---
title: "ADR MVP-1: MVP scope, session model, and dependency decisions"
date: 2026-10-19
status: decided
decides: [MVP-D1, MVP-D2, MVP-D7, MVP-D8, MVP-D9, MVP-D10]
source: docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md
plan: docs/plans/2026-10-19-intercom-go-mvp-primary-scenario-plan.md
governing: docs/design-docs/intercom-go-backend-architecture.md
backlog: [064-F, 064.001-T]
stash_ids: [A92E3FA0]
---

# ADR MVP-1: MVP scope, session model, and dependency decisions

## Status

Decided on 2026-10-19 (shipment 056-S, task 064.001-T). The operator was away
(dark mode) and left no override comment on 064.001-T, so each recommended default in the
deliberation is ratified by exception. Any later change requires a new ADR that supersedes
this one.

## Context

The deliberation (`docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md`)
proposes an MVP of the primary scenario: one workspace, one Copilot SDK session, one
operator with a local TUI surface and an authenticated remote surface. `cmd/intercom` and
`cmd/intercom-ctl` are still stubs, and C3 to C10 of the architecture correction plan are
not started. Six decisions are not operator-blocking and are recorded here.

## Decisions

### MVP-D1: Primary scenario

The MVP is the seven-step scenario in the deliberation (start, SDK session, loopback
listener with per-process credentials, local TUI plus remote client, first-responder
permission arbitration, gap-free reconnect, clean shutdown). Definition of done:

* The headless end-to-end test (fake agent, real hub, real authenticated WebSocket server,
  Go reference client, TUI model driven in-process) passes for scenario steps 4 to 7.
* The opt-in live-SDK test (`INTERCOM_LIVE_SDK_TESTS`) passes for steps 2, 5, and 7 against
  a real runtime.

Rationale: this is the smallest scenario that exercises every co-equal-surface and
tunnel-boundary risk in design sections 4, 5, and 6.

### MVP-D2: Cut is backend core first, then TUI and authenticated WebSocket lanes

Option (c) of the deliberation: build the agent adapter, hub, arbitration, and shutdown as
a backend core, then the TUI lane and the authenticated WebSocket lane in either order.

Rationale: a TUI-only cut never proves co-equal arbitration or the tunnel boundary, and a
remote-only cut leaves the operator with no local surface. Rejected: (a) TUI only, (b)
remote only. Ship executes one shipment at a time (P-001, P-016), so the lanes are
dependency facts and not concurrent execution.

### MVP-D7: One session per process

One workspace and one Copilot session per process. Multiple sessions per process (Q6/H4)
remain deferred to C9 scope.

Rationale: it removes session routing from the hub, registry, and protocol, and matches the
per-workspace process lifecycle in design section 3.

### MVP-D8: No persistence in the MVP

No D4a session pointer and no `database` or `retention_days` use. A restart starts a fresh
session. Because `gen` (design section 5.2) changes on every process start, a reconnecting
client holding a stale `(gen, seq)` cursor is forced to a full re-hydrate, so a restart
cannot corrupt a client view. Design section 5.1 is amended to say the D4a pointer is cut
from the MVP and stays a post-MVP item (H6, C11).

Consequence: `Client.ResumeSession` and `Session.GetEvents` are not on the MVP path, so the
resume-path de-duplication re-verification in design section 5.1 is not an MVP gate.

### ModeEmpty and AvailableTools

Decision: the MVP runs the SDK in its **standard mode**. It does not set
`ClientOptions.Mode = ModeEmpty` and does not set an explicit `SessionConfig.AvailableTools`
allowlist. This is the default recorded for the choice, to be revisited after the MVP
(design section 7.3 asks only that the mode be evaluated).

Rationale:

* `ModeEmpty` requires `BaseDirectory`, `SessionFS`, or a `URIConnection` for persistent
  session state (design section 7.3). That pulls persistence-adjacent configuration into a
  scope that cut persistence (MVP-D8).
* An explicit `AvailableTools` list would need a tool-by-tool review, which is a separate
  piece of work with its own security analysis.
* The MVP has one session per process and one workspace per process (MVP-D7), so the
  multi-tenant isolation motivation in section 7.3 does not apply yet.

Accepted risk and mitigation: in standard mode the agent has its built-in tools, and
telemetry and user-environment settings are not disabled as `ModeEmpty` would do (design
7.3). The permission path is the intended control over shell and write actions, but that is
an assumption until it is tested: the live-SDK test (MVP-D1) must show that a shell call and
a write call each reach the permission handler, and the MVP does not ship until it does.
The MVP also treats
the authorization matrix (ADR MVP-3) as mandatory, denies unknown permission kinds by
default, and never starts the SDK with a flag that skips permission prompts (for example
`--dangerously-skip-permissions`, which design section 8 already removes from the example
config). Revisit trigger: C9 multi-session work, or the first operator request to run a
restricted tool set.

### MVP-D9: `coder/websocket` is a direct dependency

The WebSocket library is `github.com/coder/websocket`. It is already an indirect dependency
(`go.mod`, v1.8.15, pulled in by the SDK). The first transport task promotes it to a direct
`require` and the module stays on a version at or above the one already resolved.

Rationale: design section 2 requires ping control that can run inside the write pump's own
`select` (design section 5.5), and reusing the SDK's library avoids a second WebSocket
implementation in the binary. It also supports handler-selected subprotocols, which ADR
MVP-2 needs for browser credential carriage.

### MVP-D10: CLI version floor and SDK protocol

Floor: Copilot CLI/runtime `1.0.84-1`, SDK protocol version `3`
(`SDKProtocolVersion = 3`), as validated by the C2 proving spike
(`docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md`). The adapter asserts both
at `Client.Start`, regardless of whether the CLI is embedded or resolved through
`[copilot].cli_path`. A mismatch is a startup error, not a warning. The SDK module stays
pinned at `v1.0.11`.

## Consequences

* Plan tasks that depend on this ADR (2.4 and 3.6, through 064.004-T) may rely on these
  decisions without asking the operator again.
* Out of scope for the MVP, recorded and not forgotten: SPA implementation, session
  persistence and the D4a pointer, multiple sessions per process, Dev Tunnel provisioning,
  `intercom-ctl`, workspace-tooling transport (Q7), the delegation engine (RC-5),
  automated git remediation (RC-6/R8), web push, Slack, and any ACP surface.
* Stash entries and deliberations 003-DL to 005-DL are not consumed by this ADR.

## References

* Deliberation: `docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md`
* Plan: `docs/plans/2026-10-19-intercom-go-mvp-primary-scenario-plan.md`
* Design: `docs/design-docs/intercom-go-backend-architecture.md` sections 3, 5.1, 7.2, 7.3, 9
* Companion ADRs: `docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md`,
  `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`
