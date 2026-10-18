---
title: "Stage contract divergence ledger (050-S, task 020.004-T)"
date: 2026-10-17
status: "active"
source: "docs/plans/2026-10-17-intercom-go-020f-stage-contract-enforcement-plan.md"
backlog_task: "020.004-T"
---

This ledger records the intentional divergence between the autoharness-generated Stage contract files and this repository's CI. It also records the limits of the gate that enforces the divergence.

## Divergent surfaces

| Surface | Divergence | Properties |
|---|---|---|
| `.github/policies/workflow-policies.md` | Carries the corrected Stage default-branch contract. The generated baseline does not. | P1 to P6 |
| `.github/agents/_stage.agent.md` | Carries the corrected Stage default-branch contract, including the Step 1.9 checklist registration. The generated baseline does not. | P1 to P6 |
| `.github/workflows/ci.yml` | Job `lint` runs the blocking step "Run stage contract live gate". The generated baseline has no such step. | P1 to P6 |

P1 to P6 are the closed set of corrected-contract properties defined in the plan's "Named corrected-contract properties" section. The detector in `tools/gatecheck/internal/stagecontract/` evaluates only that set. A regeneration that reverts either contract file makes `TestLiveContract` fail.

## The gate

The step runs `go test -json -count=1 -mod=readonly -run '^TestLiveContract$' ./tools/gatecheck/internal/stagecontract/` and then requires a JSON `pass` event for `TestLiveContract` through `jq -e`. A run with no tests, or a skipped test, emits no `pass` event and fails the step. The step is hermetic, has a timeout, no `if:` guard and no `continue-on-error`. `tools/gatecheck/ciwiring_test.go` pins this shape.

## Limits

* **TAMPER-EVIDENT, NOT TAMPER-PROOF.** A `pull_request` workflow runs from the PR head. A single PR can change the contract files, the detector and the workflow together and still pass. The gate reliably catches accidental regeneration reverts only.
* **Skipped lint skips the gate.** When the `changes` job reports no code change (docs-only, backlog-only or `.autoharness`-only PRs), job `lint` is skipped and so is this step. This is accepted.
* **CODEOWNERS is advisory.** `.github/CODEOWNERS` lists the two contract files and `tools/gatecheck/internal/stagecontract/`, but required code-owner review is not enabled.

## Operator-only follow-up

Real enforcement needs one of the following, both outside this task:

* branch protection or a ruleset that requires code-owner review on `main`, after an approval path for unattended PRs exists, or
* a base-ref required workflow that runs the detector from the default branch rather than the PR head.

## Rollback

Rollback is coordinated. Never revert the `ci.yml` step alone: the wiring assertions in `ciwiring_test.go` would fail, and the live-contract and mutation-proof tests would assert a gate that no longer exists. Either:

1. revert the step together with its `ciwiring_test.go` assertions and the dependent detector tests in one change, and update this ledger to record the gate as removed, or
2. revert the whole 050-S implementation release as a unit.

Both must leave `go test ./...` green.
