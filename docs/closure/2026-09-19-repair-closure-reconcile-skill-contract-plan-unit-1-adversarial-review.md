---
mode: report-only
scope: "diff main..feat/repair-closure-and-reconcile-skill-contract-plan-unit-1"
shipment: 025-S
feature: 028-F
tasks: [028.001-T, 028.002-T, 028.003-T]
date: 2026-09-19
---

# Adversarial Review — Repair closure and reconcile skill contract (plan unit 1)

## Verdict

**READY**

No CRITICAL or MAJOR severity finding was raised by any reviewer. No HIGH-confidence
consensus finding, no MEDIUM-confidence majority/plurality finding, and no P0/P1 finding
was identified. All findings are MINOR severity, LOW confidence (each flagged by exactly
one reviewer), and are advisory only — none block merge.

## Methodology Note (deviation from standard protocol)

`git diff main..feat/repair-closure-and-reconcile-skill-contract-plan-unit-1` could not be
executed in this session — no direct shell/CLI tool was reachable from this agent or from
any dispatched subagent (all attempts, across three agent types, returned a sub-agent-depth
/ no-shell-tool refusal). The review was instead conducted by:

1. Reading the full current (post-change) content of all three affected files directly via
   the file-view tool (the working tree's `HEAD` is already
   `feat/repair-closure-and-reconcile-skill-contract-plan-unit-1`).
2. Cross-referencing that content against the user-supplied task description (028.001-T,
   028.002-T, 028.003-T), which specifies the pre-change wording and the exact original
   line numbers (1043, 1168, 503) the description claims for the three
   `src/autoharness/**` citations in `shipment-reconcile/SKILL.md`.
3. Independently grep/full-file-verifying that exactly three `src/autoharness` occurrences
   exist in `shipment-reconcile/SKILL.md` today, at lines matching the claimed locations,
   with the two claimed-edited ones bearing the new "authoritative / optional upstream
   reference / may be absent" wording and the one claimed-untouched one retaining plain
   citation wording consistent with pre-repair phrasing.
4. Independently checking `tests/integration/ship_feature_completion_contract_test.go`
   (the only test found to read `shipment-reconcile/SKILL.md`) for substring assertions
   that could collide with the reworded citations — none were found to overlap.

This substitutes for, but does not exactly reproduce, a literal diff. Reviewers were told
to treat the task description as authoritative for "what changed since main" and to focus
on internal correctness/consistency of the current state plus scope discipline. This is
disclosed as a methodology limitation, not a masked gap: the three reviewers independently
corroborate (a) the test's parsing logic is genuinely file-driven (no hardcoded target
pattern for the documented path itself), and (b) no reviewer found any semantic problem,
contract collision, or scope violation in the SKILL.md wording changes or the byte-unchanged
claim for line 503.

## Reviewer Pool

3 reviewers dispatched in parallel (no anchor route configured/available in this session;
used the anchor-unavailable 3-reviewer mapping):

