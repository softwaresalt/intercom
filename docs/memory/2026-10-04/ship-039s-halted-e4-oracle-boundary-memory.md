---
title: "039-S halted at E-T4 oracle boundary"
date: 2026-10-04
agent: ship
feature: 049-F
shipment: 039-S
task: 049.004-T
mode: dark (P-017)
status: resolved
---

## Halt reason

E-T4 cannot pass the required full test suite without changing the D-T1a oracle.
The authorized 049.002-T AC-E2.6 required that oracle's fixture inputs remain
the explicit frozen 19-name set from the 031-S merge. Instead,
`TestOracle_FixtureCorpus_Parity` in `writepath_oracle_test.go` still uses a
dynamic `filepath.Glob` over the flat fixture directory. Adding E-T4 alias and
split-selector fixtures therefore makes the legacy oracle compare constructs
it was explicitly required to exclude.

AC-E2.6 also limits the D-T1a oracle to one authorized adaptation and states
that any further oracle change halts (H-3). The failing full suite exposed the
missing frozen fixture boundary. No further oracle change, alternative test
suppression, or out-of-scope repair was attempted. This is an unmet in-scope
acceptance condition, not a P-021 deferred-scope-expansion finding; no stash
entry was created.

## State

* Halt time: `2026-10-04T08:38:10Z`.
* Branch: `feat/migrate-write-path-scanner-to-go-ast-post-m4-unit-e`.
* Committed HEAD: `20715ceb905716a4d1138f77dc5dd8dd9052fc6e`.
* Shipment 039-S: active. Tasks 049.001-T, 049.002-T, and 049.003-T are done;
  049.004-T remains active.
* No PR exists. No E-T4 change has been staged, committed, or pushed.
* The only uncommitted E-T4 paths are `writepath.go`, the writepath golden JSON,
  and the four E-T4 flat fixtures: `accept-alias-copilot.go`,
  `reject-alias-os-writefile.go`, `reject-dot-import-os.go`, and
  `reject-split-selector.go`.
* The active structured checkpoint `checkpoint-20261004-073226.json` remains
  untouched; the engine implementation and backlog operations stay within
  Ship's authority boundary.

## Verification and next step

The E-T4 harness and writepath-package tests excluding the frozen oracle tests
passed. `go vet ./...`, `go build ./...`, formatting, and `golangci-lint run`
passed. The full `go test ./...` failed in
`TestOracle_FixtureCorpus_Parity` for the alias and split-selector fixtures.
`git diff --check` passed. The task-specific red run before implementation
failed on the expected alias, dot-import, split-selector, and missing-golden
cases.

H-3 requires operator disposition before any further oracle change. Preserve
the uncommitted E-T4 work and resume only after the authorized oracle boundary
is resolved. Do not proceed to 049.005-T, open a PR, or merge while the
required full suite is failing. Engram remains degraded and must not be
restarted.

## Resolution (ORCH-D1)

The Orchestrator determined that AC-E2.6's explicit fixture-input freeze was
incomplete in the already-done 049.002-T, not a new oracle change. The exact
same-surface completion was authorized without changing the legacy scanner,
expectations, selector list, or restriction logic.

Corrective commit `aeb3313576f1b1e00b500d1aa532560699ff0a5e` pins exactly the
19 fixture names and order from the 031-S merge tree `c04d975`. The change is
separate from E-T4. The 049.007 freeze guard was observed failing before the
edit and passing afterward; the oracle parity tests, `gofmt -l .`,
`go vet ./...`, `go test ./...`, and `go build ./...` passed afterward.

E-T4 is now complete. Implementation commit
`59bdd951ca55d6f29b653b0048b73b5f510ec2a0` contains the scanner, fixtures,
and additive goldens; backlog archive commit
`10e38c7c037850e8e52f2b16ad75ff8f5d4a67ea` records 049.004-T as done.
The task's backlog commit-tracking field was not updated: Ship's Role Boundary
does not authorize that unlisted backlog mutation, so the task ID remains
traceable through the commit subject and this memory record.

Before PR creation, readiness evidence and the adversarial review must disclose:
"AC-E2.6 input-freeze landed in a follow-up corrective commit rather than the
E-T2 commit 8aaafd7; authorized by Orchestrator ORCH-D1." The review must
confirm that the corrective commit contains only the exact 19-name freeze.
The 049.002-T backlog commit-tracking field was also not updated under the
Ship Role Boundary; its commit subject and this memory record preserve the
association. The next task is 049.005-T. No PR exists yet.
