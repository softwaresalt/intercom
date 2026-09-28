---
title: "Post-merge closure: 028-S / 031-F — Merge-strategy structural verification gate (plan unit 4)"
description: "Post-merge closure artifact for shipment 028-S / feature 031-F"
status: "complete"
tags:
  - "operational-closure"
  - "028-S"
  - "031-F"
  - "post-merge"
date: 2026-09-28
mode: post-merge
shipment: 028-S
feature: 031-F
pr: 75
merge_commit_sha: 8c5deba9f811b69dacca97c287f203a10a9f36df
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 028-S / 031-F — Merge-strategy structural verification gate (plan unit 4)

## Summary

Shipment `028-S` delivers `031-F`, "Merge-strategy structural verification gate". It adds
`scripts/check-merge-strategy.sh`, an advisory structural checker. The checker asserts that the repository's
`allow_squash_merge` and `allow_rebase_merge` settings are both `false`, which makes P-009 (merge-commit only) a
verified property rather than a convention. It has three verdicts:

* `PASS`: both settings are `false`.
* `FAIL`: either setting is `true`.
* `SKIP`: the settings cannot be read, or a value is not a boolean.

The shipment has three parts:

* **CI wiring**: a `merge-strategy structural verification` job in `.github/workflows/ci.yml`, wired into `ci gate`
  `needs:` and the LOCAL DIVERGENCE enumeration.
* **Fixtures**: 8 JSON fixtures under `scripts/testdata/mergestrategy/`.
* **Operator documentation**: `docs/merge-strategy-gate.md`, covering the GITHUB_TOKEN feasibility finding, the
  operator trigger, and the advisory-to-required promotion path.

This is the verification half only. Changing the repository settings is operator-only. No production `cmd/` or
`internal/` Go code changed.

The PR also carried an operator-authorized one-off exception under P-021 C4. Commit `fdf075f` resolves stash
`DCD67C30`: it renames the 027-S repair closure artifact to the dated non-shipment form, which unblocked
`TestOperationalClosurePostMergeFilenameConformance`.

## Invariants to preserve

* A value that is missing, null, a string, or any other non-boolean type yields `SKIP` and never `PASS`. An empty
  evaluator verdict exits non-zero.
* The CI job is advisory by default and blocks nothing while it reports `SKIP` under the default `GITHUB_TOKEN`.
* The workflow declares no invalid `permissions:` scope. There is no `administration` key; actionlint stays clean.
* The live scan passes the API response to the evaluator through a temp file, not stdin, because the program heredoc
  occupies stdin.
* The LOCAL DIVERGENCE header lists the new job, so `autoharness tune` does not drop it.

## Validator evidence

Per `.autoharness/workspace-profile.yaml` `runtime_validation.validator_manifest`, the only runtime surface declared
for this repository is `cli` (`go run ./cmd/... --help` smoke). This shipment does not touch `cmd/intercom` or
`cmd/intercom-ctl`, so the CLI probe does not apply to this diff and no probe evidence is recorded for it.

The surface that did change is the CI gate script. It was exercised directly on PR #75 at reviewed HEAD `b59fb31`
(no code changed after `32519f5`):

* `gofmt -l .` → no output; `go vet ./...` → exit 0; `go build ./...` → exit 0
* `go test -count=1 ./...` → all packages `ok`. This includes `TestOperationalClosurePostMergeFilenameConformance`,
  which is green again after `fdf075f`.
* `bash scripts/check-merge-strategy.sh --self-test` → 16 fixture checks and 2 exit-code checks pass
* `actionlint .github/workflows/ci.yml` → exit 0
* A live `scripts/check-merge-strategy.sh` run with an admin user token → `PASS` (both settings `false`)
* All 14 CI checks were green at merge HEAD `b59fb31`, including `ci gate`. The `merge-strategy` job printed the
  expected `SKIP` under `GITHUB_TOKEN`.
* The merge used a merge commit: `8c5deba9f811b69dacca97c287f203a10a9f36df`. The repository allows merge commits
  only (`allow_merge_commit=true`, `allow_squash_merge=false`, `allow_rebase_merge=false`), which satisfies P-009.
  `gh pr merge 75 --merge --match-head-commit b59fb31…`, no `--admin`.
* P-018 Copilot-review gate: `NOT_APPLICABLE` (no engagement, enforcement `auto`). It was re-run at the last mile
  before the merge.
* P-016 pipeline-topology: `pass` at the pre-merge lifecycle check and at the pre-closure lifecycle check on
  `post-merge/028-s-031-f-merge-strategy-gate` (`BRANCH_POST_MERGE_CLOSURE_ELIGIBLE`).
* Merge confirmation: `gh pr view 75` → `MERGED` at `2026-09-28T05:59:10Z`, and
  `git merge-base --is-ancestor 8c5deba origin/main` → exit 0.

Verdict: `PASS_WITH_FOLLOW_UP`. The follow-ups are listed below.

## Pre-deploy audits

None apply. This shipment has no migration, feature flag, runtime config, or access change. Wiring a PAT or App
secret would be an access change, and it is deferred to `124AE9DE`.

## Deployment / rollout path

Merge-only. The job runs on every future PR and push to `main`. There is no separate deploy step.

## Post-deploy checks

* On the next PR, confirm that the `merge-strategy structural verification` job runs and reports `SKIP` with the
  "cannot read settings" reason, and that `ci gate` stays green.
