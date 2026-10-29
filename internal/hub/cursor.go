package hub

// Position is a (gen, seq) resume cursor (design §5.2): the pair, never seq
// alone, because a restarted process renumbers from zero.
type Position struct {
	Gen string
	Seq uint64
}

// Action is the client-side verdict for one received frame.
type Action int

const (
	// ActionApply: the frame is the next expected envelope; apply it.
	ActionApply Action = iota + 1
	// ActionDiscard: the frame was already applied (duplicate); ignore it.
	ActionDiscard
	// ActionRehydrate: a gap, out-of-order receipt, generation change or
	// missing baseline was detected; the client must reconnect for a full
	// hydrate (design §5.3).
	ActionRehydrate
)

func (a Action) String() string {
	switch a {
	case ActionApply:
		return "apply"
	case ActionDiscard:
		return "discard"
	case ActionRehydrate:
		return "rehydrate"
	default:
		return "invalid"
	}
}

// Cursor is the client-side idempotence and gap-detection helper (design
// §5.3). It keys on (gen, seq), discards anything already applied, and — the
// part "discard what you've already applied" alone gets wrong — detects gaps:
// receiving 105 after 100 yields ActionRehydrate rather than applying 105 and
// then silently discarding the late 101..104 forever.
//
// Consumer contract (design §4, §5.3): the TUI's channel is drop-oldest with a
// forced re-hydrate, so a consumer reading around the hub's drain may see a
// gap before the hydrate frame arrives. The cursor answers ActionRehydrate on
// the gap and then stays stale: every following non-hydrate frame also answers
// ActionRehydrate and MUST NOT be applied (the consumer discards it) until a
// session.hydrate frame arrives, which resets the cursor. A hydrate behind the
// current position within the same generation is itself rejected.
//
// A Cursor is owned by one consumer goroutine and is not safe for concurrent
// use. The zero value has no baseline.
type Cursor struct {
	pos   Position
	have  bool
	stale bool
}

// Apply classifies f and advances the cursor when f is the next envelope.
func (c *Cursor) Apply(f Frame) Action {
	if f.Kind() == KindSessionHydrate {
		if c.have && f.Gen() == c.pos.Gen && f.Seq() < c.pos.Seq {
			c.stale = true // a snapshot older than what was applied is bogus
			return ActionRehydrate
		}
		c.pos, c.have, c.stale = Position{Gen: f.Gen(), Seq: f.Seq()}, true, false
		return ActionApply
	}
	if f.Kind() == KindServerShutdown {
		return c.applyShutdown(f)
	}
	if !c.have || c.stale || f.Gen() != c.pos.Gen {
		c.stale = true
		return ActionRehydrate
	}
	switch {
	case f.Seq() <= c.pos.Seq:
		return ActionDiscard
	case f.Seq() == c.pos.Seq+1:
		c.pos.Seq = f.Seq()
		return ActionApply
	default:
		c.stale = true
		return ActionRehydrate
	}
}

// applyShutdown handles the terminal server.shutdown frame. It is always
// applicable, even across a sequence gap, because it is the last frame of the
// connection and dropping it would hide the shutdown from the consumer. A
// frame at or behind the current position within the same generation is a
// duplicate. Anything that is not the contiguous successor leaves the cursor
// stale so Position never offers a resume point that skipped frames.
func (c *Cursor) applyShutdown(f Frame) Action {
	same := c.have && f.Gen() == c.pos.Gen
	switch {
	case same && f.Seq() <= c.pos.Seq:
		return ActionDiscard
	case same && !c.stale && f.Seq() == c.pos.Seq+1:
		c.pos.Seq = f.Seq()
	default:
		c.pos, c.have, c.stale = Position{Gen: f.Gen(), Seq: f.Seq()}, true, true
	}
	return ActionApply
}

// Position returns the last-applied (gen, seq) for a reconnect, and false if
// the cursor has no baseline or requires a re-hydrate (a stale position must
// not be offered as a resume point).
func (c *Cursor) Position() (Position, bool) {
	if !c.have || c.stale {
		return Position{}, false
	}
	return c.pos, true
}

// NeedsRehydrate reports whether a gap or generation change was detected and
// no hydrate has been applied since.
func (c *Cursor) NeedsRehydrate() bool { return c.stale || !c.have }
