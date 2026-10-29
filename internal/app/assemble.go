package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/config"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport"
	"github.com/softwaresalt/intercom-go/internal/tui"
)

const (
	// hubSessionID names the process's one session (ADR MVP-D7); it is
	// stamped on every envelope.
	hubSessionID = "intercom"
	// teardownBudget bounds a forced teardown (start failure, Stop).
	teardownBudget = 15 * time.Second
)

// maxApproval caps [timeouts].approval_seconds so the seconds-to-Duration
// conversion cannot overflow. config Validate already bounds the value
// (config.MaxApprovalTimeoutSeconds); this one-year clamp is defence in depth
// for programmatic configs that bypass Validate.
const maxApproval = 365 * 24 * time.Hour

// defaultTeardownPermissions is the budget for failing parked permission asks
// in a forced teardown when [hub.ShutdownTimeouts].Permissions is unset; it
// matches the hub's own default for step 0.
const defaultTeardownPermissions = 5 * time.Second

// unconsumedKeys lists, sorted, the config keys that are accepted and
// validated but not yet consumed by any component and that differ from
// [config.Default]. Only key names are returned, never values.
func unconsumedKeys(cfg *config.Config) []string {
	def := config.Default()
	var keys []string
	add := func(differs bool, name string) {
		if differs {
			keys = append(keys, name)
		}
	}
	add(cfg.MaxConcurrentSessions != def.MaxConcurrentSessions, "max_concurrent_sessions")
	add(cfg.RetentionDays != def.RetentionDays, "retention_days")
	add(cfg.OperatorDetailLevel != def.OperatorDetailLevel, "operator_detail_level")
	add(cfg.Database != def.Database, "database.path")
	add(cfg.Timeouts.PromptSeconds != def.Timeouts.PromptSeconds, "timeouts.prompt_seconds")
	add(cfg.Timeouts.WaitSeconds != def.Timeouts.WaitSeconds, "timeouts.wait_seconds")
	add(cfg.Stall.Enabled != def.Stall.Enabled, "stall.enabled")
	add(cfg.Stall.InactivityThresholdSeconds != def.Stall.InactivityThresholdSeconds, "stall.inactivity_threshold_seconds")
	add(cfg.Stall.EscalationThresholdSeconds != def.Stall.EscalationThresholdSeconds, "stall.escalation_threshold_seconds")
	add(cfg.Stall.MaxRetries != def.Stall.MaxRetries, "stall.max_retries")
	add(cfg.Stall.DefaultNudgeMessage != def.Stall.DefaultNudgeMessage, "stall.default_nudge_message")
	add(!maps.Equal(cfg.Commands, def.Commands), "commands")
	slices.Sort(keys)
	return keys
}

// approvalTimeout maps [timeouts].approval_seconds to the hub's permission
// deadline; 0 selects the hub default.
func approvalTimeout(seconds uint64) time.Duration {
	if seconds >= uint64(maxApproval/time.Second) {
		return maxApproval
	}
	return time.Duration(seconds) * time.Second //nolint:gosec // bounded just above
}

// App is a started process: the handles callers (cmd, tests) use.
type App struct {
	log       *slog.Logger
	cfg       *config.Config
	workspace string

	hub        *hub.Hub
	hubCancel  context.CancelFunc
	hubRunDone chan struct{}
	remote     *transport.Remote
	serveDone  chan struct{}
	agent      agent.Agent
	ui         *localUI

	mu     sync.Mutex
	mintMu sync.Mutex         // serializes mint/re-issue so two cannot interleave their role order
	creds  []transport.Issued // live credentials, for LOCAL display only
	// rejectCh is a capacity-1 mailbox for rejected-upgrade notices that keeps
	// the highest Total; producers write it under mu, the pump is the one reader.
	rejectCh chan tui.SecurityNoticeMsg
	// credsCleared is sticky: once set (shutdown, failed start) no credential
	// is ever published into creds again. Guarded by mu.
	credsCleared bool
	// zeroAfterTUI zeroes the plaintext in creds once the TUI has been given
	// it. Set before the TUI starts and never written afterwards.
	zeroAfterTUI bool
	// afterMint is a test hook called after each successful remote.Mint and
	// before the credential is published; nil in production.
	afterMint func(hub.Role)
	// runErrs are failures during the run (TUI, listener) joined into err.
	runErrs []error

	done     chan struct{}
	doneOnce sync.Once

	// shut is the shutdown state: see shutdownState.
	shut shutdownState
}

