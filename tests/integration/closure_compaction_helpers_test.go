package integration

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertClosureBandArchived verifies every path in a closure band is archived
// (not deleted), absent from docs/closure, and listed in the band summary.
func assertClosureBandArchived(t *testing.T, root, summaryRel string, paths []string) {
	t.Helper()

	summaryBytes, summaryErr := os.ReadFile(filepath.Join(root, "docs", "closure", summaryRel))
	if summaryErr != nil {
		t.Errorf("summary %s missing: %v", summaryRel, summaryErr)
	}
	summary := string(summaryBytes)

	for _, name := range paths {
		if _, err := os.Stat(filepath.Join(root, "docs", "archive", "closure", name)); err != nil {
			t.Errorf("%s not archived under docs/archive/closure: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(root, "docs", "closure", name)); err == nil {
			t.Errorf("%s still present in docs/closure", name)
		} else if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("stat docs/closure/%s: %v", name, err)
		}
		if summaryErr == nil && !strings.Contains(summary, "`"+name+"`") {
			t.Errorf("summary does not list %s in backticks", name)
		}
	}
}
