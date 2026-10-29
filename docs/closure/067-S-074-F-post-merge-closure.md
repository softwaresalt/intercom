---
title: "067-S / 074-F post-merge operational closure"
description: "BLOCKED pending operator live re-verification of three release-required manual checkpoints. Post-merge closure for hub fold, publish and lifecycle hardening (067-S, covering feature 074-F). Feature PR 202 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: blocked
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
closure_status: BLOCKED
releasability: BLOCKED
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. 074.001-T changed the live tui.Model.applyToolEnded backfill only, and 074.002-T changed adapter shutdown timing and Hub.Run's worker join, which the TUI stop path depends on."
    satisfied: false
    disposition: "deferred-live-recheck (automated evidence only; operator live re-run recommended)"
    evidence: "The prior live console run (docs/closure/evidence/065-S-tui-render-live-evidence.md, commit 2920ed1) predates this shipment and does not verify the current shutdown path, so it is not claimed as satisfying this checkpoint. Automated evidence on the closure branch (merged main a2cacf3): go test -race ./internal/tui/... ./internal/app/... ./internal/hub/... and ./tests/integration/... all ok, including the shutdown and wedged-TUI bounded-stop tests."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed remote hydration (Hub.rehydrate buffer drain) and publish behavior, not tunnel registration."
    satisfied: false
    disposition: "deferred-live-recheck (not exercised; tunnel code unchanged)"
    evidence: "No live Dev Tunnel run was performed for 067-S. Automated WebSocket connect and hydrate coverage in ./tests/integration/... passed with -race. Last live evidence: docs/closure/evidence/062-S-live-verification-evidence.md (predates this shipment)."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment changed permission handling (aborted-turn settling, turn-scoped ask identity via PermissionRequest.TurnID)."
    satisfied: false
    disposition: "deferred-live-recheck (automated evidence only; operator live re-run recommended)"
    evidence: "No live two-surface approval run was performed for 067-S. Automated coverage of permission routing, aborted-turn settling and next-turn ask preservation passed in go test -race ./internal/hub/... and ./tests/integration/... on the closure branch."
---

* Merge: PR 202 merged as a merge commit (P-009) at a2cacf3 with reviewed HEAD b7a2934, verified in origin/main history.
* Local review: concurrency and correctness persona reviews found no P0-P2 findings. P3 advisories are recorded as residual risk: abortedTurnIDs cap eviction, empty-TurnID ask matching, and dual non-blocking sends on operator action failure.
* Scope shipped: 074.001-T to 074.005-T (Fold.applyDelta/applyToolEnded and live tui applyToolEnded parity, Hub.publish stub fold-error dedupe and markAllStale, Hub.rehydrate remote buffer drain, aborted-turn permission settling with turn-scoped ask identity, non-blocking Adapter.recheckForce and AdapterStopTimeout validation, operator lane worker join, and sanitized operator_action_failed KindError).
* Decisions: `apperr.KindIO` is used for the bounded lane-join error because `apperr.KindInternal` does not exist and adding a kind is outside the task file list. `internal/hub/permission.go` was also touched (`intakeAsk`, `permReasonTurnAborted`) because `handleAsk` lives there. The explicit operator P-001 exception over 062-S applied to this shipment.
* a1 covering-feature completion: 074-F moved active to done before the cascade; descendants 074.001-T through 074.005-T were done with no other descendants.
* Close: classify-only selected `cascade`; the mutating `autoharness shipment cascade-close` run (`run_id: eb9db0a3a48b454ab835440fd09a029a`) invoked `backlogit shipment ship 067-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 067-S --feature 074-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 067-S record, members 074-F and 074.001-T..074.005-T, and the evidence record; no out-of-manifest descendants. backlogit 1.12.0 prints `validating member ...` progress lines on stdout; the first evidence parse failed on them. The operator updated `autoharness.shipment_close.postclose.parse_ship_response`, and Ship re-parsed the captured stdout excerpt (ParsedResult, shipment_status `shipped`, seven archived ids) and finalized the evidence record to `mutation_state: completed`, `postcondition_verdict: pass` (every filesystem postcondition passed: baseline_invariant, disposition_byte_identical, parent_id_preserved, shipment_archived_shipped, no missing or unexpected archived ids, no linked-deliberation drift).
* Post-close verification: 067-S (`archived_status: shipped`), 074-F and 074.001-T through 074.005-T are present in `.backlogit/archive/`; no archive deletions (P-007); no shipment lock remains. Reconcile reports: `.backlogit/reconcile/067-S-pre-20261028T0441Z-queued.md`, `.backlogit/reconcile/067-S-pre-20261028T0608Z-done.md` and `.backlogit/reconcile/067-S-post-20261029T0013Z.md`. `backlogit sync` indexed 761 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (hub, permission, adapter and TUI code changed; no flag or protocol change). Evidence, 2026-10-28 on Windows from the closure branch (merged main a2cacf3): `go build ./cmd/intercom` exited 0; `--help` lists the unchanged flags; started with stdout redirected, the binary refused to start with exit code 1 (`config: stdout is not a terminal ...`), the documented non-TTY refusal. `go test -race ./internal/hub/... ./internal/app/... ./internal/tui/... ./tests/integration/... -count=1` all passed. No live interactive console, Dev Tunnel or two-surface tool-approval run was performed for this shipment; those three manual checkpoints are recorded as unsatisfied conditions above rather than claimed from older evidence.
* Releasability: `BLOCKED`; closure status `BLOCKED` pending operator disposition (the closure-evidence gate accepts READY_WITH_CONDITIONS only when every condition is satisfied: true, and Ship may not self-disposition unsatisfied release-required checkpoints). Automated evidence shows no defect and no deployment, tag or publish step applies, but the three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) need operator live re-verification on return. Rollback is a revert of merge commit a2cacf3. Owner: repository maintainer.

## Source artifact cleanup

* 074-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields (only `harness_status` and `scheduler_baseline_claim`); nothing archived or skipped here. The carrier stash entries (2C74C5FC, 550437DB, 8769C9A0) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* No new follow-ups were identified during post-merge closure. The P3 advisories above remain residual-risk notes.

## Degradations

* The first cascade-close evidence parse failed on backlogit 1.12.0 progress lines; resolved by the operator's `parse_ship_response` update and a re-parse of the captured stdout (see Close).
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-28-067-s-ship-compacted.md`.
