# Shipment Close-Path Classification Report

- **Shipment**: 036-S
- **Mode**: classify-close-path (READ-ONLY; manually executed per skill algorithm — no
  `classify-close-path` CLI subcommand exists in this workspace's `autoharness` build)
- **Classified at**: 2026-09-30T19:21:40-07:00 (post a1-transition state)

## Step 0(a): Manifest Load

`backlogit shipment get 036-S` → `custom_fields.items` = 14 IDs:
`046-F, 046.001-T, 046.002-T, 046.003-T, 046.004-T, 046.005-T, 046.006-T, 046.007-T,
046.008-T, 046.009-T, 046.010-T, 046.011-T, 046.012-T, 046.013-T`

Shipment `dependencies`: `[034-S]` (blocks-type predecessor; 034-S is archived/shipped).
Shipment declared status: `active`.

## Step 0(b)/(c): Pre-Close Snapshot and Classification

| ID | artifact_type | declared status | parent_id | location |
|---|---|---|---|---|
| 046-F | feature | done | - (root; no parent_id) | archive |
| 046.001-T | task | done | 046-F | archive |
| 046.002-T | task | done | 046-F | archive |
| 046.003-T | task | done | 046-F | archive |
| 046.004-T | task | done | 046-F | archive |
| 046.005-T | task | done | 046-F | archive |
| 046.006-T | task | done | 046-F | archive |
| 046.007-T | task | done | 046-F | archive |
| 046.008-T | task | done | 046-F | archive |
| 046.009-T | task | done | 046-F | archive |
| 046.010-T | task | done | 046-F | archive |
| 046.011-T | task | done | 046-F | archive |
| 046.012-T | task | done | 046-F | archive |
| 046.013-T | task | done | 046-F | archive |

No snapshot ambiguity (no ID found in both queue+archive) or missing record.

**Feature-member qualification (`046-F`)**:
- Root: confirmed — no `parent_id` declared in `046-F`'s frontmatter.
- Full coverage at every depth: confirmed — exhaustive scan of `.backlogit/queue/` and
  `.backlogit/archive/` for any record declaring `parent_id: 046.0XX-T` (grandchildren of
  `046-F`) found **zero** matches. Direct children = exactly the 13 manifest tasks
  (`046.001-T`..`046.013-T`); no deeper descendants exist.
- Set-equal to manifest: manifest = `{046-F} ∪ {13 tasks}` = exactly the feature's full
  descendant closure. No manifest member falls outside this set; no descendant falls outside
  the manifest.
- Terminal: confirmed — no manifest member declares `046-F` as its parent beyond the 13
  already-enumerated tasks (same scan).

**Linked-deliberation extension**: `046-F.custom_fields` = `{harness_status: pending}` — no
`source_deliberation_id` field. Description/references text scanned with
`\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b`: no match (the `references` list contains a file path
`docs/decisions/2026-09-28-...-deliberation.md`, not a literal deliberation-ID token). **No
linked deliberation IDs to extend the snapshot with.**

**Mixed qualification**: n=1 feature member only (`046-F`); not applicable — single feature,
fully qualifying.

## Verdict

- `CLOSE_PATH_VERDICT`: **CASCADE**
- `VERDICT_REASON`: **FULLY_COVERED_ROOT**
- `CLASSIFICATION_BINDING`: `b8ba3fba87bf4bfed1d9c237e075542dae43da2b92b9bef59afdc21b17daebd7`
- `CLASSIFIED_AT`: 2026-09-30T19:21:40-07:00

### Binding computation inputs (canonical v1 serialisation)

```
v1
shipment=036-S
verdict=CASCADE
reason=FULLY_COVERED_ROOT
skill=1c7d523de56e3aa8f1597c75365c9c8c9a29c3d44c9e098d40dac9d34d573ff2
engine=backlogit version 1.11.0 (latest: v1.11.0 -- up to date)
manifest=046-F,046.001-T,046.002-T,046.003-T,046.004-T,046.005-T,046.006-T,046.007-T,046.008-T,046.009-T,046.010-T,046.011-T,046.012-T,046.013-T
deps=034-S
status=active
046-F<US>feature<US>done<US>-<US>archive
046.001-T<US>task<US>done<US>046-F<US>archive
046.002-T<US>task<US>done<US>046-F<US>archive
046.003-T<US>task<US>done<US>046-F<US>archive
046.004-T<US>task<US>done<US>046-F<US>archive
046.005-T<US>task<US>done<US>046-F<US>archive
046.006-T<US>task<US>done<US>046-F<US>archive
046.007-T<US>task<US>done<US>046-F<US>archive
046.008-T<US>task<US>done<US>046-F<US>archive
046.009-T<US>task<US>done<US>046-F<US>archive
046.010-T<US>task<US>done<US>046-F<US>archive
046.011-T<US>task<US>done<US>046-F<US>archive
046.012-T<US>task<US>done<US>046-F<US>archive
046.013-T<US>task<US>done<US>046-F<US>archive
```
(`<US>` = ASCII 0x1F Unit Separator; lines joined with `\n`, no trailing newline, SHA-256 of
UTF-8 bytes.)

Note: `skill=` hash is SHA-256 of `.github/skills/shipment-reconcile/SKILL.md` as it exists on
`main` at merge commit `5fdd75aec21bb9492aabad08603a5479c1da6846` (verified unchanged by this
shipment's own merge via the earlier mandatory pre-self-close context reload diff check).
