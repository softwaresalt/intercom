---
title: "Residual Hardening Triage — 19-entry stash disposition, grouping, and staging order"
date: 2026-09-04
status: accepted
agent: Stage
mode: DARK_FACTORY
supersedes: none
governs:
  - docs/plans/2026-09-04-intercom-go-ci-supply-chain-hardening-plan.md
  - docs/plans/2026-09-04-intercom-go-script-guard-hardening-plan.md
  - docs/plans/2026-09-04-intercom-go-apperr-taxonomy-correction-plan.md
  - docs/plans/2026-09-04-intercom-go-pathsafe-config-containment-plan.md
---

# Residual Hardening Triage (Stage, dark-factory cycle post-005-S)

## Scope Fence (EXACT — do not expand)

This deliberation covers **exactly 19** stash entries carried forward after
shipment `005-S` closed:

`4989A42D A92E3FA0 EFFAA358 90350C9A F7C6420D BEDD2E70 35D76D5E EF9352FB
90EE7758 8ACF7110 4A91CA81 7774C9CA 3D61B5A8 9A4C8749 2130906D 7028FBE7
8C2D578D B6203CCA 78D13775`

**EXCLUDED by operator directive** (created *during* 005-S; P-021 reserves them
for the next Stage cycle): `A0A2D049`, `6E701953`, `309FBF5A`. They were not
triaged, planned, harvested, or archived in this cycle. The active stash
contains 22 entries; 22 − 3 = 19, so scope is exactly reconciled.

## Problem Frame

Phases C1 (`004-S`) and C2 (`005-S`) are shipped. The governing architecture is
**design revision 2** under
`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`.
The next *feature* phase is **C3 (agent adapter / ACL)**.

The 19 carried entries are almost entirely **residual hardening debt** deferred
under P-021 C1 from shipments 001-S through 004-S. The question is not "what
feature comes next" — it is **what must be true before C3 adds new surface on
top of these primitives**.

Operator prioritization rules 3, 4, 5 and 6 are decisive here:

* reliability and security supersede feature work;
* feature work supersedes documentation-only work;
* simplicity supersedes complexity;
* **refactoring existing code for composability, simplicity, interoperability
  or reliability takes priority over adding features.**

Rule 6 is the controlling rule for this cycle. C3 will introduce an agent
adapter that constructs errors through `internal/apperr` and resolves paths
through `internal/pathsafe`. Correcting those primitives *after* C3 would mean
correcting them across a strictly larger blast radius. Therefore **all four
harvested shipments are correctly ordered ahead of C3**, and C3 is deliberately
not harvested in this cycle.

## Verification Performed (ground truth, not assumption)

Every disposition below was checked against the live post-005-S tree rather
than trusted from the stash text.

| Claim | Verified | Evidence |
|---|---|---|
| `pip install autoharness` unpinned in CI | TRUE | `.github/workflows/ci.yml:325` |
| gitleaks runs working-tree only | TRUE | `ci.yml:210` — `gitleaks detect --source . --no-git -v` |
| retired-architecture scan runs bare (no `--self-test`) | TRUE | `ci.yml:145`; `--self-test` implemented at `check-retired-architecture.sh:273` but never invoked by CI |
| `CODEOWNERS` absent | TRUE | `.github/CODEOWNERS` does not exist |
| `apperr` still carries retired Kinds | TRUE | `apperr.go:22,30,44` (`KindSlack`/`KindIPC`/`KindACP`); sentinels at `sentinel.go:22,26,33` |
| `Kind` has no exported Stringer | TRUE | only unexported `func (k Kind) prefix() string` at `apperr.go:50` |
| `allKinds` pinned to a hard count | TRUE | `apperr_test.go:29` asserts `len(allKinds) != 14` |
| `.golangci.yml` has no `printf.funcs` entry | TRUE | file exists (152 bytes), no govet/printf settings |
| `splitPath` is a one-line wrapper | TRUE | `pathsafe.go:78` wraps `filepathSplitList` at `pathsafe.go:109` |
| `hasPathPrefix` exists as described | TRUE | `pathsafe.go:166` |
| `validate.go` rule 7 reimplements traversal detection | TRUE | `containsDotDotSegment(c.Database.Path)` in `internal/config/validate.go` |
| `pathsafe` exported surface is narrow | TRUE | only `Root`, `NewRoot`, `Root.Path`, `Root.Resolve` |
| go.mod pins toolchain above language floor | TRUE | `go 1.24` + `toolchain go1.26.5`; rationale was inline-only at triage time (now also recorded in the decision note produced by this cycle) |
| Plan Constitution table mismatches ratified numbering | TRUE | ratified IV = "CLI Workspace Containment", VIII = "Explicit Safety Modes"; plan claims "Least surprise" / "Dependency minimalism" |
| `docs/oracle-pin.md` has any live consumer | **FALSE** | referenced only by historical P1/P2-era plan/decision prose; **no workflow, no script** references it |
| `.env.local` ever committed | **FALSE** | no commit touches it; no `.env*` path ever added in history; ignored via `.gitignore:36` (`.env.*`) |

