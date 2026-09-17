---
title: "017-S closure plan — review attempt 5 findings"
date: 2026-09-16
plan: docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md
contract: docs/plans/2026-09-16-017-S-closure-canonical-contract.md
reviewing_sha: 5004c84
dispatch_mode: multi-agent
reviewers: correctness, constitution, scope-boundary
decision: FAIL
---

# Attempt 5 — revision 6 (canonical-contract remaster)

Full-plan multi-persona review of the remastered plan **and** the canonical contract. All three
personas returned **FAIL**. Seven distinct P0 findings. **Stage independently verified every P0
against the workspace; none was a reviewer misreading.**

## Persona verdicts

| Persona | Verdict | P0 | P1 | P2 |
|---|---|---|---|---|
| Correctness | FAIL | 3 | 4 | 8 |
| Constitution | FAIL | 3 | — | — |
| Scope boundary | FAIL | 1 | 5 | 8 |

## P0 findings, with Stage's independent verification

| ID | Finding | Stage verification | Status |
|---|---|---|---|
| C-P0-1 / SB-05 | PF-2 demands a reviewing-commit SHA and literal `dispatch_mode:` / `decision: PASS` markers the review ledger cannot contain | Ledger L614 columns were `Attempt/Revision/dispatch_mode/Reviewers/decision` — no SHA column; values are table cells, not literals | **CONFIRMED** |
| C-P0-2 | S-10 backlog allowlist names only `.backlogit/queue/`, but `018.008-T → done` relocates the record to `.backlogit/archive/` | `.backlogit/registry.yaml` routes `done` → `path: archive`. Halt lands **post-claim** | **CONFIRMED** |
| C-P0-3 / F-01 | No allowlist dispositions the artifacts the plan's own mandated calls emit | `.backlogit/reconcile/` is **not** gitignored (`git check-ignore` exit 1) and already carries committed reports for 006-S, 008-S, 009-S, 012-S, 021-S. S-2/S-16/S-18/S-20 all emit there | **CONFIRMED** |
| F-02 | R-5 names no execution site for a revert targeting `origin/main` | R-3 and R-4 both pin "on the closure branch"; R-5 (L417) pins nothing. P-010 L238 forbids Ship "Commit or push directly to `main`" | **CONFIRMED** |
| F-03 | The P-002/P-004 harness chain is entirely absent | P-002 precondition is literal ("the task carries the `harness-ready` label"); `harness-architect` skill **is** installed; plan + contract contain **zero** mentions of either | **CONFIRMED** |
| SB-01 | §10 / D-M3 forbids amending P-010, but S-8's mandated work **is** a P-010 amendment | `018.008-T` SURFACE 1 is `.github/policies/workflow-policies.md (P-010)`; plan L464 forbids "Amending P-015, **P-010**, …" | **CONFIRMED** |
| SB-02 / SB-03 | Allowlist is self-certifying (`{impl_paths}` defined as whatever S-8 changed), and `{harness_paths}` binding via bare `git status` ingests the claim's own backlog writes | Both mechanisms read as described; neither enforces the "no production implementation code" invariant | **CONFIRMED** |

## Convergence across personas

Independent convergence is the strongest signal here:

- **Allowlist unsatisfiability** — all three personas (C-P0-2, C-P0-3, F-01, SB-02, SB-03).
- **PF-2 unsatisfiable** — correctness + scope (C-P0-1, SB-05).
- **AC-16 contradicts §8.2** — correctness + scope (P1-5, SB-12).
- **`{post_paths}` / R-6 consumer error** — correctness + scope (P2-9, SB-13).
- **§9 condition 9 mis-anchored to PF-7** — correctness + scope (P2-8, SB-10).
- **§8.2 trigger list diverges from contract D-J8** — correctness + scope (P2-14c, SB-06).

## What the remaster did fix

Recorded so the next revision does not regress it:

- Variable binding order is sound; every variable is bound before use.
- `{post_paths}` is no longer a precondition of its own producer.
- The §8.1 halt table has a row for every step with a defined regime.
- `{merge_sha}` is correctly excluded as a rollback target.
- R-3 faithfully reproduces prerequisite §10.3.2 (ref-not-commit target).
- R-5's two-parent check precedes `-m 1`.
- PF-7.1/7.2 required tokens match the committed probe artifacts exactly.
- PF-2 is **not** circular.
- No duplicate or contradictory rule IDs.
- The `019.007-T` carve-out is now coherent across D-H2/D-H3/S-9/S-10/AC-18.
- The `_ship.agent.md` L832/L861-869 contradiction is correctly dispositioned as stash `11B75632`
  and is **not** smuggled into `017-S` — explicitly confirmed by the scope auditor.
- Manifest integrity is **clean**: the scope auditor traced every path including all §8.1 rows and
  R-1…R-8 and found nothing that adds, removes, or substitutes a member.

## Diagnosis

The remaster achieved its objective — **internal consistency** — and the verified-correct list
above is substantial and real. Review has now moved to a layer that no attempt has systematically
examined: **the external interface.**

Five of seven P0s are collisions with reality outside the plan — tool routing behavior,
tool-emitted artifacts, an installed policy obligation, and the recorded scope of the very task the
plan mandates implementing. None is an internal-logic defect.

## Why the next step is not "patch these seven"

The patch-then-review loop has failed five consecutive times. Decisively: **one of this attempt's
findings was introduced by Stage's own fix in the immediately preceding turn.** After declaring
revision 6 mechanically verified, Stage widened §8.2's trigger list to name S-19/S-21/S-22 without
updating contract D-J8 — and two personas independently flagged exactly that divergence. The defect
class reproduced itself inside the very turn that claimed to have eliminated it.

**Prescribed next step — a systematic external-interface measurement pass:**

1. Enumerate every external surface the plan touches: each backlogit operation's effect on record
   location, every artifact any mandated tool call emits, every installed policy obligation the
   route triggers, and the recorded scope of every item the plan mutates.
2. Measure each empirically with read-only probes.
3. Record each as an `M-` row in the canonical contract.
4. Only then regenerate the plan from the contract.

## Design-level finding

Three of the seven P0s are instances of a single design choice: the **exhaustive allowlist**
(`anything else ⇒ HALT`). Such a rule must perfectly predict every path a multi-tool chain touches,
and its failure mode is catastrophic — a post-claim halt with no unclaim operation, stranding
`017-S` `active` in the P-001 single-active slot.

Consider inverting it: a **deny-list of forbidden surfaces** plus a positive "no production
implementation code" assertion. That formulation is robust to unanticipated tool emissions, which
is precisely the class of failure that has dominated this review lineage.

## Scope captured (P-021)

- **`DB12DA37`** — P-002's `harness-ready` ordering is structurally unsatisfiable on any
  shipment-claim route, because the shipment claim auto-activates descendant tasks before any
  harness can exist. General to the workspace, not specific to `017-S`; resolving it would require
  amending P-002 or changing claim semantics, either of which exceeds this shipment's frozen scope.
- **`11B75632`** (captured earlier, re-confirmed correct this attempt) — `_ship.agent.md`
  cascade-authority contradiction.
