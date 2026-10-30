---
type: circuit-breaker
timestamp: 2026-10-30T04:03:35Z
agent: "Orchestrator"
skill: "pr-lifecycle"
breaker_type: universal
operation: "go test (contained policy snapshot validation)"
attempts: 3
identity: "policy-4dcf82e-tools-gatecheck-temp-containment"
---

# Circuit Breaker - Policy Snapshot Validation

## Failure Chain

### Attempt 1

* Exit: 1
* Target: `go test ./...` in the immutable policy snapshot
* Cwd: `logs\validation\policy-4dcf82e`
* Phase: final policy-correction validation before PR creation
* Stable error: `temporary directory is outside workspace`
* Evidence: validator agent `cd89c451-ed0f-4d05-b8f7-0622ce46ecab`, turn 0
* Scratch was a sibling of the repository-under-test, inside the actual workspace.
* Five assertion failures occurred within one native invocation.

### Attempt 2

* Exit: 1
* Target: the `tools/gatecheck` package in complete-package validation
* Cwd: the same immutable snapshot
* Phase: final policy-correction validation before PR creation
* Stable error: `temporary directory is outside workspace`
* Evidence: the same validator agent, turn 1
* Scratch was inside the snapshot, but the containment failures remained.
* Five assertion failures occurred within one native invocation.

### Attempt 3

* Exit: 1
* Target: `go test .`
* Cwd: `logs\validation\policy-4dcf82e\tools\gatecheck`
* Phase: third and final counted validation attempt
* Stable error: `temporary directory is outside workspace`
* Evidence: the same validator agent, turn 2
* The validator reported initial TEMP/TMP/TMPDIR and managed scratch beneath the
  exact package directory. The same five tests still failed.
* No fourth attempt is authorized.

## Affected Tests

* `TestCheckRetiredArchitectureWrapper_WrongTreeRootAnchoring`
* `TestCheckWritePathWrapper_LegitimateModesReachBuild`
* `TestInstallGoShim_UnexpectedInvocationFailsClearly`
* `TestCheckWritePathWrapper_WrongTreeRootAnchoring`
* `TestSameBashFile_ComparesByIdentity`

## Context

* Corrective branch: `fix/dark-staging-publication`
* Pushed corrective HEAD: `4dcf82efeb8549a1c1aea3bb466fccb659a81ab8`
* Scope: explicit operator-authorized P-010/P-017 and Stage/Orchestrator handoff
  correction, with focused Go contract regressions
* Focused RED/GREEN, format, vet, primary full build, and independent final-HEAD
  review passed. These do not replace the incomplete full-suite gate.
* Complete inventory contains 21 packages. Twelve passed in the last
  package-specific validation; the seven remaining packages and integration
  group were not run after the root gatecheck failure.
* The final snapshot build was not run after validation stopped.
* No corrective PR was created or merged. Original shipment `073-S` was not
  claimed or implemented, and its staging branch remains preserved at
  `3a6f5e8ae5707436645e610d7f7d4f25d90093ac`.
* Primary worktree was clean before this checkpoint was added.
* The Stage artifact branch remains unpublished. Dark mode is halted; normal
  scoped merge authorization remains recorded, but no gate bypass is allowed.
* Engram readiness separately failed twice earlier. No additional readiness
  probe or model/transport workaround is authorized by this checkpoint.

## Evidence and Logging Controls

* Assertion counts were initially mistaken for invocation counts by the
  validator. The parent reconciled the record to two native failures before
  the third attempt. No native failure was erased or reset.
* The first two failures and the final failure share a concrete error and
  target family; the failure chain is not provisional.
* The validator created no raw output log. References are bounded agent
  summaries, not raw payload or environment captures.
* Any future capture remains limited to ignored, canonically contained
  `logs\diagnostics` paths, 1 MiB or 10,000 lines, and the command timeout.
  Inspection is limited to 64 KiB or 500 lines. Secrets and raw environment
  values must not be captured.
* No source, policy, fixture assertion, test selector, timeout, or containment
  implementation was changed during validation.
* The immutable test fixture is not an implementation worktree; it contains
  no development commits or PR actions. Its named ignored directories remain
  pending approved cleanup.

## Resolution

Circuit breaker triggered after three native failures. Further execution of
this validation operation requires explicit operator disposition. Inspection
of already captured evidence is allowed; it does not authorize another run.

The policy correction and original CI-reliability shipment are incomplete.
See `dark-staging-correction-memory.md` for the corrective implementation
history. Do not report the correction as merged or the original operation
as resumed.

## Method Repair Checkpoint (2026-10-30T04:53Z)

This is a separate, explicitly operator-approved operation
(`validation-method-temp-contract-20261030`, reply "ok, apply the recommended
fixes now"). It is not a fourth attempt. `attempts: 3` and the OPEN status above
are unchanged. The five tests, the full suite and `tools/gatecheck` were not run.

* One standalone diagnostic ran once, native exit 0, and replicated the
  `workspaceTempDir` helper from source. Details, line citations, evidence
  paths and hashes are in `validation-method-repair-memory.md`.
* In all five failing tests the helper's `t.TempDir()` is the first one of the
  test, and the helper overrides the initial TEMP, TMP and TMPDIR. The three
  attempts varied only those variables, so they could not have changed the
  outcome.
* The failure reproduces when GOTMPDIR points outside the package directory,
  or when an earlier `t.TempDir()` cached an outside parent. GOTMPDIR in the
  validator environment is the leading hypothesis and remains UNPROVEN: it was
  never captured.
* On this machine's PATH, `bash` is the WSL launcher, and the helper's gate
  would skip these tests. The validator failed rather than skipped, so its
  PATH differed. That too was not captured.
* No defect in the existing helper is proven. No fix was applied or proposed.
* Gate provenance: the constitution (Principle II and Quality Gates) and
  AGENTS.md require the full `go test ./...` before merge, and nothing read
  allows a silent waiver. Linux PR CI would run `go test -race ./...` but only
  after a PR exists, and whether it may count is an operator disposition.
* Further execution of the validation operation still needs explicit operator
  disposition, as stated in Resolution above.
