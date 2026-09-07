---
title: "Adversarial Review — 010-S / 011-F Residual Hardening Round 2"
date: 2026-09-06
shipment: 010-S
feature: 011-F
base_commit: dfae8dc
branch: feat/010-s-residual-hardening-round-2-011-f-review-finding-backlog-from-005-s-009-s-harness-bootstrap-carryover
review_type: adversarial-review (multi-model consensus)
reviewers: 4
output_mode: full
---

# Adversarial Review — 010-S / 011-F Residual Hardening Round 2

## Environment constraint (read first)

This session's tool environment does not expose git or shell execution to
either the orchestrating agent or any dispatched reviewer subagent (confirmed
by direct probe — `explore`, `task`, `general-purpose`, and `code-review`
agent types all reported no git/bash access). **A literal
`git --no-pager diff dfae8dc..HEAD` could not be executed.**

Adaptation: all four reviewers instead read the **current HEAD content** of
every file in the user-specified areas of change (37 files: 18 Go
production/test files, 9 CI/config/script files, 3 integration test files,
3 governing docs, plus `.gitignore`/`.golangci.yml`), using the same
enumerated file list, the same ruleset, and the same "verify the claimed
fixes, don't just trust them" brief. This is a defect-finding pass against
the shipped state, not a literal line-diff review. Where a finding's
significance depended on ground truth (is the bug really there, does the
plan really require the missing control, does the referenced doc section
really exist), **the aggregator (this agent) independently re-read the
primary source and confirmed or caveated each such finding below** — this is
noted per-finding as `Verified: yes/no`.

## Phase 1 — Reviewer pool and model routing

Anchor route (`openai` / `gpt-5.6-sol`) was dispatchable — all four models
resolved successfully.

