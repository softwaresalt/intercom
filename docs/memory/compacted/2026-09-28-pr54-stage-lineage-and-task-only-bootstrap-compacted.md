---
title: "Compacted memory - PR54 Stage lineage, task-only finalization, and true-lineage cycles"
date: 2026-09-28
status: compacted
pr: 54
release_units:
  - 017-S
  - 021-S
features:
  - 018-F
  - 022-F
source_count: 21
compacted_from:
  - docs/archive/memory/2026-09-11-stage-artifact-branch-pr-policy-gap-session.md
  - docs/archive/memory/2026-09-11/copilot-stage-review-request-incident.md
  - docs/archive/memory/2026-09-12-stage-bootstrap-replan-rev2.md
  - docs/archive/memory/2026-09-12-stage-replan-018F-017S.md
  - docs/archive/memory/2026-09-12-stage-rev17-adversarial-remediation.md
  - docs/archive/memory/2026-09-12-stage-rev18-adversarial-remediation.md
  - docs/archive/memory/2026-09-12-stage-rev5-adversarial-remediation.md
  - docs/archive/memory/2026-09-12-stage-rev6-adversarial-remediation.md
  - docs/archive/memory/2026-09-12-stage-task-only-shipment-finalization-session.md
  - docs/archive/memory/2026-09-12-stage-three-shipment-bootstrap-replan.md
  - docs/archive/memory/2026-09-13-stage-a-only-exceptional-cycle-4.md
  - docs/archive/memory/2026-09-13-stage-a-only-final-review-remediation-cycle-1.md
  - docs/archive/memory/2026-09-13-stage-a-only-remediation-cycle-2.md
  - docs/archive/memory/2026-09-13-stage-a-only-remediation-cycle-3-escalation.md
  - docs/archive/memory/2026-09-13-stage-a-only-simplification.md
  - docs/archive/memory/2026-09-13-stage-compound-root-remediation-attempt-8-candidate.md
  - docs/archive/memory/2026-09-13-stage-recovery-torn-state-halt.md
  - docs/archive/memory/2026-09-13-stage-true-lineage-attempt-5-fail.md
  - docs/archive/memory/2026-09-13-stage-true-lineage-attempt-6-reauthor.md
  - docs/archive/memory/2026-09-13-stage-true-lineage-attempt-7-option-a.md
  - docs/archive/memory/2026-09-14-stage-true-lineage-attempt-8-gate-failed.md
preserved_latest_checkpoints:
  - docs/memory/2026-09-12-stage-taskonly-replan-and-harvest-session.md
  - docs/memory/2026-09-14-stage-operator-accepted-residual-disposition.md
---

# Compacted: PR54 Stage lineage, task-only finalization, and true-lineage cycles

## Outcome

The archived records preserve the long Stage lineage that led to PR #54 and the later task-only
shipment finalization work. The final shipped state is already summarized in
`docs/memory/compacted/2026-09-16-pr54-stage-pipeline-policy-gap-compacted.md`: PR #54 merged via
merge commit `f7112e0693804ac75cd4511d2ee3eaeac15583e4` and staged the 021-S / 022-F work without
claiming it. This compaction keeps the deeper Stage reasoning that would otherwise be lost.

Two most recent checkpoints were intentionally left in `docs/memory/` rather than archived:
`2026-09-12-stage-taskonly-replan-and-harvest-session.md` for the unmerged
`chore/stage-task-only-shipment-finalization` branch lineage, and
`2026-09-14-stage-operator-accepted-residual-disposition.md` for the last operator-disposition
record in the PR54 true-lineage chain.

## Stage artifact branch and PR policy gap

The 126 KB session record documents the evolution from a simple authorization correction into a
full producer/consumer verification contract. The durable conclusions are:

* The root problem was not only whether Stage may write artifacts, but whether the actor that must
  verify them can observe the artifact commit on the correct branch. Authorization without a
  discoverable handback deadlocks the Orchestrator.
* Stage must produce explicit handback fields: `stage_branch`, `stage_head_commit`,
  `stage_outcome`, `stage_artifact_paths`, and `shipment_id` when formed. Defaults degrade without
  wedging old sessions, but verification fails closed when the path set is empty or unreadable.
