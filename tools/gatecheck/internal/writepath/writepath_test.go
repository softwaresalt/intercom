package writepath

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// repoRoot returns the repository top-level directory from this test
// package's own path (tools/gatecheck/internal/writepath), so tests can
// exercise the real fixture corpus and (for the self-test/self-test-
// integrity golden captures) the real tracked internal/**, cmd/** tree
// without depending on the working directory `go test` happens to use.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, "..", "..", "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return root
}

func requireWritepathHarnessTask(t *testing.T) {
	t.Helper()
	name := strings.TrimPrefix(t.Name(), "TestHarness_")
	taskNumber := strings.SplitN(name, "_", 2)[0]
	if len(taskNumber) != 6 {
		t.Fatalf("invalid harness test name %q", t.Name())
	}
	task := taskNumber[:3] + "." + taskNumber[3:] + "-T"
	if os.Getenv("WRITEPATH_HARNESS_TASK") != task {
		t.Skipf("set WRITEPATH_HARNESS_TASK=%s to run this task's pre-implementation harness", task)
	}
}

func scanSourceForHarness(t *testing.T, relPath, source string) ([]string, error) {
	t.Helper()
	defer func() {
		if marker := recover(); marker != nil {
			t.Fatalf("%v", marker)
		}
	}()
	return scanSource(relPath, source)
}

func assertFindingsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("findings = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("finding[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

type writepathGolden struct {
	Selectors       []string             `json:"selectors"`
	FixtureFindings []fixtureFindingsRow `json:"fixture_findings"`
	StreamCaptures  map[string]streamRow `json:"stream_captures"`
	Filebased       []filebasedRow       `json:"filebased"`
}

type fixtureFindingsRow struct {
	Name           string   `json:"name"`
	ExpectRejected bool     `json:"expect_rejected"`
	Findings       []string `json:"findings"`
}

type streamRow struct {
	Cmd      []string `json:"cmd"`
	Cwd      string   `json:"cwd"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	ExitCode int      `json:"exit_code"`
}

type filebasedRow struct {
	Name           string   `json:"name"`
	RawBytes       []int    `json:"raw_bytes"`
	DecodeError    bool     `json:"decode_error"`
	TranslatedText string   `json:"translated_text"`
	Findings       []string `json:"findings"`
	ErrorRepr      string   `json:"error_repr"`
}

func loadWritepathGolden(t *testing.T) writepathGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "writepath_golden.json"))
	if err != nil {
		t.Fatalf("read writepath_golden.json: %v", err)
	}
	var g writepathGolden
	if err := json.Unmarshal(data, &g); err != nil {
		t.Fatalf("unmarshal writepath_golden.json: %v", err)
	}
	return g
}

func rawBytesToString(vals []int) []byte {
	b := make([]byte, len(vals))
	for i, v := range vals {
		b[i] = byte(v)
	}
	return b
}

func TestSelectors_MatchesGolden(t *testing.T) {
	g := loadWritepathGolden(t)
	if len(Selectors) != len(g.Selectors) {
		t.Fatalf("Selectors has %d entries, golden has %d", len(Selectors), len(g.Selectors))
	}
	for i, sel := range g.Selectors {
		if Selectors[i] != sel {
			t.Fatalf("Selectors[%d] = %q, want %q", i, Selectors[i], sel)
		}
	}
}

func TestScanFile_FixtureFindings_MatchGolden(t *testing.T) {
	root := repoRoot(t)
	g := loadWritepathGolden(t)
	for _, row := range g.FixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			relPath := "scripts/testdata/writepath/" + row.Name
			findings, err := scanFile(root, relPath)
			if err != nil {
				t.Fatalf("scanFile(%q): %v", relPath, err)
			}
			rejected := len(findings) > 0
			if rejected != row.ExpectRejected {
				t.Fatalf("scanFile(%q) rejected=%v, want %v (findings=%v)", relPath, rejected, row.ExpectRejected, findings)
			}
			if len(findings) != len(row.Findings) {
				t.Fatalf("scanFile(%q) findings=%v, want %v", relPath, findings, row.Findings)
			}
			for i := range findings {
				if findings[i] != row.Findings[i] {
					t.Fatalf("scanFile(%q) findings[%d]=%q, want %q", relPath, i, findings[i], row.Findings[i])
				}
			}
		})
	}
}

func TestFilebased_ReadTextAndScan_MatchGolden(t *testing.T) {
	g := loadWritepathGolden(t)
	for _, row := range g.Filebased {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, row.Name)
			if err := os.WriteFile(path, rawBytesToString(row.RawBytes), 0o644); err != nil {
				t.Fatalf("materialize fixture: %v", err)
			}

			got, err := pysem.ReadText(path)
			if row.DecodeError {
				if !errors.Is(err, pysem.ErrInvalidUTF8) {
					t.Fatalf("ReadText(%s) error = %v, want ErrInvalidUTF8", row.Name, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadText(%s): %v", row.Name, err)
			}
			if got != row.TranslatedText {
				t.Fatalf("ReadText(%s) = %q, want %q", row.Name, got, row.TranslatedText)
			}

			findings, err := scanSource(row.Name, got)
			if err != nil {
				t.Fatalf("scanSource(%s): %v", row.Name, err)
			}
			if len(findings) != len(row.Findings) {
				t.Fatalf("scanSource(%s) findings=%v, want %v", row.Name, findings, row.Findings)
			}
			for i := range findings {
				if findings[i] != row.Findings[i] {
					t.Fatalf("scanSource(%s) findings[%d]=%q, want %q", row.Name, i, findings[i], row.Findings[i])
				}
			}
		})
	}
}

func TestFindSelector_LookaroundTable(t *testing.T) {
	cases := []struct {
		name string
		line string
		sel  string
		want bool
	}{
		{"exact hit", `x := os.WriteFile("a", nil, 0o644)`, "os.WriteFile", true},
		{"preceded by word char", `xos.WriteFile()`, "os.WriteFile", false},
		{"preceded by dot", `.os.WriteFile()`, "os.WriteFile", false},
		{"preceded by non-ASCII letter", `caféos.WriteFile()`, "os.WriteFile", false},
		{"followed by word rune", `os.WriteFileX()`, "os.WriteFile", false},
		{"followed by non-word (paren)", `os.WriteFile(`, "os.WriteFile", true},
		{"no occurrence", `nothing to see here`, "os.WriteFile", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findSelector(tc.line, tc.sel)
			if got != tc.want {
				t.Fatalf("findSelector(%q, %q) = %v, want %v", tc.line, tc.sel, got, tc.want)
			}
		})
	}
}

func TestShouldScan(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"internal/config/validate.go", true},
		{"cmd/gatecheck/main.go", true},
		{"internal/config/validate_test.go", false},
		{"internal/config/README.md", false},
		{"tools/gatecheck/internal/pysem/classes.go", false},
		{"pkg/other/file.go", false},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			if got := shouldScan(tc.path); got != tc.want {
				t.Fatalf("shouldScan(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestRun_SelfTest_MatchesGolden(t *testing.T) {
	root := repoRoot(t)
	g := loadWritepathGolden(t)
	row, ok := g.StreamCaptures["self_test"]
	if !ok {
		t.Fatal("golden missing stream_captures.self_test")
	}

	var stdout, stderr strings.Builder
	code := Run("--self-test", root, DefaultGitRunner, &stdout, &stderr)

	if code != row.ExitCode {
		t.Errorf("exit code = %d, want %d", code, row.ExitCode)
	}
	if stdout.String() != row.Stdout {
		t.Errorf("stdout = %q, want %q", stdout.String(), row.Stdout)
	}
	if stderr.String() != row.Stderr {
		t.Errorf("stderr = %q, want %q", stderr.String(), row.Stderr)
	}
}

func TestRun_SelfTestIntegrity_MatchesGolden(t *testing.T) {
	root := repoRoot(t)
	g := loadWritepathGolden(t)
	row, ok := g.StreamCaptures["self_test_integrity"]
	if !ok {
		t.Fatal("golden missing stream_captures.self_test_integrity")
	}

	var stdout, stderr strings.Builder
	code := Run("--self-test-integrity", root, DefaultGitRunner, &stdout, &stderr)

	if code != row.ExitCode {
		t.Errorf("exit code = %d, want %d", code, row.ExitCode)
	}
	if stdout.String() != row.Stdout {
		t.Errorf("stdout = %q, want %q", stdout.String(), row.Stdout)
	}
	if stderr.String() != row.Stderr {
		t.Errorf("stderr = %q, want %q", stderr.String(), row.Stderr)
	}
}

func TestRun_UnknownMode_MatchesGolden(t *testing.T) {
	root := repoRoot(t)
	g := loadWritepathGolden(t)
	row, ok := g.StreamCaptures["unknown_mode"]
	if !ok {
		t.Fatal("golden missing stream_captures.unknown_mode")
	}

	var stdout, stderr strings.Builder
	code := Run("--bogus", root, DefaultGitRunner, &stdout, &stderr)

	if code != row.ExitCode {
		t.Errorf("exit code = %d, want %d", code, row.ExitCode)
	}
	if stdout.String() != row.Stdout {
		t.Errorf("stdout = %q, want %q", stdout.String(), row.Stdout)
	}
	if stderr.String() != row.Stderr {
		t.Errorf("stderr = %q, want %q", stderr.String(), row.Stderr)
	}
}

func TestRun_RepoMode_CurrentTreeIsClean(t *testing.T) {
	root := repoRoot(t)
	var stdout, stderr strings.Builder
	code := Run("", root, DefaultGitRunner, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(\"\", ...) code = %d, stderr = %q, want 0 (clean tree)", code, stderr.String())
	}
	if stdout.String() != "" || stderr.String() != "" {
		t.Fatalf("Run(\"\", ...) on a clean tree should be silent; stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// --- Fail-closed ACs ---

func TestRun_GitError_FailsClosed(t *testing.T) {
	root := repoRoot(t)
	failingGit := func(string) ([]byte, error) {
		return nil, errors.New("exit status 128")
	}
	var stdout, stderr strings.Builder
	code := Run("", root, failingGit, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr.String(), "::error::") {
		t.Fatalf("stderr = %q, want ::error:: prefix", stderr.String())
	}
}

func TestRun_ReadError_FailsClosed(t *testing.T) {
	root := repoRoot(t)
	missingFileGit := func(string) ([]byte, error) {
		return []byte("internal/does-not-exist-anywhere.go\n"), nil
	}
	var stdout, stderr strings.Builder
	code := Run("", root, missingFileGit, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr.String(), "::error::") {
		t.Fatalf("stderr = %q, want ::error:: prefix", stderr.String())
	}
}

func TestRun_InvalidUTF8_FailsClosed(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	badPath := filepath.Join(root, "internal", "bad.go")
	if err := os.WriteFile(badPath, []byte("package bad\n// caf\xff broken\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	git := func(string) ([]byte, error) {
		return []byte("internal/bad.go\n"), nil
	}
	var stdout, stderr strings.Builder
	code := Run("", root, git, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr.String(), "::error::") {
		t.Fatalf("stderr = %q, want ::error:: prefix", stderr.String())
	}
}

func TestRun_EmptySelection_FailsClosed_ED7(t *testing.T) {
	root := t.TempDir()
	emptyGit := func(string) ([]byte, error) {
		return []byte(""), nil
	}
	var stdout, stderr strings.Builder
	code := Run("", root, emptyGit, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "ED-7") {
		t.Fatalf("stderr = %q, want ED-7 mention", stderr.String())
	}
}

func TestRunFixtureSelfTest_MissingFixtureDir(t *testing.T) {
	root := t.TempDir()
	res := runFixtureSelfTest(root)
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1", res.Code)
	}
	if !strings.HasPrefix(res.Stderr, "fixture dir not found: ") {
		t.Fatalf("Stderr = %q, want %q prefix", res.Stderr, "fixture dir not found: ")
	}
	if strings.Contains(res.Stderr, "::error::") {
		t.Fatalf("Stderr = %q, must NOT carry an ::error:: prefix (Python message text is verbatim)", res.Stderr)
	}
}

func TestRunFixtureSelfTest_NoFixturesDiscovered(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "scripts", "testdata", "writepath"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	res := runFixtureSelfTest(root)
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1", res.Code)
	}
	if !strings.HasPrefix(res.Stderr, "no fixtures discovered under ") {
		t.Fatalf("Stderr = %q, want %q prefix", res.Stderr, "no fixtures discovered under ")
	}
}

// createFileCall builds a masked syscall.CreateFile call with the given
// argument list, as the scan path would see it.
func createFileCall(args string) string {
	return gomask.MaskGoNonCode("syscall.CreateFile(" + args + ")\n")
}

// allowedCreateFileArgs is the only argument shape D-2' admits.
const allowedCreateFileArgs = "p, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0"

// firstExtent extracts the extent following the first occurrence of sel in
// masked text.
func firstExtent(t *testing.T, masked, sel string) extent {
	t.Helper()
	pos := strings.Index(masked, sel)
	if pos < 0 {
		t.Fatalf("no %s in %q", sel, masked)
	}
	return extractExtent(masked, pos+len(sel))
}

// TestOccurrenceAllowed_ReparseWindowsCreateFile is AC-D2.1: the real
// reparse_windows.go extent, located by content, is allowed, keyed on the
// call's arguments and never on the file path.
func TestOccurrenceAllowed_ReparseWindowsCreateFile(t *testing.T) {
	const sel = "syscall.CreateFile"
	raw, err := pysem.ReadText(filepath.Join(repoRoot(t), "internal", "pathsafe", "reparse_windows.go"))
	if err != nil {
		t.Fatal(err)
	}
	masked := gomask.MaskGoNonCode(raw)
	offsets := absoluteOccurrences(t, masked, sel)
	if len(offsets) != 1 {
		t.Fatalf("masked occurrences = %d, want 1", len(offsets))
	}
	if !occurrenceAllowed(sel, extractExtent(masked, offsets[0]+len(sel))) {
		t.Error("the live metadata-only CreateFile call must be allowed")
	}
	if !occurrenceAllowed(sel, firstExtent(t, createFileCall(allowedCreateFileArgs), sel)) {
		t.Error("the single-line allowed shape must be allowed")
	}
	if occurrenceAllowed("os.Remove", firstExtent(t, createFileCall(allowedCreateFileArgs), sel)) {
		t.Error("the allowance must be keyed on the syscall.CreateFile selector only")
	}
}

// TestOccurrenceAllowed_RejectionTable is AC-D2.2: every non-conforming
// extent is rejected (fail closed, INV-2).
func TestOccurrenceAllowed_RejectionTable(t *testing.T) {
	const sel = "syscall.CreateFile"
	const (
		oe = "syscall.OPEN_EXISTING"
		bs = "syscall.FILE_FLAG_BACKUP_SEMANTICS"
	)
	callWith := func(access, disposition, flags string) string {
		return createFileCall("p, " + access + ", 0, nil, " + disposition + ", " + flags + ", 0")
	}
	cases := []struct {
		name   string
		masked string
	}{
		{"non-call reference", gomask.MaskGoNonCode("f := syscall.CreateFile\n")},
		{"unbalanced extent", gomask.MaskGoNonCode("syscall.CreateFile(" + allowedCreateFileArgs + "\n")},
		{"mismatched extent", gomask.MaskGoNonCode("syscall.CreateFile(" + allowedCreateFileArgs + "]\n")},
		{"six arguments", createFileCall("p, 0, 0, nil, " + oe + ", " + bs)},
		{"eight arguments", createFileCall(allowedCreateFileArgs + ", 0")},
		{"access 0x0", callWith("0x0", oe, bs)},
		{"access 00", callWith("00", oe, bs)},
		{"access (0)", callWith("(0)", oe, bs)},
		{"access uint32(0)", callWith("uint32(0)", oe, bs)},
		{"access named constant", callWith("noAccess", oe, bs)},
		{"access GENERIC_WRITE", callWith("syscall.GENERIC_WRITE", oe, bs)},
		{"CREATE_NEW", callWith("0", "syscall.CREATE_NEW", bs)},
		{"CREATE_ALWAYS", callWith("0", "syscall.CREATE_ALWAYS", bs)},
		{"OPEN_ALWAYS", callWith("0", "syscall.OPEN_ALWAYS", bs)},
		{"TRUNCATE_EXISTING", callWith("0", "syscall.TRUNCATE_EXISTING", bs)},
		{"numeric disposition", callWith("0", "3", bs)},
		{"DELETE_ON_CLOSE alone", callWith("0", oe, "syscall.FILE_FLAG_DELETE_ON_CLOSE")},
		{"DELETE_ON_CLOSE OR-ed", callWith("0", oe, bs+"|syscall.FILE_FLAG_DELETE_ON_CLOSE")},
		{"OPEN_REPARSE_POINT OR-ed", callWith("0", oe, bs+"|syscall.FILE_FLAG_OPEN_REPARSE_POINT")},
		{"brace composite", createFileCall("T{p, 0, a, b, c, d, e}.Args()")},
		{"call-valued argument", createFileCall("name(), 0, 0, nil, " + oe + ", " + bs + ", 0")},
		{"index expression", createFileCall("p[0], 0, 0, nil, " + oe + ", " + bs + ", 0")},
		{"tag-shaped raw string", gomask.MaskGoNonCode("syscall.CreateFile(p, 0, `json:\"x\"`, nil, " + oe + ", " + bs + ", 0)\n")},
		{"interpreted string argument", createFileCall("p, 0, \"rw\", nil, " + oe + ", " + bs + ", 0")},
		{"rune argument", createFileCall("p, 0, 'x', nil, " + oe + ", " + bs + ", 0")},
		{"leading empty segment", createFileCall(", 0, 0, nil, " + oe + ", " + bs + ", 0")},
		{"two trailing empty segments", createFileCall(allowedCreateFileArgs + ",,")},
		{"empty call", createFileCall("")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if occurrenceAllowed(sel, firstExtent(t, tc.masked, sel)) {
				t.Errorf("%q must be rejected", tc.masked)
			}
		})
	}
}

