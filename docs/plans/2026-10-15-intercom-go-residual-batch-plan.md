---
title: "Residual batch: Go gatecheck hardening, scripts/CI/Windows fixes, closure doc accuracy"
description: "Implementation plan for three release units harvested from stash residuals (P-021 C6 deliberation: docs/decisions/2026-10-15-intercom-go-residual-batch-triage-deliberation.md)"
source: "docs/decisions/2026-10-15-intercom-go-residual-batch-triage-deliberation.md"
status: "reviewed"
tags:
  - "gatecheck"
  - "plan"
  - "stage"
---

# Plan: residual batch (three release units)

Source deliberation: `docs/decisions/2026-10-15-intercom-go-residual-batch-triage-deliberation.md`
(decisions D-RB-1..D-RB-9). Execution order: shipment 1, then 2, then 3, each chained behind
046-S by `blocks` edges. Stage creates no code. Every unit below is **new authorized scope** for
the shipment that carries it (the P-021 C1 deferral that kept it out of its origin shipment does
not apply here).

Requires plan hardening: yes

<!-- plan-review-attempt: 2 -->
Plan-review attempt 1 (2026-10-15): FAIL (two P1: V3 premise stale and wrong domain; U9 wrong
file). Revised in place: V3 re-scoped to the Go engine, U9 re-pointed at `scripts/lib/gatecheck-run.sh`
with a retired-arch non-regression, oversized units split (U4, U13, V1, V2, V5), ordering added,
verification list extended, front matter set to `reviewed`. Re-review (attempt 2) verdict: **ADVISORY**,
no P0/P1; P2 items (U9 test seam, U18 ordering behind U15, U10/U11/U18 expected-halt notes, U11 file
list, U6 title) were applied. Independent post-harvest review (gemini-4-argon) findings verified and
applied: V3 split into V3 + V3b (`writepath_golden.json` + `tools/gatecheck/main_test.go`); U10 split
into U10 + U10b (`check_retired_architecture_wrapper_test.go:278`); U18 split into U18 + U18b
(`writepath_test.go`); U7 split into U7 + U7b + U7c with explicit file lists and dependency edges on
U3/U5/U6; U6 target resolved to `runRepoScan` containment finding (`retiredarch.go:609`, leaving
`scanPath` in `select.go:180-196` unchanged); U4/U4b test files and print-site `::error::` ownership
convention specified (`checkignore.go` included in callee convention); A744E584 staged locally in
049-S as W0 + W0b (reverting local Ship hand-edits in `_ship.agent.md` and `shipment-reconcile/SKILL.md`
back to upstream space form) and 9C6C2078 archived as refuted; U15 AC3 and V1 test count (5 tests)
corrected. Proceeding under the Orchestrator's directive; override at staging-PR review.

## Plan Review

decision: ADVISORY
personas: correctness (subagent), scope-boundary (subagent), attempt 1 FAIL then attempt 2 ADVISORY.

Sizing: every unit is `size` (volume) and `complexity` (difficulty) as two independent axes.
2-hour rule: fewer than 3 files, fewer than 5 functions, fewer than 4 test scenarios per task.
If an implementer finds a unit exceeds that, Ship halts and captures a split (P-021 C2); it does
not widen the unit.

## Release unit 1: Gatecheck Go hardening and test hygiene (stash clusters A and B)

Domain: Go code and Go tests under `tools/gatecheck`. Posture default: test-first (red then green)
where behavior changes; characterization-first for pure refactors.

### Sub-epic 1.1: unignore package

