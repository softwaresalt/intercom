---
title: "Deliberation — Three-shipment bootstrap for task-only shipment finalization"
date: 2026-09-12
status: superseded-in-part
superseded_in_part_by: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
superseded_scope: "the A → B → C three-shipment chain, for the CURRENT release scope only; A's own analysis stands"
agent: Stage
governs: features 022-F, 023-F, 024-F; shipments 021-S, 022-S (archived), 023-S (archived)
supersedes_scope_of: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
source_stash: A10EF3D0 (archived)
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Deliberation — Three-shipment bootstrap for task-only shipment finalization

> **SUPERSEDED IN PART — read this first.**
> The **three-shipment chain** decided here is **superseded for the current release
> scope** by
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (verdict `SIMPLIFICATION_VALID`).
>
> **What changed.** This deliberation assumed `017-S` must stay a **task-only** manifest
> and therefore required a new `TASK_ONLY_FINALIZE` close verdict (**B**) and its
> classifier (**C**). Shipment **A** introduces **Step 6.1(a1)**, which moves a covering
> feature `active -> done` before pre-mode. That dissolves the assumption: `017-S` is
> restored to a **fully-covered root** (13 members) and closes by the **existing** P-015
> `CASCADE` exception. B and C are therefore **not** current prerequisites.
>
> **What still stands.** Everything about **A** (`021-S` / `022-F` / `022.001-T`), the
> adopted rev-6 `MUST_REPLAN` premises, and the `021.001-T` one-queued-test-bearing-task
> arithmetic. **B and C are deferred generalized platform work**, `blocked`, with their
> shipment records `022-S` / `023-S` archived non-destructively. Their rev-2 correctness
> notes are preserved for whenever they are re-entered under fresh, separately reviewed
> shipments.
>
> The chain reasoning below is retained **as historical record**. Do not route from it.

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
| 1 | Defer CONSISTENCY edits, re-measure | **Rejected (rev 2)** — was "partially adopted" as C's overflow valve; §7.3 removes the valve entirely. Deferring any site is no longer authorized |
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
| **C** Skill impl. | `shipment-reconcile/SKILL.md` + 1 Go test | fully-covered root *(rev 2 — was task-only)* | the existing `CASCADE` *(rev 2 — was the new `TASK_ONLY_FINALIZE`)* |

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
* **After C merges.** The token appears and the classifier activates. **Rev 2:** C does
  **not** self-close by it — C's manifest is a fully-covered root, so T3 excludes
  `TASK_ONLY_FINALIZE` and C closes by `CASCADE` like A and B. The activated verdict's
  first live selection is at **`017-S`**. Still coherent; the runtime proof is relocated to
  Probe 21's post-half against the exact `017-S` 12-member fixture (plan C §8.4), and the
  residual that no *live* task-only closure precedes `017-S` is recorded as plan C **R-8**.
  See §8.1 for the measured basis.

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
| **A** | Role Boundary narrowing (~8) + parent-completion step incl. applicability/idempotence (~15) + C1 (~4) + C2 (~4) + C3/C4/C5 delegation (~12) + C6 (~4) + Go test 13 rows (~21) + verification (~10) | **~78 min ≈ 1.3 h** |
| **B** | A1 (~1.5) + A2/A3 (~3) + A4/A5/A6 (~12) + A7 dormant block (~20) + A8 (~1) + Go test 12 rows (~19) + verification (~10) | **~67 min ≈ 1.1 h** |
| **C** | 8 MUST clause edits (~28.5) + B17 two lines (~3) + 7 CONSISTENCY (~10.5) + B11 (~3) + B18 (~29) + Go test 13 rows (~21) + verification (~10) | **~105 min ≈ 1.75 h** |

All three close under the 2-hour rule. **A and B carry comfortable margin; C is tight
(~15 min).**

**Rev 2 re-measurement.** A **+7** (the §5.1 three-way a1 applicability and §5.2 idempotent
resume, plus Go rows 12–13). B **+1** (the §7.1 row-count text said 11 while the table
listed 12; corrected to 12). C **−1.5** net (`B17` +1.5 for its second line, `B18` +1 for
the token-activation rule, Go test +1 for row 13, committed inventory re-scan **−5**
removed). Evidence generation — including Probe 21's relocated post-half — is budgeted
under **Ship closure**, not task coding.

