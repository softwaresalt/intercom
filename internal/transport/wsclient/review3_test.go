package wsclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// F-4: no credential (or any caller-chosen data) can travel in the URL.
func TestDialRefusesUserinfoAndQuery(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()
	host := strings.TrimPrefix(srv.URL, "http://")

	for _, u := range []string{
		"ws://user:SECRETVALUE@" + host + "/ws",
		"ws://SECRETVALUE@" + host + "/ws",
		"ws://" + host + "/ws?token=SECRETVALUE",
		"ws://" + host + "/ws?x=1",
		"ws://" + host + "/ws?gen=g&seq=1", // the client adds the cursor itself
		"ws://" + host + "/ws?",
		"ws://" + host + "/ws#frag", // fragments are never sent, but could hide data
		"ws://" + host + "/ws#",
		"ws:opaque/ws", // opaque URL: no host, no path
		"wss:example.com/ws",
		"ws:///ws", // missing host
		"wss://user:SECRETVALUE@example.com/ws",
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := Dial(ctx, Options{URL: u, Credential: "SECRETVALUE"})
		cancel()
		if !errors.Is(err, ErrInvalidURL) {
			t.Fatalf("Dial(%q) = %v; want ErrInvalidURL", u, err)
		}
		if strings.Contains(err.Error(), "SECRETVALUE") {
			t.Fatalf("error leaks the credential: %v", err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Fatalf("server saw %d requests for refused URLs", n)
	}
}

// The only query the client sends is the resume cursor it adds itself, and
// the credential is only in the header.
func TestDialSendsOnlyCursorQuery(t *testing.T) {
	got := make(chan *http.Request, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r
		c, err := websocket.Accept(w, r, nil)
		if err == nil {
			_ = c.CloseNow()
		}
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c, err := Dial(ctx, Options{
		URL: "ws://" + strings.TrimPrefix(srv.URL, "http://") + "/ws", Credential: "SECRETVALUE",
		Resume: &hub.Position{Gen: "g1", Seq: 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Abort()
	r := <-got
	if r.URL.RawQuery != "gen=g1&seq=5" {
		t.Fatalf("query = %q; want only gen and seq", r.URL.RawQuery)
	}
	if strings.Contains(r.URL.String(), "SECRETVALUE") || r.URL.User != nil {
		t.Fatal("credential present in the URL")
	}
}

// F-13: an explicit read limit sized for hydrate snapshots.
func TestReadLimitCoversHydrateBoundAndIsEnforced(t *testing.T) {
	bound := hub.SnapshotOverhead + hub.DefaultMaxSnapshotBytes +
		hub.DefaultMaxPendingPermissions*hub.PermissionEntryHydrateBytes + 1024
	if ReadLimit < bound {
		t.Fatalf("ReadLimit %d is below the hub's default hydrate bound %d", ReadLimit, bound)
	}
	sizes := make(chan int, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = c.CloseNow() }()
		ctx := r.Context()
		for range 2 {
			n := <-sizes
			if err := c.Write(ctx, websocket.MessageText, []byte(`"`+strings.Repeat("x", n-2)+`"`)); err != nil {
				return
			}
		}
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := Dial(ctx, Options{URL: "ws://" + strings.TrimPrefix(srv.URL, "http://") + "/ws", Credential: "x"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Abort()

	sizes <- bound // far above the library's 32 KiB default
	if _, err := c.Next(ctx); err != nil && !errors.Is(err, ErrUnrecognizedFrame) {
		t.Fatalf("hydrate-sized frame rejected: %v", err)
	}
	sizes <- ReadLimit + 1
	if _, err := c.Next(ctx); err == nil || errors.Is(err, ErrUnrecognizedFrame) {
		t.Fatalf("frame over ReadLimit not refused: %v", err)
	}
}
