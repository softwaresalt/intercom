package app

import (
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

func dialApp(t *testing.T, a *App, role hub.Role) *wsclient.Client {
	t.Helper()
	var tok string
	for _, c := range a.Credentials() {
		if c.Role == role {
			tok = c.Token.Reveal()
		}
	}
	if tok == "" {
		t.Fatalf("no credential minted for role %v", role)
	}
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: a.WSURL(), Credential: tok})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(c.Abort)
	return c
}

func TestStartReachesReady(t *testing.T) {
	f := newFactory()
	o, ws := baseOptions(t, f)
	a := startApp(t, o)

	root, err := pathsafe.NewRoot(ws)
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Workspace(); got != root.Path() {
		t.Errorf("Workspace() = %q; want canonical %q", got, root.Path())
	}
	if n := f.invoked(); n != 1 {
		t.Fatalf("agent factory invoked %d times; want exactly 1 (one session per process)", n)
	}
	fa := f.fake()
	if fa.deps.WorkingDirectory != root.Path() {
		t.Errorf("agent working directory = %q; want the pathsafe-resolved %q", fa.deps.WorkingDirectory, root.Path())
	}
	if fa.deps.Ingress == nil || fa.deps.Asks == nil || fa.deps.HubDone == nil {
		t.Error("agent was not wired to the hub (ingress/asks/done)")
	}
	if starts, _ := fa.counts(); starts != 1 {
		t.Errorf("agent started %d times; want 1", starts)
	}

	tcp, ok := a.Addr().(*net.TCPAddr)
	if !ok || !tcp.IP.IsLoopback() || tcp.Port == 0 {
		t.Fatalf("Addr() = %v; want a bound loopback TCP address", a.Addr())
	}
	if got, want := a.WSURL(), "ws://"+a.Addr().String()+"/ws"; got != want {
		t.Errorf("WSURL() = %q; want %q", got, want)
	}
	if n := len(a.Credentials()); n != 2 {
		t.Fatalf("minted %d credentials; want 2 (observer, approver)", n)
	}

	// The hub is running and draining the agent's ingress: an emitted event
	// reaches a remote client through the real hub and WebSocket.
	c := dialApp(t, a, hub.RoleApprover)
	fa.emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "hello from the agent"})
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	for {
		m, err := c.Next(ctx)
		if err != nil {
			t.Fatalf("waiting for the agent message: %v", err)
		}
		if m.Envelope != nil && m.Envelope.Kind == hub.KindAgentMessage {
			break
		}
	}
}

func TestStartRefusesBadInvocation(t *testing.T) {
	cases := map[string]func(t *testing.T) Options{
		"missing config file": func(t *testing.T) Options {
			f := newFactory()
			o, _ := baseOptions(t, f)
			o.ConfigPath = t.TempDir() + "/nope.toml"
			o.NewAgent = f.build
			return o
		},
		"malformed toml": func(t *testing.T) Options {
			f := newFactory()
			o, ws := baseOptions(t, f)
			o.ConfigPath = writeConfig(t, ws, "this is = = not toml", "")
			return o
		},
		"workspace root does not exist": func(t *testing.T) Options {
			f := newFactory()
			o, ws := baseOptions(t, f)
			o.ConfigPath = writeConfig(t, ws+"/missing", "", "")
			return o
		},
		"remote lifetime out of range": func(t *testing.T) Options {
			f := newFactory()
			o, ws := baseOptions(t, f)
			o.ConfigPath = writeConfig(t, ws, "", "[remote]\ntoken_lifetime_seconds = 1\n")
			return o
		},
		"credential-shaped remote key": func(t *testing.T) Options {
			f := newFactory()
			o, ws := baseOptions(t, f)
			o.ConfigPath = writeConfig(t, ws, "", "[remote]\napi_token = \"x\"\n")
			return o
		},
		"workspace outside every authorized root": func(t *testing.T) Options {
			f := newFactory()
			o, _ := baseOptions(t, f)
			o.Workspace = t.TempDir()
			return o
		},
		"no agent factory": func(t *testing.T) Options {
			f := newFactory()
			o, _ := baseOptions(t, f)
			o.NewAgent = nil
			return o
		},
	}
	for name, mk := range cases {
		t.Run(name, func(t *testing.T) {
			o := mk(t)
			// Count factory use through a wrapper so "nothing started" is checkable.
			var built int
			if inner := o.NewAgent; inner != nil {
				o.NewAgent = func(d AgentDeps) (agent.Agent, error) { built++; return inner(d) }
			}
			ctx, cancel := context.WithTimeout(context.Background(), waitFor)
			defer cancel()
			a, err := Start(ctx, o)
			if err == nil {
				stopApp(t, a)
				t.Fatal("Start succeeded; want a startup error (process exits non-zero)")
			}
			if a != nil {
				t.Error("Start returned a non-nil App alongside an error")
			}
			if built != 0 {
				t.Errorf("agent built %d times before the invocation was refused; nothing may start on invalid input", built)
			}
		})
	}
}

