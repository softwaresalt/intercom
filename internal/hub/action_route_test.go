package hub

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---- 067.005-T: prompt-send and turn-abort through the hub ----

// fakeAgent records the calls the hub makes on behalf of operators.
type fakeAgent struct {
	mu         sync.Mutex
	prompts    []string
	aborts     int
	sendCh     chan string   // a prompt, as it enters Send
	abortCh    chan struct{} // an abort, as it enters Abort
	block      chan struct{} // when non-nil, Send blocks until closed or its ctx ends
	abortBlock chan struct{} // when non-nil, Abort blocks until closed or its ctx ends
	sendErr    error
	sendCtxE   chan error // Send's ctx error when it returns because of it
}

func newFakeAgent() *fakeAgent {
	return &fakeAgent{sendCh: make(chan string, 64), abortCh: make(chan struct{}, 64), sendCtxE: make(chan error, 8)}
}

func (a *fakeAgent) Send(ctx context.Context, prompt string) error {
	a.mu.Lock()
	a.prompts = append(a.prompts, prompt)
	block, err := a.block, a.sendErr
	a.mu.Unlock()
	a.sendCh <- prompt
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			a.sendCtxE <- ctx.Err()
			return ctx.Err()
		}
	}
	return err
}

func (a *fakeAgent) Abort(ctx context.Context) error {
	a.mu.Lock()
	a.aborts++
	block := a.abortBlock
	a.mu.Unlock()
	a.abortCh <- struct{}{}
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (a *fakeAgent) promptsSeen() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.prompts...)
}

func awaitPrompt(t *testing.T, a *fakeAgent) string {
	t.Helper()
	select {
	case p := <-a.sendCh:
		return p
	case <-time.After(testWait):
		t.Fatal("the agent never received the prompt")
		return ""
	}
}

func awaitAbort(t *testing.T, a *fakeAgent) {
	t.Helper()
	select {
	case <-a.abortCh:
	case <-time.After(testWait):
		t.Fatal("the agent never received the abort")
	}
}

func prompt(text string) OperatorAction { return OperatorAction{Type: OpPrompt, Prompt: text} }

func TestObserverPromptAndAbortAreRefusedAndNeverReachTheAgent(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	obs := connectRole(t, h, ClientRemote, RoleObserver)
	tui := connectRole(t, h, ClientTUI, 0)

	for _, act := range []OperatorAction{prompt("rm everything"), {Type: OpAbortTurn}} {
		res := do(t, obs, act)
		if res.Accepted || res.Reason != ReasonRoleNotPermitted {
			t.Fatalf("observer %v = %+v, want role_not_permitted", act.Type, res)
		}
	}
	// A later authorized prompt is the only thing the agent ever sees.
	if res := do(t, tui, prompt("legit")); !res.Accepted {
		t.Fatalf("tui prompt = %+v", res)
	}
	if got := awaitPrompt(t, ag); got != "legit" {
		t.Fatalf("agent saw %q first: a refused prompt leaked through", got)
	}
	if n := len(ag.promptsSeen()); n != 1 {
		t.Fatalf("agent saw %d prompts, want 1", n)
	}
	select {
	case <-ag.abortCh:
		t.Fatal("a refused abort reached the agent")
	default:
	}
}

func TestApproverPromptReachesTheAgentExactlyOnce(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag}) // remote_approve_destructive stays off: steering is allowed
	appr := connectRole(t, h, ClientRemote, RoleApprover)

	if res := do(t, appr, prompt("first")); !res.Accepted {
		t.Fatalf("approver prompt = %+v", res)
	}
	if got := awaitPrompt(t, ag); got != "first" {
		t.Fatalf("agent got %q", got)
	}
	// FIFO on one lane: a sentinel arriving next proves "first" was not
	// delivered twice.
	do(t, appr, prompt("sentinel"))
	if got := awaitPrompt(t, ag); got != "sentinel" {
		t.Fatalf("second delivery was %q: the prompt was duplicated", got)
	}
}

func TestApproverAndTUIAbortReachTheAgent(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	tui := connectRole(t, h, ClientTUI, 0)
	for _, c := range []*Client{appr, tui} {
		if res := do(t, c, OperatorAction{Type: OpAbortTurn}); !res.Accepted {
			t.Fatalf("abort = %+v", res)
		}
		awaitAbort(t, ag)
	}
}

