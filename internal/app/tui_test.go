package app

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// syncBuf is a goroutine-safe TUI output sink.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func waitOutput(t *testing.T, out *syncBuf, sub string) {
	t.Helper()
	deadline := time.Now().Add(waitFor)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), sub) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("TUI output never contained %q; got:\n%s", sub, out.String())
}

func TestTUIAttachesToHubAsTheSingleLocalClient(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	in, inW := io.Pipe()
	out := &syncBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() }) // release bubbletea's input reader
	o.TUI = TUIOptions{Input: in, Output: out, Width: 80, Height: 24}
	a := startApp(t, o)

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	deadline := time.Now().Add(waitFor)
	for {
		st, err := a.Hub().Stats(ctx)
		if err == nil && st.TUIAttached {
			if st.Remote != 0 {
				t.Errorf("remote clients = %d; want 0", st.Remote)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("the TUI never attached to the hub (stats %+v, err %v)", st, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	f.fake().emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "tui sees this"})
	waitOutput(t, out, "tui sees this")
}
