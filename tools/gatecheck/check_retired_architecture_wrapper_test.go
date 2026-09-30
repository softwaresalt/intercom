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

// TestCheckRetiredArchitectureWrapper_RejectsUnknownFirstArgument covers the
// PR #83 (round-4/round-5) Copilot findings: gatecheck_invoke appends the
// trusted "--root <ROOT>" ahead of any args
// scripts/check-retired-architecture.sh forwards, but main.go's parseRoot
// scans the WHOLE arg list for "--root"/"--root=" and the LAST occurrence
// wins — so blindly forwarding "$@" let a caller-supplied --root silently
// override the trusted engine-location anchor (breaking the C-3
// root-anchoring contract).
//
// Round 4 closed that gap by rejecting any invocation with more than one
// argument, but round 5 correctly flagged that as an unrelated CLI-contract
// change: the pre-M2-T11 wrapper's own `case "${1:-}"` dispatch inspected
// only the first argument and silently ignored $2 onward rather than
// rejecting it. The final fix captures only the first argument (still
// validated against the same three-way allowlist, still exit 2 on an
// unrecognized value) and forwards ONLY that single validated value — never
// "$@" — to gatecheck_invoke, so any later argument (a duplicate --root
// included) is inert exactly as it was pre-M2-T11.
//
// This test covers the unresolved case: an invalid FIRST argument is still
// rejected with exit 2 before gatecheck_build/gatecheck_invoke ever run. It
// never needs a working Go toolchain shim and is unaffected by the
// unrelated Windows short-path issue that TestGatecheckBuild_* et al. hit
// when they actually build.
//
// Local-machine caveat (not a defect in this test or the fix it covers):
// requireBash() does a plain exec.LookPath("bash"), and on a Windows
// development machine that also has WSL installed, that resolves to
// C:\Windows\System32\bash.exe (the WSL launcher) ahead of Git-Bash, whose
// different filesystem view mangles the POSIX-style script path
// toBashPath() produces. This is the exact same pre-existing PATH-shadowing
// class already affecting every other test in this file (the ones with
// generated wrapper.sh fixtures) and is not introduced by this change.
// Verified independently via a direct Git-Bash invocation (both a plain
// exit-code check and a `bash -x` trace) that the guard clause and the
// single-argument forwarding behave exactly as asserted below.
// GitHub-hosted runners (ubuntu-latest, the blocking gate, and
// windows-latest, advisory-only) do not have WSL pre-installed, so this
// PATH-shadowing does not occur there.
func TestCheckRetiredArchitectureWrapper_RejectsUnknownFirstArgument(t *testing.T) {
	bash := requireBash(t)
	script := retiredArchWrapperPath(t)

	const usageMsg = "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]"

	cases := []struct {
		name string
		args []string
	}{
		{"bare --root flag as first argument is rejected", []string{"--root", "/tmp/evil"}},
		{"bare --root= form as first argument is rejected", []string{"--root=/tmp/evil"}},
		{"unknown first argument is rejected", []string{"--bogus"}},
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

// TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst asserts
// the CLI-surface-preservation half of the round-5 finding: when the first
// argument is one of the three allowlisted modes, every later argument
// (including a would-be root-override attempt) is silently ignored — never
// forwarded to gatecheck_invoke — exactly reproducing the pre-M2-T11
// wrapper's own "$2 onward is inert" contract. This is verified via a
// `bash -x` trace of the actual gatecheck_invoke call line rather than
// waiting on a real `go build`, so a slow/unavailable Go toolchain cannot
// make this test flaky; any resulting build failure happens strictly after
// the point this test inspects.
func TestCheckRetiredArchitectureWrapper_IgnoresArgumentsAfterTheFirst(t *testing.T) {
	bash := requireBash(t)
	script := retiredArchWrapperPath(t)

	cases := []struct {
		name         string
		args         []string
		wantInvoke   string // exact "gatecheck_invoke retired-arch ..." trace line, sans leading "+ "
		mustNotMatch string // a substring that must never appear anywhere in the trace
	}{
		{
			name:         "self-test mode ignores a trailing root-override attempt",
			args:         []string{"--self-test", "--root", "/tmp/evil"},
			wantInvoke:   "gatecheck_invoke retired-arch --self-test",
			mustNotMatch: "/tmp/evil",
		},
		{
			name:         "self-test-integrity mode ignores a second allowlisted mode",
			args:         []string{"--self-test-integrity", "--self-test"},
			wantInvoke:   "gatecheck_invoke retired-arch --self-test-integrity",
			mustNotMatch: "gatecheck_invoke retired-arch --self-test-integrity --self-test",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			traceArgs := append([]string{"-x", script}, tc.args...)
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
			if !strings.Contains(out, "+ "+tc.wantInvoke) {
				t.Fatalf("trace missing expected single-argument invoke line %q: %q", tc.wantInvoke, out)
			}
			if strings.Contains(out, tc.mustNotMatch) {
				t.Fatalf("trace unexpectedly contains %q (argument leaked past the guard): %q", tc.mustNotMatch, out)
			}
		})
	}
}

// TestCheckRetiredArchitectureWrapper_AllowlistedModesPassValidation covers
// the accompanying green path: each mode the pre-M2-T11 wrapper accepted
// must still clear the restored allowlist guard (i.e. the guard rejects
// exactly the invalid first-argument shapes above and nothing else). This
// test only asserts the process gets past the guard clause — it
// intentionally does NOT assert a final exit code, since clearing the guard
// proceeds into gatecheck_build (a real `go build`), which this test does
// not want to depend on. Passing the guard is observable because the
// process no longer prints the argument-validation usage message on
// stderr.
func TestCheckRetiredArchitectureWrapper_AllowlistedModesPassValidation(t *testing.T) {
	bash := requireBash(t)
	script := retiredArchWrapperPath(t)

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

// retiredArchWrapperPath resolves the absolute, bash-friendly path to
// scripts/check-retired-architecture.sh relative to this test file's own
// package directory (tools/gatecheck).
func retiredArchWrapperPath(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return toBashPath(filepath.Join(repoRoot, "scripts", "check-retired-architecture.sh"))
}
