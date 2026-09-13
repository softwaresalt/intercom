---
title: "Plan — Ship covering-feature completion and close-path delegation (Foundation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 3
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

## 4. Exact inventory — 6 clause sites (all MUST) + 2 additive

Line anchors **re-verified against the installed file at `e7e981d`**; all six match rev 6
Inventory Table §6.1-C exactly.

| # | Line | Clause | Change |
|---|---|---|---|
| **C1** | 756–764 | Mandatory pre-self-close context reload — names only Ship instructions + `shipment-reconcile` | Also reload **P-015**; positively verify merged tokens **and the merge commit**; add the §6 authority-escalation carve-out with the **pinned reconciled wording** of §4.2 |
| **C2** | 789–791 | safe-close summary prescribing `backlogit move <shipment_id> --status shipped` | Correct: the engine **refuses** this for shipments (exit 9, Probe 18). Replace with a pointer to the skill's authoritative step 8 |
| **C3** | 794–795 | *"**Do NOT call** `backlogit shipment ship` … **unless** the P-015 VERIFIED FULLY-COVERED-ROOT EXCEPTION applies"* | Replace the named single exception with: do not call it unless **the machine-selected verdict** authorizes it |
| **C4** | 802–823 | the re-derived classification prose (binary selector + drifted `children` rule) | **Delete the restatement.** Replace with generic delegation |
| **C5** | 821–823 | *"invoke the cascade … **in place of** the safe-close sequence above"* | Subsumed by C4's delegation block; single-exception framing removed |
| **C6** | new bullet | — | **NEW** delegation bullet: Ship invokes whichever close path the verdict names, and **no** path absent from the verdict |

| Additive | Content |
|---|---|
| **ADD-1** | Role Boundary table row (line 38) — narrow feature-completion authority |
| **ADD-2** | New **Step 6.1(a1)** — covering-feature completion gate |

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

### 5.1 Step 6.1(a1) applicability — three outcomes, stated exhaustively (rev 2; re-scoped rev 3)

Rev 1 specified what a1 does when it applies, but never said what a1 does when the
manifest has **no** feature member. An unspecified step invites an implementer to either
skip the gate silently or invent a transition. Both are wrong. The step MUST be written
with an explicit three-way outcome:

| Case | Condition | Outcome |
|---|---|---|
| **(i) NOT APPLICABLE** | The manifest contains **zero** feature members | **Explicit no-op.** a1 records `A1_NOT_APPLICABLE: no feature member in manifest` and proceeds directly to Step 6.1(a). This is a **successful** outcome, not a skip and not a failure |
| **(ii) APPLIES — proceed** | The manifest contains a feature member and **all five** §5 conditions hold for it | Perform `active -> done`; record the transition |
| **(iii) APPLIES — halt** | The manifest contains a feature member and **any** §5 condition is unmet | **HALT, fail closed.** Surface which condition failed. Do **not** proceed to Step 6.1(a). Do **not** widen any condition in-flight |

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

**The distinction between (i) and (iii) is load-bearing.** Case (i) is "there is nothing
here for this gate to do". Case (iii) is "there is something here and it is not in the
required state". Collapsing them — treating an unmet guard on a real feature member as a
benign no-op — would let a shipment close with a covering feature left `active` or with
descendants unfinished, which is precisely the corruption the gate exists to prevent. A
zero-feature-member manifest must never be reported as a guard failure either: that would
make any task-only shipment un-closable at the gate.

Asserted by **AC-15** (case i) and **AC-16** (case iii).

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

**Contract text and ordering on `_ship.agent.md` only.** **13 rows**:

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

Rows 7–10 are the **widening guards**: they fail if this change authorized anything new.
Row 11 is the **coherence guard**: it fails if C1 ships with the carve-out and the
"close under the just-merged contract" directive unreconciled.
**Rows 12–13 are the applicability guards** (rev 2): they fail if a1 ships without an
explicit not-applicable outcome — which would leave any task-only manifest ambiguous at
the gate — or without idempotent resume, which would make a partially-completed A
permanently unclosable.

