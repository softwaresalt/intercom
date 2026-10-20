package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// waitFor bounds every wait: it expires only on a failing run.
const waitFor = 20 * time.Second

// recorder is a goroutine-safe ordered event log shared by fakes.
type recorder struct {
	mu     sync.Mutex
	events []string
}

func (r *recorder) add(e string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

// fakeAgent is an injected agent.Agent: no CLI child, every call recorded.
type fakeAgent struct {
	rec  *recorder
	deps AgentDeps

	startErr error
	// stopHook runs inside Stop (before it returns).
	stopHook func(context.Context)

	mu     sync.Mutex
	starts int
	stops  int
	sends  []string
}

func (f *fakeAgent) Start(context.Context) (agent.SessionHandle, error) {
	f.mu.Lock()
	f.starts++
	f.mu.Unlock()
	f.rec.add("agent.start")
	if f.startErr != nil {
		return agent.SessionHandle{}, f.startErr
	}
	return agent.SessionHandle{ID: "fake-session"}, nil
}

func (f *fakeAgent) Send(_ context.Context, p string) error {
	f.mu.Lock()
	f.sends = append(f.sends, p)
	f.mu.Unlock()
	return nil
}

func (f *fakeAgent) Abort(context.Context) error {
	f.rec.add("agent.abort")
	return agent.ErrNotRunning
}

func (f *fakeAgent) Stop(ctx context.Context) error {
	f.mu.Lock()
	f.stops++
	f.mu.Unlock()
	f.rec.add("agent.stop")
	if f.stopHook != nil {
		f.stopHook(ctx)
	}
	return nil
}

func (f *fakeAgent) counts() (starts, stops int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.starts, f.stops
}

// emit pushes an agent event into the hub through the injected ingress, the
// way the real adapter's callback does.
func (f *fakeAgent) emit(ev agent.Event) { f.deps.Ingress <- ev }

// factory records how many times it was invoked and hands out one fakeAgent.
type factory struct {
	mu       sync.Mutex
	calls    int
	agent    *fakeAgent
	rec      *recorder
	startErr error
	buildErr error
}

func newFactory() *factory { return &factory{rec: &recorder{}} }

func (f *factory) build(d AgentDeps) (agent.Agent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.buildErr != nil {
		return nil, f.buildErr
	}
	f.agent = &fakeAgent{rec: f.rec, deps: d, startErr: f.startErr}
	return f.agent, nil
}

func (f *factory) invoked() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *factory) fake() *fakeAgent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.agent
}

// writeConfig writes a config whose default_workspace_root is root. top holds
// extra top-level keys; tables holds extra table sections.
func writeConfig(t *testing.T, root, top, tables string) string {
	t.Helper()
	body := fmt.Sprintf("default_workspace_root = '%s'\n%s\n%s\n", root, top, tables)
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// freePort returns a loopback TCP port that was free a moment ago.
func freePort(t *testing.T) uint16 {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	return uint16(ln.Addr().(*net.TCPAddr).Port) //nolint:gosec // a TCP port fits in 16 bits
}

// logBuf is a goroutine-safe log sink.
type logBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *logBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *logBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func newTestLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// baseOptions is a headless option set over a fresh workspace: ephemeral
// port, no TUI, injected fake agent.
func baseOptions(t *testing.T, f *factory) (Options, string) {
	t.Helper()
	ws := t.TempDir()
	return Options{
		ConfigPath: writeConfig(t, ws, "", ""),
		Port:       0,
		PortSet:    true,
		NewAgent:   f.build,
		TUI:        TUIOptions{Disabled: true},
	}, ws
}

// startApp starts an app and registers a forced-teardown cleanup.
func startApp(t *testing.T, o Options) *App {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	a, err := Start(ctx, o)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		sctx, scancel := context.WithTimeout(context.Background(), waitFor)
		defer scancel()
		_ = a.Stop(sctx)
	})
	return a
}
