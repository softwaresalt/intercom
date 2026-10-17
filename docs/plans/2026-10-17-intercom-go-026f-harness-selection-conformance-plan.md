---
title: "Implementation plan: Ship harness selection conformance repair (026-F, DB12DA37), re-plan rev 2"
date: 2026-10-17
status: draft
source: docs/decisions/2026-10-17-intercom-go-026f-o1-closure-and-surface-adjudication-decision.md
supersedes: docs/plans/2026-09-18-intercom-go-harness-selection-conformance-plan.md (single-unit shape; plan-review attempt 1 FAIL)
feature: 026-F
tasks: [026.001-T, 026.002-T, 026.003-T]
---

# Implementation plan - 026-F re-plan (rev 2)

## Problem Frame

`backlogit shipment claim` (Ship Step 0.5 item 4) auto-transitions manifest tasks `queued -> active` before Ship Step 2. Step 2 item 1 (`_ship.agent.md:322`) lists only `queued` tasks, which is then empty; item 4 (`:329`) passes vacuously; Step 3 item 2 (`:362-367`) replaces the P-002 Enforcement filter ("`harness-ready` AND `queued`") with Step 3 item 1's derived executable set, discarding BOTH predicates. Net: Step 4 can execute a task that was never harnessed (P-004 red-phase bypass; DB12DA37). `workflow-policies.md:50` already requires the `harness-ready` filter, so this is a conformance repair of Ship, not a policy amendment.

A second, latent defect must be handled by the repair itself: a naive repair makes Step 2 scaffold the WHOLE manifest batch up front. For any multi-task shipment with test-bearing tasks that leaves several red harnesses at once, and Step 4.3's full `go test ./...` gate fails after the first task (the G1 deadlock; it was masked on the shipment route only because Step 2 was vacuous). With the one-task rule withdrawn, the repair must make harness generation **just-in-time per task** on the shipment route.

Operator authority (2026-10-17): O-1 closed, Option B (substring/structural tests are an admissible acceptance surface; no linter engine needed); one-task rule superseded. O-1..O-4 and C1-C8 dispositions: see the source decision.

## Design (the repaired contract on the shipment route)

1. **Route discriminator (C7).** The *shipment route* is any run in which a `shipment_id` is recorded as the session scope (Step 0.5 primary item 5 or fallback item 4). Both routes claim and auto-activate. Any other run keeps today's rules verbatim (legacy route).
2. **One derivation (C1/C2).** Step 3 item 1's derived executable set (`queued` or `active` task members, manifest-filtered) is the single read-only derivation, evaluated FIRST and once by Step 2 on the shipment route and reused by Step 3. It is the task-MEMBERSHIP list, not the execution boundary.
3. **Just-in-time per-task harness (G1).** On the shipment route Step 2 item 3 runs for ONE task at a time, in dependency order, immediately before that task's Step 4.1, invoking `harness-architect` with that task's ID passed explicitly as `tasks`; Ship does not proceed until the skill confirms Compilation PASS, Red Phase CONFIRMED and the `harness-ready` label is applied. At most one task's red harness exists at any Step 4.3 gate.
4. **Per-task P-002 gate at Step 4.1.** Before a task is claimed/confirmed `active` and implemented, Ship verifies it carries `harness-ready`; a task lacking it is a FAIL-CLOSED HALT (never a silent skip, never a proceed) citing P-002's Violation Action. This is the single execution gate on the shipment route (C2). The Step 0.5 shipment claim is a backlog side effect and is not the P-002 gate.
5. **harness-architect** accepts an explicitly listed `active` task as in scope and no longer excludes it as "non-ready".

## Constitution Check

