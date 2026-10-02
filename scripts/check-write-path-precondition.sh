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
# (26 selectors, matching writepath.Selectors exactly; the last six were
# appended by 034.004-T. syscall.CreateFile is reported unless the D-2'
# predicate proves the exact metadata-only reparse-probe call shape.)
#
# ACCEPTED RESIDUAL (recorded in the shipment plan): (*os.File).Write* is
# not textually decidable by a grep-shaped detector without full type
# information (a method call site names no package qualifier), so this
# detector intentionally does not attempt it and relies on the
# package-qualified constructors above instead.
#
# RESIDUAL EVASION SURFACE (034.007-T, AC-4.7 -- recorded, never silently
# left unhandled):
#   1. Named import aliases -- KNOWN OPEN, pending Unit E (feature 049-F,
#      shipment 039-S). An aliased import of a write-capable package
#      (import o "os" -> o.WriteFile(...), or an alias of database/sql /
#      go.etcd.io/bbolt) is not detected.
#   2. pathsafe.NewRoot receiver / Root.Resolve first-caller tracking --
#      KNOWN OPEN, pending Unit E (feature 049-F, shipment 039-S). No
#      tripwire exists. The pathsafe risk-register triggers ("forced the
#      moment a real write path exists", root.go) and feature 038-F's
#      "once a live caller exists" trigger stay awaited, not monitored.
#   3. Dot-imports, blank imports, and local identifiers shadowing a
#      package name (including a local syscall identifier, which the D-2'
#      predicate trusts by spelling).
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
#   6. Write primitives outside the selector set are not detected:
#      ioutil.WriteFile, ioutil.TempFile, ioutil.TempDir; syscall write and
#      namespace calls other than syscall.CreateFile and syscall.Write
#      (syscall.WriteFile, Open, Unlink, Rename, Mkdir, CreateHardLink,
#      DeleteFile); and golang.org/x/sys/windows and golang.org/x/sys/unix
#      equivalents. None occurs in internal/** or cmd/** at 9b299c8.
#      Widening the selector set is out of D-031-2's scope and is tracked
#      as stash entry 458F9385.
#   7. Selector split by a newline or comment -- KNOWN OPEN, closed by
#      Unit E's AST engine (feature 049-F). Go inserts no semicolon after
#      ".", so "syscall." + newline + "CreateFile(...)" and
#      "os./**/WriteFile(...)" are valid Go that gofmt preserves; after
#      masking neither contains the contiguous selector text, so every
#      selector is evaded.
#
# RESIDUAL-RISK STATEMENT: Until feature 049-F ships, this gate is a
# qualified-selector tripwire, not a complete mechanical proof. Items 1, 2,
# 6 and 7 are known open fail-open surfaces. At 9b299c8 there are zero
# aliased write-capable imports, zero production Root.Resolve callers and
# zero item-6 primitives in internal/**/cmd/**, so items 1, 2 and 6 are
# not exploited today. None of the four is mechanically guarded. The
# compensating control is human and agent PR review against this list.
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
