---
title: "068-S live verification evidence (real Win32 console, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render, devtunnel-connect and tool-approval-routing manual checkpoints for the 068-S / 075-F closure: the real intercom.exe built from the merged 068-S code (main 2fff94b) ran in a real Windows console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients, and through a live Dev Tunnel. Credentials are redacted. No token text is recorded."
shipment: 068-S
feature: 075-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing` conditions of
`docs/closure/068-S-075-F-post-merge-closure.md`. It comes from live runs, not
CI, using the same method as `docs/closure/evidence/067-S-live-verification-evidence.md`.
Shipment 068-S changed shutdown broadcast and drain behavior (terminal
`server.shutdown` delivery, write-ack drain), added in-place hub
`Client.Rehydrate`, and made credential re-issue honor a cancelled context, so
the shutdown step is checked from the remote surfaces as well as the console.

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` on the
  post-merge closure branch, whose Go source is identical to `main` at
  `2fff94b6722f` (the 068-S merge, PR 205); the closure branch differs from
  `main` only by backlog and docs files.
* Console: a temporary, untracked Go verifier called `FreeConsole` then
  `AllocConsole` (a real conhost with a 120 by 9001 buffer and a 120 by 30
  window), opened `CONIN$` and `CONOUT$`, and gave them to `intercom.exe` as
  stdin and stdout so Bubble Tea ran its real Windows console loop. Stderr went
  to a file. Three sentinel lines (`SCROLLBACK-SENTINEL-1..3`) were written to
  the main buffer before launch.
* Screen reads: `GetConsoleScreenBufferInfo` and `ReadConsoleOutputCharacterW`
  on a freshly opened `CONOUT$` per read. Key input: real `KEY_EVENT_RECORD`
  down and up pairs through `WriteConsoleInputW`.
* Remote surfaces: the in-repo reference client `internal/transport/wsclient`
  connected as `approver` and as `observer`. The bearer tokens were read from
  the rendered credential panel in memory only.
* Agent: the live Copilot CLI (authenticated) through the production adapter.
  Prompt: read `notes.txt` (a file inside the temporary workspace) and reply
  `DONE`. The verifier approved only the `read` of that exact file.
* Redaction: every captured string passed through an exact-match redactor for
  both tokens plus a generic 32-or-more-character token pattern before it was
  saved. Raw tokens and screens were never written to disk, and the verifier,
  its temporary workspace and its transcripts were deleted afterwards.
* Runs: two full loopback runs on the same binary (run 2 is the primary record
  below; run 1 gave the same results) and one Dev Tunnel run.

## Step A: credential panel rendered (run 2)

```text
intercom | idle
Credentials (shown once - copy them now)
Endpoint: ws://127.0.0.1:59348/ws

approver: <REDACTED-TOKEN>
  fingerprint 89b16163, expires 2026-10-29T06:23:55Z

observer: <REDACTED-TOKEN>
  fingerprint 06830e61, expires 2026-10-29T06:23:55Z

Esc dismiss   Ctrl+R re-issue (revokes these)
...
credentials are shown once: copy them now (Esc dismisses)
> _
```

## Step B: Esc dismiss

`Esc` (`0x1B`) was injected. The panel and dismiss hint were gone and the main
view rendered:

```text
intercom | idle
No activity yet. Type a prompt and press Enter.
...
connected
> _
```

## Step C: WebSocket upgrade and rejected-upgrade status line

* The approver and observer upgrades both succeeded and each received a
  `session.hydrate` envelope.
* One unauthenticated `GET /ws` returned `HTTP 401`. The TUI status row became
  `rejected upgrades: 1 (last: 401 credential_rejected)`.

## tool-approval-routing

Round 1 (answered from the local Win32 console TUI): the remote approver sent
the prompt (`action.accepted`). The hub published `permission.requested`
(`kind=read`, summary the absolute path of `notes.txt` in the temp workspace) to
**both** remote clients, and the TUI rendered the modal in `CONOUT$`:

