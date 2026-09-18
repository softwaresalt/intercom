---
title: "Deliberation — Scoped Option 3 harness contract amendment (B1 / B2 / DB12DA37)"
date: 2026-09-18
status: decided — Option 3-S adopted in analysis but NOT adoptable this cycle; Unit 1 BLOCKED and not claimable; Unit 2 blocked; no shipment
agent: Stage
governs: stash DB12DA37; blockers B1, B2 from docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md §9.2
supersedes: nothing — extends the Option 1 decision's §9.4 unblock path
base_commit: 101e974a2890e7ba82aacb22d734866fddaf973b
stage_branch: chore/stage-shipment-scoped-harness-contract-amendment
measured_against: installed surfaces at 101e974 (PR #61 merge)
outcome: OPTION_3_S_SCOPED_CONFORMANCE_AMENDMENT — TERMINAL BLOCKED, NO SHIPMENT, NO TASK AUTHORIZED
separability: NOT one inseparable unit — two units with an explicit DAG
revision: 1
---

## 1. Problem frame

The Option 1 decision (`docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md`)
terminated BLOCKED with seven blockers. Its §9.4 assigns **B1** and **B2** to a future Stage
session as "the Option 3 amendment work item". This deliberation is that session.

Three defects are in scope:

| Ref | As previously stated |
|---|---|
| **B1** | No execution slot for a pre-claim red harness. "Uncommitted on `main` trips Ship Step 0.5 item 3a's dirty-tree halt; committed to `main` places a red harness on the default branch outside any PR gate; an operator branch trips `BRANCH_MISMATCH`. **Every route fails a gate.**" |
| **B2** | No Stage→`main` artifact handoff. Ship Step 0.5 item 3a branches from `main`, so Stage's `.backlogit/` mutations on `chore/stage-*` are absent from Ship's branch. Declared **circular**. |
| **`DB12DA37`** | Shipment claim auto-activates descendant tasks `queued → active` before Ship Step 2's `queued`-only selector runs ⇒ Step 2 vacuous ⇒ silent **P-004 red-phase bypass**; existing authorization language names only **P-002**. |

The operator's constraint set: preserve P-001, P-002/P-004 TDD safety, P-009,
P-011/P-014/P-016, P-015 closure semantics; do not authorize around P-004; do not resurrect
retired rev-11…rev-15 mechanisms; adopt a shipment-scoped harness contract **only if** its
execution slot, branch/PR ownership, harness-ready timing and claim transition are fully
executable.

## 2. Measured evidence (installed text at `101e974`)

### 2.1 The three disagreeing selectors

`.github/agents/_ship.agent.md:322` — **Step 2 item 1**:

> 1. List all tasks for the target feature or chore that are in `queued` status.

`.github/agents/_ship.agent.md:338-358` — **Step 3 item 1**:

> Apply the exhaustive, positive status rule: KEEP `queued` and `active`; … **This derived set — not a
> queued-status-only list — is wired into item 2's ready queue below**: it is the actual task-membership
> boundary that Step 4 executes, so an `active` member of this derived set is never omitted from the
> ready queue …

`.github/agents/_ship.agent.md:362-367` — **Step 3 item 2**:

> 2. List all tasks with `harness-ready` label and `queued` status for the target feature or chore. When
>    operating under a Stage-prepared shipment (item 1 above ran), **replace this queued-only membership with
>    item 1's derived executable set**: include every task in that set regardless of whether its status is
>    `queued` or `active` …

### 2.2 P-002 already carries the enforcement clause that would have stopped this

`.github/policies/workflow-policies.md:36-52`:

> | Gate Point | Queue building (Step 2) and task claiming (Step 3)           |
>
> **Precondition** (ship): The task carries the `harness-ready` label.
>
> **Enforcement** (ship): Filter ready queue to only tasks carrying the `harness-ready` label.
>
> **Violation Action**: Halt and suggest running the harness-architect.

### 2.3 P-004 binds the producer, not the consumer

`.github/policies/workflow-policies.md:78-92`:

> | Applies To | `ship` (via harness-architect skill)         |
>
> **Statement**: The harness-architect must confirm the red phase … **before the `harness-ready` label is applied.**
>
> **Violation Action**: Do NOT apply `harness-ready` label. Halt and report the failure.

### 2.4 Ship Step 0.5 item 3a admits a pre-existing matching branch

`.github/agents/_ship.agent.md:180-196`:

> - If already on a branch matching this shipment (e.g., `feat/{slug}` or `chore/{slug}`): log
>   `WORKTREE_TOPOLOGY_OK` and `BRANCH_OK: {branch_name}` and proceed to step 4.
> - If on `main` (the default branch):
>   a. Verify the worktree is clean … If any output appears, halt.
>   …
>   d. Create the shipment branch … `git checkout -b feat/{feature-slug}` where `{feature-slug}` is
>      derived from the shipment title: lowercase, spaces replaced with hyphens.
> - If on an unrelated non-default branch: halt with `BRANCH_MISMATCH: …`

### 2.5 harness-architect accepts an explicit task set but its exclusion rule is ambiguous

`.github/skills/harness-architect/SKILL.md:54-62`:

> 2. If `${input:tasks}` is present, restrict the scope to that explicit task set.
> 3. Exclude blocked, done, or **otherwise non-ready** work items.

`SKILL.md` contains **zero** occurrences of `commit` or `branch` (searched; NOT FOUND), and names no
permitted invoker (NOT FOUND for `Ship`/`Stage`).

### 2.6 Tool mechanics

`backlogit shipment` exposes `add | claim | create | get | list | return-blocked | ship`.
There is **no `unclaim`/`release`**, and `claim` takes **no flag** to suppress descendant
auto-activation. The claim remains a one-way door. (`backlogit` 1.10.1)

`backlogit update` **does** accept `--complexity` for tasks ("task-only planning metadata …
body-preserving, mutually exclusive with other field flags"). The Option 1 decision's recorded
sizing degradation does **not** reproduce on 1.10.1; structured emission is available.

## 3. Findings

### F-1 — `DB12DA37`'s root cause is a single localized clause, not a policy defect

The bypass does not require amending P-002 or P-004. P-002 §2.2 **already** mandates
*"Filter ready queue to only tasks carrying the `harness-ready` label"* with gate point
*"Queue building (Step 2) and task claiming (Step 3)"*.

Ship Step 3 item 2 (§2.1) implements that filter on the non-shipment path (`harness-ready` **and**
`queued`), then — on the shipment path — **replaces the whole membership predicate** with item 1's
derived set. The word *replace* discards **both** conjuncts: the status filter (intended) **and the
`harness-ready` filter (unintended)**. Item 1's rule is purely status-based and never mentions
`harness-ready`.

So the bypass mechanism is exactly: **Step 3 item 2's substitution clause silently drops P-002's
Enforcement conjunct on the shipment-claim route.** Ship is *non-conformant with an installed policy
it already carries*. The remedy is a **conformance repair**, not an amendment, and it *strengthens*
rather than weakens TDD safety.

This also corrects the B3 mis-scoping recorded in the Option 1 decision: the consumer-side duty that
is violated is **P-002's Enforcement/Precondition** (Ship executing an unharnessed task). P-004 binds
the *producer* (harness-architect must not apply the label before red is confirmed) and is **not**
itself violated — nothing mislabels anything. The observable damage is a red-phase bypass; the
violated clause is P-002. No P-004 authorization is needed, sought, or granted.

### F-2 — B1 is REFUTED on the measured text: a legal pre-claim harness slot exists

B1 asserts *"every route fails a gate"* over three routes. The measured Step 0.5 item 3a (§2.4)
enumerates **four** arms, and B1's analysis omits the first:

| Route | B1's claim | Measured outcome |
|---|---|---|
| Uncommitted on `main` | dirty-tree halt | **Confirmed** — halt |
| Committed to `main` | red harness on default branch outside PR gate | **Confirmed** — and also violates Ship's own "Forbidden: commit or push directly to `main`" |
| **Unrelated** non-default branch | `BRANCH_MISMATCH` | **Confirmed** — but only for *unrelated* branches |
| **Branch matching the shipment slug** (`feat/{slug}` / `chore/{slug}`) | *not considered* | **`BRANCH_OK` → proceed to step 4.** No dirty-tree check applies on this arm; no `BRANCH_MISMATCH`. |

The fourth arm is the execution slot B1 declares missing. The red harness is committed **on the
shipment's own feature branch**, which is the branch that becomes the PR — so the harness is *inside*
the PR gate, never on `main`, and never orphaned. Branch/PR ownership is unambiguous: Ship's Role
Boundary already grants *"Create and checkout feature/chore branches, commit, push"*, and the branch
is the same one Step 0.5 3a(d) would itself have created.

**B1 therefore downgrades from a blocker to a specification gap**: the slot is *derivable* from
installed text but not *named* by it. Naming it is in scope for Unit 1; waiting on it is not.

### F-3 — the amendment that dissolves the pre-claim requirement entirely

F-2 rescues the pre-claim mitigation, but the cleaner repair removes the need for it. If Step 2
item 1's selector becomes **manifest-scoped and status-inclusive** (mirroring Step 3 item 1's already
authoritative derived set), then on the claim route Step 2 is **no longer vacuous**: it harnesses the
`active` manifest task in-session, on Ship's own feature branch, post-claim. There is then no
pre-claim harness, no orphan-file question, and no operator hand-off step at all.

Two properties make this safe rather than a loosening:

* **It cannot widen the batch.** For a Stage-assembled shipment the manifest task set is a **subset**
  of the covering feature's task set, so the harness batch size is `≤` today's. G1 (the N>1 red-batch
  deadlock) is therefore never made worse, and under the standing Option 1 shape (one task per
  feature per shipment) it stays exactly `N = 1`.
* **Step 4.3 is untouched.** The full-suite `go test ./...` gate at
  `.github/agents/_ship.agent.md:435-447` is **not** narrowed to the manifest. The workspace's
  broadest regression detector is preserved intact — the explicit caution the Option 1 decision
  raised against Option 3 (§3.3 item 1) does not apply to this scoped form.

### F-4 — a third surface must move with it (the consumer the 7-FAIL precedent predicts)

If Step 2 passes an explicitly-listed **`active`** task to harness-architect, `SKILL.md:60`'s
*"Exclude blocked, done, or otherwise non-ready work items"* becomes decisive and is **ambiguous**
for `active`: a literal reading may exclude the very task Step 2 just selected, re-vacuating Step 2
through a different door.

`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
records **seven consecutive plan-review FAILs** on exactly this class — a producer-scoped fix that
left a consumer surface open. The skill's exclusion rule is therefore **in scope, not adjacent**.

### F-5 — B2 is separable, and is not circular for this unit

B2 is real but mis-classified as a universal blocker. `.github/agents/_orchestrator.agent.md:270-297`
**does** define a Staging Artifact Merge Gate that commits to `chore/stage-{shipment_id}`, pushes,
opens a PR, waits for merge, then verifies `git show origin/main:.backlogit/queue/{shipment_id}.md`.
The mechanism exists and is in working use: PRs **#59, #60, #61** all reached `main` through exactly
this operator-driven staging-PR route.

B2's genuine defects are narrower than "no handoff exists":

1. Item 3 **names no actor**, while Stage's Role Boundary states *"The operator — never Stage — pushes
   the Stage artifact branch and opens/approves the resulting merge-commit staging PR; no Orchestrator
   authority to do so is granted or assumed."*
2. Item 3(e) directs *"Attempt a direct push to `main` first"*, which contradicts P-009
   (merge-commit-only) and P-010 (role boundary).

Neither defect blocks a unit whose deliverable is **not itself** the handback record. The circularity
the Option 1 decision recorded is specific to `025.001-T` (whose deliverable *is* the handoff
contract). **This unit is not that unit**, so it is delivered through the same operator-driven route
that merged #59/#60/#61 — including the staging PR that closes this very session.

## 4. Separability — the operator's question answered

**They are NOT one inseparable contract unit.** The measured coupling is:

* **B1 + `DB12DA37` are inseparable from each other.** Both live on the single harness
  selection/enforcement seam spanning Ship Step 2 item 1, Ship Step 3 item 2 and
  harness-architect Step 1. Splitting them is unsafe in both directions:
  * Repair Step 3 item 2 alone ⇒ every shipment-claim route now **halts** correctly on P-002
    (safe, but nothing can ship — a total block).
  * Repair Step 2 alone ⇒ the vacuity closes, but an `active` task that still lacks
    `harness-ready` for any other reason continues to execute unguarded (the bypass survives).
  Only the pair yields *"harnessed in-session **and** provably enforced."*
* **B2 is separable.** Different surfaces (Orchestrator + Stage, not Ship), different failure mode
  (artifact visibility, not test-gate bypass), different consumers, and an independently working
  operator-driven route (F-5).

### DAG

```text
                 ┌───────────────────────────────────────────┐
                 │ UNIT 1  (026-F)                           │
                 │ Harness selection + enforcement           │
                 │ conformance — Ship Step 2 / Step 3 item 2 │
                 │ / harness-architect Step 1                │
                 │ resolves: DB12DA37, B1, B3                │
                 │ STATUS: blocked — NOT claimable (§12)     │
                 └───────────────────┬───────────────────────┘
                                     │ sequencing preference (NOT a blocking edge)
                                     ▼
                 ┌───────────────────────────────────────────┐
                 │ UNIT 2  (027-F)                           │
                 │ Staging-artifact handoff actor + authority│
                 │ Orchestrator Step 1.5 items 3 / 3(e)      │
                 │ resolves: B2                              │
                 │ STATUS: blocked — own deliberation owed   │
                 └───────────────────┬───────────────────────┘
                                     │ blocks
                                     ▼
                 ┌───────────────────────────────────────────┐
                 │ 025-F / 025.001-T  Stage handback record  │
                 │ (remains blocked; B4–B6 remediable after) │
                 └───────────────────────────────────────────┘
```

Unit 1 → Unit 2 is a **sequencing** edge, not a hard functional prerequisite: Unit 2 does not consume
Unit 1's output. It is asserted because **P-001** admits exactly one top-level release unit in flight,
and because Unit 1 is the unit that makes any shipment-claim route TDD-safe — Unit 2 should ship
*through* a repaired route rather than ahead of it.

## 5. Options considered

| # | Option | Verdict |
|---|---|---|
| **A** | Full Option 3 as originally framed — amend P-002, P-004, Step 2, **Step 4.3**, harness-architect; scope the green gate to the manifest | **REJECTED.** Narrowing Step 4.3 trades away the broadest regression detector (Option 1 decision §3.3 item 1). Amending P-002/P-004 is unnecessary once F-1 shows Ship is merely non-conformant with P-002 as written. |
| **B** | Change claim semantics so the claim does not auto-activate descendants | **REJECTED — not executable.** §2.6: no flag, no `unclaim`, no engine surface. Would require a backlogit engine change, which is outside this workspace's contract surfaces and would strand the claim's one-way-door property. |
| **C** | Keep the Option 1 pre-claim-precondition mitigation and merely *document* the F-2 branch slot | **REJECTED as primary.** Leaves Step 2 permanently vacuous, leaves P-002's Enforcement conjunct unimplemented, and assigns the workspace's hardest gate to a manual operator step on every future shipment. Retained only as the **bootstrap** route for Unit 1 itself (§7). |
| **D** | Repair Step 3 item 2 only (enforcement without selection) | **REJECTED.** Correct but totalizing: every shipment-claim route halts, nothing ships. §4. |
| **E — ADOPTED: Option 3-S (scoped)** | Manifest-scope + status-include Ship Step 2 item 1; restore the `harness-ready` conjunct in Step 3 item 2's substitution; disambiguate harness-architect's exclusion rule for explicitly-listed `active` tasks; **name** the F-2 branch slot. **No change to P-002, P-004, or Step 4.3.** | **ADOPTED.** |

## 6. Chosen direction — Option 3-S

> **Status note (§12).** Option 3-S is the option selected *by this analysis*, but it was **NOT
> adopted for execution this cycle**: plan-review returned FAIL and the unit is BLOCKED with no
> shipment and no authorized task. The surface design below stands as the starting position for the
> O-3/O-4 re-scope, not as a claimable work item.

Three surfaces plus tests, delivered as **one atomic red→green task** (cardinality `N = 1`
preserved, per Option 1):

* **S1 — `.github/agents/_ship.agent.md` Step 2 item 1.** Replace the feature-scoped `queued`-only
  selector with the shipment-manifest derived executable set (`{queued, active}`) when operating
  under a Stage-prepared shipment; retain the existing feature-scoped `queued` behaviour verbatim on
  the non-shipment/direct-invocation path.
* **S2 — `.github/agents/_ship.agent.md` Step 3 item 2.** In the substitution clause, restore P-002's
  Enforcement conjunct: the derived set is further filtered to members carrying `harness-ready`, and
  any derived member lacking it is a **fail-closed halt** citing P-002's Violation Action
  (*"Halt and suggest running the harness-architect"*) — never a silent skip, never a proceed.
* **S3 — `.github/skills/harness-architect/SKILL.md` Step 1 item 3.** Make the exclusion rule
  unambiguous: when `${input:tasks}` is supplied explicitly, an `active` task is **in scope**;
  `blocked`/`done`/archived remain excluded. Also record that harness artifacts are authored on the
  invoking session's current feature/chore branch (the F-2 slot), never on `main`.
* **S4 — tests.** Behavioral/contract tests under `tests/integration/` proving: (a) the shipment-route
  selector includes an `active` manifest member; (b) a derived member lacking `harness-ready` halts;
  (c) the non-shipment path is unchanged; (d) Step 4.3's full-suite text is unmodified
  (anti-regression pin).

### Non-goals (explicit scope fence)

* Step 4.3 is **not** narrowed. P-002 and P-004 text are **not** amended.
* No `TASK_ONLY_FINALIZE`; the manifest is a fully-covered root ⇒ `CASCADE`, so the `023-F`/`024-F`
  fired-trigger is **provably not fired**.
* No rev-11…rev-15 mechanism is reintroduced in explicit or disguised form: no harness-state
  manifest, no terminal witness, no `-gatetask` selector, no surface-predicate activation, no
  MP0–MP6 mutation-proof series. Option 3-S is a selector-scope change plus restoration of an
  already-installed enforcement conjunct — categorically different from bridging machinery that
  existed to let an N>1 red batch survive Step 4.3.
* Orchestrator Step 1.5 is **not** touched (that is Unit 2).

## 7. Bootstrap — how Unit 1 ships before its own fix exists

> **Status note (§12).** This bootstrap is **DESIGNED, NOT INSTANTIATED**. No shipment was formed,
> no task was created, and step 2 below self-blocks at the C3/C4 acceptance-surface defect. It is
> retained as prospective design for the O-4 re-scope only; nothing here is executable today.

Unit 1 repairs the route it must itself travel, so its first claim runs under the **unrepaired**
contract. The bootstrap is Option C applied once, and it is exactly the proven `017-S` precedent
(`018.008-T` carried `harness-ready` at claim time, so Step 2's partition routed it to *"Already
harnessed … skip these"* and no bypass occurred):

1. Stage commits backlog/shipment artifacts on `chore/stage-*`; **operator** pushes and merges the
   staging PR (F-5 route). Artifacts reach `origin/main`.
2. **Operator** creates `feat/{shipment-slug}` from fresh `main` and invokes **harness-architect**
   for the single task — producing a real red harness committed on that branch and the
   `harness-ready` label (P-004 discharged by the producer, on its own terms).
3. Ship starts on that branch ⇒ Step 0.5 3a logs `BRANCH_OK` (F-2, fourth arm) ⇒ claims ⇒ task goes
   `active`.
4. Step 2 is vacuous (unrepaired), but harmlessly so: the task **is** genuinely harnessed, so the
   property P-004 protects holds on the artifact itself. This is disclosure of a known-vacuous step,
   **not** an authorization to execute unharnessed work.
5. Step 3 derives `{task}`; Step 4 implements; the one harness goes green; Step 4.3 full suite is
   green at `N = 1`; PR; merge; `CASCADE` closure over the fully-covered root.

**Pre-claim precondition (hard gate, recorded on the shipment):** the member task MUST carry
`harness-ready` from an actual `harness-architect` run **before** the shipment is claimed. If it does
not, the shipment MUST NOT be claimed. Executor: **operator**. Stage neither performs this run nor
authorizes any deviation from it.

## 8. Policy preservation

| Policy | Treatment |
|---|---|
| **P-001** | One top-level release unit in flight. `021-F`/`021.001-T` are terminal (`done`). Unit 2 is created `blocked`, so it is not a competing active unit. One shipment only. |
| **P-002** | **Strengthened.** S2 implements the Enforcement clause that is currently unimplemented on the shipment route. Text unamended. |
| **P-004** | **Untouched and unbypassed.** Producer-side duty unchanged; no authorization around it is sought or granted. S1 restores a real red phase to the claim route. |
| **P-009** | Merge-commit-only. Stage opens no PR; the staging PR is operator-driven. Unit 2 will remove Orchestrator item 3(e)'s direct-push-to-`main` direction, which currently conflicts. |
| **P-010** | Stage writes only decision/plan/backlog artifacts on its own `chore/stage-*` branch. No source, test, or config mutation this session. |
| **P-011 / P-014 / P-016** | Ship's branch-creation, local-review-readiness and worktree-topology gates are untouched. Stage ran a single-worktree topology precheck at Step 1.9 and created no additional worktree. |
| **P-015** | Closure semantics unchanged. **No shipment was formed and no manifest exists** (§12.4), so no closure classification is invoked and no protected set arises. The `CASCADE` / `FULLY_COVERED_ROOT` shape described in §7 is prospective design only, contingent on a future re-scope under O-4. |
| **P-021** | `DB12DA37` triaged under C6 (deliberation forced by the marker); C5 duplicate scan CLEAN; late-identifier reconciliation ran with no result (§9). |

## 9. `DB12DA37` P-021 C5/C6 disposition

* **Duplicate detection (unconditional):** performed 2026-09-18 over the active stash (21 entries)
  and `.backlogit/archive/stash.jsonl`. **CLEAN — no duplicate found.** Related-but-distinct entries
  confirmed not duplicates: `3F546E63` (claim-*last* expressibility — a different remedy on the same
  claim seam, retained), `11B75632` (Ship CASCADE binding harmonization — closure surface, unrelated).
* **Late-identifier reconciliation (triggered by `review-thread ID: N/A`):** searched all Ship-owned
  residual-risk/closure records citing `DB12DA37` (20 files). **No late identifier found.** The `N/A`
  is a truthful terminal record — the finding arose from an internal multi-persona plan-review
  (constitution persona F-03), never from a GitHub review thread, so no thread ID can exist. The
  recorded `N/A` **stands**; reconciliation completes as a no-op and does not gate this deliberation.
* **Disposition:** `DB12DA37` remains **ACTIVE** — updated in place and **not** archived, because no
  shipment delivered it and it is therefore **not consumed** (§12.4). Its root-cause localization and
  the B1 refutation are recorded on the entry so the next session starts from the measured position.
  `3F546E63` and `11B75632` also remain **active** for future triage.

## 10. Residual risks

* **R-1 — bootstrap depends on an operator action.** §7 step 2 is manual. If the operator claims
  without it, Ship executes an unharnessed task. Mitigation: the precondition is recorded on the
  shipment record itself as a hard gate; Stage halts there. After Unit 1 merges, the requirement
  disappears permanently (S1 makes Step 2 do it in-session).
* **R-2 — CASCADE closure is destructive and irreversible.** Classified `ActionRisk: destructive`;
  requires recorded operator approval and a backlog-state unwind procedure distinct from `git revert`.
  Carried into the plan.
* **R-3 — cross-surface closure risk.** The compound 7-FAIL precedent applies. The plan MUST build
  the surface matrix (producer, every consumer, authority grant, refusal path, tests, probes) against
  one bound snapshot (`101e974`) before any edit. F-4's third surface is the first instance already
  caught.
* **R-4 — three files in one task.** Exceeds the "fewer than 3 files" heuristic by one. Justified:
  `N = 1` is mandatory (a second task reintroduces G1), and the precedent `018.008-T` shipped a
  3-file contract task (2 instruction surfaces + 1 test). Recorded, not hidden.
* **R-5 — stale prose IDs (inherited).** `019-F`, `019.004-T`, `019.008-T`, `019.009-T`, `020.004-T`
  still carry prose references to the retired ID `019.007-T`. Structured edges are correct. Untouched
  here; flagged for the next session that opens the `019-F` strand.

## 11. Open questions

**Superseded by §12.5 — this section's original "none blocking" finding did not survive plan-review.**
**O-1 is BLOCKING**: the acceptance surface for agent-interpreted contract text is an operator
determination Stage may not self-grant, and it gates everything else. O-2 (P-002 Statement clause) is
also owed. Nothing is routed to Ship and no task exists, so no implementation-time selection of the
S4 test mechanism arises. Whether Unit 2 also folds in `3F546E63` (claim-last expressibility) remains
a triage decision for Unit 2's own deliberation, not resolved here.

## 12. TERMINAL OUTCOME — decision stands, unit BLOCKED, no shipment

`plan-review` attempt 1 returned **FAIL** (`dispatch_mode: multi-agent`, 2× P0, 14× P1). Stage Step 4
option **(c)** was taken and **P-005 recorded**. Full findings:
`docs/plans/2026-09-18-intercom-go-harness-selection-conformance-plan.md` § *Plan Review — attempt 1*.

### 12.1 What is settled and survives

The **analysis of §§1–5 stands and was independently re-measured by three personas**:

* **F-1 upheld** — `DB12DA37`'s root cause is localized to `_ship.agent.md:362-367`'s substitution
  clause discarding P-002's Enforcement conjunct. Step 3 item 1 (`:339-361`) contains zero
  occurrences of `harness-ready`. **No P-002/P-004 amendment is needed**; the defect is
  non-conformance with an installed policy.
* **F-2 upheld** — **B1 is REFUTED.** `_ship.agent.md:189`'s fourth arm (*"If already on a branch
  matching this shipment … log `BRANCH_OK` … proceed"*) is a legal pre-claim harness slot, inside
  the PR gate, never on `main`. Prior art did not consider-and-reject it; B1 was enumerated over
  only three routes. **B1 downgrades from blocker to specification gap.**
* **F-5 upheld** — B2 is separable and not circular for a unit whose deliverable is not the
  handback record itself.
* **AC-6 analogue upheld** — no rev-11…rev-15 mechanism is present in explicit or disguised form.
* **Scope preservation upheld** — P-001, P-004, P-009, P-010, P-015 preserved; the
  `023-F`/`024-F` trigger provably unfired; Step 4.3 correctly left un-narrowed.

### 12.2 Why Option 3-S is nonetheless not adoptable this cycle

Against the operator's admission gate — adopt only if **execution slot, branch/PR ownership,
harness-ready timing, and claim transition** are *all* fully executable — the measured result is
**2 of 4 PASS, 2 of 4 FAIL**:

| Gate | Result |
|---|---|
| Execution slot | **PASS** — F-2; `_ship.agent.md:189` |
| Branch/PR ownership | **PASS** — Ship Role Boundary `:36-42`; the shipment branch becomes the PR |
| Harness-ready timing | **FAIL** — P-002's **Statement** row (`workflow-policies.md:44`, *"may only **claim** and implement a task after … red phase"*) binds the *claim*, not only implementation. Option 3-S's steady state harnesses **post-claim**, so the shipment-level auto-activation still precedes red confirmation. Never adjudicated. |
| Claim transition | **FAIL — not repairable.** `backlogit` 1.10.1 exposes no `unclaim`/`release` and no flag to suppress descendant auto-activation (measured, §2.6). The one-way door and the auto-activation are both engine properties, unreachable from any contract surface in this workspace. |

### 12.3 The structural blockers (C1–C8)

C1/C2 (Step 2↔Step 3 temporal incoherence and duelling execution-boundary authority), C5 (the real
surface inventory is ~8–9 clause sites across ≥4 files — including `harness-architect/SKILL.md:50`
and `:56`, two further `ready`-scoped doors) and C6 (the `manifest ⊆ feature` cardinality proof is
invalid) together expand the unit far past the 2-hour rule — and it **cannot be split**, because a
second queued test-bearing task under one covering feature reproduces the G1 deadlock (D4). This is
the same structural trap that terminated the Option 1 decision.

C3/C4 are the harder pair: the unit has **no admissible acceptance surface**. An "unchanged"
assertion is green-on-arrival and can never satisfy P-004's *"expected failure markers for every
test function"*, so harness-architect must refuse the label and the bootstrap self-blocks at PA-3.
A behavioural harness — the surface the adopted `022.001-T`/IV-5 decision requires — has **no engine
to drive**, because Ship Steps 2–3 are agent-interpreted prose rather than code. Declaring IV-5
non-applicable is an **operator determination Stage may not self-grant**.

### 12.4 Disposition

* **No shipment created.** No task moved to `queued`. Nothing routed to Ship.
* **`026-F` created `blocked`** — carries C1–C8 in-record as the Option 3-S re-plan scope.
* **`027-F` created `blocked`** — carries B2 (Orchestrator Step 1.5 actor/authority defect).
* **`DB12DA37` remains ACTIVE** — updated in place, **not** archived: it is not consumed, because no
  shipment delivered it. Its root-cause localization and the B1 refutation are recorded on the entry
  so the next session starts from the measured position rather than re-deriving it.
* `019-F`, `020-F`, `023-F`, `024-F`, `025-F` — **unchanged, all still `blocked`**.
* `3F546E63`, `11B75632` — remain **active** in the stash.

### 12.5 Unblock path — owner: a future Stage session, not Ship

Ordered, and deliberately smaller than this session attempted:

1. **O-1 (operator determination, blocks everything else).** Decide the acceptance surface for
   agent-interpreted contract text: is a structural/substring contract test admissible as the IV-5
   surface when no executable engine exists, or must a contract-linter engine be built first? Stage
   cannot self-grant this.
2. **O-2 (Stage).** Adjudicate P-002's **Statement** clause against post-claim in-session harnessing —
   either the claim in *"may only claim and implement"* means Ship Step 4.1's **task** claim (guarded
   by S2), or Option 3-S's steady state is non-conformant. This is a one-question deliberation.
3. **O-3 (Stage).** Rebuild the surface matrix **whole**, per the compound stop rule both personas
   invoked — all ~8–9 clause sites, four columns populated with verbatim current text and literal
   pinned replacements, probe rows included, `Depends on` edges explicit. Do not patch locally.
4. **O-4 (Stage).** Only then re-scope Unit 1 against a matrix that is actually closed, and resolve
   C1 by deciding **where** the authoritative selected-batch derivation lives (Step 2 as producer with
   Step 3 as consumer) rather than cross-referencing forward.
5. **O-5.** B2 / `027-F` may proceed independently — it is not gated on O-1…O-4.

Until O-1 and O-3 are closed, **no shipment may be formed over `026-F`, and it must not be routed to
Ship.**

### 12.6 Known residual (inherited, unremediated)

`019-F`, `019.004-T`, `019.008-T`, `019.009-T` and `020.004-T` still carry prose references to the
retired ID `019.007-T`. Structured dependency edges are correct. Untouched here; flagged for the next
session that opens the `019-F` strand.
