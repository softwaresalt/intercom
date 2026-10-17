---
title: "026-F re-stage decision: O-1 closed (Option B), O-2 adjudicated, O-3 whole-surface matrix, O-4 re-scope"
date: 2026-10-17
status: accepted
type: deliberation-outcome
scope: 026-F (Ship harness selection conformance repair, DB12DA37)
operator_authority: "O-1 closed by the operator, Option B (recorded via the Orchestrator, 2026-10-17); one-task rule superseded by operator direction"
supersedes_in_part: docs/plans/2026-09-18-intercom-go-harness-selection-conformance-plan.md (single-unit shape, plan-review attempt 1 FAIL)
---

# 026-F re-stage decision (2026-10-17)

## 1. O-1 - CLOSED by operator authority (Option B)

The operator has closed O-1 with **Option B**, recorded authority dated 2026-10-17:
**a structural/substring contract test IS an admissible acceptance surface for
the 026-F fix** (DB12DA37 harness-selection conformance). No contract-linter
engine has to be built first. This resolves plan-review attempt 1 findings
**C3** and **C4**, which both rested on "no admissible acceptance surface".

Consequences recorded here:

* The adopted 022.001-T / IV-5 preference for behavioural fixtures does not
  apply to this fix: Ship Steps 2-3 are agent-interpreted prose with no engine
  to drive, and the operator has determined that substring/structural tests are
  sufficient. (Existing precedent for the test style:
  `tests/integration/ship_feature_completion_contract_test.go`, which states its
  own needles are supplemental evidence; here, by operator determination, they
  are the acceptance surface.)
* **The red phase** for each task is a test that FAILS because the required
  selection wording/predicate - **queued-status AND `harness-ready`, drawn from
  the shipment manifest** - is ABSENT from the Ship / harness-architect surface
  it covers, and PASSES after the edit. Needles are cut from the post-edit
  pinned wording; every assertion resolves inside its own subject.
* "Unchanged" assertions (old T3/T4) are green-on-arrival and are NOT test
  functions. Their protected invariants (P-002 and P-004 text, Step 4.3 text,
  direct-invocation behaviour) are proved by diff review at closure (section 6),
  never by a test that is green before the edit.

## 2. One-task rule SUPERSEDED (decision note)

