---
title: "070-S live verification evidence (real Windows pseudoconsole, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render, devtunnel-connect and tool-approval-routing manual checkpoints for the 070-S / 077-F closure: the real intercom.exe built from the merged 070-S code (main af87a21) ran in a Windows ConPTY console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients, and through a live Dev Tunnel. Credentials are redacted. No token text is recorded."
shipment: 070-S
feature: 077-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing` conditions of the 070-S / 077-F post-merge closure. It
comes from live runs, not CI. Shipment 070-S changed the Copilot adapter
(permission sanitization and re-check, latch quiescing, `session.shutdown`
classification, late terminal idle delivery, ForceStop wait accounting) and
`App.Err` reporting; it did not change the TUI, the transport or tunnel
registration, so the checks below confirm that those surfaces still work with
the new adapter and that the shutdown path behaves on the real SDK.

All runs took place on 2026-10-29 at 22:13Z to 22:14Z (15:13 to 15:14 local,
UTC-7).

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` on the
  post-merge closure branch, whose Go source (every `*.go` file, `go.mod`,
  `go.sum`) is identical to `main` at `af87a21b342a` (the 070-S merge, PR 210).
* Console: the binary ran under a Windows ConPTY pseudoconsole (120 by 30) so
  Bubble Tea used its real Windows console path; key input went through the
  pseudoconsole input pipe and the screen was read from its output stream (VT
  sequences stripped). This is a pseudoconsole, not a dedicated conhost window.
* Remote surfaces: the in-repo reference client `internal/transport/wsclient`
  connected as `approver` and as `observer`. The bearer tokens were read from
  the rendered credential panel in memory only.
* Agent: the live Copilot CLI (`C:\Tools\copilot.exe`, authenticated) through
  the production adapter, Copilot startup timeout 120 s. Prompts asked the
  agent to read `notes.txt` (a file in the temporary workspace) and reply
  `DONE`.
* Redaction: every logged string passed through an exact-match redactor for
  both tokens. Raw tokens were never written to this record. The temporary
  verifier, its workspace, config, database, logs and the temporary
  `intercom.exe` were deleted afterwards.
* Runs: run 2 (loopback) is the primary record; run 3 is the Dev Tunnel run.

## Step A: credential panel rendered (run 2)

The credential panel (`Credentials (shown once - copy them now)`, the
`Endpoint: ws://127.0.0.1:<port>/ws` line, and `approver:` / `observer:` token
lines) was readable from the pseudoconsole 5.7 s after launch. The Dev Tunnel
run rendered the same panel after 5.9 s.

## Step C: WebSocket upgrade and rejected upgrade

* Approver and observer both authenticated and each received `session.hydrate`
  as its first frame.
* A connection with a wrong credential was refused with `HTTP 401`
  (`credential_rejected` in the server log).

## tool-approval-routing

Round 1 (answered from the local TUI): the approver sent a prompt, the live
agent asked to read `notes.txt`, and the hub published `permission.requested`
(`perm-1`, `permission_kind: read`, summary the absolute path of `notes.txt`,
a `toolu_...` tool call id) to **both** remote clients while the TUI showed the
`Permission request` modal.

* A `y` sent about 50 ms after the modal appeared did not arm approval (no
  `APPROVE?` prompt). After 500 ms the screen showed the lock-out notice
  `wait a moment before approving` and the hint `y approve`, so the early key
  was refused during the 150 ms lock-out.
* A second `y` showed `APPROVE? press Enter to confirm`; `Enter` confirmed.
  Both remote clients received `permission.resolved` with
  `decision=allow_once`, `actor=local-tui`, `request_id=perm-1`, and the turn
  completed to `agent.idle` on both.

Round 2 (answered from the remote approver): a second prompt produced
`permission.requested` (`perm-2`). The observer's `resolve_permission` was
refused (`action.refused`, `reason=role_not_permitted`). The approver's
`resolve_permission` was accepted and both clients received
`permission.resolved` (`allow_once`, `actor=remote-approver`). The TUI modal
cleared by itself and the turn completed to `agent.idle`.

## Step D: shutdown with a parked permission ask (run 2, 077-F path)

