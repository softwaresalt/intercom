#!/usr/bin/env bash
set -euo pipefail

# check-merge-strategy.sh
#
# Merge-strategy structural verification gate (031-F / U1-T1, shipment
# 028-S). Constitution Principle XI / P-009 require merge-commit-only PR
# merges. This checker queries the live GitHub repository's merge-strategy
# settings and asserts allow_squash_merge == false and
# allow_rebase_merge == false.
#
# Authority split (deliberation D-5): flipping the settings themselves is
# GitHub repository administration, outside both Stage's and Ship's role
# boundary. This script implements ONLY the standing verification -- it
# never mutates a repository setting (AC-1.7).
#
# M3-T11 (docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md,
# plan section C-3/M3-T11): evaluate_json()'s Python heredoc is replaced by
# `gatecheck_invoke merge-strategy-evaluate "$src"`
# (tools/gatecheck/internal/mergestrategy.Evaluate, M3-T9). The
# surrounding bash orchestration (run_self_test, run_repo_scan,
# evaluate_response, verdict_exit_code) is UNCHANGED by this switch -- only
# evaluate_json()'s own internals move to Go, and the python3/python
# interpreter probe at the top of this script is dropped along with it.
# gatecheck_build runs ONCE at script start, before either mode branch;
# evaluate_json stays callable inside `$(... || true)` exactly as before,
# since gatecheck_invoke's own exit status becomes evaluate_json's exit
# status. Exit codes and ordered stdout/stderr bytes are preserved (ED-2
# and ED-8 are the only permitted deltas; see m3.md for the parent-vs-head
# parity evidence).
#
# M3-T12 (fold 150364D2, R-12): (1) every evaluate_response temp file is
# tracked in MERGE_STRATEGY_TMP_FILES and removed by this script's single
# composed cleanup() trap, never by an imperative `rm -f` inside the
# function body, so a file is never leaked even if evaluate_response exits
# early; (2) the self-test table additionally asserts
# verdict_exit_code's PASS->0, SKIP->0 and FAIL->1 rows explicitly,
# alongside the pre-existing unrecognized->2 assertion; (3) a
# `git rev-parse --show-toplevel` guard runs BEFORE gatecheck_build (ED-1):
# a run outside a git checkout now exits 2 with an `::error::` message
# instead of propagating git's own `set -e` failure code, so no build is
# ever attempted outside a checkout. See m3.md for the red-first proof
# (verdict_exit_code mapping deliberately broken in a scratch copy to show
# the new table rows failing) and the ED-1 non-checkout demonstration;
# both are evidence for THIS task, kept separate from M3-T11's own
# evidence section.
#
# Usage:
#   scripts/check-merge-strategy.sh
#     Resolves the current repository via `gh repo view` and queries
#     `gh api repos/{owner}/{repo}`, then asserts allow_squash_merge == false
#     and allow_rebase_merge == false. Prints one of PASS / FAIL / SKIP plus
#     a reason. Exit code 0 for PASS or SKIP, exit code 1 for FAIL, exit
#     code 2 if the evaluator produces no recognized verdict. An
#     unauthorized, failed, or malformed API read (including non-boolean
#     field values, unless the other field is already `true`, which is a
#     FAIL) is reported SKIPPED with a
#     reason and is NEVER treated as a pass (AC-1.3); it also never fails
#     the job by itself, so a default GITHUB_TOKEN that cannot read these
#     fields does not block CI (see docs/decisions and 031.002-T /
#     031.005-T for the credential finding and promotion trigger).
#   scripts/check-merge-strategy.sh --self-test
#     Verifies every committed scripts/testdata/mergestrategy/*.json fixture
#     against its expected verdict (encoded in the filename prefix:
#     pass-/fail-/skip-). Each fixture is checked twice: once by file path
#     and once through the same in-memory response transport the live
#     repo scan uses. Also asserts an unrecognized verdict maps to exit 2.
#     Drives stubbed JSON so the checker is testable
#     with no network and no token (AC-1.1). Does not query the live
#     repository.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gatecheck-run.sh
source "${SCRIPT_DIR}/lib/gatecheck-run.sh"

