---
title: "Post-merge closure: 029-S / 032-F — Extract gate Python engines to importable modules (plan unit 5)"
description: "Post-merge closure artifact for shipment 029-S / feature 032-F"
status: "complete"
tags:
  - "operational-closure"
  - "029-S"
  - "032-F"
  - "post-merge"
date: 2026-09-28
mode: post-merge
shipment: 029-S
feature: 032-F
pr: 77
merge_commit_sha: 16519360c551ee44c0405fc436bb8b4a8718c4d5
compaction_status: pending
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 029-S / 032-F — Extract gate Python engines to importable modules (plan unit 5)

## Summary

Shipment `029-S` delivers `032-F`, "Extract gate Python engines to importable modules". This is Unit 2 of
`docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md` (AC-2.1..AC-2.8). What it did:

* Moved the Python engine of `scripts/check-retired-architecture.sh` out of its bash heredoc (about 1050 lines) and
  into `scripts/lib/retired_arch.py`. The script is now a thin wrapper.
* Moved the Go masker verbatim into `scripts/lib/gomask.py`, using the retired-arch superset as the canonical
  version. `scripts/check-write-path-precondition.sh` now imports this shared masker, and its local clone was
  deleted.
* Re-derived the disk-read selection pathspec pin.
* Added a stdlib `unittest` suite in `scripts/lib/tests/` (37 tests).
* Added two steps to the CI `lint` job: a SHA-pinned `actions/setup-python` 3.12, and a blocking
  "Python gate-engine lint and unit tests" step.
* Gave `.gitignore` the `__pycache__/` and `*.pyc` entries.

Gate behaviour is preserved byte for byte; see `docs/closure/029-S-032-F-behaviour-preservation-evidence.md`. No
production `cmd/` or `internal/` Go code changed.

Task commits:

| Task | Commit |
|---|---|
| 032.001-T | `8b44682` |
| 032.002-T | `d6c2c88` |
| 032.003-T | `ad9b3a8` |
| 032.004-T | `2247b3c` |
| 032.005-T | `aa9298c` |
| 032.006-T | `3042ae5` |
| 032.007-T | `ec39da9` |
| review-fix | `a09bb0c` |

## Invariants to preserve

* The gate verdicts and output of `check-retired-architecture.sh` are byte-identical to the pre-extraction heredoc
  engine, in all three modes: repo, `--self-test` and `--self-test-integrity`. The same holds for
  `check-write-path-precondition.sh`.
* There is exactly one canonical Go masker (`scripts/lib/gomask.py`). Neither wrapper carries a local copy, and a
  unit test asserts that both import it.
* Importing `scripts/lib/retired_arch.py` has no side effects: argv, cwd and environment are read only inside
  `main()`. An AST guard enforces this, including for decorators and default arguments.
* The wrappers locate their engine through `BASH_SOURCE`, so they work from any cwd. They also export
  `PYTHONDONTWRITEBYTECODE=1`, so a gate run never leaves `__pycache__` in the working tree.
* The "Python gate-engine lint and unit tests" CI step blocks unconditionally; it has no advisory toggle. It is listed
  in the ci.yml LOCAL DIVERGENCE header, so an `autoharness tune` re-render re-applies it.

## Validator evidence

`.autoharness/workspace-profile.yaml` `runtime_validation.validator_manifest` declares only one runtime surface for
this repository: `cli`, checked by a `go run ./cmd/... --help` smoke run. This shipment does not touch `cmd/intercom`
or `cmd/intercom-ctl`, so that probe does not apply to this diff and no probe evidence is recorded.

The surfaces that did change are the CI gate scripts and the CI workflow. Evidence at reviewed HEAD `09e3d4b`, where
no code has changed since `a09bb0c`:

* **Go**: `gofmt -l .` produced no output. `go vet ./...` and `go build ./...` exited 0, and
  `go test -count=1 ./...` reported every package `ok`.
* **Python unit tests**: `python3 -m unittest discover -s scripts/lib/tests` reported 37 OK.
* **Gate scripts** (all PASS):
  * retired-arch and write-path gates, each in repo, `--self-test` and `--self-test-integrity` modes;
  * gitignore append-only and un-ignore regression gates, with their self-tests;
  * the merge-strategy `--self-test`.
* **Behaviour preservation**: the evidence re-captured at the review-fix HEAD is byte-identical to the recorded
  post-extraction evidence (`3042ae5`).
* **CI**: all 14 checks were green at merge HEAD `09e3d4b`, including `ci gate`. In the `lint` job, both
  "Set up Python (gate engines)" and "Python gate-engine lint and unit tests" succeeded (run 36392723237). That run
  is the real-CI check for 032.006-T.
* **Merge**: `gh pr merge 77 --merge --match-head-commit 09e3d4b…`, without `--admin`, produced merge commit
  `16519360c551ee44c0405fc436bb8b4a8718c4d5`. The repository allows merge commits only (`allow_merge_commit=true`,
  `allow_squash_merge=false`, `allow_rebase_merge=false`), which satisfies P-009.
* **P-018 Copilot-review gate**: `NOT_APPLICABLE`, with no engagement and enforcement `auto`. It was re-run at the last
  mile before the merge.
* **P-016 pipeline-topology**: passed at both lifecycle checks. The pre-merge check ran on the feature branch with a
  single worktree. The pre-closure check ran on `post-merge/029-s-032-f-extract-gate-python-engines` and returned
  `BRANCH_POST_MERGE_CLOSURE_ELIGIBLE`.
* **Merge confirmation**: `gh pr view 77` reports `MERGED` at `2026-09-28T23:46:39Z`, and
  `git merge-base --is-ancestor 1651936 origin/main` exits 0.

