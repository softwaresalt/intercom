---
title: "TUI permission-modal safety rules and pinning every clock in tests"
date: 2026-10-20
tags: ["go", "bubbletea", "tui", "permissions", "testing", "linux", "clock"]
shipment: 061-S
severity: high
---

# TUI permission modal and test clock lessons (061-S)

## 1. Pin every clock the code under test reads

The hub exposes both `Config.Clock` and `Config.Now`. A test pinned only one, so timing depended on the real clock through the other. It passed on Windows and flaked on Linux. It was caught only by running a cross-compiled Linux test binary (`GOOS=linux go test -c`, run under WSL). Rule: when a component reads time through more than one injected seam, tests must pin all of them, and Linux binaries should be run before pushing timing-sensitive changes.

## 2. Modal approve safety

- Use a single shared predicate for "is the modal rendered" and use it from both `View` and the `approvable` check. Copilot found approve allowed at a terminal too small for `View` to render the modal, so the operator could approve something never displayed.
- Call `syncModal` after every mutation of the pending set. `already_resolved` and `request_unknown` dropped the head request without resynchronizing, leaving a stale modal.
- Gate approve on the request being fully viewed and require two keys; deny stays available at all times (blank-summary requests are deny-only).

## 3. Input fidelity versus display sanitization

Typed prompt input must be sent to the agent exactly as typed; sanitization applies only to what is displayed. Mixing them corrupts user intent. Display sanitization of invisible characters is deferred (stash E9223340).
