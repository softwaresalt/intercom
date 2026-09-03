---
title: "Ship session memory — intercom-go Foundation (001-S)"
date: 2026-09-03
shipment: 001-S
feature: 001-F
branch: feat/intercom-go-foundation-module-skeleton-and-hardened-ci-gates
---

# Ship Session Summary — 001-S

## Scope

Shipment `001-S`: "intercom-go Foundation: Module Skeleton and Hardened CI Gates".
Covering feature `001-F`, 2 tasks (`001.001-T`, `001.002-T`), 8 subtasks, all
complete. Plan: `docs/plans/2026-09-03-intercom-go-foundation-plan.md`.

## Items completed (all 8 subtasks + 2 tasks + covering feature — status `done`)

| ID | Title | Commits |
|---|---|---|
| 001.001.001-ST | Initialize Go module and build hygiene | 633d5ea, 3aa7db0 |
| 001.001.002-ST | Add cmd/intercom entrypoint | eddfb06, be4a9aa |
| 001.001.003-ST | Add cmd/intercom-ctl entrypoint | 7bb5929, 1a5ff00 |
| 001.002.001-ST | Add hardened core CI workflow | 7a1b847, 830f6d1 |
| 001.002.002-ST | Add format and lint gates | a3a0b7a, a5cdb7f |
| 001.002.003-ST | Add static-analysis/vuln/secret-scan gates | 64c4bce, 3a4a15f |
| 001.001.004-ST | Add canonical cross-compile build script | 4a8dc6b, 8b4073f |
| 001.002.004-ST | Add cross-compile verification matrix | af6d5db, 3d1ab85 |
| — | Review-fix (P1 + in-scope P2 findings) | 04dbbd0 |
| — | Parents/feature marked done | 6d88f4c |
| — | Deferred P-021 findings captured to stash | fa62c6b |

## Operator checkpoints resolved (per task instructions, without stopping)

* Module path: `github.com/softwaresalt/intercom-go` — confirmed.
* Go language floor: `go 1.22` — confirmed (workspace standard overrides brief's "1.21+").

## Dirty-state handling

* Pre-existing modified `.gitignore` and untracked `.claude/`, plus
  `.backlogit/hooks_queue.jsonl`, were **stashed** (`git stash push -u -- .gitignore
  .claude .backlogit/hooks_queue.jsonl`, stash@{0}) before branch creation so A1's
  append-only `.gitignore` diff would be computed against the pristine committed
  version, not conflated with unrelated operator edits. **The stash must be
  restored after shipment closure** (post-merge Step 6, per task instructions).
* `references/herdr` (the Rust oracle) is a nested git repo, never staged/touched.

## Quality gate evidence (local, final pass before PR)

* `go mod verify`: all modules verified
* `go mod tidy`: no dirty diff
* `go build -mod=readonly ./...`: exit 0
* `go vet ./...`: exit 0
* `gofmt -l .`: clean
* `goimports -l .`: clean
* `golangci-lint run ./...` (v2.13.2, pinned): 0 issues
* `staticcheck ./...` (2026.2.1, pinned): clean
* `govulncheck ./...` (v1.7.0, pinned): 0 vulnerabilities in code paths called
* `go test -race -mod=readonly ./...`: all packages pass
* `scripts/build.ps1`: 8/8 artifacts produced; `go version -m` confirms
  `CGO_ENABLED=0`; negative test (path-escape refusal) passes
* gitleaks positive control (local, not committed): planted Stripe-format test
  key detected, exit 1 — proves the gate is wired (I7)

## Review gate

4 persona reviewers (Go Reviewer, Correctness Reviewer, Constitution Reviewer,
Security Reviewer) run in parallel, report-only mode, against
`git diff main` (25 files). All four returned `READY_WITH_FOLLOWUPS`. 1 P1
found (unbounded subprocess call in a new integration test) — fixed. 4 in-scope
P2s fixed (error logging, CI needs-condition robustness, test-quality gaps,
case-sensitive path guard). 6 out-of-scope findings captured to stash under
P-021 C2 (EFFAA358, 90350C9A, F7C6420D, BEDD2E70, 35D76D5E, EF9352FB) rather
than fixed, since they fall outside the 8 authorized subtask units' contracts.

## Branch / PR state

Branch: `feat/intercom-go-foundation-module-skeleton-and-hardened-ci-gates`.
Not yet pushed / PR not yet created as of this checkpoint — next step is
Step 5 PR lifecycle.

## Next steps

1. Push branch, create PR via pr-lifecycle skill.
2. Handle CI/fix-ci cycles if the hosted GitHub Actions run surfaces anything
   the local equivalents didn't catch.
3. P-014/P-018 gates, operator merge approval.
4. Post-merge: shipment-reconcile safe-close, operational-closure,
   compact-context (P-020), **restore the parked stash** (`git stash pop`),
   source-artifact cleanup (no `source_stash_id`/`source_deliberation_id` on
   001-F per current record — verify at closure time).
