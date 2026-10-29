---
title: "071-S live verification evidence (real Windows pseudoconsole, live Copilot runtime and Dev Tunnel, 2026-10-29 UTC)"
description: "Live verification of the tui-render, devtunnel-connect and tool-approval-routing manual checkpoints for the 071-S / 078-F closure: the real intercom.exe built from the merged 071-S code (main 67d3f5c) ran in a Windows ConPTY console against the live Copilot CLI runtime, with the local TUI and two authenticated remote WebSocket clients, on loopback and through a live Dev Tunnel. Credentials are redacted. No token text is recorded."
shipment: 071-S
feature: 078-F
date: 2026-10-29
---

This is the evidence for the `tui-render`, `devtunnel-connect` and
`tool-approval-routing` conditions of the 071-S / 078-F post-merge closure. It
comes from live runs, not CI. Shipment 071-S changed `hub.Cursor` and the
`wsclient` cursor (`server.shutdown` is always applicable across sequence gaps)
and the TUI permission modal (`wrapHeader` indents soft-wrapped `kind:` and
`tool call:` continuation lines). The checks below confirm that the modal still
renders correctly with the new code on the real console path, that approval
routing and shutdown still work against the real SDK, and that the wsclient
cursor change did not disturb a live session, including delivery of
`server.shutdown` to both remote clients.

All runs took place on 2026-10-29 at about 23:15Z to 23:17Z (16:15 to 16:17
local, UTC-7).

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` on the
  post-merge closure branch, whose Go source is identical to `main` at
  `67d3f5c03e89` (the 071-S merge, PR 212).
* Console: the binary ran under a Windows ConPTY pseudoconsole (120 by 30), so
  Bubble Tea used its real Windows console path; key input went through the
  pseudoconsole input pipe and the screen was read from its output stream (VT
  sequences stripped). This is a pseudoconsole, not a dedicated conhost window.
* Remote surfaces: the in-repo reference client `internal/transport/wsclient`
  connected as `approver` and as `observer`. The bearer tokens were read from
  the rendered credential panel in memory only.
* Agent: the live Copilot CLI (`C:\Tools\copilot.exe`, authenticated) through
  the production adapter. Prompts asked the agent to read `notes.txt` (a file in
  the temporary workspace) and reply `DONE`.
* Redaction: every logged string passed through an exact-match redactor for
  both tokens. Raw tokens were never written to this record. The temporary
  verifier (a throwaway program, deleted without being committed), its
  workspace, config, database and temporary `intercom.exe` were deleted
  afterwards.
* Runs: run A (loopback) and run B (Dev Tunnel) executed the same script: one
  intercom process, three prompt rounds, then Ctrl+C.

## Step A: credential panel rendered (both runs)

The credential panel (`Credentials (shown once - copy them now)`, the
`Endpoint: ws://127.0.0.1:<port>/ws` line, and the `approver:` and `observer:`
token lines with fingerprints and expiry) was readable from the pseudoconsole in
both runs. The first pre-check run also showed the `terminal too small` notice
only until the first resize was delivered, then the full TUI drew.

## Step C: WebSocket upgrade and rejected upgrade

* Approver and observer both authenticated and each received `session.hydrate`
  (run A over `ws://127.0.0.1`, run B over `wss://<tunnel-host>/ws`).
* A connection with a wrong credential was refused with `HTTP 401` in both runs
  (the TUI footer showed `rejected upgrades: 1 (last: 401 credential_rejected)`).

## tool-approval-routing

Round 1 (answered from the local TUI), identical in both runs: the approver sent
a prompt, the live agent asked to read `notes.txt`, and the hub published
`permission.requested` (`perm-1`, `permission_kind: read`, summary the absolute
path of `notes.txt`, a `toolu_...` tool call id) to **both** remote clients while
the TUI showed the `Permission request` modal. The modal's header lines read
`kind: read` and `tool call: toolu_...`, each on its own line at the modal's
inner left edge (the 078.002-T `wrapHeader` path).

* A `y` sent about 50 ms after the modal appeared did not arm approval (no
  `APPROVE?` prompt; the footer read `approval unlocks in a moment`), so the
  early key was refused during the 150 ms lock-out.