### 7.3 C's declared overflow valve — **REMOVED in rev 2**

Rev 1 declared a pre-authorized overflow valve for C: crossing 100 minutes with
CONSISTENCY edits outstanding would defer a bounded 5-clause subset (`B2 B4 B6 B14 B16`,
~7.5 min) to a follow-up hygiene stash entry, with `B7`/`B12` explicitly non-deferrable.

**That valve is removed in its entirety.** Review established the valve was itself the
defect: it authorized marking `024.001-T` **`Done`** while clause sites named in that
task's own completion acceptance criteria were unimplemented. A completion criterion the
completion path is pre-authorized to skip is not a criterion. The five "deferrable"
clauses are all statements about which close paths exist or are permitted; leaving any of
them describing a two-verdict world after a third verdict ships leaves the contract
**incorrect**, not merely uneven — the same defect class rev 1 already recognised for
`B7` and `B12`. The line between the two groups did not survive review.

**The rule now:** every MUST and CONSISTENCY site in plan C §3 is implemented. An overrun
is **HALT and return to Stage** — never a partial `Done`, never a deferral. C's budget is
re-stated honestly at **~105 min / ~15 min margin** (plan C §12), with the margin
acknowledged as still thinner than `B18`'s own estimating variance and **no** slack
mechanism to hide that. See plan C §12.1.

### 7.4 Cross-surface coupling matrix

An earlier draft claimed **CO-1 (the token string) is the only cross-shipment coupling**.
Review established that is **wrong** — it is the only coupling requiring a *literal string
match*, but not the only one. The full set, each with a producer, consumer and guard:

| # | Coupling | Producer | Consumer | Guard |
|---|---|---|---|---|
| **CO-1** | Capability token `task-only-finalize/v1` | C (skill frontmatter) | B (P-015 gate) | **C's** exact-parity test row 2. *(B cannot assert parity — B merges before C exists; B's row 3 asserts only that A7 states the token requirement with an exact version.)* |
| **CO-2** | Verdict literal `TASK_ONLY_FINALIZE` | B (authorizes) | C (emits), A (invokes) | B row 1, C rows 3–4 |
| **CO-3** | Classifier order `CASCADE → TASK_ONLY_FINALIZE → SAFE_CLOSE/HALT` | B | C (selector) | B row 2, C row 3 |
| **CO-4** | Topology `T1–T5` ↔ guards `G1–G5` | B (asserts) | C (implements) | C row 12 (cite-not-restate); **semantic**, not textual |
| **CO-5** | Zero-feature-member mutual exclusion | B | C | B row 6 |
| **CO-6** | Verdict-delegation contract (Ship invokes only what the verdict names) | A | B, C | A rows 4, 9 |
| **CO-7** | Reload set must include P-015 + skill | A (C1) | C (activation visibility) | A row 5 |
| **CO-8** | Feature-completion authority (a1) | A | B (needs it to close) | `022-S depends_on 021-S` |
| **CO-9** | `P1–P10` ordering vs P-007 archive restoration | C | C | C row 6 |
| **CO-10** | Engine CLI version + digest | C | C | §5 HALT |
| **CO-11** | **(rev 2 — re-scoped)** P-001 single-in-flight vs a live covering feature left by a task-only close | engine claim semantics | `017-S` routing | **Mechanical**: C's manifest re-shape puts `024-F` inside the manifest so the cascade archives it (Probe 24). **Not** prose disposition — Probe 23 proved no feature-endpoint dependency edge is constructible |
| **CO-12** | **(rev 2 — new)** Post-archive `--phase lifecycle` topology gate fails closed | Probe 20 | A, B, C closure sequences | `a0` runs **while the shipment is still active**; **no** lifecycle invocation after the archive. A AC-19, B AC-16, C AC-21 |
| **CO-13** | **(rev 2 — new)** Installed Ship lifecycle order (4.5 before 5; 6.0 branch before 6.1(e); closure artifacts + P-020 before the closure PR push) | `_ship.agent.md` | A, B, C closure sequences | A AC-18, B AC-15, C AC-22 |
| **CO-14** | **(rev 2 — new)** Token-activation session boundary: a session must not select a verdict whose authorizing token it merged | C (`B18`) | C's own closure | C row 13, C AC-14; **distinct from** A's Role-Boundary-scoped §6 carve-out, which is deliberately **not** widened |

