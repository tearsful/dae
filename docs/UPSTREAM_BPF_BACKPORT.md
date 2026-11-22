# Upstream bpf/tproxy backport status

Branch: `chore/upstream-bpf-tproxy`

## Landed (cherry-pick or manual port)

| Upstream | Notes |
|---|---|
| `#844` ipv6 `parse_transport` | cherry-pick |
| `#845` bpf timer | cherry-pick |
| `#912` remove sockmap fast redirect | cherry-pick |
| `#925` LAN ingress direct-route filter | cherry-pick |
| `#939` fwmark on non-SYN LAN ingress | cherry-pick |
| `#986` GCC 15 CO-RE | cherry-pick |
| `#1087` `TC_ACT_PIPE` on wan_egress pass-through | **manual port** (tree diverged from L2/L3 split) |
| `#1172` clang-23 cmdline index bound | **manual port** (`asm` barrier instead of `barrier_var`) |
| `#991` cmdline pname parsing | already present in tearsful tree (kept + clang-23 bound) |
| `#995` `bpf_get_current_comm` fallback | **manual port** via `PARAM.has_bpf_get_current_task` (no stub rewrite) |

Also: `common/bitlist` Tighten / Cap expectations aligned with upstream.

## Still deferred (hard conflict with tearsful tree)

These need a larger rebase onto upstream control/kern (L2/L3 program split, kdae sync, stub layout):

| Upstream | Why deferred |
|---|---|
| `#822` separate L2/L3 netdev programs | massive tproxy.c / loader rewrite |
| `#1160` avoid 4-byte L3 loads | depends on L2/L3 split + deleted verifier test |
| `#1099` kdae full sync | full datapath rewrite — do as dedicated rebase |

Until those land, prefer keeping the landed safety/correctness fixes above rather than waiting on a full upstream rebase.
