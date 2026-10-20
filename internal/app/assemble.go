package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
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

	done     chan struct{}
	doneOnce sync.Once
	stopOnce sync.Once
	stopErr  error
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
	workspace, err := resolveWorkspace(cfg, o.Workspace)
	if err != nil {
		return nil, err
	}
	port := cfg.HTTPPort
	if o.PortSet {
		port = o.Port
	}

	a := &App{log: log, cfg: cfg, workspace: workspace, done: make(chan struct{})}
	if err := a.assemble(ctx, o, port); err != nil {
		tctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), teardownBudget)
		defer cancel()
		if terr := a.teardown(tctx); terr != nil {
			log.Error("teardown after failed start", "error", terr)
		}
		a.finish()
		return nil, err
	}
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
		Hub:    h,
		Remote: cfg.Remote,
		Port:   port,
		Now:    o.Now,
		Rand:   o.CredentialRand,
		Logger: a.log,
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
		if err := remote.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("remote listener failed", "error", err)
		}
	}()

	if err := a.mintAll(); err != nil {
		return err
	}

	if !o.TUI.Disabled {
		// The UI outlives Start: Start's ctx bounds startup only, so the feed
		// (whose lifetime is its Start ctx) gets a context that is not cancelled
		// with it; ui.stop releases it.
		ui, err := startUI(context.WithoutCancel(ctx), h, o.TUI, a.tuiReissue)
		if err != nil {
			return fmt.Errorf("start tui: %w", err)
		}
		a.ui = ui
		ui.prog.Send(a.credentialsMsg(a.Credentials())) // the one display of the credentials: local TUI only
		go func() {
			<-ui.done
			a.finish()
		}()
	}

	a.log.Info("intercom ready", "addr", a.remote.Addr().String(), "workspace", a.workspace)
	return nil
}

// mintAll issues the approver and observer credentials. Only the role, the
// fingerprint and the expiry are ever logged.
func (a *App) mintAll() error {
	_, err := a.mint(nil)
	return err
}

// credentialRoles are the roles minted, in display order.
var credentialRoles = [...]hub.Role{hub.RoleApprover, hub.RoleObserver}

// mint issues a fresh credential for every role, each revoking that role's
// previous one, and publishes the live set. A failure part-way keeps the
// roles already replaced and the old credential of the rest, so Credentials
// always describes what is actually live. Only role, fingerprint and expiry
// are logged.
func (a *App) mint(current []transport.Issued) ([]transport.Issued, error) {
	a.mintMu.Lock()
	defer a.mintMu.Unlock()
	next := make([]transport.Issued, len(credentialRoles))
	copy(next, current)
	var firstErr error
	for i, role := range credentialRoles {
		iss, err := a.remote.Mint(role)
		if err != nil {
			firstErr = fmt.Errorf("mint %s credential: %w", roleName(role), err)
			break
		}
		a.log.Info("credential issued", "role", roleName(role), "fingerprint", iss.Fingerprint, "expires_at", iss.ExpiresAt)
		next[i] = iss
		a.mu.Lock()
		a.creds = append([]transport.Issued(nil), next...)
		a.mu.Unlock()
	}
	return next, firstErr
}

// Reissue mints a new credential for every role and revokes the previous ones
// (re-issue; ADR MVP-2). Sockets authenticated by a revoked credential are
// closed by the transport. The plaintext is for the local TUI only.
func (a *App) Reissue(context.Context) ([]transport.Issued, error) {
	fresh, err := a.mint(a.Credentials())
	if err != nil {
		return nil, err
	}
	a.log.Info("credentials re-issued", "approver", fresh[0].Fingerprint, "observer", fresh[1].Fingerprint)
	return fresh, nil
}

// tuiReissue is the TUI's re-issue key handler.
func (a *App) tuiReissue(ctx context.Context) (tui.CredentialsMsg, error) {
	fresh, err := a.Reissue(ctx)
	if err != nil {
		return tui.CredentialsMsg{}, err
	}
	return a.credentialsMsg(fresh), nil
}

// credentialsMsg is the one place the plaintext leaves the minted records: it
// builds the local TUI's credential panel, which is delivered by
// tea.Program.Send and nowhere else.
func (a *App) credentialsMsg(issued []transport.Issued) tui.CredentialsMsg {
	msg := tui.CredentialsMsg{URL: a.WSURL(), Credentials: make([]tui.Credential, 0, len(issued))}
	for _, iss := range issued {
		msg.Credentials = append(msg.Credentials, tui.Credential{
			Role:        roleName(iss.Role),
			Secret:      iss.Token.Reveal(),
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

// Done is closed when the process has nothing left to do: the local TUI has
// exited or the process has been stopped.
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

// Stop forcibly tears the process down: TUI, remote surface, agent, hub. It is
// the start-failure and test-cleanup path and does NOT follow the design §3.1
// shutdown order; it is idempotent.
func (a *App) Stop(ctx context.Context) error {
	a.stopOnce.Do(func() {
		a.stopErr = a.teardown(ctx)
		a.finish()
	})
	return a.stopErr
}

// teardown stops whatever was started, in an order that keeps the hub alive
// until the agent is gone (SDK teardown produces events and unblocks handlers
// that need a live hub).
func (a *App) teardown(ctx context.Context) error {
	var errs []error
	if a.ui != nil {
		a.ui.stop()
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
