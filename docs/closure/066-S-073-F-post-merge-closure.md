---
title: "066-S / 073-F post-merge operational closure"
description: "Post-merge closure for the maintainability refactors of internal/app (066-S, covering feature 073-F). Feature PR 198 merged as a merge commit; cascade close completed on the closure-specific approval relayed by the Orchestrator (see the Close bullet)."
status: complete
date: 2026-10-27
mode: post-merge
shipment: 066-S
feature: 073-F
pr: 198
merge_commit_sha: 65780da0d83a53182a56460093b80cc132b5d945
merge_commit: 65780da0d83a53182a56460093b80cc132b5d945
head_sha_reviewed: d588c7e4a558393632a72111ba252b8bef1d1520
closure_branch: post-merge/073-f-maintainability-refactors-of-internal-app
close_path: cascade
close_evidence: docs/closure/evidence/066-S-073-F-close-evidence.json
closure_status: READY_WITH_CONDITIONS
releasability: READY_WITH_CONDITIONS
compaction_status: done
conditions:
  - id: "p-002-harness-red-phase"
    description: "P-002 requires harness-architect to report Compilation PASS and Red Phase CONFIRMED (and the harness-ready label) for each task before implementation. The three 066-S tasks were zero-behavior-change refactors built characterization-first against the existing green suite, so no red phase exists and no harness-ready label was applied; this is a literal P-002 deviation that needs explicit operator disposition."
    satisfied: false
    disposition: "awaiting operator disposition"
    evidence: "Baseline and per-task green runs of go test -race ./internal/app/... are recorded in docs/memory/compacted/2026-10-27-066-s-ship-compacted.md and the PR 198 body; characterization tests late_agent_test.go and reap_test.go were added. Code is already merged (65780da); the disposition is to accept the deviation or direct a follow-up through Stage."
  - id: "tui-render"
    description: "Manual checkpoint required for release in the workspace profile (surface cli): the Bubble Tea TUI renders and accepts local key input without breaking terminal scrollback. This shipment is a move-only refactor of internal/app (the TUI watcher goroutine became a method and the shutdown steps became helpers), so the rendering code, key handling and TUI stop path are unchanged."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/065-S-tui-render-live-evidence.md covers the unchanged TUI surface (real Win32 console, injected key events, screen-buffer reads, Ctrl+C quit with exit code 0 and scrollback intact). Not re-run live for 066-S: the race-enabled unit and end-to-end suites (TestStopWedgedTUIIsBoundedAndReportsAnError, TestWedgedTUIShutdownIsBoundedAndSurfacesInAppErr, TestE2EShutdown, and the tui_test.go lifecycle tests) pass on the merged HEAD, and a startup smoke check on merged main reproduced the documented non-TTY refusal."
---

