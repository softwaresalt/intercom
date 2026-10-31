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
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	copilotsdk "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/pathsafe"
)

const (
	// ValidatedCLIVersion is the Copilot CLI/runtime version floor validated
	// by the C2 proving spike (ADR MVP-D10). A lower version fails Start.
	ValidatedCLIVersion = "1.0.84-1"

	// DefaultStopTimeout is the default Config.StopTimeout.
	DefaultStopTimeout = 5 * time.Second

	defaultOpTimeout = 30 * time.Second

	// SDKRuntimeShutdownTimeout is the SDK's bound on the graceful runtime
	// shutdown phase of Client.Stop, during which it holds its lifecycle lock.
	SDKRuntimeShutdownTimeout = 10 * time.Second
	// SDKProcessExitTimeout is the SDK's bound on reaping the CLI process after
	// the runtime shutdown phase, still under the same lifecycle lock.
	SDKProcessExitTimeout = 10 * time.Second
	// DefaultSDKStopLockTimeout is the worst-case time an in-flight
	// Client.Stop can hold the SDK lifecycle lock that ForceStop must acquire.
	DefaultSDKStopLockTimeout = SDKRuntimeShutdownTimeout + SDKProcessExitTimeout
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
	// StopTimeout bounds callback quiescing in Stop and, independently, the wait
	// for the SDK's ForceStop (past which Stop reports agent.ErrCleanupIncomplete).
	// Default 5s. The hub's shutdown Force budget must exceed it (see
	// hub.ShutdownTimeouts.Force).
	StopTimeout time.Duration
	// DefaultOpTimeout applies to Send/Abort when the caller's context has no
	// deadline. Default 30s.
	DefaultOpTimeout time.Duration
	// SDKStopLockTimeout is the extra time ForceStop may wait because an
	// in-flight Client.Stop holds the SDK lifecycle lock. Zero defaults to
	// DefaultSDKStopLockTimeout for the real SDK client and to 0 for injected
	// fake clients; a negative value is clamped to 0.
	SDKStopLockTimeout time.Duration

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

// isRealSDKClient marks the production SDK client (see Config.SDKStopLockTimeout).
func (realClient) isRealSDKClient() {}

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
	unmapped unmappedLogGate // once-per-type warn gate for unmapped SDK events
	stop     chan struct{}   // closed once at Stop; never the ingress channel
	stopOnce sync.Once

	mu            sync.Mutex
	started       bool
	stopped       bool // no new operations accepted
	stopping      bool // a Stop is running or has finished
	clientStarted bool
	session       sdkSession
	startCancel   context.CancelFunc // cancels an in-flight Start
	startDone     chan struct{}      // closed when Start returns
	stopDone      chan struct{}      // closed when the first Stop finishes
	forceOnce     sync.Once          // single-flight ForceStop
	forceDone     chan struct{}      // closed when ForceStop returns
	// forceUnconfirmed is sticky: set when a ForceStop wait ended without the
	// SDK call returning (runtime death unconfirmed). It is never cleared;
	// doStop re-checks forceDone so a Start-failure cleanup that timed out is
	// not forgotten once fail() has marked the client not started.
	forceUnconfirmed bool
	// clientStopStart/clientStopActive record an in-flight Client.Stop, which
	// holds the SDK lifecycle lock ForceStop needs. Guarded by mu.
	clientStopStart  time.Time
	clientStopActive bool
	stopResult       error
	ops              context.Context // cancelled by Stop; parents Send/Abort calls
	opsCancel        context.CancelFunc
}

// New validates cfg and builds an Adapter backed by the real SDK client.
func New(cfg Config) (*Adapter, error) {
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	opts := &copilotsdk.ClientOptions{WorkingDirectory: cfg.WorkingDirectory}
	if cfg.CLIPath != "" {
		opts.Connection = copilotsdk.StdioConnection{Path: cfg.CLIPath}
	}
	return newAdapter(cfg, realClient{c: copilotsdk.NewClient(opts)}), nil
}

func newWithClient(cfg Config, c sdkClient) (*Adapter, error) {
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	return newAdapter(cfg, c), nil
}

