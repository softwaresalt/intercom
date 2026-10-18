---
title: "Compacted memory — Ship 050-S (feature 020-F): Stage default-branch contract enforcement"
date: 2026-10-17
shipment: 050-S
feature: 020-F
status: shipped
pr_main: 157
merge_commit_sha: 30ca417a9cb32770fb7ca6671c36f90fef09f1e4
compacted_from:
  - docs/archive/memory/050-s-ship-memory.md
  - docs/archive/memory/050-s-ship-halt-memory.md
---

# Compacted Memory — Ship 050-S (020-F)

## Run

* Mode: DARK_MODE_ACTIVE (P-017), scope 050-S only; merge pre-authorized; no admin fallback.
* Branch `feat/050-s-mechanical-enforcement-of-the-stage-default-branch-contract` from main a40347d; pre_claim and post_claim topology gates passed; claim verified active.
* Tasks (JIT per-task harness, red then green): 020.001-T 8a46672 / fe4e9ee; 020.002-T 6a52918 / 4380c82; 020.003-T 6c76ffa / 9ce4c8a; 020.006-T 2a0e584 / 4dc3ea0; 020.004-T 0ada8be / 8bb4fe4; 020.005-T 23d626f / 02a620c. Review remediation: 16e6ad7 (red tests) / 07e8f7e (fixes).

## Decisions and deviations

* The mutation proof mutates the policies copy (detectP3 reads the P3 grant from workflow-policies.md), not `_stage.agent.md` as the plan text said; the plan allowed either with an exactly-one-occurrence assertion.
* Done-task archive moves were withheld from the PR (plan C-1 convention) and landed in the closure PR.
* Telemetry begin/record not exercised (delegated subagents).

## Review

* Correctness, Security, Scope reviewers: no P0/P1. P2 findings fixed (P6 historical, lazy continuation, P5 missing rows, P4 scope, P3 assertion).
* Deferred P3 items captured as stash 6333B903 (P-021 C2).

## Merge and closure

* PR #157 merged as 30ca417a9cb32770fb7ca6671c36f90fef09f1e4; CI green; copilot-review gate SATISFIED after 3 fix rounds.
* Closure branch `post-merge/020-stage-default-branch-contract-enforcement`: 020-F moved active to done; stale queue copies of 020.001..006-T removed; lifecycle gate OK; `cascade-close --classify-only` exit 0 (CASCADE, engine VERIFIED); evidence `docs/closure/evidence/050-S-020-F-close-evidence.json`.
* Halt checkpoint (SUPERSEDED/resolved): the mutating cascade-close required explicit per-closure operator approval; the operator approved and it succeeded.

## Open follow-ups

* Operator-owned, not performed: required code-owner review in branch protection / base-ref required workflow.
