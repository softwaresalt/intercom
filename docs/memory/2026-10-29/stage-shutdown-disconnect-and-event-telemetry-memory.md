---
title: "Stage Session Memory — End-to-End Shutdown Reliability, Remote Disconnect, and Live Copilot Resilience (Capability Set 1)"
description: "Session memory for Stage execution under DARK_MODE_ACTIVE (P-017) on Capability Set 1 (407291BE, 70BEBF29, FA1C38F7, 80996C79, A999ABD6, DA760FBB, E7B78125)."
date: 2026-10-29
agent: stage
branch: "chore/stage-shutdown-reliability-and-event-telemetry"
shipments:
  - "070-S"
  - "071-S"
  - "072-S"
features:
  - "077-F"
  - "078-F"
  - "079-F"
consumed_stash_ids:
  - "407291BE"
  - "70BEBF29"
  - "FA1C38F7"
  - "80996C79"
  - "A999ABD6"
  - "DA760FBB"
  - "E7B78125"
---

## Session Summary

* **Mode**: `DARK_MODE_ACTIVE` (P-017); Stage Artifact Branch Gate (Step 1.9) passed on `chore/stage-shutdown-reliability-and-event-telemetry` branched from `refs/remotes/origin/main` (`078713bc50482b40c7b879f7bb4b037b7f6cb304`, code baseline `069-S` `5f561d2c6f25ff01db6ddc298e2392bf2380de27`).
* **P-021 C5/C6 Triage Audit**:
  * **Unconditional Duplicate Scan (Trigger A)**: CLEAN across all active (`.backlogit/stash.jsonl`) and archived (`.backlogit/archive/stash.jsonl`) stash entries and `.backlogit/queue/`.
  * **Late-Identifier Reconciliation (Trigger B)**: Reconciled `407291BE` and `70BEBF29` (`PR #202`, `067-S`, merge commit `a2cacf32c29b5e3a0ee596f01f14686ceeb3775d`), `FA1C38F7`, `DA760FBB`, and `E7B78125` (`PR #175`, `057-S`, merge commit `e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2`), `80996C79` (`PR #205`, `068-S`, merge commit `2fff94b6722ff763aafc04a5e44432f2733bd58c`), and `A999ABD6` (`PR #207`, `069-S`, merge commit `5f561d2c6f25ff01db6ddc298e2392bf2380de27`).
* **Artifacts Produced**:
  * Deliberation: `docs/decisions/2026-10-29-intercom-go-shutdown-disconnect-and-event-telemetry-deliberation.md`
  * Implementation Plan (hardened + reviewed): `docs/plans/2026-10-29-intercom-go-shutdown-disconnect-and-event-telemetry-plan.md` (`dispatch_mode: multi-agent`, `decision: PASS`)
* **Structured Sizing / Complexity Note**: Every task (`077.001-T`..`077.004-T`, `078.001-T`..`078.002-T`, `079.001-T`..`079.006-T`) carries structured `size: S` (`size_source: agent`, `size_ruleset_version: stage-2h-rule-v1`) in frontmatter and prose-embedded `[Size: S | Complexity: <low|medium>]` in the task description (as `.backlogit` task schema defines structured `size` without a separate `complexity` frontmatter field).

## Harvested Backlog Hierarchy & Shipments

