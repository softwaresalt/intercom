package hub

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// ClientKind distinguishes the permanent local TUI member of the broadcast
// set from remote (WebSocket) clients (design §5.4, §5.5).
type ClientKind int

const (
	// ClientTUI is the local Bubble Tea surface. It is a permanent member of
	// the broadcast set: it is never dropped for backpressure, only
	// overflowed and re-hydrated (design §4, §5.4). At most one may attach.
	ClientTUI ClientKind = iota + 1
	// ClientRemote is a remote (WebSocket) surface. A full buffer drops it
	// (design §5.5).
	ClientRemote
)

// ConnectMode reports how a client's stream was started.
type ConnectMode int

const (
	// ModeNone is the zero value; a connected client never has it.
	ModeNone ConnectMode = iota
	// ModeHydrate: the first frame is a session.hydrate snapshot.
	ModeHydrate
	// ModeReplay: the first frames are the replayed envelopes seq+1..HEAD.
	ModeReplay
)

// Default per-client buffer sizes.
const (
	DefaultRemoteBuffer = 64
	DefaultTUIBuffer    = 256
	// minClientBuffer leaves room for a hydrate frame plus one live frame.
	minClientBuffer = 2
)

// Client is one registered consumer of the hub's broadcast. The hub writes to
// the buffered send channel and never closes it; the consumer's pump alone
// owns the transport. A client is dropped by closing its done channel exactly
// once (sync.Once), so a hub drop racing a consumer-side failure can never
// double-close (design §5.5).
type Client struct {
	id   uint64
	kind ClientKind
	mode ConnectMode
	send chan Frame
	done chan struct{}
	once sync.Once

	// stale is owned by the hub goroutine: the client missed envelopes (or the
	// hub forced a re-hydrate) and is skipped by broadcast until it has been
	// served a fresh session.hydrate.
	stale bool
	// lastHydrate (hub-owned) rate-limits forced re-hydrates.
	lastHydrate time.Time
	// notify is the hub's capacity-1 membership wake channel, set by the hub
	// before the client is handed out. close signals it (non-blocking) so the
	// hub re-evaluates the broadcast set promptly when a surface goes away
	// (067.008-T) instead of at its next unrelated step.
	notify chan<- struct{}
	// role and hub are set by the hub, once, before the client is handed out
	// (so reads from any goroutine happen after the write). role is the
	// authorization role (ADR MVP-3); it can never be changed by a message.
	role Role
	hub  *Hub

	// inflight counts frames the hub has queued that the consumer has not yet
	// finished with (queued, or dequeued but not acknowledged). It is
	// incremented before a send and decremented on a failed send, a hub drain
	// or a consumer acknowledgement.
	inflight atomic.Int64
	// terminal (hub-owned) is set once the terminal server.shutdown frame has
	// been queued: the client is skipped by hydration, staleness marking and
	// later broadcasts so that frame is never drained or replaced.
	terminal bool
}

// sendFrame queues f without blocking and reports whether it was queued. The
// in-flight count is raised before the send so a consumer that dequeues and
// acknowledges immediately can never drive it negative.
func (c *Client) sendFrame(f Frame) bool {
	c.inflight.Add(1)
	select {
	case c.send <- f:
		return true
	default:
		c.inflight.Add(-1)
		return false
	}
}

// drainOneFrame discards the oldest queued frame, reporting whether one existed.
func (c *Client) drainOneFrame() bool {
	select {
	case <-c.send:
		c.inflight.Add(-1)
		return true
	default:
		return false
	}
}

// drainAllFrames discards every queued frame.
func (c *Client) drainAllFrames() {
	for c.drainOneFrame() {
	}
}

// ErrClientClosed reports an action attempted through a client that has been
// closed or dropped.
var ErrClientClosed = errors.New("hub: client closed")

// Role reports the client's authorization role.
func (c *Client) Role() Role { return c.role }

// Do submits one operator action to the hub and returns its verdict. The
// identity the action is performed as comes from this Client's own kind and
// role, fixed when the hub registered it; nothing in act can change it. A
// refusal is an ActionResult with a Reason, delivered to this caller only; the
// error is non-nil only when the hub could not be reached (context ended, hub
// stopped, client closed).
func (c *Client) Do(ctx context.Context, act OperatorAction) (ActionResult, error) {
	if c.isClosed() {
		return ActionResult{}, apperr.Wrap(apperr.KindPolicy, ErrClientClosed)
	}
	p := principal{surface: surfaceRemote, role: c.role}
	if c.kind == ClientTUI {
		p = principal{surface: surfaceTUI, role: RoleApprover}
	}
	return c.hub.submitFrom(ctx, c, p, act)
}

func newClient(id uint64, kind ClientKind, buffer int) *Client {
	return &Client{id: id, kind: kind, send: make(chan Frame, bufferFor(kind, buffer)), done: make(chan struct{})}
}

// bufferFor resolves a requested client buffer to its effective capacity.
func bufferFor(kind ClientKind, buffer int) int {
	if buffer <= 0 {
		buffer = DefaultRemoteBuffer
		if kind == ClientTUI {
			buffer = DefaultTUIBuffer
		}
	}
	if buffer < minClientBuffer {
		buffer = minClientBuffer
	}
	if buffer > MaxClientBuffer {
		buffer = MaxClientBuffer
	}
	return buffer
}

// Frames is the receive side of the client's buffered envelope channel. It is
// never closed; consumers select on Done as well.
//
// Consumer contract: feed every received frame to a [Cursor] and apply it only
// on ActionApply. A forced re-hydrate (TUI overflow, degraded signal, oversize
// frame) drops the queued backlog oldest-first, so a consumer may observe a
// gap before the session.hydrate arrives; the Cursor reports ActionRehydrate
// and the consumer discards frames until that hydrate is applied.
func (c *Client) Frames() <-chan Frame { return c.send }

// Done is closed when the client is dropped or closed.
func (c *Client) Done() <-chan struct{} { return c.done }

// Kind reports the client kind.
func (c *Client) Kind() ClientKind { return c.kind }

// Mode reports how the stream was started (hydrate or replay).
func (c *Client) Mode() ConnectMode { return c.mode }

// Close unregisters the client. It is idempotent and safe from any goroutine;
// the hub removes a closed client at its next step.
func (c *Client) Close() { c.close() }

// close reports whether this call performed the close.
func (c *Client) close() bool {
	closed := false
	c.once.Do(func() {
		close(c.done)
		closed = true
	})
	if closed && c.notify != nil {
		trySend(c.notify, struct{}{})
	}
	return closed
}

func (c *Client) isClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}
