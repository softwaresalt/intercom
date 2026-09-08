---
title: "Shipment 011-S pathsafe containment memory"
date: 2026-09-07
branch: feat/011-s-pathsafe-containment-correctness-and-cross-platform-coverage
feature: 012-F
shipment: 011-S
---

## Completed tasks

* 012.001-T through 012.009-T completed
* Red evidence captured for ancestor-walk regressions, Windows junction coverage, and stripUNCPrefix regressions
* Final verification completed with `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./internal/pathsafe ./tests/integration`, and `bash ./scripts/check-write-path-precondition.sh --self-test`

## Files changed

* `internal/pathsafe/pathsafe.go`
* `internal/pathsafe/root.go`
* `internal/pathsafe/contain_test.go`
* `internal/pathsafe/symlink_test.go`
* `internal/pathsafe/junction_windows_test.go`
* `internal/pathsafe/root_test.go`
* `internal/pathsafe/root_windows_test.go`
* `internal/pathsafe/lexical_test.go`
* `tests/integration/output_path_guard_test.go`
* `tests/integration/symlink_privilege_windows_test.go`
* `tests/integration/symlink_privilege_other_test.go`

## Decisions

* `checkSymlinkEscape` now uses a final-component `os.Stat` probe, strict-ancestor `os.Lstat` probes, ENOENT-only ascent, and explicit rejection of non-directory strict ancestors
* Windows junction coverage required an additional `os.Readlink`-based reparse-point resolution path because `filepath.EvalSymlinks` does not resolve a junction when given the junction path itself on this host
* `stripUNCPrefix` now re-forms extended UNC paths and retains non-Windows-absolute stripped results unchanged
* Integration symlink privilege handling now mirrors the internal fail-closed predicate split by build tag

## Observed discrepancies

* 012.009-T referenced `TestNormalizeAcceptsCleanEquivalentForms`, but the current file's truncated comment sits above `TestNormalizePreservesSafeInteriorTraversal`; the comment was repaired at the actual current function anchor

## Post-implementation review (Ship, same session)

* Ran multi-persona review (Security, Correctness, Go, Scope Boundary,
  Maintainability) against the full diff before PR creation.
* Security + Maintainability reviewers independently converged on the
  same finding in `checkSymlinkEscape`: a manual `os.Readlink` +
  relative-target rejoin + second `filepath.EvalSymlinks` call was
  provably redundant with the immediately preceding
  `filepath.EvalSymlinks(ancestor)` call. Removed in commit `9158a05`,
  along with two Go-idiom nits (os.IsNotExist -> errors.Is normalization
  in new junction test cleanup; missing doc comment on a duplicated
  Windows privilege-error constant). Re-verified full green
  build/vet/test/`-race` after the cleanup — no test assertions changed.
* All other findings (byte-identical GO-14/BF5DE670 register text,
  012.009-T's function-name deviation) were independently verified by
  Ship directly (`git diff`, file inspection) rather than taken on
  trust.

## Next steps

* Backlogit status moves are complete for 012.001-T through 012.009-T;
  committed alongside this memory update.
* Proceeding to adversarial multi-model consensus review, then PR
  creation, CI, Copilot review loop, and merge.