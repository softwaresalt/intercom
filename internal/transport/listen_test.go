package transport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

func fakeResolver(addrs ...string) resolveFunc {
	return func(_ context.Context, _ string) ([]netip.Addr, error) {
		out := make([]netip.Addr, 0, len(addrs))
		for _, a := range addrs {
			out = append(out, netip.MustParseAddr(a))
		}
		return out, nil
	}
}

func TestListenRefusesNonLoopbackBind(t *testing.T) {
	cases := []struct {
		name     string
		host     string
		resolver resolveFunc
	}{
		{"ipv4 any", "0.0.0.0", nil},
		{"ipv6 any", "::", nil},
		{"empty host", "", nil},
		{"private ipv4", "192.168.1.10", nil},
		{"public ipv4", "8.8.8.8", nil},
		{"ipv6 global", "2001:db8::1", nil},
		{"mapped non-loopback", "::ffff:8.8.8.8", nil},
		{"zoned ipv6", "fe80::1%lo0", nil},
		{"hostname non-loopback", "evil.example", fakeResolver("203.0.113.9")},
		{"hostname mixed", "mixed.example", fakeResolver("127.0.0.1", "203.0.113.9")},
		{"hostname unresolvable", "nx.example", fakeResolver()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := listen(context.Background(), tc.host, 0, tc.resolver)
			if err == nil {
				_ = ln.Close()
				t.Fatalf("listen(%q) succeeded; non-loopback bind must be refused", tc.host)
			}
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Kind() != apperr.KindPolicy {
				t.Fatalf("listen(%q) error = %v; want apperr KindPolicy", tc.host, err)
			}
		})
	}
}

func TestListenBindsLoopback(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "localhost"} {
		t.Run(host, func(t *testing.T) {
			ln, err := listen(context.Background(), host, 0, nil)
			if err != nil {
				t.Skipf("loopback %s unavailable on this host: %v", host, err)
			}
			defer func() { _ = ln.Close() }()
			if err := AssertLoopbackAddr(ln.Addr()); err != nil {
				t.Fatalf("bound address %v is not loopback: %v", ln.Addr(), err)
			}
		})
	}
}

func TestAssertLoopbackAddr(t *testing.T) {
	bad := []net.Addr{
		&net.TCPAddr{IP: net.IPv4zero, Port: 1},
		&net.TCPAddr{IP: net.IPv6unspecified, Port: 1},
		&net.TCPAddr{IP: net.ParseIP("10.0.0.1"), Port: 1},
		&net.TCPAddr{IP: nil, Port: 1},
		&net.UnixAddr{Name: "x", Net: "unix"},
	}
	for _, a := range bad {
		if err := AssertLoopbackAddr(a); err == nil {
			t.Errorf("AssertLoopbackAddr(%v) = nil; want error", a)
		}
	}
	good := []net.Addr{
		&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1},
		&net.TCPAddr{IP: net.IPv6loopback, Port: 1},
	}
	for _, a := range good {
		if err := AssertLoopbackAddr(a); err != nil {
			t.Errorf("AssertLoopbackAddr(%v) = %v; want nil", a, err)
		}
	}
}

func TestNewServerRefusesNonLoopbackAtStartup(t *testing.T) {
	_, err := NewServer(ServerConfig{Host: "0.0.0.0", Handler: http.NotFoundHandler()})
	if err == nil {
		t.Fatal("NewServer on 0.0.0.0 succeeded")
	}
}

func TestServerShutsDownWithinDeadline(t *testing.T) {
	block := make(chan struct{})
	started := make(chan struct{})
	h := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		select { // a stalled handler that only ends when the server context ends
		case <-block:
		case <-r.Context().Done():
		}
	})
	srv, err := NewServer(ServerConfig{Host: "127.0.0.1", Handler: h})
	if err != nil {
		t.Fatal(err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve() }()

	reqDone := make(chan struct{})
	go func() {
		defer close(reqDone)
		resp, err := http.Get("http://" + srv.Addr().String() + "/")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	begin := time.Now()
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown: %v", err)
	}
	if d := time.Since(begin); d > 3*time.Second {
		t.Fatalf("Shutdown took %v; want within the deadline", d)
	}
	close(block)
	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Serve returned %v; want ErrServerClosed", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after Shutdown")
	}
	<-reqDone
}
