---
title: "Decision — Harness execution model for deferred Stage-enforcement work (021.001-T)"
date: 2026-09-18
status: decided — downstream unit BLOCKED, no shipment
agent: Stage
governs: task 021.001-T; feature 021-F; deferred features 019-F, 020-F; new feature 025-F
resolves: 021.001-T — AC-1, AC-2, AC-3, AC-6, AC-7 discharged; AC-4 NOT ACHIEVED (plan-review FAIL ×2); AC-5 correctly not performed (no shipment). See §6 and §9.
outcome: OPTION_1_SINGLE_TASK_FEATURES_AND_SHIPMENTS — unit not safely claimable (B1–B7, §9.2)
base_commit: 1bef819843771e028ce8f1cff4dfffdae920a43e
stage_branch: chore/stage-harness-execution-model-readiness-decision-defer
plan_review_attempts: 2
plan_review_verdict: FAIL (attempt 1), FAIL (attempt 2)
shipment: none — AC-5 gated on AC-4, which did not pass
revision: 3
revision_note: "Rev 3 records the TERMINAL BLOCKED OUTCOME after plan-review attempt 2 returned FAIL (5× P0). Adds §9 (blockers B1–B7, disposition, unblock path, known residual); §7 retitled DESIGNED, NOT INSTANTIATED with the unmet queued-status premise called out; §6 AC-4 marked NOT ACHIEVED and AC-5 marked correctly-not-performed. Remediates attempt-2 findings N2 (§3.1 Step 3 proof moved off the refuted pre-adoption ID), N15 (verdict token corrected HALT → BLOCK, §3.2(c)) and N16 (AC-4 no longer defers to a nonexistent section). Rev 2 remediated plan-review attempt 1 (FAIL — F1/F2/F3 P0, F4–F7 P1): restructure performed and measured (§4, incl. the unanticipated backlogit adopt ID remap 019.007-T → 025.001-T); Step 2 proof corrected to the dual-route framing (§3.1); closure path corrected from safe-close to CASCADE / FULLY_COVERED_ROOT (§3.1); R1 mitigation mechanism corrected (§5.3); §5.4 rewritten (ID-prefix risk eliminated, R-A/R-B/R-C added)."
---

# Harness execution model for deferred Stage-enforcement work

**Chosen option: OPTION 1 — single-task features and shipments.**
Rejected: Option 2 (one-queued-child serialization), Option 3 (shipment-scoped
harness contract amendment). No fourth mechanism is introduced. No retired
rev-11…rev-15 mechanism is resurrected.

## 1. Problem frame

`021.001-T` is the readiness gate for every deferred Stage-enforcement feature.
It produces a decision artifact, not code, and belongs to no shipment.

The defect it must resolve is a three-way contract interaction:

| Surface | Exact installed text | Consequence |
|---|---|---|
| **P-004** | every generated test function must be **red** before implementation begins | N harnessed tasks ⇒ N red functions |
| **Ship Step 2 item 1** (`.github/agents/_ship.agent.md:322`) | *"List all tasks for the target feature or chore that are in `queued` status."* | the harness batch is scoped to the **covering FEATURE**, not to the shipment manifest |
| **Ship Step 4.3** (`.github/agents/_ship.agent.md:435-441`) | *"After the build-feature skill reports success: 1. **Lint**: `go vet ./...` … 3. **Full Test Suite**: `go test ./...`. If any gate fails, return to the build-feature skill for a fix iteration."* | the **full** suite must be green after **every** task |

For any covering feature holding **more than one queued test-bearing task**:
Step 2 harnesses all N in one batch (all red by P-004), task 1 is implemented and
goes green, functions 2…N remain red **by design**, Step 4.3's full-suite gate
fails, `build-feature` burns its 5 attempts implementing future work, and no task
can complete. This is the G1 deadlock.

Revisions 10–15 of the predecessor plan attempted to bridge this with surface
predicates, a harness-state manifest, a terminal witness, a `-gatetask` selector
and an MP0–MP6 mutation-proof series. All five failed adversarial review. Their
defects are recorded as **NON-GOVERNING provenance** in the historical appendix of
`docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md`.

**The withdrawn guidance.** The earlier recommendation to "split into a sequence of
single-task SHIPMENT MANIFESTS" is invalid and stays withdrawn: Step 2's batch is
scoped to the feature's queued tasks, so several single-task manifests over one
multi-task feature still yield one N-function red batch and still deadlock. The
split must be at the **feature** level, not the manifest level.

## 2. Evidence base — the four contract surfaces AC-2 requires

### 2.1 Ship Step 2 item 1 — feature-scoped queued-task listing

`.github/agents/_ship.agent.md:316-333`:

> ### Step 2: Harness Generation (P-002 / P-004)
> Ensure every task in the target feature or chore has a passing test harness before any implementation begins. This step runs once, up front — not in a loop.
> 1. List all tasks for the target feature or chore that are in `queued` status.
> 2. Partition the task list:
>    * **Already harnessed**: tasks carrying the `harness-ready` label — skip these.
>    * **Needs harness**: tasks without the `harness-ready` label — scaffold these.
> 3. If any tasks need harnesses, invoke the **harness-architect** skill for the batch.
> 4. After scaffolding completes, confirm every queued task now carries the `harness-ready` label. If any task still lacks it, halt and report the gap rather than proceeding with a partial set.

Two load-bearing properties: the selector is **`parent_id`-scoped to the feature**
(the shipment manifest is not consulted), and the status filter is **`queued` only**.

