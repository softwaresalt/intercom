package tui

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Default feed sizing and bounds.
const (
	// DefaultQueueSize is the capacity of the TUI's own buffered channel.
	DefaultQueueSize = 256
	// dialTimeout bounds one (re)connect to the hub.
	dialTimeout = 5 * time.Second
	// DefaultRedialRetries is the number of redial retries after the first
	// failed attempt when FeedOptions.RedialRetries is 0.
	DefaultRedialRetries = 3
	// DefaultRedialBackoff is the pause between redial attempts when
	// FeedOptions.RedialBackoff is 0.
	DefaultRedialBackoff = 25 * time.Millisecond
)

// queue is the TUI's own buffered channel with a drop-oldest overflow policy
// (design §4). It is deliberately a plain FIFO of untouched items: nothing is
// ever merged or re-parsed (no coalescing), because that would be off-hub
// interpretation of envelopes §5.2 freezes as immutable.
//
// There is exactly one producer (the pump goroutine) and one consumer (the
// deliver goroutine). Only the producer drops, so a push never blocks.
type queue[T any] struct {
	ch      chan T
	dropped atomic.Uint64
}

func newQueue[T any](size int) *queue[T] {
	if size < 1 {
		size = DefaultQueueSize
	}
	return &queue[T]{ch: make(chan T, size)}
}

// push appends v, evicting the oldest queued items while the queue is full. It
// reports whether anything was dropped. It never blocks.
func (q *queue[T]) push(v T) (dropped bool) {
	for {
		select {
		case q.ch <- v:
			return dropped
		default:
		}
		select {
		case <-q.ch:
			dropped = true
			q.dropped.Add(1)
		default: // the consumer emptied it meanwhile
		}
	}
}

// drain discards everything queued (a re-hydrate supersedes it all).
func (q *queue[T]) drain() {
	for {
		select {
		case <-q.ch:
		default:
			return
		}
	}
}

// item is one queued unit: a hub frame, or an informational resync note.
type item struct {
	frame  hub.Frame
	resync bool
}

// FeedOptions configures a [Feed].
type FeedOptions struct {
	// Dial opens a TUI connection to the hub (see [HubDialer]).
	Dial Dial
	// Sink receives the messages for the Bubble Tea program, normally
	// (*tea.Program).Send. It may block (a stalled renderer); that blocks only
	// the deliver goroutine. It must return once the program has exited, as
	// Program.Send does, so Close can finish.
	Sink func(tea.Msg)
	// QueueSize is the capacity of the TUI's own channel. 0 selects
	// [DefaultQueueSize].
	QueueSize int
	// RedialRetries is the number of redial retries after the first failed
	// attempt (in-place rehydrate failure or an unexpected disconnect).
	// 0 selects [DefaultRedialRetries]; a negative value disables retries.
	RedialRetries int
	// RedialBackoff is the pause between redial attempts. 0 selects
	// [DefaultRedialBackoff].
	RedialBackoff time.Duration
}

// FeedStats are the feed's counters.
type FeedStats struct {
	// Dropped counts envelopes evicted from the TUI's own channel.
	Dropped uint64
	// Rehydrates counts forced re-hydrates (reconnects) the feed performed.
	Rehydrates uint64
}

// Feed connects the TUI to the hub as a client of canonical state (design §4,
// §5.3). A pump goroutine drains the hub's per-client channel into the TUI's
// own drop-oldest queue, and a separate deliver goroutine hands queued items
// to the Bubble Tea program. The hub therefore never waits on the renderer:
// it writes to its own non-blocking client channel, the pump never blocks on
// the program, and only the deliver goroutine can stall.
//
// When the queue overflows the oldest items are dropped and the feed forces a
// re-hydrate by reconnecting (a fresh dial always starts with a
// session.hydrate). The consumer's [hub.Cursor] sees the gap and ignores
// frames until that hydrate arrives.
//
// A Feed also is the model's [Ingress]: Do forwards to the current
// connection, so actions keep working across re-hydrates.
type Feed struct {
	opt FeedOptions
	q   *queue[item]

	mu      sync.Mutex // guards conn, cancel, started and closed
	conn    Conn
	cancel  context.CancelFunc
	started bool
	closed  bool

	rehydrates atomic.Uint64
	rehydrateC chan rehydrateReq

	wg       sync.WaitGroup
	pumpDone chan struct{}
	pumpErr  error // written before pumpDone closes
}

