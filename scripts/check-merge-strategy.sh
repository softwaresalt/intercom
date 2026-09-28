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

if command -v python3 >/dev/null 2>&1; then
  PYTHON_BIN=python3
elif command -v python >/dev/null 2>&1; then
  PYTHON_BIN=python
else
  echo "python3 or python is required" >&2
  exit 2
fi

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

FIXTURE_DIR="scripts/testdata/mergestrategy"

# evaluate_json <json-file>
#
# Reads JSON from the given file. Prints "<VERDICT> <reason>" on a single
# line and exits 0 for PASS/SKIP, 1 for FAIL. The payload MUST be passed as
# a file: Python's stdin is occupied by the program heredoc below, so a
# payload piped or here-stringed into this function would never be read.
evaluate_json() {
  local src="$1"
  "$PYTHON_BIN" - "$src" <<'PY'
import json
import sys

src = sys.argv[1]
with open(src, 'r', encoding='utf-8') as fh:
    raw = fh.read()


def emit(verdict, reason):
    print(f"{verdict} {reason}")
    sys.exit(0 if verdict in ("PASS", "SKIP") else 1)


if not raw.strip():
    emit("SKIP", "empty API response (unauthorized or unreachable)")

try:
    data = json.loads(raw)
except json.JSONDecodeError as exc:
    emit("SKIP", f"could not parse API response as JSON: {exc}")

if (
    not isinstance(data, dict)
    or "allow_squash_merge" not in data
    or "allow_rebase_merge" not in data
):
    emit(
        "SKIP",
        "allow_squash_merge/allow_rebase_merge absent from API response "
        "(unauthorized token or field not exposed to this credential)",
    )

squash = data["allow_squash_merge"]
rebase = data["allow_rebase_merge"]

if squash is True and rebase is True:
    emit("FAIL", "both allow_squash_merge and allow_rebase_merge are true")
elif squash is True:
    emit("FAIL", "allow_squash_merge is true")
elif rebase is True:
    emit("FAIL", "allow_rebase_merge is true")
elif not isinstance(squash, bool) or not isinstance(rebase, bool):
    emit(
        "SKIP",
        "allow_squash_merge/allow_rebase_merge present but not boolean "
        f"(got {type(squash).__name__}/{type(rebase).__name__}); "
        "malformed response is never a pass",
    )
else:
    emit("PASS", "allow_squash_merge and allow_rebase_merge are both false")
PY
}

# evaluate_response <json-string>
#
# Live repo-scan transport: writes the in-memory API response to a temp
# file and evaluates it via evaluate_json. Prints the evaluator's output
# (possibly empty if the evaluator crashed); never fails by itself.
evaluate_response() {
  local tmp
  tmp="$(mktemp)"
  printf '%s' "$1" > "$tmp"
  local output
  output="$(evaluate_json "$tmp" || true)"
  rm -f "$tmp"
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
    # file-path check above.
    local transport_output
    transport_output="$(evaluate_response "$(cat "$path")")"
    local transport_actual="${transport_output%% *}"
    if [ "$transport_actual" = "$expected" ]; then
      echo "PASS $name (repo-scan transport): got $transport_actual as expected"
    else
      failures+=("$name (repo-scan transport): expected $expected, got $transport_actual ($transport_output)")
    fi
  done <<< "$discovered"

  # An empty or unrecognized evaluator verdict must map to exit 2, never
  # to success.
  local rc
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

  local output
  output="$(evaluate_response "$response")"
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
