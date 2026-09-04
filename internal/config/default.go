package config

// Default returns a *Config pre-populated with the package's non-zero
// defaults. Absent TOML keys leave these values untouched; an explicitly
// written zero/false value overrides them (requirement R5, contract in
// decode.go's Decode).
func Default() *Config {
	return &Config{
		MaxConcurrentSessions: 3,
		HTTPPort:              3000,
		Timeouts: TimeoutConfig{
			ApprovalSeconds: 3600,
			PromptSeconds:   1800,
			WaitSeconds:     0,
		},
		Stall: StallConfig{
			Enabled:                    true,
			InactivityThresholdSeconds: 300,
			EscalationThresholdSeconds: 120,
			MaxRetries:                 3,
			DefaultNudgeMessage:        "Continue working on the current task. Pick up where you left off.",
		},
		RetentionDays: 30,
		Database: DatabaseConfig{
			Path: "data/agent-rc.db",
		},
		OperatorDetailLevel: DetailStandard,
	}
}