> **Rev 3 note.** The row count is **unchanged at 13**, and every row's assertion text is
> unchanged. Only the *rationale* for row 12 is re-scoped: rev 2 justified it by `017-S`
> being task-only, which is no longer the case (§5.1). Row 12 now stands on **defensive
> totality** of the step. The Go test is **not** modified by rev 3.

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
| 9 | **S2** | **Fresh session.** Full checkpoint-recovery protocol (§9.1) — enumeration, anomaly gate, explicit owner selection, operator confirmation, same cursor, resolve-after-resume. Loads the new Role Boundary + Step 6.1(a1) from merged `main` |
| 10 | S2 | Verify the merge SHA is an ancestor of `origin/main`; validate the **same** active shipment `021-S` |
| 11 | S2 | **Step 6.0 Post-Merge Branch Protocol** — `git checkout main`, `git pull`, `git checkout -b post-merge/022-ship-feature-completion`. **Created BEFORE any post-merge backlog mutation**, because Step 6.1(e) commits `.backlogit/` and those commits must not land on `main` |
| 12 | S2 | **Step 6.1(a0)** `--phase lifecycle` topology gate — run **while `021-S` is still `active`**. *(Probe 20 tripwire: post-archive this gate fails closed on every route; it must never be re-run after the close.)* |
| 13 | S2 | **Step 6.1(a1)** — §5.1 case **(ii)**: `022-F` is a manifest member and all five §5 conditions hold → `022-F active -> done` |
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

### 9.1 S2's checkpoint recovery is the FULL protocol, not a resume shortcut (rev 2)

Rev 1 said only that S2 *"restores the checkpoint"*. That is the step most likely to be
short-circuited, so it is specified in full. S2 MUST:

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
6. **Resume the same cursor** — the **same** single active shipment `021-S`. No parallel
   resume, no new worktree (P-001/P-016).
7. **Resolve only after a confirmed successful resume** — `backlogit_resolve_checkpoint`
   for that **one** selected checkpoint only. No bulk sweep, never a `stage`-owned record.
8. **Fail closed with no fresh-start fallback** — an invalid, torn, or unreadable
   checkpoint halts to the operator. S2 MUST NOT discard it and begin fresh work.

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

> **Verification scope of this claim — stated honestly (rev 3, internal review P2).**
> The `CASCADE` result above is established at the **classification** level: the manifest
> shape provably satisfies P-015's fully-covered-root preconditions, and pre-mode provably
> accepts pre-archived members. It is **not** established at the **engine-execution**
> level *for this exact member shape*. A's own closure (§9) archives a clean root with
> **no** truly-archived members; Probe 15 measured the **old** root-*excluded* 12-member
> shape, which is the inverse case. **No committed probe or Go test exercises a
> root-*included* cascade whose siblings are genuinely `status: archived`.**
>
> This is a **confidence** residual, not a safety one. The path is **fail-closed** at two
> independent gates — the skill's cascade step 2 HALTs on non-empty `returned_ids`, and
> step 3's two-set gate HALTs on any required/returned mismatch — so a mis-derivation
> cannot silently corrupt the backlog. The realistic downside is that `017-S` **HALTs**
> and falls back to `SAFE_CLOSE`, re-exposing the exit-9 blocker. That outcome returns
> `017-S` to **Stage**, costs no data, and is recoverable.
>
> **Not a current blocker**, and deliberately **not** charged to A: `017-S`'s closure is a
> separate shipment executed after A ships. Recorded as a follow-up — a root-included
> cascade fixture probe against the exact 13-member shape — to be run before `017-S` is
> routed, **not** before A is.

