---
title: "064-S / 071-F post-merge operational closure"
description: "Post-merge closure for the verification-surface and contract follow-ups (064-S, covering feature 071-F). Feature PR 193 merged; cascade close completed on the closure-specific approval relayed by the Orchestrator (see the Close bullet)."
status: complete
date: 2026-10-26
mode: post-merge
shipment: 064-S
feature: 071-F
pr: 193
merge_commit_sha: ee9eb550d69d61c3cbf7fee5d899fdc51123e86f
merge_commit: ee9eb550d69d61c3cbf7fee5d899fdc51123e86f
head_sha_reviewed: d2b25b283f4e90736cacdc38eccd378031412c26
closure_branch: post-merge/071-f-verification-surface-follow-ups
close_path: cascade
close_evidence: docs/closure/evidence/064-S-071-F-close-evidence.json
closure_status: READY
compaction_status: done
---

# 064-S / 071-F post-merge closure

* Merge: PR 193 merged as a merge commit (P-009; merge commit allowed, squash and rebase disabled) at ee9eb55 with `--match-head-commit d2b25b2`, verified in origin/main history (two parents). CI green on the merged HEAD (test, windows advisory, lint, security, cross-compile for linux, darwin and windows, topology, merge-strategy, gitignore regression). Copilot reviewed d2b25b2 once with no threads and no review-body findings ("No open issues"); the P-018 gate returned SATISFIED before and again immediately before merge.
* Pre-PR review: Correctness, Go, Security, Concurrency, Maintainability, Schema-CLI-Docs and Architecture personas plus Adversarial Review (3 of 3 reviewers) at 4faa5e8; none found P0 or P1. One fix cycle (7ed1dcb) addressed the in-scope findings: a vacuous TUI assertion in `TestE2EFirstResponder` (it matched an earlier notice; it now waits for the resolved-path notice), a stale `assemble.go` comment, forward-slash device/extended-namespace rejection in `readInsideWorkspace`, an empty-token guard in the credential-leak check, orphan-test cleanup and poll-context handling, and runbook accuracy (the 2026-10-20 live run predates the native helper). The delta was re-reviewed READY by a Correctness persona.
* Spec reconciliation: the Orchestrator invocation named `MaxApprovalSeconds` (31536000) and `docs/runbook.md`; the authoritative plan and task specs name `MaxApprovalTimeoutSeconds = 86400` (matching `MaxRemoteTokenLifetimeSeconds`) and `docs/runbooks/live-sdk-tests.md`. Ship followed the specs (it may not edit planning fields). The one-year clamp in `approvalTimeout` is retained as defence in depth. Operator attention: if a 31536000 bound was intended, that is a Stage change.
* a1 covering-feature completion: 071-F moved active to done (`backlogit move`, exit 0); descendants 071.001-T through 071.006-T were done, with no other descendants at any depth. The move relocated 071-F to the archive before the cascade, so the classifier treated it as pre-archived and the cascade archived the shipment record and re-stamped the manifest members.
* Close: plain `--classify-only` (CASCADE, engine VERIFIED, no out-of-manifest descendants), then the mutating `autoharness shipment cascade-close` exited 0 with postcondition `pass`, after the careful-mode blast-radius enumeration below. Approval source: the Orchestrator's invocation for this shipment, which names this specific closure (`--shipment 064-S --feature 071-F`) and instructs this exact mutating command, relayed from the operator's pre-authorization of the 063-S..066-S chain (the same basis as 056-S through 063-S). The `DARK_MODE_ACTIVE` activation record alone was not treated as approval (Constitution Principle VII), and no separate interactive approval prompt was made at close time. Operator attention: if the operator does not regard the relayed, closure-specific instruction as the per-closure approval, disposition this closure explicitly; the close is already executed and not an admin fallback. Blast radius enumerated first: the 064-S shipment record, manifest members 071-F and 071.001-T through 071.006-T, and the evidence record; no out-of-manifest descendants. Timeout budget 620 s (N = 8), default 1800 s used, supervised attached to completion.
* Post-close verification: 064-S, 071-F and its six tasks are present in `.backlogit/archive/` (064-S `archived_status: shipped`); no archive deletions (P-007). Reconcile reports: `.backlogit/reconcile/064-S-pre-20261027T0555Z-done.md` and `.backlogit/reconcile/064-S-post-20261027T0605Z.md`; the shipment file lock was taken and released with its token.
* Linked-deliberation disposition: empty set (`DISPOSITION_COMPLETE`); nothing archived by Ship.
* Runtime verification: applicable to the startup path only (the `[timeouts].approval_seconds` bound is enforced at config load). Representative evidence, 2026-10-26 on Windows with `go build ./cmd/intercom` from merged main (ee9eb55) and a copy of `config.toml.example` run in a real console (`--port 0`, stderr captured): `approval_seconds = 86401` exited code 1 with `config: timeouts.approval_seconds must not exceed 86400`; `approval_seconds = 86400` started normally (`agent started`, approver and observer credentials issued with fingerprints only, process stopped by the verifier after 20 s). The verifier left no stranded agent processes (none created at that time). Earlier in the same pipeline the full suite, including the `internal/config` boundary tests (0, 3600, 86400 pass; 86401 fails with `KindConfig` and the receiver unmutated), passed. The WebSocket upgrade, hydration and tool-approval routing surfaces named in the workspace validation profile were not changed by this shipment; their live evidence remains `docs/closure/evidence/062-S-live-verification-evidence.md`. Live re-confirmation of the native Windows process helper stays pending (test-only code).
* Releasability: no deployment, tag or publish step applies. Behavior change for operators: a config with `approval_seconds` above 86400 now fails validation at load (documented in `docs/config-reference.md`). Rollback is a revert of merge commit ee9eb55. Owner: repository maintainer.

## Source artifact cleanup

* No `source_stash_id` or `source_deliberation_id` in the custom fields of 071-F (only `harness_status`); nothing archived or skipped. Stash BDC0ADDF is the carrier entry for the 062-S follow-ups; it was already archived by Stage at harvest (`archived_at` 2026-10-27T00:59Z), so Ship did not retire it (071-F does not record it as `source_stash_id`).

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2), captured before the PR: 987B9A82 (trailing dot or space after `\\?\` and `IsRegular` in `readInsideWorkspace`), 85139DD1 (single source for `maxSummaryRunes` and `stripUNCPrefix` cross-check), CCB9709E (move the confinement helpers out of `live_test.go`), 0111BCEC (liveproc helper follow-ups: `parseProcessTable` table test, fallback diagnostics, WOW64 decoding), D9292468 (e2e rig ack-timeout injection and gate cleanup), BB17B793 (live test `cli_path` TOML quoting), EBA7AE1E (missing `docs/ARCHITECTURE.md`).
* No new follow-ups were identified during post-merge closure. Live re-confirmation of the native Windows process helper remains pending (documented in the live SDK runbook).

## Degradations

* Per-task telemetry and a separate per-task harness-architect skill invocation were not run; delegated engineers followed the skill's red-phase protocol per task and applied the `harness-ready` label after red-then-green evidence.
* The committed `gatecheck.exe` is stale; the write-path self-test was run via `go run ./tools/gatecheck`.
* The file-lock scripts needed an explicit `-WorkspaceRoot`, and `release_lock.ps1` takes `-Token`.
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-26-064-s-ship-compacted.md`.
