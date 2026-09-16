---
title: "Decision — A-only simplification of the bootstrap chain"
date: 2026-09-13
status: decided
agent: Stage
governs: features 022-F, 023-F, 024-F, 018-F; shipments 021-S, 017-S
supersedes_scope_of: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
superseded_scope: "the three-shipment A → B → C chain, for the CURRENT release scope only"
plan: docs/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md
deferred_plans:
  - docs/plans/2026-09-12-intercom-go-p015-taskonly-authorization-plan.md
  - docs/plans/2026-09-12-intercom-go-taskonly-finalize-skill-implementation-plan.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
source_stash: A10EF3D0 (archived)
---

# Decision — A-only simplification of the bootstrap chain

> **Verdict: `SIMPLIFICATION_VALID`.** The current outcome needs only **A**
> (`021-S` / `022-F` / `022.001-T`). `TASK_ONLY_FINALIZE` shipments **B** and **C** are
> **deferred generalized platform work**, not current prerequisites.

## 1. Problem frame

The three-shipment bootstrap (deliberation
`docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md`) was
built to deliver a **`TASK_ONLY_FINALIZE`** close path so that a **task-only** manifest
could be finalized. Its chain was `017-S → 023-S (C) → 022-S (B) → 021-S (A)`.

That chain is sound **for the problem it was posed**. But the problem was posed against a
premise that this decision retires: that `017-S` **must** remain a **task-only**
12-member manifest, and therefore **must** wait for a new task-only close verdict to be
invented, authorized and implemented.

## 2. The premise that changed

The rev-18 `018-F` contract excluded `018-F` from `017-S`'s manifest. Its stated reason
was precise and, at the time, correct:

> *"018-F is live in `.backlogit/queue/` with status `active` … and **NO Ship step ever
> moves a covering FEATURE to done** — Step 4.5 moves TASKS only. 018-F would therefore
> classify `status-mismatch`, pre-mode would return HALT."*

The load-bearing clause is **"no Ship step ever moves a covering feature to done"**.

**Shipment A creates exactly that step.** A's `ADD-2` introduces **Step 6.1(a1)**, an
operative post-task / pre-reconcile covering-feature completion gate, and A's `ADD-1`
grants the narrow `active -> done` Role Boundary authority it needs.

Once A is merged, the rev-18 exclusion rationale **no longer holds**, because the
condition it was conditioned on no longer holds.

## 3. Options considered

| # | Option | Assessment |
|---|---|---|
| **O-1** | Keep the full `A → B → C` chain; close `017-S` by the new `TASK_ONLY_FINALIZE` verdict | **Rejected.** Invents a new policy verdict, a capability token, a classifier, G0–G13 / P1–P10 guards and a digest-bound invocation — ~3 further hours across two shipments — to solve a problem that A's own Step 6.1(a1) already dissolves. Highest blast radius of the three options: it amends P-015's verdict set |
| **O-2 (CHOSEN)** | **Ship A only.** Restore `017-S` to a fully-covered root and close it by the **existing** P-015 `CASCADE` exception | **Chosen.** Zero new verdicts, zero new policy surface, zero new skill machinery. Reuses a path that is already written, already reviewed, and already exercised by A's own closure |
| **O-3** | Ship A, then hand-dispose `017-S` by operator action at Step 6 | **Rejected.** Leaves the workspace-wide `SAFE_CLOSE` exit-9 blocker live on the route, and substitutes an untracked manual step for a mechanical one |

## 4. Decision

**O-2.** The current scope is **A only**.

### 4.1 How A closes itself

Unchanged from the A plan. `021-S`'s manifest is the fully-covered root
`[022-F, 022.001-T]`. After A merges, a **fresh** Ship session (never the merging session
— the §6 authority-escalation rule) loads the new Role Boundary, runs Step 6.1(a1) to move
`022-F active -> done`, and the **existing** P-015 `VERIFIED FULLY-COVERED-ROOT EXCEPTION`
(`CASCADE`) closes `021-S`.

A therefore **closes by a path that already exists**. It authorizes no new verdict.

### 4.2 How `017-S` then closes

`017-S` is **restored to full-root coverage**: `018-F` is added back to the manifest,
giving **13** members — the root `018-F` plus all 12 existing descendant members. With
A live:

1. Step 6.1(a1) applies as **case (ii)** — `018-F` is a present feature member, the five
   §5 conditions hold, so `018-F active -> done`.
