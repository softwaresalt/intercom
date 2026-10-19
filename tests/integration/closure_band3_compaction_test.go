package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// closureBand3ArchivedPaths is the frozen D3 candidate list (063.003-T) from
// docs/plans/2026-10-18-intercom-go-stage-batch-bcd-plan.md, Appendix D3.
var closureBand3ArchivedPaths = []string{
	"021-S-022-F-post-merge-closure.md",
	"2026-09-16-021-s-022-f-ship-covering-feature-completion-runtime-verification.md",
	"017-S-018-F-post-merge-closure.md",
	"024-S-027-F-post-merge-closure.md",
	"2026-09-19-024-s-027-f-staging-artifact-handoff-actor-runtime-verification.md",
	"2026-09-19-repair-closure-reconcile-skill-contract-plan-unit-1-adversarial-review.md",
	"025-S-028-F-post-merge-closure.md",
	"026-S-029-F-post-merge-closure.md",
	"027-S-030-F-post-merge-closure.md",
	"2026-09-20-026-s-029-f-repair-ship-agent-execution-contract-runtime-verification.md",
}

const closureBand3Summary = "2026-09-20-closure-compacted-band3-closure-summary.md"

// TestClosureBand3CompactionArchivesRecordsAndIndexesThem verifies the D3
// closure band is archived (not deleted) and indexed by one summary record.
func TestClosureBand3CompactionArchivesRecordsAndIndexesThem(t *testing.T) {
	root := repoRoot(t)

	summaryPath := filepath.Join(root, "docs", "closure", closureBand3Summary)
	summaryBytes, summaryErr := os.ReadFile(summaryPath)
	if summaryErr != nil {
		t.Errorf("summary %s missing: %v", closureBand3Summary, summaryErr)
	}
	summary := string(summaryBytes)

	for _, name := range closureBand3ArchivedPaths {
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

// TestClosureCompactionArchivesAll37Records verifies the D1-D3 bands together
// account for 37 archived closure records.
func TestClosureCompactionArchivesAll37Records(t *testing.T) {
	root := repoRoot(t)

	var all []string
	all = append(all, closureBand1ArchivedPaths...)
	all = append(all, closureBand2ArchivedPaths...)
	all = append(all, closureBand3ArchivedPaths...)

	if len(all) != 37 {
		t.Fatalf("frozen D1-D3 lists total %d paths, want 37", len(all))
	}

	archived := 0
	for _, name := range all {
		if _, err := os.Stat(filepath.Join(root, "docs", "archive", "closure", name)); err == nil {
			archived++
		}
	}
	if archived != 37 {
		t.Errorf("archived closure records = %d, want 37", archived)
	}
}
