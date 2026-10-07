---
title: "Pipeline-topology gate — operator runbook"
description: "Observable behavior, triage, and enforcement controls for the pipeline-topology gate."
date: 2026-10-06
status: operational
tags:
  - ci
  - pipeline-topology
---

# Pipeline-topology gate: operator runbook

## Purpose

The pipeline-topology gate checks the shipment and workspace conditions that protect:

- **P-001:** only one top-level release unit is active at a time.
- **P-016:** Ship implementation work stays on one shipment-owned branch and one implementation worktree.

The gate is provided by the external `autoharness` command. This runbook documents its
repository-visible invocation and observable results; it does not assert or reproduce an
internal file-selection pattern. Treat the emitted result as authoritative for the invocation
that produced it.

The CI entrypoint runs an ambient check against the committed, synchronized backlog and branch.
It is a re-validation point, not a distributed lock or lease: it cannot serialize claims made
from another checkout whose state has not yet reached CI. The runner's ephemeral worktree is not
used as evidence about a developer's local worktree topology.

## Where the gate runs

The GitHub Actions job **`pipeline-topology (ambient)`** invokes
`scripts/ci-topology-check.sh`. That script runs the gate in `ci` / `ambient` mode and reports a
server-side signal for the P-001/P-016 invariants. The repository's aggregate **`ci gate`** is the
check intended for branch-protection configuration.

For shipment execution, an operator or Ship can run the local, shipment-scoped gate:

```powershell
autoharness gate pipeline-topology --mode agent --shipment <shipment-id> --phase <phase> --json
```

Use the phase requested by the Ship workflow (for example, `pre_claim` before a claim,
`post_claim` immediately after it, or `lifecycle` before lifecycle work). A local result is
shipment- and worktree-aware; do not substitute a CI ambient result for a required local phase.

## Reading a failure

The CI entrypoint prints the resolved backlog root and the command it runs. For a gate response,
inspect the JSON fields `phase`, `target_shipment_id`, `exit_code`, `blocked`, `invalid`, and
`checks`. Each check reports a `name`, `status`, optional `token`, `message`, and `details`.
Preserve this output when handing a failure to the workspace owner; the phase, token, and details
identify what needs attention without guessing at hidden selection rules.

Observable outcomes are:

| Result | Meaning | Operator response |
|---|---|---|
| Exit `0`, `blocked: false`, `invalid: false` | The invoked gate passed. | Continue only with the next step allowed by the workflow. |
| Exit `1`, `blocked: true` (or CI text `BLOCK`) | A checked topology invariant did not hold. | Use the reported check, token, and details to reconcile shipment ownership, active work, branch, or worktree state. |
| Exit `2`, `invalid: true` (or CI text `INVALID`) | The invocation or gate configuration is invalid. | Correct the invocation or installed configuration; do not treat this as a topology pass. |

The CI wrapper also emits an `::error::` message and exits nonzero if `autoharness` is missing,
the backlog-root configuration is invalid or ambiguous, or another wrapper prerequisite fails.
These are wrapper/configuration failures, not topology verdicts. They remain visible in the logs,
but the workflow's advisory `continue-on-error` is attached to the entire wrapper step, so it also
allows these nonzero exits without failing the CI job. The required posture makes them fail the
step and job. Checkout and pinned dependency-installation failures occur in earlier, separate
steps and remain job failures in either posture.

## Operator remediation

1. Save the complete failing output, including the phase, target shipment, token, check name, and
   details. Do not infer a different target from a branch name or a guessed path pattern.
2. For an active-shipment conflict, identify the currently active top-level release unit and
   finish or reconcile that unit through the approved Ship/Stage workflow before starting another.
   Do not silently change unrelated shipment or task state to make the gate pass.
3. For a branch-ownership result, use the branch associated with the target shipment. Resolve a
   mismatch by checking out or creating the correct shipment branch; do not continue implementation
   on an unrelated branch.
4. For a worktree-topology result, stop Ship execution and verify the other worktree's owner and
   state. Ship must not create or use a parallel implementation worktree. Do not delete a worktree
   that may contain uncommitted work; obtain the owner's disposition first.
5. For an `INVALID` result or a wrapper setup error, repair the invocation, backlog-root
   configuration, or CI installation. Do not reinterpret a missing tool or malformed state as
   `PASS`.
6. Re-run the same required gate phase after remediation and retain the passing output with the
   execution record. Do not use an override to bypass a failed safety check.

## Advisory and required enforcement

The repository variable **`PIPELINE_TOPOLOGY_GATE_REQUIRED`** controls whether a topology
`BLOCK` fails the CI job. The workflow compares its value with the string `'true'` using
GitHub Actions expression semantics, which compare strings case-insensitively:

- **Unset or any value not equal to `true` ignoring case (default advisory posture):** the
  topology-verdict step reports its result, but `continue-on-error` allows the job to remain
  successful on a `BLOCK`.
- **A value equal to `true` ignoring case (required posture):** a topology `BLOCK` fails the step
  and job, so the aggregate `ci gate` fails. Setting lowercase `true` is the documented operator
  setting; case variants also compare equal.

This toggle changes the CI job's treatment of nonzero exits from the wrapper step; it does not
change the gate's own `PASS`/`BLOCK` result. In advisory mode, wrapper `BLOCK`, `INVALID`, missing
tool, and wrapper-configuration/setup failures are reported but do not fail the job. In required
mode, any such nonzero wrapper exit fails the step and job. Checkout and pinned dependency
installation happen in separate earlier steps and remain failures in either posture. Changing
the repository variable takes effect without re-rendering the workflow.

## Post-merge closure evidence

Use the shipped post-merge closure artifact naming convention when recording release closure:
`docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`. The shipment and covering-feature
IDs make the closure record directly traceable to the release unit. This form is specific to
post-merge closure; other closure-related reports can have different names.

## Local diagnostics retention

Keep transient local diagnostic output under an ignored workspace directory, not in the repository
root. Keep maintained diagnostic tooling under `scripts/`; the root is not a scratch space and a
tooling script must not be duplicated there. The existing `.autoharness/.gitignore` comment and
`staging/` rule document the local staging boundary; retain that ignore configuration rather than
adding generated diagnostic output to version control.
