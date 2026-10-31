---
title: "Stage Session Memory — Permission Sanitization, Confusable URL Defense, CLI Path Trust, and Process Lifecycle Hardening (Option 2)"
description: "Session memory for Stage execution under DARK_MODE_ACTIVE (P-017) on Option 2: Capability Set 1 (2EE9F3ED, D3D0E9D5, F24EC587, 2C45EFAE, 2F1B6D4B) and Capability Set 2 (24CE7E1D, 5E4834CF, AE317BD3, 4B7308F2, 7BFA4182, 26E15024, 864FDF7A, D953F24D)."
date: 2026-10-31
agent: stage
branch: "chore/stage-permission-cli-path-and-app-lifecycle"
shipments:
  - "078-S"
  - "079-S"
features:
  - "085-F"
  - "086-F"
consumed_stash_ids:
  - "2EE9F3ED"
  - "D3D0E9D5"
  - "F24EC587"
  - "2C45EFAE"
  - "2F1B6D4B"
  - "24CE7E1D"
  - "5E4834CF"
  - "AE317BD3"
  - "4B7308F2"
  - "7BFA4182"
  - "26E15024"
  - "864FDF7A"
  - "D953F24D"
---

## Session Summary

* **Mode**: `DARK_MODE_ACTIVE` (P-017); Stage Artifact Branch Gate (Step 1.9) passed on `chore/stage-permission-cli-path-and-app-lifecycle` branched from `refs/remotes/origin/main` (`f00ef564f49ff38a0ad9c3d9f756eb3a11edb59d`, code baseline `077-S` `3b1135530989eb8fe60a55fcab768ca61fd3cd74`).
* **P-021 C5/C6 Triage Audit**:
  * **Unconditional Duplicate Scan (Trigger A)**: 11 entries `CLEAN` across `.backlogit/stash.jsonl`, `.backlogit/archive/stash.jsonl`, and `.backlogit/queue/`; `7BFA4182` (`065-S`) and `26E15024` (`066-S`) identified as `COMPLEMENTARY OVERLAP` on moving `watchUI`/`pumpRejects`/`onReject` into `internal/app/ui.go`, reconciled into unified tasks in `086-F`, and both archived at Step 5.6.
  * **Late-Identifier Reconciliation (Trigger B)**: Reconciled `2EE9F3ED`, `D3D0E9D5`, `F24EC587` (`PR #228`, `076-S` / `083-F`, merge `e15937714e489f6c55485651029db7cee031b91f`), `2C45EFAE`, `2F1B6D4B` (`PR #225`, `075-S` / `082-F`, merge `a2898942a12c70fe23be050d698ae4622153d97f`), `24CE7E1D`, `5E4834CF`, `AE317BD3`, `7BFA4182`, `864FDF7A`, `D953F24D` (`PR #195`, `065-S` / `072-F`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5`), and `4B7308F2`, `26E15024` (`PR #198`, `066-S` / `073-F`, merge `65780da0d83a53182a56460093b80cc132b5d945`).
* **Artifacts Produced**:
  * Deliberation: `docs/decisions/2026-10-31-intercom-go-permission-sanitization-cli-path-and-app-lifecycle-deliberation.md`
  * Implementation Plan (hardened + reviewed): `docs/plans/2026-10-31-intercom-go-permission-sanitization-cli-path-and-app-lifecycle-plan.md` (`dispatch_mode: multi-agent`, `decision: PASS` across all 6 personas including Anchor `gpt-6.1-sol`)
* **Structured Sizing / Complexity Note**: Every task (`085.001-T`..`085.005-T`, `086.001-T`..`086.009-T`) carries structured `size: S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`) in frontmatter and prose-embedded `[Size: S | Complexity: <low|medium>]` in the task description (as `.backlogit/header-def.yaml` defines structured `size` without a separate `complexity` frontmatter field).

## Harvested Backlog Hierarchy & Shipments

