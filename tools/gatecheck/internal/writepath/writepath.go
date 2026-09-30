// Package writepath reimplements the repository's write-path precondition
// gate (scripts/check-write-path-precondition.sh) on top of
// tools/gatecheck/internal/{pysem,gomask}, so the invariant -- no
// destructive filesystem write primitive exists in any non-test Go file
// under internal/** or cmd/** -- can be enforced without a Python
// interpreter.
//
// Run mirrors the bash wrapper's own top-level dispatch (bare invocation,
// --self-test, --self-test-integrity, or an unknown flag) rather than only
// the inner Python script's mode argument, so a future CLI wrapper (M1-T10)
// only needs to forward argv and the resolved --root.
package writepath

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/gomask"
	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Selectors is the ordered list of 20 qualified write-primitive selectors
// this gate detects, ported verbatim from the SELECTORS list in
// scripts/check-write-path-precondition.sh (including the three
// adversarial-review additions noted there).
var Selectors = []string{
	"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.RemoveAll",
	"os.Rename", "os.Mkdir", "os.MkdirAll", "os.Symlink", "os.Chmod",
	"os.Truncate", "io.Copy", "sql.Open", "bbolt.Open",
	"os.CreateTemp", "os.MkdirTemp", "os.Link", "os.Chown", "os.Lchown",
	"os.Chtimes",
}

const (
	exceptionName = "011.003-T's Constitution Check exception (internal/config/validate.go rule 7)"
	registerName  = "the consolidated risk register (internal/pathsafe package doc, root.go)"
)

// GitRunner runs `git ls-files -- internal/** cmd/**` rooted at root and
// returns its raw stdout bytes (for pysem.GitText decoding), or an error if
// the process could not be started or exited non-zero. It is injectable so
// tests can simulate a missing/failing git without depending on the real
// repository tree (M1-T8 AC: "an injectable gitRunner").
type GitRunner func(root string) ([]byte, error)

