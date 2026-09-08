---
title: "012-S retire-the-D6a-gate-narrowing adversarial review"
description: "Multi-model adversarial review of the internal/config/** -> internal/** gate broadening, the retiredgo fixture suite, the config.toml.example TOML-dispatch P0 fix, and the D6c governance annotations"
mode: "report-only"
branch: "feat/012-s-retire-the-d6a-gate-narrowing-broaden-u-e1a-to-internal"
reviewers: 3
---

# 012-S retire-the-D6a-gate-narrowing adversarial review

## 0. Session tooling constraint (read first)

This review session had **no shell/bash/git tool access** available to it or
to any dispatched sub-agent (a hard sub-agent-depth limit blocked every
attempt to invoke `git`, `bash`, or the Go toolchain, including through
nested `task`/`general-purpose` agents). This was confirmed empirically
across five independent probe attempts before proceeding.

**Consequently, the following requested independent verifications could NOT
be executed and are UNVERIFIED by this review:**

* `git diff --stat main...HEAD` / `git diff main...HEAD` (the review instead
  relied on the requester's provided file list and diff summary, cross-checked
  against full `view`-based reads of every changed file's current, on-branch
  content)
* `bash scripts/check-retired-architecture.sh --self-test`
* `gofmt -l .`, `go vet ./...`, `go test ./...`, `go build ./...`
* `git log --oneline -- scripts/check-retired-architecture.sh` and
  `git show 6dec85a --stat` (the D6c commit-hash claim)
* Live `gitleaks` execution against the new fixtures

**What this review DID do instead (static verification, high confidence):**

* Read the complete current content of all 7 changed files, plus
  `internal/apperr/{apperr,sentinel}.go`, `internal/copilotprobe/{client,
  fixture,permission}.go`, `internal/pathsafe/{lexical,pathsafe,root}.go`,
  `config.toml.example`, `.github/workflows/ci.yml`, and the relevant
  constitution/Go/CI-security instruction files.
* Manually traced the Python scanning logic (regexes, state machines, dispatch
  functions) line-by-line rather than trusting behavior claims.
* Manually confirmed, by direct token search, that none of the files newly
  brought into scope by the `internal/config/**` → `internal/**` broadening
  contain any forbidden token as a Go identifier — i.e. the real repo scan
  would pass on the current tree even though the tool could not execute it.
* Dispatched 3 independent static-analysis-only reviewer agents (below) with
  the full file contents embedded directly in their prompts.

**Operator action required before merge:** run the self-test and full Go
quality-gate sequence for real (`bash scripts/check-retired-architecture.sh
--self-test`, `gofmt -l .`, `go vet ./...`, `go test ./...`, `go build ./...`)
and confirm commit `6dec85a`'s existence/content, since none of these were
observed directly in this session.

## 1. Model route assignment

| Reviewer | Route | Model | Notes |
|---|---|---|---|
| Anchor Reviewer | Anchor review route | `gpt-5.6-sol`, effort `high` | Dispatched successfully; persona = architecture + constitution + P-021 scope discipline |
| Reviewer-A (Tier 1) | Fast/cheap | `gpt-5.4-mini` | Persona = correctness + Go-specific + maintainability |
| Reviewer-C (Tier 3) | Frontier | `claude-opus-5` | Persona = security (gitleaks/evasion) + schema/CLI/docs coupling + concurrency |

`reviewers: 3` with a dispatchable anchor route maps to **Anchor + Reviewer-A
(Tier 1) + Reviewer-C (Tier 3)** per the skill's count-to-slot table. No
alternate-provider override was requested; no declared fallback was needed —
all three reviewer slots dispatched successfully on the first attempt.

## 2. Independently verified facts (not reviewer opinions)

