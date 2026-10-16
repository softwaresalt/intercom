package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckWritePathWrapper_RejectsRootInjectionBeforeBuild(t *testing.T) {
	bash := requireBash(t)
	script := writePathWrapperPath(t)

	cases := []struct {
		name string
		args []string
	}{
		{"bare --root flag", []string{"--root", "/tmp/evil"}},
		{"--root= form", []string{"--root=/tmp/evil"}},
		{"--root after a legitimate flag", []string{"--self-test", "--root", "/tmp/evil"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trace, exitCode := runWritePathWrapperWithTrace(t, bash, script, tc.args...)
			if exitCode != 2 {
				t.Errorf("exit code = %d, want 2 (trace=%q)", exitCode, trace)
			}
			if !strings.Contains(trace, "::error::unrecognized argument") {
				t.Errorf("trace missing root-injection guard diagnostic: %q", trace)
			}
			if strings.Contains(trace, "gatecheck_build") {
				t.Errorf("trace reached gatecheck_build before rejecting root injection: %q", trace)
			}
		})
	}
}

func TestCheckWritePathWrapper_LegitimateModesReachBuild(t *testing.T) {
	bash := requireBash(t)
	script := writePathWrapperPath(t)
	shimDir := installGoShimDir(t, bash)

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "repo_scan"},
		{name: "--self-test", args: []string{"--self-test"}},
		{name: "--self-test-integrity", args: []string{"--self-test-integrity"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// One install per test (see wrapper_harness_test.go); each subtest
			// starts from empty shim logs so its call counts stay isolated.
			for _, name := range []string{"go-calls.log", "build-output-path.log", "generated-binary-calls.log"} {
				if err := os.Remove(filepath.Join(shimDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("reset shim log %s: %v", name, err)
				}
			}
			goShimLog := filepath.Join(shimDir, "go-calls.log")
			trace, exitCode := runWritePathWrapperWithTrace(t, bash, script, tc.args...)
			if exitCode != 0 {
				t.Errorf("exit code = %d, want 0 (trace=%q)", exitCode, trace)
			}
			if strings.Contains(trace, "::error::unrecognized argument") {
				t.Errorf("root-injection guard rejected a legitimate invocation: %q", trace)
			}
			if !strings.Contains(trace, "gatecheck_build") {
				t.Errorf("trace never reached gatecheck_build: %q", trace)
			}
			goShimCalls, err := os.ReadFile(goShimLog)
			if err != nil {
				t.Fatalf("read go shim calls: %v", err)
			}
			calls := strings.Split(strings.TrimSpace(string(goShimCalls)), "\n")
			if len(calls) != 2 {
				t.Errorf("go shim received %d invocations, want exactly env GOEXE and build: %q", len(calls), goShimCalls)
				return
			}
			if calls[0] != "env GOEXE" {
				t.Errorf("first go shim invocation = %q, want %q", calls[0], "env GOEXE")
			}
			if !strings.HasPrefix(calls[1], "-C ") ||
				!strings.Contains(calls[1], " build -trimpath -o ") ||
				!strings.HasSuffix(calls[1], " ./tools/gatecheck") {
				t.Errorf("second go shim invocation did not receive the expected build command: %q", calls[1])
			}
			expectedGoExe := platformGoExe()
			buildOutputPath, err := os.ReadFile(filepath.Join(filepath.Dir(goShimLog), "build-output-path.log"))
			if err != nil {
				t.Fatalf("read generated binary output path: %v", err)
			}
			if got, want := path.Base(strings.TrimSpace(string(buildOutputPath))), "gatecheck"+expectedGoExe; got != want {
				t.Errorf("generated binary output name = %q, want %q", got, want)
			}

			sentinelLog := filepath.Join(filepath.Dir(goShimLog), "generated-binary-calls.log")
			sentinel, err := os.ReadFile(sentinelLog)
			if err != nil {
				t.Fatalf("read generated binary sentinel: %v", err)
			}
			if got, want := strings.TrimSpace(string(sentinel)), "invoked"; got != want {
				t.Errorf("generated binary sentinel = %q, want %q", got, want)
			}
		})
	}
}

func TestInstallGoShim_UnexpectedInvocationFailsClearly(t *testing.T) {
	bash := requireBash(t)
	installGoShim(t, bash)

	cmd := exec.Command(bash, "-c", "go unexpected-argument")
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start go shim invocation: %v", err)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	var err error
	select {
	case err = <-done:
	case <-time.After(20 * time.Second):
		killErr := cmd.Process.Kill()
		waitErr := <-done
		t.Fatalf("go shim invocation timed out after 20s (kill err=%v, wait err=%v; output=%q)", killErr, waitErr, output.String())
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("unexpected go shim invocation error = %v, want exit code 97 (output: %q)", err, output.String())
	}
	if got, want := exitErr.ExitCode(), 97; got != want {
		t.Errorf("unexpected go shim exit code = %d, want %d (output: %q)", got, want, output.String())
	}
	if !strings.Contains(output.String(), "unexpected go shim invocation") {
		t.Errorf("unexpected go shim invocation output lacks diagnostic: %q", output.String())
	}
}

func runWritePathWrapperWithTrace(t *testing.T, bash, script string, args ...string) (string, int) {
	t.Helper()
	return runWritePathWrapperInDir(t, bash, script, "", args...)
}

// runWritePathWrapperInDir runs the wrapper with dir as its working directory
// ("" keeps the test's own). It is the trace runner the wrong-tree rows use to
// place the caller's cwd inside a fixture checkout (055.001-T).
func runWritePathWrapperInDir(t *testing.T, bash, script, dir string, args ...string) (string, int) {
	t.Helper()

	traceArgs := append([]string{"-x", script}, args...)
	cmd := exec.Command(bash, traceArgs...)
	cmd.Dir = dir
	var trace bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &trace

	if startErr := cmd.Start(); startErr != nil {
		t.Fatalf("start wrapper: %v", startErr)
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(20 * time.Second):
		killErr := cmd.Process.Kill()
		waitErr := <-done
		t.Fatalf("wrapper timed out after 20s (kill err=%v, wait err=%v; trace=%q)", killErr, waitErr, trace.String())
	}

	exitCode := 0
	var exitErr *exec.ExitError
	switch {
	case errors.As(runErr, &exitErr):
		exitCode = exitErr.ExitCode()
	case runErr != nil:
		t.Fatalf("unexpected error running wrapper: %v (trace=%q)", runErr, trace.String())
	}

	return trace.String(), exitCode
}

func writePathWrapperPath(t *testing.T) string {
	t.Helper()
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return toBashPath(filepath.Join(repoRoot, "scripts", "check-write-path-precondition.sh"))
}

// TestCheckWritePathWrapper_WrongTreeRootAnchoring proves the write-path
// wrapper derives ROOT from its own checkout, never from the ambient Git
// environment or the caller's working directory (055.001-T, DBE25DF5, AS-5).
// The generated binary records every argument it receives, so each engine row
// reads the --root the wrapper hands it and compares that root with the
// checkout by file identity.
func TestCheckWritePathWrapper_WrongTreeRootAnchoring(t *testing.T) {
	bash := requireBash(t)
	base := installGoShimDir(t, bash)
	for _, row := range wrongTreeRows() {
		t.Run(row.name, func(t *testing.T) {
			runWrongTreeRow(t, bash, base, writePathWrapperRel, row)
		})
	}
}
