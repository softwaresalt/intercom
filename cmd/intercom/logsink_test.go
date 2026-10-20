package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/app"
)

func withTerminalStderr(t *testing.T, buf *bytes.Buffer, term bool) {
	t.Helper()
	origErr, origTerm := stderr, isTerminal
	stderr = buf
	isTerminal = func(io.Writer) bool { return term }
	t.Cleanup(func() { stderr, isTerminal = origErr, origTerm })
}

// TestLogSink pins the destination rule: the TUI owns the terminal, so logs
// never go to a terminal stderr while it is enabled.
func TestLogSink(t *testing.T) {
	var buf bytes.Buffer
	yes := func(io.Writer) bool { return true }
	no := func(io.Writer) bool { return false }
	if got := logSink(true, &buf, yes); got != io.Discard {
		t.Error("TUI enabled + terminal stderr: logs must be discarded")
	}
	if got := logSink(true, &buf, no); got != io.Writer(&buf) {
		t.Error("TUI enabled + redirected stderr: logs must go to stderr")
	}
	if got := logSink(false, &buf, yes); got != io.Writer(&buf) {
		t.Error("TUI disabled: logs must go to stderr")
	}
}

func TestIsTerminalWriterRejectsNonFiles(t *testing.T) {
	if isTerminalWriter(&bytes.Buffer{}) {
		t.Error("a buffer is not a terminal")
	}
	f, err := os.CreateTemp(t.TempDir(), "log")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if isTerminalWriter(f) {
		t.Error("a regular file is not a terminal")
	}
}

// TestProcessLoggerSilentOnTerminalStderr: the logger handed to the process
// writes nothing to a terminal stderr, so the alt-screen is never corrupted.
func TestProcessLoggerSilentOnTerminalStderr(t *testing.T) {
	var buf bytes.Buffer
	withTerminalStderr(t, &buf, true)
	err := execute(t, func(_ context.Context, o app.Options) error {
		o.Logger.Error("from the process")
		return nil
	}, "--log-level=debug")
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("logs reached a terminal stderr while the TUI is enabled:\n%s", buf.String())
	}
}

// TestProcessLoggerWritesJSONToRedirectedStderr: with stderr redirected
// (2> intercom.log) the logs are kept, honouring --log-level.
func TestProcessLoggerWritesJSONToRedirectedStderr(t *testing.T) {
	var buf bytes.Buffer
	withTerminalStderr(t, &buf, false)
	err := execute(t, func(_ context.Context, o app.Options) error {
		o.Logger.Info("kept")
		o.Logger.Debug("filtered")
		return nil
	}, "--log-level=info")
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, `"msg":"kept"`) || strings.Contains(out, "filtered") {
		t.Fatalf("redirected stderr log = %q; want the info line only, as JSON", out)
	}
}
