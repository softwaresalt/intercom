package transport

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

const (
	credentialBytes = 32 // 256 bits (ADR MVP-2)
	redacted        = "[REDACTED]"
)

// Sentinel causes for credential failures. They are for logs and for the
// per-action re-validation decision; the HTTP layer maps every one of them to
// the same generic 401 so the response is not an oracle (ADR MVP-2). All are
// wrapped with apperr.KindUnauthorized by the store.
var (
	// ErrCredentialRejected is returned by Verify for every failure (wrong,
	// malformed, expired, revoked, wrong generation): the caller cannot tell
	// which.
	ErrCredentialRejected = errors.New("transport: credential rejected")
	// ErrCredentialExpired reports a credential past its fixed lifetime.
	ErrCredentialExpired = errors.New("transport: credential expired")
	// ErrCredentialRevoked reports a revoked credential (re-issue, explicit
	// revoke or shutdown).
	ErrCredentialRevoked = errors.New("transport: credential revoked")
	// ErrGenerationMismatch reports a credential minted for another process
	// generation.
	ErrGenerationMismatch = errors.New("transport: credential generation mismatch")
	// ErrShuttingDown is joined with ErrCredentialRevoked when the revocation
	// was the shutdown RevokeAll: the connection must be refused but not
	// closed, so the server.shutdown broadcast and drain can finish (ADR
	// MVP-3).
	ErrShuttingDown = errors.New("transport: shutting down")
)

// Token is a bearer credential. It redacts itself in every formatting verb,
// in JSON/text marshalling and in slog; the only way to read the plaintext is
// the explicit [Token.Reveal], used once to show it to the operator.
type Token struct{ s string }

// Reveal returns the plaintext credential. Call it only to display the value
// to the operator once; never log or store the result.
func (t Token) Reveal() string { return t.s }

// String returns "[REDACTED]".
func (Token) String() string { return redacted }

// GoString returns "[REDACTED]" so %#v cannot leak the value.
func (Token) GoString() string { return redacted }

// Format redacts the value for every fmt verb.
func (Token) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// MarshalText redacts the value.
func (Token) MarshalText() ([]byte, error) { return []byte(redacted), nil }

// MarshalJSON redacts the value.
func (Token) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(redacted)), nil }

// LogValue redacts the value for slog.
func (Token) LogValue() slog.Value { return slog.StringValue(redacted) }

// Issued is the result of a mint: the one-time plaintext plus public
// metadata. Only ID, Role, Fingerprint and ExpiresAt may be logged.
type Issued struct {
	// ID identifies the credential record (for Revoke).
	ID uint64
	// Role is the role the credential grants.
	Role hub.Role
	// Token is the plaintext, shown once.
	Token Token
	// Fingerprint is the first 8 hex characters of the SHA-256 digest, safe
	// to display and log to identify the live credential.
	Fingerprint string
	// ExpiresAt is the fixed expiry.
	ExpiresAt time.Time
}

// record is the stored form: a digest and metadata, never the plaintext.
// state fields are guarded by CredentialStore.mu.
type record struct {
	id        uint64
	role      hub.Role
	digest    [sha256.Size]byte
	gen       string
	expiresAt time.Time

	revoked  bool
	shutdown bool          // revoked by RevokeAll (no done signal)
	revCh    chan struct{} // closed on re-issue/Revoke
}

// CredentialConfig configures a [CredentialStore].
type CredentialConfig struct {
	// Gen is the process generation credentials are bound to (design §5.2).
	Gen string
	// Lifetime is the fixed lifetime from issue ([remote].token_lifetime_seconds).
	Lifetime time.Duration
	// Now is the clock; default time.Now.
	Now func() time.Time
	// Rand is the entropy source; default crypto/rand.Reader.
	Rand io.Reader
}

// LifetimeFromSeconds converts [remote].token_lifetime_seconds to a duration.
func LifetimeFromSeconds(s int) time.Duration { return time.Duration(s) * time.Second }

