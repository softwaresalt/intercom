---
title: "Compacted memory — pathsafe Windows canonicalization symmetry and Windows lint coverage (015-S)"
date: 2026-09-10
shipment: 015-S
feature: 016-F
status: shipped
pr: 48
merge_commit_sha: fbd8d8d0ab679e254b2b817af69b8e8e673626c5
compacted_from:
  - docs/archive/memory/2026-09-10-stage-015-s-session-summary.md
  - docs/archive/memory/2026-09-10-015-s-pathsafe-windows-canonicalization-pre-pr-checkpoint.md
---

# Compacted memory — 015-S / 016-F pathsafe Windows canonicalization symmetry

- Release unit: shipment 015-S, feature 016-F, tasks 016.001-T–016.008-T
- Date: 2026-09-10
- Outcome: **SHIPPED** — PR #48 merged (merge commit
  `fbd8d8d0ab679e254b2b817af69b8e8e673626c5`), shipment cascade-closed via
  the P-015 verified fully-covered-root path, post-merge closure complete.
- Original verbose checkpoints archived to `docs/archive/memory/`:
  `2026-09-10-stage-015-s-session-summary.md`,
  `2026-09-10-015-s-pathsafe-windows-canonicalization-pre-pr-checkpoint.md`.

## Stage phase (queue → review → harvest)

- Queued shipment 015-S covering feature 016-F + 8 tasks, harvesting stash
  entries `4104AF54`, `E428AB46`, `E4C5413F`. `BF5DE670` deliberately
  deferred again (write-path gate trigger re-measured, still HAS NOT FIRED
  — third consecutive cycle with the same result).
- Plan review: FAIL at attempt 1 (two P1s: B3/long-path handling not a hard
  C2 prerequisite; C2 silently changed `NewRoot`'s Windows error type from
  `*fs.PathError` to raw `syscall.Errno`) → both fixed → PASS at attempt 2.
- Key plan decisions: **D-2** (`NewRoot` adopts `canonicalizeReparse`,
  provable no-op on POSIX), **D-4** (BF5DE670 deferred, evidence-backed),
  **D-5** (corrected a twice-repeated false claim that EvalSymlinks calls
  GetFinalPathNameByHandleW — must not re-enter any artifact), belt-and-
  suspenders retention of `stripUNCPrefix` at both call sites (root.go
  D-2, pathsafe.go B2/AC3).
- Workspace degradation noted (not fixed, out of scope): task artifact
  type has no structured `complexity` field; `size` is structured,
  complexity recorded as labeled prose.

## Ship phase (implement → review → PR → merge → close)

- TDD build loop across all 8 tasks: 016.001-T (Windows lint CI gate),
  016.002-T (GetFinalPathNameByHandleW guard before CreateFile),
  016.003-T (internalize stripUNCPrefix postcondition), 016.004-T
  (long-path prefixing, paired hard-prerequisite with 016.007-T),
  016.005-T (RED junction-root availability lock), 016.006-T (critical
  containment-preservation lock), 016.007-T (critical green change:
  `NewRoot` → `canonicalizeReparse`), 016.008-T (risk-register/doc
  reconciliation).
- Empirical correction during 016.007-T: `GetFinalPathNameByHandleW`
  normalizes a `\\?\Volume{GUID}\...` root back to its DOS drive-letter
  form when the volume has one mounted (VOLUME_NAME_DOS default) — D-6's
  GUID-preservation guarantee is a property of `stripUNCPrefix` as a pure
  function, not something that survives full `canonicalizeReparse`
  round-trip when a drive letter exists. Test assertion corrected
  accordingly.
- Standard + 6-persona adversarial review (Go/Security/Correctness/
  Concurrency/Maintainability/Architecture): zero P0/P1. One
  plan-cross-check false positive (4 reviewers flagged `stripUNCPrefix`
  retention as apparent dead code — already adjudicated by D-2/B2-AC3;
  resolved with clarifying comments only, no code removal). One
  genuinely out-of-scope finding (`checkSymlinkEscape`'s error wrapping
  inconsistent with `NewRoot`'s — AC3 scopes error-surface preservation to
  NewRoot only) deferred per P-021 C2 as entry `7ADAC481`. Advisory-only
  hardening ideas consolidated as stash `6B751D8B`. Full detail:
  `docs/closure/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-adversarial-review.md`.
- Post-merge bookkeeping correction: 016.005-T was never transitioned to
  `done` during the build loop despite its tests existing and passing
  green after 016.007-T landed — caught during pre-archive reconciliation
  and corrected before shipment closure (no code change).
- **Process gap, disclosed transparently**: Copilot PR review was not
  actively engaged before merge — only a GraphQL `suggestedActors` probe
  (empty) and an unsuffixed `gh pr edit --add-reviewer` (failed) were
  tried, not the previously-documented working `[bot]`-suffixed REST
  fallback (`docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`).
  `autoharness gate copilot-review` returned `NOT_APPLICABLE` both
  pre-merge and at the last-mile re-check, so merge proceeded per the
  gate's own contract — but the documented fallback should have been tried
  first. Once merged, a retry of the REST call registered no actual review
  request (window closed). Compensating action: an independent fresh
  Correctness-Reviewer pass was run against the final merged diff
  post-merge; it found zero new issues beyond what the original 6-persona
  review already surfaced. New compound entry captures the lesson:
  `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`.
- Shipment safe-close used the P-015 verified fully-covered-root cascade
  path (`backlogit shipment ship`): 016-F is a root feature (no
  `parent_id`) whose full child set (exactly the 8 manifest tasks, per
  `size_composition.members`) matches the manifest with no additional
  members; `returned_ids` was empty, confirming exact match.
- Source artifact cleanup: `custom_fields.source_stash_id` /
  `source_deliberation_id` absent on both `015-S` and `016-F` (harvested
  stash IDs `4104AF54`/`E428AB46`/`E4C5413F` cited in prose description
  only, already consumed/removed by Stage's harvest — confirmed absent
  from `.backlogit/stash.jsonl`). Third consecutive shipment (after
  013-S, 014-S) confirming this is the repository's Stage-harvest norm,
  not a defect. `BF5DE670` untouched throughout, per operator's binding
  scope contract.

## Invariants preserved

- Windows root/candidate canonicalization asymmetry: CLOSED (013-S
  regression, closed by 016.007-T).
- `BF5DE670` (write-through-dangling-symlink / TOCTOU / hardlink-count
  mitigation): still deferred, trigger unfired, untouched by this
  shipment.
- `stripUNCPrefix` belt-and-suspenders retention at both call sites: an
  intentional, documented design decision (D-2, B2/AC3) — do not remove in
  a future pass without revisiting the decision.
