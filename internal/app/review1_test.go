package app

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

// gatedFailReader is a TUI input whose Read blocks until released, then fails.
type gatedFailReader struct {
	release chan struct{}
	err     error
}

func (r *gatedFailReader) Read([]byte) (int, error) {
	<-r.release
	return 0, r.err
}

// TestTUIFailureExitsNonZero: a TUI whose Run fails after startup (here: its
// input errors) shuts the process down and the failure is in Err, so the
// process exits non-zero.
func TestTUIFailureExitsNonZero(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	boom := errors.New("tty input lost")
	in := &gatedFailReader{release: make(chan struct{}), err: boom}
	var once sync.Once
	releaseIn := func() { once.Do(func() { close(in.release) }) }
	t.Cleanup(releaseIn)
	out := &syncBuf{}
	o.TUI = TUIOptions{Input: in, Output: out, Width: 80, Height: 24}
	a := startApp(t, o)
	waitOutput(t, out, "shown once") // the TUI is up

	releaseIn()
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("a failed TUI did not shut the process down")
	}
	if err := a.Err(); !errors.Is(err, boom) {
		t.Fatalf("Err() = %v; want it to wrap the TUI failure (non-zero exit)", err)
	}
}

// TestCleanTUIQuitIsNotAnError: our own stop() (Quit/Kill) is not a failure.
func TestCleanTUIQuitIsNotAnError(t *testing.T) {
	r := newUIRig(t)
	r.app.Shutdown("test")
	select {
	case <-r.app.Done():
	case <-time.After(waitFor):
		t.Fatal("no shutdown")
	}
	if err := r.app.Err(); err != nil {
		t.Fatalf("Err() = %v after a clean quit; want nil", err)
	}
}

// TestTUIRefusesANonTerminalStdout: with the default output (the process's
// stdout) the credential panel must not be drawn to a file or pipe, so startup
// fails closed, before anything is started.
func TestTUIRefusesANonTerminalStdout(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	o.TUI = TUIOptions{IsTerminal: func(*os.File) bool { return false }}
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	a, err := Start(ctx, o)
	if err == nil {
		stopApp(t, a)
		t.Fatal("Start succeeded with a non-terminal stdout; credentials could be redirected to a file")
	}
	if f.invoked() != 0 {
		t.Error("the agent was built before the terminal check")
	}
}

// TestInjectedTUIOutputNeedsNoTerminal: an injected Output (tests) is not
// subject to the terminal check.
func TestInjectedTUIOutputNeedsNoTerminal(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	in, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	o.TUI = TUIOptions{Input: in, Output: &syncBuf{}, IsTerminal: func(*os.File) bool { return false }}
	startApp(t, o)
}

// budgetRand is an entropy source that satisfies a fixed number of Read calls
// and then fails. Each Mint takes exactly one entropy read, so calls counts
// mints; the test is thereby independent of the credential's byte length. It
// fills every read with a rising byte sequence and records each one, so a test
// can derive the credential a given mint produced via token.
type budgetRand struct {
	mu    sync.Mutex
	calls int
	next  byte
	reads [][]byte
}

func (r *budgetRand) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.calls <= 0 {
		return 0, errors.New("entropy exhausted")
	}
	r.calls--
	got := make([]byte, len(p))
	for i := range p {
		r.next++
		p[i] = r.next
		got[i] = r.next
	}
	r.reads = append(r.reads, got)
	return len(p), nil
}

// token is the credential that mint n (0-based) produced: the encoding of the
// n-th entropy read.
func (r *budgetRand) token(t *testing.T, n int) string {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if n >= len(r.reads) {
		t.Fatalf("entropy read %d never happened (%d reads)", n, len(r.reads))
	}
	return base64.RawURLEncoding.EncodeToString(r.reads[n])
}

func partialRig(t *testing.T) (*App, []transport.Issued) {
	t.Helper()
	f := newFactory()
	o, _ := baseOptions(t, f)
	// one entropy read per mint: every role at start, then one more (the
	// re-issue's approver) so the re-issue's observer fails.
	o.CredentialRand = &budgetRand{calls: len(credentialRoles) + 1}
	a := startApp(t, o)
	return a, a.Credentials()
}

