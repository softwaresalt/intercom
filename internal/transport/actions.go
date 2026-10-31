package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Wire op names (inbound) and reply types (outbound). They are the SPA's
// contract; see package wsclient for the narrative description.
const (
	opResolvePermission = "resolve_permission"
	opPrompt            = "prompt"
	opAbortTurn         = "abort_turn"

	// Wire ops that exist only at the local TUI: they decode and are refused
	// with ReasonLocalOnlyAction, never sent to the hub.
	opRevealCredentials  = "reveal_credentials"
	opReissueCredentials = "reissue_credentials"
	opQuit               = "quit"
	opShutdown           = "shutdown"

	// MaxFeedbackBytes bounds the deny feedback accepted from the wire.
	MaxFeedbackBytes = hub.MaxDecisionFeedbackBytes
	// ReasonLocalOnlyAction is the refusal for an op that may only be taken at
	// the local TUI.
	ReasonLocalOnlyAction = string(hub.ReasonLocalOnlyAction)

	decisionAllowOnce = "allow_once"
	decisionDeny      = "deny"

	// ReplyAccepted and ReplyRefused are the transport-owned reply types sent
	// to the acting client only (never broadcast).
	ReplyAccepted = "action.accepted"
	ReplyRefused  = "action.refused"

	// maxEchoedRequestID bounds the client-supplied request id echoed back.
	maxEchoedRequestID = 128
)

// errProtocol marks an inbound message that is not a well-formed action: the
// connection gets a policy-violation close (ADR MVP-3).
var errProtocol = errors.New("transport: protocol violation")

// decodeAction strictly decodes one inbound text message into an operator
// action. The message must be exactly one JSON object whose keys are drawn
// from {op, request_id, decision, feedback, prompt}, each at most once, each
// with a string value (JSON null is NOT accepted as an empty string); "op" is required. Anything else (malformed JSON,
// trailing data, unknown or duplicate fields, non-string values, an unknown
// decision) is a protocol violation. An op that is a well-formed string but
// not a known operation decodes to hub.OpUnknown, which the hub refuses as
// action_unknown (ADR MVP-3) without closing the connection. The role and
// surface are never read from the message.
func decodeAction(data []byte) (hub.OperatorAction, error) {
	da, err := decodeWireAction(data)
	return da.act, err
}

// decodeWireAction is the strict decoder behind decodeAction (same grammar and
// violations). It additionally reports the local-only ops (reveal/reissue
// credentials, quit, shutdown) so the session can refuse them with
// local_only_action instead of the generic action_unknown, and it drops
// feedback on allow_once and bounds it on deny.
func decodeWireAction(data []byte) (decodedAction, error) {
	bad := func() (decodedAction, error) {
		return decodedAction{}, apperr.Wrap(apperr.KindPolicy, errProtocol)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return bad()
	}
	fields := make(map[string]string, 5)
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return bad()
		}
		key, ok := kt.(string)
		if !ok {
			return bad()
		}
		switch key {
		case "op", "request_id", "decision", "feedback", "prompt":
		default:
			return bad()
		}
		if _, dup := fields[key]; dup {
			return bad()
		}
		// Decode the raw value first: json.Unmarshal of null into a string
		// succeeds silently, so a null (or any non-string) must be refused
		// explicitly. Strictness: only a JSON string is a string value.
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil || len(raw) == 0 || raw[0] != '"' {
			return bad()
		}
		var v string
		if err := json.Unmarshal(raw, &v); err != nil {
			return bad()
		}
		fields[key] = v
	}
	if t, err := dec.Token(); err != nil || t != json.Delim('}') {
		return bad()
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return bad() // trailing data
	}
	op, ok := fields["op"]
	if !ok || op == "" {
		return bad()
	}
	act := hub.OperatorAction{RequestID: fields["request_id"], Prompt: fields["prompt"]}
	switch op {
	case opPrompt:
		act.Type = hub.OpPrompt
	case opAbortTurn:
		act.Type = hub.OpAbortTurn
	case opResolvePermission:
		act.Type = hub.OpResolvePermission
		switch fields["decision"] {
		case decisionAllowOnce:
			act.Decision.Kind = agent.DecisionAllowOnce
		case decisionDeny:
			act.Decision.Kind = agent.DecisionDeny
		default:
			return bad()
		}
		if act.Decision.Kind == agent.DecisionDeny {
			act.Decision.Feedback = truncUTF8Bytes(fields["feedback"], MaxFeedbackBytes)
		}
	case opRevealCredentials, opReissueCredentials, opQuit, opShutdown:
		return decodedAction{
			act:         hub.OperatorAction{Type: hub.OpUnknown, RequestID: fields["request_id"]},
			localOnlyOp: op,
		}, nil
	default:
		act = hub.OperatorAction{Type: hub.OpUnknown, RequestID: fields["request_id"]}
	}
	return decodedAction{act: act}, nil
}