### 2.2 Ship Step 3 item 1 — manifest executable-set derivation, fail-closed status rule

`.github/agents/_ship.agent.md:335-360`:

> 1. **Shipment-manifest executable-set derivation …**: The shipment manifest (`custom_fields.items` recorded at Step 0.5) is the **closure membership record** — it is never the executable task set and is never mutated to make execution proceed. Before assembling the queue in item 2 below, filter the manifest to task artifacts (IDs ending `T`; the covering feature is resolved through `parent_id` and is never executed — the 097-S task-only-manifest precedent), THEN read each task record's status; artifact-type filtering always precedes any status read. Apply the exhaustive, positive status rule: KEEP `queued` and `active`; SKIP-AND-REPORT an archived member as `pre_archived_skipped` …; REPORT an already-`done` member separately as `already_done`; **ANY OTHER, MISSING, OR UNREADABLE status is a FAIL-CLOSED HALT, never a skip.** … If the derived executable set is EMPTY while the manifest is non-empty, HALT and report …

Two load-bearing properties: the positive status rule admits exactly
`{queued, active}` as executable and `{archived, done}` as reported skips — **`blocked`
is in none of these sets and is therefore a HALT**; and an **empty** derived set over a
non-empty manifest is itself a HALT.

### 2.3 Ship Step 4.3 — the green gate

`.github/agents/_ship.agent.md:435-441`, quoted verbatim in §1. The gate is the
**full** `go test ./...`, evaluated **after every task**, with failure routed back
into `build-feature` rather than tolerated.

### 2.4 P-015 / `shipment-reconcile` — protected set and closure semantics

`.github/policies/workflow-policies.md:414-438`:

> **Precondition**: The shipment manifest (`items`) has been read, and the **protected set** — the covering feature plus every unshipped sibling task that is not in the manifest — has been computed from **expected IDs** (partial-feature detection scanning both queue and archive). A clean git baseline for `.backlogit/` has been captured, and a **baseline integrity gate** has confirmed every protected-set member is present in `.backlogit/queue/` before any manifest item is archived. **A protected-set member already archived or missing at baseline is treated as a pre-existing cascade and halts closure.**
>
> **Required Check (verify-after-each invariant)**: After archiving each single manifest item … confirm every protected-set member still exists in `.backlogit/queue/` … this exemption applies to **manifest items only** — the protected set has **no** pre-archived exemption …

`.github/skills/shipment-reconcile/SKILL.md:611-630`:

> 2. **Compute the protected set** (partial-feature detection):
>    * Derive the covering feature ID from the manifest item hierarchy (e.g. a task `055.002-T` belongs to feature `055-F`).
>    * If the covering feature ID is **not** in the manifest `items`, this is a **partial-feature shipment**. Add the covering feature to the protected set.
>    * Enumerate every task sharing the covering feature's hierarchy prefix whose ID is **not** in the manifest `items` (the unshipped siblings) … Add each to the protected set.
>    * **Sequence-aware exclusion (serial partial-feature shipments)**: when a sibling belongs to a predecessor shipment in the same feature-split sequence, exclude it from the protected set **ONLY** when that predecessor shipment record itself has **verified archived provenance** `archived_status: shipped` … Mere archive-file presence, `archived_status: active|queued|blocked|abandoned`, generic `status: archived` without shipped/done provenance, or missing/ambiguous provenance are **NOT** sufficient — in those cases the sibling stays protected fail-closed.

And the fully-covered-root exception, `.github/policies/workflow-policies.md`
item 1:

> 1. **Root-and-fully-covered, for every feature member.** Each feature member of the manifest MUST be a **root** (it declares no `parent_id`) AND MUST be **fully covered** (every one of its **descendants — at every depth** … enumerated by walking the full `parent_id` graph live from `.backlogit/queue/` and `.backlogit/archive/` … is also present as a manifest member).

## 3. Option evaluation

### 3.1 OPTION 1 — single-task features and shipments — **CHOSEN**

Shape: each strand is decomposed into covering features that each hold **exactly one**
queued task; manifest = `[feature, task]`.

