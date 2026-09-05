package copilotprobe

import (
	"context"
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// newProbeClient starts a real Copilot SDK client rooted at a disposable
// temporary working directory, so this spike never touches the actual repo
// checkout. It skips the test cleanly (never fails it) when a runtime
// connection cannot be established -- the CLI/credentials-unavailable
// disposition every B2-B5 acceptance criterion requires.
func newProbeClient(t *testing.T, ctx context.Context) *copilot.Client {
	t.Helper()
	client := NewClient(&copilot.ClientOptions{WorkingDirectory: t.TempDir()})
	if err := client.Start(ctx); err != nil {
		t.Skipf("copilot SDK runtime unavailable, recording UNPROVEN: Start() failed: %v", err)
	}
	t.Cleanup(func() { _ = client.Stop() })
	return client
}

// newProbeSession creates a session on client, skipping cleanly on failure.
func newProbeSession(t *testing.T, ctx context.Context, client *copilot.Client, cfg *copilot.SessionConfig) *copilot.Session {
	t.Helper()
	session, err := client.CreateSession(ctx, cfg)
	if err != nil {
		t.Skipf("copilot SDK session creation failed, recording UNPROVEN: %v", err)
	}
	t.Cleanup(func() { _ = session.Disconnect() })
	return session
}

// probeDeadline is the hard per-subtask deadline applied across B2-B5 so no
// probe can hang unbounded (plan requirement: "no test hangs unbounded").
const probeDeadline = 90 * time.Second
