package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.006-T: shutdown executor ordering with a fake agent ----

// recorder is the ordered log of what the executor and the hub did. Entries
// from the fakes are written only after the frame consumer has caught up with
// the hub's head, so a frame published before a fake call is recorded before
// it: the log order is then causal, not scheduler luck.
type recorder struct {
	mu   sync.Mutex
	ev   []string
	seen atomic.Uint64 // highest seq the frame consumer has recorded
	h    *Hub
}

func (r *recorder) add(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ev = append(r.ev, s)
}

func (r *recorder) events() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.ev)
}

// mark records a fake's invocation after syncing frames. It also records
// HUB-DONE-EARLY if the hub had already stopped: the hub must outlive every
// step before step 6.
func (r *recorder) mark(name string) {
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if s, err := r.h.Stats(ctx); err == nil {
		for r.seen.Load() < s.Head && ctx.Err() == nil {
			time.Sleep(time.Millisecond)
		}
	}
	select {
	case <-r.h.Done():
		r.add("HUB-DONE-EARLY:" + name)
	default:
	}
	r.add(name)
}

// consume records every non-hydrate frame c receives.
func (r *recorder) consume(c *Client, stop <-chan struct{}) {
	go func() {
		for {
			select {
			case f := <-c.Frames():
				if f.Kind() == KindSessionHydrate {
					continue
				}
				name := "frame:" + string(f.Kind())
				if f.Kind() == KindPermissionResolved {
					var p PermissionResolvedPayload
					var env Envelope
					_ = json.Unmarshal(f.Bytes(), &env)
					_ = json.Unmarshal(env.Payload, &p)
					name += ":" + p.Reason
				}
				r.add(name)
				if f.Seq() > r.seen.Load() {
					r.seen.Store(f.Seq())
				}
			case <-stop:
				return
			}
		}
	}()
}

type fakeShutdownAgent struct {
	rec       *recorder
	stopBlock chan struct{} // when non-nil Stop waits for it (or ctx)
	waitFor   chan struct{} // when non-nil Stop waits until this closes: a session that cannot stop while a handler is parked
	stops     atomic.Int32
	stopErr   error
	// abortHook runs inside Abort (after it is recorded): a real session
	// answers an abort with an idle event on ingress.
	abortHook func()
	abortErr  error
	aborts    atomic.Int32
	// stopIgnoresCtx makes Stop block on stopBlock without honouring ctx.
	stopIgnoresCtx bool
}

func (a *fakeShutdownAgent) Start(context.Context) (agent.SessionHandle, error) {
	return agent.SessionHandle{ID: "s"}, nil
}
func (a *fakeShutdownAgent) Send(context.Context, string) error { return nil }
func (a *fakeShutdownAgent) Abort(context.Context) error {
	a.aborts.Add(1)
	a.rec.mark("agent.abort")
	if a.abortHook != nil {
		a.abortHook()
	}
	return a.abortErr
}
func (a *fakeShutdownAgent) Stop(ctx context.Context) error {
	a.stops.Add(1)
	a.rec.mark("agent.stop")
	if a.stopIgnoresCtx && a.stopBlock != nil {
		<-a.stopBlock
		return a.stopErr
	}
	for _, ch := range []chan struct{}{a.waitFor, a.stopBlock} {
		if ch == nil {
			continue
		}
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return a.stopErr
}

type fakeGate struct {
	rec       *recorder
	h         *Hub
	revokeErr error
	closeErr  error
	block     chan struct{} // never closed: a gate call that ignores its context
	admitted  atomic.Int32  // clients the hub admitted from inside a gate call
}

func (g *fakeGate) probeAdmission() {
	for _, kind := range []ClientKind{ClientRemote, ClientTUI} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		c, err := g.h.Connect(ctx, ConnectOptions{Kind: kind})
		cancel()
		if err == nil {
			g.admitted.Add(1)
			c.Close()
		} else if !errors.Is(err, ErrShuttingDown) {
			g.rec.add("PROBE-ERR:" + err.Error())
		}
	}
}

func (g *fakeGate) RevokeCredentials(context.Context) error {
	g.rec.mark("gate.revoke")
	g.probeAdmission()
	if g.block != nil {
		<-g.block
	}
	return g.revokeErr
}

func (g *fakeGate) CloseListener(context.Context) error {
	g.rec.mark("gate.close")
	g.probeAdmission()
	return g.closeErr
}

type shutdownRig struct {
	h     *Hub
	rec   *recorder
	ag    *fakeShutdownAgent
	gate  *fakeGate
	tui   *Client
	rem   *Client
	stop  chan struct{}
	logs  *bytes.Buffer
	logMu *sync.Mutex
}