## P-021 C5/C6 Intake Obligations

**(A) Duplicate detection — UNCONDITIONAL, run over all 15 marker-carrying
entries this cycle.** Result: **CLEAN — no duplicates.** Four near-collisions
were examined and each resolved to genuinely distinct findings:

* `EFFAA358` / `BEDD2E70` / `90EE7758` all cite source task `001.002.001-ST`,
  but are respectively a version-pin, an ownership gate, and a docs note.
* `B6203CCA` / `78D13775` both target `check-retired-architecture.sh`, but are
  respectively CI wiring and masking-state-machine logic.
* `7774C9CA` / `8C2D578D` both target the `apperr` taxonomy; `8C2D578D` itself
  records that they are DISTINCT and neither supersedes the other. Confirmed.
* `4A91CA81` / `3D61B5A8` both target `NewRoot`, but are respectively cause
  preservation and directory-type validation.

No entry carried a `DISCOVERY-STATUS: AMBIGUOUS` or `LOOKUP-UNAVAILABLE` token;
three carried explicit "not ambiguous" tokens. The clean-scan outcome is
recorded here explicitly so it is distinguishable from a scan that never ran.

**(B) Late-identifier reconciliation — no-op this cycle.** All 15 marker
entries were already reconciled by the prior Stage cycle (PR #4 for 001-S
entries, #8 for 002-S, #11 for 003-S, #15 for 004-S), sourced from the
Ship-owned closure records under `docs/closure/`. Every residual `N/A` is a
**terminal, truthful** record: those closure records state 0 PR review comments
and 0 formal reviews requested, so no review thread ever existed to recover.
Reconciliation is idempotent — no concrete identifier was overwritten and no
`N/A` was written over a concrete value. `EF9352FB`'s `task=N/A` is likewise
terminal: it was a repo-wide scan finding never bound to a task unit.

## Decisions

### D-A — `8ACF7110` (oracle-drift gate) is OBSOLETE. Archive.

This is the reassessment the operator directive called for. The entry proposes
a CI gate that reads `docs/oracle-pin.md` and fails when the Rust checkout's
`HEAD` drifts from `41df772`.

Three independent grounds retire it:

1. **The oracle is demoted.** Governing decision **D8** (recorded in `4989A42D`)
   demotes `softwaresalt/agent-intercom` from behavioral oracle to *historical
   reference only*. Parity with it is explicitly **no longer a definition of
   done**.
2. **The thing the gate protects is retired.** The entry's own stated purpose is
   protecting unattended execution of **phases P2–P14** — an ordering that
   `4989A42D` marks **VOID** in its entirety, along with the 56/22/46/6 parity
   corpus the pin's verification method depends on.
3. **There is no live consumer.** Verified: no workflow and no script reads
   `docs/oracle-pin.md`. The document self-describes as "documentary only".

Building the gate would spend a CI surface defending a contract that no longer
exists, against a repository that is no longer authoritative — a direct rule-5
(simplicity) violation. The entry's own listed alternative (vendoring oracle
fixtures) is retired for the same reason.

`docs/oracle-pin.md` is **left in place, untouched**: it is a truthful
historical record of how P1 was verified, and deleting it would destroy
provenance for shipped work. Archiving the stash entry retires the *action*
without destroying the *record*.

### D-B — `EF9352FB` (Tavily key) is a real security item, but is BLOCKED on operator capability.

Split the entry into its repo-side and operator-side halves:

* **Repo-side containment — VERIFIED CLOSED this cycle, read-only.** The key was
  never committed: no commit touches `.env.local`, no `.env*` path was ever
  added anywhere in history, and the file is ignored by `.gitignore:36`. There
  is no in-repo remediation left to perform, and therefore **no code change to
  harvest**. The file's contents were **not read** and no key material appears
  in any artifact produced by this cycle.
* **Operator-side rotation — BLOCKED.** Rotating a live Tavily credential
  requires authenticating to an external provider console. That is an operator
  capability Stage does not hold and must not simulate.