* **P-002 (O-2 adjudication, decision section 3).** Not amended. Stated precisely: P-002's Gate Point is "Queue building (Step 2) and task claiming (Step 3)"; the Ship act gated is selecting a task for implementation (Step 4.1 Claim Task and Step 4.2). This plan makes the Precondition (`harness-ready`) and Enforcement hold for each task immediately before it is claimed and implemented. The adjudication is Stage's interpretation of installed text, recorded for plan-review; an operator may ratify it, but it is not asserted as an operator waiver and the plan does not claim more than the text supports.
* **P-004.** Each task's harness is produced by `harness-architect`, which confirms red and only then applies the label (its postcondition). Each test function in THIS feature's own test file is red first with an expected failure marker.
* **P-003.** Parent feature 026-F; every task has parent, >=1 AC (listed per task below), size and complexity (complexity as prose; the task type has no structured field). No intermediate sub-epic exists in this workspace's feature->task hierarchy (tasks parent directly to the covering feature, as for all prior features).
* **P-006** hardening below; **P-010** Stage writes no source; **P-016** no extra worktrees.
* **Protected, byte-unchanged:** P-002 and P-004 text, Step 4.3 text, Step 0.5 item 4 claim text (I1/I2).
* **IV-5 / 022.001-T.** The behavioural-fixture preference is NOT applied to 026-F only, by operator O-1 Option B; it stays in force elsewhere. The needle tests prove the contract WORDING, not Ship behaviour; the behavioural check is the Runtime Verification run.

## Requirements Trace

| Req | Action | Task |
|---|---|---|
| Per-task P-002 gate, membership vs execution boundary, empty-queue, broadcast, Step 3 trigger on route | `_ship.agent.md` Step 3 items 1, 2, 4, preamble; Step 4.1 | 026.001-T |
| Route discriminator, single derivation first, JIT per-task harness, non-vacuous guard | `_ship.agent.md` Step 2 preamble, items 1-4, closing paragraph; Step 3 item 1 single-derivation note | 026.002-T |
| harness-architect scope and active-task handling | `harness-architect/SKILL.md` lines 12, 18, 38, 50, 54-56, 60 | 026.003-T |
| Invariants I1-I7 and inventory closure | shipment-level PR-readiness ACs (below) | all |

## Surface matrix (whole; O-3; bound to origin/main at plan time; built before any edit)

| # | Role | Site | Current | Action | Task | Verified by |
|---|---|---|---|---|---|---|
| M1 | Producer preamble | `_ship.agent.md:318` | "every task in the target feature or chore" | "every task in the selected batch (item 1)" | 002 | T2a |
| M2 | Producer selector | `:322` | feature-scoped `queued` only | route discriminator; shipment-route batch = Step 3 item 1 derivation evaluated first; legacy `queued` clause retained verbatim | 002 | T2a |
| M3 | Producer partition | `:324-325` | partition of listed tasks | "Partition the selected batch" | 002 | T2b |
| M4 | Producer invocation | `:326` (item 3) | batch invocation up front | shipment route: JIT one task at a time before its Step 4.1, task ID passed as `tasks`; legacy route unchanged | 002 | T2b |
| M5 | Producer vacuity guard | `:329` (item 4) | "every queued task now carries" | confirm the current task (shipment route) or every task in the batch (legacy); empty selected batch on shipment route is a halt | 002 | T2b |
| M6 | Producer closing para | `:331-333` | "prefer the queue operation to assemble the task set" | add: on the shipment route the batch comes from the Step 3 item 1 derivation, never from a queued-only queue operation | 002 | T2b |
| M7 | Consumer preamble | `:337` | "Now that all tasks are harnessed" | "Now that the selected batch is identified (Step 2 item 1)" | 001 | T1c |
| M8 | Consumer derivation | `:338-361` item 1 | declares itself "the actual task-membership boundary that Step 4 executes"; trigger "Stage-prepared shipment" | membership boundary; trigger = shipment route; execution gate is Step 4.1 | 001 (boundary, trigger), 002 (single-derivation note) | T1a, T2a |
| M9 | Consumer queue | `:362-367` item 2 | substitution drops `harness-ready`; trigger "Stage-prepared shipment" | trigger = shipment route; membership list in dependency order; P-002 check applied per task at Step 4.1 | 001 | T1b |
| M10 | Consumer empty queue | `:369` item 4 | "empty after harness generation" | covers a shipment-route halt | 001 | T1c |
| M11 | Consumer broadcast | `:371` | "count of queued items" | "count of ready-queue tasks" (`:968` table row reads `{count} tasks` and is unchanged) | 001 | T1c |
| M12 | Per-task gate | `:377-379` Step 4.1 | status -> active | add the per-task P-002 gate (Design 3-4) before the claim | 001 | T1c |
| M13 | Skill selection | `harness-architect/SKILL.md:12,18,38,50,54,56,60` | "ready" scoping; "otherwise non-ready" | default `queued`; explicitly listed `active` in scope; exclude blocked/done/archived | 003 | T3a |
| M14 | Authority grant | `workflow-policies.md` P-002 (36-52), P-004 (84-92) | mandates filter | **PROTECTED** | - | git diff |
| M15 | Anti-regression | `_ship.agent.md` Step 4.3 (435-447), Step 0.5 item 4 claim | full-suite gate; claim | **PROTECTED** | - | git diff -U0 |
| M16 | Consistent, no change | `build-feature/SKILL.md:12`, `_ship.agent.md:431` harness_cmd | consistent | verified unchanged | - | git diff |
| M17 | Consistent, no change | `backlog-integration.instructions.md:124` "Listing Ready Tasks" | generic backlog query example, not a Ship selector | verified unchanged | - | git diff |

