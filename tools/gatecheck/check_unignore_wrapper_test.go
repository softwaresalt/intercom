package main

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCheckUnignoreWrapper_RejectsRootInjectionAnywhere covers the same
// root containment as the retired-arch wrapper's allowlist (see
// check_retired_architecture_wrapper_test.go): gatecheck_invoke prepends
// the trusted "--root <ROOT>" ahead of any args
// scripts/check-unignore-regression.sh forwards, and main.go's parseRoot
// now treats that first "--root"/"--root=" option as authoritative and
// rejects any later root option. The trusted root is therefore
// authoritative; this wrapper's denylist is retained as defense-in-depth.
//
// Unlike retired-arch (which forwards only a single validated mode
// string), unignore's real CLI surface needs multiple flags forwarded
// (--self-test, --base-ref <ref>, --head-ref <ref>), so instead of
// collapsing to one argument, the wrapper checks every forwarded argument
// against a literal "--root"/"--root=*" denylist before gatecheck_build
// ever runs. This test asserts that guard fires regardless of where
// --root appears in the argument list (first, middle, or last), and
// before any other validation.
//
// Local-machine caveat (not a defect in this test): requireBash() does a
// plain exec.LookPath("bash"), and on a Windows development machine that
// also has WSL installed, that resolves to C:\Windows\System32\bash.exe
// (the WSL launcher) ahead of Git-Bash, whose different filesystem view
// mangles the POSIX-style script path toBashPath() produces. This is the
// exact same pre-existing PATH-shadowing class documented in
// check_retired_architecture_wrapper_test.go and is not introduced by
// this change. GitHub-hosted runners (ubuntu-latest, the blocking gate,
// and windows-latest, advisory-only) do not have WSL pre-installed, so
// this PATH-shadowing does not occur there.
func TestCheckUnignoreWrapper_RejectsRootInjectionAnywhere(t *testing.T) {
	bash := requireBash(t)
	script := unignoreWrapperPath(t)

	cases := []struct {
		name string
		args []string
	}{
		{"bare --root flag as sole argument", []string{"--root", "/tmp/evil"}},
		{"bare --root= form as sole argument", []string{"--root=/tmp/evil"}},
		{"--root after a legitimate flag", []string{"--self-test", "--root", "/tmp/evil"}},
		{"--root= after a legitimate flag", []string{"--self-test", "--root=/tmp/evil"}},
		{"--root between two legitimate flags", []string{"--base-ref", "HEAD~1", "--root", "/tmp/evil", "--head-ref", "HEAD"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(bash, append([]string{script}, tc.args...)...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			runErr := cmd.Run()
			exitCode := 0
			var exitErr *exec.ExitError
			switch {
			case errors.As(runErr, &exitErr):
				exitCode = exitErr.ExitCode()
			case runErr != nil:
				t.Fatalf("unexpected error running wrapper: %v (stderr=%q)", runErr, stderr.String())
			}

			if exitCode != 2 {
				t.Fatalf("exit code = %d, want 2 (stderr=%q)", exitCode, stderr.String())
			}
			if !strings.Contains(stderr.String(), "::error::unrecognized argument: --root") {
				t.Fatalf("stderr missing root-injection guard message: %q", stderr.String())
			}
		})
	}
}

// TestCheckUnignoreWrapper_LegitimateFlagsClearTheGuard asserts the
// CLI-surface-preservation half: every flag combination unignore's real
// surface needs (--self-test; --base-ref/--head-ref; -h/--help) clears the
// root-injection guard and reaches gatecheck_build, so the guard never
// collapses the wrapper's multi-flag forwarding contract down to a
// single-argument allowlist the way retired-arch's did. Verified via a
// `bash -x` trace rather than waiting on a real `go build`, so a slow or
// unavailable Go toolchain cannot make this test flaky.
func TestCheckUnignoreWrapper_LegitimateFlagsClearTheGuard(t *testing.T) {
	bash := requireBash(t)
	script := unignoreWrapperPath(t)

	for _, args := range [][]string{
		{"--self-test"},
		{"--base-ref", "HEAD~1"},
		{"--base-ref", "HEAD~1", "--head-ref", "HEAD"},
		{"-h"},
		{"--help"},
	} {
		name := strings.Join(args, "_")
		t.Run(name, func(t *testing.T) {
			traceArgs := append([]string{"-x", script}, args...)
			cmd := exec.Command(bash, traceArgs...)
			var combined bytes.Buffer
			cmd.Stdout = &combined
			cmd.Stderr = &combined

			if startErr := cmd.Start(); startErr != nil {
				t.Fatalf("start wrapper: %v", startErr)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(20 * time.Second):
				_ = cmd.Process.Kill()
				<-done
			}

			out := combined.String()
			if strings.Contains(out, "::error::unrecognized argument") {
				t.Fatalf("root-injection guard rejected a legitimate invocation: %q", out)
			}
			if !strings.Contains(out, "gatecheck_build") {
				t.Fatalf("trace never reached gatecheck_build, guard may not have been cleared: %q", out)
			}
		})
	}
}

// unignoreWrapperPath resolves the absolute, bash-friendly path to
// scripts/check-unignore-regression.sh relative to this test file's own
// package directory (tools/gatecheck).
func unignoreWrapperPath(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return toBashPath(filepath.Join(repoRoot, "scripts", "check-unignore-regression.sh"))
}
