---
title: "Adversarial Review — M3 unignore-regression and merge-strategy Go port"
mode: report-only
date: 2026-09-30
shipment: 036-S
feature: 046-F
tasks: [046.001-T, 046.002-T, 046.003-T, 046.004-T, 046.005-T, 046.006-T, 046.007-T, 046.008-T, 046.009-T, 046.010-T, 046.011-T, 046.012-T, 046.013-T]
reviewed_range: 0092a08..HEAD (fe83bc5)
branch: feat/036-s-gate-engine-go-migration-m3-unignore-regression-and-merge-strategy-engines
reviewers: 3
post_remediation_review:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: skipped
---

# Adversarial Review — M3 unignore-regression / merge-strategy Go port

**Mode: report-only.** No files were modified. No `safe_auto` fixes were applied; Phase 7
(post-remediation re-review) is therefore **skipped** (nothing to re-review).

## 0. Methodology note and a disclosed tooling limitation (read before the findings)

This review targeted the diff `0092a08..HEAD` (`fe83bc5`). During Phase 1 preparation it was
discovered that **neither this orchestrator nor its dispatched reviewer subagents have reliable
`git diff`/`git show`/`git log` execution in this sandbox**: most invocations either failed
outright, or — in one directly observed case — returned the *current* working-tree content of
`scripts/check-unignore-regression.sh` when asked for `git show 0092a08:scripts/check-unignore-regression.sh`,
i.e. a fabricated/stale result rather than the actual historical blob. That fabricated attempt was
discarded and not used for any finding below.

To keep the review grounded in verifiable fact rather than hallucinated diff output, this review
was conducted as a **direct audit of the post-change code against the plan's authoritative task/AC
spec, the ED (enumerated-delta) table, and the unit's own evidence document**, rather than a literal
line-by-line diff. The three reviewer subagents were explicitly instructed not to fabricate git
output and to work only from `view`-tool reads of the current tree plus the embedded plan/evidence
text (which itself documents the retired Python engine's exact semantics in comment form, and whose
goldens were independently confirmed — see §2 — to be genuine external captures, not Go-vs-Go
self-reference). This is a materially different evidence basis than a mechanical diff review, and
is disclosed here so the reader can weigh it accordingly. It does **not** reduce confidence in the
findings below, which were each independently re-traced by this orchestrator directly against the
current source (cited by absolute path and function name), but it does mean "no diff noise outside
the stated file list was observed" is an assumption inherited from the evidence document's own
`git diff --stat`-equivalent claims, not independently re-verified bit-for-bit here.

## 1. Reviewer panel

| Reviewer | Route | Model | Notes |
|---|---|---|---|
| Anchor Reviewer | Anchor route | `gpt-5.6-sol` (reasoning: high) | Dispatched successfully |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | Dispatched successfully |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | Dispatched successfully |

3-reviewer anchor-dispatchable mapping per protocol (Anchor + Tier 1 + Tier 3). No declared
fallback was needed — all three routes dispatched and returned structured JSON findings.

## 2. Mandate-item disposition (summary, before the finding list)

1. **ED-1 (merge-strategy outside-checkout exit 2)** — correctly implemented in
   `scripts/check-merge-strategy.sh` (`git rev-parse --show-toplevel` guard runs before
   `gatecheck_build`, prints `::error::` and exits 2 on failure). Scoped correctly: no other path
   through the script is affected. Tested via the evidence doc's manual non-checkout demonstration;
   no Go unit test directly exercises this bash-level guard, but this matches the pattern already
   established for `check-retired-architecture.sh`'s analogous (and intentionally different, exit
   128 vs 2) guard. No finding raised by any reviewer.
2. **ED-5 (additive self-test rows)** — correctly implemented: `verdict_exit_code`'s PASS→0/SKIP→0/
   FAIL→1 assertions are added to the existing self-test loop without removing or reordering any
   prior row, and the evidence doc records a genuine red-first proof (FAIL→return 0 mutation caught
   by the new assertion). No finding raised by any reviewer.
3. **ED-8 (reason-text wording)** — correctly scoped to the three named exempted cases
   (`nan-literal`, `trailing-data-after-object`, `two-concatenated-objects`) in
   `mergestrategy/golden_test.go`'s `runGoldenCase`; every other case's reason text is asserted
   byte-for-byte against the golden. No finding raised by any reviewer.
