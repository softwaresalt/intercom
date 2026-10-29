package tui

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// testWait bounds waits that only ever expire on a failing run: every test
// waits on a condition (channel receive, state poll), never on elapsed time.
const testWait = 20 * time.Second

var fixedClock = func() time.Time { return time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC) }

// fakeIngress records the actions it is asked to perform.
type fakeIngress struct {
	mu     sync.Mutex
	calls  []hub.OperatorAction
	result hub.ActionResult
	err    error
}

func (f *fakeIngress) Do(_ context.Context, act hub.OperatorAction) (hub.ActionResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, act)
	return f.result, f.err
}

func (f *fakeIngress) snapshot() []hub.OperatorAction {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]hub.OperatorAction(nil), f.calls...)
}

func asciiRenderer() *lipgloss.Renderer { return lipgloss.NewRenderer(io.Discard) }

func newModel(t *testing.T, ing Ingress, w, h int) Model {
	t.Helper()
	m := New(Options{Ingress: ing, Renderer: asciiRenderer(), NoApprovalDelay: true})
	m, _ = step(m, tea.WindowSizeMsg{Width: w, Height: h})
	return m
}

// step runs one Update and returns the concrete model.
func step(m Model, msg tea.Msg) (Model, tea.Cmd) {
	next, cmd := m.Update(msg)
	return next.(Model), cmd
}

func keyRunes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func typeText(m Model, s string) Model {
	for _, r := range s {
		m, _ = step(m, keyRunes(string(r)))
	}
	return m
}

// rig is a real hub plus an attached TUI client, used as the generator of
// genuine frames (Frame has no public constructor) and as the real ingress.
type rig struct {
	t *testing.T
	h *hub.Hub
	c *hub.Client
}

func newRig(t *testing.T, cfg hub.Config) *rig {
	t.Helper()
	cfg.SessionID = "sess"
	cfg.Clock = fixedClock
	if cfg.Now == nil {
		// Pin the hub's own clock too: remaining_ms (and so the modal's
		// 'expires:' line) derives from it, and must not depend on wall time.
		cfg.Now = fixedClock
	}
	h, err := hub.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- h.Run(context.Background()) }()
	t.Cleanup(func() {
		h.Stop()
		select {
		case <-done:
		case <-time.After(testWait):
			t.Error("hub Run did not return after Stop")
		}
	})
	r := &rig{t: t, h: h}
	r.c = r.connect()
	return r
}

func (r *rig) connect() *hub.Client {
	r.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	c, err := r.h.Connect(ctx, hub.ConnectOptions{Kind: hub.ClientTUI})
	if err != nil {
		r.t.Fatalf("connect: %v", err)
	}
	return c
}

func (r *rig) send(evs ...agent.Event) {
	r.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	for _, ev := range evs {
		if err := r.h.Send(ctx, ev); err != nil {
			r.t.Fatalf("hub send: %v", err)
		}
	}
}

func (r *rig) frame() hub.Frame { return recvFrame(r.t, r.c) }

func recvFrame(t *testing.T, c *hub.Client) hub.Frame {
	t.Helper()
	select {
	case f := <-c.Frames():
		return f
	case <-time.After(testWait):
		t.Fatal("timed out waiting for a frame")
		return hub.Frame{}
	}
}

// feedFrames receives n frames and applies them to m.
func (r *rig) feedFrames(m Model, n int) Model {
	r.t.Helper()
	for i := 0; i < n; i++ {
		m, _ = step(m, FrameMsg{Frame: r.frame()})
	}
	return m
}

func (r *rig) untilKind(m Model, kind hub.Kind) Model {
	r.t.Helper()
	for {
		f := r.frame()
		m, _ = step(m, FrameMsg{Frame: f})
		if f.Kind() == kind {
			return m
		}
	}
}

func delta(id, text string) agent.Event {
	return agent.Event{Kind: agent.KindDelta, MessageID: id, Text: text}
}

func message(id, text string) agent.Event {
	return agent.Event{Kind: agent.KindAgentMessage, MessageID: id, Text: text}
}

// view returns the rendered frame with trailing spaces trimmed per line.
func view(m Model) string { return normalize(m.View()) }

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

// golden compares got with testdata/<name>.golden; UPDATE_GOLDEN=1 rewrites it.
func golden(t *testing.T, name, got string) {
	t.Helper()
	path := "testdata/" + name + ".golden"
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path) //nolint:gosec // fixed testdata path
	if err != nil {
		t.Fatalf("read golden (run with UPDATE_GOLDEN=1 to create): %v", err)
	}
	if w := strings.ReplaceAll(string(want), "\r\n", "\n"); w != got+"\n" {
		t.Fatalf("golden %s mismatch\n--- want ---\n%s\n--- got ---\n%s", name, w, got+"\n")
	}
}
