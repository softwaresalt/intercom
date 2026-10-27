// Package app is the per-workspace process assembly (design §3, §6): it loads
// and validates the config, resolves the workspace root through pathsafe,
// then builds and starts the agent adapter, hub, remote transport and local
// TUI in the normative order, and tears them down again.
//
// It exists so the whole process can be driven headlessly: every
// environment-dependent input (agent runtime, port, TUI terminal I/O, clock,
// logger, entropy, signals) is an [Options] field, so tests and the
// end-to-end scenario inject fakes and cmd/intercom injects the real ones.
package app

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// AgentDeps is everything an [AgentFactory] needs from the assembled process.
type AgentDeps struct {
	// WorkingDirectory is the workspace root the agent runs in (resolved
	// through pathsafe).
	WorkingDirectory string
	// CLIPath is [copilot].cli_path ("" resolves the CLI from PATH).
	CLIPath string
	// StartupTimeout is [copilot].startup_timeout_seconds.
	StartupTimeout time.Duration
	// Ingress, Asks and HubDone connect the agent to the hub.
	Ingress chan<- agent.Event
	Asks    chan<- agent.PermissionAsk
	HubDone <-chan struct{}
	// Logger is the process logger.
	Logger *slog.Logger
}

// AgentFactory builds the (not yet started) agent for a process.
type AgentFactory func(AgentDeps) (agent.Agent, error)

// TUIOptions configures the local terminal UI.
type TUIOptions struct {
	// Disabled runs the process with no local TUI (tests).
	Disabled bool
	// Input and Output replace the terminal; nil uses the process's
	// stdin/stdout.
	Input  io.Reader
	Output io.Writer // nil: stdout, which must then be a terminal
	// Width and Height, when both set, give the program an initial window
	// size (a non-terminal Output reports none).
	Width, Height int
	// IsTerminal reports whether a file is a terminal; nil uses the real
	// check. Tests inject it to exercise the fail-closed refusal.
	IsTerminal func(*os.File) bool
	// RetainPlaintext keeps the credential plaintext readable through
	// [App.Credentials] after the TUI has been given it. It is a test seam;
	// by default the plaintext is zeroed once delivered to the TUI.
	RetainPlaintext bool
}

// Options configures [Start].
type Options struct {
	// ConfigPath is the config file.
	ConfigPath string
	// Workspace is the workspace directory; "" selects default_workspace_root.
	Workspace string
	// Port overrides [config.Config.HTTPPort] when PortSet.
	Port    uint16
	PortSet bool
	// NewAgent builds the agent. Required.
	NewAgent AgentFactory
	// Logger receives structured logs. Default: discard.
	Logger *slog.Logger
	// Now is the clock for credentials and the hub. Default time.Now.
	Now func() time.Time
	// CredentialRand overrides credential entropy (tests only).
	CredentialRand io.Reader
	// Signals delivers interrupt/terminate notifications. Nil: none.
	Signals <-chan os.Signal
	// Timeouts are the shutdown step budgets (zero fields use defaults).
	Timeouts hub.ShutdownTimeouts
	// TUI configures the local UI.
	TUI TUIOptions
}