| Slot | Route | Model | Result |
|---|---|---|---|
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | `[]` — no findings |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` | 1 MINOR finding |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | 8 MINOR findings |

All 3 reviewers completed before aggregation began.

## Consensus Findings (HIGH confidence — flagged by all 3 reviewers)

None.

## Majority Findings (MEDIUM confidence — flagged by >1.5 reviewers, i.e. all 3 or a strict majority)

None.

## Plurality Findings (MEDIUM confidence — flagged by 2 of 3 reviewers, not a strict majority)

None. (Reviewer-B and Reviewer-C both examined `placeholderClass`'s `id` branch at the same
code location, but reached **opposite** conclusions — B says the class is too permissive, C
says it is too narrow — so this is recorded as two independent LOW-confidence unique
observations below, not a genuine duplicate finding.)

## Unique Findings (LOW confidence — flagged by exactly 1 reviewer)

All findings below target `tests/integration/operational_closure_post_merge_filename_test.go`.
None target the `SKILL.md` wording changes (028.001-T/028.003-T); no reviewer raised a
concern there.

| # | Severity | Reviewer | Rule | Line | Issue | Fix |
|---|---|---|---|---|---|---|
| 1 | MINOR | B (Tier 2) | test-rigor-permissive-regex | ~96 | `placeholderClass`'s `id` branch compiles to a permissive class (`[0-9]+(?:\.[0-9]+)*-[A-Za-z]+`) for both `{shipment_id}` and `{feature_id}`, loose enough to accept non-conforming id shapes as "conformant" (e.g. mixed-case/multi-letter suffixes, reversed segment lengths) | Tighten to the real backlogit id grammar or add a companion negative-case assertion to catch id-format drift |
| 2 | MINOR | C (Tier 3) | vacuous-assertion-on-unknown-placeholder | ~112 | `placeholderClass`'s default arm (`[^/]+`) for any token other than `YYYY-MM-DD`/`slug`/`id` can, if the documented form is reworded to unrecognized token names, produce a regex that matches everything the glob already returns — silently turning the test into a tautology instead of failing loudly | Make unknown tokens fatal (`t.Fatalf` on unrecognized token) instead of silently falling back to a permissive default |
| 3 | MINOR | C (Tier 3) | heading-match-not-anchored | ~49 | `strings.Index(skillMD, "## Output")` is not anchored to a line start, so it would incorrectly match inside a deeper heading like `### Output ...` if such text is ever added elsewhere in the file; currently safe, but silently fragile | Anchor the search on a line-start match (e.g. `\n## Output\n` or per-line `TrimSpace(line) == "## Output"`) |
| 4 | MINOR | C (Tier 3) | parse-rule-keyword-ambiguity | ~65 | `findBacktickPathOnLineContaining` takes the *first* `## Output` line merely mentioning "post-merge" carrying any `docs/closure/...` path; today only one such line exists, but the parse rule doesn't defend against a second post-merge-mentioning line ever appearing with its own path | Require the matched line to also look like an Output-form bullet (e.g. contains "Closure artifact at"), and fail on more than one candidate rather than silently taking the first |
| 5 | MINOR | C (Tier 3) | comment-implementation-mismatch | ~11 | Header comment claims the test "never hardcodes the pattern", but the artifact universe is selected via a hardcoded glob `*-post-merge-closure.md`, which does hardcode the literal suffix of the convention under test | Soften the comment's claim scope, or derive the glob from the parsed template's literal tail after the last placeholder |
| 6 | MINOR | C (Tier 3) | id-placeholder-class-too-narrow | ~110 | The same `id` class discussed in finding #1 (opposite framing): it mandates a numeric-prefix + hyphen + purely-alphabetic-suffix shape, which would false-negative on a legitimate future variant (e.g. a multi-feature closure filename), misattributing a test-assumption gap to the artifact | Document the assumed id shape explicitly in the comment, or widen the class and treat multi-id filenames as an explicit, separately reported case |
| 7 | MINOR | C (Tier 3) | coverage-gap-generic-forms-unguarded | ~150 | Nothing in the new test asserts that the pre-merge/post-deploy generic Output form (the "leave other modes intact" half of 028.001-T) is still documented; deleting it would leave the new test fully green | Add an assertion that the `## Output` section still documents the generic `{YYYY-MM-DD}-{slug}-closure.md` form |
| 8 | MINOR | C (Tier 3) | environment-coupled-hard-failure | ~163 | An empty glob match is a `t.Fatalf`; if `docs/closure/` is ever legitimately empty of `*-post-merge-closure.md` artifacts in some workspace state, the suite fails for a reason unrelated to doc/artifact drift | Keep the anti-vacuous-pass guard but state the invariant explicitly, or `t.Skip` with a clear message when the universe is empty |
| 9 | MINOR | C (Tier 3) | shared-test-package-namespace-hygiene | ~45 | New package-level identifiers (`extractOutputSection`, `placeholderClass`, `compileClosureFilenameRegex`, etc.) are added to the shared `integration` package with no current collision, but generic names are likely to be reused by a future doc-conformance test, risking a redeclaration compile error | Prefix helpers/vars with a domain-specific name, e.g. `closureExtractOutputSection` |

