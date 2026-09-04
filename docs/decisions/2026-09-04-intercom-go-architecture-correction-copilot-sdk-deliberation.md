---
title: "Architecture Correction: GitHub Copilot SDK, Bubble Tea + SPA, Slack Retirement"
description: "Governing architecture correction for intercom-go — retires Slack and the custom ACP broker, adopts the official GitHub Copilot SDK for Go as the agent boundary, and fixes the dual operator surface (Bubble Tea TUI + SPA over Dev Tunnels)"
topic: "intercom-go corrective architecture after operator decision superseding Slack/custom-ACP assumptions"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/design-docs/intercom-go-backend-architecture.md"
  - "docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md"
supersedes:
  - "docs/decisions/2026-07-06-go-port-reference-brief.md"
  - "docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md"
  - "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
tags:
  - "architecture"
  - "governing-decision"
  - "copilot-sdk"
  - "slack-retirement"
  - "correction"
---

# Architecture Correction (GOVERNING)

> **Status: GOVERNING.** This artifact is the authoritative architecture record
> for intercom-go as of 2026-09-04. Where it conflicts with any earlier
> decision, plan, design document, or stash entry, **this artifact wins.**

## Problem Frame

### What happened

intercom-go was staged and shipped through three phases (`001-S` foundation,
`002-S` apperr/pathsafe, `003-S` P2 configuration) under an architecture
inherited from a Rust predecessor (`softwaresalt/agent-intercom`). That
inherited architecture assumed:

* Slack as the remote operator surface (Socket Mode, slash commands, modals,
  channel/team/member IDs, Slack-shaped persistence);
* an application-owned **ACP broker** speaking the Agent Client Protocol over
  TCP to a **headless Copilot CLI**, decoded with `coder/acp-go-sdk`;
* an iOS client as the remote UI;
* a "feature-complete Rust port" as the definition of done, with the Rust
  repository as a **behavioral oracle** governing operator UI, transport,
  config schema, and credential precedence.

A P3 credentials plan built on those assumptions reached PR #13, which was
**closed unmerged as obsolete**. Stage halted.

### The operator correction (binding)

The operator has issued an explicit architecture correction that resolves that
halt and supersedes the Slack/custom-ACP assumptions. The corrected
architecture is stated in *Decision* below and is binding on all downstream
planning.

### Why this is load-bearing now

Three shipments of code and documentation already exist. The longer the
corrected architecture stays unrecorded, the more work is built against a
retired design. The P2 configuration package in particular has **already
shipped a Slack-shaped, ACP-shaped, headless-host-shaped schema to `main`**
(see *Shipped P2 Audit*). Every future phase that reads `internal/config`
inherits that contamination. Correcting the record is therefore the
dependency root of all remaining work.

### Success criteria

1. A single unambiguous governing architecture document exists, and the stale
   design document no longer contradicts it.
2. The application-owned architecture is explicitly separated from the Copilot
   SDK's internal CLI transport, so no future phase re-implements ACP.
3. Every shipped surface carrying a retired assumption is classified with an
   explicit remediation disposition (retain / remove / rename / migrate).
4. The backlog and stash cannot be used to accidentally implement retired
   scope.
5. The SDK integration is grounded in verified public API, or an explicit
   spike is defined where uncertainty remains.

### Explicitly out of scope

* SPA framework selection (React/Vue/Svelte/etc.) — not specified by the
  operator, deliberately deferred.
* SPA implementation of any kind.
* Product code changes (Stage does not write product code).
* Rewriting archived historical evidence beyond adding superseded markers and
  cross-links.
* Persistence engine selection beyond what *Decision D7* fixes.

---

## Research Findings

### F1 — The official GitHub Copilot SDK for Go exists, is public, and is GA

Verified by direct source inspection at a pinned tag (not from memory or search
snippets):

| Property | Verified value |
|---|---|
| Repository | `github/copilot-sdk` — public, MIT, default branch `main` |
| Go module path | `github.com/github/copilot-sdk/go` |
| Go language floor | **`go 1.24`** (`go/go.mod`, and `go/README.md`: "Go 1.24 or later") |
| Tag scheme | `go/vX.Y.Z` (78 Go tags) |
| Latest **stable** | **`go/v1.0.11`** = commit **`a550258d5c37bd662197536992a23d633bfe5804`** (2026-08-14) |
| Stability statement | Root `README.md`: "The GitHub Copilot SDK is generally available and follows semantic versioning." |
| Direct dependencies | `google/jsonschema-go`, `klauspost/compress`, `coder/websocket`, `ebitengine/purego`, `google/uuid`, `otel`, `otel/trace` |

`proxy.golang.org/.../@latest` and pkg.go.dev both resolve to `v1.0.11`.
Everything newer (`v1.0.12-preview.0`, `v1.0.13-preview.0..4`) is a prerelease.

### F2 — The SDK owns the CLI transport; the application does not

Root `README.md` states verbatim:

> All SDKs communicate with the Copilot CLI server via JSON-RPC:
> `Your Application → SDK Client → JSON-RPC → Copilot CLI (server mode)`.
> The SDK manages the CLI process lifecycle automatically.

Confirmed in source (`go/client.go`): the SDK imports
`github.com/github/copilot-sdk/go/internal/jsonrpc2`, builds `exec.Command`,
and passes `--stdio` or `--port N`. **The JSON-RPC codec lives under
`internal/` and is therefore not importable by host applications.**

