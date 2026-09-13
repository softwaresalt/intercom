---
title: "Deliberation — Three-shipment bootstrap for task-only shipment finalization"
date: 2026-09-12
status: decided
agent: Stage
governs: features 022-F, 023-F, 024-F; shipments 021-S, 022-S, 023-S
supersedes_scope_of: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
source_stash: A10EF3D0 (archived)
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Deliberation — Three-shipment bootstrap for task-only shipment finalization

## 1. Problem frame

Plan rev 6 of the task-only shipment finalization unit returned **`MUST_REPLAN`**. The
verdict is correct and is **not** re-litigated here. Its two findings are adopted as
premises:

| Premise | Source | Status |
|---|---|---|
| The unit prices at **~2.5–2.7 h** against a **closed 28-site** clause inventory | rev 6 §10.1 | adopted |
| The unit **cannot be split under one covering feature** | rev 6 §10.2 | adopted |

Rev 6 §10.2 grounds the no-split finding in two independent constraints:

1. **T2** admits exactly **one live manifest member** at closure. Two completed tasks
   present two live members, so the prerequisite could not classify itself under the
   contract it ships.
2. The instruction surfaces must be **mutually consistent before the unit self-closes**.
   A partial landing leaves P-015 authorizing a path `_ship.agent.md` still forbids.

A third constraint, inherited from `021.001-T` and the `018-F` re-plan decision, binds
any decomposition:

1. **P-004** requires every generated test function red before implementation; **Ship
   Step 4.3** requires the full `go test ./...` green after **every** task; **Ship Step 2
   item 1** harnesses all **queued tasks of the covering feature** in one batch.
   Therefore any covering feature with **more than one queued test-bearing task
   deadlocks**. This is arithmetic, not sequencing.

So the admissible shape is fixed before any option is considered: **one queued
test-bearing task per covering feature, one covering feature per shipment.**

The question this deliberation answers is therefore **not** "how do we split the task"
but: **is there a decomposition into dependency-ordered single-task shipments in which
every intermediate merge leaves a coherent, fail-closed, executable contract?**

## 2. What rev 6 left open

Rev 6 §10.1 enumerated four candidate directions and deliberately took none:

| # | Direction | Disposition here |
|---|---|---|
| 1 | Defer CONSISTENCY edits, re-measure | **Partially adopted** — retained as C's declared overflow valve (§7.3), not as the primary mechanism |
| 2 | Move the skill's normative bulk into P-015 | **Rejected** — §4 option 2 |
| 3 | Operator-authorized 2-hour-rule exception (P-005 deviation) | **Rejected** — §4 option 3 |
| 4 | Change the self-hosting approach so the unit need not close itself | **ADOPTED and extended** — §4 option 4 |

Direction 4 is the load-bearing insight, and rev 6 stated its consequence precisely:
changing the self-hosting approach *"dissolves the T2 one-live-member constraint and
re-enables splitting."*

This deliberation establishes **how** — and finds that the dissolution is available
today, with no new mechanism, because **the existing P-015 `CASCADE` exception already
closes a fully-covered-root manifest**. T2 binds only the *task-only* close shape. A
manifest of `[feature, its one task]` is a fully-covered root, closes under the
**existing, already-reviewed** exception, and imposes no one-live-member constraint at
all.

## 3. The decisive discovery

Two findings, verified against the installed files at `e7e981d`, make the decomposition
both possible and necessary.

### 3.1 Ship's duplicated classification has already drifted unsafe

`_ship.agent.md` (lines ~801–823) **re-derives** P-015's classification in its own prose
instead of delegating to it. That duplicate is **already wrong**:

| Surface | Coverage rule |
|---|---|
| `workflow-policies.md` P-015 item 1 (v1.24.0) | *"every one of its **descendants — at every depth, not only direct children** — enumerated by walking the full `parent_id` graph"*; and explicitly: *"A check limited to direct children is insufficient"* |
| `_ship.agent.md` line ~810 | *"every one of its **children**, enumerated live from `.backlogit/queue/` + `.backlogit/archive/`, is also a manifest member"* |

