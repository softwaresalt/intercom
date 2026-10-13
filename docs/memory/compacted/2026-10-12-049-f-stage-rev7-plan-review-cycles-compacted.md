---
title: "049-F Stage rev 7 plan-review cycles compacted memory"
description: "Dense summary of the Stage go/ast spike, the rev 7 re-plan of Unit E, rounds 4 to 6 of the plan-review gate, and the operator hold-lift (D-049-11)."
date: 2026-10-12
compacted_by: "compact-context (P-020, target all)"
feature: 049-F
shipment: 039-S
source_files:
  - docs/archive/memory/stage-049f-go-ast-spike-replan-memory.md
  - docs/archive/memory/stage-049f-rev7-round4-memory.md
  - docs/archive/memory/stage-049f-rev7-round5-memory.md
  - docs/archive/memory/stage-049f-rev7-round6-memory.md
  - docs/archive/memory/stage-049f-rev7-hold-lift-memory.md
---

# 049-F Stage rev 7 plan-review cycles: compacted memory (2026-10-02 to 2026-10-03)

This summary replaces five Stage memory files (about 27 KB together). The verbatim originals are archived under `docs/archive/memory/` with the names listed in `source_files`.

## Spike and re-plan (2026-10-02)

* The go/ast spike (stash `1EEBECA5`, operator decision D-049-1) ran after 031-S merged at `961b652`. The spike doc is `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`.
* G-1 to G-7 were feasible. G-3 showed exact verdict parity: zero divergences across 19 fixtures, 3 file-based inputs, 17 tracked files and 155 `.go` files.
* Plan revision 7 re-planned Unit E as a linear chain: E-T2, E-T3, E-T4, E-T5, E-T7, E-T6. New task 049.007-T folds in stash `458F9385`.
* Pre-PR adversarial review of the spike: three reviewers, all ADVISORY with no P0 or P1. Fixes were applied in a follow-up commit.
* Staging PR #99 Copilot review took three iterations. Fixes included making the spike harness fail on divergence, extending the import-path scan, and correcting the H-5 evidence pointer.

## Plan-review gate results

| Round | Authority | Anchor gpt-5.6-sol | claude-opus-5.5 | gemini-3.8-flash |
|---|---|---|---|---|
| 1 to 3 (standard cap) | FAIL in all three cycles | FAIL | ADVISORY then FAIL | ADVISORY, then empty, so grok-4.7 fallback FAIL |
| 4 (D-049-8) | Operator-authorised exception | FAIL: GPT4-1 P1, GPT4-2 and GPT4-3 P2 | FAIL: OP4-1 to OP4-4 P1 | PASS |
| 5 (D-049-9) | Operator-authorised exception | FAIL: GPT5-1 and GPT5-2 P1 | ADVISORY | PASS |
| 6 (D-049-10) | Conditional operator authorisation | FAIL: GPT6-1 P1 | ADVISORY | ADVISORY |

* Round 4 verified defects: `writepath_extent_test.go:82` does not parse because automatic semicolon insertion follows `os.Remove`. Predicate rows at `:467` and `:468` are parse errors. The AC-D1.3 port did not discriminate, because the mask erased the separators it relied on. Amendments R4-1 to R4-8 were applied without a re-gate.
* Round 5: GPT5-1 was accepted. The predicate-port rows are built from unmasked source, and `:464` to `:466` fail to parse in masked form. GPT5-2 was rejected as a false positive, because a synthetic scratch module gave a false "0 issues" for `unparam`. The real package with the repo config showed the issue only before R4-3, and zero issues after. Amendments R5-1 to R5-4 were applied without a re-gate.
* Round 6: GPT6-1 was verified. The E-T3 predicate table had no non-bare-operand rejection row for `Args[3]` or `Args[6]`, no STRING or CHAR row for `Args[0]`, and no call-valued row for `Args[2]`. A go/parser probe found zero divergences between the full predicate and a fail-open variant. That made it a coverage gap, not a classification defect, so the classification-only exception did not apply. Amendments R6-1 to R6-6 were applied without a re-gate.

## Operator decision and hold-lift (D-049-11)

* On 2026-10-03 the operator accepted R6-1 as written and lifted the hold, without another review round. This was recorded as OPERATOR-LIFT, an override of the D-049-7 hold-lift rule. The plan-review verdict of record stayed round 6 FAIL.
* 049-F moved from `blocked` to `queued` with a direct `backlogit move`. `--force-gates` was not used. 039-S became eligible, with the manifest 049-F plus 049.001-T to 049.007-T.
* Pre-PR review of the decision record: round 1 had three findings, all fixed, including a P1 requiring the AC-E1.2 discharge note. Round 2 had one P3, fixed. Round 3 was clean. Copilot review of PR #104 had four in-scope comments in cycle 1, all fixed.
* Implementation-time obligations recorded for Ship: the discharge note for AC-E1.2 (049.001-T), the R4 to R6 amendments for 049.002-T, 049.003-T and 049.005-T, and a mutation check for the nine positional rejection rows.

## Learnings

* Review cycles were not converging. Since cycle 2, every rev 7 round found new P1s in the E-T2 and E-T3 test-port detail. The mechanical fix is an executor-side `go/parser` probe of every ported row.
* A coverage gap cannot be caught by a test-first build, because a missing row never fails. The 0-divergence probe distinguishes classification defects from coverage defects.
* Reviewer lint claims must be reproduced on the real package, not a synthetic module.
* Gemini PASS verdicts missed issues that Stage verified, and some of its tables had wrong row labels. Treat them as weak evidence.

## Failed approaches

* Plan-review cycles 1 to 4 and rounds 5 and 6 all returned FAIL on the rev 7 gate.
* gemini-3.8-flash returned empty responses. grok-4.7 was the fallback.
* A synthetic scratch module gave a false negative for `unparam`.

## Outcome

049-F was later completed under 039-S. The 039-S closure merged via PR #111 on 2026-10-07. The closure record is `docs/closure/039-S-049-F-post-merge-closure.md`.
