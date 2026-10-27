package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/tui"
)

// syncBuf is a goroutine-safe TUI output sink.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func waitOutput(t *testing.T, out *syncBuf, sub string) {
	t.Helper()
	deadline := time.Now().Add(waitFor)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), sub) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("TUI output never contained %q; got:\n%s", sub, out.String())
}

func TestTUIAttachesToHubAsTheSingleLocalClient(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	in, inW := io.Pipe()
	out := &syncBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() }) // release bubbletea's input reader
	o.TUI = TUIOptions{Input: in, Output: out, Width: 80, Height: 24}
	a := startApp(t, o)

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	deadline := time.Now().Add(waitFor)
	for {
		st, err := a.Hub().Stats(ctx)
		if err == nil && st.TUIAttached {
			if st.Remote != 0 {
				t.Errorf("remote clients = %d; want 0", st.Remote)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the TUI never attached to the hub (stats %+v, err %v)", st, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// The credential panel is shown first and covers the transcript; Esc
	// dismisses it (the plaintext is then dropped from the model).
	waitOutput(t, out, "shown once")
	if _, err := inW.Write([]byte("\x1b")); err != nil {
		t.Fatal(err)
	}
	f.fake().emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "tui sees this"})
	waitOutput(t, out, "tui sees this")
}

const (
	testStopBudget = 100 * time.Millisecond
	testKillWait   = 100 * time.Millisecond
	testSlack      = 1 * time.Second
)

// fakeUI is a localUI with no Bubble Tea program: only the stop machinery and
// its seams, so the stop budgets can be exercised with short timers.
func fakeUI(done chan struct{}, quit, kill func()) *localUI {
	return &localUI{
		done:       done,
		feed:       tui.NewFeed(tui.FeedOptions{}),
		quit:       quit,
		kill:       kill,
		stopBudget: testStopBudget,
		killWait:   testKillWait,
	}
}

// within runs fn and fails the test if it has not returned within limit.
func within(t *testing.T, limit time.Duration, what string, fn func()) time.Duration {
	t.Helper()
	start := time.Now()
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		fn()
	}()
	select {
	case <-finished:
	case <-time.After(limit):
		t.Fatalf("%s did not complete within %s", what, limit)
	}
	return time.Since(start)
}

// uiHub is a running hub with no agent, the minimum startUI attaches to.
func uiHub(t *testing.T) *hub.Hub {
	t.Helper()
	h, err := hub.New(hub.Config{SessionID: hubSessionID, Agent: &lateAgent{}})
	if err != nil {
		t.Fatal(err)
	}
	ran := make(chan struct{})
	go func() {
		defer close(ran)
		_ = h.Run(context.Background())
	}()
	t.Cleanup(func() {
		h.Stop()
		<-ran
	})
	return h
}

func uiTestOptions(t *testing.T) (TUIOptions, *syncBuf) {
	t.Helper()
	in, inW := io.Pipe()
	out := &syncBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	return TUIOptions{Input: in, Output: out, Width: 80, Height: 24}, out
}

func TestStartUIPreCancelledContextStartsNothing(t *testing.T) {
	h := uiHub(t)
	o, out := uiTestOptions(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var ui *localUI
	var err error
	within(t, waitFor, "startUI", func() { ui, err = startUI(ctx, h, o, tui.CredentialsMsg{}, nil) })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("startUI error = %v; want context.Canceled", err)
	}
	if ui != nil {
		t.Fatal("startUI returned a UI for a cancelled start context")
	}
	if got := out.String(); got != "" {
		t.Errorf("a program was launched for a cancelled start context; output:\n%s", got)
	}
}

// handshakeCtx reports cancellation only once its Done channel has been
// consumed, so cancellation lands exactly at the ready handshake.
type handshakeCtx struct {
	context.Context
	closed    chan struct{}
	consulted atomic.Bool
}

func newHandshakeCtx() *handshakeCtx {
	c := &handshakeCtx{Context: context.Background(), closed: make(chan struct{})}
	close(c.closed)
	return c
}

func (c *handshakeCtx) Done() <-chan struct{} { c.consulted.Store(true); return c.closed }

func (c *handshakeCtx) Err() error {
	if c.consulted.Load() {
		return context.Canceled
	}
	return nil
}

