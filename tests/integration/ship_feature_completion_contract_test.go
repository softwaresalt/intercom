package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Supplemental 33-row static contract test, per
// docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
// §6/§6.1. Every PRESENT/ABSENT needle below is a verbatim substring cut from
// the installed instruction text (the plan's pinned replacement wording, or,
// for guard/anchor rows, the pre-existing installed text) -- never authored
// from prose or intent, per the plan's Derivation rule.
//
// The 053-F expand/contract migration is complete, so this file is STRICT:
// every row pins the new contract text only. Rows that retired contract text
// (classify-close-path, CLASSIFICATION_BINDING, classification_binding,
// RECONCILE_FAIL_CLASSIFICATION, RECONCILE_FAIL_CASCADE_UNBOUND, CLOSE_PATH_VERDICT,
// fully-covered-root and their refusal wording) assert that the retired token is
// ABSENT from the surface that carried it. The policy version-history table keeps
// fully-covered-root verbatim, so the policy absence check skips those history
// rows (see liveLines).
//
// Per Stage deliberation D-2 (001-DL), this static needle test is
// SUPPLEMENTAL evidence only: it proves the instruction/skill prose states the
// required contract and no longer states the retired one. It cannot by itself
// satisfy the close-path invariants IV-1..IV-5; those rest on behavioral
// evidence of the cascade-close command, not on prose.
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

// assertEitherForm fails only when neither the old contract nor the new
// contract holds, so an either-form row cannot pass vacuously.
func assertEitherForm(t *testing.T, row, oldDesc string, oldHolds bool, newDesc string, newHolds bool) {
	t.Helper()
	if !oldHolds && !newHolds {
		t.Fatalf("row %s: neither old contract (%s) nor new contract (%s) holds", row, oldDesc, newDesc)
	}
}

func assertPresentEither(t *testing.T, content, oldNeedle, newNeedle, row string) {
	t.Helper()
	assertEitherForm(t, row,
		fmt.Sprintf("present %q", oldNeedle), countOccurrences(content, oldNeedle) >= 1,
		fmt.Sprintf("present %q", newNeedle), countOccurrences(content, newNeedle) >= 1)
}

func assertPresentExactlyEither(t *testing.T, content, oldNeedle, newNeedle, row string) {
	t.Helper()
	assertEitherForm(t, row,
		fmt.Sprintf("exactly once %q", oldNeedle), countOccurrences(content, oldNeedle) == 1,
		fmt.Sprintf("exactly once %q", newNeedle), countOccurrences(content, newNeedle) == 1)
}

func assertPresentAtLeastEither(t *testing.T, content, oldNeedle, newNeedle string, n int, row string) {
	t.Helper()
	assertEitherForm(t, row,
		fmt.Sprintf("present >=%d times %q", n, oldNeedle), countOccurrences(content, oldNeedle) >= n,
		fmt.Sprintf("present >=%d times %q", n, newNeedle), countOccurrences(content, newNeedle) >= n)
}

// assertAbsentEither passes when the retired needle is absent or when the
// replacement marker is present. Used where the replacement contract still
// contains the retired needle verbatim, so plain absence cannot hold.
func assertAbsentEither(t *testing.T, content, oldNeedle, newMarker, row string) {
	t.Helper()
	assertEitherForm(t, row,
		fmt.Sprintf("absent %q", oldNeedle), countOccurrences(content, oldNeedle) == 0,
		fmt.Sprintf("present %q", newMarker), countOccurrences(content, newMarker) >= 1)
}

// assertAbsentPattern fails when pattern matches anywhere in content.
func assertAbsentPattern(t *testing.T, content, pattern, row string) {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("row %s: invalid pattern %q: %v", row, pattern, err)
	}
	if loc := re.FindStringIndex(content); loc != nil {
		t.Fatalf("row %s: expected pattern to be ABSENT, but it matched at byte %d: %q", row, loc[0], pattern)
	}
}

