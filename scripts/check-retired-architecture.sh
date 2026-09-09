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
#     Runs three checks: (1) verifies the committed TOML fixture suite in
#     scripts/testdata/retired-*.toml against scripts/testdata/retired-manifest.json
#     and the committed Go fixture suite in scripts/testdata/retiredgo/*.go against
#     scripts/testdata/retiredgo-manifest.json; (2) verifies the real selection
#     logic structurally (every tracked internal/** non-test, non-testdata .go
#     file is selected, internal/** test/testdata paths and scripts/ itself are
#     excluded, selection is non-empty, and config.toml.example dispatches to the
#     TOML engine); (3) verifies the real tracked tree passes a repo scan. Exits 0
#     only when all three checks succeed. Semantics UNCHANGED by 015.001-T.
#   scripts/check-retired-architecture.sh --self-test-integrity (015.001-T)
#     Additive integrity-only mode: runs ONLY checks (1) and (2) above (the
#     fixture suites and the selection-logic self-test). Does NOT run the
#     repo scan (3). Used by ci.yml's integrity step so that step stays
#     unconditionally blocking while the separately-toggled verdict step
#     (see below) owns the real repo-scan enforcement decision.
#
# Every invocation emits a GitHub Actions ::notice:: line naming the
# resolved mode (repo / self-test / self-test-integrity) so enforcement
# posture is falsifiable from run history (H8).

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
    {
        # 015.002-T: masking bypass seam. This suite is SEPARATE from the
        # shared "go" suite above (untouched, per rev-3 C-P1-2) and is
        # reject-only (per rev-4 ESC-P2-a). It still declares
        # engines: ['go'] so the pre-existing `engine_name not in
        # suite['engines']` guard (engine_for_path always returns 'go' for
        # a .go file) passes unchanged; only the ENGINE FUNCTION is
        # overridden via 'engine_override' below to the unmasked scan
        # (scan_go(path, mask=False)), so suite-to-engine selection is
        # suite-scoped rather than keyed on engine_for_path.
        "name": "go-differential",
        "glob": "retiredgo-differential/*.go",
        "manifest_path": root / "scripts" / "testdata" / "retiredgo-differential-manifest.json",
        "engines": ["go"],
        "engine_override": [("go-unmasked", lambda path: scan_go(path, mask=False))],
    },
]

go_identifier_re = re.compile(r'\b[A-Za-z_][A-Za-z0-9_]*\b')
bare_key_re = re.compile(r'[A-Za-z_][A-Za-z0-9_]*')
# component_re removed by 015.005-T: split_camel_acronym() replaces the
# regex-based acronym/camelCase splitter (V1 fix -- see split_camel_acronym
# docstring for the acronym+plural-suffix disambiguation it now performs).


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


def split_camel_acronym(chunk: str):
    """Segment one underscore-delimited chunk into camelCase/acronym
    components.

    015.005-T (V1 fix): a maximal uppercase run of length >= 2 immediately
    followed by lowercase letters is ambiguous between two cases:
      (a) genuine acronym-then-new-word, e.g. 'IPCNames' -> ['IPC', 'Names']
          (the last uppercase letter of the run starts the new word); and
      (b) acronym-then-bare-plural-suffix, e.g. 'ChannelIDs' -> keep 'IDs'
          together as ONE token rather than peeling the run's last letter
          into a spurious one-character token ('channel', 'i', 'ds') that
          can never match a 2-part forbidden sequence.
    The two are disambiguated by checking whether the lowercase tail
    following the run is EXACTLY 's' or 'es' (case (b), glue) or something
    longer (case (a), split before the last uppercase letter). This check
    is intentionally NOT anchored to end-of-string: 'TeamIDsCache' must
    still glue 'IDs' even though 'Cache' follows, because the plural-tail
    rule fires the moment the lowercase run itself is exactly 's'/'es',
    regardless of what comes after it.
    """
    n = len(chunk)
    tokens: list[str] = []
    i = 0
    while i < n:
        ch = chunk[i]
        if ch.isdigit():
            j = i
            while j < n and chunk[j].isdigit():
                j += 1
            tokens.append(chunk[i:j])
            i = j
            continue
        if ch.isupper():
            j = i
            while j < n and chunk[j].isupper():
                j += 1
            run_len = j - i
            if run_len == 1:
                k = j
                while k < n and chunk[k].islower():
                    k += 1
                tokens.append(chunk[i:k])
                i = k
                continue
            if j < n and chunk[j].islower():
                k = j
                while k < n and chunk[k].islower():
                    k += 1
                tail = chunk[j:k]
                if tail in ('s', 'es'):
                    tokens.append(chunk[i:k])
                else:
                    tokens.append(chunk[i:j - 1])
                    tokens.append(chunk[j - 1:k])
                i = k
                continue
            tokens.append(chunk[i:j])
            i = j
            continue
        j = i
        while j < n and chunk[j].islower():
            j += 1
        tokens.append(chunk[i:j])
        i = j
    return tokens


