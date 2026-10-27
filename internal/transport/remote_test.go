package transport

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

// remoteRig is a running hub plus a composed Remote (070.001-T).
type remoteRig struct {
	t      *testing.T
	hub    *hub.Hub
	remote *Remote
	clock  *fakeClock
	logs   *captureHandler
}

func newRemoteRig(t *testing.T, rc config.RemoteConfig) *remoteRig {
	t.Helper()
	return newRemoteRigWith(t, rc, nil)
}

// newRemoteRigWith is newRemoteRig with a hook to adjust the RemoteOptions.
func newRemoteRigWith(t *testing.T, rc config.RemoteConfig, mutate func(*RemoteOptions)) *remoteRig {
	t.Helper()
	h, err := hub.New(hub.Config{SessionID: "sess"})
	if err != nil {
		t.Fatal(err)
	}
	runDone := make(chan struct{})
	go func() { defer close(runDone); _ = h.Run(context.Background()) }()
	clock := newFakeClock()
	logs := &captureHandler{}
	opts := RemoteOptions{Hub: h, Remote: rc, Port: 0, Now: clock.Now, Logger: slog.New(logs)}
	if mutate != nil {
		mutate(&opts)
	}
	r, err := NewRemote(opts)
	if err != nil {
		h.Stop()
		<-runDone
		t.Fatalf("NewRemote: %v", err)
	}
	go func() { _ = r.Serve() }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := r.Shutdown(ctx); err != nil {
			t.Errorf("remote shutdown: %v", err)
		}
		h.Stop()
		<-runDone
	})
	return &remoteRig{t: t, hub: h, remote: r, clock: clock, logs: logs}
}

func (rr *remoteRig) url() string { return "ws://" + rr.remote.Addr().String() + WSPath }

func defaultRemote() config.RemoteConfig { return config.Default().Remote }

func TestNewRemoteRejectsMissingHub(t *testing.T) {
	if _, err := NewRemote(RemoteOptions{Remote: defaultRemote()}); err == nil {
		t.Fatal("NewRemote without a hub succeeded")
	}
}

func TestNewRemoteBindsLoopbackAndPairsOnBind(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	if err := AssertLoopbackAddr(rr.remote.Addr()); err != nil {
		t.Fatalf("bound address is not loopback: %v", err)
	}
	// OnBind pairing: the handler learned the ephemeral port, so the loopback
	// Host name for that port authenticates; an unpaired handler would refuse it.
	iss, err := rr.remote.Mint(hub.RoleApprover)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: rr.url(), Credential: iss.Token.Reveal()})
	if err != nil {
		t.Fatalf("dial through the composed remote: %v", err)
	}
	defer c.Abort()
}

func TestNewRemoteMapsRemoteConfig(t *testing.T) {
	rc := defaultRemote()
	rc.TokenLifetimeSeconds = 120
	rc.MaxConnections = 3
	rc.AllowedOrigins = []string{"https://ui.example.test"}
	rr := newRemoteRig(t, rc)

	iss, err := rr.remote.Mint(hub.RoleObserver)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := iss.ExpiresAt, rr.clock.Now().Add(120*time.Second); !got.Equal(want) {
		t.Errorf("expiry = %v; want %v (token_lifetime_seconds on the injected clock)", got, want)
	}
	if got := rr.remote.server.ln.max; got != 3 {
		t.Errorf("connection cap = %d; want 3 (max_connections)", got)
	}

	dial := func(origin string) (*http.Response, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hdr := http.Header{"Authorization": {"Bearer " + iss.Token.Reveal()}, "Origin": {origin}}
		c, resp, err := websocket.Dial(ctx, rr.url(), &websocket.DialOptions{HTTPHeader: hdr})
		if c != nil {
			_ = c.CloseNow()
		}
		return resp, err
	}
	if _, err := dial("https://ui.example.test"); err != nil {
		t.Errorf("allowed origin refused: %v", err)
	}
	resp, err := dial("https://evil.example.test")
	if err == nil {
		t.Fatal("origin outside allowed_origins was upgraded")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign origin response = %v; want 403", resp)
	}
}

func TestNewRemoteRefusesInvalidRemoteConfig(t *testing.T) {
	h, err := hub.New(hub.Config{SessionID: "sess"})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Stop()
	rc := defaultRemote()
	rc.AllowedOrigins = []string{"https://UI.example.test/path"}
	if _, err := NewRemote(RemoteOptions{Hub: h, Remote: rc}); err == nil {
		t.Fatal("a non-canonical allowed origin was accepted")
	}
	rc = defaultRemote()
	rc.TokenLifetimeSeconds = 0
	if _, err := NewRemote(RemoteOptions{Hub: h, Remote: rc}); err == nil {
		t.Fatal("a zero credential lifetime was accepted")
	}
}

func TestNewRemoteRefusesNonLoopbackHost(t *testing.T) {
	h, err := hub.New(hub.Config{SessionID: "sess"})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Stop()
	if _, err := NewRemote(RemoteOptions{Hub: h, Remote: defaultRemote(), Host: "0.0.0.0"}); err == nil {
		t.Fatal("a non-loopback bind host was accepted")
	}
}

