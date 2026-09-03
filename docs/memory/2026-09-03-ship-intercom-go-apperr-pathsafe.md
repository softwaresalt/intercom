---
title: "Ship Session Memory — 002-S: intercom-go P1 Error Taxonomy and Workspace Path Containment"
date: 2026-09-03
shipment: 002-S
feature: 002-F
branch: feat/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment
---

# Ship Session Memory — 002-S

## Items Completed

All 3 tasks / 8 subtasks in the shipment manifest, strictly in the plan's
mandated dependency order (`A1 → A2 → A3 → B1 → B2 → B3 → B4 → C1`):

* `002.001-T` (`internal/apperr`): A1 (kinds + display), A2 (sentinels +
  `errors.Is`), A3 (constructors + `Unwrap`) — all 3 subtasks done.
* `002.002-T` (`internal/pathsafe`): B1 (`Root` canonicalization), B2
  (lexical rejection + normalize), B3 (containment + `Resolve`), B4 (symlink
  escape detection) — all 4 subtasks done.
* `002.003-T` (provenance): C1 (`docs/oracle-pin.md`) — 1 subtask done.
* Covering feature `002-F` marked done; all queue items archived to
  `.backlogit/archive/`.

## Commits (on `feat/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment`)

1. `84636de` feat(apperr): define 14 error kinds and display contract (unit A1)
2. `3659f59` feat(apperr): add kind sentinels and errors.Is support (unit A2)
3. `48666e8` feat(apperr): add constructors and cause-preserving Unwrap (unit A3)
4. `c74b1dd` feat(pathsafe): add validated Root canonicalization type (unit B1)
5. `12b150b` feat(pathsafe): implement lexical rejection and component normalization (unit B2)
6. `f1e3ee1` feat(pathsafe): enforce containment in Root.Resolve (unit B3)
7. `7f35806` feat(pathsafe): detect symlink escapes for existing paths (unit B4)
8. `6fb1fd8` docs(oracle-pin): record behavioral-oracle pin and parity trace (unit C1)
9. `ead9a47` chore(backlog): archive completed 002-F feature and its 3 tasks / 8 subtasks
10. `f3a7fda` fix(pathsafe): close symlinked-ancestor containment gap and GOOS-gate isRooted (review remediation)
11. `3a54498` chore(stash): capture 5 deferred P2/P3 review findings under P-021 C2

Reviewed HEAD: `3a54498986de99953f76520566f9617f47fe558083` (short `3a54498`).

## In-Session Findings (beyond the plan's literal mechanism)

Two real Go-specific gaps were found and closed during implementation/review,
both same-contract-surface (required to deliver the NON-NEGOTIABLE workspace
path-containment control the plan charters):

1. **Windows rooted-vs-absolute gap**: `filepath.IsAbs("/etc/passwd")` is
   `false` on Windows (no volume; "rooted" not "absolute"), so the plan's
   literal `VolumeName`/`IsAbs` mechanism under-rejects a mandatory negative
   test case (`/etc/passwd`, required by acceptance criterion B2.2 on every
   platform). Closed by adding `isRooted()`.
2. **Symlinked-intermediate-directory bypass (P1, convergent across 3
   independent reviewers)**: `checkSymlinkEscape` originally gated
   re-validation on `os.Stat(resolved)` succeeding for the *full* path. A
   symlinked intermediate directory pointing outside the workspace, combined
   with a non-existent leaf (the dominant new-file-write shape), bypassed
   detection entirely. Closed by walking up to the nearest existing
   ancestor and re-asserting containment there.
3. **P2**: `isRooted`'s backslash check was not `GOOS`-gated, risking false
   rejection of legitimate relative paths on non-Windows. Closed.

Both P1/P2 fixes are locked in by new regression tests
(`TestResolveRejectsSymlinkedIntermediateDirEscapingRoot`,
`TestNormalizeDoesNotOverRejectLiteralBackslashOnNonWindows`).

## Validation Evidence

* `gofmt -l .` — clean (empty output).
* `go vet ./...` — clean.
* `golangci-lint run ./...` — 0 issues.
* `staticcheck ./...` — clean.
* `govulncheck ./...` — 0 vulnerabilities affecting this code (2
  package-level / 6 module-level pre-existing advisories not called by our
  code, consistent with `001-S`'s CVE baseline).
* `go test -race ./...` — all packages pass, including
  `internal/apperr` and `internal/pathsafe`'s full symlink-privilege-gated
  suite (symlink privilege available in this environment; tests did not
  skip vacuously).
* `CGO_ENABLED=0 go build ./...` — succeeds (CGO-free mandate preserved).
* `go.mod`/`go.sum` unchanged — zero new dependencies (stdlib only).

## Review Gate

Five parallel structured reviews (Go Reviewer, Correctness Reviewer,
Security Reviewer, Architecture Strategist, Constitution Reviewer),
report-only mode, against the full diff and the reviewed plan:

* All five returned **READY_WITH_FOLLOWUPS** initially, converging on one
  **P1** finding (the symlinked-intermediate-directory bypass) and one real
  **P2** (`isRooted` GOOS-gating).
* Both P1/P2 remediated this session (commit `f3a7fda`); full validation
  suite re-confirmed green afterward.
* Remaining P2/P3 findings (5 items) captured as stash entries per P-021 C2
  (see below) — none touch the pinned behavioral contract.

## Deferred Stash Entries (P-021 C2, threadless/pre-PR path)

| ID | Priority | Summary |
|---|---|---|
| `4A91CA81` | low | `NewRoot` drops cause on `apperr.New` construction (plan-mandated mechanism; deferred, requires deliberation) |
| `7774C9CA` | low | Go-idiom P3 advisories: `splitPath` wrapper, `Kind` `Stringer`, `kindSentinel` bounds/drift test |
| `3D61B5A8` | low | `hasPathPrefix` filesystem-root edge case; `NewRoot` no directory-type check; GO-14 dangling-symlink regression test gap |
| `9A4C8749` | medium | `Newf` missing from `.golangci.yml` `printf.funcs` allowlist (touches shared lint config — deliberation required) |
| `2130906D` | medium | Plan's Constitution Check table uses non-matching principle numbers vs. ratified `constitution.instructions.md` (Stage-owned artifact correction) |

## Branch / PR State

Branch `feat/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment`
pushed to origin; PR to be opened against `main` with merge-commit strategy
(P-009). Awaiting CI green and explicit operator merge approval (standard
sequential mode, not dark mode).

## Next Steps

1. Push branch, open PR, monitor hosted CI to green.
2. Run §1.9 pre-merge review readiness gate and P-018 Copilot-review gate
   before presenting as merge-ready.
3. Present PR readiness summary to operator; wait for explicit merge
   approval — do not auto-merge.
4. On approval: merge via merge commit, run post-merge closure (branch
   `post-merge/002-f-intercom-go-p1-error-taxonomy-and-workspace-path-containment`),
   safe-close shipment `002-S`, produce
   `docs/closure/002-S-002-F-post-merge-closure.md`, invoke P-020
   compaction, restore any parked local state.