The earlier decision (plan section U1 "single atomic unit", "cardinality N = 1
is mandatory", decision D4 "cannot be split without reproducing the G1
deadlock") rested on the one-task-per-feature/shipment rule. By operator
direction there is no such rule: the only constraint is that size and
complexity are reasonable and managed. The rule is **not applied**. The G1
deadlock (Step 4.3 full-suite green after every task vs several simultaneously
red harness functions) is avoided on its own merits by per-task just-in-time
harness authoring: each task's test file additions are written only when that
task starts, tasks are strictly sequential, and every task ends full-suite
green. The unit is split into several tasks because the surface is wide, not to
dodge the deadlock.

## 3. O-2 - P-002 Statement clause adjudicated

P-002 **Statement** (`workflow-policies.md:44`): "The ship agent may only
**claim and implement** a task after the harness-architect has confirmed that
the test harness compiles and all tests fail in the red phase."

**Question.** Does the backlogit shipment claim (Step 0.5 item 4), which flips
manifest tasks `queued -> active` before any harness exists, violate the
Statement?

**Adjudication (conformance reading, no amendment).** No, for four reasons
grounded in the installed text:

1. P-002's own **Gate Point** is "Queue building (Step 2) and task claiming
   (Step 3)". Both gate points are Ship steps that run AFTER Step 0.5's
   shipment claim by construction. The policy therefore places its gate
   downstream of the shipment claim; it cannot have meant the shipment claim.
2. The act the policy names is Ship's own **Step 4.1 "Claim Task"** ("Update
   task status to `active`", `_ship.agent.md:377-379`), which runs after Steps 2
   and 3 and immediately before implementation. On the shipment route that
   status flip has already happened as a backlogit shipment-claim side effect;
   the Ship act that P-002 gates (selecting a task for the ready queue and
   implementing it) still occurs only at Steps 3-4.
3. P-002's **Precondition** ("the task carries the `harness-ready` label") and
   **Enforcement** ("Filter ready queue to only tasks carrying the
   `harness-ready` label") are properties of the READY QUEUE, not of the
   backlog status field. Conformance therefore requires that no task reach the
   Step 4 ready queue without the label, which is exactly what this feature
   restores.
4. The Violation Action ("Halt and suggest running the harness-architect")
   confirms the intended failure mode: Ship halts instead of implementing an
   unharnessed task. It does not require the backlog status to be `queued`.

**Limits of the adjudication.** It is an interpretation of installed text by
Stage, recorded for plan-review. It does not claim P-002/P-004 are amended, does
not authorize any bypass, and does not weaken the requirement that the label
precede implementation. It also removes the basis for the 020-F plan-review
constitution P1 only to the extent plan-review accepts it; 020-F stays held
(blocked) until 026-F ships regardless.

## 4. O-3 - surface matrix rebuilt whole (all clause sites)

The matrix of record is in the plan (section "Surface matrix", rows M1-M17). It
inventories every clause site across four files plus the unchanged protected
sites, per the compound stop rule, instead of the three-site list that failed
review. Sites added over the failed plan: `_ship.agent.md` Step 2 preamble
(line 318), partition (324-325), Step 3 preamble (337), empty-queue (369) and
broadcast (371/968), Step 4.1 (377-379), harness_cmd (431); and
`harness-architect/SKILL.md` lines 12, 38, 50, 54-56, 60 (and the new
branch-locus note). `build-feature/SKILL.md:12` and the policy rows are
verified consistent and unchanged.

## 5. C1-C8 dispositions

| ID | Disposition |
|---|---|
| C1 temporal incoherence (Step 2 cites Step 3 item 1) | Resolved without restructuring: Step 3 item 1 is a read-only pure function of manifest + statuses. Step 2 states an explicit evaluation order (evaluate Step 3 item 1's derivation first, read-only, once) and Step 3 item 1 states it is that same single derivation consumed by both steps. One authoritative copy; no second status rule. |
| C2 two execution-boundary authorities | Item 1's "actual task-membership boundary that Step 4 executes" is narrowed to a MEMBERSHIP boundary; the single execution gate is the per-task P-002 check at Step 4.1 (membership list in Step 3, label check per task after its just-in-time harness). One authority. |
| C3 / C4 acceptance surface | Closed by O-1 Option B (section 1). |
| C5 incomplete inventory | Closed by the whole matrix (section 4). A closure grep inventory is recorded in the PR readiness block. |
| C6 cardinality proof | The N=1 premise is dropped (section 2). G1 safety is rebuilt: on the shipment route harness generation is just-in-time per task (one task's red harness at a time, in dependency order, immediately before its Step 4.1), so Step 4.3 full-suite green holds at every task; no cardinality claim is made. The repair therefore also prevents the G1 deadlock that a naive whole-batch Step 2 would create for every multi-task shipment. |
| C7 route discriminator | Defined in the plan: the shipment route is any run where a `shipment_id` is recorded as session scope (Step 0.5 primary item 5 OR fallback item 4); both claim and auto-activate. The only non-shipment route is a run with no shipment support, which keeps the feature-scoped `queued`-only rule verbatim. A third tested route covers the fallback. |
| C8 plan-artifact authorship by Ship | Removed. The matrix is Stage-authored in the plan; Ship produces no plan artifact. The per-task tests encode the matrix rows. |

## 6. O-4 - unit re-scoped

Three tasks (026.001-T, 026.002-T, 026.003-T), strictly sequential, one covering feature, one shipment (plan
section "Implementation Units"). Size/complexity per task reasonable; no
single-task constraint.

## 7. Bootstrap and DB12DA37

Executing 026-F itself runs on the unrepaired Ship (Step 2 vacuous). The plan's common protocol therefore requires, for EVERY 026 task, a harness-architect invocation with the task ID explicit before Step 4.2, red confirmed and harness-ready applied by the skill, plus red-then-green evidence and a P-005 annotation. That gives P-002 Precondition and P-004 postcondition at implementation. The O-2 adjudication is Stage's interpretation of installed text; it is not an operator waiver and is not claimed as literal compliance with the Statement's wording about the shipment claim. If plan-review still returns an unresolved P0/P1 needing operator authority, Stage halts (no shipment).

## 8. Out of scope (unchanged)

025-F / 025.001-T / 019-F tasks (B2-B6, Orchestrator Step 1.5 handback), 020-F
and shipment 050-S (held until 026-F ships), `3750C37C`, `A92E3FA0` and other
stash entries. P-002/P-004/Step 4.3/Step 0.5 claim text are not edited.
