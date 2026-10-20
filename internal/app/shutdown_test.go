package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// ---- goroutine leak detection ----

// ownGoroutines returns the IDs of live goroutines that run this module's
// production code (internal/**), excluding test goroutines and the reference
// client's. A goroutine a test started has a *_test.go frame; production
// goroutines (hub, transport, feed, shutdown executor) do not.
func ownGoroutines() map[string]string {
	var buf bytes.Buffer
	_ = pprof.Lookup("goroutine").WriteTo(&buf, 2)
	own := map[string]string{}
	for _, block := range strings.Split(buf.String(), "\n\n") {
		if !strings.Contains(block, "intercom-go/internal/") ||
			strings.Contains(block, "_test.go") ||
			strings.Contains(block, "testing.tRunner") ||
			strings.Contains(block, "transport/wsclient") {
			continue
		}
		head, _, _ := strings.Cut(block, "\n")
		fields := strings.Fields(head) // "goroutine 42 [chan receive]:"
		if len(fields) >= 2 {
			own[fields[1]] = block
		}
	}
	return own
}

// assertNoLeaks waits (bounded) until no production goroutine started since
// baseline is left.
func assertNoLeaks(t *testing.T, baseline map[string]string) {
	t.Helper()
	deadline := time.Now().Add(waitFor)
	var leaked []string
	for {
		leaked = leaked[:0]
		for id, block := range ownGoroutines() {
			if _, was := baseline[id]; !was {
				leaked = append(leaked, block)
			}
		}
		if len(leaked) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d goroutine(s) leaked after shutdown:\n%s", len(leaked), strings.Join(leaked, "\n\n"))
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// ---- shutdown probes ----

// probe is what the fake agent observes about the rest of the process at the
// moment the shutdown executor calls it.
type probe struct {
	listenerRefused bool // new TCP connections are refused (step 0 closed the listener)
	shuttingDown    bool // the hub has begun shutdown (step 0 ran through the hub)
	hubRunning      bool // the hub is still draining ingress (it stops only at step 6)
}

func takeProbe(a *App) probe {
	var p probe
	c, err := net.DialTimeout("tcp", a.Addr().String(), time.Second)
	if err == nil {
		_ = c.Close()
	}
	p.listenerRefused = err != nil
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	st, err := a.Hub().Stats(ctx)
	p.shuttingDown = err == nil && st.ShuttingDown
	select {
	case <-a.Hub().Done():
	default:
		p.hubRunning = true
	}
	return p
}

type probes struct {
	mu sync.Mutex
	m  map[string]probe
}

func (p *probes) set(label string, v probe) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.m == nil {
		p.m = map[string]probe{}
	}
	p.m[label] = v
}

func (p *probes) get(label string) (probe, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.m[label]
	return v, ok
}

// permAsk is a permission request whose handler records when it is told
// "user not available" (step 0) before it acknowledges.
type permAsk struct {
	ask      agent.PermissionAsk
	reply    chan agent.PermissionDecision
	outcome  *agent.PermissionOutcome
	resolved chan struct{}
}

func newPermAsk() *permAsk {
	reply := make(chan agent.PermissionDecision, 1)
	resolved := make(chan struct{})
	outcome := &agent.PermissionOutcome{}
	return &permAsk{reply: reply, outcome: outcome, resolved: resolved, ask: agent.PermissionAsk{
		Request:  agent.PermissionRequest{ToolCallID: "call-1", Kind: agent.PermissionKindRead, RawKind: "read", Summary: "read a file"},
		Reply:    reply,
		Outcome:  outcome,
		Resolved: resolved,
	}}
}

// serve plays the adapter's handler: wait for the hub's decision, record it,
// then deliver it (the hub's step 0 waits for this acknowledgement).
func (p *permAsk) serve(rec *recorder, got chan<- agent.PermissionDecision) {
	d := <-p.reply
	rec.add("permission.unavailable")
	p.outcome.Deliver(d)
	close(p.resolved)
	got <- d
}

// frameLog collects the envelope kinds a remote client receives.
type frameLog struct {
	mu    sync.Mutex
	kinds []string
}

func (l *frameLog) add(k string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.kinds = append(l.kinds, k)
}

func (l *frameLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.kinds...)
}

