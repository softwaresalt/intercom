---
title: "coder/websocket close semantics, Linux-only CI timing, and secret-scanner pitfalls in transport tests"
date: 2026-10-20
tags: ["go", "websocket", "coder-websocket", "concurrency", "ci", "gitleaks", "testing", "linux"]
shipment: 060-S
severity: high
---

# WebSocket transport pitfalls (060-S)

## 1. `Conn.CloseNow` does nothing useful once `Conn.Close` has begun

`github.com/coder/websocket` `Close` writes the close frame and then waits up to 5 seconds for the peer's close frame
(`waitCloseHandshake`). `CloseNow` takes the `casClosing` branch when a close is already in flight: it only waits for the
connection to be marked closed and never closes the socket. A "bounded abort" timer that calls `CloseNow` therefore did not
bound anything: a peer that never reads held its session and its connection slot for the library's 5s instead of the
configured `CloseGrace`.

Fix: the abort must close the underlying `net.Conn` (hand it to the session through `http.Server.ConnContext`; the slot
accounting wrapper's `Close` is idempotent) and cancel the read-pump context, then call `CloseNow`. Test with a peer that
never reads and assert both `ActiveSessions()==0` and `ActiveConnections()==0` within the grace. This passed on Windows by
accident of timing and failed on Linux CI, so run the transport tests against a Linux binary
(`GOOS=linux go test -c`, run through WSL) before pushing.

## 2. Graceful `http.Server.Shutdown` waits on pre-handshake connections

A connection that never completes its HTTP request is in `StateNew` and counts as idle only after 5 seconds. Draining
hijacked WebSocket sessions after `Shutdown` returns lets that wait consume the caller's whole deadline. Run the session drain
concurrently with `Shutdown`.

## 3. Sample WebSocket keys trip gitleaks

The RFC 6455 sample `Sec-WebSocket-Key` literal is flagged by the `generic-api-key` rule. Build a key at runtime (16 bytes,
base64) in tests instead of committing the literal.

## 4. Test hygiene

* The hub hands a prompt to the agent asynchronously relative to the acceptance reply: poll for the effect, do not sample once.
* A producer that floods a "fast" client with large frames overflows its buffer on a loaded runner; pace the producer and
  assert which client was dropped.
* When editing Go tests from PowerShell, a replacement string in single quotes keeps `\n` literal and turns the rest of the
  line into a comment (the test still compiles and passes, silently losing an assertion). Use the edit tool.
