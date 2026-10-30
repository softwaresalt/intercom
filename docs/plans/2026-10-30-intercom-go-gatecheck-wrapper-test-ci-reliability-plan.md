---
title: "Gatecheck wrapper test CI reliability and Git Bash sentinel stability: implementation plan"
description: "Implementation plan for Option B (51C7CD42, AD132ACF): wire hermetic go shim and explicit timeout failure in TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst, and anchor installGoShim sentinel logs via toBashPath with test -ef TMPDIR identity verification."
date: 2026-10-30
source: "docs/decisions/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-deliberation.md"
stash_ids:
  - "51C7CD42"
  - "AD132ACF"
---

## Problem Frame

Two deferred-scope-expansion bugs cause intermittent test failures in `tools/gatecheck`:

1. **`51C7CD42` (`priority: high`, `kind: bug`, originating from `078.002-T` / `078-F` / `071-S` / `PR 212`)**:
   In `tools/gatecheck/check_retired_architecture_wrapper_test.go` (lines 103-156), `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` executes `bash -x scripts/check-retired-architecture.sh` across two subtests (`--self-test` and `--self-test-integrity`) without installing the package's hermetic `go` shim (`installGoShim(t, bash)`). Because `scripts/check-retired-architecture.sh` runs `gatecheck_build` (line 207) before `gatecheck_invoke retired-arch` (lines 214/216), each subtest spawns a real `go build -trimpath -o ... ./tools/gatecheck` compilation and runs the real binary. Under CI `go test -race -mod=readonly ./...` load, building `./tools/gatecheck` took 21-25 seconds and exceeded the hard-coded `time.After(20 * time.Second)` timer on three consecutive PR 212 runs. Furthermore, the timeout arm in `select` (lines 139-144) kills the process without calling `t.Fatalf`, causing execution to fall through to `if !strings.Contains(out, "+ "+tc.wantInvoke)` and misreport a slow build as a missing `gatecheck_invoke` line.
2. **`AD132ACF` (`priority: low`, `kind: bug`, originating from `061.001-T` / `061-F` / `053-S` / `PR 165`)**:
   In `tools/gatecheck/wrapper_harness_test.go` (lines 58-127), `installGoShim(t, bash)` sets `TMPDIR`, `TMP`, and `TEMP` to `dir` via `workspaceTempDir(t)` and then computes `bashDir := resolveBashPath(t, bash, dir)` to set `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG`. Under Git for Windows (MSYS2), `/etc/fstab` mounts `/tmp` as `usertemp` (bound to the Windows `TMP`/`TEMP` of whichever MSYS2 process initializes the shared `msys-2.0.dll` mount table). When `resolveBashPath` runs in an isolated MSYS2 mount table, `cd "$directory" && pwd` returns `"/tmp"`, setting `GATECHECK_SENTINEL_LOG` to `"/tmp/generated-binary-calls.log"`. When `TestCheckWritePathWrapper_LegitimateModesReachBuild` (`tools/gatecheck/check_write_path_wrapper_test.go`, lines 45-113) subsequently spawns wrapper subprocesses during a parallel `go test ./...` run where a concurrent MSYS2 process holds `/tmp` pointed at another `TMP` directory, the generated sentinel binary writes `"invoked\n"` to `/tmp/generated-binary-calls.log` in that other directory while `go-calls.log` and `build-output-path.log` (addressed via `${shim_dir}`) land in `shimDir`. Go's `os.ReadFile(filepath.Join(filepath.Dir(goShimLog), "generated-binary-calls.log"))` then fails with file-not-found. Additionally, `installGoShim`'s `managed_tmpdir` check (lines 74-78) compares `$6` against `${managed_tmpdir}` by string prefix alone without a `test -ef` directory-identity fallback when `/tmp` and `/c/...` name the same directory.

## Requirements Trace

