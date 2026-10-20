package copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// 066.008-T: adapter-to-ingress integration. The REAL adapter callback path
// (onEvent -> normalize -> blocking ingress send) is driven through the
// existing fakeSession.emit seam into a REAL hub. copilot.Config has no
// degraded-channel field, so the degraded path is exercised from the adapter
// side by signalling the hub's capacity-1 Degraded channel directly.

const intWait = 5 * time.Second

func emitMessage(s *fakeSession, i int) {
	s.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantMessageData{MessageID: fmt.Sprintf("m%03d", i), Content: fmt.Sprintf("%d", i)}})
}

type hubRig struct {
	hub     *hub.Hub
	h       *harness
	runDone chan error
	cancel  context.CancelFunc
}

// newHubRig wires a real hub's ingress into the adapter harness. The hub is
// NOT running yet so the test controls when ingress drains.
func newHubRig(t *testing.T, ingressBuf int) *hubRig {
	t.Helper()
	hb, err := hub.New(hub.Config{SessionID: "sess-1", IngressBuffer: ingressBuf})
	if err != nil {
		t.Fatal(err)
	}
	r := &hubRig{hub: hb, runDone: make(chan error, 1)}
	r.h = newHarness(t, func(c *Config, _ *fakeClient) { c.Ingress = hb.Ingress() })
	if _, err := r.h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *hubRig) run() {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go func() { r.runDone <- r.hub.Run(ctx) }()
}

func (r *hubRig) stop(t *testing.T) {
	t.Helper()
	if r.cancel == nil {
		return
	}
	r.cancel()
	select {
	case <-r.runDone:
	case <-time.After(intWait):
		t.Error("hub Run did not return")
	}
}

type tally struct {
	hydrateSeq  uint64
	snapEntries int
	frames      []hub.Frame // non-hydrate frames in arrival order
	hydrates    []hub.Frame
}

func entryCount(t *testing.T, f hub.Frame) int {
	t.Helper()
	var env hub.Envelope
	if err := json.Unmarshal(f.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	var snap struct {
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(env.Payload, &snap); err != nil {
		t.Fatal(err)
	}
	return len(snap.Entries)
}

// consume reads until a frame with seq >= wantLast arrives.
func consume(t *testing.T, c *hub.Client, wantLast uint64) tally {
	t.Helper()
	var tl tally
	var last uint64
	deadline := time.After(intWait)
	for last < wantLast {
		select {
		case f := <-c.Frames():
			if f.Kind() == hub.KindSessionHydrate {
				tl.hydrates = append(tl.hydrates, f)
				if len(tl.hydrates) == 1 {
					tl.hydrateSeq = f.Seq()
					tl.snapEntries = entryCount(t, f)
				}
			} else {
				tl.frames = append(tl.frames, f)
			}
			last = f.Seq()
		case <-deadline:
			t.Fatalf("stalled at seq %d, want %d", last, wantLast)
		}
	}
	return tl
}

func TestIntegrationSaturatedIngressBlockingSendNeverDropsSeqContinuous(t *testing.T) {
	const ingressBuf, prefill, burst = 2, 2, 150
	r := newHubRig(t, ingressBuf)
	defer r.stop(t)

	// Saturate ingress BEFORE the hub drains: fill the buffer directly, then
	// drive the adapter callback, which must block (never drop, never spin).
	for i := 0; i < prefill; i++ {
		r.hub.Ingress() <- agent.Event{Kind: agent.KindIdle}
	}
	emitted := make(chan struct{})
	go func() {
		for i := 0; i < burst; i++ {
			emitMessage(r.h.client.session, i)
		}
		close(emitted)
	}()
	select {
	case <-emitted:
		t.Fatal("adapter callback returned with saturated ingress: events dropped")
	case <-time.After(100 * time.Millisecond):
	}

	// A client queued before the hub starts: hydrate + stream must be gap-free.
	var c *hub.Client
	connected := make(chan error, 1)
	go func() {
		var err error
		c, err = r.hub.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientRemote, Buffer: burst + 16})
		connected <- err
	}()
	r.run()
	if err := <-connected; err != nil {
		t.Fatal(err)
	}
	select {
	case <-emitted:
	case <-time.After(intWait):
		t.Fatal("adapter callback never unblocked after the hub started draining")
	}

	total := uint64(prefill + burst)
	tl := consume(t, c, total)
	if tl.hydrateSeq > total || len(tl.hydrates) != 1 {
		t.Fatalf("hydrates = %d at seq %d", len(tl.hydrates), tl.hydrateSeq)
	}
	next := tl.hydrateSeq + 1
	var lastText = -1
	for _, f := range tl.frames {
		if f.Seq() != next {
			t.Fatalf("seq continuity broken: got %d want %d", f.Seq(), next)
		}
		next++
		if f.Kind() == hub.KindAgentMessage {
			var env hub.Envelope
			_ = json.Unmarshal(f.Bytes(), &env)
			var p hub.MessagePayload
			_ = json.Unmarshal(env.Payload, &p)
			var n int
			if _, err := fmt.Sscanf(p.Text, "%d", &n); err != nil || n <= lastText {
				t.Fatalf("callback order broken: %q after %d", p.Text, lastText)
			}
			lastText = n
		}
	}
	if next != total+1 {
		t.Fatalf("stream ended at %d, want %d", next-1, total)
	}
	// No event was dropped: the hub folded every one (entries = messages).
	fresh, err := r.hub.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientRemote, Buffer: 4})
	if err != nil {
		t.Fatal(err)
	}
	if n := entryCount(t, <-fresh.Frames()); n != burst {
		t.Fatalf("hub folded %d message entries, want %d (a dropped event corrupts the fold)", n, burst)
	}
}

