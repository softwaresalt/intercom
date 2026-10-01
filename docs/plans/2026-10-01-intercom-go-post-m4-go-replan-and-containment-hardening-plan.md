---
title: "Plan — Post-M4 Go re-plan (030-S, 032-S) and gatecheck CLI containment hardening"
description: "Retargets held plan units 6 and 8 from the retired Python engines onto tools/gatecheck, and plans a new containment-hardening unit from stash entries 7223218F, 50E6F22C and 44F8CC48"
source_document: docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md
agent: Stage
date: 2026-10-01
stage_branch: chore/stage-post-m4-followups-and-go-replan
bound_snapshot: 9b299c8 (origin/main)
revision: 3
requires_plan_hardening: yes
---

# Plan — Post-M4 Go re-plan and gatecheck CLI containment hardening

**Source document:**
`docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md`

**Scope.** Three units. Units A and B retarget held shipments 030-S and 032-S
onto the Go gate engines. Unit C is a new release unit harvested from stash
group G1 plus G6.

**Out of scope.** 031-S / 034-F — blocked pending the operator's Option A / B
selection (deliberation §4). Stash groups G2, G3, G5 — deferred with recorded
rationale (deliberation §1.5).

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

## Dependency graph and execution order

```
Unit A (030-S):  A-T1 → A-T2               (ALP-1: one atomic landing, one commit)

Unit B (032-S):  B-T1 → B-T2 → B-T3        (B-T1 is the test-first red)

Unit C (new):    C-T1 → C-T2 → C-T3        (C-T2 is the test-first red)
                 C-T1 → C-T4
                 C-T5  (independent)
                 C-T6  (independent)
```

**Recommended shipment order: Unit C → Unit B → Unit A.**

* **Unit C first** — the only unit closing a *live* security-adjacent gap
  (`check-write-path-precondition.sh` has no `--root` guard today), and C-T1
  touches shared CLI dispatch that both other units' engines run through. Landing
  it first means A and B develop against the final dispatch contract.
* **Unit B second** — self-contained in one package, no design fork, lowest risk.
* **Unit A last** — the only unit that must mutate `pin.go` in lockstep with the
  code the pin asserts against (formalised as **ALP-1**, INV-5); benefits from a
  quiet tree, since its interior boundary is the plan's only sanctioned red gate.

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

*Generated by Copilot*