func TestHubNeverCallsTheAdapterSynchronously(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{})
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)

	if res := do(t, tui, prompt("slow")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag) // Send is now blocked inside the adapter

	// The hub is not parked in that call: it still serves hub steps, folds and
	// broadcasts events, and an abort overtakes the blocked prompt.
	stats(t, h)
	send(t, h, msgEvent("m1", "still flowing"))
	recvKind(t, tui, KindAgentMessage)
	if res := do(t, tui, OperatorAction{Type: OpAbortTurn}); !res.Accepted {
		t.Fatalf("abort = %+v", res)
	}
	awaitAbort(t, ag)
	close(ag.block)
}

func TestPromptQueueFullIsRefusedNotBlocked(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{})
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)

	do(t, tui, prompt("in flight"))
	awaitPrompt(t, ag)
	for i := 0; i < operatorPromptQueue; i++ {
		if res := do(t, tui, prompt("queued")); !res.Accepted {
			t.Fatalf("queued prompt %d = %+v", i, res)
		}
	}
	res := do(t, tui, prompt("one too many"))
	if res.Accepted || res.Reason != ReasonAgentUnavailable {
		t.Fatalf("overflow prompt = %+v, want agent_unavailable", res)
	}
	close(ag.block)
}

func TestNoAgentAttachedRefusesWithAgentUnavailable(t *testing.T) {
	h := startHub(t, Config{})
	tui := connectRole(t, h, ClientTUI, 0)
	for _, act := range []OperatorAction{prompt("x"), {Type: OpAbortTurn}} {
		if res := do(t, tui, act); res.Accepted || res.Reason != ReasonAgentUnavailable {
			t.Fatalf("%v with no agent = %+v", act.Type, res)
		}
	}
}

func TestEmptyPromptIsMalformed(t *testing.T) {
	h := startHub(t, Config{Agent: newFakeAgent()})
	tui := connectRole(t, h, ClientTUI, 0)
	if res := do(t, tui, prompt("  \n")); res.Accepted || res.Reason != ReasonActionUnknown {
		t.Fatalf("blank prompt = %+v", res)
	}
}

func TestAgentCallIsBoundedByTheOperatorTimeout(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{}) // never released: only the deadline can end Send
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: 30 * time.Millisecond})
	tui := connectRole(t, h, ClientTUI, 0)
	do(t, tui, prompt("hangs"))
	select {
	case err := <-ag.sendCtxE:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Send ended with %v, want the operator deadline", err)
		}
	case <-time.After(testWait):
		t.Fatal("a hung agent call was never bounded")
	}
}

func TestHubStopCancelsAnInFlightAgentCall(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{})
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	do(t, tui, prompt("hangs"))
	awaitPrompt(t, ag)
	h.Stop()
	select {
	case err := <-ag.sendCtxE:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Send ended with %v, want cancellation", err)
		}
	case <-time.After(testWait):
		t.Fatal("hub stop left an operator call running")
	}
}

func TestAgentFailureIsLoggedWithoutThePromptText(t *testing.T) {
	var buf bytes.Buffer
	var mu sync.Mutex
	lg := slog.New(slog.NewTextHandler(&lockedWriter{mu: &mu, w: &buf}, nil))
	ag := newFakeAgent()
	ag.sendErr = errors.New("sdk exploded")
	h := startHub(t, Config{Agent: ag, Logger: lg})
	tui := connectRole(t, h, ClientTUI, 0)
	do(t, tui, prompt("super secret prompt body"))
	awaitPrompt(t, ag)
	deadline := time.Now().Add(testWait)
	for {
		mu.Lock()
		out := buf.String()
		mu.Unlock()
		if strings.Contains(out, "sdk exploded") {
			if strings.Contains(out, "super secret prompt body") {
				t.Fatalf("the prompt text was logged: %s", out)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("the agent error was swallowed; log = %q", out)
		}
		time.Sleep(time.Millisecond)
	}
}

type lockedWriter struct {
	mu *sync.Mutex
	w  *bytes.Buffer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}
