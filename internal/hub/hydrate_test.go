package hub

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestConnectHydratesAtHighWaterMarkThenStreams(t *testing.T) {
	h := startHub(t, Config{})
	for i := 0; i < 5; i++ {
		send(t, h, msgEvent("m"+string(rune('a'+i)), "text"))
	}
	settle(t, h, 5)

	c, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote, Buffer: 16})
	if err != nil {
		t.Fatal(err)
	}
	if c.Mode() != ModeHydrate {
		t.Fatalf("Mode = %v, want ModeHydrate", c.Mode())
	}
	f := recv(t, c)
	if f.Kind() != KindSessionHydrate || f.Seq() != 5 || f.Gen() != h.Gen() {
		t.Fatalf("first frame = %s seq %d gen %s; want hydrate seq 5", f.Kind(), f.Seq(), f.Gen())
	}
	var snap struct {
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(decode(t, f).Payload, &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 5 {
		t.Fatalf("hydrate carries %d entries, want 5", len(snap.Entries))
	}
	send(t, h, msgEvent("next", "t"))
	if g := recv(t, c); g.Seq() != 6 || g.Kind() != KindAgentMessage {
		t.Fatalf("after hydrate got %s seq %d, want agent.message seq 6", g.Kind(), g.Seq())
	}
}

func TestClientRegisteredMidStreamReceivesNoGap(t *testing.T) {
	h := startHub(t, Config{IngressBuffer: 8, MaxRemote: 32})
	const total, nClients = 900, 15 // total+1 hydrate frame fits MaxClientBuffer
	pctx, pcancel := context.WithTimeout(context.Background(), 4*testWait)
	defer pcancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < total; i++ {
			if err := h.Send(pctx, msgEvent("m", "x")); err != nil {
				t.Errorf("producer: %v", err)
				return
			}
		}
	}()
	type result struct {
		hydrateSeq uint64
		got        []uint64
	}
	var mu sync.Mutex
	var results []result
	var cw sync.WaitGroup
	for i := 0; i < nClients; i++ {
		time.Sleep(time.Millisecond)
		c, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote, Buffer: MaxClientBuffer})
		if err != nil {
			t.Fatal(err)
		}
		cw.Add(1)
		go func() {
			defer cw.Done()
			var r result
			first := true
			for {
				select {
				case f := <-c.Frames():
					if first {
						if f.Kind() != KindSessionHydrate {
							t.Errorf("first frame is %s, want session.hydrate", f.Kind())
							return
						}
						r.hydrateSeq, first = f.Seq(), false
					} else {
						r.got = append(r.got, f.Seq())
					}
					// A client that connected after the final event is complete
					// once its hydrate already covers total.
					if f.Seq() == total {
						mu.Lock()
						results = append(results, r)
						mu.Unlock()
						return
					}
				case <-time.After(testWait):
					return
				}
			}
		}()
	}
	wg.Wait()
	cw.Wait()
	if len(results) != nClients {
		t.Fatalf("only %d/%d clients saw the end of the stream", len(results), nClients)
	}
	for i, r := range results {
		want := r.hydrateSeq + 1
		for _, s := range r.got {
			if s != want {
				t.Fatalf("client %d: gap/dup after hydrate seq %d: got %d want %d", i, r.hydrateSeq, s, want)
			}
			want++
		}
	}
}
func TestHydrationsAboveCapAreDeferredNotBlocking(t *testing.T) {
	const clients, perIter = 12, 2
	h, err := New(Config{SessionID: "s", HydrationsPerIteration: perIter, MaxRemote: 64})
	if err != nil {
		t.Fatal(err)
	}
	// Queue every connect request before the hub runs so the first iteration
	// sees them all at once.
	var wg sync.WaitGroup
	got := make(chan *Client, clients)
	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, cerr := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote, Buffer: 8})
			if cerr != nil {
				t.Error(cerr)
				return
			}
			got <- c
		}()
	}
	queueDeadline := time.Now().Add(testWait)
	for len(h.connects) < clients {
		if time.Now().After(queueDeadline) {
			t.Fatalf("only %d/%d connects queued", len(h.connects), clients)
		}
		time.Sleep(time.Millisecond)
	}
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(context.Background()) }()
	defer func() { h.Stop(); <-runDone }()

	// The hub keeps draining ingress while hydrations are being deferred.
	for i := 0; i < 5; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	wg.Wait()
	close(got)
	n := 0
	for c := range got {
		n++
		if c.Mode() != ModeHydrate {
			t.Fatalf("mode = %v", c.Mode())
		}
	}
	if n != clients {
		t.Fatalf("%d/%d connects completed", n, clients)
	}
	s := settle(t, h, 5)
	if s.MaxHydrationsPerIteration > perIter {
		t.Fatalf("hydrations in one iteration = %d, cap %d", s.MaxHydrationsPerIteration, perIter)
	}
	if s.HydrationsDeferred == 0 {
		t.Fatalf("expected deferrals with %d queued connects and cap %d: %+v", clients, perIter, s)
	}
}

func TestOverflowedTUIIsForcedToRehydrate(t *testing.T) {
	h := startHub(t, Config{})
	tui := connect(t, h, ClientTUI, 2)
	for i := 0; i < 50; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, 50)
	// Consume like a TUI pump: a hydrate resets the cursor, every other frame
	// must be exactly the next seq. The stream must end at 50 with at least
	// one forced re-hydrate in it, and never skip.
	// A frame that is not the next seq means the hub dropped frames and a
	// replacement snapshot is pending (the Cursor contract): discard stale
	// live frames until a session.hydrate arrives. The hub may deliver later
	// live frames before that snapshot, which is legal, not a gap.
	var next uint64 = 1
	rehydrates := 0
	forced := 0
	awaiting := false
	deadline := time.After(testWait)
	for next <= 50 {
		select {
		case f := <-tui.Frames():
			if f.Kind() == KindSessionHydrate {
				rehydrates++
				if awaiting || rehydrates > 1 { // first hydrate is the connect snapshot
					forced++
				}
				awaiting = false
				next = f.Seq() + 1
				continue
			}
			if awaiting {
				continue
			}
			if f.Seq() != next {
				if f.Seq() < next {
					t.Fatalf("duplicate/out-of-order: got seq %d, want %d", f.Seq(), next)
				}
				awaiting = true // skipped frames: wait for the forced hydrate
				continue
			}
			next++
		case <-deadline:
			t.Fatalf("TUI stream stalled at seq %d (rehydrates=%d)", next, rehydrates)
		}
	}
	if forced == 0 {
		t.Fatalf("overflowed TUI was never force re-hydrated (rehydrates=%d)", rehydrates)
	}
	if s := stats(t, h); s.Rehydrations == 0 || s.DroppedClients != 0 || !s.TUIAttached {
		t.Fatalf("stats = %+v", s)
	}
}
