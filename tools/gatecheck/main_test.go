package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_RootOccurrences(t *testing.T) {
	trustedRoot := t.TempDir()
	overrideRoot := t.TempDir()

	const subcommand = "test-root-occurrences"
	var dispatched bool
	var gotRoot string
	registerSubcommand(subcommand, func(_ []string, root string, _ io.Reader, _ io.Writer, _ io.Writer) int {
		dispatched = true
		gotRoot = root
		return 0
	})
	t.Cleanup(func() { delete(subcommands, subcommand) })

	tests := []struct {
		name           string
		rootArgs       []string
		wantCode       int
		wantRoot       string
		wantDiagnostic string
	}{
		{
			name:     "trusted only",
			rootArgs: []string{"--root", trustedRoot},
			wantCode: 0,
			wantRoot: trustedRoot,
		},
		{
			name:           "trusted plus trailing --root",
			rootArgs:       []string{"--root", trustedRoot, "--root", overrideRoot},
			wantCode:       1,
			wantDiagnostic: "--root specified more than once",
		},
		{
			name:           "trusted plus trailing --root=",
			rootArgs:       []string{"--root", trustedRoot, "--root=" + overrideRoot},
			wantCode:       1,
			wantDiagnostic: "--root specified more than once",
		},
		{
			name:           "--root without a value",
			rootArgs:       []string{"--root"},
			wantCode:       1,
			wantDiagnostic: "--root requires a value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dispatched = false
			gotRoot = ""
			args := append([]string{subcommand}, tc.rootArgs...)
			var stdout, stderr bytes.Buffer
			code := run(args, strings.NewReader(""), &stdout, &stderr)
			if code != tc.wantCode {
				t.Fatalf("code = %d, want %d (stderr: %q)", code, tc.wantCode, stderr.String())
			}
			if tc.wantDiagnostic != "" && !strings.Contains(stderr.String(), tc.wantDiagnostic) {
				t.Fatalf("stderr = %q, want diagnostic containing %q", stderr.String(), tc.wantDiagnostic)
			}
			if tc.wantRoot != "" {
				if !dispatched {
					t.Fatal("subcommand was not dispatched")
				}
				if gotRoot != tc.wantRoot {
					t.Fatalf("dispatched root = %q, want trusted root %q", gotRoot, tc.wantRoot)
				}
			} else if dispatched {
				t.Fatalf("subcommand was dispatched with root %q after an invalid --root occurrence", gotRoot)
			}
		})
	}
}

// TestRun_NoArgs verifies that invoking run() with no arguments prints usage
// to stderr and returns 2 (C-2: the gatecheck tool's own usage errors return
// 2; no gate observes this because the wrappers never call it that way).
func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want usage text")
	}
}

// TestRun_UnknownSubcommand verifies that an unrecognized sub-command name
// prints usage to stderr and returns 2, matching the no-args case.
func TestRun_UnknownSubcommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"bogus-subcommand"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want usage text")
	}
}

// TestRun_MissingRoot verifies that a recognized sub-command without a
// --root flag returns 1 with an ::error:: message on stderr (C-3: the
// common --root flag is parsed once in run(), before dispatch).
func TestRun_MissingRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"write-path"}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "::error::") {
		t.Fatalf("stderr = %q, want ::error:: message", stderr.String())
	}
}

// TestRun_NonDirectoryRoot verifies that a --root pointing at a non-directory
// path (or a nonexistent path) returns 1 with an ::error:: message.
func TestRun_NonDirectoryRoot(t *testing.T) {
	tmp := t.TempDir()
	notADir := filepath.Join(tmp, "not-a-dir.txt")
	if err := os.WriteFile(notADir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}

	cases := []struct {
		name string
		root string
	}{
		{"file, not directory", notADir},
		{"nonexistent path", filepath.Join(tmp, "does-not-exist")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{"write-path", "--root", tc.root}, strings.NewReader(""), &stdout, &stderr)
			if code != 1 {
				t.Fatalf("code = %d, want 1", code)
			}
			if !strings.Contains(stderr.String(), "::error::") {
				t.Fatalf("stderr = %q, want ::error:: message", stderr.String())
			}
		})
	}
}

// TestRun_EachStub verifies that every remaining pre-registered
// sub-command stub returns exitError{1, "<name>: not yet ported"} when
// given a valid --root. "write-path" is excluded: M1-T10 replaced its stub
// with the real internal/writepath dispatch (see TestRun_WritePath_*
// below). "retired-arch" is excluded: M2-T11 replaced its stub with the
// real internal/retiredarch dispatch (see TestRun_RetiredArch_* below).
// "unignore" is excluded: M3-T7 replaced its stub with the real
// internal/unignore dispatch (see internal/unignore's own test suite).
// "merge-strategy-evaluate" is excluded: M3-T11 replaced its stub with the
// real internal/mergestrategy dispatch (see internal/mergestrategy's own
// test suite). All four originally-registered sub-commands are now real,
// so `names` is empty; this test is retained as the structural harness for
// any future sub-command that lands behind a "not yet ported" stub before
// its own engine is wired in.
func TestRun_EachStub(t *testing.T) {
	root := t.TempDir()
	names := []string{}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run([]string{name, "--root", root}, strings.NewReader(""), &stdout, &stderr)
			if code != 1 {
				t.Fatalf("code = %d, want 1", code)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
			want := name + ": not yet ported"
			if got := strings.TrimRight(stderr.String(), "\n"); got != want {
				t.Fatalf("stderr = %q, want %q", got, want)
			}
		})
	}
}