# ED-1 (M3-T12, 150364D2 item 3): a run outside a git checkout must exit 2
# with a message, rather than propagating git's own `rev-parse` failure
# code under `set -e`. This guard runs BEFORE gatecheck_build, so no build
# is ever attempted outside a checkout.
if ! ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
  echo "::error::check-merge-strategy.sh must be run from inside a git checkout" >&2
  exit 2
fi
cd "$ROOT"

# MERGE_STRATEGY_TMP_FILES (M3-T12, 150364D2 item 1): every temp file used
# for evaluate_response's live-transport payload is appended here (by the
# CALLER, in the parent shell -- never inside evaluate_response itself,
# since every call site invokes it via `$(...)`, which forks a subshell
# whose array mutations are invisible to the parent) and removed only by
# the single composed cleanup() trap below, never by an imperative `rm -f`
# inside any function body, so a file is never leaked even if a call
# exits early.
MERGE_STRATEGY_TMP_FILES=()

cleanup() {
  local f
  for f in "${MERGE_STRATEGY_TMP_FILES[@]:-}"; do
    [ -n "$f" ] && rm -f -- "$f"
  done
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

FIXTURE_DIR="scripts/testdata/mergestrategy"

# evaluate_json <json-file>
#
# Reads JSON from the given file path via the Go evaluator
# (gatecheck merge-strategy-evaluate, M3-T9/M3-T11). Prints "<VERDICT>
# <reason>" on a single line and exits 0 for PASS/SKIP, 1 for FAIL -- the
# built binary's own exit code becomes this function's exit code, so a
# caller capturing this function's stdout via `$(... || true)` observes
# exactly the same (output, exit-code) pair the retired Python heredoc
# produced.
evaluate_json() {
  local src="$1"
  gatecheck_invoke merge-strategy-evaluate "$src"
}

# evaluate_response <tmp-file> <json-string>
#
# Live repo-scan transport: writes the in-memory API response to the given
# temp file and evaluates it via evaluate_json. Prints the evaluator's
# output (possibly empty if the evaluator crashed); never fails by itself.
#
# IMPORTANT: every call site invokes this function via `$(evaluate_response
# ...)`, and command substitution always forks a subshell in bash -- any
# `mktemp`/array-append performed INSIDE this function's own body would be
# confined to that subshell and never visible to the parent shell whose
# cleanup() EXIT trap drains MERGE_STRATEGY_TMP_FILES, silently leaking one
# temp file per call (this was a real regression in an earlier revision of
# this script, see m3.md's adversarial-review remediation notes). To keep
# M3-T12 item 1's tracked-and-trap-cleaned guarantee real, the caller
# allocates the temp file via `mktemp` and appends it to
# MERGE_STRATEGY_TMP_FILES itself, in the parent shell, BEFORE calling this
# function -- this function only ever receives an already-registered path
# and writes to it.
evaluate_response() {
  local tmp="$1"
  local payload="$2"
  printf '%s' "$payload" > "$tmp"
  local output
  output="$(evaluate_json "$tmp" || true)"
  printf '%s\n' "$output"
}

# verdict_exit_code <evaluator-output>
#
# Maps "<VERDICT> <reason>" to 0 (PASS/SKIP), 1 (FAIL), or 2 for an empty
# or unrecognized verdict (evaluator broke) -- never success.
verdict_exit_code() {
  case "${1%% *}" in
    PASS|SKIP) return 0 ;;
    FAIL) return 1 ;;
    *)
      echo "ERROR merge-strategy evaluator produced no recognized verdict" >&2
      return 2
      ;;
  esac
}

