#!/usr/bin/env bash
set -euo pipefail

# check-unignore-regression.sh
#
# Un-ignore regression gate (011.012-T, resolves 11ECB954). Detects when a
# path that was actually ignored moves to un-ignored across a change,
# using git's OWN ignore matcher (never a gitignore engine reimplemented in
# bash/Python) via `git check-ignore`.
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
#   Part 2 (secondary): any path CURRENTLY untracked-and-ignored in the
#     working tree must still have been ignored at base-ref, evaluated via
#     a detached worktree at base-ref (`git worktree add --detach`), since
#     git check-ignore reads the working tree, not a ref. KNOWN BOUNDARY:
#     this part is INERT on a fresh CI checkout (no untracked cruft exists
#     to enumerate) -- part 1 carries the real coverage in CI. This
#     boundary is deliberate, not a bug, and is recorded here rather than
#     silently assumed away.
#
# Per-path evaluation is ALWAYS literal and individual (never a single
# batched `git check-ignore` invocation): `git check-ignore` exits 0 if ANY
# argument is ignored, so a batched call would mask an entry that silently
# became un-ignored. Glob-shaped denylist entries are passed as literal
# arguments (never shell-expanded).
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

mode="check"
base_ref=""
head_ref="HEAD"

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
    -h|--help)
      echo "Usage: $0 --self-test | --base-ref <ref> [--head-ref <ref>]" >&2
      exit 0
      ;;
    *)
      echo "::error::unrecognized argument: $1" >&2
      exit 2
      ;;
  esac
done

