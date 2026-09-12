---
title: "Decision — Deferred-work split, readiness lock, and P-015 closure repair"
date: 2026-09-12
status: decided
agent: Stage
governs: features 018-F / 019-F / 020-F / 021-F; shipments 017-S / 018-S / 019-S / 020-S
supersedes_scope_of: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
strand_a_plan: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
strand_b_plan: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md
---

# Decision — Deferred-work split, readiness lock, and P-015 closure repair

Remediation cycle for the adversarial plan/decomposition review at `14d64e4`
(branch `chore/stage-pipeline-policy-gap`). Verdict was **MUST_REMEDIATE**: the one-task
decomposition is sound and must not be broadened; fix the current blockers and make deferred work
mechanically non-claimable and internally coherent **without widening `017-S`**.

`017-S` was **not widened**. Its executable set is still exactly one task.

## 1. P-015 closure semantics — the current blocker, and its proof

### 1.1 The defect

`017-S`'s manifest was `[018-F, 018.008-T]`. `018-F` has 12 descendants: `018.008-T` (queued) and
11 pre-archived legacy items (`018.001-T`, `018.002-T`, `018.003-T` and their eight subtasks). The
prior plan called the resulting `size_composition` roll-up "cosmetic". **It was not.** Traced
through the binding `shipment-reconcile` procedure:

1. **Step 0(c) classification** → `SAFE_CLOSE`, because `018-F` was **not fully covered** (its
   descendants at every depth were not all manifest members).
2. **Safe-close step 2** computed the **protected set** — the covering feature plus every
   unshipped sibling task sharing the hierarchy prefix that is not in the manifest, scanning both
   `queue/` and `archive/`. The 11 archived legacy items qualified. The **sequence-aware
   exclusion** did not apply: it requires a predecessor shipment with verified
   `archived_status: shipped` provenance, and these were never in any shipment, so they stay
   protected **fail-closed**.
3. **Safe-close step 3 baseline integrity gate** requires **every** protected-set member to be
   present in `.backlogit/queue/`, and grants the protected set **no pre-archived exemption**
   ("the `pre-archived` exemption applies to manifest items only — never to the protected set").
   All 11 are in `.backlogit/archive/`.
4. Result: `HALT — cascade detected, revert required` at Ship Step 6.

**`017-S` was mechanically unclosable.** This was a P0 that would have surfaced only after the PR
merged.

### 1.2 The repair, and why it is valid

Add the **complete descendant closure at every depth** to the manifest → 13 members. The
classifier then returns **`CASCADE`**, and **a `CASCADE` manifest has no protected set by
construction** (full coverage is itself a Step 0(c) precondition), so the false-cascade halt
cannot arise.

Every P-015 precondition verified mechanically against the live workspace:

| Precondition | Result |
|---|---|
| Every feature member is a **root** | `018-F.parent_id` is null ✔ |
| **Fully covered** at every depth (full `parent_id` walk over `queue/` + `archive/`) | 12/12 descendants present ✔ |
| Manifest contains **nothing extra** | every non-feature member's ancestry leads to `018-F` ✔ |
| No torn record (present in both `queue/` and `archive/`) | none ✔ |
| Childlessness positively verified | N/A — `018-F` has descendants |
| Linked deliberation | `018-F` has no `custom_fields.source_deliberation_id` and no ID matching the engine matcher `\b(?:DL\d+\|[0-9]+(?:\.[0-9]+)*-DL)\b` → cascade archives no deliberation ✔ |

`required_ids` for the two-set gate = `{017-S, 018-F, 018.008-T}`. The 11 truly-`status: archived`
members have no transition to report and are **correctly absent** from `archived_ids` — expressly
tolerated by P-015 exception item 7.

### 1.3 Operation used, and archived-status preservation

The **supported** operation `backlogit shipment add` (MCP `backlogit_add_to_shipment`) was used —
the documented frontmatter fallback was **not** needed. Verified by file hash: every archived
record is **byte-identical** before and after, still declaring `status: archived` /
`archived_status: queued`. Nothing was unarchived, moved, or executed.

### 1.4 Ship's executable set is unchanged

Verified against Ship Step 3 item 1's exhaustive positive status rule (KEEP `queued`/`active`;
SKIP-AND-REPORT `archived` as `pre_archived_skipped`; REPORT `done` as `already_done`; anything
else is a fail-closed halt), after the artifact-type filter:

```text
SHIP EXECUTABLE SET (KEEP):  ['018.008-T']
  pre_archived_skipped:      ['018.001-T', '018.002-T', '018.003-T']
  already_done:              []
  FAIL-CLOSED HALT:          none
```

The eight subtasks are not task artifacts and never enter the derivation.

## 2. Interim pipeline behaviour — stated honestly

`018.008-T` delivers the branch **gate** only. **Stage commit production is not in the current
task** (it is `019.009-T`). The honest terminal and the only reachable completion sequence are
recorded in §R16.5 of the governing plan and in `018-F`'s goals:

* Stage ends with bounded artifacts **uncommitted** on `chore/stage-{scope-slug}`.
* Installed Orchestrator Step 1.5 is **main-scoped** (`git log origin/main..main` is empty for a
  side branch) — it can neither discover the Stage branch nor guarantee PR-only persistence.
* Absent manual action it **fails closed** at Step 1.5 step 4 with `STAGING_GATE_FAIL`. That halt
  is the intended interim behaviour.
* The reachable sequence: Stage hands back → Orchestrator (under its own installed Step 1.5
  item 3 authority) **or** the operator verifies branch/worktree state → commits the Stage
  artifact roots → pushes → opens and merges the staging PR → pulls the default branch →
  re-verifies the manifest on `origin/main`. Only then may Ship be routed.

**No new authority is created.** P-010 enumerates no Orchestrator MAY/MUST NOT bullets beyond
"must not perform Stage or Ship work directly"; the commit/push/PR authority is written in the
Orchestrator's **own** installed Step 1.5 item 3. If an operator judges that arm to exceed the
Orchestrator role boundary, **the operator performs the sequence directly** — always available
under installed permissions, no contract change. A safe manual path therefore exists, so this
remediation returns **MUST_REMEDIATE-resolved**, not `MUST_REPLAN`.

**Withdrawn overclaims**: "identical to present behavior"; "no skipped assertions anywhere in the
suite"; automatic Orchestrator branch discovery; automatic PR persistence; no-shipment semantics;
generalized CI enforcement; fixture platform; regeneration proof. All either withdrawn or moved to
deferred work.

## 3. Deferred-work split — Strand A / Strand B

The former `019-F` mixed **persistence semantics** with **reusable enforcement infrastructure**.
Split into two cohesive features, each with its own shipment and plan:

| Strand | Feature | Shipment | Tasks |
|---|---|---|---|
| A — Stage handback, Orchestrator PR persistence, no-shipment semantics | `019-F` | `018-S` | `019.004-T`, `019.007-T`, `019.008-T`, `019.009-T` |
| B — detector, fixture corpus, event parser, CI, regeneration proof | `020-F` | `019-S` | `020.001-T` … `020.005-T` |

Moved with backlogit's supported `adopt` (re-IDs, rewrites cross-references, records
`origin_feature`). Nothing destructively removed.

| Former ID | New ID | Former ID | New ID |
|---|---|---|---|
| `019.001-T` | `020.001-T` | `019.005-T` | `020.004-T` |
| `019.002-T` | `020.002-T` | `019.006-T` | `020.005-T` |
| `019.003-T` | `020.003-T` | | |

`018-S`'s manifest was left dangling by `adopt` (the documented backlogit gap — `adopt` does not
rewrite `custom_fields.items`, and no supported removal operation exists). Repaired by the
precedented narrow fallback: direct `custom_fields.items` edit + immediate `backlogit sync`.
Result: `[019-F, 019.004-T, 019.007-T, 019.008-T, 019.009-T]`, zero unresolved members. Both
deferred manifests are **fully covered** and therefore P-015-`CASCADE`-qualified by construction.

### 3.1 Dependencies rebuilt

Removed as semantically wrong after the split: `019.004-T → 020.003-T` (Orchestrator discovery
does not need the detector), `020.004-T → 019.004-T` and `020.004-T → 019.007-T` (CI wiring does
not need the persistence tasks). Added: `019.004-T → 018.008-T`, `019.004-T → 019.007-T`,
`020.004-T → 020.003-T`. The strands are now independent; each depends only on `018.008-T` and its
own chain. Verified: **no cycle** in the full graph, **no `current → future` edge**.

## 4. Mechanical readiness lock

Prose labels are not enforcement. Both deferred shipments are made **mechanically ineligible**
through two independent mechanisms:

1. **Status.** `018-S` and `019-S` are `status: blocked`. Orchestrator Step 2 triggers on `queued`
   shipments only; Ship Step 0.5 item 1b independently rejects a shipment that is neither
   `queued` nor `active`.
2. **Dependency.** Each carries a `blocks` edge to readiness-gate shipment `020-S`. Orchestrator
   Step 2's explicit pre-claim re-check treats an unshipped blocking predecessor as a **hard**
   eligibility gate that outranks queue position.

**Design constraint discovered during implementation**: backlogit refuses a shipment `blocks` edge
whose other endpoint is not a shipment (`add shipment block: prerequisite 021.001-T has type
"task"; both endpoints must be shipments`). The readiness-decision item is therefore modelled as a
**lock-token shipment** `020-S` (`status: blocked`, manifest `[021-F, 021.001-T]`, never claimed,
never executed) rather than as a bare task. `021-F` / `021.001-T` are `blocked` and belong to no
other shipment, so they are never an unshipped sibling of a deferred feature and create no P-015
protected-set interaction.

**No cycle**: `020-S` has no outgoing dependency.

