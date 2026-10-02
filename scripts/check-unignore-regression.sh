#!/usr/bin/env bash
set -euo pipefail

# check-unignore-regression.sh
#
# Un-ignore regression gate (011.012-T, resolves 11ECB954). Detects when a
# path that was actually ignored moves to un-ignored across a change,
# using git's OWN ignore matcher (never a gitignore engine reimplemented in
# bash or any other language) via `git check-ignore`.
#
# RULE (old-vs-new behavioural differential, NOT a textual "no negations"
# ban -- that would forbid every functional gitignore negation, since git
# only gives a `!` line effect when a preceding pattern already ignores the
# path):
#
#   No path that is actually ignored at base-ref may become un-ignored at
#   head-ref.
#
# TWO-PART CHECK:
#   Part 1 (primary): a fixed, committed denylist of paths that MUST
#     remain ignored, asserted at head-ref via `git check-ignore --no-index`
#     (works without the path needing to exist). The denylist references
#     ONLY paths guaranteed ignored by HEAD + 011.001-T -- it must never
#     depend on droppable stowaway content (revision 5 standing rule).
#   Part 2 (secondary): any path currently tracked-but-ignored, or
#     untracked-and-ignored, in the working tree is cross-checked so a
#     silent regression cannot hide behind stowaway files. KNOWN BOUNDARY:
#     this part is INERT on a fresh CI checkout (no untracked/stowaway
#     cruft exists to enumerate) -- part 1 carries the real coverage in
#     CI. This boundary is deliberate, not a bug, and is recorded here
#     rather than silently assumed away.
#
# Per-path evaluation is ALWAYS literal and individual (never a single
# batched `git check-ignore` invocation as positional arguments): `git
# check-ignore` exits 0 if ANY argument is ignored, so a batched call would
# mask an entry that silently became un-ignored. Glob-shaped denylist
# entries are passed as literal arguments (never shell-expanded).
#
# NON-VACUITY: the two counts (denylist entries evaluated; differential
# paths evaluated) are reported SEPARATELY, and the check fails if the
# DENYLIST count is zero (the differential count may legitimately be zero
# on a clean checkout -- see the boundary note above).
#
# NO WAIVER MECHANISM (deliberate, YAGNI -- HEAD has zero negations and no
# concrete waiver case exists). A deliberate un-ignore that trips this gate
# is RECORDED AND ESCALATED to the operator, never self-approved through an
# unowned bypass in a security gate.
#
# BOUNDARIES (recorded, not assumed): this checker inspects only the ROOT
# .gitignore; the CI job invoking it is pull_request-gated (a `push` event
# has no stable base-ref); and Part 2 is inert on a clean CI checkout (see
# above).
#
# M3-T7 (docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md,
# plan section C-3/M3-T7): the engine that used to run under an embedded
# interpreter now runs entirely in Go, at
# tools/gatecheck/internal/unignore (M3-T6). This wrapper is reduced to a
# caller-argument guard plus the shared build/invoke/cleanup runner
# (scripts/lib/gatecheck-run.sh, C-3): it builds the gatecheck binary once
# per invocation and forwards this script's own arguments straight through
# to `gatecheck unignore`, which absorbs the --self-test/--base-ref/
# --head-ref/-h/--help parsing itself. The interpreter
# prerequisite probe is dropped along with the embedded engine it guarded. Exit codes and
# ordered stdout/stderr bytes are preserved (ED-3 is the only permitted
# delta; see m3.md for the parent-vs-head parity evidence).
#
# Caller-argument allowlist (same root cause as
# check-retired-architecture.sh's own root-injection guard), retained as
# defense-in-depth: gatecheck_invoke prepends the trusted "--root ${ROOT}",
# and main.go's parseRoot treats that first root option as authoritative and
# rejects any later "--root"/"--root=" option. The trusted root therefore
# cannot be overridden. Unlike retired-arch (which forwards only a single
# validated mode string), unignore's real CLI surface needs multiple flags
# forwarded (--self-test, --base-ref <ref>, --head-ref <ref>), so instead of
# collapsing to one argument, every forwarded argument is checked against a
# literal "--root"/"--root=*" denylist before being passed through.
#
# Usage:
#   scripts/check-unignore-regression.sh --self-test
#     Landing precondition (011.012-T revision 4/5): asserts every
#     denylist entry is ignored at HEAD -- this is the guard that prevents
#     a mis-specified denylist from ever shipping again. Also runs
#     synthetic accept/reject scenarios (built as ephemeral, throwaway git
#     repos, since a full ignore-differential scenario is a git-history
#     fixture, not flat text) and the real check against this repo.
#   scripts/check-unignore-regression.sh --base-ref <ref> [--head-ref <ref>]
#     Runs the real two-part check. --head-ref defaults to HEAD (the
#     current checkout). --base-ref has no default and must be resolved
#     explicitly by the caller (CI passes
#     github.event.pull_request.base.sha), matching
#     check-gitignore-append-only.sh's own fail-closed convention.

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

gatecheck_invoke unignore "$@"
exit $?
