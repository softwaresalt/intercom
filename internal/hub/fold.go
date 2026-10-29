package hub

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// Declared caps on the session-state projection (design §5.3: "capped by
// entry count and serialized bytes, with the cap declared"). They are the
// defaults for [Limits]; hub [Config] exposes them as option fields.
const (
	// DefaultMaxEntries bounds the number of transcript entries retained.
	DefaultMaxEntries = 1000
	// DefaultMaxSnapshotBytes bounds the serialized size of all retained
	// entries.
	DefaultMaxSnapshotBytes = 256 << 10
	// DefaultMaxEntryTextBytes bounds the raw text kept for one entry.
	DefaultMaxEntryTextBytes = 16 << 10
	// SnapshotOverhead is the fixed allowance for the snapshot's non-entry
	// fields: a snapshot never exceeds MaxBytes+SnapshotOverhead bytes (the
	// byte cap counts each entry's separating comma).
	SnapshotOverhead = 1024

	// maxIdentBytes bounds ids and tool names stored in an entry.
	maxIdentBytes = 128
	// entryOverhead bounds the encoded size of an entry beyond its escaped text:
	// two identifiers at worst-case 6x JSON escaping plus keys and punctuation.
	entryOverhead = 2048
	// maxJSONExpansion is the worst-case byte expansion of JSON string escaping.
	maxJSONExpansion = 6
)

// Limits are the fold's declared caps. Zero fields select the defaults.
type Limits struct {
	// MaxEntries caps the retained entry count; oldest entries are evicted.
	MaxEntries int
	// MaxBytes caps the summed serialized size of retained entries.
	MaxBytes int
	// MaxEntryTextBytes caps the raw text of a single entry; longer text is
	// truncated and the entry flagged.
	MaxEntryTextBytes int
}

func (l Limits) withDefaults() Limits {
	if l.MaxEntries == 0 {
		l.MaxEntries = DefaultMaxEntries
	}
	if l.MaxBytes == 0 {
		l.MaxBytes = DefaultMaxSnapshotBytes
	}
	if l.MaxEntryTextBytes == 0 {
		l.MaxEntryTextBytes = DefaultMaxEntryTextBytes
	}
	return l
}