// CredentialStore mints and verifies per-role bearer credentials (ADR MVP-2).
// Credentials come from the process (crypto/rand), never from config.toml.
// It keeps only a SHA-256 digest and metadata per live credential, one per
// role. All state is guarded by one mutex.
type CredentialStore struct {
	cfg CredentialConfig

	mu       sync.Mutex
	live     map[hub.Role]*record
	nextID   uint64
	shutdown bool

	conns    atomic.Uint64
	compareN atomic.Int64
}

// NewCredentialStore validates cfg and returns an empty store.
func NewCredentialStore(cfg CredentialConfig) (*CredentialStore, error) {
	if cfg.Gen == "" {
		return nil, apperr.New(apperr.KindConfig, "transport: credential store needs a generation")
	}
	if cfg.Lifetime <= 0 {
		return nil, apperr.New(apperr.KindConfig, "transport: credential lifetime must be positive")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Rand == nil {
		cfg.Rand = rand.Reader
	}
	return &CredentialStore{cfg: cfg, live: make(map[hub.Role]*record, 2)}, nil
}

// String describes the store without any credential material.
func (s *CredentialStore) String() string { return "transport.CredentialStore{}" }

// GoString describes the store without any credential material.
func (s *CredentialStore) GoString() string { return s.String() }

func (s *CredentialStore) compareCount() int64 { return s.compareN.Load() }

func validRole(r hub.Role) bool { return r == hub.RoleObserver || r == hub.RoleApprover }

// Mint issues a fresh credential for role and revokes the role's previous
// one immediately (re-issue). A failed entropy read is a hard error
// (nothing is issued).
func (s *CredentialStore) Mint(role hub.Role) (Issued, error) {
	if !validRole(role) {
		return Issued{}, apperr.New(apperr.KindPolicy, "transport: credentials exist only for the observer and approver roles")
	}
	var raw [credentialBytes]byte
	if _, err := io.ReadFull(s.cfg.Rand, raw[:]); err != nil {
		return Issued{}, apperr.Wrapf(apperr.KindIO, err, "transport: read entropy for credential")
	}
	tok := base64.RawURLEncoding.EncodeToString(raw[:])
	clear(raw[:])
	digest := sha256.Sum256([]byte(tok))

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shutdown {
		return Issued{}, apperr.Wrap(apperr.KindPolicy, ErrShuttingDown)
	}
	if old := s.live[role]; old != nil {
		s.revokeLocked(old, false)
	}
	s.nextID++
	rec := &record{
		id: s.nextID, role: role, digest: digest, gen: s.cfg.Gen,
		expiresAt: s.cfg.Now().Add(s.cfg.Lifetime), revCh: make(chan struct{}),
	}
	s.live[role] = rec
	return Issued{
		ID: rec.id, Role: role, Token: Token{s: tok},
		Fingerprint: hex.EncodeToString(digest[:4]), ExpiresAt: rec.expiresAt,
	}, nil
}

func (s *CredentialStore) revokeLocked(r *record, shutdown bool) {
	if r.revoked {
		return
	}
	r.revoked = true
	r.shutdown = shutdown
	if !shutdown {
		close(r.revCh) // signals connections using it (ADR MVP-3: hub done signal)
	}
	if s.live[r.role] == r {
		delete(s.live, r.role)
	}
}

// Revoke revokes the credential with the given ID immediately and signals
// every connection authenticated with it. Unknown IDs are ignored.
func (s *CredentialStore) Revoke(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.live {
		if r.id == id {
			s.revokeLocked(r, false)
			return
		}
	}
}

// RevokeAll revokes every credential for shutdown (ADR MVP-2 §7): new
// admission and privileged actions are refused, but connections are not
// signalled, so the server.shutdown broadcast and bounded drain finish and the
// write pumps close the sockets. Minting is refused afterwards.
func (s *CredentialStore) RevokeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shutdown = true
	for _, r := range s.live {
		s.revokeLocked(r, true)
	}
}

