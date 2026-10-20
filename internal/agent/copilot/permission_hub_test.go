package copilot

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// ---- 067.002-T: the REAL permission handler against a REAL hub ----

// permRig runs a real hub, wires the real adapter's ask channel and hub-done
// channel to it, and parks the SDK permission handler in a goroutine.
type permRig struct {
	hub     *hub.Hub
	h       *harness
	tui     *hub.Client
	cancel  context.CancelFunc
	runDone chan error
}

func newPermRig(t *testing.T, cfg hub.Config) *permRig {
	t.Helper()
	cfg.SessionID = "sess-1"
	hb, err := hub.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r := &permRig{hub: hb, runDone: make(chan error, 1)}
	r.h = newHarness(t, func(c *Config, _ *fakeClient) {
		c.Ingress = hb.Ingress()
		c.Asks = hb.Asks()
		c.HubDone = hb.Done()
	})
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go func() { r.runDone <- hb.Run(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.runDone:
		case <-time.After(intWait):
			t.Error("hub Run did not return")
		}
	})
	if _, err := r.h.ad.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	conn, err := hb.Connect(context.Background(), hub.ConnectOptions{Kind: hub.ClientTUI, Buffer: 64})
	if err != nil {
		t.Fatal(err)
	}
	r.tui = conn
	<-conn.Frames() // initial hydrate
	return r
}

// park invokes the SDK permission handler as the SDK would, on its own
// goroutine, and reports its eventual decision.
func (r *permRig) park() <-chan rpc.PermissionDecision {
	res := make(chan rpc.PermissionDecision, 1)
	handler := r.h.client.createCfg.OnPermissionRequest
	go func() {
		d, _ := handler(&rpc.PermissionRequestShell{FullCommandText: "echo hi"}, copilotsdk.PermissionInvocation{})
		res <- d
	}()
	return res
}

func (r *permRig) nextOfKind(t *testing.T, kind hub.Kind) hub.Envelope {
	t.Helper()
	deadline := time.After(intWait)
	for {
		select {
		case f := <-r.tui.Frames():
			if f.Kind() != kind {
				continue
			}
			var env hub.Envelope
			if err := json.Unmarshal(f.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			return env
		case <-deadline:
			t.Fatalf("no %s envelope", kind)
		}
	}
}

func TestHubTimeoutAnswersTheRealHandlerAndResolvedFollowsItsReturn(t *testing.T) {
	r := newPermRig(t, hub.Config{PermissionTimeout: 150 * time.Millisecond})
	res := r.park()
	r.nextOfKind(t, hub.KindPermissionRequested)

	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("handler returned %q, want user-not-available from the hub timeout", d.Kind())
		}
	case <-time.After(intWait):
		t.Fatal("hub timeout never released the handler")
	}
	env := r.nextOfKind(t, hub.KindPermissionResolved)
	var p hub.PermissionResolvedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Decision != "unavailable" || p.Actor != hub.ActorHub || p.Reason != "timeout" {
		t.Fatalf("resolved = %+v", p)
	}
}

func TestHandlerStaysParkedUntilTheHubDecides(t *testing.T) {
	r := newPermRig(t, hub.Config{PermissionTimeout: time.Hour})
	res := r.park()
	r.nextOfKind(t, hub.KindPermissionRequested)
	select {
	case d := <-res:
		t.Fatalf("handler returned %q on its own deadline; only the hub may decide", d.Kind())
	case <-time.After(200 * time.Millisecond):
	}
}

func TestAdapterStopWhileRequestPendingEmitsAborted(t *testing.T) {
	r := newPermRig(t, hub.Config{PermissionTimeout: time.Hour})
	res := r.park()
	r.nextOfKind(t, hub.KindPermissionRequested)

	if err := r.h.ad.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	select {
	case d := <-res:
		if d.Kind() != rpc.PermissionDecisionKindUserNotAvailable {
			t.Fatalf("handler returned %q after Stop", d.Kind())
		}
	case <-time.After(intWait):
		t.Fatal("Stop did not release the handler")
	}
	env := r.nextOfKind(t, hub.KindPermissionAborted)
	var p hub.PermissionAbortedPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Reason != "handler_gone" {
		t.Fatalf("aborted = %+v", p)
	}
}

// The handler must never carry a deadline of its own (design §5.4 item 4):
// two clocks in two goroutines produce a silent authorization divergence. This
// guard fails if any timer or deadline construct appears in the handler.
func TestPermissionHandlerContainsNoTimerOrDeadline(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "permission.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	banned := map[string]bool{
		"time.After": true, "time.NewTimer": true, "time.Sleep": true, "time.AfterFunc": true,
		"time.Tick": true, "context.WithTimeout": true, "context.WithDeadline": true,
	}
	found := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "handle" || fn.Recv == nil {
			continue
		}
		found = true
		ast.Inspect(fn, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if id, ok := sel.X.(*ast.Ident); ok && banned[id.Name+"."+sel.Sel.Name] {
				t.Errorf("permissionBridge.handle uses %s.%s: the hub owns the timeout", id.Name, sel.Sel.Name)
			}
			return true
		})
	}
	if !found {
		t.Fatal("permissionBridge.handle not found")
	}
}
