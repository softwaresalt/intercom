//go:build windows

package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAddLongPathPrefixVerdicts pins the U5/AC2 prefixing predicate
// explicitly (016.004-T): a boundary test covering the MAX_PATH threshold,
// plus the classes review flagged as negative coverage (U5/AC4) -- an
// already-extended `\\?\Volume{GUID}\...` input must not be double-prefixed,
// and a `\\.\` device-namespace path must not be prefixed at all. RED
// against the current package: addLongPathPrefix does not yet exist.
func TestAddLongPathPrefixVerdicts(t *testing.T) {
	longSuffix := strings.Repeat("a", 300)
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "short path unaffected",
			in:   `C:\short\path`,
			want: `C:\short\path`,
		},
		{
			name: "long absolute path gets prefixed",
			in:   `C:\` + longSuffix,
			want: uncPrefix + `C:\` + longSuffix,
		},
		{
			name: "boundary just below threshold unaffected",
			in:   `C:\` + strings.Repeat("a", longPathThreshold-len(`C:\`)-1),
			want: `C:\` + strings.Repeat("a", longPathThreshold-len(`C:\`)-1),
		},
		{
			name: "boundary at threshold gets prefixed",
			in:   `C:\` + strings.Repeat("a", longPathThreshold-len(`C:\`)),
			want: uncPrefix + `C:\` + strings.Repeat("a", longPathThreshold-len(`C:\`)),
		},
		{
			name: "relative long path is not prefixed",
			in:   longSuffix,
			want: longSuffix,
		},
		{
			name: "already-extended volume GUID path is not double-prefixed",
			in:   `\\?\Volume{12345678-1234-1234-1234-123456789abc}\` + longSuffix,
			want: `\\?\Volume{12345678-1234-1234-1234-123456789abc}\` + longSuffix,
		},
		{
			name: "already \\?\\-prefixed drive path is not double-prefixed",
			in:   uncPrefix + `C:\` + longSuffix,
			want: uncPrefix + `C:\` + longSuffix,
		},
		{
			name: "device namespace path is never prefixed",
			in:   `\\.\` + longSuffix,
			want: `\\.\` + longSuffix,
		},
		{
			name: "bare long UNC path is not prefixed (AC5 residual)",
			in:   `\\server\share\` + longSuffix,
			want: `\\server\share\` + longSuffix,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := addLongPathPrefix(tc.in)
			if got != tc.want {
				t.Fatalf("addLongPathPrefix(len=%d) = %q, want %q", len(tc.in), got, tc.want)
			}
		})
	}
}

// TestAddLongPathPrefixRoundTripsWithStripUNCPrefix verifies the
// prefix/strip round-trip required by U5/AC3: prefixing a long path and
// then stripping it via stripUNCPrefix (U6's internalized postcondition)
// recovers the original input.
func TestAddLongPathPrefixRoundTripsWithStripUNCPrefix(t *testing.T) {
	long := `C:\` + strings.Repeat("a", 300)
	prefixed := addLongPathPrefix(long)
	if prefixed == long {
		t.Fatalf("addLongPathPrefix(%q) did not add a prefix for a long path", long)
	}
	got := stripUNCPrefix(prefixed)
	if got != long {
		t.Fatalf("stripUNCPrefix(addLongPathPrefix(%q)) = %q, want %q", long, got, long)
	}
}

// canonicalForComparison normalizes want through filepath.EvalSymlinks so
// test assertions are not defeated by an incidental Windows 8.3 short-name
// alias in a t.TempDir() path (e.g. "DEWILL~1" vs "dewilliams") -- a naming
// artifact of the test environment, not a canonicalizeReparse defect.
// filepath.EvalSymlinks performs the same long-name normalization Windows
// applies internally, giving an independent oracle for the expected value.
func canonicalForComparison(t *testing.T, want string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) returned error: %v", want, err)
	}
	return stripUNCPrefix(resolved)
}

// TestCanonicalizeReparsePlainDirectory verifies an ordinary directory with
// no reparse point resolves to itself (014.003-T AC6, case 1).
func TestCanonicalizeReparsePlainDirectory(t *testing.T) {
	dir := t.TempDir()
	want := canonicalForComparison(t, dir)

	got, err := canonicalizeReparse(dir)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", dir, err)
	}
	got = stripUNCPrefix(got)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", dir, got, want)
	}
}

// TestCanonicalizeReparseInRootJunction verifies a live directory junction
// whose target lies inside the workspace resolves to the real (in-root)
// target rather than the junction's own lexical path (014.003-T AC6, case
// 2).
func TestCanonicalizeReparseInRootJunction(t *testing.T) {
	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	targetDir := filepath.Join(root.Path(), "real-target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create in-root junction target: %v", err)
	}

	linkPath := filepath.Join(root.Path(), "junction-in-root")
	createDirectoryJunction(t, linkPath, targetDir)

	got, err := canonicalizeReparse(linkPath)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", linkPath, err)
	}
	got = stripUNCPrefix(got)
	if !hasPathPrefix(got, root.Path()) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want a path inside root %q", linkPath, got, root.Path())
	}
	want := canonicalForComparison(t, targetDir)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", linkPath, got, want)
	}
}

// TestCanonicalizeReparseOutOfRootJunction verifies a live directory
// junction whose target lies outside the workspace resolves to the true,
// outside-root target -- the core defect this helper closes (014.003-T
// AC6, case 3).
func TestCanonicalizeReparseOutOfRootJunction(t *testing.T) {
	outsideRoot := t.TempDir()

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-out")
	createDirectoryJunction(t, linkPath, outsideRoot)

	got, err := canonicalizeReparse(linkPath)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", linkPath, err)
	}
	got = stripUNCPrefix(got)
	if hasPathPrefix(got, root.Path()) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want a path outside root %q", linkPath, got, root.Path())
	}
	want := canonicalForComparison(t, outsideRoot)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", linkPath, got, want)
	}
}

// TestCanonicalizeReparseReturnsNormalizedPath verifies canonicalizeReparse
// internalizes the stripUNCPrefix postcondition itself (016.003-T, U6): the
// raw GetFinalPathNameByHandleW result carries a leading `\\?\` extended-path
// prefix, and every existing caller of this function had to strip it
// independently. This test asserts the returned value is ALREADY normalized
// -- no leading `\\?\` -- so the postcondition holds for any caller, not just
// the ones that remember to call stripUNCPrefix themselves. RED against the
// current implementation: GetFinalPathNameByHandleW's raw result is returned
// unmodified today.
func TestCanonicalizeReparseReturnsNormalizedPath(t *testing.T) {
	dir := t.TempDir()

	got, err := canonicalizeReparse(dir)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", dir, err)
	}
	if strings.HasPrefix(got, uncPrefix) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want no leading %q prefix (postcondition must be internalized)", dir, got, uncPrefix)
	}
	want := canonicalForComparison(t, dir)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", dir, got, want)
	}
}

// TestCanonicalizeReparsePreservesVolumeGUIDPrefix pins D-6: a
// `\\?\Volume{GUID}\...` result must NOT be stripped, because
// stripUNCPrefix's own contract only re-forms drive-letter and UNC-share
// forms -- a Volume{GUID} path has no non-`\\?\` equivalent. This proves
// U6's internalization does not regress that non-stripping behavior for a
// path shape canonicalizeReparse cannot itself construct from a real
// filesystem probe, so it is verified directly against stripUNCPrefix, the
// function U6 now calls internally.
func TestCanonicalizeReparsePreservesVolumeGUIDPrefix(t *testing.T) {
	in := `\\?\Volume{12345678-1234-1234-1234-123456789abc}\some\path`
	got := stripUNCPrefix(in)
	if got != in {
		t.Fatalf("stripUNCPrefix(%q) = %q, want unchanged (Volume{GUID} forms are not strippable)", in, got)
	}
}

// TestCanonicalizeReparseResolvesDeeplyNestedInRootAncestor is the U5/AC1
// end-to-end long-path fixture: a deeply-nested in-root ancestor exceeding
// MAX_PATH must resolve successfully through canonicalizeReparse rather than
// being spuriously rejected by a raw, unprefixed syscall.CreateFile call.
//
// Measured residual (U5/AC1, plan H3): on this shipment's verified
// development environment (Windows 10.0.26200, LongPathsEnabled=0), a raw
// syscall.CreateFile call already succeeds for a >1000-character absolute
// path with NO `\\?\` prefix -- so this fixture does not reproduce a
// pre-fix failure here, and this test cannot serve as a fail-before/
// pass-after RED proof on this host. This is the exact environment
// sensitivity the plan's H3 mitigation anticipated for the >MAX_PATH
// fixture. Per that mitigation, this end-to-end case is retained as a
// non-regression lock (must pass on every environment, before and after),
// while the deterministic, OS-independent predicate coverage above
// (TestAddLongPathPrefixVerdicts, TestAddLongPathPrefixRoundTripsWithStripUNCPrefix)
// is the covering test for the accepted limitation recorded in the package
// risk register (016.008-T): environments that DO enforce classic MAX_PATH
// without the registry opt-in are protected by addLongPathPrefix's
// predicate, which is proven correct in isolation even though this
// particular host cannot exercise the failure branch it guards against.
func TestCanonicalizeReparseResolvesDeeplyNestedInRootAncestor(t *testing.T) {
	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	cur := root.Path()
	seg := strings.Repeat("a", 50)
	for len(cur) < longPathThreshold+50 {
		cur = filepath.Join(cur, seg)
		if err := os.Mkdir(cur, 0o755); err != nil {
			t.Fatalf("failed to build deeply-nested fixture at len=%d: %v", len(cur), err)
		}
	}
	if len(cur) < longPathThreshold {
		t.Fatalf("fixture path length %d did not reach longPathThreshold %d", len(cur), longPathThreshold)
	}

	got, err := canonicalizeReparse(cur)
	if err != nil {
		t.Fatalf("canonicalizeReparse(len=%d) returned unexpected error: %v -- want the long in-root ancestor to resolve successfully", len(cur), err)
	}
	if !hasPathPrefix(got, root.Path()) {
		t.Fatalf("canonicalizeReparse(len=%d) = %q, want a path inside root %q", len(cur), got, root.Path())
	}
}

// TestCanonicalizeReparseNonExistentPathErrors verifies a non-existent path
// returns an error rather than a zero-value success (014.003-T AC6, case
// 4).
func TestCanonicalizeReparseNonExistentPathErrors(t *testing.T) {
	rootDir := t.TempDir()
	missing := filepath.Join(rootDir, "does-not-exist")

	_, err := canonicalizeReparse(missing)
	if err == nil {
		t.Fatalf("canonicalizeReparse(%q) = nil error, want an error for a non-existent path", missing)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("canonicalizeReparse(%q) error = %v, want it to satisfy errors.Is(err, fs.ErrNotExist)", missing, err)
	}
}