1. **Shipment `070-S`** (Covering Feature `077-F`: *"Copilot adapter Stop/ForceStop resilience, latch/permission hardening, and App.Err stuck-lane reporting (407291BE, 70BEBF29, FA1C38F7)"*, declared DAG root with `labels: ["dag-root"]` / `predecessor_source: "declared_root"` under explicit operator P-001 policy exception over `062-S`, code baseline `069-S` `5f561d2`):
   * `077.001-T` (`size: S`, `complexity: low`, `code`): Harden `sanitizeBound` `Zl`/`Zp` replacement, cap `PermissionRequest` identifiers (`sanitizeIdent` at `128` runes), and re-check `gone()` before ask send (`FA1C38F7(a, b, d)`).
   * `077.002-T` (`size: S`, `complexity: low`, `code`, blocked by `077.001-T`): Quiesce callback latch on failed `Start`, classify `session.shutdown` informational, and deliver post-quiesce terminal `session.idle` non-blockingly (`FA1C38F7(c, f)`).
   * `077.003-T` (`size: S`, `complexity: medium`, `code`, blocked by `077.002-T`): Account for in-flight `Client.Stop()` `startStopMux` lock contention (`SDKRuntimeShutdownTimeout = 10s`, `ForceStopBudget()`, `DefaultShutdownForce = 12s`) in `Adapter.forceStopWait()` and `hub.NewShutdown` (`407291BE`).
   * `077.004-T` (`size: S`, `complexity: low`, `code`, blocked by `077.003-T`): Record `Hub.Run`'s `ErrCleanupIncomplete` in `App.Err` via `a.addErr` before `hubRunDone` closes (`70BEBF29`).
2. **Shipment `071-S`** (Covering Feature `078-F`: *"Remote Cursor/wsclient terminal server.shutdown gap handling and TUI permission modal kind:/tool call: continuation anti-spoofing (80996C79, A999ABD6)"*, blocked by `070-S`):
   * `078.001-T` (`size: S`, `complexity: low`, `code`): Treat `KindServerShutdown` as always-applicable (`ActionApply` / `VerdictApply`) across sequence gaps in `hub.Cursor` and `wsclient` while setting `c.stale = true` on non-contiguous terminal delivery and discarding duplicates (`80996C79`).
   * `078.002-T` (`size: S`, `complexity: low`, `code`, blocked by `078.001-T`): Indent soft-wrap continuation lines of `kind:` and `tool call:` via `wrapHeader` in `tui.Model.permissionBody` (`A999ABD6`).
3. **Shipment `072-S`** (Covering Feature `079-F`: *"agent.Event rune-safe bounds, tool.ended/reasoning enrichment, write-diff/shell-URL surfacing, and unmapped-event log rate-limiting (DA760FBB, FA1C38F7(e), E7B78125)"*, blocked by `071-S`):
   * `079.001-T` (`size: S`, `complexity: low`, `code`): Surface write diff/new-file-content (with explicit `[write preview truncated]` notice when clipped) and shell `possibleUrls` in `PermissionRequest.Summary` (`FA1C38F7(e)`).
   * `079.002-T` (`size: S`, `complexity: medium`, `code`, blocked by `079.001-T`): Add `agent.Event` reasoning kinds, `Truncated` flag, `tool.ended` `Text` extraction, single-pass `boundUTF8`, and `unmappedLogGate` (`DA760FBB`, `E7B78125`).
   * `079.003-T` (`size: S`, `complexity: low`, `code`, blocked by `079.002-T`): Map `AssistantReasoningData` and `AssistantReasoningDeltaData` and wire `unmappedLogGate` in `Adapter.onEvent` (`DA760FBB`, `E7B78125`).
   * `079.004-T` (`size: S`, `complexity: low`, `code`, blocked by `079.003-T`): Map and fold `agent.reasoning`, `agent.reasoning.delta`, `tool.ended` `Text`, and `Event.Truncated` in `internal/hub` (`DA760FBB`).
   * `079.005-T` (`size: S`, `complexity: low`, `code`, blocked by `079.004-T`): Publish live `agent.reasoning.delta`, rate-limit hub unmapped event warns, and render reasoning and tool text in TUI (`DA760FBB`, `E7B78125`).
   * `079.006-T` (`size: S`, `complexity: low`, `docs`, blocked by `079.005-T`): Amend normative sections 3.1, 4.2, 4.3, and 5.2 in `docs/design-docs/intercom-go-backend-architecture.md` (`407291BE`, `FA1C38F7(f)`, `E7B78125`, `DA760FBB`).

## Consumed Stash Archival & Deferred Scope

* Archived all 7 consumed stash entries via `backlogit stash archive`: `407291BE`, `70BEBF29`, `FA1C38F7`, `80996C79`, `A999ABD6`, `DA760FBB`, `E7B78125`.
* All other stash entries and queue items outside Capability Set 1 were left untouched.
