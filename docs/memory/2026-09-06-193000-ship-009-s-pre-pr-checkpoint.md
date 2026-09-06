# Ship 009-S — Pre-PR Checkpoint

Date: 2026-09-06T19:30:00Z (session-local)
Shipment: 009-S (pathsafe and config containment hardening, 010-F)
Branch: `feat/009-s-pathsafe-and-config-containment-hardening`
Dark mode: active (scope: 009-S only; merge_approval_pre_authorized=false; admin_fallback_pre_authorized=false)

## Items completed (all 5 tasks + covering feature, all status=done/archived)

| Task | Summary | Commit |
|---|---|---|
| 010.001-T | Preserve underlying cause in NewRoot via apperr.Wrapf | e856fc4 |
| 010.002-T | Reject non-directory target in NewRoot via os.Stat | 3cd528e |
| 010.003-T | Fix hasPathPrefix doubled separator at fs root (conditional fix); drop splitPath wrapper | a3fbd5a (fix) + a2af1f9 (test coverage, folded in during 010.005-T pass) |
| 010.004-T | Add dangling-symlink regression test (GO-14) — no prod change | 8c7ce47 |
| 010.005-T | Pure-move containsDotDotSegment -> pathsafe.ContainsDotDotSegment | 52a4033 |
| 010-F | Covering feature — marked done, harness_status=passing | — |

## Branch state

- Base: `main` @ 44b1cda5531d34bbe32b9d2b5c277e0a137fece3 (fetched, up to date, no divergence)
- HEAD: `52a4033f86dec386205646766b5a9ee955d25f5a`
- 6 commits ahead of main, all source-only (backlog metadata changes deferred to Step 6 safe-close)
- Pre-existing dirty/untracked carryover (operator-preserved, stash 9F9A3CB2 for .gitignore/start.ps1; unrelated
  environment/tooling files) verified isolated from `internal/pathsafe/`, `internal/apperr/`, `internal/config/`
  scope and NOT staged/committed in any shipment commit.

## Quality gates (final pass, HEAD 52a4033)

- `go vet ./...` — PASS
- `gofmt -l .` — clean (no files listed)
- `go build ./...` — PASS
- `go test ./...` — PASS (all packages, including tests/integration)

## Review gate (report-only, HEAD 52a4033)

Reviewers: Security Reviewer, Go Reviewer, Correctness Reviewer, Maintainability Reviewer, Constitution Reviewer.
All five: **READY_WITH_FOLLOWUPS**. Zero P0/P1 findings across all reviewers. All P2/P3 findings classified
out-of-scope under P-021 C1 and captured as deferred stash entries (threadless path, pre-PR):

| Stash ID | Theme | Priority |
|---|---|---|
| BF5DE670 | Consolidate pathsafe TOCTOU/dangling-symlink-write-through/hardlink (SEC-5) tracking | medium |
| E89E5D00 | internal/pathsafe Windows-specific security branches never exercised by ubuntu-only CI | medium |
| 798002CB | pathsafe.ContainsDotDotSegment exported with weaker guarantee than Root/Resolve | low |
| 8D953C4B | config.Validate rule 7 database.path absolute-path gap vs Principle III (pre-existing) | medium |
| 8472E0A1 | Consolidated minor P3 code-quality/test-coverage cluster (8 sub-items) | low |

No P0/P1 remain unresolved. No in-scope fixes were deferred (C3 symmetric guard satisfied — every review finding was
genuinely outside the 010-F U1-U6 authorized change set).

## Next steps

1. Push branch, create PR via pr-lifecycle skill with `## Local Review Readiness` block (HEAD 52a4033,
   outcome READY_WITH_FOLLOWUPS, 5 follow-up stash IDs above, full-build evidence as captured above).
2. Run P-018 copilot-review gate and P-014 §1.9 readiness gate before presenting as merge-ready.
3. **Halt at merge approval gate** — `merge_approval_pre_authorized=false`. Do not merge without explicit operator
   approval. Preserve branch/PR state for resumption.
4. On approval: verify merge commit strategy (P-009), re-run last-mile P-018/§1.9 checks, merge, then proceed to
   Step 6 post-merge closure (post-merge/009-s-pathsafe-and-config-containment-hardening branch, backlog safe-close,
   operational-closure, compound-refresh, compact-context, source artifact cleanup).
