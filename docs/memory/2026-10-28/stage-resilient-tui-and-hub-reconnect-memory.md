---
title: "Stage Session Memory — Resilient TUI and Hub Reconnect Experience (Capability Set 1)"
description: "Session memory for Stage execution under DARK_MODE_ACTIVE (P-017) on Capability Set 1 (1DCB6108, 550437DB, E9223340, 4AFE4DA4, AEF57AF1, 8769C9A0, 2C74C5FC)."
date: 2026-10-28
agent: stage
branch: "chore/stage-resilient-tui-and-hub-reconnect"
shipments:
  - "067-S"
  - "068-S"
  - "069-S"
features:
  - "074-F"
  - "075-F"
  - "076-F"
consumed_stash_ids:
  - "1DCB6108"
  - "550437DB"
  - "E9223340"
  - "4AFE4DA4"
  - "AEF57AF1"
  - "8769C9A0"
  - "2C74C5FC"
---

## Session Summary

* **Mode**: `DARK_MODE_ACTIVE` (P-017); Stage Artifact Branch Gate (Step 1.9) passed on `chore/stage-resilient-tui-and-hub-reconnect` branched from `refs/remotes/origin/main` (`7617a00fc29660cac87dd8d67a177518e0d68ec2`, code baseline `066-S` `65780da0d83a53182a56460093b80cc132b5d945`).
* **P-021 C5/C6 Triage Audit**:
  * **Unconditional Duplicate Scan (Trigger A)**: CLEAN across all active (`.backlogit/stash.jsonl`) and archived (`.backlogit/archive/stash.jsonl`) stash entries and `.backlogit/queue/`.
  * **Late-Identifier Reconciliation (Trigger B)**: Reconciled `2C74C5FC` (`PR #177`, `058-S`, merge commit `ec447f7e7fe6140bc99da28802e96b4317860aa3`), `550437DB` / `4AFE4DA4` / `8769C9A0` (`PR #179`, `059-S`, merge commit `8c9f0f299fb2784ec0cbe03b843c84402927595d`), `1DCB6108` / `E9223340` (`PR #183`, `061-S`, merge commit `8bbd57f5907f50c769467708dceee167a4444493`), and `AEF57AF1` (`PR #195`, `065-S`, merge commit `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5`); `thread: N/A (pre-PR local review)` stands as a truthful terminal record.
* **Artifacts Produced**:
  * Deliberation: `docs/decisions/2026-10-28-intercom-go-resilient-tui-and-hub-reconnect-deliberation.md`
  * Implementation Plan (hardened + reviewed): `docs/plans/2026-10-28-intercom-go-resilient-tui-and-hub-reconnect-plan.md` (`dispatch_mode: multi-agent`, `decision: PASS`)
* **Structured Sizing / Complexity Note**: Every task (`074.001-T`..`074.005-T`, `075.001-T`..`075.004-T`, `076.001-T`..`076.005-T`) carries structured `size: S` (`size_source: agent`, `size_ruleset_version: stage-2h-rule-v1`) in frontmatter and prose-embedded `[Size: S | Complexity: <low|medium>]` in the task description (as `.backlogit` task schema defines structured `size` without a separate `complexity` frontmatter field).

## Harvested Backlog Hierarchy & Shipments

1. **Shipment `067-S`** (Covering Feature `074-F`: *"Hub fold, publish, and lifecycle hardening (2C74C5FC, 550437DB, 8769C9A0)"*, declared DAG root with `labels: ["dag-root"]` / `predecessor_source: "declared_root"` under explicit operator P-001 policy exception over `062-S`, code baseline `066-S` `65780da`):
   * `074.001-T` (`size: S`, `complexity: low`, `code`): Harden `Fold` delta cap re-marshalling and live/hydrated `tool.ended` name/truncation parity across `internal/hub/fold.go` and `internal/tui/model.go` (`2C74C5FC(d, e)`).
   * `074.002-T` (`size: S`, `complexity: low`, `code`, blocked by `074.001-T`): Harden `Hub.publish` stub fold-error deduplication, unbuildable-stub `markAllStale`, remote `rehydrate` drain, and `ShutdownConfig.AdapterStopTimeout` validation (`2C74C5FC(a, c, f)`, `8769C9A0`).
   * `074.003-T` (`size: S`, `complexity: medium`, `code`, blocked by `074.002-T`): Settle/resolve permissions on aborted `agent.idle` with turn-scoped ask identity (`PermissionRequest.TurnID`, `activeTurnID`, `abortedTurnIDs`, `intakeAsk`, `drainAbortedTurnAsks`) and preserve next-turn asks dequeued before `KindTurnStart` (`550437DB`).
   * `074.004-T` (`size: S`, `complexity: low`, `code`, blocked by `074.003-T`): Make `Adapter.recheckForce` non-blocking after failed `Start`, export `(*Adapter).StopTimeout()`, wire `permissionBridge` turn-scoped `TurnID` stamping and `turnStop` cancellation in `internal/agent/copilot/permission.go` and `internal/agent/copilot/adapter.go`, migrate `TestShutdownThatCannotConfirmForceStopReportsOrphanRisk`, and add registered-callback turn-abort regression test in `shutdown_hub_test.go` (`8769C9A0`, `550437DB`).
   * `074.005-T` (`size: S`, `complexity: low`, `code`, blocked by `074.004-T`): Join operator lane workers in `Hub.Run` before `Done()` closes and publish sanitized `KindError` (`operator_action_failed`) on non-shutdown async operator `Send`/`Abort` failures (`550437DB`).
