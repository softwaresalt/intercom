package hub

import (
	"encoding/json"
	"strconv"
	"time"

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
	KindAgentTurnStart      Kind = "agent.turn.start"
	KindAgentTurnEnd        Kind = "agent.turn.end"
	KindAgentIdle           Kind = "agent.idle"
	KindPermissionRequested Kind = "permission.requested"
	KindPermissionResolved  Kind = "permission.resolved"
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
	}
	// TurnPayload is the payload of agent.turn.start and agent.turn.end.
	TurnPayload struct {
		TurnID string `json:"turn_id"`
	}
	// IdlePayload is the payload of agent.idle.
	IdlePayload struct {
		Aborted bool `json:"aborted"`
	}
	// PermissionRequestedPayload is provisional; shipment S4 owns the final
	// arbitration payloads.
	PermissionRequestedPayload struct {
		RequestID      string `json:"request_id"`
		ToolCallID     string `json:"tool_call_id"`
		PermissionKind string `json:"permission_kind"`
		Summary        string `json:"summary"`
	}
	// PermissionResolvedPayload is provisional; see PermissionRequestedPayload.
	PermissionResolvedPayload struct {
		RequestID string `json:"request_id"`
		Decision  string `json:"decision"`
		Actor     string `json:"actor"`
	}
	// ToolPayload is the payload of tool.started and tool.ended.
	ToolPayload struct {
		ToolCallID string `json:"tool_call_id"`
		ToolName   string `json:"tool_name"`
		Success    bool   `json:"success,omitempty"`
	}
	// ErrorPayload is the payload of hub-produced error envelopes.
	ErrorPayload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		RawType string `json:"raw_type,omitempty"`
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
// re-parsed. The output is byte-identical to json.Marshal of [Envelope].
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

// maxRawTypeLen bounds the SDK type name echoed in an unknown-event error.
const maxRawTypeLen = 128

// mapEvent converts a normalized agent event into an envelope kind and
// payload. An unmapped event becomes an error envelope (design §4.3: never a
// silent drop).
func mapEvent(ev agent.Event) (Kind, any) {
	kind, ok := KindFor(ev.Kind)
	if !ok {
		return KindError, ErrorPayload{
			Code:    ErrCodeUnknownEvent,
			Message: "unrecognised agent event",
			RawType: sanitize(ev.RawType, maxRawTypeLen),
		}
	}
	switch kind {
	case KindAgentMessage, KindAgentMessageDelta:
		return kind, MessagePayload{MessageID: ev.MessageID, Text: ev.Text}
	case KindAgentTurnStart, KindAgentTurnEnd:
		return kind, TurnPayload{TurnID: ev.TurnID}
	case KindAgentIdle:
		return kind, IdlePayload{Aborted: ev.Aborted}
	default: // tool.started, tool.ended
		return kind, ToolPayload{ToolCallID: ev.ToolCallID, ToolName: ev.ToolName, Success: ev.Success}
	}
}

// sanitize bounds s to max bytes (on a rune boundary) and replaces control
// characters so hub-produced diagnostics cannot smuggle terminal escapes.
func sanitize(s string, max int) string {
	out := make([]rune, 0, len(s))
	n := 0
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			r = '?'
		}
		sz := len(string(r))
		if n+sz > max {
			break
		}
		out = append(out, r)
		n += sz
	}
	return string(out)
}
