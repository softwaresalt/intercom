package copilot

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"

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
	ev := rpc.SessionEvent{Data: &rpc.SessionErrorData{}}
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

func TestIsInformationalClassification(t *testing.T) {
	for _, raw := range []string{"session.usage_info", "assistant.usage", "model.call_start", "session.start", "pending_messages.modified"} {
		if !isInformational(raw) {
			t.Errorf("%q should be informational", raw)
		}
	}
	for _, raw := range []string{"model.call_failure", "session.error", "", "totally.new"} {
		if isInformational(raw) {
			t.Errorf("%q must stay visible as unknown", raw)
		}
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
	got := normalize(rpc.SessionEvent{Data: &rpc.SessionErrorData{}}, nil)
	if got.Kind != agent.KindUnknown {
		t.Fatalf("Kind = %q, want unknown", got.Kind)
	}
}

func TestNormalizeToolEndedExtractsResultAndErrorText(t *testing.T) {
	ok := normalize(rpc.SessionEvent{Data: &rpc.ToolExecutionCompleteData{
		ToolCallID: "c1", Success: true, Result: &rpc.ToolExecutionCompleteResult{Content: "done output"},
	}}, nil)
	if ok.Text != "done output" || !ok.Success {
		t.Errorf("success Text = %q success=%v", ok.Text, ok.Success)
	}
	fail := normalize(rpc.SessionEvent{Data: &rpc.ToolExecutionCompleteData{
		ToolCallID: "c2", Error: &rpc.ToolExecutionCompleteError{Message: "boom"},
	}}, nil)
	if fail.Text != "boom" || fail.Success {
		t.Errorf("error Text = %q success=%v", fail.Text, fail.Success)
	}
	bare := normalize(rpc.SessionEvent{Data: &rpc.ToolExecutionCompleteData{ToolCallID: "c3", Success: true}}, nil)
	if bare.Text != "" {
		t.Errorf("no result must leave Text empty, got %q", bare.Text)
	}
}

func TestNormalizeBoundsTextAndIdentifiersRuneSafely(t *testing.T) {
	big := strings.Repeat("é", MaxEventTextBytes) // 2 bytes per rune
	got := normalize(rpc.SessionEvent{Data: &rpc.AssistantMessageData{MessageID: strings.Repeat("i", 300), Content: big}}, nil)
	if !got.Truncated {
		t.Error("oversized text must set Truncated")
	}
	if len(got.Text) > MaxEventTextBytes || !utf8.ValidString(got.Text) {
		t.Errorf("Text len=%d valid=%v", len(got.Text), utf8.ValidString(got.Text))
	}
	if len(got.MessageID) != MaxEventIdentBytes {
		t.Errorf("MessageID len = %d, want %d", len(got.MessageID), MaxEventIdentBytes)
	}
	small := normalize(rpc.SessionEvent{Data: &rpc.AssistantMessageData{MessageID: "m", Content: "ok"}}, nil)
	if small.Truncated {
		t.Error("small event must not be Truncated")
	}
}

func TestBoundUTF8SinglePassRuneSafe(t *testing.T) {
	if s, tr := boundUTF8("hello", 5); s != "hello" || tr {
		t.Errorf("exact fit = %q,%v", s, tr)
	}
	if s, tr := boundUTF8("hello", 3); s != "hel" || !tr {
		t.Errorf("ascii cut = %q,%v", s, tr)
	}
	for _, in := range []string{"aé", "a€", "a😀", "😀😀😀"} {
		for max := 0; max <= len(in); max++ {
			s, tr := boundUTF8(in, max)
			if !utf8.ValidString(s) || len(s) > max {
				t.Errorf("boundUTF8(%q,%d) = %q (valid=%v)", in, max, s, utf8.ValidString(s))
			}
			if tr != (len(in) > max) {
				t.Errorf("boundUTF8(%q,%d) truncated=%v", in, max, tr)
			}
		}
	}
}

func TestUnmappedLogGateWarnsOncePerTypeThenDebug(t *testing.T) {
	log, buf := newBufLogger()
	var g unmappedLogGate
	ev := rpc.SessionEvent{Data: &rpc.SessionErrorData{}}
	for range 3 {
		if got := normalizeWithGate(ev, log, &g); got.Kind != agent.KindUnknown {
			t.Fatalf("Kind = %q", got.Kind)
		}
	}
	if n := strings.Count(buf.String(), "level=WARN"); n != 1 {
		t.Errorf("WARN count = %d, want 1\n%s", n, buf.String())
	}
	if n := strings.Count(buf.String(), "level=DEBUG"); n != 2 {
		t.Errorf("DEBUG count = %d, want 2\n%s", n, buf.String())
	}
	if !g.first("brand.new.type") || g.first("brand.new.type") {
		t.Error("first must be true once per type")
	}
}

func TestUnmappedLogGateBoundsMemory(t *testing.T) {
	var g unmappedLogGate
	for i := range maxWarnedEventTypes + 50 {
		g.first(fmt.Sprintf("type.%d", i))
	}
	if len(g.warned) > maxWarnedEventTypes {
		t.Errorf("warned set grew to %d, cap %d", len(g.warned), maxWarnedEventTypes)
	}
}

func TestBoundUTF8InvalidInputIsBoundedAndNeverSplitsValidRunes(t *testing.T) {
	cont := strings.Repeat("\x80", 1<<16) // continuation bytes only: not valid UTF-8
	s, tr := boundUTF8(cont, 10)
	if !tr || len(s) < 10-(utf8.UTFMax-1) || len(s) > 10 {
		t.Errorf("invalid input cut = len %d truncated=%v, want within 3 bytes of the limit", len(s), tr)
	}
	// Valid text after an invalid prefix is still cut on a rune boundary.
	mixed := "\xff\xfe" + strings.Repeat("é", 10)
	if got, _ := boundUTF8(mixed, 5); !utf8.ValidString(got[2:]) {
		t.Errorf("valid tail split mid-rune: %q", got)
	}
}
