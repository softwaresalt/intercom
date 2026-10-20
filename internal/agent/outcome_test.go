package agent

import (
	"sync"
	"testing"
)

func TestOutcomeDeliverThenRevokeReportsTheDeliveredDecision(t *testing.T) {
	var o PermissionOutcome
	if _, ok := o.Delivered(); ok {
		t.Fatal("a fresh outcome reports a delivery")
	}
	want := PermissionDecision{Kind: DecisionAllowOnce}
	if !o.Deliver(want) {
		t.Fatal("Deliver on a fresh outcome must succeed")
	}
	if o.Revoke() {
		t.Fatal("Revoke after Deliver must fail: the SDK already got the decision")
	}
	if got, ok := o.Delivered(); !ok || got != want {
		t.Fatalf("Delivered = %+v, %v; want %+v, true", got, ok, want)
	}
}

func TestOutcomeRevokeThenDeliverRefusesTheHandler(t *testing.T) {
	var o PermissionOutcome
	if !o.Revoke() {
		t.Fatal("Revoke on a fresh outcome must succeed")
	}
	if o.Deliver(PermissionDecision{Kind: DecisionAllowOnce}) {
		t.Fatal("Deliver after Revoke must fail: an allow must never reach the SDK")
	}
	if _, ok := o.Delivered(); ok {
		t.Fatal("a revoked outcome must not report a delivery")
	}
	if !o.Revoke() {
		t.Fatal("Revoke is idempotent")
	}
}

func TestOutcomeFirstDeliveryWins(t *testing.T) {
	var o PermissionOutcome
	first := PermissionDecision{Kind: DecisionUnavailable}
	if !o.Deliver(first) || o.Deliver(PermissionDecision{Kind: DecisionAllowOnce}) {
		t.Fatal("only the first Deliver may succeed")
	}
	if got, _ := o.Delivered(); got != first {
		t.Fatalf("Delivered = %+v, want the first %+v", got, first)
	}
}

// Exactly one of Deliver and Revoke wins, whatever the interleaving.
func TestOutcomeDeliverAndRevokeAreMutuallyExclusive(t *testing.T) {
	for i := 0; i < 500; i++ {
		var o PermissionOutcome
		var delivered, revoked bool
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); delivered = o.Deliver(PermissionDecision{Kind: DecisionAllowOnce}) }()
		go func() { defer wg.Done(); revoked = o.Revoke() }()
		wg.Wait()
		if delivered == revoked {
			t.Fatalf("iteration %d: delivered=%v revoked=%v; exactly one must win", i, delivered, revoked)
		}
	}
}
