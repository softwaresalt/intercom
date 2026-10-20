package hub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// idleHub builds a hub that is never run: tests drive its hub-goroutine-only
// steps directly from the test goroutine, which is then the sole owner.
func idleHub(t *testing.T, cfg Config) *Hub {
	t.Helper()
	if cfg.SessionID == "" {
		cfg.SessionID = "sess"
	}
	h, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func heapInUse() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// ---- 1: caps apply before concatenation; retained strings are clones ----

//go:noinline
func applyHugeStrings(f *Fold) {
	big := strings.Repeat("é", 5<<20) // 10 MiB
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "a", Text: "small"})
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "a", Text: big}) // concat path
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "b", Text: big}) // first-delta path
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: big, Text: "x"})
	f.Apply(agent.Event{Kind: agent.KindToolStarted, ToolCallID: big, ToolName: big})
	f.Apply(agent.Event{Kind: agent.KindTurnStart, TurnID: big})
	f.Apply(agent.Event{Kind: agent.KindAgentMessage, MessageID: big, Text: big})
}

func TestFoldDoesNotPinHugeSDKStrings(t *testing.T) {
	f := mustFold(t, Limits{})
	before := heapInUse()
	applyHugeStrings(f)
	after := heapInUse()
	runtime.KeepAlive(f)
	if after > before && after-before > 2<<20 {
		t.Fatalf("fold retains %d bytes after applying huge SDK strings: retained strings pin their source", after-before)
	}
}

func TestFoldDeltaCapsBeforeConcatenating(t *testing.T) {
	f := mustFold(t, Limits{})
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "a", Text: "small"})
	big := strings.Repeat("x", 10<<20)
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "a", Text: big})
	runtime.ReadMemStats(&m1)
	if d := m1.TotalAlloc - m0.TotalAlloc; d > 1<<20 {
		t.Fatalf("Apply allocated %d bytes for an oversize delta: concatenated before capping", d)
	}
}

func TestFoldDeltaCapIsRuneSafeAcrossBudgetBoundary(t *testing.T) {
	f := mustFold(t, Limits{MaxEntryTextBytes: 1024, MaxBytes: 64 << 10})
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: strings.Repeat("a", 1023)})
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: "éé"}) // 1 byte of budget left
	var snap struct {
		Entries []entry `json:"entries"`
	}
	if err := json.Unmarshal(f.Snapshot(), &snap); err != nil {
		t.Fatal(err)
	}
	e := snap.Entries[0]
	if len(e.Text) != 1023 || !e.Truncated {
		t.Fatalf("text len %d truncated %v, want 1023/true (rune not split)", len(e.Text), e.Truncated)
	}
}

// ---- 6: identifier truncation flag parity; delta storm parity ----

func TestFoldIdentTruncationSetsTruncatedFlags(t *testing.T) {
	f := mustFold(t, Limits{})
	f.Apply(agent.Event{Kind: agent.KindToolStarted, ToolCallID: strings.Repeat("c", 5000), ToolName: "t"})
	var snap struct {
		Entries   []entry `json:"entries"`
		Truncated bool    `json:"truncated"`
	}
	if err := json.Unmarshal(f.Snapshot(), &snap); err != nil {
		t.Fatal(err)
	}
	if !snap.Entries[0].Truncated || !snap.Truncated {
		t.Fatalf("ident truncation not flagged: entry=%v snapshot=%v", snap.Entries[0].Truncated, snap.Truncated)
	}
	// a truncated ident must not stop later text accumulation
	g := mustFold(t, Limits{})
	id := strings.Repeat("m", 5000)
	g.Apply(agent.Event{Kind: agent.KindDelta, MessageID: id, Text: "a"})
	g.Apply(agent.Event{Kind: agent.KindDelta, MessageID: id, Text: "b"})
	if err := json.Unmarshal(g.Snapshot(), &snap); err != nil {
		t.Fatal(err)
	}
	if snap.Entries[0].Text != "ab" {
		t.Fatalf("text = %q, want ab", snap.Entries[0].Text)
	}
}

