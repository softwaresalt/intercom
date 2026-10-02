package writepath

import (
	"encoding/json"
	"errors"
	"os"
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

			findings := scanText(row.Name, gomask.MaskGoNonCode(got))
			if len(findings) != len(row.Findings) {
				t.Fatalf("scanText(%s) findings=%v, want %v", row.Name, findings, row.Findings)
			}
			for i := range findings {
				if findings[i] != row.Findings[i] {
					t.Fatalf("scanText(%s) findings[%d]=%q, want %q", row.Name, i, findings[i], row.Findings[i])
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
		{"tag-shaped raw string", "syscall.CreateFile(p, 0, `json:\"x\"`, nil, " + oe + ", " + bs + ", 0)\n"},
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
// the finding; the selector is passed as a parameter because
// syscall.CreateFile is not yet in Selectors.
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
