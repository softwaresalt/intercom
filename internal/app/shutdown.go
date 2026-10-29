package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// remoteShutdownBudget bounds the final transport teardown that follows the
// ordered sequence (every client was already closed at step 5, so this only
// reaps the connection goroutines and the HTTP server).
const remoteShutdownBudget = 10 * time.Second

// shutdownState groups the App's shutdown bookkeeping: req is closed once by
// the first trigger (signal, TUI quit, Shutdown call) and supervise then runs
// the §3.1 sequence; ctx bounds the whole sequence and is cancelled (hurried)
// by a second signal.
type shutdownState struct {
	req      chan struct{}
	once     sync.Once
	reason   string          // written before req closes
	ctx      context.Context //nolint:containedctx // process-lifetime shutdown budget, cancelled by hurry
	hurry    context.CancelFunc
	timeouts hub.ShutdownTimeouts
	err      error // aggregated shutdown errors; written before done closes
}

// Shutdown requests the ordered shutdown of design §3.1. It is the single
// trigger behind a SIGINT/SIGTERM, a quit in the local TUI, and callers (the
// end-to-end test). It never blocks and is idempotent: only the first call
// (and its reason) counts. Wait on [App.Done] for completion.
func (a *App) Shutdown(reason string) {
	a.shut.once.Do(func() {
		a.shut.reason = reason
		close(a.shut.req)
	})
}

// Err reports the aggregated shutdown errors once [App.Done] is closed (nil
// before that, and nil for a clean shutdown). A non-nil Err means the process
// should exit non-zero: a step failed, the shutdown was interrupted, or the
// agent's CLI child may be orphaned (hub.ErrOrphanRisk).
func (a *App) Err() error {
	select {
	case <-a.done:
		return a.shut.err
	default:
		return nil
	}
}

// onServeExit handles the remote listener's accept loop ending. An orderly
// close (nil, or http.ErrServerClosed from Shutdown) is not a failure; anything
// else is logged, kept for [App.Err] and triggers the ordered shutdown, so a
// dead listener cannot leave a process that looks alive but serves nobody.
func (a *App) onServeExit(err error) {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return
	}
	a.log.Error("remote listener failed", "error", err)
	a.addErr(fmt.Errorf("remote listener: %w", err))
	a.Shutdown("remote listener failed")
}

// addErr records a failure to be joined into [App.Err]. Record it before the
// Shutdown call it accompanies: the shutdown sequence reads it after shut.req
// closes.
func (a *App) addErr(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.runErrs = append(a.runErrs, err)
}

func (a *App) takeErrs() []error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]error(nil), a.runErrs...)
}

// watchSignals turns interrupt/terminate notifications into shutdown. Every
// signal cancels a startup still in progress (even when a shutdown is already
// under way, e.g. the operator quit the TUI mid-startup). The first one also
// requests the ordered shutdown; if a shutdown is already in progress it
// hurries it at once instead. Each later one hurries it (cancels the shutdown
// context, which makes the agent stop escalate to a force stop) without ever
// skipping a step. It returns when the process is done.
func (a *App) watchSignals(sig <-chan os.Signal, cancelStart context.CancelFunc) {
	first := true
	for {
		select {
		case s := <-sig:
			cancelStart()
			if first && !a.shuttingDown() {
				first = false
				a.log.Info("signal received; shutting down", "signal", s.String())
				a.Shutdown("signal: " + s.String())
				continue
			}
			first = false
			a.log.Warn("signal received during shutdown; hurrying it", "signal", s.String())
			a.shut.hurry()
		case <-a.done:
			return
		}
	}
}

// shuttingDown reports whether a shutdown has been requested.
func (a *App) shuttingDown() bool {
	select {
	case <-a.shut.req:
		return true
	default:
		return false
	}
}

// adapterStopTimeout reports the agent's own ForceStop wait when it exposes one
// (the real Copilot adapter), and 0 for agents that do not (in-process test
// fakes), which keeps their short Force budgets valid.
func adapterStopTimeout(ag agent.Agent) time.Duration {
	if st, ok := ag.(interface{ StopTimeout() time.Duration }); ok {
		return st.StopTimeout()
	}
	return 0
}

