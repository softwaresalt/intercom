package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/tui"
)

// uiStopBudget bounds how long stopping the program waits for it to leave
// before it is killed outright.
const uiStopBudget = 3 * time.Second

// uiKillWaitBudget bounds how long stop waits for the program to leave after
// it was killed, before reporting the TUI as wedged.
const uiKillWaitBudget = uiStopBudget

// localUI is the single local TUI session: the Bubble Tea program, and the
// feed that connects it to the hub as its one in-process client (design §4).
type localUI struct {
	prog *tea.Program
	feed *tui.Feed
	done chan struct{} // closed when prog.Run has returned
	err  error         // prog.Run's error; valid after done

	stopping atomic.Bool // set by stop: a kill then is ours, not a failure
	stopOnce sync.Once

	watchDone chan struct{} // closed by the UI watcher once it recorded any runtime error; nil before it is registered
	stopErr   error         // stop's outcome, valid after stopOnce completed

	// Test seams: zero values select the program's own Quit/Kill and the
	// package budgets.
	quit, kill           func()
	stopBudget, killWait time.Duration
}

// readyMsg is the startup handshake: the program has entered its event loop.
type readyMsg struct{}

// failure is the program's error, excluding the kill our own stop issues. It
// is valid after done.
func (u *localUI) failure() error {
	select {
	case <-u.done:
	default:
		return nil // u.err is only valid once done is closed
	}
	if u.err == nil || (u.stopping.Load() && errors.Is(u.err, tea.ErrProgramKilled)) {
		return nil
	}
	return u.err
}

// requireTerminal fails closed when the TUI would draw to the process's own
// stdout and that is not a terminal: the credential panel must never be
// redirected to a file or pipe. An injected Output (tests) is not checked.
func requireTerminal(o TUIOptions) error {
	if o.Output != nil {
		return nil
	}
	isTerm := o.IsTerminal
	if isTerm == nil {
		isTerm = func(f *os.File) bool { return term.IsTerminal(f.Fd()) }
	}
	if !isTerm(os.Stdout) {
		return errors.New("stdout is not a terminal: the local TUI shows the credentials once and will not draw them to a file or pipe; run intercom in an interactive terminal")
	}
	return nil
}

// startUI builds the program, attaches it to the hub and runs it on its own
// goroutine. The model never touches the hub directly: actions go out through
// the feed's ingress and state comes back as hub envelopes.
//
// creds is the initial credential set. It is part of the model's construction,
// not a message sent after start, so it cannot be delivered after (and clobber)
// a Ctrl+R re-issue.
func startUI(startCtx context.Context, h *hub.Hub, o TUIOptions, creds tui.CredentialsMsg, reissue tui.Reissuer) (*localUI, error) {
	if err := startCtx.Err(); err != nil {
		return nil, err
	}
	ui := &localUI{done: make(chan struct{})}

	var prog *tea.Program // assigned before the feed starts, so the sink never sees nil
	ui.feed = tui.NewFeed(tui.FeedOptions{
		Dial: tui.HubDialer(h, 0),
		Sink: func(m tea.Msg) { prog.Send(m) },
	})
	ready := make(chan struct{})
	var readyOnce sync.Once
	opts := []tea.ProgramOption{
		tea.WithoutSignalHandler(), // signals are owned by the process
		tea.WithFilter(func(_ tea.Model, m tea.Msg) tea.Msg {
			if _, ok := m.(readyMsg); ok {
				readyOnce.Do(func() { close(ready) })
				return nil
			}
			return m
		}),
	}
	if o.Input != nil {
		opts = append(opts, tea.WithInput(o.Input))
	}
	if o.Output != nil {
		opts = append(opts, tea.WithOutput(o.Output))
	} else {
		opts = append(opts, tea.WithAltScreen())
	}
	prog = tea.NewProgram(tui.New(tui.Options{Ingress: ui.feed, Reissue: reissue, Credentials: creds}), opts...)
	ui.prog = prog
	ui.quit, ui.kill = prog.Quit, prog.Kill

	go func() {
		defer close(ui.done)
		_, ui.err = prog.Run()
	}()
	// Handshake: do not report ready after a failed terminal startup. Send
	// returns when the loop took the message or the program ended.
	go prog.Send(readyMsg{})
	select {
	case <-ready:
	case <-startCtx.Done():
		_ = ui.stop()
		return nil, startCtx.Err()
	case <-ui.done:
		if ui.err != nil {
			return nil, fmt.Errorf("terminal startup: %w", ui.err)
		}
		return nil, errors.New("terminal startup: the program exited immediately")
	}
	if err := startCtx.Err(); err != nil { // cancelled as the handshake completed
		_ = ui.stop()
		return nil, err
	}
	if o.Width > 0 && o.Height > 0 {
		prog.Send(tea.WindowSizeMsg{Width: o.Width, Height: o.Height})
	}
	// The feed outlives startup: its lifetime is its Start context, so it does
	// not follow startCtx; ui.stop releases it.
	if err := ui.feed.Start(context.WithoutCancel(startCtx)); err != nil {
		_ = ui.stop()
		return nil, err
	}
	if err := startCtx.Err(); err != nil { // cancelled while the feed was starting
		_ = ui.stop()
		return nil, err
	}
	return ui, nil
}

// stop quits the program, waits for it (killing it after a bound) and then
// releases the hub connection. It is idempotent and always returns the same
// result: non-nil when the program did not leave even after being killed.
//
// Quit and Kill are dispatched asynchronously because Bubble Tea's Quit blocks
// in Send until the event loop takes the message and Kill may block too; the
// wait after each is bounded by its budget regardless. Kill cancels the
// program's context, which also releases a parked Quit.
func (u *localUI) stop() error {
	u.stopOnce.Do(func() {
		u.stopping.Store(true)
		quit, kill := u.quit, u.kill
		if quit == nil {
			quit = u.prog.Quit
		}
		if kill == nil {
			kill = u.prog.Kill
		}
		stopBudget, killWait := u.stopBudget, u.killWait
		if stopBudget <= 0 {
			stopBudget = uiStopBudget
		}
		if killWait <= 0 {
			killWait = uiKillWaitBudget
		}

		go quit()
		if !waitClosed(u.done, stopBudget) {
			go kill()
			if !waitClosed(u.done, killWait) {
				u.stopErr = fmt.Errorf("local tui: the program did not exit within %s of Quit and %s of Kill", stopBudget, killWait)
				u.feed.Close()
				return
			}
		}
		u.feed.Close()
		// The watcher records a runtime error before it closes watchDone; it
		// is nil while startup cleanup runs ahead of the watcher's registration.
		if u.watchDone != nil {
			<-u.watchDone
		}
	})
	return u.stopErr
}

// waitClosed reports whether ch closed within d.
func waitClosed(ch <-chan struct{}, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ch:
		return true
	case <-t.C:
		return false
	}
}
