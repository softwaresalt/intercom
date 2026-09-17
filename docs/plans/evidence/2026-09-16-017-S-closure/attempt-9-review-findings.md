---
title: "Plan review attempt 9 — 017-S closure plan revision 12"
date: 2026-09-17
status: complete
agent: Stage
shipment: 017-S
attempt: 9
revision_reviewed: 12
reviewing_sha: 8472ba7
dispatch_mode: multi-agent
decision: FAIL
---

# Plan review attempt 9 — revision 12 @ `8472ba7`

Operator authorization: `AUTHORIZE PLAN REVIEW ATTEMPT 9 FOR 017-S — current-HEAD confirmation
review of revision 12`.

Scope: full multi-persona current-HEAD confirmation review of the canonical contract, the
remastered plan, the `017-S` eligibility wording and the PR #57 readiness block — specifically to
confirm the attempt-8 advisory remediations introduced no contradictions.

## Verdicts

| Persona | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|
| Correctness | **FAIL** | 0 | **2** | 3 | 9 |
| Constitution | ADVISORY | 0 | 0 | 7 | 7 |
| Scope boundary | ADVISORY | 0 | 0 | 3 | 4 |

**Aggregate decision: FAIL.** The operator's `decision: PASS` condition ("if there are no P0/P1
findings") was **not met**, so no PASS was recorded. The FAIL is driven solely by the P1
halt-routing contradiction below.

The confirmation review achieved its stated purpose: **five of the attempt-8 remediations were
measured to have introduced new defects or left propagation gaps**, which is exactly the class of
regression a confirmation review exists to catch.

## P1 — post-claim halts routed to Stage (HIGH, all three personas)

A `C3` finding can only arise during **S-8** or **S-10**, both of which are post-claim. Contract
`D-J11` is categorical: *every halt at or after S-3 routes to the OPERATOR*. Stage cannot dispose
of a shipment at all (P-010). Four locations contradicted this:

| Location | Defect |
|---|---|
| contract `D-J10` | "C3 (defect in this plan ⇒ halt to Stage)" |
| plan `AC-23` | "or C3 (halt to Stage)" — certified the wrong routing as acceptance-passing |
| plan §8.2 post-restore | "the shipment returns to Stage" on the R-6/R-7 **success** path |
| contract `D-K12` | same post-restore defect |

Aggravating factor: the plan's own precedence rule (contract governs) made the **wrong** routing
authoritative over the correctly-remediated plan step S-10.

**Remediated** in all four locations, contract-first. Halts now route to the operator with Stage
re-measurement recorded as advisory input only.

## HIGH consensus — §13 Principle III/IV mis-citation

The row claimed "§6.1 **P5** rejects any changed path outside the repository". P5 is the
*production-code* invariant (no Go / `go.mod` / `go.sum` change outside `{impl_paths}`) asserted by
a repo-scoped diff — it is not the containment predicate. Introduced by attempt-8 remediation #9.

**Remediated** — re-cited to `D1(a)` (repo-scoped diff), `D1(e)` (governance denial), `P1` and §10,
with an explicit note that P5 is not the containment predicate.

## Attempt-8 remediations that regressed

| Remediation | Regression | Disposition |
|---|---|---|
| #9 §13 principle rows | false P5 citation; over-broad IX/VII claims; 7 principles unmapped | fixed; Principles I, II, V, VI, VIII, X, XI added |
| #10 §13 P-013 narrative strip | **over-corrected** — deleted the escalation route `D-J13` requires the plan to name | route restored (`gpt-5.6-sol`/`openai`/`high` + `ESCALATION_DEGRADED` fallback) |
| #8 §4.2 P-004 | not propagated to `S-7` or contract `D-D7` | propagated; "all five" → "all six" literals |
| #5 contract `D-H3` `M-1` | not propagated to plan §3 consumer block or §6.1 `P2` | propagated |
| #3 plan §8.1 `M-13` | not propagated to contract `D-J4` | propagated |

This is the **dominant defect class** for the ninth consecutive attempt: *a remediation validated
against the finding it answers but never propagated to other clauses referencing the changed
mechanism.* Attempt 9 established its measured taxonomy — it runs in **three** directions:
(1) plan fixed, contract not; (2) contract fixed, plan not; (3) one section fixed, sibling sections
in the **same** artifact not. Mechanism-reference closure is now run in all three.

## Safety findings remediated

* **§8.1 S-7 halt row** directed discarding the harness working tree at a **post-claim** halt — an
  unapproved destructive act under Constitution Principle VII that destroys the exact red-phase
  evidence P-004 exists to produce, and contradicted the adjacent row "leave the branch intact as
  evidence". Replaced with an additive commit-and-preserve disposition; generalized as `D-J14`.
* **Principle VIII violated by omission** — no safety mode was declared anywhere. Added `D-A7`:
  S-0 declares `careful` + `freeze-scope` before any mutation (`AC-33`).
* **S-8 had no pass criterion** — "implement to green" was unfalsifiable and asymmetric with the
  five-literal red phase. Added `D-D8` and `AC-32`: five green literals.
* **PF-4 never positively verified `018-F` root-ness**, which is P-015 precondition 1. Added
  `D-D9`.
* **S-9 had no exit contract.** Added `D-D10`, including the `M-17` post-transition re-read path.

## Other remediations

`{run_log}` bound as a real artifact (`D-G6`); the `{harness_paths}` union-cardinality bound now
actually asserted at S-8; `{pre_cascade_sha}` first consumer corrected to S-15 step 5; R-5 /
`D-K7` "contains **both**" → "all three"; P-018 added to S-22; `AC-1` extended to the S-6/S-13
topology gates; `AC-24` scoped to the governance-surface clause it uniquely owns (it duplicated
`AC-18`); PF execution point stated; `markdownlint` given real gate steps (`AC-34`); P-007 archive
postcheck made concrete; PF-2 literal-marker contradiction removed from the status prose.

## Explicitly clean (all personas, re-measured)

Manifest holds at exactly **13** throughout; P-009 merge-commit-only with **no** squash or rebase
limb anywhere including rollback; `PA-017-CASCADE` escrow's nine conditions intact; all three
P-021 deferrals (`DB12DA37`, `11B75632`, `3F546E63`) correctly held out of scope; P-020
non-blocking semantics reproduced rather than inverted; circuit-breaker handling stricter than
installed; all ID families contiguous; no variable consumed before binding.

## Reviewer tooling limitation

All three personas had **read-only file access with no shell**, so none could execute a
measurement command. Every finding requiring command execution was independently re-measured by
the Stage agent before acceptance; the four P1 locations and all five regression claims were
confirmed by direct read before remediation.

## Outcome

Revision 12 → **revision 13**. All findings were in-scope under P-021 C1 (defects in the artifacts
Stage is authoring); **no P-021 capture was required** and `017-S` was not widened. This was
review-fix cycle 1 of a maximum 3.

Revision 13 requires a **separately operator-authorized attempt 10**. Stage does not
self-authorize review attempts — the P-013 circuit is open and operator-acknowledged.
