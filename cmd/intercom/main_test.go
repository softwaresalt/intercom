package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestRootCmd_HelpExitsZero asserts --help succeeds (maps to process exit 0).
func TestRootCmd_HelpExitsZero(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("--help returned error, want nil (exit 0): %v", err)
	}
}

// TestRootCmd_UnknownFlagExitsNonZero asserts an unrecognized flag fails
// (maps to process exit non-zero).
func TestRootCmd_UnknownFlagExitsNonZero(t *testing.T) {
	cmd := newRootCmd()
	cmd.SetArgs([]string{"--does-not-exist"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	if err := cmd.Execute(); err == nil {
		t.Fatal("unknown flag returned nil error, want non-nil (exit non-zero)")
	}
}

// TestRootCmd_LogLevelDebugEmitsJSON asserts --log-level=debug produces a
// parseable JSON log line on stderr.
func TestRootCmd_LogLevelDebugEmitsJSON(t *testing.T) {
	var buf bytes.Buffer
	orig := stderr
	stderr = &buf
	defer func() { stderr = orig }()

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--log-level=debug"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	// RunE deliberately returns the not-implemented sentinel; the log line
	// is still emitted before that return.
	_ = cmd.Execute()

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected a JSON log line on stderr, got none")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(line), &payload); err != nil {
		t.Fatalf("stderr output did not parse as JSON: %v\noutput: %s", err, line)
	}
	if _, ok := payload["msg"]; !ok {
		t.Fatalf("expected JSON log payload to include a msg field: %v", payload)
	}
}

// TestRootCmd_LogLevelErrorSuppressesInfo proves --log-level actually
// filters output (not just that some line appears at debug): the Info-level
// startup line must not appear when the level is set to error.
func TestRootCmd_LogLevelErrorSuppressesInfo(t *testing.T) {
	var buf bytes.Buffer
	orig := stderr
	stderr = &buf
	defer func() { stderr = orig }()

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--log-level=error"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	if strings.TrimSpace(buf.String()) != "" {
		t.Fatalf("expected no log output at level=error for an Info-level message, got: %s", buf.String())
	}
}
