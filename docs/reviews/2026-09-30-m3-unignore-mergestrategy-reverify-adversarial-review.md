# Targeted Re-Verification: M3 Unignore/Merge-Strategy Fixes (C-1, M-1)

**Date**: 2026-09-30
**Mode**: Report-only, targeted re-verification (NOT a full re-review)
**Scope**: Two previously-identified BLOCKING findings from
`docs/reviews/2026-09-30-m3-unignore-mergestrategy-go-port-adversarial-review.md`
**Reviewers**: 3 (same roster as prior round)

## Model Route Assignment

| Reviewer | Route | Model | Notes |
|---|---|---|---|
| Anchor Reviewer | Anchor route | `gpt-5.6-sol` via dispatchable anchor route | Dispatched successfully |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | Dispatched successfully |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | Dispatched successfully |

No Anchor/alternate-provider fallback was needed; all three reviewer slots dispatched
and returned results. Reviewer-B (Tier 2 standard) was not included for this 3-reviewer
targeted pass, per the explicit instruction in the task ("reviewers: 3, same roster as
before: Anchor, Tier-1, Tier-3").

**Sandbox constraint honored**: all three reviewers were instructed to read current
file content directly via a file-view tool rather than `git diff`/`git show`, per the
prior round's tool-reliability issue. All three reviewers' evidence cites verbatim
code/test content (function bodies, exact assertions), not diff output.

## Finding C-1 (CRITICAL) — `splitNulTerminated` silent-swallow of invalid UTF-8

**File**: `tools/gatecheck/internal/unignore/checks.go`
**Claimed fix**: `splitNulTerminated` returns `([]string, error)`; both call sites in
`runDifferentialCheck` check and propagate with `::error::`-prefixed messages.

| Reviewer | Verdict |
|---|---|
| Anchor (gpt-5.6-sol) | RESOLVED |
| Tier-1 (gpt-5.4-mini) | RESOLVED |
| Tier-3 (claude-opus-5) | RESOLVED |

**Consensus: 3/3 — RESOLVED (HIGH confidence)**

Evidence (converged across all three reviewers):

- `splitNulTerminated(raw []byte) ([]string, error)` now calls `pysem.GitText(raw)`
  and does `return nil, err` on failure — no swallow.
- Both call sites in `runDifferentialCheck` check the returned error:
  - `ls-files --others -z` branch: `allUntracked, splitErr := splitNulTerminated(untrackedOut); if splitErr != nil { return 0, nil, fmt.Errorf("::error::git ls-files --others produced invalid UTF-8 output: %v", splitErr) }`
  - `diff --name-only -z` branch: `diffPaths, splitErr := splitNulTerminated(diffOut); if splitErr != nil { return 0, nil, fmt.Errorf("::error::git diff --name-only produced invalid UTF-8 output: %v", splitErr) }`
- Both messages are `::error::`-prefixed and propagate to exit 1 at the `runCheck`/
  `runSelfTest` layer — same exit code and verdict as the Python engine's
  `UnicodeDecodeError` traceback, satisfying ED-2 ("same exit and same verdict; only
  the message changes").
- Tests genuinely exercise the fixed path and would fail under the old behavior:
  - `TestSplitNulTerminated_InvalidUTF8Propagates` feeds
    `"valid-path\x00\xff\xfe-invalid-utf8-path\x00"` and `t.Fatalf`s if `err == nil`.
    The old single-return-value swallow behavior would not even compile against this
    two-return-value test, let alone pass the assertion.
  - `TestRunDifferentialCheck_InvalidUTF8FailsClosed` has two subtests (one injecting
    invalid UTF-8 via a stub `GitRunner` into the `ls-files` branch, one into the
    `diff` branch), each asserting `err != nil`. Under the old swallow-to-nil
    behavior, the decode failure would silently collapse the candidate union to
    empty, hit the `len(candidates) == 0` vacuous-pass branch, and return
    `(0, nil, nil)` — both subtests would fail (no error produced).
- No other call site of `splitNulTerminated` exists in the package that bypasses the
  new error check (`checkignore.go`/`unignore.go` use `pysem.GitText` directly with
  their own independent error handling).

**Remaining issues**: none reported by any reviewer.

## Finding M-1 (MAJOR) — `evaluate_response` temp-file leak via subshell scoping

**File**: `scripts/check-merge-strategy.sh`
**Claimed fix**: `evaluate_response <tmp-file> <json-string>`; both call sites
(`run_self_test`'s transport-fixture loop, `run_repo_scan`) allocate the temp file
and register it in `MERGE_STRATEGY_TMP_FILES` in the parent shell, before calling
`evaluate_response`.

| Reviewer | Verdict |
|---|---|
| Anchor (gpt-5.6-sol) | RESOLVED |
| Tier-1 (gpt-5.4-mini) | RESOLVED |
| Tier-3 (claude-opus-5) | RESOLVED |

**Consensus: 3/3 — RESOLVED (HIGH confidence)**

Evidence (converged across all three reviewers):

- `evaluate_response()` body is now exactly:
  ```bash
  evaluate_response() {
    local tmp="$1"
    local payload="$2"
    printf '%s' "$payload" > "$tmp"
    local output
    output="$(evaluate_json "$tmp" || true)"
    printf '%s\n' "$output"
  }
  ```
  No `mktemp` call and no `MERGE_STRATEGY_TMP_FILES+=` append inside the function body.
- Both call sites allocate and register in parent-shell scope, as plain statements
  (not nested inside any `$(...)`), before invoking `evaluate_response`:
  - `run_self_test`'s transport-fixture loop:
    `transport_tmp="$(mktemp)"; MERGE_STRATEGY_TMP_FILES+=("$transport_tmp"); transport_output="$(evaluate_response "$transport_tmp" "$(cat "$path")")"`
  - `run_repo_scan`:
    `tmp="$(mktemp)"; MERGE_STRATEGY_TMP_FILES+=("$tmp"); output="$(evaluate_response "$tmp" "$response")"`
  - In both cases the `mktemp`/array-append statements execute directly in the
    enclosing function's shell (no subshell), so they ARE visible to the parent
    script's `MERGE_STRATEGY_TMP_FILES` array that `cleanup()` drains on EXIT.
- No remaining one-argument call site of `evaluate_response` exists (full-file read
  shows exactly three occurrences: the definition/doc-comment and the two
  two-argument call sites above) — no signature-mismatch bug.
- The subshell-scoping hazard is genuinely eliminated: the `$(evaluate_response ...)`
  substitution still forks a subshell, but that subshell only *writes to* an
  already-registered path; it performs no `mktemp`/array mutation of its own, so an
  early exit inside it cannot leak an untracked file. The temp-file lifecycle
  (allocate → register → remove via the single `cleanup()` EXIT trap) is now
  entirely in parent-shell scope, matching the trap's `for f in
  "${MERGE_STRATEGY_TMP_FILES[@]:-}"; do rm -f -- "$f"; done` drain.

**Remaining issues**: none reported by any reviewer.

## Overall Consensus

| Finding | Severity | Consensus | Verdict |
|---|---|---|---|
| C-1 | CRITICAL | 3/3 | **RESOLVED** |
| M-1 | MAJOR | 3/3 | **RESOLVED** |

**OVERALL: CLEARED**

Both previously-blocking findings are confirmed resolved by unanimous, independent
re-verification against current file content. No new findings were raised (out of
scope for this confirmation pass, per instruction). No backlog work items are
generated — both findings are closed.

## Post-Remediation Re-Review

Not applicable — this was a report-only confirmation pass over already-applied fixes,
not a round that itself applied new `safe_auto` fixes. No Phase 7 cycle was run.

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: "skipped"
```
