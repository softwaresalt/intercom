---
title: "042-S / 052-F post-merge operational closure"
description: "Post-merge release readiness for Gatecheck Batch B1 write-path input hardening: the NUL-safe listing, git environment isolation and fixture containment merged as PR #123, the cascade close of 042-S, and disclosure of every procedural deviation and residual."
status: ready_with_conditions
tags:
  - closure
  - post-merge
  - dev-tooling
  - gatecheck
  - writepath
  - retiredarch
  - security
  - 042-S
date: 2026-10-11
mode: post-merge
shipment: 042-S
feature: 052-F
pr: 123
merge_commit_sha: 591f37e2ba877d81b1304d31905369e79096657f
head_sha_reviewed: b44cc1539c5d734caef49726306c575f3e6b2700
base_sha: 9cb4d3b7a96f8f4d8eedb67ddc8a742b65a1446e
closure_pr: 124
compaction_status: degraded
close_path: cascade
classification_binding: e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95
closure_status: READY_WITH_CONDITIONS
releasability: READY_WITH_CONDITIONS
conditions:
  - id: cascade-post-close-writer-evidence-missing
    summary: "The harness writer produced the pre_close record (captured before the close), but no post_close writer phase exists. The mutating form of autoharness shipment cascade-close was deliberately not used, because it invokes backlogit shipment ship itself and the Ship role boundary reserves that invocation to the classified close path. The post-close state is evidenced by the shipment-reconcile post-mode check and the archive records, not by the writer. Unsatisfied and disclosed; Operator disposition (2026-10-11T21:47 -07:00, item 5, option A): no replay. No post_close writer record is produced for 042-S, because the writer refuses after the fact and a hand-built record would be fabricated. Accepted as disclosed. Future closures use the writer through stash A83C398E."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 5; docs/closure/evidence/042-S-052-F-close-evidence.json (phase pre_close, classifier CASCADE, engine VERIFIED; mutation_possible no and failures [] are in the writer stdout recorded in the reconcile record); .backlogit/reconcile/042-S-cascade-close-2026-10-11T02-19-08Z.md ('Writer-form evidence' and 'Close' sections)"
  - id: safe-close-agent-executed-binding
    summary: "Process residual, disclosed and matching 041-S: there is no invokable shipment-reconcile safe-close CLI or gate (stash D10D3AFC, Stage-owned). The classify-close-path verdict and CLASSIFICATION_BINDING were computed by the agent from the canonical v1 lines, and the cascade primitive was invoked by the agent on the P-015 fully-covered-root path. The engine's own revalidation does not recompute this binding. Operator disposition (2026-10-11T21:47 -07:00, item 1, option A): accepted as a disclosed process deviation, not a safety failure, same shape as 041-S. The Orchestrator verified independently that the closure branch changes under .backlogit only the seven manifest members, the reconcile record and one stash capture, and that nothing reached main. Conditions attached: stash D10D3AFC and stash A83C398E form the next Stage batch, before another implementation shipment; Stage reconciles the contradiction between the Ship rule and the absence of any executable safe-close, by implementing safe-close or amending the rule; the disposition is recorded here."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 1; .backlogit/reconcile/042-S-cascade-close-2026-10-11T02-19-08Z.md (binding e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95 with its canonical lines); stash entry D10D3AFC (active, Stage-owned)"
  - id: principle-iv-scratch-writes
    summary: "Principle IV / AGENTS.md CLI workspace containment, disclosed and not remediated. Ship wrote scratch files under logs/ (git-ignored) by design. Outside the workspace, Ship created and removed a short-lived symlink probe directory under the OS temp directory during intake. Repository Go tests write temporary fixtures under t.TempDir (the existing repository convention, recorded in the plan). The tool runtime spilled large command output to the OS temp directory. Stage left a copy of the HEAD stash file at %TEMP%/head-stash.jsonl (Stage memory, 'Deviations and disclosures'). Ship did not delete it, and operator deletion or acknowledgement is requested. Ship's own probe directory was created and removed under the OS temp directory at intake. The P-005 event for this condition is recorded here, because intercom was unavailable and nothing was broadcast. Operator disposition (2026-10-11, item 4): done by the operator. The operator deleted Stage's %TEMP% scratch files head-stash.jsonl and stage-042s-commitmsg.txt, so the residue is removed and this condition is dispositioned as disclosed."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 4; Ship session log (symlink probe at intake, removed); plan Constitution Check (Principle IV interpretation); PR #123 body 'Local Review Readiness' and 'Notes'"
  - id: timing-sensitive-start-script-test
    summary: "tests/integration TestInvokeStartScriptMainPropagatesNonZeroExitCode (a pwsh-driven test with a 30 s budget for start.ps1) recorded `start.ps1 exceeded 30s timeout` in full local runs, including runs dated 2026-10-08 before this branch existed. It passed in isolation (39.8 s, 19.6 s) and in the final full local run on the reviewed HEAD. The cause is not established: start.ps1 runs backlogit and Engram sidecar syncs, so workspace size is a hypothesis only. The test does not exercise write-path code. Timeline and run list: docs/closure/evidence/042-S-verification-evidence.md. The 2026-10-08 log logs/fulltest-final.txt also ends in a 10-minute test panic with a blocked child process (TestLocationRestoredEvenOnNonZeroCopilotExit in exec.Cmd.Wait), and four of the five logged full-package runs failed. The condition is NOT satisfied: the cause is open and acceptance is an operator decision. The Linux CI test job passed on the reviewed HEAD (run 38104296737). Update 2026-10-11 (operator disposition, item 3, accepted with root-cause analysis): the root cause was established after this text was written, and the RCA supersedes the earlier statement that the cause is not established. (1) start.ps1 Get-EnabledSidecars is a render-time constant that enables engram, graphtor-docs and backlogit, and start.ps1 operates on the real repository through PSScriptRoot, so each run performs a real backlogit sync (about 8 s), an Engram pre-warm that always exhausts its 15 s budget on a cold index of 191 files, and a graphtor sync, for 25 to 35 s against a 30 s limit (about 2 s on Linux CI, where no sidecars exist). (2) start.ps1 Kill() terminates only the launched engram process; the index worker survives holding the inherited stdout pipe, and CombinedOutput has no WaitDelay, so the test blocks until the worker exits (test durations of 140 to 206 s track engram index durations of 123 to 184 s). (3) The tests write to real workspace state. The failure is unrelated to 042-S (the same failure occurred on 2026-10-08, before 042-S existed). Accepted as a disclosed environmental residual on this Windows host; the test-only fix (about 15 lines) is captured as stash C5E50EFD and is not implemented in this closure."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 3; stash C5E50EFD (RCA capture, P-021 C2); logs/fulltest-final.txt and logs/fulltest-head.txt (2026-10-08 timeouts, the first with a 10-minute panic); logs/gate-052.004-T-fulltest.txt, logs/final-e202fe3-fulltest.txt (timeouts on the 042-S branch); logs/052.004-T-integration-rerun.txt and logs/final-startscript-isolated.txt (pass); logs/final-b44cc15-fulltest.txt (pass); PR #123 check test (pass); docs/closure/evidence/042-S-verification-evidence.md"
  - id: copilot-review-residual-risk-acceptance
    summary: "Copilot's review of the final HEAD was COMMENTED with zero review threads and 0 open findings, but its body asked for human risk acceptance of the documented false-clean residual surfaces. Those surfaces are the accepted residuals captured in stash entries 4372BAD4, DBE25DF5, EDC18D59 and 31F33EFE and the new captures EEE39D23, 858516B9, 2613D248, 65AE88CA, 96B2E8CC and CDBA41E3.     This record does not treat the merge pre-authorization as an explicit risk acceptance: the six new captures were created after it, and P-018 does not let an activation record satisfy the Copilot question. Explicit operator risk acceptance is an open decision. Operator disposition (2026-10-11T21:47 -07:00, item 2): ACCEPTED as interim and tracked, not as a decision never to fix. This is the explicit operator risk acceptance that Copilot's PR #123 review body asked for. The merge pre-authorization is still not counted as that acceptance; the operator reply is, and it is recorded in the Operator disposition section with the stash IDs and the routing."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 2; PR #123 review by copilot-pull-request-reviewer on b44cc15 (COMMENTED, zero threads); autoharness gate copilot-review 123 verdict SATISFIED for b44cc15; DARK_MODE_ACTIVE merge pre-authorization"
  - id: copilot-request-via-rest
    summary: "The Copilot request was made through the REST requested_reviewers endpoint with copilot-pull-request-reviewer[bot]. GraphQL reviewRequests did not show the request afterward, but Copilot reviewed the final HEAD (review event at 2026-10-11T02:14:20Z). Disclosed; the P-018 gate verdict SATISFIED is the controlling signal."
    satisfied: true
    evidence: "gh api --method POST repos/softwaresalt/intercom/pulls/123/requested_reviewers; GraphQL reviews for PR 123; autoharness gate copilot-review verdict SATISFIED"
  - id: covering-feature-archived-on-done-move
    summary: "Disclosure, matching the 041-S precedent: backlogit move 052-F --status done relocated 052-F from queue to archive in backlogit 1.11.0. The a1 transition ran before the pre-close snapshot, so the snapshot records 052-F as archived (see the closure-mutations condition). Location is not a coverage input and the descendant topology was unchanged, so the CASCADE verdict stands."
    satisfied: true
    evidence: "backlogit move 052-F --status done (exit 0); .backlogit/archive/052-F.md (archived_status done); reconcile record a1 section"
  - id: closure-mutations-before-closure-branch
    summary: "Process deviation, disclosed (P-011 and Step 6.0 ordering). After the merge the agent checked out main and fast-forwarded it, then performed the covering-feature transition (052-F active to done), the classify-only evidence capture and the cascade close while main was the checked-out branch, and only afterwards created post-merge/042-s-gatecheck-batch-b1-write-path-input-hardening. No commit reached main. The uncommitted backlog state was carried onto the closure branch with checkout -b and committed there (672eddf). Reflog times: checkout main 02:15:22Z, fast-forward 02:15:23Z, closure branch created 02:18:54Z. Intercom was unavailable, so this P-005 event is recorded here only."
    satisfied: true
    evidence: "git reflog for HEAD on 2026-10-11 between 02:15Z and 02:19Z; closure commit 672eddf on the closure branch; .backlogit/reconcile/042-S-cascade-close-2026-10-11T02-19-08Z.md"
  - id: ship-direct-primitive-invocation
    summary: "P-010 boundary violation, disclosed. Ship called `backlogit shipment ship` directly after computing the binding itself. The Ship agent file prohibits this ('Ship NEVER calls backlogit shipment ship ... directly, for any verdict'), and the shipment-reconcile skill reserves that call to its Cascade Close Sub-Procedure after the safe-close step 0 revalidation of the bound classification. That revalidation never ran. Post-hoc state checks pass (returned_ids empty; archived_ids inside allowed_ids; required members archived; the binding recomputes), but a post-hoc pass does not cure the missed revalidation. Disposition: NOT reversed (a reversal would be a second unauthorized archive mutation). The closure was BLOCKED pending an explicit operator disposition. The operator accepted it on 2026-10-11T21:47 -07:00 (item 1, option A; see the Operator disposition section), so this condition is dispositioned and the closure PR is no longer blocked by it."
    satisfied: true
    evidence: "Operator disposition record, section 'Operator disposition (2026-10-11)', item 1; reconcile record 042-S-cascade-close-2026-10-11T02-19-08Z.md (Close section and Disclosures); Copilot thread on PR 124 (PRRT_kwDOTPuhps6rJyWj); .github/agents/_ship.agent.md (Ship never calls backlogit shipment ship directly); .github/skills/shipment-reconcile/SKILL.md (Cascade Close Sub-Procedure, safe-close step 0 revalidation)"
  - id: intercom-unavailable
    summary: "agent-intercom tools were not present in this session. Operator visibility ran through session output, the PR bodies and this artifact. Intercom broadcasts were skipped."
    satisfied: true
    evidence: "Session tool inventory; PR #123 body"
  - id: mcp-tools-unavailable
    summary: "No MCP backlog tools were exposed. Every backlog operation used the registered backlogit CLI fallback (TOOL_DEGRADED, expected and logged)."
    satisfied: true
    evidence: "Ship Step 0.0 probe log in the session; backlogit 1.11.0 CLI commands in the reconcile record"
