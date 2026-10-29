package tui

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Sanitize makes agent-sourced text safe to put on a terminal. Every control
// rune (C0, DEL and C1 such as U+009B CSI / U+009D OSC), the Unicode line and
// paragraph separators and every format rune (bidi override and isolate
// controls, zero-width characters, BOM) is replaced by a visible <U+XXXX>
// marker, and invalid UTF-8 by <invalid>. The operator therefore sees that
// something hidden was present instead of having it silently dropped or,
// worse, interpreted by the terminal (escape injection, text spoofing).
// Newline is kept (callers wrap on it) and tab becomes four spaces.
//
// The hub and the adapter already sanitize some fields; the TUI does not rely
// on that: it is the last hop before the terminal.
func Sanitize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch {
		case r == utf8.RuneError && size == 1:
			b.WriteString("<invalid>")
		case r == '\n':
			b.WriteByte('\n')
		case r == '\t':
			b.WriteString("    ")
		case isHiddenRune(r):
			fmt.Fprintf(&b, "<U+%04X>", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeLine is [Sanitize] for single-line fields (ids, tool names, status
// text): newlines are shown as markers too.
func SanitizeLine(s string) string {
	return strings.ReplaceAll(Sanitize(s), "\n", "<U+000A>")
}

// isHiddenRune reports runes that render as nothing (or as blank filler) and
// so must be shown as a visible marker.
func isHiddenRune(r rune) bool {
	return unicode.IsControl(r) || r == '\u2028' || r == '\u2029' ||
		unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) ||
		unicode.Is(unicode.Variation_Selector, r) ||
		r == '\u2800' || r == '\u3164' || r == '\u115f' || r == '\u1160' ||
		r == '\uffa0' || r == '\u034f' || r == '\u17b4' || r == '\u17b5'
}

// isBlankSummary reports whether s shows nothing meaningful: every rune is
// whitespace, hidden, or a replacement/invalid rune.
func isBlankSummary(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) && !isHiddenRune(r) && r != utf8.RuneError {
			return false
		}
	}
	return true
}
