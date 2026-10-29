package hub

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// EnvelopeVersion is the envelope schema version carried in the "v" field.
const EnvelopeVersion = 1

// Kind is the envelope kind vocabulary. It is intercom-owned and deliberately
// not the SDK's (design §5.2).
type Kind string

// The initial envelope kinds (design §5.2).
const (
	KindSessionHydrate      Kind = "session.hydrate"
	KindAgentMessage        Kind = "agent.message"
	KindAgentMessageDelta   Kind = "agent.message.delta"
	KindAgentReasoning      Kind = "agent.reasoning"
	KindAgentReasoningDelta Kind = "agent.reasoning.delta"
	KindAgentTurnStart      Kind = "agent.turn.start"
	KindAgentTurnEnd        Kind = "agent.turn.end"
	KindAgentIdle           Kind = "agent.idle"
	KindPermissionRequested Kind = "permission.requested"
	KindPermissionResolved  Kind = "permission.resolved"
	KindPermissionAborted   Kind = "permission.aborted"
	KindToolStarted         Kind = "tool.started"
	KindToolEnded           Kind = "tool.ended"
	KindError               Kind = "error"
	KindServerShutdown      Kind = "server.shutdown"
)

// Error codes carried in [ErrorPayload.Code] for hub-produced error
// envelopes.
const (
	// ErrCodeUnknownEvent reports an agent event the adapter could not map.
	ErrCodeUnknownEvent = "unknown_event"
	// ErrCodeDegraded reports an out-of-band degraded signal from the adapter.
	ErrCodeDegraded = "degraded"
	// ErrCodeFrameTooLarge replaces a live envelope that exceeded the frame cap.
	ErrCodeFrameTooLarge = "frame_too_large"
)

// Envelope is the decoded wire shape of one envelope (design §5.2). It is
// used by consumers to decode [Frame.Bytes]; the hub itself never publishes
// an Envelope value, only pre-marshalled frames.
type Envelope struct {
	V         int             `json:"v"`
	Gen       string          `json:"gen"`
	Seq       uint64          `json:"seq"`
	TS        string          `json:"ts"`
	Kind      Kind            `json:"kind"`
	SessionID string          `json:"session_id"`
	Payload   json.RawMessage `json:"payload"`
}

// Kind-specific payloads. Field order is the wire order.
type (
	// MessagePayload is the payload of agent.message and agent.message.delta.
	MessagePayload struct {
		MessageID string `json:"message_id"`
		Text      string `json:"text"`
		Truncated bool   `json:"truncated,omitempty"`
	}
	// TurnPayload is the payload of agent.turn.start and agent.turn.end.
	TurnPayload struct {
		TurnID    string `json:"turn_id"`
		Truncated bool   `json:"truncated,omitempty"`
	}
	// IdlePayload is the payload of agent.idle.
	IdlePayload struct {
		Aborted bool `json:"aborted"`
	}
	// PermissionRequestedPayload is the payload of permission.requested. It is
	// broadcast to every surface, observers included (ADR MVP-3).
	PermissionRequestedPayload struct {
		RequestID      string `json:"request_id"`
		ToolCallID     string `json:"tool_call_id"`
		PermissionKind string `json:"permission_kind"`
		Summary        string `json:"summary"`
		// RemainingMS is the time left before the hub-owned timeout resolves
		// the request to UserNotAvailable.
		RemainingMS int64 `json:"remaining_ms,omitempty"`
	}
	// PermissionResolvedPayload is the payload of permission.resolved: it is
	// published only after the agent adapter's handler acknowledged receiving
	// the decision (design §5.4 item 7).
	PermissionResolvedPayload struct {
		RequestID string `json:"request_id"`
		Decision  string `json:"decision"`
		// Actor is the winning surface: local-tui, remote-approver or hub.
		Actor string `json:"actor"`
		// Reason qualifies a hub-made decision (timeout, shutdown,
		// no_surface); it is empty when an operator answered.
		Reason string `json:"reason,omitempty"`
	}
	// PermissionAbortedPayload is the payload of permission.aborted: the
	// request ended without the agent receiving a decision.
	PermissionAbortedPayload struct {
		RequestID string `json:"request_id"`
		Reason    string `json:"reason"`
	}
	// ToolStartedPayload is the payload of tool.started.
	ToolStartedPayload struct {
		ToolCallID string `json:"tool_call_id"`
		ToolName   string `json:"tool_name"`
		Truncated  bool   `json:"truncated,omitempty"`
	}
	// ToolEndedPayload is the payload of tool.ended; success is always present.
	ToolEndedPayload struct {
		ToolCallID string `json:"tool_call_id"`
		ToolName   string `json:"tool_name"`
		Success    bool   `json:"success"`
		Text       string `json:"text,omitempty"`
		Truncated  bool   `json:"truncated,omitempty"`
	}
	// ErrorPayload is the payload of hub-produced error envelopes.
	ErrorPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	// ShutdownPayload is the payload of server.shutdown.
	ShutdownPayload struct {
		Reason string `json:"reason"`
	}
)

