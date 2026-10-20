// Package wsclient is the reference Go client for the intercom-go WebSocket
// transport. It exists for two reasons: the integration tests drive the
// server through it, and it documents the wire protocol for SPA authors in
// executable form. Everything below is the protocol; the code is only its
// illustration.
//
// # Endpoint
//
// The server listens on loopback only and is reached through a tunnel. The
// WebSocket endpoint is
//
//	ws(s)://<host>/ws
//
// and nothing else is served. The request Host must be the loopback name for
// the bound port or one of [remote].allowed_hosts, otherwise 403.
//
// # Authentication (before the upgrade; there is no in-band login)
//
// Credentials are per-role (observer or approver) 256-bit bearer values shown
// to the operator on the local TUI once. They are NEVER sent in the URL, a
// cookie, or the first message. Exactly one of two carriages is used:
//
//   - Native clients (this package, CLI tools): the request header
//     "Authorization: Bearer <credential>" ([CarriageHeader]). Preferred.
//   - Browsers, which cannot set headers on a WebSocket: offer the
//     subprotocols "intercom.v1" and "bearer.<credential>" together, e.g.
//     new WebSocket(url, ["intercom.v1", "bearer." + credential])
//     ([CarriageSubprotocol]). The server selects only "intercom.v1"; the
//     credential entry is never echoed. A request that offers no "intercom.v1"
//     is refused. The subprotocol header can appear in tunnel or edge logs, so
//     prefer the header carriage whenever the client can set one.
//
// Offering both carriages, a malformed credential, or a credential in the URL
// is refused. Every credential failure (missing, malformed, wrong, expired,
// revoked, wrong generation) returns the same bare HTTP 401 with no body, so a
// client cannot tell them apart; it must ask the operator for a new one.
//
// A browser attaches an Origin header the page cannot control. The server
// accepts it only if it exactly matches [remote].allowed_origins (empty by
// default => every browser origin gets 403, even with a valid credential). A
// client that sends no Origin header (native clients) is accepted with a valid
// credential alone. Rejection statuses: 503 over the connection cap, 403 bad
// Host or Origin, 401 any credential failure, 400 malformed cursor.
//
// # Connecting, hydrating and resuming (design §5.3)
//
// The first frame after the upgrade is either:
//
//   - a "session.hydrate" envelope: a full snapshot (including
//     pending_permissions), sent on a first connect, when the client's
//     generation differs, or when the server's replay ring no longer covers the
//     client's position; or
//   - the replay of every envelope with seq+1 ... HEAD, when the client
//     resumed with a (gen, seq) cursor whose gen matches and whose seq is still
//     in the ring.
//
// To resume, put the last APPLIED position in the upgrade URL's query:
//
//	/ws?gen=<gen>&seq=<seq>
//
// (both or neither; these are not secrets). The client keys on (gen, seq),
// never seq alone: a restarted server has a new gen and renumbers from zero.
// It applies an envelope only if it is a hydrate or exactly seq+1 of the
// current position, discards duplicates (seq <= position), and treats a gap, an
// out-of-order frame or a gen change as a signal to reconnect WITHOUT a cursor
// to obtain a fresh hydrate. [Client.Next] returns that verdict with every
// envelope.
//
// # Frames the server sends (text messages, one JSON object each)
//
// Envelopes, broadcast to every client (hub.Envelope):
//
//	{"v":1,"gen":"<uuidv7>","seq":42,"ts":"<RFC3339Nano>","kind":"agent.message",
//	 "session_id":"...","payload":{...}}
//
// where kind is one of session.hydrate, agent.message, agent.message.delta,
// agent.turn.start, agent.turn.end, agent.idle, permission.requested,
// permission.resolved, permission.aborted, tool.started, tool.ended, error,
// server.shutdown. Payload deadlines for permissions are ts + remaining_ms,
// never the client's receipt time.
//
// Replies to your own actions, sent to you only (never broadcast, no seq):
//
//	{"v":1,"type":"action.accepted","op":"prompt"}
//	{"v":1,"type":"action.refused","op":"resolve_permission",
//	 "reason":"role_not_permitted","request_id":"r1","winner":"..."}
//
// Refusal reasons: role_not_permitted, destructive_remote_refused,
// already_resolved (winner names the winning actor), request_unknown,
// action_unknown, agent_unavailable, shutting_down, credential_revoked,
// credential_expired, generation_mismatch, rate_limited.
//
// # Frames the client sends (text messages only, at most 16 KiB each)
//
// Exactly one JSON object with string values and only these keys, each at most
// once:
//
//	{"op":"prompt","prompt":"text"}
//	{"op":"abort_turn"}
//	{"op":"resolve_permission","request_id":"r1","decision":"allow_once"|"deny","feedback":"..."}
//
// The role is never sent: it comes from the credential. An observer's
// privileged actions are refused with role_not_permitted. An unknown "op"
// string is refused with action_unknown and the connection stays open; any
// other malformed message (invalid JSON, unknown or duplicate keys, non-string
// values, binary frames, trailing data) is a protocol violation and the server
// closes the connection with status 1008.
//
// # Closing
//
// The server closes with a standard close frame; the client reconnects and
// resumes. 1008 (policy violation, reason "session ended", "protocol
// violation", "idle timeout" or "rate limit") means the credential is no
// longer valid or the client misbehaved: do not retry blindly with the same
// credential. 1009 means a message was too large. 1001 means the server is
// shutting down, the client was too slow (its buffer overflowed and it was
// dropped) or the connection ended; reconnect with the last applied cursor. A
// credential is valid for a fixed lifetime (default 4 hours, no refresh) and
// the connection is closed when it expires or is re-issued; there is no
// ping/pong duty on the client beyond keeping a reader running (the server
// sends pings and any compliant WebSocket library answers them while reading).
// A connection with no inbound activity for the idle timeout is closed.
package wsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Wire constants (mirrors of the server's; duplicated so this package never
// imports the server and can be read standalone).
const (
	// Path is the WebSocket endpoint path.
	Path = "/ws"
	// Subprotocol is the one subprotocol the server selects.
	Subprotocol  = "intercom.v1"
	bearerPrefix = "bearer."
)

// Carriage selects how the credential is presented.
type Carriage int

const (
	// CarriageHeader sends "Authorization: Bearer <credential>" (native).
	CarriageHeader Carriage = iota
	// CarriageSubprotocol offers "intercom.v1" and "bearer.<credential>"
	// (browser style).
	CarriageSubprotocol
)

// Options configures [Dial].
type Options struct {
	// URL is the ws:// or wss:// endpoint (path [Path]). The credential must
	// never be put in it; a query holding a credential is refused by the
	// server.
	URL string
	// Credential is the bearer credential.
	Credential string
	// Carriage selects header or subprotocol authentication.
	Carriage Carriage
	// Origin, when set, is sent as the Origin header (browser simulation).
	Origin string
	// Resume is the last applied (gen, seq) position from a previous
	// connection; nil requests a full hydrate.
	Resume *hub.Position
	// HTTPClient optionally overrides the HTTP client used for the handshake.
	HTTPClient *http.Client
}

// HandshakeError reports a refused upgrade. It carries only the HTTP status,
// which is all a client may learn.
type HandshakeError struct{ Status int }

func (e *HandshakeError) Error() string {
	return "wsclient: upgrade refused with HTTP status " + strconv.Itoa(e.Status)
}

// Verdict is the cursor's decision about one envelope.
type Verdict int

const (
	// VerdictNone is used for replies, which carry no cursor position.
	VerdictNone Verdict = iota
	// VerdictApply: the envelope is next in sequence (or a hydrate); apply it.
	VerdictApply
	// VerdictDiscard: a duplicate already applied; ignore it.
	VerdictDiscard
	// VerdictRehydrate: a gap, reorder or gen change; reconnect without a
	// cursor and discard frames until the hydrate arrives.
	VerdictRehydrate
)

// Reply is a transport reply to this client's own action.
type Reply struct {
	V         int    `json:"v"`
	Type      string `json:"type"`
	Op        string `json:"op"`
	Reason    string `json:"reason"`
	Winner    string `json:"winner"`
	RequestID string `json:"request_id"`
}

// Accepted reports an "action.accepted" reply.
func (r Reply) Accepted() bool { return r.Type == "action.accepted" }

// Message is one received frame: an Envelope with its cursor Verdict, or a
// Reply.
type Message struct {
	Raw      []byte
	Envelope *hub.Envelope
	Reply    *Reply
	Verdict  Verdict
}

// ErrUnrecognizedFrame reports a server frame that is neither an envelope nor
// a reply.
var ErrUnrecognizedFrame = errors.New("wsclient: unrecognized frame")

// Client is one connected session. It is not safe for concurrent use except
// that Send and Next may run in different goroutines (the underlying library
// permits one reader and one writer).
type Client struct {
	conn *websocket.Conn
	cur  cursor
}

// Dial authenticates and upgrades. On refusal it returns a *HandshakeError.
func Dial(ctx context.Context, o Options) (*Client, error) {
	u, err := url.Parse(o.URL)
	if err != nil {
		return nil, fmt.Errorf("wsclient: bad URL: %w", err)
	}
	if o.Resume != nil {
		q := u.Query()
		q.Set("gen", o.Resume.Gen)
		q.Set("seq", strconv.FormatUint(o.Resume.Seq, 10))
		u.RawQuery = q.Encode()
	}
	dopts := &websocket.DialOptions{HTTPClient: o.HTTPClient}
	hdr := http.Header{}
	switch o.Carriage {
	case CarriageSubprotocol:
		dopts.Subprotocols = []string{Subprotocol, bearerPrefix + o.Credential}
	default:
		hdr.Set("Authorization", "Bearer "+o.Credential)
	}
	if o.Origin != "" {
		hdr.Set("Origin", o.Origin)
	}
	dopts.HTTPHeader = hdr
	conn, resp, err := websocket.Dial(ctx, u.String(), dopts)
	if err != nil {
		if resp != nil {
			return nil, &HandshakeError{Status: resp.StatusCode}
		}
		return nil, fmt.Errorf("wsclient: dial failed: %w", scrub(err, o.Credential))
	}
	c := &Client{conn: conn}
	if o.Resume != nil {
		c.cur = cursor{gen: o.Resume.Gen, seq: o.Resume.Seq, have: true}
	}
	return c, nil
}

// scrub guards against a library error echoing the credential.
func scrub(err error, secret string) error {
	if secret != "" && strings.Contains(err.Error(), secret) {
		return errors.New("connection error (details withheld)")
	}
	return err
}

// Next reads and classifies the next frame. A close frame or transport error
// ends the session; inspect it with [CloseStatus].
func (c *Client) Next(ctx context.Context) (Message, error) {
	typ, data, err := c.conn.Read(ctx)
	if err != nil {
		return Message{}, err
	}
	if typ != websocket.MessageText {
		return Message{}, ErrUnrecognizedFrame
	}
	m := Message{Raw: data}
	var probe struct {
		Type string   `json:"type"`
		Kind hub.Kind `json:"kind"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return m, ErrUnrecognizedFrame
	}
	switch {
	case probe.Type == "action.accepted" || probe.Type == "action.refused":
		var r Reply
		if err := json.Unmarshal(data, &r); err != nil {
			return m, ErrUnrecognizedFrame
		}
		m.Reply = &r
	case probe.Kind != "":
		var env hub.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return m, ErrUnrecognizedFrame
		}
		m.Envelope = &env
		m.Verdict = c.cur.apply(env)
	default:
		return m, ErrUnrecognizedFrame
	}
	return m, nil
}

// Position returns the last applied (gen, seq) for a reconnect; false if there
// is no baseline or a re-hydrate is required.
func (c *Client) Position() (hub.Position, bool) { return c.cur.position() }

// Prompt sends a prompt (approver only).
func (c *Client) Prompt(ctx context.Context, text string) error {
	return c.send(ctx, map[string]string{"op": "prompt", "prompt": text})
}

// AbortTurn aborts the in-flight turn (approver only).
func (c *Client) AbortTurn(ctx context.Context) error {
	return c.send(ctx, map[string]string{"op": "abort_turn"})
}

// ResolvePermission answers a pending permission request (approver only).
func (c *Client) ResolvePermission(ctx context.Context, requestID string, allow bool, feedback string) error {
	m := map[string]string{"op": "resolve_permission", "request_id": requestID, "decision": "deny"}
	if allow {
		m["decision"] = "allow_once"
	}
	if feedback != "" {
		m["feedback"] = feedback
	}
	return c.send(ctx, m)
}

// SendRaw sends an arbitrary text message (protocol-violation tests).
func (c *Client) SendRaw(ctx context.Context, msg string) error {
	return c.conn.Write(ctx, websocket.MessageText, []byte(msg))
}

func (c *Client) send(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.conn.Write(ctx, websocket.MessageText, b)
}

// Close performs the close handshake.
func (c *Client) Close() error { return c.conn.Close(websocket.StatusNormalClosure, "") }

// Abort drops the connection without a handshake (simulates a network loss).
func (c *Client) Abort() { _ = c.conn.CloseNow() }

// CloseStatus extracts the close frame's code and reason from a read error.
func CloseStatus(err error) (code int, reason string, ok bool) {
	var ce websocket.CloseError
	if errors.As(err, &ce) {
		return int(ce.Code), ce.Reason, true
	}
	return 0, "", false
}

// cursor is the client-side idempotence and gap detector of design §5.3,
// keyed on (gen, seq).
type cursor struct {
	gen   string
	seq   uint64
	have  bool
	stale bool
}

func (c *cursor) apply(e hub.Envelope) Verdict {
	if e.Kind == hub.KindSessionHydrate {
		if c.have && e.Gen == c.gen && e.Seq < c.seq {
			c.stale = true // a snapshot older than what was applied is bogus
			return VerdictRehydrate
		}
		c.gen, c.seq, c.have, c.stale = e.Gen, e.Seq, true, false
		return VerdictApply
	}
	if !c.have || c.stale || e.Gen != c.gen {
		c.stale = true
		return VerdictRehydrate
	}
	switch {
	case e.Seq <= c.seq:
		return VerdictDiscard
	case e.Seq == c.seq+1:
		c.seq = e.Seq
		return VerdictApply
	default:
		c.stale = true
		return VerdictRehydrate
	}
}

func (c *cursor) position() (hub.Position, bool) {
	if !c.have || c.stale {
		return hub.Position{}, false
	}
	return hub.Position{Gen: c.gen, Seq: c.seq}, true
}