## Task-by-Task Assessment

### 028.001-T — operational-closure/SKILL.md Output section
No reviewer flagged any issue. The post-merge form is additive and explicit; the
pre-merge/post-deploy generic form is left intact and adjacent. Semantically sound.

### 028.002-T — new parsing conformance test
Correctness of the core claim ("parses from file, does not hardcode the target pattern")
holds: `parsePostMergeClosurePathTemplate` reads the template string from the live
`## Output` section text, and `compileClosureFilenameRegex` derives its match regex from
that parsed template via `placeholderTokenPattern`/`placeholderClass`, rather than embedding
a fixed target string anywhere. All 9 unique findings above are edge-case robustness/rigor
observations about the parsing and regex-construction logic (LOW confidence, MINOR severity)
— none identify an actual false pass/fail on the current repository state, and none were
corroborated by a second reviewer with the same conclusion.

### 028.003-T — shipment-reconcile/SKILL.md wording changes
No reviewer flagged any semantic, consistency, or contract-collision issue with the two
reworded `src/autoharness/**` citations, or with the claim that the third citation
(original line 503) remains byte-unchanged. Independent verification (this agent, not a
reviewer) confirmed:
* Exactly 3 occurrences of `src/autoharness` exist in the file.
* The two reworded citations (Mixed-Role Detection Audit + Telemetry section;
  Related Artifacts section) both now carry the "installed skill markdown is authoritative
  / optional upstream reference / may be absent in other workspaces" pattern.
* The third citation (Step 0(c), Safe-Close Mode) retains plain, unhedged citation wording
  consistent with the pre-repair description — no reviewer or independent check found any
  divergence suggesting it was touched.
* `tests/integration/ship_feature_completion_contract_test.go` — the only test found that
  reads `shipment-reconcile/SKILL.md` — asserts exclusively on `classify-close-path` mode
  mechanics substrings that do not overlap with any of the three citation excerpts;
  no wording change in this task's scope collides with that test's contract.

### Scope discipline (P-021)
No reviewer identified scope expansion beyond the three described tasks. All findings are
confined to robustness/rigor observations within 028.002-T's own new test file.

## Remediation Plan

All 9 unique findings carry the same priority score (confidence_weight LOW=1 ×
severity_weight MINOR=2 = **2**) and the same action class:

| Priority | Finding # | File | Action Class | Notes |
|---|---|---|---|---|
| 2 | 1–9 | `tests/integration/operational_closure_post_merge_filename_test.go` | `advisory` | Optional hardening of the new conformance test's edge-case handling; none block merge; human judgment on which (if any) to act on |

No `safe_auto`, `gated_auto`, or `manual` entries are required — there is no CRITICAL/MAJOR
or P0/P1 finding in this review.

## Backlog Work Item Entries (P0/P1 only)

None. No P0 or P1 finding was raised by any reviewer.

## Post-Remediation Re-Review

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: "skipped"
```

Skipped per `mode: report-only` (user-requested) and because no `safe_auto` fix was
generated (no CRITICAL/MAJOR/P0/P1 finding exists to auto-remediate).

## Summary

**READY** — merge is not blocked. Optional follow-up: consider addressing findings #2–#9
(unknown-placeholder fatality, heading anchoring, parse-rule line disambiguation, glob/pattern
comment precision, generic-form regression guard, empty-universe failure mode, and package
namespace hygiene) as a low-priority test-hardening follow-up in a future task; none are
required before merging shipment 025-S / feature 028-F plan unit 1.