func dialToken(a *App, tok string) error {
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: a.WSURL(), Credential: tok})
	if err == nil {
		c.Abort()
	}
	return err
}

// TestPartialReissueReportsTheAuthoritativeLiveSet: when the second role's
// mint fails, the approver was already replaced (its old credential revoked).
// Reissue returns the live set with the error, Credentials() agrees, and the
// old approver no longer authenticates while the untouched observer does.
func TestPartialReissueReportsTheAuthoritativeLiveSet(t *testing.T) {
	a, old := partialRig(t)
	live, err := a.Reissue(context.Background())
	if err == nil {
		t.Fatal("Reissue succeeded although the observer mint had no entropy")
	}
	if len(live) != 2 {
		t.Fatalf("Reissue returned %d credentials with the error; want the live set (2)", len(live))
	}
	if live[0].Token.Reveal() == old[0].Token.Reveal() {
		t.Error("the approver was not replaced")
	}
	if live[1].Token.Reveal() != old[1].Token.Reveal() {
		t.Error("the observer changed although its mint failed")
	}
	got := a.Credentials()
	for i := range live {
		if got[i].Token.Reveal() != live[i].Token.Reveal() {
			t.Errorf("Credentials()[%d] differs from the set Reissue reported", i)
		}
	}
	if dialToken(a, old[0].Token.Reveal()) == nil {
		t.Error("the revoked approver credential still authenticates")
	}
	if err := dialToken(a, live[0].Token.Reveal()); err != nil {
		t.Errorf("the live approver credential is refused: %v", err)
	}
	if err := dialToken(a, old[1].Token.Reveal()); err != nil {
		t.Errorf("the still-live observer credential is refused: %v", err)
	}
}

// TestTUIReissueFailureCarriesTheLiveSet: the TUI handler hands the live set
// to the panel on a partial failure, so it never shows the revoked approver.
func TestTUIReissueFailureCarriesTheLiveSet(t *testing.T) {
	a, old := partialRig(t)
	msg, err := a.tuiReissue(context.Background())
	if err == nil {
		t.Fatal("tuiReissue succeeded although the observer mint had no entropy")
	}
	if len(msg.Credentials) != 2 {
		t.Fatalf("failure message carries %d credentials; want the live set (2)", len(msg.Credentials))
	}
	if msg.Credentials[0].Secret == old[0].Token.Reveal() {
		t.Error("the panel would keep showing the revoked approver credential")
	}
	if msg.Credentials[0].Secret != a.Credentials()[0].Token.Reveal() {
		t.Error("the panel's approver is not the live approver")
	}
}

// TestConcurrentReissueLeavesExactlyOneLiveSet runs under -race: re-issues
// serialize, every returned set is internally consistent, and only the final
// set authenticates.
func TestConcurrentReissueLeavesExactlyOneLiveSet(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	a := startApp(t, o)

	const n = 6
	sets := make([][]transport.Issued, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := a.Reissue(context.Background())
			if err != nil {
				t.Errorf("Reissue: %v", err)
				return
			}
			sets[i] = s
		}()
	}
	wg.Wait()
	if t.Failed() {
		return
	}
	final := a.Credentials()
	matches := 0
	for _, s := range sets {
		if len(s) != 2 {
			t.Fatalf("a Reissue returned %d credentials", len(s))
		}
		if s[0].Token.Reveal() == final[0].Token.Reveal() && s[1].Token.Reveal() == final[1].Token.Reveal() {
			matches++
			continue
		}
		for _, c := range s {
			if dialToken(a, c.Token.Reveal()) == nil {
				t.Errorf("a superseded %v credential still authenticates", c.Role)
			}
		}
	}
	if matches != 1 {
		t.Fatalf("%d returned sets match the live set; want exactly 1", matches)
	}
	for _, c := range final {
		if err := dialToken(a, c.Token.Reveal()); err != nil {
			t.Errorf("the live %v credential is refused: %v", c.Role, err)
		}
	}
}