P-015 v1.24.0 corrected exactly this defect in the policy and the duplicate was never
updated. A manifest `[feature, task]` where that task owns an out-of-manifest subtask
**passes Ship's copy and fails the policy** — which is the precise corruption the
v1.24.0 amendment was written to prevent.

This is a **pre-existing latent defect discovered during this re-plan**, not a
consequence of it. It converts "replace the duplicate with a delegation" from a
stylistic preference into a **correctness fix that stands on its own merits**.

It also supplies the structural key: once Ship **delegates** rather than **restates**,
every future change to the set of authorized close paths is confined to P-015 and the
skill. **Ship never needs editing again** for `TASK_ONLY_FINALIZE`. That is what
severs the "all three files must land together" coupling in rev 6 §10.2 finding 2.

### 3.2 Pre-mode requires every manifest member to be `done`

`_ship.agent.md` Step 6.1(a) invokes `shipment-reconcile` with `expected_status: done`
and *"verifies that every manifest item is present in queue with `status: done`"*.

So a manifest of `[022-F, 022.001-T]` requires the **covering feature** to be `done`
before pre-mode. Ship's Role Boundary currently permits only *"move **tasks** to
active/done"* — there is **no** feature-completion authority. The fully-covered-root
shape is therefore **unreachable today**, which is why the original unit was forced into
the task-only shape whose close path is the very thing being built.

Granting that authority — narrowly, fail-closed, positioned immediately before pre-mode —
is what makes the fully-covered-root shape reachable, and it is a **single-surface,
small, self-contained change to `_ship.agent.md` only**.

## 4. Options considered

### Option 1 — Three-shipment dependency-ordered bootstrap (**CHOSEN**)

Decompose by **surface**, one surface per shipment, ordered so that each merge leaves a
coherent contract and each shipment closes by a path that is **already live when it
needs it**.

| Shipment | Surface | Manifest shape | Closes by |
|---|---|---|---|
| **A** Foundation | `_ship.agent.md` + 1 Go test | fully-covered root | **existing** `CASCADE` |
| **B** Policy gate | `workflow-policies.md` P-015 + 1 Go test | fully-covered root | A's feature-completion + existing `CASCADE` |
| **C** Skill impl. | `shipment-reconcile/SKILL.md` + 1 Go test | **task-only** | the new `TASK_ONLY_FINALIZE` |

The bootstrap property: **A and B close by the existing, already-reviewed `CASCADE`
path. Only C closes by the new path — and by then A, B and C's own change are all
live.** Nothing closes by a path that does not yet exist.

**Intermediate-state coherence** (the requirement rev 6 §10.2 finding 2 protects):

* **After A merges.** Ship delegates close-path selection to the machine verdict. P-015
  and the skill are unchanged, so the only verdicts available are `CASCADE` and
  `SAFE_CLOSE`/`HALT` — exactly today's behavior, minus the drifted duplicate. Coherent.
* **After B merges.** P-015 authorizes `TASK_ONLY_FINALIZE` **only when the installed
  reconciliation skill advertises the required capability/schema token**. No installed
  skill advertises it, so it can never be selected. Behavior is byte-for-byte today's
  behavior. **Fail-closed by construction, not by discipline.** Coherent.
* **After C merges.** The token appears, the classifier activates, and C self-closes
  task-only. Coherent, and the self-close is the durable runtime proof.

The dormant-token gate is what makes B safe to land alone. It is not a feature flag
that someone must remember to leave off; **absence of the token is the default and
authorization is impossible without it.**

**Cost.** Three features, three shipments, three closure cycles instead of one.

### Option 2 — Move the skill's normative bulk into P-015

