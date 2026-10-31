---
title: "077-S / 084-F compacted Ship handoff"
description: "Consolidated Ship record for transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting and pumpRejects rate-limiting"
date: 2026-10-31
status: shipped
tags: [memory, compacted, transport, authz, 077-S]
source:
  - docs/archive/memory/2026-10-31/077-s-ship-session-checkpoint.md
  - docs/archive/memory/2026-10-31/077-s-post-merge-awaiting-cascade-approval.md
---

## Execution

* Built 084.001-T through 084.008-T: hub `ConnectOptions.Validate` atomic revalidation (`5107b42`), wire feedback bounds and close on hub-step failure (`24755a8`), off-accept-loop over-cap 503 refusal worker (`a849ab7`), `RecordReject` accounting and `grantValidator` (`c799270`), depguard layering (`84dcf99`), wsclient URL checks (`e6c3f53`), rate-limited `pumpRejects` (`82a11d2`), ADR MVP-2 and runbook (`616f07a`)
* Review fixes `dc065d1`: validator leak on hydrate failure and refusal queue drain race
* Local review found no P0/P1. PR 230 merged by merge commit `3b1135530989eb8fe60a55fcab768ca61fd3cd74`

## Decisions

* `dispatchWire` closes the socket after a hub-step credential failure only when `revalidate` reports `closeConn`, so shutdown revocation keeps the drain open
* Config conformance test asserts containment only; the transport is intentionally more lenient
* Deferred scope captured under P-021 C2: stash `7919A9A1` (wsclient still imports `internal/hub`)
* `internal/tui` flaked once under `-race` load and passed on reruns; the package is untouched

## Closure

* Attempt 1 of the mutating cascade-close exited 6; its evidence is preserved at `docs/closure/evidence/077-S-084-F-close-evidence.attempt1-exit6.json`
* Step A (`backlogit shipment repair-evidence` for 084.006-T and 084.007-T, then sync) repaired the blocker
* Restored and resolved `checkpoint-20261031-181948.json` after operator confirmation
* Operator approved at 2026-10-31T18:23:21Z. Attempt 2 exited 0, `postcondition_verdict: pass`, `shipment_record_archived_status: shipped`, ten records archived, `returned_ids` empty, disposition complete with no dispositions
* Post-mode reconcile: `.backlogit/reconcile/077-S-post-2026-10-31T18-30-00Z.md`, `PROCEED`. Evidence: `docs/closure/evidence/077-S-084-F-close-evidence.json`