* A second `y` (after 800 ms) showed `APPROVE? press Enter to confirm`; `Enter`
  confirmed. Both remote clients received `permission.resolved` with
  `decision=allow_once`, `actor=local-tui`, `request_id=perm-1`, and the turn
  completed to `agent.idle` on both.

Round 2 (answered from the remote approver): a second prompt produced
`permission.requested` (`perm-2`) with the same `kind:` / `tool call:` modal
header lines. The observer's `resolve_permission` was refused (`action.refused`,
`reason=role_not_permitted`). The approver's `resolve_permission` was accepted
and both clients received `permission.resolved` (`allow_once`,
`actor=remote-approver`). The turn completed to `agent.idle` on both.

## Step D: shutdown with a parked permission ask (078-F path)

Round 3 sent a third prompt and waited until the agent had a `permission.requested`
ask parked mid-turn, then sent Ctrl+C to the TUI.

| Measure | Run A (loopback) | Run B (Dev Tunnel) |
|---|---|---|
| `intercom.exe` exit code | 0 | 0 |
| Time from keystroke to exit | 1.227 s | 0.692 s |
| `server.shutdown` received by approver | yes | yes |
| `server.shutdown` received by observer | yes | yes |
| Token text in the record | none | none |

Both remote clients received `server.shutdown` even though a permission ask was
parked and the event stream had advanced through several turns (the
always-applicable terminal delivery path that 078.001-T hardened). The parked ask
was released rather than left hanging and the process exited with code 0.

## Dev Tunnel connect (devtunnel-connect)

A temporary tunnel was created with `devtunnel create --allow-anonymous` (CLI
1.0.2014, logged in), given one port (38471) and hosted with `devtunnel host`.
Anonymous tunnel access was used only so the reference client needs no tunnel
token; intercom's own bearer authentication stayed enforced and the tunnel host
`<id>-38471.usw3.devtunnels.ms` was listed in `[remote].allowed_hosts`.

* The approver and observer authenticated over `wss://<tunnel-host>/ws` and each
  received `session.hydrate` as the first frame.
* A wrong credential through the tunnel got `HTTP 401`.
* Prompt rounds 1 to 3 ran through the tunnel with the same results as run A.
* Ctrl+C exited `intercom.exe` with code 0; both tunnel clients received
  `server.shutdown`.
* Afterwards the `devtunnel host` process was stopped and the tunnel was deleted
  with `devtunnel delete`.

## Assertion summary

| Assertion | Result |
|---|---|
| tui-render: credential panel rendered with both roles and the endpoint | pass |
| tui-render: modal `kind:` and `tool call:` header lines render on their own lines | pass |
| tui-render: local key input accepted (`y`, `Enter`, Ctrl+C) and the TUI quit with exit code 0 | pass |
| tool-approval-routing: `permission.requested` delivered to the TUI and both remote clients | pass |
| tool-approval-routing: `y` inside the 150 ms lock-out refused; later `y` then `Enter` approved (`allow_once`, `local-tui`) | pass |
| tool-approval-routing: remote approver resolved, observer refused (`role_not_permitted`) | pass |
| shutdown with a parked ask: exit 0, `server.shutdown` delivered to both remote clients | pass |
| devtunnel-connect: wss connect, hydrate, prompt rounds, 401, clean shutdown | pass |
| No stray `intercom`, verifier or `devtunnel` process; tunnel deleted | pass |
| No token text in this record | pass |

## Limits

* Windows ConPTY only. Windows Terminal, a dedicated conhost window, macOS and
  Linux were not exercised, and scrollback preservation was not re-measured.
* The remote surface was the Go reference client, not a browser.
* The soft-wrap indentation of an over-long `kind:` or `tool call:` value
  cannot be produced by the live agent (live kinds are short and tool call ids
  are about 30 characters), so that branch of 078.002-T is covered by the
  `internal/tui` unit tests only. Likewise, a gapped or foreign-generation
  `server.shutdown` (078.001-T) needs a sequence gap the live runs did not
  produce; the live runs confirm the contiguous path and are otherwise covered
  by the `hub` and `wsclient` cursor tests.
* The stderr JSON log lines were not captured separately from the screen; the
  exit code and the `server.shutdown` frames are the recorded shutdown evidence.
* Copilot processes owned by other sessions were already running on the machine
  and were not touched.
