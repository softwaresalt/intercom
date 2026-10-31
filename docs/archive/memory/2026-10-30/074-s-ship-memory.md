---
title: "074-S Ship session memory"
description: "Ship session memory for shipment 074-S / feature 081-F: PR 223 merged, awaiting per-closure cascade-close approval."
date: 2026-10-30
agent: ship
shipment: 074-S
feature: 081-F
status: awaiting-cascade-approval
---

## Completed

* Intake: index sync, queued-with-active-work check, pre_claim topology gate, intake reconcile (`.backlogit/reconcile/074-S-pre-queued.md`), claim, post_claim global verification (`CLAIM_VERIFY_OK`)
* 081.001-T to 081.005-T built test-first (red confirmed, then green), each committed and moved to `done`
* Review gate: Security, Correctness and Adversarial reviewers; no P0/P1; in-scope findings fixed (20e1cef); two out-of-scope findings captured as stash ACD578AE and D46C1B38 (P-021 C2)
* PR 223 opened, CI first failed `lint` (unparam), fixed in 7c0a3be, then all 14 checks green; P-018 `NOT_APPLICABLE`; merged by merge commit 966f6a9 under dark-mode pre-authorization

## Decisions

* Clipped writes stay remotely approvable when `remote_approve_destructive = true`; recorded as accepted exposure in ADR MVP-3 (074-S) and deferred to stash ACD578AE, since new remote policy exceeds the authorized change (P-021 C1)
* `PreviewClipped` is honoured only for write-kind asks in the hub
* Ran `golangci-lint run ./...` locally after the CI lint failure; it was available all along and should be part of the pre-push gate
* The `file-lock` script needs `-WorkspaceRoot` on this workspace (widening guard)

## State

* Branch: `post-merge/081-f-write-permission-preview-safety-shell-url-anti-spoofing-and-unknown-kind-local-tui-authorization`
* 081-F moved to `done` (Step 6.1.a1 S4). 074-S still `active`. Pre-done reconcile `PROCEED`; classify-only selected `cascade`; lock released
* Evidence: `docs/closure/evidence/074-S-081-F-close-evidence.json` (`pre_close`)

## Next steps

* Operator approves the mutating `autoharness shipment cascade-close` run
* Re-acquire the file lock, run the approved command, then the INV-12 disposition, `shipment-reconcile mode: post`, source-stash cleanup, `compact-context` (P-020), finalize the closure artifact and run the closure-evidence gate, backlog index resync, commit, push and open the closure PR for approval
* Hold 075-S until closure and compaction complete
