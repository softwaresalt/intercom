---
shipment_id: 034-S
mode: classify-close-path
generated_at: 2026-09-30T04:00:19Z
---

# Classify-Close-Path Report — 034-S

## Manifest

`items`: `044-F`, `044.001-T`, `044.002-T`, `044.003-T`, `044.004-T`, `044.005-T`, `044.006-T`, `044.007-T`, `044.008-T`, `044.009-T`, `044.010-T`, `044.011-T`, `044.012-T`

## Per-Member Evidence

| ID | artifact_type | declared status (pre-close) | parent_id | location | precondition result |
|---|---|---|---|---|---|
| 044-F | feature | done | (none — root) | archive | root ✓, fully covered at every depth ✓ |
| 044.001-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.002-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.003-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.004-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.005-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.006-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.007-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.008-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.009-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.010-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.011-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |
| 044.012-T | task | done | 044-F | archive | descendant of 044-F, in manifest ✓ |

The full-depth descendant scan of 044-F (over `.backlogit/queue/` and `.backlogit/archive/`, matching on `parent_id`) returns exactly the 12 manifest tasks. No other item declares `044-F` as its `parent_id`.

None of the 12 tasks has its own subtasks (no item declares any `044.0xx-T` as `parent_id`).

There are no linked deliberations:
- 044-F has no `custom_fields.source_deliberation_id`.
- The `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` matcher finds nothing in its description or references. The description cites "deliberation D-1 to D-8" (the decision-record section labels, not a `DL`-form artifact ID) and the decision record path `docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`, neither of which matches the pattern.

## Verdict

`CLOSE_PATH_VERDICT: CASCADE`
`VERDICT_REASON: FULLY_COVERED_ROOT`

## Classification Binding

`CLASSIFICATION_BINDING: a36a46c5ad8ca807e2701645e33bc424dca8f6a1a8f7c1d866d0dccc9e920837`
(canonical v1 serialisation in `.backlogit/reconcile/034-S-classification-input.txt`)

`CLASSIFIED_AT`: 2026-09-30T04:00:19Z