**Stop rule.** If a finding reappears at an adjacent surface, Ship halts, records a DEFERRED/BLOCKED note on the task, and writes NO plan or matrix artifact (C8; Ship's P-010 boundary); Stage rebuilds the matrix.

## Pinned wording (needle source; full clauses, never bare tokens)

* **PW-1 (M8/001)**: replace `it is the actual task-membership boundary` ... `that Step 4 executes, so an \`active\` member of this derived set is never omitted from the ready queue,` with `it is the task-MEMBERSHIP boundary (which tasks are candidates), not the execution boundary: the single execution gate is the per-task P-002 check at Step 4.1, so an \`active\` member of this derived set is never omitted from the ready queue's candidates,`; and in item 1 and item 2 replace the trigger `when operating under a Stage-prepared shipment` with `when on the shipment route (the Step 2 item 1 route discriminator)`.
* **PW-2 (M9/001)**: in item 2, after `regardless of whether its status is \`queued\` or \`active\`,` insert `in dependency order; on the shipment route this list is a membership list only, and the P-002 \`harness-ready\` check is applied to each task at Step 4.1 after its just-in-time harness (Step 2 item 3), where a member lacking the label is a FAIL-CLOSED HALT,`.
* **PW-3 (M7/M10/M11/001)**: `Now that the selected batch is identified (Step 2 item 1), construct the execution queue:`; item 4 gains `; on the shipment route an empty queue while the derived executable set is non-empty is likewise a halt`; broadcast `with the count of ready-queue tasks`.
* **PW-4 (M12/001)**: new paragraph at the start of Step 4.1: `**P-002 per-task gate (shipment route)**: before claiming this task, run Step 2 item 3 for this single task (just-in-time harness), then verify the task carries the \`harness-ready\` label. A task lacking it is a FAIL-CLOSED HALT, never a silent skip and never a proceed, citing P-002's Violation Action ("Halt and suggest running the harness-architect"). The Step 0.5 shipment claim has already set manifest tasks \`active\` as a backlog side effect; it is not the P-002 gate.`
* **PW-5 (M1-M6/002)**: preamble `every task in the selected batch (item 1)`; item 1 `Select the batch. **Route discriminator**: the *shipment route* is any run in which a \`shipment_id\` is recorded as the session scope (Step 0.5 primary item 5 or fallback item 4); on that route the Step 0.5 claim has already activated manifest tasks. On the shipment route the batch is the Step 3 item 1 derived executable set, admitting both \`queued\` and \`active\` members; Ship evaluates that single read-only derivation FIRST, once, and Step 3 reuses it. On any other route, list all tasks for the target feature or chore that are in \`queued\` status.`; item 2 `Partition the selected batch`; item 3 adds `On the shipment route, scaffold ONE task at a time in dependency order, immediately before that task's Step 4.1, passing that task's ID explicitly as \`tasks\` (a whole-batch up-front scaffold would leave several red harnesses at once and break the Step 4.3 full-suite gate after the first task); do not proceed until Compilation PASS, Red Phase CONFIRMED and the \`harness-ready\` label are recorded. On any other route, invoke the skill for the batch as before.`; item 4 `On the shipment route confirm the current task now carries the \`harness-ready\` label; on any other route confirm every task in the batch does. ... An empty selected batch on the shipment route is a halt, never a vacuous pass.`; closing paragraph gains `On the shipment route the batch comes from the Step 3 item 1 derivation, never from a queued-only queue operation.`; Step 3 item 1 gains `(This is the single derivation Step 2 item 1 evaluates first on the shipment route; it is not recomputed.)`.
* **PW-6 (M13/003)**: description `defaults to all \`queued\` tasks`; line 18 `ready work items` -> `in-scope work items`; broadcast `Tasks in scope`; line 50 `When omitted, use all \`queued\` tasks under the feature. An \`active\` task is in scope only when it is listed explicitly in \`${input:tasks}\`.`; Step 1 heading `Select the in-scope task set`; item 1 `Load the feature or chore and its in-scope descendants (\`queued\`, plus explicitly listed \`active\`)`; item 3 `Exclude blocked, done, and archived work items. An explicitly listed \`active\` task is in scope and MUST NOT be excluded as non-ready.`

## Implementation Units

**Common protocol (every task).** Posture TEST-FIRST, just-in-time: the task's first commit is its own red test additions, written only when the task starts and never pushed red. The red marker is an assertion failing because the pinned wording is ABSENT (and, where stated, the old wording is still PRESENT). **Per-task harness gate:** before Step 4.2 of each task, Ship invokes `harness-architect` with that task's ID explicitly as `tasks`, confirms Compilation PASS and Red Phase CONFIRMED, and applies `harness-ready` through the skill; Step 4.2 does not start otherwise. The task END commit has the full `go test ./...` green. Tasks are strictly sequential 001 -> 002 -> 003. Needles are whole pinned clauses cut from PW-1..PW-6, sliced by their own list-item/step anchors (a local `section(content, startAnchor, endAnchor)` helper), never bare tokens; every assertion resolves in its own subject; "old text ABSENT" assertions use the exact superseded clause. The red and green outputs are recorded by running the exact test command against the real file before and after. Every task's PR readiness block records: (a) red-then-green evidence; (b) the session-start commit SHA of `_ship.agent.md` and `harness-architect/SKILL.md` and a statement that Ship applied the Step 2/3 text as loaded at session start (any mid-run read of edited text is reported); (c) the baseline `git grep` hit list and closure result below; (d) the P-005 annotation: "DB12DA37 claim-ordering condition disclosed; per-task harness-architect gate and red-then-green evidence recorded; O-2 adjudication is Stage's interpretation (decision section 3), no operator waiver asserted".

**Shipment-level closure ACs (in every task's readiness block).** (i) `git diff origin/main...HEAD --name-only` lists only `.github/agents/_ship.agent.md`, `.github/skills/harness-architect/SKILL.md`, `tests/integration/harness_selection_conformance_test.go` (so `workflow-policies.md` is absent: I2); (ii) `git diff origin/main...HEAD -U0 -- .github/agents/_ship.agent.md` has hunk ranges disjoint from Step 0.5 item 4 and Step 4.3 (I1); (iii) inventory: run `git grep -n -E "harness-ready|ready tasks|ready descendants|ready work items|non-ready|every task in the target|all tasks are harnessed|every queued task|queued status|queued items|task-membership" -- .github` at the start of 001 and record the baseline hit list; at the end every hit must be either in the baseline or inside a matrix row M1-M13, and every changed baseline hit must be in M1-M13; any other hit STOPS the run (stop rule).

### U1 - 026.001-T: Ship Step 3 and Step 4.1 consumer conformance (Size M, complexity medium)

Files (2): `.github/agents/_ship.agent.md` (M7-M12), `tests/integration/harness_selection_conformance_test.go` (NEW; adds local helpers `readSurface(t, relpath)` and `section(...)`, and reuses `repoRoot`, `assertPresent`, `assertAbsent`, `assertOrderedBefore`, `countOccurrences` that already exist in package `integration`). Tests (3): **T1a** boundary + trigger (PW-1; old "actual task-membership boundary" and "when operating under a Stage-prepared shipment" ABSENT from item 1/2 slices); **T1b** item 2 dependency-order/membership-only/per-task-check clause (PW-2, inside the item 2 slice, after the `queued` or `active` clause); **T1c** preamble, item 4, broadcast, and the Step 4.1 per-task P-002 gate with halt and Violation Action citation (PW-3, PW-4; old "Now that all tasks are harnessed" and "count of queued items" ABSENT). ACs: AC-1 each pinned clause present in its own slice; AC-2 each superseded clause absent; AC-3 Step 4.3, Step 0.5 item 4 and policy text byte-identical (closure i-ii); AC-4 full suite green; plus common protocol.

### U2 - 026.002-T: Ship Step 2 producer conformance (Size M, complexity medium)

Files (2): `_ship.agent.md` (M1-M6 and the Step 3 item 1 single-derivation note), the test file (append). Tests (2): **T2a** `TestStep2_ShipmentRoute_BatchFromSingleDerivation` (preamble; item 1 discriminator naming both Step 0.5 primary item 5 and fallback item 4; derivation evaluated first and once; `queued` and `active`; the FULL legacy sentence "On any other route, list all tasks for the target feature or chore that are in `queued` status."; the single-derivation note inside Step 3 item 1; old "every task in the target feature or chore" ABSENT); **T2b** `TestStep2_JustInTimeTaskAndNonVacuousGuard` (partition over the selected batch; item 3 JIT one-task-at-a-time with the explicit `tasks` clause and the Step 4.3 rationale; item 4 current-task confirmation and empty-batch halt; closing paragraph; old "every queued task now carries" ABSENT). ACs mirror U1.

### U3 - 026.003-T: harness-architect scope (Size S, complexity medium)

Files (2): `.github/skills/harness-architect/SKILL.md` (M13), the test file (append). Tests (1): **T3a** `TestHarnessArchitect_ExplicitActiveTaskInScope` (PW-6 needles incl. line 50 and item 3 wording; old "defaults to all ready tasks", "ready work items" and "Exclude blocked, done, or otherwise non-ready work items" ABSENT in that file only). ACs mirror U1; this task also runs and records the closure inventory (iii) as the final state.

## Dependency Graph

`026.001-T -> 026.002-T -> 026.003-T` strictly sequential (blocks edges). 026-F is the covering feature. 025-F/019-F chain is NOT planned here; 020-F/050-S stay held until 026-F ships.

## Decisions and Rationale

* Per-task JIT harness on the shipment route is part of the repair, not an add-on: without it the repair re-creates the G1 deadlock for every multi-task shipment (masked today only by Step 2's vacuity).
* The execution gate moves to Step 4.1 per task because Step 3 runs once before any JIT harness exists; Step 3 item 2 becomes membership only. This keeps ONE execution gate (C2) and ONE derivation (C1, decision D6).
* Needle tests are the acceptance surface by operator determination (O-1 Option B); IV-5 stays in force elsewhere.
* The branch-locus note of the earlier plan is dropped: it was not required by the repair (YAGNI) and JIT harness authoring happens on the shipment's own branch by Ship's existing topology gate.
* Three tasks, strictly sequential, each <=2 files, <=3 test functions.

## Risks and Caveats

* **DB12DA37 bootstrap.** 026-F's own run uses the unrepaired Ship (Step 2 vacuous). Per-task `harness-architect` invocation before Step 4.2 (common protocol) gives P-002 Precondition and P-004 postcondition at implementation for every 026-F task. Residual: Ship may skip the protocol; the readiness evidence is merge-gating for the operator.
* **Existing tests** may pin edited phrases; the full-suite gate catches this; Ship updates only phrases it changed and reports any other failure (never weakens an unrelated test). Pre-task check: `git grep -n -E "harness-ready|queued|task-membership boundary|ready tasks" -- tests/`.
* **Wording tests can be gamed**; accepted by the operator; the inventory grep is the second defense.
* **Mid-branch states** (after 001 only): Step 4.1 gate halts fail-closed if a task lacks the label, never executes unharnessed; both land in one PR.
* **Not planned here**: 025-F/025.001-T, 019.004/008/009-T (each needs its own re-plan and review).
* **P-021**: the empty `**Format**: ` ` `` gate at `_ship.agent.md:440` is out of scope; capture as DEFERRED SCOPE EXPANSION if still present.

## Plan Hardening Signals

* Public API/contract change: PRESENT. Security/compliance-sensitive: PRESENT (P-002/P-004 enforcement path). Migration/destructive: ABSENT. External integration/operator checkpoint: PRESENT (merge approval pre-authorized for scope). High rollout/rollback risk: PRESENT, moderate (a wrong halt stops every shipment run).

Requires plan hardening: yes

## Runtime Verification and Closure

Runtime surface: Ship agent behaviour on the next shipment run. Verification: the three test groups green; the first subsequent multi-task shipment run (050-S after its hold is lifted) shows Step 2 selecting a non-empty batch, one task at a time harnessed JIT, the Step 4.1 per-task gate enforced, and Step 4.3 green at every task, recorded in its readiness block. Rollback: revert the 026-F release as a unit (three files); never a single clause.

## Plan Hardening

**Required:** yes. **Consulted:** `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` (needles anchored to their own slice; run the exact command against the real file before and after), `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` (no manual substitution of skills), `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md` (bounded rule surface; no fake reds), the 2026-09-18 review (C1-C8), the 020-F review record (constitution P1 class).

| ProposedAction | ActionRisk | Approval |
|---|---|---|
| Edit Ship Step 4.1 per-task gate | high (wrong halt stops every run) | PR review; merge approval pre-authorized for scope |
| Edit Ship Step 2/3 selection | high | same |
| Edit harness-architect scope | medium | same |
| Add substring contract tests | low | none |

**Invariants.** I1 Step 4.3 and Step 0.5 claim text byte-unchanged; I2 P-002/P-004 text byte-unchanged; I3 legacy route keeps the feature-scoped `queued` rule verbatim; I4 no task reaches Step 4.2 on the shipment route without `harness-ready` applied by harness-architect after confirmed red; I5 single derivation, single execution gate; I6 no manifest/witness/selector-flag mechanism reintroduced (rev-11..rev-15 non-governing); I7 structural integrity of both edited files outside the edited clauses (heading/list read-back).

**Verification depth.** Pre-task baseline (clean branch, `go vet ./...`, `go test ./...` green); red; green. Blocked path: if a pinned edit breaks an unrelated existing test, Ship stops and reports.

**Closure.** Monitoring: next two multi-task shipment runs. Rollback trigger: a run halting at Step 2/4.1 for a task that carries `harness-ready`, or executing an unharnessed task, or a Step 4.3 failure caused by simultaneous red harnesses. Rollback: revert as a unit. Owner: repository owner. Validation window: next two shipment runs.

**Review-gate capability risk.** plan-review must emit literal `dispatch_mode:` and `decision:` markers; degraded dispatch does not relax FAIL on any P0/P1.

**Unresolved operator decisions.** None blocking. O-2 is recorded as Stage's interpretation; operator ratification is welcome but not required if plan-review accepts that the per-task harness-architect gate satisfies P-002 Precondition and P-004 postcondition at implementation. Stage halts if plan-review still returns a P0/P1 needing operator authority.

## Plan Review - attempt 2 of rev 2 (FAIL; escalation threshold reached)

`dispatch_mode: single-agent-declared-degradation` for every persona (no nested dispatch); `decision: FAIL`. Personas: Constitution FAIL (2 P1), Architecture FAIL (1 P1), Scope Boundary ADVISORY (3 P2); the earlier attempt-1 FAIL of this rev (Constitution FAIL, Architecture FAIL, Scope ADVISORY, Learnings FAIL) led to the rev-2 redesign above. Attempt-1 findings P1-1 (route discriminator vs Step 3 trigger) and P1-2 (whole-batch harness) are confirmed CLOSED by rev 2.

**Remaining P1s (all closable by plan text; none needs an operator waiver; the O-2 adjudication was held defensible):**

1. **Constitution P1-A - bootstrap order.** The shipment claim has already set 026.001-T..003-T `active`, and `harness-architect/SKILL.md:38` ("otherwise non-ready") is fixed only by the LAST task (M13), so the per-task harness-architect protocol cannot run for the earlier tasks. Remediation: reorder so the skill task is FIRST; state that, for that first task only, harness-architect relies on its explicit-`tasks` precedence (reading: an explicitly listed `active` task is not "non-ready"); if the skill refuses, Ship halts fail-closed and reports (no bypass). Delete any claim that the protocol "satisfies" P-002/P-004 for every task until true.
2. **Constitution P1-B - red marker.** `harness-architect` Step 5.2 requires failure markers `panic("not implemented: <reason>")`; the plan's red state (substring assertion failure) would be rejected and the label withheld. Remediation: pin `t.Fatalf("not implemented: pinned wording absent: ...")` as the marker; the harness-architect skill (not Ship) authors/confirms the red test file as each task's first commit; scope the red-phase check to the task's new test functions with `-run`.
3. **Architecture P1 - Step 2 line 316 contradiction.** `_ship.agent.md:316` says Step 2 "runs once, up front - not in a loop" and "before any implementation begins"; the JIT design contradicts it and no matrix row covers it. Remediation: add matrix row M1b + PW clause ("legacy route: once, up front; shipment route: just-in-time per task (item 3), before that task's Step 4.1; the up-front rule does not apply"), a T2a needle, and an ABSENT assertion for the old sentence.

**P2 remediation queue (apply with the P1s):** order tasks skill -> Step 2 -> Step 3/Step 4.1 (resolves the dangling reference after the first task); Step 2 item 2 "(on the shipment route: the current task)"; Step 3 preamble/item 4 "empty after membership derivation" and broadcast "candidate tasks"; align "before Step 4.1" everywhere; collapse whitespace in `readSurface`/`section` and pin the item 2 trigger as "When operating under a Stage-prepared shipment (item 1 above ran)"; key the inventory baseline by file + matched text (every NEW hit text must be a PW-pinned clause); keep T-scenarios <=3 per task; note the "harness generation completes" checkpoint trigger and harness-architect Step 5.2 scoping as explicit rows; widen the inventory to "up front|before any implementation|runs once" over `.github/agents/_orchestrator.agent.md`, `_stage.agent.md`, prompts, instructions.

<!-- plan-review-attempt: 1 --> FAIL (rev-2 first pass; redesigned)
<!-- plan-review-attempt: 2 --> FAIL (threshold: 2 consecutive FAILs); remediation queue above is UNAPPLIED; Stage stops per Escalation Protocol