| Requirement ID | Source Reference | Requirement Summary | Target Implementation Unit |
|---|---|---|---|
| `R-1` | `51C7CD42` (`docs/decisions/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-deliberation.md`, Option 1 Unit 1) | Eliminate unshimmed `go build` execution inside `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` by wiring `installGoShim(t, bash)` and per-subtest shim log reset. | Unit 1 (`080.001-T`) |
| `R-2` | `51C7CD42` (`docs/decisions/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-deliberation.md`, Option 1 Unit 1) | Fail explicitly with `t.Fatalf` (including kill/wait errors and captured trace output) if the 20s subprocess safety timer expires in `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst`, and assert exit code `0` and deterministic `PS4=+ ` / `LC_ALL=C` trace environment. | Unit 1 (`080.001-T`) |
| `R-3` | `AD132ACF` (`docs/decisions/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-deliberation.md`, Option 1 Unit 2) | Anchor `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG` in `installGoShim` (`tools/gatecheck/wrapper_harness_test.go`) to `toBashPath(dir)` rather than `resolveBashPath(t, bash, dir)` so sentinel logs never route through the volatile MSYS2 `/tmp` mount alias. | Unit 2 (`080.002-T`) |
| `R-4` | `AD132ACF` (`docs/decisions/2026-10-30-intercom-go-gatecheck-wrapper-test-ci-reliability-deliberation.md`, Option 1 Unit 2) | Add a `test -ef` directory-identity fallback (`[ ! "$(dirname "$(dirname "$6")")" -ef "${managed_tmpdir}" ]`) to the `managed_tmpdir` check in `installGoShim`'s `go` shim script, normalize the written `shim` script via `normalizeScript(shim)`, reset `generated-binary-args.log` alongside the other shim logs in `TestCheckWritePathWrapper_LegitimateModesReachBuild`, include `(trace=%q)` on log-read failures, and cover both behaviors with a dedicated unit test. | Unit 2 (`080.002-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | Compliant | Changes are confined to Go `_test.go` files in `tools/gatecheck`; all error returns (`os.Remove`, `os.ReadFile`, `exec.ExitError` via `errors.As`) are handled explicitly with zero `golangci-lint` / `go vet` / `gofmt` findings. |
| II. Test-First Development (NON-NEGOTIABLE) | Compliant | Both units are `tests`-domain tasks with explicit red-then-green verification sequences so `harness-architect` observes a compiling red test before `build-feature` applies the green change. |
| III. Workspace Isolation and Security Boundaries | Compliant | All temporary directories use `workspaceTempDir(t)` (contained within the workspace working directory) and `t.TempDir()`. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | Compliant | All file paths resolve inside `C:\Source\GitHub\intercom`. |
| V. Structured Observability | Compliant | Timeout and log-read failure assertions include full `bash -x` trace output and exit diagnostics. |
| VI. Single Responsibility | Compliant | Zero new external dependencies; reuses standard-library packages and existing `wrapper_harness_test.go` helpers. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | Compliant | No destructive commands are used or planned. |
| VIII. Explicit Safety Modes for Elevated Risk | Compliant | Executed under `investigate-first` and `freeze-scope` modes with explicit `ProposedAction` / `ActionRisk: moderate` / `ActionResult: approved` tracking. |
| IX. Git-Friendly Persistence | Compliant | Planning and backlog artifacts use UTF-8 Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | Compliant | Retrieved context via `engram` CLI and targeted `backlogit` queries. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Compliant | Stage commits only to `chore/stage-gatecheck-wrapper-test-ci-reliability`; downstream PRs merge via merge commits only. |

## Implementation Units

### Release Unit Container (Top-Level Chore-Classified Feature `080-F`)

* **Classification**: Semantic **chore** (internal CI test reliability and harness maintenance). Per installed `.backlogit/config.yaml` (`backlogit metadata types`), `feature` (`{NNN}-F`) is the only level-1 parent artifact type (`allowed_children: [task, review]`), so the top-level container is created with `--type feature` (`080-F`), labeled `chore`, and titled `"Gatecheck wrapper test CI reliability (51C7CD42, AD132ACF)"`.

