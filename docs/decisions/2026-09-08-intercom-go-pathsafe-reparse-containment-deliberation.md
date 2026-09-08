---
title: "Deliberation — Pathsafe Final-Component Reparse-Point Containment"
date: 2026-09-08
status: accepted
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 014-F / shipment 013-S
---

# Deliberation — Pathsafe Final-Component Reparse-Point Containment

- **Date**: 2026-09-08
- **Agent**: Stage (P-017 dark-factory cycle, post-012-S)
- **Mode**: `DARK_MODE_ACTIVE`, operator AFK, hard limit of exactly **ONE** shipment
- **Scope under consideration**: all 22 active stash entries (full dark scope), of which
  **4 are selected** for the single shipment: `700B41CE`, `F133AB7E`, `2362BBB5`, `AD0D9D1F`
- **Governing predecessors**:
  - `docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md` (findings F0–F11)
  - `docs/closure/011-S-012-F-post-merge-closure.md` (deferred-follow-up ledger)
  - `docs/decisions/2026-09-07-intercom-go-pathsafe-containment-correctness-deliberation.md`
  - `internal/pathsafe/root.go` package risk register (GO-14, 700B41CE, SEC-5, 5FE4A7BE)
  - `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`

---

## 1. Framing — what problem are we actually solving?

`internal/pathsafe` is the workspace path-containment control. Brief section 10 marks
workspace isolation **NON-NEGOTIABLE**. The control's single enforcement point is
`checkSymlinkEscape` in `internal/pathsafe/pathsafe.go`.

That function currently probes the **final** path component with `os.Stat` and every
**strict ancestor** with `os.Lstat`. That one asymmetry — chosen deliberately by
`012.003-T` to preserve the GO-14 oracle-parity acceptance "by construction" — is the
root cause of **both** open containment defects:

| Defect | Final-component input | Why it is accepted today |
|---|---|---|
| `700B41CE` (**critical**) | live Windows directory junction → outside root | `os.Stat` transparently follows `IO_REPARSE_TAG_MOUNT_POINT`, so the loop breaks on iteration 1; `filepath.EvalSymlinks` never resolves mount points; `hasPathPrefix` then compares the **lexically unchanged** in-root path and passes. |
| `F133AB7E` (GO-14) | dangling symlink → outside root | `os.Stat` fails `ENOENT`, the walk ascends to the (in-root) parent, `EvalSymlinks(parent)` succeeds and is contained → accept. The symlink itself is never examined. |

Both yield the **same primitive**: a caller that writes through the returned path writes
outside the workspace root. `700B41CE` additionally yields a **read** primitive and was
**empirically reproduced on this host** (an outside-root secret file was read through an
accepted path). It requires **no elevated privilege** — junction creation, unlike
`os.Symlink`, needs neither `SeCreateSymbolicLinkPrivilege` nor Developer Mode.

**This is the central insight of this deliberation:** these are not two independent bugs
that happen to share a file. They are one defect class — *"the final path component's
reparse-point identity is never examined"* — with two triggers. Any correct fix touches
the same three lines of probe logic. Fixing one without the other is not possible without
leaving the seam half-corrected.

---

## 2. Stash triage — classification of all 22 in-scope entries

### 2.1 Shape and precedence classification

Per the Step 1 precedence rule, entries carrying the literal `DEFERRED SCOPE EXPANSION`
marker are forced onto the `deliberate` route regardless of shape or size. **14 entries**
carry it. `F133AB7E` and `2362BBB5` carry `DEFERRED REVIEW FINDING(S)` and both
self-declare `Requires deliberation: yes`; they are treated identically.

### 2.2 Contextual grouping analysis

Four contextually coherent clusters emerged from code-surface analysis:

**Cluster A — `internal/pathsafe` containment security** (4 entries)
`700B41CE` (critical bug), `F133AB7E` (medium bug), `2362BBB5` (low bug), `AD0D9D1F` (medium task).
All four modify `checkSymlinkEscape` and/or the `root.go` risk register. Single function,
single package, single review lineage (the 011-S adversarial review).
Estimated scope: 8 tasks × 2h = 16h. Risk: **high blast radius, narrow surface**.

**Cluster B — `scripts/check-retired-architecture.sh` detection quality** (9 entries)
`24B63533`, `8988120D`, `6285779C`, `6C24E2E4`, `8D603E29`, `9A14D3B7`, `E403E3C5`,
`09CD6ACF`, `F4F4A959`. All CI-gate detection/robustness work on one shell script.
Highly coherent, but **zero critical**, and the gate is an explicitly **anti-accident,
not anti-adversary** control (D6b). Estimated scope: ~10 tasks = 20h.

**Cluster C — documentation and script hygiene** (5 entries)
`F47DB9A9`, `2787DA56`, `4C5BEC23`, `1C6C3B46`, `9D45E62E`. All low. Documentation
ranks last in operator priority.

**Cluster D — standing trackers and operator-blocked** (4 entries)
`A92E3FA0`, `4989A42D` (trackers), `EF9352FB` (operator-only credential rotation),
`BF5DE670` (deferred mitigation, trigger unfired).

### 2.3 Selection — Decision D-1

**SELECTED: Cluster A.**

Rationale against the operator's stated priority order (reliability and security supersede
feature work; then composability/simplifying refactors; then features; then documentation;
within that, order by priority then type bug → review → feature → task → spike):

1. Cluster A contains the **only `critical` entry in the entire dark scope** (`700B41CE`),
   and it is an **empirically reproduced, unprivileged, live containment bypass** in a
   control the Brief marks NON-NEGOTIABLE. Operator rule 1 makes this unambiguous.
2. The operator's brief explicitly directs that `700B41CE` "should normally lead unless
   evidence shows it is unsafe or blocked." Evidence gathered this cycle shows it is
   **neither**: the mechanism is fully traced, the remediation fork is documented, the
   Windows test harness (`createDirectoryJunction`) already exists, and pre-existence in
   `main` is confirmed. It leads.
3. Cluster B is genuinely coherent and larger, but every entry is medium/low and the gate
   it hardens is a hygiene control, not a security boundary. It cannot outrank a live
   containment bypass.
4. Clusters C and D are documentation or non-actionable.

Cluster A is also the **smallest coherent unit that fully contains the critical defect** —
it does not drag unrelated surfaces into the blast radius of a NON-NEGOTIABLE control.

---

## 3. Membership decisions within Cluster A

### D-2 — `F133AB7E` (GO-14) is **mandatorily coupled**, not optional scope creep

Three independent grounds:

1. **Technical coupling.** The fix for `700B41CE` must examine the final component with
   `os.Lstat` before trusting `os.Stat`. That *necessarily* changes the dangling-final-symlink
   verdict, because a dangling symlink is exactly "an `Lstat` entry that exists and is a
   reparse point." The GO-14 acceptance is preserved today only *by construction* of the
   probe ordering the fix must change. There is no fix for `700B41CE` that leaves GO-14 untouched.
2. **In-code procedure.** `root.go`'s register states the flip procedure verbatim: *"A future
   shipment may flip the GO-14 regression lock only if it explicitly reconsiders this
   final-component acceptance, updates the lock, and lands the replacement boundary in the
   SAME change (see stash F133AB7E)."* This shipment is that change.
