package pysem

import "unicode"

// finalSigma is the codepoint CPython's str.lower() emits for U+03A3 GREEK
// CAPITAL LETTER SIGMA when the Unicode SpecialCasing.txt "Final_Sigma"
// context condition holds (U+03C2 GREEK SMALL LETTER FINAL SIGMA). Every
// other occurrence of Sigma lowercases to the regular U+03C3 GREEK SMALL
// LETTER SIGMA, which is exactly what unicode.ToLower already returns.
const (
	greekCapitalSigma = 0x03A3
	greekFinalSigma   = 0x03C2
)

// IsUpper reports whether r is uppercase under CPython's str.isupper()
// single-character semantics: the Unicode derived "Uppercase" property
// (category Lu, plus the Other_Uppercase codepoints in otherUpperTable --
// Roman numerals and circled/squared Latin letters -- that carry the
// property without belonging to Lu).
func IsUpper(r rune) bool {
	return unicode.IsUpper(r) || unicode.Is(otherUpperTable, r)
}

// IsLower reports whether r is lowercase under CPython's str.islower()
// single-character semantics: the Unicode derived "Lowercase" property
// (category Ll, plus the Other_Lowercase codepoints in otherLowerTable --
// modifier letters, small Roman numerals, and circled Latin letters --
// that carry the property without belonging to Ll).
func IsLower(r rune) bool {
	return unicode.IsLower(r) || unicode.Is(otherLowerTable, r)
}

// Lower reproduces CPython's str.lower() over s, including the one
// unconditional multi-rune SpecialCasing.txt mapping (U+0130 LATIN CAPITAL
// LETTER I WITH DOT ABOVE -> U+0069 U+0307, i.e. "i" followed by a
// combining dot above) and the Final_Sigma context condition for U+03A3.
// Every other rune uses unicode.ToLower's simple case mapping, which
// already agrees with CPython for ordinary 1:1 mappings including
// titlecase letters such as U+01C5 -> U+01C6 and already-lowercase letters
// such as U+00DF (which CPython's lower() leaves unchanged: the ß -> "ss"
// expansion is a lower-to-upper-only special case, never applied by
// lower()).
func Lower(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+4)
	for i, r := range runes {
		switch r {
		case 0x0130: // LATIN CAPITAL LETTER I WITH DOT ABOVE
			out = append(out, 0x0069, 0x0307)
		case greekCapitalSigma:
			if isFinalSigma(runes, i) {
				out = append(out, greekFinalSigma)
			} else {
				out = append(out, unicode.ToLower(r))
			}
		default:
			out = append(out, unicode.ToLower(r))
		}
	}
	return string(out)
}

// isCased reports whether r has the Unicode Cased property for the
// practical purposes of the Final_Sigma condition: general category Lu,
// Ll, or Lt. (The derived Cased property also covers a small set of
// Other_Uppercase/Other_Lowercase modifier letters outside these three
// categories; none appear in this port's supported corpus.)
func isCased(r rune) bool {
	return IsUpper(r) || IsLower(r) || unicode.IsTitle(r)
}

// isCaseIgnorable reports whether r has the Unicode Case_Ignorable
// property used by the Final_Sigma condition to skip over intervening
// punctuation/marks when looking for the nearest cased letter: general
// category Mn, Me, Cf, Lm, or Sk, or one of the specific
// MidLetter/MidNumLet/Single_Quote word-break punctuation marks that also
// carry Case_Ignorable.
func isCaseIgnorable(r rune) bool {
	switch r {
	case 0x0027, // APOSTROPHE (Single_Quote)
		0x002E,         // FULL STOP (MidNumLet)
		0x003A,         // COLON (MidLetter)
		0x00B7,         // MIDDLE DOT (MidLetter)
		0x0387,         // GREEK ANO TELEIA (MidLetter)
		0x05F4,         // HEBREW PUNCTUATION GERSHAYIM (MidLetter)
		0x2018, 0x2019, // single quotation marks (MidNumLet)
		0x2024, // ONE DOT LEADER (MidNumLet)
		0x2027: // HYPHENATION POINT (MidLetter)
		return true
	}
	return unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf, unicode.Lm, unicode.Sk)
}

// isFinalSigma implements the Unicode SpecialCasing.txt Final_Sigma
// context condition for the capital Sigma at runes[i]: the nearest
// non-case-ignorable rune before i must be cased, AND the nearest
// non-case-ignorable rune after i (if any) must NOT be cased.
func isFinalSigma(runes []rune, i int) bool {
	before := false
	for j := i - 1; j >= 0; j-- {
		if isCaseIgnorable(runes[j]) {
			continue
		}
		before = isCased(runes[j])
		break
	}
	if !before {
		return false
	}
	for j := i + 1; j < len(runes); j++ {
		if isCaseIgnorable(runes[j]) {
			continue
		}
		if isCased(runes[j]) {
			return false
		}
		break
	}
	return true
}