This is the single most important finding for this correction: it confirms the
operator's framing exactly. The agent transport is an SDK implementation
detail. An application-owned ACP broker is not merely unnecessary — the
protocol layer is deliberately sealed away from hosts.

Nuance that must not be lost: *transport selection* is public even though the
*codec* is internal. `RuntimeConnection` is a **sealed interface** (unexported
methods, so hosts cannot implement custom transports) with four provided
implementations: `StdioConnection`, `TCPConnection`, `URIConnection`,
`InProcessConnection` (experimental, build-tag gated).

### F3 — The SDK's public API covers every boundary this product needs

Verified signatures at `go/v1.0.11`:

```go
// Client lifecycle
func NewClient(options *ClientOptions) *Client   // returns no error; validation happens at Start
func (c *Client) Start(ctx context.Context) error
func (c *Client) Stop() error
func (c *Client) ForceStop()

// Session lifecycle
func (c *Client) CreateSession(ctx context.Context, config *SessionConfig) (*Session, error)
func (c *Client) ResumeSession(ctx context.Context, sessionID string, config *ResumeSessionConfig) (*Session, error)
func (s *Session) Send(ctx context.Context, options MessageOptions) (string, error)
func (s *Session) SendAndWait(ctx context.Context, options MessageOptions) (*SessionEvent, error)
func (s *Session) Abort(ctx context.Context) error       // interrupt in-flight turn
func (s *Session) Disconnect() error
func (s *Session) GetEvents(ctx context.Context) ([]SessionEvent, error)   // <-- see D4

// Events: callback registration returning an unsubscribe func
type SessionEventHandler func(event SessionEvent)
func (s *Session) On(handler SessionEventHandler) func()

// Permissions
type PermissionHandlerFunc func(request PermissionRequest, invocation PermissionInvocation) (rpc.PermissionDecision, error)
// wired via SessionConfig.OnPermissionRequest

// Tools
func DefineTool[T any, U any](name, description string, handler func(T, ToolInvocation) (U, error)) Tool
```

`SessionEvent.Data` is a **sealed discriminated union** (`SessionEventData`
with unexported methods). Event types include `assistant.message`,
`assistant.message_delta`, `assistant.turn_start`, `assistant.turn_end`,
`session.idle`, `permission.*`, `tool_execution.*`, and many more.

### F4 — Three SDK facts that materially shape the design

1. **`ClientOptions.Mode = ModeEmpty`** exists for multi-tenant-safe defaults
   (no built-in tools, no `environment_context`, telemetry off). It requires
   `BaseDirectory`, `SessionFS`, or a `URIConnection` for persistent session
   state. Relevant because intercom-go runs one process per workspace and must
   not leak context across workspaces.
2. **`Session.GetEvents(ctx)` returns the session's full event history from the
   runtime.** This is decisive for hydration (see *D4*).
3. **The Go SDK does not bundle the Copilot CLI.** Root `README.md`: "For Go,
   Java, and Rust SDKs, the CLI is not bundled by default." The CLI must be on
   `PATH` or located via `COPILOT_CLI_PATH`. There is also an
   application-level bundling path (`go/embeddedcli`, `go/cmd/bundler`).

### F5 — Verified SDK risks

| # | Risk | Evidence |
|---|---|---|
| R1 | `rpc.PermissionDecision` is marked **"Experimental: … may change or be removed"** yet is the mandatory return type of the permission handler every host must implement | `go doc rpc.PermissionDecision` |
| R2 | High release velocity: 39 commits touched `go/` in the 3 weeks after `v1.0.11`; 6 prerelease tags in that window | `gh api repos/github/copilot-sdk/commits?path=go` |
| R3 | Large **generated** surface (`go/types.go` 145 KB, `go/zsession_events.go` 94 KB). New union variants appear **without compile errors** — a non-exhaustive type switch silently drops events | source inspection |
| R4 | `SDKProtocolVersion = 3` couples the SDK to a Copilot CLI version, and for Go the CLI is an **unbundled operator-installed dependency**. Pinning the module does **not** pin the CLI | `go/sdk_protocol_version.go` |
| R5 | Documentation drift: `Session.SendAndWait`'s doc comment describes a `timeout` parameter that does not exist in the signature; the package-level `go doc` example does not compile against v1.0.11 | source vs. doc comparison |
| R6 | Experimental sub-surfaces to avoid: `InProcessConnection`, `Providers`/`Models` BYOK, `EnableCitations`, `SessionLimits`, `EnableMCPApps`; `ExpAssignments` is internal-only | doc comments |

### F6 — The current repository cannot compile against the SDK

`go.mod` declares `go 1.22`. The SDK requires `go 1.24`. **This is a hard,
machine-verifiable blocker** and is the narrowest possible dependency root for
the corrective work.

### F7 — Prior-art retrieval

`docs/compound/` contains one learning:
`2026-05-07-backlogit-shipment-status-constraints.md` (shipment closure must go
through `ShipShipment`, not a generic status move). Relevant to Ship, not to
this architecture decision. **Confidence: low** for this topic — no prior
architecture learnings exist. No prior art constrains this decision.

---

## Options Evaluated

The operator's directive fixes the *destination* architecture. The genuinely
open question this deliberation must settle is **how to sequence the correction
safely**, plus four sub-decisions the directive does not resolve (hydration,
persistence, config remediation shape, and Rust-oracle disposition).