> **Deliberately OUT OF A's INVENTORY — `_ship.agent.md` pre-mode summary (rev 3,
> internal review P3).** The installed Step 6.1(a) summary at **L776–777** says pre-mode
> *"verifies that every manifest item is present in queue with `status: done`"*. On
> `017-S`'s route that sentence is **literally false** — 11 of 13 members live in
> `.backlogit/archive/`, and `018-F` relocates there once a1 completes. It is a further
> instance of exactly the duplication defect §2.2 describes.
>
> It is **NOT added to A's clause inventory**, for two reasons. **(1) Budget.** A's
> inventory is closed at **6 sites** and its budget is pinned at **~78 min**; the site
> falls in the gap between C1 (756–764) and C2 (789+) and would open a seventh. **(2) It
> is not load-bearing.** The authoritative `shipment-reconcile` skill — which C1–C6 make
> Ship delegate to unconditionally — already classifies archived members as `pre-archived`
> and accepts them without a status check, so the **runtime outcome is correct
> regardless**. The defect is descriptive drift in a summary, not an operative gate.
>
> Recorded as a follow-up for a future `_ship.agent.md` pass. A reviewer MUST NOT fail A
> for its absence.

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
| Pre-mode returns `RECONCILE_FAIL` | HALT; do not proceed to close; surface the report |
| Classification returns `SAFE_CLOSE` instead of `CASCADE` | HALT — indicates the manifest is not the fully-covered root this plan asserts. Return to Stage |
| Cascade archives anything outside `allowed_ids` | P-015 violation action: `git restore`, re-verify protected set, HALT |
| S1 attempts the a1 gate or the close after its own merge | **HALT** — §6 forbids it; S1 must checkpoint and end. Return to Stage; the two-session property has been violated |
| S2 recovery hits a validation/quarantine anomaly, ambiguity, or a `stage`-owned record | **FAIL CLOSED** to operator handoff (§9.1). No restore, no resume, no resolve |
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
  precede S2's resume. **S2 executes the full §9.1 recovery protocol** — enumeration
  without status/agent filter, anomaly gate first, explicit owner selection even for a
  single candidate, `agent: ship` ownership validation, operator confirmation before
  restore, same cursor, and resolve-only-after-confirmed-resume. *Residual, stated
  honestly:* the artifact is not tamper-proof — the two-session property ultimately rests
  on process integrity, and the checkpoint is evidence, not enforcement.
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

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity is carried as **enum-validated prose**: this
workspace's `header-def.yaml` defines `size` on the task type but **no** `complexity`
field (confirmed at `.backlogit/header-def.yaml:144–166`, and in
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`). Size is structured.

| Work item | Estimate |
|---|---|
| Role Boundary grant (ADD-1) | ~8 min |
| Step 6.1(a1) parent-completion step (ADD-2), incl. §5.1 three-way applicability + §5.2 idempotent resume | ~15 min |
| C1 reload extension | ~4 min |
| C2 safe-close summary correction | ~4 min |
| C3 + C4 + C5 generic delegation | ~12 min |
| C6 delegation bullet | ~4 min |
| Go table-driven test (**13** rows) + ≤4 helpers | ~21 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~78 min ≈ 1.3 h** |

Rates are rev 6 §10.1's own, so this is directly comparable to the measurement that
produced `MUST_REPLAN`. Margin to the 2-hour rule: **~42 min**.

*(Rev-2 delta: ADD-2 **+5** for the §5.1/§5.2 applicability and idempotence text; Go test
**+2** for rows 12–13. Total ~71 → ~78 min, margin ~49 → ~42 min.)*

*(Rev-3 delta: **none**. The A-only simplification changes no surface, no clause site and
no test row, so the budget stands at **~78 min** with a **~42 min** margin. Rev 2's
comparative remark that A was *"the second-most comfortable of the three"* is withdrawn —
there is no longer a set of three to compare against.)*

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

**Hardening 6 — what could make this plan wrong.** If `expected_status: done` is not
actually enforced over feature members, the a1 gate is unnecessary (harmless but
pointless). If claim already sets features `active`, PO-1 passes trivially. Neither
falsifies the delegation fix, which stands on §2.2 alone. The plan is therefore
**falsifiable in both directions** and its two deliveries are **independently
justified**.
