package pathsafe

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestNormalizeRejectsTraversal verifies escape attempts are rejected with
// "path attempts to escape workspace".
func TestNormalizeRejectsTraversal(t *testing.T) {
	cases := []string{
		"../secret.txt",
		"src/../../secret.txt",
		"subdir/../../escape.txt",
	}
	for _, c := range cases {
		_, err := normalize(c)
		if err == nil {
			t.Fatalf("normalize(%q) = nil error, want error containing %q", c, escapeMsg)
		}
		if !strings.Contains(err.Error(), escapeMsg) {
			t.Fatalf("normalize(%q) error = %q, want to contain %q", c, err.Error(), escapeMsg)
		}
	}
}

// TestNormalizeRejectsAbsoluteAndVolumeRelative verifies absolute and
// volume-relative candidates are rejected with the absolute-path message,
// including the Windows drive-relative form (finding SEC-1, SEC-2).
func TestNormalizeRejectsAbsoluteAndVolumeRelative(t *testing.T) {
	cases := []string{"/etc/passwd"}
	if runtime.GOOS == "windows" {
		cases = append(cases, `C:\Windows\system32`, "C:foo")
	}

	for _, c := range cases {
		_, err := normalize(c)
		if err == nil {
			t.Fatalf("normalize(%q) = nil error, want error containing %q", c, absMsg)
		}
		if !strings.Contains(err.Error(), absMsg) {
			t.Fatalf("normalize(%q) error = %q, want to contain %q", c, err.Error(), absMsg)
		}
	}
}

// TestNormalizeDoesNotOverRejectLiteralBackslashOnNonWindows verifies that a
// relative candidate merely beginning with a literal backslash byte is not
// rejected as "rooted" on non-Windows platforms, where '\' has no
// path-separator meaning. This guards the isRooted GOOS-gating fix: only the
// leading-'/' check applies unconditionally; the leading-'\' check is scoped
// to runtime.GOOS == "windows".
func TestNormalizeDoesNotOverRejectLiteralBackslashOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("backslash is a path separator on windows; this case is windows-drive-relative there, not a literal filename byte")
	}

	components, err := normalize(`\odd-filename.txt`)
	if err != nil {
		t.Fatalf(`normalize("\odd-filename.txt") returned unexpected error: %v`, err)
	}
	if len(components) != 1 || components[0] != `\odd-filename.txt` {
		t.Fatalf(`normalize("\odd-filename.txt") = %v, want a single literal component`, components)
	}
}

// TestNormalizePreservesSafeInteriorTraversal verifies a/b/../c.txt
// normalizes to a/c.txt and ./src/main.go normalizes to src/main.go
// (finding GO-13, proving Clean equivalence).
func TestNormalizePreservesSafeInteriorTraversal(t *testing.T) {
	cases := map[string]string{
		"a/b/../c.txt":  filepath.Join("a", "c.txt"),
		"./src/main.go": filepath.Join("src", "main.go"),
	}
	for input, want := range cases {
		components, err := normalize(input)
		if err != nil {
			t.Fatalf("normalize(%q) returned unexpected error: %v", input, err)
		}
		got := filepath.Join(components...)
		if got != want {
			t.Fatalf("normalize(%q) joined = %q, want %q", input, got, want)
		}
	}
}

// TestNormalizeTreatsUnexpandedEnvVarSegmentLiterally pins 011.008-T
// (resolves 8472E0A1 item (g)): a candidate segment shaped like an
// unexpanded environment-variable reference ("$VAR" or "%VAR%") is treated
// as an ordinary, literal path component -- never substituted with an
// environment value. This proves normalize performs no expansion of any
// kind; expansion (if wanted) is entirely the caller's responsibility.
func TestNormalizeTreatsUnexpandedEnvVarSegmentLiterally(t *testing.T) {
	cases := []string{
		"$VAR/file.txt",
		"%VAR%/file.txt",
		"${VAR}/file.txt",
	}
	for _, input := range cases {
		components, err := normalize(input)
		if err != nil {
			t.Fatalf("normalize(%q) returned unexpected error: %v", input, err)
		}
		if len(components) == 0 {
			t.Fatalf("normalize(%q) = no components, want at least one literal component", input)
		}
		first := components[0]
		if !strings.ContainsAny(first, "$%") {
			t.Fatalf("normalize(%q) first component = %q, want the literal, unexpanded env-var-shaped segment preserved verbatim", input, first)
		}
	}
}

// TestNormalizeAcceptsLegitimateDollarAndPercentFilenames is the paired
// positive test for 011.008-T: real filenames containing '$' or '%' (both
// legal on their native platforms -- e.g. Windows' own top-level
// "$Recycle.Bin" directory, and '%' as a legal POSIX filename byte) must
// still validate successfully. No refusal class is added by 011.008-T (J2:
// zero verdict change).
func TestNormalizeAcceptsLegitimateDollarAndPercentFilenames(t *testing.T) {
	cases := []string{
		"$Recycle.Bin/desktop.ini",
		"$WinREAgent/config",
		"reports/100%-complete.txt",
	}
	for _, input := range cases {
		components, err := normalize(input)
		if err != nil {
			t.Fatalf("normalize(%q) returned unexpected error: %v, want acceptance (no refusal class added)", input, err)
		}
		if len(components) == 0 {
			t.Fatalf("normalize(%q) = no components, want a valid component list", input)
		}
	}
}

func TestContainsDotDotSegment(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "windows-style traversal", path: `..\..\etc`, want: true},
		{name: "posix traversal", path: "../secret", want: true},
		{name: "safe relative path", path: "data/agent-rc.db", want: false},
		{name: "absolute path without traversal", path: "/var/lib/agent-rc.db", want: false},
		{name: "interior dots are not traversal", path: "data/.../agent-rc.db", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContainsDotDotSegment(tc.path); got != tc.want {
				t.Fatalf("ContainsDotDotSegment(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
