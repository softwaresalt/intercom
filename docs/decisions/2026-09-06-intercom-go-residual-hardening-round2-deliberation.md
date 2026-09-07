---
title: "Residual Hardening Round 2 — 16-entry stash disposition, single-shipment grouping, and execution order"
date: 2026-09-06
status: accepted
agent: Stage
mode: DARK_FACTORY
supersedes: none
governs:
  - docs/plans/2026-09-06-intercom-go-residual-hardening-round2-plan.md
---

# Residual Hardening Round 2 (Stage, dark-factory cycle post-009-S)

## Scope Fence (EXACT — do not expand)

This deliberation covers **exactly 16** stash entries, the authoritative
bounded scope supplied at activation:

`4989A42D, A92E3FA0, EF9352FB, A0A2D049, 6E701953, 309FBF5A, 11ECB954,
AFE0E95C, DA945722,707FE72B, 9F9A3CB2, BF5DE670, E89E5D00, 798002CB,
8D953C4B, 8472E0A1`

No entry outside this list is triaged, and no work outside these entries is
harvested. Operator constraint: **exactly ONE queued shipment**.

## Starting State (verified, not assumed)

* `backlogit shipment list` → `[]`. No active and no queued shipments.
* `backlogit queue view` → empty. Every backlog item `001-F`…`010-F` is
  `archived`. Next free IDs: feature `011-F`, shipment `010-S`.
* `backlogit checkpoint list` → 1 checkpoint, `ship`-owned, `resolved`,
  `needs_quarantine: 0`. Zero Stage-owned candidates → **zero-candidate
  normal startup**, not a recovery event.
* MCP tool surface is not exposed to this session. All backlogit operations
  run through the registry-declared **CLI fallback** (`TOOL_DEGRADED`,
  P-012-compliant — no ad hoc filesystem substitution).

## Problem Frame

All 16 entries are residue, not new product direction. Fifteen are
`DEFERRED SCOPE EXPANSION` captures produced by the review gates of
shipments `005-S`–`009-S` under P-021 C1/C2; one (`9F9A3CB2`) is an
operator-preserved working-tree carryover with an explicit
operator-authorized *stowaway* disposition.

The question is not "what should we build next" — the forward roadmap is
phases C3–C11 in
`docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md`. The
question is: **which of this residue is genuinely implementable now, in what
order, and what must be dispositioned rather than built?**

## P-021 Intake Obligations Discharged

Per P-021 C5/C6, the two independently triggered obligations were evaluated
over every `DEFERRED SCOPE EXPANSION` entry in scope.

**(A) Duplicate detection — UNCONDITIONAL, run over all 15 marked entries.**

* Result: **CLEAN with two near-miss pairs deliberately NOT merged as
  duplicates**, plus one true overlap merged.
* `DA945722` × `E89E5D00` — **true overlap, MERGED.** Both are the same root
  cause (no Windows CI runner) over two different contract surfaces
  (`tests/integration/output_path_guard_test.go` junction test vs.
  `internal/pathsafe` unit branches). `E89E5D00` explicitly anticipates and
  disclaims duplication with `DA945722`. They are not duplicate *captures*
  (neither is a re-capture of the other), so **neither is archived**; they
  are jointly resolved by one work unit.
* `11ECB954` × `9F9A3CB2` — related but distinct, NOT merged: `11ECB954`
  extends the *checker*; `9F9A3CB2` carries the *diff*. Confirmed by
  `9F9A3CB2`'s own duplicate scan.
* `798002CB` × `8D953C4B` — both touch `ContainsDotDotSegment`, but one is
  an export-surface concern and the other a validation-policy concern.
  Distinct.
* No `DISCOVERY-STATUS: AMBIGUOUS` or `LOOKUP-UNAVAILABLE` token present on
  any entry.

**(B) Late-identifier reconciliation — triggered by any `N/A` source ref.**

