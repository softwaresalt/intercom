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
#     Scans tracked files in internal/** (excluding *_test.go and any
#     testdata/ directory), config.toml.example, and cmd/**. Exits 0 when no
#     retired-architecture token is found as a Go identifier or TOML key.
#   scripts/check-retired-architecture.sh --self-test
#     Verifies the committed TOML fixture suite in scripts/testdata/retired-*.toml
#     against scripts/testdata/retired-manifest.json and the committed Go fixture
#     suite in scripts/testdata/retiredgo/*.go against
#     scripts/testdata/retiredgo-manifest.json, then verifies the real tracked
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
from __future__ import annotations

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

fixture_suites = [
    {
        "name": "toml",
        "glob": "retired-*.toml",
        "manifest_path": root / "scripts" / "testdata" / "retired-manifest.json",
        "engines": ["toml"],
    },
    {
        "name": "go",
        "glob": "retiredgo/*.go",
        "manifest_path": root / "scripts" / "testdata" / "retiredgo-manifest.json",
        "engines": ["go"],
    },
]

go_identifier_re = re.compile(r'\b[A-Za-z_][A-Za-z0-9_]*\b')
bare_key_re = re.compile(r'[A-Za-z_][A-Za-z0-9_]*')
component_re = re.compile(r'[A-Z]+(?=[A-Z][a-z]|$)|[A-Z]?[a-z]+|[0-9]+')


def should_scan_repo_path(path: str) -> bool:
    if path == 'config.toml.example':
        return True
    if path.startswith('cmd/'):
        return path.endswith('.go')
    if not path.startswith('internal/'):
        return False
    if '/testdata/' in path or path.endswith('_test.go'):
        return False
    return path.endswith('.go')


def engine_for_path(path: Path):
    if path.name == 'config.toml.example':
        return 'toml'
    if path.suffix == '.go':
        return 'go'
    if path.suffix == '.toml':
        return 'toml'
    return None


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


def self_test_engines_for_name(engine_name: str):
    if engine_name == 'go':
        return [('go', scan_go)]

    if engine_name == 'toml':
        engines = [('tomllib', scan_toml_with_tomllib)] if tomllib is not None else []
        engines.append(('fallback', scan_toml_with_fallback))
        return engines

    return []


def scan_path(path: Path):
    engine_name = engine_for_path(path)
    if engine_name == 'go':
        return scan_go(path)
    if engine_name == 'toml':
        return scan_toml(path)
    return []


def select_repo_paths():
    proc = subprocess.run(
        ['git', 'ls-files', '--', 'config.toml.example', 'cmd/**', 'internal/**'],
        cwd=root,
        text=True,
        capture_output=True,
        check=True,
    )
    return sorted(path for path in proc.stdout.splitlines() if should_scan_repo_path(path))


def expected_internal_repo_paths():
    proc = subprocess.run(
        ['git', 'ls-files', '--', 'internal/**'],
        cwd=root,
        text=True,
        capture_output=True,
        check=True,
    )

    expected = []
    for path in proc.stdout.splitlines():
        if not path.startswith('internal/'):
            continue
        if '/testdata/' in path or path.endswith('_test.go'):
            continue
        if path.endswith('.go'):
            expected.append(path)
    return sorted(expected)


def load_fixture_manifest(manifest_path: Path):
    data = json.loads(manifest_path.read_text(encoding='utf-8'))
    if not isinstance(data, dict):
        raise SystemExit(f"invalid fixture manifest shape: {manifest_path.as_posix()}")
    return data


def run_repo_scan():
    rel_paths = select_repo_paths()

    findings = []
    for rel_path in rel_paths:
        findings.extend(scan_path(root / rel_path))

    if findings:
        print('\n'.join(findings), file=sys.stderr)
        raise SystemExit(1)


def report_assertion(name: str, ok: bool, success: str, failure: str, failures: list[str]):
    if ok:
        print(f"PASS {name}: {success}")
        return

    print(f"FAIL {name}: {failure}")
    failures.append(f"{name}: {failure}")


