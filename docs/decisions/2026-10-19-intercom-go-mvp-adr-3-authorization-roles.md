---
title: "ADR MVP-3: authorization role matrix"
date: 2026-10-19
status: decided
decides: [MVP-D4]
source: docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md
plan: docs/plans/2026-10-19-intercom-go-mvp-primary-scenario-plan.md
governing: docs/design-docs/intercom-go-backend-architecture.md
backlog: [064-F, 064.003-T]
stash_ids: [A92E3FA0]
---

# ADR MVP-3: authorization role matrix

## Status

Decided on 2026-10-19 (shipment 056-S, task 064.003-T). MVP-D4 was operator-blocking. The
operator was away (dark mode) and left no override comment on the task, so the recommended
default is ratified and recorded as a default. A later operator override supersedes this ADR
through a new ADR. This ADR blocks the arbitration task (plan S4) and the WebSocket
transport (plan S5).

## Context

Design section 5.4 lets the first authorized responder answer a pending permission request,
and design section 6.1 notes that any connected surface could approve an agent shell
command. Authentication (ADR MVP-2) says who a caller is. This ADR decides what each caller
may do. Option (b) of the deliberation was chosen: observer and approver roles, with remote
destructive approval opt-in. Rejected: (a) all surfaces equal, and (c) TUI-only approvals
(which would make the remote surface read-only and would not prove co-equal arbitration).

## Decision: surfaces and roles

| Surface | Role | How it is established |
|---|---|---|
| Local TUI | `approver`, always | The TUI is an in-process client, not a WebSocket client. Its role is fixed by construction and needs no credential. |
| Remote client with the observer credential | `observer` | ADR MVP-2 credential minted for `observer` |
| Remote client with the approver credential | `approver` | ADR MVP-2 credential minted for `approver` |

A connection has exactly one role, taken from the credential record that authenticated it.
A role cannot be changed by a message. A remote connection can never be `local`.

## Decision: permission kind classes

Permission requests are classified by the adapter (design 7.1) into intercom-owned kinds.
The kind and its class are stored in the hub's pending-request record when the request
is registered. They are read only from that record and never from an inbound client
message, so a client cannot relabel a `shell` request as `read`.

* **Non-destructive:** `read`.
* **Modeled destructive:** `shell`, `write`, and `url`. A remote approver may approve these
  only when `[remote].remote_approve_destructive` is true.
* **Unknown or unmodeled:** `other`, `unknown`, and any kind the adapter or hub does not
  recognize. These are **local-TUI-only for approval** (074-S amendment): a remote approval is
  refused with `destructive_remote_refused` even when `remote_approve_destructive` is true,
  because the request carries only its raw kind. Remote `deny` stays allowed. Moving a kind
  between classes requires an amendment to this ADR, never a silent code change.

The remote approval opt-in is `[remote].remote_approve_destructive` (default `false`).

## Decision: action-by-role matrix

"Y" means allowed, "N" means refused with the reason in the right-hand column.

| Action | Local TUI | Remote approver | Remote observer | Refusal reason when refused |
|---|---|---|---|---|
| Observe: connect, hydrate, replay, receive all envelopes | Y | Y | Y | n/a (credential failures use ADR MVP-2) |
| Prompt-send (steer the agent) | Y | Y | N | `role_not_permitted` |
| Abort the current turn | Y | Y | N | `role_not_permitted` |
| Approve or deny a `read` permission | Y | Y | N | `role_not_permitted` |
| Deny any permission (any kind) | Y | Y | N | `role_not_permitted` |
| Approve a modeled destructive permission (`shell`, `write`, `url`) | Y | only if `remote_approve_destructive` is true, otherwise N | N | `destructive_remote_refused` (approver); `role_not_permitted` (observer) |
| Approve an unknown or unmodeled permission (`other`, `unknown`, unrecognized) | Y | N (never, even with `remote_approve_destructive`) | N | `destructive_remote_refused` (approver); `role_not_permitted` (observer) |
| Reveal credentials, re-issue credentials | Y | N | N | `local_only_action` |
| Quit or shut down the process | Y | N | N | `local_only_action` |
| Any other or malformed action type | refused | refused | refused | `action_unknown` (default deny) |

Notes:

* Denial is always safe, so an approver may deny any kind remotely even when remote approval
  of that kind is refused.
