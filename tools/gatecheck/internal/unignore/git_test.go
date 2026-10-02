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

func TestRootTreeHasGitignore(t *testing.T) {
	entry := func(name string) string {
		return "100644 blob 0123456789abcdef0123456789abcdef01234567\t" + name + "\x00"
	}
	cases := []struct {
		name    string
		listing string
		want    bool
		wantErr bool
	}{
		{name: "empty listing", listing: "", want: false},
		{name: "exact entry", listing: entry("README.md") + entry(".gitignore"), want: true},
		{name: "near-miss names only", listing: entry(".gitignore-extra") + entry("x.gitignore") + entry(".gitignore.bak"), want: false},
		{name: "malformed record", listing: "100644 blob deadbeef .gitignore\x00", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := rootTreeHasGitignore([]byte(tc.listing))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// realExitError returns a genuine *exec.ExitError (git exiting non-zero),
// so fake runners can reproduce a git process that ran and failed, which
// exitCode distinguishes from a process that never started.
func realExitError(t *testing.T) error {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	err := exec.Command("git", "this-is-not-a-command").Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError from an unknown git subcommand, got %v", err)
	}
	return err
}

// TestRootGitignoreTextAt_ShowFailureClassification drives the non-HEAD
// failure branches with a scripted GitRunner: every outcome other than a
// resolvable ref with a positively absent root .gitignore is an error.
func TestRootGitignoreTextAt_ShowFailureClassification(t *testing.T) {
	exitErr := realExitError(t)
	absentListing := "100644 blob 0123456789abcdef0123456789abcdef01234567\tREADME.md\x00"
	presentListing := absentListing + "100644 blob 0123456789abcdef0123456789abcdef01234567\t.gitignore\x00"

	type reply struct {
		stdout string
		err    error
	}
	cases := []struct {
		name    string
		ref     string
		replies map[string]reply
		wantErr bool
		wantSub string
	}{
		{
			name:    "git could not start",
			ref:     "main",
			replies: map[string]reply{"show": {err: errors.New("exec: not found")}},
			wantErr: true,
			wantSub: "could not run",
		},
		{
			name:    "leading dash ref",
			ref:     "-x",
			replies: map[string]reply{"show": {err: exitErr}},
			wantErr: true,
			wantSub: "not a valid ref",
		},
		{
			name: "unresolvable ref",
			ref:  "nope",
			replies: map[string]reply{
				"show":      {err: exitErr},
				"rev-parse": {err: exitErr},
			},
			wantErr: true,
			wantSub: "cannot resolve ref",
		},
		{
			name: "rev-parse returns no object",
			ref:  "main",
			replies: map[string]reply{
				"show":      {err: exitErr},
				"rev-parse": {stdout: "\n"},
			},
			wantErr: true,
			wantSub: "no object",
		},
		{
			name: "ls-tree fails",
			ref:  "main",
			replies: map[string]reply{
				"show":      {err: exitErr},
				"rev-parse": {stdout: "abc123\n"},
				"ls-tree":   {err: exitErr},
			},
			wantErr: true,
			wantSub: "ls-tree",
		},
		{
			name: "present but unshowable",
			ref:  "main",
			replies: map[string]reply{
				"show":      {err: exitErr},
				"rev-parse": {stdout: "abc123\n"},
				"ls-tree":   {stdout: presentListing},
			},
			wantErr: true,
			wantSub: "exists",
		},
		{
			name: "resolvable ref with absent .gitignore",
			ref:  "main",
			replies: map[string]reply{
				"show":      {err: exitErr},
				"rev-parse": {stdout: "abc123\n"},
				"ls-tree":   {stdout: absentListing},
			},
			wantErr: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			fake := func(_ string, _ []byte, args ...string) ([]byte, []byte, error) {
				calls = append(calls, args[0])
				r, ok := tc.replies[args[0]]
				if !ok {
					t.Fatalf("unexpected git %v", args)
				}
				return []byte(r.stdout), []byte("scripted stderr"), r.err
			}
			text, err := rootGitignoreTextAt(t.TempDir(), tc.ref, fake)
			if text != "" {
				t.Fatalf("expected empty text, got %q", text)
			}
			if !tc.wantErr {
				if err != nil {
					t.Fatalf("expected nil error, got %v (calls %v)", err, calls)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error, got nil (calls %v)", calls)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err, tc.wantSub)
			}
		})
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
