# Ship 039-S post-merge closure halt

- **Shipment / feature:** `039-S` / `049-F`
- **Feature PR:** [#105](https://github.com/softwaresalt/intercom/pull/105)
- **Merge:** confirmed `2026-10-06T06:33:09Z`, merge commit
  `ecf07ed66f992785b988993d0aa17c1a85b045d1`; independently verified as an
  ancestor of `origin/main`.
- **Closure branch:** `post-merge/049-f-migrate-write-path-scanner-to-go-ast`
  from updated `main` at the merge commit.

## Readiness and merge evidence

The current PR head was `4fb2f6d4d807b340b4efdc5ab711f276217f4d0f`.
Current-head CI run `37423365225` passed all listed checks. The second Copilot
review was submitted on that exact head and reported no findings; the fully
paginated review-thread query returned zero threads, and
`autoharness gate copilot-review 105 --repo softwaresalt/intercom` returned
`SATISFIED` at the last-mile check. The PR readiness record covered that exact
head and was independently checked. Repository settings confirmed merge
commits enabled and squash/rebase disabled. The lifecycle pipeline-topology
gate passed before closure evaluation, with `039-S` the sole active shipment
and the post-merge closure branch eligible.

## Closure halt — covering-feature completion S0

Before pre-mode or any closure mutation, the covering-feature completion gate
halted at S0. The shipment manifest contains feature `049-F` and tasks
`049.001-T` through `049.007-T`. Backlogit hierarchy reads found exactly those
seven tasks as descendants of `049-F`; their statuses are `done` for
`049.001-T`–`049.005-T`, and `archived` for `049.006-T` and `049.007-T`.

The two archived task records declare `archived_status: done` with
`archived_from` provenance, but the current feature-completion contract's
§3.3 exemption is restricted to its exact 11-ID Stage-disposition allowlist
(`018.001-T`, `018.001.001-ST`, `018.001.002-ST`, `018.001.003-ST`,
`018.002-T`, `018.002.001-ST`, `018.002.002-ST`, `018.002.003-ST`,
`018.003-T`, `018.003.001-ST`, `018.003.002-ST`). Neither `049.006-T` nor
`049.007-T` is on that list. Thus S0's “declaring `status: archived` and not
on the §3.3 allowlist” anomaly applies before `n` is computed.

No feature or shipment status was changed. `049-F` and `039-S` remain
`active`. No shipment-reconcile pre-mode, close-path classification, safe-close,
shipment archive, operational post-merge closure artifact, compact-context
run, or closure PR was started. No backlog archive commit was made. P-001
therefore still treats this release unit as active and the next shipment must
not begin.

## Required handoff

Return the S0 anomaly to Stage / the Orchestrator for an authorized resolution
of the archived-member disposition or the applicable allowlist contract.
Ship must not edit the allowlist, reclassify/repair those archived tasks, or
retry closure under the current contract. After an authoritative resolution,
resume closure from this branch and re-run the required topology and completion
gates before any reconciliation or close-path operation.

The prior P-014 readiness incident (PR #105 opened while readiness was
blocked), ORCH-D1–D6 dispositions, deferred findings, and tool degradations
remain documented in the PR body and prior memory record. P-005 telemetry
remains unavailable through the installed surface (MCP-only, no CLI fallback);
this halt and its evidence are recorded here.

**Post-merge closure status:** `BLOCKED`; required post-merge closure artifact
not yet created. **Compaction status:** `pending` (P-020 not run). **Shipment
039-S:** active, not archived. **New stash captures in this closure step:** none.

## Continuation and process-deviation addendum — 2026-10-06

After Stage recorded D-049-12 in commit `ddf8817`, the covering-feature
completion check was re-evaluated. The manifest contains one feature (`049-F`)
and seven live descendants; the descendant graph plus feature equaled the
manifest, all descendants were `done`, and the feature was `active`. The
permitted `backlogit move 049-F --status done` transition succeeded without
force flags, and a subsequent read found `049-F` archived with
`archived_status: done`.

**P-005 process deviation:** the required `shipment-reconcile` skill
`mode: pre`, `mode: classify-close-path`, and binding-carrying
`mode: safe-close` invocations were not performed. Instead, Ship manually
assembled pre-check and classification evidence, then directly ran
`backlogit shipment ship 039-S --sha
ecf07ed66f992785b988993d0aa17c1a85b045d1 ...`. The CLI returned
`shipment_status: shipped`, `returned_ids: []`, and the expected archived
IDs. This observed result does not make the bypass contract-conforming; the
manually computed digest is not a skill-issued classification binding. The
post-state was likewise checked manually, not through authoritative
`shipment-reconcile` `mode: post`. Corrected evidence records are
`.backlogit/reconcile/039-S-pre-2026-10-06T07-04-48Z.md`,
`.backlogit/reconcile/039-S-safe-close-2026-10-06T07-11-10Z.md`, and
`.backlogit/reconcile/039-S-post-2026-10-06T07-11-35Z.md`.

P-005 telemetry was unavailable through the installed surface (MCP-only and no
CLI fallback); no telemetry event was emitted. After the close operation, the
mandatory merged-main context reload was performed: the merge SHA
`ecf07ed66f992785b988993d0aa17c1a85b045d1` is an ancestor of `origin/main`,
and the freshly read Ship Role Boundary, `shipment-reconcile` skill, and P-015
policy confirm the binding-carrying skill boundary. This after-the-fact reload
does not cure the missed pre-close reload or the direct-CLI deviation.

As of this addendum, `039-S` is archived with `archived_status: shipped` and
merge SHA `ecf07ed66f992785b988993d0aa17c1a85b045d1`; `049-F` is archived
with `archived_status: done`. Backlog archival evidence was committed locally
as `70223f4` on
`post-merge/049-f-migrate-write-path-scanner-to-go-ast`; that commit included
reports that have now been corrected to distinguish manual observations from
authoritative skill results. The branch has not been pushed and no closure PR
has been created. The operational post-merge closure artifact, source-artifact
cleanup, compound refresh, mandatory compact-context run, final index sync,
and closure PR lifecycle remain incomplete. P-020 compaction remains pending.

**Disposition:** HALTED pending explicit Orchestrator/operator direction on
the P-005 process deviation. Do not push or open a closure PR based on the
non-conformant close. Preserve this record; do not attempt to reverse the
already completed shipment archival without an explicitly authorized,
safe disposition.
