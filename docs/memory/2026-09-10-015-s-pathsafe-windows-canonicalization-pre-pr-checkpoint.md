# 015-S Pre-PR Checkpoint — Pathsafe Windows Canonicalization Symmetry

**Date**: 2026-09-10
**Shipment**: 015-S (feature 016-F, tasks 016.001-T..016.008-T)
**Branch**: `feat/015-s-pathsafe-windows-canonicalization-symmetry-and-windows-lint-coverage`
**Mode**: Ship dark-factory execution (DARK_MODE_ACTIVE), operator AFK, merge pre-authorized

## Status at this checkpoint

All 8 tasks + covering feature implemented, tested, and reviewed. About to
proceed to PR creation (Step 5).

## Items completed

- 016.001-T: blocking `golangci-lint run (GOOS=windows)` CI lint step. DONE.
- 016.002-T: `GetFinalPathNameByHandleW` `.Find()` panic-avoidance guard. DONE.
- 016.003-T: internalized `stripUNCPrefix` postcondition inside `canonicalizeReparse`. DONE.
- 016.004-T: `addLongPathPrefix` long-path (`>MAX_PATH`) prefixing before `CreateFile`. DONE (paired with 016.007-T).
- 016.005-T: RED junction-rooted-workspace regression lock (root/candidate asymmetry). DONE (flipped to GREEN by 016.007-T).
- 016.006-T: containment-preservation lock (junction root + inner escaping reparse point). DONE.
- 016.007-T (C2): `NewRoot` now routes through `canonicalizeReparse` instead of `filepath.EvalSymlinks`, closing the 013-S-introduced root/candidate canonicalization asymmetry. DONE (paired with 016.004-T).
- 016.008-T (C3): risk register + doc-comment reconciliation (asymmetry recorded CLOSED, long-path environment-sensitivity recorded as accepted residual, F8/D-5 accuracy verified). DONE.
- 016-F: covering feature marked done.

## Commits on this branch (in order)

1. `3c8f45b` — chore(015-S): claim shipment 015-S and activate covered items
2. `9ed2457` — ci(016.001-T): add blocking GOOS=windows golangci-lint step to lint job
3. `406f16a` — fix(016.002-T): guard GetFinalPathNameByHandleW resolution without a panic path
4. `2e66913` — fix(016.003-T): internalize the stripUNCPrefix postcondition in canonicalizeReparse
5. `e3bb7fb` — chore(015-S): archive completed tasks 016.001-T, 016.002-T, 016.003-T
6. `87979b7` — fix(016.004-T): long-path (>MAX_PATH) prefixing before CreateFile
7. `4e0f8c6` — test(016.005-T): RED junction-rooted workspace regression lock (availability)
8. `d948198` — test(016.006-T): containment-preservation lock for junction-rooted workspaces
9. `6dcab9e` — fix(016.004-T,016.007-T): route NewRoot through canonicalizeReparse (C2)
10. `45895aa` — docs(016.008-T): reconcile risk register and doc comments for the 013-S root/candidate asymmetry (C3)
11. `467afed` — chore(016-F): mark covering feature done
12. `3c6927b` — docs(review-remediation): clarify deliberate stripUNCPrefix retention, fix Op label, CI comment tense, build tag
13. `65d213e` — chore(P-021): capture deferred scope expansion 7ADAC481
14. `6997700` — chore: stash advisory pathsafe hardening follow-ups (6B751D8B)

## Review gate outcome (Step 4.4)

Standard review + 5-persona adversarial pool (Go Reviewer, Security Reviewer,
Correctness Reviewer, Concurrency Reviewer, Maintainability Reviewer,
Architecture Strategist) run against the full `main...HEAD` diff.

**Result: READY_WITH_FOLLOWUPS. Zero P0/P1 across all reviewers.**

