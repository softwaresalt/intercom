package main

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

// Shim install cadence (057.009-T): every wrapper test installs the go shim
// once per top-level test, never once per subtest. The shim directory is the
// managed TMP itself, so a second install in the same test would make Git Bash
// spell the same directory two ways (/tmp and /c/...) and fail the self-check.
// Subtests that need isolated logs clear or address their own log files inside
// the one shim directory instead of reinstalling.

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

// installGoShim installs the go shim and returns the shim directory.
func installGoShim(t *testing.T, bash string) string {
	t.Helper()

	dir := workspaceTempDir(t)
	bashDir := resolveBashPath(t, bash, dir)
	shimPath := filepath.Join(dir, "go")
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
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("locate go shim: %v (output: %q)", err, output)
	}
	if got, want := strings.TrimSpace(string(output)), path.Join(bashDir, "go"); !strings.EqualFold(filepath.ToSlash(got), filepath.ToSlash(want)) {
		t.Fatalf("go resolves to %q, want test shim %q", got, want)
	}
	return dir
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
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("resolve Bash path for %q: %v (output: %q)", windowsOrNativePath, err, output)
	}
	return strings.TrimSpace(string(output))
}

// wrongTreeLayout selects how the wrapper checkout is laid out for one
// wrong-tree row (055.001-T, D-WB2-4).
type wrongTreeLayout int

const (
	// layoutOwnRepo: the checkout is a repository in its own right.
	layoutOwnRepo wrongTreeLayout = iota
	// layoutSymlink: the wrapper is invoked through a symlink to an own-repo checkout.
	layoutSymlink
	// layoutNoRepo: no .git exists at or above the checkout (GIT_CEILING_DIRECTORIES stops discovery at the row directory).
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
	name           string
	layout         wrongTreeLayout
	ambient        string
	subdir         string
	cwdDecoy       bool
	cwdOutside     bool
	readonlyGitDir bool
	// readonlyConfigGlobal sets an unexported readonly GIT_CONFIG_GLOBAL, which
	// the environment scrub never sees, so only the export refusal arm catches it.
	readonlyConfigGlobal bool
	// failCompgen shadows compgen with a function that fails, the named test-only
	// seam (exported into the wrapper through BASH_ENV) that forces a failed
	// environment listing in gatecheck_repo_root (057.008-T).
	failCompgen bool
	wantEngine  bool
	wantExit    int
	wantDiag    string
}

// writePathWrapperRel is the write-path wrapper, repo-relative.
const writePathWrapperRel = "scripts/check-write-path-precondition.sh"

// wrongTreeRows is the shared wrong-tree table (055.001-T, 055.003-T). Both
// gate wrappers that anchor through gatecheck_repo_root run these rows, so they
// cannot drift apart. At the red commit the decoy, mixed-case and nested rows
// are red for both wrappers, and so is the subdirectory row for the write-path
// wrapper; the subdirectory row is green on arrival for the retired-arch wrapper
// (regression guard). The control, symlink, no-repository, caller-outside and
// readonly rows are characterization rows, green on current code.
func wrongTreeRows() []wrongTreeRow {
	return []wrongTreeRow{
		{name: "decoy GIT_DIR and GIT_WORK_TREE", layout: layoutOwnRepo, ambient: "decoy", wantEngine: true},
		{name: "mixed-case ambient names", layout: layoutOwnRepo, ambient: "mixed-case", wantEngine: true},
		{name: "GIT_DIR alone from a subdirectory", layout: layoutOwnRepo, ambient: "dir-only", subdir: "internal", wantEngine: true},
		{name: "no ambient control", layout: layoutOwnRepo, wantEngine: true},
		{name: "symlinked checkout directory", layout: layoutSymlink, wantEngine: true},
		{name: "no repository", layout: layoutNoRepo, wantExit: 128, wantDiag: "not a git repository"},
		{name: "no repository, caller inside a decoy repository", layout: layoutNoRepo, cwdDecoy: true, wantExit: 128, wantDiag: "not a git repository"},
		{name: "caller outside any work tree", layout: layoutOwnRepo, cwdOutside: true, wantExit: 128, wantDiag: "not a git repository"},
		{name: "readonly GIT_DIR refuses in caller-precondition ordering", layout: layoutOwnRepo, readonlyGitDir: true, wantExit: 2, wantDiag: "cannot unset GIT_DIR"},
		{name: "unexported readonly GIT_CONFIG_GLOBAL refuses the config pin", layout: layoutOwnRepo, readonlyConfigGlobal: true, wantExit: 2, wantDiag: "cannot set GIT_CONFIG_GLOBAL"},
		{name: "failed compgen listing refuses before git runs", layout: layoutOwnRepo, failCompgen: true, wantExit: 2, wantDiag: "cannot list the environment"},
		{name: "nested checkout", layout: layoutNested, wantExit: 2, wantDiag: "is not the top level of its repository"},
	}
}