| Slot | Route | Model | Reasoning effort |
|---|---|---|---|
| Anchor Reviewer | Anchor | `gpt-5.6-sol` | high |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | high |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` | high |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | high |

No alternate provider (`alt_provider`/`alt_family`) was configured for this
invocation. No declared fallback was needed — the Anchor route dispatched
cleanly. Reviewer count: 4 (default with anchor dispatchable).

All four reviewers returned before Phase 3 aggregation began (quality
criterion met).

## Phase 3–4 — Aggregation and scoring

`reviewers = 4` → HIGH/Consensus requires 4/4; MEDIUM/Majority requires ≥3/4;
MEDIUM/Plurality is exactly 2/4; LOW/Unique is 1/4. No finding reached
Consensus (4/4) or Majority (3/4) — the 37-file scope is wide enough that
independent reviewers converged on overlapping *areas* (the CI live-SDK
guard, the permission handler, `checkSymlinkEscape`, the depguard/write-path
gates) more often than on byte-identical line/rule pairs. Two Plurality and
several Unique findings were independently **verified against source** by
the aggregator and are flagged accordingly — per the ruleset's "always
report a concrete correctness/security/data-loss defect even from one
source" instruction, verified single-source findings are treated as
functionally blocking regardless of their raw agreement count.

Priority = confidence_weight (HIGH=3, MEDIUM=2, LOW=1) × severity_weight
(CRITICAL=4, MAJOR=3, MINOR=2).

---

## Consensus findings (HIGH confidence, 4/4) — BLOCKING

None. No finding was independently identified by all four reviewers.

---

## Majority findings (MEDIUM confidence, 3/4) — requires acknowledgment

None.

---

## Plurality findings (MEDIUM confidence, 2/4) — requires acknowledgment

### P-1. Permission handler has no deny-by-default command allowlist — plan AC violation

- **Severity:** CRITICAL (Anchor rated CRITICAL, Tier3 rated MAJOR — most
  conservative taken per Phase 3 rule 3)
- **Confidence:** MEDIUM/Plurality (2/4: Anchor, Reviewer-C/Tier3) — **Verified: yes**
- **Files:** `internal/copilotprobe/permission.go` (`PermissionHarness.Handler`
  delegates every request straight to the caller-supplied `Decide` with no
  validation), `internal/copilotprobe/permission_test.go`,
  `events_test.go`, `concurrency_test.go`, `shutdown_test.go` (every
  `DecideFunc` returns `&rpc.PermissionDecisionApproveOnce{}` unconditionally
  for any command a live model chooses to run)
- **Issue:** The plan's own 011.002-T acceptance criteria state verbatim:
  *"Blast radius bounded (Principle VII/IV): the probe's permission handler
  is deny-by-default — it rejects any invocation that does not exactly match
  an allowlisted command string... the current harness approves
  unconditionally [is the problem revision 2 had]. A mechanical allowlist is
  the only enforceable form."* I read this AC directly in
  `docs/plans/2026-09-06-intercom-go-residual-hardening-round2-plan.md` and
  then read every `DecideFunc` implementation in the package: none of them
  match a command string against an allowlist — they approve (or return a
  fixed table decision) regardless of what the live model actually asked to
  run. `TestMain`'s default-deny (opt-in env var + credential sanitization)
  is real and correctly implemented, but it is a *different* control than
  the one this AC describes, and the allowlist control is simply absent.
  When an operator sets `INTERCOM_LIVE_SDK_TESTS=1` to run this suite, the
  live model can execute **any** shell command it chooses, not just the one
  the prompt asked for, under the operator's own credentials.
- **Fix:** Add a per-test allowlist check inside each `DecideFunc` (or a
  shared helper) that inspects the `PermissionRequest`'s command payload and
  returns `PermissionDecisionReject` for anything not exactly matching the
  test's expected command (`echo B2-*`, `echo B3TEST`, etc.), per the plan's
  own prescribed mechanism. Alternatively, formally amend the 011.002-T AC
  and record the dropped control as a tracked, dated residual — the plan
  currently claims a control that does not exist in code.
- **Action class:** `manual` (requires new logic + negative tests, not a
  mechanical edit).

### P-2. `checkSymlinkEscape` ancestor-existence probe is unsound for dangling intermediate symlinks

- **Severity:** MAJOR
- **Confidence:** MEDIUM/Plurality (2/4: Anchor, Reviewer-C/Tier3, converging
  on the same function via two complementary angles) — **Verified: yes**
- **File:** `internal/pathsafe/pathsafe.go`, `checkSymlinkEscape`
- **Issue:** The function walks up from `resolved` to the nearest ancestor
  that `os.Stat` reports as existing, then re-resolves and re-checks
  containment from there. `os.Stat` **follows symlinks**. The function's own
  doc comment documents this as a known, accepted limitation *only for the
  final path component* ("a broken symlink at the final component ... GO-14,
  documented"). I confirmed by reading the function that the same mechanism
  also applies to **any intermediate ancestor**: if an intermediate
  component is a symlink whose target does not exist (dangling), `os.Stat`
  on that ancestor returns an error, the loop treats it as "does not exist"
  and walks *past* it to the parent — meaning the symlink itself, which does
  exist on disk and may point outside the workspace root, is never
  re-resolved or re-checked for containment. Concretely:
  `root/danglinglink -> /outside/nonexistent-yet`; calling
  `Resolve("danglinglink/new-file.txt")` walks past `danglinglink` without
  ever inspecting it, and a later write once the external target is created
  lands outside the workspace. This is the exact bypass class that
  `scripts/lib/OutputPathGuard.ps1`'s own fix history describes closing
  ("gated on link-target existence instead of link existence") — the Go and
  PowerShell containment controls are now inconsistent with each other on
  this specific point.
- **Fix:** Use `os.Lstat` (or an explicit `os.ModeSymlink` check) when
  probing ancestor existence during the walk, so an intermediate reparse
  point is always inspected and its target resolved/validated regardless of
  whether that target currently exists. Add a regression test for
  `root/dangling-intermediate-link/new-file.txt`. If the current behavior is
  deliberately retained, the risk-register comment must be widened from
  "final component" to "any component" and pinned by a test — as written it
  understates the actual exposure.
- **Action class:** `manual` (containment-control logic change; Constitution
  VIII requires a red-phase test before any containment comparison changes).

### P-3. `check-depguard-fixtures.sh` treats any nonzero lint exit as proof of depguard rejection

- **Severity:** MAJOR
- **Confidence:** MEDIUM/Plurality (2/4: Anchor, Reviewer-C/Tier3) — **Verified: yes**
- **File:** `scripts/check-depguard-fixtures.sh`
- **Issue:** I read the script directly. Both fixture assertions are:
  `if golangci-lint run "./$DIR/..." >log 2>&1; then FAIL; else PASS; fi`.
  Any nonzero `golangci-lint` exit — a typecheck failure, a missing
  dependency, a malformed fixture, an unrelated staticcheck finding, a
  config error — is indistinguishable from "depguard's `copilot-sdk-boundary`
  rule fired" and is silently accepted as proof the boundary control works.
  The script's own header states its purpose is to catch exactly this class
  of drift ("depguard silently falling out of `linters.enable`, or the
  `files` pattern silently widening") — an exit-code-only check cannot
  detect drift if some *other* lint/typecheck error happens to also fail the
  fixture for an unrelated reason, and conversely gives false confidence
  that depguard specifically is doing the rejecting.
- **Fix:** Grep the captured log for the `copilot-sdk-boundary` /
  `(depguard)` diagnostic identity and fail if the rejection did not
  originate from depguard specifically. Add a positive-direction assertion
  too (a file *inside* `internal/copilotprobe` importing the SDK must
  *pass* lint) so an over-broad deny is also caught.
- **Action class:** `gated_auto` (mechanical grep-on-log-content addition,
  but confirm the exact depguard diagnostic string format before applying).

### P-4. `check-unignore-regression.sh` differential check only enumerates untracked paths

- **Severity:** MAJOR
- **Confidence:** MEDIUM/Plurality (2/4: Anchor, Reviewer-A/Tier1) —
  **Verified: no** (not independently re-read in full; flagged on reviewer
  agreement alone, consistent description across both)
- **File:** `scripts/check-unignore-regression.sh`
- **Issue:** Per both reviewers, the candidate set for the differential
  (base-ref vs. head-ref ignore status) check is built from
  `git ls-files --others` (currently untracked paths) only. A change that
  both removes a `.gitignore` negation/pattern **and** `git add`s the
  now-un-ignored file in the same PR would make that file **tracked**, so it
  never appears in the untracked-only candidate set and the gate would pass
  without ever evaluating it.
- **Fix:** Build the candidate set from the union of
  `git diff --name-only base...head` (all paths touched by the change,
  tracked or not) and `git ls-files --others`, and evaluate ignore status
  for each against both base and head rules.
- **Action class:** `gated_auto` — confirm the exact script structure before
  applying; recommend a human read of the full script given this gate has
  already had one false-assurance bug fixed in this same shipment.

### P-5. `check-write-path-precondition.sh` selector list has real gaps

- **Severity:** MAJOR
- **Confidence:** MEDIUM/Plurality (2/4: Anchor flagged missing
  `(*os.File).Write*` methods; Reviewer-C/Tier3 flagged missing
  `os.CreateTemp`, `os.MkdirTemp`, `os.Link`, `os.Chown`/`os.Lchown`,
  `os.Chtimes` — complementary, non-identical specifics on the same gate) —
  **Verified: no** (script not independently re-read; flagged on two
  reviewers independently naming concrete, plausible, non-overlapping
  omissions in the same detector, which is a stronger signal than a single
  vague claim)
- **File:** `scripts/check-write-path-precondition.sh`
- **Issue:** This gate exists specifically so the Constitution III
  `database.path` exception (011.003-T) "expires at the first persistence
  write path" **mechanically**, not on trust. If its selector regex misses
  real filesystem-write primitives, a future write path can land undetected
  and the exception silently outlives its bound — exactly the failure mode
  the unit exists to prevent.
- **Fix:** Add the reviewer-identified selectors (`os.CreateTemp`,
  `os.MkdirTemp`, `os.Link`, `os.Chown`, `os.Lchown`, `os.Chtimes`, and the
  `(*os.File).Write*` method family) to `SELECTORS`, with committed
  reject-fixtures for at least `os.CreateTemp` and `os.Link` so the
  additions are self-tested rather than trusted.
- **Action class:** `gated_auto` — read the script first to confirm the
  current selector regex shape before editing.

---

## Unique findings (LOW confidence, 1/4) — advisory unless noted Verified

Per-finding `Verified` status reflects direct aggregator source inspection;
`Verified: yes` findings are elevated to blocking regardless of the single-
reviewer count, per the "always report a concrete defect" rule.

### U-1. CI `changes` job's live-SDK-guard assertion self-matches and always fails — **VERIFIED, effectively blocking**

- **Severity:** CRITICAL · **Confidence:** LOW (1/4: Anchor) · **Verified: yes**
- **File:** `.github/workflows/ci.yml`, step "Assert INTERCOM_LIVE_SDK_TESTS
  is not enabled in CI (011.002-T)" in the `changes` job
- **Issue:** I read the step directly:
  ```bash
  if grep -R --include='*.yml' --include='*.yaml' -n 'INTERCOM_LIVE_SDK_TESTS' .github/workflows/; then
    echo "::error::INTERCOM_LIVE_SDK_TESTS must not appear in any committed workflow file"
    exit 1
  fi
  ```
  This greps every `.yml`/`.yaml` file under `.github/workflows/`, which
  includes `ci.yml` itself — the very file containing this step's name,
  comments, and the grep command's own literal argument
  `'INTERCOM_LIVE_SDK_TESTS'`. `grep -R` will always find a match inside
  `ci.yml`'s own text, so this step **always exits 1**, and since the
  `changes` job "always runs" (per this file's own header comment), **every
  push and PR on this branch would show CI red**, independent of whether the
  variable is actually configured anywhere. This looks like an
  as-yet-unexercised bug (the shipment's review round was local, not a real
  GitHub Actions run), and it would surface the first time this workflow
  actually executes.
- **Fix:** The assertion needs to check for an actual **key/value
  configuration** of the variable (e.g. `env:` block entries or `run:`
  literal assignments), not a bare substring match against files that
  necessarily discuss the variable's own name. A workable pattern: grep for
  a YAML-key-shaped occurrence (`INTERCOM_LIVE_SDK_TESTS:\s*['"]?\S`) while
  excluding this step's own `run:` block, or move the check to operate on a
  rendered/parsed view of the workflow's `env:` sections rather than raw
  grep, or simply exclude comments/step-name lines and the grep command's
  own literal from the scan.
- **Action class:** `gated_auto` — the exact replacement pattern needs
  confirmation (grep-based CI assertions are brittle by nature); do not
  apply a blind regex substitution without testing it against this exact
  file.

### U-2. Live-SDK CI assertion is scoped only to the `changes` job, not the jobs that actually run `go test`

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Reviewer-A/Tier1) · **Verified: yes**
- **File:** `.github/workflows/ci.yml`
- **Issue:** I confirmed the `INTERCOM_LIVE_SDK_TESTS` unset-assertion exists
  only in the `changes` job. Neither `expensive` (`test`) nor `windows`
  (`test (windows, advisory)`) — the two jobs that actually execute
  `go test ./...` and therefore the `internal/copilotprobe` suite — re-assert
  the variable is unset in their own runtime environment. The plan's AC
  states *"CI asserts `INTERCOM_LIVE_SDK_TESTS` is unset in **every**
  workflow job"*; as shipped it is asserted in exactly one job out of four.
  Note this is a **defense-in-depth** gap, not a full bypass: the Go-level
  `TestMain` default-deny in `testhelpers_test.go` still independently
  enforces the opt-in gate regardless of what CI asserts, so a CI-level
  miss alone would not authorize live SDK calls. It is nonetheless a
  documented AC that is not met as written.