// TestLineReportsSelector_AllowedCallCannotHideWritingCall is AC-D2.3: an
// allowed call and a writing call on one line, in both orders, still yield
// the finding; the selector is passed as a parameter to exercise the
// per-selector evaluator directly.
func TestLineReportsSelector_AllowedCallCannotHideWritingCall(t *testing.T) {
	const sel = "syscall.CreateFile"
	allowed := "syscall.CreateFile(" + allowedCreateFileArgs + ")"
	writing := "syscall.CreateFile(p, syscall.GENERIC_WRITE, 0, nil, syscall.CREATE_ALWAYS, 0, 0)"
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"allowed alone", allowed, false},
		{"allowed then writing", allowed + "; " + writing, true},
		{"writing then allowed", writing + "; " + allowed, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			masked := gomask.MaskGoNonCode(tc.line)
			if got := lineReportsSelector(masked, masked, 0, true, sel); got != tc.want {
				t.Errorf("lineReportsSelector(%q) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
	masked := gomask.MaskGoNonCode(allowed)
	if !lineReportsSelector(masked, masked, 0, false, sel) {
		t.Error("with the allowance disabled the allowed call must still be reported (fail closed)")
	}
}

// TestHarness_049002_SyntaxErrorsFailClosed covers AC-E2.3 and G-2, including
// a parser error that returns a partial AST and code-position line separators
// that must fail closed rather than being treated as Python-style newlines.
func TestHarness_049002_SyntaxErrorsFailClosed(t *testing.T) {
	requireWritepathHarnessTask(t)
	cases := []struct {
		name        string
		source      string
		wantPartial bool
	}{
		{
			name:        "partial AST from missing closing brace",
			source:      "package p\nimport \"os\"\nfunc f() { _ = os.Remove(\"x\")\n",
			wantPartial: true,
		},
		{
			name:   "malformed declaration",
			source: "package p\nfunc f( {",
		},
		{
			name:   "form feed in code",
			source: "package p\nfunc f() { _ = 1\f; }\n",
		},
		{
			name:   "vertical tab in code",
			source: "package p\nfunc f() { _ = 1\v; }\n",
		},
		{
			name:   "unicode line separator in code",
			source: "package p\nfunc f() { _ = 1\u2028; }\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file, parseErr := parser.ParseFile(token.NewFileSet(), "invalid.go", tc.source, parser.ParseComments|parser.AllErrors)
			if parseErr == nil {
				t.Fatal("test input unexpectedly parses")
			}
			if tc.wantPartial && file == nil {
				t.Fatal("partial-AST test input did not produce a partial AST")
			}
			findings, err := scanSourceForHarness(t, "invalid.go", tc.source)
			if err == nil || len(findings) != 0 {
				t.Fatalf("not implemented: 049.002-T: parse error must return no findings and fail closed, got findings=%q err=%v", findings, err)
			}
		})
	}
}

// TestHarness_049002_LineDirectivePinsPosition covers the PositionFor(pos,
// false) obligation: a //line directive cannot move the finding off the
// physical source line.
func TestHarness_049002_LineDirectivePinsPosition(t *testing.T) {
	requireWritepathHarnessTask(t)
	const relPath = "line-map.go"
	source := "//line deceptive.go:900\npackage p\nimport \"os\"\nfunc f() { _ = os.Remove(\"x\") }\n"
	got, err := scanSourceForHarness(t, relPath, source)
	if err != nil {
		t.Fatalf("scanSource: %v", err)
	}
	assertFindingsEqual(t, got, []string{"line-map.go:4: write primitive 'os.Remove' found"})
}

// TestHarness_049002_GoldenParityThroughScanSource routes the existing
// fixture golden through the new unmasked-source boundary without editing
// the golden or widening the frozen oracle corpus.
func TestHarness_049002_GoldenParityThroughScanSource(t *testing.T) {
	requireWritepathHarnessTask(t)
	root := repoRoot(t)
	golden := loadWritepathGolden(t)
	for _, row := range golden.FixtureFindings {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			relPath := "scripts/testdata/writepath/" + row.Name
			source, err := pysem.ReadText(filepath.Join(root, filepath.FromSlash(relPath)))
			if err != nil {
				t.Fatalf("ReadText(%q): %v", relPath, err)
			}
			got, err := scanSourceForHarness(t, relPath, source)
			if err != nil {
				t.Fatalf("scanSource(%q): %v", relPath, err)
			}
			assertFindingsEqual(t, got, row.Findings)
		})
	}
}

// TestHarness_049002_FilebasedParityThroughScanSource keeps the existing
// BOM/CRLF/lone-CR verdict contract on the source boundary. Invalid UTF-8
// remains rejected by pysem.ReadText before scanSource is entered.
func TestHarness_049002_FilebasedParityThroughScanSource(t *testing.T) {
	requireWritepathHarnessTask(t)
	golden := loadWritepathGolden(t)
	for _, row := range golden.Filebased {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			if row.DecodeError {
				path := filepath.Join(t.TempDir(), row.Name)
				if err := os.WriteFile(path, rawBytesToString(row.RawBytes), 0o644); err != nil {
					t.Fatalf("materialize invalid UTF-8 input: %v", err)
				}
				if _, err := pysem.ReadText(path); !errors.Is(err, pysem.ErrInvalidUTF8) {
					t.Fatal("invalid UTF-8 filebased input did not retain the ReadText rejection")
				}
				return
			}
			got, err := scanSourceForHarness(t, row.Name, row.TranslatedText)
			if err != nil {
				t.Fatalf("scanSource(%q): %v", row.Name, err)
			}
			assertFindingsEqual(t, got, row.Findings)
		})
	}
}

// TestHarness_049002_RetiredArchPinReanchored verifies the R6-6 handoff:
// the scan-loop pin must consume scanSource, retain the canonical masker,
// and stop positively depending on scanText.
func TestHarness_049002_RetiredArchPinReanchored(t *testing.T) {
	requireWritepathHarnessTask(t)
	root := repoRoot(t)
	retiredPath := filepath.Join(root, "tools", "gatecheck", "internal", "retiredarch", "writepath_mask_test.go")
	retiredSource, err := os.ReadFile(retiredPath)
	if err != nil {
		t.Fatalf("read retiredarch pin: %v", err)
	}
	retired, err := parser.ParseFile(token.NewFileSet(), retiredPath, retiredSource, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse retiredarch pin: %v", err)
	}
	var flow *ast.FuncDecl
	for _, decl := range retired.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "writePathMaskFlow" {
			flow = fn
			break
		}
	}
	if flow == nil || astHasIdent(flow.Body, "scanText") {
		t.Fatal("not implemented: 049.002-T: retiredarch writePathMaskFlow still positively depends on scanText")
	}

	productionPath := filepath.Join(root, "tools", "gatecheck", "internal", "writepath", "writepath.go")
	productionSource, err := os.ReadFile(productionPath)
	if err != nil {
		t.Fatalf("read writepath source: %v", err)
	}
	production, err := parser.ParseFile(token.NewFileSet(), productionPath, productionSource, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse writepath source: %v", err)
	}
	functions := make(map[string]*ast.FuncDecl)
	for _, decl := range production.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			functions[fn.Name.Name] = fn
		}
	}
	if !astHasCallIdent(functions["scanFile"], "scanSource") ||
		!astHasSelectorCall(functions["scanSource"], "gomask", "MaskGoNonCode") ||
		!astHasSelectorCall(functions["scanSource"], "pysem", "SplitLines") {
		t.Fatal("not implemented: 049.002-T: scanSource is not wired through the canonical masker and SplitLines")
	}
	if !strings.Contains(string(retiredSource), "canonicalMaskerDef = \"internal/gomask/gomask.go:MaskGoNonCode\"") ||
		!strings.Contains(string(retiredSource), "TestMaskerDefinedExactlyOnce") {
		t.Fatal("not implemented: 049.002-T: retiredarch canonical-mask/single-definition protections are missing")
	}
}