* Prompt-send by a remote approver is allowed when `remote_approve_destructive` is false.
  Any destructive action the prompt triggers still needs a local approval, so the opt-in
  controls the risky decision and not the steering.
* `permission.requested` envelopes, including destructive requests, are broadcast to every
  surface so that observers and approvers see the same modal state.

## Decision: refusal behaviour

* Every refused action returns a hub-produced refusal message, `action.refused`, to the
  **refused surface only** (not broadcast), carrying `reason` (one of the stable codes
  below), the action identifier, and for the loser case the winning actor. It is a
  non-sequenced per-client control message, not a sequenced envelope, so design 5.2's
  broadcast rule does not apply. The transport task must name it in the protocol
  definition, and a refusal that is followed by a connection close must be flushed by the
  write pump before the close.
* Stable reason codes: `role_not_permitted`, `destructive_remote_refused`,
  `local_only_action`, `already_resolved` (first-responder loser, design 5.4 item 3),
  `credential_expired`, `credential_revoked`, `generation_mismatch`, `request_unknown`,
  `action_unknown`.
* Authorization runs **before** the first-responder compare-and-swap. The order inside the
  hub step is: re-validation (below), then the pending-request lookup (`request_unknown`
  for an ID that does not exist or is already finished, `already_resolved` for the loser),
  then the role and kind check, then the CAS. Message size limits are owned by the
  transport task (plan 5.5). A refused action does
  not resolve, consume, or time out the pending request; the request stays open for an
  authorized surface (for example the local TUI) until the hub-owned timeout. A refusal is
  never silent and never broadcast.
* The `permission.resolved` envelope carries the winning actor as the role and surface
  (for example `local-tui` or `remote-approver`), never a credential or a digest.

## Decision: per-message re-validation

Connection-time authentication is not enough (design 6.1). For every inbound privileged
action (every row of the matrix other than observe), the hub step re-validates, in this order:

1. The credential record that authenticated the connection still exists and is not revoked.
2. The credential has not expired, using the hub's own clock.
3. The record's `gen` equals the current process `gen`.
4. The connection's role permits the action (matrix above).
5. For approvals, the permission kind class and `remote_approve_destructive`.

Failures 1 to 3 do not apply to the local TUI, which has no credential record and is never
closed by them. During shutdown (after credentials are revoked) a privileged action is
refused without closing the connection, so the `server.shutdown` broadcast and drain finish.
For remote connections otherwise they send the refusal and then close the connection (policy-violation close
issued by the write pump, per design 5.5), because the credential is no longer valid.
Failures 4 and 5 send the refusal and keep the connection open.

Observe-only connections never send a privileged action, so expiry and revocation reach
them by a connection-level mechanism: each connection's write pump owns an expiry timer set
to the credential's expiry, and the hub signals the connection's `done` channel when a
credential is revoked by re-issue. Shutdown revocation (ADR MVP-2) does not signal `done`;
the write pumps close sockets after the design 3.1 broadcast and bounded drain. A revoked or expired credential therefore
ends a connection within one timer firing or one hub step, not at the next action. There is
no periodic sweep. The credential store is hub-owned (ADR MVP-2), and the upgrade handler
registers a connection with a revocation re-check in the same hub step. Re-validation steps 1 to 3 run for every decoded inbound action, including one that is then refused as
`action_unknown`. A frame that fails to decode is a protocol violation: the connection gets
a policy-violation close. When the hub-owned permission timeout fires, the outcome is deny
(`UserNotAvailable`, design 5.4).

## Addendum: hub-added reason codes and transport-owned codes (shipment 059-S)

* The hub adds three reason codes beyond the stable list above, for conditions the
  list does not cover: `agent_unavailable` (the hub cannot hand a prompt or
  abort to the agent: none configured, or its operator queue is full),
  `shutting_down` (the credential-less local TUI attempted a privileged action
  after shutdown step 0; a remote surface gets `credential_revoked` instead) and
  `client_closed` (see below).
* `credential_expired`, `generation_mismatch` and `local_only_action`, and the
  wire shape of `action.refused`, are owned by the transport (plan S5): the
  hub's arbitration returns the codes it can decide itself and does not
  produce those.
