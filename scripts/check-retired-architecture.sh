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
#     Verifies every committed scripts/testdata/retired-*.toml fixture against
#     scripts/testdata/retired-manifest.json, then verifies the real tracked
#     tree passes. Exits 0 only when both checks succeed.

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
import json
import re
import subprocess
import sys
from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover - defensive fallback only
    tomllib = None

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

fixture_glob = "retired-*.toml"
fixture_manifest_path = root / "scripts" / "testdata" / "retired-manifest.json"

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
            if parts[i:i + len(want)] == want:
                return token
    return None


def mask_go_non_code(text: str) -> str:
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


# Defensive fallback when tomllib is unavailable. The primary path uses the
# real TOML parser because parse-based key walking avoids this whole class of
# masking bugs.
def strip_toml_comment(line: str, state: dict[str, bool]):
    out = []
    i = 0
    in_basic = state['in_basic']
    in_literal = state['in_literal']
    in_multiline_basic = state['in_multiline_basic']
    in_multiline_literal = state['in_multiline_literal']

    while i < len(line):
        if in_multiline_basic:
            if line.startswith('"""', i):
                backslashes = 0
                j = i - 1
                while j >= 0 and line[j] == '\\':
                    backslashes += 1
                    j -= 1
                if backslashes % 2 == 0:
                    out.extend('"""')
                    i += 3
                    in_multiline_basic = False
                    continue
            out.append(line[i])
            i += 1
            continue

        if in_multiline_literal:
            if line.startswith("'''", i):
                out.extend("'''")
                i += 3
                in_multiline_literal = False
                continue
            out.append(line[i])
            i += 1
            continue

        ch = line[i]
        nxt3 = line[i:i + 3]

        if in_basic:
            out.append(ch)
            if ch == '\\' and i + 1 < len(line):
                out.append(line[i + 1])
                i += 2
                continue
            if ch == '"':
                in_basic = False
            i += 1
            continue

        if in_literal:
            out.append(ch)
            if ch == "'":
                in_literal = False
            i += 1
            continue

        if nxt3 == '"""':
            out.extend('"""')
            i += 3
            in_multiline_basic = True
            continue
        if nxt3 == "'''":
            out.extend("'''")
            i += 3
            in_multiline_literal = True
            continue
        if ch == '#':
            break
        out.append(ch)
        if ch == '"':
            in_basic = True
        elif ch == "'":
            in_literal = True
        i += 1

    state['in_basic'] = in_basic
    state['in_literal'] = in_literal
    state['in_multiline_basic'] = in_multiline_basic
    state['in_multiline_literal'] = in_multiline_literal
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


def report_toml_key(path: Path, key_path: list[str], segment: str, token: str) -> str:
    return (
        f"{path.as_posix()}: retired token {token!r} in TOML key path "
        f"{'.'.join(key_path)!r} (segment {segment!r})"
    )


def walk_toml_value(path: Path, prefix: list[str], value, findings: list[str]):
    if isinstance(value, dict):
        for key, child in value.items():
            key_path = prefix + [key]
            token = matches_forbidden_parts(key.lower().split('_'))
            if token:
                findings.append(report_toml_key(path, key_path, key, token))
            walk_toml_value(path, key_path, child, findings)
        return

    if isinstance(value, list):
        for child in value:
            walk_toml_value(path, prefix, child, findings)


def scan_toml_with_tomllib(path: Path):
    try:
        parsed = tomllib.loads(path.read_text(encoding='utf-8'))
    except Exception as exc:
        return [f"{path.as_posix()}: TOML parse error (fail-closed): {exc}"]

    findings = []
    walk_toml_value(path, [], parsed, findings)
    return findings


