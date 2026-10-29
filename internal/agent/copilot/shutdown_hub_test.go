package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// ---- 067.007-T: shutdown steps 2-4 against the REAL adapter ----
//
// The REAL adapter (latch, Stop, ForceStop escalation, permission bridge) and a
// REAL hub run under the REAL shutdown executor; only the SDK client/session
// (the existing fakeClient/fakeSession seam) and the transport gate are fakes.
// fakeClient.childAlive models the CLI child process, so "no stranded CLI
// child" is an assertion, not a hope.

type rigGate struct {
	revokes, closes atomic.Int32
}

func (g *rigGate) RevokeCredentials(context.Context) error { g.revokes.Add(1); return nil }
func (g *rigGate) CloseListener(context.Context) error     { g.closes.Add(1); return nil }

type frameLog struct {
	mu    sync.Mutex
	kinds []string
}

func (f *frameLog) add(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.kinds = append(f.kinds, s)
}

func (f *frameLog) has(s string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, k := range f.kinds {
		if k == s {
			return true
		}
	}
	return false
}

type lockedBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func (l *lockedBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.String()
}

type shutdownRig struct {
	hub    *hub.Hub
	h      *harness
	gate   *rigGate
	frames *frameLog
	logs   *lockedBuf
	stopC  chan struct{}
}

func newShutdownRig(t *testing.T) *shutdownRig {
	t.Helper()
	hb, err := hub.New(hub.Config{SessionID: "sess-1"})
	if err != nil {
		t.Fatal(err)
	}
	r := &shutdownRig{hub: hb, gate: &rigGate{}, frames: &frameLog{}, logs: &lockedBuf{}, stopC: make(chan struct{})}
	r.h = newHarness(t, func(c *Config, _ *fakeClient) {
		c.Ingress = hb.Ingress()
		c.Asks = hb.Asks()
		c.HubDone = hb.Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- hb.Run(ctx) }()
	t.Cleanup(func() {
		close(r.stopC)
		cancel()
		select {
		case <-runDone:
		case <-time.After(intWait):
			t.Error("hub Run did not return")
		}
	})
	if _, err := r.h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	tui, err := hb.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientTUI, Buffer: 128})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for {
			select {
			case f := <-tui.Frames():
				name := string(f.Kind())
				if f.Kind() == hub.KindPermissionResolved {
					var env hub.Envelope
					var p hub.PermissionResolvedPayload
					_ = json.Unmarshal(f.Bytes(), &env)
					_ = json.Unmarshal(env.Payload, &p)
					name += ":" + p.Reason
				}
				r.frames.add(name)
			case <-r.stopC:
				return
			}
		}
	}()
	return r
}