// Start assembles and starts the process; it returns once the process is
// ready: config loaded and validated, workspace contained, hub running, agent
// started (the SDK's Start is the readiness gate, design §3), listener
// serving, credentials minted and the local TUI attached. On any failure
// everything already started is torn down and no App is returned.
//
// Startup order: config -> workspace (pathsafe) -> hub -> remote bind (not yet
// serving) -> agent build+start -> serve -> credentials -> TUI. The hub runs
// first because the agent's callbacks block on its ingress; the listener is
// bound early so a port conflict fails before the CLI child is spawned, but
// serves only once the agent is ready and credentials exist.
func Start(ctx context.Context, o Options) (*App, error) {
	if o.NewAgent == nil {
		return nil, apperr.New(apperr.KindConfig, "app: an agent factory is required")
	}
	log := o.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	if o.Now == nil {
		o.Now = time.Now
	}

	if !o.TUI.Disabled {
		if err := requireTerminal(o.TUI); err != nil {
			return nil, apperr.Wrap(apperr.KindConfig, err)
		}
	}

	cfg, report, err := config.Load(o.ConfigPath)
	if err != nil {
		return nil, err
	}
	for _, w := range report.Warnings {
		log.Warn("config warning", "warning", w)
	}
	if len(report.UnknownKeys) > 0 {
		log.Warn("config has unknown keys", "keys", report.UnknownKeys)
	}
	if keys := unconsumedKeys(cfg); len(keys) > 0 {
		log.Warn("config has accepted-but-unconsumed keys", "keys", keys)
	}
	workspace, err := resolveWorkspace(cfg, o.Workspace)
	if err != nil {
		return nil, err
	}
	port := cfg.HTTPPort
	if o.PortSet {
		port = o.Port
	}

	a := &App{
		log: log, cfg: cfg, workspace: workspace,
		shut:     shutdownState{timeouts: o.Timeouts, req: make(chan struct{})},
		done:     make(chan struct{}),
		rejectCh: make(chan tui.SecurityNoticeMsg, 1),
	}
	// The shutdown context is deliberately NOT derived from ctx or from any
	// signal: a SIGINT must trigger the ordered sequence, not cancel it (and
	// the hub runs on a context nothing cancels before step 6).
	a.shut.ctx, a.shut.hurry = context.WithCancel(context.Background())

	// ctx (and a signal) bound startup only.
	startCtx, cancelStart := context.WithCancel(ctx)
	defer cancelStart()
	if o.Signals != nil {
		go a.watchSignals(o.Signals, cancelStart)
	}
	if err := a.assemble(startCtx, o, port); err != nil {
		tctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownBudget)
		defer cancel()
		if terr := a.teardown(tctx); terr != nil {
			log.Error("teardown after failed start", "error", terr)
			err = errors.Join(err, fmt.Errorf("teardown after failed start: %w", terr))
		}
		a.shut.hurry()
		a.finish()
		return nil, err
	}
	go a.supervise()
	return a, nil
}

