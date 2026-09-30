// Package pysem re-derives the subset of CPython's str/re-module Unicode
// semantics that ported gatecheck engines depend on but Go's unicode/
// regexp packages do not reproduce exactly (see plan C-4).
package pysem

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Strip reproduces str.strip() called with no arguments: it trims leading
// and trailing runes for which IsSpace reports true, leaving interior
// whitespace untouched.
func Strip(s string) string {
	runes := []rune(s)
	start := 0
	for start < len(runes) && IsSpace(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && IsSpace(runes[end-1]) {
		end--
	}
	return string(runes[start:end])
}

// isLineBoundary reports whether r is one of the line-boundary runes
// recognized by CPython's str.splitlines(): \n, \r, \v, \f, \x1c, \x1d,
// \x1e, \x85 (NEL) and U+2028/U+2029 (LINE/PARAGRAPH SEPARATOR). \r\n is
// handled as a combined two-rune boundary by the caller.
func isLineBoundary(r rune) bool {
	switch r {
	case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
		return true
	}
	return false
}

// SplitLines reproduces str.splitlines() (called with no arguments, i.e.
// keepends=False): it splits s on the boundaries enumerated by
// isLineBoundary, treating a "\r\n" pair as a single boundary, dropping
// each boundary from the output, and never emitting a trailing empty
// element for a string that ends exactly on a boundary. An empty input
// yields an empty (non-nil) slice, matching Python's `"".splitlines() ==
// []` rather than `[""]`.
func SplitLines(s string) []string {
	runes := []rune(s)
	out := make([]string, 0)
	start := 0
	i := 0
	for i < len(runes) {
		r := runes[i]
		if !isLineBoundary(r) {
			i++
			continue
		}
		end := i
		if r == '\r' && i+1 < len(runes) && runes[i+1] == '\n' {
			i += 2
		} else {
			i++
		}
		out = append(out, string(runes[start:end]))
		start = i
	}
	if start < len(runes) {
		out = append(out, string(runes[start:]))
	}
	return out
}

// Repr reproduces repr(str) as used by gatecheck's "!r"-formatted
// findings: the quote-choice rule (prefer single quotes, switch to double
// quotes only when the string contains a literal single quote and no
// double quote), backslash escaping of the chosen quote character and of
// literal backslashes, the \n/\r/\t short escapes, \xNN/\uNNNN/\UNNNNNNNN
// escapes for non-printable runes (str.isprintable(), reproduced here via
// unicode.IsPrint -- both definitions are "category L/M/N/P/S, plus the
// ASCII space, and nothing else"), and printable non-ASCII runes kept
// verbatim.
func Repr(s string) string {
	quote := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == rune(quote) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, r)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r <= 0xffff:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// WordBoundary reproduces the regular-expression `\b` assertion (over
// pysem's IsWord, not Go's narrower ASCII word-class) at the given UTF-8
// byte offset into text: true when exactly one of the rune immediately
// before offset and the rune immediately at offset is a "word" rune (an
// absent rune, at either the start or the end of text, counts as
// not-a-word).
func WordBoundary(text string, offset int) bool {
	var beforeWord, afterWord bool
	if offset > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:offset])
		if r != utf8.RuneError || size > 1 {
			beforeWord = IsWord(r)
		}
	}
	if offset < len(text) {
		r, size := utf8.DecodeRuneInString(text[offset:])
		if r != utf8.RuneError || size > 1 {
			afterWord = IsWord(r)
		}
	}
	return beforeWord != afterWord
}
