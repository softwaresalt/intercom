"""Canonical Go non-code masker shared by the repository's Go-source gates.

032.002-T (shipment 029-S, plan unit 2 / U2-T2): this module is the ONE
definition of mask_go_non_code() (AC-2.1). It is the retired-architecture
SUPERSET chosen by 032.001-T -- see the CANONICAL GO MASKER DECISION note in
scripts/check-retired-architecture.sh for the rationale and the enumerated
write-path verdict delta set. The body below is moved verbatim from the
former check-retired-architecture.sh heredoc (behaviour-preserving with
respect to the chosen canonical implementation); it is imported by
scripts/lib/retired_arch.py and scripts/check-write-path-precondition.sh.

mask_go_non_code() blanks out comment, interpreted-string, rune, and
raw-string literal contents while preserving total length and line
structure, so a token appearing only in non-code text is never mistaken for
code and finding line numbers stay aligned with the source file. A raw-string
literal whose ENTIRE content matches struct_tag_re is left visible (struct
tags carry config keys); an unterminated raw string at EOF fails closed.

This module reads no process state at import time and has no side effects.
"""

from __future__ import annotations

import re

__all__ = ['mask_go_non_code', 'struct_tag_re']

# Fix (post-015.007-T review, Copilot): the original `(\s*\w+:"[^"]*")+`
# was more permissive than Go's actual struct tag grammar in two ways --
# (a) `\w+` allowed a key to start with a digit, which is not a valid Go
# identifier and never appears as a real struct tag key; and (b) `\s*`
# before every repetition (including between pairs) allowed adjacent
# `key1:"v1"key2:"v2"` with zero separating whitespace, whereas Go's real
# struct tag format requires pairs to be whitespace-separated. Both gaps
# let non-tag raw string literals that merely look tag-like be classified
# as a struct tag and unmasked, increasing false-positive scan exposure.
# The tightened pattern below requires each key to start with a letter or
# underscore, and requires at least one whitespace character between
# successive pairs (only the very first pair may be preceded by optional
# leading whitespace, matching Go's own leading-space-trim behavior).
struct_tag_re = re.compile(
    r'^\s*[A-Za-z_]\w*:"[^"]*"(?:\s+[A-Za-z_]\w*:"[^"]*")*\s*$'
)


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
