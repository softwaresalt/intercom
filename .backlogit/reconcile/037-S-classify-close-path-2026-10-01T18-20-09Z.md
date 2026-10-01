# Shipment Close-Path Classification Report

- **Shipment**: 037-S
- **Mode**: classify-close-path (READ-ONLY; agent-executed per the skill algorithm via a
  session-local snapshot helper — no `classify-close-path` CLI subcommand exists in this
  workspace's `autoharness` build)
- **Classified at**: 2026-10-01T18:20:09Z (post a1-transition state)

## Step 0(a): Manifest Load

`backlogit shipment get 037-S` → `custom_fields.items` = 10 IDs:
`047-F, 047.001-T, 047.002-T, 047.003-T, 047.004-T, 047.005-T, 047.006-T, 047.007-T, 047.008-T, 047.009-T`

Shipment `dependencies`: `[035-S, 036-S]` (blocks-type predecessors; both archived/shipped).
Shipment declared status: `active`.

## Step 0(b)/(c): Pre-Close Snapshot and Classification

| ID | artifact_type | declared status | parent_id | location |
|---|---|---|---|---|
| 047-F | feature | done | - (root; no parent_id) | queue |
| 047.001-T | task | done | 047-F | queue |
| 047.002-T | task | done | 047-F | queue |
| 047.003-T | task | done | 047-F | queue |
| 047.004-T | task | done | 047-F | queue |
| 047.005-T | task | done | 047-F | queue |
| 047.006-T | task | done | 047-F | queue |
| 047.007-T | task | done | 047-F | queue |
| 047.008-T | task | done | 047-F | queue |
| 047.009-T | task | done | 047-F | queue |

No snapshot ambiguity (no ID found in both queue+archive) or missing record.

**Feature-member qualification (`047-F`)**:
- Root: confirmed — no `parent_id` declared in `047-F`'s frontmatter.
- Full coverage at every depth: confirmed — exhaustive scan of `.backlogit/queue/` and
  `.backlogit/archive/` for any record declaring `parent_id` equal to any `047.0XX-T`
  (grandchildren of `047-F`) found **zero** matches. Direct children = exactly the 9 manifest
  tasks (`047.001-T`..`047.009-T`); no deeper descendants exist.
- Set-equal to manifest: manifest = `{047-F} ∪ {9 tasks}` = exactly the feature's full
  descendant closure.
- Terminal: confirmed (same scan).

**Linked-deliberation extension**: `047-F.custom_fields` = `{harness_status: pending}` — no
`source_deliberation_id` field. Description/references text scanned with
`\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b`: no match (the `references` list contains a file path
`docs/decisions/2026-09-28-...-deliberation.md`, not a literal deliberation-ID token). **No
linked deliberation IDs to extend the snapshot with.**

**Mixed qualification**: n=1 feature member only (`047-F`); not applicable.

## Verdict

- `CLOSE_PATH_VERDICT`: **CASCADE**
- `VERDICT_REASON`: **FULLY_COVERED_ROOT**
- `CLASSIFICATION_BINDING` (corrected): `77985d03e202c5ff61f1ba431d6d087aa7f2b349395eeceaadd3ca41b263a5b4`
- `CLASSIFICATION_BINDING` (as issued at 18:20:09Z, defective):
  `0f9be8a39309d70cd01b641871f84de68fc0abe5a80bce03fe394a0c230792d9`
- `CLASSIFIED_AT`: 2026-10-01T18:20:09Z

### Correction: snapshot location defect (disclosed post-close)

The session-local snapshot helper recorded `location = archive` for all ten members. Their
actual pre-close location was `.backlogit/queue/`. Three pieces of authoritative evidence show
this:

1. At 18:19:59Z, the pre-mode reconcile report found all ten members in `.backlogit/queue/` with
   status `done`.
2. Every archived record carries `archived_from: .backlogit/queue/<id>.md` and
   `archived_status: done`. Each record's `updated_at` falls between 18:20:44Z and 18:20:57Z,
   which is after the 18:20:42Z cascade mutation. The cascade therefore moved them from queue
   to archive, so they cannot have been in archive at 18:20:09Z.
3. In merge commit `ec6d1d9`, all ten records and `037-S.md` are under `.backlogit/queue/`.

The snapshot table and binding inputs above now hold the corrected `queue` values. Recomputing
the canonical v1 serialisation from them gives `77985d03…a5b4`. Running the same serialisation
with `archive` reproduces the as-issued value `0f9be8a3…92d9`, which confirms that location is
the only field that differed.

The verdict does not change. The qualification algorithm uses location only to detect snapshot
ambiguity (an ID present in both queue and archive), and there was none. Root status, full
coverage at every depth, set-equality with the manifest, and terminal status all hold as
recorded. For the effect of this defect on the safe-close Step 0 revalidation, see the
cascade-close report.

### Binding computation inputs (canonical v1 serialisation, corrected locations)

```
v1
shipment=037-S
verdict=CASCADE
reason=FULLY_COVERED_ROOT
skill=1c7d523de56e3aa8f1597c75365c9c8c9a29c3d44c9e098d40dac9d34d573ff2
engine=backlogit version 1.11.0 (latest: v1.11.0 -- up to date)
manifest=047-F,047.001-T,047.002-T,047.003-T,047.004-T,047.005-T,047.006-T,047.007-T,047.008-T,047.009-T
deps=035-S,036-S
status=active
047-F<US>feature<US>done<US>-<US>queue
047.001-T<US>task<US>done<US>047-F<US>queue
047.002-T<US>task<US>done<US>047-F<US>queue
047.003-T<US>task<US>done<US>047-F<US>queue
047.004-T<US>task<US>done<US>047-F<US>queue
047.005-T<US>task<US>done<US>047-F<US>queue
047.006-T<US>task<US>done<US>047-F<US>queue
047.007-T<US>task<US>done<US>047-F<US>queue
047.008-T<US>task<US>done<US>047-F<US>queue
047.009-T<US>task<US>done<US>047-F<US>queue
```
(`<US>` = ASCII 0x1F Unit Separator; lines joined with `\n`, no trailing newline, SHA-256 of
UTF-8 bytes.)

Note: `skill=` hash is SHA-256 of `.github/skills/shipment-reconcile/SKILL.md` as it exists on
`main` at merge commit `ec6d1d9938559292363508134ddbf462a62e564d` (`git diff ec6d1d9~1 ec6d1d9`
shows the file unchanged by this shipment's merge).
