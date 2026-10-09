---
title: "Stage session: Gatecheck Batch A correctness (shipment 041-S queued)"
date: 2026-10-08
agent: stage
status: complete
branch: chore/stage-gatecheck-batch-a-correctness
base: "main@135ca59"
deliberation: docs/decisions/2026-10-08-intercom-go-gatecheck-batch-a-correctness-deliberation.md
plan: docs/plans/2026-10-08-intercom-go-gatecheck-batch-a-correctness-plan.md
feature_id: 051-F
shipment_id: 041-S
---

# Stage session memory: Gatecheck Batch A correctness

## Status

**Complete. Shipment 041-S is queued** and covers feature 051-F plus six tasks. The next
step belongs to Ship: claim 041-S after the operator has pushed this branch and merged the
staging PR.

## Session inputs

* **Orchestrator `stage next` on Batch A.** Mode is sequential and non-dark. The route was
  claude-opus-5.5 / anthropic / high.
* **Operator instructions (2026-10-08T14:12-07:00):**
  * Carry forward the `.gitignore` line `.github/copilot/` and commit it on the Stage branch.
  * Carry forward `.backlogit/stash.jsonl` (stat-dirty only at session start).
  * Run `stage next` on Batch A.
* **Batch A:** 4537B2F6, 0D643BE8 and D44D8BDF.
* **Explicitly excluded:** DC921AF6, 3750C37C, 0ECC1895, writepath-harness and Root.Resolve
  entries, and every other stash entry.

## Steps

| Step | Outcome |
|---|---|
| 0.0 Tool gate | `ALL_TOOLS_OK`: backlogit v1.11.0 CLI. agent-intercom is not installed, so there were no broadcasts. |
| 0.1 Index sync | `INDEX_SYNC_OK` |
| Recovery | 67 checkpoints, all resolved. Zero active, no anomalies. Normal startup. |
| 1 Triage | 0D643BE8 carries the DEFERRED SCOPE EXPANSION marker, so it was forced to deliberate (P-021 C6). The duplicate scan was clean. Late-identifier reconciliation recovered PR #112 from `docs/closure/040-S-050-F-post-merge-closure.md`; no review thread was found, so the thread N/A stands. 4537B2F6 and D44D8BDF are task-shaped. |
| 1.5 Grouping | The operator pre-selected Batch A. Covering feature: "Gatecheck Batch A correctness". |
| 1.8 Learnings | Pin canonical-text maintenance compound; staging-premise verification compound. |
| 1.9 Branch gate | PASS. `chore/stage-gatecheck-batch-a-correctness` created from `origin/main` 135ca59. |
| 2 Deliberation | D-BA-1/1a (`-z` plus inline NUL parse, atomic pin refreeze), D-BA-2 (empty-selection guard in `runRepoScan`), D-BA-3 (narrowed containment: writepath and unignore), D-BA-4, D-BA-5. |
| 3 Plan + hardening | Units U1-U6 with `Requires plan hardening: yes`. A `## Plan Hardening` section was written. |
| 4 Plan review | **PASS** on attempt 1. Four parallel personas (Go, Scope, Security Lens, Constitution) each returned ADVISORY, and every finding was resolved in the plan text. |
| 5 Harvest | Feature 051-F and tasks 051.001-T to 051.006-T, with sections, sizes and dependency edges. |
| 5.5 Shipment | 041-S, `queued`, 7 items, feature first. |
| 5.6 Archive | 4537B2F6, 0D643BE8 and D44D8BDF archived (non-destructive `stash archive`). |

## Premise verification (against main@135ca59)

* **4537B2F6: CONFIRMED.**
  * `select.go:40` builds `ls-files --` without `-z`.
  * `selectRepoPaths` (`select.go:196-217`) uses `GitText` plus `SplitLines`, so quoted or
    U+2028/U+0085 paths are dropped and the scan false-cleans.
* **0D643BE8: CONFIRMED.** `runRepoScan` exits 0 when the selection is empty.
* **D44D8BDF: PARTIALLY SATISFIED.**
  * Retiredarch is already contained by PR #112 (`containedRegularFile`).
  * The mergestrategy premise was refuted: it reads no git-selected worktree files.
  * Still open: the writepath repo scan (`writepath.go` `runRepoScan`) and the unignore HEAD
    root `.gitignore` read (`unignore/git.go`). Both are planned.

## Backlog created

| ID | Unit | Title | Size | Complexity | Blocked by |
|---|---|---|---|---|---|
| 051-F | n/a | Gatecheck Batch A correctness | n/a | n/a | n/a |
| 051.001-T | U1 | Centralise retiredarch fake ls-files listing helper | XS | trivial | none |
| 051.002-T | U2 | Atomic NUL-safe retiredarch selection with pin refreeze | M | medium | 051.001-T |
| 051.003-T | U3 | Fail-closed empty selection in retiredarch repo scan | S | low | 051.001-T |
| 051.004-T | U4 | Containment check for writepath repo-scan reads | S | medium | none |
| 051.005-T | U5 | Reject non-regular unignore root .gitignore at HEAD | XS | low | none |
| 051.006-T | U6 | Replace env-isolation vectors neutralised by ls-files -z | S | low | 051.002-T |

* **Shipment order:** 051-F, 051.001-T, 051.002-T, 051.006-T, 051.003-T, 051.004-T,
  051.005-T.
* **Estimate:** about 12h (6 tasks x 2h).

**Degradation: complexity.** `features.sizing` holds for size: `--size` was applied with
`--size-source agent` and `--size-ruleset-version 2h-rule-v1`. However,
`backlogit update --complexity` failed with "artifact type task does not define a
complexity field". Complexity is therefore recorded as labelled prose in each task's
`implementation-notes` section, in the form `Size: X | Complexity: Y`.

## Stash dispositions

* **4537B2F6:** consumed into 051.002-T, 051.001-T and 051.006-T. Archived.
* **0D643BE8:** consumed into 051.003-T. Archived.
  * Reconciliation: PR N/A was updated to PR #112, with provenance.
  * The thread was annotated "no late identifier found".
  * The duplicate scan was clean.
* **D44D8BDF:** consumed and narrowed into 051.004-T and 051.005-T. Archived.
  * Retiredarch was already satisfied by PR #112.
  * The mergestrategy premise was refuted.
* **New DEFERRED SCOPE EXPANSION entries (active):**
  * **B83F53BB** (high, bug): writepath `DefaultGitRunner` has no `-z` and no environment
    isolation. This is the same false-clean class as 4537B2F6, and it is the residual most
    likely to be exploited.
  * **05E12A6F** (low, task): the writepath fixture self-test reads fixtures without
    `Lstat` containment.

## Residual risks

* **RK-3 and RK-4 (plan):**
  * 051.002-T is an atomic 5-file commit (plan EX-1): pin-text drift risk, mitigated by the
    plan's §CANON text and its rollback rule.
  * 051.005-T needs a symlink-capable environment to observe its red. If none is available,
    halt under P-005.
* **B83F53BB:** the writepath selection is still not `-z` safe. 051.004-T containment cannot
  protect paths that the runner drops. No such path is tracked today.
* **05E12A6F:** fixture self-test containment remains open (low priority).

## Deviations

* **`.gitignore` committed outside the Stage artifact roots.** The operator explicitly
  authorized this. It went in a separate `chore(config)` commit on the Stage branch.
* **Complexity recorded as prose** (see "Degradation: complexity" above).

## Next steps

1. The operator pushes `chore/stage-gatecheck-batch-a-correctness` and opens the
   merge-commit staging PR.
2. Ship claims **041-S**.
