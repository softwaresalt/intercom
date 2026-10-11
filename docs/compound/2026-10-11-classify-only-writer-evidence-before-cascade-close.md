# Capture pre-close writer evidence before a CASCADE close

Date: 2026-10-11. Source: shipment 042-S closure (feature 052-F, PR #123). Related: the 041-S closure residual `cascade-evidence-record-missing` (docs/closure/041-S-051-F-post-merge-closure.md).

## Problem

`autoharness shipment cascade-close` refuses to write a close-evidence record for a shipment that is no longer open. The 041-S closure ran the backlogit cascade directly and so had no writer-form record, which needed an operator waiver.

## What worked (042-S)

1. Merge and gates first: merge confirmation, `pipeline-topology --phase lifecycle`, and the covering-feature completion gate (one `active -> done` move, no force flags).
2. Writer-form pre-close evidence while the shipment was still open: `autoharness shipment cascade-close --classify-only --shipment <S> --feature <F> --sha <merge_sha> --workspace . --json`. It exited 0 with `classifier_verdict: CASCADE`, `engine_verdict: VERIFIED`, `mutation_possible: no`, and wrote a `pre_close` record under `docs/closure/evidence/`.
3. The shipment-reconcile binding was computed from the canonical v1 lines and recorded in the reconcile record, so a reviewer can recompute it.
4. The close ran as `backlogit shipment ship <S> --sha <merge_sha> --message ... --author ...`. Result checks: `returned_ids` empty, and the two-set `allowed_ids` and `required_ids` gate passed on `archived_ids`.

## What not to do

* Do not use the mutating `autoharness shipment cascade-close` from Ship. It invokes `backlogit shipment ship` itself, which bypasses the classified-close boundary of the shipment-reconcile skill. Use `--classify-only` for the evidence and keep the close on the skill path.
* Do not run `--classify-only` after the close. A post-close invocation is refused, so the `post_close` writer phase cannot be produced by the harness without the mutating form. Record that as a condition rather than reconstructing it.

## Open items (tracked)

* No invokable safe-close CLI or gate exists yet (stash D10D3AFC). The binding is agent-computed and recorded, not tool-enforced.
* The `post_close` writer phase needs a verified step that does not invoke the ship primitive from Ship.
