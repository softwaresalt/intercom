package copilotprobe

import (
	"strings"
	"testing"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// allowlistedDecide wraps a DecideFunc with a mechanical, deny-by-default
// shell-command allowlist (011.002-T AC, resolves A0A2D049's Principle
// VII/IV blast-radius-bounding requirement).
//
// Command ENUMERATION (a fixed literal list checked ahead of time) is
// unsatisfiable in general: every test in this package prompts a LIVE
// model, so the exact command executed is model-determined, not
// statically knowable before the probe runs. The enforceable form
// instead is per-invocation: each test knows the ONE command its own
// prompt explicitly asked the model to run ("...run exactly this command
// and nothing else: <cmd>"), and this wrapper mechanically rejects any
// shell-kind permission request whose FullCommandText does not match
// that single expected command exactly. A live model that attempts to
// run ANY other command -- whether adversarial, hallucinated, or merely
// a reasonable-seeming deviation -- is denied, never approved.
//
// Only PermissionRequestKindShell requests are gated; every other
// request kind passes through to onAllowed unchanged, since arbitrary
// shell execution (not e.g. a read/write/URL/custom-tool request) is
// this control's specific blast-radius concern.
//
// Fail-closed on shape mismatch: if request.Kind() reports
// PermissionRequestKindShell but the concrete value does not actually
// assert to rpc.PermissionRequestShell (an SDK contract violation this
// probe cannot itself repair), the test fails loudly via t.Fatalf rather
// than silently approving or silently denying.
func allowlistedDecide(t *testing.T, allowedCommand string, onAllowed DecideFunc) DecideFunc {
	t.Helper()
	return func(n int, request copilot.PermissionRequest, invocation copilot.PermissionInvocation) rpc.PermissionDecision {
		if request.Kind() == rpc.PermissionRequestKindShell {
			shellReq, ok := request.(rpc.PermissionRequestShell)
			if !ok {
				t.Fatalf("allowlistedDecide: request.Kind() == PermissionRequestKindShell but the concrete value did not assert to rpc.PermissionRequestShell (request=%#v)", request)
			}
			actual := strings.TrimSpace(shellReq.FullCommandText)
			if actual != allowedCommand {
				t.Logf("allowlistedDecide: REJECTING out-of-allowlist shell command %q (this test's allowlist permits exactly %q)", actual, allowedCommand)
				feedback := "command not on this test's allowlist"
				return &rpc.PermissionDecisionReject{Feedback: &feedback}
			}
		}
		return onAllowed(n, request, invocation)
	}
}

// TestAllowlistedDecideRejectsOutOfAllowlistCommand is a direct, live-SDK-
// independent unit test of the allowlist mechanism itself (011.002-T AC):
// a shell-kind request whose FullCommandText does not match the allowlisted
// command must be REJECTED, and onAllowed (the test's own intended
// decision) must never be invoked for it.
func TestAllowlistedDecideRejectsOutOfAllowlistCommand(t *testing.T) {
	onAllowedCalled := false
	decide := allowlistedDecide(t, "echo B2-ApproveOnce", func(n int, request copilot.PermissionRequest, invocation copilot.PermissionInvocation) rpc.PermissionDecision {
		onAllowedCalled = true
		return &rpc.PermissionDecisionApproveOnce{}
	})

	adversarial := rpc.PermissionRequestShell{FullCommandText: "rm -rf /"}
	decision := decide(0, adversarial, copilot.PermissionInvocation{})

	if _, rejected := decision.(*rpc.PermissionDecisionReject); !rejected {
		t.Fatalf("allowlistedDecide(%q) for out-of-allowlist command %q = %#v, want *rpc.PermissionDecisionReject", "echo B2-ApproveOnce", adversarial.FullCommandText, decision)
	}
	if onAllowedCalled {
		t.Fatalf("onAllowed was invoked for an out-of-allowlist command; the allowlist must short-circuit before delegating")
	}
}

// TestAllowlistedDecideAllowsMatchingCommand pins the paired positive case:
// a shell-kind request whose FullCommandText matches the allowlisted
// command exactly delegates to onAllowed and its decision is returned
// unchanged.
func TestAllowlistedDecideAllowsMatchingCommand(t *testing.T) {
	onAllowedCalled := false
	want := &rpc.PermissionDecisionApproveOnce{}
	decide := allowlistedDecide(t, "echo B2-ApproveOnce", func(n int, request copilot.PermissionRequest, invocation copilot.PermissionInvocation) rpc.PermissionDecision {
		onAllowedCalled = true
		return want
	})

	matching := rpc.PermissionRequestShell{FullCommandText: "echo B2-ApproveOnce"}
	decision := decide(0, matching, copilot.PermissionInvocation{})

	if decision != rpc.PermissionDecision(want) {
		t.Fatalf("allowlistedDecide(%q) for matching command = %#v, want the onAllowed decision %#v", "echo B2-ApproveOnce", decision, want)
	}
	if !onAllowedCalled {
		t.Fatalf("onAllowed was NOT invoked for a matching, allowlisted command")
	}
}

// TestAllowlistedDecidePassesThroughNonShellRequests pins that only
// shell-kind requests are gated -- every other request kind must delegate
// to onAllowed unconditionally, since arbitrary shell execution (not e.g.
// a read/write/URL/custom-tool request) is this control's specific
// blast-radius concern.
func TestAllowlistedDecidePassesThroughNonShellRequests(t *testing.T) {
	onAllowedCalled := false
	want := &rpc.PermissionDecisionApproveOnce{}
	decide := allowlistedDecide(t, "echo B2-ApproveOnce", func(n int, request copilot.PermissionRequest, invocation copilot.PermissionInvocation) rpc.PermissionDecision {
		onAllowedCalled = true
		return want
	})

	nonShell := rpc.PermissionRequestRead{}
	decision := decide(0, nonShell, copilot.PermissionInvocation{})

	if decision != rpc.PermissionDecision(want) {
		t.Fatalf("allowlistedDecide for a non-shell request = %#v, want the onAllowed decision %#v (non-shell requests must pass through ungated)", decision, want)
	}
	if !onAllowedCalled {
		t.Fatalf("onAllowed was NOT invoked for a non-shell request")
	}
}
