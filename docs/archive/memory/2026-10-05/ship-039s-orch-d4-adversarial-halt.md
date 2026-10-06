---
title: "039-S ORCH-D4 adversarial review halt"
date: 2026-10-05
agent: ship
feature: 049-F
shipment: 039-S
task: 049.005-T
mode: dark (P-017)
status: blocked
branch: feat/migrate-write-path-scanner-to-go-ast-post-m4-unit-e
reviewed_head: 8aca231566f98cab9ae62bf5234fd38992f57de8
pr: none
---

## Outcome

`DARK_MODE_HALTED` before push, PR creation, Copilot review, CI, merge, or
shipment closure. The ORCH-D4-authorized one-cycle fix was implemented and
locally verified. The fresh adversarial review returned `BLOCKED` on a new
in-scope P1 verification finding. The operator's terminal rule prohibits
another fix cycle in this session. No source change was made after the review.

The feature branch remains the active workspace branch. Shipment `039-S` and
feature `049-F` remain active; all manifest tasks were already done or archived.
No task status was changed. Task `049.005-T` now tracks the fix commit below,
and a concise task comment records this halt.

## ORCH-D4 fix cycle

The operator explicitly authorized exactly one bounded fix cycle for the two
confirmed `049.005-T` false negatives:

1. `(root).Resolve(...)` after a supported `pathsafe.NewRoot` binding.
2. A loop calling `root.Resolve(...)` before a later plain
   `root, err = pathsafe.NewRoot(...)` assignment in the same function.

`tools/gatecheck/internal/writepath/writepath_test.go` gained task-gated
regression cases for both inputs and an unbound-`pathsafe.Root` negative
control. The red run failed with zero findings for both positive cases while
the negative control passed. Production code then unwraps nested
`ast.ParenExpr` receivers and discovers supported NewRoot bindings before
inspecting calls within each function. The nested function boundary and other
task scope were retained.

* Commit: `8aca231566f98cab9ae62bf5234fd38992f57de8`
* Subject: `fix: detect parenthesized and forward NewRoot receivers (049.005-T)`
* Paths: `tools/gatecheck/internal/writepath/writepath.go` and
  `tools/gatecheck/internal/writepath/writepath_test.go`.
* Task tracking: `backlogit update 049.005-T --commit ...` succeeded and
  `backlogit get 049.005-T` verified the commit association.

### TDD and quality evidence

The task harness command was run with
`WRITEPATH_HARNESS_TASK=049.005-T` and Git-for-Windows `bin` and `usr\bin`
prepended to `PATH` in the same PowerShell process.

* Red: parenthesized receiver test reported zero findings; forward-assignment
  test reported zero findings; unbound-root control passed.
* Green: `go test ./tools/gatecheck/internal/writepath -run
  '^TestHarness_049005_' -count=1` passed.
* `gofmt -l .`: no output.
* `go vet ./...`: passed.
* `go test ./...`: passed.
* `go test -race ./tools/gatecheck/internal/writepath`: passed.
* `go build ./...`: passed.
* `git diff --check`: passed.

## Adversarial review and readiness

The review covered HEAD `8aca231566f98cab9ae62bf5234fd38992f57de8` against
merge base `6d8be46ee23a72b1bcd8db99e2a9946dcfd22e99`. The partition diffs were
generated under ignored `logs/review/039-s/`. A coordinator verification
confirmed that the saved primary diffs exactly equal the Git diffs, that the
P2 split reconstructs all 38 `writepath_test.go` hunks (879 + 867 changed
lines), and that the partition union covers all 46 changed paths.

The explicit coverage manifest was:

| Partition | Coverage acknowledgement |
|---|---|
| P1: `writepath.go` full file and diff (7 primary hunks) | Anchor `gpt-6.1-sol`, T2 `claude-sonnet-5`, and T3 `claude-opus-5`; anchor and T2 also acknowledged the full file. |
| P2a/P2b: `writepath_test.go` (38 hunks, split at hunk boundaries) | Anchor and T2; T3 cross-cutting reviewer explicitly mapped all 38 hunks, including all three new ORCH-D4 regression tests. |
| P3: extent tests, oracle tests, golden JSON (17 primary hunks) | T2; T3 explicitly acknowledged all 8 zero-context oracle hunks in the cross-cutting diff. |
| P4: retiredarch mask tests and precondition script (9 hunks) | T1 `gpt-5.4-mini`, including the final diagnostic hunk. |
| P5: all 18 fixture files | T1, with an explicit acknowledgement for each fixture hunk. |
| P6: `.backlogit` and docs (21 files) | T1, including the complete backlog and memory partition. |
| Cross-cutting: production, test, and oracle diffs | Distinct T3 reviewer explicitly acknowledged 36 production hunks, all 38 test hunks, and all 8 oracle hunks by file and hunk. |