4. **Golden/test design (vacuity check)** — both `unignoreGolden`/`mergestrategyGolden` are loaded
   from committed JSON files via `os.ReadFile` + `json.Unmarshal` and compared against live `Run`/
   `Evaluate` output; this is a genuine external-capture-vs-Go-output comparison, not a self-referential
   Go-vs-Go test. Confirmed directly by this orchestrator reading both `golden_test.go` files and a
   sample of both golden JSON files' contents (plausible, non-trivial captured Python output,
   including the exact `python_version: "3.14.3"` provenance tag). One LOW-confidence completeness
   gap was raised (see Unique Findings, item U-2).
5. **Wiring / dead Python-fallback check** — `register_unignore.go` and
   `register_merge_strategy_evaluate.go` both unconditionally dispatch to the Go `internal/unignore`
   and `internal/mergestrategy` packages with no stub/fallback branch; both wrapper scripts
   (`check-unignore-regression.sh`, `check-merge-strategy.sh`) contain no `python`/`python3`
   invocation anywhere in their current content, and `main_test.go`'s `TestRun_EachStub` was
   correctly emptied (with an explanatory comment) rather than left stale. No reviewer found any
   residual or conditional Python invocation path. **Clean.**
6. **Already-captured deferred items** (`7223218F`, `5A8EC1BC`, `978D2946`) and the
   already-confirmed-faithful-parity items (mergestrategy verdict ordering; unignore batch
   stdin transport shape) from `m3.md` §16 — **not re-raised**. This orchestrator's own reading
   of the code agrees with the existing disposition in each case; no reviewer disputed it either.

## 3. Consensus findings (confidence: HIGH — flagged by all 3 reviewers)

### C-1. `splitNulTerminated` silently fail-opens the un-ignore differential check on invalid-UTF-8 paths (INV-1 / ED-2 violation)

* **Severity**: CRITICAL
* **Confidence**: HIGH (3/3 reviewers, independently; orchestrator-verified by direct trace)
* **File**: `tools/gatecheck/internal/unignore/checks.go`, `splitNulTerminated` (~line 147) and its
  two call sites in `runDifferentialCheck` (~lines 74–91)
