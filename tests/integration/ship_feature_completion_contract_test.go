package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Supplemental 33-row static contract test, per
// docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
// §6/§6.1. Every PRESENT/ABSENT needle below is a verbatim substring cut from
// the plan's §4.2 pinned replacement wording (or, for guard/anchor rows, from
// the plan's own H0 measurement of the pre-existing installed text) -- never
// authored from prose or intent, per the plan's Derivation rule.
//
// Per Stage deliberation D-2 (001-DL), this static needle test is
// SUPPLEMENTAL evidence only: it proves the instruction/skill prose says the
// right thing. It cannot by itself satisfy IV-1..IV-5; the fixture-backed
// behavioral tests in shipment_close_path_behavior_test.go are the
// authoritative acceptance surface for those invariants.
// ---------------------------------------------------------------------------

func loadInstructionSurfaces(t *testing.T) map[string]string {
	t.Helper()
	root := repoRoot(t)
	files := map[string]string{
		"ship":   filepath.Join(root, ".github", "agents", "_ship.agent.md"),
		"policy": filepath.Join(root, ".github", "policies", "workflow-policies.md"),
		"skill":  filepath.Join(root, ".github", "skills", "shipment-reconcile", "SKILL.md"),
	}
	out := make(map[string]string, len(files))
	for key, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading instruction surface %s (%s): %v", key, path, err)
		}
		out[key] = string(data)
	}
	return out
}

func countOccurrences(haystack, needle string) int {
	if needle == "" {
		return 0
	}
	count := 0
	idx := 0
	for {
		pos := strings.Index(haystack[idx:], needle)
		if pos < 0 {
			return count
		}
		count++
		idx += pos + len(needle)
	}
}

func assertPresent(t *testing.T, content, needle, row string) {
	t.Helper()
	if countOccurrences(content, needle) < 1 {
		t.Fatalf("row %s: expected needle to be PRESENT, but it is ABSENT: %q", row, needle)
	}
}

func assertPresentAtLeast(t *testing.T, content, needle string, n int, row string) {
	t.Helper()
	got := countOccurrences(content, needle)
	if got < n {
		t.Fatalf("row %s: expected needle present >=%d times, got %d: %q", row, n, got, needle)
	}
}

func assertPresentExactly(t *testing.T, content, needle string, row string) {
	t.Helper()
	got := countOccurrences(content, needle)
	if got != 1 {
		t.Fatalf("row %s: expected needle present exactly once, got %d: %q", row, got, needle)
	}
}

func assertAbsent(t *testing.T, content, needle, row string) {
	t.Helper()
	if got := countOccurrences(content, needle); got != 0 {
		t.Fatalf("row %s: expected needle to be ABSENT, but found %d occurrence(s): %q", row, got, needle)
	}
}

func assertOrderedBefore(t *testing.T, content, first, second, row string) {
	t.Helper()
	firstIdx := strings.Index(content, first)
	secondIdx := strings.Index(content, second)
	if firstIdx < 0 {
		t.Fatalf("row %s: ordering anchor not found: %q", row, first)
	}
	if secondIdx < 0 {
		t.Fatalf("row %s: ordering anchor not found: %q", row, second)
	}
	if firstIdx >= secondIdx {
		t.Fatalf("row %s: expected %q to appear before %q, but it did not", row, first, second)
	}
}

