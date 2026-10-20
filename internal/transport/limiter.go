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
// socket before it is closed.
const refuseWriteTimeout = 200 * time.Millisecond

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
}

func newLimitListener(l net.Listener, limit int) *limitListener {
	if limit <= 0 {
		limit = DefaultMaxConnections
	}
	return &limitListener{Listener: l, max: int64(limit)}
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
			_ = c.SetWriteDeadline(time.Now().Add(refuseWriteTimeout))
			_, _ = c.Write([]byte(refusal503))
			_ = c.Close()
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
