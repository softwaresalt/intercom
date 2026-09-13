---
title: "Plan — Ship covering-feature completion and close-path delegation (Foundation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 5
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

> **Rev 5 — remediation cycle 2 of the attempt-1 `FAIL`. The Probe-25 hybrid is APPLIED.**
> Rev 4's own `## Plan Review` returned `FAIL` with 1 P0 and 12 P1s. Cycle 1 closed five
> (A-4, A-5, A-14, A-15, A-16). **Rev 5 closes the remaining eleven** — A-1, A-2, A-3,
> A-6…A-13 — plus the cheap same-surface P2/P3 queue. The single structural change:
>
> 1. **The §9.4 fixture is re-seated as a Stage pre-claim gate with a read-only Ship
>    re-verification (§9.4, §9.4a).** This is the **hybrid** the attempt-1 review
>    recommended and the operator selected. **Stage has already authored, executed and
>    committed Probe 25** — `PROBE25_RESULT=PASS`, all eight criteria, measured at `4fd21c5`
>    **before** `021-S` is claimed. Ship's fresh S2 session now only **re-verifies that
>    committed evidence read-only**. This closes **A-1** (no post-merge no-override
>    deadlock) and **A-2** (Ship never authors plan evidence, so no P-010 violation).
> 2. **Engine identity is portable (§9.4, A-3).** The contract resolves `backlogit` at
>    runtime via the registered command and compares its SHA-256 to the committed digest.
>    No runtime requirement hardcodes an absolute path; the observed path is recorded as
>    telemetry only.
> 3. **a1 is now a strictly-ordered, non-overlapping selector (§5.1, A-6)** with the
>    pre-archived descendant anomaly gate evaluated **before** `n` is computed (**A-10**).
> 4. **Test rows 14–16 are compound and non-vacuous (§8.1, A-7)** with all four old literals
>    and all four new phrases pinned verbatim. Row count **16 → 18**; budget **~92 → ~99 min**.
> 5. **A `## Constitution Check` section is added (A-13)**, and §9.1/AC-9 gain top-level
>    `session_id` (**A-8**) and an S2-specific fail-closed zero-candidate rule (**A-9**).
>
> **Rev 5 is submitted to a FRESH `plan-review` as attempt 2 — and it returned
> `decision: FAIL`.** The gate closed every attempt-1 P0/P1, but found **1 new P0 and 7 new
> P1s** in rev 5's own constructs. The decisive one: **§5.1's new S0 anomaly gate HALTs on
> A's own manifest**, because `022.001-T` is `done` (and therefore archive-located) by the
> time a1 runs — two personas converged on it and the committed Probe-25 transcript proves
> it. **Rev 5 is NOT harvest-ready and `021-S` must NOT be claimed.** See
> `## Plan Review — attempt 2` for the full finding set and the cycle-3 remediation queue.

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

> **Rev 5 — the two PROCEED conditions rev 4 dropped (plan-review finding B-11, normative).**
> `PROCEED` is **not** "every item is `matched` or `pre-archived`" alone. The authoritative
> rule carries two further conjuncts that the rev-4 summary silently lost, and a reconciled
> sentence that omits them would be a *new* inaccuracy replacing the old one:
>
> * **no orphans** — the orphan scan must find none; and
> * **`record-consistent`** — the shipment-record-status classification must be
>   `record-consistent`; any other value is `HALT — operator reconcile required`.
>
> **C8 additionally carries the lock acquisition and the orphan scan** in the same sentence
> region. The reconciliation MUST preserve those clauses intact. An implementer who rewrites
> the sentence wholesale and drops the single-writer lock or the orphan scan has introduced a
> **worse** defect than the queue-only wording being fixed. Asserted by the ABSENT/PRESENT
> literals pinned in §8.1 rows 14–15 and by new row **17**.

> **Rev 5 — C7 must COEXIST with the adjacent scope note (plan-review finding B-14).**
> Immediately after C7, at L257+, `_ship.agent.md` carries a
> `Scope note (139-F/139.001-T)` which scopes Step 0.5 item 6 to **uniform-status** intake.
> `017-S`'s intake manifest is **not** uniform-status — it is 2 live + 11 pre-archived. C7's
> replacement MUST be applied **without deleting or contradicting** that scope note: the note
> narrows *when item 6's check applies*, while C7 corrects *what that check accepts when it
> does apply*. These are orthogonal. Asserted by new row **18** (the scope note survives).

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

> **Rev 5 — the delegation TRUST BOUNDARY (plan-review finding A-11, normative).**
> Delegation is **not** unbounded trust in whatever the skill emits. A delegates exactly one
> thing — **selection among close paths that already exist** — and nothing else.
>
> **The allowed verdict set is CLOSED at its current membership.** Ship may act on exactly:
>
> | Verdict | Close path |
> |---|---|
> | `CASCADE` | the P-015 verified fully-covered-root cascade sub-procedure |
> | `SAFE_CLOSE` | the default per-item safe-close procedure (skill steps 1–10) |
> | `HALT` | no close; report and stop |
>
> **Ship MUST HALT — fail closed, no mutation — on ANY of:** a verdict outside that set (a
> name the classifier invented or a future version added); an **absent**, **empty**,
> **unparseable** or **ambiguous** classifier result; a classifier that is **not installed
> or not readable**; or a result whose `policy_id` is not `P-015` or whose recorded policy
> version does not match the P-015 amendment this plan was reviewed against.
>
> **P-015 and the skill must AGREE.** The verdict is authoritative only when the skill's
> machine-checkable classification and P-015's stated preconditions select the **same**
> close path. On **disagreement**, Ship HALTs and returns the shipment to **Stage** — it
> never prefers one source, and never breaks the tie itself.
>
> **No unversioned future capability is authorized.** A grants Ship no ability to act on a
> verdict that does not exist today. If a future change adds one, that change must carry its
> own review and its own authorization; it does **not** inherit A's delegation. This is the
> answer to the objection that `shipment-reconcile/SKILL.md` sits outside A's 2-file surface
> and outside its review: A does not trust that file's *future* contents, because the verdict
> set Ship may act on is pinned **here**, in A's own reviewed file.
>
> **Why this is not a re-derivation.** Ship still does not compute *which* path applies —
> that remains delegated. Ship only checks that the answer it was handed is **one of the
> answers that exist**, and that its two authorities concur. Validating an input is not
> restating a classification.

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

### 5.0 What "pre-archived descendant" means for THIS grant (rev 5, normative, fail-closed)

**Plan-review finding A-10.** Condition 1 exempts *pre-archived* descendants from the `done`
check. Rev 4 inherited pre-mode's definition — **archive-located, accepted without a status
check** — which is correct *for pre-mode* but **unsafe as a completion exemption**. Under
that reading, an unfinished descendant that is merely archive-located stops counting for
condition 1, is not live for condition 2, and can keep an intact `parent_id` for condition 3
— letting Ship complete a covering feature whose work is **not done**.

**The exemption is therefore NOT granted on archive-location, and NOT on `archived_status:
queued`.** Being `archived_status: queued` **does not prove completion** and is never
treated as proving it.

A manifest descendant is **excluded from condition 1** only when **all five** hold:

1. **P-015 pre-mode classifies it `pre-archived`** — the classification is taken from the
   authoritative scan, never re-derived by a1;
2. **exactly one location** — the record exists in `.backlogit/archive/` and **not** in
   `.backlogit/queue/`. A record in **both** is a torn/duplicate state; a record in
   **neither** is missing;
3. **immutable across the gate** — its content hash and `parent_id` are unchanged from the
   pre-gate snapshot;
4. **valid archive provenance** — a well-formed `archived_from`, a present `archived_status`,
   and a resolvable `artifact_type`; and
5. **a recorded Stage decision maps it as a superseded / non-obligatory closure member** —
   an explicit, named disposition that this descendant is **not** part of the obligation the
   covering feature's completion asserts.

**Condition 5 is the load-bearing one.** Conditions 1–4 establish that the record is
*genuinely and stably archived*; only a **Stage decision** establishes that it is
*not owed*. Without it the exemption would still be inferring completion from location.

**`017-S` — the exact 11 IDs, mapped.** Stage's disposition covers precisely these eleven
pre-archived manifest descendants of `018-F`, and no others:

| # | ID | Kind | Parent |
|---|---|---|---|
| 1 | `018.001-T` | task | `018-F` |
| 2 | `018.001.001-ST` | subtask | `018.001-T` |
| 3 | `018.001.002-ST` | subtask | `018.001-T` |
| 4 | `018.001.003-ST` | subtask | `018.001-T` |
| 5 | `018.002-T` | task | `018-F` |
| 6 | `018.002.001-ST` | subtask | `018.002-T` |
| 7 | `018.002.002-ST` | subtask | `018.002-T` |
| 8 | `018.002.003-ST` | subtask | `018.002-T` |
| 9 | `018.003-T` | task | `018-F` |
| 10 | `018.003.001-ST` | subtask | `018.003-T` |
| 11 | `018.003.002-ST` | subtask | `018.003-T` |

**Any pre-archived manifest descendant NOT on this list ⇒ HALT.** The list is the
authorization; membership is not inferable. If `017-S`'s live manifest at its own gate
presents a pre-archived descendant that is absent here, a1 **halts and returns to Stage**.

**`021-S` (A itself) has NO pre-archived descendants.** Its manifest is `[022-F,
022.001-T]`, both live. The exemption is therefore **never exercised on A's own route** —
it exists for `017-S` and is gated fail-closed for everything else.

**The general authority stays narrow.** This is an exemption from *one* condition, for
*enumerated* IDs, under a *recorded* Stage decision. It is not a general rule that archived
descendants don't count, and it must not be generalized into one.

### 5.1 Step 6.1(a1) applicability — a STRICTLY ORDERED, non-overlapping selector (rev 2; re-scoped rev 3; made total in rev 4; **ordered and disambiguated in rev 5**)

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

> **Rev 5 — the selector is a strictly ordered sequence, not a set of predicates
> (plan-review finding A-6).** Rev 4's table let an already-`done` feature member match
> **two** rows at once: case (iii) fires on *"any §5 condition unmet"* and condition 5 is
> *"live status is exactly `active`"*, so a `done` member is told to **HALT**; while §5.2
> **simultaneously** tells it to proceed via idempotent resume. That is the same
> dual-action defect rev 4 fixed for `n > 1`, left unfixed one row above it.
>
> **The fix is ordering, not re-wording.** a1 executes steps **S0 → S4 in order** and stops
> at the first that applies. §5.2 is **not nested inside** any case; it is reached only at
> **S3**. **No input can satisfy two steps** — each step's guard is disjoint from every
> earlier one by construction, because every step after S0 assumes all earlier guards failed.

**S0 — ANOMALY GATE (before `n` is computed).** Evaluated over every manifest member,
**first**, because a malformed record makes the multiplicity count itself untrustworthy.
**HALT, fail closed, no mutation** on any of:

* a member whose `artifact_type` is **missing, empty, or unresolvable**;
* a member present in **both** `.backlogit/queue/` and `.backlogit/archive/` (torn/duplicate);
* a member present in **neither** (missing);
* a member with **malformed archive provenance** (absent/ill-formed `archived_from` or
  `archived_status` on an archive-located record);
* an **archive-located member that is not on the §5.0 recorded Stage disposition list**.

Only if S0 raises nothing does a1 compute `n`.

