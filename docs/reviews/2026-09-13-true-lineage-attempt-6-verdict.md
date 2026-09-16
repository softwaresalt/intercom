# TRUE-LINEAGE ATTEMPT 6 — plan-review verdict

**Artifact**: `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md` revision 10
**Date**: 2026-09-13 · **Attempt**: 6 · **Dispatch**: multi-agent, 7/7 personas returned
**Anchor**: Architecture Strategist — `openai` / `gpt-5.6-sol` / `high`

## Verdict

**FAIL** — `P0 = 1`, `P1 = 7` (deduplicated), `P2 = 23` raw, `P3 = 21` raw.

**No attempt 7 is authorized.** The plan was **not patched** in response to this verdict, per
explicit operator direction. This file is the compact record; the governing plan carries only a
three-line gate-state marker pointing here.

## Per-persona results

| Persona | Route | Verdict | P0 | P1 | Lead findings |
|---|---|---|---|---|---|
| Architecture Strategist **(anchor)** | `gpt-5.6-sol` / high | FAIL | 0 | 2 | `A-1`, `A-2` |
| Constitution Reviewer | default | FAIL | 0 | 2 | `C-1`, `C-2` |
| Go Reviewer | default | FAIL | 0 | 1 | `G-1` |
| Scope Boundary Auditor | default | FAIL | 0 | 1 | `S-1` |
| Security Lens Reviewer | default | FAIL | 0 | 1 | `SEC-1` |
| Agent-Native Parity | default | FAIL | **1** | 3 | `P-1`, `P-2`, `P-3`, `P-4` |
| Learnings Researcher | default | FAIL | 0 | 2 | `L-1`, `L-2` |

## Deduplicated blockers

### K-1 — P0 — the installed skill exposes no pre-mutation verdict boundary

*Raised independently by 5 of 7 personas (`A-1`, `SEC-1`, `S-1`, `P-1`, `P-2`) and verified
directly by Stage against the installed file.*

The plan's central safety guarantee (§3.5, §10.1, AC-14, and condition 1 of both `PA-*`
records) is **"a reported non-`CASCADE` verdict ⇒ HALT BEFORE ANY CLOSE MUTATION."** Measured
against `.github/skills/shipment-reconcile/SKILL.md`:

* **L510** — on a non-`CASCADE` selection Step 0(c) *"continue[s] to step 1 below"*, entering
  the mutating Safe-Close steps 1–10 directly. Step 4 archives every manifest item; step 8
  ships and archives the shipment record.
* **L821** — the verdict is recorded **only** in the cascade-close report, at Cascade Close
  Sub-Procedure **step 5**, i.e. *after* the destructive `backlogit_ship_shipment` call.
* The safe-close report records no verdict token at all, so `SAFE_CLOSE` is **never an
  observable value**; its real signature is *verdict absent*, detectable only post-closure.

Consequences: the HALT is unreachable; §10.1's row *"no CLOSE mutation has occurred; `021-S`
stays `active`, is not shipped and not archived"* is **factually false**; a close could execute
outside `PA-021-CASCADE`'s approved action (Principle VII); and §10.3's restore is scoped to
post-`CASCADE` failures only, leaving an unintended safe-close execution with **no defined
recovery**.

This is the **same defect class as `J-5`** — asserting a contract the installed tooling does not
implement — relocated from the classification API to the mutation boundary. `J-5`'s remediation
corrected *where the verdict comes from* but not *when it becomes readable*.

### K-2 — P1 — CS3, CS5 and CS6 are not independently verifiable

`A-2`, `G-2`, `G-3`. Row 9's two needles both sit inside the **CS4** block, yet AC-3 credits row
9 as proof for CS3 and CS5 as well. CS6 has no anchor and no dedicated needle. An H1 that edits
only CS4 yields a fully green suite with 2–3 of 10 mandatory sites unapplied. AC-4's "no
classification predicate" claim is likewise undetectable: four measured survivor sentences in
the CS4 block (`it is a root (no parent_id)` ×1, `The cascade close path is permitted` ×1,
`childlessness is positively verified` ×1, `The manifest must contain nothing beyond` ×1) are
pinned by no row. This is `J-12`'s defect class surviving its own remediation.

### K-3 — P1 — row 2's needle is unsatisfiable by the plan's own pinned wording

`G-1`, `S-5`. Row 2 pins `all five conditions are conjunctive` (lowercase). The only pinned
grant wording (§4.1 CS10, which ADD-1 must match) reads *"all five conditions in the Ship Role
Boundary hold"* — not a match; §3.2's own prose is *"**All** five conditions are conjunctive"* —
also not a match for a case-sensitive `strings.Contains`. An executor applying the pinned
wording leaves row 2 **red at H1**, so the plan's red→green claim is not derivable from its own
normative text. Additionally, nothing pins the `at every depth` quantifier or the set-equality
conjunct **in `_ship.agent.md`**, so the agent file could land a weaker grant than §3.2.1 with
the suite green — the exact drift this feature exists to fix.

