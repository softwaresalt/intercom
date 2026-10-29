---
title: "071-S / 078-F post-merge operational closure"
description: "READY after live verification of the release-required manual checkpoints (docs/closure/evidence/071-S-live-verification-evidence.md). Post-merge closure for Remote Cursor/wsclient terminal server.shutdown gap handling and TUI permission modal kind:/tool call: continuation anti-spoofing (071-S, covering feature 078-F). Feature PR 212 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-29
mode: post-merge
shipment: 071-S
feature: 078-F
pr: 212
merge_commit_sha: 67d3f5c03e893dde348ecdec08c6fd7b4c69a686
merge_commit: 67d3f5c03e893dde348ecdec08c6fd7b4c69a686
head_sha_reviewed: 1317960f343a3f8df7e6bb04d7c6c6dc54b0d249
closure_branch: post-merge/078-f-remote-cursor-wsclient-terminal-server-shutdown-gap-handling-and-tui-permission-modal-kind-tool-call-continuation-anti-spoofing
close_path: cascade
close_evidence: docs/closure/evidence/071-S-078-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input. 071-S changed the permission modal header wrapping (wrapHeader)."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/071-S-live-verification-evidence.md: on 2026-10-29 UTC the real intercom.exe built from the merged 071-S code (main 67d3f5c) ran in a Windows ConPTY console. The credential panel rendered (secrets redacted), the permission modal showed its kind: and tool call: header lines, local keys (y, Enter, Ctrl+C) were accepted, and Ctrl+C quit with exit code 0, including with a permission ask parked mid-turn. A separate ConPTY run confirmed scrollback preservation at stream level (sentinel lines written to the main buffer before launch, alternate screen entered then left on exit, no ESC[3J or ESC[2J emitted, exit 0). ConPTY only (not a dedicated conhost window); the soft-wrap branch for over-long header values is covered by unit tests only."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed the wsclient cursor, not tunnel registration or the transport."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/071-S-live-verification-evidence.md (Dev Tunnel section): on 2026-10-29 UTC the merged binary ran behind a live Dev Tunnel with intercom bearer auth enforced. The approver and observer authenticated over wss and each received session.hydrate, three prompt rounds round-tripped, a wrong credential got HTTP 401, and Ctrl+C exited 0 with both clients receiving server.shutdown. The tunnel was deleted afterwards."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment changed the modal rendering and the cursor that carries permission events to remote clients."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/071-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested (permission_kind read, absolute-path summary, toolu_ tool call id) reached the local TUI and both remote clients. An early y inside the 150 ms lock-out was refused, then y and Enter produced permission.resolved (allow_once, actor local-tui) on both remotes. A remote approver resolution produced permission.resolved (actor remote-approver) and the observer was refused (role_not_permitted). A third ask parked mid-turn was released during Ctrl+C shutdown with exit 0. Repeated on loopback and through the Dev Tunnel."
---

* Merge: PR 212 merged as a merge commit (P-009) at 67d3f5c with reviewed HEAD 1317960, verified in origin/main history. CI was green on the final HEAD (an earlier `tools/gatecheck` wrapper timing flake on e8aa687 did not recur). The P-018 gate `autoharness gate copilot-review 212` returned `SATISFIED` for the final HEAD. Merge used `--match-head-commit` under the `DARK_MODE_ACTIVE` pre-authorization; no admin fallback was used.
* Local review: single-reviewer adversarial pass found no P0/P1/P2 and six P3 observations; in-scope P3 test gaps (foreign-generation and repeated gapped `server.shutdown`) were added. Outcome `READY_WITH_FOLLOWUPS`, with follow-up handling by the stash entries below.
* Scope shipped: 078.001-T (`server.shutdown` always applicable across sequence gaps in `hub.Cursor` and `wsclient` cursor; duplicates discarded, contiguous advances, gapped or foreign leaves the cursor stale), 078.002-T (`tui.Model.wrapHeader` indents soft-wrap continuation lines of `kind:` and `tool call:` in the permission modal).
* a1 covering-feature completion: 078-F moved active to done before the cascade; descendants 078.001-T and 078.002-T were done with no other descendants.
* Close: classify-only selected `cascade` (`engine_verdict: VERIFIED`); the mutating `autoharness shipment cascade-close --shipment 071-S --feature 078-F` run invoked the shipment close and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure; the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 071-S record, members 078-F, 078.001-T and 078.002-T, and the evidence record; no out-of-manifest descendants. Postconditions all passed (`postcondition_verdict: pass`).
* Post-close verification: 071-S (`archived_status: shipped`), 078-F, 078.001-T and 078.002-T are present in `.backlogit/archive/`; no archive deletions (P-007); the shipment lock was released with its token. Reconcile reports: `.backlogit/reconcile/071-S-pre-20261029T2230Z-queued.md`, `.backlogit/reconcile/071-S-pre-20261029T2258Z-done.md` and `.backlogit/reconcile/071-S-post-20261029T2305Z.md`. `backlogit sync` indexed 778 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (TUI and cursor code changed). The live checkpoints were exercised on 2026-10-29 UTC under a ConPTY console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients, on loopback and through a live Dev Tunnel; see `docs/closure/evidence/071-S-live-verification-evidence.md`. Full local gates before the PR: `gofmt -l internal`, `go vet ./...`, `go build ./...`, `go test -race ./...` and `gatecheck write-path --self-test` all passed.
* Releasability: `READY`. All three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) are satisfied by live evidence (see `conditions`). Residual risk: a gapped or foreign-generation terminal `server.shutdown` and an over-long wrapped modal header cannot be produced by the live agent; they are covered by the `hub`, `wsclient` and `internal/tui` unit tests only. No deployment, tag or publish step applies. Rollback is a revert of merge commit 67d3f5c. Owner: repository maintainer.

## Source artifact cleanup

* 078-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields; nothing archived or skipped here. The carrier stash entries (80996C79, A999ABD6) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* FF578342 (low, task): deduplicate the `server.shutdown` cursor state machine shared by `hub.Cursor` and `wsclient.cursor` (P-021 C2 deferred scope expansion).
* 7BDC38FF (low, task): introduce a single shared helper for indenting continuation lines across every tui modal header (P-021 C2).
* 51C7CD42 (high, bug): `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` in `tools/gatecheck` uses a hard 20 s wait that times out under CI runner load (seen on PR 210 and PR 212; rerun passes) (P-021 C2).
* No new follow-ups were created during post-merge closure.

## Degradations

* None. Compaction (P-020): done. See `docs/memory/compacted/2026-10-29-071-s-ship-compacted.md`.