### Option A — Document-only correction, defer all code remediation

Write the governing decision and revise the design doc. Change nothing else.
Fix `internal/config` whenever a later phase happens to need it.

* **Pros:** smallest possible slice; zero risk to shipped code.
* **Cons:** leaves a contaminated, *validated* config schema on `main` that
  actively enforces Slack semantics (rule 4 requires `channel_id` on every
  workspace; rule 8 enforces channel uniqueness; rule 6 requires `host_cli`).
  Any P4 work loads that config and inherits the contamination. Also leaves
  `host_cli_args = ["--dangerously-skip-permissions"]` in the shipped example.
* **Effort:** low. **Fit:** poor — fails success criterion 3.

### Option B — Big-bang corrective shipment (docs + full config rewrite + SDK integration)

One shipment: governing docs, complete `internal/config` rewrite, `go.mod`
bump, SDK client wrapper, event envelope, WebSocket hub, Bubble Tea shell.

* **Pros:** one coherent landing; no intermediate inconsistent state.
* **Cons:** violates the 2-hour rule at task level and produces an oversized
  shipment the operator explicitly prohibited. Couples an *unverified* SDK
  integration to a *well-understood* config remediation, so an SDK surprise
  blocks the config fix. Nothing is independently rollback-able.
* **Effort:** high. **Fit:** poor — explicitly prohibited by the directive.

### Option C — Governance + config remediation + SDK proving spike (RECOMMENDED)

One bounded corrective shipment containing exactly three cohesive things:

1. **Governance:** governing decision artifact, revised design document,
   superseded markers and cross-links on stale artifacts.
2. **Config remediation:** rewrite the shipped `internal/config` schema,
   defaults, validation rules, example, and reference doc to the corrected
   architecture — removing Slack, ACP, and headless-host surfaces.
3. **SDK proving spike:** raise the `go.mod` floor to 1.24, add the pinned SDK
   dependency, and prove the three risky surfaces (permission round-trip,
   event union handling, abort/shutdown) in a throwaway harness — producing a
   findings artifact, **not** production wiring.

* **Pros:** the config remediation is fully understood *today* and is the
  thing actively causing wrong work; the spike de-risks R1/R3/R5 before any
  architecture is committed to code. Each of the three has an independent
  rollback point. Cohesive: all three are "stop building on the wrong
  foundation."
* **Cons:** the config rewrite is a breaking change to a package that just
  shipped; needs deliberate migration notes. Spike output is not shippable
  product value on its own.
* **Effort:** medium. **Fit:** strong.

### Option D — Governance-only slice, then a separate config slice, then a separate spike

Same content as C, split across three sequential shipments.

* **Pros:** maximally small slices; each independently reviewable.
* **Cons:** three full Stage→Ship cycles to reach the same state. Governance
  alone does not stop wrong work — the contaminated schema still sits on
  `main` through two more cycles. Ceremony cost exceeds the risk reduction,
  since all three parts share one rationale and one review context.
* **Effort:** medium-high (aggregate). **Fit:** moderate.

### Trade-off Comparison

| Criterion | A (docs only) | B (big bang) | **C (governance+config+spike)** | D (three slices) |
|---|---|---|---|---|
| Stops further wrong work | ✗ no | ✓ yes | **✓ yes** | partial (delayed) |
| Respects 2-hour task rule | ✓ | ✗ | **✓** | ✓ |
| Oversized shipment risk | none | **high** | **low** | none |
| Independent rollback points | 1 | 0 | **3** | 3 |
| De-risks SDK before commitment | ✗ | ✗ (couples) | **✓** | ✓ |
| Cycles to safe state | ∞ | 1 | **1** | 3 |
| Dark-factory readiness | poor | poor | **good** | good |

---

## Decision

### D0 — Governing architecture (binding, from the operator)

```text
┌─────────────────────────────────────────────────────────────────────────┐
│  APPLICATION BOUNDARY — owned by intercom-go                            │
│                                                                         │
│   Bubble Tea TUI  ──┐                                                   │
│   (local operator)  │                                                   │
│                     ├──► OperatorGateway ──► Canonical SessionState ──┐ │
│   SPA (remote)  ────┘    (event envelopes,     (backend is the         │ │
│     ▲                     hydrate/reconnect,    single source of truth)│ │
│     │                     first-responder                              │ │
│  WebSocket                arbitration)                                 │ │
│     ▲                                                                  │ │
│  HTTP/WS endpoint                                                      │ │
│     ▲                                                                  ▼ │
│  Microsoft Dev Tunnels                            copilot.Client / Session│
└───────────────────────────────────────────────────────────┬─────────────┘
                                                            │  ← SDK BOUNDARY
        ══════════════════════════════════════════════════  │
                                                            ▼
                             SDK-INTERNAL (NOT owned by intercom-go)
                             JSON-RPC over stdio/TCP  →  Copilot CLI (server mode)
```

1. **Agent boundary is the official SDK.** intercom-go integrates
   `github.com/github/copilot-sdk/go`. It does **not** implement an ACP broker,
   an ACP wire protocol, or a direct headless-Copilot-CLI lifecycle.
2. **The SDK's JSON-RPC-to-CLI transport is an SDK implementation detail.**
   That the SDK spawns and manages a Copilot CLI server over JSON-RPC (F2) is
   *inside* the SDK boundary. It is never an intercom-go-owned architecture,
   never re-implemented, and never depended on at the wire level. The
   codec is under `internal/` and is unimportable by design.
