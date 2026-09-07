// Package integration exercises start.ps1's restored capabilities (sub-epic
// F, 011.016-T/011.017-T/011.018-T) via the same Go-shells-out-to-pwsh
// harness pattern established in output_path_guard_test.go -- not Pester
// (no Pester dependency exists or is added in this repository).
package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const startScriptTimeout = 30 * time.Second

// filterEnvKeys returns base with any entry whose key (case-insensitively,
// matching Windows environment variable semantics) matches a key in
// overrideKeys removed, so a subsequent append of that key is a true
// override rather than a possibly-shadowed duplicate.
func filterEnvKeys(base []string, overrideKeys map[string]string) []string {
	if len(overrideKeys) == 0 {
		return append([]string{}, base...)
	}
	filtered := make([]string, 0, len(base))
	for _, kv := range base {
		name, _, found := strings.Cut(kv, "=")
		if !found {
			filtered = append(filtered, kv)
			continue
		}
		skip := false
		for overrideKey := range overrideKeys {
			if strings.EqualFold(name, overrideKey) {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, kv)
		}
	}
	return filtered
}

// runStartScriptSnippet dot-sources the real start.ps1 (so every function
// defined in it is available, without triggering Invoke-StartScriptMain --
// see the `$MyInvocation.InvocationName -ne '.'` guard at the bottom of the
// file) and then runs snippet, which may call any function start.ps1
// defines. env, if non-nil, OVERRIDES (not merely appends to) the child
// process environment for each named key -- ambient values for these keys
// are filtered out of the inherited environment first (Windows environment
// variables are case-insensitive, and a naive append of a duplicate key
// cannot be relied on to win over an ambient dev-machine value such as a
// pre-set COPILOT_EXE_PATH or a PATH that already contains a real
// graphtor-docs/copilot installation).
func runStartScriptSnippet(t *testing.T, root, snippet string, env map[string]string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), startScriptTimeout)
	defer cancel()

	startPS1 := filepath.Join(root, "start.ps1")
	helperScript := filepath.Join(t.TempDir(), "invoke-start-snippet.ps1")
	helperContent := ". \"" + startPS1 + "\"\n" + snippet + "\n"
	if err := os.WriteFile(helperScript, []byte(helperContent), 0o600); err != nil {
		t.Fatalf("writing snippet helper script: %v", err)
	}

	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-File", helperScript)
	cmd.Dir = t.TempDir()
	cmd.Env = filterEnvKeys(os.Environ(), env)
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("start.ps1 snippet exceeded %s timeout; output so far:\n%s", startScriptTimeout, out)
	}
	return string(out), err
}

// writeStubExecutable writes a tiny pwsh-shebang-free "executable" (a
// simple pwsh script invoked directly by full path, matching how
// start.ps1's own functions accept an explicit exe path override) that
// records its own invocation (arguments) to a marker file and exits with
// exitCode.
func writeStubExecutable(t *testing.T, dir, name string, exitCode int, stdout string) string {
	t.Helper()
	path := filepath.Join(dir, name+".ps1")
	content := fmt.Sprintf(
		"param([Parameter(ValueFromRemainingArguments=$true)][string[]]$StubArgs)\n"+
			"Add-Content -LiteralPath '%s.invoked' -Value ($StubArgs -join ' ')\n"+
			"Write-Output '%s'\n"+
			"exit %d\n",
		strings.ReplaceAll(path, "'", "''"), stdout, exitCode,
	)
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("writing stub executable %q: %v", path, err)
	}
	return path
}

func TestGetEnabledSidecarsReturnsWorkspaceConfiguredSet(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)

	out, err := runStartScriptSnippet(t, root, "(Get-EnabledSidecars) -join ','", nil)
	if err != nil {
		t.Fatalf("Get-EnabledSidecars failed: %v\noutput:\n%s", err, out)
	}
	got := strings.TrimSpace(out)
	if got != "engram,graphtor-docs,backlogit" {
		t.Fatalf("Get-EnabledSidecars = %q, want %q", got, "engram,graphtor-docs,backlogit")
	}
}

