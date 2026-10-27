package app

import (
	"context"
	"errors"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// TestLateAgentForwardsOnlyOnceBound characterizes lateAgent: until an agent
// is bound every operator action is refused with agent.ErrNotRunning; once
// bound, Send and Abort reach the bound agent through get.
func TestLateAgentForwardsOnlyOnceBound(t *testing.T) {
	var l lateAgent
	ctx := context.Background()
	if l.get() != nil {
		t.Fatal("get() returned an agent before bind")
	}
	if err := l.Send(ctx, "hi"); !errors.Is(err, agent.ErrNotRunning) {
		t.Fatalf("Send before bind = %v; want agent.ErrNotRunning", err)
	}
	if err := l.Abort(ctx); !errors.Is(err, agent.ErrNotRunning) {
		t.Fatalf("Abort before bind = %v; want agent.ErrNotRunning", err)
	}

	fa := &fakeAgent{rec: &recorder{}}
	l.bind(fa)
	if l.get() != agent.Agent(fa) {
		t.Fatal("get() did not return the bound agent")
	}
	if err := l.Send(ctx, "hi"); err != nil {
		t.Fatalf("Send after bind: %v", err)
	}
	if err := l.Abort(ctx); err != nil {
		t.Fatalf("Abort after bind: %v", err)
	}
	if len(fa.sends) != 1 || fa.sends[0] != "hi" {
		t.Fatalf("bound agent sends = %v; want [hi]", fa.sends)
	}
	if got := fa.rec.snapshot(); len(got) != 1 || got[0] != "agent.abort" {
		t.Fatalf("bound agent events = %v; want [agent.abort]", got)
	}
}
