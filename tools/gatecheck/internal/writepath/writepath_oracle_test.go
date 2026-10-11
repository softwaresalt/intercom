package writepath

// Frozen differential oracle for write-path detection parity (034.010-T,
// plan Unit D, D-T1a). This file pins the pre-Unit-D masked-text scanner
// (verbatim copies of the original scanText and findSelector over the
// original 20 selectors) and asserts that production findings restricted to
// those 20 selectors are identical in text, line and order.
//
// LIFECYCLE: the legacy side and its expectations are frozen. Two edits are
// authorised. First, the single AC-E2.6 adaptation in Unit E (production
// side, input list, and moving unparseable-input expectations to stricter
// fail-closed assertions). Second, the B83F53BB listing-format adaptation
// (D-BW-4, staging PR for Gatecheck Batch B1): the tracked-tree test decodes
// the runner's listing through decodeLsFilesListing, the runner-format
// adapter in writepath_test.go, instead of pysem.GitText and
// pysem.SplitLines. Any other edit is an H-3 stop.

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// oracleFrozenSelectors is the frozen copy of the original 20 selectors.
var oracleFrozenSelectors = []string{
	"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
	"os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
	"os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
	"os.CreateTemp", "os.MkdirTemp", "os.Link", "os.Chown", "os.Lchown",
	"os.Chtimes",
}

// oracleLegacyFindSelector is a verbatim copy of the pre-Unit-D findSelector.
func oracleLegacyFindSelector(line, sel string) bool {
	start := 0
	for start <= len(line) {
		idx := strings.Index(line[start:], sel)
		if idx < 0 {
			return false
		}
		pos := start + idx
		end := pos + len(sel)
		if !pysem.PrecededByWordOrDot(line, pos) && !pysem.FollowedByWord(line, end) {
			return true
		}
		start = pos + 1
	}
	return false
}

// oracleLegacyScanText is a verbatim copy of the pre-Unit-D scanText, bound
// to the frozen selector list and the frozen findSelector copy.
func oracleLegacyScanText(relPath, maskedText string) []string {
	var findings []string
	for i, line := range pysem.SplitLines(maskedText) {
		lineNo := i + 1
		for _, sel := range oracleFrozenSelectors {
			if oracleLegacyFindSelector(line, sel) {
				findings = append(findings, fmt.Sprintf("%s:%d: write primitive %s found", relPath, lineNo, pysem.Repr(sel)))
			}
		}
	}
	return findings
}

// oracleLegacyScanFile mirrors scanFile's read and mask steps exactly.
func oracleLegacyScanFile(root, relPath string) ([]string, error) {
	text, err := pysem.ReadText(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return nil, err
	}
	return oracleLegacyScanText(relPath, gomask.MaskGoNonCode(text)), nil
}

// oracleRestrictToFrozen keeps only the findings whose selector token, parsed
// from the right, is exactly one of the frozen 20 selectors (an exact set
// lookup, never a substring test: 'io.CopyN' must not count as 'io.Copy').
// A finding whose shape cannot be parsed fails the test (fail-closed).
func oracleRestrictToFrozen(t *testing.T, findings []string) []string {
	t.Helper()
	frozen := make(map[string]bool, len(oracleFrozenSelectors))
	for _, sel := range oracleFrozenSelectors {
		frozen[pysem.Repr(sel)] = true
	}
	const marker = "write primitive "
	var kept []string
	for _, f := range findings {
		body, ok := strings.CutSuffix(f, " found")
		i := strings.LastIndex(body, marker)
		if !ok || i < 0 {
			t.Fatalf("oracle: unparseable production finding %q", f)
		}
		if frozen[body[i+len(marker):]] {
			kept = append(kept, f)
		}
	}
	return kept
}

func oracleAssertEqual(t *testing.T, label string, legacy, production []string) {
	t.Helper()
	restricted := oracleRestrictToFrozen(t, production)
	if len(restricted) != len(legacy) {
		t.Fatalf("%s: restricted production findings=%q, legacy oracle=%q", label, restricted, legacy)
	}
	for i := range legacy {
		if restricted[i] != legacy[i] {
			t.Fatalf("%s: finding[%d]=%q, legacy oracle=%q", label, i, restricted[i], legacy[i])
		}
	}
}

// oracleCompareFile runs both sides over one repo-relative file. If either
// side fails to read or decode the file, both sides must fail.
func oracleCompareFile(t *testing.T, root, relPath string) {
	t.Helper()
	legacy, legacyErr := oracleLegacyScanFile(root, relPath)
	production, prodErr := scanFile(root, relPath)
	if legacyErr != nil || prodErr != nil {
		if legacyErr == nil || prodErr == nil {
			t.Fatalf("%s: error parity broken: legacy err=%v, production err=%v", relPath, legacyErr, prodErr)
		}
		return
	}
	oracleAssertEqual(t, relPath, legacy, production)
}

