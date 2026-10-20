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
	// ErrTooManyClients reports a remote Connect refused at the MaxRemote cap.
	ErrTooManyClients = errors.New("hub: too many remote clients")
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
	// Limits are the declared fold caps (entries, bytes). Zero fields use
	// the Default* constants.
	Limits Limits
	// Clock supplies envelope timestamps and the generation timestamp.
	// Default time.Now.
	Clock func() time.Time
	// Logger receives hub diagnostics. Default: discard.
	Logger *slog.Logger
	// MaxRemote caps concurrently registered remote clients (the TUI does not
	// count). 0 selects DefaultMaxRemote; values above MaxRemoteHardLimit are
	// clamped.
	MaxRemote int
	// ReplayRingBytes caps the total encoded bytes retained for replay.
	ReplayRingBytes int
	// MaxFrameBytes caps one live envelope (hydrate frames are bounded by the
	// fold caps instead). An envelope over the cap is replaced, at the same
	// seq, by a small bounded error frame. 0 derives it from the fold limits.
	MaxFrameBytes int
	// Now supplies the time used for rate limiting (not envelope timestamps;
	// see Clock). Default time.Now.
	Now func() time.Time

	// ---- TEST-ONLY MUTATION HOOK (066.005-T; accepted trade-off, see S4) ----
	// mutHydrateThenRegister, when set, makes a hydrating connect snapshot in
	// one hub step and register in a later one — the forbidden
	// hydrate-then-register split of design §5.3. It exists solely so the
	// property test can prove it detects that defect. Never set in production.
	mutHydrateThenRegister bool
	// ---- end test-only mutation hook ----
}