Verdict: `PASS_WITH_FOLLOW_UP`. The follow-ups are listed below.

## Pre-deploy audits

None apply. The shipment involves no migration, feature flag, runtime config change, or access change.

## Deployment / rollout path

Merge-only. The extracted engines and the new unit-test step run on every future PR and on every push to `main`.
There is no separate deploy step.

## Post-deploy checks

* On the next PR, confirm the `lint` job still runs the retired-arch and write-path gates and the
  "Python gate-engine lint and unit tests" step, and that each one passes.
* Confirm that no `scripts/lib/__pycache__/` directory appears in the working tree after a local gate run.

## Risky action record

No destructive, irreversible, or elevated-privilege action was taken.

* **Merge**: the operator approved it under P-014 ("PR 77: Merge approved"). It used a merge commit only and was
  bound to HEAD `09e3d4b` through `--match-head-commit`.
* **Covering-feature completion (a1, S4)**: all five conditions held. `backlogit move 032-F --status done` moved
  `032-F` from `active` to `done`; it exited 0, and a re-read showed `done`.
* **Shipment closure**: `shipment-reconcile` `classify-close-path` returned `CASCADE` / `FULLY_COVERED_ROOT` with
  binding `7da9434d…`. The binding was revalidated against live state before any mutation. Safe-close then delegated
  to the Cascade Close Sub-Procedure, which ran `backlogit shipment ship 029-S --sha 1651936…`. Results:
  * `returned_ids=[]`.
  * `archived_ids`, `allowed_ids` and `required_ids` are the same set: {029-S, 032-F, 032.001-T..032.007-T}.
  * `parent_id` was preserved on all seven tasks.
  * `029-S` has `archived_status: shipped`.

  See `.backlogit/reconcile/029-S-safe-close-20260928T234811Z.md`.

## Healthy signals

* The `lint` job's retired-arch and write-path gate verdicts stay green on `main`.
* The Python unit-test step passes on every PR.

## Failure signals

* A gate's verdict or output changes on a PR that does not touch the scanned surfaces. That would be a
  behaviour-preservation regression in the extracted engine.
* The Python unit-test step fails because a module cannot be imported or the masker diverges.
* `__pycache__` or `*.pyc` files appear in a working tree or trip the gitignore or unignore gates.

## Monitoring plan

Watch the standard CI dashboard on `main`, in particular the `lint` job steps named above, over the next several PRs
(030-S onward).

## Rollback trigger

Revert if any of these happens:

* A false-positive or false-negative gate verdict that the extraction caused.
* The Python unit-test step blocking unrelated PRs on an interpreter or environment issue.

## Rollback procedure

Open a standard merge-commit revert PR for merge commit `16519360c551ee44c0405fc436bb8b4a8718c4d5`. That restores the
heredoc engines exactly. There is no data or state migration.

For a narrower fix, the unit-test step can be made `continue-on-error` while the root cause is fixed.

## Validation window

The next 2–3 PRs on `main`: 030-S and the later shipments in the gate-reliability plan.

## Owner

The Ship agent and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (stash; Stage-owned triage)

Each of these is a P-021 DEFERRED SCOPE EXPANSION entry. They are capture-only, Stage-owned, and were captured before
the PR opened, on the threadless path.

* `C312BD4C`: the masker's `struct_tag_re` `\s+` matches newlines, which contradicts a comment. The behaviour is
  pinned by a characterization test and was left unchanged.
* `40C421EF`: pin Python for the gate steps themselves. They currently use the runner's default `python3`, which
  predates this shipment and is not a regression.
* `54EF986C`: replace the stdlib stand-in with a real, hash-pinned Python linter.
* `8387758F`: add write-path struct-tag (D-1) fixtures.

Post-merge closure produced no new follow-ups. Ship created or modified no stash entries during this closure.

## Source artifact cleanup

* `032-F` has no `custom_fields.source_stash_id`. Its description cites "Source: stash 6C24E2E4 item 2", but Step 7
  keys only on the structured field, so it skipped the entry and left it for Stage. Because the stash is multi-item,
  archiving it on one item's shipment would also be premature.
* `032-F` has no `custom_fields.source_deliberation_id`. Skipped. Its references cite
  `docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md`, which is a file
  artifact, not a backlogit deliberation.
* Archived: none. Skipped: `6C24E2E4` (no structured link).

## Knowledge graduation

* No updates were needed to `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/` or `docs/product-specs/`. The
  change is internal CI tooling, and the ci.yml LOCAL DIVERGENCE header was updated in the PR.
* Compound capture: `docs/compound/2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`.
  It covers engine path resolution via `BASH_SOURCE`, bytecode pollution, and side-effect-free import, which review
  found during the extraction.
* `compound-refresh`: not needed. The one related entry,
  `docs/compound/2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md`, covers
  `check-merge-strategy.sh`, which this shipment left unchanged. It is still accurate, so it is kept and
  cross-referenced from the new entry.

## Compaction status (P-020)

`pending`. Ship Step 6 item 8 finalizes this field after it runs `compact-context`.

## Releasability evidence

**READY.**

* Local review readiness: `READY_WITH_FOLLOWUPS` (P0=0, P1=0) at reviewed HEAD `09e3d4b`, unchanged through the
  merge.
* All 14 CI checks were green at the merge HEAD, and the merge used a merge commit (P-009).
* P-018 `NOT_APPLICABLE`, P-016 passed, and the P-014 operator approval is recorded.
* No runtime service surface changed. The gate scripts, their self-tests, the behaviour-preservation diff, and the
  new unit-test step are all green.
* The residual follow-ups are advisory and do not block release. They are recorded above for Stage.
