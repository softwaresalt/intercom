package hub

import (
	"sync"
	"time"
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
