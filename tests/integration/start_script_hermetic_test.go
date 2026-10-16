package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// hermeticSidecarNames are the executables start.ps1's Get-EnabledSidecars
// can launch; hermeticStartEnv guarantees none is reachable via PATH.
var hermeticSidecarNames = []string{"backlogit", "engram", "graphtor-docs"}

// hermeticStartScript writes a byte copy of the real <root>/start.ps1 into a
// private t.TempDir and returns the copy's path, so $PSScriptRoot,
// .env.local, .copilot, .engram and .graphtor resolve in scratch space
// rather than the real repository (D-HS-1).
func hermeticStartScript(t *testing.T, root string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "start.ps1"))
	if err != nil {
		t.Fatalf("reading real start.ps1: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "start.ps1")
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		t.Fatalf("writing scratch start.ps1: %v", err)
	}
	return dst
}

// writeCwdEchoStub writes a copilot stub (a pwsh script, like
// writeStubExecutable, which stays untouched) that prints
// "stub copilot ran" and "STUB_CWD=<Get-Location>" and exits with exitCode.
// start.ps1 does Push-Location $PSScriptRoot before launching copilot, so
// STUB_CWD reveals which script directory the launcher really ran from.
func writeCwdEchoStub(t *testing.T, dir, name string, exitCode int) string {
	t.Helper()
	path := filepath.Join(dir, name+".ps1")
	content := fmt.Sprintf(
		"Write-Output 'stub copilot ran'\n"+
			"Write-Output (\"STUB_CWD=\" + (Get-Location).Path)\n"+
			"exit %d\n", exitCode)
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("writing cwd-echo stub %q: %v", path, err)
	}
	return path
}

// dirHasSidecar reports whether dir contains an executable named after any
// sidecar (bare name, .ps1, or any PATHEXT extension on Windows).
func dirHasSidecar(dir string) bool {
	exts := []string{"", ".ps1"}
	if runtime.GOOS == "windows" {
		pathExt := os.Getenv("PATHEXT")
		if pathExt == "" {
			pathExt = ".COM;.EXE;.BAT;.CMD"
		}
		for _, e := range strings.Split(pathExt, ";") {
			if e != "" {
				exts = append(exts, strings.ToLower(e))
			}
		}
	}
	for _, name := range hermeticSidecarNames {
		for _, ext := range exts {
			info, err := os.Stat(filepath.Join(dir, name+ext))
			if err != nil || info.IsDir() {
				continue
			}
			if runtime.GOOS == "windows" || info.Mode()&0o111 != 0 {
				return true
			}
		}
	}
	return false
}

// hermeticStartEnv returns an environment for running the scratch start.ps1
// with every PATH directory that contains a sidecar executable dropped
// (the scrub is deliberately coarse: a directory such as C:\Tools is removed
// whole), plus the caller's overrides, and the absolute path of pwsh
// resolved against the parent PATH (exec.Command("pwsh") would resolve with
// the parent PATH, not cmd.Env). If pwsh itself lives in a sidecar-bearing
// directory the test is skipped. Residual, accepted (D-HS-2): ambient
// GITHUB_* tokens and the read-only `gh auth token` call.
func hermeticStartEnv(t *testing.T, base []string, overrides map[string]string) ([]string, string) {
	t.Helper()
	pwshExe, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh not available on PATH")
	}
	if abs, err := filepath.Abs(pwshExe); err == nil {
		pwshExe = abs
	}
	if dirHasSidecar(filepath.Dir(pwshExe)) {
		t.Skipf("pwsh directory %q also holds a sidecar executable; cannot scrub PATH without losing pwsh", filepath.Dir(pwshExe))
	}

	keys := map[string]string{"PATH": ""}
	for k := range overrides {
		keys[k] = ""
	}
	env := filterEnvKeys(base, keys)

	var kept []string
	for _, dir := range filepath.SplitList(pathValue(base)) {
		if dir == "" || dirHasSidecar(dir) {
			continue
		}
		kept = append(kept, dir)
	}
	env = append(env, "PATH="+strings.Join(kept, string(os.PathListSeparator)))
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	return env, pwshExe
}

// pathValue returns the value of the PATH entry in env, matching the key
// case-insensitively (Windows spells it "Path").
func pathValue(env []string) string {
	for _, kv := range env {
		name, val, ok := strings.Cut(kv, "=")
		if ok && strings.EqualFold(name, "PATH") {
			return val
		}
	}
	return ""
}

