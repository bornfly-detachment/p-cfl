# p-cfl

Single-binary Pattern CFL implementation for PRVSE P layer.

It implements the MVP from `/Users/Shared/product-docs/prd-core/p-cfl/p-cfl.md`:

- `encode`: external information -> immutable PatternRef.
- `decode`: PatternRef -> original `what` without loss.
- `derive`: parent PatternRef + new content -> new PatternRef with derivation metadata.
- `measure_uncertainty`: KL divergence against an explicit receiver prior.
- `measure_info_density`: Shannon entropy/density check.

Properties:

- Go stdlib only; `go.mod` has an empty `require` block.
- Single binary: `p-cfl`.
- Stdin JSON / stdout JSON.
- Exit code: `0=pass`, `1=fail`, `2=n/a`.
- Append-only JSONL store with record hash chain.
- No update/delete API; all change semantics go through `derive`.

## Quick start

```sh
go build -o p-cfl .
mkdir -p .p-cfl-state
printf '%s\n' '{"api":"encode","store_dir":".p-cfl-state","input":{"message":"bornfly original pattern text with enough information density"},"from":{"type":"model","model_id":"demo","prompt_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","call_ts":1},"where":{"project":"demo"}}' | ./p-cfl
```

## Examples (4 主要矛盾覆盖)

| Script | Mains contradictions covered |
|---|---|
| `examples/roundtrip.sh` | encode/decode/measure_info_density 主线（通用 happy path） |
| `examples/derive_chain.sh` | ④ 衍生协议 — external ingest → structuring derive；父 Pattern 永存 |
| `examples/tamper_detect.sh` | ④ append-only + hash 链 — 篡改 JSONL 后 `chain_valid=false` |
| `examples/info_metrics.sh` | ② KL 散度 + ③ Shannon 熵 — 低熵 input 拒收，KL 度量不确定性下降 |

TODO（后续轮次）：`examples/bornfly_signed.sh` — 需要独立 Ed25519 keypair generator helper（违反 single binary 风格暂留）。当前 `TestFromAuthenticationModes` 已覆盖 bornfly 签名验签路径。

## Source note

`from.type=human` is accepted for direct human-entered information and starts as `candidate`. `from.type=bornfly` remains the signed, stronger source that can start as `internal`.
