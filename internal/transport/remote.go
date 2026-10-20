package transport

import (
	"context"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// RemoteOptions configures [NewRemote].
type RemoteOptions struct {
	// Hub is the hub clients attach to. Required.
	Hub *hub.Hub
	// Remote is the validated [remote] config section.
	Remote config.RemoteConfig
	// Host is the loopback bind host (default 127.0.0.1). A non-loopback host
	// is refused.
	Host string
	// Port is the TCP port to bind; 0 selects an ephemeral port.
	Port uint16
	// Now is the clock for credential lifetime and the action rate limiter
	// (default time.Now).
	Now func() time.Time
	// Rand is the credential entropy source (default crypto/rand).
	Rand io.Reader
	// Logger receives transport diagnostics (default discard).
	Logger *slog.Logger
}

// Remote is the composed remote surface: credential store, upgrade handler
// and loopback server, wired together so a caller cannot forget the pairings
// ([Server]'s OnBind to [Handler.SetBoundAddr], its Drain to [Handler.Drain]).
// Credentials are bound to the hub's generation and live for
// [config.RemoteConfig.TokenLifetimeSeconds].
type Remote struct {
	creds   *CredentialStore
	server  *Server
	handler *Handler
}

// NewRemote maps cfg onto the transport configuration and binds the loopback
// listener. Nothing is accepted until [Remote.Serve]. Credentials are not
// minted here; see [Remote.Mint].
func NewRemote(o RemoteOptions) (*Remote, error) {
	if o.Hub == nil {
		return nil, apperr.New(apperr.KindConfig, "transport: remote needs a hub")
	}
	creds, err := NewCredentialStore(CredentialConfig{
		Gen:      o.Hub.Gen(),
		Lifetime: LifetimeFromSeconds(o.Remote.TokenLifetimeSeconds),
		Now:      o.Now,
		Rand:     o.Rand,
	})
	if err != nil {
		return nil, err
	}
	handler, err := NewHandler(Config{
		Hub:            o.Hub,
		Creds:          creds,
		AllowedOrigins: o.Remote.AllowedOrigins,
		AllowedHosts:   o.Remote.AllowedHosts,
		Now:            o.Now,
		Logger:         o.Logger,
	})
	if err != nil {
		return nil, err
	}
	server, err := NewServer(ServerConfig{
		Host:           o.Host,
		Port:           o.Port,
		Handler:        handler,
		Drain:          handler.Drain,
		MaxConnections: o.Remote.MaxConnections,
		Logger:         o.Logger,
		OnBind:         handler.SetBoundAddr,
	})
	if err != nil {
		handler.cancel() // release the handler's context; nothing was admitted
		return nil, err
	}
	return &Remote{creds: creds, server: server, handler: handler}, nil
}

// Addr returns the bound loopback address.
func (r *Remote) Addr() net.Addr { return r.server.Addr() }

// Serve accepts connections until the listener is closed; it then returns
// [net/http.ErrServerClosed].
func (r *Remote) Serve() error { return r.server.Serve() }

// Mint issues a credential for role, revoking that role's previous one.
func (r *Remote) Mint(role hub.Role) (Issued, error) { return r.creds.Mint(role) }

// Shutdown closes the listener, cancels every session and waits (bounded by
// ctx) for their goroutines. Pass a budget of at least [MinShutdownBudget].
func (r *Remote) Shutdown(ctx context.Context) error { return r.server.Shutdown(ctx) }