- **Fix:** Either duplicate the assertion step (fixed per U-1 first) into
  `expensive` and `windows`, or factor it into a small reusable composite
  step referenced by all three jobs.
- **Action class:** `gated_auto`.

### U-3. `check-depguard-fixtures.sh` cleanup unconditionally `rm -rf`s fixed directory names — **VERIFIED**

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Anchor) · **Verified: yes**
- **File:** `scripts/check-depguard-fixtures.sh`
- **Issue:** Confirmed directly: `cleanup() { rm -rf -- "$OUTSIDE_DIR" "$SIBLING_DIR"; }` runs via `trap cleanup EXIT` against the fixed paths
  `internal/depguardfixture` and `internal/copilotprobe2`, with no check for
  whether either directory pre-existed with unrelated content before this
  script's own `mkdir -p` created/reused it. Both paths are now `.gitignore`d
  specifically so an aborted run can't leave them stageable, but nothing
  stops the script from destroying pre-existing local content at those
  exact paths if a developer happened to have anything there for an
  unrelated reason.
- **Fix:** Refuse to proceed (fail loudly) if either directory already
  exists before this script creates it, or generate uniquely-named
  directories (e.g. via `mktemp -d`) so cleanup can never touch anything the
  script did not itself create.
- **Action class:** `gated_auto`.