func TestShipFeatureCompletionContract33Rows(t *testing.T) {
	files := loadInstructionSurfaces(t)
	ship := files["ship"]
	policy := files["policy"]
	skill := files["skill"]

	t.Run("row_1_ADD1_grant_present", func(t *testing.T) {
		assertPresent(t, ship, "complete one covering feature `active -> done`", "1")
	})

	t.Run("row_2a_ADD1_conjunctive_failclosed", func(t *testing.T) {
		assertPresent(t, ship, "all five conditions below are conjunctive and fail-closed", "2a")
	})
	t.Run("row_2b_ADD1_every_depth", func(t *testing.T) {
		assertPresent(t, ship, "no live descendant remains at every depth", "2b")
	})
	t.Run("row_2c_ADD1_set_equality", func(t *testing.T) {
		assertPresent(t, ship, "the descendant graph union the feature is set-equal to the manifest", "2c")
	})
	t.Run("row_2d_ADD1_scoping_clause", func(t *testing.T) {
		assertPresent(t, ship, "scoped to the manifest of the shipment this session has claimed and whose live status is exactly `active`", "2d")
	})

	t.Run("row_3_ADD2_a1_ordering", func(t *testing.T) {
		const a0Anchor = "a0. **TOPOLOGY_GATE: lifecycle"
		const aAnchor = "a. **Pre-archive reconciliation gate"
		const a1Needle = "a1. **Covering-feature completion gate**"
		assertPresent(t, ship, a1Needle, "3")
		assertOrderedBefore(t, ship, a0Anchor, a1Needle, "3")
		assertOrderedBefore(t, ship, a1Needle, aAnchor, "3")
	})

	t.Run("row_4_CS4_delegation_boundary", func(t *testing.T) {
		assertPresent(t, ship, "the `shipment-reconcile` skill's read-only `mode: classify-close-path` boundary selects the close path", "4")
	})

	t.Run("row_5_CS1_p015_merge_commit", func(t *testing.T) {
		assertPresent(t, ship, "also re-read `P-015` and verify the merge commit", "5")
	})
	t.Run("row_6_CS1_authority_escalation", func(t *testing.T) {
		assertPresent(t, ship, "A context reload MUST NOT widen the authority envelope", "6")
	})

	t.Run("row_7_CS4_drifted_children_only_gone", func(t *testing.T) {
		assertAbsent(t, ship, "it is fully covered (every one of its children,", "7")
	})

	t.Run("row_8_guard_no_TASK_ONLY_FINALIZE", func(t *testing.T) {
		assertAbsent(t, ship, "TASK_ONLY_FINALIZE", "8")
	})

	t.Run("row_9a_CS4_predicate_absent", func(t *testing.T) {
		assertAbsent(t, ship, "permitted **only** when, for **every** feature member of the manifest: it is a", "9a")
	})
	t.Run("row_9b_CS4_predicate_absent", func(t *testing.T) {
		assertAbsent(t, ship, "enumerates to zero children, that childlessness is **positively verified**", "9b")
	})
	t.Run("row_9c_CS4_predicate_absent", func(t *testing.T) {
		assertAbsent(t, ship, "The manifest must contain nothing beyond the", "9c")
	})
	t.Run("row_9d_CS4_predicate_absent", func(t *testing.T) {
		assertAbsent(t, ship, "qualification is never per-member, and no feature ID is ever special-cased.", "9d")
	})

	t.Run("row_10_CS2_direct_status_gone_pointer_present", func(t *testing.T) {
		assertPresentExactly(t, ship, "<shipment_id> --status shipped", "10")
		assertPresent(t, ship, "the authoritative close sequence defined by the `shipment-reconcile` skill", "10")
	})

	t.Run("row_11_CS1_scoped_role_boundary_only", func(t *testing.T) {
		assertPresent(t, ship, "scoped to Role Boundary changes only", "11")
	})

	t.Run("row_12_ADD2_notapplicable_and_haltnaming", func(t *testing.T) {
		assertPresent(t, ship, "A1_NOT_APPLICABLE", "12")
		assertPresent(t, ship, "HALT, naming the unmet condition", "12")
	})

	t.Run("row_13_ADD2_already_done", func(t *testing.T) {
		assertPresent(t, ship, "A1_ALREADY_DONE", "13")
	})

	t.Run("row_14_CS7_compound", func(t *testing.T) {
		assertPresentExactly(t, ship, "every manifest item is present in `.backlogit/queue/` with the", "14")
		assertPresent(t, ship, "defers every per-item status decision to that skill's `mode: pre` classification", "14")
	})
	t.Run("row_15_CS8_compound", func(t *testing.T) {
		assertPresentExactly(t, ship, "queue with `status: done`, and scans for orphan items.", "15")
		assertPresent(t, ship, "defers every per-item status decision to the `mode: pre` classification at `expected_status: done`", "15")
	})

	t.Run("row_16a_ADD2_anomaly_before_n", func(t *testing.T) {
		assertPresent(t, ship, "evaluated before `n` is computed", "16a")
	})
	t.Run("row_16b_ADD2_multiple_feature_members", func(t *testing.T) {
		assertPresent(t, ship, "A1_MULTIPLE_FEATURE_MEMBERS", "16b")
		assertPresent(t, ship, "halt with no feature mutation", "16b")
	})

	t.Run("row_17_CS8_preserve_orphan_scan", func(t *testing.T) {
		assertPresentAtLeast(t, ship, "scans for orphan items", 2, "17")
	})
	t.Run("row_18_CS7_CS15_preserve_scope_note", func(t *testing.T) {
		assertPresentExactly(t, ship, "Scope note (139-F/139.001-T)", "18")
	})

	t.Run("row_19_CS9_scoped_to_safe_close", func(t *testing.T) {
		assertAbsent(t, ship, "proves the protected set and halts fail-closed on any cascade or provenance", "19")
		assertPresent(t, ship, "proves the protected set on the safe-close path", "19")
		assertPresent(t, ship, "no protected set by construction", "19")
	})

	t.Run("row_20a_CS10_grant_in_policy", func(t *testing.T) {
		assertPresent(t, policy, "complete one covering feature `active -> done`", "20a")
	})
	t.Run("row_20b_guard_no_018_dot_in_policy", func(t *testing.T) {
		assertAbsent(t, policy, "018.", "20b")
	})
	t.Run("row_20c_guard_no_release_ids_in_policy", func(t *testing.T) {
		assertAbsent(t, policy, "021-", "20c")
		assertAbsent(t, policy, "022-", "20c")
	})

	t.Run("row_21a_CS11_binding_input_row", func(t *testing.T) {
		assertPresent(t, skill, "digest returned by `mode: classify-close-path`", "21a")
	})
	t.Run("row_21b_CS12_three_verdict_tokens", func(t *testing.T) {
		assertPresent(t, skill, "CLOSE_PATH_VERDICT: CASCADE", "21b")
		assertPresent(t, skill, "CLOSE_PATH_VERDICT: SAFE_CLOSE", "21b")
		assertPresent(t, skill, "CLOSE_PATH_VERDICT: BLOCK", "21b")
	})
	t.Run("row_21c_CS11_classify_close_path_count", func(t *testing.T) {
		assertPresentAtLeast(t, skill, "classify-close-path", 2, "21c")
	})

	t.Run("row_22a_CS12_section_heading", func(t *testing.T) {
		assertPresent(t, skill, "### Classify-Close-Path Mode", "22a")
	})
	t.Run("row_22b_CS12_no_mutation_clause", func(t *testing.T) {
		assertPresent(t, skill, "performs no archive, no status transition and no record mutation", "22b")
	})
	t.Run("row_22c_CS12_failclosed_block_clause", func(t *testing.T) {
		assertPresent(t, skill, "every unsupported, mixed, ambiguous, torn, missing or incompletely enumerated case returns `CLOSE_PATH_VERDICT: BLOCK`", "22c")
	})

	t.Run("row_23a_CS13_bound_refusal_and_branch_selects", func(t *testing.T) {
		assertPresent(t, skill, "every other bound verdict halts before any close mutation with `RECONCILE_FAIL_CLASSIFICATION_REFUSED`", "23a")
		assertPresent(t, skill, "the bound verdict alone selects the branch", "23a")
	})
	t.Run("row_23b_CS13_default_gone_bound_only_present", func(t *testing.T) {
		assertAbsent(t, skill, "* **SAFE_CLOSE selected** (default, including any classifier error,", "23b")
		assertPresent(t, skill, "(bound verdict `SAFE_CLOSE` only)", "23b")
	})
	t.Run("row_23c_CS13_unbound_D1_refusal", func(t *testing.T) {
		assertPresent(t, skill, "halts before any close mutation with `RECONCILE_FAIL_CASCADE_UNBOUND`", "23c")
		assertPresent(t, skill, "performs zero archive, zero status transition and zero record mutation", "23c")
	})
	t.Run("row_23d_CS13_invalid_binding_refusal", func(t *testing.T) {
		assertPresent(t, skill, "halts before any close mutation with `RECONCILE_FAIL_CLASSIFICATION_INVALID`", "23d")
	})

	t.Run("row_24a_CS14_drift_token_and_readonly", func(t *testing.T) {
		assertPresent(t, skill, "RECONCILE_FAIL_CLASSIFICATION_DRIFT", "24a")
		assertPresent(t, skill, "**`mode: classify-close-path` is strictly READ-ONLY.**", "24a")
	})
	t.Run("row_24b_guard_four_mode_names_survive", func(t *testing.T) {
		assertPresentAtLeast(t, skill, "mode: detect-mixed-role", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: safe-close", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: pre", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: post", 1, "24b")
	})

	t.Run("row_25_CS3_unknown_token_halt", func(t *testing.T) {
		assertPresent(t, ship, "HALT on any result token the installed classification did not name", "25")
	})

	t.Run("row_26_CS5_binding_carried", func(t *testing.T) {
		assertAbsent(t, ship, "in place of the safe-close sequence above for this shipment's", "26")
		assertPresent(t, ship, "carrying the returned `CLASSIFICATION_BINDING` into that call", "26")
	})

	t.Run("row_27_CS6_never_unbound", func(t *testing.T) {
		assertPresent(t, ship, "Ship MUST NOT invoke `mode: safe-close` without a `classification_binding`", "27")
	})

	t.Run("row_28_CS10_self_contained_in_policy", func(t *testing.T) {
		assertPresent(t, policy, "the descendant graph union the feature is set-equal to the manifest", "28")
		assertPresent(t, policy, "scoped to the manifest of the shipment this session has claimed and whose live status is exactly `active`", "28")
	})

	t.Run("row_29_CS15_intake_before_claim", func(t *testing.T) {
		const intakeNeedle = "Intake reconciliation check — run BEFORE the claim"
		const claimNeedle = "Record `shipment_id` as the session scope"
		assertPresent(t, ship, intakeNeedle, "29")
		assertOrderedBefore(t, ship, intakeNeedle, claimNeedle, "29")
	})

	t.Run("row_30_CS12_cascade_within_section_block", func(t *testing.T) {
		const sectionHeading = "### Classify-Close-Path Mode"
		const nextSectionHeading = "### Safe-Close Mode"
		startIdx := strings.Index(skill, sectionHeading)
		if startIdx < 0 {
			t.Fatalf("row 30: section heading not found: %q", sectionHeading)
		}
		endIdx := strings.Index(skill[startIdx:], nextSectionHeading)
		if endIdx < 0 {
			t.Fatalf("row 30: could not bound section block: %q not found after heading", nextSectionHeading)
		}
		block := skill[startIdx : startIdx+endIdx]
		assertPresent(t, block, "CLOSE_PATH_VERDICT: CASCADE", "30")
	})

	t.Run("row_31_CS16_binding_carried_old_literal_gone", func(t *testing.T) {
		assertAbsent(t, ship, "skill with `mode: safe-close`, `shipment_id`, and the", "31")
		assertPresent(t, ship, "carrying the `classification_binding` from the preceding", "31")
		assertPresent(t, ship, "An unbound safe-close invocation is refused by the skill", "31")
	})

	t.Run("row_32_CS6_handles_every_refusal_token", func(t *testing.T) {
		assertPresent(t, ship, "RECONCILE_FAIL_CASCADE_UNBOUND", "32")
		assertPresent(t, ship, "as a HALT with no mutation performed", "32")
	})

	t.Run("row_33_CS13_migration_no_success_shaped_fallback", func(t *testing.T) {
		assertPresent(t, skill, "Unbound invocations are rejected, not defaulted", "33")
		assertPresent(t, skill, "refused, never served on a legacy path", "33")
		assertPresent(t, skill, "no deprecation window, no grace period and no success-shaped fallback", "33")
	})

	t.Run("guard_no_018_dot_in_ship", func(t *testing.T) {
		assertAbsent(t, ship, "018.", "guard")
	})
}

// TestCascadeRoutesThroughSafeClose is the 029.005-T contract test (plan
// docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md
// §3.3, PIN-830/PIN-861). Its single red->green assertion pair is cut
// VERBATIM from the pinned replacement wording per the binding compound rule
// in docs/compound/workflow-issues/cross-artifact-contract-closure-requires-
// every-surface-2026-09-13.md: needles are never authored from prose or
// intent.
//
// PRESENT (post-repair only): "designates `shipment-reconcile` `mode: safe-
// close`" -- introduced by PIN-830, absent from the pre-repair file.
// ABSENT (post-repair): "invoke the cascade" -- the pre-repair `:861` opener,
// replaced in full by PIN-861.
//
// This is deliberately the ONLY assertion in this test (plan §3.2: "SINGLE
// red-to-green assertion only"); the earlier green-on-arrival second
// assertion is intentionally not implemented here.
func TestCascadeRoutesThroughSafeClose(t *testing.T) {
	files := loadInstructionSurfaces(t)
	ship := files["ship"]

	assertPresent(t, ship, "designates `shipment-reconcile` `mode: safe-close`", "U2-T3")
	assertAbsent(t, ship, "invoke the cascade", "U2-T3")
}
