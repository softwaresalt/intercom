---
title: "Pipeline-topology gate — CI rollout"
description: "Threat model and staged advisory-to-required adoption of the topology CI backstop."
date: 2026-10-06
status: operational
tags:
  - ci
  - pipeline-topology
  - rollout
---

# Pipeline-topology gate — CI rollout

## Rollout objective

The `topology-check` job is an ambient CI backstop for the P-001 single-active-release-unit and
P-016 single implementation branch/worktree invariants. It validates the synchronized backlog
and the branch under CI. It is a detector at sync time, not a lock or a distributed lease; another
checkout whose changes have not reached CI can still race.

The repository variable `PIPELINE_TOPOLOGY_GATE_REQUIRED` controls whether a reported topology
`BLOCK` fails the CI job. Leave it unset for advisory observation. Set it to `true` to make a
gate `BLOCK` fail the step, job, and aggregate `ci gate` without requiring a workflow re-render.
The workflow's GitHub Actions string comparison is case-insensitive, so case variants of `true`
also select required enforcement. This is a staged enforcement control, not a change to the gate's
own verdict.

## Staged advisory-to-required rollout

1. **Observe in advisory mode.** Keep `PIPELINE_TOPOLOGY_GATE_REQUIRED` unset. Review the
   `pipeline-topology (ambient)` result on ordinary pull requests and pushes. A `BLOCK` remains
   visible in the step output even though the CI job can proceed.
2. **Triage before promotion.** For every `BLOCK`, retain its JSON output and reconcile the
   reported active-shipment or branch condition through the authorized shipment workflow. Treat
   `INVALID`, missing-tool, ambiguous backlog-root, and wrapper setup failures as configuration
   defects, not topology findings. They remain visible in logs; because the advisory
   `continue-on-error` is attached to the entire wrapper step, these nonzero exits are also
   non-blocking for the job until required enforcement is enabled.
3. **Promote deliberately.** After the workspace owner confirms operators can diagnose and
   remediate the observed output, set repository variable `PIPELINE_TOPOLOGY_GATE_REQUIRED` to
   `true`. Keep the branch-protection requirement on the aggregate `ci gate` job, not an
   individual topology step.
4. **Verify promotion.** Open a controlled change or run a normal CI event; confirm a passing
   topology verdict succeeds and a genuine topology `BLOCK` fails the job and aggregate check.
   Revert the repository variable to an unset advisory posture if the required check exposes an
   operational blocker; record the reason and the follow-up before another promotion attempt.

`continue-on-error` is attached to the entire `bash scripts/ci-topology-check.sh` step, not only
to a topology verdict. In advisory mode, a missing `autoharness` binary or invalid/ambiguous
backlog configuration still emits a visible wrapper error, but its nonzero exit does not fail the
job. Required mode makes any nonzero wrapper exit fail the step, job, and aggregate `ci gate`.
Checkout and pinned dependency installation occur in earlier separate steps and remain hard
failures in either posture. This distinction is important when promoting the toggle: advisory mode
does not enforce wrapper/configuration failures.

## Threat Model & CODEOWNERS Hardening

The current workflow uses GitHub's `pull_request` event. For that event, the proposed pull-request
head supplies the workflow content that CI executes. Therefore a malicious or compromised pull
request that can edit the enforcement workflow, the CI topology-check entrypoint, or its relevant
checker surface may also weaken or bypass the check in the same change. This CI job is useful for
catching accidental or careless local bypasses that do not intentionally modify the enforcement
surface; it is not a non-bypassable security boundary against a hostile PR.

Recommended defense-in-depth:

- Add a `CODEOWNERS` rule covering `.github/workflows/ci.yml`,
  `scripts/ci-topology-check.sh`, and any in-repository policy or configuration that controls this
  CI enforcement surface.
- Require an approval from the responsible trusted owner for changes matching those rules, and
  verify the review requirement is enforced by repository rules/settings rather than relying on
  the file alone.
- Keep CI permissions least-privileged, and review changes to the job, entrypoint, pinned
  installation inputs, and aggregation dependencies as security-sensitive.
- Re-check the rules and effective required status when the enforcement surface changes; a
  CODEOWNERS declaration without a matching required-review policy is not enforcement.

A stronger architecture could move validation to a trusted `pull_request_target` workflow that
checks out and inspects the proposed content without executing its workflow or scripts. That is
an alternative design with a different trust boundary and careful token/checkout risks; it is
outside this documentation-only rollout and is not enabled by this job.

## Post-merge closure evidence

Post-merge release closure is recorded at
`docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`. This is the repaired,
shipment-and-feature-addressable naming convention for post-merge closure artifacts. Do not
confuse it with the separate dated naming forms used for other closure-related reports.
