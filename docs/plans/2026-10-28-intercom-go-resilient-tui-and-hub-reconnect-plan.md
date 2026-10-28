---
title: "Resilient TUI and Hub Reconnect Experience Implementation Plan"
description: "Implementation plan for Capability Set 1 across three chained single-feature shipments (067-S..069-S): hub fold/publish/lifecycle hardening (074-F), shutdown broadcast/drain fidelity and in-place Client.Rehydrate (075-F), and resilient TUI reconnect, sanitization, and viewport hardening (076-F)."
source: "docs/decisions/2026-10-28-intercom-go-resilient-tui-and-hub-reconnect-deliberation.md"
status: "ready"
date: 2026-10-28
tags:
  - "hub"
  - "tui"
  - "transport"
  - "reconnect"
  - "shutdown"
  - "permissions"
---

## Problem Frame

With `063-S..066-S` merged and closed on `main` (`066-S` merge commit `65780da0d83a53182a56460093b80cc132b5d945`, closure commit `7617a00fc29660cac87dd8d67a177518e0d68ec2`), Capability Set 1 (**Resilient TUI & Hub Reconnect Experience**) resolves seven deferred-scope-expansion stash entries (`1DCB6108`, `550437DB`, `E9223340`, `4AFE4DA4`, `AEF57AF1`, `8769C9A0`, and `2C74C5FC`) spanning `internal/hub`, `internal/agent/copilot`, `internal/transport`, `internal/app`, and `internal/tui`.

