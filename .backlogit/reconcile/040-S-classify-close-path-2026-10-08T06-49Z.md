# Shipment Close-Path Classification Report

- **Shipment**: 040-S
- **Mode**: classify-close-path (read-only).
  - Executed per the skill algorithm. This workspace's `autoharness` build has no `classify-close-path` CLI subcommand.
  - Snapshot and binding were computed by a deterministic session script, `%TEMP%\binding040s.ps1`. It reads frontmatter from `.backlogit/queue/` and `.backlogit/archive/`.
  - The serializer was validated by reproducing the 036-S precedent binding `b8ba3fba87bf4bfed1d9c237e075542dae43da2b92b9bef59afdc21b17daebd7` byte for byte from that report's recorded inputs.
- **Classified at**: 2026-10-08T06:49Z (approximate; the binding was computed in the session immediately before the 06:50:14Z pre-close backup) (state after the a1 transition; 050-F moved to `done` and was auto-archived by backlogit)
- **Base**: branch `post-merge/050-f-retiredarch-correctness` from `main@49ff6b92121f60f3dfa5d7eff15e0dc89cc81f22`

## Step 0(a): Manifest load

- `040-S` `custom_fields.items` = `050-F, 050.001-T, 050.002-T, 050.003-T, 050.004-T, 050.006-T` (6 IDs).
- `050.005-T` does not exist by plan design (U4 and U5 land as one task, 050.004-T, under ALP-2).
- Shipment `dependencies`: `[039-S]`. This is the blocks-type predecessor; 039-S is archived and shipped. It is rendered as the bare ID `039-S`, which is the frontmatter list value and matches the 036-S/037-S precedent rendering.
- Shipment declared status: `active` (in `.backlogit/queue/040-S.md`).

## Step 0(b)/(c): Pre-close snapshot and classification

| ID | artifact_type | declared status | parent_id | location |
|---|---|---|---|---|
| 050-F | feature | done | - (root) | archive |
| 050.001-T | task | done | 050-F | archive |
| 050.002-T | task | done | 050-F | archive |
| 050.003-T | task | done | 050-F | archive |
| 050.004-T | task | done | 050-F | archive |
| 050.006-T | task | done | 050-F | archive |

There is no snapshot ambiguity: no ID was found in both queue and archive, and no record is missing.

**Feature-member qualification (`050-F`; n=1 by `artifact_type`)**:

- **Root**: confirmed. `050-F` declares no `parent_id`.
- **Full coverage at every depth**: confirmed. I walked the full `parent_id` graph to a fixed point over all 380 queue and archive records, starting at `050-F`. The walk returned exactly `{050-F, 050.001-T, 050.002-T, 050.003-T, 050.004-T, 050.006-T}`, and no record declares a parent under any `050.*` task.
- **Set-equal to manifest**: the feature plus its descendant closure is exactly the manifest.
- **Terminal**: confirmed by the same walk.

**Linked-deliberation extension**:

- `050-F.custom_fields` is `{harness_status: pending}`, with no `source_deliberation_id`.
- I scanned the description and references with `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` and found no match. The references list a file path (`docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md`), not a deliberation-ID token.
- There are no linked deliberation IDs.

**Mixed qualification**: not applicable. There is a single fully qualifying feature member.

## Verdict

- `CLOSE_PATH_VERDICT`: **CASCADE**
- `VERDICT_REASON`: **FULLY_COVERED_ROOT**
- `CLASSIFICATION_BINDING`: `17bab11e9114840724e3f4928c0e0ab939d4f49be7de89c354e51dd12359d182`
- `CLASSIFIED_AT`: 2026-10-08T06:49Z (approximate; the binding was computed in the session immediately before the 06:50:14Z pre-close backup)

### Binding computation inputs (canonical v1 serialisation)

```
v1
shipment=040-S
verdict=CASCADE
reason=FULLY_COVERED_ROOT
skill=1c7d523de56e3aa8f1597c75365c9c8c9a29c3d44c9e098d40dac9d34d573ff2
engine=backlogit version 1.11.0 (latest: v1.11.0 -- up to date)
manifest=050-F,050.001-T,050.002-T,050.003-T,050.004-T,050.006-T
deps=039-S
status=active
050-F<US>feature<US>done<US>-<US>archive
050.001-T<US>task<US>done<US>050-F<US>archive
050.002-T<US>task<US>done<US>050-F<US>archive
050.003-T<US>task<US>done<US>050-F<US>archive
050.004-T<US>task<US>done<US>050-F<US>archive
050.006-T<US>task<US>done<US>050-F<US>archive
```

`<US>` is the ASCII 0x1F Unit Separator. Lines are joined with `\n`, with no trailing newline, and the digest is the SHA-256 of the UTF-8 bytes.

The `skill=` hash is the SHA-256 of `.github/skills/shipment-reconcile/SKILL.md` at `main@49ff6b9`. This was verified as part of the mandatory context reload before self-close: `git diff f50dba4 49ff6b9 -- .github/` is empty, so this shipment did not modify the skill.

This `CLASSIFICATION_BINDING` is carried into `mode: safe-close`. Safe-close Step 0 MUST recompute it from a fresh snapshot and match it before any mutation.