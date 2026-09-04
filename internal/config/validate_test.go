package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// TestValidateRule1MaxConcurrentSessionsZero covers unit C1's acceptance
// criterion (i): MaxConcurrentSessions = 0 yields exactly
// "config: max_concurrent_sessions must be greater than zero".
func TestValidateRule1MaxConcurrentSessionsZero(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.MaxConcurrentSessions = 0
	cfg.HostCLI = "unused"

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for MaxConcurrentSessions == 0")
	}
	if got, want := err.Error(), "config: max_concurrent_sessions must be greater than zero"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

// TestValidateRule2aEmptyDefaultWorkspaceRootDoesNotResolveToCWD covers
// unit C1's acceptance criterion (ii): DefaultWorkspaceRoot = "" yields
// "config: default_workspace_root must be set" and does not resolve to
// the CWD.
func TestValidateRule2aEmptyDefaultWorkspaceRootDoesNotResolveToCWD(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = ""

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for an empty default_workspace_root")
	}
	if got, want := err.Error(), "config: default_workspace_root must be set"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
	if cfg.DefaultWorkspaceRoot != "" {
		t.Errorf("DefaultWorkspaceRoot = %q after a failed validation, want it left empty (not resolved to the CWD)", cfg.DefaultWorkspaceRoot)
	}
}

// TestValidateRule2bNonExistentRootYieldsInvalidMessage covers unit C1's
// acceptance criterion (iii): a non-existent root yields a message
// starting with "config: default_workspace_root invalid:".
func TestValidateRule2bNonExistentRootYieldsInvalidMessage(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = filepath.Join(t.TempDir(), "does-not-exist")

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for a non-existent default_workspace_root")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Validate error is not *apperr.Error: %v", err)
	}
	if !strings.HasPrefix(appErr.Error(), "config: default_workspace_root invalid:") {
		t.Errorf("Validate error = %q, want prefix %q", appErr.Error(), "config: default_workspace_root invalid:")
	}
}

// TestValidateRule2bValidRootCanonicalizesWithoutUNCPrefix covers unit
// C1's acceptance criterion (iv): a valid t.TempDir() root is rewritten to
// its canonicalized absolute form with no \\?\ prefix.
func TestValidateRule2bValidRootCanonicalizesWithoutUNCPrefix(t *testing.T) {
	dir := t.TempDir()
	cfg := Default()
	cfg.DefaultWorkspaceRoot = dir

	_, err := cfg.Validate()
	if err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
	if !filepath.IsAbs(cfg.DefaultWorkspaceRoot) {
		t.Errorf("DefaultWorkspaceRoot = %q, want an absolute path", cfg.DefaultWorkspaceRoot)
	}
	if strings.HasPrefix(cfg.DefaultWorkspaceRoot, `\\?\`) {
		t.Errorf("DefaultWorkspaceRoot = %q, want no \\\\?\\ UNC prefix", cfg.DefaultWorkspaceRoot)
	}
}

// TestValidateOnlyMutatesReceiverAfterFullSuccess is an additional guard
// (beyond the specific rule-6 forward reference in the plan's C1
// acceptance (v), verified once rule 6 exists in hostcli_test.go): a
// config that fails rule 1 (the very first rule) leaves
// DefaultWorkspaceRoot completely untouched, proving the deferred-commit
// discipline holds even when the failure occurs before the canonicalization
// step runs at all.
func TestValidateOnlyMutatesReceiverAfterFullSuccess(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	original := cfg.DefaultWorkspaceRoot
	cfg.MaxConcurrentSessions = 0

	if _, err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for MaxConcurrentSessions == 0")
	}
	if cfg.DefaultWorkspaceRoot != original {
		t.Errorf("DefaultWorkspaceRoot = %q after a failed validation, want unchanged %q", cfg.DefaultWorkspaceRoot, original)
	}
}
