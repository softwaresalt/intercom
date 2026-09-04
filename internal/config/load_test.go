package config

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestDecodeRejectsCaseFoldCollision(t *testing.T) {
	data := `default_workspace_root = "."
operator_detail_level = "minimal"
Operator_Detail_Level = "verbose"
`
	_, _, err := Decode(data)
	if err == nil {
		t.Fatal("Decode returned nil error for a case-fold key collision")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Decode error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Decode error kind = %v, want KindConfig", appErr.Kind())
	}
}

func TestDecodeRequiresDefaultWorkspaceRoot(t *testing.T) {
	_, _, err := Decode(`operator_detail_level = "standard"
`)
	if err == nil {
		t.Fatal("Decode returned nil error for a config omitting default_workspace_root")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Decode error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Decode error kind = %v, want KindConfig", appErr.Kind())
	}
}

func TestLoadNonExistentPathYieldsKindConfigWithGuidance(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")
	_, _, err := Load(missing)
	if err == nil {
		t.Fatal("Load returned nil error for a non-existent path")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Load error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Load error kind = %v, want KindConfig", appErr.Kind())
	}
	if !strings.Contains(appErr.Error(), missing) {
		t.Errorf("Load error = %q, want it to contain the path %q", appErr.Error(), missing)
	}
	if !strings.Contains(appErr.Error(), "--config") {
		t.Errorf("Load error = %q, want operator guidance text mentioning --config", appErr.Error())
	}
}

func TestLoadRejectsOversizedFileWithoutDecoding(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized.toml")
	oversized := make([]byte, MaxConfigBytes+1)
	for i := range oversized {
		oversized[i] = 'a'
	}
	if err := os.WriteFile(path, oversized, 0o600); err != nil {
		t.Fatalf("failed to write oversized fixture: %v", err)
	}

	_, _, err := Load(path)
	if err == nil {
		t.Fatal("Load returned nil error for an oversized config file")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Load error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Load error kind = %v, want KindConfig", appErr.Kind())
	}
}

func TestLoadDelegatesToDecodeOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	data := "default_workspace_root = \"" + escapeTOMLString(dir) + "\"\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load returned nil *Config")
	}
}

func escapeTOMLString(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

func TestDecodeDoesNotRejectCaseVariantMapEntriesAsCollisions(t *testing.T) {
	data := `default_workspace_root = "."
[commands]
build = "go build ./..."
Build = "make build"
`

	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error for distinct case-variant map keys: %v", err)
	}
	if got, want := len(cfg.Commands), 2; got != want {
		t.Errorf("len(Commands) = %d, want %d (both entries preserved)", got, want)
	}
	if cfg.Commands["build"] != "go build ./..." || cfg.Commands["Build"] != "make build" {
		t.Errorf("Commands = %v, want both distinct keys preserved verbatim", cfg.Commands)
	}
}

func TestDecodeRetiredNestedMapCaseFoldCollisionIsFatal(t *testing.T) {
	data := `default_workspace_root = "."
[slack.markdown_upload_extensions]
".md" = "text/markdown"
".MD" = "text/uppercase-markdown"
`

	_, _, err := Decode(data)
	if err == nil {
		t.Fatal("Decode returned nil error for a legacy slack.markdown_upload_extensions case-fold collision")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Decode error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Decode error kind = %v, want KindConfig", appErr.Kind())
	}
}

func TestDecodeStillRejectsCollisionOnMapFieldsOwnName(t *testing.T) {
	data := `default_workspace_root = "."
[commands]
build = "go build"

[Commands]
test = "go test"
`

	_, _, err := Decode(data)
	if err == nil {
		t.Fatal("Decode returned nil error for a case-fold collision on the map field's own name (commands vs Commands)")
	}
}

func TestFilterLeafKeysDoesNotMisclassifyByteWisePrefixWithoutSeparator(t *testing.T) {
	data := `default_workspace_root = "."
ab = 1
abc = 2
`

	_, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}

	want := []string{"ab", "abc"}
	if !slices.Equal(report.UnknownKeys, want) {
		t.Errorf("UnknownKeys = %v, want %v (both keys survive; neither is a segment-wise ancestor of the other)", report.UnknownKeys, want)
	}
}

func TestDecodeRequiredKeyPrecheckPrecedesValidate(t *testing.T) {
	data := `max_concurrent_sessions = 0
[copilot]
cli_path = "./bin/copilot"
`

	_, _, err := Decode(data)
	if err == nil {
		t.Fatal("Decode returned nil error for a config omitting default_workspace_root")
	}
	if got, want := err.Error(), "config: default_workspace_root must be set"; got != want {
		t.Errorf("Decode error = %q, want %q (required-key pre-check precedes Validate)", got, want)
	}
}
