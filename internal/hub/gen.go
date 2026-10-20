package hub

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// NewGen returns a UUIDv7 (RFC 9562) built from the 48-bit Unix millisecond
// timestamp of at plus 74 bits of crypto/rand entropy. It is the process
// generation identity (design §5.2): time-ordered, never a resettable
// counter, and unique even for two constructions inside the same
// millisecond.
func NewGen(at time.Time) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", apperr.Wrapf(apperr.KindIO, err, "hub: read entropy for generation id")
	}
	ms := uint64(at.UnixMilli()) // #nosec G115 -- pre-epoch clocks wrap harmlessly; entropy still disambiguates
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = b[6]&0x0f | 0x70 // version 7
	b[8] = b[8]&0x3f | 0x80 // RFC 4122 variant

	var out [36]byte
	hex.Encode(out[0:8], b[0:4])
	out[8] = '-'
	hex.Encode(out[9:13], b[4:6])
	out[13] = '-'
	hex.Encode(out[14:18], b[6:8])
	out[18] = '-'
	hex.Encode(out[19:23], b[8:10])
	out[23] = '-'
	hex.Encode(out[24:36], b[10:16])
	return string(out[:]), nil
}
