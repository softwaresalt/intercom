package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "repo_scan"},
		{name: "--self-test", args: []string{"--self-test"}},
		{name: "--self-test-integrity", args: []string{"--self-test-integrity"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			goShimLog := installGoShim(t, bash)
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

func workspaceTempDir(t *testing.T) string {
	t.Helper()

	workspace, err := os.Getwd()
	if err != nil {
		t.Fatalf("get workspace working directory: %v", err)
	}
	workspace, err = filepath.Abs(workspace)
	if err != nil {
		t.Fatalf("resolve workspace working directory: %v", err)
	}
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, workspace)
	}

	dir := t.TempDir()
	absoluteDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolve temporary directory: %v", err)
	}
	relativeDir, err := filepath.Rel(workspace, absoluteDir)
	if err != nil {
		t.Fatalf("check temporary directory containment: %v", err)
	}
	if relativeDir == ".." ||
		filepath.IsAbs(relativeDir) ||
		strings.HasPrefix(relativeDir, ".."+string(filepath.Separator)) {
		t.Fatalf("temporary directory is outside workspace")
	}
	for _, name := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(name, absoluteDir)
	}
	return dir
}

func installGoShim(t *testing.T, bash string) string {
	t.Helper()

	dir := workspaceTempDir(t)
	bashDir := resolveBashPath(t, bash, dir)
	shimPath := filepath.Join(dir, "go")
	logPath := filepath.Join(dir, "go-calls.log")
	t.Setenv("GATECHECK_GOEXE", platformGoExe())
	t.Setenv("GATECHECK_SENTINEL_LOG", path.Join(bashDir, "generated-binary-calls.log"))
	shim := `#!/usr/bin/env bash
set -eu
shim_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
printf '%s\n' "$*" >> "${shim_dir}/go-calls.log"
unexpected() {
  printf 'unexpected go shim invocation (exit code 97): %s\n' "$*" >&2
  exit 97
}
case "${1-}" in
env)
  if [ "$#" -eq 2 ] && [ "$2" = "GOEXE" ]; then
    printf '%s\n' "${GATECHECK_GOEXE}"
    exit 0
  fi
  ;;
-C)
  if [ "$#" -ne 7 ] ||
    [ "$3" != "build" ] ||
    [ "$4" != "-trimpath" ] ||
    [ "$5" != "-o" ] ||
    [ "$7" != "./tools/gatecheck" ]; then
    unexpected "$*"
  fi
  managed_tmpdir="${TMPDIR%/}"
  case "$6" in
    "${managed_tmpdir}"/*) ;;
    *) unexpected "gatecheck output is outside managed TMPDIR: $6 (TMPDIR=$TMPDIR)" ;;
  esac
  printf '%s\n' "$6" > "${shim_dir}/build-output-path.log"
  printf '#!/usr/bin/env bash\nset -eu\nprintf "invoked\\n" >> "${GATECHECK_SENTINEL_LOG:?}"\nexit 0\n' > "$6"
  chmod +x "$6"
  exit 0
  ;;
esac
unexpected "$*"
`
	if err := os.WriteFile(shimPath, []byte(shim), 0o755); err != nil {
		t.Fatalf("write go shim: %v", err)
	}
	if err := os.Chmod(shimPath, 0o755); err != nil {
		t.Fatalf("make go shim executable: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	cmd := exec.Command(bash, "-c", "command -v go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("locate go shim: %v (output: %q)", err, output)
	}
	if got, want := strings.TrimSpace(string(output)), path.Join(bashDir, "go"); !strings.EqualFold(filepath.ToSlash(got), filepath.ToSlash(want)) {
		t.Fatalf("go resolves to %q, want test shim %q", got, want)
	}
	return logPath
}

func platformGoExe() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func resolveBashPath(t *testing.T, bash, windowsOrNativePath string) string {
	t.Helper()

	cmd := exec.Command(bash, "-c", `directory="$1"
if command -v cygpath >/dev/null 2>&1; then
  directory="$(cygpath -u "$directory")"
elif command -v wslpath >/dev/null 2>&1; then
  directory="$(wslpath -u "$directory")"
fi
cd "$directory" && pwd`, "bash", windowsOrNativePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("resolve Bash path for %q: %v (output: %q)", windowsOrNativePath, err, output)
	}
	return strings.TrimSpace(string(output))
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
