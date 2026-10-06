#!/usr/bin/env bash
set -euo pipefail

# check-write-path-precondition.sh
#
# Mechanical CI gate (011.004-T, partially resolves BF5DE670) making the
# database.path write-path precondition (see the Constitution Check
# exception recorded in internal/config/validate.go, rule 7, and the
# consolidated risk register in internal/pathsafe's package doc, root.go)
# a machine-enforced invariant rather than a documentation-only promise.
#
# INVARIANT: no destructive filesystem write primitive exists in any
# non-test Go file under internal/** or cmd/**. J1 in the shipment plan
# names this the enforcing gate.
#
# Detector scope: qualified selectors only (never a bare identifier
# alternation, which would false-positive on names like
# copilot.CreateSession):
#   os.WriteFile os.Create os.OpenFile os.Remove os.RemoveAll os.Rename
#   os.Mkdir os.MkdirAll os.Symlink os.Chmod os.Truncate io.Copy
#   sql.Open bbolt.Open os.CreateTemp os.MkdirTemp os.Link os.Chown
#   os.Lchown os.Chtimes syscall.CreateFile syscall.Write os.OpenRoot
#   os.Root io.CopyN io.CopyBuffer
#   ioutil.WriteFile ioutil.TempFile ioutil.TempDir
#   syscall.WriteFile syscall.Open syscall.Unlink syscall.Rename syscall.Mkdir
#   syscall.Rmdir syscall.CreateHardLink syscall.DeleteFile syscall.MoveFile
#   syscall.RemoveDirectory syscall.CreateDirectory syscall.CreateSymbolicLink
#   syscall.Truncate syscall.Creat
#   windows.WriteFile windows.CreateFile windows.DeleteFile windows.MoveFile
#   windows.MoveFileEx windows.CreateDirectory windows.RemoveDirectory
#   windows.CreateHardLink windows.CreateSymbolicLink windows.SetEndOfFile
#   windows.SetFileInformationByHandle
#   unix.Open unix.Openat unix.Openat2 unix.Creat unix.Write unix.Pwrite
#   unix.Unlink unix.Unlinkat unix.Rename unix.Renameat unix.Renameat2
#   unix.Mkdir unix.Mkdirat unix.Rmdir unix.Link unix.Linkat unix.Symlink
#   unix.Symlinkat unix.Truncate unix.Ftruncate unix.Chmod unix.Fchmodat
# (76 selectors, matching writepath.Selectors exactly. The first 20 were
# ported from the retired Python list, six were appended by 034.004-T, and
# the final 50 by 049.007-T. syscall.CreateFile is reported unless the D-2'
# predicate proves the exact metadata-only reparse-probe call shape.)
#
# ACCEPTED RESIDUAL (recorded in the shipment plan): (*os.File).Write* is
# not textually decidable by a grep-shaped detector without full type
# information (a method call site names no package qualifier), so this
# detector intentionally does not attempt it and relies on the
# package-qualified constructors above instead.
#
# CLOSED RESIDUALS (Unit E):
#   1. Named import aliases -- closed by 049.004-T and
#      reject-alias-os-writefile.go.
#   7. Selector split by a newline or comment -- closed by 049.002-T and
#      pinned by 049.004-T's reject-split-selector.go.
#
# RESIDUAL EVASION SURFACE (034.007-T, AC-4.7 -- recorded, never silently
# left unhandled):
#   2. pathsafe.NewRoot receiver / Root.Resolve first-caller tracking --
#      same-function bindings are closed by 049.005-T and
#      reject-resolve-first-caller.go. Cross-function flows, struct fields,
#      package-level variables, and method values remain residuals.
#   3. Dot imports of write-capable packages are reported at the import spec,
#      closed by 049.004-T and reject-dot-import-os.go. Blank imports are
#      inert. Local identifiers shadowing a package name (including a local
#      syscall identifier, which the D-2' predicate trusts by spelling)
#      remain residual.
#   4. os.Root method calls and (*os.File).Write* (the latter recorded
#      above).
#   5. Undecidable or non-simple call shape -> rejected. A
#      syscall.CreateFile reached via a wrapper or function value, with an
#      extent that does not balance, with any argument that is not a bare
#      operand (call, composite literal, index, string or raw string), or
#      with any access, disposition or flags spelling outside the D-2'
#      predicate's exact tokens is rejected, not exempted (D-2' rules 2-7).
#      This is a
#      false-positive surface: a future legitimate metadata-only call in
#      such a shape trips the gate and needs an explicit, reviewed
#      widening. It is never a silent hole.
#   6. Write primitives outside the selector set are not detected. E-T7
#      closes the 50 enumerated ioutil, syscall, x/sys/windows and x/sys/unix
#      primitives in 049.007-T and fixtures reject-ioutil-write-primitives.go,
#      reject-syscall-namespace-primitives.go and
#      reject-xsys-write-primitives.go. The uncovered same-family remainder
#      includes metadata and attribute writes
#      (syscall.Chmod/Fchmod/Chown/Utimes/SetFileAttributes,
#      unix.Fchmod/Chown/Fchown/Lchown/Utimes/Setxattr,
#      windows.SetFileAttributes), positional and vector writes
#      (syscall.Pwrite, unix.Writev/Pwritev), syscall.Ftruncate,
#      syscall.Link/Symlink, unix.Mknod/Mknodat, and any other
#      un-enumerated write-capable symbol in those packages. This remainder
#      is deferred as stash entry C0D28448.
#   8. Dynamic invocation via syscall.NewLazyDLL / LazyProc.Call,
#      syscall.Syscall* and the golang.org/x/sys equivalents can reach any
#      OS write API without a write-named selector. syscall.NewLazyDLL is
#      live in production, so it cannot become a finding without an
#      allowance design. It is outside Unit E and deferred as stash entry
#      FE2F02FF.
#
# RESIDUAL-RISK STATEMENT: This gate remains a narrow Go AST tripwire, not a
# complete mechanical proof. Items 2, 4, 5, 6 and 8, plus local package-name
# shadowing in item 3, remain residual surfaces. The compensating control is
# human and agent PR review against this list.
#
# ANTI-GOAL: no TOCTOU/hardlink mitigation mechanism is added here. This
# script only detects the FIRST write path arriving; it does not mitigate
# anything once one does.
#
# M1-T10 (docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md,
# plan section C-3/M1-T10): the engine that used to run under an embedded
# interpreter now runs entirely in Go, at
# tools/gatecheck/internal/writepath. This wrapper is reduced to argument
# parsing plus the shared build/invoke/cleanup runner
# (scripts/lib/gatecheck-run.sh, C-3): it builds the gatecheck binary once
# per invocation and forwards this script's own first argument straight
# through as `gatecheck write-path`'s mode flag, exactly mirroring the
# retired dispatch below (repo mode, --self-test, --self-test-integrity,
# or an unrecognized flag -> usage + exit 2). Exit codes and ordered
# stdout/stderr bytes are preserved (ED-3/ED-5 are the only permitted
# deltas; see m1.md for the parent-vs-head parity evidence).
#
# Usage:
#   scripts/check-write-path-precondition.sh
#     Scans tracked, non-test Go files under internal/** and cmd/**. Exits 0
#     when no write primitive is found.
#   scripts/check-write-path-precondition.sh --self-test
#     Verifies every committed scripts/testdata/writepath/*.go fixture
#     against its expected verdict, then verifies the real tracked tree
#     passes. Exits 0 only when both succeed.
#   scripts/check-write-path-precondition.sh --self-test-integrity
#     Integrity-only mode (030.001-T, mirrors
#     check-retired-architecture.sh's --self-test-integrity split): runs
#     ONLY the fixture self-test above. Does NOT run the repo scan. Used
#     by ci.yml's unconditionally-blocking integrity step so it stays
#     blocking while a separately WRITE_PATH_GATE_ADVISORY-toggled verdict
#     step owns the repo scan. AC-0.5: verdicts at bound snapshot
#     14d44e3c are unchanged -- this mode addition changes no selector or
#     masking logic.

for arg in "$@"; do
  case "$arg" in
    --root | --root=*)
      echo "::error::unrecognized argument: $arg" >&2
      exit 2
      ;;
  esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gatecheck-run.sh
source "${SCRIPT_DIR}/lib/gatecheck-run.sh"

ROOT="$(git rev-parse --show-toplevel)"

cleanup() {
  gatecheck_cleanup
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

gatecheck_build
build_rc=$?
if [ "${build_rc}" -ne 0 ]; then
  exit "${build_rc}"
fi

gatecheck_invoke write-path "$@"
exit $?
