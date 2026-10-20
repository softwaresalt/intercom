package hub

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// Sentinel errors, always returned wrapped in an apperr.Error.
var (
	// ErrStopped reports an operation against a hub that has stopped.
	ErrStopped = errors.New("hub stopped")
	// ErrAlreadyRunning reports a second Run on the same hub.
	ErrAlreadyRunning = errors.New("hub already running")
)

// Defaults for [Config].
const (
	DefaultIngressBuffer = 256
	// DefaultHydrationsPerIteration caps full hydrations per loop iteration.
	DefaultHydrationsPerIteration = 4
	// DefaultReplayRing is the default replay ring capacity in envelopes.
	DefaultReplayRing = 1024
	connectQueue      = 64
)

// Config configures a [Hub]. Zero values select the documented defaults.
type Config struct {
	// SessionID is stamped on every envelope (one session per process,
	// ADR MVP-D7).
	SessionID string
	// IngressBuffer is the capacity of the ingress channel. Default
	// [DefaultIngressBuffer].
	IngressBuffer int
	// HydrationsPerIteration caps full hydrations served per select-loop
	// iteration; excess requests are deferred to later iterations, never
	// blocking ingress (design §5.3). Default [DefaultHydrationsPerIteration].
	HydrationsPerIteration int
	// ReplayRing is how many recent envelopes are retained for reconnect
	// replay. Default [DefaultReplayRing].
	ReplayRing int
	// mutHydrateThenRegister is a TEST-ONLY mutation switch (066.005-T): when
	// set, a hydrating connect snapshots in one hub step and registers in a
	// later one — the forbidden hydrate-then-register split of design §5.3.
	// It exists solely so the property test can prove it detects that defect.
	mutHydrateThenRegister bool
	// Limits are the declared fold caps (entries, bytes). Zero fields use
	// the Default* constants.
	Limits Limits
	// Clock supplies envelope timestamps and the generation timestamp.
	// Default time.Now.
	Clock func() time.Time
	// Logger receives hub diagnostics. Default: discard.
	Logger *slog.Logger
}

// Stats is a point-in-time view of the hub, read inside a hub step.
type Stats struct {
	// Head is the last assigned seq.
	Head uint64
	// Remote counts registered remote clients; TUIAttached reports the TUI.
	Remote      int
	TUIAttached bool
	// DroppedClients counts remote clients dropped by the hub for a full buffer.
	DroppedClients uint64
	// TUIOverflows counts TUI buffer overflows (each forces a re-hydrate).
	TUIOverflows uint64
	// MaxHydrationsPerIteration is the largest number of hydrations served in
	// one loop iteration so far; HydrationsDeferred counts requests that had
	// to wait for a later iteration because the cap was reached.
	MaxHydrationsPerIteration int
	HydrationsDeferred        uint64
	// Hydrations counts full hydrates served to new connects; Rehydrations
	// counts forced re-hydrates of registered clients.
	Hydrations      uint64
	DegradedSignals uint64
	Replays         uint64
	Rehydrations    uint64
}

// Hub is the single-writer event hub. See the package documentation. Every
// field below the "hub-owned" marker is touched only by the goroutine inside
// Run, so none of it needs a lock (design §4.1).
type Hub struct {
	cfg Config
	gen string

	ingress  chan agent.Event
	degraded chan string // capacity 1: adapter fault signal, out-of-band (§4.2)
	connects chan *connectReq
	statsReq chan chan Stats

	stopCh   chan struct{}
	stopOnce sync.Once
	done     chan struct{}
	running  atomic.Bool

	// hub-owned state.
	seq     uint64
	ring    []Frame // circular replay ring of the most recent sequenced frames
	ringLen int
	ringPos int // index of the next write
	fold    *Fold
	clients map[uint64]*Client
	tui     *Client
	nextID  uint64
	st      Stats
	pending []*connectReq  // connect requests waiting for hydration budget
	staleQ  []*Client      // clients awaiting a forced re-hydrate
	wake    chan time.Time // closed; selecting on it forces another iteration
	// test-only mutation state (see Config.mutHydrateThenRegister).
	mutLate  []*Client
	mutReady []*Client
}

