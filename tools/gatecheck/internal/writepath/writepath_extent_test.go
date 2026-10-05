package writepath

import (
	"go/parser"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

func wholeWritepathFile(t *testing.T, body string) string {
	return wholeWritepathFileWithImports(t, `"os"`, body)
}

func wholeWritepathFileWithImports(t *testing.T, imports, body string) string {
	t.Helper()
	source := "package p\nimport " + imports + "\nfunc f() {\n" + body + "\n}\n"
	assertWrapperSelectorQualifiersImported(t, t.Name(), source)
	return source
}

// TestScanSource_ClassificationsKeepFindingText ports the parseable
// AC-D1.2 classifications to whole-file AST scans. A selector reference and
// nested call remain findings, with the established finding text.
func TestScanSource_ClassificationsKeepFindingText(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"selector reference", "f := os.Remove"},
		{"call", "os.Remove(a)"},
		{"nested call", "os.Remove(inner(a, b))"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := wholeWritepathFile(t, tc.body)
			got, err := scanSource("x.go", source)
			if err != nil {
				t.Fatalf("scanSource: %v", err)
			}
			want := []string{"x.go:4: write primitive 'os.Remove' found"}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("scanSource findings = %q, want %q", got, want)
			}
		})
	}
}

// TestScanSource_OneFindingPerSelectorPerLine preserves the R4-3 y.go source
// literal while checking the unchanged one-finding-per-line contract through
// the AST scanner and a complete Go file wrapper.
func TestScanSource_OneFindingPerSelectorPerLine(t *testing.T) {
	body := "os.Remove(a); os.Remove(b); os.Rename(c, d)\nxos.Remove(e)"
	source := wholeWritepathFileWithImports(t, `("os"; xos "errors")`, body)
	got, err := scanSource("y.go", source)
	if err != nil {
		t.Fatalf("scanSource: %v", err)
	}
	want := []string{
		"y.go:4: write primitive 'os.Remove' found",
		"y.go:4: write primitive 'os.Rename' found",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanSource findings = %q, want %q", got, want)
	}
}

// TestScanSource_SyntaxErrorFailsClosed ports the old undecidable extents to
// G-2 parser failures. Go-code separators and automatic semicolon insertion
// cannot turn malformed source into a partial scan.
func TestScanSource_SyntaxErrorFailsClosed(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"unbalanced arguments", "os.Remove(a, (b"},
		{"mismatched bracket", "os.Remove(a, [b)"},
		{"mismatched brace", "os.Remove(a}"},
		{"newline inserts semicolon", "os.Remove \t\n(a, b)"},
		{"form feed in code", "os.Remove\f(a)"},
		{"line separator in code", "os.Remove(a,\u2028 os.Remove(b))"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := wholeWritepathFile(t, tc.body)
			if _, err := parser.ParseFile(token.NewFileSet(), "syntax-error.go", source, parser.ParseComments|parser.AllErrors); err == nil {
				t.Fatal("parser accepted a row required to fail closed")
			}
			got, err := scanSource("syntax-error.go", source)
			if err == nil || len(got) != 0 {
				t.Fatalf("scanSource = %q, %v; want no findings and a parse error", got, err)
			}
		})
	}
}

