---
title: "067-S live verification evidence (real Win32 console, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render and tool-approval-routing manual checkpoints for the 067-S / 074-F closure: the real intercom.exe built from main (0fd7936, containing 067-S merge a2cacf3) ran in a real Windows console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients. Credentials are redacted. No token text is recorded."
shipment: 067-S
feature: 074-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing`
conditions of `docs/closure/067-S-074-F-post-merge-closure.md`. It comes from
live runs, not CI. It uses the same method as
`docs/closure/evidence/065-S-tui-render-live-evidence.md`, extended with live
WebSocket clients and a live model turn.

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` from `main` at
  `0fd793639319` (merge of PR 203), which contains the 067-S merge commit
  `a2cacf32c29b` (PR 202).
* Console: a temporary, untracked Go verifier called `FreeConsole` then
  `AllocConsole` (a real conhost with a 120 by 9001 buffer and a 120 by 30
  window), opened `CONIN$` and `CONOUT$`, and gave them to `intercom.exe` as
  stdin and stdout so Bubble Tea ran its real Windows console loop. Stderr went
  to a file. Three sentinel lines (`SCROLLBACK-SENTINEL-1..3`) were written to the
  main buffer before launch.
* Screen reads: `GetConsoleScreenBufferInfo` and `ReadConsoleOutputCharacterW`
  on a freshly opened `CONOUT$` per read. Key input: real `KEY_EVENT_RECORD`
  down and up pairs through `WriteConsoleInputW`.
* Remote surfaces: the in-repo reference client `internal/transport/wsclient`
  connected as `approver` and as `observer` over `ws://127.0.0.1:<port>/ws`.
  The bearer tokens were read from the rendered credential panel in memory only.
* Agent: the live Copilot CLI (authenticated) through the production adapter.
  Prompt: read `notes.txt` (a file inside the temporary workspace) and reply
  `DONE`. The verifier approved only a `read` of that exact file and denied any
  other request.
* Redaction: every captured string passed through an exact-match redactor for
  both tokens plus a generic 32-or-more-character token pattern before it was
  saved. Raw tokens and screens were never written to disk, and the verifier
  and its temporary directory were deleted afterwards. Config: a copy of
  `config.toml.example` with the workspace, database path and Copilot startup
  timeout (120 s) adjusted, and a free loopback port per run.
* Runs: three full runs on the same binary. Run 2 (approve paths, clean) is
  the primary record below. Run 3 repeated the scenario with both permission
  requests denied. Run 1 is reported under Observed anomaly.

## Step A: credential panel rendered (run 2)

```text
intercom | idle
Credentials (shown once - copy them now)
Endpoint: ws://127.0.0.1:53643/ws

approver: <REDACTED-TOKEN>
  fingerprint 377df69c, expires 2026-10-29T04:38:33Z

observer: <REDACTED-TOKEN>
  fingerprint 7c51c5fc, expires 2026-10-29T04:38:33Z

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

* The approver and observer upgrades both succeeded (HTTP 101 implied by a
  successful dial) and each received a `session.hydrate` envelope (seq 0).
* One unauthenticated `GET /ws` returned `HTTP 401`. The TUI status row became
  `rejected upgrades: 1 (last: 401 credential_rejected)`.

## tool-approval-routing

Round 1 (answered from the local Win32 console TUI, run 2): the remote approver
sent the prompt (`action.accepted`). The hub published `permission.requested`
(`kind=read`, summary the absolute path of `notes.txt` in the temp workspace) to
**both** remote clients, and the TUI rendered the modal in `CONOUT$`:

```text
╭──────────────────────────────────────────────────────╮
│ Permission request                                   │
│ expires: 01:38:36Z                                   │
│ kind: read                                           │
│ tool call: toolu_...                                 │
│                                                      │
│ <temp>\run\ws\notes.txt                              │
│ ...                                                  │
│ y approve (then Enter to confirm)   n deny           │
│ lines 1-4 of 4                                       │
╰──────────────────────────────────────────────────────╯
```

The verifier injected `y`; the screen showed `APPROVE? press Enter to confirm`;
it injected `Enter`. Both remote clients received `permission.resolved` with
`decision=allow_once`, `actor=local-tui` and the same request id; the modal
cleared and the notice row read `permission allow_once by local-tui`. A late
`resolve_permission` for the same request from the approver client was refused
(`action.refused`, `request_unknown`), and the turn completed to `agent.idle`.

Round 2 (answered from the remote approver, run 2): a second prompt produced a
new `permission.requested` for the same read. The TUI showed the modal. An
observer `resolve_permission` attempt was refused (`action.refused`,
`role_not_permitted`). The approver's `resolve_permission` was accepted; both
clients received `permission.resolved` with `decision=allow_once`,
`actor=remote-approver`; the TUI modal cleared by itself and the notice row
read `permission allow_once by remote-approver`. The turn completed to
`agent.idle`.

Run 3 (both requests denied) gave the same routing: `decision=deny` by
`local-tui` (round 1) and by `remote-approver` (round 2), the TUI modal cleared
each time, the late duplicate was refused with `request_unknown`, and shutdown
was clean.

## Step D: Ctrl+C and shutdown (runs 2 and 3)

Key: Ctrl+C as real key events (Control down, `C` down with character `0x03` and
`LEFT_CTRL_PRESSED`, `C` up, Control up).

| Measure | Run 2 | Run 3 |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 709 ms | 1061 ms |
| stderr `shutdown started` reason | `tui quit` | `tui quit` |
| stderr `shutdown complete` | `"ok":true` | `"ok":true` |
| Main-buffer sentinel rows 0 to 2 | unchanged | unchanged |
| Buffer, cursor, window after exit | 120 x 9001, (0, 3), (0, 0, 119, 29), same as before | same |
| TUI debris in main buffer | none | none |
| Token text in stderr | none | none |
| New `copilot` processes after exit | 0 | 0 |
| `intercom.exe` left running | 0 | 0 |

While running, `intercom.exe` had one child, `copilot.exe`, and it was gone
after exit. Stderr also contained the existing warnings
`config has accepted-but-unconsumed keys` (names only) and
`agent event arrived after quiesce; discarded` for a late `session.shutdown`;
neither affected the exit code.

## Dev Tunnel connect (devtunnel-connect)

All runs in this record took place between 2026-10-29T00:35Z and 00:50Z
(2026-10-28 in local time, UTC-7), after the merge of PR 203.

A separate run used the same binary and console method behind a live Dev Tunnel
(`devtunnel host -p <port> -a`, CLI 1.0.2014, logged in). Anonymous tunnel
access was used only so the reference client needs no tunnel token; intercom's
own bearer authentication stayed enforced, and the tunnel host was listed in
`[remote].allowed_hosts`. The tunnel was temporary and was stopped afterwards.

* Registration: the tunnel registered and printed a public URL
  (`https://<id>-<port>.usw3.devtunnels.ms`).