3. **Local operator UX:** Bubble Tea terminal application.
4. **Remote operator UX:** SPA frontend (framework choice out of scope).
5. **Remote transport:** backend HTTP/WebSocket endpoint exposed through
   Microsoft Dev Tunnels.
6. **Deployment:** one isolated Go process per workspace, orchestrated and
   launched by autoharness.
7. **Slack is completely retired.** No Socket Mode, no slash commands or
   modals, no Slack credentials, no team/channel/member IDs, no Slack-shaped
   persistence, no Slack-driven workflows.

### D1 — SDK version pin

Pin **`github.com/github/copilot-sdk/go v1.0.11`** (tag `go/v1.0.11`, commit
`a550258d5c37bd662197536992a23d633bfe5804`).

Rationale: it is the only stable Go release; `@latest` and pkg.go.dev agree
with it, so `go get` default behaviour and the pin do not diverge; HEAD is 39
`go/` commits past it with no release notes and would force a pseudo-version.

**Public API uncertainty does not block implementation** — the API is verified
and GA (F1, F3). But R1/R3/R5 are real, so *D2* requires a spike before the
SDK is wired into production paths.

### D2 — SDK integration is gated behind a proving spike

Before any production SDK wiring, a bounded spike must prove exactly three
surfaces against the pin, because these are precisely where documentation and
reality diverged during research:

* **S1 — permission round-trip:** a real `PermissionRequestShell` reaching a
  `PermissionHandlerFunc` and each decision variant being accepted (risk R1).
* **S2 — event union handling:** enumerating `SessionEvent.Data` variants and
  proving that an unknown variant is observable rather than silently dropped
  (risk R3).
* **S3 — cancellation/shutdown:** `Abort` mid-turn, then `Stop`, then
  `ForceStop` as an escape hatch, with a deadline-bearing context (risks R5,
  and the missing `SendAndWait` timeout parameter).

The spike produces a findings artifact. It does **not** produce production
wiring.

### D3 — Anti-corruption layer is mandatory

All SDK types are wrapped behind an intercom-go-owned interface at exactly one
boundary. Specifically:

* `rpc.PermissionDecision` (experimental, R1) is converted to an
  intercom-go-owned domain decision type in one file.
* Every `SessionEvent.Data` type switch carries a `default:` branch that logs
  the unhandled `event.Type()` (R3).
* Application code never imports `github.com/github/copilot-sdk/go/rpc`
  outside the adapter package.

This makes an SDK breaking change cost one file rather than the codebase.

### D4 — Hydration is SDK-backed, not an in-memory replica

The stale design document specified an in-memory `State Hydrator` holding the
full transcript. `Session.GetEvents(ctx)` (F4.2) returns the session's event
history from the runtime, and `ResumeSession` reattaches to a persisted
session.

Therefore: **the backend's canonical `SessionState` is a projection built by
folding SDK events, and the SDK runtime is the durable event source.** The
backend keeps a bounded in-memory projection for fast hydration and
first-responder arbitration, but it is a **cache, not the system of record.**

#### D4a — The session-pointer bootstrap (REQUIRED; closes a hole in D4)

D4 as first written was **incomplete**, and the gap is load-bearing:
`Client.ResumeSession(ctx, sessionID, cfg)` requires a `sessionID`, and
`Session.GetEvents(ctx)` is a method on a `*Session` that exists only *after* a
successful resume. If the session ID lives **only** in the in-memory
projection, it dies with the process — so on crash there is no session ID, no
resume, no history, and D4 would merely relocate the durability conflict rather
than resolve it.

**Decision:** intercom-go persists a **minimal durable session pointer** per
workspace — the current `session_id`, its workspace path, and the SDK/CLI
protocol version it was created under. Nothing else. It is written on session
creation and cleared on clean session end.

This is deliberately **not** a reopening of D7. The distinction is exact:

| | Session pointer (D4a — decided now) | Session/transcript persistence (D7 — still deferred) |
|---|---|---|
| Content | one ID + workspace path + protocol version | messages, tool calls, approvals, audit, retention |
| Size | a few hundred bytes | unbounded, growing |
| Owner | intercom-go | undecided |
| Purpose | locate the SDK-owned session after restart | *be* the record |

Persisting a pointer to someone else's durable store is not the same as
owning a durable store. D7 stays deferred; `database` and `retention_days`
stay untouched and inert.

**Cardinality:** the pointer store is a **set keyed by `session_id`**, not a
single slot. Q6 (whether one process hosts more than one concurrent session)
is not scheduled to be answered until C9, and a singular pointer would force a
schema and resume-path rewrite five phases later if Q6 resolves to
multi-session. A set is cardinality-neutral, so the Q6 answer cannot invalidate
the C4 artifact.

**Lifecycle semantics (required — the pointer is the sole key to the only
durable record; lose it and the transcript is unreachable):**

* **Write ordering:** write **after** `CreateSession` succeeds, using an
  atomic replace (write-temp-then-rename) with an fsync. A crash between
  create and write yields an orphan session, which is recoverable; the reverse
  order yields a pointer to a session that never existed, which is not.
* **A pointer is a HINT, never an assertion.** A failed `ResumeSession` —
  stale, garbage-collected, or expired session — MUST fall back to creating a
  new session and emitting an `error` envelope. It MUST NOT be a hard startup
  failure.
