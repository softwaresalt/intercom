---
title: "Plan — Ship covering-feature completion and close-path delegation (Foundation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 1
feature: 022-F
task: 022.001-T
shipment: 021-S
deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Plan — Ship covering-feature completion and close-path delegation (Foundation)

> **Requires plan hardening**: **yes — applied in rev 1** (see `## Plan Hardening`).

- **Shipment**: `021-S` (A, Foundation) — first of three in the bootstrap chain
- **Closes by**: the **existing** P-015 `VERIFIED FULLY-COVERED-ROOT EXCEPTION` (`CASCADE`)
- **Authorizes**: nothing new in the close-path verdict set

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

## 2. Why this is first

Two independent reasons, both verified at `e7e981d`:

**2.1 Pre-mode makes the shape unreachable today.** `_ship.agent.md` Step 6.1(a) invokes
`shipment-reconcile` with `expected_status: done` and verifies *"every manifest item is
present in queue with `status: done`"*. A manifest `[022-F, 022.001-T]` therefore needs
the **feature** `done`. Ship's Role Boundary permits only *"move **tasks** to
active/done"* (line 38). No feature-completion authority exists, so no fully-covered-root
manifest can ever reach closure. B's manifest has the same shape, so **B cannot close
until A is live**.

**2.2 The duplicate has already drifted unsafe.** This is a pre-existing defect found
during re-planning, independent of the bootstrap:

| Surface | Coverage rule |
|---|---|
| P-015 item 1 (v1.24.0), `workflow-policies.md:436` | *"every one of its **descendants — at every depth, not only direct children**"*, and: *"A check limited to direct children is insufficient"* |
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
| **A-1** | Role Boundary table row (line 38) — narrow feature-completion authority |
| **A-2** | New **Step 6.1(a1)** — covering-feature completion gate |

### 4.1 Delegation wording constraint (normative)

The replacement text MUST NOT enumerate the authorized verdicts. It delegates:

> Ship MUST invoke the close path named by the machine-checkable classification defined
> in **P-015** and implemented by the **`shipment-reconcile`** skill, and MUST NOT invoke
> any close path the classification did not name. Ship does not re-derive, restate,
> extend or narrow that classification.

This is what lets B and C land **without touching this file again**. It also forbids Ship
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

### 4.3 Stage precondition — manifest re-shape (owned, not assumed)

