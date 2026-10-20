package hub

import (
	"context"
	"strings"
)

// Operator-initiated agent calls (067.005-T). The hub step authorizes the
// action and hands it to a worker lane over a buffered channel with
// select/default; the agent is only ever called from the lane goroutines, never
// from the hub (design §4.1: Send and Abort can block for their whole
// deadline, and a blocked hub stops draining ingress). Prompts and aborts have
// separate lanes so an abort is never queued behind a slow prompt-send.

// routeToAgent is the hub step for OpPrompt and OpAbortTurn.
func (h *Hub) routeToAgent(r *actionReq) ActionResult {
	if ok, why := authorize(authzInput{p: r.from, op: r.act.Type, remoteApproveDestructive: h.cfg.RemoteApproveDestructive}); !ok {
		return ActionResult{Reason: why}
	}
	if h.cfg.Agent == nil {
		return ActionResult{Reason: ReasonAgentUnavailable}
	}
	switch r.act.Type {
	case OpPrompt:
		if strings.TrimSpace(r.act.Prompt) == "" {
			return ActionResult{Reason: ReasonActionUnknown} // nothing to send
		}
		if !trySend(h.promptQ, r.act.Prompt) {
			return ActionResult{Reason: ReasonAgentUnavailable}
		}
		// A turn is about to start: until the hub folds its turn.start and
		// then an idle, the session is not idle, so WaitIdle must not be
		// released by the pre-prompt flag. (If the send then fails no turn
		// starts and the flag stays false; the shutdown idle wait is bounded.)
		h.agentIdle = false
	default: // OpAbortTurn
		// One queued abort already cancels the turn; a second adds nothing, so
		// a full queue still means the request is satisfied.
		trySend(h.abortQ, struct{}{})
	}
	return ActionResult{Accepted: true}
}

// runOperatorLanes starts the prompt and abort workers. They end when ctx is
// cancelled (hub stop), which also cancels any in-flight agent call, so
// neither can outlive the hub.
func (h *Hub) runOperatorLanes(ctx context.Context) {
	if h.cfg.Agent == nil {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-h.lanesStop:
				return
			case text := <-h.promptQ:
				h.callAgent(ctx, "prompt", func(c context.Context) error { return h.cfg.Agent.Send(c, text) })
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-h.lanesStop:
				return
			case <-h.abortQ:
				h.callAgent(ctx, "abort", h.cfg.Agent.Abort)
			}
		}
	}()
}

// stopLanes is shutdown step 0 for the operator lanes, on the hub goroutine:
// whatever is queued is dropped and the lanes are told to cancel their
// in-flight call and take nothing more, so no operator prompt can start a turn
// after the shutdown's own abort. It is idempotent.
func (h *Hub) stopLanes() {
	select {
	case <-h.lanesStop:
	default:
		close(h.lanesStop)
	}
	for {
		select {
		case <-h.promptQ:
		case <-h.abortQ:
		default:
			return
		}
	}
}

// callAgent runs one bounded agent call and logs a failure. The log carries
// the action name and the error, never the prompt text (it may hold secrets).
// A call picked up after shutdown step 0, or still running when it begins, is
// not made or is cancelled.
func (h *Hub) callAgent(ctx context.Context, action string, call func(context.Context) error) {
	select {
	case <-h.lanesStop:
		return
	default:
	}
	callCtx, cancel := context.WithTimeout(ctx, h.cfg.OperatorActionTimeout)
	defer cancel()
	go func() { // ends with the call: callCtx is cancelled on return
		select {
		case <-h.lanesStop:
			cancel()
		case <-callCtx.Done():
		}
	}()
	if err := call(callCtx); err != nil {
		h.cfg.Logger.Warn("hub: operator action failed", "action", action, "err", err)
	}
}
