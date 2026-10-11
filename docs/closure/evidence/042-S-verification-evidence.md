---
title: "042-S verification evidence (committed, reproducible from the repository)"
description: "Commands, exit codes, run identifiers, and timings for shipment 042-S, cited by the post-merge closure artifact. The git-ignored logs/ files are local backups of the same runs."
shipment: 042-S
feature: 052-F
pr: 123
reviewed_head: b44cc1539c5d734caef49726306c575f3e6b2700
merge_commit_sha: 591f37e2ba877d81b1304d31905369e79096657f
date: 2026-10-11
---

# 042-S verification evidence

## Required CI on PR #123 (reviewed HEAD b44cc15)

Workflow run `38104296737` on the reviewed HEAD. Every check passed. Job URLs follow the pattern `https://github.com/softwaresalt/intercom/actions/runs/38104296737/job/<job_id>`.

| Check | Result | Duration | Job id |
|---|---|---|---|
| test | pass (the required Linux job; the PR check name is `test`) | 1m38s | 114366421477 |
| test (windows, advisory) | pass at job level only: its test step is continue-on-error while `WINDOWS_GATE_REQUIRED` is unset, so a failing go test leaves the job green | 2m2s | 114366421435 |
| lint | pass | 1m48s | 114366421439 |
| security | pass | 56s | 114366421419 |
| cross-compile (linux/amd64) | pass | 24s | 114366436480 |
| cross-compile (darwin/amd64) | pass | 22s | 114366436528 |
| cross-compile (darwin/arm64) | pass | 21s | 114366436570 |
| cross-compile (windows/amd64) | pass | 18s | 114366436530 |
| merge-strategy structural verification | pass | 24s | 114366396124 |
| gitignore append-only + un-ignore regression (I6) | pass | 17s | 114366396079 |
| pipeline-topology (ambient) | pass | 12s | 114366396190 |
| ci gate | pass | 3s | 114366753651 |

## Local pre-PR gate on HEAD b44cc15

| Command | Exit | Note |
|---|---|---|
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l .` | 0, empty | |
| `goimports -l .` | 0, empty | |
| `golangci-lint run ./...` | 0 | 0 issues |
| `GOOS=linux go vet ./tools/gatecheck/...` | 0 | |
| `go test -count=1 ./...` | 0 | 14 packages ok (log `logs/final-b44cc15-fulltest.txt`, local) |

## Live tree (run at 5eb2e78; e202fe3 changed only a test file, and b44cc15 only backlog and stash records)

The exit codes below were observed in the session command output. Those commands print nothing on success, so the log files under logs/ are empty.

| Command | Exit |
|---|---|
| `go run ./tools/gatecheck write-path --root .` | 0 |
| `go run ./tools/gatecheck write-path --root . --self-test` | 0 |
| `go run ./tools/gatecheck write-path --root . --self-test-integrity` | 0 |
| `go run ./tools/gatecheck retired-arch --root .` | 0 |
| `bash scripts/check-write-path-precondition.sh` | 0 |

## Targeted test evidence

* `TestRunFixtureSelfTest_LinkedFixture_FailsClosed`: 5 of 5 subtests PASS, 0 SKIP, on a symlink-capable Windows host (row 1 ran locally, not only in CI).
* `TestScannedPathsFromListing_Table` (7 subtests), `TestRunRepoScan_BadListing_FailsClosed` (14 subtests), `TestDefaultGitRunner_NonRepository_FailsClosed`: PASS, 0 SKIP (fix cycle 3, e202fe3).

## Timing condition (pre-existing, intermittent)

`tests/integration` contains pwsh-driven start-script tests with a 30 s budget for `start.ps1` (`start_script_test.go`, `startScriptTimeout`). Recorded timeouts:

| Run | Date | Result |
|---|---|---|
| `logs/fulltest-final.txt` | 2026-10-08 (before the 042-S branch existed) | `TestInvokeStartScriptMainPropagatesNonZeroExitCode`: `start.ps1 exceeded 30s timeout` |
| `logs/fulltest-head.txt` | 2026-10-08 (before the 042-S branch existed) | the same timeout |
| `logs/gate-052.004-T-fulltest.txt` | 2026-10-10 (042-S branch) | `TestInvokeStartScriptMainPropagatesNonZeroExitCode` (157.5 s) and `TestLocationRestoredEvenOnNonZeroCopilotExit` (32.0 s, `location-check helper exceeded 30s timeout`) |
| `logs/final-e202fe3-fulltest.txt` | 2026-10-10 (042-S branch) | `TestInvokeStartScriptMainPropagatesNonZeroExitCode` |

Isolated reruns passed: `logs/052.004-T-integration-rerun.txt` (39.8 s; a package-level `ok` line without `-v`, so per-test verdicts are not in that log) and `logs/final-startscript-isolated.txt` (19.6 s). The final full run on the reviewed HEAD (`logs/final-b44cc15-fulltest.txt`) passed the whole package.

Cause: not established. The timeout predates this branch, so the 042-S workspace growth is a hypothesis only. The 041-S closure already recorded a start.ps1 timeout on a pre-toolchain baseline. The Linux CI test job, which is authoritative, passed.
