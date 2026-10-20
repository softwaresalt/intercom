package transport

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// Pre-authentication HTTP server limits (ADR MVP-2, MVP-D12). The exact
// numeric values are named here, as the ADR requires.
const (
	// HandshakeTimeout is the fixed read deadline for the HTTP request line
	// and headers of an upgrade. A client that has not completed its headers
	// within it is disconnected.
	HandshakeTimeout = 10 * time.Second
	// MaxHeaderBytes bounds the request header block (ADR MVP-2: request
	// headers are size-limited).
	MaxHeaderBytes = 16 << 10
	// idleHTTPTimeout bounds keep-alive connections for non-upgrade requests.
	idleHTTPTimeout = 30 * time.Second
	// MinShutdownBudget is the smallest [Server.Shutdown] context budget that
	// cannot end in a deadline error merely because a peer never answered the
	// close handshake: the pump aborts such a socket after DefaultCloseGrace,
	// and one second of margin covers scheduling and the hub drain.
	MinShutdownBudget = DefaultCloseGrace + time.Second
)

// resolveFunc resolves a host name to addresses (injectable for tests).
type resolveFunc func(ctx context.Context, host string) ([]netip.Addr, error)

func defaultResolver(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

// AssertLoopbackAddr reports an error unless a is a TCP address whose IP is a
// loopback address (127.0.0.0/8 or ::1). It is the bind assertion of design
// §6: the listener binds loopback only; exposure is exclusively via the
// tunnel. A nil or unspecified address (0.0.0.0, ::) fails.
func AssertLoopbackAddr(a net.Addr) error {
	ta, ok := a.(*net.TCPAddr)
	if !ok || ta == nil {
		return apperr.New(apperr.KindPolicy, "transport: listener address is not a TCP address")
	}
	ip, ok := netip.AddrFromSlice(ta.IP)
	if !ok || ip.Zone() != "" || !ip.Unmap().IsLoopback() {
		return apperr.New(apperr.KindPolicy, "transport: listener address is not loopback")
	}
	return nil
}

// loopbackTarget resolves host to the single loopback address to bind. A
// literal IP must itself be loopback; a name must resolve, and EVERY address
// it resolves to must be loopback (a name that also resolves elsewhere is
// refused rather than trusting the resolver's ordering).
func loopbackTarget(ctx context.Context, host string, resolve resolveFunc) (netip.Addr, error) {
	refuse := func() (netip.Addr, error) {
		return netip.Addr{}, apperr.New(apperr.KindPolicy,
			"transport: bind host must be a loopback address (127.0.0.1, ::1 or a name resolving only to loopback)")
	}
	if host == "" {
		return refuse()
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" || !ip.Unmap().IsLoopback() {
			return refuse()
		}
		return ip.Unmap(), nil
	}
	if resolve == nil {
		resolve = defaultResolver
	}
	addrs, err := resolve(ctx, host)
	if err != nil || len(addrs) == 0 {
		return refuse()
	}
	for _, a := range addrs {
		if a.Zone() != "" || !a.Unmap().IsLoopback() {
			return refuse()
		}
	}
	return addrs[0].Unmap(), nil
}

