package config

import (
	"os"
	"strings"
	"testing"
)

// configReferencePath locates docs/config-reference.md, which lives at the
// repository root — two directories above this package.
func configReferencePath(t *testing.T) string {
	t.Helper()
	return "../../docs/config-reference.md"
}

func TestConfigReferenceCoversDefaultFieldNames(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatalf("failed to read docs/config-reference.md: %v", err)
	}
	doc := string(data)

	fieldNames := []string{
		"Timeouts.ApprovalSeconds",
		"Timeouts.PromptSeconds",
		"Stall.Enabled",
		"Stall.InactivityThresholdSeconds",
		"Stall.EscalationThresholdSeconds",
		"Stall.MaxRetries",
		"Stall.DefaultNudgeMessage",
		"RetentionDays",
		"MaxConcurrentSessions",
		"HTTPPort",
		"Database.Path",
		"OperatorDetailLevel",
	}

	for _, name := range fieldNames {
		if !strings.Contains(doc, name) {
			t.Errorf("docs/config-reference.md is missing default field name %q", name)
		}
	}
}

func TestConfigReferenceCoversValidationMessages(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatalf("failed to read docs/config-reference.md: %v", err)
	}
	doc := string(data)

	messages := []string{
		"max_concurrent_sessions must be greater than zero",
		"default_workspace_root must be set",
		"default_workspace_root invalid:",
		"workspace_id cannot be empty in [[workspace]] entry",
		"duplicate workspace_id '{id}' in [[workspace]] entries",
		"copilot.cli_path '{path}' does not exist",
		"copilot.cli_path '{path}' must not be drive-relative; use an absolute path, a bare name, or empty",
		"copilot.cli_path '{path}' must not be relative; use an absolute path, a bare name, or empty",
		"workspace path invalid for workspace_id '{id}': {err}",
		"database.path must not contain '..' segments",
	}

	for _, msg := range messages {
		if !strings.Contains(doc, msg) {
			t.Errorf("docs/config-reference.md is missing validation message %q", msg)
		}
	}
}
