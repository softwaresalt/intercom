---
title: "Plan — Ship covering-feature completion and close-path delegation (Foundation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 4
feature: 022-F
task: 022.001-T
shipment: 021-S
deliberation: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
superseded_deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Plan — Ship covering-feature completion and close-path delegation (Foundation)

> **Requires plan hardening**: **yes — applied in rev 2, re-affirmed in rev 3** (see
> `## Plan Hardening`).

> **Rev 3 — A is now the SOLE prerequisite.**
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (`SIMPLIFICATION_VALID`) retires the three-shipment chain for the current scope.
> Shipments **B** (`022-S`) and **C** (`023-S`) are **deferred generalized platform
> work** — `blocked`, records archived non-destructively, **not** current prerequisites.
> **A's own scope, surface, inventory, budget and closure sequence are UNCHANGED.**
> What changed is only what comes *after* A: the route now goes **A → `017-S`**, and
> `017-S` is a **feature-present** `CASCADE` case, not the zero-feature example rev 2
> used.

> **Rev 4 — A-only final-review remediation cycle 1. Five substantive corrections.**
> Rev 3 passed internal review but the A-only final review found five blockers. All are
> fixed here, and **none of them widens A's surface beyond its two files**.
>
> 1. **Inventory reopened from 6 to 8 clause sites (§4).** Rev 3 deliberately scoped the
>    stale queue-only pre-mode summary at L776–777 *out* of the inventory to protect a
>    pinned budget. That was the wrong trade: the summary is **operative instruction text
>    Ship reads at the gate**, it is **literally false on `017-S`'s route**, and a *second*
>    instance of the same defect exists at **L255–256** (Step 0.5 item 6) which rev 3 never
>    detected at all. Both are now **C7** and **C8**. The budget is **recalculated, not
>    preserved**: ~78 → **~92 min**, margin ~42 → **~28 min**. Test rows 13 → **16**.
> 2. **Step 6.1(a1)'s selector is now exhaustive over member multiplicity (§5.1).** Rev 3's
>    three cases silently assumed *at most one* feature member. A **two-feature manifest
>    reached an unspecified branch.** Case **(iv)** now fails closed with **no** feature
>    mutation.
> 3. **S2's checkpoint recovery now cites the FULL installed protocol (§9.1, AC-9).** Rev 3
>    listed 8 steps and **omitted the `restore → prune/gate → resume` ordering, the Engram
>    reachability gate, and the never-prune allowlist** — all three of which are installed,
>    mandatory, and fail-closed in this workspace.
> 4. **The root-included 13-member fixture is now a NON-BYPASSABLE A closure gate (§9.4,
>    AC-20).** Rev 3 recorded it as a follow-up "before `017-S`, not before A". That left
>    the only unproven step in the chain guarded by nothing. It is now a **gate on A's own
>    close**.
> 5. **Formal `plan-review` gate executed (`## Plan Review`).** Rev 3 carried an internal
>    prose review only. The real gate ran as **attempt 1**, dispatching all seven selected
>    personas (including the `gpt-5.6-sol` anchor route) with machine-readable
>    `dispatch_mode:` / `decision:` markers. **It returned `decision: FAIL`.** Rev 4 is
>    therefore **NOT harvest-ready** and `021-S` must not be claimed; see `## Plan Review`
>    for the full finding set and the remediation queue.

- **Shipment**: `021-S` (A, Foundation) — **the sole prerequisite**; the only eligible shipment
- **Closes by**: the **existing** P-015 `VERIFIED FULLY-COVERED-ROOT EXCEPTION` (`CASCADE`)
- **Authorizes**: nothing new in the close-path verdict set
- **Releases**: `017-S` (13-member fully-covered root), which then closes by the **same existing** `CASCADE`

## 1. Objective

Make the fully-covered-root manifest shape **reachable** by Ship, and stop Ship from
restating a classification it should delegate.

Three deliveries, one surface plus one test:

1. **Narrow Role Boundary authority** for Ship to complete a covering feature
   `active -> done`, fail-closed.
2. **An operative post-task / pre-reconcile parent-completion step** that exercises it.
3. **Replace the duplicated close-path exception wording** in `_ship.agent.md` with a
   generic delegation to the authoritative P-015 + `shipment-reconcile` machine-selected
   verdict — **without authorizing any new verdict**.

Current behavior after this lands remains exactly `CASCADE` or `SAFE_CLOSE`/`HALT`.

## 2. Why this is the sole prerequisite

Two independent reasons, both verified at `e7e981d`:

**2.1 Pre-mode makes the shape unreachable today.** `_ship.agent.md` Step 6.1(a) invokes
`shipment-reconcile` with `expected_status: done` and verifies *"every manifest item is
present in queue with `status: done`"*. A manifest `[022-F, 022.001-T]` therefore needs
the **feature** `done`. Ship's Role Boundary permits only *"move **tasks** to
active/done"* (line 38). No feature-completion authority exists, so no fully-covered-root
manifest can ever reach closure.

**`017-S` has exactly this shape**, now that its root `018-F` has been restored to the
manifest (13 members). So `017-S` cannot close until A is live — which is precisely why A
is the sole prerequisite, and why nothing else is needed between them.

**2.2 The duplicate has already drifted unsafe.** This is a pre-existing defect found
during re-planning, independent of the bootstrap:

| Surface | Coverage rule |
|---|---|
| P-015 item 1 (v1.24.0), `workflow-policies.md:438` | *"every one of its **descendants — at every depth, not only direct children**"*, and: *"A check limited to direct children is insufficient"* |
| `_ship.agent.md:810` | *"every one of its **children**, enumerated live from `.backlogit/queue/` + `.backlogit/archive/`"* |

A manifest `[feature, task]` whose task owns an out-of-manifest subtask **passes Ship's
copy and fails the policy** — precisely the corruption v1.24.0 was written to prevent.
Delegation removes the drift **and** the class of defect.

## 3. Surface — exactly 2 files

| # | File | Change |
|---|---|---|
| 1 | `.github/agents/_ship.agent.md` | Role Boundary row; new Step 6.1(a1); C1–C6 |
| 2 | `tests/integration/ship_feature_completion_contract_test.go` | one table-driven test fn + ≤4 helpers |

**New production code files: 0.** No policy file, no skill file, no source change.

## 4. Exact inventory — 8 clause sites (all MUST) + 2 additive

Line anchors **re-verified against the installed file at `d19b954`**. C1–C6 match rev 6
Inventory Table §6.1-C exactly. **C7 and C8 are new in rev 4** and are *not* in rev 6's
table — C8 was recorded there but deliberately excluded, and C7 was never detected.

> **Label order is allocation order, not line order.** Sorted by line the sites run
> **C7 (255) → C1 (756) → C8 (776) → C2 (789) → C3 (794) → C4 (802) → C5 (821)**. C1–C6
> keep their rev-6 labels so cross-revision traceability survives; the two new sites take
> the next free labels. An implementer MUST apply by **anchor text**, never by label order
> and never by line ordinal.

| # | Line | Clause | Change |
|---|---|---|---|
| **C1** | 756–764 | Mandatory pre-self-close context reload — names only Ship instructions + `shipment-reconcile` | Also reload **P-015**; positively verify merged tokens **and the merge commit**; add the §6 authority-escalation carve-out with the **pinned reconciled wording** of §4.2 |
| **C2** | 789–791 | safe-close summary prescribing `backlogit move <shipment_id> --status shipped` | Correct: the engine **refuses** this for shipments (exit 9, Probe 18). Replace with a pointer to the skill's authoritative step 8 |
| **C3** | 794–795 | *"**Do NOT call** `backlogit shipment ship` … **unless** the P-015 VERIFIED FULLY-COVERED-ROOT EXCEPTION applies"* | Replace the named single exception with: do not call it unless **the machine-selected verdict** authorizes it |
| **C4** | 802–823 | the re-derived classification prose (binary selector + drifted `children` rule) | **Delete the restatement.** Replace with generic delegation |
| **C5** | 821–823 | *"invoke the cascade … **in place of** the safe-close sequence above"* | Subsumed by C4's delegation block; single-exception framing removed |
| **C6** | new bullet | — | **NEW** delegation bullet: Ship invokes whichever close path the verdict names, and **no** path absent from the verdict |
| **C7** | 255–256 | **(rev 4)** Step 0.5 item 6 intake pre-mode summary: *"verifies every manifest item is present in `.backlogit/queue/` with the expected status"* — **queue-only** | Reconcile to the authoritative skill semantics (§4.4): an item is accepted when it is a **queue item at the expected status** **OR** an **archive-located `pre-archived` member accepted without a status check** |
| **C8** | 776–777 | **(rev 4)** Step 6.1(a) pre-archive gate summary: *"verifies that every manifest item is present in queue with `status: done`"* — **queue-only** | Same reconciliation as C7, with `expected_status: done` |

| Additive | Content |
|---|---|
| **ADD-1** | Role Boundary table row (line 38) — narrow feature-completion authority |
| **ADD-2** | New **Step 6.1(a1)** — covering-feature completion gate |

### 4.0 Why C7 and C8 were reopened (rev 4 — reverses a rev-3 decision)

Rev 3 recorded L776–777 as *"deliberately OUT OF A's INVENTORY"* on two stated grounds —
budget, and *"not load-bearing … descriptive drift in a summary, not an operative gate"* —
and instructed that *"a reviewer MUST NOT fail A for its absence."* **Rev 4 withdraws that
instruction and both grounds.**

1. **It is operative, not descriptive.** `_ship.agent.md` is the agent's **instruction
   surface**. A sentence telling Ship what a gate verifies *is* what Ship acts on when it
   reads the step. The rev-3 argument — that the skill computes the right answer anyway —
   proves the *runtime* is safe; it does not make a false instruction harmless. An agent
   reading *"every manifest item is present in queue with `status: done`"* against
   `017-S`, where **11 of 13 members are not in queue at all**, has every reason to
   conclude the manifest is broken and halt, or to "help" by restoring archived members to
   the queue. That is a **fabricated** failure on the exact route A exists to enable.
2. **The same defect exists twice, and rev 3 found only one instance.** C7 (L255–256) is
   the intake-side twin. Rev 3's inventory was therefore not merely under-budgeted, it was
   **incomplete** — which is precisely the §2.2 duplication-drift failure mode this plan
   was written to eliminate, reproduced inside the plan's own inventory.
3. **Budget is an output, not a constraint.** Pinning an inventory to protect a number
   inverts the relationship. The correct response is to **recalculate** (§12): ~92 min,
   margin ~28 min, still inside the 2-hour rule with room. Had it not fit, the correct
   response would have been HALT-and-return-to-Stage (§10.3), never silent exclusion.

### 4.4 C7/C8 reconciled pre-mode wording (normative)

Both sites MUST be reconciled to the **authoritative** `shipment-reconcile` pre-mode
classification, which is (skill `## Classification` table, and its `PROCEED` rule):

| Classification | Condition |
|---|---|
| `matched` | Queue file present **AND** declared status matches `expected_status` |
| `pre-archived` | **No queue file, but an archive file exists** — already archived before this shipment ran; **valid**, accepted **without** a status check |
| `missing` | No queue **or** archive file for this manifest item |
| `status-mismatch` | Queue file present but declared status ≠ `expected_status` |

`PROCEED` requires every item to be `matched` **or** `pre-archived`, with no orphans. The
replacement sentences MUST therefore express **both** accepted dispositions. Neither
sentence may be left in the queue-only form. Asserted by Go test rows **14** and **15**.

> **Scope guard.** C7 and C8 correct a **summary of an existing gate**. They change no
> threshold, add no classification, and grant no authority — the skill's semantics are
> already what they are. This is A's `_ship.agent.md` file, so both sites are inside A's
> declared 2-file surface; `shipment-reconcile/SKILL.md` is **not** touched.

> **Rev 2 — additive-item label rename, recorded for traceability.** Rev 1 labelled these
> two additive items `A-1` and `A-2`. That collided visually with plan **B**'s clause-site
> labels `A1`–`A8` (B's sites are named `A*` because rev 6's Inventory Table §6.1-A covers
> `workflow-policies.md`). Because the three plans are reviewed together, `A-1` and `A1`
> appearing in the same review were a genuine mis-read hazard. They are renamed
> **`ADD-1`/`ADD-2`** here. **Traceability:** rev 1 `A-1` ≡ rev 2 `ADD-1`; rev 1 `A-2` ≡
> rev 2 `ADD-2`. No content changed — label only. The cross-plan label map is:
>
> | Plan | Surface | Clause-site labels | Origin |
> |---|---|---|---|
> | **A** (this plan) | `_ship.agent.md` | `C1`–`C6` + `ADD-1`/`ADD-2` | rev 6 Inventory Table §6.1-C |
> | **B** | `workflow-policies.md` | `A1`–`A8` | rev 6 Inventory Table §6.1-A |
> | **C** | `shipment-reconcile/SKILL.md` | `B1`–`B18` | rev 6 Inventory Table §6.1-B |
>
> The letters track the **rev-6 inventory section**, not the shipment. That is confusing
> but it is the traceable naming, so it is documented rather than silently re-lettered.

### 4.1 Delegation wording constraint (normative)

The replacement text MUST NOT enumerate the authorized verdicts. It delegates:

