package transport

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

const (
	// WSPath is the only path the transport serves.
	WSPath = "/ws"
	// SubprotocolV1 is the one WebSocket subprotocol the handler selects. It
	// is non-secret: the credential-carrying "bearer.<token>" entry a browser
	// offers is never selected, so it is never echoed in the response.
	SubprotocolV1 = "intercom.v1"
	// bearerSubprotocolPrefix prefixes the credential-carrying subprotocol.
	bearerSubprotocolPrefix = "bearer."
	// maxCredentialLen bounds the presented credential before digesting.
	maxCredentialLen = 128
)

// Config configures a [Handler].
type Config struct {
	// Hub is the event hub clients attach to. Required.
	Hub *hub.Hub
	// Creds is the credential store. Required.
	Creds *CredentialStore
	// AllowedOrigins is [remote].allowed_origins: an exact-match allowlist.
	// Empty (the default) rejects every request that carries an Origin.
	AllowedOrigins []string
	// AllowedHosts is [remote].allowed_hosts, in addition to the loopback
	// names for the bound port.
	AllowedHosts []string
	// Now is the clock for the per-connection action token bucket ONLY;
	// default time.Now. Credential validity and the pump's expiry timer come
	// from the credential store's own clock ([CredentialStore.Remaining]), so
	// the pump and Validate can never disagree about expiry.
	Now func() time.Time
	// PingInterval, WriteTimeout and CloseGrace override the Default* timing
	// constants (tests).
	PingInterval time.Duration
	WriteTimeout time.Duration
	CloseGrace   time.Duration
	// ClientBuffer is the hub send-buffer capacity per remote client
	// (default hub.DefaultRemoteBuffer).
	ClientBuffer int
	// MaxMessageBytes caps one inbound message (default
	// [DefaultMaxMessageBytes]); a larger one closes the connection.
	MaxMessageBytes int
	// IdleTimeout closes a connection with no inbound activity (messages or
	// pongs) for this long (default [DefaultIdleTimeout]). A pong answering a
	// server ping IS activity, so a peer that keeps answering pings never
	// idles out; the credential lifetime is what bounds such a session.
	IdleTimeout time.Duration
	// ActionTimeout bounds one hub action round trip (default
	// [DefaultActionTimeout], reduced to fit a small WriteTimeout). The read
	// pump is blocked for the duration of an action and cannot answer pings
	// meanwhile, so NewHandler requires that ActionTimeout leave a pong margin
	// under WriteTimeout (see [actionFitsWriteTimeout]) and refuses a
	// configuration that does not.
	ActionTimeout time.Duration
	// ActionRate (actions/second) and ActionBurst configure the
	// per-connection inbound token bucket; MaxRateViolations is how many
	// consecutive rate-limited actions close the connection. Defaults:
	// [DefaultActionRate], [DefaultActionBurst], [DefaultMaxRateViolations].
	ActionRate        float64
	ActionBurst       int
	MaxRateViolations int
	// Logger receives transport diagnostics. The transport never logs
	// headers, URLs or request dumps. Default: discard.
	Logger *slog.Logger
}

// Handler is the WebSocket upgrade handler: it authenticates BEFORE the
// upgrade (design §6, ADR MVP-2) and then serves the connection.
type Handler struct {
	cfg     Config
	log     *slog.Logger
	origins map[string]struct{}
	hosts   map[string]struct{}
	port    atomic.Uint32 // bound port, for the loopback Host names

	ctx    context.Context //nolint:containedctx // handler-lifetime context, canceled by Drain
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// admitMu is the admission fence: wg.Add happens only under it while
	// draining is false, and Drain sets draining under it before waiting, so
	// an Add can never race the Wait.
	admitMu  sync.Mutex
	draining bool

	pingEvery     time.Duration
	writeTimeout  time.Duration
	closeGrace    time.Duration
	idleTimeout   time.Duration
	actionTimeout time.Duration
	maxMessage    int
	actionRate    float64
	actionBurst   int
	maxViolations int
	sessions      atomic.Int64
	pings         atomic.Int64
}

