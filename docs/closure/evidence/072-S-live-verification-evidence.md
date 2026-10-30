---
title: "072-S live verification evidence (real Windows pseudoconsole, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render, devtunnel-connect and tool-approval-routing manual checkpoints for the 072-S / 079-F closure: the real intercom.exe built from the merged 072-S code (main d0d863f) ran in a Windows ConPTY console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients, on loopback and through a live Dev Tunnel. Credentials are redacted. No token text is recorded."
shipment: 072-S
feature: 079-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing` conditions of the 072-S / 079-F post-merge closure. It
comes from live runs, not CI. Shipment 072-S changed `agent.Event` bounds
(`boundUTF8`, `Truncated`), the agent adapter (write-diff and shell `possibleUrls`
surfacing in `PermissionRequest.Summary`, reasoning events, the unmapped-event log
gate), the hub fold and publish path (`agent.reasoning`, `agent.reasoning.delta`,
`tool.ended` text, rate-limited unmapped warns) and the TUI rendering of
reasoning and tool text. The checks below confirm that the modal and credential
panel still render on the real console path, that approval routing and shutdown
still work against the real SDK with the new adapter and hub code, and that live
event delivery to two remote clients is undisturbed.

All runs took place on 2026-10-30 at about 01:02Z to 01:04Z (2026-10-29 18:02 to
18:04 local, UTC-7).

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` on the post-merge
  closure branch, whose Go source is identical to `main` at `d0d863fae5ec`
  (the 072-S merge, PR 214).
* Console: the binary ran under a Windows ConPTY pseudoconsole (120 by 30), so
  Bubble Tea used its real Windows console path; key input went through the
  pseudoconsole input pipe and the screen was read from its output stream (VT
  sequences stripped). This is a pseudoconsole, not a dedicated conhost window.
* Remote surfaces: the in-repo reference client `internal/transport/wsclient`
  connected as `approver` and as `observer`. The bearer tokens were read from
  the rendered credential panel in memory only.
* Agent: the live Copilot CLI (`C:\Tools\copilot.exe`, authenticated) through
  the production adapter. Prompts asked the agent to read `notes.txt` (a file in
  a temporary workspace) with its file-read tool and reply `DONE`.
* Redaction: every logged string passed through an exact-match redactor for
  both tokens. Raw tokens were never written to this record. The temporary
  verifier (a throwaway program that sets `STARTF_USESTDHANDLES` so the child
  sees the pseudoconsole), its workspace, config, database, logs and temporary
  `intercom.exe` were deleted afterwards without being committed.
* Runs: run A (loopback, port 38470) and run B (Dev Tunnel, port 38471) executed
  the same script: one intercom process, three prompt rounds, then Ctrl+C. Run A
  also performed the scrollback check. Run A: 21 assertions, 0 failed. Run B: 17
  assertions, 0 failed.

## tui-render

* The credential panel (`Credentials (shown once - copy them now)`, the
  `Endpoint: ws://127.0.0.1:<port>/ws` line, and the `approver:` and `observer:`
  token lines) rendered and was readable from the pseudoconsole in both runs.
* The `Permission request` modal drew with its `kind:` and `tool call:` header
  lines in every round that the local TUI was asked (rounds 1 to 3, both runs).
* Local key input was accepted: `y`, `Enter` and Ctrl+C. The TUI quit with exit
  code 0.

### Scrollback preservation

A separate ConPTY run (120 by 30, run A script, same merged code) wrote three
sentinel lines (`SCROLLBACK-SENTINEL-1`, `-2`, `-3`, each followed by `(written
before launching intercom.exe)`) to the main screen buffer through `cmd.exe` and
then launched `intercom.exe`, let it run 12 s, and sent Ctrl+C. The raw
pseudoconsole output stream was inspected:

* The three sentinel lines were emitted before the TUI's alternate-screen enter
  (`ESC[?1049h`), so they sit in the main buffer.
* The TUI left the alternate screen on exit (`ESC[?1049l`), restoring the main
  buffer; `intercom.exe` exited with code 0.
* No scrollback-clear (`ESC[3J`) and no main-buffer clear (`ESC[2J`) was emitted
  after leaving the alternate screen.

This is a stream-level check of ConPTY output, not a screenshot of the visible
rows; it matches the design (`tea.WithAltScreen()` in `internal/app/ui.go`) and
the earlier 065-S and 071-S evidence. It is ConPTY only.

## tool-approval-routing

Round 1 (answered from the local TUI), identical in both runs: the approver sent
a prompt, the live agent asked to read `notes.txt`, and the hub published
`permission.requested` (`perm-1`) to **both** remote clients while the TUI showed
the `Permission request` modal.