// assemble builds and starts the components in order, recording each in a as
// it comes up so a failure can be torn down by a.teardown.
func (a *App) assemble(ctx context.Context, o Options, port uint16) error {
	cfg := a.cfg

	// Hub first: the agent's callbacks block on its ingress.
	late := &lateAgent{}
	h, err := hub.New(hub.Config{
		SessionID:                hubSessionID,
		Clock:                    o.Now,
		Now:                      o.Now,
		Logger:                   a.log,
		MaxRemote:                cfg.Remote.MaxConnections,
		RemoteApproveDestructive: cfg.Remote.RemoteApproveDestructive,
		PermissionTimeout:        approvalTimeout(cfg.Timeouts.ApprovalSeconds),
		Agent:                    late,
	})
	if err != nil {
		return err
	}
	a.hub = h
	hubCtx, hubCancel := context.WithCancel(context.Background()) // never the signal context: see Shutdown
	a.hubCancel = hubCancel
	a.hubRunDone = make(chan struct{})
	go func() {
		defer close(a.hubRunDone)
		if err := h.Run(hubCtx); err != nil {
			a.log.Error("hub stopped with error", "error", err)
			// Recorded before hubRunDone closes, so the shutdown that awaits it
			// joins the failure (e.g. a stuck lane worker) into Err().
			a.addErr(fmt.Errorf("hub: %w", err))
		}
	}()

	remote, err := transport.NewRemote(transport.RemoteOptions{
		Hub:      h,
		Remote:   cfg.Remote,
		Port:     port,
		Now:      o.Now,
		Rand:     o.CredentialRand,
		Logger:   a.log,
		OnReject: a.onReject,
	})
	if err != nil {
		return err
	}
	a.remote = remote

	startup := time.Duration(cfg.Copilot.StartupTimeoutSeconds) * time.Second //nolint:gosec // bounded by config validation
	ag, err := o.NewAgent(AgentDeps{
		WorkingDirectory: a.workspace,
		CLIPath:          cfg.Copilot.CLIPath,
		StartupTimeout:   startup,
		Ingress:          h.Ingress(),
		Asks:             h.Asks(),
		HubDone:          h.Done(),
		Logger:           a.log,
	})
	if err != nil {
		return fmt.Errorf("build agent: %w", err)
	}
	a.agent = ag
	late.bind(ag)

	startCtx, cancel := context.WithTimeout(ctx, startup)
	handle, err := ag.Start(startCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("agent start: %w", err)
	}
	a.log.Info("agent started", "session", handle.ID)

	a.serveDone = make(chan struct{})
	go func() {
		defer close(a.serveDone)
		a.onServeExit(remote.Serve())
	}()

	if err := a.mintAll(); err != nil {
		return err
	}

	if !o.TUI.Disabled {
		// startUI honours the startup ctx (a cancelled start fails it) and itself
		// gives the feed a context that outlives startup; ui.stop releases it.
		// The one display of the credentials: the local TUI only. They are given
		// to the model at construction so no later message can race Ctrl+R.
		creds := a.credentialsMsg(a.Credentials())
		a.zeroAfterTUI = !o.TUI.RetainPlaintext
		if a.zeroAfterTUI {
			a.zeroPlaintext()
		}
		ui, err := startUI(ctx, h, o.TUI, creds, a.tuiReissue)
		if err != nil {
			return fmt.Errorf("start tui: %w", err)
		}
		a.ui = ui
		// Registered before the watcher starts, so a stop that waits on it can
		// never miss the runtime error it records.
		ui.watchDone = make(chan struct{})
		go a.watchUI(ui)
	}

	a.log.Info("intercom ready", "addr", a.remote.Addr().String(), "workspace", a.workspace)
	return nil
}

// onReject is the transport's rejected-upgrade observer. It runs on request
// goroutines, so it only stores the latest notice in the capacity-1 mailbox:
// it never blocks, never reads a.ui (the TUI may not exist yet, or any more)
// and never spawns work per request. The mailbox keeps the notice with the
// highest Total, because notices from concurrent request goroutines can arrive
// out of order. Producers hold mu, so after draining the slot the send cannot
// find the mailbox full.
func (a *App) onReject(status int, reason string, total uint64) {
	msg := tui.SecurityNoticeMsg{Status: status, Reason: reason, Total: total}
	a.mu.Lock()
	defer a.mu.Unlock()
	select {
	case pending := <-a.rejectCh:
		if pending.Total > msg.Total {
			msg = pending
		}
	default:
	}
	select {
	case a.rejectCh <- msg:
	default:
	}
}