* **Unclean shutdown:** "cleared on clean session end" means a crash leaves a
  pointer to a logically-ended session. That is acceptable precisely because
  of the hint rule above.
* **Protocol-version mismatch** (recorded at creation vs. the CLI present at
  restart): refuse the resume, start fresh, emit an `error` envelope. Do not
  attempt a cross-protocol resume.
* **Writer:** the **agent adapter** (§7.1) owns the write, on the
  session-creation path, **off the hub goroutine**. A filesystem write plus
  fsync inside the hub would violate the never-stall invariant; a write
  elsewhere would create a second writer of session-identity state. The hub
  holds only the in-memory copy.

**H6** must therefore cover these semantics, not merely the file and location.

**Corollary — `seq` needs a generation anchor.** Because a restarted process
re-folds `GetEvents` from zero, hub-assigned `seq` numbering restarts. A client
reconnecting with a pre-restart cursor would otherwise be replayed against a
*different* numbering — silent transcript corruption with no detection path.
The resume cursor is therefore the pair **`(generation, seq)`**, where
`generation` changes on every process start. A cursor whose `generation` does
not match the hub's forces a **full re-hydration**, never a replay. See design
§5.2/§5.3.

This resolves the load-bearing durability conflict recorded in stash
`A92E3FA0` (in-memory state cannot support crash-resume) without inventing a
bespoke persistence layer.

### D5 — First-responder arbitration is backend-authoritative

Both operator surfaces (TUI and SPA) may answer a pending permission request.
The backend holds the single pending-request registry and resolves exactly one
responder by compare-and-swap on the request ID. Losing responders receive an
explicit `permission.resolved` envelope carrying the winning actor, so their UI
clears deterministically rather than by timeout.

### D6 — Slack retirement is total and enforced mechanically

Not merely "unused" — **absent**, with a machine-checkable gate. A CI check
must fail the build if `slack`, `socketmode`, `channel_id`, `team_id`, or
`acp` reappear as identifiers or config keys.

**Gate scope (normative):** `internal/config/**` **excluding `*_test.go` and
any `testdata/` directory**, plus `config.toml.example`, plus `cmd/**`.
Token list: `slack`, `socketmode`, `channel_id`, `team_id`, `acp`, `host_cli`,
`ipc_name`.

Three exclusions are deliberate and each is evidence-driven, not a weakening:

1. **`internal/apperr` is out of scope** (hence `internal/config/**`, not
   `internal/**`). `apperr` legitimately declares `KindSlack`, `KindIPC`, and
   `KindACP` in its 14-variant taxonomy (with `String()` values `"slack"`,
   `"ipc"`, `"acp"`). A gate scoped to `internal/**` would fail immediately on
   a package this correction deliberately freezes.
2. **`*_test.go` and `testdata/` are out of scope.** The corrected package
   MUST retain legacy-shape fixtures containing `[slack]`, `[acp]`,
   `host_cli`, `ipc_name`, and `channel_id` as TOML string literals — they are
   exactly what proves the tolerant-decode migration path (invariant I4). A
   gate that reddens on its own required migration fixtures can never pass.
3. **Documentation, `docs/**`, archived artifacts, Migration sections, and Go
   comments are exempt.**

`cmd/**` is *added* to scope because the retired `intercom-ctl` IPC binary
lives there. It is verified clean today.

**D6a — the residual apperr taxonomy contamination is acknowledged, not
hidden.** `KindSlack`/`KindIPC`/`KindACP` are themselves retired-architecture
surfaces. Removing or renaming them is a **separate error-taxonomy contract
change** with its own blast radius (the taxonomy is pinned by a sentinel-count
test), and it belongs in its own slice. It is captured as a tracked deferred
item rather than silently absorbed here or silently ignored by the gate.

### D6b — Gate durability caveat (honest limits)

`.github/workflows/ci.yml` carries the header
`# Generated by autoharness | Template: ci/ci.yml.tmpl`, so a hand-added step
can be reverted by the next harness render. Further, the workflow's own
threat-model comment records that a `pull_request`-triggered workflow runs the
file **from the PR head**, so the same PR that reintroduces a retired surface
can edit the gate. The gate is therefore an **anti-accident control, not an
anti-adversary control**, and must be documented as such. Making it a
*required* check without the CODEOWNERS protection tracked in stash
`BEDD2E70` would convey assurance it does not have.

### D7 — Persistence is deferred, not decided here

The 7-table SQLite schema in the port brief is Slack-shaped (`slack_ts` in 5 of
7 tables). Under *D4* the SDK runtime owns session durability, so the original
schema's justification largely evaporates. Any residual persistence need
(audit log, retention) is a **separate future deliberation**. This artifact
deliberately does **not** decide it, and no phase may assume the 7-table
schema.

### D8 — The Rust repository is demoted from oracle to historical reference

`C:\Source\GitHub\intercom` (`softwaresalt/agent-intercom` @ `41df772`) is
**historical reference only**. It is **not** a behavioral oracle for operator
UI, transport, configuration schema, or credentials. Its remaining legitimate
value is narrow: general Go-port lessons and non-Slack, non-ACP algorithmic
detail. The "feature-complete Rust port" definition of done is **retired** —
parity with the Rust implementation is no longer a success criterion, because
the two products no longer share an operator surface, a transport, or an agent
boundary.

`references\herdr` remains a non-governing local multiplexer/TUI reference,
untouched and never committed.