func TestStartUICancelledDuringReadyHandshakeCleansUp(t *testing.T) {
	h := uiHub(t)
	o, _ := uiTestOptions(t)

	var ui *localUI
	var err error
	within(t, waitFor, "startUI", func() { ui, err = startUI(newHandshakeCtx(), h, o, tui.CredentialsMsg{}, nil) })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("startUI error = %v; want context.Canceled", err)
	}
	if ui != nil {
		t.Fatal("startUI returned a UI after the start context was cancelled")
	}
	deadline := time.Now().Add(waitFor)
	for {
		st, serr := h.Stats(context.Background())
		if serr == nil && !st.TUIAttached {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the cancelled startup left the TUI attached (stats %+v, err %v)", st, serr)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestStartUIFeedStartFailureCleansUpBeforeWatcherRegistered(t *testing.T) {
	h, err := hub.New(hub.Config{SessionID: hubSessionID, Agent: &lateAgent{}})
	if err != nil {
		t.Fatal(err)
	}
	h.Stop() // never run: the feed's dial fails
	o, _ := uiTestOptions(t)

	var ui *localUI
	within(t, waitFor, "startUI", func() { ui, err = startUI(context.Background(), h, o, tui.CredentialsMsg{}, nil) })
	if err == nil || ui != nil {
		t.Fatalf("startUI = (%v, %v); want a nil UI and the feed's error", ui, err)
	}
}

func TestStopWithoutWatcherDoesNotWaitForIt(t *testing.T) {
	done := make(chan struct{})
	u := fakeUI(done, func() { close(done) }, func() {})
	if u.watchDone != nil {
		t.Fatal("watchDone must be nil until the UI watcher is registered")
	}
	within(t, testSlack, "stop", func() {
		if err := u.stop(); err != nil {
			t.Errorf("stop: %v", err)
		}
	})
}

func TestStopWaitsForRegisteredWatcher(t *testing.T) {
	done := make(chan struct{})
	u := fakeUI(done, func() { close(done) }, func() {})
	u.watchDone = make(chan struct{})
	var watcherFinished atomic.Bool
	go func() {
		<-done
		time.Sleep(50 * time.Millisecond)
		watcherFinished.Store(true)
		close(u.watchDone)
	}()
	within(t, testSlack, "stop", func() { _ = u.stop() })
	if !watcherFinished.Load() {
		t.Fatal("stop returned before the UI watcher recorded the runtime error")
	}
}

func TestFailureIsNotReadBeforeDone(t *testing.T) {
	u := &localUI{done: make(chan struct{})}
	got := make(chan error, 1)
	go func() { got <- u.failure() }()
	if err := <-got; err != nil {
		t.Fatalf("failure before done = %v; want nil", err)
	}
	u.err = errors.New("boom")
	close(u.done)
	if err := u.failure(); err == nil {
		t.Fatal("failure after done lost the program error")
	}
}

func TestStopStalledBeforeQuitEscalatesToKill(t *testing.T) {
	done := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	var killed atomic.Bool
	u := fakeUI(done,
		func() { <-release }, // a loop that never receives Quit
		func() { killed.Store(true); close(done) })

	var err error
	elapsed := within(t, testStopBudget+testKillWait+testSlack, "stop", func() { err = u.stop() })
	if err != nil {
		t.Errorf("stop: %v; the kill released the program, so the stop succeeded", err)
	}
	if !killed.Load() {
		t.Error("stop never escalated to Kill")
	}
	if elapsed > testStopBudget+testKillWait+testSlack/2 {
		t.Errorf("stop took %s; want about the stop budget (%s)", elapsed, testStopBudget)
	}
}

func TestStopWithBlockingKillIsStillBounded(t *testing.T) {
	done := make(chan struct{}) // never closed: the wedged program does not leave
	release := make(chan struct{})
	t.Cleanup(func() { close(release) }) // frees the intentionally blocked kill
	u := fakeUI(done, func() { <-release }, func() { <-release })

	var err error
	within(t, testStopBudget+testKillWait+testSlack, "stop with a kill that blocks", func() { err = u.stop() })
	if err == nil {
		t.Fatal("stop of a wedged TUI with a blocking kill returned nil")
	}
}

func TestStopWedgedTUIIsBoundedAndReportsAnError(t *testing.T) {
	done := make(chan struct{}) // never closed: Kill does not release it
	var killed atomic.Bool
	u := fakeUI(done, func() {}, func() { killed.Store(true) })

	var err error
	within(t, testStopBudget+testKillWait+testSlack, "stop", func() { err = u.stop() })
	if err == nil {
		t.Fatal("stop of a wedged TUI returned nil")
	}
	if !killed.Load() {
		t.Error("stop never killed the wedged TUI")
	}
	if again := u.stop(); again != err {
		t.Errorf("second stop = %v; want the same error %v", again, err)
	}
}

func tuiOptionsForApp(t *testing.T, o *Options) {
	t.Helper()
	in, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	o.TUI = TUIOptions{Input: in, Output: &syncBuf{}, Width: 80, Height: 24}
}

func TestRuntimeTUIFailureIsJoinedIntoAppErr(t *testing.T) {
	o, _ := baseOptions(t, newFactory())
	tuiOptionsForApp(t, &o)
	a := startApp(t, o)

	a.ui.prog.Kill() // not issued by stop: a genuine runtime failure
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("the app did not shut down after the TUI failed")
	}
	err := a.Err()
	if err == nil || !strings.Contains(err.Error(), "tui:") {
		t.Fatalf("App.Err() = %v; want it to carry the TUI failure", err)
	}
}

func TestWedgedTUIShutdownIsBoundedAndSurfacesInAppErr(t *testing.T) {
	o, _ := baseOptions(t, newFactory())
	tuiOptionsForApp(t, &o)
	a := startApp(t, o)

	prog := a.ui.prog
	t.Cleanup(prog.Kill) // release the real program the wedged seams hide
	a.ui.quit = func() {}
	a.ui.kill = func() {}
	a.ui.stopBudget = testStopBudget
	a.ui.killWait = testKillWait

	a.Shutdown("test")
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("the app did not finish shutting down with a wedged TUI")
	}
	if a.Err() == nil {
		t.Fatal("App.Err() is nil although the TUI could not be stopped")
	}
}