**S1 — `n == 0` ⇒ NOT APPLICABLE.** Record `A1_NOT_APPLICABLE: no feature member in
manifest` and proceed to Step 6.1(a). A **successful** outcome — not a skip, not a failure,
and **no** transition of any kind.

**S2 — `n > 1` ⇒ AMBIGUOUS, HALT.** Record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`. **Evaluate
no §5 condition and mutate no feature**, not even one that would individually qualify.
Return to **Stage** for manifest disposition. *(Placed before the `n == 1` branches so the
single-member steps below may assume a unique referent.)*

**S3 — `n == 1` and that member's live status is exactly `done` ⇒ IDEMPOTENT NO-OP —
but only after revalidation.** Read the member's live status per **PO-1b** (§10.1).
Re-evaluate §5 conditions **1, 2, 3 and 4** — every non-status guard. Then:

* **all four hold** ⇒ record `A1_ALREADY_DONE: {feature_id}` and proceed to Step 6.1(a)
  **without error and without re-issuing the transition**;
* **any of the four fails** ⇒ **HALT** (this is the S4 failure disposition), because a
  feature that is `done` while its descendants are *not* is a corruption, not a resume.

Condition 5 is **not** evaluated at S3 — reaching S3 *is* the status determination. This is
the only step that may observe a `done` member, so no other step can give a second answer.

**S4 — `n == 1` and live status is exactly `active` ⇒ evaluate the full §5 guard.**

* **all five conditions hold** ⇒ perform `active -> done` on that member and record the
  transition;
* **any condition unmet** ⇒ **HALT**, fail closed. Surface which condition failed. Do **not**
  proceed to Step 6.1(a). Do **not** widen any condition in-flight.

**S5 — `n == 1` and live status is anything else** (`queued`, `blocked`, `review`, or any
other value) ⇒ **HALT**, fail closed. Record the observed status. Return to **Stage**.

| Step | Guard (assumes every earlier guard failed) | Outcome |
|---|---|---|
| **S0** | any member anomaly | **HALT** — no `n`, no mutation |
| **S1** | `n == 0` | explicit no-op ⇒ proceed to `a` |
| **S2** | `n > 1` | **HALT** — no condition evaluated, no mutation |
| **S3** | `n == 1`, status `done` | conditions 1–4 revalidate ⇒ idempotent no-op; else **HALT** |
| **S4** | `n == 1`, status `active` | all five ⇒ `active -> done`; else **HALT** |
| **S5** | `n == 1`, any other status | **HALT** |

**Totality and disjointness.** S1/S2/S3/S4/S5 partition the input exactly: `n` is either
`0`, `>1`, or `==1`; and in the `==1` case the live status is either `done`, `active`, or
something else. Every input reaches **exactly one** step, and **no input reaches two**.

> **Legacy case-label map (rev 2–4 → rev 5).** Prior revisions, the acceptance criteria and
> the appended review all refer to cases (i)–(iv). The labels are retained for traceability
> and map onto the ordered steps as follows. Where a rev-4 label was ambiguous, the map
> records which step now owns the input:
>
> | Legacy label | Rev-5 step | Note |
> |---|---|---|
> | case (i) — not applicable | **S1** | unchanged |
> | case (ii) — applies, proceed | **S4** (all five hold) | `done`-member sub-case **moved out** to S3 |
> | case (iii) — applies, halt | **S4** (guard unmet), **S5** (wrong status), **S3** (revalidation failure) | rev 4's single label covered three distinct inputs |
> | case (iv) — ambiguous, halt | **S2** | unchanged; now evaluated **before** the `n == 1` steps |
> | §5.2 idempotent resume | **S3** | no longer nested inside case (ii) |
> | *(new in rev 5)* | **S0** | the anomaly gate — no legacy label existed |

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

**The distinction between S1 and S4/S5 is load-bearing.** S1 is "there is nothing
here for this gate to do". S4/S5 are "there is something here and it is not in the
required state". Collapsing them — treating an unmet guard on a real feature member as a
benign no-op — would let a shipment close with a covering feature left `active` or with
descendants unfinished, which is precisely the corruption the gate exists to prevent. A
zero-feature-member manifest must never be reported as a guard failure either: that would
make any task-only shipment un-closable at the gate.

**S2 is fail-closed for a different reason than S4.** S4 halts because a known
member is in the wrong state. S2 halts because **a1 cannot determine which member the
grant is about.** §5 grants completion of *"a covering feature"* — singular, and scoped to
"the active shipment's manifest". With two feature members there is no unique referent, so
**the grant does not apply at all** and a1 has no authority to mutate either one. Halting
without evaluating conditions is deliberate: evaluating them first would invite an
implementer to proceed when exactly one of the several happens to qualify, which is the
manifest-ordering-dependent behavior this step exists to forbid.

Asserted by **AC-15** (S1), **AC-16** (S4), **AC-17** (S3) and **AC-21** (the full ordering).

### 5.2 Idempotent resume (rev 2; **re-seated as selector step S3 in rev 5**)

A1 MUST be **idempotent**. If the feature member's live status is already `done` when a1
runs, a1 reaches **selector step S3** (§5.1) — **not** case (ii), and **not** case (iii).

> **Rev 5 — the ordering is now explicit, and the no-op is CONDITIONAL.** Rev 4 stated the
> resume as an unconditional proceed, which (a) collided with case (iii)'s halt and (b)
> would have let a1 wave through a feature that is `done` while its descendants are not.
> **S3 records `A1_ALREADY_DONE: {feature_id}` and proceeds ONLY after §5 conditions 1, 2,
> 3 and 4 — every non-status guard — are RE-EVALUATED and all hold.** If any of the four
> fails, S3 **HALTs**. An already-`done` feature is a resume only when everything that made
> it completable is still true.

This is not hypothetical tidiness — it is required by A's own two-session closure. If S2
completes `022-F -> done` and then fails anywhere between a1 and the close (a
`RECONCILE_FAIL`, an evidence re-verification failure, an operator halt), the recovery
session re-enters at a1 with the feature already `done`. Without idempotent resume,
condition 5 (*"live status is exactly `active`"*) would fail closed and A could **never** be
closed by any session — a permanently stuck shipment created by its own partial success.

Idempotent resume is scoped tightly: it recognises the **already-satisfied terminal state**
of the transition a1 itself performs. It does **not** relax conditions 1–4 — they are
**re-verified on every entry, including this one** — and it does **not** authorize any
transition from any status other than `active`. A feature in `queued`, `blocked`, or any
other status reaches **S5** — **HALT**. Asserted by **AC-17**.

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

**Contract text and ordering on `_ship.agent.md` only.** **18 rows** (rev 5: 16 → 18).

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
| 12 | Step 6.1(a1) states the **zero-feature-member explicit no-op** (S1) and that an unmet guard on a present feature member **halts** (S4) | **ordering/consistency** |
| 13 | Step 6.1(a1) states **idempotent resume** — a feature member already `done` proceeds without error and without re-issuing the transition, **after conditions 1–4 revalidate** (S3) | **ordering/consistency** |
| 14 | **(C7)** **COMPOUND** — Step 0.5 item 6's pre-mode summary **PRESENT**: the reconciled archive-located disposition phrase; **AND ABSENT**: the exact old literal | **compound +/−** |
| 15 | **(C8)** **COMPOUND** — Step 6.1(a)'s pre-archive gate summary **PRESENT**: the reconciled phrase; **AND ABSENT**: the exact old literal | **compound +/−** |
| 16 | **COMPOUND** — Step 6.1(a1)'s selector states **`n > 1` halts with no feature mutation** (S2) **AND** the anomaly-gate-before-`n` rule (S0) | **compound** |
| 17 | **(rev 5, B-11)** C8's reconciled sentence **PRESERVES** the single-writer lock clause **and** the orphan scan clause | **preservation** |
| 18 | **(rev 5, B-14)** C7's adjacent `Scope note (139-F/139.001-T)` **survives** the C7 replacement | **preservation** |

#### 8.1.1 Pinned literals (rev 5, normative — plan-review finding A-7)

Rows 14–15 were **vacuously satisfiable** in rev 4: §8.1 said they *"reuse the same
substring-assertion helper"*, i.e. a whole-file `strings.Contains(content, "pre-archived")`.
That passes if the token appears **anywhere** in the ~94 KB file while L255–256 and
L776–777 keep their exact queue-only text — and it may be **green at H0**, never
transitioning red→green. Since rows 14–15 (AC-22) are the **entire mechanical
justification** for reopening the inventory, that would let the plan go all-green with its
headline defect intact.

**Every row below is COMPOUND: assert a long distinctive PRESENT phrase that only the
reconciled wording can contain, AND assert the site's exact old literal is ABSENT.**
All needles are **single-line** (C-6: the target file may be CRLF on Windows) and are
pinned **verbatim** here:

| Row | MUST BE ABSENT (exact current literal) | MUST BE PRESENT (distinctive reconciled phrase) |
|---|---|---|
| 14 (C7) | ``every manifest item is present in `.backlogit/queue/` with the`` | ``an archive-located `pre-archived` member is accepted without a status check`` |
| 15 (C8) | ``queue with `status: done`, and scans for orphan items.`` | ``either a queue item at `status: done` or an archive-located `pre-archived` member`` |
| 7 | ``it is fully covered (every one of its children,`` | ``descendants — at every depth, not only direct children`` |
| 10 | ``<shipment_id> --status shipped`` | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` |

**Negative needles are pinned to the FULL drifted clause, never a bare token (B-3).** A bare
`children` needle can **never** go green — `_ship.agent.md` legitimately uses the word
elsewhere — and would permanently block H1. Row 7 therefore asserts absence of the
**full single-line clause** above, not the word.

**Row 3's ordering anchors are unique (B-2).** Rev 4 anchored on `"Step 6.1(a1)"`, which
**does not exist** in the installed file — it uses `a0.`/`a.`/`b.` list markers, so a naive
`strings.Index("a.")` matches the first `a.` in a 94 KB document. Pinned anchors:

| Anchor | Literal |
|---|---|
| `a0` | ``TOPOLOGY_GATE: lifecycle (before closure/safe-close)`` |
| `a` | ``Pre-archive reconciliation gate (mandatory)`` |
| `a1` (new, unique) | ``a1. **Covering-feature completion gate (P-015 grant)**`` |

Row 3 asserts `index(a0) < index(a1) < index(a)`, all three found, each exactly once.

**Rows 17–18 assert PRESERVATION, not replacement.** Row 17 pins ``(via the `file-lock`
skill)`` and ``scans for orphan items`` as still-present after C8's rewrite; row 18 pins
``Scope note (139-F/139.001-T)`` as still-present after C7's. These exist because the most
likely implementation error is a wholesale sentence replacement that silently deletes
surrounding clauses.

**H0 red is `1 of 1` by construction (B-4).** An 18-row table naturally yields many reds. A
single **marker-fatal guard runs BEFORE the row loop**: if the file does not contain the
marker `not implemented: ship feature-completion delegation contract`, the test fails once
with that marker and **returns**, so H0 is exactly one failure. At H1 the marker is gone and
the 18 rows execute. Each row runs under its own `t.Run(row.name, …)` for traceability
(C-2), so the table is **real assertions from the start**, not a stub.

**Assertions are stdlib only (B-5).** `technology-go.instructions.md` recommends testify,
but **testify is not in `go.mod`**; importing it would add a dependency and break AC-11's
2-file surface. Use `strings.Contains` / `strings.Index` + `t.Errorf`. **Reuse the existing
package-level `repoRoot(t)` from `build_script_test.go`** — redefining it is a
duplicate-declaration **compile failure** at H0 (C-5).

