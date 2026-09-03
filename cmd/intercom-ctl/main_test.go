package main

import (
	"bytes"
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
