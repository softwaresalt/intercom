package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- A: Send is bounded by shutdown and ctx ----

func TestSendAfterStopReturnsPromptly(t *testing.T) {
	h, err := New(Config{SessionID: "s", IngressBuffer: 1})
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(context.Background()) }()
	stats(t, h)
	h.Stop()
	<-runDone
	for i := 0; i < 3; i++ { // ingress has room for one: none may block or succeed silently
		ctx, cancel := context.WithTimeout(context.Background(), testWait)
		start := time.Now()
		err := h.Send(ctx, msgEvent("m", "x"))
		cancel()
		if !errors.Is(err, ErrStopped) {
			t.Fatalf("Send #%d after stop = %v, want ErrStopped", i, err)
		}
		if time.Since(start) > time.Second {
			t.Fatalf("Send after stop took %v", time.Since(start))
		}
	}
}

func TestSendWithSaturatedIngressHonoursContext(t *testing.T) {
	h, err := New(Config{IngressBuffer: 1}) // never run: nothing drains
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Send(context.Background(), msgEvent("m", "x")); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	err = h.Send(ctx, msgEvent("m", "y"))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Send = %v, want context deadline", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("Send ignored ctx for %v", time.Since(start))
	}
	// A stop releases a Send blocked on a saturated ingress too.
	released := make(chan error, 1)
	go func() { released <- h.Send(context.Background(), msgEvent("m", "z")) }()
	time.Sleep(20 * time.Millisecond)
	h.Stop()
	select {
	case err := <-released:
		if !errors.Is(err, ErrStopped) {
			t.Fatalf("blocked Send released with %v, want ErrStopped", err)
		}
	case <-time.After(testWait):
		t.Fatal("Send stayed wedged after Stop")
	}
}

// ---- B: degraded boundary forbids replay across it ----

func TestReconnectCursorOlderThanDegradedEventFullHydrates(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 64})
	publishN(t, h, 3)
	h.SignalDegraded("x")
	settle(t, h, 4) // degraded error envelope is seq 4
	publishN(t, h, 2)
	cases := []struct {
		seq  uint64
		mode ConnectMode
	}{{2, ModeHydrate}, {3, ModeHydrate}, {4, ModeReplay}, {5, ModeReplay}, {6, ModeReplay}}
	for _, tc := range cases {
		c := connectResume(t, h, &Position{Gen: h.Gen(), Seq: tc.seq}, 32)
		if c.Mode() != tc.mode {
			t.Errorf("cursor %d: Mode = %v, want %v", tc.seq, c.Mode(), tc.mode)
		}
		c.Close()
	}
}

// ---- C: replay never exactly fills the buffer ----

func TestReplayRequiresOneSlotOfHeadroom(t *testing.T) {
	h := startHub(t, Config{ReplayRing: 64})
	publishN(t, h, 8)
	full := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 0}, 8) // n == buffer
	if full.Mode() != ModeHydrate {
		t.Fatalf("n == buffer: Mode = %v, want ModeHydrate", full.Mode())
	}
	full.Close()
	fits := connectResume(t, h, &Position{Gen: h.Gen(), Seq: 1}, 8) // n == buffer-1
	if fits.Mode() != ModeReplay {
		t.Fatalf("n == buffer-1: Mode = %v, want ModeReplay", fits.Mode())
	}
}

// ---- D: Connect is bounded by ctx even if Run never started ----

func TestConnectHonoursContextWhenHubNotRunning(t *testing.T) {
	h, err := New(Config{SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res := make(chan error, 1)
	go func() {
		_, err := h.Connect(ctx, ConnectOptions{Kind: ClientRemote})
		res <- err
	}()
	select {
	case err := <-res:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Connect = %v, want context deadline", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Connect wedged on a hub that is not running")
	}
	// The abandoned request is discarded, not registered, once Run starts.
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(context.Background()) }()
	defer func() { h.Stop(); <-runDone }()
	if s := stats(t, h); s.Remote != 0 {
		t.Fatalf("abandoned connect registered a client: %+v", s)
	}
}

func TestAbandonedConnectIsNeverRegistered(t *testing.T) {
	g, err := New(Config{SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	req := &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)}
	req.abandoned.Store(true) // the requester gave up before the hub completed it
	g.completeConnect(req, false)
	if len(g.clients) != 0 {
		t.Fatalf("abandoned client left registered: %d", len(g.clients))
	}
	select {
	case r := <-req.reply:
		if r.client != nil && !r.client.isClosed() {
			t.Fatal("client for an abandoned request was left open")
		}
	default:
	}
}

