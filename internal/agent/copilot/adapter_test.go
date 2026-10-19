package copilot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

var _ agent.Agent = (*Adapter)(nil)

func absTestPath() string { return filepath.Join(os.TempDir(), "copilot") }

// ---- fake SDK seam ----

type fakeSession struct {
	mu                  sync.Mutex
	handler             copilotsdk.SessionEventHandler
	unsubscribe         atomic.Int32
	sendCtxHasDeadline  bool
	sentPrompt          string
	sendErr             error
	abortErr            error
	abortCtxHasDeadline bool
	disconnects         atomic.Int32
}

func (s *fakeSession) ID() string { return "sess-1" }
func (s *fakeSession) On(h copilotsdk.SessionEventHandler) func() {
	s.mu.Lock()
	s.handler = h
	s.mu.Unlock()
	return func() { s.unsubscribe.Add(1) }
}
func (s *fakeSession) emit(ev copilotsdk.SessionEvent) {
	s.mu.Lock()
	h := s.handler
	s.mu.Unlock()
	h(ev)
}
func (s *fakeSession) Send(ctx context.Context, o copilotsdk.MessageOptions) (string, error) {
	_, s.sendCtxHasDeadline = ctx.Deadline()
	s.sentPrompt = o.Prompt
	return "m", s.sendErr
}
func (s *fakeSession) Abort(ctx context.Context) error {
	_, s.abortCtxHasDeadline = ctx.Deadline()
	return s.abortErr
}
func (s *fakeSession) Disconnect() error { s.disconnects.Add(1); return nil }

type fakeClient struct {
	startBlock  bool
	startErr    error
	status      *copilotsdk.GetStatusResponse
	statusErr   error
	session     *fakeSession
	createCfg   *copilotsdk.SessionConfig
	createCalls atomic.Int32
	stopErr     error
	stops       atomic.Int32
	forceStops  atomic.Int32
}

func newFakeClient() *fakeClient {
	return &fakeClient{
		status:  &copilotsdk.GetStatusResponse{Version: ValidatedCLIVersion, ProtocolVersion: copilotsdk.SDKProtocolVersion},
		session: &fakeSession{},
	}
}

func (c *fakeClient) Start(ctx context.Context) error {
	if c.startBlock {
		<-ctx.Done()
		return ctx.Err()
	}
	return c.startErr
}
func (c *fakeClient) GetStatus(context.Context) (*copilotsdk.GetStatusResponse, error) {
	return c.status, c.statusErr
}
func (c *fakeClient) CreateSession(_ context.Context, cfg *copilotsdk.SessionConfig) (sdkSession, error) {
	c.createCalls.Add(1)
	c.createCfg = cfg
	return c.session, nil
}
func (c *fakeClient) Stop() error { c.stops.Add(1); return c.stopErr }
func (c *fakeClient) ForceStop()  { c.forceStops.Add(1) }

type harness struct {
	ad      *Adapter
	client  *fakeClient
	ingress chan agent.Event
	asks    chan agent.PermissionAsk
	hubDone chan struct{}
}

