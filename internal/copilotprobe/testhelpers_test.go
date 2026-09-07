package copilotprobe

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

// liveSDKTestsEnvVar is the opt-in gate for tests in this package that make
// real Copilot SDK connections (which may consume ambient credentials and
// execute real shell commands via the probed permission handler). See
// 011.002-T (resolves A0A2D049).
const liveSDKTestsEnvVar = "INTERCOM_LIVE_SDK_TESTS"

// liveSDKTestsDenied is a package-level denial flag set once by TestMain.
// It defaults to true (denied) so that any test executed before TestMain
// hypothetically ran would still observe the fail-closed value.
var liveSDKTestsDenied = true

// liveSDKTestsAllowed implements default-deny value semantics for
// INTERCOM_LIVE_SDK_TESTS: enabled only when raw parses as boolean true via
// strconv.ParseBool (which accepts the exact string "1" among others).
// "0", "false", "off", the empty string, and any unparseable value all deny
// -- there is no fail-open path on a parse error.
func liveSDKTestsAllowed(raw string) bool {
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return v
}

// credentialEnvVarsToSanitize lists ambient credential environment variables
// that must never silently authorize a live Copilot SDK connection when the
// opt-in gate is denied.
var credentialEnvVarsToSanitize = []string{
	"GITHUB_TOKEN",
	"GITHUB_PERSONAL_ACCESS_TOKEN",
	"COPILOT_TOKEN",
	"GH_TOKEN",
}

// TestMain is the fail-closed backstop (011.002-T). When
// INTERCOM_LIVE_SDK_TESTS is not enabled per liveSDKTestsAllowed, ambient
// credential environment variables are sanitized from this test binary's
// own process environment and liveSDKTestsDenied is set so per-test helpers
// (newProbeClient / newProbeClientManualLifecycle) refuse to start a real
// SDK connection. This backstop deliberately still calls m.Run() in every
// case -- it must never skip or exit the whole package, which would mask
// regressions and block its own verification test from executing.
func TestMain(m *testing.M) {
	if liveSDKTestsAllowed(os.Getenv(liveSDKTestsEnvVar)) {
		liveSDKTestsDenied = false
	} else {
		liveSDKTestsDenied = true
		for _, key := range credentialEnvVarsToSanitize {
			os.Unsetenv(key)
		}
	}
	os.Exit(m.Run())
}

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
	if liveSDKTestsDenied {
		t.Skipf("live Copilot SDK tests disabled (default-deny): set %s=1 to opt in", liveSDKTestsEnvVar)
	}
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
