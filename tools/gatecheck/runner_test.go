// Package main test coverage for M1-T9: the shared bash runner
// scripts/lib/gatecheck-run.sh (C-3). These tests never touch the real
// gatecheck binary build; every scenario runs bash against a temp copy of
// the runner with PATH-shim fakes standing in for `go` and the built
// gatecheck binary, per the plan's M1-T9 acceptance criteria.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// requireBash returns the path to a bash interpreter, or skips the test.
// This is the ONLY bash-availability gate the runner tests use.
func requireBash(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found on PATH; skipping runner tests (M1-T9)")
	}
	return path
}

// driveLetterRE matches a Windows absolute drive-letter path prefix
// (case-insensitive), after the path has already been slash-normalized.
var driveLetterRE = regexp.MustCompile(`^([A-Za-z]):/`)

// toBashPath converts a Windows drive-letter path into the MSYS POSIX form
// Git-Bash's PATH auto-translation expects when the value is composed into
// a colon-joined list (e.g. C:/Users/x -> /c/Users/x). It is a no-op on any
// input that is not a Windows drive-letter path, so it is safe to call
// unconditionally on every platform; on Linux/macOS paths never match the
// drive-letter pattern and pass through unchanged.
func toBashPath(p string) string {
	p = filepath.ToSlash(p)
	if m := driveLetterRE.FindStringSubmatch(p); m != nil {
		return "/" + strings.ToLower(m[1]) + "/" + p[len(m[0]):]
	}
	return p
}

// normalizeScript strips any CR so a script embedded in this .go source
// file (which may itself be checked out with CRLF line endings on a
// Windows/core.autocrlf tree) is always written to disk as pure LF. Bash
// tolerates CRLF inconsistently; every synthetic script this test writes is
// normalized before it touches the filesystem.
func normalizeScript(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// posixPwdToWindows resolves a bash-reported POSIX path (as produced by
// `cd ... && pwd` inside GATECHECK_SRC's own derivation) back to its
// Windows form using bash's own `pwd -W` builtin. This is required instead
// of a regex-based reverse conversion because Git-Bash mounts the Windows
// TEMP directory under a special /tmp alias rather than a plain
// drive-letter path, which only bash itself can resolve authoritatively.
func (f *runnerFixture) posixPwdToWindows(t *testing.T, posixPath string) string {
	t.Helper()
	out, err := exec.Command(f.bash, "-c", fmt.Sprintf("cd %q && pwd -W", posixPath)).Output()
	if err != nil {
		t.Fatalf("resolve posix path %q via pwd -W: %v", posixPath, err)
	}
	return filepath.FromSlash(strings.TrimSpace(string(out)))
}

// writeScript normalizes and writes an executable script to path.
func writeScript(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(normalizeScript(content)), 0o755); err != nil {
		t.Fatalf("write script %s: %v", path, err)
	}
}

// runnerFixture bundles the on-disk scaffolding every runner scenario
// shares: a temp copy of the real gatecheck-run.sh (so GATECHECK_SRC
// resolves to the fixture root, never the real repo), a fake `go` on PATH,
// and scratch directories for marker/log files the fakes write to.
type runnerFixture struct {
	bash        string
	root        string // GATECHECK_SRC once sourced: scripts/lib/gatecheck-run.sh lives under here
	runnerPath  string
	fakeBinDir  string
	markerDir   string
	logDir      string
	realBinSrc  string // the fake "built" gatecheck binary template, copied by the fake `go` build step
	goSrcLog    string
	goArgsLog   string
	binArgsLog  string
	tmpPathLog  string // wrapper writes $GATECHECK_TMP here right after gatecheck_build
	resumedFlag string // wrapper writes this file if execution resumes past a signal trap
}