### Unit 1: Shim `go` build and fail explicitly on timeout in `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` (`080.001-T`)

* **Source Stash ID**: `51C7CD42` (originating refs: `078.002-T`, `078-F`, `071-S`, `PR 212`, `review-thread: N/A`)
* **Priority**: `high`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: stage-2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `tests` (width-isolated to `_test.go` in `tools/gatecheck`)
* **Execution Posture**: `test-first`
* **Files Affected (1)**:
  * `tools/gatecheck/check_retired_architecture_wrapper_test.go`
* **Functions Changed / Added (2)**:
  * `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` (modified)
  * `runRetiredArchWrapperWithTrace` (added helper in `tools/gatecheck/check_retired_architecture_wrapper_test.go`)
* **Test Scenarios (2)**:
  1. Subtest `"self-test mode ignores a trailing root-override attempt"` runs against the hermetic `go` shim (`installGoShim(t, bash)`), exits `0`, records `go` shim invocations in `filepath.Join(shimDir, "go-calls.log")`, emits `+ gatecheck_invoke retired-arch --self-test`, and omits `/tmp/evil`.
  2. Subtest `"self-test-integrity mode ignores a second allowlisted mode"` runs against the hermetic `go` shim (`installGoShim(t, bash)`), exits `0`, records `go` shim invocations in `filepath.Join(shimDir, "go-calls.log")`, emits `+ gatecheck_invoke retired-arch --self-test-integrity`, and omits `gatecheck_invoke retired-arch --self-test-integrity --self-test`.
* **Detailed Changes**:
  1. Extract helper `runRetiredArchWrapperWithTrace(t *testing.T, bash, script string, args ...string) (string, int)` in `tools/gatecheck/check_retired_architecture_wrapper_test.go` (mirroring `runWritePathWrapperWithTrace` / `runWritePathWrapperInDir` in `check_write_path_wrapper_test.go`):
     * Call `t.Helper()` first; construct `cmd := exec.Command(bash, append([]string{"-x", script}, args...)...)` with `cmd.Env = append(os.Environ(), "PS4=+ ", "LC_ALL=C")` and combined stdout/stderr buffer.
     * Wait on `cmd.Wait()` with a `20 * time.Second` safety deadline; if the timer fires, kill `cmd.Process`, drain `<-done`, and call `t.Fatalf("wrapper timed out after 20s (kill err=%v, wait err=%v; out=%q)", killErr, waitErr, combined.String())` instead of silently falling through.
     * Extract and return `(combined.String(), exitCode)` using `switch { case errors.As(runErr, &exitErr): exitCode = exitErr.ExitCode(); case runErr != nil: t.Fatalf("unexpected error running wrapper: %v (out=%q)", runErr, combined.String()) }`.
  2. In `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst(t *testing.T)`:
     * Update the function doc comment (lines 97-102) so it accurately states that `gatecheck_build` runs before `gatecheck_invoke` and is satisfied by `installGoShim(t, bash)` rather than a real `go build`.
     * Call `shimDir := installGoShim(t, bash)` once before the `for _, tc := range cases` loop.
     * At the start of each subtest, remove `"go-calls.log"`, `"build-output-path.log"`, `"generated-binary-calls.log"`, and `"generated-binary-args.log"` inside `shimDir` (ignoring `os.ErrNotExist`).
     * Invoke `out, exitCode := runRetiredArchWrapperWithTrace(t, bash, script, tc.args...)`, assert `exitCode == 0`, assert `strings.Contains(out, "+ "+tc.wantInvoke)`, assert `!strings.Contains(out, tc.mustNotMatch)`, and assert `filepath.Join(shimDir, "go-calls.log")` is non-empty (confirming `gatecheck_build` hit the hermetic `go` shim rather than the real Go toolchain, without reading `generated-binary-calls.log` so Unit 1 has zero runtime coupling to Unit 2).
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add the `runRetiredArchWrapperWithTrace` stub (or add the `filepath.Join(shimDir, "go-calls.log")` shim-usage assertion in `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` before calling `installGoShim(t, bash)`), confirm `go vet ./tools/gatecheck/...` passes and `go test -run TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst ./tools/gatecheck/...` fails with the expected failure marker.
  * **Green phase**: Wire `shimDir := installGoShim(t, bash)` and the full `runRetiredArchWrapperWithTrace` implementation, and confirm `go test -race ./tools/gatecheck/...` passes in <1s.
