package copilot

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// liveSDKTestsEnvVar is the workspace-wide default-deny opt-in for tests that
// make real Copilot SDK connections (design §7.6; convention inherited from
// the retired disposable C2 spike package).
const liveSDKTestsEnvVar = "INTERCOM_LIVE_SDK_TESTS"

// liveSDKTestsAllowed reports whether raw opts in: strconv.ParseBool, and
// deny on any parse error or empty value.
func liveSDKTestsAllowed(raw string) bool {
	if raw == "" {
		return false
	}
	v, err := strconv.ParseBool(raw)
	return err == nil && v
}

func TestLiveSDKTestsAllowedDefaultDenySemantics(t *testing.T) {
	cases := []struct {
		raw  string
		want bool
	}{
		{"", false}, {"1", true}, {"0", false}, {"false", false},
		{"off", false}, {"true", true}, {"garbage", false},
	}
	for _, tc := range cases {
		if got := liveSDKTestsAllowed(tc.raw); got != tc.want {
			t.Errorf("liveSDKTestsAllowed(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

// TestLiveStartAndStop exercises the real SDK and a real Copilot runtime. It
// is skipped unless INTERCOM_LIVE_SDK_TESTS opts in.
func TestLiveStartAndStop(t *testing.T) {
	if !liveSDKTestsAllowed(os.Getenv(liveSDKTestsEnvVar)) {
		t.Skipf("%s not enabled; skipping live SDK test", liveSDKTestsEnvVar)
	}
	hubDone := make(chan struct{})
	defer close(hubDone)
	ad, err := New(Config{
		WorkingDirectory: t.TempDir(),
		StartupTimeout:   60 * time.Second,
		StopTimeout:      10 * time.Second,
		Ingress:          make(chan agent.Event, 64),
		Asks:             make(chan agent.PermissionAsk, 1),
		HubDone:          hubDone,
		Logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	h, err := ad.Start(context.Background())
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if h.ID == "" {
		t.Error("empty session id")
	}
	if err := ad.Stop(context.Background()); err != nil {
		t.Errorf("Stop: %v", err)
	}
}
