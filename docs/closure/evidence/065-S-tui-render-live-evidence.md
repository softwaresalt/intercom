---
title: "065-S tui-render live evidence (real Win32 console, 2026-10-27)"
description: "Live verification of the tui-render manual checkpoint for the 065-S closure: the real intercom.exe built from main ran in a real Windows console (conhost), and the verifier injected real key events and read the rendered screen buffer. Credentials are redacted. No token text is recorded."
shipment: 065-S
feature: 072-F
date: 2026-10-27
---

This is the evidence for the `tui-render` condition of
`docs/closure/065-S-072-F-post-merge-closure.md`: the Bubble Tea TUI renders and
accepts local key input without breaking terminal scrollback. It comes from a
live run, not CI.

## Method

* Binary: `intercom.exe` built with `go build ./cmd/intercom` from `main` at
  commit `2920ed1444306fa6ce2c5d15e11ce849d99a995a` (the merge of PR 196, which
  contains the 065-S work from PR 195).
* Console: a temporary, untracked Go verifier allocated its own real Win32
  console (`FreeConsole`, then `AllocConsole`, so a real `conhost.exe` with a
  console input buffer and a 120 by 30 screen buffer). It opened `CONIN$` and
  `CONOUT$` as inheritable handles and gave them to `intercom.exe` as stdin and
  stdout, so the child saw a real terminal and Bubble Tea ran its real Windows
  console loop. Stderr went to a file.
* Deviation from the planned `CREATE_NEW_CONSOLE` plus `AttachConsole` route:
  Go's process start gives a child with `cmd.Stdin` and `cmd.Stdout` unset NUL
  handles (it always sets `STARTF_USESTDHANDLES`), which would defeat the
  terminal check. Sharing the verifier's own real console with explicit
  `CONIN$` and `CONOUT$` handles reaches the same state: a real conhost with real
  console handles on the child's stdin and stdout.
* Screen reads: `GetConsoleScreenBufferInfo` and `ReadConsoleOutputCharacterW`
  on a freshly opened `CONOUT$` for each read, so each read addresses the active
  screen buffer, including the alternate buffer the TUI draws in.
* Keystrokes: real `KEY_EVENT_RECORD` key-down and key-up pairs through
  `WriteConsoleInputW` on `CONIN$`.
* Redaction: every token-shaped run of 20 or more characters on a captured
  row was replaced with `<REDACTED-TOKEN>` before anything was saved. The
  verifier's raw screens were never written to disk, and the temporary
  directory was deleted afterwards. A few non-secret strings were also matched
  by the redactor (the word `accepted-but-unconsumed`, the sentinel labels and a
  session id); they are restored or left redacted below as noted.
* Port: the verifier picked a free loopback port (58401) and passed
  `--port 58401`. Config: a copy of `config.toml.example`.

## Step A: credential panel rendered

The panel appeared about 12 s after launch (the agent adapter start comes
first). Screen rows 0 to 10 and the bottom two rows of the 30-row screen:

```text
intercom | idle
Credentials (shown once - copy them now)
Endpoint: ws://127.0.0.1:58401/ws

approver: <REDACTED-TOKEN>
  fingerprint ad4b1ac6, expires 2026-10-27T12:47:06Z

observer: <REDACTED-TOKEN>
  fingerprint 4c71677c, expires 2026-10-27T12:47:06Z

Esc dismiss   Ctrl+R re-issue (revokes these)
...
credentials are shown once: copy them now (Esc dismisses)
> _
```

Checks: panel title, both role rows, the endpoint with the bound port and the
dismiss hint were present. Console state at this point: the child console
listed 3 attached processes (the verifier, intercom and one more, not
identified), the console input mode was `0x0088`
(`ENABLE_EXTENDED_FLAGS` and `ENABLE_WINDOW_INPUT`), so `ENABLE_PROCESSED_INPUT`
was off and Ctrl+C arrives as a key event for the TUI rather than as a console
signal.

## Step B: dismiss keystroke

Key: `Esc` (virtual key `0x1B`, down then up). `internal/tui/model.go` handles
`tea.KeyEsc` with the panel open by calling `dismissCredentials`. Result, rows 0
to 1 and the bottom two rows:

```text
intercom | idle
No activity yet. Type a prompt and press Enter.
...
connected
> _
```

The credential panel and the dismiss hint were gone and the main transcript and
status view rendered.

## Step C: rejected-upgrade status line

The verifier sent two unauthenticated requests to `http://127.0.0.1:58401/ws`;
both returned `HTTP 401`. The status row (second from the bottom) became:

```text
rejected upgrades: 2 (last: 401 credential_rejected)
> _
```

Stderr agreed: two `transport: upgrade rejected` records with `status` 401 and
`reason` `credential_rejected`.

## Step D: clean quit and scrollback

Key: Ctrl+C as real key events (left Control down, `C` down with character
`0x03` and `LEFT_CTRL_PRESSED`, `C` up, Control up). `internal/tui/model.go`
returns `tea.Quit` for `ctrl+c`.

* Exit: `intercom.exe` exited with code 0, 1627 ms after the keystroke (the
  documented budget is 3 s for `Quit`, then 3 s after `Kill`). Total uptime 15 s.
* Shutdown path: stderr logged `shutdown started` with `reason` `tui quit`, then
  `shutdown complete` with `ok` true. This confirms the TUI key path, not an OS
  signal, ended the process.
* Scrollback: three sentinel lines
  (`SCROLLBACK-SENTINEL-1`, `-2` and `-3`, each followed by
  `(written before launching intercom.exe)`) were written to the main screen
  buffer before launch. After exit they were still on the same rows, unchanged,
  and the main buffer held none of the TUI text (no `Credentials (shown once`
  and no `rejected upgrades`).

```text
                       before launch      after exit
buffer size            120 x 9001         120 x 9001
cursor (x, y)          (0, 3)             (0, 3)
window (l, t, r, b)    (0, 0, 119, 29)    (0, 0, 119, 29)
```

* Child processes: the `copilot.exe` process set was the same before and after
  the run (five pre-existing host processes, no new `copilot` process), and no
  `intercom.exe` remained.

## Stderr log (redacted)

```text
INFO  intercom starting
WARN  config has accepted-but-unconsumed keys  keys=["commands"]
INFO  agent started  session=<redacted>
INFO  credential issued  role=approver fingerprint=ad4b1ac6
INFO  credential issued  role=observer fingerprint=4c71677c
INFO  intercom ready  addr=127.0.0.1:58401
WARN  transport: upgrade rejected  status=401 reason=credential_rejected
WARN  transport: upgrade rejected  status=401 reason=credential_rejected
INFO  shutdown started  reason="tui quit"
WARN  agent event arrived after quiesce; discarded  event_type=session.shutdown
INFO  shutdown complete  ok=true
```

The `agent event arrived after quiesce; discarded` warning is the existing
quiesce guard dropping a late `session.shutdown` event; it did not affect the
exit code.

## Assertion summary

| Assertion | Result |
|---|---|
| A: credential panel rendered with both roles and the endpoint | pass |
| B: Esc dismissed the panel and the main view rendered | pass |
| C: status line showed `rejected upgrades: 2` | pass |
| D: exit code 0 through the TUI quit path | pass |
| D: scrollback sentinels intact, no TUI debris in the main buffer | pass |
| No stranded `copilot` child | pass |
| No token text in this record | pass |

## Limits

* Windows conhost only. Windows Terminal, and macOS or Linux terminals, were
  not exercised.
* The bounded-stop wedge case remains covered by the race-enabled tests named in
  the closure artifact, not by this run.
