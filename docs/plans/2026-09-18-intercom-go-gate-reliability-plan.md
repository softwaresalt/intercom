---
title: "Implementation Plan — Merge-strategy governance and merge-blocking gate reliability (S-3…S-6)"
date: 2026-09-18
status: plan-review attempt 1 FAIL (4× P0) → revised; attempt 2 ADVISORY (0 P0, 4 P1) → P1s applied in revision 4
agent: Stage
source_document: docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md
governs: four release units — merge-strategy verification, engine extraction, scan-scope unification, write-path gate hardening
bound_snapshot: 14d44e3c1f28b321e8db24ab8cafcba346996133
stage_branch: chore/stage-ship-pipeline-contract-repair
requires_plan_hardening: yes
revision: 4
---

<!-- plan-review-attempt: 2 -->
<!-- attempt 2 verdict: ADVISORY. Findings applied in revision 4; see §9. No attempt 3 submitted. -->

> **Revision 3 — corrections from plan-review attempt 1 (FAIL, 4× P0, 5× P1).** The review refuted
> this plan's central premise. Summary of what changed:
>
> * **The "live scan-scope drift" claim is WITHDRAWN.** `select_repo_paths` (`:823–831`) globs
>   `['git','ls-files','--','config.toml.example','cmd/**','internal/**']` — `cmd/**` **is**
>   enumerated. The `internal/**`-only glob at `:836–837` belongs to the self-test oracle
>   `expected_internal_repo_paths` (`:834–852`) and is never on the repo-scan path. Independently
>   re-verified by Stage against the source.
> * **The unit order is inverted**: extraction now precedes scope unification (§7, H-3).
> * **The masker "clone" is not identical** and unification cannot be verdict-neutral (§4).
> * **Two merge-blocking assertions the plan never mentioned** — the `selection pathspec pin`
>   (`:1013–1059`) and the `selection cmd/ coverage` assertion (`:983–1000`) — now have explicit
>   tasks and acceptance criteria.
>
> Full finding-by-finding resolution in §8.

**Requires plan hardening: yes** — every unit here modifies a **merge-blocking control** reachable
from `ci.yml`'s required `ci-gate` job, one of which (`check-write-path-precondition.sh`) guards a
containment boundary the Brief marks NON-NEGOTIABLE. See §7.

## 1. Objective

Close a structural gap in the repository's merge-strategy guarantee, make two merge-blocking gate
scripts maintainable and testable, and close four measured evasion vectors in the write-path gate.

## 2. Unit 0 — Write-path gate kill switch (S-0, **prerequisite to Units 2 and 4**)

