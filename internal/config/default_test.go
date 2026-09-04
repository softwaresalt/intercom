package config

import "testing"

func TestDefaultNonZeroValuesMatchContract(t *testing.T) {
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
		{"HTTPPort", d.HTTPPort, uint16(3000)},
		{"Database.Path", d.Database.Path, "data/agent-rc.db"},
		{"OperatorDetailLevel", d.OperatorDetailLevel, DetailStandard},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("Default().%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestDefaultOperatorDetailLevelIsNotIotaZero(t *testing.T) {
	if Default().OperatorDetailLevel != DetailStandard {
		t.Errorf("Default().OperatorDetailLevel = %v, want DetailStandard", Default().OperatorDetailLevel)
	}
	if DetailStandard == OperatorDetailLevel(0) {
		t.Error("DetailStandard must not equal the iota zero value (DetailMinimal); a struct-literal Config would silently default to minimal")
	}
}

func TestDefaultTimeoutsWaitSecondsIsZero(t *testing.T) {
	if got := Default().Timeouts.WaitSeconds; got != 0 {
		t.Errorf("Default().Timeouts.WaitSeconds = %v, want 0", got)
	}
}
