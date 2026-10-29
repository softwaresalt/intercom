---
title: "068-S / 075-F post-merge operational closure"
description: "READY after live verification of the release-required manual checkpoints (docs/closure/evidence/068-S-live-verification-evidence.md). Post-merge closure for shutdown broadcast/drain fidelity and in-place hub Client.Rehydrate (068-S, covering feature 075-F). Feature PR 205 merged as a merge commit; cascade close completed with the per-closure approval relayed by the Orchestrator."
status: complete
date: 2026-10-29
mode: post-merge
shipment: 068-S
feature: 075-F
pr: 205
merge_commit_sha: 2fff94b6722ff763aafc04a5e44432f2733bd58c
merge_commit: 2fff94b6722ff763aafc04a5e44432f2733bd58c
head_sha_reviewed: 766a15cb5e24b1c7fdc69608fa454adef6284ff4
closure_branch: post-merge/075-f-shutdown-broadcast-drain-fidelity-and-in-place-rehydrate
close_path: cascade
close_evidence: docs/closure/evidence/068-S-075-F-close-evidence.json
closure_status: READY
releasability: READY
compaction_status: done
conditions:
  - id: "tui-render"
    description: "Manual checkpoint required for release (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. 075.002-T/075.004-T changed the hub client surface the TUI attaches to (in-place Client.Rehydrate, write-ack drain) and the shutdown path the TUI quit key triggers."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/068-S-live-verification-evidence.md: on 2026-10-29 UTC (2026-10-28 local) the real intercom.exe built from the merged 068-S code (main 2fff94b) ran in a real Win32 console. The credential panel rendered (secrets redacted), Esc dismissed it, one unauthenticated upgrade showed 'rejected upgrades: 1 (last: 401 credential_rejected)', and Ctrl+C quit through the TUI path with exit code 0 (516 ms and 670 ms in two runs) with scrollback sentinels intact. Windows conhost only."
  - id: "devtunnel-connect"
    description: "Dev Tunnel registration and authenticated remote connect. This shipment changed shutdown delivery to remote clients (terminal server.shutdown, write-ack drain), not tunnel registration."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/068-S-live-verification-evidence.md (Dev Tunnel section): on 2026-10-29 UTC the merged binary ran behind a live Dev Tunnel (devtunnel host, anonymous tunnel access, intercom bearer auth still enforced, host listed in allowed_hosts). The approver and observer authenticated over wss through the tunnel and each received session.hydrate, a prompt round-tripped to agent.idle through the tunnel, a wrong credential got HTTP 401, and Ctrl+C exited 0 with shutdown complete ok; both tunnel clients received server.shutdown then close 1001."
  - id: "tool-approval-routing"
    description: "Tool-approval first-responder routing across the local TUI and a remote client. This shipment changed the hub client lifecycle (in-place Client.Rehydrate keeps the TUI registered so a pending permission is not reaped; empty-set reap waits for pending connects)."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/068-S-live-verification-evidence.md: against the live Copilot runtime, permission.requested reached the local Win32 console TUI (modal rendered in CONOUT$) and both remote clients; a TUI y then Enter approval produced permission.resolved (allow_once, actor local-tui) on both remote surfaces; a remote approver resolution cleared the TUI modal (actor remote-approver); the observer was refused (role_not_permitted) and a late duplicate was refused (request_unknown)."
---

