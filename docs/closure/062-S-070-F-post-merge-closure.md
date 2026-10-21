---
title: "062-S / 070-F post-merge operational closure"
description: "Post-merge closure for MVP S7 process wiring and end-to-end verification (062-S, covering feature 070-F). Feature PR 185 merged; cascade close completed with operator-authored pre-approval. Final shipment of the MVP DAG."
status: complete
date: 2026-10-20
mode: post-merge
shipment: 062-S
feature: 070-F
pr: 185
merge_commit_sha: 6290151f08fcb300b2bd787c15ce28382242b233
merge_commit: 6290151f08fcb300b2bd787c15ce28382242b233
head_sha_reviewed: 6930bde0aef49acb74ce306b578a4532dc254d2c
closure_branch: post-merge/062-mvp-s7-process-wiring
close_path: cascade
close_evidence: docs/closure/evidence/062-S-070-F-close-evidence.json
closure_status: READY_WITH_CONDITIONS
releasability: READY_WITH_CONDITIONS
compaction_status: done
conditions:
  - id: "live-sdk-test-pass"
    description: "The opt-in live SDK test (INTERCOM_LIVE_SDK_TESTS; scenario steps 2, 5 and 7 against a real Copilot runtime) has not been run. The MVP definition of done requires it to pass. Operator-owned: it needs an installed, authenticated Copilot CLI and cannot run in the unattended session."
    satisfied: false
    disposition: "partial-evidence"
    evidence: "Partial: docs/closure/evidence/062-S-live-verification-evidence.md section A (2026-10-20, operator-present, Copilot CLI 1.0.100): TestLiveCopilotEndToEnd and TestLiveStartAndStop passed and SQ-b verified for the read kind. Still missing: a recorded live run showing shell and write calls reach the permission handler (ADR MVP-1), and SQ-c re-verification (inconclusive; design section 4.4 stays a conservative, unverified assumption)."
  - id: "devtunnel-connect"
    description: "Manual checkpoint required for release in the workspace profile: Dev Tunnel registration succeeds (anonymous access off) and an authenticated remote client connects, including Authorization header pass-through. Tunnel creation is outside intercom. Operator-owned."
    satisfied: true
    disposition: "satisfied-for-reference-client"
    evidence: "docs/closure/evidence/062-S-live-verification-evidence.md section B (2026-10-20, devtunnel CLI 1.0.2014): private tunnel registered with --allow-anonymous false; Go reference client connected over wss with the Authorization header unmodified. Not tested: real browser, second machine, TUI-launched binary, Origin/allowed_origins, long-lived reconnect."
  - id: "tool-approval-routing-live"
    description: "Manual checkpoint required for release: first-responder routing across the local TUI and a remote client against a real runtime. Verified headlessly only (internal/app e2e tests)."
    satisfied: false
    disposition: "partial-evidence"
    evidence: "Partial: docs/closure/evidence/062-S-live-verification-evidence.md sections A and B (2026-10-20): remote-only approval of a read permission through the tunnel and in TestLiveCopilotEndToEnd, both with headless harnesses (the TUI was disabled). Still missing: a live run with the local TUI and a remote client both connected, showing first-responder arbitration; shell and write kinds were not exercised live."
---

# 062-S / 070-F post-merge closure

* Merge: PR 185 merged as a merge commit (P-009; merge commit allowed, squash and rebase disabled) at 6290151 with `--match-head-commit 6930bde`, verified in origin/main history. CI green on the merged HEAD; P-018 Copilot-review gate SATISFIED. Copilot completed three rounds with real findings and a fourth clean round: (1) live SQ probe ignored `INTERCOM_LIVE_CLI_PATH`; stale out-of-order re-issue results; credential panel clipped at small terminal sizes; (2) startup credential delivery could race Ctrl+R; (3) an armed permission approval was not disarmed by the credential shortcuts (a real safety bug). Each was fixed with tests, pushed, replied to, resolved through GraphQL, and review re-requested.
* Pre-PR review: seven persona subagents (Security, Concurrency, Correctness, Go, Maintainability, Architecture, Schema-CLI-Docs) plus Adversarial Review (2 of 3 reviewers responded; the third model returned nothing); no P0 or P1. One fix cycle (828dc6c) with Security, Concurrency and Correctness delta reviews; Copilot-round commits c5737b7, c410f9b were delta-reviewed (Correctness, Concurrency) and 6930bde was inspected directly.
* Dependencies: `charmbracelet/x/term` promoted to a direct dependency; `pflag` promoted by `go mod tidy`. govulncheck clean.
* a1 covering-feature completion: 070-F was moved to done by the task-completion loop before the post-merge gate (a protocol deviation; the move succeeded and the feature archived with the PR); descendants 070.001-T through 070.006-T were done with no other descendants. Cascade classification did not object.
* Close: plain `--classify-only` (CASCADE, engine VERIFIED), then the mutating `autoharness shipment cascade-close` exited 0 with every postcondition passing, under explicit operator-authored pre-approval (dark-mode scope 056-S..062-S; not an admin fallback). Default timeout used, supervised attached to completion.
* Post-close verification: 062-S, 070-F and its tasks are present in `.backlogit/archive/`; no archive deletions (P-007). The reconcile checks were performed inline; no separate report under `.backlogit/reconcile/` (degradation, as in earlier shipments).
* Linked-deliberation disposition: none; nothing archived by Ship.
* Runtime verification at merge time: headless only. The live SDK tests (`INTERCOM_LIVE_SDK_TESTS`), the Dev Tunnel connect checkpoint and live first-responder routing were not executed in the unattended session. Successful archival did not satisfy them. They were run later, in an operator-present session; see "Live evidence update (2026-10-20)".
* Releasability: `READY_WITH_CONDITIONS` (see conditions). No deployment, tag or publish step applies. Rollback is a revert of merge commit 6290151. Owner: repository maintainer. Monitoring: first operator-run live test.

