---
title: "072-S ship session memory (pre-compaction)"
description: "Ship session memory for shipment 072-S (covering feature 079-F): execution, review, merge of PR 214, and post-merge closure state."
shipment: 072-S
feature: 079-F
date: 2026-10-29
status: completed
---

## Scope

Shipment 072-S (feature 079-F): `agent.Event` rune-safe bounds (`boundUTF8`,
`Truncated`), `tool.ended` and reasoning enrichment, write-diff and shell
`possibleUrls` surfacing in `PermissionRequest.Summary`, and unmapped-event log
rate-limiting. Tasks 079.001-T to 079.006-T, all done. Source stash entries were
retired by Stage at harvest. Branch
`feat/072-s-agent-event-rune-safe-bounds-tool-ended-reasoning-enrichment-write-diff-shell-url-surfacing-and-unmapped-event-log-rate-limiting`.
PR 214 merged as a merge commit `d0d863fae5ec4bdc1d197b35800223a786ed87c5`
(reviewed HEAD `54e4988a331686aea4d0222afbe8fd0abe0345e9`).

## Decisions

* The clipped write-permission preview keeps `out.Truncated=false` as the plan
  (Unit 3.1 AC 2) specifies; making it fail closed is out of scope under P-021 C1
  and was captured as stash 773B2CB2 (bug, high, requires deliberation).
* `boundUTF8` look-back is bounded; the hub fold flags turn truncation; clipped
  event text and the hub unmapped-type key are cloned so oversized SDK strings
  are not pinned (commit 54e4988).
* Design doc sections 3.1, 4.2, 4.3 and 5.2 were amended (079.006-T).

## Closure state

* a1: 079-F moved active to done before the cascade (all six descendants done).
* Classify-only selected `cascade` (`engine_verdict: VERIFIED`). The mutating
  `autoharness shipment cascade-close --shipment 072-S --feature 079-F` ran with
  the per-closure approval relayed by the Orchestrator and exited 0
  (`postcondition_verdict: pass`). `DARK_MODE_ACTIVE` alone was not treated as
  approval.
* Archive verified: 072-S (`archived_status: shipped`), 079-F, 079.001-T to
  079.006-T; no archive deletions (P-007). Lock released via
  `.github/skills/file-lock/scripts/release_lock.ps1` (the path `scripts/` in the
  task text does not exist). `backlogit sync` indexed 779 artifacts.
* Live verification: ConPTY, live Copilot, loopback (21 assertions) and Dev Tunnel
  (17 assertions), all pass; tunnel deleted; record in
  `docs/closure/evidence/072-S-live-verification-evidence.md`.

## Learnings

* The ConPTY harness must set `STARTF_USESTDHANDLES` or the child inherits the
  parent's redirected stdio and refuses to draw the credential panel.
* The throwaway verifier lived under `tools/_verify072` (ignored by `./...`) and
  was deleted without commit.

## Next steps

Commit and push the closure branch, open the closure PR with the readiness block,
request Copilot review, pass CI and the P-018 gate, merge, return to `main`.
Follow-up stash: 773B2CB2 (created before merge); none new at closure.
