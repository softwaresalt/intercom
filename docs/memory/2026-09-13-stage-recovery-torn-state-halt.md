---
title: "Stage recovery session — torn-state HALT on checkpoint-20260913-194412"
date: 2026-09-13
agent: Stage
session_id: stage-2026-09-13-recovery-194412
branch: chore/stage-pipeline-policy-gap
status: halted-awaiting-operator-authorization
restored_checkpoint: checkpoint-20260913-194412.json
checkpoint_disposition: NOT RESOLVED (torn state)
---

# Stage recovery session — torn-state HALT

Operator selected `checkpoint-20260913-194412.json` and authorized resume. Restore
succeeded; **resume did not complete**. The checkpoint was **not resolved**, because the
recorded state and the on-disk state diverge in ways my Crash-Resumption / Startup Recovery
Protocol classifies as **torn**.

## 1. Recovery protocol execution

| Step | Result |
|---|---|
| Unfiltered enumeration (`checkpoint list`, no `--agent`/`--status` filter) | 16 checkpoints; 1 `ship`, 15 `stage` |
| Anomaly/quarantine gate (run first, on full enumeration) | **Clean** — no invalid, quarantined, or malformed record |
| Active `stage` candidates | 2 — `checkpoint-20260913-194412.json`, `checkpoint-20260913-084125.json` |
| Operator selection | Explicit, unique: `checkpoint-20260913-194412.json` |
| Owner validation (`agent` field) | `stage` — ownership matches |
| Restore (`checkpoint get`) | Succeeded; schema_version 1, valid |
| Prune-on-restore (engram reachable, `engram stats` OK) | Performed read-select-summarize; **preserved** active cursor (`021-S`/`022-F`/`022.001-T`), unresolved-checkpoint pointer (`084125`), and gate verdicts (attempt 4 FAIL, H-1..H-9) |
| Resume | **HALTED** — torn state (§2) |
| `checkpoint resolve` | **NOT invoked** — protocol forbids resolution on torn/ambiguous state |
| `checkpoint-20260913-084125.json` | **UNTOUCHED**, remains `active` (reported only) |

## 2. Torn state — post-checkpoint divergence

The checkpoint was written at `2026-09-13T19:44:12Z` (12:44:12 PDT) with
`base_head: 0e18cd1`. Work landed **after** it that the checkpoint does not record.

| Dimension | Checkpoint (recorded) | On disk (observed) |
|---|---|---|
| HEAD | `0e18cd1` | `b65b3e3` (+2 commits) |
| Governing plan | `docs/plans/2026-09-12-...-foundation-plan.md` | `docs/plans/2026-09-13-...-decided-plan.md` (new path) |
| Plan revision | 7 | 8 |
| Foundation plan | live at `docs/plans/` | moved to `docs/archive/plans/`, `status: superseded` |
| Gate verdict | attempt **4**, FAIL, 9 open P1 | attempt **2**, `decision: PENDING` |
| Untracked anomaly file | present, needs provenance adjudication | **absent** |

* `4abea02` (12:44:51 PDT) — rev 7 remediation + attempt 4 FAIL. Consistent with the checkpoint.
* `b65b3e3` (12:57:48 PDT) — **13 minutes after the checkpoint**; "consolidate plan A into a
  decided plan and archive the rev history". Not recorded in any checkpoint.
* Uncommitted working-tree edits (decided plan last written 13:59:07 PDT) — a further
  in-flight revision, still uncommitted.

## 3. Blocking findings

### F-A — Review-cycle counter reset across an artifact rename (circuit breaker)

True lineage, from the archived plan's own append-only markers:
`attempt 1 → 2 → 3 → 3 → 4`. Checkpoint phases corroborate independently:
attempt 1 FAIL → 2 FAIL → 3 FAIL (`cycles-exhausted-ESCALATED`) → 4 FAIL
(`exceptional-cycle-spent`).

The consolidated artifact restarts its counter at `plan-review-attempt: 1`, and the
uncommitted edit bumps it to `2`, with `decision: PENDING`.

Effect: the plan-review circuit breaker (max 2 re-entry cycles; attempt ≥3 escalates) reads
as "attempt 2 of 3" on the new path, when the work item is actually at **attempt 5** with the
exceptional budget already **SPENT**. Continuing the gate here would consume a cycle that no
authorization covers. I did not run it.

### F-B — The 9 open P1 blockers are absent from the sole governing artifact

`H-1`..`H-9` occur **0 times** in the decided plan. They survive only in the backlog records
(`021-S`, `022-F`, `022.001-T`, `017-S`), the memory file, and the archived plan.

The decided plan declares itself "the SOLE GOVERNING ARTIFACT" and instructs readers:
"Do not apply ... review findings from the archived file." The archive is exactly where
H-1..H-9 live. The §13 correction table tracks a **different, earlier** set (`F1`–`F8`).

Net effect: 9 unresolved P1 blockers were made non-governing by relocation rather than by
remediation. I will not ratify that silently.

### F-C — Untracked bug-report file vanished, unexplained

`docs/decisions/2026-09-13-autoharness-append-only-plan-review-loop-bug-report.md` is absent
from disk **and** from all git history on all branches (`git log --all -- <path>` is empty).
The checkpoint recorded it as present, not Stage-authored, and deliberately uncommitted.

Stage does not own it and has no evidence of its disposition. Not adopted, not recreated,
not committed. Flagged for provenance adjudication.

### F-D — Backlog split-brain persists

`022-F`'s body still points at "Plan rev 3: `docs/plans/2026-09-12-...-foundation-plan.md`" —
a stale revision at a path that is now archived. The consolidation appended a superseding
block rather than correcting the stale pointer. This is the recurrence the checkpoint names
as "the 4th backlog split-brain" (H-2).

## 4. State verified unchanged (no execution occurred)

* `021-S` `queued` — **not claimed**, not exercised. `017-S` `queued`, ineligible.
* `022-F` `queued`, `022.001-T` `queued`. No active shipment.
* `PA-021-CASCADE` / `PA-017-CASCADE`: recorded, **unexercised**.
* `backlogit doctor`: `No issues found.` Engram reachable (297 symbols, 100% coverage).
* PR #54 OPEN, head `8999867`, base `main`, `CLEAN`/`MERGEABLE`. Local branch is **39 commits
  ahead** of the PR head. Not pushed.
* One worktree. No branch created. No source/test/config file touched.

## 5. Authorization required to proceed

Stage cannot complete without **one** of:

1. **Explicit authorization for a further exceptional plan-review/remediation cycle**, naming
   the true lineage (this would be **attempt 5**), since the budget is recorded SPENT; **and**
2. **Adjudication of the governing-artifact swap** — either (a) H-1..H-9 are carried forward
   into the decided plan and remediated under the authorized cycle, or (b) each is explicitly
   withdrawn on the record with rationale. Silent lapse via the archive pointer is not
   acceptable.
3. **Provenance ruling on F-C** — whether the vanished bug-report file was operator-removed
   (no action) or lost (needs reconstruction by its owner).
4. **Ruling on the counter reset** — whether the decided plan's counter is corrected to
   continue the true lineage, or the reset is ratified with reasons.

Until then: plan stays `planned`, **not** harvest-ready; `021-S` not claimable; no push to
PR #54; no harvest; no shipment assembly.