* A queued action whose connection closed before the hub processed it is
  refused (re-validation runs at the hub step), uniformly for every registered
  client: `credential_revoked` for a remote connection, and `client_closed` for
  the local TUI, which has no credential to revoke.
* The hub denies a truncated permission request itself (the adapter already
  does; the hub does not rely on that): an operator cannot approve text it
  cannot see, so such a request is never registered or broadcast.
* Ack-deadline and shutdown settlement: a decision is final only when the
  handler reports what it returned to the SDK. When the acknowledgement is
  overdue, or shutdown step 0 begins, a decision the handler had not yet
  delivered is revoked (it can only ever become `UserNotAvailable`) and the
  request is published `permission.aborted`; a decision the handler already
  delivered is published `permission.resolved` per the handler-reported
  outcome, or `permission.aborted` if the handler had downgraded it.
* Unknown kinds remain destructive for remote approvers (default-deny), but as of the
  074-S amendment below they are no longer covered by `remote_approve_destructive`.

## Amendment (074-S, 2026-10-30)

* **Unknown kinds are local-TUI-only for approval.** `other`, `unknown`, and any
  undefined kind are refused to remote approvers with `destructive_remote_refused`
  regardless of `remote_approve_destructive`; the opt-in covers only `shell`,
  `write`, and `url`. Remote `deny` and local TUI approval are unchanged.
  This supersedes the earlier statement that the opt-in covers unknown kinds.
* **Clipped write previews are flagged out of band.** A write whose diff or content
  preview was clipped carries the boolean `preview_clipped` on `permission.requested`
  and `session.hydrate`. The local TUI shows a column-0 warning header and asks for an
  explicit clipped-write confirmation; the in-text `[write preview truncated]`
  notice is informational only and can be imitated by file content.
* **`urls:` framing is adapter-owned.** Shell possible-URLs are sanitized and
  command lines imitating the `urls:` line are indented.
* `[timeouts].approval_seconds` precedence is documented in `docs/config-reference.md`.

## Accepted exposure

An observer sees full transcripts: prompts, agent output, tool arguments, command lines, and
any file content the agent prints. The MVP does no redaction, so the observer credential
must be treated as sensitive, almost as sensitive as the approver credential. This is
accepted because redaction needs a separate classification design, which is out of scope.

**Read exfiltration.** A remote approver may approve `read` permissions and prompt the agent.
`read` is not scoped to the workspace root in the MVP, so a leaked approver credential could
make the agent read files outside the workspace and the observer stream would show them.
This is accepted for the MVP and reduced by the credential lifetime and revocation in
ADR MVP-2. Scoping `read` to the workspace root (using `internal/pathsafe`) is a candidate
post-MVP control and is not decided here.

## Consequences

* The arbitration task implements authorization as a pure function of (role, action, kind
  class, config) called before the CAS, with a table-driven test over this matrix, including
  unknown kinds.
* The WebSocket task implements the re-validation order above and the refusal envelope.
* Config validation must reject `remote_approve_destructive` values that are not boolean and
  document that enabling it lets a leaked approver credential approve shell and write
  actions.

## Implementation addendum (060-S, 2026-10-20)

Recorded after implementing the WebSocket task in `internal/transport`:

* **Close 1011 `unavailable`** is sent when the hub cannot serve an action (the hub is
  unavailable); the client reconnects later with its last cursor.
* **A role mismatch at re-validation closes the connection.** If the grant's role differs
  from the role the hub client was registered with, the action is refused and the
  connection is closed (defence in depth; the two come from the same credential).
* **Revocation closes via the store's `Grant.Revoked` signal** with 1008 `session ended`;
  expiry and failed re-validation use the same neutral close. Shutdown revocation refuses
  without closing so the drain can finish.
* **`local_only_action` is unrepresentable remotely.** The wire protocol has no operation
  that maps to it, so an unknown `op` string yields the hub's `action_unknown` refusal and
  the connection stays open.
* **Per-connection action rate limiting is implemented** (`rate_limited`, close after
  repeated violations); only upgrade rate limiting remains deferred (see ADR MVP-2).

## References

* Design sections 5.4, 6, 6.1, 7.1
* ADR MVP-2: `docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md`
* ADR MVP-1: `docs/decisions/2026-10-19-intercom-go-mvp-adr-1-scope-and-session.md`