// Grant is the credential record that authenticated one connection (ADR
// MVP-2 §7: "bound to each connection"). It holds no secret.
type Grant struct {
	rec     *record
	actorID string
}

// Role is the role the credential grants. The transport takes the role passed
// to the hub from here, never from a message.
func (g *Grant) Role() hub.Role { return g.rec.role }

// ActorID is the per-connection actor identifier (unique per connection).
func (g *Grant) ActorID() string { return g.actorID }

// ExpiresAt is the credential's fixed expiry.
func (g *Grant) ExpiresAt() time.Time { return g.rec.expiresAt }

// CredentialID is the ID of the credential record.
func (g *Grant) CredentialID() uint64 { return g.rec.id }

// Revoked is closed when the credential is revoked by re-issue or Revoke. It
// is NOT closed by shutdown revocation.
func (g *Grant) Revoked() <-chan struct{} { return g.rec.revCh }

// String describes the grant without credential material.
func (g *Grant) String() string { return "grant{" + g.actorID + "}" }

// GoString describes the grant without credential material.
func (g *Grant) GoString() string { return g.String() }

// Verify authenticates a presented credential against the live credentials
// and, on success, returns a new per-connection Grant. The presented value is
// digested and compared against EVERY candidate with
// crypto/subtle.ConstantTimeCompare over fixed-size digests, without early
// exit, so neither the presented length nor which candidate matched leaks via
// timing. Every failure is the same [ErrCredentialRejected] (wrapped
// KindUnauthorized); the error never contains the presented value.
func (s *CredentialStore) Verify(presented, currentGen string) (*Grant, error) {
	digest := sha256.Sum256([]byte(presented))

	s.mu.Lock()
	var match *record
	// Fixed candidate order over both roles; never break out early.
	for _, role := range [...]hub.Role{hub.RoleObserver, hub.RoleApprover} {
		cand := s.live[role]
		var want [sha256.Size]byte // zero digest compares as a miss for an absent role
		if cand != nil {
			want = cand.digest
		}
		eq := subtle.ConstantTimeCompare(digest[:], want[:])
		s.compareN.Add(1)
		if cand != nil && eq == 1 {
			match = cand
		}
	}
	var err error
	if match == nil {
		err = ErrCredentialRejected
	} else if verr := s.checkLocked(match, currentGen); verr != nil {
		err = ErrCredentialRejected
	}
	s.mu.Unlock()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindUnauthorized, err)
	}
	return &Grant{rec: match, actorID: "conn-" + strconv.FormatUint(s.conns.Add(1), 10)}, nil
}

// checkLocked is the shared revocation/expiry/gen check (re-validation steps
// 1 to 3 of ADR MVP-3).
func (s *CredentialStore) checkLocked(r *record, currentGen string) error {
	if r.revoked {
		if r.shutdown {
			return errors.Join(ErrCredentialRevoked, ErrShuttingDown)
		}
		return ErrCredentialRevoked
	}
	if !s.cfg.Now().Before(r.expiresAt) {
		return ErrCredentialExpired
	}
	if r.gen != currentGen {
		return ErrGenerationMismatch
	}
	return nil
}

// Validate re-checks a connection's credential: not revoked, not expired,
// bound to the current generation. It is called before EVERY privileged
// action and never trusts state captured at connect.
func (s *CredentialStore) Validate(g *Grant, currentGen string) error {
	if g == nil || g.rec == nil {
		return apperr.Wrap(apperr.KindUnauthorized, ErrCredentialRevoked)
	}
	s.mu.Lock()
	err := s.checkLocked(g.rec, currentGen)
	s.mu.Unlock()
	if err != nil {
		return apperr.Wrap(apperr.KindUnauthorized, err)
	}
	return nil
}
