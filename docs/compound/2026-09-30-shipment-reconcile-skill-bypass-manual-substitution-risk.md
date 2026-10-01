---
title: "shipment-reconcile without an installed CLI gate: manual classification is not license to skip the safe-close binding"
date: 2026-09-30
category: process-compliance
tags: [backlogit, shipment-reconcile, p-005, p-021, dark-mode, closure]
---

# A skill document's mode-dispatch sequence must still be followed even absent a CLI wrapper

## Symptom

During 035-S's post-merge closure, no dedicated `classify-close-path`
autoharness gate/CLI was installed in this workspace (established precedent
from 034-S: "manual structural classification remains the only method").
That true fact was over-generalized into skipping a second, independent
requirement: the `shipment-reconcile` skill's mandated sequence is
`mode: classify-close-path` → obtain a `CLASSIFICATION_BINDING` (a SHA-256
digest over the canonical pre-close snapshot) → `mode: safe-close` carrying
that binding, which revalidates it immediately before any mutation (a TOCTOU
guard) and only then enters the Cascade Close Sub-Procedure. Absence of a
CLI wrapper for *classification* does not remove the requirement to compute
and carry a binding before invoking the *cascade primitive*.

What actually happened: the classification predicates (root check, full-depth
descendant coverage, no grandchildren, no linked deliberation) were evaluated
manually and correctly, reaching the right verdict
(`CASCADE` / `FULLY_COVERED_ROOT`). But the cascade primitive
(`backlogit shipment ship 035-S ...`) was then invoked **directly**, with no
binding ever computed and no `mode: safe-close` call in between. Copilot
review caught this on the follow-on post-merge-closure PR (#84) as a Medium
finding: "The closure bypassed mandatory classification-binding
revalidation and directly invoked the cascade operation."

## Root cause

Two genuinely separate gaps were conflated into one:

1. **No CLI/gate wraps `mode: classify-close-path`** — true, and correctly
   handled by performing the equivalent structural checks manually, per the
   skill's own "other workspaces implement the equivalent check directly"
   guidance.
2. **The skill's own mode-dispatch discipline — compute/carry a binding,
   never call the cascade primitive except from inside a bound
   `safe-close`** — this is a *procedural* requirement of the skill
   document itself (SKILL.md, not a CLI feature), and has no "no gate
   installed" exception. It must be followed by literally executing the
   skill's documented algorithm step by step (including the binding
   computation), not by substituting a different, narrower manual check
   (classification only) and then treating the primitive as directly
   callable.

Because gap #1 had a clean, previously-validated manual substitute, it was
easy to implicitly extend that same "manual is fine here" reasoning to gap
#2, where no such substitute is sanctioned — the binding and its
revalidation exist specifically to guard against drift between
classification and mutation, and skipping them removes that guard even if,
in this particular execution, no drift occurred.

## Disposition

Not retroactively repairable (the shipment was already correctly archived),
and **not** fully re-verified either: the safe-close report's own
Steps 2–4 (two-set gate over `allowed_ids`/`required_ids`/`archived_ids`,
`returned_ids` empty, `parent_id` preservation re-read post-close) confirm
**final archive consistency** — the output set matches the manifest and
structural preservation held — but that is a narrower guarantee than what
`CLASSIFICATION_BINDING` covers. The binding also snapshots shipment
dependencies/status, skill and engine identity, and each member's declared
type, pre-close status, and resolved location (SKILL.md §Canonical Binding
Format); Steps 2–4 cannot reconstruct those inputs from the archived output
alone, so the residual uncertainty from skipping the pre-mutation
revalidation of those specific inputs is **retained, not closed**, by the
post-hoc checks. The gap was disclosed transparently in the reconcile report
and the closure artifact's releasability evidence rather than silently
corrected or hidden — per Copilot's explicit ask:
"record the process violation and obtain an explicit disposition... rather
than presenting this as a conforming skill run."

## Fix / Prevention

When invoking `shipment-reconcile` for a shipment close and no CLI/gate
wraps the skill:

1. Still execute `mode: classify-close-path`'s full algorithm manually,
   INCLUDING computing the canonical `CLASSIFICATION_BINDING` digest per
   the skill's documented format (`v1` + shipment + verdict + reason +
   skill-hash + engine-version + sorted manifest + deps + status + one
   line per snapshot member, SHA-256 of the UTF-8 bytes) — not just the
   classification predicates.
2. Never call the cascade primitive (or any direct archival mutation) outside of a live
   `mode: safe-close` invocation. The binding is not a standalone artifact to compute, record,
   and then present as authorization for a separate direct call — `mode: safe-close` itself must
   actually run: it recomputes the binding from a **freshly-taken** snapshot immediately before
   mutating, revalidates it against the value just classified, and only then enters the Cascade
   Close Sub-Procedure internally. A manually-computed binding that is merely recorded alongside
   a direct cascade call reproduces the exact same bypass this entry exists to prevent, just with
   a binding-shaped artifact attached — it does not perform the mandated live revalidation and
   must not be treated as equivalent to it.
