//go:build linux || darwin

package writepath

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// unixTwoPrimitivesGo holds two write primitives, so a fixture built from it
// yields two findings and exercises the joined text of a FAIL line.
const unixTwoPrimitivesGo = `package p

import "os"

func f() {
	_ = os.Remove("x")
	_ = os.Remove("y")
}
`

// TestRunFixtureSelfTest_SinkEscapesControlText pins the fixture self-test
// sinks. The rows live in a linux || darwin file because the fixture names
// embed a newline, which Windows cannot create. The PASS and FAIL lines and the
// joined findings must escape that name, and no output may echo the absolute
// root. The not-a-contained-regular-file line is a characterization row, because
// errorLine already escapes it. The PASS and FAIL rows pin the escape.
func TestRunFixtureSelfTest_SinkEscapesControlText(t *testing.T) {
	const (
		fixtureRel = "scripts/testdata/writepath"
		rejectName = "reject-a\n::error::x.go"
		acceptName = "accept-a\n::error::x.go"
		linkName   = "reject-a\n::error::y.go"
	)

	t.Run("PASS line escapes the fixture name", func(t *testing.T) {
		root := t.TempDir()
		writeText(t, filepath.Join(root, fixtureRel, rejectName), rejectGoSource)

		res := runFixtureSelfTest(root)
		assertNoRawControl(t, res.Stdout)
		if strings.Contains(res.Stdout, root) {
			t.Fatalf("stdout echoes absolute root %q: %q", root, res.Stdout)
		}
		want := `PASS reject-a\x0a::error::x.go: rejected as expected` + "\n"
		if res.Code != 0 || res.Stdout != want || res.Stderr != "" {
			t.Fatalf("got (Code=%d, stdout=%q, stderr=%q), want (0, %q, empty stderr)",
				res.Code, res.Stdout, res.Stderr, want)
		}
	})

	t.Run("FAIL line and joined findings escape the fixture name", func(t *testing.T) {
		root := t.TempDir()
		writeText(t, filepath.Join(root, fixtureRel, acceptName), unixTwoPrimitivesGo)

		res := runFixtureSelfTest(root)
		assertNoRawControl(t, res.Stderr)
		if strings.Contains(res.Stderr, root) {
			t.Fatalf("stderr echoes absolute root %q: %q", root, res.Stderr)
		}
		want := `FAIL accept-a\x0a::error::x.go: expected clean, got findings: ` +
			`scripts/testdata/writepath/accept-a\x0a::error::x.go:6: write primitive 'os.Remove' found; ` +
			`scripts/testdata/writepath/accept-a\x0a::error::x.go:7: write primitive 'os.Remove' found` + "\n"
		if res.Code != 1 || res.Stdout != "" || res.Stderr != want {
			t.Fatalf("got (Code=%d, stdout=%q, stderr=%q), want (1, empty stdout, %q)",
				res.Code, res.Stdout, res.Stderr, want)
		}
	})

	t.Run("not a contained regular file line escapes the name (characterization)", func(t *testing.T) {
		temp := t.TempDir()
		root := filepath.Join(temp, "repo")
		target := filepath.Join(temp, "outside", "y.go")
		writeText(t, target, rejectGoSource)
		mkdirAll(t, filepath.Join(root, fixtureRel))
		createSymlink(t, target, filepath.Join(root, fixtureRel, linkName))

		res := runFixtureSelfTest(root)
		assertNoRawControl(t, res.Stderr)
		if strings.Contains(res.Stderr, temp) {
			t.Fatalf("stderr echoes absolute temp path %q: %q", temp, res.Stderr)
		}
		wantPrefix := `::error::scripts/testdata/writepath/reject-a\x0a::error::y.go: ` +
			`not a contained regular file: scripts/testdata/writepath/reject-a\x0a::error::y.go is not a regular file (mode `
		if res.Code != 1 || res.Stdout != "" || !strings.HasPrefix(res.Stderr, wantPrefix) {
			t.Fatalf("got (Code=%d, stdout=%q, stderr=%q), want (1, empty stdout, prefix %q)",
				res.Code, res.Stdout, res.Stderr, wantPrefix)
		}
	})
}

// TestRunRepoScan_NewlineNonSelectedNameEscapes is the Linux-only outcome-4 row
// for a real newline in a non-selected directory name. Windows cannot create
// such a name, so the row lives in this linux || darwin file.
func TestRunRepoScan_NewlineNonSelectedNameEscapes(t *testing.T) {
	const rel = "internal/a\n::error::x"
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, filepath.FromSlash(rel)))

	res := runRepoScan(root, stubGitListing(rel))
	want := `::error::internal/a\x0a::error::x: not a contained regular file: `
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1 (stderr=%q)", res.Code, res.Stderr)
	}
	if n := strings.Count(res.Stderr, "\n"); n != 1 {
		t.Fatalf("stderr has %d lines, want exactly one: %q", n, res.Stderr)
	}
	assertNoRawControl(t, res.Stderr)
	if !strings.HasPrefix(res.Stderr, want) {
		t.Fatalf("stderr = %q, want it to start with %q", res.Stderr, want)
	}
}

// TestRunFixtureSelfTest_FIFOFixtureRefused is a characterization row, green on
// arrival (FIFO already refused by containment). A named pipe reject-fifo.go
// beside a valid fixture must be refused as not a regular file. The self-test
// runs in a goroutine, so a regression that opens the FIFO fails at the timeout
// instead of hanging. Cleanup opens the FIFO O_RDWR|O_NONBLOCK, which unblocks a
// stuck opener, then waits for the goroutine. No t.Skip.
func TestRunFixtureSelfTest_FIFOFixtureRefused(t *testing.T) {
	const (
		fixtureRel = "scripts/testdata/writepath"
		waitLimit  = 10 * time.Second
	)
	root := t.TempDir()
	fixtureDir := filepath.Join(root, filepath.FromSlash(fixtureRel))
	writeText(t, filepath.Join(fixtureDir, "accept-ok.go"), cleanGoSource)
	fifo := filepath.Join(fixtureDir, "reject-fifo.go")
	if err := syscall.Mkfifo(fifo, 0o644); err != nil {
		t.Fatalf("mkfifo %s: %v", fifo, err)
	}

	results := make(chan Result, 1)
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		results <- runFixtureSelfTest(root)
	}()
	t.Cleanup(func() {
		if f, err := os.OpenFile(fifo, os.O_RDWR|syscall.O_NONBLOCK, 0); err == nil {
			_ = f.Close()
		}
		select {
		case <-finished:
		case <-time.After(waitLimit):
			t.Errorf("runFixtureSelfTest still blocked %v after the FIFO was unblocked", waitLimit)
		}
	})

	var res Result
	select {
	case res = <-results:
	case <-time.After(waitLimit):
		t.Fatalf("runFixtureSelfTest did not return within %v: it may have opened the FIFO", waitLimit)
	}

	wantPrefix := `::error::scripts/testdata/writepath/reject-fifo.go: ` +
		`not a contained regular file: scripts/testdata/writepath/reject-fifo.go is not a regular file (mode `
	assertNoRawControl(t, res.Stderr)
	if res.Code != 1 || !strings.HasPrefix(res.Stderr, wantPrefix) {
		t.Fatalf("got (Code=%d, stdout=%q, stderr=%q), want (1, prefix %q)",
			res.Code, res.Stdout, res.Stderr, wantPrefix)
	}
}
