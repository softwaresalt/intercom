---
title: "Gatecheck wrapper test CI reliability and Git Bash sentinel stability (51C7CD42, AD132ACF)"
description: "Deliberation for Option B (51C7CD42, AD132ACF): eliminate unshimmed go build timeout flake in TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst and fix MSYS2 /tmp sentinel path aliasing in installGoShim / TestCheckWritePathWrapper_LegitimateModesReachBuild."
date: 2026-10-30
topic: "Gatecheck wrapper test CI reliability (51C7CD42, AD132ACF)"
depth: "standard"
decision_status: "decided"
promoted_to: "plan"
stash_ids:
  - "51C7CD42"
  - "AD132ACF"
linked_artifacts:
  - "docs/plans/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-plan.md"
tags:
  - "ci"
  - "gatecheck"
  - "testing"
  - "windows"
  - "git-bash"
---

## Problem Frame

Option B (`run_id=option-b-ci-reliability-2026-10-30`, `DARK_MODE_ACTIVE` activated at `2026-10-30T02:11:46Z`) scopes exactly two deferred-scope-expansion bug entries affecting `tools/gatecheck` wrapper tests in CI and local Git Bash runs:

* **`51C7CD42`** (`priority: high`, `kind: bug`): `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` (`tools/gatecheck/check_retired_architecture_wrapper_test.go`, lines 103-156) waits on a hard-coded `time.After(20 * time.Second)` deadline while `scripts/check-retired-architecture.sh` runs a real `go build` of `./tools/gatecheck` in each subtest. Under CI `-race` load (`go test -race -mod=readonly ./...`), the build took 21-25 seconds and failed three consecutive runs on PR 212 (versus ~12 seconds locally and passing on PR 210). When the 20-second deadline expires, the `select` block silently kills the subprocess without calling `t.Fatalf`, causing the subsequent trace assertion (`+ gatecheck_invoke retired-arch ...`) to fail as if argument forwarding were broken.
* **`AD132ACF`** (`priority: low`, `kind: bug`): `TestCheckWritePathWrapper_LegitimateModesReachBuild` (`tools/gatecheck/check_write_path_wrapper_test.go`, lines 45-113) is intermittently flaky under Git Bash on Windows (`missing generated-binary-calls.log sentinel in some runs`, following an earlier `/tmp/go` shim self-check spelling mismatch that was fixed in `061.003-T`).

### Safety Mode and Strict-Safety Classification

Per Constitutional Principle VIII and `.github/instructions/strict-safety.instructions.md`, this session operates under **`investigate-first`** (for root-cause verification of both test flakes before defining acceptance criteria) and **`freeze-scope`** (constraining planning and harvest strictly to `51C7CD42` and `AD132ACF` wrapper test reliability plus required Stage artifacts, excluding Option A write approvals, upstream harness templates, `80E3F89C(b)/(d)/(e)` production gatecheck items, and unrelated wrapper tests).

* **`ProposedAction`**: Scoped Stage deliberation, implementation planning, plan review, backlog harvest, shipment assembly, and consumed-stash archival for `51C7CD42` and `AD132ACF` on branch `chore/stage-gatecheck-wrapper-test-ci-reliability`
* **`ActionRisk`**: `moderate` (non-destructive planning and backlog state mutations; downstream execution touches only `tools/gatecheck/*_test.go`)
* **`approval_required`**: `false` for authorized non-destructive Stage work; `true` for publishing/merging the Stage artifact branch PR and for any downstream Ship merge/admin/destructive operations (`merge_approval_pre_authorized=false`, `admin_fallback_pre_authorized=false`)
* **`ActionResult`**: `approved` (for Stage pipeline execution within declared scope)

### P-021 C5 and C6 Triage and Provenance Audit

Both selected entries carry the literal `DEFERRED SCOPE EXPANSION` token and `requires deliberation: yes`, triggering mandatory Step 2 deliberation before planning (P-021 C6) as well as the two independent Step 1 triage obligations (Unconditional Duplicate Detection [Trigger A] and Late-Identifier Reconciliation [Trigger B]):

