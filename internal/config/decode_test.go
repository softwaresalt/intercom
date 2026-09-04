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

// TestDecodeReportsUnknownKeysTolerantly covers unit B1's acceptance
// criterion (i): a config with wobble = 1 and [nonsense]\nx = 2 loads
// without a decode error and reports both dotted paths sorted.
func TestDecodeReportsUnknownKeysTolerantly(t *testing.T) {
	data := "default_workspace_root = \".\"\nhost_cli = \"claude\"\nwobble = 1\n\n[nonsense]\nx = 2\n"
	cfg, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Decode returned nil *Config")
	}

	want := []string{"nonsense.x", "wobble"}
	if !slices.Equal(report.UnknownKeys, want) {
		t.Errorf("UnknownKeys = %v, want %v (sorted)", report.UnknownKeys, want)
	}
}

// TestDecodeMalformedTOMLYieldsKindConfig covers unit B1's acceptance
// criterion (ii): malformed TOML yields apperr.KindConfig.
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

// TestDecodeSlackDetailLevelThroughUnmarshalText covers unit B1's
// acceptance criterion (iii): slack_detail_level = "verbose" decoded
// through toml.Decode yields DetailVerbose, proving the decoder honours
// UnmarshalText on a value field.
func TestDecodeSlackDetailLevelThroughUnmarshalText(t *testing.T) {
	cfg, _, err := Decode("default_workspace_root = \".\"\nhost_cli = \"claude\"\nslack_detail_level = \"verbose\"\n")
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.SlackDetailLevel != DetailVerbose {
		t.Errorf("SlackDetailLevel = %v, want DetailVerbose", cfg.SlackDetailLevel)
	}
}

// TestDecodeCapsUnknownKeysAt64PlusMarker covers unit B1's acceptance
// criterion (iv), the count-bound half: a config with 200 unknown keys
// reports 64 entries plus a "… N more" marker.
func TestDecodeCapsUnknownKeysAt64PlusMarker(t *testing.T) {
	var b strings.Builder
	b.WriteString("default_workspace_root = \".\"\nhost_cli = \"claude\"\n")
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
	if !strings.Contains(marker, "136") || !strings.Contains(marker, "more") {
		t.Errorf("marker = %q, want to mention 136 more entries", marker)
	}
}

// TestDecodeEscapesControlCharactersInUnknownKeys covers unit B1's
// acceptance criterion (iv), the escaping half: a key containing \n is
// escaped rather than passed through raw. toml.Key.String() itself
// renders any key requiring quoting (including one containing a raw
// newline) using Go-syntax-style quoting, so the raw newline never reaches
// the caller; this test locks that observable behavior for
// Report.UnknownKeys.
func TestDecodeEscapesControlCharactersInUnknownKeys(t *testing.T) {
	data := "default_workspace_root = \".\"\nhost_cli = \"claude\"\n\"has\\nnewline\" = 1\n"
	_, report, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if len(report.UnknownKeys) != 1 {
		t.Fatalf("len(UnknownKeys) = %d, want 1", len(report.UnknownKeys))
	}
	if strings.ContainsRune(report.UnknownKeys[0], '\n') {
		t.Errorf("UnknownKeys[0] = %q contains a raw newline; must be escaped", report.UnknownKeys[0])
	}
	if !strings.Contains(report.UnknownKeys[0], `\n`) {
		t.Errorf("UnknownKeys[0] = %q, want an escaped \\n marker", report.UnknownKeys[0])
	}
}

// TestSanitizeKeyPathTruncatesOnRuneBoundary is a regression for a review
// finding: truncating a sanitized key path at a raw byte index can split
// a multi-byte UTF-8 rune straddling the boundary, producing an invalid
// UTF-8 string. The truncated result must always be valid UTF-8, no
// matter where the multi-byte rune falls relative to maxUnknownKeyBytes.
func TestSanitizeKeyPathTruncatesOnRuneBoundary(t *testing.T) {
	// "é" is 2 bytes (0xC3 0xA9) in UTF-8. Build a path whose multi-byte
	// rune straddles the maxUnknownKeyBytes boundary.
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
