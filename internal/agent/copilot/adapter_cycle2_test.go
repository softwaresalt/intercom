package copilot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// A concurrent second Stop waits for the first and reports its result.
func TestConcurrentStopWaitsForFirst(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 2 * time.Second
		f.stopBlock = make(chan struct{})
	})
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	firstReturned := make(chan struct{})
	go func() { _ = h.ad.Stop(context.Background()); close(firstReturned) }()
	deadline := time.Now().Add(2 * time.Second)
	for h.client.stops.Load() == 0 { // first Stop is inside client.Stop
		if time.Now().After(deadline) {
			t.Fatal("first Stop never reached client.Stop")
		}
		time.Sleep(time.Millisecond)
	}
	second := make(chan struct{})
	go func() { _ = h.ad.Stop(context.Background()); close(second) }()
	select {
	case <-second:
		t.Fatal("second Stop returned before the first completed")
	case <-time.After(100 * time.Millisecond):
	}
	close(h.client.stopBlock)
	for _, ch := range []chan struct{}{second, firstReturned} {
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatal("Stop never returned")
		}
	}
	if h.client.stops.Load() != 1 {
		t.Errorf("client.Stop called %d times, want 1", h.client.stops.Load())
	}
}

// Events emitted while CreateSession is still running are not lost: the
// handler is registered at creation, not afterwards (design 4.2: no silent drop).
func TestEventsDuringCreateSessionAreDelivered(t *testing.T) {
	h := newHarness(t, func(_ *Config, f *fakeClient) {
		f.preEvents = []copilotsdk.SessionEvent{{Data: &rpc.AssistantTurnStartData{TurnID: "early"}}}
	})
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case ev := <-h.ingress:
		if ev.TurnID != "early" {
			t.Fatalf("ev = %+v", ev)
		}
	default:
		t.Fatal("event emitted during CreateSession was dropped")
	}
}

func TestSentinelErrors(t *testing.T) {
	h := newHarness(t, nil)
	if err := h.ad.Send(context.Background(), "x"); !errors.Is(err, agent.ErrNotRunning) {
		t.Errorf("Send before Start = %v, want ErrNotRunning", err)
	}
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := h.ad.Start(context.Background()); !errors.Is(err, agent.ErrAlreadyStarted) {
		t.Errorf("second Start = %v, want ErrAlreadyStarted", err)
	}
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Send(context.Background(), "x"); !errors.Is(err, agent.ErrNotRunning) {
		t.Errorf("Send after Stop = %v, want ErrNotRunning", err)
	}
	h2 := newHarness(t, nil)
	_ = h2.ad.Stop(context.Background())
	if _, err := h2.ad.Start(context.Background()); !errors.Is(err, agent.ErrStopped) {
		t.Errorf("Start after Stop = %v, want ErrStopped", err)
	}
}

// Stop cancels an in-flight Send rather than leaving the caller to wait out
// its deadline.
func TestStopCancelsInFlightSend(t *testing.T) {
	h := newHarness(t, nil)
	h.client.session.sendBlock = true
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	res := make(chan error, 1)
	go func() { res <- h.ad.Send(context.Background(), "hello") }()
	time.Sleep(50 * time.Millisecond)
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case err := <-res:
		if err == nil {
			t.Fatal("in-flight Send should fail when Stop cancels it")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight Send was not cancelled by Stop (would wait out its 30s default deadline)")
	}
}

// A failed Start releases anything already wired to the stop signal.
func TestFailedStartClosesStopSignal(t *testing.T) {
	h := newHarness(t, func(_ *Config, f *fakeClient) {
		f.status = &copilotsdk.GetStatusResponse{Version: "1.0.0", ProtocolVersion: 3}
	})
	if _, err := h.ad.Start(context.Background()); err == nil {
		t.Fatal("expected version failure")
	}
	select {
	case <-h.ad.stop:
	default:
		t.Fatal("failed Start must close the stop signal so parked handlers are released")
	}
}

func TestEventDiscardAfterQuiesceIsLoggedAtWarn(t *testing.T) {
	log, buf := newBufLogger()
	h := newHarness(t, func(c *Config, _ *fakeClient) { c.Logger = log })
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnEndData{TurnID: "t"}})
	if !strings.Contains(buf.String(), "level=WARN") {
		t.Errorf("post-quiesce discard must be logged at warn, got %q", buf.String())
	}
}

func TestParseVersionRejectsSignedComponents(t *testing.T) {
	for _, v := range []string{"1.0.+84", "1.0.-84", "+1.0.84", "1.0.84-+1"} {
		if _, err := parseVersion(v); err == nil {
			t.Errorf("parseVersion(%q) should fail", v)
		}
	}
}

// A failed Start closes the callback latch exactly as Stop does, so a late SDK
// callback cannot be admitted after the failure.
func TestFailedStartQuiescesCallbackLatch(t *testing.T) {
	h := newHarness(t, func(_ *Config, f *fakeClient) {
		f.status = &copilotsdk.GetStatusResponse{Version: "1.0.0", ProtocolVersion: 3}
	})
	if _, err := h.ad.Start(context.Background()); err == nil {
		t.Fatal("expected version failure")
	}
	if h.ad.latch.enter() {
		h.ad.latch.exit()
		t.Fatal("a failed Start must close the callback latch")
	}
}

func TestSessionShutdownIsInformationalAndTerminalIdleAfterQuiesceIsDeliveredWhenIngressHasRoom(t *testing.T) {
	if !isInformational("session.shutdown") {
		t.Fatal(`"session.shutdown" must be classified informational`)
	}
	log, buf := newBufLogger()
	h := newHarness(t, func(c *Config, _ *fakeClient) { c.Logger = log })
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.ad.bridge.setTurn("t1")
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	// A post-disconnect session.shutdown is dropped quietly (no Warn).
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.SessionShutdownData{}})
	if strings.Contains(buf.String(), "level=WARN") {
		t.Errorf("post-quiesce session.shutdown must not warn: %q", buf.String())
	}
	select {
	case ev := <-h.ingress:
		t.Fatalf("session.shutdown leaked to ingress: %+v", ev)
	default:
	}

	// A terminal idle that arrives after quiesce is still delivered when
	// ingress has room, and the bridge turn is closed out.
	aborted := true
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.SessionIdleData{Aborted: &aborted}})
	select {
	case ev := <-h.ingress:
		if ev.Kind != agent.KindIdle || !ev.Aborted {
			t.Fatalf("event = %+v, want aborted idle", ev)
		}
	default:
		t.Fatal("terminal idle after quiesce was dropped although ingress had room")
	}
	if _, stopCh := h.ad.bridge.currentTurn(); stopCh == nil {
		t.Fatal("bridge lost its turn channel")
	} else {
		select {
		case <-stopCh:
		default:
			t.Error("an aborted terminal idle must abort the bridge turn")
		}
	}

	// A full ingress never blocks the late callback: it falls through to Warn.
	for len(h.ingress) < cap(h.ingress) {
		h.ingress <- agent.Event{}
	}
	done := make(chan struct{})
	go func() {
		h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.SessionIdleData{}})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("late terminal idle blocked on a full ingress")
	}
}
