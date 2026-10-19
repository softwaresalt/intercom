package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// closureBand1ArchivedPaths is the frozen D1 candidate list (063.001-T) from
// docs/plans/2026-10-18-intercom-go-stage-batch-bcd-plan.md, Appendix D1.
var closureBand1ArchivedPaths = []string{
	"001-S-001-F-post-merge-closure.md",
	"002-S-002-F-post-merge-closure.md",
	"003-S-003-F-post-merge-closure.md",
	"004-S-005-F-post-merge-closure.md",
	"005-S-006-F-post-merge-closure.md",
	"2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md",
	"006-S-007-F-post-merge-closure.md",
	"007-S-008-F-post-merge-closure.md",
	"008-S-009-F-post-merge-closure.md",
	"009-S-010-F-post-merge-closure.md",
	"2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md",
	"010-S-011-F-post-merge-closure.md",
	"011-S-012-F-post-merge-closure.md",
	"2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md",
}

const closureBand1Summary = "2026-09-07-closure-compacted-band1-closure-summary.md"

// TestClosureBand1CompactionArchivesRecordsAndIndexesThem verifies the D1
// closure band is archived (not deleted) and indexed by one summary record.
func TestClosureBand1CompactionArchivesRecordsAndIndexesThem(t *testing.T) {
	root := repoRoot(t)

	summaryPath := filepath.Join(root, "docs", "closure", closureBand1Summary)
	summaryBytes, summaryErr := os.ReadFile(summaryPath)
	if summaryErr != nil {
		t.Errorf("summary %s missing: %v", closureBand1Summary, summaryErr)
	}
	summary := string(summaryBytes)

	for _, name := range closureBand1ArchivedPaths {
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