Rejected. Rev 6 already identified the defect: *"It cannot be compressed to a pointer
without moving the contract somewhere that Ship does not reload."* It also inverts the
correct layering — P-015 is the authorization surface, the skill is the procedure
surface — and would leave the skill not self-contained. It reduces one task's budget by
relocating work, not by decomposing it, so the aggregate still exceeds 2 h in one unit.

### Option 3 — Operator-authorized 2-hour-rule exception

Rejected. The 2-hour rule is a **reliability** constraint (agent success drops below 50%
past ~2 h), not a bureaucratic one. Waiving measurement does not change the failure
probability. It would also set a precedent that a structurally indivisible unit may
exceed budget, when §3 shows this unit is **not** structurally indivisible — the
indivisibility was an artifact of the task-only close shape, which Option 1 removes.

### Option 4 — Two shipments (fold A into B)

Rejected. Ship's Role Boundary change and the P-015 amendment are different surfaces
with different risk profiles, and folding them re-creates a ~2.2 h unit. More
decisively, **A must be live before B can close**: B's manifest is a fully-covered root,
so B's closure needs A's feature-completion authority. A combined A+B shipment would have
to close using an authority introduced in its own merge — the exact mid-session
escalation §5 prohibits.

## 5. The authority-escalation rule (new, load-bearing)

A merged change that alters Ship's own **Role Boundary** must **not** take effect
mid-session.

The existing *"Mandatory pre-self-close context reload"* requires Ship to re-read freshly
merged `main` instructions before closing. Read naively, that would let the session that
merged A immediately use the authority A introduced — an agent granting itself new
authority by merging a change to its own contract.

The distinction this deliberation records, and which A makes explicit:

* A reload MAY update the **procedure** a session follows.
* A reload MUST NOT widen the **authority envelope** of the session performing it. Role
  Boundary is a session-identity constraint fixed at session start.

Therefore the session that merges A **checkpoints and ends**. A **fresh** session loads
the new Role Boundary and completes the closure. This is both the safety property and
the self-hosting proof: *no authority is gained mid-session*, demonstrated by execution
rather than asserted.

## 6. Inventory split — exact, no duplication, no gap

Rev 6's closed inventory is **28 clause sites** = 20 MUST + 8 CONSISTENCY, over rows
`A1–A6`, `B1–B10`, `B12–B17`, `C1–C6`. The remaining six rows (`A7`, `A8`, `B11`, `B18`,
`D1`, `D2`) are additive blocks and artifacts, budgeted separately in rev 6 §10.1. Total
distinct rows: **34**.

| Shipment | Surface | Clause sites (of the 28) | Additive rows |
|---|---|---|---|
| **A** | `_ship.agent.md` | `C1 C2 C3 C4 C5 C6` = **6** (6 MUST) | own Go test |
| **B** | `workflow-policies.md` | `A1 A2 A3 A4 A5 A6` = **6** (5 MUST + 1 CONS) | `A7`, `A8`, own Go test |
| **C** | `shipment-reconcile/SKILL.md` | `B1–B10`, `B12–B17` = **16** (9 MUST + 7 CONS) | `B11`, `B18`, §6.1-E inventory, own Go test, runtime probes |

**6 + 6 + 16 = 28.** Exact partition. Every site lands in exactly one shipment; no site
is dropped; no obligation is duplicated.

Rev 6 internal inconsistency noted and resolved: `B17` is classed CONSISTENCY in Table
§6.1-B but budgeted as MUST-mechanical in §10.1. This re-plan follows **§10.1** (MUST),
which is the stricter reading, and assigns it to C with the rest of the skill surface.

## 7. Budgets

Per-task estimates use rev 6 §10.1's own per-unit rates, so they are directly comparable
to the measurement that produced `MUST_REPLAN`.