func TestPortFlagOverridesConfig(t *testing.T) {
	cfgPort := freePort(t)
	f := newFactory()
	o, ws := baseOptions(t, f)
	o.ConfigPath = writeConfig(t, ws, "http_port = "+itoa(cfgPort), "")

	o.PortSet = false
	a := startApp(t, o)
	if got := a.Addr().(*net.TCPAddr).Port; got != int(cfgPort) {
		t.Errorf("without --port the listener is on %d; want the configured %d", got, cfgPort)
	}

	f2 := newFactory()
	o.NewAgent = f2.build
	o.PortSet, o.Port = true, 0
	b := startApp(t, o)
	if got := b.Addr().(*net.TCPAddr).Port; got == int(cfgPort) || got == 0 {
		t.Errorf("--port 0 listener is on %d; want an ephemeral port distinct from the config's %d", got, cfgPort)
	}
}

func TestAgentStartFailureTearsEverythingDown(t *testing.T) {
	f := newFactory()
	f.startErr = errors.New("cli did not start")
	o, _ := baseOptions(t, f)
	port := freePort(t)
	o.Port = port

	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	a, err := Start(ctx, o)
	if err == nil {
		stopApp(t, a)
		t.Fatal("Start succeeded though the agent failed to start")
	}
	if !errors.Is(err, f.startErr) {
		t.Errorf("error %v does not wrap the agent start failure", err)
	}
	fa := f.fake()
	if fa == nil {
		t.Fatal("the agent was never built")
	}
	if _, stops := fa.counts(); stops != 1 {
		t.Errorf("agent Stop called %d times after a failed Start; want 1 (reap the CLI child)", stops)
	}
	select {
	case <-fa.deps.HubDone:
	case <-time.After(waitFor):
		t.Fatal("hub still running after a failed start")
	}
	if c, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", itoa(port)), time.Second); err == nil {
		_ = c.Close()
		t.Fatal("listener still accepting after a failed start")
	}
}

func TestLogsCarryNoCredentialMaterial(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	logs := &logBuf{}
	o.Logger = newTestLogger(logs)
	a := startApp(t, o)

	creds := a.Credentials()
	if len(creds) == 0 {
		t.Fatal("no credentials minted")
	}
	out := logs.String()
	if !strings.Contains(out, "ready") {
		t.Errorf("startup logged no ready message:\n%s", out)
	}
	for _, c := range creds {
		if tok := c.Token.Reveal(); tok == "" || strings.Contains(out, tok) {
			t.Fatalf("log output carries credential material (or the token was empty):\n%s", out)
		}
		if !strings.Contains(out, c.Fingerprint) {
			t.Errorf("log lacks the %v credential fingerprint %q", c.Role, c.Fingerprint)
		}
	}
}

func itoa(p uint16) string { return strconv.Itoa(int(p)) }