* Merge: PR 198 merged as a merge commit (P-009) at 65780da with `--match-head-commit d588c7e`, verified in origin/main history. CI green on the merged HEAD (test, windows advisory, lint, security, cross-compile for linux, darwin and windows, topology, merge-strategy, gitignore regression); merge state was CLEAN.
* Pre-PR review: Adversarial Review (3 of 3 reviewers), Correctness/Go/Concurrency personas and Maintainability/Architecture/Security personas; none found P0 or P1. All reviewers reported they had no shell and reviewed the working tree by reading, so Ship verified the `teardown` and `supervise` diff against `main` directly and ran every gate itself. One fix cycle (3e61cc7) addressed in-scope P3 findings: two stale comments left by the field rename, a shadowed loop index in `mint`, the `awaitHub` ordering precondition (callers must run `stopRemote` first) and the `partialRig` call budget derived from `len(credentialRoles)`. A stray scratch file created by a reviewer (`.review_scratch_none`) was removed before any commit.
* Copilot (P-018): requested after PR creation; it reviewed d588c7e with "approval recommended", no inline threads and no open issues. The gate returned SATISFIED (`exit_code: 0`, no unresolved threads) and again immediately before merge. No fix cycle was needed.
* Behavior-neutral deviations from a literal move-only (all in the PR body): in `teardown` `hubCancel()` now precedes `<-hubRunDone`, matching `supervise`'s long-standing order (`hub.Stop()` is idempotent); `supervise` also calls the idempotent `hub.Stop()` through `reap`; a TUI stop error now joins `a.shut.err` through `reap` rather than through `addErr`, so the order of messages in the joined error can differ when both the remote shutdown and the TUI stop fail (`errors.Is` is unaffected); the reject-notice pump starts inside `watchUI` instead of `assemble`, still after `ui.watchDone` is assigned synchronously. Failed-start `teardown` keeps its original order (hub `BeginShutdown`, UI, remote, agent, hub) because the agent stops between the UI and remote steps and the hub wait; it composes the same nil-guarded helpers as `reap` instead of calling it as one unit.
* Spec reconciliation: plan Unit 4.2 and task 073.002-T name `a.credMu`, `a.addUIErr`/`a.takeUIErr` and `zeroPlaintextCredentials`, and the plan describes the teardown order as UI stop before `BeginShutdown`. The code uses `a.mu`, `addErr`/`takeErrs` and `zeroPlaintext`, and `BeginShutdown` runs first (as on main); the existing names and order were preserved. Ship may not edit plan or planning fields, so the plan text is left for Stage to correct if desired.
* a1 covering-feature completion: 073-F moved active to done (`backlogit move`, exit 0); descendants 073.001-T through 073.003-T were done with no other descendants at any depth. The move relocated 073-F to the archive before the cascade, so the classifier treated it as pre-archived.
* Close: plain `--classify-only` (CASCADE, engine VERIFIED, selected path `cascade`), then the mutating `autoharness shipment cascade-close` run (which invoked `backlogit shipment ship 066-S` and recorded it in the evidence record) exited 0 with postcondition `pass`, after the careful-mode blast-radius enumeration below. Approval (Constitution Principle VII): the explicit per-closure approval was obtained before the mutating run. It is the operator's command approval for this specific closure (`--shipment 066-S --feature 073-F`), relayed verbatim in the Orchestrator's invocation of this Ship session, which states that the operator "relayed per-closure command approval for this specific closure ... in dark factory mode while AFK, matching 056-S..065-S". The approval was given in advance and relayed rather than prompted interactively at close time, and the invocation text is not a file in the repository, so it is cited here rather than linked. The `DARK_MODE_ACTIVE` activation record alone was not treated as approval. The close is not an admin fallback. Blast radius enumerated first: the 066-S shipment record, manifest members 073-F and 073.001-T through 073.003-T, and the evidence record; no out-of-manifest descendants. Default 1800 s timeout (budget B = 462 s for N = 5), supervised attached to completion.
* Post-close verification: 066-S, 073-F and its three tasks are present in `.backlogit/archive/` (066-S `archived_status: shipped`); no archive deletions (P-007). Reconcile reports: `.backlogit/reconcile/066-S-pre-20261027T1457Z-queued.md`, `.backlogit/reconcile/066-S-pre-20261027T1540Z-done.md` and `.backlogit/reconcile/066-S-post-20261027T1545Z.md`; the shipment file lock was taken and released with its token.
* Linked-deliberation disposition: empty set (`DISPOSITION_COMPLETE`); nothing archived by Ship.
* Runtime verification: applicable (startup and teardown code in `internal/app` was restructured, with no change to flags, protocol or TUI rendering). Evidence, 2026-10-27 on Windows from merged main (65780da): `go build ./cmd/intercom` exited 0; `--help` lists the unchanged flags (`--config`, `--log-level`, `--port`, `--workspace`); started with `--config config.toml.example --port 0` and stdout redirected to a file, the binary logged `intercom starting` and refused to start with exit code 1 (`config: stdout is not a terminal: the local TUI shows the credentials once and will not draw them to a file or pipe; run intercom in an interactive terminal`), exactly the behavior documented in `docs/runbooks/intercom-operator-runbook.md`. No token or secret text appeared in the output. The interactive console path (TUI render, key input, quit) and the tunnel and WebSocket surfaces were not changed and were not re-run live; their evidence is `docs/closure/evidence/065-S-tui-render-live-evidence.md` and `docs/closure/evidence/062-S-live-verification-evidence.md`, and the shutdown ordering is covered by the race-enabled `internal/app` unit and end-to-end tests that passed locally and in CI.
* Releasability: `READY_WITH_CONDITIONS`. The one unsatisfied condition is the P-002 deviation recorded in the frontmatter (`p-002-harness-red-phase`), which concerns process, not the shipped artifact: nothing blocks release. No behavior change for operators, no deployment, tag or publish step applies. The release-required `tui-render` checkpoint covers a surface this shipment did not change (see `conditions` in the frontmatter); the other manual checkpoints (`devtunnel-connect`, `tool-approval-routing`) cover surfaces it did not change either. Rollback is a revert of merge commit 65780da. Owner: repository maintainer.

## Source artifact cleanup

* No `source_stash_id` or `source_deliberation_id` in the custom fields of 073-F (only `harness_status`); nothing archived or skipped. Stash AD8472DD is the carrier entry for the 073-F scope; it was already archived by Stage at harvest (`backlogit stash get AD8472DD` reports not found among active entries), so Ship did not retire it.

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2), captured before the PR: 4B7308F2 (credential `zeroLocked` helper and `mint` early-out on `credsCleared`), 2F9C75F8 (reap fake-seam test and a `lateAgent` bind-versus-send race test), 26E15024 (move `watchUI`, `pumpRejects`, `onReject` into `ui.go`, `teardown` beside `reap`, and a `credentialState` struct).
* No new follow-ups were identified during post-merge closure.

## Degradations

* Per-task telemetry and a separate per-task harness-architect skill invocation were not run. The refactors are zero-behavior-change and characterization-first: the existing `go test -race ./internal/app/...` suite was confirmed green before the first change and after each task, and two small characterization tests were added (`late_agent_test.go`, `reap_test.go`). No red phase exists for a pure refactor, so the `harness-ready` label was not applied by Ship (the tasks archived without it); the P-002 red-phase gate was therefore not satisfied in its literal form. The tasks were not halted at Step 4.1 as P-002 directs. This is recorded as the unsatisfied condition `p-002-harness-red-phase` and the closure is `READY_WITH_CONDITIONS` until the operator dispositions it.
* Ship could not dispatch the `review` skill as a single invocation; the persona and Adversarial Review subagents were dispatched directly, and all reported they had no shell.
* The committed `gatecheck.exe` is stale; the write-path self-test was run via `go run ./tools/gatecheck write-path --self-test --root .`.
* The file-lock scripts needed an explicit `-WorkspaceRoot`, and `release_lock.ps1` takes `-Token`.
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-27-066-s-ship-compacted.md`.