func astHasIdent(node ast.Node, name string) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return !found
	})
	return found
}

func astHasCallIdent(node ast.Node, name string) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if ok && id.Name == name {
			found = true
		}
		return !found
	})
	return found
}

func astHasSelectorCall(node ast.Node, qualifier, name string) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != name {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if ok && id.Name == qualifier {
			found = true
		}
		return !found
	})
	return found
}

func harnessWritepathFile(imports, body string) string {
	return "package p\nimport " + imports + "\nfunc f() {\n" + body + "\n}\n"
}

func harnessFixtureSource(t *testing.T, name string) string {
	t.Helper()
	testName := strings.SplitN(t.Name(), "/", 2)[0]
	taskNumber := strings.SplitN(strings.TrimPrefix(testName, "TestHarness_"), "_", 2)[0]
	if len(taskNumber) != 6 {
		t.Fatalf("invalid harness test name %q", testName)
	}
	taskDir := taskNumber[:3] + "." + taskNumber[3:]
	path := filepath.Join(repoRoot(t), "scripts", "testdata", "writepath", "harness", taskDir, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read gated fixture %q: %v", path, err)
	}
	return string(data)
}

func harnessFinding(relPath, source string) string {
	const selector = "syscall.CreateFile"
	pos := strings.Index(source, selector)
	if pos < 0 {
		return ""
	}
	line := strings.Count(source[:pos], "\n") + 1
	return fmt.Sprintf("%s:%d: write primitive '%s' found", relPath, line, selector)
}