### D9 — Sequencing: Option C

Adopt **Option C**. One bounded corrective shipment: governance + config
remediation + SDK proving spike, with three independent rollback points.

#### D9a — Amendment (2026-09-04, post-review): Option C executes as C1 + C2

Option C is executed as **two** sequential shipments rather than one, and this
amendment is recorded explicitly so the deviation is auditable rather than
inferred from the plan:

* **C1** — governance (this artifact + design rev 2, persisted by Stage),
  config remediation, Go 1.24 floor, anti-regression gate.
* **C2** — SDK proving spike: adds the pinned dependency **with its first real
  import**, and proves S1/S2/S3.

**Reason:** `go mod tidy` prunes modules that have no importer, and CI runs a
hard dirty-check on `go mod tidy`. Adding `copilot-sdk/go` in C1 — where the
Non-Goals forbid using it — would be mechanically unstable in CI. Splitting
also honours the operator's explicit "do not create an oversized implementation
shipment" directive. The *content* of Option C is unchanged; only its shipment
boundary moved.

---

## Rejected Alternatives

* **Option A (docs only)** — rejected because the contaminated config schema
  is *actively validated* on `main`. Rule 4 makes `channel_id` mandatory on
  every workspace entry and rule 6 makes `host_cli` mandatory; a config file
  written for the corrected architecture literally **fails to load today**.
  Documentation alone cannot fix an enforced schema.
* **Option B (big bang)** — rejected: oversized shipment, explicitly
  prohibited, and it couples verified work (config) to unverified work (SDK).
* **Option D (three slices)** — rejected: triples cycle count without reducing
  risk, and leaves the harmful schema on `main` for two extra cycles. The
  three parts share one rationale and one review context, so splitting them
  fragments the review rather than sharpening it.
* **Retaining `coder/acp-go-sdk`** — rejected: the official SDK seals the
  protocol layer under `internal/` (F2). An ACP decoder would decode a
  protocol intercom-go no longer speaks.
* **Keeping Slack behind a feature flag** — rejected: the operator retired
  Slack completely, and `slack_ts` contaminates the persistence schema, so a
  flag would preserve the coupling it is meant to hide.

---

## Unresolved Questions

| # | Question | Disposition |
|---|---|---|
| Q1 | Copilot CLI minimum version floor for SDK protocol v3 | **Unknown — not documented.** Research found only `SDKProtocolVersion = 3`. The spike (D2) must record the CLI version it validated against, and that becomes the provisional floor. |
| Q2 | SPA framework | Deliberately out of scope per the operator. |
| Q3 | Dev Tunnel authentication model | Assumption to validate: `--allow-anonymous false` with GitHub identity gating at the tunnel edge, with the backend treating the tunnel as an **untrusted** boundary and performing its own token check. Not yet operator-confirmed. |
| Q4 | Whether to embed the Copilot CLI via `go/embeddedcli` | Deferred. Removes the operator-install variable (R4) but adds binary size and licensing considerations. |
| Q5 | Residual persistence need after D4 | Deferred to a future deliberation (D7). |
| Q6 | Multi-session-per-workspace semantics | The corrected topology is one process per workspace; whether that process hosts >1 concurrent Copilot session determines whether `max_concurrent_sessions` survives. Provisionally retained, semantics re-founded. |

---

## Risks and Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| SDK experimental permission API changes (R1) | high | Anti-corruption layer (D3); spike S1 proves the round-trip before commitment |
| Silent event-variant drop from generated union (R3) | high | Mandatory `default:` branch that logs unknown `event.Type()` (D3); spike S2 |
| SDK/CLI version skew — module pin does not pin CLI (R4) | high | Record the validated CLI version in the spike; add a startup version/protocol assertion; consider embedding (Q4) |
| Config rewrite breaks the just-shipped P2 package | low-medium | Blast radius verified mechanically, not assumed: `internal/config` has **exactly one non-test importer**, `cmd/intercom/main.go`, and it consumes **only** the `config.DefaultConfigPath` constant for a Cobra flag default — no schema type, no `Load`, no `Validate`. Both `cmd` binaries still return a `not implemented` sentinel from `RunE`. So a schema rewrite touches the package and its own tests, plus at most one flag-default line. Migration notes still required in `docs/config-reference.md`. |
| Correction is itself built on a wrong assumption | medium | Every architectural claim in this artifact is traced to verified source at a pinned commit; assumptions are labelled as such (Q1, Q3) |
| Slack/ACP surfaces creep back in | medium | Mechanical CI gate (D6) rather than reviewer diligence |
| Dev Tunnel treated as a trust boundary it is not | medium | Backend performs its own authentication regardless of tunnel gating (Q3) |

---

## Shipped P2 Audit — `internal/config` contamination classification

Audited against the actual shipped code on `main` (merge `2a54791`), not
against the plan. Every surface receives an explicit disposition.

**Verified blast radius:** `internal/config` has exactly one non-test
importer — `cmd/intercom/main.go` — which consumes only
`config.DefaultConfigPath` as a Cobra flag default. Neither `cmd` binary has
implemented behaviour yet (both `RunE` return a `not implemented` sentinel).

### A. Schema surfaces (`internal/config/config.go`)