const fakeGoTemplate = `#!/usr/bin/env bash
set -u
if [ -n "${GH_TOKEN:-}" ] || [ -n "${GITHUB_TOKEN:-}" ]; then
  : > "${MARKER_DIR}/go-token-leak"
fi
if [ "${GOTOOLCHAIN:-}" != "local" ] || [ "${GOFLAGS:-}" != "-mod=readonly" ]; then
  : > "${MARKER_DIR}/go-env-missing"
fi
if [ -n "${GO_ARGS_LOG:-}" ]; then
  printf '%s\n' "$*" >> "${GO_ARGS_LOG}"
fi
if [ "$1" = "env" ] && [ "$2" = "GOEXE" ]; then
  printf ''
  exit 0
fi
if [ "$1" = "-C" ]; then
  src="$2"
  if [ -n "${GO_SRC_LOG:-}" ]; then
    printf '%s\n' "$src" >> "${GO_SRC_LOG}"
  fi
  out=""
  prev=""
  for arg in "$@"; do
    if [ "$prev" = "-o" ]; then
      out="$arg"
    fi
    prev="$arg"
  done
  case "${FAKE_GO_BUILD_MODE:-success}" in
    fail)
      echo "fake compile error" >&2
      exit 1
      ;;
    *)
      cp "${FAKE_GO_BIN_SRC}" "${out}"
      chmod +x "${out}"
      exit 0
      ;;
  esac
fi
echo "unrecognized fake go invocation: $*" >&2
exit 99
`

const fakeBinTemplate = `#!/usr/bin/env bash
set -u
if [ -n "${GH_TOKEN:-}" ] || [ -n "${GITHUB_TOKEN:-}" ]; then
  : > "${MARKER_DIR}/bin-token-leak"
fi
if [ -n "${BIN_ARGS_LOG:-}" ]; then
  printf '%s\n' "$*" >> "${BIN_ARGS_LOG}"
fi
case "${FAKE_BIN_MODE:-exit}" in
  stream)
    printf 'OUT1\n'
    sleep 0.05
    printf 'ERR1\n' >&2
    sleep 0.05
    printf 'OUT2\n'
    sleep 0.05
    printf 'ERR2\n' >&2
    exit "${FAKE_BIN_EXIT:-0}"
    ;;
  sleep)
    sleep "${FAKE_BIN_SLEEP:-30}"
    exit "${FAKE_BIN_EXIT:-0}"
    ;;
  *)
    exit "${FAKE_BIN_EXIT:-0}"
    ;;
esac
`

func newRunnerFixture(t *testing.T) *runnerFixture {
	t.Helper()
	bash := requireBash(t)

	root := t.TempDir()
	libDir := filepath.Join(root, "scripts", "lib")
	if err := os.MkdirAll(libDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts/lib: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "scripts", "lib", "gatecheck-run.sh"))
	if err != nil {
		t.Fatalf("read real gatecheck-run.sh: %v", err)
	}
	runnerPath := filepath.Join(libDir, "gatecheck-run.sh")
	writeScript(t, runnerPath, string(raw))

	fakeBinDir := t.TempDir()
	writeScript(t, filepath.Join(fakeBinDir, "go"), fakeGoTemplate)

	scratch := t.TempDir()
	markerDir := filepath.Join(scratch, "markers")
	logDir := filepath.Join(scratch, "logs")
	for _, d := range []string{markerDir, logDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	realBinSrc := filepath.Join(scratch, "fakebin-template.sh")
	writeScript(t, realBinSrc, fakeBinTemplate)

	return &runnerFixture{
		bash:        bash,
		root:        root,
		runnerPath:  runnerPath,
		fakeBinDir:  fakeBinDir,
		markerDir:   markerDir,
		logDir:      logDir,
		realBinSrc:  realBinSrc,
		goSrcLog:    filepath.Join(logDir, "go-src.log"),
		goArgsLog:   filepath.Join(logDir, "go-args.log"),
		binArgsLog:  filepath.Join(logDir, "bin-args.log"),
		tmpPathLog:  filepath.Join(logDir, "gatecheck-tmp.log"),
		resumedFlag: filepath.Join(logDir, "resumed.flag"),
	}
}

