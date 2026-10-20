package transport

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/apperr"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{t: time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)} }
func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

const testGen = "01900000-0000-7000-8000-000000000001"

func newTestStore(t *testing.T, clk *fakeClock) *CredentialStore {
	t.Helper()
	s, err := NewCredentialStore(CredentialConfig{Gen: testGen, Lifetime: time.Hour, Now: clk.Now})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustMint(t *testing.T, s *CredentialStore, role hub.Role) Issued {
	t.Helper()
	is, err := s.Mint(role)
	if err != nil {
		t.Fatal(err)
	}
	return is
}

func TestMintProduces256BitBase64URLCredentials(t *testing.T) {
	s := newTestStore(t, newFakeClock())
	obs := mustMint(t, s, hub.RoleObserver)
	app := mustMint(t, s, hub.RoleApprover)
	for _, is := range []Issued{obs, app} {
		raw, err := base64.RawURLEncoding.DecodeString(is.Token.Reveal())
		if err != nil || len(raw) != 32 {
			t.Fatalf("token must be unpadded base64url of 32 bytes, got len=%d err=%v", len(raw), err)
		}
	}
	if obs.Token.Reveal() == app.Token.Reveal() {
		t.Fatal("observer and approver credentials must be independent")
	}
	if _, err := s.Mint(hub.RoleNone); err == nil {
		t.Fatal("minting RoleNone must fail")
	}
	if _, err := s.Mint(hub.Role(9)); err == nil {
		t.Fatal("minting an invalid role must fail")
	}
}

func TestMintFailsClosedOnEntropyFailure(t *testing.T) {
	s, err := NewCredentialStore(CredentialConfig{Gen: testGen, Lifetime: time.Hour, Rand: failReader{}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Mint(hub.RoleApprover); err == nil {
		t.Fatal("entropy failure must be a hard error")
	}
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestNewCredentialStoreValidatesConfig(t *testing.T) {
	if _, err := NewCredentialStore(CredentialConfig{Gen: "", Lifetime: time.Hour}); err == nil {
		t.Error("empty gen must be refused")
	}
	if _, err := NewCredentialStore(CredentialConfig{Gen: testGen, Lifetime: 0}); err == nil {
		t.Error("zero lifetime must be refused")
	}
}

func TestVerifyAcceptsValidAndBindsRoleAndConnection(t *testing.T) {
	clk := newFakeClock()
	s := newTestStore(t, clk)
	obs := mustMint(t, s, hub.RoleObserver)
	app := mustMint(t, s, hub.RoleApprover)

	g1, err := s.Verify(obs.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if g1.Role() != hub.RoleObserver {
		t.Fatalf("role = %v", g1.Role())
	}
	g2, err := s.Verify(app.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if g2.Role() != hub.RoleApprover {
		t.Fatalf("role = %v", g2.Role())
	}
	g3, err := s.Verify(app.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if g2.ActorID() == g3.ActorID() || g2.ActorID() == "" {
		t.Fatalf("each connection needs a distinct actor ID, got %q and %q", g2.ActorID(), g3.ActorID())
	}
	if want := clk.Now().Add(time.Hour); !g2.ExpiresAt().Equal(want) {
		t.Fatalf("expiry = %v want %v", g2.ExpiresAt(), want)
	}
	if err := s.Validate(g2, testGen); err != nil {
		t.Fatalf("Validate fresh grant: %v", err)
	}
}

func TestVerifyRejectsWrongExpiredRevokedAndWrongGen(t *testing.T) {
	clk := newFakeClock()
	s := newTestStore(t, clk)
	app := mustMint(t, s, hub.RoleApprover)
	obs := mustMint(t, s, hub.RoleObserver)

	for _, bad := range []string{"", "x", strings.Repeat("A", 43), app.Token.Reveal() + "A", app.Token.Reveal()[1:], strings.Repeat("x", 100000)} {
		if _, err := s.Verify(bad, testGen); !errors.Is(err, ErrCredentialRejected) {
			t.Errorf("wrong credential accepted or wrong error: %v", err)
		}
	}
	// wrong generation
	if _, err := s.Verify(app.Token.Reveal(), "other-gen"); !errors.Is(err, ErrCredentialRejected) {
		t.Errorf("wrong gen: %v", err)
	}
	// revoked
	s.Revoke(obs.ID)
	if _, err := s.Verify(obs.Token.Reveal(), testGen); !errors.Is(err, ErrCredentialRejected) {
		t.Errorf("revoked: %v", err)
	}
	// expired
	clk.Advance(time.Hour)
	if _, err := s.Verify(app.Token.Reveal(), testGen); !errors.Is(err, ErrCredentialRejected) {
		t.Errorf("expired at the boundary: %v", err)
	}
	// every failure is KindUnauthorized
	_, err := s.Verify("nope", testGen)
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Kind() != apperr.KindUnauthorized {
		t.Errorf("error kind = %v", err)
	}
}

func TestValidateReportsSpecificReasons(t *testing.T) {
	clk := newFakeClock()
	s := newTestStore(t, clk)
	app := mustMint(t, s, hub.RoleApprover)
	g, err := s.Verify(app.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(g, "other"); !errors.Is(err, ErrGenerationMismatch) {
		t.Errorf("gen mismatch: %v", err)
	}
	clk.Advance(time.Hour + time.Second)
	if err := s.Validate(g, testGen); !errors.Is(err, ErrCredentialExpired) {
		t.Errorf("expired: %v", err)
	}
	clk.t = clk.t.Add(-2 * time.Hour)
	s.Revoke(app.ID)
	if err := s.Validate(g, testGen); !errors.Is(err, ErrCredentialRevoked) {
		t.Errorf("revoked: %v", err)
	}
	select {
	case <-g.Revoked():
	default:
		t.Error("Revoke must signal the grant's Revoked channel")
	}
}

func TestReissueRevokesPreviousCredential(t *testing.T) {
	s := newTestStore(t, newFakeClock())
	old := mustMint(t, s, hub.RoleApprover)
	g, err := s.Verify(old.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	fresh := mustMint(t, s, hub.RoleApprover)
	if _, err := s.Verify(old.Token.Reveal(), testGen); err == nil {
		t.Fatal("old credential must be rejected after re-issue")
	}
	if err := s.Validate(g, testGen); !errors.Is(err, ErrCredentialRevoked) {
		t.Fatalf("existing grant must be revoked on re-issue: %v", err)
	}
	select {
	case <-g.Revoked():
	default:
		t.Fatal("re-issue must signal Revoked for connections using the old credential")
	}
	if _, err := s.Verify(fresh.Token.Reveal(), testGen); err != nil {
		t.Fatalf("new credential must verify: %v", err)
	}
	// the other role is unaffected
	obs := mustMint(t, s, hub.RoleObserver)
	mustMint(t, s, hub.RoleApprover)
	if _, err := s.Verify(obs.Token.Reveal(), testGen); err != nil {
		t.Fatalf("observer must survive approver re-issue: %v", err)
	}
}

func TestRevokeAllRefusesWithoutSignallingDone(t *testing.T) {
	s := newTestStore(t, newFakeClock())
	app := mustMint(t, s, hub.RoleApprover)
	g, _ := s.Verify(app.Token.Reveal(), testGen)
	s.RevokeAll()
	if _, err := s.Verify(app.Token.Reveal(), testGen); err == nil {
		t.Fatal("shutdown revocation must block new admission")
	}
	err := s.Validate(g, testGen)
	if !errors.Is(err, ErrCredentialRevoked) || !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("Validate after RevokeAll = %v; want revoked+shutting down", err)
	}
	select {
	case <-g.Revoked():
		t.Fatal("shutdown revocation must NOT signal done (ADR MVP-2 §7)")
	default:
	}
	if _, err := s.Mint(hub.RoleApprover); err == nil {
		t.Fatal("no minting after shutdown revocation")
	}
}

func TestVerifyScansAllCandidatesWithoutEarlyExit(t *testing.T) {
	s := newTestStore(t, newFakeClock())
	mustMint(t, s, hub.RoleObserver)
	app := mustMint(t, s, hub.RoleApprover)
	before := s.compareCount()
	if _, err := s.Verify(app.Token.Reveal(), testGen); err != nil {
		t.Fatal(err)
	}
	// A match on the first or the last candidate performs the same number of
	// comparisons as a miss.
	matched := s.compareCount() - before
	before = s.compareCount()
	_, _ = s.Verify("definitely-wrong", testGen)
	missed := s.compareCount() - before
	if matched != 2 || missed != 2 {
		t.Fatalf("comparisons matched=%d missed=%d; want 2 each (no early exit)", matched, missed)
	}
}

func TestTokenIsRedactedEverywhere(t *testing.T) {
	s := newTestStore(t, newFakeClock())
	is := mustMint(t, s, hub.RoleApprover)
	secret := is.Token.Reveal()

	outs := []string{
		fmt.Sprint(is.Token), fmt.Sprintf("%v", is.Token), fmt.Sprintf("%+v", is.Token),
		fmt.Sprintf("%#v", is.Token), fmt.Sprintf("%s", is.Token), fmt.Sprintf("%q", is.Token), //nolint:gosimple
		fmt.Sprintf("%x", is.Token), fmt.Sprintf("%v", is), fmt.Sprintf("%+v", &is), fmt.Sprintf("%#v", is),
		is.Token.String(), fmt.Sprintf("%d", is.Token),
	}
	if b, err := json.Marshal(is); err == nil {
		outs = append(outs, string(b))
	}
	if b, err := is.Token.MarshalText(); err == nil {
		outs = append(outs, string(b))
	}
	var buf bytes.Buffer
	slog.New(slog.NewTextHandler(&buf, nil)).Info("x", "tok", is.Token, "issued", is)
	outs = append(outs, buf.String())
	g, _ := s.Verify(secret, testGen)
	outs = append(outs, fmt.Sprintf("%v %+v %#v", g, g, g), fmt.Sprintf("%v %+v", s, s))
	_, verr := s.Verify(secret+"x", testGen)
	outs = append(outs, fmt.Sprint(verr))
	_, verr = s.Verify(secret, "wrong-gen")
	outs = append(outs, fmt.Sprint(verr))
	for i, o := range outs {
		if strings.Contains(o, secret) {
			t.Fatalf("output %d leaks the credential: %q", i, o)
		}
	}
	if is.Token.String() != "[REDACTED]" {
		t.Fatalf("String() = %q", is.Token.String())
	}
	if len(is.Fingerprint) != 8 {
		t.Fatalf("fingerprint = %q; want 8 hex chars", is.Fingerprint)
	}
}

func TestLifetimeFromConfig(t *testing.T) {
	if got := LifetimeFromSeconds(14400); got != 4*time.Hour {
		t.Fatalf("got %v", got)
	}
}
