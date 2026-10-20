package hub

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// permState is the lifecycle of one registry entry.
type permState uint8

const (
	// permPending: unresolved; the first matching action wins it.
	permPending permState = iota + 1
	// permDecided: a winner was chosen; the entry stays until the handler
	// acknowledges (or the hub aborts it), so a losing responder gets an
	// explicit already_resolved.
	permDecided
)

// pendingPerm is one registry entry. Its kind is recorded from the adapter's
// classification when the request is registered and is never taken from an
// inbound message (ADR MVP-3).
type pendingPerm struct {
	id         string
	ask        agent.PermissionAsk
	kind       agent.PermissionKind
	toolCallID string
	summary    string
	deadline   time.Time
	state      permState
	winner     string
	decision   agent.PermissionDecision
	reason     string
}

// casOutcome reports how a compare-and-swap attempt ended.
type casOutcome uint8

const (
	casWon casOutcome = iota + 1
	casLost
	casUnknown
)

// permRegistry is the hub-owned registry of permission requests (design §5.4).
// It is touched only by the hub goroutine, so it needs no lock.
type permRegistry struct {
	max   int
	next  uint64
	byID  map[string]*pendingPerm
	order []*pendingPerm
}

func newPermRegistry(max int) *permRegistry {
	return &permRegistry{max: max, byID: make(map[string]*pendingPerm)}
}

func (r *permRegistry) len() int { return len(r.order) }

// register adds a pending entry for ask, or reports false when the registry is
// full. The entry's kind and summary are recorded here, from the adapter's
// classification, and are the only source policy ever reads (ADR MVP-3).
func (r *permRegistry) register(ask agent.PermissionAsk, deadline time.Time) (*pendingPerm, bool) {
	if len(r.order) >= r.max {
		return nil, false
	}
	r.next++
	e := &pendingPerm{
		id:         "perm-" + strconv.FormatUint(r.next, 10),
		ask:        ask,
		kind:       ask.Request.Kind,
		toolCallID: ask.Request.ToolCallID,
		summary:    ask.Request.Summary,
		deadline:   deadline,
		state:      permPending,
	}
	r.byID[e.id] = e
	r.order = append(r.order, e)
	return e, true
}

func (r *permRegistry) lookup(id string) *pendingPerm { return r.byID[id] }

// cas is the first-responder compare-and-swap (design §5.4 item 3): it moves a
// pending entry to decided exactly once. The hub goroutine is the only caller,
// so the compare and the swap are indivisible. A later attempt reports
// casLost with the entry (so the loser can be told the winner), and an id that
// is not registered reports casUnknown.
func (r *permRegistry) cas(id, actor string, d agent.PermissionDecision, reason string) (casOutcome, *pendingPerm) {
	e := r.byID[id]
	switch {
	case e == nil:
		return casUnknown, nil
	case e.state != permPending:
		return casLost, e
	}
	e.state, e.winner, e.decision, e.reason = permDecided, actor, d, reason
	return casWon, e
}

