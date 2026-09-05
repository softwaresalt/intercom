#!/usr/bin/env bash
# check-gitignore-append-only.sh
#
# Enforces invariant I6: .gitignore is append-only. Precise definition
# (007-F / U4a-U4b, resolves stash F7C6420D):
#
#   I6 holds iff the OLD file's non-comment, non-blank lines appear as an
#   ordered SUBSEQUENCE of the NEW file's non-comment, non-blank lines.
#
# Insertions anywhere pass; deletions and reorders fail. One mechanical
# predicate, not ad-hoc cases.
#
# STATUS (007.004-T / U4a): this file currently ships the CLI surface,
# fixture harness, and --self-test driver ONLY. The core subsequence
# predicate below is an INTENTIONAL STUB (Principle II, test-first): the
# fixtures under scripts/testdata/gitignore/ are authored and committed
# BEFORE the detection logic exists, and MUST fail against this stub to
# prove they assert something real. 007.005-T (U4b) replaces
# `subsequence_check` with the real implementation and wires the
# git-ref comparison mode; nothing else in this file's interface changes.
#
# Usage:
#   check-gitignore-append-only.sh --self-test
#       Runs every fixture pair under scripts/testdata/gitignore/ and
#       reports each fixture by name (AC-2). Exits non-zero if any
#       fixture's actual result does not match its recorded expectation.
#
#   check-gitignore-append-only.sh [--base-ref <ref>] [--head-ref <ref>]
#                                   [--file <path>] [--allow-no-merge-base]
#       Compares --file (default: .gitignore) as it existed at --base-ref
#       against --head-ref (default: HEAD). --base-ref has NO default: it
#       must be resolved explicitly by the caller (CI passes
#       `github.event.pull_request.base.sha`). FAIL-CLOSED (007.005-T
#       AC-4): if --base-ref is not supplied and cannot otherwise be
#       resolved, this script FAILS. The only way to skip the comparison
#       is the explicit --allow-no-merge-base flag, which CI never passes.

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
FIXTURE_ROOT="${SCRIPT_DIR}/testdata/gitignore"

mode="check"
base_ref=""
head_ref="HEAD"
target_file=".gitignore"
allow_no_merge_base="false"

usage() {
  cat >&2 <<'EOF'
Usage:
  check-gitignore-append-only.sh --self-test
  check-gitignore-append-only.sh --base-ref <ref> [--head-ref <ref>] [--file <path>] [--allow-no-merge-base]
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --self-test)
      mode="self-test"
      shift
      ;;
    --base-ref)
      base_ref="${2:-}"
      shift 2
      ;;
    --head-ref)
      head_ref="${2:-}"
      shift 2
      ;;
    --file)
      target_file="${2:-}"
      shift 2
      ;;
    --allow-no-merge-base)
      allow_no_merge_base="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "::error::unrecognized argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

# non_comment_lines FILE
#
# Extracts the non-comment, non-blank lines from FILE, in order. A line is
# a comment if its first non-whitespace character is '#'; a line is blank
# if it contains only whitespace.
non_comment_lines() {
  local file="$1"
  grep -Ev '^[[:space:]]*(#|$)' -- "$file"
}

# subsequence_check OLD_FILE NEW_FILE
#
# Returns 0 if OLD_FILE's non-comment/non-blank lines are an ordered
# subsequence of NEW_FILE's; returns 1 otherwise.
#
# INTENTIONAL STUB (007.004-T / U4a): the real subsequence predicate is
# implemented in 007.005-T (U4b). This stub deliberately does not
# implement the check and always reports "not implemented" (exit 2) so
# that every committed fixture demonstrably fails against it (AC-3),
# proving the fixtures assert something real before the logic exists.
subsequence_check() {
  echo "check-gitignore-append-only: subsequence_check not yet implemented (007.005-T / U4b)" >&2
  return 2
}

run_self_test() {
  local overall_status=0
  local fixture_dir fixture_name expected actual_status old_file new_file

  if [ ! -d "$FIXTURE_ROOT" ]; then
    echo "::error::fixture root not found: $FIXTURE_ROOT" >&2
    return 2
  fi

  for fixture_dir in "$FIXTURE_ROOT"/*/; do
    [ -d "$fixture_dir" ] || continue
    fixture_name="$(basename -- "$fixture_dir")"
    old_file="${fixture_dir}old.gitignore"
    new_file="${fixture_dir}new.gitignore"
    expected_file="${fixture_dir}expected"

    if [ ! -f "$old_file" ] || [ ! -f "$new_file" ] || [ ! -f "$expected_file" ]; then
      echo "FAIL  ${fixture_name}: missing fixture file(s)"
      overall_status=1
      continue
    fi

    expected="$(tr -d '[:space:]' < "$expected_file")"

    set +e
    subsequence_check "$old_file" "$new_file"
    actual_status=$?
    set -e

    case "$actual_status" in
      0) actual="pass" ;;
      1) actual="fail" ;;
      *) actual="error(${actual_status})" ;;
    esac

    if [ "$actual" = "$expected" ]; then
      echo "PASS  ${fixture_name}: expected=${expected} actual=${actual}"
    else
      echo "FAIL  ${fixture_name}: expected=${expected} actual=${actual}"
      overall_status=1
    fi
  done

  return "$overall_status"
}

run_check() {
  if [ -z "$base_ref" ]; then
    if [ "$allow_no_merge_base" = "true" ]; then
      echo "::warning::no --base-ref supplied; skipping I6 check via explicit --allow-no-merge-base"
      return 0
    fi
    echo "::error::--base-ref was not supplied and no merge-base could be resolved. FAIL-CLOSED (007.005-T AC-4): pass --base-ref explicitly, or pass --allow-no-merge-base to skip deliberately (CI never does)." >&2
    return 1
  fi

  echo "check-gitignore-append-only: base=${base_ref} head=${head_ref} file=${target_file}"

  local old_tmp new_tmp
  old_tmp="$(mktemp)"
  new_tmp="$(mktemp)"
  trap 'rm -f "$old_tmp" "$new_tmp"' RETURN

  if ! git show "${base_ref}:${target_file}" > "$old_tmp" 2>/dev/null; then
    echo "::error::could not read ${target_file} at ${base_ref}" >&2
    return 1
  fi
  if ! git show "${head_ref}:${target_file}" > "$new_tmp" 2>/dev/null; then
    echo "::error::could not read ${target_file} at ${head_ref}" >&2
    return 1
  fi

  if subsequence_check "$old_tmp" "$new_tmp"; then
    echo "check-gitignore-append-only: PASS (I6 holds: ${base_ref}..${head_ref} for ${target_file})"
    return 0
  else
    echo "::error::check-gitignore-append-only: FAIL (I6 violated: ${base_ref}..${head_ref} for ${target_file} is not append-only)" >&2
    return 1
  fi
}

if [ "$mode" = "self-test" ]; then
  run_self_test
  exit $?
else
  run_check
  exit $?
fi
