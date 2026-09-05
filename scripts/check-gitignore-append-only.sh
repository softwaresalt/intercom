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
# STATUS: 007.004-T (U4a) shipped the CLI surface, fixture harness, and
# --self-test driver first, with the core subsequence predicate as an
# intentional stub that every fixture failed against (test-first,
# Principle II). 007.005-T (U4b) replaces the stub with the real
# `subsequence_check` implementation below; the CLI surface and
# git-ref comparison mode (`run_check`) are unchanged from U4a.
#
# REMEDIATION (adversarial re-review, 2026-09-04): two gaps fixed prior to
# merge, both verified: (1) `non_comment_lines` now strips trailing `\r`
# before classification, so a CRLF-vs-LF-only difference between base and
# head can never be misread as a deletion/reorder (verified: a
# content-identical CRLF/LF pair now returns pass, previously a bare-CR
# artifact could have been miscompared as distinct text); (2) `run_check`
# now distinguishes "target_file did not exist at base_ref at all" (a
# genuinely first-ever add -- e.g. the very first `.gitignore` a repo ever
# commits) from "base_ref itself does not resolve" -- the former is
# treated as an empty OLD (trivially compliant, per subsequence_check's
# own empty-OLD semantics) rather than failing closed; the latter (an
# actually-unresolvable ref) still fails closed exactly as AC-4 requires.
# Verified with a synthetic empty-tree commit as --base-ref.
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
# Extracts the non-comment, non-blank lines from FILE, in order, via
# stdout; returns 0 on success (including a legitimately empty result --
# e.g. an all-comment/blank file) or 2 on a genuine read/filter failure. A
# line is a comment if its first non-whitespace character is '#'; a line
# is blank if it contains only whitespace. Trailing carriage returns (CRLF
# line endings) are stripped before classification/comparison so a CRLF vs
# LF difference alone (no content change) can never be misread as a
# deletion/reorder.
#
# HARDENING (Copilot review, 2026-09-05): `grep -Ev` exits 1 when it finds
# zero matching lines -- a LEGITIMATE, expected outcome for an all-comment
# or blank file (or the first-ever-file carve-out's empty OLD), not an
# error. `sed` and `grep` are each run as a SEPARATE command-substitution
# assignment (not one piped one-liner) specifically so each stage's own
# exit status can be checked independently and unambiguously: a piped
# `sed | grep` inside a single `$(...)` runs in a subshell whose internal
# `PIPESTATUS` does not propagate to the outer shell, and `pipefail` alone
# can report the rightmost (grep's) exit code even when sed is the one
# that actually failed. Each assignment is the condition of an `if`, which
# is exempt from `set -e` propagation, so the exit status is captured
# explicitly without ever toggling the global `errexit` setting (toggling
# it with a bare `set +e`/`set -e` pair would be unsafe here: it is GLOBAL
# shell state, not scoped to this function, so it could clobber whatever
# errexit state an outer caller was itself relying on). grep's exit 1 is
# treated as "valid empty result"; any OTHER nonzero exit from either
# command is a genuine failure and is reported as such rather than
# silently treated as "no lines".
non_comment_lines() {
  local file="$1"
  local raw_content filtered_content sed_status grep_status

  if raw_content="$(sed 's/\r$//' -- "$file")"; then
    sed_status=0
  else
    sed_status=$?
  fi
  if [ "$sed_status" -ne 0 ]; then
    echo "::error::sed failed reading ${file} (exit ${sed_status})" >&2
    return 2
  fi

  if filtered_content="$(printf '%s\n' "$raw_content" | grep -Ev '^[[:space:]]*(#|$)')"; then
    grep_status=0
  else
    grep_status=$?
  fi
  if [ "$grep_status" -ne 0 ] && [ "$grep_status" -ne 1 ]; then
    echo "::error::grep failed filtering ${file} (exit ${grep_status})" >&2
    return 2
  fi

  if [ -n "$filtered_content" ]; then
    printf '%s\n' "$filtered_content"
  fi
  return 0
}

