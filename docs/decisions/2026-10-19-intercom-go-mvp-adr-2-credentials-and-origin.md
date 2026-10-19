---
title: "ADR MVP-2: tunnel credential, carriage, Origin, and Host decisions"
date: 2026-10-19
status: decided
decides: [MVP-D3, MVP-D5, MVP-D6, MVP-D11, MVP-D12]
source: docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md
plan: docs/plans/2026-10-19-intercom-go-mvp-primary-scenario-plan.md
governing: docs/design-docs/intercom-go-backend-architecture.md
backlog: [064-F, 064.002-T]
stash_ids: [A92E3FA0]
---

# ADR MVP-2: tunnel credential, carriage, Origin, and Host decisions

## Status

Decided on 2026-10-19 (shipment 056-S, task 064.002-T). MVP-D3 and MVP-D6 were marked
operator-blocking. The operator was away (dark mode) and left no override comment on the
task, so the recommended defaults are ratified here and recorded as defaults, not as
operator-confirmed choices. A later operator override supersedes this ADR through a new ADR.

## Context

Design section 6 treats the Dev Tunnel as an untrusted boundary and requires the backend to
authenticate every WebSocket upgrade itself. Design section 6.1 lists controls that must be
specified before C6/C10: Origin and cross-site WebSocket hijacking (CSWSH), credential
carriage, token replay and expiry, and per-message authorization. Role semantics are decided
in ADR MVP-3. Any connected surface can approve an agent shell command (design 5.4), so a
leaked or replayed credential is equivalent to remote code execution on the operator's
machine unless the controls below hold.

## Decision: credential model (MVP-D3)

Option (a) of the deliberation: per-process random bearer credentials. Rejected for MVP:
(b) GitHub identity token validation (a post-MVP upgrade) and (c) edge-only authentication
(forbidden by design section 6).

1. **Generation.** At process start the backend mints one credential per role (`observer`
   and `approver`, ADR MVP-3) from `crypto/rand`: 32 random bytes (256 bits), encoded as
   unpadded base64url. The two credentials are independent; neither is derivable from the
   other. A failed `crypto/rand` read is a fatal startup error.
2. **Storage and comparison.** The process keeps only a SHA-256 digest of each credential
   plus its metadata (role, issued-at, expiry, `gen`, revoked flag). Presented credentials
   are digested and compared with `crypto/subtle.ConstantTimeCompare`. The plaintext is held
   only long enough to show it to the operator.
3. **Display.** Credentials are shown only on the local TUI, once per mint or re-issue
   (the operator must copy it then; a later view requires re-issue, so no plaintext is
   retained). A short fingerprint (the first 8 hex characters of the SHA-256 digest) may be
   shown and logged to identify which credential is live. The plaintext is never written to a log, an envelope, a diagnostic dump, `config.toml`, the
   process command line, or disk. Logs and diagnostics redact the `Authorization` header
   and the `Sec-WebSocket-Protocol` header.
4. **Secrets come from the environment, not config.** No credential may be read from
   `config.toml`. The MVP mints credentials and accepts no operator-supplied value. If a
   later change allows operator-supplied credentials, they must come from an environment
   variable, and config validation must reject any credential-shaped key under `[remote]`
   (a key whose name contains `token`, `secret`, `credential`, `password`, or `bearer`,
   other than the documented `token_lifetime_seconds`).
5. **Lifetime.** Each credential has a fixed lifetime from issue, set by
   `[remote].token_lifetime_seconds` (default 14400, that is 4 hours; accepted range 60 to
   86400). There is no sliding expiry and no refresh endpoint. An expired credential is
   treated as absent.
6. **Renewal.** For the MVP, renewal is operator re-issue through the local TUI. Re-issue
   mints a new credential for the chosen role, revokes the old one immediately, and the
   per-message re-validation (ADR MVP-3) closes remote connections that used the revoked
   credential. The operator hands the new credential to the remote user out of band.
