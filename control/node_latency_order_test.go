/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2022-2026, daeuniverse Organization <dae@v2raya.org>
 */

package control

import (
	"testing"
	"time"

	"github.com/daeuniverse/dae/component/outbound/dialer"
)

func TestOrderedLatencySamplesRing(t *testing.T) {
	ln := dialer.NewLatenciesN(3)
	ln.AppendLatency(10 * time.Millisecond)
	ln.AppendLatency(20 * time.Millisecond)
	ln.AppendLatency(30 * time.Millisecond)
	// Ring full; next append drops 10ms.
	ln.AppendLatency(40 * time.Millisecond)

	got := orderedLatencySamples(ln.Snapshot())
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}
	want := []time.Duration{20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d]=%v, want %v (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestOrderedLatencySamplesPartial(t *testing.T) {
	ln := dialer.NewLatenciesN(5)
	ln.AppendLatency(5 * time.Millisecond)
	ln.AppendLatency(15 * time.Millisecond)
	got := orderedLatencySamples(ln.Snapshot())
	if len(got) != 2 || got[0] != 5*time.Millisecond || got[1] != 15*time.Millisecond {
		t.Fatalf("unexpected partial order: %v", got)
	}
}

func TestPreferNodeLatencySnapshotAliveWins(t *testing.T) {
	ms := int32(999)
	dead := NodeLatencySnapshot{Alive: false, LatencyMs: &ms}
	alive := NodeLatencySnapshot{Alive: true, LatencyMs: &ms}
	if !preferNodeLatencySnapshot(alive, dead) {
		t.Fatal("alive should beat dead")
	}
	if preferNodeLatencySnapshot(dead, alive) {
		t.Fatal("dead should not beat alive")
	}
}

func TestPreferNodeLatencySnapshotRTTAndCheckedAt(t *testing.T) {
	low := int32(30)
	high := int32(80)
	older := time.Unix(100, 0)
	newer := time.Unix(200, 0)

	if !preferNodeLatencySnapshot(
		NodeLatencySnapshot{Alive: true, LatencyMs: &low},
		NodeLatencySnapshot{Alive: true, LatencyMs: &high},
	) {
		t.Fatal("lower RTT should win")
	}
	if preferNodeLatencySnapshot(
		NodeLatencySnapshot{Alive: true, LatencyMs: nil},
		NodeLatencySnapshot{Alive: true, LatencyMs: &low},
	) {
		t.Fatal("nil LatencyMs should not beat a measured sample")
	}
	if !preferNodeLatencySnapshot(
		NodeLatencySnapshot{Alive: true, LatencyMs: &low, CheckedAt: newer},
		NodeLatencySnapshot{Alive: true, LatencyMs: &low, CheckedAt: older},
	) {
		t.Fatal("newer CheckedAt should win on equal RTT")
	}
}

func TestSnapshotFromProbeObservationMessages(t *testing.T) {
	ms := 12 * time.Millisecond
	alive := snapshotFromProbeObservation("ss://a", true, dialer.DialerProbeObservationSnapshot{
		HasLatency: true,
		Latency:    ms,
		Alive:      true,
		Message:    "12ms",
	})
	if alive.Message != "12ms" || alive.LatencyMs == nil || *alive.LatencyMs != 12 {
		t.Fatalf("alive success: %+v", alive)
	}

	probeFail := snapshotFromProbeObservation("ss://a", false, dialer.DialerProbeObservationSnapshot{
		HasLatency: false,
		Alive:      false,
		Message:    "dial tcp: i/o timeout",
	})
	if probeFail.Message != "dial tcp: i/o timeout" {
		t.Fatalf("probe fail message: %q", probeFail.Message)
	}

	// Traffic/other path marked dead; LastProbe still a success sample.
	stale := snapshotFromProbeObservation("ss://a", false, dialer.DialerProbeObservationSnapshot{
		HasLatency: true,
		Latency:    ms,
		Alive:      true,
		Message:    "12ms",
	})
	if stale.Message != "unavailable" {
		t.Fatalf("stale success while dead should not keep Nms text, got %q", stale.Message)
	}
	if stale.Alive {
		t.Fatal("stale snapshot should report currentAlive=false")
	}
	if stale.LatencyMs == nil || *stale.LatencyMs != 12 {
		t.Fatalf("stale dead snapshot should keep last LatencyMs: %+v", stale)
	}
}
