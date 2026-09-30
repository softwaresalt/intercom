package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// TestRun_EachStub verifies that every one of the four pre-registered
// sub-command stubs returns exitError{1, "<name>: not yet ported"} when
// given a valid --root.
func TestRun_EachStub(t *testing.T) {
	root := t.TempDir()
	names := []string{"write-path", "retired-arch", "unignore", "merge-strategy-evaluate"}
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
