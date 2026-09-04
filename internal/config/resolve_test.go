package config

import "testing"

// TestResolveChannelIDKnownWorkspace covers unit D1's acceptance
// criterion (i): a known workspace resolves to its channel.
func TestResolveChannelIDKnownWorkspace(t *testing.T) {
	cfg := Default()
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1"},
		{WorkspaceID: "W2", ChannelID: "C2"},
	}

	got, ok := cfg.ResolveChannelID("W2")
	if !ok {
		t.Fatal("ResolveChannelID(\"W2\") returned ok=false, want true")
	}
	if got != "C2" {
		t.Errorf("ResolveChannelID(\"W2\") = %q, want %q", got, "C2")
	}
}

// TestResolveChannelIDUnknownWorkspace covers unit D1's acceptance
// criterion (ii): an unknown workspace returns false.
func TestResolveChannelIDUnknownWorkspace(t *testing.T) {
	cfg := Default()
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", ChannelID: "C1"}}

	if _, ok := cfg.ResolveChannelID("does-not-exist"); ok {
		t.Error("ResolveChannelID(\"does-not-exist\") returned ok=true, want false")
	}
}

// TestWorkspaceRootForChannelUsesMappingPathOrDefault covers unit D1's
// acceptance criterion (iii): a channel whose mapping sets path returns
// that path, and one without returns DefaultWorkspaceRoot.
func TestWorkspaceRootForChannelUsesMappingPathOrDefault(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1", Path: "/explicit/path"},
		{WorkspaceID: "W2", ChannelID: "C2"},
	}

	if got, want := cfg.WorkspaceRootForChannel("C1"), "/explicit/path"; got != want {
		t.Errorf("WorkspaceRootForChannel(\"C1\") = %q, want %q", got, want)
	}
	if got, want := cfg.WorkspaceRootForChannel("C2"), "/default/root"; got != want {
		t.Errorf("WorkspaceRootForChannel(\"C2\") = %q, want %q", got, want)
	}
}

// TestResolversSafeOnDefaultConfigWithNoMappings covers unit D1's
// acceptance criterion (iv): all three resolvers are safe on a Default()
// config with no mappings.
func TestResolversSafeOnDefaultConfigWithNoMappings(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"

	if _, ok := cfg.ResolveChannelID("anything"); ok {
		t.Error("ResolveChannelID on an empty Workspaces slice returned ok=true, want false")
	}
	if _, ok := cfg.ResolveWorkspaceByChannelID("anything"); ok {
		t.Error("ResolveWorkspaceByChannelID on an empty Workspaces slice returned ok=true, want false")
	}
	if got, want := cfg.WorkspaceRootForChannel("anything"), "/default/root"; got != want {
		t.Errorf("WorkspaceRootForChannel on an empty Workspaces slice = %q, want %q", got, want)
	}
}

// TestResolveWorkspaceByChannelIDReturnsValueNotPointer locks the
// by-value return contract: mutating the returned WorkspaceMapping must
// not affect the Config's underlying slice.
func TestResolveWorkspaceByChannelIDReturnsValueNotPointer(t *testing.T) {
	cfg := Default()
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", ChannelID: "C1", Label: "original"}}

	m, ok := cfg.ResolveWorkspaceByChannelID("C1")
	if !ok {
		t.Fatal("ResolveWorkspaceByChannelID(\"C1\") returned ok=false, want true")
	}
	m.Label = "mutated"

	if cfg.Workspaces[0].Label != "original" {
		t.Errorf("Workspaces[0].Label = %q after mutating the returned value, want unchanged %q", cfg.Workspaces[0].Label, "original")
	}
}
