---
title: "intercom-go MVP primary scenario: DAG implementation plan"
date: 2026-10-19
source: docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md
governing: docs/design-docs/intercom-go-backend-architecture.md
stash_ids: [A92E3FA0]
---

## Summary

Seven shipments, 7 features, 46 tasks deliver the primary scenario defined in the
deliberation (operator starts a per-workspace intercom, a Copilot SDK session is hosted,
events and permission requests reach the local TUI and an authenticated remote surface,
first-responder approval and steering work, reconnect is gap-free, shutdown is clean).
This plan decomposes roadmap phases C3 to C10 of
`docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md` and re-orders them
into a true dependency DAG: security decisions are front-loaded, hydration is built with
the hub (not after transport), and the TUI and WebSocket lanes are independent after
arbitration. Ship executes one shipment at a time (P-001/P-016); parallel lanes below are
dependency facts and planning-overlap opportunities, not concurrent execution. Stage
writes no code.

Requires plan hardening: yes

## Source of truth

Deliberation `docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md`
(MVP-D1 to MVP-D10). Where the design is silent (WS library, token carriage) the
deliberation defaults apply and the S1 ADRs finalize them.

## Review revision log

Plan review attempt 1 (Architecture Strategist: ADVISORY; Security Lens: FAIL) found
missing edges, a missing config task, a missing adapter-to-ingress integration task,
overloaded tasks, and security gaps (token carriage, per-role credentials, expiry
ownership, action authorization, pre-auth DoS limits, shutdown ordering, token display).
This revision adds tasks 3.6 to 3.9, 4.8, 5.7, 5.8, 7.6, splits overloaded tasks, adds the
missing edges, and moves each security control to an owning task. See the Plan Review
section at the end.

## Nodes

Sizes: XS, S (about 1h), M (about 2h). Complexity is a separate axis. One skill domain per
task. Code and tests tasks are test-first. "Blocks-on" lists predecessors (every entry is a
real `blocks` edge).

### S1 Feature F1: MVP scope and security decisions (gating)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 1.1 | ADR: MVP scope, one session per process, no persistence, `ModeEmpty` and `AvailableTools` choice, WS library, CLI version floor (MVP-D1,D2,D7,D8,D9,D10) | docs | S | low | none | 3, 5.1, 7.2, 7.3, 9 |
| 1.2 | ADR: credential model and carriage per client type (SPA subprotocol, native header), per-role credentials, lifetime (renewal for MVP is operator re-issue through the local TUI), binding, replay, secrets from env not config, `Origin` and `Host` rules, tunnel anonymous-off confirmation (MVP-D3,D5,D6) | docs | S | medium | none | 6, 6.1 |
| 1.3 | ADR: role matrix covering approve, prompt-send, abort, observe; destructive and unknown permission kinds deny by default; observer visibility of transcripts; per-message re-validation (MVP-D4) | docs | S | medium | none | 5.4, 6.1 |
| 1.4 | Amend design 5.1, 6.1, 8 (new validation rule group), 9 to "decided"; MVP product-spec pointer; annotate 043-F | docs | XS | low | 1.1, 1.2, 1.3 | 5.1, 6.1, 8, 9 |

Decision tasks record the recommended default. The operator overrides by commenting on the
task before S1 ships; the ADR then records the override.

### S2 Feature F2: Agent adapter anti-corruption layer (C3)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 2.1 | Domain types and `Agent` interface in `internal/agent` (no SDK import); SDK adapter lives in subpackage `internal/agent/copilot` | code | M | medium | none | 7.1 |
| 2.2 | Event normalization table with logging `default:` for unknown variants | code | M | medium | 2.1 | 4.3, 5.2 |
| 2.3 | Permission decision conversion (single file) and blocking bridge with hub-done exit | code | M | medium | 2.1 | 5.4, 7.1 |
| 2.4 | Client `Start` with re-introduced startup timeout field, CLI version assertion, session create per ADR 1.1 | code | M | medium | 2.1, 1.1 | 3, 7.2, 7.3, 8 |
| 2.5 | Callback entrant latch, unsubscribe and bounded quiesce | code | S | medium | 2.4 | 4.2 |
| 2.6 | `Send` and `Abort` operations with deadline contexts | code | S | low | 2.4 | 7.5 |
| 2.7 | Delete `internal/copilotprobe`, re-point depguard allowlist at `internal/agent/copilot` only, record live-test convention | config | S | low | 2.2, 2.3, 2.5, 2.6 | 7.6 |