7. **Binding.** A credential is bound to the process `gen` (design 5.2) and so cannot
   survive a restart. It is also bound to each connection: the server records which
   credential record authenticated a connection, and every privileged action re-checks that
   record for revocation, expiry, and `gen`. All credentials are revoked at shutdown, in
   the design 3.1 shutdown order, at the step that stops accepting work (plan task 4.6).
   Revocation blocks new admission and further privileged actions only. It does not close
   sockets itself: remote clients still receive the `server.shutdown` broadcast (design 3.1
   step 1) and the write pump closes each socket after the bounded drain. The credential
   store is owned by the hub goroutine (or guarded by one lock), and a connection registers
   with a revocation re-check in the same step, so an upgrade racing a re-issue cannot
   register with a revoked credential.
8. **Replay stance.** This is a bearer credential: anyone holding it within its lifetime can
   use it. The MVP accepts that and mitigates it with TLS at the tunnel, no URL carriage,
   header redaction, the 4-hour default lifetime, immediate revocation on re-issue and
   shutdown, and the connection cap below. Single-use upgrade tickets and channel binding
   are deferred (see "Deferred controls").

## Decision: credential carriage per client type (MVP-D3)

The credential is never carried in the URL (path or query), because URLs are logged by
proxies, tunnels, and browsers.

| Client type | Carriage | Rule |
|---|---|---|
| Native clients (Go reference client, CLI tools) | `Authorization: Bearer <credential>` request header on the upgrade | Preferred. Browsers cannot set this header on a WebSocket. |
| Browsers (future SPA) | WebSocket subprotocol: the client offers `Sec-WebSocket-Protocol: intercom.v1, bearer.<credential>` | The handler selects only `intercom.v1` and never echoes the `bearer.` entry. A request that offers no `intercom.v1` is rejected. |

Neither carriage may use cookies or any ambient credential. If both a header and a
subprotocol credential are present, the request is rejected as ambiguous. A malformed or
missing credential returns the same generic 401 as a wrong one, so the response is not an
oracle.

**Accepted residual risk.** The subprotocol carriage can appear in tunnel or edge logs
(`Sec-WebSocket-Protocol` is a request header that an intermediary may record). Mitigations
are the bounded lifetime, re-issue revocation, and a recommendation that native clients use
the `Authorization` header. This is accepted for the MVP because browsers have no other
non-URL option.

## Decision: Origin and CSWSH (MVP-D5)

* A WebSocket upgrade is not covered by the same-origin policy, so the backend validates
  `Origin` itself on every upgrade, using exact, case-insensitive matching of scheme, host,
  and port against `[remote].allowed_origins`.
* The default allowlist is empty. A present `Origin` that is not on the list is rejected
  (403). With an empty list, every browser origin is therefore rejected. That is intended:
  the SPA is out of MVP scope, and the operator adds the SPA origin when one exists.
* A missing `Origin` is accepted only together with a valid credential. Native clients
  normally send no `Origin`. A missing `Origin` with a missing or invalid credential is
  rejected.
* Canonical form: scheme and host are compared case-insensitively, a trailing dot on a host
  is stripped, an internationalized host is compared in its ASCII (punycode) form, and the
  default port (`:443` for `https`, `:80` for `http`) is equivalent to an omitted port.
  Entries must already be written in this canonical form, and validation rejects others.
* Wildcards (`*`, scheme-only, or host-only patterns) are not allowed in
  `allowed_origins` or `allowed_hosts`. Config validation rejects them. The library's skip-verification and
  wildcard origin options must not be used.
* Authentication is by explicit bearer credential only. Ambient cookies are neither read nor
  set, which removes the CSWSH precondition. A matching `Origin` never replaces a credential.
* Check order on upgrade (cheapest first, fail closed): connection cap, `Host`, `Origin`,
  credential (presence, digest compare, expiry, `gen`, revocation), role. Rejections use
  these statuses and no response body detail: 503 (cap), 403 (`Host` or `Origin`), 401
  (any credential failure, one generic response for missing, malformed, wrong, expired, or
  revoked). The distinct statuses reveal only which public check failed, not anything about
  a credential, and are accepted.

## Decision: Host validation (MVP-D11)