func newShutdownRig(t *testing.T, cfg Config) *shutdownRig {
	t.Helper()
	r := &shutdownRig{stop: make(chan struct{}), logs: &bytes.Buffer{}, logMu: &sync.Mutex{}}
	r.h = startHub(t, cfg)
	r.rec = &recorder{h: r.h}
	r.ag = &fakeShutdownAgent{rec: r.rec}
	r.gate = &fakeGate{rec: r.rec, h: r.h}
	r.tui = connectRole(t, r.h, ClientTUI, 0)
	r.rem = connectRole(t, r.h, ClientRemote, RoleApprover)
	r.rec.consume(r.tui, r.stop)
	t.Cleanup(func() { close(r.stop) })
	return r
}

func (r *shutdownRig) newShutdown(t *testing.T, mutate func(*ShutdownConfig)) *Shutdown {
	t.Helper()
	cfg := ShutdownConfig{
		Hub: r.h, Agent: r.ag, Gate: r.gate, Reason: "operator quit",
		Timeouts: ShutdownTimeouts{Permissions: 2 * time.Second, Gate: time.Second, Announce: time.Second,
			Abort: time.Second, Idle: time.Second, Stop: 2 * time.Second, Drain: 500 * time.Millisecond},
		Logger: slog.New(slog.NewTextHandler(&lockedWriter{mu: r.logMu, w: r.logs}, nil)),
	}
	if mutate != nil {
		mutate(&cfg)
	}
	s, err := NewShutdown(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func (r *shutdownRig) logText() string {
	r.logMu.Lock()
	defer r.logMu.Unlock()
	return r.logs.String()
}

// waitUntil polls cond (each poll is cheap) until it holds or testWait passes.
func waitUntil(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(testWait)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func waitShutdown(t *testing.T, s *Shutdown) {
	t.Helper()
	select {
	case <-s.Done():
	case <-time.After(3 * testWait):
		t.Fatal("shutdown did not finish: deadlock")
	}
}

// parkHandler plays the SDK permission handler: it asks the hub, parks until
// the hub replies (or the hub is gone), then returns. released closes when it
// has returned with the hub's decision.
func parkHandler(t *testing.T, h *Hub) (released chan struct{}, decision *atomic.Int32) {
	t.Helper()
	released = make(chan struct{})
	decision = &atomic.Int32{}
	a := newTestAsk(agent.PermissionKindShell, "rm -rf /")
	a.post(t, h)
	go func() {
		select {
		case d := <-a.reply:
			a.outcome.Deliver(d)
			decision.Store(int32(d.Kind))
		case <-h.Done():
			a.outcome.Deliver(agent.PermissionDecision{Kind: agent.DecisionUnavailable})
			decision.Store(int32(agent.DecisionUnavailable))
		}
		close(a.resolved) // the handler returns: acknowledgement
		close(released)
	}()
	return released, decision
}

func indexOf(t *testing.T, ev []string, name string) int {
	t.Helper()
	i := slices.Index(ev, name)
	if i < 0 {
		t.Fatalf("event %q missing from %v", name, ev)
	}
	return i
}

func TestShutdownOrderProvesEachStep(t *testing.T) {
	r := newShutdownRig(t, Config{})
	released, decision := parkHandler(t, r.h)
	waitUntil(t, "the ask to be registered", func() bool { return stats(t, r.h).PendingPermissions > 0 })

	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitShutdown(t, s)
	if err := s.Err(); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	select {
	case <-released:
	default:
		t.Fatal("the permission handler was never released")
	}
	if agent.DecisionKind(decision.Load()) != agent.DecisionUnavailable {
		t.Fatalf("handler decision = %v, want UserNotAvailable", decision.Load())
	}

	ev := r.rec.events()
	order := []string{
		"frame:permission.resolved:shutdown", // step 0: pending failed through the hub, with a shutdown reason
		"gate.revoke",                        // step 0: credentials revoked ...
		"gate.close",                         // ... and the listener closed
		"frame:server.shutdown",              // step 1: broadcast before any teardown
		"agent.stop",                         // steps 2-4: the agent is stopped only after the broadcast
	}
	last := -1
	for _, name := range order {
		i := indexOf(t, ev, name)
		if i <= last {
			t.Fatalf("%q out of order in %v", name, ev)
		}
		last = i
	}
	for _, e := range ev {
		if len(e) > 14 && e[:14] == "HUB-DONE-EARLY" {
			t.Fatalf("the hub stopped before step 6: %v", ev)
		}
	}
	// Step 5 before step 6: by the time the hub is done, every client's done
	// has been closed.
	select {
	case <-r.h.Done():
	default:
		t.Fatal("step 6 did not stop the hub")
	}
	for name, c := range map[string]*Client{"tui": r.tui, "remote": r.rem} {
		select {
		case <-c.Done():
		default:
			t.Fatalf("%s client done is open after shutdown", name)
		}
	}
	if n := r.ag.stops.Load(); n != 1 {
		t.Fatalf("agent stopped %d times", n)
	}
}

func TestShutdownDoesNotDeadlockWithABlockedPermissionHandler(t *testing.T) {
	r := newShutdownRig(t, Config{})
	released, _ := parkHandler(t, r.h)
	waitUntil(t, "the ask to be registered", func() bool { return stats(t, r.h).PendingPermissions > 0 })
	// A session that cannot stop while the handler is parked: Stop returns
	// only after the handler did. Nothing but step 0 can release it.
	r.ag.waitFor = released

	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitShutdown(t, s)
	if err := s.Err(); err != nil {
		t.Fatalf("shutdown error (a deadlock ends in a stop timeout): %v", err)
	}
}

func TestNoClientIsAdmittedAfterStepZero(t *testing.T) {
	r := newShutdownRig(t, Config{})
	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitShutdown(t, s)
	if n := r.gate.admitted.Load(); n != 0 {
		t.Fatalf("%d clients were admitted while the gate calls ran", n)
	}
	for _, e := range r.rec.events() {
		if len(e) > 9 && e[:9] == "PROBE-ERR" {
			t.Fatalf("probe failed for the wrong reason: %s", e)
		}
	}
	for _, kind := range []ClientKind{ClientRemote, ClientTUI} {
		if c, err := r.h.Connect(context.Background(), ConnectOptions{Kind: kind}); err == nil {
			c.Close()
			t.Fatalf("kind %v connected after shutdown", kind)
		}
	}
}

func TestBeginShutdownClosesAdmissionAsksAndPrivilegedActions(t *testing.T) {
	h := startHub(t, Config{Agent: newFakeAgent()})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)

	if err := h.BeginShutdown(t.Context()); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	if !stats(t, h).ShuttingDown {
		t.Fatal("hub does not report shutting down")
	}
	// Admission is closed for both kinds.
	for _, kind := range []ClientKind{ClientRemote, ClientTUI} {
		_, err := h.Connect(t.Context(), ConnectOptions{Kind: kind})
		if !errors.Is(err, ErrShuttingDown) {
			t.Fatalf("Connect(%v) = %v, want ErrShuttingDown", kind, err)
		}
	}
	// A new ask is answered UserNotAvailable immediately and never registered,
	// so no handler can park after step 0.
	a := newTestAsk(agent.PermissionKindRead, "late")
	a.post(t, h)
	if d := a.takeReply(t); d.Kind != agent.DecisionUnavailable {
		t.Fatalf("late ask decision = %v", d.Kind)
	}
	if n := stats(t, h).PendingPermissions; n != 0 {
		t.Fatalf("pending = %d", n)
	}
	// Privileged actions are refused without closing the connection.
	if res := do(t, appr, prompt("x")); res.Accepted || res.Reason != ReasonCredentialRevoked {
		t.Fatalf("remote action after revoke = %+v, want credential_revoked", res)
	}
	if res := do(t, tui, prompt("x")); res.Accepted || res.Reason != ReasonShuttingDown {
		t.Fatalf("tui action during shutdown = %+v, want shutting_down", res)
	}
	for _, c := range []*Client{tui, appr} {
		select {
		case <-c.Done():
			t.Fatal("a refusal closed the connection; the broadcast and drain must still finish")
		default:
		}
	}
}

func TestHubKeepsDrainingIngressWhileTheAgentStops(t *testing.T) {
	r := newShutdownRig(t, Config{IngressBuffer: 2})
	r.ag.stopBlock = make(chan struct{})
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Stop = 10 * time.Second })
	s.Start(context.Background())

	waitUntil(t, "the executor to reach Agent.Stop", func() bool { return slices.Contains(r.rec.events(), "agent.stop") })
	// SDK teardown produces events; the hub must still fold and broadcast them
	// (far more than the ingress buffer holds: a hub that stopped draining
	// would wedge the sender).
	for i := 0; i < 20; i++ {
		send(t, r.h, msgEvent("m", "teardown event"))
	}
	waitUntil(t, "all 20 teardown events to be broadcast (the hub stopped draining ingress)", func() bool {
		n := 0
		for _, e := range r.rec.events() {
			if e == "frame:agent.message" {
				n++
			}
		}
		return n == 20
	})
	close(r.ag.stopBlock)
	waitShutdown(t, s)
}

