---
shipment_id: 025-S
mode: classify-close-path
timestamp: 2026-09-20T07:23:00Z
---

# Close-Path Classification Report — 025-S

## Step 0(a) — Manifest Load

`items`: `028-F`, `028.001-T`, `028.002-T`, `028.003-T`

## Step 0(b) — Snapshot

| Item | Queue | Archive | Snapshot |
|---|---|---|---|
| 028-F | absent | present (`status: done`) | clean |
| 028.001-T | absent | present (`status: done`) | clean |
| 028.002-T | absent | present (`status: done`) | clean |
| 028.003-T | absent | present (`status: done`) | clean |

No item present in both queue and archive (no torn/ambiguous snapshot). No item missing
from both locations.

## Step 0(c) — Classification

* Feature members in manifest: `028-F` (n=1).
* `028-F.parent_id`: none declared → `028-F` is a root feature.
* Descendant enumeration (via `backlogit get 028-F` `size_composition.members`, read
  immediately prior to the covering-feature completion transition): exactly
  `028.001-T`, `028.002-T`, `028.003-T` — no other descendants at any depth.
* Manifest task members (`028.001-T`, `028.002-T`, `028.003-T`) are set-equal to
  `028-F`'s full descendant set. `028-F` is therefore fully covered at every depth by
  this manifest.
* Every manifest member is a descendant of (or is) the qualifying feature — no
  manifest member falls outside `028-F`'s subtree.
* Only one feature member present — no mixed qualification is possible.
* No enumeration error, no ambiguity, no classifier error encountered.

## Verdict

* `CLOSE_PATH_VERDICT`: `CASCADE`
* `VERDICT_REASON`: `FULLY_COVERED_ROOT`
* `VERDICT_EVIDENCE`: `028-F` is a root feature (no `parent_id`) whose complete
  descendant set `{028.001-T, 028.002-T, 028.003-T}` is set-equal to the shipment
  manifest's task members; all four manifest members (`028-F` + 3 tasks) are already
  archived with `status: done`.
* `CLASSIFICATION_BINDING`: `sha256(025-S|028-F|028.001-T,028.002-T,028.003-T|FULLY_COVERED_ROOT|20260920T072300Z)`
