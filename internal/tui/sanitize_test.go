package tui

import (
	"strings"
	"testing"
	"unicode"
)

func TestSanitizeNeutralisesTerminalEscapeInjection(t *testing.T) {
	cases := map[string]string{
		"csi":                "a\x1b[31mred\x1b[0m",
		"osc title":          "x\x1b]0;pwned\x07y",
		"c1 csi":             "a\u009b2Jb",
		"c1 osc":             "a\u009d0;t\u009cb",
		"del":                "a\x7fb",
		"nul":                "a\x00b",
		"cr overwrite":       "safe\rEVIL",
		"bidi override":      "ok\u202eexe.txt",
		"bidi isolate":       "ok\u2066hidden\u2069",
		"zero width":         "a\u200bb\ufeffc",
		"line separator":     "a\u2028b\u2029c",
		"invalid utf8":       "a\xffb",
		"hangul filler":      "a\u3164b",
		"hangul choseong":    "a\u115fb",
		"hangul jungseong":   "a\u1160b",
		"halfwidth filler":   "a\uffa0b",
		"braille blank":      "a\u2800b",
		"cgj":                "a\u034fb",
		"khmer inherent aq":  "a\u17b4b",
		"khmer inherent aa":  "a\u17b5b",
		"private use":        "a\ue000b",
		"variation selector": "a\ufe0fb",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			out := Sanitize(in)
			for _, r := range out {
				if r == '\n' || r == ' ' {
					continue
				}
				if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' {
					t.Fatalf("Sanitize(%q) = %q still holds %U", in, out, r)
				}
			}
			if out == in {
				t.Fatalf("Sanitize(%q) changed nothing", in)
			}
		})
	}
}

func TestSanitizeMakesRemovalVisible(t *testing.T) {
	got := Sanitize("ok\u202eexe")
	if !strings.Contains(got, "<U+202E>") {
		t.Fatalf("hidden character must be shown, got %q", got)
	}
}

func TestSanitizeKeepsPlainTextNewlinesAndExpandsTabs(t *testing.T) {
	in := "line one\nline\ttwo ünïcode ✓"
	want := "line one\nline    two ünïcode ✓"
	if got := Sanitize(in); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestSanitizeLine(t *testing.T) {
	if got := SanitizeLine("a\nb\x1bc"); strings.ContainsAny(got, "\n\x1b") {
		t.Fatalf("SanitizeLine left a newline or escape: %q", got)
	}
}

func TestIsBlankSummaryRejectsInvisibleAndReplacementOnlyText(t *testing.T) {
	blank := []string{"", " \t\n", "\u3164", "\u2800", "\u200b", "\ufffd", "\xff", "\ue000\n\t", "\u034f\u17b4\u17b5\ufe0f", "\u115f\u1160\uffa0 \ufffd"}
	for _, s := range blank {
		if !isBlankSummary(s) {
			t.Errorf("isBlankSummary(%q) = false, want true", s)
		}
	}
	visible := []string{"git status", "\u3164x", "\ufffdok", "a"}
	for _, s := range visible {
		if isBlankSummary(s) {
			t.Errorf("isBlankSummary(%q) = true, want false", s)
		}
	}
}