// supervise is the dedicated shutdown goroutine: it waits for the first
// trigger and then runs the design §3.1 sequence through [hub.Shutdown] — never
// on the hub goroutine, the Bubble Tea goroutine or the signal watcher.
//
// The hub runs on a context nothing cancels before step 6, and the sequence's
// own context is independent of the signal that triggered it, so a SIGINT
// cannot stop the hub before step 0 has failed the pending permissions.
func (a *App) supervise() {
	defer a.finish()
	defer a.clearCreds() // before finish closes done, whatever path returns
	defer a.shut.hurry() // release the shutdown context

	<-a.shut.req
	a.log.Info("shutdown started", "reason", a.shut.reason)

	var errs []error
	sd, err := hub.NewShutdown(hub.ShutdownConfig{
		Hub:      a.hub,
		Agent:    a.agent,
		Gate:     a.remote, // step 0: revoke credentials, close the listener
		Reason:   a.shut.reason,
		Timeouts: a.shut.timeouts,
		// Always forwarded (never gated on the Force budget) so an explicit
		// Force shorter than the real adapter's stop timeout is rejected by
		// NewShutdown instead of silently skipping the validation.
		AdapterStopTimeout: adapterStopTimeout(a.agent),
		Logger:             a.log,
	})
	if err != nil {
		// Cannot happen with the components assembled above; if it does, still
		// unwind rather than strand the agent.
		errs = append(errs, fmt.Errorf("shutdown executor: %w", err))
		a.log.Error("cannot build the shutdown executor; forcing teardown", "error", err)
		tctx, cancel := context.WithTimeout(context.Background(), teardownBudget)
		errs = append(errs, a.teardown(tctx))
		cancel()
		a.shut.err = errors.Join(append(a.takeErrs(), errs...)...)
		return
	}
	sd.Start(a.shut.ctx)
	<-sd.Done() // steps 0-6; the hub is stopped
	if e := sd.Err(); e != nil {
		errs = append(errs, e) // each step failure was already logged by the executor
	}

	// Everything the hub could not know about: the HTTP server and its
	// connection goroutines, then the (already stopped) hub's run goroutine,
	// the accept loop and the local UI. A TUI that could not be stopped is
	// part of reap's error, so it lands in a.shut.err ahead of close(a.done).
	rctx, cancel := context.WithTimeout(context.WithoutCancel(a.shut.ctx), remoteShutdownBudget)
	if e := a.reap(rctx); e != nil {
		errs = append(errs, e)
	}
	cancel()
	a.shut.err = errors.Join(append(a.takeErrs(), errs...)...) // run failures (TUI, listener) first
	a.log.Info("shutdown complete", "ok", a.shut.err == nil)
}

// reap releases everything the hub's shutdown executor does not know about,
// once the agent is stopped: it drops the credentials (sticky, see
// clearCreds), shuts the remote transport down within rctx, stops and awaits
// the hub's run goroutine and the accept loop, and finally stops the local TUI.
// Every component is nil-guarded so it is safe on a partially assembled App.
// It returns the joined failures (remote shutdown, TUI stop) for the caller to
// record. The failed-start [App.teardown] cannot call it as one unit because it
// stops the UI and the listener before the agent; it composes the same steps.
func (a *App) reap(rctx context.Context) error {
	a.clearCreds()
	var errs []error
	if err := a.stopRemote(rctx); err != nil {
		errs = append(errs, err)
	}
	a.awaitHub()
	if err := a.stopUI(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// stopRemote shuts the remote transport down within ctx; nil when there is no
// remote.
func (a *App) stopRemote(ctx context.Context) error {
	if a.remote == nil {
		return nil
	}
	if err := a.remote.Shutdown(ctx); err != nil {
		a.log.Error("remote shutdown failed", "error", err)
		return fmt.Errorf("remote shutdown: %w", err)
	}
	return nil
}

// awaitHub stops the hub (idempotent), cancels its run context and waits for
// its run goroutine and the listener's accept loop to finish. The accept loop
// only returns once the remote transport has been shut down, so callers must
// run stopRemote first (reap and teardown both do).
func (a *App) awaitHub() {
	if a.hub != nil {
		a.hub.Stop()
		a.hubCancel()
		<-a.hubRunDone
	}
	if a.serveDone != nil {
		<-a.serveDone
	}
}

// stopUI stops the local TUI and returns its stop failure; nil when there is
// no TUI.
func (a *App) stopUI() error {
	if a.ui == nil {
		return nil
	}
	if err := a.ui.stop(); err != nil {
		a.log.Error("local tui stop failed", "error", err)
		return err
	}
	return nil
}