// Frame is one immutable, pre-marshalled envelope. All fields are unexported
// and Bytes returns a fresh copy, so one Frame can be shared by the TUI pump
// and any number of write pumps without a data race (design §5.2): freeze at
// seq assignment, broadcast bytes.
type Frame struct {
	gen  string
	seq  uint64
	kind Kind
	data []byte // never exposed, never mutated after newFrame
}

// Bytes returns a copy of the envelope's JSON encoding. Mutating the result
// does not affect any other holder of the frame.
func (f Frame) Bytes() []byte {
	out := make([]byte, len(f.data))
	copy(out, f.data)
	return out
}

// Len is the encoded size in bytes.
func (f Frame) Len() int { return len(f.data) }

// Seq is the envelope sequence number. For a session.hydrate frame it is the
// high-water mark the snapshot reflects.
func (f Frame) Seq() uint64 { return f.seq }

// Gen is the process generation of the hub that produced the frame.
func (f Frame) Gen() string { return f.gen }

// Kind is the envelope kind.
func (f Frame) Kind() Kind { return f.kind }

// newFrame marshals payload and freezes the envelope. It is only ever called
// from the hub goroutine (and from tests).
func newFrame(gen string, seq uint64, ts time.Time, kind Kind, sessionID string, payload any) (Frame, error) {
	var raw []byte
	switch p := payload.(type) {
	case nil:
		raw = []byte("{}")
	case json.RawMessage:
		raw = p
	case []byte:
		raw = p
	default:
		b, err := json.Marshal(payload)
		if err != nil {
			return Frame{}, apperr.Wrapf(apperr.KindIO, err, "hub: marshal %s payload", kind)
		}
		raw = b
	}
	return encodeFrame(gen, seq, ts, kind, sessionID, raw)
}

// encodeFrame assembles the envelope bytes by concatenation so that a
// pre-marshalled payload (the incremental hydrate snapshot) is copied, not
// re-parsed. The output decodes to the same [Envelope] as json.Marshal of it
// and, for payloads produced by json.Marshal (which HTML-escapes <, > and &),
// matches it byte for byte (see the escaping golden test); a caller-supplied
// raw payload is embedded as given.
func encodeFrame(gen string, seq uint64, ts time.Time, kind Kind, sessionID string, payload []byte) (Frame, error) {
	genQ, err := json.Marshal(gen)
	if err != nil {
		return Frame{}, apperr.Wrapf(apperr.KindIO, err, "hub: marshal gen")
	}
	sessQ, err := json.Marshal(sessionID)
	if err != nil {
		return Frame{}, apperr.Wrapf(apperr.KindIO, err, "hub: marshal session id")
	}
	b := make([]byte, 0, 96+len(genQ)+len(sessQ)+len(payload))
	b = append(b, `{"v":`...)
	b = strconv.AppendInt(b, EnvelopeVersion, 10)
	b = append(b, `,"gen":`...)
	b = append(b, genQ...)
	b = append(b, `,"seq":`...)
	b = strconv.AppendUint(b, seq, 10)
	b = append(b, `,"ts":"`...)
	b = ts.UTC().AppendFormat(b, "2006-01-02T15:04:05Z")
	b = append(b, `","kind":"`...)
	b = append(b, kind...)
	b = append(b, `","session_id":`...)
	b = append(b, sessQ...)
	b = append(b, `,"payload":`...)
	b = append(b, payload...)
	b = append(b, '}')
	return Frame{gen: gen, seq: seq, kind: kind, data: b}, nil
}