* `EF9352FB` was already reconciled on 2026-09-04 (PR `#4` recovered from
  `docs/closure/001-S-001-F-post-merge-closure.md`; thread ID terminal `N/A`
  because that closure record documents 0 PR review comments). **Idempotent
  no-op this cycle** — not re-reconciled, not overwritten.
* `707FE72B` already carries concrete refs (PR `22`, Copilot review thread
  on `OutputPathGuard.ps1`). No gap.
* The remaining 13 marked entries record `PR=N/A (pre-PR)` and
  `thread=N/A (no thread — local review)`. Searched the Ship-owned
  residual-risk records under `docs/closure/` for the entry IDs as join
  keys: **no late identifier surfaced for any of them.** This is recorded
  explicitly rather than left unexplained. Per P-021, these `N/A`s **STAND
  as truthful terminal records** — every one was raised by a *local persona
  review gate* that never reached a PR thread, so no identifier was ever
  created to recover. Reconciliation completes as a no-op and is **not** a
  gate on deliberation, planning, or harvest.

No Ship write was requested and no entry was destructively removed. The
single-write capture invariant and the C5 capture-only carve-out are
preserved unweakened.

## Disposition of the 3 Non-Implementable Entries

These are dispositioned **inside Stage with evidence**, not silently
omitted, and are deliberately excluded from the shipment.

### D-1 · `A92E3FA0` (high, feature) → **RETAIN as standing tracker**

Not implementable work. It is the governing product-direction record and the
**only index of five unresolved operator decisions** (Q3/H3 Dev Tunnel auth
→ blocks C10; Q6/H4 sessions-per-process → blocks C9; H5 permission
authorization → blocks C5; H6 session-pointer semantics → blocks C4; Q7
workspace tooling reach). Archiving it destroys that index.

C3 (agent adapter) is the next *feature* phase and is deliberately **NOT**
harvested this cycle: operator ordering places reliability/security work
ahead of feature expansion, and C3 builds directly on `internal/apperr`,
`internal/pathsafe`, and `internal/config` — all three of which this round
corrects. Harvesting C3 now would enlarge its blast radius against a
still-moving base. **Carried forward unchanged.**

### D-2 · `4989A42D` (medium, feature) → **RETAIN; completion condition measured and UNMET**

Its attached completion condition is: archive when **(a)** `009-F` has
closed **AND** **(b)** the U-E1a retired-architecture gate has been
broadened back to `internal/**`.

* (a) **MET** — `009-F` is `archived`.
* (b) **NOT MET** — measured directly:
  `scripts/check-retired-architecture.sh:75` still reads
  `if not path.startswith('internal/config/')` and line 391 still passes
  `git ls-files -- config.toml.example cmd/** internal/config/**`. The D6a
  narrowing is intact.

Because (b) is unmet the tracker **must be retained**; retiring it now would
remove the contamination-detection mechanism exactly as C3 begins landing
new code. Broadening the gate is *not* harvested this round: no stash entry
requests it, and doing so unbidden would be scope expansion (a declared stop
condition). Recorded here so the condition is measurable next cycle.

### D-3 · `EF9352FB` (high, task) → **BLOCKED — OPERATOR ACTION REQUIRED; isolated from the shipment**

Handled under the credential-safety constraint: **no secret material was
read, revealed, copied, or committed.** Only metadata-level, read-only
containment checks were performed, and no key value was ever loaded.

Repo-side containment was previously verified **CLOSED** and nothing in this
cycle changes it: `.env.local` has no commit history, no `.env*` path was
ever added anywhere in history, and `.gitignore` covers the pattern. There
is therefore **no in-repo remediation left and no code change to harvest.**

The residual action — rotating a live Tavily credential — requires
authenticating to an external provider console. That authority is **not held
by Stage and must not be simulated**. Per the activation constraint, this
portion is dispositioned explicitly rather than leaked or faked:
**operator-only, outside repository boundaries.**

**TIME-SENSITIVE:** the key remains in plaintext on operator disk, so
exposure persists until the operator rotates it. Deliberately isolated into
**no shipment** so it cannot stall the 13 implementable entries. This is a
reported blocker, not a silent omission.

