---
title: "Implementation plan: Ship harness selection conformance repair (026-F, DB12DA37), re-plan rev 3"
date: 2026-10-17
status: draft
source: docs/decisions/2026-10-17-intercom-go-026f-o1-closure-and-surface-adjudication-decision.md
supersedes: docs/plans/2026-09-18-intercom-go-harness-selection-conformance-plan.md (rev 1, FAIL); rev 2 of this file (git history, FAIL x2)
feature: 026-F
tasks: [026.001-T, 026.002-T, 026.003-T, 026.004-T]
---

# Implementation plan - 026-F re-plan (rev 3)

## Problem Frame

`backlogit shipment claim` (Ship Step 0.5 item 4) auto-transitions manifest tasks `queued -> active` before Ship Step 2. Step 2 item 1 (`_ship.agent.md:322`) lists only `queued` tasks, which is then empty; item 4 (`:329`) passes vacuously; Step 3 item 2 (`:362-367`) replaces the P-002 Enforcement filter ("`harness-ready` AND `queued`") with Step 3 item 1's derived executable set, discarding BOTH predicates. Net: Step 4 can execute a task that was never harnessed (P-004 red-phase bypass; DB12DA37). `workflow-policies.md:50` already requires the `harness-ready` filter, so this is a conformance repair of Ship, not a policy amendment.

A latent defect must be handled by the repair itself: a naive repair makes Step 2 scaffold the WHOLE manifest batch up front. For a multi-task shipment with test-bearing tasks that leaves several red harnesses at once and Step 4.3's full `go test ./...` gate fails after the first task (the G1 deadlock; masked on the shipment route only because Step 2 was vacuous). With the one-task rule withdrawn, the repair makes harness generation **just-in-time per task** on the shipment route.

Operator authority (2026-10-17): O-1 closed, Option B (substring/structural tests are an admissible acceptance surface); one-task rule superseded. O-1..O-4 and C1-C8: see the source decision.

## Design (the repaired contract on the shipment route)

1. **Route discriminator (C7).** The *shipment route* is any run in which a `shipment_id` is recorded as session scope (Step 0.5 primary item 5 or fallback item 4). Any other run is the *legacy route* and keeps today's rules verbatim.
2. **One derivation (C1/C2).** Step 3 item 1's derived executable set (`queued` or `active` task members, manifest-filtered) is the single read-only derivation, evaluated FIRST and once by Step 2 on the shipment route and reused by Step 3. It is a MEMBERSHIP list, not an execution boundary.
3. **Just-in-time per-task harness (G1).** On the shipment route Step 2 item 3 runs for ONE task at a time, in dependency order, immediately before that task's Step 4.1, invoking `harness-architect` with that task's ID passed explicitly as `tasks` (never the whole batch); Ship proceeds only after the skill confirms Compilation PASS, Red Phase CONFIRMED and the `harness-ready` label is applied. At most one task's red harness exists at any Step 4.3 gate.
4. **Per-task P-002 gate at Step 4.1.** Before a task is claimed/confirmed `active` and implemented, Ship verifies it carries `harness-ready`; a task lacking it is a FAIL-CLOSED HALT (never a silent skip, never a proceed) citing P-002's Violation Action. This is the single execution gate on the shipment route. The Step 0.5 shipment claim is a backlog side effect and is not the P-002 gate.
5. **harness-architect** treats an explicitly listed `active` task as in scope and no longer excludes it as "non-ready".

## Constitution Check