* Merge: PR 205 merged as a merge commit (P-009) at 2fff94b with reviewed HEAD 766a15c, verified in origin/main history. CI was green (the run concluded `success`; one job, `detect code changes`, still displayed as pending in the checks list although the workflow run was complete and the PR merge state was `CLEAN`). The P-018 gate `autoharness gate copilot-review 205` returned `SATISFIED`; Copilot's review found no issues.
* Local review: concurrency and correctness persona reviews found no P0/P1. The two in-scope P3 findings on `handleRehydrate` (terminal client refusal, duplicate stale queueing) were fixed in c088900. Residual-risk P3 advisories: `mint` takes `mintMu` without selecting on ctx; `Drained` assumes write-ack is enabled before any frame is dequeued (true in `newSession`).
* Scope shipped: 075.001-T to 075.004-T (hub intake sweep, pending-connect reap guard, terminal `server.shutdown` delivery to stale clients; write-ack `Drained` tracking and in-place `Client.Rehydrate`; `transport.session` write-ack and unconditional `AdapterStopTimeout`; `App.Reissue` honoring a cancelled ctx and concurrent remote drain in `shutdownScenario`).
* Decisions: terminal clients still receive later non-terminal frames when their buffer has room (existing shutdown tests require post-shutdown frames) and skip them silently when full; the plan text said to skip them. `reapUnlessConnecting` is the single guarded reap used at both `Hub.Run` call sites. The explicit operator P-001 exception over 062-S applied to this shipment. Process note: for 075.002-T..075.004-T the `harness-ready` label was applied after red was confirmed but after the production edit had been written (not yet run), recorded in the session memory.
* a1 covering-feature completion: 075-F moved active to done before the cascade; descendants 075.001-T through 075.004-T were done with no other descendants.
* Close: classify-only selected `cascade` (`engine_verdict: VERIFIED`); the mutating `autoharness shipment cascade-close` run (`run_id: 7a29b7fc268a4c51b4add74391db8124`) invoked `backlogit shipment ship 068-S` and exited 0, with the per-closure approval relayed by the Orchestrator for this specific closure (`--shipment 068-S --feature 075-F`); the `DARK_MODE_ACTIVE` record alone was not treated as approval. Blast radius: the 068-S record, members 075-F and 075.001-T..075.004-T, and the evidence record; no out-of-manifest descendants. Postconditions all passed (`postcondition_verdict: pass`, shipment_status `shipped`, six archived ids, no missing or unexpected ids, no linked-deliberation drift).
* Post-close verification: 068-S (`archived_status: shipped`), 075-F and 075.001-T through 075.004-T are present in `.backlogit/archive/`; no archive deletions (P-007); the shipment lock was released with its token. Reconcile reports: `.backlogit/reconcile/068-S-pre-20261029T0102Z-queued.md`, `.backlogit/reconcile/068-S-pre-20261029T0210Z-done.md` and `.backlogit/reconcile/068-S-post-20261029T0225Z.md`. `backlogit sync` indexed 763 artifacts.
* Linked-deliberation disposition: `DISPOSITION_COMPLETE`; nothing archived by Ship outside the close path.
* Runtime verification: applicable (hub, transport, app shutdown and credential code changed; no flag or protocol change). The live checkpoints were exercised on 2026-10-29 UTC (2026-10-28 local) in a real Win32 console against the live Copilot runtime with the local TUI plus approver and observer WebSocket clients, and through a live Dev Tunnel; see `docs/closure/evidence/068-S-live-verification-evidence.md`. Full local gates before the PR: `gofmt -l`, `go vet ./...`, `go build ./...`, `go test -race ./...` and `gatecheck write-path --self-test` all passed.
* Releasability: `READY`. All three release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) are satisfied by live evidence (see `conditions`). Residual risk: the stale-client terminal delivery and in-place `Client.Rehydrate` paths are covered by race-enabled tests rather than a live stale-client scenario, and the deferred Cursor/wsclient item below. No deployment, tag or publish step applies. Rollback is a revert of merge commit 2fff94b. Owner: repository maintainer.

## Source artifact cleanup

* 075-F carries no `source_stash_id` or `source_deliberation_id` in its custom fields (only `harness_status` and `scheduler_baseline_claim`); nothing archived or skipped here. The carrier stash entries (4AFE4DA4, AEF57AF1, 1DCB6108 and parts of 2C74C5FC and 8769C9A0) were handled by Stage at harvest.

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2), captured during local review: 80996C79 (`DEFERRED SCOPE EXPANSION`: the hub may drop the oldest queued frame to fit `server.shutdown`, creating a seq gap that makes a Cursor-based consumer discard it; Cursor/wsclient should treat `KindServerShutdown` as always applicable). Provisional priority low; requires deliberation by Stage.
* The P3 advisories above remain residual-risk notes.

## Degradations

* None. Compaction (P-020): done. See `docs/memory/compacted/2026-10-29-068-s-ship-compacted.md`.
