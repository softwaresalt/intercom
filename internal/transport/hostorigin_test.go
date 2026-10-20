package transport

import (
	"strings"
	"testing"
)

func TestCanonHostPort(t *testing.T) {
	cases := []struct {
		in         string
		host, port string
		ok         bool
	}{
		{"example.com", "example.com", "", true},
		{"EXAMPLE.com:8080", "example.com", "8080", true},
		{"example.com.", "example.com", "", true},
		{"example.com.:443", "example.com", "443", true},
		{"example.com..", "", "", false},
		{"example.com..:80", "", "", false},
		{".example.com", "", "", false},
		{"a..b", "", "", false},
		{".", "", "", false},
		{"127.0.0.1:9", "127.0.0.1", "9", true},
		{"[::1]:8080", "[::1]", "8080", true},
		{"[::1]", "[::1]", "", true},
		{"[0:0:0:0:0:0:0:1]:8080", "[::1]", "8080", true},
		{"[0000:0000:0000:0000:0000:0000:0000:0001]", "[::1]", "", true},
		{"[2001:0DB8:0:0:0:0:0:1]:1", "[2001:db8::1]", "1", true},
		{"[::1", "", "", false},
		{"::1", "", "", false}, // unbracketed IPv6
		{"[::1]x", "", "", false},
		{"[fe80::1%25eth0]:80", "", "", false}, // zone
		{"[127.0.0.1]:80", "", "", false},      // bracketed IPv4
		{"example.com:0", "", "", false},
		{"example.com:00", "", "", false},
		{"example.com:1", "example.com", "1", true},
		{"example.com:65535", "example.com", "65535", true},
		{"example.com:65536", "", "", false},
		{"example.com:080", "example.com", "80", true}, // leading zeros normalise
		{"example.com:0000080", "", "", false},         // over-long port text
		{"example.com:", "", "", false},                // empty port
		{"example.com:8a", "", "", false},
		{"example.com:+80", "", "", false},
		{"example.com:-1", "", "", false},
		{"example.com:8:9", "", "", false},
		{"user@example.com", "", "", false},
		{"example.com/path", "", "", false},
		{"exa mple.com", "", "", false},
		{"exämple.com", "", "", false},
		{"", "", "", false},
		{strings.Repeat("a", 300), "", "", false},
	}
	for _, tc := range cases {
		host, port, ok := canonHostPort(tc.in)
		if host != tc.host || port != tc.port || ok != tc.ok {
			t.Errorf("canonHostPort(%q) = (%q, %q, %v); want (%q, %q, %v)", tc.in, host, port, ok, tc.host, tc.port, tc.ok)
		}
	}
}

func TestHostKey(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"localhost:8080", "localhost:8080", true},
		{"LocalHost.:8080", "localhost:8080", true},
		{"[::1]:8080", "[::1]:8080", true},
		{"[0:0:0:0:0:0:0:1]:8080", "[::1]:8080", true},
		// a Host default port is NOT equated with an omitted port
		{"example.com:80", "example.com:80", true},
		{"example.com", "example.com", true},
		{"example.com:", "", false},
		{"example.com:0", "", false},
		{"example.com:65536", "", false},
		{"example.com..", "", false},
	}
	for _, tc := range cases {
		got, ok := hostKey(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("hostKey(%q) = (%q, %v); want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
	a, _ := hostKey("example.com:80")
	b, _ := hostKey("example.com")
	if a == b {
		t.Error("Host default port must not be equated with an omitted port")
	}
}

func TestOriginKey(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"https://app.example", "https://app.example", true},
		{"HTTPS://App.Example", "https://app.example", true},
		{"https://app.example:443", "https://app.example", true}, // default port dropped
		{"http://app.example:80", "http://app.example", true},
		{"http://app.example:443", "http://app.example:443", true},
		{"https://app.example:80", "https://app.example:80", true},
		{"https://app.example.", "https://app.example", true},
		{"https://app.example..", "", false},
		{"https://[::1]:8443", "https://[::1]:8443", true},
		{"http://[0:0:0:0:0:0:0:1]:3000", "http://[::1]:3000", true},
		{"http://[::1]:80", "http://[::1]", true},
		{"https://user@app.example", "", false}, // userinfo
		{"https://user:pw@app.example", "", false},
		{"https://app.example/", "", false},
		{"https://app.example/path", "", false},
		{"https://app.example?x=1", "", false},
		{"https://app.example:0", "", false},
		{"https://app.example:65536", "", false},
		{"https://app.example:", "", false},
		{"https://", "", false},
		{"null", "", false},
		{"ftp://app.example", "", false},
		{"app.example", "", false},
		{"https://*.example.com", "", false},
	}
	for _, tc := range cases {
		got, ok := originKey(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("originKey(%q) = (%q, %v); want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestNewHandlerRefusesWildcardAndNonCanonicalAllowlists(t *testing.T) {
	e := newEnv(t, envOpts{})
	base := func() Config { return Config{Hub: e.hub, Creds: e.creds} }
	for _, o := range []string{"*", "https://*.example.com", "null", "https://app.example/", "https://user@app.example", "app.example", "https://app.example:0"} {
		cfg := base()
		cfg.AllowedOrigins = []string{o}
		if _, err := NewHandler(cfg); err == nil {
			t.Errorf("NewHandler accepted allowed origin %q", o)
		}
	}
	for _, h := range []string{"*", "*.example.com", "example.com/", "user@example.com", "example.com:", "example.com:0", "example.com..", "::1"} {
		cfg := base()
		cfg.AllowedHosts = []string{h}
		if _, err := NewHandler(cfg); err == nil {
			t.Errorf("NewHandler accepted allowed host %q", h)
		}
	}
	cfg := base()
	cfg.AllowedOrigins = []string{"https://app.example", "http://[::1]:3000"}
	cfg.AllowedHosts = []string{"app.example", "[::1]:8080"}
	if _, err := NewHandler(cfg); err != nil {
		t.Fatalf("canonical entries refused: %v", err)
	}
}