func newHarness(t *testing.T, mutate func(*Config, *fakeClient)) *harness {
	t.Helper()
	h := &harness{
		client:  newFakeClient(),
		ingress: make(chan agent.Event, 8),
		asks:    make(chan agent.PermissionAsk, 1),
		hubDone: make(chan struct{}),
	}
	cfg := Config{
		WorkingDirectory: t.TempDir(),
		StartupTimeout:   time.Second,
		StopTimeout:      time.Second,
		Ingress:          h.ingress,
		Asks:             h.asks,
		HubDone:          h.hubDone,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	if mutate != nil {
		mutate(&cfg, h.client)
	}
	ad, err := newWithClient(cfg, h.client)
	if err != nil {
		t.Fatalf("newWithClient: %v", err)
	}
	h.ad = ad
	return h
}

// ---- 065.004-T: Start ----

func TestStartCreatesSingleSessionAndConfiguresIt(t *testing.T) {
	h := newHarness(t, nil)
	handle, err := h.ad.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if handle.ID != "sess-1" {
		t.Errorf("handle = %+v", handle)
	}
	cfg := h.client.createCfg
	if cfg == nil || cfg.OnPermissionRequest == nil {
		t.Fatal("session must be created with a permission handler")
	}
	if cfg.Streaming == nil || !*cfg.Streaming {
		t.Error("streaming must be enabled so deltas flow")
	}
	if cfg.WorkingDirectory == "" {
		t.Error("working directory must be passed to the session")
	}
	if _, err := h.ad.Start(context.Background()); err == nil {
		t.Fatal("second Start must fail: one session per process (ADR MVP-D7)")
	}
	if h.client.createCalls.Load() != 1 {
		t.Errorf("CreateSession called %d times, want 1", h.client.createCalls.Load())
	}
}

func TestStartHonoursConfiguredDeadline(t *testing.T) {
	h := newHarness(t, func(c *Config, f *fakeClient) {
		c.StartupTimeout = 50 * time.Millisecond
		f.startBlock = true
	})
	start := time.Now()
	_, err := h.ad.Start(context.Background())
	if err == nil {
		t.Fatal("Start must surface the deadline as an error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error should wrap DeadlineExceeded, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Error("Start overran its configured deadline")
	}
	if h.client.forceStops.Load() == 0 {
		t.Error("a failed Start must tear the client down")
	}
}

func TestStartSurfacesClientErrors(t *testing.T) {
	h := newHarness(t, func(_ *Config, f *fakeClient) { f.startErr = errors.New("boom") })
	if _, err := h.ad.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("Start error = %v, want wrapped boom", err)
	}
}

func TestStartRejectsVersionMismatch(t *testing.T) {
	cases := []struct {
		name    string
		status  copilotsdk.GetStatusResponse
		wantMsg string
	}{
		{"old version", copilotsdk.GetStatusResponse{Version: "1.0.83", ProtocolVersion: 3}, "1.0.84-1"},
		{"old revision", copilotsdk.GetStatusResponse{Version: "1.0.84", ProtocolVersion: 3}, "1.0.84-1"},
		{"unparseable", copilotsdk.GetStatusResponse{Version: "dev", ProtocolVersion: 3}, "unparseable"},
		{"protocol mismatch", copilotsdk.GetStatusResponse{Version: "1.0.84-1", ProtocolVersion: 2}, "protocol"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t, func(_ *Config, f *fakeClient) { st := tc.status; f.status = &st })
			_, err := h.ad.Start(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.wantMsg) {
				t.Fatalf("Start error = %v, want containing %q", err, tc.wantMsg)
			}
			if h.client.createCalls.Load() != 0 {
				t.Error("no session may be created on a version mismatch")
			}
			if h.client.forceStops.Load() == 0 {
				t.Error("client must be torn down on a version mismatch")
			}
		})
	}
}

func TestStartAcceptsNewerVersions(t *testing.T) {
	for _, v := range []string{"1.0.84-1", "1.0.84-2", "1.0.85", "1.1.0", "2.0.0-1"} {
		h := newHarness(t, func(_ *Config, f *fakeClient) {
			f.status = &copilotsdk.GetStatusResponse{Version: v, ProtocolVersion: 3}
		})
		if _, err := h.ad.Start(context.Background()); err != nil {
			t.Errorf("version %q rejected: %v", v, err)
		}
	}
}

func TestVersionAtLeast(t *testing.T) {
	cases := []struct {
		got, floor string
		want       bool
		wantErr    bool
	}{
		{"1.0.84-1", "1.0.84-1", true, false},
		{"1.0.84", "1.0.84-1", false, false},
		{"1.0.83-9", "1.0.84-1", false, false},
		{"1.0.84-10", "1.0.84-2", true, false},
		{"x", "1.0.84-1", false, true},
		{"1.0", "1.0.84-1", false, true},
		{"1.0.84-beta", "1.0.84-1", false, true},
		{"", "1.0.84-1", false, true},
	}
	for _, tc := range cases {
		got, err := versionAtLeast(tc.got, tc.floor)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("versionAtLeast(%q,%q) = %v,%v want %v err=%v", tc.got, tc.floor, got, err, tc.want, tc.wantErr)
		}
	}
}

