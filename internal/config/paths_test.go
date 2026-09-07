package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRule6NonExistentWorkspacePathInvalid(t *testing.T) {
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{
		{WorkspaceID: "W1", Path: filepath.Join(t.TempDir(), "does-not-exist")},
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

func TestValidateRule6ValidWorkspacePathCanonicalizes(t *testing.T) {
	wsDir := t.TempDir()
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", Path: wsDir}}

	_, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
	if !filepath.IsAbs(cfg.Workspaces[0].Path) {
		t.Errorf("Workspaces[0].Path = %q, want an absolute canonical path", cfg.Workspaces[0].Path)
	}
}

// TestValidateRule6WorkspacePathRejectsRegularFile (011.006-T item (c),
// resolves 8472E0A1 item (c)) pins rule 6's integration WIRING to
// pathsafe.NewRoot at the Validate() layer: a workspace path that points at
// an existing REGULAR FILE (not a directory) must be rejected. This is
// deliberately NOT a duplicate of pathsafe's own
// TestNewRootRejectsFilePathAsWorkspaceRoot (root_test.go), which pins
// NewRoot's own behavior in isolation -- this test instead pins that
// Validate() actually calls NewRoot for each non-empty [[workspace]].path
// and correctly propagates its rejection through the config-layer error
// message, rather than silently accepting the file.
func TestValidateRule6WorkspacePathRejectsRegularFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "not-a-directory.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create file workspace-path candidate: %v", err)
	}

	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", Path: filePath}}

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a workspace path pointing at a regular file")
	}
	want := "config: workspace path invalid for workspace_id 'W1': "
	if got := err.Error(); len(got) < len(want) || got[:len(want)] != want {
		t.Errorf("Validate error = %q, want prefix %q", got, want)
	}
}

func TestValidateRule7DatabasePathRejectsDotDotSegment(t *testing.T) {
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

func TestValidateRule7AbsoluteDatabasePathAccepted(t *testing.T) {
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

// TestValidateRule7FailureLeavesWorkspacePathUncommitted re-founds the
// deferred-commit invariant (Decision R10, Protected Invariant I6) for the
// per-entry Workspaces[i].Path commit specifically: rule 6 computes a
// canonicalized local for a valid workspace entry, and a later rule 7
// failure must discard that local rather than writing it back to the
// receiver. TestValidateOnlyMutatesReceiverAfterFullSuccess in
// validate_test.go covers the same invariant for DefaultWorkspaceRoot only;
// this test closes the parallel gap for Workspaces[i].Path, whose commit is
// a separate loop over a separate local slice in Validate().
func TestValidateRule7FailureLeavesWorkspacePathUncommitted(t *testing.T) {
	wsDir := t.TempDir()
	cfg := newValidBaseConfig(t)
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", Path: wsDir}}
	cfg.Database.Path = "../../etc/db"

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a database.path containing '..' segments")
	}
	if cfg.Workspaces[0].Path != wsDir {
		t.Errorf("Workspaces[0].Path = %q after a rule-7 failure, want unchanged %q (rule 6's canonicalization must not commit)", cfg.Workspaces[0].Path, wsDir)
	}
}
