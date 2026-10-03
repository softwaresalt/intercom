---
title: "Stage memory — 049-F rev 7 round-5 plan-review (D-049-9), FAIL (hold kept)"
date: 2026-10-03
agent: stage
feature: 049-F
shipment: 039-S
mode: dark (P-017)
status: hold-kept
base_commit: 121e972
---

# Stage memory: 049-F rev 7 round 5 (D-049-9)

## Outcome

* At 2026-10-03T00:13-07:00 the operator authorised **exactly one** more
  plan-review round, round 5, over plan rev 7 Unit E as amended by
  R4-1..R4-8. It is recorded as **D-049-9**, the second exception to the
    3-cycle cap for the Unit E rev 7 gate after D-049-8 (the third overall,
    after D-030-6). It does not authorise a round 6.
* **The round-5 gate result is FAIL. The 049-F STAGE HOLD stays.**

  | Reviewer | Model | Verdict | Findings |
  |---|---|---|---|
  | Anchor | gpt-5.6-sol (high) | FAIL | GPT5-1 P1, GPT5-2 P1 |
  | Second | claude-opus-5.5 | ADVISORY | OP5-1 P2, OP5-2..OP5-4 P3 |
  | Third | gemini-3.8-flash | PASS | none (no grok substitution needed) |

  All three reported REVIEWED_HEAD `121e972dbabf20d9b6bbf27b9cff6694cf04ebc6`.
* **Stage checked each P1 mechanically.** It used a temporary in-package
  `go/parser` probe test, plus a golangci-lint experiment made as a local
  edit. Both were reverted, so the tracked `tools/` tree is unchanged.
  * **GPT5-1 (P1) was ACCEPTED, and OP5-1 (P2) reports the same defect.**
    R4-2 did not say that the predicate-port rows are built from unmasked
    source. If you probe the existing masked `tc.masked` values, `:464`,
    `:465` and `:466` also fail to parse, and R4-2 would then move the
    STRING/CHAR rejection pins to the syntax-error test. Disposition: R5-1.
  * **GPT5-2 (P1) was REJECTED as a false positive.** Stage re-pointed the
    E-T2 callers on the real package. golangci-lint 2.13.2, run with the repo
    config, reported `scanText - relPath always receives "x.go" (unparam)`.
    With R4-3 applied, it reported `0 issues`. The reviewer's synthetic scratch
    package did not reproduce this behaviour. Its secondary point, the
    Constitution row VI count "E-T2 4", was accepted at P3. Disposition: R5-4.
* The P3 findings were dispositioned as follows:
  * R5-2 (OP5-2): the AC-D2.1 port row.
  * R5-3 (OP5-4): import housekeeping.
  * OP5-3: D-049-9 is recorded in the plan (changelog item 12 and the
    "Round 5" section) and in the deliberation (§4.9 and the §5 row).
* **Amendments R5-1..R5-4 are applied but are NOT re-gated (they are un-gated).**
  They touch:
  * the plan: E-T3, AC-E3.2, AC-E3.3, the Constitution row VI, changelog
    item 12, and the "Round 5" review section with
    `<!-- plan-review-attempt: 5 -->`;
  * the deliberation;
  * the backlog: the new `round5` section and the ACs of 049.003-T, and the
    `hold` sections of 049-F and 039-S.
* **Escalation is ESCALATION_DEGRADED**, because there is no engram handoff
  surface. The decision goes to the operator.
* **No stash entries were created.** Every finding was on the same contract
  surface (P-021 C1).

## State at handoff

* **049-F is `blocked`.** It was not moved and `--force-gates` was not used.
* **039-S is `queued` and on hold** (do NOT claim). Its manifest is 049-F plus
  049.001-T..049.007-T, and no change was needed. It depends on 031-S, which
  has shipped.
* `backlogit sync` and `backlogit doctor` are clean.
* This memory is committed in the staging PR from branch
  `chore/stage-049f-rev7-round5-regate`, before that PR merges. The PR number,
  the Copilot iterations and the merge SHA are therefore not known here. Stage
  reports them in its return to the Orchestrator once the merge completes.

## Lessons

* Round 5 had the fewest verified P1s of any rev 7 round: one. It is the same
  defect class as before, row-level parseability in the E-T3 port. This time
  the cause was masked versus unmasked input, not the row text.
* Reviewer lint claims must be reproduced on the **real package**. A
  synthetic scratch module gave a false "0 issues" for `unparam`. A local edit
  that is then reverted, run with the repo config, settles the question.
* Gemini's PASS verdicts have twice missed issues that Stage verified. In
  round 5 it also mislabelled the R4-x amendments. Its PASS should be treated
  as weak evidence.

## Next step

Operator disposition for 049-F/039-S is required (HALT). No round 6 is
authorised. The alternatives recorded in the plan's round-4 and round-5
pattern notes each need a new, recorded operator decision.