def split_identifier(name: str):
    parts = []
    for chunk in name.split('_'):
        parts.extend(token.lower() for token in split_camel_acronym(chunk))
    return parts or [name.lower()]


_VOCAB_WORDS = sorted(
    {word for parts in forbidden_parts.values() for word in parts},
    key=len,
    reverse=True,
)


def segment_whole(word: str):
    """Whole-part decomposition (015.005-T, V3 / P3 concat model): return
    every way `word` can be segmented EXACTLY and COMPLETELY into known
    vocabulary words (`_VOCAB_WORDS`), or [] if no complete segmentation
    exists. This is deliberately NOT a substring search: a word like
    'hostclient' is never treated as containing 'host'+'cli' because the
    leftover 'ent' can never be covered by a vocabulary word, so no
    complete segmentation exists and segment_whole returns [].
    """
    n = len(word)
    dp: list[list[list[str]]] = [[] for _ in range(n + 1)]
    dp[0] = [[]]
    for end in range(1, n + 1):
        segmentations: list[list[str]] = []
        for tok in _VOCAB_WORDS:
            start = end - len(tok)
            if start >= 0 and dp[start] and word[start:end] == tok:
                for prefix in dp[start]:
                    segmentations.append(prefix + [tok])
        dp[end] = segmentations
    return dp[n]


def window_matches(parts, want):
    """Shared window/suffix rule for both matching models (015.005-T): a
    forbidden sequence `want` matches `parts` if some contiguous window of
    `parts` equals `want`, EXCEPT that the final element of the window may
    carry a plural 's'/'es' suffix relative to the final element of
    `want`. The suffix check is scoped to the matched WINDOW, not the end
    of the whole identifier -- e.g. 'TeamIDsCache' matches the
    ['team','id'] window even though 'Cache' follows."""
    span = len(want)
    for i in range(len(parts) - span + 1):
        window = parts[i:i + span]
        if window[:-1] != want[:-1]:
            continue
        last, want_last = window[-1], want[-1]
        if last == want_last:
            return True
        for suffix in ('es', 's'):
            if last.endswith(suffix) and last[:-len(suffix)] == want_last:
                return True
    return False


def matches_forbidden_sequence(parts):
    """Camel/acronym-boundary model: matches a forbidden sequence against
    the case-boundary-derived component list produced by split_identifier
    (covers V1, V2)."""
    for token, want in forbidden_parts.items():
        if len(parts) >= len(want) and window_matches(parts, want):
            return token
    return None


def matches_forbidden_concat(parts):
    """No-separator concat model (covers V3): for each individual
    component that split_identifier could not further decompose by case
    boundary (e.g. the whole of 'socketmode'), attempt a full whole-part
    decomposition and re-apply the same window/suffix rule to it."""
    for component in parts:
        for segmentation in segment_whole(component):
            for token, want in forbidden_parts.items():
                if len(segmentation) >= len(want) and window_matches(segmentation, want):
                    return token
    return None


def matches_forbidden_parts(parts):
    """Single dispatch point (015.005-T) for the two separately named
    matching models. Returns (token, model_name) so findings can report
    which model fired, or None."""
    token = matches_forbidden_sequence(parts)
    if token:
        return token, 'sequence'
    token = matches_forbidden_concat(parts)
    if token:
        return token, 'concat'
    return None


struct_tag_re = re.compile(r'^(\s*\w+:"[^"]*")+\s*$')


