package config

import (
	"errors"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func TestDecodeMinimalConfigKeepsAllDefaultsIntact(t *testing.T) {
	cfg, _, err := Decode(`default_workspace_root = "."
`)
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
	if cfg.HTTPPort != want.HTTPPort {
		t.Errorf("HTTPPort = %v, want %v", cfg.HTTPPort, want.HTTPPort)
	}
	if cfg.Database != want.Database {
		t.Errorf("Database = %+v, want %+v", cfg.Database, want.Database)
	}
	if cfg.OperatorDetailLevel != want.OperatorDetailLevel {
		t.Errorf("OperatorDetailLevel = %v, want %v", cfg.OperatorDetailLevel, want.OperatorDetailLevel)
	}
}

func TestDecodeExplicitFalseOverridesTrueDefault(t *testing.T) {
	data := `default_workspace_root = "."
[stall]
enabled = false
`
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.Stall.Enabled != false {
		t.Errorf("Stall.Enabled = %v, want false", cfg.Stall.Enabled)
	}
	if cfg.Stall.InactivityThresholdSeconds != 300 {
		t.Errorf("Stall.InactivityThresholdSeconds = %v, want 300 (default preserved)", cfg.Stall.InactivityThresholdSeconds)
	}
}

func TestDecodeExplicitZeroOverridesNonZeroDefault(t *testing.T) {
	data := `default_workspace_root = "."
[timeouts]
approval_seconds = 0
`
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error: %v", err)
	}
	if cfg.Timeouts.ApprovalSeconds != 0 {
		t.Errorf("Timeouts.ApprovalSeconds = %v, want 0", cfg.Timeouts.ApprovalSeconds)
	}
	if cfg.Timeouts.PromptSeconds != 1800 {
		t.Errorf("Timeouts.PromptSeconds = %v, want 1800 (default preserved)", cfg.Timeouts.PromptSeconds)
	}
}

func TestDecodeLoneNonCanonicalSpellingDecodes(t *testing.T) {
	data := `default_workspace_root = "."
Operator_Detail_Level = "verbose"
`
	cfg, _, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode returned unexpected error for a lone non-canonical spelling: %v", err)
	}
	if cfg.OperatorDetailLevel != DetailVerbose {
		t.Errorf("OperatorDetailLevel = %v, want DetailVerbose", cfg.OperatorDetailLevel)
	}
}

func TestDecodeRejectsOutOfRangeIntegers(t *testing.T) {
	tests := []string{
		`default_workspace_root = "."
http_port = -1
`,
		`default_workspace_root = "."
http_port = 70000
`,
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