// TestSignalDuringAnInProgressShutdownHurries: a shutdown already requested
// (here by a TUI quit) is hurried by the very first signal, not the second.
func TestSignalDuringAnInProgressShutdownHurries(t *testing.T) {
	f := newFactory()
	f.stopBlock = true
	o, _ := baseOptions(t, f)
	sig := make(chan os.Signal, 4)
	o.Signals = sig
	o.Timeouts = hub.ShutdownTimeouts{Stop: time.Minute, Force: time.Minute}
	a := startApp(t, o)

	a.Shutdown("tui quit")
	f.rec.waitEvent(t, "agent.stop") // stuck inside Stop
	sig <- os.Interrupt
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("a signal during an in-progress shutdown did not hurry it")
	}
}

// TestServeFailureTriggersShutdownAndIsReported: a listener that dies with
// anything but ErrServerClosed shuts the process down non-zero.
func TestServeFailureTriggersShutdownAndIsReported(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	a := startApp(t, o)

	a.onServeExit(nil)
	a.onServeExit(http.ErrServerClosed) // orderly close: not a failure
	if a.shuttingDown() {
		t.Fatal("an orderly listener close triggered a shutdown")
	}

	boom := errors.New("accept failed")
	a.onServeExit(boom)
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("a failed listener did not shut the process down")
	}
	if err := a.Err(); !errors.Is(err, boom) {
		t.Fatalf("Err() = %v; want it to wrap the listener failure", err)
	}
}

func assertNoPlaintext(t *testing.T, a *App) {
	t.Helper()
	for i, c := range a.Credentials() {
		if c.Token.Reveal() != "" {
			t.Errorf("Credentials()[%d] still holds plaintext", i)
		}
	}
}

// TestTUIReissueRoleZeroFailureReturnsEmptyMessage: when the first role's
// mint fails nothing was replaced and nothing is shown; the handler returns an
// empty message so a dismissed panel stays closed.
func TestTUIReissueRoleZeroFailureReturnsEmptyMessage(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	o.CredentialRand = &budgetRand{calls: 2} // start-up only
	a := startApp(t, o)

	msg, err := a.tuiReissue(context.Background())
	if err == nil {
		t.Fatal("tuiReissue succeeded although the approver mint had no entropy")
	}
	if msg.URL != "" || len(msg.Credentials) != 0 {
		t.Fatalf("message = %+v; want an empty CredentialsMsg", msg)
	}
	live, err := a.Reissue(context.Background())
	if err == nil || len(live) != 0 {
		t.Fatalf("Reissue = %d credentials, %v; want none and an error", len(live), err)
	}
}

// TestReissueAfterRevocationReturnsNoCredentials: once shutdown revoked the
// credentials a re-issue yields nothing, reports the revocation, and the
// TUI handler returns an empty message.
func TestReissueAfterRevocationReturnsNoCredentials(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	a := startApp(t, o)
	a.Shutdown("test")
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("no shutdown")
	}

	live, err := a.Reissue(context.Background())
	if len(live) != 0 {
		t.Fatalf("Reissue after revocation returned %d credentials; want none", len(live))
	}
	if !errors.Is(err, transport.ErrCredentialRevoked) || !errors.Is(err, transport.ErrShuttingDown) {
		t.Fatalf("Reissue error = %v; want ErrCredentialRevoked joined with ErrShuttingDown", err)
	}
	msg, err := a.tuiReissue(context.Background())
	if err == nil || msg.URL != "" || len(msg.Credentials) != 0 {
		t.Fatalf("tuiReissue = %+v, %v; want an empty message and an error", msg, err)
	}
	assertNoPlaintext(t, a)
}

