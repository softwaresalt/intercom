package copilot

import (
	"context"
	"sync"
	"time"
)

// latch is the one-way "no new entrants" gate plus in-flight refcount that
// quiesces SDK callbacks (design §4.2). A bare WaitGroup is unsafe here
// because wg.Add(1) inside a callback races wg.Wait() in the shutdown path.
//
// Once quiesce has been called, enter always fails; quiesce returns when every
// callback that entered before the latch closed has called exit, or when its
// deadline passes.
type latch struct {
	mu     sync.Mutex
	closed bool
	n      int
	idle   chan struct{} // closed once closed && n == 0
	idleOK bool
}

func newLatch() *latch {
	return &latch{idle: make(chan struct{})}
}

// enter registers an in-flight callback. It reports false (and registers
// nothing) once the latch is closed.
func (l *latch) enter() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return false
	}
	l.n++
	return true
}

// exit deregisters a callback previously admitted by enter.
func (l *latch) exit() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.n--
	l.signalIdleLocked()
}

func (l *latch) signalIdleLocked() {
	if l.closed && l.n == 0 && !l.idleOK {
		l.idleOK = true
		close(l.idle)
	}
}

// quiesce closes the latch and waits up to timeout for in-flight callbacks to
// drain. It reports whether they drained. It is idempotent.
func (l *latch) quiesce(timeout time.Duration) bool {
	return l.quiesceCtx(context.Background(), timeout)
}

// quiesceCtx is quiesce that also gives up when ctx ends.
func (l *latch) quiesceCtx(ctx context.Context, timeout time.Duration) bool {
	l.mu.Lock()
	l.closed = true
	l.signalIdleLocked()
	l.mu.Unlock()

	// Check for an already-drained latch first so a zero or expired timeout
	// cannot randomly lose the race against a ready idle channel.
	select {
	case <-l.idle:
		return true
	default:
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-l.idle:
		return true
	case <-ctx.Done():
		return false
	case <-timer.C:
		return false
	}
}
