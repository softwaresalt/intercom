// Package pysem reimplements the small set of CPython str/re character- and
// text-classification primitives the Python gate engines relied on
// (str.isspace/isalnum/isdigit/isdecimal/isupper/islower/lower/strip/
// splitlines, repr(), and the write-path selector regex's \w-based
// lookarounds). Go's unicode package and CPython's Unicode database do not
// classify every codepoint identically, so each helper here is deliberately
// re-derived and pinned against tools/gatecheck/internal/pysem/testdata/
// pysem_golden.json (captured from CPython 3.14.3 / unicodedata 16.0.0)
// rather than assumed equivalent to its closest unicode.IsXxx counterpart.
package pysem

import "unicode"

// IsSpace reports whether r is whitespace under CPython's str.isspace()
// semantics: Unicode category Zs, Zl, Zp, or bidirectional class WS, B, or
// S. Go's unicode.IsSpace already covers every case CPython treats as
// whitespace except the four C1-range separator controls FS/GS/RS/US
// (U+001C-U+001F), which CPython's bidirectional-class-B rule includes but
// Go's Latin-1 fast path does not.
func IsSpace(r rune) bool {
	switch r {
	case 0x1C, 0x1D, 0x1E, 0x1F:
		return true
	}
	return unicode.IsSpace(r)
}

// IsWord reports whether r is a "word" character under the semantics the
// write-path engine's selector regex relies on for its \w character class:
// underscore, or CPython's str.isalnum() (isalpha() or isnumeric()).
// unicode.IsLetter covers isalpha() (Unicode category L); unicode.IsNumber
// covers the numeric categories (Nd, Nl, No) that back isnumeric() for
// every rune in this package's supported corpus.
func IsWord(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}

// IsDecimal reports whether r is a decimal digit under CPython's
// str.isdecimal() semantics (Unicode category Nd only). This is exactly
// Go's unicode.IsDigit, which is defined over the same Nd range table.
func IsDecimal(r rune) bool {
	return unicode.IsDigit(r)
}

// IsDigit reports whether r is a digit under CPython's str.isdigit()
// semantics: category Nd (IsDecimal), plus every codepoint whose Unicode
// Numeric_Type is Digit rather than Decimal (superscript/subscript digits,
// circled and parenthesized digits, and a handful of legacy numeral
// blocks -- see digit_table.go).
func IsDigit(r rune) bool {
	return IsDecimal(r) || unicode.Is(extraDigit, r)
}