// listen binds a TCP listener on a loopback address and re-asserts, on the
// bound socket, that it really is loopback (defence against a resolver or
// platform surprise). It never returns a non-loopback listener.
func listen(ctx context.Context, host string, port uint16, resolve resolveFunc) (net.Listener, error) {
	ip, err := loopbackTarget(ctx, host, resolve)
	if err != nil {
		return nil, err
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", net.JoinHostPort(ip.String(), strconv.Itoa(int(port))))
	if err != nil {
		return nil, apperr.Wrapf(apperr.KindIO, err, "transport: listen on loopback port %d failed", port)
	}
	if err := AssertLoopbackAddr(ln.Addr()); err != nil {
		_ = ln.Close()
		return nil, err
	}
	return ln, nil
}

// ServerConfig configures a [Server].
type ServerConfig struct {
	// Host is the loopback bind host. Default "127.0.0.1". A non-loopback
	// host is refused at startup.
	Host string
	// Port is the TCP port ([config.Config.HTTPPort]); 0 selects an ephemeral
	// port.
	Port uint16
	// Handler serves every request. Required.
	Handler http.Handler
	// Drain, when set, is called during Shutdown after the base context is
	// canceled; it waits (bounded by ctx) for hijacked connection goroutines
	// that http.Server does not track.
	Drain func(ctx context.Context) error
	// HandshakeTimeout overrides [HandshakeTimeout] (tests).
	HandshakeTimeout time.Duration
	// MaxConnections is [remote].max_connections: the cap on concurrent
	// sockets, counting unauthenticated ones. Default [DefaultMaxConnections].
	MaxConnections int
	// Logger receives server diagnostics (including net/http errors) through
	// slog. Default: discard.
	Logger *slog.Logger
	// OnBind, when set, is called with the bound loopback address before
	// Serve (used to tell the handler its port).
	OnBind func(net.Addr)

	resolver resolveFunc // test seam
}

// Server is the loopback-only HTTP server that hosts the WebSocket upgrade
// handler. Its lifecycle: [NewServer] binds (and asserts) the listener,
// [Server.Serve] runs until [Server.Shutdown].
type Server struct {
	ln     *limitListener
	srv    *http.Server
	cancel context.CancelFunc
	drain  func(ctx context.Context) error

	lnClosed atomic.Bool // the listener was closed by CloseListener/Shutdown
}

// NewServer binds the loopback listener and returns a Server. A non-loopback
// bind address is refused with a KindPolicy error and nothing is bound.
func NewServer(cfg ServerConfig) (*Server, error) {
	if cfg.Handler == nil {
		return nil, apperr.New(apperr.KindConfig, "transport: server handler is required")
	}
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	raw, err := listen(context.Background(), host, cfg.Port, cfg.resolver)
	if err != nil {
		return nil, err
	}
	ln := newLimitListener(raw, cfg.MaxConnections)
	logger := cfg.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	hs := cfg.HandshakeTimeout
	if hs <= 0 {
		hs = HandshakeTimeout
	}
	base, cancel := context.WithCancel(context.Background())
	if cfg.OnBind != nil {
		cfg.OnBind(ln.Addr())
	}
	hsrv := &http.Server{
		Handler:           cfg.Handler,
		ReadHeaderTimeout: hs,
		ReadTimeout:       hs,
		IdleTimeout:       idleHTTPTimeout,
		MaxHeaderBytes:    MaxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		BaseContext:       func(net.Listener) context.Context { return base },
	}
	// Only /ws is served and every pre-upgrade rejection must release its
	// connection-cap slot at once, so HTTP keep-alive is never useful.
	hsrv.SetKeepAlivesEnabled(false)
	return &Server{ln: ln, cancel: cancel, drain: cfg.Drain, srv: hsrv}, nil
}

// ActiveConnections reports the sockets currently holding a connection slot.
func (s *Server) ActiveConnections() int { return int(s.ln.active.Load()) }

// Addr returns the bound (loopback) listener address.
func (s *Server) Addr() net.Addr { return s.ln.Addr() }

// Serve accepts connections until [Server.CloseListener] or
// [Server.Shutdown]; it then returns [http.ErrServerClosed].
func (s *Server) Serve() error {
	err := s.srv.Serve(s.ln)
	if err != nil && s.lnClosed.Load() && errors.Is(err, net.ErrClosed) {
		return http.ErrServerClosed
	}
	return err
}

// closeListener closes the bound listener once, ignoring net.ErrClosed.
func (s *Server) closeListener() error {
	s.lnClosed.Store(true)
	if err := s.ln.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return apperr.Wrap(apperr.KindIO, err)
	}
	return nil
}

// CloseListener ONLY stops accepting new connections: the bound listener is
// closed, but no session is touched, no context is canceled and nothing is
// drained, so established connections keep running. It is the primitive for
// step 0 of the hub.Gate shutdown sequence (stop admitting before the
// server.shutdown broadcast); [Server.Shutdown] remains the final step that
// tears sessions down. It is idempotent and ctx is honoured only before the
// (non-blocking) close.
func (s *Server) CloseListener(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return apperr.Wrap(apperr.KindIO, err)
	}
	return s.closeListener()
}

// Shutdown stops accepting (closing the listener explicitly, so it also works
// when Serve was never called), cancels every in-flight request/connection
// context, and waits for them until ctx ends, then force-closes whatever is
// left. A peer that never completes the close handshake holds its session for
// the close grace, so callers should pass a budget of at least
// [MinShutdownBudget]; below that a deadline error is possible. It returns
// within ctx's deadline plus a small close margin.
func (s *Server) Shutdown(ctx context.Context) error {
	s.cancel()
	lerr := s.closeListener()
	err := s.srv.Shutdown(ctx)
	if errors.Is(err, net.ErrClosed) {
		err = nil // the listener was already closed by closeListener
	}
	if err == nil {
		err = lerr
	}
	if err != nil {
		_ = s.srv.Close()
	}
	if s.drain != nil {
		if derr := s.drain(ctx); derr != nil && err == nil {
			err = derr
		}
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
