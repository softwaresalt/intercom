package tui

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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

func statusOf(m Model) string { return normalize(m.statusLine()) }

func TestStatusLineHasNoRejectedUpgradesInitially(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 120, 12)
	if s := statusOf(m); strings.Contains(s, "rejected upgrades") {
		t.Fatalf("status shows a rejection counter before any notice: %q", s)
	}
}

func TestSecurityNoticeRendersCountAndLastReason(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 120, 12)
	m, cmd := step(m, SecurityNoticeMsg{Status: 401, Reason: "credential_rejected", Total: 3})
	if cmd != nil {
		t.Fatalf("a notice must not produce a Cmd")
	}
	want := "rejected upgrades: 3 (last: 401 credential_rejected)"
	if s := statusOf(m); !strings.Contains(s, want) {
		t.Fatalf("status = %q; want it to contain %q", s, want)
	}
	if v := view(m); !strings.Contains(v, want) {
		t.Fatalf("view lacks the notice:\n%s", v)
	}
}

func TestSecurityNoticeOutOfOrderTotalIsIgnored(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 120, 12)
	m, _ = step(m, SecurityNoticeMsg{Status: 403, Reason: "host_not_allowed", Total: 5})
	m, _ = step(m, SecurityNoticeMsg{Status: 401, Reason: "credential_rejected", Total: 2})
	if s := statusOf(m); !strings.Contains(s, "rejected upgrades: 5 (last: 403 host_not_allowed)") {
		t.Fatalf("stale notice replaced a newer one: %q", s)
	}
	m, _ = step(m, SecurityNoticeMsg{Status: 400, Reason: "not_an_upgrade", Total: 5})
	if s := statusOf(m); !strings.Contains(s, "rejected upgrades: 5 (last: 400 not_an_upgrade)") {
		t.Fatalf("an equal Total must be applied: %q", s)
	}
}

func TestSecurityNoticeZeroTotalIsNotRendered(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 120, 12)
	m, _ = step(m, SecurityNoticeMsg{Status: 400, Reason: "x", Total: 0})
	if s := statusOf(m); strings.Contains(s, "rejected upgrades") {
		t.Fatalf("zero total rendered: %q", s)
	}
}

func TestSecurityNoticeReasonIsSanitized(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 200, 12)
	m, _ = step(m, SecurityNoticeMsg{Status: 401, Reason: "bad\x1b\x07 \u202ereason\u009b\nline", Total: 1})
	for _, s := range []string{m.statusLine(), m.View()} {
		for _, bad := range []string{"\x1b[", "\x07", "\u202e", "\u009b"} {
			if strings.Contains(s, bad) {
				t.Fatalf("notice leaks %q: %q", bad, s)
			}
		}
	}
	s := statusOf(m)
	if strings.Contains(s, "\n") {
		t.Fatalf("status line is not a single line: %q", s)
	}
	if !strings.Contains(s, "rejected upgrades: 1 (last: 401 bad<U+001B><U+0007>") || !strings.Contains(s, "<U+202E>") {
		t.Fatalf("hidden characters should be shown as markers: %q", s)
	}
}

func TestFoldDeltaAtCapDoesNotRemarshalWhenAlreadyTruncated(t *testing.T) {
	f, err := hub.NewFold(hub.Limits{MaxEntryTextBytes: 1024, MaxBytes: 16 << 10})
	if err != nil {
		t.Fatal(err)
	}
	f.Apply(delta("m", strings.Repeat("a", 1024)))
	f.Apply(delta("m", "overflow")) // clipped: entry becomes truncated
	ev := delta("m", "more")
	allocs := testing.AllocsPerRun(50, func() { f.Apply(ev) })
	// One allocation remains: the independent copy of the message id. A
	// re-marshal of the entry would add several more.
	if allocs > 1 {
		t.Fatalf("delta at cap on an already-truncated entry allocated %v times; want <= 1 (no re-marshal)", allocs)
	}
}

func TestToolEndedBackfillsNameAndPropagatesTruncationLiveAndHydrated(t *testing.T) {
	r := newRig(t, hub.Config{})
	live := newModel(t, &fakeIngress{}, 60, 12)
	long := strings.Repeat("n", 200)
	r.send(
		agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c1"},
		agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c1", ToolName: "grep", Success: true},
		agent.Event{Kind: agent.KindToolStarted, ToolCallID: "c2", ToolName: "x"},
		agent.Event{Kind: agent.KindToolEnded, ToolCallID: "c2", ToolName: long, Success: true},
	)
	live = r.feedFrames(live, 5)
	r.c.Close()
	r.c = r.connect()
	hydrated := newModel(t, &fakeIngress{}, 60, 12)
	hydrated, _ = step(hydrated, FrameMsg{Frame: r.frame()})
	if len(live.entries) != 2 || len(hydrated.entries) != 2 {
		t.Fatalf("entries live=%d hydrated=%d, want 2", len(live.entries), len(hydrated.entries))
	}
	if live.entries[0].Tool != "grep" || !live.entries[1].Truncated {
		t.Fatalf("live entries wrong: %+v", live.entries)
	}
	for i := range live.entries {
		l, h := live.entries[i], hydrated.entries[i]
		if l.Tool != h.Tool || l.Truncated != h.Truncated {
			t.Fatalf("entry %d live/hydrated mismatch: live=%+v hydrated=%+v", i, l, h)
		}
	}
}

