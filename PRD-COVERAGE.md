# PRD Coverage

Source PRD: `/Users/Shared/product-docs/prd-core/p-cfl/p-cfl.md`.

## Required APIs

- `encode(input, from, where)` → implemented by `api=encode`.
- `decode(pattern_ref)` → implemented by `api=decode`.
- `derive(parent_ref, new_content, modify_type, from)` → implemented by `api=derive`.
- `measure_uncertainty(pattern_ref, receiver_prior)` → implemented by `api=measure_uncertainty`.
- `measure_info_density(pattern_ref)` → implemented by `api=measure_info_density`.

## Four main contradictions

1. **Information pattern recognition**
   - Deterministic metadata extraction in `patternMetadata`.
   - Emergent metadata includes physical type, origin, top-level keys, field count, token distribution.
2. **Reducing information uncertainty**
   - `measure_uncertainty` computes KL divergence from explicit receiver prior to deterministic posterior token distribution.
3. **Ensuring information density**
   - `measure_info_density` computes Shannon entropy and threshold pass/fail.
   - `encode`/`derive` reject content below the entropy threshold.
4. **Spacetime consistency and tamper resistance**
   - Stored patterns always include `from`, auto-injected `when`, `where`, and immutable `what`.
   - JSONL append-only store carries `prev_hash` + `record_hash` chain.
   - There is no update/delete/lock/unlock API; mutation attempts fail and point to `derive`.

## From authentication modes

- `bornfly`: Ed25519 signature over canonical `what` / `new_content`.
- `cfl`: requires `cfl_id`, `layer=L0|L1|L2`, and non-empty `cert`.
- `model`: requires `model_id`, `sha256:` `prompt_hash`, and positive `call_ts`.
- `v-check`: requires `v_cfl_id` and `sha256:` `record_hash`.

## Derive modify types

- `standalone-modify`
- `structuring`
- `adaptation`
- `mod`

All derive writes create a new PatternRef and preserve the parent.
