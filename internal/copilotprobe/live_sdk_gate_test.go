package copilotprobe

import (
	"os"
	"testing"
)

// TestLiveSDKTestsAllowedDefaultDenySemantics directly unit-tests the gate
// helper's decision function (011.002-T). This test does NOT call t.Skip --
// a test that has already skipped itself cannot assert anything about its
// own gate decision, so the pure decision function is exercised directly
// with a table of raw values instead of relying solely on the ambient
// process environment.
func TestLiveSDKTestsAllowedDefaultDenySemantics(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"unset/empty", "", false},
		{"exact 1", "1", true},
		{"exact 0", "0", false},
		{"false", "false", false},
		{"off (unparseable)", "off", false},
		{"true (parses true)", "true", true},
		{"garbage", "enable-please", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := liveSDKTestsAllowed(tc.raw)
			if got != tc.want {
				t.Errorf("liveSDKTestsAllowed(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}

// TestLiveSDKGateDeniedByDefaultInThisRun asserts the actual package-level
// denial flag TestMain computed for THIS test run. In the default CI/dev
// posture (INTERCOM_LIVE_SDK_TESTS unset), this must observe
// liveSDKTestsDenied == true, proving the fail-closed backstop actually ran
// and set the flag rather than merely defining an unused helper.
func TestLiveSDKGateDeniedByDefaultInThisRun(t *testing.T) {
	if liveSDKTestsAllowed(os.Getenv(liveSDKTestsEnvVar)) {
		t.Skipf("%s is explicitly enabled for this run; default-deny assertion not applicable", liveSDKTestsEnvVar)
	}
	if !liveSDKTestsDenied {
		t.Fatalf("liveSDKTestsDenied = false with %s unset/disabled; TestMain backstop did not apply default-deny", liveSDKTestsEnvVar)
	}
}
