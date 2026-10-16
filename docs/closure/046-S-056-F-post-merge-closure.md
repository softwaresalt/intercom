---
title: "046-S / 056-F post-merge operational closure"
description: "Post-merge closure for hermetic and containable start-script integration tests (046-S, covering feature 056-F): feature PR 141 merged, the operator-approved cascade close (postcondition pass), the archive relocation, and the completed P-020 compaction."
status: ready
tags:
  - closure
  - post-merge
  - tests
  - start-script
  - shipment-close
  - 046-S
  - 056-F
date: 2026-10-16
mode: post-merge
shipment: 046-S
feature: 056-F
pr: 141
feature_pr: 141
merge_commit: 75e8755d8b8f7f98a8b952bb529352fa6590550a
merge_commit_sha: 75e8755d8b8f7f98a8b952bb529352fa6590550a
feature_merge_commit_sha: 75e8755d8b8f7f98a8b952bb529352fa6590550a
releasability: READY
conditions:
  - id: cascade-close-operator-approval
    description: "Explicit per-closure operator approval for the mutating autoharness shipment cascade-close run. The operator approved it for 046-S / 056-F at SHA 75e8755d; Ship ran it (space-separated --message/--author form) and it exited 0."
    satisfied: true
    evidence: docs/closure/evidence/046-S-056-F-close-evidence.json
  - id: archive-relocation-and-postcondition
    description: "Archive relocation of 046-S, 056-F and tasks 056.001-T..056.003-T, committed on the closure branch after the post-mode PROCEED. The evidence postcondition verdict is pass, with no failures and no unexpected archived IDs."
    satisfied: true
    evidence: .backlogit/reconcile/046-S-post-20261016T063000Z.md
  - id: p020-compaction-completion
    description: "P-020 compact-context ran with target all and succeeded; originals archived under docs/archive/memory, none deleted."
    satisfied: true
    evidence: docs/memory/compacted/2026-10-16-046-s-compact-context-report.md
head_sha_reviewed: a6bb6fc27fa97135932a28dad14760f28a1ac335
code_sha_reviewed: 1dedea6
base_sha: 84f4915
closure_base_sha: 75e8755d8b8f7f98a8b952bb529352fa6590550a
closure_branch: post-merge/056-hermetic-start-script-tests
close_path: cascade
close_evidence: docs/closure/evidence/046-S-056-F-close-evidence.json
closure_status: READY
compaction_status: done
---

# 046-S / 056-F post-merge operational closure

## Summary

Shipment 046-S delivered covering feature 056-F: the two start-script launcher integration tests
(`TestInvokeStartScriptMainPropagatesNonZeroExitCode`, `TestLocationRestoredEvenOnNonZeroCopilotExit`)
now run a scratch byte copy of `start.ps1` under a PATH with no backlogit, engram or graphtor-docs
directories, and are bounded by `cmd.WaitDelay` so a timed-out launcher cannot block on a pipe held by an
orphaned descendant. Test-only; `start.ps1` is unchanged. On the Windows dev host with sidecars installed the
tests dropped from 25-35 s (and unbounded hangs) to about 2.5 s and 3 s.

Feature PR 141 merged at `75e8755` with a merge commit. The mutating cascade close was run by Ship with explicit
per-closure operator approval and reported `postcondition_verdict: pass`, five archived IDs (046-S, 056-F, three
tasks), and `shipment_status: shipped`.

## Merge and gate record

* Feature PR 141, head `a6bb6fc27fa97135932a28dad14760f28a1ac335`, merged head-pinned (`--match-head-commit`)
  with a merge commit and no admin flag. Merge commit `75e8755d8b8f7f98a8b952bb529352fa6590550a`, confirmed as an
  ancestor of `origin/main`.
* P-009: repository allows merge commits only (squash and rebase disallowed).
* Local review: formal review of code at `1dedea6`, READY_WITH_FOLLOWUPS (no P0/P1/P2; five advisory P3s accepted as
  residual risk). The post-review delta (backlog-state commit `80a7433`, and the one-line `//nolint:unparam` commit
  `a6bb6fc`) was checked directly by Ship and was not formally re-reviewed.
* CI on `a6bb6fc`: lint, test, security, ci gate and the advisory Windows test passed. The first run failed lint
  (`unparam` on `writeStubExecutable`), fixed in `a6bb6fc`.
* P-018 copilot-review gate: NOT_APPLICABLE (pass); never forced.

## Close path