* Occasionally run the live checker with an admin token as a manual spot check:
  `GH_TOKEN=<admin> bash scripts/check-merge-strategy.sh` → expect `PASS`.

## Risky action record

No destructive, irreversible, or elevated-privilege action was taken.

* **Merge**: operator-approved (P-014, "PR 75: Merge approved"), merge commit only, bound to HEAD `b59fb31`.
* **Covering-feature completion (a1, S4)**: all five conditions held, so `backlogit move 031-F --status done` moved
  `031-F` `active -> done` (exit 0; re-read `done`).
* **Shipment closure**: `shipment-reconcile` `classify-close-path` returned `CASCADE` / `FULLY_COVERED_ROOT` with
  binding `a0b976c8…`. The binding was revalidated against live state before any mutation. Safe-close then delegated
  to the Cascade Close Sub-Procedure: `backlogit shipment ship 028-S --sha 8c5deba…`.
  * `returned_ids=[]`
  * `archived_ids = allowed_ids = required_ids` = {028-S, 031-F, 031.001-T..031.005-T}
  * `parent_id` was preserved on all five tasks
  * `028-S` `archived_status: shipped`

  See `.backlogit/reconcile/028-S-safe-close-20260928T060059Z.md`.

## Healthy signals

* The `merge-strategy` job reports `SKIP` under `GITHUB_TOKEN` (or `PASS` once `124AE9DE` wires a secret), and
  `ci gate` stays green.
* The manual admin-token spot check keeps returning `PASS`.

## Failure signals

* The job reports `PASS` under the default `GITHUB_TOKEN`. That token cannot read the settings, so a `PASS` would
  mean the unreadable-to-PASS guard has regressed.
* The job reports `FAIL`, which means someone enabled squash or rebase merging in the repository settings.
* GitHub rejects the workflow, for example because of an invalid permission scope. That would block every CI run.

## Monitoring plan

Watch the standard CI dashboard on `main`, specifically the `merge-strategy structural verification` job output on
the next several PRs (029-S onward).

## Rollback trigger

The job breaks workflow parsing or blocks unrelated PRs, or its verdicts prove to be wrong.

## Rollback procedure

Revert merge commit `8c5deba9f811b69dacca97c287f203a10a9f36df` through a standard merge-commit revert PR. There is no
data or state migration. Alternatively, drop the job from `ci gate` `needs:` while keeping the script.

## Validation window

The next 2–3 PRs on `main` (029-S and later shipments in the gate-reliability plan).

## Owner

The Ship agent and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (stash; Stage-owned triage)

* `150364D2`: 3 P3 checker-hardening items (trap cleanup, exit-code table self-test, running outside a git checkout).
  Deferred when the review-fix circuit breaker was reached.
* `124AE9DE` (P-021 deferred scope expansion): wire a PAT or GitHub App secret so CI can read the settings.
* `C8914513` (P-021 deferred scope expansion): required mode should fail closed on `SKIP`.
* `C98B92F0` (P-021 deferred scope expansion): closure artifact `mode` vs. filename conformance
  (`DISCOVERY-STATUS: AMBIGUOUS DCD67C30`).
* `DCD67C30`: resolved by operator-authorized exception `fdf075f`. Stage should archive it; Ship did not modify the
  entry.

No new follow-ups came out of post-merge closure.

## Source artifact cleanup

* `031-F` has no `custom_fields.source_stash_id`. Its description cites "Source: stash B6EF23CC", but Step 7 keys
  only on the structured field, so the entry was skipped and left for Stage to handle. Ship changed no stash entries.
* `031-F` has no `custom_fields.source_deliberation_id`. Skipped.
* Archived: none. Skipped: `B6EF23CC` (no structured link).

## Knowledge graduation

* No update was needed for `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/`, or `docs/product-specs/`: the
  change is limited to CI tooling, and `docs/merge-strategy-gate.md` shipped with the PR.
* Compound capture: `docs/compound/2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md` (the review
  needed 3 cycles).
* `compound-refresh`: not needed. No existing `docs/compound/` entry is superseded.

## Compaction status (P-020)

`done`. Ship Step 6 item 8 ran `compact-context` with `target: all`.

* **Memory**: the two 028-S-scoped checkpoints (`2026-09-21-ship-028-s-halted-pre-existing-test-failure.md` and
  `2026-09-27-ship-028-s-pr75-open-awaiting-merge-approval.md`) were compacted into
  `docs/memory/compacted/2026-09-28-028-s-031-f-compacted.md`. The originals moved to `docs/archive/memory/`.
* **Plans**: no consolidation. The gate-reliability plan still has outstanding units (029-S through 033-S).
* **Closure records**: no compaction. Every closure record is younger than 14 days.

## Releasability evidence

**READY.**

* Local review readiness: `READY_WITH_FOLLOWUPS` (P0=0, P1=0) at reviewed HEAD `b59fb31`, unchanged through the
  merge.
* All 14 CI checks were green at merge HEAD, and the merge used a merge commit (P-009).
* P-018 `NOT_APPLICABLE`; P-016 passed; P-014 operator approval was recorded.
* No runtime service surface changed. The gate script's self-test, actionlint, and the live admin-token scan are all
  green.
* The residual follow-ups are advisory and do not block release. They are recorded above for Stage.