3. If genuinely uncertain whether a "manual equivalent" is sufficient for
   any given mode (classification vs. safe-close vs. cascade), re-read the
   skill's own text for that mode section specifically — "no CLI gate
   installed" is a fact about tooling, not a blanket license to skip
   skill-mandated procedural steps that have no CLI dependency at all.
4. Cross-reference: this complements (does not duplicate) the existing
   034-S precedent note about manual classification being an accepted
   substitute for the *missing CLI* — this entry narrows that precedent to
   classification-only and explicitly excludes the binding/cascade
   sequencing from the "manual substitute is fine" umbrella.

## Recurred in 036-S (same category of gap, not successfully applied — corrected 2026-10-01)

036-S's post-merge closure (feature 046-F, M3 unignore-regression/merge-strategy Go port)
**reproduced this exact gap**, not applied this entry's fix/prevention steps successfully. The
canonical `CLASSIFICATION_BINDING` was computed per the skill's documented `v1` format during
the manually-executed `classify-close-path` step, but the subsequent `backlogit shipment ship`
call was a **direct** invocation of the cascade primitive, with no intervening `mode: safe-close`
call to freshly recompute and compare that binding. This is exactly the pattern item 2 above
warns against: "a manually-computed binding that is merely recorded alongside a direct cascade
call reproduces the exact same bypass this entry exists to prevent... it does not perform the
mandated live revalidation and must not be treated as equivalent to it." An earlier draft of
this entry incorrectly claimed this closure "followed this entry's fix/prevention steps in full"
and that "no process-compliance gap was disclosed" — both claims were inaccurate and were
corrected (via Copilot review on PR #86) in `docs/closure/036-S-046-F-post-merge-closure.md`'s
Risky Action Record and Releasability evidence, which now accurately disclose the gap.

**Net assessment across both occurrences**: 035-S computed no binding at all; 036-S computed and
recorded a binding but never used it to gate or revalidate the subsequent mutation. Both are the
same root-cause category — no conforming `mode: safe-close` call ever actually ran — and the
difference between them is cosmetic (a binding value present in the record) rather than
substantive (the live revalidation step was absent both times). This confirms the entry's
prevention guidance was *correctly written* (item 2 predicted exactly this failure mode) but was
not *followed* a second time, reinforcing that a narrative claim of compliance is not a substitute
for literally invoking the mandated step. A stash entry (`D10D3AFC`, captured during 036-S's own
closure review) proposes a structural remediation: make `mode: safe-close` actually invokable (a
real CLI/gate subcommand) so the live recompute-and-compare step cannot be silently approximated
away by an agent reasoning about "no state changed, so recomputation would be identical" — that
reasoning is precisely the failure mode to guard against, since it is unverifiable from inside the
same session that is making the claim. See
`docs/closure/036-S-046-F-post-merge-closure.md` and
`.backlogit/reconcile/036-S-classify-close-path-2026-09-30T19-21-40Z.md` /
`036-S-cascade-close-2026-10-01T02-25-49Z.md` for the full binding computation and the corrected
disclosure.

## Partially applied in 037-S (revalidation performed; tool enforcement still absent)

037-S's post-merge closure (2026-10-01, closure PR #88; implementation PR #87) applied item 2 of the fix and prevention guidance.

* `classify-close-path` produced `CLASSIFICATION_BINDING`
  `0f9be8a39309d70cd01b641871f84de68fc0abe5a80bce03fe394a0c230792d9` from a snapshot at
  18:20:09Z. A post-close review later found that this value encoded member location wrongly;
  the corrected binding is `77985d03e202c5ff61f1ba431d6d087aa7f2b349395eeceaadd3ca41b263a5b4`.
* Before the cascade mutation, a fresh snapshot was taken at 18:20:12Z and the binding was
  recomputed independently. The two values matched exactly.
* `backlogit shipment ship` ran only after that match. The two-set gate then passed:
  archived ids equal allowed ids equal required ids, a set of 11.

**Same-helper blind spot (new evidence from 037-S).** One session-local helper produced both
snapshots, and it wrongly encoded every member's `location` as `archive` when the members were
actually in `queue`. Because both snapshots shared the defect, the recompute-and-compare
matched without validating location. The verdict happened to be unaffected, since location only
feeds the ambiguity check. The lesson is general: recomputing a binding with the same
agent-authored code only proves that the code is deterministic, not that it is correct. Real
revalidation needs a tool-enforced implementation, or at minimum an independent cross-check
against filesystem or archive provenance.

This is the first closure since this entry was written to perform the recompute and compare,
rather than only recording a binding.

The residual is unchanged in kind:

* No invokable `safe-close` exists, so the recompute and the comparison were agent-executed by a
  session-local helper, not a tool-enforced atomic call.
* Under two minutes separated the fresh snapshot from the mutation, in a single-agent session
  with no concurrent backlog writers.

Stash `D10D3AFC`, an invokable `safe-close` CLI or gate, therefore stays relevant. See
`docs/closure/037-S-047-F-post-merge-closure.md` and
`.backlogit/reconcile/037-S-cascade-close-2026-10-01T18-21-40Z.md`.