---
shipment_id: 029-S
mode: classify-close-path
generated_at: 2026-09-28T23:48:11Z
---

# Classify-Close-Path Report — 029-S

## Manifest

`items`: `032-F`, `032.001-T`, `032.002-T`, `032.003-T`, `032.004-T`, `032.005-T`, `032.006-T`, `032.007-T`

## Per-Member Evidence

| ID | artifact_type | declared status (pre-close) | parent_id | location | precondition result |
|---|---|---|---|---|---|
| 032-F | feature | done | (none — root) | archive | root ✓, fully covered at every depth ✓ |
| 032.001-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.002-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.003-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.004-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.005-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.006-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |
| 032.007-T | task | done | 032-F | archive | descendant of 032-F, in manifest ✓ |

The full-depth descendant scan of 032-F returns exactly the 7 manifest tasks. The scan looked for `parent_id` in {032-F, 032.00[1-7]-T} across `.backlogit/queue/` and `.backlogit/archive/`.

None of the 7 tasks has its own subtasks.

There are no linked deliberations:
- 032-F has no `custom_fields.source_deliberation_id`.
- The `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` matcher finds nothing in its description or references. The references are file paths, and the description cites stash 6C24E2E4.

## Verdict

`CLOSE_PATH_VERDICT: CASCADE`
`VERDICT_REASON: FULLY_COVERED_ROOT`

## Classification Binding

`CLASSIFICATION_BINDING: 7da9434d3f50248e90430bc3ee4c6bbd645447c9a11eda6fb6e7aa501bdd1795`
(canonical v1 serialisation in `.backlogit/reconcile/029-S-classification-input.txt`)

`CLASSIFIED_AT`: 2026-09-28T23:48:11Z