func TestNewValidatesConfig(t *testing.T) {
	base := func() Config {
		return Config{
			StartupTimeout: time.Second,
			Ingress:        make(chan agent.Event, 1),
			Asks:           make(chan agent.PermissionAsk, 1),
			HubDone:        make(chan struct{}),
		}
	}
	cases := []struct {
		name   string
		mutate func(*Config)
		ok     bool
	}{
		{"valid empty cli path", func(*Config) {}, true},
		{"bare name", func(c *Config) { c.CLIPath = "copilot" }, true},
		{"absolute", func(c *Config) { c.CLIPath = absTestPath() }, true},
		{"relative dot", func(c *Config) { c.CLIPath = "./bin/copilot" }, false},
		{"relative dotdot", func(c *Config) { c.CLIPath = `..\x\copilot` }, false},
		{"relative subdir", func(c *Config) { c.CLIPath = "bin/copilot" }, false},
		{"drive relative", func(c *Config) { c.CLIPath = "C:copilot" }, false},
		{"nil ingress", func(c *Config) { c.Ingress = nil }, false},
		{"nil asks", func(c *Config) { c.Asks = nil }, false},
		{"nil hubdone", func(c *Config) { c.HubDone = nil }, false},
		{"zero startup timeout", func(c *Config) { c.StartupTimeout = 0 }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base()
			tc.mutate(&cfg)
			_, err := New(cfg)
			if (err == nil) != tc.ok {
				t.Fatalf("New error = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

// ---- event flow (065.002/005 wiring, callback contract 4.2) ----

func TestCallbackDeliversNormalizedEventToIngress(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantMessageDeltaData{MessageID: "m", DeltaContent: "x"}})
	select {
	case ev := <-h.ingress:
		if ev.Kind != agent.KindDelta || ev.Text != "x" {
			t.Fatalf("ev = %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no event delivered")
	}
}

func TestCallbackUnknownEventIsForwardedNotDropped(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantReasoningData{}})
	select {
	case ev := <-h.ingress:
		if ev.Kind != agent.KindUnknown {
			t.Fatalf("ev = %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("unknown event dropped")
	}
}

func TestCallbackBlockingSendNeverDropsWhenIngressFull(t *testing.T) {
	h := newHarness(t, func(c *Config, _ *fakeClient) { c.Ingress = make(chan agent.Event) })
	// replace harness ingress with the unbuffered one for reading
	ing := make(chan agent.Event)
	h.ad.cfg.Ingress = ing
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	delivered := make(chan struct{})
	go func() {
		h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnStartData{TurnID: "t"}})
		close(delivered)
	}()
	select {
	case <-delivered:
		t.Fatal("callback returned without the ingress consuming (silent drop)")
	case <-time.After(50 * time.Millisecond):
	}
	if ev := <-ing; ev.Kind != agent.KindTurnStart {
		t.Fatalf("ev = %+v", ev)
	}
	<-delivered
}

// ---- 065.005-T: quiesce ----

func TestStopReleasesBlockedCallbackAndNeverClosesIngress(t *testing.T) {
	h := newHarness(t, nil)
	ing := make(chan agent.Event) // unbuffered, nobody reads: callback blocks
	h.ad.cfg.Ingress = ing
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	cbDone := make(chan struct{})
	go func() {
		h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnStartData{TurnID: "t"}})
		close(cbDone)
	}()
	time.Sleep(50 * time.Millisecond)

	stopped := make(chan error, 1)
	go func() { stopped <- h.ad.Stop(context.Background()) }()
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not return within its quiesce deadline")
	}
	select {
	case <-cbDone:
	case <-time.After(time.Second):
		t.Fatal("blocked callback was not released by Stop")
	}
	if h.client.session.unsubscribe.Load() != 1 {
		t.Errorf("unsubscribe called %d times, want 1", h.client.session.unsubscribe.Load())
	}
	select {
	case _, ok := <-ing:
		if !ok {
			t.Fatal("ingress channel was closed by the adapter")
		}
	default:
	}
	// A late callback after Stop must neither panic nor deliver.
	h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnEndData{TurnID: "t"}})
}

