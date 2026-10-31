---
title: "076-S / 083-F post-merge operational closure"
description: "Remote clipped-write approval refusal, shell/write summary invisible-rune and homoglyph anti-spoofing, and shared TUI modal header continuation helper (076-S, covering feature 083-F). Code PR 228 merged and the approved cascade-close completed."
status: complete
date: 2026-10-31
mode: post-merge
shipment: 076-S
feature: 083-F
pr: 228
merge_commit_sha: e15937714e489f6c55485651029db7cee031b91f
merge_commit: e15937714e489f6c55485651029db7cee031b91f
head_sha_reviewed: cf7aa865b96b5ab02ada869c9c95926f5c89d3c3
closure_branch: post-merge/083-f-remote-clipped-write-approval-refusal
close_path: cascade
close_evidence: docs/closure/evidence/076-S-083-F-close-evidence.json
closure_status: READY_WITH_CONDITIONS
releasability: READY
compaction_status: done
conditions:
  - id: "per-closure-approval"
    description: "Explicit operator approval for the destructive mutating autoharness shipment cascade-close run for 076-S / 083-F."
    satisfied: true
    disposition: "satisfied"
    evidence: "Operator replied 'Approved.' at 2026-10-31T07:20:00Z (2026-10-31T00:20:00.275-07:00) for the mutating cascade-close on 076-S / 083-F at e15937714e489f6c55485651029db7cee031b91f, and approved resuming checkpoint-20261031-071316.json."
  - id: "cascade-close-executed"
    description: "The mutating cascade-close run exits 0 with postcondition_verdict pass and the shipment archived as shipped."
    satisfied: true
    disposition: "satisfied"
    evidence: "docs/closure/evidence/076-S-083-F-close-evidence.json (phase post_close): exit 0, postcondition_verdict pass, shipment_record_archived_status shipped, 6 records archived (076-S, 083-F, 083.001-T..083.004-T), returned_ids empty, parent_id_preserved true, linked_deliberation_disposition complete (no dispositions)."
  - id: "runtime-verification"
    description: "Runtime verification for the changed surfaces (hub permission authorization, copilot adapter summary sanitization, TUI modal headers)."
    satisfied: true
    disposition: "satisfied"
    evidence: "CLI cold-start smoke (go run ./cmd/intercom --help) exit 0 on the post-merge branch; unit and integration tests for hub.authorize/resolvePermission, internal/agent/copilot sanitizers and internal/tui modal header helper passed locally (go vet, go build, gofmt, go test ./..., go test -race) and in PR 228 CI. No DevTunnel wire protocol change."
---

* State: PR 228 merged and 076-S closed. 083-F and its four tasks are archived; the shipment record is archived with `archived_status: shipped`
* Merge: PR 228 merged as a merge commit (P-009) at e159377 (`e15937714e489f6c55485651029db7cee031b91f`), reviewed HEAD cf7aa86, verified in `origin/main` history. Normal merge under the dark-mode pre-authorization; no admin fallback
* Scope shipped: `hub.authorize` refuses remote approval of PreviewClipped write permissions (`ReasonDestructiveRemoteRefused`, even with `remote_approve_destructive = true`) while preserving local-TUI approval and remote deny; `Hub.resolvePermission` clears and bounds deny feedback; copilot adapter sanitizers hardened against invisible filler runes and `urls:` homoglyphs; shared `modalHeader` / `renderModalHeaders` helper with the column-0 header invariant; ADR MVP-3, backend architecture and config reference updated
* Task commits: `318167f` (hub refusal), `54ab92b` (summary sanitization), `a5f0ee0` (TUI modal header), `c9600b3` (docs); state commits `90c6561`, `46149a0`, `cf7aa86`
* Review: Security and Correctness reviewers found no P0 or P1; outcome `READY_WITH_FOLLOWUPS`
* Pre-close reconcile: `.backlogit/reconcile/076-S-pre-20261031T071245Z.md`, `PROCEED`
* Close: classifier CASCADE, `engine_verdict: VERIFIED`, `selected_close_path: cascade`. Mutating run approved by the operator, exit 0, `postcondition_verdict: pass`, `returned_ids` empty, `parent_id_preserved: true`
* Post-archive reconcile: `.backlogit/reconcile/076-S-post-2026-10-31T07-25-00Z.md`, `PROCEED`; P-007 check found no archive deletions; lock `.backlogit/queue/.076-S.md.lock` released and removed
* Recovery: `checkpoint-20261031-071316.json` restored on operator confirmation and resolved after successful resume

## Source artifact cleanup

* 083-F carries no `custom_fields.source_stash_id`; the Stage-owned source stashes (ACD578AE, D46C1B38, 7BDC38FF, 026CF750) were not archived by Ship
* `source_deliberation_id`: no linked-deliberation dispositions in the close report (`DISPOSITION_COMPLETE`, empty); nothing archived by Ship

## Follow-ups (stash, capture-only, P-021 C2)

* D3D0E9D5: replace the hand-written `urls:` homoglyph denylist with NFKC plus confusable-skeleton folding and tolerate interior invisible runes
* 2EE9F3ED: sanitize operator deny feedback (control, bidi, newline, invalid UTF-8) and add a non-positive-n guard to `truncUTF8`
* F24EC587: decide whether summary sanitization should preserve emoji variation selectors, ZWJ/ZWNJ and private-use icon glyphs
* 077-S (084-F) remains queued; the Orchestrator routes it after this closure PR merges

## Compaction (P-020)

* Status: done. `docs/memory/compacted/2026-10-31-076-s-ship-compacted.md` consolidates the 076-S Ship session and closure memory; both sources moved to `docs/archive/memory/2026-10-30/`