func TestShutdownAggregatesAndLogsStepErrorsAndKeepsGoing(t *testing.T) {
	r := newShutdownRig(t, Config{})
	errRevoke, errClose, errStop := errors.New("revoke boom"), errors.New("close boom"), errors.New("stop boom")
	r.gate.revokeErr, r.gate.closeErr, r.ag.stopErr = errRevoke, errClose, errStop
	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitShutdown(t, s)

	for _, want := range []error{errRevoke, errClose, errStop} {
		if !errors.Is(s.Err(), want) {
			t.Errorf("Err() = %v lacks %v", s.Err(), want)
		}
	}
	for _, want := range []string{"revoke boom", "close boom", "stop boom"} {
		if !bytes.Contains([]byte(r.logText()), []byte(want)) {
			t.Errorf("log lacks %q: errors must be logged, never swallowed\n%s", want, r.logText())
		}
	}
	ev := r.rec.events()
	for _, name := range []string{"gate.close", "frame:server.shutdown", "agent.stop"} {
		indexOf(t, ev, name) // an earlier failure must not skip later steps
	}
	select {
	case <-r.h.Done():
	default:
		t.Fatal("the hub was not stopped after step errors")
	}
}

func TestHungGateIsBoundedAndShutdownContinues(t *testing.T) {
	r := newShutdownRig(t, Config{})
	r.gate.block = make(chan struct{})
	t.Cleanup(func() { close(r.gate.block) })
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Gate = 50 * time.Millisecond })
	s.Start(context.Background())
	waitShutdown(t, s)
	if !errors.Is(s.Err(), context.DeadlineExceeded) {
		t.Fatalf("Err() = %v, want the gate deadline", s.Err())
	}
	indexOf(t, r.rec.events(), "agent.stop")
}

