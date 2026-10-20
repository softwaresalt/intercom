package tui

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

func TestUpdatePerformsNoIngressCallAndOnlyReturnsCmd(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m := newModel(t, ing, 80, 24)
	m = typeText(m, "hello")

	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter on a prompt must return a Cmd")
	}
	if got := ing.snapshot(); len(got) != 0 {
		t.Fatalf("Update itself called the ingress (%d calls); side effects must live in the Cmd", len(got))
	}
	res := cmd()
	calls := ing.snapshot()
	if len(calls) != 1 || calls[0].Type != hub.OpPrompt || calls[0].Prompt != "hello" {
		t.Fatalf("cmd should submit exactly one OpPrompt through the ingress, got %+v", calls)
	}
	m, _ = step(m, res)
	if len(m.input) != 0 {
		t.Fatalf("accepted prompt should clear the input, got %q", string(m.input))
	}
}

func TestAbortGoesThroughIngressOnly(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m := newModel(t, ing, 80, 24)
	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlX})
	if cmd == nil || len(ing.snapshot()) != 0 {
		t.Fatal("abort must be a Cmd, not a direct call")
	}
	cmd()
	if c := ing.snapshot(); len(c) != 1 || c[0].Type != hub.OpAbortTurn {
		t.Fatalf("want one OpAbortTurn, got %+v", c)
	}
}

func TestTypedInputIsSentVerbatimButFiltered(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m := newModel(t, ing, 80, 24)
	emoji := "\U0001F469\u200D\U0001F4BB"
	m = typeText(m, "hi "+emoji+"\x1b\x07\u009b!")
	if got := string(m.input); got != "hi "+emoji+"!" {
		t.Fatalf("input = %q, want format runes kept and controls dropped", got)
	}
	v := m.View()
	for _, bad := range []string{"\x1b", "\x07", "\u009b", "\u200d"} {
		if strings.Contains(v, bad) {
			t.Fatalf("view leaks %q: %q", bad, v)
		}
	}
	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	cmd()
	c := ing.snapshot()
	if len(c) != 1 || c[0].Prompt != "hi "+emoji+"!" {
		t.Fatalf("sent %+v, want the typed text verbatim", c)
	}
}

func TestRefusedPromptKeepsInputAndShowsReason(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Reason: hub.ReasonAgentUnavailable}}
	m := newModel(t, ing, 80, 24)
	m = typeText(m, "hi")
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = step(m, cmd())
	if string(m.input) != "hi" {
		t.Fatalf("refused prompt must keep the text, got %q", string(m.input))
	}
	if !strings.Contains(view(m), string(hub.ReasonAgentUnavailable)) {
		t.Fatalf("refusal reason not shown:\n%s", view(m))
	}
}

func TestNoIngressYieldsErrorResultNotPanic(t *testing.T) {
	m := newModel(t, nil, 80, 24)
	m = typeText(m, "x")
	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	msg, ok := cmd().(ActionResultMsg)
	if !ok || msg.Err == nil {
		t.Fatalf("want ActionResultMsg with error, got %#v", msg)
	}
}

// TestPackageHasNoDirectAgentOrSDKReach proves structurally that the TUI
// package cannot call session.Send or reach the Copilot SDK: its non-test
// sources import neither the SDK nor the adapter.
func TestPackageHasNoDirectAgentOrSDKReach(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f) //nolint:gosec // package-local source files
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(fset, f, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range af.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if strings.Contains(p, "copilot-sdk") || strings.HasSuffix(p, "internal/agent/copilot") {
				t.Errorf("%s imports %s: the TUI must act only through hub ingress", f, p)
			}
		}
	}
}

// ---- golden frames ----

func TestGoldenEmpty(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	m = r.feedFrames(m, 1)
	golden(t, "empty", view(m))
}

func TestGoldenStreaming(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	r.send(
		agent.Event{Kind: agent.KindTurnStart, TurnID: "t1"},
		message("m1", "Done reading the repository."),
		agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c1", ToolName: "grep"},
		agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c1", ToolName: "grep", Success: true},
		delta("m2", "I will now update the "),
		delta("m2", "parser so that streaming text wraps at the viewport width and keeps going"),
	)
	m = r.feedFrames(m, 7) // hydrate + 6 live envelopes
	golden(t, "streaming", view(m))
}

func TestGoldenError(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	m = r.feedFrames(m, 1)
	r.h.SignalDegraded("adapter fault")
	m = r.untilKind(m, hub.KindSessionHydrate)
	golden(t, "error_degraded", view(m))

	m, _ = step(m, DisconnectedMsg{Err: hub.ErrStopped})
	golden(t, "error_disconnected", view(m))
}

// ---- transcript viewport ----

func fillTranscript(r *rig, n int) {
	evs := make([]agent.Event, 0, n)
	for i := 0; i < n; i++ {
		evs = append(evs, message("m"+strconv.Itoa(i), "line "+strconv.Itoa(i)))
	}
	r.send(evs...)
}