In-scope P2/P3 findings remediated directly (commit `3c6927b`):
- Clarified `stripUNCPrefix` retention at both call sites (`root.go`,
  `pathsafe.go`) as a deliberate, plan-authorized (D-2/B2-AC3)
  belt-and-suspenders decision, not redundant dead code (4 reviewers flagged
  this before the clarifying comments were added).
- Fixed `wrapPathError`'s hardcoded `"lstat"` Op label to `"open"` (2 reviewers).
- Reworded the CI lint-step comment to avoid an overclaiming past-tense
  verification date (1 reviewer).
- Added explicit `//go:build windows` tag to `root_windows_test.go` for
  consistency with its sibling file (2 reviewers).
- Re-review pass (Correctness Reviewer) confirmed all four items resolved
  correctly against the plan text, with two more P3 nits (CI comment tense
  precision, D-2/B2-AC3 citation precision) fixed in the same commit.

Out-of-scope finding deferred per P-021 (capture-first, threadless path,
zero prior duplicates found on discovery):
- **Deferred entry 7ADAC481**: Architecture Strategist P1 — `checkSymlinkEscape`'s
  `canonicalizeReparse` error is not normalized through `wrapPathError` the
  way `NewRoot`'s is. Classified out of scope because the 015-S plan's
  C2/AC3 error-surface-preservation requirement is explicitly scoped to
  "NewRoot's failure branches" only; `checkSymlinkEscape`'s error wrapping
  is pre-existing, unmodified-by-this-diff behavior from 013-S/014-S.

Advisory-only follow-ups consolidated into one stash entry (not P-021
deferred-scope, since none require out-of-scope work — genuine future
hardening opportunities):
- **Stash entry 6B751D8B**: `addLongPathPrefix`'s `\\.\` branch reachability
  test-design nuance, MAX_PATH UTF-16-vs-UTF-8 length precision,
  `getFinalPathNameByHandle` buffer-pooling performance opportunity, CI lint
  job runtime-doubling scoping opportunity.

## Reviewed HEAD vs current HEAD

Review was performed against the diff through commit `3c6927b` (last commit
touching source/test/CI files). Two subsequent commits (`65d213e`, `6997700`)
touch only `.backlogit/stash.jsonl` (P-021 capture + follow-up stash
bookkeeping) — confirmed via `git diff 3c6927b..HEAD --stat -- ':!.backlogit'`
returning empty. No re-review required; review coverage is current as of
HEAD `6997700`.

## Quality gates (all green at HEAD `6997700`)

- `go build ./...` — pass
- `go vet ./...` — pass
- `go test ./...` (full repo, all packages) — pass
- `golangci-lint run ./...` (default/linux) — 0 issues
- `golangci-lint run ./...` with `GOOS=windows` — 0 issues
- `python -c "import yaml; yaml.safe_load(...)"` on `ci.yml` — pass
- `actionlint .github/workflows/ci.yml` — pass
- Cross-compile check `GOOS=linux go build ./... && go vet ./...` — pass (done earlier in session)

## P-021 scope discipline

BF5DE670 (deferred stash, no production write caller) was NOT touched,
altered, archived, harvested, or broadened into, per the operator's explicit
DARK_MODE_ACTIVE contract. Confirmed via `git status`/diff review — no file
under this shipment's diff references or modifies anything related to
BF5DE670's scope.

## Next steps

1. Push branch, create PR (Step 5) with `## Local Review Readiness` block.
2. Handle CI failures via fix-ci workflow if any arise.
3. Wait patiently through Copilot review iterations (P-018 hard gate).
4. Classify every Copilot comment under P-021; fix in-scope, defer-capture
   out-of-scope, always reply-before-resolve.
5. Verify merge-commit-only strategy (P-009), obtain final approval
   (pre-authorized per DARK_MODE_ACTIVE contract, but gates must still pass).
6. Merge, then proceed to Step 6 post-merge closure (shipment safe-close,
   operational-closure, compact-context, source-artifact cleanup for
   4104AF54/E428AB46/E4C5413F, post-merge closure PR).
