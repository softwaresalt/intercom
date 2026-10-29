---
title: "070-S / 077-F post-merge operational closure"
description: "READY after live verification of the release-required manual checkpoints (docs/closure/evidence/070-S-live-verification-evidence.md). Post-merge closure for Copilot adapter Stop/ForceStop resilience, latch/permission hardening and App.Err stuck-lane reporting (070-S, covering feature 077-F). Feature PR 210 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-29
mode: post-merge
shipment: 070-S
feature: 077-F
pr: 210
merge_commit_sha: af87a21b342a75bd02b0a5d5c59fbbba6581bc51
merge_commit: af87a21b342a75bd02b0a5d5c59fbbba6581bc51
head_sha_reviewed: 69bbf9fc64a0f614945ea62f4d43272d2a84a85f
closure_branch: post-merge/077-f-copilot-adapter-stop-forcestop-resilience-latch-permission-hardening-and-app-err-stuck-lane-reporting
close_path: cascade
close_evidence: docs/closure/evidence/070-S-077-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input. 070-S did not change the TUI; it changed the adapter and hub shutdown path the TUI runs on."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/070-S-live-verification-evidence.md: on 2026-10-29 UTC the real intercom.exe built from the merged 070-S code (main af87a21) ran in a Windows ConPTY console. The credential panel rendered (secrets redacted), local keys (y, Enter, Ctrl+C) were accepted, and Ctrl+C quit with exit code 0 and shutdown ok=true, including with a permission ask parked mid-turn. ConPTY only (not a dedicated conhost window); scrollback was not re-measured because the TUI was unchanged."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed neither tunnel registration nor the transport."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/070-S-live-verification-evidence.md (Dev Tunnel section): on 2026-10-29 UTC the merged binary ran behind a live Dev Tunnel with intercom bearer auth enforced. The approver and observer authenticated over wss and each received session.hydrate, a prompt round-tripped to agent.idle, a wrong credential got HTTP 401, and Ctrl+C exited 0 with both clients receiving server.shutdown then close 1001. The tunnel was deleted afterwards."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment hardened the adapter permission bridge (Zl/Zp and identifier sanitization, pre-send gone() re-check)."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/070-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested (permission_kind read, absolute-path summary, toolu_ tool call id) reached the local TUI and both remote clients. An early y inside the 150 ms lock-out was refused, then y and Enter produced permission.resolved (allow_once, actor local-tui) on both remotes. A remote approver resolution cleared the TUI modal (actor remote-approver) and the observer was refused (role_not_permitted). A third ask parked mid-turn was released during Ctrl+C shutdown with exit 0."
---

* Merge: PR 210 merged as a merge commit (P-009) at af87a21 with reviewed HEAD 69bbf9f, verified in origin/main history. CI was green on the final HEAD (one unrelated `tools/gatecheck` wrapper timing flake passed on rerun). The P-018 gate `autoharness gate copilot-review 210` returned `SATISFIED` after one Copilot thread was fixed and resolved.
* Local review: Concurrency, Correctness and Security reviewers found no P0/P1 and no out-of-scope expansion; three P2 test-quality findings were fixed (3f89a09). Outcome `READY_WITH_FOLLOWUPS`. Residual P3 observations: `forceStopWait` is evaluated once when ForceStop starts; a late terminal idle is dropped when ingress is full during shutdown; a stale non-abort idle arriving after the next turn start clears the new turn id (fails closed).
* Scope shipped: 077.001-T (Zl/Zp sanitization, `sanitizeIdent` with a 128-rune cap, pre-send `gone()` re-check), 077.002-T (latch quiesce on failed Start, informational `session.shutdown`, non-blocking late terminal idle delivery), 077.003-T (`SDKStopLockTimeout`, `forceStopWait`, `ForceStopBudget`, `hub.NewShutdown` budget folding and Force validation), 077.004-T (`Hub.Run` errors recorded in `App.Err` before `hubRunDone` closes).
* a1 covering-feature completion: 077-F moved active to done before the cascade; descendants 077.001-T through 077.004-T were done with no other descendants.
* Close: classify-only selected `cascade` (`engine_verdict: VERIFIED`); the mutating `autoharness shipment cascade-close` run invoked `backlogit shipment ship 070-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 070-S --feature 077-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. The relayed `--approved-by` and `--approval-mode` flags are not accepted by the installed CLI (first attempt exited 2 with no mutation), so the run used the supported flags. Blast radius: the 070-S record, members 077-F and 077.001-T..077.004-T, and the evidence record; no out-of-manifest descendants. Postconditions all passed (`postcondition_verdict: pass`, shipment_status `shipped`, six archived ids, no missing or unexpected ids).
* Post-close verification: 070-S (`archived_status: shipped`), 077-F and 077.001-T through 077.004-T are present in `.backlogit/archive/`; no archive deletions (P-007); the shipment lock was released with its token. Reconcile reports: `.backlogit/reconcile/070-S-pre-20261029T2130Z-queued.md`, `.backlogit/reconcile/070-S-pre-20261029T2215Z-done.md` and `.backlogit/reconcile/070-S-post-20261029T2225Z.md`. `backlogit sync` indexed 775 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (adapter and shutdown code changed). The live checkpoints were exercised on 2026-10-29 UTC under a ConPTY console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients, and through a live Dev Tunnel; see `docs/closure/evidence/070-S-live-verification-evidence.md`. Full local gates before the PR: `gofmt -l`, `go vet ./...`, `go build ./...`, `go test -race ./...` and `gatecheck write-path --self-test` all passed.
* Releasability: `READY`. All three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) are satisfied by live evidence (see `conditions`). Residual risk: the ForceStop lock-wait extension and the stuck-lane `App.Err` recording need a wedged `Client.Stop` or lane worker that the live runs did not produce; they are covered by race-enabled tests only. No deployment, tag or publish step applies. Rollback is a revert of merge commit af87a21. Owner: repository maintainer.

## Source artifact cleanup

* 077-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields; nothing archived or skipped here. The carrier stash entries (407291BE, 70BEBF29, FA1C38F7) were handled by Stage at harvest; FA1C38F7 items (a,b,c,d,f) shipped here, and any remaining sub-item stays with Stage.

## Follow-ups (stash, capture-only)

* None captured. No P-021 C2 deferred-scope entries were created in this shipment.
* Pre-existing timing-sensitive test noted: `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` (CI-only timeout on PR 210, passed on rerun).
* Tooling note for Stage/operator: the installed `autoharness shipment cascade-close` rejects `--approved-by` and `--approval-mode`; the Orchestrator relay text should drop them.

## Degradations

* None. Compaction (P-020): done. See `docs/memory/compacted/2026-10-29-070-s-ship-compacted.md`.