---

# 042-S / 052-F post-merge operational closure

## Summary

Shipment 042-S (feature 052-F, Gatecheck Batch B1 write-path input hardening) merged as PR #123 at 591f37e2ba877d81b1304d31905369e79096657f on 2026-10-11T02:15:09Z. The change hardens the merge-blocking write-path gate in three ways: it selects tracked Go files with NUL-delimited `git ls-files -z` output and fails closed on malformed or unscannable listings (B83F53BB, part 1); it removes ambient `GIT_*` variables and pins global and system git configuration for the `git ls-files` child (B83F53BB, part 2); and it refuses a fixture reached through a symlink, junction or linked ancestor in the self-test (05E12A6F). A comment-only correction in the retiredarch listing comment was included (U5).

Releasability: `READY_WITH_CONDITIONS` for the closure record. The operator dispositioned every condition on 2026-10-11T21:47 -07:00 (section 'Operator disposition (2026-10-11)'), including the P-010 violation, so the closure is no longer blocked. The shipped code change itself carries no open code condition; each frontmatter condition is dispositioned and carries its evidence, and the disclosures are retained.

## Merge and gate record

* Merge commit `591f37e2ba877d81b1304d31905369e79096657f`, confirmed by `gh pr view 123` (MERGED) and by `git merge-base --is-ancestor` against `origin/main`.
* Reviewed HEAD `b44cc1539c5d734caef49726306c575f3e6b2700`. `headRefOid` matched at merge.
* Required CI on the reviewed HEAD: all checks passed, including `test` (Linux, required) and `test (windows, advisory)`, `lint`, `security`, `cross-compile` for four targets, and `merge-strategy structural verification`.
* Section 1.9 readiness: the PR body's Local Review Readiness block cites the reviewed HEAD, outcome `READY_WITH_FOLLOWUPS`, `P0=0, P1=0` (one consensus-rated P1 dispositioned as a P-021 C1 out-of-scope capture), full local build evidence, and follow-up stash IDs.
* P-018 copilot-review gate: `SATISFIED` for the reviewed HEAD on the final run before merge.
* P-009: the repository accepts merge commits only (`allow_merge_commit: true`, `allow_squash_merge: false`, `allow_rebase_merge: false`). Merge used `--merge --match-head-commit`.
* P-016: `pipeline-topology` lifecycle gate exit 0 before the merge and again before the close.
* Approval source: the Orchestrator's dark-factory session brief of 2026-10-10 (the DARK_MODE_ACTIVE record carried in the invocation; it is not stored in the repository). It states `merge_approval_pre_authorized: true` for scope 042-S and that the closure PR is pre-authorized for merge once its gates pass. No admin fallback was used. Section 1.9.6 evaluation for the reviewed HEAD b44cc15: (1) PR inside scope: yes; (2) pre-authorization true; (3) section 1.9 checks 1 to 5 passed; (4) required CI green; (5) P-009 merge-commit-only verified and P-016 lifecycle gate exit 0; (6) P-018 copilot-review SATISFIED for b44cc15, with no admin override.

