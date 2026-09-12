---
title: "Decision — Deferred-work split, shipment-free readiness, and 017-S closure repair"
date: 2026-09-12
revision: 2
status: decided
agent: Stage
governs: "features 018-F / 019-F / 020-F / 021-F; shipment 017-S (018-S / 019-S / 020-S archived)"
supersedes_scope_of: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
strand_a_plan: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
strand_b_plan: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md
---

# Decision — Deferred-work split, shipment-free readiness, and `017-S` closure repair

**Revision 2 (rev-18 remediation cycle, 2026-09-12).** Second bounded Stage remediation cycle,
against the adversarial re-review of `9e44bfe` (branch `chore/stage-pipeline-policy-gap`). Verdict
was **MUST_REMEDIATE**, not `MUST_REPLAN`: the one-task `017-S` decomposition and the two cohesive
future feature strands are **preserved**; only mechanics were corrected. Rev 1 of this artifact is
in git at `9e44bfe`.

`017-S` was **not widened**. Its executable set is still exactly one task, `018.008-T`.

**What rev 2 changed, in one line each:**

1. `017-S`'s manifest became **task-only (12 entries)**; `018-F` was removed so Step 6 pre-mode can
   ever pass (§1).
2. The safe-close **step-8 shipment-record closure conflict** was traced, proven tool-blocked, and
   escalated as a **P0** rather than asserted away (§1.6).
3. The interim persistence route became **operator-only**, and two rev-17 Orchestrator-authority
   claims were withdrawn as factually wrong (§2).
4. The **invalid `blocked` shipment lock** was retired: `018-S`, `019-S`, `020-S` archived; deferred
   features are `blocked` with **no shipment** (§4).
5. The harness was made **P-004-compliant unconditionally** (mandatory marker, explicit H0
   evidence), and Step 1.9's algorithm was tightened to exact anchors, P-016 worktree classification,
   and a branch ownership/identity discriminator (§9).


## 1. Shipment `017-S` closure — the current blocker, re-diagnosed and re-repaired

> **Rev 18 supersedes the rev-17 finding.** The rev-17 repair (a 13-entry `CASCADE` manifest
> including `018-F`) fixed the *protected-set* defect but introduced an **earlier, fatal** one: the
> shipment could never reach closure, because Ship Step 6's pre-mode gate would halt first.

### 1.1 What rev 17 got right, and what it missed

Rev 17 correctly established that the original two-entry manifest `[018-F, 018.008-T]` was
unclosable: it classified `SAFE_CLOSE`, the 11 archived non-manifest descendants of `018-F` entered
the **protected set**, and safe-close step 3's baseline integrity gate — which grants the protected
set **no pre-archived exemption** — would have halted with
`HALT — cascade detected, revert required`.

Rev 17 then added the full descendant closure so the classifier would return `CASCADE`. That is
where it missed: **including `018-F` makes `018-F` a manifest ITEM**, and Ship Step 6 item 1.a
invokes `shipment-reconcile` with `mode: pre` and **`expected_status: done`**.

### 1.2 Proof that the rev-17 shape could never close

