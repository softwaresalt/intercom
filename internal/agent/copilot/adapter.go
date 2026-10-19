// Package copilot is the single Copilot SDK adapter: the only package in
// intercom-go permitted to import github.com/github/copilot-sdk (design §7.1;
// enforced by the copilot-sdk-boundary depguard rule). It converts SDK events,
// permission requests and decisions to and from the intercom-owned types in
// internal/agent so that no SDK type leaks past this package.
//
// SDK integration rules honoured here:
//   - §4.2 callback contract: blocking send to a buffered hub ingress, no
//     silent drop, one-way latch + refcount for quiescing, ingress never closed.
//   - §4.3 event-union safety: every type switch over event data has a
//     logging default branch (see normalize).
//   - §7.2 version pinning: CLI version floor and protocol asserted at Start.
//   - §7.5: every SDK call carries a deadline-bearing context.
//   - ADR MVP-D7: exactly one session per Adapter.
package copilot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

const (
	// ValidatedCLIVersion is the Copilot CLI/runtime version floor validated
	// by the C2 proving spike (ADR MVP-D10). A lower version fails Start.
	ValidatedCLIVersion = "1.0.84-1"

	defaultStopTimeout = 5 * time.Second
	defaultOpTimeout   = 30 * time.Second
)

// Config configures an Adapter.
type Config struct {
	// CLIPath selects the Copilot CLI binary: empty (SDK resolution), a bare
	// name, or an absolute path. Relative and drive-relative paths are
	// rejected (design §6.2).
	CLIPath string
	// WorkingDirectory is the workspace root for the runtime and session.
	WorkingDirectory string
	// StartupTimeout bounds Start ([copilot].startup_timeout_seconds).
	StartupTimeout time.Duration
	// StopTimeout bounds callback quiescing in Stop. Default 5s.
	StopTimeout time.Duration
	// DefaultOpTimeout applies to Send/Abort when the caller's context has no
	// deadline. Default 30s.
	DefaultOpTimeout time.Duration

	// Ingress receives normalized events. It must be buffered and drained by
	// a hub that never blocks (§4.1). The adapter never closes it.
	Ingress chan<- agent.Event
	// Asks receives permission asks for the hub to answer.
	Asks chan<- agent.PermissionAsk
	// HubDone is closed when the hub is gone; pending permission requests
	// then resolve to UserNotAvailable.
	HubDone <-chan struct{}

	Logger *slog.Logger
}

// sdkSession and sdkClient are the narrow test seam over the SDK. They are
// unexported and only ever satisfied inside this package.
type sdkSession interface {
	ID() string
	On(copilotsdk.SessionEventHandler) func()
	Send(context.Context, copilotsdk.MessageOptions) (string, error)
	Abort(context.Context) error
	Disconnect() error
}

type sdkClient interface {
	Start(context.Context) error
	GetStatus(context.Context) (*copilotsdk.GetStatusResponse, error)
	CreateSession(context.Context, *copilotsdk.SessionConfig) (sdkSession, error)
	Stop() error
	ForceStop()
}

type realClient struct{ c *copilotsdk.Client }

func (r realClient) Start(ctx context.Context) error { return r.c.Start(ctx) }
func (r realClient) GetStatus(ctx context.Context) (*copilotsdk.GetStatusResponse, error) {
	return r.c.GetStatus(ctx)
}
func (r realClient) CreateSession(ctx context.Context, cfg *copilotsdk.SessionConfig) (sdkSession, error) {
	s, err := r.c.CreateSession(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return realSession{s: s}, nil
}
func (r realClient) Stop() error { return r.c.Stop() }
func (r realClient) ForceStop()  { r.c.ForceStop() }

type realSession struct{ s *copilotsdk.Session }

func (r realSession) ID() string { return r.s.SessionID }
func (r realSession) On(h copilotsdk.SessionEventHandler) func() {
	return r.s.On(h)
}
func (r realSession) Send(ctx context.Context, o copilotsdk.MessageOptions) (string, error) {
	return r.s.Send(ctx, o)
}
func (r realSession) Abort(ctx context.Context) error { return r.s.Abort(ctx) }
func (r realSession) Disconnect() error               { return r.s.Disconnect() }

// Adapter implements agent.Agent over the Copilot SDK.
type Adapter struct {
	cfg    Config
	client sdkClient
	log    *slog.Logger
	bridge *permissionBridge

	latch    *latch
	stop     chan struct{} // closed once at Stop; never the ingress channel
	stopOnce sync.Once

	mu            sync.Mutex
	started       bool
	stopped       bool
	clientStarted bool
	session       sdkSession
	unsubscribe   func()
}

// New validates cfg and builds an Adapter backed by the real SDK client.
func New(cfg Config) (*Adapter, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	opts := &copilotsdk.ClientOptions{WorkingDirectory: cfg.WorkingDirectory}
	if cfg.CLIPath != "" {
		opts.Connection = copilotsdk.StdioConnection{Path: cfg.CLIPath}
	}
	return newAdapter(cfg, realClient{c: copilotsdk.NewClient(opts)}), nil
}

func newWithClient(cfg Config, c sdkClient) (*Adapter, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}
	return newAdapter(cfg, c), nil
}

