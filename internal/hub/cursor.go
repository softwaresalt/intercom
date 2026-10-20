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
// then silently discarding the late 101..104 forever. Once a re-hydrate is
// required the cursor stays in that state (every non-hydrate frame answers
// ActionRehydrate) until a session.hydrate frame resets it.
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
		c.pos, c.have, c.stale = Position{Gen: f.Gen(), Seq: f.Seq()}, true, false
		return ActionApply
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