func TestStreamingDeltasKeepScrollPosition(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	fillTranscript(r, 60)
	r.send(delta("s", "streaming start"))
	m = r.feedFrames(m, 62)
	if !m.vp.AtBottom() {
		t.Fatal("a viewport at the bottom must follow new output")
	}

	m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgUp})
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgUp})
	off := m.vp.YOffset
	if off == 0 || m.vp.AtBottom() {
		t.Fatalf("setup: expected to be scrolled up, offset=%d", off)
	}
	r.send(delta("s", " and more"), delta("s", " and more"), message("late", "new entry"))
	m = r.feedFrames(m, 3)
	if m.vp.YOffset != off {
		t.Fatalf("scroll position lost: offset %d -> %d", off, m.vp.YOffset)
	}
	if got := m.entries[len(m.entries)-2].Text; got != "streaming start and more and more" {
		t.Fatalf("delta not appended to the streaming entry: %q", got)
	}
	if m.vp.AtBottom() {
		t.Fatal("must not jump to the bottom while the operator is scrolled up")
	}

	m, _ = step(m, tea.KeyMsg{Type: tea.KeyEnd})
	r.send(delta("s2", "tail"))
	m = r.feedFrames(m, 1)
	if !m.vp.AtBottom() {
		t.Fatal("after returning to the bottom the view must follow again")
	}
}

func TestRehydrateReplacesTranscriptWithoutDuplicates(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	r.send(message("a", "alpha"), message("b", "beta"), delta("c", "gam"))
	m = r.feedFrames(m, 4)
	before := len(m.entries)
	if before != 3 {
		t.Fatalf("setup: %d entries", before)
	}

	// A duplicate live frame (same seq) must be discarded, not re-applied.
	r.send(delta("c", "ma"))
	f := r.frame()
	m, _ = step(m, FrameMsg{Frame: f})
	m, _ = step(m, FrameMsg{Frame: f})
	if got := m.entries[2].Text; got != "gamma" {
		t.Fatalf("duplicate frame re-applied: %q", got)
	}

	// Re-hydrate through a fresh client: same state, replaced not appended.
	r.c.Close()
	r.c = r.connect()
	hyd := r.frame()
	for i := 0; i < 2; i++ { // applying the same hydrate twice is idempotent too
		m, _ = step(m, FrameMsg{Frame: hyd})
		if len(m.entries) != before {
			t.Fatalf("hydrate #%d: %d entries, want %d (duplicates)", i+1, len(m.entries), before)
		}
	}
	seen := map[string]bool{}
	for _, e := range m.entries {
		if seen[e.ID] {
			t.Fatalf("duplicate entry id %q", e.ID)
		}
		seen[e.ID] = true
	}
	if m.entries[2].Text != "gamma" || m.entries[2].Status != "streaming" {
		t.Fatalf("hydrated entry wrong: %+v", m.entries[2])
	}

	// And the stream continues cleanly after the hydrate.
	r.send(message("c", "gamma done"))
	m = r.feedFrames(m, 1)
	if len(m.entries) != 3 || m.entries[2].Status != "complete" {
		t.Fatalf("post-hydrate live frame mis-applied: %+v", m.entries)
	}
}

func TestGapForcesWaitForHydrate(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 60, 12)
	m = r.feedFrames(m, 1)
	r.send(message("a", "one"), message("b", "two"), message("c", "three"))
	_ = r.frame() // seq of "a" is lost on purpose
	m, _ = step(m, FrameMsg{Frame: r.frame()})
	m, _ = step(m, FrameMsg{Frame: r.frame()})
	if len(m.entries) != 0 {
		t.Fatalf("frames after a gap must not be applied: %+v", m.entries)
	}
	if !strings.Contains(view(m), "resync") {
		t.Fatalf("gap not visible in status:\n%s", view(m))
	}
}

func TestStatusShowsDegradedAndShutdown(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 80, 12)
	m = r.feedFrames(m, 1)
	r.h.SignalDegraded("x")
	m = r.untilKind(m, hub.KindSessionHydrate)
	if !strings.Contains(view(m), "DEGRADED") {
		t.Fatalf("degraded not visible:\n%s", view(m))
	}
	m, _ = step(m, ResyncMsg{})
	if !strings.Contains(view(m), "resync") {
		t.Fatalf("resync not visible:\n%s", view(m))
	}
}

func TestTranscriptTextIsSanitized(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := newModel(t, &fakeIngress{}, 80, 12)
	r.send(
		message("m\x1b[2J", "hello \x1b]0;owned\x07 \u202eevil \u009b2J"),
		agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c", ToolName: "ba\x1bsh\u2066"},
	)
	m = r.feedFrames(m, 3)
	v := m.View()
	for _, bad := range []string{"\x1b", "\x07", "\u202e", "\u009b", "\u2066"} {
		if strings.Contains(v, bad) {
			t.Fatalf("view leaks %q:\n%q", bad, v)
		}
	}
	if !strings.Contains(v, "<U+202E>") {
		t.Fatalf("hidden characters should be shown as markers:\n%s", v)
	}
}