Rows 7–10 are the **widening guards**: they fail if this change authorized anything new.
Row 11 is the **coherence guard**: it fails if C1 ships with the carve-out and the
"close under the just-merged contract" directive unreconciled.
**Rows 12–13 are the applicability guards** (rev 2): they fail if a1 ships without an
explicit not-applicable outcome — which would leave any task-only manifest ambiguous at
the gate — or without idempotent resume, which would make a partially-completed A
permanently unclosable.
**Rows 14–15 are the truthfulness guards** (rev 4, made non-vacuous in rev 5): they fail if
either pre-mode summary ships still asserting that every manifest item lives in
`.backlogit/queue/` — an assertion Probe 25 **measured false** for **all 13** members of a
completed manifest (`MEMBERS_QUEUE_LOCATED=0`).
**Row 16 is the totality guard**: it fails if a1's selector ships without a specified
`n > 1` branch or without the S0 anomaly gate.
**Rows 17–18 are the preservation guards** (rev 5).

> **Rev 3 note (superseded by rev 4, and again by rev 5).** Rev 3 recorded the row count as
> **unchanged at 13**. **Rev 4** added rows 14–16 (→ 16). **Rev 5** adds rows 17–18 (→ **18**)
> and makes rows 14–15 and 7/10 **compound** with verbatim-pinned literals. Rows 1–13 keep
> their numbering and assertion text, so rev 2/rev 3 traceability is intact.

**Helper budget.** Still **one** table-driven test function and **≤4 helper functions**
(§3). The helpers are: `repoRoot` *(reused, not redefined)*, `mustContain`, `mustNotContain`,
and `indexOfUnique`. Compound rows call `mustContain` **and** `mustNotContain` on the same
row — **two calls, not a new helper**. Ordering rows call `indexOfUnique`. **Helper count is
unchanged at ≤4**, and the function count is unchanged at **1**, so Ship Step 4.3's
one-function constraint still holds.

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
| **0a** | **Stage** | **PROBE 25 — root-included cascade fixture (§9.4). DISCHARGED.** Authored, executed and **committed** by Stage **before `021-S` is claimed**. `PROBE25_RESULT=PASS`, 8/8 criteria. **`021-S` is not claimable until this evidence is committed and PASSing.** Ship never authors this artifact (P-010) |
| 1 | S1 | Verify on `main`; Step 0.5 pre-claim topology gate; **claim `021-S`**. Ship moves **`022.001-T -> active`**. `022-F` is left **`active`** — **measured, not assumed** (Probe 22 ARM AB) |
| 2 | S1 | H0 red → implement → H1 green |
| 3 | S1 | **Step 4.3 quality gates; Step 4.4 review gate** |
| 4 | S1 | **Step 4.5 Complete Task — commit, then `022.001-T -> done`.** *(Installed order: Step 4.5 (L517) precedes Step 5 (L554). The task reaches `done` BEFORE the implementation PR merges.)* |
| 5 | S1 | **Step 5 PR lifecycle** — full gate sequence, `--phase lifecycle` topology gate, build, push, PR, operator approval, **merge** |
| 6 | S1 | **Step 6 Merge Confirmation Gate** — `gh pr view` state `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | **Reload merged `main`.** Detect that the merged change alters **Ship's own Role Boundary**. Per §6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (§9.1 payload contract: `schema_version: 1`, `agent: ship`, **`session_id`**, `phase: awaiting-fresh-session-parent-completion`, `resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA; domain data under `context`) and **END**. **Record the emitted checkpoint filename** — S2 must match it exactly |
| 9 | **S2** | **Fresh session.** Full checkpoint-recovery protocol (§9.1) — enumeration, anomaly gate, **S2-specific fail-closed zero-candidate rule**, exact-filename match, explicit owner selection, `agent: ship` ownership, operator confirmation, **restore → prune/gate → resume** (Engram-reachability fail-closed; never-prune allowlist honored), same cursor, resolve-after-confirmed-resume. Loads the new Role Boundary + Step 6.1(a1) from merged `main` |
| 10 | S2 | Verify the merge SHA is an ancestor of `origin/main`; validate the **same** active shipment `021-S` |
| 11 | S2 | **Step 6.0 Post-Merge Branch Protocol** — `git checkout main`, `git pull`, `git checkout -b post-merge/022-ship-feature-completion`. **Created BEFORE any post-merge backlog mutation**, because Step 6.1(e) commits `.backlogit/` and those commits must not land on `main` |
| 12 | S2 | **Step 6.1(a0)** `--phase lifecycle` topology gate — run **while `021-S` is still `active`**. *(Probe 20 tripwire: post-archive this gate fails closed on every route; it must never be re-run after the close.)* |
| **12a** | **S2** | **READ-ONLY RE-VERIFICATION of Stage's committed Probe-25 evidence (§9.4a).** All five checks: exact topology, PASS fields, evidence commit **ancestry** of the merge SHA, **currently-resolved** engine digest, and no stale/missing record. **S2 authors, modifies and re-runs NOTHING.** FAIL ⇒ HALT with the §9.4a recovery path; `021-S` stays `active`, no a1 mutation |
| 13 | S2 | **Step 6.1(a1)** — §5.1 step **S4** (`n == 1`, status `active`): `022-F` is a manifest member and all five §5 conditions hold → `022-F active -> done` |
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

0. **Know the expected checkpoint before enumerating. (rev 5 — plan-review findings A-8/A-9.)**
   S1's step-8 checkpoint is written through `backlogit_create_checkpoint` (MCP) **or**
   `backlogit checkpoint create --state-dump` (CLI) — **both are valid**; the payload MUST
   carry, **top-level**, all of `schema_version: 1`, **`agent: ship`**, **`session_id`**,
   `phase`, and `resume_hint`, with domain data nested under `context`.
   **`session_id` is REQUIRED top-level** by the installed Checkpoint Payload Contract; rev 4
   omitted it, so a plan-following agent would have written a **non-conforming** checkpoint
   that step 2's own anomaly gate then rejects. S1 **records the emitted filename**
   (`checkpoint-YYYYMMDD-HHMMSS.json`) and the `session_id` in its handoff. **S2 carries both
   as the expected identity** and matches them **exactly**.
1. **Enumerate** via `backlogit_list_checkpoints` with `consumer_id: "ship"` and **no**
   `status`/`agent` filter — a quarantined or schema-invalid record must not be silently
   excluded by the query itself.
2. **Run the anomaly gate FIRST** — any validation error, quarantine flag, or
   missing/malformed required field in **any** enumerated summary ⇒ **FAIL CLOSED** to
   operator handoff, evaluated **before** the candidate-count check.
2a. **S2-SPECIFIC CANDIDATE RULE — FAIL CLOSED on zero, multiple, or mismatch. (rev 5.)**
   The installed protocol treats **zero** active `ship` candidates as *normal startup —
   explicitly not a failure*. **That continuation does NOT apply to this handoff**, and an
   agent that applies it would skip parent completion entirely and "proceed with fresh work".
   For S2, whose entire purpose presupposes S1's checkpoint:
   * **zero** active `ship` candidates ⇒ **FAIL CLOSED** to operator handoff. There is **no
     normal zero-candidate path at this step**;
   * **more than one** active `ship` candidate ⇒ **FAIL CLOSED** (ambiguity is never resolved
     by picking);
   * a candidate whose **filename** does not match the exact filename S1 emitted, or whose
     **`session_id`** does not match S1's recorded `session_id` ⇒ **FAIL CLOSED**.

   Exactly **one** candidate, matching **both** the expected filename **and** `session_id`,
   may proceed. Anything else halts to the operator.
3. **Require explicit owner selection** — never auto-pick, **even when exactly one
   candidate is returned** and even when it matches. The operator selects one checkpoint by
   filename; ambiguity fails closed.
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
| 4 | **Step 6.1(a1)** — §5.1 step **S4** (`n == 1`, status `active`): `018-F` is a present feature member; conditions hold; `018-F active -> done`. The 11 pre-archived descendants are exempt from condition 1 **only** under the §5.0 enumerated-ID disposition |
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

### 9.4 Root-included cascade fixture — STAGE PRE-CLAIM EVIDENCE (rev 5, hybrid)

> **Rev 5 — this section is RE-SEATED. Rev 4's placement is WITHDRAWN.**
> Rev 4 made this an S2-run, no-override gate at §9 step 12a — *after* A's Role Boundary
> change had already merged to `main` at step 5. The attempt-1 review found two blocking
> defects in that placement, and both are structural rather than editorial:
>
> * **A-1 (P0)** — a **no-override** gate placed **after an irreversible merge** creates an
>   **unrecoverable stall**: on FAIL the contract change is already deployed to every future
>   Ship session while `021-S` is stuck `active`/in-flight, and no in-role actor can clear it
>   (Stage may not close shipments; Ship is halted by the gate). It also contradicted this
>   plan's own §10.2 `git revert` rollback — precisely the override §9.4 forbade.
> * **A-2 (P1)** — it assigned **authoring and committing** `docs/plans/evidence/probe25-*`
>   to **Ship**, which Ship's Role Boundary **forbids** under **P-010**. Executed literally,
>   S2 would record a P-010 violation and halt, and **A could never close**.
>
> **The hybrid resolves both.** **Stage authors, executes and commits the evidence BEFORE
> `021-S` is claimed. Ship re-verifies it READ-ONLY.** Nothing has merged when the
> measurement happens, so a FAIL costs a re-plan and nothing else; and Ship never touches a
> planning artifact, so P-010 is never engaged.

**Ownership — unambiguous.**

| Actor | Does | Never does |
|---|---|---|
| **Stage** (§9 step **0a**, pre-claim) | authors, executes and **commits** `probe25-*.{ps1,txt}` | — |
| **Ship S2** (§9 step **12a**, post-merge) | **read-only re-verification** of the committed evidence (§9.4a) | author, modify, re-run or regenerate planning evidence |

**STATUS: EXECUTED AND COMMITTED.** Stage ran Probe 25 at `4fd21c5`, **before** `021-S` is
claimable. `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, all eight criteria PASS.

**Gate definition.**

| Property | Value |
|---|---|
| **Who runs it** | **Stage**, at §9 step **0a** — pre-claim, pre-merge, pre-implementation |
| **When** | **Before `021-S` is claimed.** A plan/PR readiness gate: `021-S` is **not claimable** until this evidence is committed and PASSing |
| **Fixture shape** | The **exact** `017-S` shape: a **13-member** manifest = **root feature included** + 12 descendants, of which **11 are genuinely archived** and 1 is a live task. Not an approximation, not the Probe-15 root-excluded inverse |
| **Workspace** | A **disposable, isolated, live-config-seeded** backlogit workspace — **never** the live one. Same pattern as Probes 18/20/22/24 |
| **Engine binding** | **Portable identity check.** The engine is resolved **at runtime** from the **registered command** (`Get-Command backlogit`); its SHA-256 is compared to the **committed digest** `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`. **No runtime requirement hardcodes an absolute path.** The evidence **records** the observed path as telemetry. **Mismatch ⇒ HALT** |
| **Evidence location** | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.ps1` **and** `…/probe25-root-included-cascade-fixture.txt` — **committed**, alongside probes 14–24 |
| **Authorization** | **CLI-only**, matching probes 18–24. No MCP equivalence is claimed |

