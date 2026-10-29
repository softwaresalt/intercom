---
title: "End-to-End Shutdown Reliability, Remote Disconnect, and Live Copilot Resilience Implementation Plan"
description: "Implementation plan for Capability Set 1 across three chained single-feature shipments (070-S..072-S): Copilot adapter Stop/ForceStop resilience, latch/permission hardening, and App.Err stuck-lane reporting (077-F); remote Cursor/wsclient terminal server.shutdown gap handling and TUI permission modal kind:/tool call: continuation anti-spoofing (078-F); and agent.Event rune-safe bounds, tool.ended/reasoning enrichment, write-diff/shell-URL surfacing, and unmapped-event log rate-limiting (079-F)."
source: "docs/decisions/2026-10-29-intercom-go-shutdown-disconnect-and-event-telemetry-deliberation.md"
status: "ready"
date: 2026-10-29
tags:
  - "shutdown"
  - "copilot"
  - "hub"
  - "wsclient"
  - "tui"
  - "permissions"
  - "telemetry"
---

## Problem Frame

With `067-S..069-S` merged and closed on `main` (`069-S` merge commit `5f561d2c6f25ff01db6ddc298e2392bf2380de27`, closure commit `078713bc50482b40c7b879f7bb4b037b7f6cb304`), Capability Set 1 (**End-to-End Shutdown Reliability, Remote Disconnect & Live Copilot Resilience**) resolves seven active stash entries (`407291BE`, `70BEBF29`, `FA1C38F7`, `80996C79`, `A999ABD6`, `DA760FBB`, and `E7B78125`) spanning `internal/agent`, `internal/agent/copilot`, `internal/app`, `internal/hub`, `internal/transport/wsclient`, `internal/tui`, and `docs/design-docs/intercom-go-backend-architecture.md`.

