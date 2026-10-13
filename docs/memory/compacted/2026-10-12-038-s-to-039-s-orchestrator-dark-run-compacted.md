---
title: "038-S to 039-S Orchestrator dark run compacted memory"
description: "Dense summary of the 2026-10-02 Orchestrator dark-mode run: four shipments closed, 039-S halted on the 049-F hold, and the operator rulings and environment lessons recorded during the run."
date: 2026-10-12
compacted_by: "compact-context (P-020, target all)"
scope: [038-S, 032-S, 030-S, 031-S, 039-S]
source_files:
  - docs/archive/memory/orchestrator-dark-run-038s-to-039s-memory.md
---

# 038-S to 039-S Orchestrator dark run: compacted memory (2026-10-02)

This summary replaces `orchestrator-dark-run-038s-to-039s-memory.md` (5,023 bytes). The verbatim original is archived at `docs/archive/memory/orchestrator-dark-run-038s-to-039s-memory.md`.

## Outcome at record time

The run was `DARK_MODE_HALTED`. Four of five scoped shipments shipped and closed. 039-S was held.

| Shipment | Feature PR (merge) | Closure PR (merge) | Closure status |
|---|---|---|---|
| 038-S / 048-F | #91 (`0fc58ce`) | #92 (`38f1d92`) | READY |
| 032-S / 035-F | #93 (`532cf52`) | #94 (`698e2bc`) | READY_WITH_CONDITIONS |
| 030-S / 033-F | #95 (`e1d61a7`) | #96 (`8a517bb`) | READY_WITH_CONDITIONS |
| 031-S / 034-F | #97 (`c04d975`) | #98 (`961b652`) | READY |
| 039-S / 049-F | not claimed | not applicable | STAGE HOLD kept |

* Staging PR #90 (`7fad096`) added the `038-S` to `037-S` blocks edge, which fixed `UNSEQUENCED_SHIPMENT`.
* Staging PR #99 (`272c404`) added the go/ast spike and the rev 7 re-plan.
* Every PR used a merge commit, no `--admin`, and passed the P-018 Copilot gate. Every PR also had multi-model adversarial review, with gpt-5.6-sol as the anchor.
* Later outcome (verified in git history): 039-S and 049-F closed via PR #111 on 2026-10-07. The closure record `docs/closure/039-S-049-F-post-merge-closure.md` is `ready_with_conditions`, with `compaction_status: done`.

## Why 039-S halted

* The go/ast spike (stash `1EEBECA5`) found the migration feasible: G-1 to G-7 were feasible, and G-3 showed exact verdict parity with zero divergences.
* The rev 7 plan-review gate returned FAIL in all three allowed cycles, with P1 findings from every reviewer in cycle 3.
* 049-F stayed `blocked`. The hold was enforced mechanically by Ship Step 0.5 `shipment-reconcile`, not by the topology gate, which checks topology only.
* Options left to the operator: authorise a fourth plan-review round, remediate the cycle 3 P1s, or re-scope 049-F. The rev 7 cycles are summarised in `docs/memory/compacted/2026-10-12-049-f-stage-rev7-plan-review-cycles-compacted.md`.

## Orchestrator rulings

* Closure PRs use `autoharness gate pipeline-topology --mode manual --phase ambient --json`. `--phase lifecycle` applies only while a shipment is active.
* `backlogit hooks poll|ack` is `TOOL_DEGRADED`, not `TOOL_UNAVAILABLE`, because the backlogit MCP was absent.
* Reviewer git preflight failures: pin the diff to a `logs/*.patch` file and have each reviewer print `REVIEWED_HEAD`.
* Missing root-level lock scripts (stash `9F824B64`): the bundled copies under `.github/skills/file-lock/scripts/` are the reference. 038-S used an atomic create-new lock. 032-S recorded the formal condition `reconcile-lock-procedural-deviation`.

## Failed approaches and environment notes

* 038-S needed five Ship attempts. They halted on hooks availability, the review-fix cycle cap, reviewer git preflight, a stuck engram daemon (fixed by stopping its PID; it respawned healthy), and misapplying the lifecycle gate to the closure PR.
* `write_agent` cannot continue a sync-mode `_Ship` task. Start a new task with a complete prompt instead.
* gemini-3.8-flash sometimes returned empty output. grok-4.7 was the working substitute. Pinned review patches need Git Bash ahead of WSL bash on `PATH`.

## Stash state recorded at the time

* Created during the run: `B29A565E`, `F5958BBC`, `48EA04C1`, `3750C37C`, `0ECC1895`, `FE2F02FF`, `C0D28448`.
* Referenced again: `D10E82EC`, `9F824B64`, `D10D3AFC`, `DC921AF6`, `D7BF9F74`.
* This summary does not change any stash entry. Triage stays with Stage.
