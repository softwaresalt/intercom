---
title: "Stage session — A-only simplification of the bootstrap chain"
date: 2026-09-13
agent: Stage
session_id: stage-2026-09-13-a-only-simplification
branch: chore/stage-pipeline-policy-gap
base_head: 270a8e8
decision: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
plan: docs/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md
shipment: 021-S
status: complete
---

# Stage session — A-only simplification of the bootstrap chain

**Verdict applied: `SIMPLIFICATION_VALID`.** The three-shipment `A → B → C` bootstrap is
retired for the current scope. Only **A** (`021-S` / `022-F` / `022.001-T`) is a current
prerequisite; `017-S` follows it and closes by the **existing** P-015 `CASCADE` exception.

## 1. Checkpoint recovery

* Enumerated **all** checkpoints with no `status`/`agent` filter — 11 total,
  `needs_quarantine: 0`, `quarantined: 0`. Anomaly gate ran **first** and found nothing,
  so no fail-closed halt.
* Exactly **one** `stage`-owned `active` candidate: `checkpoint-20260913-052637.json`,
  matching the operator's explicit selection. Ownership validated (`agent: stage`).
* Restored (`valid: true`), superseded its obsolete cursor
  `shipment_chain: [021-S, 022-S, 023-S, 017-S]`, and resolved it owner-exclusively.
* No `ship`-owned checkpoint was read, restored, resumed, pruned or resolved.

## 2. Backlog mutations — executed in the window-free order

| # | Mutation | Result |
|---|---|---|
| 1 | `dep add 017-S 021-S --type blocks` **while `017-S → 023-S` still blocked** | both edges present; no eligibility window |
| 2 | `shipment add 017-S 018-F` | manifest **12 → 13**; `covering_feature` resolves to `018-F` |
| 3 | `archive 022-S`, `archive 023-S` | non-destructive; `archived_from`, `archived_status: queued`, `items`, `dependencies` all preserved |
| 4 | `move 023-F / 023.001-T / 024-F / 024.001-T --status blocked` | all four blocked; members of no live shipment |
| 5 | `dep remove 017-S 023-S` **(last)** | `017-S.dependencies = [021-S]` exactly |

**Probe 24 evidence (read-back).** `017-S` manifest = exactly 13:
`018-F`, `018.008-T`, `018.001-T` + 3 `-ST`, `018.002-T` + 3 `-ST`, `018.003-T` + 2 `-ST`.

**P-015 pre-archived semantics recorded.** 11 pre-archived members **accepted** without a
status check; **required IDs are exactly three** — shipment `017-S`, root `018-F`, live
task `018.008-T`; `returned_ids` **empty**; parent IDs **preserved** (no reparenting, no
orphans).

## 3. Final state

* **Live shipments: exactly 2** — `017-S` (queued, deps `[021-S]`, 13 items, cover
  `018-F`) and `021-S` (queued, **no deps**, 2 items, cover `022-F`).
* **Only `021-S` is eligible.** Single edge, no cycle, no transient window.
* `022-S`, `023-S` archived. `023-F`, `023.001-T`, `024-F`, `024.001-T` blocked.
* Unrelated `019-F`, `020-F`, `021-F` untouched.
* One queued child under `022-F` (`022.001-T`) and one under `018-F` (`018.008-T`).

## 4. Artifact reconciliation

* **New** — the A-only decision (superseding, with options, consequences, and §7
  follow-ups).
* **Superseded in part** — three-shipment deliberation; policy-gap re-plan decision
  (§4.2 manifest shape); deferred-split/readiness-lock decision; task-only deliberation.
* **Deferred** — B and C plans, with rev-2 correctness notes (Probes 22/23/24, the
  non-constructible feature-prerequisite barrier, the agent-driven execution boundary)
  preserved verbatim for re-entry.
* **Plan A → rev 3** — sole prerequisite; routes to `017-S` not B; `017-S` re-scoped from
  the zero-feature case (i) to the **feature-present case (ii)**; rollback chain now
  A-only; budget **unchanged at ~78 min / ~42 min margin**; test rows **unchanged at 13**.
* **Umbrella → rev 8** — role-change banner, §14.0, §14.0.1, gate 9 and §14.1 all rewritten
  to the A → `017-S` route.
* **`018-F` → rev 19** — 13-entry fully-covered root; the rev-18 exclusion rationale
  withdrawn with its repeal explained; `SAFE_CLOSE` exit-9 recorded as off-route.
* **`022-F` / `022.001-T`** renamed and rewritten to the Ship covering-feature completion
  foundation; `021-S` retitled to match.
* Deferred banners added to the four B/C queue records and to nine historical memory
  snapshots.

## 5. Degradation recorded

`complexity` is **not settable** as a structured field. Re-verified this session:
`backlogit update 022.001-T --complexity high` fails with
`artifact type "task" does not define a complexity field` (exit 1), even though the
registry advertises `features.sizing: true` and the CLI exposes `--complexity`. The **WIT
type definition**, not the CLI surface, is the gate. `size` **is** structured and
preserved (`size: M`, `size_source: agent`, `size_ruleset_version: stage-2h-rule-v1`).
Complexity is carried as enum-validated prose (`high`). See
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`.

## 6. Internal plan review

Two reviewers, same-surface findings remediated in place.

* **Correctness — no P0/P1.** Architecture, authority scoping, fresh-session handoff,
  exact `017-S` CASCADE walk, budgets (§12 sums to **78**; margin **42**; §8.1 has
  literally **13** rows) and actor gates all verified **sound** against the installed
  `_ship.agent.md`, P-015 and `shipment-reconcile/SKILL.md`.
* **P2 remediated** — narrowed the "closes by the same existing exception" claim to the
  **classification** level and recorded that no probe exercises a root-*included* cascade
  with genuinely archived siblings; documented the two fail-closed gates and added a
  follow-up to probe it **before `017-S` is routed** (not before A).
* **P3 remediated** — corrected the framing of the stale
  `src/autoharness/gates/shipment_closure.py` reference: it lives in the **installed
  skill** on the `CASCADE` route of **both** A and `017-S`, not in C alone.
* **P3 remediated** — `_ship.agent.md` L776–777's pre-mode summary is false for
  pre-archived members; **explicitly scoped out** of A's 6-site inventory with rationale
  (budget pinned at ~78 min; the authoritative skill behaves correctly regardless) and
  recorded as a follow-up.
* **Scope audit** — no scope creep; zero source, test, skill, agent or policy files
  touched. Stale current-instruction claims found and fixed in the policy-gap plan
  §R16.2, the deferred-split decision, the task-only deliberation, four B/C queue records
  and nine memory snapshots.

## 7. Validation

`backlogit sync` clean (238 artifacts, `parse_failures: 0`) · `backlogit doctor` — **no
issues found** · `markdownlint` — **0 issues** · one worktree · no push, no PR, no claim.

## 8. Next actor — the OPERATOR

1. Independent adversarial review of the **A-only architecture** (scope in the session
   report).
2. **PR #54** — operator-only per §11 and Probe 17. No agent push was performed.

Ship must **NOT** claim `021-S` or `017-S` before PR #54 merges and `021-S` is present on
`main`. After that, route **`021-S` only**, then `017-S`.
