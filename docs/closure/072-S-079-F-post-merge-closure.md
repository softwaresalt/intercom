---
title: "072-S / 079-F post-merge operational closure"
description: "READY after live verification of the release-required manual checkpoints (docs/closure/evidence/072-S-live-verification-evidence.md). Post-merge closure for agent.Event rune-safe bounds, tool.ended and reasoning enrichment, write-diff and shell-URL surfacing, and unmapped-event log rate-limiting (072-S, covering feature 079-F). Feature PR 214 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-29
mode: post-merge
shipment: 072-S
feature: 079-F
pr: 214
merge_commit_sha: d0d863fae5ec4bdc1d197b35800223a786ed87c5
merge_commit: d0d863fae5ec4bdc1d197b35800223a786ed87c5
head_sha_reviewed: 54e4988a331686aea4d0222afbe8fd0abe0345e9
closure_branch: post-merge/079-f-agent-event-rune-safe-bounds-tool-ended-reasoning-enrichment-write-diff-shell-url-surfacing-and-unmapped-event-log-rate-limiting
close_path: cascade
close_evidence: docs/closure/evidence/072-S-079-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input. 072-S changed TUI rendering of reasoning and tool text."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/072-S-live-verification-evidence.md: on 2026-10-29 UTC the real intercom.exe built from the merged 072-S code (main d0d863f) ran in a Windows ConPTY console. The credential panel rendered (secrets redacted), the permission modal showed its kind: and tool call: header lines, local keys (y, Enter, Ctrl+C) were accepted, and Ctrl+C quit with exit code 0, including with a permission ask parked mid-turn. A separate ConPTY run confirmed scrollback preservation at stream level (sentinel lines written to the main buffer before launch, alternate screen entered then left on exit, no ESC[3J or ESC[2J emitted, exit 0). ConPTY only (not a dedicated conhost window)."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed agent events and hub publishing, not tunnel registration or the transport."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/072-S-live-verification-evidence.md (Dev Tunnel section): on 2026-10-29 UTC the merged binary ran behind a live Dev Tunnel with intercom bearer auth enforced. The approver and observer authenticated over wss and each received session.hydrate, three prompt rounds round-tripped, a wrong credential got HTTP 401, and Ctrl+C exited 0 with both clients receiving server.shutdown. The tunnel was deleted afterwards."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment changed the permission summary content (write diff, shell URLs) and the event path that carries permission events to remote clients."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/072-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested reached the local TUI and both remote clients. An early y inside the lock-out was refused, then y and Enter produced permission.resolved (allow_once, actor local-tui) on both remotes. A remote approver resolution produced permission.resolved (actor remote-approver) and the observer was refused (role_not_permitted). A third ask parked mid-turn was released during Ctrl+C shutdown with exit 0. Repeated on loopback and through the Dev Tunnel."
---

* Merge: PR 214 merged as a merge commit (P-009) at d0d863f with reviewed HEAD 54e4988, verified in origin/main history. Merge used the `DARK_MODE_ACTIVE` pre-authorization; no admin fallback was used.
* Scope shipped: 079.001-T (write diff, new-file content and shell `possibleUrls` in `PermissionRequest.Summary`), 079.002-T (`agent.Event` reasoning kinds, `Truncated`, `tool.ended` text, `boundUTF8`, `unmappedLogGate`), 079.003-T (reasoning mapping and gate wiring in `Adapter.onEvent`), 079.004-T (hub fold of reasoning, `tool.ended` text and `Truncated`), 079.005-T (live reasoning deltas, hub unmapped warn rate-limit, TUI rendering), 079.006-T (design doc sections 3.1, 4.2, 4.3 and 5.2).
* a1 covering-feature completion: 079-F moved active to done before the cascade; descendants 079.001-T to 079.006-T were done with no other descendants.
* Close: classify-only selected `cascade` (`engine_verdict: VERIFIED`); the mutating `autoharness shipment cascade-close --shipment 072-S --feature 079-F` run exited 0 with the closure-specific command approval relayed by the Orchestrator after `--classify-only` and blast-radius verification. The `DARK_MODE_ACTIVE` record alone was not treated as approval (Constitution Principle VII), matching 056-S to 071-S. Blast radius: the 072-S record, members 079-F and 079.001-T to 079.006-T, and the evidence record; no out-of-manifest descendants. Postconditions all passed (`postcondition_verdict: pass`).
* Post-close verification: 072-S (`archived_status: shipped`), 079-F and 079.001-T to 079.006-T are present in `.backlogit/archive/`; no archive deletions (P-007); the shipment lock was released with its token. Reconcile reports: `.backlogit/reconcile/072-S-pre-20261029T2330Z-queued.md`, `.backlogit/reconcile/072-S-pre-20261030T0100Z-done.md` and `.backlogit/reconcile/072-S-post-20261030T0130Z.md`. `backlogit sync` indexed 779 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE` (no linked deliberations); nothing archived by Ship outside the close path.
* Runtime verification: applicable (agent, hub and TUI code changed). The live checkpoints were exercised on 2026-10-29 UTC under a ConPTY console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients, on loopback (21 assertions) and through a live Dev Tunnel (17 assertions), all passing; see `docs/closure/evidence/072-S-live-verification-evidence.md`.
* Releasability: `READY`. All three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) are satisfied by live evidence (see `conditions`). Residual risk: the live prompts used a read permission, so write-diff and shell-URL summary surfacing and live reasoning deltas are covered by the `agent`, `hub` and `internal/tui` unit tests only. No deployment, tag or publish step applies. Rollback is a revert of merge commit d0d863f. Owner: repository maintainer.

## Source artifact cleanup

* 079-F carries no `source_stash_id` or `source_deliberation_id` that remained live; nothing archived or skipped here. Source stash entries were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* 773B2CB2 (high, bug): make a clipped write-permission preview fail closed or add a structured clipped flag and TUI banner, instead of asking the operator to approve a partially shown write (P-021 C2 deferred scope expansion; requires deliberation).
* No new follow-ups were created during post-merge closure.

## Degradations

* None. Compaction (P-020): done. See `docs/memory/compacted/2026-10-29-072-s-ship-compacted.md`; the original is archived at `docs/archive/memory/2026-10-29/072-s-ship-memory.md`.
