# Ship checkpoint — 038-S / Unit C

## Completed

- **048.001-T** — first `--root` occurrence is authoritative; repeated
  `--root` / `--root=` is rejected. Commit:
  `187bf4475ff49e29cbaae8abc06b2f56385be8f6`.
- **048.002-T** — added write-path wrapper root-denylist tests. Commit:
  `2c4c4d8`. This is the plan-declared C-T2 red phase; the new rejection cases
  fail until 048.003-T adds the wrapper guard.
- **048.003-T** — added the per-argument `--root` denylist to the write-path
  wrapper before `gatecheck_build`. Commit:
  `15eb080c07f4194afe465ec1ea69c0e931a383ee`.
- **048.004-T** — corrected the five wrapper and test rationale comments to
  describe first-wins parsing and defense-in-depth guards. Commit:
  `5b59b8a5888645f27843c36e93817b4ecdab8dd9`.
- **048.005-T** — recorded the three verified Python-parity behaviors in the
  merge-strategy and unignore package documentation as accepted, scoped
  Principle III exceptions. No executable code or behavior changed. Commit:
  `6374d14f2768e603eccf1de6c510199bf1e8508c`.
- **048.006-T** — appended the M4-retired / defensive-historical note beside
  the existing Python bytecode ignore rules without removing or reordering
  either rule. Commit:
  `d2bd6d9ba19002a8e384b92081ddc032b0e3f011`.

## Verification and decisions

- C-T1 test-first red observed for repeated root options; green after the
  parser fix. All four required wrapper self-tests returned exit 0 both before
  and after the change.
- C-T1 `go vet ./...`, `go test ./...`, `go build ./...`, `golangci-lint`,
  `staticcheck`, `goimports`, and LF-normalized gofmt checks passed.
- C-T2 `go vet ./...`, `go build ./...`, `golangci-lint`, `staticcheck`,
  `goimports`, and gofmt passed. The full suite's only failures are the
  explicitly expected C-T2 pre-C3 rejection assertions: the unguarded wrapper
  reaches `gatecheck_build`, while the new harness requires rejection before
  build. At this intermediate boundary the engine returns the C-T1 duplicate
  root diagnostic rather than the C-T3 wrapper diagnostic; C-T3 must make these
  assertions pass.
- C-T3 turned the C-T2 rejection assertions green. Its full Go test, build,
  vet, lint, staticcheck, goimports, and LF-normalized gofmt checks passed.
- C-T4 changed comments only; all four wrapper self-tests and the full task
  quality gates passed.
- C-T5's two package GoDoc commands, targeted tests, full `go test ./...`,
  `go vet ./...`, `go build ./...`, golangci-lint, staticcheck, goimports,
  LF-normalized gofmt, and diff checks passed. The commit changes only package
  comments; accepted parity exceptions are not fixed or expanded.
- C-T6's append-only and unignore checker self-tests passed, as did both
  checkers' base/head CI entry points for base
  `7fad096fffc274756998b145a2b7d5d5604f2e4e` and HEAD. The append-only checker
  reported PASS; unignore reported PASS with 9 denylist and 6244 differential
  cases evaluated. Go tests, vet, build, lint, staticcheck, goimports,
  LF-normalized gofmt, and `git diff --check` passed.
- Shipment-level verification on the final source tree passed:
  `go test ./...` (all packages), `go vet ./...`, `go build ./...`,
  `golangci-lint run ./...` (0 issues), `staticcheck ./...`,
  `goimports -l .`, and a UTF-8-safe LF-normalized gofmt audit (0 files need
  formatting). The required pipeline-topology `lifecycle` gate passed for
  038-S on this feature branch.
- P-001 check: 048-F is the only active top-level feature; there are no active
  chores. No PR exists yet for this branch.
- Adversarial review on `06d2807` reported three in-scope P3 items (C-T4
  rationale-comment wording, C-T2 legitimate-mode harness timeout behavior,
  and C-T1 duplicate-root test edge coverage) plus one out-of-scope P3
  suggestion to add a rationale comment to the C-T3 wrapper guard. The
  out-of-scope item was captured before disposition as stash **D10E82EC**;
  its source refs are task 048.003-T, feature 048-F, shipment 038-S, PR N/A,
  and review-thread N/A. The entry includes
  `DISCOVERY-STATUS: LOOKUP-UNAVAILABLE`: active stash was queryable and no
  source-ref match was found, but the installed backlogit CLI/index exposes no
  enumeration of archived stash entries. No thread exists on this pre-PR
  threadless path. Stage owns later triage.
- Remediated all three C1 in-scope findings in commit
  `6a6eb9d8289650e28e40b9e6efc6cb9bd698be31`: corrected the stale shared-root
  comment, bounded the legitimate-mode wrapper harness with a temporary Go
  shim (without spawning a real build), and added both repeated-root ordering
  cases. Focused tests and the full Go quality gates passed. The C3 deferred
  comment was not added to code.
- The Windows shell-test environment needs
  `C:\Program Files\Git\bin` at the front of `PATH`; the WSL launcher named
  `bash.exe` otherwise translates paths incorrectly. A transient
  `rm: Device or resource busy` appeared while cleaning the wrapper's generated
  executable on one rejection subtest; the next full run had no cleanup error.
  No task code was changed to work around this platform behavior.
- Backlog CLI fallback used because backlogit MCP tools are absent. Telemetry
  begin returned `disabled` for both completed tasks.

## Current state

- Shipment: `038-S`, active. Feature: `048-F`, active.
- Branch: `feat/038-s-harden-gatecheck-cli-argument-containment`.
- Working tree includes the shipment/task lifecycle updates and the intake
  reconciliation report. All six task-scoped commits are committed and
  048.001-T through 048.006-T are done. The `.backlogit/` lifecycle changes
  are intentionally uncommitted until post-merge safe-close; keep them in the
  worktree across the feature-branch merge and closure-branch creation. No PR
  is open. C-T5 and C-T6 source locks were released after their respective
  task verification.
- Next: run report-only local review and post-remediation adversarial review on
  the final current HEAD. The first adversarial report noted that exact model
  identities were unavailable; the committed diff was independently verified
  locally, but cross-model diversity remains unverified unless the rerun
  confirms it. Then commit this memory checkpoint, prepare current-HEAD
  readiness evidence, push/create the feature PR, and follow the required
  Copilot-review / CI / merge gates. Do not use admin fallback.
- **Stop condition:** if an authorized change makes any existing engine
  self-test change verdict or exit code, halt and return to Stage.
