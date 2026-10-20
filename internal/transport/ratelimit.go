package transport

import (
	"errors"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// Transport-owned refusal reason codes (ADR MVP-3 addendum). The hub-owned
// codes (role_not_permitted, credential_revoked, ...) are hub.Reason values.
const (
	// ReasonCredentialExpired: the connection's credential is past its
	// lifetime.
	ReasonCredentialExpired = "credential_expired"
	// ReasonGenerationMismatch: the credential is bound to another process
	// generation.
	ReasonGenerationMismatch = "generation_mismatch"
	// ReasonRateLimited: the connection exceeded its inbound action rate.
	ReasonRateLimited = "rate_limited"
)

// Defaults for the per-connection inbound action limiter.
const (
	// DefaultActionRate is the sustained inbound actions per second.
	DefaultActionRate = 5.0
	// DefaultActionBurst is the bucket size.
	DefaultActionBurst = 10
	// DefaultMaxRateViolations is how many consecutive rate-limited actions
	// close the connection.
	DefaultMaxRateViolations = 20
)

// tokenBucket is a classic token bucket owned by one goroutine (the read
// pump), so it needs no lock.
type tokenBucket struct {
	rate   float64 // tokens per second
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

func newTokenBucket(rate float64, burst int, now func() time.Time) *tokenBucket {
	return &tokenBucket{rate: rate, burst: float64(burst), tokens: float64(burst), last: now(), now: now}
}

// allow takes one token if available. A clock that goes backwards never
// mints tokens.
func (b *tokenBucket) allow() bool {
	t := b.now()
	if el := t.Sub(b.last).Seconds(); el > 0 {
		b.tokens = min(b.burst, b.tokens+el*b.rate)
	}
	b.last = t
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// verdict is the outcome of per-action re-validation.
type verdict struct {
	// reason is empty when the action may proceed.
	reason string
	// closeConn says the connection must be closed (through the write pump)
	// after the refusal is sent.
	closeConn bool
}

// revalidate re-checks, before EVERY inbound action, everything connect-time
// authentication established (ADR MVP-3, design §6.1): the credential record
// still exists and is not revoked, has not expired by the store's own clock,
// is bound to the current gen, and its role is the one the hub registered the
// client with (the role captured at connect is never trusted on its own).
// Failures end the connection except shutdown revocation, which refuses
// without closing so the server.shutdown broadcast and drain can finish.
// Role/kind policy (steps 4 and 5) is the hub's authorize, whose refusal is
// surfaced to this client only.
func (h *Handler) revalidate(g *Grant, clientRole hub.Role) verdict {
	if g == nil {
		return verdict{reason: string(hub.ReasonCredentialRevoked), closeConn: true}
	}
	if err := h.cfg.Creds.Validate(g, h.cfg.Hub.Gen()); err != nil {
		switch {
		case errors.Is(err, ErrShuttingDown):
			return verdict{reason: string(hub.ReasonCredentialRevoked)}
		case errors.Is(err, ErrCredentialExpired):
			return verdict{reason: ReasonCredentialExpired, closeConn: true}
		case errors.Is(err, ErrGenerationMismatch):
			return verdict{reason: ReasonGenerationMismatch, closeConn: true}
		default:
			return verdict{reason: string(hub.ReasonCredentialRevoked), closeConn: true}
		}
	}
	if g.Role() != clientRole {
		return verdict{reason: string(hub.ReasonRoleNotPermitted), closeConn: true}
	}
	return verdict{}
}