// New constructs a Hub and captures its generation identity exactly once.
func New(cfg Config) (*Hub, error) {
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if cfg.IngressBuffer < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: IngressBuffer must not be negative")
	}
	if cfg.IngressBuffer == 0 {
		cfg.IngressBuffer = DefaultIngressBuffer
	}
	if cfg.HydrationsPerIteration < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: HydrationsPerIteration must not be negative")
	}
	if cfg.HydrationsPerIteration == 0 {
		cfg.HydrationsPerIteration = DefaultHydrationsPerIteration
	}
	if cfg.ReplayRing < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: ReplayRing must not be negative")
	}
	if cfg.ReplayRing == 0 {
		cfg.ReplayRing = DefaultReplayRing
	}
	fold, err := NewFold(cfg.Limits)
	if err != nil {
		return nil, err
	}
	gen, err := NewGen(cfg.Clock())
	if err != nil {
		return nil, err
	}
	wake := make(chan time.Time)
	close(wake)
	return &Hub{
		cfg:      cfg,
		gen:      gen,
		ingress:  make(chan agent.Event, cfg.IngressBuffer),
		degraded: make(chan string, 1),
		connects: make(chan *connectReq, connectQueue),
		statsReq: make(chan chan Stats, connectQueue),
		stopCh:   make(chan struct{}),
		done:     make(chan struct{}),
		fold:     fold,
		ring:     make([]Frame, cfg.ReplayRing),
		clients:  make(map[uint64]*Client),
		wake:     wake,
	}, nil
}

// Gen returns the process generation, captured once at construction.
func (h *Hub) Gen() string { return h.gen }

// Ingress is the buffered channel the agent adapter sends normalized events
// on. The adapter performs a BLOCKING send (design §4.2) — safe because the
// hub never blocks — and must never close the channel.
func (h *Hub) Ingress() chan<- agent.Event { return h.ingress }

// maxDegradedReason bounds the reason echoed into an error envelope.
const maxDegradedReason = 256

// Degraded is the dedicated capacity-1 channel the adapter uses to report a
// fault on the ingress path (design §4.2). The adapter must not build an
// envelope itself and must not route through ingress (the unavailable
// resource); the hub selects on this channel, builds the error envelope and
// forces every client to re-hydrate. The channel is never closed.
func (h *Hub) Degraded() chan<- string { return h.degraded }

// SignalDegraded reports a degraded condition without ever blocking. If a
// signal is already pending the new one coalesces into it: one pending
// signal already guarantees an error envelope and a full re-hydrate.
func (h *Hub) SignalDegraded(reason string) { trySend(h.degraded, reason) }

// handleDegraded is the hub-side degraded step: it builds the error envelope
// (sequenced, through the single publish step) and forces re-hydration of
// every client, because the fault means envelopes may have been lost
// upstream of the hub.
func (h *Hub) handleDegraded(reason string) {
	h.st.DegradedSignals++
	msg := "agent event path degraded: " + sanitize(reason, maxDegradedReason)
	h.fold.ApplyError(ErrCodeDegraded, msg)
	h.publish(KindError, ErrorPayload{Code: ErrCodeDegraded, Message: msg})
	for _, c := range h.clients {
		if !c.stale && !c.isClosed() {
			c.stale = true
			h.staleQ = append(h.staleQ, c)
		}
	}
}

// Done is closed when Run has returned.
func (h *Hub) Done() <-chan struct{} { return h.done }

// Stop asks Run to return and waits for it. It is idempotent and safe to call
// before Run (a later Run returns immediately).
func (h *Hub) Stop() {
	h.stopOnce.Do(func() { close(h.stopCh) })
	if h.running.Load() {
		<-h.done
	}
}

// Run is the hub goroutine: a single select loop and the sole writer of the
// canonical state. It returns after ctx is cancelled or Stop is called,
// having broadcast a server.shutdown envelope.
func (h *Hub) Run(ctx context.Context) error {
	if !h.running.CompareAndSwap(false, true) {
		return apperr.Wrap(apperr.KindAlreadyConsumed, ErrAlreadyRunning)
	}
	defer close(h.done)
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	defer timer.Stop()
	for {
		// Hydrations run at the top of each iteration under the per-iteration
		// cap, so a burst of reconnects cannot monopolise the loop (§5.3).
		immediate, retryIn := h.serviceHydrations()
		var wake, retry <-chan time.Time
		if immediate {
			wake = h.wake // closed: forces another iteration for deferred work
		}
		if retryIn > 0 {
			timer.Reset(retryIn)
			retry = timer.C
		}
		select {
		case <-ctx.Done():
			h.shutdown("context cancelled")
			return nil
		case <-h.stopCh:
			h.shutdown("stopped")
			return nil
		case ev := <-h.ingress:
			h.handleEvent(ev)
		case reason := <-h.degraded:
			h.handleDegraded(reason)
		case req := <-h.connects:
			h.intake(req)
		case <-wake:
		case <-retry:
		case reply := <-h.statsReq:
			h.sweep()
			s := h.st
			s.Head = h.seq
			s.Remote, s.TUIAttached = h.membership()
			trySend(reply, s)
		}
		h.mutFlush()
	}
}