| Surface | Disposition | Rationale |
|---|---|---|
| `DefaultWorkspaceRoot` | **RETAIN** | Workspace rooting survives; still needed for the per-workspace process (D0.6) |
| `Slack SlackConfig` (`channel_id`, `markdown_upload_extensions`) | **REMOVE** | Slack retired (D0.7). Whole `[slack]` section deleted |
| `SlackDetailLevel` + `DetailMinimal/Standard/Verbose` | **RENAME → migrate** | The *concept* (operator verbosity) survives and applies to TUI and SPA; the Slack-bound name does not. Rename to `operator_detail_level`; keep the three levels and the deliberate non-zero `Standard` default |
| `HostCLI` (`host_cli`) | **MIGRATE (semantics re-founded)** | Nuance: this does **not** simply delete. The Go SDK does not bundle the Copilot CLI (F4.3) and resolves it via `PATH` or `COPILOT_CLI_PATH`, and `StdioConnection.Path` is public. So a CLI-path knob is still legitimate — but re-founded as `[copilot].cli_path` meaning "which runtime binary the SDK should use", **not** "the headless host process we spawn". Validation rule 6 (`host_cli` required) is **ELIMINATED**, not weakened: empty is a defined, valid value meaning "resolve from `PATH`" |
| `HostCLIArgs` (`host_cli_args`) | **REMOVE** | Maps to `StdioConnection.Args`, which is SDK-internal plumbing no operator should hand-tune. See finding **P2-SEC-1** below |
| `ACP ACPConfig` (`max_sessions`, `startup_timeout_seconds`, `max_msg_rate`, `http_port`) | **REMOVE — all four fields deleted** | The `[acp]` section is retired wholesale (D0.1/D0.2). All four defaults are dropped. `max_sessions` is **deleted, not migrated** — its semantics are already covered by `max_concurrent_sessions` (Q6). `startup_timeout_seconds` is **deleted in C1 and re-introduced in phase C2** under `[copilot]`, alongside its first real consumer (the SDK `Client.Start` deadline); adding it in C1 would ship an unconsumed key that no validation rule guards, so an explicitly written `0` would silently become a zero deadline. `max_msg_rate` and `acp.http_port` are deleted outright |
| `Commands map[string]string` | **RETAIN (re-scoped)** | Verified as a generic prompt-alias map (example: `review = "review the changes..."`), not Slack slash-commands. Survives as operator command aliases usable from both TUI and SPA |
| `HTTPPort` | **RETAIN (re-scoped)** | Becomes the single backend HTTP/WebSocket listener port for SPA traffic via Dev Tunnels (D0.5) |
| `IPCName` (`ipc_name`) | **REMOVE (deferred confirm)** | IPC + `intercom-ctl` belong to the retired one-server-many-workspaces topology. Under D0.6 (one process per workspace, autoharness-orchestrated) there is no multi-workspace control plane to address. Default value is literally `"agent-intercom"` — the retired product's name |
| `Timeouts` (`approval_seconds`, `prompt_seconds`, `wait_seconds`) | **RETAIN** | Approval/prompt timeouts map cleanly onto SDK permission handling and turn deadlines |
| `Stall` (5 fields) | **RETAIN** | Stall detection/nudge is agent-surface-agnostic and survives intact |
| `RetentionDays` | **DEFER** | Meaningful only once D7 (persistence) is decided. Do not delete, do not build on |
| `Database DatabaseConfig` | **DEFER** | Same as above. Under D4 the SDK runtime owns session durability, so this may shrink to an audit log or vanish |
| `Workspaces []WorkspaceMapping` | **MIGRATE (field-level)** | `workspace_id`, `label`, `path` **retain**; `channel_id` **removes** (Slack). Under one-process-per-workspace the list itself is under review (Q6) but is not deleted in this slice |
| `MaxConcurrentSessions` | **RETAIN (semantics re-founded)** | Re-founded as concurrent Copilot **SDK** sessions inside one workspace process, not ACP sessions across workspaces. Bound to Q6 |
| `Report` / `UnknownKeys` / `Warnings` | **RETAIN** | Architecture-neutral tolerant-decode machinery; no contamination |
| `MaxConfigBytes` | **RETAIN** | Security hardening, no oracle counterpart |

### B. Validation contract (`internal/config/validate.go`) — 10 rules

| Rule | Disposition |
|---|---|
| 1 — `max_concurrent_sessions` non-zero | **RETAIN** (re-founded per Q6) |
| 2a/2b — `default_workspace_root` set + canonicalizes | **RETAIN** |
| 3 — per-entry `workspace_id` non-empty | **RETAIN** |
| **4 — per-entry `channel_id` non-empty** | **REMOVE** — Slack-mandatory field |
| 5 — `workspace_id` uniqueness | **RETAIN** |
| **6 — `host_cli` must be set** | **ELIMINATED as a rule** — not relaxed. Empty `cli_path` is a defined valid value ("resolve from `PATH`"), so no check, no message, and no contract entry survives |
| 7 — absolute `host_cli` must exist | **RETAIN, rebased and WIDENED** onto `[copilot].cli_path`. This single rule slot now carries the whole `cli_path` contract, including the new relative-path rejection (design §6.2) |
| **8 — duplicate `channel_id` across entries** | **REMOVE** — dies with rule 4 |
| 9 — each non-empty `[[workspace]].path` canonicalizes | **RETAIN** |
| 10 — `database.path` no `..` segment | **RETAIN, but see stash `7028FBE7`** — reimplements a weaker traversal check instead of reusing `internal/pathsafe`. Unchanged by this correction; that deferred entry stays valid |

