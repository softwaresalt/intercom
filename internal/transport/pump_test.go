package transport

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

func authHeader(tok string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+tok)
	return h
}

func dialWS(t *testing.T, e *env, tok string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{HTTPHeader: authHeader(tok)})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

func readEnvelope(t *testing.T, c *websocket.Conn, within time.Duration) hub.Envelope {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), within)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var env hub.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		t.Fatalf("frame is not a JSON envelope: %v: %q", err, data)
	}
	return env
}

// TestOnlyPumpWritesOrCloses is a structural guard of design §5.5: every call
// of Write, Ping, Close or CloseNow on the socket in this package's non-test
// source must sit inside the write pump (or its close helper).
func TestOnlyPumpWritesOrCloses(t *testing.T) {
	allowed := map[string]bool{"writePump": true, "writeFrame": true, "pingPeer": true, "closeSocket": true, "fail": true}
	forbidden := map[string]bool{"Write": true, "Ping": true, "Close": true, "CloseNow": true, "CloseRead": true, "Subprotocol": false}
	fset := token.NewFileSet()
	files, _ := filepath.Glob("*.go")
	checked := 0
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				// closures inside an allowed func count as inside it, except
				// for the abort-after-done timer, which is an explicit CloseNow.
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || !forbidden[sel.Sel.Name] {
					return true
				}
				recv := ""
				if inner, ok := sel.X.(*ast.SelectorExpr); ok {
					recv = inner.Sel.Name
				} else if id, ok := sel.X.(*ast.Ident); ok {
					recv = id.Name
				}
				if recv != "sock" && recv != "conn" {
					return true
				}
				checked++
				if !allowed[fd.Name.Name] {
					t.Errorf("%s: %s.%s called in %s; only the write pump may write or close the socket",
						f, recv, sel.Sel.Name, fd.Name.Name)
				}
				return true
			})
		}
	}
	if checked == 0 {
		t.Fatal("guard matched no socket calls; the receiver naming convention changed")
	}
}

func TestPumpDeliversFramesInOrderWithConcurrentPings(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) { c.PingInterval = 20 * time.Millisecond }})
	c := dialWS(t, e, e.app.Token.Reveal())
	first := readEnvelope(t, c, 3*time.Second)
	if first.Kind != hub.KindSessionHydrate {
		t.Fatalf("first frame = %s; want session.hydrate", first.Kind)
	}
	const n = 150
	go func() {
		for i := 0; i < n; i++ {
			e.event("m")
			if i%25 == 0 {
				time.Sleep(15 * time.Millisecond) // let pings interleave
			}
		}
	}()
	var last uint64
	for i := 0; i < n; i++ {
		env := readEnvelope(t, c, 5*time.Second)
		if env.Seq != last+1 && last != 0 {
			t.Fatalf("seq %d after %d: pings corrupted or reordered the stream", env.Seq, last)
		}
		last = env.Seq
	}
	if got := e.handler.PingsSent(); got < 2 {
		t.Fatalf("pings sent = %d; the ticker case never fired in the pump", got)
	}
}

func TestSlowClientIsDroppedAndNeverBlocksHubOrOthers(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) {
		c.ClientBuffer = 4
		c.WriteTimeout = 400 * time.Millisecond
		c.CloseGrace = 300 * time.Millisecond
	}})
	slow := dialWS(t, e, e.obs.Token.Reveal()) // never reads
	fast := dialWS(t, e, e.app.Token.Reveal())

	var wg sync.WaitGroup
	received := make(chan int, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		count := 0
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, _, err := fast.Read(ctx)
			cancel()
			if err != nil {
				received <- count
				return
			}
			count++
		}
	}()

	big := strings.Repeat("x", 15<<10)
	sent := 0
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		begin := time.Now()
		e.event(big)
		if d := time.Since(begin); d > time.Second {
			t.Fatalf("hub Send blocked for %v because of a slow client", d)
		}
		sent++
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		st, err := e.hub.Stats(ctx)
		cancel()
		if err == nil && st.DroppedClients >= 1 {
			break
		}
		if sent%50 == 0 {
			time.Sleep(5 * time.Millisecond)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	st, err := e.hub.Stats(ctx)
	cancel()
	if err != nil || st.DroppedClients < 1 {
		t.Fatalf("slow client was never dropped (stats=%+v err=%v) after %d events", st, err, sent)
	}
	// the dropped client's session ends (socket closed) while the fast one lives on
	eventually(t, "slow session closed", 5*time.Second, func() bool { return e.handler.ActiveSessions() == 1 })
	e.event("after-drop")

	// slow client eventually sees its socket die
	rctx, rcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer rcancel()
	var rerr error
	for rerr == nil {
		_, _, rerr = slow.Read(rctx)
	}
	if errors.Is(rerr, context.DeadlineExceeded) {
		t.Fatalf("slow client's socket was never closed: %v", rerr)
	}
	_ = fast.Close(websocket.StatusNormalClosure, "")
	wg.Wait()
	if n := <-received; n < sent {
		t.Fatalf("fast client received %d of >=%d events: a slow client starved it", n, sent)
	}
}

func TestExpiryClosesSocketViaPumpWithPolicyViolation(t *testing.T) {
	e := newEnv(t, envOpts{lifetime: 400 * time.Millisecond})
	c := dialWS(t, e, e.obs.Token.Reveal()) // observe-only: no actions to trip re-validation
	readEnvelope(t, c, 2*time.Second)       // hydrate
	begin := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	var ce websocket.CloseError
	if !errors.As(err, &ce) {
		t.Fatalf("read error = %v; want a close frame", err)
	}
	if ce.Code != websocket.StatusPolicyViolation {
		t.Fatalf("close code = %d; want policy violation", ce.Code)
	}
	if strings.Contains(strings.ToLower(ce.Reason), "expire") || strings.Contains(ce.Reason, e.obs.Token.Reveal()) {
		t.Fatalf("close reason %q must be neutral", ce.Reason)
	}
	if d := time.Since(begin); d > 3*time.Second {
		t.Fatalf("expiry close took %v", d)
	}
	eventually(t, "session gone", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 0 })
}

func TestReissueClosesConnectionViaPump(t *testing.T) {
	e := newEnv(t, envOpts{})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	if _, err := e.creds.Mint(hub.RoleApprover); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	var ce websocket.CloseError
	if !errors.As(err, &ce) || ce.Code != websocket.StatusPolicyViolation {
		t.Fatalf("read error = %v; want policy-violation close", err)
	}
}

func TestSessionsLeaveNoGoroutinesAfterShutdown(t *testing.T) {
	base := runtime.NumGoroutine()
	e := newEnv(t, envOpts{})
	var clients []*websocket.Conn
	for i := 0; i < 4; i++ {
		c := dialWS(t, e, e.app.Token.Reveal())
		readEnvelope(t, c, 2*time.Second)
		clients = append(clients, c)
	}
	e.event("x")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.srv.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if n := e.handler.ActiveSessions(); n != 0 {
		t.Fatalf("%d sessions outlived Shutdown", n)
	}
	for _, c := range clients {
		_ = c.CloseNow()
	}
	e.hub.Stop()
	eventually(t, "goroutines to return to baseline", 5*time.Second, func() bool {
		return runtime.NumGoroutine() <= base+2
	})
}
