package copilot

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
)

// Copilot review: a failed Start must not hang when ForceStop blocks.
func TestFailedStartReturnsWhenForceStopBlocks(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 100 * time.Millisecond
		f.forceBlock = make(chan struct{})
		f.status = &copilotsdk.GetStatusResponse{Version: "1.0.0", ProtocolVersion: 3}
	})
	defer close(h.client.forceBlock)

	res := make(chan error, 1)
	go func() { _, err := h.ad.Start(context.Background()); res <- err }()
	select {
	case err := <-res:
		if err == nil || !strings.Contains(err.Error(), "below the validated floor") {
			t.Fatalf("Start error = %v, want the original startup error preserved", err)
		}
		if !strings.Contains(err.Error(), "cleanup") {
			t.Errorf("Start error = %v, want the cleanup timeout joined", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("failed Start hung on a blocking ForceStop")
	}
}

// Copilot review: Stop's fallback ForceStop is bounded and Stop still
// publishes its result so concurrent callers are released.
func TestStopBoundedWhenForceStopBlocks(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 100 * time.Millisecond
		f.forceBlock = make(chan struct{})
	})
	defer close(h.client.forceBlock)
	h.client.session.disconnectBlock = make(chan struct{})
	defer close(h.client.session.disconnectBlock)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	res := make(chan error, 1)
	go func() { res <- h.ad.Stop(context.Background()) }()
	select {
	case err := <-res:
		if err == nil || !strings.Contains(err.Error(), "cleanup") {
			t.Fatalf("Stop error = %v, want incomplete cleanup reported", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stop hung on a blocking ForceStop")
	}
	second := make(chan struct{})
	go func() { _ = h.ad.Stop(context.Background()); close(second) }()
	select {
	case <-second:
	case <-time.After(2 * time.Second):
		t.Fatal("stopDone was never closed; concurrent Stop callers would hang")
	}
}

// Review finding A: once the caller's context is done, graceful teardown is
// over. Stop must go straight to ForceStop and must not launch a graceful
// Disconnect or Client.Stop that would race the force kill.
func TestStopWithADoneContextSkipsGracefulStepsAndForceStops(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := h.ad.Stop(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop error = %v, want context.Canceled", err)
	}
	time.Sleep(50 * time.Millisecond) // a (wrongly) launched graceful call would have run by now
	if n := h.client.session.disconnects.Load(); n != 0 {
		t.Errorf("Disconnect ran %d times after the context was done", n)
	}
	if n := h.client.stops.Load(); n != 0 {
		t.Errorf("graceful Client.Stop ran %d times after the context was done", n)
	}
	if n := h.client.forceStops.Load(); n != 1 {
		t.Errorf("ForceStop ran %d times, want 1", n)
	}
	if h.client.childAlive.Load() {
		t.Error("the CLI child survived Stop")
	}
}

// A graceful step that failed or timed out means the runtime is going to be
// force-killed: the remaining graceful steps are skipped, not attempted.
func TestStopSkipsClientStopAfterDisconnectTimesOut(t *testing.T) {
	h := newHarness(t, func(c *Config, _ *fakeClient) { c.StopTimeout = 100 * time.Millisecond })
	h.client.session.disconnectBlock = make(chan struct{})
	defer close(h.client.session.disconnectBlock)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Stop(context.Background()); err == nil {
		t.Fatal("a Disconnect that timed out must be reported")
	}
	if n := h.client.stops.Load(); n != 0 {
		t.Errorf("graceful Client.Stop ran %d times after Disconnect timed out, want straight to ForceStop", n)
	}
	if n := h.client.forceStops.Load(); n != 1 {
		t.Errorf("ForceStop ran %d times, want 1", n)
	}
}

// Copilot review: Stop observes cancellation of the caller's context during
// the first call's teardown, and still force-stops the runtime.
func TestStopObservesContextCancellation(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 10 * time.Second
		f.stopBlock = make(chan struct{})
	})
	defer close(h.client.stopBlock)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan error, 1)
	go func() { res <- h.ad.Stop(ctx) }()
	deadline := time.Now().Add(2 * time.Second)
	for h.client.stops.Load() == 0 { // inside the blocked SDK Stop
		if time.Now().After(deadline) {
			t.Fatal("Stop never reached client.Stop")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	select {
	case err := <-res:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Stop error = %v, want it to wrap context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stop ignored caller cancellation (waited out StopTimeout)")
	}
	if h.client.forceStops.Load() == 0 {
		t.Error("a cancelled Stop must still force-stop the runtime")
	}
}

// ForceStop is single-flight: concurrent cleanup paths (failed Start and Stop)
// share one SDK ForceStop call.
func TestForceStopIsSingleFlight(t *testing.T) {
	h := newHarness(t, nil)
	done := make(chan struct{}, 8)
	for i := 0; i < 8; i++ {
		go func() { _ = h.ad.forceStop(); done <- struct{}{} }()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	if n := h.client.forceStops.Load(); n != 1 {
		t.Fatalf("SDK ForceStop called %d times, want exactly 1", n)
	}
}

func TestStopObservesContextCancellationDuringStart(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StartupTimeout = 10 * time.Second
		c.StopTimeout = 10 * time.Second
		f.createGate = make(chan struct{}) // CreateSession ignores ctx
		f.createEntered = make(chan struct{})
	})
	defer close(h.client.createGate)
	go func() { _, _ = h.ad.Start(context.Background()) }()
	<-h.client.createEntered
	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan error, 1)
	go func() { res <- h.ad.Stop(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case err := <-res:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Stop error = %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stop ignored cancellation while waiting for an in-flight Start")
	}
}
