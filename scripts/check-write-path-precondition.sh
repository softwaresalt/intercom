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
#   sql.Open bbolt.Open
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
# Usage:
#   scripts/check-write-path-precondition.sh
#     Scans tracked, non-test Go files under internal/** and cmd/**. Exits 0
#     when no write primitive is found.
#   scripts/check-write-path-precondition.sh --self-test
#     Verifies every committed scripts/testdata/writepath/*.go fixture
#     against its expected verdict, then verifies the real tracked tree
#     passes. Exits 0 only when both succeed.

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

run_mode() {
  local mode="$1"
  "$PYTHON_BIN" - "$mode" <<'PY'
from __future__ import annotations

import re
import subprocess
import sys
from pathlib import Path

mode = sys.argv[1]
root = Path.cwd()

SELECTORS = [
    "os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
    "os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
    "os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
]

SELECTOR_RES = [
    (sel, re.compile(r'(?<![\w.])' + re.escape(sel) + r'(?![\w])'))
    for sel in SELECTORS
]

FIXTURE_DIR = root / "scripts" / "testdata" / "writepath"

REGISTER_NAME = "the consolidated risk register (internal/pathsafe package doc, root.go)"
EXCEPTION_NAME = "011.003-T's Constitution Check exception (internal/config/validate.go rule 7)"


def mask_go_non_code(text: str) -> str:
    # Same state machine as scripts/check-retired-architecture.sh's
    # mask_go_non_code: blanks out comments/string/rune literal contents
    # (preserving line structure and total length) so a selector name
    # appearing only in a comment or string is never mistaken for real code.
    code, line_comment, block_comment, string, raw_string, rune = range(6)
    state = code
    out = []
    i = 0
    while i < len(text):
        ch = text[i]
        nxt = text[i + 1] if i + 1 < len(text) else ''

        if state == code:
            if ch == '/' and nxt == '/':
                out.extend('  ')
                i += 2
                state = line_comment
                continue
            if ch == '/' and nxt == '*':
                out.extend('  ')
                i += 2
                state = block_comment
                continue
            if ch == '"':
                out.append(' ')
                i += 1
                state = string
                continue
            if ch == '`':
                out.append(' ')
                i += 1
                state = raw_string
                continue
            if ch == "'":
                out.append(' ')
                i += 1
                state = rune
                continue
            out.append(ch)
            i += 1
            continue

        if state == line_comment:
            if ch == '\n':
                out.append('\n')
                state = code
            else:
                out.append(' ')
            i += 1
            continue

        if state == block_comment:
            if ch == '*' and nxt == '/':
                out.extend('  ')
                i += 2
                state = code
            else:
                out.append('\n' if ch == '\n' else ' ')
                i += 1
            continue

        if state == string:
            if ch == '\\' and nxt:
                out.extend('  ')
                i += 2
                continue
            out.append('\n' if ch == '\n' else ' ')
            i += 1
            if ch == '"':
                state = code
            continue

        if state == raw_string:
            out.append('\n' if ch == '\n' else ' ')
            i += 1
            if ch == '`':
                state = code
            continue

        if state == rune:
            if ch == '\\' and nxt:
                out.extend('  ')
                i += 2
                continue
            out.append('\n' if ch == '\n' else ' ')
            i += 1
            if ch == "'":
                state = code
            continue

    return ''.join(out)


def scan_file(path: Path):
    findings = []
    masked = mask_go_non_code(path.read_text(encoding='utf-8'))
    for line_no, line in enumerate(masked.splitlines(), start=1):
        for sel, pattern in SELECTOR_RES:
            if pattern.search(line):
                findings.append(f"{path.as_posix()}:{line_no}: write primitive {sel!r} found")
    return findings


def should_scan(rel_path: str) -> bool:
    if rel_path.endswith('_test.go') is True:
        return False
    if not rel_path.endswith('.go'):
        return False
    return rel_path.startswith('internal/') or rel_path.startswith('cmd/')


def run_repo_scan():
    proc = subprocess.run(
        ['git', 'ls-files', '--', 'internal/**', 'cmd/**'],
        cwd=root, text=True, capture_output=True, check=True,
    )
    rel_paths = [p for p in proc.stdout.splitlines() if should_scan(p)]

    findings = []
    for rel_path in rel_paths:
        findings.extend(scan_file(root / rel_path))

    if findings:
        print('\n'.join(findings), file=sys.stderr)
        print(
            f"::error::a destructive filesystem write primitive was found under "
            f"internal/** or cmd/**. This trips the write-path precondition "
            f"recorded in {EXCEPTION_NAME} and tracked in {REGISTER_NAME}. "
            f"RETIREMENT PROCEDURE: re-evaluate every finding in the risk "
            f"register against this new write call site, land a mitigation "
            f"(or an explicit, re-justified acceptance) in the SAME change, "
            f"and update both the register and the Constitution Check "
            f"exception to reflect the arrival of a real write path.",
            file=sys.stderr,
        )
        raise SystemExit(1)


def run_fixture_self_test():
    if not FIXTURE_DIR.is_dir():
        raise SystemExit(f"fixture dir not found: {FIXTURE_DIR.as_posix()}")

    failures = []
    discovered = sorted(FIXTURE_DIR.glob('*.go'))
    if not discovered:
        raise SystemExit(f"no fixtures discovered under {FIXTURE_DIR.as_posix()}")

    for path in discovered:
        findings = scan_file(path)
        rejected = bool(findings)
        name = path.name
        if name.startswith('reject-'):
            if rejected:
                print(f"PASS {name}: rejected as expected")
            else:
                failures.append(f"{name}: expected rejection (write primitive), got clean")
        elif name.startswith('accept-'):
            if rejected:
                detail = '; '.join(findings)
                failures.append(f"{name}: expected clean, got findings: {detail}")
            else:
                print(f"PASS {name}: clean as expected")
        else:
            failures.append(f"{name}: fixture filename must start with 'accept-' or 'reject-'")

    if failures:
        print('\n'.join(f"FAIL {f}" for f in failures), file=sys.stderr)
        raise SystemExit(1)


if mode == 'repo':
    run_repo_scan()
elif mode == 'self-test':
    run_fixture_self_test()
else:
    raise SystemExit(f'unknown mode: {mode}')
PY
}

case "${1:-}" in
  "")
    run_mode repo
    ;;
  --self-test)
    run_mode self-test
    run_mode repo
    echo "self-test passed: fixtures matched expectations and the tracked tree is clean"
    ;;
  *)
    echo "usage: scripts/check-write-path-precondition.sh [--self-test]" >&2
    exit 2
    ;;
esac