// 011.016-T AC: "enabledSidecars gating is proven to suppress a sidecar
// whose command exists but whose capability pack is absent."
func TestSidecarGatingSuppressesDisabledSidecarEvenWhenCommandExists(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	stubDir := t.TempDir()
	stub := writeStubExecutable(t, stubDir, "backlogit-stub", 0, "")

	snippet := fmt.Sprintf(
		"Invoke-BacklogitSync -EnabledSidecars @() -BacklogitExePath '%s'",
		strings.ReplaceAll(stub, "'", "''"),
	)
	out, err := runStartScriptSnippet(t, root, snippet, nil)
	if err != nil {
		t.Fatalf("Invoke-BacklogitSync failed: %v\noutput:\n%s", err, out)
	}
	if _, statErr := os.Stat(stub + ".invoked"); statErr == nil {
		t.Fatalf("backlogit stub was invoked despite 'backlogit' being ABSENT from EnabledSidecars (gating did not suppress it)")
	}
}

func TestSidecarGatingRunsEnabledSidecar(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	stubDir := t.TempDir()
	stub := writeStubExecutable(t, stubDir, "backlogit-stub", 0, "")

	snippet := fmt.Sprintf(
		"Invoke-BacklogitSync -EnabledSidecars @('backlogit') -BacklogitExePath '%s'",
		strings.ReplaceAll(stub, "'", "''"),
	)
	out, err := runStartScriptSnippet(t, root, snippet, nil)
	if err != nil {
		t.Fatalf("Invoke-BacklogitSync failed: %v\noutput:\n%s", err, out)
	}
	if _, statErr := os.Stat(stub + ".invoked"); statErr != nil {
		t.Fatalf("backlogit stub was NOT invoked even though 'backlogit' IS in EnabledSidecars: %v\noutput:\n%s", statErr, out)
	}
}

// 011.016-T AC: restores the graphtor-docs.exe fallback under
// .graphtor/bin/, exercised behaviourally (not merely parsed).
func TestResolveGraphtorDocsCommandFallsBackToLocalInstall(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	graphtorBinDir := filepath.Join(scratchRoot, ".graphtor", "bin")
	if err := os.MkdirAll(graphtorBinDir, 0o755); err != nil {
		t.Fatalf("creating .graphtor/bin fixture dir: %v", err)
	}
	stubPath := filepath.Join(graphtorBinDir, "graphtor-docs.exe")
	if err := os.WriteFile(stubPath, []byte("stub"), 0o700); err != nil {
		t.Fatalf("writing graphtor-docs.exe stub: %v", err)
	}

	snippet := fmt.Sprintf(
		"$cmd = Resolve-GraphtorDocsCommand -ScriptRoot '%s'; if ($cmd) { Write-Output $cmd.Source } else { Write-Output 'NONE' }",
		strings.ReplaceAll(scratchRoot, "'", "''"),
	)
	// PATH override: this dev/CI machine may have a REAL graphtor-docs on
	// PATH (e.g. a global capability-pack install), which would otherwise
	// win over the fallback and make this test pass vacuously regardless
	// of whether the fallback logic works. A minimal PATH proves the
	// fallback is actually exercised.
	out, err := runStartScriptSnippet(t, root, snippet, map[string]string{
		"PATH": `C:\Windows\System32;C:\Windows`,
	})
	if err != nil {
		t.Fatalf("Resolve-GraphtorDocsCommand failed: %v\noutput:\n%s", err, out)
	}
	got := strings.TrimSpace(out)
	if got != stubPath {
		t.Fatalf("Resolve-GraphtorDocsCommand fallback = %q, want %q", got, stubPath)
	}
}

// 011.016-T AC: ai_tools.copilot_cli.exe_path forwarding, at the same
// precedence tier COPILOT_EXE_PATH would occupy, without overriding an
// operator's own already-set COPILOT_EXE_PATH.
func TestResolveCopilotExecutableForwardsConfiguredExePath(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)

	// This dev/CI machine may have COPILOT_EXE_PATH/COPILOT_EXE already
	// set ambiently -- explicitly clear both so the forwarding path (not
	// an ambient operator override) is what this test actually proves.
	out, err := runStartScriptSnippet(t, root, `Resolve-CopilotExecutable -ConfiguredExePath 'C:\configured\copilot.exe'`, map[string]string{
		"COPILOT_EXE_PATH": "",
		"COPILOT_EXE":      "",
	})
	if err != nil {
		t.Fatalf("Resolve-CopilotExecutable failed: %v\noutput:\n%s", err, out)
	}
	got := strings.TrimSpace(out)
	if got != `C:\configured\copilot.exe` {
		t.Fatalf("Resolve-CopilotExecutable = %q, want the configured exe_path forwarded", got)
	}
}