* **Issue**: `splitNulTerminated` calls `pysem.GitText(raw)` on the raw `-z`-delimited output of
  `git ls-files --others -z` and `git diff --name-only -z`. If `GitText` returns
  `pysem.ErrInvalidUTF8` (which it will for any single byte sequence in that output that is not
  valid UTF-8 — and git does not enforce valid-UTF-8 filenames on Linux), the function **swallows
  the error and returns `nil`** instead of propagating it. Because this function supplies *both*
  halves of `runDifferentialCheck`'s de-duplicated candidate union, a single invalid-UTF-8-named
  path anywhere in the untracked-or-touched set collapses the **entire** candidate list to empty —
  not just the one offending path. `runDifferentialCheck` then returns `(evaluated=0, failures=nil,
  err=nil)`, which `runCheck` cannot distinguish from the *legitimate* "clean checkout, nothing to
  evaluate" case (the plan's own C-5/§6 design explicitly permits `differential_evaluated == 0` as
  non-vacuous — only the denylist count is required to be non-zero). The check therefore reports
  **PASS** for a diff that may contain a genuine un-ignore regression, as long as the diff (or the
  working tree's untracked set) also happens to contain any single non-UTF-8-named file. The
  original Python engine decoded this same git output via `text=True`, which raises
  `UnicodeDecodeError` — an uncaught exception, exit 1 (fail-closed). ED-2 explicitly permits *only*
  a message-text change for this exact class ("Invalid-UTF-8 input, or a git/read error: today a
  Python traceback with exit 1; after, a one-line `::error::` message with exit 1 ... **Same exit and
  same verdict**"); this implementation instead changes the verdict from fail (exit 1) to pass
  (exit 0) — an unenumerated divergence, i.e. an INV-1 violation, in a security-relevant gate whose
  entire purpose is to prevent exactly this class of silent regression. The in-code comment
  asserting the swallow is "strictly more conservative than a crash" is backwards: an empty
  candidate set makes Part 2 of the check vacuous, not conservative.
* **Scope**: `splitNulTerminated`/`runDifferentialCheck` are themselves M3-T5 work (this unit), not
  pre-existing code — this is squarely in scope for 036-S.
* **Fix**: Change `splitNulTerminated`'s signature to return `([]string, error)` and propagate
  `pysem.ErrInvalidUTF8` (and any other decode error) to both call sites in `runDifferentialCheck`;
  have `runDifferentialCheck` return that error so `runCheck`/`runSelfTest` surface it as the ED-2
  one-line `::error::` message and return exit 1, matching the Python engine's fail-closed exit.
  Add a red test: a fake `GitRunner` whose `ls-files`/`diff` output contains an invalid-UTF-8 byte
  sequence, asserting `runCheck` returns exit 1 (not 0).
* **Action class**: `manual` (the fix is a signature change touching two call sites plus a new
  negative-path test; not auto-applicable without review of the resulting error message wording).

## 4. Majority findings (confidence: MEDIUM — flagged by 2 of 3 reviewers)

### M-1. `MERGE_STRATEGY_TMP_FILES` temp-file registration happens inside a subshell and is never seen by the cleanup trap (M3-T12 AC regression / resource leak)

* **Severity**: MAJOR
* **Confidence**: MEDIUM-HIGH (2/3 reviewers — Anchor and Tier-3; orchestrator independently
  re-derived and confirms this is real, standard bash subshell semantics, not a model artifact)
* **File**: `scripts/check-merge-strategy.sh`, `evaluate_response` (~line 136) and its two call
  sites in `run_self_test` and `run_repo_scan`
* **Issue**: `evaluate_response` appends its `mktemp` file path to the `MERGE_STRATEGY_TMP_FILES`
  array, which the script's single composed `cleanup()` EXIT trap is supposed to drain. However,
  **every** call site invokes `evaluate_response` inside command substitution —
  `transport_output="$(evaluate_response "$(cat "$path")")"` in `run_self_test`, and
  `output="$(evaluate_response "$response")"` in `run_repo_scan`. Command substitution
  unconditionally forks a subshell in bash; any variable (including array) mutation performed
  inside that subshell is confined to it and is never visible to the parent shell process whose
  `EXIT` trap eventually runs. The result: `MERGE_STRATEGY_TMP_FILES` is **always empty** in the
  parent shell, `cleanup()`'s loop over it removes nothing, and every self-test run (8 fixtures ×
  1 transport call each) and every live repo-scan leaks one `mktemp` file. This is the exact defect
  class M3-T12 item 1 and the evidence document (`m3.md` §10 item 1) claim to have fixed — the
  claim is incorrect as implemented. (The prior pattern — an imperative `rm -f` *inside*
  `evaluate_response` itself — would actually have worked, since removal and creation would occur
  in the same subshell; the refactor to a shared array + single trap broke this.)
* **Scope**: `check-merge-strategy.sh` M3-T12 work — in scope for 036-S.
* **Fix**: Either restore file removal inside `evaluate_response` itself (safe: created and
  consumed in the same subshell invocation), or restructure so the temp file path is allocated in
  the parent shell (e.g. have the caller `mktemp` the file and pass the path into
  `evaluate_response`/`evaluate_json` rather than allocating it inside the function that runs in a
  subshell). Add a test asserting no `MERGE_STRATEGY_TMP_FILES`-tracked file survives after a
  `--self-test` run (e.g. snapshot `$TMPDIR` entries before/after).
* **Action class**: `gated_auto` (small, mechanical, low-risk fix; confirm the chosen approach
  before applying, since it changes `evaluate_response`'s responsibility).

## 5. Plurality findings (confidence: MEDIUM, >1 reviewer but not majority)

None — with 3 reviewers, the "plurality but not majority" band (exactly a non-majority multiple
flag, e.g. 2 of 5) does not apply; every multi-reviewer finding here already met the >50% majority
threshold (2 of 3).

## 6. Unique findings (confidence: LOW — flagged by exactly one reviewer)

### U-1. `gatecheck_build`'s `go env GOEXE` probe is not failure-guarded before `set -e`

* **Severity**: reviewer-assigned CRITICAL (orchestrator assesses MAJOR-at-most — see scope caveat)
* **Confidence**: LOW (1/3 reviewers — Anchor only)
* **File**: `scripts/lib/gatecheck-run.sh`, `gatecheck_build` (~line 45):
  `goexe="$(env -u GH_TOKEN -u GITHUB_TOKEN GOTOOLCHAIN=local GOFLAGS=-mod=readonly go env GOEXE)"`
* **Issue**: both wrapper scripts run under `set -euo pipefail`. A bare assignment statement whose
  right-hand side is a failing command substitution (`var=$(failing-cmd)`) is **not** exempt from
  `set -e` in bash — if `go` is missing from `PATH` or this probe otherwise fails, the script
  terminates immediately with the raw underlying exit code (typically 127), never reaching the
  `go build` call that is explicitly guarded with `|| rc=$?` and the designed
  `::error::gatecheck build failed (exit N)` / exit-2 contract ED-3 requires ("Missing interpreter
  becomes a build failure ... Exit stays 2"). Orchestrator confirmed this matches standard bash
  `set -e` assignment-substitution semantics; the claim is technically sound.
* **Scope caveat**: `scripts/lib/gatecheck-run.sh` was added in **M1-T9**, not modified by any
  M3 task (M3-T7/T11's file lists are `scripts/check-unignore-regression.sh`/
  `scripts/check-merge-strategy.sh` + the two `register_*.go` files + `m3.md` — this shared runner
  is not among them, and nothing in `m3.md` describes touching it). This is therefore most likely a
  **pre-existing latent defect in shared M1 infrastructure**, not something introduced or modified
  by 036-S/046-F. It is preserved here as a disclosed, low-confidence, **out-of-scope** observation
  for a separate backlog item — it should not block this unit's merge, but it affects every
  wrapper built on this runner (write-path, retired-arch, unignore, merge-strategy alike).
* **Suggested fix** (for the separate backlog item): wrap the `goexe` probe the same way the
  `go build` call already is (`|| rc=$?`), and treat its failure through the same
  `::error::gatecheck build failed` / exit-2 path.
* **Action class**: `advisory` (out of scope for this unit's merge decision; route to backlog).

### U-2. Golden's `Scenarios` field (name + verdict rows) is loaded but never directly asserted against

* **Severity**: reviewer-assigned MAJOR (orchestrator assesses MINOR — see rationale)
* **Confidence**: LOW (1/3 reviewers — Anchor only)
* **File**: `tools/gatecheck/internal/unignore/golden_test.go` (`unignoreGolden.Scenarios` field),
  consumed nowhere by name in `unignore_test.go`
* **Issue**: the golden JSON's `scenarios` array (5 named rows: `landing-precondition`,
  `reject-existing-file-unignored`, `accept-nonexistent-negation`,
  `reject-tracked-file-unignored-in-same-change`, `real-repo denylist check`) is decoded into
  `unignoreGolden.Scenarios` but no test iterates it directly to assert per-scenario pass/fail.
* **Orchestrator's assessment**: this is a real but low-impact completeness gap, not a parity hole.
  Every one of those five rows' PASS lines is a literal substring of `g.SelfTest.Stdout`, which
  `TestRun_SelfTest_MatchesGolden` *does* compare byte-for-byte end-to-end against the Go
  implementation's actual stdout. So the same assertion is effectively made, just indirectly
  through the full-stream comparison rather than by iterating the structured `Scenarios` field.
  The practical risk is an orphaned/dead JSON field inviting future drift (someone edits
  `Scenarios` expecting it to be load-bearing when it is not), not a missed regression class today.
  Downgraded from the reviewer's MAJOR to MINOR.
* **Fix**: either add a small loop asserting `g.Scenarios[i].Verdict` against each scenario's own
  return value for self-documentation, or remove the field/add a comment noting it is
  evidence-only and superseded by the full-stream comparison.
* **Action class**: `advisory`.

## 7. Remediation plan (ordered by priority = confidence_weight × severity_weight)

| # | Finding | Confidence | Severity | Score | Action class | File |
|---|---|---|---|---|---|---|
| 1 | C-1 `splitNulTerminated` fail-open on invalid UTF-8 | HIGH (3) | CRITICAL (4) | 12 | `manual` | `tools/gatecheck/internal/unignore/checks.go` |
| 2 | M-1 `MERGE_STRATEGY_TMP_FILES` subshell leak | MEDIUM (2) | MAJOR (3) | 6 | `gated_auto` | `scripts/check-merge-strategy.sh` |
| 3 | U-1 `goexe` probe unguarded under `set -e` | LOW (1) | CRITICAL (4) | 4 | `advisory` (out of scope — route to backlog against `scripts/lib/gatecheck-run.sh`, M1) | `scripts/lib/gatecheck-run.sh` |
| 4 | U-2 orphaned `Scenarios` golden field | LOW (1) | MINOR (2, downgraded) | 2 | `advisory` | `tools/gatecheck/internal/unignore/golden_test.go` |

No fixes were applied (report-only mode). Phase 7 post-remediation re-review is **skipped**
(`status: skipped`, `cycles_run: 0`) because nothing was changed.

## 8. Backlog work item entries (P0/P1 findings)

```yaml
type: bug
title: "INV-1/ED-2: splitNulTerminated fail-opens unignore differential check on invalid-UTF-8 paths"
description: >
  tools/gatecheck/internal/unignore/checks.go's splitNulTerminated swallows
  pysem.ErrInvalidUTF8 from `git ls-files --others -z` / `git diff --name-only -z`
  output and returns nil instead of propagating the error, collapsing
  runDifferentialCheck's entire candidate universe to empty when ANY single
  untracked-or-touched path has a non-UTF-8 name. This lets a crafted or
  incidental non-UTF-8-named file mask a genuine un-ignore regression elsewhere
  in the same diff, reporting PASS where the original Python engine would have
  raised UnicodeDecodeError (uncaught, exit 1). This is an unenumerated
  verdict/exit-code divergence (ED-2 permits only a message-text change for
  this class, not a verdict change), violating this unit's INV-1 parity
  invariant in a security-relevant gate.
file: "tools/gatecheck/internal/unignore/checks.go"
line: 147
severity: "CRITICAL"
confidence: "HIGH"
fix: >
  Change splitNulTerminated to return ([]string, error) and propagate decode
  errors to both call sites in runDifferentialCheck; surface the error as the
  ED-2 ::error:: message with exit 1 from runCheck/runSelfTest. Add a red test
  with a fake GitRunner emitting an invalid-UTF-8 byte sequence, asserting
  exit 1.
linked_review: "docs/closure/2026-09-30-m3-unignore-mergestrategy-go-port-adversarial-review.md"
```

```yaml
type: bug
title: "M3-T12 regression: MERGE_STRATEGY_TMP_FILES leak via subshell-confined array append"
description: >
  check-merge-strategy.sh's evaluate_response appends its mktemp path to
  MERGE_STRATEGY_TMP_FILES, but every call site invokes evaluate_response
  inside command substitution ($(...)), which always forks a subshell in
  bash. The array mutation is therefore confined to that subshell and never
  reaches the parent shell whose EXIT trap drains the array, so every
  self-test run and every live repo-scan leaks one mktemp-created temp file.
  This reintroduces the leak class M3-T12 item 1 and m3.md section 10 item 1
  claim to have fixed.
file: "scripts/check-merge-strategy.sh"
line: 136
severity: "MAJOR"
confidence: "MEDIUM"
fix: >
  Restore file removal inside evaluate_response itself (same subshell that
  created the file), or restructure so the temp file is allocated in the
  parent shell and passed in rather than created inside a function that only
  ever runs inside command substitution. Add a test asserting no leaked temp
  file survives a --self-test run.
linked_review: "docs/closure/2026-09-30-m3-unignore-mergestrategy-go-port-adversarial-review.md"
```

## 9. Already-triaged items re-confirmed, not re-raised

Per the evidence document's own §16 disposition, and this review's independent agreement with it:

* `7223218F` (high) — `--root` guard duplicated per-wrapper, `check-write-path-precondition.sh` has
  none. Disposition stands: deferred, correctly out of scope for 036-S.
* `5A8EC1BC` (low) — three independent `GitRunner` shapes, no shared abstraction. Disposition
  stands.
* `978D2946` (medium) — no durable CRLF-drift prevention mechanism. Disposition stands.
* Mergestrategy verdict-ordering (true-overrides-malformed-sibling) and unignore's plain-newline
  (not NUL-delimited) batch stdin transport — both independently re-examined by this review's
  reviewers and judged faithful parity, consistent with the evidence doc's own prior investigation.

## 10. Overall readiness verdict

## **BLOCKED**

Rationale: the unanimous (3/3) **C-1** finding is a genuine, in-scope (M3-T5), unenumerated
verdict-divergence from the original Python engine in a security-relevant merge-blocking gate — the
exact class of regression INV-1 and the ED table exist to foreclose, and it directly contradicts
`m3.md` §15's self-certification ("No other verdict, exit code, or control-flow behavior was
changed"). The majority (2/3) **M-1** finding independently contradicts a specific, separately
evidenced M3-T12 acceptance criterion ("every evaluate_response temp file is removed ... so a file
is never leaked"). Both are narrowly scoped, mechanically fixable defects — neither requires
re-architecture — but both must be fixed and re-verified (ideally with the two red tests described
above) before this unit is safe to merge. The two LOW-confidence unique findings (U-1, U-2) do not
block merge: U-1 is most likely pre-existing M1 infrastructure outside this diff's scope and should
be routed to its own backlog item; U-2 is a minor test-hygiene nitpick.

Once C-1 and M-1 are fixed, re-run this same 3-reviewer panel against the changed files only
(Phase 7 equivalent) before re-assessing readiness.