run_py() {
  "$PYTHON_BIN" - "$@" <<'PY'
from __future__ import annotations

import subprocess
import sys
import tempfile
from pathlib import Path

# Revision 5 denylist: ONLY paths guaranteed ignored by HEAD + 011.001-T.
# Never reference a pattern introduced by droppable content (standing rule).
DENYLIST = [
    ".env",
    ".env.local",
    "config.toml",
    ".claude/instructions.md",
    ".github/copilot/settings.local.json",
    ".autoharness/gates/pipeline-topology-force-audit.log",
    ".backlogit/hooks_queue.jsonl",
]


def git(args, cwd=None, check=True):
    return subprocess.run(
        ["git"] + args, cwd=cwd, text=True, capture_output=True, check=check,
    )


def root_gitignore_text_at(repo_dir: Path, ref: str) -> str:
    # HEAD is read directly from the checked-out working tree (the ref IS
    # what's on disk); any other ref is read via `git show` against the
    # repo's own history, never a worktree checkout (avoids the nested-
    # .gitignore-discovery / real-filesystem-noise hazard below).
    if ref == "HEAD":
        gi = repo_dir / ".gitignore"
        return gi.read_text(encoding="utf-8") if gi.exists() else ""
    proc = subprocess.run(
        ["git", "show", f"{ref}:.gitignore"], cwd=str(repo_dir), capture_output=True, text=True,
    )
    if proc.returncode != 0:
        # Absent at that ref is a legitimate (if unlikely) state -- treated
        # as an empty ignore file, not an error.
        return ""
    return proc.stdout


def make_scratch_gitignore(scratch_root: Path, name: str, content: str) -> Path:
    # Isolates evaluation to ONLY the root .gitignore's content (the
    # checker's documented boundary: "inspects only the root .gitignore").
    # A bare `git check-ignore --no-index` run against the REAL repo
    # working tree or a real worktree checkout would instead walk the
    # ENTIRE real directory tree looking for nested .gitignore files --
    # including untracked, machine-local ones (e.g. installed plugin
    # content under .copilot/) that have nothing to do with what this
    # shipment's changes actually touch, producing false regressions keyed
    # off unrelated local machine state rather than the actual PR diff.
    # Writing the extracted content into an otherwise-empty scratch
    # directory and running --no-index there eliminates that noise
    # entirely and makes the check deterministic and reproducible.
    scratch_dir = scratch_root / name
    scratch_dir.mkdir(parents=True, exist_ok=True)
    # git check-ignore (even with --no-index) still requires SOME git
    # repository context to run at all; a bare directory with only a
    # .gitignore file and no .git raises "not a git repository". This
    # scratch repo is never committed to or used for any real git
    # operation beyond ignore-pattern matching.
    git(["init", "-q"], cwd=str(scratch_dir))
    (scratch_dir / ".gitignore").write_text(content, encoding="utf-8")
    return scratch_dir


def is_ignored(cwd: Path, rel_path: str) -> bool:
    # --no-index: works without the path needing to exist on disk. Exit 0
    # = ignored, exit 1 = not ignored, anything else = a real error (never
    # silently treated as either verdict).
    proc = subprocess.run(
        ["git", "check-ignore", "--no-index", "-q", "--", rel_path],
        cwd=str(cwd), capture_output=True, text=True,
    )
    if proc.returncode not in (0, 1):
        raise SystemExit(
            f"::error::git check-ignore errored for {rel_path!r} in {cwd}: "
            f"exit {proc.returncode}: {proc.stderr.strip()}"
        )
    return proc.returncode == 0


def is_ignored_batch(cwd: Path, rel_paths: list[str]) -> list[bool]:
    # Batched form of is_ignored, used for the differential check's
    # potentially large candidate universe (Part 2): a per-path subprocess
    # spawn does not scale to thousands of untracked files. Uses
    # `--verbose --non-matching --stdin` (per 011.012-T's own AC) rather
    # than passing all paths as positional arguments, since a bare batched
    # `git check-ignore path1 path2 ...` exits 0 if ANY argument matches --
    # exactly the masking hazard this checker exists to avoid. git
    # preserves one output line per input line, in order, under
    # --non-matching, so results are aligned positionally rather than by
    # re-parsing the (potentially C-quoted) pathname field back out of the
    # output.
    if not rel_paths:
        return []
    stdin_payload = ("\n".join(rel_paths) + "\n").encode("utf-8")
    proc = subprocess.run(
        ["git", "check-ignore", "--no-index", "--verbose", "--non-matching", "--stdin"],
        cwd=str(cwd), input=stdin_payload, capture_output=True,
    )
    if proc.returncode not in (0, 1):
        raise SystemExit(
            f"::error::git check-ignore --stdin errored in {cwd}: "
            f"exit {proc.returncode}: {proc.stderr.decode('utf-8', 'replace').strip()}"
        )
    stdout_text = proc.stdout.decode("utf-8", "replace")
    lines = stdout_text.split("\n")
    if lines and lines[-1] == "":
        lines = lines[:-1]
    if len(lines) != len(rel_paths):
        raise SystemExit(
            f"::error::git check-ignore --stdin returned {len(lines)} lines for "
            f"{len(rel_paths)} input paths in {cwd}; cannot safely align results"
        )
    return [parse_check_ignore_verbose_line(line) for line in lines]


def parse_check_ignore_verbose_line(line: str) -> bool:
    # `--verbose` reports the LAST matching pattern regardless of whether
    # it is a positive pattern or a negation (`!pattern`) -- a naive "any
    # match at all" reading would misclassify a negation match (which
    # RE-INCLUDES the path, i.e. NOT ignored) as ignored. Format is
    # "source:linenum:pattern<TAB>pathname", or exactly "::<TAB>pathname"
    # for a non-matching path (--non-matching).
    prefix, _, _ = line.partition("\t")
    if prefix == "::":
        return False
    parts = prefix.split(":", 2)
    pattern = parts[2] if len(parts) == 3 else ""
    return not pattern.startswith("!")


def run_denylist_check(repo_dir: Path, scratch_root: Path, ref: str = "HEAD"):
    """Part 1. Returns (evaluated_count, failures)."""
    content = root_gitignore_text_at(repo_dir, ref)
    scratch_dir = make_scratch_gitignore(scratch_root, f"denylist-{ref.replace('/', '_')}", content)
    failures = []
    for path in DENYLIST:
        if not is_ignored(scratch_dir, path):
            failures.append(path)
    return len(DENYLIST), failures


def run_differential_check(repo_dir: Path, scratch_root: Path, base_ref: str, head_ref: str = "HEAD"):
    """Part 2. Returns (evaluated_count, failures).

    Candidate universe: ALL untracked paths REALLY PRESENT in the working
    tree, ignored or not (git ls-files --others WITHOUT --exclude-standard)
    -- deliberately broader than "--ignored", because the whole point of a
    regression is a path that is untracked and NO LONGER ignored at head;
    such a path would never appear under an --ignored-filtered enumeration
    at head. This enumeration is the only place real on-disk untracked
    content is consulted; the actual ignored-status EVALUATION for each
    candidate is always against the isolated root-.gitignore-only scratch
    directories built above, never the real tree (see make_scratch_gitignore
    for why).
    """
    proc = git(["ls-files", "--others", "-z"], cwd=str(repo_dir))
    # -z (NUL-terminated, unquoted) avoids git's C-style quoting of
    # unusual/non-ASCII filenames that a newline-delimited
    # `git ls-files --others` (no -z) would otherwise apply -- a quoted
    # representation fed back into check-ignore --stdin would not match
    # the real on-disk path, silently mis-evaluating that candidate's
    # ignored/not-ignored verdict.
    all_untracked = [p for p in proc.stdout.split("\0") if p]

    if not all_untracked:
        return 0, []

    base_content = root_gitignore_text_at(repo_dir, base_ref)
    head_content = root_gitignore_text_at(repo_dir, head_ref)
    base_scratch = make_scratch_gitignore(scratch_root, f"diff-base-{base_ref.replace('/', '_')}", base_content)
    head_scratch = make_scratch_gitignore(scratch_root, f"diff-head-{head_ref.replace('/', '_')}", head_content)

    base_ignored = is_ignored_batch(base_scratch, all_untracked)
    candidates = [p for p, ignored in zip(all_untracked, base_ignored) if ignored]
    evaluated = len(candidates)
    if not candidates:
        return 0, []
    head_ignored = is_ignored_batch(head_scratch, candidates)
    failures = [p for p, ignored in zip(candidates, head_ignored) if not ignored]
    return evaluated, failures


def run_ls_files_check(repo_dir: Path):
    proc = git(["ls-files", "-i", "-c", "--exclude-standard"], cwd=str(repo_dir))
    tracked_ignored = [line for line in proc.stdout.splitlines() if line]
    return tracked_ignored


def do_self_test_landing_precondition(scratch_root: Path):
    repo_dir = Path.cwd()
    evaluated, failures = run_denylist_check(repo_dir, scratch_root, ref="HEAD")
    if failures:
        print(
            "FAIL landing-precondition: the following denylist entries are NOT "
            f"ignored at HEAD: {failures}",
            file=sys.stderr,
        )
        return False
    print(f"PASS landing-precondition: all {evaluated} denylist entries are ignored at HEAD")
    return True


def make_scenario_repo(tmp_root: Path, name: str, old_gitignore: str, new_gitignore: str, existing_paths):
    repo = tmp_root / name
    repo.mkdir(parents=True)
    git(["init", "-q", "-b", "main"], cwd=str(repo))
    git(["config", "user.email", "fixture@example.invalid"], cwd=str(repo))
    git(["config", "user.name", "fixture"], cwd=str(repo))

    (repo / ".gitignore").write_text(old_gitignore, encoding="utf-8")
    git(["add", ".gitignore"], cwd=str(repo))
    git(["commit", "-q", "-m", "base"], cwd=str(repo))
    base_ref = git(["rev-parse", "HEAD"], cwd=str(repo)).stdout.strip()

    (repo / ".gitignore").write_text(new_gitignore, encoding="utf-8")
    for rel in existing_paths:
        p = repo / rel
        p.parent.mkdir(parents=True, exist_ok=True)
        p.write_text("fixture content\n", encoding="utf-8")
    git(["add", ".gitignore"], cwd=str(repo))
    git(["commit", "-q", "-m", "head"], cwd=str(repo))

    return repo, base_ref


def run_scenario(tmp_root: Path, name: str, old_gitignore: str, new_gitignore: str, existing_paths, expect_reject: bool):
    repo, base_ref = make_scenario_repo(tmp_root, name, old_gitignore, new_gitignore, existing_paths)
    scratch_root = tmp_root / f"{name}-scratch"
    evaluated, failures = run_differential_check(repo, scratch_root, base_ref)
    rejected = bool(failures)
    if expect_reject:
        if rejected:
            print(f"PASS {name}: rejected as expected (regressed paths: {failures})")
            return True
        print(f"FAIL {name}: expected rejection, got clean (evaluated={evaluated})", file=sys.stderr)
        return False
    else:
        if rejected:
            print(f"FAIL {name}: expected clean, got rejection: {failures}", file=sys.stderr)
            return False
        print(f"PASS {name}: clean as expected (evaluated={evaluated})")
        return True


def do_self_test_scenarios():
    ok = True
    with tempfile.TemporaryDirectory() as tmp:
        tmp_root = Path(tmp)

        # Scenario A: an EXISTING, previously-ignored file is un-ignored by
        # a negation added at head. Must be REJECTED.
        ok &= run_scenario(
            tmp_root,
            "reject-existing-file-unignored",
            old_gitignore="foo/bar.secret\n",
            new_gitignore="foo/bar.secret\n!foo/bar.secret\n",
            existing_paths=["foo/bar.secret"],
            expect_reject=True,
        )

        # Scenario B: a negation for a path that does NOT exist (the
        # stowaway's actual case) is added at head. Must be ACCEPTED --
        # part 2 only evaluates currently-existing untracked-ignored
        # candidates, and a non-existent path is never such a candidate.
        ok &= run_scenario(
            tmp_root,
            "accept-nonexistent-negation",
            old_gitignore="foo/\n",
            new_gitignore="foo/\n!foo/keep.txt\n",
            existing_paths=[],
            expect_reject=False,
        )

    return ok


mode_arg = sys.argv[1] if len(sys.argv) > 1 else ""

if mode_arg == "self-test":
    overall_ok = True
    with tempfile.TemporaryDirectory() as scratch_tmp:
        scratch_root = Path(scratch_tmp)
        overall_ok &= do_self_test_landing_precondition(scratch_root)
        overall_ok &= do_self_test_scenarios()

        # Also run the real two-part denylist check against this actual
        # repo (root .gitignore content only -- never real nested
        # .gitignore/plugin noise, see make_scratch_gitignore).
        repo_dir = Path.cwd()
        denylist_evaluated, denylist_failures = run_denylist_check(repo_dir, scratch_root, ref="HEAD")
        if denylist_evaluated == 0:
            print("FAIL non-vacuity: denylist evaluated count is zero", file=sys.stderr)
            overall_ok = False
        if denylist_failures:
            print(f"FAIL real-repo denylist check: {denylist_failures}", file=sys.stderr)
            overall_ok = False
        else:
            print(f"PASS real-repo denylist check: {denylist_evaluated} entries all ignored")

        tracked_ignored = run_ls_files_check(repo_dir)
        if tracked_ignored:
            print(f"FAIL: tracked files are ignored (git ls-files -i -c): {tracked_ignored}", file=sys.stderr)
            overall_ok = False
        else:
            print("PASS: git ls-files -i -c --exclude-standard is empty")

    if not overall_ok:
        sys.exit(1)
    print("check-unignore-regression --self-test: PASS")
elif mode_arg == "check":
    base_ref = sys.argv[2]
    head_ref = sys.argv[3]
    repo_dir = Path.cwd()

    if head_ref != "HEAD":
        current = git(["rev-parse", "HEAD"]).stdout.strip()
        target = git(["rev-parse", head_ref]).stdout.strip()
        if current != target:
            raise SystemExit(
                f"::error::--head-ref {head_ref!r} does not match the current "
                f"checkout HEAD ({current}); this checker evaluates the CURRENT "
                f"working tree as head-ref, matching how CI checks out the PR head."
            )

    with tempfile.TemporaryDirectory() as scratch_tmp:
        scratch_root = Path(scratch_tmp)
        denylist_evaluated, denylist_failures = run_denylist_check(repo_dir, scratch_root, ref="HEAD")
        diff_evaluated, diff_failures = run_differential_check(repo_dir, scratch_root, base_ref, head_ref="HEAD")

    print(f"check-unignore-regression: denylist_evaluated={denylist_evaluated} differential_evaluated={diff_evaluated}")

    ok = True
    if denylist_evaluated == 0:
        print("::error::non-vacuity violation: denylist evaluated count is zero", file=sys.stderr)
        ok = False
    if denylist_failures:
        print(f"::error::denylist regression -- no longer ignored at head-ref: {denylist_failures}", file=sys.stderr)
        ok = False
    if diff_failures:
        print(f"::error::differential regression -- ignored at base-ref but not at head-ref: {diff_failures}", file=sys.stderr)
        ok = False

    tracked_ignored = run_ls_files_check(repo_dir)
    if tracked_ignored:
        print(f"::error::tracked files report as ignored (git ls-files -i -c --exclude-standard): {tracked_ignored}", file=sys.stderr)
        ok = False

    if not ok:
        print(
            "::error::un-ignore regression detected. No waiver mechanism exists "
            "(deliberate, YAGNI) -- escalate to the operator for explicit review "
            "rather than bypassing this gate.",
            file=sys.stderr,
        )
        sys.exit(1)
    print("check-unignore-regression: PASS")
else:
    raise SystemExit(f"unknown mode: {mode_arg!r}")
PY
}

case "$mode" in
  self-test)
    run_py self-test
    ;;
  check)
    if [ -z "$base_ref" ]; then
      echo "::error::--base-ref was not supplied. FAIL-CLOSED: pass --base-ref explicitly (CI passes github.event.pull_request.base.sha)." >&2
      exit 1
    fi
    run_py check "$base_ref" "$head_ref"
    ;;
esac
