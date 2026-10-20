package transport

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func hydrations(t *testing.T, e *env) uint64 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	st, err := e.hub.Stats(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return st.Hydrations
}

// Item 2: once Drain has begun, no new upgrade is admitted.
func TestDrainRefusesNewUpgradesWith503(t *testing.T) {
	e := newEnv(t, envOpts{})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := e.handler.Drain(ctx); err != nil {
		t.Fatal(err)
	}
	before := hydrations(t, e)
	wantStatus(t, e.rawUpgrade(bearer(e.app.Token.Reveal()), ""), http.StatusServiceUnavailable, "upgrade after Drain")
	if got := hydrations(t, e); got != before {
		t.Fatal("a refused upgrade still registered with the hub")
	}
}

// Item 2: Drain racing with concurrent upgrades must be race-free (run with
// -race) and must never leave a session running after Drain returns.
func TestDrainConcurrentWithUpgradesIsRaceFree(t *testing.T) {
	e := newEnv(t, envOpts{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/ws", nil)
			req.Host = e.hostPort()
			req.Header.Set("Authorization", "Bearer "+e.app.Token.Reveal())
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
			req.Header.Set("Sec-WebSocket-Version", "13")
			req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
			e.handler.ServeHTTP(httptest.NewRecorder(), req)
		}()
		if i == 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_ = e.handler.Drain(ctx)
			}()
		}
	}
	wg.Wait()
	if n := e.handler.ActiveSessions(); n != 0 {
		t.Fatalf("%d sessions running after Drain", n)
	}
}

// Item 3: a credential holder's plain GET (no upgrade headers) is refused
// with a generic 400 BEFORE the hub registration.
func TestNonUpgradeRequestsRejectedBeforeHubRegistration(t *testing.T) {
	e := newEnv(t, envOpts{})
	full := map[string]string{
		"Authorization":         "Bearer " + e.app.Token.Reveal(),
		"Connection":            "Upgrade",
		"Upgrade":               "websocket",
		"Sec-WebSocket-Version": "13",
		"Sec-WebSocket-Key":     "dGhlIHNhbXBsZSBub25jZQ==",
	}
	without := func(k string) map[string]string {
		m := map[string]string{}
		for kk, v := range full {
			if kk != k {
				m[kk] = v
			}
		}
		return m
	}
	with := func(k, v string) map[string]string {
		m := without(k)
		m[k] = v
		return m
	}
	cases := map[string]map[string]string{
		"plain GET":          {"Authorization": full["Authorization"]},
		"no Upgrade":         without("Upgrade"),
		"wrong Upgrade":      with("Upgrade", "h2c"),
		"no Connection":      without("Connection"),
		"Connection no upg":  with("Connection", "keep-alive"),
		"no version":         without("Sec-WebSocket-Version"),
		"no key":             without("Sec-WebSocket-Key"),
		"empty key":          with("Sec-WebSocket-Key", ""),
		"keep-alive,Upgrade": with("Connection", "keep-alive, Upgrade"), // valid: control
	}
	before := hydrations(t, e)
	for name, hdr := range cases {
		req, err := http.NewRequest(http.MethodGet, "http://"+e.hostPort()+WSPath, nil)
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range hdr {
			if v != "" {
				req.Header.Set(k, v)
			}
		}
		cl := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
		resp, err := cl.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		_ = resp.Body.Close()
		if name == "keep-alive,Upgrade" {
			if resp.StatusCode != http.StatusSwitchingProtocols {
				t.Fatalf("%s: status %d; a valid upgrade must pass", name, resp.StatusCode)
			}
			continue
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("%s: status %d; want 400", name, resp.StatusCode)
		}
	}
	// the only registration is the valid control request
	if got := hydrations(t, e); got > before+1 {
		t.Fatalf("hydrations grew by %d; non-upgrade requests registered with the hub", got-before)
	}
	// ...and the method must be GET
	req, _ := http.NewRequest(http.MethodPost, "http://"+e.hostPort()+WSPath, nil)
	for k, v := range full {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST status %d; want 400", resp.StatusCode)
	}
}

// readRawResponse reads one HTTP response head from c and reports whether
// "Connection: close" was present.
func readRawResponse(t *testing.T, c net.Conn) (status string, closes bool) {
	t.Helper()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	br := bufio.NewReader(c)
	line, err := br.ReadString('\n')
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	status = strings.TrimSpace(line)
	for {
		l, err := br.ReadString('\n')
		if err != nil || strings.TrimSpace(l) == "" {
			return status, closes
		}
		if strings.EqualFold(strings.TrimSpace(l), "connection: close") {
			closes = true
		}
	}
}

