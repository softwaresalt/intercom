package transport

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// env is a running hub + credential store + handler + loopback server.
type env struct {
	t       *testing.T
	hub     *hub.Hub
	creds   *CredentialStore
	handler *Handler
	srv     *Server
	app     Issued
	obs     Issued
	logs    *captureHandler
}

type envOpts struct {
	origins  []string
	hosts    []string
	maxConns int
	lifetime time.Duration
	hubCfg   hub.Config
	cfgTweak func(*Config)
	srvTweak func(*ServerConfig)
	clock    *fakeClock
}

func nowOf(c *fakeClock) func() time.Time {
	if c == nil {
		return nil
	}
	return c.Now
}

func newEnv(t *testing.T, o envOpts) *env {
	t.Helper()
	if o.maxConns == 0 {
		o.maxConns = 8
	}
	if o.lifetime == 0 {
		o.lifetime = time.Hour
	}
	hc := o.hubCfg
	if hc.SessionID == "" {
		hc.SessionID = "sess"
	}
	h, err := hub.New(hc)
	if err != nil {
		t.Fatal(err)
	}
	hubCtx, hubCancel := context.WithCancel(context.Background())
	runDone := make(chan struct{})
	go func() { defer close(runDone); _ = h.Run(hubCtx) }()

	creds, err := NewCredentialStore(CredentialConfig{Gen: h.Gen(), Lifetime: o.lifetime, Now: nowOf(o.clock)})
	if err != nil {
		t.Fatal(err)
	}
	logs := &captureHandler{}
	cfg := Config{
		Hub: h, Creds: creds,
		AllowedOrigins: o.origins, AllowedHosts: o.hosts,
		Logger: slog.New(logs),
		Now:    nowOf(o.clock),
	}
	if o.cfgTweak != nil {
		o.cfgTweak(&cfg)
	}
	handler, err := NewHandler(cfg)
	if err != nil {
		t.Fatal(err)
	}
	sc := ServerConfig{Host: "127.0.0.1", Handler: handler, MaxConnections: o.maxConns, Logger: slog.New(logs), Drain: handler.Drain, OnBind: handler.SetBoundAddr}
	if o.srvTweak != nil {
		o.srvTweak(&sc)
	}
	srv, err := NewServer(sc)
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve() }()

	e := &env{t: t, hub: h, creds: creds, handler: handler, srv: srv, logs: logs}
	e.app, _ = creds.Mint(hub.RoleApprover)
	e.obs, _ = creds.Mint(hub.RoleObserver)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			t.Errorf("server shutdown: %v", err)
		}
		h.Stop()
		hubCancel()
		<-runDone
	})
	return e
}

func (e *env) hostPort() string { return e.srv.Addr().String() }
func (e *env) wsURL() string    { return "ws://" + e.hostPort() + WSPath }

// event pushes one agent message through the hub.
func (e *env) event(text string) {
	e.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.hub.Send(ctx, agent.Event{Kind: agent.KindAgentMessage, MessageID: "m", Text: text}); err != nil {
		e.t.Fatal(err)
	}
}

// rawUpgrade sends a hand-built WebSocket upgrade request and returns the
// response (never upgraded by the HTTP client; a 101 is returned as-is).
func (e *env) rawUpgrade(hdr map[string]string, query string) *http.Response {
	e.t.Helper()
	u := "http://" + e.hostPort() + WSPath
	if query != "" {
		u += "?" + query
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		e.t.Fatal(err)
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", testWebSocketKey())
	for k, v := range hdr {
		if k == "Host" {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
	}
	cl := &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{DisableKeepAlives: true},
	}
	resp, err := cl.Do(req)
	if err != nil {
		e.t.Fatalf("raw upgrade: %v", err)
	}
	e.t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func bodyOf(r *http.Response) string {
	b, _ := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	return string(b)
}

// testWebSocketKey returns a syntactically valid Sec-WebSocket-Key (16 bytes, base64) built at runtime, so no key-shaped literal sits in the tree for secret scanners.
func testWebSocketKey() string { return base64.StdEncoding.EncodeToString(make([]byte, 16)) }

// waitPrompts waits until the fake agent has received at least n prompts and returns them. The hub hands a prompt to the agent asynchronously relative to the acceptance reply, so a single sample can race.
func waitPrompts(t *testing.T, fa *fakeAgent, n int) []string {
	t.Helper()
	var got []string
	eventually(t, "agent prompts", 3*time.Second, func() bool {
		got, _ = fa.snapshot()
		return len(got) >= n
	})
	return got
}