* **Acceptance Criteria**:
  1. `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` installs `installGoShim(t, bash)` once per test, resets the four shim log files per subtest, and verifies `go-calls.log` is written in `shimDir` so no real `go build` subprocess runs during the test.
  2. `runRetiredArchWrapperWithTrace` pins `PS4=+ ` and `LC_ALL=C` in `cmd.Env` and calls `t.Fatalf` with kill/wait errors and trace output if the 20-second deadline expires, and `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` asserts `exitCode == 0` along with the existing `wantInvoke` and `mustNotMatch` trace assertions.

### Unit 2: Anchor `installGoShim` sentinel logs via `toBashPath` and verify `TMPDIR` by file identity (`080.002-T`)

* **Source Stash ID**: `AD132ACF` (originating refs: `061.001-T`, `061-F`, `053-S`, `PR 165` [reconciled], `review-thread: N/A`)
* **Priority**: `low`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: stage-2h-rule-v1`)
* **Complexity**: `medium`
* **Skill Domain**: `tests` (width-isolated to `_test.go` in `tools/gatecheck`)
* **Execution Posture**: `test-first`
* **Files Affected (2)**:
  * `tools/gatecheck/wrapper_harness_test.go`
  * `tools/gatecheck/check_write_path_wrapper_test.go`
* **Functions Changed / Added (3)**:
  * `installGoShim` in `tools/gatecheck/wrapper_harness_test.go` (modified)
  * `TestCheckWritePathWrapper_LegitimateModesReachBuild` in `tools/gatecheck/check_write_path_wrapper_test.go` (modified)
  * `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` in `tools/gatecheck/check_write_path_wrapper_test.go` (added)
* **Test Scenarios (3)**:
  1. `installGoShim(t, bash)` sets `GATECHECK_SENTINEL_LOG` to `path.Join(toBashPath(shimDir), "generated-binary-calls.log")` and `GATECHECK_ARGS_LOG` to `path.Join(toBashPath(shimDir), "generated-binary-args.log")` (never `/tmp/generated-binary-*.log`), and executing the generated sentinel binary writes `"invoked"` to `filepath.Join(shimDir, "generated-binary-calls.log")` even when `/tmp` is not used.
  2. The `go` shim's `managed_tmpdir` guard accepts a build output path `$6` (`<subdir>/gatecheck<exe>`) when `$(dirname "$(dirname "$6")")` and `${managed_tmpdir}` refer to the same directory on disk via different path spellings (`resolveBashPath(t, bash, shimDir)` vs `toBashPath(shimDir)`, checked via `test -ef`), and still exits `97` with `"gatecheck output is outside managed TMPDIR"` when `$6` is placed under a different directory outside `TMPDIR`.
  3. `TestCheckWritePathWrapper_LegitimateModesReachBuild` resets all four shim logs (`go-calls.log`, `build-output-path.log`, `generated-binary-calls.log`, `generated-binary-args.log`) before each subtest, includes `(trace=%q)` in all log-read `t.Fatalf` calls, and passes all three subtests (`repo_scan`, `--self-test`, `--self-test-integrity`).
* **Detailed Changes**:
  1. In `tools/gatecheck/wrapper_harness_test.go` (`installGoShim`, lines 58-127):
     * Set `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG` using `toBashPath(dir)` (`path.Join(toBashPath(dir), "generated-binary-calls.log")` and `path.Join(toBashPath(dir), "generated-binary-args.log")`), matching `runWrongTreeRow` (lines 257-258) so the sentinel logs use the deterministic `/c/...` path on Windows rather than `bashDir` (`resolveBashPath`, which returns `"/tmp"` when MSYS2 mounts `TMP` at `/tmp`). Keep `bashDir` (or `toBashPath(dir)`) for the `sameBashFile(t, bash, got, want)` check at line 120.
     * In the embedded `go` shim script (lines 74-78), keep the `case "$6" in "${managed_tmpdir}"/*) ;;` fast branch and replace the wildcard failure arm with an identity fallback:
       ```bash
       managed_tmpdir="${TMPDIR%/}"
       case "$6" in
         "${managed_tmpdir}"/*) ;;
         *)
           if [ ! "$(dirname "$(dirname "$6")")" -ef "${managed_tmpdir}" ]; then
             unexpected "gatecheck output is outside managed TMPDIR: $6 (TMPDIR=$TMPDIR)"
           fi
           ;;
       esac
       ```
     * Write the `go` shim using `normalizeScript(shim)` (`os.WriteFile(shimPath, []byte(normalizeScript(shim)), 0o755)`) so CRLF can never enter the script or its `SENTINEL` heredoc.
  2. In `tools/gatecheck/check_write_path_wrapper_test.go`:
     * In `TestCheckWritePathWrapper_LegitimateModesReachBuild`, include `"generated-binary-args.log"` in the per-subtest log-reset slice and include `(trace=%q)` in the `t.Fatalf` messages for reading `go-calls.log`, `build-output-path.log`, and `generated-binary-calls.log`.
     * Add `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity(t *testing.T)` verifying:
       (a) `os.Getenv("GATECHECK_SENTINEL_LOG") == path.Join(toBashPath(shimDir), "generated-binary-calls.log")` and `os.Getenv("GATECHECK_ARGS_LOG") == path.Join(toBashPath(shimDir), "generated-binary-args.log")`,
       (b) after pre-creating `filepath.Join(shimDir, "sub")` via `os.MkdirAll`, setting `TMPDIR` to `toBashPath(shimDir)+"/sub/.."` (an un-normalized alias spelling of `shimDir` that does not string-prefix-match `toBashPath(shimDir)+"/sub/gatecheck..."` on either Linux or Windows) and invoking `go -C <root> build -trimpath -o <toBashPath(shimDir)>/sub/gatecheck<exe> ./tools/gatecheck` succeeds with exit `0` via the `test -ef` fallback and produces a runnable sentinel binary that writes `"invoked"` to `filepath.Join(shimDir, "generated-binary-calls.log")`, and
       (c) after pre-creating `<outsideDir>/sub` via `os.MkdirAll`, invoking `go -C <root> build -trimpath -o <toBashPath(outsideDir)>/sub/gatecheck<exe> ./tools/gatecheck` where `<outsideDir>` is a separate `t.TempDir()` fails with exit code `97` and `"gatecheck output is outside managed TMPDIR"`.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` in `tools/gatecheck/check_write_path_wrapper_test.go` (which sets `TMPDIR` to `toBashPath(shimDir)+"/sub/.."` while passing `toBashPath(shimDir)+"/sub/gatecheck"+platformGoExe()` as `-o`, and checks `GATECHECK_SENTINEL_LOG` against `path.Join(toBashPath(shimDir), "generated-binary-calls.log")`). Confirm `go vet ./tools/gatecheck/...` passes and `go test -run TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity ./tools/gatecheck/...` fails deterministically on both Linux and Windows before `wrapper_harness_test.go` is updated.
  * **Green phase**: Apply the `toBashPath(dir)`, `test -ef` `managed_tmpdir` fallback, and `normalizeScript(shim)` updates in `tools/gatecheck/wrapper_harness_test.go` and the log-reset/diagnostic updates in `tools/gatecheck/check_write_path_wrapper_test.go`, and confirm `go test -race ./tools/gatecheck/...` passes.
* **Acceptance Criteria**:
  1. `installGoShim` in `tools/gatecheck/wrapper_harness_test.go` sets `GATECHECK_SENTINEL_LOG` and `GATECHECK_ARGS_LOG` using `toBashPath(dir)`, normalizes the written `go` shim via `normalizeScript(shim)`, and validates `$6` against `${managed_tmpdir}` using `test -ef` (`[ ! "$(dirname "$(dirname "$6")")" -ef "${managed_tmpdir}" ]`) when string prefix matching does not match.
  2. `TestCheckWritePathWrapper_LegitimateModesReachBuild` resets all four shim log files per subtest and includes `(trace=%q)` in log-read `t.Fatalf` diagnostics, and `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` verifies both the `toBashPath` log anchoring and the `-ef` `TMPDIR` containment check (accepting aliased spellings of `shimDir` and rejecting outside paths with exit code `97`).

## Dependency Graph

* **Shipment-level DAG**:
  * `073-S` (`blocks`-depends on `072-S`, which is `shipped` and `READY` with `compaction_status: done` in `docs/closure/072-S-079-F-post-merge-closure.md`).
* **Task-level DAG inside `080-F`**:
  * `080.001-T` (`priority: high`, `tools/gatecheck/check_retired_architecture_wrapper_test.go`) — no task predecessor (independent).
  * `080.002-T` (`priority: low`, `tools/gatecheck/wrapper_harness_test.go`, `tools/gatecheck/check_write_path_wrapper_test.go`) — no task predecessor (independent).
  * **Rationale for no inter-task `blocks` edge**: Unit 1 and Unit 2 modify disjoint files and do not change `installGoShim`'s function signature. Per operator instructions, no artificial dependency is manufactured between independent fixes; priority ordering (`high` then `low`) sequences `080.001-T` before `080.002-T`.

## Decisions and Rationale

* **D-1 (Hermetic `go` shim in `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` instead of raising the 20s timeout)**: `scripts/check-retired-architecture.sh` invokes `gatecheck_build` before `gatecheck_invoke`. Installing `installGoShim(t, bash)` fulfills the test's own doc comment (*"rather than waiting on a real `go build`, so a slow/unavailable Go toolchain cannot make this test flaky"*), cuts runtime from 12-25s to <0.5s, and pairs with an explicit `t.Fatalf` in the 20s timeout branch so a hung process never masquerades as a missing trace line.
* **D-2 (`toBashPath(dir)` for sentinel logs and `test -ef` for `managed_tmpdir` in `installGoShim`)**: Under MSYS2 / Git Bash, `resolveBashPath(t, bash, dir)` returns `"/tmp"` when `dir` matches `TMP`, which breaks whenever a concurrent MSYS2 process in `go test ./...` holds `/tmp` pointed at another `TMP` directory. Using `toBashPath(dir)` (`/c/...`) matches `runWrongTreeRow`'s proven pattern (`057.009-T`), and adding `test -ef` to `managed_tmpdir` matches `sameBashFile` (`061.003-T`).
* **D-3 (Chore classification represented via installed `feature` container `080-F` with `chore` label)**: Per `.backlogit/config.yaml` and `backlogit metadata types`, `feature` (`-F`) is the only level-1 parent work-item type in this workspace. `080-F` is created as `--type feature` with `labels: ["chore"]` and explicit chore classification in its description and shipment so Ship uses a `chore/` branch.

## Risks and Caveats

* **Cross-platform red-phase reproduction for Unit 2**: On Linux (`ubuntu-latest`), `resolveBashPath(t, bash, shimDir)` and `toBashPath(shimDir)` return the same string unless an alias (such as a symlink or non-normalized path) is used. To ensure `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` fails deterministically in the red phase on both Linux and Windows (where symlinks may require `ERROR_PRIVILEGE_NOT_HELD`), the test can pass an un-normalized POSIX path containing `/./` or `/../<base>` (e.g. `toBashPath(shimDir) + "/."` or `path.Clean` vs `toBashPath(shimDir) + "/sub/../sub/gatecheck"`) as the `-o` path relative to `TMPDIR`: `"${managed_tmpdir}"/*` fails the literal string prefix check when `TMPDIR` is `toBashPath(shimDir)+"/sub/.."` while `-o` is `toBashPath(shimDir)+"/child/gatecheck"`, whereas `[ "$(dirname "$(dirname "$6")")" -ef "${managed_tmpdir}" ]` succeeds on every platform without needing `os.Symlink` privileges!
* **Windows host `PATH` ordering**: As documented in `053-S` closure notes, local Windows runs must place Git for Windows `bash` ahead of `C:\Windows\System32\bash.exe` (WSL launcher) on `PATH` so `requireBash(t)` runs rather than skipping via `requireWindowsCompatibleBash`.

## Plan Hardening Signals

* **Public API, schema, or contract change**: **Absent** — modifies only internal test files (`tools/gatecheck/*_test.go`); no CLI, HTTP, WebSocket, or config schema changes.
* **Security, auth, permission, or compliance-sensitive behavior**: **Absent** — test-shim path handling only; `managed_tmpdir` containment in `installGoShim` remains fail-closed via `test -ef`.
* **Migration, backfill, destructive data/config action, or irreversible step**: **Absent** — no data, schema, or config migrations and no destructive operations.
* **External integration, operator checkpoint, or external dependency**: **Absent** — hermetic unit tests in `tools/gatecheck` with no external service calls.
* **High runtime, rollout, or rollback risk**: **Absent** — test-only changes in `tools/gatecheck`; rollback is a clean `git revert` of the shipment merge commit.

Requires plan hardening: no

## Runtime Verification and Closure

* **Changed Runtime Surfaces**: **None** (`tools/gatecheck/*_test.go` test-only changes; no changes to `cmd/`, `internal/`, or production `tools/gatecheck/*.go` / `scripts/*.sh`).
* **Verification Expectations for Ship**:
  * Run `gofmt -l .`, `go vet ./...`, `go test -race ./tools/gatecheck/...`, `go test -race ./...`, and `go build ./...`.
  * On Windows hosts, run `go test -v ./tools/gatecheck/...` with Git for Windows `bash` first on `PATH` so `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst`, `TestCheckWritePathWrapper_LegitimateModesReachBuild`, and `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` execute without skipping.
* **Operational Closure**:
  * Post-merge operational closure (`docs/closure/{shipment}-{feature}-post-merge-closure.md`) with `Runtime verification: not applicable (test-only changes in tools/gatecheck)` and mandatory P-020 `compact-context` invocation.

## Plan Review

* `dispatch_mode: multi-agent`
* `decision: PASS`
* **Gate Decision and Rationale**: **PASS**. All five applicable reviewer personas (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, and `Architecture Strategist`) ran as independent subagents (`TOOL_OK: reviewer-subagent-dispatch`). Zero P0 or P1 findings were raised. One P2 finding (`Architecture Strategist`: aligning Unit 1's Test Scenarios wording with its Detailed Changes and Acceptance Criteria so Unit 1 asserts `go-calls.log` rather than `generated-binary-calls.log`, preserving complete runtime independence from Unit 2) and three P3 advisory refinements (`Go Reviewer` and `Architecture Strategist`: `t.Helper()` + non-`ExitError` `t.Fatalf` branch in `runRetiredArchWrapperWithTrace`, updating the stale doc comment on `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst`, and pre-creating `<subDir>` via `os.MkdirAll` while setting `TMPDIR = toBashPath(shimDir)+"/sub/.."` in `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity`) were folded directly into the plan above prior to harvest.
* **Plan Hardening Requirement**: `Requires plan hardening: no` (all five hardening signals absent; P-006 satisfied).
* **Degradations**: `TOOL_DEGRADED: graphtor-docs — declared fallback: direct local documentation reads` (no `graphtor-docs` MCP surface in session; all referenced files are local repo files read directly). `TOOL_OK: reviewer-subagent-dispatch` (`dispatch_mode: multi-agent`).

### Persona Coverage

| Persona | Installed Identity File | Dispatch Mode | Outcome |
|---|---|---|---|
| Constitution Reviewer | `.github/agents/subagents/constitution-reviewer.agent.md` | Subagent (`multi-agent`) | 0 findings (`PASS`) |
| Go Reviewer | `.github/agents/subagents/go-reviewer.agent.md` | Subagent (`multi-agent`) | 0 P0/P1/P2, 3 P3 advisory (folded into Unit 1 and Unit 2) |
| Scope Boundary Auditor | `.github/agents/subagents/scope-boundary-auditor.agent.md` | Subagent (`multi-agent`) | 0 findings (`PASS`) |
| Learnings Researcher | `.github/agents/subagents/learnings-researcher.agent.md` | Subagent (`multi-agent`) | 0 P0/P1 (`ALIGNED`, 37 files searched in `docs/compound/`) |
| Architecture Strategist | `.github/agents/subagents/architecture-strategist.agent.md` | Subagent (`multi-agent`) | 0 P0/P1, 1 P2 (resolved in Unit 1), 2 P3 advisory (folded into Unit 2 / noted for future triage) |
| Agent-Native Parity Reviewer | `.github/agents/subagents/agent-native-parity-reviewer.agent.md` | Not triggered (no MCP tools or agent/user parity surfaces touched) | N/A |
| Security Lens Reviewer | `.github/agents/subagents/security-lens-reviewer.agent.md` | Not triggered (no auth/authz, API, secrets, or trust-boundary surfaces touched) | N/A |

### Findings Summary

* **P0 / P1 (Blocking)**: None.
* **P2 (Resolved before harvest)**:
  * **P2-1 (`Architecture Strategist`)**: Unit 1's Test Scenarios initially mentioned asserting both `go` shim and sentinel invocations in `shimDir`, whereas Detailed Changes and Acceptance Criteria specified asserting `go-calls.log` and the `+ gatecheck_invoke` trace line. Updated Unit 1's Test Scenarios to specify `filepath.Join(shimDir, "go-calls.log")` so Unit 1 has zero runtime coupling to Unit 2's `GATECHECK_SENTINEL_LOG` fix.
* **P3 (Advisory — folded into plan)**:
  * **P3-1 (`Go Reviewer`)**: Specified `t.Helper()` and the `case runErr != nil: t.Fatalf(...)` arm in `runRetiredArchWrapperWithTrace` (Unit 1).
  * **P3-2 (`Go Reviewer`)**: Specified updating the stale comment on `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` (Unit 1).
  * **P3-3 (`Go Reviewer` / `Architecture Strategist`)**: Specified pre-creating `filepath.Join(shimDir, "sub")` and `<outsideDir>/sub` with `os.MkdirAll` and using `TMPDIR = toBashPath(shimDir)+"/sub/.."` in `TestInstallGoShim_SentinelLogsAvoidTmpAliasAndCheckTmpdirIdentity` (Unit 2) so the red phase fails deterministically on both Linux and Windows without requiring symlink privileges.
  * **P3-4 (`Architecture Strategist`)**: Noted that `TestCheckRetiredArchitectureWrapper_AllowlistedModesPassValidation` and `TestCheckUnignoreWrapper_LegitimateFlagsClearTheGuard` remain untouched per Option B's strict `freeze-scope` boundary.
* **Runtime Verification and Operational Closure**: Verified complete; no runtime surfaces are modified (`tools/gatecheck/*_test.go` only), and post-merge operational closure with mandatory P-020 `compact-context` is specified.

