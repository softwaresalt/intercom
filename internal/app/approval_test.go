package app

import (
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// TestApprovalSecondsBoundsAPermissionRequest: [timeouts].approval_seconds is
// the hub's permission deadline. With approval_seconds = 1 an unanswered
// request resolves to UserNotAvailable within seconds; before the mapping the
// hub's own 2 minute default applied and this wait expired.
func TestApprovalSecondsBoundsAPermissionRequest(t *testing.T) {
	f := newFactory()
	o, ws := baseOptions(t, f)
	o.ConfigPath = writeConfig(t, ws, "", "[timeouts]\napproval_seconds = 1\n")
	a := startApp(t, o)

	// With no surface connected the hub answers at once; an approver must be
	// attached for the request to wait for its deadline.
	_ = dialApp(t, a, hub.RoleApprover)
	ask := newPermAsk()
	decided := make(chan agent.PermissionDecision, 1)
	go ask.serve(f.rec, decided)
	f.fake().deps.Asks <- ask.ask

	select {
	case d := <-decided:
		if d.Kind != agent.DecisionUnavailable {
			t.Fatalf("decision = %v; want UserNotAvailable on timeout", d.Kind)
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("a request unanswered past approval_seconds=1 never timed out (config says %d)", a.Config().Timeouts.ApprovalSeconds)
	}
}

func TestApprovalTimeoutClampsHugeValues(t *testing.T) {
	if got := approvalTimeout(0); got != 0 {
		t.Errorf("approvalTimeout(0) = %v; want 0 (hub default)", got)
	}
	if got := approvalTimeout(3600); got != time.Hour {
		t.Errorf("approvalTimeout(3600) = %v; want 1h", got)
	}
	if got := approvalTimeout(1 << 63); got <= 0 {
		t.Errorf("approvalTimeout(huge) = %v; want a positive clamped duration (no overflow)", got)
	}
}
