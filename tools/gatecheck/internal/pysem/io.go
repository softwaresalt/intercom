package pysem

import (
	"errors"
	"os"
	"strings"
	"unicode/utf8"
)

// ErrInvalidUTF8 is returned by ReadText and GitText when their input is
// not valid UTF-8, reproducing CPython's Path.read_text(encoding='utf-8')
// raising UnicodeDecodeError on the same bytes. Ported engines fold this
// into the ED-2 error class: a synthetic "::error::"-prefixed finding at
// exit 1, replacing Python's traceback (plan §3 C-4, §3 ED-2).
var ErrInvalidUTF8 = errors.New("pysem: invalid UTF-8 input")

// translateNewlines reproduces the universal-newline translation Python
// applies on text-mode reads: "\r\n" becomes "\n", and any remaining lone
// "\r" becomes "\n" as well. The two-character sequence is replaced first
// so a CRLF pair is never double-translated into two newlines.
func translateNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

// ReadText reproduces Path.read_text(encoding='utf-8'): it reads the file
// at path, rejects invalid UTF-8 with ErrInvalidUTF8 (ED-2) before doing
// anything else with the bytes, and otherwise applies translateNewlines.
// A file-system read error (missing file, permission denied, etc.) is
// returned unwrapped, exactly as os.ReadFile reports it.
func ReadText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(data) {
		return "", ErrInvalidUTF8
	}
	return translateNewlines(string(data)), nil
}

// GitText reproduces the text-mode decoding of a subprocess.run(...,
// text=True) output: reject invalid UTF-8 with ErrInvalidUTF8 (ED-2), then
// apply the same universal-newline translation as ReadText.
func GitText(out []byte) (string, error) {
	if !utf8.Valid(out) {
		return "", ErrInvalidUTF8
	}
	return translateNewlines(string(out)), nil
}

// PrecededByWordOrDot is the explicit replacement for the lookbehind
// `(?<![\w.])`, which RE2 cannot express: it reports whether the rune
// immediately before the given UTF-8 byte offset is a "word" rune (per
// IsWord) or a literal '.'. Offset 0 (nothing before it) is never
// preceded.
func PrecededByWordOrDot(text string, offset int) bool {
	if offset <= 0 {
		return false
	}
	r, size := utf8.DecodeLastRuneInString(text[:offset])
	if r == utf8.RuneError && size <= 1 {
		return false
	}
	return r == '.' || IsWord(r)
}

// FollowedByWord is the explicit replacement for the lookahead
// `(?![\w])`, which RE2 cannot express: it reports whether the rune
// immediately at the given UTF-8 byte offset is a "word" rune (per
// IsWord). An offset at or past the end of text is never followed by
// anything.
func FollowedByWord(text string, offset int) bool {
	if offset >= len(text) {
		return false
	}
	r, size := utf8.DecodeRuneInString(text[offset:])
	if r == utf8.RuneError && size <= 1 {
		return false
	}
	return IsWord(r)
}
