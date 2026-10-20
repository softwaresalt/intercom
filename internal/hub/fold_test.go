package hub

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

func recordedSequence() []agent.Event {
	return []agent.Event{
		{Kind: agent.KindTurnStart, TurnID: "t1"},
		{Kind: agent.KindDelta, MessageID: "m1", Text: "Hel"},
		{Kind: agent.KindDelta, MessageID: "m1", Text: "lo"},
		{Kind: agent.KindToolStarted, ToolCallID: "c1", ToolName: "bash"},
		{Kind: agent.KindToolEnded, ToolCallID: "c1", ToolName: "bash", Success: true},
		{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "Hello!"},
		{Kind: agent.KindTurnEnd, TurnID: "t1"},
		{Kind: agent.KindIdle},
	}
}

const recordedGolden = `{"entries":[` +
	`{"kind":"message","id":"m1","text":"Hello!","status":"complete"},` +
	`{"kind":"tool","id":"c1","tool_name":"bash","status":"succeeded"}` +
	`],"turn_active":false,"turn_id":"t1","truncated":false,"evicted_count":0}`

func mustFold(t *testing.T, l Limits) *Fold {
	t.Helper()
	f, err := NewFold(l)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFoldIsDeterministicForRecordedSequence(t *testing.T) {
	a, b := mustFold(t, Limits{}), mustFold(t, Limits{})
	for _, ev := range recordedSequence() {
		a.Apply(ev)
		b.Apply(ev)
	}
	if got := string(a.Snapshot()); got != recordedGolden {
		t.Fatalf("snapshot mismatch\n got: %s\nwant: %s", got, recordedGolden)
	}
	if string(a.Snapshot()) != string(b.Snapshot()) {
		t.Fatal("two folds of the same sequence diverged")
	}
}

func TestFoldStreamingMidTurnState(t *testing.T) {
	f := mustFold(t, Limits{})
	seq := recordedSequence()
	for _, ev := range seq[:3] {
		f.Apply(ev)
	}
	want := `{"entries":[{"kind":"message","id":"m1","text":"Hello","status":"streaming"}],` +
		`"turn_active":true,"turn_id":"t1","truncated":false,"evicted_count":0}`
	if got := string(f.Snapshot()); got != want {
		t.Fatalf("mid-turn snapshot\n got: %s\nwant: %s", got, want)
	}
}

func TestFoldSnapshotIsValidJSON(t *testing.T) {
	f := mustFold(t, Limits{})
	for _, ev := range recordedSequence() {
		f.Apply(ev)
	}
	f.Apply(agent.Event{Kind: agent.KindUnknown, RawType: "weird"})
	f.ApplyError(ErrCodeDegraded, "stalled")
	var v map[string]any
	if err := json.Unmarshal(f.Snapshot(), &v); err != nil {
		t.Fatalf("snapshot is not JSON: %v", err)
	}
}

func TestFoldEntryCapEvictsOldestAndFlags(t *testing.T) {
	f := mustFold(t, Limits{MaxEntries: 5})
	for i := 0; i < 20; i++ {
		f.Apply(agent.Event{Kind: agent.KindAgentMessage, MessageID: fmt.Sprintf("m%d", i), Text: "x"})
		if f.Len() > 5 {
			t.Fatalf("entry cap breached: %d", f.Len())
		}
	}
	if f.Len() != 5 || f.EvictedCount() != 15 {
		t.Fatalf("len=%d evicted=%d, want 5/15", f.Len(), f.EvictedCount())
	}
	snap := string(f.Snapshot())
	if !strings.Contains(snap, `"id":"m19"`) || strings.Contains(snap, `"id":"m14"`) || !strings.Contains(snap, `"id":"m15"`) {
		t.Fatalf("oldest not evicted / newest not kept: %s", snap)
	}
	if !strings.Contains(snap, `"truncated":true`) || !strings.Contains(snap, `"evicted_count":15`) {
		t.Fatalf("eviction flags missing: %s", snap)
	}
}

func TestFoldByteCapHoldsUnderOversizedStream(t *testing.T) {
	l := Limits{MaxEntries: 10_000, MaxBytes: 4096, MaxEntryTextBytes: 256}
	f := mustFold(t, l)
	big := strings.Repeat("a", 3000)
	for i := 0; i < 500; i++ {
		f.Apply(agent.Event{Kind: agent.KindAgentMessage, MessageID: fmt.Sprintf("m%d", i), Text: big})
		if n := len(f.Snapshot()); n > l.MaxBytes+SnapshotOverhead {
			t.Fatalf("snapshot %d bytes exceeds cap %d+%d at i=%d", n, l.MaxBytes, SnapshotOverhead, i)
		}
	}
	if f.EvictedCount() == 0 {
		t.Fatal("expected eviction under oversized stream")
	}
	if !strings.Contains(string(f.Snapshot()), `"truncated":true`) {
		t.Fatal("truncated flag not set")
	}
}

func TestFoldDeltaStormOnOneMessageIsBounded(t *testing.T) {
	l := Limits{MaxEntryTextBytes: 256}
	f := mustFold(t, l)
	for i := 0; i < 10_000; i++ {
		f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "m", Text: "0123456789"})
	}
	snap := f.Snapshot()
	if len(snap) > l.withDefaults().MaxBytes+SnapshotOverhead {
		t.Fatalf("snapshot unbounded: %d", len(snap))
	}
	if !strings.Contains(string(snap), `"truncated":true`) {
		t.Fatalf("entry truncation not flagged: %s", snap)
	}
}

func TestFoldSnapshotIsAPointerFreeCopy(t *testing.T) {
	f := mustFold(t, Limits{})
	f.Apply(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m", Text: "one"})
	s1 := f.Snapshot()
	keep := string(s1)
	for i := range s1 {
		s1[i] = 'Z'
	}
	f.Apply(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m2", Text: "two"})
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "m3", Text: "three"})
	if string(f.Snapshot()) == string(s1) {
		t.Fatal("snapshot aliased fold storage")
	}
	s2 := f.Snapshot()
	f.Apply(agent.Event{Kind: agent.KindDelta, MessageID: "m3", Text: "more"})
	if strings.Contains(string(s2), "more") {
		t.Fatal("earlier snapshot changed after later fold write")
	}
	if !strings.HasPrefix(keep, `{"entries"`) {
		t.Fatal("sanity")
	}
}

func TestFoldUnknownAndErrorBecomeEntries(t *testing.T) {
	f := mustFold(t, Limits{})
	f.Apply(agent.Event{Kind: agent.KindUnknown, RawType: "weird"})
	f.ApplyError(ErrCodeDegraded, "stalled")
	if f.Len() != 2 {
		t.Fatalf("len=%d want 2", f.Len())
	}
}

func TestNewFoldRejectsInconsistentLimits(t *testing.T) {
	if _, err := NewFold(Limits{MaxBytes: 100, MaxEntryTextBytes: 1000}); err == nil {
		t.Fatal("expected error when a single entry could exceed the byte cap")
	}
	if _, err := NewFold(Limits{MaxEntries: -1}); err == nil {
		t.Fatal("expected error for negative cap")
	}
}
