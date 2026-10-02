---
title: "Plan — Post-M4 Go re-plan (030-S, 031-S, 032-S), gatecheck CLI containment hardening, and the write-path go/ast follow-up"
description: "Retargets held plan units 6, 7 and 8 from the retired Python engines onto tools/gatecheck, plans a new containment-hardening unit from stash entries 7223218F, 50E6F22C and 44F8CC48, and splits unit 7 (031-S) into a masked-text hardening increment plus a deferred go/ast follow-up (049-F / 039-S)"
source_document: docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md
agent: Stage
date: 2026-10-01
stage_branch: chore/stage-post-m4-followups-and-go-replan
bound_snapshot: 9b299c8 (origin/main)
revision: 4
requires_plan_hardening: yes
---

# Plan — Post-M4 Go re-plan and gatecheck CLI containment hardening

**Source document:**
`docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md`

**Scope.** Five units. Units A and B retarget held shipments 030-S and 032-S
onto the Go gate engines. Unit C is a new release unit harvested from stash
group G1 plus G6. Unit D re-plans held shipment 031-S as the **Option A
masked-text hardening increment** (decision D-031-2), folding in stash entries
B72E9715 and 8E9F8E55. Unit E is the **new, deferred `go/ast` follow-up**
(feature 049-F, shipment 039-S) that receives the parts of the original unit 7
that are hard as text scanning and that a `go/ast` engine would replace.

**Out of scope.** Stash groups G2, G3, G5 — deferred with recorded rationale
(deliberation §1.5). Stash entry **DD0BB60F** (fold into 030-S) — **not
approved** by the D-031-2 operator decision and left untouched.

**Requires plan hardening: yes.** See `## Plan Hardening`.

> **Revision 2** incorporates plan-review round 1 (FAIL, 2×P1 + 6×P2 + 5×P3).
> Material corrections: the Unit A "silent pin" hazard narrative was **factually
> inverted** and is rewritten (A-T2); the A-T1/A-T2 split was not independently
> landable under the `unused` linter and is restructured; bare gate invocations
> that always exit 1 are corrected; Unit C is split from 4 to 6 tasks to respect
> the 2-hour rule and width isolation.

> **Revision 3** incorporates plan-review round 2 (FAIL, 1×P1 + 1×P2 + 5×P3) and
> round 3 (ADVISORY, 9×P3). Material corrections: A-T1 was still not
> independently landable — it necessarily reddens the pathspec pin — so **INV-5**
> and the **ALP-1** atomic-landing-pair declaration were added with a bounded
> INV-1 carve-out and a pair-boundary gate (AC-A2.6); the refactor-prohibition
> comments A-T1 supersedes are now owned (AC-A1.7); **C-T2 and C-T3 were
> inverted** so the wrapper test supplies the red before the guard lands.
>
> **Gate status: ADVISORY — proceed with operator awareness.** No P1 or P2
> findings remain. The round-3 P3 residuals are applied in this revision.

> **Revision 4** records operator decision **D-031-2** (deliberation §4.5): 031-S
> is **split**. Units A, B and C are **unchanged** by this revision. Unit D
> (031-S, revised) is the Option A increment — a frozen detection oracle,
> call-extent extraction (D-1′), the access-mode allowance predicate (D-2′), six
> new selectors (four planned plus B72E9715's `io.CopyN`/`io.CopyBuffer`),
> residual documentation, and three fixture tasks (the last harvested from
> 8E9F8E55). D-1′/D-2′ are the Go-era amendments of D-1/D-2 recorded as
> deliberation **D-031-3** after rev-4 review round 1. Unit E (049-F / 039-S,
> new) holds the deferred `go/ast` work — a Stage-executed spike, the engine
> migration, the allowance-predicate port, named import alias resolution
> (034.006-T, re-parented) and the `Root.Resolve`/`NewRoot` tripwire (034.005-T,
> re-parented). Unit E is **dependency-gated on 031-S** and carries its own
> STAGE HOLD pending the spike. Hardening entries for both units are added to
> H-1..H-5 and a new **H-6 Rollback** section. Revision 4 is gated separately
> under `## Plan Review — Revision 4` (rounds 1–2 FAIL, remediated; **round 3
> ADVISORY, no P0/P1 → 031-S STAGE HOLD LIFTED**); the revision-3 verdict above
> covers Units A–C only.

---

## Invariants (all units)

* **INV-1 — No gate verdict changes.** No unit may change the PASS/FAIL verdict
  of any gate on the tracked tree at `9b299c8`, except where an acceptance
  criterion names the change explicitly. Every unit verifies against its own
  parent commit, not a fixed snapshot.
* **INV-2 — Fail-closed direction only.** Where behaviour changes, it may only
  make the gate scan more, reject more, or error louder. No unit may widen an
  allowance or convert an error into a silent pass.
* **INV-3 — No write primitives introduced.** No destructive filesystem write
  primitive may be introduced into `internal/**` or `cmd/**`; doing so is a named
  stop condition (it would trip the very gate being hardened).
* **INV-4 — Lint-clean on every task boundary.** `ci.yml` runs blocking
  `golangci-lint run ./...` (job `lint`, `.golangci.yml` `default: standard`,
  which includes `unused`) and `staticcheck ./...` (job `security`), and the lint
  job formats the whole tree **including `testdata`**. Therefore **every task
  must leave the tree lint-clean and gofmt/goimports-clean on its own** — a task
  may not land an unused declaration awaiting a later consumer.
* **INV-5 — Atomic landing pairs are declared, never implied.** Where two tasks
  are genuinely not independently landable — one necessarily reddens a gate that
  **was passing on the tracked tree** and that only its partner can re-green —
  they must be declared an **atomic landing pair** here, carry an explicit
  bounded INV-1 carve-out naming which gate is red and for how long, and move
  their full-suite gate to the **pair boundary**. Ship lands such a pair as a
  **single commit**. Exactly one pair exists in this plan: **ALP-1 = {A-T1,
  A-T2}**. No other task in this plan may be sequenced this way without amending
  this invariant.
  > **Test-first red phases are explicitly excluded.** A task that adds a *new*
  > failing test which its successor turns green (B-T1 → B-T2, C-T2 → C-T3) is
  > the constitutionally required harness-before-code posture, not an atomic
  > landing pair: the red is a newly authored assertion, not a previously-passing
  > gate changing verdict. Those pairs land as separate commits. The distinction
  > is **"was this assertion green on the tracked tree at `9b299c8`?"** — for
  > ALP-1 it was; for B-T1 and C-T2 it did not exist. (B-T1 inverts an existing
  > assertion at `git_test.go:103-121`, which is an authored change to a test
  > this plan explicitly owns, not a collateral gate break.)

---

## Unit A — 030-S / 033-F: Unify retired-architecture scan scope declaration

**Retarget.** `scripts/lib/retired_arch.py` → `tools/gatecheck/internal/retiredarch/`.

**Premise correction carried forward (still true in Go):** there is **no live
scan-scope drift**. `selectRepoPaths` *does* enumerate `cmd/**`
(`select.go:126`). This unit removes a maintenance hazard; it does **not** fix a
live defect. Any report claiming a fixed defect here is wrong.

**The guarded decision (D4/AG-5) that survived the port:** `shouldScanRepoPath`
(`select.go:50-63`) is **not** a flat prefix disjunction. `cmd/` intentionally
**includes** `_test.go` and `testdata/` (`:54`); `internal/` **excludes** both
(`:57`, `:60`); `config.toml.example` (`:51`) is a third arm with no prefix at
all. A naive one-prefix-list unification would silently narrow `cmd/` coverage
while looking correct. This is the unit's central hazard.

### ALP-1 — A-T1 and A-T2 are an atomic landing pair (INV-5)

**A-T1 is not independently landable, by construction.** `pin.go` asserts against
`select.go`'s *source text*, so the instant AC-A1.2 moves the scope literals out
of `selectRepoPaths`/`shouldScanRepoPath`, `collectStringLits` finds nothing,
`containsAll` returns false (`pin.go:76`, `:80`), and the following all go red:

* `TestCheckPathspecPin_UnmodifiedCopy_Accepted` and `TestSelectionPathspecPin_LiveTree`;
* every `writeMutatedCopy` subtest — its anchor
  `git(root, "config.toml.example", "cmd/**", "internal/**")` must appear exactly
  once or the helper calls `t.Fatalf`;
* `TestCheckPathspecPin_PrefixMutation_Rejected` (`pin_test.go:133-141`), whose
  separate anchor `strings.HasPrefix(path, "internal/")` (`select.go:57`) AC-A1.2
  also removes;
* `TestRunRepoSelectionSelfTest_AssertionNamesAndOrder_MatchGolden`
  (`selftest_selection_test.go:11-43`), which fails on `out.failed`;
* the `"selection pathspec pin (AC-6/AG-1)"` assertion in `selftest_selection.go`
  (`:278-280`) — so `scripts/check-retired-architecture.sh --self-test` and
  `--self-test-integrity` flip PASS → FAIL.

This list is **exhaustive**: no other test or gate may be red at the ALP-1
interior boundary. This is `select.go:5-9`'s documented fail-closed design
working correctly, not a defect to engineer around.

Merging A-T1 and A-T2 into one task is **rejected**: it would span `select.go`,
`pin.go` and `pin_test.go` (3 files, ~5 functions) and breach the 2-hour rule.

**Therefore, per INV-5:**
* **Bounded INV-1 carve-out.** Between A-T1 and A-T2 — and *only* in the interior
  of ALP-1 — the retired-architecture pin assertion and the `pin_test.go` cases
  named above are **expected RED**. No other gate may be red at that boundary.
  The carve-out closes at A-T2.
* **Ship lands ALP-1 as a single commit.** A-T1 must not be pushed, reported
  complete, or CI-gated on its own.
* **The full-suite gate moves to the pair boundary** (AC-A2.6). A-T1's own
  acceptance is compile-, lint- and non-pin-test-scoped (AC-A1.6).

### A-T1 — Introduce `scanScope` and derive both consumers from it

Size **M**, complexity **medium**. Single file: `select.go`. **Half of ALP-1 —
not independently landable (INV-5).**

Introduce one `scanScope` declaration stating, per arm, both its `git ls-files`
pathspec form **and** its test/testdata policy, and rewire **both** consumers in
the same change: `selectRepoPaths` (`select.go:126`) builds its argument vector
from `scanScope`, and `shouldScanRepoPath` (`select.go:50-63`) evaluates from it.
Neither may restate a scope literal.

> **Why declaration and rewiring are one task (INV-4).** An unexported
> `scanScope` with no consumer is an `unused`/U1000 finding and would redden both
> blocking CI gates. A declaration-only task is therefore not independently
> landable. Scope stays within the 2-hour rule because it is one file and two
> functions.

**Acceptance criteria**
* AC-A1.1 — Exactly one `scanScope` declaration exists in `select.go`, carrying
  per-arm pathspec form and per-arm test policy **as data**, not as code paths.
* AC-A1.2 — Neither the pathspec construction nor the path predicate restates a
  scope literal; both read from `scanScope`.
* AC-A1.3 — The governing decision (D6/D6c as broadened by 012-S) is cited in the
  declaration so a future scope change is a one-line edit that cannot half-apply.
* AC-A1.4 — The selected path set is **identical to the set computed at this
  change's own parent commit** (not at a fixed snapshot).
* AC-A1.5 — The `cmd/`-includes-tests vs `internal/`-excludes-tests asymmetry and
  the `config.toml.example` arm are each preserved and each independently
  asserted. The existing assertions at `selftest_selection.go:185-197` and
  `:256-276` must still pass **unmodified in intent**.
* AC-A1.6 — Scoped to this task's own boundary (per ALP-1's carve-out):
  `go build ./...`, `golangci-lint run ./...` and `staticcheck ./...` are clean
  (INV-4), and `go test ./tools/gatecheck/... -count=1` passes **except for** the
  tests exhaustively enumerated in the ALP-1 declaration, which are expected red
  until A-T2. Those are gated at the pair boundary by AC-A2.6 — they
  are **not** waived. Any *other* red test fails this task.
