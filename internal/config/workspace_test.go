package config

import "testing"

// newValidBaseConfig returns a Default()-based config with a valid
// canonicalizable DefaultWorkspaceRoot and a non-empty HostCLI (a bare
// name, so rule 7's absolute-path-exists check never applies), suitable
// as a base for mapping-stage and channel-uniqueness tests that must
// reach rules 3-5/8 without failing rule 1/2a/2b/6 first.
func newValidBaseConfig(t *testing.T) *Config {
	t.Helper()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.HostCLI = "intercom-host-cli-placeholder"
	return cfg
}

// TestValidateRule3EmptyWorkspaceID covers unit C2's acceptance criterion
// (i): an empty workspace_id yields the exact rule-3 message.
func TestValidateRule3EmptyWorkspaceID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "", ChannelID: "C1"}}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for an empty workspace_id")
	}
	if got, want := err.Error(), "config: workspace_id cannot be empty in [[workspace]] entry"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule4EmptyChannelID covers unit C2's acceptance criterion
// (ii): an empty channel_id yields the exact rule-4 message.
func TestValidateRule4EmptyChannelID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", ChannelID: ""}}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for an empty channel_id")
	}
	if got, want := err.Error(), "config: channel_id cannot be empty in [[workspace]] entry"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule5DuplicateWorkspaceID covers unit C2's acceptance
// criterion (iii): a duplicate workspace_id yields the exact rule-5
// message.
func TestValidateRule5DuplicateWorkspaceID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1"},
		{WorkspaceID: "W1", ChannelID: "C2"},
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a duplicate workspace_id")
	}
	if got, want := err.Error(), "config: duplicate workspace_id 'W1' in [[workspace]] entries"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule8DuplicateChannelID covers unit C2's acceptance
// criterion (iv): a duplicate channel_id yields the exact rule-8 message.
func TestValidateRule8DuplicateChannelID(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1"},
		{WorkspaceID: "W2", ChannelID: "C1"},
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a duplicate channel_id")
	}
	want := "config: duplicate channel_id 'C1' in [[workspace]] entries; ACP routing requires each channel_id to map to exactly one workspace"
	if got := err.Error(); got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}
