package hub

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// streamChecker is an INDEPENDENT oracle for the no-gap / no-duplicate
// property (it deliberately does not reuse Cursor, so a Cursor bug cannot
// mask a hub bug). Frames of one logical consumer, in the order consumed
// across any number of connections, must satisfy per (gen, seq):
//   - a hydrate may jump forward (or stay) but never go back;
//   - every other frame is exactly last+1: a larger value is a GAP, a smaller
//     or equal one a DUPLICATE / out-of-order delivery.
type streamChecker struct {
	gen  string
	last uint64
	have bool
}

func (c *streamChecker) check(f Frame) (violation string) {
	if f.Gen() != c.gen {
		return fmt.Sprintf("gen %q != hub gen %q", f.Gen(), c.gen)
	}
	if f.Kind() == KindSessionHydrate {
		if c.have && f.Seq() < c.last {
			violation = fmt.Sprintf("hydrate seq %d went backwards from %d", f.Seq(), c.last)
		}
		c.last, c.have = f.Seq(), true
		return violation
	}
	switch {
	case !c.have:
		violation = fmt.Sprintf("seq %d delivered before any hydrate", f.Seq())
	case f.Seq() > c.last+1:
		violation = fmt.Sprintf("GAP: seq %d after %d", f.Seq(), c.last)
	case f.Seq() <= c.last:
		violation = fmt.Sprintf("DUPLICATE: seq %d after %d", f.Seq(), c.last)
	}
	c.last, c.have = f.Seq(), true // resync so one fault is reported once
	return violation
}

type propertyOpts struct {
	cfg       Config
	events    int
	workers   int
	tuiWorker bool
}

// runProperty drives one concurrent producer while clients connect, drop and
// reconnect, and returns every violation of the no-gap/no-duplicate property
// (plus liveness failures).
func runProperty(t *testing.T, o propertyOpts) []string {
	t.Helper()
	o.cfg.SessionID = "sess"
	o.cfg.MaxRemote = MaxRemoteHardLimit // workers may leave clients registered until dropped
	h, err := New(o.cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(ctx) }()
	defer func() { cancel(); <-runDone }()

	var (
		mu         sync.Mutex
		violations []string
		stop       atomic.Bool
	)
	report := func(who, msg string) {
		mu.Lock()
		violations = append(violations, who+": "+msg)
		mu.Unlock()
		stop.Store(true)
	}

	deadline := time.Now().Add(8 * time.Second)
	var wg sync.WaitGroup
	worker := func(id int, kind ClientKind) {
		defer wg.Done()
		who := fmt.Sprintf("worker %d", id)
		rng := rand.New(rand.NewSource(int64(id) + 1))
		chk := streamChecker{gen: h.Gen()}
		for chk.last < uint64(o.events) || !chk.have {
			if stop.Load() {
				return
			}
			if time.Now().After(deadline) {
				report(who, fmt.Sprintf("liveness: stuck at seq %d of %d", chk.last, o.events))
				return
			}
			opts := ConnectOptions{Kind: kind, Buffer: 4 + rng.Intn(60)}
			if chk.have && rng.Intn(4) != 0 { // mostly resume, sometimes connect fresh
				opts.Resume = &Position{Gen: h.Gen(), Seq: chk.last}
			}
			cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Second)
			c, err := h.Connect(cctx, opts)
			ccancel()
			if err != nil {
				if stop.Load() {
					return
				}
				report(who, "connect: "+err.Error())
				return
			}
			quota := 1 + rng.Intn(200) // frames to read before voluntarily dropping
			for read := 0; read < quota && chk.last < uint64(o.events); {
				select {
				case f := <-c.Frames():
					if v := chk.check(f); v != "" {
						report(who, v)
						c.Close()
						return
					}
					read++
					if rng.Intn(16) == 0 {
						runtime.Gosched() // be a little slow: provoke overflow and drops
					}
				case <-c.Done():
					read = quota // dropped by the hub: reconnect and resume
				case <-time.After(20 * time.Millisecond):
					if stop.Load() || time.Now().After(deadline) {
						c.Close()
						return
					}
				}
			}
			if kind == ClientTUI {
				c.Close() // the TUI exiting and re-attaching is legal
			} else if rng.Intn(2) == 0 {
				c.Close()
			}
			if chk.last >= uint64(o.events) {
				c.Close()
			}
		}
	}
	id := 0
	for ; id < o.workers; id++ {
		wg.Add(1)
		go worker(id, ClientRemote)
	}
	if o.tuiWorker {
		wg.Add(1)
		go worker(id, ClientTUI)
	}

	go func() {
		for i := 0; i < o.events && !stop.Load(); i++ {
			select {
			case h.Ingress() <- agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: "x"}:
			case <-time.After(5 * time.Second):
				report("producer", "ingress send wedged")
				return
			}
			if i%97 == 0 {
				runtime.Gosched()
			}
		}
	}()

	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	return append([]string(nil), violations...)
}

func TestPropertyNoGapNoDuplicateUnderConcurrentConnectDropReconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("slow property variants skipped in -short mode")
	}
	base := runtime.NumGoroutine()
	for seed, ring := range []int{16, 64, 1024} { // wrapped, borderline, fully covering
		v := runProperty(t, propertyOpts{
			cfg:       Config{ReplayRing: ring, IngressBuffer: 8, HydrationsPerIteration: 2},
			events:    2500,
			workers:   10,
			tuiWorker: true,
		})
		if len(v) > 0 {
			t.Fatalf("variant %d (ring %d): %d violation(s); first: %s", seed, ring, len(v), v[0])
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > base && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > base {
		t.Fatalf("goroutine leak: %d > baseline %d", n, base)
	}
}

// TestPropertyDetectorTripsOnHydrateThenRegisterMutation is the mutation
// check (066.005-T): a test-only switch makes the hub snapshot and then
// register in separate steps — the exact defect design §5.3 forbids — and the
// property harness MUST report a gap. If it does not, the property test is not
// actually guarding the one-step invariant.
func TestPropertyDetectorTripsOnHydrateThenRegisterMutation(t *testing.T) {
	var last []string
	for attempt := 0; attempt < 3; attempt++ {
		last = runProperty(t, propertyOpts{
			cfg:     Config{ReplayRing: 16, IngressBuffer: 8, mutHydrateThenRegister: true},
			events:  2500,
			workers: 10,
		})
		for _, v := range last {
			if containsAny(v, "GAP", "DUPLICATE", "liveness") {
				t.Logf("mutation detected as expected: %s", v)
				return
			}
		}
	}
	t.Fatalf("mutation NOT detected: violations=%v", last)
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
