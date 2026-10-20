package copilot

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// probeClient wraps the real SDK client and instruments the two callbacks the
// adapter registers at CreateSession, the same measurement the retired C2
// spike took (design 4.2 SQ-b, design 4.4 SQ-c; docs/decisions/
// 2026-09-04-intercom-go-c2-sdk-spike-findings.md):
//
//	SQ-b  OnEvent invocations are serialised: the in-flight count never exceeds 1
//	SQ-c  permission dispatch is independent of event dispatch: events keep
//	      arriving while a permission handler is parked
type probeClient struct {
	realClient

	inflight    atomic.Int64
	maxInflight atomic.Int64
	events      atomic.Int64
	permission  atomic.Int64 // permission handler entries
}

func (p *probeClient) CreateSession(ctx context.Context, cfg *copilotsdk.SessionConfig) (sdkSession, error) {
	onEvent, onPerm := cfg.OnEvent, cfg.OnPermissionRequest
	cfg.OnEvent = func(ev copilotsdk.SessionEvent) {
		n := p.inflight.Add(1)
		defer p.inflight.Add(-1)
		for m := p.maxInflight.Load(); n > m; m = p.maxInflight.Load() {
			if p.maxInflight.CompareAndSwap(m, n) {
				break
			}
		}
		p.events.Add(1)
		onEvent(ev)
	}
	cfg.OnPermissionRequest = func(req copilotsdk.PermissionRequest, inv copilotsdk.PermissionInvocation) (rpc.PermissionDecision, error) {
		p.permission.Add(1)
		return onPerm(req, inv)
	}
	return p.realClient.CreateSession(ctx, cfg)
}

// TestLiveSQbSQcReverification re-verifies, against the pinned SDK and a real
// runtime, the two empirical findings the adapter relies on. Opt-in with
// INTERCOM_LIVE_SDK_TESTS (see liveSDKTestsAllowed); it never approves
// anything: the one permission request it receives is held for a while and
// then DENIED.
//
//	INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/agent/copilot -run LiveSQbSQc -v
//
// A failure here is a signal to revert to the conservative postures in design
// 4.2 (guard callback-local state) and 4.4 (assume head-of-line blocking).
func TestLiveSQbSQcReverification(t *testing.T) {
	if !liveSDKTestsAllowed(os.Getenv(liveSDKTestsEnvVar)) {
		t.Skipf("%s not enabled; skipping live SDK SQ-b/SQ-c re-verification", liveSDKTestsEnvVar)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("intercom live test notes\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	hubDone := make(chan struct{})
	var closeHub sync.Once
	defer closeHub.Do(func() { close(hubDone) })
	ingress := make(chan agent.Event, 256)
	asks := make(chan agent.PermissionAsk, 1)
	// INTERCOM_LIVE_CLI_PATH (docs/runbooks/live-sdk-tests.md) overrides the PATH
	// lookup, for both the adapter's validation and the SDK's spawn.
	cliPath := os.Getenv("INTERCOM_LIVE_CLI_PATH")
	cfg := Config{
		CLIPath:          cliPath,
		WorkingDirectory: dir,
		StartupTimeout:   120 * time.Second,
		StopTimeout:      10 * time.Second,
		Ingress:          ingress,
		Asks:             asks,
		HubDone:          hubDone,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	opts := &copilotsdk.ClientOptions{WorkingDirectory: dir}
	if cliPath != "" {
		opts.Connection = copilotsdk.StdioConnection{Path: cliPath}
	}
	probe := &probeClient{realClient: realClient{c: copilotsdk.NewClient(opts)}}
	ad, err := newWithClient(cfg, probe)
	if err != nil {
		t.Fatalf("newWithClient: %v", err)
	}

	// Drain the ingress and note when the turn goes idle.
	idle := make(chan struct{})
	var idleOnce sync.Once
	go func() {
		for {
			select {
			case ev := <-ingress:
				if ev.Kind == agent.KindIdle {
					idleOnce.Do(func() { close(idle) })
				}
			case <-hubDone:
				return
			}
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err := ad.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		sctx, scancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer scancel()
		if err := ad.Stop(sctx); err != nil {
			t.Errorf("Stop: %v", err)
		}
	}()

	prompt := os.Getenv("INTERCOM_LIVE_PROMPT")
	if prompt == "" {
		prompt = "Use your file reading tool to read the file notes.txt in the current directory, then reply with only the word DONE."
	}
	if err := ad.Send(ctx, prompt); err != nil {
		t.Fatalf("Send: %v", err)
	}

	var ask agent.PermissionAsk
	select {
	case ask = <-asks:
	case <-time.After(3 * time.Minute):
		t.Fatalf("no permission request arrived (events seen: %d); the runtime did not ask for the read, so SQ-c could not be exercised (set INTERCOM_LIVE_PROMPT to a prompt that triggers a permission request)", probe.events.Load())
	}

	// SQ-c: the handler is now parked on our reply. Hold it and watch whether
	// events keep being dispatched meanwhile.
	atAsk := probe.events.Load()
	hold := time.NewTimer(30 * time.Second)
	defer hold.Stop()
	poll := time.NewTicker(100 * time.Millisecond)
	defer poll.Stop()
	observed := false
wait:
	for !observed {
		select {
		case <-poll.C:
			observed = probe.events.Load() > atAsk
		case <-hold.C:
			break wait
		}
	}
	if !observed {
		t.Errorf("SQ-c not re-verified: no session event was dispatched in 30s while a permission handler was parked (events at the request: %d). Either the runtime had nothing more to say while waiting, or dispatch is head-of-line blocked; inspect manually before trusting design 4.4", atAsk)
	}
	t.Logf("SQ-c: %d session event(s) dispatched while the permission handler was parked", probe.events.Load()-atAsk)

	// Never approve in a live test.
	ask.Reply <- agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "intercom live test: denied"}

	select {
	case <-idle:
	case <-time.After(3 * time.Minute):
		t.Fatal("the turn never went idle after the permission was denied")
	}

	// SQ-b: across the whole turn no two OnEvent invocations overlapped.
	if got := probe.maxInflight.Load(); got != 1 {
		t.Errorf("SQ-b: max concurrent OnEvent invocations = %d over %d events; want exactly 1 (serialised)", got, probe.events.Load())
	}
	if probe.events.Load() < 2 {
		t.Errorf("SQ-b: only %d event(s) sampled; too few to say anything", probe.events.Load())
	}
	t.Logf("SQ-b: max in-flight %d across %d events (%d permission handler entries)", probe.maxInflight.Load(), probe.events.Load(), probe.permission.Load())
}
