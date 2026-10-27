package app

import (
	"context"
	"sync/atomic"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// lateAgent breaks the construction cycle between the hub and the agent: the
// hub needs an AgentControl at New, but the adapter needs the hub's ingress,
// so the hub is given this forwarder and the real agent is bound afterwards.
// Until bound (or after a failed build) operator actions are refused with
// agent.ErrNotRunning, which the hub reports as agent_unavailable.
type lateAgent struct{ p atomic.Pointer[agent.Agent] }

func (l *lateAgent) bind(a agent.Agent) { l.p.Store(&a) }

// get returns the bound agent, or nil until one is bound.
func (l *lateAgent) get() agent.Agent {
	a := l.p.Load()
	if a == nil {
		return nil
	}
	return *a
}

func (l *lateAgent) Send(ctx context.Context, prompt string) error {
	a := l.get()
	if a == nil {
		return agent.ErrNotRunning
	}
	return a.Send(ctx, prompt)
}

func (l *lateAgent) Abort(ctx context.Context) error {
	a := l.get()
	if a == nil {
		return agent.ErrNotRunning
	}
	return a.Abort(ctx)
}
