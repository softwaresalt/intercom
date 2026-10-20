package hub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// Gate is the admission surface the shutdown executor closes at step 0
// (design §3.1, plan 4.6). The transport implements it: revoke every
// credential so no upgrade or action authenticates, and stop accepting
// connections. Both calls must be idempotent.
type Gate interface {
	// RevokeCredentials invalidates all credentials (ADR MVP-2: revoked at
	// shutdown). It must not signal connections' done channels: write pumps
	// close their sockets after the shutdown broadcast and drain (ADR MVP-3).
	RevokeCredentials(ctx context.Context) error
	// CloseListener stops accepting new connections.
	CloseListener(ctx context.Context) error
}

// ShutdownTimeouts are the per-step budgets. A zero field selects its default.
type ShutdownTimeouts struct {
	// Permissions bounds step 0's wait for the handlers to take their
	// UserNotAvailable (default 5s).
	Permissions time.Duration
	// Gate bounds each Gate call (default 5s).
	Gate time.Duration
	// Announce bounds the server.shutdown hub step (default 2s).
	Announce time.Duration
	// Abort bounds Agent.Abort (default 5s).
	Abort time.Duration
	// Idle bounds the wait for a positive idle signal (default 5s).
	Idle time.Duration
	// Stop bounds Agent.Stop, after which the adapter escalates to ForceStop
	// (default 10s).
	Stop time.Duration
	// Drain bounds the write-pump drain before sockets are aborted (design
	// §3.1 step 5: 1-2 s; default 2s).
	Drain time.Duration
}

func (t ShutdownTimeouts) valid() bool {
	for _, d := range []time.Duration{t.Permissions, t.Gate, t.Announce, t.Abort, t.Idle, t.Stop, t.Drain} {
		if d < 0 {
			return false
		}
	}
	return true
}

func (t ShutdownTimeouts) withDefaults() ShutdownTimeouts {
	def := func(d *time.Duration, v time.Duration) {
		if *d == 0 {
			*d = v
		}
	}
	def(&t.Permissions, 5*time.Second)
	def(&t.Gate, 5*time.Second)
	def(&t.Announce, 2*time.Second)
	def(&t.Abort, 5*time.Second)
	def(&t.Idle, 5*time.Second)
	def(&t.Stop, 10*time.Second)
	def(&t.Drain, 2*time.Second)
	return t
}

// drainPollInterval is how often step 5 re-checks the client queues.
const drainPollInterval = 5 * time.Millisecond

// ShutdownConfig configures a [Shutdown].
type ShutdownConfig struct {
	Hub   *Hub
	Agent agent.Agent
	Gate  Gate
	// Reason is carried in server.shutdown (default "shutdown").
	Reason   string
	Timeouts ShutdownTimeouts
	Logger   *slog.Logger
}

// Shutdown is the design §3.1 shutdown executor. The sequence runs on a
// dedicated goroutine, never on the hub goroutine: steps 0, 2 and 3 block on
// results that can only arrive through the hub, so running them on the hub
// would stop it draining ingress and deadlock the shutdown. The hub keeps
// draining ingress through steps 0-5 and stops only at step 6.
type Shutdown struct {
	cfg  ShutdownConfig
	once sync.Once
	done chan struct{}
	err  error
}