| Shipment | Work | Estimate |
|---|---|---|
| **A** | Role Boundary narrowing (~8) + parent-completion step (~10) + C1 (~4) + C2 (~4) + C3/C4/C5 delegation (~12) + C6 (~4) + Go test (~18) + verification (~10) | **~70 min ≈ 1.2 h** |
| **B** | A1 (~1.5) + A2/A3 (~3) + A4/A5/A6 (~12) + A7 dormant block (~20) + A8 (~1) + Go test (~18) + verification (~10) | **~66 min ≈ 1.1 h** |
| **C** | 8 MUST clause edits (~28.5) + B17 (~1.5) + 7 CONSISTENCY (~10.5) + B11 (~3) + B18 (~28) + §6.1-E (~5) + Go test (~20) + verification (~10) | **~107 min ≈ 1.8 h** |

All three close under the 2-hour rule. **A and B carry comfortable margin; C is tight
(~13 min).**

### 7.3 C's declared overflow valve

C's margin is thin enough to name a pre-authorized reduction rather than leave the
implementer to improvise one. If C crosses **100 minutes** with CONSISTENCY edits
outstanding, a **bounded 5-clause subset** (`B2 B4 B6 B14 B16`, ~7.5 min) is deferred to a
follow-up hygiene stash entry. Those five are branch-local descriptive statements that
remain **literally true** after the amendment.

**`B7` and `B12` are explicitly NOT deferrable** (review correction): they assert
*sufficiency* and *exclusivity* respectively, which the amendment falsifies, so deferring
them would leave the contract **incorrect**, not merely uneven.

**The valve's limits are stated, not glossed.** It sheds ~7.5 min and therefore cannot
rescue a `B18` or Go-test overrun. That case is an explicit **HALT and return to Stage**.
Deferring any **MUST** site is **not** authorized. See plan C §12.1.

### 7.4 Cross-surface coupling matrix

An earlier draft claimed **CO-1 (the token string) is the only cross-shipment coupling**.
Review established that is **wrong** — it is the only coupling requiring a *literal string
match*, but not the only one. The full set, each with a producer, consumer and guard:

| # | Coupling | Producer | Consumer | Guard |
|---|---|---|---|---|
| **CO-1** | Capability token `task-only-finalize/v1` | C (skill frontmatter) | B (P-015 gate) | exact-parity test row, **both** sides |
| **CO-2** | Verdict literal `TASK_ONLY_FINALIZE` | B (authorizes) | C (emits), A (invokes) | B row 1, C rows 3–4 |
| **CO-3** | Classifier order `CASCADE → TASK_ONLY_FINALIZE → SAFE_CLOSE/HALT` | B | C (selector) | B row 2, C row 3 |
| **CO-4** | Topology `T1–T5` ↔ guards `G1–G5` | B (asserts) | C (implements) | C row 12 (cite-not-restate); **semantic**, not textual |
| **CO-5** | Zero-feature-member mutual exclusion | B | C | B row 6 |
| **CO-6** | Verdict-delegation contract (Ship invokes only what the verdict names) | A | B, C | A rows 4, 9 |
| **CO-7** | Reload set must include P-015 + skill | A (C1) | C (activation visibility) | A row 5 |
| **CO-8** | Feature-completion authority (a1) | A | B (needs it to close) | `022-S depends_on 021-S` |
| **CO-9** | `P1–P10` ordering vs P-007 archive restoration | C | C | C row 6 |
| **CO-10** | Engine CLI version + digest | C | C | §5 HALT |
| **CO-11** | P-001 single-in-flight vs C's live `024-F` | C | `017-S` routing | plan C §8.3 disposition |

**CO-4 is semantic, not textual, and that is the honest characterization.** `G1–G5`
necessarily *implement* `T1–T5`. C cites rather than restates them (row 12), which prevents
textual duplication but does **not** eliminate the semantic dependency. Drift between them
is a **real residual**, accepted: a CI coupling gate is out of scope for this chain.

**CO-8 and CO-11 are why the ordering is load-bearing**, not merely tidy: B cannot close
before A is live, and `017-S` cannot route before `024-F` is disposed of.