func (r *permRegistry) remove(id string) {
	if _, ok := r.byID[id]; !ok {
		return
	}
	delete(r.byID, id)
	for i, e := range r.order {
		if e.id == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
}

// unresolved returns the pending (not decided) entries in registration order.
func (r *permRegistry) unresolved() []*pendingPerm {
	var out []*pendingPerm
	for _, e := range r.order {
		if e.state == permPending {
			out = append(out, e)
		}
	}
	return out
}

// permKindName is the wire name of an adapter permission classification.
func permKindName(k agent.PermissionKind) string {
	switch k {
	case agent.PermissionKindShell:
		return "shell"
	case agent.PermissionKindWrite:
		return "write"
	case agent.PermissionKindRead:
		return "read"
	case agent.PermissionKindURL:
		return "url"
	case agent.PermissionKindOther:
		return "other"
	default:
		return "unknown"
	}
}

// realAfter is the production permission timer.
func realAfter(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTimer(d)
	return t.C, func() { t.Stop() }
}

// Reasons the hub itself resolves a request (carried in
// permission.resolved) or abandons it (carried in permission.aborted).
const (
	permReasonNoSurface = "no_surface"
	permReasonTimeout   = "timeout"

	permAbortHandlerGone      = "handler_gone"
	permAbortReplyUnavailable = "reply_unavailable"
	permAbortAckTimeout       = "ack_timeout"
)

// decisionName is the wire name of a decision.
func decisionName(k agent.DecisionKind) string {
	switch k {
	case agent.DecisionAllowOnce:
		return "allow_once"
	case agent.DecisionDeny:
		return "deny"
	case agent.DecisionUnavailable:
		return "unavailable"
	default:
		return "invalid"
	}
}

// decide resolves e on the hub's own authority (timeout, shutdown, empty
// broadcast set): the same first-responder CAS an operator action goes
// through, then the hand-off to the handler.
func (h *Hub) decide(e *pendingPerm, actor string, d agent.PermissionDecision, reason string) {
	if out, _ := h.perm.cas(e.id, actor, d, reason); out != casWon {
		return
	}
	h.handoff(e)
}

// handoff delivers the decided entry's decision to the handler and starts the
// acknowledgement wait. Finality is the hand-off, not the CAS (design §5.4
// item 7): permission.resolved is published only once the handler has
// acknowledged (onHandlerReturned); if the handler is already gone, or the
// capacity-1 reply send takes the default branch, the request is abandoned
// with permission.aborted instead, so no surface renders "allowed" for a
// decision the SDK never received. The hub never blocks here.
func (h *Hub) handoff(e *pendingPerm) {
	e.deadline = h.cfg.Now().Add(h.cfg.PermissionAckTimeout) // the ack wait is hub-timed too
	select {
	case <-e.ask.Resolved:
		h.abortPerm(e, permAbortHandlerGone)
		return
	default:
	}
	if !trySend(e.ask.Reply, e.decision) {
		h.abortPerm(e, permAbortReplyUnavailable)
	}
}

// watchAsk reports the handler's return (Resolved closed) to the hub. It
// exits when the hub is done, so it cannot outlive the hub; the handler always
// returns (it selects on hub-done and its own stop), so it cannot outlive the
// request either.
func (h *Hub) watchAsk(id string, resolved <-chan struct{}) {
	select {
	case <-resolved:
	case <-h.done:
		return
	}
	select {
	case h.acks <- id:
	case <-h.done:
	}
}

// onHandlerReturned is the acknowledgement step. A decided entry whose reply
// has been consumed (the buffer is empty) was received: publish resolved. If
// the decision is still sitting in the reply buffer the handler returned by
// another path (stop, hub-done): abort. An entry that was never decided means
// the handler is gone without an answer (SDK Abort/Stop, CLI-side timeout):
// abort. An id that is no longer registered is a late duplicate: ignore.
func (h *Hub) onHandlerReturned(id string) {
	e := h.perm.lookup(id)
	if e == nil {
		return
	}
	if e.state == permDecided && len(e.ask.Reply) == 0 {
		h.finishResolved(e)
		return
	}
	h.abortPerm(e, permAbortHandlerGone)
}

// finishResolved publishes permission.resolved and retires the entry.
func (h *Hub) finishResolved(e *pendingPerm) {
	h.perm.remove(e.id)
	h.publish(KindPermissionResolved, PermissionResolvedPayload{
		RequestID: e.id,
		Decision:  decisionName(e.decision.Kind),
		Actor:     e.winner,
		Reason:    e.reason,
	})
}

// abortPerm publishes permission.aborted and retires the entry.
func (h *Hub) abortPerm(e *pendingPerm, reason string) {
	h.perm.remove(e.id)
	h.publish(KindPermissionAborted, PermissionAbortedPayload{RequestID: e.id, Reason: reason})
}

// expirePermissions is the hub-owned clock firing. A pending entry past its
// deadline resolves to UserNotAvailable through the same path an operator
// answer takes; a decided entry past its ack deadline is aborted. The hub is
// the only clock: the handler runs none (design §5.4 item 4).
func (h *Hub) expirePermissions() {
	now := h.cfg.Now()
	for _, e := range slices.Clone(h.perm.order) {
		if e.deadline.After(now) {
			continue
		}
		switch e.state {
		case permPending:
			h.decide(e, ActorHub, agent.PermissionDecision{Kind: agent.DecisionUnavailable}, permReasonTimeout)
		case permDecided:
			h.abortPerm(e, permAbortAckTimeout)
		}
	}
}

// armPermissionTimer keeps one wake-up armed for the earliest registry
// deadline, re-arming only when that deadline changes.
func (h *Hub) armPermissionTimer() {
	var earliest time.Time
	for _, e := range h.perm.order {
		if earliest.IsZero() || e.deadline.Before(earliest) {
			earliest = e.deadline
		}
	}
	if earliest.IsZero() {
		h.disarmPermissionTimer()
		return
	}
	if h.permC != nil && h.permArmed.Equal(earliest) {
		return
	}
	h.disarmPermissionTimer()
	h.permC, h.permStop = h.cfg.permAfter(max(earliest.Sub(h.cfg.Now()), 0))
	h.permArmed = earliest
}

func (h *Hub) disarmPermissionTimer() {
	if h.permStop != nil {
		h.permStop()
	}
	h.permC, h.permStop, h.permArmed = nil, nil, time.Time{}
}

// reapEmptyBroadcastSet resolves every unresolved request to UserNotAvailable
// the moment the broadcast set is empty by the permanent-TUI rule (design
// §5.4): the last remote surface is gone AND the TUI is not attached. An
// answer nobody can give is definitionally unobtainable, so waiting out the
// deadline would only keep the agent parked. The TUI counts even when it is
// stale (overflowed): it is never dropped for backpressure, so presence is
// registration, not liveness of its buffer. It runs at the end of every hub
// iteration and is free while nothing is pending.
func (h *Hub) reapEmptyBroadcastSet() {
	pending := h.perm.unresolved()
	if len(pending) == 0 {
		return
	}
	h.sweep() // a client that closed itself is not a member
	if remote, tui := h.membership(); remote != 0 || tui {
		return
	}
	for _, e := range pending {
		h.decide(e, ActorHub, agent.PermissionDecision{Kind: agent.DecisionUnavailable}, permReasonNoSurface)
	}
}

// Asks is the channel the agent adapter sends permission asks on. The hub
// answers each by sending exactly one decision on the ask's Reply (capacity 1,
// select/default) and never blocks on it.
func (h *Hub) Asks() chan<- agent.PermissionAsk { return h.asks }

// handleAsk registers one permission request and broadcasts it, both in this
// single hub iteration and in that order (design §5.4 item 1): were the
// broadcast first, a fast operator answer could reach the hub before the entry
// existed and be rejected, denying the agent a permission the operator
// granted.
func (h *Hub) handleAsk(ask agent.PermissionAsk) {
	if ask.Reply == nil || ask.Resolved == nil {
		// Cannot be answered or abandoned: fail closed rather than register an
		// entry nothing can ever finish.
		h.cfg.Logger.Error("hub: malformed permission ask discarded")
		if ask.Reply != nil {
			trySend(ask.Reply, agent.PermissionDecision{Kind: agent.DecisionUnavailable})
		}
		return
	}
	select {
	case <-ask.Resolved:
		h.cfg.Logger.Debug("hub: permission ask already abandoned; discarded")
		return
	default:
	}
	if len(ask.Request.Summary) > MaxPermissionSummaryBytes {
		// An operator cannot approve what cannot be shown in full, and the
		// hydrate section must stay bounded: deny without registering.
		h.cfg.Logger.Warn("hub: permission summary over the bound; request denied", "bytes", len(ask.Request.Summary))
		trySend(ask.Reply, agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "request is too large to review; denied"})
		return
	}
	e, ok := h.perm.register(ask, h.cfg.Now().Add(h.cfg.PermissionTimeout))
	if !ok {
		h.cfg.Logger.Warn("hub: permission registry full; request denied as unavailable", "max", h.cfg.MaxPendingPermissions)
		trySend(ask.Reply, agent.PermissionDecision{Kind: agent.DecisionUnavailable})
		return
	}
	h.publish(KindPermissionRequested, PermissionRequestedPayload{
		RequestID:      e.id,
		ToolCallID:     e.toolCallID,
		PermissionKind: permKindName(e.kind),
		Summary:        e.summary,
		RemainingMS:    h.remainingMS(e),
	})
	go h.watchAsk(e.id, ask.Resolved)
}

