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

// TestCheckRetiredArchitectureWrapper_RejectsCallerRootOverride covers the
// PR #83 (round-4) Copilot finding: gatecheck_invoke appends the trusted
// "--root <ROOT>" ahead of any args scripts/check-retired-architecture.sh
// forwards, but main.go's parseRoot scans the WHOLE arg list for
// "--root"/"--root=" and the LAST occurrence wins — so blindly forwarding
// "$@" let a caller-supplied --root silently override the trusted
// engine-location anchor (breaking the C-3 root-anchoring contract).
//
// The fix restores the pre-M2-T11 wrapper's own allowlist dispatch (only
// "", --self-test, or --self-test-integrity as the sole argument) so any
// other invocation — including one smuggling a second --root — is rejected
// with exit 2 before gatecheck_build/gatecheck_invoke ever run. Every case
// here exercises only that guard clause, so it never needs a working Go
// toolchain shim and is unaffected by the unrelated Windows short-path
// issue that TestGatecheckBuild_* et al. hit when they actually build.
//
// Local-machine caveat (not a defect in this test or the fix it covers):
// requireBash() does a plain exec.LookPath("bash"), and on a Windows
// development machine that also has WSL installed, that resolves to
// C:\Windows\System32\bash.exe (the WSL launcher) ahead of Git-Bash,
// whose different filesystem view mangles the POSIX-style script path
// toBashPath() produces. This is the exact same pre-existing PATH-shadowing
// class already affecting every other test in this file (the ones with
// generated wrapper.sh fixtures) and is not introduced by this change.
// Verified independently via a direct Git-Bash invocation (both a plain
// exit-code check and a `bash -x` trace) that the guard clause behaves
// exactly as asserted below. GitHub-hosted runners (ubuntu-latest, the
// blocking gate, and windows-latest, advisory-only) do not have WSL
// pre-installed, so this PATH-shadowing does not occur there.
func TestCheckRetiredArchitectureWrapper_RejectsCallerRootOverride(t *testing.T) {
	bash := requireBash(t)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	script := toBashPath(filepath.Join(repoRoot, "scripts", "check-retired-architecture.sh"))

	const usageMsg = "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]"

	cases := []struct {
		name string
		args []string
	}{
		{"bare --root flag is rejected", []string{"--root", "/tmp/evil"}},
		{"bare --root= form is rejected", []string{"--root=/tmp/evil"}},
		{"self-test with a second --root is rejected", []string{"--self-test", "--root", "/tmp/evil"}},
		{"self-test-integrity with a second --root is rejected", []string{"--self-test-integrity", "--root=/tmp/evil"}},
		{"unknown first argument is rejected", []string{"--bogus"}},
		{"two otherwise-valid modes are rejected (extra arg)", []string{"--self-test", "--self-test-integrity"}},
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
			if !strings.Contains(stderr.String(), usageMsg) {
				t.Fatalf("stderr missing usage message: %q", stderr.String())
			}
		})
	}
}

// TestCheckRetiredArchitectureWrapper_AllowlistedModesPassValidation covers
// the accompanying green path: each mode the pre-M2-T11 wrapper accepted
// must still clear the restored allowlist guard (i.e. the guard rejects
// exactly the invalid shapes above and nothing else). This test only
// asserts the process gets past the guard clause — it intentionally does
// NOT assert a final exit code, since clearing the guard proceeds into
// gatecheck_build (a real `go build`), which this test does not want to
// depend on. Passing the guard is observable because the process no longer
// prints the argument-validation usage message on stderr.
func TestCheckRetiredArchitectureWrapper_AllowlistedModesPassValidation(t *testing.T) {
	bash := requireBash(t)

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	script := toBashPath(filepath.Join(repoRoot, "scripts", "check-retired-architecture.sh"))

	const usageMsg = "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]"

	for _, args := range [][]string{
		{},
		{"--self-test"},
		{"--self-test-integrity"},
	} {
		name := "no-args"
		if len(args) > 0 {
			name = args[0]
		}
		t.Run(name, func(t *testing.T) {
			// bash -x traces every executed command; grepping for the
			// case/if guard lines (rather than waiting on the real build)
			// confirms the guard was cleared without depending on a Go
			// toolchain being invokable in this test environment.
			traceArgs := append([]string{"-x", script}, args...)
			cmd := exec.Command(bash, traceArgs...)
			var combined bytes.Buffer
			cmd.Stdout = &combined
			cmd.Stderr = &combined
			// gatecheck_build performs a real `go build`; bound this so a
			// slow/unavailable toolchain cannot hang the test. Any
			// resulting failure/timeout happens strictly after the guard
			// clause this test cares about.
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
			if strings.Contains(out, usageMsg) {
				t.Fatalf("guard clause rejected an allowlisted invocation: %q", out)
			}
			if !strings.Contains(out, "gatecheck_build") {
				t.Fatalf("trace never reached gatecheck_build, guard may not have been cleared: %q", out)
			}
		})
	}
}