// opName is the canonical wire name echoed in replies (never client text).
func opName(t hub.OpType) string {
	switch t {
	case hub.OpPrompt:
		return opPrompt
	case hub.OpAbortTurn:
		return opAbortTurn
	case hub.OpResolvePermission:
		return opResolvePermission
	default:
		return "unknown"
	}
}

// reply is the wire shape of a transport-owned reply.
type reply struct {
	V         int    `json:"v"`
	Type      string `json:"type"`
	Op        string `json:"op,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Winner    string `json:"winner,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// replyFor builds the reply for a hub verdict. It cannot fail for these
// field types; on the impossible marshal error it returns a fixed refusal.
func replyFor(act hub.OperatorAction, res hub.ActionResult) []byte {
	r := reply{V: 1, Type: ReplyAccepted, Op: opName(act.Type)}
	if !res.Accepted {
		r.Type = ReplyRefused
		r.Reason = string(res.Reason)
		r.Winner = res.Winner
	}
	r.RequestID = echoRequestID(act.RequestID)
	return marshalReply(r)
}

// echoRequestID bounds the client-supplied request id to maxEchoedRequestID
// bytes, cutting on a rune boundary so the echo stays valid UTF-8.
func echoRequestID(id string) string {
	return truncUTF8Bytes(id, maxEchoedRequestID)
}

// refusalFor builds a refusal with a transport-owned reason code, echoing the
// offending action's request id where it carried one.
func refusalFor(act hub.OperatorAction, reason string) []byte {
	return marshalReply(reply{
		V: 1, Type: ReplyRefused, Op: opName(act.Type), Reason: reason,
		RequestID: echoRequestID(act.RequestID),
	})
}

func marshalReply(r reply) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"v":1,"type":"action.refused","reason":"internal"}`)
	}
	return b
}

// decodedAction is the transport-internal result of decoding one inbound
// message: the hub action plus, for the wire ops that may only be taken at
// the local TUI, the canonical op name to echo in the refusal. It keeps
// hub.OperatorAction.Prompt strictly for OpPrompt text.
type decodedAction struct {
	act hub.OperatorAction
	// localOnlyOp is non-empty for a well-formed op that is refused on the wire.
	localOnlyOp string
}

// truncUTF8Bytes truncates s to at most maxBytes bytes, cutting on a rune
// boundary so the result stays valid UTF-8.
func truncUTF8Bytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// refusalForWire builds a refusal for a decoded wire action: a local-only op
// echoes its own canonical name, every other action the hub op name.
func refusalForWire(da decodedAction, reason string) []byte {
	if da.localOnlyOp != "" {
		return marshalReply(reply{
			V: 1, Type: ReplyRefused, Op: da.localOnlyOp, Reason: reason,
			RequestID: echoRequestID(da.act.RequestID),
		})
	}
	return refusalFor(da.act, reason)
}
