package transport

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// pipeListener hands out pre-built in-memory connections, so a test controls
// exactly which sockets the limiter sees and how they behave.
type pipeListener struct {
	conns  chan net.Conn
	closed chan struct{}
	once   sync.Once
}

func newPipeListener() *pipeListener {
	return &pipeListener{conns: make(chan net.Conn, 256), closed: make(chan struct{})}
}

func (p *pipeListener) Accept() (net.Conn, error) {
	select {
	case c := <-p.conns:
		return c, nil
	case <-p.closed:
		return nil, net.ErrClosed
	}
}

func (p *pipeListener) Close() error {
	p.once.Do(func() { close(p.closed) })
	return nil
}

func (p *pipeListener) Addr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }

func waitUntil(t *testing.T, within time.Duration, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// Over-cap refusals must neither block nor write on the accept loop: every
// refused socket is counted at once and its 503 is written by a worker, a slot
// freed mid-flood is admitted promptly, a saturated refusal queue degrades to
// closing the socket (still counted), and Close stops the worker.
func TestLimitListenerOverCapRefusalOffAcceptLoopAndCountsReject(t *testing.T) {
	var rejected atomic.Int64
	pl := newPipeListener()
	ll := newLimitListener(pl, 1, func() { rejected.Add(1) })

	admittedClient, admittedServer := net.Pipe()
	defer func() { _ = admittedClient.Close() }()
	pl.conns <- admittedServer

	type accepted struct {
		c   net.Conn
		err error
	}
	results := make(chan accepted, 4)
	go func() {
		for {
			c, err := ll.Accept()
			results <- accepted{c, err}
			if err != nil {
				return
			}
		}
	}()
	first := <-results
	if first.err != nil {
		t.Fatal(first.err)
	}

	// Pipes that nobody reads: a synchronous 503 write would stall the accept
	// loop for the full write timeout per socket (10 x 50ms = 500ms).
	const burst = 10
	var clients []net.Conn
	for i := 0; i < burst; i++ {
		cl, sv := net.Pipe()
		clients = append(clients, cl)
		pl.conns <- sv
	}
	waitUntil(t, 250*time.Millisecond, "every over-cap socket counted off the accept loop", func() bool {
		return rejected.Load() == burst
	})

	// A freed slot is admitted without waiting behind the refusals.
	_ = first.c.Close()
	next, nextSrv := net.Pipe()
	defer func() { _ = next.Close() }()
	pl.conns <- nextSrv
	select {
	case r := <-results:
		if r.err != nil {
			t.Fatal(r.err)
		}
		defer func() { _ = r.c.Close() }()
	case <-time.After(250 * time.Millisecond):
		t.Fatal("a freed slot was not admitted promptly: the accept loop is blocked on refusals")
	}

	// Saturation: far more than the queue holds is still counted, and the
	// accept goroutine keeps running.
	before := rejected.Load()
	const flood = 3 * overCapRefuseQueueSize
	for i := 0; i < flood; i++ {
		cl, sv := net.Pipe()
		clients = append(clients, cl)
		pl.conns <- sv
	}
	waitUntil(t, time.Second, "flooded refusals counted", func() bool {
		return rejected.Load() == before+flood
	})

	// Close drains and stops the worker.
	if err := ll.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Close: %v", err)
	}
	select {
	case <-ll.workerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("the refusal worker did not exit after Close")
	}
	for _, c := range clients {
		_ = c.Close()
	}
}

// A refused client eventually reads the 503.
func TestLimitListenerOverCapWritesRefusal503(t *testing.T) {
	pl := newPipeListener()
	ll := newLimitListener(pl, 1, nil)
	defer func() { _ = ll.Close() }()

	holdCl, holdSv := net.Pipe()
	defer func() { _ = holdCl.Close() }()
	pl.conns <- holdSv
	go func() {
		for {
			if _, err := ll.Accept(); err != nil {
				return
			}
		}
	}()
	waitUntil(t, time.Second, "first connection admitted", func() bool { return ll.active.Load() == 1 })

	cl, sv := net.Pipe()
	defer func() { _ = cl.Close() }()
	pl.conns <- sv
	buf := make([]byte, len(refusal503))
	_ = cl.SetReadDeadline(time.Now().Add(2 * time.Second))
	n := 0
	for n < len(buf) {
		m, err := cl.Read(buf[n:])
		if err != nil {
			t.Fatalf("read refusal: %v (got %q)", err, buf[:n])
		}
		n += m
	}
	if string(buf) != refusal503 {
		t.Fatalf("refusal = %q; want %q", buf, refusal503)
	}
}
