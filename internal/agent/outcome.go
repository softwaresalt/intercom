package agent

import "sync"

// PermissionOutcome is the one-shot arbitration point between the permission
// handler (the adapter) and the hub for ONE ask. It is what makes "the SDK got
// exactly this decision" a recorded fact instead of something the hub infers
// from a channel buffer:
//
//   - the handler calls [PermissionOutcome.Deliver] with the exact decision it is
//     about to return to the SDK, BEFORE it returns (and so before the ask's
//     Resolved channel closes). Deliver fails if the hub revoked the ask; the
//     handler then returns UserNotAvailable;
//   - the hub calls [PermissionOutcome.Revoke] when it will not wait for the
//     handler any longer (ack deadline, shutdown). Revoke succeeds only if the
//     handler has not delivered yet, and from then on nothing is delivered: an
//     allow can never reach the SDK after the surfaces were told it was
//     abandoned. If Revoke fails the handler already delivered, and
//     [PermissionOutcome.Delivered] says exactly what.
//
// Exactly one of Deliver and Revoke wins, so there is no window in which the
// handler has consumed a decision but the hub cannot tell. The zero value is
// ready to use. Methods are safe for concurrent use.
type PermissionOutcome struct {
	mu      sync.Mutex
	state   outcomeState
	decided PermissionDecision
}

type outcomeState uint8

const (
	outcomeOpen outcomeState = iota
	outcomeDelivered
	outcomeRevoked
)

// Deliver records d as the decision returned to the SDK. It reports false, and
// records nothing, if the outcome was revoked: the caller must then return
// UserNotAvailable. Only the first successful call records a decision; a later
// call reports false.
func (o *PermissionOutcome) Deliver(d PermissionDecision) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.state != outcomeOpen {
		return false
	}
	o.state, o.decided = outcomeDelivered, d
	return true
}

// Revoke prevents any later delivery. It reports true if the handler has not
// delivered (it is idempotent), and false if a decision was already delivered.
func (o *PermissionOutcome) Revoke() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.state == outcomeDelivered {
		return false
	}
	o.state = outcomeRevoked
	return true
}

// Delivered returns the decision the handler delivered to the SDK, if any.
func (o *PermissionOutcome) Delivered() (PermissionDecision, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.decided, o.state == outcomeDelivered
}