func TestDeltaStormHydratedAndLiveAgreeOnTruncationAndBound(t *testing.T) {
	h := idleHub(t, Config{})
	chunk := strings.Repeat("d", 100)
	const n = 400 // 40 000 bytes > the 16 KiB cap
	for i := 0; i < n; i++ {
		h.handleEvent(agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: chunk})
	}
	var live strings.Builder
	liveTrunc := false
	for _, f := range h.replayFrom(0) {
		var p MessagePayload
		if err := json.Unmarshal(decode(t, f).Payload, &p); err != nil {
			t.Fatal(err)
		}
		live.WriteString(p.Text)
		liveTrunc = liveTrunc || p.Truncated
	}
	var snap struct {
		Entries   []entry `json:"entries"`
		Truncated bool    `json:"truncated"`
	}
	if err := json.Unmarshal(h.fold.Snapshot(), &snap); err != nil {
		t.Fatal(err)
	}
	folded := snap.Entries[0]
	if !folded.Truncated || !snap.Truncated || !liveTrunc {
		t.Fatalf("truncated parity broken: fold entry=%v snapshot=%v live=%v", folded.Truncated, snap.Truncated, liveTrunc)
	}
	if live.Len() > DefaultMaxEntryTextBytes {
		t.Fatalf("live accumulation is %d bytes, cap %d", live.Len(), DefaultMaxEntryTextBytes)
	}
	if live.String() != folded.Text {
		t.Fatalf("live accumulation (%d bytes) differs from hydrated text (%d bytes)", live.Len(), len(folded.Text))
	}
}

// ---- 2: bounded connect intake ----

func flood(n int, kind ClientKind) []*connectReq {
	reqs := make([]*connectReq, n)
	for i := range reqs {
		reqs[i] = &connectReq{opts: ConnectOptions{Kind: kind}, reply: make(chan connectResp, 1)}
	}
	return reqs
}

func TestIntakeBoundsPendingAndRejectsExcessImmediately(t *testing.T) {
	h := idleHub(t, Config{MaxRemote: 2})
	limit := h.pendingLimit()
	reqs := flood(limit+20, ClientRemote)
	for _, r := range reqs {
		h.intake(r)
	}
	if len(h.pending) > limit {
		t.Fatalf("pending = %d, want <= %d", len(h.pending), limit)
	}
	rejected := 0
	for _, r := range reqs {
		select {
		case resp := <-r.reply:
			var ae *apperr.Error
			if !errors.Is(resp.err, ErrTooManyClients) || !errors.As(resp.err, &ae) || ae.Kind() != apperr.KindPolicy {
				t.Fatalf("excess reply = %v, want KindPolicy ErrTooManyClients", resp.err)
			}
			rejected++
		default:
		}
	}
	if rejected != len(reqs)-len(h.pending) || rejected < 20 {
		t.Fatalf("rejected %d of %d with %d pending", rejected, len(reqs), len(h.pending))
	}
}

func TestIntakeDrainsAFixedBatchPerCall(t *testing.T) {
	h := idleHub(t, Config{MaxRemote: MaxRemoteHardLimit})
	for i := 0; i < connectQueue-1; i++ {
		h.connects <- &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)}
	}
	h.intake(&connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)})
	if len(h.pending) > connectBatch {
		t.Fatalf("one intake took %d requests, batch is %d", len(h.pending), connectBatch)
	}
	if len(h.connects) == 0 {
		t.Fatal("intake drained the whole queue in one call")
	}
}

