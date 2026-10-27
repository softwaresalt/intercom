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

func TestValidateApprovalTimeoutUpperBound(t *testing.T) {
	cases := []struct {
		name    string
		seconds uint64
		wantErr string
	}{
		{"zero ok", 0, ""},
		{"default ok", 3600, ""},
		{"maximum ok", MaxApprovalTimeoutSeconds, ""},
		{"one over maximum rejected", MaxApprovalTimeoutSeconds + 1, "config: timeouts.approval_seconds must not exceed 86400"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newValidBaseConfig(t)
			cfg.Timeouts.ApprovalSeconds = tc.seconds

			_, err := cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate returned nil error for approval_seconds = %d", tc.seconds)
			}
			var appErr *apperr.Error
			if !errors.As(err, &appErr) {
				t.Fatalf("Validate error is not *apperr.Error: %v", err)
			}
			if appErr.Kind() != apperr.KindConfig {
				t.Errorf("Validate error kind = %v, want %v", appErr.Kind(), apperr.KindConfig)
			}
			if got := appErr.Error(); got != tc.wantErr {
				t.Errorf("Validate error = %q, want %q", got, tc.wantErr)
			}
		})
	}
}

func TestValidateApprovalTimeoutFailureLeavesReceiverUnmutated(t *testing.T) {
	cfg := newValidBaseConfig(t)
	wsDir := t.TempDir()
	cfg.Workspaces = []WorkspaceMapping{{WorkspaceID: "W1", Path: wsDir}}
	cfg.Timeouts.ApprovalSeconds = MaxApprovalTimeoutSeconds + 1
	rootBefore := cfg.DefaultWorkspaceRoot
	pathBefore := cfg.Workspaces[0].Path

	if _, err := cfg.Validate(); err == nil {
		t.Fatal("Validate returned nil error for approval_seconds over the maximum")
	}
	if cfg.DefaultWorkspaceRoot != rootBefore {
		t.Errorf("DefaultWorkspaceRoot = %q after a failed validation, want unchanged %q", cfg.DefaultWorkspaceRoot, rootBefore)
	}
	if cfg.Workspaces[0].Path != pathBefore {
		t.Errorf("Workspaces[0].Path = %q after a failed validation, want unchanged %q", cfg.Workspaces[0].Path, pathBefore)
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