// invokeStyle selects how the generated wrapper calls gatecheck_invoke, so
// the exit-code-passthrough AC can be proven "via gatecheck_invoke inside
// $(...) and || rc=$?" as well as directly.
type invokeStyle int

const (
	invokeDirect invokeStyle = iota
	invokeSubst
	invokeOrRC
)

// wrapperOpts configures the generated wrapper script for one scenario.
type wrapperOpts struct {
	subcommand  string
	invoke      invokeStyle
	ownCleanup  string // extra shell snippet composed into cleanup()
	afterInvoke string // extra shell snippet run after gatecheck_invoke returns (resumption probe)
	skipTmpLog  bool
	skipBuild   bool // for the double-cleanup-harmless scenario
	rawBody     string
}

func (f *runnerFixture) writeWrapper(t *testing.T, dir string, opts wrapperOpts) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("set -u\n")
	// Prepend the fake bin dir ahead of bash's OWN (already MSYS-translated)
	// $PATH. FAKE_BIN_DIR_POSIX is supplied already in POSIX form by the Go
	// harness (toBashPath); joining it onto bash's own $PATH with ':' here
	// — instead of pre-splicing a mixed-format PATH string before bash ever
	// starts — is what actually resolves correctly under Git-Bash on
	// Windows (verified empirically: pre-splicing a POSIX prefix onto a
	// still-untranslated Windows-style PATH before process start does not
	// resolve, because MSYS's translation runs once at process start over
	// the whole raw value).
	b.WriteString("export PATH=\"${FAKE_BIN_DIR_POSIX}:${PATH}\"\n")
	fmt.Fprintf(&b, "source %q\n", f.runnerPath)

	if opts.rawBody != "" {
		b.WriteString(opts.rawBody)
		wrapperPath := filepath.Join(dir, "wrapper.sh")
		writeScript(t, wrapperPath, b.String())
		return wrapperPath
	}

	b.WriteString("cleanup() {\n")
	if opts.ownCleanup != "" {
		b.WriteString("  " + opts.ownCleanup + "\n")
	}
	b.WriteString("  gatecheck_cleanup\n")
	b.WriteString("}\n")
	b.WriteString("trap cleanup EXIT\n")
	b.WriteString("trap 'exit 130' INT\n")
	b.WriteString("trap 'exit 143' TERM\n\n")

	if !opts.skipBuild {
		b.WriteString("gatecheck_build\n")
		b.WriteString("build_rc=$?\n")
		b.WriteString("if [ \"${build_rc}\" -ne 0 ]; then\n  exit \"${build_rc}\"\nfi\n")
		if !opts.skipTmpLog {
			fmt.Fprintf(&b, "printf '%%s' \"${GATECHECK_TMP:-}\" > %q\n", f.tmpPathLog)
		}
	}

	sub := opts.subcommand
	if sub == "" {
		sub = "write-path"
	}

	switch opts.invoke {
	case invokeSubst:
		fmt.Fprintf(&b, "out=$(gatecheck_invoke %s \"$@\")\nrc=$?\nprintf '%%s' \"$out\"\n", sub)
	case invokeOrRC:
		fmt.Fprintf(&b, "rc=0\ngatecheck_invoke %s \"$@\" || rc=$?\n", sub)
	default:
		fmt.Fprintf(&b, "gatecheck_invoke %s \"$@\"\nrc=$?\n", sub)
	}

	if opts.afterInvoke != "" {
		b.WriteString(opts.afterInvoke + "\n")
	}
	b.WriteString("exit \"${rc}\"\n")

	wrapperPath := filepath.Join(dir, "wrapper.sh")
	writeScript(t, wrapperPath, b.String())
	return wrapperPath
}