## The 13 Implementable Entries — Options Considered

### Option 1 — One shipment, one flat covering feature (13 sibling tasks)

Simple manifest. **Rejected:** mixes Go internals, CI topology, PowerShell
scripts, ignore policy, and governance docs as flat siblings with no
expressed cohesion, and gives the executor no ordering signal beyond ID
sequence. Encodes no dependency between the checker and the ignore-policy
change it must validate.

### Option 2 — One shipment, covering feature + 7 domain sub-epics (CHOSEN)

Groups the 13 entries by **shared code surface and artifact class**, and
orders the sub-epics by the operator's product ranking. Satisfies the
one-shipment constraint while preserving width isolation per task and
encoding a real execution order. **Chosen.**

### Option 3 — Multiple shipments split by artifact class

Cleanest artifact-class isolation (Go internals / CI / harness bootstrap).
**Rejected: violates the hard one-shipment constraint.** Recorded because it
is the honest runner-up, and because artifact-class isolation is a genuine
value being traded away — mitigated in Option 2 by sub-epic boundaries and
per-task width isolation rather than by separate manifests.

**Constraint-vs-policy check:** the one-shipment constraint was tested
against mandatory policy and does **not** force a violation. Width isolation
and the 2-hour rule are *per-task* constraints, satisfiable inside a
multi-sub-epic feature. P-016 is unaffected (one sequential unit, no
parallel worktree). Artifact-class isolation is a *grouping preference*, not
a NON-NEGOTIABLE principle. **No halt required.**

## Decisions

### D-A · Covering feature = residual hardening, not feature expansion

`011-F` covers exactly the 13 implementable entries. No C3 work, no new
product capability. This honours "reliability and security over features"
and "simplicity over complexity."

### D-B · `BF5DE670` is RESCOPED from *mitigate* to *consolidate* — the decisive evidence

The entry asks for a mitigation plan (Lstat-before-write, `O_EXCL`
revalidation, hardlink-count checks) for three residual risks: the GO-14
write-through-dangling-symlink primitive, the Resolve→use TOCTOU window, and
the `EvalSymlinks`-ignores-hardlinks gap (SEC-5).

**Measured fact:** a scan of every non-test `.go` file under `internal/` for
`os.WriteFile|Create|OpenFile|Remove|RemoveAll|Rename|Mkdir` returns
**zero matches**. `pathsafe` currently gates **no destructive write path at
all** — there is no call site to protect.

Building TOCTOU mitigation now would be speculative engineering against an
API that does not yet exist, and would very likely be rebuilt at C4–C6 when
real write paths land. The entry's own trigger is *"before pathsafe gates
any live destructive file-write path"* — that trigger has **not fired**.

**Decision:** deliver the *consolidation* the entry also asks for (a single
tracked risk register unifying the three currently-scattered doc comments)
plus an explicit, checkable **precondition gate** that fires when the first
write path appears. Defer the mitigation mechanism itself to that trigger.
Cheaper, honest, and directly supported by "simplicity over complexity."

### D-C · `A0A2D049` is promoted to FIRST — it is a live safety exposure, not a policy nicety

The entry reads as testing policy. Inspection shows something sharper:
`internal/copilotprobe/testhelpers_test.go:77` skips **only when `Start()`
fails**. Where ambient credentials are present, the probe suite makes live,
credential-consuming SDK calls *and* `permission_test.go` deliberately
provokes **real shell-command execution**. The gate today is *implicit
credential availability* — meaning any developer or CI runner that happens
to hold credentials executes this silently.

That is an active safety hole under Constitution VII/VIII, so it is ordered
**first**, ahead of every other entry regardless of its `medium` provisional
priority.

**Conflict resolved:** `309FBF5A` proposes *deleting* `internal/copilotprobe`
at C3, which would discard this work. Deleting now is rejected — it is
explicitly C3-owned forward scope, and C3 is not scheduled this round.
Adding a small explicit opt-in gate is cheap, is correct while the package
lives, and dies harmlessly with it. **Safety now beats efficiency later.**