// liveLines returns content without the policy version-history table rows
// (lines beginning "| 1."). Those rows keep retired tokens verbatim as history,
// so absence checks against live policy text must skip them.
func liveLines(content string) string {
	var b strings.Builder
	for _, line := range strings.SplitAfter(content, "\n") {
		if strings.HasPrefix(line, "| 1.") {
			continue
		}
		b.WriteString(line)
	}
	return b.String()
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
		assertPresent(t, ship, "`autoharness shipment cascade-close --classify-only --shipment {shipment_id}", "4")
		assertAbsent(t, ship, "classify-close-path", "4")
	})

	t.Run("row_5_CS1_p015_merge_commit", func(t *testing.T) {
		assertPresent(t, ship, "Mandatory pre-self-close context reload", "5")
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
		assertAbsent(t, ship, "`backlogit move <shipment_id> --status", "10")
		assertPresent(t, ship, "`shipment-reconcile` is authoritative", "10")
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
		assertPresent(t, ship, "must not be re-derived here", "19")
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
		assertPresent(t, skill, "`--classify-only` first, on every close.", "21a")
		assertAbsent(t, skill, "digest returned by `mode: classify-close-path`", "21a")
	})
	t.Run("row_21b_CS12_three_verdict_tokens", func(t *testing.T) {
		assertPresent(t, skill, "0 (`CASCADE`)", "21b")
		assertPresent(t, skill, "3 (`SAFE_CLOSE` selected: classifier `SAFE_CLOSE`, or engine `UNVERIFIED`)", "21b")
		assertPresent(t, skill, "HALT. Nothing was mutated.", "21b")
		assertAbsent(t, skill, "CLOSE_PATH_VERDICT", "21b")
	})
	t.Run("row_21c_CS11_classify_close_path_count", func(t *testing.T) {
		assertPresentAtLeast(t, skill, "cascade-close --classify-only", 2, "21c")
		assertAbsent(t, skill, "classify-close-path", "21c")
	})

	t.Run("row_22a_CS12_section_heading", func(t *testing.T) {
		assertPresent(t, skill, "**Command routing (192-F): `--classify-only` first, on every close.**", "22a")
		assertAbsent(t, skill, "### Classify-Close-Path Mode", "22a")
	})
	t.Run("row_22b_CS12_no_mutation_clause", func(t *testing.T) {
		assertPresent(t, skill, "Classification is read-only", "22b")
		assertAbsent(t, skill, "performs no archive, no status transition and no record mutation", "22b")
	})
	t.Run("row_22c_CS12_failclosed_block_clause", func(t *testing.T) {
		assertPresent(t, skill, "any other or unrecognized exit (including 1)", "22c")
		assertAbsent(t, skill, "returns `CLOSE_PATH_VERDICT: BLOCK`", "22c")
	})

	t.Run("row_23a_CS13_bound_refusal_and_branch_selects", func(t *testing.T) {
		assertPresent(t, skill, "HALT. Nothing was mutated.", "23a")
		assertPresent(t, skill, "Route on the command's exit code", "23a")
		assertAbsent(t, skill, "RECONCILE_FAIL_CLASSIFICATION", "23a")
		assertAbsent(t, skill, "the bound verdict alone selects the branch", "23a")
	})
	t.Run("row_23b_CS13_default_present_bound_only_absent", func(t *testing.T) {
		assertPresent(t, skill, "* **SAFE_CLOSE selected** (default, including any classifier error,", "23b")
		assertAbsent(t, skill, "(bound verdict `SAFE_CLOSE` only)", "23b")
	})
	t.Run("row_23c_CS13_unbound_D1_refusal", func(t *testing.T) {
		assertPresent(t, skill, "HALT. Nothing was mutated.", "23c")
		assertPresent(t, skill, "Nothing was mutated", "23c")
		assertAbsent(t, skill, "RECONCILE_FAIL_CASCADE_UNBOUND", "23c")
		assertAbsent(t, skill, "performs zero archive, zero status transition and zero record mutation", "23c")
	})
	t.Run("row_23d_CS13_invalid_binding_refusal", func(t *testing.T) {
		assertPresent(t, skill, "HALT. Nothing was mutated.", "23d")
		assertAbsent(t, skill, "RECONCILE_FAIL_CLASSIFICATION_INVALID", "23d")
	})

	t.Run("row_24a_CS14_drift_token_and_readonly", func(t *testing.T) {
		assertPresent(t, skill, "any difference from the `cascade`-selected `--classify-only` record", "24a")
		assertPresent(t, skill, "Classification is read-only", "24a")
		assertAbsent(t, skill, "RECONCILE_FAIL_CLASSIFICATION_DRIFT", "24a")
		assertAbsent(t, skill, "**`mode: classify-close-path` is strictly READ-ONLY.**", "24a")
	})
	t.Run("row_24b_guard_four_mode_names_survive", func(t *testing.T) {
		assertPresentAtLeast(t, skill, "mode: detect-mixed-role", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: safe-close", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: pre", 1, "24b")
		assertPresentAtLeast(t, skill, "mode: post", 1, "24b")
	})

	t.Run("row_25_CS3_unknown_token_halt", func(t *testing.T) {
		assertPresent(t, ship, "exit-code routing table lives in the", "25")
		assertAbsent(t, ship, "HALT on any result token the installed classification did not name", "25")
	})

	t.Run("row_26_CS5_binding_carried", func(t *testing.T) {
		assertAbsent(t, ship, "in place of the safe-close sequence above for this shipment's", "26")
		assertAbsent(t, ship, "CLASSIFICATION_BINDING", "26")
		assertPresent(t, ship, "`close_evidence` (the evidence record path)", "26")
	})

	t.Run("row_27_CS6_never_unbound", func(t *testing.T) {
		assertAbsent(t, ship, "classification_binding", "27")
		assertPresent(t, ship, "whose closure the closure-evidence gate refuses.", "27")
	})

	t.Run("row_28_CS10_self_contained_in_policy", func(t *testing.T) {
		assertPresent(t, policy, "the descendant graph union the feature is set-equal to the manifest", "28")
		assertPresent(t, policy, "scoped to the manifest of the shipment this session has claimed and whose live status is exactly `active`", "28")
		assertAbsent(t, liveLines(policy), "fully-covered-root", "28")
	})

	t.Run("row_29_CS15_intake_before_claim", func(t *testing.T) {
		const intakeNeedle = "Intake reconciliation check — run BEFORE the claim"
		const claimNeedle = "Record `shipment_id` as the session scope"
		assertPresent(t, ship, intakeNeedle, "29")
		assertOrderedBefore(t, ship, intakeNeedle, claimNeedle, "29")
	})

	t.Run("row_30_CS12_cascade_within_section_block", func(t *testing.T) {
		const sectionHeading = "### Step 0(c) Command Routing: cascade-close --classify-only (192-F)"
		const sectionEnd = "<!-- cascade-close-routing:END step-0c -->"
		startIdx := strings.Index(skill, sectionHeading)
		if startIdx < 0 {
			t.Fatalf("row 30: section heading not found: %q", sectionHeading)
		}
		endIdx := strings.Index(skill[startIdx:], sectionEnd)
		if endIdx < 0 {
			t.Fatalf("row 30: could not bound section block: %q not found after heading", sectionEnd)
		}
		block := skill[startIdx : startIdx+endIdx]
		assertPresent(t, block, "0 (`CASCADE`)", "30")
		assertAbsent(t, skill, "### Classify-Close-Path Mode", "30")
		assertAbsent(t, skill, "CLOSE_PATH_VERDICT: CASCADE", "30")
	})

	t.Run("row_31_CS16_binding_carried_old_literal_gone", func(t *testing.T) {
		assertAbsent(t, ship, "skill with `mode: safe-close`, `shipment_id`, and the", "31")
		assertAbsent(t, ship, "carrying the `classification_binding` from the preceding", "31")
		assertPresent(t, ship, "every close, on either path, starts", "31")
		assertPresent(t, ship, "whose closure the closure-evidence gate refuses.", "31")
	})
	t.Run("row_32_CS6_handles_every_refusal_token", func(t *testing.T) {
		assertPresent(t, ship, "exit-code routing table lives in the", "32")
		assertAbsent(t, ship, "RECONCILE_FAIL_CASCADE_UNBOUND", "32")
		assertAbsent(t, ship, "as a HALT with no mutation performed", "32")
	})
	t.Run("row_33_CS13_migration_no_success_shaped_fallback", func(t *testing.T) {
		assertPresent(t, skill, "never substitute `SAFE_CLOSE`", "33")
		assertPresent(t, skill, "there is no hand-run fallback", "33")
		assertPresent(t, skill, "do not fall back to cascade", "33")
		assertAbsent(t, skill, "Unbound invocations are rejected, not defaulted", "33")
		assertAbsent(t, skill, "refused, never served on a legacy path", "33")
		assertAbsent(t, skill, "no deprecation window, no grace period and no success-shaped fallback", "33")
	})

	t.Run("guard_no_018_dot_in_ship", func(t *testing.T) {
		assertAbsent(t, ship, "018.", "guard")
	})
}