| # | Item | Result |
|---|---|---|
| (a) | P0 TOML dispatch bug (`config.toml.example` suffix is `.example`, not `.toml`) | **Genuinely fixed.** `engine_for_path()` special-cases `path.name == 'config.toml.example'` *before* the suffix checks, and this is wired into both the real `scan_path()` production path and a dedicated self-test assertion (`dispatch config.toml.example`). Confirmed non-vacuous: the assertion would fail if the special case were removed. |
| (b) | `should_scan_repo_path` prefix-matching exact-segment safety | **Safe.** `path.startswith('internal/')` (trailing slash included) cannot false-positive on a sibling like `internalfoo/x.go`. `'/testdata/' in path` (slashes on both sides) cannot false-positive on `internal/mytestdatafoo/x.go`. No repo path today (`internal/{apperr,config,copilotprobe,pathsafe}`) exercises a genuine ambiguous case, so this is confirmed correct by construction, not by a repo-specific coincidence. |
| (c) | Real-tree compatibility of the broadened scope | **Confirmed clean.** Direct read of every non-test `.go` file newly brought into scope (`internal/apperr/{apperr,sentinel}.go`, `internal/copilotprobe/{client,fixture,permission}.go`, `internal/pathsafe/{lexical,pathsafe,root}.go`) shows zero occurrences of any forbidden token as a Go identifier. `internal/apperr`'s taxonomy is confirmed to be the corrected 11-variant set with no `KindSlack`/`KindIPC`/`KindACP` remaining, supporting D6c's stated reason. |
| (d) | Gitleaks/secret-scan exposure of the 2 new Go fixtures + 1 new JSON manifest | **Clean by inspection.** No `xox[baprs]-` prefix, no `hooks.slack.com`, no high-entropy string assigned to anything, no realistic-looking credential. The literal word "slack" appears only as a camelCase-decomposed identifier fragment (`SlackTeamID`) and in a comment — gitleaks' stock rules are prefix/entropy-keyed, not bare-keyword-keyed, so this should not trip a stock rule. **Not observed via a live gitleaks run** — recommend confirming the CI `security` job is green post-push. |
| (e) | `retiredgo-manifest.json` validity/sort order | Valid JSON, 2 keys, **sorted** ascending. (Pre-existing, out-of-scope `retired-manifest.json` for the TOML suite is *not* sorted — a pre-existing inconsistency, not introduced by this diff, not flagged as a defect by any reviewer.) |
| (f) | Exactly 2 Go fixtures, 1 accept / 1 reject | Confirmed: `retired-accept-comment-only.go` (accept), `retired-reject-slack-team-id.go` (reject). Matches task 013.003-T's constraint. |
| (g) | "Exactly two governance docs" constraint (task 006) | The **decision doc** (D6c amendment) and the **plan doc** (5 historical-note annotations) are the two files touched under that specific constraint. `docs/memory/012-s-013.001-red-evidence.md` is a distinct artifact type (RED-evidence capture) produced by an earlier task (013.001-T) in the same chain, not a governance doc — none of the three reviewers flagged this as a scope violation. Interpreted as compliant, but noting explicitly since the raw file count under `docs/` for this shipment is 3, not 2. |
| | 5 historical-note annotations preserve original content | Directly confirmed via `view` for 3 of 5 (I5 row, PA-4/risk-table area, final Post-Review Remediation Record) and cross-checked via a targeted sub-agent extraction for all 5 — every annotation is an *appended* blockquote immediately following (not replacing) original prose. No original content was deleted or rewritten. |
| | D6c section uniqueness | Appears exactly once in the deliberation doc, immediately after D6b, consistent with governing-doc convention. |
| (g) | Goroutine/concurrency concerns | **None.** Single Python process per `scan_with_mode` invocation, no threading/async, no shared mutable state across process boundaries. All three reviewers independently confirmed this. |
| (h) | Markdownlint cleanliness | **UNVERIFIED** — no markdownlint execution available in this session. |
| | D6c commit-hash claim (`6dec85a`) | **UNVERIFIED** — flagged by 2 of 3 reviewers as needing manual `git log`/`git show` confirmation before treating "introduced already narrowed, never broadened back" as an established fact in a governing decision document. |