// Limits on remote clients.
const (
	DefaultMaxRemote   = 8
	MaxRemoteHardLimit = 1024
	// MaxClientBuffer is the hard cap on a client channel capacity.
	MaxClientBuffer = 1024
	// DefaultReplayRingBytes bounds the replay ring by total encoded bytes.
	DefaultReplayRingBytes = 4 << 20

	minFrameBytes    = 512 // room for the bounded error stub
	frameHeaderBytes = 1024
	statsQueue       = 16
)

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
	// mu guards started/stopped so that Run-start and Stop are mutually
	// consistent: Stop either observes a started Run (and waits for it) or
	// finishes the hub itself before any Run can start.
	mu      sync.Mutex
	started bool
	stopped bool // Stop ran before Run: the hub finished without ever running

	// hub-owned state.
	seq         uint64
	degradedSeq uint64  // seq of the latest degraded error envelope (0 = none)
	ring        []Frame // circular replay ring of the most recent sequenced frames
	ringLen     int
	ringPos     int // index of the next write
	ringBytes   int // total encoded bytes of the retained frames
	fold        *Fold
	clients     map[uint64]*Client
	tui         *Client
	nextID      uint64
	st          Stats
	pending     []*connectReq  // connect requests waiting for hydration budget
	staleQ      []*Client      // clients awaiting a forced re-hydrate
	wake        chan time.Time // closed; selecting on it forces another iteration
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
	if cfg.ReplayRingBytes < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: ReplayRingBytes must not be negative")
	}
	if cfg.ReplayRingBytes == 0 {
		cfg.ReplayRingBytes = DefaultReplayRingBytes
	}
	if cfg.MaxRemote < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: MaxRemote must not be negative")
	}
	if cfg.MaxRemote == 0 {
		cfg.MaxRemote = DefaultMaxRemote
	}
	if cfg.MaxRemote > MaxRemoteHardLimit {
		cfg.MaxRemote = MaxRemoteHardLimit
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	fold, err := NewFold(cfg.Limits)
	if err != nil {
		return nil, err
	}
	switch {
	case cfg.MaxFrameBytes < 0 || (cfg.MaxFrameBytes > 0 && cfg.MaxFrameBytes < minFrameBytes):
		return nil, apperr.Newf(apperr.KindConfig, "hub: MaxFrameBytes must be 0 or >= %d", minFrameBytes)
	case cfg.MaxFrameBytes == 0:
		// Room for one maximum-size entry at worst-case JSON escaping plus
		// the envelope header.
		cfg.MaxFrameBytes = maxJSONExpansion*fold.lim.MaxEntryTextBytes + entryOverhead + frameHeaderBytes
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
		statsReq: make(chan chan Stats, statsQueue),
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
// on. The adapter performs a BLOCKING send (design §4.2) — safe while the hub
// runs because the hub never blocks — and must never close the channel.
//
// Shutdown-order rule (design §3.1): the adapter MUST be stopped before the
// hub. Once Run has returned nothing drains ingress, so a raw blocking send
// can wedge forever. Producers that cannot guarantee the order should use
// [Hub.Send], which also selects on hub shutdown and a context.
func (h *Hub) Ingress() chan<- agent.Event { return h.ingress }

// Send delivers ev to the hub's ingress, blocking while ingress is saturated
// (no event is ever dropped) but never past hub shutdown or ctx: it returns
// an apperr-wrapped [ErrStopped] once the hub has stopped, or the wrapped
// context error. Send never closes ingress.
func (h *Hub) Send(ctx context.Context, ev agent.Event) error {
	select {
	case <-h.done:
		return apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-h.stopCh:
		return apperr.Wrap(apperr.KindIO, ErrStopped)
	default:
	}
	select {
	case h.ingress <- ev:
		return nil
	case <-h.done:
		return apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-h.stopCh:
		return apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-ctx.Done():
		return apperr.Wrap(apperr.KindIO, ctx.Err())
	}
}

// maxDegradedReason bounds the reason written to the log.
const maxDegradedReason = 256

// Fixed wire/fold messages: adapter-controlled text never reaches remote
// clients; the sanitized detail goes to the log only.
const (
	msgDegraded     = "agent event path degraded"
	msgUnknownEvent = "unrecognised agent event"
	msgFrameTooBig  = "envelope exceeded the maximum frame size"
)

// SignalDegraded is the single public entry for the adapter to report a fault
// on the ingress path (design §4.2). It writes to a dedicated capacity-1
// channel and never blocks: if a signal is already pending the new one
// coalesces into it, because one pending signal already guarantees an error
// envelope and a full re-hydrate. The adapter must not build an envelope
// itself and must not route through ingress (the unavailable resource).
func (h *Hub) SignalDegraded(reason string) { trySend(h.degraded, reason) }

// handleDegraded is the hub-side degraded step: it builds the error envelope
// (sequenced, through the single publish step) and forces re-hydration of
// every client, because the fault means envelopes may have been lost
// upstream of the hub. It also records the boundary: a reconnect cursor older
// than the degraded envelope must full-hydrate, never replay across it.
func (h *Hub) handleDegraded(reason string) {
	h.st.DegradedSignals++
	h.cfg.Logger.Warn("hub: agent event path degraded", "reason", sanitize(reason, maxDegradedReason))
	h.fold.ApplyError(ErrCodeDegraded, msgDegraded)
	h.publish(KindError, ErrorPayload{Code: ErrCodeDegraded, Message: msgDegraded})
	h.degradedSeq = h.seq
	for _, c := range h.clients {
		if !c.stale && !c.isClosed() {
			c.stale = true
			h.staleQ = append(h.staleQ, c)
		}
	}
}

// Done is closed when Run has returned (or when Stop finished a hub that
// never ran).
func (h *Hub) Done() <-chan struct{} { return h.done }

// Stop asks Run to return and waits for it. It is idempotent and safe to call
// before Run: the hub is then finished without running and a later Run
// returns [ErrStopped]. Stop never returns before a concurrently starting Run
// has finished.
func (h *Hub) Stop() {
	h.stopOnce.Do(func() { close(h.stopCh) })
	h.mu.Lock()
	wasStarted := h.started
	if !wasStarted {
		h.started, h.stopped = true, true
		close(h.done)
	}
	h.mu.Unlock()
	if wasStarted {
		<-h.done
	}
}

// Run is the hub goroutine: a single select loop and the sole writer of the
// canonical state. It returns after ctx is cancelled or Stop is called,
// having broadcast a server.shutdown envelope and closed every registered
// client's Done channel (pumps should drain Frames() once Done is closed to
// pick up the shutdown frame).
func (h *Hub) Run(ctx context.Context) error {
	h.mu.Lock()
	if h.started {
		stopped := h.stopped
		h.mu.Unlock()
		if stopped {
			return apperr.Wrap(apperr.KindAlreadyConsumed, ErrStopped)
		}
		return apperr.Wrap(apperr.KindAlreadyConsumed, ErrAlreadyRunning)
	}
	h.started = true
	h.mu.Unlock()
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
		// Built once, fed to both the fold and the wire. The adapter-supplied
		// type name is logged (sanitized) but never reaches clients.
		h.cfg.Logger.Warn("hub: unmapped agent event", "kind", string(ev.Kind), "raw_type", sanitize(ev.RawType, maxRawTypeLen))
		ep := unknownEventError()
		h.fold.ApplyError(ep.Code, ep.Message)
		h.publish(KindError, ep)
		return
	}
	h.fold.Apply(ev)
	kind, payload := mapEvent(ev, h.fold.lim.MaxEntryTextBytes)
	h.publish(kind, payload)
}

// publish is THE single step that assigns seq and enqueues to all clients
// (design §5.2): no other code path constructs a sequenced envelope. A frame
// over MaxFrameBytes is replaced, at the same seq, by a bounded error stub.
func (h *Hub) publish(kind Kind, payload any) {
	h.seq++
	f, err := newFrame(h.gen, h.seq, h.cfg.Clock(), kind, h.cfg.SessionID, payload)
	if err == nil && f.Len() > h.cfg.MaxFrameBytes {
		h.cfg.Logger.Warn("hub: envelope exceeds max frame size", "kind", string(kind), "bytes", f.Len(), "max", h.cfg.MaxFrameBytes)
		f, err = newFrame(h.gen, h.seq, h.cfg.Clock(), KindError, h.cfg.SessionID,
			ErrorPayload{Code: ErrCodeFrameTooLarge, Message: msgFrameTooBig})
	}
	if err != nil {
		h.seq--
		h.cfg.Logger.Error("hub: build envelope failed", "kind", string(kind), "err", err)
		return
	}
	h.ringAppend(f)
	h.broadcast(f)
}

// ringAppend retains f for reconnect replay. The ring is bounded by frame
// count (overwriting the oldest) and by total encoded bytes (evicting the
// oldest until within ReplayRingBytes), so coverage is always judged against
// the range actually retained.
func (h *Hub) ringAppend(f Frame) {
	if h.ringLen == len(h.ring) {
		h.ringBytes -= h.ring[h.ringPos].Len() // the slot being overwritten is the oldest
	} else {
		h.ringLen++
	}
	h.ring[h.ringPos] = f
	h.ringPos = (h.ringPos + 1) % len(h.ring)
	h.ringBytes += f.Len()
	for h.ringBytes > h.cfg.ReplayRingBytes && h.ringLen > 0 {
		oldest := (h.ringPos - h.ringLen + len(h.ring)) % len(h.ring)
		h.ringBytes -= h.ring[oldest].Len()
		h.ring[oldest] = Frame{}
		h.ringLen--
	}
}

// replayable reports whether a resume cursor can be served by replay: same
// generation, not ahead of HEAD, not older than the latest degraded envelope
// (envelopes may have been lost across it, so the client must full-hydrate),
// covered by the retained ring (seq+1 still held), and the replay leaves at
// least one free slot in the client's buffer — a replay is never allowed to
// block or exactly fill the buffer, so a too-large one falls back to a full
// hydrate (§5.3).
func (h *Hub) replayable(p *Position, buffer int) bool {
	if p == nil || p.Gen != h.gen || p.Seq > h.seq || p.Seq < h.degradedSeq {
		return false
	}
	n := h.seq - p.Seq // frames to replay
	return n <= uint64(h.ringLen) && n < uint64(buffer)
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

// shutdown broadcasts server.shutdown through the normal single step, then
// closes every registered client's Done (via its sync.Once) — including stale
// clients that never received the shutdown frame — so write pumps can exit
// instead of waiting on a hub that will never send again. A pump should drain
// Frames() after Done to pick up the queued shutdown frame.
func (h *Hub) shutdown(reason string) {
	h.publish(KindServerShutdown, ShutdownPayload{Reason: reason})
	for id, c := range h.clients {
		c.close()
		h.unregister(id, c)
	}
	for _, c := range h.mutLate {
		c.close()
	}
	for _, c := range h.mutReady {
		c.close()
	}
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
		// Never wait on the hub here: it may not be running. Mark the request
		// abandoned — the hub discards it (or closes the client it registers
		// in the race window) when it processes it — and close a client whose
		// reply already arrived. The hub stores the reply before it re-checks
		// the flag, so one side always sees the other.
		req.abandoned.Store(true)
		select {
		case r := <-req.reply:
			if r.client != nil {
				r.client.Close()
			}
		default:
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
	now := h.cfg.Now()
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
		replay := h.replayable(req.opts.Resume, bufferFor(req.opts.Kind, req.opts.Buffer))
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
	if req.abandoned.Load() {
		trySend(req.reply, connectResp{err: apperr.New(apperr.KindIO, "hub: connect abandoned")})
		return
	}
	h.sweep()
	if req.opts.Kind == ClientTUI && h.tui != nil {
		trySend(req.reply, connectResp{err: apperr.New(apperr.KindAlreadyConsumed, "hub: local TUI already attached")})
		return
	}
	if req.opts.Kind == ClientRemote {
		if remote, _ := h.membership(); remote >= h.cfg.MaxRemote {
			trySend(req.reply, connectResp{err: apperr.Wrap(apperr.KindPolicy, ErrTooManyClients)})
			return
		}
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
	if req.abandoned.Load() { // requester gave up while we were completing: do not keep the client
		c.close()
		h.unregister(c.id, c)
	}
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
