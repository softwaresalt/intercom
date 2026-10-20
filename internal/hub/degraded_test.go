package hub

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStalledConsumersNeverBlockHub(t *testing.T) {
	h := startHub(t, Config{IngressBuffer: 4})
	connect(t, h, ClientRemote, 2) // never read
	connect(t, h, ClientTUI, 2)    // never read
	const total = 20000
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for i := 0; i < total; i++ {
			h.Ingress() <- msgEvent("m", "x") // adapter-style blocking send
		}
	}()
	select {
	case <-finished:
	case <-time.After(testWait):
		t.Fatal("producer wedged: hub stopped draining ingress because of stalled consumers")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		s, err := h.Stats(ctx)
		if err != nil {
			t.Fatalf("hub unresponsive: %v", err)
		}
		if s.Head == total {
			if s.DroppedClients != 1 || !s.TUIAttached {
				t.Fatalf("stats = %+v", s)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
}

func TestDegradedChannelIsCapacityOneAndSignalNeverBlocks(t *testing.T) {
	h, err := New(Config{}) // not running: nobody drains the channel
	if err != nil {
		t.Fatal(err)
	}
	if got := cap(h.degraded); got != 1 {
		t.Fatalf("cap(Degraded) = %d, want 1", got)
	}
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			h.SignalDegraded("x")
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(testWait):
		t.Fatal("SignalDegraded blocked on a full degraded channel")
	}
}

func TestDegradedSignalYieldsHubBuiltErrorEnvelopeAndForcedRehydrate(t *testing.T) {
	h := startHub(t, Config{})
	clients := []*Client{connect(t, h, ClientTUI, 16), connect(t, h, ClientRemote, 16)}
	h.SignalDegraded("ingress saturated")
	for i, c := range clients {
		e := recv(t, c)
		if e.Kind() != KindError || e.Seq() != 1 || e.Gen() != h.Gen() {
			t.Fatalf("client %d: first frame %s seq %d, want hub-built error seq 1", i, e.Kind(), e.Seq())
		}
		var p ErrorPayload
		if err := json.Unmarshal(decode(t, e).Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Code != ErrCodeDegraded {
			t.Fatalf("error code = %q", p.Code)
		}
		hy := recv(t, c)
		if hy.Kind() != KindSessionHydrate || hy.Seq() != 1 {
			t.Fatalf("client %d: second frame %s seq %d, want forced hydrate at seq 1", i, hy.Kind(), hy.Seq())
		}
		if !strings.Contains(string(hy.Bytes()), ErrCodeDegraded) {
			t.Fatalf("client %d: hydrate snapshot does not record the fault", i)
		}
	}
	s := stats(t, h)
	if s.DegradedSignals != 1 || s.Rehydrations < 2 {
		t.Fatalf("stats = %+v", s)
	}
	// Streaming resumes after the forced re-hydrate with no gap.
	send(t, h, msgEvent("m", "x"))
	for i, c := range clients {
		if f := recv(t, c); f.Seq() != 2 {
			t.Fatalf("client %d: seq %d after rehydrate, want 2", i, f.Seq())
		}
	}
}

func TestDegradedReasonNeverReachesTheWire(t *testing.T) {
	h := startHub(t, Config{})
	c := connect(t, h, ClientRemote, 8)
	h.SignalDegraded("bad\x1b[31m" + strings.Repeat("A", 5000))
	var p ErrorPayload
	if err := json.Unmarshal(decode(t, recv(t, c)).Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Message != msgDegraded {
		t.Fatalf("wire message = %q, want the fixed %q", p.Message, msgDegraded)
	}
}

// bareSends returns the line of every channel send statement in src that is
// not the communication clause of a select (i.e. could block unconditionally).
func bareSends(t *testing.T, filename string, src any) []int {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	inSelect := map[*ast.SendStmt]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if cc, ok := n.(*ast.CommClause); ok {
			if s, ok := cc.Comm.(*ast.SendStmt); ok {
				inSelect[s] = true
			}
		}
		return true
	})
	var lines []int
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SendStmt); ok && !inSelect[s] {
			lines = append(lines, fset.Position(s.Pos()).Line)
		}
		return true
	})
	return lines
}

func TestBareSendDetectorFlagsUnconditionalSends(t *testing.T) {
	const bad = "package p\nfunc f(c chan int) {\n\tc <- 1\n}\n"
	if got := bareSends(t, "bad.go", bad); len(got) != 1 || got[0] != 3 {
		t.Fatalf("detector missed a bare send: %v", got)
	}
	const good = "package p\nfunc f(c chan int) {\n\tselect {\n\tcase c <- 1:\n\tdefault:\n\t}\n}\n"
	if got := bareSends(t, "good.go", good); len(got) != 0 {
		t.Fatalf("detector flagged a select send: %v", got)
	}
}

// TestHubPackageHasNoBareChannelSends enforces design §4.1 structurally: every
// send in the hub's production code is a select communication clause (with a
// default, or racing a done/ctx case), never an unconditional blocking send.
func TestHubPackageHasNoBareChannelSends(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if lines := bareSends(t, name, src); len(lines) > 0 {
			t.Errorf("%s: bare (potentially blocking) channel send at lines %v", name, lines)
		}
	}
}