func TestStopIsIdempotentAndSafeBeforeStart(t *testing.T) {
	h := newHarness(t, nil)
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatalf("Stop before Start: %v", err)
	}
	h2 := newHarness(t, nil)
	if _, err := h2.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h2.ad.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h2.ad.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop: %v", err)
	}
	if h2.client.stops.Load() != 1 {
		t.Errorf("client.Stop called %d times, want 1", h2.client.stops.Load())
	}
}

func TestStopFallsBackToForceStop(t *testing.T) {
	h := newHarness(t, func(_ *Config, f *fakeClient) { f.stopErr = errors.New("stuck") })
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Stop(context.Background()); err == nil {
		t.Fatal("Stop must report the client stop failure")
	}
	if h.client.forceStops.Load() != 1 {
		t.Error("failed Stop must escalate to ForceStop")
	}
}

func TestStopRaceWithCallbacks(t *testing.T) {
	h := newHarness(t, nil)
	h.ad.cfg.Ingress = make(chan agent.Event, 1)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				h.client.session.emit(copilotsdk.SessionEvent{Data: &rpc.AssistantTurnStartData{TurnID: "t"}})
			}
		}()
	}
	if err := h.ad.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	wg.Wait()
}

// ---- 065.006-T: Send / Abort ----

func TestSendAndAbortRequireStart(t *testing.T) {
	h := newHarness(t, nil)
	if err := h.ad.Send(context.Background(), "hi"); err == nil {
		t.Error("Send before Start must fail")
	}
	if err := h.ad.Abort(context.Background()); err == nil {
		t.Error("Abort before Start must fail")
	}
}

func TestSendEnforcesDefaultDeadline(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Send(context.Background(), "hello"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !h.client.session.sendCtxHasDeadline {
		t.Error("Send must always pass a deadline-bearing context to the SDK")
	}
	if h.client.session.sentPrompt != "hello" {
		t.Errorf("prompt = %q", h.client.session.sentPrompt)
	}
	if err := h.ad.Abort(context.Background()); err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if !h.client.session.abortCtxHasDeadline {
		t.Error("Abort must always pass a deadline-bearing context to the SDK")
	}
}

func TestSendAbortWrapErrors(t *testing.T) {
	sentinel := errors.New("sdk failure")
	h := newHarness(t, nil)
	h.client.session.sendErr = sentinel
	h.client.session.abortErr = sentinel
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	for name, err := range map[string]error{
		"send":  h.ad.Send(context.Background(), "x"),
		"abort": h.ad.Abort(context.Background()),
	} {
		if !errors.Is(err, sentinel) || !strings.Contains(err.Error(), name) {
			t.Errorf("%s error = %v, want wrapped with context", name, err)
		}
	}
}

func TestSendRejectsEmptyPromptAndCancelledContext(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := h.ad.Send(context.Background(), "  "); err == nil {
		t.Error("blank prompt must be rejected")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := h.ad.Send(ctx, "x"); err == nil {
		t.Error("cancelled context must be rejected")
	}
}

// SQ-a: Abort must return promptly even while a permission handler is blocked
// waiting on the hub.
func TestAbortReturnsPromptlyWithBlockedPermissionHandler(t *testing.T) {
	h := newHarness(t, nil)
	if _, err := h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	handler := h.client.createCfg.OnPermissionRequest
	blocked := make(chan struct{})
	go func() {
		close(blocked)
		_, _ = handler(&rpc.PermissionRequestShell{FullCommandText: "x"}, copilotsdk.PermissionInvocation{})
	}()
	<-blocked
	<-h.asks // handler is now parked waiting for a reply

	start := time.Now()
	if err := h.ad.Abort(context.Background()); err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("Abort did not return promptly with a blocked permission handler")
	}
	close(h.hubDone) // release the parked handler
}
