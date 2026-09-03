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
