---
title: "Deliberation — Post-M4 stash triage and Go re-plan of held shipments 030-S/031-S/032-S"
description: "Dispositions for 8 post-M4 deferred-scope-expansion stash entries, and the retarget of held plan units 6/7/8 from the retired Python gate engines onto the Go engines under tools/gatecheck"
topic: "Post-M4 follow-ups: stash triage + Go re-plan of 030-S (033-F), 031-S (034-F), 032-S (035-F)"
depth: "deep"
decision_status: "partially-decided"
promoted_to: "plan (030-S, 032-S); BLOCKED pending operator decision (031-S)"
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
related_shipments:
  - 030-S (033-F)
  - 031-S (034-F)
  - 032-S (035-F)
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

### 2.3 Re-planned task set (2 tasks, was 3)

* **033.001-T (rewritten)** — Introduce a single `scanScope` declaration in
  `select.go` describing, per arm, its pathspec form and its test/testdata
  policy. Size S, complexity medium.
* **033.002-T (rewritten)** — Derive both the `git ls-files` argument vector in
  `selectRepoPaths` and the `shouldScanRepoPath` predicate from `scanScope`;
  neither restates a literal. Update `pin.go`'s literal targets in lockstep.
  Size M, complexity medium.
* **033.003-T — DROPPED** (satisfied by `pin.go`; its intent survives as ACs on
  033.002-T).

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

## 4. Re-plan — 031-S / 034-F "Harden write-path gate" — **BLOCKED, operator decision required**

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

---

## 5. Decision summary

| ID | Decision |
|---|---|
| D-000-1 | All three held shipments' `blocks` edges are satisfied (025-S/027-S/029-S/037-S all archived). Edges **retained** as historical record. |
| D-030-1 | 030-S re-planned onto Go; 2 tasks retargeted. **HOLD LIFTED** on plan-review ADVISORY (round 3, no P1/P2). A-T1/A-T2 declared atomic landing pair ALP-1. |
| D-030-2 | 033.003-T dropped — `pin.go` already supersedes it. Intent carried as ACs. Hazard narrative corrected at plan-review round 1: the pin fails **loudly**, the real risk is a **vacuously green** pin from emptied literal lists. |
| D-030-3 | C312BD4C stays deferred, but **re-scoped**: fix the stale comment, not the regex. |
| D-031-1 | 031-S **BLOCKED** — operator must choose Option A (masked-text port, recommended) or Option B (`go/ast` rewrite, needs a fresh spike). **HOLD NOT LIFTED.** |
| D-032-1 | 032-S re-planned onto Go; all 3 tasks retargeted. **HOLD LIFTED** on plan-review ADVISORY (round 3, no P1/P2). |
| D-032-2 | Valid-ref/absent-file must stay `("", nil)`; invalid-ref must become a distinct error. |
| D-032-3 | `git_test.go:103-121` must be **inverted**, not merely supplemented. |
| D-S-1 | G1 (7223218F + 50E6F22C + 44F8CC48) harvested into a new gatecheck-containment feature/shipment. |
| D-S-2 | G2 (D10D3AFC + 9F824B64) deferred to a dedicated harness deliberation. |
| D-S-3 | G3 (5A8EC1BC) deferred — its own stated fourth-engine trigger has not fired. |
| D-S-4 | G5 (978D2946) deferred as a standalone CI-hygiene unit. |

---

*Generated by Copilot*
