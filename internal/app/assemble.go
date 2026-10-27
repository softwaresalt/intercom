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
	"sync/atomic"
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

// lateAgent breaks the construction cycle between the hub and the agent: the
// hub needs an AgentControl at New, but the adapter needs the hub's ingress,
// so the hub is given this forwarder and the real agent is bound afterwards.
// Until bound (or after a failed build) operator actions are refused with
// agent.ErrNotRunning, which the hub reports as agent_unavailable.
type lateAgent struct{ p atomic.Pointer[agent.Agent] }

func (l *lateAgent) bind(a agent.Agent) { l.p.Store(&a) }

func (l *lateAgent) Send(ctx context.Context, prompt string) error {
	a := l.p.Load()
	if a == nil {
		return agent.ErrNotRunning
	}
	return (*a).Send(ctx, prompt)
}

func (l *lateAgent) Abort(ctx context.Context) error {
	a := l.p.Load()
	if a == nil {
		return agent.ErrNotRunning
	}
	return (*a).Abort(ctx)
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
	// rejectCh is a capacity-1 latest-only mailbox for rejected-upgrade
	// notices; producers write it under mu, the pump is the one reader.
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

	// Shutdown: shutReq is closed once by the first trigger (signal, TUI quit,
	// Shutdown call); supervise then runs the §3.1 sequence. shutCtx bounds
	// the whole sequence and is cancelled (hurried) by a second signal.
	shutReq    chan struct{}
	shutOnce   sync.Once
	shutReason string          // written before shutReq closes
	shutCtx    context.Context //nolint:containedctx // process-lifetime shutdown budget, cancelled by hurry
	hurry      context.CancelFunc
	timeouts   hub.ShutdownTimeouts
	err        error // aggregated shutdown errors; written before done closes
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
		log: log, cfg: cfg, workspace: workspace, timeouts: o.Timeouts,
		done: make(chan struct{}), shutReq: make(chan struct{}),
		rejectCh: make(chan tui.SecurityNoticeMsg, 1),
	}
	// The shutdown context is deliberately NOT derived from ctx or from any
	// signal: a SIGINT must trigger the ordered sequence, not cancel it (and
	// the hub runs on a context nothing cancels before step 6).
	a.shutCtx, a.hurry = context.WithCancel(context.Background())

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
		}
		a.hurry()
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
		pumpDone := make(chan struct{})
		go func() {
			defer close(pumpDone)
			a.pumpRejects(ui)
		}()
		go func() {
			defer close(ui.watchDone)
			<-ui.done
			<-pumpDone
			if err := ui.failure(); err != nil {
				a.log.Error("tui exited with error", "error", err)
				a.addErr(fmt.Errorf("tui: %w", err))
				a.Shutdown("tui failed")
				return
			}
			a.Shutdown("tui quit")
		}()
	}

	a.log.Info("intercom ready", "addr", a.remote.Addr().String(), "workspace", a.workspace)
	return nil
}

// onReject is the transport's rejected-upgrade observer. It runs on request
// goroutines, so it only stores the latest notice in the capacity-1 mailbox:
// it never blocks, never reads a.ui (the TUI may not exist yet, or any more)
// and never spawns work per request. Producers hold mu, so after dropping a
// stale notice the send cannot find the mailbox full.
func (a *App) onReject(status int, reason string, total uint64) {
	msg := tui.SecurityNoticeMsg{Status: status, Reason: reason, Total: total}
	a.mu.Lock()
	defer a.mu.Unlock()
	select {
	case <-a.rejectCh:
	default:
	}
	select {
	case a.rejectCh <- msg:
	default:
	}
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

// mintAll issues the approver and observer credentials. Only the role, the
// fingerprint and the expiry are ever logged.
func (a *App) mintAll() error {
	_, err := a.mint()
	return err
}

// credentialRoles are the roles minted, in display order.
var credentialRoles = [...]hub.Role{hub.RoleApprover, hub.RoleObserver}

// mint issues a fresh credential for every role, each revoking that role's
// previous one, and publishes the live set after every role. It always returns
// the authoritative live set, with the error when a mint fails part-way: the
// roles already replaced keep their new credential (their old one is revoked)
// and the rest keep theirs, so neither Credentials nor a caller ever shows a
// revoked credential. The snapshot the set starts from is taken inside mintMu,
// so concurrent re-issues cannot start from a stale set. Only role,
// fingerprint and expiry are logged.
func (a *App) mint() ([]transport.Issued, error) {
	a.mintMu.Lock()
	defer a.mintMu.Unlock()
	if a.shuttingDown() {
		return nil, a.revokedForShutdown(nil)
	}
	next := a.Credentials() // under mintMu: the only writer of creds
	if len(next) != len(credentialRoles) {
		next = make([]transport.Issued, len(credentialRoles))
	}
	for i, role := range credentialRoles {
		iss, err := a.remote.Mint(role)
		if err != nil {
			if errors.Is(err, transport.ErrShuttingDown) || a.shuttingDown() {
				clear(next)
				return nil, a.revokedForShutdown(err)
			}
			err = fmt.Errorf("mint %s credential: %w", roleName(role), err)
			if i == 0 {
				return nil, err // nothing was replaced: there is no set to show
			}
			return next, err
		}
		a.log.Info("credential issued", "role", roleName(role), "fingerprint", iss.Fingerprint, "expires_at", iss.ExpiresAt)
		next[i] = iss
		if a.afterMint != nil {
			a.afterMint(role)
		}
		a.mu.Lock()
		if a.credsCleared || a.shuttingDown() {
			// Shutdown revoked (and cleared) the credentials while this mint was
			// in flight: never republish them.
			a.credsCleared = true
			a.creds = nil
			a.mu.Unlock()
			clear(next)
			return nil, a.revokedForShutdown(nil)
		}
		a.creds = append([]transport.Issued(nil), next...)
		a.mu.Unlock()
	}
	return next, nil
}

// revokedForShutdown clears the credentials and builds the error a mint
// returns once shutdown revoked them; cause, when non-nil, is the mint failure.
func (a *App) revokedForShutdown(cause error) error {
	a.clearCreds()
	err := errors.Join(transport.ErrCredentialRevoked, transport.ErrShuttingDown)
	if cause != nil {
		err = errors.Join(err, cause)
	}
	return err
}

// clearCreds drops the live credentials and makes that terminal: no later mint
// can publish into creds again. It takes only mu, so it is safe to call with
// or without mintMu held.
func (a *App) clearCreds() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.credsCleared = true
	for i := range a.creds {
		a.creds[i].Token = transport.Token{}
	}
	a.creds = nil
}

