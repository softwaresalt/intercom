package copilot

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLatchEnterExit(t *testing.T) {
	l := newLatch()
	if !l.enter() {
		t.Fatal("enter should succeed before quiesce")
	}
	l.exit()
	if !l.quiesce(time.Second) {
		t.Fatal("quiesce with no in-flight callbacks should return true")
	}
}

func TestLatchRefusesNewEntrantsAfterQuiesce(t *testing.T) {
	l := newLatch()
	l.quiesce(time.Second)
	if l.enter() {
		t.Fatal("enter must fail after the latch closed")
	}
}

func TestLatchQuiesceWaitsForInFlight(t *testing.T) {
	l := newLatch()
	if !l.enter() {
		t.Fatal("enter")
	}
	var returned atomic.Bool
	done := make(chan bool, 1)
	go func() { r := l.quiesce(2 * time.Second); returned.Store(true); done <- r }()
	time.Sleep(50 * time.Millisecond)
	if returned.Load() {
		t.Fatal("quiesce returned while a callback was in flight")
	}
	l.exit()
	select {
	case ok := <-done:
		if !ok {
			t.Fatal("quiesce should report drained after the callback exited")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("quiesce did not return after exit")
	}
}

func TestLatchQuiesceHonoursDeadline(t *testing.T) {
	l := newLatch()
	l.enter() // never exits
	start := time.Now()
	if l.quiesce(50 * time.Millisecond) {
		t.Fatal("quiesce must report not-drained at deadline")
	}
	if time.Since(start) > time.Second {
		t.Fatal("quiesce exceeded its deadline")
	}
}

func TestLatchQuiesceIdempotent(t *testing.T) {
	l := newLatch()
	l.quiesce(time.Second)
	if !l.quiesce(time.Second) {
		t.Fatal("second quiesce should still report drained")
	}
}

// TestLatchCallbackRacingShutdown races many entrants against quiesce; under
// -race this proves no WaitGroup-style Add/Wait misuse and no panic.
func TestLatchCallbackRacingShutdown(t *testing.T) {
	for iter := 0; iter < 50; iter++ {
		l := newLatch()
		var wg sync.WaitGroup
		var inside atomic.Int64
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					if l.enter() {
						inside.Add(1)
						time.Sleep(time.Microsecond)
						inside.Add(-1)
						l.exit()
					}
				}
			}()
		}
		if !l.quiesce(2 * time.Second) {
			t.Fatal("quiesce failed to drain")
		}
		if n := inside.Load(); n != 0 {
			t.Fatalf("%d callbacks still inside after quiesce returned", n)
		}
		wg.Wait()
	}
}
