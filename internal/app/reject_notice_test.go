package app

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/tui"
)

func newRejectApp() *App {
	return &App{rejectCh: make(chan tui.SecurityNoticeMsg, 1)}
}

func TestOnRejectBeforeUICoalescesLatestWithoutBlocking(t *testing.T) {
	a := newRejectApp() // a.ui is nil: the callback must not need it
	within(t, testSlack, "1000 rejections with no consumer", func() {
		for i := uint64(1); i <= 1000; i++ {
			a.onReject(401, "credential_rejected", i)
		}
	})
	select {
	case msg := <-a.rejectCh:
		if msg.Total != 1000 || msg.Status != 401 || msg.Reason != "credential_rejected" {
			t.Fatalf("mailbox holds %+v; want the latest notice (total 1000)", msg)
		}
	default:
		t.Fatal("no notice was retained")
	}
	select {
	case msg := <-a.rejectCh:
		t.Fatalf("mailbox held a second notice: %+v", msg)
	default:
	}
}

func TestOnRejectKeepsHighestTotalWhenNoticesArriveOutOfOrder(t *testing.T) {
	a := newRejectApp()
	a.onReject(403, "host_not_allowed", 6)
	a.onReject(401, "credential_rejected", 5)
	select {
	case msg := <-a.rejectCh:
		if msg.Total != 6 || msg.Status != 403 || msg.Reason != "host_not_allowed" {
			t.Fatalf("mailbox holds %+v; want the notice with the highest total (6)", msg)
		}
	default:
		t.Fatal("no notice was retained")
	}
	if len(a.rejectCh) != 0 {
		t.Fatal("mailbox held a second notice")
	}
}

func TestOnRejectIsSafeConcurrently(t *testing.T) {
	a := newRejectApp()
	var wg sync.WaitGroup
	within(t, 5*time.Second, "concurrent rejections", func() {
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := uint64(1); i <= 200; i++ {
					a.onReject(400, "not_an_upgrade", i)
				}
			}()
		}
		wg.Wait()
	})
	if len(a.rejectCh) != 1 {
		t.Fatalf("mailbox length = %d; want 1", len(a.rejectCh))
	}
}

func TestPumpRejectsDeliversToTUIAndStopsWhenItExits(t *testing.T) {
	h := uiHub(t)
	o, out := uiTestOptions(t)
	ui, err := startUI(context.Background(), h, o, tui.CredentialsMsg{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	a := newRejectApp()
	pumped := make(chan struct{})
	go func() {
		defer close(pumped)
		a.pumpRejects(ui)
	}()

	a.onReject(401, "credential_rejected", 1)
	a.onReject(403, "host_not_allowed", 2)
	waitOutput(t, out, "rejected upgrades: 2 (last: 403 host_not_allowed)")

	if err := ui.stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-pumped:
	case <-time.After(testSlack):
		t.Fatal("pumpRejects did not return after the TUI exited")
	}
	within(t, testSlack, "a rejection after the TUI exited", func() {
		for i := uint64(3); i < 50; i++ {
			a.onReject(401, "credential_rejected", i)
		}
	})
}

func TestRejectedUpgradeIsShownOnLocalTUI(t *testing.T) {
	f := newFactory()
	o, _ := baseOptions(t, f)
	tuiOpts, out := uiTestOptions(t)
	o.TUI = tuiOpts
	a := startApp(t, o)

	url := strings.Replace(a.WSURL(), "ws://", "http://", 1)
	resp, err := http.Get(url) //nolint:noctx // loopback request to the test server
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode < 400 {
		t.Fatalf("a plain GET was not rejected: %d", resp.StatusCode)
	}
	waitOutput(t, out, "rejected upgrades: 1 (last: ")
	if strings.Contains(out.String(), "token=") {
		t.Fatal("notice output carries a query string")
	}
}

// A flood of rejected upgrades must reach the TUI as at most one update per
// interval: the first notice after idle is immediate, the rest coalesce in the
// mailbox (highest Total wins) and are delivered when the cooldown ends.
func TestPumpRejectsRateLimitsAndCoalescesBurstNotices(t *testing.T) {
	const interval = 150 * time.Millisecond
	a := newRejectApp()
	type sent struct {
		at  time.Time
		msg tui.SecurityNoticeMsg
	}
	var mu sync.Mutex
	var got []sent
	snapshot := func() []sent {
		mu.Lock()
		defer mu.Unlock()
		return append([]sent(nil), got...)
	}
	done := make(chan struct{})
	pumped := make(chan struct{})
	go func() {
		defer close(pumped)
		a.pumpRejectsTo(done, interval, func(m tui.SecurityNoticeMsg) {
			mu.Lock()
			got = append(got, sent{time.Now(), m})
			mu.Unlock()
		})
	}()

	start := time.Now()
	a.onReject(401, "credential_rejected", 1)
	waitFor := func(what string, cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(testSlack)
		for !cond() {
			if time.Now().After(deadline) {
				t.Fatalf("timed out waiting for %s", what)
			}
			time.Sleep(time.Millisecond)
		}
	}
	waitFor("the first notice", func() bool { return len(snapshot()) >= 1 })
	if d := snapshot()[0].at.Sub(start); d > interval/2 {
		t.Fatalf("first notice after idle took %v; want immediate", d)
	}
	for i := uint64(2); i <= 200; i++ { // a burst during the cooldown
		a.onReject(401, "credential_rejected", i)
	}
	waitFor("the coalesced notice", func() bool { return len(snapshot()) >= 2 })
	time.Sleep(interval / 3) // nothing more may follow within the next cooldown
	s := snapshot()
	if len(s) != 2 {
		t.Fatalf("burst produced %d sends; want 2 (first, then one coalesced)", len(s))
	}
	if s[1].msg.Total != 200 {
		t.Fatalf("coalesced notice Total = %d; want the highest (200)", s[1].msg.Total)
	}
	if gap := s[1].at.Sub(s[0].at); gap < interval-interval/10 {
		t.Fatalf("sends %v apart; want at least one interval (%v)", gap, interval)
	}

	close(done)
	select {
	case <-pumped:
	case <-time.After(testSlack):
		t.Fatal("pumpRejectsTo did not return after done closed")
	}
}

// Closing done mid-cooldown stops the timer wait at once.
func TestPumpRejectsStopsMidCooldownWhenDoneCloses(t *testing.T) {
	a := newRejectApp()
	sent := make(chan struct{}, 1)
	done := make(chan struct{})
	pumped := make(chan struct{})
	go func() {
		defer close(pumped)
		a.pumpRejectsTo(done, time.Hour, func(tui.SecurityNoticeMsg) { sent <- struct{}{} })
	}()
	a.onReject(401, "credential_rejected", 1)
	select {
	case <-sent:
	case <-time.After(testSlack):
		t.Fatal("first notice not delivered")
	}
	a.onReject(401, "credential_rejected", 2) // waits out an hour-long cooldown
	close(done)
	select {
	case <-pumped:
	case <-time.After(testSlack):
		t.Fatal("pumpRejectsTo waited for the cooldown timer after done closed")
	}
}
