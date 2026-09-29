# Ship session checkpoint — 029-S feature PR open, awaiting operator merge approval

**Date**: 2026-09-28
**Branch**: `feat/029-s-extract-gate-python-engines` (base `beeb84f`)
**Shipment**: 029-S is **active**. Covering feature 032-F is active (it moves to done only at post-merge closure). Tasks 032.001-T through 032.007-T are done and archived, with commits tracked.
**Plan**: `docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md` Unit 2 (AC-2.1..AC-2.8)

## Status: feature PR open, WAITING FOR OPERATOR MERGE APPROVAL (P-014)

Not dark mode, and admin fallback is not authorized. The merge must use a merge commit (P-009).

## Commits per task

| Task | Commit | Summary |
|---|---|---|
| claim | `f05d0ca` | claim 029-S (032-F active) |
| harness | `3173e97` | red-phase stdlib unittest harness; `.gitignore` gains `__pycache__/`, `*.pyc` |
| 032.001-T | `8b44682` | canonical masker decision (retired-arch superset); expected write-path delta set = EMPTY |
| 032.002-T | `d6c2c88` | `scripts/lib/gomask.py` verbatim move; captured deferred stash C312BD4C |
| 032.003-T | `ad9b3a8` | `scripts/lib/retired_arch.py` engine (no import-time argv/cwd reads) |
| 032.004-T | `2247b3c` | retired-arch script reduced to a wrapper; pathspec pin re-derived |
| 032.005-T | `aa9298c` | write-path checker imports the shared masker; local clone deleted |
| 032.006-T | `3042ae5` | ci.yml lint job: pinned setup-python 3.12 + blocking unit-test step |
| 032.007-T | `ec39da9` | `docs/closure/029-S-032-F-behaviour-preservation-evidence.md` |
| tracking | `e4b7973` | task commit SHAs recorded |
| review fix | `a09bb0c` | in-scope hardening from local review (engine path via BASH_SOURCE, PYTHONDONTWRITEBYTECODE, test tightening) |
| capture | `6a10e46` | P-021 deferred captures 40C421EF, 54EF986C, 8387758F |

## Gates (local, at the review-fix HEAD `a09bb0c` and later)

- `gofmt -l .`: empty. `go vet ./...`, `go build ./...` and `go test -count=1 ./...`: all ok.
- `python3 -m unittest discover -s scripts/lib/tests`: 37 tests pass.
- Retired-arch (repo, --self-test, --self-test-integrity) and write-path (repo, --self-test, --self-test-integrity): PASS.
- gitignore append-only + self-test, unignore regression + self-test, merge-strategy --self-test: PASS.
- Re-captured evidence at the review-fix HEAD is byte-identical to the recorded post-extraction evidence (3042ae5).
- pipeline-topology lifecycle gate: exit 0.

## Local review

- The Adversarial Review has 3 reviewers (Anchor gpt-5.6-sol, Tier 1 gpt-5.4-mini, Tier 3 claude-opus-5), with outcome READY_WITH_FOLLOWUPS and P0=0, P1=0.
  - Review range: `beeb84f..e4b7973`.
  - Findings 2, 3, 4 and 6 were in scope (P-021 C1/C3) and were fixed in `a09bb0c`. Ship reviewed the fix cycle directly (1 of 3 cycles).
  - Findings 1, 5 and 7 are out of scope and were captured on the threadless pre-PR path:
    - 40C421EF: pin Python for the gate steps
    - 54EF986C: use a real pinned linter
    - 8387758F: add write-path struct-tag fixtures
- Deferred entries this shipment: C312BD4C, 40C421EF, 54EF986C, 8387758F (all capture-only; Stage triages).

## Next steps

1. Check that the PR CI is green, and confirm the new "Python gate-engine lint and unit tests" step ran and passed. That completes the real-CI-run check for 032.006-T.
2. Handle any review threads: fix, push, reply, then resolve. Out-of-scope threads follow capture, reply citing the ID, resolve, residual-risk.
3. Run the P-018 `autoharness gate copilot-review <pr> --repo softwaresalt/intercom --enforcement auto`.
4. Wait for **explicit operator merge approval**. Then:
   - re-run the §1.9 and P-018 last-mile gates;
   - `gh pr merge <pr> --merge --match-head-commit <head>`;
   - run the Step 6 post-merge closure on the `post-merge/029-s-032-f-extract-gate-python-engines` branch.