// NewShutdown validates cfg and applies defaults.
func NewShutdown(cfg ShutdownConfig) (*Shutdown, error) {
	switch {
	case cfg.Hub == nil:
		return nil, apperr.New(apperr.KindConfig, "hub: shutdown needs a hub")
	case cfg.Agent == nil:
		return nil, apperr.New(apperr.KindConfig, "hub: shutdown needs an agent")
	case cfg.Gate == nil:
		return nil, apperr.New(apperr.KindConfig, "hub: shutdown needs a gate")
	case !cfg.Timeouts.valid():
		return nil, apperr.New(apperr.KindConfig, "hub: shutdown timeouts must not be negative")
	}
	if cfg.Reason == "" {
		cfg.Reason = "shutdown"
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Timeouts = cfg.Timeouts.withDefaults()
	return &Shutdown{cfg: cfg, done: make(chan struct{})}, nil
}

// Start launches the shutdown sequence on its own goroutine. It is idempotent:
// only the first call runs the sequence. ctx is the shutdown's overall budget:
// cancelling it hurries the agent and drain steps (the adapter then force-stops
// its runtime), but steps 0 and 1, the gate and step 6 always run.
func (s *Shutdown) Start(ctx context.Context) {
	s.once.Do(func() { go s.run(ctx) })
}

// Done is closed when the sequence has finished, the hub being stopped.
func (s *Shutdown) Done() <-chan struct{} { return s.done }

// Err reports the aggregated step errors; valid after Done. Every error is
// also logged as it happens.
func (s *Shutdown) Err() error { return s.err }

func (s *Shutdown) run(ctx context.Context) {
	var errs []error
	t, h := s.cfg.Timeouts, s.cfg.Hub
	// step runs fn under its own budget, logs a failure and records it; a
	// failed step never skips the later ones.
	step := func(name string, parent context.Context, budget time.Duration, fn func(context.Context) error) error {
		err := boundedStep(parent, budget, fn)
		if err == nil {
			return nil
		}
		err = fmt.Errorf("shutdown %s: %w", name, err)
		s.cfg.Logger.Error("shutdown step failed", "step", name, "err", err)
		errs = append(errs, err)
		return err
	}
	// Steps whose effect must happen even if the overall budget is spent keep
	// their own deadline but not the parent's cancellation.
	always := context.WithoutCancel(ctx)

	// Step 0: fail every pending permission through the hub (and close
	// admission, atomically with it), then revoke credentials and close the
	// listener so no client is admitted after this step.
	_ = step("0 fail pending permissions", always, t.Permissions, h.BeginShutdown)
	_ = step("0 revoke credentials", always, t.Gate, s.cfg.Gate.RevokeCredentials)
	_ = step("0 close listener", always, t.Gate, s.cfg.Gate.CloseListener)
	// Step 1: tell the operators why, before any teardown.
	_ = step("1 broadcast server.shutdown", always, t.Announce, func(c context.Context) error {
		return h.AnnounceShutdown(c, s.cfg.Reason)
	})
	// Steps 2-4: the agent. See agentStage for how the design's ordering maps
	// onto the adapter's Stop.
	s.agentStage(ctx, step)
	// Step 5: bounded drain of the write pumps, then abort the sockets.
	_ = step("5 drain clients", ctx, t.Drain+time.Second, func(c context.Context) error {
		return h.drainAndCloseClients(c, t.Drain)
	})
	// Step 6: the hub outlives everything above.
	h.Stop()
	if err := ctx.Err(); err != nil {
		errs = append(errs, fmt.Errorf("shutdown interrupted: %w", err))
	}
	s.err = errors.Join(errs...)
	close(s.done)
}

// agentStage is design §3.1 steps 2-4.
//
// Step 3 first: Agent.Abort the in-flight turn, then wait for a POSITIVE idle
// signal from the hub (a real idle event with ingress drained, never a stale
// flag), bounded by a deadline. Then Agent.Stop, which performs steps 2 and 4
// in the adapter: it closes the callback latch and waits (bounded) for
// in-flight callbacks to drain (step 2), shuts the runtime down with
// Client.Stop under the deadline and escalates to ForceStop when that fails or
// expires (step 4), aggregating its own errors. Step 2's quiesce runs inside
// Stop, after the idle wait, rather than before the abort as the design lists
// it: the adapter registers its handler at session creation (no separate
// unsubscribe handle), and closing the latch first would discard the very
// idle event step 3 waits for (agent.Agent.Stop documents this order).
//
// A failed or timed-out step is logged and aggregated, and never skips the
// stop: the runtime and its CLI child must be torn down regardless.
func (s *Shutdown) agentStage(ctx context.Context, step func(string, context.Context, time.Duration, func(context.Context) error) error) {
	t := s.cfg.Timeouts
	var noSession atomic.Bool // Abort reported there is no session to abort
	err := step("3 abort turn", ctx, t.Abort, func(c context.Context) error {
		err := s.cfg.Agent.Abort(c)
		if errors.Is(err, agent.ErrNotRunning) || errors.Is(err, agent.ErrStopped) {
			noSession.Store(true) // nothing to abort is not a failure
			return nil
		}
		return err
	})
	// With nothing to abort there is no idle to wait for, and after a failed
	// abort the wait could only burn its deadline.
	if err == nil && !noSession.Load() {
		_ = step("3 await idle", ctx, t.Idle, s.cfg.Hub.WaitIdle)
	}
	_ = step("2+4 stop agent", ctx, t.Stop, s.cfg.Agent.Stop)
}

// boundedStep runs fn under a budget-bound child of parent. fn may ignore its
// context (a transport or SDK call that blocks): the step still returns at the
// deadline, leaving that call to finish on its own goroutine.
func boundedStep(parent context.Context, budget time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, budget)
	defer cancel()
	res := make(chan error, 1)
	go func() { trySend(res, fn(ctx)) }() // capacity 1, sole sender: never drops
	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		select { // a result that is already ready beats the deadline
		case err := <-res:
			return err
		default:
			return ctx.Err()
		}
	}
}