// remainingMS is the time left on e's deadline in whole milliseconds, never
// below 1 for a live entry so a client always sees a positive budget.
func (h *Hub) remainingMS(e *pendingPerm) int64 {
	return max(e.deadline.Sub(h.cfg.Now()).Milliseconds(), 1)
}

// handleAction processes one operator action inside a hub step. Authorization
// (067.004) runs before the CAS; until then this path is reachable only from
// in-package callers, never from a transport.
func (h *Hub) handleAction(r *actionReq) {
	trySend(r.reply, h.processAction(r))
}

func (h *Hub) processAction(r *actionReq) ActionResult {
	switch r.act.Type {
	case OpResolvePermission:
		return h.resolvePermission(r)
	default:
		return ActionResult{Reason: ReasonActionUnknown}
	}
}

// resolvePermission is the first-responder arbitration step.
func (h *Hub) resolvePermission(r *actionReq) ActionResult {
	if k := r.act.Decision.Kind; k != agent.DecisionAllowOnce && k != agent.DecisionDeny {
		return ActionResult{Reason: ReasonActionUnknown} // an operator answers allow or deny only
	}
	switch out, e := h.perm.cas(r.act.RequestID, r.from.label(), r.act.Decision, ""); out {
	case casWon:
		h.handoff(e)
		return ActionResult{Accepted: true}
	case casLost:
		return ActionResult{Reason: ReasonAlreadyResolved, Winner: e.winner}
	default:
		return ActionResult{Reason: ReasonRequestUnknown}
	}
}

// submit delivers an action to the hub and waits for its verdict. It is the
// only way an action reaches the hub step; it blocks on the hub only until ctx
// ends or the hub stops, and the hub's reply goes through a capacity-1 channel.
func (h *Hub) submit(ctx context.Context, from principal, act OperatorAction) (ActionResult, error) {
	req := &actionReq{from: from, act: act, reply: make(chan ActionResult, 1)}
	select {
	case h.actions <- req:
	case <-ctx.Done():
		return ActionResult{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		return ActionResult{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-h.stopCh:
		return ActionResult{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	}
	select {
	case res := <-req.reply:
		return res, nil
	case <-ctx.Done():
		return ActionResult{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		select { // the hub may have answered just before finishing
		case res := <-req.reply:
			return res, nil
		default:
			return ActionResult{}, apperr.Wrap(apperr.KindIO, ErrStopped)
		}
	}
}