// Item 4: every pre-upgrade rejection says Connection: close and releases its
// cap slot promptly, without the client closing the socket.
func TestRejectedRequestReleasesSlotPromptly(t *testing.T) {
	e := newEnv(t, envOpts{maxConns: 1})
	addr := e.hostPort()
	cases := map[string]string{
		"401 no credential": "GET /ws HTTP/1.1\r\nHost: " + addr + "\r\n\r\n",
		"403 bad host":      "GET /ws HTTP/1.1\r\nHost: evil.example\r\n\r\n",
		"404 path":          "GET /other HTTP/1.1\r\nHost: " + addr + "\r\n\r\n",
		"400 not upgrade": "GET /ws HTTP/1.1\r\nHost: " + addr + "\r\nAuthorization: Bearer " +
			e.app.Token.Reveal() + "\r\n\r\n",
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			c := dialRaw(t, e)
			if _, err := c.Write([]byte(req)); err != nil {
				t.Fatal(err)
			}
			status, closes := readRawResponse(t, c)
			if !closes {
				t.Fatalf("%s: response lacks Connection: close", status)
			}
			eventually(t, "rejected socket's slot released", time.Second, func() bool {
				return e.srv.ActiveConnections() == 0
			})
		})
	}
}

// Item 4: the over-cap 503 may not stall the accept loop for long, and a
// flood of refused sockets must not prevent admitting a freed slot.
func TestOverCapRefusalDoesNotStallAcceptLoop(t *testing.T) {
	if refuseWriteTimeout > 50*time.Millisecond {
		t.Fatalf("refuseWriteTimeout = %v; the accept loop must not stall longer than 50ms per refusal", refuseWriteTimeout)
	}
	e := newEnv(t, envOpts{maxConns: 1})
	hold := dialRaw(t, e)
	eventually(t, "slot taken", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 1 })
	for i := 0; i < 40; i++ { // refused sockets that never read their 503
		c, err := net.DialTimeout("tcp", e.hostPort(), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
	}
	// The accept loop is FIFO: once a LATER over-cap socket gets its 503 the
	// whole flood has been processed (none can be admitted after the release
	// below and squat on the freed slot).
	if got := statusOfRaw(dialRaw(t, e), 5*time.Second); !strings.Contains(got, "503") {
		t.Fatalf("over-cap socket behind a refusal flood got %q; want 503", got)
	}
	_ = hold.Close()
	eventually(t, "freed slot admitted despite the refusal flood", 3*time.Second, func() bool {
		c, err := net.DialTimeout("tcp", e.hostPort(), time.Second)
		if err != nil {
			return false
		}
		defer func() { _ = c.Close() }()
		_, _ = c.Write([]byte("GET /ws HTTP/1.1\r\nHost: " + e.hostPort() + "\r\n\r\n"))
		return strings.Contains(statusOfRaw(c, time.Second), "401")
	})
}

// Item 6: the action round trip (which blocks the read pump, hence pong
// handling) must finish well inside the ping write timeout.
func TestActionTimeoutLeavesPongMargin(t *testing.T) {
	if !actionFitsWriteTimeout(DefaultActionTimeout, DefaultWriteTimeout) {
		t.Fatalf("DefaultActionTimeout %v leaves no margin under DefaultWriteTimeout %v", DefaultActionTimeout, DefaultWriteTimeout)
	}
	e := newEnv(t, envOpts{})
	_, err := NewHandler(Config{Hub: e.hub, Creds: e.creds, WriteTimeout: time.Second, ActionTimeout: time.Second})
	if err == nil {
		t.Fatal("NewHandler accepted an action timeout that can starve ping replies")
	}
	if _, err := NewHandler(Config{Hub: e.hub, Creds: e.creds, WriteTimeout: time.Second, ActionTimeout: 500 * time.Millisecond}); err != nil {
		t.Fatalf("valid timeouts refused: %v", err)
	}
	h, err := NewHandler(Config{Hub: e.hub, Creds: e.creds, WriteTimeout: 400 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if !actionFitsWriteTimeout(h.actionTimeout, h.writeTimeout) {
		t.Fatalf("defaulted action timeout %v does not fit write timeout %v", h.actionTimeout, h.writeTimeout)
	}
}

// Item 7: Shutdown within MinShutdownBudget completes without a deadline
// error even when a peer never answers the close handshake.
func TestShutdownBudgetCoversCloseGrace(t *testing.T) {
	if MinShutdownBudget < DefaultCloseGrace+time.Second {
		t.Fatalf("MinShutdownBudget %v is below closeGrace %v + margin", MinShutdownBudget, DefaultCloseGrace)
	}
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) { c.CloseGrace = 300 * time.Millisecond }})
	c := dialWS(t, e, e.app.Token.Reveal()) // never reads again: never answers the close
	readEnvelope(t, c, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond+time.Second)
	defer cancel()
	if err := e.srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned %v merely because closeGrace elapsed", err)
	}
	if n := e.handler.ActiveSessions(); n != 0 {
		t.Fatalf("%d sessions outlived Shutdown", n)
	}
}

// Item 7: Shutdown closes the bound listener even when Serve never ran.
func TestShutdownClosesListenerWithoutServe(t *testing.T) {
	srv, err := NewServer(ServerConfig{Host: "127.0.0.1", Handler: http.NotFoundHandler()})
	if err != nil {
		t.Fatal(err)
	}
	addr := srv.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	if c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond); err == nil {
		_ = c.Close()
		t.Fatal("listener still accepting after Shutdown (Serve was never called)")
	}
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("second Shutdown: %v", err)
	}
}
