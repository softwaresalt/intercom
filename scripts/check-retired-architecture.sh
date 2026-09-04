#!/usr/bin/env bash
set -euo pipefail

# Retired-architecture regression gate for the corrected C1 config surface.
#
# Threat-model limits (D6b): this is an anti-accident control, not an
# anti-adversary control. CI runs workflow files from the PR head, so a PR
# can edit the workflow or this script. The CI workflow is also
# autoharness-generated and may overwrite the wiring step on a future render.
#
# Usage:
#   scripts/check-retired-architecture.sh
#     Scans tracked files in internal/config/** (excluding *_test.go and any
#     testdata/ directory), config.toml.example, and cmd/**. Exits 0 when no
#     retired-architecture token is found as a Go identifier or TOML key.
#   scripts/check-retired-architecture.sh --self-test
#     Verifies the committed fixture in scripts/testdata/ would be rejected,
#     then verifies the real tracked tree passes. Exits 0 only when both
#     checks succeed.

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

scan_with_mode() {
  local mode="$1"
  "$PYTHON_BIN" - "$mode" <<'PY'
import re
import subprocess
import sys
from pathlib import Path

mode = sys.argv[1]
root = Path.cwd()

forbidden_parts = {
    "slack": ["slack"],
    "socketmode": ["socket", "mode"],
    "channel_id": ["channel", "id"],
    "team_id": ["team", "id"],
    "acp": ["acp"],
    "host_cli": ["host", "cli"],
    "ipc_name": ["ipc", "name"],
}

go_identifier_re = re.compile(r'\b[A-Za-z_][A-Za-z0-9_]*\b')
bare_key_re = re.compile(r'[A-Za-z_][A-Za-z0-9_]*')
component_re = re.compile(r'[A-Z]+(?=[A-Z][a-z]|$)|[A-Z]?[a-z]+|[0-9]+')

def should_scan_repo_path(path: str) -> bool:
    if path == 'config.toml.example':
        return True
    if path.startswith('cmd/'):
        return path.endswith('.go')
    if not path.startswith('internal/config/'):
        return False
    if '/testdata/' in path or path.endswith('_test.go'):
        return False
    return path.endswith('.go')

def split_identifier(name: str):
    parts = []
    for chunk in name.split('_'):
        parts.extend(m.group(0).lower() for m in component_re.finditer(chunk))
    return parts or [name.lower()]

def matches_forbidden_parts(parts):
    for token, want in forbidden_parts.items():
        if len(parts) < len(want):
            continue
        for i in range(len(parts) - len(want) + 1):
            if parts[i:i+len(want)] == want:
                return token
    return None

def mask_go_non_code(text: str) -> str:
    code, line_comment, block_comment, string, raw_string, rune = range(6)
    state = code
    out = []
    i = 0
    while i < len(text):
        ch = text[i]
        nxt = text[i+1] if i + 1 < len(text) else ''

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

def strip_toml_comment(line: str) -> str:
    out = []
    in_basic = False
    in_literal = False
    escaped = False
    for ch in line:
        if in_basic:
            out.append(ch)
            if escaped:
                escaped = False
            elif ch == '\\':
                escaped = True
            elif ch == '"':
                in_basic = False
            continue
        if in_literal:
            out.append(ch)
            if ch == "'":
                in_literal = False
            continue
        if ch == '#':
            break
        out.append(ch)
        if ch == '"':
            in_basic = True
        elif ch == "'":
            in_literal = True
    return ''.join(out)

def scan_go(path: Path):
    findings = []
    masked = mask_go_non_code(path.read_text(encoding='utf-8'))
    for line_no, line in enumerate(masked.splitlines(), start=1):
        for match in go_identifier_re.finditer(line):
            token = matches_forbidden_parts(split_identifier(match.group(0)))
            if token:
                findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in Go identifier {match.group(0)!r}")
    return findings

def scan_toml(path: Path):
    findings = []
    current_table = []
    for line_no, raw_line in enumerate(path.read_text(encoding='utf-8').splitlines(), start=1):
        line = strip_toml_comment(raw_line).strip()
        if not line:
            continue
        if line.startswith('[') and line.endswith(']'):
            header = line.strip('[]').strip()
            current_table = bare_key_re.findall(header)
            for key in current_table:
                token = matches_forbidden_parts(key.lower().split('_'))
                if token:
                    findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML table key {key!r}")
            continue
        if '=' not in line:
            continue
        lhs = line.split('=', 1)[0]
        keys = current_table + bare_key_re.findall(lhs)
        if not keys:
            continue
        # For TOML, evaluate each path segment as snake_case words so
        # host_cli_args still exposes the retired host_cli token.
        for key in keys:
            token = matches_forbidden_parts(key.lower().split('_'))
            if token:
                findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML key {key!r}")
    return findings

if mode == 'repo':
    proc = subprocess.run(
        ['git', 'ls-files', '--', 'config.toml.example', 'cmd/**', 'internal/config/**'],
        cwd=root,
        text=True,
        capture_output=True,
        check=True,
    )
    rel_paths = [p for p in proc.stdout.splitlines() if should_scan_repo_path(p)]
elif mode == 'fixture':
    rel_paths = ['scripts/testdata/retired-architecture-fixture.toml']
else:
    raise SystemExit(f'unknown mode: {mode}')

findings = []
for rel_path in rel_paths:
    path = root / rel_path
    if rel_path.endswith('.go'):
        findings.extend(scan_go(path))
    elif rel_path.endswith('.toml'):
        findings.extend(scan_toml(path))

if findings:
    print('\n'.join(findings), file=sys.stderr)
    raise SystemExit(1)
PY
}

case "${1:-}" in
  "")
    scan_with_mode repo
    ;;
  --self-test)
    if scan_with_mode fixture; then
      echo "self-test failed: fixture was not rejected" >&2
      exit 1
    fi
    scan_with_mode repo
    echo "self-test passed: fixture was rejected and the tracked tree is clean"
    ;;
  *)
    echo "usage: scripts/check-retired-architecture.sh [--self-test]" >&2
    exit 2
    ;;
esac
