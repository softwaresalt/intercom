package config

import "testing"

// TestDefaultHasAll17NonZeroValues covers unit A2's acceptance criterion:
// a table-driven test asserts all 17 values from the Default Contract
// (plan doc, oracle config.rs:167-341).
func TestDefaultHasAll17NonZeroValues(t *testing.T) {
	d := Default()

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"Timeouts.ApprovalSeconds", d.Timeouts.ApprovalSeconds, uint64(3600)},
		{"Timeouts.PromptSeconds", d.Timeouts.PromptSeconds, uint64(1800)},
		{"Stall.Enabled", d.Stall.Enabled, true},
		{"Stall.InactivityThresholdSeconds", d.Stall.InactivityThresholdSeconds, uint64(300)},
		{"Stall.EscalationThresholdSeconds", d.Stall.EscalationThresholdSeconds, uint64(120)},
		{"Stall.MaxRetries", d.Stall.MaxRetries, uint32(3)},
		{"Stall.DefaultNudgeMessage", d.Stall.DefaultNudgeMessage, "Continue working on the current task. Pick up where you left off."},
		{"RetentionDays", d.RetentionDays, uint32(30)},
		{"MaxConcurrentSessions", d.MaxConcurrentSessions, uint32(3)},
		{"ACP.MaxSessions", d.ACP.MaxSessions, uint32(5)},
		{"ACP.StartupTimeoutSeconds", d.ACP.StartupTimeoutSeconds, uint64(30)},
		{"ACP.MaxMsgRate", d.ACP.MaxMsgRate, uint32(10)},
		{"ACP.HTTPPort", d.ACP.HTTPPort, uint16(3001)},
		{"HTTPPort", d.HTTPPort, uint16(3000)},
		{"IPCName", d.IPCName, "agent-intercom"},
		{"Database.Path", d.Database.Path, "data/agent-rc.db"},
		{"SlackDetailLevel", d.SlackDetailLevel, DetailStandard},
	}

	if len(tests) != 17 {
		t.Fatalf("test table declares %d entries, want 17 (one per Default Contract row)", len(tests))
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("Default().%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestDefaultSlackDetailLevelIsNotIotaZero covers unit A2's acceptance
// criterion: Default().SlackDetailLevel == DetailStandard *and*
// DetailStandard != SlackDetailLevel(0).
func TestDefaultSlackDetailLevelIsNotIotaZero(t *testing.T) {
	if Default().SlackDetailLevel != DetailStandard {
		t.Errorf("Default().SlackDetailLevel = %v, want DetailStandard", Default().SlackDetailLevel)
	}
	if DetailStandard == SlackDetailLevel(0) {
		t.Error("DetailStandard must not equal the iota zero value (DetailMinimal); a struct-literal Config would silently default to minimal")
	}
}

// TestDefaultTimeoutsWaitSecondsIsZero covers unit A2's acceptance
// criterion: Default().Timeouts.WaitSeconds == 0 (oracle parity: 0 means
// "no timeout", the only field with a zero default).
func TestDefaultTimeoutsWaitSecondsIsZero(t *testing.T) {
	if got := Default().Timeouts.WaitSeconds; got != 0 {
		t.Errorf("Default().Timeouts.WaitSeconds = %v, want 0", got)
	}
}