func TestResolveCopilotExecutableDoesNotOverrideOperatorEnv(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)

	out, err := runStartScriptSnippet(t, root, `Resolve-CopilotExecutable -ConfiguredExePath 'C:\configured\copilot.exe'`, map[string]string{
		"COPILOT_EXE_PATH": `C:\operator\copilot.exe`,
	})
	if err != nil {
		t.Fatalf("Resolve-CopilotExecutable failed: %v\noutput:\n%s", err, out)
	}
	got := strings.TrimSpace(out)
	if got != `C:\operator\copilot.exe` {
		t.Fatalf("Resolve-CopilotExecutable = %q, want the operator's own COPILOT_EXE_PATH to win", got)
	}
}

// 011.017-T AC: no loaded value is ever echoed, logged, or written to a
// transcript.
func TestImportDotEnvLocalNeverEchoesLoadedValues(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	envLocalPath := filepath.Join(scratchRoot, ".env.local")
	secretValue := "SUPER-SECRET-VALUE-011017T"
	if err := os.WriteFile(envLocalPath, []byte(fmt.Sprintf("SOME_LOADED_SECRET=%s\n", secretValue)), 0o600); err != nil {
		t.Fatalf("writing .env.local fixture: %v", err)
	}

	snippet := fmt.Sprintf(
		"Import-DotEnvLocal -EnvLocalPath '%s'; Write-Output 'done-no-echo'",
		strings.ReplaceAll(envLocalPath, "'", "''"),
	)
	out, err := runStartScriptSnippet(t, root, snippet, nil)
	if err != nil {
		t.Fatalf("Import-DotEnvLocal failed: %v\noutput:\n%s", err, out)
	}
	if strings.Contains(out, secretValue) {
		t.Fatalf("secret value was echoed in output: %s", out)
	}
	if !strings.Contains(out, "done-no-echo") {
		t.Fatalf("expected marker output not found: %s", out)
	}
}

// 011.017-T AC: the regex widening is justified against a concrete key
// set -- a real lowercase/mixed-case key is loaded correctly.
func TestImportDotEnvLocalAcceptsMixedCaseKeys(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	envLocalPath := filepath.Join(scratchRoot, ".env.local")
	if err := os.WriteFile(envLocalPath, []byte("Mixed_Case_Key=loaded-ok\n"), 0o600); err != nil {
		t.Fatalf("writing .env.local fixture: %v", err)
	}

	snippet := fmt.Sprintf(
		"Import-DotEnvLocal -EnvLocalPath '%s'; Write-Output $env:Mixed_Case_Key",
		strings.ReplaceAll(envLocalPath, "'", "''"),
	)
	out, err := runStartScriptSnippet(t, root, snippet, nil)
	if err != nil {
		t.Fatalf("Import-DotEnvLocal failed: %v\noutput:\n%s", err, out)
	}
	got := strings.TrimSpace(out)
	if got != "loaded-ok" {
		t.Fatalf("mixed-case key was not loaded: got %q, want %q (output: %s)", got, "loaded-ok", out)
	}
}

// 011.017-T AC: exit-code propagation is asserted concretely -- a non-zero
// Copilot exit produces a non-zero launcher exit.
func TestInvokeStartScriptMainPropagatesNonZeroExitCode(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	stubCopilot := writeStubExecutable(t, scratchRoot, "copilot-stub", 7, "stub copilot ran")

	// Run the FULL main entry point (not just a snippet) via
	// COPILOT_EXE_PATH pointing at the stub, with the sidecars all
	// resolving to nothing (no real backlogit/engram/graphtor-docs on
	// PATH in this constrained environment is fine -- absence is
	// non-fatal by design).
	ctx, cancel := context.WithTimeout(context.Background(), startScriptTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-File", filepath.Join(root, "start.ps1"))
	cmd.Dir = scratchRoot
	cmd.Env = append(os.Environ(),
		"COPILOT_EXE_PATH="+stubCopilot,
		"COPILOT_HOME="+filepath.Join(scratchRoot, ".copilot"),
		"ENGRAM_DATA_DIR="+filepath.Join(scratchRoot, ".engram"),
	)
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("start.ps1 exceeded %s timeout; output so far:\n%s", startScriptTimeout, out)
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected a non-zero exit (ExitError), got err=%v\noutput:\n%s", err, out)
	}
	if exitErr.ExitCode() != 7 {
		t.Fatalf("launcher exit code = %d, want 7 (propagated from the stub Copilot CLI)\noutput:\n%s", exitErr.ExitCode(), out)
	}
	if !strings.Contains(string(out), "stub copilot ran") {
		t.Fatalf("expected stub Copilot CLI output not found: %s", out)
	}
}

