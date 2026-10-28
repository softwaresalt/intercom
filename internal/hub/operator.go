package hub

import (
	"context"
	"strings"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// Operator-initiated agent calls (067.005-T). The hub step authorizes the
// action and hands it to a worker lane over a buffered channel with
// select/default; the agent is only ever called from the lane goroutines, never
// from the hub (design §4.1: Send and Abort can block for their whole
// deadline, and a blocked hub stops draining ingress). Prompts and aborts have
// separate lanes so an abort is never queued behind a slow prompt-send.

// ErrCodeOperatorActionFailed is the error envelope code published when an
// operator-initiated agent call (prompt send, turn abort) fails.
const ErrCodeOperatorActionFailed = "operator_action_failed"

// Fixed, sanitized client-facing messages for a failed operator action: the
// underlying error (which may carry adapter or credential text) is only logged.
const (
	msgPromptSendFailed = "operator prompt send failed"
	msgTurnAbortFailed  = "operator turn abort failed"
)

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
		// released by the pre-prompt flag. The count is given back
		// (retractPrompts) only for a prompt that provably never reached
		// Agent.Send: dropped at shutdown step 0, or picked up by a lane
		// whose context was already cancelled. A Send that was invoked and
		// returned an error is delivery-uncertain (the runtime may have
		// emitted turn.start before the call timed out), so it is NOT
		// retracted: the turn.start/idle accounting stays authoritative and
		// the bounded idle wait covers a prompt that never starts a turn.
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
	h.lanesWG.Add(2)
	go func() {
		defer h.lanesWG.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case text := <-h.promptQ:
				if invoked := h.callAgent(ctx, "prompt", func(c context.Context) error { return h.cfg.Agent.Send(c, text) }); !invoked {
					// Agent.Send was never invoked (the lanes were cancelled
					// first), so this prompt provably started no turn: let the
					// hub give back the count it took when it accepted it. The
					// send never races ctx.Done(): the report must not be lost
					// to the very cancellation that caused it.
					trySend(h.promptFailed, struct{}{})
				}
			}
		}
	}()
	go func() {
		defer h.lanesWG.Done()
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

// defaultLaneJoinBudget bounds how long Run waits for the lane workers to
// observe cancellation before it gives up on them.
const defaultLaneJoinBudget = 100 * time.Millisecond

// waitLanes joins the lane workers after their context was cancelled, bounded:
// a worker stuck in an agent call that ignores its context must not keep
// Stop, Done or the shutdown executor from finishing. On a timeout the worker
// is left to finish on its own and the join reports
// agent.ErrCleanupIncomplete.
func (h *Hub) waitLanes() error {
	budget := defaultLaneJoinBudget
	if h.cfg.OperatorActionTimeout > 0 {
		budget = min(h.cfg.OperatorActionTimeout, budget)
	}
	done := make(chan struct{})
	go func() { // ends when the workers do; they always observe cancellation or finish their call
		h.lanesWG.Wait()
		close(done)
	}()
	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case <-done:
		for { // nothing publishes after Run: drop what late failures queued
			select {
			case <-h.actionFailed:
			case <-h.opErrs:
			default:
				return nil
			}
		}
	case <-timer.C:
		h.cfg.Logger.Error("hub: operator lane worker ignored cancellation", "err", agent.ErrCleanupIncomplete)
		return apperr.Wrap(apperr.KindIO, agent.ErrCleanupIncomplete) // no KindInternal exists; KindIO is the lifecycle kind
	}
}

// opErr is the cause of a failed operator call.
type opErr struct {
	action string
	err    error
}

// handleActionFailed folds and publishes a sanitized error for a failed
// operator call. It does not retract h.promptsPending: a Send that was invoked
// and failed is delivery-uncertain (see routeToAgent).
func (h *Hub) handleActionFailed(action string) {
	if h.closing {
		return
	}
	msg := msgTurnAbortFailed
	if action == "prompt" {
		msg = msgPromptSendFailed
	}
	h.fold.ApplyError(ErrCodeOperatorActionFailed, msg)
	h.publish(KindError, ErrorPayload{Code: ErrCodeOperatorActionFailed, Message: msg})
}

// stopLanes is shutdown step 0 for the operator lanes, on the hub goroutine:
// whatever is queued is dropped and the lanes' context is cancelled
// synchronously. Cancelling the context cancels the context of every agent call
// the lanes make, so no call can begin with a live context after this returns.
// It does not wait for a lane that is already inside Agent.Send or Abort: that
// call sees its context cancelled and returns on its own, and the shutdown's own
// Abort and idle wait cover the turn it may have started. Dropped prompts never
// ran, so their effect on the idle flag is retracted. It is idempotent.
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

// retractPrompts is the hub-side bookkeeping for n accepted prompts that
// provably never reached Agent.Send (dropped at step 0, or picked up after the
// lanes were cancelled): they can start no turn, so they no longer hold the
// session busy. It must not be used for a Send that was invoked and failed:
// that outcome cannot be correlated with a turn.start, and retracting could
// consume the count of a different, later prompt. The idle flag is restored only
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
// It reports whether the call was invoked at all, regardless of its outcome:
// false means it provably never ran (picked up after shutdown step 0), true
// means it ran and may have had an effect even if it returned an error. One
// still running when step 0 happens is cancelled (its context is a child of
// the lanes' context).
func (h *Hub) callAgent(ctx context.Context, action string, call func(context.Context) error) bool {
	if ctx.Err() != nil {
		return false
	}
	callCtx, cancel := context.WithTimeout(ctx, h.cfg.OperatorActionTimeout)
	defer cancel()
	if err := call(callCtx); err != nil {
		h.cfg.Logger.Warn("hub: operator action failed", "action", action, "err", err)
		if ctx.Err() == nil { // a cancelled call is shutdown, not a fault to report
			// Non-blocking: a late-returning stalled call must never block
			// on a hub whose Run has already exited.
			trySend(h.actionFailed, action)
			trySend(h.opErrs, opErr{action: action, err: err})
		}
	}
	return true
}
