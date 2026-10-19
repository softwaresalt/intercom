package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// closureBand2ArchivedPaths is the frozen D2 candidate list (063.002-T) from
// docs/plans/2026-10-18-intercom-go-stage-batch-bcd-plan.md, Appendix D2.
var closureBand2ArchivedPaths = []string{
	"012-S-013-F-post-merge-closure.md",
	"013-S-014-F-post-merge-closure.md",
	"014-S-015-F-post-merge-closure.md",
	"2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md",
	"2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md",
	"2026-09-08-013-s-014-f-pathsafe-reparse-containment-runtime-verification.md",
	"2026-09-08-014-s-015-f-retired-arch-gate-hardening-runtime-verification.md",
	"2026-09-08-retired-arch-gate-hardening-adversarial-review.md",
	"015-S-016-F-post-merge-closure.md",
	"016-S-017-F-post-merge-closure.md",
	"2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-adversarial-review.md",
	"2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-runtime-verification.md",
	"2026-09-11-016-s-017-f-pathsafe-error-surface-parity-runtime-verification.md",
}

const closureBand2Summary = "2026-09-10-closure-compacted-band2-closure-summary.md"

// TestClosureBand2CompactionArchivesRecordsAndIndexesThem verifies the D2
// closure band is archived (not deleted) and indexed by one summary record.
func TestClosureBand2CompactionArchivesRecordsAndIndexesThem(t *testing.T) {
	root := repoRoot(t)

	summaryPath := filepath.Join(root, "docs", "closure", closureBand2Summary)
	summaryBytes, summaryErr := os.ReadFile(summaryPath)
	if summaryErr != nil {
		t.Errorf("summary %s missing: %v", closureBand2Summary, summaryErr)
	}
	summary := string(summaryBytes)

	for _, name := range closureBand2ArchivedPaths {
		if _, err := os.Stat(filepath.Join(root, "docs", "archive", "closure", name)); err != nil {
			t.Errorf("%s not archived under docs/archive/closure: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(root, "docs", "closure", name)); err == nil {
			t.Errorf("%s still present in docs/closure", name)
		}
		if summaryErr == nil && !strings.Contains(summary, name) {
			t.Errorf("summary does not list %s", name)
		}
	}
}
