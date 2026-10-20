package transport

import (
	"net/url"
	"strconv"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

const (
	// maxGenLen bounds the generation identifier in the resume cursor.
	maxGenLen = 64
	// QueryGen and QuerySeq are the URL query parameters carrying the
	// client's last-applied (gen, seq) cursor on (re)connect (design §5.3).
	// They are not secrets; the credential never travels in the URL.
	QueryGen = "gen"
	QuerySeq = "seq"
)

// parseCursor reads the optional resume cursor from the upgrade URL's query.
// Neither parameter means "no cursor" (full hydrate). Exactly one, a repeated
// parameter, an over-long gen, or a non-decimal seq is malformed (ok=false).
// The hub decides replay versus hydrate from the cursor (gen equal and the
// replay ring covers seq, else hydrate); the transport never second-guesses.
func parseCursor(q url.Values) (*hub.Position, bool) {
	gens, hasGen := q[QueryGen]
	seqs, hasSeq := q[QuerySeq]
	if !hasGen && !hasSeq {
		return nil, true
	}
	if len(gens) != 1 || len(seqs) != 1 || gens[0] == "" || len(gens[0]) > maxGenLen {
		return nil, false
	}
	seq, err := strconv.ParseUint(seqs[0], 10, 64)
	if err != nil {
		return nil, false
	}
	return &hub.Position{Gen: gens[0], Seq: seq}, true
}
