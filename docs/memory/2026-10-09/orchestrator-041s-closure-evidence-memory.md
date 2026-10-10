---
title: "Orchestrator memory: 041-S cascade-evidence record and closure-fix routing"
description: "Why the autoharness cascade-evidence writer cannot produce a record for the closed 041-S shipment, the waiver proposed in its place, the Principle IV handling of the %TEMP% scratch files, and the routing-config carry-forward."
date: 2026-10-09
agent: orchestrator
shipment: 041-S
feature: 051-F
---

## Operator decisions received

* Produce `docs/closure/evidence/041-S-051-F-close-evidence.json` with the autoharness cascade-evidence writer.
* Remove the scratch files left under `%TEMP%`.
* Carry the `.autoharness/config.yaml` edit into the commit submitted for the closure-evidence PR.

## Findings

### The writer cannot produce a record for 041-S

`autoharness shipment cascade-close --classify-only --shipment 041-S --feature 051-F --sha fcc61f3ceea06a96d268070008b45ab5536ba1fb --workspace . --json` exited 2 with `evidence_path: null` and the failure `shipment '041-S' is no longer open (location archive, status 'archived'); no verdict record after the fact`. The attempt wrote no record. It released its pair lock and left two empty directories, `docs/closure/evidence/` and `.autoharness/gates/cascade-close/`.

The refusal is by design, not a defect:

* `run_preclose` requires the shipment to be in the queue with status `queued` or `active`, so both writer modes refuse a closed shipment.
* A valid `cascade` record is phase `post_close`. It carries a pre-close snapshot with classifier `CASCADE` and engine `VERIFIED`, a captured `backlogit shipment ship` invocation (argv, timestamps, exit 0), and the parsed result. None of that exists for 041-S, because the agent ran the cascade directly on the P-015 path.
* The validator re-derives the verdict from the recorded raw inputs. A hand-built record would be fabricated evidence, so none was written.

### The evidence requirement postdates the installed harness

* The workspace manifest records `autoharness_version: 1.5.0`. No workspace agent, skill, or workflow invokes `autoharness gate closure-evidence` or `autoharness shipment cascade-close`.
* The prior closure `040-S-050-F` fails the same gate at `close_path`, so no closure in this repository carries an evidence record.
* Routing reads only the frontmatter predicate (`topology._closure_artifact_complete`). It needs `compaction_status` done or degraded, and every condition `satisfied: true` with non-empty evidence. The evidence file is not part of routing. The ambient topology gate passes on `main`.

## Decisions

* Proposed the alternative the operator was offered earlier: an operator waiver of condition `cascade-evidence-record-missing`, carried by a closure-fix PR. The PR is prepared to merge-ready only. P-014 requires the operator's explicit merge approval, and that approval is the recorded waiver.
* Carried `.autoharness/config.yaml` verbatim into the same commit, as instructed. `autoharness verify-workspace` reports 0 blockers and 0 strict schema blockers. It reports 32 `ROUTE_VARIABLE_STALE` warnings, because the installed agent routing variables were rendered before the routing change. CI does not run `verify-workspace`.
* Declined to delete the `%TEMP%` scratch files. Principle IV (NON-NEGOTIABLE) bars agents from deleting outside the workspace, and the 040-S and 041-S closures followed the same rule. Provided `logs/remove-temp-scratch.ps1` (dry run by default, exact names only) for the operator to run.

## Scratch files under %TEMP% from the 040-S and 041-S runs

* 040-S: `gatecheck-040s.exe`, `binding040s.ps1`, and `040s-preclose-backup\` (639 files)
* 041-S and H0: `ship-h0-carry\`, `stash-before-h0-captures.jsonl`, and the `h0-*` and `h041-*` message and log files named in the script
* Probable: `stage_stash.txt` (Stage dump, 2026-10-08 15:19)

## Next steps

* Operator: review and approve the closure-fix PR (the merge approval records the waiver), and run `logs/remove-temp-scratch.ps1`.
* Stage: triage the follow-up stash entry for adopting the 192-F closure flow and re-rendering stale routing variables with `tune-harness`.
* After merge: confirm `autoharness gate closure-evidence --path docs/closure/041-S-051-F-post-merge-closure.md --shipment 041-S` advances past `frontmatter_predicate`.