| Stash ID | Originating Source Refs | Trigger A: Unconditional Duplicate Scan | Trigger B: Late-Identifier Reconciliation |
|---|---|---|---|
| `51C7CD42` | `task: 078.002-T`, `feature: 078-F`, `shipment: 071-S`, `PR: 212` (merge commit `67d3f5c03e893dde348ecdec08c6fd7b4c69a686`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned all 145 active entries in `.backlogit/stash.jsonl`, `.backlogit/archive/stash.jsonl`, and `.backlogit/queue/`; no duplicate entry describes `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` or its 20-second build wait. | **No late identifier found**: `PR 212` was recorded at capture; searched `docs/closure/071-S-078-F-post-merge-closure.md` and `docs/memory/compacted/2026-10-29-071-s-ship-compacted.md` by join key `51C7CD42` — no review thread ever existed for this CI runner timing failure, so `review-thread: N/A` stands as the truthful terminal record (no-op). |
| `AD132ACF` | `task: 061.001-T`, `feature: 061-F`, `shipment: 053-S`, `PR: 165` (reconciled from `PR: N/A`; merge commit `222e6e29321d5fcfca23da99170cfe03e6c09b82`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned all 145 active stash entries, archived stash entries, and queue items; no duplicate entry describes the missing `generated-binary-calls.log` sentinel in `TestCheckWritePathWrapper_LegitimateModesReachBuild`. Related `80E3F89C(a)` (`sameBashFile` identity check) and `80E3F89C(c)` (junction normalization) were already promoted and shipped in `053-S` (`061.003-T`, `061.004-T`). | **Reconciled `PR: 165`; no late review-thread identifier found**: Searched `docs/closure/053-S-061-F-post-merge-closure.md` and `docs/memory/compacted/2026-10-18-053-s-ship-compacted.md` by join key `AD132ACF`; recovered `PR 165` from the `053-S` closure record and updated `AD132ACF` in place via `backlogit stash edit` after the Step 1.9 branch gate passed. `review-thread: N/A` has no late identifier found (local Git Bash test observation) and stands as the truthful terminal record. |

## Research Findings

### Prior Learnings Consulted (`docs/compound/`)

The `Learnings Researcher` pass (`confidence: high`, 37 compound documents searched) surfaced five directly relevant records:

* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`: Documents MSYS2 / Git Bash `/tmp` mount aliasing under Windows `TEMP`, `toBashPath` `/c/...` normalization, and filesystem identity checks (`sameBashFile` / `os.SameFile`).
* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`: Documents natively launchable fake `go` / binary shims in `tools/gatecheck` tests that write sentinel marker files to prove execution.
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`: Documents heavy CPU and filesystem contention in `tools/gatecheck` tests under `go test -race` parallel execution.
* `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`: Documents that CI's `test` job runs `go test -race -mod=readonly ./...` with CGO enabled across the workspace, magnifying subprocess `go build` latency.
* `docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md`: Documents wrapper script `--root` dispatch and environment scrubbing conventions in `tools/gatecheck`.

### Root-Cause Analysis (`investigate-first`)

#### 1. `51C7CD42` — Why `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` Times Out in CI

In `tools/gatecheck/check_retired_architecture_wrapper_test.go` (lines 93-156):

* The function comment (lines 97-102) states: *"This is verified via a `bash -x` trace of the actual `gatecheck_invoke` call line rather than waiting on a real `go build`, so a slow/unavailable Go toolchain cannot make this test flaky; any resulting build failure happens strictly after the point this test inspects."*
* However, in `scripts/check-retired-architecture.sh` (lines 207-218), `gatecheck_build` runs **before** `gatecheck_invoke retired-arch`, and if `gatecheck_build` returns non-zero the wrapper exits before `gatecheck_invoke` is ever reached.
* Unlike `TestCheckWritePathWrapper_LegitimateModesReachBuild` (`check_write_path_wrapper_test.go`, line 48) and `TestCheckRetiredArchitectureWrapper_WrongTreeRootAnchoring` (`check_retired_architecture_wrapper_test.go`, line 278), `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` does **not** call `installGoShim(t, bash)`. Each of its two subtests (`--self-test` and `--self-test-integrity`) therefore runs a full unshimmed `go -C "${GATECHECK_SRC}" build -trimpath -o "${GATECHECK_BIN}" ./tools/gatecheck` subprocess followed by the real compiled `gatecheck retired-arch` scan.
* Under `go test -race -mod=readonly ./...` in CI, compiling `./tools/gatecheck` inside the wrapper subprocess takes 21-25 seconds, exceeding `time.After(20 * time.Second)`.
* Worse, the timeout branch in `select` (lines 139-144) kills `cmd.Process` and drains `done` without calling `t.Fatalf`, so a build timeout falls through to `if !strings.Contains(out, "+ "+tc.wantInvoke)` and reports a misleading missing-invocation failure.

#### 2. `AD132ACF` — Why `TestCheckWritePathWrapper_LegitimateModesReachBuild` Loses `generated-binary-calls.log` Under Git Bash

In `tools/gatecheck/wrapper_harness_test.go` (lines 22-127) and `tools/gatecheck/check_write_path_wrapper_test.go` (lines 45-113):

* `installGoShim(t, bash)` first calls `dir := workspaceTempDir(t)`, which sets `TMPDIR`, `TMP`, and `TEMP` to `absoluteDir` (`dir`). It then calls `bashDir := resolveBashPath(t, bash, dir)` (line 60) and sets:
  * `GATECHECK_SENTINEL_LOG = path.Join(bashDir, "generated-binary-calls.log")`
  * `GATECHECK_ARGS_LOG = path.Join(bashDir, "generated-binary-args.log")`
* Under Git for Windows (MSYS2), `/etc/fstab` mounts `/tmp` as `usertemp`, bound to the Windows `TMP`/`TEMP` of whichever MSYS2 process initializes the shared `msys-2.0.dll` mount table. When `resolveBashPath(t, bash, dir)` runs `cygpath -u "$directory"` and `cd "$directory" && pwd` with `TMP` set to `dir` and no competing MSYS2 process holding the mount table, MSYS2 aliases `dir` to `/tmp`, so `bashDir` is returned as `"/tmp"` and `GATECHECK_SENTINEL_LOG` is set to `"/tmp/generated-binary-calls.log"`.
* When `TestCheckWritePathWrapper_LegitimateModesReachBuild` subsequently runs `runWritePathWrapperWithTrace` during a parallel `go test ./...` run where another concurrent MSYS2 process (spawned by another test with the host's default `TMP`/`TEMP` or a different `workspaceTempDir`) holds the shared MSYS2 mount table, `/tmp` in that child MSYS2 session resolves to the other process's `TMP` directory rather than `shimDir`.
* Meanwhile, inside the `go` shim (`wrapper_harness_test.go`, lines 65-79), `go-calls.log` and `build-output-path.log` are written to `${shim_dir}` derived from `$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)`, so both files are created inside `shimDir`. The generated sentinel binary then executes `printf 'invoked\n' >> "${GATECHECK_SENTINEL_LOG:?}"` and exits `0`, writing `invoked` to `/tmp/generated-binary-calls.log` in the other `TMP` directory instead of `shimDir`. When Go reads `filepath.Join(filepath.Dir(goShimLog), "generated-binary-calls.log")`, the file is missing in `shimDir`.
* Note the existing prior art in the same file: `runWrongTreeRow` (`wrapper_harness_test.go`, lines 247-258) explicitly noted this `/tmp` vs `/c/...` aliasing hazard (*"Each row's logs are addressed by their `/c/...` form instead"*) and set `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG` using `toBashPath(rowDir)` (`/c/...`), which uses MSYS2's static cygdrive mount and never depends on `/tmp`. While `061.003-T` (`sameBashFile`) fixed the `command -v go` check at line 120 to compare by `test -ef` file identity, lines 62-63 of `installGoShim` still used `bashDir` (`/tmp`), and lines 74-78 of `installGoShim` (`case "$6" in "${managed_tmpdir}"/*)`) still compared `$6` against `${managed_tmpdir}` by string prefix alone without a `test -ef` directory-identity fallback.

## Options Evaluated

### Option 1: Hermetic Go-Shim Wiring for `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` and `/c/...` + Identity Hardening in `installGoShim` (Recommended)

Address each bug at its verified root cause in two independent, test-domain tasks under a single chore-classified release unit (`080-F` / shipment `073-S`):

1. **Unit 1 (`51C7CD42`, `priority: high`, `size: S`, `complexity: low`) — `tools/gatecheck/check_retired_architecture_wrapper_test.go`**:
   * Call `installGoShim(t, bash)` in `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` so `gatecheck_build` runs the fast hermetic `go` shim (~5-10 ms) rather than compiling `./tools/gatecheck` twice under `go test -race`, fulfilling the function's documented intent.
   * Reset the shim log files (`go-calls.log`, `build-output-path.log`, `generated-binary-calls.log`, `generated-binary-args.log`) at the start of each subtest and pin `PS4=+ ` and `LC_ALL=C` in `cmd.Env` (matching `runWritePathWrapperInDir`).
   * In the `select` timeout arm, fail explicitly via `t.Fatalf("wrapper timed out after 20s (kill err=%v, wait err=%v; out=%q)", killErr, waitErr, combined.String())` instead of silently falling through to the trace assertion, and assert that the shimmed wrapper run exits with code `0` (`runErr == nil`).
2. **Unit 2 (`AD132ACF`, `priority: low`, `size: S`, `complexity: medium`) — `tools/gatecheck/wrapper_harness_test.go` and `tools/gatecheck/check_write_path_wrapper_test.go`**:
   * In `installGoShim` (`tools/gatecheck/wrapper_harness_test.go`), set `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG` using `toBashPath(dir)` (`path.Join(toBashPath(dir), "generated-binary-calls.log")` and `path.Join(toBashPath(dir), "generated-binary-args.log")`), matching `runWrongTreeRow` so sentinel logs never route through the volatile MSYS2 `/tmp` mount alias.
   * In `installGoShim`'s generated `go` shim script, keep the `case "$6" in "${managed_tmpdir}"/*)` fast path and add an identity fallback `if [ ! "$(dirname "$(dirname "$6")")" -ef "${managed_tmpdir}" ]; then unexpected ...; fi` so `/tmp/tmp.XXXX` and `/c/...` spellings of the same `TMPDIR` pass while any output path outside `TMPDIR` still fails with exit code `97`. Also normalize the `go` shim script via `normalizeScript(shim)` before writing to disk.
   * In `TestCheckWritePathWrapper_LegitimateModesReachBuild` (`tools/gatecheck/check_write_path_wrapper_test.go`), reset `generated-binary-args.log` alongside `go-calls.log`, `build-output-path.log`, and `generated-binary-calls.log` before each subtest, include `(trace=%q)` in the `t.Fatalf` messages when reading `go-calls.log`, `build-output-path.log`, or `generated-binary-calls.log`, and add a regression test (`TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity`) verifying that `GATECHECK_SENTINEL_LOG` / `GATECHECK_ARGS_LOG` use `toBashPath(shimDir)` and that the `go` shim accepts `/tmp` vs `/c/...` spellings of `TMPDIR` by identity while still rejecting an outside output path with exit code `97`.

* **Pros**: Eliminates both root causes deterministically; keeps the two fixes in disjoint files (`check_retired_architecture_wrapper_test.go` vs `wrapper_harness_test.go` + `check_write_path_wrapper_test.go`) with no artificial inter-task dependency; speeds up `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` from 12-25s to <0.5s.
* **Cons**: Requires careful test-first red/green verification on both tasks during Ship execution.
* **Effort estimate**: Low (`2 tasks × S`).
* **Fit**: Exact match to scope, root causes, and 2-hour granularity constraints.

### Option 2: Increase Subprocess Timeouts Only (`20s` to `60s`)

Raise the `20 * time.Second` timeout in `check_retired_architecture_wrapper_test.go` and `check_write_path_wrapper_test.go` to `60 * time.Second` without wiring `installGoShim` or changing `GATECHECK_SENTINEL_LOG`.

* **Pros**: Minimal diff in test files.
* **Cons**: Does not fix `AD132ACF` at all (since `missing generated-binary-calls.log` is caused by MSYS2 `/tmp` mount aliasing, not a 20s timeout); leaves `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` compiling `./tools/gatecheck` twice under `-race` in CI and wasting 25+ seconds of runner time; leaves silent timeout swallowing in `select`.
* **Effort estimate**: Low.
* **Fit**: Poor — fails to fix `AD132ACF` and leaves unnecessary compilation load in CI.

### Option 3: Broad Refactor of All Gatecheck Wrapper Tests and Shell Scripts

Refactor `check_retired_architecture_wrapper_test.go`, `check_unignore_wrapper_test.go`, `check_write_path_wrapper_test.go`, `wrapper_harness_test.go`, and `scripts/lib/gatecheck-run.sh` into a single unified test harness and also address residual items in `80E3F89C(b)/(d)/(e)`.

* **Pros**: Unifies every wrapper test file at once.
* **Cons**: Violates `freeze-scope` and P-021 C1/C6 scope boundaries by pulling unauthorized stash items (`80E3F89C(b)/(d)/(e)`) and adjacent test files (`check_unignore_wrapper_test.go`) into Option B.
* **Effort estimate**: High.
* **Fit**: Disqualified by explicit operator scope freeze.

## Trade-off Comparison

| Criterion | Option 1 (Targeted Shim + Path Identity Fixes) | Option 2 (Timeout Bump Only) | Option 3 (Broad Wrapper + Script Refactor) |
|---|---|---|---|
| Fixes `51C7CD42` CI race timeout | Yes — replaces 21-25s real `go build` with ~5ms `installGoShim` + explicit timeout `t.Fatalf` | Partial — masks slowness until CI load exceeds 60s | Yes |
| Fixes `AD132ACF` Git Bash sentinel flake | Yes — anchors sentinel logs via `toBashPath` (`/c/...`) + `test -ef` `TMPDIR` check | No — timeout has no effect on `/tmp` mount aliasing | Yes |
| Respects P-021 / Option B `freeze-scope` | Yes — strictly scoped to `51C7CD42` and `AD132ACF` | Yes | No — pulls adjacent files and `80E3F89C` residuals into scope |
| Task independence & 2-hour rule | Yes — 2 independent `S` tasks on disjoint files | Yes | No — couples multiple files and domains |

## Decision

Adopt **Option 1** and promote immediately to implementation planning (`docs/plans/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-plan.md`).

* **Release Unit Classification**: Semantic **chore** (internal CI test reliability and harness maintenance, not a user-facing product feature). Because `.backlogit/config.yaml` (`backlogit metadata types`) defines `feature` (`{NNN}-F`) as the sole level-1 work-item container and does not define a separate `chore` artifact type, the level-1 backlog item is created with `--type feature` (`080-F`) and labeled `chore` with explicit chore classification in its title/description and shipment (`073-S`).
* **Task Independence (No Fabricated Dependencies)**: Unit 1 (`080.001-T`, `51C7CD42`, `priority: high`) modifies only `tools/gatecheck/check_retired_architecture_wrapper_test.go` and calls the existing `installGoShim(t, bash)` function without signature changes. Unit 2 (`080.002-T`, `AD132ACF`, `priority: low`) modifies only `tools/gatecheck/wrapper_harness_test.go` and `tools/gatecheck/check_write_path_wrapper_test.go`. Because their file sets and APIs are disjoint, no artificial `blocks` edge is added between `080.001-T` and `080.002-T`; Priority ordering (`high` then `low`) naturally sequences `080.001-T` ahead of `080.002-T`. At the shipment level, `073-S` carries the explicit `blocks` dependency edge `073-S -> 072-S` chaining after the last completed and `READY` shipment `072-S`.

## Rejected Alternatives

* **Option 2 (Timeout Bump Only)** was rejected because `AD132ACF` is caused by MSYS2 `/tmp` path aliasing between `resolveBashPath` and child wrapper invocations, not by process timeout, and because running two unshimmed `go build ./tools/gatecheck` compilations inside `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` directly contradicts that test's own design contract.
* **Option 3 (Broad Wrapper + Script Refactor)** was rejected because the operator froze Option B strictly to `51C7CD42` and `AD132ACF`.

## Unresolved Questions

None. Both root causes are verified directly from the repository source and commit history.

## Risks and Mitigations

* **Risk — Red-phase test verification for test-only tasks (P-002 / P-004)**: Both tasks modify `_test.go` files in `tools/gatecheck`.
  * **Mitigation**: Each task specifies a concrete test-first red-then-green sequence so Ship's `harness-architect` and `build-feature` phases can observe a deterministic red assertion before applying the green fix.
* **Risk — Cross-platform behavior between Linux CI (`ubuntu-latest`) and Windows Git Bash (`windows-latest` / local host)**: `toBashPath` and `test -ef` behave identically on Linux and MSYS2, but path assumptions could regress if Windows-specific branches are not guarded.
  * **Mitigation**: `toBashPath` is already a no-op on POSIX paths and converts `C:/...` to `/c/...` on Windows; `test -ef` is standard across GNU/MSYS Bash and already proven by `sameBashFile` (`061.003-T`).