### S3 Feature F3: Hub core, envelope, hydration, config (C4 and C7)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 3.1 | Envelope v1, `gen` from UUIDv7/ULID captured once, immutable pre-marshalled publish | code | S | medium | 2.1 | 5.2 |
| 3.2 | Session-state fold with entry and byte caps declared | code | M | medium | 3.1 | 5.1, 5.3 |
| 3.3 | Hub loop: ingress, sole writer, client registry with client kinds (permanent local TUI member, remote), seq-and-broadcast single step | code | M | medium | 3.1 | 4, 4.1, 5.4 |
| 3.4 | Connect hydrate as one hub step with per-iteration hydration cap | code | M | medium | 3.2, 3.3 | 5.3 |
| 3.5 | No-gap/no-duplicate property test with concurrent producer under `-race` | tests | S | medium | 3.4, 3.9 | 5.3 |
| 3.6 | Config: `[remote]` fields (token lifetime, `allowed_origins`, `max_connections`, `remote_approve_destructive`) and validation | config | M | medium | 1.4 | 8 |
| 3.7 | Hub non-blocking-send enforcement, capacity-1 `degraded` channel, hub-produced `error` envelope and forced re-hydrate | code | S | medium | 3.3 | 4.1, 4.2 |
| 3.8 | Adapter-to-ingress integration test: blocking send, degraded signal, no silent drop | tests | S | medium | 2.5, 3.3, 3.7 | 4.2 |
| 3.9 | `(gen, seq)` reconnect, bounded replay ring, coverage check, gap and idempotence rules | code | M | medium | 3.4 | 5.3 |

### S4 Feature F4: Permission arbitration, operator actions, shutdown (C5 and shutdown half of C9)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 4.1 | Pending registry, register-before-broadcast, CAS first responder | code | M | medium | 3.3 | 5.4 |
| 4.2 | Hub-owned timeout, capacity-1 reply, `permission.resolved` and `permission.aborted` | code | M | medium | 4.1, 2.3 | 5.4 |
| 4.3 | Pending permissions and remaining deadlines carried in hydrate | code | S | low | 4.1, 3.4 | 5.4 |
| 4.4 | Authorization policy enforcement per ADR 1.3 (roles, remote destructive refusal with explicit reason, unknown kinds denied) | code | M | medium | 4.1, 1.3, 3.6 | 5.4, 6.1 |
| 4.5 | Route prompt-send and turn-abort actions through hub to the adapter, authorized by the role matrix | code | S | medium | 3.3, 2.6, 4.4 | 4.1, 6.1 |
| 4.6 | Shutdown executor order test with fake agent: step 0 also revokes credentials and closes the listener via a `Gate` interface; hub keeps draining to step 6 | code | M | medium | 4.2, 2.5 | 3.1 |
| 4.7 | Shutdown steps 2 to 4 against real adapter: abort, positive idle, `Stop`, `ForceStop` fallback | code | S | medium | 4.6, 2.4, 2.6 | 3.1 |
| 4.8 | Empty-broadcast-set fast resolve using the permanent TUI membership rule | code | S | medium | 4.1, 3.3 | 5.4 |

### S5 Feature F5: Authenticated WebSocket transport (C6 and C10)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 5.1 | Loopback-only listener lifecycle and bind-address assertion | code | S | low | none | 6 |
| 5.2 | Per-role credential mint (observer, approver), constant-time verify, lifetime, bound to `gen` and connection, `Revoke`, per-connection actor ID | code | S | medium | 1.2, 1.3, 3.1, 3.6 | 6.1 |
| 5.3 | Upgrade handler: authenticate before upgrade, `Origin` and `Host` checks, subprotocol selection | code | M | medium | 5.1, 5.2, 3.6 | 6, 6.1 |
| 5.4 | Write pump: single writer, `done`/`sync.Once` drop protocol, ping inside pump, drop-on-full, credential-expiry deadline case closes via pump | code | M | medium | 3.3, 3.7, 5.3 | 5.5 |
| 5.5 | Read pump: read size limit, idle deadline, action decode, hydrate or replay on connect | code | M | medium | 5.3, 4.5, 3.4, 3.9 | 5.3 |
| 5.7 | Per-message authorization re-validation (token, role, expiry, `gen`) and inbound action rate limit | code | M | medium | 5.5, 4.4 | 6.1 |
| 5.8 | Pre-auth limits: handshake timeout, connection cap counting unauthenticated sockets; header and credential redaction test | code | S | medium | 5.3 | 6.1 |
| 5.6 | Go reference client and integration tests: slow client, CSWSH, expired and wrong-role token, mid-connection expiry close, reconnect replay | tests | M | medium | 5.4, 5.7, 5.8 | 5.5, 6.1 |