func TestOracle_FixtureCorpus_Parity(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join("scripts", "testdata", "writepath")
	fixtures := []string{
		"accept-clean.go",
		"accept-mentions-in-comment.go",
		"accept-non-tag-raw-string-selector.go",
		"accept-syscall-createfile-metadata.go",
		"reject-createtemp.go",
		"reject-io-copyn-copybuffer.go",
		"reject-link.go",
		"reject-os-chown.go",
		"reject-os-chtimes.go",
		"reject-os-lchown.go",
		"reject-os-mkdirtemp.go",
		"reject-os-openroot.go",
		"reject-os-root-type.go",
		"reject-struct-tag-selector.go",
		"reject-syscall-createfile-evasion.go",
		"reject-syscall-createfile-write.go",
		"reject-syscall-write.go",
		"reject-tag-shaped-raw-string-expr.go",
		"reject-writefile.go",
	}
	for _, name := range fixtures {
		rel := filepath.ToSlash(filepath.Join(fixtureDir, name))
		t.Run(name, func(t *testing.T) {
			oracleCompareFile(t, root, rel)
		})
	}
}

func TestOracle_FilebasedGoldenInputs_Parity(t *testing.T) {
	g := loadWritepathGolden(t)
	if len(g.Filebased) == 0 {
		t.Fatal("oracle: golden has no filebased inputs")
	}
	for _, row := range g.Filebased {
		t.Run(row.Name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, row.Name), rawBytesToString(row.RawBytes), 0o644); err != nil {
				t.Fatalf("materialize input: %v", err)
			}
			if row.DecodeError {
				_, legacyErr := oracleLegacyScanFile(dir, row.Name)
				_, prodErr := scanFile(dir, row.Name)
				if !errors.Is(legacyErr, pysem.ErrInvalidUTF8) || !errors.Is(prodErr, pysem.ErrInvalidUTF8) {
					t.Fatalf("invalid UTF-8 must fail both sides: legacy err=%v, production err=%v", legacyErr, prodErr)
				}
				return
			}
			oracleCompareFile(t, dir, row.Name)
		})
	}
}

func TestOracle_TrackedProductionTree_Parity(t *testing.T) {
	root := repoRoot(t)
	out, err := DefaultGitRunner(root)
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	var rels []string
	for _, p := range decodeLsFilesListing(t, out) {
		if shouldScan(p) {
			rels = append(rels, p)
		}
	}
	if len(rels) == 0 {
		t.Fatal("oracle: no tracked internal/** or cmd/** Go files")
	}
	for _, rel := range rels {
		t.Run(rel, func(t *testing.T) {
			oracleCompareFile(t, root, rel)
		})
	}
}

// TestOracle_HandBuiltBoundaryInputs_Parity pins the production boundary to
// parsed Go source while preserving the frozen masked-text oracle for valid
// inputs. Go rejects \f, \v and U+2028 in code position.
func TestOracle_HandBuiltBoundaryInputs_Parity(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		parseable bool
		want      []string
	}{
		{
			name: "form-feed-in-code.go",
			text: "package p\nfunc f() { _ = 1\f; }\n",
		},
		{
			name: "vertical-tab-in-code.go",
			text: "package p\nfunc f() { _ = 1\v; }\n",
		},
		{
			name: "unicode-line-separator-in-code.go",
			text: "package p\nfunc f() { _ = 1\u2028; }\n",
		},
		{
			name:      "comment-and-string-boundaries.go",
			text:      "package p\n\nimport \"os\"\n\n// a\fb\vc\u2028d\nvar s = \"e\ff\vg\u2028h\"\n\nfunc f() error { return os.Remove(s) }\n",
			parseable: true,
			want: []string{
				"comment-and-string-boundaries.go:8: write primitive 'os.Remove' found",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, parseErr := parser.ParseFile(token.NewFileSet(), tc.name, tc.text, parser.ParseComments|parser.AllErrors)
			if !tc.parseable {
				if parseErr == nil {
					t.Fatal("code-position boundary unexpectedly parses as Go")
				}
				findings, err := scanSource(tc.name, tc.text)
				if err == nil || len(findings) != 0 {
					t.Fatalf("scanSource must fail closed on invalid Go, findings=%q err=%v", findings, err)
				}
				return
			}
			if parseErr != nil {
				t.Fatalf("hand-built input must be a complete Go file: %v", parseErr)
			}
			masked := gomask.MaskGoNonCode(tc.text)
			legacy := oracleLegacyScanText(tc.name, masked)
			if strings.Join(legacy, "\n") != strings.Join(tc.want, "\n") {
				t.Fatalf("legacy oracle findings=%q, want %q", legacy, tc.want)
			}
			production, err := scanSource(tc.name, tc.text)
			if err != nil {
				t.Fatalf("scanSource(%q): %v", tc.name, err)
			}
			oracleAssertEqual(t, tc.name, legacy, production)
		})
	}
}
