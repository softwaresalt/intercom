---
title: "Oracle Pin — agent-intercom Behavioral Oracle"
date: 2026-09-03
status: documentary
---

# Oracle Pin — agent-intercom Behavioral Oracle

This document records the behavioral-oracle pin for the `intercom-go` port and
the parity traceability table for phase P1 (`internal/apperr`,
`internal/pathsafe`). It is **documentary only**: the oracle is out-of-tree
and this pin must be re-verified before being trusted by any future phase
that depends on it.

## Pinned Oracle

* **Repository:** `softwaresalt/agent-intercom`
* **Pinned commit:** `41df772`
* **Local read-only path:** `C:\Source\GitHub\intercom` (this workspace's
  sibling checkout, mounted read-only for parity verification — never
  modified, staged, or committed from `intercom-go`)
* **Access:** read-only, out-of-tree. No write, no `git add`/`commit`/
  `checkout`/`stash`/`clean`/`restore`, and no branch operations were issued
  against this path during phase P1.

## Rejected Mis-Mount

`references/herdr` in this repository is **not** the agent-intercom oracle.
It is an unrelated Rust terminal multiplexer, tracked as:

* **Repository:** `ogulcancelik/herdr`
* **Pinned commit:** `b07ba9ce`

`references/herdr` was not read, modified, or otherwise used as a source of
behavioral truth for phase P1. It remains untouched.

## Verification Method

The oracle pin was verified by:

1. Confirming the local checkout's remote URL resolves to
   `softwaresalt/agent-intercom` and its `HEAD` is `41df772`.
2. Cross-checking the oracle's module map (`src/errors.rs`,
   `src/diff/path_safety.rs`, `src/config.rs`) against the units this phase
   ports.
3. Confirming the oracle's `src/` line count is approximately 21762 LOC.
4. Confirming the oracle's test-corpus counts match the expected
   56 / 22 / 46 / 6 distribution (total unit / integration / behavioral /
   other test groupings referenced by the source deliberation and plan).

Because the oracle lives outside this repository's version control, this
verification is a point-in-time record, not a continuously enforced
constraint. A drift-detection mechanism is explicitly deferred (see finding
ARCH-5 in the source plan's residual findings) rather than added to this
slice.

## Oracle Parity Traceability (16 tests → covering unit)

| # | Oracle test (`tests/unit/…`) | Pins | Covered by |
|---|---|---|---|
| 1 | `error_tests::acp_error_display_starts_with_acp_prefix` | prefix `acp:` | A1.1, A1.2 |
| 2 | `error_tests::acp_error_display_includes_message` | `"acp: stream closed"` | A1.2 |
| 3 | `error_tests::acp_error_message_no_trailing_period` | no trailing `.` | A1.1 |
| 4 | `error_tests::acp_error_is_distinct_from_io_error` | prefix distinctness | A1.1 |
| 5 | `error_tests::acp_error_is_distinct_from_mcp_error` | prefix distinctness | A1.1 |
| 6 | `error_tests::acp_error_implements_std_error_trait` | satisfies `error` | A1.2 |
| 7 | `error_tests::acp_error_debug_representation` | kind recoverable | A2.1, A3.3 |
| 8 | `path_validation::allows_path_inside_workspace` | relative resolves in-root | B3.1 |
| 9 | `path_validation::rejects_traversal` | `../secret.txt` | B2.1 |
| 10 | `path_validation::rejects_deep_traversal` | `src/../../secret.txt` | B2.1 |
| 11 | `path_validation::allows_relative_subdirectory` | nested relative | B3.1 |
| 12 | `path_validation::allows_dot_segment` | `./src/main.rs` | B2.3, B3.1 |
| 13 | `path_validation::rejects_workspace_root_boundary` | `subdir/../../escape.txt` | B2.1 |
| 14 | `path_validation::rejects_symlink_escape` | symlink out of root | B4.1 |
| 15 | `path_validation::path_safety_allows_non_existent_file` | non-existent accepted | B4.3 |
| 16 | `path_validation::path_safety_rejects_invalid_workspace` | bad root | B1.2 |

Go-port-specific tests beyond oracle parity (added by review findings):
B2.2 (`C:foo` drive-relative, SEC-1), B3.2 (`<root>-evil` sibling, GO-6),
B3.3 (empty candidate, SEC-2), B2.3 (`a/b/../c.txt` `Clean`-equivalence,
GO-13), A3.2 (`Wrap` nil guard, GO-3), B1.3 (`NewRoot` idempotence).

An additional Go-specific correctness gap was found and closed during
implementation, beyond the plan's original mechanism: on Windows,
`filepath.IsAbs`/`filepath.VolumeName` alone do not reject a Unix-style
rooted candidate such as `/etc/passwd` (Go treats it as drive-relative, not
absolute, on that platform). `internal/pathsafe.normalize` additionally
rejects any candidate beginning with a path separator without a volume name,
so this required negative case (plan acceptance criterion B2.2) is
rejected uniformly across platforms.

## Review-Gate Remediation (this session, post-implementation)

Independent structured review (Go, Correctness, Security, Architecture,
Constitution personas, report-only mode) surfaced two same-contract-surface
defects in `internal/pathsafe`, both remediated before merge:

* **P1 (Correctness/Security/Constitution, convergent finding):**
  `checkSymlinkEscape` originally gated symlink-escape re-validation on
  `os.Stat(resolved)` succeeding for the *full* candidate path. Because the
  dominant real-world use case is creating a *new* file (a non-existent
  leaf component), a symlinked *intermediate* directory pointing outside the
  workspace would silently bypass detection: `os.Stat` on the non-existent
  leaf returns `ENOENT`, taking the "non-existent, accept" branch without
  ever resolving the symlinked ancestor. Fixed by walking up from the
  resolved candidate to the nearest *existing* ancestor, resolving that
  ancestor's symlinks, and re-asserting containment there — regardless of
  whether the final leaf exists. Locked in by
  `TestResolveRejectsSymlinkedIntermediateDirEscapingRoot`.
* **P2 (Correctness):** `isRooted`'s leading-backslash check was
  unconditional, which would over-reject a legitimate workspace-relative
  candidate that merely begins with a literal backslash byte on non-Windows
  platforms (where `\` has no path-separator meaning). Fixed by scoping the
  backslash branch to `runtime.GOOS == "windows"`; the leading-`/` check
  remains unconditional (redundant-but-harmless on Unix, essential on
  Windows). Locked in by
  `TestNormalizeDoesNotOverRejectLiteralBackslashOnNonWindows`.

Residual P2/P3 findings from the same review round (a docs/lint-config gap
around `Newf` printf-safety, a `Kind` `Stringer` nicety, `NewRoot` not
verifying its target is a directory, a dangling-symlink regression-test gap,
and the plan's Constitution Check table using non-matching principle
numbers) do not touch the pinned behavioral contract and are captured as
deferred stash entries under P-021 rather than expanding this shipment's
scope.

## Status

This pin is documentary. Re-verify the commit, module map, and test-corpus
counts above against the live oracle checkout before relying on this record
in any later phase (P2 and beyond).