// Errors returned by [Feed.Start].
var (
	// ErrFeedClosed reports a Start on a feed that was already closed.
	ErrFeedClosed = errors.New("tui: feed closed")
	// ErrFeedStarted reports a second Start on the same feed.
	ErrFeedStarted = errors.New("tui: feed already started")
)

// NewFeed builds an unstarted feed.
func NewFeed(o FeedOptions) *Feed {
	if o.RedialRetries == 0 {
		o.RedialRetries = DefaultRedialRetries
	}
	if o.RedialBackoff <= 0 {
		o.RedialBackoff = DefaultRedialBackoff
	}
	return &Feed{
		opt:        o,
		q:          newQueue[item](o.QueueSize),
		pumpDone:   make(chan struct{}),
		rehydrateC: make(chan rehydrateReq),
	}
}

// rehydrateReq asks the pump (the owner of the connection) to rehydrate.
type rehydrateReq struct {
	ctx   context.Context
	reply chan error // buffered: the pump never blocks on a gone caller
}

// Rehydrate forces a re-hydrate of the feed's connection now, in place when
// the connection supports it, otherwise by redialling. It returns
// [ErrNotConnected] when the feed is not started, closed or has no connection.
func (f *Feed) Rehydrate(ctx context.Context) error {
	f.mu.Lock()
	ok := f.started && !f.closed && f.conn != nil
	f.mu.Unlock()
	if !ok {
		return ErrNotConnected
	}
	req := rehydrateReq{ctx: ctx, reply: make(chan error, 1)}
	select {
	case f.rehydrateC <- req:
	case <-f.pumpDone:
		return ErrNotConnected
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-req.reply:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Start dials the hub once, synchronously (so a failure is reported to the
// caller), and starts the pump and deliver goroutines. They run until ctx is
// cancelled, [Feed.Close] is called, or the hub goes away.
func (f *Feed) Start(ctx context.Context) error {
	f.mu.Lock()
	switch {
	case f.closed:
		f.mu.Unlock()
		return ErrFeedClosed
	case f.started:
		f.mu.Unlock()
		return ErrFeedStarted
	}
	f.started = true
	f.mu.Unlock()

	dctx, dcancel := context.WithTimeout(ctx, dialTimeout)
	conn, err := f.opt.Dial(dctx)
	dcancel()

	f.mu.Lock()
	defer f.mu.Unlock()
	if err != nil {
		f.started = false // nothing was started: a retry is allowed
		return err
	}
	if f.closed { // Close ran while dialling
		conn.Close()
		return ErrFeedClosed
	}
	ctx, f.cancel = context.WithCancel(ctx)
	f.conn = conn
	f.wg.Add(2)
	go f.pump(ctx, conn)
	go f.deliver(ctx)
	return nil
}

// Do implements [Ingress] over the current hub connection.
func (f *Feed) Do(ctx context.Context, act hub.OperatorAction) (hub.ActionResult, error) {
	f.mu.Lock()
	c := f.conn
	f.mu.Unlock()
	if c == nil {
		return hub.ActionResult{}, ErrNotConnected
	}
	return c.Do(ctx, act)
}

// Stats returns the feed counters.
func (f *Feed) Stats() FeedStats {
	return FeedStats{Dropped: f.q.dropped.Load(), Rehydrates: f.rehydrates.Load()}
}

// Close stops both goroutines, releases the hub connection and waits for them.
// It is idempotent, and a Start after it fails with [ErrFeedClosed].
func (f *Feed) Close() {
	f.mu.Lock()
	f.closed = true
	cancel := f.cancel
	f.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	f.wg.Wait()
	f.mu.Lock()
	c := f.conn
	f.conn = nil
	f.mu.Unlock()
	if c != nil {
		c.Close()
	}
}

// pump owns the hub connection: it moves frames from the hub's channel into
// the drop-oldest queue and performs the forced re-hydrate on overflow. It
// never blocks on the program.
func (f *Feed) pump(ctx context.Context, conn Conn) {
	defer f.wg.Done()
	defer close(f.pumpDone)
	// Whatever way the pump ends (ctx cancel, hub gone, failed re-dial) it
	// releases its hub connection, so a missed Close cannot hold the hub's
	// single TUI slot. Close on a Conn is idempotent.
	defer func() {
		f.mu.Lock()
		f.conn = nil
		f.mu.Unlock()
		conn.Close()
	}()
	for {
		select {
		case fr := <-conn.Frames():
			if !f.q.push(item{frame: fr}) {
				continue
			}
			// Overflow: items were dropped, so this consumer has a gap. Only a
			// fresh snapshot can repair it.
			next, err := f.reconnect(ctx, conn)
			if err != nil {
				f.pumpErr = err
				return
			}
			conn = next
		case req := <-f.rehydrateC:
			next, err := f.reconnect(req.ctx, conn)
			req.reply <- err
			if err != nil {
				f.pumpErr = err
				return
			}
			conn = next
		case <-conn.Done():
			// Pick up what the hub queued before it went away (the shutdown
			// envelope).
			shutdown := false
		drain:
			for {
				select {
				case fr := <-conn.Frames():
					if fr.Kind() == hub.KindServerShutdown {
						shutdown = true
					}
					f.q.push(item{frame: fr})
				default:
					break drain
				}
			}
			if shutdown || ctx.Err() != nil {
				f.pumpErr = hub.ErrStopped
				return
			}
			// Dropped without a shutdown notice: recover with a fresh dial.
			conn.Close()
			next, err := f.redial(ctx)
			if err != nil {
				f.pumpErr = err
				return
			}
			conn = next
		case <-ctx.Done():
			return
		}
	}
}

// reconnect forces a re-hydrate. It prefers an in-place rehydrate of old, which
// keeps the client registered so a pending permission never loses its surface
// (the hub's empty-broadcast-set rule, §5.4, would otherwise resolve it to
// UserNotAvailable in the close-and-redial window). When old cannot rehydrate
// in place, or that fails, the old client is closed and the feed redials with
// bounded retry. Everything queued is discarded because the hydrate that heads
// the refreshed stream supersedes it.
func (f *Feed) reconnect(ctx context.Context, old Conn) (Conn, error) {
	if r, ok := old.(Rehydrater); ok {
		dctx, cancel := context.WithTimeout(ctx, dialTimeout)
		err := r.Rehydrate(dctx)
		cancel()
		if err == nil {
			f.rehydrates.Add(1)
			f.q.drain()
			f.q.push(item{resync: true})
			return old, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	old.Close()
	return f.redial(ctx)
}

// redial dials a fresh connection, retrying a failed dial up to RedialRetries
// times with RedialBackoff between attempts. On success the queue is replaced
// by a resync note; on failure the feed has no connection.
func (f *Feed) redial(ctx context.Context) (Conn, error) {
	retries := max(f.opt.RedialRetries, 0)
	var err error
	for attempt := 0; attempt <= retries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(f.opt.RedialBackoff):
			case <-ctx.Done():
				err = ctx.Err()
			}
		}
		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}
		var next Conn
		dctx, cancel := context.WithTimeout(ctx, dialTimeout)
		next, err = f.opt.Dial(dctx)
		cancel()
		if err != nil {
			continue
		}
		f.mu.Lock()
		f.conn = next
		f.mu.Unlock()
		f.rehydrates.Add(1)
		f.q.drain()
		f.q.push(item{resync: true})
		return next, nil
	}
	f.mu.Lock()
	f.conn = nil
	f.mu.Unlock()
	return nil, err
}

// deliver hands queued items to the program. It is the only goroutine that
// can be stalled by a slow renderer.
func (f *Feed) deliver(ctx context.Context) {
	defer f.wg.Done()
	send := func(it item) {
		if it.resync {
			f.opt.Sink(ResyncMsg{})
			return
		}
		f.opt.Sink(FrameMsg{Frame: it.frame})
	}
	for {
		select {
		case it := <-f.q.ch:
			send(it)
		case <-f.pumpDone:
			if ctx.Err() != nil {
				return // Close or cancel: the program is going away, say nothing
			}
			for {
				select {
				case it := <-f.q.ch:
					send(it)
				default:
					err := f.pumpErr
					if err == nil {
						err = errors.New("tui: feed ended")
					}
					f.opt.Sink(DisconnectedMsg{Err: err})
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
