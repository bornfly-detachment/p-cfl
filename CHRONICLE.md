# p-cfl Chronicle ↔ Git ledger

- Recorded at: 2026-05-18 (Asia/Shanghai)
- GitHub repository: https://github.com/bornfly-detachment/p-cfl
- Local implementation source: /Users/Shared/codex-workspace/p-cfl
- Local Git sync clone: /Users/Shared/intelligence-resource-management/runtime/go/p-cfl
- PRD source: /Users/Shared/product-docs/prd-core/p-cfl/
- Runtime scope: Pattern CFL MVP runtime, state/event evidence, and pattern execution primitives.
- Provenance note: the repository content was cross-checked against the codex workspace CFL directory before this ledger was committed, so Chronicle and Git point to the same artifact boundary.
- Verification command: `go test ./...`

## 2026-05-18 handoff record

This record anchors the p-cfl code to its standalone GitHub repository after Claude Code-created repository setup and Codex-side local CFL implementation work. Future changes should update this ledger or add a dated Chronicle entry in the same repository.

## 2026-05-19 IRM shared resource relocation

Moved the local Git sync clone into the shared Intelligence Resource Management Go CFL resource pool:

```text
/Users/Shared/intelligence-resource-management/runtime/go/p-cfl
```

This path is now the canonical local runtime path for cross-project calls. The previous loose top-level `/Users/Shared/p-cfl` path should be treated as retired.