// DefaultGitRunner is the production GitRunner: it shells out to
// `git ls-files -- internal/** cmd/**` with root as the working directory.
func DefaultGitRunner(root string) ([]byte, error) {
	cmd := exec.Command("git", "ls-files", "--", "internal/**", "cmd/**")
	cmd.Dir = root
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// shouldScan reports whether relPath (a forward-slash, repo-root-relative
// path as produced by `git ls-files`) is a non-test Go source file under
// internal/** or cmd/**. Ported verbatim from should_scan().
func shouldScan(relPath string) bool {
	if strings.HasSuffix(relPath, "_test.go") {
		return false
	}
	if !strings.HasSuffix(relPath, ".go") {
		return false
	}
	return strings.HasPrefix(relPath, "internal/") || strings.HasPrefix(relPath, "cmd/")
}

// findSelector reports whether line contains sel as a qualified selector:
// not preceded by a word rune or '.', and not followed by a word rune. It
// is the explicit replacement for Python's
// re.compile(r'(?<![\w.])' + re.escape(sel) + r'(?![\w])').search(line):
// only whether at least one satisfying occurrence exists matters (a
// boolean hit), not an enumeration of every occurrence, matching
// pattern.search's truthiness semantics.
func findSelector(line, sel string) bool {
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

// scanText scans already-decoded, newline-translated, masked text for
// every selector hit, returning one finding string per (line, selector)
// pair in Selectors order, formatted exactly as Python's
// f"{relPath}:{line_no}: write primitive {sel!r} found" (via pysem.Repr).
// relPath is echoed back into the finding text verbatim and must already
// be in the caller's desired display form (forward-slash, repo-relative).
func scanText(relPath, maskedText string) []string {
	var findings []string
	for i, line := range pysem.SplitLines(maskedText) {
		lineNo := i + 1
		for _, sel := range Selectors {
			if findSelector(line, sel) {
				findings = append(findings, fmt.Sprintf("%s:%d: write primitive %s found", relPath, lineNo, pysem.Repr(sel)))
			}
		}
	}
	return findings
}

// scanFile reads the file at filepath.Join(root, filepath.FromSlash(relPath))
// through pysem.ReadText (rejecting invalid UTF-8 and translating
// newlines), masks it via gomask.MaskGoNonCode, and scans the result.
// relPath is echoed back into finding text verbatim.
func scanFile(root, relPath string) ([]string, error) {
	text, err := pysem.ReadText(filepath.Join(root, filepath.FromSlash(relPath)))
	if err != nil {
		return nil, err
	}
	return scanText(relPath, gomask.MaskGoNonCode(text)), nil
}

// Result carries the ordered stdout/stderr text and process-style exit
// code a single mode produced, mirroring writepath_golden.json's
// stream_captures shape.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// errorLine formats a one-line, "::error::"-prefixed ED-2 message: the
// replacement for what would otherwise be an uncaught Python exception
// (a git failure, a file read failure, or invalid UTF-8) propagating as a
// traceback. context names the file or operation that failed.
func errorLine(context string, err error) string {
	return fmt.Sprintf("::error::%s: %v\n", context, err)
}

// runRepoScan performs the repo-mode invariant scan: enumerate every
// tracked internal/**, cmd/** Go source (via git, injectable), scan each,
// and fail closed on a git error, a per-file read/decode error (ED-2), any
// finding, or an empty selection (ED-7).
func runRepoScan(root string, git GitRunner) Result {
	out, err := git(root)
	if err != nil {
		return Result{Stderr: errorLine("git ls-files", err), Code: 1}
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return Result{Stderr: errorLine("git ls-files output", err), Code: 1}
	}

	var relPaths []string
	for _, p := range pysem.SplitLines(listing) {
		if p == "" {
			continue
		}
		if shouldScan(p) {
			relPaths = append(relPaths, p)
		}
	}

	if len(relPaths) == 0 {
		return Result{
			Stderr: "::error::write-path repo scan matched no files under internal/** or cmd/** (ED-7)\n",
			Code:   1,
		}
	}

	var findings []string
	for _, rel := range relPaths {
		fs, err := scanFile(root, rel)
		if err != nil {
			return Result{Stderr: errorLine(rel, err), Code: 1}
		}
		findings = append(findings, fs...)
	}

	if len(findings) > 0 {
		var b strings.Builder
		for _, f := range findings {
			b.WriteString(f)
			b.WriteByte('\n')
		}
		_, _ = fmt.Fprintf(&b,
			"::error::a destructive filesystem write primitive was found under internal/** or cmd/**. "+
				"This trips the write-path precondition recorded in %s and tracked in %s. "+
				"RETIREMENT PROCEDURE: re-evaluate every finding in the risk register against this new "+
				"write call site, land a mitigation (or an explicit, re-justified acceptance) in the SAME "+
				"change, and update both the register and the Constitution Check exception to reflect the "+
				"arrival of a real write path.\n",
			exceptionName, registerName,
		)
		return Result{Stderr: b.String(), Code: 1}
	}

	return Result{Code: 0}
}

// runFixtureSelfTest performs ONLY the fixture self-test: every
// scripts/testdata/writepath/*.go fixture is scanned and checked against
// its accept-/reject- filename prefix. It is shared, verbatim, by both the
// "self-test" and "self-test-integrity" top-level modes; whether the repo
// scan also runs afterward is Run's concern, not this function's.
func runFixtureSelfTest(root string) Result {
	fixtureDir := filepath.Join(root, "scripts", "testdata", "writepath")
	fixtureDirPosix := filepath.ToSlash(fixtureDir)

	info, statErr := os.Stat(fixtureDir)
	if statErr != nil || !info.IsDir() {
		return Result{Stderr: fmt.Sprintf("fixture dir not found: %s\n", fixtureDirPosix), Code: 1}
	}

	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		return Result{Stderr: fmt.Sprintf("fixture dir not found: %s\n", fixtureDirPosix), Code: 1}
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".go") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	if len(names) == 0 {
		return Result{Stderr: fmt.Sprintf("no fixtures discovered under %s\n", fixtureDirPosix), Code: 1}
	}

	var stdout strings.Builder
	var failures []string
	for _, name := range names {
		relPath := "scripts/testdata/writepath/" + name
		findings, err := scanFile(root, relPath)
		if err != nil {
			return Result{Stderr: errorLine(relPath, err), Code: 1}
		}
		rejected := len(findings) > 0
		switch {
		case strings.HasPrefix(name, "reject-"):
			if rejected {
				_, _ = fmt.Fprintf(&stdout, "PASS %s: rejected as expected\n", name)
			} else {
				failures = append(failures, fmt.Sprintf("%s: expected rejection (write primitive), got clean", name))
			}
		case strings.HasPrefix(name, "accept-"):
			if rejected {
				failures = append(failures, fmt.Sprintf("%s: expected clean, got findings: %s", name, strings.Join(findings, "; ")))
			} else {
				_, _ = fmt.Fprintf(&stdout, "PASS %s: clean as expected\n", name)
			}
		default:
			failures = append(failures, fmt.Sprintf("%s: fixture filename must start with 'accept-' or 'reject-'", name))
		}
	}

	if len(failures) > 0 {
		var b strings.Builder
		for _, f := range failures {
			b.WriteString("FAIL " + f + "\n")
		}
		return Result{Stdout: stdout.String(), Stderr: b.String(), Code: 1}
	}

	return Result{Stdout: stdout.String(), Code: 0}
}

// Run dispatches the top-level CLI-style modes exactly like
// scripts/check-write-path-precondition.sh's case statement:
//
//	flag == ""                      -> repo-mode scan only
//	flag == "--self-test"           -> fixture self-test, then repo scan,
//	                                    then a trailing banner on success
//	flag == "--self-test-integrity" -> fixture self-test only, then a
//	                                    trailing banner on success
//	anything else                   -> usage to stderr, exit 2
//
// It writes ordered stdout/stderr bytes to the given writers and returns
// the process-style exit code (M1-T8; M1-T10 wires this into
// `gatecheck write-path`).
func Run(flag, root string, git GitRunner, stdout, stderr io.Writer) int {
	switch flag {
	case "":
		res := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		return res.Code

	case "--self-test":
		res := runFixtureSelfTest(root)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		if res.Code != 0 {
			return res.Code
		}
		repoRes := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, repoRes.Stdout)
		_, _ = io.WriteString(stderr, repoRes.Stderr)
		if repoRes.Code != 0 {
			return repoRes.Code
		}
		_, _ = io.WriteString(stdout, "self-test passed: fixtures matched expectations and the tracked tree is clean\n")
		return 0

	case "--self-test-integrity":
		res := runFixtureSelfTest(root)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		if res.Code != 0 {
			return res.Code
		}
		_, _ = io.WriteString(stdout, "self-test-integrity passed: fixtures matched expectations (repo scan skipped, 030.001-T)\n")
		return 0

	default:
		_, _ = io.WriteString(stderr, "usage: scripts/check-write-path-precondition.sh [--self-test|--self-test-integrity]\n")
		return 2
	}
}