# subsequence_check OLD_FILE NEW_FILE
#
# Returns 0 if OLD_FILE's non-comment/non-blank lines are an ordered
# subsequence of NEW_FILE's; returns 1 if not; returns 2 on a genuine
# extraction failure from either file (distinct from "not a subsequence").
# Implemented (007.005-T / U4b) as a standard two-pointer subsequence
# walk: advance the OLD pointer only on a match while scanning NEW once;
# OLD is a subsequence of NEW iff every OLD line was matched by the time
# NEW is exhausted. An empty OLD (no non-comment/non-blank lines) is
# trivially a subsequence of anything.
#
# HARDENING (Copilot review, 2026-09-05): the two `non_comment_lines`
# calls are now MATERIALIZED via direct command-substitution assignment
# (`var="$(...)"`), not process substitution consumed by `mapfile`. A
# failure inside a process substitution (`< <(...)`) does not reliably
# propagate to trigger `set -e` in the consuming command, since `mapfile`
# itself still "succeeds" regardless of what the substituted command's
# exit status was -- which could silently produce a truncated/empty array
# and a false PASS. Each assignment is the condition of an `if` (same
# rationale as `non_comment_lines` above: this avoids ever toggling the
# global `errexit` setting) so its exit status is checked explicitly
# before any comparison proceeds.
subsequence_check() {
  local old_file="$1" new_file="$2"
  local old_content new_content old_status new_status
  local -a old_lines new_lines

  if old_content="$(non_comment_lines "$old_file")"; then
    old_status=0
  else
    old_status=$?
  fi
  if [ "$old_status" -ne 0 ]; then
    echo "::error::could not extract non-comment lines from ${old_file}" >&2
    return 2
  fi

  if new_content="$(non_comment_lines "$new_file")"; then
    new_status=0
  else
    new_status=$?
  fi
  if [ "$new_status" -ne 0 ]; then
    echo "::error::could not extract non-comment lines from ${new_file}" >&2
    return 2
  fi

  if [ -n "$old_content" ]; then
    mapfile -t old_lines <<< "$old_content"
  else
    old_lines=()
  fi
  if [ -n "$new_content" ]; then
    mapfile -t new_lines <<< "$new_content"
  else
    new_lines=()
  fi

  local old_len=${#old_lines[@]} new_len=${#new_lines[@]}
  local i=0 j=0

  while [ "$i" -lt "$old_len" ] && [ "$j" -lt "$new_len" ]; do
    if [ "${old_lines[$i]}" = "${new_lines[$j]}" ]; then
      i=$((i + 1))
    fi
    j=$((j + 1))
  done

  [ "$i" -eq "$old_len" ]
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

  if ! git rev-parse --verify --quiet "${base_ref}^{commit}" >/dev/null; then
    echo "::error::--base-ref '${base_ref}' does not resolve to a valid commit" >&2
    return 1
  fi
  if ! git rev-parse --verify --quiet "${head_ref}^{commit}" >/dev/null; then
    echo "::error::--head-ref '${head_ref}' does not resolve to a valid commit" >&2
    return 1
  fi

  local old_tmp new_tmp
  old_tmp="$(mktemp)"
  new_tmp="$(mktemp)"
  trap 'rm -f "$old_tmp" "$new_tmp"' RETURN

  # First-ever-file carve-out: if target_file did not exist at base_ref at
  # all (a valid commit, just no version of this path), there is nothing
  # prior to have deleted or reordered -- treat it as an empty OLD, which
  # subsequence_check already handles as trivially compliant (any NEW
  # content is vacuously a superset of nothing). This is NOT the fail-open
  # "no --base-ref supplied" case above (AC-4): base_ref itself is valid
  # and explicit; only the path is absent there.
  #
  # POSITIVELY confirms absence via `git ls-tree` rather than inferring it
  # from any nonzero `git cat-file -e` exit (re-review finding, 2026-09-04):
  # a bare nonzero exit cannot distinguish "path genuinely absent" from
  # other failure modes (e.g. a corrupted/unreadable tree or blob object),
  # so a real error could otherwise be silently mistaken for "first-ever
  # add" and pass. `git ls-tree --name-only <ref> -- <path>` exits non-zero
  # only when the TREE lookup itself fails (already-validated base_ref, so
  # this should not happen); a healthy lookup that simply finds no matching
  # path exits 0 with empty output -- that combination is the only case
  # treated as "absent".
  local base_ls_tree_output
  if ! base_ls_tree_output="$(git ls-tree --name-only "${base_ref}" -- "${target_file}" 2>&1)"; then
    echo "::error::git ls-tree failed while checking whether ${target_file} exists at ${base_ref}: ${base_ls_tree_output}" >&2
    return 1
  fi

  if [ -z "$base_ls_tree_output" ]; then
    : > "$old_tmp"
    echo "::notice::${target_file} did not exist at ${base_ref} (git ls-tree confirmed absent); treating as a first-ever add (trivially append-only)"
  elif ! git show "${base_ref}:${target_file}" > "$old_tmp" 2>/dev/null; then
    echo "::error::${target_file} is listed in ${base_ref}'s tree but could not be read (unexpected)" >&2
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
