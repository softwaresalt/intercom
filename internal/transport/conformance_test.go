package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

const modulePath = "github.com/softwaresalt/intercom-go"

// The [remote] settings that config validates are the ones the transport
// consumes: defaults line up, and every allowed_origins / allowed_hosts entry
// that config accepts is accepted by the transport (config is the stricter gate).
func TestRemoteConfigToTransportWiringConformance(t *testing.T) {
	if config.DefaultRemoteMaxConnections != DefaultMaxConnections {
		t.Fatalf("max connections default drifted: config=%d transport=%d", config.DefaultRemoteMaxConnections, DefaultMaxConnections)
	}
	if DefaultMaxConnections != hub.DefaultMaxRemote {
		t.Fatalf("max connections default drifted: transport=%d hub=%d", DefaultMaxConnections, hub.DefaultMaxRemote)
	}

	base := config.Default()
	base.DefaultWorkspaceRoot = t.TempDir()
	if _, err := base.Validate(); err != nil {
		t.Skipf("default config does not validate in this environment: %v", err)
	}
	h, err := hub.New(hub.Config{SessionID: "sess"})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Stop()

	configAccepts := func(rc config.RemoteConfig) bool {
		c := config.Default()
		c.DefaultWorkspaceRoot = base.DefaultWorkspaceRoot
		c.Remote = rc
		_, err := c.Validate()
		return err == nil
	}
	transportAccepts := func(rc config.RemoteConfig) bool {
		r, err := NewRemote(RemoteOptions{Hub: h, Remote: rc})
		if err != nil {
			return false
		}
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = r.Shutdown(sctx)
		return true
	}

	origins := []string{
		"https://ui.example.test", "http://localhost:8080", "https://UI.example.test", "https://ui.example.test/",
		"https://ui.example.test/path", "ftp://ui.example.test", "https://user@ui.example.test", "", "null",
		"http://[::1]:3000", "https://ui.example.test:443",
	}
	for _, o := range origins {
		rc := config.Default().Remote
		rc.AllowedOrigins = []string{o}
		// config is the strict gate (exact canonical form); the transport also
		// canonicalizes. The contract is containment: nothing config accepts may
		// be refused by the transport.
		if c, tr := configAccepts(rc), transportAccepts(rc); c && !tr {
			t.Errorf("allowed_origins %q: config accepts but the transport refuses", o)
		}
	}
	hosts := []string{
		"example.test", "example.test:8080", "EXAMPLE.test", "[::1]:8080", "::1", "example.test:0",
		"example.test:65536", "exa mple.test", "", "127.0.0.1", "example.test:80", "https://example.test",
	}
	for _, hs := range hosts {
		rc := config.Default().Remote
		rc.AllowedHosts = []string{hs}
		if c, tr := configAccepts(rc), transportAccepts(rc); c && !tr {
			t.Errorf("allowed_hosts %q: config accepts but the transport refuses", hs)
		}
	}
}

