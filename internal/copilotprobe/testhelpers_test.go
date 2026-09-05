package copilotprobe

import (
	"context"
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// cleanupTimeout bounds every t.Cleanup teardown call in this package so a
// hung Stop()/Disconnect() (exactly the deadlock scenario B4/S3 probes for)
// can never hang the whole test binary. Discovering that a stage does not
// return within this bound is itself logged, never silently swallowed.
const cleanupTimeout = 15 * time.Second

// boundedStop calls client.Stop() with a bounded wait, logging (never
// failing) if it does not return in time. Errors are intentionally
// discarded beyond logging: cleanup is best-effort and must never fail an
// otherwise-passing test.
func boundedStop(t *testing.T, client *copilot.Client) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- client.Stop() }()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("cleanup: client.Stop() returned err=%v", err)
		}
	case <-time.After(cleanupTimeout):
		t.Logf("cleanup: client.Stop() did not return within %s; abandoning wait (leaked goroutine accepted for cleanup safety)", cleanupTimeout)
	}
}

// boundedDisconnect calls session.Disconnect() with a bounded wait, same
// rationale as boundedStop.
func boundedDisconnect(t *testing.T, session *copilot.Session) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- session.Disconnect() }()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("cleanup: session.Disconnect() returned err=%v", err)
		}
	case <-time.After(cleanupTimeout):
		t.Logf("cleanup: session.Disconnect() did not return within %s; abandoning wait (leaked goroutine accepted for cleanup safety)", cleanupTimeout)
	}
}

// newProbeClient starts a real Copilot SDK client rooted at a disposable
// temporary working directory, so this spike never touches the actual repo
// checkout. It skips the test cleanly (never fails it) when a runtime
// connection cannot be established -- the CLI/credentials-unavailable
// disposition every B2-B5 acceptance criterion requires.
//
// The returned client is automatically, boundedly stopped via t.Cleanup.
// Tests that manage their own Stop/ForceStop shutdown ladder (S3/SQ-a) MUST
// use newProbeClientManualLifecycle instead, so the SDK's documented
// "call stop exactly once" contract is never violated by a competing
// automatic cleanup.
func newProbeClient(t *testing.T, ctx context.Context) *copilot.Client {
	t.Helper()
	client := newProbeClientManualLifecycle(t, ctx)
	t.Cleanup(func() { boundedStop(t, client) })
	return client
}

// newProbeClientManualLifecycle starts a real Copilot SDK client identically
// to newProbeClient but registers NO automatic Stop cleanup: the caller owns
// the full stop lifecycle and MUST call client.Stop() or client.ForceStop()
// itself exactly once.
func newProbeClientManualLifecycle(t *testing.T, ctx context.Context) *copilot.Client {
	t.Helper()
	client := NewClient(&copilot.ClientOptions{WorkingDirectory: t.TempDir()})
	if err := client.Start(ctx); err != nil {
		t.Skipf("copilot SDK runtime unavailable, recording UNPROVEN: Start() failed: %v", err)
	}
	return client
}

// newProbeSession creates a session on client, skipping cleanly on failure.
// The returned session is automatically, boundedly disconnected via
// t.Cleanup.
func newProbeSession(t *testing.T, ctx context.Context, client *copilot.Client, cfg *copilot.SessionConfig) *copilot.Session {
	t.Helper()
	session, err := client.CreateSession(ctx, cfg)
	if err != nil {
		t.Skipf("copilot SDK session creation failed, recording UNPROVEN: %v", err)
	}
	t.Cleanup(func() { boundedDisconnect(t, session) })
	return session
}

// probeDeadline is the hard per-subtask deadline applied across B2-B5 so no
// probe can hang unbounded (plan requirement: "no test hangs unbounded").
const probeDeadline = 90 * time.Second
