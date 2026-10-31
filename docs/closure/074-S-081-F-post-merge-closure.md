---
title: "074-S / 081-F post-merge operational closure"
description: "BLOCKED pending per-closure approval. Write-permission preview safety, shell URL anti-spoofing, and unknown-kind local-TUI authorization (074-S, covering feature 081-F). Code PR 223 merged; the mutating cascade-close has not run and awaits explicit per-closure operator approval."
status: in-progress
date: 2026-10-30
mode: post-merge
shipment: 074-S
feature: 081-F
pr: 223
merge_commit_sha: 966f6a977ef6590e1d7a9155c3f64d0ee8861061
merge_commit: 966f6a977ef6590e1d7a9155c3f64d0ee8861061
head_sha_reviewed: 7c0a3be184de4c5dba2c92978a2614d9f77d5081
closure_branch: post-merge/081-f-write-permission-preview-safety-shell-url-anti-spoofing-and-unknown-kind-local-tui-authorization
close_path: cascade
close_evidence: docs/closure/evidence/074-S-081-F-close-evidence.json
closure_status: BLOCKED
releasability: READY_WITH_CONDITIONS
compaction_status: pending
conditions:
  - id: "per-closure-approval"
    description: "Explicit operator approval for the destructive mutating autoharness shipment cascade-close run for 074-S / 081-F."
    satisfied: false
    disposition: "pending"
    evidence: "DARK_MODE_ACTIVE does not waive per-closure approval (Constitution Principle VII, P-017 rule 8). Ship stopped before the mutating run."
  - id: "cascade-close-executed"
    description: "The mutating cascade-close run exits 0 with postcondition_verdict pass and the shipment archived as shipped."
    satisfied: false
    disposition: "pending"
    evidence: "Read-only classification only: docs/closure/evidence/074-S-081-F-close-evidence.json (phase pre_close, classifier CASCADE, engine VERIFIED, selected_close_path cascade, mutation_possible no)."
  - id: "runtime-verification"
    description: "Runtime verification for changed runtime surfaces (agent, hub, tui)."
    satisfied: true
    disposition: "satisfied-automated"
    evidence: "CLI cold-start smoke (go run ./cmd/intercom --help) exit 0; unit and e2e coverage for live permission.requested, session.hydrate, TUI render, authz matrix and hub-level remote refusal; all 14 CI checks green. Live ConPTY, DevTunnel and tool-approval-routing manual checkpoints were not run in this session and remain a release condition."
---

* State: merged, not yet closed. 074-S is `active`; 081-F was moved `active -> done` at Step 6.1.a1 S4 (archived by the backlog tool); 081.001-T..081.005-T are `done` in `.backlogit/archive/`
* Merge: PR 223 merged as a merge commit (P-009) at 966f6a9, reviewed HEAD 7c0a3be, verified in `origin/main` history. Normal merge under the dark-mode pre-authorization (`DARK_MODE_MERGE_AUTHORIZED`); no admin fallback
* Scope shipped: `PreviewClipped` across agent, adapter, hub (`preview_clipped` on `permission.requested` and `session.hydrate`, write kind only) and TUI (column-0 warning, `APPROVE CLIPPED WRITE?` hints); `shellSummary` URL sanitization and `urls:` line indentation; unknown, other and undefined permission kinds local-TUI-only for approval; ADR MVP-3 074-S amendment, design 5.2, 5.3, 6.1, config-reference, `config.toml.example`, operator runbook
* Task commits: 081.001-T `2664cb4`, 081.002-T `9804454`, 081.003-T `f9ff532`, 081.004-T `9289ecd`, 081.005-T `f2c03b6`; review follow-ups `20e1cef`; lint fix `7c0a3be`
* CI: canonical run 41151704527, all 14 checks PASS (first run failed `lint` with unparam in a test helper; fixed in 7c0a3be and re-run green). P-018 `NOT_APPLICABLE` PASS (Copilot review not engaged)
* Review: Security Reviewer, Correctness Reviewer and Adversarial Review (3 reviewers) found no P0 or P1. In-scope findings fixed in 20e1cef; Go Reviewer had no shell and could not run
* Pre-close reconcile: `.backlogit/reconcile/074-S-pre-done.md`, `PROCEED`, with the `file-lock` token held on `.backlogit/queue/074-S.md` (explicit `-WorkspaceRoot` required on this workspace). The lock was released cleanly for the approval pause and must be re-acquired before the approved mutating run and `mode: post`
* Close classification (read-only): `selected_close_path: cascade`, classifier CASCADE, `engine_verdict: VERIFIED`, `mutation_possible: no`
* Telemetry: `autoharness telemetry begin` was not run for the per-task loop; no telemetry context exists and none is reconstructed

## Source artifact cleanup

* Pending until after the cascade close. 081-F carries no `custom_fields.source_stash_id`; source stash IDs 773B2CB2 and 70315264 must be confirmed archived by Stage or the Ship post-merge Step 7 pass
* Linked-deliberation disposition: pending the `shipment-reconcile` INV-12 step in the mutating close

## Follow-ups (stash, capture-only)

* ACD578AE: decide whether the hub should refuse remote approval of `PreviewClipped` write permissions (requires deliberation; accepted exposure recorded in the ADR MVP-3 074-S amendment)
* D46C1B38: invisible-rune and homoglyph hardening of `urls:` framing in adapter summaries
* Live ConPTY, DevTunnel and tool-approval-routing manual checkpoints for the new clipped-write modal are not exercised; capture as a Stage follow-up if the operator wants them tracked

## Degradations

* Compaction (P-020): pending. `compact-context` runs after the cascade close; until then `compaction_status` stays `pending` and the next shipment (075-S) must stay held
