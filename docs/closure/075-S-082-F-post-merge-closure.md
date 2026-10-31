---
title: "075-S / 082-F post-merge operational closure"
description: "Copilot cli_path classification and resolution hardening and [remote] config validation (075-S, covering feature 082-F). Code PR 225 merged; the mutating cascade-close awaits per-closure operator approval."
status: in-progress
date: 2026-10-31
mode: post-merge
shipment: 075-S
feature: 082-F
pr: 225
merge_commit_sha: a2898942a12c70fe23be050d698ae4622153d97f
merge_commit: a2898942a12c70fe23be050d698ae4622153d97f
head_sha_reviewed: c019210324abcadbc1750dd3a1f552413aaf6f2b
closure_branch: post-merge/082-f-copilot-cli-path-classification-resolution-hardening-and-remote-config-validation
close_path: cascade
close_evidence: docs/closure/evidence/075-S-082-F-close-evidence.json
closure_status: BLOCKED
releasability: BLOCKED
compaction_status: pending
conditions:
  - id: "per-closure-approval"
    description: "Explicit operator approval for the destructive mutating autoharness shipment cascade-close run for 075-S / 082-F."
    satisfied: false
    disposition: "pending"
    evidence: "Not yet requested of the operator. P-017 dark mode does not waive this approval (Constitution Principle VII)."
  - id: "cascade-close-executed"
    description: "The mutating cascade-close run exits 0 with postcondition_verdict pass and the shipment archived as shipped."
    satisfied: false
    disposition: "pending"
    evidence: "Classify-only recorded the pre_close evidence (CASCADE, engine VERIFIED); the mutating run has not been executed."
  - id: "runtime-verification"
    description: "Runtime verification for the changed surfaces (config validation and the Copilot adapter constructor)."
    satisfied: true
    disposition: "satisfied-automated"
    evidence: "CLI cold-start smoke (go run ./cmd/intercom --help) exit 0 on the merged main; unit coverage for ClassifyCLIPath, validateCLIPath, resolveCLIPath, [remote] unknown-key and hex IPv4 rejection; all 14 CI checks green on PR 225."
---

* State: PR 225 merged, post-merge closure in progress. 082-F is `done` (archived); 075-S is still `active` until the approved mutating cascade-close runs
* Merge: PR 225 merged as a merge commit (P-009) at a289894, reviewed HEAD c019210, verified in `origin/main` history. Normal merge under the dark-mode pre-authorization (`DARK_MODE_MERGE_AUTHORIZED`); no admin fallback
* Scope shipped: shared `pathsafe.ClassifyCLIPath`; `internal/config` `validateCLIPath` and `internal/agent/copilot` `resolveCLIPath` hardened (UNC and device rejection, regular-file check, single `LookPath` resolution with deferred commit); `[remote]` unknown-key rejection, hex IPv4 host rejection, `DefaultRemoteMaxConnections` locked to `hub.DefaultMaxRemote`; design doc 6.2 and 8 and config reference updated
* Task commits: 082.001-T `a416b85`, 082.002-T `f12cfa4`, 082.003-T `e9c3f65`, 082.004-T `4069cd4`, 082.005-T `f1ee9c5`; review follow-ups `c019210`
* CI: all 14 checks PASS on PR 225. P-018 `NOT_APPLICABLE` PASS (Copilot review not engaged)
* Review: Security, Correctness and Maintainability reviewers found no P0 or P1. The Adversarial Review agent could not dispatch sub-reviewers and returned a single-reviewer provisional result; this gap is disclosed in the PR body
* Pre-close reconcile: `.backlogit/reconcile/075-S-pre-done.md`, `PROCEED`, with the `file-lock` token held on `.backlogit/queue/075-S.md`, then released for the approval pause
* Close: classifier CASCADE, `engine_verdict: VERIFIED`, `selected_close_path: cascade`; the mutating run awaits per-closure operator approval

## Source artifact cleanup

* Pending: runs after the cascade-close. Source stash IDs 56CE7077, FD093EFC and FC76C5D8 are to be confirmed archived; the linked-deliberation disposition comes from the cascade-close report

## Follow-ups (stash, capture-only)

* 2C45EFAE: consolidate the duplicated bare-name and absolute-path resolution steps of `validateCLIPath` and `resolveCLIPath` into shared helpers
* 2F1B6D4B: cli_path residual hardening (symlink and mapped-drive canonicalization, PATHEXT, execute bit, `%q` in messages, empty cli_path SDK behavior); requires deliberation

## Degradations

* Compaction (P-020): pending; `compact-context` runs after the cascade-close
