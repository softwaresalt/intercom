package transport

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func wantStatus(t *testing.T, resp *http.Response, want int, what string) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("%s: status = %d; want %d", what, resp.StatusCode, want)
	}
}

func bearer(tok string) map[string]string { return map[string]string{"Authorization": "Bearer " + tok} }

func TestUnauthenticatedRequestNeverUpgrades(t *testing.T) {
	e := newEnv(t, envOpts{})
	cases := map[string]map[string]string{
		"missing":          nil,
		"wrong":            bearer("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
		"empty bearer":     {"Authorization": "Bearer "},
		"bare scheme":      {"Authorization": "Bearer"},
		"basic scheme":     {"Authorization": "Basic " + e.app.Token.Reveal()},
		"two tokens":       {"Authorization": "Bearer " + e.app.Token.Reveal() + " extra"},
		"no space":         {"Authorization": "Bearer" + e.app.Token.Reveal()},
		"token only":       {"Authorization": e.app.Token.Reveal()},
		"cookie not used":  {"Cookie": "token=" + e.app.Token.Reveal()},
		"x-header":         {"X-Auth-Token": e.app.Token.Reveal()},
		"subprotocol only": {"Sec-WebSocket-Protocol": "bearer." + e.app.Token.Reveal()},
	}
	for name, hdr := range cases {
		t.Run(name, func(t *testing.T) {
			resp := e.rawUpgrade(hdr, "")
			wantStatus(t, resp, http.StatusUnauthorized, name)
			if resp.Header.Get("Sec-WebSocket-Accept") != "" || resp.Header.Get("Upgrade") != "" {
				t.Fatal("unauthenticated response carries upgrade headers")
			}
		})
	}
}

func TestAuthFailureResponsesAreIdentical(t *testing.T) {
	e := newEnv(t, envOpts{})
	// expired and revoked credentials
	exp := newEnv(t, envOpts{lifetime: time.Nanosecond})
	time.Sleep(5 * time.Millisecond)
	old := e.app
	e.app, _ = e.creds.Mint(2) // re-issue revokes `old`

	type shot struct {
		status int
		hdr    string
		body   string
	}
	take := func(env *env, hdr map[string]string) shot {
		r := env.rawUpgrade(hdr, "")
		var hs []string
		for k, v := range r.Header {
			if k == "Date" {
				continue
			}
			hs = append(hs, k+"="+strings.Join(v, ","))
		}
		return shot{r.StatusCode, strings.Join(sorted(hs), ";"), bodyOf(r)}
	}
	ref := take(e, nil)
	for name, got := range map[string]shot{
		"wrong":   take(e, bearer("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")),
		"revoked": take(e, bearer(old.Token.Reveal())),
		"expired": take(exp, bearer(exp.app.Token.Reveal())),
		"garbage": take(e, map[string]string{"Authorization": "Bearer !!!"}),
	} {
		if got != ref {
			t.Errorf("%s response differs from missing-credential response:\n got %+v\nwant %+v", name, got, ref)
		}
	}
	if ref.status != http.StatusUnauthorized {
		t.Fatalf("status = %d", ref.status)
	}
}

func sorted(s []string) []string {
	for i := range s {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
	return s
}

func TestNativeClientWithoutOriginAcceptedOnlyWithValidCredential(t *testing.T) {
	e := newEnv(t, envOpts{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	h := http.Header{"Authorization": {"Bearer " + e.app.Token.Reveal()}}
	c, resp, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{HTTPHeader: h})
	if err != nil {
		t.Fatalf("valid native client refused: %v", err)
	}
	defer func() { _ = c.CloseNow() }()
	wantStatus(t, resp, http.StatusSwitchingProtocols, "valid native")
	if got := resp.Header.Get("Sec-WebSocket-Protocol"); got != "" {
		t.Fatalf("a header-authenticated client negotiated protocol %q", got)
	}
	// without credential: 401 and no Origin involved
	wantStatus(t, e.rawUpgrade(nil, ""), http.StatusUnauthorized, "no origin no cred")
}

func TestCrossOriginUpgradeRejected(t *testing.T) {
	e := newEnv(t, envOpts{}) // empty allowlist: every browser origin rejected
	good := bearer(e.app.Token.Reveal())
	for _, origin := range []string{
		"https://evil.example", "http://127.0.0.1:" + "1", "null", "http://" + e.hostPort(), "https://evil.example:8443",
	} {
		hdr := map[string]string{"Authorization": good["Authorization"], "Origin": origin}
		wantStatus(t, e.rawUpgrade(hdr, ""), http.StatusForbidden, "origin "+origin)
	}
	// Cross-origin rejected even with no credential: Origin precedes credential.
	wantStatus(t, e.rawUpgrade(map[string]string{"Origin": "https://evil.example"}, ""), http.StatusForbidden, "origin no cred")
}

func TestOriginAllowlistExactMatch(t *testing.T) {
	e := newEnv(t, envOpts{origins: []string{"https://app.example", "http://localhost:5173"}})
	good := bearer(e.app.Token.Reveal())
	with := func(origin string) map[string]string {
		return map[string]string{"Authorization": good["Authorization"], "Origin": origin}
	}
	for _, bad := range []string{
		"https://app.example.evil.example", "http://app.example", "https://app.example:8443",
		"https://evilapp.example", "https://app.example/", "https://sub.app.example", "http://localhost:5174",
	} {
		wantStatus(t, e.rawUpgrade(with(bad), ""), http.StatusForbidden, bad)
	}
	// allowlisted origin without a credential is still 401, never an upgrade
	wantStatus(t, e.rawUpgrade(map[string]string{"Origin": "https://app.example"}, ""), http.StatusUnauthorized, "allowed origin no cred")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, ok := range []string{"https://app.example", "HTTPS://APP.EXAMPLE", "https://app.example:443", "https://app.example.", "http://localhost:5173"} {
		h := http.Header{}
		h.Set("Authorization", good["Authorization"])
		h.Set("Origin", ok)
		c, resp, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{HTTPHeader: h})
		if err != nil {
			t.Fatalf("allowlisted origin %q refused: %v", ok, err)
		}
		wantStatus(t, resp, http.StatusSwitchingProtocols, ok)
		_ = c.CloseNow()
	}
}

func TestHostHeaderValidated(t *testing.T) {
	e := newEnv(t, envOpts{hosts: []string{"tunnel.example", "other.example:9443"}})
	auth := bearer(e.app.Token.Reveal())
	port := e.hostPort()[strings.LastIndex(e.hostPort(), ":")+1:]
	bad := []string{
		"evil.example", "evil.example:" + port, "127.0.0.1:1", "localhost:1", "127.0.0.1", "", "rebind.127.0.0.1.nip.io:" + port,
		"tunnel.example:8080", "other.example", "tunnel.example@evil.example", "[::1]:1", "2130706433:" + port,
	}
	for _, h := range bad {
		if h == "" {
			continue // Go substitutes the URL host for an empty Host
		}
		hdr := map[string]string{"Authorization": auth["Authorization"], "Host": h}
		wantStatus(t, e.rawUpgrade(hdr, ""), http.StatusForbidden, "host "+h)
	}
	// Host is checked before credentials
	wantStatus(t, e.rawUpgrade(map[string]string{"Host": "evil.example"}, ""), http.StatusForbidden, "bad host no cred")

	for _, h := range []string{
		"127.0.0.1:" + port, "localhost:" + port, "LOCALHOST:" + port, "tunnel.example", "TUNNEL.example.", "other.example:9443",
	} {
		hdr := map[string]string{"Authorization": auth["Authorization"], "Host": h}
		wantStatus(t, e.rawUpgrade(hdr, ""), http.StatusSwitchingProtocols, "allowed host "+h)
	}
}

func TestCredentialInURLIsNeverAccepted(t *testing.T) {
	e := newEnv(t, envOpts{})
	tok := e.app.Token.Reveal()
	for _, q := range []string{"token=" + tok, "access_token=" + tok, "bearer=" + tok, "Authorization=Bearer%20" + tok, "key=" + tok, "credential=" + tok} {
		wantStatus(t, e.rawUpgrade(nil, q), http.StatusUnauthorized, "url only: "+q)
		// even alongside a valid header the leak is refused
		wantStatus(t, e.rawUpgrade(bearer(tok), q), http.StatusUnauthorized, "url+header: "+q)
	}
	// a benign query (cursor) with a header credential is fine
	wantStatus(t, e.rawUpgrade(bearer(tok), "gen="+e.hub.Gen()+"&seq=0"), http.StatusSwitchingProtocols, "benign query")
}

func TestSubprotocolCarriageSelectsOnlyNonSecretProtocol(t *testing.T) {
	e := newEnv(t, envOpts{})
	tok := e.app.Token.Reveal()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, e.wsURL(), &websocket.DialOptions{Subprotocols: []string{"intercom.v1", "bearer." + tok}})
	if err != nil {
		t.Fatalf("subprotocol carriage refused: %v", err)
	}
	defer func() { _ = c.CloseNow() }()
	if got := resp.Header.Get("Sec-WebSocket-Protocol"); got != "intercom.v1" {
		t.Fatalf("selected subprotocol = %q; want only intercom.v1", got)
	}
	for k, v := range resp.Header {
		if strings.Contains(strings.Join(v, ","), tok) {
			t.Fatalf("response header %s echoes the credential", k)
		}
	}
	if c.Subprotocol() != "intercom.v1" {
		t.Fatalf("negotiated = %q", c.Subprotocol())
	}

	// Order-independent, and a bad bearer is generic 401.
	r := e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "bearer." + tok + ", intercom.v1"}, "")
	wantStatus(t, r, http.StatusSwitchingProtocols, "reordered")
	r = e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "intercom.v1, bearer.wrong"}, "")
	wantStatus(t, r, http.StatusUnauthorized, "wrong bearer")
	// no intercom.v1 offered => same generic 401
	r = e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "bearer." + tok}, "")
	wantStatus(t, r, http.StatusUnauthorized, "no intercom.v1")
	// two bearer entries / header AND subprotocol => ambiguous => 401
	r = e.rawUpgrade(map[string]string{"Sec-WebSocket-Protocol": "intercom.v1, bearer." + tok + ", bearer." + tok}, "")
	wantStatus(t, r, http.StatusUnauthorized, "two bearers")
	r = e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok, "Sec-WebSocket-Protocol": "intercom.v1, bearer." + tok}, "")
	wantStatus(t, r, http.StatusUnauthorized, "ambiguous")
	// intercom.v1 alone plus header auth is fine and selects intercom.v1
	r = e.rawUpgrade(map[string]string{"Authorization": "Bearer " + tok, "Sec-WebSocket-Protocol": "intercom.v1"}, "")
	wantStatus(t, r, http.StatusSwitchingProtocols, "header + intercom.v1")
	if got := r.Header.Get("Sec-WebSocket-Protocol"); got != "intercom.v1" {
		t.Fatalf("selected %q", got)
	}
}

func TestOnlyWSPathServed(t *testing.T) {
	e := newEnv(t, envOpts{})
	req, _ := http.NewRequest(http.MethodGet, "http://"+e.hostPort()+"/other", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	wantStatus(t, resp, http.StatusNotFound, "other path")
}

func TestWebsocketLibraryOriginCheckIsNotDisabled(t *testing.T) {
	// The handler authorises Origin itself; the library must still run with
	// InsecureSkipVerify=false (verified by the exported constant the handler
	// builds its AcceptOptions from).
	if acceptOptions().InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify must be false")
	}
	if len(acceptOptions().OriginPatterns) != 0 {
		t.Fatal("OriginPatterns must be empty (handled by the handler's own pre-checks)")
	}
	if got := acceptOptions().Subprotocols; len(got) != 1 || got[0] != SubprotocolV1 {
		t.Fatalf("selectable subprotocols = %v; want only %s", got, SubprotocolV1)
	}
}
