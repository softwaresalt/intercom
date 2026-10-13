---
title: "Stage memory — 049-F rev 7 round-6 plan-review (D-049-10), FAIL (hold kept)"
date: 2026-10-03
agent: stage
feature: 049-F
shipment: 039-S
mode: dark (P-017)
status: hold-kept
base_commit: 4c36944
---

# Stage memory: 049-F rev 7 round 6 (D-049-10)

## Outcome

* At 2026-10-03T21:10-07:00 the operator wrote, verbatim: "I approve a sixth
  round of review; however, if the findings again identify are all about how
  test inputs are classified and would be caught anyway through test-first
  build, then lift the hold and let the code review be the gate."
  * This is recorded as **D-049-10**.
  * It authorises exactly one round, round 6, over rev 7 Unit E as amended by
    R4-* and R5-*.
  * The conditional hold-lift rule applies to round 6 only. It does not
    authorise a round 7.
* **The round-6 gate result is FAIL. The 049-F STAGE HOLD stays.**

  | Reviewer | Model | Verdict | Findings |
  |---|---|---|---|
  | Anchor | gpt-5.6-sol (high) | FAIL | GPT6-1 P1 |
  | Second | claude-opus-5.5 | ADVISORY | OP6-1 P2, OP6-2..OP6-6 P3 |
  | Third | gemini-3.8-flash | ADVISORY | GM6-1 P2 (no grok substitution needed) |

  All three reported REVIEWED_HEAD `4c369448de039049e70158c9714bbeb0a2f9bb76`.
* **GPT6-1 is verified.** It is the same finding as PRE5-1, OP6-1 and GM6-1.
  * The defect: in the E-T3 predicate table, `Args[3]` and `Args[6]` have no
    non-bare-operand rejection row. `Args[0]` has no STRING/CHAR row, and
    `Args[2]` has no call-valued row.
  * The evidence: a throwaway `go/parser` probe under `logs/` compared the
    full G-4 predicate with a fail-open variant that checks only `Args[0]` and
    `Args[2]`. Over every planned row the two showed **0 divergences**.
  * Conditional rule: (a) classification-only is **NO**. It is a coverage gap
    (missing rows), not the classification of an existing input, and doubt
    counts as not qualifying. The anchor tagged it yes, and Stage's first draft
    said "arguably yes"; Copilot review on PR #103 corrected this. (b) caught
    by test-first is **NO**, because a weak predicate passes every specified
    test.
  * The operator's FAIL list names exactly this case: "a test that could pass
    while asserting the wrong thing" and "AC/requirement correctness a passing
    test would not expose". The anchor's verdict is also FAIL, so the gate is
    **FAIL**.
* **Every classification claim was confirmed by all three reviewers**,
  independently:
  * the parse-error set `{:443, :444, :467, :468}` from unmasked source;
  * masked `:464`–`:466` do not parse;
  * the R4-4 line discrimination;
  * the R4-3 `unparam` boundary;
  * a clean E-T2/E-T3 scratch lint.
* **Amendments R6-1..R6-6 are applied but are NOT re-gated (they are
  un-gated).**
  * The amendments are:
    * R6-1: nine parseable rejection rows, closing PRE5-1;
    * R6-2: rename the test-message strings;
    * R6-3: give the predicate a name that is not on the deletion list;
    * R6-4: add the `writepath_extent_test.go:236` citation;
    * R6-5: remove the `gomask` test import;
    * R6-6: the `retiredarch` pin must not depend on `scanText`.
  * They touch:
    * the plan: E-T2, E-T3, AC-E2.7, AC-E3.2..AC-E3.4, changelog item 13,
      the header outcome, and the "Round 6" section with
      `<!-- plan-review-attempt: 6 -->`;
    * the deliberation: §4.9 and the D-049-10 row;
    * the backlog: the `round6` sections of 049.002-T and 049.003-T, their
      ACs, and the `hold` sections of 049-F and 039-S.
* **Escalation is ESCALATION_DEGRADED**, so the decision goes to the operator.
* **No stash entries were created.** Every finding was on the same contract
  surface (P-021 C1).

## State at handoff

* **049-F is `blocked`.** It was not moved and `--force-gates` was not used.
* **039-S is `queued` and on hold** (do NOT claim). Its manifest is 049-F plus
  049.001-T..049.007-T, and no change was needed. It depends on 031-S, which
  has shipped.
* **039-S is the next shipment.** It must be started by Ship in a **NEW
  session**, and only after a new operator disposition lifts the hold. Stage
  then moves 049-F to `queued` and removes the hold sections.
* `backlogit sync` and `backlogit doctor` are clean.
* This memory is committed in the staging PR from branch
  `chore/stage-049f-rev7-round6-regate`, before that PR merges. The PR number,
  the Copilot iterations and the merge SHA are therefore reported in Stage's
  return to the Orchestrator.

## Lessons

* Round 6 is the first rev 7 round where every classification claim held.
  The remaining P1 is a coverage gap (missing rows), not a mis-classification.
  A coverage gap cannot be caught by a test-first build, because a missing row
  never fails.
* The 0-divergence probe is a quick, mechanical way to tell "classification"
  defects from "coverage" defects. Run both the full predicate and a weakened
  predicate over the planned table. If they agree on every row, the table does
  not pin the requirement.
* Gemini's row labels were wrong again (`:81` for `:80`, and invented
  predicate-row names), although its parse verdicts matched. Treat its tables
  as weak evidence.

## Next step

HALT, pending an operator disposition for 049-F/039-S. No round 7 is
authorised. The options recorded in the plan's Round 6 pattern note are:
* accept R6-1 as written and lift the hold;
* authorise a round 7;
* amend the D-049-7 lift rule.

Each option needs a new, recorded operator decision.
