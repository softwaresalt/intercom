// Package config implements the agent-intercom config.toml schema ported
// from the behavioral oracle (softwaresalt/agent-intercom @ 41df772,
// src/config.rs). It declares the schema types, the 17 non-zero defaults,
// tolerant TOML decoding, one unified validation pass, and workspace-to-
// channel routing.
//
// internal/config is, and remains, a secret-free file-schema package: the
// oracle's four secret-bearing fields (app_token, bot_token, team_id,
// authorized_user_ids) are #[serde(skip)] in the oracle and are omitted
// entirely here. Credential resolution is a later, separate phase (P3) that
// must compose a distinct value with *Config rather than adding secret
// fields to it (Decision R7).
package config

import "github.com/softwaresalt/intercom-go/internal/apperr"

// DefaultConfigPath matches the oracle's --config default (main.rs:60).
const DefaultConfigPath = "config.toml"

// MaxConfigBytes bounds a config file read (security hardening, no oracle
// counterpart — divergence V9).
const MaxConfigBytes = 1 << 20 // 1 MiB

// Config is the top-level configuration schema, matching the oracle's
// GlobalConfig TOML key names exactly (requirement R1).
type Config struct {
	DefaultWorkspaceRoot  string             `toml:"default_workspace_root"`
	Slack                 SlackConfig        `toml:"slack"`
	MaxConcurrentSessions uint32             `toml:"max_concurrent_sessions"`
	HostCLI               string             `toml:"host_cli"`
	HostCLIArgs           []string           `toml:"host_cli_args"`
	Commands              map[string]string  `toml:"commands"`
	HTTPPort              uint16             `toml:"http_port"`
	IPCName               string             `toml:"ipc_name"`
	Timeouts              TimeoutConfig      `toml:"timeouts"`
	Stall                 StallConfig        `toml:"stall"`
	RetentionDays         uint32             `toml:"retention_days"`
	Database              DatabaseConfig     `toml:"database"`
	SlackDetailLevel      SlackDetailLevel   `toml:"slack_detail_level"`
	ACP                   ACPConfig          `toml:"acp"`
	Workspaces            []WorkspaceMapping `toml:"workspace"`
}

// SlackConfig is the [slack] section. The oracle's secret-bearing
// app_token, bot_token, and team_id fields are #[serde(skip)] and are
// deliberately omitted here (Decision R6).
type SlackConfig struct {
	ChannelID                string            `toml:"channel_id"`
	MarkdownUploadExtensions map[string]string `toml:"markdown_upload_extensions"`
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

// ACPConfig is the [acp] section.
type ACPConfig struct {
	MaxSessions           uint32 `toml:"max_sessions"`
	StartupTimeoutSeconds uint64 `toml:"startup_timeout_seconds"`
	MaxMsgRate            uint32 `toml:"max_msg_rate"`
	HTTPPort              uint16 `toml:"http_port"`
}

// DatabaseConfig is the [database] section.
type DatabaseConfig struct {
	Path string `toml:"path"`
}

// WorkspaceMapping is one [[workspace]] entry.
type WorkspaceMapping struct {
	WorkspaceID string `toml:"workspace_id"`
	ChannelID   string `toml:"channel_id"`
	Label       string `toml:"label"`
	Path        string `toml:"path"`
}

// SlackDetailLevel controls how much detail is rendered in Slack messages.
// DetailStandard is the default and is deliberately not the iota zero
// value (DetailMinimal is), so a Config built by struct literal rather than
// Default() cannot silently end up "minimal".
type SlackDetailLevel uint8

const (
	// DetailMinimal renders the least detail.
	DetailMinimal SlackDetailLevel = iota
	// DetailStandard renders the oracle's default level of detail.
	DetailStandard
	// DetailVerbose renders the most detail.
	DetailVerbose
)

// UnmarshalText implements encoding.TextUnmarshaler so the TOML decoder can
// decode a string value directly into a SlackDetailLevel field. It accepts
// exactly "minimal", "standard", and "verbose" and rejects anything else
// with an apperr.KindConfig error.
func (d *SlackDetailLevel) UnmarshalText(text []byte) error {
	switch string(text) {
	case "minimal":
		*d = DetailMinimal
	case "standard":
		*d = DetailStandard
	case "verbose":
		*d = DetailVerbose
	default:
		return apperr.Newf(apperr.KindConfig, "slack_detail_level: unknown value %q; must be one of minimal, standard, verbose", string(text))
	}
	return nil
}

// Report carries bounded, non-fatal load observations for the caller to
// log. Library code never logs directly (Principle V).
type Report struct {
	// UnknownKeys is a sorted, sanitized, capped list of dotted TOML paths
	// present in the document but not in the Config schema.
	UnknownKeys []string
	// Warnings holds non-fatal advisories, such as the host_cli PATH
	// advisory. Never fatal.
	Warnings []string
}