**Proof against Step 2 item 1 — both routes, stated exactly.** *(Corrected at plan-review
attempt 1, finding F3: the original text proved only the non-claim route and was contradicted
by this artifact's own §5.2.)* Step 2 runs **after** the Step 0.5.4 claim, and the claim
auto-activates manifest tasks (`DB12DA37`, §5.1). The selector must therefore be evaluated on
two distinct routes:

* **Non-claim route** (feature targeted directly, no shipment claim has run): the selector
  lists the feature's `queued` tasks. A feature with exactly one task yields a batch of
  exactly **one**; `harness-architect` generates **one** red function.
* **Shipment-claim route** (the route this workspace actually uses): the task is already
  `active` when Step 2 runs, so item 1 returns **∅** and Step 2 is **vacuous** — it scaffolds
  nothing, and item 4's confirmation passes vacuously over the empty set. This is the
  `DB12DA37` consequence developed in §5.2, and Option 1 does **not** repair it.

**What Option 1 does and does not prove here.** Option 1 does **not** claim Step 2 produces a
batch of one on the claim route. It proves the property that actually matters for the G1
deadlock: **the total number of harnessed-but-unimplemented test functions in the tree is at
most one, on either route.** On the non-claim route that bound is enforced by Step 2's batch
size; on the claim route it is enforced by the §5.3 pre-claim `harness-architect` run, which
harnesses exactly the one task in the feature. Cardinality 1 is what Step 4.3 needs (§3.1
below); *which step produced the harness* is immaterial to that gate. No contract text is
amended; the deadlock is removed by making N = 1.

The contrast with the deadlock is exact: G1 arises only when Step 2 emits **N > 1** red
functions for one feature. A vacuous Step 2 emits **zero**, which cannot deadlock Step 4.3 —
it fails a *different* way (executing an unharnessed task), which is §5.2's finding and is
mitigated in §5.3, not here.

**Proof against Step 3 item 1.** *(Corrected at plan-review attempt 2, finding N2: rev 2
updated the Step 2 and P-015 proofs to the post-adoption ID but left this third proof on
the refuted pre-restructure ID. The corrected manifest is the one 025-F now records.)*
Manifest `[025-F, 025.001-T]`. Artifact-type filter
(IDs ending `T`) runs first and removes `025-F` — the covering feature "is resolved
through `parent_id` and is never executed". Remaining member `025.001-T` is read at
status `queued` ⇒ **KEEP**. Derived executable set = `{025.001-T}`, non-empty, so the
empty-set HALT does not fire. **No manifest member is ever in `blocked` status**, so
the fail-closed "ANY OTHER status" HALT cannot fire. This is the same derivation that
`018.008-T` recorded and that shipment `017-S` executed successfully.

**Proof against Step 4.3.** After the single task is implemented, its one harness
function is green and **no other red function exists anywhere in the tree** — there is
no second harnessed-but-unimplemented task by construction (this is the cardinality-1
property established above, which holds on both routes). `go test ./...` is green. The gate
passes on its first evaluation. This is the precise point at which Option 1 discharges G1.

**Proof against P-015 / `shipment-reconcile`.** *(Corrected at plan-review attempt 1,
finding F2: the original text asserted a safe-close protected-set computation that this
manifest never reaches.)* The manifest is `[025-F, 025.001-T]`. Step 0(c) classification runs
**before** any close path is entered, and `shipment-reconcile` SKILL.md:378 gives the
exhaustive, normative mapping:

> | Every feature member is a root, fully covered at every depth, set-equal to the manifest | `CASCADE` / `FULLY_COVERED_ROOT` |

This manifest matches on every conjunct: `025-F` declares no `parent_id` (**root**); its full
`parent_id`-graph descendant set is exactly `{025.001-T}`, which is a manifest member (**fully
covered at every depth**); and the manifest contains nothing beyond the qualifying root and its
descendants (**set-equal**). The classifier therefore returns **`CASCADE` / `FULLY_COVERED_ROOT`**,
and SKILL.md:576-577 binds the branch: *"On match the bound verdict alone selects the branch:
`CASCADE` enters the Cascade Close Sub-Procedure."* The Cascade Close Sub-Procedure
*"[r]eplaces steps 1–10 above entirely for this shipment's closure; there is no partial mixing
of the two paths."*

Consequently the **protected set does not arise at all** — not because it computes to ∅ under
safe-close step 2, but because safe-close steps 1–10 never execute. SKILL.md states this
directly: a manifest that qualifies for `CASCADE` *"has no protected set by construction (full
coverage is itself a Step 0(c) precondition), so no protected set arises."* The entire class of
P-015 failures quoted in §2.4 — the baseline-integrity-gate halt, the verify-after-each
invariant, the "protected-set member already archived or missing ⇒ pre-existing cascade ⇒ halt
closure" terminal — is **structurally unreachable for this manifest**, which is the strongest
available closure guarantee and is exactly the guarantee `017-S` relied on.

The closure gate that *does* apply is the cascade path's two-set `allowed_ids` / `required_ids`
post-condition (workflow-policies Amendment Log 1.21.0–1.23.0): fail on
`archived_ids − allowed_ids` (unexpected artifact) and fail on `required_ids − archived_ids`
(required artifact not archived), evaluated as two separately-labelled, independently-failing
conditions. For this manifest `required_ids = {025-F, 025.001-T, shipment record}` — the
shipment record and a qualifying feature member are `required_ids` members **unconditionally**
(1.22.0/1.23.0) — and there is no third artifact to appear as unexpected.

