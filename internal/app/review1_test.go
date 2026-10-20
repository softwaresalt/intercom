package app

import (
	"context"
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

// budgetRand is an entropy source that fails once its byte budget is spent.
type budgetRand struct {
	mu     sync.Mutex
	budget int
	next   byte
}

func (r *budgetRand) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.budget <= 0 {
		return 0, errors.New("entropy exhausted")
	}
	n := min(len(p), r.budget)
	for i := 0; i < n; i++ {
		r.next++
		p[i] = r.next
	}
	r.budget -= n
	return n, nil
}

const credBytes = 32

func partialRig(t *testing.T) (*App, []transport.Issued) {
	t.Helper()
	f := newFactory()
	o, _ := baseOptions(t, f)
	// two credentials at start, then one more: the re-issue's observer fails.
	o.CredentialRand = &budgetRand{budget: 3 * credBytes}
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
	select {
	case <-a.shutReq:
		t.Fatal("an orderly listener close triggered a shutdown")
	default:
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