The listener binds loopback only. The request `Host` header (including its port, which is
compared after the same canonicalization as Origin: case-insensitive, trailing dot stripped,
IDN in punycode form, default port equal to an omitted port) must be in an allowed host set,
to defend against DNS rebinding and misrouted tunnel traffic. The set is the loopback names
for the bound port (`127.0.0.1`, `[::1]`, and `localhost`) plus the hosts listed in
`[remote].allowed_hosts` (default empty). A tunnel-facing deployment adds the tunnel
hostname. MVP-D11 is new relative to the deliberation (which named `Host` validation only
in passing) and introduces one config field, `allowed_hosts`, in the `[remote]` rule group
recorded in the design section 8 amendment. Plan task 3.6 lists the other `[remote]` fields
but not this one; the task that implements config must include it (reported as a follow-up
for Stage, since changing the plan is outside this shipment).

## Decision: pre-authentication limits (MVP-D12)

`[remote].max_connections` bounds concurrent upgrades and live remote connections
(default 8). Requests over the cap get 503 before any credential work. The handshake has a
fixed read deadline of 10 seconds, request headers are size-limited, and no application
message is read before authentication succeeds. The exact numeric values for the deadline
and header limit belong to the transport task, which must name them as constants.
Post-authentication read-size limits, idle deadlines, and action rate limits are owned by
the transport tasks (plan 5.5 and 5.7), not decided here.

**Accepted residual risk.** The cap counts unauthenticated sockets, so someone who can reach
the tunnel can hold the slots for the handshake window and delay legitimate reconnects.
This is a denial of service on the remote surface only. It cannot affect the local TUI or
the agent. All tunnel traffic arrives through one relay, so per-IP limiting would not help
anyway. It is accepted for the MVP; a reserved slot pool for authenticated reconnects is a
candidate later improvement. A remote approver locked out this way cannot answer a pending
request remotely, which then falls to the local TUI or the hub-owned timeout.

## Decision: tunnel anonymous access (MVP-D6)

The runbook requires the Dev Tunnel to be created with `--allow-anonymous false`
(gating on GitHub identity at the Azure edge). This is recorded as the **default
confirmation**: the operator did not comment, and the tunnel is created outside the
process, so the backend cannot verify it. Backend authentication in this ADR is mandatory
regardless of tunnel settings, so a tunnel misconfigured to allow anonymous access still
does not expose an unauthenticated agent.

## Controls that are decided and controls that are deferred

| Design 6.1 control | MVP status |
|---|---|
| Origin and CSWSH | Decided, implemented in the MVP transport lane |
| Bearer credential instead of ambient cookie | Decided |
| Per-message authorization | Decided here (revalidation inputs) and in ADR MVP-3 (rules) |
| Token replay and expiry | Decided: fixed lifetime, revocation, `gen` binding, accepted bearer replay |
| Credential carriage per client type | Decided |
| GitHub identity token validation (MVP-D3 option b) | Deferred: post-MVP upgrade, needs an identity provider integration the MVP does not otherwise need |
| Single-use upgrade tickets or channel binding | Deferred: needs a ticket-minting endpoint and client flow; bearer replay accepted for now |
| Automatic credential rotation or refresh | Deferred: renewal is operator re-issue through the TUI |
| Per-IP or per-credential upgrade rate limiting | Deferred: only the global connection cap and handshake deadline apply before authentication; post-authentication limits belong to plan 5.5 and 5.7 |
| Credentials in the subprotocol hidden from edge logs | Deferred: accepted residual risk above |

## Consequences

* Plan tasks for credential minting, the upgrade handler, config `[remote]` validation, and
  the TUI reveal and re-issue action implement exactly this ADR.
* Config `[remote]` fields: `token_lifetime_seconds`, `allowed_origins`, `allowed_hosts`,
  `max_connections`, and `remote_approve_destructive` (ADR MVP-3).
* The operator runbook must state `--allow-anonymous false` and warn that the credential is
  a bearer secret.
* Dev Tunnels may use its own edge-authentication header. The runbook and the end-to-end
  check must verify that the `Authorization` header reaches the backend unmodified through
  the tunnel; if it does not, native clients need another carriage and this ADR is amended.

## References

* Design sections 3.1, 5.2, 5.4, 5.5, 6, 6.1, 8
* ADR MVP-1: `docs/decisions/2026-10-19-intercom-go-mvp-adr-1-scope-and-session.md`
* ADR MVP-3: `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`
* Deliberation: `docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md`