1. **Shipment `078-S`** (Covering Feature `085-F`: *"Permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, and Copilot cli_path resolution hardening (2EE9F3ED, D3D0E9D5, F24EC587, 2C45EFAE, 2F1B6D4B)"*, declared DAG root with `labels: ["dag-root"]`):
   * `085.001-T` (`size: S`, `complexity: low`, `code`): Guard `truncUTF8` non-positive bounds and sanitize operator deny feedback in `internal/hub` (`2EE9F3ED`).
   * `085.002-T` (`size: S`, `complexity: medium`, `code`, blocked by `085.001-T`): Sanitize adapter deny feedback, fold NFKC `urls:` confusables with interior invisibles, and lock emoji/ZWJ/PUA policy in `internal/agent/copilot` and `internal/tui` (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`).
   * `085.003-T` (`size: S`, `complexity: medium`, `code`, blocked by `085.002-T`): Add shared `pathsafe.ResolveCLIPath` and `SanitizeCLIPathForMessage` with canonicalization and batch/ADS/exec checks in `internal/pathsafe` (`2C45EFAE`, `2F1B6D4B`).
   * `085.004-T` (`size: S`, `complexity: low`, `code`, blocked by `085.003-T`): Wire `pathsafe.ResolveCLIPath` and `SanitizeCLIPathForMessage` into `internal/config` and `internal/agent/copilot` (`2C45EFAE`, `2F1B6D4B`).
   * `085.005-T` (`size: S`, `complexity: low`, `docs`, blocked by `085.004-T`): Update backend architecture design doc and config reference for permission sanitization and `cli_path` hardening (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`).
2. **Shipment `079-S`** (Covering Feature `086-F`: *"Process lifecycle, credential hygiene, localUI supervision, and config.Unconsumed architecture (24CE7E1D, 5E4834CF, AE317BD3, 4B7308F2, 7BFA4182, 26E15024, 864FDF7A, D953F24D)"*, blocked by `078-S`):
   * `086.001-T` (`size: S`, `complexity: low`, `code`): Convert `transport.Token` to a wipeable byte slice with `Wipe` and `Clone` in `internal/transport` (`AE317BD3`).
   * `086.002-T` (`size: S`, `complexity: low`, `code`, blocked by `086.001-T`): Render metadata-only partial `Reissue` credentials as `(already shown)` and surface one-time unconsumed-keys TUI notice in `internal/tui` (`5E4834CF`, `D953F24D`).
   * `086.003-T` (`size: S`, `complexity: low`, `code`, blocked by `086.002-T`): Add `config.Unconsumed` with reflect-based schema completeness test and derive runbook test from it (`864FDF7A`).
   * `086.004-T` (`size: S`, `complexity: medium`, `code`, blocked by `086.003-T`): Harden `app` credentials with `zeroTokensLocked`, `mint` `credsCleared` early-out, metadata-only partial `Reissue`, and disjoint `Token` cloning (`4B7308F2`, `5E4834CF`, `AE317BD3`).
   * `086.005-T` (`size: S`, `complexity: low`, `code`, blocked by `086.004-T`): Unexport `TUIOptions.retainPlaintext` in `internal/app` (`AE317BD3`).
   * `086.006-T` (`size: S`, `complexity: low`, `code`, blocked by `086.005-T`): Group `App` credential fields into `credentialState` mirroring `shutdownState` (`26E15024`).
   * `086.007-T` (`size: S`, `complexity: medium`, `code`, blocked by `086.006-T`): Move `teardown` beside `reap` in `shutdown.go`, clear credentials at shutdown start in `supervise`, and test `NewShutdown` teardown fallback (`24CE7E1D`, `26E15024`, `864FDF7A`, `D953F24D`).
   * `086.008-T` (`size: S`, `complexity: medium`, `code`, blocked by `086.007-T`): Move `watchUI`, `pumpRejects`, and `onReject` into `ui.go`, own `watchDone` inside `startUI`, wire `UnconsumedKeys`, and expose `App.RejectedUpgrades` (`7BFA4182`, `26E15024`, `D953F24D`).
   * `086.009-T` (`size: S`, `complexity: low`, `docs`, blocked by `086.008-T`): Update backend architecture §3.1 and operator runbook for failed-start `teardown`, partial `Reissue` display, and unconsumed-keys TUI notice (`5E4834CF`, `864FDF7A`, `D953F24D`).

## Consumed Stash Archival

* Archived all 13 consumed stash entries via `backlogit stash archive`: `2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`, `24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`, `7BFA4182`, `26E15024`, `864FDF7A`, `D953F24D`.