**Precedent.** Already proven in practice: the reduced `017-S` closed over a
fully-covered root with executable set exactly `[018.008-T]`, merged at `5d69c727`
(PR #58), post-merge closure merged at `b0b1868` (PR #59).

**Cost, recorded honestly.** More features, more shipments, more closure cycles; and
each feature must be a **genuinely coherent release unit**, not a wrapper around one
task. §4 applies that test rather than assuming it.

### 3.2 OPTION 2 — machine-enforced one-queued-child serialization — **REJECTED**

Shape: one covering feature per strand; at most one child `queued` at a time, the rest
`blocked`, promoted one at a time.

**(a) Step 2 item 1 — PASSES.** With exactly one child `queued`, the feature-scoped
`queued` filter yields a batch of one. Option 2 is *not* refuted here.

**(b) Step 3 item 1 — the manifest membership is a forced dilemma, and both horns fail.**

* *Blocked siblings **in** the manifest.* `blocked` is absent from the exhaustive
  positive rule's KEEP set `{queued, active}` and from its reported-skip set
  `{archived ⇒ pre_archived_skipped, done ⇒ already_done}`. It therefore lands in
  "ANY OTHER … status is a FAIL-CLOSED HALT, never a skip". **Every** run of such a
  shipment halts before the ready queue is built. Fatal.
* *Blocked siblings **out** of the manifest.* Step 3 is satisfied, but the siblings
  become P-015 **protected-set** members via the "every task sharing the covering
  feature's hierarchy prefix whose ID is not in the manifest `items`" bullet — which is
  **not** conditioned on the partial-feature branch. This is survivable for the *first*
  shipment (the blocked siblings are still in `.backlogit/queue/`), but it forces the
  whole sequence onto the **sequence-aware exclusion** path at every subsequent closure,
  where a sibling archived by a predecessor shipment stays protected fail-closed unless
  that predecessor record carries **verified** `archived_status: shipped` provenance.
  Any gap, legacy normalisation, or ambiguity in that provenance halts closure with
  `HALT — cascade detected, revert required` on a backlog that is in fact correct.

**(c) P-015 closure — the covering-feature disposal is unsound either way.**

* If the covering feature is placed in the first manifest, closure archives it while
  three `blocked` children remain in `.backlogit/queue/`, **orphaning** them: Step 2
  item 1's feature-scoped selector and Step 3's `parent_id` feature resolution then both
  point at an archived parent for every subsequent task in the strand.
* If the covering feature is kept out of every manifest until the last, each shipment is
  a **genuinely task-only manifest**. The authorised closure verdict set is today exactly
  `CASCADE, SAFE_CLOSE, BLOCK` (`CLOSE_PATH_VERDICT` tokens, `shipment-reconcile/SKILL.md:375-376`;
  corrected at plan-review attempt 2, finding N15 — an earlier revision wrote `HALT`, which is
  the *action* the mode takes on a `BLOCK` verdict, not a verdict token) — `TASK_ONLY_FINALIZE`
  does **not exist** in any installed
  contract surface (verified: zero matches for `TASK_ONLY_FINALIZE` under `.github/`).
  It is the deliverable of `023-F`/`024-F`, which are **intentionally blocked** behind a
  FIRED TRIGGER defined as *"a genuinely task-only shipment that cannot be re-shaped to a
  covered root"*. Option 2 would manufacture exactly that class of shipment as its normal
  operating mode. Option 1 re-shapes every unit to a covered root and therefore
  **provably does not fire that trigger**, preserving the intentional block.

**(d) No compensating benefit.** Option 2 does not reduce shipment count: a task cannot
be promoted until its predecessor has shipped, so it needs the **same N shipments** as
Option 1 while additionally requiring blocked-sibling manifest exclusion, protected-set
growth, and the sequence-aware provenance path. It is strictly more machinery for
strictly equal throughput. **Rejected.**

### 3.3 OPTION 3 — reviewed shipment-scoped harness contract amendment — **REJECTED FOR NOW**

Shape: amend P-002/P-004, Ship Step 2 and Step 4.3, and the `harness-architect` skill so
the harness batch and the green gate are scoped to the **shipment manifest** rather than
the covering feature.

Option 3 is the only option that would *also* close the Step 2 / Step 0.5 status gap
described in §5.2, and it is not rejected on merit. It is rejected **for this decision**
because:

1. It amends the rule that **detects** the problem. Step 4.3's full-suite gate is the
   workspace's broadest regression detector; narrowing it to the manifest trades a real
   safety property for convenience, and a mis-scoped amendment silently stops detecting
   cross-task regressions.
2. `021.001-T` **AC-3** requires that any contract amendment be *"scoped, deliberated and
   reviewed as its own work item **before any deferred shipment is unlocked**"*. Choosing
   Option 3 therefore unlocks **nothing** this session and defers all downstream work
   behind a second full deliberation/review cycle.
3. Option 1 already discharges the deadlock **with no amendment at all**. Amending a
   safety contract to solve a problem that a decomposition already solves is unjustified.

Option 3 is recorded as the **standing remedy of record** for the residual defects in
§5, to be picked up as its own deliberated, separately reviewed work item — not here.

## 4. Coherence test for the chosen shape (the Option 1 cost, applied)

Option 1 is only admissible where the single-task feature is a genuine release unit.
Applying that test:

| Strand | Verdict |
|---|---|
| **`020-F`** (5 tasks: fixture corpus → stubbed driver → detector → CI wiring → mutation proof) | **NOT decomposable now.** These are fragments of one mechanism; a fixture corpus with no detector, or a driver over a stub, is not independently releasable. Decomposing it under Option 1 would produce exactly the wrapper-around-one-task that Option 1 forbids. `020-F` **stays blocked**, unchanged. |
| **`019-F`** (4 tasks, distinct contract surfaces) | **Decomposable at the dependency root only.** Dependency DAG measured from the backlog: `019.007-T → 018.008-T` (archived **done**); `019.008-T → 019.007-T`; `019.009-T → {018.008-T, 019.007-T}`; `019.004-T → {018.008-T, 019.007-T}`. `019.007-T` is the **unique** node whose every dependency is satisfied. |

`019.007-T` passes the coherence test on its own merits, not by convenience: it delivers
the **Stage handback record** — the single output the Orchestrator's Step 1.5 consumes.
Its own record states *"Without it the Orchestrator's handback requirement has no producer
and every pipeline run degrades."* A contract producer with a named external consumer is a
release unit. It is one file (`.github/agents/_stage.agent.md`), one step section, size **M**,
single skill domain (contract text). *(After the restructure below it carries the ID
`025.001-T`; `019.007-T` is retained in this artifact wherever the pre-restructure state is
being described.)*

**Restructuring — PERFORMED AND MEASURED (2026-09-18).** *(Corrected at plan-review attempt 1,
finding F1: the original text asserted this in the past tense before the mutation existed.)*
The following Stage mutations were executed against the live backlog and their results read
back, not assumed:

| # | Action | Measured result |
|---|---|---|
| 1 | Create covering feature, root, `status: queued` | **`025-F`** — "Stage handback record emission contract", `parent_id: null` |
| 2 | `backlogit adopt 019.007-T --parent 025-F` | **ID REMAPPED `019.007-T` → `025.001-T`**; `parent_id: 025-F`; tool reported `rewritten_artifact_ids: [018.008-T, 019.004-T, 019.008-T, 019.009-T]` |
| 3 | Verify dependency edges survived | `025.001-T → 018.008-T (blocks)`; and `019.004-T`, `019.008-T`, `019.009-T` each `→ 025.001-T (blocks)` — all four re-pointed automatically, none lost |
| 4 | Rewrite the task's stale `DEFERRED / BLOCKED / NO SHIPMENT / Do not execute` header | Withdrawn and replaced; the §5.3 pre-claim precondition and the U1 scope fence are now carried on the task record itself |
| 5 | Verify strand preservation | `019-F` **blocked** with `019.004-T`, `019.008-T`, `019.009-T`; `020-F` **blocked** (5 tasks); `023-F`, `024-F` **blocked** — all unchanged |

**The ID remap is load-bearing and was not anticipated.** Because `backlogit adopt` renumbers
the adopted artifact into its new parent's sequence, the task's ID now shares the covering
feature's prefix. `shipment-reconcile`'s ID-prefix covering-feature derivation and P-015's
`parent_id` walk therefore **agree on `025-F`**, and the divergence risk originally recorded in
§5.4 is **eliminated at the root** rather than merely tolerated. §5.4 is rewritten accordingly.

**Deliberately NOT yet performed.** `025.001-T` remains `status: blocked`. Moving it
`blocked → queued` is part of AC-5's post-review step — *"Only after AC-1 through AC-4, Stage
CREATES A FRESH SHIPMENT over the reviewed shape **and moves its member tasks to queued**"* —
and is sequenced after the review gate passes, together with shipment creation. It is recorded
here as an **owed, ordered Stage action**, not as a completed one:

> **A1 (Stage, post-review)**: `025.001-T` `blocked → queued`, then create the shipment over
> `[025-F, 025.001-T]`. Invariant **I2** (no manifest member in `blocked` status) is satisfied
> **by A1's ordering**: the task reaches `queued` before it is ever a manifest member.

`019-F` retains `019.004-T`, `019.008-T`, `019.009-T` and **remains `blocked`**; its remaining
tasks are each still gated on `025.001-T` by explicit, tool-rewritten dependency edges.
`023-F` and `024-F` are **untouched and remain blocked** on their own intentional
fired-trigger policy block.

## 5. Feasibility and risk — including the systemic `DB12DA37` defect

### 5.1 `DB12DA37` — P-002 claim ordering is structurally unsatisfiable

Active stash entry `DB12DA37` (`DEFERRED SCOPE EXPANSION`, priority medium,
`DISCOVERY-STATUS: CLEAN`), canonical-contract rows **D-D6**/**D-M8** of
`docs/plans/2026-09-16-017-S-closure-canonical-contract.md`:

> **P-002 is DISCLOSED AS UNSATISFIABLE and requires an EXPLICIT RECORDED OPERATOR AUTHORIZATION to proceed — disclosure alone is NOT sufficient.** … the shipment claim at S-3 auto-activates `018.008-T`, so P-002's "task claim after `harness-ready`" ordering cannot hold on any shipment-claim route. **P-002's Violation Action is literally "Halt and suggest running the harness-architect."**

> **D-M8** … it *"affects every shipment-claim route in this workspace"* and resolving it *"would require amending P-002 or changing claim semantics"*.

This defect is **option-independent**: it is a property of the shipment-claim route, not
of the harness batch size. Option 1 does **not** fix it and this artifact does **not**
authorise around it.

### 5.2 Sharpened consequence discovered during this deliberation — silent harness bypass

Composing `DB12DA37`'s measured auto-activation with §2.1's exact text yields a
consequence **not** previously recorded, and it is recorded here rather than absorbed:

1. Ship Step 0.5 item 1a passes (fresh `queued` shipment, `queued` task — no
   `SHIPMENT_STATE_INCONSISTENT`).
2. The Step 0.5.4 claim auto-transitions `019.007-T` `queued → active`.
3. Step 2 item 1 lists the feature's tasks **in `queued` status** — now **zero**.
4. Step 2 item 3 scaffolds nothing; item 4's *"confirm every queued task now carries the
   `harness-ready` label"* is **vacuously true** over the empty set, so its halt never fires.
5. Step 3 item 1 keeps `active` members, so the executable set is `{019.007-T}` and
   Step 4 executes a task **that was never harnessed**.

The result is a **P-004 red-phase bypass**, materially worse than the P-002 ordering
deviation alone. Note the internal inconsistency it exposes: Step 3's rule deliberately
tolerates `active` (KEEP), while Step 2's selector does not — the two steps disagree about
the post-claim status of the same task.

**Why the `017-S` precedent does not cover it.** `017-S` shipped safely because
`018.008-T` **already carried the `harness-ready` label** at claim time
(`.backlogit/archive/018.008-T.md:17`), so Step 2's partition routed it to
*"Already harnessed … skip these"* and no bypass occurred. `019.007-T` carries no such
label — its own notes read *"HARNESS: deferred"*. The precedent's safety therefore does
**not** transfer, and asserting it would have been the silent authorisation the operator
prohibited.

### 5.3 Mitigation carried into the plan — contract-conformant, no amendment

The shipment is created carrying a **named, recorded, verifiable PRE-CLAIM PRECONDITION**:

> `019.007-T` MUST carry the `harness-ready` label, produced by an actual
> `harness-architect` run, **BEFORE** the shipment is claimed. If it does not, the
> shipment MUST NOT be claimed.

This mitigation amends nothing. *(Corrected at plan-review attempt 1, finding F3: the original
text claimed the mitigation works by "routing the task into Step 2's *Already harnessed*
partition". That mechanism is false — on the claim route the task is absent from Step 2 item 1's
`queued`-only list entirely, so it is never partitioned at all.)*

**The mitigation's actual mechanism.** It does not repair Step 2; Step 2 remains vacuous on the
claim route. It works by making Step 2 **irrelevant**: a real, red, `harness-architect`-produced
harness and its `harness-ready` label already exist on the task *before* execution begins, so the
property P-004 actually protects — *"every generated test function red before implementation"* —
holds on the artifact itself rather than on Step 2's bookkeeping. Step 4 then executes a task
that **is** harnessed, which is the outcome §5.2's bypass would otherwise deny. It additionally
restores P-002's ordering on its own terms (harness-ready precedes the claim), and it is
precisely the direction `DB12DA37`'s own *"SCOPE WHEN PICKED UP"* field nominates:
*"harness-ready required before the SHIPMENT claim for any manifest containing unimplemented
tasks"*.

**What it does not fix.** Step 2's vacuity on the claim route is untouched, and item 4's
confirmation still passes vacuously. Repairing the *contract* — so Step 2's selector is
manifest-scoped and status-inclusive like Step 3's — is squarely Option 3's amendment scope
(§3.3) and is explicitly not attempted here.

**Authority boundary.** Satisfying the precondition is an **operator** action: the operator
invokes the `harness-architect` skill directly, as the always-available substitute actor. No
installed surface defines a pre-claim harness slot for an *agent* (P-004 applies to `ship` via
Step 2, which is post-claim; P-010 forbids Stage), and that gap is recorded as residual **R-C**
in §5.4. Stage neither performs the run nor authorises the residual P-002 deviation; Stage
records the precondition on the shipment and halts there. The residual P-002 deviation, if the
operator elects to proceed without satisfying the precondition, still requires the explicit
recorded operator authorisation of the `D-D6` precedent, with **absence, ambiguity, or silence ⇒
HALT, zero mutation**.

### 5.4 Residual risks after the measured restructure

**Formerly-recorded risk, now ELIMINATED — ID-prefix vs `parent_id` derivation.** The original
draft of this artifact recorded a divergence risk: the task would retain the ID `019.007-T`
under `025-F`, so `shipment-reconcile`'s ID-prefix heuristic (*"a task `055.002-T` belongs to
feature `055-F`"*) would derive `019-F` while P-015's fully-covered-root check walks the
`parent_id` graph and derives `025-F`. **Measurement refuted the premise**: `backlogit adopt`
renumbered the artifact to **`025.001-T`** (§4), so both derivations now resolve to `025-F`.
The risk is closed at the root, not tolerated. The underlying *contract* divergence between the
two surfaces is real but is no longer reachable from this manifest; it is carried as candidate
scope for the Option 3 amendment item (§3.3).

**Live residual R-A — the closure path is CASCADE, and it is destructive.** Per §3.1, this
manifest classifies `CASCADE` / `FULLY_COVERED_ROOT`, which enters the Cascade Close
Sub-Procedure and invokes the destructive `backlogit_ship_shipment` engine operation. This is
the *correct and intended* path for a fully-covered root — it is the same path `017-S` used —
but it is **irreversible** and must be classified and approved as destructive, not waved
through as "safe-close". The plan carries this as `ActionRisk: destructive` with a recorded
operator approval requirement and a backlog-state unwind procedure distinct from `git revert`.

**Live residual R-B — P-001 single-release-unit.** `021-F` and `021.001-T` are `active` while
this decision is being authored. P-001's gate point is Ship **Step 1** (pre-flight), which runs
**after** the Step 0.5.4 claim — so a stale `active` top-level unit would let Ship take the
one-way-door claim and *then* halt. Mitigation: Stage completes `021.001-T` and `021-F` to a
terminal state in this same session, **before** the shipment is claimable, and the plan adds
P-001 to its pre-claim environment prechecks.

**Live residual R-C — no contractual slot for the pre-claim harness run.** §5.3's mitigation is
substantively sound but has no *installed* contract slot: P-004 applies to `ship` via the
`harness-architect` skill, and Ship's only invocation of that skill is Step 2, which is
post-claim. P-010 forbids Stage from running it. The executor is therefore named explicitly as
the **operator**, invoking the `harness-architect` skill directly — the always-available
substitute actor — with the gap itself recorded as candidate scope for the Option 3 amendment
item. Naming an unowned action would have left the plan's hardest gate assigned to no one.

## 6. AC disposition

Dispositions are recorded against **measured** state. AC-4 and AC-5 are sequenced and were
**not** claimed before they were true *(corrected at plan-review attempt 1, finding F5: the
original table marked both ✅ while the review that AC-4 depends on was still running, which
was circular).*

| AC | Disposition |
|---|---|
| **AC-1** | This artifact frames the problem (§1), evaluates all three admissible options (§3), records **Option 1** with rationale (§3.1) and rejected-option reasoning (§3.2, §3.3). ✅ |
| **AC-2** | Option 1 proven against Ship Step 2 item 1 (both routes, §3.1), Ship Step 3 item 1 including the fail-closed status rule (§3.1), Step 4.3 (§3.1), and P-015 / `shipment-reconcile` closure (§3.1, corrected to the `CASCADE` / `FULLY_COVERED_ROOT` path the classifier actually returns) — with exact text quoted from the installed surfaces at `1bef819`. ✅ |
| **AC-3** | Option 1 requires **no** contract amendment. Option 3's amendment is explicitly deferred to its own scoped, deliberated, separately reviewed work item and unlocks nothing here (§3.3). The two contract gaps this decision surfaced — Step 2's claim-route vacuity (§5.2) and the absent pre-claim harness slot (§5.4 R-C) — are routed to that item, not fixed here. ✅ |
| **AC-4** | `impl-plan` → `plan-harden` (REQUIRED) → `plan-review` executed for the unlocked unit (`docs/plans/2026-09-18-intercom-go-stage-handback-record-plan.md`). Review attempt 1 returned **FAIL** (3× P0, 4× P1); both artifacts were remediated and re-submitted. Attempt 2 returned **FAIL** (5× P0). **NOT ACHIEVED** — the process ran in full and to contract, but no passing or acceptable verdict was obtained. See §9. ❌ |
| **AC-5** | Gated on AC-4. AC-4 did not pass ⇒ **correctly NOT performed**. No shipment exists; action **A1** (`025.001-T` `blocked → queued`) was **not** run, so the member task remains `blocked` and cannot be claimed. This is the compliant outcome, not an omission. ⊘ |
| **AC-6** | No retired rev-11…rev-15 mechanism (harness-state manifest, terminal witness, `-gatetask` selector, surface-predicate activation, MP0–MP6) is reintroduced, in either explicit or disguised form. None was considered; the historical appendix was read and the mechanisms remain NON-GOVERNING provenance. **No re-deliberation is requested.** ✅ |
| **AC-7** | The chosen option is recorded as exactly one of the three enumerated: **(1) single-task features/shipments**. No fourth mechanism is proposed. ✅ |

## 7. Resulting shipment shape — **DESIGNED, NOT INSTANTIATED**

**No shipment was created.** This section records the shape Option 1 *proves correct*, which
remains the target shape once §9's blockers are cleared. Every row below describes the
intended manifest, not live state. The two status preconditions the proofs in §3.1 assume
(`025.001-T` at `queued`, `021-F`/`021.001-T` terminal) are called out explicitly.

| Field | Value |
|---|---|
| Covering feature | **`025-F`** — Stage handback record emission contract (root, `parent_id: null`, `status: queued`) — **created** |
| Member task | **`025.001-T`** — Emit Stage handback record with concrete artifact paths (`size: M`, complexity `medium` as prose — see degradation note) — **adopted, ID remapped from `019.007-T`**; **live status `blocked`**, because action A1 was correctly not run (AC-5). The §3.1 Step 3 proof assumes `queued`; that premise is **unmet in the terminal state** and becomes true only when A1 runs after a passing review. |
| Manifest | `[025-F, 025.001-T]` — fully-covered root, set-equal |
| Executable set (Ship Step 3 item 1) | `{025.001-T}` — exactly one; artifact-type filter removes `025-F` before any status read |
| Closure classification | **`CASCADE` / `FULLY_COVERED_ROOT`** (`shipment-reconcile` SKILL.md:378). Cascade Close Sub-Procedure replaces safe-close steps 1–10 entirely; **no protected set arises by construction**. Closure gate is the two-set `allowed_ids` / `required_ids` post-condition. |
| Dependencies encoded | `025.001-T → 018.008-T` (satisfied, archived done); `019.004-T`, `019.008-T`, `019.009-T` → `025.001-T` (retained, cross-feature, tool-rewritten) |
| **Pre-claim precondition (hard gate)** | `025.001-T` must carry `harness-ready` from an actual `harness-architect` run **before** the shipment is claimed (§5.3). Executor: **operator**. |
| **Pre-claim precondition (hard gate)** | P-001: `021-F` / `021.001-T` must be terminal before the claim (§5.4 R-B) |
| Sizing degradation | `size: M` is a structured field; **`complexity` is carried as prose only** — `backlogit update --complexity` rejects it: *"artifact type `task` does not define a complexity field"*. Recorded per the structured-emission capability gate. |

## 8. Scope preserved

* `019-F` — remains **blocked** with `019.004-T`, `019.008-T`, `019.009-T`.
* `020-F` — remains **blocked**, untouched; fails the Option 1 coherence test (§4).
* `023-F`, `024-F` — remain **blocked** on their intentional fired-trigger policy block.
  Option 1 provably does not fire that trigger (§3.2(c)).
* No push, no PR, no claim, no build. Stage role boundary held throughout.

## 9. TERMINAL OUTCOME — **DECISION MADE, UNIT BLOCKED, NO SHIPMENT**

### 9.1 What is settled

The **decision is complete and stands**: Option 1 (single-task features/shipments) is the
chosen mechanism, proven against all four contract surfaces AC-2 names. Two independent
adversarial review rounds contested this decision and **upheld it on every analytical point
they tested** — the `CASCADE`/`FULLY_COVERED_ROOT` classification, the I2 `blocked → queued`
ordering, the non-firing of the `023-F`/`024-F` task-only trigger, AC-6's no-retired-mechanism
assertion, and the disclosure (not authorisation-around) of `DB12DA37`. Attempt 2's reviewers
independently re-measured the live backlog and confirmed the restructure as recorded in §4.

### 9.2 What is blocked, and why it is a contract gap rather than an artifact defect

`plan-review` returned **FAIL** on both permitted attempts. The attempt-2 P0 findings are not
corrections to the reasoning above; they are **structural gaps in the installed execution
surface** that make the downstream unit unsafe to claim *however it is planned*:

| ID | Blocker | Why it cannot be remediated inside this unit |
|---|---|---|
| **B1** (N6) | **No execution slot for the pre-claim harness.** §5.3's mitigation requires `harness-architect` to run and `harness-ready` to be applied *before* the claim, because the claim auto-activates the task and Step 2 item 1 lists only `queued` tasks (§5.2). | No installed surface defines where those red harness files live. Uncommitted on `main` trips Ship Step 0.5 item 3a's dirty-tree halt; committed to `main` places a red harness on the default branch outside any PR gate; an operator branch trips `BRANCH_MISMATCH`. Every route fails a gate. The mitigation has nowhere to stand. |
| **B2** (N14) | **No Stage→`main` artifact handoff.** Ship Step 0.5 item 3a checks out `main` and branches from it, so Stage's `.backlogit/` mutations on `chore/stage-*` — the shipment record and the `queued` status — are absent from Ship's branch. | **Circular.** The missing staging-artifact handoff is precisely the defect `025.001-T` was written to make specifiable. The unit cannot be delivered through the gap it exists to close. |
| **B3** (N7) | **Wrong policy on the escape route.** The residual-deviation authorisation was scoped to `P-002` while the actual consequence of an unharnessed execution is a **`P-004`** red-phase bypass; the route also contradicts the plan's own invariant I6. | Re-scoping the authorisation to P-004 would be Stage authorising a red-phase bypass — outside Stage's role boundary and contrary to AC-3's no-amendment constraint. |
| **B4** (N3/N4/N11) | **Closure procedure not executable as specified.** Unbound cascade (`RECONCILE_FAIL_CASCADE_UNBOUND`), missing `returned_ids == []` assertion, undefined `allowed_ids`, and an agent-driven classifier whose self-computed digest provides no real drift protection. | The first three are plan defects and *are* fixable; the fourth is a property of the installed skill (no executable classifier exists in this repo) and is not. |
| **B5** (N8) | **Unsatisfiable edit placement.** In `_stage.agent.md`, `#### Pre-Summary Verification Gate` is immediately followed by the summary bullets with **no intervening heading**, so inserting a `####` heading between them necessarily re-parents those bullets. | The task's own placement requirement is unsatisfiable as written without a restoring heading — a scope change the fence forbids. |
| **B6** (N13) | **Stage Step 5.5 item 3 scope guard.** `add_to_shipment` may only be called for IDs the immediately preceding harvest returned; `025.001-T` is an **adopted pre-existing queue item**, not a harvest output. | Stage's own contract forbids assembling the designed manifest from an adopted item. |
| **B7** | **`DB12DA37` unresolved.** The systemic P-002 claim-ordering defect remains open and un-authorised. | Explicitly out of scope: this decision **discloses** it (§5.1–§5.2) rather than authorising around it, per the operator's instruction. |

### 9.3 Disposition

Per the operator's standing instruction — *"if no option yields a safely claimable
implementation unit, complete the decision task with a precise blocked outcome and create no
shipment"* — the terminal outcome is:

* **`021.001-T` → `done`**, with this artifact as its deliverable. **`021-F` → `done`.**
  Completing them also discharges risk R-B (§5.4): P-001 would otherwise fire at Ship Step 1
  *after* the one-way-door claim.
* **`025-F` → `blocked`**, **`025.001-T` remains `blocked`**, both carrying B1–B7 in-record.
* **No shipment created.** Action **A1 not run**.
* `019-F`, `020-F`, `023-F`, `024-F` — **unchanged**, all still `blocked`.
* **P-005 recorded** for the two-attempt `plan-review` FAIL (Stage Step 4 option (c)).

### 9.4 Unblock path — owner: a future Stage session, not Ship

B1 and B2 are **contract-surface gaps** belonging to the **Option 3 amendment work item**
(shipment-scoped harness selection) and/or to `019-F`'s own persistence-route scope. They must
be deliberated, planned and reviewed on their own terms before `025-F` can be unblocked. B3–B6
are then remediable inside a re-planned `025.001-T`. Until B1 and B2 are closed, **no shipment
may be formed over `025-F` and it must not be routed to Ship.**

### 9.5 Known residual

`019-F`, `019.004-T`, `019.008-T`, `019.009-T` and `020.004-T` retain **prose** references to
the pre-adoption ID `019.007-T`, which no longer resolves. Their **structured** dependency
edges were rewritten correctly by `backlogit adopt` and were verified. The ID mapping
`019.007-T → 025.001-T` is recorded here, on `025-F` and on `025.001-T`. Not remediated in
this session: all five records are `blocked` and unroutable, and rewriting their prose risks
damaging unrelated content for no live benefit. Flagged for the next session that touches the
`019-F` strand.

