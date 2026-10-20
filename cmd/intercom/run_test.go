package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/app"
)

func execute(t *testing.T, run runFunc, args ...string) error {
	t.Helper()
	cmd := newRootCmdWith(run)
	cmd.SetArgs(args)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	return cmd.Execute()
}

// TestRootCmd_FlagsBecomeAppOptions pins the process flag surface (070.001-T):
// workspace, config path and port reach the assembly, and the production seams
// (agent factory, signals, logger) are populated.
func TestRootCmd_FlagsBecomeAppOptions(t *testing.T) {
	var got app.Options
	err := execute(t, func(_ context.Context, o app.Options) error { got = o; return nil },
		"--config", "x.toml", "--workspace", "ws", "--port", "4123")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got.ConfigPath != "x.toml" || got.Workspace != "ws" || got.Port != 4123 || !got.PortSet {
		t.Errorf("options = %+v; want config x.toml, workspace ws, port 4123 (set)", got)
	}
	if got.NewAgent == nil || got.Signals == nil || got.Logger == nil {
		t.Errorf("production seams missing: agent=%v signals=%v logger=%v", got.NewAgent != nil, got.Signals != nil, got.Logger != nil)
	}
	if got.TUI.Disabled {
		t.Error("the local TUI is disabled in the production invocation")
	}
}

func TestRootCmd_PortUnsetKeepsConfigPort(t *testing.T) {
	var got app.Options
	if err := execute(t, func(_ context.Context, o app.Options) error { got = o; return nil }); err != nil {
		t.Fatal(err)
	}
	if got.PortSet {
		t.Error("PortSet true without --port: the config's http_port must apply")
	}
	if got.ConfigPath != "config.toml" {
		t.Errorf("default config path = %q", got.ConfigPath)
	}
}

func TestRootCmd_RunErrorExitsNonZero(t *testing.T) {
	boom := errors.New("boom")
	if err := execute(t, func(context.Context, app.Options) error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("Execute error = %v; want the run error (process exits non-zero)", err)
	}
}

func TestRootCmd_PortOutOfRangeRejected(t *testing.T) {
	called := false
	err := execute(t, func(context.Context, app.Options) error { called = true; return nil }, "--port", "70000")
	if err == nil || called {
		t.Fatalf("out-of-range --port accepted (err=%v, run called=%v)", err, called)
	}
}

// TestRootCmd_InvalidConfigExitsNonZeroWithoutStarting drives the real
// assembly (not an injected run): a missing config file must fail startup.
func TestRootCmd_InvalidConfigExitsNonZeroWithoutStarting(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.toml")
	if err := execute(t, runApp, "--config", missing); err == nil {
		t.Fatal("startup with a missing config returned nil; want a non-zero exit")
	}
	bad := filepath.Join(t.TempDir(), "bad.toml")
	if err := os.WriteFile(bad, []byte("default_workspace_root = \n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := execute(t, runApp, "--config", bad); err == nil {
		t.Fatal("startup with an invalid config returned nil; want a non-zero exit")
	}
}