// ---- E: shutdown closes every client's done ----

func TestShutdownClosesEveryClientDoneIncludingStale(t *testing.T) {
	h, err := New(Config{SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan error, 1)
	go func() { runDone <- h.Run(context.Background()) }()
	remote := connect(t, h, ClientRemote, 16)
	tui := connect(t, h, ClientTUI, 2)
	for i := 0; i < 6; i++ { // overflows the TUI: it is stale at shutdown
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, 6)
	h.Stop()
	<-runDone
	for name, c := range map[string]*Client{"remote": remote, "tui": tui} {
		select {
		case <-c.Done():
		case <-time.After(time.Second):
			t.Fatalf("%s client's Done not closed at shutdown", name)
		}
	}
	// The shutdown frame was queued ahead of the close for the healthy client.
	var last Frame
	for drained := false; !drained; {
		select {
		case f := <-remote.Frames():
			last = f
		default:
			drained = true
		}
	}
	if last.Kind() != KindServerShutdown {
		t.Fatalf("last frame = %s, want server.shutdown", last.Kind())
	}
}

// ---- F: Stop cannot return before a concurrently starting Run finished ----

func TestStopNeverReturnsBeforeConcurrentRunFinished(t *testing.T) {
	for i := 0; i < 500; i++ {
		h, err := New(Config{SessionID: "s"})
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); _ = h.Run(context.Background()) }()
		go func() {
			defer wg.Done()
			h.Stop()
			select {
			case <-h.Done():
			default:
				t.Error("Stop returned but Done is not closed")
			}
		}()
		wg.Wait()
	}
}

// ---- G: live payload caps, frame cap, byte-capped ring ----

func TestLiveEnvelopeTextAndIdentsAreTruncatedWithFlag(t *testing.T) {
	h := startHub(t, Config{Limits: Limits{MaxEntryTextBytes: 1024, MaxBytes: 64 << 10}})
	c := connect(t, h, ClientRemote, 8)
	big := strings.Repeat("<", 1<<20) // 1 MiB; html-escapes 6x on the wire
	send(t, h, agent.Event{Kind: agent.KindAgentMessage, MessageID: strings.Repeat("i", 5000), Text: big})
	send(t, h, agent.Event{Kind: agent.KindToolStarted, ToolCallID: strings.Repeat("c", 5000), ToolName: strings.Repeat("n", 5000)})
	send(t, h, agent.Event{Kind: agent.KindTurnStart, TurnID: strings.Repeat("t", 5000)})
	for i, kind := range []Kind{KindAgentMessage, KindToolStarted, KindAgentTurnStart} {
		f := recv(t, c)
		if f.Kind() != kind {
			t.Fatalf("frame %d kind %s want %s", i, f.Kind(), kind)
		}
		if f.Len() > 16<<10 {
			t.Fatalf("frame %d is %d bytes: live payload not capped", i, f.Len())
		}
		if !strings.Contains(string(f.Bytes()), `"truncated":true`) {
			t.Fatalf("frame %d lacks truncated:true", i)
		}
	}
}

func TestPublishEnforcesMaxFrameBytesWithBoundedStub(t *testing.T) {
	h, err := New(Config{SessionID: "s", MaxFrameBytes: 8 << 10})
	if err != nil {
		t.Fatal(err)
	}
	h.publish(KindAgentMessage, MessagePayload{MessageID: "m", Text: strings.Repeat("x", 1<<20)})
	if h.seq != 1 || h.ringLen != 1 {
		t.Fatalf("seq=%d ringLen=%d: oversize frame must still occupy its seq", h.seq, h.ringLen)
	}
	f := h.ring[0]
	if f.Len() > 8<<10 || f.Seq() != 1 || f.Kind() != KindError {
		t.Fatalf("frame = %s seq %d %d bytes, want bounded error stub at seq 1", f.Kind(), f.Seq(), f.Len())
	}
}