> **Rev 5 — portability (plan-review finding A-3), and the Hardening 5 reconciliation.**
> Rev 4 pinned the gate to the absolute path `C:\Tools\backlogit.exe`, so any second
> operator, CI runner or toolchain move produced `DIGEST_GATE=FAIL` → HALT → (under rev 4's
> no-override rule) a bricked A. **That path binding is removed from the contract.** What
> the contract requires is an **identity** assertion, not a **location** assertion:
>
> * **resolve** `backlogit` at runtime via the registered command / `Get-Command`;
> * **compare** the resolved executable's SHA-256 to the committed digest;
> * **record** the observed path (`ENGINE_PATH_OBSERVED`) as evidence telemetry only;
> * on **mismatch**, **HALT** — the measurement cannot be attributed to a known engine build.
>
> A HALT here is now cheap and recoverable: it happens **pre-claim**, so it returns the work
> to Stage with nothing merged and nothing in flight. **Hardening 5 is reconciled**
> accordingly below: A's *execution* path still introduces no digest-bound engine call; A's
> *verification* path is digest-gated, and that gate now sits **before** the irreversible
> step rather than after it.

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

**PASS criteria — all of them, conjunctive. MEASURED RESULT: 8/8 PASS.**

| # | Criterion | Measured |
|---|---|---|
| 1 | `DIGEST_GATE=PASS` against the committed digest, engine resolved via registered command | **PASS** |
| 2 | pre-mode with `expected_status: done` returns **`PROCEED`**, with the 11 archived members classified **`pre-archived`**, the 2 completed members **`matched`**, and **no** `status-mismatch`, `missing` or torn/duplicate record | **PASS** |
| 3 | classification returns **`CASCADE`** (not `SAFE_CLOSE`, not `HALT`) | **PASS** |
| 4 | the cascade close returns **`returned_ids: []`** — empty | **PASS** |
| 5a | **`archived_ids − allowed_ids` is EMPTY** — nothing unexpected was archived | **PASS** |
| 5b | **`required_ids − archived_ids` is EMPTY** — nothing required was left unarchived | **PASS** |
| 6 | every surviving record's **`parent_id` is preserved** | **PASS** |
| 7 | the shipment record reaches **`shipped`** and archives with `archived_status: shipped` | **PASS** |
| 8 | **nothing outside the manifest** is archived, moved or deleted | **PASS** |

Criteria **5a and 5b are two separately-labelled, independently-failing conditions**
(P-015 amendment 1.21.0), never a single reconciliation. `allowed_ids` is the manifest
**plus the shipment record itself** — the close archives its own record and the engine
reports it. `archived_ids` is a **transition log, not a manifest echo**: a genuinely
`status: archived` member has no transition to report and is correctly absent, which is
exactly why full-set equality is the wrong test and why the 1.19.0 equality claim was
withdrawn. The probe evaluates 5a over the **union** of the observed pre/post snapshot log
**and the engine's own reported `archived_ids`**, so a mutation visible to one but not the
other cannot pass vacuously.

**Measured highlights** (full transcript in the committed `.txt`):

```text
MEMBERS_QUEUE_LOCATED=0          MEMBERS_ARCHIVE_LOCATED=13
C7_C8_QUEUE_ONLY_SUMMARY_FALSIFIED=True
PREMODE_MATCHED=2  PREMODE_PRE_ARCHIVED=11  PREMODE_RESULT=PROCEED
CLASSIFICATION_VERDICT=CASCADE
archived_ids=[001.004-T, 001-F, 001-S]   returned_ids=[]
SET_EQ_A__archived_minus_allowed=EMPTY   SET_EQ_B__required_minus_archived=EMPTY
LIVE_BACKLOG_UNMUTATED=True              PROBE25_RESULT=PASS
```

> **An unplanned but load-bearing measurement: `MEMBERS_QUEUE_LOCATED=0`.** At the pre-mode
> gate, **zero of the 13 members are in `.backlogit/queue/`** — the live registry routes
> `done|accepted|rejected|archived` to `archive/`, so even the two *correctly completed*
> members leave the queue. This **independently falsifies the queue-only pre-mode summaries**
> at C7 and C8 by measurement rather than by argument, and it strengthens §4.0: the stale
> sentences are wrong not only for the 11 legacy members but for **every** member of a
> completed manifest.

**Failure semantics — fail closed, and RECOVERABLE (rev 5).**

* **Pre-claim (Stage, the normal case).** A FAIL, or a digest mismatch, **HALTs before
  `021-S` is claimed**. Nothing has merged, nothing is in flight, and the cost is a return
  to Stage for re-planning. `021-S` stays **unclaimed**; `017-S → 021-S` stays unsatisfied.
  **This is the gate that actually protects the route**, and it has already run: **PASS**.
* **Evidence missing or not committed ⇒ `021-S` is NOT CLAIMABLE.** This is a plan/PR
  readiness gate, not a closure-time gate.
* **Post-merge (Ship S2 re-verification) — see §9.4a for the recovery path.** Rev 4's
  *"no operator override, no `--force`, no proceed-with-residual"* clause is **WITHDRAWN**:
  as written it contradicted §10.2's `git revert` rollback and left no in-role actor able
  to clear a stalled `021-S`.

#### 9.4a Ship S2 re-verification — READ-ONLY (rev 5, normative)

At §9 step **12a**, the fresh S2 session performs a **read-only** re-verification of the
**already-committed** evidence. **S2 authors nothing, modifies nothing, and re-runs nothing.**

**What S2 checks — all five, conjunctive:**

1. **Exact topology.** The evidence records `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`,
   `PRE_ARCHIVED=11`, `ARCHIVED_STATUS_FLAVOUR=queued`. **An exact-match check** — any other
   topology means the measurement is not about `017-S`'s shape.
2. **PASS fields.** `PROBE25_RESULT=PASS` and `FAILED_CRITERIA=0`, and each of the eight
   criterion tokens reads `PASS`.
3. **Evidence commit ancestry.** Both evidence files are **present and tracked**, and the
   commit that introduced them is an **ancestor of the merge commit** S2 verified at §9
   step 6 (`git merge-base --is-ancestor {evidence_sha} {merge_sha}`). This is what makes
   "stale" machine-checkable: evidence must **pre-date** the merge, not be regenerated after.
4. **Current resolved engine digest.** S2 resolves `backlogit` **the same portable way**
   (registered command / `Get-Command`) and compares its **current** SHA-256 to the digest
   recorded in the evidence. This detects toolchain drift **between** measurement and close.
5. **No stale or missing record.** Neither evidence file is absent, empty, or modified
   relative to its committed blob (`git diff --quiet HEAD -- {paths}`).

**S2's role boundary is preserved absolutely.** Every one of the five is a **read**:
`git log`, `git merge-base`, `git diff --quiet`, `Get-FileHash`, and text inspection. S2
never writes to `docs/plans/`. This is what closes **A-2**.

**If S2 re-verification FAILS — the recovery path (closes A-1):**

1. **`021-S` remains `active`.** It is **not** marked `shipped`, **not** archived, and the
   a1 mutation is **not** performed. `017-S → 021-S` stays unsatisfied.
2. **S2 HALTs and returns to the operator**, recording which of the five checks failed.
3. **The operator has real, in-role options** — this is the override rev 4 wrongly forbade:
   * **re-verify** after resolving toolchain drift (check 4 is the common benign cause), or
   * direct **Stage** to re-run Probe 25 pre-claim and re-commit fresh evidence — Stage owns
     this artifact, so this is always available, or
   * **revert A** per §10.2, after holding `017-S` per §10.2's obligation.
4. **`021-S` is left claimable-and-active for a future Ship session**, not stranded: because
   no a1 mutation and no close occurred, a later S2 re-entry is a clean resume (§5.1 **S4**,
   or **S3** if a prior attempt had already completed the feature).

**The impossible contradiction is gone.** There is no longer a no-override clause sitting
against a rollback that requires an override, and no step assigns Ship a P-010-forbidden
write. The gate is still **non-bypassable** — S2 may not close A without a PASS — but a
failure is now **recoverable by a named actor** instead of deadlocked.

**What this gate does NOT claim.** It measures the **engine** on a faithful fixture; it is
not a proof about the live `017-S` records at the moment `017-S` is later claimed. The
live-record re-read obligations at `017-S`'s own gate are unchanged and still fail-closed —
the same relationship PO-1a/PO-1b have (§10.1).

#### 9.4b Inherited CASCADE residual — correctly classified (rev 5)

**Plan-review finding A-12** observed that the cascade's `returned_ids` / two-set /
`parent_id` checks run **after** the cascade has already mutated the backlog, so a
destructive mutation is *detected*, not *prevented*.

**This is a pre-existing P-015 residual, and A neither creates nor removes it.** Stating
that plainly matters, because the alternative — claiming Probe 25 prevents it — would be
an overclaim:

* **Not a new cascade path.** A introduces **no** verdict and **no** new close path. The
  `CASCADE` invocation A rides is the identical call every fully-covered-root closure
  already makes today. The detect-after-mutate ordering is a property of **P-015 and the
  engine**, inherited unchanged.
* **What Probe 25 actually does** is act as a **pre-claim gate**: it measures, on an exact
  topology match, that this cascade shape behaves as classified. **Measured: `returned_ids`
  empty, both set equations EMPTY, all `parent_id`s preserved, nothing outside the manifest
  touched, and the live backlog unmutated.**
* **What Probe 25 does NOT do.** It does **not** prevent a future mutation, does **not**
  install a rollback, and does **not** make the detect-after-mutate window smaller. If the
  live topology at closure differs from the measured topology, the fixture's PASS does not
  transfer — which is exactly why §9.4a check 1 is an **exact topology match that HALTs on
  mismatch**.
* **Rollback is not claimed.** Rev 4's implied restore-on-halt is **not** asserted here.
  Remediating the detect-after-mutate ordering (a pre-cascade restorable snapshot and an
  unconditional `git restore -- .backlogit/queue/ .backlogit/archive/` on every cascade HALT
  path) is a **recorded follow-up against P-015 and the skill** — **not charged to A**,
  whose 2-file surface touches neither.

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

* `active` → proceed (**S4**).
* `done` → proceed via **idempotent resume** (**S3**), and only after §5 conditions 1–4
  revalidate; if any fails, **HALT**.
* anything else → **HALT** (**S5**). Do not widen §5 condition 5 in-flight. Return to
  **Stage**.

> **Rev 5 — PO-1b reads by ID, and the tool is pinned (finding B-10/B-13).** `backlogit get
> 022-F` **or** the MCP `backlogit_get_item` are **both** acceptable — the load-bearing rule
> is *read by status/ID, never by path*, not CLI-exclusivity, and rev 4 mandated the CLI with
> no justification. The **a1 transition itself** is pinned: it MUST be performed with
> `backlogit_move_item` / `backlogit move {id} --status done` — **never** a direct file edit.
> Ship MUST NOT invoke `backlogit update --complexity`, which is advertised but fails (exit 1)
> in this workspace (finding B-12).

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

