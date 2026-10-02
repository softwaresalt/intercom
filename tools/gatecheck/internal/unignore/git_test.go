package unignore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExitCode(t *testing.T) {
	if got := exitCode(nil); got != 0 {
		t.Fatalf("exitCode(nil) = %d, want 0", got)
	}
	if got := exitCode(errors.New("boom")); got != -1 {
		t.Fatalf("exitCode(generic error) = %d, want -1", got)
	}
}

func TestExitCode_ExitError(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	// `git this-is-not-a-command` reliably exits 1.
	cmd := exec.Command("git", "this-is-not-a-command")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected git to fail on an unknown subcommand")
	}
	if got := exitCode(err); got <= 0 {
		t.Fatalf("exitCode(exec.ExitError) = %d, want > 0", got)
	}
}

func TestRootGitignoreTextAt_HEAD_ReadsDiskDirectly(t *testing.T) {
	dir := t.TempDir()
	// No .gitignore present: must degrade to empty content, no error.
	text, err := rootGitignoreTextAt(dir, "HEAD", DefaultGitRunner)
	if err != nil {
		t.Fatalf("unexpected error for absent .gitignore: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty content for absent .gitignore, got %q", text)
	}

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("foo\nbar\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	text, err = rootGitignoreTextAt(dir, "HEAD", DefaultGitRunner)
	if err != nil {
		t.Fatalf("unexpected error reading present .gitignore: %v", err)
	}
	if text != "foo\nbar\n" {
		t.Fatalf("got %q, want %q", text, "foo\nbar\n")
	}
}

func TestRootGitignoreTextAt_ValidRef(t *testing.T) {
	git := newIsolatedGitRunner(t)
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := git(dir, nil, args...); err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, stderr)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("committed-content\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	run("add", ".gitignore")
	run("commit", "-q", "-m", "base")

	text, err := rootGitignoreTextAt(dir, "HEAD", git)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "committed-content\n" {
		t.Fatalf("got %q, want %q", text, "committed-content\n")
	}

	// Also confirm resolving a non-HEAD ref reads the same content via
	// `git show <ref>:.gitignore`.
	stdout, _, revErr := git(dir, nil, "rev-parse", "HEAD")
	if revErr != nil {
		t.Fatalf("rev-parse HEAD: %v", revErr)
	}
	sha := string(stdout)
	// Trim the trailing newline git rev-parse emits.
	for len(sha) > 0 && (sha[len(sha)-1] == '\n' || sha[len(sha)-1] == '\r') {
		sha = sha[:len(sha)-1]
	}
	text, err = rootGitignoreTextAt(dir, sha, git)
	if err != nil {
		t.Fatalf("unexpected error resolving by sha: %v", err)
	}
	if text != "committed-content\n" {
		t.Fatalf("got %q, want %q", text, "committed-content\n")
	}
}

// initRepoWithCommit creates a git repository in a fresh temp directory
// with a single commit containing only a placeholder file (no root
// .gitignore), and returns the directory. The repository is otherwise
// valid, so any failure to resolve a baseline from it is attributable to
// the ref (or the absent file) alone.
func initRepoWithCommit(t *testing.T, git GitRunner) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := git(dir, nil, args...); err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, stderr)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "placeholder.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write placeholder: %v", err)
	}
	run("add", "placeholder.txt")
	run("commit", "-q", "-m", "base without .gitignore")
	return dir
}

// TestRootGitignoreTextAt_InvalidRef_ReturnsError pins the 035-F fix
// (AC-B1.1/AC-B1.2): an unresolvable ref must surface a distinct,
// non-nil error at the point of use, independent of the upstream
// `git diff` guard in runDifferentialCheck. Before the fix, an invalid
// ref silently degraded to an empty baseline, so no ignore rule could
// ever be observed as removed and the merge-blocking regression check
// passed vacuously.
func TestRootGitignoreTextAt_InvalidRef_ReturnsError(t *testing.T) {
	git := newIsolatedGitRunner(t)
	dir := initRepoWithCommit(t, git)

	const badRef = "this-ref-does-not-exist"
	text, err := rootGitignoreTextAt(dir, badRef, git)
	if err == nil {
		t.Fatalf("expected an error for unresolvable ref %q, got nil (text %q): an invalid ref must not degrade to an empty baseline", badRef, text)
	}
	if text != "" {
		t.Fatalf("expected empty text alongside the error, got %q", text)
	}
	if !strings.Contains(err.Error(), badRef) {
		t.Fatalf("error %q does not name the unresolvable ref %q", err, badRef)
	}
}

// TestRootGitignoreTextAt_ValidRefAbsentGitignore_ReturnsEmpty guards the
// other half of the 035-F distinction (plan stop condition): a ref that
// resolves but simply has no root .gitignore is a legitimate empty
// baseline, NOT an error. Over-strictness here would turn a legitimately
// absent .gitignore into a merge-blocking failure.
func TestRootGitignoreTextAt_ValidRefAbsentGitignore_ReturnsEmpty(t *testing.T) {
	git := newIsolatedGitRunner(t)
	dir := initRepoWithCommit(t, git)

	// "main" is a valid, non-HEAD ref, so this exercises the
	// `git show <ref>:.gitignore` path rather than the HEAD disk read.
	text, err := rootGitignoreTextAt(dir, "main", git)
	if err != nil {
		t.Fatalf("expected an empty baseline for a valid ref without .gitignore, got error: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty content for absent .gitignore at a valid ref, got %q", text)
	}
}

func TestMakeScratchGitignore(t *testing.T) {
	git := newIsolatedGitRunner(t)
	root := t.TempDir()

	scratchDir, err := makeScratchGitignore(root, "example", "foo\n", git)
	if err != nil {
		t.Fatalf("makeScratchGitignore: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(scratchDir, ".gitignore"))
	if err != nil {
		t.Fatalf("read scratch .gitignore: %v", err)
	}
	if string(data) != "foo\n" {
		t.Fatalf("got %q, want %q", string(data), "foo\n")
	}

	// The scratch directory must be a real git repository, since
	// check-ignore --no-index requires one.
	ignored, ignErr := isIgnored(git, scratchDir, "foo")
	if ignErr != nil {
		t.Fatalf("isIgnored on scratch repo: %v", ignErr)
	}
	if !ignored {
		t.Fatal("expected foo to be ignored in the scratch repo")
	}
}