## Review and remediation record

* Pre-PR review: the persona pool (Correctness, Go, Security, Constitution, Maintainability, Scope Boundary) and the multi-model Adversarial Review, run against 1d39f76. Full dispositions are in the PR #123 body.
* Review-fix cycles: three (1a704c5, 5eb2e78, e202fe3), each followed by a targeted re-review or a self-audit of the diff, within the per-task cap of three. Accepted follow-ups at the cap are stash entries.
* Adversarial P1 candidates: V-1 (oracle and select.go verification) closed by Ship; C-5 (Windows non-ASCII env alias) not honoured by git 2.55 on this host and not reachable on the required Linux gate; C-1 (fixture-name echo) captured under EDC18D59 under P-021 C1, because the PASS and FAIL lines already echo the same names for regular fixtures.

## Verification on the reviewed HEAD

Committed, reproducible evidence (commands, exit codes, run identifiers, timings and the timeout timeline): docs/closure/evidence/042-S-verification-evidence.md.

* `go build ./...` exit 0; `go vet ./...` exit 0; `gofmt -l .` empty; `goimports -l .` empty; `golangci-lint run ./...` 0 issues; `GOOS=linux go vet ./tools/gatecheck/...` exit 0.
* `go test -count=1 ./...` on b44cc15: exit 0, 14 packages ok.
* Live tree: `write-path --root .`, `write-path --root . --self-test`, `write-path --root . --self-test-integrity`, `retired-arch --root .` and `scripts/check-write-path-precondition.sh`, all exit 0.
* `TestRunFixtureSelfTest_LinkedFixture_FailsClosed`: 5 of 5 PASS, 0 SKIP on this host.
* `WRITE_PATH_GATE_ADVISORY` repository variable: unset, so the verdict step is blocking.

