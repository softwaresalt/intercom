---
title: "045-S / 055-F post-merge operational closure"
description: "Post-merge closure for gatecheck root anchoring and MSYS environment stripping (045-S, covering feature 055-F): feature PR 135 and closure PR 136 merged, the operator-approved cascade close (postcondition pass), the archive relocation, and the P-020 compaction that remains pending."
status: ready
tags:
  - closure
  - post-merge
  - dev-tooling
  - shipment-close
  - 045-S
  - 055-F
date: 2026-10-15
mode: post-merge
shipment: 045-S
feature: 055-F
pr: 135
feature_pr: 135
closure_pr: 136
merge_commit: 526fccf46669891c4b9a191dc97449cd30e797a2
merge_commit_sha: 526fccf46669891c4b9a191dc97449cd30e797a2
feature_merge_commit_sha: 526fccf46669891c4b9a191dc97449cd30e797a2
releasability: READY
conditions:
  - id: cascade-close-operator-approval
    description: "Explicit per-closure operator approval for the mutating autoharness shipment cascade-close run. The operator ran the command and approved it. Ship did not re-run it."
    satisfied: true
    evidence: docs/closure/evidence/045-S-055-F-close-evidence.json
  - id: archive-relocation-and-postcondition
    description: "Archive relocation of 045-S and its nine manifest items, committed on the archive branch after the post-mode PROCEED. The evidence postcondition verdict is pass, with no failures and no unexpected archived IDs."
    satisfied: true
    evidence: .backlogit/reconcile/045-S-post-20261015T173631Z.md
  - id: p020-compaction-completion
    description: "P-020 compaction of the remaining candidates recorded in the compaction report completes, and compaction_status is then set to done. Until then the closure is not complete, and routing for the next shipment stays blocked."
    satisfied: true
    evidence: docs/memory/compacted/2026-10-15-045-s-compact-context-report.md
head_sha_reviewed: df7d8fd6baf46af3b698e5d50da49311b6c66a09
code_sha_reviewed: df7d8fd
base_sha: f038032
closure_base_sha: 526fccf46669891c4b9a191dc97449cd30e797a2
closure_branch: post-merge/055-f-closure-045-s
archive_branch: post-merge/045-s-archive-relocation
close_path: cascade
close_evidence: docs/closure/evidence/045-S-055-F-close-evidence.json
closure_status: READY
compaction_status: done
---

# 045-S / 055-F post-merge operational closure

## Summary

Shipment 045-S delivered covering feature 055-F: the gatecheck write-path and retired-architecture
wrappers derive their root only from their own checkout through `gatecheck_repo_root`, fail closed
when the checkout is not a repository top, and the Go git children drop Git Bash and Cygwin
translation variables. Feature PR 135 merged at `526fccf`. Closure PR 136 merged at `1b06f4c`
with the closure record.

The mutating cascade close was run by the operator with explicit per-closure approval. Ship did not
re-run it. The evidence record reports `postcondition_verdict: pass`, no failures, no unexpected
archived IDs, and nine archived IDs (045-S, 055-F, and seven tasks). Shipment 045-S is now
`archived` with status `shipped`. This branch carries the archive relocation and the post-mode
reconciliation record.

P-020 compaction completed on `post-merge/045-s-compaction` (`compaction_status: done`), so every
closure condition is satisfied and the closure-evidence gate can pass. See the Compaction section.

## Merge and gate record

* Feature PR 135, head `df7d8fd6baf46af3b698e5d50da49311b6c66a09`, merged head-pinned with a merge
  commit (no admin flag). Merge commit `526fccf46669891c4b9a191dc97449cd30e797a2`, confirmed as an
  ancestor of `origin/main`. Parents: `f038032` and `df7d8fd`.
* P-009: repository allows merge commits only (`allow_merge_commit` true, squash and rebase false).
* P-016: `pipeline-topology` passed at pre_claim, post_claim, and lifecycle.
* CI on head `df7d8fd`: required checks passed.
* P-018 copilot-review gate on PR 135: `SATISFIED` on `df7d8fd` after Copilot approved the head with
  zero findings.
* Closure PR 136, head `5f01138`, merged head-pinned with a merge commit (no admin flag). Merge
  commit `1b06f4c6030eec30bd7105196ebabb220e98ef41`, confirmed in `origin/main` history.
* P-018 on PR 136: one Copilot finding (compaction status), fixed in `5f01138`, replied and resolved.
  The re-review on `5f01138` had zero open findings, and the gate returned `SATISFIED`.

## Scope of the merged change