## 3. Consensus findings (confidence: HIGH — flagged by all 3 reviewers)

### C-1. Self-test's "selection structural inclusion" assertion has partial vacuous-truth / empty-set-pass risk

* **Severity:** MAJOR (most conservative of the three characterizations — see below)
* **File:** `scripts/check-retired-architecture.sh`
* **Location:** `expected_internal_repo_paths()` / `run_repo_selection_self_test()` (`selection structural inclusion` assertion)
* **Issue:** All three reviewers independently converged on the same code region from different angles:
  * The Anchor Reviewer and Reviewer-A both noted that `expected_internal_repo_paths()` re-implements the *exact same* filter conditions (`internal/` prefix, `/testdata/` exclusion, `_test.go` exclusion, `.go` suffix) as `should_scan_repo_path()`, rather than deriving the expected set through a genuinely independent method. A bug shared by both implementations would not be caught by this assertion.
  * Reviewer-C (Tier 3) went further and identified a concrete, severe failure mode: because both `internal_actual` and `internal_expected` are derived from the *same* `git ls-files -- internal/**` pathspec call, if that call ever returns zero rows (e.g. `GIT_LITERAL_PATHSPECS=1`/`GIT_NOGLOB_PATHSPECS=1` in the environment causing glob patterns to match literally/nothing, or a sparse checkout), both sides are empty, `internal_actual == internal_expected` trivially holds, and the assertion **passes** while the gate silently scans zero internal files. The compensating `selection non-empty` check only asserts the *aggregate* `rel_paths` is non-empty, which `cmd/**` + `config.toml.example` alone can satisfy — this exact insensitivity is empirically visible in the committed RED evidence (`docs/memory/012-s-013.001-red-evidence.md`), which shows `PASS selection non-empty: selected 11 tracked repo paths` on the very run where 8 internal files were missing from selection.
* **Fix:** Add an explicit non-emptiness assertion scoped to `internal/` specifically (e.g. `len(internal_expected) > 0` and `len(internal_actual) > 0`), and consider deriving the expected set via a genuinely independent method (e.g. a hardcoded/floor count, or `os.walk` rather than reusing the same filter predicate) so a shared bug in the exclusion logic cannot make the assertion self-validate.
* **Action class:** `gated_auto` — the fix is a small, well-specified assertion addition; confirm before applying since it changes self-test pass/fail behavior.

## 4. Majority findings (confidence: MEDIUM — flagged by 2 of 3 reviewers)

### M-1. `config.toml.example`'s presence in the real selection set is asserted only in isolation, not end-to-end

* **Severity:** MINOR (both reviewers rated this MINOR)
* **Flagged by:** Reviewer-A (Tier 1), Reviewer-C (Tier 3)
* **File:** `scripts/check-retired-architecture.sh`
* **Issue:** The new `dispatch config.toml.example` self-test assertion proves `engine_for_path(Path('config.toml.example')) == 'toml'`, but nothing asserts that `'config.toml.example'` actually appears in `select_repo_paths()`'s real output. If the file were renamed, untracked, or dropped from the pathspec, the TOML engine would again never execute against the real tree — the same "pass produced by an engine that never runs" failure class as the P0 bug this shipment fixes — while this specific assertion would keep passing. Reviewer-A additionally notes `'config.toml.example'` is hardcoded as a special case in *two* places (`should_scan_repo_path` and `engine_for_path`), which is itself a duplication smell independent of the coverage gap.
* **Fix:** Add `report_assertion('selection includes config.toml.example', 'config.toml.example' in rel_paths, ...)` alongside the existing dispatch assertion.
* **Action class:** `gated_auto`.

### M-2. D6c's commit-hash claim (`6dec85a`) is unverifiable in this review and should be confirmed before merge

