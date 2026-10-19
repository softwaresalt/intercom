---
title: "intercom-go MVP: primary end-to-end scenario, scope cut, and gating decisions"
date: 2026-10-19
stash_ids: [A92E3FA0]
status: proposed (operator ratification requested for MVP-D3, MVP-D4, MVP-D6; all others ratify-by-exception)
governing: docs/design-docs/intercom-go-backend-architecture.md (revision 2)
---

## Frame

Operator request (2026-10-19, normal pipeline): reprioritize an MVP of the full primary
scenario ahead of the queued deliberations (003-DL to 005-DL) and the gate/tooling
residual stash, and stage a DAG to reach it.

Verified starting state (main at 28b4850): `cmd/intercom` and `cmd/intercom-ctl` return
`errNotImplemented`. Production packages are `internal/config`, `internal/pathsafe`, and
`internal/apperr`. `internal/copilotprobe` is a disposable SDK spike that nothing imports
(design section 7.6). The roadmap C1 (config remediation) and C2 (SDK proving spike) are
shipped; C3 to C10 of `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md`
are not started. Stash A92E3FA0 (feature-shaped, governing product direction, no
`DEFERRED SCOPE EXPANSION` marker) is the single traced stash entry; feature 043-F is its
blocked placeholder.

Intake obligations (P-021 C5/C6): A92E3FA0 carries no deferred-expansion marker, so the
forced-deliberation precedence does not apply and the duplicate scan is not triggered by it.
A keyword scan of the active stash found no other entry that is an MVP prerequisite. Clean.
No unrelated stash entry is consumed or re-prioritized.

## The primary scenario (MVP-D1)

One workspace, one session, one operator with two surfaces.

1. The operator starts `intercom` for a workspace. Config loads and validates
   (`internal/config`, 7 rules); the workspace root resolves through `internal/pathsafe`.
2. The process starts the Copilot SDK client (`Start` is the readiness gate, design
   section 3), asserts the validated CLI version, and creates one session.
3. The listener binds loopback only. A per-process bearer credential is minted and shown
   to the local operator. The operator (or autoharness) exposes the port through a Dev
   Tunnel; tunnel provisioning is outside the process.
4. The local Bubble Tea TUI renders the streamed transcript. A remote client authenticates
   (bearer plus `Origin` check), hydrates, and receives the same envelopes.
5. The operator sends a prompt (steer). The agent runs a turn and requests a permission.
   Both surfaces show the request. The first authorized responder wins (design 5.4); the
   loser is told why. A destructive request answered from a remote surface is refused per
   the authorization model unless policy allows it.
6. The remote client drops and reconnects with its `(gen, seq)` cursor; it replays or
   re-hydrates with no gap and no duplicate.
7. The operator quits. Shutdown runs the section 3.1 order; no CLI child process remains.

Definition of done: the headless end-to-end test (fake agent through the real hub, real
authenticated WebSocket server, Go reference client, and TUI model driven in-process)
passes for steps 4 to 7, and the opt-in live-SDK test (`INTERCOM_LIVE_SDK_TESTS`) passes
for steps 2, 5, and 7 against a real runtime.

## In scope

* Agent adapter (design 7.1): intercom-owned interface, single-file permission conversion,
  event normalization with logging `default:` (4.3), callback latch (4.2), version assertion.
* Hub (4.1): sole writer, non-blocking sends, envelope v1 with `(gen, seq)`, bounded
  projection, hydrate and replay as single hub steps (5.2, 5.3).
* First-responder arbitration, hub-owned timeout, capacity-1 reply, `permission.aborted`,
  authorization policy (5.4, 6.1).
* Shutdown executor (3.1).
* HTTP/WS transport: loopback bind, bearer auth, `Origin` validation, per-message
  authorization, write pump and drop-on-full backpressure (5.5, 6).
* Bubble Tea TUI: viewport transcript, own buffered feed with drop-oldest and forced
  re-hydrate, permission modal, prompt entry (4, 4.1).
* Process wiring, signal handling, operator runbook, headless and live end-to-end tests.
* Deleting `internal/copilotprobe` and re-pointing the depguard allowlist (7.6 entry actions).

## Out of scope (recorded, not forgotten)

* SPA implementation and framework (design section 9). The remote surface for MVP is the
  documented authenticated WebSocket protocol plus a Go reference client used by tests.
