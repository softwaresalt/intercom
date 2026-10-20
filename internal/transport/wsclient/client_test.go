package wsclient

import (
	"testing"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

func env(kind hub.Kind, gen string, seq uint64) hub.Envelope {
	return hub.Envelope{V: 1, Gen: gen, Seq: seq, Kind: kind}
}

func TestCursorAppliesDiscardsAndDetectsGaps(t *testing.T) {
	var c cursor
	if v := c.apply(env(hub.KindAgentMessage, "g", 1)); v != VerdictRehydrate {
		t.Fatalf("no baseline must demand a hydrate, got %v", v)
	}
	if _, ok := c.position(); ok {
		t.Fatal("stale cursor offered a resume position")
	}
	if v := c.apply(env(hub.KindSessionHydrate, "g", 10)); v != VerdictApply {
		t.Fatal("hydrate must apply")
	}
	for _, tc := range []struct {
		seq  uint64
		want Verdict
	}{{11, VerdictApply}, {11, VerdictDiscard}, {10, VerdictDiscard}, {13, VerdictRehydrate}, {12, VerdictRehydrate}} {
		if v := c.apply(env(hub.KindAgentMessage, "g", tc.seq)); v != tc.want {
			t.Fatalf("seq %d: %v want %v", tc.seq, v, tc.want)
		}
	}
	if v := c.apply(env(hub.KindSessionHydrate, "g", 20)); v != VerdictApply {
		t.Fatal("a hydrate must recover a stale cursor")
	}
	if p, ok := c.position(); !ok || p.Seq != 20 {
		t.Fatalf("position = %+v %v", p, ok)
	}
	if v := c.apply(env(hub.KindAgentMessage, "other-gen", 21)); v != VerdictRehydrate {
		t.Fatal("a gen change must demand a hydrate")
	}
	if v := c.apply(env(hub.KindSessionHydrate, "other-gen", 3)); v != VerdictApply {
		t.Fatal("a hydrate from a new generation applies")
	}
}

func TestDialErrorNeverCarriesCredential(t *testing.T) {
	err := scrub(errString("dial failed for Bearer SECRETVALUE here"), "SECRETVALUE")
	if err.Error() == "" || contains(err.Error(), "SECRETVALUE") {
		t.Fatalf("scrub leaked: %v", err)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