**CO-4 is semantic, not textual, and that is the honest characterization.** `G1–G5`
necessarily *implement* `T1–T5`. C cites rather than restates them (row 12), which prevents
textual duplication but does **not** eliminate the semantic dependency. Drift between them
is a **real residual**, accepted: a CI coupling gate is out of scope for this chain.

**CO-8 is why the ordering is load-bearing**, not merely tidy: B cannot close before A is
live.

**CO-11 changed kind in rev 2, and that is the point.** Rev 1 discharged it with a
Stage-owned disposition step — a **prose** gate. Probe 23 established that the mechanical
alternative everyone would reach for (a `blocks` edge from `017-S` to `024-F`) is
**refused by the engine**: *"both endpoints must be shipments"*. With no feature-endpoint
edge available and no shipment `blocked` status, there was **no valid barrier to build**.
Rev 2 therefore removes the hazard rather than gating it: `024-F` joins C's manifest and
the engine's own cascade archives it (Probe 24 — `archived_ids` contains the feature,
**0** other active top-level units remain, successor eligible with no further action).
The residual case — `017-S`'s own `018-F` — is recorded rather than fixed **because
`017-S` is terminal and has no downstream consumer to block**.

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

### 8.1 Rev 2 amendment — C's manifest shape, forced by measurement

The decision above is **retained unchanged** in its structure: three shipments, the same
IDs (`021-S`, `022-S`, `023-S`), the same dependency chain, the same inventory split, the
same one-feature-one-task shape. **One thing changed**: C's manifest.

| | Rev 1 | Rev 2 |
|---|---|---|
| `023-S` manifest | task-only `[024.001-T]` | fully-covered root `[024.001-T, 024-F]` |
| C's close path | the **new** `TASK_ONLY_FINALIZE` (self-proof) | the **existing** `CASCADE` |
| `024-F` afterwards | live `active`, disposed by a Stage follow-up | **archived by the cascade** |
| First live use of `TASK_ONLY_FINALIZE` | C itself | **`017-S`** |
| Barrier before `017-S` | Stage disposition step (prose) | **none needed** |

**Why.** Rev 1 assumed a task-only close would leave `024-F` in a state that a Stage
follow-up could tidy before `017-S` routed. Three executed probes falsified the assumption
and then falsified the remedy:

* **Probe 22** — claiming a task-only shipment moves the **out-of-manifest** covering
  feature to `active`, and it **stays `active`** after the close. `024-F` would be a live
  top-level release unit at `017-S`'s P-001 gate (which tests for status `Active`).
* **Probe 23** — the obvious mechanical barrier, `017-S depends_on 024-F`, is **refused by
  the engine**: *"both endpoints must be shipments"*. No feature-endpoint dependency edge
  can be written at all, and backlogit defines no shipment `blocked` status. **There was
  no valid barrier available to build.**
* **Probe 24** — with the feature **inside** the manifest, the cascade archives it
  (`archived_ids` contains the feature), leaving **zero** other active top-level units and
  the successor eligible with no barrier and no follow-up. The control arm confirms the
  difference is attributable to manifest shape alone.

**What this costs, stated plainly.** C no longer proves the new path by closing with it.
That was rev 1's strongest property and it is genuinely weakened. It is **relocated, not
deleted**: plan C §8.4's Probe 21 post-half exercises the classifier against the **exact
`017-S` 12-member fixture** — a closer analogue of the real consumer than C's own one-task
manifest ever was — and plan C carries the un-hedged residual (**R-8**) that no *live*
task-only closure occurs before `017-S` depends on one.

**Why this was preferred to returning `MUST_REPLAN`.** The alternative to re-shaping is a
chain that closes C successfully and then **deadlocks**: `017-S` claimed, P-001 tripped by
a feature nothing is authorized to dispose, and no constructible edge to prevent it. Given
a measured choice between a weakened proof and a measured deadlock, the weakened proof
wins — and the weakening is recorded here rather than absorbed silently.