To satisfy Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and maintain strict architectural layering, the work is organized into three sequential single-feature shipments (`067-S -> 068-S -> 069-S`). As in `063-S..066-S`, shipment `067-S` is declared a DAG root (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`) under the operator's explicit P-001 policy exception over `062-S`'s two open operator-present live runtime conditions (`live-sdk-test-pass` and `tool-approval-routing-live`), with `066-S` (`65780da0d83a53182a56460093b80cc132b5d945`) documented as the code baseline.

## Requirements Trace

| Source Requirement | Stash Origin | Target Shipment & Feature | Implementation Unit(s) |
|---|---|---|---|
| Early-return in `Fold.applyDelta` when already truncated at cap and `add == ""`; backfill `Tool` name and propagate truncation flag in `Fold.applyToolEnded` | `2C74C5FC(d, e)` | `067-S` / `074-F` | Unit 1.1 (`074.001-T`) |
| Avoid duplicate fold error entry on `Hub.publish` stub path when `kind == KindError`; call `markAllStale` if even the error stub fails; drain full remote buffer on `Hub.rehydrate` instead of dropping | `2C74C5FC(a, c, f)` | `067-S` / `074-F` | Unit 1.2 (`074.002-T`) |
| Join operator lane workers in `Hub.Run`; fold and publish `KindError` (`operator_action_failed`) on non-shutdown async `Send`/`Abort` failure; settle/resolve pending permissions (`turn_aborted`) and reject late asks (`turnAborted`) on aborted `agent.idle` | `550437DB` | `067-S` / `074-F` | Unit 1.3 (`074.003-T`) |
| Make `Adapter.recheckForce` non-blocking after failed `Start`; export `(*Adapter).StopTimeout()` and validate `ShutdownConfig.AdapterStopTimeout` coupling (migrating `TestShutdownThatCannotConfirmForceStopReportsOrphanRisk`); document `Hub.Run` ctx, `Shutdown.Done`, and `promptsPending` contracts | `8769C9A0`, `550437DB` | `067-S` / `074-F` | Unit 1.4 (`074.004-T`) |
| Sweep closed clients at `Hub.intake` start; honor pending connects in `reapEmptyBroadcastSet` (with post-`serviceHydrations` reap and `membershipWake` on abandoned `Connect`); deliver terminal `server.shutdown` to stale/full-buffer clients while suppressing subsequent stale hydration; add `Client` `inflight`/`terminal` frame helpers (`sendFrame`, `drainOneFrame`, `drainAllFrames`) | `4AFE4DA4(a, b, c, d)`, `2C74C5FC(b)`, `AEF57AF1(a)` | `068-S` / `075-F` | Unit 2.1 (`075.001-T`) |
| Add `Client` write-ack completion (`EnableWriteAck`, `AckWrite`, `Drained`) and check `c.Drained()` in `drainAndCloseClients`; add in-place `(*hub.Client).Rehydrate(ctx)` and hub control handler `(*Hub).rehydrateClient` (`ctlRehydrate`) | `4AFE4DA4(b)`, `AEF57AF1(a)`, `1DCB6108` (hub half) | `068-S` / `075-F` | Unit 2.2 (`075.002-T`) |
| Wire `EnableWriteAck`/`AckWrite` in `transport.session.writePump` and unconditional `AdapterStopTimeout: adapterStopTimeout(a.agent)` in `internal/app/shutdown.go` | `4AFE4DA4(b)`, `AEF57AF1(a)`, `8769C9A0(c)` | `068-S` / `075-F` | Unit 2.3 (`075.003-T`) |
| Propagate `ctx` in `App.Reissue`/`tuiReissue` (`mint(ctx)`) and drain `remote` concurrently in `e2e_test.go` `shutdownScenario` | `AEF57AF1(b, c)` | `068-S` / `075-F` | Unit 2.4 (`075.004-T`) |
| Use in-place `Conn.Rehydrate` in `tui.Feed` on overflow, add bounded redial retry (`RedialRetries`, `RedialBackoff`), and auto-redial on unexpected `conn.Done()` without `server.shutdown` | `1DCB6108` (TUI half), `E9223340(c)` | `069-S` / `076-F` | Unit 3.1 (`076.001-T`) |
| Sanitize non-`Cf` invisible blank runes (`U+3164`, `U+2800`, `U+034F`, `Co`, variation selectors) and add `isBlankSummary` (including `utf8.RuneError`) | `E9223340(d)` | `069-S` / `076-F` | Unit 3.2 (`076.002-T`) |
| Enforce `isBlankSummary` deny-only in `newPermission` and wrap-aware 2-space summary indentation against column-0 header spoofing | `E9223340(d, e)` | `069-S` / `076-F` | Unit 3.3 (`076.003-T`) |
| Add modal-generation pure-Elm `ApprovalDelay` (`DefaultApprovalDelay = 150ms` in `tui.New`, explicit `NoApprovalDelay` zero-delay test path, `ApprovalReadyMsg{RequestID, ModalGen}` rejecting stale timers after modal resets), disarm `m.armed` on mouse/resync/hydrate, and trigger `Rehydrater.Rehydrate` (`ResyncFailedMsg`) on stuck `resync` | `E9223340(c, f)` | `069-S` / `076-F` | Unit 3.4 (`076.004-T`) |
| Cap live `tui.Model` transcript entries (`MaxEntries`) with delta-invalidated per-entry wrap cache; truncate prompt by `lipgloss.Width`; enable mouse reporting in `Model.Init()` (`tea.EnableMouseCellMotion`) and handle `tea.MouseMsg` wheel scroll; note `(a)` already absorbed in `062-S` | `E9223340(a, b, g)` | `069-S` / `076-F` | Unit 3.5 (`076.005-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | PASS | All production code uses Go 1.25.0 (`GOTOOLCHAIN=go1.26.10`), explicit `apperr` error wrapping, bounded waits, and zero `panic` calls in library packages. |
| II. Test-First Development (NON-NEGOTIABLE) | PASS | Every implementation unit specifies red-phase failing test assertions before production code changes and verifies green via `go test -race ./...`. |
| III. Workspace Isolation and Security Boundaries | PASS | Zero new filesystem write calls or `pathsafe.Root.Resolve` call sites (`go run ./tools/gatecheck write-path --self-test --root .` stays 100% green). Operator action failure envelopes use fixed sanitized messages and never echo prompt text or credentials. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | PASS | All modifications are confined to the repository working tree. |
| V. Structured Observability | PASS | Asynchronous operator `Send`/`Abort` failures are both logged (`slog.Logger.Warn` without prompt text) and published as `KindError` (`operator_action_failed`) envelopes so TUI and remote operators observe them. |
| VI. Single Responsibility | PASS | Zero new external dependencies in `go.mod` (`lipgloss.Width` is already in `github.com/charmbracelet/lipgloss v1.1.0`). |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | PASS | No destructive commands are used; consumed stash entries are archived via non-destructive `backlogit stash archive`. |
| VIII. Explicit Safety Modes | PASS | Freeze-scope mode active on the 7 chartered stash entries of Capability Set 1. |
| IX. Git-Friendly Persistence | PASS | Deliberation, plan, session memory, and backlog items use Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | PASS | Targeted code and backlog queries used throughout staging. |
| XI. Merge Commit History Preservation | PASS | All shipments will merge via `--merge` commits (P-009). |

## Implementation Units

### Sub-Epic 1: Shipment `067-S` — Covering Feature `074-F` (Hub fold, publish, and lifecycle hardening)

#### Unit 1.1 (`074.001-T`): Harden `Fold` delta cap re-marshalling and `tool.ended` name/truncation parity (`2C74C5FC(d, e)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/hub/fold.go`
  * `internal/hub/fold_test.go`
* **Changes needed**:
  * In `(*Fold).applyDelta` (`internal/hub/fold.go:193`): change `if add == "" && !clipped { return }` to `if add == "" && (cur.Truncated || !clipped) { return }` (after updating `f.lastDelta` and `if clipped { f.anyTrunc = true }`), so once a streaming entry has reached `MaxEntryTextBytes` (`remaining == 0`) and already has `cur.Truncated == true`, subsequent deltas do not call `f.replace(i, cur)` and do not re-marshal the 16 KiB entry.
  * In `(*Fold).applyToolEnded` (`internal/hub/fold.go:212-225`): compute `name, t2 := f.identFlag(ev.ToolName)` before `if i := f.find("tool", id, statusRunning); i >= 0`. Inside that `i >= 0` branch, update `cur.Status = status`, set `if cur.Tool == "" && name != "" { cur.Tool = name }`, and set `cur.Truncated = cur.Truncated || t1 || t2` before calling `f.replace(i, cur)` so a folded `tool.ended` entry matches the live `mapEvent` `ToolEndedPayload` name and truncation flags.
  * Add unit tests in `internal/hub/fold_test.go`:
    1. `TestFoldDeltaAtCapDoesNotRemarshalWhenAlreadyTruncated`: fill a streaming entry to `MaxEntryTextBytes` (16 KiB), apply one delta that clips and sets `Truncated: true`, and then measure `runtime.ReadMemStats` across 1,000 further 1-byte deltas to assert total allocations stay well below 64 KiB (proving `marshalEntry` is not called per delta once already truncated).
    2. `TestFoldToolEndedBackfillsNameAndPropagatesTruncation`: verify that a `KindToolStarted` with an empty `ToolName` followed by a `KindToolEnded` with a non-empty `ToolName` (and a case with an oversize `> 128`-byte `ToolName` on `KindToolEnded`) backfills `tool_name` and sets `truncated: true` on the folded entry and snapshot.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestFoldDeltaAtCap|TestFoldToolEnded"`

#### Unit 1.2 (`074.002-T`): Harden `Hub.publish` stub fold-error deduplication, unbuildable-stub `markAllStale`, and stale remote buffer drain on `rehydrate` (`2C74C5FC(a, c, f)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/hub/hub.go`
  * `internal/hub/review3_test.go`
* **Changes needed**:
  * In `(*Hub).publish` (`internal/hub/hub.go:671-686`):
    * When `stub` is true and building the `ErrCodeFrameTooLarge` stub itself returns `err != nil`, set `h.degradedSeq = h.seq` and call `h.markAllStale()` before returning `0, false` so registered clients are forced to re-hydrate when `h.fold` was already mutated by the caller prior to `publish`.
    * When `stub` is true and the stub frame is built, guard `h.fold.ApplyError(ErrCodeFrameTooLarge, msgFrameTooBig)` with `if kind != KindError` so callers publishing `KindError` (which already called `h.fold.ApplyError`) do not produce a duplicate second error entry in `h.fold`.
  * In `(*Hub).rehydrate` (`internal/hub/hub.go:1089-1114`): remove the `if c.kind != ClientTUI { h.overflow(c.id, c); return false }` branch so that when `trySend(c.send, f)` finds `c.send` full on any stale client (`ClientTUI` or `ClientRemote`), `rehydrate` drains the superseded frames from `c.send` and enqueues the fresh `session.hydrate` snapshot `f` without dropping the remote client.
  * Add unit tests in `internal/hub/review3_test.go`:
    1. `TestPublishStubDoesNotDuplicateFoldErrorForKindError`: configure a hub where a `KindError` payload exceeds `MaxFrameBytes` (or triggers the stub path after `h.fold.ApplyError`) and assert `h.fold.Len() == 1` rather than `2`.
    2. `TestPublishUnbuildableStubStillMarksClientsStale`: force `encodeFrame` to fail on the stub path (e.g., via a test-only invalid payload / hook or unencodable state) after mutating `h.fold` and assert `c.stale == true` and `len(h.staleQ) == 1`.
    3. `TestRehydrateDrainsFullStaleRemoteBufferInsteadOfDropping`: connect a `ClientRemote` with `Buffer: 2`, fill its buffer with 1 live frame plus the `SignalDegraded` error frame (`len(c.send) == 2`), run `serviceHydrations`, and assert the remote client is NOT dropped (`!c.isClosed()`, `DroppedClients == 0`) and its buffer now holds the fresh `KindSessionHydrate` frame.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestPublishStub|TestPublishUnbuildable|TestRehydrateDrainsFullStaleRemote"`

#### Unit 1.3 (`074.003-T`): Join operator lane workers in `Hub.Run`, publish `KindError` on async operator `Send`/`Abort` failure, and resolve pending permissions on aborted `agent.idle` (`550437DB`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/hub.go`
  * `internal/hub/operator.go`
  * `internal/hub/operator_lanes_test.go`
* **Changes needed**:
  * In `internal/hub/operator.go` and `internal/hub/hub.go`:
    * Update `runOperatorLanes(ctx context.Context, wg *sync.WaitGroup)` to call `wg.Add(2)` AFTER `if h.cfg.Agent == nil { return }` before launching the prompt and abort worker goroutines (`defer wg.Done()` in each), and replace the `select { case h.promptFailed <- struct{}{}: case <-h.done: }` inside `if !invoked` with `trySend(h.promptFailed, struct{}{})` (since `h.promptFailed` has capacity `operatorPromptQueue + 1`, `trySend` never blocks, never races `<-ctx.Done()`, and satisfies `TestHubPackageHasNoBareChannelSends`). In `(*Hub).Run`, declare `var lanesWG sync.WaitGroup`, register `defer lanesWG.Wait()` immediately after `defer close(h.done)` and BEFORE `opsCtx, opsCancel := context.WithCancel(ctx); defer opsCancel()` so on return `opsCancel()` executes first, `lanesWG.Wait()` joins both workers second, and `close(h.done)` executes last.
    * Define `const ErrCodeOperatorActionFailed = "operator_action_failed"` and fixed messages `msgPromptSendFailed = "operator prompt send failed"` and `msgTurnAbortFailed = "operator turn abort failed"` in `internal/hub/hub.go`, plus `actionFailed chan string` (capacity `operatorPromptQueue + 2`) on `Hub`. In `(*Hub).callAgent` (`internal/hub/operator.go`), when `err := call(callCtx); err != nil`, after logging the warning, if `ctx.Err() == nil` (i.e. the lanes' parent context was not cancelled by `stopLanes` or `Hub.Stop`), call `trySend(h.actionFailed, action)`. In `(*Hub).Run`, add `case action := <-h.actionFailed: h.handleActionFailed(action)` which (when `!h.closing`) maps `"abort"` to `msgTurnAbortFailed` and `"prompt"` to `msgPromptSendFailed`, calls `h.fold.ApplyError(ErrCodeOperatorActionFailed, msg)`, and calls `h.publish(KindError, ErrorPayload{Code: ErrCodeOperatorActionFailed, Message: msg})`.
    * Add `turnAborted bool` on `Hub` and `func (h *Hub) intakeAsk(ask agent.PermissionAsk)` in `internal/hub/hub.go` (called from `case ask := <-h.asks:` in `(*Hub).Run`): when `h.turnAborted` is `true`, if `ask.Reply != nil`, check `select { case <-ask.Resolved: default: trySend(ask.Reply, agent.PermissionDecision{Kind: agent.DecisionUnavailable}) }` and return immediately without calling `h.handleAsk(ask)` (so late permission asks arriving from an aborted turn after `KindIdle{Aborted: true}` are rejected at once and never park in `h.perm`). In `(*Hub).handleEvent` (`internal/hub/hub.go`): under `case agent.KindTurnStart:`, clear `h.turnAborted = false`; under `case agent.KindIdle:`, when `ev.Aborted` is true, set `h.turnAborted = true`, drain any asks already queued in `h.asks` via `h.intakeAsk(ask)` (`for { select { case ask := <-h.asks: h.intakeAsk(ask) default: break } }`), iterate `slices.Clone(h.perm.order)` and call `h.settleDecided(e, permReasonTurnAborted)` on any `e.state == permDecided` entry (atomically revoking any unconsumed approval via `e.ask.Outcome.Revoke()`), and finally resolve every unresolved permission in `h.perm.unresolved()` to `agent.DecisionUnavailable` with reason `"turn_aborted"` (`const permReasonTurnAborted = "turn_aborted"`) via `h.decide(e, ActorHub, agent.PermissionDecision{Kind: agent.DecisionUnavailable}, permReasonTurnAborted)` before publishing `KindAgentIdle`.
  * Add unit tests in `internal/hub/operator_lanes_test.go`:
    1. `TestHubRunJoinsOperatorLaneWorkersBeforeDoneCloses`: block `Agent.Send` until its `ctx` is cancelled (plus a brief post-cancel hook before `Send` returns), call `h.Stop()`, and assert the worker's exit flag is `true` by the time `<-h.Done()` unblocks.
    2. `TestOperatorSendAndAbortFailuresPublishSanitizedErrorEnvelope`: configure `fakeAgent` to fail `Send` and `Abort`, invoke both via `tui.Do`, and assert `tui.Frames()` receives `KindError` envelopes with `Code == ErrCodeOperatorActionFailed` and fixed messages without leaking the prompt text, while `h.fold.Snapshot()` also contains the error entries.
    3. `TestAbortedIdleEventSettlesDecidedResolvesPendingAndRejectsLateAsks`: verify that on `agent.Event{Kind: agent.KindIdle, Aborted: true}`, (a) an unresolved `permPending` entry and an unacknowledged `permDecided` entry registered before the idle event are resolved/settled with `permReasonTurnAborted` (`DecisionUnavailable` / `permission.aborted`), (b) an ask queued in `h.asks` at the moment of the aborted idle is answered `DecisionUnavailable`, and (c) in the `idle-before-ask` race where a late `PermissionAsk` from the aborted turn arrives on `h.Asks()` *after* `agent.Event{Kind: agent.KindIdle, Aborted: true}` has already been processed (and before the next `agent.KindTurnStart`), the late ask is immediately rejected with `DecisionUnavailable` without registering in `h.perm` (`Stats.PendingPermissions == 0`), whereas a subsequent `agent.Event{Kind: agent.KindTurnStart}` clears `h.turnAborted` so asks in the new turn register normally.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestHubRunJoins|TestOperatorSendAndAbortFailures|TestAbortedIdleEvent|TestFailedPromptSend|TestLateSendFailure"`

#### Unit 1.4 (`074.004-T`): Make `Adapter.recheckForce` non-blocking after failed `Start` and validate `ShutdownConfig.AdapterStopTimeout` coupling (`8769C9A0`, `550437DB`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/copilot/adapter.go`
  * `internal/hub/shutdown.go`
  * `internal/agent/copilot/shutdown_hub_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/adapter.go`: export `const DefaultStopTimeout = 5 * time.Second` (keeping `const defaultStopTimeout = DefaultStopTimeout`) and add `func (a *Adapter) StopTimeout() time.Duration { return a.cfg.StopTimeout }` so callers can query the adapter's effective `ForceStop` timeout budget independently of `ShutdownTimeouts.Force`. In `(*Adapter).recheckForce()` (`adapter.go:676-685`), replace `return a.forceStop()` with a non-blocking check on `a.forceDone`:
    ```go
    select {
    case <-a.forceDone:
        a.mu.Lock()
        a.forceUnconfirmed = false
        a.mu.Unlock()
        return nil
    default:
        return fmt.Errorf("agent: %w: ForceStop did not return within %v", agent.ErrCleanupIncomplete, a.cfg.StopTimeout)
    }
    ```
    so `Stop` after a failed `Start` with a wedged `ForceStop` does not wait a second `StopTimeout`.
  * In `internal/hub/shutdown.go`: add `AdapterStopTimeout time.Duration` to `ShutdownConfig`. In `NewShutdown`: when `cfg.AdapterStopTimeout == 0`, populate it from `if st, ok := cfg.Agent.(interface{ StopTimeout() time.Duration }); ok { cfg.AdapterStopTimeout = st.StopTimeout() }`; reject `cfg.AdapterStopTimeout < 0` or `(cfg.AdapterStopTimeout > 0 && cfg.Timeouts.Force > 0 && cfg.Timeouts.Force < cfg.AdapterStopTimeout)` with an `apperr.KindConfig` error; when `cfg.AdapterStopTimeout > 0 && cfg.Timeouts.Force == 0`, set `cfg.Timeouts.Force = max(DefaultShutdownForce, cfg.AdapterStopTimeout+2*time.Second)` before `withDefaults()`. Document the `Hub.Run` signal-independent context requirement, `Shutdown.Done` lifecycle, and `promptsPending` 1-to-1 `turn.start` / bounded `WaitIdle` behavior in `internal/hub/shutdown.go`.
  * In `internal/agent/copilot/shutdown_hub_test.go`: migrate `TestShutdownThatCannotConfirmForceStopReportsOrphanRisk` (`shutdown_hub_test.go:298`) by setting `r.h.ad.cfg.StopTimeout = 200 * time.Millisecond` before `r.newShutdown` (so its explicit `c.Timeouts.Force = 300 * time.Millisecond` satisfies `Force >= StopTimeout` under the new `NewShutdown` validation while still exercising orphan-risk reporting on a wedged `ForceStop`), add `TestStopAfterFailedStartRecheckForceIsNonBlocking` asserting that `h.ad.Stop(context.Background())` after a failed `Start` (with `StopTimeout = 200ms`) returns `agent.ErrCleanupIncomplete` in `< 50ms` (proving `recheckForce` did not re-wait `StopTimeout`), and add `TestNewShutdownValidatesAdapterStopTimeoutCoupling` verifying `hub.NewShutdown` rejects `Force < AdapterStopTimeout` (both when `AdapterStopTimeout` is set explicitly and when derived from `(*copilot.Adapter).StopTimeout()` with an explicit `300ms` `Force` against a `1s` `StopTimeout` adapter) and scales default `Force` when `AdapterStopTimeout > DefaultShutdownForce`, while preserving short `Force` budgets for in-process test fakes whose `StopTimeout` is `0`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot/... ./internal/hub/...`

---

### Sub-Epic 2: Shipment `068-S` — Covering Feature `075-F` (Shutdown broadcast/drain fidelity and in-place hub `Client.Rehydrate`)

#### Unit 2.1 (`075.001-T`): Sweep closed clients at intake, honor pending connects on reap, and deliver terminal `server.shutdown` to stale clients (`4AFE4DA4(a, b, c, d)`, `2C74C5FC(b)`, `AEF57AF1(a)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/client.go`
  * `internal/hub/hub.go`
  * `internal/hub/review1_test.go`
* **Changes needed**:
  * In `internal/hub/client.go`: add `inflight atomic.Int64` and hub-owned `terminal bool` on `Client` with `func (c *Client) sendFrame(f Frame) bool` (increments `c.inflight.Add(1)` BEFORE `select { case c.send <- f: return true default: c.inflight.Add(-1); return false }`), `func (c *Client) drainOneFrame() bool` (`select { case <-c.send: c.inflight.Add(-1); return true default: return false }`), and `func (c *Client) drainAllFrames()` (`for c.drainOneFrame() {}`). *(Note: `(*Client).Rehydrate(ctx)` and `EnableWriteAck`/`AckWrite`/`Drained` are introduced in Unit 2.2 (`075.002-T`) alongside `h.rehydrateClient` and `drainAndCloseClients` so Unit 2.1 compiles and passes all tests independently.)*
  * In `internal/hub/hub.go`:
    * Call `h.sweep()` at the top of `(*Hub).intake` before `h.admitPending(req)` so closed remote clients are removed from `h.clients` before `admitPending` checks `live+remote >= h.pendingLimit()`, and skip already-abandoned requests in `admitPending`.
    * Add `func (h *Hub) hasPendingConnect() bool` returning `true` when `!h.closing` and either `len(h.connects) > 0` or `h.pending` contains at least one request with `!p.abandoned.Load()`. In `(*Hub).Connect`, after `req.abandoned.Store(true)` on `<-ctx.Done()`, call `trySend(h.membershipWake, struct{}{})`, and in `(*Hub).Run`, guard both the post-`h.serviceHydrations()` and end-of-loop `h.reapEmptyBroadcastSet()` calls with `if !h.hasPendingConnect() { h.reapEmptyBroadcastSet() }` so an in-flight connect/reconnect is not raced by `permReasonNoSurface`.
    * Route all hub sends/drains on `c.send` in `completeConnect`, `broadcast`, and `rehydrate` through `c.sendFrame(f)` and `c.drainAllFrames()`. In `(*Hub).broadcast`: when `f.Kind() == KindServerShutdown`, deliver to every non-closed client (including `c.stale` and full-buffer clients) via `deliverTerminal(c, f)`: if `!c.sendFrame(f)` on a full buffer, call `c.drainOneFrame()` and `c.sendFrame(f)`, and set `c.stale = false` and `c.terminal = true`. Skip `c.terminal` clients in `serviceHydrations()`, `markAllStale()`, and subsequent non-terminal `broadcast()` calls so a subsequent hub-loop iteration (hydration pass or post-announce event) never drains, replaces, or overflow-drops the queued terminal `server.shutdown` frame before the client reads its channel.
  * Add unit tests in `internal/hub/review1_test.go`:
    1. `TestIntakeSweepsClosedRemotesBeforeCheckingPendingLimit`: register `MaxRemote` remote clients, close them via `c.Close()` without running another hub step, and immediately call `h.intake` with `pendingLimit()` new connect requests, asserting none are falsely rejected with `ErrTooManyClients`.
    2. `TestStaleClientReceivesServerShutdownBeforeClose`: overflow a `ClientTUI` (`Buffer: 2`) so `c.stale == true` and `len(c.send) == 2`, trigger `AnnounceShutdown`, advance the hub clock past `staleRetryInterval` and drive another hub-loop iteration (e.g. a subsequent event and hydration pass) before the stale client reads `c.Frames()`, and assert `c.Frames()` still delivers `KindServerShutdown` intact rather than replacing it with a `session.hydrate` snapshot.
    3. `TestReapEmptyBroadcastSetWaitsForQueuedPendingConnectAndReapsOnAbandon`: queue a connect request in `h.pending` (with no registered client) and post a permission ask, asserting `reapEmptyBroadcastSet` keeps the permission pending while the connect is live and resolves it to `no_surface` once the connect is abandoned and discarded by `serviceHydrations`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestIntakeSweeps|TestStaleClientReceivesServerShutdown|TestReapEmptyBroadcastSet|TestShutdownClosesEveryClientDoneIncludingStale"`

#### Unit 2.2 (`075.002-T`): Add `Client` write-ack `Drained` tracking in `drainAndCloseClients` and in-place `(*hub.Client).Rehydrate(ctx)` (`4AFE4DA4(b)`, `AEF57AF1(a)`, `1DCB6108` hub half)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/client.go`
  * `internal/hub/shutdown.go`
  * `internal/hub/permission_empty_test.go`
* **Changes needed**:
  * In `internal/hub/client.go`:
    * Add `writeAck atomic.Bool` on `Client` with `func (c *Client) EnableWriteAck()` (`c.writeAck.Store(true)`), `func (c *Client) AckWrite()` (`c.inflight.Add(-1)`), and `func (c *Client) Drained() bool` (`c.isClosed() || (len(c.send) == 0 && (!c.writeAck.Load() || c.inflight.Load() <= 0))`).
    * Add `func (c *Client) Rehydrate(ctx context.Context) error`: check `if c.isClosed() { return apperr.Wrap(apperr.KindPolicy, ErrClientClosed) }` and delegate to `c.hub.rehydrateClient(ctx, c)`.
  * In `internal/hub/shutdown.go`:
    * In `queuesEmpty(cs []*Client)`: check `if !c.Drained() { return false }` so `drainAndCloseClients` waits until every WebSocket client with `EnableWriteAck()` has acknowledged writing its final dequeued frame via `AckWrite()`.
    * Add `ctlRehydrate` to `ctlOp` and `client *Client` to `ctlReq`, plus `func (h *Hub) rehydrateClient(ctx context.Context, c *Client) error`. In `(*Hub).handleCtl` under `case ctlRehydrate:`: if `h.closing`, reply with `apperr.Wrap(apperr.KindPolicy, ErrShuttingDown)`; if `r.client == nil || r.client.isClosed() || h.clients[r.client.id] != r.client`, reply with `apperr.Wrap(apperr.KindPolicy, ErrClientClosed)`; otherwise, while keeping `r.client` registered in `h.clients` (and `h.tui == r.client` for `ClientTUI`), call `r.client.drainAllFrames()`, call `h.rehydrate(r.client)` (updating `r.client.lastHydrate = h.cfg.Now()`), and reply `nil` on success.
  * Add unit tests in `internal/hub/permission_empty_test.go`:
    1. `TestClientDrainedTracksWriteAckAcrossSendAndDrain`: verify `c.Drained()` remains `false` after `<-c.Frames()` when `c.EnableWriteAck()` is set until `c.AckWrite()` is called, and that `c.drainAllFrames()` resets `inflight` cleanly.
    2. `TestDrainAndCloseClientsWaitsForAckWrite`: enable `c.EnableWriteAck()` on a connected client, dequeue `server.shutdown` from `c.Frames()` without calling `c.AckWrite()`, assert `drainAndCloseClients` stays blocked while `writeFrame` is simulated in flight, and assert `drainAndCloseClients` returns immediately once `c.AckWrite()` is called.
    3. `TestClientRehydrateInPlaceKeepsPendingPermissionAlive`: when `ClientTUI` is the only attached surface and a permission request is pending, fill `tui.Frames()` with live messages, call `tui.Rehydrate(ctx)`, and assert `tui.Frames()` now starts with a fresh `KindSessionHydrate` carrying `pending_permissions`, `Stats.Rehydrations` incremented, and the permission remains pending.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestClientDrained|TestDrainAndCloseClientsWaitsForAckWrite|TestClientRehydrateInPlace"`

#### Unit 2.3 (`075.003-T`): Wire `EnableWriteAck`/`AckWrite` in `transport.session.writePump` and unconditional `AdapterStopTimeout` in `internal/app/shutdown.go` (`4AFE4DA4(b)`, `AEF57AF1(a)`, `8769C9A0(c)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/session.go`
  * `internal/app/shutdown.go`
  * `internal/transport/pump_test.go`
* **Changes needed**:
  * In `internal/transport/session.go`: call `client.EnableWriteAck()` in `newSession`, and in `(*session).writePump` under `case f := <-s.client.Frames():`, call `err := s.writeFrame(ctx, f.Bytes()); s.client.AckWrite(); if err != nil { s.fail(); return }` so `c.Drained()` turns `true` only after `s.sock.Write` completes.
  * In `internal/app/shutdown.go`: define `func adapterStopTimeout(ag agent.Agent) time.Duration` (returning `st.StopTimeout()` when `st, ok := ag.(interface{ StopTimeout() time.Duration }); ok`, such as `*copilot.Adapter`'s 5s `DefaultStopTimeout`, and `0` for in-process test fakes) and pass `AdapterStopTimeout: adapterStopTimeout(a.agent)` unconditionally (never gated on `a.shut.timeouts.Force >= DefaultStopTimeout`) in `hub.ShutdownConfig` inside `(*App).supervise()` so an explicit `Force < AdapterStopTimeout` with the real adapter is rejected by `NewShutdown` validation rather than silently skipping validation, while preserving short `Force` budgets for in-process test fakes.
  * In `internal/transport/pump_test.go`: add `TestWritePumpAcksFrameOnlyAfterSocketWriteCompletes` verifying that `s.client.Drained()` is `true` only after `writePump` has written the hub frame to the WebSocket connection.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport/... ./internal/app/...`

#### Unit 2.4 (`075.004-T`): Propagate `ctx` in `App.Reissue`/`tuiReissue` and drain remote concurrently in `shutdownScenario` (`AEF57AF1(b, c)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/credentials.go`
  * `internal/app/credentials_test.go`
  * `internal/app/e2e_test.go`
* **Changes needed**:
  * In `internal/app/credentials.go`: change `mint()` to `mint(ctx context.Context) ([]transport.Issued, error)` (with `mintAll` passing `context.Background()` and `Reissue(ctx context.Context)` passing `ctx`). In `mint(ctx)`, check `if err := ctx.Err(); err != nil { return nil, err }` before `a.mintMu.Lock()` and again immediately after `a.mintMu.Lock()` (before the role loop) so a cancelled `ctx` passed to `Reissue` or `tuiReissue` returns `ctx.Err()` without revoking or rotating any credential.
  * In `internal/app/credentials_test.go`: add `TestReissueHonoursCancelledContext` verifying that calling `r.app.Reissue(cancelledCtx)` (and `r.app.tuiReissue(cancelledCtx)`) returns `context.Canceled` and leaves the previously minted approver and observer credentials valid (`dialToken` still succeeds with 101).
  * In `internal/app/e2e_test.go` (`shutdownScenario`): start draining `remote` concurrently right after `trigger(r)` (`drainDone := make(chan struct{}); go func() { defer close(drainDone); remote.drain() }()`) and wait on `<-drainDone` after `<-r.app.Done()` so the remote WebSocket peer completes the close handshake immediately during step 5 / `reap` without waiting out the 2-second `closeGrace` timer (`AEF57AF1(b)`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestReissueHonoursCancelledContext|TestE2EShutdown"`

---

### Sub-Epic 3: Shipment `069-S` — Covering Feature `076-F` (Resilient TUI reconnect, sanitization, modal anti-spoofing, and viewport hardening)

#### Unit 3.1 (`076.001-T`): Use in-place `Conn.Rehydrate` in `tui.Feed` with bounded redial retry and dropped-TUI recovery (`1DCB6108` TUI half, `E9223340(c)` feed half)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/ingress.go`
  * `internal/tui/feed.go`
  * `internal/tui/feed_test.go`
* **Changes needed**:
  * In `internal/tui/ingress.go`: add `Rehydrate(ctx context.Context) error` to the `Conn` interface (implemented by `*hub.Client` in `075.001-T`/`075.002-T`) and define `type Rehydrater interface { Rehydrate(ctx context.Context) error }`.
  * In `internal/tui/feed.go`:
    * Add `RedialRetries int` (default `defaultRedialRetries = 3`) and `RedialBackoff time.Duration` (default `defaultRedialBackoff = 25 * time.Millisecond`) to `FeedOptions`.
    * Add `rehydrateReq chan chan error` (capacity 1) on `Feed` and implement `func (f *Feed) Rehydrate(ctx context.Context) error`: check `f.started && !f.closed && f.conn != nil` under `f.mu` (returning `ErrNotConnected` immediately when false), send a buffered `reply := make(chan error, 1)` on `f.rehydrateReq` (selecting on `<-f.pumpDone` and `<-ctx.Done()`), and wait for the result on `reply` so the model can request an in-place re-hydrate through the pump goroutine. Guard `if conn != nil { conn.Close() }` in `pump`'s deferred cleanup.
    * In `(*Feed).reconnect(ctx, old)`: first attempt in-place `old.Rehydrate(dctx)` (with `dialTimeout`). When `old.Rehydrate(dctx)` succeeds, increment `f.rehydrates.Add(1)`, call `f.q.drain()`, push `item{resync: true}`, and return `old, nil` without closing `old`! Only if `old.Rehydrate(dctx)` returns an error, close `old`, clear `f.conn = nil` under `f.mu`, and call `f.redial(ctx)` which attempts `f.opt.Dial` up to `max(f.opt.RedialRetries, defaultRedialRetries)` times (waiting `RedialBackoff` on `ctx` between transient errors, stopping immediately on `ctx.Err() != nil`, `errors.Is(err, hub.ErrStopped)`, or `errors.Is(err, hub.ErrShuttingDown)`).
    * In `(*Feed).pump`: on `case <-conn.Done():`, drain remaining frames from `conn.Frames()` while recording whether `fr.Kind() == hub.KindServerShutdown` was seen; if `KindServerShutdown` was seen or `ctx.Err() != nil`, set `f.pumpErr = hub.ErrStopped` and return; otherwise (unexpected TUI client close while the hub is still running), call `f.redial(ctx)` to re-attach and re-hydrate the TUI client.
  * Add unit tests in `internal/tui/feed_test.go`:
    1. `TestOverflowRehydratesInPlaceWithoutAutoDenyingPendingPermission`: attach a `Feed` (with no remote clients), post a permission ask to the hub, overflow the feed with a message burst while the sink is stalled, release the sink, and assert `f.Stats().Rehydrates > 0` AND the pending permission is still unresolved in the hub and present on the converged model.
    2. `TestFeedRedialRetriesTransientFailureAndRecoversDroppedClient`: test both an unexpected `conn.Close()` while the hub is running and a `Dial` wrapper that fails transiently before succeeding within `RedialRetries`, asserting the feed recovers and delivers the new `session.hydrate`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestOverflow|TestFeed|TestStalled"`

#### Unit 3.2 (`076.002-T`): Sanitize non-`Cf` invisible blank runes (`U+3164`, `U+2800`, `U+034F`, `Co`, variation selectors) and add `isBlankSummary` (`E9223340(d)` sanitize half)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/tui/sanitize.go`
  * `internal/tui/sanitize_test.go`
* **Changes needed**:
  * In `internal/tui/sanitize.go`: extract `func isHiddenRune(r rune) bool` returning `true` for `unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) || unicode.Is(unicode.Variation_Selector, r) || r == '\u2800' || r == '\u3164' || r == '\u115f' || r == '\u1160' || r == '\uffa0' || r == '\u034f' || r == '\u17b4' || r == '\u17b5'`, and use `isHiddenRune(r)` in `Sanitize(s)`. Add `func isBlankSummary(s string) bool` returning `true` when every decoded rune in `s` is `unicode.IsSpace(r)`, `isHiddenRune(r)`, or `r == utf8.RuneError` (covering both invalid UTF-8 bytes and `U+FFFD` replacement characters emitted by `copilot.sanitizeBound` for `Cf`/control characters).
  * In `internal/tui/sanitize_test.go`: add test cases in `TestSanitizeNeutralisesTerminalEscapeInjection` for `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+2800`, `U+034F`, `U+17B4`, `U+17B5`, `U+E000` (`Co`), and `U+FE0F` (`Variation_Selector`), and add `TestIsBlankSummaryRejectsInvisibleAndReplacementOnlyText`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestSanitize|TestIsBlankSummary"`

#### Unit 3.3 (`076.003-T`): Enforce `isBlankSummary` deny-only in `newPermission` and wrap-aware 2-space summary indentation (`E9223340(d, e)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/permission.go`
  * `internal/tui/permission_test.go`
  * `internal/tui/testdata/modal.golden`
* **Changes needed**:
  * In `internal/tui/permission.go`:
    * In `newPermission`: when `strings.TrimSpace(p.Summary) == ""`, keep `out.Summary, out.Empty = emptyRequestText, true`; when `strings.TrimSpace(p.Summary) != ""` and `isBlankSummary(p.Summary)` is true, set `out.Summary = out.Summary + "\n" + emptyRequestText` and `out.Empty = true` so invisible-only or `U+FFFD`-only summaries display both their `<U+XXXX>` markers and `(empty request - deny only)` and cannot be approved.
    * In `syncModal` / `(permission).body()`: wrap the `"kind: " + p.Kind` and `"tool call: " + p.ToolCallID` header lines to `m.modalInnerWidth()`, and wrap `p.Summary` to `max(m.modalInnerWidth()-2, 1)` before prefixing every resulting summary line with `"  "` so neither newline-separated nor soft-wrapped summary lines can ever spoof column-0 `kind:` or `tool call:` headers. Update `internal/tui/testdata/modal.golden` to reflect the 2-space indentation on `  git status --short`.
  * In `internal/tui/permission_test.go`: extend `TestBlankSummaryIsDenyOnly` with `"\u3164"`, `"\u2800"`, `"\u200b"`, `"\ufffd"`, and `"\ue000\n\t"`, and add `TestPermissionSummaryIsIndentedAgainstNewlineAndSoftWrapHeaderSpoofing`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestModal|TestBlankSummary|TestPermissionSummary"`

#### Unit 3.4 (`076.004-T`): Add modal-generation pure-Elm `ApprovalDelay`, disarm `m.armed` on mouse/resync/hydrate, and trigger `Rehydrater.Rehydrate` on stuck `resync` (`E9223340(c, f)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/model.go`
  * `internal/tui/permission.go`
  * `internal/tui/permission_test.go`
* **Changes needed**:
  * In `internal/tui/model.go` and `internal/tui/permission.go`:
    * **Pure-Elm `ApprovalDelay` with nonzero `DefaultApprovalDelay` and modal-generation token (`E9223340(f)`)**: export `const DefaultApprovalDelay = 150 * time.Millisecond` in `internal/tui/model.go`, and add `ApprovalDelay time.Duration`, `NoApprovalDelay bool`, and `After func(time.Duration) <-chan time.Time` to `Options`. In `New(o Options)`, when `o.NoApprovalDelay || o.ApprovalDelay < 0`, set `m.approvalDelay = 0` (explicit zero-delay path for synchronous unit tests); when `o.ApprovalDelay == 0 && !o.NoApprovalDelay`, set `m.approvalDelay = DefaultApprovalDelay` (`150 * time.Millisecond`, so `startUI` in `internal/app/ui.go:119` enforces the 150ms delay in production automatically); and default `m.after` to `time.After` when `o.After == nil`. Define `type ApprovalReadyMsg struct{ RequestID string; ModalGen uint64 }`, and add `modalReady bool` and monotonic `modalGen uint64` on `Model`. Whenever `syncModal` opens a new modal or resets the current modal (`p.RequestID != m.modalFor || reset`, including on `tea.WindowSizeMsg` resize or re-hydrate), increment `m.modalGen++`. When `m.approvalDelay <= 0`, `syncModal` sets `m.modalReady = true` immediately. When `m.approvalDelay > 0`, `syncModal` sets `m.modalReady = false` whenever `p.RequestID != m.modalFor || reset`, and `Update` (across `tea.WindowSizeMsg`, `FrameMsg`, and `ActionResultMsg` advancing the head request) returns a `tea.Cmd` (`m.approvalTimerCmd(p.RequestID, m.modalGen)`) that waits on `m.after(m.approvalDelay)` and returns `ApprovalReadyMsg{RequestID: id, ModalGen: gen}`. On `ApprovalReadyMsg`, verify both `len(m.pending) > 0 && m.pending[0].RequestID == msg.RequestID && m.modalFor == msg.RequestID` AND `m.modalGen == msg.ModalGen` before setting `m.modalReady = true` (rejecting stale timers from earlier resets of the same request). In `onModalKey("y")` / `approvable()`, require `m.modalReady` (setting `m.notice = "wait a moment before approving"` if pressed before `modalReady`), while keeping `case "n":` (deny) immediate at all times.
    * **Model-initiated re-hydrate on stuck `resync` & `KindServerShutdown` gap handling (`E9223340(c)`)**: define `type ResyncFailedMsg struct{ Err error }` and add `resyncRequested bool` on `Model` (cleared in `applyHydrate` on a valid snapshot and on `ResyncFailedMsg`). In `Model.onFrame`: if `f.Kind() == hub.KindServerShutdown`, set `m.shutdown = true` and `m.resync = false` immediately (even if `m.cursor.Apply(f)` returned `hub.ActionRehydrate` on a stale client). Otherwise, when `onFrame` enters `m.resync = true` (on `hub.ActionRehydrate` or an undecodable `hydrateBody`) and `!m.resyncRequested`, if `r, ok := m.ingress.(Rehydrater); ok`, set `m.resyncRequested = true` and return a `tea.Cmd` that calls `r.Rehydrate(ctx)` bounded by `m.timeout` (returning `ResyncFailedMsg{Err: err}` on error). Also clear `m.armed = ""` on `ResyncMsg` and in `applyHydrate`.
  * In `internal/tui/permission_test.go`: update `modalModel`, `twoPending`, and the direct synchronous modal tests in `permission_test.go` to opt into the explicit zero-delay path (`NoApprovalDelay: true`), and add `TestApprovalDelayBlocksApproveUntilReadyMsgAndRejectsStaleModalGenAfterResize` (constructing `New(Options{Ingress: ing, Renderer: asciiRenderer(), After: fakeAfter})` with `ApprovalDelay` left at `0` to verify `DefaultApprovalDelay = 150ms` is active by default in production, and asserting that an older `ApprovalReadyMsg` with a stale `ModalGen` arriving after a `tea.WindowSizeMsg` resize reset is ignored without unlocking approval early) and `TestCursorGapAndBadHydrateTriggerModelRehydrateAndClearArmed`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestApprovalDelay|TestCursorGap"`

#### Unit 3.5 (`076.005-T`): Add live transcript entry cap, delta-invalidated wrap cache, display-width prompt truncation, and mouse wheel initialization/scrolling (`E9223340(b, g)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/messages.go`
  * `internal/tui/model.go`
  * `internal/tui/model_test.go`
* **Changes needed**:
  * In `internal/tui/messages.go`: add unexported cache fields `wrapped string` and `wrapWidth int` to `entry`.
  * In `internal/tui/model.go`:
    * Add `MaxEntries int` to `Options` (defaulting `m.maxEntries` to `hub.DefaultMaxEntries = 1000` when `<= 0`).
    * **Live entry cap & delta-invalidated wrap cache (`E9223340(b)`)**: add `capEntries()` called after appending to `m.entries` in `applyMessage`, `applyToolStarted`, `applyToolEnded`, and `applyError`, evicting the oldest entries beyond `m.maxEntries` and adding the evicted count to `m.evicted`. Whenever `applyMessage` mutates an existing entry at index `i` (such as appending a streaming delta under `case isDelta && i >= 0:` or updating `Text` under `case i >= 0:`), clear `m.entries[i].wrapped = ""` and `m.entries[i].wrapWidth = 0` while leaving all other unchanged entries' cached strings intact. In `applyToolEnded`, also backfill `if m.entries[i].Tool == "" && name != "" { m.entries[i].Tool = name }`, set `m.entries[i].Truncated = m.entries[i].Truncated || p.Truncated`, and clear `m.entries[i].wrapped = ""` and `m.entries[i].wrapWidth = 0`. In `refreshTranscript` / `renderTranscript`, compute and store `m.entries[i].wrapped = m.wrap(entryText(m.entries[i]), m.vp.Width)` and `m.entries[i].wrapWidth = m.vp.Width` only when `m.entries[i].wrapped == "" || m.entries[i].wrapWidth != m.vp.Width`, reusing `m.entries[i].wrapped` otherwise.
    * **Mouse reporting initialization, display-width prompt truncation & mouse wheel scroll (`E9223340(g)`)**: in `Model.Init()`, return `tea.EnableMouseCellMotion` so Bubble Tea enables mouse reporting when the TUI starts. In `Model.View()`, replace `if rs := []rune(prompt); len(rs) > m.width` with display-width truncation using `lipgloss.Width(prompt) > m.width` (trimming runes from the front of the sanitized input tail until `lipgloss.Width("> ..." + tail) <= m.width` when `m.width > 5`). In `Model.Update`, handle `tea.MouseMsg` (`tea.MouseButtonWheelUp` and `tea.MouseButtonWheelDown`), first clearing `m.armed = ""` and then scrolling `m.modal` (and calling `m.noteSeen()`) when `len(m.pending) > 0`, scrolling credentials (`m.scrollCredentials(-1)` / `scrollCredentials(1)`) when `m.credOpen`, or scrolling `m.vp` otherwise.
  * In `internal/tui/model_test.go`: add `TestLiveTranscriptCapAndWrapCacheInvalidatesMutatedDeltaEntry` (asserting consecutive streaming deltas at the same viewport width render newly appended text while unchanged prior entries reuse their cached render) and `TestWideRunePromptTruncatesToDisplayWidthAndMouseInitAndWheelScrolls` (asserting `m.Init() != nil` returns the `tea.EnableMouseCellMotion` command and `tea.MouseMsg` wheel events scroll modal/credentials/transcript).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui/...`

## Dependency Graph

### Shipment-Level DAG

* `067-S` (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`, under explicit operator P-001 policy exception over `062-S`, code baseline `066-S` `65780da`; Covering Feature `074-F`: `074.001-T -> 074.002-T -> 074.003-T -> 074.004-T`)
  * `<-blocks-- 068-S` (Covering Feature `075-F`: `075.001-T -> 075.002-T -> 075.003-T -> 075.004-T`)
    * `<-blocks-- 069-S` (Covering Feature `076-F`: `076.001-T -> 076.002-T -> 076.003-T -> 076.004-T -> 076.005-T`)

### Intra-Shipment Task Dependencies

* **Within `067-S` (`074-F`)**: `074.001-T` -> `074.002-T` -> `074.003-T` -> `074.004-T` (serialized because `074.002-T` and `074.003-T` both touch `internal/hub/hub.go`).
* **Within `068-S` (`075-F`)**: `075.001-T` -> `075.002-T` -> `075.003-T` -> `075.004-T` (serialized because `075.002-T` and `075.003-T` build on `075.001-T`'s `Client` write-ack helpers).
* **Within `069-S` (`076-F`)**: `076.001-T` -> `076.002-T` -> `076.003-T` -> `076.004-T` -> `076.005-T` (serialized across `internal/tui/permission.go` and `internal/tui/model.go`).

## Decisions and Rationale

1. **Three single-feature shipments (`067-S..069-S`)**: Satisfies Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and orders the hub `Client.Rehydrate` API (`068-S`) before its `internal/tui` consumer (`069-S`).
2. **Protocol decision for async operator `Send`/`Abort` failures (`550437DB`)**: Because `OpPrompt` and `OpAbortTurn` run off-hub on operator lane workers after `Client.Do` has already returned `Accepted: true`, non-shutdown failures in `callAgent` are sent over `actionFailed` to the hub select loop, which folds and publishes a `KindError` (`code: "operator_action_failed"`) envelope with a fixed sanitized message. This uses the existing `KindError` wire vocabulary without leaking prompt secrets or triggering a false degraded-mode re-hydrate.
3. **In-place `(*hub.Client).Rehydrate` (`1DCB6108`)**: Executing `ctlRehydrate` inside a single hub step drains `c.send` via `c.drainAllFrames()` and enqueues a fresh `session.hydrate` while keeping `c` registered in `h.clients` (`h.tui == c`), closing the no-TUI window during `tui.Feed` overflow recovery.
4. **Lock-free write-completion tracking (`4AFE4DA4(b)` / `AEF57AF1(a)`)**: Using `EnableWriteAck()` / `AckWrite()` / `Drained()` (`writeAck atomic.Bool` and `inflight atomic.Int64`, incremented in `c.sendFrame` before sending to `c.send` and decremented in `c.AckWrite` and `c.drainOneFrame`/`c.drainAllFrames`) on `hub.Client` ensures `drainAndCloseClients` (`c.Drained()`) waits for `transport.session.writePump` to finish `writeFrame` on the final dequeued frame before closing `c.Done()`, with zero race window even on direct `select` handoff and zero overhead for raw `<-c.Frames()` consumers in unit tests.

## Risks and Caveats

* **`TestHubPackageHasNoBareChannelSends` AST guard (`internal/hub/degraded_test.go`)**: Every channel send in `internal/hub/*.go` production files MUST be a `select` communication clause (never a bare `ch <- v` statement). All new sends in `internal/hub` (`actionFailed`, `ctlRehydrate`, `c.sendFrame`, etc.) must use `trySend` or a `select` statement with `<-ctx.Done()` / `<-h.done`.
* **Golden file coupling (`internal/tui/testdata/modal.golden`)**: Indenting `permission.Summary` by 2 spaces in `(permission).body()` requires updating `modal.golden` in `076.003-T` so `TestModalShowsRequestAndApproveGoesThroughIngress` passes.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Adds `(*hub.Client).Rehydrate(ctx)`, `(*hub.Client).EnableWriteAck()`, `AckWrite()`, `Drained()`, `ErrCodeOperatorActionFailed`, `ShutdownConfig.AdapterStopTimeout`, `copilot.DefaultStopTimeout`, `tui.Conn.Rehydrate(ctx)`, `(*tui.Feed).Rehydrate(ctx)`, `tui.DefaultApprovalDelay`, `tui.ApprovalReadyMsg`, `tui.ResyncFailedMsg`, and `tui.Options.{MaxEntries, ApprovalDelay, NoApprovalDelay, After}`.
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Modifies permission modal sanitization (`isHiddenRune`, `isBlankSummary`), wrap-aware summary indentation against header spoofing, aborted-idle pending-permission settlement (`permReasonTurnAborted`), and empty-broadcast-set pending-connect handling.
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive action.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — No new external dependencies in `go.mod`.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Modifies hub operator lane lifecycle, WebSocket write-pump drain synchronization at shutdown step 5, and TUI feed overflow/reconnect recovery.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `067-S` (`074-F`)**: Hub fold, publish, operator lane lifecycle, and copilot adapter stop-recheck behavior. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/... ./internal/agent/copilot/... ./internal/app/...` and `go run ./tools/gatecheck write-path --self-test --root .`.
* **Shipment `068-S` (`075-F`)**: Hub shutdown broadcast/drain, transport write-pump completion tracking, `App.Reissue` context propagation, and in-place `Client.Rehydrate`. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/... ./internal/transport/... ./internal/app/...`.
* **Shipment `069-S` (`076-F`)**: TUI feed reconnect/redial, modal sanitization/indentation, transcript cap/cache, prompt width truncation, and mouse scroll. Verified headlessly via `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui/... ./internal/app/...`; for the `tui-render` release checkpoint, Ship exercises the real `intercom.exe` binary in a Win32 console (or carries forward / updates live TUI render evidence per the release-observability instructions). Rollback for each shipment is a clean `git revert` of its merge commit.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was triggered by three signals: (1) internal package contract additions across `internal/hub`, `internal/agent/copilot`, `internal/transport`, and `internal/tui`, (2) permission-arbitration and TUI terminal-sanitization security controls (`1DCB6108`, `550437DB`, `E9223340(d, e, f)`, `4AFE4DA4(c)`), and (3) concurrency-sensitive changes in `Hub.Run` lane worker joining, `drainAndCloseClients` write-completion tracking, and `tui.Feed` in-place re-hydration.

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, V, VI, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`

### 2. Protected Invariants

1. **No bare channel sends in `internal/hub` (`TestHubPackageHasNoBareChannelSends`)**: Every channel send in `internal/hub/*.go` production files must live inside a `select` statement (or use `trySend`), never a bare `ch <- v`.
2. **Zero new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives**: `go run ./tools/gatecheck write-path --self-test --root .` must exit 0 after every task.
3. **No direct SDK or adapter imports in `internal/tui` (`TestPackageHasNoDirectAgentOrSDKReach`)**: `internal/tui` non-test files must not import `copilot-sdk` or `internal/agent/copilot`.
4. **Delivery-uncertain failed `Send` idle accounting (`TestFailedPromptSendKeepsTheSessionBusy`, `TestLateSendFailureOfAStartedPromptDoesNotConsumeALaterPromptsCount`)**: Publishing `KindError` (`operator_action_failed`) when an invoked `Agent.Send` fails in `callAgent` must NOT retract `h.promptsPending`; retraction remains restricted to prompts that provably never reached `Agent.Send` (`!invoked` or dropped in `stopLanes`).
5. **Single choke-point approval disarm and immediate deny (`TestAccidentalTypingCannotApprove`, `TestDenyIsImmediateAndGoesThroughIngress`)**: Any key other than `Enter` (as well as `tea.MouseMsg`, `ResyncMsg`, and `applyHydrate`) disarms `m.armed`, and `n` (deny) remains immediately available even when `ApprovalDelay > 0` or the modal is not yet scrolled to the end.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Harden `Fold` (`applyDelta`, `applyToolEnded`), `Hub.publish` / `rehydrate`, `Hub.Run` operator lane joining, `operator_action_failed` error envelope publication, aborted-idle permission settlement, and `Adapter.recheckForce` / `ShutdownConfig.AdapterStopTimeout` (`067-S` / `074-F`). | `local edit` (hub & adapter hardening) | `moderate` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `067-S` merge commit | `planned` |
| `PA-02` | Sweep closed remotes at `intake`, honor pending connects in `reapEmptyBroadcastSet`, deliver `server.shutdown` to stale clients, track WebSocket frame write completion (`EnableWriteAck`/`AckWrite`/`Drained`) in `drainAndCloseClients`, wire `AdapterStopTimeout` in `internal/app/shutdown.go`, propagate `ctx` in `App.Reissue`, and add in-place `(*hub.Client).Rehydrate` (`068-S` / `075-F`). | `local edit` (hub, transport & app hardening) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `068-S` merge commit | `planned` |
| `PA-03` | Wire in-place `Conn.Rehydrate` and bounded redial retry in `tui.Feed`, sanitize non-`Cf` invisible runes (`U+3164`, `U+2800`, `U+034F`, `Co`, variation selectors), wrap-and-indent permission summaries, add pure-Elm `ApprovalDelay`, cap live transcript entries with wrap caching, add display-width prompt truncation, and support mouse scroll (`069-S` / `076-F`). | `local edit` (TUI security & UX hardening) | `moderate` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `069-S` merge commit | `planned` |

### 4. Deepened Verification, Rollback, and Review-Gate Requirements

* **AST and structural guards**: Run `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub -run TestHubPackageHasNoBareChannelSends` after every `internal/hub` task and `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui -run TestPackageHasNoDirectAgentOrSDKReach` after every `internal/tui` task.
* **Review-gate capability requirement (P-012)**: Plan review (Step 4) MUST spawn the always-on and triggered cross-model persona subagents (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, `Security Lens Reviewer`) plus `Adversarial Review`, and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* **Unresolved operator decisions**: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating all P1, P2, and P3 findings from the multi-agent and inline persona review pass directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01`..`PA-03` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Refinement) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (3 P2, 2 P3) | **PASS** |
| **Go Reviewer** | Inline Rubric Fallback (declared degradation on empty subagent response) | Caller (`gemini-4-argon`) | PASS (0 P0, 0 P1) | **PASS** |
| **Scope Boundary Auditor** | Inline Rubric Fallback (declared degradation on empty subagent response) | Caller (`gemini-4-argon`) | PASS (0 P0, 0 P1) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | PASS (6 high-confidence solutions verified) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | FAIL -> Remediated (3 P1, 3 P2) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (3 P2, 2 P3) | **PASS** |
| **Concurrency Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (2 P1, 2 P2) | **PASS** |
| **Correctness Reviewer** | Inline Rubric Fallback (declared degradation on empty subagent response) | Caller (`gemini-4-argon`) | PASS (0 P0, 0 P1) | **PASS** |
| **Adversarial Review** | Inline Rubric Fallback (declared degradation on empty subagent response) | Caller (`gemini-4-argon`) | PASS (0 P0, 0 P1) | **PASS** |

### Findings and Applied Remediations

1. **[P1 — Remediated] Unit 2.1 (`075.001-T`), Unit 2.2 (`075.002-T`), & Unit 2.3 (`075.003-T`) Hub-to-Transport Write-Completion Contract (`EnableWriteAck` / `AckWrite` / `Drained`)**:
   * *Finding*: When `transport.session.writePump` is parked in the blocking `select` on `case f := <-s.client.Frames():`, Go's channel runtime hands an incoming `server.shutdown` frame directly to the waiting receiver without incrementing `len(c.send)` (`qcount == 0`), leaving a window before `BeginWrite()` where both `len(c.send) == 0` and `writing == 0` hold. In addition, the deliberation and Requirements Trace used `EnableWriteAck`/`AckWrite` while Unit 2.2 used `BeginWrite`/`EndWrite`.
   * *Remediation applied*: Unified the deliberation and plan on `EnableWriteAck()`, `AckWrite()`, `Drained()`, `sendFrame(f)`, `drainOneFrame()`, and `drainAllFrames()` using `writeAck atomic.Bool` and `inflight atomic.Int64` on `hub.Client`, where `c.sendFrame(f)` increments `c.inflight.Add(1)` BEFORE `select { case c.send <- f: ... }` and `c.AckWrite()` decrements `c.inflight.Add(-1)` only after `writeFrame` returns in `writePump` (with `c.drainOneFrame`/`c.drainAllFrames` decrementing `inflight` for hub-drained frames). This eliminates the direct-handoff `select` race completely with zero polling delay.
2. **[P1 — Remediated] Unit 1.3 (`074.003-T`) `runOperatorLanes` `!invoked` Retraction Select and Aborted-Idle Permission Settlement**:
   * *Finding*: Adding `case <-ctx.Done():` to the `h.promptFailed` send select inside `if !invoked` in `runOperatorLanes` would randomly drop `promptFailed` ~50% of the time during `stopLanes()` because `!invoked` only holds when `ctx.Done()` is already closed, leaking `h.promptsPending`. Furthermore, on `agent.KindIdle` with `Aborted: true`, `handleEvent` must also drain any asks queued in `h.asks` and call `h.settleDecided(e, permReasonTurnAborted)` on `permDecided` entries (revoking unconsumed approvals via `Outcome.Revoke()`) before resolving `h.perm.unresolved()`.
   * *Remediation applied*: Replaced the `h.promptFailed` send select with `trySend(h.promptFailed, struct{}{})` (capacity `operatorPromptQueue + 1`), placed `wg.Add(2)` after `if h.cfg.Agent == nil { return }`, and updated `handleEvent`'s aborted-idle branch to drain `h.asks`, settle `permDecided` entries via `h.settleDecided(e, permReasonTurnAborted)`, and resolve `h.perm.unresolved()` entries with `permReasonTurnAborted`.
3. **[P1/P2 — Remediated] Units 3.1–3.5 (`076.001-T`..`076.005-T`) Pure-Elm `ApprovalDelay`, `ResyncFailedMsg` Recovery, `KindServerShutdown` Gap Handling, Soft-Wrap Indentation, and Principle II Test-First Pairing**:
   * *Finding*: (a) Reading `time.Now()` inside `onModalKey` violates the pure Elm boundary of `tui.Model`; (b) setting `m.resyncRequested = true` without a failure message (`ResyncFailedMsg`) can permanently suppress future resync attempts if `Rehydrate` errors; (c) a stale client receiving `KindServerShutdown` across a sequence gap would trigger `ActionRehydrate` instead of setting `m.shutdown = true`; (d) `isBlankSummary` needed to cover `r == utf8.RuneError` (`U+FFFD` emitted by `copilot.sanitizeBound` for `Cf`/control runes) and `U+034F`, and `(permission).body()` needed to wrap `p.Summary` to `m.modalInnerWidth() - 2` before prefixing `"  "` so soft-wrapped lines cannot spoof column-0 headers; and (e) every production code unit must include its own red-phase `_test.go` assertions (Principle II).
   * *Remediation applied*: Replaced `Now` in `ApprovalDelay` with a pure-Elm `tea.Cmd` (`After func(time.Duration) <-chan time.Time` emitting `ApprovalReadyMsg{RequestID: id}` that sets `m.modalReady = true`), added `ResyncFailedMsg` to clear `m.resyncRequested` on failure, handled `f.Kind() == hub.KindServerShutdown` directly in `onFrame` before `ActionRehydrate`, disarmed `m.armed` on `tea.MouseMsg`/`ResyncMsg`/`applyHydrate`, extended `isHiddenRune` and `isBlankSummary` (`r == utf8.RuneError` for any size), wrapped `p.Summary` to `max(m.modalInnerWidth()-2, 1)` before indenting each line by `"  "`, wired `AdapterStopTimeout: copilot.DefaultStopTimeout` in `internal/app/shutdown.go` (`075.003-T`), and structured `068-S` (`075.001-T`..`075.004-T`) and `069-S` (`076.001-T`..`076.005-T`) so every single unit modifies `<= 3` files and pairs its production changes with its own red-phase `_test.go` tests.
4. **[P2 — Remediated] Unit 2.1 (`075.001-T`) Post-`serviceHydrations` `reapEmptyBroadcastSet` and Abandoned `Connect` Wake**:
   * *Finding*: If a pending connect in `h.pending` deferred `reapEmptyBroadcastSet()` and was then discarded as abandoned or refused in `serviceHydrations()` at the top of the next iteration, `Hub.Run` would enter `select` without re-running `reapEmptyBroadcastSet()`.
   * *Remediation applied*: Added `if !h.hasPendingConnect() { h.reapEmptyBroadcastSet() }` immediately after `h.serviceHydrations()` and at the end of the loop in `Hub.Run`, skipped already-abandoned requests in `admitPending`, and called `trySend(h.membershipWake, struct{}{})` after `req.abandoned.Store(true)` in `Hub.Connect`.
5. **[PR #201 Copilot Review — Remediated] Units 1.4, 2.1, 2.2, 2.3, 3.4, & 3.5 (`074.004-T`, `075.001-T`..`075.003-T`, `076.004-T`..`076.005-T`)**:
   * *Findings*: (1) `deliverTerminal` in Unit 2.1 must clear `c.stale = false` (and mark `c.terminal = true` so `serviceHydrations`, `markAllStale`, and subsequent non-terminal `broadcast` calls skip `c`) when enqueuing `server.shutdown` so a subsequent hub-loop iteration cannot drain and replace the terminal `server.shutdown` frame with a `session.hydrate` snapshot; (2) `(*Client).Rehydrate(ctx)` (`internal/hub/client.go`) must be introduced in Unit 2.2 (`075.002-T`) alongside its hub control handler `(*Hub).rehydrateClient` (`internal/hub/shutdown.go`) rather than in Unit 2.1 (`075.001-T`), keeping every task self-compilable and within `<= 3` files; (3) `internal/app/shutdown.go` in Unit 2.3 (`075.003-T`) must pass `AdapterStopTimeout: adapterStopTimeout(a.agent)` (`(*copilot.Adapter).StopTimeout()` from Unit 1.4 / `074.004-T`, returning `0` for in-process test fakes) unconditionally rather than gating on `Force >= DefaultStopTimeout`, so an explicit `Force < 5s` with the real adapter is rejected by `NewShutdown` validation; (4) `ApprovalReadyMsg` in Unit 3.4 (`076.004-T`) must carry a monotonic `ModalGen uint64` incremented whenever `syncModal` opens or resets a modal (such as on `tea.WindowSizeMsg` resize) and reject stale generations; (5) `Model.Init()` in Unit 3.5 (`076.005-T`) must return `tea.EnableMouseCellMotion` so Bubble Tea enables mouse reporting during initialization; and (6) `applyMessage` in Unit 3.5 (`076.005-T`) must invalidate `m.entries[i].wrapped = ""` and `m.entries[i].wrapWidth = 0` when appending a streaming delta to an existing entry while retaining unchanged entries' cached strings.
   * *Remediation applied*: Updated Units 1.4, 2.1, 2.2, 2.3, 3.4, and 3.5 in the plan, deliberation, session memory, and backlog tasks (`074.004-T`, `075.001-T`, `075.002-T`, `075.003-T`, `076.004-T`, `076.005-T`) with all six fixes and corresponding red-then-green unit test assertions while keeping every task strictly `<= 3` files.
6. **[PR #201 Copilot Round 2 Review — Remediated] Units 1.3, 1.4, & 3.4 (`074.003-T`, `074.004-T`, `076.004-T`)**:
   * *Findings*: (1) `ApprovalDelay` in Unit 3.4 (`076.004-T`) defaulted to zero so `startUI` in `internal/app/ui.go:119` never activated the delay in production; (2) draining `h.asks` once in Unit 1.3 (`074.003-T`) when `agent.Event{Kind: agent.KindIdle, Aborted: true}` is processed misses late permission asks from the aborted turn that arrive on `h.asks` *after* the idle event (`internal/agent/agent.go:143-145` leaves ask/event ordering undefined); and (3) auto-populating `ShutdownConfig.AdapterStopTimeout` in Unit 1.4 (`074.004-T`) causes `TestShutdownThatCannotConfirmForceStopReportsOrphanRisk` in `internal/agent/copilot/shutdown_hub_test.go:298` (which pairs a `1s` `StopTimeout` adapter with `Force: 300ms`) to fail in `NewShutdown` unless migrated within `074.004-T`.
   * *Remediation applied*: (1) Exported `DefaultApprovalDelay = 150 * time.Millisecond` in `internal/tui/model.go` and made `tui.New(o Options)` select `DefaultApprovalDelay` when `o.ApprovalDelay == 0 && !o.NoApprovalDelay` (so `startUI` in `internal/app/ui.go:119` enforces the 150ms delay in production automatically) while adding `Options.NoApprovalDelay` (`or ApprovalDelay < 0`) as the explicit zero-delay path for synchronous unit tests in `internal/tui/permission_test.go`; (2) added `turnAborted bool` on `Hub` and `(*Hub).intakeAsk(ask)` in `internal/hub/hub.go` (set to `true` on `agent.KindIdle` with `Aborted: true` and cleared on `agent.KindTurnStart`) so any late `PermissionAsk` arriving after an aborted idle is rejected immediately with `agent.DecisionUnavailable` without registering in `h.perm`, and added the `idle-before-ask` regression case to `TestAbortedIdleEventSettlesDecidedResolvesPendingAndRejectsLateAsks`; and (3) changed `074.004-T`'s test file from `internal/agent/copilot/adapter_review_test.go` to `internal/agent/copilot/shutdown_hub_test.go` (keeping `074.004-T` at 3 files) and included the migration of `TestShutdownThatCannotConfirmForceStopReportsOrphanRisk` (`r.h.ad.cfg.StopTimeout = 200 * time.Millisecond` so `Force: 300ms >= StopTimeout: 200ms`) alongside `TestStopAfterFailedStartRecheckForceIsNonBlocking` and `TestNewShutdownValidatesAdapterStopTimeoutCoupling`.

