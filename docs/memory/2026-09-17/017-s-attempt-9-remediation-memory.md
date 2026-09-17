---
title: "017-S plan-review attempt 9 — FAIL, revision 13 remediation"
description: "Stage session memory: attempt-9 confirmation review of revision 12 failed on 2 P1s; remediated contract-first into revision 13 and published to PR #57."
status: active
tags: [stage, 017-S, plan-review, remediation]
date: 2026-09-17
agent: stage
shipment: 017-S
feature: 018-F
pr: 57
---

# 017-S — plan-review attempt 9 and revision 13 remediation

## Outcome

Operator-authorized **plan-review attempt 9** (current-HEAD confirmation review of
revision 12 at `8472ba7`) returned **FAIL**. The operator's conditional instruction to
record `decision: PASS` applied only if there were no P0/P1 findings; two P1s were
raised, so **no PASS was recorded**. The fallback instruction — remediate
same-contract-surface findings — was executed instead.

| Persona | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|
| Correctness | FAIL | 0 | 2 | 3 | 9 |
| Constitution | ADVISORY | 0 | 0 | 7 | 7 |
| Scope boundary | ADVISORY | 0 | 0 | 3 | 4 |

This was **review-fix cycle 1 of a maximum 3**.

## The P1

Post-claim `C3` halts (arising only at S-8/S-10) were routed to Stage in **four**
locations. Contract `D-J11` is categorical: every halt at or after S-3 routes to the
**operator**. Stage cannot dispose of a shipment at all (P-010). The personas found
three locations; measurement found the fourth.

Aggravating factor: the plan's own precedence rule made the **wrong** contract routing
authoritative over the correctly-remediated plan step, so fixing the plan alone would
have left the defect governing.

## Five attempt-8 remediations regressed

This is what the confirmation review existed to catch:

1. #9 (§13 constitution rows) introduced a **false P5 citation** and over-broad
   Principle IX / VII claims.
2. #10 **over-corrected and deleted the P-013.6 escalation route** that contract
   `D-J13` requires the plan to name.
3. #8, #5, #3 each fixed the clause the finding named and failed to propagate.

## Dominant defect class — three-direction taxonomy (measured, settled)

Roughly 30 findings across attempts share one shape: a remediation validated against the
finding it answers but never propagated to other clauses referencing the changed
mechanism. Attempt 9 confirmed it runs in **three** directions:

1. plan fixed, contract not (`M-13` to `D-J4`)
2. contract fixed, plan not (`M-1` to plan section 6.1 `P2`)
3. one section fixed, **sibling sections in the same artifact** not
   (section 4.2 fixed; `S-7` and `D-D7` not)

**Mechanism-reference closure must be run in all three directions after every changed
mechanism.** This is the single highest-value process fix found this session.

## Settled measurement — the contested P-004 question

P-004's own literal precondition requires `go test ./...` non-zero; `018.008-T` AC-13
supplies the scoped-run literals. **Both** are required. Section 4.2 was correct; `S-7`
and `D-D7` were the incomplete surfaces. The scoped run is what pins the red phase to
this function rather than unrelated breakage.

## Changes

- Contract `2026-09-16-017-S-closure-canonical-contract.md`: 266 to 277 lines, 12 edits,
  6 new decision rows (`D-A7`, `D-D8`, `D-D9`, `D-D10`, `D-G6`, `D-J14`).
- Plan `2026-09-16-intercom-go-017-S-closure-plan.md`: 999 to 1066 lines, revision 13,
  about 25 edits, 3 new acceptance criteria (`AC-32`, `AC-33`, `AC-34`).
- New evidence `evidence/2026-09-16-017-S-closure/attempt-9-review-findings.md`.
- `.backlogit/queue/017-S.md` ground 3 rewritten.

Remediation was performed **contract-first**, per the plan's own precedence rule.

## Validation

- markdownlint-cli2 on plan, contract and evidence: **0 issues**
- `AC-1` through `AC-34` contiguous, no gaps; all `D-` families contiguous
- Mechanism-reference closure run in all three directions; zero residual
  stale routing or stale cardinality phrases
- `backlogit sync` and `backlogit doctor`: no issues, 245 artifacts
- `017-S` manifest re-verified at exactly **13** members
- `017-S` to `021-S` (blocks) satisfied — `021-S` is `archived`
- `018-F` is `queued` with **no `parent_id`**, independently confirming the new
  `PF-4` root-ness gate (`D-D9`) is satisfiable

## Scope

Every finding was in-scope under **P-021 C1** — each is a defect in an artifact Stage is
authoring. **No P-021 capture was required** and `017-S` was not widened. The three
standing deferrals (`DB12DA37`, `11B75632`, `3F546E63`) remain out of scope.

## State

- Commit `2abcbbe` pushed to `stage/017-S-closure-plan`; PR #57 `OPEN` / `MERGEABLE` /
  `CLEAN`; `ci gate` SUCCESS; Go jobs correctly SKIPPED (docs-only).
- P-018: Copilot never engaged (no reviews, requests or threads), so the gate is
  not applicable.
- Checkpoint `checkpoint-20260916-193114.json` **resolved** earlier this session after
  Engram returned and the mandatory bounded prune-on-restore completed.
  `checkpoint-20260913-194412.json` and `checkpoint-20260913-084125.json` remain
  **active and untouched**.

## Next steps

1. **Operator authorizes plan-review attempt 10 against revision 13.** Stage does not
   self-authorize — the P-013 circuit is open and operator-acknowledged.
2. On a non-FAIL verdict, operator authorizes the PR #57 merge (P-014 / P-009), landing
   the plan on `origin/main`.
3. Only then does `017-S` eligibility ground 3 discharge and the shipment become
   claimable by Ship.