* AC-A1.7 — The refactor prohibitions that A-T1 deliberately supersedes are
  rewritten in the same change: the whole `select.go:4-9` paragraph (both the
  "requires each literal to appear as a `*ast.BasicLit` inside the correct
  function body" premise at `:4-7` and the "Do not refactor these literals into
  shared constants, helper variables, or another file…" prohibition at `:7-9`)
  and `select.go:121-124` ("literal string arguments in THIS call … never a
  variable, slice literal defined elsewhere, or constant reference"). All must
  state the post-`scanScope` contract and name `scanScope` as the single
  permitted home for the literals. Leaving any of them in place is a false
  instruction to the next reader.

### A-T2 — Re-anchor the pathspec pin and add a negative control

Size **S**, complexity **medium**. `pin.go` plus its test. **Half of ALP-1 —
closes the carve-out and carries the pair-boundary gate (INV-5).**

**Corrected hazard (plan-review round 1, finding 3).** The round-1 narrative
claimed the pin would "silently stop pinning while still passing". That is
**wrong and inverted**. `checkPathspecPin` sets
`PathspecOK = containsAll(lits, pathspecPinLiterals)` (`pin.go:76`) and `PrefixOK`
likewise (`pin.go:80`). If the literals move into `scanScope`, `collectStringLits`
finds none, `containsAll` returns **false**, and the "selection pathspec pin
(AC-6/AG-1)" assertion in `selftest_selection.go` **FAILS loudly**. `select.go:5-9`
documents this as deliberate fail-closed behaviour.

**The real hazard is the opposite one, and it is unguarded today:**
`containsAll(_, [])` returns **`true`** — `containsAll` spans `pin.go:117-128`
and its vacuous `return true` is reached at `:127` when `wanted` is empty. So the
tempting way to make A-T1 land green — emptying `pathspecPinLiterals`
(`pin.go:33`) or `prefixPinLiterals` (`pin.go:37`) during this lockstep edit —
yields a **vacuously green pin that asserts nothing**. That, not a moved literal,
is what this task must prevent.

Re-anchor `checkPathspecPin` onto whichever declaration now holds the literals.
Note `findFuncBody` resolves only `*ast.FuncDecl`, so anchoring on a
package-level `scanScope` (`*ast.GenDecl`) requires a sibling collector.

> **Not a self-comparison.** `pin.go:9-15` records that `pathspecPinLiterals` and
> `prefixPinLiterals` are **this file's own independent expectation, never derived
> from `select.go`'s text**. Anchoring the pin on `scanScope` therefore remains
> genuinely falsifiable. (Round 1's AC-A2.5 wrongly called this a
> self-comparison, creating a contradiction with AC-A2.1; that claim is
> withdrawn.)

**Acceptance criteria**
* AC-A2.1 — The pin remains **source-text-anchored**: it parses the on-disk
  `select.go` via `go/parser` and compares against independently-authored literal
  lists that are **never derived from `select.go`'s own text**.
* AC-A2.2 — **Non-vacuity guard:** a committed test asserts that
  `pathspecPinLiterals` and `prefixPinLiterals` are each **non-empty**, so the
  `containsAll(_, []) == true` vacuous-pass mode (`pin.go:127`) cannot be
  reached by emptying them.
* AC-A2.3 — **Negative control, as a committed test artifact** (not a one-time
  manual demonstration): a deliberately narrowed scope — e.g. a fixture omitting
  the `cmd/**` arm — makes the pin **FAIL**. A pin that passes against a narrowed
  scope fails this task.
* AC-A2.4 — All pre-existing `--self-test` and `--self-test-integrity` cases pass.
* AC-A2.5 — `golangci-lint run ./...` and `staticcheck ./...` clean (INV-4).
* AC-A2.6 — **ALP-1 pair-boundary gate (INV-5).** At this task's completion the
  carve-out is closed: `go test ./tools/gatecheck/... -count=1` passes in full
  — including every test enumerated in the ALP-1 declaration and deferred by
  AC-A1.6 — and `scripts/check-retired-architecture.sh`, `--self-test` and
  `--self-test-integrity` all return to PASS. ALP-1 is reported complete only
  here, as one commit.
* AC-A2.7 — The pin's own fixtures and prose are re-anchored with it: the
  `writeMutatedCopy` anchors in `pin_test.go` (including
  `TestCheckPathspecPin_PrefixMutation_Rejected`'s separate
  `strings.HasPrefix(path, "internal/")` anchor at `pin_test.go:133-141`) target
  `scanScope`, and `pin.go:1-7` — which states the pin requires the literals
  "INSIDE the `selectRepoPaths`/`shouldScanRepoPath` function bodies" — is
  rewritten to describe the `scanScope` anchor. (`pin.go:9-15`, the H-11
  self-comparison note, stays true and is **not** modified.)

### Dropped from Unit A

* **033.003-T — DROPPED.** The Python `inspect.getsource()` pin became a stronger
  `go/parser` AST pin during the port (`pin.go:56`, `:156-162`), and the asymmetry
  assertions already exist (`selftest_selection.go:185-197`, `:256-276`).
  Re-implementing it would be green-on-arrival. Its surviving intent is carried as
  AC-A2.1 through AC-A2.3.

**Known residual (not closed here).** Stash `DD0BB60F` records that
`checkPathspecPin` only verifies pin literals occur *somewhere* in the target
body (`pin.go:76`), not that they are the actual arguments of the `git` call.
AC-A2.3 narrows but does not close this.

---

## Unit B — 032-S / 035-F: Repair unignore-regression checker ref resolution

**Retarget.** Python `root_gitignore_text_at` → `rootGitignoreTextAt`
(`tools/gatecheck/internal/unignore/git.go:84`).

**Measured correction, restated for Go.** The checker does **not** currently fail
open in practice: `runDifferentialCheck` fails closed on an unresolvable ref
(`checks.go:90-96`) **before** the use sites at `checks.go:124` and `:128`. The
defect is **latent defence-in-depth**, not a live escape. It is worth fixing
because the guard and the use site are separated and a future refactor could
reorder them.

**What the port made worse.** `rootGitignoreTextAt` returns `"", nil` on **any**
non-zero `git show` exit (`git.go:100-104`), discarding stderr
(`stdout, _, err := …`) and collapsing *invalid ref* and *legitimately absent
`.gitignore`* into one indistinguishable result.

### B-T1 — Invert and extend the direct baseline-resolution tests

Size **S**, complexity **medium**. Posture: **test-first — this task supplies the
red and is authored first.** Single file: `git_test.go`.

Tests must drive `rootGitignoreTextAt` **directly**, not through the scenario
harness — the end-to-end path is short-circuited by the upstream guard and cannot
reach the defect. A scenario-level test would be green-on-arrival and certify
nothing.

**This task is partly a test *correction*.** `git_test.go:103-121`
(`TestRootGitignoreTextAt_InvalidRef_DegradesToEmpty`) currently asserts
`err == nil` and `text == ""` for an invalid ref — i.e. it **pins the defect as
correct** — and its comment at `:103-107` names 035-F. It must be inverted and
renamed, and that comment is **owned by this task**, not by B-T3.

Required cases: (a) a **valid non-HEAD ref** whose `.gitignore` is **absent** →
`("", nil)`; (b) an **invalid/nonexistent ref** → distinct non-nil error.
Case (a) is **not optional**: without it, B-T2 could regress the legitimate
absent-file path into an error undetected.

**Acceptance criteria**
* AC-B1.1 — Both cases drive `rootGitignoreTextAt` directly, not via the scenario
  harness.
* AC-B1.2 — The inverted `:103-121` assertion and case (b) both **FAIL against the
  pre-change function**. Leaving the original degrade-to-empty assertion in place
  fails this task.
* AC-B1.3 — The stale characterisation comment at `git_test.go:103-107` is
  corrected as part of this task.
* AC-B1.4 — Pre-existing coverage at `git_test.go:35-43` (HEAD/absent → `""`) and
  `:58-101` (valid ref/present) still passes unmodified.

### B-T2 — Fail closed on an unresolvable base ref at the point of use

Size **S**, complexity **medium**. Depends on B-T1 (turns its red green). Single
file: `git.go`.

Make `rootGitignoreTextAt` distinguish its two failure modes: a valid ref whose
`.gitignore` is absent returns `("", nil)`; an unresolvable ref returns a
**distinct non-nil error**. This requires inspecting `git show`'s stderr/exit
shape — currently discarded at `git.go:99` — not merely its exit status, which is
why this is S/medium rather than XS.

**Acceptance criteria**
* AC-B2.1 — An unresolvable ref produces a distinct non-nil error at the point of
  use, independent of the upstream `checks.go:90-96` guard.
* AC-B2.2 — A valid ref whose `.gitignore` is absent still returns `("", nil)`.
  The legitimate absent-file path is **not** converted into an error.
* AC-B2.3 — The HEAD path (`git.go:85-98`) is unchanged: absent `.gitignore` on
  disk still yields `""`.
* AC-B2.4 — Every B-T1 test passes; existing merge-blocking behaviour is unchanged
  for resolvable refs (INV-1).

> **Declared interim inconsistency.** B-T2 changes the failure semantics but
> leaves the doc comments at `git.go:75-83` and `:101-103` still asserting the old
> "both degrade to an empty baseline" contract. They are knowingly false for
> exactly one task boundary and are corrected by B-T3, which owns them. This is a
> comment-only lag, not an INV-5 atomic-landing pair: no gate is red between B-T2
> and B-T3.

### B-T3 — Correct the stale baseline-semantics comments in `git.go`

Size **XS**, complexity **low**. Documentation-only. Single file: `git.go`.

Correct `git.go:75-83` and `git.go:101-103`, which both still state that an
invalid ref and an absent `.gitignore` "both degrade to an empty baseline".

**Scope boundary:** `git_test.go:103-107` is **not** in scope here — B-T1 owns it.
This task touches `git.go` comments only.

**Acceptance criteria**
* AC-B3.1 — Each corrected comment names the actual mechanism (`git show` of the
  ref-qualified path) and the corrected failure mode.
* AC-B3.2 — No executable line in `git.go` is modified by this task.

---

## Unit C — NEW: gatecheck CLI containment hardening

Harvested from stash **7223218F** (high), **50E6F22C** (medium), **44F8CC48** (low).

**The live gap.** `parseRoot` (`main.go:74-102`) scans the **whole** forwarded
argument list for `--root`/`--root=` with **last-occurrence-wins** semantics
(`root = args[i+1]`, no `break`). `gatecheck-run.sh:66` appends the trusted
`--root ${ROOT}` **ahead of** whatever a wrapper forwards. Therefore any wrapper
forwarding more than a single fixed non-flag argument must carry its own
`--root` denylist.

Two wrappers do, each hand-rolled differently:
`check-retired-architecture.sh:231-249` (allowlists the first arg only) and
`check-unignore-regression.sh:94-101` (per-argument denylist loop).
**`check-write-path-precondition.sh:84` forwards raw `"$@"` with no guard at
all** — a real override gap today.

Verified safe to change: all four `register_*.go` subcommands forward args and
none consumes `--root`; no wrapper passes two `--root`s; `main_test.go` contains
no test pinning last-wins semantics.

### C-T1 — Make the trusted `--root` authoritative in `parseRoot`

Size **S**, complexity **medium**. Posture: test-first. `main.go` + `main_test.go`.

Change `parseRoot` so the **first** `--root` occurrence wins and any **subsequent**
`--root`/`--root=` is a hard error (exit 1, named diagnostic) rather than a silent
override. Because `gatecheck-run.sh:66` always prepends the trusted root,
first-wins makes it authoritative for **all four** engines at once, removing the
need for every wrapper to carry its own copy of the guard. This is the central
cross-cutting fix (option (a) of 7223218F); **C-T3**'s wrapper denylist is
defence-in-depth, not the primary control.

**Acceptance criteria**
* AC-C1.1 — A caller-supplied `--root` or `--root=` appearing **after** the
  trusted one causes a non-zero exit with a named diagnostic; it never overrides.
* AC-C1.2 — The trusted root prepended by `gatecheck_invoke` is the one used, for
  all four subcommands.
* AC-C1.3 — Existing single-`--root` invocations are unaffected (INV-1),
  demonstrated by the **actual** self-test entry points:
  `scripts/check-retired-architecture.sh --self-test`,
  `scripts/check-write-path-precondition.sh --self-test`,
  `scripts/check-unignore-regression.sh --self-test`, and
  `scripts/check-merge-strategy.sh --self-test`. (The merge-strategy self-test is
  implemented in the **wrapper** `run_self_test` at
  `check-merge-strategy.sh:171`; `mergestrategy.Run` itself has no self-test mode.)
* AC-C1.4 — Table-driven tests in `main_test.go` cover: trusted-only, trusted +
  trailing `--root`, trusted + trailing `--root=`, and `--root` with no value.

### C-T2 — Add the write-path wrapper guard test

Size **S**, complexity **low**. Go test only. Posture: test-first. Depends on C-T1.

Add `tools/gatecheck/check_write_path_wrapper_test.go` mirroring
`check_unignore_wrapper_test.go`, including that file's reliance on the guard
firing before `gatecheck_build` (`check_unignore_wrapper_test.go:26-27`).

> **Ordering (plan-review round 2, finding P3-1).** This test task is sequenced
> **before** the shell guard it covers, so Unit C keeps the same test-first
> posture C-T1 adopts: C-T2 supplies the red, C-T3 turns it green. The earlier
> guard-then-test ordering left the new guard unverified for one boundary.
>
> Split from C-T3 because mirroring the unignore test means two test functions
> plus a path helper and a trace helper. The unignore mirror has **10 scenarios**
> (5 rejection + 5 acceptance); write-path's legitimate CLI surface is narrower
> (`no args`, `--self-test`, `--self-test-integrity` — unknown flags already exit
> 2 via `writepath.Run`), so the realistic mirror is ~5 rejection + 3 acceptance.
> Either way, bundled with the shell edit this would exceed the 2-hour rule and
> mix the shell and Go-test skill domains, so the split stands.

**Acceptance criteria**
* AC-C2.1 — Rejection cases cover `--root`, `--root=…`, and a `--root` following
  a legitimate flag; acceptance cases cover `--self-test` and
  `--self-test-integrity`.
* AC-C2.2 — Each rejection subtest **asserts** that `gatecheck_build` was not
  reached, using the same trace mechanism as
  `check_unignore_wrapper_test.go:103` (`bash -x` plus a trace scan). The
  assertion is authored here and **goes green at C-T3** — pre-C-T3 the wrapper
  has no guard, so every rejection case necessarily reaches
  `gatecheck_build` (`check-write-path-precondition.sh:78`). This AC owns the
  assertion's *existence and shape*; AC-C3.2 owns its *satisfaction*.
* AC-C2.3 — The rejection cases **fail** against the pre-C-T3 wrapper. A suite
  that is green on arrival fails this task.
* AC-C2.4 — The file is gofmt and goimports clean (INV-4).

> **Red phase is bounded and declared.** C-T2's rejection cases are red until
> C-T3. Per INV-5's test-first exclusion this is **not** an atomic landing pair:
> the red assertions are newly authored and did not exist on the tracked tree at
> `9b299c8`, and `scripts/check-write-path-precondition.sh` itself stays PASS
> throughout. C-T2 and C-T3 may land as separate commits.

### C-T3 — Add the per-argument `--root` denylist to the write-path wrapper

Size **XS**, complexity **trivial**. Shell only. Depends on C-T2.

Add to `scripts/check-write-path-precondition.sh` the per-argument denylist loop
used at `check-unignore-regression.sh:94-101`.

**Placement is load-bearing:** the guard must run **before** `gatecheck_build`
(currently `check-write-path-precondition.sh:78`), mirroring the unignore
analogue which guards at `:94` before sourcing and building. A post-build
placement makes every rejection test pay a full `go build`.

**Acceptance criteria**
* AC-C3.1 — Every forwarded argument is checked against a literal
  `--root`/`--root=*` denylist before pass-through; a match exits non-zero with
  the same `::error::unrecognized argument:` diagnostic shape the unignore wrapper
  uses.
* AC-C3.2 — The guard executes **before** `gatecheck_build`.
* AC-C3.3 — Legitimate invocations (no args, `--self-test`,
  `--self-test-integrity`) are unaffected.
* AC-C3.4 — Every C-T2 scenario now passes.

### C-T4 — Correct the wrapper rationale comments superseded by C-T1

Size **XS**, complexity **trivial**. Documentation-only. Depends on C-T1.

C-T1 makes five live rationale comments factually wrong; each currently
documents the superseded last-wins security model:
`check-unignore-regression.sh:70-72`, `check-retired-architecture.sh:210-211`
(the "LAST occurrence wins" sentence) and `:224-227` (the "`$2` onward ignored"
rationale that depends on it), `check_unignore_wrapper_test.go:14-20`, and
`check_retired_architecture_wrapper_test.go:14-20`.

**Acceptance criteria**
* AC-C4.1 — All five comments describe the post-C-T1 model: the trusted root is
  authoritative because the **first** occurrence wins, and the wrapper denylists
  are retained as defence-in-depth.
* AC-C4.2 — No executable line is modified by this task.

### C-T5 — Record gatecheck's CLI containment posture as explicit residuals

Size **S**, complexity **low**. Documentation-only. Independent.

Document — do not silently inherit — the three Principle III items from 50E6F22C,
**all three confirmed faithful parity with the retired Python original, not
regressions introduced by the port**:

1. `mergestrategy.Run` accepts `root` and discards it (`evaluate.go:192`
   `_ = root`), reading the caller-supplied path via `os.ReadFile` (`:59`) with no
   containment check. Required by the live-transport design: the wrapper's
   temp-file payload is created in the OS temp directory, outside the repo root,
   by design. Matches the original `evaluate_json`'s `open(sys.argv[1])`.
2. `unignore`'s self-test creates scenario/scratch repos via `os.MkdirTemp("", …)`
   outside root. Matches the original `tempfile.TemporaryDirectory()`.
3. `mergestrategy` silently evaluates only `args[0]` (`:197`) and ignores extra
   positional arguments. Matches the original's unconditional `sys.argv[1]`.

**Acceptance criteria**
* AC-C5.1 — All three are recorded in the `mergestrategy`/`unignore` package docs
  with their parity justification and the evidence that each matches the retired
  Python.
* AC-C5.2 — The record states explicitly that these are **accepted, scoped
  Principle III exceptions**, not undiscovered holes.
* AC-C5.3 — **No behaviour is changed by this task.** In particular item (3) is
  **not** fixed here. The distinguishing factor is **INV-1's explicit-AC
  carve-out**, not exit-code or trigger novelty: C-T1 also adds a new trigger to
  a currently-passing invocation shape, but it does so under an acceptance
  criterion (AC-C1.1) that names the change. Item (3) has no such authorisation,
  so changing it would be an unauthorised INV-1 deviation. (For completeness:
  `mergestrategy.Run` already returns 2 for usage at `evaluate.go:193-195`, so
  the fix would reuse an existing exit code — the exit code was never the
  obstacle.) Recorded as a candidate ED entry for a future revision.

### C-T6 — Correct the stale `.gitignore` gate-engine comment

Harvested from 44F8CC48. Size **XS**, complexity **trivial**. Independent.

`.gitignore:126-127` still reads "Python bytecode caches written when the
extracted gate engines under `scripts/lib/` are imported or executed (032-F,
shipment 029-S)". Those engines were retired in M4 (037-S). Append a clarifying
line noting the rule is now defensive/historical.

`.gitignore` is an **append-only gated file** and a different contract surface
from the Go code — hence a comment **addition**, never a rewrite or deletion.

**Acceptance criteria**
* AC-C6.1 — A comment line is **appended** near the existing block noting the
  `scripts/lib/` engines were retired in M4 (037-S) and the rule is now
  defensive/historical.
* AC-C6.2 — The existing `__pycache__/` and `*.pyc` rules are **not** removed or
  reordered — removing an ignore rule is exactly what
  `check-unignore-regression.sh` blocks.
* AC-C6.3 — Both gates pass **using their real CI entry points** (neither script
  may be invoked bare — both fail closed with exit 1 when given no ref
  arguments): `scripts/check-gitignore-append-only.sh --self-test` plus the
  `--base-ref`/`--head-ref` form used at `ci.yml:464`, and
  `scripts/check-unignore-regression.sh --self-test` plus the `--base-ref`/
  `--head-ref` form used at `ci.yml:486`.

---

## Unit D — 031-S / 034-F (revised): Harden the write-path gate — masked-text increment

**Decision.** D-031-2 (operator, 2026-10-01): **split 031-S.** This unit keeps the
Option A work that **carries over** to a later `go/ast` engine — the D-1/D-2
mechanism, the selector set, the residual record and, above all, the fixture
corpus, which becomes Unit E's **verdict-parity regression corpus**. The parts
that are hard as text scanning and that `go/ast` would replace (named import
aliases, `NewRoot` receiver tracking) move to Unit E.

**Rev-4 plan-review amendments (deliberation D-031-3).** Round 1 of the rev-4
review returned **FAIL** (1 P0, 6 P1). Stage amended the Python-era mechanism
decisions for the Go engine:

* **D-1′** keeps the `SplitLines` line loop plus a byte cursor, so line parity
  and the `retiredarch` scan-loop pin both hold. It tracks `()[]{}` and
  enumerates every occurrence.
* **D-2′** adds exact creation-disposition and flags tokens, a bare-operand rule
  and per-occurrence evaluation.
* **D-T1** and **D-T5** are split under the 2-hour rule, into new tasks 034.010-T
  and 034.011-T.

Neither D-1′ nor D-2′ widens what the gate allows. Each only narrows the
allowance or preserves parity.

**Retarget.** `scripts/check-write-path-precondition.sh` (Python body, retired) →
`tools/gatecheck/internal/writepath/writepath.go`. Fixtures stay at
`scripts/testdata/writepath/*.go` (`runFixtureSelfTest`, `writepath.go:220-221`).

**Facts re-measured at `9b299c8` that shape every task:**

* `scanText` (`writepath.go:112-124`) is line-based over `pysem.SplitLines`;
  `findSelector` (`:89-102`) is `strings.Index` plus `pysem` boundary checks.
  Findings are emitted **per (line, selector) pair in `Selectors` order** and a
  selector hit **at most once per line** — both are pinned by the golden.
* `pysem.SplitLines` uses **Python `splitlines` boundaries** (`\n`, `\r`, `\r\n`
  **and** `\v`, `\f`, `\x1c`–`\x1e`, `\x85`, `\u2028`, `\u2029`). D-1's
  "count `\n` before the offset" is therefore **not** automatically identical to
  the current line numbering. D-1′ (D-T1) closes this hazard by construction:
  it keeps the `SplitLines` loop and adds a byte cursor.
* `retiredarch/writepath_mask_test.go` **pins writepath's scan-loop shape**
  (`:15-29`, `:265-302`):
  * `scanText` must range over `pysem.SplitLines` of its second parameter in a
    top-level statement;
  * `scanFile` must pass a direct `gomask.MaskGoNonCode` call;
  * no other non-test writepath source may reference `scanText`.

  The original D-1 ("`scanText` stops iterating lines") would break this pin.
  This is the reason for D-1′.
* `writepath_golden.json` pins **three** things that this unit changes:
  the `selectors` array (`TestSelectors_MatchesGolden`), per-fixture
  `fixture_findings` rows, and the `self_test` / `self_test_integrity` stream
  captures, whose stdout carries one `PASS` line per fixture. **Every task that
  adds a selector or a fixture must update the golden in the same commit** or
  `go test ./tools/gatecheck/...` goes red (INV-4).
* The target call `internal/pathsafe/reparse_windows.go:164-172` still has
  **exactly 7 arguments with arg index 1 the literal `0`** (`:166`, followed by
  a comment the masker blanks) and a trailing comma before `)`. The four
  comment-only `syscall.CreateFile` mentions are at `:24`, `:28`, `:33`, `:145`.
* None of `syscall.CreateFile` (as a code reference other than `:164`),
  `syscall.Write`, `os.OpenRoot`, `os.Root`, `io.CopyN`, `io.CopyBuffer` occurs in
  any of the **17** non-test `.go` files under `internal/**`/`cmd/**`.
  `io.CopyN` is **not** caught by the existing `io.Copy` selector
  (`FollowedByWord` rejects the trailing `N`), confirming B72E9715.
* The only named import alias in scope is `copilot` →
  `github.com/github/copilot-sdk/go` (`internal/copilotprobe/{client,fixture,permission}.go`),
  which is not write-capable. There are **zero** production `.Resolve(` callers.
* No atomic landing pair is needed in this unit (INV-5 unchanged): the predicate
  (D-T2) lands **before** the selector that needs it (D-T3), so the repo scan
  never sees an unallowed `syscall.CreateFile`. Every intermediate task state
  stays inside the single 031-S shipment PR, so the scan-level positive control
  arriving one task after D-T3 (in D-T5a) never reaches `main` separately.

### D-T0 — 034.001-T: SATISFIED by the port (no executable scope)

Handled exactly as 033.003-T was (Unit A, "Dropped from Unit A"). The spike at
`docs/decisions/2026-09-19-intercom-go-call-extent-extraction-and-allowance-predicate-spike.md`
exists and records D-1..D-4; this revision re-anchors D-1/D-2 onto the Go engine.
The task is retitled `SATISFIED (post-M4 Go re-plan) …`, keeps `queued` status and
its 031-S membership, and is **removed from the live dependency chain** (edge
`034.002-T → 034.001-T` deleted) so a no-op item cannot gate real work.

> **Lifecycle constraint (same as 033.003-T).** `.backlogit/hooks.yaml` allows
> only `queued → active|blocked`; there is no `queued → archived|rejected`
> transition, so Stage cannot give the item a terminal state. Removal from the
> manifest **is** possible (`shipment return-blocked`, see Unit E) but would leave
> a permanently queued orphan; retention keeps it visible in the manifest Ship
> already reconciles. **Named prerequisite for 031-S closure** (shared with
> 033.003-T in 030-S): before Ship's closure classification, the operator either
> authorises Ship to close it as an explicit no-op item (counted in the
> two-set gate as "satisfied-by-prior-work", not as delivered scope), or gives
> both items a terminal state. If the first option is taken, the closure
> artifact records it as an **explicit, disclosed deviation** from the
> two-set gate computation (compound lesson
> `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`).
> A narrative compliance claim does not satisfy this.

### D-T1a — 034.010-T (NEW): Commit a frozen differential oracle for detection parity

Size **S**, complexity **low**. Posture: characterization-first (declared
**green on arrival** — it pins today's behaviour). One new file,
`tools/gatecheck/internal/writepath/writepath_oracle_test.go`. No dependency.

Split out of D-T1 at plan-review rev-4 (2-hour rule; P1 oracle-drift finding).
Copy the pre-change `scanText` **and** `findSelector` (`writepath.go:89-124`)
verbatim into the test file under test-only names (never the identifier
`scanText` — `retiredarch`'s `TestWritePathScanFileConsumesCanonicalMask` only
inspects non-test files, but the names must not collide), together with a
**frozen copy of the original 20 selectors**. The oracle compares the frozen
legacy scan against production `scanText` **restricted to findings that name
one of the frozen 20 selectors**.

The restriction is an **exact set lookup** on the selector parsed from the
finding text, compared against `pysem.Repr(sel)` for each frozen selector. The
token is parsed **from the right** (`TrimSuffix(" found")`, then the text after
the `LastIndex` of `"write primitive "`), because the path is free text. It is never a substring
test, because `'io.CopyN'` and `'io.CopyBuffer'` contain `io.Copy`.

D-T3 only **appends** selectors, and the D-2′ allowance only ever applies to
`syscall.CreateFile`, which is not in the frozen 20. The restriction therefore
keeps the oracle valid, **unedited**, through all of Unit D.

**Oracle lifecycle (rev-4 round-2 P1 fix).** The oracle is a **Unit D**
masked-text harness. Its **legacy side and expectations are frozen for its
whole lifetime**. Exactly three things may be adapted, **exactly once**, in
E-T2, by AC-E2.6:
* the production side;
* the input list;
* the expectations for **unparseable** inputs, which may only move to
  fail-closed. That is the only authorised edit, and it
can only make the oracle stricter. No other task may edit the oracle (H-3).

**Acceptance criteria**
* AC-D1a.1 — The oracle runs over:
  * every `scripts/testdata/writepath/*.go` fixture;
  * every `filebased` golden input;
  * every tracked non-test `internal/**`/`cmd/**` file.

  Each input is masked exactly as `scanFile` masks it. The restricted
  production findings must equal the oracle's findings in text, line and order.
  An input that `pysem.ReadText` rejects (invalid UTF-8) is not a parity input:
  the oracle asserts that **both** sides error.
* AC-D1a.2 — Hand-built inputs pin line numbering to the `pysem.SplitLines`
  boundary set:
  * one input contains `\f`, `\v` and `\u2028` in **code position**, with a
    selector hit on each following line;
  * one input contains the same boundary runes **inside a comment and inside an
    interpreted string**, followed by a selector-hit line.

  The masker turns in-comment and in-string boundaries into spaces, so the
  legacy line number of that hit is the **masked** line number. This is a parity
  hazard Unit E's G-3 must reproduce. Every hand-built input except the
  code-position one is a **complete, parseable Go file**, with a `package`
  clause.
* AC-D1a.3 — Green on arrival at its own parent commit; no production file is
  modified; lint clean.

### D-T1 — 034.002-T: Call-extent extraction over masked text (D-1′)

Size **S**, complexity **medium** (already de-risked by the 2026-09-19 spike's
D-1 feasibility proof). Posture: characterization-first.
`writepath.go` + `writepath_test.go`. Depends on D-T1a.

Implement **D-1′** (D-1 as amended by deliberation D-031-3; the extent
algorithm is unchanged, the scan loop is not):

* **Keep the line loop.** `scanText` keeps its signature and its single
  top-level `range pysem.SplitLines(maskedText)`. It additionally maintains a
  **byte cursor** into the whole masked text: each line is exactly
  `maskedText[cur:cur+len(line)]`, and the cursor then advances past exactly
  one boundary (`\r\n` as two bytes, otherwise the boundary rune's UTF-8 width).
  This is why D-1's "count `\n` before the offset" is dropped.

  The cursor **fails closed**:
  * it asserts `maskedText[cur:cur+len(line)] == line`;
  * it bounds-checks the boundary decode on the last line.

  On any mismatch, `scanText` **disables the allowance for the rest of that
  text**, so every occurrence is reported. This fails closed as a false
  positive, with no signature change. `scanText` returns `[]string` and has no
  error channel, and the oracle and `TestFilebased_*` call it single-valued.

  The mismatch is unreachable through `scanFile`, because:
  * `ReadText` rejects invalid UTF-8 and reports it as ED-2 `::error::` via
    `errorLine`;
  * `MaskGoNonCode` emits only input runes, `' '` and `'\n'`.

  A direct-call test with invalid UTF-8 pins the fallback. Line numbers stay **identical by construction**, and the `retiredarch`
  architectural pin
  (`retiredarch/writepath_mask_test.go:15-29,265-302`: `scanText` must range
  over `pysem.SplitLines` of its second parameter at top level, `scanFile` must
  pass `gomask.MaskGoNonCode` directly, and nothing else may reference
  `scanText`) passes **unmodified**.
* **Enumerate every occurrence.** For each `(line, selector)` it enumerates
  **every** non-word-bounded occurrence of the selector on that line, not only the
  first. The legacy `findSelector` stops at the first match.
* **Extent extraction** for each occurrence, at its absolute offset, over the
  whole masked text:
  * skip whitespace (newlines included);
  * require `(`;
  * walk to the balanced `)` while tracking a **bracket stack over `()`, `[]` and `{}`**. A mismatched closer is **undecidable**, so is reaching EOF;
  * split the interior at commas where the stack depth is exactly 1, i.e. inside the call's own parentheses only.

  A selector followed by anything but `(` is a **non-call reference**.
* **No verdict change.** Per D-4, the extent result **cannot change any verdict**
  in this task (no predicate yet). A line still yields **at most one finding
  per selector**, in `(line, Selectors index)` order.
* **Consumed, not discarded (INV-4).** The per-occurrence evaluator consumes
  the extent classification, which feeds the allowance hook. In D-T1 that hook
  returns "not allowed" for every classification, and D-T2 replaces it. No
  computed-and-unused value is introduced (staticcheck SA4006).
* **`findSelector` stays a production function.** It is re-expressed as the
  first enumerated occurrence, so `TestFindSelector_LookaroundTable`
  (`writepath_test.go:166-189`) passes **unmodified**. Its named production
  caller is the per-line fast-path guard (`if !findSelector(line, sel) {
  continue }`) ahead of full enumeration. It must not become a test-only
  declaration.

**Acceptance criteria**
* AC-D1.1 — For the masked text of `internal/pathsafe/reparse_windows.go`, the
  extractor recovers the `syscall.CreateFile` call (located **by content**, not by
  a hard-coded line number) as one balanced extent of **8** raw depth-1 segments
  (7 arguments plus the trailing-comma artifact). Segment 1 is `0` after
  `strings.TrimSpace`, and segment 4 is `syscall.OPEN_EXISTING`. The four
  comment-only mentions produce **no** occurrence.
* AC-D1.2 — Unit tests classify **non-call**, **unbalanced (EOF)**, and
  **mismatched-bracket** extents as such. Each still yields the unchanged
  finding text `write primitive '<sel>' found`, so no new classification text
  appears.
* AC-D1.3 — A multi-line extent whose lines are separated by `\f` or `\u2028`
  maps every occurrence back to the correct `SplitLines` line (cursor test).
* AC-D1.4 — D-T1a's oracle, every golden test, and `retiredarch`'s
  `TestWritePathScanFileConsumesCanonicalMask` / `TestWritePathMaskFlow_*` pass
  **unmodified**. All three `check-write-path-precondition.sh` entry points
  pass. `golangci-lint`/`staticcheck` are clean (INV-1, INV-4). The extractor is
  reached from the production scan path, with no test-only declaration (INV-4).
* AC-D1.5 — If D-1′ cannot be followed as specified, or the `retiredarch` pin
  cannot stay unmodified, **HALT and return to Stage** (P-010). Do not
  re-decide the mechanism, and do not edit `retiredarch`.

### D-T2 — 034.003-T: Access-mode allowance predicate (D-2′)

Size **S**, complexity **medium**. Posture: test-first.
`writepath.go` + `writepath_test.go`. Depends on D-T1.

Implement **D-2′** (D-2 as amended by D-031-3). A single occurrence is
**allowed** only when **all** of the following hold:

1. the selector is exactly `syscall.CreateFile`;
2. the occurrence is a real call whose extent balanced under D-1′;
3. the extent interior contains **none** of `(` `)` `[` `]` `{` `}`, a
   backquote, or a double quote. This rejects call arguments, composite
   literals, index expressions, generic instantiation and `T{…}.Args()`-style
   multi-value smuggling.

   The quote check can only catch **visible tag-shaped raw strings**. `gomask`
   blanks interpreted strings and rune literals, delimiters included, so those
   are caught by rule 3b or by type checking.
   * **3b.** After discarding **one** trailing empty segment, **every** segment
     is non-empty after `strings.TrimSpace`. A blanked string or rune argument
     leaves an empty segment and is rejected.
4. after discarding **one** trailing empty segment there are **exactly 7**
   arguments;
5. argument 1 (`dwDesiredAccess`), trimmed, is **exactly the token `0`**;
6. argument 4 (`dwCreationDisposition`), trimmed, is **exactly the token
   `syscall.OPEN_EXISTING`**. Access `0` with `CREATE_NEW`, `CREATE_ALWAYS`,
   `OPEN_ALWAYS` or `TRUNCATE_EXISTING` can still create or modify a file;
7. argument 5 (`dwFlagsAndAttributes`), trimmed, is **exactly the token
   `syscall.FILE_FLAG_BACKUP_SEMANTICS`** and nothing else. No `|`
   combination, no other flag, and no operator is allowed. This excludes
   `FILE_FLAG_DELETE_ON_CLOSE` and every other flag.

   Round 2 narrowed this rule to the live call's exact flag. An earlier draft
   also allowlisted `FILE_FLAG_OPEN_REPARSE_POINT`, which no caller uses.

Any other spelling (`0x0`, `00`, `(0)`, `uint32(0)`, a named constant, or an
equivalent numeric disposition) is **rejected**. The predicate can only fail
as a false positive (INV-2).

The predicate runs on **each occurrence independently**. A line emits a finding
for a selector when **any** occurrence of it on that line is not allowed, and
the at-most-one-finding de-duplication is applied **after** the predicate. This
way an allowed call never hides a writing call on the same line.

The predicate is wired into the production scan path keyed on the selector
string, so it stays **dormant until D-T3** adds `syscall.CreateFile` to
`Selectors`. Its scan-level positive control therefore lands with the D-T5a
fixtures, inside the same shipment PR.

**Acceptance criteria**
* AC-D2.1 (AC-4.2, function level) — The real `reparse_windows.go` extent
  (located by content) is **allowed**. The allowance is keyed on the call's
  arguments, never on the file path.
* AC-D2.2 — Table tests **reject** each of the following:
  * a non-call reference, an unbalanced extent, and a mismatched-bracket extent;
  * 6 and 8 arguments;
  * each non-`0` access spelling listed above;
  * `CREATE_NEW`, `CREATE_ALWAYS`, `OPEN_ALWAYS` and `TRUNCATE_EXISTING`;
  * `FILE_FLAG_DELETE_ON_CLOSE`, alone and OR-ed with the allowed flag;
  * a brace composite (`T{p, 0, a, b, c, d, e}.Args()`);
  * a call-valued argument;
  * a visible tag-shaped raw-string argument (rule 3);
  * an interpreted-string or rune argument in an unconstrained slot (rule 3b,
    empty segment);
  * `syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT`
    (rule 7 exactness).
* AC-D2.3 — A per-occurrence test proves that an allowed call and a writing call
  on **one line**, in both orders, produce the finding. Until D-T3 adds
  `syscall.CreateFile` to `Selectors`, this test calls the per-line evaluator
  with the selector **passed as a parameter**, not through `scanText`.

*Scenario counting (2-hour rule):* a table-driven test counts as **one** test
scenario. D-T2 therefore has three scenarios: AC-D2.1 allow, the AC-D2.2
rejection table, and AC-D2.3 per-occurrence.
* AC-D2.4 — These tests are authored first and **fail** before the predicate
  exists (red phase; newly authored assertions, so not an INV-5 pair).
* AC-D2.5 — D-T1a's oracle, the golden tests and the `retiredarch` pin are
  unmodified and green; `Selectors` is unchanged; lint is clean.
* AC-D2.6 — If D-2′ cannot be followed as specified, **HALT and return to Stage**.

### D-T3 — 034.004-T: Add six decidable selectors (folds in B72E9715)

Size **S**, complexity **low**. `writepath.go` +
`testdata/writepath_golden.json`. Depends on D-T2.

Append `syscall.CreateFile`, `syscall.Write`, `os.OpenRoot`, `os.Root`,
`io.CopyN` and `io.CopyBuffer` to `Selectors` (20 → **26**). Also correct the
`Selectors` doc comment (`writepath.go:28-31`, "ordered list of 20 … ported
verbatim"), which this change makes false.

Scope corrections carried from the original task:
* `os.Root` **method** calls (`root.Create`) are **not** decidable by
  qualified-selector matching and are **not** attempted (residual, D-T4). Only
  `os.Root` type-use and `os.OpenRoot` construction are added.
* `syscall.Write` does **not** match `syscall.WriteFile`, because
  `FollowedByWord` rejects the trailing `File` (residual item 6).
* The unlisted DB-driver selectors item stays removed (no database dependency
  in `go.mod`).

**Acceptance criteria**
* AC-D3.1 — The six selectors are present and appended **after** the existing
  20, so existing finding order is unchanged. The golden `selectors` array is
  updated in the **same commit**: red first (golden edited, test fails), then
  `Selectors`.
* AC-D3.2 (AC-4.2, gate level) — `scripts/check-write-path-precondition.sh`
  (bare) still **PASSES** on the tracked tree: D-T2 allows the
  `reparse_windows.go` call, and the comment-only mentions are masked.
  `TestRun_RepoMode_CurrentTreeIsClean` passes unmodified.
* AC-D3.3 (AC-4.8) — Zero false positives across the 17 non-test files.
* AC-D3.4 — `--self-test` and `--self-test-integrity` pass. Their golden stream
  captures are unchanged, since no fixture is added in this task. D-T1a's oracle
  stays green **unmodified**: it is restricted to the frozen 20.
* AC-D3.5 — Stash B72E9715 is cited in the commit body.

### D-T4 — 034.007-T: Record the residual evasion surface honestly

Size **S**, complexity **low**. Documentation only.
`scripts/check-write-path-precondition.sh` (header comment) + `writepath.go`
(package doc comment). Depends on D-T3.

Update the wrapper's "Detector scope" list to the 26 selectors and record —
never silently leave unhandled — the residual surface (AC-4.7):

1. **Named import aliases — KNOWN OPEN, pending Unit E (049-F).** An aliased
   import of a write-capable package (`import o "os"` → `o.WriteFile(…)`, or
   an alias of `database/sql` / `go.etcd.io/bbolt`) is **not detected**.
2. **`pathsafe.NewRoot` receiver / `Root.Resolve` first-caller tracking — KNOWN
   OPEN, pending Unit E.** No tripwire exists. The pathsafe risk-register
   triggers ("forced the moment a real write path exists", `root.go`) and
   feature 038-F's "once a live caller exists" trigger stay **awaited, not
   monitored**.
3. Dot-imports, blank imports, and local identifiers shadowing a package name
   (including a local `syscall` identifier, which the D-2′ predicate trusts by
   spelling).
4. `os.Root` **method** calls and `(*os.File).Write*` (the latter already
   recorded).
5. **Undecidable or non-simple call shape → rejected.** A `syscall.CreateFile`
   in any of these shapes is **rejected, not exempted** (D-2′ rules 2–7):
   * reached via a wrapper or function value;
   * with an extent that does not balance;
   * with any argument that is not a bare operand (call, composite literal,
     index, string or raw string);
   * with any access, disposition or flags spelling outside D-2′'s exact
     tokens.

   This is a **false-positive** surface. A future legitimate metadata-only call
   in such a shape trips the gate and needs an explicit, reviewed widening. It
   is never a silent hole.
6. **Write primitives outside the selector set.** These are not detected:
   * `ioutil.WriteFile`, `ioutil.TempFile` and `ioutil.TempDir`;
   * `syscall` write and namespace calls other than `syscall.CreateFile` and
     `syscall.Write`: `syscall.WriteFile`, `Open`, `Unlink`, `Rename`, `Mkdir`,
     `CreateHardLink` and `DeleteFile`;
   * `golang.org/x/sys/windows` and `golang.org/x/sys/unix` equivalents.

   None occurs in `internal/**` or `cmd/**` at `9b299c8`. Widening the selector
   set is **out of D-031-2's scope** and is tracked as a stash candidate.
7. **Selector split by a newline or comment — KNOWN OPEN, closed by Unit E's
   AST engine.** Go inserts no semicolon after `.`, so both of these are valid
   Go that gofmt preserves:
   * `syscall.` + newline + `CreateFile(p, GENERIC_WRITE, …)`;
   * `os./**/WriteFile(…)`.

   After masking, neither contains the contiguous selector text, so **every**
   selector, not only the new ones, is evaded. This predates the plan and is
   recorded here because the risk statement must not understate the fail-open
   set. It is pinned as a reject fixture in E-T4.

**Residual-risk statement (required text, verbatim intent).** "Until feature
049-F ships, this gate is a **qualified-selector tripwire, not a complete
mechanical proof**. Items 1, 2, 6 and 7 are **known open fail-open surfaces**.
At `9b299c8` there are zero aliased write-capable imports, zero production
`Root.Resolve` callers and zero item-6 primitives in `internal/**`/`cmd/**`, so
items 1, 2 and 6 are not exploited today. None of the four is mechanically
guarded. The compensating control is human and agent PR review against this
list."

**Acceptance criteria**
* AC-D4.1 (AC-4.7) — All seven items and the residual-risk statement are present
  in **both** locations. Items 1 and 2 name feature **049-F** / shipment
  **039-S** as the closing unit, and the gate is **not** presented as complete.
* AC-D4.2 — The wrapper's detector-scope list matches `Selectors` exactly.
* AC-D4.3 — No executable line is modified; all gate entry points still pass.

*Ship guidance (not a task AC; it applies during PR review, after D-T4 is
done):* **review-loop terminator** (compound lesson
`2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`). During 031-S PR
review, a bypass already listed here is answered by citing its item and 049-F.
It does not reopen the masked-text engine.

### D-T5a — 034.008-T: Verdict-boundary fixtures for the access-mode allowance

Size **S**, complexity **medium**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T3.

Fixtures:

* `accept-syscall-createfile-metadata.go`: the metadata-only call alone.
* `reject-syscall-createfile-write.go`: the **positive control**. It contains
  the metadata-only call, a `GENERIC_WRITE` call on another line, and one line
  holding an allowed call **and** a writing call.
* `reject-syscall-createfile-evasion.go`: one line per D-2′ rejection shape:
  * access `0` + `CREATE_NEW`;
  * access `0` + `FILE_FLAG_DELETE_ON_CLOSE`;
  * access `0` + an OR-ed flag combination;
  * a brace-composite multi-value argument;
  * a tag-shaped raw-string argument;
  * a function-value reference.
* `reject-tag-shaped-raw-string-expr.go`: a tag-shaped raw string **outside a
  struct field** (`` q := `x:"os.Remove"` ``). It pins the masker behaviour that
  a `Field.Tag`-only `go/ast` walk would silently flip, so it is a parity
  hazard for Unit E's G-3.

**Acceptance criteria**
* AC-D5.1 (AC-4.1, AC-4.3) — Each fixture is self-tested with the expected
  verdict and has a `fixture_findings` golden row pinning the **exact** finding
  text (selector name and line). The positive-control row pins findings
  **only** on the writing lines, and none on the metadata-only line. That is
  the committed proof the allowance is per-occurrence, not file-level.
* AC-D5.2 — Falsifiability is **committed, not demonstrated once**. Each golden
  row names its selector, so removing that selector, disabling the predicate
  (accept fixture), or widening it (evasion fixture) fails
  `TestScanFile_FixtureFindings_MatchGolden`. Declared honestly (H-4): the
  fixtures are **green on arrival** against D-T3; their red phase lives in
  D-T2/D-T3.
* AC-D5.3 (AC-4.9) — All fixtures are gofmt- and goimports-clean (the lint job
  formats `testdata`).
* AC-D5.4 — `self_test` and `self_test_integrity` stream captures are updated in
  the same commit. D-T1a's oracle stays green unmodified, and all gate entry
  points pass.
* AC-D5.5 (AC-4.5) — No TOCTOU/hard-link mitigation and **no write primitive**
  is introduced outside `testdata` (INV-3, a named stop condition).
* AC-D5.6 — Each fixture's header comment states it is part of the
  **verdict-parity corpus for the 049-F `go/ast` migration**.

### D-T5b — 034.011-T (NEW): Presence fixtures for the new selectors

Size **S**, complexity **low**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T5a (serialised only
because both edit the golden).

Split out of D-T5 at plan-review rev-4 (2-hour rule). Fixtures:

* `reject-syscall-write.go`;
* `reject-os-openroot.go`;
* `reject-os-root-type.go`: an `os.Root`-typed declaration or parameter **with
  no `os.OpenRoot` call**;
* `reject-io-copyn-copybuffer.go`: one B72E9715 selector per line, two golden
  findings.

**Acceptance criteria**
* AC-D5b.1 — Each fixture is rejected and has a golden row pinning its exact
  finding(s). Stream captures are updated in the same commit.
* AC-D5b.2 — Declared **green on arrival**; committed falsifiability via the
  selector-naming golden rows (as AC-D5.2).
* AC-D5b.3 — gofmt/goimports clean; each header comment marks it as 049-F parity
  corpus; stash B72E9715 is cited in the commit body.

### D-T6 — 034.009-T (NEW): Regression fixtures for pre-existing selectors (8E9F8E55)

Size **S**, complexity **trivial**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T5b (serialised only
because both edit the golden).

Add `reject-os-mkdirtemp.go`, `reject-os-chown.go`, `reject-os-lchown.go` and
`reject-os-chtimes.go`. 8E9F8E55 names all four. The operator's summary named
only three, but directed that the stash entry be folded in. `os.Chtimes` is
included because the entry's own text lists it and it shares the gap.

It is flagged for operator confirmation. If the operator declines, Ship drops
the fourth fixture; no other task is affected.

*Fixture counting (2-hour rule; applies to D-T5a, D-T5b and D-T6):* a fixture
data file plus its golden row and stream-capture lines count as **one**
artifact and **one** scenario of a single table-driven golden test. These
tasks are therefore two artifacts: the fixture set and the golden.

**Acceptance criteria**
* AC-D6.1 — Each fixture is rejected and has a golden row pinning its exact
  finding. Stream captures are updated in the same commit.
* AC-D6.2 — Declared **green on arrival** (the selectors predate this unit).
  These are regression/parity fixtures whose committed falsifiability is the
  selector-naming golden row (as AC-D5.2).
* AC-D6.3 — gofmt/goimports clean; each header comment marks it as 049-F parity
  corpus; stash 8E9F8E55 is cited in the commit body.

### Moved out of Unit D

| Task | Disposition |
|---|---|
| 034.005-T `Root.Resolve` tripwire | Removed from 031-S via `shipment return-blocked`, re-parented to 049-F via `backlogit adopt`, rewritten as E-T5 |
| 034.006-T named import aliases | Same path, rewritten as E-T4 |

---

## Unit E — 049-F / 039-S (NEW): Migrate the write-path scanner to `go/ast` and close the alias and `Resolve` residuals

**Gating.** (1) `blocks` edge **039-S → 031-S**: Unit E cannot be claimed before
Unit D ships, because Unit D's fixtures are its parity corpus. (2) **STAGE HOLD**
on 049-F: E-T2..E-T6 are **provisional** and are re-planned against the spike
findings, then re-gated through plan-review, before the hold lifts.

**Why a hold and not just an edge.** Option B **invalidates** the original
spike's mechanism decisions (deliberation §4.2), so the spike that replaces them
must exist **before** implementation tasks are final. Spikes and their findings
are Stage artifacts (P-010: Ship may not create or modify spike or plan
artifacts), so the investigation is **Stage-executed** under the P-016
spike-worktree exception; the in-shipment task E-T1 is the read-only gate that
verifies it, mirroring the original 034.001-T.

**Invariant carried by every Unit E task — gomask is retained.**
`tools/gatecheck/internal/gomask` **must not be deleted or narrowed**:
`retiredarch` depends on it independently of `writepath`. Unit E may remove
`writepath`'s own import of it, nothing more.

### E-T1 — 049.001-T: Go-era `go/ast` spike (Stage-executed) and its verification gate

Size **XS** (Ship's verification half), complexity **trivial**. Read-only.
Depends on 034.009-T (the last task of the Unit D corpus chain; the full
corpus must exist on `main`).

Stage produces `docs/decisions/{date}-intercom-go-writepath-go-ast-spike.md`
answering, with evidence against the live corpus:

* **G-1 Feasibility and boundary** — `go/parser` + `go/ast` reproduce every
  finding's text, line and order across the full corpus.
  * Define a **source-level detection boundary**: it takes the decoded,
    **unmasked** `pysem.ReadText` output and returns findings **plus an
    error**. The masked-text `scanText` signature can carry neither valid Go
    syntax nor parse errors.
  * Decide how `retiredarch/writepath_mask_test.go`'s writepath half is
    re-anchored without weakening `retiredarch`'s canonical-mask and
    single-definition protections.
* **G-2 Fail-closed on unparseable input** — **any** `go/parser` error, even
  when a partial AST is also returned, makes that file an `::error::` ED-2
  failure with exit 1. The partial AST is **never scanned and never skipped**,
  and no `AllErrors` partial-scan mode is used.
  * The existing `filebased` inputs (BOM, CRLF, lone-CR, invalid UTF-8) keep
    their verdicts.
  * Positions use `fset.PositionFor(pos, false)`, so `//line` directives cannot
    move findings; a `//line` fixture pins this.
* **G-3 Verdict parity** — every existing fixture, every Unit D fixture and the
  live tree keep their verdicts. **Named hazards:**
  * **Struct tags.** `reject-struct-tag-selector.go` and
    `reject-tag-shaped-raw-string-expr.go` are rejected today **only** because
    the masker leaves tag-shaped raw strings visible. Under `go/ast` these are
    `*ast.BasicLit`s, so a naive `SelectorExpr` walk (or a `Field.Tag`-only
    walk) would flip them to **accept**, a fail-open (INV-2).
  * **Line numbering.** `go/token` counts lines on `\n` only, while the golden
    uses `pysem.SplitLines` boundaries (`\v`, `\f`, `\x1c`–`\x1e`, `\x85`,
    `\u2028`, `\u2029`). Offsets must be mapped through `SplitLines` boundaries
    **of the masked text**.
    * Boundary runes inside comments or strings are erased by the masker, so
      the legacy line number of a later hit is the **masked** line number
      (AC-D1a.2). Parity must reproduce this.
    * Code-position `\f`, `\v` and `\u2028` are **not Go whitespace**, so they
      are `go/parser` errors and fall under G-2 fail-closed.
    * Verify `go/scanner`'s whitespace and illegal-character handling, and
      `go/token`'s line counting, against the pinned GOROOT source, per compound
      lesson `2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`.
  * **Oracle adaptation.** Confirm that AC-E2.6 items 1–3 suffice, or HALT.
    Any adaptation beyond them returns to Stage and plan-review. The spike may
    not redesign the oracle after seeing where it disagrees (H-3).
* **G-4** — D-2′ mapped onto `*ast.CallExpr` as a **representation change only,
  never a widening**. Every D-2′ rejection is preserved:
  * bare-operand arguments only;
  * exactly 7 `Args`;
  * `Args[1]` an INT `*ast.BasicLit` with value `0`;
  * `Args[4]` exactly `syscall.OPEN_EXISTING`;
  * `Args[5]` exactly `syscall.FILE_FLAG_BACKUP_SEMANTICS`;
  * the `X` of the callee, `Args[4]` and `Args[5]` each resolving to the import
    path `"syscall"`;
  * `call.Ellipsis == token.NoPos`.

  Any allowance the masked engine rejects needs a separately approved policy
  exception with committed regression cases.
* **G-5** — named-alias resolution via `*ast.ImportSpec`. A **dot import** of
  any write-capable package fails closed (a finding), and a blank import is
  inert.
* **G-6** — `NewRoot` receiver binding: syntactic intra-function tracking vs
  `go/types`; cost and false-negative profile of each.
* **G-7** — sizing of E-T2..E-T6 against the 2-hour rule.

**Acceptance criteria (Ship, read-only)**
* AC-E1.1 — The spike document exists and records G-1..G-7 with a decision each.
* AC-E1.2 — The post-spike re-plan gives E-T2..E-T6 each:
  * a citation of the spike;
  * an explicit file list and test-scenario count within the 2-hour rule;
  * at least one verifiable AC;
  * a red/green phase declaration.

  None is left at complexity `high` without a recorded de-risking step.
* AC-E1.3 — If either is unmet, **HALT and return to Stage**; decide nothing.

### E-T2 — 049.002-T: Replace masked-text detection with `go/parser` + `go/ast` (provisional)

Size **M**, complexity **high** (de-risked by E-T1). `writepath.go` +
`writepath_test.go`. Depends on E-T1.

Swap the detection engine in **one commit**. Presence selectors become
`*ast.SelectorExpr` matches on the package identifier. Findings keep the exact
text, line and order. Parse failure fails closed (G-2), the struct-tag rule is
preserved per G-3, and production detection goes through the G-1 source-level
boundary. The masked-text extractor and predicate stay in place for
`syscall.CreateFile` until E-T3. E-T2 owns the G-1 re-anchoring of
`retiredarch`'s writepath scan-loop pin.

**Acceptance criteria**
* AC-E2.1 — The golden JSON is **byte-identical** (`selectors`,
  `fixture_findings`, stream captures, `filebased`). The test harness may be
  re-routed through the G-1 boundary so that the golden tests exercise the
  production engine and not a retained legacy path. The Unit D corpus is the
  parity contract.
* AC-E2.2 — `reject-struct-tag-selector.go` and
  `reject-tag-shaped-raw-string-expr.go` stay **rejected**.
* AC-E2.3 — A committed unparseable-input test, including a partial-AST case,
  fails closed with exit 1, and a `//line` test pins positions.
* AC-E2.4 — `gomask` is not deleted or modified, and `retiredarch` keeps its
  canonical-mask protection.
* AC-E2.5 — Closure evidence includes a before/after byte diff of all three
  gate entry points' output (compound lesson
  `2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`).
* AC-E2.6 — **The single authorised oracle adaptation** (D-T1a lifecycle). In
  the same commit as the engine swap, D-T1a's oracle is adapted as follows:
  1. **Production side.** It is re-pointed to the G-1 source-level boundary and
     fed the original unmasked source. The legacy side stays fed the masked
     text, so the oracle tests the production engine and never a retained
     legacy path.
  2. **Fixture inputs.** The fixture input set is frozen to an **explicit name
     list** equal to the corpus on `main` at 031-S merge. E-T4 and E-T5
     fixtures (alias, dot-import, split-selector, `Resolve`) are therefore never
     oracle inputs.
  3. **Unparseable inputs.** AC-D1a.2's code-position `\f`/`\v`/`\u2028` input
     is moved from a parity assertion to a G-2 **fail-closed assertion**
     (error, exit 1). This is a stricter verdict, never a weaker one.
     Invalid UTF-8 already fails at `ReadText` (AC-D1a.1), so it is unchanged.

  The legacy implementation, the frozen 20 and every other expectation stay
  **unedited**. Any further oracle change HALTs (H-3).

### E-T3 — 049.003-T: Port the allowance predicate onto `*ast.CallExpr` (provisional)

Size **S**, complexity **medium**. `writepath.go` + `writepath_test.go`.
Depends on E-T2.

Re-express D-2′ over the AST (G-4) as a representation change only, and delete
the masked-text extent extractor.

**Acceptance criteria**
* AC-E3.1 — All Unit D predicate tests and fixtures keep their verdicts. The
  golden JSON stays **byte-identical**, and the positive-control row still pins
  only the writing lines.
* AC-E3.2 — Every D-2′ rejection is preserved (INV-2). In particular, only an
  INT `*ast.BasicLit` `0` access, an exact `syscall.OPEN_EXISTING` disposition
  and the exact `syscall.FILE_FLAG_BACKUP_SEMANTICS` flag are allowed, on a non-variadic call whose `X` resolves to
  import path `"syscall"`. Any widening HALTs (H-2 E).

### E-T4 — 049.004-T (re-parented 034.006-T): Resolve named import aliases (provisional)

Size **M**, complexity **medium** (re-sized after E-T1). Depends on E-T2.

Map **named** aliases to canonical import paths via `*ast.ImportSpec` and match
selectors on the canonical path. Dot-imports and blank imports follow the G-5
disposition; shadowing stays a residual.

**Acceptance criteria**
* AC-E4.1 — New fixtures have golden rows:
  * `reject-alias-os-writefile.go` (aliased `os`);
  * `reject-dot-import-os.go` (G-5);
  * `accept-alias-copilot.go` (the live `copilot` alias shape);
  * `reject-split-selector.go`: a selector split by a newline after `.` and by
    an inline comment. It pins D-T4 residual item 7, which the E-T2 AST engine
    closes. It lives here because E-T2/E-T3's golden must stay byte-identical.
* AC-E4.2 — The golden changes are **additive only**: new rows and the
  matching new `PASS` lines in the stream captures, updated atomically with the
  fixtures. Every pre-existing `selectors` and `fixture_findings` row stays
  unchanged.

### E-T5 — 049.005-T (re-parented 034.005-T): `Root.Resolve` first-caller tripwire (provisional)

Size **M**, complexity **medium**. Depends on E-T4 (the tripwire must resolve
`internal/pathsafe` through aliases).

Fire only when a file imports `internal/pathsafe` **and** calls `.Resolve(` on a
receiver bound from `pathsafe.NewRoot` (binding strategy per G-6). A bare
`.Resolve(` match is insufficient.

**Acceptance criteria**
* AC-E5.1 (AC-4.4) — Fires on a fixture introducing a production
  `Root.Resolve` caller; does **not** fire on the tracked tree (zero production
  callers, re-measured at the parent commit).
* AC-E5.2 — Does not fire on an unrelated `.Resolve(` receiver.
* AC-E5.3 — Golden changes are additive only (as AC-E4.2).

### E-T6 — 049.006-T: Retire closed residuals from the gate documentation

Size **XS**, complexity **trivial**. Documentation only. Depends on E-T3 and
E-T5.

Remove items 1, 2 and **7** from D-T4's residual record in both locations,
plus item 3's dot-import clause if G-5 closed it. Item 7 cites E-T2 and E-T4's
`reject-split-selector.go`. Replace the residual-risk statement with the
post-migration posture. Items 4–6 remain.

**Acceptance criteria**
* AC-E6.1 — Each removed residual item cites the task and fixture that closed it.
  Every retained item is still present verbatim, and AC-D4.2's detector-scope
  check still holds.

---

## Constitution Check (Units D and E)

This maps Units D and E against `.github/instructions/constitution.instructions.md`.
Units A–C were checked in revisions 1–3.

| Principle | How Units D/E comply | Justified deviation |
|---|---|---|
| I. Safety-First Go | Gatecheck tool code only. INV-4 (lint clean, no dead code). The D-T1 cursor fails closed with no panic. E-T2 parse errors are exit 1. | None |
| II. Test-First (NON-NEGOTIABLE) | Red/green phases are declared per task (H-4). D-T2 has a genuine red phase (AC-D2.4). D-T1a, D-T5a, D-T5b and D-T6 are **declared** green on arrival as characterization or regression fixtures, with committed falsifiability via selector-naming golden rows. | Green-on-arrival fixtures are disclosed, not claimed as red phases |
| III. Workspace Isolation | No write primitive outside `testdata` (INV-3, AC-D5.5). The gate *is* the workspace-write tripwire. Units D/E only narrow what it allows (INV-2). | None |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | The residual-risk statement (D-T4) names every known open fail-open surface: items 1, 2, 6, 7. Deferral of alias and `Resolve` coverage is operator-approved (D-031-2) and bounded by 039-S, with the H-5 re-measure. | Interval fail-open residuals are accepted by the operator, visible, and not hidden |
| V. Structured Observability | Finding text is unchanged (`write primitive '<sel>' found`). ED-2 `::error::` lines are retained. | None |
| VI. Single Responsibility | Each task covers one skill domain: code, docs or fixtures. Scenario counting is stated (table-driven = 1). | None |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | No backlog item deleted. `return-blocked`/`adopt` were operator-directed (D-031-2) and recorded with an ID map (H-6). `WRITE_PATH_GATE_ADVISORY` stays operator-only (H-3). | None |
| VIII. Explicit Safety Modes | `WRITE_PATH_GATE_ADVISORY` is the explicit elevated-risk lever. Ship must not use it (H-3). | None |
| IX. Git-Friendly Persistence | Golden JSON and fixtures are committed; golden changes are atomic with fixtures. | None |
| X. Agent Context Efficiency | Tasks are S/XS within the 2-hour rule. Unit E stays provisional until the spike, so no speculative detail. | None |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | One commit per task; rollback by revert in reverse dependency order (H-6). No squash or rebase is implied. | None |

## Dependency graph and execution order

```
Unit A (030-S):  A-T1 → A-T2               (ALP-1: one atomic landing, one commit)

Unit B (032-S):  B-T1 → B-T2 → B-T3        (B-T1 is the test-first red)

Unit C (new):    C-T1 → C-T2 → C-T3        (C-T2 is the test-first red)
                 C-T1 → C-T4
                 C-T5  (independent)
                 C-T6  (independent)

Unit D (031-S):  D-T1a → D-T1 → D-T2 → D-T3 ─┬→ D-T4                (docs)
                                             └→ D-T5a → D-T5b → D-T6 (golden-serialised)
                 (034.010 → 034.002 → 034.003 → 034.004 → {034.007, 034.008};
                  034.008 → 034.011 → 034.009)
                 D-T0 (034.001-T) SATISFIED — detached from the chain

Unit E (039-S):  [039-S blocks-on 031-S]   [049-F depends-on 034-F]
                 D-T6 → E-T1 → E-T2 → E-T3 ─┐
                                E-T2 → E-T4 → E-T5 ─┴→ E-T6
                 (034.009 → 049.001 → 049.002 → 049.003;
                  049.002 → 049.004 → 049.005; 049.003 + 049.005 → 049.006)
```

**Recommended shipment order: 038-S (C) → 032-S (B) → 030-S (A) → 031-S (D) →
039-S (E).**

* **Unit C first** — the only unit closing a *live* security-adjacent gap
  (`check-write-path-precondition.sh` has no `--root` guard today), and C-T1
  touches shared CLI dispatch that both other units' engines run through. Landing
  it first means A and B develop against the final dispatch contract.
* **Unit B second** — self-contained in one package, no design fork, lowest risk.
* **Unit A last** of the re-planned trio — the only unit that must mutate `pin.go` in lockstep with the
  code the pin asserts against (formalised as **ALP-1**, INV-5); benefits from a
  quiet tree, since its interior boundary is the plan's only sanctioned red gate.
* **Unit D (031-S) fourth** — independent of A–C at the file level (only
  `writepath` and its fixtures), but it benefits from C-T1's final dispatch
  contract and C-T2/C-T3's write-path wrapper guard, which share
  `check-write-path-precondition.sh`. D-T4 edits that wrapper's **header
  comment** only; C-T3 edits its **body** — sequencing C first avoids a textual
  conflict.
* **Unit E (039-S) last, and gated** — hard `blocks` edge on 031-S plus a STAGE
  HOLD pending the Stage-executed spike. The alias and `Resolve` residuals stay
  open for that interval (H-5); keep 039-S directly behind 031-S.

This supersedes the earlier post-M4 suggestion (032-S → 030-S → 031-S), which
predated both Unit C's discovery and the 031-S option fork.

---

## Plan Hardening

### H-1 — Blast radius

| Unit | Surface | Worst case |
|---|---|---|
| A | `retiredarch` scope declaration + its own pin | **Vacuously green pin.** Emptying `pathspecPinLiterals`/`prefixPinLiterals` to make the refactor land yields `containsAll(_, []) == true` (`pin.go:127`) — the pin passes while asserting nothing. Guarded by AC-A2.2. Secondary: ALP-1 split across two commits, leaving the pin gate red on `main`. Guarded by INV-5 + AC-A2.6. |
| B | One function + its tests | Over-strict error converts a legitimately absent `.gitignore` into a merge-blocking failure. Bounded to one gate; guarded by AC-B2.2. |
| C | `parseRoot`, shared by **all four** engines | A regression breaks every gate at once. Guarded by AC-C1.3's four real self-test entry points. |
| D | `writepath` scanner — the **only** mechanical enforcement of the write-path precondition (merge-blocking at `ci.yml:333-347` unless the operator sets `WRITE_PATH_GATE_ADVISORY`) plus its fixtures/golden. No product code. | **Fail-open:** D-T1's per-occurrence enumeration drops or mis-lines a hit, or D-T2's predicate allows a writing or file-creating `syscall.CreateFile` (non-zero access, non-`OPEN_EXISTING` disposition, `DELETE_ON_CLOSE`, smuggled composite/raw-string arguments, or a same-line pairing) — a silent hole in a security gate. Guarded by AC-D1a.1 (frozen differential oracle over corpus and live tree), AC-D2.2/AC-D2.3 (rejection table and same-line test), AC-D5.1 (fixture-level positive control and evasion corpus), and the golden. **Fail-closed:** line/order drift reddens golden or the repo scan — loud, blocks merges until fixed. **Collateral:** `retiredarch`'s writepath scan-loop pin — kept unmodified by D-1′ (AC-D1.4). |
| E | Same scanner, **engine swap** | **Fail-open:** the tag-shaped-raw-string rule, a partial-AST-on-parse-error scan, `//line`/`\n`-only line drift, or a predicate *widening* during the port flips a reject to accept. Guarded by G-2/G-3/G-4, AC-E2.1–E2.3, AC-E3.2 and the Unit D parity corpus. Collateral: deleting `gomask` would break `retiredarch` — forbidden by the Unit E invariant; the `retiredarch` writepath pin is re-anchored by E-T2 (G-1). |

### H-2 — Named stop conditions

Halt and return to Stage rather than deciding:
* **A** — if `scanScope` cannot express all three arms (including the
  prefix-less `config.toml.example`) without a special case; or if AC-A2.3's
  negative control cannot be made to fail; or if re-anchoring the pin would
  require deriving the expected literals **from `select.go`'s own text** (which
  *would* make it a true self-comparison). Do **not** empty the pin literal lists
  and do **not** weaken the pin to make the refactor land.
* **B** — if `git show`'s exit/stderr shape cannot distinguish invalid-ref from
  absent-file portably. Do **not** satisfy AC-B2.1 by making absent-file an error.
* **C** — if first-wins `parseRoot` changes any existing engine's verdict or exit
  code. Do **not** proceed by relaxing AC-C1.3.
* **D** — if D-1′ or D-2′ cannot be implemented exactly as specified (AC-D1.5,
  AC-D2.6); if D-T1a's frozen oracle (restricted to the frozen 20 selectors)
  shows **any** divergence; if `retiredarch`'s writepath pin would need editing;
  or if adding the six selectors makes the bare repo scan fail. Do **not** edit
  the oracle or golden rows for pre-existing fixtures, do **not** widen any D-2′
  token or allowlist, and do **not** attempt alias or receiver tracking in this
  unit (that is Unit E's scope).
* **E** — if the spike cannot establish fail-closed parse handling (including
  partial ASTs) or full verdict parity (including both tag-shaped fixtures and
  `SplitLines` line numbering); if E-T2/E-T3 would change **any** golden byte, or
  E-T4/E-T5 would change any **pre-existing** golden row; if the predicate port
  would allow anything D-2′ rejects; or if the migration would require deleting
  or modifying `gomask`.
* **All** — if any task would introduce a write primitive into `internal/**` or
  `cmd/**` (INV-3), or would leave the tree lint-dirty (INV-4), or would require
  an atomic landing pair not declared in INV-5.

### H-3 — Controls that must not be weakened to make work land

* **AC-A2.2** (pin literals non-empty) and **AC-A2.3** (committed negative
  control) exist specifically because the refactor's easiest failure mode is a
  green pin that pins nothing. Neither may be relaxed.
* **AC-B1.2** (inverted assertion must fail pre-change) exists because the unit
  could otherwise be reported complete while `git_test.go:103-121` still pins
  fail-open behaviour.
* **AC-C5.3** (C-T5 changes no behaviour) exists because "fixing" item (3) would
  change the verdict of a currently-passing invocation shape without an
  authorising acceptance criterion (INV-1's explicit-AC carve-out).
* **INV-5 / AC-A2.6** (ALP-1 lands as one commit) exists because the cheapest way
  to make A-T1 "pass" on its own is to weaken or empty the pin. The pair boundary
  is where the pin must be proven, and it may not be moved earlier.
* **AC-D1a.1** (frozen differential oracle) exists because "line numbers
  unchanged" is otherwise asserted only for the handful of golden fixtures, and
  the `SplitLines`-vs-`\n` hazard would pass them while mis-lining a live
  finding. It is restricted to the frozen 20 selectors (exact set lookup) so
  that it is never edited in Unit D; its **only** permitted adaptation is
  AC-E2.6 (production side re-pointed to the G-1 boundary, input list frozen by
  name, unparseable inputs moved to stricter fail-closed assertions) — an oracle
  that is edited whenever it disagrees is no oracle.
* **AC-D2.2 / AC-D2.3 / AC-D5.1** (rejection table, same-line test, positive
  control pinning only the writing lines) exist because a file-level or
  first-occurrence exemption passes every "metadata call is allowed" test.
* **AC-D1.4** (`retiredarch` pin unmodified) exists because editing a
  cross-package architectural guard to make a refactor land is exactly the
  weakening H-3 forbids.
* **AC-D4.1** (residual-risk statement names 049-F) exists because deferring
  alias and receiver coverage is only acceptable if it is **visible** to every
  reviewer of the gate in the interval.
* **AC-E2.1 / AC-E3.1** (golden byte-identical across E-T2/E-T3) and
  **AC-E4.2 / AC-E5.3** (additive-only thereafter) are the parity contract;
  editing a pre-existing golden row to make the migration land defeats the Unit D
  corpus.
* **`WRITE_PATH_GATE_ADVISORY`** — Ship must **not** set it, request it, or rely
  on it to land any Unit D or Unit E work; a D/E PR that is green only because
  the repo-scan step was advisory is not green. It is an operator-only
  emergency lever (H-6). *Recommendation (requires operator approval before it
  becomes policy):* the operator logs each use with a restore condition and
  re-scans any commit merged while it was set with the bare gate before
  clearing it.

### H-4 — Green-on-arrival declared honestly

* **A-T1's derivation-agreement property is green-on-arrival** — the two
  derivations already agree today. The genuinely new assertions are AC-A1.5
  (asymmetry preservation) and A-T2's AC-A2.3 (negative control), which are red
  against a naive flat-prefix unification.
* **A-T2's re-anchoring half is red** (the pin breaks the moment A-T1 moves the
  literals, by design — `select.go:5-9`); its non-vacuity guard (AC-A2.2) is
  green-on-arrival today and exists to stay green.
* **B-T1 is the red phase** for Unit B; B-T2 turns it green.
* **B-T3, C-T4, C-T5 and C-T6 are documentation-only** and claim no red phase.
* **C-T1 has a genuine red phase** (test-first on `main_test.go`).
* **C-T2 is the red phase** for the wrapper guard; **C-T3 is the green phase.**
  Neither is green-on-arrival: AC-C2.3 requires C-T2's rejection cases to fail
  against the pre-C-T3 wrapper, and AC-C3.4 requires C-T3 to turn them green.
  C-T3 must therefore be verified, not assumed.
* **D-T1a is green on arrival by design** — a characterization oracle that
  pins today's behaviour. **D-T1's** new extent and cursor assertions (AC-D1.1–
  AC-D1.3) are red before the extractor exists. **D-T2 has a genuine red phase**
  (AC-D2.4). **D-T3's red** is the golden `selectors` edit (AC-D3.1).
  **D-T4 is documentation-only.** **D-T5a, D-T5b and D-T6 fixtures are green on
  arrival** — declared, with committed falsifiability via selector-naming golden
  rows (AC-D5.2, AC-D5b.2, AC-D6.2).
* **Unit E** red/green phases are declared by the post-spike re-plan; E-T1 is
  read-only.

### H-5 — Residuals carried, not closed

* `DD0BB60F` — `pin.go:76` checks literal *presence in body*, not *argument
  position*. AC-A2.3 narrows but does not close it.
* `C312BD4C` — the `gomask` multiline-tag contradiction (`gomask.go:138-141`
  comment vs `gomask_test.go:57` pinned behaviour) stays open; deliberation
  D-030-3 re-scopes it to a comment fix.
* `50E6F22C` item (3) — extra-positional-argument tolerance stays open pending a
  new ED entry (AC-C5.3).
* `5A8EC1BC` — three `GitRunner` shapes and exported mutable
  `Denylist`/`Selectors` stay open; trigger not fired.
* **Deferred alias coverage (D-031-2) — KNOWN OPEN fail-open surface from 031-S
  ship until 039-S ship.** An aliased or dot import of any package whose
  qualifier appears in `Selectors` (`os`, `io`, `syscall`, `database/sql`,
  `go.etcd.io/bbolt`) evades the gate. *Likelihood:* low — zero instances at `9b299c8`; the sole alias in scope
  is `copilot` (not write-capable), and Go convention rarely aliases those
  packages. *Impact:* high — a write path could land without tripping the
  precondition, so the pathsafe TOCTOU/hard-link register triggers would fire
  unnoticed. *Compensating controls:* AC-D4.1's residual record in the gate's own
  header, PR review, and E-T4.
* **Deferred `NewRoot`/`Resolve` tracking (D-031-2) — KNOWN OPEN over the same
  interval.** The first production `Root.Resolve` caller lands unmonitored, so
  feature 038-F's trigger and the register's "real write path" triggers stay
  *awaited*, not *monitored*. *Likelihood:* medium — product work (043-F) will
  eventually add callers. *Impact:* medium — 038-F's black-box verification and
  the TOCTOU mitigation decision are skipped until someone notices.
  *Compensating controls:* the residual record, PR review, and E-T5. 038-F
  now carries a `blocks` edge on 049-F (Harvest Record), so its trigger cannot
  be mistaken as satisfied by 031-S.
* **Split selector (D-T4 item 7) — KNOWN OPEN until E-T2** (pinned by E-T4's
  `reject-split-selector.go`). Pre-existing; affects all selectors. *Likelihood:*
  very low (gofmt-hostile, never idiomatic). *Impact:* high. *Compensating
  controls:* the residual record and PR review.
* **Selector coverage gap (D-T4 item 6) — KNOWN OPEN, not scheduled.**
  `ioutil.*` write helpers, `syscall` write/namespace calls beyond the two
  selected, and `golang.org/x/sys` equivalents are undetected. *Likelihood:*
  low (none at `9b299c8`; `ioutil` is deprecated). *Impact:* high. Widening the
  selector set is outside D-031-2's operator-approved scope; it is recorded as a
  stash candidate for a later operator decision.
* **Interval bound and re-measurement (recommendation — requires operator
  approval before it becomes policy).** Neither deferred residual has a time
  bound if 039-S stalls. Stage recommends re-measuring at each Stage session
  until 039-S ships, recording command and verbatim output in the session
  memory:
  `git grep -nE '^[[:space:]]*(import[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*|\.)[[:space:]]+"(os|io|syscall|database/sql|go\.etcd\.io/bbolt)"' -- 'internal/**' 'cmd/**' ':!*_test.go' | grep -v 'return "'`
  (package set derived from the `Selectors` qualifiers; trailing comments
  tolerated; the `grep -v` drops `return "io"`-style string literals such as
  `internal/apperr/apperr.go:67`) and
  `git grep -nE '\.Resolve\(|pathsafe\.NewRoot\(' -- 'internal/**' 'cmd/**' ':!*_test.go' ':!internal/pathsafe/**'`
  and, for item 6,
  `git grep -nE '(ioutil\.(WriteFile|TempFile|TempDir)|syscall\.(WriteFile|Open|Unlink|Rename|Mkdir|CreateHardLink|DeleteFile)|golang\.org/x/sys)\b' -- 'internal/**' 'cmd/**' ':!*_test.go'`
  (POSIX `[[:space:]]`, not GNU `\s`, per compound lesson
  `2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`; the
  item-6 command's only current hit is the comment at
  `internal/pathsafe/reparse_windows.go:15`, re-verified at `2c05b6e`).
  At `9b299c8` the first is empty and the second lists only `internal/config/validate.go`
  (`pathsafe.NewRoot` calls at `:42`/`:67` and a comment at `:32`); any new hit is a
  trigger to promote 039-S to the head of the queue. A blocking CI tripwire was
  considered and **not** adopted: it would re-implement the deferred 034.006-T /
  034.005-T scope in text form, contrary to D-031-2.

### H-6 — Rollback

* **Units A–C** — revert the task commit (ALP-1 as its single commit). No data
  or schema state is involved.
* **Unit D** — every task is its own commit; **revert in reverse dependency
  order** (D-T6 → D-T5b → D-T5a → D-T4 → D-T3 → D-T2 → D-T1 → D-T1a). The one
  unsafe partial revert is **reverting D-T2 while D-T3 remains**:
  `syscall.CreateFile` would then be a selector with no allowance, and the bare
  repo scan fails on `reparse_windows.go:164` — fail-closed, but merge-blocking.
  Revert D-T3 with or before D-T2. Emergency lever (operator-only, not a Ship
  action): the `WRITE_PATH_GATE_ADVISORY` repository variable makes the repo-scan
  step advisory while the integrity self-test stays blocking (`ci.yml:333-347`).
* **Unit E** — roll back by **landed dependency closure**: revert E-T6, E-T5,
  E-T4, E-T3, then E-T2 (with their fixtures, golden additions and the
  `retiredarch` re-anchoring). Only at the immediate E-T2 boundary — before E-T3
  deletes the masked-text extractor — is a single revert sufficient. `gomask` is
  retained so the masked engine can be restored cleanly; the Unit D corpus is
  the parity contract in **both** directions. `WRITE_PATH_GATE_ADVISORY` does
  not bypass the blocking integrity self-test, so it is not a substitute for
  rollback.
* **Backlog** — `shipment return-blocked` removal is reversible with
  `shipment add 031-S <id>` while the item is in no other shipment.
  `backlogit adopt` **renumbers** IDs (034.006-T → 049.004-T, 034.005-T →
  049.005-T); re-adopting back to 034-F would renumber again, so the ID map in
  the deliberation (§4.5) is the rollback record. **Harvest read-back
  obligation** (compound lessons `2026-09-04-backlogit-sizing-is-wit-gated.md`,
  `2026-05-07-backlogit-shipment-status-constraints.md`): every backlog mutation
  is run with stderr visible and its exit code checked, then read back
  (`shipment get`, `dep list`, item files); the resulting ID map and edge
  directions are recorded in the session memory.

---

## Verification

Gate scripts must be invoked at their **real entry points**. Neither
`check-unignore-regression.sh` nor `check-gitignore-append-only.sh` may be run
bare — both fail closed with exit 1 when no ref arguments are supplied.

```bash
go build ./...
go test ./tools/gatecheck/... -count=1
golangci-lint run ./...
staticcheck ./...

scripts/check-retired-architecture.sh
scripts/check-retired-architecture.sh --self-test
scripts/check-retired-architecture.sh --self-test-integrity

scripts/check-write-path-precondition.sh
scripts/check-write-path-precondition.sh --self-test
scripts/check-write-path-precondition.sh --self-test-integrity

scripts/check-merge-strategy.sh --self-test
scripts/check-merge-strategy.sh

scripts/check-unignore-regression.sh --self-test
scripts/check-unignore-regression.sh --base-ref <base> --head-ref <head>

scripts/check-gitignore-append-only.sh --self-test
scripts/check-gitignore-append-only.sh --base-ref <base> --head-ref <head>
```

Each unit additionally verifies its selected-path / verdict set against **its own
parent commit**, per INV-1.

**ALP-1 exception (INV-5).** The full suite above is expected to be red at the
A-T1 → A-T2 interior boundary, bounded to the tests exhaustively enumerated in
the ALP-1 declaration. Run the full suite at the **pair boundary** (AC-A2.6),
not at A-T1.

**Test-first red phases (not INV-5 pairs).** `go test ./tools/gatecheck/...` is
also expected red at the B-T1 → B-T2 and C-T2 → C-T3 boundaries, bounded to the
assertions those tasks author (AC-B1.2, AC-C2.3). No *gate script* changes
verdict at either boundary, and both land as separate commits.

**Golden/fixture coupling (Units D and E).** `writepath_golden.json` pins
`selectors`, every `fixture_findings` row, and the `--self-test` /
`--self-test-integrity` stream captures (one `PASS` line per fixture). Any
selector or fixture addition (D-T3, D-T5a, D-T5b, D-T6) **must** update the golden in
the same commit, or `go test ./tools/gatecheck/...` goes red. E-T2 and E-T3 must
leave the golden **byte-identical** (AC-E2.1, AC-E3.1); E-T4 and E-T5 may only
**add** rows and their stream-capture `PASS` lines (AC-E4.2, AC-E5.3). D-T2's red phase is a `go test` red
only; no gate script changes verdict at that boundary because the predicate is
dormant until D-T3 adds `syscall.CreateFile`.

---

<!-- plan-review-attempt: 3 -->
<!-- revision: 3 — round 2 (FAIL) remediations: P1-1 ALP-1/INV-5, P2-1 AC-A1.7,
     P3-1 C-T2/C-T3 inversion, P3-2 B-T2 interim note, P3-3 AC-C5.3 rationale,
     P3-4 citation drift, P3-5 scenario count.
     Round 3 (ADVISORY, 9xP3) residuals all applied: INV-5 test-first exclusion,
     exhaustive ALP-1 red enumeration + AC-A2.7, AC-C2.2 reword, C-T4 fifth
     comment, AC-A1.7 widened to select.go:4-9, H-4 green-phase correction,
     C-T1 cross-reference, frontmatter revision bump, C-T2 sizing note. -->
<!-- GATE: plan-review round 3 = ADVISORY. No P1/P2 outstanding. Proceeds to
     harvest under the ADVISORY disposition with operator awareness. -->

## Plan Review — Revision 4 (Units D and E only)

<!-- plan-review-attempt-rev4: 1 -->

### Round 1 — FAIL

dispatch_mode: multi-agent
decision: FAIL

Six personas were dispatched as subagents. Five ran on the caller's model:
Constitution Reviewer, Go Reviewer, Scope Boundary Auditor, Learnings
Researcher and Security Lens Reviewer. Security Lens was triggered because the
write-path gate is a security control. The sixth, Architecture Strategist, ran
through the `model_routing.anchor_review` route (`gpt-6.1-sol`, xhigh). Plan
hardening was required and present: H-1..H-6.

Merged counts after deduplication: **1 P0, 6 P1**, about 12 P2, and P3s.

| ID | Sev | Source | Finding | Remediation (rev 4, round-1 fixes) |
|---|---|---|---|---|
| R4-P0-1 | P0 | Security | Paren-only depth lets brace composites or tag-shaped raw strings fake D-2's 7-segment shape | D-1′ `()[]{}` stack; D-2′ bare-operand rule; evasion fixture (D-T5a) |
| R4-P1-1 | P1 | Constitution, Go, Security | Access `0` + `CREATE_*`/`OPEN_ALWAYS` creates files; `DELETE_ON_CLOSE` unchecked | D-2′ rules 6–7; table and fixture rejects; G-4 carries them |
| R4-P1-2 | P1 | Constitution, Go | Differential oracle breaks at D-T3/D-T5 (legacy copy has no allowance) | Oracle restricted to a **frozen 20-selector** set; split into 034.010-T |
| R4-P1-3 | P1 | Security | First-occurrence-only `findSelector` lets an allowed call hide a writing call on the same line | Per-occurrence predicate, de-duplication afterwards (AC-D2.3, D-T5a fixture) |
| R4-P1-4 | P1 | Security | G-2 too weak: partial AST scanned on parse error; `//line` shifts positions | G-2 any-error = exit 1, no partial scan, `PositionFor(pos,false)`, `//line` fixture |
| R4-P1-5 | P1 | Architecture | `retiredarch/writepath_mask_test.go` pins `scanText`'s top-level `SplitLines` loop; D-1's whole-text loop breaks it | D-1′ keeps the loop + byte cursor; pin must pass unmodified (AC-D1.4); E-T2 owns re-anchoring (G-1) |
| R4-P1-6 | P1 | Learnings | `go/token` counts `\n` only — line-number hazard for Unit E | Named G-3 hazard; map through `SplitLines`; oracle inputs retained |
| R4-P2-a | P2 | Constitution, Scope, Architecture | "Unit E golden byte-identical" contradicts E-T4/E-T5 fixture additions | Byte-identical for E-T2/E-T3; additive-only for E-T4/E-T5 |
| R4-P2-b | P2 | Constitution, Scope, Go | D-T1 and D-T5 exceed the 2-hour rule | D-T1 → D-T1a + D-T1; D-T5 → D-T5a + D-T5b |
| R4-P2-c | P2 | Go | No fixture pins tag-shaped raw strings outside struct fields | `reject-tag-shaped-raw-string-expr.go` (D-T5a) |
| R4-P2-d | P2 | Architecture, Go | `scanTextWith` is not an engine-swap boundary; masked text cannot parse | Seam dropped from Unit D; G-1 defines a source-level boundary |
| R4-P2-e | P2 | Architecture | E-T3 AST port could widen the allowance | G-4 / AC-E3.2: representation-only; widening HALTs |
| R4-P2-f | P2 | Architecture, Constitution | Unit E rollback by single revert is wrong once E-T3+ land | H-6 rollback by landed dependency closure |
| R4-P2-g | P2 | Security, Constitution | Uncovered write primitives (`ioutil.*`, other `syscall`, `x/sys`) unrecorded | D-T4 residual item 6; H-5 entry; stash candidate |
| R4-P2-h | P2 | Security | `WRITE_PATH_GATE_ADVISORY` could be used to land D/E work | H-3 control: unset while D/E in flight; logged; re-scan |
| R4-P2-i | P2 | Learnings, Security | Re-measure grep too narrow, unrecorded | Broadened grep; recorded output; marked as a recommendation pending operator approval |
| R4-P2-j | P2 | Learnings | D-T0 queued no-op affects closure | Named prerequisite for 031-S closure |
| R4-P2-k | P2 | Learnings | Bypass review loop risk on the 031-S PR | AC-D4.4 review-loop terminator |
| R4-P2-l | P2 | Security | No automated tripwire for the deferred residuals | **Declined, with rationale.** It would re-implement deferred 034.005-T/034.006-T in text form, contrary to D-031-2. The re-measure recommendation stands. |
| R4-P2-m | P2 | Learnings | Backlog mechanics (return-blocked/adopt/edge direction) need read-back | Applied at harvest; tool output recorded in the session memory |

P3s applied:
* findSelector copy;
* the fixed finding text;
* content-located tests;
* `strings.TrimSpace`;
* reverse-order rollback wording;
* the "reverting D-T2 while D-T3 remains" wording;
* the redundant edge D-T5 → E-T1 removed;
* D-T5's dependency on D-T4 replaced with D-T3;
* E-T1 re-plan obligations (AC-E1.2);
* the `Ellipsis` and import-path checks (G-4);
* the dot-import fail-closed rule (G-5).

P3s acknowledged and not applied:
* a header-vs-`Selectors` consistency test (AC-D4.2 already requires the
  match; a test would be new scope);
* a CI `::notice::` for residuals.

<!-- plan-review-attempt: 2 -->

### Round 2 (re-entry 1 of 2) — FAIL

dispatch_mode: multi-agent
decision: FAIL

The same six personas were re-dispatched against the round-1 remediation. The
Architecture Strategist again ran on the `anchor_review` route (`gpt-6.1-sol`,
xhigh).

**Round-1 results:**
* Every round-1 P0/P1 was confirmed **RESOLVED** by the persona that raised it:
  R4-P0-1, R4-P1-1 to R4-P1-6.
* Security Lens found no valid Go that D-2′ exempts while writing.
* The Go and Architecture reviewers confirmed the D-1′ cursor is byte-correct,
  and that the `retiredarch` pin passes **unmodified**.
* The Architecture reviewer confirmed the task DAG is correct and minimal.

**Persona verdicts:**

| Persona | Model | Outstanding P0/P1/P2/P3 | Verdict |
|---|---|---|---|
| Security Lens | caller | 0 / 0 / 3 / 2 | ADVISORY |
| Go Reviewer | caller | 0 / 0 / 3 / 4 | ADVISORY |
| Constitution | caller | 0 / 0 / 1 / 5 | ADVISORY |
| Architecture Strategist | anchor (`gpt-6.1-sol`) | 0 / **1** / 0 / 0 | FAIL |
| Scope Boundary Auditor | caller | 0 / 0 / 1 / 7 | ADVISORY |
| Learnings Researcher | caller | 0 / 0 / 1 / 3 | ADVISORY |

**Round-2 findings and dispositions:**

| ID | Sev | Source | Finding | Remediation |
|---|---|---|---|---|
| R4b-P1-1 | P1 | Architecture (also Go N3 P2, Constitution N-1 P2, Learnings P2) | D-T1a's oracle is bound to masked `scanText`, but G-1 replaces that boundary. Alias fixtures and code-position `\f`/`\v`/`\u2028` inputs (parser errors under G-2) break "stays green unedited". | Oracle lifecycle: legacy side frozen; **single authorised adaptation AC-E2.6** (production side → G-1 on unmasked source; fixture list frozen by name at 031-S merge; unparseable inputs → stricter G-2 fail-closed assertions); G-3 names the hazards; H-3 updated |
| R4b-P2-1 | P2 | Security, Go | Rule 3's quote check cannot see interpreted strings or runes (gomask blanks the delimiters); the AC-D2.2 "string argument" row was unmeetable | New **rule 3b** (non-empty segments); rule 3 reworded; AC-D2.2 rows split |
| R4b-P2-2 | P2 | Go | A substring frozen-20 filter would count `io.CopyN` as `io.Copy` | Exact set lookup on the parsed selector token (D-T1a) |
| R4b-P2-3 | P2 | Security | A selector split by a newline or comment evades every selector; the risk statement understated the fail-open set | D-T4 **residual item 7**; risk statement names items 1, 2, 6, 7; H-5 entry; E-T4 `reject-split-selector.go` |
| R4b-P2-4 | P2 | Scope | D-T2's scenario count exceeds the 2-hour rule | Table-driven = one scenario (stated); D-T2 = 3 scenarios. **Not split**: the rules share one evaluator, and a split would leave a half-predicate commit |
| R4b-P2-5 | P2 | Security | Declining the CI tripwire relies on an unapproved re-measure | **Accepted as open:** the re-measure recommendation is surfaced as an operator decision at hold-lift |
| R4b-P3 | P3 | various | Rule 7 under-specified; G-4 constants' `X`; AC-D2.3 before D-T3; `findSelector` disposition; in-comment boundary case; cursor fail-closed; D-T1 dead code; item-6 grep; deliberation D-T0 mismatch; D-031-3 acknowledgment field; D-T5/D-T6 file counting; `os.Chtimes` confirmation; `FILE_FLAG_OPEN_REPARSE_POINT` speculative; H-3 operator duties; AC-D4.4 not task-verifiable; `[[:space:]]`; D-T0 disclosed deviation; read-back obligation | All applied. Rule 7 is narrowed to exact `FILE_FLAG_BACKUP_SEMANTICS`. H-3's operator logging is now a recommendation. AC-D4.4 is moved to Ship guidance. `os.Chtimes` is flagged for operator confirmation. G-5 dot-import fail-closed is confirmed at re-plan |

<!-- plan-review-attempt: 3 -->

### Round 3 (re-entry 2 of 2, final) — ADVISORY

dispatch_mode: multi-agent
decision: ADVISORY

The personas that owned round-2 findings were re-dispatched: Architecture
Strategist on the anchor route (`gpt-6.1-sol`, xhigh), Security Lens, Go
Reviewer, and Constitution Reviewer. Scope Boundary and Learnings had no
round-2 P0/P1, and their round-2 P2/P3 items were all remediated.

**Round-2 P1 and P2 status:**
* **R4b-P1-1 is RESOLVED**, confirmed by the Architecture reviewer (anchor
  route). The tracked-tree inputs stay live as extra regression coverage, and
  any divergence HALTs.
* Every round-2 P2/P3 is resolved, or accepted as an open operator decision
  (R4b-P2-5).
* The live `reparse_windows.go:164-172` call still passes every D-2′ rule,
  including rule 3b and the narrowed rule 7 (Security Lens).

| Persona | Model | P0 / P1 / P2 / P3 | Verdict |
|---|---|---|---|
| Architecture Strategist | anchor (`gpt-6.1-sol`) | 0 / 0 / 0 / 1 | ADVISORY |
| Security Lens | caller | 0 / 0 / 1 (accepted-open R4b-P2-5) / 2 | ADVISORY |
| Go Reviewer | caller | 0 / 0 / 1 / 3 | ADVISORY |
| Constitution | caller | 0 / 0 / 1 / 4 | ADVISORY |

**Round-3 advisory findings, all applied after the verdict (no re-review
needed at ADVISORY):**
* **Go P2.** The cursor-mismatch rule said "ED-2 `::error::`", which the
  `[]string` signature cannot carry. The rule now disables the allowance for
  the rest of the text, which fails closed as a false positive. The case is
  unreachable through `scanFile`, and a direct-call invalid-UTF-8 test pins it.
* **Constitution P2.** The plan had no Constitution Check section. Added:
  **Constitution Check (Units D and E)**.
* **P3s:**
  * E-T6 retires residual item 7;
  * `findSelector` has a named production caller (fast-path guard);
  * hand-built oracle inputs are complete Go files;
  * invalid UTF-8 handling is moved to `ReadText` both-sides-error;
  * the selector token is parsed from the right;
  * the lifecycle paragraph lists the unparseable-expectations adaptation;
  * the G-3 oracle bullet is now "confirm or HALT";
  * the deliberation's D-031-3 row is updated.

**Gate outcome.** The verdict is ADVISORY with **no P0/P1 outstanding**, which
meets the operator's D-031-2 hold-lift condition. The **031-S STAGE HOLD is
LIFTED**.

The remaining open items are operator decisions, not plan defects:
* acknowledging D-031-3;
* approving the H-5 re-measure (on which declining the CI tripwire rests,
  R4b-P2-5);
* `os.Chtimes` in D-T6;
* the terminal state for 034.001-T and 033.003-T;
* the selector-set widening (item 6).

The 049-F STAGE HOLD stays in place until the Go-era spike re-plan, as designed.

---

## Harvest Record — Revision 4 (2026-10-02)

Every mutation used `C:\Tools\backlogit.exe` v1.11.0 with stderr visible and
exit codes checked, followed by a read-back (`shipment get`, `dep list` and the
item files). This satisfies H-6. All expected IDs were assigned exactly as
planned.

| Plan unit | Backlog ID | Disposition |
|---|---|---|
| D (feature) | 034-F | Retitled "Harden write-path precondition gate (masked-text increment)". The `blocked` status moved to `queued` (**hold lifted**). Edges to 030-F/032-F retained (D-000-1). |
| D-T0 | 034.001-T | Retitled `SATISFIED (post-M4 Go re-plan) …`. Kept `queued` in 031-S with **no live edges** (edge 034.002→034.001 removed). Lifecycle: no queued→terminal transition in `hooks.yaml`, so closure is a named operator prerequisite (shared with 033.003-T). |
| D-T1a | 034.010-T | NEW, S |
| D-T1 | 034.002-T | Rewritten (D-1′), S |
| D-T2 | 034.003-T | Rewritten (D-2′), S |
| D-T3 | 034.004-T | Rewritten (six selectors; folds B72E9715), S |
| D-T4 | 034.007-T | Rewritten (seven residuals + residual-risk statement), S |
| D-T5a | 034.008-T | Rewritten, M→S |
| D-T5b | 034.011-T | NEW, S |
| D-T6 | 034.009-T | NEW, S (folds 8E9F8E55) |
| E (feature) | 049-F | NEW. Status `blocked` = the **STAGE HOLD** pending the spike re-plan. |
| E-T1 | 049.001-T | NEW, XS |
| E-T2 | 049.002-T | NEW, M |
| E-T3 | 049.003-T | NEW, S |
| E-T4 | 049.004-T | **was 034.006-T**: `shipment return-blocked` from 031-S → `move --status queued` → `adopt --parent 049-F`. Body rewritten, M. |
| E-T5 | 049.005-T | **was 034.005-T**: same path as E-T4. Body rewritten, M. |
| E-T6 | 049.006-T | NEW, XS |
| E (shipment) | 039-S | NEW, "Migrate write-path scanner to go/ast (post-M4 unit E)". |

**Final manifests (read back).**
* 031-S, retitled "Harden write-path gate, masked-text increment (post-M4 unit
  D)": `034-F, 034.001-T, 034.002-T, 034.003-T, 034.004-T, 034.007-T,
  034.008-T, 034.010-T, 034.011-T, 034.009-T`. Manifest order is not execution
  order; Ship follows the edges, and 034.010-T is the chain root.
* 039-S: `049-F, 049.001-T … 049.006-T`.

**Edges (read back; `A → B` means A depends on B, type `blocks`).**
* Removed:
  * 034.002→034.001, 034.007→034.006, 034.008→034.005 and 034.008→034.007;
  * the adopt-rewritten 049.004→034.004 and 049.005→034.003.
* Added:
  * 034.002→034.010, 034.007→034.004, 034.008→034.004, 034.011→034.008 and
    034.009→034.011;
  * 049.001→034.009, 049.002→049.001, 049.003→049.002, 049.004→049.002,
    049.005→049.004, 049.006→049.003 and 049.006→049.005;
  * 049-F→034-F and **039-S→031-S**;
  * **038-F→049-F** (see below).

**Consequence found at harvest — 038-F re-pointed.** 038-F ("Black-box verify
`Root.Resolve` once a live caller exists") names "the `Root.Resolve`
first-caller tripwire delivered by 034.005-T" as its trigger, and its only edge
pointed at 034-F. The split moved that tripwire to 049.005-T. Stage therefore:
* added edge 038-F→049-F, keeping 038-F→034-F as historical record;
* appended a marked description section recording the re-identification.

The legacy text and the `blocked` status are unchanged. Without this, 038-F
would read as unblocked once 031-S ships, even though its trigger mechanism
does not exist yet.

**Stash.** 8E9F8E55 → 034.009-T and B72E9715 → 034.004-T (with the
`reject-io-copyn-copybuffer.go` presence fixture in 034.011-T). Each entry's
text was edited to carry the reconciled `PR=#71` and its harvest target. Both
were then **archived** (`backlogit stash archive`, non-destructive).

**Degraded tooling.** `update --complexity` fails ("task does not define a
complexity field"). Complexity is recorded as prose
(`Size: X | Complexity: y`) in every task description. Size is set as a
structured field (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`).

*Generated by Copilot*
