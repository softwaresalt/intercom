---
title: "Dark staging publication corrective maintenance"
description: "Bounded corrective handoff for the parent resuming shipment 073-S"
date: 2026-10-30
status: blocked
tags: [memory, policy, corrective-maintenance]
---

## Scope and authority

* Operator authorized this separate correction and normal scoped PR publication
  and merge after required gates: "Let's apply that policy correction now and
  then resume the original operation."
* Corrective branch: `fix/dark-staging-publication`, based on refreshed
  `origin/main` at `19753551a37314718292d922a57d12fed440efad`
* Original Stage branch `chore/stage-gatecheck-wrapper-test-ci-reliability`
  remains preserved at `3a6f5e8ae5707436645e610d7f7d4f25d90093ac`.
  No original 073-S/080-F implementation or staging publication is authorized
  in this executor session.
* No shipment was claimed. No new backlog IDs were created.
* Backlog continuity key: `corrective-dark-staging-2026-10-30`, written through
  registered `backlogit memory save`
* ENGRAM_DEGRADED: two prior daemon startup failures retained without replay.
  Known-path and bounded local fallback only; no external retrieval.

## Decisions and modified surfaces

* P-010/P-017 and both agent handoffs now distinguish normal operator-only
  publication from recorded-dark Orchestrator publication of already committed,
  owned, allowed, scope-bound Stage artifacts.
* Stage remains forbidden from PR/build/implementation work. No Orchestrator
  staging commits, direct default push, shipment claims or source/config edits
  are granted.
* Current-HEAD independent readiness, required CI, P-009/P-014/P-018 remain
  gates. Explicit dark activation/clarification supplies scoped normal merge
  approval, never admin or destructive approval.
* Clean unpublished Stage branches explicitly enter the publication arm.
* Modified tests: `staging_gate_actor_contract_test.go` adds three focused
  boundary cases; `stage_branch_gate_test.go` replaces only the incompatible
  operator-only assertion while retaining Stage's prohibition.
* Policy registry advances from 1.26.0 to 1.27.0; constitution remains unchanged.

## Evidence at 2026-10-30T03:25:43Z

* RED: focused tests failed before policy edits: missing Orchestrator grant,
  normal-mode arm, activation source and updated Stage handoff.
* GREEN: focused staging tests plus
  `TestShipFeatureCompletionContract33Rows` passed.
* An initial guard run rejected sentence-ending `P-018.` as retired task
  prefix `018.`; reworded the new changelog sentence without weakening the guard.
* `gofmt -l .`: no output
* `go vet ./...`: passed
* `go test ./...`: first run failed in the pre-existing
  `TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst` timeout
  (original 073-S scope). No wrapper code changed. One bounded full retry with
  warm caches and workspace-local scratch passed.
* `go build ./...`: passed
* Independent review of `0f74407` found an in-scope P2: Step 4's manifest-only
  success arm did not enforce reviewed versions of all scoped artifacts. Added
  a failing regression before correcting it to blob-ID/absence comparisons.
  Strengthened scoped merge assertions for independent review, current-HEAD
  PASS and readiness invalidation after HEAD changes (P3 test-quality finding).
* Reviewers had no Git execution tool; final review must use a coordinator-captured
  committed diff, not treat incomplete initial reads as readiness.
* Current-HEAD final review and PR/CI/merge: blocked, not asserted complete
* Merge settings verified: merge commits enabled, squash/rebase disabled
* One worktree only; no destructive command, force or admin fallback performed

## Safety and compaction

* Freeze-scope ProposedAction: non-destructive correction of the three named
  policy/agent surfaces and two existing tests, plus required continuity/evidence
* ActionRisk: high (authorization wording); ActionResult: applied under explicit
  operator approval, publication still gated
* Rollback: a separate reviewed revert PR, never reset or force
* Destructive/admin actions remain blocked without separate authority.
* Pre-review compact-context assessment: unrelated memories/plans/closure
  artifacts excluded from this corrective scope; active original shipment and
  this maintenance checkpoint preserved. No unrelated archive/move performed.

## Final blocker at 2026-10-30T03:41:28Z

The final full suite failed, so this correction is NOT merge-ready. No PR was
opened and no merge, admin bypass, force operation, deletion or shipment closure
was attempted.

The final run confined `TEMP`, `TMP` and `TMPDIR` to the ignored workspace path
`logs/test-tmp`. Existing integration fixtures use the process temporary root
as an intentionally outside-repository destination. Relocating it inside this
repository changes that premise, producing these failures:

* `TestBuildScript_RefusesOutputPathOutsideRepoRoot`: expected refusal, got success
* `TestResolveOutputPathWithinRoot/rejects_directory_symlink_escape`
* `TestResolveOutputPathWithinRoot/rejects_dangling_symlink_escape`
* `TestResolveOutputPathWithinRoot/rejects_junction_escape`

The three link cases resolved to paths below `dist/output-path-guard-*`, with
targets now inside the repository rather than outside it. Evidence is the
bounded `go test ./...` output from command session 164 (native exit 1). The
gatecheck packages passed in that run. No containment implementation or fixture
was changed, and no selector/skip was used to claim a full-suite pass. Moving
scratch outside the authorized tree or changing these unrelated tests is not
part of this correction.

Final narrow evidence:

* Second RED observed before the remote-version correction: missing
  `compare Git blob IDs`
* Final focused staging tests and ship/destructive guards: GREEN, uncached
  `go test ./tests/integration -run
  'Test(DarkStagingPublicationContract|StageBranchGate_ContractCorrected|StagingGateActorContract|ShipFeatureCompletionContract33Rows)$'
  -count=1`
* Final format and vet: passed
* Full build after the final correction: passed
* Independent initial review: P0=0, P1=0, one deduplicated P2 and one P3,
  both addressed. Review completeness was blocked because reviewer tools could
  read named files but could not execute the committed Git diff. Do not
  describe this as final current-HEAD readiness.
* Final independent review, required remote CI and P-018 have not run.
* Closure status: BLOCKED before PR; no post-merge closure is applicable.
* Compaction: scoped assessment only, active blocked checkpoint preserved;
  no unrelated memories/plans/closure files moved or deleted.

## Next steps

The parent must resolve the full-suite fixture/environment constraint without
silently weakening containment or required checks, then obtain independent
review of a captured exact committed diff. Only after all gates pass may the
corrective PR be opened and normally merged with merge-commit strategy. Admin
and destructive actions still require separate explicit authority.

Original 073-S cannot resume under the proposed corrected rules until this
correction is reviewed and merged. Do not publish the original Stage branch from
this executor session. Preserve its commit `3a6f5e8` unchanged.

## Validation method repair at 2026-10-30T04:53Z

The full-suite blocker remains open (breaker at 3 attempts, not reset). A
separately approved standalone diagnostic found that the gatecheck helper
confines temp to its own package directory, and that the integration tests need
a temp root outside the repository, so one global TEMP value cannot serve both.
The temp contract, same-process recipe, gate provenance and evidence links are
in `validation-method-repair-memory.md`. The full suite was not rerun, and the
merge-readiness gate is still unmet.