// TestTUIDeliveryZeroesRetainedPlaintext: by default the plaintext is zeroed
// in App.Credentials() once the TUI has been given it, while the credentials
// still authenticate and keep their public metadata.
func TestTUIDeliveryZeroesRetainedPlaintext(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	rnd := &budgetRand{calls: 4}
	o.CredentialRand = rnd
	in, inW := io.Pipe()
	out := &syncBuf{}
	t.Cleanup(func() { _ = inW.Close(); _ = in.Close() })
	o.TUI = TUIOptions{Input: in, Output: out, Width: 100, Height: 30}
	a := startApp(t, o)

	approver, observer := rnd.token(t, 0), rnd.token(t, 1)
	waitOutput(t, out, approver) // the TUI was given the plaintext
	waitOutput(t, out, observer)

	creds := a.Credentials()
	if len(creds) != 2 {
		t.Fatalf("Credentials() = %d entries; want 2", len(creds))
	}
	for i, c := range creds {
		if c.Token.Reveal() != "" {
			t.Errorf("Credentials()[%d] retains plaintext after TUI delivery", i)
		}
		if c.Fingerprint == "" || c.ExpiresAt.IsZero() {
			t.Errorf("Credentials()[%d] lost its public metadata: %+v", i, c)
		}
	}
	if creds[0].Role != hub.RoleApprover || creds[1].Role != hub.RoleObserver {
		t.Errorf("roles = %v, %v; want approver, observer", creds[0].Role, creds[1].Role)
	}
	for _, tok := range []string{approver, observer} {
		if err := dialToken(a, tok); err != nil {
			t.Errorf("the zeroed-in-process credential no longer authenticates: %v", err)
		}
	}
}

// TestShutdownClearsCredentials: a normal shutdown leaves no plaintext in
// App.Credentials(), whether the TUI is disabled or retains plaintext, and
// without any Reissue call.
func TestShutdownClearsCredentials(t *testing.T) {
	cases := map[string]func(t *testing.T) *App{
		"tui disabled": func(t *testing.T) *App {
			f := newFactory()
			o, _ := baseOptions(t, f)
			return startApp(t, o)
		},
		"tui retains plaintext": func(t *testing.T) *App { return newUIRig(t).app },
	}
	for name, start := range cases {
		t.Run(name, func(t *testing.T) {
			a := start(t)
			if tokens(a)[0] == "" {
				t.Fatal("precondition: plaintext should be retained before shutdown")
			}
			a.Shutdown("test")
			select {
			case <-a.Done():
			case <-time.After(waitFor):
				t.Fatal("no shutdown")
			}
			assertNoPlaintext(t, a)
			if n := len(a.Credentials()); n != 0 {
				t.Errorf("Credentials() has %d entries after shutdown; want none", n)
			}
		})
	}
}

// TestReissuePausedAfterMintCannotRepublishAfterShutdown: a Reissue that
// paused after its final successful mint and resumes once shutdown revoked
// and cleared the credentials must refuse to publish them.
func TestReissuePausedAfterMintCannotRepublishAfterShutdown(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	a := startApp(t, o)

	reached, resume := make(chan struct{}), make(chan struct{})
	a.afterMint = func(role hub.Role) {
		if role == hub.RoleObserver { // the final mint of the re-issue
			close(reached)
			<-resume
		}
	}
	type result struct {
		live []transport.Issued
		err  error
	}
	res := make(chan result, 1)
	go func() {
		live, err := a.Reissue(context.Background())
		res <- result{live, err}
	}()
	select {
	case <-reached:
	case <-time.After(waitFor):
		t.Fatal("the re-issue never reached its final mint")
	}

	a.Shutdown("test")
	select {
	case <-a.Done():
	case <-time.After(waitFor):
		t.Fatal("shutdown did not complete while a re-issue was paused")
	}
	close(resume)

	select {
	case r := <-res:
		if !errors.Is(r.err, transport.ErrCredentialRevoked) {
			t.Fatalf("Reissue error = %v; want ErrCredentialRevoked", r.err)
		}
		if len(r.live) != 0 {
			t.Errorf("Reissue returned %d credentials after revocation", len(r.live))
		}
	case <-time.After(waitFor):
		t.Fatal("Reissue never returned")
	}
	assertNoPlaintext(t, a)
	if n := len(a.Credentials()); n != 0 {
		t.Errorf("Credentials() has %d entries; the revoked set was republished", n)
	}
}
