package config

import (
	"fmt"
	"os"
	"reflect"
	"sort"
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

// TestConfigReferenceCoversEverySchemaKeyPath mechanically enforces
// schema-to-docs coupling: it walks Config's toml struct tags with the
// same collectTOMLKeyPaths helper example_test.go already uses to check
// config.toml.example, and asserts every resulting dotted key path appears
// backtick-quoted in docs/config-reference.md. Without this, a schema
// field could be added or removed without the reference doc's key tables
// being caught out of date by any test — only the hand-maintained
// fieldNames/messages lists above would need updating, and nothing forces
// that update to happen.
func TestConfigReferenceCoversEverySchemaKeyPath(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatalf("failed to read docs/config-reference.md: %v", err)
	}
	doc := string(data)

	all := collectTOMLKeyPaths(reflect.TypeOf(Config{}), "")
	allSet := make(map[string]struct{}, len(all))
	for _, p := range all {
		allSet[p] = struct{}{}
	}

	// Only leaf paths are checked against a literal backtick-quoted dotted
	// form: a struct- or slice-of-struct-typed section (e.g. "copilot",
	// "workspace") is rendered in docs/config-reference.md as a TOML table
	// header ("## `[copilot]`", "## `[[workspace]]`"), not as a bare
	// backtick-quoted identifier, so it is intentionally excluded here
	// rather than producing a false-positive gap. A map-typed field (e.g.
	// "commands") is rendered the same way, with a synthetic
	// "commands.<name>" placeholder key rather than the bare field name, so
	// it is exempted identically.
	mapFields := make(map[string]struct{})
	configType := reflect.TypeOf(Config{})
	for i := range configType.NumField() {
		field := configType.Field(i)
		tag := field.Tag.Get("toml")
		if tag != "" && tag != "-" && field.Type.Kind() == reflect.Map {
			mapFields[tag] = struct{}{}
		}
	}

	var leaves []string
	for _, p := range all {
		if _, isMap := mapFields[p]; isMap {
			continue
		}
		isPrefixOfAnother := false
		for other := range allSet {
			if other != p && strings.HasPrefix(other, p+".") {
				isPrefixOfAnother = true
				break
			}
		}
		if !isPrefixOfAnother {
			leaves = append(leaves, p)
		}
	}
	sort.Strings(leaves)

	var missing []string
	for _, path := range leaves {
		if !strings.Contains(doc, fmt.Sprintf("`%s`", path)) {
			missing = append(missing, path)
		}
	}
	if len(missing) != 0 {
		t.Errorf("docs/config-reference.md is missing backtick-quoted key path(s): %v", missing)
	}
}