func (r *shutdownRig) newShutdown(t *testing.T, mutate func(*hub.ShutdownConfig)) *hub.Shutdown {
	t.Helper()
	cfg := hub.ShutdownConfig{
		Hub: r.hub, Agent: r.h.ad, Gate: r.gate, Reason: "test",
		Timeouts: hub.ShutdownTimeouts{Permissions: 2 * time.Second, Gate: time.Second, Announce: time.Second,
			Abort: time.Second, Idle: 2 * time.Second, Stop: 3 * time.Second, Drain: 200 * time.Millisecond},
		Logger: slog.New(slog.NewTextHandler(r.logs, nil)),
	}
	if mutate != nil {
		mutate(&cfg)
	}
	s, err := hub.NewShutdown(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func waitDone(t *testing.T, s *hub.Shutdown) {
	t.Helper()
	select {
	case <-s.Done():
	case <-time.After(3 * intWait):
		t.Fatal("shutdown did not finish")
	}
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(intWait)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func (r *shutdownRig) emit(ev rpc.SessionEventData) {
	r.h.client.session.emit(copilotsdk.SessionEvent{Data: ev})
}

// activeTurn starts a turn and waits until the hub has folded it.
func (r *shutdownRig) activeTurn(t *testing.T) {
	t.Helper()
	r.emit(&rpc.AssistantTurnStartData{TurnID: "t1"})
	eventually(t, "the hub to fold turn.start", func() bool {
		s, err := r.hub.Stats(context.Background())
		return err == nil && s.Head >= 1
	})
}

func idleAborted() rpc.SessionEventData {
	aborted := true
	return &rpc.SessionIdleData{Aborted: &aborted}
}

func (r *shutdownRig) assertNoStrandedChild(t *testing.T) {
	t.Helper()
	eventually(t, "the CLI child to be terminated", func() bool { return !r.h.client.childAlive.Load() })
}

func TestShutdownAbortsWaitsForPositiveIdleThenStopsTheRealAdapter(t *testing.T) {
	r := newShutdownRig(t)
	if !r.h.client.childAlive.Load() {
		t.Fatal("harness: the CLI child never came up")
	}
	r.activeTurn(t)

	var abortsAtStop atomic.Int32
	r.h.client.onStop = func() { abortsAtStop.Store(r.h.client.session.aborts.Load()) }
	idleRelease := make(chan struct{})
	r.h.client.session.abortHook = func() {
		go func() { // the SDK answers an abort with an idle event, asynchronously
			<-idleRelease
			r.emit(idleAborted())
		}()
	}
	s := r.newShutdown(t, func(c *hub.ShutdownConfig) { c.Timeouts.Idle = 10 * time.Second })
	s.Start(context.Background())

	eventually(t, "Abort", func() bool { return r.h.client.session.aborts.Load() == 1 })
	time.Sleep(100 * time.Millisecond)
	if n := r.h.client.stops.Load(); n != 0 {
		t.Fatalf("Client.Stop ran (%d) before the positive idle signal: stale-flag shutdown", n)
	}
	close(idleRelease)
	waitDone(t, s)

	if err := s.Err(); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	if abortsAtStop.Load() != 1 {
		t.Fatalf("Client.Stop ran after %d aborts, want Abort first", abortsAtStop.Load())
	}
	c := r.h.client
	if c.stops.Load() != 1 || c.forceStops.Load() != 0 {
		t.Fatalf("Stop=%d ForceStop=%d, want a clean graceful stop", c.stops.Load(), c.forceStops.Load())
	}
	if c.session.disconnects.Load() != 1 {
		t.Fatalf("session disconnected %d times", c.session.disconnects.Load())
	}
	if r.gate.revokes.Load() != 1 || r.gate.closes.Load() != 1 {
		t.Fatalf("gate revokes=%d closes=%d", r.gate.revokes.Load(), r.gate.closes.Load())
	}
	r.assertNoStrandedChild(t)
}

func TestStopExceedingItsDeadlineTriggersForceStopAndNoChildIsStranded(t *testing.T) {
	r := newShutdownRig(t)
	r.h.client.stopBlock = make(chan struct{}) // the runtime ignores the graceful stop
	t.Cleanup(func() { close(r.h.client.stopBlock) })
	s := r.newShutdown(t, func(c *hub.ShutdownConfig) { c.Timeouts.Stop = 300 * time.Millisecond })
	s.Start(context.Background())
	waitDone(t, s)

	eventually(t, "ForceStop", func() bool { return r.h.client.forceStops.Load() >= 1 })
	if n := r.h.client.forceStops.Load(); n != 1 {
		t.Fatalf("ForceStop ran %d times, want exactly once (single-flight)", n)
	}
	r.assertNoStrandedChild(t)
	if s.Err() == nil {
		t.Fatal("a stop that blew its deadline must be reported, not swallowed")
	}
}

// Review finding A: a wedged graceful stop and a cancelled shutdown context must
// still end in a completed ForceStop BEFORE the executor reports Done.
func TestCancelledShutdownContextStillWaitsForForceStopToKillTheChild(t *testing.T) {
	r := newShutdownRig(t)
	c := r.h.client
	c.stopBlock = make(chan struct{})  // the runtime ignores the graceful stop
	c.forceBlock = make(chan struct{}) // and ForceStop takes a while to kill the child
	var once sync.Once
	releaseForce := func() { once.Do(func() { close(c.forceBlock) }) }
	t.Cleanup(func() { close(c.stopBlock); releaseForce() })
	s := r.newShutdown(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	eventually(t, "the graceful Client.Stop to wedge", func() bool { return c.stops.Load() == 1 })
	cancel()
	eventually(t, "ForceStop to start", func() bool { return c.forceStops.Load() == 1 })
	select {
	case <-s.Done():
		t.Fatal("Done closed while ForceStop was still killing the CLI child")
	case <-time.After(300 * time.Millisecond):
	}
	if !c.childAlive.Load() {
		t.Fatal("harness: the child died before ForceStop completed")
	}
	releaseForce()
	waitDone(t, s)
	if c.childAlive.Load() {
		t.Fatal("shutdown finished with the CLI child alive")
	}
	if s.Err() == nil {
		t.Fatal("a cancelled shutdown must report its interruption")
	}
}

func TestShutdownThatCannotConfirmForceStopReportsOrphanRisk(t *testing.T) {
	r := newShutdownRig(t)
	c := r.h.client
	c.stopBlock = make(chan struct{})
	c.forceBlock = make(chan struct{}) // ForceStop never returns
	t.Cleanup(func() { close(c.stopBlock); close(c.forceBlock) })
	// Force (300ms) must not undercut the adapter's own ForceStop wait, so the
	// adapter bound is lowered below it: the hub then gives up on a wedged
	// ForceStop and reports the orphan risk.
	r.h.ad.cfg.StopTimeout = 200 * time.Millisecond
	s := r.newShutdown(t, func(c *hub.ShutdownConfig) {
		c.Timeouts.Stop = 100 * time.Millisecond
		c.Timeouts.Force = 300 * time.Millisecond
	})
	s.Start(context.Background())
	waitDone(t, s)
	if !errors.Is(s.Err(), hub.ErrOrphanRisk) {
		t.Fatalf("Err() = %v, want an orphan-risk report", s.Err())
	}
	if !c.childAlive.Load() {
		t.Fatal("harness: the child should still be alive")
	}
}

// Review cycle 2 N1: the adapter gives up on ForceStop at its own StopTimeout
// (1s in the harness), BEFORE the hub's hard budget (Stop+Force) is spent. The
// adapter then returns a "cleanup incomplete" error; the hub must treat that
// as an unconfirmed reap (orphan risk), not as an ordinary returned error, and
// must not close Done before the adapter has answered.
func TestForceStopOutlastingTheAdapterBoundReportsOrphanRiskNotAnOrdinaryError(t *testing.T) {
	r := newShutdownRig(t)
	c := r.h.client
	c.stopBlock = make(chan struct{})
	c.forceBlock = make(chan struct{}) // ForceStop never returns
	t.Cleanup(func() { close(c.stopBlock); close(c.forceBlock) })
	s := r.newShutdown(t, func(c *hub.ShutdownConfig) {
		c.Timeouts.Stop = 100 * time.Millisecond
		c.Timeouts.Force = 5 * time.Second // outlasts the adapter's own 1s ForceStop bound
	})
	start := time.Now()
	s.Start(context.Background())
	select {
	case <-s.Done():
		t.Fatalf("Done closed after %v, before the adapter's ForceStop wait ended", time.Since(start))
	case <-time.After(500 * time.Millisecond):
	}
	waitDone(t, s)
	if !errors.Is(s.Err(), hub.ErrOrphanRisk) {
		t.Fatalf("Err() = %v, want an orphan-risk report for an unconfirmed ForceStop", s.Err())
	}
	if !errors.Is(s.Err(), agent.ErrCleanupIncomplete) {
		t.Fatalf("Err() = %v, want the adapter's cleanup-incomplete cause kept in the chain", s.Err())
	}
	if !c.childAlive.Load() {
		t.Fatal("harness: the child should still be alive")
	}
}

// The hub's default Force budget must outlast the adapter's default ForceStop
// wait (its StopTimeout), or the hub would give up on a stop the adapter is
// still waiting out.
func TestDefaultForceBudgetOutlastsTheAdapterDefaultStopTimeout(t *testing.T) {
	if hub.DefaultShutdownForce < DefaultStopTimeout+time.Second {
		t.Fatalf("hub default Force %v must exceed the adapter's default StopTimeout %v by a margin", hub.DefaultShutdownForce, DefaultStopTimeout)
	}
}

// ---- 074.004: bounded recheck and turn-scoped permission asks ----

// A Stop after a failed Start whose ForceStop is still wedged must report
// ErrCleanupIncomplete at once: the wedge was already waited out, so the
// recheck is a non-blocking look, not another StopTimeout-long wait.
func TestStopAfterFailedStartRecheckForceIsNonBlocking(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StopTimeout = 200 * time.Millisecond
		f.forceBlock = make(chan struct{})
		f.status = &copilotsdk.GetStatusResponse{Version: "1.0.0", ProtocolVersion: 3}
	})
	defer close(h.client.forceBlock)
	if _, err := h.ad.Start(context.Background()); !errors.Is(err, agent.ErrCleanupIncomplete) {
		t.Fatalf("Start error = %v, want ErrCleanupIncomplete", err)
	}
	start := time.Now()
	err := h.ad.Stop(context.Background())
	if !errors.Is(err, agent.ErrCleanupIncomplete) {
		t.Fatalf("Stop = %v, want ErrCleanupIncomplete", err)
	}
	if d := time.Since(start); d >= 50*time.Millisecond {
		t.Fatalf("Stop took %v; the force recheck must not wait again", d)
	}
}

// turnRig is a real hub and the real adapter, with the ask channel proxied so
// the test sees the TurnID each ask carries on its way to the hub.
type turnRig struct {
	hub   *hub.Hub
	h     *harness
	mu    sync.Mutex
	asked []string
}

func newTurnRig(t *testing.T) *turnRig {
	t.Helper()
	hb, err := hub.New(hub.Config{SessionID: "sess-1"})
	if err != nil {
		t.Fatal(err)
	}
	r := &turnRig{hub: hb}
	proxy := make(chan agent.PermissionAsk, 8)
	r.h = newHarness(t, func(c *Config, _ *fakeClient) {
		c.Ingress = hb.Ingress()
		c.Asks = proxy
		c.HubDone = hb.Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- hb.Run(ctx) }()
	fwdDone := make(chan struct{})
	go func() {
		defer close(fwdDone)
		for {
			select {
			case ask := <-proxy:
				r.mu.Lock()
				r.asked = append(r.asked, ask.Request.TurnID)
				r.mu.Unlock()
				select {
				case hb.Asks() <- ask:
				case <-hb.Done():
					return
				}
			case <-hb.Done():
				return
			}
		}
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-runDone:
		case <-time.After(intWait):
			t.Error("hub Run did not return")
		}
		<-fwdDone
	})
	if _, err := r.h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	// An attached surface keeps asks pending (an empty broadcast set resolves
	// them at once).
	if _, err := hb.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientTUI, Buffer: 128}); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *turnRig) emit(ev rpc.SessionEventData) {
	r.h.client.session.emit(copilotsdk.SessionEvent{Data: ev})
}