### D-D · `309FBF5A` is SPLIT along the code/planning seam

* **(ii) mechanical import-boundary gate → BUILD NOW.** `.golangci.yml`
  already exists and golangci-lint v2.13.2 is already pinned in the `lint`
  job, so a `depguard` rule is native and near-free. It must **allowlist
  `internal/copilotprobe`** (today's sole importer) or it cannot pass; the
  allowlist becomes the exact deletion checklist at C3. This also *de-risks*
  C3, whose published exit contract already demands "no package outside
  `internal/agent` imports `copilot-sdk/go/rpc` (mechanical gate)."
* **(i) "add a C3 planning acceptance criterion" → DOCS.** Forward planning,
  correctly C3-owned. Discharged as a plan amendment, not code.

### D-E · Execution order: the checker precedes the ignore-policy change

**Measured fact:** `git show HEAD:.gitignore` contains **zero** negation
lines. All six (`!.env.example`, `!plugin/.mcp.json`, and four `.engram/*`)
are introduced by the `9F9A3CB2` stowaway.

So the stowaway is the **first-ever introduction of negation semantics into
this repo's ignore policy** — precisely the un-ignore hazard class that
`11ECB954` flags as unguarded by the append-only checker, and precisely what
`9F9A3CB2`'s own validation obligation (c) demands be cross-checked.

**Decision:** harden the checker **before** the `.gitignore` stowaway ships,
so the checker is the instrument that validates it. The new rule must detect
a negation that *re-includes a path already ignored by an earlier pattern*
(a true un-ignore regression) — it must **not** ban negations outright,
which would be both wrong and a scope overreach.

If the stowaway's negations trip the hardened rule, the
operator-authorized stowaway protocol governs verbatim: **drop it from the
PR and return it to the stash — do not block the feature and do not discard
the change.**

### D-F · `9F9A3CB2` is SPLIT into ignore-policy and launcher units

The entry itself recommends this split, and the two halves are independent
with different artifact classes and different risk. Both ride as
**disclosed stowaways** per the operator-authorized disposition — this
shipment *is* the "next feature round" that disposition names.

`start.ps1` is an **autoharness-generated** file whose
`Generated by autoharness` marker was deleted while carrying two genuine
improvements (bounded Engram pre-warm budget; `Push`/`Pop-Location`
anchoring) and **eight capability regressions**. Reconciliation direction:
**restore the dropped HEAD capabilities and the marker, keep the two
improvements.**

**Boundary disposition:** pushing the improvements upstream into the
autoharness template lives in `softwaresalt/autoharness`, a **different
repository**. That is outside this repo's boundary and is **not harvested**;
it is recorded as a follow-up so the value is not lost to the next
regeneration.

### D-G · Windows CI is added, but `-race` is omitted on that runner

`E89E5D00` + `DA945722` are jointly resolved by one Windows job. Prior
learning `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md` is
directly load-bearing: `windows-latest` provides **no cgo by default**, and
`go test -race` **fails closed (exit 2)**, it does not silently skip. A
Windows job naively copying the Linux `-race` step would hard-fail CI.

**Decision:** Windows job runs unit tests **without** `-race`; race coverage
remains Linux-only and that limitation is stated explicitly. New actions
must be **full 40-hex-SHA pinned** per `ci-security.instructions.md`, and
the job must be added to the `ci-gate` `needs:` array (currently 8 entries)
or it is decorative.

### D-H · `8472E0A1` is split by skill domain, not shipped as one task

Its eight items (a)–(h) span tests, refactors, and doc notes. As a single
task it would violate width isolation, so it is decomposed into
domain-isolated subtasks.

### D-I · `8D953C4B` resolution is chosen at implementation, both branches pre-authorized

Rule 7 permits an absolute `database.path` with only a literal `..` scan,
conflicting textually with Constitution III. Two legitimate resolutions:
constrain the path under workspace containment, or record a formal
Constitution Check exception. Constraining is preferred (it removes the
divergence rather than blessing it), but `database.path` may legitimately
need to live outside the workspace, and D7 defers persistence design.
**Either outcome is acceptable; an undocumented divergence is not.**

## Execution Order (sub-epics, product-ranked)

| # | Sub-epic | Entries | Rationale |
|---|---|---|---|
| A | Live safety exposure | `A0A2D049` | Active unguarded credential + shell execution |
| B | Containment correctness | `8D953C4B`, `BF5DE670`, `798002CB`, `8472E0A1` | Security/constitution correctness in shipped code |
| C | Platform verification coverage | `E89E5D00`, `DA945722` | Security branches currently unverified |
| D | Script guard correctness | `707FE72B` | Fail-safe defect; real but non-exploitable |
| E | Supply-chain & boundary gates | `11ECB954`, `AFE0E95C`, `309FBF5A`(ii) | Preventive gates; E precedes F per D-E |
| F | Harness bootstrap carryover | `9F9A3CB2` | Stowaways; validated by E's hardened checker |
| G | Governance docs | `6E701953`, `309FBF5A`(i) | Docs last per product ranking |

Ordering obeys: reliability/security → features → docs; and the one hard
internal dependency **E → F** established by D-E.

## Open Questions (non-blocking)

1. Should the U-E1a gate be broadened to `internal/**` (closing `4989A42D`)?
   Deferred — not requested by any in-scope entry; raising it here so it is
   measurable next cycle.
2. Does `database.path` legitimately need to escape the workspace root
   (D-I branch selection)? Resolved at implementation; both branches
   pre-authorized.
3. Upstream autoharness template push for the `start.ps1` improvements
   (D-F) — different repository, out of boundary.

## Outcome

**13** entries → one covering feature `011-F` across **7** sub-epics, in one
queued shipment. **3** entries dispositioned in Stage with evidence
(`A92E3FA0` retain, `4989A42D` retain with measured unmet condition,
`EF9352FB` blocked/operator-only). Proceed to
`docs/plans/2026-09-06-intercom-go-residual-hardening-round2-plan.md`.

---

# Addendum (same session, 2026-09-06) — post-review decision corrections

The multi-persona plan review gate returned **FAIL** on plan revision 1
(4×P0, ~20×P1). Three findings invalidated reasoning in *this* deliberation,
not merely in the plan, so they are corrected here.

## D-I **REVISED** — `8D953C4B` resolves via branch (c), not (a) or (b)

The original D-I pre-authorized two branches and *preferred* branch (a)
(route `database.path` through `pathsafe.NewRoot`/`Resolve`). Review found
**both branches defective**:

* **Branch (a) is infeasible against the current API.** `pathsafe.normalize`
  rejects *every* absolute candidate, and `NewRoot` requires an
  **already-existing directory**. `database.path` is a not-yet-existing file
  (default `data/agent-rc.db`; `data/` does not exist), and rule 7
  deliberately permits absolute paths as "an explicit, visible operator
  privilege". Branch (a) would therefore reject *legitimate in-root absolute
  paths* and force a new `pathsafe` capability — colliding head-on with
  invariant J2. The original escape condition ("select (b) only if a test
  shows (a) breaks a legitimate operator configuration") would never have
  fired, because the failure is **API-shaped, not config-shaped**.
* **Branch (b) alone is unsafe.** Principle IV is NON-NEGOTIABLE and has no
  write-side exception mechanism, so an unbounded exception could
  pre-authorize a breach the moment persistence lands.

**Revised decision — branch (c):** record a **dated, bounded Constitution
Check exception that EXPIRES at the first persistence write path**, and
register containment as a mandatory precondition on the D7/C4 persistence
work.

This is not a compromise but the factually correct position: Principle IV is
**not breached today**, because the same measurement underpinning D-B shows
*nothing opens or writes* `database.path`. The divergence is **latent**. The
exception's expiry trigger is bound to the *same* mechanical gate as D-B's
precondition, so it cannot outlive its own basis.

**Consequence for `798002CB`:** under branch (c) the rule-7 call site
survives, so `pathsafe.ContainsDotDotSegment` does **not** become dead
exported API. The dead-API interaction review raised between `8D953C4B` and
`798002CB` is thereby dissolved rather than merely sequenced, and `798002CB`
stays a docs-only unit.

## D-E **RE-SCOPED** — correct ordering, wrong rule and wrong granularity

The ordering conclusion (harden the instrument before the subject) **stands
and is confirmed by evidence** recorded literally here at review request:

```text
$ git show HEAD:.gitignore | grep -c '^!'
0
```

All six negations are introduced by the stowaway. Two corrections:

1. **The rule definition was wrong.** D-E specified detecting "a negation that
   re-includes a path already matched by an earlier ignore pattern" — which is
   the definition of *every functional gitignore negation*, since git only
   gives a `!` line effect when a preceding pattern already ignores the path.
   Implemented literally it would have banned all effective negations: exactly
   the outcome D-E itself forbids. **Replaced by an old-vs-new behavioural
   differential** — no path *actually ignored at base-ref* may become
   un-ignored at head-ref — evaluated with **git's own matcher** rather than a
   gitignore engine reimplemented in bash.

   This also resolves the false-positive risk: the six negation targets do not
   exist, so a behavioural differential correctly does **not** flag them, while
   still catching a genuine regression. The predicted "stowaway almost
   certainly dropped" outcome is therefore **withdrawn**.

2. **The dependency was declared at the wrong granularity.** The real edge is
   `011.012-T → 011.015-T` only. `AFE0E95C` (installer pinning) and
   `309FBF5A`(ii) (depguard) bear no relation to the stowaway and were
   gratuitously ordered behind sub-epic F. Restated at task level.

   Invariant J9's enforcement claim is also corrected: the CI checker compares
   **base→head**, so intra-PR commit order has no effect on its verdict. What
   matters is that both changes are present in the same head state with all
   six verdicts recorded — the ordering's value is evidentiary and
   local-development sequencing, not a mechanical gate.

## D-J **NEW** — the four preserve-only ignore lines land FIRST, not last

Review identified a sequencing exposure created by the original E→F ordering:
the four confirmed-NOT-IGNORED preserve-only paths would have stayed
un-ignored for the **entire** shipment, two of which carry machine-local
command/environment payloads, with only a manual `git status` check standing
between them and an accidental `git add -A`.

Those four lines are **pure additions**: they cannot violate I6's
ordered-subsequence predicate and cannot trip any un-ignore rule. They are
therefore **not stowaway content and not droppable** — they are ordinary
shipment hardening, split out and sequenced **first** (`011.001-T`), ahead of
every other unit. Only the six negation lines remain behind the
`011.012-T → 011.015-T` dependency, which is the only part D-E's rationale
ever required.

## Scope correction — `9F9A3CB2` is PARTIALLY resolved

`9F9A3CB2`'s operator-authorized include list names **eight** paths; the
original plan covered only `.gitignore` and `start.ps1`. Archiving the entry
on that basis would have silently dropped six operator-preserved paths,
including two tracked backlog-state files the entry explicitly requires be
disclosed as "backlog-state carryover, NOT feature content".

`9F9A3CB2` is therefore recorded as **partially resolved** and **remains
active** in the stash until Ship's PR disclosure completes.

* `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`
  — load-bearing for D-G's `-race` omission (`windows-latest` has no cgo;
  `go test -race` fails closed with exit 2). *(Path corrected in revision 3;
  the original D-G text cited a non-existent `docs/compound/` root path.)*

## Unchanged

D-A, D-B, D-C, D-D, D-F, D-G, D-H stand as written. In particular D-B's
rescope (consolidate, do not mitigate) was **independently confirmed** by two
reviewers who re-ran the write-path measurement, and D-C's promotion of the
live-SDK gate to first was confirmed against
`internal/copilotprobe/testhelpers_test.go`. The three non-implementable
dispositions were audited and found **legitimate and evidence-based, not
disguised omissions**.
