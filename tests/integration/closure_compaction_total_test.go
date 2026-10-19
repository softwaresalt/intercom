// Frozen snapshot guard for the one-time 055-S/063-F closure compaction.
// Delete or supersede it when a later compaction re-bands or archives these records.

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

// TestClosureCompactionArchivesAll37Records verifies the D1-D3 bands together
// account for 37 archived closure records.
func TestClosureCompactionArchivesAll37Records(t *testing.T) {
	root := repoRoot(t)

	bands := []struct {
		name  string
		paths []string
		want  int
	}{
		{"band1", closureBand1ArchivedPaths, 14},
		{"band2", closureBand2ArchivedPaths, 13},
		{"band3", closureBand3ArchivedPaths, 10},
	}

	total := 0
	for _, b := range bands {
		if len(b.paths) != b.want {
			t.Errorf("%s lists %d paths, want %d", b.name, len(b.paths), b.want)
		}
		total += len(b.paths)
		for _, name := range b.paths {
			if _, err := os.Stat(filepath.Join(root, "docs", "archive", "closure", name)); err != nil {
				t.Errorf("%s (%s) not archived: %v", name, b.name, err)
			}
		}
	}
	if total != 37 {
		t.Errorf("frozen D1-D3 lists total %d paths, want 37", total)
	}
}
