---
title: "Implementation plan: Ship harness selection conformance repair (026-F, DB12DA37), re-plan rev 9 (PR-review fixes)"
date: 2026-10-17
status: draft
source: docs/decisions/2026-10-17-intercom-go-026f-o1-closure-and-surface-adjudication-decision.md
supersedes: docs/plans/2026-09-18-intercom-go-harness-selection-conformance-plan.md (rev 1, FAIL); rev 2 of this file (git history, FAIL x2)
feature: 026-F
tasks: [026.001-T, 026.002-T, 026.003-T, 026.004-T]
---

# Implementation plan - 026-F re-plan (rev 9)

**Operator ratifications (2026-10-17, recorded in the decision note section 0):** (1) the O-2 reading of P-002 is ratified, no `skip_policy` needed; (2) an explicitly listed `active` task is not "non-ready" (reading R); (3) P-003 "parent sub-epic" is satisfied by the feature->task hierarchy. Rev 4 resolves the attempt-3 Constitution P1 on that basis and applies the attempt-3 Scope P1 and the queued P2s; it is validated by a ratification-resolution pass on the changed text only.

## Problem Frame

`backlogit shipment claim` (Ship Step 0.5 item 4) auto-transitions manifest tasks `queued -> active` before Ship Step 2. Step 2 item 1 (`_ship.agent.md:322`) lists only `queued` tasks, which is then empty; item 4 (`:329`) passes vacuously; Step 3 item 2 (`:362-367`) replaces the P-002 Enforcement filter ("`harness-ready` AND `queued`") with Step 3 item 1's derived executable set, discarding BOTH predicates. Net: Step 4 can execute a task that was never harnessed (P-004 red-phase bypass; DB12DA37). `workflow-policies.md:50` already requires the `harness-ready` filter, so this is a conformance repair of Ship, not a policy amendment.

A latent defect must be handled by the repair itself: a naive repair makes Step 2 scaffold the WHOLE manifest batch up front. For a multi-task shipment with test-bearing tasks that leaves several red harnesses at once and Step 4.3's full `go test ./...` gate fails after the first task (the G1 deadlock; masked on the shipment route only because Step 2 was vacuous). With the one-task rule withdrawn, the repair makes harness generation **just-in-time per task** on the shipment route.

Operator authority (2026-10-17): O-1 closed, Option B (substring/structural tests are an admissible acceptance surface); one-task rule superseded. O-1..O-4 and C1-C8: see the source decision.

## Design (the repaired contract on the shipment route)

1. **Route discriminator (C7).** The *shipment route* is any run in which a `shipment_id` is recorded as session scope (Step 0.5 primary item 5 or fallback item 4) OR in which a task reaches Step 4.1 already `active`; on any ambiguity the run is treated as the shipment route. Any other run is the *legacy route* (backlogit without shipment support, where Step 0.5 does not apply) and keeps today's rules semantically retained. The Step 4.1 P-002 gate applies on EVERY route (on the legacy route it is redundant with the Step 3 item 2 filter, which stays).
2. **One derivation (C1/C2).** Step 3 item 1's derived executable set (`queued` or `active` task members, manifest-filtered) is the single read-only derivation, evaluated FIRST and once by Step 2 on the shipment route and reused by Step 3. It is a MEMBERSHIP list, not an execution boundary.
3. **Just-in-time per-task harness (G1).** On the shipment route Step 2 item 3 runs for ONE task at a time, in dependency order, immediately before that task's Step 4.1, invoking `harness-architect` with that task's ID passed explicitly as `tasks` (never the whole batch); Ship proceeds only after the skill confirms Compilation PASS, Red Phase CONFIRMED and the `harness-ready` label is applied. At most one task's red harness exists at any Step 4.3 gate.
4. **Per-task P-002 gate at Step 4.1.** Before a task is claimed/confirmed `active` and implemented, Ship verifies it carries `harness-ready`, applied only by `harness-architect`; a task lacking it is a FAIL-CLOSED HALT (never a silent skip, never a proceed) citing P-002's Violation Action; implementation (Step 4.2) never begins for a task whose gate has not passed in this iteration. This is the single execution gate on the shipment route. The Step 0.5 shipment claim is a backlog side effect and is not the P-002 gate.
5. **harness-architect** treats an explicitly listed `active` task as in scope and no longer excludes it as "non-ready".

## Constitution Check

* **P-002 (O-2 reading, RATIFIED by the operator 2026-10-17; decision section 0).** Not amended; no `skip_policy` is needed or used. The operator ratified that per-task enforcement at Step 4.1 is the gate on the Ship act and that the Step 0.5 shipment claim activating tasks before a harness exists is not the gated act. P-002's Gate Point is "Queue building (Step 2) and task claiming (Step 3)"; the Ship act gated is selecting a task for implementation (Step 4.1 Claim Task and Step 4.2). The plan makes the Precondition (`harness-ready`) and Enforcement hold for each task immediately before it is claimed and implemented. The reading was Stage's interpretation of installed text; it now carries operator authority.
* **P-004.** Each task's harness is produced by `harness-architect`, which confirms red and only then applies the label (its postcondition). Red for THIS feature's tests is recognised exactly as the skill's Step 5.2 defines it (see "Red marker").
* **P-003.** Parent feature 026-F; every task has parent, >=1 AC, size and complexity (complexity as prose; the task type has no structured field). Tasks parent directly to the feature: the operator ratified (2026-10-17) that P-003's "parent sub-epic" is satisfied by the feature->task hierarchy of this workspace.
* **P-006** hardening below. **P-010** Stage writes no source. **P-016** no extra worktrees.
* **Protected, byte-unchanged:** P-002 and P-004 text, Step 4.3 text, Step 0.5 item 4 claim text.
* **IV-5 / 022.001-T.** The behavioural-fixture preference is NOT applied to 026-F only, by O-1 Option B; it stays in force elsewhere. The needle tests prove the contract WORDING, not Ship behaviour; the behavioural check is the Runtime Verification run.