func newAdapter(cfg Config, c sdkClient) *Adapter {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.StopTimeout <= 0 {
		cfg.StopTimeout = DefaultStopTimeout
	}
	if cfg.DefaultOpTimeout <= 0 {
		cfg.DefaultOpTimeout = defaultOpTimeout
	}
	if _, real := c.(interface{ isRealSDKClient() }); real && cfg.SDKStopLockTimeout == 0 {
		cfg.SDKStopLockTimeout = DefaultSDKStopLockTimeout
	}
	if cfg.SDKStopLockTimeout < 0 {
		cfg.SDKStopLockTimeout = 0
	}
	a := &Adapter{
		cfg:    cfg,
		client: c,
		log:    cfg.Logger,
		bridge: newPermissionBridge(cfg.Asks, cfg.HubDone, cfg.Logger),
		latch:  newLatch(),
		stop:   make(chan struct{}),
	}
	a.ops, a.opsCancel = context.WithCancel(context.Background())
	a.bridge.stop = a.stop
	return a
}

func validateConfig(cfg *Config) error {
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
	resolved, err := resolveCLIPath(cfg.CLIPath)
	if err != nil {
		return err
	}
	cfg.CLIPath = resolved
	return nil
}

// resolveCLIPath re-applies the §6.2 classification at the point of spawn
// (defence in depth; config.Validate is the primary gate) and returns the
// path to hand to the SDK. UNC, device-namespace, drive-relative and relative
// paths are rejected; an absolute path must be an existing regular file; a
// bare name is resolved once via exec.LookPath to an absolute regular file.
func resolveCLIPath(p string) (string, error) {
	switch pathsafe.ClassifyCLIPath(p) {
	case pathsafe.CLIPathEmpty:
		return "", nil
	case pathsafe.CLIPathUNCOrDevice:
		return "", fmt.Errorf("agent: cli_path %q must not be a UNC or device-namespace path", p)
	case pathsafe.CLIPathDriveRelative:
		return "", fmt.Errorf("agent: cli_path %q must not be drive-relative", p)
	case pathsafe.CLIPathRelative:
		return "", fmt.Errorf("agent: cli_path %q must not be relative", p)
	case pathsafe.CLIPathAbs:
		info, err := os.Stat(p)
		if err != nil {
			return "", fmt.Errorf("agent: cli_path %q does not exist: %w", p, err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("agent: cli_path %q must be a regular file", p)
		}
		return p, nil
	default: // pathsafe.CLIPathBare
		resolved, err := exec.LookPath(p)
		if err != nil {
			return "", fmt.Errorf("agent: cli_path %q not found in PATH: %w", p, err)
		}
		abs, err := filepath.Abs(resolved)
		if err != nil {
			return "", fmt.Errorf("agent: resolve cli_path %q: %w", p, err)
		}
		if pathsafe.ClassifyCLIPath(abs) != pathsafe.CLIPathAbs {
			return "", fmt.Errorf("agent: cli_path %q must not resolve to a UNC or device-namespace path", p)
		}
		info, err := os.Stat(abs)
		if err != nil {
			return "", fmt.Errorf("agent: stat resolved cli_path %q: %w", p, err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("agent: cli_path %q must resolve to a regular file", p)
		}
		return abs, nil
	}
}

// Start launches the runtime, asserts its version and protocol, and creates
// the single session.
func (a *Adapter) Start(ctx context.Context) (agent.SessionHandle, error) {
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		return agent.SessionHandle{}, fmt.Errorf("agent: start: %w", agent.ErrStopped)
	}
	if a.started {
		a.mu.Unlock()
		return agent.SessionHandle{}, fmt.Errorf("agent: start: one session per process: %w", agent.ErrAlreadyStarted)
	}
	a.started = true
	// The SDK launches the CLI with exec.Command (not CommandContext), so
	// cancelling this startup context when Start returns does not kill the
	// runtime; it only bounds the startup RPCs.
	ctx, cancel := context.WithTimeout(ctx, a.cfg.StartupTimeout)
	done := make(chan struct{})
	a.startCancel = cancel
	a.startDone = done
	a.mu.Unlock()
	defer close(done)
	defer cancel()

	// fail tears the runtime down. Marking the adapter stopped makes a
	// failed Start terminal, and closing the stop signal releases anything
	// already wired to it (a parked permission handler, a blocked callback).
	fail := func(err error) (agent.SessionHandle, error) {
		a.mu.Lock()
		a.stopped = true
		a.clientStarted = false
		a.mu.Unlock()
		a.stopOnce.Do(func() { close(a.stop) })
		a.latch.quiesce(0) // a failed Start closes the callback latch like Stop
		// Cleanup gets its own finite budget so a wedged ForceStop cannot
		// hold Start (or startDone) open; the original error is preserved.
		if cerr := a.forceStop(); cerr != nil {
			err = errors.Join(err, cerr)
		}
		return agent.SessionHandle{}, err
	}
	// interrupted reports a Stop that landed while Start was in flight.
	interrupted := func() bool {
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.stopped
	}
	errInterrupted := fmt.Errorf("agent: start interrupted: %w", agent.ErrStopped)

	if err := a.client.Start(ctx); err != nil {
		return fail(fmt.Errorf("agent: start copilot client: %w", err))
	}
	a.mu.Lock()
	a.clientStarted = true
	a.mu.Unlock()
	if interrupted() {
		return fail(errInterrupted)
	}

	status, err := a.client.GetStatus(ctx)
	if err != nil {
		return fail(fmt.Errorf("agent: query copilot runtime status: %w", err))
	}
	if err := assertRuntime(status); err != nil {
		return fail(err)
	}

	// The event handler is registered at session creation (SessionConfig.OnEvent),
	// not afterwards via Session.On, so events emitted during creation are not
	// lost (design §4.2: no silent drop). The latch refuses entrants after
	// Stop; the SDK stops invoking the handler once the session disconnects.
	streaming := true
	sess, err := a.client.CreateSession(ctx, &copilotsdk.SessionConfig{
		OnPermissionRequest: a.bridge.handler(),
		OnEvent:             a.onEvent,
		Streaming:           &streaming,
		WorkingDirectory:    a.cfg.WorkingDirectory,
	})
	if err != nil {
		return fail(fmt.Errorf("agent: create session: %w", err))
	}

	// Publish the session atomically with the stopped check so a Stop cannot
	// slip between them and leave a live session behind.
	a.mu.Lock()
	if a.stopped {
		a.mu.Unlock()
		_ = boundedCall(context.Background(), a.cfg.StopTimeout, sess.Disconnect)
		return fail(errInterrupted)
	}
	a.session = sess
	a.mu.Unlock()
	return agent.SessionHandle{ID: sess.ID()}, nil
}

// RequiredProtocolVersion is the SDK protocol version asserted at Start
// (ADR MVP-D10). It is pinned here rather than following the SDK constant so
// an SDK bump cannot silently move the pin.
const RequiredProtocolVersion = 3

func assertRuntime(st *copilotsdk.GetStatusResponse) error {
	if st == nil {
		return errors.New("agent: copilot runtime returned no status")
	}
	if st.ProtocolVersion != RequiredProtocolVersion {
		return fmt.Errorf("agent: copilot runtime protocol version %d does not match required %d", st.ProtocolVersion, RequiredProtocolVersion)
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

// versionAtLeast compares "MAJOR.MINOR.PATCH[-N]" versions numerically. The
// "-N" suffix follows Copilot CLI's prerelease convention (semver
// precedence): a final release such as 1.0.84 outranks 1.0.84-1, so a missing
// suffix sorts above any numeric one. A leading "v" and "+build" metadata are
// tolerated; any other suffix is an error (fail closed).
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
	v = strings.TrimPrefix(v, "v")
	v, _, _ = strings.Cut(v, "+")
	core, rev, hasRev := strings.Cut(v, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return out, fmt.Errorf("want MAJOR.MINOR.PATCH[-N], got %q", v)
	}
	for i, p := range parts {
		n, err := parseDigits(p)
		if err != nil {
			return out, fmt.Errorf("non-numeric component %q in %q", p, v)
		}
		out[i] = n
	}
	out[3] = math.MaxInt // final release: above every prerelease number
	if hasRev {
		n, err := parseDigits(rev)
		if err != nil {
			return out, fmt.Errorf("non-numeric revision %q in %q", rev, v)
		}
		out[3] = n
	}
	return out, nil
}

// parseDigits parses a non-empty string of ASCII digits (strconv.Atoi would
// also accept a leading sign).
func parseDigits(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty")
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errors.New("not a digit string")
		}
	}
	return strconv.Atoi(s)
}