// TestCascadeRoutesThroughCommand pins the post-migration close-path routing
// (053-F, plan U11). CASCADE routes only through the mutating
// `autoharness shipment cascade-close` command, after explicit operator
// approval for that closure; the direct `backlogit shipment ship` route stays
// forbidden. The a0 lifecycle gate, the a1 covering-feature completion gate and
// the 6.0 Post-Merge Branch Protocol heading must remain present. No retired
// classification token may appear in the Ship agent, and the 029.005-T
// "invoke the cascade" opener stays absent (the "invoke the cascade-close" form
// is allowed).
func TestCascadeRoutesThroughCommand(t *testing.T) {
	files := loadInstructionSurfaces(t)
	ship := files["ship"]

	assertPresent(t, ship, "obtain explicit operator approval for this specific closure **before** the mutating run. Only then run `autoharness shipment cascade-close` with no `--classify-only`", "U11-a")
	assertPresent(t, ship, "never calls `backlogit shipment ship` / `backlogit_ship_shipment` directly, for any verdict", "U11-a")

	assertPresent(t, ship, "a0. **TOPOLOGY_GATE: lifecycle", "U11-b")
	assertPresent(t, ship, "a1. **Covering-feature completion gate**", "U11-b")
	assertPresent(t, ship, "#### Step 6.0: Post-Merge Branch Protocol", "U11-b")

	for _, retired := range []string{"classify-close-path", "CLASSIFICATION_BINDING", "classification_binding"} {
		assertAbsent(t, ship, retired, "U11-c")
	}
	assertAbsentPattern(t, ship, `invoke the cascade([^-]|$)`, "U11-c")
}