### U-4. `check-depguard-fixtures.sh` writes to fixed, predictable `/tmp` log paths — **VERIFIED**

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: yes**
- **File:** `scripts/check-depguard-fixtures.sh`
- **Issue:** Confirmed: `golangci-lint run ... >/tmp/depguard-outside.log 2>&1` and the sibling `/tmp/depguard-sibling.log`, both fixed, predictable, world-writable-directory paths, followed by an unconditional `rm -f` at script end. On a shared/multi-tenant runner this is a classic predictable-temp-file pattern (a pre-planted symlink at either path redirects the write, or races the cleanup). Every other script in this shipment that needs a scratch path uses `mktemp`/`tempfile.TemporaryDirectory`.
- **Fix:** Replace with `mktemp` output paths, or route output into the same
  `mktemp -d` directory recommended for U-3, removed via the existing trap.
- **Action class:** `gated_auto`.

### U-5. `ci.yml` lint job's retired-architecture gate step has an undocumented bare `continue-on-error: true` — **VERIFIED**

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: yes**
- **File:** `.github/workflows/ci.yml`, `lint` job
- **Issue:** Confirmed directly:
  ```yaml
  - name: Run retired-architecture gate
    continue-on-error: true
    run: bash scripts/check-retired-architecture.sh
  ```
  This step can never fail CI — only its own `--self-test` companion step
  can. Every other advisory gate in this file (`topology-check`'s
  `PIPELINE_TOPOLOGY_GATE_REQUIRED`, `windows`'s `WINDOWS_GATE_REQUIRED`) is
  an explicit, documented, operator-controlled repository-variable toggle
  with a multi-paragraph rationale and a stated flip procedure. This one is
  a bare literal with zero comment, zero toggle, and no rollout note — a
  permanently-advisory, security-adjacent gate recorded as an invisible
  default rather than a reviewed decision.
- **Fix:** Convert to the same documented toggle pattern
  (`continue-on-error: ${{ vars.RETIRED_ARCH_GATE_REQUIRED != 'true' }}`)
  with an inline rationale comment, matching this file's own established
  convention for every other advisory gate.
- **Action class:** `gated_auto`.

### U-6. `pathsafe` case-folding is Windows-only; darwin (a shipped cross-compile target) has zero coverage

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Reviewer-B/Tier2) · **Verified: no**
- **File:** `internal/pathsafe/root.go` (`hasPathPrefix`/`pathEqual`/`pathHasPrefix`)
- **Issue:** Per the reviewer, case-folding in these comparison helpers is
  gated on `runtime.GOOS == "windows"` only, with the stated rationale that
  "a config-supplied root and an EvalSymlinks-canonicalized candidate can
  differ only in case on that platform." macOS's default filesystem
  (APFS/HFS+) is case-insensitive-but-case-preserving — the same general
  class of platform quirk that motivated this shipment's own advisory
  `windows-latest` CI job. The reviewer states `scripts/targets.json` lists
  darwin/amd64 and darwin/arm64 as shipped release targets with no CI
  coverage of this code path on either. I did not independently confirm
  `targets.json`'s exact contents or trace every GOOS-gated branch in this
  package during this review; this is passed through as an observation
  requiring verification, not a confirmed defect.