// onEvent is the SDK callback. It follows the §4.2 contract: admitted by the
// latch, then a blocking send to ingress. The send is abandoned only when the
// adapter is stopping or the hub is gone, never because ingress is full.
func (a *Adapter) onEvent(ev copilotsdk.SessionEvent) {
	if raw := string(ev.Type()); isInformational(raw) {
		a.log.Debug("informational agent SDK event ignored", "event_type", raw)
		return
	}
	if !a.latch.enter() {
		if _, ok := ev.Data.(*rpc.SessionIdleData); ok && a.deliverLateIdle(ev) {
			return
		}
		a.log.Warn("agent event arrived after quiesce; discarded", "event_type", string(ev.Type()))
		return
	}
	defer a.latch.exit()

	e := normalizeWithGate(ev, a.log, &a.unmapped)
	switch e.Kind {
	case agent.KindTurnStart:
		e.TurnID = a.bridge.setTurn(e.TurnID)
	case agent.KindIdle:
		if e.Aborted {
			if id := a.bridge.abortTurn(); e.TurnID == "" {
				e.TurnID = id
			}
		} else {
			a.bridge.clearTurn()
		}
	}
	// Deliver without blocking whenever ingress has room, even during
	// shutdown, so an event is only abandoned when it genuinely cannot be
	// delivered.
	select {
	case a.cfg.Ingress <- e:
		return
	default:
	}
	select {
	case a.cfg.Ingress <- e:
	case <-a.stop:
		a.log.Warn("agent event discarded during shutdown", "event_type", string(ev.Type()))
	case <-a.cfg.HubDone:
		a.log.Warn("agent event discarded: hub is gone", "event_type", string(ev.Type()))
	}
}