func (r *turnRig) askCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.asked)
}

func (r *turnRig) lastTurnID() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.asked) == 0 {
		return ""
	}
	return r.asked[len(r.asked)-1]
}

// callHandler invokes the registered SDK permission callback on its own
// goroutine and returns the channel carrying its decision.
func (r *turnRig) callHandler() <-chan rpc.PermissionDecision {
	res := make(chan rpc.PermissionDecision, 1)
	handler := r.h.client.createCfg.OnPermissionRequest
	go func() {
		d, _ := handler(&rpc.PermissionRequestShell{FullCommandText: "ls"}, copilotsdk.PermissionInvocation{})
		res <- d
	}()
	return res
}

func awaitUnavailable(t *testing.T, what string, res <-chan rpc.PermissionDecision) {
	t.Helper()
	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("%s: decision = %q, want user-not-available", what, d.Kind())
		}
	case <-time.After(intWait):
		t.Fatalf("%s: the handler never returned", what)
	}
}

func TestPermissionBridgeStampsTurnIDAndCancelsLateAbortedTurnAskThroughRegisteredCallback(t *testing.T) {
	t.Run("abort after turn start", func(t *testing.T) {
		r := newTurnRig(t)
		r.emit(&rpc.AssistantTurnStartData{TurnID: "t1"})
		inflight := r.callHandler()
		eventually(t, "the ask to reach the hub", func() bool { return r.askCount() == 1 })
		if got := r.lastTurnID(); got != "t1" {
			t.Fatalf("ask TurnID = %q, want the active turn t1", got)
		}
		r.emit(idleAborted())
		awaitUnavailable(t, "in-flight ask of the aborted turn", inflight)

		late := r.callHandler() // a callback reaching currentTurn() after the abort
		awaitUnavailable(t, "late ask of the aborted turn", late)
		if n := r.askCount(); n != 1 {
			t.Fatalf("%d asks reached the hub; the late aborted-turn ask must not be enqueued", n)
		}

		// The next turn's ask is accepted even before the hub has folded its
		// turn.start.
		r.emit(&rpc.AssistantTurnStartData{TurnID: "t2"})
		r.callHandler()
		eventually(t, "the next turn's ask to reach the hub", func() bool { return r.askCount() == 2 })
		if got := r.lastTurnID(); got != "t2" {
			t.Fatalf("next ask TurnID = %q, want t2", got)
		}
		eventually(t, "the next turn's ask to be registered", func() bool {
			s, err := r.hub.Stats(context.Background())
			return err == nil && s.PendingPermissions == 1
		})
	})
	t.Run("abort with no turn start", func(t *testing.T) {
		r := newTurnRig(t)
		r.emit(idleAborted())
		awaitUnavailable(t, "ask after an abort with no turn start", r.callHandler())
		if n := r.askCount(); n != 0 {
			t.Fatalf("%d asks reached the hub; an ask after an abort must not be enqueued", n)
		}
	})
}

