package copilotprobe

import (
	"context"
	"sync"

	copilot "github.com/github/copilot-sdk/go"
)

// TurnResult captures the outcome of one turn driven through DriveTurn.
type TurnResult struct {
	// Events holds every SessionEvent observed for the turn, in arrival order.
	Events []copilot.SessionEvent
	// Ended is true when an AssistantTurnEndData event was observed before
	// ctx was done. False means DriveTurn returned because ctx expired.
	Ended bool
}

// DriveTurn is the shared turn/streaming fixture consumed by B3 (event-union
// proving, SQ-d, R5) and B5 (SQ-b callback re-entrancy, SQ-c dispatch
// independence). It subscribes to session events, sends prompt, forwards
// every observed event to onEvent (the "instrumented event-callback hook"
// both subtasks require) as it arrives, and returns once an
// AssistantTurnEndData event is seen or ctx is done -- whichever first.
//
// onEvent may be nil. It MUST NOT block indefinitely: DriveTurn's own
// termination depends on the session's internal delivery, not on onEvent
// returning quickly, but a slow onEvent will still delay this turn's
// completion signal because callback delivery order is exactly what SQ-b
// probes.
func DriveTurn(ctx context.Context, session *copilot.Session, prompt string, onEvent func(copilot.SessionEvent)) (*TurnResult, error) {
	result := &TurnResult{}
	var mu sync.Mutex
	done := make(chan struct{})
	var closeOnce sync.Once

	unsubscribe := session.On(func(event copilot.SessionEvent) {
		mu.Lock()
		result.Events = append(result.Events, event)
		mu.Unlock()

		if onEvent != nil {
			onEvent(event)
		}

		if _, ok := event.Data.(*copilot.AssistantTurnEndData); ok {
			closeOnce.Do(func() { close(done) })
		}
	})
	defer unsubscribe()

	if _, err := session.Send(ctx, copilot.MessageOptions{Prompt: prompt}); err != nil {
		return result, err
	}

	select {
	case <-done:
		result.Ended = true
	case <-ctx.Done():
		result.Ended = false
	}

	mu.Lock()
	defer mu.Unlock()
	return result, nil
}