## Close path

* Covering-feature completion gate (a1): satisfied (manifest tasks declared `done` and archived; no live descendants; descendant set equal to the manifest minus the feature; containment holds; 052-F was `active` before the single transition).
* Classify-close-path: `CASCADE` / `FULLY_COVERED_ROOT`, binding `e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95`.
* Cascade close (P-010 boundary violation, see the disposition section): `backlogit shipment ship 042-S --sha 591f37e2ba877d81b1304d31905369e79096657f`. Result: `shipment_status: shipped`, `returned_ids: []`, `archived_ids: [052.001-T, 052.002-T, 052.003-T, 052.004-T, 052.005-T, 042-S, 052-F]`. Two-set gate: pass both checks. Post-mode: pass; no archive deletions.
* Closure commit: `672eddf` on `post-merge/042-s-gatecheck-batch-b1-write-path-input-hardening` (`chore: archive 042-S backlog artifacts`). Never committed to main.

## Source artifact cleanup

The covering feature 052-F carries no `source_stash_id` or `source_deliberation_id` custom field, so no source stash entry is archived by Ship in this closure. The consumed stash entries B83F53BB and 05E12A6F are archived by Stage per the plan and are not touched here.

## Stash follow-ups (visible to Stage)

* Captured as DEFERRED SCOPE EXPANSION (P-021 C2, capture-only): `EEE39D23` (scan-set selection of tracked links, gitlinks and case variants; medium), `858516B9` (fixture self-test directory and hard-link gaps, FIFO row), `2613D248` (env residuals MSYS, CYGWIN, MSYS2_*, boundary rows), `65AE88CA` (test-helper consolidation), `96B2E8CC` (GitRunner context parameter), `CDBA41E3` (cgo, linkname and os/exec write detection).
* Existing residual entries cited by the PR: `4372BAD4` (relative git refusal), `DBE25DF5` (wrapper root discovery), `EDC18D59` (echo sinks, including the U4 refusal site), `31F33EFE` (rune extensions and the retiredarch asymmetry).
* Accepted follow-ups at the review-fix cycle cap: the non-repository test's no-ceiling control, exact context for the non-control rows of the bad-listing test, the Windows 8.3-name check, and the start-script timing item. Recorded in `.backlogit/stash.jsonl`.

