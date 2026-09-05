---
title: "Adversarial Review (Pass 2) — CI supply-chain and gate integrity hardening (007-F)"
date: 2026-09-04
branch: feat/006-s-ci-supply-chain-hardening
shipment: 006-S
feature: 007-F
mode: report-only
pass: 2
prior_pass: "docs/plans/2026-09-04-intercom-go-ci-supply-chain-hardening-plan.md (Post-Review Remediation Record)"
---

# Adversarial Review (Pass 2) — 007-F CI Supply-Chain Hardening

**Scope**: implementation on `feat/006-s-ci-supply-chain-hardening` vs `main` (6 commits, no Go
source changes) against the already-hardened plan at
`docs/plans/2026-09-04-intercom-go-ci-supply-chain-hardening-plan.md`.

**Files reviewed**: `.github/constraints/autoharness-lock.txt`, `.github/workflows/ci.yml`,
`.github/workflows/secret-scan-history.yml`, `.github/CODEOWNERS`,
`scripts/check-gitignore-append-only.sh`, `scripts/testdata/gitignore/{insertion,deletion,reorder}/*`.

**Purpose of this pass**: the FIRST adversarial pass (documented in the plan's Post-Review
Remediation Record) already forced remediation of U6 cross-shipment ordering, U3's false
self-modifying-PR-gap closure claim, U1's non-hermetic `==` pin, and U4's fail-open base
resolution. This pass reviews the **actual implementation** against that hardened plan, hunting
for anything the first pass — which reviewed the plan, not the code — could not have caught.

## Reviewer panel

4 independent model instances, dispatched in parallel, each given identical file contents and
identical focus-area instructions, each returning structured JSON findings only:

| Slot | Route | Model |
|---|---|---|
| Anchor Reviewer | Anchor review route | `gpt-5.6-sol` (reasoning effort: high) |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` |

No alternate-provider substitution was requested for this invocation; no declared fallback was
needed — all 4 slots dispatched successfully. All 4 reviewers returned before aggregation began.

## Verdict

# **FAIL — blocking remediation required before merge**

The blocking items are small and mechanical (no redesign needed): one CODEOWNERS coverage gap
that directly contradicts the plan's own AC-1 language, and one hash-pinning hermeticity gap that
undermines U1 AC-2's central claim. Everything else in the queue below is MAJOR/MINOR and can be
fixed in the same pass or explicitly deferred with rationale per the plan's own conventions.

---

## 1. Consensus findings (confidence: HIGH — flagged by all 4 reviewers)

### C-1. CODEOWNERS does not cover its own stated scope — the hash-pin lock file is unowned
**Severity: CRITICAL | Confidence: HIGH (4/4) | Priority score: 3×4 = 12**

`.github/CODEOWNERS` states: *"Covers ALL CI-critical scripts and workflow surfaces, or none
(partial ownership coverage here would be a misleading control)."* All four reviewers
independently identified that this claim is false as shipped:

- `.github/constraints/autoharness-lock.txt` — the hash-pinned dependency manifest that
  `--require-hashes` trusts, arguably the single most CI-critical **non-workflow** artifact this
  shipment introduces — matches no CODEOWNERS pattern.
- `scripts/testdata/gitignore/**` — the fixtures that define pass/fail semantics for the I6
  checker — also matches no pattern.

This is not cosmetic: the entire point of U3 is ownership metadata as a *prerequisite* for a
future enforcing control. If code-owner review is ever turned on (an explicit future operator
decision per R1), a PR could still edit the lock file's pinned hashes/versions — the actual
trust boundary of U1 — without triggering any ownership signal, while every other CI-critical
path does. The plan's own AC-1 text ("or none") makes this a direct acceptance-criterion failure,
not a nice-to-have, and it is a regression class similar to the false-completeness claims the
first pass already forced out of U3's self-modifying-PR-gap language.

**Fix**: add `/.github/constraints/ @softwaresalt` and `/scripts/testdata/ @softwaresalt` (or the
gitignore-specific subdirectory) to CODEOWNERS. Trivial, one-line-per-path fix.

---

## 2. Majority findings (confidence: MEDIUM — flagged by 3 of 4 reviewers)

### M-1. `--require-hashes` is not end-to-end hermetic
**Severity: CRITICAL | Confidence: MEDIUM/Majority (3/4: Anchor, Reviewer-A, Reviewer-C) | Priority: 2×4 = 8**

Three independent reviewers converged on the same underlying gap via different mechanisms:

- The install runs in the runner's **ambient** pip environment. `--require-hashes` verifies
  packages pip actually decides to fetch/install; it does not force re-verification of an
  already-satisfied/pre-installed distribution, and it provides no isolation guarantee by itself.
- The command has no `--only-binary=:all:` (that flag was used only at **resolve** time when
  generating the lock, not at **install** time). If a listed wheel becomes unusable on the runner
  for any reason, pip can fall back to an sdist for the same pinned version, which triggers PEP
  517 build isolation and fetches the build backend (setuptools/etc.) and *its* dependencies
  **without hash verification** — a live, unhashed code-execution path inside the exact job this
  shipment exists to lock down.
- Neither pip itself, nor its own bundled setuptools/wheel, nor the Python patch version from
  `actions/setup-python` are pinned or hash-verified anywhere — the verification mechanism is
  itself an unpinned component (see M-2, related but scored separately since it is the installer
  rather than the payload).

**Fix**: add `--only-binary=:all:` (and consider `--no-deps` given the lock claims to be a
complete closure) to the install invocation; document that this flag is mandatory at install
time, not merely at lock-generation time.

### M-2. The installer itself (pip/setuptools/wheel/Python patch) is outside the pinned boundary
**Severity: MAJOR | Confidence: MEDIUM/Majority (3/4: Anchor, Reviewer-B, Reviewer-C) | Priority: 2×3 = 6**

`--require-hashes` constrains only the packages named in the requirements file. The pip binary
doing the installing — and any setuptools/wheel it bundles — comes from whatever
`actions/setup-python@...` installs for `'3.12'` that week, unpinned and unverified. The AC this
shipment claims is "hermetic install"; what actually ships is "hermetic **payload**, ambient
**installer**." This is a smaller surface than M-1 but the same class of gap, and should at
minimum be disclosed rather than left implicit given the shipment's stated purpose.

**Fix**: either pin/hash-verify a minimal pip bootstrap, or add an explicit comment in the lock
file disclosing that the installer itself is outside the pinned boundary so the guarantee is not
overstated (matches the plan's own AC-6 spirit: disclose residual non-hermeticity rather than
imply closure that wasn't achieved).

### M-3. `.gitignore` line comparison has no CRLF/whitespace normalization
**Severity: MINOR | Confidence: MEDIUM/Majority (3/4: Reviewer-A, Reviewer-B, Reviewer-C) | Priority: 2×2 = 4**

`subsequence_check` does raw string equality on `git show` blob lines. A pure CRLF introduction
(e.g. a Windows contributor, or a `.gitattributes` normalization commit) makes every affected line
fail to match anywhere in the new file's line set, and the gate reports a false I6 **violation**
on a change that removed nothing. This fails in the safe direction (blocks rather than admits a
real deletion) but is untested — none of the three shipped fixtures exercise it — and is a
blocking false positive on a job now wired into `ci-gate`.

**Fix**: strip trailing `\r` (and optionally trailing whitespace) in `non_comment_lines` before
comparison; add a CRLF fixture.

---

## 3. Plurality findings (confidence: MEDIUM — flagged by 2 of 4 reviewers, not a strict majority)

### P-1. Lock file's claimed "complete 7-package closure" may be missing `typing-extensions`
**Severity: CRITICAL | Confidence: MEDIUM/Plurality (2/4: Anchor, Reviewer-C) | Priority: 2×4 = 8**

Both reviewers independently made the same specific, falsifiable technical claim: `referencing`
(and transitively `jsonschema`) commonly declare `typing-extensions>=4.4.0; python_version<'3.13'`
in their metadata. The lock targets `python-version: 312` — the marker would be true — yet
`typing-extensions` is absent from the 7-package closure the file's header claims is complete and
verified. If this claim is accurate, `pip install --require-hashes` would refuse to install (a
requirement lacking a hash fails hashes-required mode), meaning the install as shipped may never
have been run to a clean, verified success on the actual target interpreter — which would call
into question U1 AC-5 ("the topology-check job still passes on a clean tree").

**This claim could not be independently verified in this review** (no network/package-index
access was available to either this orchestrating review or its reviewer subagents). Two
independent models converging on the identical, specific dependency-marker claim is stronger
signal than a typical single-source finding, but it is not confirmed. **This is a MUST-VERIFY
item, not an assumed-true one.**

**Required action before merge**: re-run the documented resolution procedure
(`pip download autoharness==1.4.11 --platform manylinux_2_17_x86_64 --python-version 312
--implementation cp --abi cp312 --only-binary=:all: -d <scratch>`) in a clean environment and
confirm whether `typing-extensions` (or any other package) appears in the resolved set. If it
does, add it with its hash and correct the "7 packages, verified as complete" claim. If it does
not, note in the PR body that this was explicitly re-verified and the closure is confirmed
complete as shipped.

### P-2. `.gitignore` checker hard-fails when the base ref has no version of the file (illegitimate rejection of the first-ever `.gitignore` PR)
**Severity: MAJOR | Confidence: MEDIUM/Plurality (2/4: Reviewer-B, Reviewer-C) | Priority: 2×3 = 6**

`git show "${base_ref}:${target_file}"` failing is treated identically whether the file
legitimately did not exist at base (the normal state for a PR introducing `.gitignore` for the
first time, or first introducing a nested one via `--file`) or the ref itself is corrupt/garbage.
An append-only invariant over an empty base is vacuously satisfied and should PASS; instead it
hard-FAILs with a generic "could not read" message, which is safe-direction but conflates two very
different failure modes and blocks a legitimate PR with no fixture covering this case.

**Fix**: distinguish "ref resolves, path absent" (treat old content as empty, PASS with a notice)
from "ref itself unresolvable" (hard fail) using `git cat-file -e` / `git rev-parse --verify`
before falling back to the current behavior. Add a fixture.

### P-3. Wheel/platform-tag lock fragility — forced resolution tag may already diverge from the runner's native pip resolution
**Severity: MAJOR | Confidence: MEDIUM/Plurality (2/4: Reviewer-B raised the general monitoring gap, Reviewer-C raised the acute mechanism) | Priority: 2×3 = 6**

The lock was generated by *forcing* `--platform manylinux_2_17_x86_64 --abi cp312` at resolution
time, with exactly one hash recorded per package. Runner-side `pip install` does **not** honor
that forced tag — it selects the highest-priority wheel it finds compatible with the actual
runtime, which may differ (e.g. a package publishing both `manylinux_2_17` and
`manylinux_2_28`/`musllinux` variants). If the runner's native pip resolution picks a different
file than the one whose hash was recorded, the install fails hard with a hash mismatch — a
present-day divergence risk, not merely a hypothetical future one, since the forced-platform
generation procedure was never validated against an actual run on `ubuntu-latest` +
`actions/setup-python 3.12`. The documented bump procedure reruns the same forced-tag commands,
which would reproduce rather than detect this class of divergence.

**Fix**: generate the lock by capturing what pip *actually selects* on the target runner image
(e.g. run the install once on `ubuntu-latest` and record the resolved file/hash), or record every
compatible distribution hash per package (as `pip-compile --generate-hashes` would), not a single
forced-platform hash. Add a coupling note tying the lock's target ABI to the workflow's
`python-version` value so a version bump without lock regeneration is caught.

### P-4. `check-gitignore-append-only.sh --self-test` is never invoked by any workflow
**Severity: MAJOR | Confidence: MEDIUM/Plurality (2/4: Reviewer-B, Reviewer-C) | Priority: 2×3 = 6**

U4a's entire purpose — fixtures + `--self-test` authored before the detection logic — provides
zero continuous protection in CI as wired: the `gitignore-append-only` job only ever runs the
script in real `--base-ref`/`--head-ref` mode against the actual PR diff. A future edit that
subtly breaks `subsequence_check` (or a tampered `expected` fixture) is undetectable until a PR
happens to exercise the exact broken edge case.

**Fix**: add a cheap step (fits naturally in the existing `lint` job, needs no network and no
extra checkout) running `bash scripts/check-gitignore-append-only.sh --self-test`, failing the
build on any fixture mismatch.

### P-5. gitleaks non-zero exit is unconditionally treated as "secret found"
**Severity: MAJOR | Confidence: MEDIUM/Plurality (2/4: Anchor rated MINOR, Reviewer-C rated MAJOR — most conservative severity taken) | Priority: 2×3 = 6**

`gitleaks detect ... || echo "found=true" >> "$GITHUB_OUTPUT"` cannot distinguish "leak detected"
from "gitleaks crashed / bad config / OOM / unreadable repo." Reviewer-C additionally identified a
compounding defect: every downstream step's `if: steps.scan.outputs.found == 'true'` **replaces**
the implicit `success()` condition GitHub Actions would otherwise apply, so if gitleaks errors out
*before* writing `gitleaks-report.json`, the summarize step still runs, `jq` fails on the missing
file, and the handoff-issue step still attempts to run too — producing a confusing cascade of
unrelated failures and, in the worst case, **no handoff issue at all** precisely when a real leak
would need one.

**Fix**: capture the exit code explicitly (`set +e; gitleaks ...; rc=$?; set -e`); branch on
`rc==0` (clean), gitleaks' documented "leaks found" exit code (`found=true`), and any other code
(distinct "scan operational failure" path, its own alert). Add `success() &&` to the downstream
`if:` conditions and guard `jq` with a file-existence/non-empty check.

### P-6. Weekly finding creates a duplicate GitHub issue every run with no dedupe
**Severity: MINOR | Confidence: MEDIUM/Plurality (2/4: Reviewer-B, Reviewer-C) | Priority: 2×2 = 4**

Git history is immutable without a rewrite (which this shipment correctly never does). A single
true positive therefore recurs on every subsequent weekly scan, and `gh issue create` runs
unconditionally — roughly 13 duplicate high-priority security issues per unrotated finding per
quarter, which trains operators toward alert fatigue on exactly the signal this control exists to
protect.

**Fix**: search for an existing open issue with a matching title/label before creating a new one
(`gh issue list --state open --search ...`), comment on it instead of duplicating. Consider a
`.gitleaksignore` keyed by fingerprint for triaged/rotated findings, documented in the workflow
comment.

### P-7. Unescaped `jq` interpolation into markdown tables/issue bodies; no size cap
**Severity: MINOR | Confidence: MEDIUM/Plurality (3/4 total across Anchor MINOR, Reviewer-B MINOR, Reviewer-C MINOR — treated here as majority, see note) | Priority: 2×2 = 4**

*(Note: this finding was actually flagged by 3/4 reviewers — Anchor, Reviewer-B, and Reviewer-C —
so it technically belongs in the Majority section above; listed here for narrative adjacency to
the other secret-scan findings. Its score/action-class treatment already reflects Majority, not
Plurality.)*

`RuleID`/`Fingerprint`/`Commit`/`File` are interpolated raw into a markdown table with no escaping
in both the job summary and the auto-filed issue body. A path containing a literal `|` breaks
table alignment; a path containing markdown/HTML link syntax renders as clickable content inside
an auto-filed security-triage issue — a low-grade content-injection vector on a surface an
operator is meant to trust. Reviewer-C additionally flagged that a first full-history scan with
many findings can exceed `$GITHUB_STEP_SUMMARY`'s ~1MB limit or the issue-body size limit, which
would silently drop content or fail the `gh issue create` call entirely — meaning the handoff
issue this AC depends on might never be created for a large first-run finding set.

One point of **explicit reassurance, confirmed by cross-checking Reviewer-C's independent note**:
`Fingerprint`'s actual gitleaks format is `commit:file:rule:startLine` and carries no secret
material by construction — the disclosure-control concern about `Fingerprint` leaking partial
secret content specifically (raised as a targeted focus area) does **not** appear to be
substantiated by any reviewer. See Unique finding U-8 below for the remaining open question about
the on-disk report file itself.

**Fix**: escape `|` (and control/markdown-significant characters) in the `jq` projection for both
consumers; cap the rendered table (e.g. `.[0:100]` with a "N more omitted" footer) and always
retain the full count separately.

---

## 4. Unique findings (confidence: LOW — flagged by exactly 1 of 4 reviewers)

Per protocol, LOW-confidence + CRITICAL findings are preserved and flagged for action despite
single-source status; the two below are exactly that case and should not be dismissed as noise.

### U-1. **`.gitignore` append-only checker can be defeated by a single appended negation line** — CRITICAL
**Confidence: LOW/Unique (Reviewer-C only) | Priority: 1×4 = 4 (elevated in practice — see note)**

`subsequence_check` only proves that every old non-comment **line's text** still appears, in
order, in the new file. It does not — and structurally cannot, as specified — prove the *ignore
set* is non-shrinking. Appending a negation pattern (e.g. `!.env`, `!*.pem`, `!/secrets/`) is a
pure textual append: every old line survives verbatim, order is preserved, the subsequence holds,
and the gate **PASSES** — while the appended line un-ignores exactly the class of file the I6
invariant exists to keep out of the repository. This is *easier* to execute than the deletion the
gate is built to catch, and it directly defeats the entire purpose of U4/U5 with a single
one-line diff that the checker is specified to treat as a pure insertion.

This is a single-source finding, but it is a concrete, reproducible logic gap in the precise
mechanical predicate the plan defines ("old file's lines are an ordered subsequence of new
file's lines") — the predicate itself is insufficient for the invariant it's named after, not an
implementation bug in an otherwise-correct predicate. Per the protocol's explicit carve-out for
LOW-confidence + CRITICAL findings, this is escalated to the remediation queue as `gated_auto`
despite single-source status.

**Fix**: reject (or require an explicit, separately-reviewed override for) any newly added line
whose first non-whitespace character is `!`, at minimum as an interim mitigation. Stronger:
compare *effective ignore behavior* rather than raw text — e.g. run `git check-ignore` (or
`git ls-files --ignored --exclude-standard`) over a fixed probe-path set at base and head, and
fail if any path ignored at base is no longer ignored at head. Add a `negation/` fixture with
`expected=fail`.

### U-2. `topology-check`'s job-level `continue-on-error` masks *all* step failures, not just the topology verdict — MAJOR
**Confidence: LOW/Unique (Reviewer-C only) | Priority: 1×3 = 3 (elevated in practice — see note)**

`continue-on-error: ${{ vars.PIPELINE_TOPOLOGY_GATE_REQUIRED != 'true' }}` is set at **job**
scope. With the flag unset (the plan's explicitly-required current default — Non-Goals: "stays
at its current (non-required) setting"), this swallows failure from *every* step in the job,
including the hash-pinned install step this entire shipment's U1 is about. If the lock file is
wrong, a hash mismatches, PyPI is briefly unavailable, or `ci-topology-check.sh` itself crashes,
the job still reports `success` to `ci-gate` today, and would continue to do so indefinitely with
zero visible CI signal. The gate correctly passing on a clean tree at implementation time (U1
AC-5) says nothing about whether a *future* regression in the lock file would ever be caught —
the masking is total, not scoped to the topology-verdict step.

**Fix**: scope `continue-on-error` to only the step that reports the topology verdict (step-level,
not job-level), so setup/install failures fail the job unconditionally regardless of the
topology-required toggle. Alternatively, split the hash-pinned install into its own
always-required job separate from the advisory topology verdict.

### U-3. `mapfile` reads from `non_comment_lines`'s process substitution without checking `grep`'s exit status
**Severity: MAJOR | Confidence: LOW/Unique (Anchor only) | Priority: 1×3 = 3**

`mapfile -t old_lines < <(non_comment_lines "$old_file")` — if `grep` (or a future change to
`non_comment_lines`) fails inside the process substitution, `set -e` does not propagate that
failure to the parent shell in the standard bash process-substitution case, and `mapfile` itself
can still "succeed" against an empty or truncated stream, silently producing an empty
`old_lines`/`new_lines` array. An empty `old_lines` is (correctly, per the script's own logic)
*trivially a subsequence of anything* — meaning a read failure on the OLD file's lines could
silently manifest as a false PASS in a script whose entire mandate is fail-closed behavior.

**Fix**: capture `non_comment_lines`'s exit status explicitly rather than relying on `mapfile`
inside `set -e` to propagate it (e.g. write to a temp file first and check `$?`, or use `grep`'s
own explicit call with status capture instead of a bare process substitution).

### U-4. `github.event.pull_request.base.sha` is the PR's recorded base tip, not the actual merge-base
**Severity: MAJOR | Confidence: LOW/Unique (Reviewer-C only) | Priority: 1×3 = 3**

The job passes `github.event.pull_request.base.sha` as `--base-ref`. That is the base branch's tip
at PR-open/last-sync time, not `git merge-base base head`. If unrelated commits land on `main`
between the PR's recorded base and the current tip (e.g. another merged PR that legitimately
edited `.gitignore` in a way this checker would flag), this PR's `gitignore-append-only` job can
fail and blame a change the PR under review never made. This also means
`--allow-no-merge-base` — whose entire framing in the script's own help text and error messages is
about an unresolvable *merge base* — is effectively dead code in every CI-reachable path, since CI
never actually computes or needs a merge-base; it always has a `base.sha` to hand the script
directly.

**Fix**: compute the actual merge-base in the workflow (`git merge-base "$BASE_SHA" "$HEAD_SHA"`)
and pass that as `--base-ref`, failing closed if `git merge-base` itself cannot resolve one — at
which point `--allow-no-merge-base` becomes a meaningful, CI-reachable flag rather than a decoy.

### U-5. Multi-commit "remove-then-restore" blind spot within a single PR
**Severity: MAJOR | Confidence: LOW/Unique (Reviewer-B only) | Priority: 1×3 = 3**

`subsequence_check` only ever compares two snapshots — PR base vs current head. It never inspects
intermediate commits. A PR author (or a compromised/malicious commit) could remove a `.gitignore`
entry in commit N (e.g. un-ignoring a secrets-shaped path), commit a file into that now-untracked
slot, then restore the same entry in commit N+1. By the time base-vs-head is diffed, the entry is
present again in relative order and the check PASSES, even though the file was tracked in the
interim — and because this repository preserves merge-commit history (Constitution Principle XI,
explicitly protected by this very shipment), that intermediate commit remains permanently
reachable in `main`'s history after merge. The "append-only" name implies protection against
exactly this maneuver; the two-snapshot comparison does not provide it. Reviewer-B's own
suggested fix correctly notes this is meaningfully mitigated (not eliminated) by the weekly
full-history secret scan (U2) as a compensating control, since the resulting file would exist in
history in a commit the scheduled scan would eventually see and alert on — though a targeted
non-secret file (e.g. a bypass config, a debug flag) wouldn't be caught by gitleaks at all.

**Fix**: either diff every commit in the PR range against its immediate parent (not just
base..head) and fail on any single commit that removes a line later reintroduced, or explicitly
document this as a known, accepted gap partially compensated by U2, and cross-reference U2 in the
job's success messaging so operators do not over-trust the PASS signal as a complete guarantee.

### U-6. I6 invariant is enforced only for the root `.gitignore`
**Severity: MINOR | Confidence: LOW/Unique (Reviewer-C only) | Priority: 1×2 = 2**

`target_file` defaults to `.gitignore` and no CI caller ever passes `--file`. Git honors
`.gitignore` files in any directory. A PR could leave the root file untouched (PASS) while
editing or adding a nested `subdir/.gitignore` to un-ignore content there; the gate's name and its
`ci-gate` wiring imply repo-wide coverage that does not exist.

**Fix**: enumerate all `*.gitignore` paths present at base or head
(`git ls-tree -r --name-only` filtered) and run the check per-path, or explicitly document the
root-only scope in the job name and script usage text so the coverage claim isn't overstated.

### U-7. AC-0's protection pre-flight is a point-in-time snapshot with no re-verification mechanism
**Severity: MINOR | Confidence: LOW/Unique (Reviewer-C only) | Priority: 1×2 = 2**

The pre-flight (`branches/main/protection` → 404, `rulesets` → `[]`) was correctly run before
authoring, satisfying AC-0 as literally specified. But it is a snapshot at one point in time; the
CODEOWNERS header states the "no enforcement effect" fact durably. Branch protection or a ruleset
enabling `require_code_owner_reviews` could be turned on at any later moment (by the operator, by
a future shipment, or by a GitHub product default change), at which point this file silently
becomes an enforcing, single-owner, self-review-bypassable control — with nothing in this
shipment designed to notice that transition.

**Fix**: not a blocker for this shipment (AC-0 was satisfied as literally specified), but worth a
forward-looking note: a cheap periodic check re-querying `branches/main/protection` and
`rulesets?includes_parents=true`, failing loudly if `require_code_owner_reviews` becomes true
while the header still claims advisory-only, would close the staleness window. Reasonable to defer
to a future shipment given this is a monitoring gap, not a present defect.

### U-8. Whether `--redact=100` covers the on-disk `gitleaks-report.json`, not just console output, is unconfirmed
**Severity: MAJOR (as raised) | Confidence: LOW/Unique-with-rebuttal (Anchor raised the concern; Reviewer-C's independent pass explicitly asserted the opposite — that `--redact` does cover the on-disk report — without citing a source) | Priority: not numerically scored due to reviewer disagreement; treated as a verification item**

This is the one place the four reviewers **actively disagreed** rather than simply not all
noticing the same thing. Anchor flagged that the raw JSON report, which contains fields beyond
the four the plan restricts (potentially `Secret`/`Match`, commit author/email/message, entropy),
is written to disk (never uploaded, but present in the runner's filesystem for the job's
duration) under `-v` and could pose a leak risk via any later diagnostic/debug step. Reviewer-C's
independent pass asserted, in an aside, that gitleaks' `--redact=100` redacts `Secret`/`Match`
fields **in the on-disk report as well**, not merely in stdout — which if true would mean the
on-disk file is already safe by construction. Neither claim was verified against gitleaks 8.30.1's
actual source/documentation in this review (no network access was available to the reviewers or
this orchestrating pass).

**This directly answers one of the review's explicitly-requested focus areas and must be resolved
by the operator, not assumed either way.**

**Required action before merge**: confirm gitleaks 8.30.1's actual `--redact` semantics against
its documentation/source (does `--redact=N` redact matched values in `--report-format json`
output, or only in `-v`/console output?). If report-file redaction is NOT guaranteed:
(a) drop `-v` (it provides no required output and increases the surface of what's written), and
(b) immediately post-process `gitleaks-report.json` into a sanitized four-field file and delete
the raw report before the job's remaining steps run, rather than relying on the raw file never
being uploaded as the sole control. If report-file redaction IS confirmed, document the citation
in the workflow's comment block so this isn't re-litigated on the next audit pass.

---

## 5. Scope-creep check (focus area 7) — CLEAR

Independently verified by direct file inspection (not solely reviewer-reported): no Go source
files are touched; `PIPELINE_TOPOLOGY_GATE_REQUIRED`'s conditional
(`vars.PIPELINE_TOPOLOGY_GATE_REQUIRED != 'true'`) is unchanged from its pre-existing form; no
branch-protection or ruleset API call that *mutates* state appears anywhere in the diff (only the
read-only `gh api .../protection` and `gh api .../rulesets` pre-flight calls the plan's AC-0
requires); `check-retired-architecture.sh` is invoked identically to before
(`bash scripts/check-retired-architecture.sh`, no `--self-test` flag) — U6 was correctly **not**
wired in, consistent with its move to 008-F. No reviewer flagged a Non-Goals violation either. No
finding in this dimension.

---

## 6. Remediation queue (ordered by priority = confidence_weight × severity_weight)

| # | Finding | Conf. | Sev. | Score | Action class |
|---|---|---|---|---|---|
| 1 | C-1: CODEOWNERS missing lock file + fixtures | HIGH | CRITICAL | 12 | `safe_auto` |
| 2 | M-1: `--require-hashes` not fully hermetic (no `--only-binary`) | MEDIUM | CRITICAL | 8 | `gated_auto` |
| 3 | P-1: lock closure may be missing `typing-extensions` | MEDIUM | CRITICAL | 8 | `manual` (requires re-resolution) |
| 4 | M-2: installer itself unpinned | MEDIUM | MAJOR | 6 | `gated_auto` |
| 5 | P-2: hard-fail on legitimate missing base `.gitignore` | MEDIUM | MAJOR | 6 | `gated_auto` |
| 6 | P-3: wheel/platform-tag lock fragility | MEDIUM | MAJOR | 6 | `manual` |
| 7 | P-4: `--self-test` never wired into CI | MEDIUM | MAJOR | 6 | `safe_auto` |
| 8 | P-5: gitleaks exit-code conflation + missing `success()` guards | MEDIUM | MAJOR | 6 | `gated_auto` |
| 9 | M-3: CRLF/whitespace not normalized | MEDIUM | MINOR | 4 | `safe_auto` |
| 10 | P-6: duplicate handoff issues, no dedupe | MEDIUM | MINOR | 4 | `gated_auto` |
| 11 | P-7: unescaped jq/markdown injection + size limits | MEDIUM (majority) | MINOR | 4 | `safe_auto` |
| 12 | U-1: negation-line bypass of I6 checker | LOW | CRITICAL | 4* | `gated_auto` (escalated per protocol) |
| 13 | U-2: job-level `continue-on-error` masks all step failures | LOW | MAJOR | 3* | `gated_auto` (escalated — undermines U1 signal) |
| 14 | U-3: `mapfile`/process-substitution error swallowing | LOW | MAJOR | 3 | `gated_auto` |
| 15 | U-4: `base.sha` ≠ actual merge-base | LOW | MAJOR | 3 | `gated_auto` |
| 16 | U-5: multi-commit remove-then-restore blind spot | LOW | MAJOR | 3 | `advisory` (compensated by U2 scan) |
| 17 | U-8: `--redact=100` on-disk-report coverage unconfirmed | LOW (disputed) | MAJOR | n/a | `manual` (verify then fix) |
| 18 | U-6: I6 enforced only for root `.gitignore` | LOW | MINOR | 2 | `advisory` |
| 19 | U-7: AC-0 pre-flight is a point-in-time snapshot | LOW | MINOR | 2 | `advisory` (defer to future shipment) |
| 20 | U-3 (shell hygiene: undeclared locals, dead branch, trap timing) | LOW | MINOR | 2 | `advisory` |

\* Priority scores for #12 and #13 are shown per the formula, but both are called out above their
numeric rank per the protocol's explicit instruction that LOW-confidence + CRITICAL (and, by
extension here, a LOW-confidence MAJOR that structurally invalidates another finding's safety
assumption) warrants flagging "despite single source." Recommend treating both as required fixes
in this same remediation pass, not deferred.

**Recommended blocking set for this PR** (minimum to flip verdict to PASS): #1, #2, #3
(verification, not necessarily a code change if re-resolution confirms completeness), #12. These
four are the only findings that either falsify the plan's own stated acceptance criteria (#1, #3)
or leave a demonstrated, reproducible bypass of the shipment's core deliverable (#12), or
materially weaken the central hermeticity claim this shipment exists to make (#2). Everything else
in the queue is legitimate and should be fixed or explicitly deferred-with-rationale in the same
spirit as the plan's own AC-6 disclosure convention, but does not need to block this specific PR.

---

## 7. Backlog work items (P0/P1 findings)

```yaml
type: bug
title: "CODEOWNERS: cover .github/constraints/ and scripts/testdata/gitignore/ (U3 AC-1 gap)"
description: "CODEOWNERS claims to cover ALL CI-critical paths 'or none' but omits the hash-pinned lock file (.github/constraints/autoharness-lock.txt) and the gitignore-checker fixtures (scripts/testdata/gitignore/**), both of which are CI-critical trust-boundary artifacts."
file: ".github/CODEOWNERS"
line: null
severity: "CRITICAL"
confidence: "HIGH"
fix: "Add '/.github/constraints/ @softwaresalt' and '/scripts/testdata/ @softwaresalt' (or the gitignore-specific subpath) entries."
linked_review: "docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md"
---
type: bug
title: "pip install --require-hashes lacks --only-binary=:all: at install time, permitting an unhashed sdist/build-isolation fallback path"
description: "The install command only forced --only-binary=:all: at lock-generation time, not at install time. A wheel-selection failure can trigger an sdist build via PEP 517 build isolation, fetching unhashed build-backend dependencies over the network inside the job this shipment exists to lock down."
file: ".github/workflows/ci.yml"
line: null
severity: "CRITICAL"
confidence: "MEDIUM"
fix: "Add --only-binary=:all: (and consider --no-deps) to the 'Install autoharness' step's pip invocation."
linked_review: "docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md"
---
type: bug
title: "Verify autoharness-lock.txt's claimed 7-package closure is actually complete for cp312 (possible missing typing-extensions)"
description: "Two independent reviewers flagged that referencing/jsonschema commonly require typing-extensions on python_version<3.13, which is the lock's target marker condition, yet typing-extensions is absent from the lock. Unverified in this review (no network access). If the install has never actually succeeded end-to-end on the target interpreter, U1 AC-5 ('passes on a clean tree') may be unverified in practice."
file: ".github/constraints/autoharness-lock.txt"
line: null
severity: "CRITICAL"
confidence: "MEDIUM"
fix: "Re-run the documented pip download/pip hash procedure in a clean cp312 environment; add any missing package+hash, or explicitly confirm and document that the closure is complete as shipped."
linked_review: "docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md"
---
type: bug
title: ".gitignore append-only checker can be bypassed by appending a negation pattern (!path)"
description: "subsequence_check only validates that old lines survive as an ordered subsequence of new lines. Appending a negation line (e.g. !.env) is a pure textual insertion that passes the check while un-ignoring previously protected paths, defeating the I6 invariant this unit exists to enforce."
file: "scripts/check-gitignore-append-only.sh"
line: null
severity: "CRITICAL"
confidence: "LOW"
fix: "Reject newly-added lines beginning with '!' at minimum as an interim mitigation; stronger fix compares effective git-ignore behavior (git check-ignore / git ls-files --ignored) between base and head rather than raw text. Add a negation/ fixture with expected=fail."
linked_review: "docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md"
---
type: bug
title: "topology-check job-level continue-on-error masks all step failures, not just the topology verdict"
description: "continue-on-error is set at job scope keyed off PIPELINE_TOPOLOGY_GATE_REQUIRED, which is currently unset per plan Non-Goals. This means any failure in the hash-pinned install step (bad hash, missing package, index outage) is silently swallowed and reported success to ci-gate today, providing no real regression signal for U1's core deliverable."
file: ".github/workflows/ci.yml"
line: null
severity: "MAJOR"
confidence: "LOW"
fix: "Scope continue-on-error to only the step reporting the topology verdict, or split the hash-pinned install into its own always-required job separate from the advisory topology-verdict step."
linked_review: "docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md"
```

---

## 8. Post-remediation re-review

Not run in this invocation — this was a **report-only** pass with no auto-fixes applied (per the
explicit instruction to not modify any files). If the operator applies the `safe_auto`/`gated_auto`
items above, a Phase 7 post-remediation re-review cycle is recommended before merge, scoped to the
modified files only, with the same 4-reviewer panel.

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 20
  status: "skipped"
```

---

## 9. Summary for the operator

This is a strong, deliberate piece of security engineering — the first pass's fixes (hash-pinning,
fail-closed base resolution, U3's honest ownership-only scoping, U6's cross-shipment reordering)
all verified correctly in the actual code, with no regressions found by any of the 4 reviewers.
The issues this second pass surfaced are almost entirely of the shape "the stated guarantee has a
narrow, fixable gap" rather than "the design is wrong" — consistent with a shipment that has
already been through one hardening cycle. The one genuinely alarming find (U-1, the negation-line
bypass) is a single-reviewer catch, but it is concrete, reproducible, and cheap to fix, and
squarely fits the class of thing a second independent pass exists to catch. Recommend: fix the
four blocking items, re-run `--self-test` plus the negation fixture, and ship.