// baseEnv returns the environment every fixture-driven wrapper invocation
// needs: the fake `go` prepended to PATH (in MSYS POSIX form so Git-Bash's
// PATH auto-translation does not mis-split a Windows drive-letter colon),
// plus the marker/log/build-mode plumbing the fakes read.
func (f *runnerFixture) baseEnv(t *testing.T, extra map[string]string) []string {
	t.Helper()
	// PATH itself is left exactly as inherited (os.Environ()'s ambient
	// value). Git-Bash auto-translates that whole raw value to MSYS POSIX
	// form once at process start; the fake bin dir is prepended onto
	// bash's OWN already-translated $PATH from INSIDE the generated
	// wrapper script instead (see writeWrapper), which is the form that
	// actually resolves. FAKE_BIN_DIR_POSIX carries the pre-converted
	// value that line needs.
	env := append([]string{}, os.Environ()...)
	env = append(env,
		"FAKE_BIN_DIR_POSIX="+toBashPath(f.fakeBinDir),
		"ROOT="+f.root,
		"MARKER_DIR="+f.markerDir,
		"GO_SRC_LOG="+f.goSrcLog,
		"GO_ARGS_LOG="+f.goArgsLog,
		"BIN_ARGS_LOG="+f.binArgsLog,
		"FAKE_GO_BIN_SRC="+f.realBinSrc,
	)
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

func (f *runnerFixture) run(t *testing.T, wrapperPath string, args []string, extraEnv map[string]string, cwd string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	cmdArgs := append([]string{wrapperPath}, args...)
	cmd := exec.Command(f.bash, cmdArgs...)
	cmd.Env = f.baseEnv(t, extraEnv)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	code := 0
	var exitErr *exec.ExitError
	if runErr != nil {
		if errors.As(runErr, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			return outBuf.String(), errBuf.String(), -1, runErr
		}
	}
	return outBuf.String(), errBuf.String(), code, nil
}

func requireFileAbsent(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s: unexpected marker/file present at %s", what, path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("%s: stat %s: %v", what, path, err)
	}
}

func requireNoLeak(t *testing.T, f *runnerFixture) {
	t.Helper()
	requireFileAbsent(t, filepath.Join(f.markerDir, "go-token-leak"), "GH_TOKEN/GITHUB_TOKEN leaked to fake go")
	requireFileAbsent(t, filepath.Join(f.markerDir, "bin-token-leak"), "GH_TOKEN/GITHUB_TOKEN leaked to fake binary")
	requireFileAbsent(t, filepath.Join(f.markerDir, "go-env-missing"), "GOTOOLCHAIN=local / GOFLAGS=-mod=readonly missing for fake go")
}

// ---------------------------------------------------------------------
// Red tests (build failure, exit-code passthrough) committed first per AC.
// ---------------------------------------------------------------------

func TestGatecheckBuild_Failure_WrapperExits2WithErrorLine(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{})

	stdout, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_GO_BUILD_MODE": "fail",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "::error::gatecheck build failed") {
		t.Fatalf("stderr missing ::error:: build-failure line: %q", stderr)
	}
	requireNoLeak(t, f)
}

func TestGatecheckInvoke_ExitCodePassthrough(t *testing.T) {
	for _, wantExit := range []int{0, 1, 2, 3} {
		wantExit := wantExit
		for _, style := range []struct {
			name  string
			style invokeStyle
		}{
			{"direct", invokeDirect},
			{"command-substitution", invokeSubst},
			{"or-rc", invokeOrRC},
		} {
			style := style
			t.Run(fmt.Sprintf("exit-%d/%s", wantExit, style.name), func(t *testing.T) {
				f := newRunnerFixture(t)
				dir := t.TempDir()
				wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: style.style})

				_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
					"FAKE_BIN_MODE": "exit",
					"FAKE_BIN_EXIT": strconv.Itoa(wantExit),
				}, "")
				if err != nil {
					t.Fatalf("run wrapper: %v", err)
				}
				if code != wantExit {
					t.Fatalf("exit code = %d, want %d; stderr=%q", code, wantExit, stderr)
				}
				requireNoLeak(t, f)
			})
		}
	}
}

// ---------------------------------------------------------------------
// Streaming: stdout/stderr are not buffered or reordered by the runner.
// ---------------------------------------------------------------------

