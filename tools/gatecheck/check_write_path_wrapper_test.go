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
	"syscall"
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
	t.Setenv("GATECHECK_ARGS_LOG", path.Join(bashDir, "generated-binary-args.log"))
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
  cat > "$6" <<'SENTINEL'
#!/usr/bin/env bash
set -eu
printf 'invoked\n' >> "${GATECHECK_SENTINEL_LOG:?}"
for arg in "$@"; do printf '%s\n' "$arg" >> "${GATECHECK_ARGS_LOG:?}"; done
exit 0
SENTINEL
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

// wrongTreeLayout selects how the wrapper checkout is laid out for one
// wrong-tree row (055.001-T, D-WB2-4).
type wrongTreeLayout int

const (
	// layoutOwnRepo: the checkout is a repository in its own right.
	layoutOwnRepo wrongTreeLayout = iota
	// layoutSymlink: the wrapper is invoked through a symlink to an own-repo checkout.
	layoutSymlink
	// layoutNoRepo: no .git exists at or above the checkout (GIT_CEILING_DIRECTORIES stops discovery at the fixture base).
	layoutNoRepo
	// layoutNested: the checkout sits inside a parent repository without its own .git.
	layoutNested
)

// errorPrivilegeNotHeld is the Windows ERROR_PRIVILEGE_NOT_HELD code returned
// when the account may not create symbolic links.
const errorPrivilegeNotHeld = 1314

// wrongTreeRow is one row of the write-path wrong-tree table. ambient names the
// ambient Git environment set for the wrapper process, subdir the caller's
// working directory relative to the checkout. wantEngine is true when the
// engine must run with ROOT equal to the checkout; false when the wrapper must
// fail closed and never invoke the engine.
type wrongTreeRow struct {
	name       string
	layout     wrongTreeLayout
	ambient    string
	subdir     string
	wantEngine bool
}

// TestCheckWritePathWrapper_WrongTreeRootAnchoring proves the write-path
// wrapper derives ROOT from its own checkout, never from the ambient Git
// environment or the caller's working directory (055.001-T, DBE25DF5, AS-5).
// The go shim records the --root the wrapper hands the generated binary, so
// each row compares that root with the checkout by file identity.
func TestCheckWritePathWrapper_WrongTreeRootAnchoring(t *testing.T) {
	bash := requireBash(t)
	base := filepath.Dir(installGoShim(t, bash))
	rows := []wrongTreeRow{
		{name: "decoy GIT_DIR and GIT_WORK_TREE", layout: layoutOwnRepo, ambient: "decoy", wantEngine: true},
		{name: "mixed-case ambient names", layout: layoutOwnRepo, ambient: "mixed-case", wantEngine: true},
		{name: "GIT_DIR alone from a subdirectory", layout: layoutOwnRepo, ambient: "dir-only", subdir: "internal", wantEngine: true},
		{name: "no ambient control", layout: layoutOwnRepo, wantEngine: true},
		{name: "symlinked checkout directory", layout: layoutSymlink, wantEngine: true},
		{name: "no repository", layout: layoutNoRepo, wantEngine: false},
		{name: "nested checkout", layout: layoutNested, wantEngine: false},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			runWrongTreeRow(t, bash, base, row)
		})
	}
}

// runWrongTreeRow runs one row inside its own directory under base, the shim's
// base directory. The shim is installed once for the whole table: the shim
// directory is the managed TMP itself, so a per-row install would make Git Bash
// spell the same directory two ways (/tmp and /c/...) and fail the self-check.
// Each row's logs are addressed by their /c/... form instead.
func runWrongTreeRow(t *testing.T, bash, base string, row wrongTreeRow) {
	t.Helper()

	rowDir := filepath.Join(base, strings.NewReplacer(" ", "_").Replace(row.name))
	if err := os.MkdirAll(rowDir, 0o755); err != nil {
		t.Fatalf("create row directory: %v", err)
	}
	t.Setenv("GATECHECK_SENTINEL_LOG", path.Join(toBashPath(rowDir), "generated-binary-calls.log"))
	t.Setenv("GATECHECK_ARGS_LOG", path.Join(toBashPath(rowDir), "generated-binary-args.log"))
	decoy := filepath.Join(rowDir, "decoy")
	checkout := filepath.Join(rowDir, "checkout")
	writeWrongTreeTree(t, decoy)
	initWrongTreeRepo(t, decoy)
	writeWrongTreeTree(t, checkout)
	copyWrapperFixture(t, checkout)
	script := filepath.Join(checkout, "scripts", "check-write-path-precondition.sh")
	cwd := checkout

	switch row.layout {
	case layoutOwnRepo:
		initWrongTreeRepo(t, checkout)
	case layoutSymlink:
		initWrongTreeRepo(t, checkout)
		link := filepath.Join(rowDir, "link")
		if err := os.Symlink(checkout, link); err != nil {
			if isPrivilegeNotHeld(err) {
				t.Skipf("symlink creation needs a privilege this host does not grant: %v", err)
			}
			t.Fatalf("create checkout symlink: %v", err)
		}
		cwd = link
		script = filepath.Join(link, "scripts", "check-write-path-precondition.sh")
	case layoutNoRepo:
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.ToSlash(rowDir))
	case layoutNested:
		initWrongTreeRepo(t, rowDir)
	default:
		t.Fatalf("unknown layout %d", row.layout)
	}

	applyWrongTreeAmbient(t, row.ambient, decoy)
	if row.subdir != "" {
		cwd = filepath.Join(checkout, row.subdir)
	}

	trace, exitCode := runWritePathWrapperInDir(t, bash, toBashPath(script), cwd)
	marker := strings.TrimSpace(readOptionalFile(t, filepath.Join(rowDir, "generated-binary-calls.log")))

	if !row.wantEngine {
		if exitCode == 0 {
			t.Errorf("exit code = 0, want nonzero for a checkout that must fail closed (trace=%q)", trace)
		}
		if marker != "" {
			t.Errorf("engine was invoked for a checkout that must fail closed (marker=%q, trace=%q)", marker, trace)
		}
		return
	}

	if exitCode != 0 {
		t.Errorf("exit code = %d, want 0 (trace=%q)", exitCode, trace)
	}
	if marker != "invoked" {
		t.Fatalf("engine marker = %q, want %q: the shim must have run (trace=%q)", marker, "invoked", trace)
	}
	root := recordedRootArg(t, filepath.Join(rowDir, "generated-binary-args.log"))
	if !sameDirectory(t, bash, root, toBashPath(checkout)) {
		t.Errorf("engine was handed --root %q, want the wrapper checkout %q (trace=%q)", root, checkout, trace)
	}
}

