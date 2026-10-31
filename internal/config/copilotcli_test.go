package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRule5EmptyCLIPathValidWithZeroWarnings(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Copilot.CLIPath = ""

	report, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error for an empty cli_path: %v", err)
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("len(report.Warnings) = %d, want 0", len(report.Warnings))
	}
}

func TestValidateRule5AbsoluteCLIPathDoesNotExist(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist-binary")
	cfg := newValidBaseConfig(t)
	cfg.Copilot.CLIPath = missing

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a non-existent absolute cli_path")
	}
	want := "config: copilot.cli_path '" + missing + "' does not exist"
	if got := err.Error(); got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

func TestValidateRule5BareCLIPathUnresolvableProducesWarningOnly(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Copilot.CLIPath = "definitely-not-a-real-copilot-binary-xyz123"

	report, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error for an unresolvable bare cli_path: %v", err)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("len(report.Warnings) = %d, want 1", len(report.Warnings))
	}
	want := "copilot.cli_path 'definitely-not-a-real-copilot-binary-xyz123' was not resolvable via PATH"
	if got := report.Warnings[0]; got != want {
		t.Errorf("report.Warnings[0] = %q, want %q", got, want)
	}
}

func TestValidateRule5RejectsRelativeCLIPaths(t *testing.T) {
	tests := []struct {
		name    string
		cliPath string
		want    string
	}{
		{
			name:    "forward slash",
			cliPath: "./bin/copilot",
			want:    "config: copilot.cli_path './bin/copilot' must not be relative; use an absolute path, a bare name, or empty",
		},
		{
			name:    "backslash",
			cliPath: `..\x\copilot`,
			want:    `config: copilot.cli_path '..\x\copilot' must not be relative; use an absolute path, a bare name, or empty`,
		},
		{
			name:    "drive relative",
			cliPath: "C:copilot",
			want:    "config: copilot.cli_path 'C:copilot' must not be drive-relative; use an absolute path, a bare name, or empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newValidBaseConfig(t)
			cfg.Copilot.CLIPath = tt.cliPath

			_, err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate returned nil error for cli_path %q", tt.cliPath)
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("Validate error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateRule1WinsOverAllOtherViolations(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.MaxConcurrentSessions = 0
	cfg.Copilot.CLIPath = "./bin/copilot"
	cfg.Database.Path = "../data/agent-rc.db"
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", Path: filepath.Join(t.TempDir(), "does-not-exist")},
		{WorkspaceID: "W1"},
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a config violating multiple rules")
	}
	if got, want := err.Error(), "config: max_concurrent_sessions must be greater than zero"; got != want {
		t.Errorf("Validate error = %q, want the rule-1 error %q (order determinism)", got, want)
	}
}

func TestValidateRule5FailureLeavesDefaultWorkspaceRootUnmodified(t *testing.T) {
	dir := t.TempDir()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = dir
	cfg.Copilot.CLIPath = filepath.Join(t.TempDir(), "does-not-exist-binary")

	if _, err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for a non-existent absolute cli_path")
	}
	if cfg.DefaultWorkspaceRoot != dir {
		t.Errorf("DefaultWorkspaceRoot = %q after a rule-5 failure, want unchanged %q (deferred commit)", cfg.DefaultWorkspaceRoot, dir)
	}
}

func TestValidateRule5RejectsUNCAndDeviceCLIPaths(t *testing.T) {
	paths := []string{
		`\\server\share\copilot.exe`,
		"//server/share/copilot",
		`\\?\C:\copilot.exe`,
		`\\.\pipe\copilot`,
		`\??\C:\copilot.exe`,
	}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			cfg := newValidBaseConfig(t)
			cfg.Copilot.CLIPath = p

			_, err := cfg.Validate()
			if err == nil {
				t.Fatalf("Validate returned nil error for cli_path %q", p)
			}
			want := "config: copilot.cli_path '" + p + "' must not be a UNC or device-namespace path; use an absolute path, a bare name, or empty"
			if got := err.Error(); got != want {
				t.Errorf("Validate error = %q, want %q", got, want)
			}
		})
	}
}

func TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile(t *testing.T) {
	dir := t.TempDir()
	cfg := newValidBaseConfig(t)
	cfg.Copilot.CLIPath = dir

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a directory cli_path")
	}
	want := "config: copilot.cli_path '" + dir + "' must be a regular file"
	if got := err.Error(); got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}

	file := filepath.Join(t.TempDir(), "copilot")
	if err := os.WriteFile(file, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cfg = newValidBaseConfig(t)
	cfg.Copilot.CLIPath = file
	if _, err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned unexpected error for a regular-file cli_path: %v", err)
	}
	if cfg.Copilot.CLIPath != file {
		t.Errorf("CLIPath = %q, want %q", cfg.Copilot.CLIPath, file)
	}
}

func TestValidateRule5BareCLIPathResolvesToAbsWithDeferredCommit(t *testing.T) {
	binDir := t.TempDir()
	for _, name := range []string{"copilot", "copilot.exe"} {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	t.Setenv("PATH", binDir)

	cfg := newValidBaseConfig(t)
	cfg.Copilot.CLIPath = "copilot"
	report, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
	if len(report.Warnings) != 0 {
		t.Errorf("len(report.Warnings) = %d, want 0", len(report.Warnings))
	}
	if !filepath.IsAbs(cfg.Copilot.CLIPath) {
		t.Errorf("CLIPath = %q, want an absolute resolved path", cfg.Copilot.CLIPath)
	}

	// A failing later rule must leave CLIPath unmutated (deferred commit).
	cfg = newValidBaseConfig(t)
	cfg.Copilot.CLIPath = "copilot"
	cfg.Database.Path = "../escape.db"
	if _, err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for a database.path with '..'")
	}
	if cfg.Copilot.CLIPath != "copilot" {
		t.Errorf("CLIPath = %q after failed Validate, want it unmutated (%q)", cfg.Copilot.CLIPath, "copilot")
	}
}
