// Package hub is the single-writer event hub of intercom-go (design §4, §5).
//
// The hub owns the canonical, bounded session-state projection and is the
// only goroutine that assigns envelope sequence numbers. Normalized agent
// events arrive on a buffered ingress channel; the hub folds each one into
// the projection, freezes it into an immutable pre-marshalled envelope with
// the next seq, and enqueues it to every registered client in the same step.
//
// The hub never performs a blocking send (§4.1): every send is a
// select/default on a buffered channel or a capacity-1 reply channel. A
// client whose buffer is full is dropped (remote) or forcibly re-hydrated
// (the permanent local TUI member); neither ever stalls the hub.
//
// This package imports only internal/agent and internal/apperr. It never
// imports the Copilot SDK (design §7.1) and contains no WebSocket or TUI
// code; those surfaces consume [Client] values.
package hub
