/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2022-2026, daeuniverse Organization <dae@v2raya.org>
 */

package control

import (
	"time"

	"github.com/daeuniverse/dae/common/consts"
	"github.com/daeuniverse/dae/component/outbound/dialer"
)

// NodeLatencySnapshot describes the latest observable latency status for one node link.
type NodeLatencySnapshot struct {
	Link      string
	LatencyMs *int32
	Alive     bool
	Message   string
	CheckedAt time.Time
	// RecentMs is chronological TCP probe samples (oldest → newest), up to 10.
	RecentMs []int32
}

// TriggerLatencyChecks asks dialers to refresh their TCP latency samples.
func (c *ControlPlane) TriggerLatencyChecks() {
	if c == nil {
		return
	}
	for _, group := range c.outbounds {
		if group == nil {
			continue
		}
		for _, d := range group.Dialers {
			if d == nil {
				continue
			}
			d.NotifyCheck()
		}
	}
}

// ActiveTCPConnections returns the number of currently tracked incoming TCP connections.
func (c *ControlPlane) ActiveTCPConnections() (n int) {
	if c == nil {
		return 0
	}
	if manager, _ := c.controlPlaneSessionManager(); manager != nil {
		n = manager.ActiveTCPConnections()
	}
	c.inConnections.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// SnapshotNodeLatencies returns one best-effort latency snapshot per unique node link.
func (c *ControlPlane) SnapshotNodeLatencies() []NodeLatencySnapshot {
	if c == nil {
		return nil
	}

	latenciesByLink := make(map[string]NodeLatencySnapshot)
	for _, group := range c.outbounds {
		if group == nil {
			continue
		}
		for _, d := range group.Dialers {
			if d == nil || d.Property() == nil || d.Property().Link == "" {
				continue
			}

			snapshot := bestNodeLatencySnapshotForDialer(d)
			if existing, ok := latenciesByLink[snapshot.Link]; !ok || preferNodeLatencySnapshot(snapshot, existing) {
				latenciesByLink[snapshot.Link] = snapshot
			}
		}
	}

	results := make([]NodeLatencySnapshot, 0, len(latenciesByLink))
	for _, snapshot := range latenciesByLink {
		results = append(results, snapshot)
	}
	return results
}

func bestNodeLatencySnapshotForDialer(d *dialer.Dialer) NodeLatencySnapshot {
	link := ""
	if d != nil && d.Property() != nil {
		link = d.Property().Link
	}
	snapshot := NodeLatencySnapshot{
		Link:      link,
		Alive:     false,
		Message:   "no latency result",
		CheckedAt: time.Time{},
	}

	checkTypes := []*dialer.NetworkType{
		{
			L4Proto:   consts.L4ProtoStr_TCP,
			IpVersion: consts.IpVersionStr_4,
		},
		{
			L4Proto:   consts.L4ProtoStr_TCP,
			IpVersion: consts.IpVersionStr_6,
		},
	}

	var bestNetwork *dialer.NetworkType
	haveCandidate := false
	for _, networkType := range checkTypes {
		candidate, ok := nodeLatencySnapshotForNetwork(d, link, networkType)
		if !ok {
			continue
		}
		// First hit wins the slot; later hits must win via prefer rules (alive / lower RTT).
		if !haveCandidate || preferNodeLatencySnapshot(candidate, snapshot) {
			snapshot = candidate
			bestNetwork = networkType
			haveCandidate = true
		}
	}
	if bestNetwork != nil {
		attachRecentMs(d, bestNetwork, &snapshot)
	}

	return snapshot
}

func nodeLatencySnapshotForNetwork(d *dialer.Dialer, link string, networkType *dialer.NetworkType) (NodeLatencySnapshot, bool) {
	currentAlive := d.MustGetAlive(networkType)
	observation := d.SnapshotLastProbe(networkType)
	if !observation.CheckedAt.IsZero() {
		return snapshotFromProbeObservation(link, currentAlive, observation), true
	}

	// Backward-compatible fallback for inherited snapshots that predate the
	// explicit probe observation field. The next real health check will replace
	// this synthesized view with an accurate CheckedAt timestamp.
	latency, ok := d.MustGetLatencies10(networkType).LastLatency()
	if !ok {
		return NodeLatencySnapshot{}, false
	}
	alive := d.MustGetAlive(networkType)
	return snapshotFromProbeObservation(link, alive, dialer.DialerProbeObservationSnapshot{
		Alive:      alive,
		Latency:    latency,
		HasLatency: true,
		Message:    dialer.FormatLatencyMessage(&dialer.LatencyProbeResult{Alive: alive, Latency: latency}),
	}), true
}

func snapshotFromProbeObservation(link string, currentAlive bool, observation dialer.DialerProbeObservationSnapshot) NodeLatencySnapshot {
	snapshot := NodeLatencySnapshot{
		Link:      link,
		Alive:     currentAlive,
		CheckedAt: observation.CheckedAt,
	}
	if observation.HasLatency {
		latencyMs := int32(observation.Latency.Milliseconds())
		snapshot.LatencyMs = &latencyMs
		if currentAlive {
			snapshot.Message = dialer.FormatLatencyMessage(&dialer.LatencyProbeResult{
				Alive:   true,
				Latency: observation.Latency,
			})
		} else if observation.Message != "" {
			// Keep probe error text (e.g. dial/HTTP failure) instead of dropping it.
			snapshot.Message = observation.Message
		} else {
			snapshot.Message = "unavailable"
		}
	} else {
		snapshot.Message = observation.Message
	}
	if snapshot.Message == "" {
		snapshot.Message = "no latency result"
	}
	return snapshot
}

func attachRecentMs(d *dialer.Dialer, networkType *dialer.NetworkType, snapshot *NodeLatencySnapshot) {
	if d == nil || snapshot == nil || networkType == nil {
		return
	}
	ordered := orderedLatencySamples(d.MustGetLatencies10(networkType).Snapshot())
	if len(ordered) == 0 {
		return
	}
	snapshot.RecentMs = make([]int32, 0, len(ordered))
	for _, sample := range ordered {
		snapshot.RecentMs = append(snapshot.RecentMs, int32(sample.Milliseconds()))
	}
}

func orderedLatencySamples(snap dialer.LatenciesNSnapshot) []time.Duration {
	n := len(snap.Latencies)
	if n == 0 {
		return nil
	}
	out := make([]time.Duration, 0, n)
	for i := 0; i < n; i++ {
		idx := (snap.Head + i) % n
		out = append(out, snap.Latencies[idx])
	}
	return out
}

func preferNodeLatencySnapshot(next NodeLatencySnapshot, current NodeLatencySnapshot) bool {
	if next.Alive != current.Alive {
		return next.Alive
	}
	if next.LatencyMs != nil && current.LatencyMs == nil {
		return true
	}
	if next.LatencyMs == nil {
		return false
	}
	if current.LatencyMs == nil {
		return true
	}
	if *next.LatencyMs != *current.LatencyMs {
		return *next.LatencyMs < *current.LatencyMs
	}
	return next.CheckedAt.After(current.CheckedAt)
}
