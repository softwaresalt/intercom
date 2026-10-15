---
title: "045-S / 055-F post-merge operational closure"
description: "Post-merge closure for gatecheck root anchoring and MSYS environment stripping (045-S, covering feature 055-F): PR 135 merged at 526fccf, the classify-only close path (CASCADE, engine VERIFIED), and the mutating cascade close halted for per-closure operator approval."
status: ready_with_conditions
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
closure_pr: pending
merge_commit: 526fccf46669891c4b9a191dc97449cd30e797a2
merge_commit_sha: 526fccf46669891c4b9a191dc97449cd30e797a2
feature_merge_commit_sha: 526fccf46669891c4b9a191dc97449cd30e797a2
releasability: READY_WITH_CONDITIONS
conditions:
  - id: cascade-close-operator-approval
    description: "Explicit per-closure operator approval for the mutating autoharness shipment cascade-close run (classify-only selected cascade). Dark-mode authority does not satisfy it."
    satisfied: false
    evidence: docs/closure/evidence/045-S-055-F-close-evidence.json
  - id: post-approval-archive-and-closure-gate
    description: "Backlog archive commit after the approved cascade close, then the closure-evidence gate exits 0 before completion is declared."
    satisfied: false
    evidence: docs/closure/evidence/045-S-055-F-close-evidence.json
head_sha_reviewed: df7d8fd6baf46af3b698e5d50da49311b6c66a09
code_sha_reviewed: df7d8fd
base_sha: f038032
closure_branch: post-merge/055-f-closure-045-s
close_path: cascade
close_evidence: docs/closure/evidence/045-S-055-F-close-evidence.json
closure_status: READY_WITH_CONDITIONS
compaction_status: degraded
---

# 045-S / 055-F post-merge operational closure

## Summary

Shipment 045-S delivered covering feature 055-F: the gatecheck write-path and retired-architecture
wrappers derive their root only from their own checkout through `gatecheck_repo_root`, fail closed
when the checkout is not a repository top, and the Go git children drop Git Bash and Cygwin
translation variables. Feature PR 135 merged at `526fccf` after the Copilot review gate passed
for head `df7d8fd`.

The closure is not complete. The classify-only close path selected `cascade`, and the mutating
`autoharness shipment cascade-close` run needs per-closure operator approval, which dark-mode
authority does not satisfy. Shipment 045-S therefore remains `active`, and the successor shipment
046-S stays blocked by P-001 and P-020 until the close is approved and completed.

## Merge and gate record

* Feature PR 135, head `df7d8fd6baf46af3b698e5d50da49311b6c66a09`, merged head-pinned with a merge
  commit (`gh pr merge 135 --merge --match-head-commit df7d8fd...`, no admin flag). Merge commit
  `526fccf46669891c4b9a191dc97449cd30e797a2`, confirmed as an ancestor of `origin/main`.
* P-009: repository allows merge commits only (`allow_merge_commit` true, squash and rebase false).
* P-016: `pipeline-topology` gates passed at pre_claim, post_claim, and lifecycle (both before the
  feature PR and before closure).
* CI on head `df7d8fd`: all required checks passed (ci gate, cross-compile on four targets, lint,
  test, security, merge-strategy, topology, gitignore regression; Windows test advisory passed).
* P-018 copilot-review gate: `NOT_APPLICABLE` before the review was requested. The review was
  requested through the approved GraphQL route after the operator's resumption instruction.
  Copilot reviewed `df7d8fd` with "Approval recommended" and zero open findings. No inline threads
  and no body findings required disposition. The gate returned `SATISFIED` and was re-run
  unconditionally before the merge.
* Local review (P-014): three rounds on the feature diff. Final readiness `READY_WITH_FOLLOWUPS`,
  P0 0, P1 0. Follow-ups are recorded below.

## Scope of the merged change

* scripts/lib/gatecheck-run.sh: `gatecheck_repo_root` (scrubbed, checkout-anchored, fail-closed root).
* scripts/check-write-path-precondition.sh and scripts/check-retired-architecture.sh: `ROOT` bound
  through `gatecheck_repo_root`, caller precondition kept.
