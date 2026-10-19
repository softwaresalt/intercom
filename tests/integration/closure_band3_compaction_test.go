// Frozen snapshot guard for the one-time 055-S/063-F closure compaction.
// Delete or supersede it when a later compaction re-bands or archives these records.

package integration

import "testing"

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
	assertClosureBandArchived(t, repoRoot(t), closureBand3Summary, closureBand3ArchivedPaths)
}
