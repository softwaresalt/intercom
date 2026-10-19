package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultCopilotStartupTimeout(t *testing.T) {
	if got := Default().Copilot.StartupTimeoutSeconds; got != 30 {
		t.Errorf("Default().Copilot.StartupTimeoutSeconds = %d, want 30", got)
	}
}

func TestValidateCopilotStartupTimeout(t *testing.T) {
	cases := []struct {
		name    string
		seconds uint64
		wantErr string
	}{
		{"default ok", 30, ""},
		{"minimum ok", 1, ""},
		{"maximum ok", 3600, ""},
		{"zero rejected", 0, "copilot.startup_timeout_seconds must be greater than zero"},
		{"too large rejected", 3601, "copilot.startup_timeout_seconds must not exceed 3600"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Default()
			cfg.DefaultWorkspaceRoot = t.TempDir()
			cfg.Copilot.StartupTimeoutSeconds = tc.seconds
			_, err := cfg.Validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestDecodeCopilotStartupTimeoutFromTOML(t *testing.T) {
	root := strings.ReplaceAll(t.TempDir(), `\`, `/`)
	cfg, _, err := Decode("default_workspace_root = \"" + root + "\"\n[copilot]\nstartup_timeout_seconds = 45\n")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if cfg.Copilot.StartupTimeoutSeconds != 45 {
		t.Errorf("StartupTimeoutSeconds = %d, want 45", cfg.Copilot.StartupTimeoutSeconds)
	}
}

func TestConfigReferenceDocumentsStartupTimeout(t *testing.T) {
	data, err := os.ReadFile(configReferencePath(t))
	if err != nil {
		t.Fatalf("read config reference: %v", err)
	}
	doc := string(data)
	for _, want := range []string{"copilot.startup_timeout_seconds", "Copilot.StartupTimeoutSeconds"} {
		if !strings.Contains(doc, want) {
			t.Errorf("docs/config-reference.md is missing %q", want)
		}
	}
}
