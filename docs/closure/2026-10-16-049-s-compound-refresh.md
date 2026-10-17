---
title: "Compound refresh: 049-S / 059-F"
description: "Evidence-based review of compound entries touched by the closure contract design-doc and test accuracy shipment."
date: 2026-10-16
mode: apply
shipment: 049-S
feature: 059-F
pr: 147
---

# Compound refresh: 049-S / 059-F

## Scope and evidence

Searched `docs/compound/` for `cascade-close`, `--message`, `--author`, `selected_close_path`, exit-code
routing and the closure matrix. Evidence: the PR #147 merge (`009f100`) and follow-up PR #148 (`f6e97e2`,
`51f0751`, test-only fingerprint fix), the closure matrix design doc
`docs/design-docs/2026-10-11-closure-contract-surface-matrix.md`, and the 049-S ship session memory.
Engram and Graphtor were not consulted; a direct targeted search was used.

## Classifications

| Entry | Classification | Evidence |
|---|---|---|
| `2026-05-07-backlogit-shipment-status-constraints.md` | **keep** | Its `backlogit shipment ship --sha <sha> --message <msg> --author <author>` examples already use the space-separated form that the shipped work confirms as correct. |
| `2026-10-08-bound-cascade-close-procedure.md` | **keep** | Already marked superseded in part (043-S / 053-F); its space-separated form is still correct, and the historical text is retained for the record. |
| `2026-10-11-classify-only-writer-evidence-before-cascade-close.md` | **keep** | Uses the space-separated form. Nothing in 049-S contradicts it. |
| `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` | **keep** | Matched on `cascade-close`; the bypass-risk lesson is unaffected. |
| Other entries | **keep** | No overlap with this shipment. |

No entry shows the rejected `--message=` or `--author=` form. No entry described the smoke test's exit
derivation or the contract-test row that guards the equals form, so none was invalidated.

## Changes and follow-up

* No compound file was updated, consolidated, replaced, archived, deleted or marked stale.
* Follow-up (optional): a new compound entry could record that cascade-close rejects the `=` argument
  form and that the contract test guards it. The design doc and test already carry that rule.
* No stash or backlog item was changed.
