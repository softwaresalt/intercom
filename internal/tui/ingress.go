package tui

import (
	"context"
	"errors"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Ingress is the only way the TUI acts on the session: it submits one
// operator action to the hub and receives the hub's verdict (design §4.1,
// ADR MVP-3). *hub.Client and *Feed satisfy it. The TUI never holds an agent
// handle and never resolves a permission itself; Update only builds a
// tea.Cmd that calls Do.
type Ingress interface {
	Do(ctx context.Context, act hub.OperatorAction) (hub.ActionResult, error)
}

// ErrNoIngress reports an action attempted with no ingress attached.
var ErrNoIngress = errors.New("tui: no hub ingress attached")

// ErrNotConnected reports an action attempted while the feed has no live hub
// connection (before Start, mid-reconnect, or after the hub went away).
var ErrNotConnected = errors.New("tui: not connected to the hub")

// Conn is the slice of *hub.Client the feed uses. Frames is the hub-owned
// buffered channel (never closed); Done closes when the client is dropped or
// the hub stops.
type Conn interface {
	Frames() <-chan hub.Frame
	Done() <-chan struct{}
	Do(ctx context.Context, act hub.OperatorAction) (hub.ActionResult, error)
	Rehydrate(ctx context.Context) error
	Close()
}

// Rehydrater is implemented by a [Conn] that can re-serve itself a fresh
// session.hydrate in place, staying registered with the hub (so a pending
// permission keeps its surface). *hub.Client satisfies it.
type Rehydrater interface {
	Rehydrate(ctx context.Context) error
}

// Dial opens a fresh TUI connection. Every dial starts with a session.hydrate
// frame (no resume cursor), which is what makes a dial a forced re-hydrate.
type Dial func(ctx context.Context) (Conn, error)

// HubDialer returns a Dial that attaches the in-process TUI client to h.
// buffer is the hub-side client channel capacity (0 selects the hub default).
func HubDialer(h *hub.Hub, buffer int) Dial {
	return func(ctx context.Context) (Conn, error) {
		c, err := h.Connect(ctx, hub.ConnectOptions{Kind: hub.ClientTUI, Buffer: buffer})
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}
