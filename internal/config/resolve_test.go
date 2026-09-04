package config

import "testing"

func TestWorkspaceRootUsesMatchingWorkspacePath(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", Path: "/explicit/path"},
		{WorkspaceID: "W2", Path: "/other/path"},
	}

	if got, want := cfg.WorkspaceRoot("W2"), "/other/path"; got != want {
		t.Errorf("WorkspaceRoot(\"W2\") = %q, want %q", got, want)
	}
}

func TestWorkspaceRootFallsBackToDefaultWhenMappingHasNoPath(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1"}}

	if got, want := cfg.WorkspaceRoot("W1"), "/default/root"; got != want {
		t.Errorf("WorkspaceRoot(\"W1\") = %q, want %q", got, want)
	}
}

func TestWorkspaceRootFallsBackToDefaultWhenWorkspaceUnknown(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", Path: "/explicit/path"}}

	if got, want := cfg.WorkspaceRoot("does-not-exist"), "/default/root"; got != want {
		t.Errorf("WorkspaceRoot(\"does-not-exist\") = %q, want %q", got, want)
	}
}

func TestWorkspaceRootSafeOnDefaultConfigWithNoMappings(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = "/default/root"

	if got, want := cfg.WorkspaceRoot("anything"), "/default/root"; got != want {
		t.Errorf("WorkspaceRoot on an empty Workspaces slice = %q, want %q", got, want)
	}
}