**Preserved.** Shipment IDs, the dependency chain `021-S → 022-S → 023-S → 017-S`, the
28-site inventory split (6 + 6 + 16), and the one-covering-feature-one-task invariant.

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
2. **C leaves `024-F` live after task-only closure.** — **RESOLVED IN REV 2; the step this
   question mandated has been removed.** T4 does require the root parent live and outside
   the manifest, and review correctly escalated that a live `024-F` can trip Ship Step 1's
   P-001 gate (line 308) and stall `017-S`. Rev 1's answer was an ordered Stage-owned
   disposition step. **Probe 23 falsified that answer**: no dependency edge can be written
   against a feature endpoint, so the step could never be mechanically enforced — it was
   prose. **Probe 24 replaced it**: `024-F` is now a member of C's manifest and the
   engine's own cascade archives it, leaving zero other active top-level units. Plan C §8.3
   is now titled *"No `024-F` disposition step is required"* and **AC-17** asserts the
   mechanical outcome. See §7.4 CO-11 and §8.1.
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
| 2 | C leaves `024-F` live, which can trip Ship's **P-001** single-in-flight gate (line 308) and stall `017-S` | **P1** | **Superseded in rev 2.** Rev 1's answer was an ordered Stage-owned disposition step; Probe 23 showed it could never be mechanically enforced. **§8.3 now removes the hazard**: `024-F` is a manifest member and the cascade archives it (Probe 24, C AC-17) |
| 3 | Rollback described as independently revertible; reverting B with C installed leaves an **unauthorized** implemented verdict | **P1** | New **§10.1 rollback matrix**; `C → B → A` reverse order **mandatory** after activation |
| 4 | Skill references `src/autoharness/gates/shipment_closure.py`, which **does not exist** here; "0 source changes" was an unchecked assumption | **P1** | Verified absent; **§2.2** records the skill markdown as the execution boundary; stale refs reconciled inside existing `B9`/`B17` (C AC-16) |
| 5 | C's overflow valve mis-targeted, ordering-incoherent, and over-broad (`B7`/`B12` are not branch-local) | **P2** | **Superseded in rev 2.** Rev 1 rewrote the valve (CONSISTENCY last, 5-clause subset, `B7`/`B12` non-deferrable). **§12.1 now removes the valve entirely** — every site is implemented; an overrun is HALT to Stage, never a partial `Done` (C AC-18) |
| 6 | C1's carve-out and the retained *"close under the just-merged contract"* left unreconciled and untested | **P2** | **§4.2** pins the reconciled wording; new Go test row 11 guards it (A AC-14) |
| 7 | Manifest re-shape to `[022-F, 022.001-T]` had no named owner or AC | **P2** | **§4.3** names Stage as owner, pre-claim; A AC-12 |
| 8 | `PO-1` ran at closure — the most expensive moment to discover a claim-shape mismatch | **P2** | Split into **PO-1a** (pre-execution) + **PO-1b** (at gate); A AC-10 |
| 9 | "CO-1 is the only cross-shipment coupling" — false | **P2** | Replaced by the **§7.4 coupling matrix** (CO-1…**CO-14** as of rev 2); CO-4 named as semantic and its drift accepted as residual |
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
* **C's ~15 min margin** — real, bounded to one shipment, with a HALT path and **no**
  overflow valve (rev 2; the rev-1 figure was ~13.5 min against a ~106.5 min budget).
* **(rev 2)** `TASK_ONLY_FINALIZE` is not exercised by a **live** closure before `017-S`
  depends on it — plan C **R-8**, mitigated but not eliminated by Probe 21's post-half.
* **(rev 2)** `017-S`'s own task-only close will leave `018-F` `active` — plan C **R-10**,
  accepted and **not** gated, because `017-S` is terminal and has no downstream consumer.
* **A AC-9 tamper-evidence** — the checkpoint is evidence, not enforcement.

## 12. What this deliberation does not change

* The `MUST_REPLAN` verdict on rev 6 — adopted, not overturned.
* The probe evidence base (Probes 14–20) — retained; the rev-6 plan becomes the
  **umbrella evidence and decision source** that A, B and C cite.
* Live-config evidence and recovery semantics from rev 6 — retained verbatim for C.
* The PR #54 actor boundary — the **operator** remains the sole push/merge actor. No
  agent push. Unchanged from rev 6 §11 / §14.
* `017-S` — remains `queued` and unclaimed throughout.
