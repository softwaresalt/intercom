---
title: "062-S live verification evidence (operator-present session, 2026-10-20)"
description: "Summarized results of the live SDK tests and the Dev Tunnel connect check cited by the three operator-owned condition dispositions of the 062-S closure. Evidence is from live runs in an operator-present session, not from CI. No credentials or tokens are recorded."
shipment: 062-S
feature: 070-F
date: 2026-10-20
---

# 062-S live verification evidence

These results come from live runs in an operator-present session on 2026-10-20.
They are not CI results and are not reproducible from the repository alone: they
need an authenticated Copilot CLI and, for the tunnel check, a Dev Tunnels
account. No credentials or tokens appear here.

## A. Live SDK tests

Environment: Copilot CLI 1.0.100, `GH_TOKEN` authentication, Windows, go1.26.9.

`INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/app -run Live -v -count=1` passed.

* `TestLiveCopilotEndToEnd`: the real CLI was started through `app.CopilotAgent`
  with the real hub, WebSocket server and reference client. The prompt raised a
  `read` permission request that the reference client answered. Shutdown was
  clean. The CLI process tree was seen after `Start`, and no CLI child remained
  after shutdown.
* `TestLiveReadApprovalIsConfinedToWorkspace` and the other `Live` tests in the
  package passed.

`INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/agent/copilot -run Live -v -count=1`:

* `TestLiveStartAndStop`: pass.
* `TestLiveSQbSQcReverification`: failed as inconclusive.
  * SQ-b: **verified**. The maximum number of in-flight callbacks was 1 across
    106 events, with one permission handler entry.
  * SQ-c: **not re-verified**. No session event was dispatched in the 30 seconds
    that the permission handler was parked, although 60 events had been
    dispatched before the request. The runtime was probably waiting for the
    answer, but that cannot be distinguished from head-of-line blocking. A prompt
    or event source that emits events while the request is pending is needed.

Design section 4.4 SQ-c remains an unverified assumption, and the conservative
posture stays in force.

## B. Dev Tunnel live check

Environment: devtunnel CLI 1.0.2014, one user account, one machine.

Setup:

* `devtunnel create intercom-test --allow-anonymous false` was accepted; the
  access control was `{}` with no anonymous entry. The runbook's note that this
  syntax was unverified is resolved for this build.
* `devtunnel port create intercom-test -p 39173 --protocol http` and
  `devtunnel host intercom-test` produced the host
  `ht9t8j1m-39173.usw3.devtunnels.ms` (pattern `<id>-<port>.<cluster>.devtunnels.ms`).
* The backend was a temporary, uncommitted harness wiring the real Copilot agent,
  hub and transport, with `[remote].allowed_hosts` set to the bare tunnel
  hostname (no port; this worked). The real binary shows credentials only on its
  interactive TUI, which could not be driven non-interactively.
* The client was the reference client (`internal/transport/wsclient`) over
  `wss://<host>/ws` from the same machine through the public edge, with a Dev
  Tunnels access token.

Results:

1. A connection without the Dev Tunnels edge token was refused at the edge (an
   HTTP 200 sign-in page, no WebSocket upgrade), so the edge enforced access control on this tunnel.
2. With the edge token and a wrong intercom credential, intercom answered HTTP
   401.
3. An observer connected with the credential in the `Authorization` header.
4. An approver connected with the credential in the `Authorization` header. The
   header reached intercom unmodified through the tunnel, which satisfies the
   ADR MVP-2 manual per-environment check for this environment.
5. Browser-style subprotocol carriage (`intercom.v1`, `bearer.<credential>`)
   connected through the tunnel. Only the reference client was used; no real
   browser was tested.
6. A full turn completed: the prompt was accepted, a `read` permission was
   requested, a remote approver approved it, `permission.resolved` and
   `agent.idle` were observed, and shutdown completed with no error.

The tunnel and the temporary harness were deleted afterwards.

Operational finding: a tunnel client must present a Dev Tunnels access token in
the header `X-Tunnel-Authorization: tunnel <token>`, obtained with
`devtunnel token <tunnel-id> --scopes connect` (about 24 hours of lifetime). This
is separate from the intercom bearer credential in `Authorization`.

## Not tested

* A real browser.
* A second machine.
* The real binary, launched from its TUI, behind the tunnel.
* `Origin` handling and `allowed_origins` through the tunnel.
* A long-lived connection with reconnect across the tunnel.
* Permission kinds other than `read` (shell and write) routed live.
* Design section 4.4 SQ-c (see A).