// trySend is the hub's only send primitive for replies: select/default on a
// channel with spare capacity. It never blocks (design §4.1).
func trySend[T any](ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	default:
		return false
	}
}

// Stats reads hub counters inside a hub step.
func (h *Hub) Stats(ctx context.Context) (Stats, error) {
	reply := make(chan Stats, 1)
	select {
	case h.statsReq <- reply:
	case <-ctx.Done():
		return Stats{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		return Stats{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	}
	select {
	case s := <-reply:
		return s, nil
	case <-ctx.Done():
		return Stats{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		return Stats{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	}
}

func (h *Hub) membership() (remote int, tui bool) {
	for _, c := range h.clients {
		if c.kind == ClientTUI {
			tui = true
		} else {
			remote++
		}
	}
	return remote, tui
}

// ---- hub-goroutine-only steps below ----

// handleEvent folds one normalized event and publishes it. Events the
// adapter could not map become error envelopes; nothing is dropped.
func (h *Hub) handleEvent(ev agent.Event) {
	if _, ok := KindFor(ev.Kind); !ok {
		h.cfg.Logger.Warn("hub: unmapped agent event", "kind", string(ev.Kind), "raw_type", sanitize(ev.RawType, maxRawTypeLen))
	}
	h.fold.Apply(ev)
	kind, payload := mapEvent(ev)
	h.publish(kind, payload)
}

// publish is THE single step that assigns seq and enqueues to all clients
// (design §5.2): no other code path constructs a sequenced envelope.
func (h *Hub) publish(kind Kind, payload any) {
	h.seq++
	f, err := newFrame(h.gen, h.seq, h.cfg.Clock(), kind, h.cfg.SessionID, payload)
	if err != nil {
		h.seq--
		h.cfg.Logger.Error("hub: build envelope failed", "kind", string(kind), "err", err)
		return
	}
	h.ringAppend(f)
	h.broadcast(f)
}

// ringAppend retains f for reconnect replay, overwriting the oldest frame.
func (h *Hub) ringAppend(f Frame) {
	h.ring[h.ringPos] = f
	h.ringPos = (h.ringPos + 1) % len(h.ring)
	if h.ringLen < len(h.ring) {
		h.ringLen++
	}
}

// replayable reports whether a resume cursor can be served by replay: same
// generation, not ahead of HEAD, covered by the ring (seq+1 still retained),
// and the replay fits the client's buffer — a replay is never allowed to
// block, so a too-large one falls back to a full hydrate (§5.3).
func (h *Hub) replayable(p *Position, buffer int) bool {
	if p == nil || p.Gen != h.gen || p.Seq > h.seq {
		return false
	}
	n := h.seq - p.Seq // frames to replay
	return n <= uint64(h.ringLen) && n <= uint64(buffer)
}

// replayFrom copies the ring slice seq+1..HEAD in order. Callers must have
// checked replayable in the same hub step.
func (h *Hub) replayFrom(seq uint64) []Frame {
	n := int(h.seq - seq)
	out := make([]Frame, 0, n)
	start := (h.ringPos - n + len(h.ring)) % len(h.ring)
	for i := 0; i < n; i++ {
		out = append(out, h.ring[(start+i)%len(h.ring)])
	}
	return out
}

// broadcast enqueues f to every live client without ever blocking.
func (h *Hub) broadcast(f Frame) {
	for id, c := range h.clients {
		if c.isClosed() {
			h.unregister(id, c)
			continue
		}
		if c.stale {
			continue // will be served a fresh hydrate that already includes f
		}
		select {
		case c.send <- f:
		default:
			h.overflow(id, c)
		}
	}
}

// overflow applies the per-kind backpressure policy (design §4, §5.5).
func (h *Hub) overflow(id uint64, c *Client) {
	if c.kind == ClientTUI {
		// Never drop the permanent TUI member: mark it stale; it is skipped by
		// broadcast until re-hydrated (the hydrate supersedes everything it
		// missed, i.e. its queued backlog is dropped oldest-first).
		if !c.stale {
			c.stale = true
			h.staleQ = append(h.staleQ, c)
		}
		h.st.TUIOverflows++
		return
	}
	if c.close() {
		h.st.DroppedClients++
	}
	h.unregister(id, c)
}

// unregister is idempotent.
func (h *Hub) unregister(id uint64, c *Client) {
	delete(h.clients, id)
	if h.tui == c {
		h.tui = nil
	}
}

// sweep removes clients that closed themselves.
func (h *Hub) sweep() {
	for id, c := range h.clients {
		if c.isClosed() {
			h.unregister(id, c)
		}
	}
}

// shutdown broadcasts server.shutdown through the normal single step.
func (h *Hub) shutdown(reason string) {
	h.publish(KindServerShutdown, ShutdownPayload{Reason: reason})
}

// ---- connect ----

// ConnectOptions describes a Connect request.
type ConnectOptions struct {
	// Kind selects TUI or remote semantics.
	Kind ClientKind
	// Buffer is the client's channel capacity (default per kind).
	Buffer int
	// Resume, when set, is the client's last-applied (gen, seq) cursor.
	Resume *Position
}

type connectResp struct {
	client *Client
	err    error
}

type connectReq struct {
	opts      ConnectOptions
	reply     chan connectResp // capacity 1
	abandoned atomic.Bool
	deferred  bool // hub-owned: already counted as deferred
}

// Connect registers a client with the hub.
func (h *Hub) Connect(ctx context.Context, opts ConnectOptions) (*Client, error) {
	if opts.Kind != ClientTUI && opts.Kind != ClientRemote {
		return nil, apperr.New(apperr.KindConfig, "hub: invalid client kind")
	}
	req := &connectReq{opts: opts, reply: make(chan connectResp, 1)}
	select {
	case h.connects <- req:
	case <-ctx.Done():
		return nil, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		return nil, apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-h.stopCh:
		return nil, apperr.Wrap(apperr.KindIO, ErrStopped)
	}
	select {
	case r := <-req.reply:
		return r.client, r.err
	case <-h.done:
		select {
		case r := <-req.reply:
			return r.client, r.err
		default:
			return nil, apperr.Wrap(apperr.KindIO, ErrStopped)
		}
	case <-ctx.Done():
		// The hub never blocks, so the reply arrives promptly; wait for it so
		// a client registered in the race window is closed, not leaked.
		req.abandoned.Store(true)
		select {
		case r := <-req.reply:
			if r.client != nil {
				r.client.Close()
			}
		case <-h.done:
		}
		return nil, apperr.Wrap(apperr.KindIO, ctx.Err())
	}
}

// intake queues a connect request and batches every other request already
// waiting, so the per-iteration cap sees the whole burst.
func (h *Hub) intake(req *connectReq) {
	h.pending = append(h.pending, req)
	for {
		select {
		case r := <-h.connects:
			h.pending = append(h.pending, r)
		default:
			return
		}
	}
}

// staleRetryInterval rate-limits forced re-hydrates per client. A consumer
// that is stalled (or slower than the stream) would otherwise cost one
// snapshot copy every other event; with the limit a stalled TUI costs at most
// one hydrate per interval, and a live one is made consistent within it.
const staleRetryInterval = 25 * time.Millisecond

// serviceHydrations performs up to HydrationsPerIteration full hydrations:
// forced re-hydrates of already-registered (stale) clients first, then
// pending connects. The rest are deferred to the next iteration — never
// blocking ingress or the other select cases. It reports whether deferred
// work should run again immediately, and otherwise how long until the
// earliest rate-limited stale client may be re-hydrated (0 = none).
func (h *Hub) serviceHydrations() (immediate bool, retryIn time.Duration) {
	if len(h.pending) == 0 && len(h.staleQ) == 0 {
		return false, 0
	}
	budget, used := h.cfg.HydrationsPerIteration, 0
	now := time.Now()
	soonest := func(d time.Duration) {
		if retryIn == 0 || d < retryIn {
			retryIn = d
		}
	}

	var carryStale []*Client
	for _, c := range h.staleQ {
		if c.isClosed() || !c.stale {
			continue
		}
		if wait := staleRetryInterval - now.Sub(c.lastHydrate); wait > 0 {
			carryStale = append(carryStale, c)
			soonest(wait)
			continue
		}
		if used >= budget {
			carryStale = append(carryStale, c)
			immediate = true
			continue
		}
		used++
		c.lastHydrate = now
		if !h.rehydrate(c) && c.stale {
			carryStale = append(carryStale, c)
			soonest(staleRetryInterval)
		}
	}
	h.staleQ = carryStale

	var carry []*connectReq
	for _, req := range h.pending {
		if req.abandoned.Load() {
			trySend(req.reply, connectResp{err: apperr.New(apperr.KindIO, "hub: connect abandoned")})
			continue
		}
		replay := req.opts.Kind != 0 && h.replayable(req.opts.Resume, bufferFor(req.opts.Kind, req.opts.Buffer))
		if !replay && used >= budget {
			if !req.deferred {
				req.deferred = true
				h.st.HydrationsDeferred++
			}
			carry = append(carry, req)
			immediate = true
			continue
		}
		if !replay {
			used++ // replays are a ring copy, not a marshal: no hydration budget
		}
		h.completeConnect(req, replay)
	}
	h.pending = carry
	if used > h.st.MaxHydrationsPerIteration {
		h.st.MaxHydrationsPerIteration = used
	}
	return immediate, retryIn
}

// hydrateFrame builds a session.hydrate frame at the current high-water mark.
// The snapshot is the fold's pre-marshalled entries, so this is a copy plus a
// header (design §5.3), done inside the hub.
func (h *Hub) hydrateFrame() (Frame, bool) {
	f, err := encodeFrame(h.gen, h.seq, h.cfg.Clock(), KindSessionHydrate, h.cfg.SessionID, h.fold.Snapshot())
	if err != nil {
		h.cfg.Logger.Error("hub: build hydrate failed", "err", err)
		return Frame{}, false
	}
	return f, true
}

// rehydrate serves a stale client a fresh snapshot, reporting success. A
// full TUI buffer is drained oldest-first (everything queued is superseded by
// the snapshot); a full remote buffer drops the remote (§5.5).
func (h *Hub) rehydrate(c *Client) bool {
	f, ok := h.hydrateFrame()
	if !ok {
		return false
	}
	if trySend(c.send, f) {
		c.stale = false
		h.st.Rehydrations++
		return true
	}
	if c.kind != ClientTUI {
		h.overflow(c.id, c)
		return false
	}
drain:
	for {
		select {
		case <-c.send:
		default:
			break drain
		}
	}
	if trySend(c.send, f) {
		c.stale = false
		h.st.Rehydrations++
		return true
	}
	return false
}

// completeConnect is the ONE hub step of design §5.3: coverage check, ring
// copy or snapshot construction, (gen, seq) high-water capture and
// registration into the broadcast set happen here with no other hub work
// interleaved, so nothing published later can be missed, nothing already
// folded is delivered twice, and live envelopes cannot interleave with
// replayed ones.
func (h *Hub) completeConnect(req *connectReq, replay bool) {
	h.sweep()
	if req.opts.Kind == ClientTUI && h.tui != nil {
		trySend(req.reply, connectResp{err: apperr.New(apperr.KindAlreadyConsumed, "hub: local TUI already attached")})
		return
	}
	h.nextID++
	c := newClient(h.nextID, req.opts.Kind, req.opts.Buffer)
	if replay {
		c.mode = ModeReplay
		for _, f := range h.replayFrom(req.opts.Resume.Seq) {
			trySend(c.send, f) // capacity checked by replayable: cannot fail
		}
		h.st.Replays++
	} else {
		f, ok := h.hydrateFrame()
		if !ok {
			trySend(req.reply, connectResp{err: apperr.New(apperr.KindIO, "hub: build hydrate snapshot")})
			return
		}
		c.mode = ModeHydrate
		trySend(c.send, f) // fresh channel with capacity >= 2: cannot fail
		h.st.Hydrations++
	}
	if replay || !h.cfg.mutHydrateThenRegister {
		h.register(c)
	} else {
		h.mutLate = append(h.mutLate, c) // MUTATION: register one step too late
	}
	trySend(req.reply, connectResp{client: c})
}

func (h *Hub) register(c *Client) {
	h.clients[c.id] = c
	if c.kind == ClientTUI {
		h.tui = c
	}
}

// mutFlush registers mutation-delayed clients two iterations after their
// snapshot was taken. A no-op unless the test-only switch is set.
func (h *Hub) mutFlush() {
	if !h.cfg.mutHydrateThenRegister {
		return
	}
	for _, c := range h.mutReady {
		h.register(c)
	}
	h.mutReady, h.mutLate = h.mutLate, nil
}
