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
closure_pr: recorded-in-pr-body
compaction_status: done
close_path: cascade
classification_binding: e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95
closure_status: READY_WITH_CONDITIONS
releasability: READY_WITH_CONDITIONS
conditions:
  - id: cascade-post-close-writer-evidence-missing
    summary: "The harness writer produced the pre_close record (captured before the close), but no post_close writer phase exists. The mutating form of autoharness shipment cascade-close was deliberately not used, because it invokes backlogit shipment ship itself and the Ship role boundary reserves that invocation to the classified close path. The post-close state is evidenced by the shipment-reconcile post-mode check and the archive records, not by the writer. Unsatisfied and disclosed; operator decision requested on whether a post_close writer record should be produced by a separate verified step."
    satisfied: false
    evidence: "docs/closure/evidence/042-S-052-F-close-evidence.json (phase pre_close, classifier CASCADE, engine VERIFIED, mutation_possible no, failures []); .backlogit/reconcile/042-S-cascade-close-2026-10-11T02-19-08Z.md ('Writer-form evidence' and 'Close' sections)"
  - id: safe-close-agent-executed-binding
    summary: "Process residual, disclosed and matching 041-S: there is no invokable shipment-reconcile safe-close CLI or gate (stash D10D3AFC, Stage-owned). The classify-close-path verdict and CLASSIFICATION_BINDING were computed by the agent from the canonical v1 lines, and the cascade primitive was invoked by the agent on the P-015 fully-covered-root path. The engine's own revalidation does not recompute this binding."
    satisfied: true
    evidence: ".backlogit/reconcile/042-S-cascade-close-2026-10-11T02-19-08Z.md (binding e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95 with its canonical lines); stash entry D10D3AFC (active, Stage-owned)"
  - id: principle-iv-scratch-writes
    summary: "Principle IV / AGENTS.md CLI workspace containment, disclosed and not remediated. Ship wrote scratch files under logs/ (git-ignored) by design. Outside the workspace, Ship created and removed a short-lived symlink probe directory under the OS temp directory during intake. Repository Go tests write temporary fixtures under t.TempDir (the existing repository convention, recorded in the plan). The tool runtime spilled large command output to the OS temp directory. Operator acknowledgement or cleanup is requested."
    satisfied: false
    evidence: "Ship session log (symlink probe at intake, removed); plan Constitution Check (Principle IV interpretation); PR #123 body 'Local Review Readiness' and 'Notes'"
  - id: timing-sensitive-start-script-test
    summary: "tests/integration TestInvokeStartScriptMainPropagatesNonZeroExitCode (a pwsh-driven test with a 30 s budget for start.ps1) hit start.ps1 exceeded 30s timeout in three full local runs and passed in isolation (39.8 s, 19.6 s) and in the final full local run on the reviewed HEAD. start.ps1 runs the backlogit and Engram sidecar syncs, whose cost grows with workspace size, and this branch added about 370 archived artifacts. The test does not exercise write-path code. The Linux CI test job passed on the reviewed HEAD (run 38104296737). Follow-up stash entry recorded."
    satisfied: true
    evidence: "logs/gate-052.004-T-fulltest.txt, logs/final-e202fe3-fulltest.txt (timeouts); logs/052.004-T-integration-rerun.txt and logs/final-startscript-isolated.txt (pass); logs/final-b44cc15-fulltest.txt (pass); PR #123 checks (test pass)"
  - id: copilot-review-residual-risk-acceptance
    summary: "Copilot's review of the final HEAD was COMMENTED with zero review threads and 0 open findings, but its body asked for human risk acceptance of the documented false-clean residual surfaces. Those surfaces are the accepted residuals captured in stash entries 4372BAD4, DBE25DF5, EDC18D59 and 31F33EFE and the new captures EEE39D23, 858516B9, 2613D248, 65AE88CA, 96B2E8CC and CDBA41E3. The operator's explicit merge pre-authorization for PR #123 is recorded as the acceptance."
    satisfied: true
    evidence: "PR #123 review by copilot-pull-request-reviewer on b44cc15 (COMMENTED, zero threads); autoharness gate copilot-review 123 verdict SATISFIED for b44cc15; DARK_MODE_ACTIVE merge pre-authorization"
  - id: copilot-request-via-rest
    summary: "The Copilot request was made through the REST requested_reviewers endpoint with copilot-pull-request-reviewer[bot]. GraphQL reviewRequests did not show the request afterward, but Copilot reviewed the final HEAD (review event at 2026-10-11T02:14:20Z). Disclosed; the P-018 gate verdict SATISFIED is the controlling signal."
    satisfied: true
    evidence: "gh api --method POST repos/softwaresalt/intercom/pulls/123/requested_reviewers; GraphQL reviews for PR 123; autoharness gate copilot-review verdict SATISFIED"
  - id: covering-feature-archived-on-done-move
    summary: "Disclosure, matching the 041-S precedent: backlogit move 052-F --status done relocated 052-F from queue to archive in backlogit 1.11.0. The pre-close snapshot recorded the true location. Location is not a coverage input and the descendant topology was unchanged, so the CASCADE verdict stands."
    satisfied: true
    evidence: "backlogit move 052-F --status done (exit 0); .backlogit/archive/052-F.md (archived_status done); reconcile record a1 section"
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

