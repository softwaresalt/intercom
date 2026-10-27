package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport"
	"github.com/softwaresalt/intercom-go/internal/tui"
)

// mintAll issues the approver and observer credentials. Only the role, the
// fingerprint and the expiry are ever logged.
func (a *App) mintAll() error {
	_, err := a.mint()
	return err
}

// credentialRoles are the roles minted, in display order.
var credentialRoles = [...]hub.Role{hub.RoleApprover, hub.RoleObserver}

// mint issues a fresh credential for every role, each revoking that role's
// previous one, and publishes the live set after every role. It always returns
// the authoritative live set, with the error when a mint fails part-way: the
// roles already replaced keep their new credential (their old one is revoked)
// and the rest keep theirs, so neither Credentials nor a caller ever shows a
// revoked credential. The snapshot the set starts from is taken inside mintMu,
// so concurrent re-issues cannot start from a stale set. Only role,
// fingerprint and expiry are logged.
func (a *App) mint() ([]transport.Issued, error) {
	a.mintMu.Lock()
	defer a.mintMu.Unlock()
	if a.shuttingDown() {
		return nil, a.revokedForShutdown(nil)
	}
	next := a.Credentials() // under mintMu: the only writer of creds
	if len(next) != len(credentialRoles) {
		next = make([]transport.Issued, len(credentialRoles))
	}
	for i, role := range credentialRoles {
		iss, err := a.remote.Mint(role)
		if err != nil {
			if errors.Is(err, transport.ErrShuttingDown) || a.shuttingDown() {
				clear(next)
				return nil, a.revokedForShutdown(err)
			}
			err = fmt.Errorf("mint %s credential: %w", roleName(role), err)
			if i == 0 {
				return nil, err // nothing was replaced: there is no set to show
			}
			return next, err
		}
		a.log.Info("credential issued", "role", roleName(role), "fingerprint", iss.Fingerprint, "expires_at", iss.ExpiresAt)
		next[i] = iss
		if a.afterMint != nil {
			a.afterMint(role)
		}
		a.mu.Lock()
		if a.credsCleared || a.shuttingDown() {
			// Shutdown revoked (and cleared) the credentials while this mint was
			// in flight: never republish them.
			a.credsCleared = true
			for j := range a.creds {
				a.creds[j].Token = transport.Token{}
			}
			a.creds = nil
			a.mu.Unlock()
			clear(next)
			return nil, a.revokedForShutdown(nil)
		}
		a.creds = append([]transport.Issued(nil), next...)
		a.mu.Unlock()
	}
	return next, nil
}

// revokedForShutdown clears the credentials and builds the error a mint
// returns once shutdown revoked them; cause, when non-nil, is the mint failure.
func (a *App) revokedForShutdown(cause error) error {
	a.clearCreds()
	err := errors.Join(transport.ErrCredentialRevoked, transport.ErrShuttingDown)
	if cause != nil {
		err = errors.Join(err, cause)
	}
	return err
}

// clearCreds drops the live credentials and makes that terminal: no later mint
// can publish into creds again. Zeroing drops the process's own reference to
// the plaintext; Go strings are immutable, so this is not a memory scrub. It
// takes only mu, so it is safe to call with or without mintMu held.
func (a *App) clearCreds() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.credsCleared = true
	for i := range a.creds {
		a.creds[i].Token = transport.Token{}
	}
	a.creds = nil
}

// zeroPlaintext drops the process's own reference to the plaintext in the live
// credentials once the local TUI has been given it, keeping role, fingerprint
// and expiry. Go strings are immutable, so this is not a memory scrub; the TUI
// model's copy is dropped when the operator dismisses the panel.
func (a *App) zeroPlaintext() {
	a.mintMu.Lock()
	defer a.mintMu.Unlock()
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.creds {
		a.creds[i].Token = transport.Token{}
	}
}

// Reissue mints a new credential for every role and revokes the previous ones
// (re-issue; ADR MVP-2). Sockets authenticated by a revoked credential are
// closed by the transport. The plaintext is for the local TUI only. On a
// failure part-way it returns the authoritative live set together with the
// error (some roles may already have been replaced); when nothing was replaced
// or shutdown revoked the credentials it returns no set.
func (a *App) Reissue(context.Context) ([]transport.Issued, error) {
	live, err := a.mint()
	if err != nil {
		return live, err
	}
	a.log.Info("credentials re-issued", "approver", live[0].Fingerprint, "observer", live[1].Fingerprint)
	return live, nil
}

// tuiReissue is the TUI's re-issue key handler. When zeroAfterTUI is set it
// drops the process's own reference to the new plaintext once the message is
// built (not a memory scrub: Go strings are immutable); the TUI model's copy is
// dropped when the operator dismisses the panel.
func (a *App) tuiReissue(ctx context.Context) (tui.CredentialsMsg, error) {
	live, err := a.Reissue(ctx)
	if len(live) == 0 {
		return tui.CredentialsMsg{}, err
	}
	// On a partial failure some displayed credentials may already be revoked:
	// hand the panel the authoritative live set with the error.
	msg := a.credentialsMsg(live)
	if a.zeroAfterTUI {
		a.zeroPlaintext()
	}
	return msg, err
}

// credentialsMsg is the one place the plaintext leaves the minted records: it
// builds the local TUI's credential panel, which is delivered by
// tea.Program.Send and nowhere else. Entries without plaintext (already
// delivered and zeroed) are skipped.
func (a *App) credentialsMsg(issued []transport.Issued) tui.CredentialsMsg {
	msg := tui.CredentialsMsg{URL: a.WSURL(), Credentials: make([]tui.Credential, 0, len(issued))}
	for _, iss := range issued {
		secret := iss.Token.Reveal()
		if secret == "" {
			continue
		}
		msg.Credentials = append(msg.Credentials, tui.Credential{
			Role:        roleName(iss.Role),
			Secret:      secret,
			Fingerprint: iss.Fingerprint,
			ExpiresAt:   iss.ExpiresAt,
		})
	}
	return msg
}

func roleName(r hub.Role) string {
	if r == hub.RoleApprover {
		return "approver"
	}
	return "observer"
}

// Credentials returns the live minted credentials. They exist for the LOCAL
// operator only (the TUI, and the in-process end-to-end test): never log,
// persist or transmit the plaintext.
func (a *App) Credentials() []transport.Issued {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]transport.Issued(nil), a.creds...)
}
