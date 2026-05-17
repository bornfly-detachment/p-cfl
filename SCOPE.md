# Scope

`p-cfl` owns only the P-layer communication responsibilities:

1. information pattern recognition,
2. uncertainty reduction measurement,
3. information density measurement,
4. spacetime consistency and tamper resistance.

It deliberately does not perform V/R/S/E duties: truth judgement, routing legality,
subject state transitions, or evolution/global optimization.

## v1 choices for PRD open questions

- Emergent pattern recognition is deterministic: same input produces same metadata.
- `from.type=bornfly` starts as `internal`; `cfl`, `model`, and `v-check` start as `candidate`.
- Receiver prior is explicit for `measure_uncertainty`.
- Default entropy threshold is 2.0 bits/symbol.
- Physical append-only guarantee is JSONL + hash-chain + no update/delete code path.
- Derive chain has no hard depth limit.
- KL and entropy are fully implemented with stdlib math; no placeholder metrics.
