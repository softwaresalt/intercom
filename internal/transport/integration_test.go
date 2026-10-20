package transport

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

func refOpts(e *env, tok string) wsclient.Options {
	return wsclient.Options{URL: e.wsURL(), Credential: tok}
}

func dialRef(t *testing.T, e *env, o wsclient.Options) *wsclient.Client {
	t.Helper()
	if o.URL == "" {
		o.URL = e.wsURL()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := wsclient.Dial(ctx, o)
	if err != nil {
		t.Fatalf("reference client dial: %v", err)
	}
	t.Cleanup(c.Abort)
	return c
}

func nextMsg(t *testing.T, c *wsclient.Client) wsclient.Message {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m, err := c.Next(ctx)
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	return m
}

func handshakeStatus(t *testing.T, err error) int {
	t.Helper()
	var he *wsclient.HandshakeError
	if !errors.As(err, &he) {
		t.Fatalf("error = %v; want a handshake refusal", err)
	}
	return he.Status
}

func TestIntegrationSlowClientDropped(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) {
		c.ClientBuffer = 32 // large enough that the (reading) fast client is never the one dropped
		c.WriteTimeout = 400 * time.Millisecond
		c.CloseGrace = 300 * time.Millisecond
	}})
	slowOpts := refOpts(e, e.obs.Token.Reveal())
	slowOpts.Carriage = wsclient.CarriageSubprotocol
	slow := dialRef(t, e, slowOpts) // never calls Next
	fast := dialRef(t, e, refOpts(e, e.app.Token.Reveal()))

	var got atomic.Int64
	fastDone := make(chan struct{})
	go func() {
		defer close(fastDone)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			m, err := fast.Next(ctx)
			cancel()
			if err != nil {
				return
			}
			if m.Envelope != nil && m.Envelope.Kind == hub.KindAgentMessage && m.Verdict == wsclient.VerdictApply {
				got.Add(1)
			}
		}
	}()

	big := strings.Repeat("y", 15<<10)
	sent := 0
	for deadline := time.Now().Add(20 * time.Second); time.Now().Before(deadline); {
		begin := time.Now()
		e.event(big)
		sent++
		if time.Since(begin) > time.Second {
			t.Fatal("hub blocked by a slow client")
		}
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
	eventually(t, "slow client dropped, fast client kept", 8*time.Second, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		st, err := e.hub.Stats(ctx)
		return err == nil && st.DroppedClients == 1 && st.Remote == 1 && e.handler.ActiveSessions() == 1
	})
	eventually(t, "fast client received every event", 10*time.Second, func() bool { return got.Load() >= int64(sent) })

	// the slow client's socket is dead; reading drains what the kernel buffered, then fails
	rctx, rcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer rcancel()
	var rerr error
	for rerr == nil {
		_, rerr = slow.Next(rctx)
	}
	if errors.Is(rerr, context.DeadlineExceeded) {
		t.Fatalf("slow client socket never closed: %v", rerr)
	}
	fast.Abort()
	<-fastDone
}

