//go:build linux || darwin

package writepath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

// unixWriteFixture writes src at path, creating parent directories.
func unixWriteFixture(t *testing.T, path, src string) {
	t.Helper()
	u4Mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunFixtureSelfTest_SinkEscapesControlText pins the fixture self-test
// sinks (U3, red phase). The rows live in a linux || darwin file because the
// fixture names embed a newline, which Windows cannot create. The PASS and FAIL
// lines and the joined findings must escape that name, and no output may echo
// the absolute root. The not-a-contained-regular-file line is a
// characterization row, green on arrival, because errorLine already escapes it.
func TestRunFixtureSelfTest_SinkEscapesControlText(t *testing.T) {
	const (
		fixtureRel = "scripts/testdata/writepath"
		rejectName = "reject-a\n::error::x.go"
		acceptName = "accept-a\n::error::x.go"
		linkName   = "reject-a\n::error::y.go"
	)

	t.Run("PASS line escapes the fixture name", func(t *testing.T) {
		root := t.TempDir()
		unixWriteFixture(t, filepath.Join(root, fixtureRel, rejectName), u4RejectGo)

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
		unixWriteFixture(t, filepath.Join(root, fixtureRel, acceptName), unixTwoPrimitivesGo)

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
		unixWriteFixture(t, target, u4RejectGo)
		u4Mkdir(t, filepath.Join(root, fixtureRel))
		u4CreateSymlink(t, target, filepath.Join(root, fixtureRel, linkName))

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