* A `y` sent right after the modal appeared did not arm approval (no `APPROVE?`
  prompt appeared), so the early key was refused during the 150 ms lock-out.
* A second `y` (after about 1 s) showed `APPROVE? press Enter to confirm`;
  `Enter` confirmed. Both remote clients received `permission.resolved` with
  `decision=allow_once`, `actor=local-tui`, `request_id=perm-1`, and the turn
  completed to `agent.idle` on both.

Round 2 (answered from the remote approver): a second prompt produced
`permission.requested` (`perm-2`) on both clients. The observer's
`resolve_permission` was refused (`action.refused`, `reason=role_not_permitted`).
The approver's `resolve_permission` was accepted and both clients received
`permission.resolved` (`allow_once`, `actor=remote-approver`). The turn completed
to `agent.idle` on both. The new `tool.started` and `tool.ended` events were
delivered to both clients in between.

## Shutdown with a parked permission ask

Round 3 sent a third prompt and waited until the agent had a `permission.requested`
ask parked mid-turn (`perm-3`), then sent Ctrl+C to the TUI.

| Measure | Run A (loopback) | Run B (Dev Tunnel) |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 0.898 s | 2.677 s |
| `server.shutdown` received by approver | yes | yes |
| `server.shutdown` received by observer | yes | yes |
| Parked ask released (`permission.resolved`, `decision=unavailable`, `actor=hub`), run A | yes | not asserted |
| Token text in the record | none | none |

## Dev Tunnel connect (devtunnel-connect)

A temporary tunnel was created with `devtunnel create --allow-anonymous` (CLI
1.0.2014, logged in), given one port (38471) and hosted with `devtunnel host`.
Anonymous tunnel access was used only so the reference client needs no tunnel
token; intercom's own bearer authentication stayed enforced and the tunnel host
(`<id>-38471.usw3.devtunnels.ms`) was listed in `[remote].allowed_hosts`.

* The approver and observer authenticated over `wss://<tunnel-host>/ws` and each
  received `session.hydrate`.
* A wrong credential through the tunnel got `HTTP 401`.
* Prompt rounds 1 to 3 ran through the tunnel with the same results as run A.
* Ctrl+C exited `intercom.exe` with code 0; both tunnel clients received
  `server.shutdown`.
* Afterwards the `devtunnel host` process was stopped and the tunnel was deleted
  with `devtunnel delete` (output `Deleted`). No `intercom` or `devtunnel` process
  remained.

## Assertion summary

| Assertion | Result |
|---|---|
| tui-render: credential panel rendered with both roles and the endpoint | pass |
| tui-render: `Permission request` modal with `kind:` and `tool call:` lines | pass |
| tui-render: scrollback preserved (main-buffer sentinels before launch, alt-screen left on exit, no clear sequences) | pass |
| tui-render: local key input accepted (`y`, `Enter`, Ctrl+C) and the TUI quit with exit code 0 | pass |
| tool-approval-routing: `permission.requested` delivered to the TUI and both remote clients | pass |
| tool-approval-routing: `y` inside the lock-out refused; later `y` then `Enter` approved (`allow_once`, `local-tui`) | pass |
| tool-approval-routing: remote approver resolved, observer refused (`role_not_permitted`) | pass |
| shutdown with a parked ask: exit 0, `server.shutdown` delivered to both remote clients | pass |
| devtunnel-connect: wss connect, hydrate, prompt rounds, 401, clean shutdown | pass |
| Tunnel deleted; no stray processes | pass |
| No token text in this record | pass |

## Limits

* Windows ConPTY only. Windows Terminal, a dedicated conhost window, macOS and
  Linux were not exercised. Scrollback was checked at stream level only.
* The remote surface was the Go reference client, not a browser.
* The prompts used a read permission. The write-diff and shell `possibleUrls`
  surfacing (079.001-T) and live reasoning deltas (079.003-T to 079.005-T) depend
  on the agent choosing a write or shell tool or emitting reasoning; those paths
  are covered by the `agent`, `hub` and `internal/tui` unit tests, and the live
  runs only confirm the event path stayed intact for read, tool and turn events.
  The event kinds logged by the live runs were not individually asserted beyond
  the permission, tool, turn and idle kinds above.
* After the parked ask was released at shutdown the reference clients also saw a
  few `error` envelopes and a closing `agent.turn.end` and `agent.idle`; the
  clients were then closed by the server and the process exited 0. This was
  not asserted either way.
* The stderr JSON log lines were not captured separately from the screen; the
  exit code and the `server.shutdown` frames are the recorded shutdown evidence.
* Copilot processes owned by other sessions were already running on the machine
  and were not touched.
