---
title: "039-S E-T5 Root.Resolve tripwire"
date: 2026-10-04
agent: ship
feature: 049-F
shipment: 039-S
task: 049.005-T
mode: dark (P-017)
status: completed
---

## Outcome

049.005-T is complete. The scanner now reports `.Resolve` calls only when
their receiver was bound from canonical `pathsafe.NewRoot` within the same
function. Cross-function flows, struct fields, package-level variables, and
method values remain residuals.

## Implementation

* Implementation commit: `fce2b4577a2f2679582e022dc665e0f9edea44ce`.
* Backlog archival commit: `d5e0b886893c7fb75fadb76ab4a097822e4f612b`.
* The positive fixture covers `:=`, plain `=`, and `var r, err =` two-name
  NewRoot bindings; all three calls produce findings.
* The unrelated receiver fixture remains clean, and golden/self-test captures
  are additive.
* 049.005-T moved to `done` and was archived. Its backlog commit field was not
  updated because Ship's Role Boundary does not authorize that unlisted
  backlog mutation; the commit subject and this record preserve the link.

## Verification

The harness was observed red before implementation and passed afterward.
`gofmt -l .`, `go vet ./...`, `go test ./...`, `go build ./...`, affected
package lint, and `git diff --check` passed. The full test suite required
`C:\Program Files\Git\bin` and `C:\Program Files\Git\usr\bin` on `PATH` so
Windows Git Bash wrappers could locate their tools.

## Decisions and constraints

AC-E2.6 input-freeze landed in a follow-up corrective commit
`aeb3313576f1b1e00b500d1aa532560699ff0a5e` rather than the E-T2 commit
`8aaafd7`; authorized by Orchestrator ORCH-D1. That commit changes only the
oracle fixture input source to the exact 19-name list from `c04d975`.

The 049.002-T backlog commit-tracking field was not updated because the Ship
Role Boundary does not authorize that unlisted mutation. Its commit subject
and the resolved halt record preserve the association.

The active checkpoint `checkpoint-20261004-073226.json` still has a stale
049.002-T resume hint. This same-session continuation used the verified live
branch/backlog state and ORCH-D1 rather than restoring that checkpoint; it
remains unresolved and must not be cleaned up without its owner-scoped
disposition.

## Next

Continue with 049.007-T, then 049.006-T. The local adversarial review, PR,
Copilot review, CI, merge, and post-merge closure gates remain outstanding.