// NewHandler validates cfg and returns a Handler.
func NewHandler(cfg Config) (*Handler, error) {
	if cfg.Hub == nil || cfg.Creds == nil {
		return nil, apperr.New(apperr.KindConfig, "transport: handler needs a hub and a credential store")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	h := &Handler{
		cfg:     cfg,
		log:     cfg.Logger,
		origins: make(map[string]struct{}, len(cfg.AllowedOrigins)),
		hosts:   make(map[string]struct{}, len(cfg.AllowedHosts)),
	}
	for _, o := range cfg.AllowedOrigins {
		k, ok := originKey(o)
		if !ok {
			return nil, apperr.New(apperr.KindConfig, "transport: an allowed origin is not a canonical http(s) origin")
		}
		h.origins[k] = struct{}{}
	}
	for _, o := range cfg.AllowedHosts {
		k, ok := hostKey(o)
		if !ok {
			return nil, apperr.New(apperr.KindConfig, "transport: an allowed host is not a canonical host[:port]")
		}
		h.hosts[k] = struct{}{}
	}
	h.pingEvery = durOr(cfg.PingInterval, DefaultPingInterval)
	h.writeTimeout = durOr(cfg.WriteTimeout, DefaultWriteTimeout)
	h.closeGrace = durOr(cfg.CloseGrace, DefaultCloseGrace)
	h.idleTimeout = durOr(cfg.IdleTimeout, DefaultIdleTimeout)
	if cfg.ActionTimeout > 0 {
		h.actionTimeout = cfg.ActionTimeout
		if !actionFitsWriteTimeout(h.actionTimeout, h.writeTimeout) {
			return nil, apperr.New(apperr.KindConfig, "transport: action timeout leaves no margin under the write timeout; ping replies would starve")
		}
	} else {
		h.actionTimeout = DefaultActionTimeout
		if !actionFitsWriteTimeout(h.actionTimeout, h.writeTimeout) {
			h.actionTimeout = h.writeTimeout / 2
		}
	}
	h.actionRate, h.actionBurst, h.maxViolations = cfg.ActionRate, cfg.ActionBurst, cfg.MaxRateViolations
	if h.actionRate <= 0 {
		h.actionRate = DefaultActionRate
	}
	if h.actionBurst <= 0 {
		h.actionBurst = DefaultActionBurst
	}
	if h.maxViolations <= 0 {
		h.maxViolations = DefaultMaxRateViolations
	}
	h.maxMessage = cfg.MaxMessageBytes
	if h.maxMessage <= 0 {
		h.maxMessage = DefaultMaxMessageBytes
	}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	return h, nil
}

// SetBoundAddr records the listener's port so the loopback Host names
// (127.0.0.1, [::1], localhost) are accepted for that port only. Until it is
// called only allowed_hosts match.
func (h *Handler) SetBoundAddr(a net.Addr) {
	if ta, ok := a.(*net.TCPAddr); ok && ta != nil {
		h.port.Store(uint32(ta.Port)) //nolint:gosec // TCP ports fit in 16 bits
	}
}

// actionFitsWriteTimeout reports whether an action round trip of the given
// duration leaves a pong margin (one fifth of the write timeout) under the
// ping write timeout. The read pump is blocked for the whole action and
// cannot answer a ping meanwhile.
func actionFitsWriteTimeout(action, write time.Duration) bool {
	return action > 0 && action+write/5 < write
}

// admit registers one connection goroutine unless the handler is draining.
func (h *Handler) admit() bool {
	h.admitMu.Lock()
	defer h.admitMu.Unlock()
	if h.draining {
		return false
	}
	h.wg.Add(1)
	return true
}

// Drain refuses every new upgrade, cancels every connection and waits (until
// ctx ends) for their goroutines, so Shutdown leaves none behind. A peer that
// never completes the close handshake delays its session by CloseGrace, so a
// ctx budget below [MinShutdownBudget] can end in a deadline error.
func (h *Handler) Drain(ctx context.Context) error {
	h.admitMu.Lock()
	h.draining = true
	h.admitMu.Unlock()
	h.cancel()
	done := make(chan struct{})
	go func() { h.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return apperr.Wrap(apperr.KindIO, ctx.Err())
	}
}

// Pre-upgrade rejection reasons. They are fixed strings for logs; nothing
// from the request is ever logged.
const (
	rejectHost        = "host_not_allowed"
	rejectOrigin      = "origin_not_allowed"
	rejectCredential  = "credential_rejected"
	rejectHubRegister = "hub_refused"
	rejectCursor      = "bad_cursor"
	rejectNotUpgrade  = "not_an_upgrade"
	rejectDraining    = "draining"
)

// reject answers a pre-upgrade refusal. Connection: close makes the rejected
// socket release its connection-cap slot immediately instead of idling as a
// keep-alive connection.
func (h *Handler) reject(w http.ResponseWriter, status int, reason string) {
	h.log.Warn("transport: upgrade rejected", "status", status, "reason", reason)
	w.Header().Set("Connection", "close")
	if status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", "Bearer")
	}
	w.WriteHeader(status)
}