Pre-mode step 3 checks `.backlogit/queue/` **first**; when the record is found there it compares the
declared `status` to `expected_status`, with **no artifact-type exemption** (the task-artifact
filter applies only to step 5's record-status classification).

At Ship Step 6:

| Fact | Evidence |
|---|---|
| `backlogit shipment claim` moves the covering feature `queued → active` | `.backlogit/logs/016-F.jsonl`: `status_changed … reason: "shipment claimed" … to: active` |
| **No Ship step moves a covering FEATURE to `done`** | Ship Step 4.5 item 3 moves *the task*; no step in `_ship.agent.md` touches the feature |
| backlogit does **not** auto-complete a feature when its last task completes | **Verified by execution**, backlogit 1.10.1, isolated probe workspace: after `move <task> --status done`, the covering feature remained `status: active` |
| Therefore `018-F` is live in `queue/` with `status: active` at pre-mode | → classification `status-mismatch` → `HALT — operator reconcile required` |

The working precedents in this workspace (`006-S`, `009-S`, `014-S`) all had the covering feature
already `done`/archived **before** pre-mode ran — produced by an **undocumented session step that no
installed contract specifies**. This decision **does not rely on it**, per the remediation
directive.

### 1.3 The rev-18 repair: a task-only 12-entry manifest

`018-F` was removed from the manifest with the **registry-declared supported operation**
`backlogit shipment return-blocked` (registry op `return_blocked`, MCP `backlogit_return_blocked`),
then restored with `backlogit move 018-F --status queued`, which also cleared the `blocked_reason`
the removal stamped. **No hand-edit of `custom_fields.items` was required or performed** — the
narrow frontmatter fallback used in rev 17 for `018-S` was unnecessary here.

Final manifest (12 entries, task-only):

```text
018.008-T                                   (live, the sole executable task)
018.001-T  018.001.001-ST  018.001.002-ST  018.001.003-ST
018.002-T  018.002.001-ST  018.002.002-ST  018.002.003-ST
018.003-T  018.003.001-ST  018.003.002-ST
```

Ship Step 0.5 item 2 **expressly accepts task-only manifests** and resolves the covering feature
through each task's `parent_id` (the 097-S precedent).

### 1.4 Full pre/post trace — every gate, in order

| Gate | Evaluation | Outcome |
|---|---|---|
| Step 0.5 item 1a queued-with-active-work | task-artifact members: `018.008-T` `queued`, `018.001/2/3-T` `archived`; none `active`/`done` | **PASS** |
| Step 0.5 item 2 membership | task-only manifest accepted; covering feature via `parent_id` | **PASS** |
| Step 0.5 item 3 covering parent | all four tasks declare `parent_id: 018-F`; all eight subtasks declare their parent task | **PASS** |
| Step 0.5 item 4 claim | `017-S` `queued → active`; `018.008-T → active`; `018-F → active`; **archived members untouched** (verified by execution: byte-identical file hashes before/after claim) | **PASS** |
| Step 0.5 item 6 intake pre-mode (`expected_status: queued`, pre-claim) | `018.008-T` `matched`; 11 legacy `pre-archived`; 0 orphans (no queue record declares `shipment_id: 017-S`); record `queued`, no task `active`/`done` → `record-consistent` | **PROCEED** |
| Step 2 harness generation | queued tasks of `018-F` = `[018.008-T]` | **1 function, 1 marker** |
| Step 3 item 1 executable-set derivation | `artifact_type: task` filter first (**never** the ID suffix — subtask IDs end `-ST`, whose last character is also `T`), then the positive status rule | KEEP `[018.008-T]`; `pre_archived_skipped` `[018.001-T, 018.002-T, 018.003-T]`; `already_done` `[]`; fail-closed halts **none**; subtasks excluded before any status read |
| Step 4.5 completion | `move 018.008-T --status done` → relocated to `.backlogit/archive/` with declared `status: done` (registry routes `done` to `archive`) | **verified by execution** |
| Step 6 pre-mode (`expected_status: done`) | `018.008-T` not in queue → found in archive → `pre-archived` (archive branch performs **no** status check); 11 legacy `pre-archived`; 0 orphans; record `active` → `record-consistent`; **`018-F` is not a member and is never evaluated** | **PROCEED** |
| Safe-close Step 0(c) classification | zero feature members ⇒ zero qualifying roots ⇒ P-015 precondition 4 fails for every member ("a task whose ancestry does not lead back to one of the qualifying root features"); skill default rule concurs | **`SAFE_CLOSE`** |
| Safe-close step 2 protected set | covering feature `018-F` derived from hierarchy, absent from manifest → partial-feature shipment; every other `018-F` descendant at every depth **is** a member | **protected set = `{018-F}` exactly** |
| Safe-close step 3 baseline integrity gate | `018-F` present in `.backlogit/queue/` | **PASS** (this is the gate the two-entry manifest failed) |
| Safe-close step 4 archive loop | all 12 members already in `archive/` → all `pre-archived` → skipped | **zero backlog mutations** |
| Safe-close steps 5 / 7 invariants | `{018-F}` still in `.backlogit/queue/`; verified by execution that neither claim nor single-artifact archive touches a non-member covering feature | **PASS** |
| Safe-close step 8 shipment-record close | `backlogit move 017-S --status shipped` → **exit 9, refused** | **HALT** `RECONCILE_FAIL_SHIPMENT_RECORD_LIVE_STATUS` — see §1.6 |

**Allowed archive IDs** for this closure (safe-close scope) = the 12 manifest item IDs + the
shipment record `017-S`. **Required transitions** = the shipment record only; all 12 members are
already archived and have no transition to report. **Post-close state** would be: `017-S` archived
with `archived_status: shipped`; all 12 members archived; **`018-F` still live and protected in
`.backlogit/queue/`** — exactly the goal.

### 1.5 Why no other manifest shape works

| Shape | Classification | First failing gate |
|---|---|---|
| `[018.008-T]` only | `SAFE_CLOSE` | safe-close step 3 — 11 archived descendants in the protected set, no pre-archived exemption |
| `[018-F, 018.008-T]` (original) | `SAFE_CLOSE` | same as above |
| `[018-F + all 12 descendants]` (rev 17) | `CASCADE` | **Step 6 pre-mode** — `018-F` live `active` vs `expected_status: done` |
| **`[018.008-T + all 11 legacy descendants]` (rev 18)** | **`SAFE_CLOSE`** | **safe-close step 8** — tool refuses the shipment-record move |

The only shape reaching the *permitted* cascade path must contain `018-F`, and that shape fails
pre-mode first. The two failure modes are mutually exclusive; **there is no third shape.** The rev-18
shape is the one that clears every gate that any manifest shape can clear.

### 1.6 Residual blocker — **P0**, pre-existing, shape-independent, out of scope here

Safe-close step 8 closes the shipment record with
`backlogit move <shipment_id> --status shipped`. **backlogit 1.10.1 refuses that command for
shipment artifacts:**

```text
move shipment 017-S to shipped via generic path:
backlogit: shipment must be shipped via ShipShipment, not a direct status update
exit code: 9
```

**Verified by execution in this cycle** from both `queued` and `active` record states. Already
recorded in this workspace by the `009-S` safe-close report — *"the generic safe-close move path
described in this skill's steps 1–10 is not available in `backlogit` 1.10.1 for shipment records"* —
and by `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`.

Both available substitutes are forbidden by the contract's own negative scenarios: archiving an
`active` record yields `archived_status: active` → `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`, and
invoking the cascade op after a `SAFE_CLOSE` classification is a **P-005 process deviation**
("substituting … between the classification and the cascade invocation is a P-005 process deviation,
never a permitted fallback").

*Observed but not adopted*: an isolated probe showed `backlogit shipment ship` on a **task-only**
manifest archives only the manifest's live members plus the shipment record
(`archived_ids: [<task>, <shipment>]`, `returned_ids: []`), leaves the covering feature untouched in
`queue/`, and stamps `archived_status: shipped` + `commit` — i.e. it produces exactly the safe-close
postcondition, non-destructively. This is recorded as **evidence for the follow-up decision only**.
It is **not** taken this cycle, because the binding classification is `SAFE_CLOSE` and this
remediation is forbidden from amending Ship/reconcile contracts.

**Disposition — BLOCKED, escalated, not guessed.** This is **workspace-wide** (every partial-feature
shipment), **pre-existing** (predates `017-S`), and **not remediable by any manifest shape**.
Resolving it requires either a backlogit change or a separately deliberated and reviewed amendment
to `shipment-reconcile` step 8 / P-015. Recorded as a stash entry for Stage intake. Until then:
**`017-S` executes, reviews and MERGES normally, and pauses at Ship Step 6 closure for operator
disposition.**

## 2. Interim pipeline behaviour — operator-only, stated plainly

`018.008-T` delivers the branch **gate** only. **Stage commit production is not in the current
task** (it is `019.009-T`). The honest terminal:

* **Stage branch gate.** Step 1.9 prevents any tracked Stage artifact mutation on the default
  branch; all bounded Stage mutations happen on `chore/stage-{scope-slug}`.
* **Stage terminus.** Stage ends at a **named fail-closed handback**,
  `STAGE_ARTIFACTS_UNCOMMITTED`, with bounded artifacts **uncommitted** on that branch under
  `.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/` — uncommitted **because commit and
  handback automation is deferred** (`019.009-T` / `019.007-T`).
* **The Orchestrator MUST HALT.** It **MUST NOT** execute Step 1.5 item 3 (the uncommitted/unpushed
  arm) for this interim route.
* **Operator-only completion.** Only the operator may (1) verify branch, worktree topology and the
  uncommitted artifact set; (2) commit **only** the four allowed Stage roots; (3) push the side
  branch; (4) open and approve a **merge-commit** staging PR (P-009); (5) update the default branch;
  (6) re-verify the manifest on `origin/main`.
* **The current dark run pauses there.** Because the operator is AFK, automation **may not proceed
  past that checkpoint** until `019-F` ships. **The dark pipeline therefore cannot finish
  autonomously after this release unit is implemented.**

### 2.1 Two rev-17 authority claims are WITHDRAWN as factually wrong

1. *"Orchestrator Step 1.5 is naturally fail-closed for this route."* **False.** Step 1.5's step 1
   branches on `git status --short -- .backlogit/` being **dirty**, and a Stage handback leaves it
   dirty — so control reaches **item 3**, the commit/push arm, not the step-4 verification halt.
2. *"The Orchestrator acts under its OWN installed Step 1.5 item 3 authority."* **False as a grant.**
   Item 3(a) commits Stage artifacts and item 3(e) instructs *"Attempt a direct push to `main`
   first."* Both contradict the Orchestrator's own installed Role statement — *"You do NOT triage
   stash entries yourself. You do NOT write code or create PRs yourself."* — and a direct-`main`
   attempt is precisely the behaviour this release unit exists to prohibit. Step 1.5 item 3's
   dirty-artifact arm is an **internal contradiction in the installed Orchestrator contract**, not a
   usable authority.

**No Orchestrator authority is invented, asserted, or relied upon.** Repairing that contradiction is
`019.004-T`'s job, deferred and unreviewed. Until then the route is operator-only.

**Withdrawn overclaims**: "identical to present behavior"; "no skipped assertions anywhere in the
suite"; automatic Orchestrator branch discovery; automatic PR persistence; no-shipment semantics;
generalized CI enforcement; fixture platform; regeneration proof; and the two authority claims
above.

## 3. Deferred-work split — Strand A / Strand B

The former `019-F` mixed **persistence semantics** with **reusable enforcement infrastructure**.
Split into two cohesive features. Rev 17 gave each a shipment; **rev 18 removed both shipments**
(§4) — the strands themselves are unchanged and remain cohesive.

| Strand | Feature | Shipment | Tasks |
|---|---|---|---|
| A — Stage handback, staging-PR persistence, no-shipment semantics | `019-F` (blocked) | **none** (`018-S` archived) | `019.004-T`, `019.007-T`, `019.008-T`, `019.009-T` |
| B — detector, fixture corpus, event parser, CI, regeneration proof | `020-F` (blocked) | **none** (`019-S` archived) | `020.001-T` … `020.005-T` |

Moved with backlogit's supported `adopt`, which re-IDs and re-parents the task. Nothing
destructively removed.

| Former ID | New ID | Former ID | New ID |
|---|---|---|---|
| `019.001-T` | `020.001-T` | `019.005-T` | `020.004-T` |
| `019.002-T` | `020.002-T` | `019.006-T` | `020.005-T` |
| `019.003-T` | `020.003-T` | | |

### 3.1 Adoption provenance — corrected in rev 18

The structured frontmatter field **`origin_feature` on all five Strand B tasks reads `018-F`, and
that is correct**: `origin_feature` records the **original/root** origin feature and is **preserved
across successive adoptions**, never rewritten to the immediate predecessor. The **intermediate**
re-parent history — `018-F` → `019-F` → `020-F` — is carried in **prose only**, in each task body
and in §8 of the Strand B plan, because no structured field records it. The same applies to
`019.004-T` and `019.007-T` (`origin_feature: 018-F`, intermediate step `018-F` → `019-F`).

Two rev-17 claims are **withdrawn as inaccurate**: (1) that `adopt` "records
`origin_feature: 019.001-T -> 020.001-T`" — that conflated an **ID remap** with a **field value**;
and (2) the rev-17 provenance table asserting `origin_feature` = `019-F`, which **contradicted the
stored frontmatter**.

### 3.2 Dependencies rebuilt

Removed as semantically wrong after the split: `019.004-T → 020.003-T` (Orchestrator discovery
does not need the detector), `020.004-T → 019.004-T` and `020.004-T → 019.007-T` (CI wiring does
not need the persistence tasks). Added: `019.004-T → 018.008-T`, `019.004-T → 019.007-T`,
`020.004-T → 020.003-T`. **Rev 18 removed the two readiness-lock edges** `018-S → 020-S` and
`019-S → 020-S` before archiving those records, so **no dependency edge now references an archived
shipment**. The strands are independent; each depends only on `018.008-T` and its own chain.
Verified: **no cycle** in the full graph, **no `current → future` edge**.

The rev-17 note about `018-S`'s dangling manifest is now **moot**: `018-S` is archived, and the
narrow `custom_fields.items` fallback it described was **not** needed in rev 18 — `017-S`'s manifest
change used the registry-declared `return_blocked` operation instead (§1.3).

## 4. Readiness representation — the rev-17 lock was invalid and has been retired

### 4.1 `blocked` is not a valid shipment status

Rev 17 made `018-S` and `019-S` `status: blocked` and gave each a `blocks` edge to a **lock-token
shipment** `020-S`. That entire construction rests on a value the schema does not define.

Backlogit's shipment status enum is exactly `queued | active | shipped | abandoned`
(`.backlogit/header-def.yaml`; `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`).
The task/feature enum — which *does* include `blocked` — is a different enum.

**Verified by execution** against backlogit 1.10.1 in an isolated probe workspace:

| Command | Exit | Meaning |
|---|---|---|
| `backlogit move <shipment> --status blocked` | **0** | accepted — the generic mover validates against the *generic* status enum in `config.yaml`; this is the known CLI defect that wrote those values |
| `backlogit move <shipment> --status shipped` | **9** | refused: `shipment must be shipped via ShipShipment` |
| `backlogit move <shipment> --status abandoned` (from `queued`) | **1** | refused by `validate_status_transition` |
| `backlogit move <shipment> --status queued` (from `blocked`) | **0** | accepted — so `blocked` is not even a one-way state |

A lock token that can exist **only** in an invalid state is not enforcement. `backlogit doctor`
reports "No issues found" for such records, so nothing would have caught it.

### 4.2 There is no safe valid queued-shipment readiness lock

The obvious repair — leave the deferred shipments `queued` and rely on a `blocks` edge to a
predecessor — **does not work** in current backlogit semantics unless a genuine predecessor shipment
is actually **intended to ship**:

* a `queued` shipment is by definition **claimable**;
* `blocked → queued` as an "unlock transition" is not a supported shipment lifecycle edge;
* the only legitimate suppressor is an *unshipped predecessor shipment*, and inventing a predecessor
  that will never ship reproduces the lock-token defect in a different colour.

**Therefore: do not leave implementation-unready future shipments claimable.**

**This supersedes the earlier request for a queued future shipment.** The adversarial gate proved no
valid safe queued representation exists; **reliability and safety take precedence** over having a
shipment record present.

### 4.3 What was actually done

1. `backlogit dep remove 018-S 020-S` — removed.
2. `backlogit dep remove 019-S 020-S` — removed.
3. `backlogit archive 018-S` · `backlogit archive 019-S` · `backlogit archive 020-S` — the
   **official non-destructive** archive operation.

**Nothing was deleted.** Each archived record retains its ID, title and full `custom_fields.items`
manifest, gains `archived_from`, and carries `archived_status: blocked` — an honest provenance
record of the defect-written value, and a value the `shipment-reconcile` contract already enumerates
and handles (`archived_status: active|queued|blocked|abandoned` in the sequence-aware exclusion
rule).

**Why archive directly rather than normalise to `queued` first.** Normalising would have produced a
cleaner `archived_status: queued`, but at the cost of a window in which three
implementation-unready shipments were `queued` and therefore **claimable**. Zero claimable exposure
was judged more important than cosmetic provenance, and the rejected alternative is recorded here.

**Verified by execution** that archival costs no future capability: `backlogit archive <shipment>`
does **not** cascade to its members (queue files untouched), and a member of an **archived** shipment
can be freely added to a **new** shipment.

### 4.4 What replaces the lock

Deferred features `019-F`, `020-F` and the readiness strand `021-F`, and **all** their tasks, are
`status: blocked` and belong to **no shipment**.

Orchestrator Step 2 and Ship Step 0.5 both operate on **shipments**. With no shipment there is
nothing to list, nothing to select, nothing to claim, and nothing to mis-route. This is strictly
stronger than any status-based lock, because it removes the eligibility *surface* rather than trying
to mark it ineligible.

**It is acceptable — and safer — to retain blocked future features and tasks without any live
shipment until implementation readiness.**

Verified:

```text
backlogit queue view --group-by status
017-S      Stage artifact branch/PR policy gap correction        queued   shipment
018-F      Stage artifact branch/PR policy gap correction        queued   feature
018.008-T  Gate Stage artifact mutation behind a dedicated …     queued   task
019-F      Stage artifact persistence route execution            blocked  feature
020-F      Mechanical enforcement of the Stage default-branch …  blocked  feature
021-F      Harness execution-model readiness decision            blocked  feature
021.001-T  Decide and review the deferred harness execution …    blocked  task
```

`017-S` is the **only** shipment in the workspace, and no invalid shipment status remains in the
queue.

### 4.5 Reconstitution (Stage only, never Ship, never the Orchestrator)

After `021.001-T` records a **reviewed** decision, Stage: (1) restructures the target feature into
the shape that decision requires; (2) runs `impl-plan`, `plan-harden` (**required**) and
`plan-review`; (3) **only then** creates a **fresh shipment** over the reviewed shape and moves its
member tasks to `queued`. **Creating a shipment before plan review is a P-006 violation.** There is
no lock edge to remove and no shipment to un-block — that is the point.

### 4.6 `021-F` / `021.001-T` are retained, not archived

The remediation directive permitted archiving `021-F`/`021.001-T` *"if they exist only to support
that invalid lock."* They do not. `021.001-T` carries the substantive three-option harness
execution-model analysis that the reconstitution path in §4.5 and the P-006 plan review both depend
on. Only the **lock-token machinery** was removed from them; the decision content is retained,
`blocked`, and shipment-free.

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
| `018.001-T`–`018.003-T` (+ 8 subtasks) | Remain **archived**. All 11 are `017-S` manifest members **for closure only** — their membership is what empties the protected set (§1.4). `pre_archived_skipped` (tasks) or excluded at the `artifact_type` filter (subtasks). Not unarchived, not executed. |
| `018-F` | **Removed from the `017-S` manifest** in rev 18 via `shipment return-blocked`, then restored to `queued`. Remains the covering feature, resolved through `parent_id`, and is the sole protected-set member at closure. |
| `018.008-T` | **Retained**, body rewritten again in rev 18: exact-anchor Step 1.9 registration, P-016 worktree precheck, exact slug source fields + 48-byte bound + empty fallback, branch ownership/identity discriminator, mandatory P-004 marker with explicit H0 evidence, `artifact_type`-based executable filtering. |
| `019.004-T` | Body rewritten (Strand A); rev 18 replaced the readiness-lock notice with the shipment-free notice and corrected the `origin_feature` provenance wording. |
| `019.007-T`, `019.008-T`, `019.009-T` | Bodies retained; rev 18 replaced the readiness-lock notice with the shipment-free notice; all three successors of the retired `018.009-T` named in full in each. |
| `019.001-T`, `019.002-T`, `019.003-T`, `019.005-T`, `019.006-T` | **Adopted** to `020-F` as `020.001-T`…`020.005-T`; bodies fully rewritten in rev 17, provenance wording corrected in rev 18. Historical machinery is explicitly non-governing provenance, recoverable at `git show 14d64e4:.backlogit/queue/019.00N-T.md`. |
| `018-S`, `019-S`, `020-S` | **Archived** in rev 18 with `backlogit archive`, after their `blocks` edges were removed. Manifests, IDs and titles preserved; `archived_status: blocked` retained as provenance. Not deleted. |

## 7. Residual risk, accepted

| Risk | Disposition |
|---|---|
| **`SAFE_CLOSE` shipment-record close is tool-blocked (backlogit exit 9)** | **P0, pre-existing, workspace-wide, shape-independent.** See §1.6. `017-S` executes, reviews and merges normally, then **pauses at Ship Step 6 closure for operator disposition**. Escalated as a stash entry; requires a backlogit change or a separately reviewed `shipment-reconcile`/P-015 amendment, both out of scope for this cycle. |
| No mechanical CI enforcement until Strand B lands | Accepted; contract regeneration-vulnerable in the interim; tracked by `020-F`. |
| No automated persistence route until Strand A lands | Accepted, **operator-only**. Stage halts at `STAGE_ARTIFACTS_UNCOMMITTED`; the Orchestrator must halt and must not run Step 1.5 item 3; only the operator may commit, push, open/approve the merge-commit staging PR and re-verify. **The dark run cannot finish autonomously past that checkpoint while the operator is AFK.** Tracked by `019-F`. |
| Installed Orchestrator Step 1.5 item 3 contains a direct-`main` attempt that contradicts its own Role statement | Accepted and **recorded**, not exercised. Repair is `019.004-T`, deferred and unreviewed. Until then the Orchestrator MUST HALT rather than execute that arm. |
| Deferred work unschedulable until the execution model is decided | Accepted. `021.001-T` is the decision; deferred features are `blocked` with **no shipment**, so nothing is claimable. |
| Archived shipment candidates could be mistaken for live work | Mitigated: they are out of `.backlogit/queue/`, absent from `backlogit shipment list`, and carry no inbound dependency edge. Verified by execution that their members remain re-assemblable into a fresh shipment. |
| `backlogit adopt` does not rewrite shipment manifests | Known gap; **moot for `017-S`** (rev 18 used the registry-declared `return_blocked` operation instead of a frontmatter fallback). Re-recorded here for the tool owner. |
| Generic `backlogit move` accepts invalid shipment statuses (exit 0 for `--status blocked`) | **Tool defect, recorded.** It is what produced the rev-17 lock. Mitigated here by removing every such record from the queue. Reported for the tool owner. |
| Registry declares a `complexity` param the workspace schema lacks | Reporting defect only; complexity carried as enum-validated prose. Fixing it would mutate `.backlogit/header-def.yaml`, outside Stage's role boundary. |

## 8. Open questions for review

* **OQ-1** — Which of the three `021.001-T` options should the deferred strands adopt? Option 1 is
  the proven default; option 3 is the only one that removes the constraint permanently.
* **OQ-2** — **How should the safe-close step-8 closure conflict (§1.6) be resolved?** The two
  candidate resolutions are: (a) a backlogit change exposing a non-cascading shipment-record close;
  or (b) a reviewed `shipment-reconcile`/P-015 amendment recognising that a **task-only** manifest
  has no explicit feature member for the engine to force to `done`, so the cascade op is
  non-destructive for that shape. Evidence for (b) is recorded in §1.6 but **not acted on**.
  This is the highest-priority open item.
* **OQ-3** — Should `019.004-T` also repair the internal contradiction in the installed
  Orchestrator Step 1.5 item 3 (the direct-`main` attempt in 3(e) versus the Orchestrator's
  orchestration-only Role statement), or should that be a separate P-010 amendment?
* **OQ-4** — Should the generic-`move`-accepts-invalid-shipment-status defect (§4.1) be reported
  upstream to the backlogit owner as a schema-validation bug?

## 9. Harness and Step 1.9 corrections (rev 18)

### 9.1 P-004 compliance is unconditional

Rev 17 framed the expected-failure marker as a **fallback** — *"if and only if harness-architect
insists on a callable panic marker."* That was **P-004-noncompliant and is WITHDRAWN**.

P-004's precondition is `go test ./...` exiting non-zero **with expected failure markers in the
output for every test function**, and harness-architect Step 5.2 requires **all** harness tests to
fail with `panic("not implemented: <reason>")`. Neither is contingent on anything.

The corrected contract: **one** generated test function
(`TestStageBranchGate_ContractCorrected`), **one** expected marker — an unexported helper **inside**
`tests/integration/stage_branch_gate_test.go` with body
`panic("not implemented: stage branch gate contract")`, invoked **unconditionally** — **all red
before implementation**, then **full-suite green after the sole task**. No new file, no new package,
nothing under `internal/` or `cmd/`; still 3 files, 0 new production files.

**Explicit H0 evidence** (four captured checks): `go vet ./...` exit 0; the harness run exits
non-zero; the output carries the literal marker attributed to the test function; the H0 red count
over this unit's generated functions is **literally 1 of 1**.

### 9.2 Step 1.9 algorithm tightened

| Area | Rev-17 state | Rev-18 requirement |
|---|---|---|
| Checklist registration | "between the Step 1.8 and Step 2 lines" | **Byte-exact anchors** quoted with their EM DASH U+2014 and the three ASCII spaces after `Step 2`; a **normalized locator** fallback; **halt on zero or multiple matches** (`STAGE_BRANCH_GATE_ANCHOR_FAIL`); never an ordinal insertion |
| P-016 topology | absent | **`git worktree list --porcelain` first**, classify **every** attached worktree (current / explicit Stage spike-research / prohibited-ambiguous) **before** any create/select/resume; spike/research class **passes** and the P-016 exception is **not narrowed**; halt on prohibited or unclassifiable |
| Slug source | "the Step 1 classification subject" / "the covering-feature title" | **Exact source fields** per shape, including the shipment-shaped optional source |
| Normalization | "lowercase, collapsed, trimmed, length-bounded" | **Five ordered steps**, ASCII-only case folding, **explicit 48-byte maximum** justified against `max_slug_length: 60`, post-truncation hyphen trim |
| Empty slug | undefined | **Literal `stage-session` fallback**; never halt, never improvise |
| Base ref | "the configured default branch" | **Resolved from `origin/HEAD`** with a documented fallback and a **halt if both fail** — never assume `main` |
| Branch collision | "a resumption case, not a failure" | **Ownership / base-identity discriminator**: non-empty merge-base with `origin/{base_ref}` **and** a two-dot-three diff touching only the four Stage artifact roots. **Resume only on identity match; otherwise HALT.** |
| Retained unchanged | — | symbolic-HEAD equality (detached fails), expected ≠ default, categorical mutation deferral with illustrative examples, named `STAGE_BRANCH_GATE_FAIL` halt with P-005 record and no tracked mutation |

### 9.3 Executable filtering is by `artifact_type`, never ID suffix

Corrected wherever it appeared. The authoritative predicate in Ship Step 3 item 1 and in
`shipment-reconcile`'s record-status classification is the record's frontmatter
`artifact_type == "task"`. An ID-suffix heuristic is **wrong**: subtask IDs end `-ST`, whose final
character is also `T`, so a suffix test admits every subtask. This is now load-bearing, because
`017-S`'s manifest deliberately contains eight subtasks (§1.3).

