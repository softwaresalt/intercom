# Ship session checkpoint — 016-S pre-PR

**Date**: 2026-09-10
**Agent**: Ship (dark-factory execution)
**Shipment**: 016-S | **Feature**: 017-F | **Branch**: `feat/016-s-pathsafe-error-surface-parity-and-windows-long-path-precision`

## Items completed

- **017.001-T** (high): `checkSymlinkEscape`'s `canonicalizeReparse` error branch
  now routed through the existing `wrapPathError` helper (parity with
  `NewRoot`'s identical call site). No containment verdict change. New
  privilege-independent (junction-based) Windows test locks the
  `*fs.PathError` surface. Commit `6dd3ba1`.
- **017.002-T** (medium): characterization test + doc-comment correction for
  `addLongPathPrefix`'s `\.\` device-namespace exclusion branch — corrects
  the stale "unreachable" premise (measured: `filepath.IsAbs` returns true
  for all `\.\` forms tested), documents it as reachable-but-currently-
  redundant (D-4), no panic/assertion added, branch retained. Commit
  `a5a05d2`.
- **017.003-T** (medium, depends on 017.002-T): `addLongPathPrefix`'s
  MAX_PATH threshold comparison switched from UTF-8 byte length to a new
  `utf16Len` (UTF-16 code-unit count). Precision improvement only (old
  proxy could only over-prefix, never under-prefix). Verified against
  stdlib `utf16.Encode` reference including invalid-UTF-8 input. Commit
  `a5a05d2` (combined with 017.002-T per the task's own documented
  same-function/same-doc-comment merge-coordination note).
- Feature 017-F and all three tasks moved to `done`. Backlog bookkeeping
  committed (`chore(backlog): mark ... done`).

## Review posture

Local multi-persona / adversarial review completed BEFORE any PR creation
(operator directive): general code-review agent, security-review agent,
and Go Reviewer agent — all three independent passes returned **PASS,
merge-ready**, zero P0/P1/P2 findings.

P3 findings disposition:
- Missing invalid-UTF-8 test case for `utf16Len` reference test —
  **remediated** (added to `TestUTF16LenMatchesStdlibEncoding`).
- Missing comment explaining why the sibling `os.Lstat` error branch
  doesn't use `wrapPathError` — **remediated** (comment added).
- Suggestion to add a byte-length fast-path before the UTF-16 rune scan —
  **out of scope (P-021 C1)**: this re-proposes an optimization option
  ("non-allocating bespoke counter") the plan already explicitly withdrew
  as gold-plating at plan review. **Deferred via P-021 C2 capture**: stash
  entry `475E76D2` (threadless path, pre-PR local finding, no thread to
  reply to/resolve).
- Pre-existing convention (raw `t.Fatalf` vs. testify) — no action, not a
  regression introduced by this diff.

## Acceptance criteria verification (feature-level, AC-F1..AC-F5)

- **AC-F1**: `scripts/check-write-path-precondition.sh` exits 0; `--self-test`
  passes 5/5. Verified.
- **AC-F2**: `internal/pathsafe/root.go` risk register untouched (file has
  zero diff in this shipment). Verified.
- **AC-F3**: independent grep for
  `os.WriteFile/Create/CreateTemp/OpenFile/Remove/RemoveAll/Rename/Mkdir/MkdirAll/Symlink/Link/Chmod/Truncate`
  across non-test `.go` files under `internal/**` and `cmd/**` — zero
  matches (one pre-existing illustrative comment mentioning `os.Mkdir` in
  prose, not code). Verified.
- **AC-F4**: stash entries `BF5DE670` and `6B751D8B` were NOT touched or
  archived this session (no `custom_fields.source_stash_id` present on
  017-F, so post-merge source-artifact cleanup has nothing to act on for
  this feature — correctly avoids archiving either entry). `7ADAC481` was
  already archived by Stage prior to this session.
- **AC-F5**: edits confined to `internal/pathsafe/pathsafe.go`,
  `internal/pathsafe/reparse_windows.go`, and their test files (plus one new
  test file, `symlink_windows_test.go`). `root.go`, `canonicalizeReparse`,
  `lexical.go`, `reparse_other.go`, CI workflows, `go.mod`, and `scripts/`
  are untouched. Verified via `git diff --stat` across both commits.

## Quality gates (final pre-PR run)

- `gofmt -l .` — clean.
- `go vet ./...` — clean.
- `go build ./...` — clean.
- `go test ./...` — all packages pass (`cmd/intercom`, `cmd/intercom-ctl`,
  `internal/apperr`, `internal/config`, `internal/copilotprobe`,
  `internal/pathsafe`, `tests/integration`).
- `go test ./internal/pathsafe/... -race` — clean.

## Branch / topology state

- Branch: `feat/016-s-pathsafe-error-surface-parity-and-windows-long-path-precision`,
  created from `main` at `4250711`.
- `pipeline-topology` gate: PASS at `pre_claim` (x2), `post_claim`
  (`CLAIM_VERIFY_OK`), and `lifecycle` (pre-PR).
- Single worktree; no parallel/ambiguous worktrees.
- Shipment 016-S: `active` (claimed, sole active shipment verified).

## Next steps

1. Prepare PR body with Local Review Readiness block (§1.9).
2. Push branch, invoke pr-lifecycle to create PR.
3. Wait for Copilot review completion (P-018 mandatory per operator);
   remediate/defer per P-021, reply + resolve threads via `gh api graphql`.
4. P-014 operator-approval gate — merge is pre-authorized for this
   shipment's PRs per dark-mode activation record; still re-verify §1.9 +
   P-018 immediately before merge (last-mile re-check).
5. Merge via merge-commit only (P-009). Admin fallback NOT authorized this
   run — any block requiring bypass halts for operator.
6. Post-merge closure: runtime-verification (N/A — no runtime surface
   touched, this is internal library logic only, no CLI/config/behavior
   surface change), operational-closure, shipment-reconcile safe-close,
   compound-refresh if warranted, mandatory compact-context, backlog index
   resync.

## Residual risks / known deferred items

- Deferred stash entry `475E76D2`: byte-length fast-path micro-optimization
  for `utf16Len`, out of scope, low priority, Stage disposition pending.
- Feature-level AC-F1..AC-F4 preserved (BF5DE670 trigger-gated write-path
  hardening remains explicitly deferred, not fired this cycle).
