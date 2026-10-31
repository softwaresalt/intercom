---
title: "077-S / 084-F post-merge operational closure"
description: "Transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting and pumpRejects rate-limiting (077-S, covering feature 084-F). Code PR 230 merged and the approved cascade-close completed on attempt 2."
status: complete
date: 2026-10-31
mode: post-merge
shipment: 077-S
feature: 084-F
pr: 230
merge_commit_sha: 3b1135530989eb8fe60a55fcab768ca61fd3cd74
merge_commit: 3b1135530989eb8fe60a55fcab768ca61fd3cd74
head_sha_reviewed: dc065d1cc57e23b949e1f1bb1560a12a642258dd
closure_branch: post-merge/084-f-transport-wire-authz-hardening
close_path: cascade
close_evidence: docs/closure/evidence/077-S-084-F-close-evidence.json
closure_status: READY_WITH_CONDITIONS
releasability: READY
compaction_status: done
conditions:
  - id: "per-closure-approval"
    description: "Explicit operator approval for the destructive mutating autoharness shipment cascade-close run for 077-S / 084-F."
    satisfied: true
    disposition: "satisfied"
    evidence: "Operator replied 'Approved' at 2026-10-31T18:23:21Z (2026-10-31T11:23:21.048-07:00) approving the hash-verified evidence move-aside, plain classify-only, the mutating cascade-close (attempt 2 of 3) and resuming checkpoint-20261031-181948.json."
  - id: "cascade-close-executed"
    description: "The mutating cascade-close run exits 0 with postcondition_verdict pass and the shipment archived as shipped."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/077-S-084-F-close-evidence.json (phase post_close): exit 0, postcondition_verdict pass, shipment_record_archived_status shipped, 10 records archived (077-S, 084-F, 084.001-T..084.008-T), returned_ids empty, parent_id_preserved true, linked_deliberation_disposition complete (no dispositions). Attempt 1 (exit 6) preserved at docs/closure/evidence/077-S-084-F-close-evidence.attempt1-exit6.json."
  - id: "runtime-verification"
    description: "Runtime verification for the changed surfaces (transport listener, wire dispatch, hub connect validation, wsclient URL checks)."
    satisfied: true
    disposition: "satisfied"
    evidence: "CLI cold-start smoke (go run ./cmd/intercom --help) exit 0 on the post-merge branch; unit and integration tests for the transport, hub and wsclient packages passed locally (go vet, go build, gofmt, go test ./..., go test -race) and in PR 230 CI."
---

* State: PR 230 merged and 077-S closed. 084-F and its eight tasks are archived; the shipment record is archived with `archived_status: shipped`
* Merge: PR 230 merged as a merge commit (P-009) at `3b1135530989eb8fe60a55fcab768ca61fd3cd74`, reviewed HEAD `dc065d1`, verified in `origin/main` history. Normal merge under the dark-mode pre-authorization; no admin fallback
* Scope shipped: atomic hub-step credential revalidation through `ConnectOptions.Validate`; wire feedback bounds, local-only refusals and close on hub-step failure; off-accept-loop over-cap 503 refusal worker with listener `OnReject`; `RecordReject` accounting and handshake validation; depguard layering; wsclient URL fragment, opaque and host checks; rate-limited `pumpRejects`; ADR MVP-2 and runbook updates
* Task commits: `5107b42`, `24755a8`, `a849ab7`, `c799270`, `84dcf99`, `e6c3f53`, `82a11d2`, `616f07a`; review fixes `dc065d1`
* Review: no P0 or P1; outcome `READY_WITH_FOLLOWUPS`
* Pre-close reconcile: `.backlogit/reconcile/077-S-pre-20261031T084500Z.md`, `PROCEED`
* Close: classifier CASCADE, `engine_verdict: VERIFIED`, `selected_close_path: cascade`. Attempt 1 exited 6 after evidence blockers on 084.006-T and 084.007-T; those were repaired with `backlogit shipment repair-evidence` and the operator approved attempt 2, which exited 0 with `postcondition_verdict: pass`
* Post-archive reconcile: `.backlogit/reconcile/077-S-post-2026-10-31T18-30-00Z.md`, `PROCEED`; P-007 check found no archive deletions; lock `.backlogit/queue/.077-S.md.lock` released and removed
* Recovery: `checkpoint-20261031-181948.json` restored on operator confirmation and resolved after successful resume

## Source artifact cleanup

* 084-F and 077-S carry no `custom_fields.source_stash_id`; no stash entries were archived by Ship
* `source_deliberation_id`: none present; no linked-deliberation dispositions (`DISPOSITION_COMPLETE`, empty); nothing archived by Ship

## Follow-ups (stash, capture-only, P-021 C2)

* 7919A9A1: wsclient still imports `internal/hub`; decide whether to invert the dependency (captured during the code review)

## Compaction (P-020)

* Status: done. `docs/memory/compacted/2026-10-31-077-s-ship-compacted.md` consolidates the 077-S Ship session and closure memory; both sources moved to `docs/archive/memory/2026-10-31/`