Four independent reviewers across tiers returned results. The requested anchor
reasoning-effort attestation was unavailable:
`TOOL_DEGRADED: anchor-review-model (reasoning effort unattested)`. Three
other reviewers were available, so the pool minimum was met. P1 was reviewed
by the anchor and at least two other reviewers.

### Blocking finding

The anchor identified a P1 in the new forward-assignment regression source at
`writepath_test.go:1471,1477` (P2b H6): it declares `var root *pathsafe.Root`
and then assigns `root, err = pathsafe.NewRoot(".")`, but `NewRoot` returns
`(Root, error)`, not `(*Root, error)` (`internal/pathsafe/root.go:340`). The
test string parses and the repository builds because it is not type-checked;
it does not model a compiling caller. The finding is newly introduced,
in-scope under P-021 C1, and classified P1 by the anchor. T2 agreed on the
mismatch but rated it P3 fixture fidelity, so independent severity agreement
was low. The conservative P1 disposition is retained. Under the operator's
terminal rule, no additional fix cycle is allowed.

Readiness: `BLOCKED`. No P0 was reported. The two production detector fixes
were confirmed by the anchor, T2, and T3; the blocking finding is the new
regression input's type mismatch, not persistence of the original two
false-negative behaviors.

Other findings were follow-up or advisory only: additional FuncDecl/FuncLit
boundary regression coverage (P2 plurality), the task-gated tests not running
in the ordinary suite (P2 unique), and a questioned split-selector line
attribution (P3 unique, no demonstrated defect). No fix was applied.

### Commit and scope verification

After the review report, Ship verified commit path scopes directly:

* ORCH-D1 `aeb3313576f1b1e00b500d1aa532560699ff0a5e` changes only
  `writepath_oracle_test.go`; the change replaces dynamic fixture discovery
  with the exact 19-name frozen corpus.
* ORCH-D2 `3094761ae82a5c351d18e3380b7c8f839e374150` changes only
  `writepath_extent_test.go` and `writepath_test.go` for the wrapper selector
  import guard.
* ORCH-D4 `8aca231566f98cab9ae62bf5234fd38992f57de8` changes only the two
  authorized Go files listed above.

## P-021 deferred-scope captures

The review's out-of-scope findings were not fixed. Active stash and indexed
deferred-entry searches were performed. Candidate `A0996E93` concerns only
same-name shadow precision and was not reused. The test-suite candidates
`2A1AD300`, `4803163C`, and `70BFB33D` concern related but differently sourced
049.003/049.007 default-suite expansions; their identity was not positively
confirmed as this 049.005 finding. Archived-stash lookup was unavailable, so
the entries below carry the required discovery marker. All captures are
capture-only, provisional, and were read back successfully:

* `B061EE34` — parenthesized NewRoot initializer/qualifier forms (bug, medium).
* `3F66C6AD` — dot-imported pathsafe NewRoot bindings (bug, low).
* `BE900FE7` — pointer/alias receiver propagation (bug, low).
* `43614FC7` — additional function-literal and function-boundary test coverage
  (task, medium).
* `667650F1` — move new task-gated invariants into the ordinary suite (task,
  medium); payload records `DISCOVERY-STATUS: AMBIGUOUS` with candidates
  `2A1AD300`, `4803163C`, and `70BFB33D`, plus
  `DISCOVERY-STATUS: LOOKUP-UNAVAILABLE`.

Every capture records task `049.005-T`, feature `049-F`, shipment `039-S`,
`PR=N/A` (genuinely pre-PR), `review-thread=N/A` (threadless), and
`requires deliberation=true`. The task comment and this run-level / pre-PR
closure-residual record cite all five IDs. No PR or closure artifact exists.

## Operational state and next decision

* Branch remains `feat/migrate-write-path-scanner-to-go-ast-post-m4-unit-e`;
  no other worktree was present.
* The lifecycle topology gate passed for active shipment `039-S`; P-001 found
  no other active top-level feature or chore.
* No feature branch push, PR, Copilot review, CI, merge, shipment reconcile,
  closure artifact, or shipment archive occurred.
* Shipment `039-S` remains active and feature `049-F` remains active.
  Post-merge closure, compaction status, and closure PR are **N/A** because no
  feature PR was opened or merged.
* Backlogit MCP was unavailable; CLI fallback and `backlogit sync` succeeded.
  Agent-intercom was unavailable (chat visibility only). Engram remains
  degraded and was not restarted. Archived-stash lookup remains unavailable.

Handoff: preserve this branch and stop. The operator/Orchestrator must
disposition the in-scope P1 and explicitly authorize any separate work unit
before code, PR, or shipment closure resumes. ORCH-D4's single additional
cycle is spent.
