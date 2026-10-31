---
title: "074-S / 081-F compacted Ship handoff"
description: "Consolidated Ship execution and closure record for write-permission preview safety, shell URL anti-spoofing and unknown-kind local-TUI authorization"
date: 2026-10-31
status: shipped
tags: [memory, compacted, permission-preview, 074-S]
source:
  - docs/archive/memory/2026-10-30/074-s-ship-memory.md
---

## Execution

* Intake: index sync, queued-with-active-work check, pre_claim topology gate, intake reconcile, claim and post_claim global verification (`CLAIM_VERIFY_OK`)
* 081.001-T to 081.005-T built test-first (red confirmed, then green), each committed and moved to `done`
* Review gate: Security, Correctness and Adversarial reviewers found no P0 or P1. In-scope findings fixed in `20e1cef`. Two out-of-scope findings captured as stash ACD578AE and D46C1B38 (P-021 C2)
* PR 223 merged by merge commit `966f6a977ef6590e1d7a9155c3f64d0ee8861061` under dark-mode pre-authorization. CI first failed `lint` (unparam), fixed in `7c0a3be`, then all 14 checks passed. P-018 was NOT_APPLICABLE

## Decisions

* Clipped writes stay remotely approvable when `remote_approve_destructive = true`. This is recorded as accepted exposure in ADR MVP-3 (074-S) and deferred to stash ACD578AE, because new remote policy exceeds the authorized change (P-021 C1)
* `PreviewClipped` is honoured only for write-kind asks in the hub
* Run `golangci-lint run ./...` locally before push
* The `file-lock` scripts need `-WorkspaceRoot` on this workspace

## Closure

* Step a1 moved 081-F to `done`; classify-only selected CASCADE (engine VERIFIED)
* Operator approved the mutating cascade-close at 2026-10-31T01:24:27Z. The run exited 0 with `postcondition_verdict: pass`, `shipment_record_archived_status: shipped`, seven records archived and `returned_ids` empty
* Post-mode reconcile: `.backlogit/reconcile/074-S-post-done.md`, `PROCEED`. Evidence: `docs/closure/evidence/074-S-081-F-close-evidence.json`
* Hold 075-S until the closure PR merges. The Stage memory `stage-write-preview-and-cli-path-safety-memory.md` also covers 075-S and stays in place
