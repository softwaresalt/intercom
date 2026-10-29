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

func TestCursorTreatsServerShutdownAsAlwaysApplicableAcrossGaps(t *testing.T) {
	hydrated := func() *cursor {
		var c cursor
		if v := c.apply(env(hub.KindSessionHydrate, "g", 10)); v != VerdictApply {
			t.Fatal("hydrate must apply")
		}
		return &c
	}
	t.Run("contiguous", func(t *testing.T) {
		c := hydrated()
		if v := c.apply(env(hub.KindServerShutdown, "g", 11)); v != VerdictApply {
			t.Fatalf("got %v", v)
		}
		if p, ok := c.position(); !ok || p.Seq != 11 {
			t.Fatalf("position %+v %v", p, ok)
		}
	})
	t.Run("gap", func(t *testing.T) {
		c := hydrated()
		if v := c.apply(env(hub.KindServerShutdown, "g", 15)); v != VerdictApply {
			t.Fatalf("got %v", v)
		}
		if _, ok := c.position(); ok {
			t.Fatal("gapped shutdown must leave the cursor stale")
		}
	})
	t.Run("stale", func(t *testing.T) {
		c := hydrated()
		c.apply(env(hub.KindAgentMessage, "g", 14))
		if v := c.apply(env(hub.KindServerShutdown, "g", 15)); v != VerdictApply {
			t.Fatalf("got %v", v)
		}
		if _, ok := c.position(); ok {
			t.Fatal("stale cursor must stay stale")
		}
	})
	t.Run("no baseline", func(t *testing.T) {
		var c cursor
		if v := c.apply(env(hub.KindServerShutdown, "g", 3)); v != VerdictApply {
			t.Fatalf("got %v", v)
		}
		if _, ok := c.position(); ok {
			t.Fatal("no-baseline shutdown must not offer a resume position")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		c := hydrated()
		if v := c.apply(env(hub.KindServerShutdown, "g", 10)); v != VerdictDiscard {
			t.Fatalf("got %v", v)
		}
		if v := c.apply(env(hub.KindServerShutdown, "g", 9)); v != VerdictDiscard {
			t.Fatalf("got %v", v)
		}
	})
	t.Run("foreign gen and repeated gap", func(t *testing.T) {
		c := hydrated()
		if v := c.apply(env(hub.KindServerShutdown, "other", 2)); v != VerdictApply {
			t.Fatalf("foreign gen: %v", v)
		}
		if _, ok := c.position(); ok {
			t.Fatal("foreign-gen shutdown must leave the cursor stale")
		}
		if v := c.apply(env(hub.KindServerShutdown, "other", 2)); v != VerdictDiscard {
			t.Fatalf("repeat: %v", v)
		}
	})
}

func TestDialErrorNeverCarriesCredential(t *testing.T) {
	const secret = "SECRETVALUE"
	in := errString("dial failed: Bearer " + secret + " rejected by proxy")
	if !contains(in.Error(), secret) {
		t.Fatal("test input must actually embed the secret")
	}
	err := scrub(in, secret)
	if err.Error() == "" || contains(err.Error(), secret) {
		t.Fatalf("scrub leaked: %v", err)
	}
	// an error that does not embed the secret passes through unchanged
	clean := errString("connection refused")
	if got := scrub(clean, secret); got.Error() != clean.Error() {
		t.Fatalf("scrub altered a clean error: %v", got)
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