> Ship MUST invoke the close path named by the machine-checkable classification defined
> in **P-015** and implemented by the **`shipment-reconcile`** skill, and MUST NOT invoke
> any close path the classification did not name. Ship does not re-derive, restate,
> extend or narrow that classification.

This is what lets **future** close-path work land **without touching this file again** —
including the deferred `TASK_ONLY_FINALIZE` platform work, if a fired trigger ever brings
it back. It also forbids Ship
from authorizing a verdict on its own — A adds **no** verdict.

### 4.2 Pinned reconciled C1 wording (normative)

The existing clause says *"Close under the just-merged contract, not a stale in-context
copy."* The §6 carve-out says a reload MUST NOT widen the authority envelope. Left
side-by-side and unreconciled, these two sentences are in direct tension **for A's own
merge**, which changes both procedure and authority. The implementer MUST NOT leave the
reconciliation to inference. C1's reconciled text is pinned:

> Close under the just-merged contract, not a stale in-context copy — **with one
> exception: if the merged change alters this agent's own Role Boundary, the altered
> authority does NOT apply to the session performing the merge.** That session MUST write
> a checkpoint and end; a fresh session loads the new Role Boundary and completes closure.
> This carve-out is scoped to **Role Boundary** changes only; every other merged change —
> procedure, algorithm, gate ordering — takes effect on reload exactly as before.

Go test row 11 asserts the scoping clause is present, so the two sentences cannot ship
unreconciled.

### 4.3 Stage precondition — manifest re-shape (owned, and now COMPLETED)