### S6 Feature F6: Local Bubble Tea TUI (C8)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 6.1 | Add bubbletea, lipgloss, bubbles; pure Elm model skeleton with golden frames | code | M | medium | 3.1 | 2, 4 |
| 6.2 | Buffered feed channel and pump, drop-oldest, forced re-hydrate | code | S | medium | 3.3, 6.1, 3.4 | 4 |
| 6.3 | Transcript viewport with streaming deltas | code | M | medium | 6.1, 6.2, 3.2 | 2, 5.1 |
| 6.4 | Permission modal, prompt input via hub ingress `tea.Cmd`, clear on remote resolution | code | M | medium | 6.3, 4.2, 4.3, 4.5, 4.8 | 4.1, 5.4 |

### S7 Feature F7: Process wiring and MVP verification (C9 and acceptance)

| Key | Task | Domain | Size | Cx | Blocks-on | Design |
|---|---|---|---|---|---|---|
| 7.1 | `cmd/intercom` wiring: flags, config, `pathsafe` workspace root, component startup order | code | M | medium | 4.7, 5.6, 6.4 | 3, 6 |
| 7.2 | Signal and TUI-quit shutdown wiring, real credential revoke and listener close, no stranded CLI child | code | S | medium | 7.1, 4.7 | 3.1 |
| 7.3 | Headless end-to-end scenario test (fake agent, real hub and WS, reference client, TUI model) | tests | M | medium | 7.1, 7.6 | 1.1 scenario |
| 7.4 | Opt-in live-SDK end-to-end test behind `INTERCOM_LIVE_SDK_TESTS` | tests | S | medium | 7.2 | 7.6 |
| 7.5 | Operator runbook: start, tunnel exposure with anonymous access off, credential handling | docs | S | low | 7.1, 1.4, 7.6 | 3, 6 |
| 7.6 | Show credentials once on the local TUI only; log-capture test proves no credential in logs or transcript | code | S | medium | 7.1, 5.2 | 6 |

## Dependency matrix (feature and shipment level)

Cell X = row shipment has a true dependency on the column shipment.

| Row blocks-on col | S1 | S2 | S3 | S4 | S5 | S6 | S7 |
|---|---|---|---|---|---|---|---|
| S1 | | | | | | | |
| S2 | X | | | | | | |
| S3 | X | X | | | | | |
| S4 | X | X | X | | | | |
| S5 | X | | X | X | | | |
| S6 | | | X | X | | | |
| S7 | X | | X | X | X | X | |

Shipment-level blocks-edges recorded in the backlog are the minimal set: S2 on S1,
S3 on S2, S4 on S3, S5 on S4, S6 on S4, S7 on S5 and S6 (the rest hold by transitivity).
The task-level matrix is the "Blocks-on" column of each node table: 89 `blocks` edges
across 46 tasks. Every edge is a direct dependency; a few (4.5 on 3.3, 4.8 on 3.3, 6.3 on 6.1,
5.5 on 3.4, 3.5 on 3.4, 7.2 on 4.7) are also implied transitively and are kept explicit
as documentation of direct use.

## Critical path and lanes

Critical path (M = 2h, S = 1h, XS = 0.5h; lower bound with unlimited parallelism 21h):
`2.1 -> 3.1 -> 3.3 -> 4.1 -> 4.4 -> 4.5 -> 5.5 -> 5.7 -> 5.6 -> 7.1 -> 7.6 -> 7.3`.
Near-critical (20h): the same prefix then `7.1 -> 7.2 -> 7.4` or `7.1 -> 7.6 -> 7.5`.
Total effort 68.5 human-hours over 46 tasks (S1 3.5, S2 11, S3 14, S4 12, S5 13, S6 7, S7 8).
Task 4.1 (CAS) must not be wired to any transport before 4.4 lands; S4 merges as a unit.

Lanes (Ship executes one shipment at a time):

* Lane A, decisions: S1 (1.1, 1.2, 1.3 are mutually independent)
* Lane B, spine: S2, S3, S4 (sequential)
* Lane C, remote surface: S5 (after S4)
* Lane D, local surface: S6 (after S4); independent of S5, so Stage may plan or review S6
  while S5 executes (P-016 planning overlap only)
* Lane E, convergence: S7 after C and D

Queue order: S1, S2, S3, S4, S5, S6, S7. S5 precedes S6 because it carries the security
controls and the critical path.

## Shipment sequence

