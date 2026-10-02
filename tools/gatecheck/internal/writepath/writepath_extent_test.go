package writepath

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// absoluteOccurrences walks text with the same line cursor scanText uses and
// returns the absolute byte offset of every word-bounded occurrence of sel.
// It fails the test if the cursor cannot follow the text.
func absoluteOccurrences(t *testing.T, text, sel string) []int {
	t.Helper()
	var offsets []int
	cur := 0
	for _, line := range pysem.SplitLines(text) {
		lineStart := cur
		next, ok := advanceCursor(text, cur, line)
		if !ok {
			t.Fatalf("cursor lost sync at byte %d", cur)
		}
		cur = next
		for _, pos := range selectorOccurrences(line, sel) {
			offsets = append(offsets, lineStart+pos)
		}
	}
	return offsets
}

// TestExtractExtent_ReparseWindowsCreateFile is AC-D1.1: the live
// syscall.CreateFile call in internal/pathsafe/reparse_windows.go, located by
// content, is one balanced extent of 8 raw depth-1 segments, and the
// comment-only mentions yield no occurrence after masking.
func TestExtractExtent_ReparseWindowsCreateFile(t *testing.T) {
	const sel = "syscall.CreateFile"
	raw, err := pysem.ReadText(filepath.Join(repoRoot(t), "internal", "pathsafe", "reparse_windows.go"))
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(raw, sel); n < 5 {
		t.Fatalf("unmasked source mentions %s %d times, want the call plus four comment mentions", sel, n)
	}
	masked := gomask.MaskGoNonCode(raw)
	offsets := absoluteOccurrences(t, masked, sel)
	if len(offsets) != 1 {
		t.Fatalf("masked occurrences of %s = %d, want exactly 1 (comment mentions must vanish)", sel, len(offsets))
	}
	ext := extractExtent(masked, offsets[0]+len(sel))
	if ext.kind != extentBalanced {
		t.Fatalf("extent kind = %v, want extentBalanced", ext.kind)
	}
	if len(ext.segments) != 8 {
		t.Fatalf("segments = %d %q, want 8 (7 arguments plus the trailing-comma artifact)", len(ext.segments), ext.segments)
	}
	if got := strings.TrimSpace(ext.segments[1]); got != "0" {
		t.Errorf("segment 1 = %q, want %q", got, "0")
	}
	if got := strings.TrimSpace(ext.segments[4]); got != "syscall.OPEN_EXISTING" {
		t.Errorf("segment 4 = %q, want %q", got, "syscall.OPEN_EXISTING")
	}
	if got := strings.TrimSpace(ext.segments[7]); got != "" {
		t.Errorf("segment 7 = %q, want the empty trailing-comma artifact", got)
	}
}

// TestExtractExtent_Classifications is AC-D1.2 at the extractor level.
func TestExtractExtent_Classifications(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		kind     extentKind
		segments []string
	}{
		{"non-call identifier", "os.Remove", extentNonCall, nil},
		{"non-call value", "os.Remove;", extentNonCall, nil},
		{"non-call form feed is not Go whitespace", "os.Remove\f(a)", extentNonCall, nil},
		{"call", "os.Remove(a)", extentBalanced, []string{"a"}},
		{"call after whitespace and newline", "os.Remove \t\n(a, b)", extentBalanced, []string{"a", " b"}},
		{"nested depth-1 split only", "os.Remove(f(a, b), []int{1, 2}, m[k])", extentBalanced, []string{"f(a, b)", " []int{1, 2}", " m[k]"}},
		{"trailing comma artifact", "os.Remove(a,\n)", extentBalanced, []string{"a", "\n"}},
		{"empty call", "os.Remove()", extentBalanced, []string{""}},
		{"unbalanced EOF", "os.Remove(a, (b", extentUnbalanced, nil},
		{"unbalanced EOF no closer", "os.Remove(", extentUnbalanced, nil},
		{"mismatched closer", "os.Remove(a, [b)", extentMismatched, nil},
		{"mismatched brace", "os.Remove(a}", extentMismatched, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ext := extractExtent(tc.text, len("os.Remove"))
			if ext.kind != tc.kind {
				t.Fatalf("kind = %v, want %v", ext.kind, tc.kind)
			}
			if !reflect.DeepEqual(ext.segments, tc.segments) {
				t.Errorf("segments = %q, want %q", ext.segments, tc.segments)
			}
		})
	}
}

// TestScanText_ClassificationsKeepFindingText is AC-D1.2 at the scan level:
// every classification still yields the unchanged finding text.
func TestScanText_ClassificationsKeepFindingText(t *testing.T) {
	want := []string{"x.go:1: write primitive 'os.Remove' found"}
	for _, text := range []string{"f := os.Remove", "os.Remove(a, (b", "os.Remove(a, [b)", "os.Remove(a)"} {
		if got := scanText("x.go", text); !reflect.DeepEqual(got, want) {
			t.Errorf("scanText(%q) = %q, want %q", text, got, want)
		}
	}
}