**Rev 2 status: DONE.** `021-S`'s live manifest is the fully-covered root
`[022-F, 022.001-T]`, verified at this plan's revision. Rev 1 described the re-shape in the
present/future tense (*"`021-S`'s live manifest is task-only `[022.001-T]` … is a required
re-shape"*), which was accurate when drafted but became stale the moment the re-shape was
performed in the same re-plan session. An implementer reading rev 1 would have gone looking
for work that no longer existed.

The re-shape was performed by **Stage** (Ship's Role Boundary forbids editing shipment
planning fields), **before** Ship claims `021-S`.

If it had been skipped, A's manifest would have stayed task-only, the classifier would
return `SAFE_CLOSE` (no feature member ⇒ no `CASCADE`; `TASK_ONLY_FINALIZE` not yet
authorized), and A could not close. That outcome is **fail-closed** (§10.3). **AC-12** is
now a *verification* that the shape is present at claim, not an instruction to create it.

## 5. The Role Boundary grant (normative, narrow, fail-closed)

Authorized transition: a **covering feature** `active -> done`, and **only** when **all**
hold:

1. every **non-pre-archived executable descendant** (at any depth) is `done`;
2. **no** `queued` or `active` live descendant remains;
3. parent/child topology is intact — every descendant resolves to this feature by
   `parent_id`, and no descendant is orphaned or reparented;
4. it is **immediately before** `shipment-reconcile mode: pre` for the shipment whose
   manifest contains that feature;
5. the feature's live status is exactly `active`.

**Any** condition unmet → **HALT**, fail closed. No inference, no fallback, no partial
grant. The authority is **not** general feature lifecycle management: it may not move a
feature to `done` outside Step 6.1(a1), and it grants nothing over features outside the
active shipment's manifest.

### 5.1 Step 6.1(a1) applicability — four outcomes, exhaustive over member multiplicity (rev 2; re-scoped rev 3; **made total in rev 4**)

Rev 1 specified what a1 does when it applies, but never said what a1 does when the
manifest has **no** feature member. An unspecified step invites an implementer to either
skip the gate silently or invent a transition. Both are wrong.

**Rev 3's three cases were still not total.** They branched on *"the manifest contains a
feature member"* — a predicate that is equally true of a manifest with **one** feature
member and a manifest with **three**. Rev 3 therefore left the multi-feature manifest on an
**unspecified branch**, and the two readings an implementer could reach from that text are
both unsafe: complete *the first* feature member found (silently leaving the others
`active`, and making the outcome depend on manifest ordering), or complete *all* of them
(a bulk feature-lifecycle power §5 explicitly refuses to grant).

a1 MUST therefore **select on multiplicity first**, and the selector MUST be exhaustive
over `n = |{manifest members whose artifact_type is feature}|`:

| Case | Condition | Outcome |
|---|---|---|
| **(i) NOT APPLICABLE** | `n == 0` — the manifest contains **zero** feature members | **Explicit no-op.** a1 records `A1_NOT_APPLICABLE: no feature member in manifest` and proceeds directly to Step 6.1(a). This is a **successful** outcome, not a skip and not a failure |
| **(ii) APPLIES — proceed** | `n == 1` **and all five** §5 conditions hold for that one member | Perform `active -> done` on **that** member; record the transition. (If its live status is already `done`, §5.2 idempotent resume applies) |
| **(iii) APPLIES — halt** | `n == 1` **and any** §5 condition is unmet | **HALT, fail closed.** Surface which condition failed. Do **not** proceed to Step 6.1(a). Do **not** widen any condition in-flight |
| **(iv) AMBIGUOUS — halt** | `n > 1` — the manifest contains **more than one** feature member | **HALT, fail closed**, with `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`. **Evaluate no §5 condition and mutate no feature.** Do not complete any member, not even one that would individually qualify. Return to **Stage** for manifest disposition |

**`n` is computed by `artifact_type`, never by ID suffix.** Subtask IDs end `-ST`, whose
last character is also `T`; feature IDs end `-F`. A suffix test is the same class of defect
§R16.3.3 of the policy-gap plan records for the executable-set filter. Read
`artifact_type` from each manifest member's record.

> **Rev 3 — `017-S` is case (ii), NOT case (i).** Rev 2 cited `017-S` as the live example
> of case (i), on the basis that it was a **12-member task-only** manifest. That is no
> longer true. Per
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`, `017-S`'s
> root `018-F` has been **restored to the manifest**, giving **13** members and making it
> a **fully-covered root**. At `017-S`'s gate, a1 therefore takes **case (ii)**: `018-F`
> is a present feature member, the five §5 conditions hold, and a1 moves it
> `active -> done` — after which pre-mode passes and classification returns **`CASCADE`**.
>
> **Case (i) is retained regardless**, and is still asserted by Go test row 12. It is now
> a **defensive completeness** requirement rather than a case this chain is known to
> exercise: the step must be total over its input, and a future task-only manifest must
> not be able to reach an unspecified branch. Retaining it costs one recorded no-op and
> removes an entire class of ambiguity.

> **Rev 4 — both live manifests are `n == 1`; case (iv) is defensive.** `021-S` is
> `[022-F, 022.001-T]` ⇒ `n == 1`. `017-S` is the 13-member root ⇒ `n == 1` (`018-F`; the
> other 12 members are tasks and subtasks). **No current shipment reaches case (iv)** —
> it exists so that the step is **total**, exactly as case (i) does.

> **DISCLOSED NARROWING — case (iv) restricts a manifest shape P-015 explicitly authorizes
> (rev 4, plan-review finding A-5).** This is recorded honestly rather than claimed away.
>
> P-015's fully-covered-root exception is *"quantified over **every feature member of the
> manifest**"* (item 1) and item 4 speaks of *"the qualifying root feature **member(s)**"* —
> **plural is authorized by policy**. A manifest with two qualifying root features, each
> fully covered, with nothing extra, **satisfies P-015 today**.
>
> Case (iv)'s unconditional `n > 1` HALT therefore makes such a manifest **unclosable by
> Ship**, because a1 halts before pre-mode ever runs. That is a **real narrowing of
> reachability**, and it sits in tension with §4.1's constraint that Ship *"does not
> re-derive, restate, extend or **narrow** that classification."*
>
> **Both statements are kept, and the tension is resolved explicitly rather than by
> deletion:** §4.1 governs the **verdict selection** — Ship still may not re-derive or
> narrow *which close path the classifier names*. Case (iv) is an **upstream mutation
> guard**: it declines to exercise the §5 grant when the grant has no unique referent, and
> it halts **instead of** mutating. A halt that returns the shipment to Stage is not a
> re-classification.
>
> **This is nonetheless a narrowing, and the claim that A "narrows nothing" is
> WITHDRAWN.** The accepted cost: a legitimate multi-root manifest must be dispositioned by
> **Stage** (split into one shipment per covering feature, or re-planned) rather than closed
> by Ship. **Recorded follow-up**, not charged to A: make a1 set-based — validate *every*
> feature member against the same five §5 conditions under one lock and transition all of
> them — which would restore full P-015 reachability. Deliberately **not** done here: it
> widens the grant from one feature to N, which is a materially larger authority increase
> than A is scoped to make, and it has no current consumer.

**The distinction between (i) and (iii) is load-bearing.** Case (i) is "there is nothing
here for this gate to do". Case (iii) is "there is something here and it is not in the
required state". Collapsing them — treating an unmet guard on a real feature member as a
benign no-op — would let a shipment close with a covering feature left `active` or with
descendants unfinished, which is precisely the corruption the gate exists to prevent. A
zero-feature-member manifest must never be reported as a guard failure either: that would
make any task-only shipment un-closable at the gate.

**Case (iv) is fail-closed for a different reason than (iii).** (iii) halts because a known
member is in the wrong state. (iv) halts because **a1 cannot determine which member the
grant is about.** §5 grants completion of *"a covering feature"* — singular, and scoped to
"the active shipment's manifest". With two feature members there is no unique referent, so
**the grant does not apply at all** and a1 has no authority to mutate either one. Halting
without evaluating conditions is deliberate: evaluating them first would invite an
implementer to proceed when exactly one of the several happens to qualify, which is the
manifest-ordering-dependent behavior this case exists to forbid.

Asserted by **AC-15** (case i), **AC-16** (case iii) and **AC-21** (case iv).

### 5.2 Idempotent resume (rev 2)

A1 MUST be **idempotent**. If the feature member's live status is already `done` when a1
runs, a1 records `A1_ALREADY_DONE: {feature_id}` and proceeds to Step 6.1(a) **without
error and without re-issuing the transition**.

This is not hypothetical tidiness — it is required by A's own two-session closure. If S2
completes `022-F -> done` and then fails anywhere between a1 and the close (a
`RECONCILE_FAIL`, a digest mismatch, an operator halt), the recovery session re-enters at
a1 with the feature already `done`. Without idempotent resume, condition 5 (*"live status
is exactly `active`"*) would fail closed and A could **never** be closed by any session —
a permanently stuck shipment created by its own partial success.

Idempotent resume is scoped tightly: it recognises the **already-satisfied terminal state**
of the transition a1 itself performs. It does **not** relax conditions 1–4, which are
re-verified on every entry, and it does **not** authorize any transition from any status
other than `active`. A feature in `queued`, `blocked`, or any other status remains case
(iii) — **HALT**. Asserted by **AC-17**.

## 6. The authority-escalation rule (load-bearing)

A merged change to Ship's **Role Boundary** MUST NOT take effect mid-session.

* A reload MAY update the **procedure** a session follows.
* A reload MUST NOT widen the **authority envelope** of the session performing it.

Role Boundary is a **session-identity** constraint fixed at session start. Without this
rule, the existing pre-self-close reload would let the session that merged A immediately
use the authority A introduced — an agent granting itself authority by merging a change
to its own contract.

Therefore: **the session that merges A checkpoints and ends.**

## 7. Ordering (test-first; `main` stays green)

1. **H0 — red.** Add the single table-driven test function. Marker:
   `not implemented: ship feature-completion delegation contract`. Red count **1 of 1**;
   `go vet` clean; `go test ./...` non-zero.
2. **H1 — green.** Apply ADD-1, ADD-2, C1–C6. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function, so Ship Step 2's feature-scoped batch yields one
function and Step 4.3 cannot deadlock (the `021.001-T` constraint).

## 8. Verification obligations

### 8.1 What the Go test proves

**Contract text and ordering on `_ship.agent.md` only.** **16 rows**:

| # | Assertion | Kind |
|---|---|---|
| 1 | Role Boundary row contains the narrow feature-completion grant | positive |
| 2 | The grant names all five conditions of §5 | positive |
| 3 | Step 6.1(a1) exists and is positioned **after** `a0` and **before** `a` | **ordering** |
| 4 | The delegation block is present and names P-015 + `shipment-reconcile` | positive |
| 5 | The reload clause (C1) also names P-015 and the merge commit | positive |
| 6 | The authority-escalation rule (§6) is present | positive |
| 7 | `_ship.agent.md` no longer contains the drifted `children`-only coverage rule | **negative** |
| 8 | `_ship.agent.md` does **not** name `TASK_ONLY_FINALIZE` | **negative** |
| 9 | `_ship.agent.md` does **not** enumerate the authorized verdict set | **negative** |
| 10 | The `backlogit move <shipment_id> --status shipped` prescription (C2) is gone | **negative** |
| 11 | C1's reconciled wording scopes the carve-out to **Role Boundary** changes only (§4.2) | **ordering/consistency** |
| 12 | Step 6.1(a1) states the **zero-feature-member explicit no-op** (§5.1 case i) and that an unmet guard on a present feature member **halts** (case iii) | **ordering/consistency** |
| 13 | Step 6.1(a1) states **idempotent resume** — a feature member already `done` proceeds without error and without re-issuing the transition (§5.2) | **ordering/consistency** |
| 14 | **(rev 4, C7)** Step 0.5 item 6's pre-mode summary names the archive-located **`pre-archived`** disposition and is **not** queue-only | positive |
| 15 | **(rev 4, C8)** Step 6.1(a)'s pre-archive gate summary names the archive-located **`pre-archived`** disposition and is **not** queue-only | positive |
| 16 | **(rev 4)** Step 6.1(a1) states the **multi-feature-member fail-closed halt** — `n > 1` halts with no feature mutation (§5.1 case iv) | **ordering/consistency** |

Rows 7–10 are the **widening guards**: they fail if this change authorized anything new.
Row 11 is the **coherence guard**: it fails if C1 ships with the carve-out and the
"close under the just-merged contract" directive unreconciled.
**Rows 12–13 are the applicability guards** (rev 2): they fail if a1 ships without an
explicit not-applicable outcome — which would leave any task-only manifest ambiguous at
the gate — or without idempotent resume, which would make a partially-completed A
permanently unclosable.
**Rows 14–15 are the truthfulness guards** (rev 4): they fail if either pre-mode summary
ships still asserting that every manifest item lives in `.backlogit/queue/` — an assertion
that is false on `017-S`'s route for 11 of 13 members.
**Row 16 is the totality guard** (rev 4): it fails if a1's selector ships without a
specified `n > 1` branch.

> **Rev 3 note (superseded by rev 4).** Rev 3 recorded the row count as **unchanged at 13**
> and stated *"The Go test is **not** modified by rev 3."* That was true of rev 3. **Rev 4
> does modify the Go test**: three rows are added (14, 15, 16), taking the count to **16**.
> Rows 1–13 keep their numbering and their assertion text unchanged, so rev 2/rev 3
> traceability is intact. Row 12's rationale re-scoping (defensive totality rather than
> `017-S` being task-only) still stands.

**Helper budget.** Still **one** table-driven test function and **≤4 helper functions**
(§3). The three added rows are table entries, not new helpers: rows 14–15 reuse the same
substring-assertion helper as rows 1–2 and 4–6, and row 16 reuses the ordering/consistency
helper used by rows 11–13. **Helper count is unchanged.**

### 8.2 What it does **not** prove

It does **not** execute a shipment closure, does **not** prove the engine's cascade
behavior, and does **not** prove that shipment claim leaves a covering feature `active`.
That last point is §10 risk R-1 and is discharged by **probe obligation PO-1**, not by
the test. Claims stay honest.

## 9. Self-hosting closure sequence

`021-S` manifest becomes `[022-F, 022.001-T]` — a **fully-covered root**: `022-F` is a
root (no `parent_id`), its complete descendant set at every depth is exactly
`{022.001-T}` (verified: no `022.001.*-ST` exists), and the manifest contains nothing
else. It therefore qualifies under the **existing, already-reviewed** P-015 exception.

| Step | Session | Action |
|---|---|---|
| 0 | **Stage** | **Pre-execution readiness check (PO-1a)** — **DISCHARGED**, see §10.1. Probe 22 measured the claim shape; no pre-implementation blocker remains |
| 1 | S1 | Verify on `main`; Step 0.5 pre-claim topology gate; **claim `021-S`**. Ship moves **`022.001-T -> active`**. `022-F` is left **`active`** — **measured, not assumed** (Probe 22 ARM AB) |
| 2 | S1 | H0 red → implement → H1 green |
| 3 | S1 | **Step 4.3 quality gates; Step 4.4 review gate** |
| 4 | S1 | **Step 4.5 Complete Task — commit, then `022.001-T -> done`.** *(Installed order: Step 4.5 (L517) precedes Step 5 (L554). The task reaches `done` BEFORE the implementation PR merges.)* |
| 5 | S1 | **Step 5 PR lifecycle** — full gate sequence, `--phase lifecycle` topology gate, build, push, PR, operator approval, **merge** |
| 6 | S1 | **Step 6 Merge Confirmation Gate** — `gh pr view` state `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | **Reload merged `main`.** Detect that the merged change alters **Ship's own Role Boundary**. Per §6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (`schema_version: 1`, `agent: ship`, `phase: awaiting-fresh-session-parent-completion`, `resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA; domain data under `context`) and **END** |
| 9 | **S2** | **Fresh session.** Full checkpoint-recovery protocol (§9.1) — enumeration, anomaly gate, explicit owner selection, `agent: ship` ownership, operator confirmation, **restore → prune/gate → resume** (Engram-reachability fail-closed; never-prune allowlist honored), same cursor, resolve-after-confirmed-resume. Loads the new Role Boundary + Step 6.1(a1) from merged `main` |
| 10 | S2 | Verify the merge SHA is an ancestor of `origin/main`; validate the **same** active shipment `021-S` |
| 11 | S2 | **Step 6.0 Post-Merge Branch Protocol** — `git checkout main`, `git pull`, `git checkout -b post-merge/022-ship-feature-completion`. **Created BEFORE any post-merge backlog mutation**, because Step 6.1(e) commits `.backlogit/` and those commits must not land on `main` |
| 12 | S2 | **Step 6.1(a0)** `--phase lifecycle` topology gate — run **while `021-S` is still `active`**. *(Probe 20 tripwire: post-archive this gate fails closed on every route; it must never be re-run after the close.)* |
| **12a** | **S2** | **ROOT-INCLUDED CASCADE FIXTURE GATE (§9.4) — NON-BYPASSABLE.** Run and persist the exact 13-member root-included/pre-archived fixture (Probe 25). **PASS is required before any a1 mutation and before `021-S` may be marked or archived `shipped`.** FAIL, missing, stale or digest-mismatched evidence ⇒ **HALT**; `021-S` stays `active`/unshipped and `017-S → 021-S` stays unsatisfied |
| 13 | S2 | **Step 6.1(a1)** — §5.1 case **(ii)** (`n == 1`): `022-F` is a manifest member and all five §5 conditions hold → `022-F active -> done` |
| 14 | S2 | **Step 6.1(a)** pre-mode, `expected_status: done` — **both** members now `done` → `PROCEED` |
| 15 | S2 | Classification returns **`CASCADE`** (fully-covered root); **6.1(b)** close via the cascade op; **6.1(c)** P-007 archive-integrity verify; **6.1(d)** post-mode; **6.1(e)** commit `.backlogit/` **on the closure branch** |
| 16 | S2 | **Step 6 item 2** `operational-closure mode=post-merge` → `docs/closure/`; **P-020** compact-context finalizes the compaction status — **all on the closure branch, before the closure PR is pushed** |
| 17 | S2 | **Step 6 item 9** `backlogit sync`, then push the closure branch, closure PR, P-014 local review, operator approval, merge |
| 18 | S2 | **Step 6 item 10** return to `main`; `git pull` |
| 19 | — | Orchestrator routes **`017-S`** only (rev 3 — A is the sole prerequisite; there is no B) |

**Why the order in steps 4–5 and 11–17 is exactly this.** Read from the **installed**
`_ship.agent.md`, not inferred:

| Installed anchor | Consequence |
|---|---|
| Step 4.5 (L517) precedes Step 5 (L554) | task `-> done` **before** the implementation PR merges |
| Step 6.0 (L726) items 2–3 — branch from fresh `main`; *"All subsequent Step 6 work happens on this branch"* | closure branch exists **before** the 6.1(e) backlog commit |
| Step 6.1(a0) — *"the shipment-scoped check immediately preceding the safe-close mutation itself"* | lifecycle gate runs while `021-S` is still `active` |
| Step 6.1(a)→(b)→(c)→(d)→(e) | pre-mode → close → P-007 → post-mode → commit |
| Step 6 item 2 and item 8 (P-020) precede item 9 (`sync`) and the 6.0 item 4 push | closure artifacts + P-020 land on the closure branch **before** the closure PR is pushed |

Rev 1 placed `022.001-T -> done` **after** the merge, and placed operational closure + P-020
as a single late step without establishing the branch first. Both contradicted the
installed contract. Corrected above.

### 9.1 S2's checkpoint recovery is the FULL protocol, not a resume shortcut (rev 2; **completed in rev 4**)

Rev 1 said only that S2 *"restores the checkpoint"*. That is the step most likely to be
short-circuited, so it is specified in full.

> **Rev 4 — rev 3's list was itself incomplete.** Rev 3 enumerated eight steps and
> **omitted three mandatory, installed, fail-closed elements**: the
> `restore → prune/gate → resume` **ordering**, the **Engram reachability** gate, and the
> **never-prune allowlist**. All three are installed in this workspace
> (`agent-engram` is in `capability_packs`), so rev 3's list was not a safe abbreviation —
> it was a specification that, if followed literally and completely, would still **violate
> the installed protocol**. The authoritative sources are
> `.github/instructions/backlogit.instructions.md` §**Checkpoint-Recovery / Prune-on-Restore
> Protocol** and `_ship.agent.md`'s own **Crash-Resumption / Startup Recovery Protocol**.
> S2 MUST execute the installed protocol as written; the list below is a **pointer and a
> completeness checklist**, never a substitute for reading it.

S2 MUST, **in this order**:

1. **Enumerate** via `backlogit_list_checkpoints` with `consumer_id: "ship"` and **no**
   `status`/`agent` filter — a quarantined or schema-invalid record must not be silently
   excluded by the query itself.
2. **Run the anomaly gate FIRST** — any validation error, quarantine flag, or
   missing/malformed required field in **any** enumerated summary ⇒ **FAIL CLOSED** to
   operator handoff, evaluated **before** the zero-candidate check.
3. **Require explicit owner selection** — never auto-pick, **even when exactly one
   candidate is returned**. The operator selects one checkpoint by filename; ambiguity
   fails closed.
4. **Validate ownership** — CheckpointV1 `agent` MUST be exactly `ship`. A `stage`-owned
   checkpoint is never selectable here (P-001 role separation).
5. **Obtain operator confirmation before restore** — present `resume_hint` and recorded
   state. There is no automatic resume: schema V1 carries no heartbeat or lease, so elapsed
   time can never prove S1 dead.
6. **RESTORE** — load the selected checkpoint's `state_dump` via
   `backlogit_get_checkpoint`. **(rev 4)**
7. **PRUNE/GATE — before resuming, never after. (rev 4, mandatory here.)** The installed
   owner sequence is **`restore → prune/gate → resume`**, explicitly *"never restore →
   resume → prune"*. `agent-engram` **is installed and active** in this workspace, so this
   step is **not** the supported no-prune no-op available to backlogit-only installs — it
   is a required gate. S2 performs a bounded **read-select-summarize** over the restored
   `state_dump` and bound engram state, dropping **only** superseded action-observation
   history already synthesized into recorded state.
   * **Engram reachability is fail-closed. (rev 4)** If the bound engram substrate is
     **unreachable** at restore, S2 **FAILS CLOSED to operator handoff — NO prune and NO
     resume.** A bounded file-based prune degradation is explicitly **not** permitted. This
     is the installed degraded-fallback rule and is distinct from the static
     `agent-engram`-not-installed case, which does not apply here.
   * **Never-prune allowlist. (rev 4)** These three classes survive pruning **intact**,
     regardless of age or verbosity: **(a)** the active-shipment / active-task cursor — the
     single-active resumption point, here `021-S`; **(b)** the **unresolved-checkpoint
     pointer itself** — the checkpoint being resumed; **(c)** **recorded gate verdicts** —
     quality gates, review verdicts, CI/PR gate outcomes. Pruning any of these is a
     protocol violation, not an optimization.
8. **Resume the same cursor** — and only **after** step 7 completes. The **same** single
   active shipment `021-S`. No parallel resume, no new worktree (P-001/P-016).
9. **Resolve only after a confirmed successful resume** — `backlogit_resolve_checkpoint`
   for that **one** selected checkpoint only. No bulk sweep, never a `stage`-owned record.
10. **Fail closed with no fresh-start fallback** — an invalid, torn, or unreadable
    checkpoint halts to the operator. S2 MUST NOT discard it and begin fresh work.
11. **No `cleanup_checkpoints` against an unresolved record. (rev 4)** A fail-closed
    operator handoff performs no resolve and deliberately leaves the checkpoint `active`.
    That is **neither** resolution **nor** an explicit archival decision, so the record
    stays **excluded from `cleanup_checkpoints` eligibility regardless of age** until it is
    either resolved after a confirmed resume or explicitly archived/abandoned by a named
    operator decision.

### 9.2 What this proves

* **Checkpoint/resume across a session boundary** — steps 8→9, executed, not asserted.
* **No authority gained mid-session** — S1 never uses feature-completion authority;
  it is used first by S2, which did not perform the merge. This is the §6 property
  demonstrated by execution.
* **A closes by an existing path.** At step 15, P-015 still has exactly one exception.
  A introduces no verdict and closes by the pre-existing one.

### 9.3 What A releases — `017-S` (rev 3)

A's full post-merge closure (step 15) and its P-020 operational closure (step 16) are the
release condition for **`017-S`**, which is the **only** dependent. `017-S.dependencies`
is exactly `[021-S]`.

`017-S` then closes **by the same existing path A used** — no new verdict, no new gate:

| Step | Action |
|---|---|
| 1 | `017-S` becomes eligible once `021-S` is archived shipped **and** A's post-merge closure is complete (P-020 compaction status `done`/`degraded`, not `pending`) |
| 2 | Ship claims `017-S`. Manifest is the **13-member fully-covered root**: root `018-F` + `018.008-T` (live) + 11 pre-archived legacy descendants |
| 3 | Step 4.5 moves `018.008-T -> done` |
| 4 | **Step 6.1(a1)** — §5.1 **case (ii)**: `018-F` is a present feature member; conditions hold; `018-F active -> done` |
| 5 | **Step 6.1(a)** pre-mode `expected_status: done`: `018-F` `done`, `018.008-T` `done`, 11 legacy members **pre-archived** (accepted without a status check) → **`PROCEED`** |
| 6 | Classification returns **`CASCADE`** — the **existing** exception |

**Why the rev-18 `SAFE_CLOSE` exit-9 blocker is not reached.** Rev 18 recorded that a
`SAFE_CLOSE`-classified shipment halts at skill step 8 because
`backlogit move <shipment_id> --status shipped` is refused (exit 9, Probe 18), and argued
no manifest shape could avoid it *"because the only shape that reaches the permitted
cascade path is one containing `018-F`, which fails pre-mode first."* **Step 6.1(a1) is
exactly what repeals that clause** — with a1 live, a manifest containing `018-F` no longer
fails pre-mode first. `017-S` classifies `CASCADE`, never `SAFE_CLOSE`, so step 8 is never
reached. The underlying tool/contract conflict is **not fixed** here; it is simply **off
this route**, and remains a recorded follow-up.

> **Verification scope of this claim — stated honestly (rev 3, internal review P2;
> ESCALATED TO A GATE in rev 4).**
> The `CASCADE` result above is established at the **classification** level: the manifest
> shape provably satisfies P-015's fully-covered-root preconditions, and pre-mode provably
> accepts pre-archived members. It is **not** established at the **engine-execution**
> level *for this exact member shape*. A's own closure (§9) archives a clean root with
> **no** truly-archived members; Probe 15 measured the **old** root-*excluded* 12-member
> shape, which is the inverse case. **No committed probe or Go test exercises a
> root-*included* cascade whose siblings are genuinely `status: archived`.**
>
> The path is **fail-closed** at two independent gates — the skill's cascade step 2 HALTs
> on non-empty `returned_ids`, and step 3's two-set gate HALTs on any required/returned
> mismatch — so a mis-derivation cannot silently corrupt the backlog. The realistic
> downside is that `017-S` **HALTs** and falls back to `SAFE_CLOSE`, re-exposing the
> exit-9 blocker, returning `017-S` to **Stage** at no data cost.
>
> **Rev 4 — the rev-3 disposition is WITHDRAWN.** Rev 3 recorded this as *"not a current
> blocker … deliberately not charged to A"*, to be run *"before `017-S` is routed, **not**
> before A is."* That disposition is unsound **as a matter of sequencing**: A's entire
> justification for existing is that it makes `017-S`'s shape closable (§2.1, §9.3). If
> the root-included cascade does **not** behave as classified, A has shipped an
> irreversible contract change — a permanent Role Boundary authority grant — in service of
> a route that does not work, and the discovery happens only *after* A is closed and
> `017-S` is claimed. The measurement is **cheap, disposable, and available now**; the
> rev-3 ordering deferred it past the one commit that makes it expensive to act on.
> It is therefore promoted to a **non-bypassable gate on A's own close** (§9.4, **AC-20**).

> **Rev 4 — the `_ship.agent.md` pre-mode summaries are now IN A's inventory.**
> Rev 3 carried a block here scoping the L776–777 pre-mode summary *"deliberately OUT OF
> A's INVENTORY"* on budget and not-load-bearing grounds, and instructed that *"a reviewer
> MUST NOT fail A for its absence."* **That block is withdrawn in full, and so is that
> instruction.** Both stale summaries — L255–256 **and** L776–777 — are now clause sites
> **C7** and **C8**; see **§4.0** for why both rev-3 grounds fail, and **§4.4** for the
> reconciled wording. A reviewer **SHOULD** fail A if either site ships unreconciled; Go
> test rows **14–15** enforce it mechanically.

### 9.4 Root-included cascade fixture — NON-BYPASSABLE A closure gate (rev 4)

This section converts rev 3's deferred follow-up into a **hard gate on A's own closure**.
It creates **no new shipment, no new backlog item and no new dependency edge** — it is an
additional, mandatory step inside A's existing §9 closure sequence (step **12a**).

**Why this placement enforces `017-S`'s future gate without new topology.** `017-S`
depends on `021-S`. If A cannot close, `021-S` never reaches `shipped`, the edge
`017-S → 021-S` is never satisfied, and `017-S` is never eligible to be claimed. Making
the fixture a precondition of A's close therefore makes it, transitively, a precondition of
`017-S` ever running — using **only the dependency edge that already exists**.

**Gate definition.**

| Property | Value |
|---|---|
| **Who runs it** | The **fresh S2 session**, after it has loaded A's merged Role Boundary + Step 6.1(a1) from merged `main` (§9 steps 9–11) |
| **When** | §9 step **12a** — after the `a0` lifecycle gate, **before** the a1 mutation, and necessarily before any `shipped` mark or archive |
| **Fixture shape** | The **exact** `017-S` shape: a **13-member** manifest = **root feature included** + 12 descendants, of which **11 are genuinely archived** and 1 is a live task. Not an approximation, not the Probe-15 root-excluded inverse. **See the measured shape below — reproduce it byte-for-byte in kind, not merely "archived somehow".** |
| **Workspace** | A **disposable, isolated, live-config-seeded** backlogit workspace — **never** the live one. Same pattern as Probes 18/20/22/24 |
| **Engine binding** | In-script **digest gate**. Current live binary `C:\Tools\backlogit.exe`, backlogit **1.10.1**, SHA-256 **`1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`**. A digest mismatch is a **FAIL**, not a warning |
| **Evidence location** | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.ps1` **and** `…/probe25-root-included-cascade-fixture.txt` — committed, alongside probes 14–24 |
| **Authorization** | **CLI-only**, matching probes 18–24. No MCP equivalence is claimed |

#### 9.4.1 The MEASURED live shape the fixture must reproduce (rev 4 — verified at `d19b954`)

Stage traced all 13 live `017-S` members this session. The fixture is only faithful if it
reproduces **this** shape; a fixture that archives its members *some other way* measures a
different case and its PASS is worthless.

| Member class | Count | Location | `artifact_type` | `status` | `archived_status` | `parent_id` |
|---|---|---|---|---|---|---|
| Root feature `018-F` | 1 | `queue/` | `feature` | `queued` | — | **none (root)** |
| Live task `018.008-T` | 1 | `queue/` | `task` | `queued` | — | `018-F` |
| Legacy tasks `018.001-T`, `018.002-T`, `018.003-T` | 3 | `archive/` | `task` | **`archived`** | **`queued`** | `018-F` |
| Legacy subtasks (3 + 3 + 2) | 8 | `archive/` | `subtask` | **`archived`** | **`queued`** | their `018.00N-T` |

Verified: **13/13 present**, **0 missing**, **0 duplicates** (no id in both `queue/` and
`archive/`), `018-F` declares **no `parent_id`**, and every one of the 12 descendants resolves
to `018-F` by the `parent_id` chain at depth 1 or 2. The manifest contains **nothing extra**.
P-015 preconditions 1 and 4 therefore hold on the live records.

> **The load-bearing detail: `archived_status: queued`, NOT `done`.** All 11 pre-archived
> members are `status: archived` with `archived_status: **queued**`. A fixture built with
> `archived_status: done` would be a **different and easier** case, because it would satisfy
> the skill's `archived-completed(done)` role, which these records **do not**.
>
> **Why this does not block the route** (checked, not assumed): under the skill's per-task
> **ROLE** table these 11 records classify as the **`any-other-archived-status` anomaly**
> (*"an archive record whose status is NEITHER `done` NOR `archived`-with-`archived_status:
> done`"*), and that table's anomaly list says *"REPORT and HALT"*. **But that table is scoped
> to `mode: detect-mixed-role` only**, which the skill states is *"used **ONLY to DESCRIBE**
> the inconsistency in a report — **NEVER** to gate a mutation"* (013-DL Addendum G, Copilot
> PR #304 finding 1). **Pre-mode** uses the separate four-value classification
> (`matched`/`pre-archived`/`missing`/`status-mismatch`), under which an archive-located member
> is simply **`pre-archived`, valid, accepted without a status check** — so `017-S` reaches
> `PROCEED`. The two classifications are explicitly different scans; conflating them would
> produce a false halt.
>
> **Recorded residual**: if any future gate were to adopt the role table as a *gating* check,
> all 11 members would halt `017-S`. That is not the case today and is not charged to A, but
> the fixture MUST carry `archived_status: queued` so this exact condition is exercised rather
> than accidentally avoided.

**PASS criteria — all of them, conjunctive.** The fixture must record, as literal tokens:

1. `DIGEST_GATE=PASS`;
2. pre-mode with `expected_status: done` returns **`PROCEED`**, with the 11 archived
   members classified **`pre-archived`** and **no** `status-mismatch` and **no** `missing`;
3. classification returns **`CASCADE`** (not `SAFE_CLOSE`, not `HALT`);
4. the cascade close returns **`returned_ids: []`** — empty;
5. the skill's cascade **step-3 two-set gate** passes, evaluated as **two separately-labelled,
   independently-failing** conditions (P-015 amendment 1.21.0): **`archived_ids − allowed_ids`
   is EMPTY** (nothing unexpected was archived) **AND `required_ids − archived_ids` is EMPTY**
   (nothing required was left unarchived). Note that `archived_ids` is a **transition log, not
   a manifest echo** — a genuinely `status: archived` member has no transition to report and is
   correctly absent — which is precisely why full-set equality is the wrong test and why the
   1.19.0 equality claim was withdrawn;
6. every surviving record's **`parent_id` is preserved**;
7. the shipment record reaches **`shipped`** and archives with `archived_status: shipped`;
8. **nothing outside the manifest** is archived, moved or deleted.

**Failure semantics — fail closed, no bypass.**

* If the fixture **FAILS** on any criterion, or the evidence is **missing**, **stale**
  (not regenerated by this S2 session), or **digest-mismatched**, S2 **HALTs**. `021-S`
  remains **`active` and unshipped**, is **not** archived, and `017-S → 021-S` remains
  **unsatisfied**. `017-S` stays queued and ineligible.
* **Only a PASS permits A's own final close.** There is no operator override inside this
  plan, no `--force`, and no "proceed with a recorded residual" path. A residual was
  exactly rev 3's disposition and is what rev 4 withdraws.
* A HALT here returns A to **Stage** for disposition. It costs no data: the fixture
  workspace is disposable and the live backlog is untouched at step 12a because **no a1
  mutation has occurred yet** — which is precisely why the gate is placed before a1 rather
  than after.

**What a FAIL would mean.** That the root-included/pre-archived cascade does not behave as
§9.3 classifies it. In that event **`017-S` must be re-planned before A ships**, because A's
release value — making `017-S` closable — would be unproven. This is the discovery rev 3's
ordering would have deferred until after A was irreversible.

**What this gate does NOT claim.** It measures the **engine** on a faithful fixture; it is
not a proof about the live `017-S` records at the moment `017-S` is later claimed. The
live-record re-read obligations at `017-S`'s own gate are unchanged and still fail-closed —
the same relationship PO-1a/PO-1b have (§10.1).

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | `022-F` may be `queued`, not `active`, at the a1 gate, so §5 condition 5 fails closed and A cannot self-close | **(rev 2 — DISCHARGED by measurement.)** Probe 22 ARM AB measured `022-F`'s analogue at **`active`** after claim of a fully-covered-root manifest. PO-1a is closed; **PO-1b** remains as the at-the-gate re-read (§10.1). Fail-closed HALT is still correct if the live read disagrees |
| **R-2** | S1 might use the new authority anyway, defeating the proof | §6 is stated normatively **and** asserted by Go test row 6 |
| **R-3** | Delegation could be read as authorizing any verdict the skill invents | §4.1 binds Ship to verdicts the **classification** names; P-015 remains the authorization surface. Go test row 9 guards it |
| **R-4** | C2's correction touches the safe-close summary, which the deferred C plan would also edit in its own file | **No overlap, and now moot**: C2 is in `_ship.agent.md`; the skill's step 8 is C's file — disjoint surfaces. **Rev 3**: C is deferred and lands nothing, so there is no concurrent editor at all |
| **R-5** | Removing Ship's restatement loses reviewer-visible context | Accepted. The classification stays fully specified in P-015 and the skill; duplication is what caused the §2.2 drift |

### 10.1 PO-1 — covering-feature status at claim (**PO-1a discharged in rev 2**)

**PO-1a — before implementation begins (Stage / §9 step 0). STATUS: DISCHARGED.**
Rev 1 required Stage to determine what status shipment claim leaves a covering feature in,
**before** implementing, so a claim-shape mismatch could not be discovered only after A's
implementation merged. That measurement has now been executed:

> **Probe 22** (`probe22-parent-status-at-claim.ps1`, executed at `7bc0318`,
> `DIGEST_GATE=PASS`, live-config-seeded, CLI-only):
> **ARM AB** — fully-covered-root manifest `[001-F, 001.001-T]`, the exact shape of
> `021-S`:
> `ARM_AB_FEATURE_STATUS_PRE_CLAIM=queued`,
> **`ARM_AB_FEATURE_STATUS_AFTER_CLAIM=active`**,
> `ARM_AB_FEATURE_ACTIVE_AT_CLAIM=True`,
> `ARM_AB_FEATURE_DONE_MOVE_ACCEPTED=True`.

Claim leaves the covering feature **`active`**, which is exactly what §5 condition 5
requires, and the engine accepts the `active -> done` move a1 performs. **A may proceed to
implementation.** Neither of rev 1's contingency branches — amending the claim step, or
widening the authorized transition — is needed, and neither is authorized.

**PO-1b — at the gate (S2, before §9 step 13). RETAINED.** Read `022-F`'s live status with
**`backlogit get 022-F`** and record the `status` field.

> **Read by status, never by path (rev 2 correction).** An earlier draft said *"re-read
> `.backlogit/queue/022-F.md`"*. That is wrong and would have made the `done` branch below
> **unreachable**: the live `registry.yaml` routes `done|accepted|rejected|archived` to
> `archive/` and only `queued|active|blocked|review` to `queue/`, and `shipment-reconcile`
> independently classifies a live `status: done` record found in `queue/` as `conflicting`.
> A literal implementer reading the queue path for a completed feature gets file-not-found,
> falls into "anything else", and **HALTs** — producing exactly the permanently-stuck
> shipment §5.2 exists to prevent.

* `active` → proceed (case ii).
* `done` → proceed via **idempotent resume** (§5.2).
* anything else → **HALT** (case iii). Do not widen §5 condition 5 in-flight. Return to
  **Stage**.

PO-1b is retained even though PO-1a measured the general behavior, because a probe
measures the **engine**, not this specific live record at that specific moment. Widening an
authority envelope mid-execution remains forbidden; PO-1b keeps the gate fail-closed
regardless of what the probe found.

### 10.1.1 Probe 18 travels with A

C2's correction rests on the claim that the engine **refuses** `backlogit move
<shipment_id> --status shipped` for shipment artifacts (exit 9). That is **Probe 18**, and
it is A's justification, not C's. A cites
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe18-liveconfig-inbound-*`
and the compound note
`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`. Asserted by **AC-13**.

> **Rev 3 — evidence directory is RETAINED.** The evidence directory was nominally
> *"owned by C"*. C is deferred, but the directory is **not** deferred with it: A cites
> **Probe 18** here and **Probe 22** in §10.1, and the A-only decision relies on **Probe
> 24** for `017-S`'s 13-member shape. The directory stays in place and untouched.

### 10.2 Rollback

Two files on a feature branch. `git revert` the merge commit; Ship's closure behavior
returns to today's state (including, knowingly, the §2.2 drift). No backlog mutation is
performed by the implementation commit itself, so no backlog rollback is required.

**Rev 3 — the rollback chain is now A-only.** Rev 2 required a mandatory `C → B → A`
revert order because B and C would have layered on top of A's delegation. B and C are
**deferred and archived** and land on nothing, so **no downstream contract surface depends
on A's delegation**. A can be reverted **on its own**, with one obligation below.

**The one remaining obligation: hold `017-S` before reverting A.** Reverting A removes
Step 6.1(a1), which is the sole reason `017-S`'s 13-member fully-covered root can close.
A revert therefore changes the contract `017-S` was planned against. Before reverting A:

1. confirm `017-S` is **not** `active` (it must not be mid-execution);
2. explicitly **hold `017-S`** — a **Stage** action, recorded in the backlog, taken
   **before** the revert;
3. either restore `017-S` to its former task-only 12-member shape **or** re-plan it —
   do **not** re-route it on this plan once a1 is gone.

This is the general rule *"a revert changes the contract queued work was planned against"*
applied to the only dependent that now exists. It is stated here normatively; rev 2
delegated it to plan C §10.1, which is now **deferred** and must not be routed from.

### 10.3 Failure paths

| Failure | Response |
|---|---|
| H1 cannot be reached in budget | HALT, return to Stage. Never split — A is already minimal |
| a1: a **present** feature member fails any §5 condition 1–5 | **HALT** fail-closed (§5.1 case iii); surface which condition; no close |
| a1: manifest has **zero** feature members | **Explicit no-op** (§5.1 case i) — record `A1_NOT_APPLICABLE` and proceed to `a`. **Not** a failure |
| a1: feature member already `done` | **Idempotent resume** (§5.2) — record `A1_ALREADY_DONE` and proceed to `a`. **Not** a failure |
| a1: manifest has **more than one** feature member | **HALT** fail-closed (§5.1 case **iv**) — record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`; evaluate **no** §5 condition and mutate **no** feature; return to Stage |
| **(rev 4)** §9.4 fixture **FAILS** any PASS criterion | **HALT.** `021-S` stays `active`/unshipped and unarchived; `017-S → 021-S` stays unsatisfied. Return to Stage. No override, no residual-accepted path |
| **(rev 4)** §9.4 fixture evidence **missing, stale, or digest-mismatched** | **HALT** — identical to a FAIL. Stale means "not regenerated by this S2 session"; a digest mismatch is a FAIL, never a warning |
| **(rev 4)** `agent-engram` unreachable at S2 restore | **FAIL CLOSED** to operator handoff (§9.1 step 7) — **no prune and no resume**. A file-based prune degradation is not permitted |
| Pre-mode returns `RECONCILE_FAIL` | HALT; do not proceed to close; surface the report |
| Classification returns `SAFE_CLOSE` instead of `CASCADE` | HALT — indicates the manifest is not the fully-covered root this plan asserts. Return to Stage |
| Cascade archives anything outside `allowed_ids` | P-015 violation action: `git restore`, re-verify protected set, HALT |
| S1 attempts the a1 gate or the close after its own merge | **HALT** — §6 forbids it; S1 must checkpoint and end. Return to Stage; the two-session property has been violated |
| S2 recovery hits a validation/quarantine anomaly, ambiguity, or a `stage`-owned record | **FAIL CLOSED** to operator handoff (§9.1). No restore, no resume, no resolve. The checkpoint stays `active` and is **excluded from `cleanup_checkpoints`** regardless of age |
| `--phase lifecycle` gate invoked after the archive | **HALT** — Probe 20 tripwire; it fails closed with zero active shipments and nothing remains to re-claim |

## 11. Acceptance criteria

- **AC-1** Role Boundary carries the narrow grant with all five §5 conditions, fail-closed.
- **AC-2** Step 6.1(a1) exists, positioned after `a0` and immediately before `a`.
- **AC-3** The §6 authority-escalation rule is stated normatively in `_ship.agent.md`.
- **AC-4** C1–C6 are all applied; the delegation block names P-015 + `shipment-reconcile`.
- **AC-5** `_ship.agent.md` contains **no** enumeration of the authorized verdict set and
  **no** occurrence of `TASK_ONLY_FINALIZE`.
- **AC-6** The drifted `children`-only coverage rule is **absent**.
- **AC-7** Exactly **one** Go test function; H0 red 1 of 1; H1 green; `go vet` clean.
- **AC-8** `markdownlint` clean on the changed file.
- **AC-9** Closure follows §9 including the S1 checkpoint-and-end and the S2 fresh-session
  parent completion. AC-9 is **not** satisfied if one session performs both. The
  checkpoint MUST be written through `backlogit_create_checkpoint` (tool-timestamped) with
  `schema_version: 1`, `agent: ship`, a `phase`, a `resume_hint` naming `021-S`, `022-F`,
  `022.001-T` and the merge SHA, and domain data under `context`; its `created_at` must
  precede S2's resume. **S2 executes the full §9.1 recovery protocol — all eleven steps,
  in order.** AC-9 is **not** satisfied unless S2 demonstrably performed **every** element
  below; rev 3's shorter list is superseded and is **not** sufficient evidence:
  * enumeration via `backlogit_list_checkpoints` with `consumer_id: "ship"` and **no**
    `status`/`agent` filter;
  * **anomaly gate evaluated FIRST**, before the zero-candidate check;
  * **explicit owner selection by filename, even for a single candidate** — never auto-pick;
  * `agent: ship` ownership validation (a `stage`-owned record is never selectable);
  * **explicit operator confirmation before restore** (no automatic resume; no age-based
    liveness inference);
  * **the `restore → prune/gate → resume` ORDERING, executed in that order** — never
    restore → resume → prune **(rev 4)**;
  * **the Engram reachability gate** — `agent-engram` is installed here, so the prune step
    is mandatory and **not** the backlogit-only no-op; if the engram substrate is
    unreachable, S2 **FAILS CLOSED to operator handoff with no prune and no resume**, and a
    file-based prune degradation is **not** an acceptable substitute **(rev 4)**;
  * **the never-prune allowlist honored intact** — the active-shipment/active-task cursor
    (`021-S`), the unresolved-checkpoint pointer itself, and all recorded gate verdicts
    **(rev 4)**;
  * **same cursor** — the same single active shipment `021-S`, no parallel resume and no
    new worktree;
  * **resolve only after a confirmed successful resume**, for that one selected checkpoint
    only, with no bulk sweep;
  * **fail closed with no fresh-start fallback**; and a checkpoint left `active` by a
    fail-closed handoff is **excluded from `cleanup_checkpoints` regardless of age**
    **(rev 4)**.

  *Residual, stated honestly:* the artifact is not tamper-proof — the two-session property
  ultimately rests on process integrity, and the checkpoint is evidence, not enforcement.
- **AC-10** PO-1a is **discharged by Probe 22** and cited (§10.1); PO-1b executed before the a1 gate and recorded.
- **AC-11** Exactly 2 files changed. 0 new production files.
- **AC-12** `021-S`'s manifest is verified to be `[022-F, 022.001-T]` **at claim** (§4.3 — the re-shape is already complete; this is a verification, not a task).
- **AC-13** Probe 18 is cited as C2's justification (§10.1.1).
- **AC-14** C1 ships with the §4.2 pinned reconciled wording; Go test row 11 passes.
- **AC-15** **(rev 2)** Step 6.1(a1) states the **zero-feature-member explicit no-op** (§5.1 case i); Go test row 12 passes.
- **AC-16** **(rev 2)** Step 6.1(a1) states that an unmet §5 condition on a **present** feature member **HALTs** (§5.1 case iii) and is not treated as a no-op.
- **AC-17** **(rev 2)** Step 6.1(a1) states **idempotent resume** for a feature member already `done` (§5.2); Go test row 13 passes.
- **AC-18** **(rev 2)** The lifecycle order of §9 is followed: `022.001-T -> done` **before** the implementation PR merges; the closure branch created from fresh `main` **before** any post-merge backlog mutation; closure artifacts and P-020 on the closure branch **before** the closure PR is pushed; `sync` and post-mode in installed order.
- **AC-19** **(rev 2)** The `a0` `--phase lifecycle` gate runs **while `021-S` is still active**; **no** `--phase lifecycle` invocation occurs after the archive (Probe 20 tripwire).
- **AC-20** **(rev 4, NON-BYPASSABLE)** The §9.4 root-included cascade fixture is **executed by the fresh S2 session**, after it loads A's merged Role Boundary + Step 6.1(a1), at §9 step **12a** — **before** the a1 mutation and **before** `021-S` is marked or archived `shipped`. Its evidence is **persisted and committed** at `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.{ps1,txt}`, records `DIGEST_GATE=PASS` against SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`, and satisfies **all eight** PASS criteria of §9.4 over the **exact 13-member root-included manifest with 11 genuinely pre-archived members**. **Only a PASS permits A's final close.** A FAIL, or evidence that is missing, stale (not regenerated by this S2 session) or digest-mismatched, leaves `021-S` `active`/unshipped and `017-S → 021-S` unsatisfied. AC-20 admits **no** override and **no** accepted-residual path.
- **AC-21** **(rev 4)** Step 6.1(a1)'s selector is **exhaustive over member multiplicity** and computes `n` by `artifact_type`, never by ID suffix: `n == 0` ⇒ explicit no-op (case i); `n == 1` ⇒ full §5 guard evaluation with `active -> done` or §5.2 idempotent resume (cases ii/iii); `n > 1` ⇒ **HALT fail-closed with `A1_MULTIPLE_FEATURE_MEMBERS` and no feature mutation whatsoever** (case iv). Go test row 16 passes.
- **AC-22** **(rev 4)** Both stale queue-only pre-mode summaries are reconciled to the §4.4 authoritative semantics — **C7** at `_ship.agent.md` L255–256 (Step 0.5 item 6) and **C8** at L776–777 (Step 6.1(a)). Each reconciled sentence expresses **both** accepted dispositions: a **queue item at the expected status** *or* an **archive-located `pre-archived` member accepted without a status check**. Neither may remain queue-only. Go test rows **14** and **15** pass. The rev-3 instruction that *"a reviewer MUST NOT fail A for its absence"* is **withdrawn**.
- **AC-23** **(rev 4)** The inventory is **8 clause sites + 2 additive**, the recalculated budget is **~92 min** with a **~28 min** margin to the 2-hour rule (§12), and the Go test carries **16 rows** in **one** function with **≤4 helpers**. No site is excluded from the inventory in order to preserve a previously-pinned budget.

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity is carried as **enum-validated prose**: this
workspace's `header-def.yaml` defines `size` on the task type but **no** `complexity`
field (confirmed at `.backlogit/header-def.yaml:144–166`, and in
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`). Size is structured.

| Work item | Estimate |
|---|---|
| Role Boundary grant (ADD-1) | ~8 min |
| Step 6.1(a1) parent-completion step (ADD-2), incl. §5.1 **four-case** multiplicity-first applicability + §5.2 idempotent resume | ~18 min |
| C1 reload extension | ~4 min |
| C2 safe-close summary correction | ~4 min |
| C3 + C4 + C5 generic delegation | ~12 min |
| C6 delegation bullet | ~4 min |
| **C7** Step 0.5 item 6 pre-mode summary reconciliation (**rev 4**) | ~4 min |
| **C8** Step 6.1(a) pre-mode summary reconciliation (**rev 4**) | ~4 min |
| Go table-driven test (**16** rows) + ≤4 helpers | ~24 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~92 min ≈ 1.5 h** |

Rates are rev 6 §10.1's own, so this is directly comparable to the measurement that
produced `MUST_REPLAN`. Margin to the 2-hour rule: **~28 min**.

*(Rev-2 delta: ADD-2 **+5** for the §5.1/§5.2 applicability and idempotence text; Go test
**+2** for rows 12–13. Total ~71 → ~78 min, margin ~49 → ~42 min.)*

*(Rev-3 delta: **none**. The A-only simplification changes no surface, no clause site and
no test row, so the budget stood at **~78 min** with a **~42 min** margin. Rev 2's
comparative remark that A was *"the second-most comfortable of the three"* is withdrawn —
there is no longer a set of three to compare against.)*

*(**Rev-4 delta: +14 min, recalculated rather than preserved.** C7 **+4** and C8 **+4** —
the two pre-mode summary sites rev 3 excluded (§4.0). ADD-2 **+3** for §5.1 case (iv), the
multiplicity-first selector. Go test **+3** for rows 14–16. Total ~78 → **~92 min**; margin
~42 → **~28 min**. **Helper count unchanged at ≤4** and the function count unchanged at 1,
so Ship Step 4.3's one-function constraint still holds.)*

> **The margin is smaller and that is the honest number.** ~28 min of headroom against the
> 2-hour rule is still real headroom, and §12's contingency is unchanged: if actual work
> exceeds 2 h mid-execution, **HALT and return to Stage**. Rev 3 protected a ~42 min margin
> by leaving two defects out of the inventory; rev 4 spends 14 of those minutes fixing
> them. Trading margin for completeness is the correct direction — an under-scoped
> inventory does not make the work smaller, it only moves the discovery later.

> **Not charged to the task budget: the §9.4 fixture gate (AC-20).** Probe 25 is run by the
> **S2 closure session**, not by `022.001-T`, and is scoped exactly as PO-1a was: a
> Stage/closure-time measurement against a disposable workspace, outside the implementation
> task's 2-hour envelope. Its cost is real (~20–30 min of S2 session time) but it is not
> part of the harnessed task, does not affect the 2-hour rule for `022.001-T`, and does not
> consume the ~28 min margin.

*(Rev-1 review: +1 min for the C1 coherence row 11. PO-1a is a Stage/pre-execution read
and is not charged to the task budget.)*

**Contingency.** If actual work exceeds 2 h mid-execution: **HALT and return to Stage.**
Never split, never rush contract prose, never mis-size.

## Plan Hardening

**Blast radius.** One agent instruction file governing closure for every shipment. A
defect here affects all future closures, not just this chain. Mitigated by: the change
**removes** authority duplication rather than adding authority; four negative test rows;
and A closing by the pre-existing path, so a defect in the new wording cannot corrupt A's
own closure.

**Hardening 1 — the grant is a genuine authority increase.** Ship gains a power it did
not have. It is bounded by five conjunctive conditions, one position (Step 6.1(a1)), one
scope (the active shipment's manifest), and one transition (`active -> done`). It cannot
archive, reparent, create or delete. The residual — a Ship defect completing a feature
whose descendants are not all done — is caught by pre-mode's `expected_status: done`
sweep and by P-015's protected-set gate, both downstream and both fail-closed.
**Rev 2 adds the two cases rev 1 left unstated**: a manifest with **zero** feature members
is an explicit **no-op** (§5.1 case i), and a feature member already `done` is an
**idempotent resume** (§5.2). Neither widens the grant — the no-op performs no transition
at all, and idempotent resume recognises the terminal state of the *same* transition
without relaxing conditions 1–4. Both are asserted negatively-by-construction in Go test
rows 12–13.

**Hardening 2 — delegation must not become a blank cheque.** §4.1's wording binds Ship to
the verdict **named by the classification**, and P-015 remains the sole authorization
surface. Go test row 9 fails if `_ship.agent.md` ever re-enumerates verdicts. Ship cannot
authorize `TASK_ONLY_FINALIZE` — and after rev 3 **nothing currently authorizes it at
all**: the policy branch (B) and its classifier (C) are deferred, so the selectable
verdict set stays exactly `CASCADE`, `SAFE_CLOSE`, `HALT`. Go test row 8 independently
asserts `_ship.agent.md` never names the token.

**Hardening 3 — the self-hosting boundary is the highest-risk step.** The S1→S2 handoff is
where an impatient implementer would "just finish it". Three independent defenses: §6 is
normative contract text; AC-9 explicitly fails if one session does both; the Go test
asserts the rule is present in the merged file. The checkpoint `resume_hint` names
`021-S`, `022-F`, `022.001-T` and the merge SHA so S2 can validate rather than trust.
**Rev 2 specifies S2's side of that handoff in full (§9.1)** — rev 1's *"restore the
checkpoint"* was exactly the kind of one-line instruction that gets satisfied by a
plausible-looking shortcut. Auto-picking a sole candidate, filtering the enumeration by
status, or resolving before a confirmed resume would each defeat the property the handoff
exists to establish.

**Hardening 4 — R-1 was a real possibility and is now MEASURED, not merely planned for.**
Rev 1 correctly refused to assume claim leaves a covering feature `active`, and made PO-1a
a pre-execution read. Rev 2 **executed** it: Probe 22 ARM AB measured
`ARM_AB_FEATURE_STATUS_AFTER_CLAIM=active` on the exact `[feature, task]` shape, and
measured that the engine accepts the subsequent `active -> done` move. The expensive
discovery has been made cheaply and has come back **favourable**, so A proceeds. **PO-1b
is retained anyway** — a probe measures the engine, not this particular record at that
particular moment — and both non-`active`, non-`done` branches still halt and return to
Stage rather than widening the grant in-flight.

**Hardening 5 — engine coupling.** A performs **no** engine call that depends on the
1.10.1 digest beyond what every closure already does. The digest binding
(`1E106F5F…959A98`) was **C's** obligation and is **deferred with C** — A never needed it.
A's cascade invocation is the same call every
fully-covered-root closure already makes today. **Rev 3**: this is now strictly simpler —
with C deferred, **no** digest-bound engine invocation is introduced on any current route.

> **Rev 4 correction — this paragraph is no longer true as written (plan-review finding
> A-3).** §9.4's fixture gate **does** bind a digest: `DIGEST_GATE=PASS` against
> `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` is a **PASS criterion
> of a non-bypassable gate on A's own close**. The two statements are reconciled as follows,
> and the distinction is real but narrow:
>
> * **A's own execution path** introduces no digest-bound engine call — the sentence above
>   remains true of the *implementation* and of the `CASCADE` invocation itself.
> * **A's verification path** now does — the §9.4 fixture is digest-gated so that its
>   measurement cannot be silently attributed to a different engine build.
>
> **Residual, unresolved and recorded**: binding a *no-override* gate to one exact binary
> hash (and, worse, to the absolute path `C:\Tools\backlogit.exe`) makes A un-closable on
> any other machine, CI runner, or backlogit point release. That is finding **A-3** and is
> **not** fixed in remediation cycle 1, because the correct fix depends on how the operator
> adjudicates **A-1** (fixture placement). Recommended once adjudicated: discover the tool
> from the workspace rather than an absolute path, and record a digest mismatch as an
> explicit operator-acknowledged condition rather than an unoverridable FAIL.

**Hardening 6 — what could make this plan wrong.** If `expected_status: done` is not
actually enforced over feature members, the a1 gate is unnecessary (harmless but
pointless). If claim already sets features `active`, PO-1 passes trivially. Neither
falsifies the delegation fix, which stands on §2.2 alone. The plan is therefore
**falsifiable in both directions** and its two deliveries are **independently
justified**.

**Hardening 7 — the inventory itself was a blast-radius surface (rev 4).** Rev 3 closed
the inventory at 6 sites to protect a budget, and in doing so reproduced — inside the
plan's own inventory — the exact duplication-drift failure mode §2.2 exists to eliminate:
one instance of the defect was known and excluded, a second (C7, L255–256) was never
looked for. The generalized lesson is recorded as a constraint: **an inventory is closed by
exhaustive search over the surface, never by a budget target.** When the recalculated
budget does not fit, the response is HALT-and-return-to-Stage (§10.3), never silent
exclusion. AC-23 pins the recalculated numbers so a future revision cannot quietly
re-shrink them.

**Hardening 8 — totality of the a1 selector (rev 4).** A guard that is not **total** over
its input is not a guard. Rev 2 made a1 total over *state* (cases i/ii/iii + §5.2
idempotence); rev 4 makes it total over *multiplicity* (case iv). The residual risk of
case (iv) is nil-by-construction — it performs no mutation and evaluates no condition — and
its cost is one halt on a manifest shape no current shipment has. The asymmetry is
deliberate: the cost of the branch existing is a recorded halt; the cost of it not existing
is a feature completed by manifest ordering, or a bulk feature-lifecycle power the §5 grant
explicitly refuses.

**Hardening 9 — the fixture gate converts a confidence residual into a blocking
measurement (rev 4).** Rev 3 classified the unproven root-included cascade as *"a
confidence residual, not a safety one"* and deferred it past A. That classification was
defensible in isolation and wrong in **sequence**: A is an **irreversible contract change**
(a permanent Role Boundary authority grant) whose sole release value is making `017-S`
closable, so deferring the measurement past A's close means discovering a mis-classification
only once it is expensive to act on. §9.4 moves it **before** A's close **and before a1's
mutation**, so a FAIL costs nothing but a halt. The gate deliberately carries **no override
path**: an override would restore exactly the accepted-residual posture rev 4 withdraws.
Its placement — inside A's existing closure sequence, riding the pre-existing
`017-S → 021-S` edge — means it adds **no shipment, no backlog item and no dependency
edge**, so it cannot itself perturb the topology it is protecting.

## 13. Independent re-review scope (rev 4)

**All prior review scopes are SUPERSEDED**, including rev 3's A-only internal review and
every three-shipment-era scope (rev 5 §13, rev 6 §13). Review **rev 4 of this plan only**;
the deferred B/C plans are **not** in scope and must not be routed from.

| # | Item | Why it is in scope |
|---|---|---|
| 1 | **Adjudicate the §4.0 inventory reopening.** Are C7/C8 correctly classified as *operative* instruction text rather than descriptive drift? Is the rev-3 exclusion correctly withdrawn? | Reverses a rev-3 decision; the whole budget recalculation rests on it |
| 2 | **Verify the inventory is now exhaustive.** Is there a *third* stale queue-only or `children`-only restatement in `_ship.agent.md` that rev 4 still missed? | Rev 3 missed one; rev 4 claims closure at 8 sites |
| 3 | **Adjudicate the recalculated budget** (~92 min, ~28 min margin) and the ≤4-helper / 1-function claim at 16 rows | The rev-6 `MUST_REPLAN` verdict was a budget verdict; this is the same axis |
| 4 | **Audit §5.1 totality.** Is the four-case selector exhaustive over `n`? Is `artifact_type`-not-suffix correct? Does case (iv)'s no-evaluation rule introduce any way to strand a legitimate manifest? | Newly total; the multiplicity branch is new contract text |
| 5 | **Audit §9.1/AC-9 against the installed protocol.** Read `.github/instructions/backlogit.instructions.md` §Checkpoint-Recovery / Prune-on-Restore Protocol and Ship's Crash-Resumption Protocol directly; confirm rev 4 omits nothing further | Rev 3's list was incomplete; the claim is that rev 4's is not |
| 6 | **Adjudicate §9.4/AC-20 as a NON-BYPASSABLE gate.** Is placement at step 12a (pre-a1) correct? Are the eight PASS criteria sufficient to falsify the `CASCADE` classification claim? Is the no-override stance correct, or does it risk an unrecoverable stall? | Highest-risk new construct in rev 4 |
| 7 | **Verify the §9.4 gate adds no topology.** Confirm it creates no shipment, no backlog item and no dependency edge, and that the `017-S → 021-S` transitive argument holds | The gate's central design claim |
| 8 | **Re-audit the §6 authority-escalation property** under rev 4's larger surface — does anything in C7/C8/case-(iv) let S1 use the new authority? | Core self-hosting property |
| 9 | **Confirm A still authorizes no new verdict.** Selectable set must remain exactly `CASCADE`, `SAFE_CLOSE`, `HALT`; `TASK_ONLY_FINALIZE` must appear nowhere | Widening guard (rows 7–10) |
| 10 | **Adjudicate the §R16 correction** in `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` — is the 13-member feature-root/`CASCADE` route now the *only* operative statement, with withdrawn text clearly non-governing? | Governing-contract coherence; a contradiction there re-blocks `017-S` |
| 11 | **Confirm the operator-only PR boundary survives**, and that the fail-closed stance on the contradictory Orchestrator Step 1.5 (3a/3e) is correctly recorded **without** modifying the Orchestrator in this cycle | P-010/role-separation boundary |

**Out of scope, explicitly**: the deferred B/C plans; `shipment-reconcile/SKILL.md`
(including its stale `src/autoharness/...` classifier paths — recorded as a follow-up, not
a rev-4 edit); Orchestrator implementation repair (deferred to `019.004-T`); and the
pre-existing `SAFE_CLOSE`/exit-9 tool conflict, which remains off-route and recorded.

## Plan Review

<!-- plan-review-attempt: 1 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

- **Reviewed revision**: rev 4
- **Reviewed at commit**: `d19b954` (working tree, pre-commit)
- **Gate run**: attempt 1
- **Plan hardening required**: **yes** (declared in the plan header). **Present**: yes — a
  `## Plan Hardening` section exists with 9 hardening items. **Materially complete**: yes
  for the authority grant; **NO** for rev 4's new constructs (see A-3).
- **Verdict**: **FAIL** — gate rule *"Any P0 or P1 findings ⇒ FAIL"*. The plan MUST be
  revised before `harvest`. **No plan status advance**: `status:` stays `planned`.

### Dispatch capability (P-012)

| Capability | Status |
|---|---|
| reviewer-subagent-dispatch | `TOOL_OK` |
| model-specific-review-routing | `TOOL_OK` — anchor route `model_routing.anchor_review` = `openai` / `gpt-5.6-sol` / `high` dispatched |
| indexed knowledge retrieval (`agent-engram`) | `TOOL_OK` — `engram stats` reachable, 297 symbols, 100% embedding coverage |
| `agent-native-parity-reviewer` identity file | `TOOL_DEGRADED: agent-native-parity-reviewer-identity — declared fallback: rubric applied via dispatched general-purpose subagent on a cross-model route.` The persona was **covered, not skipped**; `.github/agents/subagents/agent-native-parity-reviewer.agent.md` is not installed in this workspace |

`dispatch_mode: multi-agent` is correct: subagent dispatch **and** model-specific routing
were both available and used. The single missing identity file is recorded as a declared
degradation of that persona's *identity source*, not of the dispatch surface.

### Persona coverage

| Persona | Trigger | Mode | Model/route | Findings (P0/P1/P2/P3) |
|---|---|---|---|---|
| Constitution Reviewer | always-on | subagent | `claude-opus-4.8` | 1 / 3 / 4 / 3 |
| Go Reviewer | always-on | subagent | `claude-opus-4.8` | 0 / 1 / 5 / 4 |
| Scope Boundary Auditor | always-on | subagent | `claude-opus-4.8` | 0 / 1 / 1 / 5 |
| Learnings Researcher | always-on | subagent | default | 1 / 3 / 8 / 3 |
| Architecture Strategist | always (cross-model) | subagent — **ANCHOR** | `gpt-5.6-sol` / high | 0 / 7 / 3 / 1 |
| Security Lens Reviewer | triggered — the plan is an **authorization** change | subagent (cross-model) | `gpt-5.5` / high | 0 / 4 / 1 / 0 |
| Agent-Native Parity Reviewer | triggered — MCP tools + agent-facing contract | subagent (cross-model), **degraded identity** | `grok-4.6` / high | 0 / 3 / 4 / 3 |

All seven selected personas were covered. No persona was skipped.

### Gate decision rationale

Rev 4 fixed the five blockers it set out to fix, and three independent personas
**affirmatively cleared** its most contested moves: the C7/C8 inventory reopening is
correct and the line anchors were verified against the installed file; the §9.1/AC-9
recovery expansion is **faithful** to the installed Prune-on-Restore protocol; and the
§12 budget arithmetic sums exactly to **92 min / 28 min margin**.

The plan nonetheless **FAILS**, because rev 4's *new* constructs — principally §9.4/AC-20
and §5.1 case (iv) — introduced defects of their own, and because three pre-existing
issues were surfaced that prior cycles missed. **Four separate personas independently
converged on §9.4/AC-20**, which is the strongest signal in this review.

### P0 findings — blocking

**A-1 · §9.4/AC-20 is a no-override gate placed AFTER the irreversible merge → unrecoverable stall.**
*(Constitution Reviewer P0-1; independently corroborated by Architecture Strategist, Scope
Boundary Auditor, and the P-010 finding A-2.)*
Sections: §9.4 "Failure semantics", §9 step 12a, AC-20, §10.3.
A's Role-Boundary change **merges to `main` at §9 step 5**, *before* S2 runs the fixture at
step 12a. So on FAIL the contract change is **already deployed to every future Ship
session**, while `021-S` is stuck `active`/in-flight — and §9.4 states there is *"no
operator override … no `--force`, and no 'proceed with a recorded residual' path."* No
in-role actor can clear that state: Stage may not close shipments (P-010), and Ship is
halted by the gate. It also **directly contradicts the plan's own §10.2 rollback**, which
requires an operator `git revert` — precisely the override §9.4 forbids.
**Recommendation (two viable resolutions; operator adjudication required):**
**(R1)** Move the fixture to a **Stage pre-execution gate** (the PO-1a/Probe-22 slot, §10.1)
so a FAIL costs nothing and nothing has merged; or **(R2)** keep an S2-side gate but make it
a **pre-merge** CI/quality gate on A's branch, and add an explicit operator-escalation path
(abandon `021-S` + operator-approved revert per §10.2). A **hybrid** preserves the operator's
stated intent at lowest risk: **Stage runs and commits Probe 25 pre-claim**, and **S2
re-verifies the committed evidence (presence + digest + PASS tokens) read-only** before A may
be marked/archived `shipped` — still non-bypassable, but with no deadlock and no P-010 issue.

### P1 findings — blocking

**A-2 · §9.4/§12/AC-20 assign Probe-25 authoring *and commit* to Ship (S2) → P-010 violation.**
*(Constitution Reviewer P1-2; Architecture Strategist.)* `docs/plans/evidence/…/probe25-*`
are plan/spike evidence artifacts under `docs/plans/`. Ship's Role Boundary **forbids**
creating or modifying deliberation, spike, plan, or review artifacts. Executed literally,
S2 records a P-010 violation and halts at step 12a — **A can never close**. The plan even
calls it *"a Stage/closure-time measurement"* (§12) while assigning it to Ship. Note the
plan's own precedent is the opposite: Probe 22 (PO-1a) is explicitly *"Stage / §9 step 0."*
**Fix**: resolved together with A-1 by the hybrid — Stage authors/commits, S2 verifies.

**A-3 · §9.4's digest binding is machine-specific and contradicts Hardening 5.**
*(Constitution Reviewer P1-3; Agent-Native Parity P3-10.)* A **non-bypassable, no-override**
gate is pinned to `C:\Tools\backlogit.exe` and one exact SHA-256. Any toolchain drift, CI
runner, or second operator yields `DIGEST_GATE=FAIL` → HALT → per A-1, no override → A
bricked. It also flatly contradicts **Hardening 5**, which asserts *"A performs **no** engine
call that depends on the 1.10.1 digest … A never needed it."*
**Fix**: replace the absolute path with workspace tool discovery; treat the digest as
recorded telemetry with an explicit operator-acknowledged mismatch path rather than a
no-override hard FAIL; and reconcile Hardening 5 with §9.4 — one of the two is wrong.

**A-4 · §9.4 PASS criterion 5 does not match the authoritative cascade API.**
*(Architecture Strategist — VERIFIED against P-015 amendment 1.21.0.)* §9.4 says
`required_ids` and `allowed_ids` must reconcile with *"no required/returned mismatch."*
The real gate is **two separately-labelled, independently-failing** conditions:
`archived_ids − allowed_ids` (unexpected artifact archived) and `required_ids − archived_ids`
(required artifact not archived). `returned_ids` is a **separate** step-2 check. A literal
Probe-25 implementation could satisfy criterion 5 while asserting neither real invariant.
**Fix**: restate criterion 5 as those two exact set equations and keep `returned_ids == []`
as its own criterion.

**A-5 · §5.1 case (iv) narrows a shape P-015 explicitly authorizes.**
*(Architecture Strategist — VERIFIED.)* P-015 item 1 is *"quantified over **every feature
member of the manifest**"* and item 4 says *"the qualifying root feature **member(s)**"* —
**plural is authorized by policy**. An unconditional `n > 1` HALT makes every legitimate
multi-root manifest permanently unclosable by Ship, which **falsifies the plan's repeated
claim that A "narrows nothing"** and that its blast radius is harmless.
**Note**: case (iv) itself was an explicit operator directive and is *safe* (it mutates
nothing). The defect is the **undisclosed narrowing**, not the halt.
**Fix**: either make a1 set-based (validate every feature member against the same five
conditions and transition all of them), or **keep the halt but disclose it honestly** as an
accepted narrowing with a recorded follow-up, and delete the "narrows nothing" claims.

**A-6 · a1 gives two different answers for an already-`done` feature member.**
*(Agent-Native Parity P1-2.)* Case (iii) fires on *"any §5 condition unmet"*, and condition 5
is *"live status is exactly `active`."* An already-`done` member **fails (iii) → HALT** and is
**simultaneously** told by §5.2 to **proceed via idempotent resume**. This is exactly the
dual-action class of defect rev 4 fixed for `n > 1`, left unfixed one row above it.
**Fix**: order the selector explicitly — compute `n`; then branch on status: `done` +
conditions 1–4 ⇒ §5.2 resume; `active` + all five ⇒ (ii); anything else ⇒ (iii). Do not nest
§5.2 inside (ii).

**A-7 · Go test rows 14–15 are vacuously satisfiable — the C7/C8 verification is hollow.**
*(Go Reviewer P1-1.)* §8.1 says rows 14–15 *"reuse the same substring-assertion helper"*, i.e.
a whole-file `strings.Contains(content, "pre-archived")`. That passes if the token appears
**anywhere** in the ~94 KB file while L255–256 and L776–777 keep their exact queue-only text —
and it may be **green at H0**, never transitioning red→green. It also contradicts §4.4's own
*"Neither sentence may be left in the queue-only form."* Since rows 14–15 (AC-22) are the
**entire mechanical justification** for reopening the inventory and spending +14 min, this
would let rev 4 go all-green with its headline defect intact.
**Fix**: make each row **compound** — assert PRESENT a long distinctive phrase only the
reconciled wording can contain, **AND** assert ABSENT that site's exact old literal. Pin all
four literals verbatim in §8.1.

**A-8 · Checkpoint payload omits the required top-level `session_id`.**
*(Agent-Native Parity P1-1.)* §9 step 8 and AC-9 enumerate `schema_version`, `agent`, `phase`,
`resume_hint` and `context` — but **not** `session_id`, which the installed Checkpoint Payload
Contract requires top-level. An agent following the plan rather than the overlay writes a
non-conforming checkpoint, which the §9.1 anomaly gate would then reject.
**Fix**: add `session_id` to §9 step 8 and AC-9; permit MCP `backlogit_create_checkpoint`
**or** CLI `backlogit checkpoint create --state-dump`.

**A-9 · S2 with zero checkpoint candidates is ambiguous — continue vs fail closed.**
*(Agent-Native Parity P1-3.)* The installed protocol treats zero active `ship` candidates as
**normal startup — explicitly not a failure**. But S2's entire purpose depends on S1's
checkpoint existing. As written, an agent may legitimately read "zero candidates" as
"proceed with fresh work", skipping parent completion.
**Fix**: specialize in §9.1 — for S2, zero active `ship` candidates is **FAIL CLOSED**, not
the ZERO-CANDIDATE continuation.

**A-10 · `pre-archived` descendant exemption can satisfy a1 without proving completion.**
*(Security Lens Reviewer.)* §5 condition 1 checks only *non-pre-archived* descendants for
`done`, and §4.4 defines `pre-archived` as **archive-located, accepted without a status
check**. An unfinished descendant that is merely archive-located therefore stops counting for
condition 1, is not live for condition 2, and can keep an intact `parent_id` for condition 3 —
letting Ship complete a covering feature whose work is not done. Missing/malformed
`artifact_type` also has no fail-closed rule before `n` is computed.
**Fix**: define "pre-archived executable descendant" for the **grant** as a *proven terminal*
descendant (`status: done` terminal relocation, or `status: archived` + valid
`archived_status`/`archived_from`), per the skill's own `archived-completed(done)` role. Treat
duplicate queue+archive records, archive-located `queued`/`active`, malformed provenance, and
missing `artifact_type` as a1 HALT conditions, evaluated **before** computing `n`.

**A-11 · Delegation (C3–C6) hands authorization to an unversioned skill boundary.**
*(Security Lens Reviewer; Architecture Strategist P2.)* Ship must invoke *"the close path named
by the machine-checkable classification"* and may not enumerate verdicts — but the implementing
surface, `shipment-reconcile/SKILL.md`, is **outside this plan's 2-file scope and outside its
review**. A future change there could name a new destructive verdict and Ship would obey with
no local provenance check. The plan also never says what Ship does when P-015 and the skill
**disagree**, or when the classifier returns nothing/an unknown path.
**Fix**: keep verdict names out of Ship, but require the classifier result to carry
`policy_id: P-015` + a policy version/commit, and require Ship to **HALT** on an absent,
unknown, ambiguous, or policy-mismatched verdict.

**A-12 · CASCADE guardrails detect some destructive mutations only *after* they occur.**
*(Security Lens Reviewer.)* `returned_ids`/two-set/`parent_id` checks run **after** the cascade
has already mutated the backlog; a non-empty `returned_ids` HALT does not explicitly restore
`.backlogit/queue/` and `.backlogit/archive/`. Worst case for `017-S` is the 13-member subtree
requeued/detached with `parent_id` cleared before the halt fires.
**Fix**: require a clean `.backlogit/` baseline + restorable snapshot before any CASCADE, and
make **every** cascade HALT path run an unconditional `git restore -- .backlogit/queue/
.backlogit/archive/` and re-verify before halting.

**A-13 · Missing mandatory `## Constitution Check` section.**
*(Constitution Reviewer P1-1.)* The constitution's Governance section requires **every**
implementation plan to include a "Constitution Check" mapping the work against Principles
I–XI and documenting justified violations. The plan maps exhaustively against *workflow
policies* but never against the *constitutional principles*, so its genuinely justified
deviations (task-granularity heuristics — 16 test scenarios vs "fewer than 4"; 1 fn + ≤4
helpers vs "fewer than 5 functions"; `tests/integration/` vs inline `_test.go`) are
undocumented.
**Fix**: add `## Constitution Check` with those deviations and their rationale.

**A-14 · Backlog records are a rev-3 split-brain contract.**
*(Architecture Strategist; Scope Boundary Auditor P3.)* `022.001-T` still states *"Plan rev 3"*,
*"6 SITES"*, *"13 ROWS"*, *"~78 min … ~42 min margin"*, *"AC-1..AC-19"*; `022-F` echoes rev 3.
Ship consumes **both** the queue record and the plan, so it would execute the rev-3 contract
and silently omit C7, C8, case (iv), rows 14–16 and AC-20–AC-23.
**Fix**: Stage resyncs both records to rev 4 **before** A is claimed. *(Remediated in this
session — see the Remediation Cycle 1 note below.)*

**A-15 · §R16 of the governing policy-gap plan is internally contradictory.**
*(Architecture Strategist.)* R16.2 declares the 13-member fully-covered root current, while
R16.3/R16.3.1/R16.3.1a/R16.3.2/R16.3.3 still operatively assert the manifest *"now contains 12
entries and is task-only"*, that the 13-entry `CASCADE` shape is *"WITHDRAWN"* and *"unreachable
by construction"*, and that *"neither is remediable."* None of that sits in the historical
appendix, and §R16 is declared *"the whole governing contract."* This directly contradicts the
route Plan A depends on.
**Fix**: demote the withdrawn analysis to clearly non-governing historical labels and leave one
unambiguous operative statement of the 13-member CASCADE route. *(Remediated in this session.)*

**A-16 · Rev-4 header claimed an executed `plan-review` gate (§14) that did not exist.**
*(Constitution P2-2; Scope Boundary P3; Architecture Strategist.)* The rev-4 banner asserted
*"Formal plan-review gate executed (§14) … now appended"* while the file ended at §13.
**Fix**: this section is that gate. The banner claim is corrected in the same pass.

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| B-1 | §3/AC-11 *"exactly 2 files changed"* is contradicted by AC-20's mandated `probe25-*.{ps1,txt}` commit. Qualify as "2 implementation files + N closure-evidence files" | Constitution, Scope Boundary |
| B-2 | Row 3's ordering anchors (`"Step 6.1(a1)"`) **do not exist** in `_ship.agent.md` — the file uses `a0.`/`a.`/`b.` list markers. A naive `strings.Index("a.")` matches the first `a.` in a 94 KB doc. Pin unique anchors (`TOPOLOGY_GATE: lifecycle`, `Pre-archive reconciliation gate (mandatory)`, and a new unique a1 heading) | Go Reviewer |
| B-3 | Negative needles (rows 7–10) are unpinned; a bare `children` needle can **never** go green and would permanently block H1. Pin the full drifted clause and the full `backlogit move <shipment_id> --status shipped` literal | Go Reviewer |
| B-4 | *"H0 red 1 of 1"* is not automatic — a 16-row table yields many reds. Requires a single marker-fatal guard **before** the row loop | Go Reviewer |
| B-5 | `technology-go.instructions.md` says "use testify", but testify is **not** in `go.mod`; importing it breaks AC-11's 2-file surface. Mandate stdlib assertions explicitly | Go Reviewer |
| B-6 | ~24 min for 16 rows is optimistic while **zero** of the 16 needles/anchors are pinned; the derivation churn can consume the ~28 min margin | Go Reviewer |
| B-7 | §9.4 lacks machine checks for **live-backlog isolation** (prove `FIXTURE_BACKLOG_ROOT` ≠ / not nested under live `.backlogit`; before/after `git status` on live `.backlogit`) and for **evidence redaction** (absolute user paths, emails, tokens in the committed `.txt`) | Security Lens |
| B-8 | a1's descendant traversal duplicates graph-walking already owned by P-015 + `shipment-reconcile` — recreating, for mutation eligibility, the same duplication pattern C4 removes for verdict selection | Architecture Strategist |
| B-9 | The a1 mutation occurs **before** pre-mode acquires the single-writer lock, and the ~20–30 min fixture sits between `a0` and the close, so topology can go stale inside the closure critical section (TOCTOU) | Architecture Strategist |
| B-10 | PO-1b mandates CLI `backlogit get` with no justification — `backlogit_get_item` reads by ID too. "Read by status, never by path" is right; CLI-exclusivity is not | Agent-Native Parity |
| B-11 | §4.4 drops two authoritative PROCEED conditions: **no orphans** and `record-consistent`. C8's sentence also carries the lock acquisition and orphan scan, which a careless rewrite would delete | Agent-Native Parity |
| B-12 | The complexity degradation claim is accurate but incomplete — the plan never **forbids** the advertised-but-failing `backlogit update --complexity` call (exit 1) | Agent-Native Parity |
| B-13 | a1 does not pin the move tool; specify `backlogit_move_item` / `backlogit move {id} --status done` only, never a file edit | Agent-Native Parity |
| B-14 | C7's replacement must coexist with the existing `Scope note (139-F/139.001-T)` at L257+, which scopes item 6 to *uniform-status* intake — `017-S`'s intake manifest is **not** uniform-status | Scope Boundary |
| B-15 | §6's self-hosting property is enforced only by contract text + a text-presence test; nothing mechanically stops S1 continuing. Honestly disclosed, but worth explicit operator sign-off at the S1→S2 boundary | Constitution, Security Lens, Architecture Strategist |
| B-16 | Task-granularity heuristics (16 scenarios vs "<4"; 5 functions vs "<5") breached without documented justification — ties to A-13 | Constitution |
| B-17 | The Go test is text-only and provides **no** executable blast-radius coverage for task-only/partial-feature/multi-feature manifests on the shared closure path | Architecture Strategist |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| C-1 | Principle VIII: a global-closure blast radius is never paired with an explicit careful/freeze-scope safety-mode directive | Constitution |
| C-2 | H0's `not implemented:` marker risks the stub anti-pattern; prefer running the 16 real assertions and emitting the marker from the first failure, with `t.Run` per row for traceability | Constitution, Go Reviewer |
| C-3 | `_contract_test.go` in `tests/integration/` deviates from the instructions' `tests/contract/` taxonomy (justified by precedent; no `tests/contract/` exists) | Constitution, Go Reviewer |
| C-4 | Helper taxonomy is imprecise — only row 3 is positional; rows 11–13/16 are multi-substring presence checks. Budget conclusion (≤4) still holds | Go Reviewer |
| C-5 | Reuse the existing package-level `repoRoot(t)` from `build_script_test.go`; redefining it is a duplicate-declaration compile failure at H0 | Go Reviewer |
| C-6 | Keep needles single-line — the target file may be CRLF on Windows | Go Reviewer |
| C-7 | Case (iv)'s justifying prose ("what Stage produces if a grouping session ever…") is speculative; rest the case on **totality** alone | Scope Boundary |
| C-8 | §13 items 10–11 reach into two adjacent governing documents; keep them strictly read-only coherence checks | Scope Boundary |
| C-9 | AC-9 names only the MCP create; the CLI create is equally valid. §9 step 17 names CLI `backlogit sync`; MCP is `backlogit_sync_index` | Agent-Native Parity |

### Affirmative clearances (recorded as evidence, not findings)

* **C7/C8 reopening is correct.** Both anchors verified in the installed file: L255–256
  reads *"present in `.backlogit/queue/` with the expected status"*; L776–777 reads
  *"present in queue with `status: done`"*. Neither contains `pre-archived`. Both are inside
  A's declared file #1, so the **file surface does not widen**. Rev 3's budget-driven
  exclusion was itself the scope error. *(Scope Boundary, Agent-Native Parity, Go Reviewer.)*
* **§9.1/AC-9 expansion is faithful.** The installed `backlogit.instructions.md`
  Prune-on-Restore protocol contains verbatim the `restore → prune/gate → resume` ordering,
  the three-class never-prune allowlist, the Engram-unreachable fail-closed rule, and the
  `cleanup_checkpoints` exclusion. `agent-engram` **is** installed, so these are mandatory,
  not a no-op. Nothing was invented. *(Scope Boundary, Agent-Native Parity.)*
* **Budget arithmetic is correct**: 8+18+4+4+12+4+4+4+24+10 = **92**; 120−92 = **28**.
  *(Scope Boundary.)*
* **All C1–C8 line anchors verified** against the installed file: ADD-1 L38; C7 255–256;
  C1 756–764; C8 776–777; C2 789–791; C3 794–795; C4 802–823; C5 821–823. *(Architecture
  Strategist.)*
* **The §2.2 drift is real**: Ship's copy checks direct `children`; P-015 requires
  descendants at every depth and says a direct-children check *"is insufficient"*.
  *(Architecture Strategist.)*
* **Compound library confirms** the exit-9 shipment-move refusal, the shipment lifecycle
  triple, and the `complexity`-not-settable degradation. *(Learnings Researcher.)*
* **No P-016 violation**: S1→S2 is sequential on one branch; §9.1 forbids a new worktree;
  the probe workspace is a scratch data dir, not a git worktree. *(Constitution.)*
* **A authorizes no new verdict**; the selectable set remains `CASCADE`, `SAFE_CLOSE`,
  `HALT`. *(Constitution, Architecture Strategist.)*

### Runtime verification and operational closure gaps

* **Runtime verification**: the only executable artifact is a **text-contract** test. No
  runtime behavior of the closure path is exercised by A itself (B-17). §9.4 was rev 4's
  attempt to close this gap and is itself the P0 (A-1).
* **Operational closure**: §9 steps 16–18 correctly place `operational-closure
  mode=post-merge`, P-020 compaction, `backlogit sync`, and the closure PR on the closure
  branch before push. **No gap.**
* **Rollback**: §10.2 is sound for a 2-file revert but is **contradicted by §9.4's
  no-override clause** (A-1) and does not cover the AC-20 evidence files (B-1).

### Required before re-review (P0/P1 remediation queue)

A-1 and A-2 are **jointly decisive** and require **operator adjudication** of the fixture
placement (R1 / R2 / hybrid) before any further remediation of §9.4 is meaningful.
A-3 → A-13 are remediable in-plan. A-14, A-15 and A-16 were remediated in this session.

## Plan Review — Remediation Cycle 1 disposition

This session is **remediation cycle 1** of the attempt-1 `FAIL`. Dispositions:

| Finding | Disposition |
|---|---|
| **A-14** | **CLOSED** — `022-F` and `022.001-T` resynced to rev 4 (8 sites, 16 rows, case (iv), ~92 min, AC-1…AC-23) |
| **A-15** | **CLOSED** — §R16 corrected; the withdrawn 12-entry/task-only/unreachable analysis is demoted to explicitly non-governing historical labels |
| **A-16** | **CLOSED** — this `## Plan Review` section is the executed gate; the rev-4 banner claim is corrected |
| **A-4** | **CLOSED** — §9.4 criterion 5 restated as the two independent set equations |
| **A-5** | **CLOSED (disclosure)** — the "narrows nothing" claim is withdrawn; case (iv)'s narrowing of P-015's authorized multi-root shape is now disclosed and recorded as a follow-up. The operator-directed halt itself is retained |
| **A-3** | **PARTIAL** — the Hardening 5 / §9.4 digest contradiction is reconciled in-plan; the machine-specific path binding remains pending A-1 adjudication |
| **A-1, A-2** | **OPEN — operator adjudication required.** The gate is retained exactly as directed; both findings are recorded unmodified |
| **A-6 … A-13, B-*, C-*** | **OPEN** — queued for remediation cycle 2, which **requires a fresh `plan-review` (attempt 2)** |

**The gate verdict for this plan remains `decision: FAIL`.** Remediation cycle 1 does not
convert a FAIL into a PASS: rev 5 must be re-reviewed. `status:` stays `planned`, the plan is
**not** harvest-ready, and `021-S` must **not** be claimed.

