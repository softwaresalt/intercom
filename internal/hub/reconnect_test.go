package hub

import (
	"context"
	"testing"
)

func connectResume(t *testing.T, h *Hub, pos *Position, buf int) *Client {
	t.Helper()
	c, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote, Buffer: buf, Resume: pos})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func publishN(t *testing.T, h *Hub, n int) {
	t.Helper()
	before := stats(t, h).Head
	for i := 0; i < n; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, before+uint64(n))
}

func TestReconnectGenMismatchAlwaysHydratesFully(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 64})
	publishN(t, h, 10)
	for _, seq := range []uint64{0, 3, 10} { // all covered by the ring
		c := connectResume(t, h, &Position{Gen: "some-other-gen", Seq: seq}, 32)
		if c.Mode() != ModeHydrate {
			t.Fatalf("seq %d: Mode = %v, want ModeHydrate on gen mismatch", seq, c.Mode())
		}
		if f := recv(t, c); f.Kind() != KindSessionHydrate || f.Seq() != 10 {
			t.Fatalf("seq %d: first frame %s seq %d", seq, f.Kind(), f.Seq())
		}
		c.Close()
	}
}

func TestReconnectCoveredSeqReplaysExactlyThenStreams(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 16})
	live := connect(t, h, ClientRemote, 64)
	publishN(t, h, 10)

	c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 4}, 32)
	if c.Mode() != ModeReplay {
		t.Fatalf("Mode = %v, want ModeReplay", c.Mode())
	}
	for want := uint64(5); want <= 10; want++ {
		got := recv(t, c)
		orig := recv2(t, live, want)
		if got.Seq() != want || got.Kind() != KindAgentMessage {
			t.Fatalf("replay got %s seq %d, want seq %d", got.Kind(), got.Seq(), want)
		}
		if string(got.Bytes()) != string(orig.Bytes()) {
			t.Fatalf("replayed bytes differ from original at seq %d", want)
		}
	}
	send(t, h, msgEvent("m", "x"))
	if f := recv(t, c); f.Seq() != 11 {
		t.Fatalf("after replay got seq %d, want 11 (no gap, no duplicate)", f.Seq())
	}
}

// recv2 reads frames from c until seq want (skipping earlier ones).
func recv2(t *testing.T, c *Client, want uint64) Frame {
	t.Helper()
	for {
		f := recv(t, c)
		if f.Seq() == want {
			return f
		}
	}
}

func TestReconnectAtHeadReplaysNothingThenStreams(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 16})
	publishN(t, h, 5)
	c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 5}, 8)
	if c.Mode() != ModeReplay {
		t.Fatalf("Mode = %v, want ModeReplay", c.Mode())
	}
	send(t, h, msgEvent("m", "x"))
	if f := recv(t, c); f.Seq() != 6 || f.Kind() == KindSessionHydrate {
		t.Fatalf("got %s seq %d, want seq 6", f.Kind(), f.Seq())
	}
}

func TestReconnectFreshHubAtSeqZeroIsReplay(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 4})
	c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 0}, 8)
	if c.Mode() != ModeReplay {
		t.Fatalf("Mode = %v, want ModeReplay", c.Mode())
	}
}

func TestReconnectUncoveredSeqFallsBackToFullHydrate(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 4})
	publishN(t, h, 10) // ring retains seq 7..10
	cases := []struct {
		seq  uint64
		mode ConnectMode
		desc string
	}{
		{2, ModeHydrate, "wrapped out"},
		{5, ModeHydrate, "needs 6, oldest retained is 7"},
		{6, ModeReplay, "needs 7, retained"},
		{9, ModeReplay, "needs 10"},
		{11, ModeHydrate, "ahead of HEAD (cursor from another timeline)"},
	}
	for _, tc := range cases {
		c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: tc.seq}, 16)
		if c.Mode() != tc.mode {
			t.Errorf("seq %d (%s): Mode = %v, want %v", tc.seq, tc.desc, c.Mode(), tc.mode)
		}
		if tc.mode == ModeReplay {
			if f := recv(t, c); f.Seq() != tc.seq+1 {
				t.Errorf("seq %d: first replayed = %d", tc.seq, f.Seq())
			}
		} else if f := recv(t, c); f.Kind() != KindSessionHydrate {
			t.Errorf("seq %d: first frame %s, want hydrate", tc.seq, f.Kind())
		}
		c.Close()
	}
}

