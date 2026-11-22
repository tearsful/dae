# Upstream bpf/tproxy backport status

## Landed (cherry-pick or manual port)

| Upstream | Notes |
|---|---|
| `#844` ipv6 `parse_transport` | cherry-pick |
| `#845` bpf timer | cherry-pick |
| `#912` remove sockmap fast redirect | cherry-pick |
| `#925` LAN ingress direct-route filter | cherry-pick |
| `#939` fwmark on non-SYN LAN ingress | cherry-pick |
| `#986` GCC 15 CO-RE | cherry-pick |
| `#1087` `TC_ACT_PIPE` on wan_egress pass-through | manual port |
| `#1172` clang-23 cmdline index bound | manual port |
| `#991` cmdline pname parsing | already present; kept + clang-23 bound |
| `#995` `bpf_get_current_comm` fallback | `PARAM.has_bpf_get_current_task` |
| `#822` separate L2/L3 netdev programs | **manual port**: `*_l2`/`*_l3` SEC + `linkHdrLen()` loader |
| `#1160` L3 strict-alignment loads | **covered** by `bpf_skb_load_bytes` parse path (no `parse_transport_fast` yet) |
| `#1099` kdae full sync | **merged** to main; outbound via `github.com/tearsful/outbound` ≥ v0.3.20 (stickyip, scoped cache, NewDirectDialers, datapath resolver) |
| `#1110` LAN UDP exact bind | **included** with kdae (`sock_bound_to_daddr`; skip dport 53); **closed on macOS via Docker**: `./scripts/docker-ebpf-verify.sh` runs `make ebpf` + `make ebpf-test` (all PASS) |

Also: `common/bitlist` Tighten / Cap expectations aligned with upstream.

## Local Linux loop (Docker)

On macOS with Docker Desktop enabled:

```bash
./scripts/docker-ebpf-verify.sh
```

The script mounts `debugfs`/`tracefs`, compiles real tproxy/direct_stats objects (non-stub), runs `#1110` ab_regression, and the production datapath ABI whitelist. Soft-reports `TestTCHookSetKernelUpdateAndRollback` — Docker Desktop linuxkit often cannot pass a packet through the TC chain; real Ubuntu CI still owns that case.

`control/kern/tests` opens `/sys/kernel/tracing/trace_pipe` with fallback to `/sys/kernel/debug/tracing/trace_pipe`.

## Still deferred

| Upstream | Why deferred |
|---|---|
| fast-path `parse_transport_fast` + remaining `#1160` helpers | optional perf path; bring `pkt_load16/32` together |
| `TestTCHookSetKernelUpdateAndRollback` on Docker Desktop | linuxkit TC packet path; not a dae regression — use Ubuntu CI / bare metal |

### Suggested next

1. Optionally port `parse_transport_fast` / remaining `#1164` control-plane API polish.
2. Keep wing/daed pinned to tearsful/dae main + tearsful/outbound tags.
3. Re-run `./scripts/docker-ebpf-verify.sh` after bpf/tproxy changes.