func TestShutdownAggregatesAndLogsAdapterStopErrors(t *testing.T) {
	r := newShutdownRig(t)
	r.h.client.stopErr = errors.New("cli exited with status 7")
	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitDone(t, s)

	if err := s.Err(); err == nil || !strings.Contains(err.Error(), "cli exited with status 7") {
		t.Fatalf("Err() = %v, want the adapter's stop error aggregated", err)
	}
	if !strings.Contains(r.logs.String(), "cli exited with status 7") {
		t.Fatalf("the stop error was not logged:\n%s", r.logs.String())
	}
	// The failed graceful stop escalates to ForceStop, which still kills the child.
	eventually(t, "ForceStop after a failed Stop", func() bool { return r.h.client.forceStops.Load() == 1 })
	r.assertNoStrandedChild(t)
}

func TestShutdownWithABlockedPermissionHandlerDoesNotDeadlockOrForceStop(t *testing.T) {
	r := newShutdownRig(t)
	res := make(chan rpc.PermissionDecision, 1)
	handler := r.h.client.createCfg.OnPermissionRequest
	go func() {
		d, _ := handler(&rpc.PermissionRequestShell{FullCommandText: "rm -rf /"}, copilotsdk.PermissionInvocation{})
		res <- d
	}()
	eventually(t, "the request to be broadcast", func() bool { return r.frames.has("permission.requested") })

	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitDone(t, s)

	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("handler returned %q", d.Kind())
		}
	default:
		t.Fatal("the parked handler was never released")
	}
	if err := s.Err(); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	if n := r.h.client.forceStops.Load(); n != 0 {
		t.Fatalf("ForceStop ran %d times: the handler deadlock was not avoided", n)
	}
	// Released by step 0 through the hub (resolved, shutdown reason), not by
	// the adapter's own stop path (aborted).
	eventually(t, "permission.resolved with the shutdown reason", func() bool { return r.frames.has("permission.resolved:shutdown") })
	if r.frames.has("permission.aborted") {
		t.Fatal("the request was aborted instead of failed through the hub at step 0")
	}
	r.assertNoStrandedChild(t)
}