## 8. Decision

**Adopt Option 1.** Re-plan the closed unit as a dependency-ordered three-shipment
bootstrap:

```text
A (021-S)  →  B (022-S)  →  C (023-S)  →  017-S
```

with `B depends_on A`, `C depends_on B`, `017-S depends_on C`, and the stale
`017-S -> 021-S` edge removed **only after** the `017-S -> C` edge exists, so no
eligibility window opens.

Each shipment carries exactly **one covering feature and one ≤2 h task**, satisfying
P-004 / Ship Step 4.3 / Ship Step 2 item 1 simultaneously.

## 9. Rejected-option reasoning, recorded

| Option | Why rejected |
|---|---|
| 2 — bulk into P-015 | Relocates work rather than decomposing it; inverts authorization/procedure layering; skill stops being self-contained |
| 3 — 2-h rule exception | Waives the measurement, not the failure probability; §3 shows the unit is not actually indivisible |
| 4 — two shipments | Re-creates a ~2.2 h unit; B's closure would depend on an authority introduced in its own merge (§5 violation) |

## 10. Open questions carried into planning

1. **Is `022-F` `active` at the parent-completion gate?** The authorized transition is
   `active -> done`. If shipment claim leaves the covering feature `queued`, the gate
   **fails closed** (correctly) and A cannot self-close. A's plan carries this as an
   explicit **probe obligation** and a fail-closed HALT, not an assumption. Resolved in
   A's plan §Plan Hardening.
2. **C leaves `024-F` live after task-only closure.** T4 requires the root parent feature
   live and outside the manifest, so this is by design, not a leak. **Review escalated
   this**: a live `024-F` can trip Ship Step 1's P-001 single-in-flight gate (line 308) and
   stall `017-S`. Disposition is now an **ordered, Stage-owned step** (plan C §8.3,
   AC-17), executed before `017-S` routes — **not** a vague follow-up.
3. **Engine digest.** Mismatch against
   `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` remains **HALT +
   full probe refresh** for C. Unchanged from rev 6; the rev-4.1 advisory fallback stays
   withdrawn.

## 11. Plan review record (Stage internal, multi-persona)

Four independent reviewer personas ran against the deliberation and plans A/B/C at
`e7e981d`: **Correctness**, **Scope Boundary**, **Constitution**, **Architecture**.

### 11.1 Verdict

**ADVISORY → PASS after same-surface remediation.** No finding required re-planning; all
P1 findings were same-surface fixable and were fixed in this revision.

### 11.2 Findings adopted and remediated

