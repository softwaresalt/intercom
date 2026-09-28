# Ship session checkpoint — shipment 028-S PR #75 open, awaiting operator merge approval

**Date**: 2026-09-27
**Branch**: `feat/028-s-merge-strategy-structural-verification-gate` (pushed)
**Shipment**: 028-S (`active`), covering feature 031-F, "Merge-strategy structural verification gate"
**PR**: https://github.com/softwaresalt/intercom/pull/75 (base `main`)
**Resumed from**: `docs/memory/2026-09-21-ship-028-s-halted-pre-existing-test-failure.md`

## Status: PR open, CI green, WAITING FOR OPERATOR MERGE APPROVAL (P-014)

The session did not merge. Merge needs an explicit operator approval signal.

## What this session did

1. **DCD67C30 operator-authorized one-off exception**: commit `fdf075f`. This was a separate authorization under P-021 C4 and did not expand 028-S scope.
   * Renamed `docs/closure/027-S-030-F-closure-evidence-repair-post-merge-closure.md` to `docs/closure/2026-09-21-027-s-030-f-closure-evidence-repair-closure.md`. The date comes from the artifact's frontmatter.
   * The target is the dated closure form in operational-closure SKILL.md `## Output`. That form is documented for pre-merge/post-deploy closures; it is used here because the artifact belongs to a chore with no shipment. The strict `{shipment_id}-{feature_id}-post-merge-closure.md` name would collide with the canonical `027-S-030-F-post-merge-closure.md`.
   * Updated the live references in 3 `docs/memory` files. The test, its regex, and SKILL.md are unchanged.
   * `go test ./...` is now green.
   * DCD67C30 is resolved. Stage can archive it; Ship did not change the entry.
2. **Local adversarial review**: 3 reviewers per cycle, 3 review-fix cycles in total. The circuit breaker is reached.
   * Cycle 1, at `fdf075f`: BLOCKED.
     * P0: `administration: read` is not a valid GITHUB_TOKEN scope. actionlint confirmed it, and GitHub would have rejected the whole workflow.
     * P1: the credential doc was wrong.
     * P2 ×2: non-boolean values were treated as PASS, and an empty evaluator verdict exited 0.
     * The fixes landed in `baf6f72`.
   * Cycle 2, at `36881a9`: BLOCKED on a P1 that cycle 1 missed. The live repo scan never delivered the API response to python, because the python heredoc overrode the here-string on stdin, so every live scan reported SKIP. `32519f5` fixes it with a temp-file transport, self-test coverage of that transport and of exit 2, and doc accuracy fixes.
   * Cycle 3, at `32519f5`: READY_WITH_FOLLOWUPS with P0=0 and P1=0. The three P3 items went to stash follow-up `150364D2`.
3. **P-021 deferred captures** (threadless, pre-PR): `124AE9DE` (wire a PAT or GitHub App secret), `C8914513` (required mode should fail closed on SKIP), `C98B92F0` (closure `mode` vs. name; `DISCOVERY-STATUS: AMBIGUOUS DCD67C30`). They are committed in `36881a9`.
4. **Gates**:
   * gofmt clean.
   * vet, build, and `go test -count=1 ./...` all green.
   * Self-test: 16 fixture checks and 2 exit-code checks pass.
   * actionlint exits 0.
   * The live end-to-end scan with an admin user token returned `PASS` (squash and rebase are both false).
   * pipeline-topology lifecycle gate: pass.
   * P-018 copilot-review: `NOT_APPLICABLE` (Copilot is not engaged).
5. **CI on PR #75**: every check passed, including `ci gate`. The `merge-strategy` job prints SKIP with GITHUB_TOKEN, which is the documented, expected behavior until `124AE9DE` lands.

## Branch commits (origin/main..HEAD)

The 9 prior shipment and housekeeping commits, then `fdf075f` (DCD67C30), `baf6f72`, `36881a9`, `32519f5`, `aa25250` (stash follow-up), and this checkpoint commit.

## Next steps

1. **Operator**: give explicit merge approval for PR #75. The merge must use a merge commit (P-009); the repo already has squash and rebase disabled.
2. **Ship, after approval**:
   * Last-mile P-018 re-run.
   * Run §1.9 again if HEAD advanced.
   * `gh pr merge 75 --merge`.
   * Merge confirmation gate.
   * Post-merge closure on `post-merge/028-s-031-f-merge-strategy-gate`: a1 covering-feature completion, reconcile pre → classify-close-path → safe-close → post, closure artifact `docs/closure/028-S-031-F-post-merge-closure.md`, compact-context (P-020), and source stash cleanup.
3. **Stage**: triage `124AE9DE`, `C8914513`, `C98B92F0`, and `150364D2`. Archive `DCD67C30`, which is now resolved.
