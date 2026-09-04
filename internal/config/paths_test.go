package config

import (
	"path/filepath"
	"testing"
)

// TestValidateRule9NonExistentWorkspacePathInvalid covers unit C4's
// acceptance criterion (i): a [[workspace]] with a non-existent path
// yields "config: workspace path invalid for workspace_id '{id}': …".
func TestValidateRule9NonExistentWorkspacePathInvalid(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1", Path: filepath.Join(t.TempDir(), "does-not-exist")},
	}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a non-existent workspace path")
	}
	want := "config: workspace path invalid for workspace_id 'W1': "
	if got := err.Error(); len(got) < len(want) || got[:len(want)] != want {
		t.Errorf("Validate error = %q, want prefix %q", got, want)
	}
}

// TestValidateRule9ValidWorkspacePathCanonicalizes covers unit C4's
// acceptance criterion (ii): a [[workspace]] whose path is a valid
// t.TempDir() is rewritten to its canonical form on success.
func TestValidateRule9ValidWorkspacePathCanonicalizes(t *testing.T) {
	wsDir := t.TempDir()
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", ChannelID: "C1", Path: wsDir},
	}

	_, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
	if !filepath.IsAbs(cfg.Workspaces[0].Path) {
		t.Errorf("Workspaces[0].Path = %q, want an absolute canonical path", cfg.Workspaces[0].Path)
	}
}

// TestValidateRule10DatabasePathRejectsDotDotSegment covers unit C4's
// acceptance criterion (iii): database.path = "../../etc/db" yields
// "config: database.path must not contain '..' segments".
func TestValidateRule10DatabasePathRejectsDotDotSegment(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Database.Path = "../../etc/db"

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a database.path containing '..' segments")
	}
	if got, want := err.Error(), "config: database.path must not contain '..' segments"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule10AbsoluteDatabasePathAccepted covers unit C4's
// acceptance criterion (iv): an absolute database.path is accepted
// (explicit operator privilege) and left unmodified.
func TestValidateRule10AbsoluteDatabasePathAccepted(t *testing.T) {
	cfg := newValidBaseConfig(t)
	abs := filepath.Join(t.TempDir(), "data", "agent-rc.db")
	cfg.Database.Path = abs

	_, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error for an absolute database.path: %v", err)
	}
	if cfg.Database.Path != abs {
		t.Errorf("Database.Path = %q, want unchanged %q", cfg.Database.Path, abs)
	}
}
