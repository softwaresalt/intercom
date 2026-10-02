package main

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
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

	for _, args := range [][]string{
		{"--self-test"},
		{"--self-test-integrity"},
	} {
		name := strings.Join(args, "_")
		t.Run(name, func(t *testing.T) {
			trace, _ := runWritePathWrapperWithTrace(t, bash, script, args...)
			if strings.Contains(trace, "::error::unrecognized argument") {
				t.Errorf("root-injection guard rejected a legitimate invocation: %q", trace)
			}
			if !strings.Contains(trace, "gatecheck_build") {
				t.Errorf("trace never reached gatecheck_build: %q", trace)
			}
		})
	}
}

func runWritePathWrapperWithTrace(t *testing.T, bash, script string, args ...string) (string, int) {
	t.Helper()

	traceArgs := append([]string{"-x", script}, args...)
	cmd := exec.Command(bash, traceArgs...)
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
