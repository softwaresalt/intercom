---
title: "067-S / 074-F post-merge operational closure"
description: "Post-merge closure for hub fold, publish and lifecycle hardening (067-S, covering feature 074-F). Feature PR 202 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
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
    description: "Manual checkpoint required for release in the workspace profile (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. 074.001-T changed the live tui.Model.applyToolEnded tool-name and truncation backfill only; rendering, key handling and the TUI stop path are unchanged."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/065-S-tui-render-live-evidence.md covers the TUI surface (real Win32 console, injected key events, screen-buffer reads, Ctrl+C quit with exit code 0 and scrollback intact). The applyToolEnded change is covered by `go test -race ./internal/tui/...`, which passed on the closure branch (ok)."
---

* Merge: PR 202 merged as a merge commit (P-009) at a2cacf3 with reviewed HEAD b7a2934, verified in origin/main history.
* Local review: concurrency and correctness persona reviews found no P0-P2 findings. P3 advisories are recorded as residual risk: abortedTurnIDs cap eviction, empty-TurnID ask matching, and dual non-blocking sends on operator action failure.
* Scope shipped: 074.001-T to 074.005-T (Fold.applyDelta/applyToolEnded and live tui applyToolEnded parity, Hub.publish stub fold-error dedupe and markAllStale, Hub.rehydrate remote buffer drain, aborted-turn permission settling with turn-scoped ask identity, non-blocking Adapter.recheckForce and AdapterStopTimeout validation, operator lane worker join, and sanitized operator_action_failed KindError).
* Decisions: `apperr.KindIO` is used for the bounded lane-join error because `apperr.KindInternal` does not exist and adding a kind is outside the task file list. `internal/hub/permission.go` was also touched (`intakeAsk`, `permReasonTurnAborted`) because `handleAsk` lives there. The explicit operator P-001 exception over 062-S applied to this shipment.
* a1 covering-feature completion: 074-F moved active to done before the cascade; descendants 074.001-T through 074.005-T were done with no other descendants.
* Close: classify-only selected `cascade`; the mutating `autoharness shipment cascade-close` run (`run_id: eb9db0a3a48b454ab835440fd09a029a`) invoked `backlogit shipment ship 067-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 067-S --feature 074-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 067-S record, members 074-F and 074.001-T..074.005-T, and the evidence record; no out-of-manifest descendants. backlogit 1.12.0 prints `validating member ...` progress lines on stdout; the first evidence parse failed on them. The operator updated `autoharness.shipment_close.postclose.parse_ship_response`, and Ship re-parsed the captured stdout excerpt (ParsedResult, shipment_status `shipped`, seven archived ids) and finalized the evidence record to `mutation_state: completed`, `postcondition_verdict: pass` (every filesystem postcondition passed: baseline_invariant, disposition_byte_identical, parent_id_preserved, shipment_archived_shipped, no missing or unexpected archived ids, no linked-deliberation drift).
* Post-close verification: 067-S (`archived_status: shipped`), 074-F and 074.001-T through 074.005-T are present in `.backlogit/archive/`; no archive deletions (P-007); no shipment lock remains. Reconcile reports: `.backlogit/reconcile/067-S-pre-20261028T0441Z-queued.md`, `.backlogit/reconcile/067-S-pre-20261028T0608Z-done.md` and `.backlogit/reconcile/067-S-post-20261029T0013Z.md`. `backlogit sync` indexed 761 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (hub and TUI code changed; no flag or protocol change). Evidence, 2026-10-28 on Windows from the closure branch (merged main a2cacf3): `go build ./cmd/intercom` exited 0; `--help` lists the unchanged flags; started with stdout redirected, the binary logged `intercom starting` and refused to start with exit code 1 (`config: stdout is not a terminal ...`), the documented non-TTY refusal; no secret text appeared. `go test -race ./internal/tui/...` passed. The interactive console path was not re-run live; its evidence is `docs/closure/evidence/065-S-tui-render-live-evidence.md`.
* Releasability: `READY`; closure status `READY`. No deployment, tag or publish step applies. Rollback is a revert of merge commit a2cacf3. Owner: repository maintainer.

## Source artifact cleanup

* 074-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields (only `harness_status` and `scheduler_baseline_claim`); nothing archived or skipped here. The carrier stash entries (2C74C5FC, 550437DB, 8769C9A0) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* No new follow-ups were identified during post-merge closure. The P3 advisories above remain residual-risk notes.

## Degradations

* The first cascade-close evidence parse failed on backlogit 1.12.0 progress lines; resolved by the operator's `parse_ship_response` update and a re-parse of the captured stdout (see Close).
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-28-067-s-ship-compacted.md`.