Round 3 sent a third prompt and waited until the agent had a `permission.requested` ask
parked mid-turn, then sent Ctrl+C to the TUI.

| Measure | Run 2 (loopback, ask parked) | Run 3 (Dev Tunnel, idle) |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 877 ms | 563 ms |
| stderr `shutdown started` reason | `tui quit` | `tui quit` |
| stderr `shutdown complete` | `"ok":true` | `"ok":true` |
| Remote frames at close | `server.shutdown`, then close 1001 `closing` (both clients) | same |
| Parked ask | resolved by the hub during shutdown (`permission.resolved` count 3 on both clients) | not applicable |
| Token text in stderr | none | none |

The parked ask was released rather than left hanging, `App.Err` stayed nil
(`"ok":true`) and the real SDK shut down within the hub's budget. Unlike the
069-S evidence, stderr held **no** `agent event arrived after quiesce;
discarded` warning for the post-disconnect `session.shutdown`: it is now
classified informational and dropped at debug level (077.002-T). The remaining
stderr warnings were the existing `config has accepted-but-unconsumed keys`
(names only), the expected rejected-upgrade warning, and `unhandled agent SDK
event type` / `hub: unmapped agent event` warnings for the `abort`,
`agent.interrupted` and `model.turn_failed` events that the interrupted turn
produced; none affected the exit code.

## Dev Tunnel connect (devtunnel-connect)

A temporary tunnel was created with `devtunnel create --allow-anonymous`
(CLI 1.0.2014, logged in), given one port and hosted with `devtunnel host`.
Anonymous tunnel access was used only so the reference client needs no tunnel
token; intercom's own bearer authentication stayed enforced and the tunnel
host `<id>-<port>.usw3.devtunnels.ms` was listed in `[remote].allowed_hosts`.

* The approver and observer authenticated over `wss://<tunnel-host>/ws` and each
  received `session.hydrate` as the first frame.
* A wrong credential through the tunnel got `HTTP 401`.
* A prompt sent by the approver through the tunnel round-tripped and
  `agent.idle` reached both clients.
* Ctrl+C exited `intercom.exe` with code 0 (`shutdown complete`, `"ok":true`);
  both tunnel clients received `server.shutdown` then close 1001 `closing`.
* Afterwards the `devtunnel host` process was stopped and the tunnel was
  deleted with `devtunnel delete`.

## Assertion summary

| Assertion | Result |
|---|---|
| tui-render: credential panel rendered with both roles and the endpoint | pass |
| tui-render: local key input accepted (`y`, `Enter`, Ctrl+C) and the TUI quit with exit code 0 | pass |
| tool-approval-routing: `permission.requested` delivered to the TUI and both remote clients | pass |
| tool-approval-routing: `y` inside the 150 ms lock-out refused; later `y` then `Enter` approved (`allow_once`, `local-tui`) | pass |
| tool-approval-routing: remote approver resolved, observer refused (`role_not_permitted`), TUI modal cleared | pass |
| shutdown with a parked ask: exit 0, ask released, `server.shutdown` then close 1001, `"ok":true` | pass |
| devtunnel-connect: wss connect, hydrate, prompt round trip, 401, clean shutdown | pass |
| No stray `intercom`, verifier or `devtunnel` process; tunnel deleted | pass |
| No token text in this record | pass |

## Limits

* Windows ConPTY only. Windows Terminal, a dedicated conhost window, macOS and
  Linux were not exercised, and scrollback preservation was not re-measured
  (the TUI was not changed by 070-S).
* The remote surface was the Go reference client, not a browser.
* The ForceStop lock-wait extension (077.003-T) and the stuck-lane `App.Err`
  recording (077.004-T) need a wedged `Client.Stop` or lane worker, which the
  live runs did not produce; they remain covered by the race-enabled tests
  (`TestForceStopExtendsWaitForInFlightClientStopHoldingSDKLock`,
  `TestShutdownWaitsOutClientStopLockAcrossTheGracefulCutoff`,
  `TestShutdownRecordsHubRunErrCleanupIncompleteInAppErr`).
* Copilot processes owned by other sessions were already running on the machine
  and were not touched.
