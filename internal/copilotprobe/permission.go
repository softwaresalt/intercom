package copilotprobe

import (
	"sync"

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
	Decide   DecideFunc
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
		h.mu.Lock()
		n := len(h.requests)
		h.mu.Unlock()

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
