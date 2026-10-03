---
title: "Stage memory — 049-F go/ast write-path spike, rev 7 re-plan, gate FAIL (hold kept)"
date: 2026-10-02
agent: stage
feature: 049-F
shipment: 039-S
mode: dark (P-017)
status: hold-kept
base_commit: 961b652
---

# Stage memory: 049-F go/ast spike and rev 7 re-plan

## Outcome

* I executed stash 1EEBECA5 (operator decision D-049-1) once its trigger fired
  (031-S merged at `961b652`).
* **The spike is complete.** The findings are in
  `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`.
  * G-1..G-7 are all feasible.
  * G-3 shows **exact verdict parity**, with 0 divergences across 19 fixtures,
    3 filebased inputs, 17 tracked files and 155 `.go` files.
* **Plan revision 7** re-plans Unit E as a linear chain:
  E-T2 → E-T3 → E-T4 → E-T5 → E-T7 → E-T6.
  * E-T7 is the new task 049.007-T. It folds in stash 458F9385 (residual
    item 6 widening, 50 selectors).
  * E-T6 narrows items 2 and 6 and discloses residual item 8.
* **The plan-review gate FAILed in all 3 cycles:**

  | Cycle | gpt-5.6-sol | claude-opus-5.5 | gemini-3.8-flash / grok-4.7 |
  |---|---|---|---|
  | 1 | FAIL | ADVISORY | gemini ADVISORY |
  | 2 | FAIL | FAIL | gemini ADVISORY |
  | 3 | FAIL | FAIL | gemini empty → grok FAIL |

  Each cycle surfaced new P1s, most of them in E-T3's test-port contract. The
  cycle-3 fixes are applied but **not re-gated**.
* **The hold is kept (D-049-7).** 049-F stays `blocked` and 039-S stays on hold,
  so DARK_MODE_HALTED for 039-S.
  * The escalation route is ESCALATION_DEGRADED (engram is unavailable), which
    falls back to the operator halt.

## Backlog changes

* New task 049.007-T, with `blocks` edges 049.004→049.003, 049.007→049.005 and
  049.006→049.007. 039-S lists 049.007-T.
* New `hold` section on 049-F.
* Rev 7 replan sections on 049.001-T..049.007-T, plus the post-cycle-3
  amendments on 049.003-T (ACs, replan) and 049.005-T (GK3-4).
* Stash created:
  * FE2F02FF (residual item 8);
  * C0D28448 (item-6 remainder).
* Stash archived:
  * 1EEBECA5 (the spike doc exists);
  * 458F9385 (folded into 049.007-T).

## Next steps (operator)

* Authorise one more plan-review round over rev 7, as D-030-6 did for Unit A,
  or choose another disposition.
* Lift 049-F `blocked → queued` only on ADVISORY or better with no P0/P1, then
  run the 039-S `pre_claim` topology gate.

## Failed approaches

* gemini-3.8-flash returned an empty response in cycle 3. I used grok-4.7 as
  the fallback.

## Adversarial diff review (pre-PR)

- Reviewers: gpt-5.6-sol (GD-1..5), claude-opus-5.5 (OD-1..8) and gemini-3.8-flash (GM-1..2). All three returned ADVISORY with no P0/P1.
- Fixes applied in the follow-up commit:
  - Inventory scope rescoped (GD-1).
  - AC-E3.3 grep extended with the `extent*` constants and `type extent` (GD-2); confirmed to match today's code as a positive control.
  - Spike-output pointer fixed (GD-3).
  - Archived stash disposition pointers added (GD-4/OD-4).
  - Narrow-not-close title note added (GD-5).
  - 039-S `hold` section added (OD-1).
  - Broken cross-reference fixed (OD-2/GM-1).
  - P-021 (B) corrected for FE2F02FF/C0D28448 (OD-3).
  - `unix.Chmod` (OD-5).
  - `captured_stash` frontmatter updated (OD-6/GM-2).
  - Forced worktree removal note added (OD-8).
- OD-7 was already superseded by the 049.003-T cycle-2 amendment.
- 039-S claim hold: 049-F `blocked` is mechanically enforced. Ship Step 0.5 runs `shipment-reconcile` (pre, `expected_status: queued`), so the blocked 049-F is a `status-mismatch` that halts intake (Copilot PR #99). The `hold` section is explanatory.

## Staging PR #99 Copilot review

- Iteration 1 (on `e5e2d7f`): two findings, fixed in `8978726`. The spike harness now fails on divergence, and the 039-S hold text names the mechanical guard.
- Iteration 2 (on `8978726`): two inline findings and three previously missed items. All were classified as the same contract surface (P-021 C1) and applied, not re-gated:
  - E-T5 / 049.005-T binding covers `token.ASSIGN` (`x = pathsafe.NewRoot(...)`).
  - AC-E7.3 / 049.007-T adds an import-path scan; an aliased or dot import is a HALT.
  - The harness filebased decode, read and golden paths, plus the fixture golden read and the all-module walk, now fail instead of logging. Re-run at `961b652`: the same tallies, PASS.
  - H-5 evidence pointer corrected. The commands and verbatim output (H-5 1–4, the inclusive variant, the import-path scan and the shape scan) are in `docs/decisions/assets/2026-10-02-writepath-go-ast-spike/h5-rev7-remeasure.txt`. This memory doc holds no verbatim output.
- Iteration 3 (on `484dae9`): one inline finding and one previously missed item, both applied:
  - E-T5 / 049.005-T binding uses the multi-name `var x, err = pathsafe.NewRoot(...)` form, since a single-name `var x = ...` does not compile. The AC-E5.1 fixture adds a third caller.
  - ALLGO 155 is described as the 154 tracked `.go` files at `961b652` plus the untracked harness.
- The hold is unchanged. These amendments join the operator-authorized re-gate scope.