* scripts/lib/gatecheck-run.sh: `gatecheck_repo_root` (scrubbed, checkout-anchored, fail-closed root,
  refusal on a readonly `GIT_*` variable and on an unexportable config pin).
* scripts/check-write-path-precondition.sh and scripts/check-retired-architecture.sh: `ROOT` bound
  through `gatecheck_repo_root`, caller precondition kept.
* tools/gatecheck/internal/writepath/writepath.go and retiredarch/select.go and pin.go:
  MSYS, CYGWIN and MSYS2_* stripped from the git child environment (one atomic commit for the
  retiredarch body and its canonical pin).
* Tests: shared wrong-tree table for both wrappers, ROOT shape pins, and env-table rows.

Twelve commits from `9e42d16` to `df7d8fd` on the feature branch. Git history is the record.

## Close path

* Lifecycle topology gate (a0): exit 0.
* Covering-feature completion gate (a1): all five conditions held. Every manifest task was `done` and
  archived, no live descendant remained at any depth, the descendant set equalled the manifest,
  containment held, and 055-F was `active`. Single transition `active -> done`, re-read as `done`.
  Backlogit archived 055-F as a routing effect of the done status.
* Pre-archive reconciliation (pre mode, expected_status done): record `active` classified
  `record-consistent`, 055-F pre-archived with status done, the seven tasks pre-archived, no orphans.
  Verdict PROCEED. The pre-mode report file was not persisted to `.backlogit/reconcile/` and is not reconstructed here. (`autoharness shipment cascade-close --classify-only`): exit 0, verdict CASCADE,
  engine VERIFIED (backlogit 1.11.0), `selected_close_path` cascade, `mutation_possible` no (classify
  command output). Evidence written to `docs/closure/evidence/045-S-055-F-close-evidence.json`.
* Mutating cascade close: run by the operator with explicit per-closure approval. The evidence
  record's `invocation` section shows the backlogit shipment ship call, exit 0, mutation state
  completed, started 2026-10-15T17:34:16Z, finished 2026-10-15T17:36:31Z. Its `shipment_status` is
  `shipped`, with `commit_sha` 526fccf and the nine archived IDs. The record's `phase` is
  `post_close`.
* Post-mode reconciliation: `.backlogit/reconcile/045-S-post-20261015T173631Z.md`. Recommendation
  PROCEED. Every archive file is present, the P-007 guard finds no deletions under `.backlogit/archive/`,
  and no shipment lock is held.
* Archive relocation: committed on `post-merge/045-s-archive-relocation`. The queue deletions and the
  archive additions are the operator run's changes. Ship stages only those paths, plus the stash
  captures and this record. It does not stage the operator's `.autoharness/config.yaml` routing edit.

## Deviations and reconciliation record

1. Feature commit ded479b carries a test-support harness change (shim installed once per table). No
   assertion changed. The red state was re-verified against the pre-change production files with the
   restructured harness.
2. Commit footers end with the Co-authored-by trailer requested by the operator. The literal
   `- Generated by Copilot` line appears before the trailer.
3. 055.003-T acceptance criterion 1 expected the retired-architecture subdirectory row to be red. It
   is green on arrival, because `git -C "$GATECHECK_SRC"` anchors the worktree to the checkout. The
   row stays as a regression guard. Acceptance criteria are planning fields that Ship does not edit,
   so the deviation is recorded here.
4. Ship did not formally invoke the safety-modes skill for U10, U12, U14 and U15 (careful mode). The
   blast radius was enumerated by hand, and each pair was gated before commit.
5. The mutating cascade close was run by the operator, not by Ship. The operator's approval is
   recorded in the conditions above.
6. P-020 compaction was first run as a partial pass and was completed afterwards. See the compaction section.
7. Host note: on this Windows host `bash` on PATH is WSL. The serialized suite ran with Git Bash
   first on PATH. Shim-based wrapper tests can flake under package-parallel load on this host. CI on
   Linux is unaffected, and CI on head `df7d8fd` passed.

## Risky action record

* a1 covering-feature transition `active -> done` on 055-F: backlog state only. Its routing effect
  archived 055-F. Reversible by a backlog move.
* Mutating cascade close (operator-run, explicit per-closure approval): recorded in the evidence
  record.
* No ProposedAction entries affected rollout or rollback. The feature changes CI gate scripts only.

## Source artifact cleanup

* 055-F and its seven tasks declare no `source_stash_id` and no `source_deliberation_id`. No source
  stash entry was archived by Ship.