2. Pre-mode's `expected_status: done` sweep now passes over **every** member: `018-F` is
   `done`, `018.008-T` is `done`, and the 11 legacy descendants are **pre-archived**,
   which pre-mode accepts without a status check.
3. Classification returns **`CASCADE`** — the same **existing** exception A used.

> **Verification scope (internal review P2).** This is established at the
> **classification** level. It is **not** engine-verified for this exact member shape — no
> committed probe exercises a root-*included* cascade whose siblings are genuinely
> `status: archived`. The path is **fail-closed** (the skill's cascade step 2 HALTs on
> non-empty `returned_ids`; step 3's two-set gate HALTs on mismatch), so the realistic
> downside is a HALT and return to Stage, not corruption. A root-included fixture probe is
> a recorded follow-up (§7 item 4), to be run **before `017-S` is routed** — it is **not**
> a prerequisite for A.

`017-S` never needs `TASK_ONLY_FINALIZE`, because it is **no longer a task-only
shipment**.

### 4.3 Why the rev-18 `SAFE_CLOSE` blocker disappears

Rev 18 recorded an honest, pre-existing residual: a `SAFE_CLOSE`-classified shipment halts
at skill step 8 because `backlogit move <shipment_id> --status shipped` is refused with
**exit 9** (Probe 18). Rev 18 also stated the blocker was unavoidable:

> *"NO manifest shape can avoid it, because the only shape that reaches the permitted
> cascade path is one containing 018-F, which fails pre-mode first."*

The second half of that sentence is what A repeals. With Step 6.1(a1) live, a manifest
containing `018-F` **no longer fails pre-mode first** — a1 moves the feature to `done`
immediately *before* pre-mode reads it. The cascade path becomes reachable, `017-S` never
classifies `SAFE_CLOSE`, and the exit-9 blocker is **never reached on this route**.

The underlying tool/contract conflict is **not fixed** by this decision — it is simply
**not on the route**. It is recorded in §7 as a future follow-up.

## 5. What B and C become

**Deferred generalized platform work.** Not cancelled, not deleted, not current
prerequisites.

| Unit | Disposition |
|---|---|
| **B** — `023-F`, `023.001-T` (dormant token-gated `TASK_ONLY_FINALIZE` in P-015) | `blocked`; shipment record `022-S` **archived** (non-destructively, manifest and dependency provenance preserved) |
| **C** — `024-F`, `024.001-T` (classifier + guards in `shipment-reconcile`) | `blocked`; shipment record `023-S` **archived** (non-destructively, manifest and dependency provenance preserved) |

Their plans remain on disk and remain **correct at rev 2**. Their rev-2 correctness
notes — in particular C's Probe 22 / 23 / 24 measurements about manifest shape, the
non-constructible feature-prerequisite barrier, and the agent-driven execution boundary —
are **preserved for future use**.

**If** a genuinely task-only shipment is ever required, B and C are re-entered through
**fresh, separately reviewed shipments**. They are not resurrected by re-routing the
archived records.

### 5.1 Why deferral is safe

B's entire value is a **dormant** policy branch that is *unselectable* until C ships the
capability token. C's value is the token and classifier. Neither is referenced by any
current route once this decision lands. Deferring both returns the selectable verdict set
to exactly **`CASCADE`, `SAFE_CLOSE`, `HALT`** — byte-for-byte today's behavior, which is
the same end state B itself guaranteed between its own merge and C's.

Deferral therefore removes ~3 h of work and an amendment to P-015's verdict set **without
changing the behavior of any shipped surface**.

## 6. What is retained from the three-shipment deliberation

The superseded deliberation is retained **in full** as historical reasoning. These
findings survive and are **not** re-litigated:

* The rev-6 `MUST_REPLAN` verdict and its two premises (~2.5–2.7 h; unsplittable under one
  covering feature) — **still adopted**. A-only satisfies both: A alone is ~78 min.
* The `021.001-T` arithmetic — a covering feature with more than one queued test-bearing
  task deadlocks at Ship Step 2/4.3. A has exactly **one** queued task; `018-F` has
  exactly **one** (`018.008-T`). Both respect it.
* A's own justification is **independent of the chain**. §2.2 of the A plan records that
  Ship's duplicated close-path wording has **already drifted unsafe** (`children` where
  P-015 v1.24.0 requires `descendants at every depth`). That defect is real today and is
  fixed by A regardless of whether B and C ever ship.