// writeWrongTreeTree writes a minimal internal/ and cmd/ tree, the shape the
// write-path scan walks, so both repositories in a wrong-tree row look alike.
func writeWrongTreeTree(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"internal/wrongtree/wrongtree.go": "package wrongtree\n",
		"cmd/wrongtree/main.go":           "package main\n",
	}
	for rel, body := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("create %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
}

func initWrongTreeRepo(t *testing.T, dir string) {
	t.Helper()
	runIsolatedGit(t, dir, "init", "-q")
}

// copyWrapperFixture copies the live write-path wrapper and its shared runner
// into checkout, so the wrapper derives GATECHECK_SRC from the copy's location.
func copyWrapperFixture(t *testing.T, checkout string) {
	t.Helper()
	repo := gatecheckRepoRoot(t)
	for _, rel := range []string{"scripts/check-write-path-precondition.sh", "scripts/lib/gatecheck-run.sh"} {
		data, err := os.ReadFile(filepath.Join(repo, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		dst := filepath.Join(checkout, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatalf("create %s: %v", filepath.Dir(dst), err)
		}
		if err := os.WriteFile(dst, []byte(normalizeScript(string(data))), 0o755); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}
}

// applyWrongTreeAmbient sets the ambient Git environment for one row. Windows
// environment names are case-insensitive, so the mixed-case row exercises the
// locale-free scrub on hosts that keep the name's case.
func applyWrongTreeAmbient(t *testing.T, kind, decoy string) {
	t.Helper()
	gitDir := filepath.ToSlash(filepath.Join(decoy, ".git"))
	tree := filepath.ToSlash(decoy)
	switch kind {
	case "":
	case "decoy":
		t.Setenv("GIT_DIR", gitDir)
		t.Setenv("GIT_WORK_TREE", tree)
	case "mixed-case":
		t.Setenv("Git_Dir", gitDir)
		t.Setenv("Git_Work_Tree", tree)
	case "dir-only":
		t.Setenv("GIT_DIR", gitDir)
	default:
		t.Fatalf("unknown ambient kind %q", kind)
	}
}

func readOptionalFile(t *testing.T, p string) string {
	t.Helper()
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(data)
}

// recordedRootArg returns the value the wrapper passed after --root to the
// generated binary, as recorded by the shim's sentinel.
func recordedRootArg(t *testing.T, argsLog string) string {
	t.Helper()
	data, err := os.ReadFile(argsLog)
	if err != nil {
		t.Fatalf("read engine argument log: %v", err)
	}
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(data)), "\r\n", "\n"), "\n")
	for i := 0; i+1 < len(lines); i++ {
		if lines[i] == "--root" {
			return lines[i+1]
		}
	}
	t.Fatalf("engine was not handed --root: %q", data)
	return ""
}

// sameDirectory reports whether two directories are the same on disk. Both are
// resolved through Bash to a native path first, so 8.3 short names and
// POSIX-form paths compare equal (Windows-safe).
func sameDirectory(t *testing.T, bash, a, b string) bool {
	t.Helper()
	infoA, err := os.Stat(nativeDirForBash(t, bash, a))
	if err != nil {
		t.Fatalf("stat engine root %q: %v", a, err)
	}
	infoB, err := os.Stat(nativeDirForBash(t, bash, b))
	if err != nil {
		t.Fatalf("stat checkout %q: %v", b, err)
	}
	return os.SameFile(infoA, infoB)
}

func nativeDirForBash(t *testing.T, bash, dir string) string {
	t.Helper()
	out, err := exec.Command(bash, "-c", `cd "$1" && (pwd -W 2>/dev/null || pwd)`, "bash", dir).Output()
	if err != nil {
		t.Fatalf("resolve directory %q through bash: %v", dir, err)
	}
	return strings.TrimSpace(string(out))
}

// isPrivilegeNotHeld reports whether err is the Windows privilege refusal for
// symlink creation. It is always false off Windows.
func isPrivilegeNotHeld(err error) bool {
	var errno syscall.Errno
	return runtime.GOOS == "windows" && errors.As(err, &errno) && uintptr(errno) == errorPrivilegeNotHeld
}
