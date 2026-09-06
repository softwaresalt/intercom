---
title: "Implementation Plan — Build and scan script hardening (008-F)"
date: 2026-09-04
status: reviewed
phase: residual-hardening
feature: 008-F
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolves_stash: [35D76D5E, 78D13775, B6203CCA]
depends_on: none (self-contained post-remediation)
---

# Implementation Plan — Build and scan script hardening (008-F)

## Problem Frame

Two deferred findings target **script logic** (as opposed to 007-F's CI wiring):

1. **`35D76D5E` — `scripts/build.ps1` repo-root-escape guard (invariant I5) is
   lexical only.** It normalises the output path via `GetFullPath` and compares
   prefixes. `GetFullPath` resolves `..` and separators but **does not resolve
   symlinks**, so a symlinked ancestor of the output directory can point outside
   the repo root while still passing a lexical prefix test. This is a
   Constitution-IV containment control failing open in a narrow case.
2. **`78D13775` — `check-retired-architecture.sh`'s `strip_toml_comment` masking
   is line-oriented.** It tracks single-line basic/literal string state only. It
   has no notion of TOML multi-line basic (`"""`) or literal (`'''`) strings, so
   a `#` inside such a string can be mis-read as a comment start (hiding real
   content from the scan) or content after it mis-treated as live (producing a
   false positive).

Neither has a known live exploit path today — the repo's tracked TOML uses no
multi-line strings, and no symlinked build output is configured. Both are
*guards that would fail silently* the first time the shape appears, which is the
worst failure mode for an unattended gate.

## Non-Goals

* No change to *which* patterns `check-retired-architecture.sh` treats as
  retired-architecture markers. Only the comment-masking state machine changes.
* No change to CI wiring — 007-F owns that. This shipment relies on 007-F having
  already wired `--self-test` into CI (U6), so its logic change lands with
  automated regression coverage in place.
* No Go source changes.

## Constitution Check

| Principle | Status | Units | Justification |
|---|---|---|---|
| I. Safety-First Go | N/A | — | No Go code. |
| II. Test-First Development (NON-NEGOTIABLE) | Satisfied | U2, U4 | Tests/fixtures precede both logic changes (U2→U1, U4→U3). |
| III. Workspace Isolation and Security Boundaries | Satisfied | U1 | Strengthens the repo-root containment guard. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | Satisfied | U1 | This unit *is* an I5 containment control; AC-3 forbids loosening it. |
| V. Structured Observability | Satisfied | U1, U3, U5 | Rejection messages name the resolved path and the invariant violated; the self-test reports each case by name. |
| VI. Single Responsibility | Satisfied | all | Logic (U1, U3) separated from tests/fixtures (U2, U4) and CI wiring (U5). |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | N/A | — | No destructive command is executed. |
| VIII. Explicit Safety Modes for Elevated Risk | Satisfied | U3 | The fail-open masking risk (R2) is gated by a mandatory positive-control fixture and a blocking stop condition, rather than accepted on judgement. |
| IX. Git-Friendly Persistence | Satisfied | all | Text fixtures and scripts. |
| X. Agent Context Efficiency | Satisfied | all | Small, independently reviewable units. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Satisfied | — | No history operations. |

## Implementation Units

### U1 — Make the `build.ps1` root-escape guard symlink-aware
*Resolves `35D76D5E` (part 1 of 2). Artifact class: script/code. Size M, complexity medium.*

Before the prefix comparison, fully resolve symlinks on **both** the repo root
and the resolved output directory, walking up to the nearest existing ancestor
when the output directory does not yet exist (the common case for a build
output path). Compare the *resolved* paths.

This deliberately mirrors the ancestor-walking technique already proven in
`internal/pathsafe.checkSymlinkEscape`, which was hardened during 002-S for
exactly this failure mode (a symlinked *intermediate* directory with a
non-existent leaf). Reusing the established approach rather than inventing a
second one is the point.

**Acceptance criteria**
1. An output directory reached through a symlinked ancestor pointing outside the
   repo root is **rejected**, even when the leaf directory does not yet exist.
2. A legitimate output directory inside the repo root — including one reached
   through a symlink that stays *inside* the root — is still accepted.
3. **The existing comparison policy is PRESERVED.** `build.ps1` deliberately
   compares Ordinal (case-sensitive) as a fail-closed choice. This unit adds
   canonical symlink resolution **before** that comparison and must **not**
   change it to case-insensitive. (The earlier draft required a Windows
   case-fold; review correctly identified that as *loosening* a containment
   control, contradicting the protected invariant that U1 may only tighten.
   Any change to the case policy is a separate deliberation.)
4. **Windows junctions/reparse points are covered, not just symlinks.**
   Directory junctions typically require no elevated privilege on Windows and
   provide the same lexical-in-root/external-target escape, so a
   symlink-only fix leaves the cheapest bypass open. Resolution must use a
   canonicalising API that resolves reparse points.
5. The rejection message names the resolved path and identifies invariant I5.
6. **Documented residual risk (not fixed here):** validation and use are
   separated in time, so a concurrent process could swap an ancestor between the
   check and the write (TOCTOU). This unit narrows the lexical bypass; it does
   not claim to close the race. Record this explicitly rather than implying the
   guard is atomic.

### U2 — Negative tests for symlink/junction root escape
*Resolves `35D76D5E` (part 2 of 2). Artifact class: tests. Size M, complexity medium. **Authored before U1** (Principle II).*

Add tests creating a real escape and asserting rejection, plus an in-root
positive control.

**Testability requirement:** the existing I5 test lives in
`tests/integration/build_script_test.go` and runs a ~5-minute full build matrix.
Driving new symlink cases through the whole script would dominate CI. U1 must
therefore **extract the resolve-and-compare step into a separately invocable
function**, and U2 tests that function directly rather than running a build.

**Acceptance criteria**
1. A test creates a directory symlink pointing outside the repo root, targets
   build output through it, and asserts rejection.
2. A **Windows junction** case is covered (see U1 AC-4), since junctions need no
   elevated privilege and are the cheaper bypass.
3. A positive-control test asserts an in-root symlink is accepted, guarding
   against a "fix" that simply rejects all links.
4. The existing non-symlink negative test continues to pass unchanged.
5. Tests exercise the extracted function directly; they do not invoke the full
   build matrix.
6. Where link creation is genuinely unavailable, tests **skip with an explicit,
   log-visible reason** naming the unavailable capability. A silent skip reads
   as a pass and would manufacture false confidence. **At least one non-skipped
   escape test must be observed running before this shipment closes**; if every
   escape test skips on every available platform, return the unit blocked.

### U3 — Make TOML comment masking multi-line-string aware
*Resolves `78D13775` (part 1 of 2). Artifact class: script/code. Size M, complexity high. Depends on U4.*

**Ground-truth correction from review:** `strip_toml_comment` / `scan_toml` are
**Python embedded in a heredoc** inside `scripts/check-retired-architecture.sh`
— not bash. The earlier draft's "bash state machine" framing and POSIX-
portability caveat were wrong and would have misdirected an unattended executor.

**Preferred implementation — parse, do not lex.** Replace the hand-rolled
comment masking with a real TOML parse (`tomllib`, stdlib in Python 3.11+) and
walk actual keys, **failing closed on a parse error**. Review demonstrated that a
hand-written lexer must additionally handle escaped-delimiter closes
(`"""text \\"""`), 4–5 quote runs adjacent to a closing delimiter, same-line
open/close, and inert nested delimiters — a set that is neither 2 hours of work
nor reliably completable. A parser eliminates the entire class.

**Fallback (only if `tomllib` is unavailable):** extend the existing per-line
state machine to carry multi-line string state across lines, and cap scope at
exactly the acceptance criteria below.

**Acceptance criteria**
1. A `#` inside a multi-line basic (`"""`) string is **not** treated as a
   comment start.
2. A `#` inside a multi-line literal (`'''`) string is likewise not a comment.
3. A `'''` inside a `"""` block (and vice versa) does **not** terminate it.
4. A genuine comment after the *close* of a multi-line string on the same line
   **is** still stripped.
5. An escaped-delimiter close (`"""text \\"""`) is handled correctly — the block
   ends, and a retired marker on the following line is still reported.
6. A malformed/unparseable TOML file **fails closed** (reported, not skipped).
7. All existing single-line masking behaviour is unchanged, evidenced by the
   pre-existing self-test cases passing untouched.

### U4 — Multi-line TOML fixtures and self-test cases
*Resolves `78D13775` (part 2 of 2). Artifact class: tests/fixtures. Size S, complexity medium. **Authored before U3.***

Add committed fixtures under `scripts/testdata/`, wired into `--self-test`.

**Isolation requirement:** `--self-test` currently hard-codes a single fixture,
so adding cases would force this unit to edit the scanner — mixing artifact
classes. U3 must therefore make `--self-test` **discover** fixtures (glob
`scripts/testdata/retired-*.toml` with the expected verdict encoded in the
filename or a manifest), so U4 adds **files only**.

**Acceptance criteria**
1. Fixtures exist for: a retired marker hidden inside a `"""` block (must **not**
   be reported — it is string content); a genuine retired marker after a closed
   multi-line string (**must** be reported); nested inert delimiters; an
   escaped-delimiter close; and a malformed file (must fail closed).
2. Each fixture is discovered and reported by name by `--self-test`.
3. `--self-test` fails if any case regresses.
4. The positive cases (markers that **must** still be reported) are the
   load-bearing ones: they are what prevents an over-masking bug from silently
   blinding the scanner. A fixture set containing only negative cases is
   insufficient and must be rejected at review.

### U5 — Wire `check-retired-architecture.sh --self-test` into CI
*Resolves `B6203CCA`. Artifact class: CI config. Size XS, complexity trivial. **Depends on U3 and U4 — must land last.***

Relocated here from 007-F. Add a second invocation of the script in
`--self-test` mode alongside the existing bare scan at `ci.yml:145`.

This unit lands **after** U3/U4 precisely so the test-first fixtures (which must
fail against the pre-U3 implementation) can be committed without reddening a
gate that does not yet exist.

**Acceptance criteria**
1. CI invokes both `bash scripts/check-retired-architecture.sh` (repo scan) and
   `bash scripts/check-retired-architecture.sh --self-test`.
2. The self-test step fails the job if any committed fixture produces the wrong
   verdict — a real regression gate on the detection logic, not a smoke test.
3. The existing bare-scan step keeps its current blocking/non-blocking posture
   unchanged.
4. The self-test step is green at the moment it is introduced (U3/U4 already
   landed).

## Dependency Graph

```
U2 ──► U1        (tests authored first, per Principle II)
U4 ──► U3        (fixtures authored first)
U3, U4 ──► U5    (CI gate wired only once it can be green)
```
**No cross-shipment dependency.** 008-F is self-contained; the `--self-test`
wiring that previously lived in 007-F is now U5 here.

## Post-Review Remediation Record (2026-09-04)

Adversarial multi-persona review (Security, Maintainability, Scope,
Constitution) returned **FAIL** on the pre-remediation draft. Changes:

| Finding | Severity | Remediation |
|---|---|---|
| **U1 AC-3 required a Windows case-fold**, which *loosens* the deliberately Ordinal, fail-closed comparison in `build.ps1` — contradicting the protected invariant that U1 may only tighten. | **blocker** | AC-3 rewritten to **preserve** Ordinal comparison; case policy is now explicitly out of scope. |
| **`strip_toml_comment` is Python in a heredoc, not bash.** The plan's bash framing and POSIX caveat would have misdirected an unattended executor. | **blocker** | Ground truth corrected; implementation retargeted to `tomllib` parsing with a lexer fallback. |
| **U3 was not honestly a 2-hour unit** as a hand-written TOML lexer (escaped closes, quote runs, nested inert delimiters). | major | Parse-don't-lex approach removes the class; scope capped at the ACs. |
| **A single positive-control fixture was insufficient** to prove the scanner fails closed. | major | Fixture set expanded (escaped-delimiter close, malformed file); AC-4 makes positive cases load-bearing and rejects negative-only sets. |
| **U4 could not add fixtures without editing the scanner** (hard-coded single fixture) — an artifact-class violation. | major | U3 must make `--self-test` discover fixtures by glob/manifest so U4 adds files only. |
| **U1 covered symlinks but not Windows junctions**, the cheaper unprivileged bypass. | major | AC-4 added; U2 AC-2 requires a junction test. |
| **U2 would have driven a ~5-minute build matrix** and was ordered after U1, violating test-first. | major | U1 must extract a directly testable resolve-and-compare function; graph reversed to U2→U1. |
| **A silent skip reads as a pass**, leaving a containment guard unexercised. | major | U2 AC-6 requires at least one non-skipped escape test observed before close, else return blocked. |
| TOCTOU between validation and write was unaddressed and implicitly overclaimed. | minor | Recorded as explicit documented residual risk (U1 AC-6) rather than implied closed. |
| Constitution table: truncated titles; VII/VIII misclassified. | major | Rewritten with exact ratified titles and corrected classifications. |
| `B6203CCA` arrived from 007-F to fix a cross-shipment ordering blocker. | — | Absorbed as U5, landing last. |

## Risks and Caveats

**R1 — Windows symlink creation requires privilege.** `New-Item -ItemType
SymbolicLink` needs Administrator or Developer Mode. In an unattended runner
this may be unavailable. Mitigation: U2 AC-4 requires an explicit, *visible*
skip. The danger is a skip that reads as a pass — a guard whose test silently
never runs is worse than no test, because it manufactures false confidence.
Ship must confirm the skip is surfaced in CI output.

**R2 — U3 could over-mask and blind the scanner.** A state machine that
incorrectly believes it is inside an unterminated multi-line string would mask
the remainder of the file, causing the retired-architecture scan to report
clean on genuinely contaminated config — the gate failing **open**, silently.
This is the highest-severity risk in this plan. Mitigation: U4 AC-1 requires a
positive fixture (a marker that **must** still be reported after a multi-line
string closes), so over-masking is caught by a failing self-test rather than by
an absent finding.

**R3 — Bash portability.** The scan runs under `bash` on the CI runner; the
state machine must avoid non-POSIX constructs not already used by the script.
Keep to the existing idiom set.

**R4 — `build.ps1` resolution cost.** Symlink resolution adds filesystem calls
to every build. Negligible (a handful of `stat`-equivalents), and the guard
already performs path resolution.

## Plan Hardening Signals

| Signal | Present | Note |
|---|---|---|
| Schema/contract change | No | — |
| Security-sensitive behaviour | **Yes** | U1 is a containment control; U3 affects a security-relevant scan. |
| Migration | No | — |
| External dependency | No | — |
| Fail-open potential | **Yes** | R2 — over-masking blinds the retired-architecture gate. |

**Requires plan hardening: yes.**

---

# Plan Hardening

**Hardening required: YES** (security-sensitive containment control; fail-open potential). Hardened 2026-09-04.

## Context consulted

`scripts/build.ps1` (I5 guard), `scripts/check-retired-architecture.sh` (`strip_toml_comment`, `--self-test` L273-282), `internal/pathsafe/pathsafe.go:214` (`checkSymlinkEscape`, the proven ancestor-walking pattern U1 mirrors), and the 004-S closure record deferring both findings.

## Protected invariants

* **I5 — build output must remain inside the repo root.** U1 may only make the guard *stricter*, never looser.
* **The retired-architecture scan must not report clean on contaminated config.** U3 must not blind it.
* **Existing single-line masking behaviour is unchanged.** Pre-existing self-test cases must pass untouched.
* **Existing non-symlink negative tests keep passing.**

## Risky actions (ProposedAction / ActionRisk)

### PA-1 — Rewrite `strip_toml_comment` into a cross-line state machine (U3)
**Risk: HIGH (fail-open, silent).** An over-masking bug makes the scanner report clean on genuinely contaminated config. The failure presents as an *absence* of findings, which is indistinguishable from success without a positive control.
**Control:** U4 AC-1 mandates a positive fixture — a retired marker after a closed multi-line string that **must still be reported**. Landing U3 without that fixture green is forbidden. Fixtures are authored first (U4 → U3) and must be demonstrably failing before the logic change.

### PA-2 — Modify a containment guard (U1)
**Risk: MEDIUM.** A resolution bug could reject legitimate builds (halts delivery) or accept an escape (breaches containment).
**Control:** U2 requires both a negative (escape rejected) and a positive control (in-root symlink accepted), so neither failure direction passes silently.

### PA-3 — Symlink-dependent tests in an unattended runner
**Risk: MEDIUM (false confidence).** A silent skip reads as a pass; the guard would appear tested while never being exercised.
**Control:** U2 AC-4 / U4 require the skip to be **explicit and visible in CI output**. Ship must confirm the skip line appears, not merely that the suite is green.

### Not risky — explicitly classified
Fixture files in isolation; the rejection-message wording change in U1 AC-4.

## Rollback points

U1+U2 and U3+U4 are two independent pairs; either pair reverts without affecting the other. Neither pair changes any interface consumed elsewhere.

## Deepened runtime verification

1. `--self-test` green on all fixtures, old and new.
2. **Negative control:** deliberately break the masking state machine and confirm the positive fixture fails. A masking change whose failure mode has not been observed is unverified.
3. Symlink escape test observed **running** (not skipped) on at least one platform before the shipment closes; if it skips everywhere, the unit is returned blocked rather than accepted.
4. A normal build still produces output in the expected in-root location.

## Human checkpoints

None required. If U2's symlink test cannot execute on any available platform, Ship returns that unit **blocked** for operator input rather than accepting an unexercised containment guard.