func newAdapter(cfg Config, c sdkClient) *Adapter {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.StopTimeout <= 0 {
		cfg.StopTimeout = defaultStopTimeout
	}
	if cfg.DefaultOpTimeout <= 0 {
		cfg.DefaultOpTimeout = defaultOpTimeout
	}
	return &Adapter{
		cfg:    cfg,
		client: c,
		log:    cfg.Logger,
		bridge: newPermissionBridge(cfg.Asks, cfg.HubDone, cfg.Logger),
		latch:  newLatch(),
		stop:   make(chan struct{}),
	}
}

func validateConfig(cfg Config) error {
	if cfg.Ingress == nil {
		return errors.New("agent: Config.Ingress is required")
	}
	if cfg.Asks == nil {
		return errors.New("agent: Config.Asks is required")
	}
	if cfg.HubDone == nil {
		return errors.New("agent: Config.HubDone is required")
	}
	if cfg.StartupTimeout <= 0 {
		return errors.New("agent: Config.StartupTimeout must be positive")
	}
	return validateCLIPath(cfg.CLIPath)
}

// validateCLIPath re-applies the §6.2 classification at the point of spawn
// (defence in depth; config.Validate is the primary gate). Existence of an
// absolute path is checked by config; here only the classification matters:
// a relative or drive-relative path would resolve against the process working
// directory and must never reach the SDK.
func validateCLIPath(p string) error {
	switch {
	case p == "":
		return nil
	case filepath.IsAbs(p):
		return nil
	case len(p) >= 2 && p[1] == ':' && isASCIILetter(p[0]):
		return fmt.Errorf("agent: cli_path %q must not be drive-relative", p)
	case strings.ContainsAny(p, `/\`):
		return fmt.Errorf("agent: cli_path %q must not be relative", p)
	default:
		return nil // bare name: resolved via PATH (accepted residual risk, §6.2)
	}
}

func isASCIILetter(b byte) bool { return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') }

// Start launches the runtime, asserts its version and protocol, and creates
// the single session.
func (a *Adapter) Start(ctx context.Context) (agent.SessionHandle, error) {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return agent.SessionHandle{}, errors.New("agent: adapter is stopped")
	}
	if a.started {
		a.mu.Unlock()
		return agent.SessionHandle{}, errors.New("agent: already started; one session per process")
	}
	a.started = true
	a.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, a.cfg.StartupTimeout)
	defer cancel()

	fail := func(err error) (agent.SessionHandle, error) {
		a.mu.Lock()
		a.stopped = true
		a.mu.Unlock()
		a.client.ForceStop()
		return agent.SessionHandle{}, err
	}

	if err := a.client.Start(ctx); err != nil {
		return fail(fmt.Errorf("agent: start copilot client: %w", err))
	}
	a.mu.Lock()
	a.clientStarted = true
	a.mu.Unlock()

	status, err := a.client.GetStatus(ctx)
	if err != nil {
		return fail(fmt.Errorf("agent: query copilot runtime status: %w", err))
	}
	if err := assertRuntime(status); err != nil {
		return fail(err)
	}

	streaming := true
	sess, err := a.client.CreateSession(ctx, &copilotsdk.SessionConfig{
		OnPermissionRequest: a.bridge.handler(),
		Streaming:           &streaming,
		WorkingDirectory:    a.cfg.WorkingDirectory,
	})
	if err != nil {
		return fail(fmt.Errorf("agent: create session: %w", err))
	}

	unsub := sess.On(a.onEvent)
	a.mu.Lock()
	a.session = sess
	a.unsubscribe = unsub
	a.mu.Unlock()
	return agent.SessionHandle{ID: sess.ID()}, nil
}

func assertRuntime(st *copilotsdk.GetStatusResponse) error {
	if st == nil {
		return errors.New("agent: copilot runtime returned no status")
	}
	if st.ProtocolVersion != copilotsdk.SDKProtocolVersion {
		return fmt.Errorf("agent: copilot runtime protocol version %d does not match required %d", st.ProtocolVersion, copilotsdk.SDKProtocolVersion)
	}
	ok, err := versionAtLeast(st.Version, ValidatedCLIVersion)
	if err != nil {
		return fmt.Errorf("agent: copilot CLI version %q is unparseable: %w", st.Version, err)
	}
	if !ok {
		return fmt.Errorf("agent: copilot CLI version %q is below the validated floor %s", st.Version, ValidatedCLIVersion)
	}
	return nil
}

// versionAtLeast compares "MAJOR.MINOR.PATCH[-REV]" versions numerically. A
// missing -REV is revision 0; a non-numeric suffix is an error (fail closed).
func versionAtLeast(got, floor string) (bool, error) {
	g, err := parseVersion(got)
	if err != nil {
		return false, err
	}
	f, err := parseVersion(floor)
	if err != nil {
		return false, err
	}
	for i := range g {
		if g[i] != f[i] {
			return g[i] > f[i], nil
		}
	}
	return true, nil
}

func parseVersion(v string) ([4]int, error) {
	var out [4]int
	core, rev, hasRev := strings.Cut(v, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return out, fmt.Errorf("want MAJOR.MINOR.PATCH[-REV], got %q", v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, fmt.Errorf("non-numeric component %q in %q", p, v)
		}
		out[i] = n
	}
	if hasRev {
		n, err := strconv.Atoi(rev)
		if err != nil || n < 0 {
			return out, fmt.Errorf("non-numeric revision %q in %q", rev, v)
		}
		out[3] = n
	}
	return out, nil
}

// onEvent is the SDK callback. It follows the §4.2 contract: admitted by the
// latch, then a blocking send to ingress. The send is abandoned only when the
// adapter is stopping or the hub is gone, never because ingress is full.
func (a *Adapter) onEvent(ev copilotsdk.SessionEvent) {
	if !a.latch.enter() {
		a.log.Debug("agent event arrived after quiesce; discarded", "event_type", string(ev.Type()))
		return
	}
	defer a.latch.exit()

	e := normalize(ev, a.log)
	select {
	case a.cfg.Ingress <- e:
	case <-a.stop:
		a.log.Warn("agent event discarded during shutdown", "event_type", string(ev.Type()))
	case <-a.cfg.HubDone:
		a.log.Warn("agent event discarded: hub is gone", "event_type", string(ev.Type()))
	}
}

func (a *Adapter) activeSession() (sdkSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.stopped {
		return nil, errors.New("agent: session is not running")
	}
	return a.session, nil
}

func (a *Adapter) withDeadline(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, a.cfg.DefaultOpTimeout)
}

// Send submits a prompt. SendAndWait is deliberately not used: its documented
// timeout parameter does not exist (§7.5), so a deadline-bearing context is
// always supplied instead.
func (a *Adapter) Send(ctx context.Context, prompt string) error {
	sess, err := a.activeSession()
	if err != nil {
		return err
	}
	if strings.TrimSpace(prompt) == "" {
		return errors.New("agent: send prompt: prompt is empty")
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("agent: send prompt: %w", err)
	}
	ctx, cancel := a.withDeadline(ctx)
	defer cancel()
	if _, err := sess.Send(ctx, copilotsdk.MessageOptions{Prompt: prompt}); err != nil {
		return fmt.Errorf("agent: send prompt: %w", err)
	}
	return nil
}

// Abort cancels the in-flight turn.
func (a *Adapter) Abort(ctx context.Context) error {
	sess, err := a.activeSession()
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("agent: abort turn: %w", err)
	}
	ctx, cancel := a.withDeadline(ctx)
	defer cancel()
	if err := sess.Abort(ctx); err != nil {
		return fmt.Errorf("agent: abort turn: %w", err)
	}
	return nil
}

// Stop quiesces SDK callbacks (latch, then Unsubscribe, then a bounded wait)
// and shuts the runtime down, escalating to ForceStop on failure. It is
// idempotent and never closes the ingress channel.
func (a *Adapter) Stop(ctx context.Context) error {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return nil
	}
	a.stopped = true
	sess, unsub, clientStarted := a.session, a.unsubscribe, a.clientStarted
	a.mu.Unlock()

	a.stopOnce.Do(func() { close(a.stop) })
	if unsub != nil {
		unsub()
	}

	timeout := a.cfg.StopTimeout
	if dl, ok := ctx.Deadline(); ok {
		if rem := time.Until(dl); rem < timeout {
			timeout = max(rem, 0)
		}
	}

	var errs []error
	if !a.latch.quiesce(timeout) {
		errs = append(errs, errors.New("agent: callbacks did not drain before the quiesce deadline"))
	}
	if sess != nil {
		if err := sess.Disconnect(); err != nil {
			errs = append(errs, fmt.Errorf("agent: disconnect session: %w", err))
		}
	}
	if clientStarted {
		if err := a.client.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("agent: stop copilot client: %w", err))
			a.client.ForceStop()
		}
	}
	return errors.Join(errs...)
}