// runWrongTreeRow runs one row of the wrong-tree table against the wrapper at
// wrapperRel inside its own directory under base, the shim's base directory.
// The shim is installed once per test: the shim directory is the managed TMP
// itself, so a per-row install would make Git Bash spell the same directory
// two ways (/tmp and /c/...) and fail the self-check. Each row's logs are
// addressed by their /c/... form instead.
func runWrongTreeRow(t *testing.T, bash, base, wrapperRel string, row wrongTreeRow) {
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
	copyWrapperFixture(t, checkout, wrapperRel)
	script := filepath.Join(checkout, filepath.FromSlash(wrapperRel))
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
		script = filepath.Join(link, filepath.FromSlash(wrapperRel))
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
	if row.cwdDecoy {
		cwd = decoy
	}
	if row.cwdOutside {
		outside := filepath.Join(rowDir, "outside")
		if err := os.MkdirAll(outside, 0o755); err != nil {
			t.Fatalf("create caller directory outside any work tree: %v", err)
		}
		cwd = outside
		t.Setenv("GIT_CEILING_DIRECTORIES", filepath.ToSlash(rowDir))
	}
	if row.readonlyGitDir {
		bashEnv := filepath.Join(rowDir, "readonly-git-dir.sh")
		body := "export GIT_DIR='" + filepath.ToSlash(filepath.Join(decoy, ".git")) + "'\nreadonly GIT_DIR\n"
		if err := os.WriteFile(bashEnv, []byte(body), 0o644); err != nil {
			t.Fatalf("write readonly environment file: %v", err)
		}
		t.Setenv("BASH_ENV", toBashPath(bashEnv))
	}
	if row.readonlyConfigGlobal {
		bashEnv := filepath.Join(rowDir, "readonly-config-global.sh")
		// export -n drops any inherited export attribute so the variable is
		// unexported regardless of the test process environment.
		body := "export -n GIT_CONFIG_GLOBAL\nreadonly GIT_CONFIG_GLOBAL=/nonexistent/gitconfig\n"
		if err := os.WriteFile(bashEnv, []byte(body), 0o644); err != nil {
			t.Fatalf("write readonly environment file: %v", err)
		}
		t.Setenv("BASH_ENV", toBashPath(bashEnv))
	}
	if row.failCompgen {
		bashEnv := filepath.Join(rowDir, "fail-compgen.sh")
		body := "compgen() { return 1; }\nexport -f compgen\n"
		if err := os.WriteFile(bashEnv, []byte(body), 0o644); err != nil {
			t.Fatalf("write compgen seam file: %v", err)
		}
		t.Setenv("BASH_ENV", toBashPath(bashEnv))
	}

	trace, exitCode := runWritePathWrapperInDir(t, bash, toBashPath(script), cwd)
	marker := strings.TrimSpace(readOptionalFile(t, filepath.Join(rowDir, "generated-binary-calls.log")))

	if !row.wantEngine {
		if row.wantDiag == "" {
			t.Fatalf("fail-closed row %q names no diagnostic to assert", row.name)
		}
		if exitCode != row.wantExit {
			t.Errorf("exit code = %d, want %d for a checkout that must fail closed (trace=%q)", exitCode, row.wantExit, trace)
		}
		if !diagnosticOnOutputLine(trace, row.wantDiag) {
			t.Errorf("no output line carries the fail-closed diagnostic %q (trace=%q)", row.wantDiag, trace)
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

// copyWrapperFixture copies the live wrapper at wrapperRel and the shared runner
// into checkout, so the wrapper derives GATECHECK_SRC from the copy's location.
func copyWrapperFixture(t *testing.T, checkout, wrapperRel string) {
	t.Helper()
	repo := gatecheckRepoRoot(t)
	for _, rel := range []string{wrapperRel, "scripts/lib/gatecheck-run.sh"} {
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

// diagnosticOnOutputLine reports whether diag appears on a line the wrapper
// wrote, not on a bash -x trace line (trace lines start with "+"). Without this
// a diagnostic that reaches the trace only as an echoed command would pass with
// its redirect to stderr removed.
func diagnosticOnOutputLine(trace, diag string) bool {
	for _, line := range strings.Split(trace, "\n") {
		if strings.HasPrefix(line, "+") {
			continue
		}
		if strings.Contains(line, diag) {
			return true
		}
	}
	return false
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