// KindFor maps an agent event kind to its envelope kind. It reports false
// for agent.KindUnknown and for kinds the hub does not know.
func KindFor(k agent.EventKind) (Kind, bool) {
	switch k {
	case agent.KindAgentMessage:
		return KindAgentMessage, true
	case agent.KindDelta:
		return KindAgentMessageDelta, true
	case agent.KindReasoning:
		return KindAgentReasoning, true
	case agent.KindReasoningDelta:
		return KindAgentReasoningDelta, true
	case agent.KindTurnStart:
		return KindAgentTurnStart, true
	case agent.KindTurnEnd:
		return KindAgentTurnEnd, true
	case agent.KindIdle:
		return KindAgentIdle, true
	case agent.KindToolStarted:
		return KindToolStarted, true
	case agent.KindToolEnded:
		return KindToolEnded, true
	default:
		return "", false
	}
}

// unknownEventError is the single error payload for an unmappable agent
// event. Its message is fixed: adapter-controlled text (the raw SDK type)
// must never reach remote clients or the fold.
func unknownEventError() ErrorPayload {
	return ErrorPayload{Code: ErrCodeUnknownEvent, Message: msgUnknownEvent}
}

// mapEvent converts a normalized agent event into an envelope kind and
// payload. Text is capped at textCap bytes and identifiers at maxIdentBytes
// (the same caps the fold applies), with truncated:true set when anything was
// cut, so a live envelope can never exceed the declared bounds. An unmapped
// event becomes an error envelope (design §4.3: never a silent drop).
func mapEvent(ev agent.Event, textCap int) (Kind, any) {
	kind, ok := KindFor(ev.Kind)
	if !ok {
		return KindError, unknownEventError()
	}
	switch kind {
	case KindAgentMessage, KindAgentMessageDelta, KindAgentReasoning, KindAgentReasoningDelta:
		id, t1 := capIdent(ev.MessageID)
		text := truncUTF8(ev.Text, textCap)
		return kind, MessagePayload{MessageID: id, Text: text, Truncated: t1 || len(text) < len(ev.Text) || ev.Truncated}
	case KindAgentTurnStart, KindAgentTurnEnd:
		id, t := capIdent(ev.TurnID)
		return kind, TurnPayload{TurnID: id, Truncated: t || ev.Truncated}
	case KindAgentIdle:
		return kind, IdlePayload{Aborted: ev.Aborted}
	case KindToolStarted:
		id, t1 := capIdent(ev.ToolCallID)
		name, t2 := capIdent(ev.ToolName)
		return kind, ToolStartedPayload{ToolCallID: id, ToolName: name, Truncated: t1 || t2 || ev.Truncated}
	default: // tool.ended
		id, t1 := capIdent(ev.ToolCallID)
		name, t2 := capIdent(ev.ToolName)
		text := truncUTF8(ev.Text, textCap)
		return kind, ToolEndedPayload{ToolCallID: id, ToolName: name, Success: ev.Success, Text: text,
			Truncated: t1 || t2 || len(text) < len(ev.Text) || ev.Truncated}
	}
}

// capIdent bounds an identifier to maxIdentBytes and reports truncation.
func capIdent(s string) (string, bool) {
	out := truncUTF8(s, maxIdentBytes)
	return out, len(out) < len(s)
}

// sanitize bounds s to limit bytes (on a rune boundary) and replaces every
// control rune (C0, DEL and C1 such as U+009B CSI), the Unicode line/paragraph
// separators and format runes (bidi overrides, zero-width) with '?', so
// hub-produced diagnostics cannot smuggle terminal escapes or spoof text.
func sanitize(s string, limit int) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Cf, r) {
			r = '?'
		}
		sz := utf8.RuneLen(r)
		if n+sz > limit {
			break
		}
		b.WriteRune(r)
		n += sz
	}
	return b.String()
}
