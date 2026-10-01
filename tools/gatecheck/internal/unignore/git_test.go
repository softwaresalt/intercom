package unignore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

// TestRootGitignoreTextAt_InvalidRef_DegradesToEmpty characterizes the
// 035-F defect (plan §6): an invalid ref -- or a valid ref that simply
// lacks a root .gitignore -- both degrade silently to an empty baseline
// rather than surfacing an error. This port preserves that behaviour
// faithfully; 035-F re-plans the fix for after M4.
func TestRootGitignoreTextAt_InvalidRef_DegradesToEmpty(t *testing.T) {
	git := newIsolatedGitRunner(t)
	dir := t.TempDir()
	if _, stderr, err := git(dir, nil, "init", "-q", "-b", "main"); err != nil {
		t.Fatalf("git init failed: %v: %s", err, stderr)
	}

	text, err := rootGitignoreTextAt(dir, "this-ref-does-not-exist", git)
	if err != nil {
		t.Fatalf("expected degraded-empty, not an error: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty content for invalid ref, got %q", text)
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