* **Severity:** MINOR (informational/documentation-accuracy)
* **Flagged by:** Anchor Reviewer, Reviewer-C (Tier 3)
* **File:** `docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`
* **Issue:** D6c states "git history shows the gate was introduced already narrowed in commit `6dec85a`," which is load-bearing for the amendment's "broadening for the first time, not a reversal" framing. Neither reviewer (nor the orchestrating review) had git access to confirm this commit exists or supports the claim.
* **Fix:** Manually run `git log -1 6dec85a --oneline -- scripts/check-retired-architecture.sh` and `git log -S"internal/**" -- scripts/check-retired-architecture.sh` (expected empty except the current change) before treating this as settled fact in a governing document.
* **Action class:** `manual` (documentation-accuracy verification, not a code fix).

## 5. Plurality findings

Not applicable at `reviewers: 3` — with an odd reviewer count, "flagged by
more than one but not a strict majority" has no distinct bucket (2 of 3 is
already a strict majority). See §4 for the 2-of-3 findings.

## 6. Unique findings (confidence: LOW — flagged by exactly one reviewer)

These are preserved as observations. Several describe **pre-existing**
weaknesses in code/CI structure not modified by this diff; they are valuable
backlog candidates, not regressions requiring a fix before this PR merges,
unless otherwise noted.

