package transport

import (
	"context"
	"net"
	"testing"
	"time"
)

// A connection that never completes its HTTP handshake keeps http.Server's
// graceful Shutdown waiting (a StateNew conn is only treated as idle after
// 5s). That wait must not consume the budget the session drain needs: an
// established WebSocket session must still be closed and joined well within
// the caller's deadline.
func TestShutdownDrainsSessionsDespiteStalledHandshake(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) {
		c.CloseGrace = 300 * time.Millisecond
	}})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 3*time.Second)
	eventually(t, "session registered", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 1 })

	stalled, err := net.Dial("tcp", e.hostPort())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stalled.Close() })
	if _, err := stalled.Write([]byte("GET /ws HTT")); err != nil { // never finishes the request line
		t.Fatal(err)
	}
	eventually(t, "stalled socket counted", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 2 })

	budget := MinShutdownBudget
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	begin := time.Now()
	_ = e.srv.Shutdown(ctx) // may report the stalled socket; the sessions must still be drained
	if n := e.handler.ActiveSessions(); n != 0 {
		t.Fatalf("active sessions after Shutdown = %d; the stalled handshake starved the drain", n)
	}
	if d := time.Since(begin); d > budget+2*time.Second {
		t.Fatalf("Shutdown took %v; want within the %v budget plus a small margin", d, budget)
	}
}