run_self_test() {
  if [ ! -d "$FIXTURE_DIR" ]; then
    echo "fixture dir not found: $FIXTURE_DIR" >&2
    exit 2
  fi

  local failures=()
  local discovered
  discovered=$(find "$FIXTURE_DIR" -maxdepth 1 -name '*.json' | sort)
  if [ -z "$discovered" ]; then
    echo "no fixtures discovered under $FIXTURE_DIR" >&2
    exit 2
  fi

  while IFS= read -r path; do
    local name
    name="$(basename "$path")"
    local expected=""
    case "$name" in
      pass-*) expected="PASS" ;;
      fail-*) expected="FAIL" ;;
      skip-*) expected="SKIP" ;;
      *)
        failures+=("$name: fixture filename must start with 'pass-', 'fail-', or 'skip-'")
        continue
        ;;
    esac

    local output
    output="$(evaluate_json "$path" || true)"
    local actual="${output%% *}"

    if [ "$actual" = "$expected" ]; then
      echo "PASS $name: got $actual as expected"
    else
      failures+=("$name: expected $expected, got $actual ($output)")
    fi

    # Same fixture through the live repo-scan transport (in-memory
    # response string), so a transport regression cannot hide behind the
    # file-path check above. The temp file is allocated and registered in
    # THIS (parent) shell scope, not inside evaluate_response's own
    # subshell -- see evaluate_response's doc comment for why.
    local transport_tmp
    transport_tmp="$(mktemp)"
    MERGE_STRATEGY_TMP_FILES+=("$transport_tmp")
    local transport_output
    transport_output="$(evaluate_response "$transport_tmp" "$(cat "$path")")"
    local transport_actual="${transport_output%% *}"
    if [ "$transport_actual" = "$expected" ]; then
      echo "PASS $name (repo-scan transport): got $transport_actual as expected"
    else
      failures+=("$name (repo-scan transport): expected $expected, got $transport_actual ($transport_output)")
    fi
  done <<< "$discovered"

  # verdict_exit_code's full exit-code table (M3-T12, 150364D2 item 2):
  # PASS->0, SKIP->0, FAIL->1, and an empty or unrecognized verdict->2,
  # never success.
  local rc
  for row in "PASS reason:0" "SKIP reason:0" "FAIL reason:1"; do
    local sample="${row%%:*}"
    local expected_rc="${row##*:}"
    if verdict_exit_code "$sample" 2>/dev/null; then rc=0; else rc=$?; fi
    if [ "$rc" = "$expected_rc" ]; then
      echo "PASS verdict_exit_code('${sample}'): exit ${rc} as expected"
    else
      failures+=("verdict_exit_code('${sample}'): expected exit ${expected_rc}, got ${rc}")
    fi
  done
  for bogus in "" "garbage verdict"; do
    if verdict_exit_code "$bogus" 2>/dev/null; then rc=0; else rc=$?; fi
    if [ "$rc" = "2" ]; then
      echo "PASS unrecognized verdict '${bogus}': exit 2 as expected"
    else
      failures+=("unrecognized verdict '${bogus}': expected exit 2, got $rc")
    fi
  done

  if [ "${#failures[@]}" -gt 0 ]; then
    printf 'FAIL %s\n' "${failures[@]}" >&2
    exit 1
  fi

  echo "self-test passed: all fixtures matched expected verdicts"
}

run_repo_scan() {
  local owner_repo
  if ! owner_repo="$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null)"; then
    echo "SKIP could not resolve repository via gh repo view (unauthenticated, gh unavailable, or not a GitHub repo)"
    return 0
  fi

  local response
  if ! response="$(gh api "repos/${owner_repo}" 2>/dev/null)"; then
    echo "SKIP gh api repos/${owner_repo} failed (unauthorized token or network unavailable)"
    return 0
  fi

  # The temp file is allocated and registered in THIS (parent) shell
  # scope, not inside evaluate_response's own subshell -- see
  # evaluate_response's doc comment for why.
  local tmp
  tmp="$(mktemp)"
  MERGE_STRATEGY_TMP_FILES+=("$tmp")
  local output
  output="$(evaluate_response "$tmp" "$response")"
  echo "$output"
  verdict_exit_code "$output"
}

case "${1:-}" in
  --self-test)
    run_self_test
    ;;
  "")
    run_repo_scan
    ;;
  *)
    echo "usage: scripts/check-merge-strategy.sh [--self-test]" >&2
    exit 2
    ;;
esac
