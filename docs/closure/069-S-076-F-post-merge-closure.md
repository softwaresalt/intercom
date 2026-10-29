---
title: "069-S / 076-F post-merge operational closure"
description: "READY after live verification of the release-required manual checkpoints (docs/closure/evidence/069-S-live-verification-evidence.md). Post-merge closure for resilient TUI reconnect, sanitization, modal anti-spoofing, and viewport hardening (069-S, covering feature 076-F). Feature PR 207 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-29
mode: post-merge
shipment: 069-S
feature: 076-F
pr: 207
merge_commit_sha: 5f561d2c6f25ff01db6ddc298e2392bf2380de27
merge_commit: 5f561d2c6f25ff01db6ddc298e2392bf2380de27
head_sha_reviewed: a55ec99a888fbb33eb98188962d2cedc07968ed2
closure_branch: post-merge/076-f-resilient-tui-reconnect-sanitization-modal-anti-spoofing-and-viewport-hardening
close_path: cascade
close_evidence: docs/closure/evidence/069-S-076-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. 076.001-T..076.005-T changed the TUI feed, sanitization, permission modal (indent, approval delay), transcript cap and mouse reporting."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/069-S-live-verification-evidence.md: on 2026-10-29 UTC the real intercom.exe built from the merged 069-S code (main 5f561d2) ran in a real Win32 console. The credential panel rendered (secrets redacted), Esc dismissed it, one unauthenticated upgrade showed a rejected-upgrade status, and Ctrl+C quit with exit code 0 with scrollback sentinels intact. Mouse reporting was on (console input mode showed the mouse flag) and no key input was swallowed. Mouse events themselves were not injected. Windows conhost only."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed TUI-side reconnect and redial, not tunnel registration."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/069-S-live-verification-evidence.md (Dev Tunnel section): on 2026-10-29 UTC the merged binary ran behind a live Dev Tunnel with intercom bearer auth enforced. The approver and observer authenticated over wss and each received session.hydrate, a prompt round-tripped to agent.idle, a wrong credential got HTTP 401, and Ctrl+C exited 0 with both clients receiving server.shutdown then close 1001."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment added the 150 ms approval delay, the indented permission summary, and deny-only blank summaries."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/069-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested reached the local Win32 console TUI and both remote clients. The summary rendered two columns right of the kind row. A y injected the instant the modal appeared did not arm approval; after the delay the hint 'y approve (then Enter to confirm)   n deny' showed and y then Enter produced permission.resolved (allow_once, actor local-tui) on both remotes. A remote approver resolution cleared the TUI modal (actor remote-approver); the observer was refused (role_not_permitted) and a late duplicate was refused (request_unknown)."
---

* Merge: PR 207 merged as a merge commit (P-009) at 5f561d2 with reviewed HEAD a55ec99, verified in origin/main history. CI was green on the final HEAD. The P-018 gate `autoharness gate copilot-review 207` returned `SATISFIED` after six Copilot threads were fixed and resolved across three rounds.
* Local review: initial review at 8c46ea6 returned `READY_WITH_FOLLOWUPS` with one P3. Re-reviews of the two fix deltas (f360a9f..f906c9b and f906c9b..a55ec99) returned `READY` with no P0-P2. Residual P3 advisory: the pump `shutdown` flag is not reset after a successful reconnect (cosmetic).
* Scope shipped: 076.001-T to 076.005-T (in-place `Conn.Rehydrate` with bounded redial in `tui.Feed`; sanitization of invisible blank runes and `isBlankSummary`; deny-only blank permission summaries and an indented modal summary; 150 ms `ApprovalDelay` with modal generation, disarm on resync/hydrate and rehydrate on a stuck resync; `server.shutdown` handled before the cursor-gap resync branch; capped transcript with wrap cache, prompt truncation by display width, mouse reporting and wheel scroll).
* Decisions: `app.Options.NoMouse` is set when a custom `Input` is supplied, because on Windows mouse mode restarts the Bubble Tea input reader and swallows the first key on a pipe. The feed pump stops after queuing `server.shutdown` so late broadcasts cannot evict it. Explicit operator P-001 exception over 062-S applied to this shipment. Process note: the Copilot-fix agent appended tests with `Add-Content` instead of the edit tool; gofmt was applied afterwards.
* Stash 80996C79 (Cursor/wsclient gap handling vs `server.shutdown`): the TUI half is closed by 076.004-T (`KindServerShutdown` is handled before the cursor-gap branch in `Model.onFrame`); the wsclient half remains for Stage.
* a1 covering-feature completion: 076-F moved active to done before the cascade; descendants 076.001-T through 076.005-T were done with no other descendants.
* Close: classify-only selected `cascade` (`engine_verdict: VERIFIED`); the mutating `autoharness shipment cascade-close` run (`run_id: a921f63ff6ea481bbd986a4ec4eab44f`) invoked `backlogit shipment ship 069-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 069-S --feature 076-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 069-S record, members 076-F and 076.001-T..076.005-T, and the evidence record; no out-of-manifest descendants. Postconditions all passed (`postcondition_verdict: pass`, shipment_status `shipped`, seven archived ids, no missing or unexpected ids).
* Post-close verification: 069-S (`archived_status: shipped`), 076-F and 076.001-T through 076.005-T are present in `.backlogit/archive/`; no archive deletions (P-007); the shipment lock was released with its token. Reconcile reports: `.backlogit/reconcile/069-S-pre-20261029T0236Z-queued.md`, `.backlogit/reconcile/069-S-pre-20261029T0405Z-done.md` and `.backlogit/reconcile/069-S-post-20261029T0412Z.md`. `backlogit sync` indexed 764 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (TUI code changed). The live checkpoints were exercised on 2026-10-29 UTC in a real Win32 console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients, and through a live Dev Tunnel; see `docs/closure/evidence/069-S-live-verification-evidence.md`. Full local gates before the PR: `gofmt -l`, `go vet ./...`, `go build ./...`, `go test -race ./...` and `gatecheck write-path --self-test` all passed.
* Releasability: `READY`. All three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) are satisfied by live evidence (see `conditions`). Residual risk: mouse wheel scrolling and the in-place feed rehydrate and redial paths are covered by tests, not by injected live mouse events or a live dropped-connection scenario. No deployment, tag or publish step applies. Rollback is a revert of merge commit 5f561d2. Owner: repository maintainer.

## Source artifact cleanup

* 076-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields; nothing archived or skipped here. The carrier stash entries (1DCB6108 TUI half and E9223340) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2), captured during local review: A999ABD6 (`DEFERRED SCOPE EXPANSION`: indent soft-wrap continuation lines of the permission modal `kind:` and `tool call:` values so they cannot imitate a header at column 0). Provisional priority low; requires deliberation by Stage.
* Pre-existing timing-sensitive tests noted in the ship memory: `TestStalledRendererNeverBlocksHubAndStateConverges`, `TestIntegrationExpiredWrongAndWrongRoleCredentials`, and a CI-only gatecheck wrapper timeout seen once on PR 207 (passed on rerun).

## Degradations

* None. Compaction (P-020): done. See `docs/memory/compacted/2026-10-29-069-s-ship-compacted.md`.
