//go:build windows

package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
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

// TestAddLongPathPrefixDeviceNamespaceBranchReachability is the 017.002-T
// characterization/regression lock. It CORRECTS a stale premise recorded in
// a prior stash entry -- that addLongPathPrefix's `\\.\` device-namespace
// exclusion branch is "currently unreachable given filepath.IsAbs's own
// device-path handling". That premise is FALSE: filepath.IsAbs reports true
// for every `\\.\` form below, so the preceding `!filepath.IsAbs(path)`
// guard never shadows the `\\.\` branch.
//
// This is deliberately a BLACK-BOX, order-independent characterization, not
// a white-box assertion about which specific guard line executes: it locks
// the two OBSERVABLE facts (filepath.IsAbs's verdict, and
// addLongPathPrefix's returned value) rather than internal control flow, so
// it remains a valid regression lock regardless of how the guard chain
// inside addLongPathPrefix is ordered or refactored in the future. Per D-4
// (reparse_windows.go), the `\\.\` branch is reachable but currently
// redundant with the immediately-following bare `\\` guard -- it is
// retained as self-documenting defense-in-depth, not deleted, and this test
// is what makes that retention decision enforceable rather than merely
// prose.
//
// Per Constitution Principle II's documented deviation (017.002-T,
// characterization-first posture): this test makes NO production behavior
// change (addLongPathPrefix's `\\.\` guard already existed and already
// passed these cases), so there is no failing-before/passing-after RED
// phase to drive -- it is a Feathers-style characterization test locking in
// already-verified-correct behavior, not a bug fix.
func TestAddLongPathPrefixDeviceNamespaceBranchReachability(t *testing.T) {
	longSuffix := strings.Repeat("a", 300)
	cases := []struct {
		name string
		in   string
	}{
		{name: "device path to drive-relative file", in: `\\.\C:\foo`},
		{name: "device path to physical drive", in: `\\.\PhysicalDrive0`},
		{name: "device path to UNC-style device target", in: `\\.\UNC\srv\sh\x`},
		{name: "device path at/beyond MAX_PATH threshold", in: `\\.\C:` + longSuffix},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Lock the corrected premise: filepath.IsAbs does NOT exclude
			// `\\.\` device-namespace paths, so the guard chain's
			// `!filepath.IsAbs(path)` early return never shadows the
			// `\\.\` branch for these inputs.
			if !filepath.IsAbs(tc.in) {
				t.Fatalf("filepath.IsAbs(%q) = false, want true -- this is the corrected premise: device-namespace paths ARE reported absolute", tc.in)
			}
			// Lock the observable verdict: a `\\.\` input is never
			// extended-prefixed, regardless of length.
			got := addLongPathPrefix(tc.in)
			if got != tc.in {
				t.Fatalf(`addLongPathPrefix(%q) = %q, want the input returned unchanged (device-namespace paths are never \\?\-prefixed)`, tc.in, got)
			}
		})
	}
}

// TestAddLongPathPrefixIsUTF16CodeUnitAware is the 017.003-T RED/GREEN
// lock. Windows measures MAX_PATH in UTF-16 code units, not UTF-8 bytes;
// len(path) (a UTF-8 byte count) is a conservative proxy that can only ever
// fire AT OR BEFORE the true limit (1 UTF-8 byte -> at most 1 UTF-16 unit;
// 2/3 bytes -> 1 unit; 4 bytes -> a 2-unit surrogate pair), so it can
// produce a FALSE POSITIVE (prefixing a path that did not need it) but
// never a false negative. This is a PRECISION improvement, not a
// correctness or security fix -- framing it as a vulnerability fix is an
// explicit anti-goal (plan D-5).
//
// RED against pre-017.003-T addLongPathPrefix: the multi-byte cases below
// have a UTF-8 byte length at/beyond longPathThreshold (260) but a true
// UTF-16 code-unit count well below it, so the byte-length proxy currently
// (wrongly, but safely) prefixes them.
func TestAddLongPathPrefixIsUTF16CodeUnitAware(t *testing.T) {
	base := `C:\`
	eAcute := strings.Repeat("\u00e9", 200)    // 200 runes: 400 UTF-8 bytes, 200 UTF-16 units
	emoji := strings.Repeat("\U0001F600", 100) // 100 runes: 400 UTF-8 bytes, 200 UTF-16 units (surrogate pairs)

	cases := []struct {
		name       string
		in         string
		wantPrefix bool
	}{
		{
			name:       "e-acute run: 403 UTF-8 bytes but 203 UTF-16 units, below threshold -- must NOT be prefixed",
			in:         base + eAcute,
			wantPrefix: false,
		},
		{
			name:       "emoji run: 403 UTF-8 bytes but 203 UTF-16 units (100 surrogate pairs), below threshold -- must NOT be prefixed",
			in:         base + emoji,
			wantPrefix: false,
		},
		{
			name:       "e-acute run long enough to cross the UTF-16 threshold too -- must still be prefixed",
			in:         base + strings.Repeat("\u00e9", 260),
			wantPrefix: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := addLongPathPrefix(tc.in)
			gotPrefixed := got != tc.in
			if gotPrefixed != tc.wantPrefix {
				t.Fatalf("addLongPathPrefix(%q) prefixed=%v, want prefixed=%v (utf16 units=%d, utf8 bytes=%d, threshold=%d)",
					tc.in, gotPrefixed, tc.wantPrefix, utf16Len(tc.in), len(tc.in), longPathThreshold)
			}
			if tc.wantPrefix && !strings.HasPrefix(got, uncPrefix) {
				t.Fatalf("addLongPathPrefix(%q) = %q, want it to start with the extended-length prefix %q", tc.in, got, uncPrefix)
			}
		})
	}
}

// TestUTF16LenMatchesStdlibEncoding verifies utf16Len's counting method
// (rune-by-rune surrogate-pair accumulation, avoiding an intermediate
// []rune / []uint16 allocation) agrees with the stdlib
// len(utf16.Encode([]rune(s))) reference computation the plan endorsed
// (O2-b), across ASCII, 2-3 byte BMP characters, and 4-byte
// (surrogate-pair) characters. A plain rune count (O2-a, REJECTED in the
// plan) would UNDER-count the emoji case -- 1 rune is 2 UTF-16 units for
// any codepoint above U+FFFF -- so this test would catch a regression to
// that rejected approach.
func TestUTF16LenMatchesStdlibEncoding(t *testing.T) {
	cases := []string{
		"",
		`C:\short\path`,
		strings.Repeat("a", 300),
		strings.Repeat("\u00e9", 200),
		strings.Repeat("\U0001F600", 100),
		`C:\` + strings.Repeat("\u00e9", 50) + strings.Repeat("\U0001F600", 25),
		// Invalid UTF-8 (a lone continuation byte, and a truncated
		// multi-byte sequence): both range-over-string and []rune(s)
		// decode each invalid byte to exactly one U+FFFD replacement
		// rune, so utf16Len's rune-by-rune accumulation and the stdlib
		// utf16.Encode([]rune(s)) reference necessarily agree here too
		// (review follow-up, code-review pass).
		"C:\\" + string([]byte{0xff, 0xfe}) + "\\path",
	}
	for _, s := range cases {
		want := len(utf16.Encode([]rune(s)))
		got := utf16Len(s)
		if got != want {
			t.Fatalf("utf16Len(%q) = %d, want %d (stdlib utf16.Encode reference)", s, got, want)
		}
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
