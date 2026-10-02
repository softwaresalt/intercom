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
#   os.Lchown os.Chtimes
#
# ACCEPTED RESIDUAL (recorded in the shipment plan): (*os.File).Write* is
# not textually decidable by a grep-shaped detector without full type
# information (a method call site names no package qualifier), so this
# detector intentionally does not attempt it and relies on the
# package-qualified constructors above instead.
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
