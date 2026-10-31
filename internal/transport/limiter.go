package transport

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultMaxConnections is the cap used when [ServerConfig.MaxConnections]
// is zero; it matches the [remote].max_connections default.
const DefaultMaxConnections = 8

// refuseWriteTimeout bounds the best-effort 503 written to an over-cap
// socket before it is closed. The write happens on the refusal worker, never
// on the accept loop; it is kept short so one stalled peer delays the next
// queued refusal only briefly.
const refuseWriteTimeout = 50 * time.Millisecond

const refusal503 = "HTTP/1.1 503 Service Unavailable\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"

// limitListener enforces [remote].max_connections at accept time, BEFORE any
// byte is read or any credential work happens (ADR MVP-2: connection cap
// first, 503 over cap). A slot is therefore held by every socket, including
// ones that never complete a handshake, and is released exactly once when the
// socket closes by ANY path (handshake timeout, rejection, normal close,
// hijacked WebSocket close, shutdown): the wrapped conn's Close is the only
// release point.
type limitListener struct {
	net.Listener
	max    int64
	active atomic.Int64

	// onReject, when non-nil, is called once per over-cap refusal.
	onReject func()
	// refuseQ holds over-cap sockets awaiting their 503; the refusal worker
	// drains it. done is closed by Close and workerDone when the worker exits.
	refuseQ    chan net.Conn
	workerDone chan struct{}
	done       chan struct{}
	closeOnce  sync.Once
}

// overCapRefuseQueueSize bounds the over-cap sockets waiting for their 503.
const overCapRefuseQueueSize = 64

// rejectOverCap is the OnReject reason for a connection refused at the cap.
const rejectOverCap = "connection_cap_exceeded"

func newLimitListener(l net.Listener, limit int, onReject func()) *limitListener {
	if limit <= 0 {
		limit = DefaultMaxConnections
	}
	ll := &limitListener{
		Listener: l, max: int64(limit), onReject: onReject,
		refuseQ:    make(chan net.Conn, overCapRefuseQueueSize),
		workerDone: make(chan struct{}),
		done:       make(chan struct{}),
	}
	go ll.refuseWorker()
	return ll
}

// refuseWorker writes the 503 to queued over-cap sockets in FIFO order and
// closes them. When the listener closes it closes whatever is still queued
// (without writing) and exits.
func (l *limitListener) refuseWorker() {
	defer close(l.workerDone)
	for {
		select {
		case c := <-l.refuseQ:
			_ = c.SetWriteDeadline(time.Now().Add(refuseWriteTimeout))
			_, _ = c.Write([]byte(refusal503))
			_ = c.Close()
		case <-l.done:
			for {
				select {
				case c := <-l.refuseQ:
					_ = c.Close()
				default:
					return
				}
			}
		}
	}
}

// Close stops accepting and stops the refusal worker. It is idempotent.
func (l *limitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() { close(l.done) })
	return err
}

// refuse counts an over-cap socket and queues it for its 503 without ever
// blocking: a full queue (or a closed listener) closes the socket at once.
func (l *limitListener) refuse(c net.Conn) {
	if l.onReject != nil {
		l.onReject()
	}
	select {
	case <-l.done:
		_ = c.Close()
		return
	default:
	}
	select {
	case l.refuseQ <- c:
	default:
		_ = c.Close()
	}
}

// Accept returns only admitted connections; over-cap sockets are refused with
// a 503 and closed, and Accept keeps accepting.
func (l *limitListener) Accept() (net.Conn, error) {
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if l.active.Add(1) > l.max {
			l.active.Add(-1)
			l.refuse(c)
			continue
		}
		return &slotConn{Conn: c, l: l}, nil
	}
}

// slotConn releases its slot exactly once on Close.
type slotConn struct {
	net.Conn
	l    *limitListener
	once sync.Once
}

func (c *slotConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { c.l.active.Add(-1) })
	return err
}
