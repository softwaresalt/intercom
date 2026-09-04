// Package config implements the intercom-go config.toml schema. It declares
// the schema types, the non-zero defaults, tolerant TOML decoding, one
// unified validation pass, and workspace-root lookup.
//
// internal/config is, and remains, a secret-free file-schema package:
// credential resolution is a later, separate phase that must compose a
// distinct value with *Config rather than adding secret fields to it.
package config

import "github.com/softwaresalt/intercom-go/internal/apperr"

// DefaultConfigPath matches the oracle's --config default (main.rs:60).
const DefaultConfigPath = "config.toml"

// MaxConfigBytes bounds a config file read (security hardening, no oracle
// counterpart — divergence V9).
const MaxConfigBytes = 1 << 20 // 1 MiB

// Config is the top-level configuration schema.
type Config struct {
	DefaultWorkspaceRoot  string              `toml:"default_workspace_root"`
	MaxConcurrentSessions uint32              `toml:"max_concurrent_sessions"`
	Copilot               CopilotConfig       `toml:"copilot"`
	Commands              map[string]string   `toml:"commands"`
	HTTPPort              uint16              `toml:"http_port"`
	Timeouts              TimeoutConfig       `toml:"timeouts"`
	Stall                 StallConfig         `toml:"stall"`
	RetentionDays         uint32              `toml:"retention_days"`
	Database              DatabaseConfig      `toml:"database"`
	OperatorDetailLevel   OperatorDetailLevel `toml:"operator_detail_level"`
	Workspaces            []WorkspaceMapping  `toml:"workspace"`
}

// CopilotConfig is the [copilot] section.
type CopilotConfig struct {
	CLIPath string `toml:"cli_path"`
}

// TimeoutConfig is the [timeouts] section.
type TimeoutConfig struct {
	ApprovalSeconds uint64 `toml:"approval_seconds"`
	PromptSeconds   uint64 `toml:"prompt_seconds"`
	// WaitSeconds defaults to 0 (oracle parity: 0 means "no timeout").
	WaitSeconds uint64 `toml:"wait_seconds"`
}

// StallConfig is the [stall] section.
type StallConfig struct {
	Enabled                    bool   `toml:"enabled"`
	InactivityThresholdSeconds uint64 `toml:"inactivity_threshold_seconds"`
	EscalationThresholdSeconds uint64 `toml:"escalation_threshold_seconds"`
	MaxRetries                 uint32 `toml:"max_retries"`
	DefaultNudgeMessage        string `toml:"default_nudge_message"`
}

// DatabaseConfig is the [database] section.
type DatabaseConfig struct {
	Path string `toml:"path"`
}

// WorkspaceMapping is one [[workspace]] entry.
type WorkspaceMapping struct {
	WorkspaceID string `toml:"workspace_id"`
	Label       string `toml:"label"`
	Path        string `toml:"path"`
}

// OperatorDetailLevel controls how much detail is rendered in operator-
// facing output. DetailStandard is the default and is deliberately not the
// iota zero value (DetailMinimal is), so a Config built by struct literal
// rather than Default() cannot silently end up "minimal".
type OperatorDetailLevel uint8

const (
	// DetailMinimal renders the least detail.
	DetailMinimal OperatorDetailLevel = iota
	// DetailStandard renders the default level of detail.
	DetailStandard
	// DetailVerbose renders the most detail.
	DetailVerbose
)

// UnmarshalText implements encoding.TextUnmarshaler so the TOML decoder can
// decode a string value directly into an OperatorDetailLevel field. It
// accepts exactly "minimal", "standard", and "verbose" and rejects
// anything else with an apperr.KindConfig error.
func (d *OperatorDetailLevel) UnmarshalText(text []byte) error {
	switch string(text) {
	case "minimal":
		*d = DetailMinimal
	case "standard":
		*d = DetailStandard
	case "verbose":
		*d = DetailVerbose
	default:
		return apperr.Newf(apperr.KindConfig, "operator_detail_level: unknown value %q; must be one of minimal, standard, verbose", string(text))
	}
	return nil
}

// Report carries bounded, non-fatal load observations for the caller to
// log. Library code never logs directly (Principle V).
type Report struct {
	// UnknownKeys is a sorted, sanitized, capped list of dotted TOML paths
	// present in the document but not in the Config schema.
	UnknownKeys []string
	// Warnings holds non-fatal advisories. Never fatal.
	Warnings []string
}