def run_repo_selection_self_test():
    rel_paths = select_repo_paths()
    internal_actual = sorted(path for path in rel_paths if path.startswith('internal/'))
    internal_expected = expected_internal_repo_paths()
    failures: list[str] = []

    missing_internal = sorted(set(internal_expected) - set(internal_actual))
    extra_internal = sorted(set(internal_actual) - set(internal_expected))
    report_assertion(
        # Guard against a vacuous pass: internal_actual == internal_expected is
        # trivially true when both are empty (e.g. a broken 'internal/**'
        # pathspec, wrong cwd, or a shallow/partial clone silently returning
        # nothing from both independent git ls-files calls). Asserting the
        # independently-derived expected set is itself non-empty closes that
        # gap without weakening the structural equality check below.
        'selection internal non-empty',
        len(internal_expected) > 0,
        f"independently-derived expected internal/** set is non-empty ({len(internal_expected)} paths)",
        'independently-derived expected internal/** set was empty -- git ls-files -- internal/** '
        'likely returned nothing; the structural inclusion assertion below would pass vacuously',
        failures,
    )

    report_assertion(
        'selection structural inclusion',
        internal_actual == internal_expected,
        f"selected every tracked internal non-test, non-testdata Go file ({len(internal_actual)} paths)",
        f"missing={missing_internal or ['none']} extra={extra_internal or ['none']}",
        failures,
    )

    report_assertion(
        'selection internal exclusions',
        not any(path.startswith('internal/') and (path.endswith('_test.go') or '/testdata/' in path) for path in rel_paths),
        'excluded internal test files and internal testdata paths',
        'selected an internal test or testdata path',
        failures,
    )

    predicate_guard = not should_scan_repo_path('scripts/testdata/retiredgo/x.go')
    report_assertion(
        'selection self-scan guard',
        not any(path.startswith('scripts/') for path in rel_paths) and predicate_guard,
        'did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go',
        'self-scan guard failed for scripts/ selection or predicate probe',
        failures,
    )

    report_assertion(
        'selection non-empty',
        len(rel_paths) > 0,
        f"selected {len(rel_paths)} tracked repo paths",
        'selected zero repo paths',
        failures,
    )

    report_assertion(
        'dispatch config.toml.example',
        engine_for_path(Path('config.toml.example')) == 'toml',
        'engine_for_path routes config.toml.example to the TOML engine',
        f"engine_for_path returned {engine_for_path(Path('config.toml.example'))!r}",
        failures,
    )

    # The prior assertion proves engine_for_path() dispatches correctly in
    # isolation, but does not prove config.toml.example is ever actually
    # reached by a real repo scan. Close that gap end-to-end: confirm the
    # real selection set (the same rel_paths a repo-mode run would scan)
    # includes it, so the P0 dead-dispatch bug this task fixed cannot
    # regress silently via a selection-side change instead of a
    # dispatch-side one.
    report_assertion(
        'selection includes config.toml.example',
        'config.toml.example' in rel_paths,
        'real selection set includes config.toml.example end-to-end',
        'config.toml.example was not present in the real selection set',
        failures,
    )

    if failures:
        raise SystemExit(1)


def run_fixture_self_test():
    failures = []

    for suite in fixture_suites:
        manifest_path = suite['manifest_path']
        manifest = load_fixture_manifest(manifest_path)
        fixture_dir = manifest_path.parent
        discovered_paths = sorted(fixture_dir.glob(suite['glob']), key=lambda path: path.as_posix())
        discovered = [path.name for path in discovered_paths]

        missing_manifest = sorted(set(discovered) - set(manifest))
        extra_manifest = sorted(set(manifest) - set(discovered))
        for name in missing_manifest:
            failures.append(f"{name}: discovered by {suite['glob']} but missing from {manifest_path.name}")
        for name in extra_manifest:
            failures.append(f"{name}: listed in {manifest_path.name} but not found on disk")

        for path in discovered_paths:
            name = path.name
            expectation = manifest.get(name)
            engine_name = engine_for_path(path)

            if engine_name not in suite['engines']:
                failures.append(
                    f"{name}: suite {suite['name']} expected engines {suite['engines']} but engine_for_path returned {engine_name!r}"
                )
                continue

            engines = self_test_engines_for_name(engine_name)
            if not engines:
                failures.append(f"{name}: no self-test engines configured for {engine_name!r}")
                continue

            for engine_label, engine_fn in engines:
                findings = engine_fn(path)
                rejected = bool(findings)
                label = f"{name} [{engine_label}]"

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

                failures.append(f"{label}: unknown expectation {expectation!r} in {manifest_path.name}")

    if failures:
        print('\n'.join(f"FAIL {failure}" for failure in failures), file=sys.stderr)
        raise SystemExit(1)


if mode == 'repo':
    run_repo_scan()
elif mode == 'self-test':
    run_fixture_self_test()
    run_repo_selection_self_test()
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