Disposition is therefore **blocked/unsafe pending operator**, deliberately
isolated into no shipment so it cannot stall the 13 safe entries. This satisfies
the directive to "isolate and mark unsafe/blocked while allowing other safe work
to proceed". Note the residual exposure is time-sensitive: the key sits in
plaintext on the operator's disk, so the item is reported at the top of the
handoff rather than buried in a shipment manifest.

### D-C — `4989A42D` and `A92E3FA0` are RETAINED as standing trackers.

Neither is implementable work; both are governance records that must outlive
this cycle.

* **`A92E3FA0`** is the governing product-direction record and the standing
  roadmap tracker for **C3–C11**. It also carries five unresolved operator
  decisions (Q3/H3 Dev Tunnel auth → blocks C10; Q6/H4 sessions-per-process →
  blocks C9; H5 permission authorization → blocks C5; H6 session-pointer
  semantics → blocks C4; new Q7 workspace tooling reach). Archiving it would
  destroy the only index of those blockers. **Retain.**
* **`4989A42D`** is the migration-remediation tracker for retired-architecture
  contamination surviving in shipped code. Its *current* sole item (`8C2D578D`)
  is harvested this cycle — but the tracker's job is continuous: it must catch
  contamination the U-E1a gate surfaces across C3–C11. Retiring it the moment
  its backlog momentarily empties would remove the mechanism precisely when new
  code starts landing. **Retain**, with a note that its open item is now tracked
  by feature `009-F`.

  Adversarial review challenged this, arguing an empty tracker with no
  completion condition is speculative governance state. The challenge is
  partially accepted: the retention stands, but a **concrete completion
  condition** is now attached rather than leaving it open-ended —

  > `4989A42D` is archived when **both** hold: (a) `009-F` has closed, and
  > (b) the U-E1a retired-architecture gate has been broadened back to
  > `internal/**` (reversing the D6a narrowing) and passes. Condition (b) is the
  > real signal that contamination tracking is no longer needed, because the
  > gate itself then covers what the tracker was watching by hand.

  This preserves the mechanism while making it terminable, which was the
  legitimate core of the objection.

### D-D — `2130906D` and `90EE7758` are Stage-owned artifact work, resolved in-cycle.

Both request changes to **planning/decision artifacts**, not source code:

* `2130906D` corrects a Constitution Check table inside a plan document, and its
  own text states the reason it was deferred: *"Ship does not create or modify
  plan/deliberation artifacts (Stage-owned)."*
* `90EE7758` requests a **decision note** — an artifact in `docs/decisions/`.

Harvesting these to Ship would hand Ship work its role boundary forbids.
Stage's boundary explicitly permits creating and committing decision and plan
artifacts, so Stage resolves both **in this cycle** rather than creating a
shipment that Ship could not legally execute. See the Constitution Check
Correction section below and
`docs/decisions/2026-09-04-go-toolchain-pin-maintenance-note.md`.

### D-E — Grouping: four shipments, partitioned by artifact class and file surface.

Grouping is by **code/context similarity** as directed, with artifact-class
isolation preserved inside each shipment's task set.

| Group | Surface | Entries |
|---|---|---|
| **007-F** CI supply-chain hardening | `.github/workflows/ci.yml`, `.github/CODEOWNERS`, constraints lock (declarative CI/YAML) | `EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70` |
| **008-F** Build & scan script hardening | `scripts/build.ps1`, `scripts/check-retired-architecture.sh` + its CI wiring (shell/PowerShell/Python logic) | `35D76D5E`, `78D13775`, `B6203CCA` |
| **009-F** apperr taxonomy correction | `internal/apperr/**`, `.golangci.yml` (Go contract) | `8C2D578D`, `7774C9CA` (apperr parts), `9A4C8749` |
| **010-F** pathsafe & config containment | `internal/pathsafe/**`, `internal/config/validate.go` (Go correctness) | `4A91CA81`, `3D61B5A8`, `7028FBE7`, `7774C9CA` (`splitPath` part — **final resolver**) |

**`B6203CCA` was relocated to 008-F by adversarial review.** It was originally
grouped with 007-F on artifact class (CI YAML). Review found that wiring
`--self-test` as a blocking gate *before* 008-F changed the scanner's masking
logic would make 008-F's test-first fixtures unmergeable — they must fail
against the pre-change implementation, which would redden the gate 007-F had
just installed. Correctness of ordering beat purity of artifact class, so the
unit moved to sit after the logic it gates. Both shipments are now
self-contained with no cross-shipment dependency.

