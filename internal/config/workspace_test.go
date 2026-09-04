package config

import "testing"

// newValidBaseConfig returns a Default()-based config with a valid
// canonicalizable DefaultWorkspaceRoot, suitable as a base for workspace-
// validation tests that must reach rules 3-7 without failing rule 1/2a/2b
// first.
func newValidBaseConfig(t *testing.T) *Config {
	t.Helper()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	return cfg
}

func TestValidateRule3EmptyWorkspaceID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: ""}}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for an empty workspace_id")
	}
	if got, want := err.Error(), "config: workspace_id cannot be empty in [[workspace]] entry"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

func TestValidateRule4DuplicateWorkspaceID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1"},
		{WorkspaceID: "W1"},
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a duplicate workspace_id")
	}
	if got, want := err.Error(), "config: duplicate workspace_id 'W1' in [[workspace]] entries"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}