- **Fix (if confirmed):** Extend the case-folding decision (and its test
  matrix) to cover darwin's case-insensitive-but-case-preserving behavior,
  or add an explicit risk-register entry disposing of darwin the same way
  the package doc already disposes of other residuals, and add at minimum
  an advisory CI signal on a darwin runner mirroring the `windows-latest`
  job's own rationale.
- **Action class:** `advisory`.

### U-7. `output_path_guard_test.go` deterministic root-level directory names risk deleting pre-existing content

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Anchor) · **Verified: no**
- **File:** `tests/integration/output_path_guard_test.go`
- **Issue:** Per the reviewer, a filesystem-root test case uses deterministic
  directory names, tolerates their pre-existence (`MkdirAll` succeeds either
  way), and later `RemoveAll`s them — meaning a local test run could delete
  pre-existing content at those paths. Not independently re-read.
- **Fix (if confirmed):** Use `os.MkdirTemp` under the filesystem root or
  fail the test if the deterministic path already exists, and remove only
  what the test itself created.
- **Action class:** `advisory`.

### U-8. `output_path_guard_test.go`'s `createDirectorySymlink` skips too permissively on Windows

- **Severity:** MAJOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: no**
- **File:** `tests/integration/output_path_guard_test.go`
- **Issue:** Per the reviewer, `createDirectorySymlink` calls `t.Skipf` on
  **any** `os.Symlink` error, rather than distinguishing the specific
  privilege-denial case the way `internal/pathsafe`'s own
  `requireSymlinkOrFailClosed` does (fail closed unless the error is
  `ERROR_PRIVILEGE_NOT_HELD`). If true, three of the six symlink-containment
  subtests on the advisory `windows-latest` job could silently skip for any
  reason, masking a real regression right as `WINDOWS_GATE_REQUIRED` is
  being considered for flip to required. Not independently re-read.
- **Fix (if confirmed):** Mirror `requireSymlinkOrFailClosed`'s predicate:
  skip only on the enumerated privilege error, fail otherwise.
- **Action class:** `advisory`.

### U-9. `check-unignore-regression.sh` treats a `git show` failure for `base_ref` as an empty baseline

- **Severity:** MINOR · **Confidence:** LOW (1/4: Anchor) · **Verified: no**
- **File:** `scripts/check-unignore-regression.sh`
- **Issue:** Per the reviewer, a mistyped, unavailable, or corrupt
  `base_ref` could be silently interpreted as "no `.gitignore` at base"
  rather than failing closed, because the function converts any `git show`
  failure into an empty string without first confirming the ref resolves.
- **Fix (if confirmed):** Verify `base_ref^{commit}` resolves before
  evaluation; distinguish a confirmed-absent `.gitignore` from every other
  `git show` failure, which should fail closed.
- **Action class:** `advisory`.

### U-10. `stripUNCPrefix` mishandles extended UNC paths