**On `7774C9CA` spanning two shipments.** This entry bundles three findings
across two packages: a `pathsafe.splitPath` redundancy plus two `apperr`
findings. Rather than force the whole entry into one shipment and have two
shipments both mutate `internal/pathsafe/pathsafe.go`, the `splitPath` item is
placed with the other `pathsafe` work in `010-F`. Because shipments execute
strictly sequentially (P-016, no parallel worktrees), the split creates no
concurrent-edit hazard. **`010-F` is designated the final resolver**: the entry
is archived only when 010-F closes, so it cannot be orphaned between two
"partially resolves" claims (a gap adversarial review caught).

### D-F — Staging order and dependency chain.

`007-S → 008-S → 009-S → 010-S`, strictly sequential, successors remain queued.

1. **007-S first (security + reliability of the pipeline itself).** An unpinned
   `pip install autoharness` executes on CI runs today; under dark-factory
   unattended operation, the pipeline is the trust root for everything after it.
   `BEDD2E70` additionally closes a self-modifying-PR gap that the topology-check
   job's own comments already document. Hardening the gate that validates all
   later shipments must precede those shipments.
2. **008-S second (containment guard + scanner correctness).** `35D76D5E` is a
   Constitution-IV containment guard (repo-root escape, symlink-aware).
   Sequenced after 007-S because 007-S installs the `--self-test` CI invocation
   that gives 008-S's scanner change automated regression coverage.
3. **009-S third (composability refactor, rule 6).** Correcting the `apperr`
   Kind taxonomy is a *contract* change pinned by a `len(allKinds) == 14` drift
   test. It must land before C3 multiplies the construction sites. `9A4C8749`
   rides here because registering `Newf` in `printf.funcs` is an `apperr`
   concern and protects the ~470 projected future construction sites.
4. **010-S fourth.** Depends on 009-S via a **hard symbol dependency**:
   `4A91CA81` requires a cause-preserving constructor, and adversarial review
   proved no existing `apperr` constructor can both retain a message and
   preserve the cause (`Wrap` takes no message; `New` drops the cause). 009-F
   U5 adds `Wrapf`; 010-F U1 consumes it. `7028FBE7` then de-duplicates
   traversal logic (rule 6 composability).

   *(The originally stated justification — "`pathsafe` imports `apperr`" — was
   spurious and review correctly rejected it. The dependency is real, but for a
   different reason than first recorded.)*

**Ordering challenge considered and rejected.** One reviewer argued for
`007-S → 008-S → 010-S → 009-S`, on the grounds that 010-F fixes live
containment defects while 009-F is zero-consumer taxonomy cleanup, and that
security-first should therefore promote 010-S. The argument was sound against
the draft as written, but is defeated by the `Wrapf` finding: 010-F U1 now
cannot compile without a constructor that 009-F introduces. Reversing the order
would force either a duplicated constructor or a blocked unit. The chosen order
stands, now on a verified technical dependency rather than a stylistic one.

Documentation-only work is last by rule 4 — and in fact the only pure-docs items
(`2130906D`, `90EE7758`) are Stage-owned and resolved in-cycle without consuming
a shipment slot at all.

## Options Considered and Rejected

| Option | Rejected because |
|---|---|
| One combined "residual hardening" shipment (13 entries) | ~14 tasks ≈ 28h. Violates coherent-release-unit sizing; a failure anywhere blocks all four unrelated surfaces; mixes YAML/bash/Go artifact classes in one review. |
| Harvest C3 (agent adapter) now, ahead of the debt | Direct rule-6 violation. C3 builds *on* `apperr` and `pathsafe`; correcting them afterwards enlarges the blast radius. C3 also needs its own deliberation and is out of the 19-entry scope fence. |
| Build the `8ACF7110` oracle-drift gate as specified | Defends a void contract against a demoted repository with no live consumer (D-A). |
| Archive `4989A42D` since `8C2D578D` is its only open item | Removes the contamination-tracking mechanism exactly when C3–C11 begin landing new code (D-C). |
| Harvest `2130906D` / `90EE7758` into a docs shipment | Assigns Ship work that Ship's role boundary forbids (D-D). |
| Include `EF9352FB` in 007-S as a "security" task | No in-repo change exists to make (containment already closed); the residual action is operator-only. Bundling it would block a shipment on an external console login (D-B). |

## Constitution Check Correction (resolves `2130906D`)

The ratified numbering in `.github/instructions/constitution.instructions.md`
is authoritative and is reproduced here as the canonical table for all future
plans in this workspace:

| # | Ratified principle |
|---|---|
| I | Safety-First Go |
| II | Test-First Development (NON-NEGOTIABLE) |
| III | Workspace Isolation and Security Boundaries |
| IV | CLI Workspace Containment (NON-NEGOTIABLE) |
| V | Structured Observability |
| VI | Single Responsibility |
| VII | Destructive Command Approval (NON-NEGOTIABLE) |
| VIII | Explicit Safety Modes for Elevated Risk |
| IX | Git-Friendly Persistence |
| X | Agent Context Efficiency |
| XI | Merge Commit History Preservation (NON-NEGOTIABLE) |

`docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md` carries a
non-matching table (its IV "Least surprise", VIII "Dependency minimalism").
That plan is a **historical record of shipped work** and is therefore not
rewritten; a correction banner is appended pointing here, which stops the bad
table propagating by copy into C3–C11 plans while preserving what was actually
reviewed at the time. The four plans generated by this cycle use the ratified
table above.

## Open Questions

None blocking this cycle. `EF9352FB` requires operator action but blocks no
harvested shipment. The five C3–C11 operator decisions carried by `A92E3FA0`
(Q3/H3, Q6/H4, H5, H6, Q7) are out of scope here and remain tracked there.

## Adversarial Review Record (2026-09-04)

Per operator directive, all groupings, prioritization, dispositions, and every
implementation plan were subjected to independent multi-persona adversarial
review before any shipment was marked ready. Five reviewers ran in parallel on
**four different model families** to avoid correlated blind spots.

| Reviewer | Model | Scope | Initial verdict |
|---|---|---|---|
| Scope Boundary Auditor | `gpt-5.6-sol` | Dispositions, grouping, prioritization, verification claims | **FAIL** |
| Security Reviewer | `gpt-5.6-terra` | 007-F, 008-F | **FAIL** |
| Correctness Reviewer | `claude-opus-4.8` | 009-F, 010-F | **FAIL** |
| Constitution Reviewer | `gemini-3.8-flash` | All five artifacts | **FAIL** |
| Maintainability Reviewer | `grok-4.6` | Sizing, artifact isolation, ambiguity | **FAIL** |

### Convergent blockers (found independently by 2+ reviewers)

1. **`apperr.Wrap` cannot carry a message.** 010-F U1's acceptance criteria were
   mutually unsatisfiable; an unattended Ship would have shipped a
   context-losing, test-breaking change to a security control. → 009-F U5
   (`Wrapf`) added; became the real 009→010 dependency.
2. **010-F U6 would have caused a containment regression.** A
   `normalize`-based predicate accepts `..\..\etc` on Linux, which
   `containsDotDotSegment` rejects today, and breaks the deliberately-permitted
   absolute `database.path` contract. → Rewritten as a zero-verdict-change move.
3. **009-F U4's lint config was silently inert** — golangci-lint **v2** schema
   with a v1 key and a malformed function path. → Corrected and proven by a
   committed fixture.
4. **CODEOWNERS does not close the self-modifying-PR gap and could deadlock
   unattended merges.** → Disposition narrowed to ownership metadata; blocking
   pre-flight ruleset probe added.
5. **007-F U6 before 008-F made test-first unmergeable.** → Unit relocated to
   008-F; `B6203CCA` re-dispositioned.
6. **Ground-truth errors that would strand an unattended executor:** validation
   "rule 10" is actually **rule 7** (tests in `paths_test.go`, not
   `validate_test.go`); `strip_toml_comment` is **Python in a heredoc**, not
   bash; `build.ps1` compares **Ordinal on purpose** (the draft's case-fold
   would have *loosened* a containment control).

### Disposition of findings

All blockers and majors were remediated in-cycle; each plan carries a
**Post-Review Remediation Record** mapping finding → fix. Two findings were
**partially accepted with reasoning recorded** rather than adopted wholesale:

* *Archive `4989A42D` instead of retaining it.* Retention stands, but a concrete
  terminating completion condition was added (see D-C).
* *Reorder to `010-S` before `009-S`.* Rejected on the merits — the `Wrapf`
  symbol dependency makes the original order technically required (see D-F).

Deferred out of scope, recorded rather than silently dropped:

* Semantic unification of the two traversal predicates (beyond the mechanical
  move) needs its own deliberation and an operator-facing migration decision.
* Enabling code-owner review enforcement requires first provisioning an approval
  path for unattended PRs — an operator decision, carried to the handoff.

**Verified-correct claims that survived review unchanged:** zero external use of
the retired `apperr` Kinds; the exact 11-Kind retained list; safety of iota
renumbering; the byte-identical `Error()` claim; the reality of the
doubled-separator bug at filesystem root; and the absence of any live consumer
of `docs/oracle-pin.md` (independently re-verified, supporting D-A).

**Post-remediation verdict: PASS (with recorded residual risk).** Residual items
are explicitly assigned as blocking stop conditions inside the plans (return the
unit blocked rather than guess) or escalated to the operator handoff.