| Seq | Shipment | Feature | Tasks | Independently mergeable because |
|---|---|---|---|---|
| 1 | S1 decisions | F1 | 4 | docs only |
| 2 | S2 adapter | F2 | 7 | package unused until S3; removes the dead spike |
| 3 | S3 hub and config | F3 | 9 | library packages, unit and property tests, `-race` |
| 4 | S4 arbitration and shutdown | F4 | 8 | library package, fake-agent tests |
| 5 | S5 WS transport | F5 | 8 | package plus reference client tests; no binary change |
| 6 | S6 TUI | F6 | 4 | package plus golden tests; no binary change |
| 7 | S7 wiring and e2e | F7 | 6 | first runnable intercom; e2e proves the scenario |

Each shipment leaves `main` green (build, `-race` tests, lint, depguard). No shipment
opens a network listener in a binary until S7.

## Plan Hardening

Triggers: auth and authorization behavior, a network listener, concurrency invariants.

* Authentication: per-role credentials minted with `crypto/rand` (at least 256 bits),
  constant-time compare, fixed lifetime with a documented renewal path, bound to `gen` and
  to the connection, never in a URL, never logged (fingerprint only; owned by 5.8 redaction
  test and 7.6 log-capture test). Browser clients carry the credential in a WS subprotocol
  that the handler selects; edge and proxy log exposure is an accepted, documented residual
  (ADR 1.2). Secrets such as `GitHubToken` come from the environment, never `config.toml`.
* Pre-auth hardening (5.3, 5.8): authenticate before upgrade, handshake timeout, `Host`
  validation against DNS rebinding, connection cap counting unauthenticated sockets (5.3 and 5.8).
* CSWSH: `Origin` allowlist (default empty) enforced on upgrade; no ambient credentials.
* Authorization: role matrix covers approve, prompt-send, abort, observe (ADR 1.3, 4.4,
  4.5); re-validated per message (5.7); expiry mid-connection closes the socket through
  the write pump (5.4, tested in 5.6); unknown permission kinds are denied; ambiguous
  authorization defaults to deny.
* Denial of service: read size limit and idle deadline (5.5), action rate limit (5.7),
  per-client bounded buffer with drop-on-full, bounded concurrent hydrations (3.4),
  capped projection (3.2).
* Concurrency: hub is sole writer; no hub blocking send (3.7); no off-hub envelope; the
  write pump is the only socket writer; every task runs `go test -race` and 3.5 uses a
  concurrent producer. Live-SDK test (7.4) re-verifies SQ-b and SQ-c.
* Shutdown: step 0 revokes credentials and closes the listener so no client is admitted
  mid-shutdown (4.6, 7.2); pending permissions resolve to `UserNotAvailable`.
* Data exposure: observers see full transcripts and tool output; recorded as accepted in
  ADR 1.3 unless the operator asks for scoping.
* Rollback: each task is one revertable commit; no data migration; no persistence.
* Residual risks: Dev Tunnel edge misconfiguration (mitigated by mandatory backend auth);
  SDK behavior drift on version bump (design 4.2); PATH-resolved CLI (design 6.2).

## Risks

* Task 7.1 is the first production caller of `pathsafe.Root.Resolve`; the 034.005-T
  tripwire fires. 037-F, 038-F, 039-F become unblockable but stay operator-gated.
* 2.1 defines types every later shipment imports; keep it minimal and additive.
* Estimates assume the C2 findings hold at SDK v1.0.11.

## Operator decisions outstanding

MVP-D3 (auth mechanism), MVP-D4 (authorization model), MVP-D6 (tunnel anonymous off).
Defaults are in the deliberation and in tasks 1.2 and 1.3 (D6 now lives in 1.2, so it no
longer gates the adapter spine through 1.1).

## Plan Review

Gate verdict: PASS (after one revision cycle; attempt counter 2 of 3 not exhausted).

* Attempt 1 (Architecture Strategist ADVISORY with 7 P1; Security Lens FAIL with 5 P1):
  all findings applied in this revision.
* Attempt 2: Architecture Strategist PASS; Security Lens ADVISORY (no P0/P1). Applied the
  advisories: 7.6 edges to 7.3 and 7.5, 5.3 split (new 5.8), 5.4 on 3.7, renewal stated in
  1.2, 4.1 and 4.4 wiring note, corrected path figures. Redundant-but-direct edges are
  documented rather than pruned. Residual: credential in a WS subprotocol can appear in
  edge logs (accepted, ADR 1.2).

<!-- plan-review-attempt: 2 -->
