package hub

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.002-T: hub-owned timeout, reply hand-off, resolved vs aborted ----

// fakeClock is the hub's Now for permission tests.
type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{t: time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)} }
func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// fakeTimers stands in for the hub's permission wake-up timer. Every arming
// publishes its fire channel on armed; the test fires it by sending a value.
type fakeTimers struct {
	armed chan chan time.Time
}

func newFakeTimers() *fakeTimers { return &fakeTimers{armed: make(chan chan time.Time, 64)} }

func (f *fakeTimers) after(time.Duration) (<-chan time.Time, func()) {
	ch := make(chan time.Time, 1)
	f.armed <- ch
	return ch, func() {}
}

// fire waits for the hub to arm a timer and fires the most recent one.
func (f *fakeTimers) fire(t *testing.T) {
	t.Helper()
	var ch chan time.Time
	select {
	case ch = <-f.armed:
	case <-time.After(testWait):
		t.Fatal("hub never armed its permission timer")
	}
	for { // the latest arming wins
		select {
		case ch = <-f.armed:
		default:
			ch <- time.Time{}
			return
		}
	}
}

type permRig struct {
	h     *Hub
	clock *fakeClock
	tm    *fakeTimers
}

func newPermRig(t *testing.T, cfg Config) *permRig {
	t.Helper()
	r := &permRig{clock: newFakeClock(), tm: newFakeTimers()}
	cfg.Now = r.clock.Now
	cfg.permAfter = r.tm.after
	r.h = startHub(t, cfg)
	return r
}

// fire fires the hub's currently armed permission timer. A hub step (Stats)
// comes first so every earlier event has been processed and the timer has been
// re-armed for the current earliest deadline.
func (r *permRig) fire(t *testing.T) {
	t.Helper()
	stats(t, r.h)
	r.tm.fire(t)
}

// noFrame asserts, after a hub barrier, that nothing is queued for c.
func noFrame(t *testing.T, h *Hub, c *Client, why string) {
	t.Helper()
	stats(t, h) // a hub step after the event under test
	select {
	case f := <-c.Frames():
		t.Fatalf("%s: unexpected %s frame", why, f.Kind())
	default:
	}
}

func resolvedPayload(t *testing.T, f Frame) PermissionResolvedPayload {
	t.Helper()
	var p PermissionResolvedPayload
	if err := json.Unmarshal(decode(t, f).Payload, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func abortedPayload(t *testing.T, f Frame) PermissionAbortedPayload {
	t.Helper()
	var p PermissionAbortedPayload
	if err := json.Unmarshal(decode(t, f).Payload, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// takeReply receives the hub's decision, failing the test (not hanging) when
// none arrives.
func (a *testAsk) takeReply(t *testing.T) agent.PermissionDecision {
	t.Helper()
	select {
	case d := <-a.reply:
		return d
	case <-time.After(testWait):
		t.Fatal("no decision was delivered to the handler")
		return agent.PermissionDecision{}
	}
}

// handlerAcks plays the adapter handler: it consumes the reply and then
// returns (closing Resolved), which is the acknowledgement.
func (a *testAsk) handlerAcks(t *testing.T) agent.PermissionDecision {
	t.Helper()
	d := a.takeReply(t)
	close(a.resolved)
	return d
}

func TestHubTimeoutResolvesUnavailableThroughTheHubOnly(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))

	r.clock.Advance(time.Minute)
	r.fire(t)
	select {
	case d := <-a.reply:
		if d.Kind != agent.DecisionUnavailable {
			t.Fatalf("timeout decision = %v, want unavailable", d.Kind)
		}
	case <-time.After(testWait):
		t.Fatal("hub timeout did not answer the handler")
	}
	noFrame(t, r.h, tui, "resolved must wait for the handler ack")
	close(a.resolved)
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.RequestID != id || p.Decision != "unavailable" || p.Actor != ActorHub || p.Reason != "timeout" {
		t.Fatalf("resolved = %+v", p)
	}
	// An answer arriving after the timeout finds nothing: the hub is the sole
	// writer of the decision.
	if res := resolveAs(t, r.h, tuiP(), id, allowOnce); res.Accepted {
		t.Fatalf("a post-timeout answer was accepted: %+v", res)
	}
}

func TestResolvedIsPublishedOnlyAfterHandlerAck(t *testing.T) {
	r := newPermRig(t, Config{})
	tui := connect(t, r.h, ClientTUI, 16)
	rem := connect(t, r.h, ClientRemote, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	recvKind(t, rem, KindPermissionRequested)

	if res := resolveAs(t, r.h, approverP(), id, allowOnce); !res.Accepted {
		t.Fatalf("resolve refused: %+v", res)
	}
	select {
	case d := <-a.reply:
		if d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("handler got %v", d.Kind)
		}
	case <-time.After(testWait):
		t.Fatal("decision not delivered to the handler")
	}
	noFrame(t, r.h, tui, "the handler has not acknowledged yet")
	noFrame(t, r.h, rem, "the handler has not acknowledged yet")

	close(a.resolved) // handler returned: acknowledged
	for _, c := range []*Client{tui, rem} {
		p := resolvedPayload(t, recvKind(t, c, KindPermissionResolved))
		if p.RequestID != id || p.Decision != "allow_once" || p.Actor != ActorRemoteApprover || p.Reason != "" {
			t.Fatalf("resolved = %+v", p)
		}
	}
	if n := stats(t, r.h).PendingPermissions; n != 0 {
		t.Fatalf("pending = %d after resolution", n)
	}
	// The finished request is gone: a late responder is told it is unknown.
	if res := resolveAs(t, r.h, tuiP(), id, denyDecision); res.Reason != ReasonRequestUnknown {
		t.Fatalf("late responder = %+v, want request_unknown", res)
	}
}

func TestLoserBetweenDecisionAndAckGetsAlreadyResolvedWithWinner(t *testing.T) {
	r := newPermRig(t, Config{})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	resolveAs(t, r.h, tuiP(), id, allowOnce)
	res := resolveAs(t, r.h, approverP(), id, denyDecision)
	if res.Accepted || res.Reason != ReasonAlreadyResolved || res.Winner != ActorLocalTUI {
		t.Fatalf("loser = %+v", res)
	}
	a.handlerAcks(t)
}

func TestAbortedWhenHandlerIsGoneBeforeAnyDecision(t *testing.T) {
	r := newPermRig(t, Config{})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))

	close(a.resolved) // SDK Stop / adapter shutdown: the handler returned without an answer
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "handler_gone" {
		t.Fatalf("aborted = %+v", p)
	}
	if res := resolveAs(t, r.h, tuiP(), id, allowOnce); res.Reason != ReasonRequestUnknown {
		t.Fatalf("answer to an aborted request = %+v, want request_unknown", res)
	}
	if n := stats(t, r.h).PendingPermissions; n != 0 {
		t.Fatalf("pending = %d after abort", n)
	}
}

