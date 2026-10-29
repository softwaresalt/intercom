---
title: "069-S live verification evidence (real Win32 console, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render, devtunnel-connect and tool-approval-routing manual checkpoints for the 069-S / 076-F closure: the real intercom.exe built from the merged 069-S code (main 5f561d2) ran in a real Windows console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients, and through a live Dev Tunnel. Credentials are redacted. No token text is recorded."
shipment: 069-S
feature: 076-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing` conditions of the 069-S / 076-F post-merge closure. It
comes from live runs, not CI, using the same method as
`docs/closure/evidence/068-S-live-verification-evidence.md`. Shipment 069-S
changed the local TUI: resilient reconnect, summary sanitization, a 2-space
summary indent and a 150 ms approval lock-out in the permission modal, and
viewport hardening with mouse reporting (`tea.EnableMouseCellMotion`). Those
changes drive the extra checks below.

All runs took place on 2026-10-29 between 04:09Z and 04:11Z (2026-10-28 21:09
to 21:11 in local time, UTC-7).

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` on the
  post-merge closure branch, whose Go source (every `*.go` file, `go.mod`,
  `go.sum`) is identical to `main` at `5f561d2c6f25` (the 069-S merge, PR 207);
  the closure branch differs from `main` only by backlog and docs files.
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
* Agent: the live Copilot CLI (`C:\Tools\copilot.exe`, authenticated) through
  the production adapter, with the Copilot startup timeout set to 120 s.
  Prompt: read `notes.txt` (a file inside the temporary workspace) and reply
  `DONE`. The verifier approved only a `read` whose summary named that file.
* Redaction: every captured string passed through an exact-match redactor for
  both tokens plus a generic 32-or-more-character token pattern before it was
  written. Raw tokens and screens were never written to disk. The verifier, its
  workspace, config copy, database and logs, and the temporary `intercom.exe`
  were deleted afterwards.
* Config: a copy of `config.toml.example` with workspace, database path and
  Copilot startup timeout adjusted and a free loopback port per run.
* Runs: a first loopback run found a verifier bug (the permission kind was read
  from the wrong JSON key, so the verifier denied the request instead of
  approving it); it is not used except for the console-mode observation below.
  Run 2 (loopback) is the primary record, and run 3 is the Dev Tunnel run.

## Step A: credential panel rendered (run 2)

```text
intercom | idle
Credentials (shown once - copy them now)
Endpoint: ws://127.0.0.1:60821/ws

approver: <REDACTED-TOKEN>
  fingerprint 90dffb79, expires 2026-10-29T05:10:21Z

observer: <REDACTED-TOKEN>
  fingerprint 7d43eb4a, expires 2026-10-29T05:10:21Z

Esc dismiss   Ctrl+R re-issue (revokes these)
...
credentials are shown once: copy them now (Esc dismisses)
> _
```

The panel was readable from `CONOUT$` 1.9 s after launch (run 2).

## Step B: Esc dismiss and mouse reporting in a real console

`Esc` (`0x1B`) was injected through `WriteConsoleInputW`. The panel and its
dismiss hint were gone and the main view rendered (`intercom | idle`,
`No activity yet...`).

Mouse reporting: `intercom.exe` was started with the console handles as
stdin and stdout, not through `tea.WithInput`. In that setup `app.Options.Input`
is nil, so `NoMouse` is false and `tea.EnableMouseCellMotion` is in effect;
the verifier setup does not turn it off. Observed behavior:

* The console input mode read through `GetConsoleMode(CONIN$)` while intercom
  ran was `0x98` (`ENABLE_MOUSE_INPUT` set, `ENABLE_WINDOW_INPUT`,
  `ENABLE_EXTENDED_FLAGS`; line input, echo and processed input off) in run 2
  and again in run 3 just before the Esc key.
* One earlier sample taken within about a second of the first render read
  `0x88` (mouse flag not yet set), in run 1 and in run 3. The mode had
  changed to `0x98` by the later sample in the same run, so Bubble Tea sets the
  flag shortly after startup rather than never.
* With mouse input enabled, no key was swallowed: Esc, `y`, `Enter` and Ctrl+C
  injected as real key events were all acted on (the panel dismissed, the modal
  armed and approved, and the program quit). No mouse event was injected, so
  mouse event handling itself is not covered here.

## Step C: WebSocket upgrade and rejected-upgrade status line

* Approver and observer both authenticated and each received `session.hydrate`
  as its first frame.
* One connection with a wrong credential got `HTTP 401`
  (`transport: upgrade rejected`, `status 401`, `credential_rejected` in the
  server log). The TUI status row became
  `rejected upgrades: 1 (last: 401 credential_rejected)`.

## tool-approval-routing

Round 1 (answered from the local Win32 console TUI): the remote approver sent
the prompt (`action.accepted`). The hub published `permission.requested`
(`perm-1`, kind `read`, summary the absolute path of `notes.txt` in the temp
workspace) to **both** remote clients. The verifier polled `CONOUT$`
continuously and injected `y` the instant the text `Permission request`
appeared. The screen read back 1.6 ms after detection showed the lock-out
hint, not the arming prompt:

```text
╭──────────────────────────────────────────────────────╮
│ Permission request                                   │
│ expires: 05:10:24Z                                   │
│ kind: read                                           │
│ tool call: toolu_...                                 │
│                                                      │
│   <temp>\run2\ws\notes.txt                           │
│                                                      │
│ approval unlocks in a moment   n deny                │
│ lines 1-4 of 4                                       │
╰──────────────────────────────────────────────────────╯
```

New behaviors verified:

* Summary indent: the first summary row started 2 columns to the right of the
  `kind:` row (columns 4 and 2 of the console row), so no summary line is at
  the left edge of the modal body.
* 150 ms approval delay: the `y` sent at the moment the modal appeared did not
  arm approval (no `APPROVE?` prompt). At 400 ms after the modal appeared the
  status row read `wait a moment before approving` (the refusal notice from the
  early `y`) and the hint had changed to
  `y approve (then Enter to confirm)   n deny`. The early `y` was therefore
  processed and refused during the lock-out. The 1.6 ms read happened before the
  refusal notice rendered; the notice and the later hint are the stronger
  evidence. The verifier cannot time the modal's first paint more precisely than
  its polling loop, so "within 150 ms" means "within a few milliseconds of first
  detection".
* After the delay, `y` showed `APPROVE? press Enter to confirm, any other key
  cancels`, and `Enter` confirmed.

Both remote clients received `permission.resolved` with `decision=allow_once`,
`actor=local-tui` and request id `perm-1`. A late `resolve_permission` for the
same request from the approver was refused (`action.refused`,
`request_unknown`), and the turn completed to `agent.idle` on both.

Round 2 (answered from the remote approver): a second prompt produced
`permission.requested` (`perm-2`, kind `read`) and the TUI showed the modal. An
observer `resolve_permission` attempt was refused (`action.refused`,
`role_not_permitted`). The approver's `resolve_permission` was accepted; both
clients received `permission.resolved` with `decision=allow_once`,
`actor=remote-approver`. The TUI modal cleared by itself (no
`Permission request` text 600 ms later) and the notice row read
`permission allow_once by remote-approver`. The turn completed to `agent.idle`.

## Step D: Ctrl+C and shutdown (run 2)

Key: Ctrl+C as real key events (Control down, `C` down with character `0x03`
and `LEFT_CTRL_PRESSED`, `C` up, Control up).

| Measure | Run 2 (loopback) | Run 3 (Dev Tunnel) |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 479 ms | 654 ms |
| stderr `shutdown started` reason | `tui quit` | `tui quit` |
| stderr `shutdown complete` | `"ok":true` | `"ok":true` |
| Remote frames at close | `server.shutdown`, then close 1001 `closing` (both clients) | same |
| Main-buffer sentinel rows 0 to 2 | unchanged | unchanged |
| Buffer, cursor, window after exit | 120 x 9001, (0, 3), (0, 0, 119, 29), same as before launch | same |
| Token text in stderr | none | none |

Stderr also held the existing warnings `config has accepted-but-unconsumed
keys` (names only) and `agent event arrived after quiesce; discarded` for a
late `session.shutdown`; neither affected the exit code. 

## Dev Tunnel connect (devtunnel-connect)

A temporary tunnel was created with `devtunnel create --allow-anonymous`
(CLI 1.0.2014, logged in), given one port, and hosted with `devtunnel host`.
Anonymous tunnel access was used only so the reference client needs no tunnel
token; intercom's own bearer authentication stayed enforced, and the tunnel
host `<id>-<port>.usw3.devtunnels.ms` was listed in `[remote].allowed_hosts`.

* The approver and observer authenticated over `wss://<tunnel-host>/ws` and each
  received `session.hydrate` as the first frame.
* A wrong credential through the tunnel got `HTTP 401` (server log:
  `transport: upgrade rejected`, `status 401`, `credential_rejected`).
* A prompt sent by the approver through the tunnel round-tripped and
  `agent.idle` reached both the approver and the observer.
* Ctrl+C exited `intercom.exe` with code 0 (`shutdown complete`, `"ok":true`);
  both tunnel clients received `server.shutdown` then close 1001 `closing`.
* Afterwards the `devtunnel host` process was stopped and the tunnel was
  deleted with `devtunnel delete`; an older, unrelated tunnel in the account was
  left alone.

## Assertion summary

| Assertion | Result |
|---|---|
| tui-render: credential panel rendered with both roles and the endpoint | pass |
| tui-render: Esc dismissed the panel with mouse reporting active in the real console | pass |
| tui-render: status line showed `rejected upgrades: 1` after one 401 | pass |
| tui-render: Ctrl+C quit with exit code 0, scrollback and console state intact | pass |
| tool-approval-routing: `permission.requested` delivered to the TUI and both remote clients | pass |
| tool-approval-routing: summary rendered with a 2-space indent | pass |
| tool-approval-routing: `y` at modal appearance refused; hint changes after the 150 ms delay | pass |
| tool-approval-routing: `y` then `Enter` approved; `allow_once` by `local-tui` on both remotes | pass |
| tool-approval-routing: remote approver resolved, TUI modal cleared; observer refused | pass |
| devtunnel-connect: wss connect, hydrate, prompt round trip, 401, clean shutdown | pass |
| No stray `intercom`, `devtunnel` or verifier process; tunnel deleted | pass |
| No token text in this record | pass |

## Limits

* Windows conhost only. Windows Terminal, macOS and Linux were not exercised.
* The remote surface was the Go reference client (loopback for the tui-render
  and approval runs, a live Dev Tunnel for the connect run), not a browser.
* No mouse event was injected, so mouse handling in the viewport is not covered
  live; only that enabling mouse reporting does not break key input.
* The 150 ms lock-out is verified by outcome (early `y` refused, later `y`
  accepted), not by a precise first-paint timestamp.
* The reconnect and sanitization paths of 069-S were not driven live; they
  remain covered by the race-enabled tests.
* Copilot processes owned by other sessions were already running on the
  machine before these runs and were not touched.
