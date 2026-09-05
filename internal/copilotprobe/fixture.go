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

// DriveTurn is the shared turn/streaming fixture. Actual consumers in this
// package: permission_test.go (B2, S1 permission round-trip) and
// events_test.go (B3, S2/SQ-d/R5). concurrency_test.go (B5, SQ-b/SQ-c)
// intentionally does NOT reuse this fixture: SQ-b/SQ-c need to observe
// events while a permission handler remains genuinely blocked mid-turn, a
// different termination signal than "wait for AssistantTurnEndData", so B5
// re-implements an equivalent subscribe/send loop inline with its own
// termination condition.
//
// It subscribes to session events, sends prompt, forwards every observed
// event to onEvent (the "instrumented event-callback hook" B3 requires) as
// it arrives, and returns once an AssistantTurnEndData event is seen or ctx
// is done -- whichever first.
//
// onEvent may be nil. It MUST NOT block indefinitely: DriveTurn's own
// termination depends on the session's internal delivery, not on onEvent
// returning quickly, but a slow onEvent will still delay this turn's
// completion signal because callback delivery order is exactly what SQ-b
// probes.
//
// Concurrency/synchronization note: the returned TurnResult.Events is a
// snapshot copy taken under this function's internal lock immediately
// before return, specifically so callers (who have no access to that lock)
// never read the slice concurrently with an in-flight append. This package
// does not have an independently-verified guarantee from the SDK that no
// further callback invocation can occur once session.On's returned
// unsubscribe function has been called (that guarantee is itself unverified
// SDK behaviour, outside this spike's proven criteria) -- the snapshot copy
// bounds the exposure to "events observed up to the moment of copy", not
// "no more events will ever be appended to the original backing array",
// which is why callers must use the returned copy, never assume the
// snapshot is a live view.
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
	snapshot := make([]copilot.SessionEvent, len(result.Events))
	copy(snapshot, result.Events)
	return &TurnResult{Events: snapshot, Ended: result.Ended}, nil
}