| Unit | Stash | Title | Size / Complexity | Files (expected) | Acceptance criteria |
|---|---|---|---|---|---|
| U1 | F5958BBC | Isolate TestExitCode_ExitError from git aliases | S / low | `unignore/git_test.go` | AC1: the test derives its error from `realExitError` (isolated runner), and passes with a `git-*` executable and a git alias planted on PATH. AC2: red-first: the isolation test fails on the old code with the planted alias/executable, then passes. AC3: `go test ./tools/gatecheck/internal/unignore/...` green. |
| U2 | 48EA04C1 | Verify baseline .gitignore entry is a blob | S / medium | `unignore/git.go`, `unignore/git_test.go` | AC1: `rootGitignoreTextAt` rejects a base tree whose `.gitignore` is a directory (test with a tree-typed entry fails red first). AC2: a normal blob baseline is unchanged. AC3: error classification matches the existing failed-show path. |
| U3 | C8827920 | Make the env-test liveness control mandatory | S / low | `retiredarch/select_test.go` | AC1: `TestDefaultGitRunnerIgnoresGitEnv` uses an explicit `noLiveness` field with a stated reason in place of the opt-in `live` bool. AC2: vectors a and b keep their behavior; vectors c-e keep their controls. AC3: omitting the control on a new row fails the test. |
| U4 | F9D52027 (part 1) | Stop doubled ::error:: prefix at unignore.go print sites | S / low | `unignore/unignore.go`, `unignore/unignore_test.go` | Ownership convention: print sites (`unignore.go`) own the single leading `::error::` prefix (stripping any existing leading `::error::` before formatting), while callees in `checks.go`, `git.go`, and `checkignore.go` remain unchanged (`unignore_golden.json` has no error-path scenario captures and is unaffected). AC1: the `runScenario` and sibling print sites in `unignore.go` (~134, 140, 166, 206, 250, 262, 276, 333, 337, 341) emit a single leading `::error::` prefix whether or not the returned error already starts with `::error::`. AC2: red-first tests in `unignore_test.go` assert no `::error::::error::` doubling on `runScenario` / `runSelfTest` / `runCheck` error paths. Depends on U3 and U5 (stash ordering; no file overlap, recorded for fidelity). |
| U4b | F9D52027 (part 2) | Document ::error:: prefix convention and verify checks/git/checkignore non-doubling | S / low | `unignore/checks.go`, `unignore/checks_test.go` | AC1: document the package-wide `::error::` prefix convention in `checks.go` (covering `checks.go`, `git.go`, and `checkignore.go` callees vs `unignore.go` print sites) so the package stays consistent. AC2: tests in `checks_test.go` verify that errors from `checks.go`, `git.go`, and `checkignore.go` (`isIgnored`, `isIgnoredBatch`) format with a single `::error::` prefix and that `git_test.go:248` remains green. Depends on U4. |

### Sub-epic 1.2: retiredarch package

| Unit | Stash | Title | Size / Complexity | Files | Acceptance criteria |
|---|---|---|---|---|---|
| U5 | 4A3851F0 | Narrow Windows symlink skip to errno 1314 | S / low | `retiredarch/retiredarch_test.go` | AC1: `TestRunRepoScan_FinalComponentSymlink_FailsClosed` skips only on errno 1314 and fails on any other `os.Symlink` error, as writepath `u4CreateSymlink` does. AC2: red-first via an injected non-1314 symlink error seam (or a recorded equivalent). Windows behavior is verified on the advisory Windows job and noted, not asserted locally. |
| U6 | A85D25F7 | Print repo-relative path in runRepoScan containment finding | S / low | `retiredarch/retiredarch.go`, `retiredarch/retiredarch_test.go` | AC1: in `runRepoScan` (`retiredarch.go:609`), the containment synthetic finding prints the repo-relative path (`rel`), not `filepath.ToSlash(path)` of the absolute root (SEC-4 as in writepath), while `scanPath` (`select.go:180-196`) is left unchanged so `TestRunRepoScan_DeletedFile_KeepsReadErrorText` (SB-3, ED-11) holds. AC2: a red-first test in `retiredarch_test.go` (`u3MustContain` at line 636) pins the exact repo-relative line and verifies the absolute root is not echoed. |
| U7 | 4995F8C3 (part 1) | Relabel batch unit tags in retiredarch.go and retiredarch_test.go | S / trivial | `retiredarch/retiredarch.go`, `retiredarch/retiredarch_test.go` | AC1: comment-only diff; batch unit tags (such as U3) in `retiredarch.go` and `retiredarch_test.go` are replaced by stash IDs (`990AFA71`). AC2: no non-comment line changes (`git diff -U0` shows comments only; `gofmt -l` clean). Depends on U3, U5, and U6. The writepath half is in U13b (`057.013-T`). |
| U7b | 4995F8C3 (part 2) | Relabel batch unit tags in retiredarch select_test and tomlprimary | S / trivial | `retiredarch/select_test.go`, `retiredarch/tomlprimary.go` | AC1: comment-only diff; batch unit tags in `select_test.go` (U1, U2, U4, U6) and `tomlprimary.go` (U1, U2) are replaced by stash IDs (`4537B2F6`, `D7BF9F74`, `9FC28DB9`). AC2: `git diff -U0` shows comment lines only; `gofmt -l` clean. Depends on U3 and U7. |
| U7c | 4995F8C3 (part 3) | Relabel batch unit tags in retiredarch tomlprimary_test and pin_test | S / trivial | `retiredarch/tomlprimary_test.go`, `retiredarch/pin_test.go` | AC1: comment-only diff; batch unit tags in `tomlprimary_test.go` (U1, U2) and `pin_test.go` (U5, U6) are replaced by stash IDs (`9FC28DB9`, `D7BF9F74`); `pin.go:228` (`040-S U5`) is already disambiguated and untouched. AC2: `git diff -U0` shows comment lines only; `gofmt -l` clean. Depends on U7b. |

