package copilot

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// F1: Stop racing an in-flight Start must not orphan the runtime.
func TestStopDuringStartAbortsStartAndTearsDownClient(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StartupTimeout = 10 * time.Second
		f.startBlock = true
	})
	startErr := make(chan error, 1)
	go func() { _, err := h.ad.Start(context.Background()); startErr <- err }()
	<-h.client.startEntered // Start is parked in client.Start

	stopErr := make(chan error, 1)
	go func() { stopErr <- h.ad.Stop(context.Background()) }()

	select {
	case err := <-startErr:
		if err == nil {
			t.Fatal("Start must fail when Stop interrupts it")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start was not aborted by Stop")
	}
	select {
	case <-stopErr:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not return")
	}
	if h.client.forceStops.Load() == 0 && h.client.stops.Load() == 0 {
		t.Fatal("runtime was left running: neither Stop nor ForceStop was called")
	}
	if h.client.createCalls.Load() != 0 {
		t.Error("no session may be created after Stop")
	}
}

func TestStopDuringCreateSessionDoesNotLeakSession(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 200 * time.Millisecond
		f.createGate = make(chan struct{})
		f.createEntered = make(chan struct{})
	})
	startErr := make(chan error, 1)
	go func() { _, err := h.ad.Start(context.Background()); startErr <- err }()
	<-h.client.createEntered

	if err := h.ad.Stop(context.Background()); err == nil {
		t.Log("Stop reported no error")
	}
	close(h.client.createGate) // CreateSession now completes after Stop already ran

	select {
	case err := <-startErr:
		if err == nil {
			t.Fatal("Start must fail after Stop")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return")
	}
	if h.client.session.disconnects.Load() != 1 {
		t.Errorf("late session was not disconnected (disconnects=%d)", h.client.session.disconnects.Load())
	}
	if h.client.forceStops.Load() == 0 && h.client.stops.Load() == 0 {
		t.Error("runtime was left running")
	}
	if _, err := h.ad.activeSession(); err == nil {
		t.Error("adapter must not expose a session after Stop")
	}
}

// F2: Stop is bounded even when the SDK hangs.
func TestStopBoundedWhenDisconnectHangs(t *testing.T) {
	h := newHarness(t, func(c *Config, _ *fakeClient) { c.StopTimeout = 100 * time.Millisecond })
	h.client.session.disconnectBlock = make(chan struct{})
	defer close(h.client.session.disconnectBlock)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err := h.ad.Stop(context.Background())
	if err == nil {
		t.Error("a hung Disconnect must be reported")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("Stop took %v with a hung Disconnect", time.Since(start))
	}
	if h.client.forceStops.Load() == 0 {
		t.Error("a hung Disconnect must escalate to ForceStop")
	}
}

func TestStopBoundedWhenClientStopHangs(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 100 * time.Millisecond
		f.stopBlock = make(chan struct{})
	})
	defer close(h.client.stopBlock)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err := h.ad.Stop(context.Background())
	if err == nil {
		t.Error("a hung client Stop must be reported")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("Stop took %v with a hung client Stop", time.Since(start))
	}
	if h.client.forceStops.Load() != 1 {
		t.Errorf("ForceStop called %d times, want 1", h.client.forceStops.Load())
	}
}

// F3: a pending permission request is released when the adapter stops, even
// though the hub is still alive and never answers.
func TestPermissionHandlerReleasedByAdapterStop(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	handler := h.client.createCfg.OnPermissionRequest
	res := make(chan rpc.PermissionDecision, 1)
	go func() {
		d, _ := handler(&rpc.PermissionRequestShell{FullCommandText: "x"}, copilotsdk.PermissionInvocation{})
		res <- d
	}()
	<-h.asks // parked awaiting a reply; hub still alive

	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("got %q, want user-not-available", d.Kind())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("pending permission handler was not released by Stop")
	}
}

// F6: the permission summary is bounded.
func TestPermissionSummaryIsBounded(t *testing.T) {
	long := strings.Repeat("é", maxSummaryRunes*3)
	req := toDomainRequest(&rpc.PermissionRequestShell{FullCommandText: long}, copilotsdk.PermissionInvocation{})
	if n := utf8.RuneCountInString(req.Summary); n > maxSummaryRunes+1 {
		t.Fatalf("summary has %d runes, want <= %d", n, maxSummaryRunes+1)
	}
	if !utf8.ValidString(req.Summary) {
		t.Fatal("truncation split a rune")
	}
	if !strings.HasSuffix(req.Summary, "…") {
		t.Error("truncated summary should end with an ellipsis")
	}
	short := toDomainRequest(&rpc.PermissionRequestShell{FullCommandText: "echo hi"}, copilotsdk.PermissionInvocation{})
	if short.Summary != "echo hi" {
		t.Errorf("short summary altered: %q", short.Summary)
	}
}

// F7: a zero/expired budget must not turn an idle latch into a false negative.
func TestLatchQuiesceZeroTimeoutWhenIdle(t *testing.T) {
	for i := 0; i < 200; i++ {
		l := newLatch()
		if !l.quiesce(0) {
			t.Fatal("idle latch reported not-drained with a zero timeout")
		}
	}
}

// F8: during shutdown an event is still delivered when ingress has room.
func TestCallbackDeliversDuringShutdownWhenIngressHasRoom(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.ad.stopOnce.Do(func() { close(h.ad.stop) }) // shutdown begun, latch still open
	for i := 0; i < 50; i++ {
		h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnEndData{TurnID: "t"}})
	}
	if got := len(h.ingress); got != cap(h.ingress) {
		t.Fatalf("ingress holds %d events, want it filled to %d (events with room must not be dropped)", got, cap(h.ingress))
	}
}

// F11: protocol pin is explicit rather than following the SDK constant.
func TestRequiredProtocolVersionIsPinned(t *testing.T) {
	if RequiredProtocolVersion != 3 {
		t.Fatalf("RequiredProtocolVersion = %d, want 3 (ADR MVP-D10)", RequiredProtocolVersion)
	}
	if copilotsdk.SDKProtocolVersion != RequiredProtocolVersion {
		t.Fatalf("SDK protocol %d != pinned %d: re-validate before bumping the SDK", copilotsdk.SDKProtocolVersion, RequiredProtocolVersion)
	}
}