* tools/gatecheck/internal/writepath/writepath.go and retiredarch/select.go and pin.go:
  MSYS, CYGWIN and MSYS2_* stripped from the git child environment (one atomic commit for the
  retiredarch body and its canonical pin).
* Tests: shared wrong-tree table for both wrappers, ROOT shape pins, and env-table rows.

Twelve commits from `9e42d16` to `df7d8fd` on the feature branch. Git history is the record.

## Close path

* Lifecycle topology gate (a0): exit 0.
* Covering-feature completion gate (a1): all five conditions held. Every manifest task was `done`
  and archived, no live descendant remained at any depth, the descendant set equalled the
  manifest, containment held, and 055-F was `active`. Single permitted transition `active -> done`
  on 055-F, re-read as `done`. No force flags.
* Pre-archive reconciliation (pre mode, expected_status done): lock acquired on
  `.backlogit/queue/045-S.md`, record `active` classified `record-consistent`, 055-F matched in
  queue, seven tasks `pre-archived`, no orphans. Verdict PROCEED. Lock released after the
  classification stopped at the approval boundary.
* Classify-only close (`autoharness shipment cascade-close --classify-only --shipment 045-S
  --feature 055-F --sha 526fccf46669891c4b9a191dc97449cd30e797a2 --json`): exit 0, classifier
  verdict `CASCADE`, engine verdict `VERIFIED` (backlogit 1.11.0 on the verified 1.11 minor line),
  `selected_close_path` `cascade`, `mutation_possible` `no`. Evidence written to
  `docs/closure/evidence/045-S-055-F-close-evidence.json` (phase `pre_close`, not hand-edited).
* Mutating cascade close: NOT RUN. It needs explicit per-closure operator approval, which dark-mode
  authority does not satisfy (Constitution Principle VII). Halted here.
* Post-merge backlog archive commit (`chore: archive 045-S backlog artifacts`): NOT RUN. It follows
  the mutating close and is therefore blocked with it.

## Deviations and reconciliation record

1. Feature commit ded479b carries a test-support harness change (shim installed once per table).
   No assertion changed. The red state was re-verified against the pre-change production files
   with the restructured harness.
2. Commit footers end with the Co-authored-by trailer requested by the operator. The literal
   `- Generated by Copilot` line appears before the trailer.
3. 055.003-T acceptance criterion 1 expected the retired-architecture subdirectory row to be red.
   It is green on arrival, because `git -C "$GATECHECK_SRC"` anchors the worktree to the checkout.
   The row stays as a regression guard. Acceptance criteria are planning fields that Ship does not
   edit, so the deviation is recorded here.
4. Ship did not formally invoke the safety-modes skill for U10, U12, U14 and U15 (careful mode).
   The blast radius was enumerated by hand and each pair was gated before commit.
5. Local review round 1 was blocked by an adversarial P1 (the `-C` anchor was not required by any
   test). Fixed by commit 841d1b4 and mutation-verified. Round 2 found a P1 (readonly refusal
   untested) and a P2 (diagnostic matched trace lines). Fixed by df7d8fd and mutation-verified.
6. Host note: on this Windows host `bash` on PATH is WSL. The serialized suite ran with Git Bash
   first on PATH. Shim-based wrapper tests can flake under package-parallel load on this host
   (shared Git Bash `/tmp` mount). CI on Linux is unaffected, and CI on head `df7d8fd` passed.
7. Copilot request route: the MCP tool was not available in the first session. The Orchestrator
   requested review through the approved GraphQL route. Copilot re-review after any future push must
   use the same route.

## Risky action record

* a1 covering-feature transition `active -> done` on 055-F: low risk, backlog state only, reversible
  by a backlog move. Recorded here because it is the only backlog mutation made before approval.
* No ProposedAction entries affected rollout or rollback. The feature changes CI gate scripts only.

## Source artifact cleanup