// isUpgradeRequest reports whether r is shaped like a WebSocket opening
// handshake: GET, "Upgrade: websocket", a Connection header listing the
// upgrade token, and Sec-WebSocket-Version and -Key present. It exists so a
// plain request from a credential holder never reaches Hub.Connect.
func isUpgradeRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if !headerHasToken(r.Header, "Upgrade", "websocket") || !headerHasToken(r.Header, "Connection", "upgrade") {
		return false
	}
	return r.Header.Get("Sec-WebSocket-Version") != "" && r.Header.Get("Sec-WebSocket-Key") != ""
}

// headerHasToken reports whether any comma-separated value of header name
// equals token, case-insensitively.
func headerHasToken(h http.Header, name, token string) bool {
	for _, v := range h.Values(name) {
		for _, p := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(p), token) {
				return true
			}
		}
	}
	return false
}

// acceptOptions are the library options. InsecureSkipVerify stays false and
// OriginPatterns stays empty: Origin is authorised by this package's own
// exact-match pre-check, and the library's check can only ever be stricter.
func acceptOptions() websocket.AcceptOptions {
	return websocket.AcceptOptions{
		Subprotocols:       []string{SubprotocolV1},
		InsecureSkipVerify: false,
		CompressionMode:    websocket.CompressionDisabled,
	}
}

func (h *Handler) hostAllowed(host string) bool {
	k, ok := hostKey(host)
	if !ok {
		return false
	}
	if _, ok := h.hosts[k]; ok {
		return true
	}
	p := h.port.Load()
	if p == 0 {
		return false
	}
	suffix := ":" + strconv.FormatUint(uint64(p), 10)
	switch k {
	case "127.0.0.1" + suffix, "[::1]" + suffix, "localhost" + suffix:
		return true
	}
	return false
}

// ServeHTTP implements the upgrade. Check order (ADR MVP-2, cheapest first,
// fail closed): Host (403), Origin (403), credential (401). The connection
// cap (503) is added ahead of these by the pre-auth limiter.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != WSPath {
		w.Header().Set("Connection", "close")
		http.NotFound(w, r)
		return
	}
	if !h.hostAllowed(r.Host) {
		h.reject(w, http.StatusForbidden, rejectHost)
		return
	}
	if origins := r.Header.Values("Origin"); len(origins) > 0 {
		if !h.originAllowed(origins) {
			h.reject(w, http.StatusForbidden, rejectOrigin)
			return
		}
	}
	tok, ok := extractCredential(r)
	if !ok {
		h.reject(w, http.StatusUnauthorized, rejectCredential)
		return
	}
	grant, err := h.cfg.Creds.Verify(tok, h.cfg.Hub.Gen())
	if err != nil {
		h.reject(w, http.StatusUnauthorized, rejectCredential)
		return
	}
	// After authentication (so unauthenticated probes learn nothing) and
	// before any hub registration: a plain GET by a credential holder must
	// not churn hub client registration.
	if !isUpgradeRequest(r) {
		h.reject(w, http.StatusBadRequest, rejectNotUpgrade)
		return
	}
	h.upgrade(w, r, grant)
}

func (h *Handler) originAllowed(origins []string) bool {
	if len(origins) != 1 {
		return false
	}
	k, ok := originKey(origins[0])
	if !ok {
		return false
	}
	_, ok = h.origins[k]
	return ok
}