**Verified ineligibility** using the same queue logic the Orchestrator uses:

```text
backlogit queue view --type shipment
017-S  Stage artifact branch/PR policy gap correction     queued   high
020-S  READINESS GATE — ... (lock token, never executed)  blocked  high
```

`018-S` and `019-S` do not appear at all. `017-S` is the only `queued` shipment in the workspace.

**Reconstitution (Stage only, never Ship, never the Orchestrator).** After `021.001-T` records a
reviewed decision: `backlogit dep remove {018-S|019-S} 020-S` → restructure to the shape that
decision requires → `backlogit move … --status queued` and re-queue member tasks → `impl-plan`,
`plan-harden` (**required**) and `plan-review`. Removing the edge without the review is a P-006
violation.

## 5. The harness execution-model decision gate (`021.001-T`)

### 5.1 Why the previous guidance was invalid

The rev-1 deferred plan recommended splitting into "a sequence of **single-task shipment
manifests**". **Invalid and withdrawn**: `_ship.agent.md` Step 2 item 1 lists *all tasks for the
target **feature** that are in `queued` status* and harnesses them in one batch. Several
single-task manifests over one multi-task feature still produce one N-function red batch and still
deadlock at Step 4.3.

### 5.2 The three admissible options — none selected now

| # | Option | Requirement |
|---|---|---|
| 1 | **Single-task features and shipments** | Each covering feature holds exactly one `queued` task. Proven in practice by the reduced `017-S`. No contract amendment. Cost: more features/shipments and closure cycles; each feature must be a genuinely coherent release unit. |
| 2 | **Machine-enforced one-queued-child serialization** | One feature, at most one `queued` child at a time. **Must be proven**, not asserted, against (a) Ship Step 2 item 1's `queued`-status listing; (b) Ship Step 3 item 1's fail-closed status rule — a `blocked` manifest member is a **HALT**, not a skip, so blocked siblings must stay out of the manifest or the rule must be amended; and (c) P-015 closure, where any covering-feature descendant left out of the manifest becomes a protected-set member whose archived-or-missing state halts closure. |
| 3 | **Reviewed contract amendment making harness selection shipment-scoped** | Amend P-002/P-004, `_ship.agent.md` Step 2 and 4.3, and `harness-architect` so the harness batch and the green-after-every-task gate are scoped to the shipment manifest. A deliberate safety-contract change; requires its own deliberation, risk assessment and review. |

**No speculative mechanism is selected here.** The rev-11..15 manifest / terminal-witness /
`-gatetask` / MP0–MP6 machinery must not be resurrected without explicit re-deliberation.

## 6. Old task dispositions

| Task | Disposition |
|---|---|
| `018.001-T`–`018.003-T` (+ 8 subtasks) | Remain **archived**. Now `017-S` manifest members **for closure only**; `pre_archived_skipped` by Ship. Not unarchived, not executed. |
| `018.008-T` | **Retained**, body rewritten: operative Step 1.9 gate spec, exact/normalized locator, honest harness footprint, scoped zero-skip AC. |
| `019.004-T` | Body **fully rewritten** (Strand A). Stale rev-11..15 machinery withdrawn to non-governing provenance; cross-strand dependency removed. |
| `019.007-T`, `019.008-T`, `019.009-T` | Bodies retained (already correctly scoped); readiness-lock notice added; all three successors of the retired `018.009-T` named in full in each. |
| `019.001-T`, `019.002-T`, `019.003-T`, `019.005-T`, `019.006-T` | **Adopted** to `020-F` as `020.001-T`…`020.005-T`; bodies **fully rewritten**. Historical machinery is explicitly non-governing provenance, recoverable at `git show 14d64e4:.backlogit/queue/019.00N-T.md`. |

## 7. Residual risk, accepted

| Risk | Disposition |
|---|---|
| No mechanical CI enforcement until Strand B lands | Accepted; contract regeneration-vulnerable in the interim; tracked by `020-F`. |
| No automated persistence route until Strand A lands | Accepted; **fail-closed** at `STAGING_GATE_FAIL`; manual sequence in §R16.5 is the only completion path; tracked by `019-F`. |
| Deferred work unschedulable until the execution model is decided | Accepted and **locked** behind `020-S`/`021.001-T`. |
| `backlogit adopt` does not rewrite shipment manifests | Known gap, repaired by the precedented narrow fallback; re-recorded here for the tool owner. |
| Registry declares a `complexity` param the workspace schema lacks | Reporting defect only; complexity carried as enum-validated prose. Fixing it would mutate `.backlogit/header-def.yaml`, outside Stage's role boundary. |

## 8. Open questions for review

* **OQ-1** — Which of the three `021.001-T` options should the deferred strands adopt? Option 1 is
  the proven default; option 3 is the only one that removes the constraint permanently.
* **OQ-2** — Should the lock-token shipment pattern (`020-S`) be generalized into a harness
  convention for all deferred work, or stay a one-off for this split?
