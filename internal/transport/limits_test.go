package transport

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// dialRaw opens a plain TCP connection to the server (no HTTP spoken).
func dialRaw(t *testing.T, e *env) net.Conn {
	t.Helper()
	c, err := net.DialTimeout("tcp", e.hostPort(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

// statusOfRaw reads an HTTP status line from c, or "" if it closes first.
func statusOfRaw(c net.Conn, wait time.Duration) string {
	_ = c.SetReadDeadline(time.Now().Add(wait))
	line, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

func eventually(t *testing.T, what string, within time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestConnectionsAboveCapAreRefusedIncludingUnauthenticated(t *testing.T) {
	e := newEnv(t, envOpts{maxConns: 2})
	// Two raw sockets that never authenticate (never even send a request)
	// hold both slots.
	a, b := dialRaw(t, e), dialRaw(t, e)
	eventually(t, "both slots taken", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 2 })

	c := dialRaw(t, e)
	if got := statusOfRaw(c, 2*time.Second); !strings.Contains(got, "503") {
		t.Fatalf("over-cap connection got %q; want an HTTP 503", got)
	}
	// A refused socket is closed and never counted.
	if n := e.srv.ActiveConnections(); n != 2 {
		t.Fatalf("active = %d after a refusal; want 2", n)
	}

	// Releasing one slot (any exit path) admits a new connection.
	_ = a.Close()
	eventually(t, "slot released", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 1 })
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": {"Bearer " + e.app.Token.Reveal()}}})
	if err != nil {
		t.Fatalf("connection after release refused: %v", err)
	}
	// Now full again (b + ws): a further one is refused.
	eventually(t, "full again", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 2 })
	if got := statusOfRaw(dialRaw(t, e), 2*time.Second); !strings.Contains(got, "503") {
		t.Fatalf("got %q; want 503", got)
	}
	// Closing the authenticated socket releases its slot too.
	_ = ws.CloseNow()
	_ = b.Close()
	eventually(t, "all released", 3*time.Second, func() bool { return e.srv.ActiveConnections() == 0 })
}

func TestSlowHandshakeClosedAtTimeout(t *testing.T) {
	e := newEnv(t, envOpts{maxConns: 1, srvTweak: func(sc *ServerConfig) { sc.HandshakeTimeout = 150 * time.Millisecond }})
	c := dialRaw(t, e)
	// A stalled handshake: a partial request line, then silence.
	if _, err := c.Write([]byte("GET /ws HTTP/1.1\r\nHost: ")); err != nil {
		t.Fatal(err)
	}
	begin := time.Now()
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 64)
	_, err := c.Read(buf)
	if err == nil {
		t.Fatal("expected the server to close the stalled handshake")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatalf("server did not close the stalled handshake within the deadline (%v)", time.Since(begin))
	}
	if d := time.Since(begin); d > 2*time.Second {
		t.Fatalf("closed after %v; want about the 150ms handshake timeout", d)
	}
	// The timed-out socket released its (only) slot.
	eventually(t, "slot released after handshake timeout", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 0 })
}

func TestIdleUnauthenticatedSocketClosedAtTimeout(t *testing.T) {
	e := newEnv(t, envOpts{maxConns: 1, srvTweak: func(sc *ServerConfig) { sc.HandshakeTimeout = 150 * time.Millisecond }})
	c := dialRaw(t, e) // never sends a byte
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := c.Read(make([]byte, 8)); err == nil {
		t.Fatal("expected close")
	} else if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("silent socket was never closed by the server")
	}
	eventually(t, "released", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 0 })
}

func TestLogsNeverContainCredentialsOrSecretHeaders(t *testing.T) {
	e := newEnv(t, envOpts{origins: []string{"https://app.example"}})
	tok := e.app.Token.Reveal()
	obsTok := e.obs.Token.Reveal()

	// A spread of requests: rejected and accepted, header and subprotocol
	// carriage, credential in the URL, bad host/origin, wrong credentials.
	e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok}, "")
	e.rawUpgrade(map[string]string{"Authorization": "Bearer wrongwrongwrongwrongwrong"}, "")
	e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "intercom.v1, bearer." + obsTok}, "")
	e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "bearer." + tok}, "")
	e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "intercom.v1, bearer.secretsecretsecret"}, "")
	e.rawUpgrade(nil, "token="+tok)
	e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok, "Origin": "https://evil.example"}, "")
	e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok, "Host": "evil.example"}, "")
	e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok, "Origin": "https://app.example"}, "")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{Subprotocols: []string{"intercom.v1", "bearer." + tok}})
	if err != nil {
		t.Fatal(err)
	}
	_ = ws.CloseNow()
	// malformed junk that makes net/http itself log
	c := dialRaw(t, e)
	_, _ = c.Write([]byte("Authorization: Bearer " + tok + "\r\nSec-WebSocket-Protocol: bearer." + tok + "\r\n\r\n"))
	_ = statusOfRaw(c, time.Second)
	time.Sleep(50 * time.Millisecond)

	out := e.logs.all()
	if out == "" {
		t.Fatal("test captured no log output; the capture is not wired")
	}
	for _, secret := range []string{tok, obsTok, "secretsecretsecret", "wrongwrongwrongwrongwrong", "bearer.", "Bearer ", "Authorization", "Sec-WebSocket-Protocol", "evil.example", "app.example"} {
		if strings.Contains(out, secret) {
			t.Fatalf("log output contains %q:\n%s", secret, out)
		}
	}
}
