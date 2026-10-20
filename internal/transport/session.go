package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	// DefaultActionTimeout bounds one hub action round trip. The read pump is
	// blocked for it and cannot answer a ping meanwhile, so it must stay well
	// below DefaultWriteTimeout (the ping's deadline); see
	// [actionFitsWriteTimeout], which NewHandler enforces.
	DefaultActionTimeout = 5 * time.Second
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
	// abortRead cancels the read pump's context. coder/websocket closes the connection when a Read context ends, which is the only way to abort a close handshake that Close is already waiting on (CloseNow is a no-op wait once Close has begun).
	abortRead context.CancelFunc
	// rawConn is the accepted connection (nil when the handler is not served by Server, e.g. httptest). Closing it is the abort of last resort: once Close has begun, CloseNow only waits, and the library's own handshake wait is 5s.
	rawConn net.Conn

	// read is the read-pump body; it is s.readPump except in tests that
	// exercise the panic guard.
	read func(ctx context.Context)

	// read-pump-owned (no lock): inbound rate limiting.
	bucket     *tokenBucket
	violations int
}

// replyQueue bounds queued replies per connection; a client that lets them
// pile up is not reading and is dropped.
const replyQueue = 16

func newSession(h *Handler, sock *websocket.Conn, client *hub.Client, grant *Grant) *session {
	s := &session{
		h: h, sock: sock, client: client, grant: grant,
		closeReq: make(chan closeRequest, 1),
		replies:  make(chan []byte, replyQueue),
		bucket:   newTokenBucket(h.actionRate, h.actionBurst, h.now),
	}
	s.read = s.readPump
	return s
}

// touch records inbound activity (a message or a pong), re-arming the idle
// deadline. Safe from any goroutine.
//
// Idle semantics: a pong answering the server's ping IS activity (the pump
// calls touch after each successful ping round trip). A peer that keeps
// answering pings therefore never idles out; the idle timer only ends a
// connection whose peer has gone silent, and the credential lifetime (expiry
// or revocation) is what bounds an otherwise healthy session.
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
	s.abortRead = cancel
	s.idle = time.AfterFunc(s.h.idleTimeout, func() {
		s.requestClose(websocket.StatusPolicyViolation, closeReasonIdle)
	})
	defer s.idle.Stop()
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		// A panic in the read pump (decode, dispatch, hub client) must not
		// take the process down: close the client so the write pump finishes
		// the session. The panic value is never logged (it could carry
		// message content); only its type is.
		defer func() {
			if r := recover(); r != nil {
				s.h.log.Error("transport: read pump panic recovered", "panic_type", fmt.Sprintf("%T", r))
				s.client.Close()
			}
		}()
		s.read(ctx)
	}()
	s.writePump(ctx)
	cancel()
	s.client.Close()
	<-readDone
}

// endedClose delivers the neutral "session ended" close for a credential that
// expired or was revoked. Replies already queued (notably a refusal of the
// action that exposed the problem) are flushed first so ADR MVP-3's refusal
// reaches the client before the 1008 close.
func (s *session) endedClose(ctx context.Context) {
	s.flushReplies(ctx)
	s.closeSocket(websocket.StatusPolicyViolation, closeReasonEnded)
}

// writePump is the sole writer of the socket. Frames, keep-alive pings, the
// credential-expiry deadline, revocation, close requests, drop (done) and
// shutdown are all cases of ONE select, so no second goroutine can interleave
// a frame.
func (s *session) writePump(ctx context.Context) {
	ping := time.NewTicker(s.h.pingEvery)
	defer ping.Stop()
	// The expiry timer is derived from the credential store's clock, the same
	// source Validate uses, so the two cannot disagree about expiry.
	expiry := time.NewTimer(s.h.cfg.Creds.Remaining(s.grant))
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
		// Revocation and expiry come next, ahead of frames, pings and queued
		// replies: a dead credential must not keep receiving traffic.
		select {
		case <-s.grant.Revoked():
			s.endedClose(ctx)
			return
		case <-expiry.C:
			s.endedClose(ctx)
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
			s.endedClose(ctx)
			return
		case <-s.grant.Revoked():
			s.endedClose(ctx)
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
	abort := time.AfterFunc(s.h.closeGrace, func() {
		if s.abortRead != nil {
			s.abortRead() // ends the in-flight Read, so the library tears the socket down and Close returns
		}
		if s.rawConn != nil {
			_ = s.rawConn.Close() // abort after done: releases the slot and unblocks Close
		}
		_ = s.sock.CloseNow()
	})
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
//
// The revalidate step compares the grant's role with the role the hub client
// was registered with. Both come from the same credential, so a mismatch
// cannot happen today; the comparison is intentional defence in depth (a
// future change to either side must fail closed, never silently widen
// privilege).
func (s *session) dispatch(ctx context.Context, act hub.OperatorAction) bool {
	v := s.h.revalidate(s.grant, s.client.Role())
	if v.closeConn {
		s.queueReply(refusalFor(act, v.reason))
		s.requestClose(websocket.StatusPolicyViolation, closeReasonEnded)
		<-ctx.Done() // the pump flushes the refusal, closes, then cancels
		return false
	}
	// Every other path, including the non-closing shutdown-revoked refusal,
	// goes through the per-connection limiter below, so a flood during the
	// revoked state is rate limited and eventually closed.
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
	if v.reason != "" {
		// Non-closing refusal (shutdown-revoked): it passed the limiter but is
		// not a successful action, so the violation count is not reset.
		s.queueReply(refusalFor(act, v.reason))
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