def mask_go_non_code(text: str) -> str:
    code, line_comment, block_comment, string, raw_string, rune = range(6)
    state = code
    out = []
    raw_buf: list[str] = []
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
                raw_buf = []
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
            if ch == '`':
                # 015.007-T (V4): decide whole-content struct-tag visibility
                # only now that the ENTIRE raw_string content is known --
                # this is why the content must be buffered rather than
                # masked char-by-char as it streams past. Only a literal
                # whose entire content matches the tag grammar is unmasked;
                # everything else (multiline strings, SQL/template
                # backtick literals, doc-comment-quoted backticks that
                # never reach this state at all) keeps the prior
                # fully-masked behavior unchanged.
                content = ''.join(raw_buf)
                if struct_tag_re.match(content):
                    out.append(content)
                else:
                    out.append(''.join('\n' if c == '\n' else ' ' for c in raw_buf))
                out.append(' ')
                i += 1
                state = code
                raw_buf = []
                continue
            raw_buf.append(ch)
            i += 1
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

    if state == raw_string and raw_buf:
        # Unterminated raw string at EOF: fail closed by masking the
        # buffered content exactly as the pre-015.007-T behavior did --
        # never unmask a literal whose content could not be fully
        # determined to be (or not be) a complete struct tag.
        out.append(''.join('\n' if c == '\n' else ' ' for c in raw_buf))

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


def scan_go(path: Path, mask: bool = True):
    # 015.002-T: mask=True (default) is the real, unchanged production and
    # self-test path. mask=False is a masking-bypass seam used ONLY by the
    # go-differential fixture suite's engine_override to prove the masked
    # path is not false-green on content that hides a retired token behind
    # comment/string masking.
    findings = []
    text = path.read_text(encoding='utf-8')
    masked = mask_go_non_code(text) if mask else text
    for line_no, line in enumerate(masked.splitlines(), start=1):
        for match in go_identifier_re.finditer(line):
            result = matches_forbidden_parts(split_identifier(match.group(0)))
            if result:
                token, model = result
                findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in Go identifier {match.group(0)!r} (via {model} model)")
    return findings


def decompose_toml_key(segment: str):
    """015.008-T (V5): decompose a single raw TOML key/table-header
    segment string into split_identifier-style component tokens. Hyphen
    normalisation lives HERE, at the TOML boundary -- not inside the
    shared split_identifier tokenizer -- because bare TOML keys may use
    hyphens (TOML bare-key charset is [A-Za-z0-9_-]) that Go identifiers
    never do. In the fallback lexer, bare_key_re ([A-Za-z_][A-Za-z0-9_]*)
    already splits a hyphenated key like 'channel-id' into two separate
    regex matches before this function ever sees either half, so the
    hyphen never actually reaches this normalisation there -- that
    fixture is caught by composed-path matching instead. This function
    still performs the replace defensively for the tomllib site, where a
    hyphenated key arrives as one whole dict-key string."""
    return split_identifier(segment.replace('-', '_'))


def compose_toml_parts(path_segments: list[str]):
    """015.008-T (V5): flatten decompose_toml_key() over every segment of
    a composed TOML key path (table prefix segments plus the leaf key)
    into ONE parts list, so matches_forbidden_parts can see a forbidden
    sequence that only appears once the path is composed -- e.g.
    [channel] + id, or a top-level dotted channel.id key -- neither of
    which contains the forbidden pair within a single segment alone."""
    parts: list[str] = []
    for segment in path_segments:
        parts.extend(decompose_toml_key(segment))
    return parts


def report_toml_key(path: Path, key_path: list[str], segment: str, token: str, model: str) -> str:
    return (
        f"{path.as_posix()}: retired token {token!r} in TOML key path "
        f"{'.'.join(key_path)!r} (segment {segment!r}) (via {model} model)"
    )


