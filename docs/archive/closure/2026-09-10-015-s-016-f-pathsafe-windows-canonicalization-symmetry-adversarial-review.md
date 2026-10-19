---
title: "Adversarial Review — 015-S Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage"
date: 2026-09-10
shipment: 015-S
feature: 016-F
mode: report-only
status: complete
---

# Adversarial Review — 015-S Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage

**Scope**: `git diff main..HEAD` on
`feat/015-s-pathsafe-windows-canonicalization-symmetry-and-windows-lint-coverage`
at commit `467afed` (all 8 tasks + covering feature done, pre-remediation),
reviewed against the PASS-verdicted (attempt 2) plan at
`docs/plans/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-plan.md`
and its companion deliberation/plan-review artifacts under `docs/decisions/`.

**Mode**: report-only, standard review followed by multi-persona adversarial
review, per operator requirement #2 (standard first, then adversarial
multi-model).

**Reviewer pool**: 6 independent persona reviewers (Task-tool subagents),
covering the full changed surface (`internal/pathsafe/{root,pathsafe,
reparse_windows,reparse_other}.go` + both `*_test.go` files +
`.github/workflows/ci.yml`):

| Persona | Focus |
|---|---|
| Go Reviewer | Go-specific safety, idioms, correctness |
| Security Reviewer | Injection, escape, containment, trust-boundary risk |
| Correctness Reviewer | Logic errors, edge cases, invariant violations |
| Concurrency Reviewer | Concurrent/parallel execution safety (scoped, low relevance — no new goroutines introduced) |
| Maintainability Reviewer | Complexity, coupling, dead paths, premature abstraction |
| Architecture Strategist | Cohesion, module boundaries, CI/gate design |

## Result: zero P0/P1 findings across all six reviewer passes

## Consensus finding requiring plan cross-check (false positive, resolved by clarification only)

Four of six reviewers (Go Reviewer, Correctness Reviewer, Maintainability
Reviewer, Architecture Strategist) independently flagged the retained
`stripUNCPrefix` calls in `root.go` (`NewRoot`) and `pathsafe.go`
(`checkSymlinkEscape`) as apparently redundant/dead code now that
`canonicalizeReparse` internalizes the same postcondition.

**Verification against the plan**: this is an already-adjudicated design
decision, not a defect. The plan explicitly requires both call sites be
**RETAINED as belt-and-suspenders, not removed**:
- `root.go`'s `NewRoot` call site: plan decision **D-2**.
- `pathsafe.go`'s `checkSymlinkEscape` call site: plan **B2/AC3**.

**Remediation**: clarifying comments citing D-2 and B2/AC3 were added at both
call sites (no code removal), so the intent is now discoverable without
re-litigating the decision on every future review pass.

A related false positive (Architecture Strategist, P2): the new CI
`golangci-lint run (GOOS=windows)` step being unconditionally blocking
(not advisory) was also flagged as a possible risk. Verified against plan
**AC2/AC4**, which explicitly *requires* the gate to be blocking. No action
needed beyond the comment-wording fix already made for accuracy (see below).

## Out-of-scope finding (P-021 deferred)

Architecture Strategist (rated P1 in isolation): `checkSymlinkEscape`'s
`canonicalizeReparse` error path is not wrapped via `wrapPathError` the way
`NewRoot`'s is, so its error surface is inconsistent with `NewRoot`'s.

**P-021 C1 classification**: out of scope. The plan's AC3
error-surface-preservation requirement is explicitly scoped to *"NewRoot's
failure branches"* only — `checkSymlinkEscape`'s (pre-existing, unmodified
by this shipment) error wrapping is a different, untouched contract
surface. Fixing it would require expanding the authorized change beyond
what 015-S's plan approved.

**Disposition**: captured per the P-021 C2 mandatory defer-capture
procedure. Discovery search (active + archived stash) found zero prior
matching entries. Captured as deferred entry **`7ADAC481`** (six-field
payload: token, one-sentence expansion statement, P-021 C1 rationale,
source refs [task=N/A cross-cutting, feature=016-F, shipment=015-S,
PR=N/A pre-PR at capture time, thread=N/A pre-PR], `requires deliberation:
yes`, kind=task/provisional priority=low). Threadless path (pre-PR local
review finding, no review thread existed at classification time) —
capture only, no thread reply/resolve applicable; cited in the closure
artifact's residual-risk record per C3's threadless discharge.

## In-scope remediation applied directly

1. Clarifying belt-and-suspenders comments at both `stripUNCPrefix` call
   sites (D-2, B2/AC3 citations) — see above.
2. `wrapPathError`'s `Op` label corrected from the inaccurate `"lstat"` to
   `"open"` (Go Reviewer + Correctness Reviewer, both P2 — `NewRoot`'s
   actual failing syscall class is `open`-family, not `lstat`).
3. A CI comment reworded to remove an overclaiming past-tense verification
   date (Architecture Strategist, P3).
4. An explicit `//go:build windows` tag added to `root_windows_test.go`
   (Go Reviewer, P2 — file already used Windows-only APIs and was already
   gated by the `_windows_test.go` filename suffix, but the explicit tag
   makes the constraint self-documenting and immune to a future rename).

## Targeted re-review of the remediation diff

A follow-up Correctness Reviewer pass was run against the uncommitted
remediation diff (items 1–4 above) before committing. Confirmed all four
items correctly resolved; found 2 additional minor P3 nits (CI comment
tense precision, D-2/B2-AC3 citation precision in `pathsafe.go`) — both
fixed and included in the same commit (`3c6927b`).

## Advisory-only follow-ups (not P-021 out-of-scope; future hardening ideas)

Consolidated into stash entry **`6B751D8B`**: `addLongPathPrefix`
branch-reachability test-design nuance, MAX_PATH UTF-16 precision, buffer
pooling opportunity, CI lint-runtime scoping question. None block merge;
none represent defects in the shipped scope.

## Commits

- `3c6927b` — in-scope remediation (items 1–4 plus the 2 re-review nits).
- `65d213e` — P-021 capture commit (stash entry `7ADAC481`).
- `6997700` — advisory follow-ups stash commit (stash entry `6B751D8B`).

## Outcome

**READY_WITH_FOLLOWUPS.** Zero P0/P1 at reviewed HEAD `3c6927b` (and at
final pre-PR HEAD `323e852`, which only added a memory checkpoint and
backlog bookkeeping on top — verified via `git diff --stat` excluding
`.backlogit/`, empty). One P-021 deferred entry (`7ADAC481`). One advisory
stash entry (`6B751D8B`). No re-review cycles beyond the single targeted
Correctness Reviewer pass were required.