func indexOf(s []string, v string) int {
	for i, x := range s {
		if x == v {
			return i
		}
	}
	return -1
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// watchHubStop records "hub.stopped" once the hub's Run has returned.
func watchHubStop(rec *recorder, hubDone <-chan struct{}) {
	go func() { <-hubDone; rec.add("hub.stopped") }()
}

// ---- the tests ----

// TestSIGINTRunsTheSection31Order drives the real hub, real WebSocket
// transport and real shutdown executor with a fake agent, delivers an
// interrupt, and checks the design §3.1 order with a pending permission and a
// connected remote client:
//
//	step 0  pending permission -> UserNotAvailable (through the hub), then
//	        credentials revoked and listener closed (observed at abort time)
//	step 1  server.shutdown broadcast (after permission.resolved, to the client)
//	step 3  Agent.Abort, then step 4 Agent.Stop (hub still running at both)
//	step 5  client sockets closed
//	step 6  hub stopped LAST
func TestSIGINTRunsTheSection31Order(t *testing.T) {
	baseline := ownGoroutines()
	f := newFactory()
	o, _ := baseOptions(t, f)
	sig := make(chan os.Signal, 2)
	o.Signals = sig

	var cur atomic.Pointer[App]
	var pr probes
	f.abortHook = func() { pr.set("abort", takeProbe(cur.Load())) }
	f.stopHook = func(context.Context) { pr.set("stop", takeProbe(cur.Load())) }

	a := startApp(t, o)
	cur.Store(a)
	rec := f.rec
	watchHubStop(rec, f.fake().deps.HubDone)

	// A connected remote surface (so the permission request stays pending).
	client := dialApp(t, a, hub.RoleApprover)
	var frames frameLog
	sawRequested := make(chan struct{})
	var once sync.Once
	clientGone := make(chan struct{})
	go func() {
		defer close(clientGone)
		ctx, cancel := context.WithTimeout(context.Background(), waitFor)
		defer cancel()
		for {
			m, err := client.Next(ctx)
			if err != nil {
				return
			}
			if m.Envelope != nil {
				frames.add(string(m.Envelope.Kind))
				if m.Envelope.Kind == hub.KindPermissionRequested {
					once.Do(func() { close(sawRequested) })
				}
			}
		}
	}()
	ask := newPermAsk()
	decided := make(chan agent.PermissionDecision, 1)
	go ask.serve(rec, decided)
	f.fake().deps.Asks <- ask.ask
	select {
	case <-sawRequested:
	case <-time.After(waitFor):
		t.Fatal("the remote client never saw the permission request")
	}

	sig <- os.Interrupt

	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("the process did not finish shutting down after SIGINT")
	}
	if err := a.Err(); err != nil {
		t.Fatalf("shutdown reported errors: %v (a signal that cancels the hub's context before step 0 would show up here as ErrStopped)", err)
	}
	if d := <-decided; d.Kind != agent.DecisionUnavailable {
		t.Errorf("pending permission resolved to %v; want UserNotAvailable", d.Kind)
	}
	rec.waitEvent(t, "hub.stopped")
	want := []string{"agent.start", "permission.unavailable", "agent.abort", "agent.stop", "hub.stopped"}
	if got := rec.snapshot(); !equalStrings(got, want) {
		t.Fatalf("shutdown order = %v; want %v", got, want)
	}
	for _, label := range []string{"abort", "stop"} {
		p, ok := pr.get(label)
		if !ok {
			t.Fatalf("no probe taken at agent %s", label)
		}
		if !p.listenerRefused || !p.shuttingDown || !p.hubRunning {
			t.Errorf("at agent %s: %+v; want listener closed, hub shutting down and still running (step 0 precedes the agent steps; hub stops last)", label, p)
		}
	}
	select {
	case <-clientGone:
	case <-time.After(waitFor):
		t.Fatal("the remote client's socket was never closed (step 5)")
	}
	kinds := frames.snapshot()
	if r, s := indexOf(kinds, string(hub.KindPermissionResolved)), indexOf(kinds, string(hub.KindServerShutdown)); r < 0 || s < 0 || r > s {
		t.Errorf("client saw %v; want permission.resolved before server.shutdown", kinds)
	}
	if c, err := net.DialTimeout("tcp", a.Addr().String(), time.Second); err == nil {
		_ = c.Close()
		t.Error("the listener is still open after shutdown")
	}
	select {
	case <-a.Hub().Done():
	default:
		t.Error("the hub is still running after shutdown")
	}
	assertNoLeaks(t, baseline)
}

