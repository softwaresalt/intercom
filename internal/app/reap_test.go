package app

import (
	"context"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/transport"
)

// TestReapOnAPartialAppIsNilSafeAndRevokesCredentials characterizes reap's
// nil guards: with no remote, hub, accept loop or UI assembled it still
// succeeds, and it makes the credential clearing sticky so no later mint can
// publish into creds again.
func TestReapOnAPartialAppIsNilSafeAndRevokesCredentials(t *testing.T) {
	a := &App{
		log:   newTestLogger(&logBuf{}),
		done:  make(chan struct{}),
		shut:  shutdownState{req: make(chan struct{})},
		creds: []transport.Issued{{Fingerprint: "fp"}},
	}
	if err := a.reap(context.Background()); err != nil {
		t.Fatalf("reap on a partial App = %v; want nil", err)
	}
	if got := a.Credentials(); len(got) != 0 {
		t.Fatalf("Credentials() = %v after reap; want none", got)
	}
	a.mu.Lock()
	cleared := a.credsCleared
	a.mu.Unlock()
	if !cleared {
		t.Fatal("reap did not set the sticky credsCleared flag")
	}
}
