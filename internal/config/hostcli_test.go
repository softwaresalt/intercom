package config

import (
	"path/filepath"
	"testing"
)

// TestValidateRule6EmptyHostCLI covers unit C3's acceptance criterion
// (i): empty host_cli yields "config: host_cli must be set".
func TestValidateRule6EmptyHostCLI(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.HostCLI = ""

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for an empty host_cli")
	}
	if got, want := err.Error(), "config: host_cli must be set"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule7AbsoluteHostCLIDoesNotExist covers unit C3's
// acceptance criterion (ii): an absolute non-existent host_cli yields
// "config: host_cli '{path}' does not exist".
func TestValidateRule7AbsoluteHostCLIDoesNotExist(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist-binary")
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.HostCLI = missing

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a non-existent absolute host_cli")
	}
	want := "config: host_cli '" + missing + "' does not exist"
	if got := err.Error(); got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule7BareHostCLIUnresolvableProducesWarningOnly covers unit
// C3's acceptance criterion (iii): a bare name not resolvable by
// exec.LookPath produces no error and exactly one Report.Warnings entry.
func TestValidateRule7BareHostCLIUnresolvableProducesWarningOnly(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.HostCLI = "definitely-not-a-real-binary-xyz123"

	report, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error for an unresolvable bare host_cli: %v", err)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("len(report.Warnings) = %d, want 1", len(report.Warnings))
	}
}

// TestValidateRule1WinsOverRules6And8Simultaneously covers unit C3's
// acceptance criterion (iv): a config violating rules 1, 6, and 8
// simultaneously returns the rule-1 error, proving fixed, deterministic
// rule order.
func TestValidateRule1WinsOverRules6And8Simultaneously(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.MaxConcurrentSessions = 0 // violates rule 1
	cfg.HostCLI = ""              // violates rule 6
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1"},
		{WorkspaceID: "W2", ChannelID: "C1"}, // violates rule 8
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a config violating rules 1, 6, and 8")
	}
	if got, want := err.Error(), "config: max_concurrent_sessions must be greater than zero"; got != want {
		t.Errorf("Validate error = %q, want the rule-1 error %q (order determinism)", got, want)
	}
}

// TestValidateRule6FailureLeavesDefaultWorkspaceRootUnmodified covers the
// plan's C1 acceptance criterion (v), deferred here because it requires
// rule 6 (host_cli), which does not exist until this unit: a config
// failing rule 6 leaves DefaultWorkspaceRoot unmodified, proving the
// canonicalization commit remains deferred past rule 2b even when a later
// rule fails.
func TestValidateRule6FailureLeavesDefaultWorkspaceRootUnmodified(t *testing.T) {
	dir := t.TempDir()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = dir
	cfg.HostCLI = "" // violates rule 6, after rule 2b's canonicalization runs

	if _, err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for an empty host_cli")
	}
	if cfg.DefaultWorkspaceRoot != dir {
		t.Errorf("DefaultWorkspaceRoot = %q after a rule-6 failure, want unchanged %q (deferred commit)", cfg.DefaultWorkspaceRoot, dir)
	}
}

// TestValidateRule6WinsOverRule8 covers the plan's C2 acceptance
// criterion (v), deferred here because it requires rule 6, which does not
// exist until this unit: a config violating rule 6 and rule 8
// simultaneously returns the rule-6 error, proving host_cli validation
// runs before channel-uniqueness.
func TestValidateRule6WinsOverRule8(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.HostCLI = "" // violates rule 6
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1"},
		{WorkspaceID: "W2", ChannelID: "C1"}, // violates rule 8
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a config violating rules 6 and 8")
	}
	if got, want := err.Error(), "config: host_cli must be set"; got != want {
		t.Errorf("Validate error = %q, want the rule-6 error %q (host_cli runs before channel-uniqueness)", got, want)
	}
}
