package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

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
	bad := func() (hub.OperatorAction, error) {
		return hub.OperatorAction{}, apperr.Wrap(apperr.KindPolicy, errProtocol)
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
		act.Decision.Feedback = fields["feedback"]
	default:
		act = hub.OperatorAction{Type: hub.OpUnknown}
	}
	return act, nil
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
	if id := act.RequestID; id != "" {
		if len(id) > maxEchoedRequestID {
			id = id[:maxEchoedRequestID]
		}
		r.RequestID = id
	}
	return marshalReply(r)
}

// refusalFor builds a refusal with a transport-owned reason code.
func refusalFor(act hub.OperatorAction, reason string) []byte {
	return marshalReply(reply{V: 1, Type: ReplyRefused, Op: opName(act.Type), Reason: reason})
}

func marshalReply(r reply) []byte {
	b, err := json.Marshal(r)
	if err != nil {
		return []byte(`{"v":1,"type":"action.refused","reason":"internal"}`)
	}
	return b
}