## 7. Future follow-ups — recorded, NOT current blockers

1. **`SAFE_CLOSE` step 8 / exit 9.** The workspace-wide conflict between
   `shipment-reconcile` step 8's `backlogit move --status shipped` prescription and
   backlogit 1.10.1's refusal for shipment artifacts. Off-route after this decision;
   still live for any future `SAFE_CLOSE`-classified shipment.
2. **Task-only platform work (B + C).** Re-enter only under a fired trigger — an actual
   task-only shipment that cannot be re-shaped to a covered root.
3. **Stale source references in the INSTALLED skill.** `shipment-reconcile/SKILL.md`
   safe-close **Step 0(c)** (~line 413, and again at ~1075) asserts that *"this
   self-hosting repository's own implementation lives at"*
   `src/autoharness/gates/shipment_closure.py`. **That path does not exist in
   intercom-go** — there is no `src/` directory and no executable close-path classifier;
   classification is **agent-driven from the skill markdown**.
   **Correction to an earlier framing (internal review P3):** this is **not** a C-only,
   off-route issue. Step 0(c) is on the **`CASCADE` route of both A and `017-S`**, so both
   current-scope closures traverse it. It is **non-fatal** — the classification prose in
   the skill is authoritative and an agent implements the check directly against
   `queue/` + `archive/` — but the reference is misleading on a live route and should be
   reconciled in the installed skill. C's plan reconciles it in place at `B9`/`B17`
   whenever C is re-entered; until then it stands as a recorded follow-up.
4. **Root-included cascade fixture probe.** No committed probe exercises a cascade whose
   manifest contains the root **and** genuinely `status: archived` siblings — the exact
   `017-S` 13-member shape. The path is fail-closed at two gates, so the risk is a HALT
   and return to Stage, not corruption. Run before `017-S` is routed; **not** a
   prerequisite for A.
5. **`_ship.agent.md` pre-mode summary (L776–777).** Says pre-mode verifies every manifest
   item is *"present in queue with `status: done`"*, which is false for pre-archived
   members. Deliberately outside A's 6-site inventory (see plan A §9.3); the authoritative
   skill behaves correctly regardless.

None of these gates A, `017-S`, or this session.

## 8. Consequences for the backlog

| Change | Detail |
|---|---|
| `017-S` manifest | 12 → **13** members (root `018-F` restored) |
| `017-S` dependencies | `[023-S]` → **`[021-S]`** exactly |
| `022-S`, `023-S` | **archived** (non-destructive; `archived_status: queued` preserved) |
| `023-F`, `023.001-T`, `024-F`, `024.001-T` | **blocked**; members of no live shipment |
| Eligibility | **`021-S` only.** `017-S` stays queued and blocked behind A |
| Unrelated | `019-F`, `020-F`, `021-F` untouched |

**Ordering note.** The dependency rewrite was performed **add-before-remove** —
`017-S → 021-S` was created while `017-S → 023-S` still blocked, and the stale edge was
removed **last** — so `017-S` was never transiently eligible.

## 9. Route after this decision

```text
A (021-S) merges
  → fresh Ship session: Step 6.1(a1) moves 022-F done → existing CASCADE closes 021-S
    → 017-S becomes eligible
      → Step 6.1(a1) moves 018-F done → existing CASCADE closes 017-S
```

Two shipments, one new gate, **zero** new close-path verdicts.

## 10. Post-decision status — A-only final review, remediation cycle 1 (2026-09-13)

This decision (`SIMPLIFICATION_VALID`) **stands**. Nothing below re-opens the A-only
simplification itself: the retirement of the three-shipment `A → B → C` bootstrap, the
two-shipment `A → 017-S` route, the deferral of B and C, and the `CASCADE`-not-
`TASK_ONLY_FINALIZE` close path are all **unchanged and unchallenged** by the review
described here.

What changed is the **readiness of Plan A**, and two of this decision's own recorded
follow-ups have been **re-graded**.

### 10.1 Plan A is rev 4 and its formal review gate returned FAIL

A formal `plan-review` gate was executed against
`docs/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md` rev 4 —
**attempt 1**, `dispatch_mode: multi-agent`, all seven selected personas dispatched
(including the `gpt-5.6-sol` anchor route), **`decision: FAIL`**.

**Consequences, stated plainly:**

