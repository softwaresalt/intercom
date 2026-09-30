#!/usr/bin/env bash
# gatecheck-run.sh — shared runner sourced by every gate wrapper script.
#
# This file is never executed directly; wrappers `source` it. It anchors the
# gatecheck tool source relative to its OWN location (never the caller's
# working directory), builds the tool exactly once per wrapper process into
# an OS-managed scratch directory, and invokes it with streamed (never
# buffered or captured) stdout/stderr.
#
# See docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md
# section C-3 for the full design contract (M1-T9).
#
# Provided to wrappers after sourcing:
#   GATECHECK_SRC     - absolute repo root, derived from this file's path.
#   gatecheck_build   - builds the binary once; sets GATECHECK_TMP/GATECHECK_BIN.
#   gatecheck_invoke  - runs a gatecheck subcommand; streams stdout/stderr;
#                       returns (never exits) the binary's status.
#   gatecheck_cleanup - removes GATECHECK_TMP. Idempotent: safe to call more
#                       than once, and safe when the directory was never
#                       created.
#
# This file never installs a trap. Each wrapper installs exactly three, and
# composes its own cleanup work with gatecheck_cleanup inside one function:
#   cleanup() { <wrapper's own cleanup>; gatecheck_cleanup; }
#   trap cleanup EXIT
#   trap 'exit 130' INT
#   trap 'exit 143' TERM

# Anchored on this file's location, not $PWD (matches how the Python engines
# anchor on __file__/SCRIPT_DIR). $(dirname "${BASH_SOURCE[0]}") is
# scripts/lib; ../.. is the repo root.
GATECHECK_SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# gatecheck_build builds the gatecheck binary into a fresh mktemp -d
# directory and sets GATECHECK_TMP/GATECHECK_BIN for gatecheck_invoke and
# gatecheck_cleanup to use. It is called once per wrapper process, after any
# wrapper precondition that must exit first (for example an ED-1
# `git rev-parse` guard).
#
# On build failure it prints an `::error::` line to stderr and returns 2; it
# never calls exit itself, so the wrapper controls its own exit path.
gatecheck_build() {
  GATECHECK_TMP="$(mktemp -d)"
  local goexe
  goexe="$(env -u GH_TOKEN -u GITHUB_TOKEN GOTOOLCHAIN=local GOFLAGS=-mod=readonly go env GOEXE)"
  GATECHECK_BIN="${GATECHECK_TMP}/gatecheck${goexe}"
  local rc=0
  env -u GH_TOKEN -u GITHUB_TOKEN GOTOOLCHAIN=local GOFLAGS=-mod=readonly \
    go -C "${GATECHECK_SRC}" build -trimpath -o "${GATECHECK_BIN}" ./tools/gatecheck || rc=$?
  if [ "${rc}" -ne 0 ]; then
    echo "::error::gatecheck build failed (exit ${rc})" >&2
    return 2
  fi
  return 0
}

# gatecheck_invoke <subcommand> [args...] runs the built binary with the
# common --root flag appended after the subcommand. GH_TOKEN and GITHUB_TOKEN
# are stripped from its environment: no gate engine ever needs a token.
# stdout and stderr are streamed straight through — never captured into a
# variable or command substitution here — so it is safe to call this
# function itself inside `if`, `|| rc=$?` or `$(...)` at the call site.
gatecheck_invoke() {
  local subcommand="$1"
  shift
  env -u GH_TOKEN -u GITHUB_TOKEN "${GATECHECK_BIN}" "${subcommand}" --root "${ROOT}" "$@"
}

# gatecheck_cleanup removes GATECHECK_TMP. Idempotent: safe to call twice in
# a row, and safe when GATECHECK_TMP was never set (gatecheck_build was
# never reached).
gatecheck_cleanup() {
  if [ -n "${GATECHECK_TMP:-}" ] && [ -d "${GATECHECK_TMP}" ]; then
    rm -rf -- "${GATECHECK_TMP}"
  fi
  GATECHECK_TMP=""
}