// 011.018-T AC: Push-Location/Pop-Location remain in try/finally -- the
// caller's PowerShell session location is restored even when the launched
// Copilot CLI stub exits non-zero (a mid-sequence failure).
func TestLocationRestoredEvenOnNonZeroCopilotExit(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	callerDir := t.TempDir()
	stubCopilot := writeStubExecutable(t, scratchRoot, "copilot-stub", 3, "")

	helperScript := filepath.Join(t.TempDir(), "invoke-and-check-location.ps1")
	helperContent := fmt.Sprintf(
		"Set-Location -LiteralPath '%s'\n"+
			"$before = (Get-Location).Path\n"+
			"$env:COPILOT_EXE_PATH = '%s'\n"+
			"& pwsh -NoProfile -NonInteractive -File '%s' 2>$null | Out-Null\n"+
			"$after = (Get-Location).Path\n"+
			"if ($before -ne $after) { Write-Output \"LOCATION_CHANGED:before=$before after=$after\"; exit 1 }\n"+
			"Write-Output 'LOCATION_UNCHANGED'\n",
		strings.ReplaceAll(callerDir, "'", "''"),
		strings.ReplaceAll(stubCopilot, "'", "''"),
		strings.ReplaceAll(filepath.Join(root, "start.ps1"), "'", "''"),
	)
	if err := os.WriteFile(helperScript, []byte(helperContent), 0o600); err != nil {
		t.Fatalf("writing location-check helper: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), startScriptTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-File", helperScript)
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("location-check helper exceeded %s timeout; output so far:\n%s", startScriptTimeout, out)
	}
	if err != nil {
		t.Fatalf("location-check helper failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(string(out), "LOCATION_UNCHANGED") {
		t.Fatalf("caller's PowerShell session location was not restored: %s", out)
	}
}

// 011.018-T AC: the "Generated by autoharness" marker is restored adjacent
// to an in-file "LOCAL DIVERGENCE" block enumerating the two improvements.
func TestGeneratedByAutoharnessMarkerAndLocalDivergenceBannerPresent(t *testing.T) {
	root := repoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "start.ps1"))
	if err != nil {
		t.Fatalf("reading start.ps1: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "Generated by autoharness") {
		t.Fatalf("start.ps1 is missing the 'Generated by autoharness' marker")
	}
	if !strings.Contains(text, "LOCAL DIVERGENCE") {
		t.Fatalf("start.ps1 is missing the 'LOCAL DIVERGENCE' banner")
	}
	markerIdx := strings.Index(text, "Generated by autoharness")
	bannerIdx := strings.Index(text, "LOCAL DIVERGENCE")
	if bannerIdx < markerIdx || bannerIdx-markerIdx > 500 {
		t.Fatalf("LOCAL DIVERGENCE banner is not adjacent to the Generated-by-autoharness marker (marker at %d, banner at %d)", markerIdx, bannerIdx)
	}
}

// 011.018-T AC: Invoke-EngramCommandWithProgress retains its bounded shared
// wall-clock budget and timeout kill.
func TestInvokeEngramCommandWithProgressEnforcesTimeout(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}
	root := repoRoot(t)
	scratchRoot := t.TempDir()
	// A stub engram "executable" that sleeps far longer than the budget
	// below, so the timeout-kill path is genuinely exercised.
	slowStub := filepath.Join(scratchRoot, "slow-engram.ps1")
	if err := os.WriteFile(slowStub, []byte("Start-Sleep -Seconds 30\n"), 0o700); err != nil {
		t.Fatalf("writing slow stub: %v", err)
	}

	snippet := fmt.Sprintf(
		"try {\n"+
			"  Invoke-EngramCommandWithProgress -Executable '%s' -Subcommand 'sync' -Activity 'test' -Status 'test' -Deadline ([DateTimeOffset]::UtcNow.AddMilliseconds(500))\n"+
			"  Write-Output 'NO_TIMEOUT_ERROR'\n"+
			"} catch {\n"+
			"  Write-Output \"TIMEOUT_CAUGHT:$_\"\n"+
			"}",
		strings.ReplaceAll(slowStub, "'", "''"),
	)
	out, err := runStartScriptSnippet(t, root, snippet, nil)
	if err != nil {
		t.Fatalf("snippet execution failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "TIMEOUT_CAUGHT") {
		t.Fatalf("expected the bounded wall-clock budget to be enforced (timeout caught), got: %s", out)
	}
}
