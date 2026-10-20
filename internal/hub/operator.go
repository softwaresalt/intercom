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
		// released by the pre-prompt flag. If the send then fails, or the
		// prompt is dropped at shutdown step 0, the lane reports it and the
		// flag is restored (retractPrompts) when no turn is running.
		h.promptsPending++
		h.agentIdle = false
	default: // OpAbortTurn
		// One queued abort already cancels the turn; a second adds nothing, so
		// a full queue still means the request is satisfied.
		trySend(h.abortQ, struct{}{})
	}
	return ActionResult{Accepted: true}
}

// runOperatorLanes starts the prompt and abort workers. They end when ctx is
// cancelled (hub stop, or shutdown step 0 via lanesCancel), which also cancels
// any in-flight agent call, so neither can outlive the hub. ctx is also the
// parent of every call's context: step 0 cancels it synchronously, so no call
// can begin with a live context after step 0 has returned.
func (h *Hub) runOperatorLanes(ctx context.Context) {
	if h.cfg.Agent == nil {
		return
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case text := <-h.promptQ:
				if !h.callAgent(ctx, "prompt", func(c context.Context) error { return h.cfg.Agent.Send(c, text) }) {
					// No turn was started by this prompt: let the hub restore
					// the idle flag it cleared when it accepted it.
					select {
					case h.promptFailed <- struct{}{}:
					case <-h.done:
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-h.abortQ:
				h.callAgent(ctx, "abort", h.cfg.Agent.Abort)
			}
		}
	}()
}

// stopLanes is shutdown step 0 for the operator lanes, on the hub goroutine:
// whatever is queued is dropped and the lanes' context is cancelled, which ends
// the lanes and the in-flight call before this returns, so no operator prompt
// can start a turn after the shutdown's own abort. Dropped prompts never ran, so
// their effect on the idle flag is retracted. It is idempotent.
func (h *Hub) stopLanes() {
	if h.lanesCancel != nil {
		h.lanesCancel()
	}
	dropped := 0
	for {
		select {
		case <-h.promptQ:
			dropped++
		case <-h.abortQ:
		default:
			h.retractPrompts(dropped)
			return
		}
	}
}

// retractPrompts is the hub-side bookkeeping for n accepted prompts that will
// start no turn (their Send failed, was never made, or they were dropped at
// step 0): they no longer hold the session busy. The idle flag is restored only
// when nothing else is pending and no turn is known to be running; otherwise
// the running turn's own idle event decides.
func (h *Hub) retractPrompts(n int) {
	if n <= 0 {
		return
	}
	h.promptsPending = max(h.promptsPending-n, 0)
	if h.promptsPending == 0 && !h.turnRunning {
		h.agentIdle = true
	}
}

// callAgent runs one bounded agent call and logs a failure. The log carries
// the action name and the error, never the prompt text (it may hold secrets).
// It reports whether the call was made and succeeded. A call picked up after
// shutdown step 0 is not made; one still running when it begins is cancelled
// (its context is a child of the lanes' context).
func (h *Hub) callAgent(ctx context.Context, action string, call func(context.Context) error) bool {
	if ctx.Err() != nil {
		return false
	}
	callCtx, cancel := context.WithTimeout(ctx, h.cfg.OperatorActionTimeout)
	defer cancel()
	if err := call(callCtx); err != nil {
		h.cfg.Logger.Warn("hub: operator action failed", "action", action, "err", err)
		return false
	}
	return true
}