func TestIntegrationCSWSHRejectedBeforeRegistration(t *testing.T) {
	e := newEnv(t, envOpts{origins: []string{"https://app.example"}})
	for _, carriage := range []wsclient.Carriage{wsclient.CarriageHeader, wsclient.CarriageSubprotocol} {
		o := refOpts(e, e.app.Token.Reveal()) // a VALID credential: only the Origin is wrong
		o.Carriage = carriage
		o.Origin = "https://evil.example"
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := wsclient.Dial(ctx, o)
		cancel()
		if got := handshakeStatus(t, err); got != http.StatusForbidden {
			t.Fatalf("cross-origin status = %d; want 403", got)
		}
		if strings.Contains(err.Error(), e.app.Token.Reveal()) {
			t.Fatal("error leaks the credential")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	st, err := e.hub.Stats(ctx)
	if err != nil || st.Remote != 0 || e.handler.ActiveSessions() != 0 {
		t.Fatalf("a rejected cross-origin upgrade registered with the hub: %+v %v", st, err)
	}
	// the allowlisted origin works with the credential
	o := refOpts(e, e.app.Token.Reveal())
	o.Origin = "https://app.example"
	c := dialRef(t, e, o)
	if m := nextMsg(t, c); m.Envelope == nil || m.Envelope.Kind != hub.KindSessionHydrate {
		t.Fatalf("first frame = %+v", m)
	}
}

func TestIntegrationExpiredWrongAndWrongRoleCredentials(t *testing.T) {
	clk := newFakeClockAt(time.Now())
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{clock: clk, hubCfg: hub.Config{Agent: fa}})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := wsclient.Dial(ctx, refOpts(e, "A"+e.app.Token.Reveal()[1:])); handshakeStatus(t, err) != http.StatusUnauthorized {
		t.Fatal("wrong credential not refused with 401")
	}

	// wrong role: an observer's approve and prompt are refused with a reason
	obs := dialRef(t, e, refOpts(e, e.obs.Token.Reveal()))
	nextMsg(t, obs) // hydrate
	sctx, scancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer scancel()
	if err := obs.ResolvePermission(sctx, "nope", true, ""); err != nil {
		t.Fatal(err)
	}
	if err := obs.Prompt(sctx, "hi"); err != nil {
		t.Fatal(err)
	}
	var refusals []string
	for len(refusals) < 2 {
		m := nextMsg(t, obs)
		if m.Reply != nil {
			refusals = append(refusals, m.Reply.Reason)
		}
	}
	if refusals[1] != string(hub.ReasonRoleNotPermitted) {
		t.Fatalf("observer prompt refusal = %v", refusals)
	}
	if prompts, _ := fa.snapshot(); len(prompts) != 0 {
		t.Fatal("observer prompt reached the agent")
	}

	// expired: the same credential after its lifetime is a generic 401
	clk.Advance(2 * time.Hour)
	if _, err := wsclient.Dial(ctx, refOpts(e, e.app.Token.Reveal())); handshakeStatus(t, err) != http.StatusUnauthorized {
		t.Fatal("expired credential not refused with 401")
	}
}

func TestIntegrationMidConnectionExpiryClosesWithPolicyViolation(t *testing.T) {
	e := newEnv(t, envOpts{lifetime: 1500 * time.Millisecond}) // generous: only the dial must beat it
	c := dialRef(t, e, refOpts(e, e.obs.Token.Reveal()))
	nextMsg(t, c)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := c.Next(ctx)
	code, reason, ok := wsclient.CloseStatus(err)
	if !ok || code != 1008 {
		t.Fatalf("expiry close = %v (code %d %q)", err, code, reason)
	}
	if strings.Contains(strings.ToLower(reason), "expir") {
		t.Fatalf("reason %q must be neutral", reason)
	}
}

func TestIntegrationReconnectReplayThenHydrate(t *testing.T) {
	e := newEnv(t, envOpts{hubCfg: hub.Config{ReplayRing: 4}})
	c1 := dialRef(t, e, refOpts(e, e.obs.Token.Reveal()))
	if m := nextMsg(t, c1); m.Verdict != wsclient.VerdictApply || m.Envelope.Kind != hub.KindSessionHydrate {
		t.Fatalf("first = %+v", m)
	}
	for i := 0; i < 3; i++ {
		e.event("a")
	}
	for i := 0; i < 3; i++ {
		if m := nextMsg(t, c1); m.Verdict != wsclient.VerdictApply {
			t.Fatalf("live frame verdict = %v", m.Verdict)
		}
	}
	pos, ok := c1.Position()
	if !ok {
		t.Fatal("no resumable position after applying frames")
	}
	c1.Abort() // a network blip
	e.event("b")
	e.event("c")

	o := refOpts(e, e.obs.Token.Reveal())
	o.Resume = &pos
	c2 := dialRef(t, e, o)
	for i := uint64(1); i <= 2; i++ {
		m := nextMsg(t, c2)
		if m.Envelope.Kind == hub.KindSessionHydrate || m.Envelope.Seq != pos.Seq+i || m.Verdict != wsclient.VerdictApply {
			t.Fatalf("replay frame %d = %+v (verdict %v)", i, m.Envelope, m.Verdict)
		}
	}
	c2.Abort()

	// a cursor from another generation hydrates
	o.Resume = &hub.Position{Gen: "00000000-0000-7000-8000-000000000000", Seq: pos.Seq}
	c3 := dialRef(t, e, o)
	if m := nextMsg(t, c3); m.Envelope.Kind != hub.KindSessionHydrate {
		t.Fatalf("foreign gen got %s; want hydrate", m.Envelope.Kind)
	}
	c3.Abort()

	// the ring no longer covers the old position: hydrate
	for i := 0; i < 10; i++ {
		e.event("d")
	}
	o.Resume = &pos
	c4 := dialRef(t, e, o)
	if m := nextMsg(t, c4); m.Envelope.Kind != hub.KindSessionHydrate {
		t.Fatalf("uncovered position got %s; want hydrate", m.Envelope.Kind)
	}
}

func TestIntegrationSubprotocolCarriageAndApproverFlow(t *testing.T) {
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{hubCfg: hub.Config{Agent: fa}})
	o := refOpts(e, e.app.Token.Reveal())
	o.Carriage = wsclient.CarriageSubprotocol
	c := dialRef(t, e, o)
	nextMsg(t, c)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.Prompt(ctx, "ship it"); err != nil {
		t.Fatal(err)
	}
	for {
		if m := nextMsg(t, c); m.Reply != nil {
			if !m.Reply.Accepted() {
				t.Fatalf("reply = %+v", m.Reply)
			}
			break
		}
	}
	if prompts, _ := fa.snapshot(); len(prompts) != 1 || prompts[0] != "ship it" {
		t.Fatalf("prompts = %v", prompts)
	}
}