| # | Severity | Reviewer | File | Issue | Fix |
|---|---|---|---|---|---|
| U-1 | MAJOR | Anchor | `scripts/check-retired-architecture.sh` (`split_identifier`/`matches_forbidden_parts`) | A concatenated, no-separator, all-lowercase (or single-leading-capital) Go identifier such as a hypothetical `socketmode` or `Socketmode` decomposes to a single token `["socketmode"]` via `component_re`'s `[A-Z]?[a-z]+` alternative, which has length 1 and can never match the 2-element `forbidden_parts["socketmode"] = ["socket","mode"]` sequence. CamelCase forms (`SocketMode`) ARE caught; only the smashed-together form evades. **Pre-existing logic, not modified by this diff.** | Add a check against the full lowercased identifier string (not just split parts) for tokens with no internal camelCase boundary, or extend `split_identifier` to also emit sliding-window substrings for multi-word tokens. |
| U-2 | MAJOR | Reviewer-A (Tier 1) | `scripts/check-retired-architecture.sh` (`walk_toml_value`/`scan_toml_with_fallback`) | TOML key matching only checks `key.lower().split('_')` **per individual key segment**, so (i) dotted key paths like `host.cli = "..."` parse via `tomllib` into nested tables and each segment (`host`, `cli`) is checked independently, never as the joined 2-part sequence `["host","cli"]`; (ii) TOML's legal hyphenated bare keys like `team-id = "..."` never split on `_` and are checked as a single unsplit token; (iii) a concatenated key `socketmode = "..."` has the same single-token miss as U-1. All three are real, spec-legal TOML key forms that evade the gate. **Pre-existing logic, not modified by this diff** (this diff did not touch `walk_toml_value`/`scan_toml`/`matches_forbidden_parts`). | Normalize each *full* key path (joined across nesting, hyphens treated as an additional split delimiter alongside `_`) before calling `matches_forbidden_parts`, and add fixtures for dotted/hyphenated/concatenated forms to the existing TOML fixture suite. |
| U-3 | MAJOR | Anchor | `scripts/check-retired-architecture.sh` (`scan_toml_with_fallback`) | The line that *opens* a multiline TOML string (e.g. `slack = """`) is skipped for key inspection: `strip_toml_comment` mutates `state['in_multiline_basic']` to `True` while processing that same line, and the caller's very next check (`if state['in_multiline_basic'] or ... : continue`) then sees the POST-line state and skips the line — meaning a forbidden key on the *opening* line of a multiline value is never checked by the fallback path. **This fallback path is only used when `tomllib` is unavailable** (rare; the primary/self-test-exercised path is `tomllib`-based and would still catch this correctly), but it is exercised by the self-test's `[fallback]` engine label on every run and no current fixture exercises this exact scenario. **Pre-existing logic, not modified by this diff.** | Track whether a line *started* inside a multiline string separately from whether it *ends* inside one; inspect the LHS of any line whose multiline string state changed from false→true during processing. Add a regression fixture: a forbidden key whose value opens (does not close) a multiline string on the same line. |
| U-4 | MAJOR | Reviewer-C (Tier 3) | `scripts/check-retired-architecture.sh` (`scan_path`/`engine_for_path`) | `scan_path()` returns `[]` (silently clean) for any path whose `engine_for_path()` returns `None`. `should_scan_repo_path()` and `engine_for_path()` are two independently-maintained predicates that must agree by hand; the only structural guard added by this diff is one hardcoded assertion for the literal string `config.toml.example`, which cannot catch the general case (e.g. a future selected extension with no matching engine branch). | Make dispatch disagreement fail closed: raise or emit a synthetic finding when a selected path's engine is `None`, and add a self-test assertion `all(engine_for_path(root / p) is not None for p in select_repo_paths())`. |
| U-5 | MINOR | Reviewer-C (Tier 3) | `scripts/check-retired-architecture.sh` (`run_fixture_self_test`) | A fixture suite whose directory is emptied and whose manifest is emptied to `{}` in tandem produces no failures (`discovered == [] == manifest`) — the suite "passes" with zero fixtures and zero proof. Manifest values are only validated to be dict-shaped, not restricted to `{"accept","reject"}`. | Assert `len(discovered) > 0` per suite and that each manifest contains at least one `"accept"` and one `"reject"` entry; validate the value vocabulary in `load_fixture_manifest`. |
| U-6 | MINOR | Reviewer-C (Tier 3) | `scripts/check-retired-architecture.sh` (usage docstring, top of file) | The `--self-test` usage comment describes only the TOML and Go fixture-suite checks plus "verifies the real tracked tree passes," but doesn't mention the third check class this diff added (`run_repo_selection_self_test`'s 5 structural assertions), and "Exits 0 only when both checks succeed" is now numerically stale (three checks, not two). Separately, the repo-scan usage text doesn't call out that `cmd/**/*_test.go` IS scanned (unlike `internal/**/*_test.go`, which is excluded) — an easy-to-misread asymmetry. | Update the header comment to enumerate all three self-test check classes and explicitly state the `cmd/**` test-file asymmetry. |
| U-7 | MINOR | Reviewer-C (Tier 3) | N/A (process recommendation) | Static reasoning indicates the 2 new Go fixtures should not trip gitleaks (see §2(d)), but this was not confirmed by an actual gitleaks run in this session. | Confirm the CI `security` job / a local `gitleaks detect --source . --no-git -v` run is green after this PR's fixtures are pushed. |
| U-8 | MINOR | Reviewer-C (Tier 3) | `scripts/check-retired-architecture.sh` (scope, pre-existing) | The broadening to `internal/**` does not introduce any *new* directory-name evasion vector (a hypothetical `internal/mocks/` or `internal/fixtures/` is correctly scanned, not silently excluded — only the literal `/testdata/` segment and `_test.go` suffix are excluded). However, several **pre-existing** blind spots are now implicitly covered by D6c's broader coverage claim: (i) `mask_go_non_code` blanks all string literals, so an aliased import (`import sl "…/slack-go/slack"` used as `sl.New()`) is invisible to the gate; (ii) `go.mod`/`go.sum` are outside the pathspec entirely, arguably the most likely accidental-reintroduction vector; (iii) non-`.go`/`.toml` files under `internal/` (e.g. `go:embed`'d JSON/YAML) are unscanned. | Consider a follow-up to add `go.mod` to the pathspec with a cheap line-oriented engine; record (i)/(iii) as explicit residual risk in D6c so its coverage claim isn't read as stronger than the control actually provides. |
| U-9 | MINOR | Reviewer-C (Tier 3) | `scripts/testdata/retiredgo/*.go` (scope, deliberate) | The 2-fixture cap (by design, per task 013.003-T) exercises only 2 of `mask_go_non_code`'s 6 masking states (code, line-comment); block comments, interpreted strings, raw strings, and rune literals remain unexercised by the Go fixture suite. Explicitly acknowledged as an intentional scope boundary, not a defect. | No action required for this shipment; track as a residual-coverage note for a future full masking-characterization fixture suite. |
| U-10 | MINOR | Reviewer-C (Tier 3) | `.github/workflows/ci.yml` (pre-existing, **not part of this diff's file list**) | Verified directly: the `lint` job's "Run retired-architecture gate" step runs the bare `scripts/check-retired-architecture.sh` (no args) with `continue-on-error: true`, so that step alone cannot fail CI. Real enforcement of the broadened `internal/**` scope currently depends entirely on the *next* step, "Run retired-architecture self-test" (`--self-test`, blocking), because `--self-test` mode happens to re-invoke the bare repo scan internally after its fixture checks. This wiring pre-dates this diff and `ci.yml` is not among the 7 changed files, so it is out of this PR's scope to fix, but it means D6c's "enforced repo scope" claim is *currently* enforced only as a side effect of the self-test step's internal implementation detail, not by a dedicated blocking step. | Out of scope for this PR. Track as a follow-up: either drop `continue-on-error: true` from the dedicated gate step, or remove the redundant step so exactly one blocking invocation carries enforcement. |

## 7. Remediation plan (ordered by priority = confidence_weight × severity_weight)

| Priority | Finding | Confidence | Severity | Score | Action class |
|---|---|---|---|---|---|
| 1 | C-1: self-test "selection structural inclusion" partial vacuous-truth / empty-set risk | HIGH (3) | MAJOR (3) | 9 | `gated_auto` |
| 2 | M-1: `config.toml.example` selection membership not asserted end-to-end | MEDIUM (2) | MINOR (2) | 4 | `gated_auto` |
| 2 | M-2: D6c commit-hash `6dec85a` unverified | MEDIUM (2) | MINOR (2) | 4 | `manual` |
| 4 | U-1: `split_identifier` misses concatenated Go identifiers | LOW (1) | MAJOR (3) | 3 | `gated_auto` — unusual enough (real detection gap) to flag despite single source |
| 4 | U-2: TOML key matching misses dotted/hyphenated/concatenated keys | LOW (1) | MAJOR (3) | 3 | `gated_auto` |
| 4 | U-3: TOML fallback lexer skips LHS on multiline-opening line | LOW (1) | MAJOR (3) | 3 | `gated_auto` |
| 4 | U-4: `scan_path` fails open on unmapped engine | LOW (1) | MAJOR (3) | 3 | `gated_auto` |
| 8 | U-5: fixture suite can pass with zero fixtures | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U-6: usage docstring drift (3rd check class, cmd/** asymmetry) | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U-7: confirm live gitleaks run | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U-8: pre-existing evasion blind spots (import aliasing, go.mod/go.sum, non-go/toml) | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U-9: masking state machine undertested by design | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U-10: CI enforcement is a side effect of self-test wiring (out of PR scope) | LOW (1) | MINOR (2) | 2 | `advisory` |

Note: U-1, U-2, and U-3 share a root cause (the token-matching approach's
weakness against non-camelCase, non-underscore-delimited, or dotted/hyphenated
forms) even though they were raised independently against different code
paths (Go identifier splitting vs. TOML key splitting vs. the TOML fallback
lexer specifically) and therefore could not be fuzzy-matched into a single
finding under the file+line±2+rule key. Human reviewers should treat these
three together as one systemic weakness in the retired-token matching
approach, pre-dating this diff, worth a dedicated follow-up hardening pass.

## 8. Backlog work item entries (P0/P1 findings)

```yaml
type: bug
title: "Self-test 'selection structural inclusion' assertion has vacuous-truth risk on empty git ls-files result"
description: "expected_internal_repo_paths() and should_scan_repo_path() derive from the same git ls-files -- internal/** call and duplicate the same exclusion filter; if that pathspec ever returns zero rows (e.g. GIT_LITERAL_PATHSPECS=1), both sides are empty and the assertion vacuously passes while the gate scans no internal file. The aggregate 'selection non-empty' check does not catch this because cmd/**/config.toml.example alone satisfy it."
file: "scripts/check-retired-architecture.sh"
line: 485
severity: "MAJOR"
confidence: "HIGH"
fix: "Add internal/-scoped non-emptiness assertions (len(internal_expected) > 0, len(internal_actual) > 0) and consider an independently-derived expected set."
linked_review: "docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md"
```

```yaml
type: task
title: "Assert config.toml.example is actually present in select_repo_paths(), not only that its dispatch routes to TOML"
description: "The 'dispatch config.toml.example' self-test assertion only proves engine_for_path() routing; nothing asserts the file is actually selected by select_repo_paths() during a real scan, so a rename/untrack regression would silently stop scanning it while this assertion keeps passing."
file: "scripts/check-retired-architecture.sh"
line: 518
severity: "MINOR"
confidence: "MEDIUM"
fix: "Add report_assertion('selection includes config.toml.example', 'config.toml.example' in rel_paths, ...)."
linked_review: "docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md"
```

```yaml
type: task
title: "Confirm D6c's commit-hash claim (6dec85a) via git log/git show before treating it as settled governing-doc fact"
description: "D6c states the gate was introduced already narrowed in commit 6dec85a, supporting a 'broadening for the first time, not a reversal' framing. No reviewer in this session had git access to confirm the commit's existence or content."
file: "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
line: 530
severity: "MINOR"
confidence: "MEDIUM"
fix: "Run git log -1 6dec85a --oneline -- scripts/check-retired-architecture.sh and git log -S\"internal/**\" -- scripts/check-retired-architecture.sh; correct D6c if the claim does not hold."
linked_review: "docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md"
```

## 9. Post-remediation re-review (Phase 7)

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 13
  status: "skipped"
```

Skipped by design: this invocation was run in `mode: report-only` per explicit
operator instruction ("Do NOT modify any source file — this is report-only,
read/analyze only"). No `safe_auto` fixes were applied, so Phase 7 has no
fixed files to re-review. All findings above remain open for the operator to
triage, fix, and (optionally) request a follow-up adversarial review pass
after changes land.

## 10. Overall readiness verdict

**READY_WITH_FOLLOWUPS**

Rationale:

* The core deliverables are genuinely correct: the P0 `config.toml.example`
  TOML-dispatch bug is fixed and non-vacuously covered; the scope broadening
  to `internal/**` is compatible with the current real tree (independently
  confirmed clean by direct token search); the 2-fixture Go suite is
  gitleaks-safe by inspection and correctly proves comment-masking and
  bare-identifier detection; the governance-doc annotations are additive,
  non-destructive, and consistent with the "decision doc governs" convention.
* One HIGH-confidence MAJOR finding (C-1) shows the newly-added self-test
  structural assertions have a real, if narrow, vacuous-pass risk — ironic
  given this shipment's purpose was to strengthen exactly this kind of
  assertion. This should be fixed or explicitly acknowledged before the
  shipment is considered fully closed, but it does not indicate the
  broadening itself is broken on the current tree.
* Several MAJOR-severity but LOW-confidence (single-reviewer) findings surface
  **pre-existing** detection-gap bugs in the token-matching logic
  (concatenated-identifier evasion, TOML dotted/hyphenated-key evasion, a
  TOML-fallback lexer bug) that predate this diff and are legitimate backlog
  candidates, not blockers for merging this specific change.
* This review could not execute the self-test, `gofmt`/`go vet`/`go
  test`/`go build`, or `git log` verification commands due to a hard
  tooling restriction in this session. **The operator must run these for
  real before merge** — this review's "PASS" assessments for (a)–(c) in §2
  are based on rigorous static tracing, not observed command output.