// ---- hub-side shutdown steps ----

// ctlOp enumerates the hub-control messages the shutdown executor sends.
type ctlOp uint8

const (
	ctlBegin ctlOp = iota + 1
	ctlAnnounce
	ctlClients
	ctlIdle
)

// WaitIdle blocks until the agent session is positively idle (design §3.1
// step 3): the hub has observed an idle event (or no turn ever started) AND its
// ingress is drained. It never reads the fold's idle flag while events are
// still queued, because that flag is stale until they are folded. It returns
// the context's error at its deadline, or ErrStopped if the hub stops first.
func (h *Hub) WaitIdle(ctx context.Context) error {
	_, err := h.ctlCall(ctx, ctlIdle, "")
	return err
}

// releaseIdleWaiters answers the WaitIdle callers whose condition now holds.
// It runs at the end of every hub iteration, after ingress had a chance to be
// served, so "ingress empty" is judged against the state just folded.
func (h *Hub) releaseIdleWaiters() {
	if len(h.idleWaiters) == 0 || !h.agentIdle || len(h.ingress) > 0 {
		return
	}
	for _, w := range h.idleWaiters {
		trySend(w.reply, ctlResp{})
	}
	h.idleWaiters = nil
}

type ctlResp struct {
	err     error
	clients []*Client
}

type ctlReq struct {
	op     ctlOp
	reason string
	reply  chan ctlResp // capacity 1: the hub never blocks on it
}