* Session persistence and the D4a session pointer (H6, C11). A restart starts a fresh
  session; `gen` change forces full re-hydrate, so no corruption is possible.
* Multiple sessions per process (Q6/H4, deferred to C9 scope), autoharness spawn logic,
  Dev Tunnel creation, `intercom-ctl`, workspace-tooling transport (Q7), delegation engine
  (RC-5), automated git remediation (RC-6/R8), web push, Slack, any ACP surface.
* Reliability and security residual stash (about 100 entries) and 003-DL to 005-DL.
  No real dependency exists between them and the MVP.

## Decisions

| ID | Question | Options | Recommended default | Operator-blocking |
|---|---|---|---|---|
| MVP-D1 | What is the primary scenario | Section above | As written | No |
| MVP-D2 | Cut: TUI-first, SPA-first, or backend-first | (a) TUI only; (b) remote only; (c) backend core, then TUI and authenticated WS in parallel lanes | (c). (a) never proves the co-equal arbitration or the tunnel boundary; (b) leaves the operator without a local surface | No (ratify by exception) |
| MVP-D3 | Tunnel and WS authentication (Q3/H3, design 6, 6.1) | (a) per-process random bearer; (b) GitHub identity token validation; (c) edge-only | (a): 256-bit `crypto/rand` token minted at start, fixed lifetime, bound to `gen`, revoked at shutdown, constant-time compare, carried in `Authorization` for native clients and in a handler-selected WS subprotocol for browsers (browsers cannot set `Authorization` on a WebSocket), never in the URL; separate observer and approver credentials; logs and diagnostics redact both headers. (b) is a post-MVP upgrade. (c) is forbidden by design 6 | **Yes** |
| MVP-D4 | Authorization model (H5, design 5.4, 6.1) | (a) all surfaces equal; (b) observer/approver roles, remote destructive approval opt-in; (c) TUI-only approvals | (b): TUI is always approver; remote tokens carry a role; remote approval of shell and write kinds is refused unless config opts in; every privileged action re-validates token expiry, role, and `gen` | **Yes** |
| MVP-D5 | `Origin` handling (6.1) | allowlist vs none | Config allowlist, default empty. A present `Origin` not on the list is rejected. A missing `Origin` is accepted only with a valid bearer (native clients) | No |
| MVP-D6 | Tunnel anonymous access (Q3 assumption) | Confirm disabled | Confirm `--allow-anonymous false` in the runbook. Backend auth is mandatory regardless | **Yes** (confirm) |
| MVP-D7 | Sessions per process (Q6/H4) | 1 or N | 1 for MVP. N is C9 scope | No |
| MVP-D8 | Persistence (H6/D7) | pointer, none | None in MVP | No |
| MVP-D9 | WebSocket library | `coder/websocket` vs other | `coder/websocket` (already an indirect dependency; design section 2 requires ping control from the write pump; becomes a direct dependency) | No |
| MVP-D10 | Copilot CLI version floor (Q1) | recorded | `1.0.84-1`, SDK protocol 3, per C2 findings; asserted at start | No |

## Gating deliberations and why they cannot be skipped

Design 6.1 says the controls "MUST be specified before C6/C10" and that authorization
"Blocks C5". The DAG therefore front-loads MVP-D3, MVP-D4, MVP-D5 and token lifecycle as
decision tasks in the first shipment. They gate the arbitration policy task (S4) and the
entire WebSocket transport (S5). Without recorded decisions Stage will not let those
shipments become eligible: the dependency edges enforce it.

## Risks carried into the plan

* Largest risk is concurrency correctness in the hub, arbitration, and write pumps.
  Mitigation: each unit is test-first with `-race` and a concurrent producer property test.
* The SDK pin is empirical for serialization (SQ-b) and dispatch independence (SQ-c); the
  live opt-in test re-verifies both.
* First production caller of `pathsafe.Root.Resolve` (process wiring task) will fire the
  034.005-T tripwire and makes 037-F, 038-F, 039-F unblockable. They remain
  operator-gated; this MVP does not unblock them.

## Recommendation

Proceed with option MVP-D2(c). Ratify MVP-D3, MVP-D4, MVP-D6 before shipment S1 executes
(the decision tasks record the default unless the operator overrides on the task).