// extractCredential returns the presented credential from exactly one of the
// two permitted carriages: the "Authorization: Bearer <credential>" header (native clients)
// or the "bearer.<token>" WebSocket subprotocol offered next to "intercom.v1"
// (browsers). Both together, a duplicate, a missing intercom.v1, a malformed
// value, or a credential-shaped URL query parameter all fail: the caller
// answers one generic 401 for every case. The URL is never a source.
func extractCredential(r *http.Request) (string, bool) {
	if queryCarriesCredential(r.URL.RawQuery) {
		return "", false
	}
	auth := r.Header.Values("Authorization")
	var offered []string
	for _, v := range r.Header.Values("Sec-WebSocket-Protocol") {
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				offered = append(offered, p)
			}
		}
	}
	var bearers []string
	hasV1 := false
	for _, p := range offered {
		switch {
		case p == SubprotocolV1:
			hasV1 = true
		case strings.HasPrefix(p, bearerSubprotocolPrefix):
			bearers = append(bearers, p[len(bearerSubprotocolPrefix):])
		}
	}
	switch {
	case len(auth) > 0 && len(bearers) > 0:
		return "", false // ambiguous
	case len(auth) == 1:
		const scheme = "Bearer "
		v := auth[0]
		if len(v) <= len(scheme) || !strings.EqualFold(v[:len(scheme)], scheme) {
			return "", false
		}
		return plausibleCredential(v[len(scheme):])
	case len(auth) > 1:
		return "", false
	case len(bearers) == 1 && hasV1:
		return plausibleCredential(bearers[0])
	default:
		return "", false
	}
}

// plausibleCredential bounds and charset-checks a presented credential
// before it is digested; it reveals nothing about any stored credential.
func plausibleCredential(s string) (string, bool) {
	if s == "" || len(s) > maxCredentialLen {
		return "", false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_'
		if !ok {
			return "", false
		}
	}
	return s, true
}

// queryCarriesCredential reports whether any URL query parameter NAME looks
// credential-shaped. A credential in a URL is leaked into proxy, tunnel and
// browser logs, so such a request is refused outright even if it also carries
// a valid header. Values are never inspected or logged.
func queryCarriesCredential(rawQuery string) bool {
	for _, pair := range strings.Split(rawQuery, "&") {
		name, _, _ := strings.Cut(pair, "=")
		var b strings.Builder
		for _, r := range strings.ToLower(name) {
			if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
				b.WriteRune(r)
			}
		}
		n := b.String()
		for _, frag := range [...]string{"token", "bearer", "auth", "credential", "secret", "key", "pass", "jwt", "cookie", "sig"} {
			if strings.Contains(n, frag) {
				return true
			}
		}
	}
	return false
}

// upgrade registers the client with the hub, then completes the WebSocket
// handshake and serves the connection until it ends.
func (h *Handler) upgrade(w http.ResponseWriter, r *http.Request, grant *Grant) {
	if !h.admit() { // draining: refuse before touching the hub
		h.reject(w, http.StatusServiceUnavailable, rejectDraining)
		return
	}
	defer h.wg.Done()

	resume, ok := parseCursor(r.URL.Query())
	if !ok {
		h.reject(w, http.StatusBadRequest, rejectCursor)
		return
	}
	client, err := h.cfg.Hub.Connect(r.Context(), hub.ConnectOptions{
		Kind: hub.ClientRemote, Role: grant.Role(), Buffer: h.cfg.ClientBuffer, Resume: resume,
	})
	if err != nil {
		h.reject(w, http.StatusServiceUnavailable, rejectHubRegister)
		return
	}
	defer client.Close()

	// The Origin (if any) was already authorised against the exact-match
	// allowlist; hand the library a request without it so its same-origin
	// default can never contradict that decision. It is never looser: the
	// library check is skipped only for an already-allowlisted origin.
	ar := r.Clone(r.Context())
	ar.Header.Del("Origin")
	opts := acceptOptions()
	conn, err := websocket.Accept(w, ar, &opts)
	if err != nil {
		return // Accept has already written the HTTP error
	}
	newSession(h, conn, client, grant).run()
}

func durOr(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}

// ActiveSessions reports connections currently being served.
func (h *Handler) ActiveSessions() int { return int(h.sessions.Load()) }

// PingsSent reports keep-alive pings written by the pumps (diagnostics).
func (h *Handler) PingsSent() int64 { return h.pings.Load() }

func (h *Handler) now() time.Time { return h.cfg.Now() }
