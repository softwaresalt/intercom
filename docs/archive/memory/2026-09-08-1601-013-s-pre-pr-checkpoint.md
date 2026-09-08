---
title: "013-S pathsafe reparse-point containment hardening — pre-PR checkpoint"
date: 2026-09-08
shipment: 013-S
feature: 014-F
branch: feat/013-s-pathsafe-reparse-point-containment-hardening
status: in-progress
---

# Session checkpoint: 013-S, ready for PR creation

## Items completed

All 8 manifest tasks (014.001-T through 014.008-T) are `done` in backlogit.
Feature `014-F` and shipment `013-S` remain `active` (will close at post-merge
shipment safe-close).

| Task | Summary | Key commit(s) |
|---|---|---|
| 014.001-T | RED: reparse-point containment matrix (C1/C2/C3) | `bf9cd55` |
| 014.002-T | RED: invert GO-14 dangling-final-symlink regression lock | `9308460` |
| 014.003-T | Reparse-aware canonicalization helper (platform-split) | `ef56fe0` |
| 014.004-T | GREEN: wire canonicalization into containment check | `0c51669` |
| 014.005-T | Flip fail-open terminal branch to fail-closed | `3e1cb58` |
| 014.006-T | Distinct unverifiable-target error message | `541073c` |
| 014.007-T | Close branch-coverage gaps F4/F11 (POSIX) | `4c9f8df` |
| 014.008-T | Reconcile risk register and doc comments | `c4e5ea2`, `d4ffeab` (review fix) |
| — | Adversarial-review remediation (post-014.008-T) | `6bab167` |

Full commit range: `231862a..6bab167` (15 commits) on top of `main@e48cc55`.

## Quality gates — all green

- `go build ./...` (Windows, and cross-compiled linux/amd64, darwin/arm64)
- `go vet ./...`
- `gofmt -l .` (clean; two real formatting fixes applied and re-verified
  along the way — numbered-list indentation and a stray blank line from
  edit-tool merges)
- `golangci-lint run ./...` — 0 issues (including the errcheck fix for
  `reparse_windows.go`'s `CloseHandle`)
- `go test ./...` — all packages green, including `tests/integration`
- `go test -v ./internal/pathsafe/...` — 45 tests, 0 failures (only one
  pre-existing non-Windows-only test correctly skipped on this Windows host)
- `scripts/check-write-path-precondition.sh --self-test` — green (plan AC7)

## Review gates

1. **Standard report-only code review** (code-review agent): found and this
   session fixed 3 issues (700B41CE causal misattribution, uncPrefix
   misattribution, errcheck violation). Verdict: READY_WITH_FOLLOWUPS.
2. **Multi-persona adversarial review** (3 independent reviewers,
   `docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md`):
   consensus that the core fix is sound, no newly-introduced escape-direction
   bypass. Fixed 5 findings directly (pwsh silent-skip → fail-loud, 2 new
   Resolve()-level regression-lock tests, uncPrefix comment re-corrected
   after a second-draft inaccuracy verified against actual Go stdlib
   source). Deferred 3 findings via P-021 capture (stash IDs below).
   Verdict: READY_WITH_FOLLOWUPS.

## P-021 deferred scope entries (stash)

- `4104AF54`: golangci-lint's ubuntu-only CI job structurally can't lint
  windows-tagged production files (CI workflow change, out of scope).
- `E428AB46`: NewRoot (EvalSymlinks) vs. checkSymlinkEscape
  (canonicalizeReparse) canonicalization-mechanism asymmetry if a workspace
  root is itself ever a junction/GUID volume — plausible, unconfirmed,
  needs its own design deliberation.
- `E4C5413F`: bundled U5 (long-path CreateFile prefixing asymmetry, fail-
  closed direction only), U6 (canonicalizeReparse relies on caller-side
  stripUNCPrefix rather than self-normalizing), U7 (LazyProc.Call panics
  rather than erroring on unresolvable DLL export, theoretical/unreachable
  on supported Windows versions).

None of these are P0/P1 blocking findings; all are documented residual risk
appropriate for `READY_WITH_FOLLOWUPS` follow-up handling.

## Branch state

On `feat/013-s-pathsafe-reparse-point-containment-hardening`, 15 commits
ahead of `main@e48cc55`, working tree clean. HEAD = `6bab167`.

## Next steps

1. Prepare PR body with `## Local Review Readiness` block (§1.9).
2. Invoke `pr-lifecycle` skill: push branch, create PR via `gh pr create`.
3. Request Copilot review; iterate reply/fix/resolve per comment.
4. Run §1.9 + P-018 copilot-review gate before merge.
5. Merge (merge-commit strategy) under dark-mode pre-authorization once all
   gates pass.
6. Full post-merge closure: shipment safe-close, `operational-closure`
   artifact, `compact-context` (P-020 mandatory), backlog index resync.