## Live evidence update (2026-10-20)

The three operator-owned conditions were exercised in an operator-present
session after the merge. The evidence is from live runs, not CI; the summary is
in `docs/closure/evidence/062-S-live-verification-evidence.md`. Per-condition
disposition:

* `live-sdk-test-pass`: **partial evidence, not satisfied.** `TestLiveCopilotEndToEnd`
  (including a remotely answered `read` permission, clean shutdown and no
  stranded CLI child) and `TestLiveStartAndStop` passed, and SQ-b is verified.
  `TestLiveSQbSQcReverification` failed as inconclusive: SQ-c is **not**
  re-verified, so design section 4.4 remains an unverified assumption and its
  conservative posture stays in force. ADR MVP-1 also requires a recorded live
  run showing that shell and write calls reach the permission handler; only the
  `read` kind has been exercised, so the condition stays open.
* `devtunnel-connect`: **satisfied for the Go reference client** over a private
  tunnel (devtunnel CLI 1.0.2014; `--allow-anonymous false` accepted; no edge
  token refused; wrong credential gave 401; `Authorization` header reached
  intercom unmodified; subprotocol carriage connected). Not tested: a real
  browser, a second machine, the real TUI-launched binary behind the tunnel,
  `Origin` and `allowed_origins` through the tunnel, long-lived reconnect.
* `tool-approval-routing-live`: **partial evidence, not satisfied.** A `read`
  permission was approved remotely through the tunnel and in the live app test,
  but both runs used headless harnesses with the local TUI disabled, so
  first-responder arbitration across the local TUI and a remote client was not
  exercised live. Shell and write kinds were not exercised live either.

Release readiness stays `READY_WITH_CONDITIONS`: only `devtunnel-connect` is
satisfied (for the reference client), and two conditions remain open with the
partial evidence above. The closure-evidence gate requires every condition to be
satisfied, so it keeps rejecting this artifact with `frontmatter_predicate`
until the remaining live evidence is recorded; that is the accurate state, not
a formatting defect. The open residuals are also recorded in the operator
runbook, sections 5 and 11.

## Source artifact cleanup

* No `source_stash_id` or `source_deliberation_id` in the custom fields of 070-F (only `harness_status`); nothing archived or skipped. Stash A92E3FA0 is the standing roadmap tracker and is deliberately retained.

## Follow-ups (stash, capture-only)

* Deferred scope (P-021 C2): FAE4134F (process-lifecycle hardening), AD8472DD (internal/app maintainability refactors), BDC0ADDF (verification-surface and contract follow-ups).
* Open and not absorbed: 1DCB6108, E9223340, 7EF4DA98 (depguard layering, conformance and drift tests, atomic revalidation), 550437DB, 6738A8E5, 4AFE4DA4, 026CF750 and others. Absorbed by task text: the composed transport constructor and Gate adapter of 7EF4DA98; the cmd wiring of E9223340; wiring-side avoidance of the 550437DB hazard; approval_seconds mapping (same contract surface).

## Degradations

* Per-task telemetry and a separate per-task harness-architect invocation were not run (engineer followed the skill's red-phase protocol and recorded evidence).
* Reconcile was performed inline (no `.backlogit/reconcile` report).
* `golangci-lint` and `govulncheck` required `GOTOOLCHAIN=go1.26.9`.
* Adversarial Review ran with 2 of 3 reviewers.
* `TestProcessTreeSeesAndForgetsAChild` flaked locally on Windows under load (not on Linux CI); stashed in BDC0ADDF.
* Compaction (P-020): done. See `docs/memory/compacted/2026-10-20-062-s-ship-compacted.md`. Compound learnings: `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`.
