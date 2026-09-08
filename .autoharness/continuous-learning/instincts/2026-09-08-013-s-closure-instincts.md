---
generated: 2026-09-08
scope: recent
source_observations:
  - .autoharness/continuous-learning/observations/2026-09-05.jsonl
  - .autoharness/continuous-learning/observations/2026-09-08.jsonl
release_context: 013-S / 014-F (pathsafe reparse-point containment hardening)
---

# Instincts — 013-S closure session

Clustering the 1 new observation captured during 013-S/014-F post-merge
closure against the prior 3 observations from the 008-S session (recent
scope = both observation files present in the store; the 013-S session
itself produced only 1 new observation record, but that record's own
evidence field already documents 3 distinct shipment occurrences of the
same pattern).

## Instinct 1: closure artifacts missing frontmatter reliably draw a Copilot finding (PROMOTED)

* **Pattern**: Post-merge closure docs authored by Ship without the
  established YAML frontmatter block (`title`/`date`/`mode`/`shipment`/
  `feature`/`pr`/`merge_commit_sha`/`compaction_status`/`closure_status`/
  `releasability`) reliably draw a Copilot review finding on the closure
  PR, costing a fix-reply-resolve cycle each time.
* **Evidence count**: 3 corroborating occurrences across independent
  shipments — `chore/008-s-closure-evidence-frontmatter-fix` branch,
  `post-merge/012-s-retire-d6a-gate-narrowing` commit `f2afdd9` ("fix
  closure frontmatter schema... per Copilot review"), and this session's
  PR #41 review thread `PRRT_kwDOTPuhps6gcjSY` on the 013-S closure PR.
* **Confidence**: high — reaches the configured promotion threshold
  (`continuous_learning.promotion_threshold: 3`); same defect class,
  same detection mechanism (Copilot review), same remediation
  (retrofit the frontmatter block from an existing closure doc), across
  3 independent release units.
* **Workflow phase**: closure (post-merge, PR authoring)
* **Suggested next action**: **evolve into instruction** — add an explicit
  step to the Ship agent template's Step 6 ("Post-Merge Closure") and/or
  the `operational-closure` skill instructing Ship to copy the YAML
  frontmatter block from the most recent prior `docs/closure/*.md` file
  as a template *before* first drafting a new closure artifact, rather
  than relying on Copilot review to catch the omission after the fact
  every time.

## Instinct 2: P-015 cascade classifier requires manual Python invocation, no CLI subcommand (still below threshold)

* **Pattern**: `classify_shipment_close_path` has no `autoharness gate`
  CLI subcommand exposing it; Ship must hand-invoke the Python module
  directly (via `PYTHONPATH` in the 008-S session, via a `pip`-installed
  `autoharness` package import in this 013-S session) every time a
  cascade-eligible shipment closes.
* **Evidence count**: 2 (008-S, 013-S)
* **Confidence**: low-medium (recurring, but one below promotion
  threshold)
* **Workflow phase**: closure
* **Suggested next action**: keep observing. If a third occurrence is
  logged, promote to a proposal for an `autoharness gate
  shipment-closure` CLI subcommand mirroring `pipeline-topology` and
  `copilot-review`.

## Promotion status

Instinct 1 reaches the promotion threshold (3 corroborating occurrences)
and is promoted for `evolve` in `mode: propose`. Instinct 2 and the
carried-forward 008-S instincts (backlogit auto-relocation on `done`,
batch-then-commit commit granularity) remain below threshold — kept as
low-confidence instincts to watch for repetition.
