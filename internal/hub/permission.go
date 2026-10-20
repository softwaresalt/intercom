package hub

import (
	"context"
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
