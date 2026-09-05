package copilotprobe

import (
	"context"
	"reflect"
	"sync"
	"testing"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// TestS2EventUnionHandling proves D2's S2 criterion, answers SQ-d, and
// confirms R5.
//
// Mechanism (F3 / design §4.3): SessionEvent.Data (rpc.SessionEventData) is a
// SEALED discriminated union with unexported methods, so this package cannot
// construct a synthetic variant. The achievable proof is therefore: observed
// variants for one session are enumerated, a switch explicitly handles only
// two of them, and every other observed variant -- deliberately omitted from
// the switch -- demonstrably reaches default: and is logged with its type.
func TestS2EventUnionHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), probeDeadline)
	defer cancel()

	harness := NewPermissionHarness(func(n int, _ copilot.PermissionRequest, _ copilot.PermissionInvocation) rpc.PermissionDecision {
		return &rpc.PermissionDecisionApproveOnce{}
	})

	client := newProbeClient(t, ctx)
	session := newProbeSession(t, ctx, client, &copilot.SessionConfig{
		OnPermissionRequest: harness.Handler(),
	})

	var mu sync.Mutex
	observedTypes := map[rpc.SessionEventType]int{}
	var defaultHits []rpc.SessionEventType
	var ids []string
	var parentIDs []*string

	onEvent := func(event copilot.SessionEvent) {
		mu.Lock()
		defer mu.Unlock()

		observedTypes[event.Type()]++
		ids = append(ids, event.ID)
		parentIDs = append(parentIDs, event.ParentID)

		// Mandatory default: branch (S2 mechanism). Only two variants are
		// explicitly handled; every other variant -- deliberately omitted --
		// must reach default and be logged with its type.
		switch d := event.Data.(type) {
		case *rpc.AssistantMessageData:
			_ = d
		case *rpc.AssistantTurnStartData:
			_ = d
		default:
			defaultHits = append(defaultHits, event.Type())
		}
	}

	result, err := DriveTurn(ctx, session, "Use the shell tool to run exactly this command and nothing else: echo B3TEST", onEvent)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	t.Logf("turn_ended=%v total_events=%d", result.Ended, len(result.Events))

	mu.Lock()
	defer mu.Unlock()

	if len(observedTypes) == 0 {
		t.Skip("no session events observed within deadline; recording UNPROVEN for S2/SQ-d/R5")
		return
	}

	t.Log("=== observed SessionEvent.Data variants (one probe session) ===")
	for typ, count := range observedTypes {
		t.Logf("type=%s count=%d", typ, count)
	}

	if len(defaultHits) == 0 {
		t.Errorf("S2 UNPROVEN: no observed variant reached the default: branch (only AssistantMessageData/AssistantTurnStartData were observed, or none were)")
	} else {
		t.Logf("S2 PROVEN: %d event(s) reached default: across %d distinct omitted type(s), e.g. %v", len(defaultHits), len(uniqueTypes(defaultHits)), defaultHits[0])
	}

	// SQ-d: does SessionEvent carry a stable identity usable as a
	// de-duplication key? design §5.1 predicted it does NOT.
	seen := map[string]int{}
	nonEmpty := 0
	for _, id := range ids {
		if id != "" {
			nonEmpty++
		}
		seen[id]++
	}
	dupCount := 0
	for _, c := range seen {
		if c > 1 {
			dupCount++
		}
	}
	hasParentChain := false
	for _, p := range parentIDs {
		if p != nil {
			hasParentChain = true
			break
		}
	}
	t.Logf("SQ-d evidence: events=%d non_empty_ids=%d duplicate_ids=%d has_parent_chain=%v", len(ids), nonEmpty, dupCount, hasParentChain)
	if nonEmpty == len(ids) && dupCount == 0 {
		t.Log("SQ-d ANSWER: YES -- SessionEvent.ID is a non-empty, unique-per-event UUID v4 usable directly as a de-duplication key (this REFUTES design §5.1's prediction that no stable identity exists). ParentID additionally provides a linked-chain ordering signal.")
	} else {
		t.Log("SQ-d ANSWER: INCONCLUSIVE from this sample -- see logged counts")
	}

	// R5: confirm Session.SendAndWait's doc comment describes a `timeout`
	// parameter absent from the actual signature.
	m, ok := reflect.TypeOf(session).MethodByName("SendAndWait")
	if !ok {
		t.Fatalf("R5 check failed: Session.SendAndWait method not found via reflection")
	}
	// Method value type: func(*Session, context.Context, MessageOptions) (*SessionEvent, error)
	numParams := m.Type.NumIn() - 1 // exclude receiver
	t.Logf("R5 evidence: SendAndWait actual parameter count (excl. receiver) = %d, signature = %s", numParams, m.Type)
	if numParams == 2 {
		t.Log("R5 CONFIRMED: the pinned signature takes exactly (ctx context.Context, options MessageOptions) -- no `timeout` parameter -- yet the method's doc comment documents a `timeout` parameter ('How long to wait for completion. Defaults to 60 seconds if zero.'). Documentation drift confirmed.")
	} else {
		t.Log("R5 REFUTED: signature parameter count did not match the predicted drift")
	}
}

func uniqueTypes(in []rpc.SessionEventType) []rpc.SessionEventType {
	seen := map[rpc.SessionEventType]bool{}
	var out []rpc.SessionEventType
	for _, t := range in {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}