### Sub-epic 1.3: write-path wrapper

| Unit | Stash | Title | Size / Complexity | Files | Acceptance criteria |
|---|---|---|---|---|---|
| U8 | 815C2B28 | Add readonly GIT_CONFIG_GLOBAL discriminating row | S / low | `tools/gatecheck/check_write_path_wrapper_test.go` | AC1: a shared-table row with an unexported readonly `GIT_CONFIG_GLOBAL` expects exit 2 and `cannot set GIT_CONFIG_GLOBAL`. AC2: the row fails if the refusal arm is removed (mutation check recorded). |
| U9 | BEFE6DF2 | Fail closed on a failed compgen listing | S / medium | `scripts/lib/gatecheck-run.sh` (the shared `gatecheck_repo_root`; the `done < <(compgen -e)` loop), `check_write_path_wrapper_test.go` | AC1: `gatecheck_repo_root` captures the `compgen -e` listing before iterating; a failed listing exits nonzero. AC2: a red-first test forces the failure through a named seam (for example a shell function shadowing `compgen` exported into the wrapper subprocess, or an env override the test sets); the seam is test-only and documented in the test. AC3: because the helper is shared with `check-retired-architecture.sh`, a non-regression scenario for the retired-architecture wrapper is run and recorded. Depends on U8. |
| U10 | EF11062A (part 1) | Move wrong-tree harness to its own file | S / medium | `tools/gatecheck/check_write_path_wrapper_test.go`, `tools/gatecheck/wrapper_harness_test.go` | AC1: `installGoShim`, `runWrongTreeRow`, `wrongTreeRows` and helpers moved to `wrapper_harness_test.go` with no assertion or coverage change. AC2: callers in `check_write_path_wrapper_test.go` use the directory return (with a transitional wrapper if needed until U10b updates `check_retired_architecture_wrapper_test.go:278`). AC3: the per-subtest install cadence and `LegitimateModesReachBuild` use one stated cadence (install once per test or once per subtest), recorded in a comment. Depends on U8 and U9. |
| U10b | EF11062A (part 2) | Update retiredarch wrapper test to use installGoShim directory return | S / low | `tools/gatecheck/wrapper_harness_test.go`, `tools/gatecheck/check_retired_architecture_wrapper_test.go` | AC1: `installGoShim` in `wrapper_harness_test.go` returns its directory directly, and `check_retired_architecture_wrapper_test.go:278` is updated to use `installGoShim(t, bash)` instead of `filepath.Dir(installGoShim(t, bash))`. AC2: all callers in both wrapper test files compile and pass with no coverage change. Depends on U10. |
| U11 | 63BCA7C6 | Correct wrapper header text and pin PS4 and LC_ALL | S / low | `scripts/check-write-path-precondition.sh` (header), `tools/gatecheck/check_write_path_wrapper_test.go` (xtrace diagnostic filter and row name; `scripts/lib/gatecheck-run.sh` has no xtrace filter) | AC1: the header no longer describes a usage branch and first-argument forwarding (the code forwards every argument). AC2: PS4 and LC_ALL are pinned in the xtrace diagnostic filter. AC3: the readonly row is renamed to caller-precondition ordering. Depends on U9, U10, and U10b. |