// runPwshForGuard runs a real pwsh (absolute path, -NoProfile) on script
// with the given env and returns trimmed combined output.
func runPwshForGuard(t *testing.T, pwshExe string, env []string, script string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), startScriptTimeout)
	defer cancel()
	helper := filepath.Join(t.TempDir(), "guard.ps1")
	if err := os.WriteFile(helper, []byte(script), 0o600); err != nil {
		t.Fatalf("writing guard helper: %v", err)
	}
	cmd := exec.CommandContext(ctx, pwshExe, "-NoProfile", "-NonInteractive", "-File", helper)
	cmd.Dir = t.TempDir()
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("guard pwsh failed: %v\noutput:\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

// G1: the hermetic copy is a distinct scratch script, and the functions it
// defines resolve to the scratch directory, never the repository root.
func TestHermeticStartScriptIsolatesScriptRoot(t *testing.T) {
	pwshExe, err := exec.LookPath("pwsh")
	if err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratch := hermeticStartScript(t, root)
	realScript := filepath.Join(root, "start.ps1")
	if scratch == realScript {
		t.Fatalf("hermeticStartScript returned the real script path %q", scratch)
	}
	scratchDir := filepath.Dir(scratch)

	// Dot-sourcing defines functions without running main (see the
	// InvocationName guard at the bottom of start.ps1). The defining file
	// of a function's script block locates the script's $PSScriptRoot.
	script := ". '" + strings.ReplaceAll(scratch, "'", "''") + "'\n" +
		"Split-Path -Parent (Get-Command Get-EnabledSidecars).ScriptBlock.File\n"
	got := runPwshForGuard(t, pwshExe, os.Environ(), script)

	gotInfo, err := os.Stat(got)
	if err != nil {
		t.Fatalf("stat printed script root %q: %v", got, err)
	}
	scratchInfo, err := os.Stat(scratchDir)
	if err != nil {
		t.Fatalf("stat scratch dir: %v", err)
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		t.Fatalf("stat repo root: %v", err)
	}
	if !os.SameFile(gotInfo, scratchInfo) {
		t.Fatalf("script root %q is not the scratch dir %q", got, scratchDir)
	}
	if os.SameFile(gotInfo, rootInfo) {
		t.Fatalf("script root %q is the real repository root", got)
	}
}

// G2: after hermeticStartEnv no sidecar resolves, yet pwsh still does. A
// control run under the unscrubbed env proves injected fake sidecars are
// found, so the guard is non-vacuous on sidecar-free CI.
func TestHermeticStartEnvHidesSidecars(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	fakeDir := t.TempDir()
	for _, name := range hermeticSidecarNames {
		file := name
		if runtime.GOOS == "windows" {
			file += ".exe"
		}
		if err := os.WriteFile(filepath.Join(fakeDir, file), []byte("fake"), 0o755); err != nil {
			t.Fatalf("writing fake sidecar: %v", err)
		}
	}
	base := filterEnvKeys(os.Environ(), map[string]string{"PATH": ""})
	base = append(base, "PATH="+fakeDir+string(os.PathListSeparator)+pathValue(os.Environ()))

	const probe = "Get-Command backlogit,engram,graphtor-docs -ErrorAction SilentlyContinue | ForEach-Object { $_.Name }\n"

	env, pwshExe := hermeticStartEnv(t, base, nil)
	control := runPwshForGuard(t, pwshExe, base, probe)
	for _, name := range hermeticSidecarNames {
		if !strings.Contains(control, name) {
			t.Fatalf("control: %q not found under unscrubbed env (output %q); guard would be vacuous", name, control)
		}
	}

	for _, dir := range filepath.SplitList(pathValue(env)) {
		if dir == fakeDir {
			t.Fatalf("fake sidecar dir %q still on scrubbed PATH", fakeDir)
		}
	}
	if got := runPwshForGuard(t, pwshExe, env, probe); got != "" {
		t.Fatalf("sidecars still resolvable after scrub: %q", got)
	}
	if got := runPwshForGuard(t, pwshExe, env, "Get-Command pwsh | ForEach-Object { $_.Name }\n"); got == "" {
		t.Fatalf("pwsh no longer resolves on scrubbed PATH")
	}
}
