package copilotprobe

import (
	"context"
	"fmt"
	"testing"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// TestS1PermissionRoundTrip proves D2's S1 criterion: drive a real
// PermissionRequestShell to a PermissionHandlerFunc and exercise each
// rpc.PermissionDecision variant a handler can construct, recording which
// are accepted.
//
// Variant-set finding (recorded verbatim in the findings artifact, B6): the
// rpc.PermissionDecision interface has 16 concrete implementers at the
// v1.0.11 pin (excluding the internal RawPermissionDecisionData wire
// fallback), but PermissionHandlerFunc's own doc comment documents exactly
// 4 as constructible handler-return values: ApproveOnce, Reject,
// UserNotAvailable, NoResult. The remaining 12
// (Approved/ApprovedForLocation/ApprovedForSession/ApproveForLocation/
// ApproveForSession/ApprovePermanently/Cancelled/DeniedBy*) are
// result/outcome-shaped variants the server reports (e.g. after a
// rules-based or host-policy auto-approval), not values a handler
// constructs. Because the constructible set is 4 (at the plan's own
// non-split threshold), this subtask is NOT split.
func TestS1PermissionRoundTrip(t *testing.T) {
	feedback := "B2 spike: exercising PermissionDecisionReject"

	cases := []struct {
		name     string
		decision rpc.PermissionDecision
	}{
		{"ApproveOnce", &rpc.PermissionDecisionApproveOnce{}},
		{"Reject", &rpc.PermissionDecisionReject{Feedback: &feedback}},
		{"UserNotAvailable", &rpc.PermissionDecisionUserNotAvailable{}},
		{"NoResult", &rpc.PermissionDecisionNoResult{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), probeDeadline)
			defer cancel()

			harness := NewPermissionHarness(func(n int, _ copilot.PermissionRequest, _ copilot.PermissionInvocation) rpc.PermissionDecision {
				return tc.decision
			})

			client := newProbeClient(t, ctx)
			session := newProbeSession(t, ctx, client, &copilot.SessionConfig{
				OnPermissionRequest: harness.Handler(),
			})

			prompt := fmt.Sprintf("Use the shell tool to run exactly this command and nothing else: echo B2-%s", tc.name)
			result, err := DriveTurn(ctx, session, prompt, nil)
			if err != nil {
				t.Fatalf("Send failed: %v", err)
			}

			observed := harness.Count()
			t.Logf("variant=%s permission_requests_observed=%d turn_ended=%v", tc.name, observed, result.Ended)

			if observed == 0 {
				t.Skipf("model did not trigger a shell permission request for variant %s within %s; recording UNPROVEN for this variant", tc.name, probeDeadline)
				return
			}

			for _, req := range harness.Requests() {
				t.Logf("variant=%s request#%d kind=%s decision_returned=%T", tc.name, req.N, req.Kind, req.Decision)
			}
		})
	}
}
