package copilot

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

func newBufLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func TestNormalizeMappedEvents(t *testing.T) {
	cases := []struct {
		name string
		data rpc.SessionEventData
		want agent.Event
	}{
		{"message", &rpc.AssistantMessageData{MessageID: "m1", Content: "hello"},
			agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "hello"}},
		{"delta", &rpc.AssistantMessageDeltaData{MessageID: "m1", DeltaContent: "he"},
			agent.Event{Kind: agent.KindDelta, MessageID: "m1", Text: "he"}},
		{"turn start", &rpc.AssistantTurnStartData{TurnID: "t1"},
			agent.Event{Kind: agent.KindTurnStart, TurnID: "t1"}},
		{"turn end", &rpc.AssistantTurnEndData{TurnID: "t1"},
			agent.Event{Kind: agent.KindTurnEnd, TurnID: "t1"}},
		{"idle", &rpc.SessionIdleData{},
			agent.Event{Kind: agent.KindIdle}},
		{"idle aborted", &rpc.SessionIdleData{Aborted: boolp(true)},
			agent.Event{Kind: agent.KindIdle, Aborted: true}},
		{"tool started", &rpc.ToolExecutionStartData{ToolCallID: "c1", ToolName: "shell"},
			agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c1", ToolName: "shell"}},
		{"tool ended", &rpc.ToolExecutionCompleteData{ToolCallID: "c1", Success: true},
			agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c1", Success: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log, buf := newBufLogger()
			got := normalize(rpc.SessionEvent{Data: tc.data}, log)
			if got != tc.want {
				t.Fatalf("normalize = %+v, want %+v", got, tc.want)
			}
			if buf.Len() != 0 {
				t.Errorf("mapped event must not log; got %q", buf.String())
			}
		})
	}
}

func TestNormalizeUnknownVariantLogsWarnAndIsExplicit(t *testing.T) {
	log, buf := newBufLogger()
	ev := rpc.SessionEvent{Data: &rpc.AssistantReasoningData{}}
	got := normalize(ev, log)
	if got.Kind != agent.KindUnknown {
		t.Fatalf("Kind = %q, want unknown", got.Kind)
	}
	if got.RawType == "" || got.RawType != string(ev.Type()) {
		t.Errorf("RawType = %q, want %q", got.RawType, ev.Type())
	}
	out := buf.String()
	if !strings.Contains(out, "level=WARN") || !strings.Contains(out, got.RawType) {
		t.Errorf("expected warn log naming %q, got %q", got.RawType, out)
	}
}

func TestNormalizeNilDataIsUnknownNotPanic(t *testing.T) {
	log, buf := newBufLogger()
	got := normalize(rpc.SessionEvent{}, log)
	if got.Kind != agent.KindUnknown {
		t.Fatalf("Kind = %q, want unknown", got.Kind)
	}
	if !strings.Contains(buf.String(), "level=WARN") {
		t.Errorf("nil Data must warn, got %q", buf.String())
	}
}

func TestNormalizeNilLoggerDoesNotPanic(t *testing.T) {
	got := normalize(rpc.SessionEvent{Data: &rpc.AssistantReasoningData{}}, nil)
	if got.Kind != agent.KindUnknown {
		t.Fatalf("Kind = %q, want unknown", got.Kind)
	}
}
