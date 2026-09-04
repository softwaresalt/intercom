package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// TestDecodeRejectsCaseFoldCollision covers unit B2's acceptance criterion
// (i): a config defining both host_cli and Host_CLI yields KindConfig
// (divergence V2).
func TestDecodeRejectsCaseFoldCollision(t *testing.T) {
	data := "default_workspace_root = \".\"\nhost_cli = \"foo\"\nHost_CLI = \"bar\"\n"
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

// TestDecodeRequiresDefaultWorkspaceRoot covers unit B2's acceptance
// criterion (ii): a config omitting default_workspace_root yields
// KindConfig.
func TestDecodeRequiresDefaultWorkspaceRoot(t *testing.T) {
	_, _, err := Decode("host_cli = \"foo\"\n")
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

// TestLoadNonExistentPathYieldsKindConfigWithGuidance covers unit B2's
// acceptance criterion (iii): Load on a non-existent path yields
// KindConfig whose message contains the path and operator guidance text.
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

// TestLoadRejectsOversizedFileWithoutDecoding covers unit B2's acceptance
// criterion (iv): Load on a file of MaxConfigBytes+1 yields KindConfig
// without decoding (divergence V9).
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

// TestLoadDelegatesToDecodeOnSuccess proves Load's happy path reaches
// Decode: a small, valid config file loads successfully.
func TestLoadDelegatesToDecodeOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	data := "default_workspace_root = \"" + escapeTOMLString(dir) + "\"\nhost_cli = \"claude\"\n"
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

// escapeTOMLString escapes backslashes for embedding a Windows path inside
// a TOML basic string literal in a test fixture.
func escapeTOMLString(s string) string {
	return strings.ReplaceAll(s, `\`, `\\`)
}

// TestDecodeDoesNotRejectCaseVariantMapEntriesAsCollisions is a
// regression for a review finding: findCaseFoldCollision must not treat
// two differently-cased entries inside a map-typed field (Commands,
// Slack.MarkdownUploadExtensions) as an ambiguous struct-field collision.
// Unlike a struct field, a map's runtime keys are literal, case-sensitive
// data assigned directly into the map — "build" and "Build" are two
// distinct, unambiguous named commands, not a decode collision.
func TestDecodeDoesNotRejectCaseVariantMapEntriesAsCollisions(t *testing.T) {
	data := "default_workspace_root = \".\"\nhost_cli = \"claude\"\n" +
		"[commands]\nbuild = \"go build ./...\"\nBuild = \"make build\"\n"

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

// TestDecodeRequiredKeyPrecheckPrecedesValidate locks the documented
// order-of-checks caveat in Decode's doc comment: a config that both
// omits default_workspace_root and violates Validate's rule 1
// (max_concurrent_sessions == 0) returns Decode's own required-key
// message, not Validate's rule-1 message, because the required-key
// pre-check (Decode step 4) runs and returns before Validate is ever
// invoked (Decision R8).
func TestDecodeRequiredKeyPrecheckPrecedesValidate(t *testing.T) {
	data := "max_concurrent_sessions = 0\nhost_cli = \"claude\"\n"

	_, _, err := Decode(data)
	if err == nil {
		t.Fatal("Decode returned nil error for a config omitting default_workspace_root")
	}
	if got, want := err.Error(), "config: default_workspace_root must be set"; got != want {
		t.Errorf("Decode error = %q, want %q (required-key pre-check precedes Validate)", got, want)
	}
}
