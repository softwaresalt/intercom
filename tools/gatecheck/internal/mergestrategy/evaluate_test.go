package mergestrategy

import (
	"bytes"
	"strings"
	"testing"
)

// TestEvaluate_TrailingDataNeverPasses is the M3-T9 named red test (plan
// §6 AC, and the revision-2 P1 Go Reviewer finding this closes):
// json.Decoder.Decode alone silently ignores trailing input after a
// complete JSON value, so a naive single-Decode port would turn an input
// Python's json.loads rejects as "Extra data" (SKIP) into a false PASS.
// This test fails (was committed failing, before the second-Decode/EOF
// check existed) if that regression is ever reintroduced.
func TestEvaluate_TrailingDataNeverPasses(t *testing.T) {
	cases := []string{
		`{"allow_squash_merge": false, "allow_rebase_merge": false} x`,
		`{"allow_squash_merge": false, "allow_rebase_merge": false}{"allow_squash_merge": true, "allow_rebase_merge": true}`,
	}
	for _, content := range cases {
		verdict, _, code, readErr := Evaluate("-", strings.NewReader(content))
		if readErr != "" {
			t.Fatalf("unexpected readErr: %s", readErr)
		}
		if verdict != Skip {
			t.Errorf("trailing data must never produce a verdict other than SKIP; got %s for input %q", verdict, content)
		}
		if code != 0 {
			t.Errorf("SKIP must exit 0; got %d for input %q", code, content)
		}
	}
}

// TestRun_NoPathArgument asserts the CLI entry point's own usage/exit-2
// path when no path argument is supplied.
func TestRun_NoPathArgument(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(nil, "/unused", strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on usage error, got %q", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Errorf("expected a usage message on stderr")
	}
}

// TestRun_StdinPath exercises the full Run() dispatch (not just Evaluate)
// against the "-" stdin transport for a simple PASS case, confirming the
// printed line is "<VERDICT> <reason>" exactly as evaluate_json's emit()
// would have produced.
func TestRun_StdinPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	content := `{"allow_squash_merge": false, "allow_rebase_merge": false}`
	code := Run([]string{"-"}, "/unused", strings.NewReader(content), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%s", code, stderr.String())
	}
	want := "PASS allow_squash_merge and allow_rebase_merge are both false\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}