func TestLiveTranscriptCapAndWrapCacheInvalidatesMutatedDeltaEntry(t *testing.T) {
	r := newRig(t, hub.Config{})
	m := New(Options{Renderer: asciiRenderer(), NoApprovalDelay: true, MaxEntries: 3})
	m, _ = step(m, tea.WindowSizeMsg{Width: 40, Height: 12})
	m = r.feedFrames(m, 1)
	r.send(message("m1", "one"), message("m2", "two"), message("m3", "three"), message("m4", "four"), message("m5", "five"))
	m = r.feedFrames(m, 5)
	if len(m.entries) != 3 || m.evicted != 2 || m.entries[0].ID != "m3" {
		t.Fatalf("cap not applied: len=%d evicted=%d entries=%+v", len(m.entries), m.evicted, m.entries)
	}
	if !strings.Contains(view(m), "2 earlier entries evicted") {
		t.Fatalf("eviction not shown:\n%s", view(m))
	}

	r.send(delta("s", "hel"))
	m = r.feedFrames(m, 1)
	last := len(m.entries) - 1
	if m.entries[last].wrapWidth != m.vp.Width || !strings.Contains(m.entries[last].wrapped, "hel") {
		t.Fatalf("rendered entry must be cached: %+v", m.entries[last])
	}
	m.entries[0].wrapped = "SENTINEL-CACHE" // proves an unchanged entry is not re-wrapped
	m.entries[0].wrapWidth = m.vp.Width

	r.send(delta("s", "lo"))
	m = r.feedFrames(m, 1)
	last = len(m.entries) - 1
	if !strings.Contains(m.entries[last].wrapped, "hello") {
		t.Fatalf("mutated delta entry served a stale wrap: %q", m.entries[last].wrapped)
	}
	if m.entries[0].wrapped != "SENTINEL-CACHE" || !strings.Contains(view(m), "hello") {
		t.Fatalf("unchanged entry must keep its cache and the new text must render:\n%s", view(m))
	}

	// A width change re-wraps everything.
	m, _ = step(m, tea.WindowSizeMsg{Width: 50, Height: 12})
	if m.entries[0].wrapped == "SENTINEL-CACHE" {
		t.Fatal("width change must invalidate the cache")
	}
}

func wheel(b tea.MouseButton) tea.MouseMsg {
	return tea.MouseMsg{Button: b, Action: tea.MouseActionPress}
}

func TestLargePromptTruncatesLinearlyWithCorrectTail(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 30, 10)
	m.input = []rune(strings.Repeat("a世", 100000) + "END")
	start := time.Now()
	out := m.View()
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("View with a huge prompt took %v (quadratic truncation?)", d)
	}
	lines := strings.Split(out, "\n")
	last := lines[len(lines)-1]
	if w := lipgloss.Width(last); w > 30 {
		t.Fatalf("prompt is %d columns wide on a 30 column terminal", w)
	}
	if !strings.HasPrefix(last, "> ...") || !strings.HasSuffix(last, "END_") {
		t.Fatalf("unexpected truncated prompt %q", last)
	}
}

func TestWideRunePromptTruncatesToDisplayWidthAndMouseInitAndWheelScrolls(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 30, 10)
	cmd := m.Init()
	if cmd == nil || cmd() != tea.EnableMouseCellMotion() {
		t.Fatal("Init must enable mouse cell-motion reporting")
	}
	m = typeText(m, strings.Repeat("世", 40))
	lines := strings.Split(m.View(), "\n")
	if w := lipgloss.Width(lines[len(lines)-1]); w > 30 {
		t.Fatalf("prompt is %d columns wide on a 30 column terminal", w)
	}

	t.Run("transcript", func(t *testing.T) {
		r := newRig(t, hub.Config{})
		m := newModel(t, &fakeIngress{}, 40, 10)
		m = r.feedFrames(m, 1)
		for i := 0; i < 30; i++ {
			r.send(message("m"+strconv.Itoa(i), "line"))
		}
		m = r.feedFrames(m, 30)
		before := m.vp.YOffset
		m, _ = step(m, wheel(tea.MouseButtonWheelUp))
		if m.vp.YOffset >= before {
			t.Fatalf("wheel up did not scroll the transcript (%d -> %d)", before, m.vp.YOffset)
		}
		up := m.vp.YOffset
		m, _ = step(m, wheel(tea.MouseButtonWheelDown))
		if m.vp.YOffset <= up {
			t.Fatal("wheel down did not scroll the transcript back")
		}
	})

	t.Run("modal disarms and scrolls", func(t *testing.T) {
		m, _, _ := modalModel(t, &fakeIngress{}, agent.PermissionKindShell, longSummary(30), 70, 16)
		m, _ = step(m, wheel(tea.MouseButtonWheelDown))
		if m.modal.YOffset == 0 || m.modalSeen <= m.modal.Height {
			t.Fatalf("wheel did not scroll the modal (offset=%d seen=%d)", m.modal.YOffset, m.modalSeen)
		}
		m = scrollModalToEnd(t, m)
		m, _ = step(m, keyRunes("y"))
		if m.armed == "" {
			t.Fatal("not armed")
		}
		m, _ = step(m, wheel(tea.MouseButtonWheelUp))
		if m.armed != "" {
			t.Fatal("mouse wheel must disarm a pending approval")
		}
	})

	t.Run("credentials", func(t *testing.T) {
		m := New(Options{Renderer: asciiRenderer(), NoApprovalDelay: true, Credentials: credsMsg(fakeApprover, fakeObserver)})
		m, _ = step(m, tea.WindowSizeMsg{Width: 30, Height: 6})
		if !m.credOpen {
			t.Fatal("credential panel should be open")
		}
		m, _ = step(m, wheel(tea.MouseButtonWheelDown))
		if m.credScroll == 0 {
			t.Fatal("wheel did not scroll the credential panel")
		}
	})
}