// deliverLateIdle handles a terminal session.idle that arrives after the latch
// closed (between quiesce and Disconnect). The turn is closed out on the bridge
// and, unless the hub is gone, the event is offered to ingress without ever
// blocking. It reports whether the event was delivered.
func (a *Adapter) deliverLateIdle(ev copilotsdk.SessionEvent) bool {
	e := normalize(ev, a.log)
	if e.Aborted {
		if id := a.bridge.abortTurn(); e.TurnID == "" {
			e.TurnID = id
		}
	} else {
		a.bridge.clearTurn()
	}
	select {
	case <-a.cfg.HubDone:
		return false
	default:
	}
	select {
	case a.cfg.Ingress <- e:
		return true
	default:
		return false
	}
}

func (a *Adapter) activeSession() (sdkSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.stopped {
		return nil, fmt.Errorf("agent: session is not running: %w", agent.ErrNotRunning)
	}
	return a.session, nil
}

// opContext bounds an SDK call: it carries a deadline (the caller's, or the
// default) and is also cancelled when Stop begins, so an in-flight Send or
// Abort does not outlive shutdown.
func (a *Adapter) opContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); !ok {
		var cancelTimeout context.CancelFunc
		ctx, cancelTimeout = context.WithTimeout(ctx, a.cfg.DefaultOpTimeout)
		inner, cancel := context.WithCancel(ctx)
		stopWatch := context.AfterFunc(a.ops, cancel)
		return inner, func() { stopWatch(); cancel(); cancelTimeout() }
	}
	inner, cancel := context.WithCancel(ctx)
	stopWatch := context.AfterFunc(a.ops, cancel)
	return inner, func() { stopWatch(); cancel() }
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
	ctx, cancel := a.opContext(ctx)
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
	ctx, cancel := a.opContext(ctx)
	defer cancel()
	// Cancel the turn's parked permission asks as the abort is issued, so a
	// handler never outlives the turn it belongs to.
	a.bridge.abortTurn()
	if err := sess.Abort(ctx); err != nil {
		return fmt.Errorf("agent: abort turn: %w", err)
	}
	return nil
}