// orderedWriter records each Write call's tag and payload into a
// shared, mutex-protected sequence so interleaved stdout/stderr writes from
// the fake binary can be checked for ordering after the process exits.
type orderedWriter struct {
	mu    *sync.Mutex
	order *[]string
	tag   string
}

func (w *orderedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	*w.order = append(*w.order, w.tag+":"+string(p))
	return len(p), nil
}

func TestGatecheckInvoke_StreamsNotBuffered(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: invokeDirect})

	cmd := exec.Command(f.bash, wrapper)
	cmd.Env = f.baseEnv(t, map[string]string{
		"FAKE_BIN_MODE": "stream",
		"FAKE_BIN_EXIT": "0",
	})

	var mu sync.Mutex
	var order []string
	cmd.Stdout = &orderedWriter{mu: &mu, order: &order, tag: "OUT"}
	cmd.Stderr = &orderedWriter{mu: &mu, order: &order, tag: "ERR"}

	if err := cmd.Run(); err != nil {
		t.Fatalf("run wrapper: %v", err)
	}

	joined := strings.Join(order, "|")
	// The fake binary sleeps 50ms between each write; if the runner buffered
	// or captured output instead of streaming it directly through, the
	// relative arrival order recorded here would collapse to a single
	// end-of-process burst instead of the interleaved OUT/ERR/OUT/ERR
	// sequence the fake binary actually emits.
	wantPrefixes := []string{"OUT:OUT1\n", "ERR:ERR1\n", "OUT:OUT2\n", "ERR:ERR2\n"}
	if len(order) != len(wantPrefixes) {
		t.Fatalf("recorded %d writes, want %d; joined=%q", len(order), len(wantPrefixes), joined)
	}
	for i, want := range wantPrefixes {
		if order[i] != want {
			t.Fatalf("write[%d] = %q, want %q; full order=%q", i, order[i], want, joined)
		}
	}
	requireNoLeak(t, f)
}

// ---------------------------------------------------------------------
// GH_TOKEN / GITHUB_TOKEN absence, GOTOOLCHAIN / GOFLAGS presence.
// ---------------------------------------------------------------------

func TestGatecheckBuildAndInvoke_TokensStrippedEnvPresent(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: invokeDirect})

	_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "0",
		"GH_TOKEN":      "leaked-gh-token",
		"GITHUB_TOKEN":  "leaked-github-token",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	// The whole point of this scenario: GH_TOKEN/GITHUB_TOKEN were present
	// in the AMBIENT environment (set above), yet neither fake observed
	// them, and both fakes observed GOTOOLCHAIN=local/GOFLAGS=-mod=readonly.
	requireNoLeak(t, f)
}

// ---------------------------------------------------------------------
// Root anchoring: a different working directory still builds from
// GATECHECK_SRC and passes --root through unchanged.
// ---------------------------------------------------------------------

