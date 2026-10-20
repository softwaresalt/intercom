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
//
// # Permission arbitration (design §5.4, ADR MVP-3)
//
// The agent adapter sends each permission request on [Hub.Asks]. In one hub
// iteration the hub registers it in the pending registry and only then
// broadcasts permission.requested, so a fast operator answer can never find no
// entry. The first action that matches an unresolved entry wins by
// compare-and-swap inside the hub goroutine; a later one is refused with an
// explicit reason ([ActionResult]). Authorization (roles, the remote
// destructive-approval opt-in, unknown kinds denied) is a pure function of
// hub-owned state evaluated before the CAS: the role comes from the
// registered [Client], the kind from the registry entry, never from the
// inbound [OperatorAction]. The hub owns the approval timeout, hands the
// decision to the handler over a capacity-1 reply channel with select/default,
// and publishes permission.resolved only after the handler acknowledged
// receipt (otherwise permission.aborted). Unresolved requests are carried in
// session.hydrate with their remaining time. When the broadcast set is empty
// (no remote surface and no TUI) pending requests resolve to UserNotAvailable
// at once.
//
// # Operator actions
//
// Prompt-send and turn-abort enter through [Client.Do], pass the same role
// matrix, and are handed to worker goroutines that call the agent off the hub.
//
// # Shutdown (design §3.1)
//
// [Shutdown] runs the ordered sequence on a dedicated goroutine, never the hub
// goroutine, through a [Gate] the transport implements. The hub keeps draining
// ingress until the last step stops it.
package hub