## Red marker (how the substring failure is recognised as red, no policy change)

`harness-architect/SKILL.md` Step 5.2 requires each harness test to fail with the marker `panic("not implemented: <reason>")`. The 026 tests therefore assert through two local helpers in the test file: `readSurface(t, relpath, startAnchor, endAnchor)` (reads a fixed-literal repo path, slices by anchors, collapses whitespace; any read/path/anchor error uses `t.Fatalf`, NEVER the marker, so a typo is not a fake red) and `pin(t, subject, clause, id string, wantPresent bool)` whose single failure line is `panic(fmt.Sprintf("not implemented: pinned wording %s: %s", "absent"|"still present", id))`. Only a clause present/absent mismatch emits the marker, exactly the skill's, so Step 5.2 recognises it and Step 6 may apply the label. The panic is a deliberate exception for wording-contract tests only (a later drift panics the integration package loudly and names the clause id; accepted). Because a panic aborts the test binary, the red-phase check runs EACH new test function separately (`go test ./tests/integration -run '^TestName$'`) and records the marker per function; the skill's whole-suite run is scoped this way to "the harness tests", as its own Step 5.2 text says. There is no production stub for a contract harness; that is recorded as "not applicable: contract text harness" in the skill's Step 4. The skill (not Ship) authors and confirms each task's red test file as that task's first commit. Each invocation note records the skill's Step 4 "production stubs" and Completion-Criteria "structural stubs" items as not applicable to a contract-text harness (no production code exists to stub) and instructs the skill to use the per-function `-run` filter; the exit code and marker per function go in the readiness block.

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
| M7 | Consumer preamble | `:337` | "Now that all tasks are harnessed" | "Now that the batch membership is identified (Step 2 item 1), construct the execution queue:" | 003 | T3b |
| M8 | Consumer derivation | `:338-361` item 1 (`:358-360`) | "the actual task-membership boundary that Step 4 executes"; trigger "Stage-prepared shipment" | membership boundary; trigger = shipment route; single-derivation note (note text lands in 002) | 003 (boundary, trigger), 002 (note) | T3a, T2a |
| M9 | Consumer queue | `:362-367` item 2 | substitution drops `harness-ready`; trigger "When operating under a Stage-prepared shipment (item 1 above ran)" | trigger = shipment route; membership list in dependency order; P-002 check applied per task at Step 4.1 | 003 | T3a |
| M10 | Consumer empty queue | `:369` item 4 | "empty after harness generation" | "(legacy route) after harness generation or (shipment route) after membership derivation" | 003 | T3b |
| M11 | Consumer broadcast | `:371` | "count of queued items" | "count of ready-queue tasks" (`:968` row reads `{count} tasks`, unchanged) | 003 | T3b |
| M12 | Per-task gate | `:377-379` Step 4.1 and the start of Step 4.2 (`:~422`) | status -> active; build-feature delegation | add the per-task P-002 gate before the claim (every route) and a Step 4.2 guard sentence | 004 | T4a |
| M13 | Skill selection | `harness-architect/SKILL.md:12,18,38,50,54,56,60` | "ready" scoping; "otherwise non-ready" | default `queued`; explicit `active` in scope; exclude blocked/done/archived | 001 | T1a |
| M14 | Authority grant | `workflow-policies.md` P-002 (36-52), P-004 (84-92) | mandates filter | **PROTECTED** | - | git diff |
| M15 | Anti-regression | `_ship.agent.md` Step 4.3 (435-447), Step 0.5 item 4 claim | gate; claim | **PROTECTED** | - | git diff -U0 |
| M16 | Unchanged, noted | `harness-architect/SKILL.md` Step 5.2 (`:124-127`) | "Run `go test ./...` for the harness tests" | already scoped to "the harness tests"; unchanged; scoping per Red marker | - | git diff |
| M17 | Unchanged, noted | `_ship.agent.md:1052` checkpoint trigger "harness generation completes" | fires after harness generation | unchanged; on the shipment route it fires per JIT harness, which is intended | - | git diff |
| M18 | Unchanged, consistent | `build-feature/SKILL.md:12`; `_ship.agent.md:23,431`; `backlog-integration.instructions.md:136`; `_stage.agent.md:363`; `prompts/stage-grouping-analysis.prompt.md:22` | `harness-ready` label use / queued-item wording unrelated to Ship selection | verified, no change | - | git diff |

**Baseline inventory (plan time, `git grep -n -E "<pattern>" -- .github`, pattern in Closure rules C-4; lines that map to no M-row, e.g. `_ship.agent.md` 4, 23, 89 and `_stage.agent.md` 106, 363, are unrelated prose and are expected in both baseline and final unchanged):** hits exist only at the sites listed in M1-M18 above (`_ship.agent.md` 4, 23, 89, 318, 324, 325, 329, 337, 358, 362, 369, 371, 431, 1052; `_stage.agent.md` 106, 363; `backlog-integration` 136; `stage-grouping-analysis.prompt` 17, 22; `brainstorm`/`runtime-verification`/`build-feature:29` "up front" prose unrelated; `build-feature` 12; `harness-architect` 12, 18, 43, 50, 56, 60, 132, 136, 156). `.github/agents/_orchestrator.agent.md` has no hit.

