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

Also: `common/bitlist` Tighten / Cap expectations aligned with upstream.

## Still deferred

| Upstream | Why deferred |
|---|---|
| `#1099` kdae full sync | landed on `chore/kdae-datapath-rebase` (outbound via `github.com/tearsful/outbound` v0.3.20+) |
| `#1110` LAN UDP exact bind match | depends on kdae lan-ingress local-socket bypass path |
| fast-path `parse_transport_fast` + `#1160` helpers | comes with kdae; bring `pkt_load16/32` together |

### Suggested next rebase

1. Branch from current main (L2/L3 already present).
2. Take upstream `control/kern/tproxy.c` post-kdae as base.
3. Re-apply tearsful deltas: outbound module path, `PARAM.has_bpf_get_current_task`, softwind/outbound pins.
4. Port `#1110` and remaining control-plane API (`#1164`) as follow-ups.