3. **Not scope expansion.** `F133AB7E` is an explicitly in-scope stash entry in this dark
   run whose entire content is "reconsider the GO-14 risk acceptance." Acting on it is
   executing the declared scope, not expanding it. The prior cycle's deferral reason
   ("reversing another shipment's regression-locked decision without operator input would
   be scope expansion") **no longer applies**: the operator has now placed `F133AB7E` in
   scope and directed security-first ordering.

**Consequence, recorded explicitly:** `TestResolveAllowsDanglingSymlinkAtFinalComponent`
(`internal/pathsafe/symlink_test.go`) is a live passing regression lock asserting
`Resolve("dangling-link")` returns `nil`. It must be **inverted and renamed in the same
commit as the behavior change**, never deleted. Deleting it would silently retire the
only evidence that the boundary moved deliberately.

**Oracle-divergence justification (required by D-2 ground 3):** GO-14 is framed as
oracle-parity with `agent-intercom @ 41df772 src/diff/path_safety.rs`. Closing it is a
deliberate divergence from the behavioral oracle. That divergence is justified because
governing decision **D8 demoted the Rust repository from behavioral oracle to historical
reference only** — parity with it is no longer a definition of done. A retained
arbitrary-write primitive cannot be justified by parity with a reference that no longer
governs. `700B41CE` carries no parity constraint at all (Go/Windows-runtime-specific).

### D-3 — Remediation option for `700B41CE`: **option (a), reparse-aware resolution**

The adversarial review named two options:

- **(a)** Detect a live name-surrogate reparse point at the final component and resolve its
  true target (`GetFinalPathNameByHandle`-equivalent semantics), then re-assert containment.
- **(b)** Pin `GODEBUG=winsymlink=0` module-wide via the `go.mod` `godebug` directive.

**Selected: (a).** Option (b) is rejected as the primary fix because:
- `700B41CE` itself records it as "an unverified, cross-cutting toolchain-behavior decision
  requiring its own deliberation, not a mechanical fix."
- It makes a NON-NEGOTIABLE security control depend on a **process-wide toolchain flag**.
  Any host, test harness, or downstream consumer overriding `GODEBUG` silently reopens the
  bypass. A containment control must be correct by its own construction.
- Its blast radius is every `os`/`filepath` call in the module, not just pathsafe.

**D-3a — GODEBUG independence is an acceptance criterion, not a separate fix.** The chosen
implementation must produce the **same verdict under both `winsymlink=1` (the pinned Go
1.26.5 default) and `winsymlink=0`**. Finding F7 correctly observes that today's verdict is
an undocumented-default-dependency. Verification is by a subprocess re-exec regression test
(the child sets `GODEBUG` in its environment; `//go:debug` directives cannot vary per-test).
If that proves infeasible within the task's 2-hour budget, the fallback is an explicit
documented residual plus a new stash defer — **not** a silent omission.

### D-4 — **REVISED after adversarial review**: do NOT relax the ancestor branch (F3 deferred)

**Original position (rev 1, now withdrawn):** also fix the F3 false-rejection, where a live
junction ancestor whose target is *inside* the root is unconditionally rejected because
`os.Lstat` reports `ModeIrregular` (neither `IsDir()` nor `ModeSymlink`).

**Withdrawn on two independent review findings:**

1. **Correctness P0.** Relaxing the ancestor guard while `filepath.EvalSymlinks` still cannot
   resolve mount points would compare an **unresolved, lexically in-root** path against
   `root.path` — **introducing a new intermediate-junction bypass** inside the shipment meant
   to close one.
2. **Scope P2.** F3 is not a declared member of any of the four selected stash entries. It
   *relaxes* a control (turns a current reject into an accept) and is not required by the
   definition of done — a fail-**closed** usability bug is not a containment failure.

**Revised decision:** retain the strict-ancestor `ModeIrregular` rejection unchanged. The
shipment becomes **monotonically fail-closed** — it converts accepts into rejects and never
the reverse, which is the correct risk posture for a NON-NEGOTIABLE control. F3 is recorded
as a deferred residual in the risk register.

### D-4a — The real fix is whole-prefix canonicalization, not a final-component special case

Adversarial review established that a final-component-only fix leaves **two further accepting
inputs** open, because `os.Lstat` does not follow only the *final* component — it still
traverses every intermediate one:

- `Resolve("junction/existing.txt")` — junction intermediate, leaf **exists**
- `Resolve("junction/existing-dir/missing.txt")` — ascent stops **beyond** the junction

In both, the walk terminates without ever examining the junction, and
`filepath.EvalSymlinks` leaves the mount point unresolved.

**Corrected root cause:** the defect is not *"the final component's reparse identity is
unexamined"* but *"the containment check is performed against a path that was never
mount-point-canonicalized."* The remedy is therefore to replace
`filepath.EvalSymlinks(ancestor)` with mount-point-aware canonicalization of the whole
existing prefix. GO-14 is the same defect reached via the `ENOENT` ascent path, which
reinforces D-2's coupling argument.

### D-5 — `2362BBB5`(a) and `AD0D9D1F`(F6) are the **same finding**; dedupe to one task

Both describe `checkSymlinkEscape`'s `parent == ancestor` terminal branch returning
`(resolved, nil)` — a fail-open default in a containment control. They were captured by
two different intake paths (Stage review gate vs. adversarial review). This is an
**overlap between two multi-item bundles, not a duplicate entry**: neither entry is a
subset of the other, so neither may be archived without losing its non-overlapping members.
Resolution: harvest **exactly one task** covering the flip, and record the overlap in both
entries' dispositions so it is not implemented twice.

The flip is a **zero-behavior-change hardening** — two reviewers independently traced the
branch as genuinely unreachable. It also touches the `011.006-T` coverage exclusion, which
must be updated in the same task. **Anti-goal:** do not claim a reproduced exploit for this
branch; its two hypothesized triggers were never reproduced.

### D-6 — `2362BBB5`(b): distinct error message constant

Today an intermediate entry that exists but is not statable (in-root dangling link, ELOOP,
EACCES, unresolvable reparse point) is rejected with `symlinkEscapeMsg` = "symlink target
escapes workspace". The verdict is correct; the **message asserts an escape that did not
occur** and discards the OS cause. Once real write paths land (C4–C6) this destroys the
signal a genuine escape should produce. Included: add a distinct
"target cannot be verified" constant with cause wrapping, keeping `apperr.KindPathViolation`.

### D-7 — `AD0D9D1F` residual membership

Of the 7 bundled advisory findings, **F5 (share-less UNC) and F10 (pwsh skip-vs-fail) are
already CLOSED** in commit `902c05f` (Copilot review on PR #34) — verified this cycle
against the commit's diffstat. They must **not** be re-opened.

Included (all same-surface, all cheap while the files are already open):
- **F4** — POSIX ENOTDIR-vs-ENOENT branch coverage gap
- **F6** — fail-open terminal branch (merged with D-5, one task)
- **F8** — `uncPrefix` doc comment misattributes `\\?\` to `GetFinalPathNameByHandle`
- **F9** — register entry `5FE4A7BE` embeds two contradictory STATUS/TRIGGER pairs
- **F11** — no dedicated test for the EACCES/ELOOP (`!errors.Is(err, fs.ErrNotExist)`) branch

### D-8 — CI gating: **surface, do not self-authorize** (operator decision)

`.github/workflows/ci.yml:240` sets
`continue-on-error: ${{ vars.WINDOWS_GATE_REQUIRED != 'true' }}` on the Windows `go test`
step. `WINDOWS_GATE_REQUIRED` is **unset**, so a Windows-only test failure is **masked** and
the job still reports success to `ci-gate`.

**Therefore a new Windows-only junction regression test cannot block CI today.** Claiming
this containment fix is "verified by CI" would be false: it is build-tagged out on Ubuntu
and swallowed on Windows.

Flipping `WINDOWS_GATE_REQUIRED` is a **repository-configuration change requiring explicit
operator confirmation** (Constitution VII); precedent from `011.010-T` is that Stage/Ship
must **not** self-authorize it. Resolution:
1. Ship **must** record a local Windows `go test ./internal/pathsafe/...` run as primary
   evidence (precedent: 011-S and 012-S closures did exactly this).
2. The PR body **must** disclose that the Windows-only coverage is advisory-in-CI.
3. The flip is raised as an **explicit operator decision**, carried in this deliberation
   and surfaced in the Stage report. It is **not** a blocker for queuing 013-S.

**Anti-goal:** do not add `-race` to the Windows job — `windows-latest` has no cgo and
`go test -race` fails closed with `go: -race requires cgo`.

---

## 4. Options considered for the containment fix

| Option | Description | Verdict |
|---|---|---|
| **O1** | Lstat-first final probe; reject any final component that is a reparse point whose resolved target is outside root; resolve mount points via a Windows-specific handle-based call | **SELECTED** |
| O2 | Module-wide `GODEBUG=winsymlink=0` pin | Rejected (D-3) — process-wide flag dependency for a NON-NEGOTIABLE control |
| O3 | Reject **all** final-component reparse points unconditionally (no target resolution) | Rejected — fails closed on legitimate in-root junctions/symlinks; a workspace legitimately containing an in-root junction becomes unusable. Retained as the **fallback** if handle-based resolution proves infeasible on the pinned toolchain, because fail-closed beats fail-open for this control |
| O4 | Defer again | Rejected — the defect is critical, reproduced, unprivileged, and now explicitly in operator scope |

**O1 fallback rule (recorded so Ship is not blocked mid-build):** if reparse-target
resolution cannot be implemented safely within the task budget, Ship must fall back to
**O3 (fail closed)** and defer the in-root-junction usability case to a new stash entry.
Ship must **never** fall back to leaving the bypass open.

---

## 5. Dispositions for all 18 non-selected in-scope entries

| Stash ID | Pri | Disposition | Reason |
|---|---|---|---|
| `700B41CE` | critical | **SELECTED** → 014-F | Leads the shipment |
| `F133AB7E` | medium | **SELECTED** → 014-F | Mandatorily coupled (D-2) |
| `2362BBB5` | low | **SELECTED** → 014-F | Same function; (a) merged with F6 (D-5), (b) error attribution (D-6) |
| `AD0D9D1F` | medium | **SELECTED** → 014-F | F4/F6/F8/F9/F11 same surface; F5/F10 already closed (D-7) |
| `4989A42D` | medium | **ARCHIVE** | Completion condition **MET** — see D-9 below |
| `A92E3FA0` | high | **RETAIN** (deferred) | Standing roadmap tracker and the ONLY index of 5 unresolved **operator** decisions (Q3/H3, Q6/H4, H5, H6, Q7). All 5 re-verified still open. Not agent-executable; archiving destroys the index. C3 remains deferred behind security work per operator rule 1. |
| `EF9352FB` | high | **DEFER — HARD BLOCKER, operator-only** | See D-10 |
| `BF5DE670` | medium | **RETAIN** (deferred) | Its primary ask (consolidated register) was discharged by `011.004-T`. The residual **mitigation** (Lstat-before-write, O_EXCL, hardlink checks) has an explicit trigger — "before pathsafe gates any live destructive file-write path" — which has **NOT fired**: `scripts/check-write-path-precondition.sh` still finds zero write primitives under `internal/**`/`cmd/**`. Building the machinery now is speculative engineering against an API that does not exist. |
| `24B63533` | medium | **DEFER** (Cluster B) | Retired-arch gate detection false negatives. Coherent with 8 peers; anti-accident control, no critical. Next cycle. |
| `8988120D` | medium | **DEFER** (Cluster B) | Overlaps `24B63533`(1) on `split_identifier`; must be merged into one unit next cycle |
| `6285779C` | medium | **DEFER** (Cluster B) | `mask_go_non_code` characterization fixtures |
| `6C24E2E4` | low | **DEFER** (Cluster B) | Structural refactor: single scan-scope source |
| `8D603E29` | low | **DEFER** (Cluster B) | `cmd/` branch filter asymmetry |
| `9A14D3B7` | low | **DEFER** (Cluster B) | Dispatch fails open |
| `E403E3C5` | low | **DEFER** (Cluster B) | Self-test framework robustness |
| `09CD6ACF` | low | **DEFER** (Cluster B) | Residual detection blind-spot disclaimers |
| `F4F4A959` | low | **DEFER** (Cluster B) | Undocumented `continue-on-error` on the retired-arch gate step; same CI-toggle family as D-8, belongs with Cluster B |
| `F47DB9A9` | low | **DEFER** (Cluster C) | Dangling docs references in `ci.yml` |
| `2787DA56` | low | **DEFER** (Cluster C) | `check-unignore-regression.sh` ref-resolution hardening |
| `4C5BEC23` | low | **DEFER** (Cluster C) | `check-unignore-regression.sh` stale header comment |
| `1C6C3B46` | low | **DEFER** (Cluster C) | Missing YAML frontmatter on two 2026-09-07 artifacts |
| `9D45E62E` | low | **DEFER** (Cluster C) | 5 archived stash entries lack forward-reference notes |

### D-9 — `4989A42D` completion condition is MET → **ARCHIVE**

The tracker carries an explicit, concrete archival condition: archive when **(a)** 009-F has
closed **AND** **(b)** the U-E1a retired-architecture gate has been broadened to `internal/**`
and passes.

Both measured directly against merged `HEAD` this cycle:
- **(a) MET** — 009-F and `009.001-T`..`009.006-T` are archived (unchanged from prior cycles).
- **(b) NOW MET** — `scripts/check-retired-architecture.sh:425` passes
  `git ls-files -- config.toml.example cmd/** internal/**`; the `internal/config/` prefix test
  that encoded the D6a narrowing is **gone entirely** (zero occurrences remain in the file);
  the header comment at line 13 now documents the scan as `internal/**`. "Passes" is evidenced
  by shipment 012-S merging through CI via PR #36 (`9d8937a`) with post-merge closure merged
  via PR #37 (`625a7b7`).

This is the **first cycle in which condition (b) has ever been met** — three prior cycles
measured it UNMET and retained the tracker. Its stated purpose (manual contamination tracking
until the gate covers `internal/**`) is now discharged by the mechanical gate itself. Archiving
is the action the entry's own contract prescribes; retaining it further would be ignoring a
satisfied exit condition.

### D-10 — `EF9352FB`: hard blocker, deliberately isolated into **no** shipment

Re-affirmed this cycle **without reading, revealing, copying, or committing any secret
material** — only metadata-level, read-only checks were performed and no key value was ever
loaded, printed, or stored.

Repo-side containment re-verified **CLOSED** against merged `HEAD`:
`git log --all -- .env.local` returns no commits; no `.env*` path was ever added anywhere in
history; `git ls-files -- .env*` is empty; the file remains ignored via `.gitignore:36`
(`.env.*`).

There is therefore **no in-repo remediation left and no code change to harvest**. The sole
residual action — rotating a live Tavily credential in an external provider console —
requires authenticating to a third-party system, an authority Stage does not hold and must
not simulate. It is deliberately isolated into **no shipment** so it cannot stall the 4 safe
selected entries. **TIME-SENSITIVE:** the untracked plaintext `.env.local` remains on the
operator's disk (existence checked only; contents never opened), so exposure persists until
the operator rotates the key. **Reported as a hard blocker, not a silent omission.**

---

## 6. P-021 C5/C6 intake obligations

### 6.1 Duplicate detection (unconditional, run over every deferred entry)

All 22 active entries were scanned pairwise for same-expansion duplicates.
**Result: CLEAN — no duplicate entries found.** No entry was archived as a duplicate and
no merge was performed.

Two **partial overlaps** were identified and recorded rather than merged, because in each
case neither entry is a subset of the other:
- `2362BBB5`(a) ∩ `AD0D9D1F`(F6) — the fail-open terminal branch. Handled by D-5 (one task).
- `24B63533`(1) ∩ `8988120D`(1) — `split_identifier` decomposition weakness. Both deferred;
  flagged for merge into a single unit when Cluster B is harvested.

No entry carried a `DISCOVERY-STATUS: AMBIGUOUS` or `LOOKUP-UNAVAILABLE` token.

### 6.2 Late-identifier reconciliation (triggered by `N/A` source refs)

| Entry | Field recovered | Value | Residual-risk record (join key) |
|---|---|---|---|
| `700B41CE` | PR | **#34** | `docs/closure/011-S-012-F-post-merge-closure.md` (names 700B41CE in Deferred follow-ups) |
| `AD0D9D1F` | PR | **#34** | same record (names AD0D9D1F, and records F5/F10 closed in `902c05f`) |
| `F133AB7E` | PR | **#34** | same record (011-S shipment PR) |
| `2362BBB5` | PR | **#34** | same record (011-S shipment PR) |
| `EF9352FB` | — | idempotent **no-op** | PR #4 already recovered 2026-09-04; no identifier overwritten |

`review-thread` remains a **terminal, truthful `N/A`** for all four selected entries: they
are Stage-review-gate and adversarial-review findings that predate PR thread creation, and
the governing closure record documents the PR #34 Copilot review as 2 comments, both already
dispositioned. `task=N/A` likewise stands for `700B41CE` — no covering task existed for a
defect discovered during review. Reconciliation was performed under Stage's own stash
authority; **no Ship write was requested**, preserving the C5 capture-only carve-out and the
single-write capture invariant.

---

## 7. Decisions summary

| ID | Decision |
|---|---|
| D-1 | Select Cluster A (pathsafe containment); defer Clusters B and C |
| D-2 | `F133AB7E`/GO-14 is mandatorily coupled; invert (never delete) the regression lock in the same commit; oracle divergence justified by D8 |
| D-3 | Fix via option (a) reparse-aware resolution; reject module-wide GODEBUG pin |
| D-3a | GODEBUG-independence is an acceptance criterion, verified by subprocess re-exec |
| D-4 | **Revised:** do NOT relax the ancestor branch; F3 deferred. Shipment stays monotonically fail-closed |
| D-4a | The fix is whole-prefix mount-point-aware canonicalization, not a final-component special case |
| D-5 | `2362BBB5`(a) ≡ `AD0D9D1F`(F6): one task, overlap recorded, neither entry archived |
| D-6 | Add a distinct "cannot verify target" message constant with cause wrapping |
| D-7 | `AD0D9D1F` F4/F8/F9/F11 included; F5/F10 already closed, do not re-open |
| D-8 | `WINDOWS_GATE_REQUIRED` flip is an operator decision — surface, do not self-authorize; require local Windows evidence |
| D-9 | `4989A42D` completion condition MET → archive |
| D-10 | `EF9352FB` remains a hard, operator-only blocker in no shipment |
| D-11 | Staging artifacts reach `main` via branch → PR → merge, **never** direct commit (see §8) |

---

## 8. D-11 — staging-artifact git path (P-001/P-016)

Committing Stage staging artifacts directly to `main` is a recorded **P-001/P-016
violation** in this repository: `f6d5879` did exactly that and required a corrective revert
via PR #31, a preservation branch, and reapplication via PR #33.

This cycle therefore uses the sanctioned path: a `chore/stage-013-s` branch **in the existing
worktree** (no new worktree — P-016), commit, push, open a staging-artifact PR, merge after
CI. This is the *staging-artifact* PR, which is distinct from and must not be confused with
the *implementation* PR that Ship will open for 013-S.

---

## 9. Anti-goals for shipment 013-S

- Do **not** implement TOCTOU, O_EXCL, or hardlink mitigations (`BF5DE670` trigger unfired).
- Do **not** relax the strict-ancestor `ModeIrregular` rejection (F3 deferred — D-4).
- Do **not** fix the case-folding fail-open (`5FE4A7BE`) — not in any selected entry.
- Do **not** add `-race` to the CI Windows job.
- Do **not** flip `WINDOWS_GATE_REQUIRED` (operator decision, D-8).
- Do **not** pin `GODEBUG=winsymlink` module-wide (D-3).
- Do **not** touch `scripts/check-retired-architecture.sh` (Cluster B, deferred).
- Do **not** delete the GO-14 regression lock — invert and rename it.
- Do **not** re-open `AD0D9D1F` findings F5 and F10 (closed in `902c05f`).
- Do **not** reintroduce an unexplained `os.Readlink` re-resolution: a redundant one was
  removed on review in 011-S (`9158a05`). If Readlink-based resolution returns, the commit
  must state why it is no longer redundant, or reviewers will re-raise the same finding.
- Do **not** read, print, or commit any credential material.
- Do **not** claim `700B41CE` is closed against case-folding, hardlinks, or TOCTOU — the
  closure is scoped to the reparse-point mechanism only.

---

## 10. Deferred residuals (recorded in the risk register by `014.008-T`)

This dark run's scope is frozen at the 22 listed stash IDs with **no additions permitted**,
so residuals surfaced during review are recorded as **in-code risk-register entries** rather
than as new stash entries. They remain durable, discoverable, and adjacent to the acceptances
they qualify.

1. **Case-folding fail-open** — `pathEqual`/`pathHasPrefix` fold case unconditionally on
   Windows; a case-sensitivity-enabled NTFS/WSL tree admits a case-only-aliased outside
   sibling. Already partly recorded under `5FE4A7BE`.
2. **F3 in-root junction ancestor false-rejection** — fail-closed usability bug; deferred
   per D-4.
3. **GODEBUG `winsymlink` both-settings empirical proof** — the fix is GODEBUG-independent by
   construction; a subprocess re-exec harness proving it empirically is not built this cycle.

---

## 11. Independent adversarial review of this deliberation

Reviewed by three independent models in parallel — Security (`gpt-5.6-sol`), Correctness
(`gemini-3.8-flash`), Scope Boundary (`claude-opus-4.8`). The panel raised 1 P0, 4 P1, and
several P2/P3 findings against plan rev 1; all P0 and P1 findings are resolved in plan rev 2
(§10 of the plan carries the full disposition table).

Panel confirmations material to this deliberation:

- **Selection (D-1) upheld** — Cluster A is correct under the operator priority order; the
  strongest counter-argument (Cluster B is larger and more coherent) fails because a
  hygiene control cannot outrank a live, reproduced containment bypass.
- **D-2 upheld** — including `F133AB7E` is **legitimate in-scope work, not scope expansion**:
  `root.go`'s register contains an explicit flip procedure naming that exact stash, and the
  operator placed it in scope. Does not trip the scope-expansion stop condition.
- **D-9 fact-checked and confirmed** — `scripts/check-retired-architecture.sh` scans
  `internal/**` (line 83 prefix test, line 424 pathspec, lines 494–498 structural self-test)
  with **zero** remaining `internal/config` occurrences. Archiving `4989A42D` is correct and
  not premature. One P3 noted: the archival is autonomous while the operator is AFK; it is
  non-destructive (`stash archive`, retrievable) and is surfaced in the Stage report.
- **D-7 fact-checked and confirmed** — `AD0D9D1F` findings F5 and F10 are closed in `902c05f`
  (verified in `root.go:174-196`, `root_test.go:142-143`, `junction_windows_test.go:21-23`).
- **All 22 dispositions verified** — 4 selected + 18 dispositioned, none missing, none
  double-counted, none hand-waved.
