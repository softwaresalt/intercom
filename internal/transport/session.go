package transport

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Timing defaults for a session's pumps.
const (
	// DefaultPingInterval is the keep-alive ping period (a case in the write
	// pump's select; there is no other pinger).
	DefaultPingInterval = 30 * time.Second
	// DefaultWriteTimeout bounds one socket write or ping.
	DefaultWriteTimeout = 10 * time.Second
	// DefaultCloseGrace bounds the close handshake before the abort.
	DefaultCloseGrace = 2 * time.Second

	// closeReasonEnded is the neutral reason sent when a credential expires,
	// is revoked, or fails re-validation: it reveals nothing about why.
	closeReasonEnded = "session ended"
	// closeReasonGoingAway is sent on drop or shutdown.
	closeReasonGoingAway = "closing"
)

// closeRequest asks the write pump to close the socket (the read pump and
// any other goroutine must not touch the socket's write side: design §5.5).
type closeRequest struct {
	status websocket.StatusCode
	reason string
}

// session is one authenticated, registered WebSocket connection. Ownership:
//   - writePump (this goroutine's caller) is the ONLY goroutine that calls
//     sock.Write, sock.Ping, sock.Close and, after done, sock.CloseNow;
//   - readPump only reads and signals through client.Close or closeReq.
type session struct {
	h      *Handler
	sock   *websocket.Conn
	client *hub.Client
	grant  *Grant

	closeReq chan closeRequest // capacity 1: first request wins
}

func newSession(h *Handler, sock *websocket.Conn, client *hub.Client, grant *Grant) *session {
	return &session{h: h, sock: sock, client: client, grant: grant, closeReq: make(chan closeRequest, 1)}
}

// requestClose asks the write pump to close the connection with status and a
// neutral reason. It never blocks and never touches the socket.
func (s *session) requestClose(status websocket.StatusCode, reason string) {
	select {
	case s.closeReq <- closeRequest{status: status, reason: reason}:
	default: // a close is already pending
	}
}

// run serves the session until it ends. The calling goroutine becomes the
// write pump; the read pump runs in one more goroutine that is always joined
// before run returns, so nothing outlives the session.
func (s *session) run() {
	s.h.sessions.Add(1)
	defer s.h.sessions.Add(-1)

	ctx, cancel := context.WithCancel(s.h.ctx)
	defer cancel()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		s.readPump(ctx)
	}()
	s.writePump(ctx)
	cancel()
	s.client.Close()
	<-readDone
}

// writePump is the sole writer of the socket. Frames, keep-alive pings, the
// credential-expiry deadline, revocation, close requests, drop (done) and
// shutdown are all cases of ONE select, so no second goroutine can interleave
// a frame.
func (s *session) writePump(ctx context.Context) {
	ping := time.NewTicker(s.h.pingEvery)
	defer ping.Stop()
	expiry := time.NewTimer(s.grant.ExpiresAt().Sub(s.h.now()))
	defer expiry.Stop()

	for {
		// done has priority over queued frames: a dropped client's backlog
		// is stale and is never flushed.
		select {
		case <-s.client.Done():
			s.closeSocket(websocket.StatusGoingAway, closeReasonGoingAway)
			return
		default:
		}
		select {
		case f := <-s.client.Frames():
			if err := s.writeFrame(ctx, f.Bytes()); err != nil {
				s.fail()
				return
			}
		case <-ping.C:
			if err := s.pingPeer(ctx); err != nil {
				s.fail()
				return
			}
		case <-expiry.C:
			s.closeSocket(websocket.StatusPolicyViolation, closeReasonEnded)
			return
		case <-s.grant.Revoked():
			s.closeSocket(websocket.StatusPolicyViolation, closeReasonEnded)
			return
		case rq := <-s.closeReq:
			s.closeSocket(rq.status, rq.reason)
			return
		case <-s.client.Done():
			s.closeSocket(websocket.StatusGoingAway, closeReasonGoingAway)
			return
		case <-ctx.Done():
			s.closeSocket(websocket.StatusGoingAway, closeReasonGoingAway)
			return
		}
	}
}

func (s *session) writeFrame(ctx context.Context, data []byte) error {
	wctx, cancel := context.WithTimeout(ctx, s.h.writeTimeout)
	defer cancel()
	return s.sock.Write(wctx, websocket.MessageText, data)
}

func (s *session) pingPeer(ctx context.Context) error {
	pctx, cancel := context.WithTimeout(ctx, s.h.writeTimeout)
	defer cancel()
	s.h.pings.Add(1)
	return s.sock.Ping(pctx)
}

// fail ends the session after a write/ping error: the socket is already
// unusable, so it is dropped (done closed) and aborted by the pump.
func (s *session) fail() {
	s.client.Close()
	_ = s.sock.CloseNow()
}

// closeSocket sends the close frame and tears the socket down. done is closed
// FIRST, which is what licenses the bounded abort timer to call CloseNow
// from another goroutine if the peer never completes the close handshake.
func (s *session) closeSocket(status websocket.StatusCode, reason string) {
	s.client.Close()
	abort := time.AfterFunc(s.h.closeGrace, func() { _ = s.sock.CloseNow() })
	defer abort.Stop()
	_ = s.sock.Close(status, reason)
}

// readPump reads (and for now discards) inbound messages until the socket
// ends, then drops the client so the write pump closes down. It never writes
// or closes the socket.
func (s *session) readPump(ctx context.Context) {
	for {
		if _, _, err := s.sock.Read(ctx); err != nil {
			s.client.Close()
			return
		}
	}
}