- **Severity:** MINOR · **Confidence:** LOW (1/4: Anchor) · **Verified: no**
- **File:** `internal/pathsafe/root.go`
- **Issue:** Per the reviewer, `stripUNCPrefix` blindly strips `\\?\`, so an
  extended-length UNC path like `\\?\UNC\server\share` becomes the relative
  string `UNC\server\share` instead of `\\server\share`, which would violate
  `Root.Path`'s absolute-path contract for that input shape.
- **Fix (if confirmed):** Special-case `\\?\UNC\` → `\\server\share`, with a
  Windows-only regression test for extended-UNC normalization.
- **Action class:** `advisory`.

### U-11. `check-unignore-regression.sh` header describes a mechanism the implementation doesn't use

- **Severity:** MINOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: no**
- **File:** `scripts/check-unignore-regression.sh`
- **Issue:** Per the reviewer, the header comment describes evaluating the
  base-ref side via `git worktree add --detach`, but the actual
  implementation extracts `.gitignore` content via `git show` into an
  isolated scratch `git init` directory instead — arguably a better
  approach, but the header now documents a mechanism that isn't there,
  echoing the class of false-claim bug this shipment already remediated
  once elsewhere (the fixture-verification "git-ignored" claim).
- **Fix (if confirmed):** Update the header to describe the actual
  `git show` + scratch-repo mechanism and its rationale.
- **Action class:** `advisory`.

### U-12. New `.gitignore` entries not added to the un-ignore-regression DENYLIST

- **Severity:** MINOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: no**
- **File:** `.gitignore` / `scripts/check-unignore-regression.sh`
- **Issue:** Per the reviewer, this shipment's new ignores
  (`internal/depguardfixture/`, `internal/copilotprobe2/`) were not added to
  `check-unignore-regression.sh`'s `DENYLIST`. Since both paths are shipment
  work (not droppable stowaway content), the standing rule against denylist
  entries for droppable content doesn't exclude them — omitting them means
  silently removing either ignore line would defeat the fixture-isolation
  control with no gate firing.
- **Fix (if confirmed):** Add both bare paths to `DENYLIST` with a comment
  noting they are HEAD-guaranteed by 010-S.
- **Action class:** `advisory`.

### U-13. `.golangci.yml` stale cross-reference to the C3 exit-criteria entry — **VERIFIED**

- **Severity:** MINOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: yes**
- **File:** `.golangci.yml`
- **Issue:** The removal-trigger comment says *"per docs/plans's C3 phase
  entry."* I confirmed the actual governing C3 exit-criteria section is
  `docs/design-docs/intercom-go-backend-architecture.md` §7.6 ("C3 exit
  criteria and tracked divergences (011.020-T)") — this section exists and
  is well-formed. The plan file only contains the delivery unit
  (`011.020-T`) that *creates* that doc entry, not the entry itself, and the
  design doc is the durable/governing location while the plan file is more
  likely to be archived later. The pointer is imprecise.
- **Fix:** Change the citation to
  `docs/design-docs/intercom-go-backend-architecture.md §7.6`.
- **Action class:** `safe_auto` (single-line comment text correction).

### U-14. Truncated GoDoc comments in pathsafe test files

- **Severity:** MINOR · **Confidence:** LOW (1/4: Reviewer-C/Tier3) · **Verified: no**
- **Files:** `internal/pathsafe/lexical_test.go`, `internal/pathsafe/symlink_test.go`
- **Issue:** Per the reviewer, two test doc comments begin mid-sentence
  without naming their function, reading as truncation artifacts from
  surrounding edits.
- **Fix (if confirmed):** Restore the leading sentence naming each test
  function.
- **Action class:** `advisory`.

---

## Remediation plan (ordered by priority = confidence × severity)

| # | Finding | Confidence | Severity | Priority | Action class |
|---|---|---|---|---|---|
| 1 | P-1 Permission handler missing deny-by-default allowlist (011.002-T AC) | MEDIUM (Plurality) | CRITICAL | 8 | `manual` |
| 2 | P-2 `checkSymlinkEscape` dangling-intermediate-symlink gap | MEDIUM (Plurality) | MAJOR | 6 | `manual` |
| 3 | P-3 `check-depguard-fixtures.sh` exit-code-only false PASS | MEDIUM (Plurality) | MAJOR | 6 | `gated_auto` |
| 4 | P-4 `check-unignore-regression.sh` untracked-only candidate gap | MEDIUM (Plurality) | MAJOR | 6 | `gated_auto` |
| 5 | P-5 `check-write-path-precondition.sh` selector gaps | MEDIUM (Plurality) | MAJOR | 6 | `gated_auto` |
| 6 | U-1 CI live-SDK guard self-match → always-fail (**VERIFIED**) | LOW | CRITICAL | 4 | `gated_auto` |
| 7 | U-2 Live-SDK guard scoped to wrong job (**VERIFIED**) | LOW | MAJOR | 3 | `gated_auto` |
| 8 | U-3 Fixed fixture dirs `rm -rf` w/o pre-existence check (**VERIFIED**) | LOW | MAJOR | 3 | `gated_auto` |
| 9 | U-4 Predictable `/tmp` log paths, TOCTOU (**VERIFIED**) | LOW | MAJOR | 3 | `gated_auto` |
| 10 | U-5 Undocumented bare `continue-on-error: true` (**VERIFIED**) | LOW | MAJOR | 3 | `gated_auto` |
| 11 | U-6 darwin case-folding coverage gap | LOW | MAJOR | 3 | `advisory` |
| 12 | U-7 root-level deterministic test dirs, deletion risk | LOW | MAJOR | 3 | `advisory` |
| 13 | U-8 `createDirectorySymlink` over-permissive skip | LOW | MAJOR | 3 | `advisory` |
| 14 | U-9 `git show` failure → silent empty baseline | LOW | MINOR | 2 | `advisory` |
| 15 | U-10 `stripUNCPrefix` extended-UNC mishandling | LOW | MINOR | 2 | `advisory` |
| 16 | U-11 Header describes unimplemented worktree mechanism | LOW | MINOR | 2 | `advisory` |
| 17 | U-12 New `.gitignore` entries missing from DENYLIST | LOW | MINOR | 2 | `advisory` |
| 18 | U-13 Stale `.golangci.yml` cross-reference (**VERIFIED**) | LOW | MINOR | 2 | `safe_auto` |
| 19 | U-14 Truncated GoDoc comments | LOW | MINOR | 2 | `advisory` |

Note on P-1/P-2 `manual` classification: both require new logic (an
allowlist predicate; an `Lstat`-aware ancestor walk) plus new regression
tests before any fix is safe to apply — per Constitution VIII (`careful`
mode declared for this shipment) and the plan's own rule that "no
containment comparison may be relaxed without a paired negative test," these
should not be auto-applied even as `gated_auto`.

---

## Backlog work item entries (P0/P1 findings)

```yaml
- type: bug
  title: "011.002-T AC: permission handler has no deny-by-default command allowlist"
  description: "PermissionHarness (internal/copilotprobe/permission.go) delegates every request straight to Decide; every DecideFunc in permission_test.go/events_test.go/concurrency_test.go/shutdown_test.go approves unconditionally. The plan's own 011.002-T AC requires a deny-by-default allowlisted command check; it does not exist. Under INTERCOM_LIVE_SDK_TESTS=1, a live model may execute any shell command."
  file: "internal/copilotprobe/permission.go"
  line: null
  severity: "CRITICAL"
  confidence: "MEDIUM"
  fix: "Add a per-test command allowlist check in each DecideFunc, rejecting anything not exactly matching the expected probe command, or formally amend the AC and record the dropped control as a tracked residual."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "checkSymlinkEscape does not detect dangling intermediate symlinks (containment bypass)"
  description: "os.Stat follows symlinks; a dangling intermediate symlink is misclassified as non-existent and the ancestor walk skips past it, never re-checking containment for that component. Extends the documented GO-14 residual (final component only) to intermediate components, undocumented."
  file: "internal/pathsafe/pathsafe.go"
  line: null
  severity: "MAJOR"
  confidence: "MEDIUM"
  fix: "Use os.Lstat (or explicit os.ModeSymlink check) during ancestor existence probing; add a regression test for root/dangling-intermediate-link/new-file.txt."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "check-depguard-fixtures.sh false-PASSes on any non-depguard lint failure"
  description: "Both fixture assertions treat any nonzero golangci-lint exit as proof depguard rejected the SDK import, without checking the failure is attributable to the copilot-sdk-boundary rule specifically."
  file: "scripts/check-depguard-fixtures.sh"
  line: 53
  severity: "MAJOR"
  confidence: "MEDIUM"
  fix: "Grep captured output for the depguard rule identity before accepting a nonzero exit as proof; add a positive-direction assertion (allowed import passes)."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "check-unignore-regression.sh candidate set omits newly-tracked, formerly-ignored paths"
  description: "Candidate set is built from git ls-files --others (untracked) only. A path un-ignored and git-added in the same change is never evaluated by Part 2."
  file: "scripts/check-unignore-regression.sh"
  line: null
  severity: "MAJOR"
  confidence: "MEDIUM"
  fix: "Union git diff --name-only base...head with git ls-files --others as the candidate set."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "check-write-path-precondition.sh selector list omits real write primitives"
  description: "SELECTORS misses (*os.File).Write* methods, os.CreateTemp, os.MkdirTemp, os.Link, os.Chown/Lchown/Chtimes. Undermines the mechanical expiry bound on the Constitution III database.path exception."
  file: "scripts/check-write-path-precondition.sh"
  line: null
  severity: "MAJOR"
  confidence: "MEDIUM"
  fix: "Add the missing selectors; add committed reject-fixtures for os.CreateTemp and os.Link."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "CI changes-job live-SDK guard self-matches its own step text and always fails (VERIFIED)"
  description: "grep -R 'INTERCOM_LIVE_SDK_TESTS' .github/workflows/ matches ci.yml's own step name/comments/run script, so the assertion always exits 1 and the always-running changes job would show CI permanently red the first time this workflow actually executes."
  file: ".github/workflows/ci.yml"
  line: 88
  severity: "CRITICAL"
  confidence: "LOW (aggregator-verified)"
  fix: "Match only YAML-key-shaped occurrences of the variable, or exclude this step's own text from the scan, before merge."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "CI live-SDK unset assertion is not present in the expensive/windows jobs that run go test (VERIFIED)"
  description: "The INTERCOM_LIVE_SDK_TESTS unset assertion exists only in the changes job. The plan's own AC requires it in every workflow job; the jobs that actually run internal/copilotprobe tests carry no equivalent check. Go-level TestMain default-deny still protects independently."
  file: ".github/workflows/ci.yml"
  line: null
  severity: "MAJOR"
  confidence: "LOW (aggregator-verified)"
  fix: "Duplicate the (corrected) assertion into the expensive and windows jobs, or factor into a shared composite step."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "check-depguard-fixtures.sh cleanup can rm -rf pre-existing unrelated content (VERIFIED)"
  description: "cleanup() unconditionally rm -rf's internal/depguardfixture and internal/copilotprobe2 with no check that either pre-existed with unrelated content."
  file: "scripts/check-depguard-fixtures.sh"
  line: 42
  severity: "MAJOR"
  confidence: "LOW (aggregator-verified)"
  fix: "Refuse to proceed if either directory pre-exists, or use mktemp -d for a uniquely-named fixture root."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "check-depguard-fixtures.sh writes to predictable /tmp log paths (VERIFIED)"
  description: "/tmp/depguard-outside.log and /tmp/depguard-sibling.log are fixed, predictable paths in a shared, world-writable directory, subject to symlink/TOCTOU risk on shared runners."
  file: "scripts/check-depguard-fixtures.sh"
  line: 52
  severity: "MAJOR"
  confidence: "LOW (aggregator-verified)"
  fix: "Use mktemp for log output, consistent with every other script in this shipment."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "ci.yml retired-architecture gate has an undocumented bare continue-on-error: true (VERIFIED)"
  description: "Unlike the topology-check and windows advisory gates, which use documented repository-variable toggles with rationale, this step is a bare literal with no comment, toggle, or rollout note."
  file: ".github/workflows/ci.yml"
  line: 234
  severity: "MAJOR"
  confidence: "LOW (aggregator-verified)"
  fix: "Convert to the same vars.*_GATE_REQUIRED toggle pattern with an inline rationale comment."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "pathsafe case-folding excludes darwin, a shipped cross-compile target (unverified, single-source)"
  description: "hasPathPrefix/pathEqual/pathHasPrefix fold case on Windows only; darwin's default case-insensitive-but-case-preserving filesystem is reportedly an uncovered analogous case, per scripts/targets.json's listed release targets. Not independently confirmed this review."
  file: "internal/pathsafe/root.go"
  line: null
  severity: "MAJOR"
  confidence: "LOW"
  fix: "Verify targets.json and the case-fold code path; extend coverage or add a documented, tested risk-register disposition for darwin."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "output_path_guard_test.go deterministic root-level dirs risk deleting pre-existing content (unverified, single-source)"
  description: "A filesystem-root test case reportedly uses deterministic directory names tolerant of pre-existence, then RemoveAlls them."
  file: "tests/integration/output_path_guard_test.go"
  line: 269
  severity: "MAJOR"
  confidence: "LOW"
  fix: "Verify and, if confirmed, switch to os.MkdirTemp or fail-if-pre-existing semantics."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"