Releasability: `READY_WITH_CONDITIONS`. The conditions are listed in the frontmatter, and each one carries its evidence.

## Merge and gate record

* Merge commit `591f37e2ba877d81b1304d31905369e79096657f`, confirmed by `gh pr view 123` (MERGED) and by `git merge-base --is-ancestor` against `origin/main`.
* Reviewed HEAD `b44cc1539c5d734caef49726306c575f3e6b2700`. `headRefOid` matched at merge.
* Required CI on the reviewed HEAD: all checks passed, including `test` (Linux, required) and `test (windows, advisory)`, `lint`, `security`, `cross-compile` for four targets, and `merge-strategy structural verification`.
* Section 1.9 readiness: the PR body's Local Review Readiness block cites the reviewed HEAD, outcome `READY_WITH_FOLLOWUPS`, `P0=0, P1=0` (one consensus-rated P1 dispositioned as a P-021 C1 out-of-scope capture), full local build evidence, and follow-up stash IDs.
* P-018 copilot-review gate: `SATISFIED` for the reviewed HEAD on the final run before merge.
* P-009: the repository accepts merge commits only (`allow_merge_commit: true`, `allow_squash_merge: false`, `allow_rebase_merge: false`). Merge used `--merge --match-head-commit`.
* P-016: `pipeline-topology` lifecycle gate exit 0 before the merge and again before the close.
* Operator approval source: the DARK_MODE_ACTIVE activation record (`merge_approval_pre_authorized: true`), scope 042-S only. No admin fallback was used.

## Review and remediation record

* Pre-PR review: the persona pool (Correctness, Go, Security, Constitution, Maintainability, Scope Boundary) and the multi-model Adversarial Review, run against 1d39f76. Full dispositions are in the PR #123 body.
* Review-fix cycles: three (1a704c5, 5eb2e78, e202fe3), each followed by a targeted re-review or a self-audit of the diff, within the per-task cap of three. Accepted follow-ups at the cap are stash entries.
* Adversarial P1 candidates: V-1 (oracle and select.go verification) closed by Ship; C-5 (Windows non-ASCII env alias) not honoured by git 2.55 on this host and not reachable on the required Linux gate; C-1 (fixture-name echo) captured under EDC18D59 under P-021 C1, because the PASS and FAIL lines already echo the same names for regular fixtures.

## Verification on the reviewed HEAD

