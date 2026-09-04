package config

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestDecodeReportsUnknownKeysTolerantly(t *testing.T) {
	data := `default_workspace_root = "."
host_cli = "claude"
wobble = 1

[nonsense]
x = 2
`
	cfg, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Decode returned nil *Config")
	}

	want := []string{"host_cli", "nonsense.x", "wobble"}
	if !slices.Equal(report.UnknownKeys, want) {
		t.Errorf("UnknownKeys = %v, want %v (sorted)", report.UnknownKeys, want)
	}
}

func TestDecodeReportsRetiredKeysFromShippedShapeConfig(t *testing.T) {
	data := `default_workspace_root = "."
host_cli = "claude"
host_cli_args = ["--dangerously-skip-permissions"]
ipc_name = "agent-intercom"
slack_detail_level = "standard"

[slack]
channel_id = "C0123456789"

[acp]
max_sessions = 5
startup_timeout_seconds = 30
max_msg_rate = 10
http_port = 3001

[[workspace]]
workspace_id = "W1"
channel_id = "C1"

[[workspace]]
workspace_id = "W2"
channel_id = "C2"
`

	_, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error for a shipped-shape config: %v", err)
	}

	want := []string{
		"acp.http_port",
		"acp.max_msg_rate",
		"acp.max_sessions",
		"acp.startup_timeout_seconds",
		"host_cli",
		"host_cli_args",
		"ipc_name",
		"slack.channel_id",
		"slack_detail_level",
		"workspace.channel_id",
		"workspace.channel_id",
	}
	if !slices.Equal(report.UnknownKeys, want) {
		t.Errorf("UnknownKeys = %v, want %v", report.UnknownKeys, want)
	}
}

func TestDecodeMalformedTOMLYieldsKindConfig(t *testing.T) {
	_, _, err := Decode("this is not = = = valid toml [[[")
	if err == nil {
		t.Fatal("Decode returned nil error for malformed TOML")
	}

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Decode error is not *apperr.Error: %v", err)
	}
	if appErr.Kind() != apperr.KindConfig {
		t.Errorf("Decode error kind = %v, want KindConfig", appErr.Kind())
	}
}

func TestDecodeOperatorDetailLevelThroughUnmarshalText(t *testing.T) {
	cfg, _, err := Decode(`default_workspace_root = "."
operator_detail_level = "verbose"
`)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.OperatorDetailLevel != DetailVerbose {
		t.Errorf("OperatorDetailLevel = %v, want DetailVerbose", cfg.OperatorDetailLevel)
	}
}

func TestDecodeCapsUnknownKeysAt64PlusMarker(t *testing.T) {
	var b strings.Builder
	b.WriteString(`default_workspace_root = "."
host_cli = "claude"
`)
	for i := range 200 {
		fmt.Fprintf(&b, "key%03d = %d\n", i, i)
	}

	_, report, err := Decode(b.String())
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if len(report.UnknownKeys) != 65 {
		t.Fatalf("len(UnknownKeys) = %d, want 65 (64 entries + marker)", len(report.UnknownKeys))
	}

	marker := report.UnknownKeys[64]
	if !strings.Contains(marker, "137") || !strings.Contains(marker, "more") {
		t.Errorf("marker = %q, want to mention 137 more entries", marker)
	}
}

func TestDecodeEscapesControlCharactersInUnknownKeys(t *testing.T) {
	data := "default_workspace_root = \".\"\nhost_cli = \"claude\"\n\"has\\nnewline\" = 1\n"
	_, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if len(report.UnknownKeys) != 2 {
		t.Fatalf("len(UnknownKeys) = %d, want 2", len(report.UnknownKeys))
	}

	var escaped string
	for _, key := range report.UnknownKeys {
		if strings.Contains(key, `\n`) {
			escaped = key
			break
		}
	}
	if escaped == "" {
		t.Fatalf("UnknownKeys = %v, want one entry containing an escaped \\n marker", report.UnknownKeys)
	}
	if strings.ContainsRune(escaped, '\n') {
		t.Errorf("escaped unknown key %q contains a raw newline; must be escaped", escaped)
	}
}

func TestSanitizeKeyPathTruncatesOnRuneBoundary(t *testing.T) {
	prefix := strings.Repeat("a", maxUnknownKeyBytes-1)
	path := prefix + "é" + "trailing"

	got := sanitizeKeyPath(path)

	if !utf8.ValidString(got) {
		t.Fatalf("sanitizeKeyPath(%d-byte input) = %q, not valid UTF-8", len(path), got)
	}
	if len(got) > maxUnknownKeyBytes {
		t.Errorf("len(sanitizeKeyPath(...)) = %d, want <= %d", len(got), maxUnknownKeyBytes)
	}
}
