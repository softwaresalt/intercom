---
title: "Deliberation — Post-M4 stash triage and Go re-plan of held shipments 030-S/031-S/032-S"
description: "Dispositions for 8 post-M4 deferred-scope-expansion stash entries, and the retarget of held plan units 6/7/8 from the retired Python gate engines onto the Go engines under tools/gatecheck"
topic: "Post-M4 follow-ups: stash triage + Go re-plan of 030-S (033-F), 031-S (034-F), 032-S (035-F)"
depth: "deep"
decision_status: "decided"
promoted_to: "plan (030-S, 032-S, 038-S; 031-S revised per D-031-2; new follow-up 049-F/039-S)"
agent: Stage
date: 2026-10-01
stage_branch: chore/stage-post-m4-followups-and-go-replan
bound_snapshot: 9b299c8 (origin/main, PR #88 merged; 037-S/047-F closed)
source_stash:
  - 44F8CC48
  - D10D3AFC
  - C312BD4C
  - 7223218F
  - 5A8EC1BC
  - 978D2946
  - 50E6F22C
  - 9F824B64
  - 8E9F8E55  # D-031-2 fold-in (034.009-T)
  - B72E9715  # D-031-2 fold-in (034.004-T)
  - DD0BB60F  # D-030-4 harvest (033.004-T + 033.005-T + 033.006-T), 2026-10-01 follow-up session
  - 1EEBECA5  # D-049-2 spike executed (049.001-T Stage half), 2026-10-02
  - 458F9385  # D-049-3 fold-in (049.007-T, E-T7), 2026-10-02
captured_stash:
  - 458F9385  # D-S-5 item-6 selector widening (task, low)
  - 1EEBECA5  # D-049-1 spike-schedule trigger (spike, high)
  - FE2F02FF  # D-049-5 residual item 8, DEFERRED SCOPE EXPANSION (2026-10-02)
  - C0D28448  # D-049-3 item-6 remainder, DEFERRED SCOPE EXPANSION (2026-10-02)
related_shipments:
  - 030-S (033-F)
  - 031-S (034-F)
  - 032-S (035-F)
  - 039-S (049-F)  # D-031-2 follow-up
---

# Deliberation — Post-M4 stash triage and Go re-plan

## 0. Framing

The Python→Go gate-engine migration (034-S..037-S / features 044-F..047-F) is
complete and closed at `9b299c8`. The gate engines now live in Go under
`tools/gatecheck`; `scripts/lib/*.py` is retired (only `gatecheck-run.sh` and
`OutputPathGuard.ps1` remain under `scripts/lib/`).

Two consequences drive this deliberation:

1. Eight deferred-scope-expansion stash entries accumulated across the M2/M3/M4
   shipments and need dispositions.
2. Held shipments 030-S, 031-S and 032-S were planned against the **now-retired
   Python engines**. Every one of their 14 tasks names Python identifiers
   (`scripts/lib/retired_arch.py`, `scan_file`, `masked.splitlines()`,
   `root_gitignore_text_at`, `SystemExit`, `inspect.getsource()`). They carry a
   STAGE HOLD: "re-plan onto Go required before claim."

### Blocking matrix — re-measured at this snapshot

| Shipment | Declared blocks | Status at `9b299c8` |
|---|---|---|
| 030-S | 029-S, 037-S | both **archived** (shipped) |
| 031-S | 027-S, 029-S, 037-S | all three **archived** (shipped) |
| 032-S | 037-S | **archived** (shipped) |

**All dependency edges on the held shipments are satisfied.** No shipment is
blocked by an unshipped predecessor. The only remaining gate is the STAGE HOLD
itself, which this deliberation exists to resolve.

Per backlogit convention, satisfied `blocks` edges on archived shipments are
**retained, not removed** — they are the historical sequencing record, and
removing them would erase the provenance of why these units waited. They do not
gate claim because the predecessors are archived.

---

## 1. Stash triage — all 8 entries

All eight entries carry the literal `DEFERRED SCOPE EXPANSION` marker, so per the
Step 1 precedence rule every one is forced onto the `deliberate` route (P-021 C6)
regardless of shape or size. This document is that deliberation artifact.

### 1.1 Duplicate detection (P-021 C5, unconditional)

Ran over all 27 active stash entries across eight theme patterns
(root-injection, gitrunner, masker, safe-close, file-lock, crlf, gitignore,
containment). **Result: CLEAN.** No two of the eight target entries describe the
same expansion, and no third entry duplicates any of them.

The scan did surface four *adjacent but distinct* out-of-scope entries that bear
on these re-plans; they are recorded in §4 as recommended fold-ins, not merged.

### 1.2 Late-identifier reconciliation (P-021 C6)

Recovered from Ship-owned residual-risk/closure records under Stage's own stash
authority. No Ship write required; the existing entries are updated in place.

| Entry | Field was `N/A` | Outcome |
|---|---|---|
| 44F8CC48 | PR, review-thread | **PR #87 recovered** (037-S implementation PR, `docs/closure/037-S-047-F-post-merge-closure.md`). Review-thread `N/A` **stands** — pre-PR local adversarial review, never a GitHub thread. |
| D10D3AFC | task, review-thread | task `N/A` **stands** (no implementing task exists, truthful). Review-thread: **no late thread ID found** — the PR #86 Copilot findings are inline comment IDs (4151321043/…103/…194/…224/…255), not GraphQL review threads. PR #86 already recorded. |
| C312BD4C | — | **Already reconciled** 2026-09-28 (PR #77 recovered). Re-run is a **no-op** (idempotent); no identifier overwritten. |
| 7223218F | PR, review-thread | **PR #85 recovered** (036-S implementation PR). Review-thread `N/A` **stands** — local persona review. |
| 5A8EC1BC | PR, review-thread | **PR #85 recovered**. Review-thread `N/A` **stands** — local persona review. |
| 978D2946 | PR, review-thread | **PR #85 recovered**. Review-thread `N/A` **stands** — local persona review. |
| 50E6F22C | — | Clean at capture (PR #85 + three thread IDs). No reconciliation needed. |
| 9F824B64 | task | task `N/A` **stands** (no implementing task). PR #88 + thread `PRRT_kwDOTPuhps6oFWve` already recorded. |

A `N/A` that stands is a **truthful terminal record**, not a shortfall, and never
gates planning.

### 1.3 Verification against current code

Every claim was re-verified against the Go tree at `9b299c8` before disposition.
No entry was accepted on its own say-so.

| Entry | Claim | Verified? |
|---|---|---|
| 44F8CC48 | `.gitignore:126-127` comment still says the gate engines live under `scripts/lib/` | **CONFIRMED** — comment reads "the extracted gate engines under `scripts/lib/` are imported or executed (032-F, shipment 029-S)"; those engines no longer exist. Rule is now purely defensive. |
| D10D3AFC | `shipment-reconcile` `mode: safe-close` has no invokable implementation | **CONFIRMED** — skill exists at `.github/skills/shipment-reconcile/`; no CLI/gate subcommand implements safe-close. |
| C312BD4C | Go masker unmasks multiline tag-shaped raw strings, contradicting its own comment | **CONFIRMED** — `isStructTag` (`gomask.go:205`) uses `pysem.IsSpace`, which includes `\n`; the doc comment at `gomask.go:138-141` still claims "multiline strings … keep the fully-masked behavior". Contradiction **survived the port**. Behaviour is pinned by `TestMaskGoNonCode_MultilineTagShapedRawStringPinned` (`gomask_test.go:57`), whose own comment names C312BD4C and says it "must not be fixed here". |
| 7223218F | `check-write-path-precondition.sh` has no `--root` denylist | **CONFIRMED** — `parseRoot` (`main.go:74-100`) scans the whole arg list with **last-occurrence-wins** (`root = args[i+1]`, no `break`); the wrapper does `gatecheck_invoke write-path "$@"` (`:84`) with **no guard**, while retired-arch (allowlists first arg only, `:231-249`) and unignore (per-arg denylist, `:94-99`) each carry their own differently-shaped guard. **Live gap.** |
| 5A8EC1BC | Three independent `GitRunner` shapes; exported mutable `Denylist`/`Selectors` | **CONFIRMED** — `writepath.go:50` `func(root string) ([]byte, error)`; `retiredarch/select.go:28` `func(root string, pathspecs ...string) ([]byte, error)`; `unignore/git.go:38` `func(dir string, stdin []byte, args ...string) ([]byte, []byte, error)`. `writepath.Selectors` (`:32`) and `unignore.Denylist` (`checks.go:18`) are exported mutable package vars. |
| 978D2946 | No automated CRLF/line-ending gate | **CONFIRMED** — `.gitattributes` declares `eol=lf` but no pre-commit hook or CI step enforces it. |
| 50E6F22C | mergestrategy discards `root`; temp dirs outside root; extra args ignored | **CONFIRMED** — `evaluate.go:192` `_ = root`; `os.ReadFile(path)` at `:59` with no containment check; `path := args[0]` at `:197` with no argv-length validation. All three are faithful parity with the retired Python. |
| 9F824B64 | `scripts/acquire_lock.ps1`/`.sh` do not exist | **CONFIRMED** — no lock scripts anywhere under `scripts/`. The `file-lock` skill exists but its scripts do not. |

**All eight entries are valid and still live.** None is obsolete; none targets
code that no longer exists in a way that voids the finding. 44F8CC48 and
C312BD4C both *reference* retired Python, but each has a live Go-side or
contract-surface successor.

### 1.4 Thematic grouping

Grouped per the Orchestrator heuristic; artifact classes are **not** mixed.

| Group | Class | Entries | Disposition |
|---|---|---|---|
| **G1 — gatecheck CLI containment hardening** | Go tooling | 7223218F (high), 50E6F22C (medium) | **HARVEST** → new feature + shipment. 50E6F22C's own text proposes this consolidation. |
| **G2 — harness/skill closure tooling** | template/skill | D10D3AFC (high), 9F824B64 (medium) | **DEFER** with recorded rationale → needs its own design deliberation. 9F824B64's own text says "triage alongside D10D3AFC". |
| **G3 — gatecheck internal architecture** | Go tooling | 5A8EC1BC (low) | **DEFER** — trigger not yet fired (see §1.5). |
| **G4 — masker semantics** | Go tooling | C312BD4C (low) | **FOLD** into 030-S re-plan as a scoped decision; see §2.1. |
| **G5 — CI/line-ending hygiene** | CI tooling | 978D2946 (medium) | **DEFER** — standalone, no dependency on the re-plans. |
| **G6 — .gitignore comment** | gated-file docs | 44F8CC48 (low) | **FOLD** into G1's shipment as a trailing docs task — same session, distinct file, append-only surface honoured. |

Note G1 and G6 are different artifact classes (Go tooling vs a gated config
file). They are placed in one shipment only because 44F8CC48 is a single
comment line with no code coupling, and splitting a one-line comment fix into
its own shipment would cost more ceremony than it saves. This is recorded as a
deliberate, bounded exception rather than a silent mixing.

### 1.5 Deferral rationales (explicit, per entry)

* **D10D3AFC + 9F824B64 (G2)** — deferred, **not** rejected. Both are
  high-signal and D10D3AFC is the *third* consecutive closure to disclose the
  same category of gap (035-S computed no binding; 036-S computed but never
  revalidated; 037-S skipped the lock). But D10D3AFC names two mutually
  exclusive candidate resolutions ((a) a real `safe-close` subcommand vs (b) a
  Ship-side hard gate), and 9F824B64 requires reconciling the
  `shipment-reconcile` SKILL.md single-writer-lock constraint against
  `concurrency.instructions.md`, which says lock *only* under concurrent access
  — a genuine contract conflict. Choosing between (a) and (b) and resolving that
  conflict is a harness-architecture decision that must not be made as a
  by-product of a gate-engine re-plan session. **Recommended next action:** a
  dedicated Stage session with the `deliberate` skill on G2 alone.
* **5A8EC1BC (G3)** — deferred on its **own stated trigger**. The entry itself
  asks whether consolidation is warranted "now (three call sites) or deferred
  until a fourth engine port makes the duplication cost clearly outweigh
  consolidation risk." There is no fourth engine and no fourth port is planned,
  so the trigger has **not fired**. Consolidating three correct, independently
  tested call sites into a shared abstraction now would mutate three shipped
  engines for a maintainability gain with no forcing function. Revisit when a
  fourth engine lands or when a cross-cutting git-isolation change is actually
  required. The exported-mutable-var half (`Denylist`/`Selectors`) is real but
  is defence-in-depth with no importer outside the module.
* **978D2946 (G5)** — deferred as standalone. Genuinely worth doing (two
  consecutive migration units hit a Windows line-ending hazard caught only by
  manual `bash -n`), but it is new CI/hook infrastructure with no coupling to
  any gate-engine re-plan, and folding it in would mix a CI-infrastructure
  artifact class into a Go-engine shipment. Should become its own small unit.

---

## 2. Re-plan — 030-S / 033-F "Unify retired-architecture scan scope declaration"

### 2.1 What the port changed

The three tasks targeted `scripts/lib/retired_arch.py`. The equivalent Go code is
`tools/gatecheck/internal/retiredarch/`.

| Old task | Verdict | Evidence |
|---|---|---|
| 033.001-T — introduce one `SCAN_SCOPE` declaration | **NEEDED**, retarget | No single scope declaration exists. Scope arms live in `shouldScanRepoPath` (`select.go:50`): `config.toml.example` (`:51`), `cmd/` (`:54`), `internal/` (`:57`), internal exclusions `/testdata/` + `_test.go` (`:60`). |
| 033.002-T — derive ls-files vector *and* predicate from it | **NEEDED**, retarget | Still duplicated: the pathspec vector is a separate literal list in `selectRepoPaths` → `git(root, "config.toml.example", "cmd/**", "internal/**")` (`select.go:125-126`), restating what `shouldScanRepoPath` restates again. |
| 033.003-T — keep the pin source-text-anchored | **SATISFIED by the port** | The Python `inspect.getsource()` pin became a **stronger** AST pin: `checkPathspecPin` (`pin.go:56`) parses the on-disk `select.go` with `go/parser`, collects `*ast.BasicLit` string literals via `collectStringLits` (`pin.go:102`), and asserts `pathspecPinLiterals` (`:33`) occur in `selectRepoPaths` and `prefixPinLiterals` (`:37`) in `shouldScanRepoPath`. `SelectionPathspecPin` (`:156-162`) reads the real repo file. The cmd-includes-tests / internal-excludes-tests asymmetry is already asserted in `selftest_selection.go:187-198` and `:256-281`. |

**The critical nuance survived the port intact**: `cmd/` still accepts any `.go`
including tests, while `internal/` still excludes `_test.go` and `/testdata/`.
The naive-flat-prefix-unification hazard that 033.001-T was written to guard
against is therefore **still real** in Go.

### 2.2 Decisions

* **D-030-1 — Keep the unit; retarget both live tasks to Go.** The maintenance
  hazard (scope stated in three places) is unchanged by the port. Confirmed
  live.
* **D-030-2 — Drop 033.003-T as an implementation task; demote it to acceptance
  criteria on 033.002-T.** `pin.go` already does source-text anchoring *better*
  than the Python original. Re-implementing it would be green-on-arrival work.
  But the pin must not be allowed to *regress* during the 033.001/033.002
  refactor. **Correction (plan-review round 1, finding 3):** the risk is **not**
  that the pin "silently stops pinning". Moving the literals makes
  `collectStringLits` find nothing, `containsAll` return **false**, and the pin
  **FAIL loudly** — `select.go:5-9` documents this fail-closed behaviour as
  deliberate. The real, unguarded risk runs the other way: `containsAll(_, [])`
  returns **`true`** (`pin.go:127`), so *emptying* `pathspecPinLiterals` or
  `prefixPinLiterals` to make the refactor land green yields a **vacuously green
  pin that asserts nothing**. That is the failure mode the carried ACs must
  prevent (plan AC-A2.2 non-vacuity guard, AC-A2.3 committed negative control).
  The risk is carried as explicit ACs rather than as a task.
* **D-030-3 — C312BD4C (G4) folds in here, as a decision, not as code.**
  `retiredarch` is the heaviest consumer of `gomask`. But the masker semantic
  change must **still not** ride on this unit: 033-F is a behaviour-preserving
  refactor of scope *declaration*, and changing what the masker unmasks would
  change retired-arch verdicts — exactly the coupling D-4 of the migration
  deliberation forbade. **C312BD4C stays deferred**, and this deliberation
  records the second half of its disposition: the contradiction is now between
  the Go comment (`gomask.go:138-141`) and the Go pinned test
  (`gomask_test.go:57`), so the **cheap, correct, zero-risk resolution is to fix
  the comment, not the regex.** The comment is wrong; the behaviour is
  deliberately pinned. That reduces C312BD4C from "semantic change to a
  merge-blocking gate masker" to "correct a stale comment" — a very different
  risk profile. Recorded as a recommended re-scope for a future session.
* **D-030-4 — Harvest DD0BB60F as A-T3, split into A-T3a (033.004-T),
  A-T3b (033.005-T) and A-T3c (033.006-T). This is a deliberate
  design change to D-6 / INV-6 (operator decision, 2026-10-01 follow-up).**
  * **Amended source.** D-6 / INV-6 of
    `docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`
    (INV-6 at :969, "pathspec pin stays source-text-anchored"; M2-T8 at :516;
    R-6 at :55) defined the pin as **presence**. Each pin literal had to
    appear as a `*ast.BasicLit` somewhere inside the
    `selectRepoPaths`/`shouldScanRepoPath` body. That plan is not edited; this
    record and plan rev 5 carry the amendment.
  * **What is preserved.**
    * Source-text anchoring: the pin parses the on-disk `select.go`.
    * H-11 independence: `pathspecPinLiterals`/`prefixPinLiterals` are never
      derived from `select.go`, and `pin.go:9-15` is unchanged.
    * No `go/types`, no execution of `select.go`.
    * Containment semantics: expected ⊆ resolved, so widening the scope is
      still allowed (INV-2).
  * **What changes.** A literal counts **only when it is bound to the
    consuming call**:
    * The pathspec half resolves from the arguments of the single `GitRunner`
      call in `selectRepoPaths`.
    * The prefix half resolves from the prefix arguments of `strings.HasPrefix`
      in `shouldScanRepoPath`.
    * Each argument must be reached through **one documented canonical
      syntactic path** to the corresponding `scanScope` field, enforced as a
      file-wide **use-whitelist** (plan A-T3). Direct literal arguments are
      not accepted, because AC-A1.2 forbids restating them (rev-5 review
      round 2).
    * Dead literals never count.
    * Zero or several git calls, an unrecognised form, an empty resolved set,
      or an empty expected set each **fail closed**.
  * **Rationale.**
    1. **It closes a real fail-open.** A narrowed call plus a retained dead
       literal keeps today's pin green, and this weakens D-030-2's "pin
       supersedes 033.003-T" reasoning (§4.4).
    2. **ALP-1 widens the gap.** Once the literals sit in a package-level
       `scanScope`, "the literal exists" says even less about "the literal
       reaches git".
    3. **Python parity is no longer binding.** M4 retired the Python engine, so
       the presence definition was only a parity artefact.
    4. **The change is fail-closed only** (INV-2). It can only reject more.
  * **Sequencing.** The chain is `033.004-T → 033.002-T`,
    `033.005-T → 033.004-T` and `033.006-T → 033.005-T` (all `blocks`). Each
    task lands as its own commit after ALP-1. ALP-1 is **not** split, and
    INV-5 still declares exactly one pair. Each task's red is a newly authored
    assertion inside that task, so INV-5's test-first exclusion applies.
  * **Why it is split.**
    * Rev-5 plan-review round 1 found that a single A-T3 implied 7 test
      scenarios and about 5 functions, which breaches the 2-hour rule.
    * Round 2 found that a two-way split still left the pathspec task at 5
      production functions once the shared use-whitelist checker was counted.

    The operator's "a new task" is therefore delivered as **three** serial
    tasks:
    * A-T3a (033.004-T) covers `containsAll`'s non-vacuous empty-input
      contract;
    * A-T3b (033.005-T) binds the pathspec half;
    * A-T3c (033.006-T) binds the prefix half.

    A-T1 gains AC-A1.8, which pins a bindable canonical shape.
  * **Acceptance criteria.**
    * The `containsAll(_, [])` vacuously-green hazard (`pin.go:127`) is
      AC-A3a.1, a test-first pure-predicate test.
    * Committed reject tables have a "narrowed call plus a retained dead
      literal" row: AC-A3b.1 row (a) for the pathspec half, and AC-A3c.1
      row (b) for the prefix analogue. Each row is RED against the prior pin,
      or is declared green on arrival under the plan's red-phase rule.
  * **Stop condition.** If binding is not syntactically establishable for
    ALP-1's landed shape, or a task exceeds its budget, the task HALTs to
    Stage. It never falls back to presence.
  * **Re-gate.** Unit A goes through plan-review again (plan rev 5, "Plan
    Review — Revision 5 (Unit A)").

### 2.3 Re-planned task set (2 live tasks, was 3; rev 5 adds 3)

* **033.001-T (rewritten)** — Introduce a single `scanScope` declaration in
  `select.go` describing, per arm, its pathspec form and its test/testdata
  policy. Size S, complexity medium.
* **033.002-T (rewritten)** — Derive both the `git ls-files` argument vector in
  `selectRepoPaths` and the `shouldScanRepoPath` predicate from `scanScope`;
  neither restates a literal. Update `pin.go`'s literal targets in lockstep.
  Size M, complexity medium.
* **033.003-T — DROPPED** (satisfied by `pin.go`; its intent survives as ACs on
  033.002-T). **Rev 5:** closure as a verified no-op inside 030-S is authorised
  (D-000-2, §4.7).
* **033.004-T — NEW (rev 5, D-030-4)** — A-T3a: make `containsAll`
  non-vacuous on empty input. Size XS, complexity low, test-first. Depends on
  033.002-T.
* **033.005-T — NEW (rev 5, D-030-4)** — A-T3b: bind the pathspec half to the
  git call. Size M, complexity medium, test-first. Depends on 033.004-T.
* **033.006-T — NEW (rev 5, D-030-4)** — A-T3c: bind the prefix half to the
  `strings.HasPrefix` call. Size S, complexity medium, test-first. Depends on
  033.005-T.

---

## 3. Re-plan — 032-S / 035-F "Repair unignore-regression checker ref resolution"

### 3.1 What the port changed — and it got *worse*

The Python equivalent of `root_gitignore_text_at` is `rootGitignoreTextAt`
(`unignore/git.go:84`).

The original finding said the Python did **not** fail open: a `SystemExit` guard
fired before the use site, making the defect latent defence-in-depth. **That is
still structurally true in Go** — `runDifferentialCheck` runs
`git diff --name-only -z baseRef headRef` and fails closed on an invalid ref
(`checks.go:90-97`) before the use sites at `checks.go:124` and `:128`.

But the Go function itself is **explicitly fail-open**: for a non-HEAD ref it
runs `git show <ref>:.gitignore` (`git.go:99`) and returns `"", nil` on **any**
non-zero exit (`git.go:100-104`). An invalid ref and a legitimately absent
`.gitignore` are **collapsed into the same empty-baseline result**, with no way
for a caller to distinguish them.

Worse for the re-plan: **`git_test.go:103-121` currently asserts the degraded
behaviour as correct**, and its comment names 035-F as the defect. So the test
must be *inverted*, not merely added to — a detail the Python-era plan could not
have known.

| Old task | Verdict | Evidence |
|---|---|---|
| 035.001-T — fail closed at point of use | **NEEDED**, retarget | `rootGitignoreTextAt` (`git.go:84-104`) returns `"", nil` on any `git show` failure. Guard/use-site separation preserved at `checks.go:90-97` vs `:124`/`:128`. |
| 035.002-T — two direct function-level tests | **NEEDED**, retarget + **expanded** | `git_test.go:35-43` covers HEAD/absent→`""`. `git_test.go:58-101` covers valid ref/present. **Missing**: valid *non-HEAD* ref with absent `.gitignore` → `""`, and invalid ref → distinct error. The existing `git_test.go:103-121` asserts the *wrong* behaviour and must be inverted. |
| 035.003-T — correct stale comments | **NEEDED**, retarget | `git.go:75-83` and `git.go:101-103` both still describe the degrade-to-empty behaviour; `git_test.go:103-107` characterises it as the defect. All three need correcting *after* the behaviour change. |

### 3.2 Decisions

* **D-032-1 — Keep all three tasks; retarget to Go.** All three are live. This
  is the **cleanest** of the three re-plans: no design fork, direct
  identifier-for-identifier translation.
* **D-032-2 — The distinction that must be preserved is valid-ref/absent-file vs
  invalid-ref.** This is the whole point of the unit and the Go code currently
  conflates them. `rootGitignoreTextAt` must return `("", nil)` for a valid ref
  whose `.gitignore` is absent, and a **distinct non-nil error** for an
  unresolvable ref. Distinguishing these requires inspecting `git show`'s
  *stderr/exit-code shape*, not just its exit status — which is why the task is
  S/medium and not XS.
* **D-032-3 — 035.002-T is now partly a test *correction*, not only a test
  addition.** Its ACs must state that `git_test.go:103-121` is inverted and must
  FAIL before the change and PASS after. Without this the unit could be reported
  complete while the old assertion still pins fail-open behaviour.

### 3.3 Re-planned task set (3 tasks, unchanged count)

All three rewritten onto `tools/gatecheck/internal/unignore/git.go`. See the plan
document for full text.

---

## 4. Re-plan — 031-S / 034-F "Harden write-path gate" — **decided: SPLIT (D-031-2, §4.5)**

### 4.1 What the port changed

| Old task | Verdict |
|---|---|
| 034.001-T — verify the spike doc | **SATISFIED** — the spike at `docs/decisions/2026-09-19-…-spike.md` exists and records D-1..D-4. |
| 034.002-T — call-extent extraction | **NEEDED** — `scanText` (`writepath.go:112-119`) is still line-based over `pysem.SplitLines`; `findSelector` (`:89-101`) is `strings.Index` + boundary checks. No extent, no args. |
| 034.003-T — access-mode allowance predicate | **NEEDED** — no allowance exists. The target call at `internal/pathsafe/reparse_windows.go:164-171` **still has exactly 7 args with arg index 1 == literal `0`**, so D-2's predicate still applies verbatim. |
| 034.004-T — four new selectors | **NEEDED** — `Selectors` (`writepath.go:32-37`) has the original 20; `syscall.CreateFile`, `syscall.Write`, `os.OpenRoot`, `os.Root` are all absent. |
| 034.005-T — Root.Resolve tripwire | **NEEDED** — no `pathsafe.NewRoot` receiver tracking anywhere. |
| 034.006-T — named import alias resolution | **NEEDED** — textual selector matching cannot resolve aliases. |
| 034.007-T — document residuals | **NEEDED** — residuals undocumented. |
| 034.008-T — fixtures + positive control | **NEEDED** — current fixtures cover only existing selector/mask vectors. |

### 4.2 The fork

`writepath` does **not** parse Go. It is text-and-mask: it calls
`gomask.MaskGoNonCode` (`writepath.go:24`, `:127-134`) and scans the masked text
line by line. The original spike's D-1/D-2 were designed in that world.

Meanwhile `retiredarch/pin.go:56` and `scango.go` **already use `go/parser` and
`go/ast`** in this very module.

**Option A — faithful port of D-1/D-2 onto masked text.**
Implement paren-depth call-extent walking over the masked string exactly as D-1
specifies, recovering line numbers by counting newlines before the match offset.

* *For:* The spike's feasibility proof still holds — it was validated against the
  real masker (now `gomask.MaskGoNonCode`) and the real
  `reparse_windows.go` call, both of which still exist unchanged. D-1 is
  described in language-agnostic terms ("offset-based paren-depth extraction over
  the masked text") and ports directly. Lowest risk; no verdict-semantics change
  beyond the authorised deltas. 034.001-T's verification gate stays meaningful.
* *Against:* 034.006-T (import aliases) requires hand-rolling import-block
  parsing that `go/ast` gives for free and exactly. 034.005-T (receiver bound
  from `pathsafe.NewRoot`) is close to undecidable textually. The gate keeps
  accreting hand-rolled parsing that the stdlib already does correctly.

**Option B — move the write-path scanner to `go/ast`.**
Replace mask-and-scan with a real parse: `*ast.CallExpr` gives extent and args
exactly, `*ast.BasicLit` gives the literal `0`, `*ast.ImportSpec` gives aliases,
and `os.Root` type-use becomes a `*ast.SelectorExpr` in type position.

* *For:* Makes 034.002, 034.003, 034.005 and 034.006 dramatically simpler **and
  more correct**. Eliminates the undecidable-call-extent residual that 034.007-T
  exists to document. In-repo precedent (`pin.go`, `scango.go`). Zero new
  dependency — `go/ast` is stdlib, so the original "no Go parser and no new
  dependency" constraint evaporates in Go (it meant "don't shell out to a Go
  parser from Python").
* *Against:* This is a **scanner-engine replacement**, not a hardening increment.
  It changes the gate's failure modes wholesale: an unparseable file must become
  a fail-closed finding rather than a skip, and comment/string handling moves
  from the masker to the parser, which could shift verdicts on the existing
  17-file corpus. It **invalidates the spike's D-1/D-2 mechanism decisions**
  (they describe masked-text offset walking), so 034.001-T's verification gate
  becomes meaningless and a **fresh Go-era spike would be required** before
  implementation. `gomask` cannot be deleted either way — `retiredarch` still
  depends on it.

### 4.3 Recommendation — and why this is not Stage's call alone

**Stage recommends Option A**, on three grounds: the existing spike remains valid
and its feasibility evidence is still anchored to live code; Option A is a
hardening increment to a merge-blocking gate rather than an engine swap; and it
is the only option that can proceed **without a new spike**, so 031-S can be
claimed sooner.

However, Option B is genuinely stronger on the merits for 034.005-T and
034.006-T, and a reasonable operator may prefer paying a one-time spike cost to
stop hand-rolling a Go parser inside a security gate.

Per the Stage behavioural constraint "never promote to plan without the
operator's explicit confirmation of the deliberation outcome", and because this
fork changes the shape of **six of eight tasks**, Stage does **not** decide it
unilaterally.

> **031-S remains ON HOLD pending operator selection of Option A or Option B.**
> 030-S and 032-S are unaffected by this fork and proceed now.

### 4.4 Recommended fold-ins for 031-S (operator decision, out of this session's 8-entry scope)

The duplicate scan surfaced three adjacent entries that belong with 031-S
whichever option is chosen. They are **flagged, not harvested** — they are
outside the eight entries this session was scoped to:

* **8E9F8E55** (low) — add `reject-*.go` fixtures for the pre-existing
  `os.MkdirTemp`, `os.Chown`, `os.Lchown` selectors. Natural companion to
  034.008-T.
* **B72E9715** (low) — add `io.CopyN` and `io.CopyBuffer` to the selector list.
  Natural companion to 034.004-T.
* **DD0BB60F** (low) — `pin.go:76`'s AST pathspec pin only checks that pin
  literals appear *somewhere* in the function body, not that they are the actual
  arguments of the git call; a narrowed selector plus a retained-but-dead literal
  would keep the pin green. **This directly weakens the D-030-2 reasoning above**
  and is a natural companion to the re-planned 033.002-T in **030-S**.

### 4.5 D-031-2 — Operator decision: SPLIT 031-S (2026-10-01, explicit, confirmed)

**Decision.** The operator chose neither pure option. 031-S is **split**:

* **031-S stays the Option A hardening increment** — the masked-text work whose
  value carries over to Option B: D-1 call extents (034.002-T), the D-2
  access-mode predicate targeting `internal/pathsafe/reparse_windows.go:164-171`
  (034.003-T), four new selectors **plus B72E9715's `io.CopyN`/`io.CopyBuffer`**
  (034.004-T, 20 → 26), honest residual documentation (034.007-T) and fixtures
  plus a positive control (034.008-T). **8E9F8E55** is harvested into a new
  sibling 034.009-T, split out of 034.008-T under the 2-hour rule. The 031-S
  fixtures are explicitly the **verdict-parity regression corpus** for the
  later `go/ast` migration.
* **Deferred to a new follow-up feature 049-F / shipment 039-S** — the parts
  that are hard as text scanning and that Option B replaces: named import alias
  resolution (034.006-T → 049.004-T) and the `Root.Resolve`/`NewRoot` tripwire
  (034.005-T → 049.005-T), preceded by a Go-era `go/ast` spike (049.001-T:
  feasibility, fail-closed handling of unparseable files, verdict parity against
  the existing corpus plus the 031-S fixtures) and the engine migration itself
  (049.002-T, 049.003-T), then a docs close-out (049.006-T). **`gomask` is not
  deleted** — `retiredarch` depends on it.
* **034.001-T is SATISFIED** by the port (same disposition as 033.003-T under
  D-030-2): retitled, its edge from 034.002-T removed, left `queued` in the
  manifest. Lifecycle constraint: `.backlogit/hooks.yaml` allows only
  `queued → active|blocked`. There is no `queued → archived|rejected`
  transition, so Stage cannot give the item a terminal state.

  **Named prerequisite for 031-S closure** (plan D-T0): before Ship's closure
  classification, the operator does one of two things.
  * Authorise Ship to close it as an explicit no-op. This is recorded as a
    **disclosed deviation** from the two-set gate in the closure artifact.
  * Or give it, and 033.003-T, a terminal state.
* **DD0BB60F was NOT approved** in this decision and is left untouched in the
  stash (still a 030-S fold-in candidate awaiting a separate operator call).
  *Update (2026-10-01 follow-up): the operator approved it. It is harvested as
  033.004-T, 033.005-T and 033.006-T under D-030-4 (§2.2) and archived.*

**Why this is sound.** It lands every Option-A unit whose work survives the
`go/ast` migration. D-1/D-2 become the migration's behavioural spec, the new
selectors carry over unchanged, and the fixtures become its parity contract.
It defers exactly the two units that are disproportionately hard in a text
scanner: alias resolution needs import-block parsing, and receiver tracking
needs type/flow information. Neither has a live instance at `9b299c8`: zero
aliased `os`/`io`/`syscall` imports in scope, and zero production `.Resolve(`
callers.

**Accepted residual (made explicit).** Between 031-S ship and 039-S ship,
aliased imports evade the gate and the first `Root.Resolve` caller lands
unmonitored. These are recorded as **KNOWN OPEN** residuals in the gate's own
header (034.007-T AC) with a residual-risk statement, and the plan's H-5
assesses likelihood and impact and gives a per-session re-measurement command.

**Backlog mechanics (correcting the earlier memory note).**
`backlogit shipment return-blocked` **is** a supported member-removal path. It
removes the item from the manifest and sets it to `blocked`. Prior art restores
`queued` with `move --status queued`. The two deferred tasks were removed that
way and then re-parented with `backlogit adopt`, which **renumbers** them. ID
map (the rollback record): `034.006-T → 049.004-T`, `034.005-T → 049.005-T`.
Nothing was deleted.

The split also re-points **038-F**, whose trigger was "the `Root.Resolve`
tripwire delivered by 034.005-T". Edge 038-F → 049-F was added, keeping the
038-F → 034-F edge as historical record. A marked description section records
the re-identification, so 038-F cannot read as unblocked once 031-S ships. The
full harvest read-back is in the plan's `## Harvest Record — Revision 4`.

**Gating.** 039-S `blocks`-depends on 031-S, and 049-F depends on 034-F, so the
follow-up cannot be claimed first. 049-F also carries a **STAGE HOLD** until
049.001-T's spike findings exist and the re-planned E-T2..E-T6 pass plan-review.

**Hold disposition (conditional).** The 031-S hold is lifted **only if**
plan-review of the revised 031-S unit and the follow-up unit returns ADVISORY or
better with no P0/P1. The outcome is recorded in the plan's
`## Plan Review — Revision 4` and in the D-031-2 row of §5.

#### 4.5.1 Triage of the two harvested stash entries (P-021 C5/C6)

| Entry | Duplicate scan (C5, unconditional) | Late-identifier reconciliation (C6) | Harvested into |
|---|---|---|---|
| **8E9F8E55** | **CLEAN** over all 24 active stash entries. Adjacent 8E18CCF5 and 56B16321 are distinct expansions and were not merged. | `PR=N/A` → **recovered PR #71**, joined on the entry ID from Ship's records `docs/archive/memory/2026-09-20-ship-027-s-pr-ready-awaiting-merge-approval.md` (cites both entry IDs; the PR was created immediately after) and `docs/closure/027-S-030-F-post-merge-closure.md` (`pr: 71`). `review-thread=N/A` **stands** as a truthful terminal record: a threadless local review, no late thread found. | 034.009-T (031-S). The stash text also lists `os.Chtimes`, so it is included. |
| **B72E9715** | **CLEAN** (same scan). | Same as above: **PR #71 recovered**; `review-thread=N/A` stands. | 034.004-T (031-S). |

Both entries were annotated in place with the recovered PR and their promotion
targets, and then **archived** (`backlogit stash archive`, non-destructive)
after harvest.

### 4.6 D-031-3 — Go-era amendments to D-1/D-2 (Stage, from rev-4 plan-review round 1)

The rev-4 plan-review round 1 returned **FAIL**: 1 P0 and 6 P1s. Six personas
were dispatched; the Security Lens, Go, Constitution, Architecture and
Learnings personas raised the blocking findings. The D-1/D-2 mechanism
decisions were taken in the Python era, on 2026-09-19. Three defects show they
no longer hold for the Go engine, and the P1s carry the rest of the
remediation:

1. **P0 — paren-only depth is bypassable.** Commas inside braces, or inside
   tag-shaped raw strings (which the masker leaves visible), sit at paren
   depth 1. They can fake a 7-argument list with `0` in slot 1 while the real
   call writes, e.g. `syscall.CreateFile(T{p, 0, …}.Args())`.
2. **P1 — access `0` is not "metadata-only".** `CreateFile` with access `0`
   plus `CREATE_NEW`, `CREATE_ALWAYS` or `OPEN_ALWAYS` still creates a file, and
   `FILE_FLAG_DELETE_ON_CLOSE` deletes one.
3. **P1 — "scanText stops iterating lines" breaks a Go-era architectural pin.**
   `retiredarch/writepath_mask_test.go` requires `scanText` to range over
   `pysem.SplitLines` at top level.

**Decision (Stage, within planning authority; flagged to the operator).**

* **D-1′** — the extent algorithm is kept. In addition:
  * the `SplitLines` line loop is **retained** and gains a byte cursor;
  * depth is tracked over `()[]{}`, and a mismatched closer is undecidable;
  * every occurrence on a line is enumerated.
* **D-2′** — D-2's four conditions, **plus**:
  * a bare-operand rule: the extent may contain no inner brackets, braces or
    parens, and no quotes or backquotes;
  * `Args[4]` must be exactly `syscall.OPEN_EXISTING`;
  * `Args[5]` must use only the allowlist `FILE_FLAG_BACKUP_SEMANTICS` /
    `FILE_FLAG_OPEN_REPARSE_POINT` (narrowed in round 2, see below);
  * the predicate is evaluated per occurrence, with the once-per-line
    de-duplication applied **after** it.
* **Task splits under the 2-hour rule.** The frozen differential oracle moves
  out of 034.002-T into new **034.010-T**. The new-selector presence fixtures
  move out of 034.008-T into new **034.011-T**.

Both amendments only **narrow** the allowance or preserve parity. Neither
widens what the gate admits. The live `internal/pathsafe/reparse_windows.go:164`
call still passes every D-2′ condition: access `0`, `OPEN_EXISTING`,
`FILE_FLAG_BACKUP_SEMANTICS`, bare operands.

The operator's D-031-2 scope is unchanged. Unit E's G-4 now requires the
`go/ast` port of D-2′ to be a representation change only.

**Round-2 narrowings (rev-4 plan-review round 2).** Round 2 found one P1:
D-T1a's oracle could not survive Unit E unedited. Stage made these further
narrowings:

* **Rule 3b (new).** Every segment must be non-empty. `gomask` blanks
  interpreted strings and runes, delimiters included.
* **Rule 7 (narrowed).** The flags argument must be **exactly**
  `syscall.FILE_FLAG_BACKUP_SEMANTICS`, with no `|` combinations.
  `FILE_FLAG_OPEN_REPARSE_POINT` is dropped because no caller uses it.
* **Oracle lifecycle.** The oracle's legacy side is frozen. It has exactly one
  authorised adaptation, in E-T2 (plan AC-E2.6), and that adaptation can only
  make it stricter.
* **Residual item 7.** A selector split by a newline or comment is now recorded
  as a known-open fail-open surface.

**Operator acknowledgment of D-031-3: ACKNOWLEDGED (2026-10-01, D-031-4, §4.7).**
D-031-3 is a Stage amendment made under planning authority and gated by the
operator's D-031-2 hold-lift condition. The operator has now explicitly
acknowledged it as the Go-engine adaptation of the call-extent (D-1) and
access-mode (D-2) rules.

### 4.7 Operator decisions on the rev-4 open items (2026-10-01 follow-up, explicit, final)

These were recorded by Stage in the follow-up session on branch
`chore/stage-post-m4-followups-and-go-replan`, as routed by the Orchestrator
under P-013.5.

* **D-031-4 — D-031-3 ACKNOWLEDGED.** The operator acknowledges D-1′/D-2′ as
  the adaptation of the call-extent (D-1) and access-mode (D-2) rules to the Go
  engine. Nothing is left pending on 031-S's D-031-3 dependency.
* **D-031-5 — H-5 re-measure is POLICY.** It replaces a CI tripwire. Every
  Stage session re-measures the known residuals until 039-S ships: aliased or
  dot write-capable imports, production `Root.Resolve`/`pathsafe.NewRoot`
  callers, and item-6 primitives. The three plan H-5 commands are used, and
  their verbatim output goes in the session memory. Any new hit promotes 039-S
  to the head of the queue. This closes R4b-P2-5: declining the CI tripwire
  now rests on an approved control. The first run under the policy is recorded
  in plan H-5 (rev 5) and in the session memory: no new trigger.
* **D-031-6 — `os.Chtimes` CONFIRMED in 034.009-T.** All four D-T6 fixtures
  stay in scope.
* **D-000-2 — Verified no-op closure AUTHORISED for 034.001-T (031-S) and
  033.003-T (030-S).** Ship closes each **inside its own shipment** through the
  normal path `queued → active → done`. Evidence at the pre-task-completion gate
  (`evidence_required: true`) shows the work is already satisfied:
  * for 034.001-T, the spike document
    `docs/decisions/2026-09-19-intercom-go-call-extent-extraction-and-allowance-predicate-spike.md`
    plus D-031-3, acknowledged as D-031-4;
  * for 033.003-T, `pin.go` supersedes it. Cite the symbols
    `checkPathspecPin` and `SelectionPathspecPin` plus the SHA at closure.
    Close it after 033.006-T lands, or, if any A-T3 task halts to Stage, cite the
    post-ALP-1 pin instead.

  For both items, Ship records the `harness-ready` precondition (P-002/P-004)
  at claim as a disclosed `skip_policy: P-002`, scoped to that item and citing
  this decision. Stage's reading needs operator confirmation, which is
  recorded as an open item.

  Each closure artifact counts the item as **satisfied-by-prior-work**, not as
  delivered scope, and discloses that. `.backlogit/hooks.yaml` is **not**
  changed. This resolves the D-031-2 "named prerequisite" for both 030-S and
  031-S closure.
* **D-049-1 — Go-era `go/ast` spike SCHEDULED.** E-T1 / 049.001-T's spike
  (G-1..G-7) runs **after 031-S merges and before 039-S is claimed**, because
  G-3 parity needs the Unit D corpus, ending at 034.009-T, on `main`.
  * **Execution:** Stage runs it under the P-016 time-boxed spike-worktree
    exception.
  * **Trigger:** stash entry **1EEBECA5** (kind spike, priority high), which the
    next Stage triage after 031-S closes picks up. If triaged earlier, it is
    deferred with the reason "trigger not fired" and archived only once the
    spike document exists.
  * **Backstop:** the `blocked_stale` hook on blocked 049-F (7 days,
    Stage-subscribed).
  * **Why this mechanism:** a stash entry is backlogit's native Stage-intake
    surface and survives across sessions. It is the only Stage-owned queue the
    next cycle is required to read (Step 1). A checkpoint was rejected because
    resolving it is session-scoped and recovery-oriented, not scheduling.
  * **Where recorded:** 049-F, 049.001-T and plan E-T1.
  * **Operator rationale (accepted Q&A on why `go/ast`).** The write-path gate
    is CI tooling that guards intercom-go's own source, not runtime code.
    `go/ast` gives:
    * exact call boundaries;
    * import-alias resolution;
    * `NewRoot`/`Root.Resolve` receiver tracking;
    * immunity to split selectors.

    None of those gaps is exploited today. Their value grows as the Copilot
    SDK client gains real workspace write paths. 039-S therefore stays behind
    the spike and is sequenced after 031-S.
* **D-S-5 — Item-6 selector widening CAPTURED** as stash entry **458F9385**
  (kind task, priority low). It covers:
  * `ioutil.WriteFile/TempFile/TempDir`;
  * `syscall.WriteFile/Open/Unlink/Rename/Mkdir/CreateHardLink/DeleteFile`;
  * the `x/sys/windows` and `x/sys/unix` equivalents.

  It cites plan D-T4 residual item 6 and R4-P2-g. **Stage recommends folding it
  into 039-S / 049-F at the post-spike re-plan**, because `go/ast` import-path
  resolution makes detection exact. Plan residual item 6 now cites 458F9385
  instead of "stash candidate".
* **D-030-4 — DD0BB60F harvested into 030-S as 033.004-T, 033.005-T and 033.006-T.** This is a
  deliberate design change to D-6 / INV-6; see §2.2 for the full record.
* **D-030-5 — Gate outcome (2026-10-01): FAIL, so 030-S goes back on STAGE
  HOLD.**
  * **The review.** The rev-5 Unit A plan-review round 3 was the final allowed
    re-entry. It returned **FAIL**: the Security Lens raised five P1 findings,
    SEC-1..SEC-5. Each is a fail-open narrowing that the canonical binding
    contract still accepts:
    * exclude/magic pathspecs, and prefix shadowing, defeat expected ⊆
      resolved;
    * the self-test has no probe for non-test `cmd/**/*.go`;
    * the `selectRepoPaths` shape is unpinned;
    * the `DefaultGitRunner` body is unpinned;
    * prefix control flow outside the loop is not analysed.
  * **The hold.** Decision 7 lets the hold stay lifted only at ADVISORY or
    better with no P0/P1, so **033-F stays `blocked`**. Both 033-F and 030-S
    now carry a STAGE HOLD section.
  * **The harvest.** The tasks were still harvested, because decision 7 directs
    it, but under the held feature. Each is marked STAGE HOLD, and its
    acceptance criteria are expected to change at remediation.
  * **Escalation.** The protocol resolved to `ESCALATION_DEGRADED`, so this
    goes to operator review.
  * **Options.** (a) Stage remediates in rev 6, then one operator-authorised
    re-review (recommended). (b) LR-6 narrowing: ALP-1 + A-T3a now, with
    A-T3b/A-T3c held. (c) Risk acceptance of SEC-1..SEC-5.
  * **Where recorded.** Plan, `## Plan Review — Revision 5 (Unit A)`.

**P-021 obligations for this follow-up.**
* **(A) Duplicate scan, unconditional.** It ran over the stash for the two new
  captures and for DD0BB60F. It found no duplicate (the `ioutil`, `go/ast`
  spike and 049.001 keyword scan matched only the new entries), so the scan is
  CLEAN. DD0BB60F carries no `DEFERRED SCOPE EXPANSION` marker.
* **(B) Late-identifier reconciliation.** Not triggered for these entries: none
  carries an `N/A` source ref.

### 4.8 Operator decisions on the rev-5 open items (2026-10-01 second follow-up, explicit, final)

* **D-030-6 — OPTION A: Stage remediates in plan rev 6, with ONE
  operator-authorised review round 4.**
  * **The decision.** The operator chose option (a) of D-030-5. Stage revises
    Unit A to rev 6, with a changelog, fixing SEC-1..SEC-5 and the round-3 P2s.
  * **Exception to the cap.** The operator explicitly authorises **one**
    additional plan-review round (round 4) on the revised Unit A. This is an
    operator-authorised exception to the 3-cycle cap. It is not repeatable,
    and **no round 5** may be run.
  * **Conditions.** If round 4 returns ADVISORY or better with no P0/P1, lift
    the hold: 033-F → `queued`, remove the hold notes, and clear do-not-claim.
    If it returns FAIL, keep the hold and report.
  * **Contract surface.** Fixes stay on the retiredarch pin's scan-scope
    integrity. Anything outside it is a P-021 DEFERRED SCOPE EXPANSION entry.
    ALP-1 is not split, and INV-5 is respected.
  * **The rev-6 design.** The use-whitelist is replaced by a **closed-world,
    frozen-declaration pin**:
    * §A-CANON canonical texts in `pin.go`, compared token-for-token with
      `select.go`'s selection declarations (`import`, `GitRunner`,
      `DefaultGitRunner`, `scanArm`, `scanScope`, `shouldScanRepoPath`,
      `selectRepoPaths`);
    * a closed world and confinement for `select.go`;
    * `scopeDataOK` data invariants;
    * package closure.
  * **Finding map.**
    * SEC-1: shared rules, token equality, and `scopeDataOK`'s magic,
      duplicate and shadowing rejection.
    * SEC-2: new A-T4 (`cmd/x/main.go` probe).
    * SEC-3: `selectRepoPaths` frozen.
    * SEC-4: `DefaultGitRunner` frozen.
    * SEC-5: `shouldScanRepoPath` frozen.
  * **New tasks.**
    * **033.007-T** (A-T3d, package closure; S, medium; test-first).
    * **033.008-T** (A-T4; XS, low; test-first).
    * Edges: 033.007-T → 033.006-T, 033.008-T → 033.002-T,
      033.007-T → 033.008-T (all `blocks`).
  * **P-021 captures** (P-021 C6 deliberation required before planning):
    * **DC921AF6** (R-A1): runner wiring outside `select.go`, and
      `scanPath`/`engineForPath` dispatch.
    * **D7BF9F74** (R-A2): the git environment and config, including redirect
      variables, and `pysem` integrity.
* **D-030-7 — Gate outcome (2026-10-01): round 4 ADVISORY, so the 030-S hold
  is LIFTED.**
  * **Verdicts.** Six personas, no P0/P1. Security Lens, Go, Scope,
    Constitution and Learnings returned ADVISORY. The Architecture
    Strategist anchor (gpt-6.1-sol) returned PASS. SEC-1..SEC-7, GO-1..GO-4,
    SC-1..SC-4 and CN-1..CN-3 are all CLOSED.
  * **Amendments.** All P2s (SEC4-1..3, GO4-1..2, SC4-1, CN4-1..2,
    LR4-1..3) and most P3s were applied as text inside the reviewed design.
    No re-review was run.
  * **Backlog.** 033-F is `blocked → queued`. 033-F, 030-S and
    033.004/005/006-T carry HOLD LIFTED sections. Record: plan,
    `## Plan Review — Revision 6 (Unit A)`.
* **D-000-3 — P-002 test-first SKIP AUTHORISED, item-scoped.**
  * **Scope.** The skip covers the verified no-op closures of **034.001-T**
    (031-S) and **033.003-T** (030-S) ONLY. Ship may record a P-002 test-first
    skip for those two items, citing D-000-3, as part of the
    `queued → active → done` no-op closure with gate evidence.
  * **Limits.** It is not a general waiver, and it does not extend to any
    task that adds or changes code. It discharges the CN-2 / AC-D0.4 /
    AC-X3.4 hard precondition that D-000-2 left open.
  * **Where recorded.** Both tasks (`description` and `acceptance-criteria`),
    plan D-T0, "Dropped from Unit A", and the Constitution Check P-002
    deviation entry.
* **D-032-4 — 035-F moved `blocked → queued`.**
  * **Why.** This is consistent with D-032-1 (the 032-S hold was lifted at
    rev 4, but the status stayed `blocked`, a prose-only lift) and with
    034-F's handling.
  * **Path.** A direct `backlogit move 035-F --status queued`, exit 0. No
    force was used.
  * **Result.** 032-S is claimable: `queued`, with 037-S archived.

**P-021 obligations for this second follow-up.**
* **(A) Duplicate scan, unconditional.** It ran over the 25 active stash
  entries for `DC921AF6` and `D7BF9F74`, using the keywords runner,
  `selectRepoPaths`, pathspec, `GIT_`, `pysem`, `scanPath`, `engineForPath`
  and retiredarch. **No duplicate was found**, so the scan is CLEAN. Three
  adjacent entries were checked and found to be different expansions:
  * `990AFA71`: symlink escape in `scanPath`'s lexical join. That is path
    containment, not dispatch integrity.
  * `5A8EC1BC`: consolidation of the three GitRunner shapes. That is
    maintainability, not environment integrity, although a future
    `D7BF9F74` design should consider it.
  * `9FF9EEB4`: fixture-manifest parity.

  None was merged or archived.
* **(B) Late-identifier reconciliation.** Both entries carry `PR N/A` and
  `review-thread N/A`. They are Stage plan-review captures, not
  PR-review-thread captures, and no Ship residual-risk record cites them yet.
  The result is **no late identifier found**, and the `N/A` stands as a
  truthful record.

### 4.9 Post-spike re-plan of 049-F / 039-S (2026-10-02, Stage, dark factory mode)

**Trigger.** Stash `1EEBECA5` (D-049-1) fired when 031-S / 034-F merged (PR
#97, closure PR #98, `main` @ `961b652`). That merge put 034.009-T, and with it
the full Unit D verdict-parity corpus, on `main`.

**Spike.** Stage ran task 049.001-T (plan E-T1) under the P-016 explicit,
time-boxed spike worktree exception. The worktree was a detached
`logs/spike-049-wt` at `961b652`, and the spike code was a throwaway
`zz_spike_test.go`. Stage made no tracked mutation in the worktree and removed
it afterwards. Findings are in
`docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`:

* **G-1** — FEASIBLE: `scanSource(relPath, src)` with
  `parser.ParseComments|parser.SkipObjectResolution`. `gomask` stays the
  single masker for the line map.
* **G-2** — FEASIBLE: any parse error is ED-2 (exit 1), using
  `PositionFor(pos, false)`.
* **G-3** — **exact verdict parity**, with 0 divergences across:
  * the 19 fixtures;
  * the 3 `filebased` inputs;
  * the 17 tracked in-scope files;
  * all 155 module `.go` files.

  The oracle adaptation is the single authorised one: freeze its glob to the
  19 names, and turn the `\f`/`\v`/`U+2028` cases into fail-closed
  assertions.
* **G-4** — a representation change only. The `"syscall"` import-path check
  is strictly stricter.
* **G-5** — additive alias resolution. The only live alias is `copilot`, and
  it is not write-capable.
* **G-6** — syntactic intra-function `NewRoot` → `Resolve` binding.
  `go/types` is rejected.
* **G-7** — every task is within the 2-hour rule. E-T2 is de-risked from
  `high` to `medium`.

**Decisions.**

* **D-049-2 — The `go/ast` migration is FEASIBLE; re-plan, do not abandon.**
  * **Evidence.** The spike shows exact parity. The migration therefore
    removes residual item 7 (split selector) at no verdict cost, and enables
    import-path keying.
  * **Rejected.** Abandoning 049-F, which would keep the masked-text engine
    and residuals 1, 3, 6 and 7 open indefinitely.
* **D-049-3 — Fold `458F9385` (residual item 6) into 049-F as new task E-T7
  (049.007-T).**
  * **Why now.** The rev-5 recommendation (D-S-5) was conditional on G-5.
    G-5 confirms that import-path resolution keys aliased and dot-imported
    `golang.org/x/sys` uses to their canonical path. Spelled-qualifier
    matching (the G-5 additivity rule) still treats any identifier spelled
    `windows` or `unix` as that qualifier. A name collision with an unrelated
    local package or variable therefore stays a **fail-closed false
    positive**: it is reported, never missed, and AC-E7.3 measures zero such
    collisions at `961b652`. The fail-open concern that blocked the widening
    under the masked-text engine is resolved; the collision is not eliminated.
  * **Scope.** 50 selectors appended (`ioutil` 3, `syscall` 14, `windows` 11,
    `unix` 22), 76 in total. The 458F9385 enumeration is extended with the
    remaining namespace / ACL-relevant primitives of the same families
    (`Rmdir`, `MoveFile`, `Truncate`, `Creat`, the `*at` variants,
    `unix.Chmod`; `syscall.Chmod` stays uncovered and is recorded in
    C0D28448).
    These were enumerated from the families named in the stash entry, so this
    widens neither the families nor the gate's contract surface.
  * **Item 6 is narrowed, not closed** (plan-review rev 7 PR7-1 / SB-1). The
    50-name list does not cover every same-family write-capable symbol: for
    example metadata/attribute writes, `syscall.Pwrite`, `unix.Writev`/
    `Pwritev`, `syscall.Ftruncate`/`Link`/`Symlink` and `unix.Mknod`. E-T6
    records the remainder as retained item 6. Enumerating the full surface is
    a separate contract decision, captured per P-021 C2 as `DEFERRED SCOPE
    EXPANSION` stash **C0D28448** (low).
  * **Constraints.**
    * Golden changes are additive only (AC-E7.2).
    * The D-T1a oracle stays bound to the frozen 20.
    * No allowance is added for `windows.CreateFile`.
    * Removing, reordering or renaming an existing selector HALTs (H-2 E).
  * **Rejected.** A separate shipment for the widening. It would need a
    second golden-serialisation chain against the same `writepath.go` with no
    benefit.
* **D-049-4 — Linearise Unit E: E-T2 → E-T3 → E-T4 → E-T5 → E-T7 → E-T6.**
  * **Why.** E-T3 introduces the per-file import table, which E-T4, E-T5 and
    E-T7 extend. Every task edits `writepath.go` and the golden.
    Linearisation gives one table and one serial golden history, and avoids
    a merge race on the `selectors` array.
  * **Edges.** New `blocks` edges are `049.004-T → 049.003-T`,
    `049.007-T → 049.005-T` and `049.006-T → 049.007-T`. The rev-4 edges are
    retained but redundant.
* **D-049-5 — New residual item 8, dynamic proc invocation: disclose and
  defer.**
  * **What it covers.** `syscall.NewLazyDLL` / `LazyProc.Call`,
    `syscall.Syscall*` and the x/sys equivalents can reach any OS write API
    with no write-named selector.
  * **Why it is not in E-T7.** `syscall.NewLazyDLL` is **live** at
    `internal/pathsafe/reparse_windows.go:19`, for the metadata-only
    `GetFinalPathNameByHandleW` probe. Making it a finding would need an
    allowance design, which is a new gate contract, not a widening.
  * **Disposition.** E-T6 records it in both residual locations, and it is
    captured per P-021 C2 as `DEFERRED SCOPE EXPANSION` stash **FE2F02FF**
    (task, low, requires deliberation).
  * **Duplicate scan.** The scan for `LazyDLL`, `NewProc`, `Syscall(` and
    "dynamic proc" across the active stash was **CLEAN**. The nearest entry,
    `0ECC1895`, covers env-mutation selectors in the retiredarch closure,
    which is a different expansion.
* **D-049-6 — The 049.001-T lifecycle: the Stage half is closed, and the task
  stays `queued` for Ship's verification half.**
  * **The plan rule.** E-T1 has two halves:
    * the Stage-executed spike, now complete;
    * a Ship-executed, read-only verification gate (XS / trivial,
      AC-E1.1..AC-E1.3). It checks that the spike doc and the re-plan exist
      before any engine change.
  * **Why it stays queued.** Moving 049.001-T to `done` from Stage would skip
    that gate. It would also be a P-010 overreach: Stage must not close
    shipment work on Ship's behalf. `.backlogit/hooks.yaml` allows no
    `queued → done` transition in any case.
  * **Disposition.** Stage records the spike-complete evidence on the task
    (section `spike`) and leaves it `queued` as the first item of 039-S.
    Ship's claim of 039-S runs the read-only verification, and Ship closes
    the task through the normal `queued → active → done` path.

**H-5 re-measure (policy D-031-5), at `961b652`.**
* The alias / dot-import command is empty. Its positive control matches the
  live `copilot` alias.
* The `NewRoot` / `Resolve` command lists only
  `internal/config/validate.go:32/:42/:67`, with zero `.Resolve(` callers.
* The item-6 command lists only the comment at
  `internal/pathsafe/reparse_windows.go:15`.

There is **no new trigger**.

**P-021 obligations.**
* **(A) Duplicate scan, unconditional.**
  * `458F9385` is not tagged `DEFERRED SCOPE EXPANSION` (it is a Stage
    D-S-5 capture), and no other active entry describes selector widening.
    The scan is **CLEAN**.
  * The residual-8 scan is CLEAN, as noted under D-049-5.
* **(B) Late-identifier reconciliation.**
  * It does not apply to the two entries triaged in this session (1EEBECA5
    and 458F9385), because neither carries an `N/A` source-ref field.
  * The two new captures, FE2F02FF and C0D28448, record `PR N/A` and
    `review-thread N/A` truthfully: they come from plan-review findings, not
    PR threads.
  * (B) runs for them when Stage triages them in a later session.

**Plan-review outcome (rev 7).** It is recorded under
`## Plan Review — Revision 7 (Unit E)` in the plan, and in the D-049-7 row of
the summary table below.

---

## 5. Decision summary

| ID | Decision |
|---|---|
| D-000-1 | All three held shipments' `blocks` edges are satisfied (025-S/027-S/029-S/037-S all archived). Edges **retained** as historical record. |
| D-030-1 | 030-S re-planned onto Go; 2 tasks retargeted. **HOLD LIFTED** on plan-review ADVISORY (round 3, no P1/P2). A-T1/A-T2 declared atomic landing pair ALP-1. *Hold re-imposed by D-030-5 (rev 5).* |
| D-030-2 | 033.003-T dropped — `pin.go` already supersedes it. Intent carried as ACs. Hazard narrative corrected at plan-review round 1: the pin fails **loudly**, the real risk is a **vacuously green** pin from emptied literal lists. |
| D-030-3 | C312BD4C stays deferred, but **re-scoped**: fix the stale comment, not the regex. |
| D-031-1 | 031-S **BLOCKED** — operator must choose Option A (masked-text port, recommended) or Option B (`go/ast` rewrite, needs a fresh spike). **HOLD NOT LIFTED.** *Superseded by D-031-2.* |
| D-031-2 | Operator **SPLIT** 031-S (§4.5). 031-S keeps the Option A increment (034.002/003/004/007/008-T + new 034.009-T; B72E9715 and 8E9F8E55 folded in); 034.001-T SATISFIED. 034.005-T/034.006-T re-parented to new **049-F / 039-S** (`go/ast` spike + migration), gated by `blocks` 039-S → 031-S and a STAGE HOLD on 049-F. Hold lift conditional on rev-4 plan-review — outcome: **round 3 ADVISORY, no P0/P1 → 031-S HOLD LIFTED** (rounds 1–2 FAIL, remediated via D-031-3). 049-F hold retained pending spike re-plan. |
| D-031-3 | Go-era amendments **D-1′/D-2′** from rev-4 plan-review rounds 1–2 (§4.6): `SplitLines` loop retained + fail-closed byte cursor; `()[]{}` depth; per-occurrence predicate; bare-operand rule + rule 3b (non-empty segments); exact `syscall.OPEN_EXISTING`; flags exactly `syscall.FILE_FLAG_BACKUP_SEMANTICS`; frozen oracle with single authorised E-T2 adaptation; residual item 7. Oracle and presence fixtures split into new 034.010-T / 034.011-T. Narrows only; never widens. **Operator acknowledgment: ACKNOWLEDGED (D-031-4).** |
| D-030-4 | DD0BB60F harvested as **033.004-T** (A-T3a, `containsAll` empty-input contract), **033.005-T** (A-T3b, pathspec half) and **033.006-T** (A-T3c, prefix half). The task was split at rev-5 review rounds 1 and 2 to meet the 2-hour rule. The pin binds to its consuming calls through one canonical path per half, enforced as a use-whitelist: a deliberate design change to D-6/INV-6 that keeps source-text anchoring and replaces presence with call binding. A-T1 gains AC-A1.8 (bindable shape). All three tasks are sequenced after ALP-1 and are not part of it. Unit A is re-gated in plan rev 5. |
| D-031-4 | Operator **acknowledged** D-031-3. |
| D-031-5 | H-5 per-session re-measure is **policy**, replacing a CI tripwire. Closes R4b-P2-5. |
| D-031-6 | `os.Chtimes` **confirmed** in 034.009-T. |
| D-000-2 | Verified no-op closure **authorised** for 034.001-T (031-S) and 033.003-T (030-S): `queued → active → done` with gate evidence. `hooks.yaml` unchanged. |
| D-049-1 | Go-era `go/ast` spike **scheduled** after 031-S merges and before 039-S is claimed. Stage-executed (P-016). Trigger is stash 1EEBECA5; backstop is `blocked_stale`. |
| D-030-5 | Rev-5 Unit A plan-review round 3 (final re-entry) returned **FAIL** (Security Lens SEC-1..SEC-5, P1). **030-S back on STAGE HOLD** (033-F `blocked`). A-T3 tasks harvested under the hold per decision 7. Escalation is `ESCALATION_DEGRADED`, so this goes to operator review. Remediation options (a)/(b)/(c) are open. *Resolved by D-030-6/D-030-7.* |
| D-030-6 | Operator chose **OPTION A**. Plan rev 6 adds the closed-world frozen-declaration pin, new **033.007-T** (A-T3d) and **033.008-T** (A-T4), and P-021 stash **DC921AF6** / **D7BF9F74**. **One operator-authorised plan-review round 4**, an exception to the 3-cycle cap; no round 5. |
| D-030-7 | Round 4 returned **ADVISORY**, no P0/P1 (anchor PASS). **030-S HOLD LIFTED**: 033-F `blocked → queued`, hold notes replaced, do-not-claim cleared. P2s applied as text. |
| D-000-3 | **P-002 test-first skip authorised** for the verified no-op closures of **034.001-T** and **033.003-T ONLY**, citing D-000-3. Not a general waiver. Discharges the D-000-2 open precondition. |
| D-032-4 | **035-F `blocked → queued`** by direct `move` (exit 0). **032-S claimable.** |
| D-S-5 | Item-6 selector widening captured as stash **458F9385** (task, low). Recommended for 039-S / 049-F. |
| D-032-1 | 032-S re-planned onto Go; all 3 tasks retargeted. **HOLD LIFTED** on plan-review ADVISORY (round 3, no P1/P2). |
| D-032-2 | Valid-ref/absent-file must stay `("", nil)`; invalid-ref must become a distinct error. |
| D-032-3 | `git_test.go:103-121` must be **inverted**, not merely supplemented. |
| D-S-1 | G1 (7223218F + 50E6F22C + 44F8CC48) harvested into a new gatecheck-containment feature/shipment. |
| D-S-2 | G2 (D10D3AFC + 9F824B64) deferred to a dedicated harness deliberation. |
| D-S-3 | G3 (5A8EC1BC) deferred — its own stated fourth-engine trigger has not fired. |
| D-S-4 | G5 (978D2946) deferred as a standalone CI-hygiene unit. |
| D-049-2 | The `go/ast` spike (049.001-T, Stage half) is **complete**, with **exact verdict parity** (19/3/17/155, 0 divergences). The migration is FEASIBLE: **re-plan, not abandon**. Findings: `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`. |
| D-049-3 | 458F9385 folded in as **E-T7 (049.007-T)**. 50 selectors are appended, import-path keyed, for 76 in total. Golden changes are additive only, and the oracle stays frozen at 20. Item 6 is **narrowed, not closed**; the remainder is deferred as stash **C0D28448**. |
| D-049-4 | Unit E is linearised as E-T2 → E-T3 → E-T4 → E-T5 → E-T7 → E-T6. The new `blocks` edges are 049.004→049.003, 049.007→049.005 and 049.006→049.007. |
| D-049-5 | New residual item 8 (dynamic proc invocation, with a live `NewLazyDLL` at `reparse_windows.go:19`) is disclosed and deferred as stash **FE2F02FF** (low). |
| D-049-6 | 049.001-T stays `queued` for Ship's read-only verification half (AC-E1.1..AC-E1.3). Stage records the spike evidence only, and does not close the task (P-010). |
| D-049-7 | The rev-7 Unit E plan-review returned **FAIL in all 3 cycles**; cycle 3 is the final re-entry (gpt-5.6-sol FAIL, claude-opus-5.5 FAIL, gemini-3.8-flash empty → grok-4.7 FAIL). Each cycle found new P1s in E-T3's test-port contract (SB3-1/GPT3-1, GK3-1). The cycle-3 fixes are applied but **not re-gated**. **The 049-F STAGE HOLD is KEPT (`blocked`)**, and **039-S stays on hold**, so DARK_MODE_HALTED for 039-S. `ESCALATION_DEGRADED` (engram) means operator review. **Operator action required:** authorise a re-gate round (as D-030-6 did) or choose another disposition. The spike result (D-049-2) is unaffected. |

---

*Generated by Copilot*
