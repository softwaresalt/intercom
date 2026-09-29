---
title: "Deliberation — Retire gate-engine Python by migrating CI guard engines to Go, and sequence held shipments 030-S/031-S/032-S"
description: "How stash C44E2C1F (Python→Go gate-engine migration) interacts with held plan units 6/7/8, and the dispositions of the related stash entries"
topic: "C44E2C1F Python-to-Go gate-engine migration vs. held shipments 030-S, 031-S, 032-S"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
agent: Stage
date: 2026-09-28
stage_branch: chore/stage-stash-python-to-go-migration
bound_snapshot: 1651936 (origin/main, PR #77 merged; closure PR #78 open)
source_stash:
  - C44E2C1F
reconciled_stash:
  - C312BD4C
  - 40C421EF
  - 54EF986C
  - 8387758F
  - 150364D2
  - C98B92F0
  - 124AE9DE
  - C8914513
  - DCD67C30
  - B6EF23CC (already archived 2026-09-19)
  - 6C24E2E4 (already archived 2026-09-19)
linked_artifacts:
  - "docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md"
  - "docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md"
  - "docs/plans/2026-09-18-intercom-go-checker-test-docs-hygiene-plan.md"
tags:
  - "ci-gates"
  - "python-retirement"
  - "go-migration"
  - "verdict-parity"
  - "shipment-sequencing"
---

## Operator authority for this session

The operator directed the following, and the Orchestrator relayed it: "Hold [030-S, 031-S, 032-S] until Stage has triaged
C44E2C1F." Stage is to deliberate. If the decision is clear, Stage continues through plan, harden, review
and harvest. Stage treats that as conditional pre-authorization to promote this outcome to planning.

The outcome takes effect only on the operator's own gate: the operator pushes the Stage artifact branch and merges the
staging PR. Nothing here is claimable before then. If the operator rejects the recommendation below, the staging PR is
where that happens.

## Problem Frame

Stash **C44E2C1F** (feature, medium) asks to remove all workspace Python except autoharness. The Python to remove is
the CI guard-script engines:

* `scripts/lib/gomask.py` (160 lines)
* `scripts/lib/retired_arch.py` (989 lines)
* their unittest suite `scripts/lib/tests/*.py` (5 files, 533 lines), which landed via 029-S / PR #77
* the inline Python heredocs still embedded in `scripts/check-merge-strategy.sh` (about 50 lines),
  `scripts/check-unignore-regression.sh` (about 417 lines) and `scripts/check-write-path-precondition.sh`
  (about 136 lines)

That is about 2,300 lines of Python, and every line of it backs a gate the `ci-gate` job depends on. The requested
properties are:

* stdlib-only Go that uses `go/ast` and `go/parser`
* `go test` coverage
* thin wrappers or direct CI invocation
* identical exit codes, verdicts, and `--self-test` / `--self-test-integrity` modes
* the source-text-anchored selection pathspec pin
* before/after evidence, as 029-S produced
* the lint-job Python steps removed
* the autoharness Python pin kept (pipeline-topology job setup-python, the hash-locked pip bootstrap, and the
  deploy-harness prerequisite)

Three queued shipments from the 2026-09-18 staging cycle change exactly this Python:

| Shipment | Feature | Plan unit | What it changes | Python-specific design surface |
|---|---|---|---|---|
| 030-S | 033-F | gate-reliability plan §4 (overall unit 6) | Single `SCAN_SCOPE` declaration in `retired_arch.py` | Tasks name `retired_arch.py`. U3-T3 anchors the pin with `inspect.getsource()`. |
| 031-S | 034-F | gate-reliability plan §5 (overall unit 7) | Write-path gate hardening: selectors, call extent, allowance predicate, alias resolution, `Root.Resolve` tripwire | 034.002-T hard-codes D-1: "paren-depth walk over masked text … **No Go parser**". 034.006-T specifies "parse the import block" with the regex scanner. |
| 032-S | 035-F | hygiene plan §2 (overall unit 8) | `root_gitignore_text_at` must fail closed on an invalid ref | Tasks target the Python function, its line numbers and its heredoc comments. |

Shipment 033-S (036-F, pipeline-topology docs) touches no gate-engine code. It is **not** affected and stays unblocked.

**Question.** How should C44E2C1F be sequenced against 030-S, 031-S and 032-S? The options weigh rework, gate-verdict
parity risk, and the safety of merge-blocking gates.

**Success criteria.**

1. No merge-blocking gate changes its verdict unintentionally at any point.
2. Every intended verdict change is enumerated and evidenced separately from the port.
3. Little work is thrown away.
4. Each step can be reverted.
5. The held shipments cannot be claimed out of order.

**Out of scope.**

* Autoharness Python, including the topology-check job and the deploy-harness prerequisite.
* Any product (`cmd/**`, `internal/**`) code.
* New detection semantics. Those belong to 034-F.
* Changing merge-strategy SKIP semantics. That is C8914513.
* Credential wiring. That is 124AE9DE.

## Research Findings

Every finding below was measured against `origin/main` at `1651936` or the pinned toolchain. None is recalled from memory.

* **F-1: `go run` collapses exit codes.** This was checked in the pinned toolchain source (go1.26.5,
  `$GOROOT/src/cmd/go/internal/run/run.go:56`): "The exit status of Run is not the exit status of the compiled
  binary." `base.RunStdin` calls `base.Errorf`, which calls `SetExitStatus(1)`. If a wrapper used
  `go run ./tools/…`, every non-zero verdict would become exit 1 and `exit status N` would be printed to stderr.
  The gates' exit-2 contract (for example the merge-strategy error class) would then be silently lost. **Wrappers
  must build to a temporary path and run the binary. They must not use `go run`.** This refines the stash's
  "`go run`" suggestion.
* **F-2: The tool cannot live under `cmd/**` or `internal/**`.** `check-write-path-precondition.sh` rejects every
  write primitive (`os.WriteFile`, `os.MkdirTemp`, `os.CreateTemp`, …) in non-test Go under `internal/**` and
  `cmd/**`. The retired-architecture gate scans `cmd/**` including test files. The ported unignore engine needs
  scratch directories and files, and the retired-arch engine carries the retired vocabulary. A tool under either
  root would make the gates scan themselves and trip the write-path gate. **The tool must live outside both scan
  roots.** Candidate: `tools/gatecheck/` in the root module.
* **F-3: The root module already runs the test surface.** The existing jobs cover `./...`:
  `go build ./...` (ci.yml:214), `go vet ./...` (:216), `go test -race ./...` (:220), the advisory Windows
  `go test ./...` (:277), `golangci-lint run ./...` for both GOOS values (:377, :392) and `govulncheck ./...`
  (:423). A root-module `tools/gatecheck` package is tested, linted and vulnerability-scanned with no new CI steps.
  Cross-compile builds only `./cmd/${name}` (:547), so the tool is never released. The `changes` path filter is a
  fail-closed denylist (`'**'` minus docs and backlog), so `tools/**` changes always trigger the gates.
* **F-4: TOML has no stdlib parser.** `retired_arch.py` scans `config.toml.example` with `tomllib`. It also keeps
  a hand-written fallback lexer, and `--self-test` requires the **two engines to agree** on every TOML fixture
  (`self_test_engines_for_name`). Go's standard library has no TOML parser. `github.com/BurntSushi/toml v1.6.0` is
  already a **direct** dependency of the root module (go.mod). Using it adds no new module, go.sum entry or
  supply-chain surface, and it preserves the dual-engine agreement check. A stdlib-only port would collapse to the
  fallback engine alone and lose that check.
* **F-5: The masker is a lexical state machine, not a parser.** `gomask.mask_go_non_code` is a 6-state
  character-level lexer. The only exception is a whole-content struct-tag regex on raw strings. A faithful Go port
  is mechanical, but three **Python-vs-Go semantic traps** would break byte-identical parity if ported naively:
  * Python iterates code points; a Go `for i := 0; i < len(s); i++` iterates bytes. A multi-byte rune would be
    masked to N spaces instead of 1.
  * Python `re` `\s` and `\w` are Unicode-aware for `str`. Python `\s` includes `\x1c`–`\x1f`, `\x85` and Unicode
    Z. Go RE2 `\s` is `[\t\n\f\r ]` and `\w` is ASCII.
  * The write-path selectors use a `(?<![\w.])` lookbehind that RE2 does not support.

  `retired_arch.py` also relies on `str.splitlines()`, which splits on `\r`, `\v`, `\f`, `\x1c`–`\x1e`, `\x85`,
  `U+2028` and `U+2029`, and on Unicode `\b`. Line-number parity depends on reproducing these exactly.
* **F-6: Two CI jobs have no Go toolchain.** The lint job already sets up Go (ci.yml:293) before its gate steps
  (:298–338). The `gitignore-append-only` job (:448, which runs `check-unignore-regression.sh`) and the
  `merge-strategy` job (:674) have **no `setup-go` step**. Today both rely on the runner's default `python3`.
  Migrating them adds an SHA-pinned `actions/setup-go` step to each. That changes the autoharness-generated file,
  so the LOCAL DIVERGENCE header must be updated.
* **F-7: The held units' designs do not survive a language change.** Their *contracts* (the ACs) do.
  * 034-F's D-1/D-2 design (paren-depth call extent over masked text, "No Go parser") exists only because the
    Python gate is line/regex-based. In Go, `go/ast.CallExpr.Args` gives the call extent and argument list
    directly. Import aliases come from `ast.File.Imports`. The `Root.Resolve` tripwire's "receiver bound from
    `pathsafe.NewRoot`" rule becomes an assignment-tracking AST walk. The regex design would be built and then
    thrown away.
  * 033-F's pin anchor (`inspect.getsource`) has a direct Go analogue: parse the tool's own source with
    `go/parser`.
  * 035-F's fix targets a function the port rewrites.
* **F-8: Prior learnings apply.**
  * `docs/compound/2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md`: the merge-strategy heredoc
    once consumed stdin. Parity evidence must exercise the **live stdin transport**, not only by-path fixtures.
  * `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`: pins must stay
    source-anchored, and each criterion must be red at the change's own parent.
  * `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`: build a
    per-gate surface matrix covering the engine, wrapper, ci.yml step, `ci-gate.needs`, fixtures, docs and pin.
  * `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`: verify stdlib behaviour
    in GOROOT, as F-1 does.
* **F-9: Gates run from the PR's own checkout.** A regression in a ported engine therefore does not block the PR
  that reverts the wrapper. Kill switches exist for retired-arch (`RETIRED_ARCH_GATE_ADVISORY`) and write-path
  (`WRITE_PATH_GATE_ADVISORY`). The merge-strategy gate is advisory by default. The unignore-regression gate has
  **no** toggle. Until the final retirement unit, reverting a single wrapper back to the still-present Python
  engine is a complete rollback.

## Options Evaluated

### Option A: Fold units 6/7/8 into the migration and supersede 030-S/031-S/032-S

Port and harden in one release stream. `SCAN_SCOPE`, the new write-path detections and the ref-resolution fix would
be implemented directly in Go.

* **Pros:** zero throwaway work. 034-F gets the simpler `go/ast` design from the start.
* **Cons:**
  * Port bugs and intended verdict changes land in the same diffs, so the before/after evidence can no longer tell
    them apart. That defeats success criterion 2 on merge-blocking gates, one of which guards a NON-NEGOTIABLE
    containment boundary.
  * backlogit has no shipment-cancel/supersede operation. Superseding means destructive deletes or orphaned queued
    shipments, and closing shipments is Ship's authority.
  * The plan-review history and ACs of three reviewed features would have to be rebuilt.
* **Effort:** high, in a single large stream. **Parity risk:** high.

### Option B: Parity-preserving port first, then re-plan units 6/7/8 onto Go (**recommended**)

1. Port every engine to a root-module Go tool (`tools/gatecheck`) with **byte-identical** masked output and
   identical verdicts and exit codes. Every non-identity is pre-enumerated.
2. Switch each gate's wrapper one at a time, with golden-file evidence captured from the Python engine at the
   change's parent.
3. Delete the Python and its CI steps only after all wrappers have switched.
4. The held shipments stay in the queue with their reviewed contracts intact. They are blocked on the retirement
   shipment, and Stage re-plans their tasks onto the Go code before they can be claimed.

* **Pros:**
  * Parity evidence stays clean: the port changes no verdict, and every later change is a separately evidenced delta.
  * Each wrapper switch is revertable to the still-present Python until the last unit.
  * 034-F's rework makes it *smaller* (AST replaces a hand-rolled extent walker).
  * No destructive backlog operations.
* **Cons:**
  * Some held-task text is rewritten (a Stage re-plan pass).
  * 034-F's defense-in-depth detections land later. The trigger is unfired and 034-F is detection-only, so the
    exposure delta is nil.
  * The known-latent 035-F defect is ported faithfully and fixed one shipment later.
* **Effort:** medium-high, spread over four small shipments. **Parity risk:** low. Parity is measured, not assumed.

### Option C: Ship units 6/7/8 in Python first, then migrate

* **Pros:** the hardening and ref fix land soonest. No re-plan is needed.
* **Cons:**
  * About 15 tasks, roughly 25 h, go into Python code the migration deletes. That includes the D-1 extent walker
    that `go/ast` makes redundant.
  * The port's parity surface roughly doubles: it would also have to reproduce alias resolution, the allowance
    predicate and the tripwire byte-for-byte.
  * Python gate surface grows at the moment the operator decided to retire it.
* **Effort:** highest in total. **Parity risk:** medium, from the larger surface.

## Trade-off Comparison

| Criterion | A: fold | B: port → re-plan | C: Python first |
|---|---|---|---|
| Unintended verdict-change risk on merge-blocking gates | High (mixed diffs) | **Low** (identity port, golden-pinned) | Medium (larger port surface) |
| Throwaway work | None | Low (task text re-plan only) | High (~25 h Python + regex extent walker) |
| Rollback per step | Poor (hardening + port coupled) | **Per-gate wrapper revert** until retirement | Per-gate, but after doubled port |
| Time-to-hardening (034-F) | Mid | Latest (acceptable: unfired trigger, detection-only) | Earliest |
| Backlog integrity | Destructive supersede / orphan shipments | **Preserved** (blocks + re-plan) | Preserved |
| Fit with "retire Python" | Good | **Good** | Poor (grows Python first) |

## Decision

**Adopt Option B.**

### D-1: Sequencing

Four new features and shipments, which the plan names M1–M4:

| Unit | Scope | Blocked by |
|---|---|---|
| M1 | Tool scaffold, masker and write-path port | 029-S closure |
| M2 | Retired-architecture port | M1 |
| M3 | Unignore-regression and merge-strategy port | M1 |
| M4 | Python retirement, CI and docs | M2 and M3 |

Held shipments **030-S, 031-S and 032-S** each gain an explicit `blocks` dependency on the **M4** shipment. Their
covering features **033-F, 034-F and 035-F** move to `blocked` status with a recorded reason: re-plan onto Go is
required before claim.

After M4 ships, Stage re-plans their **tasks** in place against the Go tool and keeps each feature's reviewed ACs:

* 034-F's D-1/D-2 design is replaced by `go/ast`.
* 033-F's pin is anchored with `go/parser` on the tool's own source.
* 035-F targets the Go baseline-resolution function.

Suggested post-M4 order: 032-S (smallest; fail-closed repair), then 030-S, then 031-S. **033-S is untouched and stays
claimable.**

### D-2: Tool location and packaging

The tool is a root-module package at `tools/gatecheck/`: one `main` with a gate sub-command per engine, and one
subpackage per engine. It is outside both gate scan roots (F-2), covered by the existing `./...` CI surface (F-3),
and never released.

The rejected alternative is a nested module. It would need a second toolchain pin, separate test/lint/vuln steps,
and a duplicate TOML dependency pin.

**Plan-review refinement (2026-09-28, plan revision 2):**

* The engine packages live under `tools/gatecheck/internal/`, so Go's `internal` visibility rule enforces the
  import isolation at compile time.
* Gate selection cannot reach them. The pathspecs are root-anchored, and the prefix predicates test paths
  relative to the repository root.
* Plan-review also asked whether to gate the held shipments per gate or on M4. They stay gated on M4 for three
  reasons: evidence integrity, verdict-neutral rollback until M4, and a single re-plan pass (plan §8 PD-5). The
  per-gate alternative is recorded as operator question Q-4.

### D-3: Invocation

`scripts/check-*.sh` stay as **thin wrappers**. Each builds `./tools/gatecheck` to a `mktemp -d` path, runs the
binary, passes the exit status through unchanged, and removes the temp directory in a trap. They **never** use
`go run` (F-1). Keeping the wrappers leaves the ci.yml `run:` lines, `ci-gate.needs`, pre-push hooks and docs stable.

A build failure exits non-zero and is never 0. Its exact code is pinned per gate in the plan.

### D-4: Parity-first masker

M1 ports `gomask` as a **rune-level** state machine. The struct-tag matcher emulates Python's Unicode `\s`/`\w`
semantics (F-5). Byte-identical masked output is asserted against goldens that the Python masker generates over
every tracked `.go` file in `cmd/**`, `internal/**` and `scripts/testdata/**`.

`go/ast`/`go/parser` are used in the port where they *preserve* semantics, namely the source-text-anchored pin (D-6).
They are **not** used to replace the masker's detection semantics. That is a verdict change, and it belongs to the
re-planned 034-F and to C312BD4C.

This refines the stash's "use go/ast instead of the regex masker" and is recorded as unresolved question Q-1.

### D-5: TOML

The primary engine uses the existing `BurntSushi/toml` dependency as the `tomllib` analogue. The ported fallback
lexer is kept, and `--self-test` keeps the dual-engine agreement check (F-4). This is the one deviation from
"stdlib only": it uses an already-required module and adds nothing to go.sum. Recorded as Q-2.

### D-6: Pin anchoring

The selection pathspec pin stays **source-text-anchored**. At runtime the Go engine parses its *own* selection source
file from disk with `go/parser` and requires the pathspec and prefix literals **inside** the selection functions'
bodies. It fails closed when the source cannot be read.

A module-constant self-comparison is forbidden. That is H-11 of the gate-reliability plan, carried forward.

### D-7: Parity definition

Parity means an identical exit code and an identical normalized finding set: `path:line:` plus token and model for
retired-arch, `path:line: selector` for write-path, and per-scenario verdicts for unignore and merge-strategy.
`--self-test` PASS/FAIL assertion names and outcomes must also be identical.

Exception-message text from different parsers is excluded, for example `tomllib` vs `BurntSushi` parse-error prose.
The verdict and fail-closed reject must still match. Every other non-identity must be listed in advance as an
**enumerated delta**, with justification.

The deltas enumerated at decision time are:

* **ED-1: Merge-strategy outside a checkout.** Today `git rev-parse` fails under `set -e` with git's exit code.
  After the change, the wrapper guards it and exits 2 with a message. This is 150364D2 item 3, fail-closed in both
  directions.
* **ED-2: Invalid UTF-8 source file.** Python raises an uncaught `UnicodeDecodeError`, printing a traceback and
  exiting 1. Go must fail closed with a synthetic finding and exit 1, not a traceback. The verdict is unchanged; only
  the message differs.

### D-8: Folded stash work

* **8387758F** (write-path struct-tag D-1 fixtures) becomes M1's first characterization task. The fixtures are
  pinned against the *Python* engine before the port, so they guard the port as well.
* **150364D2** (merge-strategy hardening) is folded into M3:
  * item 1: the wrapper trap cleans up temp files
  * item 2: Go table test asserting PASS→0, SKIP→0, FAIL→1 and error→2
  * item 3: ED-1

## Rejected Alternatives

* **Option A** is rejected because it mixes port and behaviour change on merge-blocking gates, which destroys the
  parity evidence. It also needs shipment supersession that backlogit cannot perform non-destructively, and that
  would cross into Ship's closure authority.
* **Option C** is rejected because it builds roughly 25 h of Python, including a regex call-extent walker that
  `go/ast` makes redundant. The migration would then delete it and port a doubled parity surface.
* **Nested Go module for the tool** is rejected because it needs a second toolchain pin, dedicated CI
  test/lint/vuln steps, and a duplicate TOML pin, all without a safety gain.
* **Tool under `cmd/` or `internal/`** is rejected because it trips the write-path gate and makes gates scan
  themselves (F-2).
* **Direct CI invocation, no wrappers** is rejected for now. It churns every gate step, `ci-gate` wiring and the
  pre-push hooks for no parity benefit. A later clean-up can inline it.
* **Stdlib-only TOML (fallback lexer as sole engine)** is rejected by default because it loses the dual-engine
  agreement control (F-4). It is retained as the Q-2 alternative.

## Stash Triage and Reconciliation (P-021 C5/C6)

Every deferred-scope-expansion entry below was routed through this deliberation, as the precedence rule requires.

Duplicate detection (A) ran unconditionally over all of them. The late-identifier reconciliation (B) sources are
Ship-owned residual-risk records:

* `docs/closure/028-S-031-F-post-merge-closure.md` on `origin/main`, which cites PR #75 (lines 66 and 109)
* `docs/closure/029-S-032-F-post-merge-closure.md` on the closure PR #78 branch, which cites PR #77 (line 123) and
  lists these entries at lines 181–186
* merge commit `1651936`, "Merge pull request #77"

| Entry | Shape | Disposition | Late-identifier reconciliation (B) | Duplicate scan (A) |
|---|---|---|---|---|
| **C44E2C1F** | feature | **Promoted** to M1–M4 (this decision). Consumed, then archived at Step 5.6. | n/a (not a deferred capture) | Clean. C312BD4C, 40C421EF, 54EF986C, 8387758F and 150364D2 are *related* but each describes a distinct change. |
| **C312BD4C** | deferred task | **Remains deferred (active).** A masker semantic change must not ride on the parity port (D-4). Retargeted to the Go masker after M1. The characterization test is ported to Go unchanged in M1. Decide after M4, alongside the 034-F re-plan. | PR `N/A` → **#77** recovered (029-S closure record; merge `1651936`). Review thread `N/A` stands (pre-PR local finding). | Clean. 8387758F pins write-path D-1 fixtures, not masker newline semantics. |
| **40C421EF** | deferred task | **Superseded → archived.** After M4 no gate runs on Python, so there is no interpreter to pin. Reopen if M4 is abandoned. | `N/A` → **#77** recovered. Thread `N/A` stands. | Clean |
| **54EF986C** | deferred task | **Superseded → archived.** M4 removes the Python lint step. The Go code is covered by the existing `golangci-lint`. Reopen if M4 is abandoned. | `N/A` → **#77** recovered. Thread `N/A` stands. | Clean |
| **8387758F** | deferred task | **Folded → consumed → archived.** Becomes M1's first task (D-8). | `N/A` → **#77** recovered. Thread `N/A` stands. | Clean |
| **150364D2** | task | **Folded → consumed → archived.** Absorbed into M3 (D-8). | n/a. The source already names shipment 028-S. PR recorded as **#75**. | Clean |
| **C98B92F0** | deferred task | **Remains deferred (active), not selected.** Closure-naming conformance is unrelated to gate engines. | `N/A` → **#75** recovered (028-S closure record). Thread `N/A` stands. | Clean. The `DISCOVERY-STATUS: AMBIGUOUS DCD67C30` seed was checked: DCD67C30 is the resolved *rename*, while C98B92F0 is the *mode-vs-filename conformance gap*. Distinct expansions, not a duplicate. |
| **124AE9DE** | deferred task | **Remains deferred (active), not selected.** Credential wiring is language-independent. M3 must not change the SKIP-under-`GITHUB_TOKEN` behaviour it relies on. | `N/A` → **#75** recovered. Thread `N/A` stands. | Clean |
| **C8914513** | deferred bug | **Remains deferred (active), not selected.** M3 preserves the current SKIP semantics byte-for-byte. If the operator later decides fail-closed required mode, the change lands in Go. | `N/A` → **#75** recovered. Thread `N/A` stands. | Clean |
| **DCD67C30** | deferred bug | **Resolved → archived.** Operator-authorized exception commit `fdf075f` is an ancestor of `origin/main` (verified with `git merge-base --is-ancestor`). | `N/A` → **#75** recovered. Thread `N/A` stands. | Clean |
| **B6EF23CC** | (archived 2026-09-19) | **No action.** Already split: the verification half went to 031-F (shipped in 028-S / PR #75), and the settings half went to 042-F (operator-only, blocked). M3 must preserve 031-F's verdict contract. | n/a (archived) | n/a |
| **6C24E2E4** | (archived 2026-09-19) | **No action.** Item 2 went to 032-F (shipped in 029-S). Item 1 is 033-F (held here, re-plan onto Go per D-1). | n/a (archived) | n/a |

Reconciled identifiers are written into each surviving entry in place, under Stage stash authority. No second entry is
created and Ship writes nothing.

## Unresolved Questions (operator)

None of these blocks harvest. Each has a default that the plan implements.

* **Q-1: `go/ast` scope.**
  * Stage's default: `go/ast` enters in the re-planned 034-F and C312BD4C, as intended verdict changes. The M1 masker
    is a byte-identical lexical port.
  * Alternative: replace the masker with `go/scanner`/`go/ast` inside M1, with an enumerated delta set. Higher
    parity risk.
* **Q-2: TOML dependency.**
  * Stage's default: `BurntSushi/toml`, which is already a direct dependency (D-5).
  * Alternative: strict stdlib-only, with the fallback lexer as the sole engine. This loses the dual-engine
    agreement control.
* **Q-3: Post-M4 order of the held shipments.**
  * Stage's default: 032-S, then 030-S, then 031-S.
  * The operator may reorder. Each is independently blocked on M4.
* **Q-4: How the held shipments are gated** (added by plan-review).
  * Stage's default: all three are gated on M4 (plan §8 PD-5).
  * The alternative is a per-gate predecessor for each: 032-S on M3, 030-S on M2, 031-S on M1. That lets them
    start earlier, at the cost of evidence and rollback complexity.

## Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Silent verdict drift from Python/Go semantic differences (F-5) | Byte-identical masked-output goldens over the whole tracked Go corpus and fixtures. Explicit Unicode `\s`/`\w`, `splitlines()` and lookbehind emulation tests. Goldens generated from Python at the change's parent. |
| Exit-code collapse (F-1) | Build-then-exec wrappers. Per-gate exit-code table assertions in `go test` and in wrapper evidence. |
| Self-comparison pin | D-6: go/parser anchoring to the tool's own on-disk source, which fails closed. |
| Unignore gate has no kill switch | Per-gate wrapper switch in its own task. Revert-to-Python is a one-file rollback until M4. Scenario goldens cover every existing self-test scenario. |
| Autoharness regeneration drops the setup-go additions | LOCAL DIVERGENCE header updated in the same change (F-6). |
| Held shipments claimed with stale Python task text | Double lock: a `blocks` dependency on the M4 shipment, plus covering features set `blocked` with a re-plan reason. |
| Live stdin transport regression in merge-strategy | Evidence must include a piped `gh api`-shaped JSON run, not only by-path fixtures (F-8). |
