package tui

import (
	"strings"
	"testing"
	"unicode"
)

func TestSanitizeNeutralisesTerminalEscapeInjection(t *testing.T) {
	cases := map[string]string{
		"csi":            "a\x1b[31mred\x1b[0m",
		"osc title":      "x\x1b]0;pwned\x07y",
		"c1 csi":         "a\u009b2Jb",
		"c1 osc":         "a\u009d0;t\u009cb",
		"del":            "a\x7fb",
		"nul":            "a\x00b",
		"cr overwrite":   "safe\rEVIL",
		"bidi override":  "ok\u202eexe.txt",
		"bidi isolate":   "ok\u2066hidden\u2069",
		"zero width":     "a\u200bb\ufeffc",
		"line separator": "a\u2028b\u2029c",
		"invalid utf8":   "a\xffb",
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