2. **Shipment `068-S`** (Covering Feature `075-F`: *"Shutdown broadcast/drain fidelity and in-place hub Client.Rehydrate (4AFE4DA4, AEF57AF1, 1DCB6108)"*, blocked by `067-S`):
   * `075.001-T` (`size: S`, `complexity: low`, `code`): Sweep closed clients at intake, honor pending connects on reap, and deliver terminal `server.shutdown` to stale clients while suppressing subsequent stale hydration (`4AFE4DA4(a, b, c, d)`, `2C74C5FC(b)`, `AEF57AF1(a)`).
   * `075.002-T` (`size: S`, `complexity: medium`, `code`, blocked by `075.001-T`): Add `Client` write-ack `Drained` tracking in `drainAndCloseClients` and in-place `Client.Rehydrate` (`4AFE4DA4(b)`, `AEF57AF1(a)`, `1DCB6108` hub half).
   * `075.003-T` (`size: S`, `complexity: low`, `code`, blocked by `075.002-T`): Wire `EnableWriteAck`/`AckWrite` in `transport.session.writePump` and unconditional `AdapterStopTimeout` in `internal/app/shutdown.go` (`4AFE4DA4(b)`, `AEF57AF1(a)`, `8769C9A0(c)`).
   * `075.004-T` (`size: S`, `complexity: low`, `code`, blocked by `075.003-T`): Propagate `ctx` in `App.Reissue`/`tuiReissue` and drain `remote` concurrently in `shutdownScenario` (`AEF57AF1(b, c)`).
3. **Shipment `069-S`** (Covering Feature `076-F`: *"Resilient TUI reconnect, sanitization, modal anti-spoofing, and viewport hardening (1DCB6108, E9223340)"*, blocked by `068-S`):
   * `076.001-T` (`size: S`, `complexity: medium`, `code`): Use in-place `Conn.Rehydrate` in `tui.Feed` with bounded redial retry and dropped-TUI recovery (`1DCB6108` TUI half, `E9223340(c)` feed half).
   * `076.002-T` (`size: S`, `complexity: low`, `code`, blocked by `076.001-T`): Sanitize non-`Cf` invisible blank runes and add `isBlankSummary` (`E9223340(d)` sanitize half).
   * `076.003-T` (`size: S`, `complexity: low`, `code`, blocked by `076.002-T`): Enforce `isBlankSummary` deny-only in `newPermission` and wrap-aware 2-space summary indentation (`E9223340(d, e)`).
   * `076.004-T` (`size: S`, `complexity: low`, `code`, blocked by `076.003-T`): Add modal-generation `ApprovalDelay` (`DefaultApprovalDelay = 150ms` by default in `tui.New`, explicit `NoApprovalDelay` zero-delay test path), disarm armed approval on resync/hydrate, and trigger `Rehydrate` on stuck `resync` (`E9223340(c, f)`).
   * `076.005-T` (`size: S`, `complexity: low`, `code`, blocked by `076.004-T`): Add live transcript entry cap, delta-invalidated wrap cache, display-width prompt truncation, and mouse wheel initialization/scrolling (`E9223340(b, g)`).

## Consumed Stash Archival & Deferred Scope

* Archived all 7 consumed stash entries via `backlogit stash archive`: `2C74C5FC`, `550437DB`, `4AFE4DA4`, `8769C9A0`, `1DCB6108`, `E9223340`, `AEF57AF1`.
* All other stash entries and queue items outside Capability Set 1 were left untouched.