* `go build ./...` exit 0; `go vet ./...` exit 0; `gofmt -l .` empty; `goimports -l .` empty; `golangci-lint run ./...` 0 issues; `GOOS=linux go vet ./tools/gatecheck/...` exit 0.
* `go test -count=1 ./...` on b44cc15: exit 0, 14 packages ok.
* Live tree: `write-path --root .`, `write-path --root . --self-test`, `write-path --root . --self-test-integrity`, `retired-arch --root .` and `scripts/check-write-path-precondition.sh`, all exit 0.
* `TestRunFixtureSelfTest_LinkedFixture_FailsClosed`: 5 of 5 PASS, 0 SKIP on this host.
* `WRITE_PATH_GATE_ADVISORY` repository variable: unset, so the verdict step is blocking.

## Close path

* Covering-feature completion gate (a1): satisfied (manifest tasks declared `done` and archived; no live descendants; descendant set equal to the manifest minus the feature; containment holds; 052-F was `active` before the single transition).
* Classify-close-path: `CASCADE` / `FULLY_COVERED_ROOT`, binding `e204d053e4ad1925f723d9eddcb1462df16f982ec175aab1a7bd060bc672cd95`.
* Cascade close: `backlogit shipment ship 042-S --sha 591f37e2ba877d81b1304d31905369e79096657f`. Result: `shipment_status: shipped`, `returned_ids: []`, `archived_ids: [052.001-T, 052.002-T, 052.003-T, 052.004-T, 052.005-T, 042-S, 052-F]`. Two-set gate: pass both checks. Post-mode: pass; no archive deletions.
* Closure commit: `672eddf` on `post-merge/042-s-gatecheck-batch-b1-write-path-input-hardening` (`chore: archive 042-S backlog artifacts`). Never committed to main.

## Source artifact cleanup

The covering feature 052-F carries no `source_stash_id` or `source_deliberation_id` custom field, so no source stash entry is archived by Ship in this closure. The consumed stash entries B83F53BB and 05E12A6F are archived by Stage per the plan and are not touched here.

## Stash follow-ups (visible to Stage)

* Captured as DEFERRED SCOPE EXPANSION (P-021 C2, capture-only): `EEE39D23` (scan-set selection of tracked links, gitlinks and case variants; medium), `858516B9` (fixture self-test directory and hard-link gaps, FIFO row), `2613D248` (env residuals MSYS, CYGWIN, MSYS2_*, boundary rows), `65AE88CA` (test-helper consolidation), `96B2E8CC` (GitRunner context parameter), `CDBA41E3` (cgo, linkname and os/exec write detection).
* Existing residual entries cited by the PR: `4372BAD4` (relative git refusal), `DBE25DF5` (wrapper root discovery), `EDC18D59` (echo sinks, including the U4 refusal site), `31F33EFE` (rune extensions and the retiredarch asymmetry).
* Accepted follow-ups at the review-fix cycle cap: the non-repository test's no-ceiling control, exact context for the non-control rows of the bad-listing test, the Windows 8.3-name check, and the start-script timing item. Recorded in `.backlogit/stash.jsonl`.

## Risky action record

* Oracle: one authorized edit (PA-2, staging PR #121 at bc32606), limited to the LIFECYCLE header and the tracked-tree decode block.
* Frozen-surface deviation: one comment-only edit in `retiredarch/select.go` (PA-5), declared in the plan.
* Shipment cascade close: performed on the P-015 fully-covered-root path after the classification, under the operator's dark-factory scope.
* Feature transition: one `active -> done` move for 052-F, with no force flags.
* No destructive git operation. No history rewrite. No merge through admin fallback.

## Compaction (P-020)

`compaction_status` is `done`. The compact-context run (P-020, bounded to the mandatory just-closed memory set) consolidated `docs/memory/2026-10-10/stage-gatecheck-batch-b1-write-path-input-hardening-memory.md` and `docs/memory/2026-10-11/ship-042-s-session-memory.md` into `docs/memory/compacted/2026-10-11-042-s-052-f-compacted.md`, and archived both originals under `docs/archive/memory/` (no deletions). Other memory files were observed and left untouched by design (threshold-gated selection, not part of this unit).