// TestHarness_049003_ASTCallPolicy carries the original D-2' rejection
// contract over whole, unmasked Go files. The name-bound lookalike case is
// the E-T3 red discriminator: only the syscall import path retains the
// metadata-only allowance.
func TestHarness_049003_ASTCallPolicy(t *testing.T) {
	requireWritepathHarnessTask(t)
	const (
		selector = "syscall.CreateFile"
		open     = "syscall.OPEN_EXISTING"
		flags    = "syscall.FILE_FLAG_BACKUP_SEMANTICS"
	)
	reject := func(args string) string { return selector + "(" + args + ")" }
	cases := []struct {
		name     string
		imports  string
		body     string
		wantNone bool
	}{
		{
			name:     "allowed exact syscall call",
			imports:  `"syscall"`,
			body:     reject(allowedCreateFileArgs),
			wantNone: true,
		},
		{
			name:     "allowed selector bare operand",
			imports:  `"syscall"`,
			body:     reject("pkg.path, 0, 0, nil, " + open + ", " + flags + ", 0"),
			wantNone: true,
		},
		{
			name:     "allowed unary bare operand",
			imports:  `"syscall"`,
			body:     reject("p, 0, -offset, nil, " + open + ", " + flags + ", 0"),
			wantNone: true,
		},
		{
			name:     "allowed star bare operand",
			imports:  `"syscall"`,
			body:     reject("p, 0, 0, *value, " + open + ", " + flags + ", 0"),
			wantNone: true,
		},
		{
			name:     "allowed binary bare operand",
			imports:  `"syscall"`,
			body:     reject("p, 0, left+right, nil, " + open + ", " + flags + ", 0"),
			wantNone: true,
		},
		{
			name:    "8 arguments with trailing string",
			imports: `"syscall"`,
			body:    reject(allowedCreateFileArgs + `, "x"`),
		},
		{
			name:    "variadic call",
			imports: `"syscall"`,
			body:    selector + "(args...)",
		},
		{
			name:    "non-call reference",
			imports: `"syscall"`,
			body:    "var f = " + selector,
		},
		{
			name:    "six arguments",
			imports: `"syscall"`,
			body:    reject("p, 0, 0, nil, " + open + ", " + flags),
		},
		{
			name:    "access 0x0",
			imports: `"syscall"`,
			body:    reject("p, 0x0, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "access 00",
			imports: `"syscall"`,
			body:    reject("p, 00, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "access parenthesized zero",
			imports: `"syscall"`,
			body:    reject("p, (0), 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "access conversion call",
			imports: `"syscall"`,
			body:    reject("p, uint32(0), 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "access named constant",
			imports: `"syscall"`,
			body:    reject("p, noAccess, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "access write constant",
			imports: `"syscall"`,
			body:    reject("p, syscall.GENERIC_WRITE, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "wrong disposition",
			imports: `"syscall"`,
			body:    reject("p, 0, 0, nil, syscall.CREATE_NEW, " + flags + ", 0"),
		},
		{
			name:    "numeric disposition",
			imports: `"syscall"`,
			body:    reject("p, 0, 0, nil, 3, " + flags + ", 0"),
		},
		{
			name:    "wrong flags",
			imports: `"syscall"`,
			body:    reject("p, 0, 0, nil, " + open + ", syscall.FILE_FLAG_DELETE_ON_CLOSE, 0"),
		},
		{
			name:    "additional flags",
			imports: `"syscall"`,
			body:    reject("p, 0, 0, nil, " + open + ", " + flags + "|syscall.FILE_FLAG_DELETE_ON_CLOSE, 0"),
		},
		{
			name:    "call-valued argument at Args[0]",
			imports: `"syscall"`,
			body:    reject("name(), 0, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "index argument at Args[0]",
			imports: `"syscall"`,
			body:    reject("p[0], 0, 0, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "string argument at Args[2]",
			imports: `"syscall"`,
			body:    reject("p, 0, \"rw\", nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "rune argument at Args[2]",
			imports: `"syscall"`,
			body:    reject("p, 0, 'x', nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "tag-shaped raw string at Args[2]",
			imports: `"syscall"`,
			body:    reject("p, 0, `json:\"x\"`, nil, " + open + ", " + flags + ", 0"),
		},
		{
			name:    "empty call",
			imports: `"syscall"`,
			body:    selector + "()",
		},
		{
			name:    "alias named syscall is not canonical syscall",
			imports: `syscall "example.invalid/other"`,
			body:    reject(allowedCreateFileArgs),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			relPath := "ast-predicate.go"
			source := harnessWritepathFile(tc.imports, tc.body)
			if _, err := parser.ParseFile(token.NewFileSet(), relPath, source, parser.ParseComments|parser.AllErrors); err != nil {
				t.Fatalf("parseable rejection/control case %q failed to parse: %v", tc.name, err)
			}
			got, err := scanSourceForHarness(t, relPath, source)
			if err != nil {
				t.Fatalf("scanSource: %v", err)
			}
			if tc.wantNone {
				assertFindingsEqual(t, got, nil)
				return
			}
			assertFindingsEqual(t, got, []string{harnessFinding(relPath, source)})
		})
	}
}

// TestHarness_049003_ParserErrorSetPinsUnmaskedInputs requires exactly the
// four known unparseable predicate rows (:443, :444, :467, :468); quoted and
// call-valued R6-1 inputs are checked separately as parseable rejections.
func TestHarness_049003_ParserErrorSetPinsUnmaskedInputs(t *testing.T) {
	requireWritepathHarnessTask(t)
	cases := []struct {
		name string
		body string
	}{
		{"unbalanced extent", "syscall.CreateFile(" + allowedCreateFileArgs},
		{"mismatched closer", "syscall.CreateFile(" + allowedCreateFileArgs + "])"},
		{"leading empty argument", "syscall.CreateFile(, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)"},
		{"two trailing empty arguments", "syscall.CreateFile(" + allowedCreateFileArgs + ",,)"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			relPath := "unparseable-predicate.go"
			source := harnessWritepathFile(`"syscall"`, tc.body)
			if _, err := parser.ParseFile(token.NewFileSet(), relPath, source, parser.ParseComments|parser.AllErrors); err == nil {
				t.Fatalf("probe case %q unexpectedly parses; expected the pinned R5-1 error set", tc.name)
			}
			findings, err := scanSourceForHarness(t, relPath, source)
			if err == nil || len(findings) != 0 {
				t.Fatalf("not implemented: 049.003-T: parser errors must fail closed, findings=%q err=%v", findings, err)
			}
		})
	}
}

// TestHarness_049003_R6_1PositionalRejections is mutation-sensitive: every
// generic argument position, especially Args[3] and Args[6], has parseable
// STRING/CHAR rejection cases; call-valued Args[2]/[3]/[6] are also pinned.
func TestHarness_049003_R6_1PositionalRejections(t *testing.T) {
	requireWritepathHarnessTask(t)
	parts := strings.Split(allowedCreateFileArgs, ",")
	cases := []struct {
		name     string
		index    int
		replaced string
	}{
		{"string Args[0]", 0, `"p"`},
		{"rune Args[0]", 0, `'p'`},
		{"call Args[2]", 2, `name()`},
		{"string Args[3]", 3, `"sa"`},
		{"rune Args[3]", 3, `'s'`},
		{"call Args[3]", 3, `name()`},
		{"string Args[6]", 6, `"t"`},
		{"rune Args[6]", 6, `'t'`},
		{"call Args[6]", 6, `name()`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string(nil), parts...)
			args[tc.index] = tc.replaced
			source := harnessWritepathFile(`"syscall"`, "syscall.CreateFile("+strings.Join(args, ", ")+")")
			file, err := parser.ParseFile(token.NewFileSet(), "r6-1.go", source, parser.ParseComments|parser.AllErrors)
			if err != nil || file == nil {
				t.Fatalf("R6-1 source for Args[%d] must parse, file=%v err=%v", tc.index, file != nil, err)
			}
			got, err := scanSourceForHarness(t, "r6-1.go", source)
			if err != nil {
				t.Fatalf("scanSource: %v", err)
			}
			assertFindingsEqual(t, got, []string{harnessFinding("r6-1.go", source)})
		})
	}
}

// TestHarness_049003_SameLineAllowanceCannotHideWrite pins AC-D2.3 through
// scanSource rather than the retired text-line evaluator.
func TestHarness_049003_SameLineAllowanceCannotHideWrite(t *testing.T) {
	requireWritepathHarnessTask(t)
	allowed := "syscall.CreateFile(" + allowedCreateFileArgs + ")"
	writing := "syscall.CreateFile(p, syscall.GENERIC_WRITE, 0, nil, syscall.CREATE_ALWAYS, 0, 0)"
	cases := []struct {
		name string
		body string
		want int
	}{
		{"allowed alone", allowed, 0},
		{"allowed then writing", allowed + "; " + writing, 1},
		{"writing then allowed", writing + "; " + allowed, 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			relPath := "same-line.go"
			source := harnessWritepathFile(`"syscall"`, tc.body)
			got, err := scanSourceForHarness(t, relPath, source)
			if err != nil {
				t.Fatalf("scanSource: %v", err)
			}
			if len(got) != tc.want {
				t.Fatalf("findings = %q, want %d for %q", got, tc.want, tc.name)
			}
		})
	}
}

// TestHarness_049004_AliasDotAndSplitFixtures covers all four E-T4 fixture
// scenarios. Fixtures stay in a nested harness directory until the green
// implementation phase promotes them into the flat production corpus and
// adds the matching golden/stream rows.
func TestHarness_049004_AliasDotAndSplitFixtures(t *testing.T) {
	requireWritepathHarnessTask(t)
	cases := []struct {
		name string
		want []string
	}{
		{
			name: "reject-alias-os-writefile.go",
			want: []string{"scripts/testdata/writepath/reject-alias-os-writefile.go:4: write primitive 'os.WriteFile' found"},
		},
		{
			name: "reject-dot-import-os.go",
			want: []string{"scripts/testdata/writepath/reject-dot-import-os.go:2: write primitive 'os.*' found"},
		},
		{
			name: "accept-alias-copilot.go",
			want: nil,
		},
		{
			name: "reject-split-selector.go",
			want: []string{
				"scripts/testdata/writepath/reject-split-selector.go:4: write primitive 'os.WriteFile' found",
				"scripts/testdata/writepath/reject-split-selector.go:6: write primitive 'os.WriteFile' found",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			source := harnessFixtureSource(t, tc.name)
			if _, err := parser.ParseFile(token.NewFileSet(), tc.name, source, parser.ParseComments|parser.AllErrors); err != nil {
				t.Fatalf("fixture %q must parse: %v", tc.name, err)
			}
			relPath := "scripts/testdata/writepath/" + tc.name
			got, err := scanSourceForHarness(t, relPath, source)
			if err != nil {
				t.Fatalf("scanSource(%q): %v", relPath, err)
			}
			assertFindingsEqual(t, got, tc.want)
		})
	}
}

// TestHarness_049004_GoldenRowsAreAdditive pins AC-E4.1/AC-E4.2 once the
// gated fixtures are promoted: the previous selector list is unchanged and
// every new fixture receives an additive finding row and PASS capture.
func TestHarness_049004_GoldenRowsAreAdditive(t *testing.T) {
	requireWritepathHarnessTask(t)
	golden := loadWritepathGolden(t)
	want := map[string][]string{
		"reject-alias-os-writefile.go": {
			"scripts/testdata/writepath/reject-alias-os-writefile.go:4: write primitive 'os.WriteFile' found",
		},
		"reject-dot-import-os.go": {
			"scripts/testdata/writepath/reject-dot-import-os.go:2: write primitive 'os.*' found",
		},
		"accept-alias-copilot.go": nil,
		"reject-split-selector.go": {
			"scripts/testdata/writepath/reject-split-selector.go:4: write primitive 'os.WriteFile' found",
			"scripts/testdata/writepath/reject-split-selector.go:6: write primitive 'os.WriteFile' found",
		},
	}
	for name, findings := range want {
		var row *fixtureFindingsRow
		for i := range golden.FixtureFindings {
			if golden.FixtureFindings[i].Name == name {
				row = &golden.FixtureFindings[i]
				break
			}
		}

		if row == nil {
			t.Fatalf("not implemented: 049.004-T: golden row %q is missing", name)
		}
		if row.ExpectRejected != (len(findings) > 0) {
			t.Fatalf("golden row %q rejected=%v, want %v", name, row.ExpectRejected, len(findings) > 0)
		}
		assertFindingsEqual(t, row.Findings, findings)
		for _, captureName := range []string{"self_test", "self_test_integrity"} {
			capture, ok := golden.StreamCaptures[captureName]
			if !ok || !strings.Contains(capture.Stdout, "PASS "+name+":") {
				t.Fatalf("not implemented: 049.004-T: %s lacks the additive PASS line for %s", captureName, name)
			}
		}
	}
	if len(golden.Selectors) != len(Selectors) {
		t.Fatalf("E-T4 must not change the selector count: golden=%d selectors=%d", len(golden.Selectors), len(Selectors))
	}
	for i := range golden.Selectors {
		if golden.Selectors[i] != Selectors[i] {
			t.Fatalf("E-T4 changed Selectors[%d] from the golden value", i)
		}
	}
}

// TestHarness_049005_NewRootReceiversTripFirstCaller verifies all three
// compiling intra-function bindings required by AC-E5.1/R4-8.
func TestHarness_049005_NewRootReceiversTripFirstCaller(t *testing.T) {
	requireWritepathHarnessTask(t)
	const fixtureName = "reject-resolve-first-caller.go"
	source := harnessFixtureSource(t, fixtureName)
	if _, err := parser.ParseFile(token.NewFileSet(), fixtureName, source, parser.ParseComments|parser.AllErrors); err != nil {
		t.Fatalf("NewRoot fixture must parse: %v", err)
	}
	for _, form := range []string{
		"root, err := pathsafe.NewRoot",
		"root, err = pathsafe.NewRoot",
		"var r, err = pathsafe.NewRoot",
	} {
		if !strings.Contains(source, form) {
			t.Fatalf("fixture is missing the required two-name binding form %q", form)
		}
	}
	if strings.Count(source, ".Resolve(") != 3 {
		t.Fatal("fixture must contain exactly three Root.Resolve callers")
	}
	relPath := "scripts/testdata/writepath/" + fixtureName
	got, err := scanSourceForHarness(t, relPath, source)
	if err != nil {
		t.Fatalf("scanSource(%q): %v", relPath, err)
	}
	if len(got) != 3 {
		t.Fatalf("not implemented: 049.005-T: got %d findings for the three NewRoot-bound Resolve callers: %q", len(got), got)
	}
	var lines []int
	for offset := strings.Index(source, ".Resolve("); offset >= 0; {
		lines = append(lines, strings.Count(source[:offset], "\n")+1)
		next := strings.Index(source[offset+len(".Resolve("):], ".Resolve(")
		if next < 0 {
			break
		}
		offset += len(".Resolve(") + next
	}
	for _, line := range lines {
		found := false
		for _, finding := range got {
			if strings.Contains(finding, fmt.Sprintf(":%d:", line)) && strings.Contains(finding, "Resolve") {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("not implemented: 049.005-T: no Root.Resolve finding at fixture line %d: %q", line, got)
		}
	}
}

// TestHarness_049005_UnrelatedResolveIsNotATrigger is the negative control:
// importing pathsafe does not turn an unrelated receiver's Resolve method
// into a finding.
func TestHarness_049005_UnrelatedResolveIsNotATrigger(t *testing.T) {
	requireWritepathHarnessTask(t)
	const fixtureName = "accept-unrelated-resolve.go"
	source := harnessFixtureSource(t, fixtureName)
	if _, err := parser.ParseFile(token.NewFileSet(), fixtureName, source, parser.ParseComments|parser.AllErrors); err != nil {
		t.Fatalf("unrelated receiver fixture must parse: %v", err)
	}
	relPath := "scripts/testdata/writepath/" + fixtureName
	got, err := scanSourceForHarness(t, relPath, source)
	if err != nil {
		t.Fatalf("scanSource(%q): %v", relPath, err)
	}
	assertFindingsEqual(t, got, nil)
}

// TestHarness_049005_TrackedTreeAndGoldenStayAdditive pins the clean tracked
// tree and requires additive golden rows for the positive and negative
// fixtures without changing the existing selector list.
func TestHarness_049005_TrackedTreeAndGoldenStayAdditive(t *testing.T) {
	requireWritepathHarnessTask(t)
	root := repoRoot(t)
	result := runRepoScan(root, DefaultGitRunner)
	if result.Code != 0 || result.Stderr != "" {
		t.Fatalf("not implemented: 049.005-T: tracked production tree must stay clean, code=%d stderr=%q", result.Code, result.Stderr)
	}
	golden := loadWritepathGolden(t)
	for _, fixture := range []string{"reject-resolve-first-caller.go", "accept-unrelated-resolve.go"} {
		var row *fixtureFindingsRow
		for i := range golden.FixtureFindings {
			if golden.FixtureFindings[i].Name == fixture {
				row = &golden.FixtureFindings[i]
				break
			}
		}
		if row == nil {
			t.Fatalf("not implemented: 049.005-T: golden row %q is missing", fixture)
		}
		if fixture == "reject-resolve-first-caller.go" && (!row.ExpectRejected || len(row.Findings) != 3) {
			t.Fatalf("not implemented: 049.005-T: positive golden row must retain three tripwire findings: %+v", *row)
		}
		if fixture == "accept-unrelated-resolve.go" && (row.ExpectRejected || len(row.Findings) != 0) {
			t.Fatalf("unrelated Resolve golden row must remain clean: %+v", *row)
		}
		for _, captureName := range []string{"self_test", "self_test_integrity"} {
			capture, ok := golden.StreamCaptures[captureName]
			if !ok || !strings.Contains(capture.Stdout, "PASS "+fixture+":") {
				t.Fatalf("not implemented: 049.005-T: %s lacks the additive PASS line for %s", captureName, fixture)
			}
		}
	}
	if len(golden.Selectors) != len(Selectors) {
		t.Fatalf("E-T5 must not change the selector count: golden=%d selectors=%d", len(golden.Selectors), len(Selectors))
	}
}

// TestHarness_049007_ImportPathSelectorsFindExpectedPrimitives covers all
// three E-T7 fixtures. The nested fixture paths keep these cases out of the
// flat golden corpus until the E-T7 implementation promotes them.
func TestHarness_049007_ImportPathSelectorsFindExpectedPrimitives(t *testing.T) {
	requireWritepathHarnessTask(t)
	cases := []struct {
		name string
		want []string
	}{
		{
			name: "reject-ioutil-write-primitives.go",
			want: []string{
				"scripts/testdata/writepath/reject-ioutil-write-primitives.go:5: write primitive 'unix.*' found",
				"scripts/testdata/writepath/reject-ioutil-write-primitives.go:9: write primitive 'ioutil.WriteFile' found",
				"scripts/testdata/writepath/reject-ioutil-write-primitives.go:10: write primitive 'ioutil.TempFile' found",
				"scripts/testdata/writepath/reject-ioutil-write-primitives.go:11: write primitive 'ioutil.TempDir' found",
			},
		},
		{
			name: "reject-syscall-namespace-primitives.go",
			want: []string{
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:6: write primitive 'syscall.WriteFile' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:7: write primitive 'syscall.Open' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:8: write primitive 'syscall.Unlink' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:9: write primitive 'syscall.Rename' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:10: write primitive 'syscall.Mkdir' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:11: write primitive 'syscall.Rmdir' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:12: write primitive 'syscall.CreateHardLink' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:13: write primitive 'syscall.DeleteFile' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:14: write primitive 'syscall.MoveFile' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:15: write primitive 'syscall.RemoveDirectory' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:16: write primitive 'syscall.CreateDirectory' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:17: write primitive 'syscall.CreateSymbolicLink' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:18: write primitive 'syscall.Truncate' found",
				"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:19: write primitive 'syscall.Creat' found",
			},
		},
		{
			name: "reject-xsys-write-primitives.go",
			want: []string{
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:9: write primitive 'windows.WriteFile' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:10: write primitive 'windows.CreateFile' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:11: write primitive 'windows.DeleteFile' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:12: write primitive 'windows.MoveFile' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:13: write primitive 'windows.MoveFileEx' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:14: write primitive 'windows.CreateDirectory' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:15: write primitive 'windows.RemoveDirectory' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:16: write primitive 'windows.CreateHardLink' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:17: write primitive 'windows.CreateSymbolicLink' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:18: write primitive 'windows.SetEndOfFile' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:19: write primitive 'windows.SetFileInformationByHandle' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:23: write primitive 'unix.Open' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:24: write primitive 'unix.Openat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:25: write primitive 'unix.Openat2' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:26: write primitive 'unix.Creat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:27: write primitive 'unix.Write' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:28: write primitive 'unix.Pwrite' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:29: write primitive 'unix.Unlink' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:30: write primitive 'unix.Unlinkat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:31: write primitive 'unix.Rename' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:32: write primitive 'unix.Renameat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:33: write primitive 'unix.Renameat2' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:34: write primitive 'unix.Mkdir' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:35: write primitive 'unix.Mkdirat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:36: write primitive 'unix.Rmdir' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:37: write primitive 'unix.Link' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:38: write primitive 'unix.Linkat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:39: write primitive 'unix.Symlink' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:40: write primitive 'unix.Symlinkat' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:41: write primitive 'unix.Truncate' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:42: write primitive 'unix.Ftruncate' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:43: write primitive 'unix.Chmod' found",
				"scripts/testdata/writepath/reject-xsys-write-primitives.go:44: write primitive 'unix.Fchmodat' found",
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			source := harnessFixtureSource(t, tc.name)
			file, err := parser.ParseFile(token.NewFileSet(), tc.name, source, parser.ParseComments|parser.AllErrors)
			if err != nil || file == nil {
				t.Fatalf("fixture %q must parse, file=%v err=%v", tc.name, file != nil, err)
			}
			if tc.name == "reject-ioutil-write-primitives.go" && !hasWritepathImport(file, "golang.org/x/sys/unix", ".") {
				t.Fatal("ioutil fixture must include the required unix dot import")
			}
			if tc.name == "reject-xsys-write-primitives.go" {
				if !hasWritepathImport(file, "golang.org/x/sys/windows", "win") {
					t.Fatal("x/sys fixture must alias the windows import")
				}
				if !hasWritepathImport(file, "golang.org/x/sys/unix", "") {
					t.Fatal("x/sys fixture must import unix without an alias")
				}
			}
			relPath := "scripts/testdata/writepath/" + tc.name
			got, err := scanSourceForHarness(t, relPath, source)
			if err != nil {
				t.Fatalf("scanSource(%q): %v", relPath, err)
			}
			assertFindingsEqual(t, got, tc.want)
		})
	}
}

// hasWritepathImport reports whether parsed contains an import whose path and
// explicit local name match the requested values; an empty wantName requires
// an unaliased import.
func hasWritepathImport(parsed *ast.File, wantPath, wantName string) bool {
	for _, spec := range parsed.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		if path != wantPath {
			continue
		}
		if wantName == "" && spec.Name == nil {
			return true
		}
		if spec.Name != nil && spec.Name.Name == wantName {
			return true
		}
	}
	return false
}

func writepathHarnessFingerprint[T any](t *testing.T, value T) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal golden fingerprint input: %v", err)
	}
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func stripWritepathPassLines(stdout string, fixtureNames []string) (string, map[string]int) {
	counts := make(map[string]int, len(fixtureNames))
	for _, name := range fixtureNames {
		counts[name] = 0
	}
	lines := strings.Split(stdout, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		removed := false
		for _, name := range fixtureNames {
			if strings.HasPrefix(line, "PASS "+name+":") {
				counts[name]++
				removed = true
				break
			}
		}
		if !removed {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n"), counts
}

// TestHarness_049007_GoldenIsAdditiveAndKeepsTheD1aOracleFrozen pins the
// exact selector append order, all 50 E-T7 findings, new stream PASS lines,
// and the pre-existing Unit D golden payloads.
func TestHarness_049007_GoldenIsAdditiveAndKeepsTheD1aOracleFrozen(t *testing.T) {
	requireWritepathHarnessTask(t)
	oldSelectors := []string{
		"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
		"os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
		"os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
		"os.CreateTemp", "os.MkdirTemp", "os.Link", "os.Chown", "os.Lchown",
		"os.Chtimes", "syscall.CreateFile", "syscall.Write", "os.OpenRoot",
		"os.Root", "io.CopyN", "io.CopyBuffer",
	}
	addedSelectors := []string{
		"ioutil.WriteFile", "ioutil.TempFile", "ioutil.TempDir",
		"syscall.WriteFile", "syscall.Open", "syscall.Unlink", "syscall.Rename",
		"syscall.Mkdir", "syscall.Rmdir", "syscall.CreateHardLink", "syscall.DeleteFile",
		"syscall.MoveFile", "syscall.RemoveDirectory", "syscall.CreateDirectory",
		"syscall.CreateSymbolicLink", "syscall.Truncate", "syscall.Creat",
		"windows.WriteFile", "windows.CreateFile", "windows.DeleteFile",
		"windows.MoveFile", "windows.MoveFileEx", "windows.CreateDirectory",
		"windows.RemoveDirectory", "windows.CreateHardLink", "windows.CreateSymbolicLink",
		"windows.SetEndOfFile", "windows.SetFileInformationByHandle",
		"unix.Open", "unix.Openat", "unix.Openat2", "unix.Creat", "unix.Write",
		"unix.Pwrite", "unix.Unlink", "unix.Unlinkat", "unix.Rename", "unix.Renameat",
		"unix.Renameat2", "unix.Mkdir", "unix.Mkdirat", "unix.Rmdir", "unix.Link",
		"unix.Linkat", "unix.Symlink", "unix.Symlinkat", "unix.Truncate",
		"unix.Ftruncate", "unix.Chmod", "unix.Fchmodat",
	}
	wantSelectors := append(append([]string(nil), oldSelectors...), addedSelectors...)
	if len(addedSelectors) != 50 || len(wantSelectors) != 76 {
		t.Fatalf("E-T7 harness list is malformed: additions=%d total=%d", len(addedSelectors), len(wantSelectors))
	}
	golden := loadWritepathGolden(t)
	if len(Selectors) != len(wantSelectors) {
		t.Fatalf("not implemented: 049.007-T: Selectors has %d entries, want %d", len(Selectors), len(wantSelectors))
	}
	if len(golden.Selectors) != len(wantSelectors) {
		t.Fatalf("not implemented: 049.007-T: golden selectors has %d entries, want %d", len(golden.Selectors), len(wantSelectors))
	}
	for i, want := range wantSelectors {
		if Selectors[i] != want {
			t.Fatalf("not implemented: 049.007-T: Selectors[%d]=%q, want %q", i, Selectors[i], want)
		}
		if golden.Selectors[i] != want {
			t.Fatalf("not implemented: 049.007-T: golden Selectors[%d]=%q, want %q", i, golden.Selectors[i], want)
		}
	}
	if len(oracleFrozenSelectors) != len(oldSelectors[:20]) {
		t.Fatalf("not implemented: 049.007-T: D-T1a oracle selector count=%d, want frozen 20", len(oracleFrozenSelectors))
	}
	for i, want := range oldSelectors[:20] {
		if oracleFrozenSelectors[i] != want {
			t.Fatalf("not implemented: 049.007-T: D-T1a oracle selector[%d]=%q, want frozen %q", i, oracleFrozenSelectors[i], want)
		}
	}

	taskFixtures := map[string][]string{
		"reject-ioutil-write-primitives.go": {
			"scripts/testdata/writepath/reject-ioutil-write-primitives.go:5: write primitive 'unix.*' found",
			"scripts/testdata/writepath/reject-ioutil-write-primitives.go:9: write primitive 'ioutil.WriteFile' found",
			"scripts/testdata/writepath/reject-ioutil-write-primitives.go:10: write primitive 'ioutil.TempFile' found",
			"scripts/testdata/writepath/reject-ioutil-write-primitives.go:11: write primitive 'ioutil.TempDir' found",
		},
		"reject-syscall-namespace-primitives.go": {
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:6: write primitive 'syscall.WriteFile' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:7: write primitive 'syscall.Open' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:8: write primitive 'syscall.Unlink' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:9: write primitive 'syscall.Rename' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:10: write primitive 'syscall.Mkdir' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:11: write primitive 'syscall.Rmdir' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:12: write primitive 'syscall.CreateHardLink' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:13: write primitive 'syscall.DeleteFile' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:14: write primitive 'syscall.MoveFile' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:15: write primitive 'syscall.RemoveDirectory' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:16: write primitive 'syscall.CreateDirectory' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:17: write primitive 'syscall.CreateSymbolicLink' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:18: write primitive 'syscall.Truncate' found",
			"scripts/testdata/writepath/reject-syscall-namespace-primitives.go:19: write primitive 'syscall.Creat' found",
		},
		"reject-xsys-write-primitives.go": {
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:9: write primitive 'windows.WriteFile' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:10: write primitive 'windows.CreateFile' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:11: write primitive 'windows.DeleteFile' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:12: write primitive 'windows.MoveFile' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:13: write primitive 'windows.MoveFileEx' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:14: write primitive 'windows.CreateDirectory' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:15: write primitive 'windows.RemoveDirectory' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:16: write primitive 'windows.CreateHardLink' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:17: write primitive 'windows.CreateSymbolicLink' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:18: write primitive 'windows.SetEndOfFile' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:19: write primitive 'windows.SetFileInformationByHandle' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:23: write primitive 'unix.Open' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:24: write primitive 'unix.Openat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:25: write primitive 'unix.Openat2' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:26: write primitive 'unix.Creat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:27: write primitive 'unix.Write' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:28: write primitive 'unix.Pwrite' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:29: write primitive 'unix.Unlink' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:30: write primitive 'unix.Unlinkat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:31: write primitive 'unix.Rename' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:32: write primitive 'unix.Renameat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:33: write primitive 'unix.Renameat2' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:34: write primitive 'unix.Mkdir' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:35: write primitive 'unix.Mkdirat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:36: write primitive 'unix.Rmdir' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:37: write primitive 'unix.Link' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:38: write primitive 'unix.Linkat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:39: write primitive 'unix.Symlink' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:40: write primitive 'unix.Symlinkat' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:41: write primitive 'unix.Truncate' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:42: write primitive 'unix.Ftruncate' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:43: write primitive 'unix.Chmod' found",
			"scripts/testdata/writepath/reject-xsys-write-primitives.go:44: write primitive 'unix.Fchmodat' found",
		},
	}
	for _, name := range []string{"reject-ioutil-write-primitives.go", "reject-syscall-namespace-primitives.go", "reject-xsys-write-primitives.go"} {
		var row *fixtureFindingsRow
		for i := range golden.FixtureFindings {
			if golden.FixtureFindings[i].Name == name {
				if row != nil {
					t.Fatalf("not implemented: 049.007-T: golden contains duplicate fixture row %q", name)
				}
				row = &golden.FixtureFindings[i]
			}
		}
		if row == nil {
			t.Fatalf("not implemented: 049.007-T: golden row %q is missing", name)
		}
		if !row.ExpectRejected || len(row.Findings) != len(taskFixtures[name]) {
			t.Fatalf("not implemented: 049.007-T: golden row %q must reject with %d findings, got %+v", name, len(taskFixtures[name]), *row)
		}
		for i, want := range taskFixtures[name] {
			if row.Findings[i] != want {
				t.Fatalf("not implemented: 049.007-T: golden row %q finding[%d]=%q, want %q", name, i, row.Findings[i], want)
			}
		}
	}

	addedFixtureNames := []string{
		"reject-alias-os-writefile.go", "reject-dot-import-os.go", "accept-alias-copilot.go", "reject-split-selector.go",
		"reject-resolve-first-caller.go", "accept-unrelated-resolve.go",
		"reject-ioutil-write-primitives.go", "reject-syscall-namespace-primitives.go", "reject-xsys-write-primitives.go",
	}
	oldRows := make([]fixtureFindingsRow, 0, 19)
	for _, row := range golden.FixtureFindings {
		if !containsWritepathFixture(addedFixtureNames, row.Name) {
			oldRows = append(oldRows, row)
		}
	}
	if len(oldRows) != 19 || writepathHarnessFingerprint(t, oldRows) != "ac30a140477ed314a7cb3057bc57c1fa3236cce65e08e967d9cf352fd0540014" {
		t.Fatalf("not implemented: 049.007-T: pre-existing fixture findings changed or the frozen fixture set is not additive")
	}

	streams := make(map[string]streamRow, len(golden.StreamCaptures))
	for name, capture := range golden.StreamCaptures {
		capture.Stdout, _ = stripWritepathPassLines(capture.Stdout, addedFixtureNames)
		streams[name] = capture
	}
	for _, captureName := range []string{"self_test", "self_test_integrity"} {
		capture, ok := golden.StreamCaptures[captureName]
		if !ok {
			t.Fatalf("not implemented: 049.007-T: stream capture %q is missing", captureName)
		}
		_, counts := stripWritepathPassLines(capture.Stdout, []string{
			"reject-ioutil-write-primitives.go", "reject-syscall-namespace-primitives.go", "reject-xsys-write-primitives.go",
		})
		for _, name := range []string{"reject-ioutil-write-primitives.go", "reject-syscall-namespace-primitives.go", "reject-xsys-write-primitives.go"} {
			if counts[name] != 1 {
				t.Fatalf("not implemented: 049.007-T: %s must contain exactly one PASS line for %s, got %d", captureName, name, counts[name])
			}
		}
	}
	if writepathHarnessFingerprint(t, streams) != "73005ef06aadac71349111be7dd3f2486a951623c9d48d56176b9bd7fb0426eb" {
		t.Fatalf("not implemented: 049.007-T: pre-existing stream capture bytes changed")
	}
	if writepathHarnessFingerprint(t, golden.Filebased) != "c06679b1feec220a212b90a8e7b3bd7d4310852dc58fa483999fe621559c8813" {
		t.Fatalf("not implemented: 049.007-T: pre-existing filebased golden bytes changed")
	}
}

func containsWritepathFixture(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}

const e7ParentSelectorScan = `(^|[^A-Za-z0-9_.])(ioutil\.(WriteFile|TempFile|TempDir)|syscall\.(WriteFile|Open|Unlink|Rename|Mkdir|Rmdir|CreateHardLink|DeleteFile|MoveFile|RemoveDirectory|CreateDirectory|CreateSymbolicLink|Truncate|Creat)|windows\.(WriteFile|CreateFile|DeleteFile|MoveFile|MoveFileEx|CreateDirectory|RemoveDirectory|CreateHardLink|CreateSymbolicLink|SetEndOfFile|SetFileInformationByHandle)|unix\.(Open|Openat|Openat2|Creat|Write|Pwrite|Unlink|Unlinkat|Rename|Renameat|Renameat2|Mkdir|Mkdirat|Rmdir|Link|Linkat|Symlink|Symlinkat|Truncate|Ftruncate|Chmod|Fchmodat))([^A-Za-z0-9_]|$)|golang\.org/x/sys`

func runWritepathGitGrep(t *testing.T, root, revision, pattern string) []string {
	t.Helper()
	cmd := exec.Command("git", "grep", "-nE", pattern, revision, "--", "internal/**", "cmd/**")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil
		}
		t.Fatalf("git grep at %s: %v", revision, err)
	}
	return strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
}

// TestHarness_049007_ParentTreeScansAreClean pins the rev-7 inclusive
// selector scan and the import-path alias scan against the task's parent
// commit. The sole selector match is the documented comment; all import-path
// hits must be unaliased import specs.
func TestHarness_049007_ParentTreeScansAreClean(t *testing.T) {
	requireWritepathHarnessTask(t)
	root := repoRoot(t)
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = root
	revisionBytes, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD: %v", err)
	}
	revision := strings.TrimSpace(string(revisionBytes))
	if revision == "" {
		t.Fatal("git rev-parse HEAD returned an empty revision")
	}
	result := runRepoScan(root, DefaultGitRunner)
	if result.Code != 0 || result.Stdout != "" || result.Stderr != "" {
		t.Fatalf("not implemented: 049.007-T: tracked internal/** and cmd/** scan must remain clean, got %+v", result)
	}
	selectorHits := runWritepathGitGrep(t, root, revision, e7ParentSelectorScan)
	if len(selectorHits) != 1 ||
		!strings.Contains(selectorHits[0], "internal/pathsafe/reparse_windows.go:15:") ||
		!strings.Contains(selectorHits[0], "//") ||
		!strings.Contains(selectorHits[0], "golang.org/x/sys/windows") {
		t.Fatalf("not implemented: 049.007-T: inclusive 50-selector scan at parent %s returned unexpected hits: %q", revision, selectorHits)
	}
	importHits := runWritepathGitGrep(t, root, revision, `"(io/ioutil|syscall|golang\.org/x/sys/(windows|unix))"`)
	if len(importHits) != 4 {
		t.Fatalf("not implemented: 049.007-T: parent import-path scan found %d hits, want the four recorded unaliased imports: %q", len(importHits), importHits)
	}
	for _, hit := range importHits {
		parts := strings.SplitN(hit, ":", 4)
		if len(parts) != 4 {
			t.Fatalf("not implemented: 049.007-T: parent import-path hit is not an unaliased import spec: %q", hit)
		}
		importSpec := strings.TrimSpace(parts[3])
		importSpec = strings.TrimPrefix(importSpec, "import ")
		if !strings.HasPrefix(importSpec, `"`) || !strings.HasSuffix(importSpec, `"`) {
			t.Fatalf("not implemented: 049.007-T: parent import-path hit is aliased, dot-imported, or not an import spec: %q", hit)
		}
	}
}

// TestHarness_049007_D1aFixtureSetRemainsFrozen protects AC-E2.6's original
// 19-name fixture corpus when E-T7 promotes its three new cases into the
// ordinary golden corpus.
func TestHarness_049007_D1aFixtureSetRemainsFrozen(t *testing.T) {
	requireWritepathHarnessTask(t)
	oraclePath := filepath.Join(repoRoot(t), "tools", "gatecheck", "internal", "writepath", "writepath_oracle_test.go")
	source, err := os.ReadFile(oraclePath)
	if err != nil {
		t.Fatalf("read D-T1a oracle source: %v", err)
	}
	oracleSource := string(source)
	frozenFixtures := []string{
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
	for _, name := range frozenFixtures {
		if !strings.Contains(oracleSource, `"`+name+`"`) {
			t.Fatalf("not implemented: 049.007-T: D-T1a oracle does not pin frozen fixture %q", name)
		}
	}
	if strings.Contains(oracleSource, "filepath.Glob(") {
		t.Fatalf("not implemented: 049.007-T: D-T1a oracle must not expand its frozen fixture list from the promoted flat corpus")
	}
}
