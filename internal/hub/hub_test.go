package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

const testWait = 5 * time.Second

// startHub constructs a hub, runs it, and registers a cleanup that stops it
// and fails the test if goroutines leak.
func startHub(t *testing.T, cfg Config) *Hub {
	t.Helper()
	if cfg.SessionID == "" {
		cfg.SessionID = "sess"
	}
	base := runtime.NumGoroutine()
	h, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(context.Background()) }()
	t.Cleanup(func() {
		h.Stop()
		select {
		case <-runDone:
		case <-time.After(testWait):
			t.Error("Run did not return after Stop")
			return
		}
		deadline := time.Now().Add(2 * time.Second)
		for runtime.NumGoroutine() > base {
			if time.Now().After(deadline) {
				buf := make([]byte, 1<<16)
				n := runtime.Stack(buf, true)
				t.Errorf("goroutine leak: %d > baseline %d\n%s", runtime.NumGoroutine(), base, buf[:n])
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
	return h
}

func recv(t *testing.T, c *Client) Frame {
	t.Helper()
	select {
	case f := <-c.Frames():
		return f
	case <-time.After(testWait):
		t.Fatal("timed out waiting for a frame")
		return Frame{}
	}
}

// connect registers a client and discards the initial hydrate frame, if any.
func connect(t *testing.T, h *Hub, kind ClientKind, buf int) *Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	c, err := h.Connect(ctx, ConnectOptions{Kind: kind, Buffer: buf})
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	if c.Mode() == ModeHydrate {
		if f := recv(t, c); f.Kind() != KindSessionHydrate {
			t.Fatalf("first frame kind = %q, want hydrate", f.Kind())
		}
	}
	return c
}

func send(t *testing.T, h *Hub, ev agent.Event) {
	t.Helper()
	select {
	case h.Ingress() <- ev:
	case <-time.After(testWait):
		t.Fatal("timed out sending to ingress")
	}
}

func msgEvent(id string, text string) agent.Event {
	return agent.Event{Kind: agent.KindAgentMessage, MessageID: id, Text: text}
}

func decode(t *testing.T, f Frame) Envelope {
	t.Helper()
	var e Envelope
	if err := json.Unmarshal(f.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	return e
}

func stats(t *testing.T, h *Hub) Stats {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	s, err := h.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	return s
}

// settle polls Stats (each a hub step) until Head reaches want, i.e. every
// event sent so far has been processed by the hub.
func settle(t *testing.T, h *Hub, want uint64) Stats {
	t.Helper()
	deadline := time.Now().Add(testWait)
	for {
		s := stats(t, h)
		if s.Head >= want {
			return s
		}
		if time.Now().After(deadline) {
			t.Fatalf("Head = %d, want %d", s.Head, want)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestSeqMonotonicAndDeliveryOrderMatchesSeq(t *testing.T) {
	h := startHub(t, Config{IngressBuffer: 16})
	const producers, perProducer = 4, 250
	const total = producers * perProducer
	clients := []*Client{
		connect(t, h, ClientTUI, total+8),
		connect(t, h, ClientRemote, total+8),
		connect(t, h, ClientRemote, total+8),
	}
	for p := 0; p < producers; p++ {
		go func(p int) {
			for i := 0; i < perProducer; i++ {
				h.Ingress() <- msgEvent(fmt.Sprintf("p%d", p), fmt.Sprintf("%d", i))
			}
		}(p)
	}
	for ci, c := range clients {
		lastPerProducer := map[string]int{}
		for want := uint64(1); want <= total; want++ {
			f := recv(t, c)
			if f.Seq() != want || f.Gen() != h.Gen() {
				t.Fatalf("client %d: got (gen %s, seq %d), want (gen %s, seq %d)", ci, f.Gen(), f.Seq(), h.Gen(), want)
			}
			var p MessagePayload
			if err := json.Unmarshal(decode(t, f).Payload, &p); err != nil {
				t.Fatal(err)
			}
			var n int
			if _, err := fmt.Sscanf(p.Text, "%d", &n); err != nil {
				t.Fatal(err)
			}
			if prev, ok := lastPerProducer[p.MessageID]; ok && n != prev+1 {
				t.Fatalf("client %d: producer %s order broken: %d after %d", ci, p.MessageID, n, prev)
			}
			lastPerProducer[p.MessageID] = n
		}
	}
	if s := stats(t, h); s.Head != total {
		t.Fatalf("Head = %d, want %d", s.Head, total)
	}
}

func TestUnknownEventIsPublishedAsErrorEnvelope(t *testing.T) {
	h := startHub(t, Config{})
	c := connect(t, h, ClientRemote, 8)
	send(t, h, agent.Event{Kind: agent.KindUnknown, RawType: "assistant.reasoning"})
	f := recv(t, c)
	if f.Kind() != KindError || f.Seq() != 1 {
		t.Fatalf("frame = %s seq %d, want error seq 1", f.Kind(), f.Seq())
	}
}

func TestTUIClientIsNeverDroppedForBackpressure(t *testing.T) {
	h := startHub(t, Config{})
	tui := connect(t, h, ClientTUI, 2)
	for i := 0; i < 50; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	s := settle(t, h, 50)
	if s.Head != 50 {
		t.Fatalf("Head = %d, want 50", s.Head)
	}
	if !s.TUIAttached || s.DroppedClients != 0 {
		t.Fatalf("TUI dropped: %+v", s)
	}
	if s.TUIOverflows == 0 {
		t.Fatalf("expected overflow handling to have run: %+v", s)
	}
	select {
	case <-tui.Done():
		t.Fatal("TUI client done channel closed by hub")
	default:
	}
}

func TestRemoteClientDroppedWhenBufferFullOthersUnaffected(t *testing.T) {
	h := startHub(t, Config{})
	slow := connect(t, h, ClientRemote, 2)
	healthy := connect(t, h, ClientRemote, 64)
	tui := connect(t, h, ClientTUI, 64)
	for i := 0; i < 20; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	s := settle(t, h, 20)
	if s.DroppedClients != 1 || s.Remote != 1 || !s.TUIAttached {
		t.Fatalf("stats = %+v, want 1 dropped, 1 remote, TUI attached", s)
	}
	select {
	case <-slow.Done():
	case <-time.After(testWait):
		t.Fatal("slow client's done channel was not closed")
	}
	// The send channel is never closed by the hub: after draining buffered
	// frames the receive must block (default), not return a closed zero value.
	for drained := false; !drained; {
		select {
		case _, ok := <-slow.Frames():
			if !ok {
				t.Fatal("hub closed the client's send channel")
			}
		default:
			drained = true
		}
	}
	for _, c := range []*Client{healthy, tui} {
		for want := uint64(1); want <= 20; want++ {
			if f := recv(t, c); f.Seq() != want {
				t.Fatalf("seq = %d, want %d", f.Seq(), want)
			}
		}
	}
}

func TestUnregistrationIsIdempotent(t *testing.T) {
	h := startHub(t, Config{})
	c := connect(t, h, ClientRemote, 2)
	if s := stats(t, h); s.Remote != 1 {
		t.Fatalf("precondition: Remote = %d, want 1", s.Remote)
	}
	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func() { c.Close(); done <- struct{}{} }()
	}
	for i := 0; i < 8; i++ {
		<-done
	}
	for i := 0; i < 10; i++ { // hub-side drop racing an already closed client
		send(t, h, msgEvent("m", "x"))
	}
	c.Close()
	if s := stats(t, h); s.Remote != 0 || s.DroppedClients != 0 {
		t.Fatalf("stats = %+v, want 0 remote and 0 hub drops (client closed itself)", s)
	}
}

func TestSecondTUIAttachIsRejected(t *testing.T) {
	h := startHub(t, Config{})
	first := connect(t, h, ClientTUI, 4)
	_, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientTUI})
	if err == nil {
		t.Fatal("second TUI attach succeeded")
	}
	first.Close() // the TUI exiting frees the slot
	if s := stats(t, h); s.TUIAttached {
		t.Fatalf("TUI still attached after Close: %+v", s)
	}
	connect(t, h, ClientTUI, 4)
}

func TestRunTwiceFailsAndStopIsIdempotent(t *testing.T) {
	h := startHub(t, Config{})
	stats(t, h) // Run is up: the next Run must be the second
	if err := h.Run(context.Background()); err == nil {
		t.Fatal("second Run succeeded")
	}
	h.Stop()
	h.Stop()
	select {
	case <-h.Done():
	case <-time.After(testWait):
		t.Fatal("Done not closed after Stop")
	}
	if _, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote}); err == nil {
		t.Fatal("Connect after Stop succeeded")
	}
}

func TestShutdownPublishesServerShutdownEnvelope(t *testing.T) {
	cfg := Config{SessionID: "sess"}
	h, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(ctx) }()
	c := connect(t, h, ClientRemote, 8)
	cancel()
	f := recv(t, c)
	if f.Kind() != KindServerShutdown || f.Seq() != 1 {
		t.Fatalf("frame = %s seq %d, want server.shutdown seq 1", f.Kind(), f.Seq())
	}
	<-runDone
}