def scan_toml_with_fallback(path: Path):
    # Used as the real scan engine only when tomllib is unavailable (pre-3.11
    # interpreters), but --self-test exercises this function directly against
    # every fixture on every run regardless of interpreter version, so it is
    # never untested dead code even though tomllib is expected to be present
    # in every currently supported environment.
    findings = []
    current_table = []
    state = {
        'in_basic': False,
        'in_literal': False,
        'in_multiline_basic': False,
        'in_multiline_literal': False,
    }
    for line_no, raw_line in enumerate(path.read_text(encoding='utf-8').splitlines(), start=1):
        line = strip_toml_comment(raw_line, state).strip()
        if not line:
            continue
        if not (state['in_multiline_basic'] or state['in_multiline_literal']) and line.startswith('[') and line.endswith(']'):
            header = line.strip('[]').strip()
            current_table = bare_key_re.findall(header)
            for key in current_table:
                token = matches_forbidden_parts(key.lower().split('_'))
                if token:
                    findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML table key {key!r}")
            continue
        if state['in_multiline_basic'] or state['in_multiline_literal'] or '=' not in line:
            continue
        lhs = line.split('=', 1)[0]
        keys = current_table + bare_key_re.findall(lhs)
        for key in keys:
            token = matches_forbidden_parts(key.lower().split('_'))
            if token:
                findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML key {key!r}")

    # Fail closed (AC-6): an unterminated string at EOF (single-line basic/
    # literal OR multi-line basic/literal) means this lexer's simplified
    # state tracking cannot vouch for the rest of the file. Reporting nothing
    # here would be exactly the silent fail-open masking bug (R2) this
    # shipment exists to close, just relocated into the defensive fallback
    # instead of the primary tomllib path. in_basic/in_literal are included
    # (not just the multiline flags) because this simplified per-line lexer
    # tolerates an unterminated single-line string spanning multiple physical
    # lines, and content on those "swallowed" lines must not be silently lost.
    if state['in_multiline_basic'] or state['in_multiline_literal'] or state['in_basic'] or state['in_literal']:
        findings.append(f"{path.as_posix()}: unterminated string at EOF (fail-closed)")

    return findings


def scan_toml(path: Path):
    if tomllib is not None:
        return scan_toml_with_tomllib(path)
    return scan_toml_with_fallback(path)


def scan_path(path: Path):
    if path.suffix == '.go':
        return scan_go(path)
    if path.suffix == '.toml':
        return scan_toml(path)
    return []


def load_fixture_manifest():
    data = json.loads(fixture_manifest_path.read_text(encoding='utf-8'))
    if not isinstance(data, dict):
        raise SystemExit(f"invalid fixture manifest shape: {fixture_manifest_path.as_posix()}")
    return data


def run_repo_scan():
    proc = subprocess.run(
        ['git', 'ls-files', '--', 'config.toml.example', 'cmd/**', 'internal/config/**'],
        cwd=root,
        text=True,
        capture_output=True,
        check=True,
    )
    rel_paths = [p for p in proc.stdout.splitlines() if should_scan_repo_path(p)]

    findings = []
    for rel_path in rel_paths:
        findings.extend(scan_path(root / rel_path))

    if findings:
        print('\n'.join(findings), file=sys.stderr)
        raise SystemExit(1)


def run_fixture_self_test():
    manifest = load_fixture_manifest()
    fixture_dir = fixture_manifest_path.parent
    discovered = sorted(path.name for path in fixture_dir.glob(fixture_glob))

    failures = []
    missing_manifest = sorted(set(discovered) - set(manifest))
    extra_manifest = sorted(set(manifest) - set(discovered))
    for name in missing_manifest:
        failures.append(f"{name}: discovered by {fixture_glob} but missing from retired-manifest.json")
    for name in extra_manifest:
        failures.append(f"{name}: listed in retired-manifest.json but not found on disk")

    for name in discovered:
        expectation = manifest.get(name)
        path = fixture_dir / name

        # Exercise BOTH scan engines against every fixture, not just whichever
        # one scan_toml() would naturally pick for this interpreter. The
        # fallback lexer (scan_toml_with_fallback) previously went completely
        # unexercised whenever tomllib was importable (true on any Python
        # >=3.11, i.e. every currently supported CI/dev environment), so a
        # regression in its own EOF fail-closed handling could land with a
        # fully green self-test. Running both engines here means the
        # fallback's correctness is proven on every self-test invocation,
        # never left as untested dead code.
        engines = [('tomllib', scan_toml_with_tomllib)] if tomllib is not None else []
        engines.append(('fallback', scan_toml_with_fallback))

        for engine_name, engine_fn in engines:
            findings = engine_fn(path)
            rejected = bool(findings)
            label = f"{name} [{engine_name}]"

            if expectation == 'accept':
                if rejected:
                    detail = '; '.join(findings)
                    failures.append(f"{label}: expected clean, got findings: {detail}")
                else:
                    print(f"PASS {label}: clean as expected")
                continue

            if expectation == 'reject':
                if rejected:
                    print(f"PASS {label}: rejected as expected")
                else:
                    failures.append(f"{label}: expected rejection, got clean")
                continue

            failures.append(f"{label}: unknown expectation {expectation!r} in retired-manifest.json")

    if failures:
        print('\n'.join(f"FAIL {failure}" for failure in failures), file=sys.stderr)
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
    scan_with_mode repo
    ;;
  --self-test)
    scan_with_mode self-test
    scan_with_mode repo
    echo "self-test passed: fixtures matched expectations and the tracked tree is clean"
    ;;
  *)
    echo "usage: scripts/check-retired-architecture.sh [--self-test]" >&2
    exit 2
    ;;
esac