* **P-002 (O-2 adjudication, decision section 3).** Not amended and no waiver asserted. P-002's Gate Point is "Queue building (Step 2) and task claiming (Step 3)"; the Ship act gated is selecting a task for implementation (Step 4.1 Claim Task and Step 4.2). The plan makes the Precondition (`harness-ready`) and Enforcement hold for each task immediately before it is claimed and implemented. The adjudication is Stage's interpretation of installed text for plan-review; an operator may ratify it.
* **P-004.** Each task's harness is produced by `harness-architect`, which confirms red and only then applies the label (its postcondition). Red for THIS feature's tests is recognised exactly as the skill's Step 5.2 defines it (see "Red marker").
* **P-003.** Parent feature 026-F; every task has parent, >=1 AC, size and complexity (complexity as prose; the task type has no structured field). Tasks parent directly to the feature (no sub-epic level exists in this workspace's hierarchy, as for all prior features).
* **P-006** hardening below. **P-010** Stage writes no source. **P-016** no extra worktrees.
* **Protected, byte-unchanged:** P-002 and P-004 text, Step 4.3 text, Step 0.5 item 4 claim text.
* **IV-5 / 022.001-T.** The behavioural-fixture preference is NOT applied to 026-F only, by O-1 Option B; it stays in force elsewhere. The needle tests prove the contract WORDING, not Ship behaviour; the behavioural check is the Runtime Verification run.

## Red marker (how the substring failure is recognised as red, no policy change)

`harness-architect/SKILL.md` Step 5.2 requires each harness test to fail with the marker `panic("not implemented: <reason>")`. The 026 tests therefore assert through two local helpers in the test file, `requirePinned(t, surface, clause, id)` and `requireRetired(t, surface, clause, id)`, whose failure path is `panic(fmt.Sprintf("not implemented: pinned wording absent: %s", id))` (respectively `retired wording still present`). A missing or stale clause thus fails with exactly the skill's marker, so Step 5.2 recognises it and Step 6 may apply the label. (A drifted clause later panics the integration package loudly and names the clause id; accepted trade-off for marker conformance.) Because a panic aborts the test binary, the red-phase check runs EACH new test function separately (`go test ./tests/integration -run '^TestName$'`) and records the marker per function; the skill's whole-suite run is scoped this way to "the harness tests", as its own Step 5.2 text says. There is no production stub for a contract harness; that is recorded as "not applicable: contract text harness" in the skill's Step 4. The skill (not Ship) authors and confirms each task's red test file as that task's first commit.

## Requirements Trace

| Req | Action | Task |
|---|---|---|
| harness-architect scope and explicit-`active` handling | `harness-architect/SKILL.md` lines 12, 18, 38, 50, 54-56, 60 | 026.001-T |
| Route discriminator, single derivation first, JIT per-task harness, non-vacuous guard, once-up-front sentence | `_ship.agent.md` Step 2 preamble (`:318`), items 1-4, closing paragraph; Step 3 item 1 single-derivation note | 026.002-T |
| Membership boundary, route trigger, dependency order, empty queue, broadcast | `_ship.agent.md` Step 3 preamble, items 1, 2, 4; `:371` | 026.003-T |
| Per-task P-002 gate | `_ship.agent.md` Step 4.1 | 026.004-T |
| Invariants I1-I7, inventory closure | shipment-level readiness ACs | all |

## Surface matrix (whole; bound to origin/main at plan time; baseline inventory run at plan time)

| # | Role | Site | Current | Action | Task | Verified by |
|---|---|---|---|---|---|---|
| M1 | Producer preamble | `_ship.agent.md:318` | "Ensure every task in the target feature or chore has a passing test harness before any implementation begins. This step runs once, up front — not in a loop." | PW-5a (both sentences replaced with route-qualified text) | 002 | T2a |
| M2 | Producer selector | `:322` | feature-scoped `queued` only | route discriminator; shipment-route batch = Step 3 item 1 derivation evaluated first; legacy `queued` clause retained verbatim | 002 | T2a |
| M3 | Producer partition | `:324-325` | partition of listed tasks | "Partition the selected batch (on the shipment route: the current task)" | 002 | T2b |
| M4 | Producer invocation | `:326` item 3 | batch invocation up front | shipment route: JIT one task at a time before its Step 4.1, task ID as `tasks`; legacy unchanged | 002 | T2b |
| M5 | Producer vacuity guard | `:329` item 4 | "every queued task now carries" | confirm the current task (shipment route) or every task in the batch (legacy); empty selected batch on shipment route is a halt | 002 | T2b |
| M6 | Producer closing para | `:331-333` | "prefer the queue operation to assemble the task set" | add: on the shipment route the batch comes from the Step 3 item 1 derivation, never a queued-only queue operation | 002 | T2b |
| M7 | Consumer preamble | `:337` | "Now that all tasks are harnessed" | "Now that the batch membership is identified (Step 2 item 1), construct the candidate queue:" | 003 | T3b |
| M8 | Consumer derivation | `:338-361` item 1 (`:358-360`) | "the actual task-membership boundary that Step 4 executes"; trigger "Stage-prepared shipment" | membership boundary; trigger = shipment route; single-derivation note (note text lands in 002) | 003 (boundary, trigger), 002 (note) | T3a, T2a |
| M9 | Consumer queue | `:362-367` item 2 | substitution drops `harness-ready`; trigger "When operating under a Stage-prepared shipment (item 1 above ran)" | trigger = shipment route; membership list in dependency order; P-002 check applied per task at Step 4.1 | 003 | T3a |
| M10 | Consumer empty queue | `:369` item 4 | "empty after harness generation" | "(legacy route) after harness generation or (shipment route) after membership derivation" | 003 | T3b |
| M11 | Consumer broadcast | `:371` | "count of queued items" | "count of candidate tasks" (`:968` row reads `{count} tasks`, unchanged) | 003 | T3b |
| M12 | Per-task gate | `:377-379` Step 4.1 | status -> active | add the per-task P-002 gate before the claim | 004 | T4a |
| M13 | Skill selection | `harness-architect/SKILL.md:12,18,38,50,54,56,60` | "ready" scoping; "otherwise non-ready" | default `queued`; explicit `active` in scope; exclude blocked/done/archived | 001 | T1a |
| M14 | Authority grant | `workflow-policies.md` P-002 (36-52), P-004 (84-92) | mandates filter | **PROTECTED** | - | git diff |
| M15 | Anti-regression | `_ship.agent.md` Step 4.3 (435-447), Step 0.5 item 4 claim | gate; claim | **PROTECTED** | - | git diff -U0 |
| M16 | Unchanged, noted | `harness-architect/SKILL.md` Step 5.2 (`:124-127`) | "Run `go test ./...` for the harness tests" | already scoped to "the harness tests"; unchanged; scoping per Red marker | - | git diff |
| M17 | Unchanged, noted | `_ship.agent.md:1052` checkpoint trigger "harness generation completes" | fires after harness generation | unchanged; on the shipment route it fires per JIT harness, which is intended | - | git diff |
| M18 | Unchanged, consistent | `build-feature/SKILL.md:12`; `_ship.agent.md:23,431`; `backlog-integration.instructions.md:136`; `_stage.agent.md:363`; `prompts/stage-grouping-analysis.prompt.md:22` | `harness-ready` label use / queued-item wording unrelated to Ship selection | verified, no change | - | git diff |

**Baseline inventory (plan time, `git grep -n -E "<pattern>" -- .github`, pattern in the closure AC):** hits exist only at the sites listed in M1-M18 above (`_ship.agent.md` 4, 23, 89, 318, 324, 325, 329, 337, 358, 362, 369, 371, 431, 1052; `_stage.agent.md` 106, 363; `backlog-integration` 136; `stage-grouping-analysis.prompt` 17, 22; `brainstorm`/`runtime-verification`/`build-feature:29` "up front" prose unrelated; `build-feature` 12; `harness-architect` 12, 19(18 in view), 43, 50, 56, 60, 132, 136, 156). `.github/agents/_orchestrator.agent.md` has no hit.

**Stop rule.** If a finding reappears at an adjacent surface, Ship halts, records a DEFERRED/BLOCKED note on the task, and writes NO plan or matrix artifact (C8; Ship's P-010 boundary); Stage rebuilds the matrix.

## Pinned wording (needle source; whole clauses; matched after whitespace collapse)

`readSurface` collapses all whitespace runs to a single space (`strings.Fields` joined by " ") so wrapped source lines match; every clause below is written single-spaced.

* **PW-1 (M8/003)**: replace `it is the actual task-membership boundary that Step 4 executes, so an \`active\` member of this derived set is never omitted from the ready queue,` with `it is the task-MEMBERSHIP boundary (which tasks are candidates), not the execution boundary: the single execution gate is the per-task P-002 check at Step 4.1, so an \`active\` member of this derived set is never omitted from the candidate queue,`.
* **PW-2 (M9/003)**: item 2 trigger `When operating under a Stage-prepared shipment (item 1 above ran), replace` becomes `When on the shipment route (the Step 2 item 1 route discriminator; item 1 above ran), replace`; and after `regardless of whether its status is \`queued\` or \`active\`,` insert `in dependency order; on the shipment route this list is a membership list only, and the P-002 \`harness-ready\` check is applied to each task at Step 4.1 after its just-in-time harness (Step 2 item 3), where a member lacking the label is a FAIL-CLOSED HALT,`. Item 1's trigger `(when operating under a Stage-prepared shipment; C1–C6,` (line 339) is changed to `(when on the shipment route; C1–C6,`; add that old trigger phrase to the Retired list.
* **PW-3 (M7/M10/M11/003)**: `Now that the batch membership is identified (Step 2 item 1), construct the candidate queue:`; item 4 `If the queue is empty after harness generation (legacy route) or after membership derivation (shipment route), halt and report — there is nothing to build.`; broadcast `with the count of candidate tasks`.
* **PW-4 (M12/004)**: new first paragraph in Step 4.1: `**P-002 per-task gate (shipment route)**: before claiming this task, run Step 2 item 3 for this single task (just-in-time harness), then verify the task carries the \`harness-ready\` label. A task lacking it is a FAIL-CLOSED HALT, never a silent skip and never a proceed, citing P-002's Violation Action ("Halt and suggest running the harness-architect"). The Step 0.5 shipment claim has already set manifest tasks \`active\` as a backlog side effect; it is not the P-002 gate.`
* **PW-5 (M1-M6/002)**: (a) preamble `Ensure every task in the selected batch (item 1) has a passing test harness before that task's implementation begins. On the shipment route this step runs just-in-time per task (item 3), before that task's Step 4.1; on any other route it runs once, up front, not in a loop.`; (b) item 1 `Select the batch. **Route discriminator**: the *shipment route* is any run in which a \`shipment_id\` is recorded as the session scope (Step 0.5 primary item 5 or fallback item 4); on that route the Step 0.5 claim has already activated manifest tasks. On the shipment route the batch is the Step 3 item 1 derived executable set, admitting both \`queued\` and \`active\` members; Ship evaluates that single read-only derivation FIRST, once, and Step 3 reuses it. On any other route, list all tasks for the target feature or chore that are in \`queued\` status.`; (c) item 2 `Partition the selected batch (on the shipment route: the current task)`; (d) item 3 adds `On the shipment route, scaffold ONE task at a time in dependency order, immediately before that task's Step 4.1, passing that task's ID explicitly as \`tasks\` (a whole-batch up-front scaffold would leave several red harnesses at once and break the Step 4.3 full-suite gate after the first task); do not proceed until Compilation PASS, Red Phase CONFIRMED and the \`harness-ready\` label are recorded. On any other route, invoke the skill for the batch as before.`; (e) item 4 `On the shipment route confirm the current task now carries the \`harness-ready\` label; on any other route confirm every task in the batch does.` ... `An empty selected batch on the shipment route is a halt, never a vacuous pass.`; (f) closing paragraph gains `On the shipment route the batch comes from the Step 3 item 1 derivation, never from a queued-only queue operation.`; (g) Step 3 item 1 gains `(This is the single derivation Step 2 item 1 evaluates first on the shipment route; it is not recomputed.)`.
* **PW-6 (M13/001)**: description `defaults to all \`queued\` tasks`; line 18 `ready work items` -> `in-scope work items`; broadcast `Tasks in scope`; line 50 `When omitted, use all \`queued\` tasks under the feature. An \`active\` task is in scope only when it is listed explicitly in \`${input:tasks}\`.`; Step 1 heading `Select the in-scope task set`; item 1 `Load the feature or chore and its in-scope descendants (\`queued\`, plus explicitly listed \`active\`)`; item 3 `Exclude blocked, done, and archived work items. An explicitly listed \`active\` task is in scope and MUST NOT be excluded as non-ready.`

Retired (ABSENT after the owning task): `every task in the target feature or chore`; `This step runs once, up front — not in a loop.`; `confirm every queued task now carries`; `Now that all tasks are harnessed`; `it is the actual task-membership boundary that Step 4 executes`; `When operating under a Stage-prepared shipment (item 1 above ran)`; `with the count of queued items`; (skill) `defaults to all ready tasks`; `ready work items`; `Exclude blocked, done, or otherwise non-ready work items`.

## Implementation Units

**Common protocol (every task).** Strictly sequential 001 -> 002 -> 003 -> 004; order resolves the bootstrap (the skill fix lands first) and avoids forward references (Step 3 and Step 4.1 cite Step 2 text that already exists). Posture TEST-FIRST, just-in-time: the harness for task N is authored only when task N starts, by `harness-architect` invoked with that task's ID passed explicitly as `tasks` BEFORE Step 4.1 (claim) and Step 4.2; Compilation PASS and Red Phase CONFIRMED per function (Red marker section) are recorded and `harness-ready` is applied by the skill; Step 4.1 does not start otherwise. The red test commit is never pushed red; the task END commit has the full `go test ./...` green. **Bootstrap for 026.001-T:** the shipment claim has already set it `active` under the unrepaired skill, whose Step 1 item 2 says an explicit `tasks` set restricts the scope and whose item 3 excludes "otherwise non-ready" work; reading R = an explicitly listed `active` task is not "non-ready" (it is ready and in progress). If the skill refuses 026.001-T, Ship halts fail-closed and reports; it never bypasses. Needles are whole clauses from PW-1..PW-6 sliced by their own step/list anchors via a local `section` helper; every assertion resolves in its own subject.

**Readiness-block evidence (every task).** (a) per-function red marker output then green output (exact commands); (b) session-start commit SHA of the instruction surfaces and a statement that Ship applied Step 2/3 text as loaded at session start (any mid-run read of edited text is reported); (c) baseline-vs-final inventory result; (d) the P-005 annotation: "DB12DA37 claim-ordering condition disclosed; per-task harness-architect gate and red-then-green evidence recorded; O-2 adjudication is Stage's interpretation (decision section 3); no operator waiver asserted".

**Closure ACs (every readiness block).** (i) `git diff origin/main...HEAD --name-only` lists only `.github/agents/_ship.agent.md`, `.github/skills/harness-architect/SKILL.md`, `tests/integration/harness_selection_conformance_test.go` (so `workflow-policies.md` is absent: I2); (ii) `git diff origin/main...HEAD -U0 -- .github/agents/_ship.agent.md` has hunk ranges disjoint from Step 0.5 item 4 and Step 4.3 (I1); (iii) inventory: `git grep -n -E "harness-ready|ready tasks|ready descendants|ready work items|non-ready|every task in the target|all tasks are harnessed|every queued task|queued status|queued items|task-membership|up front|before any implementation|runs once|harness generation" -- .github` compared with the baseline keyed by FILE + MATCHED TEXT (never line numbers): every hit text that is NEW relative to the baseline must be a PW-pinned clause, and every baseline hit text that CHANGED must belong to M1-M13; any other difference STOPS the run (stop rule).

### U1 - 026.001-T: harness-architect scope (Size S, complexity medium)

Files (2): `.github/skills/harness-architect/SKILL.md` (M13), `tests/integration/harness_selection_conformance_test.go` (NEW; local helpers `readSurface`, `section`, `requirePinned`, `requireRetired`; reuses `repoRoot` and `countOccurrences` already in package `integration`). Test (1): **T1a** `TestHarnessArchitect_ExplicitActiveTaskInScope` (PW-6 clauses present incl. line 50 and item 3; retired clauses absent, in that file only). ACs: AC-1 PW-6 clauses present; AC-2 retired clauses absent; AC-3 closure ACs (i)-(iii); AC-4 full suite green; plus common protocol.

### U2 - 026.002-T: Ship Step 2 producer (Size M, complexity medium)

Files (2): `_ship.agent.md` (M1-M6 and the Step 3 item 1 note), the test file (append). Tests (2): **T2a** `TestStep2_PreambleDiscriminatorAndSingleDerivation` (PW-5 a, b, g; retired preamble and old "every task in the target feature or chore" absent; the FULL legacy sentence retained); **T2b** `TestStep2_JustInTimeTaskPartitionGuardAndClosingParagraph` (PW-5 c, d, e, f; retired "confirm every queued task now carries" absent). ACs mirror U1.

### U3 - 026.003-T: Ship Step 3 consumer (Size M, complexity medium)

Files (2): `_ship.agent.md` (M7-M11), the test file (append). Tests (2): **T3a** `TestStep3Items1And2_MembershipBoundaryRouteTriggerDependencyOrder` (PW-1, PW-2; retired clauses absent within the item 1/2 slices); **T3b** `TestStep3PreambleEmptyQueueBroadcast` (PW-3; retired preamble and "with the count of queued items" absent). ACs mirror U1.

### U4 - 026.004-T: Ship Step 4.1 per-task P-002 gate and closure (Size S, complexity medium)

Files (2): `_ship.agent.md` (M12), the test file (append). Test (1): **T4a** `TestStep41_PerTaskP002GateFailClosed` (PW-4 inside the Step 4.1 slice, ordered before the "Update task status to `active`" sentence). This task also records the final closure inventory (iii). ACs mirror U1.

## Dependency Graph

`026.001-T -> 026.002-T -> 026.003-T -> 026.004-T` strictly sequential (blocks edges). 026-F is the covering feature. 025-F/019-F chain is NOT planned here; 020-F/050-S stay held until 026-F ships.

## Decisions and Rationale

* JIT per-task harness on the shipment route is part of the repair: without it the repair re-creates the G1 deadlock for every multi-task shipment.
* The execution gate sits at Step 4.1 per task because Step 3 runs once before any JIT harness exists; Step 3 item 2 becomes membership only (one derivation, one gate: C1/C2, decision D6).
* The panic-marker helpers make substring tests recognisable as red by the skill's own Step 5.2 without touching any policy.
* Order skill -> Step 2 -> Step 3 -> Step 4.1: bootstrap works and no task cites wording defined later.
* Four tasks, each <=2 files, <=2 test functions; no single-task constraint.
* IV-5 stays in force elsewhere (O-1 Option B is scoped to 026-F).

## Risks and Caveats

* **DB12DA37 bootstrap.** 026-F's own run uses the unrepaired Ship (Step 2 vacuous). The per-task `harness-architect` invocation before Step 4.1 provides P-002 Precondition and P-004 postcondition at implementation. Residual: Ship may skip the protocol; the readiness evidence is merge-gating for the operator.
* **Panic markers** abort the integration test binary on a drifted clause (loud, names the clause id).
* **Reading R for 026.001-T** (explicit `active` task in scope); refusal halts safely.
* **Existing tests** may pin edited phrases; checked at task start with `git grep -n -E "harness-ready|queued|task-membership boundary|ready tasks" -- tests/`; Ship updates only phrases it changed and reports any other failure.
* **Mid-branch states** fail closed (Step 4.1 gate halts on an unlabelled task); all land in one PR.
* **Not planned here**: 025-F/025.001-T, 019.004/008/009-T.
* **P-021**: the empty `**Format**: ` ` `` gate at `_ship.agent.md:440` is out of scope; capture as DEFERRED SCOPE EXPANSION if still present.

## Plan Hardening Signals

* Public API/contract change: PRESENT. Security/compliance-sensitive: PRESENT (P-002/P-004 enforcement path). Migration/destructive: ABSENT. External integration/operator checkpoint: PRESENT (merge approval pre-authorized for scope). High rollout/rollback risk: PRESENT, moderate (a wrong halt stops every shipment run).

Requires plan hardening: yes

## Runtime Verification and Closure

Runtime surface: Ship agent behaviour on the next shipment run. Verification: the four test groups green; the first subsequent multi-task shipment run (050-S after its hold is lifted) shows Step 2 selecting a non-empty batch, one task harnessed JIT at a time, the Step 4.1 gate enforced and Step 4.3 green at every task, recorded in its readiness block. Rollback: revert the 026-F release as a unit (three files); never a single clause.

## Plan Hardening

**Required:** yes. **Consulted:** `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` (needles sliced to their own subject, whitespace-collapsed; run the exact command against the real file before and after), `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`, `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md` (bounded rule surface; no fake reds), the earlier reviews (C1-C8; attempt findings of rev 2).

| ProposedAction | ActionRisk | Approval |
|---|---|---|
| Edit Ship Step 4.1 per-task gate | high (wrong halt stops every run) | PR review; merge approval pre-authorized for scope |
| Edit Ship Step 2/3 selection | high | same |
| Edit harness-architect scope | medium | same |
| Add substring contract tests with panic markers | low | none |

**Invariants.** I1 Step 4.3 and Step 0.5 claim text byte-unchanged; I2 P-002/P-004 text byte-unchanged; I3 legacy route keeps the feature-scoped `queued` rule verbatim; I4 no task reaches Step 4.2 on the shipment route without `harness-ready` applied by harness-architect after confirmed red; I5 single derivation, single execution gate; I6 no manifest/witness/selector-flag mechanism reintroduced; I7 structural integrity of both edited files outside the edited clauses.

**Verification depth.** Pre-task baseline (clean branch, `go vet ./...`, `go test ./...` green); red per function; green. Blocked path: an unrelated existing test broken by a pinned edit stops the run and is reported.

**Closure.** Monitoring: next two multi-task shipment runs. Rollback trigger: a run halting at Step 2/4.1 for a task that carries `harness-ready`, executing an unharnessed task, or a Step 4.3 failure from simultaneous red harnesses. Rollback: revert as a unit. Owner: repository owner. Validation window: next two shipment runs.

**Review-gate capability risk.** plan-review must emit literal `dispatch_mode:` and `decision:` markers; degraded dispatch does not relax FAIL on any P0/P1.

**Unresolved operator decisions.** None blocking. O-2 is Stage's interpretation; operator ratification welcome, not required. Stage halts if attempt 3 returns a P0/P1 needing operator authority.

<!-- plan-review-attempt: 2 --> prior FAIL (rev 2); rev 3 applies the full remediation queue; attempt 3 pending

## Plan Review - attempt 3 of 3 (rev 3): FAIL; cap reached, Stage halted

Six genuinely separate reviewer subagents were dispatched (Constitution, Correctness, Architecture, Scope Boundary, Security, Maintainability); each reported `dispatch_mode: single-agent` internally because reviewer subagents cannot nest, so no reviewer had a deeper sub-dispatch. `decision: FAIL` (aggregate; one persona-level P1 per FAIL below). Verdicts: Constitution **FAIL**, Scope Boundary **FAIL**, Correctness ADVISORY, Architecture ADVISORY, Security ADVISORY, Maintainability ADVISORY. Rev-2 findings A (bootstrap order), B (red marker), C (`:318` sentence) were confirmed CLOSED by rev 3.

**P1 needing OPERATOR authority (Constitution).** P-002's Statement says Ship "may only claim and implement a task after the harness-architect has confirmed red"; its Enforcement says "Filter ready queue to only tasks carrying `harness-ready`". Rev 3 removes the Step 3 queue filter (moves enforcement to a per-task Step 4.1 check) and permanently leaves the Step 0.5 shipment claim, which activates tasks before any harness exists. The plan/decision say both "no operator waiver asserted" and "not claiming literal compliance"; the reviewer holds these cannot both stand and that only the operator (policy owner) can establish compliance for protected text. Required operator input: (1) a recorded ratification of the O-2 reading covering BOTH the moved Enforcement point (Step 4.1 per-task gate instead of the Step 3 queue filter) AND the pre-harness shipment claim, or an explicit scoped `skip_policy: P-002` for the 026-F bootstrap run; (2) ratification of reading R (an explicitly listed `active` task is not "non-ready") for 026.001-T; (3) acknowledgement that P-003's "parent sub-epic" is satisfied by the established feature->task hierarchy. After that: change "no waiver asserted" to cite the ratification and "ratification not required" to "obtained".

**P1 closable by plan text (Scope Boundary).** The closure inventory (iii) is not mechanically executable: the baseline is prose, `git grep` returns single physical lines while clauses wrap, and "every changed hit belongs to M1-M13" is a judgement. Fix: define the baseline as the same `git grep -n -E ... origin/main -- .github` run, strip prefix and line numbers, key on file + line text; pass rule = every added line text is a substring of a PW clause (whitespace-collapsed) and every removed line text a substring of a Retired clause; any other difference triggers the stop rule; run as a throwaway shell command, not a test.

**P2 queue (apply with the above):** (Correctness/Architecture) state that on the shipment route Step 2 executes only item 1 at its own position and items 2-4 run per task from Step 4.1; I3 "verbatim" -> "semantically retained"; PW-4 -> "run Step 2 items 2-4 for this single task (skip scaffolding if it already carries `harness-ready`)" and re-read status, skip a `done`/archived member; add the skill heading "Claim the ready task set" and line 50 "use all ready tasks under the feature" to Retired; widen inventory pattern (`ready queue|Step 2|Step 3|harness_cmd|just-in-time`); the operator-disposition path for a resumed `active` task whose red cannot be confirmed. (Security) apply the Step 4.1 gate on EVERY route and define the shipment route by what happened (Step 0.5 claim ran, or a task reaches 4.1 already `active`); PW-4 gains "Ship MUST NOT apply or edit `harness-ready` itself; valid only if harness-architect reported Compilation PASS and Red Phase CONFIRMED for this task ID in this session; skill unavailable/error = FAIL-CLOSED HALT"; Step 4.2 gains "do not begin implementation of a task for which the Step 4.1 P-002 check has not passed in this iteration"; marker emitted only for clause mismatches (file/path errors use `t.Fatalf`). (Maintainability) consider `t.Fatalf` after the first green run (panic only for the red proof) or pin 1-2 distinctive fragments per clause; keep only 3-4 defect-bearing Retired phrases permanent; state the Step 3 item 2 insertion more briefly ("membership list only; `harness-ready` enforced per task at Step 4.1"); keep the term "ready queue" consistently; give Step 4.1 the trigger and Step 2 item 3 "invoked from Step 4.1"; confirm the legacy route exists or delete it; optionally merge 026.003-T and 026.004-T. (Scope) merge the four helpers into two to stay under the function limit; closure AC (i) -> "a subset of the three files, `workflow-policies.md` absent"; drop unverifiable readiness statement (b) keeping the SHA. (Decision doc) update to four tasks, M1-M18, three edited files; delete the branch-locus and third-route claims and the `D6` citation; fix the "19(18 in view)" baseline typo.

<!-- plan-review-attempt: 3 --> FAIL (cap reached: 3 review-fix cycles); Stage halted; operator input required (above)