func TestGatecheckBuild_DifferentWorkingDirectory_StillAnchorsAndPassesRoot(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: invokeDirect})

	elsewhere := t.TempDir() // unrelated to the fixture root entirely

	_, stderr, code, err := f.run(t, wrapper, []string{"--self-test"}, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "0",
	}, elsewhere)
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}

	srcLog, err := os.ReadFile(f.goSrcLog)
	if err != nil {
		t.Fatalf("read go-src.log: %v", err)
	}
	gotSrc := strings.TrimSpace(string(srcLog))
	wantSrc := f.root
	if runtime.GOOS == "windows" {
		// gotSrc came from bash's own `cd ... && pwd` inside
		// GATECHECK_SRC's derivation, so it is in MSYS POSIX form (and,
		// for paths under the Windows TEMP directory, may use MSYS's
		// special /tmp mount alias rather than a plain drive-letter
		// conversion). Convert it back to a Windows path via bash's own
		// `pwd -W` rather than re-deriving the mapping in Go, since the
		// /tmp alias is a Git-Bash-specific mount that a regex-based
		// converter cannot reproduce.
		gotWindows := f.posixPwdToWindows(t, gotSrc)
		// Windows may report the same directory using either the short
		// (8.3) or long form of a path segment (observed for the current
		// user's profile directory on this box); compare by filesystem
		// identity rather than string equality so that alias form does
		// not produce a false failure.
		gotInfo, err := os.Stat(gotWindows)
		if err != nil {
			t.Fatalf("stat resolved windows path %q: %v", gotWindows, err)
		}
		wantInfo, err := os.Stat(wantSrc)
		if err != nil {
			t.Fatalf("stat want src %q: %v", wantSrc, err)
		}
		if !os.SameFile(gotInfo, wantInfo) {
			t.Fatalf("go -C src = %q (windows: %q), want %q (cwd was %q)", gotSrc, gotWindows, wantSrc, elsewhere)
		}
	} else {
		// On POSIX platforms GATECHECK_SRC and f.root are both already
		// native paths; bash's own `cd ... && pwd` may still resolve a
		// symlinked temp directory (e.g. macOS /tmp -> /private/tmp), so
		// compare by filesystem identity here too rather than raw string
		// equality.
		gotInfo, err := os.Stat(gotSrc)
		if err != nil {
			t.Fatalf("stat go -C src %q: %v", gotSrc, err)
		}
		wantInfo, err := os.Stat(wantSrc)
		if err != nil {
			t.Fatalf("stat want src %q: %v", wantSrc, err)
		}
		if !os.SameFile(gotInfo, wantInfo) {
			t.Fatalf("go -C src = %q, want %q (cwd was %q)", gotSrc, wantSrc, elsewhere)
		}
	}

	binArgs, err := os.ReadFile(f.binArgsLog)
	if err != nil {
		t.Fatalf("read bin-args.log: %v", err)
	}
	if !strings.Contains(string(binArgs), "--root "+f.root) {
		t.Fatalf("fake binary args %q do not contain --root %s", string(binArgs), f.root)
	}
	requireNoLeak(t, f)
}

// ---------------------------------------------------------------------
// $GATECHECK_TMP removal: normal exit, non-zero exit, INT, TERM. The
// script must not resume after a signal.
// ---------------------------------------------------------------------

func TestGatecheckCleanup_RemovesTmpOnNormalExit(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: invokeDirect})

	_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "0",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	assertTmpRemoved(t, f)
}

func TestGatecheckCleanup_RemovesTmpOnNonZeroExit(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{invoke: invokeDirect})

	_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "1",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stderr=%q", code, stderr)
	}
	assertTmpRemoved(t, f)
}

func assertTmpRemoved(t *testing.T, f *runnerFixture) {
	t.Helper()
	logged, err := os.ReadFile(f.tmpPathLog)
	if err != nil {
		t.Fatalf("read tmp-path log: %v", err)
	}
	tmpPath := strings.TrimSpace(string(logged))
	if tmpPath == "" {
		t.Fatalf("GATECHECK_TMP was never recorded (build did not run?)")
	}
	if _, err := os.Stat(tmpPath); err == nil {
		t.Fatalf("GATECHECK_TMP %s still exists after exit", tmpPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", tmpPath, err)
	}
}

func TestGatecheckCleanup_Idempotent_CalledTwiceIsHarmless(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{
		rawBody: "gatecheck_build\n" +
			"gatecheck_cleanup\n" +
			"gatecheck_cleanup\n" +
			"exit 0\n",
	})

	_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "0",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("calling gatecheck_cleanup twice was not harmless: exit %d, stderr=%q", code, stderr)
	}
}

func TestGatecheckCleanup_HarmlessWhenNeverBuilt(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{
		rawBody: "gatecheck_cleanup\nexit 0\n",
	})

	_, stderr, code, err := f.run(t, wrapper, nil, nil, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("gatecheck_cleanup before any build was not harmless: exit %d, stderr=%q", code, stderr)
	}
}