To satisfy Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and maintain strict architectural layering, the work is organized into three sequential single-feature shipments (`070-S -> 071-S -> 072-S`). Shipment `070-S` is declared a DAG root (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`) under the operator's explicit P-001 policy exception over `062-S`'s open operator-present live runtime conditions, with `069-S` (`5f561d2c6f25ff01db6ddc298e2392bf2380de27` / `078713bc50482b40c7b879f7bb4b037b7f6cb304`) documented as the code baseline.

## Requirements Trace

| Source Requirement | Stash Origin | Target Shipment & Feature | Implementation Unit(s) |
|---|---|---|---|
| Replace Unicode `Zl`/`Zp` (`U+2028`/`U+2029`) with `U+FFFD` in `sanitizeBound` while preserving `\n` and `\t` in multi-line summaries; sanitize and cap `PermissionRequest.RawKind` and `PermissionRequest.ToolCallID` at `maxIdentRunes = 128` via `sanitizeIdent`; re-check `if gone() { return unavailable() }` immediately before sending on `b.asks` in `permissionBridge.handle` | `FA1C38F7(a, b, d)` | `070-S` / `077-F` | Unit 1.1 (`077.001-T`) |
| Call `a.latch.quiesce(0)` in `Adapter.Start` `fail()`; add `"session.shutdown"` to `informationalTypes`; deliver terminal `*rpc.SessionIdleData` arriving between `latch.quiesce` and `sess.Disconnect` non-blockingly when `HubDone` is open and `Ingress` has room | `FA1C38F7(c, f)` | `070-S` / `077-F` | Unit 1.2 (`077.002-T`) |
| Export `SDKRuntimeShutdownTimeout = 10 * time.Second`, `SDKProcessExitTimeout = 10 * time.Second`, and `DefaultSDKStopLockTimeout = SDKRuntimeShutdownTimeout + SDKProcessExitTimeout` (`20 * time.Second`); add `Config.SDKStopLockTimeout` (defaulting to `DefaultSDKStopLockTimeout` for real SDK clients via `isRealSDKClient()` independently of `StopTimeout`, and `0` for `newWithClient` fake clients unless explicitly configured); track `clientStopStart` and `clientStopActive` in `Adapter.doStop`; extend `Adapter.forceStopWait()` by `max(SDKStopLockTimeout-time.Since(clientStopStart), 0)` when `clientStopActive` is true; export `(*Adapter).ForceStopBudget() time.Duration`; and in `hub.NewShutdown` fold `ForceStopBudget()` into `cfg.AdapterStopTimeout` so post-cutoff `Force >= ForceStopBudget()` (defaulting `Force = max(DefaultShutdownForce, AdapterStopTimeout+2s)`, `27s` for the default real adapter) | `407291BE` | `070-S` / `077-F` | Unit 1.3 (`077.003-T`) |
| Record `Hub.Run`'s returned error (`ErrCleanupIncomplete` on a stuck operator lane) in `App.Err` via `a.addErr(fmt.Errorf("hub: %w", err))` before `hubRunDone` closes in `internal/app/assemble.go`, and add an `internal/app` stuck-lane shutdown regression test | `70BEBF29` | `070-S` / `077-F` | Unit 1.4 (`077.004-T`) |
| Treat `KindServerShutdown` as always-applicable (`ActionApply` / `VerdictApply`) across sequence gaps, stale state, or missing baseline in `(*hub.Cursor).Apply` and `wsclient` `(*cursor).apply` while marking `c.stale = true` on non-contiguous terminal delivery and discarding duplicate `KindServerShutdown` frames | `80996C79` | `071-S` / `078-F` | Unit 2.1 (`078.001-T`) |
| Wrap `kind: <Kind>` and `tool call: <ToolCallID>` via `m.wrapHeader` in `internal/tui/permission.go` `permissionBody` so continuation lines (`i >= 1`) are indented by two spaces (`"  "`) within `inner` columns and cannot imitate column-0 modal headers | `A999ABD6` | `071-S` / `078-F` | Unit 2.2 (`078.002-T`) |
| Surface `PermissionRequestShell.PossibleURLs` (`"\nurls: ..."`) and `PermissionRequestWrite` `Diff` / `NewFileContents` (sanitized and bounded into remaining `maxSummaryRunes = 2048` budget with an explicit `"\n[write preview truncated]"` suffix when clipped, without auto-denying writes whose `FileName` fits and leaves room for the notice) in `toDomainRequest` | `FA1C38F7(e)` | `072-S` / `079-F` | Unit 3.1 (`079.001-T`) |
| Add `KindReasoning`, `KindReasoningDelta`, and `Event.Truncated` to `internal/agent/agent.go`; enforce single-pass rune-safe `MaxEventTextBytes` (`16384`) and `MaxEventIdentBytes` (`128`) bounds, extract `ToolExecutionCompleteData` result/error text into `Event.Text`, and deduplicate unmapped SDK event `Warn` logs via `unmappedLogGate` (`Warn` once per bounded/sanitized event type, then `Debug`) in `internal/agent/copilot/normalize.go` | `DA760FBB`, `E7B78125` | `072-S` / `079-F` | Unit 3.2 (`079.002-T`) |
| Map `*rpc.AssistantReasoningData` and `*rpc.AssistantReasoningDeltaData` in `normalizeWithGate` (`internal/agent/copilot/normalize.go`), wire per-adapter `unmappedLogGate` in `Adapter.onEvent` (`internal/agent/copilot/adapter.go`), and update `internal/agent/copilot/adapter_test.go` for mapped reasoning and per-event-type `Warn`-to-`Debug` rate limiting | `DA760FBB`, `E7B78125` | `072-S` / `079-F` | Unit 3.3 (`079.003-T`) |
| Add `KindAgentReasoning`, `KindAgentReasoningDelta`, `ToolEndedPayload.Text`, `KindFor` mappings, `mapEvent` `ev.Truncated` / `ToolEndedPayload.Text` propagation in `internal/hub/envelope.go`, and fold `"reasoning"` streaming/complete entries and `tool.ended` `Text` + `ev.Truncated` across all `Fold` branches (`applyDelta`, `applyMessage`, reasoning, `KindToolStarted`, `applyToolEnded`, and `f.lastDelta`) in `internal/hub/fold.go` | `DA760FBB` | `072-S` / `079-F` | Unit 3.4 (`079.004-T`) |
| Apply `h.fold.LastDelta()` to live `agent.KindReasoningDelta` frames and rate-limit repeated `hub: unmapped agent event` `Warn` logs per `ev.RawType` in `Hub.handleEvent` (`internal/hub/hub.go`), and render `thinking> ...` (`kindReasoning`) and `tool> <name> [<state>]: <text>` in `internal/tui/model.go` with live-versus-hydrated regression tests | `DA760FBB`, `E7B78125` | `072-S` / `079-F` | Unit 3.5 (`079.005-T`) |
| Amend normative §3.1, §4.2, §4.3 (`Event-union safety`), and §5.2 (`Event envelope`) in `docs/design-docs/intercom-go-backend-architecture.md` for `ForceStopBudget` / `startStopMux` contention, post-quiesce terminal `session.idle` / `session.shutdown` handling, once-per-event-type `warn` (then `debug`) logging, and `agent.reasoning`, `agent.reasoning.delta`, and `tool.ended` `text` envelopes | `407291BE`, `FA1C38F7(f)`, `E7B78125`, `DA760FBB` | `072-S` / `079-F` | Unit 3.6 (`079.006-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | PASS | All production code uses Go 1.25.0 (`GOTOOLCHAIN=go1.26.10`), explicit `apperr` / sentinel error wrapping, bounded waits, and zero `panic` calls in library packages. |
| II. Test-First Development (NON-NEGOTIABLE) | PASS | Every `code` implementation unit specifies red-phase failing test assertions in its colocated `_test.go` file before production code changes and verifies green via `go test -race ./...`. |
| III. Workspace Isolation and Security Boundaries | PASS | Zero new filesystem write calls or `pathsafe.Root.Resolve` call sites (`go run ./tools/gatecheck write-path --self-test --root .` stays 100% green). All SDK permission fields and event strings are sanitized and rune-bounded before crossing into the hub or TUI. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | PASS | All modifications are confined to the repository working tree. |
| V. Structured Observability | PASS | Unmapped SDK event types log their first occurrence per session at `slog.LevelWarn` and subsequent occurrences at `slog.LevelDebug` (`E7B78125`), preventing log floods while preserving diagnostic visibility; `session.shutdown` logs at `slog.LevelDebug`. |
| VI. Single Responsibility | PASS | Zero new external dependencies in `go.mod`. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | PASS | No destructive commands are used; consumed stash entries are archived via non-destructive `backlogit stash archive`. |
| VIII. Explicit Safety Modes | PASS | Freeze-scope mode active on the 7 chartered stash entries of Capability Set 1. |
| IX. Git-Friendly Persistence | PASS | Deliberation, plan, session memory, and backlog items use Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | PASS | Targeted code and backlog queries used throughout staging. |
| XI. Merge Commit History Preservation | PASS | All shipments merge via `--merge` commits (P-009). |

## Implementation Units

### Sub-Epic 1: Shipment `070-S` — Covering Feature `077-F` (Copilot adapter Stop/ForceStop resilience, latch/permission hardening, and `App.Err` stuck-lane reporting)

#### Unit 1.1 (`077.001-T`): Harden `sanitizeBound` (`Zl`/`Zp` replacement), cap and sanitize `PermissionRequest` `RawKind`/`ToolCallID`, and re-check `gone()` before sending permission asks (`FA1C38F7(a, b, d)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/agent/copilot/permission.go`
  * `internal/agent/copilot/permission_cycle2_test.go`
* **Changes needed**:
  * In `sanitizeBound` (`internal/agent/copilot/permission.go:63-84`): update the control/format rune branch to `case unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Cf, r):` so Unicode line separator (`U+2028`, `Zl`) and paragraph separator (`U+2029`, `Zp`) are replaced with `utf8.RuneError` (`U+FFFD`) while `\n` and `\t` remain preserved in multi-line summaries.
  * In `internal/agent/copilot/permission.go`: define `const maxIdentRunes = 128` and `func sanitizeIdent(s string, max int) string` which replaces `\n`, `\t`, `\u2028`, `\u2029`, `unicode.IsControl(r)`, and `unicode.Is(unicode.Cf, r)` with `utf8.RuneError` and bounds the string to at most `max` runes (appending `"…"` if clipped). In `toDomainRequest` (`internal/agent/copilot/permission.go:93-118`), capture `rawKind := string(req.Kind())`, use `rawKind` as `text` for `default:` (`PermissionKindOther`) so an oversized unknown kind still sets `out.Truncated = true` against `maxSummaryRunes`, and set `out.RawKind = sanitizeIdent(rawKind, maxIdentRunes)` and `out.ToolCallID = sanitizeIdent(out.ToolCallID, maxIdentRunes)`.
  * In `(*permissionBridge).handle` (`internal/agent/copilot/permission.go:246-254`): add a second `if gone() { return unavailable() }` check immediately before `select { case b.asks <- ask: ... }` so a stop, hub-done, or turn-abort that occurred while `toDomainRequest` ran is caught deterministically before `select` can pseudo-randomly pick a buffered `b.asks <- ask` case, and document that any simultaneous close during the send is still caught by the reply `select` and the `if d.Kind == agent.DecisionAllowOnce && gone()` downgrade.
  * Add unit tests in `internal/agent/copilot/permission_cycle2_test.go`:
    1. `TestSummaryAndIdentifiersSanitizeUnicodeSeparatorsAndCapRunes`: verify `sanitizeBound` replaces `U+2028` and `U+2029` with `U+FFFD` while preserving `\n` and `\t`, and `toDomainRequest` sanitizes control/`Cf`/`Zl`/`Zp`/newline runes in `RawKind` and `ToolCallID` and caps both at `maxIdentRunes` (`128` runes + `"…"`), while an oversized unmapped `RawKind` (`> maxSummaryRunes`) still sets `out.Truncated = true`.
    2. `TestPermissionBridgeRechecksGoneBeforeAskSend`: verify that when `b.stop` or `turnStop` is closed before `b.asks <- ask`, `handle` returns `UserNotAvailable` without enqueuing onto a buffered `b.asks` channel.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot -run "TestSummary|TestToDomainRequest|TestPermissionBridgeRechecksGoneBeforeAskSend"`

#### Unit 1.2 (`077.002-T`): Quiesce callback latch in `Adapter.Start` `fail()`, classify `session.shutdown` as informational, and preserve terminal `session.idle` arriving between quiesce and `Disconnect` (`FA1C38F7(c, f)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/copilot/adapter.go`
  * `internal/agent/copilot/normalize.go`
  * `internal/agent/copilot/adapter_cycle2_test.go`
* **Changes needed**:
  * In `(*Adapter).Start` `fail` (`internal/agent/copilot/adapter.go:258-271`): call `a.latch.quiesce(0)` immediately after `a.stopOnce.Do(func() { close(a.stop) })` so a failed `Start` (including one interrupted after `CreateSession` registered `a.onEvent`) closes the callback latch consistently with `Stop`.
  * In `informationalTypes` (`internal/agent/copilot/normalize.go:15-69`): add `"session.shutdown": {}` so the SDK's post-disconnect `session.shutdown` lifecycle event is dropped at `Debug` in `onEvent` before `a.latch.enter()`, eliminating the spurious `WARN` (`"agent event arrived after quiesce; discarded"`) on clean shutdowns.
  * In `(*Adapter).onEvent` (`internal/agent/copilot/adapter.go:384-419`): document the narrow post-quiesce terminal-idle exception in GoDoc on `onEvent` and `Stop`. When `!a.latch.enter()` returns false, check whether `_, ok := ev.Data.(*rpc.SessionIdleData); ok`. When `ok` is true, normalize `e := normalize(ev, a.log)`, update `a.bridge` (`if e.Aborted { if id := a.bridge.abortTurn(); e.TurnID == "" { e.TurnID = id } } else { a.bridge.clearTurn() }`), and if `a.cfg.HubDone` is not closed (`select { case <-a.cfg.HubDone: default: ... }`), attempt a non-blocking `select { case a.cfg.Ingress <- e: return default: }` before falling through to `a.log.Warn("agent event arrived after quiesce; discarded", "event_type", string(ev.Type()))`. Non-idle events arriving after quiesce continue to be discarded and logged at `Warn`.
  * Add unit tests in `internal/agent/copilot/adapter_cycle2_test.go`:
    1. `TestFailedStartQuiescesCallbackLatch`: verify that after a failed `Start`, `!h.ad.latch.enter()` holds (`a.latch` is closed).
    2. `TestSessionShutdownIsInformationalAndTerminalIdleAfterQuiesceIsDeliveredWhenIngressHasRoom`: verify that `isInformational("session.shutdown")` is true (and emitting a `session.shutdown` event after `Stop()` logs at `DEBUG`, never `WARN`), and that a terminal `rpc.SessionIdleData{Aborted: boolp(true)}` emitted after `Stop()` (while `HubDone` is open and `ingress` has room) updates `bridge` and is delivered non-blockingly to `ingress`, whereas a non-idle event (`AssistantTurnEndData`) after `Stop()` is still discarded and logged at `WARN`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot -run "TestFailedStartQuiescesCallbackLatch|TestSessionShutdownIsInformationalAndTerminalIdleAfterQuiesce|TestEventDiscardAfterQuiesceIsLoggedAtWarn"`

#### Unit 1.3 (`077.003-T`): Account for in-flight `Client.Stop()` `startStopMux` lock contention in `Adapter.Stop` / `ForceStop` and `ShutdownTimeouts` (`407291BE`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/copilot/adapter.go`
  * `internal/hub/shutdown.go`
  * `internal/agent/copilot/shutdown_hub_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/adapter.go`:
    * Export `const SDKRuntimeShutdownTimeout = 10 * time.Second`, `const SDKProcessExitTimeout = 10 * time.Second`, and `const DefaultSDKStopLockTimeout = SDKRuntimeShutdownTimeout + SDKProcessExitTimeout` (`20 * time.Second`), documenting `github.com/github/copilot-sdk/go@v1.0.11`'s internal `runtimeShutdownTimeout = 10s` and `killProcessAndWait()` `processExitTimeout = 10s` inside `(*copilotsdk.Client).Stop()`, during both of which `Client.Stop()` holds `startStopMux` (up to `20s` total) and blocks `Client.ForceStop()` after `ForceStop()` kills the OS process.
    * Add `func (realClient) isRealSDKClient() {}` on `realClient` (also inherited by `probeClient` in `live_sq_test.go`). Add `SDKStopLockTimeout time.Duration` to `Config`: in `newAdapter(cfg Config, c sdkClient)`, if `cfg.SDKStopLockTimeout < 0`, set `cfg.SDKStopLockTimeout = 0`; else if `cfg.SDKStopLockTimeout == 0`, set `cfg.SDKStopLockTimeout = DefaultSDKStopLockTimeout` whenever `_, ok := c.(interface{ isRealSDKClient() }); ok` — **independently of `cfg.StopTimeout`** (so a real adapter with `StopTimeout < DefaultStopTimeout`, such as `StopTimeout = 1s`, still receives the `20s` SDK lock-contention budget, while injected `newWithClient` fake clients default `SDKStopLockTimeout` to `0` unless explicitly set by a test).
    * Add `clientStopStart time.Time` and `clientStopActive bool` (guarded by `a.mu`) on `Adapter`. In `(*Adapter).doStop`, when `clientStarted` is true, wrap the `a.client.Stop` function passed to `step("stop copilot client", ...)` so it sets `a.mu.Lock(); a.clientStopStart = time.Now(); a.clientStopActive = true; a.mu.Unlock()` on entry and `a.mu.Lock(); a.clientStopActive = false; a.mu.Unlock()` via `defer` when `a.client.Stop()` returns.
    * Add `func (a *Adapter) forceStopWait() time.Duration`: starts with `w := a.cfg.StopTimeout`; reads `active, start := a.clientStopActive, a.clientStopStart` under `a.mu`; when `active && a.cfg.SDKStopLockTimeout > 0`, adds `max(a.cfg.SDKStopLockTimeout-time.Since(start), 0)` to `w`. In `(*Adapter).forceStop()`, use `wait := a.forceStopWait()` for both `time.NewTimer(wait)` and the `fmt.Errorf("agent: %w: ForceStop did not return within %v", agent.ErrCleanupIncomplete, wait)` error message.
    * Export `func (a *Adapter) ForceStopBudget() time.Duration { return a.cfg.StopTimeout + a.cfg.SDKStopLockTimeout }` so `hub.NewShutdown` can reserve the adapter's worst-case lock-contended `ForceStop` wait (`25s` by default on a real adapter, `21s` on a real adapter with `StopTimeout = 1s`, or `a.cfg.StopTimeout` on a fake client when `SDKStopLockTimeout == 0`) in the post-graceful-cutoff `Force` budget.
  * In `internal/hub/shutdown.go`:
    * Update GoDoc on `ShutdownTimeouts.Force` and `ShutdownConfig.AdapterStopTimeout` to document that because preceding graceful steps inside `Agent.Stop(sc)` (in-flight `Start` cancellation, callback latch quiescing, and `sess.Disconnect`) can consume nearly the entire `t.Stop` graceful budget before `Client.Stop()` starts, `a.forceStop()` can still wait up to the full `ForceStopBudget()` (`StopTimeout + SDKStopLockTimeout`) *after* the graceful `t.Stop` cutoff expires; therefore `Timeouts.Force` alone (not `Stop + Force`) must be at least `ForceStopBudget()`.
    * In `NewShutdown`: after checking `StopTimeout()` when `cfg.AdapterStopTimeout == 0`, query `if fb, ok := cfg.Agent.(interface{ ForceStopBudget() time.Duration }); ok { b := fb.ForceStopBudget(); if b < 0 { return nil, apperr.New(apperr.KindConfig, "hub: agent ForceStopBudget must not be negative") }; cfg.AdapterStopTimeout = max(cfg.AdapterStopTimeout, b) }`. Then existing validation (`cfg.AdapterStopTimeout > 0 && cfg.Timeouts.Force > 0 && cfg.Timeouts.Force < cfg.AdapterStopTimeout`) rejects any custom `Timeouts.Force < ForceStopBudget()` with an `apperr.KindConfig` error even when `Stop + Force >= ForceStopBudget()`, and default population (`if cfg.AdapterStopTimeout > 0 && cfg.Timeouts.Force == 0 { cfg.Timeouts.Force = max(DefaultShutdownForce, cfg.AdapterStopTimeout+2*time.Second) }`) sets `Timeouts.Force = 27s` (`25s + 2s`, total `Stop + Force = 37s`) for the default real Copilot adapter while preserving `DefaultShutdownForce = 7 * time.Second` for agents without a larger `ForceStopBudget()`.
  * Add unit tests in `internal/agent/copilot/shutdown_hub_test.go`:
    1. `TestForceStopExtendsWaitForInFlightClientStopHoldingSDKLock`: simulate `copilotsdk.Client`'s two-phase `startStopMux` lock contention in `fakeClient` (with `StopTimeout = 100 * time.Millisecond` and `SDKStopLockTimeout = 300 * time.Millisecond`, where `client.Stop` holds `startStopMux` across both a simulated `runtimeShutdownTimeout` phase and a delayed-reaping `killProcessAndWait` `processExitTimeout` phase totaling `260ms > 2*StopTimeout`, and `client.ForceStop` immediately sets `childAlive = false` and then locks `startStopMux` before returning). Verify:
       * (a) **Timed-out + delayed-reaping lock hold**: when `client.Stop` times out at `100ms` and continues holding `startStopMux` for another `160ms` of simulated `runtimeShutdownTimeout + processExitTimeout` reaping while `forceStop` runs, `forceStopWait()` extends the wait by the remaining `SDKStopLockTimeout` window so `ForceStop` completes cleanly without `agent.ErrCleanupIncomplete` or `hub.ErrOrphanRisk`;
       * (b) **Hurried lock hold**: when `ctx` is hurried `20ms` into `client.Stop` (leaving `240ms` of two-phase lock hold while `forceStop` runs), `forceStopWait()` extends the wait so `ForceStop` completes cleanly;
       * (c) **Delayed-start regression (graceful cleanup consumes `t.Stop` before `Client.Stop` begins)**: when `sess.Disconnect` consumes `150ms` of a `t.Stop = 200ms` graceful budget before `client.Stop` starts at `t = 150ms` and holds `startStopMux` for `260ms` (until `t = 410ms`, `210ms` past the `t.Stop` cutoff), `hub.Shutdown` with default `Force` (`max(DefaultShutdownForce, ForceStopBudget()+2s)` or `Force >= ForceStopBudget() = 400ms`) waits out the entire post-cutoff `forceStopWait()` and finishes with no `hub.ErrOrphanRisk`;
       * (d) **Real-client `StopTimeout < DefaultStopTimeout` lock allowance & `NewShutdown` `Force < ForceStopBudget()` rejection**: verify that a real adapter constructed via `New(Config{..., StopTimeout: time.Second})` defaults `SDKStopLockTimeout` to `DefaultSDKStopLockTimeout` (`20s`) independently of `StopTimeout` (`ForceStopBudget() == 21s`), whereas `newWithClient` fake-client adapters default `SDKStopLockTimeout` to `0` unless explicitly set, and `hub.NewShutdown` rejects custom `Timeouts.Force < ForceStopBudget()` (e.g. `Stop: 10s, Force: 5s` or `Stop: 500ms, Force: 200ms` when `ForceStopBudget() == 400ms`) with an `apperr.KindConfig` error even when `Stop + Force >= ForceStopBudget()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot ./internal/hub -run "TestForceStop|TestStop|TestShutdown|TestDefaultForceBudget|TestNewShutdown"`

#### Unit 1.4 (`077.004-T`): Record `Hub.Run`'s `ErrCleanupIncomplete` in `App.Err` via `a.addErr` before `hubRunDone` closes (`70BEBF29`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/app/assemble.go`
  * `internal/app/shutdown_test.go`
* **Changes needed**:
  * In `(*App).assemble` (`internal/app/assemble.go:223-228`): inside the `go func()` that runs `h.Run(hubCtx)` with `defer close(a.hubRunDone)`, when `err := h.Run(hubCtx); err != nil`, call `a.addErr(fmt.Errorf("hub: %w", err))` after `a.log.Error("hub stopped with error", "error", err)` so the error is recorded under `a.mu` before `a.hubRunDone` closes and is joined into `a.shut.err` (`App.Err()`) by `supervise()` (and into `teardown` in `internal/app/assemble.go:426` via `errs = append(a.takeErrs(), errs...)`).
  * Add unit test in `internal/app/shutdown_test.go`:
    1. `TestShutdownRecordsHubRunErrCleanupIncompleteInAppErr`: start an `App` whose fake agent's `Send` blocks on an unclosed `releaseSend` channel (ignoring `ctx.Done()`) after signalling `sendEntered`, submit a prompt through a connected client so the hub's operator prompt lane worker is blocked inside `Agent.Send`, trigger `a.Shutdown("stuck operator lane")`, wait for `<-a.Done()`, close `releaseSend` before `assertNoLeaks(t, baseline)`, and assert `errors.Is(a.Err(), agent.ErrCleanupIncomplete)`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestShutdownRecordsHubRunErrCleanupIncompleteInAppErr|TestSIGINT"`

### Sub-Epic 2: Shipment `071-S` — Covering Feature `078-F` (Remote `Cursor`/`wsclient` terminal `server.shutdown` gap handling and TUI permission modal `kind:`/`tool call:` continuation anti-spoofing)

#### Unit 2.1 (`078.001-T`): Treat `KindServerShutdown` as always-applicable across sequence gaps in `hub.Cursor` and `wsclient` (`80996C79`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/cursor.go`
  * `internal/transport/wsclient/client.go`
  * `internal/transport/wsclient/client_test.go`
* **Changes needed**:
  * In `(*Cursor).Apply(f Frame) Action` (`internal/hub/cursor.go:55-76`): add a branch `if f.Kind() == KindServerShutdown`:
    * If `c.have && f.Gen() == c.pos.Gen && f.Seq() <= c.pos.Seq`: return `ActionDiscard` (duplicate terminal frame).
    * If `c.have && !c.stale && f.Gen() == c.pos.Gen && f.Seq() == c.pos.Seq+1`: set `c.pos.Seq = f.Seq()` and return `ActionApply` (contiguous terminal frame keeps `c.stale == false`).
    * Otherwise (sequence gap `f.Seq() > c.pos.Seq+1` after `deliverTerminal` evicted the oldest frame from a full buffer, or `c.stale` already true, or `!c.have` / generation mismatch): set `c.pos = Position{Gen: f.Gen(), Seq: f.Seq()}`, `c.have = true`, `c.stale = true` (so `Position()` returns `false` and never claims skipped sequence numbers were applied, while a duplicate of the same `server.shutdown` sees `f.Seq() <= c.pos.Seq` and returns `ActionDiscard`), and return `ActionApply`.
  * In `(*cursor).apply(e hub.Envelope) Verdict` (`internal/transport/wsclient/client.go:399-420`): add the matching `if e.Kind == hub.KindServerShutdown` branch returning `VerdictDiscard` on `c.have && e.Gen == c.gen && e.Seq <= c.seq`, advancing `c.seq = e.Seq` with `c.stale == false` and returning `VerdictApply` when contiguous (`c.have && !c.stale && e.Gen == c.gen && e.Seq == c.seq+1`), and setting `c.gen, c.seq, c.have, c.stale = e.Gen, e.Seq, true, true` and returning `VerdictApply` otherwise.
  * Add unit tests in `internal/transport/wsclient/client_test.go`:
    1. `TestCursorTreatsServerShutdownAsAlwaysApplicableAcrossGaps`: verify for both `hub.Cursor` and `wsclient` `cursor` that (a) an in-sequence `KindServerShutdown` returns `ActionApply` / `VerdictApply` and keeps `Position()` valid, (b) a `KindServerShutdown` arriving across a sequence gap (`seq=15` after `seq=11`), while `stale == true`, or before any baseline (`!have`) returns `ActionApply` / `VerdictApply` while marking `Position()` unavailable (`ok == false`), and (c) a duplicate delivery of the same `KindServerShutdown` (`seq=15`) returns `ActionDiscard` / `VerdictDiscard`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub ./internal/transport/wsclient -run "TestCursor"`

#### Unit 2.2 (`078.002-T`): Indent soft-wrap continuation lines of `kind:` and `tool call:` in `tui.Model.permissionBody` (`A999ABD6`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/tui/permission.go`
  * `internal/tui/permission_test.go`
* **Changes needed**:
  * In `internal/tui/permission.go`: add `(m Model) wrapHeader(prefix, val string, inner int) string` that wraps `prefix + val` to `max(inner-2, 1)` via `m.wrap(prefix+val, max(inner-2, 1))`, splits on `"\n"`, and prefixes every continuation line (`for i := 1; i < len(lines); i++ { lines[i] = "  " + lines[i] }`) with two spaces while leaving `lines[0]` at column 0. Update `(m Model) permissionBody(p permission) string` (`internal/tui/permission.go:159-166`) to call `m.wrapHeader("kind: ", p.Kind, inner)` and `m.wrapHeader("tool call: ", p.ToolCallID, inner)` so single-line `kind:` and `tool call:` headers render identically at column 0 (preserving `internal/tui/testdata/modal.golden`), while any soft-wrapped continuation line is indented by two spaces within `inner` columns (`2 + (inner-2) == inner`) and cannot spoof a column-0 modal header.
  * Add unit test in `internal/tui/permission_test.go`:
    1. `TestPermissionModalIndentsWrappedKindAndToolCallContinuations`: construct a permission request with long `PermissionKind` and `ToolCallID` values containing spaces and spoofed header text (e.g., `"shell             kind: read"` and `"call-1234567890   kind: read"` in a narrow width `w = 30`), assert that in `m.permissionBody` only the first line of `"kind: ..."` and the first line of `"tool call: ..."` start at column 0 while every continuation line starts with `"  "`, and verify the wrapped continuation lines feed into `m.modal.TotalLineCount()` and `m.approvable()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestPermissionModalIndentsWrappedKindAndToolCallContinuations|TestModalShowsRequestAndApproveGoesThroughIngress"`

### Sub-Epic 3: Shipment `072-S` — Covering Feature `079-F` (`agent.Event` rune-safe bounds, `tool.ended` result/error & reasoning enrichment, write-diff/shell-URL surfacing, and unmapped-event log rate-limiting)

#### Unit 3.1 (`079.001-T`): Surface write diff/new-file-content and shell `possibleUrls` in `PermissionRequest.Summary` (`FA1C38F7(e)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/agent/copilot/permission.go`
  * `internal/agent/copilot/permission_cycle2_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/permission.go`:
    * Add `const writePreviewTruncatedSuffix = "\n[write preview truncated]"`.
    * Add `func shellSummary(r *rpc.PermissionRequestShell) string`: starts with `r.FullCommandText`, collects non-empty `u.URL` strings from `r.PossibleURLs`, and when at least one non-empty URL exists, appends `"\nurls: " + strings.Join(urls, ", ")`. Pass `shellSummary(r)` as `text` in `toDomainRequest` so `sanitizeBound(text, maxSummaryRunes)` bounds the combined shell command and URLs and sets `out.Truncated = true` (auto-denied) if the total exceeds `maxSummaryRunes`.
    * In `toDomainRequest` for `*rpc.PermissionRequestWrite`: run `fnSan, fnTrunc, fnLen := sanitizeBound(r.FileName, maxSummaryRunes)` on `r.FileName` first (preserving fail-closed auto-deny if `FileName` itself exceeds `maxSummaryRunes`). When `!fnTrunc`, select `body := r.Diff` (falling back to `deref(r.NewFileContents)` when `body == ""`); if `body == ""`, set `out.Summary, out.Truncated, out.FullLen = fnSan, false, fnLen`. When `body != ""`: compute `avail := maxSummaryRunes - fnLen - 1` and `bodySan, bodyClipped, bodyLen := sanitizeBound(body, max(avail, 0))`. If `!bodyClipped && avail >= bodyLen`, set `out.Summary = fnSan + "\n" + bodySan`, `out.Truncated = false`, and `out.FullLen = fnLen + 1 + bodyLen`. If `bodyClipped || avail < bodyLen` (the write diff/content preview was clipped to fit `maxSummaryRunes`), reserve `suffixRunes := utf8.RuneCountInString(writePreviewTruncatedSuffix)` (`26` runes): if `fnLen + suffixRunes > maxSummaryRunes` (so `FileName` leaves no room for the truncation notice within `maxSummaryRunes`), set `out.Summary, out.Truncated, out.FullLen = fnSan, true, fnLen+1+bodyLen` (fail-closed auto-deny); otherwise compute `previewBudget := max(maxSummaryRunes-fnLen-1-suffixRunes-1, 0)` (reserving 1 rune for `sanitizeBound`'s `"…"` and `suffixRunes` for `writePreviewTruncatedSuffix`), set `preview, _, _ := sanitizeBound(body, previewBudget)`, and set `out.Summary = fnSan + "\n" + preview + writePreviewTruncatedSuffix` (or `fnSan + writePreviewTruncatedSuffix` when `previewBudget == 0`), `out.Truncated = false`, and `out.FullLen = fnLen + 1 + bodyLen`, guaranteeing `utf8.RuneCountInString(out.Summary) <= maxSummaryRunes` (`2048` runes) and an explicit `[write preview truncated]` notice whenever a write preview is partial.
  * Add unit tests in `internal/agent/copilot/permission_cycle2_test.go`:
    1. `TestToDomainRequestSurfacesShellPossibleURLsAndWriteDiffOrContent`: verify (a) `PermissionRequestShell` with `PossibleURLs` appends `\nurls: ...` to `Summary` (and without `PossibleURLs` leaves `Summary == FullCommandText`), (b) `PermissionRequestWrite` with `Diff` appends the sanitized diff after `FileName\n` and with `NewFileContents` (when `Diff == ""`) appends the sanitized new file contents, and (c) a large `Diff` (`> maxSummaryRunes`) is clipped to `<= maxSummaryRunes` total runes with `"\n[write preview truncated]"` while keeping `Truncated == false` so `permissionBridge.handle` forwards the request to the hub for operator review, whereas an oversized `FileName` (`> maxSummaryRunes`, or `2023..2048` runes with a non-empty `Diff` that cannot fit the notice) sets `Truncated == true` and is denied without asking the hub.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot -run "TestToDomainRequest|TestPermissionSummaryIsBounded|TestTruncatedShellCommandIsDeniedWithoutAskingHub"`

#### Unit 3.2 (`079.002-T`): Bound and enrich `agent.Event` (`KindReasoning`, `KindReasoningDelta`, `ToolEnded` result/error `Text`, rune-safe `Truncated` caps) and rate-limit unmapped SDK event warn logs in `normalize.go` (`DA760FBB`, `E7B78125`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/agent.go`
  * `internal/agent/copilot/normalize.go`
  * `internal/agent/copilot/normalize_test.go`
* **Changes needed**:
  * In `internal/agent/agent.go`:
    * Add `KindReasoning EventKind = "agent.reasoning"` and `KindReasoningDelta EventKind = "agent.reasoning.delta"` to `EventKind`.
    * Add `Truncated bool` to `agent.Event` and update field comments on `MessageID` and `Text` to document reasoning IDs/content and `tool.ended` result/error text.
  * In `internal/agent/copilot/normalize.go`:
    * Export `const MaxEventTextBytes = 16 << 10` (`16384`) and `const MaxEventIdentBytes = 128`, plus `const maxWarnedEventTypes = 256`.
    * Add `func boundUTF8(s string, maxBytes int) (string, bool)` that in a single pass replaces invalid UTF-8 sequences with `utf8.RuneError` (`U+FFFD`, 3 bytes) and stops copying as soon as adding the next rune's encoded byte length would exceed `maxBytes`, returning `(strings.Clone(out), truncated)` without allocating proportionally to an oversized input.
    * Add `type unmappedLogGate struct { mu sync.Mutex; seen map[string]struct{} }` with `func (g *unmappedLogGate) shouldWarn(raw string) bool` that lazily initializes `if g.seen == nil { g.seen = make(map[string]struct{}) }` under `g.mu`, returns `true` and records `raw` the first time `raw` is seen (while `len(g.seen) < maxWarnedEventTypes`), and returns `false` on subsequent occurrences of `raw` (or once the cap is reached).
    * Add `func normalizeWithGate(ev rpc.SessionEvent, log *slog.Logger, gate *unmappedLogGate) agent.Event` (with `normalize(ev, log)` calling `normalizeWithGate(ev, log, nil)`):
      * In `*rpc.ToolExecutionCompleteData`: extract result/error text (`if !d.Success && d.Error != nil && d.Error.Message != ""` -> `d.Error.Message`; else if `d.Result != nil` -> `d.Result.Content`, falling back to `deref(d.Result.DetailedContent)` when `Content == ""`; else if `d.Error != nil` -> `d.Error.Message`), bound `ToolCallID` (`MaxEventIdentBytes`) and `Text` (`MaxEventTextBytes`), and set `Truncated`.
      * Bound all other event fields (`MessageID`, `TurnID`, `ToolCallID`, `ToolName` at `MaxEventIdentBytes`; `Text` at `MaxEventTextBytes`) via `boundUTF8` and set `Event.Truncated`.
      * In `default:`: sanitize and bound `raw, rawT := boundUTF8(sanitizeIdent(string(ev.Type()), MaxEventIdentBytes), MaxEventIdentBytes)` before calling `gate.shouldWarn(raw)` and logging; when `gate != nil && !gate.shouldWarn(raw)`, log `log.Debug("unhandled agent SDK event type (repeated)", "event_type", raw)`; otherwise log `log.Warn("unhandled agent SDK event type", "event_type", raw)`; return `agent.Event{Kind: agent.KindUnknown, RawType: raw, Truncated: rawT}`. (The `*rpc.AssistantReasoningData` and `*rpc.AssistantReasoningDeltaData` switch cases are added in Unit 3.3 together with `adapter_test.go` so `go test ./internal/agent/...` stays 100% green after Unit 3.2.)
  * In `internal/agent/copilot/normalize_test.go`:
    * Update `TestNormalizeUnknownVariantLogsWarnAndIsExplicit` and `TestNormalizeNilLoggerDoesNotPanic` to use `&rpc.SessionErrorData{}`.
    * Add unit tests in `internal/agent/copilot/normalize_test.go` for `ToolExecutionCompleteData` result/error text extraction, single-pass rune-safe `MaxEventTextBytes` and `MaxEventIdentBytes` capping (`Truncated: true` without splitting multi-byte UTF-8 runes or overflowing `maxBytes` on invalid UTF-8 replacement), and `normalizeWithGate` logging the first occurrence of each bounded unmapped event type at `WARN` and subsequent occurrences at `DEBUG`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/... -run "TestNormalize|TestIsInformational|TestCallback"`

#### Unit 3.3 (`079.003-T`): Map `AssistantReasoningData` / `AssistantReasoningDeltaData` in `normalizeWithGate`, wire per-adapter `unmappedLogGate` in `Adapter.onEvent`, and update `adapter_test.go` (`DA760FBB`, `E7B78125`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/copilot/normalize.go`
  * `internal/agent/copilot/adapter.go`
  * `internal/agent/copilot/adapter_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/normalize.go`: add `case *rpc.AssistantReasoningData:` (mapping to `agent.KindReasoning` with `MessageID: d.ReasoningID` bounded by `MaxEventIdentBytes`, `Text: d.Content` bounded by `MaxEventTextBytes`, and `Truncated: tID || tText`) and `case *rpc.AssistantReasoningDeltaData:` (mapping to `agent.KindReasoningDelta` with `MessageID: d.ReasoningID` bounded by `MaxEventIdentBytes`, `Text: d.DeltaContent` bounded by `MaxEventTextBytes`, and `Truncated: tID || tText`) in `normalizeWithGate`.
  * In `internal/agent/copilot/adapter.go`: add `unmapped unmappedLogGate` on `Adapter` and update `(*Adapter).onEvent` (`internal/agent/copilot/adapter.go:394`) to call `e := normalizeWithGate(ev, a.log, &a.unmapped)` instead of `normalize(ev, a.log)`.
  * In `internal/agent/copilot/adapter_test.go`: update `TestCallbackUnknownEventIsForwardedNotDropped` (`internal/agent/copilot/adapter_test.go:378`) to emit `&rpc.SessionErrorData{}` instead of `&rpc.AssistantReasoningData{}`, and add `TestCallbackMapsReasoningAndRateLimitsRepeatedUnmappedEventWarnLogs` verifying that `onEvent` forwards `AssistantReasoningData` and `AssistantReasoningDeltaData` as `agent.KindReasoning` and `agent.KindReasoningDelta` with zero warn logs, and that emitting three `&rpc.SessionErrorData{}` events logs `level=WARN` on the first event and `level=DEBUG` on the second and third events while still forwarding all three `agent.KindUnknown` events to `h.ingress`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/... -run "TestCallback|TestNormalize"`

#### Unit 3.4 (`079.004-T`): Map and fold `agent.reasoning`, `agent.reasoning.delta`, `tool.ended` result/error `Text`, and `Event.Truncated` in `internal/hub/envelope.go` and `internal/hub/fold.go` (`DA760FBB`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/envelope.go`
  * `internal/hub/fold.go`
  * `internal/hub/fold_test.go`
* **Changes needed**:
  * In `internal/hub/envelope.go`:
    * Add `KindAgentReasoning Kind = "agent.reasoning"` and `KindAgentReasoningDelta Kind = "agent.reasoning.delta"` (using `MessagePayload{MessageID, Text, Truncated}`).
    * Add `Text string` (`json:"text,omitempty"`) to `ToolEndedPayload`.
    * In `KindFor(k agent.EventKind) (Kind, bool)`: map `agent.KindReasoning -> KindAgentReasoning, true` and `agent.KindReasoningDelta -> KindAgentReasoningDelta, true`.
    * In `mapEvent(ev agent.Event, textCap int)`: handle `KindAgentReasoning` and `KindAgentReasoningDelta` alongside `KindAgentMessage` and `KindAgentMessageDelta`, include `text := truncUTF8(ev.Text, textCap)` (`Text: text`, `Truncated: ev.Truncated || t1 || t2 || len(text) < len(ev.Text)`) in `ToolEndedPayload`, and propagate `ev.Truncated` into `Truncated` across `MessagePayload`, `TurnPayload`, `ToolStartedPayload`, and `ToolEndedPayload`.
  * In `internal/hub/fold.go`:
    * In `(*Fold).Apply(ev agent.Event)`: when `ev.Truncated` is true on mapped event kinds, set `f.anyTrunc = true`; add `case agent.KindReasoningDelta:` and `case agent.KindReasoning:` sharing `applyTextDelta(ev, entryKind)` and `applyTextComplete(ev, entryKind)` for `"message"` and `"reasoning"` entries; in `applyTextDelta`, set `f.lastDelta = deltaResult{text: add, truncated: clipped || ev.Truncated || idT}`, guard early-return with `if add == "" && (cur.Truncated || (!clipped && !ev.Truncated && !idT)) { return }`, set `cur.Truncated = cur.Truncated || clipped || ev.Truncated || idT` on existing streaming entries, and set `e.Truncated = ev.Truncated || idT || txtT` (`f.lastDelta = deltaResult{text: e.Text, truncated: e.Truncated}`) on new streaming entries; in `applyTextComplete`, set `e.Truncated = ev.Truncated || idT || txtT`; in `KindToolStarted`, set `Truncated: ev.Truncated || t1 || t2`; in `(*Fold).applyToolEnded(ev)`, bound `ev.Text` via `f.capText(ev.Text)` when non-empty, set `if text != "" { cur.Text = text }` (or `entry.Text = text` when pushing a new tool entry), and set `Truncated: cur.Truncated || ev.Truncated || t1 || t2 || txtT`.
  * Add unit tests in `internal/hub/fold_test.go`:
    1. `TestFoldAndMapEventReasoningAndToolEndedTextAndTruncation`: verify `KindFor` and `mapEvent` map `agent.KindReasoning`, `agent.KindReasoningDelta`, `agent.KindToolEnded` `Text` (both within cap and clipped at `textCap`), and upstream `ev.Truncated: true`, and verify `Fold.Apply` folds streaming and complete `"reasoning"` entries (without colliding with `"message"` entries sharing the same `ID`), records `tool.ended` `Text` on both running and standalone tool entries, and propagates both upstream `ev.Truncated: true` (when `len(ev.Text) <= MaxEntryTextBytes`) and hub-side `MaxEntryTextBytes` clipping into `cur.Truncated`, `e.Truncated`, `f.LastDelta()`, and snapshot `truncated`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestFold|TestEnvelope|TestKindFor|TestMapEvent"`

#### Unit 3.5 (`079.005-T`): Publish bounded live `agent.reasoning.delta`, rate-limit repeated `hub: unmapped agent event` warn logs in `Hub.handleEvent`, and render reasoning and `tool.ended` result/error text in `tui.Model` (`DA760FBB`, `E7B78125`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/hub.go`
  * `internal/tui/model.go`
  * `internal/tui/model_test.go`
* **Changes needed**:
  * In `internal/hub/hub.go`:
    * Add `unmappedWarned map[string]struct{}` (initialized in `New`, capped at `256` entries) on `Hub`. In `(*Hub).handleEvent(ev agent.Event)` (`internal/hub/hub.go:639-646`), when `_, ok := KindFor(ev.Kind); !ok`, bound `key, _ := capIdent(ev.RawType)`; if `_, seen := h.unmappedWarned[key]; !seen && len(h.unmappedWarned) < 256`, record `h.unmappedWarned[key] = struct{}{}` and log `h.cfg.Logger.Warn("hub: unmapped agent event", "code", ErrCodeUnknownEvent, "raw_type_bytes", len(ev.RawType))`, otherwise log `h.cfg.Logger.Debug("hub: unmapped agent event (repeated)", "code", ErrCodeUnknownEvent, "raw_type_bytes", len(ev.RawType))`, while still folding and publishing `KindError` (`unknown_event`).
    * In `(*Hub).handleEvent(ev agent.Event)` (`internal/hub/hub.go:663`): update `if ev.Kind == agent.KindDelta {` to `if ev.Kind == agent.KindDelta || ev.Kind == agent.KindReasoningDelta {` so live `KindAgentReasoningDelta` envelopes carry `h.fold.LastDelta()` (`text, clipped := h.fold.LastDelta(); p.Text, p.Truncated = text, p.Truncated || clipped`), maintaining live-versus-hydrated parity for reasoning deltas.
  * In `internal/tui/model.go`:
    * Add `const kindReasoning = "reasoning"`.
    * In `onFrame`: handle `case hub.KindAgentReasoningDelta:` and `case hub.KindAgentReasoning:` via `m.applyTextEntry(env.Payload, kindReasoning, isDelta)` (with `applyMessage` delegating to `m.applyTextEntry(raw, kindMessage, isDelta)`).
    * In `applyToolEnded`: sanitize `text := Sanitize(p.Text)`, set `if text != "" { m.entries[i].Text = text }` when updating a running tool entry, and set `Text: text` when appending a standalone tool entry.
    * In `entryText(e entry)`: for `case kindTool:`, when `e.Text != ""`, append `": " + e.Text` after `fmt.Sprintf("tool> %s [%s]", e.Tool, state)` (leaving empty-text tool entries unchanged so existing golden files match byte-for-byte); add `case kindReasoning:` rendering `"thinking> " + e.Text` (plus `" [streaming]"` when `e.Status == statusStreaming`).
  * Add unit tests in `internal/tui/model_test.go`:
    1. `TestReasoningAndToolEndedResultTextLiveAndHydratedParity`: drive streaming `KindReasoningDelta`, complete `KindReasoning`, and `KindToolEnded` (carrying result/error `Text` and `Truncated: true`) through a live `hub.Hub` into both a live `tui.Model` and a freshly hydrated `tui.Model`, asserting identical `entries` (`thinking> ...` and `tool> bash [ok]: ...`) and identical rendered views, and verify repeated `agent.KindUnknown` events through `Hub` log at `WARN` once per `RawType` and at `DEBUG` on repeats.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub/... ./internal/tui/... -run "TestReasoningAndToolEndedResultTextLiveAndHydratedParity|TestGolden"`

#### Unit 3.6 (`079.006-T`): Amend normative §3.1, §4.2, §4.3, and §5.2 in `docs/design-docs/intercom-go-backend-architecture.md` (`407291BE`, `FA1C38F7(f)`, `E7B78125`, `DA760FBB`)

* **Domain**: `docs` (`size: S`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (1 file):
  * `docs/design-docs/intercom-go-backend-architecture.md`
* **Changes needed**:
  * In `docs/design-docs/intercom-go-backend-architecture.md` §3.1 (step 4): document that `Adapter.ForceStopBudget()` (`StopTimeout + SDKStopLockTimeout`) accounts for the Copilot SDK's internal `runtimeShutdownTimeout` (`10s`) plus `killProcessAndWait()` `processExitTimeout` (`10s`, `DefaultSDKStopLockTimeout = 20s` total) `startStopMux` lock-hold window when an in-flight `Client.Stop()` precedes `Client.ForceStop()`, and that `hub.NewShutdown` reserves `ForceStopBudget()` in the post-graceful-cutoff `Timeouts.Force` budget (`Force >= ForceStopBudget()`, defaulting `Force = max(DefaultShutdownForce, ForceStopBudget()+2s)` = `27s`).
  * In `docs/design-docs/intercom-go-backend-architecture.md` §4.2 (`SDK callback contract`): document that `session.shutdown` is classified as informational (`debug`) and that a terminal `session.idle` arriving between `latch.quiesce` and `sess.Disconnect` updates `permissionBridge` turn state and is delivered non-blockingly (`select`/`default`) to `ingress` when `HubDone` is open.
  * In `docs/design-docs/intercom-go-backend-architecture.md` §4.3 (`Event-union safety (normative)`): amend the logging rule to specify that the `default:` branch logs the first occurrence of each unhandled `event.Type()` per adapter session (and per `RawType` in `Hub.handleEvent`) at `warn` level and subsequent occurrences of the same type at `debug` level (with a bounded seen-type set of 256 types) so high-frequency unmapped events do not flood operator logs while still emitting `agent.KindUnknown` (never silently dropping an unknown event).
  * In `docs/design-docs/intercom-go-backend-architecture.md` §5.2 (`Event envelope`): document `agent.reasoning` and `agent.reasoning.delta` alongside `agent.message` and `agent.message.delta`, and note the optional `text` field on `tool.ended` carrying bounded tool completion result or error text.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`

## Dependency Graph

### Shipment-Level DAG

* `070-S` (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`, under explicit operator P-001 policy exception over `062-S`, code baseline `069-S` `5f561d2` / `078713b`; Covering Feature `077-F`: `077.001-T -> 077.002-T -> 077.003-T -> 077.004-T`)
  * `<-blocks-- 071-S` (Covering Feature `078-F`: `078.001-T -> 078.002-T`)
    * `<-blocks-- 072-S` (Covering Feature `079-F`: `079.001-T -> 079.002-T -> 079.003-T -> 079.004-T -> 079.005-T -> 079.006-T`)

### Intra-Shipment Task Dependencies

* **Within `070-S` (`077-F`)**: `077.001-T` -> `077.002-T` -> `077.003-T` -> `077.004-T` (serialized across `internal/agent/copilot/adapter.go` and `internal/app/shutdown_test.go`).
* **Within `071-S` (`078-F`)**: `078.001-T` -> `078.002-T`.
* **Within `072-S` (`079-F`)**: `079.001-T` -> `079.002-T` -> `079.003-T` -> `079.004-T` -> `079.005-T` -> `079.006-T` (serialized because `079.003-T` and `079.004-T` consume `agent.KindReasoning`, `agent.KindReasoningDelta`, `Event.Truncated`, and `normalizeWithGate` from `079.002-T`, and `079.005-T` consumes `KindAgentReasoning` / `KindAgentReasoningDelta` / `ToolEndedPayload.Text` from `079.004-T`).

## Decisions and Rationale

1. **Three single-feature shipments (`070-S..072-S`)**: Satisfies Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and separates shutdown/lifecycle reliability fixes (`070-S`) and remote/TUI disconnect/modal bug fixes (`071-S`) from event-telemetry enrichment (`072-S`).
2. **SDK `startStopMux` contention-aware `forceStopWait()` and post-cutoff `Force` reservation (`407291BE`)**: Because `copilotsdk.Client.Stop()` (`v1.0.11`) holds `c.startStopMux.Lock()` across both `runtimeShutdownTimeout = 10s` and `killProcessAndWait()` `processExitTimeout = 10s` (up to `20s` total) while `copilotsdk.Client.ForceStop()` kills `osProcess` immediately and then waits on `c.startStopMux.Lock()`, extending `forceStopWait()` by `max(SDKStopLockTimeout-time.Since(clientStopStart), 0)` when `clientStopActive` is true guarantees `ForceStop()` always has a full `StopTimeout` (`5s`) margin after `Client.Stop()` releases `startStopMux`. Defaulting `SDKStopLockTimeout = DefaultSDKStopLockTimeout` (`20s`) on real SDK clients (`isRealSDKClient()`) independently of `StopTimeout` protects real adapters even when `StopTimeout < 5s` while keeping injected `newWithClient` fake-client unit tests fast unless a test explicitly sets `SDKStopLockTimeout > 0`, and validating `Timeouts.Force >= ForceStopBudget()` in `hub.NewShutdown` reserves the full lock-contended force-stop budget after the graceful `t.Stop` cutoff expires.
3. **`KindServerShutdown` gap handling in `hub.Cursor` and `wsclient` (`80996C79`)**: Returning `ActionApply` / `VerdictApply` while setting `c.stale = true` when `KindServerShutdown` arrives across a sequence gap ensures consumers always observe terminal shutdown even after `deliverTerminal` evicts the oldest queued frame from a full buffer, without falsely advertising skipped sequence numbers in `Position()`.
4. **Bounding write diff/content previews without auto-denying multi-KB file writes (`FA1C38F7(e)`)**: Checking `out.Truncated` on `r.FileName` first and bounding `r.Diff` / `r.NewFileContents` into the remaining `maxSummaryRunes` budget (`rem := maxSummaryRunes - out.FullLen - 1`) surfaces diffs and new file contents to local and remote operators while still allowing legitimate file writes larger than 2,048 runes to be reviewed and approved.

## Risks and Caveats

* **`TestHubPackageHasNoBareChannelSends` AST guard (`internal/hub/degraded_test.go`)**: Every channel send in `internal/hub/*.go` production files must remain a `select` communication clause (or `trySend`).
* **Existing golden tests (`internal/hub/envelope_test.go` & `internal/tui/testdata/modal.golden`)**: `ToolEndedPayload.Text` uses `json:"text,omitempty"`, `entryText` only appends `": " + e.Text` when `e.Text != ""`, and `wrapHeader` only indents continuation lines (`i >= 1`), keeping all existing golden snapshots unchanged.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Adds `copilot.SDKRuntimeShutdownTimeout`, `copilot.SDKProcessExitTimeout`, `copilot.DefaultSDKStopLockTimeout`, `copilot.Config.SDKStopLockTimeout`, `copilot.MaxEventTextBytes`, `copilot.MaxEventIdentBytes`, `agent.KindReasoning`, `agent.KindReasoningDelta`, `agent.Event.Truncated`, `hub.KindAgentReasoning`, `hub.KindAgentReasoningDelta`, `hub.ToolEndedPayload.Text`, and amends `hub.Cursor.Apply` and `wsclient` `cursor.apply` on `KindServerShutdown`.
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Modifies permission request sanitization (`sanitizeBound` `Zl`/`Zp`, `sanitizeIdent` on `RawKind`/`ToolCallID`, write-diff and shell-URL summary formatting) and TUI permission modal `kind:`/`tool call:` soft-wrap continuation indentation (`wrapHeader`).
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive action.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — No new external dependencies in `go.mod`.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Modifies Copilot adapter `Stop`/`ForceStop` timeout coordination, `App.Err` shutdown error aggregation, and remote/local `Cursor` shutdown frame handling.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `070-S` (`077-F`)**: Copilot adapter `Stop`/`ForceStop` bounds (`ForceStopBudget`), permission request sanitization, post-quiesce `session.shutdown` / terminal `session.idle` handling, and `App.Err` stuck-lane error aggregation. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/agent/... ./internal/hub/... ./internal/app/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`.
* **Shipment `071-S` (`078-F`)**: `hub.Cursor` and `wsclient` `KindServerShutdown` gap handling and TUI permission modal `kind:`/`tool call:` continuation indentation. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/... ./internal/transport/... ./internal/tui/... ./internal/app/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`.
* **Shipment `072-S` (`079-F`)**: `agent.Event` rune-safe bounds, `tool.ended` result/error and reasoning enrichment, write-diff/shell-URL surfacing, and unmapped-event log rate-limiting. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`. Rollback for each shipment is a clean `git revert` of its merge commit.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was triggered by three signals: (1) domain and wire-envelope contract additions across `internal/agent`, `internal/agent/copilot`, `internal/hub`, `internal/transport/wsclient`, and `internal/tui`, (2) permission-modal anti-spoofing and permission-request sanitization controls (`FA1C38F7(a, b, d, e)`, `A999ABD6`), and (3) shutdown/disconnect concurrency coordination (`407291BE`, `70BEBF29`, `FA1C38F7(c, f)`, `80996C79`).

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, V, VI, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-19-hub-test-oracles-and-toolchain-pitfalls.md`

### 2. Protected Invariants

1. **No bare channel sends in `internal/hub` (`TestHubPackageHasNoBareChannelSends`)**: Every channel send in `internal/hub/*.go` production files must live inside a `select` statement (or use `trySend`), never a bare `ch <- v`.
2. **Zero new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives**: `go run ./tools/gatecheck write-path --self-test --root .` must exit 0 after every task.
3. **No direct SDK or adapter imports in `internal/tui` (`TestPackageHasNoDirectAgentOrSDKReach`)**: `internal/tui` non-test files must not import `copilot-sdk` or `internal/agent/copilot`.
4. **Fail-closed permission truncation and full-view approval gating (`TestTruncatedShellCommandIsDeniedWithoutAskingHub`, `TestLongRequestIsScrollableAndApproveRefusedUntilFullyViewed`)**: Any shell/read/url/other permission request or write `FileName` exceeding `maxSummaryRunes` (`2048` runes) continues to set `PermissionRequest.Truncated = true` and is denied by `permissionBridge.handle` without asking the hub; clipped write diff/content previews stay strictly `<= maxSummaryRunes` (`2048` runes) with an explicit `"\n[write preview truncated]"` suffix; wrapped `kind:` and `tool call:` continuation lines in `permissionBody` feed directly into `m.modal.TotalLineCount()` so `m.approvable()` requires scrolling through the entire modal body.
5. **Non-blocking post-quiesce terminal idle delivery (`TestEventDiscardAfterQuiesceIsLoggedAtWarn`)**: After `latch.quiesce` closes the callback latch, non-idle events are discarded and logged at `Warn`, while a terminal `*rpc.SessionIdleData` uses only a non-blocking `select` with `default:` so `onEvent` can never block `readLoop` after `quiesce`.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Harden `sanitizeBound` (`Zl`/`Zp`), `sanitizeIdent` (`RawKind`/`ToolCallID`), pre-send `gone()` re-check, `Adapter.Start` `fail()` latch quiescing, `session.shutdown` informational classification, post-quiesce terminal `session.idle` non-blocking delivery, `SDKStopLockTimeout`-aware `forceStopWait()` / `ForceStopBudget()` (`DefaultSDKStopLockTimeout = 20s`, post-cutoff `Timeouts.Force >= ForceStopBudget()`), and `App.assemble` `a.addErr` propagation of `Hub.Run` errors (`070-S` / `077-F`). | `local edit` (adapter, hub & app shutdown and permission hardening) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `070-S` merge commit | `planned` |
| `PA-02` | Treat `KindServerShutdown` as always-applicable across sequence gaps in `hub.Cursor.Apply` and `wsclient` `cursor.apply` (marking `c.stale = true` on non-contiguous terminal frames) and indent soft-wrapped `kind:` and `tool call:` continuation lines via `wrapHeader` in `internal/tui/permission.go` (`071-S` / `078-F`). | `local edit` (cursor, wsclient & TUI modal hardening) | `moderate` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `071-S` merge commit | `planned` |
| `PA-03` | Surface shell `PossibleURLs` and bounded write `Diff`/`NewFileContents` (with explicit `"\n[write preview truncated]"` suffix) in `PermissionRequest.Summary`, add rune-safe `MaxEventTextBytes`/`MaxEventIdentBytes` bounds, `Event.Truncated`, `KindReasoning`/`KindReasoningDelta`, `tool.ended` result/error `Text`, and per-event-type `unmappedLogGate` in `internal/agent/copilot`, fold/publish/render them in `internal/hub` and `internal/tui` (with hub-side `unmappedWarned` rate-limiting), and amend `docs/design-docs/intercom-go-backend-architecture.md` §3.1, §4.2, §4.3, and §5.2 (`072-S` / `079-F`). | `local edit` (permission preview, event telemetry, hub fold & TUI rendering) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `072-S` merge commit | `planned` |

### 4. Deepened Verification, Rollback, and Review-Gate Requirements

* **AST and structural guards**: Run `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub -run TestHubPackageHasNoBareChannelSends` after every `internal/hub` task, `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui -run TestPackageHasNoDirectAgentOrSDKReach` after every `internal/tui` task, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .` after every shipment.
* **Review-gate capability requirement (P-012)**: Plan review (Step 4) MUST spawn the always-on and triggered cross-model persona subagents (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, `Security Lens Reviewer`), using `model_routing.anchor_review` (`gpt-6.1-sol` / `high`) for `Architecture Strategist`, and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* **Unresolved operator decisions**: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating all P1, P2, and P3 findings from the 6-persona multi-agent review pass directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01`..`PA-03` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Refinement) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 0 P3) | **PASS** |
| **Go Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 1 P2, 3 P3) | **PASS** |
| **Scope Boundary Auditor** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 0 P3) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | PASS (8 high-confidence solutions verified) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | FAIL -> Remediated (2 P1, 4 P2) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 3 P2, 1 P3) | **PASS** |

### Findings and Applied Remediations

1. **[P1 — Remediated] Unit 1.3 (`077.003-T`) Worst-Case Cleanup Budget Coupling (`ForceStopBudget` vs Post-Cutoff `ShutdownTimeouts.Force`, Two-Phase SDK Lock Hold, and Real-Client Lock Allowance)**:
   * *Finding* (`Architecture Strategist` & PR #209 Copilot Review Threads `PRRT_kwDOTPuhps6xwM0A`, `PRRT_kwDOTPuhps6xwM0r`, `PRRT_kwDOTPuhps6xwM1I`): (a) Under `startStopMux`, `copilotsdk.Client.Stop()` runs both `runtimeShutdownTimeout = 10s` and `killProcessAndWait()` (`processExitTimeout = 10s`), so `Client.Stop()` can retain `startStopMux` for up to `20s` total (`DefaultSDKStopLockTimeout = 20s`) before `ForceStop()` can acquire it; (b) preceding graceful steps in `Adapter.doStop(sc)` (callback quiescing and `sess.Disconnect`) can consume nearly the entire `t.Stop` graceful window before `Client.Stop()` starts, so `a.forceStop()` can still wait up to `ForceStopBudget() = StopTimeout + SDKStopLockTimeout` (`25s` by default) *after* the graceful `t.Stop` cutoff expires — comparing `Stop + Force` with `ForceStopBudget()` double-counted `t.Stop`; and (c) gating default `SDKStopLockTimeout` on `StopTimeout >= DefaultStopTimeout` disabled SDK lock-contention protection on real adapters configured with `StopTimeout < 5s`.
   * *Remediation applied*: Exported `SDKRuntimeShutdownTimeout = 10 * time.Second`, `SDKProcessExitTimeout = 10 * time.Second`, and `DefaultSDKStopLockTimeout = SDKRuntimeShutdownTimeout + SDKProcessExitTimeout` (`20 * time.Second`); defaulted `cfg.SDKStopLockTimeout` to `DefaultSDKStopLockTimeout` in `newAdapter` whenever `c` is a real SDK client (`isRealSDKClient()`), independently of `cfg.StopTimeout` (while keeping `newWithClient` fake-client adapters at `0` unless explicitly configured); exported `(*Adapter).ForceStopBudget() time.Duration` returning `a.cfg.StopTimeout + a.cfg.SDKStopLockTimeout`; and updated `hub.NewShutdown` in `internal/hub/shutdown.go` to fold `fb.ForceStopBudget()` into `cfg.AdapterStopTimeout = max(cfg.AdapterStopTimeout, fb.ForceStopBudget())` so the post-cutoff `Timeouts.Force` budget alone satisfies `Force >= ForceStopBudget()` (defaulting `Force = max(DefaultShutdownForce, AdapterStopTimeout + 2s)` = `27s` on the default real adapter and rejecting custom `Force < ForceStopBudget()` with `apperr.KindConfig`), with delayed-reaping, delayed-start, and sub-5s real-client regression coverage in `internal/agent/copilot/shutdown_hub_test.go`.
2. **[P1 — Remediated] Units 3.2 & 3.3 (`079.002-T` & `079.003-T`) Self-Contained Green Test Suite Across Units**:
   * *Finding* (`Architecture Strategist`): Mapping `AssistantReasoningData` to `agent.KindReasoning` in `079.002-T` while deferring the update of `TestCallbackUnknownEventIsForwardedNotDropped` in `adapter_test.go` to `079.003-T` would leave `go test ./internal/agent/copilot` red between `079.002-T` and `079.003-T`.
   * *Remediation applied*: Kept `079.002-T` focused on `agent.KindReasoning`/`KindReasoningDelta`/`Event.Truncated` declarations, `boundUTF8`, `unmappedLogGate`, and `ToolExecutionCompleteData` `Text` extraction (with `normalize_test.go` migrated to `&rpc.SessionErrorData{}`), and moved the `*rpc.AssistantReasoningData` / `*rpc.AssistantReasoningDeltaData` switch cases in `normalize.go` into `079.003-T` alongside `adapter.go` and `adapter_test.go` (still 3 files), ensuring `go test -race ./internal/agent/...` is 100% green after every task.
3. **[P2 — Remediated] Unit 3.4 (`079.004-T`) Upstream `ev.Truncated` Propagation Across All `Fold` Branches and `f.lastDelta`**:
   * *Finding* (`Architecture Strategist`, `Security Lens Reviewer`, `Go Reviewer`): Because `MaxEventTextBytes` (`16384`) and `MaxEventIdentBytes` (`128`) match `Fold`'s defaults, events pre-clipped in `normalizeWithGate` arrive at `Fold` with `len(ev.Text) <= MaxEntryTextBytes` and `ev.Truncated == true`. `Fold.applyDelta` and `Fold.applyMessage` (and reasoning equivalents) must explicitly OR `ev.Truncated` into `cur.Truncated`, `e.Truncated`, and `f.lastDelta.truncated`.
   * *Remediation applied*: Updated Unit 3.4 (`079.004-T`) so `applyTextDelta` and `applyTextComplete` propagate `ev.Truncated` into `f.lastDelta.truncated`, `cur.Truncated`, `e.Truncated`, and `f.anyTrunc`, and `mapEvent` propagates both `ev.Truncated` and `len(text) < len(ev.Text)` on `ToolEndedPayload`.
4. **[P2 — Remediated] Unit 3.1 (`079.001-T`) `PermissionRequestWrite` Preview Rune Budget and Explicit `[write preview truncated]` Notice**:
   * *Finding* (`Architecture Strategist`, `Security Lens Reviewer`): Passing `rem = maxSummaryRunes - out.FullLen - 1` to `sanitizeBound` on a clipped write diff/content preview would add a 1-rune `"…"`, producing `maxSummaryRunes + 1` (`2049`) runes, and a bare `"…"` at the end of a clipped diff does not clearly warn the operator that the write preview was truncated.
   * *Remediation applied*: Reserved `utf8.RuneCountInString("\n[write preview truncated]")` (`26` runes) plus `1` rune for `"…"` inside `maxSummaryRunes` whenever a write diff/content preview is clipped (appending `"\n[write preview truncated]"` so total runes never exceed `2048`), and set `out.Truncated = true` (fail-closed auto-deny) if `FileName` is so long (`2023..2048` runes) that even the truncation notice cannot fit within `maxSummaryRunes`.
5. **[P2/P3 — Remediated] Units 3.2, 3.5, & 3.6 (`079.002-T`, `079.005-T`, `079.006-T`) Identifier/Log Bounding, Hub-Side `unmappedWarned` Rate-Limiting, Single-Pass `boundUTF8`, §3.1/§4.2 Docs, and `ActionRisk` Elevation**:
   * *Finding* (`Architecture Strategist`, `Security Lens Reviewer`, `Go Reviewer`): Bound and sanitize `raw` before `gate.shouldWarn(raw)` and logging; lazily initialize `unmappedLogGate.seen` under `g.mu`; enforce `maxBytes` in a single pass after `U+FFFD` replacement in `boundUTF8`; rate-limit repeated `hub: unmapped agent event` `Warn` logs per `RawType` in `Hub.handleEvent`; document §3.1 and §4.2 updates in `079.006-T`; and elevate `PA-01`/`PA-03` to `ActionRisk: high` and `PA-02` to `ActionRisk: moderate` in `## Plan Hardening`.
   * *Remediation applied*: Incorporated all five improvements into Units 1.2, 3.2, 3.5, 3.6, and `## Plan Hardening`.