// zeroPlaintext drops the plaintext from the live credentials once the local
// TUI has been given it, keeping role, fingerprint and expiry.
func (a *App) zeroPlaintext() {
	a.mintMu.Lock()
	defer a.mintMu.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.creds {
		a.creds[i].Token = transport.Token{}
	}
}

// Reissue mints a new credential for every role and revokes the previous ones
// (re-issue; ADR MVP-2). Sockets authenticated by a revoked credential are
// closed by the transport. The plaintext is for the local TUI only. On a
// failure part-way it returns the authoritative live set together with the
// error (some roles may already have been replaced); when nothing was replaced
// or shutdown revoked the credentials it returns no set.
func (a *App) Reissue(context.Context) ([]transport.Issued, error) {
	live, err := a.mint()
	if err != nil {
		return live, err
	}
	a.log.Info("credentials re-issued", "approver", live[0].Fingerprint, "observer", live[1].Fingerprint)
	return live, nil
}

// tuiReissue is the TUI's re-issue key handler.
func (a *App) tuiReissue(ctx context.Context) (tui.CredentialsMsg, error) {
	live, err := a.Reissue(ctx)
	if len(live) == 0 {
		return tui.CredentialsMsg{}, err
	}
	// On a partial failure some displayed credentials may already be revoked:
	// hand the panel the authoritative live set with the error.
	msg := a.credentialsMsg(live)
	if a.zeroAfterTUI {
		a.zeroPlaintext()
	}
	return msg, err
}

// credentialsMsg is the one place the plaintext leaves the minted records: it
// builds the local TUI's credential panel, which is delivered by
// tea.Program.Send and nowhere else. Entries without plaintext (already
// delivered and zeroed) are skipped.
func (a *App) credentialsMsg(issued []transport.Issued) tui.CredentialsMsg {
	msg := tui.CredentialsMsg{URL: a.WSURL(), Credentials: make([]tui.Credential, 0, len(issued))}
	for _, iss := range issued {
		secret := iss.Token.Reveal()
		if secret == "" {
			continue
		}
		msg.Credentials = append(msg.Credentials, tui.Credential{
			Role:        roleName(iss.Role),
			Secret:      secret,
			Fingerprint: iss.Fingerprint,
			ExpiresAt:   iss.ExpiresAt,
		})
	}
	return msg
}

func roleName(r hub.Role) string {
	if r == hub.RoleApprover {
		return "approver"
	}
	return "observer"
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

// Credentials returns the live minted credentials. They exist for the LOCAL
// operator only (the TUI, and the in-process end-to-end test): never log,
// persist or transmit the plaintext.
func (a *App) Credentials() []transport.Issued {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]transport.Issued(nil), a.creds...)
}

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
		permBudget := a.timeouts.Permissions
		if permBudget <= 0 {
			permBudget = defaultTeardownPermissions
		}
		pctx, cancel := context.WithTimeout(ctx, permBudget)
		if err := a.hub.BeginShutdown(pctx); err != nil {
			errs = append(errs, fmt.Errorf("fail pending permissions: %w", err))
		}
		cancel()
	}
	if a.ui != nil {
		if err := a.ui.stop(); err != nil {
			errs = append(errs, err)
		}
	}
	if a.remote != nil {
		if err := a.remote.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("remote shutdown: %w", err))
		}
	}
	if a.agent != nil {
		if err := a.agent.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("agent stop: %w", err))
		}
	}
	if a.hub != nil {
		a.hub.Stop()
		<-a.hubRunDone
		a.hubCancel()
	}
	if a.serveDone != nil {
		<-a.serveDone
	}
	return errors.Join(errs...)
}