* Lifecycle topology gate (a0): exit 0.
* Covering-feature completion gate (a1): all five conditions held; 056-F `active -> done`, re-read as `done`.
* Pre-archive reconciliation (pre mode, expected_status done): PROCEED. Report
  `.backlogit/reconcile/046-S-pre-2026-10-16T06-40-00Z-done.md`. No per-file lock was taken (lock scripts are not
  installed; single-agent workflow); the cascade-close command took its own pair lock.
* Classify-only: exit 0, CASCADE, engine VERIFIED (backlogit 1.11.0), `selected_close_path` cascade.
* Mutating cascade close: exit 0, started 2026-10-16T06:26:20Z, finished 06:27:37Z; postcondition pass.
* Post-mode: PROCEED, report `.backlogit/reconcile/046-S-post-20261016T063000Z.md`; no deletions under
  `.backlogit/archive/`.
* Linked-deliberation disposition: no dispositions, no unresolved references (DISPOSITION_COMPLETE).

## Deviations and reconciliation record

1. **`--force` misuse.** Ship ran the pre_claim topology gate once with `--force` (an operator-only flag) after a
   `BRANCH_MISMATCH` caused by a non-conforming branch name; the shipment claim followed. Audit entry:
   `.autoharness/gates/pipeline-topology-force-audit.log`. The branch was renamed to a conforming name and the unforced
   post_claim gate passed with 046-S as the sole active shipment. This is a process deviation; no gate was forced
   afterwards.
2. 056.003-T had no separate red run (the plan requires red evidence for U1 and U2 only).
3. G1 derives the scratch script directory from the dot-sourced function's script file, because a dot-sourcing
   wrapper cannot see the sourced file's `$PSScriptRoot`; the assertion is still `os.SameFile`.
4. `go test ./tools/gatecheck` has 21 wrapper-test failures on the Windows dev host (package untouched; WSL bash first
   on PATH). Linux CI and the Windows advisory lane passed.
5. The `.autoharness/config.yaml` operator edit was carried uncommitted until PR #140 merged it. It was then restored from
   main after a `git diff --quiet origin/main` identity check and a byte-compared backup.

## Risky action record

* a1 `active -> done` on 056-F: backlog state only, reversible by a backlog move.
* Mutating cascade close: operator-approved, recorded in the evidence record. Blast radius was limited to 046-S, 056-F
  and its three tasks.
* No ProposedAction entries affected rollout or rollback. Test-only change.

## Source artifact cleanup

* 056-F and its tasks declare no `source_stash_id` and no `source_deliberation_id`. No source stash entry was archived by
  Ship. The stash labels `C5E50EFD` and `26EDD633` on the 056 items name the addressed entries; archiving them is a
  discretionary stash operation that Ship does not perform, so they remain for Stage.
* `linked_deliberation_disposition` is empty in the pre_close record, and the post_close record reports no drift.

## Stash follow-ups (visible to Stage)

None created. The advisory P3 review notes (G3 10 s bound margin, early-fail cleanup of the 20 s grandchild, pid reuse in
cleanup, ambient `COPILOT_HOME`/`ENGRAM_DATA_DIR` in the location test, unbounded runner in the remaining snippet tests) are
accepted as residual risk and recorded in the PR body.

## Operational closure fields

* **Invariants to preserve:** the two launcher tests run a scratch copy under a sidecar-free PATH and are bounded by
  `startScriptWaitDelay`; G1, G2 and G3 must stay non-vacuous.
* **Validator evidence:** runtime verification not applicable (test-only change).
* **Pre-deploy audits:** none. Merge-only path.
* **Deployment or rollout path:** merge-only, already merged at `75e8755`.
* **Post-deploy checks:** CI on `main` (ci gate, test, lint, security) green; the integration tests complete in seconds on
  Linux.
* **Healthy signals:** main CI green; no start-script test timeouts.
* **Failure signals:** a start-script integration test timing out, or G1/G2/G3 failing on main.
* **Monitoring plan:** the CI workflow on `main` and the advisory Windows lane.
* **Rollback trigger:** a red integration test on `main` attributable to `75e8755`.
* **Rollback procedure:** revert merge commit `75e8755` with `git revert -m 1 75e8755` on a new branch through a reviewed PR;
  never push directly to main.
* **Validation window:** until the first green CI run on `main` after the closure merges.
* **Owner:** the operator. Ship reports and does not own the window.
* **Releasability evidence:** `READY`. Operator approval, archive relocation and P-020 compaction are satisfied.

## Compaction (P-020)

`compaction_status: done`. Report: `docs/memory/compacted/2026-10-16-046-s-compact-context-report.md`. The 046-S Ship session memory
and the 2026-10-12 Stage memory were consolidated into
`docs/memory/compacted/2026-10-16-046-s-056-f-stage-and-ship-compacted.md`; originals moved to `docs/archive/memory/`, nothing deleted.