// ---- 077.003-T: ForceStop waits out an in-flight Client.Stop's SDK lock ----

func TestForceStopExtendsWaitForInFlightClientStopHoldingSDKLock(t *testing.T) {
	t.Run("force wait is extended only while a Client.Stop holds the lock", func(t *testing.T) {
		a := newAdapter(Config{StopTimeout: 100 * time.Millisecond, SDKStopLockTimeout: 400 * time.Millisecond}, newFakeClient())
		if got := a.forceStopWait(); got != 100*time.Millisecond {
			t.Fatalf("idle forceStopWait = %v, want StopTimeout", got)
		}
		a.mu.Lock()
		a.clientStopActive, a.clientStopStart = true, time.Now()
		a.mu.Unlock()
		if got := a.forceStopWait(); got < 400*time.Millisecond || got > 500*time.Millisecond {
			t.Fatalf("forceStopWait during Client.Stop = %v, want ~Stop+lock (500ms)", got)
		}
		// Hurried Client.Stop: most of the lock allowance already elapsed.
		a.mu.Lock()
		a.clientStopStart = time.Now().Add(-390 * time.Millisecond)
		a.mu.Unlock()
		if got := a.forceStopWait(); got < 100*time.Millisecond || got > 150*time.Millisecond {
			t.Fatalf("forceStopWait after the allowance elapsed = %v, want ~StopTimeout", got)
		}
		a.mu.Lock()
		a.clientStopStart = time.Now().Add(-time.Hour)
		a.mu.Unlock()
		if got := a.forceStopWait(); got != 100*time.Millisecond {
			t.Fatalf("forceStopWait with an expired allowance = %v, want StopTimeout", got)
		}
		if got := a.ForceStopBudget(); got != 500*time.Millisecond {
			t.Fatalf("ForceStopBudget = %v, want 500ms", got)
		}
	})

	t.Run("two-phase Client.Stop contention: ForceStop is awaited past StopTimeout", func(t *testing.T) {
		h := newHarness(t, func(c *Config, f *fakeClient) {
			c.StopTimeout = 100 * time.Millisecond
			c.SDKStopLockTimeout = 600 * time.Millisecond
			f.stopBlock = make(chan struct{}) // Client.Stop holds the lock throughout
			f.forceBlock = make(chan struct{})
		})
		if _, err := h.ad.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { close(h.client.stopBlock) })
		// ForceStop is only able to proceed ~300ms after it was requested,
		// which is past StopTimeout but inside the lock allowance.
		go func() {
			for h.client.forceStops.Load() == 0 {
				time.Sleep(time.Millisecond)
			}
			time.Sleep(300 * time.Millisecond)
			close(h.client.forceBlock)
		}()
		err := h.ad.Stop(context.Background())
		if errors.Is(err, agent.ErrCleanupIncomplete) {
			t.Fatalf("Stop = %v; the wedge was within the SDK lock allowance and must not be reported incomplete", err)
		}
		if h.client.childAlive.Load() {
			t.Fatal("the CLI child was stranded")
		}
	})

	t.Run("no Client.Stop in flight keeps the plain StopTimeout bound", func(t *testing.T) {
		h := newHarness(t, func(c *Config, f *fakeClient) {
			c.StopTimeout = 100 * time.Millisecond
			c.SDKStopLockTimeout = 600 * time.Millisecond
			f.status = &copilotsdk.GetStatusResponse{Version: "1.0.0", ProtocolVersion: 3}
			f.forceBlock = make(chan struct{})
		})
		defer close(h.client.forceBlock)
		start := time.Now()
		if _, err := h.ad.Start(context.Background()); !errors.Is(err, agent.ErrCleanupIncomplete) {
			t.Fatalf("Start error = %v, want ErrCleanupIncomplete", err)
		}
		if d := time.Since(start); d > 450*time.Millisecond {
			t.Fatalf("ForceStop wait took %v without a Client.Stop in flight; the lock allowance must not apply", d)
		}
	})

	t.Run("real client defaults the lock allowance independently of StopTimeout", func(t *testing.T) {
		real := newAdapter(Config{StopTimeout: time.Second}, realClient{})
		if real.cfg.SDKStopLockTimeout != DefaultSDKStopLockTimeout || DefaultSDKStopLockTimeout != 20*time.Second {
			t.Fatalf("real-client SDKStopLockTimeout = %v, want %v (20s)", real.cfg.SDKStopLockTimeout, DefaultSDKStopLockTimeout)
		}
		if got := real.ForceStopBudget(); got != time.Second+DefaultSDKStopLockTimeout {
			t.Fatalf("real ForceStopBudget = %v", got)
		}
		fake := newAdapter(Config{}, newFakeClient())
		if fake.cfg.SDKStopLockTimeout != 0 {
			t.Fatalf("fake-client SDKStopLockTimeout = %v, want 0", fake.cfg.SDKStopLockTimeout)
		}
		if neg := newAdapter(Config{SDKStopLockTimeout: -time.Second}, newFakeClient()); neg.cfg.SDKStopLockTimeout != 0 {
			t.Fatalf("negative allowance = %v, want clamped to 0", neg.cfg.SDKStopLockTimeout)
		}
	})

	t.Run("NewShutdown rejects Force below the ForceStopBudget", func(t *testing.T) {
		hb, err := hub.New(hub.Config{SessionID: "s"})
		if err != nil {
			t.Fatal(err)
		}
		ad := newAdapter(Config{StopTimeout: 5 * time.Second}, realClient{}) // budget 25s
		mk := func(force time.Duration) error {
			_, err := hub.NewShutdown(hub.ShutdownConfig{Hub: hb, Agent: ad, Gate: &rigGate{},
				Timeouts: hub.ShutdownTimeouts{Force: force}})
			return err
		}
		if err := mk(10 * time.Second); err == nil {
			t.Fatal("Force 10s < ForceStopBudget 25s must be rejected")
		}
		if err := mk(25 * time.Second); err != nil {
			t.Fatalf("Force == ForceStopBudget must be accepted: %v", err)
		}
		if err := mk(0); err != nil {
			t.Fatalf("default Force must scale to the budget: %v", err)
		}
	})
}