### Sub-epic 1.4: writepath package

| Unit | Stash | Title | Size / Complexity | Files | Acceptance criteria |
|---|---|---|---|---|---|
| U12 | 9EC5C536 | Add no-ceiling control and exact BadListing expectations | S / medium | `writepath/writepath_test.go` | AC1: `TestDefaultGitRunner_NonRepository_FailsClosed` has a no-ceiling control proving the planted parent repo is discovered when unbounded; its doc no longer overstates. AC2: non-control `BadListing` rows use exact expectations, not `Contains` with an empty prefix. AC3: item (c) (8.3 short names) is recorded as a verification note for the Windows job, not asserted locally. |
| U13 | 65AE88CA (part 1) | Consolidate writepath fixture and link helpers | S / medium | `writepath/writepath_test.go` | AC1: one shared fixture-repo builder for the two non-ASCII fixture tests; one argument order for the link helpers; one shared git-environment scrubber name. AC2: no behavior change; EX-1 and EX-5 counts unchanged. Depends on U12. |
| U13b | 65AE88CA (part 2) + 4995F8C3 writepath half | Rename writepath test helpers and use a direct junction call | S / medium | `writepath/writepath_test.go`, `writepath/writepath.go` (comments where unit tags appear) | AC1: descriptive names replace the `u4` prefix and batch unit tags (4995F8C3 writepath half). AC2: junction creation uses a direct call, not `cmd /c mklink` (metacharacters in a TEMP path re-parse under cmd); a test uses a TEMP path with metacharacters. AC3: no behavior change; EX-1 and EX-5 counts unchanged. Depends on U13. |
| U14 | 85D5A4D5 | Route static single-line outputs through logLine | S / medium | `writepath/writepath.go`, `writepath/writepath_test.go` | AC1: the ED-7 empty-selection line and fixture constant lines use `logLine`. AC2: multi-line usage and banner constants are NOT wrapped as one unit; newline bytes unchanged. AC3: output bytes are identical (golden unchanged). Depends on U13b (same package; serializes edits to `writepath.go`). |
| U15 | 054EAAB0 | Keep accumulated PASS output on self-test failure returns | S / medium | `writepath/writepath.go`, `writepath/writepath_test.go` | AC1: `runFixtureSelfTest` failure returns (`scanFile` error at ~1188, `containedRegularFile` refusal at ~1184, and `len(failures) > 0` at ~1209) include the accumulated `Stdout`. AC2: a red-first test in `writepath_test.go` pins a failing self-test log that retains earlier PASS lines on stdout alongside the failure on stderr. AC3: existing success-path goldens (`writepath_golden.json` has no `runFixtureSelfTest` failure-path goldens) and single-fixture failure tests in `writepath_unix_test.go` remain unchanged. Depends on U14 (serializes edits to `writepath.go`). |
| U16 | A6535A74 | Writepath test hygiene | S / low | `writepath/writepath_test.go`, `writepath/writepath_unix_test.go` | AC1: `unixWriteFixture` duplication with `u4WriteText` removed. AC2: the `tc.name == selector` branch keyed on a field, so renaming a row cannot silently drop the assertion. AC3: `scannedPathsFromListing` placement considered; signature unchanged. Depends on U13b. |
| U17 | AF0C83BA | Annotate outcome-matrix wording as encounter order | S / low | the `docs/plans/` file containing the "repo-scan outcome matrix" (locate by grep; expected a 044-S/054-F era plan), `writepath/writepath.go` (`escapeLogText` doc comment) | AC1: the matrix receives a dated annotation (not a rewrite of the historical record) stating listing faults are reported in encounter order (G-3), correcting the row 2 before row 3 wording. AC2: the `escapeLogText` doc sentence about a retiredarch copy is checked and corrected. Docs/comment only; `git diff` limited to comments and the dated annotation. Depends on U16. |
| U18 | A1327CF4 (part 1) | Fix writepath.go and writepath_unix_test.go comment accuracy | S / low | `writepath/writepath.go`, `writepath/writepath_unix_test.go` | AC1: the three comment inaccuracies in these two files are corrected (`containedRegularFile` doc ~944-960 on missing components returning true early, TOCTOU window wording ~960 and ~1017 on directory components swapped after Lstat, and unix build-tag comment ~105 in `writepath_unix_test.go`). AC2: comment-only diff (`git diff -U0` shows comments only). Depends on U15 and U16. |
| U18b | A1327CF4 (part 2) | Fix writepath_test.go comment accuracy | S / low | `writepath/writepath_test.go` | AC1: the three comment inaccuracies in `writepath_test.go` are corrected (outcome-order header rows ~713, vector header (b)/(c) ~1306, and (d1)-(d3) citations ~1772-1774). AC2: comment-only diff (`git diff -U0` shows comments only). Depends on U15, U16, and U18. |