func TestReplayRingIsCappedByTotalBytes(t *testing.T) {
	const capBytes = 10_000
	h, err := New(Config{SessionID: "s", ReplayRing: 1024, ReplayRingBytes: capBytes})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		h.publish(KindAgentMessage, MessagePayload{MessageID: "m", Text: strings.Repeat("x", 900)})
		if h.ringBytes > capBytes {
			t.Fatalf("ring holds %d bytes, cap %d (i=%d)", h.ringBytes, capBytes, i)
		}
	}
	if h.ringLen >= 200 || h.ringLen == 0 {
		t.Fatalf("ringLen = %d: expected byte eviction to retain only a tail", h.ringLen)
	}
	// Coverage uses the actually retained range.
	oldest := h.seq - uint64(h.ringLen)
	if !h.replayable(&Position{Gen: h.gen, Seq: oldest}, 1024) {
		t.Fatalf("cursor %d (oldest retained - 1) should replay", oldest)
	}
	if h.replayable(&Position{Gen: h.gen, Seq: oldest - 1}, 1024) {
		t.Fatalf("cursor %d precedes the byte-evicted range but replays", oldest-1)
	}
	got := h.replayFrom(oldest)
	if len(got) != h.ringLen || got[0].Seq() != oldest+1 || got[len(got)-1].Seq() != h.seq {
		t.Fatalf("replayFrom returned wrong range: %d frames", len(got))
	}
}

// ---- H: remote client cap, buffer clamp ----

func TestMaxRemoteRefusesBeyondCapButTUIIsExempt(t *testing.T) {
	h := startHub(t, Config{MaxRemote: 2})
	connect(t, h, ClientRemote, 4)
	connect(t, h, ClientRemote, 4)
	_, err := h.Connect(context.Background(), ConnectOptions{Kind: ClientRemote})
	if !errors.Is(err, ErrTooManyClients) {
		t.Fatalf("third remote = %v, want ErrTooManyClients", err)
	}
	connect(t, h, ClientTUI, 4) // does not count
}

func TestMaxRemoteDefaultsAndClamps(t *testing.T) {
	h, _ := New(Config{})
	if h.cfg.MaxRemote != DefaultMaxRemote {
		t.Fatalf("default MaxRemote = %d", h.cfg.MaxRemote)
	}
	h, _ = New(Config{MaxRemote: 1 << 30})
	if h.cfg.MaxRemote != MaxRemoteHardLimit {
		t.Fatalf("clamped MaxRemote = %d", h.cfg.MaxRemote)
	}
	if _, err := New(Config{MaxRemote: -1}); err == nil {
		t.Fatal("negative MaxRemote accepted")
	}
}

func TestClientBufferIsClamped(t *testing.T) {
	if got := bufferFor(ClientRemote, 1<<30); got != MaxClientBuffer {
		t.Fatalf("bufferFor(huge) = %d, want %d", got, MaxClientBuffer)
	}
	h := startHub(t, Config{})
	c := connect(t, h, ClientRemote, 1<<30)
	if cap(c.send) != MaxClientBuffer {
		t.Fatalf("client capacity = %d", cap(c.send))
	}
}

// ---- I: fixed wire messages; detail only in the log ----

type recordingHandler struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (r *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (r *recordingHandler) Handle(_ context.Context, rec slog.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fmt.Fprintf(&r.buf, "%s", rec.Message)
	rec.Attrs(func(a slog.Attr) bool { fmt.Fprintf(&r.buf, " %s=%v", a.Key, a.Value); return true })
	r.buf.WriteByte('\n')
	return nil
}
func (r *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return r }
func (r *recordingHandler) WithGroup(string) slog.Handler      { return r }
func (r *recordingHandler) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func TestAdapterDetailNeverReachesWireOrFoldOnlyLog(t *testing.T) {
	rec := &recordingHandler{}
	h := startHub(t, Config{Logger: slog.New(rec)})
	c := connect(t, h, ClientRemote, 16)
	h.SignalDegraded("SECRETREASON\x1b[31m")
	send(t, h, agent.Event{Kind: agent.KindUnknown, RawType: "SECRETRAW"})
	var msgs []string
	for i := 0; i < 4; i++ { // degraded error, hydrate (forced), unknown error ...
		f := recv(t, c)
		s := string(f.Bytes())
		if strings.Contains(s, "SECRET") {
			t.Fatalf("adapter-controlled text reached the wire: %s", s)
		}
		if f.Kind() == KindError {
			var p ErrorPayload
			if err := json.Unmarshal(decode(t, f).Payload, &p); err != nil {
				t.Fatal(err)
			}
			msgs = append(msgs, p.Message)
		}
		if len(msgs) == 2 {
			break
		}
	}
	want := []string{"agent event path degraded", "unrecognised agent event"}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("wire messages = %q, want %q", msgs, want)
	}
	fresh := connect(t, h, ClientRemote, 4)
	_ = fresh
	if s := stats(t, h); s.Head < 2 {
		t.Fatal("events not processed")
	}
	logged := rec.String()
	if strings.Contains(logged, "SECRET") || strings.ContainsRune(logged, 0x1b) {
		t.Fatalf("adapter text reached the log: %q", logged)
	}
	if !strings.Contains(logged, ErrCodeDegraded) || !strings.Contains(logged, ErrCodeUnknownEvent) {
		t.Fatalf("fixed codes missing from the log: %q", logged)
	}
}