// TestScanText_OneFindingPerSelectorPerLine pins that enumerating every
// occurrence still emits at most one finding per (line, selector).
func TestScanText_OneFindingPerSelectorPerLine(t *testing.T) {
	got := scanText("x.go", "os.Remove(a); os.Remove(b); os.Rename(c, d)\nxos.Remove(e)")
	want := []string{
		"x.go:1: write primitive 'os.Remove' found",
		"x.go:1: write primitive 'os.Rename' found",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanText = %q, want %q", got, want)
	}
}

// TestSelectorOccurrences_EnumeratesEvery pins full enumeration, which the
// legacy first-match findSelector did not provide.
func TestSelectorOccurrences_EnumeratesEvery(t *testing.T) {
	line := "os.Remove(a) xos.Remove(b) os.Remove(c) os.Removed(d) .os.Remove(e) os.Remove"
	got := selectorOccurrences(line, "os.Remove")
	want := []int{0, 27, 68}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("occurrences = %v, want %v", got, want)
	}
	for _, pos := range got {
		if line[pos:pos+len("os.Remove")] != "os.Remove" {
			t.Errorf("offset %d does not point at the selector", pos)
		}
	}
	if selectorOccurrences("xos.Remove os.Removed", "os.Remove") != nil {
		t.Error("non-bounded matches must not be enumerated")
	}
}

// TestCursor_MultiLineExtentAcrossSeparators is AC-D1.3: a multi-line extent
// whose lines are separated by \f or \u2028 maps every occurrence back to the
// correct SplitLines line.
func TestCursor_MultiLineExtentAcrossSeparators(t *testing.T) {
	text := "os.Remove(a,\fos.Remove(b)\u2028)\r\nnoop\u2029os.Remove(c)\u0085os.Rename(d, e)"
	lines := pysem.SplitLines(text)
	if len(lines) != 6 {
		t.Fatalf("SplitLines = %q, want 6 lines", lines)
	}
	cur := 0
	for i, line := range lines {
		next, ok := advanceCursor(text, cur, line)
		if !ok {
			t.Fatalf("line %d: cursor lost sync", i+1)
		}
		if text[cur:cur+len(line)] != line {
			t.Fatalf("line %d: cursor slice %q != line %q", i+1, text[cur:cur+len(line)], line)
		}
		for _, pos := range selectorOccurrences(line, "os.Remove") {
			if !strings.HasPrefix(text[cur+pos:], "os.Remove") {
				t.Errorf("line %d: absolute offset %d does not point at os.Remove", i+1, cur+pos)
			}
		}
		cur = next
	}
	if cur != len(text) {
		t.Errorf("cursor ended at %d, want %d", cur, len(text))
	}

	offsets := absoluteOccurrences(t, text, "os.Remove")
	if len(offsets) != 3 {
		t.Fatalf("absolute occurrences = %v, want 3", offsets)
	}
	ext := extractExtent(text, offsets[0]+len("os.Remove"))
	if ext.kind != extentBalanced || !reflect.DeepEqual(ext.segments, []string{"a", "\fos.Remove(b)\u2028"}) {
		t.Errorf("first extent = %v %q, want balanced across the separators", ext.kind, ext.segments)
	}

	got := scanText("x.go", text)
	want := []string{
		"x.go:1: write primitive 'os.Remove' found",
		"x.go:2: write primitive 'os.Remove' found",
		"x.go:5: write primitive 'os.Remove' found",
		"x.go:6: write primitive 'os.Rename' found",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanText = %q, want %q", got, want)
	}
}

// TestCursor_InvalidUTF8FailsClosed pins the cursor's fail-closed fallback:
// pysem.SplitLines replaces invalid UTF-8 with U+FFFD, so the cursor slice no
// longer equals the line, the allowance is disabled for the rest of the text,
// and every occurrence is still reported. Both lines hold an allowed-shape
// CreateFile; a lost cursor resets to offset 0, where the first line's
// allowed extent sits, so only a disabled allowance reports both lines.
func TestCursor_InvalidUTF8FailsClosed(t *testing.T) {
	call := "syscall.CreateFile(" + allowedCreateFileArgs + ")"
	text := call + "\xff\n" + call
	lines := pysem.SplitLines(text)
	if _, ok := advanceCursor(text, 0, lines[0]); ok {
		t.Fatal("cursor accepted a line that differs from the raw text slice")
	}
	got := scanText("x.go", text)
	want := []string{
		"x.go:1: write primitive 'syscall.CreateFile' found",
		"x.go:2: write primitive 'syscall.CreateFile' found",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanText = %q, want %q", got, want)
	}
}

// TestCursor_LastLineBoundsCheck pins that the boundary decode is
// bounds-checked on the last line, with and without a trailing boundary.
func TestCursor_LastLineBoundsCheck(t *testing.T) {
	for _, text := range []string{"a\nb", "a\nb\n", "a\r\nb\r\n", "a\u2028"} {
		cur := 0
		for _, line := range pysem.SplitLines(text) {
			next, ok := advanceCursor(text, cur, line)
			if !ok {
				t.Fatalf("%q: cursor lost sync at %d", text, cur)
			}
			cur = next
		}
		if cur != len(text) {
			t.Errorf("%q: cursor ended at %d, want %d", text, cur, len(text))
		}
	}
	if _, ok := advanceCursor("ab", 1, "bc"); ok {
		t.Error("cursor accepted a line that runs past the end of the text")
	}
}
