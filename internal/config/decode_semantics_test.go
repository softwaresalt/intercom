package config

import (
	"errors"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// TestDecodeMinimalConfigKeepsAllDefaultsIntact covers unit B3's
// acceptance criterion (i): a minimal valid config yields all 17
// defaults intact.
func TestDecodeMinimalConfigKeepsAllDefaultsIntact(t *testing.T) {
	cfg, _, err := Decode("default_workspace_root = \".\"\n")
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}

	want := Default()
	if cfg.Timeouts != want.Timeouts {
		t.Errorf("Timeouts = %+v, want %+v", cfg.Timeouts, want.Timeouts)
	}
	if cfg.Stall != want.Stall {
		t.Errorf("Stall = %+v, want %+v", cfg.Stall, want.Stall)
	}
	if cfg.RetentionDays != want.RetentionDays {
		t.Errorf("RetentionDays = %v, want %v", cfg.RetentionDays, want.RetentionDays)
	}
	if cfg.MaxConcurrentSessions != want.MaxConcurrentSessions {
		t.Errorf("MaxConcurrentSessions = %v, want %v", cfg.MaxConcurrentSessions, want.MaxConcurrentSessions)
	}
	if cfg.ACP != want.ACP {
		t.Errorf("ACP = %+v, want %+v", cfg.ACP, want.ACP)
	}
	if cfg.HTTPPort != want.HTTPPort {
		t.Errorf("HTTPPort = %v, want %v", cfg.HTTPPort, want.HTTPPort)
	}
	if cfg.IPCName != want.IPCName {
		t.Errorf("IPCName = %v, want %v", cfg.IPCName, want.IPCName)
	}
	if cfg.Database != want.Database {
		t.Errorf("Database = %+v, want %+v", cfg.Database, want.Database)
	}
	if cfg.SlackDetailLevel != want.SlackDetailLevel {
		t.Errorf("SlackDetailLevel = %v, want %v", cfg.SlackDetailLevel, want.SlackDetailLevel)
	}
}

// TestDecodeExplicitFalseOverridesTrueDefault covers unit B3's acceptance
// criterion (ii): [stall]\nenabled = false yields Stall.Enabled == false,
// proving an explicitly-written false beats the non-zero true default
// (requirement R5).
func TestDecodeExplicitFalseOverridesTrueDefault(t *testing.T) {
	data := "default_workspace_root = \".\"\n[stall]\nenabled = false\n"
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.Stall.Enabled != false {
		t.Errorf("Stall.Enabled = %v, want false", cfg.Stall.Enabled)
	}
	// The rest of the [stall] section must still default.
	if cfg.Stall.InactivityThresholdSeconds != 300 {
		t.Errorf("Stall.InactivityThresholdSeconds = %v, want 300 (default preserved)", cfg.Stall.InactivityThresholdSeconds)
	}
}

// TestDecodeExplicitZeroOverridesNonZeroDefault covers unit B3's
// acceptance criterion (iii): [timeouts]\napproval_seconds = 0 yields 0,
// proving an explicitly-written zero beats the non-zero default.
func TestDecodeExplicitZeroOverridesNonZeroDefault(t *testing.T) {
	data := "default_workspace_root = \".\"\n[timeouts]\napproval_seconds = 0\n"
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.Timeouts.ApprovalSeconds != 0 {
		t.Errorf("Timeouts.ApprovalSeconds = %v, want 0", cfg.Timeouts.ApprovalSeconds)
	}
	// The rest of the [timeouts] section must still default.
	if cfg.Timeouts.PromptSeconds != 1800 {
		t.Errorf("Timeouts.PromptSeconds = %v, want 1800 (default preserved)", cfg.Timeouts.PromptSeconds)
	}
}

// TestDecodeLoneNonCanonicalSpellingDecodes covers unit B3's acceptance
// criterion (iv): a lone Host_CLI key (no collision) decodes into
// HostCLI, documenting divergence V2b — the case-fold collision rejection
// added in B2 must not over-reject a harmless non-canonical spelling with
// no colliding variant.
func TestDecodeLoneNonCanonicalSpellingDecodes(t *testing.T) {
	data := "default_workspace_root = \".\"\nHost_CLI = \"claude\"\n"
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error for a lone non-canonical spelling: %v", err)
	}
	if cfg.HostCLI != "claude" {
		t.Errorf("HostCLI = %q, want %q", cfg.HostCLI, "claude")
	}
}

// TestDecodeRejectsOutOfRangeIntegers covers unit B3's acceptance
// criterion (v): http_port = -1 and http_port = 70000 each yield a decode
// error (unsigned types make an out-of-range TOML integer a decode error,
// not a validation concern).
func TestDecodeRejectsOutOfRangeIntegers(t *testing.T) {
	tests := []string{
		"default_workspace_root = \".\"\nhttp_port = -1\n",
		"default_workspace_root = \".\"\nhttp_port = 70000\n",
	}

	for _, data := range tests {
		_, _, err := Decode(data)
		if err == nil {
			t.Errorf("Decode(%q) returned nil error, want a decode error for an out-of-range http_port", data)
			continue
		}
		var appErr *apperr.Error
		if !errors.As(err, &appErr) {
			t.Errorf("Decode(%q) error is not *apperr.Error: %v", data, err)
			continue
		}
		if appErr.Kind() != apperr.KindConfig {
			t.Errorf("Decode(%q) error kind = %v, want KindConfig", data, appErr.Kind())
		}
	}
}