## Release unit 2: Scripts, CI and Windows test fixes (cluster D)

Domain: mixed, so tasks stay single-domain each.

| Unit | Stash | Title | Size / Complexity | Domain | Acceptance criteria |
|---|---|---|---|---|---|
| V1 | 96B308D1 + 5CC5532E item 1 (part 1) | Add a bash-safe path helper and fix TestGatecheckCleanup tests | S / medium | tests (`tools/gatecheck/runner_test.go`) | AC1: a shared test helper converts a Windows temp path to a bash-safe form (forward slashes or `cygpath`), or the caller skips with a clear reason where bash semantics are required. AC2: all five `TestGatecheckCleanup_*` tests in `runner_test.go` (lines 612-719: `RemovesTmpOnNormalExit`, `RemovesTmpOnNonZeroExit`, `Idempotent_CalledTwiceIsHarmless`, `HarmlessWhenNeverBuilt`, `WrapperOwnCleanupComposes` — correcting stash 96B308D1's count of three) are verified and pass on a Windows host (verified on the advisory Windows job or a Windows host and noted with host, command and result). AC3: Linux behavior unchanged. |
| V1b | 5CC5532E item 1 (part 2) | Apply the bash-safe helper to root-package wrapper tests | S / medium | tests | AC1: the root `tools/gatecheck` package tests that invoke the bash wrapper (55 test names, exit 127 on Windows) route the path through the V1 helper, ideally at one shared invocation point. AC2: if more than two files need edits, Ship halts and captures a split by wrapper (P-021 C2) instead of widening. AC3: Linux unchanged. Depends on V1. 5CC5532E item 2 is covered by 046-S and is out of scope. |
| V2 | B4D38D63 (part 1) | Harden filename-test parsing and matching | S / medium | tests | AC1: in `tests/integration/operational_closure_post_merge_filename_test.go`, the unrecognized-placeholder fallback is fatal; the `## Output` heading match is anchored to line start; multiple post-merge-mentioning lines are disambiguated; the id placeholder regex class is tightened (or relaxed) with the stated reason. AC2: red-first tests for the fatal fallback and the anchor; the test still passes against the current docs. |
| V2b | B4D38D63 (part 2) | Add regression guard and tidy filename-test helpers | S / low | tests | AC1: a guard asserts the generic pre-merge/post-deploy Output form stays documented. AC2: the "never hardcodes" comment is softened regarding the `*-post-merge-closure.md` glob; the empty-glob hard-fatal decision is recorded as a one-line comment (keep or relax, with reason). AC3: shared integration-package helper names are prefixed. Depends on V2. |
| V3 | 8E18CCF5 (part 1) | Emit write-path gate mode notice from the Go engine | S / medium | Go code (`writepath/writepath.go`, `writepath/writepath_test.go`) | Premise corrected by plan review: `check-write-path-precondition.sh` has no mode dispatch and `check-retired-architecture.sh` no longer prints the notice (it moved into `internal/retiredarch.Run`). AC1: `internal/writepath.Run` emits `::notice::write-path gate mode=X` to stdout for `repo` (`flag == ""`), `self-test` and `self-test-integrity`, mirroring the retiredarch engine. AC2: red-first unit tests in `writepath_test.go` (including `TestRun_RepoMode_CurrentTreeIsClean`) pin the notice line per mode on stdout. Depends on U15 via the shipment chain (shipment 1 completes first). |
| V3b | 8E18CCF5 (part 2) | Update writepath golden and CLI dispatch test for mode notices | S / low | tests/golden (`writepath/testdata/writepath_golden.json`, `tools/gatecheck/main_test.go`) | AC1: `writepath_golden.json` `stream_captures` (`self_test`, `self_test_integrity`, `repo_from_subdirectory`) include the new `::notice::write-path gate mode=X\n` line on stdout (and any stream fingerprint in `writepath_test.go` is updated in V3 or accounted for). AC2: `tools/gatecheck/main_test.go` `TestRun_WritePath_DispatchesToEngine` (~263-265) asserts `stdout` equals `"::notice::write-path gate mode=repo\n"` (mirroring `TestRun_RetiredArch_DispatchesToEngine`). AC3: `go test ./tools/gatecheck/...` green. Depends on V3. |
| V4 | 56B16321 | Record --self-test as manual-only | S / low | docs/comments | AC1: per D-RB-6, the script header and the CI step comment state that `--self-test` is local/manual-only and which CI steps cover each sub-mode. AC2: the implementer confirms the coverage claim by reading `ci.yml` and halts (captures) if a sub-mode is uncovered. No new CI step. |
| V5 | 06CE25E5 (part 1) | Update stale pin example in the maintenance note | S / trivial | docs | AC1: the example block in `docs/decisions/2026-09-04-go-toolchain-pin-maintenance-note.md` no longer shows go1.26.5 as current, and carries a dated annotation rather than a silent rewrite. |
| V5b | 06CE25E5 (part 2) | Annotate historical Go 1.26.5 comments | S / trivial | Go comments | AC1: the Go 1.26.5 mentions in `internal/pathsafe/root.go` (~96, ~174) are annotated historical. AC2: comment-only diff (`gofmt -l` clean; no non-comment line changes). |

## Release unit 3: Closure contract design-doc and test accuracy (cluster C, locally editable)

| Unit | Stash | Title | Size / Complexity | Domain | Acceptance criteria |
|---|---|---|---|---|---|
| W0 | A744E584 (part 1) | Revert local cascade-close --message=/--author= deviation to space form | S / low | docs (`.github/agents/_ship.agent.md`, `.github/skills/shipment-reconcile/SKILL.md`) | AC1: revert the local Ship hand-edits (commits `2749d5d` and `b5abe5f`) in `_ship.agent.md` (~line 828) and `shipment-reconcile/SKILL.md` (~line 636) that introduced `--message=<value>` and `--author=<value>` (rejected by `autoharness shipment cascade-close` with exit 2) back to the upstream space-separated form (`--message <value>` and `--author <value>`), noting that a future `autoharness install`/`tune` re-render would also restore it. AC2: preserve the argument-safety rules (argument vector with no shell interpolation, values must not begin with `-`, no newlines or control characters). |
| W0b | A744E584 (part 2) | Add contract-test guard against --message=/--author= equals form | S / low | tests (`tests/integration/ship_feature_completion_contract_test.go`) | AC1: add a contract-test row in `ship_feature_completion_contract_test.go` asserting that neither `_ship.agent.md` nor `shipment-reconcile/SKILL.md` contains `--message=` or `--author=` (red against the pre-W0 text, green after W0). AC2: `go test ./tests/integration -run TestShipFeatureCompletionContract33Rows` green. Depends on W0. |
| W1 | 86381564 | Record replica-test deletion in the closure matrix doc | S / low | docs | AC1: the matrix rows near lines 49-50, 259, 269-270, 291, 315 and 324 record the deletion under 053.013-T instead of listing the files as present/deferred. AC2: no other content changes. |
| W2 | 762A4FCA | Derive smoke-test exit from selected_close_path | S / medium | tests | AC1: `TestCascadeCloseClassifyOnlySmoke` parses `selected_close_path` and derives the expected exit (cascade -> 0, safe_close -> 3), or gates the exit-0 assertion on a verified engine verdict. AC2: a newer or UNVERIFIED engine on PATH no longer fails the test. |
| W3 | BAB25951 | Reword the IV-1..IV-5 evidence comment | S / trivial | tests (comment) | AC1: the comment at `ship_feature_completion_contract_test.go` lines 31-32 names the smoke test's actual scope and the invariants without executed coverage. AC2: comment-only diff. Depends on W0b (serializes edits to `ship_feature_completion_contract_test.go`). |

Archived as refuted premise (not harvested): 9C6C2078, see D-RB-8. Routed upstream (not harvested): 9B7E94E9,
D5D103A1, 0606E72F, 92708A06, 8134B7B7, F2759510, 8E1E2D80, 724FA2FB (D-RB-7; stash 40BCCD69).

## Dependency graph

* Shipment chain: 046-S -> shipment 1 (047-S) -> shipment 2 (048-S) -> shipment 3 (049-S) (`blocks`).
* Within shipment 1 (047-S): U8 -> U9 -> U10 -> U10b -> U11; U12 -> U13 -> U13b -> U14 -> U15 -> U18 -> U18b (also U16 -> U18 and U16 -> U18b); U13b -> U16 -> U17; U3 and U5 -> U4 -> U4b; U3, U5, and U6 -> U7 -> U7b -> U7c.
* Within shipment 2 (048-S): V1 -> V1b; V2 -> V2b; V3 -> V3b (after shipment 1 for `writepath.go` serialization); V4,
  V5, V5b independent.
* Within shipment 3 (049-S): W0 -> W0b -> W3; W1 and W2 independent (W0/W0b placed first in the manifest).
* Counts: shipment 1 (047-S) is 18 entries -> 24 tasks (U1-U18 plus U4b, U7b, U7c, U10b, U13b, U18b); shipment 2 (048-S) is 6 entries
  -> 9 tasks (V1, V1b, V2, V2b, V3, V3b, V4, V5, V5b); shipment 3 (049-S) is 4 entries -> 5 tasks (W0, W0b, W1-W3).
  Total 28 entries -> 38 tasks.

## Plan Hardening

Risk signals: U9 (fail-closed script behavior), U15 (retains accumulated PASS lines on self-test failure returns),
U14 (output chokepoint), V1 (broad Windows test surface), V3/V3b (mode notice changes writepath goldens and CLI dispatch test), W0/W0b (reverts local deviation in rendered closure instructions).

| Risk | Mitigation / verification |
|---|---|
| Gate semantics drift while "hardening" | Each behavior-changing unit is test-first with a recorded red run; refactor units (U10, U10b, U13, U16) assert unchanged counts and goldens. |
| Golden churn hides a regression (V3b) | Only the named golden fields (`stream_captures` in `writepath_golden.json` for V3b) may change; U4 and U15 do not alter existing success-path goldens. |
| Windows-only behavior cannot be verified locally on Linux CI | The Windows job is advisory (`WINDOWS_GATE_REQUIRED` unset). Units V1, U5 and U12(c) record Linux CI as the regression guard and Windows as a manual/advisory verification, never an unrecorded claim. |
| Shipment size (shipment 1 has 24 tasks) | One owner and one package family; every task is bounded to <=2 files; the Ship circuit breaker applies. If review shows an over-wide unit, Ship captures it rather than expanding (P-021). |
| Rollback | Every unit is a separate commit; revert is per-unit. No data migration. |
| Concurrent Ship activity | 046-S is queued; these shipments carry `blocks` edges and are never claimed before 046-S ships. |

Verification commands (Ship runs, not Stage): `go build ./...`, `go vet ./...`,
`GOOS=windows go vet ./tools/gatecheck/...`, `gofmt -l .`, `go test -race ./tools/gatecheck/...`,
`bash scripts/check-write-path-precondition.sh` modes, `actionlint` when a workflow file changes,
`golangci-lint run` as configured. Comment-only units (U7, U17, U18, V4, V5, V5b, W3) are checked
with `git diff -U0` limited to comment lines. Windows-only claims (U5, V1, V1b, U12(c)) are
recorded in the task closure note as host, command and result, never asserted from Linux CI.