func TestIntegrationDegradedPathIsOutOfBandAndHubBuildsErrorEnvelope(t *testing.T) {
	const ingressBuf, prefill, burst = 2, 2, 60
	r := newHubRig(t, ingressBuf)
	defer r.stop(t)

	for i := 0; i < prefill; i++ {
		r.hub.Ingress() <- agent.Event{Kind: agent.KindIdle}
	}
	emitted := make(chan struct{})
	go func() {
		for i := 0; i < burst; i++ {
			emitMessage(r.h.client.session, i)
		}
		close(emitted)
	}()
	select {
	case <-emitted:
		t.Fatal("callback not blocked on saturated ingress")
	case <-time.After(100 * time.Millisecond):
	}

	// Adapter-side fault report while ingress is saturated: it must NOT need
	// ingress (the unavailable resource) and must not block.
	signalled := make(chan struct{})
	go func() { r.hub.SignalDegraded("callback blocked on ingress"); close(signalled) }()
	select {
	case <-signalled:
	case <-time.After(time.Second):
		t.Fatal("degraded signal blocked while ingress was saturated")
	}

	c, errc := (*hub.Client)(nil), make(chan error, 1)
	go func() {
		var err error
		c, err = r.hub.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientTUI, Buffer: 4 * burst})
		errc <- err
	}()
	r.run()
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	select {
	case <-emitted:
	case <-time.After(intWait):
		t.Fatal("callback never unblocked")
	}

	// prefill + burst events + exactly one hub-built error envelope.
	total := uint64(prefill + burst + 1)
	deadline := time.After(intWait)
	var sawError bool
	var last uint64
	for last < total {
		select {
		case f := <-c.Frames():
			if f.Kind() == hub.KindError {
				sawError = true
				if f.Gen() != r.hub.Gen() || f.Seq() == 0 {
					t.Fatalf("error envelope not hub-sequenced: %d %q", f.Seq(), f.Gen())
				}
			}
			last = f.Seq()
		case <-deadline:
			t.Fatalf("stalled at %d/%d", last, total)
		}
	}
	if !sawError {
		// A forced re-hydrate may have superseded the queued error frame, but
		// then its snapshot must record the fault; check via a fresh hydrate.
		fresh, err := r.hub.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientRemote, Buffer: 4})
		if err != nil {
			t.Fatal(err)
		}
		if n := entryCount(t, <-fresh.Frames()); n != burst+1 {
			t.Fatalf("fold has %d entries, want %d messages + 1 error", n, burst+1)
		}
	}
	s, err := r.hub.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s.DegradedSignals != 1 || s.Head != total {
		t.Fatalf("stats = %+v, want 1 degraded signal and head %d", s, total)
	}
}

func TestIntegrationAdapterStopNeverClosesHubIngress(t *testing.T) {
	r := newHubRig(t, 4)
	r.run()
	defer r.stop(t)
	emitMessage(r.h.client.session, 1)
	ctx, cancel := context.WithTimeout(context.Background(), intWait)
	defer cancel()
	if err := r.h.ad.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	// A closed ingress would panic here (send on closed channel).
	select {
	case r.hub.Ingress() <- agent.Event{Kind: agent.KindIdle}:
	case <-time.After(intWait):
		t.Fatal("ingress unusable after adapter Stop")
	}
	deadline := time.Now().Add(intWait)
	for {
		s, err := r.hub.Stats(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if s.Head == 2 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("head = %d, want 2", s.Head)
		}
		time.Sleep(time.Millisecond)
	}
}

// TestIntegrationOracleDetectsLossyIngressSend is the negative control for the
// no-drop assertions above: a select/default (non-blocking) ingress send — the
// exact behaviour §4.2 forbids — does lose events against a saturated ingress,
// and the same accounting used above would flag it.
func TestIntegrationOracleDetectsLossyIngressSend(t *testing.T) {
	hb, err := hub.New(hub.Config{IngressBuffer: 2})
	if err != nil {
		t.Fatal(err)
	}
	delivered := 0
	for i := 0; i < 10; i++ {
		select {
		case hb.Ingress() <- agent.Event{Kind: agent.KindIdle}:
			delivered++
		default:
		}
	}
	if delivered == 10 {
		t.Fatal("oracle cannot detect loss: all 10 sends were accepted by a 2-slot ingress")
	}
}