* The approver and the observer authenticated with their bearer tokens over
  `wss://<tunnel-host>/ws` and each received `session.hydrate` as the first frame.
* A prompt sent by the approver through the tunnel returned `action.accepted`
  and the turn reached `agent.idle`.
* A wrong credential through the tunnel was refused with HTTP 401.
* Ctrl+C exited with code 0 and stderr logged `shutdown complete` with
  `"ok":true`.

This covers Dev Tunnel registration, authenticated remote connect and initial
hydration. It does not exercise the `Hub.rehydrate` stale-client buffer drain
that 067-S changed: no live stale-buffer scenario was run, and that behavior is
covered by `TestRehydrateDrainsFullStaleRemoteBufferInsteadOfDropping` in the
race-enabled suite.

## Observed anomaly: one intermittent failed shutdown (run 1)

Run 1 (earlier, same binary and script, with both requests denied because of a
path-comparison bug in the verifier that treated the 8.3 temp path as outside
the workspace) completed every routing assertion identically, but the Ctrl+C
exit was not clean: `intercom.exe` exited with code 1 after 11.7 s, stderr logged
`shutdown step failed` for `2+4 stop agent` with
`agent did not confirm it stopped; the CLI child may be orphaned: agent: stop
copilot client: timed out; agent cleanup incomplete: ForceStop did not return
within 5s`, then `shutdown complete` with `ok:false`. The process still exited
(bounded, no hang), the scrollback sentinels were intact, no `copilot` process
was left, and no `intercom.exe` remained. In that run the model had retried
after the denial with a long stream of reasoning events before the final turn
completed. Two further runs on the same binary, one of them the same
all-denied scenario (run 3), shut down cleanly, and
`INTERCOM_LIVE_SDK_TESTS=1 go test -v ./internal/app -run TestLiveCopilotEndToEnd`
also passed (8.9 s, no stranded child). The failure was therefore not
reproduced and is recorded as residual risk with a follow-up stash entry (see
the closure artifact), not as a blocker: the 067-S bounded-stop contract held
(the process exited within its budget and reported the failure through its exit
code and log rather than hanging).

## Assertion summary

| Assertion | Result |
|---|---|
| tui-render: credential panel rendered with both roles and the endpoint | pass |
| tui-render: Esc dismissed the panel and the main view rendered | pass |
| tui-render: status line showed `rejected upgrades: 1` after one 401 | pass |
| tui-render: Ctrl+C quit with exit code 0 through the TUI path, scrollback intact | pass (runs 2 and 3); run 1 exited 1 (see anomaly) |
| ws-upgrade: two authenticated clients upgraded and hydrated; unauthenticated got 401 | pass |
| tool-approval-routing: `permission.requested` delivered to the TUI and both remote clients | pass |
| tool-approval-routing: TUI-first resolution observed on both remote surfaces | pass |
| tool-approval-routing: remote-first resolution cleared the TUI modal; observer refused | pass |
| No stranded `copilot` child | pass (all three runs) |
| No token text in this record | pass |

## Limits

* Windows conhost only. Windows Terminal, macOS and Linux were not exercised.
* The remote surface was the Go reference client (loopback for the tui-render
  and approval runs, a live Dev Tunnel for the connect run), not a browser.
* The bounded-stop wedge case remains covered by the race-enabled tests named in
  the closure artifact, not by these runs.
* A simultaneous two-surface answer race (both surfaces resolving in the same
  instant) was not forced live; the late-duplicate refusal and observer refusal
  were.
