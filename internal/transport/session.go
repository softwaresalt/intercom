package transport

import (
	"context"
	"errors"
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
	// DefaultIdleTimeout closes a connection with no inbound activity.
	DefaultIdleTimeout = 90 * time.Second
	// DefaultActionTimeout bounds one hub action round trip.
	DefaultActionTimeout = 10 * time.Second
	// DefaultMaxMessageBytes is the inbound message size limit. Actions are
	// small JSON objects; this is generous for a prompt.
	DefaultMaxMessageBytes = 16 << 10

	// closeReasonEnded is the neutral reason sent when a credential expires,
	// is revoked, or fails re-validation: it reveals nothing about why.
	closeReasonEnded = "session ended"
	// closeReasonGoingAway is sent on drop or shutdown.
	closeReasonGoingAway   = "closing"
	closeReasonIdle        = "idle timeout"
	closeReasonProtocol    = "protocol violation"
	closeReasonRate        = "rate limit"
	closeReasonUnavailable = "unavailable"
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
	replies  chan []byte       // transport replies to this client, written by the pump
	idle     *time.Timer

	// read-pump-owned (no lock): inbound rate limiting.
	bucket     *tokenBucket
	violations int
}

// replyQueue bounds queued replies per connection; a client that lets them
// pile up is not reading and is dropped.
const replyQueue = 16

func newSession(h *Handler, sock *websocket.Conn, client *hub.Client, grant *Grant) *session {
	return &session{
		h: h, sock: sock, client: client, grant: grant,
		closeReq: make(chan closeRequest, 1),
		replies:  make(chan []byte, replyQueue),
		bucket:   newTokenBucket(h.actionRate, h.actionBurst, h.now),
	}
}

// touch records inbound activity (a message or a pong), re-arming the idle
// deadline. Safe from any goroutine.
func (s *session) touch() {
	if s.idle != nil {
		s.idle.Reset(s.h.idleTimeout)
	}
}

// queueReply hands a reply to the write pump. It never blocks: a full queue
// means the client is not reading, so it is closed through the pump.
func (s *session) queueReply(b []byte) {
	select {
	case s.replies <- b:
	default:
		s.requestClose(websocket.StatusPolicyViolation, closeReasonEnded)
	}
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
	s.idle = time.AfterFunc(s.h.idleTimeout, func() {
		s.requestClose(websocket.StatusPolicyViolation, closeReasonIdle)
	})
	defer s.idle.Stop()
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
		case b := <-s.replies:
			if err := s.writeFrame(ctx, b); err != nil {
				s.fail()
				return
			}
		case <-ping.C:
			if err := s.pingPeer(ctx); err != nil {
				s.fail()
				return
			}
			s.touch() // a pong is inbound activity
		case <-expiry.C:
			s.closeSocket(websocket.StatusPolicyViolation, closeReasonEnded)
			return
		case <-s.grant.Revoked():
			s.closeSocket(websocket.StatusPolicyViolation, closeReasonEnded)
			return
		case rq := <-s.closeReq:
			s.flushReplies(ctx) // a refusal precedes the close that follows it
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

// flushReplies writes any queued replies without blocking for new ones.
func (s *session) flushReplies(ctx context.Context) {
	for {
		select {
		case b := <-s.replies:
			if err := s.writeFrame(ctx, b); err != nil {
				return
			}
		default:
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

// readPump reads inbound messages until the socket ends. It never writes or
// closes the socket: protocol violations and every other close go through
// requestClose to the write pump, and replies through queueReply. A read
// error (peer close, oversize, abort) drops the client so the pump finishes.
func (s *session) readPump(ctx context.Context) {
	s.sock.SetReadLimit(int64(s.h.maxMessage))
	for {
		typ, data, err := s.sock.Read(ctx)
		if err != nil {
			s.client.Close()
			return
		}
		s.touch()
		if typ != websocket.MessageText {
			s.violation(ctx)
			return
		}
		act, err := decodeAction(data)
		if err != nil {
			s.violation(ctx)
			return
		}
		if !s.dispatch(ctx, act) {
			return
		}
	}
}

// violation asks the pump for a policy-violation close and waits for it (the
// pump cancels ctx once the socket is closed), so the pump, not this
// goroutine, decides the close status.
func (s *session) violation(ctx context.Context) {
	s.requestClose(websocket.StatusPolicyViolation, closeReasonProtocol)
	<-ctx.Done()
}

// dispatch submits one decoded action to the hub as this connection's client
// and queues the verdict for this client only. It reports whether the read
// loop should continue.
func (s *session) dispatch(ctx context.Context, act hub.OperatorAction) bool {
	if v := s.h.revalidate(s.grant, s.client.Role()); v.reason != "" {
		s.queueReply(refusalFor(act, v.reason))
		if v.closeConn {
			s.requestClose(websocket.StatusPolicyViolation, closeReasonEnded)
			<-ctx.Done() // the pump flushes the refusal, closes, then cancels
			return false
		}
		return true
	}
	if !s.bucket.allow() {
		s.violations++
		s.queueReply(refusalFor(act, ReasonRateLimited))
		if s.violations >= s.h.maxViolations {
			s.requestClose(websocket.StatusPolicyViolation, closeReasonRate)
			<-ctx.Done()
			return false
		}
		return true
	}
	s.violations = 0
	actx, cancel := context.WithTimeout(ctx, s.h.actionTimeout)
	defer cancel()
	res, err := s.client.Do(actx, act)
	if err != nil {
		if ctx.Err() == nil && !errors.Is(err, hub.ErrClientClosed) {
			s.requestClose(websocket.StatusInternalError, closeReasonUnavailable)
			<-ctx.Done()
		}
		return false
	}
	s.queueReply(replyFor(act, res))
	return true
}
