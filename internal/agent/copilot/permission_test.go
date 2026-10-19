package copilot

import (
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

func TestDecisionRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   agent.PermissionDecision
		kind rpc.PermissionDecisionKind
	}{
		{"allow", agent.PermissionDecision{Kind: agent.DecisionAllowOnce}, rpc.PermissionDecisionKindApproveOnce},
		{"deny", agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "no"}, rpc.PermissionDecisionKindReject},
		{"unavailable", agent.PermissionDecision{Kind: agent.DecisionUnavailable}, rpc.PermissionDecisionKindUserNotAvailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sdk := toSDKDecision(tc.in)
			if sdk.Kind() != tc.kind {
				t.Fatalf("sdk kind = %q, want %q", sdk.Kind(), tc.kind)
			}
			back, ok := fromSDKDecision(sdk)
			if !ok || back != tc.in {
				t.Fatalf("round trip = %+v ok=%v, want %+v", back, ok, tc.in)
			}
		})
	}
}

func TestToSDKDecisionFailsClosed(t *testing.T) {
	// An unset/invalid decision must never become an approval.
	got := toSDKDecision(agent.PermissionDecision{})
	if got.Kind() != rpc.PermissionDecisionKindReject {
		t.Fatalf("invalid decision mapped to %q, want reject", got.Kind())
	}
}

func TestFromSDKDecisionUnsupportedVariant(t *testing.T) {
	if _, ok := fromSDKDecision(&rpc.PermissionDecisionApprovePermanently{}); ok {
		t.Fatal("unsupported SDK variant must report ok=false")
	}
}

func testRequest() copilot.PermissionRequest {
	return &rpc.PermissionRequestShell{FullCommandText: "echo hi", ToolCallID: strp("c1")}
}

func TestPermissionHandlerReturnsOnlyHubDecision(t *testing.T) {
	asks := make(chan agent.PermissionAsk, 1)
	hubDone := make(chan struct{})
	b := newPermissionBridge(asks, hubDone, nil)

	go func() {
		ask := <-asks
		if ask.Request.Kind != agent.PermissionKindShell || ask.Request.RawKind != string(rpc.PermissionRequestKindShell) || ask.Request.Summary != "echo hi" ||
			ask.Request.SessionID != "s1" || ask.Request.ToolCallID != "c1" {
			t.Errorf("unexpected ask request %+v", ask.Request)
		}
		ask.Reply <- agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "nope"}
	}()

	got, err := b.handle(testRequest(), copilot.PermissionInvocation{SessionID: "s1"})
	if err != nil {
		t.Fatalf("handle error: %v", err)
	}
	back, ok := fromSDKDecision(got)
	if !ok || back.Kind != agent.DecisionDeny || back.Feedback != "nope" {
		t.Fatalf("handler returned %+v ok=%v, want the hub's deny", back, ok)
	}
}

func TestPermissionHandlerUserNotAvailableWhenHubDone(t *testing.T) {
	asks := make(chan agent.PermissionAsk) // unbuffered, nobody receiving
	hubDone := make(chan struct{})
	b := newPermissionBridge(asks, hubDone, nil)

	res := make(chan rpc.PermissionDecision, 1)
	go func() {
		d, _ := b.handle(testRequest(), copilot.PermissionInvocation{})
		res <- d
	}()
	time.Sleep(20 * time.Millisecond)
	close(hubDone)
	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("got %q, want user-not-available", d.Kind())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after hubDone closed")
	}
}

func TestPermissionHandlerUserNotAvailableWhenHubDoneWhileAwaitingReply(t *testing.T) {
	asks := make(chan agent.PermissionAsk, 1)
	hubDone := make(chan struct{})
	b := newPermissionBridge(asks, hubDone, nil)

	res := make(chan rpc.PermissionDecision, 1)
	go func() {
		d, _ := b.handle(testRequest(), copilot.PermissionInvocation{})
		res <- d
	}()
	<-asks // ask delivered, hub never replies
	close(hubDone)
	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("got %q, want user-not-available", d.Kind())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after hubDone closed")
	}
}

func TestPermissionHandlerPrefersHubDoneWhenAlreadyClosed(t *testing.T) {
	asks := make(chan agent.PermissionAsk, 1)
	hubDone := make(chan struct{})
	close(hubDone)
	b := newPermissionBridge(asks, hubDone, nil)
	d, err := b.handle(testRequest(), copilot.PermissionInvocation{})
	if err != nil || d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
		t.Fatalf("got %v err=%v, want user-not-available", d, err)
	}
	if len(asks) != 0 {
		t.Fatal("must not enqueue an ask once the hub is gone")
	}
}

func TestPermissionHandlerRaceClean(t *testing.T) {
	asks := make(chan agent.PermissionAsk)
	hubDone := make(chan struct{})
	b := newPermissionBridge(asks, hubDone, nil)
	go func() {
		for ask := range asks {
			ask.Reply <- agent.PermissionDecision{Kind: agent.DecisionAllowOnce}
		}
	}()
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				d, _ := b.handle(testRequest(), copilot.PermissionInvocation{})
				if d.Kind() != rpc.PermissionDecisionKindApproveOnce && d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
					t.Errorf("unexpected kind %q", d.Kind())
				}
			}
			done <- struct{}{}
		}()
	}
	time.Sleep(10 * time.Millisecond)
	for i := 0; i < 8; i++ {
		if i == 4 {
			close(hubDone)
		}
		<-done
	}
}