// Stop quiesces SDK callbacks (close the latch, then a bounded wait for
// in-flight callbacks to drain)
// and shuts the runtime down, escalating to ForceStop on failure. Every
// blocking SDK step is bounded by min(StopTimeout, ctx remaining), so Stop's
// worst case is a small multiple of StopTimeout even when the CLI hangs.
//
// Stop also interrupts an in-flight Start. It is idempotent: a concurrent or
// later call waits for the first to finish and returns its result. It never
// closes the ingress channel.
func (a *Adapter) Stop(ctx context.Context) error {
	a.mu.Lock()
	if a.stopping {
		done := a.stopDone
		a.mu.Unlock()
		select {
		case <-done:
		case <-ctx.Done():
			return fmt.Errorf("agent: waiting for stop in progress: %w", ctx.Err())
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		return a.stopResult
	}
	a.stopping = true
	a.stopped = true
	a.stopDone = make(chan struct{})
	cancelStart, startDone := a.startCancel, a.startDone
	a.mu.Unlock()

	err := a.doStop(ctx, cancelStart, startDone)

	a.mu.Lock()
	a.stopResult = err
	close(a.stopDone)
	a.mu.Unlock()
	return err
}

func (a *Adapter) budget(ctx context.Context) time.Duration {
	d := a.cfg.StopTimeout
	if dl, ok := ctx.Deadline(); ok {
		if rem := time.Until(dl); rem < d {
			d = max(rem, 0)
		}
	}
	return d
}

func (a *Adapter) doStop(ctx context.Context, cancelStart context.CancelFunc, startDone <-chan struct{}) error {
	a.stopOnce.Do(func() { close(a.stop) })
	a.opsCancel() // cancel in-flight Send/Abort

	var errs []error
	needForce := false
	ctxNoted := false
	// noteCtx records caller cancellation once and arranges a force stop: the
	// remaining waits return immediately, but the runtime must still die.
	noteCtx := func() {
		needForce = true
		if !ctxNoted {
			ctxNoted = true
			errs = append(errs, fmt.Errorf("agent: stop interrupted: %w", ctx.Err()))
		}
	}

	// Interrupt and await an in-flight Start so a late-arriving session or
	// client cannot outlive Stop.
	if cancelStart != nil {
		cancelStart()
	}
	if startDone != nil {
		t := time.NewTimer(a.budget(ctx))
		select {
		case <-startDone:
		case <-ctx.Done():
			noteCtx()
		case <-t.C:
			errs = append(errs, errors.New("agent: in-flight Start did not finish before the stop deadline"))
			needForce = true
		}
		t.Stop()
	}

	a.mu.Lock()
	sess, clientStarted := a.session, a.clientStarted
	a.mu.Unlock()

	// step runs one bounded SDK call and classifies how it ended. Once the
	// caller's context is done or an earlier step failed, graceful teardown is
	// over: the call is not attempted (it would only race the force kill) and
	// the runtime goes straight to ForceStop.
	step := func(label string, fn func() error) {
		if ctx.Err() != nil {
			noteCtx()
			return
		}
		if needForce {
			return
		}
		err := boundedCall(ctx, a.budget(ctx), fn)
		switch {
		case err == nil:
		case ctx.Err() != nil && errors.Is(err, ctx.Err()):
			noteCtx()
		default:
			errs = append(errs, fmt.Errorf("agent: %s: %w", label, err))
			needForce = true
		}
	}

	if !a.latch.quiesceCtx(ctx, a.budget(ctx)) {
		if ctx.Err() != nil {
			noteCtx()
		} else {
			errs = append(errs, errors.New("agent: callbacks did not drain before the quiesce deadline"))
		}
	}
	if sess != nil {
		step("disconnect session", sess.Disconnect)
	}
	if clientStarted {
		step("stop copilot client", a.trackedClientStop)
	}
	if needForce {
		if err := a.forceStop(); err != nil {
			errs = append(errs, err)
		}
	} else if err := a.recheckForce(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// trackedClientStop runs Client.Stop and records, for as long as it is in
// flight (even after a bounded wait gave up on it), that the SDK lifecycle
// lock is held.
func (a *Adapter) trackedClientStop() error {
	a.mu.Lock()
	a.clientStopStart, a.clientStopActive = time.Now(), true
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.clientStopActive = false
		a.mu.Unlock()
	}()
	return a.client.Stop()
}

// forceStopWait is how long a ForceStop wait may last: StopTimeout, extended by
// whatever remains of SDKStopLockTimeout while an in-flight Client.Stop still
// holds the SDK lifecycle lock that ForceStop queues behind.
func (a *Adapter) forceStopWait() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	d := a.cfg.StopTimeout
	if a.clientStopActive {
		d += max(a.cfg.SDKStopLockTimeout-time.Since(a.clientStopStart), 0)
	}
	return d
}

// ForceStopBudget is the worst-case duration of the adapter's ForceStop wait
// (StopTimeout plus the SDK lock allowance). The hub's shutdown Force budget
// must cover it.
func (a *Adapter) ForceStopBudget() time.Duration {
	return a.cfg.StopTimeout + a.cfg.SDKStopLockTimeout
}

// forceStop runs the SDK's ForceStop under its own finite budget (StopTimeout,
// independent of the caller's context, so cleanup still happens after the
// caller gives up). ForceStop can itself block on SDK lifecycle locking; if it
// does, the incomplete cleanup is reported rather than waited out.
func (a *Adapter) forceStop() error {
	// Single-flight: Start's failure path and Stop can both need a force stop;
	// the SDK's ForceStop runs at most once and every caller waits on it for a
	// bounded time.
	a.forceOnce.Do(func() {
		a.forceDone = make(chan struct{})
		go func() {
			a.client.ForceStop()
			close(a.forceDone)
		}()
	})
	wait := a.forceStopWait()
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-a.forceDone:
		return nil
	case <-t.C:
		select { // a ForceStop that has just returned beats the timer
		case <-a.forceDone:
			return nil
		default:
		}
		a.mu.Lock()
		a.forceUnconfirmed = true
		a.mu.Unlock()
		return fmt.Errorf("agent: %w: ForceStop did not return within %v", agent.ErrCleanupIncomplete, wait)
	}
}