// watchUI supervises the local TUI until its program exits: it starts the
// rejection-notice pump, waits for the program and the pump to finish, records
// a runtime TUI failure (before closing ui.watchDone, so a stop that waits on
// it can never miss the error) and then triggers the ordered shutdown. The
// caller must assign ui.watchDone synchronously before spawning watchUI.
func (a *App) watchUI(ui *localUI) {
	defer close(ui.watchDone)
	pumpDone := make(chan struct{})
	go func() {
		defer close(pumpDone)
		a.pumpRejects(ui)
	}()
	<-ui.done
	<-pumpDone
	if err := ui.failure(); err != nil {
		a.log.Error("tui exited with error", "error", err)
		a.addErr(fmt.Errorf("tui: %w", err))
		a.Shutdown("tui failed")
		return
	}
	a.Shutdown("tui quit")
}

// pumpRejects forwards the latest rejection notice to the local TUI from a
// single goroutine until the program has exited. Send returns once the program
// took the message or ended, so the loop is bounded by ui.done.
func (a *App) pumpRejects(ui *localUI) {
	for {
		select {
		case msg := <-a.rejectCh:
			ui.prog.Send(msg)
		case <-ui.done:
			return
		}
	}
}

// finish closes Done once.
func (a *App) finish() { a.doneOnce.Do(func() { close(a.done) }) }

// Done is closed when the process has finished shutting down (or failed to
// start): the §3.1 sequence ran, the hub is stopped, the listener is closed
// and the local TUI has exited.
func (a *App) Done() <-chan struct{} { return a.done }

// Hub returns the running hub.
func (a *App) Hub() *hub.Hub { return a.hub }

// Addr returns the bound loopback address.
func (a *App) Addr() net.Addr { return a.remote.Addr() }

// WSURL returns the WebSocket URL; it never carries a credential.
func (a *App) WSURL() string { return "ws://" + a.remote.Addr().String() + transport.WSPath }

// Workspace returns the pathsafe-resolved workspace root.
func (a *App) Workspace() string { return a.workspace }

// Config returns the loaded, validated configuration.
func (a *App) Config() *config.Config { return a.cfg }

// teardown is the forced unwind used only when startup fails: it stops
// whatever was started, in an order that keeps the hub alive until the agent
// is gone (SDK teardown produces events and unblocks handlers that need a live
// hub). The hub's step 0 runs first, bounded, so a parked permission ask
// resolves to UserNotAvailable at once instead of waiting on the agent stop.
// The ordered design §3.1 shutdown of a running process is [App.Shutdown].
func (a *App) teardown(ctx context.Context) error {
	a.clearCreds()
	var errs []error
	if a.hub != nil {
		permBudget := a.shut.timeouts.Permissions
		if permBudget <= 0 {
			permBudget = defaultTeardownPermissions
		}
		pctx, cancel := context.WithTimeout(ctx, permBudget)
		if err := a.hub.BeginShutdown(pctx); err != nil {
			errs = append(errs, fmt.Errorf("fail pending permissions: %w", err))
		}
		cancel()
	}
	// The UI and the listener go before the agent here (unlike a running
	// shutdown, where the agent is already gone), so teardown cannot call reap
	// as one unit: it runs the same steps reap is built from, in its own order.
	if err := a.stopUI(); err != nil {
		errs = append(errs, err)
	}
	if err := a.stopRemote(ctx); err != nil {
		errs = append(errs, err)
	}
	if a.agent != nil {
		if err := a.agent.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("agent stop: %w", err))
		}
	}
	a.awaitHub()
	errs = append(errs, a.takeErrs()...) // e.g. a Hub.Run failure recorded before hubRunDone closed
	return errors.Join(errs...)
}