func TestAbortedNotResolvedWhenHandlerReturnedWithoutTakingTheDecision(t *testing.T) {
	r := newPermRig(t, Config{})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	resolveAs(t, r.h, tuiP(), id, allowOnce)
	// The decision sits unconsumed in the reply buffer and the handler is
	// gone (it chose its stop branch): both surfaces must NOT render "allowed".
	close(a.resolved)
	f := recvKind(t, tui, KindPermissionAborted)
	if p := abortedPayload(t, f); p.RequestID != id || p.Reason != "handler_gone" {
		t.Fatalf("aborted = %+v", p)
	}
	noFrame(t, r.h, tui, "no permission.resolved may follow an abort")
}

func TestReplyChannelFullEmitsAbortedAndNeverBlocksTheHub(t *testing.T) {
	r := newPermRig(t, Config{})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.reply <- agent.PermissionDecision{Kind: agent.DecisionDeny} // reply buffer already full
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))

	resolveAs(t, r.h, tuiP(), id, allowOnce)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "reply_unavailable" {
		t.Fatalf("aborted = %+v", p)
	}
	// The hub kept running: a later event is still folded and broadcast.
	send(t, r.h, msgEvent("m1", "still alive"))
	recvKind(t, tui, KindAgentMessage)
}

func TestDecidedRequestThatIsNeverAcknowledgedIsAbortedAtTheAckDeadline(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour, PermissionAckTimeout: 5 * time.Second})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	resolveAs(t, r.h, tuiP(), id, allowOnce)
	a.takeReply(t) // received, but the handler never returns

	r.clock.Advance(5 * time.Second)
	r.fire(t)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "ack_timeout" {
		t.Fatalf("aborted = %+v", p)
	}
}

func TestTimerFiringAfterOperatorWonDoesNotOverrideTheDecision(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute, PermissionAckTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	resolveAs(t, r.h, tuiP(), id, allowOnce)

	r.clock.Advance(time.Minute) // the original request deadline passes
	r.fire(t)
	stats(t, r.h)
	a.handlerAcks(t)
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.Decision != "allow_once" || p.Actor != ActorLocalTUI {
		t.Fatalf("resolved = %+v: the timeout overrode the operator", p)
	}
	select {
	case d := <-a.reply:
		t.Fatalf("a second decision %v was sent to the handler", d.Kind)
	default:
	}
}

func TestPermissionTimeoutConfigValidation(t *testing.T) {
	if _, err := New(Config{PermissionAckTimeout: -1}); err == nil {
		t.Error("negative PermissionAckTimeout must be rejected")
	}
	if _, err := New(Config{PermissionTimeout: -1}); err == nil {
		t.Error("negative PermissionTimeout must be rejected")
	}
	if _, err := New(Config{MaxPendingPermissions: -1}); err == nil {
		t.Error("negative MaxPendingPermissions must be rejected")
	}
}
