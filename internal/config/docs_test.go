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

// TestConfigReferenceCoversAll17DefaultFieldNames covers unit E2's
// acceptance criterion: every one of the 17 default field names appears
// verbatim in docs/config-reference.md, so doc completeness is enforced
// rather than assumed.
func TestConfigReferenceCoversAll17DefaultFieldNames(t *testing.T) {
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
		"ACP.MaxSessions",
		"ACP.StartupTimeoutSeconds",
		"ACP.MaxMsgRate",
		"ACP.HTTPPort",
		"HTTPPort",
		"IPCName",
		"Database.Path",
		"SlackDetailLevel",
	}
	if len(fieldNames) != 17 {
		t.Fatalf("test table declares %d field names, want 17", len(fieldNames))
	}

	for _, name := range fieldNames {
		if !strings.Contains(doc, name) {
			t.Errorf("docs/config-reference.md is missing default field name %q", name)
		}
	}
}

// TestConfigReferenceCoversAll10ValidationMessages covers unit E2's
// acceptance criterion: all 10 validation message strings appear
// verbatim in docs/config-reference.md.
func TestConfigReferenceCoversAll10ValidationMessages(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatalf("failed to read docs/config-reference.md: %v", err)
	}
	doc := string(data)

	// One entry per Validation Contract rule (rule 2's two sub-conditions,
	// 2a and 2b, are documented together as "rule 2" but both message
	// templates are checked individually below).
	messages := []string{
		"max_concurrent_sessions must be greater than zero",      // rule 1
		"default_workspace_root must be set",                     // rule 2a
		"default_workspace_root invalid:",                        // rule 2b
		"workspace_id cannot be empty in [[workspace]] entry",    // rule 3
		"channel_id cannot be empty in [[workspace]] entry",      // rule 4
		"duplicate workspace_id '{id}' in [[workspace]] entries", // rule 5
		"host_cli must be set",                                   // rule 6
		"host_cli '{path}' does not exist",                       // rule 7
		"duplicate channel_id '{id}' in [[workspace]] entries; ACP routing requires each channel_id to map to exactly one workspace", // rule 8
		"workspace path invalid for workspace_id '{id}': {err}",                                                                      // rule 9
		"database.path must not contain '..' segments",                                                                               // rule 10
	}
	if len(messages) != 11 {
		t.Fatalf("test table declares %d messages, want 11 (10 rules; rule 2 has two message templates)", len(messages))
	}

	for _, msg := range messages {
		if !strings.Contains(doc, msg) {
			t.Errorf("docs/config-reference.md is missing validation message %q", msg)
		}
	}
}
