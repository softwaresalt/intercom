package app

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

// uiRig is an app with the real TUI attached to injected pipes, plus a
// captured log stream.
type uiRig struct {
	app  *App
	fac  *factory
	out  *syncBuf
	keys io.Writer
	logs *logBuf
}

func (r *uiRig) press(t *testing.T, s string) {
	t.Helper()
	if _, err := r.keys.Write([]byte(s)); err != nil {
		t.Fatalf("send keys: %v", err)
	}
}

func newUIRig(t *testing.T) *uiRig {
	t.Helper()
	f := newFactory()
	o, _ := baseOptions(t, f)
	in, inW := io.Pipe()
	out := &syncBuf{}
	logs := &logBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	o.TUI = TUIOptions{Input: in, Output: out, Width: 100, Height: 30, RetainPlaintext: true}
	o.Logger = newTestLogger(logs)
	a := startApp(t, o)
	return &uiRig{app: a, fac: f, out: out, keys: inW, logs: logs}
}

func tokens(a *App) []string {
	var out []string
	for _, c := range a.Credentials() {
		out = append(out, c.Token.Reveal())
	}
	return out
}

// TestCredentialsShownOnLocalTUIOnly: the minted credentials are displayed on
// the local TUI, and appear nowhere else — not in logs, not in the hub's
// hydrate payload (the bytes a remote client receives), not in later
// envelopes.
func TestCredentialsShownOnLocalTUIOnly(t *testing.T) {
	r := newUIRig(t)
	toks := tokens(r.app)
	if len(toks) != 2 {
		t.Fatalf("minted %d credentials; want 2", len(toks))
	}
	for _, tok := range toks {
		waitOutput(t, r.out, tok) // shown on the local TUI
	}
	waitOutput(t, r.out, r.app.WSURL())

	// A remote client's hydrate (raw wire bytes) and live traffic.
	c := dialApp(t, r.app, hub.RoleObserver)
	r.fac.fake().emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: "ordinary agent text"})
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	var raw strings.Builder
	sawHydrate, sawMessage := false, false
	for !sawHydrate || !sawMessage {
		m, err := c.Next(ctx)
		if err != nil {
			t.Fatalf("reading frames: %v", err)
		}
		raw.Write(m.Raw)
		if m.Envelope != nil {
			sawHydrate = sawHydrate || m.Envelope.Kind == hub.KindSessionHydrate
			sawMessage = sawMessage || m.Envelope.Kind == hub.KindAgentMessage
		}
	}
	// The in-process hub hydrate (what the TUI itself is given).
	hc, err := r.app.Hub().Connect(ctx, hub.ConnectOptions{Kind: hub.ClientRemote, Role: hub.RoleObserver})
	if err != nil {
		t.Fatal(err)
	}
	defer hc.Close()
	select {
	case f := <-hc.Frames():
		raw.Write(f.Bytes())
	case <-time.After(waitFor):
		t.Fatal("no hydrate frame from the hub")
	}

	for _, tok := range toks {
		if strings.Contains(raw.String(), tok) {
			t.Fatal("a credential appears in hydrate or envelope bytes")
		}
		if strings.Contains(r.logs.String(), tok) {
			t.Fatalf("a credential appears in the logs:\n%s", r.logs.String())
		}
	}
	for _, c := range r.app.Credentials() {
		if !strings.Contains(r.logs.String(), c.Fingerprint) {
			t.Errorf("logs lack the %v fingerprint", c.Role)
		}
	}
}

// TestReissueThroughTUIMintsNewAndRevokesOld presses the re-issue key on the
// real TUI: a new credential set is minted and shown, the old one stops
// working (new dials refused, live sockets closed), and logs carry only
// fingerprints.
func TestReissueThroughTUIMintsNewAndRevokesOld(t *testing.T) {
	r := newUIRig(t)
	before := r.app.Credentials()
	oldTokens := tokens(r.app)
	for _, tok := range oldTokens {
		waitOutput(t, r.out, tok)
	}
	live := dialApp(t, r.app, hub.RoleApprover) // a socket authenticated by the old approver credential

	r.press(t, "\x12") // Ctrl+R

	var after []string
	deadline := time.Now().Add(waitFor)
	for {
		after = tokens(r.app)
		// Both roles must differ: the set is published role by role, so the
		// approver alone changing is an incomplete re-issue.
		if len(after) == 2 && after[0] != oldTokens[0] && after[1] != oldTokens[1] {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("re-issue through the TUI never minted a new credential")
		}
		time.Sleep(5 * time.Millisecond)
	}
	for _, tok := range after {
		waitOutput(t, r.out, tok) // the new credentials are shown
	}
	for i, c := range r.app.Credentials() {
		if c.Fingerprint == before[i].Fingerprint {
			t.Errorf("%v fingerprint unchanged after re-issue", c.Role)
		}
	}

	// Old credential: refused on a fresh dial.
	dctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	_, err := wsclient.Dial(dctx, wsclient.Options{URL: r.app.WSURL(), Credential: oldTokens[0]})
	var he *wsclient.HandshakeError
	if !errors.As(err, &he) || he.Status != 401 {
		t.Fatalf("dial with the revoked credential: %v; want a 401 handshake refusal", err)
	}
	// ... and its live socket is closed by the transport.
	for {
		_, nerr := live.Next(dctx)
		if nerr != nil {
			if errors.Is(nerr, context.DeadlineExceeded) {
				t.Fatal("the socket of the revoked credential stayed open")
			}
			break
		}
	}
	// The new credential works.
	nc, err := wsclient.Dial(dctx, wsclient.Options{URL: r.app.WSURL(), Credential: after[0]})
	if err != nil {
		t.Fatalf("dial with the re-issued credential: %v", err)
	}
	nc.Abort()

	for _, tok := range append(oldTokens, after...) {
		if strings.Contains(r.logs.String(), tok) {
			t.Fatalf("a credential appears in the logs:\n%s", r.logs.String())
		}
	}
	if !strings.Contains(r.logs.String(), "credentials re-issued") {
		t.Errorf("re-issue was not logged:\n%s", r.logs.String())
	}
}

func TestReissueMintsBothRolesAndRevokesBoth(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	a := startApp(t, o)
	old := a.Credentials()
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	fresh, err := a.Reissue(ctx)
	if err != nil {
		t.Fatalf("Reissue: %v", err)
	}
	if len(fresh) != 2 {
		t.Fatalf("Reissue minted %d; want 2", len(fresh))
	}
	for i := range fresh {
		if fresh[i].Token.Reveal() == old[i].Token.Reveal() {
			t.Errorf("%v credential was not replaced", fresh[i].Role)
		}
	}
	if got := a.Credentials(); got[0].Fingerprint != fresh[0].Fingerprint {
		t.Error("Credentials() does not return the re-issued set")
	}
	for _, o := range old {
		_, err := wsclient.Dial(ctx, wsclient.Options{URL: a.WSURL(), Credential: o.Token.Reveal()})
		if err == nil {
			t.Errorf("revoked %v credential still authenticates", o.Role)
		}
	}
}
