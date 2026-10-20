package app

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/tui"
)

// uiStopBudget bounds how long stopping the program waits for it to leave
// before it is killed outright.
const uiStopBudget = 3 * time.Second

// localUI is the single local TUI session: the Bubble Tea program, and the
// feed that connects it to the hub as its one in-process client (design §4).
type localUI struct {
	prog *tea.Program
	feed *tui.Feed
	done chan struct{} // closed when prog.Run has returned
	err  error         // prog.Run's error; valid after done

	stopOnce sync.Once
}

// startUI builds the program, attaches it to the hub and runs it on its own
// goroutine. The model never touches the hub directly: actions go out through
// the feed's ingress and state comes back as hub envelopes.
func startUI(ctx context.Context, h *hub.Hub, o TUIOptions) (*localUI, error) {
	ui := &localUI{done: make(chan struct{})}

	var prog *tea.Program // assigned before the feed starts, so the sink never sees nil
	ui.feed = tui.NewFeed(tui.FeedOptions{
		Dial: tui.HubDialer(h, 0),
		Sink: func(m tea.Msg) { prog.Send(m) },
	})
	opts := []tea.ProgramOption{tea.WithoutSignalHandler()} // signals are owned by the process
	if o.Input != nil {
		opts = append(opts, tea.WithInput(o.Input))
	}
	if o.Output != nil {
		opts = append(opts, tea.WithOutput(o.Output))
	} else {
		opts = append(opts, tea.WithAltScreen())
	}
	prog = tea.NewProgram(tui.New(tui.Options{Ingress: ui.feed}), opts...)
	ui.prog = prog

	go func() {
		defer close(ui.done)
		_, ui.err = prog.Run()
	}()
	if o.Width > 0 && o.Height > 0 {
		prog.Send(tea.WindowSizeMsg{Width: o.Width, Height: o.Height})
	}
	if err := ui.feed.Start(ctx); err != nil {
		ui.stop()
		return nil, err
	}
	return ui, nil
}

// stop quits the program, waits for it (killing it after a bound) and then
// releases the hub connection. It is idempotent.
func (u *localUI) stop() {
	u.stopOnce.Do(func() {
		u.prog.Quit()
		t := time.NewTimer(uiStopBudget)
		defer t.Stop()
		select {
		case <-u.done:
		case <-t.C:
			u.prog.Kill()
			<-u.done
		}
		u.feed.Close()
	})
}