> **Rev 5 — the rev-4 contradiction is RESOLVED (plan-review finding A-1).** Rev 4's §9.4
> declared *"no operator override … no `--force`, and no 'proceed with a recorded residual'
> path"* while this section **requires an operator `git revert`** — precisely the override
> §9.4 forbade. Two clauses of the same plan directly contradicted each other, and the
> practical consequence was an **unrecoverable stall**. The no-override clause is
> **withdrawn** (§9.4). Rollback per this section is now an **explicitly available** operator
> action at every stage, and §9.4a lists it as one of three named recovery options.
>
> **Rollback also covers the evidence files (finding B-1).** `probe25-*.{ps1,txt}` are
> committed by **Stage pre-claim**, in a **separate commit** from A's 2-file implementation.
> Reverting A's merge commit therefore does **not** disturb them, and it does not need to:
> they are planning evidence, not implementation surface. If the plan is abandoned outright,
> Stage archives the evidence with the plan.

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
| a1: a **present** feature member fails any §5 condition 1–5 | **HALT** fail-closed (§5.1 **S4**); surface which condition; no close |
| a1: manifest has **zero** feature members | **Explicit no-op** (§5.1 **S1**) — record `A1_NOT_APPLICABLE` and proceed to `a`. **Not** a failure |
| a1: feature member already `done`, conditions 1–4 revalidate | **Idempotent resume** (§5.1 **S3**) — record `A1_ALREADY_DONE` and proceed to `a`. **Not** a failure |
| a1: feature member already `done`, but any of conditions 1–4 **fails** | **HALT** fail-closed (§5.1 **S3**) — a `done` feature over unfinished descendants is corruption, not a resume |
| a1: feature member in any other status (`queued`, `blocked`, `review`, …) | **HALT** fail-closed (§5.1 **S5**); record the observed status; return to Stage |
| a1: manifest has **more than one** feature member | **HALT** fail-closed (§5.1 **S2**) — record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`; evaluate **no** §5 condition and mutate **no** feature; return to Stage |
| **(rev 5)** a1: member anomaly — missing/unresolvable `artifact_type`, record in **both** queue and archive, record in **neither**, malformed archive provenance, or an archive-located member **not on the §5.0 disposition list** | **HALT** fail-closed (§5.1 **S0**) — evaluated **before** `n` is computed; no mutation |
| **(rev 5)** Probe 25 **FAILS** or digest-mismatches at Stage §9 step **0a** | **HALT pre-claim.** `021-S` is **not claimable**; nothing has merged and nothing is in flight. Return to Stage to re-plan. *(Already executed: `PROBE25_RESULT=PASS`.)* |
| **(rev 5)** S2 §9.4a **read-only re-verification** fails any of the five checks | **HALT.** `021-S` stays `active`, **no** a1 mutation, **not** shipped, **not** archived. Return to the **operator** with the §9.4a recovery path: re-verify after resolving toolchain drift, or have **Stage** re-run and re-commit Probe 25, or **revert A** per §10.2 (holding `017-S` first). `021-S` remains cleanly resumable |
| **(rev 5)** S2 checkpoint candidates: **zero**, **more than one**, or filename/`session_id` **mismatch** | **FAIL CLOSED** to operator handoff (§9.1 step **2a**). The installed "zero candidates is normal startup" continuation **does not apply to S2** |
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
  checkpoint MUST be written through **`backlogit_create_checkpoint` (MCP) *or*
  `backlogit checkpoint create --state-dump` (CLI) — both are valid (rev 5, C-9)** —
  tool-timestamped, with **top-level** `schema_version: 1`, `agent: ship`,
  **`session_id` (rev 5, A-8 — REQUIRED by the installed Checkpoint Payload Contract)**,
  a `phase`, and a `resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA,
  with domain data under `context`; its `created_at` must precede S2's resume. S1 records
  the **emitted filename** and the `session_id`; S2 matches **both exactly**.
  **S2 executes the full §9.1 recovery protocol — all twelve steps, in order.** AC-9 is
  **not** satisfied unless S2 demonstrably performed **every** element
  below; rev 3's shorter list is superseded and is **not** sufficient evidence:
  * enumeration via `backlogit_list_checkpoints` with `consumer_id: "ship"` and **no**
    `status`/`agent` filter;
  * **anomaly gate evaluated FIRST**, before the candidate-count check;
  * **the S2-SPECIFIC candidate rule (rev 5, A-9)** — **zero** active `ship` candidates is
    **FAIL CLOSED**, *not* the installed "normal startup" continuation; **more than one** is
    FAIL CLOSED; and a filename or `session_id` that does not match S1's recorded values is
    FAIL CLOSED. **There is no normal zero-candidate path at this handoff**;
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
- **AC-20** **(rev 5, RE-SEATED — supersedes rev 4's S2-run gate)** The §9.4 root-included cascade fixture (Probe 25) is **authored, executed and committed by STAGE at §9 step 0a, BEFORE `021-S` is claimed**. **`021-S` is not claimable** until the evidence is committed and PASSing. The evidence lives at `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.{ps1,txt}`, records `DIGEST_GATE=PASS` against SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` **with the engine resolved at runtime via the registered command (no hardcoded path)**, and satisfies **all eight** PASS criteria over the **exact 13-member root-included manifest with 11 genuinely pre-archived (`archived_status: queued`) members**. **STATUS: DISCHARGED** — `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, measured at `4fd21c5`. **Ship never authors, modifies, re-runs or regenerates this artifact (P-010).**
- **AC-20a** **(rev 5)** At §9 step 12a the fresh S2 session performs the §9.4a **read-only** re-verification and it passes: (1) exact topology `TOPOLOGY_MEMBERS=13` / `ROOT_INCLUDED=True` / `PRE_ARCHIVED=11` / `ARCHIVED_STATUS_FLAVOUR=queued`; (2) `PROBE25_RESULT=PASS` and `FAILED_CRITERIA=0` with all eight criterion tokens `PASS`; (3) the evidence commit is an **ancestor of the verified merge SHA**; (4) the **currently-resolved** engine SHA-256 matches the digest recorded in the evidence; (5) neither file is missing, empty, or modified relative to its committed blob. **Every check is a read.** On failure S2 HALTs, `021-S` stays `active` with **no** a1 mutation, and the operator has the §9.4a recovery path. AC-20a is **not** satisfied if S2 writes to `docs/plans/`.
- **AC-21** **(rev 5)** Step 6.1(a1)'s selector is the **strictly ordered, non-overlapping** sequence **S0 → S5** of §5.1, computing `n` by `artifact_type` and never by ID suffix: **S0** anomaly gate **before** `n`; **S1** `n == 0` ⇒ explicit no-op; **S2** `n > 1` ⇒ HALT with `A1_MULTIPLE_FEATURE_MEMBERS` and **no** feature mutation whatsoever; **S3** `n == 1` + `done` ⇒ idempotent no-op **only after conditions 1–4 revalidate**, else HALT; **S4** `n == 1` + `active` ⇒ full five-condition guard then `active -> done`; **S5** `n == 1` + any other status ⇒ HALT. **No input may satisfy two steps.** Go test row 16 passes.
- **AC-22** **(rev 4; assertions made NON-VACUOUS in rev 5)** Both stale queue-only pre-mode summaries are reconciled to the §4.4 authoritative semantics — **C7** at `_ship.agent.md` L255–256 (Step 0.5 item 6) and **C8** at L776–777 (Step 6.1(a)). Each reconciled sentence expresses **both** accepted dispositions: a **queue item at the expected status** *or* an **archive-located `pre-archived` member accepted without a status check**, **and** preserves the `PROCEED` conjuncts **no orphans** and **`record-consistent`**. Neither may remain queue-only. Go test rows **14** and **15** are **compound** — each asserts the §8.1.1 PRESENT phrase **and** the ABSENT exact old literal — so neither can pass while its site keeps the old text. The rev-3 instruction that *"a reviewer MUST NOT fail A for its absence"* is **withdrawn**.
- **AC-23** **(rev 5)** The inventory is **8 clause sites + 2 additive**, the recalculated budget is **~99 min** with a **~21 min** margin to the 2-hour rule (§12), and the Go test carries **18 rows** — of which **4 are compound** — in **one** function with **≤4 helpers**, using **stdlib assertions only** and **reusing** `repoRoot(t)`. No site is excluded from the inventory in order to preserve a previously-pinned budget.
- **AC-24** **(rev 5)** §5.0's pre-archived-descendant exemption is applied **only** to the enumerated **11 `017-S` IDs**, and **only** when all five §5.0 gates hold. `archived_status: queued` is **never** treated as proving completion. Any pre-archived manifest descendant not on the list ⇒ **HALT**. `021-S` has **none**.
- **AC-25** **(rev 5)** `_ship.agent.md` binds Ship to the **closed** verdict set `{CASCADE, SAFE_CLOSE, HALT}` (§4.1), and Ship **HALTs** on an absent, unparseable, unknown, ambiguous or policy-mismatched verdict, and on **P-015/skill disagreement**. No future or unversioned verdict is authorized by A.
- **AC-26** **(rev 5)** Rows **17** and **18** pass: C8's reconciled sentence preserves the **single-writer lock** and **orphan scan** clauses; C7's replacement leaves the adjacent `Scope note (139-F/139.001-T)` intact.

## Constitution Check

**(rev 5 — plan-review finding A-13.)** The constitution's Governance section requires every
implementation plan to map its work against Principles I–XI and to document justified
deviations. Rev 4 mapped exhaustively against *workflow policies* but never against the
*constitutional principles*, leaving genuinely justified deviations undocumented.

| Principle | Applies | Assessment |
|---|---|---|
| **I. Safety-First Go** | yes | The only Go artifact is one table-driven contract test. No `unsafe`, no goroutines, no I/O beyond reading one file. Errors surface via `t.Errorf`, never panics or ignored returns |
| **II. Test-First Development (NON-NEGOTIABLE)** | yes | **Honored literally.** §7 mandates H0 red → H1 green; the test lands **before** any `_ship.agent.md` edit. **Deviation (justified) — see D1 below** on scenario count |
| **III. Workspace Isolation and Security Boundaries** | yes | Probe 25 runs in a disposable, gitignored, workspace-contained directory seeded from live config; `FIXTURE_ISOLATED_FROM_LIVE=True` and `LIVE_BACKLOG_UNMUTATED=True` are **measured**, not asserted |
| **IV. CLI Workspace Containment (NON-NEGOTIABLE)** | yes | Every probe call is `--cwd $ws`. The engine is resolved via the **registered command**, not an absolute path (§9.4). No write escapes the repo |
| **V. Structured Observability** | yes | a1 emits named, greppable tokens (`A1_NOT_APPLICABLE`, `A1_ALREADY_DONE`, `A1_MULTIPLE_FEATURE_MEMBERS`); Probe 25 emits one `KEY=VALUE` line per criterion |
| **VI. Single Responsibility** | yes | A does exactly two things — grant one narrow authority, and replace a duplicated classification with delegation. §2.2's drift is the cost of the duplication this removes |
| **VII. Destructive Command Approval (NON-NEGOTIABLE)** | yes | The cascade is destructive. It is invoked **only** on a machine-selected `CASCADE` verdict, only inside the disposable fixture for measurement, and never against the live backlog by this plan. Ship's live cascade remains gated by P-015 |
| **VIII. Explicit Safety Modes for Elevated Risk** | yes | **Deviation (justified) — see D4 below** |
| **IX. Git-Friendly Persistence** | yes | All artifacts are line-oriented Markdown/PowerShell/Go. The backlog mutation at 6.1(e) is committed on the closure branch, never on `main` |
| **X. Agent Context Efficiency** | yes | A **removes** ~20 lines of re-derived classification prose from `_ship.agent.md` and replaces them with a delegation pointer — a net context reduction on Ship's hot path |
| **XI. Merge Commit History Preservation (NON-NEGOTIABLE)** | yes | §10.2 rollback is `git revert` of the merge commit — never a history rewrite, never a force-push. No amend anywhere in this plan |

**Documented deviations — justified, not waived.**

* **D1 — Task granularity: 18 test scenarios vs *"fewer than 4"*.** The granularity heuristic
  targets *behavioral* scenarios. These 18 are **rows in one table-driven function** over a
  **single** file, sharing one fixture and ≤4 helpers; each is a substring or index assertion,
  not an independent scenario. Collapsing them would **reduce** falsifiability — and rev 4's
  attempt to economize is exactly what produced the vacuous rows 14–15 (finding A-7).
  **Justification: assertion density, not scope creep.** The §12 budget carries them at ~29 min.
* **D2 — *"fewer than 5 functions"*: 1 test function + ≤4 helpers = 5 symbols.** The
  constraint is read as **one function under test per task**; helpers are shared machinery.
  This is also a hard Ship constraint (Step 4.3's one-function batch rule, the `021.001-T`
  deadlock), so **1 function is mandatory** and the helpers cannot be split into a second
  function without breaking it. **Justification: an externally imposed constraint.**
* **D3 — `tests/integration/` rather than an inline `_test.go` or `tests/contract/`.** The
  instructions' taxonomy names `tests/contract/`, which **does not exist** in this repo. The
  established precedent for text-contract tests here is `tests/integration/`
  (`build_script_test.go`), whose package-level `repoRoot(t)` this test **reuses**.
  **Justification: existing precedent; creating a new top-level test tree is out of scope.**
* **D4 — Principle VIII: a global-closure blast radius without an explicit safety-mode
  directive** *(finding C-1)*. A changes closure behavior for **every** future shipment.
  **Accepted with compensating controls rather than a mode declaration**: the change
  *removes* authority duplication rather than adding reach; four negative widening guards
  (rows 7–10) fail if anything new is authorized; the new grant is bounded by five
  conjunctive conditions, one position, one scope and one transition; and A closes by the
  **pre-existing** path, so a defect in the new wording cannot corrupt A's own closure.
  **Recorded as a residual**, not as a satisfied principle.

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
| Go table-driven test (**18** rows, 4 compound) + ≤4 helpers | ~29 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~99 min ≈ 1.65 h** |

Rates are rev 6 §10.1's own, so this is directly comparable to the measurement that
produced `MUST_REPLAN`. Margin to the 2-hour rule: **~21 min**.

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

> **The margin is smaller and that is the honest number.** ~21 min of headroom against the
> 2-hour rule is still real headroom, and §12's contingency is unchanged: if actual work
> exceeds 2 h mid-execution, **HALT and return to Stage**. Rev 3 protected a ~42 min margin
> by leaving two defects out of the inventory; rev 4 spent 14 of those minutes fixing them
> and rev 5 spends 7 more making the resulting assertions **actually falsifiable**. Trading
> margin for completeness is the correct direction — an under-scoped inventory does not make
> the work smaller, it only moves the discovery later.

> **Rev-5 delta: +7 min. Total ~92 → ~99 min; margin ~28 → ~21 min.** Go test **+5** for
> rows 17–18 and for making rows 7, 10, 14 and 15 **compound** with verbatim-pinned literals
> (§8.1.1) — the rev-4 versions were vacuously satisfiable, so this is the cost of the
> assertions being real. ADD-2 **+2** for the §5.1 ordered S0–S5 selector and the §5.0
> pre-archived-descendant gate. **Helper count unchanged at ≤4** and function count unchanged
> at **1**, so Ship Step 4.3's one-function constraint still holds.

> **Not charged to the task budget: Probe 25 (§9.4).** Probe 25 is run by **Stage**, at §9
> step **0a**, **pre-claim** — it is a Stage planning-time measurement against a disposable
> workspace, outside the implementation task's 2-hour envelope. **It has already been
> executed and committed** (`PROBE25_RESULT=PASS`). S2's §9.4a obligation is a **read-only
> re-verification** of that committed evidence — five read operations, ~3 min of S2 session
> time, and likewise not part of the harnessed task. Neither consumes the ~21 min margin.

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

> **Rev 5 correction — A-3 is now CLOSED, and the reconciliation is stable.** §9.4's gate
> does bind a digest, and that remains true. What changed is **where the gate sits** and
> **how the engine is identified**:
>
> * **A's own execution path** introduces no digest-bound engine call — the paragraph above
>   remains true of the *implementation* and of the `CASCADE` invocation itself.
> * **A's verification path** is digest-gated, so a measurement cannot be silently
>   attributed to a different engine build. The gate now runs **pre-claim** (Stage, §9 step
>   0a), so a mismatch costs a re-plan, not a bricked shipment.
> * **The machine-specific binding is REMOVED.** The contract resolves `backlogit` at runtime
>   from the **registered command** and compares **SHA-256 to the committed digest**. The
>   absolute path `C:\Tools\backlogit.exe` is **no longer a requirement of any runtime check**
>   — the evidence merely *records* the observed path as telemetry
>   (`ENGINE_PATH_OBSERVED`). A second operator, a CI runner, or a relocated toolchain
>   therefore passes, provided the binary is the same build.
> * **A mismatch HALTs**, and that HALT is now recoverable: pre-claim it returns to Stage;
>   at S2 re-verification it follows the §9.4a operator recovery path. The rev-4
>   no-override clause that made a mismatch unrecoverable is **withdrawn**.
>
> **Residual, recorded**: a point-release upgrade of backlogit still requires Stage to re-run
> and re-commit Probe 25 so the digest matches. That is a cheap, in-role, pre-claim action —
> not a blocker, and no longer a way to strand A.

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

## 13. Independent re-review scope (rev 5)

**All prior review scopes are SUPERSEDED**, including rev 4's §13 and every
three-shipment-era scope (rev 5 §13, rev 6 §13). Review **rev 5 of this plan only**;
the deferred B/C plans are **not** in scope and must not be routed from.

| # | Item | Why it is in scope |
|---|---|---|
| 1 | **Adjudicate the §9.4/§9.4a HYBRID.** Is Stage-authors / Ship-verifies-read-only the correct resolution of A-1 (post-merge no-override deadlock) and A-2 (P-010)? Does §9.4a's five-check re-verification remain genuinely non-bypassable while being purely read-only? Is the recovery path complete and in-role? | The structural change of rev 5; closes the only P0 |
| 2 | **Verify the committed Probe-25 evidence.** Does the transcript actually establish the 8 criteria on the exact 13-member root-included shape with `archived_status: queued`? Is the two-set gate evaluated over the engine's own `archived_ids`, not vacuously? | The plan's central empirical claim |
| 3 | **Audit §5.1's S0–S5 ordering.** Is it genuinely total *and* disjoint over (`n`, status)? Can any input still reach two steps? Is S0 correctly placed before `n` is computed? | Closes A-6; rev-4's table was ambiguous |
| 4 | **Audit §5.0's pre-archived exemption.** Is the five-gate definition fail-closed? Is the 11-ID enumeration correct for `017-S`? Is the general authority narrow enough that it cannot be read as "archived descendants don't count"? | Closes A-10 — the completion-proof gap |
| 5 | **Adjudicate §8.1.1's pinned literals.** Are the four ABSENT literals verbatim-correct against the installed file? Can any of rows 7/10/14/15 still pass vacuously or be green at H0? Do rows 17–18 actually protect the surrounding clauses? | Closes A-7; rows 14–15 are the whole justification for the inventory reopening |
| 6 | **Adjudicate the recalculated budget** (~99 min, ~21 min margin) and the ≤4-helper / 1-function claim at 18 rows | The rev-6 `MUST_REPLAN` verdict was a budget verdict; this is the same axis |
| 7 | **Audit §4.1's delegation trust boundary.** Does pinning the closed verdict set in Ship's own file contradict §4.1's "do not enumerate verdicts" constraint? Is the P-015/skill-disagreement HALT correct? | Closes A-11 without re-introducing the duplication A removes |
| 8 | **Audit §9.1's S2 candidate rule and payload.** Is `session_id` top-level correct per the installed contract? Is overriding the "zero candidates is normal startup" continuation safe and correctly scoped to S2 only? | Closes A-8/A-9 |
| 9 | **Verify §9.4b's residual classification.** Is detect-after-mutate correctly attributed to inherited P-015 rather than to A? Is the plan free of prevention/rollback overclaims? | Closes A-12 honestly rather than by assertion |
| 10 | **Audit the `## Constitution Check`.** Are D1–D4 genuine justifications rather than waivers? Is any principle mis-assessed? | Closes A-13 |
| 11 | **Confirm A still authorizes no new verdict** and that the §6 authority-escalation property survives rev 5's larger surface | Widening guard (rows 7–10); core self-hosting property |
| 12 | **Adjudicate the §R16 correction** in `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` — is R16.3 still the *sole operative* manifest contract, with withdrawn text clearly non-governing? | Governing-contract coherence; a contradiction there re-blocks `017-S`. **Read-only coherence check** (C-8) |
| 13 | **Confirm the operator-only PR boundary survives**, and that the fail-closed stance on the contradictory Orchestrator Step 1.5 (3a/3e) is correctly recorded **without** modifying the Orchestrator in this cycle | P-010/role-separation boundary. **Read-only coherence check** (C-8) |

**Out of scope, explicitly**: the deferred B/C plans; `shipment-reconcile/SKILL.md`
(including its stale `src/autoharness/...` classifier paths — recorded as a follow-up, not
a rev-5 edit); Orchestrator implementation repair (deferred to `019.004-T`); the
pre-existing `SAFE_CLOSE`/exit-9 tool conflict, which remains off-route and recorded; and
the P-015 detect-after-mutate ordering (§9.4b), which is an inherited residual recorded as a
follow-up against P-015 and the skill, not against A.

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

## Plan Review — attempt 2

<!-- plan-review-attempt: 2 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

- **Reviewed revision**: rev 5
- **Reviewed at commit**: `4fd21c5` (working tree, pre-commit)
- **Gate run**: attempt 2 (remediation cycle 2 of max 3)
- **Plan hardening required**: **yes** (declared in the header). **Present**: yes — `## Plan
  Hardening` with 9 items. **Materially complete**: **NO** — Hardening 9 still asserts the
  withdrawn no-override stance (see B-4 below).
- **Verdict**: **FAIL** — gate rule *"Any P0 or P1 findings ⇒ FAIL"*. `status:` stays
  `planned`. **`021-S` must NOT be claimed.**

### Dispatch capability (P-012)

| Capability | Status |
|---|---|
| reviewer-subagent-dispatch | `TOOL_OK` |
| model-specific-review-routing | `TOOL_OK` — anchor route `model_routing.anchor_review` = `openai` / `gpt-5.6-sol` / `high` dispatched |
| `agent-native-parity-reviewer` identity file | `TOOL_DEGRADED: agent-native-parity-reviewer-identity — declared fallback: rubric applied via dispatched subagent on a cross-model route.` Persona **covered, not skipped** |

### Persona coverage

| Persona | Mode | Model/route | P0/P1/P2/P3 |
|---|---|---|---|
| Constitution Reviewer | subagent | `claude-opus-4.8` | 0 / 1 / 2 / 2 |
| Go Reviewer | subagent | `claude-opus-4.8` | 0 / 0 / 2 / 2 |
| Scope Boundary Auditor | subagent | `claude-opus-4.8` | 0 / 0 / 1 / 3 |
| Learnings Researcher | subagent | default | 1 / 2 / 2 / 2 |
| Architecture Strategist | subagent — **ANCHOR** | `gpt-5.6-sol` / high | 1 / 4 / 3 / 0 |
| Security Lens Reviewer | subagent (cross-model) | `gpt-5.5` / high | 0 / 0 / 0 / 0 |
| Agent-Native Parity Reviewer | subagent (cross-model), degraded identity | `grok-4.6` / high | 0 / 1 / 1 / 3 |

All seven selected personas were covered. No persona was skipped.

### Gate decision rationale

**Rev 5 closed every attempt-1 blocker it targeted.** A-1, A-2, A-3, A-6 (partially),
A-7, A-8, A-9, A-10, A-11, A-12 and A-13 are affirmatively cleared below by the personas
that originally filed them — the **Security Lens Reviewer returned a clean sheet (0/0/0/0)**,
explicitly clearing all four of its attempt-1 P1s.

The plan nonetheless **FAILS**, on a defect **rev 5 itself introduced**. Two personas —
the **anchor** and the **Learnings Researcher** — independently converged on the same P0,
and **the plan's own committed Probe-25 transcript proves it**. That convergence, with
mechanical evidence, is the strongest signal in this review.

### P0 findings — blocking

**B-1 · §5.1's new S0 anomaly gate HALTs on A's own manifest — `021-S` becomes unclosable.**
*(Architecture Strategist (anchor) P0; independently corroborated by Learnings Researcher
P0-1; **confirmed by the plan's own §9.4 evidence**.)*
Sections: §5.0 gate 2, §5.1 **S0**, §9 steps 4 and 13.
S0 HALTs on *"an archive-located member that is not on the §5.0 recorded Stage disposition
list"* — a list containing only the eleven `018.*` IDs. But §9 step 4 moves **`022.001-T ->
done`** in S1, and the live `registry.yaml` routes `done|accepted|rejected|archived` to
`archive/`. By the time a1 runs at step 13, **`022.001-T` is archive-located and is not on
the list** ⇒ **S0 HALTs before `n` is computed**, `022-F` never reaches `done`, pre-mode
never returns `PROCEED`, and `021-S` stalls `active` — **after** A's irreversible Role
Boundary change merged at step 5. A second S0 bullet fires independently: a registry
relocation on `done` is not a close-archive, so `archived_status`/`archived_from` are absent
⇒ *"malformed archive provenance"* ⇒ HALT.
**This reintroduces the attempt-1 A-1 unrecoverable-stall shape through a different door**,
and it falsifies §5.0's claim that *"`021-S` has NO pre-archived descendants"* — true today,
**false at the moment a1 actually runs**.
**The plan's own evidence proves it**: `probe25-…txt` records `MEMBERS_QUEUE_LOCATED=0` and
`001.004-T loc=archive status=done`.
**Fix**: Re-base §5.0 gate 2 and every S0 archive-location trigger on **declared `status`**,
never on file location — the authoritative rule the `shipment-reconcile` skill already
states (*"never inferred from … which of `queue/`/`archive/` currently holds the record"*)
and which `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` records. Apply
the §5.0 list and provenance gates **only** to records declaring `status: archived`; accept
a uniquely-located record declaring `status: done` as a normal completed member.

### P1 findings — blocking

**B-2 · §4.4's classification table is location-first and contradicts the plan's own measurement.**
*(Learnings Researcher P1-1.)* §4.4 defines `matched` as *"**Queue file present** AND declared
status matches"*. The committed transcript reports, in one block, `MEMBERS_QUEUE_LOCATED=0`
**and** `PREMODE_MATCHED=2 PREMODE_PRE_ARCHIVED=11`. Under §4.4 those cannot both hold (zero
queue-located ⇒ `MATCHED=0`, `PRE_ARCHIVED=13`). §4.4 is the **normative source for C7/C8**,
and §8.1.1 pins row 15's PRESENT literal to that locational wording — **encoding the error
permanently into Ship's instruction surface** and letting rows 14–15 go green on wording that
is still false. That is exactly the *"new inaccuracy replacing the old one"* §4.4's own rev-5
note warns against.
**Fix**: Restate §4.4 from **declared status**; reconcile the transcript narrative; re-pin the
row 14/15 PRESENT literals.

**B-3 · §4.1/AC-25 require `_ship.agent.md` to enumerate the verdict set that AC-5/row 9 require it NOT to enumerate.**
*(Architecture Strategist P1; corroborated by Scope Boundary F-4 (P3) and Go P2-1.)*
AC-5 and Go test row 9 assert the file *"does **not** enumerate the authorized verdict set"*;
rev-5's AC-25 requires it to *"bind Ship to the **closed** verdict set `{CASCADE, SAFE_CLOSE,
HALT}`"*. **Both cannot pass under an honest test.** It also conflates `HALT` — a control
outcome with no close mutation — with the two close-path classification values.
**Fix**: Keep the no-enumeration rule in Ship's file and express the trust boundary as
**behavior** (HALT on any verdict the classification did not name; HALT on absent/ambiguous/
policy-mismatched results; HALT on P-015/skill disagreement), holding the closed set in
**this plan** as a review-time constraint. Pin row 9's needle to re-derivation phrasing, not
bare verdict tokens.

**B-4 · Hardening 9 still asserts the withdrawn no-override stance; `evidence_sha` is ambiguous.**
*(Architecture Strategist P1.)* `## Plan Hardening` item 9 still reads *"The gate deliberately
carries **no override path**"* — directly contradicting §9.4a and §10.2, which rev 5 rewrote
to **provide** one. Separately, §9.4a check 3 says the commit that *"introduced"* the evidence
must pre-date the merge, while check 5 compares current files only to `HEAD`: a post-merge
commit could **replace** the evidence while the introducing commit remains an ancestor.
**Fix**: Delete/supersede Hardening 9's no-override language. Define `evidence_sha` as the
**last** commit affecting both evidence files and require current blobs to equal that commit's.

**B-5 · §5.0's exemption has a circular dependency on pre-mode.**
*(Architecture Strategist P1.)* §5.0 gate 1 requires *"P-015 pre-mode classifies it
`pre-archived`"*, taken from the authoritative scan rather than re-derived — but §9 runs a1 at
step **13** and `shipment-reconcile mode: pre` at step **14**. Pre-mode cannot have run, and an
earlier intake scan does not satisfy the same-snapshot/immutability requirement. The `017-S`
route has **no defined provider** for a load-bearing a1 input.
**Fix**: Remove the cycle — define the exemption from the Stage disposition plus explicit
declared-status/location/provenance reads, reserving the authoritative pre-mode call for its
existing post-a1 gate.

**B-6 · The cascade reachability claim is still an overclaim, and §9.3/§9.4b/§10.3 disagree.**
*(Architecture Strategist P1.)* §9.4b's *"A neither creates nor removes it"* and *"not a new
cascade path"* are true at the **implementation-definition** level but not at the
**reachability** level: A is precisely what makes the destructive path reachable for `017-S`.
Meanwhile §9.3 still claims a HALT *"falls back to `SAFE_CLOSE`"* at *"no data cost"*, §9.4b
disclaims rollback, and §10.3 prescribes `git restore` — and **P-015 forbids
CASCADE→SAFE_CLOSE substitution once CASCADE is selected**.
**Fix**: State that A introduces new **reachability** to an inherited destructive path and
inherits its detect-after-mutate risk; delete the `SAFE_CLOSE`-fallback and "no data cost"
claims; reconcile §9.4b with §10.3 on which HALT paths carry a mandatory verified restore.

**B-7 · Backlog records `022-F` / `022.001-T` are a rev-4 split-brain against rev 5.**
*(Constitution Reviewer P1; corroborated by Learnings Researcher P1-2.)* Both records carry a
`=== REV 4 RECONCILIATION … THIS BLOCK WINS ===` block stating **16 rows**, **~92 min**,
**AC-1..AC-23**, a **four-way** selector, and *"AC-20 = the non-bypassable root-included
13-member cascade fixture gate"* — i.e. **the rev-4 S2-run post-merge gate that was the A-1
P0**. Rev 5 is **18 rows / ~99 min / AC-1..AC-26 / S0–S5 / Stage-pre-claim hybrid**. Ship
consumes both the plan and the queue record; a status flip without a content resync would
make Ship execute the superseded rev-4 contract and **re-instantiate A-1 and A-2**.
*(Learnings reported the records as "still rev 3"; that is inaccurate — the superseding rev-4
block is present. The split-brain conclusion stands.)*
**Fix**: Stage resyncs both records to rev 5 **before** `021-S` is claimable.

**B-8 · S2 step 2a's `>1` rule can strand a valid S1 checkpoint.**
*(Agent-Native Parity P1.)* Step 2a FAIL-CLOSES on **any** `>1` active `ship` candidate
**before** the identity filter, so one leftover unresolved record plus a valid S1 checkpoint
bricks S2 — and the installed protocol deliberately leaves fail-closed handoffs `active`
regardless of age, so leftovers are expected. Also, `session_id` is required for matching but
is not among the fields `backlogit_list_checkpoints` presents.
**Fix**: Filter to filename+`session_id` **first**; FAIL CLOSED only if the **matching** set
≠ 1; treat non-matching leftovers as a warning. Verify `session_id` after
`backlogit_get_checkpoint`, or document that list returns it.

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| C-1 | §12's table sums to **97**, not the pinned ~99 — the ADD-2 row was never raised 18→20 and still says "four-case" (rev-4 wording). AC-23/§13.6 pin 99/21 | Scope Boundary, Go |
| C-2 | Row 7's PRESENT needle (`descendants — at every depth…`) appears in **no** pinned replacement text; satisfying it forces re-adding the restatement §4.1 forbids. Its ABSENT needle is already non-vacuous alone | Go |
| C-3 | §8.1.1's marker guard is described with **inverted logic** ("fails if marker absent" + "at H1 the marker is gone ⇒ rows execute"). Taken literally the rows never execute and H1 never greens. Fails safe, but B-4 is not cleanly closed | Go |
| C-4 | Probe 25's `required_ids` omits the shipment record, which P-015 makes an unconditional member; criterion 5b could pass if the engine archived it but omitted it from `archived_ids`. Also: 8 criteria but **9** emitted tokens (5a/5b) | Architecture |
| C-5 | §5.0 hardcodes 11 `017-S` IDs into a global agent contract, coupling reusable policy to one release instance | Architecture, Scope Boundary |
| C-6 | Constitution Check's Principle VII row claims the cascade is *"never against the live backlog by this plan"* — false; §9 step 15 and §9.3 step 6 both run it live | Constitution |
| C-7 | §9.4a recovery option 2 (Stage re-runs and re-commits post-merge) is **structurally void** — the new evidence commit is a descendant of the merge SHA and fails check 3 | Constitution |
| C-8 | §9.4a's five checks verify **provenance only** and cannot detect the internally inconsistent transcript of B-2; add a read-only consistency check | Learnings |
| C-9 | P-018 Copilot engagement is absent from **both** merge points (§9 steps 5 and 17); the window closes permanently at merge | Learnings |
| C-10 | S0/§5.0's dual-location and provenance checks are not expressible on the registry surface (`get_item` cannot show torn duplicates); pin to `backlogit_doctor` / `backlogit_query_sql` | Agent-Native Parity |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| D-1 | AC-11 still reads literally *"Exactly 2 files changed"* without B-1's recommended qualifier | Scope Boundary |
| D-2 | "4 compound rows" (§12/AC-23) vs §8.1's Kind column marking {14,15,16} compound and {7,10} negative — inconsistent labelling | Scope Boundary, Go |
| D-3 | D4 omits the Governance-required *"simpler alternative that was rejected"*; Principle VIII arguably satisfied via freeze-scope rather than deviated | Constitution |
| D-4 | Principle VI row assesses cohesion but VI's body is dependency discipline; cite the stdlib-only/no-testify fact. Width-Isolation (markdown + Go in one task) unaddressed | Constitution |
| D-5 | Rows 10/14 ABSENT-needle single-occurrence across the 94 KB file not verified | Go |
| D-6 | §9 step 17 names CLI `backlogit sync` only; installed order is MCP `backlogit_sync_index` first | Agent-Native Parity |
| D-7 | S4 states the transition with no tool pin; the pin lives only in the PO-1b box | Agent-Native Parity |
| D-8 | AC-9 says *"all twelve steps"*; the checklist now has 13 items (0, 1, 2, 2a, 3–11) | Agent-Native Parity |
| D-9 | Rev-3 banner calls B/C shipments `blocked`, which is not a valid shipment status (`queued\|active\|shipped\|abandoned`); actual state is `archived`/`archived_status: queued` | Learnings |
| D-10 | B-12 (`backlogit update --complexity` never explicitly forbidden) is claimed closed by the rev-5 banner but is addressed nowhere in §12 | Learnings |

### Affirmative clearances (recorded as evidence, not findings)

* **Security Lens Reviewer returned 0/0/0/0** — all four of its attempt-1 P1s cleared:
  **A-10** (§5.0's five gates are fail-closed; `archived_status: queued` no longer proves
  completion), **A-11** (the closed verdict set + HALT rules prevent a future skill from
  naming a new destructive verdict), **A-12** (inherited-residual classification is honest),
  and **B-7** (isolation measured: `FIXTURE_ISOLATED_FROM_LIVE=True`,
  `LIVE_BACKLOG_UNMUTATED=True`; **no** user-profile paths, emails, tokens or secrets in the
  committed `.txt`). It also cleared the **A-1/A-2 hybrid**, noting the evidence-commit
  ancestry requirement defeats post-merge regeneration/replay.
* **A-1 (P0) RESOLVED.** *(Constitution, Architecture.)* The gate is re-seated pre-claim; a
  FAIL costs a re-plan with nothing merged; the post-merge path has an in-role escape; the
  no-override clause is withdrawn. **The rev-4 deadlock is genuinely broken.**
* **A-2 (P-010) CLOSED.** *(Constitution, Security Lens, Architecture.)* S2's five checks are
  pure reads (`git log`, `git merge-base`, `git diff --quiet`, `Get-FileHash`, text
  inspection); Stage authors and commits. Confirmed in the transcript: `PROBE_RUN_BY=Stage`.
* **A-7 (vacuousness) CLOSED.** *(Go Reviewer — verified against the installed file.)* All
  four pinned ABSENT literals are **verbatim-correct, single-line, and genuinely present** at
  their target sites: row 14 at **L255**, row 15 at **L777**, row 7 at **L810**, row 10 at
  **L790**. No compound row can pass vacuously.
* **B-2/B-3/B-5/C-5/C-6 (attempt 1) CLOSED.** *(Go Reviewer.)* Anchors
  `TOPOLOGY_GATE: lifecycle (before closure/safe-close)` (L788) and
  `Pre-archive reconciliation gate (mandatory)` (L792) both exist and are **unique**;
  needles are full single-line clauses; **testify is absent from `go.mod`** so stdlib is
  correctly mandated; `repoRoot(t)` **is** package-level in `build_script_test.go` and reuse
  avoids a duplicate-declaration compile failure.
* **A-8 CLOSED, A-9 (zero-candidate half) CLOSED, B-10/B-12/B-13 CLOSED, C-9 (create half)
  CLOSED.** *(Agent-Native Parity — verified against the installed Checkpoint Payload
  Contract.)* `session_id` **is** required top-level; both MCP and CLI create are valid; the
  S2-only zero ⇒ FAIL-CLOSED override is correct and correctly scoped.
* **No scope expansion.** *(Scope Boundary.)* All eight rev-5 additions map 1:1 to attempt-1
  findings; **no third implementation file**; `## Constitution Check` ← A-13, §9.4a ← A-1/A-2,
  §5.0 ← A-10, §4.1 ← A-11, §9.4b ← A-12, rows 17–18 ← B-11/B-14. **C-7 and C-8 (attempt 1)
  CLEARED** — speculative case-(iv) prose is gone; §13 items 12–13 are labelled read-only.
* **A-23 helper budget CLEARED** *(Go)* — `repoRoot` (reused) + `mustContain` +
  `mustNotContain` + `indexOfUnique` = ≤4 across 18 rows; one function holds; ~29 min is
  realistic **because** the literals are now pinned.
* **A-13 substantially CLOSED** *(Constitution)* — D1–D3 are genuine justifications, not
  disguised waivers. **P-016: no violation. Principle XI: compliant** (revert, never rewrite;
  no amend). **Principle II honored** (H0 red → H1 green is now a real transition).

### Required before re-review (P0/P1 remediation queue — cycle 3)

**B-1** is decisive and must be fixed first: **re-base every location-derived check on
declared `status`** (§5.0 gate 2, §5.1 S0). **B-2** is the same root cause in §4.4 and
should be fixed in the same pass — together they are one coherent correction. **B-3**
requires choosing one side of the enumeration contract. **B-4**, **B-5**, **B-6** and
**B-8** are localized text corrections. **B-7** is a Stage backlog action, **performed in
this session** (see the cycle-2 disposition below).

**Max re-entry cycles**: 2 remain consumed of 2; **cycle 3 is the last** before the
Escalation Protocol applies.

## Plan Review — Remediation Cycle 2 disposition

This session is **remediation cycle 2**. It closed **every attempt-1 P0/P1** and re-ran the
formal gate as **attempt 2**, which returned **`decision: FAIL`** on defects rev 5
introduced.

| Attempt-1 finding | Disposition |
|---|---|
| **A-1** (P0) | **CLOSED** — hybrid applied: Stage authors/commits Probe 25 pre-claim (§9 step 0a); S2 re-verifies read-only (§9.4a); no-override clause withdrawn; recovery path named |
| **A-2** | **CLOSED** — Ship never authors plan evidence; all five S2 checks are reads |
| **A-3** | **CLOSED** — engine resolved at runtime via registered command; digest compared; path recorded as telemetry only; Hardening 5 reconciled |
| **A-6** | **CLOSED** — §5.1 is a strictly ordered, non-overlapping S0–S5 selector; §5.2 re-seated as S3 |
| **A-7** | **CLOSED** — rows 7/10/14/15 compound with four verbatim-pinned ABSENT literals (§8.1.1), verified against the installed file |
| **A-8** | **CLOSED** — top-level `session_id`; MCP **or** CLI create |
| **A-9** | **CLOSED** (zero-candidate rule); residual `>1` handling re-opened as **B-8** |
| **A-10** | **CLOSED** — §5.0's five gates; `archived_status: queued` never proves completion; 11 IDs enumerated; `021-S` has none |
| **A-11** | **CLOSED** in substance; the *representation* is re-opened as **B-3** |
| **A-12** | **CLOSED** — §9.4b classifies the residual as inherited; reachability wording re-opened as **B-6** |
| **A-13** | **CLOSED** — `## Constitution Check` added with D1–D4 |
| **A-4, A-5, A-14, A-15, A-16** | **CLOSED in cycle 1** (unchanged) |
| **B-2, B-3, B-5, B-7, B-10, B-11, B-12, B-13, B-14, C-5, C-6, C-7, C-8, C-9** | **CLOSED** — same-surface P2/P3 queue absorbed into rev 5 |
| **B-1, B-4, B-6, B-9, B-15, B-16, B-17, C-1, C-2, C-3, C-4** | **DEFERRED with rationale** — see the deferral register below |

**Deferral register (explicit, not silent).**

| ID | Why deferred |
|---|---|
| **B-1** (AC-11 "2 files" vs evidence) | Substantively reconciled (§9.4/§10.2/§12); the literal AC wording remains — re-filed as **D-1** |
| **B-4** ("H0 red 1 of 1") | Addressed by the marker-fatal guard, but the guard's logic is inverted — re-filed as **C-3** |
| **B-6** (24 min optimistic) | Superseded: literals are now pinned, which removes the derivation churn B-6 priced |
| **B-9** (a1/lock TOCTOU) | Inherited P-015 ordering, not A's surface — same class as §9.4b |
| **B-15** (self-hosting rests on process integrity) | Honestly disclosed in AC-9's residual; no mechanical enforcement exists to add within 2 files |
| **B-16** (task-granularity heuristics) | **CLOSED** by `## Constitution Check` D1/D2 |
| **B-17** (no executable blast-radius coverage) | Out of A's 2-file surface; recorded against P-015/skill follow-up |
| **C-1** (Principle VIII safety mode) | **CLOSED** as deviation D4; completeness nit re-filed as **D-3** |
| **C-2** (stub anti-pattern / `t.Run` per row) | **CLOSED** — §8.1.1 mandates `t.Run` per row and real assertions |
| **C-3** (`tests/contract/` taxonomy) | **CLOSED** as deviation D3 |
| **C-4** (helper taxonomy imprecise) | Cosmetic; the ≤4 conclusion holds and was re-verified by the Go Reviewer |

**Stage action taken this cycle (closes B-7):** `022-F` and `022.001-T` are resynced to
**rev 5** and re-banner-ed **NOT CLAIMABLE — attempt 2 FAIL**.

**The gate verdict is `decision: FAIL`.** `status:` stays `planned`. The plan is **not**
harvest-ready. **`021-S` must not be claimed**, and `017-S` behind it stays ineligible.
**No backlog items are created from this plan.**

## Plan Review — AUTHORITATIVE GATE STATE (read this one)

**This section is the operative provenance record.** It is the **last** `## Plan Review`
section in this file and restates the markers of the **most recent** gate run, so a
`harvest` provenance check that reads the latest section cannot accidentally bind to
**attempt 1**. Attempt 1's section is retained **only** for audit history and is
**superseded**.

<!-- plan-review-attempt: 2 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

| Field | Value |
|---|---|
| **Latest attempt** | **2** (of max 3 — **cycle 3 is the last**) |
| **Reviewed revision** | rev 5 |
| **Decision** | **FAIL** |
| **Dispatch mode** | `multi-agent` (7/7 personas; anchor `openai`/`gpt-5.6-sol`/`high`) |
| **Harvest-ready** | **NO** |
| **Plan `status:`** | `planned` (unchanged — **not** advanced to `reviewed`) |
| **`021-S` claimable** | **NO** |
| **Blocking** | 1 P0 (**B-1**) + 7 P1 (**B-2**…**B-8**, less the closed **B-7**) |
| **Superseded** | attempt 1 (`<!-- plan-review-attempt: 1 -->`, rev 4, FAIL) |

**Harvest MUST halt on this plan.** `decision: FAIL` is the operative value.