### K-4 — P1 — `PA-021-CASCADE` names the wrong execution site

`C-1`, `P-4`. `.backlogit/queue/021-S.md` records *"Execution site: plan section 9 step 16."*
The plan places the cascade at **§9 steps 13a–15**; step 16 is operational closure / P-020
compaction. §11 states any mismatch between a recorded condition and live state voids the
approval, so the record is self-invalidating. This is `J-7`'s defect class reproduced in the
coupled record while being fixed in the plan — i.e. a **scope-of-sweep** failure, the `J-2`
pattern.

### K-5 — P1 — the rollback is itself an unapproved destructive operation

`C-2`. §10.3.2 step 2 unconditionally runs `git restore --source {pre_cascade_sha}` (unbacked
overwrite) **and** deletes `{post_paths} − {pre_paths}` (file deletion), with no approval gate —
while §11 states the operator authorization *"covers exactly the two actions below and nothing
else."* The plan declares its own mandatory recovery path unapproved. Principle VII is
NON-NEGOTIABLE and names deletion and unbacked overwrite explicitly.

### K-6 — P1 — §9 steps 1/1a produce a manifest the installed intake check cannot represent

`P-3`. After step 1 claims `021-S` and step 1a requires `022-F` to be exactly `active`, the
manifest is `[022-F active, 022.001-T queued]`. Installed Step 0.5 item 6 (`_ship.agent.md`
L253–262, the CS7 site) runs `mode: pre` with a **single** `expected_status`, so one member is
necessarily `status-mismatch` ⇒ `RECONCILE_FAIL`. The installed Scope note at L259–262 says this
outright. The plan edits that clause but never sequences the check in §9.

### K-7 — P1 — the measurement discipline omits the scope where every failure occurred

`L-1`, `L-2`. §13.4 declares Scope I (the two instruction files) and Scope A (plan + 4 records).
The **installed skill, the backlogit engine and its index schema are in neither** — yet §3.4.1,
§5 and §3.4.2 all make measurement assertions about exactly that surface, and `H-3`, `H-6`,
`J-5`, `J-9` and now `K-1` all failed **there**. §13.4 states the rule *"name the scope, give the
regex, give the number"* and then never applies it to the scope that keeps breaking. §13.5's
bullet 1 also mis-states its source: the cited compound doc teaches *independently reproduce a
claim about installed behaviour*, not *declare your count scope* — and citing the weaker lesson
is how the stronger one was evaded again.

## What revision 10 did resolve

Verified RESOLVED by multiple personas against the live tree: **`J-1`** (compact, audit-linked
lineage), **`J-2`** (five-file scope declared; all four records reauthored clean), **`J-3`**
(both path inventories defined with command, roots and capture moment), **`J-4`** (CS10 is
genuinely surgical — one P-010 bullet, no release-instance ID, intersection not union),
**`J-6`** and **`J-11`** (§9.1 defers faithfully to the installed protocol; all three never-prune
classes), **`J-7`** in the plan (escrow binding), **`J-8`**/**`J-15`** (plan-pinned blob OIDs and
engine digest, classifier byte identity), **`J-9`** as a prohibition, **`J-10`**, **`J-13`**,
**`J-14`**, **`J-16`**, **`J-17`**, **`J-18`**.

Independently re-measured clean: all 11 clause anchors resolve **exactly once**; all §4.2
ordinals exact (ADD-1 38, CS7 255, CS1 756, ADD-2 767, CS8 777, CS2 790, CS9 792, CS3 794,
CS4 802, CS5 822, CS10 241); every §6.1 H0 count correct; `go.mod` free of `testify`.

**`J-5` is the one P1 that did not survive scrutiny**, and it carried `K-1` with it.

## Residual blockers — exact

`021-S` is **not claimable**. To reach a passing gate a future authorization would need to
resolve, at minimum:

1. **`K-1`** — either add a read-only pre-mutation classification boundary to
   `shipment-reconcile` (which expands the implementation surface to a **fourth file** and is
   Ship-runtime work, not Stage planning work), **or** rewrite §3.5 / §10.1 / AC-14 and both
   `PA-*` condition-1 clauses to post-hoc validation plus a verified restore obligation for an
   unintended safe-close execution.
2. **`K-2`**, **`K-3`** — repin needles so every mandatory site is independently falsifiable and
   row 2 is satisfiable by the pinned wording.
3. **`K-4`** — one-line correction to `021-S`.
4. **`K-5`** — a third approved-action record, or an operator-approval gate before the restore.
5. **`K-6`** — sequence the intake check with a representable `expected_status`.
6. **`K-7`** — add **Scope T** (installed tooling) with command + verbatim output + pinned digest
   for every tooling claim.

## Audit

Full persona transcripts are not retained inline by design. Lineage: commits `3e0cf8f`,
`3423581`, `e9019d4`, this commit; PR #54; `docs/memory/`; `.backlogit/checkpoints/`.
