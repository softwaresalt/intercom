package unignore

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
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

func TestRootGitignoreTextAt_HEAD_SymlinkGitignore_Errors(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "secret-target")
	if err := os.WriteFile(target, []byte("secret/\n"), 0o644); err != nil {
		t.Fatalf("write symlink target: %v", err)
	}
	if err := os.Symlink(target, filepath.Join(dir, ".gitignore")); err != nil {
		if runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(1314)) {
			t.Skipf("Windows symlink privilege unavailable; Linux CI gitignore-append-only job must run this case: %v", err)
		}
		t.Fatalf("os.Symlink: %v", err)
	}

	text, err := rootGitignoreTextAt(dir, "HEAD", DefaultGitRunner)
	if err == nil {
		t.Fatalf("expected symlinked root .gitignore to fail closed, got text %q", text)
	}
	if text != "" {
		t.Fatalf("text = %q, want empty alongside the error", text)
	}
}

// TestRootGitignoreTextAt_HEAD_DirectoryGitignore_Errors is
// characterization: os.ReadFile already rejects a directory on the parent.
func TestRootGitignoreTextAt_HEAD_DirectoryGitignore_Errors(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".gitignore"), 0o755); err != nil {
		t.Fatalf("create directory .gitignore: %v", err)
	}
	text, err := rootGitignoreTextAt(dir, "HEAD", DefaultGitRunner)
	if err == nil {
		t.Fatalf("expected directory root .gitignore to error, got text %q", text)
	}
	if text != "" {
		t.Fatalf("text = %q, want empty alongside the error", text)
	}
	if !strings.Contains(err.Error(), ".gitignore") {
		t.Fatalf("error %q does not name .gitignore", err)
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

// TestRootGitignoreTextAt_OptionLikeRefs_RealGit proves against real git
// (not only the scripted runner) that refs which `git show <ref>:.gitignore`
// could previously satisfy without naming a baseline — an option-like
// argument, an empty ref, or a bare `:` index path — now fail closed with
// an error naming the ref, even when a committed .gitignore exists.
func TestRootGitignoreTextAt_OptionLikeRefs_RealGit(t *testing.T) {
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

	for _, ref := range []string{"--format=x", "", ":"} {
		t.Run(ref, func(t *testing.T) {
			text, err := rootGitignoreTextAt(dir, ref, git)
			if err == nil {
				t.Fatalf("expected an error for ref %q, got nil (text %q)", ref, text)
			}
			if text != "" {
				t.Fatalf("expected empty text alongside the error for ref %q, got %q", ref, text)
			}
			if !strings.Contains(err.Error(), "::error::") {
				t.Fatalf("error %q is not a ::error:: diagnostic", err)
			}
		})
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
// exitCode distinguishes from a process that never started. It runs a
// builtin subcommand through the suite's isolated runner (so neither a
// user-level alias nor a git-* executable on PATH can intercept it) that
// must fail: verifying a ref that cannot exist in a fresh, empty
// repository.
func realExitError(t *testing.T) error {
	t.Helper()
	git := newIsolatedGitRunner(t)
	dir := t.TempDir()
	if _, stderr, err := git(dir, nil, "init", "-q"); err != nil {
		t.Fatalf("git init failed: %v: %s", err, stderr)
	}
	_, _, err := git(dir, nil, "rev-parse", "--verify", "--quiet", "refs/heads/does-not-exist")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError from verifying a nonexistent ref, got %v", err)
	}
	return err
}

// TestRootGitignoreTextAt_NonHEADClassification drives the non-HEAD path
// with a scripted GitRunner: every outcome other than a resolved tree
// whose root .gitignore is either shown or positively absent is an error.
// The scripted runner fails the test on any git call it was not given a
// reply for, so cases that must reject before git runs also prove that no
// git process (in particular `git show`) is ever invoked for them.
func TestRootGitignoreTextAt_NonHEADClassification(t *testing.T) {
	exitErr := realExitError(t)
	const tree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
	absentListing := "100644 blob 0123456789abcdef0123456789abcdef01234567\tREADME.md\x00"
	presentListing := absentListing + "100644 blob 0123456789abcdef0123456789abcdef01234567\t.gitignore\x00"

	type reply struct {
		stdout string
		err    error
	}
	cases := []struct {
		name     string
		ref      string
		replies  map[string]reply
		wantText string
		wantSub  string // empty means a nil error is expected
	}{
		{
			name:    "empty ref rejected before git",
			ref:     "",
			wantSub: "not a valid ref",
		},
		{
			name: "option-like ref rejected before git even if show would succeed",
			ref:  "--format=x",
			replies: map[string]reply{
				"show": {stdout: "x\n"},
			},
			wantSub: "not a valid ref",
		},
		{
			name:    "rev-parse could not start",
			ref:     "main",
			replies: map[string]reply{"rev-parse": {err: errors.New("exec: not found")}},
			wantSub: "could not run",
		},
		{
			name:    "unresolvable ref",
			ref:     "nope",
			replies: map[string]reply{"rev-parse": {err: exitErr}},
			wantSub: "cannot resolve ref",
		},
		{
			name:    "rev-parse returns no object",
			ref:     "main",
			replies: map[string]reply{"rev-parse": {stdout: "\n"}},
			wantSub: "no object ID",
		},
		{
			name:    "rev-parse returns a non-hex object",
			ref:     "main",
			replies: map[string]reply{"rev-parse": {stdout: "-x\n"}},
			wantSub: "no object ID",
		},
		{
			name: "show could not start",
			ref:  "main",
			replies: map[string]reply{
				"rev-parse": {stdout: tree + "\n"},
				"show":      {err: errors.New("exec: not found")},
			},
			wantSub: "could not run",
		},
		{
			name: "ls-tree fails",
			ref:  "main",
			replies: map[string]reply{
				"rev-parse": {stdout: tree + "\n"},
				"show":      {err: exitErr},
				"ls-tree":   {err: exitErr},
			},
			wantSub: "ls-tree",
		},
		{
			name: "present but unshowable",
			ref:  "main",
			replies: map[string]reply{
				"rev-parse": {stdout: tree + "\n"},
				"show":      {err: exitErr},
				"ls-tree":   {stdout: presentListing},
			},
			wantSub: "exists",
		},
		{
			name: "resolvable ref with absent .gitignore",
			ref:  "main",
			replies: map[string]reply{
				"rev-parse": {stdout: tree + "\n"},
				"show":      {err: exitErr},
				"ls-tree":   {stdout: absentListing},
			},
		},
		{
			name: "resolvable ref with .gitignore",
			ref:  "main",
			replies: map[string]reply{
				"rev-parse": {stdout: tree + "\n"},
				"show":      {stdout: "foo\n"},
			},
			wantText: "foo\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls [][]string
			fake := func(_ string, _ []byte, args ...string) ([]byte, []byte, error) {
				calls = append(calls, args)
				r, ok := tc.replies[args[0]]
				if !ok {
					t.Fatalf("unexpected git %v", args)
				}
				return []byte(r.stdout), []byte("scripted stderr"), r.err
			}
			text, err := rootGitignoreTextAt(t.TempDir(), tc.ref, fake)
			if text != tc.wantText {
				t.Fatalf("got text %q, want %q", text, tc.wantText)
			}
			if tc.wantSub == "" {
				if err != nil {
					t.Fatalf("expected nil error, got %v (calls %v)", err, calls)
				}
			} else {
				if err == nil {
					t.Fatalf("expected an error, got nil (calls %v)", calls)
				}
				if !strings.Contains(err.Error(), tc.wantSub) {
					t.Fatalf("error %q does not contain %q", err, tc.wantSub)
				}
			}
			// Once the ref resolves, git only ever receives the resolved
			// tree ID, never the caller-supplied ref text.
			for _, call := range calls {
				switch call[0] {
				case "rev-parse":
					if want := []string{"rev-parse", "--verify", "--quiet", tc.ref + "^{tree}"}; strings.Join(call, " ") != strings.Join(want, " ") {
						t.Fatalf("rev-parse args %v, want %v", call, want)
					}
				case "show":
					if want := []string{"show", tree + ":.gitignore"}; strings.Join(call, " ") != strings.Join(want, " ") {
						t.Fatalf("show args %v, want %v", call, want)
					}
				case "ls-tree":
					if want := []string{"ls-tree", "--full-tree", "-z", tree}; strings.Join(call, " ") != strings.Join(want, " ") {
						t.Fatalf("ls-tree args %v, want %v", call, want)
					}
				}
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
