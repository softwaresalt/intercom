# Stage session — structural re-plan to the three-shipment bootstrap

- **Date:** 2026-09-12
- **Branch:** `chore/stage-pipeline-policy-gap`
- **Base HEAD:** `e7e981d`
- **Session type:** structural re-plan following rev-6 `MUST_REPLAN` (NOT a fourth micro-remediation)
- **Resumes:** `checkpoint-20260913-043350.json` (`remediation-cycle-1-complete-must-replan-on-budget`)

## 1. What this session did

Accepted the rev-6 `MUST_REPLAN` verdict and took the re-planning decision the prior
checkpoint left to the operator. Decomposed the unsplittable ~2.5–2.7 h unit into a
**dependency-ordered three-shipment bootstrap**, each shipment carrying exactly one
covering feature and one ≤2 h task.

## 2. The decisive findings

1. **Ship's duplicated P-015 classification has already drifted unsafe.**
   `_ship.agent.md:810` says *"every one of its **children**"*; P-015 item 1 (v1.24.0,
   `workflow-policies.md:436`) requires *"descendants — at every depth"* and states a
   direct-children check *"is insufficient"*. Pre-existing latent defect, discovered here.
2. **Pre-mode requires every manifest member `done`.** `_ship.agent.md` Step 6.1(a) uses
   `expected_status: done`, so a fully-covered-root manifest needs the **feature** `done`.
   Ship's Role Boundary (line 38) permits only *"move **tasks** to active/done"* — so the
   fully-covered-root shape was **unreachable**, which is why the unit was forced into the
   task-only shape whose close path it was trying to build.
3. **The existing `CASCADE` exception already closes a fully-covered root.** T2's
   one-live-member constraint binds only the *task-only* shape. So A and B can close by the
   **existing, already-reviewed** path, and only C needs the new one.

## 3. Final topology

| Shipment | Feature | Task | Manifest | Closes by |
|---|---|---|---|---|
| **A** `021-S` | `022-F` | `022.001-T` | `[022-F, 022.001-T]` | **existing** `CASCADE` |
| **B** `022-S` | `023-F` | `023.001-T` | `[023-F, 023.001-T]` | A's a1 + existing `CASCADE` |
| **C** `023-S` | `024-F` | `024.001-T` | `[024.001-T]` (task-only) | **new** `TASK_ONLY_FINALIZE` |

Dependency chain: `017-S -> 023-S -> 022-S -> 021-S`. Stale `017-S -> 021-S` removed
**after** `017-S -> 023-S` was added, so no eligibility window opened.
**Only `021-S` is eligible.**

## 4. Inventory partition (exact)

28 closed clause sites split **6 + 6 + 16**:

- **A** ← `_ship.agent.md`: `C1–C6` (6 MUST) + `A-1` Role Boundary row + `A-2` Step 6.1(a1)
- **B** ← `workflow-policies.md`: `A1–A6` (5 MUST + 1 CONS) + `A7` dormant block + `A8` log row
- **C** ← `shipment-reconcile/SKILL.md`: `B1–B10`, `B12–B17` (16) + `B11` + `B18` + evidence

No duplicated obligation, no missing site.

## 5. Budgets

A **~71 min** · B **~66 min** · C **~106.5 min**. All ≤ 2 h. Sum 243.5 min vs the rev-6
baseline 162 min; the +81 is fully explained (3 test harnesses + 3 verification passes
= +54; A's genuinely new Role Boundary/a1 scope = +18; re-rating = +9).

## 6. Review outcome

Four personas (Correctness, Scope, Constitution, Architecture). Verdict **ADVISORY →
PASS after same-surface remediation**. 13 findings adopted and fixed in-session,
including four P1s:

1. Plan B claimed items **1–7** byte-preserved while `A5` edits item 6 — corrected to
   items **1–5 + 7 + SUPERSESSION NOTE**, with item 6 scoped and a new guard row.
2. C's live `024-F` could trip Ship's **P-001** single-in-flight gate and stall `017-S` —
   now an ordered, Stage-owned disposition step (plan C §8.3, AC-17).
3. Rollback is **not** independently reversible after activation — added a matrix
   mandating `C → B → A` order.
4. The skill references `src/autoharness/gates/shipment_closure.py`, which **does not
   exist here** — verified absent; skill markdown recorded as the execution boundary.

Not adopted: the "authority-escalation rule is YAGNI" challenge (retained — Correctness
called it load-bearing and Architecture recommended preserving it).

## 7. Verification

- `backlogit sync` → 254 artifacts, `parse_failures=0`
- `backlogit doctor` → **No issues found**
- manifests exact; dependency chain acyclic; only `021-S` eligible
- each feature has exactly **one** queued child (satisfies the P-004 / Step 4.3 batch constraint)
- all shipment statuses `queued` (valid enum)
- archive **untouched** (`git status` on `.backlogit/archive/` empty); stash `A10EF3D0`
  archived lineage preserved; active stash 13 entries unchanged
- `markdownlint` → 0 issues across 5 docs
- **one** worktree; artifacts-only diff (`docs/**`, `.backlogit/**`)

## 8. Known degradation

`complexity` is **not** settable as a structured field — `.backlogit/header-def.yaml`
defines no `complexity` field on the task type and
`backlogit update --complexity high` fails validation. Carried as **enum-validated prose**
on all three tasks with an explicit degradation note. `size: M` **is** structured.
See `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`.

## 9. Next actor — the OPERATOR

**No push was performed.** Stage's output is local commits on
`chore/stage-pipeline-policy-gap`.

1. **Independent multi-model adversarial review** — scope in the deliberation §11 and the
   summary returned to the operator. This is a **new structural architecture**, so the
   rev-6 §13 scope is superseded.
2. **PR #54** — operator-only (Probe 17 stands; no agent-executable route). Carries all
   three shipments and all dependency edges **atomically**. Gates in umbrella §14.
3. After merge, Orchestrator routes **`021-S` only**, then B, C, then `017-S`, each
   sequentially under P-001/P-016 with P-020 closure between.

**Ship must NOT claim `021-S`, `022-S`, `023-S` or `017-S`.** `017-S` remains
`queued`/unclaimed.

## 10. Residuals carried forward

- **CO-4** semantic drift between `T1–T5` (policy) and `G1–G5` (skill) — accepted; CI
  coupling gate out of scope.
- **C's ~13.5 min margin** — real, bounded to one shipment, with a HALT path; valve limits
  stated explicitly.
- **Plan A AC-9 tamper-evidence** — the checkpoint is evidence, not enforcement.
- **`021.001-T` still `blocked`** though its deliberation is `decided`. Out of scope: it
  governs the separate `019-F`/`020-F` deferred strand, not this chain.