| # | Finding | Severity | Remediation |
|---|---|---|---|
| 1 | Plan B claimed items **1–7** byte-preserved while `A5` edits item 6 — unimplementable as written | **P1** | Preserved region corrected to items **1–5 + 7 + SUPERSESSION NOTE**; item 6 scoped by `A5`; new test row 12 bounds it to scoping only (B §4.1, AC-6) |
| 2 | C leaves `024-F` live, which can trip Ship's **P-001** single-in-flight gate (line 308) and stall `017-S` | **P1** | New **§8.3**: ordered, Stage-owned disposition with named operations and preconditions, before `017-S` routes (C AC-17) |
| 3 | Rollback described as independently revertible; reverting B with C installed leaves an **unauthorized** implemented verdict | **P1** | New **§10.1 rollback matrix**; `C → B → A` reverse order **mandatory** after activation |
| 4 | Skill references `src/autoharness/gates/shipment_closure.py`, which **does not exist** here; "0 source changes" was an unchecked assumption | **P1** | Verified absent; **§2.2** records the skill markdown as the execution boundary; stale refs reconciled inside existing `B9`/`B17` (C AC-16) |
| 5 | C's overflow valve mis-targeted, ordering-incoherent, and over-broad (`B7`/`B12` are not branch-local) | **P2** | **§12.1** rewritten: CONSISTENCY sequenced **last**; deferrable subset reduced to 5 clauses/~7.5 min with per-clause justification; `B7`/`B12` non-deferrable; `B18` overrun ⇒ **HALT** |
| 6 | C1's carve-out and the retained *"close under the just-merged contract"* left unreconciled and untested | **P2** | **§4.2** pins the reconciled wording; new Go test row 11 guards it (A AC-14) |
| 7 | Manifest re-shape to `[022-F, 022.001-T]` had no named owner or AC | **P2** | **§4.3** names Stage as owner, pre-claim; A AC-12 |
| 8 | `PO-1` ran at closure — the most expensive moment to discover a claim-shape mismatch | **P2** | Split into **PO-1a** (pre-execution) + **PO-1b** (at gate); A AC-10 |
| 9 | "CO-1 is the only cross-shipment coupling" — false | **P2** | Replaced by the **§7.4 coupling matrix** (CO-1…CO-11); CO-4 named as semantic and its drift accepted as residual |
| 10 | Token bound to incidental skill revision rather than contract compatibility | **P2** | Compatibility-level token `task-only-finalize/v1` (B §5) |
| 11 | Plan A §9 step 1 showed Ship moving `022-F queued -> active` — unauthorized, and contradicted PO-1 | **P2** | §9 step 1 corrected: `022.001-T -> active` only; `022-F` status **observed**, not asserted |
| 12 | Probe 18 justifies A's `C2` but was listed only under C | **P3** | **§10.1.1** — Probe 18 travels with A (A AC-13) |
| 13 | A's AC-9 two-session proof is not tamper-evident | **P3** | AC-9 strengthened to require a tool-written `backlogit_create_checkpoint` artifact; **residual stated honestly** |

### 11.3 Findings considered and NOT adopted

| Finding | Severity | Disposition |
|---|---|---|
| Authority-escalation rule (A §6) is "borderline YAGNI" — a general invariant serving one event | P2 | **Retained.** Correctness review called it load-bearing and Architecture review explicitly recommended preserving it. A procedural-only handoff would leave the privilege-escalation hazard unstated in the contract |
| Backlog items `022-F`/`022.001-T`/`021-S` still carry superseded rev-5 scope | P1 | **Not a plan defect** — reviewers ran **before** harvest. Resolved by Step 5 of this session |
| `021.001-T` still `blocked` though its deliberation is `decided` | P3 | **Out of scope.** `021-F`/`021.001-T` is the separate deferred-readiness strand governing `019-F`/`020-F`; its AC-5 does not govern this prerequisite chain. Recorded as a remaining observation |
| Architecture verdict **UNSOUND** pending classifier definition + `024-F` disposition + coupling contract | P1 | **Conditions met.** All three were the reviewer's stated preconditions for `SOUND WITH RESERVATIONS`, and all three are remediated above (items 2, 3, 4, 9). The reviewer's own recommendation — *"Keep the A → B → C ordering"* — is adopted unchanged |

### 11.4 Residuals carried forward

* **CO-4 semantic drift** between `T1–T5` (policy) and `G1–G5` (skill) — accepted; a CI
  coupling gate is out of scope.
* **C's ~13.5 min margin** — real, bounded to one shipment, with a HALT path.
* **A AC-9 tamper-evidence** — the checkpoint is evidence, not enforcement.

## 12. What this deliberation does not change

* The `MUST_REPLAN` verdict on rev 6 — adopted, not overturned.
* The probe evidence base (Probes 14–20) — retained; the rev-6 plan becomes the
  **umbrella evidence and decision source** that A, B and C cite.
* Live-config evidence and recovery semantics from rev 6 — retained verbatim for C.
* The PR #54 actor boundary — the **operator** remains the sole push/merge actor. No
  agent push. Unchanged from rev 6 §11 / §14.
* `017-S` — remains `queued` and unclaimed throughout.
