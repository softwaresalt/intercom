---
title: "074-S / 081-F post-merge operational closure"
description: "Write-permission preview safety, shell URL anti-spoofing, and unknown-kind local-TUI authorization (074-S, covering feature 081-F). Code PR 223 merged and the operator-approved cascade-close completed."
status: complete
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
closure_status: READY_WITH_CONDITIONS
releasability: READY_WITH_CONDITIONS
compaction_status: done
conditions:
  - id: "per-closure-approval"
    description: "Explicit operator approval for the destructive mutating autoharness shipment cascade-close run for 074-S / 081-F."
    satisfied: true
    disposition: "satisfied"
    evidence: "Operator replied 'cascade-close command approved' at 2026-10-31T01:24:27Z (2026-10-30T18:24:27.239-07:00) for the 7-record mutating cascade-close command on 074-S / 081-F."
  - id: "cascade-close-executed"
    description: "The mutating cascade-close run exits 0 with postcondition_verdict pass and the shipment archived as shipped."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/074-S-081-F-close-evidence.json (phase post_close): exit 0, postcondition_verdict pass, shipment_record_archived_status shipped, 7 records archived (074-S, 081-F, 081.001-T..081.005-T), returned_ids empty."
  - id: "runtime-verification"
    description: "Runtime verification for changed runtime surfaces (agent, hub, tui)."
    satisfied: true
    disposition: "satisfied-automated"
    evidence: "CLI cold-start smoke (go run ./cmd/intercom --help) exit 0; unit and e2e coverage for live permission.requested, session.hydrate, TUI render, authz matrix and hub-level remote refusal; all 14 CI checks green. Prior live ConPTY and DevTunnel baseline in docs/closure/evidence/072-S-live-verification-evidence.md."
---

* State: merged and closed. 074-S is archived as `shipped`; 081-F and 081.001-T..081.005-T are archived under `.backlogit/archive/`. Cascade-close exited 0 with `postcondition_verdict: pass` and empty `returned_ids`
* Merge: PR 223 merged as a merge commit (P-009) at 966f6a9, reviewed HEAD 7c0a3be, verified in `origin/main` history. Normal merge under the dark-mode pre-authorization (`DARK_MODE_MERGE_AUTHORIZED`); no admin fallback
* Scope shipped: `PreviewClipped` across agent, adapter, hub (`preview_clipped` on `permission.requested` and `session.hydrate`, write kind only) and TUI (column-0 warning, `APPROVE CLIPPED WRITE?` hints); `shellSummary` URL sanitization and `urls:` line indentation; unknown, other and undefined permission kinds local-TUI-only for approval; ADR MVP-3 074-S amendment, design 5.2, 5.3, 6.1, config-reference, `config.toml.example`, operator runbook
* Task commits: 081.001-T `2664cb4`, 081.002-T `9804454`, 081.003-T `f9ff532`, 081.004-T `9289ecd`, 081.005-T `f2c03b6`; review follow-ups `20e1cef`; lint fix `7c0a3be`
* CI: canonical run 41151704527, all 14 checks PASS (first run failed `lint` with unparam in a test helper; fixed in 7c0a3be and re-run green). P-018 `NOT_APPLICABLE` PASS (Copilot review not engaged)
* Review: Security Reviewer, Correctness Reviewer and Adversarial Review (3 reviewers) found no P0 or P1. In-scope findings fixed in 20e1cef; Go Reviewer had no shell and could not run
* Pre-close reconcile: `.backlogit/reconcile/074-S-pre-done.md`, `PROCEED`, with the `file-lock` token held on `.backlogit/queue/074-S.md` (explicit `-WorkspaceRoot` required on this workspace). The lock was re-acquired for the approved mutating run and `mode: post` (`.backlogit/reconcile/074-S-post-done.md`, `PROCEED`), then released
* Close: classifier CASCADE, `engine_verdict: VERIFIED`, `selected_close_path: cascade`; the mutating run was approved per closure by the operator
* Telemetry: `autoharness telemetry begin` was not run for the per-task loop; no telemetry context exists and none is reconstructed

## Source artifact cleanup

* 081-F carries no `custom_fields.source_stash_id`; source stash IDs 773B2CB2 and 70315264 were already archived by Stage, so nothing was re-archived
* Linked-deliberation disposition: `DISPOSITION_COMPLETE` with empty dispositions

## Follow-ups (stash, capture-only)

* ACD578AE: decide whether the hub should refuse remote approval of `PreviewClipped` write permissions (requires deliberation; accepted exposure recorded in the ADR MVP-3 074-S amendment)
* D46C1B38: invisible-rune and homoglyph hardening of `urls:` framing in adapter summaries
* Live ConPTY, DevTunnel and tool-approval-routing manual checkpoints for the new clipped-write modal are not exercised; capture as a Stage follow-up if the operator wants them tracked

## Degradations

* Compaction (P-020): done. `docs/memory/compacted/2026-10-31-074-s-ship-compacted.md`; source memory archived under `docs/archive/memory/2026-10-30/`
* Startup checkpoint: the orchestrator named `checkpoint-20261031-004132.json`, which does not exist. The only active ship checkpoint for 074-S is `checkpoint-20261031-004958.json`; it was not resolved because it was not the operator-selected file
