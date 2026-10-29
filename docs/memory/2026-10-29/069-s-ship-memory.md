---
title: "069-S ship session memory"
description: "Ship session memory for shipment 069-S (076-F resilient TUI reconnect, sanitization, modal anti-spoofing, viewport hardening)"
date: 2026-10-29
status: "in-progress"
---

## Items completed

* 076.001-T: in-place `Conn.Rehydrate` in `tui.Feed` with bounded redial (commit d85d671)
* 076.002-T: sanitize invisible blank runes, `isBlankSummary` (commit 6a3e4be)
* 076.003-T: deny-only blank summaries and indented modal summary (commit 22646ca)
* 076.004-T: 150ms approval delay, stuck-resync rehydrate, shutdown across gap (commit 43abaea)
* 076.005-T: transcript cap, wrap cache, prompt width, mouse wheel (commit 80c3132)

## Decisions

* Added `Options.NoMouse` in `internal/app/ui.go`, set when a custom `Input` is supplied. On Windows,
  mouse mode restarts the Bubble Tea input reader and swallows the first key on a pipe, breaking e2e tests.
* Local review: READY_WITH_FOLLOWUPS. P3 (kind/tool-call wrap spoofing) deferred as stash A999ABD6 (P-021 C2).

## Known flakes (pre-existing, timing-sensitive)

* `TestStalledRendererNeverBlocksHubAndStateConverges`
* `TestIntegrationExpiredWrongAndWrongRoleCredentials`

## Next steps

* Push, open PR, CI plus Copilot review (P-018), merge, post-merge closure.