Net effect: **10 rules → 7**. The arithmetic is stated explicitly because this
package's history already contains two counting errors:

* Rules **4** (`channel_id` required) and **8** (`channel_id` uniqueness) are
  **removed** → 10 − 2 = 8.
* Rule **6** (`host_cli` must be set) is **eliminated as an enumerated rule**,
  not merely weakened. "Empty `cli_path` is valid" means there is no check, no
  error message, and no contract entry left — the rule ceases to exist → 8 − 1
  = **7**.
* The retained absolute-path existence check (former rule 7) is therefore the
  **only** surviving `cli_path` rule.

Post-change contract, enumerated by name (this list, not the number, is
normative): (1) `max_concurrent_sessions` non-zero; (2) `default_workspace_root`
set and canonicalizes (2a/2b — one rule, two message templates, matching the
shipped counting convention); (3) `workspace_id` non-empty; (4) `workspace_id`
unique; (5) **`cli_path` is well-formed** — empty, or a bare name, or an
absolute path that exists; a **relative** path is rejected; (6)
`[[workspace]].path` canonicalizes; (7) `database.path` has no `..` segment.

**Rule 5 is one rule with three branches, not three rules.** The relative-path
rejection (design §6.2) lives inside this slot deliberately, so the
compensating control added for PA-1 does not silently become an unenumerated
eighth check that escapes both the contract pin and the doc-completeness test.

**`cli_path` classification predicate (normative, because "bare name" and
"relative" are otherwise indistinguishable):**

```text
cli_path == ""                                    -> valid (SDK resolves via PATH/COPILOT_CLI_PATH)
filepath.IsAbs(cli_path)                          -> valid iff the file exists
matches ^[A-Za-z]: and not absolute               -> REJECT (Windows drive-relative, e.g. "C:copilot")
contains '/' or '\' (but not absolute)            -> REJECT (relative path)
otherwise (no separator)                          -> bare name; PATH advisory only, never fatal
```

Both separators are tested regardless of `GOOS`, matching the existing
`containsDotDotSegment` precedent in this package, because `config.toml` is
portable text that may be authored on a different platform than it runs on.

### C. Operator-facing artifacts

| Artifact | Disposition |
|---|---|
| `config.toml.example` | **REWRITE.** Currently ships `[slack]`, `[acp]`, `host_cli = "claude"`, `channel_id` on the workspace entry, and `slack_detail_level`. A bidirectional drift test (`example_test.go`) locks example↔schema, so both move together |
| `docs/config-reference.md` | **REWRITE** (26 contaminated matches). A completeness test (`docs_test.go`) locks doc↔schema, so this is mechanically enforced |
| `internal/config/hostcli_test.go` | **RENAME + REWRITE** (25 matches) — largest single test-surface change |
| `default.go` — 17 non-zero defaults | **RECOUNT.** Removing `[acp]` (4 defaults), `ipc_name`, and Slack detail renaming changes the pinned count. The "17 defaults" invariant must be recomputed and re-pinned, not preserved |
| `cmd/intercom-ctl` | **DEFER (flag only).** Stub binary for the retired IPC control plane; no config coupling today. Do not extend it |

### D. Findings requiring explicit operator attention

* **P2-SEC-1 (security, high).** The shipped `config.toml.example` contains
  `host_cli_args = ["--dangerously-skip-permissions"]`. Under the corrected
  architecture, permission handling is a **first-class product responsibility**
  (`SessionConfig.OnPermissionRequest`, D5 first-responder arbitration). An
  example that instructs the agent host to skip permissions is directly
  contrary to the product's core value proposition and is a
  copy-paste-into-production hazard. It must be removed with the rest of
  `host_cli_args`, and its removal called out in the migration notes.
* **P2-SEC-2 (security, informational).** `host_cli = "claude"` in the shipped
  example names a **different vendor's** agent CLI. Harmless as a placeholder
  under the old headless-host model; actively misleading under D0.1.
* **P2-COMPAT-1 (correctness, high).** Rules 4 and 6 mean a `config.toml`
  written for the *corrected* architecture **fails to load today** — it would
  be rejected for a missing `channel_id` and a missing `host_cli`. This is the
  concrete proof that Option A (documentation-only correction) is insufficient.
* **P2-NAME-1 (informational).** `database.path` default is
  `"data/agent-rc.db"` in `default.go` but `"data/agent-intercom.db"` in the
  example. Both are legacy names; both are deferred with D7, but the
  divergence should be recorded so the drift test is not mistaken for
  agreement.

---

## Missing-Context Caveats

These are stated explicitly so downstream planning does not mistake them for
settled facts:

1. **Q1 (CLI version floor) is genuinely unknown** and is not resolvable from
   the SDK repository. It must be pinned empirically by the spike.
2. **Q3 (Dev Tunnel auth) is an assumption, not operator-confirmed.** The
   design records it as an assumption and requires backend-side authentication
   independent of it.
3. **The SDK's public API was verified at `v1.0.11`, not at HEAD.** A public
   API diff between the pin and HEAD was not performed. This is acceptable
   because the pin is what will be used.
4. **SPA framework, SPA implementation, and the SPA↔backend authentication
   handshake beyond the envelope contract are unspecified** and deliberately
   left so.
5. **The event envelope schema in the revised design is intercom-go-owned and
   newly authored.** It is not derived from the SDK's wire types and must not
   be assumed to mirror them.

