package hub

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

var uuidV7 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewGenIsUUIDv7(t *testing.T) {
	g, err := NewGen(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !uuidV7.MatchString(g) {
		t.Fatalf("gen %q is not a UUIDv7", g)
	}
}

func TestNewGenEmbedsMillisecondTimestamp(t *testing.T) {
	at := time.UnixMilli(0x0123456789AB)
	g, err := NewGen(at)
	if err != nil {
		t.Fatal(err)
	}
	if len(g) < 13 || g[:13] != "01234567-89ab" {
		t.Fatalf("gen %q does not carry the 48-bit ms timestamp", g)
	}
}

func TestHubConstructionsNeverShareGenInFastRestart(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 2000; i++ {
		h, err := New(Config{SessionID: "s"})
		if err != nil {
			t.Fatal(err)
		}
		g := h.Gen()
		if !uuidV7.MatchString(g) {
			t.Fatalf("hub gen %q is not a UUIDv7", g)
		}
		if _, dup := seen[g]; dup {
			t.Fatalf("gen %q repeated across constructions (iteration %d)", g, i)
		}
		seen[g] = struct{}{}
	}
}

func TestGenCapturedOnceAtConstruction(t *testing.T) {
	h, err := New(Config{SessionID: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if a, b := h.Gen(), h.Gen(); a != b || a == "" {
		t.Fatalf("Gen() unstable or empty: %q vs %q", a, b)
	}
}

func fixedTS() time.Time { return time.Date(2026, 9, 4, 19, 12, 3, 999, time.UTC) }

type goldenCase struct {
	kind    Kind
	payload any
	want    string
}

// goldenCases pins the exact wire bytes for each of the 12 initial kinds
// (design §5.2).
func goldenCases() []goldenCase {
	const head = `{"v":1,"gen":"G","seq":7,"ts":"2026-09-04T19:12:03Z","kind":"`
	return []goldenCase{
		{KindSessionHydrate, json.RawMessage(`{"entries":[],"turn_active":false,"truncated":false,"evicted_count":0}`),
			head + `session.hydrate","session_id":"sess","payload":{"entries":[],"turn_active":false,"truncated":false,"evicted_count":0}}`},
		{KindAgentMessage, MessagePayload{MessageID: "m1", Text: "hi"},
			head + `agent.message","session_id":"sess","payload":{"message_id":"m1","text":"hi"}}`},
		{KindAgentMessageDelta, MessagePayload{MessageID: "m1", Text: "h"},
			head + `agent.message.delta","session_id":"sess","payload":{"message_id":"m1","text":"h"}}`},
		{KindAgentTurnStart, TurnPayload{TurnID: "t1"},
			head + `agent.turn.start","session_id":"sess","payload":{"turn_id":"t1"}}`},
		{KindAgentTurnEnd, TurnPayload{TurnID: "t1"},
			head + `agent.turn.end","session_id":"sess","payload":{"turn_id":"t1"}}`},
		{KindAgentIdle, IdlePayload{Aborted: true},
			head + `agent.idle","session_id":"sess","payload":{"aborted":true}}`},
		{KindPermissionRequested, PermissionRequestedPayload{RequestID: "r1", ToolCallID: "c1", PermissionKind: "shell", Summary: "ls"},
			head + `permission.requested","session_id":"sess","payload":{"request_id":"r1","tool_call_id":"c1","permission_kind":"shell","summary":"ls"}}`},
		{KindPermissionResolved, PermissionResolvedPayload{RequestID: "r1", Decision: "allow_once", Actor: "tui"},
			head + `permission.resolved","session_id":"sess","payload":{"request_id":"r1","decision":"allow_once","actor":"tui"}}`},
		{KindToolStarted, ToolPayload{ToolCallID: "c1", ToolName: "bash"},
			head + `tool.started","session_id":"sess","payload":{"tool_call_id":"c1","tool_name":"bash"}}`},
		{KindToolEnded, ToolPayload{ToolCallID: "c1", ToolName: "bash", Success: true},
			head + `tool.ended","session_id":"sess","payload":{"tool_call_id":"c1","tool_name":"bash","success":true}}`},
		{KindError, ErrorPayload{Code: "degraded", Message: "ingress stalled"},
			head + `error","session_id":"sess","payload":{"code":"degraded","message":"ingress stalled"}}`},
		{KindServerShutdown, ShutdownPayload{Reason: "operator"},
			head + `server.shutdown","session_id":"sess","payload":{"reason":"operator"}}`},
	}
}

func TestEnvelopeGoldenForAll12Kinds(t *testing.T) {
	cases := goldenCases()
	if len(cases) != 12 {
		t.Fatalf("golden table has %d kinds, want 12", len(cases))
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			f, err := newFrame("G", 7, fixedTS(), tc.kind, "sess", tc.payload)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(f.Bytes()); got != tc.want {
				t.Fatalf("golden mismatch\n got: %s\nwant: %s", got, tc.want)
			}
			if f.Seq() != 7 || f.Gen() != "G" || f.Kind() != tc.kind {
				t.Fatalf("accessors wrong: %d %q %q", f.Seq(), f.Gen(), f.Kind())
			}
			var env Envelope
			if err := json.Unmarshal(f.Bytes(), &env); err != nil {
				t.Fatalf("bytes are not a valid Envelope: %v", err)
			}
			if env.V != EnvelopeVersion || env.Kind != tc.kind || env.SessionID != "sess" {
				t.Fatalf("decoded envelope wrong: %+v", env)
			}
		})
	}
}

func TestFrameBytesAreDefensivelyCopied(t *testing.T) {
	f, err := newFrame("G", 1, fixedTS(), KindAgentIdle, "sess", IdlePayload{})
	if err != nil {
		t.Fatal(err)
	}
	first := f.Bytes()
	want := string(first)
	for i := range first {
		first[i] = 'X'
	}
	if got := string(f.Bytes()); got != want {
		t.Fatalf("mutating a returned slice altered the published envelope:\n got: %s\nwant: %s", got, want)
	}
	// A value copy of the frame (as handed to another consumer) is unaffected too.
	g := f
	second := g.Bytes()
	second[0] = 'Y'
	if string(f.Bytes()) != want {
		t.Fatal("second consumer's mutation leaked into the shared envelope")
	}
}

func TestKindForMapsAgentEventKinds(t *testing.T) {
	cases := map[agent.EventKind]Kind{
		agent.KindAgentMessage: KindAgentMessage,
		agent.KindDelta:        KindAgentMessageDelta,
		agent.KindTurnStart:    KindAgentTurnStart,
		agent.KindTurnEnd:      KindAgentTurnEnd,
		agent.KindIdle:         KindAgentIdle,
		agent.KindToolStarted:  KindToolStarted,
		agent.KindToolEnded:    KindToolEnded,
	}
	for in, want := range cases {
		got, ok := KindFor(in)
		if !ok || got != want {
			t.Errorf("KindFor(%q) = %q,%v want %q,true", in, got, ok, want)
		}
	}
	if k, ok := KindFor(agent.KindUnknown); ok || k != "" {
		t.Errorf("KindFor(unknown) = %q,%v want \"\",false", k, ok)
	}
}

func TestMapEventUnknownBecomesErrorEnvelopeNotDropped(t *testing.T) {
	kind, payload := mapEvent(agent.Event{Kind: agent.KindUnknown, RawType: "assistant.reasoning"})
	if kind != KindError {
		t.Fatalf("unknown event kind = %q, want %q", kind, KindError)
	}
	ep, ok := payload.(ErrorPayload)
	if !ok || ep.Code != ErrCodeUnknownEvent || ep.RawType != "assistant.reasoning" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestMapEventKnownKindsCarryFields(t *testing.T) {
	kind, payload := mapEvent(agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c", ToolName: "n", Success: true})
	if kind != KindToolEnded {
		t.Fatal(kind)
	}
	if tp, ok := payload.(ToolPayload); !ok || tp != (ToolPayload{ToolCallID: "c", ToolName: "n", Success: true}) {
		t.Fatalf("payload = %#v", payload)
	}
}