func TestConnectFloodDoesNotStarveIngressOrStop(t *testing.T) {
	h := startHub(t, Config{MaxRemote: 4, IngressBuffer: 8})
	var stop atomic.Bool
	var wg sync.WaitGroup
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for !stop.Load() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				c, err := h.Connect(ctx, ConnectOptions{Kind: ClientRemote, Buffer: 4})
				cancel()
				if err == nil {
					c.Close()
				}
			}
		}()
	}
	const events = 3000
	sendDone := make(chan error, 1)
	go func() {
		for i := 0; i < events; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := h.Send(ctx, agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: "x"})
			cancel()
			if err != nil {
				sendDone <- err
				return
			}
			if i%500 == 0 {
				h.SignalDegraded("flood")
			}
		}
		sendDone <- nil
	}()
	select {
	case err := <-sendDone:
		if err != nil {
			t.Fatalf("ingress starved under connect flood: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("ingress did not drain under connect flood")
	}
	stop.Store(true)
	wg.Wait()
	if s := settle(t, h, events); s.DegradedSignals == 0 {
		t.Fatal("degraded signal never serviced under flood")
	}
}

// ---- 3: admission is checked before the hydration budget is charged ----

func TestRefusedConnectsDoNotConsumeHydrationBudget(t *testing.T) {
	h := idleHub(t, Config{MaxRemote: 1, HydrationsPerIteration: 1})
	h.register(newClient(100, ClientRemote, 4))
	h.register(newClient(101, ClientTUI, 4))
	var reqs []*connectReq
	for i := 0; i < 3; i++ {
		reqs = append(reqs, &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)})
		reqs = append(reqs, &connectReq{opts: ConnectOptions{Kind: ClientTUI}, reply: make(chan connectResp, 1)})
	}
	h.pending = reqs
	h.serviceHydrations()
	if h.st.HydrationsDeferred != 0 || h.st.MaxHydrationsPerIteration != 0 || len(h.pending) != 0 {
		t.Fatalf("refused connects consumed budget: deferred=%d used=%d pending=%d",
			h.st.HydrationsDeferred, h.st.MaxHydrationsPerIteration, len(h.pending))
	}
	for i, r := range reqs {
		select {
		case resp := <-r.reply:
			if resp.err == nil {
				t.Fatalf("request %d unexpectedly admitted", i)
			}
		default:
			t.Fatalf("request %d got no reply", i)
		}
	}
}

// ---- 4: a replaced or failed frame forces re-hydrate and a degraded boundary ----

func assertStubBoundary(t *testing.T, h *Hub, c *Client, wantSeq uint64) {
	t.Helper()
	if h.seq != wantSeq || h.ringLen == 0 {
		t.Fatalf("seq=%d ringLen=%d, want seq %d published", h.seq, h.ringLen, wantSeq)
	}
	last := h.ring[(h.ringPos-1+len(h.ring))%len(h.ring)]
	if last.Kind() != KindError || last.Seq() != wantSeq || !strings.Contains(string(last.Bytes()), ErrCodeFrameTooLarge) {
		t.Fatalf("last frame = %s seq %d %s, want frame_too_large stub", last.Kind(), last.Seq(), last.Bytes())
	}
	if !c.stale || len(h.staleQ) != 1 {
		t.Fatalf("client stale=%v staleQ=%d, want forced re-hydrate", c.stale, len(h.staleQ))
	}
	if h.degradedSeq != wantSeq {
		t.Fatalf("degradedSeq = %d, want %d", h.degradedSeq, wantSeq)
	}
	if h.replayable(&Position{Gen: h.gen, Seq: wantSeq - 1}, 64) {
		t.Fatal("cursor older than the stub still replays across it")
	}
	if !strings.Contains(string(h.fold.Snapshot()), ErrCodeFrameTooLarge) {
		t.Fatal("stub not recorded in the fold")
	}
}

func TestFrameTooLargeStubMarksClientsStaleAndSetsDegradedBoundary(t *testing.T) {
	h := idleHub(t, Config{MaxFrameBytes: 512})
	c := newClient(1, ClientRemote, 8)
	h.register(c)
	h.handleEvent(msgEvent("m", strings.Repeat("x", 2000)))
	assertStubBoundary(t, h, c, 1)
}

func TestNewFrameFailurePublishesBoundedStubNotSilentDrop(t *testing.T) {
	h := idleHub(t, Config{})
	c := newClient(1, ClientRemote, 8)
	h.register(c)
	seq, ok := h.publish(KindAgentMessage, make(chan int)) // unmarshalable payload
	if !ok || seq != 1 {
		t.Fatalf("publish = (%d, %v), want stub at seq 1", seq, ok)
	}
	assertStubBoundary(t, h, c, 1)
}

// ---- 5 + 7: cursor contract ----