func TestFoldDoesNotRecordAdapterTextForUnknownEvents(t *testing.T) {
	f := mustFold(t, Limits{})
	f.Apply(agent.Event{Kind: agent.KindUnknown, RawType: "SECRETRAW"})
	if strings.Contains(string(f.Snapshot()), "SECRET") {
		t.Fatalf("fold leaks raw type: %s", f.Snapshot())
	}
}

// ---- J: sanitize ----

func TestSanitizeReplacesC1BidiAndLineSeparators(t *testing.T) {
	in := "a\u009bb\u202ec\u2028d\u2029e\u200bf\x1b\x7fg\u0085h"
	got := sanitize(in, 100)
	if got != "a?b?c?d?e?f??g?h" {
		t.Fatalf("sanitize = %q", got)
	}
}

func TestSanitizeBoundsBytesOnRuneBoundary(t *testing.T) {
	got := sanitize("ééééé", 5) // 2 bytes each: at most two runes
	if got != "éé" {
		t.Fatalf("sanitize = %q", got)
	}
}

// ---- K: snapshot never exceeds bound at the caps ----

func TestSnapshotSizeBoundHoldsWithManySmallEntries(t *testing.T) {
	l := Limits{MaxEntries: 100_000, MaxBytes: 300_000}
	f := mustFold(t, l)
	for i := 0; i < 20_000; i++ {
		f.Apply(agent.Event{Kind: agent.KindToolStarted, ToolCallID: fmt.Sprintf("c%06d", i), ToolName: "b"})
		if i%50 != 0 && i < 4000 {
			continue
		}
		if n := len(f.Snapshot()); n > l.MaxBytes+SnapshotOverhead {
			t.Fatalf("snapshot %d > %d+%d at i=%d (entries=%d)", n, l.MaxBytes, SnapshotOverhead, i, f.Len())
		}
	}
	if f.EvictedCount() == 0 {
		t.Fatal("byte cap never engaged")
	}
}

// ---- L: tool.ended always carries success ----

func TestToolEndedAlwaysCarriesSuccessAndStartedOmitsIt(t *testing.T) {
	k, p := mapEvent(agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c", ToolName: "n"}, 1024)
	b, _ := json.Marshal(p)
	if k != KindToolEnded || !strings.Contains(string(b), `"success":false`) {
		t.Fatalf("tool.ended payload = %s", b)
	}
	_, p = mapEvent(agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c", ToolName: "n"}, 1024)
	b, _ = json.Marshal(p)
	if strings.Contains(string(b), "success") {
		t.Fatalf("tool.started payload carries success: %s", b)
	}
}

// ---- N: injectable rate-limit clock ----

func TestStaleRehydrateRateLimitUsesConfigNow(t *testing.T) {
	var now atomic.Int64
	base := time.Unix(1_000_000, 0)
	h := startHub(t, Config{Now: func() time.Time { return base.Add(time.Duration(now.Load())) }})
	tui := connect(t, h, ClientTUI, 2)
	for i := 0; i < 4; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, 4)
	waitRehydrations := func(n uint64) bool {
		deadline := time.Now().Add(testWait)
		for time.Now().Before(deadline) {
			if stats(t, h).Rehydrations >= n {
				return true
			}
			time.Sleep(5 * time.Millisecond)
		}
		return false
	}
	if !waitRehydrations(1) {
		t.Fatal("first re-hydrate never served")
	}
	// Overflow again: the buffer holds the hydrate plus one frame.
	for i := 0; i < 4; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, 8)
	time.Sleep(150 * time.Millisecond) // far beyond staleRetryInterval in real time
	if r := stats(t, h).Rehydrations; r != 1 {
		t.Fatalf("Rehydrations = %d with a frozen injected clock, want 1 (rate limit must use Config.Now)", r)
	}
	now.Store(int64(time.Second))
	if !waitRehydrations(2) {
		t.Fatal("re-hydrate not served after the injected clock advanced")
	}
	_ = tui
}

// ---- 075.001-T: intake sweep, pending-connect reap guard, terminal shutdown ----