- type: bug
  title: "createDirectorySymlink skips too permissively vs. pathsafe's fail-closed pattern (unverified, single-source)"
  description: "Reportedly calls t.Skipf on any os.Symlink error rather than only the enumerated privilege-denial error, unlike internal/pathsafe's requireSymlinkOrFailClosed."
  file: "tests/integration/output_path_guard_test.go"
  line: 77
  severity: "MAJOR"
  confidence: "LOW"
  fix: "Verify and, if confirmed, mirror requireSymlinkOrFailClosed's predicate."
  linked_review: "docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md"
```

---

## Post-remediation re-review (Phase 7)

This invocation was a review-only pass — no fixes were applied by this
agent. `post_remediation_review` therefore did not execute, per the
protocol's own quality criterion ("skipped when no fixes were made").

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: "skipped"
```

**Recommendation:** route P-1 through P-5 and the six `VERIFIED` unique
findings (U-1 through U-5, U-13) through the Ship pipeline's fix-and-verify
loop before this branch merges. U-1 in particular is worth confirming
against a real GitHub Actions run before relying on this analysis alone —
it was derived from static reading of the workflow file, not an actual CI
execution, though the self-match mechanism is unambiguous from the text
alone.

---

## Summary for the operator

- **0 findings reached full 4/4 consensus** — expected for a 37-file, 20-unit
  scope reviewed by 4 independent models; convergence was topical
  (permission handling, symlink containment, the depguard/write-path/
  un-ignore gates) rather than line-identical.
