---
title: "Stage session — Harden retiredarch gate integrity (resumed: Option B split; shipment 040-S queued)"
date: 2026-10-07
agent: stage
status: complete
prior_status: halted-escalation
branch: chore/stage-harden-retiredarch-gate-integrity
plan: docs/plans/2026-10-07-intercom-go-retiredarch-correctness-plan.md
superseded_plan: docs/plans/2026-10-07-intercom-go-retiredarch-gate-integrity-plan.md
shipment_id: 040-S
deliberation: docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md
---

# Stage session memory: Harden retiredarch gate integrity

## Status

**Complete. Shipment 040-S is queued.** The session resumed under Option B (split), planned and
reviewed the narrowed plan, and harvested 050-F. See "Resumed session" below.

### Initial halt (historical)

The first pass **halted at Step 4: plan review**. Attempts 1, 2 and 3 on the original 14-unit
plan all returned FAIL, which exhausted the two re-entry cycles. Per Stage policy, the
escalation protocol (P-013.6) was invoked and the session halted for an operator decision.

At the time of that halt:

* No harvest, shipment or stash archival had happened.
* No backlog item or shipment had been created.
* All six stash entries were still **active**.

The operator's Option B disposition later superseded this state.

## Session inputs

* **Orchestrator `stage next`**, approved by the operator at 2026-10-07T13:55-07:00 ("let's go
  with the recommendation"). The approval was relayed by the Orchestrator.
* **Batch:** six task-shaped stash entries. All carry a DEFERRED SCOPE EXPANSION, so all were
  routed to deliberate.

| Stash ID | Priority | Plan units (rev 3) |
|---|---|---|
| `3750C37C` | high | RA-12, RA-13, RA-14 |
| `9FC28DB9` | high bug | RA-1, RA-2 |
| `0ECC1895` | — | RA-11 |
| `DC921AF6` | — | RA-6..RA-10 |
| `D7BF9F74` | — | RA-4, RA-5 |
| `990AFA71` | — | RA-3 |

## Steps completed

| Step | Outcome |
|---|---|
| 0.0 tools | `TOOL_OK` (CLI). MCP-only operations are degraded: hook polling, `add_to_shipment`, `stash_edit` |
| 0.1 index sync | `INDEX_SYNC_OK` |
| Checkpoint recovery | Zero stage-owned candidates, so a normal start |
| 1 triage | All six entries task-shaped, with a DEFERRED SCOPE EXPANSION, so deliberate was forced. Duplicate scan (extended) **CLEAN** |
| 1, late-identifier reconciliation | `3750C37C`/`0ECC1895`: PR #95. `DC921AF6`/`D7BF9F74`: residual carrier PR #95. `990AFA71`: thread `PRRT_kwDOTPuhps6nd8rH` (PR #83), recovered via the GitHub API |
| 1.5 grouping | Operator-selected grouping, all six. Covering feature: "Harden retiredarch gate integrity" |
| 1.8 learnings | Medium confidence; consulted |
| 1.9 branch gate | PASS. `chore/stage-harden-retiredarch-gate-integrity` from origin/main `372ab38`. Dirty-worktree deviation disclosed: operator-owned `.gitignore` and `.backlogit/stash.jsonl` CRLF noise carried over, never staged |
| 2 deliberation | D-RA-1..D-RA-6, plus addenda after attempts 1 and 2 |
| 3 impl-plan + plan-harden | Revision 3: RA-1..RA-14. `Requires plan hardening: yes`. Hardening section inline |
| 4 plan-review | Attempt 1 **FAIL** (6 personas). Attempt 2 **FAIL** (4). Attempt 3 **FAIL** (4): Arch 3 P1, Security 2 P1, Go 1 P1, Scope ADVISORY |
| 5 / 5.5 / 5.6 | **Not run** (halted) |

## Escalation payload (P-013.6 contract)

```yaml
threshold_kind: review_fix_cycles            # Stage plan-review re-entry cap
threshold_count: 3                           # attempts; 2 re-entry cycles exhausted
failure_summary: >
  Plan revision 3 for the six-entry retiredarch gate-integrity batch failed final review
  with 6 new P1s. Three concern the RA-4/RA-5 git exec-location guard (`gitPathInside`):
  it is absent from the select.go closed world and frozen sets, its filepath.Abs/EvalSymlinks
  conflict with the existing init-reachability security tests, and filepath.Rel errors
  across Windows volumes. One is a red row made non-red by indexDecls' blank-identifier key.
  Two are RA-10 dispatch-table bypasses: raw or escaped literal spellings, and assembly or
  .syso plus bodyless funcs. Suspected root cause: the pin trust-root hardening units
  (RA-6..RA-14) chase an open-ended adversarial surface. Each review cycle has found new
  in-model bypasses in the AST-rule approach. The bug-fix units (RA-1..RA-3) have had no
  P1 since revision 2.
last_n_action_refs:
  - plan-review attempt 1 (rev 1): 6 persona subagents
  - plan-review attempt 2 (rev 2): 4 persona subagents (arch on anchor route gpt-6.1-sol)
  - plan-review attempt 3 (rev 3): subagents arch-review-3, security-review-3, go-review-3, scope-review-3
last_n_observation_refs:
  - "plan § Plan Review — attempt 1 (revision 1)"
  - "plan § Plan Review — attempt 2 (revision 2)"
  - "plan § Plan Review — attempt 3 (revision 3, FINAL)"
artifact_refs:
  - docs/plans/2026-10-07-intercom-go-retiredarch-gate-integrity-plan.md
  - docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md
  - stash: 3750C37C, 9FC28DB9, 0ECC1895, DC921AF6, D7BF9F74, 990AFA71
evidence_path: ""                            # telemetry not enabled
resumption_checkpoint_ref: docs/memory/2026-10-07/stage-retiredarch-gate-integrity-memory.md
resolved_escalation_route:
  model_family: gpt-6.1-sol                  # no stage.escalation, no legacy flat → tier3
  model_provider: openai
  reasoning_effort: xhigh
```

**Route resolution and guard checks:**

* **Config reload (H6).** `.autoharness/config.yaml` was re-read at the halt and parsed
  successfully.
  * `model_routing.stage.escalation` is absent.
  * The legacy flat `model_routing.escalation` is absent.
  * No both-present ambiguity (H2) applies.
* **Same-route guard.** The resolved tier3 route is not the same as Stage's own role route
  (`claude-opus-5.5`/anthropic/high), so the guard does not fire.
* **Handoff surface.** The engram CLI is present (`C:\Tools\engram.exe`). The handoff is this
  `docs/memory` document, which is ingested into engram context.
* **No re-execution.** Stage does not run a fourth plan-review attempt (the circuit is open).

## Operator decision required (choose one)

* **(A) Authorize one extra revision cycle (revision 4).**
  * This overrides the review-cycle cap.
  * It folds in every attempt-3 P1 with the candidate fixes listed in the plan.
  * The main open design question is A3-P1-2: the inside-root exec guard versus the existing
    init-reachability tests.
  * It risks more novel bypass findings in the pin-hardening units.
* **(B) Recommended: split the batch** (reliability first, and it keeps proven-clean fixes
  moving):
  * **Shipment 1** (re-staged quickly):
    * RA-1/RA-2 (`9FC28DB9`, the high-priority false-clean bug);
    * RA-3 (`990AFA71`, symlink containment);
    * RA-4/RA-5 (`D7BF9F74`, git env isolation), **without** the inside-root exec guard.
      `cmd.Err` and an absolute `cmd.Path` are kept. In-checkout `git` stays an accepted
      residual under R-A2a, which removes A3-P1-1, A3-P1-2 and G3-P1-1.
  * **Shipment 2:** `DC921AF6`, `3750C37C` and `0ECC1895` (RA-6..RA-14) go back to deliberation
    with a bounded threat model. This deliberation should decide whether package-`main`,
    assembly and sibling-package init vectors are covered by AST rules or by an out-of-band
    governance control (for example, CODEOWNERS or required review on `tools/gatecheck/**`).
* **(C) Re-deliberate the whole batch.** Replace the escalating AST-freeze approach (D-RA-4,
  D-RA-5, D-RA-6) with a governance control plus a minimal pin, then re-plan.

## Deferred and new candidate stash entries (not created; pending the operator decision)

* **Cross-engine symlink containment** (writepath/unignore/mergestrategy). Source: `990AFA71`,
  PR #83, `PRRT_kwDOTPuhps6nd8rH`. Low priority.
* **R-A2d (new, found by the Go reviewer in attempt 3):** `selectRepoPaths` reads
  `git ls-files` without `-z`. With the default `core.quotePath=true`, any quoted path
  (non-ASCII, `"`, `\`, control characters) under `cmd/`/`internal/` is silently dropped by
  `shouldScanRepoPath` and never scanned. This is a fail-open, so it belongs in reliability
  and security; candidate priority high. It is out of this batch's scope, because scope
  changes were forbidden.
* **Compound-learning candidate:** maintaining pin canonical texts, and the observation that
  AST-rule trust roots invite unbounded adversarial review cycles.

## Next steps on resume

1. The operator picks (A), (B) or (C).
2. **(B):** re-run Stage from Step 2. Add a D-RA-7 split addendum, write a new plan for
   shipment 1 (RA-1..RA-5 minus the guard), and get a fresh review cycle (counter resets for
   the new plan). Then harvest, shipment, and archive of `9FC28DB9`, `990AFA71` and
   `D7BF9F74` only. The other three entries stay active for the shipment 2 deliberation.
3. **(A):** write revision 4. Re-review needs explicit operator override of the cap,
   recorded in the plan.
4. Stash the candidate entries above under whichever option is chosen.

## Resumed session (2026-10-07, after the escalation halt)

**Disposition.** Option B: split the batch. This was the operator/Orchestrator decision at
about 14:30-07:00, recorded as D-RA-7 in the decision record. It was not a 4th review cycle.
The old 14-unit plan is marked superseded and kept for history.

* **Step 1.9 gate:** PASS on resume. The branch was unchanged and the diff touched only Stage
  artifact roots.
* **New stash entries:**
  * 4537B2F6: HIGH bug, `git ls-files` without `-z`, select.go:33.
  * D44D8BDF: LOW, cross-engine symlink containment.
  * The duplicate scan was clean.
* **Narrowed plan:** `docs/plans/2026-10-07-intercom-go-retiredarch-correctness-plan.md`.
  It has 6 units (U1..U6) and covers 9FC28DB9, 990AFA71 and D7BF9F74. It was hardened inline
  (PA-1..PA-7).
* **Plan review (fresh 3-attempt cap):**
  * Attempt 1 (rev 1): FAIL, with P1s AS-1, AS-2/G-1/CR-2 and CR-1.
  * Attempt 2 (rev 2): ADVISORY with no P1. The advisories were folded into rev 3 in place.
  * The rev-1 text was replaced in place: it was deleted before being re-read and never
    committed. The attempt-1 record is kept in the plan.
* **Harvest:**
  * Feature 050-F.
  * Tasks 050.001-T (U1, S), 050.002-T (U2, S), 050.003-T (U3, S), 050.004-T (U4+U5, L)
    and 050.006-T (U6, S).
  * U4 and U5 were first harvested separately, as 050.004-T and 050.005-T. They were merged
    into 050.004-T after the PR #109 review: Ship's per-task gates cannot hold a
    deliberately red, uncommittable U4-only task. 050.005-T was removed.
  * Complexity is recorded as prose, because the task type has no complexity field
    (degradation flagged).
* **DAG:** 050.002-T is blocked by 050.001-T. 050.006-T is blocked by 050.004-T.
  050.003-T is independent.
  * 050.004-T is a single commit (ALP-2).
* **Shipment:** 040-S is queued. It holds 050-F plus 5 tasks in dependency order.
* **Archived:** 9FC28DB9, 990AFA71 and D7BF9F74.
* **Annotated (left active for re-deliberation with a bounded threat model; CODEOWNERS is
  the alternative):** DC921AF6, 3750C37C and 0ECC1895.
* **Compound:**
  `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`.
* **Accepted residuals:** R-A2a (git earlier on PATH, including inside the checkout or via
  `$GITHUB_PATH`), R-A2b, R-A2c, R-A2d (= 4537B2F6), R-A2e (git < 2.32), R-A2f (lost
  `safe.directory`), R-T1 (TOCTOU), R-L1 (local reparse placeholders), R-C1 (= D44D8BDF).
* **Next:** Ship claims 040-S. The operator pushes this Stage branch and opens the staging PR.