```text
╭──────────────────────────────────────────────────────╮
│ Permission request                                   │
│ expires: 03:23:58Z                                   │
│ kind: read                                           │
│ tool call: toolu_...                                 │
│                                                      │
│ <temp>\out\ws\notes.txt                              │
│ ...                                                  │
│ y approve (then Enter to confirm)   n deny           │
│ lines 1-4 of 4                                       │
╰──────────────────────────────────────────────────────╯
```

The verifier injected `y`; the screen showed `APPROVE?`; it injected `Enter`.
Both remote clients received `permission.resolved` with `decision=allow_once`,
`actor=local-tui` and the same request id (`perm-1`). A late
`resolve_permission` for the same request from the approver client was refused
(`action.refused`, `request_unknown`), and the turn completed to `agent.idle`.

Round 2 (answered from the remote approver): a second prompt produced a new
`permission.requested` (`perm-2`). The TUI showed the modal. An observer
`resolve_permission` attempt was refused (`action.refused`,
`role_not_permitted`). The approver's `resolve_permission` was accepted; both
clients received `permission.resolved` with `decision=allow_once`,
`actor=remote-approver`; the TUI modal cleared by itself and the notice row
read `permission allow_once by remote-approver`. The turn completed to
`agent.idle`.

## Step D: Ctrl+C and shutdown, as seen from the console and the remote clients

Key: Ctrl+C as real key events (Control down, `C` down with character `0x03` and
`LEFT_CTRL_PRESSED`, `C` up, Control up).

| Measure | Run 1 | Run 2 |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 516 ms | 670 ms |
| stderr `shutdown started` reason | `tui quit` | `tui quit` |
| stderr `shutdown complete` | `"ok":true` | `"ok":true` |
| Approver last frame before close | `server.shutdown` | `server.shutdown`, then close 1001 `closing` |
| Observer frames at close | ..., `agent.idle`, `server.shutdown`, close 1001 `closing` | same |
| Main-buffer sentinel rows 0 to 2 | unchanged | unchanged |
| Buffer, cursor, window after exit | 120 x 9001, (0, 3), (0, 0, 119, 29), same as before launch | same |

Both remote clients received the terminal `server.shutdown` frame before the
server closed their connection with status 1001, which is the behavior 075.001-T
and 075.003-T changed (terminal delivery and write-acknowledged drain). In run 1
the approver's close status was not captured because the verifier read the
close error before the reader goroutine had delivered it (a verifier read race,
not a server result); run 2 captured it. The stale-client terminal delivery and
the in-place `Client.Rehydrate` paths are not driven by this live scenario; they
are covered by `TestStaleClientReceivesServerShutdownBeforeClose` and
`TestClientRehydrateInPlaceKeepsPendingPermissionAlive` under `go test -race`.

## Dev Tunnel connect (devtunnel-connect)

The merged binary ran behind an existing Dev Tunnel with anonymous connect
access (`devtunnel host`, tunnel port 49321, host
`0hqc8hlt-49321.usw3.devtunnels.ms` listed in `[remote].allowed_hosts`;
intercom bearer authentication still enforced).

* Approver and observer authenticated over `wss://` through the tunnel and each
  received `session.hydrate`.
* A wrong credential through the tunnel got `HTTP 401` (server log:
  `transport: upgrade rejected`, `status 401`, `credential_rejected`).
* A prompt sent by the approver through the tunnel round-tripped
  (`action.accepted`, `agent.turn.start`, `agent.message.delta`,
  `agent.message`, `agent.turn.end`, `agent.idle`) and the observer also saw
  `agent.idle`.
* Ctrl+C exited `intercom.exe` with code 0 in 797 ms (`shutdown complete`,
  `"ok":true`); both tunnel clients received `server.shutdown` then close 1001
  `closing`.

## Cleanup

The `devtunnel host` process was stopped. The temporary verifier, its
workspace, database and transcripts, and the temporary `intercom.exe` were
deleted before commit. No token text appears in this file.