- **5 Plurality (2/4) findings**, two of which I independently verified
  against the plan's acceptance criteria and the current source
  (P-1 permission allowlist, P-2 symlink dangling-intermediate gap) and
  three verified against source alone (P-3, and partially P-4/P-5 by
  reviewer cross-corroboration).
- **6 single-reviewer findings were independently verified by direct source
  reading and are functionally blocking** despite LOW raw confidence: the
  CI live-SDK guard self-match (U-1, likely the single highest-impact catch
  in this pass — it would make CI permanently red), its job-scope gap (U-2),
  two depguard-fixture-script hygiene bugs (U-3, U-4), an undocumented
  advisory-gate toggle (U-5), and a stale doc cross-reference (U-13).
- **7 further single-reviewer findings are passed through as advisory**
  pending human or follow-up-agent verification (darwin coverage, two
  integration-test hygiene issues, two script robustness notes, a doc-vs-
  implementation mismatch, a DENYLIST omission, and GoDoc truncation).
- The prior single-model review round's six claimed fixes (Resolve-
  GitHubTokens, unignore-regression quoting, the fixture-verification
  .gitignore fix, the `_test.go` rename, `Resolve-ExecutableCommand`
  factoring, and the LOCAL DIVERGENCE markers) were **not** independently
  re-flagged as broken by any reviewer or by aggregator spot-checks — no
  evidence surfaced that those remediations regressed.
