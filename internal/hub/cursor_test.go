package hub

import "testing"

func shutdownCursorFrame(t *testing.T, kind Kind, gen string, seq uint64) Frame {
	t.Helper()
	f, err := newFrame(gen, seq, fixedTS(), kind, "s", nil)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestCursorApplyTreatsServerShutdownAsAlwaysApplicableAcrossGaps(t *testing.T) {
	hydrate := func(c *Cursor, seq uint64) {
		t.Helper()
		if a := c.Apply(shutdownCursorFrame(t, KindSessionHydrate, "g", seq)); a != ActionApply {
			t.Fatalf("hydrate: %v", a)
		}
	}
	shutdown := func(seq uint64) Frame { return shutdownCursorFrame(t, KindServerShutdown, "g", seq) }

	t.Run("contiguous", func(t *testing.T) {
		var c Cursor
		hydrate(&c, 10)
		if a := c.Apply(shutdown(11)); a != ActionApply {
			t.Fatalf("got %v", a)
		}
		if p, ok := c.Position(); !ok || p.Seq != 11 {
			t.Fatalf("position %+v %v", p, ok)
		}
	})
	t.Run("gap", func(t *testing.T) {
		var c Cursor
		hydrate(&c, 10)
		if a := c.Apply(shutdown(15)); a != ActionApply {
			t.Fatalf("got %v", a)
		}
		if _, ok := c.Position(); ok {
			t.Fatal("gapped shutdown must leave the cursor stale")
		}
	})
	t.Run("stale", func(t *testing.T) {
		var c Cursor
		hydrate(&c, 10)
		c.Apply(shutdownCursorFrame(t, KindAgentMessage, "g", 14)) // gap -> stale
		if a := c.Apply(shutdown(15)); a != ActionApply {
			t.Fatalf("got %v", a)
		}
		if _, ok := c.Position(); ok {
			t.Fatal("stale cursor must stay stale")
		}
	})
	t.Run("no baseline", func(t *testing.T) {
		var c Cursor
		if a := c.Apply(shutdown(3)); a != ActionApply {
			t.Fatalf("got %v", a)
		}
		if _, ok := c.Position(); ok {
			t.Fatal("no-baseline shutdown must not offer a resume position")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		var c Cursor
		hydrate(&c, 10)
		if a := c.Apply(shutdown(10)); a != ActionDiscard {
			t.Fatalf("got %v", a)
		}
		if a := c.Apply(shutdown(9)); a != ActionDiscard {
			t.Fatalf("got %v", a)
		}
	})
}