// TestScanSource_MultiLineCallAcrossSeparators pins AST call classification
// across ordinary Go newlines while preserving raw-string selector hits and
// SplitLines numbering for U+2028 and form-feed characters in a struct tag.
func TestScanSource_MultiLineCallAcrossSeparators(t *testing.T) {
	for _, separator := range []struct {
		name  string
		value string
	}{
		{"line separator", "\u2028"},
		{"form feed", "\f"},
	} {
		t.Run(separator.name, func(t *testing.T) {
			source := "package p\nimport \"os\"\ntype S struct { F string `json:\"tag" + separator.value + "os.Remove\"` }\n" +
				"func f() {\n_ = os.Remove(\n\"call\",\n)\n}\n"
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "separators.go", source, parser.ParseComments|parser.AllErrors)
			if err != nil {
				t.Fatalf("discriminating fixture must parse: %v", err)
			}
			assertWrapperSelectorQualifiersImported(t, separator.name, source)
			tokenFile := fset.File(file.Pos())
			if tokenFile == nil {
				t.Fatal("parser did not associate a token.File")
			}

			masked := gomask.MaskGoNonCode(source)
			tagOffset := strings.Index(source, "os.Remove")
			callOffset := strings.LastIndex(source, "os.Remove")
			if tagOffset < 0 || callOffset <= tagOffset {
				t.Fatal("fixture must contain a tag hit and a later call")
			}
			tagLine := len(pysem.SplitLines(masked[:tagOffset] + "x"))
			callLine := len(pysem.SplitLines(masked[:callOffset] + "x"))
			if tokenFile.Line(tokenFile.Pos(tagOffset)) == tagLine ||
				tokenFile.Line(tokenFile.Pos(callOffset)) == callLine {
				t.Fatal("raw-tag fixture is not discriminating: SplitLines and go/token lines must differ")
			}

			got, err := scanSource("separators.go", source)
			if err != nil {
				t.Fatalf("scanSource: %v", err)
			}
			want := []string{
				"separators.go:" + strconv.Itoa(tagLine) + ": write primitive 'os.Remove' found",
				"separators.go:" + strconv.Itoa(callLine) + ": write primitive 'os.Remove' found",
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("scanSource findings = %q, want %q", got, want)
			}
		})
	}
}

// TestSelectorOccurrences_EnumeratesEvery pins full raw-string boundary
// enumeration, which preserves selector hits inside raw literals.
func TestSelectorOccurrences_EnumeratesEvery(t *testing.T) {
	line := "os.Remove(a) xos.Remove(b) os.Remove(c) os.Removed(d) .os.Remove(e) os.Remove"
	got := selectorOccurrences(line, "os.Remove")
	want := []int{0, 27, 68}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("occurrences = %v, want %v", got, want)
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

// TestHarness_049003_ParseableMultilineRawTagPinsSplitLines is the
// non-vacuous R4-4 line-mapping case. A visible raw-string struct tag carries
// U+2028 before both the tag selector and a later call, so masked-text
// SplitLines numbering differs from go/token's physical line positions.
func TestHarness_049003_ParseableMultilineRawTagPinsSplitLines(t *testing.T) {
	requireWritepathHarnessTask(t)
	const relPath = "raw-tag-line-map.go"
	source := "package p\nimport \"os\"\ntype S struct { F string `json:\"tag\u2028os.Remove\"` }\nfunc f() {\n_ = os.Remove(\"call\")\n}\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, relPath, source, parser.ParseComments|parser.AllErrors)
	if err != nil {
		t.Fatalf("discriminating fixture must parse: %v", err)
	}
	assertWrapperSelectorQualifiersImported(t, t.Name(), source)
	tokenFile := fset.File(file.Pos())
	if tokenFile == nil {
		t.Fatal("parser did not associate a token.File")
	}
	masked := gomask.MaskGoNonCode(source)
	tagOffset := strings.Index(source, "os.Remove")
	callOffset := strings.LastIndex(source, "os.Remove")
	if tagOffset < 0 || callOffset <= tagOffset {
		t.Fatal("fixture must contain the tag hit and a later call")
	}
	tagSplitLine := len(pysem.SplitLines(masked[:tagOffset] + "x"))
	callSplitLine := len(pysem.SplitLines(masked[:callOffset] + "x"))
	if tokenFile.Line(tokenFile.Pos(tagOffset)) == tagSplitLine ||
		tokenFile.Line(tokenFile.Pos(callOffset)) == callSplitLine {
		t.Fatal("raw-tag fixture is not discriminating: SplitLines and go/token lines must differ")
	}
	got, err := scanSource(relPath, source)
	if err != nil {
		t.Fatalf("scanSource: %v", err)
	}
	want := []string{
		relPath + ":" + strconv.Itoa(tagSplitLine) + ": write primitive 'os.Remove' found",
		relPath + ":" + strconv.Itoa(callSplitLine) + ": write primitive 'os.Remove' found",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("scanSource findings = %q, want %q", got, want)
	}
}
