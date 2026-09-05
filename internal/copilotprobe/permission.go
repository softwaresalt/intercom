package copilotprobe

import (
	"sync"
	"sync/atomic"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// RecordedPermissionRequest is one observed permission request/decision pair.
type RecordedPermissionRequest struct {
	N        int
	Kind     rpc.PermissionRequestKind
	Request  copilot.PermissionRequest
	Decision rpc.PermissionDecision
}

// DecideFunc computes the decision for the n-th (0-indexed) permission
// request observed by a PermissionHarness.
type DecideFunc func(n int, request copilot.PermissionRequest, invocation copilot.PermissionInvocation) rpc.PermissionDecision

// PermissionHarness drives a caller-supplied decision function for every
// permission request delivered to a PermissionHandlerFunc, recording each
// request/decision pair. It is built by B2 (S1 proof) and reused by B4
// (blocked-handler condition for SQ-a) and B5 (SQ-c dispatch-independence
// probe), per the plan's dependency graph.
type PermissionHarness struct {
	mu       sync.Mutex
	requests []RecordedPermissionRequest
	nextN    atomic.Int64 // reserves each request's N independently of
	// slice-append timing, so N is well-defined even under concurrent
	// handler invocation (SQ-b/SQ-c genuinely probe for concurrent
	// dispatch; N must not be a length-based TOCTOU race). atomic.Int64
	// (not a raw int64) avoids the classic 32-bit alignment footgun for
	// sync/atomic operations on 64-bit values and is the idiomatic choice
	// in modern Go (>= 1.19).
	Decide DecideFunc
}

// NewPermissionHarness constructs a harness that delegates every incoming
// permission request to decide.
func NewPermissionHarness(decide DecideFunc) *PermissionHarness {
	return &PermissionHarness{Decide: decide}
}

// Handler returns a copilot.PermissionHandlerFunc bound to this harness,
// suitable for SessionConfig.OnPermissionRequest.
func (h *PermissionHarness) Handler() copilot.PermissionHandlerFunc {
	return func(request copilot.PermissionRequest, invocation copilot.PermissionInvocation) (rpc.PermissionDecision, error) {
		// Reserve this request's index atomically, independent of the
		// (unlocked, potentially long-blocking) Decide call below and the
		// later re-lock for the append -- this is the actual arrival-order
		// index, not a length snapshot that could collide under
		// concurrent invocation.
		n := int(h.nextN.Add(1) - 1)

		decision := h.Decide(n, request, invocation)

		h.mu.Lock()
		h.requests = append(h.requests, RecordedPermissionRequest{
			N:        n,
			Kind:     request.Kind(),
			Request:  request,
			Decision: decision,
		})
		h.mu.Unlock()

		return decision, nil
	}
}

// Requests returns a snapshot of every request/decision pair observed so far.
func (h *PermissionHarness) Requests() []RecordedPermissionRequest {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]RecordedPermissionRequest, len(h.requests))
	copy(out, h.requests)
	return out
}

// Count returns the number of permission requests observed so far.
func (h *PermissionHarness) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.requests)
}