* Plan A `status:` remains `planned`. The plan is **not** harvest-ready.
* **`021-S` must not be claimed.** `017-S` is blocked behind it and is therefore also not
  claimable. The route in §9 above is **correct but not yet executable**.
* Two findings — **A-1** and **A-2** — require **operator adjudication** before remediation
  cycle 2 can proceed. Both concern §9.4 of Plan A, which is **new in rev 4 and did not
  exist when this decision was taken**.

### 10.2 Re-graded follow-up — the root-included cascade fixture

§7 of this decision recorded the root-included 13-member cascade fixture as a follow-up to
be run *"before `017-S` is routed"*. Plan A rev 4 **promoted it to a non-bypassable gate on
A's own close** (§9.4 / AC-20) on the ground that A is an irreversible contract change whose
sole release value is making `017-S` closable.

**That promotion is itself now contested**, by four independent reviewers:

* **A-1** — the gate sits **after** A's PR merges to `main`, so on FAIL the Role Boundary
  change is already deployed while `021-S` is stuck in-flight, and §9.4 grants **no operator
  override** — which also contradicts Plan A's own §10.2 `git revert` rollback.
* **A-2** — the gate assigns **authoring and committing** `docs/plans/evidence/probe25-*` to
  **Ship (S2)**, but Ship's Role Boundary **forbids** creating or modifying plan artifacts
  (P-010). Executed literally, S2 halts and A can never close.

**Unresolved, pending operator adjudication.** The recommended resolution — which preserves
the operator's stated intent at the lowest risk — is the **hybrid**: **Stage** runs and
commits Probe 25 **before `021-S` is claimed**, and the fresh S2 session **re-verifies the
committed evidence read-only** (presence + digest + PASS tokens) before A may be marked or
archived `shipped`. That keeps the gate non-bypassable, removes the deadlock, and removes
the P-010 conflict. **No change has been made to §9.4 in remediation cycle 1** — the gate is
recorded exactly as directed, with both findings attached.

### 10.3 New follow-up — a disclosed narrowing (finding A-5)

Not previously recorded anywhere. P-015's fully-covered-root exception is *"quantified over
**every feature member of the manifest**"* and item 4 says *"the qualifying root feature
**member(s)**"* — **plural is authorized by policy**. Plan A's new a1 case (iv) halts
whenever more than one feature member is present, which **narrows** a manifest shape P-015
permits. The halt is safe (it mutates nothing) and is retained; the **claim that A "narrows
nothing" is withdrawn**. Interim rule: a multi-root manifest must be dispositioned by
**Stage**. Recorded as stash `D8397D20`.

### 10.4 Governing-contract contradiction repaired (finding A-15)

§R16 of `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` is
now **rev 20**. Rev 19 had corrected §R16.2 to the 13-entry fully-covered root — the shape
**this decision** established — but left §R16.3 through §R16.3.3 operatively asserting the
**opposite**: a 12-entry task-only manifest, a "WITHDRAWN" and "unreachable by construction"
`CASCADE` shape, and *"neither is remediable from a Stage planning cycle"*. Since §R16 is
declared the whole governing contract, an executor reading §R16.3 would have built the
withdrawn shape. Rev 20 makes §R16.3 the **sole operative** manifest contract and relabels
§R16.3.1 / §R16.3.1a / §R16.3.2 **HISTORICAL — NON-GOVERNING**. **This repair makes the
governing plan agree with this decision**; it does not change this decision.

### 10.5 Other follow-ups recorded this cycle

* Stash **`4029DABB`** — `shipment-reconcile/SKILL.md` cites `src/autoharness/...` classifier
  paths that **do not exist** in this Go workspace. Confirmed **out of Plan A's 2-file scope**
  and correctly left unedited; recorded rather than scope-crept.
* Orchestrator Step 1.5's item 3 contradiction (3(a) commit, 3(e) direct-`main` push attempt)
  is now covered by an **explicit fail-closed runtime directive** in §R16.5 of the policy-gap
  plan. The Orchestrator itself was **not modified**; its repair stays deferred to
  `019.004-T`. The **operator-only PR boundary is preserved unweakened** (Probe 17).

### 10.6 What this means for §9's route

The route diagram in §9 is **unchanged and still correct**. It is simply **gated earlier than
this decision anticipated**: not by `017-S`'s readiness, but by Plan A's own review verdict.
`021-S` is the only eligible shipment *by topology*, and it is **not claimable** *by gate*.
