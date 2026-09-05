# Architecture Design: intercom-go Backend

> **Revision 2 — 2026-09-04. GOVERNING.**
>
> This revision **supersedes revision 1** ("Architecture Design: Copilot Remote
> Multiplexer (Backend)"), which specified `coder/acp-go-sdk`, an
> application-owned ACP broker over TCP, a headless Copilot CLI managed by this
> process, and an iOS remote client. **All four of those are retired.**
> Revision 1 is preserved verbatim in git history at commit `2a54791`.
>
> Governing decision:
> [`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`](../decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md).
> Where this document and any earlier artifact disagree, the governing
> decision wins.

## 0. What changed from revision 1

| Revision 1 (retired) | Revision 2 (governing) |
|---|---|
| `coder/acp-go-sdk` as the protocol contract | Official `github.com/github/copilot-sdk/go`, pinned `v1.0.11` |
| Application-owned **ACP broker goroutine** holding a TCP connection to the CLI | **No broker.** The SDK owns client/session lifecycle and its own transport |
| Application manages a **headless Copilot CLI** | The **SDK** spawns and manages the Copilot CLI; intercom-go never does |
| Two ports allocated (ACP + WS) | **One** port allocated (HTTP/WS). The SDK's transport is not an externally allocated port |
| **iOS** remote client | **SPA** remote client (framework out of scope) |
| `gorilla/websocket` | WebSocket library choice deferred to implementation; note the SDK already depends on `coder/websocket` |
| In-memory `SessionState` is the system of record | Backend projection is a **cache**; the SDK runtime is the durable event source |
| Slack as an operator surface | **Slack fully retired** |
| Go 1.21+ | **Go 1.24+** (SDK floor) |

---

## 1. System Overview

intercom-go is a per-workspace backend that brokers a GitHub Copilot agent
session between **two co-equal operator surfaces**: a local Bubble Tea
terminal application and a remote SPA reached over a Microsoft Dev Tunnel.

The backend holds the **canonical session state**. Both surfaces are views
onto it. Neither surface talks to the agent directly.

### 1.1 The two boundaries (do not conflate them)

```text
+==========================================================================+
|  APPLICATION BOUNDARY - designed, owned, and tested by intercom-go       |
|                                                                          |
|   Bubble Tea TUI --+                                                     |
|   (local)          |                                                     |
|                    +--> OperatorGateway --> Canonical SessionState       |
|   SPA -------------+    . event envelopes    (single source of truth     |
|    ^                    . hydrate/reconnect   for both surfaces)         |
|    |                    . first-responder               |                |
|  WebSocket              . backpressure                  |                |
|    ^                                                    v                |
|  HTTP/WS listener                              Agent adapter (ACL)       |
|    ^                                           copilot.Client / Session  |
|  Microsoft Dev Tunnels                                  |                |
+=========================================================+================+
                                                          |
   ---------------------  SDK BOUNDARY  -------------------+----------------
                                                           v
        SDK-INTERNAL - NOT owned, NOT re-implemented, NOT depended upon
              JSON-RPC (stdio / TCP)  -->  Copilot CLI (server mode)
```

**Normative statement.** The GitHub Copilot SDK internally communicates with
and manages a Copilot CLI server over JSON-RPC. That is an **SDK
implementation detail**. intercom-go:

* does **not** implement an ACP broker;
* does **not** implement or parse any ACP wire protocol;
* does **not** manage a headless Copilot CLI lifecycle directly;
* does **not** depend on the JSON-RPC framing, method names, or error codes.

This is enforced by the SDK's own design: the codec lives in
`github.com/github/copilot-sdk/go/internal/jsonrpc2` and is **unimportable**
by host applications. `RuntimeConnection` is a sealed interface — hosts select
a provided transport but cannot implement one.

The **one** legitimate application-level concern in this area is *runtime
location*: the Go SDK does not bundle the Copilot CLI, so the operator may
need to tell the SDK where it lives (`[copilot].cli_path` →
`StdioConnection.Path` / `COPILOT_CLI_PATH`). Configuring *where the SDK finds
its runtime* is not the same as *owning the runtime's lifecycle*.

### 1.2 Topology

One isolated Go process per workspace, spawned by autoharness. This gives 1:1
parity between a terminal tab, a repository context, and a UI state, and
isolates workspace lifecycles completely.

---

## 2. Technology Stack

| Concern | Choice | Notes |
|---|---|---|
| Language | **Go 1.24+** | Hard floor: `github.com/github/copilot-sdk/go` declares `go 1.24`. The repository is currently at `go 1.22` and **must be raised**. |
| Agent SDK | `github.com/github/copilot-sdk/go` **v1.0.11** (tag `go/v1.0.11`, commit `a550258d5c37bd662197536992a23d633bfe5804`) | Only stable Go release; GA with SemVer. See §7 for risk controls. |
| TUI | `charmbracelet/bubbletea` + `lipgloss` + `bubbles/viewport` | Elm architecture decouples terminal rendering from asynchronous event streams. `viewport` is required so continuous token streaming does not destroy native scrollback. |
| WebSocket | **Deferred to implementation** | Note the SDK already pulls in `coder/websocket`; reusing it avoids a second WS implementation in the binary. Whatever is chosen must expose ping/pong control so an idle tunnel is not dropped while the agent is "thinking". |
| Tunnel | Microsoft Dev Tunnels | Outbound TLS relay; no inbound firewall rules. See §6 for the trust-boundary caveat. |
| Config | `BurntSushi/toml` (already shipped) | See §8. |

**Retired:** `coder/acp-go-sdk`, Slack SDKs, Socket Mode, any ACP type
matchers.

---

## 3. Per-Workspace Process Lifecycle (autoharness integration)

1. **Port allocation.** autoharness identifies **one** free local TCP port for
   the Go HTTP/WS listener. *(Revision 1 allocated two; the ACP port is gone —
   the SDK's transport is internal and needs no externally allocated port.)*
2. **Subprocess spawn.** autoharness spawns the intercom-go binary with the
   allocated port, the target workspace directory, and the config path.
3. **SDK client start.** The process constructs `copilot.NewClient(...)` and
   calls `Start(ctx)`. Note `NewClient` returns no error — **validation
   failures surface at `Start`**, so `Start` is the real readiness gate.
4. **Tunnel registration.** autoharness adds the WS port to the persistent Dev
   Tunnel.
5. **Client discovery.** The SPA connects to the tunnel-provided URL.

### 3.1 Shutdown ordering (normative)

Shutdown must run in this order. The naive ordering (stop clients → abort →
disconnect → stop) **deadlocks**: a blocked `PermissionHandlerFunc` is waiting
on a hub channel, so stopping the delivery path first means nothing can ever
answer it, the session never becomes idle, `Disconnect`'s precondition is
unsatisfiable, `Stop()` blocks closing sessions, blows its deadline, and
`ForceStop()` strands the CLI child — the exact outcome this section exists to
prevent.

**Shutdown executor (normative, and load-bearing).** The sequence below runs on
a **dedicated shutdown goroutine — never on the hub goroutine.** Steps 0, 2 and
3 block on results that can only arrive *through* the hub (a permission
resolution, a callback quiesce, an idle signal fed by ingress). Running them on
the hub would stop the select loop draining ingress, which wedges the
blocking-send callback contract (§4.2) and makes the idle signal unproducible —
a hard deadlock ending in `ForceStop` and a stranded CLI child, the exact
outcome this section exists to prevent. **The hub MUST keep draining ingress
through steps 0–5 and stops only at step 6.**

Normative sequence:

0. **Fail all pending permissions first.** Resolve every entry in the pending
   registry to `UserNotAvailable` *through the hub* and broadcast
   `permission.resolved` with a shutdown reason. Until this completes, the
   permission handler goroutines cannot return and the session cannot go idle.
1. **Broadcast `server.shutdown`** so operators learn why the session ended.
   This happens *before* any teardown, not after.
2. **Unsubscribe and quiesce SDK callbacks.** Call the `func()` returned by
   `Session.On`, then wait (bounded) for in-flight callbacks to drain via a
   `WaitGroup`/refcount. **Never close the ingress channel** — use a separate
   `done` channel. Closing ingress while a callback is in flight panics on an
   SDK-owned goroutine.
3. **`Session.Abort(ctx)`** any in-flight turn, then wait for a *positive*
   idle signal after ingress quiesces, bounded by a deadline. Do not read a
   stale fold-derived idle flag while events are still queued.
4. **`Client.Stop()`** — the SDK's documented graceful path (close sessions →
   request runtime shutdown → close JSON-RPC → terminate the CLI process it
   spawned). If it exceeds its deadline, **`Client.ForceStop()`**.
   *Note:* `Stop()` already closes sessions, so an explicit
   `Session.Disconnect()` is optional and is omitted here — it adds a race
   (step 3's idleness precondition) and buys nothing.
5. **Bounded drain of client write pumps (1-2 s), then abort sockets.** Each client's `done` MUST be closed (via its `sync.Once`) before the abort, per §5.5's drop protocol. An
   unbounded drain can block for minutes on a client with a closed TCP window
   — precisely what §5.5 forbids.
6. **Stop the hub last.** The hub must outlive the SDK, because SDK teardown
   produces events and unblocks handlers that require a live hub.

Every step is `context`-bounded. `Stop()` aggregates errors; they are logged,
never swallowed.

**Open spike question (C2):** does `Session.Abort` unblock an already-blocked
`PermissionHandlerFunc`? If it does, step 0 can be simplified. Until answered,
assume it does not.

---

## 4. Concurrency Model

Within one workspace process:

1. **Main goroutine — Bubble Tea.** `tea.NewProgram().Run()`. Consumes
   `tea.Msg` only; never touches shared state directly.
   **The TUI is fed through its own buffered channel and pump goroutine, not
   by a synchronous `program.Send` from the hub.** A stalled renderer must not
   block the hub select loop — that would stop ingress draining, backpressure
   SDK event delivery, and strand the permission path. TUI overflow policy is **drop-oldest only — never coalesce**. Coalescing would require re-parsing and merging payloads on the pump goroutine, i.e. off-hub interpretation of envelopes that §5.2 freezes as immutable/pre-marshalled. On overflow the TUI forces re-hydration.
2. **Agent adapter.** Registers `Session.On(handler)` — the SDK delivers
   events by **callback**, not by channel or iterator. The handler normalises
   the event and hands it to the hub. See §4.2 for the normative blocking rule.
3. **HTTP/WS listener.** On upgrade, spawns two goroutines per client:
   `wsReadPump` (operator actions inbound) and `wsWritePump` (envelopes
   outbound, one buffered channel per client).
4. **Hub.** A single `select` loop owning the canonical state. All mutation
   happens here, so the state needs no mutex for writes — **the hub goroutine
   is the sole writer**.

### 4.1 Sole-writer invariants (normative)

* **The hub NEVER performs a blocking send.** Every hub-side send is either
  to a buffered channel with a `select`/`default` fallback, or to a capacity-1
  reply channel. A blocking hub send halts envelope production for every
  surface and is the root cause of the wedge described in §5.4.
* **Local TUI actions mutate nothing directly.** A Bubble Tea `tea.Cmd`
  goroutine sends onto the hub ingress channel; `Update` never calls
  `session.Send` or resolves a permission itself. Otherwise the TUI becomes a
  second writer and the no-mutex model is void.
* **No sequenced envelope is constructed off-hub.** See §5.2.
* **"Immutable snapshot" means pre-marshalled bytes or a deep-copied,
  pointer-free value** — a shallow copy of `SessionState` leaks slice backing
  arrays and maps, and races the hub's next append. Marshal inside the hub.

### 4.2 SDK callback contract (normative)

The SDK callback "must not block" and "must push onto hub ingress" are in
tension when ingress is full. Resolve it explicitly rather than leaving it to
the implementer:

* Ingress is **buffered**, and the callback performs a **blocking send** —
  this is the single contract; there is no "bounded wait" variant.
* This is safe **only** because of the companion invariant "the hub never
  blocks" (§4.1) — the hub therefore always drains. Note the invariant is
  about blocking *sends*; it does **not** by itself bound hub *latency*. See
  §5.3's cost budget, which is what keeps a hub stall from becoming
  unbounded SDK backpressure.
* **Silent drop is forbidden.** Dropping an event corrupts the §5.1 fold and
  breaks `seq` continuity.
* **Degraded-path signalling is out-of-band.** If the adapter must report a
  fault on this path, it MUST NOT construct an envelope itself (§5.2 forbids
  off-hub sequenced envelopes) and MUST NOT route through ingress (which is by
  definition the unavailable resource). It signals on a dedicated capacity-1
  `degraded` channel that the hub selects on; the **hub** then produces the
  `error` envelope and forces full re-hydration.

**Callback quiescing requires a latch, not a bare `WaitGroup`.** Because
re-entrancy is unspecified (below), `wg.Add(1)` inside a callback races
`wg.Wait()` in the shutdown path — Go requires a positive-delta `Add` that
starts from zero to happen-before `Wait`. Use a one-way "no new entrants"
latch (an atomic closed-flag checked and incremented under a mutex, or an
equivalent refcount with a close latch), so `Unsubscribe` composes safely with
an already-entered callback.

**Re-entrancy is DETERMINED (2026-09-04, shipment 005-S / B7).** `Session.On`
callback invocation is **serialised** at the pinned SDK version
(`github.com/github/copilot-sdk/go@v1.0.11`): the phase-C2 proving spike
(`internal/copilotprobe`) instrumented the callback with an atomic in-flight
counter across a token-streaming turn and observed a maximum concurrent
in-flight count of **1** across 62 sampled invocations — no overlapping
invocation occurred. See
`docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md` (SQ-b).

> **Normative adapter rule.** The adapter MAY rely on `Session.On` delivering
> events to a single callback instance one at a time (no concurrent
> invocation to guard against at this pin). This is an **empirical
> observation of the pinned version's current behaviour, not a documented SDK
> guarantee** — the SDK's own API surface still carries no serialisation
> contract in its type signatures or documentation. The adapter MUST NOT
> silently assume this holds across a future SDK version bump; re-verify
> empirically (re-run the B5 probe, or an equivalent) before or alongside any
> `copilot-sdk/go` version upgrade, and revert to the conservative
> concurrent-safe posture below if re-verification is not performed.
> Regardless of serialisation, the adapter still MUST NOT rely on callback
> ordering *across* `Unsubscribe`/re-subscribe boundaries, and the
> "no new entrants" latch above remains required — serialisation within one
> subscription does not eliminate the shutdown race the latch guards against.

**C2 closure status: CLOSED for SQ-b.** This amendment satisfies design
§4.2's closure condition ("C2 may not close until the design has been amended
with the chosen fallback").

### 4.3 Event-union safety (normative)

`SessionEvent.Data` is a **sealed, code-generated discriminated union**. New
variants can appear in a patch release **without any compile error**. Every
type switch over `event.Data` MUST have a `default:` branch that logs the
unhandled `event.Type()` at warn level. Silently dropping unknown events is
prohibited.

---

## 5. Shared-State Architecture (Bubble Tea + SPA)

### 5.1 Canonical state and the durability model

The hub owns a `SessionState` projection built by **folding SDK events**. It
is a **cache for fast hydration and arbitration — not the system of record.**
The SDK runtime is the durable event source: `Session.GetEvents(ctx)` returns
the session's event history, and `Client.ResumeSession(ctx, id, ...)`
reattaches to a persisted session.

**Bootstrap (D4a).** `ResumeSession` needs a `sessionID`, and `GetEvents` is a
method on a `*Session` that exists only after a successful resume. The session
ID therefore cannot live only in the in-memory projection or it dies with the
process. intercom-go persists a **minimal durable session pointer** per
workspace — `session_id`, workspace path, protocol version, and nothing else.
This is a pointer to someone else's durable store, not a durable store; D7
(full persistence) remains deferred.

**Resume must not duplicate or gap at the SDK boundary.** Registering the live
callback and calling `GetEvents` is the same race §5.3 solves for clients.
Duplicates here are **not** benign: `seq` is assigned *after* normalisation, so
the same SDK event folded twice receives two different `seq` values and the
`seq`-keyed idempotence rule cannot detect it. Normative order:
**register the callback first, then call `GetEvents`, then de-duplicate while
folding.**

**De-duplication key: DETERMINED (2026-09-04, shipment 005-S / B7).**
`SessionEvent` carries a stable identity: `SessionEvent.ID` (`rpc.SessionEvent.ID`)
is a non-empty, unique-per-event UUID v4, generated when the event is
emitted — the phase-C2 proving spike observed 76/76 unique IDs with zero
duplicates across one probe session, and `ParentID` additionally provides a
linked-chain ordering signal. See
`docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md` (SQ-d). This
**refutes** this section's prior prediction that no stable identity exists;
the two-branch specification below is collapsed to the applicable branch:

* **De-duplicate on `SessionEvent.ID`.** The adapter normalises it to an
  opaque, intercom-go-owned `source_event_id` **inside the ACL** — the hub
  never sees or interprets an SDK field, preserving §7.1 and keeping the
  envelope free of SDK wire types.

C2 spike question (d) is answered; the blocking gate on C4 (alongside H6) is
lifted for this question.

### 5.2 Event envelope

One envelope type flows to both surfaces. It is **intercom-go-owned and newly
authored** — it deliberately does *not* mirror SDK wire types, so an SDK
change does not become a client-visible protocol change.

```jsonc
{
  "v": 1,                        // envelope schema version
  "gen": "01JC…",                // process generation; changes every start
  "seq": 1421,                   // monotonic within (gen); the resume cursor
  "ts": "2026-09-04T19:12:03Z",
  "kind": "agent.message.delta", // intercom-go vocabulary, NOT the SDK's
  "session_id": "...",
  "payload": { }                 // kind-specific
}
```

**The resume cursor is the pair `(gen, seq)`, never `seq` alone.** A restarted
process re-folds `GetEvents` from zero, so `seq` numbering restarts. A client
presenting a pre-restart `seq` against a new hub would be replayed from a
*different* numbering — silent transcript corruption with no detection path. A
cursor whose `gen` does not match the hub's forces **full re-hydration**, never
a replay.

**`gen` MUST never repeat across process lifetimes** — "changes every start" is
too weak, because an in-memory counter reset to 1 each start, or a
second-granularity timestamp under fast crash-restart, satisfies the letter and
reproduces the corruption. `gen` MUST be drawn from a random or
monotonic-time-plus-entropy source (ULID / UUIDv7), MUST NOT be a resettable
counter, and MUST be captured once at hub construction.

**Normative production rules:**

* `seq` assignment **and** enqueue-to-all-clients are a **single hub step**.
  No sequenced envelope may be constructed off-hub — including
  `server.shutdown`, WS-layer `error` envelopes, and timeout-path
  `permission.resolved`. *Design rule: if `seq` needs `atomic`, the invariant
  is already broken* — atomic assignment guarantees monotonic numbers but not
  matching broadcast order, so a client can receive 102 before 101 and
  silently drop 101.
* **Envelopes are immutable after publication.** One envelope is shared by the
  TUI pump and N write pumps; any consumer stamping a per-client field creates
  an N-way data race. Freeze at `seq` assignment; broadcast pre-marshalled
  bytes or an immutable value.

**Envelope kinds (initial set):** `session.hydrate`, `agent.message`,
`agent.message.delta`, `agent.turn.start`, `agent.turn.end`, `agent.idle`,
`permission.requested`, `permission.resolved`, `tool.started`, `tool.ended`,
`error`, `server.shutdown`.

### 5.3 Hydration and reconnect

* **Connect.** Snapshot construction, `(gen, seq)` high-water capture, and
  registration into the broadcast set are **ONE indivisible hub step**.
  Splitting them picks a failure mode: hydrate-then-register loses every
  envelope produced in between *permanently and undetectably* (the client's
  cursor was already set to the high-water mark), whereas register-then-hydrate
  yields duplicates, which the idempotence rule below already handles.
  Duplicates are recoverable; gaps are not.
* **Snapshot cost (stated, not asserted).** Marshal **inside** the hub against an explicitly **bounded** projection. The projection is a fold over an unbounded event history, so it MUST be capped (by entry count and serialized bytes) with the cap declared in the C4 exit contract, and the hub MUST cap **concurrent hydrations per select iteration** — otherwise N clients reconnecting after one tunnel blip marshal serially inside the select loop. Preferred implementation: maintain the snapshot as incrementally pre-marshalled bytes so hydration is a copy, not a marshal. Building it outside the hub
  means a reference escaped and races the next fold write; building an
  unbounded snapshot inside the hub stalls the select loop and starves both
  ingress and the permission reply path.
* **Reconnect.** The client sends its last-seen `(gen, seq)`. If `gen` differs
  → full hydrate. If `gen` matches and the ring still covers `seq` → replay
  `seq+1…HEAD`. Otherwise → full hydrate. **Coverage check, slice copy, and
  registration are one hub step**, or the ring can wrap mid-replay, or live
  envelopes can interleave with replayed ones.
* **Idempotence and gap detection.** Clients key on `(gen, seq)` and discard
  anything already applied — **and must additionally detect gaps**. "Discard
  what you've already applied" alone silently converts out-of-order delivery
  into permanent loss (receive 105 then 99 → 99 discarded forever). A detected
  gap forces re-hydration.
* Never block to serve a replay; drop to full hydrate instead.

### 5.4 First-responder action arbitration

Both surfaces may answer a pending permission request. The backend is
authoritative. The SDK's `PermissionHandlerFunc` is **synchronous** — it blocks
an SDK goroutine until a decision is returned — which makes the ordering rules
below load-bearing rather than stylistic.

1. **Register the pending request in the hub registry BEFORE broadcasting**,
   both inside a single hub select-loop iteration. This ordering is mandatory.
   Broadcast-first admits this interleaving: hub broadcasts at T0 → a fast
   operator answers at T1 → the hub finds no unresolved entry and *rejects* the
   answer → registration lands at T2 → nothing ever answers → the handler times
   out to `UserNotAvailable`. **The agent is denied a permission the operator
   granted.**
2. The `permission.requested` envelope is then broadcast to all surfaces.
3. The first inbound action matching an **unresolved** entry wins, by
   compare-and-swap inside the hub goroutine. Later actions are rejected with
   an explicit reason — never silently ignored.
4. **The hub owns the timeout, not the handler.** The handler must not run its
   own deadline and return independently. Two clocks in two goroutines produce
   a silent authorization divergence: the handler's deadline fires at T and
   returns `UserNotAvailable` to the SDK, an operator action arrives at T+ε,
   the hub — unaware the handler gave up — CASes a winner and broadcasts
   `permission.resolved{allow}`. Both surfaces render "allowed" while the agent
   was told "user not available". The mirror case is worse. Therefore: the
   timeout is a hub message, the hub is the **sole writer** of the decision,
   and the handler returns **only** the value the hub sends it.
5. **The reply channel has capacity 1**, and the hub sends with
   `select`/`default`. An unbuffered reply channel whose handler has already
   gone away blocks the hub select loop **forever**, halting all envelope
   production, starving every write pump, and timing out every subsequent
   permission — one slow operator wedges the whole process.
   **The handler always has an exit path:** it selects on the reply channel
   **and** a hub-`done` channel, defaulting to `UserNotAvailable` if the hub
   goes away. "The handler returns only what the hub sends" must not mean "the
   handler blocks forever if the hub never sends".
6. The winner's decision is converted by the adapter (§7.1) into an
   `rpc.PermissionDecision`.
7. **Finality is the hand-off, not the CAS.** The hub broadcasts
   `permission.resolved` (carrying the **winning actor**) only after the
   handler has acknowledged receipt. If the capacity-1 send takes the
   `default` branch, or the handler is known gone, the hub MUST instead emit
   `permission.aborted` — otherwise both surfaces render "allowed" while the
   SDK received nothing, which is the original dual-owner defect in miniature.
   A third party can also abandon the request (SDK `Abort`, `Stop`/`ForceStop`,
   or a CLI-side RPC timeout); those paths resolve through the same
   `permission.aborted` envelope.

**Pending permissions are arbitration state, not fold state.** `session.hydrate`
MUST carry unresolved pending permissions with their remaining deadlines —
otherwise a reconnecting SPA never re-renders the modal and the request dies by
timeout with an operator watching it. **Membership is explicit: the local TUI counts as a permanent member of the broadcast set** (it is never dropped for backpressure — §5.5 — only overflowed and re-hydrated). So "last surface gone" means the last *remote* surface disconnected **and** the TUI has exited. When the broadcast set empties by that definition, resolve immediately to `UserNotAvailable` rather than
burning the full deadline waiting for an answer that is definitionally
unobtainable.

**Head-of-line blocking is assumed until disproven.** If the SDK dispatches
`OnPermissionRequest` on the same goroutine or queue as session events, a
blocked handler freezes *all* event delivery for the approval window. Whether
permission dispatch is independent of event dispatch is a **required C2 spike
question**. Until answered, do not design a UI that expects streaming to
continue behind a modal.

**Authorization, not just authentication.** Any *connected* surface can approve
an agent shell command. Connection-time authentication alone is therefore not
sufficient; an explicit authorization model (who may approve what) MUST be
decided before phase C5. See §6.

### 5.5 WebSocket ownership and backpressure

* **Ownership.** Exactly one goroutine writes to a given socket
  (`wsWritePump`). The hub never writes to a socket directly; it writes to the
  client's buffered channel. **Keep-alive pings are writes**, so the ping
  ticker MUST fire as a `case` inside the write pump's own `select` — a
  separate ping goroutine or library keep-alive is a second writer and
  corrupts frames. `Write`, `Ping`, and `Close(status, reason)` are forbidden
  from the hub, the read pump, and the shutdown path; only a socket **abort**
  (`CloseNow`-style) is permitted cross-goroutine, and only after `done`.
* **Backpressure.** If a client's buffer is full, that client is **dropped**,
  not blocked. One slow remote operator must never stall the agent event
  stream, the local TUI, or other clients.
* **Drop protocol (normative).** Each client has a `done` channel closed
  exactly once via `sync.Once`. The hub drops a client by closing `done` and
  **never** closes the client's send channel — otherwise a hub drop racing a
  read-pump failure produces a double-close or send-on-closed-channel panic.
  The **write pump alone** writes the close frame and closes the socket.
  Unregistration is an idempotent hub step.
* A dropped client reconnects and rehydrates via §5.3 — which is why the
  reconnect path must be correct.
* **Local TUI is not a WebSocket client.** It receives envelopes through its
  own buffered channel and pump (§4). It is subject to the same arbitration
  rules but not to socket backpressure.

---

## 6. Dev Tunnel Boundary, Auth, and Authorization

* Dev Tunnels provide an **outbound** TLS relay; no inbound firewall change.
* **Assumption (not operator-confirmed, tracked as Q3 in the governing
  decision):** the tunnel is created with anonymous access disabled
  (`--allow-anonymous false`) and gates on GitHub identity at the Azure edge.
* **Normative regardless of the above:** the backend treats the tunnel as an
  **untrusted** boundary and performs its own authentication on every
  WebSocket upgrade. Edge gating is defence in depth, never the only control.
  A misconfigured tunnel must not equal an open agent.
* The listener binds **loopback** only. Exposure is exclusively via the tunnel.
* Copilot credentials never traverse the tunnel. The SPA authenticates to
  *intercom-go*; intercom-go authenticates to *Copilot* independently
  (`ClientOptions.GitHubToken`, or the logged-in CLI user).

### 6.1 Controls that MUST be specified before C6/C10 (not yet decided)

Authentication alone is insufficient here, because §5.4 lets **any connected
surface approve an agent shell command**. The following are recorded as
required decisions, not as implemented controls:

* **Origin / CSWSH.** A WebSocket upgrade is not subject to the same-origin
  policy and browsers attach cookies cross-origin. `Origin` must be validated
  on upgrade, and the SPA must authenticate with a bearer credential rather
  than an ambient cookie.
* **Authorization model.** Who may approve *what*. At minimum, distinguish an
  observer from an approver, and decide whether a remote surface may approve
  destructive shell commands at all, or only the local TUI may. **Blocks C5.**
* **Per-message authorization, not connection-time only.** A long-lived socket
  authorized once at upgrade grants approval rights for the socket's whole
  lifetime. Re-validate on privileged actions.
* **Token replay and expiry.** Bind the credential to the connection and bound
  its lifetime.

### 6.2 Copilot CLI resolution is a security-relevant path

`[copilot].cli_path` selects a binary that is **spawned as a child process and
given agent capabilities**. Empty means "resolve from `PATH` /
`COPILOT_CLI_PATH`", which is the SDK's own documented behaviour — but that
makes CLI resolution a search-order-dependent trust decision.

**Classification predicate (normative).** "Bare name" and "relative path" are
otherwise indistinguishable, and `filepath.IsAbs` alone cannot separate them,
so the discriminator is stated explicitly:

```text
cli_path == ""                              -> valid  (SDK resolves via PATH / COPILOT_CLI_PATH)
filepath.IsAbs(cli_path)                    -> valid iff the file exists, else ERROR
matches ^[A-Za-z]: and not absolute         -> ERROR  (Windows drive-relative, e.g. "C:copilot")
contains '/' or '\' (and not absolute)      -> ERROR  (relative path)
otherwise (no separator, no drive prefix)   -> bare name: PATH advisory only, never fatal
```

Both separators are tested regardless of `GOOS`, following the existing
`containsDotDotSegment` precedent in `internal/config`, because `config.toml`
is portable text that may be authored on a different platform than it runs on.

**The drive-prefix branch is not redundant, and omitting it is a real bypass.**
`C:copilot` is non-absolute *and* contains no separator, so a naive predicate
would classify it as a bare name and pass it through with only an advisory.
But Go's `exec.LookPath` on Windows treats any string containing `:` as a
**path** rather than a `PATH` search, and `StdioConnection.Path` would then
spawn it drive-relative to the process working directory — exactly the
unchecked, CWD-relative execution the relative-path rejection exists to
prevent. It is therefore rejected explicitly, ahead of the separator test.
UNC paths are already absolute and are handled by the `IsAbs` branch.

Rationale per branch:

* **Absolute** — validated for existence, so it fails closed.
* **Relative** (`./bin/copilot`, `..\x\copilot`) — **rejected**. It is neither
  existence-checked nor containment-checked and resolves against the process
  working directory, which is attacker-influenceable in a way an operator is
  unlikely to reason about.
* **Bare name** — resolves through ambient `PATH`. This is the SDK's default
  behaviour and is **not made worse** by intercom-go, but it is a real
  PATH-hijacking surface on both Windows and POSIX. It is recorded as an
  **accepted residual risk**, surfaced as a non-fatal advisory, not treated as
  validated.

This whole contract occupies **one** validation rule slot (rule 5), so the
compensating control does not become an unenumerated eighth check.

### 6.3 Config-sourced prompt content

`[commands]` values are operator-authored strings that become **prompts sent
to an agent**. Config is therefore a prompt-injection surface. Treat
`config.toml` as a trusted operator artifact with the same handling as code,
and never populate `[commands]` from a remote or untrusted source.

---

## 7. SDK Integration Rules (normative)

### 7.1 Anti-corruption layer

All SDK types are confined to one adapter package.

* `rpc.PermissionDecision` is documented **"Experimental: ... may change or be
  removed"**, yet it is the mandatory return type of the permission handler.
  It is therefore converted to and from an intercom-go-owned domain type in
  **exactly one file**.
* No package outside the adapter imports
  `github.com/github/copilot-sdk/go/rpc`.
* Event normalisation into §5.2 envelopes happens only in the adapter.

### 7.2 Version pinning

* Pin the module to `v1.0.11`.
* **Pinning the module does not pin the Copilot CLI.** The SDK declares
  `SDKProtocolVersion = 3` and, for Go, the CLI is an unbundled,
  operator-installed dependency. The validated CLI version must be recorded
  and asserted at startup. Embedding the CLI (`go/embeddedcli`) is a deferred
  option that would remove this variable entirely.

### 7.3 Multi-tenant hygiene

Because each process is workspace-scoped, evaluate
`ClientOptions.Mode = ModeEmpty` (no built-in tools, no
`environment_context`, telemetry off) with an explicit
`SessionConfig.AvailableTools`. `ModeEmpty` requires `BaseDirectory`,
`SessionFS`, or a `URIConnection` for persistent session state.

### 7.4 Surfaces to avoid until de-flagged

`InProcessConnection`, `Providers`/`Models` BYOK, `EnableCitations`,
`SessionLimits`, `EnableMCPApps`, `ExpAssignments`.

### 7.5 Known documentation drift

`Session.SendAndWait`'s doc comment describes a `timeout` parameter that
**does not exist**. Always pass a deadline-bearing `context`. Trust
signatures over prose.

---

## 8. Configuration Impact

The shipped `internal/config` package encodes the retired architecture and is
remediated by the corrective shipment. Full per-surface classification is in
the governing decision's *Shipped P2 Audit*. Summary:

* **Removed:** `[slack]`, **the entire `[acp]` section (all four fields)**,
  `host_cli_args`, `ipc_name`, `[[workspace]].channel_id`; validation rules 4
  and 8; the channel-routing API in `internal/config/workspace.go`
  (`ResolveChannelID`, `ResolveWorkspaceByChannelID`, and
  `WorkspaceRootForChannel` re-founded on `workspace_id`). `acp.max_sessions`
  is **deleted** (its semantics are already covered by
  `max_concurrent_sessions`, Q6); `acp.startup_timeout_seconds` is **deleted in
    C1** and re-introduced **in C3** with its first real consumer (the SDK
    `Client.Start` deadline).
    > **Amended 2026-09-04 (005-S / A2).** Previously scheduled for
    > re-introduction in C2. C2 (the Copilot SDK proving spike, shipment 005-S)
    > produces **no production consumer** — `internal/copilotprobe` is
    > explicitly disposable per its Non-Goals — so re-introducing this field in
    > C2 would ship a config surface with no reader, reproducing exactly the
    > kind of coupling C1 removed. Deferred one phase to C3, where the first
    > real `Client.Start` call exists to read it. See
    > `docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md` (Decisions and
    > Rationale) and
    > `docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`.
* **Renamed:** `slack_detail_level` → `operator_detail_level`.
* **Migrated:** `host_cli` → `[copilot].cli_path`. Empty means resolve from
  `PATH`; a bare name gets a non-fatal advisory; an absolute path must exist;
  a **relative path is rejected** (§6.2).
* **Retained:** workspace root, workspace mappings (minus `channel_id`),
  timeouts, stall, commands, `http_port`, tolerant-decode reporting.
* **Deferred:** `database`, `retention_days` — pending the persistence
  decision (D7), which this revision deliberately does not make. Note D4a's
  session pointer is **not** this store.

**Post-change validation contract: exactly 7 rules** — (1)
`max_concurrent_sessions` non-zero; (2) `default_workspace_root` set and
canonicalizes; (3) `workspace_id` non-empty; (4) `workspace_id` unique; (5)
**`cli_path` well-formed** (empty | bare name | existing absolute; relative
rejected — one rule, three branches, per §6.2); (6) `[[workspace]].path`
canonicalizes; (7) `database.path` has no `..` segment. Rules 4 and 8 are
removed and rule 6's presence check is **eliminated** (not relaxed), so it
drops out of the enumeration entirely: 10 − 2 − 1 = 7.

**Security note carried forward:** the shipped example contains
`host_cli_args = ["--dangerously-skip-permissions"]`. Under this architecture
permission handling is a core product responsibility (§5.4). That example is
removed as part of the remediation.

---

## 9. Non-Goals

* SPA framework selection or SPA implementation.
* Any ACP protocol implementation.
* Slack support of any kind.
* A persistence schema (deferred).
* Parity with the Rust `agent-intercom` implementation. That repository is
  **historical reference only**, not a behavioural oracle for operator UI,
  transport, configuration, or credentials.
* **RC-2 — Sidecar process supervision over Named Pipes / IPC.** Rejected for
  now (deferred as **Q7**, see §9.1 below): inverts the decided supervision
  topology, spans three repositories with no authorization, and is
  Windows-only against a cross-platform posture.
* **RC-3 — React / iOS / VAPID web-push re-introduction.** Rejected: directly
  contradicts the 2026-09-04 correction's UI and transport decisions.
* **RC-5 — Clean-room ephemeral `_stage`/`_ship` delegation engine.** Deferred
  to ≥ C9 (tracked as Q6): presumes an unresolved sessions-per-process
  decision.
* **RC-6 — Fail-closed circuit breaker performing automated `git reset` /
  stash writes against the operator's checkout.** Deferred, with the risk
  flagged as **R8** (see §9.1): if revived, any automated convergence
  remediation must target a dedicated worktree or branch, never the
  operator's own checkout.

### 9.1 Deferred and adoptable material register

Recorded here (rather than left to survive only in the untracked
IMPL-DESIGN candidate document,
`docs/design-docs/intercom-architecture-implementation-design.md`) per
decision D14:

| ID | Item | Disposition | Source |
|---|---|---|---|
| **Q7** | How intercom reaches workspace tooling (`agent-engram`, `graphtor-docs`, `backlogit`) — transport undecided (Named Pipes / IPC was proposed via RC-2 and rejected for now) | Open question, future deliberation | RC-2; IMPL-DESIGN §2 |
| **R8** | An automated fail-closed circuit breaker performing `git reset` or stash writes could destroy operator work if revived | Risk, flagged; deferred with RC-6 | RC-6; IMPL-DESIGN §4.2 |
| Event bus (adoptable) | Central unidirectional fan-out event bus: strict payload typing, buffered channels, `context.Context` cancellation, timeout thresholds so a dropped mobile connection never stalls the SDK loop | **Adoptable** — carries forward into design rev 3 (C5–C11 planning); materially consistent with and sharpens §4 (Concurrency), §5.2 (event envelope), and §5.5 (WebSocket backpressure) | IMPL-DESIGN §7, Phase 3 "Telemetry & State Bus", step 8 |
| TUI layout (adoptable) | Bubble Tea Elm architecture (`tea.Model`/`Update`/`View`), `lipgloss` 2D layout, intervention modals | **Adoptable** — consistent with and adds useful layout detail beyond §2's stack (`bubbletea` + `lipgloss` + `bubbles/viewport`) | IMPL-DESIGN §6.1 |

### 9.2 Terminology (D12)

The acronym **"ACP"** is retired as a product term **in both expansions** —
*Agent Client Protocol* and *Agent Control Plane* — per decision D12
(`docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`).
Use "control plane" in prose where needed; the acronym itself must never
appear as a Go identifier, a TOML key, or any other code/config surface in a
governing artifact. The retired term survives only inside the preserved,
non-governing IMPL-DESIGN candidate document (its title and body predate this
decision and are left unmodified per D11).