* The stash labels `stash-DBE25DF5` and `stash-2613D248` on the 055 items record the stash entries the
  shipped scope addresses. They are not declared source fields, and archiving them would be a
  discretionary stash archive, which Ship does not perform. DBE25DF5 is addressed for the write-path
  and retired-architecture wrappers. Its remaining scope (the unignore and merge-strategy wrappers)
  stays with stash 1B85D0F4.
* `linked_deliberation_disposition` is empty in the pre_close record, and the post_close record reports
  no drift.

## Stash follow-ups (visible to Stage)

Captured during the review remediation as `DEFERRED SCOPE EXPANSION` entries (P-021 C2, capture-only):

* 83C1369E: Linux-discriminating recorder for the case-insensitive scrub and config pins.
* 815C2B28: export-pin refusal row for an unexported readonly `GIT_CONFIG_GLOBAL`.
* C37F75D5: shell-side MSYS scope decision (the shell scrub keeps MSYS, CYGWIN and MSYS2_*).
* BEFE6DF2: capture `compgen -e` before iterating, so a failed listing fails closed.
* EF11062A: move the shared wrong-tree harness to a harness file.
* 63BCA7C6: write-path header paragraph and PS4/LC_ALL hygiene in the xtrace filter.

Closure follow-ups (P-020):

* 51117F2D: reference audit and compaction of the eligible memory groups. Work done, see Compaction.
* 3D161831: compaction of the completed-work memory candidates. Work done, see Compaction.

Both entries stay in the stash for Stage to archive. Ship is capture-only for stash.

Accepted residuals named in the 055-F definition of done: 1B85D0F4 (other wrappers), 49982A4E
(engine `--root`), and 36BE1F61 (main.go sinks). Existing entries are cited, not duplicated:
5A8EC1BC (gitRunnerEnv copy and parity), 585FCD6D (Unicode and Windows case folding, safe.directory),
4372BAD4 (relative git in writepath).

## Operational closure fields

* **Invariants to preserve:** gatecheck roots derive only from the checkout. A non-top-level checkout,
  a readonly `GIT_*` variable, or a failed config pin refuses with exit 2 or 128, and never reaches the
  engine. The MSYS filter stays identical in both gitRunnerEnv copies.
* **Validator evidence:** runtime verification not applicable. The change touches CI gate scripts and
  gatecheck tools only.
* **Pre-deploy audits:** none. Merge-only path.
* **Deployment or rollout path:** merge-only, already merged at `526fccf`.
* **Post-deploy checks:** the CI run on `main` for `1b06f4c` (ci gate, test, lint, security) must be green.
  The next gatecheck run on `main` must pass the wrong-tree and pin tests.
* **Healthy signals:** main CI green, and gatecheck tests green on Linux (the blocking lane).
* **Failure signals:** a red gatecheck test on main, or a wrapper that resolves a root other than its
  checkout.
* **Monitoring plan:** the CI workflow on `main`, and the advisory Windows test lane.
* **Rollback trigger:** a red gatecheck test on `main` attributable to `526fccf`.
* **Rollback procedure:** revert merge commit `526fccf` with `git revert -m 1 526fccf` on a new branch,
  through a reviewed PR. Never push directly to main.
* **Validation window:** until the first green CI run on `main` after the closure merges.
* **Owner:** the operator. The Ship agent reports and does not own the window.
* **Releasability evidence:** `READY`. The operator approval, the archive relocation, and the P-020
  compaction are all satisfied.

## Compaction (P-020)

`compaction_status: done`. The first run (Phase 1 and 2 only) is in
`docs/memory/compacted/2026-10-15-045-s-compact-context-report.md`. The completion pass is recorded in
that report's "Completion pass" section and covers stash follow-ups 51117F2D and 3D161831.

Compacted (Phase 3), originals moved with `git mv` to `docs/archive/memory/`, nothing deleted:

* `2026-10-12/043-s-ship-session-checkpoint-4.md` and `2026-10-12/043-s-ship-closure-session-memory.md`
  into `docs/memory/compacted/2026-10-12-043-s-053-f-ship-checkpoint-4-and-closure-compacted.md`.
* `2026-10-13/044-s-054-f-closure-pr-134-final-session-memory.md` into
  `docs/memory/compacted/2026-10-13-044-s-054-f-pr-134-review-loop-compacted.md`.

Retained after the reference audit (no Phase 3): the 2026-09-18 group (three files, held by blocked 025-F,
026-F, 037-F and 043-F, live stash DB12DA37, and a live plan reference), 2026-10-01 (14 days, not older than
the threshold, live stash entries, blocked 038-F), 2026-10-07, 2026-10-11 and the 2026-10-12 stage memory
(inside the threshold, live stash entries, queued 046-S and 056-F). The audit table is in the report.