func TestCursorDiscardsLiveFramesAfterGapUntilHydrate(t *testing.T) {
	var c Cursor
	gen := "g"
	if a := c.Apply(mkFrame(t, gen, 10, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("hydrate = %v", a)
	}
	if a := c.Apply(mkFrame(t, gen, 14, KindAgentMessageDelta)); a != ActionRehydrate { // gap
		t.Fatalf("gap = %v, want rehydrate", a)
	}
	for seq := uint64(15); seq < 20; seq++ { // live frames behind the gap are never applied
		if a := c.Apply(mkFrame(t, gen, seq, KindAgentMessageDelta)); a == ActionApply {
			t.Fatalf("seq %d applied while a re-hydrate is pending", seq)
		}
	}
	if _, ok := c.Position(); ok {
		t.Fatal("stale cursor offers a resume position")
	}
	if a := c.Apply(mkFrame(t, gen, 20, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("hydrate after gap = %v", a)
	}
	if a := c.Apply(mkFrame(t, gen, 21, KindAgentMessageDelta)); a != ActionApply {
		t.Fatalf("next live frame = %v, want apply", a)
	}
}

func TestCursorRejectsSameGenHydrateBehindCurrentPosition(t *testing.T) {
	var c Cursor
	c.Apply(mkFrame(t, "g", 50, KindSessionHydrate))
	c.Apply(mkFrame(t, "g", 51, KindAgentMessageDelta))
	if a := c.Apply(mkFrame(t, "g", 40, KindSessionHydrate)); a != ActionRehydrate {
		t.Fatalf("backwards hydrate = %v, want rehydrate", a)
	}
	if !c.NeedsRehydrate() {
		t.Fatal("cursor accepted a backwards hydrate")
	}
	if a := c.Apply(mkFrame(t, "g", 60, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("forward hydrate = %v", a)
	}
	// equal seq and a different generation are legitimate
	if a := c.Apply(mkFrame(t, "g", 60, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("same-seq hydrate = %v", a)
	}
	if a := c.Apply(mkFrame(t, "other", 1, KindSessionHydrate)); a != ActionApply {
		t.Fatalf("new-gen hydrate = %v", a)
	}
}

// ---- 8: only a fixed code and a length are logged ----

func TestDegradedAndUnknownLogsCarryNoAdapterText(t *testing.T) {
	rec := &recordingHandler{}
	h := idleHub(t, Config{Logger: slog.New(rec)})
	h.handleDegraded("https://u:SECRETTOKEN@host/path\x1b[31m")
	h.handleEvent(agent.Event{Kind: agent.KindUnknown, RawType: "SECRETRAW https://x?token=SECRETTOKEN"})
	logged := rec.String()
	if strings.Contains(logged, "SECRET") || strings.Contains(logged, "https") {
		t.Fatalf("adapter text reached the log: %q", logged)
	}
	for _, want := range []string{ErrCodeDegraded, ErrCodeUnknownEvent, "bytes="} {
		if !strings.Contains(logged, want) {
			t.Fatalf("log lacks %q: %q", want, logged)
		}
	}
}

// ---- 9: golden for escaping ----

func TestFramePayloadEscapingMatchesJSONMarshal(t *testing.T) {
	p := MessagePayload{MessageID: "m<1>", Text: "a<b>&c\u2028\"d\""}
	f, err := newFrame("gen", 7, fixedTS(), KindAgentMessage, "sess&<>", p)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	want, err := json.Marshal(Envelope{V: EnvelopeVersion, Gen: "gen", Seq: 7, TS: "2026-09-04T19:12:03Z",
		Kind: KindAgentMessage, SessionID: "sess&<>", Payload: raw})
	if err != nil {
		t.Fatal(err)
	}
	if string(f.Bytes()) != string(want) {
		t.Fatalf("frame differs from json.Marshal:\n got %s\nwant %s", f.Bytes(), want)
	}
	if !strings.Contains(string(f.Bytes()), `\u003c`) || strings.ContainsAny(string(f.Bytes()), "<>&") {
		t.Fatalf("html escaping missing: %s", f.Bytes())
	}
}