*(New in revision 4 — closes attempt-2 P1 #1 and P1 #2.)*

**Why this exists.** Revision 3 put the advisory toggle in U4-T2, behind all of S-4. But **U2-T5**
re-points `check-write-path-precondition.sh` at the canonical masker and **AC-2.4 explicitly
anticipates verdict deltas** from that canonicalization. Between U2-T5 and U4-T2 the gate runs in
`ci.yml`'s lint job with no `continue-on-error`, so a canonicalization delta could block every PR
including its own revert. H-1's mitigation was ordered *after* the change that most needs it.

**Why a plain `continue-on-error` is wrong.** `check-write-path-precondition.sh`'s case block
(`:276–280`) runs the **fixture self-test and the repo scan in one invocation**, and `ci.yml`
invokes only `--self-test`. Attaching `continue-on-error` to that single step would silently make
the *fixture-integrity* self-test advisory too — a broken checker would stop blocking. The
`RETIRED_ARCH_GATE_ADVISORY` pattern this was meant to mirror works precisely because it **splits** a
toggled verdict step from an unconditionally-blocking `--self-test-integrity` step.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U0-T1 | Split `check-write-path-precondition.sh` into a fixtures-only integrity mode and a verdict mode | Add `--self-test-integrity` (fixtures only, no repo scan) alongside the existing combined `--self-test`. Pure mode addition; no selector or masker change. | S | medium |
| U0-T2 | Wire two CI steps and a `WRITE_PATH_GATE_ADVISORY` toggle | `.github/workflows/ci.yml`. Unconditionally-blocking `--self-test-integrity` step; separately toggled verdict step using **step-level** `continue-on-error` (so the job result stays `success` and `ci-gate` does not fail). Update the LOCAL DIVERGENCE enumeration (`:18–41`). | S | medium |

* **AC-0.1** `--self-test-integrity` runs fixtures only and never the tracked-tree repo scan.
* **AC-0.2** The integrity step is **unconditionally blocking**; the advisory toggle cannot disable it.
* **AC-0.3** The verdict step honours `WRITE_PATH_GATE_ADVISORY` via **step-level**
  `continue-on-error`.
* **AC-0.4** The LOCAL DIVERGENCE enumeration lists both steps.
* **AC-0.5** Verdicts at the bound snapshot are unchanged — this unit adds a mode and a toggle only.

**Ordering: S-0 must merge before U2-T5 and before Unit 4.** This is a hard constraint, not a
preference.

## 3. Unit 1 — Merge-strategy structural guarantee (S-3)

**Problem.** Constitution Principle XI / P-009 require merge-commit-only. Measured 2026-09-17 on
`softwaresalt/intercom`: `allow_merge_commit`, `allow_squash_merge` and `allow_rebase_merge` are all
`true`. The guarantee rests on per-merge operator discipline, not on structure.

**Authority split (deliberation D-5).** Flipping the settings is GitHub repository administration —
outside both Stage's and Ship's role boundary. The *standing verification* is implementable and is
what this unit ships, **advisory-first**, mirroring the existing `PIPELINE_TOPOLOGY_GATE_REQUIRED`
toggle in `ci.yml`.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U1-T1 | Add `scripts/check-merge-strategy.sh` | Query merge settings via `gh api repos/{owner}/{repo}` and assert `allow_squash_merge == false` and `allow_rebase_merge == false`. `--self-test` drives stubbed JSON fixtures so the checker is testable with no network and no token. | M | medium |
| U1-T2 | Measure token feasibility and record the credential requirement | *(new in rev 3, was P2)* Determine empirically whether the workflow's default `GITHUB_TOKEN` can read `allow_squash_merge`/`allow_rebase_merge`. Record the finding and, if it cannot, name the credential that promotion to required would need. | S | medium |
| U1-T3 | Wire the checker into CI behind a `MERGE_STRATEGY_GATE_REQUIRED` toggle **and into `ci-gate`** | `.github/workflows/ci.yml`. Advisory (non-blocking) when unset; blocking when set. *(rev 3, was P2)* The job **must** be added to `ci-gate`'s `needs:` array (`ci.yml:610`, results string at `:614`) or it can never block a merge even once promoted. *(rev 4)* Advisory mode must be implemented as **step-level** `continue-on-error` so the job result stays `success` — a job-level failure would fail `ci-gate` and contradict AC-1.2. | S | medium |
| U1-T4 | Update the `ci.yml` LOCAL DIVERGENCE enumeration | *(new in rev 3, was P2)* `ci.yml` is autoharness-generated; its header (`:18–41`) enumerates the local divergences that must be re-applied after regeneration. Add the merge-strategy job so it is not silently dropped at the next install/tune. | XS | trivial |
| U1-T5 | Document the operator trigger and the advisory→required promotion | Record: the two settings to disable, how to verify, the credential from U1-T2, and the exact promotion condition. | S | low |

### Acceptance criteria

* **AC-1.1** `--self-test` covers: both settings false (pass), squash enabled (fail), rebase enabled
  (fail), both enabled (fail), and unauthorized/absent API response (**skip**, never pass).
* **AC-1.2** The CI job is advisory by default and blocks no PR on merge.
* **AC-1.3** An unauthorized or failed API read reports `SKIPPED` with a reason and never a pass.
* **AC-1.4** The job appears in `ci-gate`'s `needs:` array and in its results aggregation string.
* **AC-1.5** The `ci.yml` LOCAL DIVERGENCE enumeration includes the new job.
* **AC-1.6** The doc names the two settings, the verification command, the required credential, and
  the promotion trigger.
* **AC-1.7** No repository setting is changed by this unit. That remains operator-only.

## 3. Unit 2 — Extract gate Python engines to importable modules (S-4) *(was S-5; now first)*

**Problem.** A 1053-line Python engine is sealed inside a quoted bash heredoc
(`check-retired-architecture.sh:123–1176`), so it cannot be imported, unit-tested in isolation, or
linted. This is the root cause of the duplicated Go masker shared with
`check-write-path-precondition.sh`.

**Why this is now first (rev 3).** Revision 2 put scope unification first on the strength of a
"broken scope" that does not exist. With that premise withdrawn, the only real constraint is that
every subsequent change is cheaper and safer inside a real module — and doing scope unification
first would force the merge-blocking `selection pathspec pin` assertion (`:1013–1059`) to be
re-derived **twice**.

**Two hazards this unit must handle explicitly (rev 3, both were P0/P1):**

1. **The disk-read pin.** `:1013–1059` locates the Python by *reading the shell script from disk*
   (because `__file__` and `inspect.getsource()` are unavailable under a heredoc) and asserts scope
   literals appear inside the `select_repo_paths` / `should_scan_repo_path` regions. After
   extraction the wrapper no longer contains those definitions, `extract_function_region` returns
   `None`, and the assertion **fails closed**, blocking CI. It must be re-pointed as part of this
   unit.
2. **The masker is not an identical clone.** The retired-architecture version (`:432–548`) has
   raw-string buffering, struct-tag unmasking via `struct_tag_re`, and an unterminated-raw-string
   EOF fail-closed branch; the write-path version (`check-write-path-precondition.sh:91–175`) has
   none. The divergence is documented as deliberate at `check-retired-architecture.sh:92–104`, which
   explicitly leaves the convergence direction undecided. Unification therefore **cannot** be
   verdict-neutral for the write-path gate.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U2-T1 | Record the canonical Go masker decision | **AC-2.3 already forecloses the choice**: adopting the write-path masker would drop `struct_tag_re` unmasking and change retired-arch verdicts on the 015.007-T struct-tag and go-differential fixtures. The canonical implementation is therefore the **retired-arch superset**. This task records the rationale against the undecided note at `:92–104` and **enumerates the expected write-path verdict deltas**. Decision record only; no code moves. | S | medium |
| U2-T2 | Extract the canonical masker to `scripts/lib/gomask.py` | New module. Behaviour-preserving with respect to the chosen canonical implementation. | M | medium |
| U2-T3 | Extract the retired-architecture engine to `scripts/lib/retired_arch.py` | New module importing `gomask`. Carries the three dispatch modes and module-level state (`forbidden_parts`, `fixture_suites`, `struct_tag_re`, the `tomllib`-absent fallback). *(rev 4, was P2)* **Must also convert the two import-time bindings at `:137–139` — `mode = sys.argv[1]` and `root = Path.cwd()`** — into a function parameter and an explicit repo-root resolver. Left as-is, importing the module raises `IndexError` under any argv the caller did not fake and silently binds the importer's cwd, defeating AC-2.8 and making U3-T3's unit test cwd-dependent. | M | medium |
| U2-T4 | Reduce `check-retired-architecture.sh` to an argv-dispatch wrapper and re-point the disk-read pin | Heredoc removed. The `selection pathspec pin` assertion (`:1004–1059`) is re-derived to read `scripts/lib/retired_arch.py` — now an ordinary importable module, so `inspect.getsource()` is available and the disk-read hack can be retired. **The pin stays source-text-anchored**; it must not degrade into a module-constant self-comparison (see U3-T3). | M | medium |
| U2-T5 | Re-point `check-write-path-precondition.sh` at `scripts/lib/gomask.py` and re-baseline its verdicts | Delete the local masker; import the shared module. Capture and justify every verdict delta arising from U2-T1's canonicalization. **Blocked by S-0** — must not merge before the kill switch exists. | M | medium |
| U2-T7 | Wire a Python lint/test step into CI | *(new in rev 4, was P2)* `ci.yml`'s lint job runs `gofmt`/`goimports`/`golangci-lint` and the three shell gates only; there is **no** pytest/ruff/mypy step anywhere in `.github/workflows`, so AC-2.8's "importable and lintable" and U3-T3's derivation-agreement test would never execute in CI. Add a pinned Python lint/test step (and add it to the LOCAL DIVERGENCE enumeration), **or** require the new assertions to live inside the `--self-test-integrity` path CI already runs unconditionally. | S | medium |
| U2-T6 | Capture pre/post behaviour-preservation evidence | Evidence task, no production change. Capture set is **named**: `check-retired-architecture.sh` bare repo scan, `--self-test`, `--self-test-integrity`; and `check-write-path-precondition.sh --self-test`. Captured at the bound snapshot and again post-extraction. | S | low |

### Acceptance criteria

* **AC-2.1** The canonical Go masker is defined in exactly one place, and U2-T1's rationale is
  recorded in-repo alongside the superseded note at `check-retired-architecture.sh:92–104`.
* **AC-2.2** All four named capture-set runs pass post-extraction. **Precedence over AC-2.4
  (rev 4):** `check-write-path-precondition.sh --self-test` runs fixtures *and* the tracked-tree
  repo scan in one invocation (`:276–280`), so the two criteria can collide. Rule: enumerated
  deltas are permitted **only on fixtures**. Any delta landing on a tracked `cmd/**` or
  `internal/**` file must be remediated within the same change, or it is a defect.
* **AC-2.3** `check-retired-architecture.sh` verdicts are **identical** pre- and post-extraction
  across all three of its modes. Any difference is a defect.
* **AC-2.4** `check-write-path-precondition.sh` verdict deltas, if any, are **exactly** those
  enumerated and justified by U2-T1. An unenumerated delta is a defect. *(This replaces revision 2's
  unsatisfiable "no detection behaviour changes" for this script.)*
* **AC-2.5** The `selection pathspec pin` assertion passes post-extraction, re-derived against the
  extracted module rather than the shell script.
* **AC-2.6** The `selection cmd/ coverage (AG-5/D4)` assertion (`:983–1000`) still passes.
* **AC-2.7** `check-retired-architecture.sh` contains no embedded Python heredoc.
* **AC-2.8** `scripts/lib/*.py` are importable and lintable as ordinary modules.

## 4. Unit 3 — Scan-scope unification inside the extracted module (S-5) *(was S-4; demoted and re-ordered)*

**Problem, corrected.** 6C24E2E4 item (1) reports the scan scope is expressed twice, in two
vocabularies, ~630 lines apart: a pathspec literal in `select_repo_paths` (`:825`) and a prefix test
in `should_scan_repo_path` (`:188–197`). That duplication is **real**.

**What is NOT true (rev 3).** The two do **not** disagree. `cmd/**` is both enumerated and admitted.
There is no live drift, and no surface is silently unscanned. The duplication is additionally
**already mechanically pinned** by the `selection pathspec pin` assertion, which is exactly the
mitigation 6C24E2E4 itself describes ("013-F mitigates this only with assertions, not by removing
the duplication"). This unit therefore removes a maintenance hazard; it does **not** fix a live
defect, and it is ranked accordingly.

**Deliberate asymmetry that must survive (rev 3, was P2).** `should_scan_repo_path` is not a flat
prefix disjunction. `cmd/` intentionally **includes** `_test.go` and `testdata/` paths while
`internal/` excludes them, and `config.toml.example` is a third arm. This is a guarded decision
(015.011-T, AG-5/D4) locked by the `selection cmd/ coverage` assertion. A naive "one prefix list"
unification would silently narrow `cmd/` coverage while appearing to satisfy the criterion.

**Blocked by S-4.** Doing this inside a real, importable module lets the derivation be unit-tested
directly and avoids re-deriving the pin twice.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U3-T1 | Introduce a single `SCAN_SCOPE` declaration carrying per-arm test-file policy | `scripts/lib/retired_arch.py`. One structure declaring, per scope arm, its pathspec form **and** whether `_test.go`/`testdata/` are included. | M | medium |
| U3-T2 | Derive both the `ls-files` argument vector and the predicate from `SCAN_SCOPE` | Same module. Neither restates a scope literal. | M | medium |
| U3-T3 | Re-derive the pathspec pin against `SCAN_SCOPE` and add a derivation-agreement unit test | Same module. *(rev 4, closes attempt-2 P1 #4)* The pin **must remain source-text-anchored** — via `inspect.getsource()` on the `SCAN_SCOPE` declaration in the extracted module, or against a committed golden selected-path fixture. It must **not** become "assert `SCAN_SCOPE` contains `cmd/**`", which is a self-comparison: the in-source rationale at `check-retired-architecture.sh:1004–1013` states that nothing outside the process could ever disagree with such an assertion, which is exactly why the original pin reads source text. Degrading it would satisfy AC-3.3 while destroying the AC-6/AG-1 control. A new unit test asserts the pathspec-selected and predicate-selected sets are identical **and** that the `cmd/`-includes-tests asymmetry is preserved. | M | medium |

**Red phase — declared honestly (rev 3).** The derivation-agreement half of U3-T3 is
**green-on-arrival**: the two derivations already agree, as the existing `selection cmd/ coverage`
assertion proves. This unit is therefore **characterization-and-refactor**, not defect-repair, and
this plan does not claim a red phase for it. The genuinely new assertion is the *asymmetry
preservation* check, which is red against a naive flat-prefix unification — i.e. it guards the
refactor, which is its actual purpose.

### Acceptance criteria

* **AC-3.1** Exactly one `SCAN_SCOPE` declaration exists; neither the pathspec construction nor the
  predicate restates a scope literal.
* **AC-3.2** The `cmd/` includes-test-files / `internal/` excludes-test-files asymmetry and the
  `config.toml.example` arm are all preserved, and each is asserted.
* **AC-3.3** The re-derived pathspec pin passes, and the `selection cmd/ coverage (AG-5/D4)`
  assertion still passes unmodified in intent.
* **AC-3.4** The selected path set is **identical to the set computed at the change's own parent
  commit** — not at the bound snapshot. *(rev 4, was P3: S-5 runs after S-4 and parallel to S-6, so
  any unrelated file added under `cmd/**` or `internal/**` in the interim would otherwise fail this
  criterion spuriously while the refactor is correct.)*
* **AC-3.5** All pre-existing `--self-test` and `--self-test-integrity` cases still pass.
* **AC-3.6** The governing decision (D6/D6c, as broadened by 012-S) is cited in the declaration so a
  future scope change is a one-line edit that cannot half-apply.

## 5. Unit 4 — Write-path precondition gate hardening and Resolve-caller tripwire (S-6)

**Problem (measured).** `check-write-path-precondition.sh:68–79` compiles 20 literal qualified
selectors with a `(?<![\w.])` lookbehind. Four evasion vectors are confirmed undetected:
(a) import aliasing (`import fs "os"; fs.WriteFile(...)`); (b) `syscall.CreateFile` /
`syscall.Write`; (c) Go 1.24 `os.Root` file-creating methods; (d) DB drivers beyond `sql.Open` /
`bbolt.Open`. There is no tripwire for `pathsafe.Root.Resolve` gaining its first production caller.

**Decidability correction (rev 3, was P1).** The gate's documented design (`:15–17`, `:27–32`)
matches **qualified selectors only** and explicitly refuses bare-identifier alternation because of
false positives; `(*os.File).Write` is already an accepted residual for that reason. `root.Create`
carries no package qualifier, and the corpus already binds a local named `root`
(`internal/config/validate.go:42`). Detecting `os.Root` *methods* is therefore **not decidable** by
this gate's design. Corrected approach: add the **decidable construction sites** `os.OpenRoot` and
`os.Root`, and record method-call detection as an explicit residual under AC-4.7.

**Advisory toggle required (rev 3, was P1).** `ci.yml`'s lint job runs this checker with **no**
`continue-on-error`, unlike the adjacent retired-arch verdict step gated by
`RETIRED_ARCH_GATE_ADVISORY`. A false positive from this unit would block every PR *including the
PR that reverts it*. The hardened selectors must land behind an equivalent advisory toggle.

**Blocked by S-4.** Hardening before extraction would re-clone the masker S-4 exists to unify.

**Scope fence.** This unit hardens **detection**. It does **not** implement the TOCTOU / hardlink /
dangling-symlink mitigation, which remains blocked on the unfired trigger (a live destructive write
path; re-measured absent). Adding a write primitive is a named stop condition.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U4-T1 | Extend the selector list with decidable qualified entry points | Add `syscall.CreateFile`, `syscall.Write`, `os.OpenRoot`, `os.Root`. *(rev 4, was P2)* The "unlisted DB-driver `Open` selectors" item is **removed as unbounded and unfalsifiable** — `go.mod` declares no database dependency at all (BurntSushi/toml, copilot-sdk, cobra + indirects), so even the existing `sql.Open`/`bbolt.Open` selectors are speculative. Vector (d) is **deferred as an unfired trigger**, consistent with this plan's own stop-condition discipline. | S | medium |
| U4-T1b | Add call-extent extraction and an access-mode allowance predicate | *(new in rev 4, closes attempt-2 P1 #3)* AC-4.2 is unsatisfiable under a line-oriented scanner: `scan_file` iterates `masked.splitlines()` and matches per line, while the live call in `internal/pathsafe/reparse_windows.go` spans multiple lines with the desired-access literal `0` on the line **after** the `syscall.CreateFile` token. No allowance mechanism exists today at all. This task builds call-extent (argument-aware) extraction plus the allowance predicate AC-4.2 and AC-4.3 depend on. | L | high |
| U4-T2 | *(moved to Unit 0 in revision 4)* | The advisory toggle and the `--self-test-integrity` split are now **S-0**, a prerequisite that must merge before U2-T5 and before this unit. See §2. | — | — |
| U4-T3 | Resolve **named** import aliases to canonical package paths | *(split in rev 3, was over-sized)* Parse the import block; map named aliases (live in-repo: `copilot "github.com/github/copilot-sdk/go"`) to package paths; match on the canonical path. | M | medium |
| U4-T4 | Record the undecidable alias residuals explicitly | *(new in rev 3)* Dot-imports (`import . "os"` → a bare `WriteFile(` call, undecidable without type info), blank imports, and local identifiers shadowing an import name are **out of scope and documented as residual**, not silently unhandled. | S | low |
| U4-T5 | Add a `Root.Resolve` first-caller tripwire with a stated matching rule | *(rule added in rev 3, was P3)* Fire only when a file both imports `internal/pathsafe` **and** contains a `.Resolve(` call on a receiver bound from `pathsafe.NewRoot`. A bare `.Resolve(` text match is insufficient and would false-positive on unrelated code. This is the detection mechanism for the blocked 37FAB8C2 and BF5DE670-mitigation triggers. | M | medium |
| U4-T6 | Add `--self-test` fixtures for every vector, including a positive control | Fixtures for named-alias write, `syscall` write, `os.OpenRoot`, an unlisted DB driver, and a **positive control**: a *writing* `syscall.CreateFile` inside an allowance-scoped file must still be **rejected**. All fixtures must be `gofmt`/`goimports` clean, since `ci.yml`'s lint job formats the whole tree including `testdata`. | M | medium |

### Acceptance criteria

* **AC-4.1** `--self-test` catches every vector in U4-T6; each fixture fails pre-change.
* **AC-4.2** The existing metadata-only `syscall.CreateFile` in `reparse_windows.go` does not trip
  the gate, and the allowance is keyed on the **call's access mode**, not on the file.
* **AC-4.3** *(positive control)* A *writing* `syscall.CreateFile` in the same allowance-scoped file
  is still **rejected**. An allowance that fails this criterion is a defect.
* **AC-4.4** The tripwire fires on a fixture introducing a production `Root.Resolve` caller and does
  not fire at the bound snapshot (zero callers, re-measured), using U4-T5's stated matching rule.
* **AC-4.5** No TOCTOU/hardlink mitigation is added. No write primitive is introduced anywhere.
* **AC-4.6** The hardened gate runs behind the **S-0** advisory toggle (AC-0.3), with S-0 already
  merged; the LOCAL DIVERGENCE enumeration is updated.
* **AC-4.7** The gate's documentation states its residual evasion surface honestly — at minimum
  `os.Root` method calls, dot-imports, and identifier shadowing — and does not present itself as a
  complete mechanical tripwire.
* **AC-4.8** Zero false positives across the 17 existing non-test `.go` files.
* **AC-4.9** All new fixtures are `gofmt` and `goimports` clean.

## 6. Dependencies

* **S-0 blocks S-4's U2-T5** and **blocks S-6 entirely** — the kill switch must exist before any
  change that can move write-path verdicts (attempt-2 P1 #1).
* **S-4 blocks S-5** (unify scope inside a real module; avoid re-deriving the pin twice).
* **S-4 blocks S-6** (avoid re-cloning the masker).
* **S-5 and S-6 are independent of each other.** No edge is invented between them.
* **S-3 is independent** of all of the above and of Plan 1.
* S-6 is a prerequisite *detector* for blocked 37FAB8C2 and the blocked BF5DE670 mitigation; those
  carry dependency edges on S-6 so their triggers are monitored rather than merely awaited.

## 8. Plan Hardening

**H-1 — Merge-blocking controls with no kill switch.** The write-path checker runs in `ci.yml`'s
lint job with no `continue-on-error`; a false positive blocks every PR including its own revert.
Mitigation *(corrected in rev 4)*: **S-0 is a standalone prerequisite** landing the
`--self-test-integrity` split and the step-level advisory toggle **before** U2-T5 (whose
canonicalization AC-2.4 expects to move verdicts) and before Unit 4. Revision 3 placed the toggle
behind all of S-4, leaving U2-T5 unprotected.

**H-2 — Refactor masquerading as improvement.** Mitigation: AC-2.3 requires byte-identical
retired-arch verdicts; AC-2.4 requires every write-path delta be pre-enumerated and justified by
U2-T1, replacing revision 2's unsatisfiable zero-delta claim.

**H-3 — Ordering was derived from a false premise.** Revision 2 ordered scope-unification first to
fix drift that does not exist. Corrected: S-4 (extract) → S-5 (unify) and S-4 → S-6 (harden), with
S-5 ∥ S-6. Only the constraints that survive measurement are encoded.

**H-4 — Invisible merge-blocking assertions.** Two assertions the earlier revision never mentioned
— the disk-read `selection pathspec pin` (`:1013–1059`) and `selection cmd/ coverage`
(`:983–1000`) — would have failed closed mid-refactor. Mitigation: U2-T4 and U3-T3 own them
explicitly; AC-2.5, AC-2.6, AC-3.3 assert them.

**H-5 — Over-broad allowance re-opening the hole.** Mitigation: AC-4.2 keys the allowance on access
mode, and AC-4.3 adds a positive control that fails if the allowance is file-scoped.

**H-6 — Asserting an undecidable capability.** Revision 2 promised `os.Root` *method* detection the
gate's design cannot deliver. Mitigation: U4-T1 detects decidable construction sites; U4-T4 and
AC-4.7 record the residual rather than claiming coverage.

**H-7 — Advisory theatre.** An advisory gate never promoted is no gate. Mitigation: AC-1.4 requires
`ci-gate` wiring so promotion is *possible*; AC-1.6 requires the trigger be written down; U1-T2
measures the credential that promotion needs.

**H-8 — Generated-file drift.** `ci.yml` is autoharness-generated. Mitigation: U1-T4 and AC-4.6
require the LOCAL DIVERGENCE enumeration be updated in the same change.

**H-9 — Task sizing.** Revision 2's "reduce a 1199-line script to a wrapper" and "resolve import
aliases" were both over-sized. Mitigation: U2 is split into seven tasks with the canonicalization
*decision* separated from the code move; U4's alias work is split into named-alias resolution
(U4-T3) and an explicit residual record (U4-T4). *(rev 4)* U4-T1b is sized **L/high** honestly
rather than being smuggled into U4-T1 at size S.

**H-10 — Assertions with no execution vehicle.** *(new in rev 4)* There is no Python lint/test step
in any workflow, so the extracted modules' unit tests would never run in CI. Mitigation: U2-T7 adds
one, or relocates the assertions into the unconditionally-run `--self-test-integrity` path.

**H-11 — Self-comparison pins.** *(new in rev 4)* The pathspec pin's whole value is that it is
anchored to source text an outside process can disagree with. Mitigation: U3-T3 forbids degrading
it to a module-constant self-comparison.

## 9. Review findings applied (attempt 2 → revision 4)

| Finding | Sev | Resolution |
|---|---|---|
| Advisory toggle ordered **after** U2-T5, the change AC-2.4 expects to move write-path verdicts | P1 | **Unit 0 (S-0)** created as a hard prerequisite to U2-T5 and Unit 4; H-1 rewritten |
| `continue-on-error` on the combined `--self-test` step would silently make the fixture-integrity check advisory | P1 | U0-T1 splits `--self-test-integrity`; AC-0.2 keeps it unconditionally blocking |
| AC-4.2 unsatisfiable — `scan_file` is line-oriented, the live call spans lines, no allowance mechanism exists | P1 | **U4-T1b** added (size L/high) for call-extent extraction + allowance predicate |
| U3-T3's pin would degrade into a tautological self-comparison, destroying the AC-6/AG-1 control | P1 | U3-T3 now mandates `inspect.getsource()` / golden-fixture anchoring |
| AC-2.2 ⊥ AC-2.4 — `--self-test` runs fixtures *and* the repo scan | P2 | Precedence rule added to AC-2.2: deltas permitted on fixtures only |
| `mode = sys.argv[1]` and `root = Path.cwd()` at `:137–139` block importability | P2 | Named in U2-T3's scope |
| No pytest/ruff/mypy step exists anywhere — new assertions would never run | P2 | **U2-T7** added; H-10 |
| "unlisted DB-driver selectors" unbounded and unfalsifiable; `go.mod` has no DB dependency | P2 | Item removed; vector (d) deferred as an unfired trigger |
| Frontmatter pre-declared its own verdict | P2 | Status records the actual attempt-2 verdict |
| AC-3.4 pinned to the bound snapshot would fail spuriously | P3 | Re-pinned to the change's own parent commit |
| U2-T1's admissible outcome set has exactly one member | P3 | Task reduced to recording the rationale + delta enumeration |
| `ci-gate` `needs:` citation drift; advisory mode must be step-level | P3 | U1-T3 corrected to `:610`/`:614` and step-level `continue-on-error` |
| `selection pathspec pin` cited as `:1013–1059` | P3 | Corrected to `:1004–1059` |

## 10. Review findings applied (attempt 1 → revision 3)

| Finding | Severity | Resolution |
|---|---|---|
| Unit 2's central premise false — `cmd/**` **is** enumerated at `:825` | P0 | Claim withdrawn; deliberation M-5 corrected (§10 C-1); unit demoted to duplication-removal and re-ordered |
| Declared red phase cannot go red (derivations already agree) | P0 | §4 declares the unit characterization-and-refactor with **no** claimed red phase; the new assertion guards the refactor instead |
| AC-2.1 ⊥ AC-2.4 — `selection pathspec pin` reads literals from disk | P0 | U2-T4 and U3-T3 own the pin re-derivation; AC-2.5 and AC-3.3 assert it |
| Maskers are not identical clones; unification not verdict-neutral | P0 | U2-T1 chooses and justifies a canonical implementation; AC-2.4 requires enumerated, justified deltas |
| Extraction breaks the disk-read pin mechanism | P1 | U2-T4 retires the disk-read hack in favour of `inspect.getsource()` on a real module |
| `os.Root` method detection not decidable by this gate's design | P1 | U4-T1 targets `os.OpenRoot`/`os.Root` construction sites; AC-4.7 records the residual |
| AC-4.2 allowance satisfiable by a file-level exemption | P1 | AC-4.2 keys on access mode; AC-4.3 adds a positive control |
| No kill switch on the most false-positive-prone unit | P1 | U4-T2 lands an advisory toggle (AC-4.6) |
| U4-T2 alias resolution over-sized and under-specified | P1 | Split into U4-T3 (named aliases) and U4-T4 (documented residuals) |
| U2/U3 sizing optimistic; tasks not separable | P2 | Unit 2 re-split into six tasks; decision separated from code move |
| Merge-strategy job absent from `ci-gate` `needs:` | P2 | U1-T3 + AC-1.4 |
| `ci.yml` LOCAL DIVERGENCE list not updated | P2 | U1-T4 + AC-1.5, AC-4.6 |
| `gh api` token feasibility unverified | P2 | U1-T2 + AC-1.6 |
| Naive unification would destroy the guarded `cmd/` asymmetry | P2 | §4 states the asymmetry; AC-3.2 asserts it |
| Citation drift (`:65–78`, `:861`, `:193–194`) | P3 | Corrected throughout to `:68–79`, `:862`, `:188–197` |
| Tripwire matching rule unstated | P3 | U4-T5 states the rule; AC-4.4 binds to it |
| Fixture formatting would break the lint job | P3 | U4-T6 + AC-4.9 |
| AC-3.2 capture set unnamed | P3 | U2-T6 names all four runs |