func TestGatecheckCleanup_WrapperOwnCleanupComposes(t *testing.T) {
	f := newRunnerFixture(t)
	dir := t.TempDir()
	ownMarker := filepath.Join(f.logDir, "own-cleanup.marker")
	wrapper := f.writeWrapper(t, dir, wrapperOpts{
		invoke:     invokeDirect,
		ownCleanup: fmt.Sprintf(": > %q", ownMarker),
	})

	_, stderr, code, err := f.run(t, wrapper, nil, map[string]string{
		"FAKE_BIN_MODE": "exit",
		"FAKE_BIN_EXIT": "0",
	}, "")
	if err != nil {
		t.Fatalf("run wrapper: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr)
	}
	if _, err := os.Stat(ownMarker); err != nil {
		t.Fatalf("wrapper's own cleanup did not run (composition broken): %v", err)
	}
	assertTmpRemoved(t, f)
}

// ---------------------------------------------------------------------
// shellcheck: clean if available, absence recorded otherwise.
// ---------------------------------------------------------------------

func TestGatecheckRunSh_ShellcheckCleanIfAvailable(t *testing.T) {
	path, err := exec.LookPath("shellcheck")
	if err != nil {
		t.Log("shellcheck not found on PATH; recording its absence per M1-T9 AC (not a failure)")
		return
	}
	target := filepath.Join("..", "..", "scripts", "lib", "gatecheck-run.sh")
	cmd := exec.Command(path, target)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("shellcheck reported issues in gatecheck-run.sh:\n%s", out.String())
	}
}

// ---------------------------------------------------------------------
// Signal handling: INT -> 130, TERM -> 143, cleanup runs, no resumption.
// See runner_signal_unix_test.go / runner_signal_windows_test.go for the
// OS-specific process-group helpers these tests call.
// ---------------------------------------------------------------------

const (
	sigINT  = 2
	sigTERM = 15
)

func runSignalScenario(t *testing.T, f *runnerFixture, signum int, wantExit int) {
	t.Helper()
	dir := t.TempDir()
	wrapper := f.writeWrapper(t, dir, wrapperOpts{
		invoke:      invokeDirect,
		afterInvoke: fmt.Sprintf(": > %q", f.resumedFlag),
	})

	cmd := exec.Command(f.bash, wrapper)
	cmd.Env = f.baseEnv(t, map[string]string{
		"FAKE_BIN_MODE":  "sleep",
		"FAKE_BIN_SLEEP": "30",
	})
	setNewProcessGroup(cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("start wrapper: %v", err)
	}

	// Give the wrapper time to reach gatecheck_build + the sleeping fake
	// binary before delivering the signal.
	time.Sleep(400 * time.Millisecond)

	if err := signalProcessGroup(cmd, signum); err != nil {
		_ = cmd.Process.Kill()
		t.Fatalf("deliver signal %d to process group: %v", signum, err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case waitErr := <-done:
		code := 0
		var exitErr *exec.ExitError
		if waitErr != nil {
			if errors.As(waitErr, &exitErr) {
				code = exitErr.ExitCode()
			} else {
				t.Fatalf("wait wrapper: %v", waitErr)
			}
		}
		if code != wantExit {
			t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, wantExit, outBuf.String(), errBuf.String())
		}
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("wrapper did not exit within 10s after signal %d (script did not resume?) stdout=%q stderr=%q", signum, outBuf.String(), errBuf.String())
	}

	requireFileAbsent(t, f.resumedFlag, "script resumed past the signal trap")
	assertTmpRemoved(t, f)
}

func TestGatecheckSignals_INT_Exits130AndCleansUp(t *testing.T) {
	f := newRunnerFixture(t)
	if !signalTestsSupported() {
		t.Skip("process-group signal delivery is not supported on this platform")
	}
	runSignalScenario(t, f, sigINT, 130)
}

func TestGatecheckSignals_TERM_Exits143AndCleansUp(t *testing.T) {
	f := newRunnerFixture(t)
	if !signalTestsSupported() {
		t.Skip("process-group signal delivery is not supported on this platform")
	}
	runSignalScenario(t, f, sigTERM, 143)
}
