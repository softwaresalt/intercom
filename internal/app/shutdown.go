package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// remoteShutdownBudget bounds the final transport teardown that follows the
// ordered sequence (every client was already closed at step 5, so this only
// reaps the connection goroutines and the HTTP server).
const remoteShutdownBudget = 10 * time.Second

// Shutdown requests the ordered shutdown of design §3.1. It is the single
// trigger behind a SIGINT/SIGTERM, a quit in the local TUI, and callers (the
// end-to-end test). It never blocks and is idempotent: only the first call
// (and its reason) counts. Wait on [App.Done] for completion.
func (a *App) Shutdown(reason string) {
	a.shutOnce.Do(func() {
		a.shutReason = reason
		close(a.shutReq)
	})
}

// Err reports the aggregated shutdown errors once [App.Done] is closed (nil
// before that, and nil for a clean shutdown). A non-nil Err means the process
// should exit non-zero: a step failed, the shutdown was interrupted, or the
// agent's CLI child may be orphaned (hub.ErrOrphanRisk).
func (a *App) Err() error {
	select {
	case <-a.done:
		return a.err
	default:
		return nil
	}
}

// watchSignals turns interrupt/terminate notifications into shutdown. The
// first one cancels a startup in progress and requests the ordered shutdown;
// each later one hurries it (cancels the shutdown context, which makes the
// agent stop escalate to a force stop) without ever skipping a step. It
// returns when the process is done.
func (a *App) watchSignals(sig <-chan os.Signal, cancelStart context.CancelFunc) {
	first := true
	for {
		select {
		case s := <-sig:
			if first {
				first = false
				a.log.Info("signal received; shutting down", "signal", s.String())
				cancelStart()
				a.Shutdown("signal: " + s.String())
				continue
			}
			a.log.Warn("another signal received; hurrying shutdown", "signal", s.String())
			a.hurry()
		case <-a.done:
			return
		}
	}
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
	defer a.hurry() // release the shutdown context

	<-a.shutReq
	a.log.Info("shutdown started", "reason", a.shutReason)

	var errs []error
	sd, err := hub.NewShutdown(hub.ShutdownConfig{
		Hub:      a.hub,
		Agent:    a.agent,
		Gate:     a.remote, // step 0: revoke credentials, close the listener
		Reason:   a.shutReason,
		Timeouts: a.timeouts,
		Logger:   a.log,
	})
	if err != nil {
		// Cannot happen with the components assembled above; if it does, still
		// unwind rather than strand the agent.
		errs = append(errs, fmt.Errorf("shutdown executor: %w", err))
		a.log.Error("cannot build the shutdown executor; forcing teardown", "error", err)
		tctx, cancel := context.WithTimeout(context.Background(), teardownBudget)
		errs = append(errs, a.teardown(tctx))
		cancel()
		a.err = errors.Join(errs...)
		return
	}
	sd.Start(a.shutCtx)
	<-sd.Done() // steps 0-6; the hub is stopped
	if e := sd.Err(); e != nil {
		errs = append(errs, e) // each step failure was already logged by the executor
	}

	// Everything the hub could not know about: the HTTP server and its
	// connection goroutines, then the (already stopped) hub's run goroutine,
	// the accept loop and the local UI.
	rctx, cancel := context.WithTimeout(context.WithoutCancel(a.shutCtx), remoteShutdownBudget)
	if e := a.remote.Shutdown(rctx); e != nil {
		errs = append(errs, fmt.Errorf("remote shutdown: %w", e))
		a.log.Error("remote shutdown failed", "error", e)
	}
	cancel()
	a.hubCancel()
	<-a.hubRunDone
	<-a.serveDone
	if a.ui != nil {
		a.ui.stop()
	}
	a.err = errors.Join(errs...)
	a.log.Info("shutdown complete", "ok", a.err == nil)
}
