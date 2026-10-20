---
title: "Process wiring: credential panel ordering, approval disarm choke point, shutdown context"
date: 2026-10-20
tags: ["go", "bubbletea", "tui", "credentials", "shutdown", "wiring"]
shipment: 062-S
severity: high
---

# Process wiring lessons (062-S)

## 1. Disarm an armed approval at one choke point

A two-key approve (`y` then Enter) is only safe if every other key disarms it. Handlers that return early (credential-panel Esc, Ctrl+R) skipped the disarm in `onModalKey`, so `y, Esc, Enter` approved. Put the disarm at the top of `onKey` for every key except Enter, before any shortcut can return. Test every shortcut and both a read and a destructive request.

## 2. Hand initial state to a model at construction

Sending initial credentials with `tea.Program.Send` after start raced user keys (Ctrl+R) and let a delayed startup message overwrite newer state and clear the in-flight guard. Pass initial state through the model constructor, make re-issue a no-op until state is known, and give re-issue success its own message type so only it clears the guard.

## 3. Shutdown and hub contexts

Run the hub on a context that signals never cancel and cancel it only after the ordered shutdown completes (design 3.1 step 6). A hub cancelled by SIGINT returns `ErrStopped` from step 0 and strands pending permissions. Run the shutdown on its own goroutine; a second signal hurries it but never skips steps.

## 4. TUI-only credentials need a terminal

If the TUI is the only place credentials are shown, refuse to start when stdout is not a terminal, and keep process logs off the alt-screen (discard when stderr is a terminal, else JSON to stderr).

## 5. Process hygiene

- Do not put a covering feature in the task-completion `backlogit move --status done` loop: it archives with the feature PR and bypasses the post-merge a1 gate.
- Some tool displays mask `Bearer` tokens as asterisks; check raw bytes before "fixing" docs.
- A Windows-only process-listing test helper (PowerShell `Win32_Process`) is load-flaky; prefer a native API (stash BDC0ADDF).
