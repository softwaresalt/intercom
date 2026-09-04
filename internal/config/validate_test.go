package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestValidateRule1MaxConcurrentSessionsZero(t *testing.T) {
	cfg := Default()
	cfg.DefaultWorkspaceRoot = t.TempDir()
	cfg.MaxConcurrentSessions = 0

	_, err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate returned nil error for MaxConcurrentSessions == 0")
	}
	if got, want := err.Error(), "config: max_concurrent_sessions must be greater than zero"; got != want {
		t.Errorf("Validate error = %q, want %q", got, want)
	}
}

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