func TestRemoteShutdownDrainsLiveSessions(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	iss, err := rr.remote.Mint(hub.RoleApprover)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: rr.url(), Credential: iss.Token.Reveal()})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Abort()
	eventually(t, "session registered", 5*time.Second, func() bool { return rr.remote.handler.ActiveSessions() == 1 })

	sctx, scancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer scancel()
	if err := rr.remote.Shutdown(sctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	// Drain pairing: Shutdown waited for the hijacked session goroutines.
	if n := rr.remote.handler.ActiveSessions(); n != 0 {
		t.Fatalf("%d sessions still running after Shutdown (Drain not paired)", n)
	}
}

type rejection struct {
	status int
	reason string
	total  uint64
}

type rejectLog struct {
	mu   sync.Mutex
	list []rejection
}

func (l *rejectLog) record(status int, reason string, total uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.list = append(l.list, rejection{status, reason, total})
}

func (l *rejectLog) snapshot() []rejection {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]rejection(nil), l.list...)
}

func TestRemoteCountsRejectedUpgradesAndInvokesOnReject(t *testing.T) {
	const badToken = "not-a-real-credential-value"
	rec := &rejectLog{}
	rc := defaultRemote()
	rr := newRemoteRigWith(t, rc, func(o *RemoteOptions) { o.OnReject = rec.record })

	if got := rr.remote.RejectedUpgrades(); got != 0 {
		t.Fatalf("RejectedUpgrades before any request = %d; want 0", got)
	}

	dial := func(hdr http.Header) int {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		c, resp, err := websocket.Dial(ctx, rr.url(), &websocket.DialOptions{HTTPHeader: hdr})
		if c != nil {
			_ = c.CloseNow()
		}
		if err == nil {
			t.Fatal("rejected dial was upgraded")
		}
		if resp == nil {
			t.Fatalf("no HTTP response: %v", err)
		}
		return resp.StatusCode
	}

	if got := dial(http.Header{"Authorization": {"Bearer " + badToken}}); got != http.StatusUnauthorized {
		t.Fatalf("bad credential status = %d; want 401", got)
	}
	if got := rr.remote.RejectedUpgrades(); got != 1 {
		t.Fatalf("RejectedUpgrades after 401 = %d; want 1", got)
	}
	if got := dial(http.Header{"Authorization": {"Bearer " + badToken}, "Origin": {"https://evil.example.test"}}); got != http.StatusForbidden {
		t.Fatalf("foreign origin status = %d; want 403", got)
	}
	if got := rr.remote.RejectedUpgrades(); got != 2 {
		t.Fatalf("RejectedUpgrades after 403 = %d; want 2", got)
	}
	if got := rr.remote.handler.RejectedUpgrades(); got != 2 {
		t.Fatalf("Handler.RejectedUpgrades = %d; want 2 (same counter as Remote)", got)
	}

	want := []rejection{
		{http.StatusUnauthorized, rejectCredential, 1},
		{http.StatusForbidden, rejectOrigin, 2},
	}
	got := rec.snapshot()
	if len(got) != len(want) {
		t.Fatalf("OnReject calls = %+v; want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("OnReject call %d = %+v; want %+v", i, got[i], want[i])
		}
		if strings.Contains(got[i].reason, badToken) {
			t.Errorf("OnReject reason %q carries the presented credential", got[i].reason)
		}
	}
}

func TestRemoteRejectedUpgradesCountsWithoutOnReject(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := websocket.Dial(ctx, rr.url(), nil)
	if err == nil {
		t.Fatal("unauthenticated dial was upgraded")
	}
	if got := rr.remote.RejectedUpgrades(); got != 1 {
		t.Fatalf("RejectedUpgrades = %d; want 1 with no OnReject configured", got)
	}
}

// TestRemoteOnRejectTotalsAreUniqueUnderConcurrency drives n simultaneous bad
// upgrades. The connection cap is raised to n so none is refused by the
// pre-auth limiter (which does not count as a handler-level rejection).
func TestRemoteOnRejectTotalsAreUniqueUnderConcurrency(t *testing.T) {
	const n = 20
	rec := &rejectLog{}
	rc := defaultRemote()
	rc.MaxConnections = n
	rr := newRemoteRigWith(t, rc, func(o *RemoteOptions) { o.OnReject = rec.record })
	var wg sync.WaitGroup
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if c, _, err := websocket.Dial(ctx, rr.url(), nil); err == nil {
				_ = c.CloseNow()
			}
		}()
	}
	wg.Wait()
	if got := rr.remote.RejectedUpgrades(); got != n {
		t.Fatalf("RejectedUpgrades = %d; want %d", got, n)
	}
	seen := make(map[uint64]bool, n)
	for _, r := range rec.snapshot() {
		if r.total < 1 || r.total > n || seen[r.total] {
			t.Fatalf("OnReject total %d out of range or duplicated", r.total)
		}
		seen[r.total] = true
	}
	if len(seen) != n {
		t.Fatalf("OnReject saw %d distinct totals; want %d", len(seen), n)
	}
}