// TestExitError_RoundTrip verifies exitError's code/message fields survive
// construction and its Error() string, and that writeExitError reproduces
// both the exact message on the writer and the code as its return value.
func TestExitError_RoundTrip(t *testing.T) {
	e := &exitError{code: 1, msg: "widget: not yet ported"}
	if e.Error() != "widget: not yet ported" {
		t.Fatalf("Error() = %q, want %q", e.Error(), "widget: not yet ported")
	}
	var stderr bytes.Buffer
	code := writeExitError(&stderr, e)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if got := strings.TrimRight(stderr.String(), "\n"); got != e.msg {
		t.Fatalf("stderr = %q, want %q", got, e.msg)
	}
}

// TestRun_PanickingStubRecovers verifies the C-2 top-level recover: a
// sub-command that panics is converted into a
// "::error::gatecheck internal error: <value>" message on stderr and exit
// code 1, never Go's default panic exit code 2. The panicking handler is
// registered only for the duration of this test, under a name that never
// collides with a production sub-command.
func TestRun_PanickingStubRecovers(t *testing.T) {
	const name = "test-panic-stub"
	registerSubcommand(name, func(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
		panic("kaboom")
	})
	t.Cleanup(func() { delete(subcommands, name) })

	var stdout, stderr bytes.Buffer
	code := run([]string{name, "--root", t.TempDir()}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	msg := stderr.String()
	if !strings.Contains(msg, "::error::gatecheck internal error:") {
		t.Fatalf("stderr = %q, want internal-error message", msg)
	}
	if !strings.Contains(msg, "kaboom") {
		t.Fatalf("stderr = %q, want panic value %q", msg, "kaboom")
	}
}

// gatecheckRepoRoot returns the repository top-level directory from this
// test file's own path (tools/gatecheck), so CLI-dispatch tests can point
// --root at the real tracked tree without depending on `go test`'s working
// directory.
func gatecheckRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, "..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return root
}

// TestRun_WritePath_DispatchesToEngine verifies M1-T10's real wiring: run()
// parses --root and forwards the remaining positional argument to
// internal/writepath.Run via runWritePath, rather than the old
// "not yet ported" stub. The engine's own behaviour (every mode, every
// fixture, every golden) is exhaustively covered by
// internal/writepath's own test suite; this test only proves the CLI-level
// plumbing (arg -> flag, --root -> root, DefaultGitRunner wiring) is
// correct.
func TestRun_WritePath_DispatchesToEngine(t *testing.T) {
	root := gatecheckRepoRoot(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"write-path", "--root", root}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("repo-mode code = %d, want 0 (stderr: %q)", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("repo-mode stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("repo-mode stderr = %q, want empty", stderr.String())
	}
}

// TestRun_WritePath_BogusFlag verifies the CLI-level plumbing surfaces
// writepath.Run's own usage/exit-2 fallback for an unrecognized mode flag,
// matching the pre-switch bash wrapper's own bogus-flag case byte for byte
// (see m1.md, M1-T10 evidence).
func TestRun_WritePath_BogusFlag(t *testing.T) {
	root := gatecheckRepoRoot(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"write-path", "--root", root, "--bogus-flag"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	want := "usage: scripts/check-write-path-precondition.sh [--self-test|--self-test-integrity]\n"
	if got := stderr.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

// TestRun_RetiredArch_DispatchesToEngine verifies M2-T11's real wiring:
// run() parses --root and forwards the remaining positional argument to
// internal/retiredarch.Run via runRetiredArch, rather than the old
// "not yet ported" stub. The engine's own behaviour (every mode, every
// fixture, every golden) is exhaustively covered by internal/retiredarch's
// own test suite; this test only proves the CLI-level plumbing (arg ->
// flag, --root -> root, DefaultGitRunner wiring) is correct.
func TestRun_RetiredArch_DispatchesToEngine(t *testing.T) {
	root := gatecheckRepoRoot(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"retired-arch", "--root", root}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("repo-mode code = %d, want 0 (stdout: %q, stderr: %q)", code, stdout.String(), stderr.String())
	}
	want := "::notice::retired-arch gate mode=repo\n"
	if got := stdout.String(); got != want {
		t.Fatalf("repo-mode stdout = %q, want %q", got, want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("repo-mode stderr = %q, want empty", stderr.String())
	}
}

// TestRun_RetiredArch_BogusFlag verifies the CLI-level plumbing surfaces
// retiredarch.Run's own usage/exit-2 fallback for an unrecognized mode
// flag, matching the pre-switch bash wrapper's own bogus-flag case byte
// for byte (see m2.md, M2-T11 evidence).
func TestRun_RetiredArch_BogusFlag(t *testing.T) {
	root := gatecheckRepoRoot(t)

	var stdout, stderr bytes.Buffer
	code := run([]string{"retired-arch", "--root", root, "--bogus-flag"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	want := "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]\n"
	if got := stderr.String(); got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}