// recheckForce re-waits (bounded) on a ForceStop that an earlier cleanup, such
// as a failed Start, left unconfirmed. It returns nil when no earlier wait
// timed out or the SDK call has since returned.
func (a *Adapter) recheckForce() error {
	a.mu.Lock()
	unconfirmed := a.forceUnconfirmed
	a.mu.Unlock()
	if !unconfirmed {
		return nil
	}
	// Non-blocking: the earlier cleanup already waited out StopTimeout.
	select {
	case <-a.forceDone:
		a.mu.Lock()
		a.forceUnconfirmed = false
		a.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("agent: %w: ForceStop has still not returned", agent.ErrCleanupIncomplete)
	}
}

// StopTimeout reports the adapter's bound on callback quiescing and on the
// wait for the SDK's ForceStop. The hub's shutdown executor reads it to keep
// its own Force budget above it.
func (a *Adapter) StopTimeout() time.Duration { return a.cfg.StopTimeout }

var errStepTimeout = errors.New("timed out")

// boundedCall runs fn and returns its result, or errStepTimeout if it has not
// returned within d, or ctx.Err() if ctx ends first. The goroutine is left to
// finish on its own; ForceStop releases it when the SDK call is wedged on a
// dead runtime.
func boundedCall(ctx context.Context, d time.Duration, fn func() error) error {
	res := make(chan error, 1)
	go func() { res <- fn() }()
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case err := <-res:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		select { // a result that is already ready beats the timer
		case err := <-res:
			return err
		default:
		}
		return errStepTimeout
	}
}