* 055-F and its seven tasks declare no `source_stash_id` and no `source_deliberation_id`. No source
  stash entry was archived.
* The stash labels `stash-DBE25DF5` and `stash-2613D248` on the 055 items record the stash entries
  that the shipped scope addresses. They are not declared source fields, and archiving them would
  be a discretionary stash archive, which Ship does not perform. DBE25DF5 is addressed for the
  write-path and retired-architecture wrappers. Its remaining scope (the unignore and merge-strategy
  wrappers) stays with stash 1B85D0F4.
* Deliberation disposition: `linked_deliberation_disposition` is empty in the pre_close record.

## Stash follow-ups (visible to Stage)

Captured during the review remediation as `DEFERRED SCOPE EXPANSION` entries (P-021 C2,
capture-only):

* 51117F2D: reference audit and compaction of the eligible memory groups (P-020 follow-up).
* 83C1369E: Linux-discriminating recorder for the case-insensitive scrub and config pins.
* 815C2B28: export-pin refusal row for an unexported readonly `GIT_CONFIG_GLOBAL`.
* C37F75D5: shell-side MSYS scope decision (the shell scrub keeps MSYS, CYGWIN and MSYS2_*).
* BEFE6DF2: capture `compgen -e` before iterating, so a failed listing fails closed.
* EF11062A: move the shared wrong-tree harness to a harness file.
* 63BCA7C6: write-path header paragraph and PS4/LC_ALL hygiene in the xtrace filter.

Existing entries cited, not duplicated: 5A8EC1BC (gitRunnerEnv copy and parity), 585FCD6D (Unicode
and Windows case folding, safe.directory), 1B85D0F4 (unignore and merge-strategy wrapper roots),
4372BAD4 (relative git in writepath), 49982A4E (engine-side top-level check).

## Operational closure fields

* **Invariants to preserve:** gatecheck roots derive only from the checkout; a non-top-level
  checkout, a readonly `GIT_*` variable, or a failed config pin refuses with exit 2 or 128 and never
  reaches the engine; the MSYS filter stays identical in both gitRunnerEnv copies.
* **Validator evidence:** runtime verification not applicable. The change touches CI gate scripts
  and gatecheck tools only, with no runtime surface.
* **Pre-deploy audits:** none. Merge-only path, no deployment.
* **Deployment or rollout path:** merge-only, already merged at `526fccf`.
* **Post-deploy checks:** the CI run on `main` for `526fccf` (ci gate, test, lint, security) must be
  green. The next gatecheck run on `main` must pass the wrong-tree and pin tests.
* **Healthy signals:** main CI green; gatecheck tests green on Linux (the blocking lane).
* **Failure signals:** a red gatecheck test on main, or a wrapper that resolves a root other than
  its checkout.
* **Monitoring plan:** the CI workflow on `main` and the advisory Windows test lane.
* **Rollback trigger:** a red gatecheck test on `main` attributable to 526fccf.
* **Rollback procedure:** revert merge commit `526fccf` with `git revert -m 1 526fccf` on a new
  branch, through a reviewed PR. Never push directly to main.
* **Validation window:** until the first green CI run on `main` after the closure merges.
* **Owner:** the operator. The Ship agent reports and does not own the window.
* **Releasability evidence:** `READY_WITH_CONDITIONS`. Condition: explicit per-closure operator
  approval for the mutating cascade close, then the archive commit and the closure-evidence gate
  pass. The conditions are unsatisfied, so the closure is not declared complete and the successor
  shipment stays blocked.

## Compaction (P-020)

`compaction_status: degraded`. compact-context was invoked for target `all` and ran Phase 1 and
Phase 2 (assessment and candidate selection). Phase 3 compacted no group, because the eligible memory
groups could not be cleared against live references in this run (blocked item 026-F, and an
unverified reference audit for the rest). The report is
`docs/memory/compacted/2026-10-15-045-s-compact-context-report.md`. Degraded is non-blocking under
P-020. The Orchestrator treats it as satisfied, and the reference audit and compaction are
captured as stash entry 51117F2D.