* Orchestrator verification for no-shipment Stage runs needs both ancestry and concrete file
  readability on `origin/main`. Directory prefixes are insufficient; `git cat-file -t` or an
  equivalent blob-type check is required to distinguish files from trees.
* Branch discovery must inspect the checked-out Stage branch (`origin/main..HEAD` after a verified
  branch assertion), not only `git status -- .backlogit/` or a default-branch range.
* Step ordering matters. A branch gate after artifact writes or a commit after summary emission is
  too late; the Stage branch gate must precede tracked artifact mutation, and artifact commit must
  precede handback.

The repeated defect class was producer/consumer drift: values were consumed before any ordered,
permitted surface produced them, or a consumer silently substituted a success state for a degraded
handoff. The final lineage hardened this through explicit production steps, pair validation, and
fail-closed verification rather than prose-only permission.

## Harness and decomposition lessons

The PR54 lineage repeatedly exposed that a multi-task harness must encode terminal-state truths,
not transient task-local truths:

* A harness assertion pinned to a transitional implementation state goes red later and wedges the
  default full suite. Per-task selectors are acceptable only when the default suite is green at each
  completed boundary.
* Activation predicates and guard probes must not read the same surfaces the harness later asserts
  on. MP5 was retired because it failed in legal mid-shipment states.
* Terminal state needed a tracked manifest and independent terminal witness, with MP6 validating
  agreement before activation. The residual that a coordinated rollback of both files is review-only
  evidence was recorded honestly rather than overclaimed.
* Later adversarial review overturned Stage's decomposition verdict: C2 fixture infrastructure and
  B3 branch/handback/path-safety work were too wide for the existing task shape. The next correct
  action was a re-plan/decomposition cycle, not another micro-fix.

## Task-only shipment finalization and A-only simplification

The archived task-only finalization sessions established the following settled facts:

* The original Ship route from the restored checkpoint was void: the referenced `018.004-T` series
  no longer existed, and live `017-S` had a one-task manifest at that point.
* Disposable backlogit probes proved `move <shipment> --status shipped` exits 9 for the task-only
  blocker; task-only close can preserve a parent when the manifest is exact, but omitted queued
  subtasks can be silently archived and omitted siblings produce non-empty `returned_ids`.
* Branch topology blocked safe harvest from a branch cut before the policy-gap artifacts reached
  `main`, because `.backlogit/` IDs are branch-scoped and would allocate conflicting `018-F` /
  `017-S` records.
* The three-shipment bootstrap was simplified to `021-S` -> `017-S`; B/C records were blocked or
  archived, and `TASK_ONLY_FINALIZE` became deferred platform work with no current consumer.

The compacted active checkpoint left in place records the final task-only harvest/replan state for
future continuation.

## True-lineage and operator-disposition lessons

Attempts 5 through 8 showed the same contract defect relocating across surfaces instead of closing:
producer, consumers, authority grant, refusal path, tests, and probes were not verified against the
same bound snapshot. The root remediation produced the compound learning
`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`.
Key durable rules:

* A local zero-occurrence result or a probe scoped to the artifact where the finding was reported is
  not proof of cross-artifact closure.
* Every pinned needle must be a verbatim substring of the bound wording unless the plan explicitly
  authorizes normalization.
* When a finding reappears under a new label, stop patching and build the full surface matrix.
* Operator-accepted residuals must be recorded as a distinct state, never relabelled as a gate PASS.

## Files and artifacts modified in the original sessions

* Stage planning, deliberation, review, and memory artifacts for PR54 / 017-S / 018-F and 021-S /
  022-F
* Backlog records for `017-S`, `018-F`, `021-S`, `022-F`, and `022.001-T`
* Dedicated compound learning for cross-artifact contract closure
* PR body/readiness updates for blocked or operator-disposition states, with no Stage-owned push or
  GitHub thread resolution where authority belonged to the Orchestrator/operator

## Open or preserved next steps

* The preserved task-only checkpoint remains the safest resume point for the unmerged local
  `chore/stage-task-only-shipment-finalization` topic.
* The preserved operator-accepted residual checkpoint remains the latest local summary of the true
  lineage before Ship can execute the staged work.
* Any future re-plan must rebuild task/function mapping, shipment membership, dependency edges, and
  two-axis sizing instead of continuing the rev-15 micro-fix loop.
