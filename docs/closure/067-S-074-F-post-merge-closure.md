---
title: "067-S / 074-F post-merge operational closure"
description: "READY after live re-verification of the release-required manual checkpoints (docs/closure/evidence/067-S-live-verification-evidence.md). Post-merge closure for hub fold, publish and lifecycle hardening (067-S, covering feature 074-F). Feature PR 202 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-28
mode: post-merge
shipment: 067-S
feature: 074-F
pr: 202
merge_commit_sha: a2cacf32c29b5e3a0ee596f01f14686ceeb3775d
merge_commit: a2cacf32c29b5e3a0ee596f01f14686ceeb3775d
head_sha_reviewed: b7a2934658d4d97af08e1997ec6ae09dabc737cb
closure_branch: post-merge/074-f-hub-fold-publish-and-lifecycle-hardening
close_path: cascade
close_evidence: docs/closure/evidence/067-S-074-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. 074.001-T changed the live tui.Model.applyToolEnded backfill only, and 074.002-T changed adapter shutdown timing and Hub.Run's worker join, which the TUI stop path depends on."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/067-S-live-verification-evidence.md: on 2026-10-28 the real intercom.exe built from main (0fd7936, containing the 067-S merge a2cacf3) ran in a real Win32 console. The credential panel rendered (secrets redacted), Esc dismissed it, one unauthenticated upgrade showed 'rejected upgrades: 1 (last: 401 credential_rejected)', and Ctrl+C quit through the TUI path with exit code 0 (709 ms and 1061 ms in two clean runs) with scrollback sentinels intact and no stranded copilot child. One earlier run exited 1 on a bounded agent-stop timeout; it is recorded in the evidence as a non-reproduced intermittent anomaly and as a follow-up stash entry. Windows conhost only."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment changed permission handling (aborted-turn settling, turn-scoped ask identity via PermissionRequest.TurnID)."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/067-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested reached the local Win32 console TUI (modal rendered in CONOUT$) and both remote clients; a TUI y then Enter approval produced permission.resolved (allow_once, actor local-tui) on both remote surfaces; a remote approver resolution cleared the TUI modal (actor remote-approver); the observer was refused (role_not_permitted) and a late duplicate was refused (request_unknown). The all-denied repeat behaved the same."
---

* Merge: PR 202 merged as a merge commit (P-009) at a2cacf3 with reviewed HEAD b7a2934, verified in origin/main history.
* Local review: concurrency and correctness persona reviews found no P0-P2 findings. P3 advisories are recorded as residual risk: abortedTurnIDs cap eviction, empty-TurnID ask matching, and dual non-blocking sends on operator action failure.
* Scope shipped: 074.001-T to 074.005-T (Fold.applyDelta/applyToolEnded and live tui applyToolEnded parity, Hub.publish stub fold-error dedupe and markAllStale, Hub.rehydrate remote buffer drain, aborted-turn permission settling with turn-scoped ask identity, non-blocking Adapter.recheckForce and AdapterStopTimeout validation, operator lane worker join, and sanitized operator_action_failed KindError).
* Decisions: `apperr.KindIO` is used for the bounded lane-join error because `apperr.KindInternal` does not exist and adding a kind is outside the task file list. `internal/hub/permission.go` was also touched (`intakeAsk`, `permReasonTurnAborted`) because `handleAsk` lives there. The explicit operator P-001 exception over 062-S applied to this shipment.
* a1 covering-feature completion: 074-F moved active to done before the cascade; descendants 074.001-T through 074.005-T were done with no other descendants.
* Close: classify-only selected `cascade`; the mutating `autoharness shipment cascade-close` run (`run_id: eb9db0a3a48b454ab835440fd09a029a`) invoked `backlogit shipment ship 067-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 067-S --feature 074-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 067-S record, members 074-F and 074.001-T..074.005-T, and the evidence record; no out-of-manifest descendants. backlogit 1.12.0 prints `validating member ...` progress lines on stdout; the first evidence parse failed on them. The operator updated `autoharness.shipment_close.postclose.parse_ship_response`, and Ship re-parsed the captured stdout excerpt (ParsedResult, shipment_status `shipped`, seven archived ids) and finalized the evidence record to `mutation_state: completed`, `postcondition_verdict: pass` (every filesystem postcondition passed: baseline_invariant, disposition_byte_identical, parent_id_preserved, shipment_archived_shipped, no missing or unexpected archived ids, no linked-deliberation drift).
* Post-close verification: 067-S (`archived_status: shipped`), 074-F and 074.001-T through 074.005-T are present in `.backlogit/archive/`; no archive deletions (P-007); no shipment lock remains. Reconcile reports: `.backlogit/reconcile/067-S-pre-20261028T0441Z-queued.md`, `.backlogit/reconcile/067-S-pre-20261028T0608Z-done.md` and `.backlogit/reconcile/067-S-post-20261029T0013Z.md`. `backlogit sync` indexed 761 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (hub, permission, adapter and TUI code changed; no flag or protocol change). Evidence, 2026-10-28 on Windows from the closure branch (merged main a2cacf3): `go build ./cmd/intercom` exited 0; `--help` lists the unchanged flags; started with stdout redirected, the binary refused to start with exit code 1 (`config: stdout is not a terminal ...`), the documented non-TTY refusal. `go test -race ./internal/hub/... ./internal/app/... ./internal/tui/... ./tests/integration/... -count=1` all passed. The live checkpoints were then exercised on 2026-10-28 from main 0fd7936 (containing a2cacf3) in a real Win32 console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients (credential panel, Esc dismiss, 401 status line, TUI-first and remote-first permission routing, Ctrl+C exit code 0 with scrollback intact, no stranded copilot child); see `docs/closure/evidence/067-S-live-verification-evidence.md`. One of three live runs exited 1 on a bounded agent-stop timeout (`ForceStop did not return within 5s`); it was not reproduced in two further runs or in `TestLiveCopilotEndToEnd`, and is recorded as residual risk and stash entry 407291BE.
* Releasability: `READY`. The release-required manual checkpoints `tui-render` and `tool-approval-routing` are satisfied by live evidence (see `conditions`). `devtunnel-connect` is no longer listed: 067-S did not change Dev Tunnel registration, and its last live evidence remains `docs/closure/evidence/062-S-live-verification-evidence.md` (the same treatment as 065-S); the changed remote hydration path was exercised live over loopback WebSocket. It was first recorded as `BLOCKED` in PR 203 and is closed by the evidence PR that updated this record. Residual risk: the one non-reproduced failed shutdown above. No deployment, tag or publish step applies. Rollback is a revert of merge commit a2cacf3. Owner: repository maintainer.

## Source artifact cleanup

* 074-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields (only `harness_status` and `scheduler_baseline_claim`); nothing archived or skipped here. The carrier stash entries (2C74C5FC, 550437DB, 8769C9A0) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2), captured during live verification: 407291BE (intermittent failed agent stop on Ctrl+C: `stop copilot client: timed out; ForceStop did not return within 5s`, observed once in three live runs).
* The P3 advisories above remain residual-risk notes.

## Degradations

* The first cascade-close evidence parse failed on backlogit 1.12.0 progress lines; resolved by the operator's `parse_ship_response` update and a re-parse of the captured stdout (see Close).
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-28-067-s-ship-compacted.md`.
