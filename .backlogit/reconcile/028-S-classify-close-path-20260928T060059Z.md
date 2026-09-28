---
shipment_id: 028-S
mode: classify-close-path
generated_at: 2026-09-28T06:00:59Z
---

# Classify-Close-Path Report — 028-S

## Manifest

`items`: `031-F`, `031.001-T`, `031.002-T`, `031.003-T`, `031.004-T`, `031.005-T`

## Per-Member Evidence

| ID | artifact_type | declared status (pre-close) | parent_id | location | precondition result |
|---|---|---|---|---|---|
| 031-F | feature | done | (none — root) | archive | root ✓, fully covered at every depth ✓ |
| 031.001-T | task | done | 031-F | archive | descendant of 031-F, in manifest ✓ |
| 031.002-T | task | done | 031-F | archive | descendant of 031-F, in manifest ✓ |
| 031.003-T | task | done | 031-F | archive | descendant of 031-F, in manifest ✓ |
| 031.004-T | task | done | 031-F | archive | descendant of 031-F, in manifest ✓ |
| 031.005-T | task | done | 031-F | archive | descendant of 031-F, in manifest ✓ |

The full-depth descendant scan of 031-F (`parent_id` in {031-F, 031.00[1-5]-T} across `.backlogit/queue/` and
`.backlogit/archive/`) returns exactly the 5 manifest tasks. None of the 5 tasks has its own subtasks. There are no linked
deliberations: 031-F has no `custom_fields.source_deliberation_id`, and the `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` matcher
finds nothing in its description or references.

## Verdict

`CLOSE_PATH_VERDICT: CASCADE`
`VERDICT_REASON: FULLY_COVERED_ROOT`

## Classification Binding

`CLASSIFICATION_BINDING: a0b976c8e71af10bc5800154a4e4ae0502812d8302493d57a6996df0d4ef2e0e`
(canonical v1 serialisation in `.backlogit/reconcile/028-S-classification-input.txt`)

`CLASSIFIED_AT`: 2026-09-28T06:00:59Z
