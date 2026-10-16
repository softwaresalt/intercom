---
title: "retiredarch closure: sibling-file side effects (3750C37C) bounded threat model"
description: "Stage deliberation over stash 3750C37C (P-021 C6): whether the retiredarch package-closure pin should constrain non-test sibling side effects such as init() spawning git, or whether the control belongs to review"
topic: "retiredarch pin.go packageClosureOK/closureFileOK does not constrain arbitrary side effects in a non-test sibling file (func init running os/exec git update-index), so the git index can be narrowed before selectRepoPaths while SelectionPathspecPin stays OK"
depth: "standard"
decision_status: "decided-pending-operator-confirmation"
promoted_to: "none (no harvest this session)"
linked_artifacts:
  - "docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md"
  - "docs/decisions/2026-10-08-intercom-go-gatecheck-batch-a-correctness-deliberation.md"
  - "docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md"
  - "docs/closure/040-S-050-F-post-merge-closure.md"
tags:
  - "gatecheck"
  - "retiredarch"
  - "threat-model"
  - "p-021"
  - "stage"
---

# Deliberation: retiredarch sibling-file side effects (3750C37C)

## Session context

* **Agent and mode:** Stage, invoked by the Orchestrator (sequential, not dark). Routing directive
  P-013.5 resolved to `gemini-4-argon` / google / high / long_context (`config.model_routing.stage`).
  The runtime could not switch models, so `ROUTING_DEGRADED` is declared; the session ran on the
  invoking model.
* **Selection:** the Orchestrator named 3750C37C as a highest-priority deliberate-first item.
* **P-021 C6:** the entry carries the `DEFERRED SCOPE EXPANSION` marker, so this artifact is
  mandatory before any planning. The prior Stage annotation (D-RA-7, 2026-10-07) already returned
  the entry "to deliberation with a BOUNDED THREAT MODEL before any re-plan".
* **Duplicate scan (A):** CLEAN. Nearest neighbours `0ECC1895` and `DC921AF6` are the other
  D-RA-7 held pin trust-root entries (same hold, different expansions).
* **Late-identifier reconciliation (B):** the entry's PR and review-thread were `N/A`. The Ship
  closure records `docs/closure/030-S-033-F-post-merge-closure.md` and
  `docs/closure/040-S-050-F-post-merge-closure.md` cite the entry as a residual. No late review
  thread or PR identifier was found; the `N/A` stands as a truthful terminal record.

## Problem frame

The pin (A-T3d) enumerates a closed set of source-text rules for the retiredarch package. A
non-test sibling file with `func init()` can run `os/exec` and mutate the git index before
`selectRepoPaths`, so the gate sees a narrowed listing while `SelectionPathspecPin` still passes.
The entry records that this existed identically on main at 698e2bc (no regression).

## Options

| Option | Description | Cost / risk |
|---|---|---|
| A | Allowlist or token-freeze every non-test file in the package (extend the pin) | Heavy. The compound learning (2026-10-07) records an unbounded AST-rule review loop for exactly this class; each new rule moves a frozen set that must be refrozen atomically with `select.go`. |
| B | New AST rule banning `init`, `os/exec` and process spawning in non-test siblings | Different contract surface (open-ended runtime side effects vs enumerated text rules). Changes gate verdicts beyond any AC (INV-1). A denylist is bypassable by indirection. |
| C | Required CODEOWNERS review on `tools/gatecheck/**` plus branch protection | Needs an operator branch-protection action and the R1 unattended-approval path. Does not change code. |
| D | Accept as a bounded, documented residual: any code merged into the gate package already runs with the gate's authority, so a pin cannot defend the package from itself | No code change; must be stated so the residual is not mistaken for a closed hole. |

## Bounded threat model

* **In scope:** accidental drift and an unreviewed change that reaches the gate package.
* **Out of scope for a source-text pin:** a malicious or careless author who can add a Go file to
  the gate's own package. That author can also edit `pin.go`, `select.go` or the workflow, so a
  closure rule is defence in depth only and cannot be the trust root.
* **Therefore** the trust root is change review of `tools/gatecheck/**`, not ever-stricter pin
  rules.

## Decision (D-3750-1)

Recommended and recorded: **C + D**, reject A and B.

1. Record the residual as accepted under the bounded threat model above (Option D).
2. The durable control is required review of `tools/gatecheck/**` (Option C). `.github/CODEOWNERS`
   is advisory today and does not cover `tools/gatecheck/**`; making it required is an **operator**
   branch-protection action. Stage does not take it and does not author a CODEOWNERS change here
   (config/source mutation is outside the Stage role boundary).
3. No code task is harvested. The stash entry stays active, annotated with this artifact, so the
   operator decision is not lost.

## Open question for the operator (blocking harvest of any follow-up)

Do you want required CODEOWNERS review on `tools/gatecheck/**` with branch protection? If yes,
re-stash a chore for the CODEOWNERS change and record the branch-protection action separately. If
no, archive 3750C37C as an accepted residual and keep D-3750-1 as the rationale.

## Status

`decided-pending-operator-confirmation`: the Orchestrator asked for a decision artifact only; no
plan or harvest follows from this entry.