// entry is one transcript item in the projection.
type entry struct {
	Kind      string `json:"kind"` // message | tool | error
	ID        string `json:"id,omitempty"`
	Text      string `json:"text,omitempty"`
	Tool      string `json:"tool_name,omitempty"`
	Status    string `json:"status,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

// Entry statuses.
const (
	statusStreaming = "streaming"
	statusComplete  = "complete"
	statusRunning   = "running"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
)

type item struct {
	e   entry
	enc []byte // pre-marshalled entry, kept in step with e
}

// Fold is the bounded session-state projection built by folding normalized
// agent events (design §5.1). It is a cache for hydration, not the system of
// record. It is deterministic (no clock, no randomness) and is NOT safe for
// concurrent use: the hub goroutine is its sole writer, and everyone else
// receives immutable snapshot bytes.
type Fold struct {
	lim        Limits
	items      []item
	bytes      int // sum of itemCost over items
	evicted    int
	anyTrunc   bool
	turnActive bool
	turnID     string
	// lastDelta is what the most recent KindDelta contributed to the
	// projection, so the hub can publish exactly that on the wire.
	lastDelta deltaResult
}

// deltaResult is the text a delta added to its entry after the cap, and
// whether the cap clipped it.
type deltaResult struct {
	text      string
	truncated bool
}

// LastDelta reports what the most recent delta contributed (clipped text and
// whether it was clipped). Live envelopes carry exactly this, so a client that
// accumulates deltas holds the same bounded text and truncated flag as a
// hydrated one.
func (f *Fold) LastDelta() (text string, truncated bool) {
	return f.lastDelta.text, f.lastDelta.truncated
}

// NewFold validates limits and returns an empty projection.
func NewFold(l Limits) (*Fold, error) {
	if l.MaxEntries < 0 || l.MaxBytes < 0 || l.MaxEntryTextBytes < 0 {
		return nil, apperr.New(apperr.KindConfig, "hub: fold limits must not be negative")
	}
	l = l.withDefaults()
	if need := maxJSONExpansion*l.MaxEntryTextBytes + entryOverhead + 1; l.MaxBytes < need {
		return nil, apperr.Newf(apperr.KindConfig,
			"hub: MaxBytes %d cannot hold one entry of MaxEntryTextBytes %d (needs >= %d)", l.MaxBytes, l.MaxEntryTextBytes, need)
	}
	return &Fold{lim: l}, nil
}

// Len is the number of retained entries.
func (f *Fold) Len() int { return len(f.items) }

// EvictedCount is the number of entries evicted by the caps so far.
func (f *Fold) EvictedCount() int { return f.evicted }

// Apply folds one normalized agent event.
func (f *Fold) Apply(ev agent.Event) {
	switch ev.Kind {
	case agent.KindDelta:
		f.applyDelta(ev, "message")
	case agent.KindReasoningDelta:
		f.applyDelta(ev, "reasoning")
	case agent.KindAgentMessage:
		f.applyMessage(ev, "message")
	case agent.KindReasoning:
		f.applyMessage(ev, "reasoning")
	case agent.KindTurnStart:
		f.turnActive, f.turnID = true, f.ident(ev.TurnID)
	case agent.KindTurnEnd:
		f.turnActive, f.turnID = false, f.ident(ev.TurnID)
	case agent.KindIdle:
		f.turnActive = false
	case agent.KindToolStarted:
		id, t1 := f.identFlag(ev.ToolCallID)
		name, t2 := f.identFlag(ev.ToolName)
		f.push(entry{Kind: "tool", ID: id, Tool: name, Status: statusRunning, Truncated: t1 || t2 || f.flagTrunc(ev)})
	case agent.KindToolEnded:
		f.applyToolEnded(ev)
	default:
		ep := unknownEventError()
		f.ApplyError(ep.Code, ep.Message)
	}
}

// ApplyError folds a hub-produced error as a transcript entry, so a
// re-hydrated client still sees that the fault happened.
func (f *Fold) ApplyError(code, msg string) {
	e := entry{Kind: "error"}
	var idT bool
	e.ID, idT = f.identFlag(code)
	var txtT bool
	e.Text, txtT = f.capText(msg)
	e.Truncated = idT || txtT
	f.push(e)
}

// flagTrunc reports the adapter-set Truncated flag, recording it in the
// snapshot-level truncation marker.
func (f *Fold) flagTrunc(ev agent.Event) bool {
	if ev.Truncated {
		f.anyTrunc = true
	}
	return ev.Truncated
}

func (f *Fold) applyDelta(ev agent.Event, kind string) {
	id, idT := f.identFlag(ev.MessageID)
	evT := f.flagTrunc(ev)
	f.lastDelta = deltaResult{}
	if i := f.find(kind, id, statusStreaming); i >= 0 {
		cur := f.items[i].e
		// Cap BEFORE concatenating: only the remaining budget of the delta is
		// ever copied, so an oversize SDK delta is neither allocated nor kept.
		remaining := f.lim.MaxEntryTextBytes - len(cur.Text)
		add, clipped := ev.Text, false
		if len(add) > remaining {
			add, clipped = truncUTF8(ev.Text, remaining), true
		}
		add = strings.Clone(add) // "" + s would otherwise alias s
		f.lastDelta = deltaResult{text: add, truncated: clipped || evT}
		if clipped {
			f.anyTrunc = true
		}
		if add == "" && (cur.Truncated || !(clipped || evT)) {
			return
		}
		cur.Text += add // fresh allocation of at most MaxEntryTextBytes
		cur.Truncated = cur.Truncated || clipped || evT
		f.replace(i, cur)
		return
	}
	e := entry{Kind: kind, ID: id, Status: statusStreaming}
	var txtT bool
	e.Text, txtT = f.capText(ev.Text)
	e.Truncated = idT || txtT || evT
	f.lastDelta = deltaResult{text: e.Text, truncated: txtT || evT}
	f.push(e)
}

func (f *Fold) applyMessage(ev agent.Event, kind string) {
	id, idT := f.identFlag(ev.MessageID)
	e := entry{Kind: kind, ID: id, Status: statusComplete}
	var txtT bool
	e.Text, txtT = f.capText(ev.Text)
	e.Truncated = idT || txtT || f.flagTrunc(ev)
	if i := f.find(kind, id, statusStreaming); i >= 0 {
		f.replace(i, e)
		return
	}
	f.push(e)
}

func (f *Fold) applyToolEnded(ev agent.Event) {
	id, t1 := f.identFlag(ev.ToolCallID)
	status := statusFailed
	if ev.Success {
		status = statusSucceeded
	}
	name, t2 := f.identFlag(ev.ToolName)
	text, t3 := f.capText(ev.Text)
	t3 = t3 || f.flagTrunc(ev)
	if i := f.find("tool", id, statusRunning); i >= 0 {
		cur := f.items[i].e
		cur.Status = status
		cur.Text = text
		if cur.Tool == "" && name != "" {
			cur.Tool = name
		}
		cur.Truncated = cur.Truncated || t1 || t2 || t3
		f.replace(i, cur)
		return
	}
	f.push(entry{Kind: "tool", ID: id, Tool: name, Text: text, Status: status, Truncated: t1 || t2 || t3})
}

// capText applies the per-entry text cap and records truncation. The result is
// always a fresh copy, so a retained entry never pins a huge SDK string.
func (f *Fold) capText(s string) (string, bool) {
	if len(s) <= f.lim.MaxEntryTextBytes {
		return strings.Clone(s), false
	}
	f.anyTrunc = true
	return strings.Clone(truncUTF8(s, f.lim.MaxEntryTextBytes)), true
}

// identFlag bounds an identifier, returns an independent copy, and records
// truncation (matching the live mapping, which sets truncated for it too).
func (f *Fold) identFlag(s string) (string, bool) {
	if len(s) <= maxIdentBytes {
		return strings.Clone(s), false
	}
	f.anyTrunc = true
	return strings.Clone(truncUTF8(s, maxIdentBytes)), true
}

func (f *Fold) ident(s string) string {
	out, _ := f.identFlag(s)
	return out
}

// find scans from the newest entry backwards.
func (f *Fold) find(kind, id, status string) int {
	for i := len(f.items) - 1; i >= 0; i-- {
		e := &f.items[i].e
		if e.Kind == kind && e.ID == id && e.Status == status {
			return i
		}
	}
	return -1
}

// itemCost is an entry's share of the byte cap: its encoding plus the
// separating comma, so the serialized snapshot is bounded by
// MaxBytes+SnapshotOverhead no matter how many entries it holds.
func itemCost(enc []byte) int { return len(enc) + 1 }

func marshalEntry(e entry) []byte {
	b, err := json.Marshal(e)
	if err != nil { // unreachable: entry holds only strings and a bool
		return []byte(`{"kind":"error","id":"fold_marshal"}`)
	}
	return b
}

func (f *Fold) replace(i int, e entry) {
	enc := marshalEntry(e)
	f.bytes += itemCost(enc) - itemCost(f.items[i].enc)
	f.items[i] = item{e: e, enc: enc}
	f.evict()
}

func (f *Fold) push(e entry) {
	enc := marshalEntry(e)
	f.items = append(f.items, item{e: e, enc: enc})
	f.bytes += itemCost(enc)
	f.evict()
}

// evict drops the oldest entries until both caps hold.
func (f *Fold) evict() {
	drop := 0
	for len(f.items)-drop > f.lim.MaxEntries || (f.bytes > f.lim.MaxBytes && len(f.items)-drop > 1) {
		f.bytes -= itemCost(f.items[drop].enc)
		f.items[drop] = item{} // release the encoding
		drop++
	}
	if drop == 0 {
		return
	}
	f.evicted += drop
	f.items = f.items[drop:]
}

// Snapshot returns the projection as freshly allocated, pointer-free JSON
// bytes: the entries are already pre-marshalled, so this is a copy and not a
// marshal (design §5.3). The result never aliases fold storage.
func (f *Fold) Snapshot() []byte {
	b := make([]byte, 0, f.bytes+len(f.items)+SnapshotOverhead/4)
	b = append(b, `{"entries":[`...)
	for i := range f.items {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, f.items[i].enc...)
	}
	b = append(b, `],"turn_active":`...)
	b = strconv.AppendBool(b, f.turnActive)
	if f.turnID != "" {
		tid, _ := json.Marshal(f.turnID) // string marshal cannot fail
		b = append(b, `,"turn_id":`...)
		b = append(b, tid...)
	}
	b = append(b, `,"truncated":`...)
	b = strconv.AppendBool(b, f.anyTrunc || f.evicted > 0)
	b = append(b, `,"evicted_count":`...)
	b = strconv.AppendInt(b, int64(f.evicted), 10)
	b = append(b, '}')
	return b
}

// truncUTF8 shortens s to at most n bytes without splitting a rune.
func truncUTF8(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
