---
title: "Ship checkpoint — 021-S RED confirmed, before instruction-surface edits"
date: 2026-09-16
shipment: 021-S
feature: 022-F
task: 022.001-T
branch: feat/021-s-ship-covering-feature-completion-foundation
agent: ship
---

# Checkpoint: 021-S — RED confirmed, entering GREEN phase

## State

* Shipment `021-S` claimed, active. Task `022.001-T` claimed, active. Feature `022-F`
  deliberately left `queued` (self-host constraint — this session must not use the new
  Role Boundary authority it introduces).
* Branch `feat/021-s-ship-covering-feature-completion-foundation` created from `main`, checked
  out, no commits yet.

## Test files created (all new, none yet committed)

1. `tests/integration/shipment_close_path_harness_test.go` — test-only fixture harness:
   disposable `backlogit` workspace builder, whole-tree snapshot/hash helper, Go
   reimplementation of `classifyClosePath` (read-only, zero mutation, zero persistence per
   D-3) and `safeClose` (binding revalidation R1-R5, dispatch to real `backlogit shipment
   ship` for CASCADE, and a snapshot+restore-collateral-drift approach for SAFE_CLOSE since
   the installed backlogit 1.10.1 engine rejects a direct `move --status shipped` on
   shipments and `shipment ship` is the ONLY primitive that can transition a shipment,
   with an observed real side effect of detaching non-member sibling tasks' `parent_id`).
2. `tests/integration/shipment_close_path_behavior_test.go` — behavioral acceptance tests
   Z1 (whole-tree equality, cascade/safe_close/6 block reasons), B4 (no mutation until
   verdict), B1 (4 refusal tokens, zero mutation), B5 (workspace drift after
   classification -> refusal), B2 (bound cascade archives exactly manifest subtree), B3
   (bound safe_close = single-artifact delta, reachable path preserved), B6 (direct engine
   path only reachable through bound dispatch), Z2/Z3/Z4 (return-value-only fields, no
   reconcile report file/dir). **ALL GREEN** after 2 rounds of real bug fixes (fixture
   design bugs, not harness-logic bugs): (1) an "extras" BLOCK fixture accidentally used
   an independently-qualifying feature instead of an orphan task; (2) `classifyClosePath`
   only computed `CLASSIFICATION_BINDING` for CASCADE/SAFE_CLOSE, not for BLOCK verdicts,
   which broke the R4 (bound-BLOCK-refused) refusal path — fixed via a defer-based
   uniform binding computation.
3. `tests/integration/ship_feature_completion_contract_test.go` — the 33-row (44
   transition-assertion / 7 guard-assertion) static needle test over the 3 instruction
   surfaces, using pinned wording extracted verbatim from plan §4.2/§6.1.
   **CONFIRMED RED**: 26 failing subtests (44 assertions) / 7 passing guard subtests —
   exactly matching the plan's stated H0 baseline.

## Next steps (not yet done)

1. Implement CS11-CS14 in `.github/skills/shipment-reconcile/SKILL.md` (producer, lands
   first per plan §7 ordering).
2. Implement CS1-CS9, CS15, CS16, ADD-1, ADD-2 in `.github/agents/_ship.agent.md`
   (consumers).
3. Implement CS10 in `.github/policies/workflow-policies.md` (authority grant).
4. Re-run the 33-row test — confirm GREEN (44 transition rows flip, 7 guard rows still
   hold).
5. Re-run the full behavioral suite — confirm still GREEN.
6. Run full quality gates: gofmt, go vet ./..., go test ./..., go build ./...,
   backlogit sync/doctor.
7. Run review skill (report-only). Remediate same-contract-surface findings only; P-021
   capture for any out-of-scope findings.
8. Commit (conventional message + Co-authored-by trailer), push, create PR with
   `## Local Review Readiness` block.
9. Manage CI/review feedback within bounded protocols. Do NOT merge — new explicit
   operator approval required for this PR.
10. STOP after PR creation / CI stabilization. Do not perform feature 022-F completion or
    021-S shipment closure in this session (self-host constraint: a fresh session uses the
    new authority after merge).

## Hygiene

* Untracked pre-existing files (`diagnostic.ps1`, `report.20260913.140024.22564.0.001.json`,
  `temp_commands.sh`, prior memory files from PR #54) are untouched and will be excluded
  from the implementation commit.
* Scratch fixture directories used for CLI experimentation
  (`C:\Temp\bltest_cascade_2`, `C:\Temp\bltest_safeclose`, `C:\Temp\bltest_safeclose2`) are
  outside the repo and irrelevant to it.
