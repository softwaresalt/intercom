---
title: "020-F re-stage decision: one-task rule superseded; just-in-time per-task harness mitigation"
date: 2026-10-17
status: accepted
type: deliberation-outcome
scope: 020-F (020.001-T .. 020.006-T; 020.006-T added at plan-review)
supersedes_in_part: docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md (section 3.1 / section 9 one-task shape, as applied to 020-F)
operator_authority: operator correction (this session) + Orchestrator dark-mode decision (operator AFK)
---

# 020-F re-stage decision

## 1. Subject and question

Can feature 020-F (five sequential test-bearing tasks) be staged as ONE queued
shipment, and under what mitigation, now that the earlier "one task per
feature/shipment" shape is withdrawn?

## 2. Decision note: the one-task rule is SUPERSEDED

The 2026-09-18 harness-execution-model decision (021.001-T, section 3.1
"Option 1 - single-task features and shipments", and the section 9 terminal
outcome built on it) treated a one-task-per-feature/shipment shape as the
required way to avoid the multi-task Ship Step 4.3 green-gate deadlock.

**Operator direction (authoritative):** there is no rule that a feature or
shipment may contain only one task, and no rule that test-bearing tasks must
each be a single-task feature in their own shipment. The only rule is that
shipment size and complexity are reasonable and reasonably managed. That
one-task rule was an unsupported self-imposed constraint. It is superseded and
is **not applied** to 020-F. Nothing in the harness-execution-model decision's
analysis of the four Ship contract surfaces is rewritten; only its
"therefore one task per shipment" conclusion is withdrawn for this unit.

## 3. Re-assessment of the earlier halt reasons for 020-F

| Earlier reason | Status now |
|---|---|
| Step 4.3 full-green deadlock with five test-bearing tasks | Resolved by plan design: just-in-time per-task harness (section 4). Harness for task N is authored only when task N starts; tasks are strictly sequential 001 -> 005, so no later task's red tests exist while an earlier task is gated. |
| B1: no legal pre-claim harness slot | Not required. The mitigation does not need a pre-claim slot: the red harness is the first commit of each task on the shipment's own feature branch, inside the PR. (Orchestrator recorded that 026-F's review refuted B1 on the same basis; the cited `_ship.agent.md` line number does not match the current file, so this decision relies on the in-PR logic, not the citation.) |
| B2: Stage -> main artifact handoff | The existing operator-opened merge-commit staging PR flow (repeatedly used: PRs #114, #121, #125, #126, #127, #128, #139) delivers the shipment record to `main` before Ship runs. No change. |
| DB12DA37: shipment claim auto-activates tasks, Ship Step 2 lists only `queued` tasks, so Step 2 is vacuous and a task can execute unharnessed (P-004 bypass) | Known conformance bug, fix owned by 026-F (separate chain, blocked on operator-only O-1). NOT fixed or authorized-around here. 020-F is made safe independent of Step 2 by task-level acceptance criteria (section 4). If plan-review raises an unresolved P0/P1 on this mitigation, Stage halts. |
| `blocked` status / "NO SHIPMENT" lock | Was the deliberate hold pending 021.001-T. 021.001-T is `done`; the hold's own lift condition (reviewed decision + impl-plan + plan-harden + plan-review, then fresh shipment) is what this session executes. |

## 4. Options considered

* **Option A - one 5-task shipment, strictly sequential, just-in-time per-task
  red harness (CHOSEN).** Cost: Ship's Step 2 up-front batch is vacuous
  (DB12DA37) so correctness rests on per-task acceptance criteria and review
  evidence rather than on Step 2 bookkeeping. Benefit: one coherent release
  unit, one PR, no policy amendment.
* **Option B - split into two shipments (001-003 detector; 004-005 wiring).**
  Rejected: adds a cross-shipment dependency and a second claim for no
  deadlock benefit (the same JIT discipline would still be required inside
  each shipment); total size (five S/M tasks, ~one day) is reasonable for one
  shipment.
* **Option C - wait for 026-F.** Rejected as unnecessary: the mitigation makes
  020-F independently safe; 026-F remains excluded and untouched.

## 5. Mitigation baked into the plan and tasks (binding on Ship)

1. Each task's FIRST commit is its own red harness (authored through
   `harness-architect` semantics: compilable, failing, with a recorded expected
   failure marker, `go vet` clean). Harness N is not written before task N starts.
2. Tasks run strictly sequentially along the dependency chain (see item 6); no task is started while its predecessor is not done.
3. Every task finishes with the full `go test ./...` green (Step 4.3). A
   committed harness must therefore turn green within the same task, and no
   artifact may leave a deliberately red test behind. Tests that depend on a
   later task use an explicit, visible skip tied to the ErrNotImplemented stub
   (002) which the next task (003) removes red-first.
4. Acceptance criteria require OBSERVED red-then-green evidence per task,
   recorded in the task comment and the PR readiness block: the red run's
   command and failure marker, the green run's command and result.
5. No policy text (P-002, P-004, Ship Step 2) is amended, and NO operator `skip_policy` waiver is asserted or claimed. The DB12DA37 claim-ordering condition (shipment claim auto-activates tasks before any harness) is a pre-existing, systemic, disclosed defect affecting every shipment; this decision does not cure it and does not claim literal P-002 compliance on the claim route. Each task's P-005 compliance annotation (DB12DA37 disclosed; no waiver asserted) in the task comment and PR readiness block makes that visible, and the operator can veto by archiving the queued shipment before Ship claims it. Plan-review attempt 2's constitution persona held this position insufficient without an operator-recorded waiver; that dissent is recorded and left for the operator.
6. Chain after plan-review split: 020.001-T -> 020.002-T -> 020.003-T (block parser) -> 020.006-T (P1-P6 rules + live-contract test, new) -> 020.004-T -> 020.005-T. Six tasks, one shipment.

## 6. Authority and escalation

The Orchestrator's dark-mode decision (operator AFK) authorized proceeding on
this basis. Stage halts and reports, rather than deciding, if plan-review
returns an unresolved P0/P1 on the section 5 mitigation.