`021-S`'s live manifest is **task-only** `[022.001-T]`. The fully-covered-root shape this
plan depends on is a **required re-shape to `[022-F, 022.001-T]`**, performed by **Stage**
(Ship's Role Boundary forbids editing shipment planning fields). It is done **before**
Ship claims `021-S`, and it is completed in the re-plan session that produced this plan.

If it were skipped, A's manifest would stay task-only, the classifier would return
`SAFE_CLOSE` (no feature member ⇒ no `CASCADE`; `TASK_ONLY_FINALIZE` not yet authorized),
and A could not close. That outcome is **fail-closed** (§10.3), but it is a precondition
with a named owner, not a hope. Asserted by **AC-12**.

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
2. **H1 — green.** Apply A-1, A-2, C1–C6. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function, so Ship Step 2's feature-scoped batch yields one
function and Step 4.3 cannot deadlock (the `021.001-T` constraint).

## 8. Verification obligations

### 8.1 What the Go test proves

**Contract text and ordering on `_ship.agent.md` only.** 10 rows:

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

Rows 7–10 are the **widening guards**: they fail if this change authorized anything new.
Row 11 is the **coherence guard**: it fails if C1 ships with the carve-out and the
"close under the just-merged contract" directive unreconciled.

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
| 0 | **Stage** | **Pre-execution readiness check (PO-1)** — read `022-F`'s live status *shape* under claim semantics **before** implementation begins (§10.1) |
| 1 | S1 | Verify on `main`; claim `021-S`. Ship moves **`022.001-T -> active`**. Whatever status claim leaves `022-F` in is **observed, not assumed** (PO-1) |
| 2 | S1 | H0 red → implement → H1 green; commit; push; PR; operator approval; merge |
| 3 | S1 | Verify merge SHA; `022.001-T -> done` |
| 4 | S1 | **Reload merged `main`.** Detect that the merged change alters **Ship's own Role Boundary**. Per §6 the new authority is **not** available to this session |
| 5 | S1 | **Write checkpoint** (`phase: awaiting-fresh-session-parent-completion`, `resume_hint` naming `021-S`, `022-F`, the merge SHA) and **END** |
| 6 | **S2** | **Fresh session.** Loads the new Role Boundary + Step 6.1(a1) from merged `main` |
| 7 | S2 | Restore checkpoint; validate the **same** active shipment `021-S`; verify merge SHA is an ancestor of `main` |
| 8 | S2 | Step 6.1(a0) topology gate → **Step 6.1(a1)**: verify §5 conditions 1–5; `022-F active -> done` |
| 9 | S2 | Step 6.1(a) pre-mode, `expected_status: done` — **both** members now `done` → `PROCEED` |
| 10 | S2 | Classification returns **`CASCADE`** (fully-covered root); close via the cascade op; P-007; post-mode; sync; verify |
| 11 | S2 | Operational closure + P-020 on the post-merge branch; closure PR; operator merge |
| 12 | — | Orchestrator routes **B** only |

### 9.1 What this proves

* **Checkpoint/resume across a session boundary** — steps 5→7, executed, not asserted.
* **No authority gained mid-session** — S1 never uses feature-completion authority;
  it is used first by S2, which did not perform the merge. This is the §6 property
  demonstrated by execution.
* **A closes by an existing path.** At step 10, P-015 still has exactly one exception.
  A introduces no verdict and closes by the pre-existing one.

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | `022-F` may be `queued`, not `active`, at the a1 gate, so §5 condition 5 fails closed and A cannot self-close | **PO-1a** (pre-execution) + **PO-1b** (at the gate), §10.1. Fail-closed HALT is correct; the plan does **not** assume claim sets features `active` |
| **R-2** | S1 might use the new authority anyway, defeating the proof | §6 is stated normatively **and** asserted by Go test row 6 |
| **R-3** | Delegation could be read as authorizing any verdict the skill invents | §4.1 binds Ship to verdicts the **classification** names; P-015 remains the authorization surface. Go test row 9 guards it |
| **R-4** | C2's correction touches the safe-close summary, which C also edits in its own file | **No overlap**: C2 is in `_ship.agent.md`; the skill's step 8 is C's file. Disjoint surfaces |
| **R-5** | Removing Ship's restatement loses reviewer-visible context | Accepted. The classification stays fully specified in P-015 and the skill; duplication is what caused the §2.2 drift |

### 10.1 PO-1 — covering-feature status at claim (**pre-execution** readiness check)

**Promoted to a pre-execution check (review correction).** PO-1 was drafted as a
closure-time read, which would have discovered a claim-shape mismatch only *after* A's
implementation merged — the most expensive possible moment. It now runs **twice**:

**PO-1a — before implementation begins (Stage / §9 step 0).** Determine what status
shipment claim leaves a covering feature in. Evidence may be taken from an already-closed
fully-covered-root shipment in `.backlogit/archive/`, or from a scratch probe. If claim
does **not** leave the covering feature `active`, **stop before implementing** and return
to Stage to choose between (i) amending the claim step to transition covering features to
`active`, or (ii) widening the authorized transition — either as its **own** reviewed
unit.

**PO-1b — at the gate (S2, before §9 step 8).** Re-read `.backlogit/queue/022-F.md` and
record the live `status`.

* `active` → proceed.
* anything else → **HALT**. Do not widen §5 condition 5 in-flight. Return to **Stage**.

Widening an authority envelope mid-execution is forbidden. PO-1a makes the expensive
discovery cheap; PO-1b keeps the gate fail-closed regardless.

### 10.1.1 Probe 18 travels with A

C2's correction rests on the claim that the engine **refuses** `backlogit move
<shipment_id> --status shipped` for shipment artifacts (exit 9). That is **Probe 18**, and
it is A's justification, not C's. A cites
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe18-liveconfig-inbound-*`
and the compound note
`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`. Asserted by **AC-13**.

### 10.2 Rollback

Two files on a feature branch. `git revert` the merge commit; Ship's closure behavior
returns to today's state (including, knowingly, the §2.2 drift). No backlog mutation is
performed by the implementation commit itself, so no backlog rollback is required.

### 10.3 Failure paths

| Failure | Response |
|---|---|
| H1 cannot be reached in budget | HALT, return to Stage. Never split — A is already minimal |
| a1 gate conditions 1–4 fail | HALT fail-closed; surface which condition; no close |
| Pre-mode returns `RECONCILE_FAIL` | HALT; do not proceed to close; surface the report |
| Classification returns `SAFE_CLOSE` instead of `CASCADE` | HALT — indicates the manifest is not the fully-covered root this plan asserts. Return to Stage |
| Cascade archives anything outside `allowed_ids` | P-015 violation action: `git restore`, re-verify protected set, HALT |

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
  checkpoint MUST be written through `backlogit_create_checkpoint` (tool-timestamped) and
  its `created_at` must precede S2's resume. *Residual, stated honestly:* the artifact is
  not tamper-proof — the two-session property ultimately rests on process integrity, and
  the checkpoint is evidence, not enforcement.
- **AC-10** PO-1a executed **before** implementation; PO-1b executed before the a1 gate; both recorded.
- **AC-11** Exactly 2 files changed. 0 new production files.
- **AC-12** `021-S`'s manifest is `[022-F, 022.001-T]` **before** claim (§4.3, Stage-owned).
- **AC-13** Probe 18 is cited as C2's justification (§10.1.1).
- **AC-14** C1 ships with the §4.2 pinned reconciled wording; Go test row 11 passes.

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity is carried as **enum-validated prose**: this
workspace's `header-def.yaml` defines `size` on the task type but **no** `complexity`
field (confirmed at `.backlogit/header-def.yaml:144–166`, and in
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`). Size is structured.

| Work item | Estimate |
|---|---|
| Role Boundary grant (A-1) | ~8 min |
| Step 6.1(a1) parent-completion step (A-2) | ~10 min |
| C1 reload extension | ~4 min |
| C2 safe-close summary correction | ~4 min |
| C3 + C4 + C5 generic delegation | ~12 min |
| C6 delegation bullet | ~4 min |
| Go table-driven test (11 rows) + ≤4 helpers | ~19 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~71 min ≈ 1.2 h** |

Rates are rev 6 §10.1's own, so this is directly comparable to the measurement that
produced `MUST_REPLAN`. Margin to the 2-hour rule: **~49 min**.

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

**Hardening 2 — delegation must not become a blank cheque.** §4.1's wording binds Ship to
the verdict **named by the classification**, and P-015 remains the sole authorization
surface. Go test row 9 fails if `_ship.agent.md` ever re-enumerates verdicts. Ship cannot
authorize `TASK_ONLY_FINALIZE`; only B can, and only under C's token.

**Hardening 3 — the self-hosting boundary is the highest-risk step.** The S1→S2 handoff is
where an impatient implementer would "just finish it". Three independent defenses: §6 is
normative contract text; AC-9 explicitly fails if one session does both; the Go test
asserts the rule is present in the merged file. The checkpoint `resume_hint` names
`021-S`, `022-F` and the merge SHA so S2 can validate rather than trust.

**Hardening 4 — R-1 is a real possibility, planned for, and now discovered early.** The
plan does **not** assume claim leaves a covering feature `active`. **PO-1a** makes it an
explicit pre-execution read, so a claim-shape mismatch surfaces **before** any
implementation is written rather than after the merge — the review correction that turned
the most expensive discovery into the cheapest one. **PO-1b** keeps the gate fail-closed
regardless. Both non-`active` branches halt and return to Stage rather than widening the
grant in-flight, deliberately preferring a planned halt over a silent authority expansion.

**Hardening 5 — engine coupling.** A performs **no** engine call that depends on the
1.10.1 digest beyond what every closure already does. The digest binding
(`1E106F5F…959A98`) is **C's** obligation. A's cascade invocation is the same call every
fully-covered-root closure already makes today.

**Hardening 6 — what could make this plan wrong.** If `expected_status: done` is not
actually enforced over feature members, the a1 gate is unnecessary (harmless but
pointless). If claim already sets features `active`, PO-1 passes trivially. Neither
falsifies the delegation fix, which stands on §2.2 alone. The plan is therefore
**falsifiable in both directions** and its two deliveries are **independently
justified**.