// The standalone client duplicates the wire contract instead of importing the
// server, so drift is caught here: what wsclient sends decodes to the intended
// hub action, and what the server replies is recognized by wsclient.
func TestWireConstantParityBetweenServerAndWSClient(t *testing.T) {
	if wsclient.Path != WSPath || wsclient.Subprotocol != SubprotocolV1 {
		t.Fatalf("wsclient path/subprotocol = %q/%q; server = %q/%q", wsclient.Path, wsclient.Subprotocol, WSPath, SubprotocolV1)
	}
	bound := hub.SnapshotOverhead + hub.DefaultMaxSnapshotBytes +
		hub.DefaultMaxPendingPermissions*hub.PermissionEntryHydrateBytes
	if wsclient.ReadLimit < bound {
		t.Fatalf("wsclient.ReadLimit %d below the hub hydrate bound %d", wsclient.ReadLimit, bound)
	}
	if ReplyAccepted != "action.accepted" || ReplyRefused != "action.refused" {
		t.Fatalf("reply types changed: %q %q", ReplyAccepted, ReplyRefused)
	}

	// Server side: record every raw message, answer with real server replies.
	got := make(chan []byte, 8)
	replies := [][]byte{
		replyFor(hub.OperatorAction{Type: hub.OpPrompt}, hub.ActionResult{Accepted: true}),
		refusalFor(hub.OperatorAction{Type: hub.OpAbortTurn, RequestID: "r1"}, ReasonLocalOnlyAction),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{SubprotocolV1}})
		if err != nil {
			return
		}
		defer func() { _ = c.CloseNow() }()
		for range 4 {
			_, data, err := c.Read(r.Context())
			if err != nil {
				return
			}
			got <- data
		}
		for _, rp := range replies {
			if err := c.Write(r.Context(), websocket.MessageText, rp); err != nil {
				return
			}
		}
		_, _, _ = c.Read(r.Context())
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{
		URL: "ws://" + strings.TrimPrefix(srv.URL, "http://") + wsclient.Path, Credential: "x",
		Carriage: wsclient.CarriageSubprotocol,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Abort()
	for _, send := range []func() error{
		func() error { return c.Prompt(ctx, "hi") },
		func() error { return c.AbortTurn(ctx) },
		func() error { return c.ResolvePermission(ctx, "r1", true, "ignored") },
		func() error { return c.ResolvePermission(ctx, "r2", false, "no") },
	} {
		if err := send(); err != nil {
			t.Fatal(err)
		}
	}
	want := []hub.OperatorAction{
		{Type: hub.OpPrompt, Prompt: "hi"},
		{Type: hub.OpAbortTurn},
		{Type: hub.OpResolvePermission, RequestID: "r1", Decision: agent.PermissionDecision{Kind: agent.DecisionAllowOnce}},
		{Type: hub.OpResolvePermission, RequestID: "r2", Decision: agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "no"}},
	}
	for i, w := range want {
		act, err := decodeAction(<-got)
		if err != nil || act != w {
			t.Fatalf("message %d decoded to %+v, %v; want %+v", i, act, err, w)
		}
	}
	m, err := c.Next(ctx)
	if err != nil || m.Reply == nil || !m.Reply.Accepted() || m.Reply.Op != opPrompt {
		t.Fatalf("accepted reply = %+v, %v", m, err)
	}
	m, err = c.Next(ctx)
	if err != nil || m.Reply == nil || m.Reply.Accepted() || m.Reply.Reason != ReasonLocalOnlyAction || m.Reply.RequestID != "r1" {
		t.Fatalf("refused reply = %+v, %v", m, err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

func directImports(t *testing.T, root, pkg string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-f", `{{join .Imports "\n"}}`, pkg)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list %s: %v", pkg, err)
	}
	return strings.Fields(string(out))
}

// The layering the depguard rules enforce is also asserted here, so a
// misconfigured or removed linter rule cannot hide an import-direction break.
func TestTransportPackageLayeringAndDepguardConfig(t *testing.T) {
	root := moduleRoot(t)
	forbid := func(pkg string, denied ...string) {
		t.Helper()
		for _, imp := range directImports(t, root, modulePath+"/"+pkg) {
			for _, d := range denied {
				full := modulePath + "/" + d
				if imp == full || strings.HasPrefix(imp, full+"/") {
					t.Errorf("%s imports %s (forbidden)", pkg, imp)
				}
			}
		}
	}
	forbid("internal/transport", "internal/app", "internal/tui", "internal/agent/copilot", "github.com/github/copilot-sdk/go")
	forbid("internal/transport/wsclient", "internal/transport", "internal/tui", "internal/app", "internal/config", "internal/pathsafe", "internal/agent")
	for _, lower := range []string{"internal/hub", "internal/agent", "internal/tui", "internal/config", "internal/pathsafe"} {
		forbid(lower, "internal/transport")
	}

	cfg, err := os.ReadFile(filepath.Join(root, ".golangci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range []string{"transport-layering-boundary:", "wsclient-standalone-boundary:", "lower-layers-no-transport:", "copilot-sdk-boundary:"} {
		if !strings.Contains(string(cfg), rule) {
			t.Errorf(".golangci.yml does not define depguard rule %s", strings.TrimSuffix(rule, ":"))
		}
	}
}