func TestIntakeSweepsClosedRemotesBeforeCheckingPendingLimit(t *testing.T) {
	g, err := New(Config{SessionID: "s", MaxRemote: 2})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < g.cfg.MaxRemote; i++ {
		req := &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)}
		g.completeConnect(req, false)
		r := <-req.reply
		if r.err != nil || r.client == nil {
			t.Fatalf("seed connect %d failed: %v", i, r.err)
		}
		r.client.Close() // closed, but the hub has not stepped since
	}
	reqs := make([]*connectReq, 0, g.pendingLimit())
	for i := 0; i < g.pendingLimit(); i++ {
		req := &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)}
		reqs = append(reqs, req)
		g.intake(req)
	}
	for i, req := range reqs {
		select {
		case r := <-req.reply:
			if errors.Is(r.err, ErrTooManyClients) {
				t.Fatalf("request %d falsely refused with ErrTooManyClients: closed remotes were not swept at intake", i)
			}
		default:
		}
	}
}

func TestStaleClientReceivesServerShutdownBeforeClose(t *testing.T) {
	var now atomic.Int64
	base := time.Unix(1_000_000, 0)
	h := startHub(t, Config{Now: func() time.Time { return base.Add(time.Duration(now.Load())) }})
	tui := connect(t, h, ClientTUI, 2)
	for i := 0; i < 4; i++ { // overflows the TUI twice; the frozen clock keeps it stale
		send(t, h, msgEvent("m", "x"))
	}
	settle(t, h, 4)
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.AnnounceShutdown(ctx, "test"); err != nil {
		t.Fatal(err)
	}
	now.Store(int64(time.Second)) // past staleRetryInterval
	send(t, h, msgEvent("m", "after"))
	for i := 0; i < 3; i++ { // drive further hub-loop iterations and hydration passes
		stats(t, h)
		time.Sleep(40 * time.Millisecond)
	}
	var kinds []string
	for drained := false; !drained; {
		select {
		case f := <-tui.Frames():
			kinds = append(kinds, string(f.Kind()))
		default:
			drained = true
		}
	}
	for _, k := range kinds {
		if k == string(KindServerShutdown) {
			return
		}
	}
	t.Fatalf("stale client never received server.shutdown intact; frames = %v", kinds)
}

func TestReapEmptyBroadcastSetWaitsForQueuedPendingConnectAndReapsOnAbandon(t *testing.T) {
	g, err := New(Config{SessionID: "s", PermissionTimeout: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	defer close(a.resolved)
	g.handleAsk(a.ask)
	req := &connectReq{opts: ConnectOptions{Kind: ClientRemote}, reply: make(chan connectResp, 1)}
	g.pending = append(g.pending, req)
	if !g.hasPendingConnect() {
		t.Fatal("hasPendingConnect = false with a live queued connect")
	}
	g.reapUnlessConnecting()
	select {
	case d := <-a.reply:
		t.Fatalf("request resolved (%v) while a connect was queued", d.Kind)
	default:
	}
	req.abandoned.Store(true)
	if g.hasPendingConnect() {
		t.Fatal("hasPendingConnect = true for an abandoned connect")
	}
	g.serviceHydrations() // discards the abandoned request
	g.reapUnlessConnecting()
	select {
	case d := <-a.reply:
		if d.Kind != agent.DecisionUnavailable {
			t.Fatalf("decision = %v, want unavailable", d.Kind)
		}
	default:
		t.Fatal("request not resolved no_surface after the connect was abandoned")
	}
}

func TestHubRateLimitsUnmappedEventWarnLogsPerType(t *testing.T) {
	rec := &recordingHandler{}
	h := startHub(t, Config{Logger: slog.New(rec)})
	c := connect(t, h, ClientRemote, 16)
	for _, raw := range []string{"dup.type", "dup.type", "other.type", "dup.type"} {
		send(t, h, agent.Event{Kind: agent.KindUnknown, RawType: raw})
	}
	// Every unmapped event is still folded and published as an error envelope.
	errs := 0
	for errs < 4 {
		if recv(t, c).Kind() == KindError {
			errs++
		}
	}
	logged := rec.String()
	if n := strings.Count(logged, "hub: unmapped agent event code="); n != 2 {
		t.Errorf("warn lines = %d, want 2 (one per distinct type)\n%s", n, logged)
	}
	if n := strings.Count(logged, "hub: unmapped agent event (repeat) code="); n != 2 {
		t.Errorf("repeat debug lines = %d, want 2\n%s", n, logged)
	}
}