**Stop rule.** If a finding reappears at an adjacent surface, Ship halts, records a DEFERRED/BLOCKED note on the task, and writes NO plan or matrix artifact (C8; Ship's P-010 boundary); Stage rebuilds the matrix.

## Pinned wording (needle source; whole clauses; matched after whitespace collapse)

`readSurface` collapses all whitespace runs to a single space (`strings.Fields` joined by " ") so wrapped source lines match; every clause below is written single-spaced.

* **PW-1 (M8/003)**: replace `it is the actual task-membership boundary that Step 4 executes, so an \`active\` member of this derived set is never omitted from the ready queue,` with `it is the task-MEMBERSHIP boundary (which tasks are candidates), not the execution boundary: the single execution gate is the per-task P-002 check at Step 4.1, so an \`active\` member of this derived set is never omitted from the ready queue,`.
* **PW-2 (M9/003)**: item 2 trigger `When operating under a Stage-prepared shipment (item 1 above ran), replace` becomes `When on the shipment route (the Step 2 item 1 route discriminator; item 1 above ran), replace`; and after `regardless of whether its status is \`queued\` or \`active\`,` insert `in dependency order; on the shipment route this list is a membership list only (\`harness-ready\` is enforced per task at Step 4.1),`. Item 1's trigger `(when operating under a Stage-prepared shipment; C1–C6,` (line 339) is changed to `(when on the shipment route; C1–C6,`; add that old trigger phrase to the Retired list.
* **PW-3 (M7/M10/M11/003)**: `Now that the batch membership is identified (Step 2 item 1), construct the execution queue:`; item 4 `If the queue is empty after harness generation (legacy route) or after membership derivation (shipment route), halt and report — there is nothing to build.`; broadcast `with the count of ready-queue tasks`.
* **PW-4 (M12/004)**: new first paragraph in Step 4.1: `**P-002 per-task gate**: before claiming this task, re-read its status (if \`done\` or archived, report the skip with the task ID rather than claim; if no task in the batch ends up executed, treat it as an empty batch: halt and report), then run Step 2 items 2-4 for this single task (skip scaffolding only if \`harness-architect\` has reported Compilation PASS and Red Phase CONFIRMED for this task ID in this session; a label without that report is not sufficient: re-invoke the skill for that task ID, and if red cannot be confirmed, HALT) and verify the task carries the \`harness-ready\` label. Ship MUST NOT apply or edit \`harness-ready\` itself: the label is valid only if \`harness-architect\` reported Compilation PASS and Red Phase CONFIRMED for this task ID in this session; if the skill is unavailable or errors, that is a FAIL-CLOSED HALT. A task lacking the label is a FAIL-CLOSED HALT, never a silent skip and never a proceed, citing P-002's Violation Action ("Halt and suggest running the harness-architect"); for a resumed \`active\` task whose red phase cannot be confirmed, Ship halts, records a DEFERRED/BLOCKED note on the task and leaves the disposition to the operator. The Step 0.5 shipment claim has already set manifest tasks \`active\` as a backlog side effect; it is not the P-002 gate.` and, at the start of Step 4.2: `Do not begin implementation of a task for which the Step 4.1 P-002 check has not passed in this iteration.`
* **PW-5 (M1-M6/002)**: (a) preamble `Ensure every task in the selected batch (item 1) has a passing test harness before that task's implementation begins. On the shipment route this step runs just-in-time per task, invoked from Step 4.1 (item 3): at this position execute item 1 only (halt if the batch is empty) and items 2-4 are executed per task from Step 4.1; on any other route it runs once, up front, not in a loop.`; (b) item 1 `Select the batch. **Route discriminator**: the *shipment route* is any run in which a \`shipment_id\` is recorded as the session scope (Step 0.5 primary item 5 or fallback item 4) or in which a task reaches Step 4.1 already \`active\`, and any ambiguity is treated as the shipment route; on that route the Step 0.5 claim has already activated manifest tasks. On the shipment route the batch is the Step 3 item 1 derived executable set, admitting both \`queued\` and \`active\` members; Ship evaluates that single read-only derivation FIRST, once, and Step 3 reuses it. On any other route, list all tasks for the target feature or chore that are in \`queued\` status.`; (c) item 2 `Partition the selected batch (on the shipment route: the current task)`; (d) item 3 adds `On the shipment route, this item is invoked from Step 4.1: scaffold ONE task at a time in dependency order, immediately before that task's Step 4.1, passing that task's ID explicitly as \`tasks\` (a whole-batch up-front scaffold would leave several red harnesses at once and break the Step 4.3 full-suite gate after the first task); do not proceed until Compilation PASS, Red Phase CONFIRMED and the \`harness-ready\` label are recorded. On any other route, invoke the skill for the batch as before.`; (e) item 4 `On the shipment route confirm the current task now carries the \`harness-ready\` label; on any other route confirm every task in the batch does.` ... `An empty selected batch on the shipment route is a halt, never a vacuous pass.`; (f) the line-320 broadcast `[SHIP] Invoking harness-architect skill` is qualified `when invoking the skill (per task from Step 4.1 on the shipment route)`; (f2) closing paragraph gains `On the shipment route the batch comes from the Step 3 item 1 derivation, never from a queued-only queue operation.`; (g) Step 3 item 1 gains `(This is the single derivation Step 2 item 1 evaluates first on the shipment route; it is not recomputed.)`.
* **PW-6 (M13/001)**: description `defaults to all \`queued\` tasks`; line 18 `ready work items` -> `in-scope work items`; broadcast `Tasks in scope`; line 50 `When omitted, use all \`queued\` tasks under the feature. An \`active\` task is in scope only when it is listed explicitly in \`${input:tasks}\`.`; Step 1 heading `Select the in-scope task set`; item 1 `Load the feature or chore and its in-scope descendants (\`queued\`, plus explicitly listed \`active\`)`; item 3 `Exclude blocked, done, and archived work items. An explicitly listed \`active\` task is in scope and MUST NOT be excluded as non-ready.`

Retired (ABSENT after the owning task): the skill heading `Claim the ready task set`; `When omitted, use all ready tasks under the feature.`; item 1's old trigger `(when operating under a Stage-prepared shipment; C1–C6,`; `every task in the target feature or chore`; `This step runs once, up front — not in a loop.`; `confirm every queued task now carries`; `Now that all tasks are harnessed`; `it is the actual task-membership boundary that Step 4 executes`; `When operating under a Stage-prepared shipment (item 1 above ran)`; `with the count of queued items`; (skill) `defaults to all ready tasks`; `ready work items`; `Exclude blocked, done, or otherwise non-ready work items`. **Permanent vs red-proof only:** only the four defect-bearing phrases `confirm every queued task now carries`, `This step runs once, up front — not in a loop.`, `it is the actual task-membership boundary that Step 4 executes` and `Exclude blocked, done, or otherwise non-ready work items` remain permanent `pin(..., wantPresent=false)` assertions; every other retired phrase is checked in the red run only and then dropped from the test. Needles are whole clauses by design (drift must be loud; operator Option B).

## Authorized replacement pairs (the span-pair rule's table; self-checked against every mandated edit PW-1..PW-6)

Pair semantics (the table cells follow these exactly): for an **in-span** pair OLD is a quoted substring of one baseline span and NEW is the text that REPLACES that substring (for an addition inside a span, OLD is the anchor substring and NEW is that anchor substring followed by the added text, so the anchor text legitimately appears in both cells); one in-span pair may consist of several non-overlapping edits inside the SAME span (AP-14). For an **insertion** pair (AP-8, AP-9 only) OLD names the unchanged left-anchor HEADING (it stays outside the change group) and NEW is the INSERTED TEXT ONLY, never including the heading. NEW is whitespace-collapsed. Each pair is owned by one task (column Task); the per-task checks that use these pairs are defined ONLY in Closure rules C-4.

| Pair | Task | File | OLD | NEW |
|---|---|---|---|---|
| AP-S1 | 001 | skill | `defaults to all ready tasks` | `defaults to all \`queued\` tasks` |
| AP-S2 | 001 | skill | `a feature's or chore's ready work items` | `a feature's or chore's in-scope work items` |
| AP-S3 | 001 | skill | `Ready tasks: {task_count}` | `Tasks in scope: {task_count}` |
| AP-S4 | 001 | skill | `When omitted, use all ready tasks under the feature.` | PW-6 line-50 sentence |
| AP-S5 | 001 | skill | `### Step 1: Claim the ready task set` | `### Step 1: Select the in-scope task set` |
| AP-S6 | 001 | skill | `its ready descendants through backlog` | `its in-scope descendants (\`queued\`, plus explicitly listed \`active\`) through backlog` (removes the last `ready descendants` match; no ADDED partner is expected) |
| AP-S7 | 001 | skill | `Exclude blocked, done, or otherwise non-ready work items.` | PW-6 item-3 sentences |
| AP-10 | 002 | ship | `Ensure every task in the target feature or chore has a passing test harness before any implementation begins. This step runs once, up front — not in a loop.` | PW-5(a) preamble (removes `any`, `runs once`, `up front`; authorized as a whole span) |
| AP-11 | 002 | ship | Step 2 item 1 current single sentence (`List all tasks for the target feature or chore that are in \`queued\` status.`) | PW-5(b) item 1 |
| AP-12 | 002 | ship | `Partition the task list:` | PW-5(c) `Partition the selected batch (on the shipment route: the current task)` |
| AP-13 | 002 | ship | `If any tasks need harnesses, invoke the **harness-architect** skill for the batch.` | same sentence + PW-5(d) addition |
| AP-14 | 002 | ship | `confirm every queued task now carries the \`harness-ready\` label.` (edit 1) and the end of that same list item (edit 2, anchor `halt and report the gap rather than proceeding with a partial set.`) | PW-5(e): edit 1 = the route-qualified confirmation sentence; edit 2 = anchor + `An empty selected batch on the shipment route is a halt, never a vacuous pass.` |
| AP-15 | 002 | ship | closing paragraph anchor `prefer the queue operation to assemble the task set.` | anchor + PW-5(f2) sentence |
| AP-16 | 002 | ship | `broadcast \`[SHIP] Invoking harness-architect skill\` before invoking the skill.` | PW-5(f) qualified broadcast |
| AP-17 | 002 | ship | end of Step 3 item 1: `task elsewhere happens to share its label or status.` | anchor + PW-5(g) note |
| AP-1 | 003 | ship | `it is the actual task-membership boundary that Step 4 executes, so an \`active\` member of this derived set is never omitted from the ready queue,` | PW-1 |
| AP-2 | 003 | ship | item 1 trigger `(when operating under a Stage-prepared shipment; C1–C6,` | `(when on the shipment route; C1–C6,` |
| AP-3 | 003 | ship | item 2 trigger `When operating under a Stage-prepared shipment (item 1 above ran), replace` | PW-2 trigger |
| AP-4 | 003 | ship | anchor `regardless of whether its status is \`queued\` or \`active\`,` | anchor + PW-2 insertion |
| AP-5 | 003 | ship | `Now that all tasks are harnessed, construct the execution queue:` | PW-3 preamble |
| AP-6 | 003 | ship | `If the queue is empty after harness generation, halt and report — there is nothing to build.` | PW-3 item 4 |
| AP-7 | 003 | ship | `with the count of queued items` | `with the count of ready-queue tasks` |
| AP-8 | 004 | ship | left-anchor heading `#### Step 4.1: Claim Task` (insertion pair; heading unchanged) | PW-4 gate paragraph (inserted text only) |
| AP-9 | 004 | ship | left-anchor heading `#### Step 4.2: Delegate to Build Feature` (insertion pair; heading unchanged) | the Step 4.2 guard sentence (inserted text only) |

Self-check result (Stage, rev 5): every mandated edit PW-1..PW-6 is covered by exactly one pair (PW-1 AP-1; PW-2 AP-3/AP-4; item-1 trigger AP-2; PW-3 AP-5/6/7; PW-4 AP-8/9; PW-5 a-g AP-10..AP-17; PW-6 AP-S1..S7). The two removals that leave no PAT residue (AP-S6, AP-10) are authorized spans, so they cannot trip the rule.

Walkthrough against Closure rules C-4 (rev 9; verification record, not a rule): in-span pairs AP-S1..S7, AP-10..AP-17 and AP-1..AP-7 each live inside one baseline span (list item, paragraph, heading or bullet) and are validated as part of that span's change group; AP-S1..S7 (skill), AP-10 (Step 2 preamble paragraph), AP-11/12/13 and the two-edit AP-14 (Step 2 list items, one span each), AP-15 (closing paragraph), AP-16 (broadcast paragraph), AP-17/AP-1/AP-2 (shared Step 3 item 1 span), AP-3/AP-4 (Step 3 item 2 span), AP-5 (Step 3 preamble paragraph), AP-6 (Step 3 item 4), AP-7 (broadcast paragraph) pass with the group-text rule. The insertion pairs AP-8 (after `#### Step 4.1: Claim Task`) and AP-9 (after `#### Step 4.2: Delegate to Build Feature`) pass because their left-anchor headings stay anchors (NEW is inserted text only, so the heading is not duplicated in the expected group text) and NEW may occupy one or more final spans (or be prepended to the next span) at the start of the following group; the unchanged paragraph that follows each heading remains an anchor. Boundary cases: no pair has a virtual anchor as its left anchor and no pair edits the first or last span of a file, so the prefix and suffix groups must be empty at every task end (any appended or prepended text stops the run); there are no deletion-only pairs and no pair deletes text at a file end. The unapplied check (rule 3) works per pair for both kinds, including pairs sharing a span with completed ones.

## Closure rules (normative - the ONLY copy of these rules)

Evaluated at the END of each task against the CUMULATIVE state through that task (`git diff origin/main...HEAD`). Task AC-H3 clauses, the readiness evidence, the pair table and the matrix refer to this section by pointer and restate only task-specific parameters.

**C-1 Changed paths.** Classify every changed path (and its status). Any path matching neither allowlist, and any deletion or rename of a pre-existing file, STOPS the run; `.github/policies/workflow-policies.md` must be absent (I2).
* **Allowlist A (implementation):** `^(\.github/agents/_ship\.agent\.md|\.github/skills/harness-architect/SKILL\.md|tests/integration/harness_selection_conformance_test\.go)$`; at the end of task N only files owned by tasks <= N may differ (001: skill + test file; 002-004: `_ship.agent.md` + test additions).
* **Allowlist B (bounded backlog/evidence artifacts the Ship/harness-architect/build-feature flow legitimately writes; harness-architect Step 6 persists labels and notes in task records, build-feature stages all changes, Ship Step 4.5 writes completion, commit association and memory checkpoints):** `^\.backlogit/queue/(026-F|026\.00[1-4]-T|051-S)\.md$`; `^\.backlogit/logs/(026-F|026\.00[1-4]-T|051-S)\.jsonl$` (append-only); `^\.backlogit/(checkpoints|reconcile)/[^/]+$`; `^\.backlogit/stash\.jsonl$` (append-only); ADDED memory notes only, matching `^docs/memory/\d{4}-\d{2}-\d{2}/[^/]+\.md$` (repo convention `{YYYY-MM-DD}/{slug}-memory.md`, also the circuit-break and per-task checkpoint forms), `^docs/memory/\d{4}-\d{2}-\d{2}-[^/]+\.md$` (existing flat form) or `^docs/memory/compacted/[^/]+\.md$` (compact-context output); pre-existing memory files may not be modified, deleted or moved. Archive moves are not part of the PR.

**C-2 B-record content.** A started or completed record (026-F, tasks <= the current task, 051-S) may change ONLY: `status`; `labels` (e.g. `harness-ready`); `custom_fields.harness_*`; the tool-managed `updated_at` and `commit` fields; body comment/evidence sections. For 051-S only `status`, `updated_at`, `commit`. Never `id`, `parent_id`, `dependencies`, `title`, `artifact_type`, `created_at`, or shipment `items`. **Not-yet-started tasks** (tasks later in the chain than the current one): the ONLY permitted change is the claim-induced `queued -> active` status transition (Ship Step 0.5 activates every manifest task at the shipment claim) with its tool-managed `updated_at`; any label, harness field, commit or body change (premature harness or implementation) STOPS the run. No other task/feature/shipment record (none of 020*, 025*, 019*, 050-S) may differ.

**C-3 Protected text (I1).** `git diff origin/main...HEAD -U0 -- .github/agents/_ship.agent.md` has hunk ranges disjoint from Step 0.5 item 4 and Step 4.3.

**C-4 Mechanical inventory (a throwaway script, not a test), anchor-aligned.** For each instruction surface in allowlist A: baseline = `git show origin/main:<file>`, final = the working file; whitespace-collapse; split into bounded spans (one heading, one paragraph, or one list item with its wrapped lines; nested list items are separate spans). **Alignment is by stable anchors, never by position:** compute a longest-common-subsequence of the two span lists using EXACT equality of the collapsed span text (headings match by their text), and add a **virtual start anchor and a virtual end anchor** (the document boundaries) to both lists. Spans in the LCS are *anchors* and must be identical. Everything between two consecutive anchors, INCLUDING the group between the virtual start anchor and the first real anchor (prefix) and the group between the last real anchor and the virtual end anchor (suffix), is a **change group**, so text added, removed or changed at the very beginning or end of a file receives the same authorization check as any other edit; an appended or prepended instruction that no cumulative pair authorizes STOPS the run: the baseline group (0 or more baseline spans) and the final group (0 or more final spans). Group text = its spans' collapsed texts joined by a single space, so how many final spans an authorized edit occupies is irrelevant (an insertion may become one or several consecutive final spans). Empty groups are normal: an insertion pair applies to an EMPTY baseline group (the final group then holds exactly NEW).

Pair kinds (every pair of the table is exactly one kind): **in-span** (OLD is a substring of one baseline span, replaced by NEW or extended by an addition at the stated anchor; this is every pair except AP-8 and AP-9) and **insertion** (AP-8, AP-9: OLD is an unchanged anchor HEADING; NEW is inserted immediately after it, i.e. at the START of the change group that begins right after that anchor, whether NEW becomes new final spans or is prepended to the following baseline span). The table contains no deletion-only pairs; deletions occur only inside in-span pairs (e.g. AP-10 removes `any`, `runs once`, `up front`).

The **cumulative pair set** at the end of task N is the union of the Authorized pairs owned by tasks <= N: **026.001-T** = AP-S1..AP-S7; **026.002-T** = those + AP-10..AP-17; **026.003-T** = those + AP-1..AP-7; **026.004-T (final)** = those + AP-8, AP-9 (all pairs; required ONLY here).
1. *Each change group:* expected group text = the baseline group text with the cumulative pairs applied (in-span pairs whose OLD lies in the group; insertion pairs whose left-anchor heading is the real anchor immediately before the group, applied at the group start; the virtual anchors can never be a pair's left anchor, so prefix/suffix groups are expected to be empty). The final group text must equal the expected text. A group to which no cumulative pair applies must be empty on both sides; any other difference STOPS the run. This also covers spans shared by several pairs (e.g. Step 3 item 1 with AP-17 from 002 and AP-1/AP-2 from 003): the shared span's final text must equal the baseline span with exactly the cumulative pairs applied.
2. *Each cumulative pair individually:* consumed exactly once (its OLD occurs exactly once in the baseline; its NEW is present in the final text at that place).
3. *Each later-task pair individually (unapplied check):* for an in-span pair, its OLD is still present unchanged and its NEW is absent from the final text; for an insertion pair, its left-anchor heading is still an anchor and its NEW text is absent from the final text. A span shared with completed-task replacements may differ from baseline by those replacements only; it is NOT required to equal baseline.
4. *PAT report:* `PAT` = `harness-ready|ready tasks|ready descendants|ready work items|non-ready|every task in the target|all tasks are harnessed|every queued task|queued status|queued items|task-membership|up front|before any implementation|runs once|harness generation|ready queue|Step 2|Step 3|harness_cmd|just-in-time`; `git grep -n -E "$PAT" origin/main -- .github` versus the working tree, restricted to files NOT in allowlist A, must be identical. Inside allowlist-A files PAT residue is reported, never decisive (change groups are validated whole by 1-3).
Ship records, per task: the cumulative set, consumed pairs, still-unapplied later pairs, the C-1/C-2 classification and the PAT report.

## Implementation Units

**Common protocol (every task).** Strictly sequential 001 -> 002 -> 003 -> 004; order resolves the bootstrap (the skill fix lands first) and avoids forward references (Step 3 and Step 4.1 cite Step 2 text that already exists). Posture TEST-FIRST, just-in-time: the harness for task N is authored only when task N starts, by `harness-architect` invoked with that task's ID passed explicitly as `tasks` BEFORE Step 4.1 (claim) and Step 4.2; Compilation PASS and Red Phase CONFIRMED per function (Red marker section) are recorded and `harness-ready` is applied by the skill; Step 4.1 does not start otherwise. The red test commit is never pushed red; the task END commit has the full `go test ./...` green. **Bootstrap for 026.001-T:** the shipment claim has already set it `active` under the unrepaired skill, whose Step 1 item 2 says an explicit `tasks` set restricts the scope and whose item 3 excludes "otherwise non-ready" work; reading R = an explicitly listed `active` task is not "non-ready" (it is ready and in progress). If the skill refuses 026.001-T, Ship halts fail-closed and reports; it never bypasses. Needles are whole clauses from PW-1..PW-6 sliced by their own step/list anchors via `readSurface`; every assertion resolves in its own subject.

**Readiness-block evidence (every task).** (a) per-function red marker output then green output (exact commands); (b) the session-start commit SHA of the instruction surfaces; (c) the Closure rules C-1..C-4 results; (d) the annotation: "DB12DA37 claim-ordering condition disclosed; per-task harness-architect gate and red-then-green evidence recorded; O-2 reading of P-002 ratified by the operator 2026-10-17 (decision section 0); no skip_policy used".

**Closure ACs.** Defined ONLY in the section "Closure rules (normative)" above; every readiness block records the results of rules C-1..C-4 at the end of the task. No rule is restated anywhere else.

### U1 - 026.001-T: harness-architect scope (Size S, complexity medium)

Files (2): `.github/skills/harness-architect/SKILL.md` (M13), `tests/integration/harness_selection_conformance_test.go` (NEW; two local helpers `readSurface` and `pin`; reuses `repoRoot` and `countOccurrences` already in package `integration`). Test (1): **T1a** `TestHarnessArchitect_ExplicitActiveTaskInScope` (PW-6 clauses present incl. line 50 and item 3; retired clauses absent, in that file only). ACs: AC-1 PW-6 clauses present; AC-2 retired clauses absent; AC-3 Closure rules C-1..C-4 (normative section); AC-4 full suite green; plus common protocol.

### U2 - 026.002-T: Ship Step 2 producer (Size M, complexity medium)

Files (2): `_ship.agent.md` (M1-M6 and the Step 3 item 1 note), the test file (append). Tests (2): **T2a** `TestStep2_PreambleDiscriminatorAndSingleDerivation` (PW-5 a, b, g; retired preamble and old "every task in the target feature or chore" absent; the FULL legacy sentence retained); **T2b** `TestStep2_JustInTimeTaskPartitionGuardAndClosingParagraph` (PW-5 c, d, e, f; retired "confirm every queued task now carries" absent). ACs mirror U1.

### U3 - 026.003-T: Ship Step 3 consumer (Size M, complexity medium)

Files (2): `_ship.agent.md` (M7-M11), the test file (append). Tests (2): **T3a** `TestStep3Items1And2_MembershipBoundaryRouteTriggerDependencyOrder` (PW-1, PW-2; retired clauses absent within the item 1/2 slices); **T3b** `TestStep3PreambleEmptyQueueBroadcast` (PW-3; retired preamble and "with the count of queued items" absent). ACs mirror U1.

### U4 - 026.004-T: Ship Step 4.1 per-task P-002 gate and closure (Size S, complexity medium)

Files (2): `_ship.agent.md` (M12), the test file (append). Test (1): **T4a** `TestStep41_PerTaskP002GateFailClosed` (PW-4 gate paragraph inside the Step 4.1 slice, ordered before the "Update task status to `active`" sentence, including the no-self-labelling and skill-error halt clauses, and the Step 4.2 guard sentence inside the Step 4.2 slice). This task also records the FINAL closure check (Closure rules C-1..C-4). ACs mirror U1.

## Dependency Graph

`026.001-T -> 026.002-T -> 026.003-T -> 026.004-T` strictly sequential (blocks edges). 026-F is the covering feature. 025-F/019-F chain is NOT planned here; 020-F/050-S stay held until 026-F ships.

## Decisions and Rationale

* JIT per-task harness on the shipment route is part of the repair: without it the repair re-creates the G1 deadlock for every multi-task shipment.
* The execution gate sits at Step 4.1 per task because Step 3 runs once before any JIT harness exists; Step 3 item 2 becomes membership only (one derivation, one gate: C1/C2).
* The panic-marker helpers make substring tests recognisable as red by the skill's own Step 5.2 without touching any policy.
* Order skill -> Step 2 -> Step 3 -> Step 4.1: bootstrap works and no task cites wording defined later.
* Four tasks, each <=2 files, <=2 test functions; no single-task constraint.
* IV-5 stays in force elsewhere (O-1 Option B is scoped to 026-F).

## Risks and Caveats

* **DB12DA37 bootstrap.** 026-F's own run uses the unrepaired Ship (Step 2 vacuous). The per-task `harness-architect` invocation before Step 4.1 provides P-002 Precondition and P-004 postcondition at implementation. Residual: Ship may skip the protocol; the readiness evidence is merge-gating for the operator.
* **Panic markers** abort the integration test binary on a drifted clause (loud, names the clause id).
* **Reading R for 026.001-T** is RATIFIED by the operator (explicitly listed `active` task is not non-ready); a refusal by the skill still halts safely.
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

**Unresolved operator decisions.** None. O-2, reading R and the P-003 sub-epic reading were ratified by the operator on 2026-10-17 (decision section 0).

<!-- plan-review-attempt: 2 --> prior FAIL (rev 2); rev 3 applies the full remediation queue; attempt 3 pending

## Plan Review - attempt 3 of 3 (rev 3): FAIL; cap reached, Stage halted

(Historical record. Any closure-mechanics proposal quoted in this and the next section is superseded by the section `Closure rules (normative)`.)

Six genuinely separate reviewer subagents were dispatched (Constitution, Correctness, Architecture, Scope Boundary, Security, Maintainability); each reported `dispatch_mode: single-agent` internally because reviewer subagents cannot nest, so no reviewer had a deeper sub-dispatch. `decision: FAIL` (aggregate; one persona-level P1 per FAIL below). Verdicts: Constitution **FAIL**, Scope Boundary **FAIL**, Correctness ADVISORY, Architecture ADVISORY, Security ADVISORY, Maintainability ADVISORY. Rev-2 findings A (bootstrap order), B (red marker), C (`:318` sentence) were confirmed CLOSED by rev 3.

**P1 needing OPERATOR authority (Constitution).** P-002's Statement says Ship "may only claim and implement a task after the harness-architect has confirmed red"; its Enforcement says "Filter ready queue to only tasks carrying `harness-ready`". Rev 3 removes the Step 3 queue filter (moves enforcement to a per-task Step 4.1 check) and permanently leaves the Step 0.5 shipment claim, which activates tasks before any harness exists. The plan/decision say both "no operator waiver asserted" and "not claiming literal compliance"; the reviewer holds these cannot both stand and that only the operator (policy owner) can establish compliance for protected text. Required operator input: (1) a recorded ratification of the O-2 reading covering BOTH the moved Enforcement point (Step 4.1 per-task gate instead of the Step 3 queue filter) AND the pre-harness shipment claim, or an explicit scoped `skip_policy: P-002` for the 026-F bootstrap run; (2) ratification of reading R (an explicitly listed `active` task is not "non-ready") for 026.001-T; (3) acknowledgement that P-003's "parent sub-epic" is satisfied by the established feature->task hierarchy. After that: change "no waiver asserted" to cite the ratification and "ratification not required" to "obtained".

**P1 closable by plan text (Scope Boundary).** The closure inventory (iii) is not mechanically executable: the baseline is prose, `git grep` returns single physical lines while clauses wrap, and "every changed hit belongs to M1-M13" is a judgement. Fix: define the baseline as the same `git grep -n -E ... origin/main -- .github` run, strip prefix and line numbers, key on file + line text; pass rule = every added line text is a substring of a PW clause (whitespace-collapsed) and every removed line text a substring of a Retired clause; any other difference triggers the stop rule; run as a throwaway shell command, not a test.

**P2 queue (apply with the above):** (Correctness/Architecture) state that on the shipment route Step 2 executes only item 1 at its own position and items 2-4 run per task from Step 4.1; I3 "verbatim" -> "semantically retained"; PW-4 -> "run Step 2 items 2-4 for this single task (skip scaffolding if it already carries `harness-ready`)" and re-read status, skip a `done`/archived member; add the skill heading "Claim the ready task set" and line 50 "use all ready tasks under the feature" to Retired; widen inventory pattern (`ready queue|Step 2|Step 3|harness_cmd|just-in-time`); the operator-disposition path for a resumed `active` task whose red cannot be confirmed. (Security) apply the Step 4.1 gate on EVERY route and define the shipment route by what happened (Step 0.5 claim ran, or a task reaches 4.1 already `active`); PW-4 gains "Ship MUST NOT apply or edit `harness-ready` itself; valid only if harness-architect reported Compilation PASS and Red Phase CONFIRMED for this task ID in this session; skill unavailable/error = FAIL-CLOSED HALT"; Step 4.2 gains "do not begin implementation of a task for which the Step 4.1 P-002 check has not passed in this iteration"; marker emitted only for clause mismatches (file/path errors use `t.Fatalf`). (Maintainability) consider `t.Fatalf` after the first green run (panic only for the red proof) or pin 1-2 distinctive fragments per clause; keep only 3-4 defect-bearing Retired phrases permanent; state the Step 3 item 2 insertion more briefly ("membership list only; `harness-ready` enforced per task at Step 4.1"); keep the term "ready queue" consistently; give Step 4.1 the trigger and Step 2 item 3 "invoked from Step 4.1"; confirm the legacy route exists or delete it; optionally merge 026.003-T and 026.004-T. (Scope) merge the four helpers into two to stay under the function limit; closure AC (i) -> "a subset of the three files, `workflow-policies.md` absent"; drop unverifiable readiness statement (b) keeping the SHA. (Decision doc) update to four tasks, M1-M18, three edited files; delete the branch-locus and third-route claims and the `D6` citation; fix the "19(18 in view)" baseline typo.

<!-- plan-review-attempt: 3 --> FAIL (cap reached: 3 review-fix cycles); operator input was required (above)

## Ratification-resolution pass (2026-10-17, rev 4)

The operator ratified O-2, reading R and the P-003 hierarchy reading (decision section 0), which resolves the attempt-3 Constitution P1; rev 4 applied the Scope P1 (mechanical inventory) and the queued P2s. Four independent reviewer subagents (Constitution, Scope Boundary, Correctness, Security) validated the changed text only: all ADVISORY, NO new P0/P1 (reviewer dispatch mode: single-agent read-only each, as nesting is unavailable). Advisory items applied after the pass: PW-4 no longer permits skipping scaffolding on a label alone (must be backed by a harness-architect PASS/CONFIRMED report for the task ID in this session); done/archived skip is reported and an all-skipped batch is an empty-batch halt; inventory compares mid-line edits by word-diff pairs and as multisets; the Step 2 broadcast `[SHIP] Invoking harness-architect skill` fires only when the skill is actually invoked (per task from Step 4.1 on the shipment route). Verdict: ADVISORY (PASS-with-advisory); proceed to harvest.

<!-- plan-review: ratification-resolution ADVISORY, rev 4; no new P0/P1 -->

## Plan Review - final (ratification-resolution, rev 5)

dispatch_mode: single-agent-declared-degradation

decision: ADVISORY

Selected-persona coverage: every selected persona was covered by its own separate reviewer subagent in the ratification-resolution pass (Constitution, Scope Boundary, Correctness, Security), validating the changed text of rev 4; the prior attempt-3 pass covered Constitution, Correctness, Architecture, Scope Boundary, Security and Maintainability. Each reviewer ran single-agent read-only because reviewer subagents cannot nest; that is the declared degradation. No P0 or P1 remains open: the attempt-3 Constitution P1 is resolved by operator ratification, the attempt-3 Scope P1 by the mechanical inventory (re-specified in rev 5 after PR review as bounded span pairs with an authorized-pair table), and the PR-review P1s on closure AC (i) (changed-path allowlists) and on the inventory rule are resolved in this revision.

Explicit authorization to harvest under ADVISORY: the operator ratified (2026-10-17, decision note section 0) the O-2 reading of P-002, reading R and the P-003 feature->task hierarchy, and directed the Stage ratification-resolution pass; harvest of 026.001-T..026.004-T and shipment 051-S proceeded on that authority.

Rev 6 to rev 9 (PR 151 rounds 2-5) changed only the closure mechanics, which now live in ONE normative section (Closure rules (normative)); the duplicated obsolete copy was deleted and every other place refers to that section by pointer. No new design issue arose; no new review cycle was run for these edits.