func TestReconnectReplayTooLargeForBufferHydratesInsteadOfBlocking(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 64})
	publishN(t, h, 20)
	c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 0}, 4) // 20 frames cannot fit 4 slots
	if c.Mode() != ModeHydrate {
		t.Fatalf("Mode = %v, want ModeHydrate (never block to serve a replay)", c.Mode())
	}
}

// ---- client-side cursor ----

func mkFrame(t *testing.T, gen string, seq uint64, kind Kind) Frame {
	t.Helper()
	f, err := newFrame(gen, seq, fixedTS(), kind, "s", nil)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestCursorAppliesInOrderAndDiscardsDuplicates(t *testing.T) {
	var cur Cursor
	if a := cur.Apply(mkFrame(t, "g", 7, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("hydrate = %v, want apply", a)
	}
	steps := []struct {
		seq  uint64
		want Action
	}{{8, ActionApply}, {9, ActionApply}, {9, ActionDiscard}, {8, ActionDiscard}, {10, ActionApply}}
	for _, s := range steps {
		if got := cur.Apply(mkFrame(t, "g", s.seq, KindAgentMessage)); got != s.want {
			t.Fatalf("seq %d: %v, want %v", s.seq, got, s.want)
		}
	}
	if p, ok := cur.Position(); !ok || p != (Position{Gen: "g", Seq: 10}) {
		t.Fatalf("Position = %+v %v", p, ok)
	}
}

func TestCursorOutOfOrderReceiptTriggersRehydrate(t *testing.T) {
	var cur Cursor
	cur.Apply(mkFrame(t, "g", 100, KindSessionHydrate))
	// 105 before 101..104: a gap, not something to apply.
	if a := cur.Apply(mkFrame(t, "g", 105, KindAgentMessage)); a != ActionRehydrate {
		t.Fatalf("105 after 100 = %v, want rehydrate", a)
	}
	// 99 afterwards must not be silently discarded as "already applied" and
	// forgotten: the cursor stays in the rehydrate-required state.
	if a := cur.Apply(mkFrame(t, "g", 99, KindAgentMessage)); a != ActionRehydrate {
		t.Fatalf("late 99 = %v, want rehydrate (sticky)", a)
	}
	if !cur.NeedsRehydrate() {
		t.Fatal("NeedsRehydrate = false")
	}
	if a := cur.Apply(mkFrame(t, "g", 110, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("hydrate = %v", a)
	}
	if cur.NeedsRehydrate() {
		t.Fatal("hydrate did not clear NeedsRehydrate")
	}
	if a := cur.Apply(mkFrame(t, "g", 111, KindAgentMessage)); a != ActionApply {
		t.Fatalf("111 after hydrate 110 = %v", a)
	}
}

func TestCursorGenChangeAndNoBaselineForceRehydrate(t *testing.T) {
	var cur Cursor
	if a := cur.Apply(mkFrame(t, "g", 1, KindAgentMessage)); a != ActionRehydrate {
		t.Fatalf("frame before any hydrate = %v, want rehydrate", a)
	}
	cur.Apply(mkFrame(t, "g", 5, KindSessionHydrate))
	if a := cur.Apply(mkFrame(t, "g2", 6, KindAgentMessage)); a != ActionRehydrate {
		t.Fatalf("gen change = %v, want rehydrate", a)
	}
	if _, ok := (&Cursor{}).Position(); ok {
		t.Fatal("empty cursor reports a Position")
	}
}

func TestCursorDrivenReconnectContinuesWithoutRehydrate(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 64})
	c1 := connect(t, h, ClientRemote, 64)
	var cur Cursor
	cur.Apply(mkFrame(t, h.Gen(), 0, KindSessionHydrate)) // baseline from the initial hydrate
	publishN(t, h, 6)
	for i := 0; i < 3; i++ {
		if a := cur.Apply(recv(t, c1)); a != ActionApply {
			t.Fatalf("frame %d: %v", i, a)
		}
	}
	c1.Close() // connection lost after seq 3
	publishN(t, h, 4)

	pos, ok := cur.Position()
	if !ok {
		t.Fatal("no position")
	}
	c2 := connectResume(t, h, &pos, 32)
	if c2.Mode() != ModeReplay {
		t.Fatalf("Mode = %v, want ModeReplay", c2.Mode())
	}
	for want := uint64(4); want <= 10; want++ {
		f := recv(t, c2)
		if a := cur.Apply(f); a != ActionApply || f.Seq() != want {
			t.Fatalf("seq %d: action %v want apply seq %d", f.Seq(), a, want)
		}
	}
}