// ctlCall delivers a control message to the hub and waits for its reply.
func (h *Hub) ctlCall(ctx context.Context, op ctlOp, reason string) (ctlResp, error) {
	req := &ctlReq{op: op, reason: reason, reply: make(chan ctlResp, 1)}
	select {
	case h.ctl <- req:
	case <-ctx.Done():
		return ctlResp{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		return ctlResp{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	case <-h.stopCh:
		return ctlResp{}, apperr.Wrap(apperr.KindIO, ErrStopped)
	}
	select {
	case r := <-req.reply:
		return r, r.err
	case <-ctx.Done():
		return ctlResp{}, apperr.Wrap(apperr.KindIO, ctx.Err())
	case <-h.done:
		select {
		case r := <-req.reply:
			return r, r.err
		default:
			return ctlResp{}, apperr.Wrap(apperr.KindIO, ErrStopped)
		}
	}
}

// BeginShutdown is design §3.1 step 0 on the hub side, one hub step: admission
// closes (no client is admitted after it, no new ask can park a handler, and
// privileged actions are refused without closing connections), and every
// unresolved permission is resolved to UserNotAvailable through the normal
// hand-off. It returns once every registry entry has been retired (the
// handlers took their decision and permission.resolved was broadcast with the
// shutdown reason, or the entry was aborted), or when ctx ends. The hub keeps
// serving ingress while it waits.
func (h *Hub) BeginShutdown(ctx context.Context) error {
	_, err := h.ctlCall(ctx, ctlBegin, "")
	return err
}

// AnnounceShutdown is step 1: broadcast server.shutdown through the normal
// single publish step. The hub's own stop later does not repeat it.
func (h *Hub) AnnounceShutdown(ctx context.Context, reason string) error {
	_, err := h.ctlCall(ctx, ctlAnnounce, reason)
	return err
}

func (h *Hub) handleCtl(r *ctlReq) {
	switch r.op {
	case ctlBegin:
		h.beginShutdown(r)
	case ctlAnnounce:
		h.announceShutdown(r.reason)
		trySend(r.reply, ctlResp{})
	case ctlClients:
		h.sweep()
		out := make([]*Client, 0, len(h.clients))
		for _, c := range h.clients {
			out = append(out, c)
		}
		trySend(r.reply, ctlResp{clients: out})
	case ctlIdle:
		h.idleWaiters = append(h.idleWaiters, r)
	default:
		trySend(r.reply, ctlResp{err: apperr.New(apperr.KindConfig, "hub: unknown control message")})
	}
}

// beginShutdown closes admission and fails every pending permission. The reply
// is held until the registry has drained.
func (h *Hub) beginShutdown(r *ctlReq) {
	h.closing = true
	for _, e := range h.perm.unresolved() {
		h.decide(e, ActorHub, agent.PermissionDecision{Kind: agent.DecisionUnavailable}, permReasonShutdown)
	}
	if h.perm.len() == 0 {
		trySend(r.reply, ctlResp{})
		return
	}
	h.drainWaiters = append(h.drainWaiters, r)
}

// notifyDrained releases the step-0 waiters once the registry is empty.
func (h *Hub) notifyDrained() {
	if h.perm.len() != 0 {
		return
	}
	for _, w := range h.drainWaiters {
		trySend(w.reply, ctlResp{})
	}
	h.drainWaiters = nil
}

// announceShutdown publishes server.shutdown once.
func (h *Hub) announceShutdown(reason string) {
	if h.announced {
		return
	}
	h.announced = true
	h.publish(KindServerShutdown, ShutdownPayload{Reason: reason})
}

// drainAndCloseClients is design §3.1 step 5. It waits, bounded by timeout,
// for every client's queue to empty (its write pump has taken every frame,
// including server.shutdown), then closes every client's done through its
// sync.Once so the pumps abort their sockets. A client that never drains is
// closed at the deadline: one stalled client must not hold shutdown (§5.5).
func (h *Hub) drainAndCloseClients(ctx context.Context, timeout time.Duration) error {
	resp, err := h.ctlCall(ctx, ctlClients, "")
	if err != nil {
		return err
	}
	defer func() {
		for _, c := range resp.clients {
			c.close()
		}
	}()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(drainPollInterval)
	defer tick.Stop()
	for {
		if queuesEmpty(resp.clients) {
			return nil
		}
		select {
		case <-tick.C:
		case <-deadline.C:
			return nil // bounded: the stalled client is aborted by the deferred close
		case <-ctx.Done():
			return nil // hurried: abort now
		}
	}
}

func queuesEmpty(cs []*Client) bool {
	for _, c := range cs {
		if len(c.send) > 0 && !c.isClosed() {
			return false
		}
	}
	return true
}