func TestDrainWaitsForTheWritePumpThenClosesDone(t *testing.T) {
	r := newShutdownRig(t, Config{})
	slow := r.rem // never read until released: its queue holds the shutdown frames
	r.ag.stopBlock = make(chan struct{})
	s := r.newShutdown(t, func(c *ShutdownConfig) {
		c.Timeouts.Drain = 10 * time.Second
		c.Timeouts.Stop = 10 * time.Second
	})
	s.Start(context.Background())
	waitUntil(t, "the executor to reach Agent.Stop", func() bool { return slices.Contains(r.rec.events(), "agent.stop") })
	close(r.ag.stopBlock) // steps 2-4 finish; step 5 begins

	// While the pump still has frames to write, done must stay open: aborting
	// the socket now would lose the server.shutdown the operator needs.
	select {
	case <-slow.Done():
		t.Fatal("done closed while the write pump had undelivered frames")
	case <-time.After(150 * time.Millisecond):
	}
	var got []Kind
	deadline := time.After(testWait)
loop:
	for {
		select {
		case f := <-slow.Frames():
			got = append(got, f.Kind())
		case <-slow.Done():
			break loop
		case <-deadline:
			t.Fatal("done never closed after the queue drained")
		}
	}
	for { // pumps drain what remains after done
		select {
		case f := <-slow.Frames():
			got = append(got, f.Kind())
			continue
		default:
		}
		break
	}
	if !slices.Contains(got, KindServerShutdown) {
		t.Fatalf("the client never received server.shutdown: %v", got)
	}
	waitShutdown(t, s)
}

func TestDrainIsBoundedForAStalledClient(t *testing.T) {
	r := newShutdownRig(t, Config{})
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Drain = 100 * time.Millisecond })
	start := time.Now()
	s.Start(context.Background()) // r.rem is never read: its queue never empties
	waitShutdown(t, s)
	if el := time.Since(start); el > 5*time.Second {
		t.Fatalf("the drain was not bounded: %v", el)
	}
	select {
	case <-r.rem.Done():
	default:
		t.Fatal("a stalled client was not aborted after the bounded drain")
	}
}

func TestShutdownStartIsIdempotentAndCancelledParentStillStopsTheHub(t *testing.T) {
	r := newShutdownRig(t, Config{})
	s := r.newShutdown(t, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // the parent is already gone
	s.Start(ctx)
	s.Start(context.Background())
	waitShutdown(t, s)
	select {
	case <-r.h.Done():
	default:
		t.Fatal("step 6 must run even when the parent context is cancelled")
	}
	if n := r.ag.stops.Load(); n > 1 {
		t.Fatalf("agent stopped %d times", n)
	}
	if s.Err() == nil {
		t.Fatal("a cancelled shutdown must report its context error")
	}
}

func TestNewShutdownValidatesItsConfig(t *testing.T) {
	h := startHub(t, Config{})
	rec := &recorder{h: h}
	ag, gate := &fakeShutdownAgent{rec: rec}, &fakeGate{rec: rec, h: h}
	for name, cfg := range map[string]ShutdownConfig{
		"no hub":   {Agent: ag, Gate: gate},
		"no agent": {Hub: h, Gate: gate},
		"no gate":  {Hub: h, Agent: ag},
		"negative": {Hub: h, Agent: ag, Gate: gate, Timeouts: ShutdownTimeouts{Drain: -1}},
	} {
		if _, err := NewShutdown(cfg); err == nil {
			t.Errorf("%s: NewShutdown accepted an invalid config", name)
		}
	}
}
