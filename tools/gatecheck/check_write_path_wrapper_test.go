package main

import (
	"bytes"
	"errors"
	"io"
	"os"
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
			goShimLog := installGoShim(t, bash)
			trace, _ := runWritePathWrapperWithTrace(t, bash, script, args...)
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
			if !strings.Contains(string(goShimCalls), "env GOEXE") || !strings.Contains(string(goShimCalls), "build") {
				t.Errorf("go shim did not receive the build sequence: %q", goShimCalls)
			}
		})
	}
}

func installGoShim(t *testing.T, bash string) string {
	t.Helper()

	dir, err := os.MkdirTemp(".", ".gatecheck-go-shim-")
	if err != nil {
		t.Fatalf("create go shim directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("remove go shim directory: %v", err)
		}
	})
	dir, err = filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolve go shim directory: %v", err)
	}
	shimPath := filepath.Join(dir, "go")
	logPath := filepath.Join(dir, "go-calls.log")
	shim := `#!/usr/bin/env bash
set -eu
printf '%s\n' "$*" >> "$GATECHECK_TEST_GO_LOG"
case "$1" in
env)
  if [ "$2" = "GOEXE" ]; then
    printf '\n'
    exit 0
  fi
  ;;
-C)
  shift
  output=""
  while [ "$#" -gt 0 ]; do
    if [ "$1" = "-o" ]; then
      shift
      output="$1"
      break
    fi
    shift
  done
  if [ -z "$output" ]; then
    exit 1
  fi
  printf '#!/usr/bin/env bash\nexit 0\n' > "$output"
  chmod +x "$output"
  exit 0
  ;;
esac
exit 1
`
	if err := os.WriteFile(shimPath, []byte(shim), 0o755); err != nil {
		t.Fatalf("write go shim: %v", err)
	}
	if err := os.Chmod(shimPath, 0o755); err != nil {
		t.Fatalf("make go shim executable: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GATECHECK_TEST_GO_LOG", toBashPath(logPath))

	cmd := exec.Command(bash, "-c", "command -v go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("locate go shim: %v (output: %q)", err, output)
	}
	if got, want := strings.TrimSpace(string(output)), toBashPath(shimPath); !strings.EqualFold(filepath.ToSlash(got), filepath.ToSlash(want)) {
		t.Fatalf("go resolves to %q, want test shim %q", got, want)
	}
	return logPath
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