def walk_toml_value(path: Path, prefix: list[str], value, findings: list[str]):
    if isinstance(value, dict):
        for key, child in value.items():
            key_path = prefix + [key]
            result = matches_forbidden_parts(compose_toml_parts(key_path))
            if result:
                token, model = result
                findings.append(report_toml_key(path, key_path, key, token, model))
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
        # 015.009-T (V6): capture the multiline flags BEFORE
        # strip_toml_comment mutates them for THIS line, and use that
        # pre-line state for the suppression decision below. A line that
        # starts OUTSIDE a multiline string remains eligible for LHS
        # inspection even though processing it OPENS one (the key is
        # legitimately written on this same line); a line that starts
        # INSIDE a multiline string remains suppressed even though
        # processing it CLOSES one (a delimiter-only closing line is
        # never itself a real key=value assignment).
        pre_line_multiline = state['in_multiline_basic'] or state['in_multiline_literal']
        line = strip_toml_comment(raw_line, state).strip()
        if not line:
            continue
        if not pre_line_multiline and line.startswith('[') and line.endswith(']'):
            header = line.strip('[]').strip()
            current_table = bare_key_re.findall(header)
            # 015.008-T (V5): check the header's OWN path as one composed
            # unit (covers a dotted header like [channel.id]), not each
            # header segment in isolation.
            result = matches_forbidden_parts(compose_toml_parts(current_table))
            if result:
                token, model = result
                findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML table key {'.'.join(current_table)!r} (via {model} model)")
            continue
        if pre_line_multiline or '=' not in line:
            continue
        lhs = line.split('=', 1)[0]
        keys = current_table + bare_key_re.findall(lhs)
        # 015.008-T (V5): the fallback must ALSO match over the composed
        # path (current_table + this line's own key segments), exactly
        # like the tomllib site, or a [channel] + id fixture is REJECT
        # under tomllib and CLEAN under fallback -- a dual-engine
        # self-test failure that can never converge.
        result = matches_forbidden_parts(compose_toml_parts(keys))
        if result:
            token, model = result
            findings.append(f"{path.as_posix()}:{line_no}: retired token {token!r} in TOML key {'.'.join(keys)!r} (via {model} model)")

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

    # 015.004-T: framework invariants. Sequenced to run before the
    # per-fixture verdict loop below so a broken suite (empty, or missing
    # accept/reject coverage) is reported clearly rather than only via a
    # vacuous "0 fixtures checked, 0 failures" pass. The differential suite
    # is the ONLY reject-only suite, named explicitly here as an allow-list
    # entry (not a general escape hatch) -- a future all-reject suite that
    # is NOT named here still trips 'suite has accept and reject coverage'.
    reject_only_suites = {'go-differential'}

    for suite in fixture_suites:
        manifest_path = suite['manifest_path']
        manifest = load_fixture_manifest(manifest_path)
        fixture_dir = manifest_path.parent
        discovered_paths = sorted(fixture_dir.glob(suite['glob']), key=lambda path: path.as_posix())
        discovered = [path.name for path in discovered_paths]

        report_assertion(
            f"suite {suite['name']} non-empty",
            len(discovered) > 0,
            f"suite discovered {len(discovered)} fixture(s)",
            f"suite {suite['name']} discovered zero fixtures via {suite['glob']}",
            failures,
        )

        expectations = {manifest.get(name) for name in discovered}
        if suite['name'] in reject_only_suites:
            report_assertion(
                f"suite {suite['name']} reject-only exemption",
                len(expectations) > 0 and expectations <= {'reject'},
                f"suite is a named reject-only exemption and all {len(discovered)} fixture(s) are 'reject'",
                f"suite {suite['name']} is declared reject-only but manifest expectations are {sorted(e for e in expectations if e is not None)!r}",
                failures,
            )
        else:
            report_assertion(
                f"suite {suite['name']} has accept and reject coverage",
                'accept' in expectations and 'reject' in expectations,
                "suite has at least one 'accept' and one 'reject' fixture",
                f"suite {suite['name']} is missing accept and/or reject coverage (found {sorted(e for e in expectations if e is not None)!r})",
                failures,
            )

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

            # 015.002-T: suite-scoped engine override takes precedence over
            # the engine-name-keyed lookup below, so a suite can exercise a
            # different engine function (e.g. unmasked scan_go) for .go
            # fixtures without affecting the shared 'go' suite, which has
            # no 'engine_override' key and is untouched.
            engines = suite.get('engine_override') or self_test_engines_for_name(engine_name)
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
elif mode == 'self-test-integrity':
    # 015.001-T: additive integrity-only mode -- fixture + selection
    # self-tests only, deliberately NOT run_repo_scan(). --self-test above
    # is UNCHANGED (still runs the repo scan); this mode exists so ci.yml's
    # integrity step can stay unconditionally blocking without also owning
    # the toggle-governed repo-scan verdict.
    run_fixture_self_test()
    run_repo_selection_self_test()
else:
    raise SystemExit(f'unknown mode: {mode}')
PY
}

case "${1:-}" in
  "")
    echo "::notice::retired-arch gate mode=repo"
    scan_with_mode repo
    ;;
  --self-test)
    echo "::notice::retired-arch gate mode=self-test"
    scan_with_mode self-test
    scan_with_mode repo
    echo "self-test passed: fixtures matched expectations and the tracked tree is clean"
    ;;
  --self-test-integrity)
    echo "::notice::retired-arch gate mode=self-test-integrity"
    scan_with_mode self-test-integrity
    echo "self-test-integrity passed: fixtures matched expectations (repo scan skipped, 015.001-T)"
    ;;
  *)
    echo "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]" >&2
    exit 2
    ;;
esac