// TestTUIQuitRunsTheSameOrderedShutdown: Ctrl+C in the local TUI ends the
// program, which triggers the identical shutdown.
func TestTUIQuitRunsTheSameOrderedShutdown(t *testing.T) {
	baseline := ownGoroutines()
	f := newFactory()
	o, _ := baseOptions(t, f)
	in, inW := io.Pipe()
	out := &syncBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	o.TUI = TUIOptions{Input: in, Output: out, Width: 100, Height: 30}

	var cur atomic.Pointer[App]
	var pr probes
	f.abortHook = func() { pr.set("abort", takeProbe(cur.Load())) }
	f.stopHook = func(context.Context) { pr.set("stop", takeProbe(cur.Load())) }

	a := startApp(t, o)
	cur.Store(a)
	watchHubStop(f.rec, f.fake().deps.HubDone)
	waitOutput(t, out, "shown once") // the TUI is up and attached

	if _, err := inW.Write([]byte("\x03")); err != nil { // Ctrl+C
		t.Fatal(err)
	}
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("the process did not shut down after the TUI quit")
	}
	if err := a.Err(); err != nil {
		t.Fatalf("shutdown reported errors: %v", err)
	}
	f.rec.waitEvent(t, "hub.stopped")
	want := []string{"agent.start", "agent.abort", "agent.stop", "hub.stopped"}
	if got := f.rec.snapshot(); !equalStrings(got, want) {
		t.Fatalf("shutdown order = %v; want %v", got, want)
	}
	for _, label := range []string{"abort", "stop"} {
		if p, ok := pr.get(label); !ok || !p.listenerRefused || !p.shuttingDown || !p.hubRunning {
			t.Errorf("at agent %s: %+v (taken=%v); want listener closed, hub shutting down and still running", label, p, ok)
		}
	}
	select {
	case <-a.ui.done:
	default:
		t.Error("the TUI program is still running after shutdown")
	}
	assertNoLeaks(t, baseline)
}

// TestShutdownIsIdempotent: any mix of triggers runs the sequence once.
func TestShutdownIsIdempotent(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	sig := make(chan os.Signal, 4)
	o.Signals = sig
	a := startApp(t, o)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); a.Shutdown("concurrent") }()
	}
	sig <- os.Interrupt
	wg.Wait()
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("shutdown did not finish")
	}
	a.Shutdown("after done") // must not panic or restart anything
	count := func(e string) int {
		n := 0
		for _, got := range f.rec.snapshot() {
			if got == e {
				n++
			}
		}
		return n
	}
	if _, stops := f.fake().counts(); stops != 1 || count("agent.abort") != 1 {
		t.Fatalf("agent stopped %d times, aborted %d times; want once each", stops, count("agent.abort"))
	}
}

// TestSecondSignalHurriesAStuckStop: a stuck Agent.Stop is waited for, never
// abandoned, until an operator's second signal cancels the shutdown context
// (the adapter then force-stops); the exit is then non-zero.
func TestSecondSignalHurriesAStuckStop(t *testing.T) {
	f := newFactory()
	f.stopBlock = true
	o, _ := baseOptions(t, f)
	sig := make(chan os.Signal, 4)
	o.Signals = sig
	o.Timeouts = hub.ShutdownTimeouts{Stop: time.Minute, Force: time.Minute}
	a := startApp(t, o)

	sig <- os.Interrupt
	f.rec.waitEvent(t, "agent.stop") // now stuck inside Stop
	select {
	case <-a.Done():
		t.Fatal("shutdown finished while Agent.Stop was still stuck")
	case <-time.After(50 * time.Millisecond):
	}
	sig <- os.Interrupt // hurry
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("a second signal did not hurry the stuck stop")
	}
	if err := a.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("Err() = %v; want a non-nil error wrapping context.Canceled (interrupted shutdown exits non-zero)", err)
	}
}

// TestSignalDuringStartupAbortsStartAndReapsTheAgent: an interrupt while the
// agent is still starting cancels the start, stops the agent (so no CLI child
// is stranded) and unwinds everything.
func TestSignalDuringStartupAbortsStartAndReapsTheAgent(t *testing.T) {
	f := newFactory()
	f.startBlock = true
	o, _ := baseOptions(t, f)
	sig := make(chan os.Signal, 2)
	o.Signals = sig
	o.Port = freePort(t)

	type result struct {
		a   *App
		err error
	}
	res := make(chan result, 1)
	go func() {
		a, err := Start(context.Background(), o)
		res <- result{a, err}
	}()
	f.rec.waitEvent(t, "agent.start")
	sig <- os.Interrupt

	select {
	case r := <-res:
		if r.err == nil {
			stopApp(t, r.a)
			t.Fatal("Start succeeded despite the interrupt")
		}
		if !errors.Is(r.err, context.Canceled) {
			t.Errorf("Start error = %v; want it to wrap context.Canceled", r.err)
		}
	case <-time.After(waitFor):
		t.Fatal("Start did not return after the interrupt")
	}
	fa := f.fake()
	if _, stops := fa.counts(); stops != 1 {
		t.Errorf("Agent.Stop called %d times; want 1 (reap the CLI child)", stops)
	}
	select {
	case <-fa.deps.HubDone:
	case <-time.After(waitFor):
		t.Error("hub still running after an aborted start")
	}
	if c, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", itoa(o.Port)), time.Second); err == nil {
		_ = c.Close()
		t.Error("listener still open after an aborted start")
	}
}