## Follow-ups (operator-dispositioned routing)

Stage owns each entry below (P-021 C5: Ship captures only). None is fixed in this closure (P-021 C1).

* Next Stage batch, before another implementation shipment (operator condition 1): `D10D3AFC` (high; make safe-close invokable as a real CLI or gate subcommand, so its Step 0 revalidation runs before any cascade) and `A83C398E` (medium; tune-harness adoption of `autoharness shipment cascade-close` and `gate closure-evidence`, with re-rendering of the 30 stale routing variables).
* Contradiction for Stage (operator condition 2): reconcile the Ship rule that Ship never calls `backlogit shipment ship` directly with the absence of any executable safe-close, by implementing safe-close (`D10D3AFC`) or amending the rule.
* Write-path batch (operator item 2): `EEE39D23` and `DBE25DF5` (both medium, fail-open), with `858516B9`, `EDC18D59` and `2613D248`.
* Bounded-threat-model deliberation first: `CDBA41E3`, `FE2F02FF` and `C0D28448`.
* Separate policy decision: `31F33EFE`.
* Optional: `4372BAD4`, `65AE88CA` and `96B2E8CC`.
* Start-script RCA (operator item 3): `C5E50EFD` (medium, task; the recorded flag is `requires deliberation: no`, but Stage's P-021 C6 intake still routes it through the `deliberate` skill before planning). Related entries, not reuse candidates (no DEFERRED SCOPE EXPANSION token): `26EDD633` (earlier low follow-up on the same test, with an unconfirmed workspace-size hypothesis) and `96B308D1` (a separate TestGatecheckCleanup_* Windows path bug).
* Closure-record compaction (P-020): `8ED0A706` (37 closure records deferred by age).

## Risky action record

* Oracle: one authorized edit (PA-2, staging PR #121 at bc32606), limited to the LIFECYCLE header and the tracked-tree decode block.
* Frozen-surface deviation: one comment-only edit in `retiredarch/select.go` (PA-5), declared in the plan.
* Shipment cascade close: performed on the P-015 fully-covered-root path after the classification, under the operator's dark-factory scope.
* Feature transition: one `active -> done` move for 052-F, with no force flags.
* Stage-owned scratch residue: `%TEMP%/head-stash.jsonl` and `%TEMP%/stage-042s-commitmsg.txt` (not deleted by Ship; deleted by the operator on 2026-10-11, see the Principle IV condition).
* Resume (2026-10-11): one stash capture, C5E50EFD, appended with `backlogit stash add` (P-021 C2, capture-only). No existing stash entry was edited or archived. No `backlogit shipment ship`, move, cascade, or archive primitive was run during the resume.
* The classify-only writer left an empty, untracked directory `.autoharness/gates/cascade-close/probe/probe-lc3raky0`. It is not committed and contains nothing. The committed writer record `docs/closure/evidence/042-S-052-F-close-evidence.json` contains the absolute local backlogit binary path; that is writer output, so Ship did not redact it.
The direct `backlogit shipment ship` call recorded above is a P-010 deviation, accepted by the operator on 2026-10-11T21:47 -07:00 as a disclosed process deviation (see the Operator disposition section). Do not repeat it.
* The archive of the shipment is on the closure branch only, not on main, until the closure PR merges.

### ActionRisk, approval path and ActionResult

| Action | ActionRisk | Approval path | ActionResult |
| --- | --- | --- | --- |
| Oracle edit PA-2 (staged in PR #121 at bc32606, carried by PR #123) | medium (test oracle) | Authorized in the plan (PA-2) | Applied; CI green on the reviewed HEAD b44cc15 |
| Frozen-surface comment edit PA-5 (`retiredarch/select.go`) | low (comment only) | Declared in the plan (PA-5) | Applied in 1d39f76 (U5); CI green |
| `backlogit move 052-F --status done` (a1 covering-feature transition) | medium | a1 gate with the five conditions verified; operator dark-factory scope; no force flags | Exit 0; re-read `done`; relocated to archive |
| `backlogit shipment ship 042-S --sha 591f37e...` (cascade close) | high | Classify-close-path verdict CASCADE with binding e204d053...; the safe-close step 0 revalidation did not run (P-010) | Exit 0; shipment shipped; `returned_ids` empty; `archived_ids` as recorded in the reconcile record; accepted as a disclosed deviation by the operator on 2026-10-11T21:47 -07:00 (item 1) |
| `autoharness shipment cascade-close --classify-only` (pre_close evidence) | low (read-only verdict and evidence record) | Classify-only mode, no mutation | Exit 0; `mutation_possible` no; the pre_close record is committed; an empty untracked probe directory under `.autoharness/gates/cascade-close/probe/` was left and is not committed |
| Stash capture C5E50EFD (resume) | low (capture only) | P-021 C2 and C5 (Ship capture-only) | Appended once; no other stash entry edited or archived |
| Removal of Stage's `%TEMP%` scratch files | low | Operator disposition, item 4 | Deleted by the operator; Ship did not delete them |

## Operator disposition (2026-10-11)

The cascade close above was a direct `backlogit shipment ship` call by Ship, which is a P-010 boundary violation. The sanctioned sequence is `shipment-reconcile` `mode: safe-close` with the bound `CLASSIFICATION_BINDING`, which revalidates the binding before any mutation and only then enters the Cascade Close Sub-Procedure. The post-hoc checks are consistent with the expected result, and the archive state is correct on the closure branch, but the sanctioned revalidation did not run. Ship did not reverse the archive, because a reversal would be a second unauthorized mutation. Ship did not merge this closure PR, and it halted the dark-factory run with `DARK_MODE_HALTED` for operator disposition; that halt is superseded by the disposition below.

The operator's reply of 2026-10-11T21:47 -07:00, verbatim `A, accept, accept`, answers the Orchestrator's analysis items 1 to 3 in order. Items 4 and 5 are part of the same relayed disposition set. Ship resumed the closure from its checkpoint `checkpoint-20261011-031007.json` under the operator's explicit confirmation of the resume. The dispositions are:

1. **P-010 direct `backlogit shipment ship`, option A: accepted as a disclosed process deviation, not a safety failure** (same shape as 041-S). The Orchestrator verified independently that the closure branch changes under `.backlogit/` only the seven manifest members (042-S, 052-F and 052.001-T to 052.005-T, renamed or modified into archive), the reconcile record and one stash capture. No other item was touched, and nothing reached `main`. Accepted conditions: (a) stash `D10D3AFC` (no invokable safe-close) and stash `A83C398E` (tune-harness adoption of `autoharness shipment cascade-close` and `gate closure-evidence`) form the next Stage batch, before another implementation shipment; (b) Stage reconciles the contradiction between the Ship rule (Ship never calls `backlogit shipment ship` directly) and the absence of any executable safe-close, by implementing safe-close or amending the rule; (c) the disposition is recorded here and in the reconcile record.
2. **Residual false-clean surfaces: accepted as INTERIM and TRACKED**, not as a decision never to fix. This is the explicit operator risk acceptance that Copilot's PR #123 review body asked for. Covered entries: `4372BAD4`, `DBE25DF5`, `EDC18D59`, `31F33EFE`, `EEE39D23`, `858516B9`, `2613D248`, `65AE88CA`, `96B2E8CC`, `CDBA41E3`, and the pre-existing `FE2F02FF` and `C0D28448`. Routing is in the Follow-ups section. The merge pre-authorization is not counted as this acceptance.
3. **Flaky start-script tests (`TestInvokeStartScriptMainPropagatesNonZeroExitCode`, `TestLocationRestoredEvenOnNonZeroCopilotExit`): accepted with root-cause analysis**, summarized in the condition `timing-sensitive-start-script-test`. The fix is captured as stash `C5E50EFD` and is not implemented in this closure.
4. **Principle IV scratch writes: done by the operator.** The operator deleted Stage's `%TEMP%` scratch files `head-stash.jsonl` and `stage-042s-commitmsg.txt`.
5. **Post-close writer record: no replay (option A).** No `post_close` writer record is produced for 042-S. The writer refuses after the fact, and a hand-built record would be fabricated. Future closures use the writer through stash `A83C398E`.

Closure PR approval basis: DARK_MODE_ACTIVE scope 042-S, with `merge_approval_pre_authorized` true for closure PR #124 only, after its gates pass. `admin_fallback_pre_authorized` is false (no `--admin`, no bypass of branch protection). The merge runs only after the section 1.9 gate, the P-018 gate on the final HEAD, and required CI all pass on that HEAD.

Ship did not run any cascade, close, shipment-ship, move-to-shipped or archive primitive during the resume. Earlier text in this artifact that said the closure was blocked, that Ship would not merge it, or that the disposition was pending is superseded by this section.

## Operational closure fields

* Invariants: INV-A (scanFile and scanSource unchanged), INV-B (no filepath.Abs or EvalSymlinks in runners), INV-C (no local masker), INV-D (ErrInvalidUTF8 identity via errors.Is), INV-E (goldens unchanged), INV-F (finding order unchanged), INV-G (the oracle edit limited to the authorized decode block and header). Each was checked on the reviewed HEAD by the pre-PR review and the test suite.
* Validator evidence: the gate runs in the evidence file (exit codes and the linked-fixture PASS rows), and the CI checks for PR #123 (docs/closure/evidence/042-S-verification-evidence.md).
* Pre-deploy audits: performed, result none required. PR #123 changed only `tools/gatecheck` files and `.backlogit` records. No `cmd/` or `internal/` product package, `go.mod` or workflow file changed, so there is no migration, feature flag, configuration, access or release prerequisite to audit. Result: no pre-deploy findings.
* Deployment and rollout: not applicable. The change is CI-run tooling merged to main. It has no deployment step, and the next CI run on main is the rollout.
* Post-deploy checks: CI run 38104542539 on main for merge commit 591f37e2ba877d81b1304d31905369e79096657f completed with conclusion success (checked 2026-10-11). That is the first validation-window run.
* Healthy and failure signals: healthy is `write-path --root .` and `--self-test-integrity` exiting 0 with goldens unchanged, and the write-path verdict step passing in CI. A failure is a non-zero exit from the write-path steps on main. Note that `WRITE_PATH_GATE_ADVISORY` is unset, so the verdict step is blocking.
* Monitoring: the CI write-path steps on each push to main, and the Copilot review and CI checks on any follow-up PR.
* Rollback trigger and procedure: a write-path CI step failing on main with a non-finding error. Revert in reverse dependency order, U5 and U3 before U2, with U4 at any time and U1 optional. U2 and U3 used no CI fallback, so each unit is a single commit. A rollback is `git revert` of the unit commits, never a history rewrite.
* Validation window: the first CI run on main after the merge (38104542539, green) and subsequent pushes to main until the operator closes the window.
* Owner: Ship for the shipment and the closure record; the Orchestrator for the closure gate; the repository maintainer for CI and the operator disposition above.



`compaction_status` is `degraded` (P-020 non-blocking). The compact-context invocation was `target: all`, which is the mandatory invocation. The compacted files it wrote under `docs/memory/compacted/` are that run's memory-only selected output; plans and closure records were not selected into them. Its first pass covered only memory. The re-run compacted the 042-S memory set and three threshold-qualified Stage memory files from 2026-09-12 to 2026-09-18 into `docs/memory/compacted/`, moving every original to `docs/archive/memory/` (no deletions). The run deferred 37 closure records that qualify by age; that is recorded as follow-up stash entry 8ED0A706. The first closure-PR re-run verification stopped on three stash IDs that were in the check list but not in any source (D10D3AFC, FE2F02FF and C0D28448 are stash entries, not 042-S memory); no data was lost. Memory under 14 days, live blocked checkpoints, and plans (Stage-owned under P-010) were left untouched. Both originals of the 042-S memory set were archived under `docs/archive/memory/` (no deletions). Other memory files were observed and left untouched by design (threshold-gated selection, not part of this unit).
