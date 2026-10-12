---
name: shipment-reconcile
description: "GI/GR reconciliation gate for shipment manifests — verifies every manifest item exists in queue (pre-mode) or archive (post-mode) with the expected status, and closes shipments with the single-artifact safe-close procedure that archives ONLY manifest item IDs one at a time and closes the shipment record via live `shipped` -> verify -> explicit archive -> verify `archived_status: shipped`, instead of the destructive cascade backlogit_ship_shipment. The destructive CASCADE close runs only through `autoharness shipment cascade-close` after per-closure operator approval, and only for the narrow, machine-verified P-015 engine-inertness case."
---

# Shipment Reconcile

Provides a double-entry (GI/GR) integrity check for shipment manifests. Run
`mode: pre` before closing a shipment and `mode: post` after the archive +
restore steps complete. Run `mode: safe-close` **in place of** the destructive
cascade `backlogit_ship_shipment` call to archive only the shipment manifest's
explicit item IDs one artifact at a time, verifying after each that the parent
feature and any unshipped sibling tasks survive. The `--classify-only` route in
Step 0(c) Command Routing (below) performs and records the P-015 engine-inertness
classification first; safe-close is entered only when that route exits 3
(`safe_close` recorded), and Ship never re-derives the close path. A CASCADE
result (exit 0) runs only the mutating `autoharness shipment cascade-close`
command, after per-closure operator approval.

> **Why safe-close exists.** `backlogit_ship_shipment` treats a shipment as a
> proxy for its covering feature and cascade-archives the whole feature subtree.
> For **partial-feature** shipments (which intentionally exclude the parent
> feature and some sibling tasks) this is destructive: it archives the parent
> feature and orphans unshipped siblings that are **not in the shipment
> manifest**. Safe-close prevents that corruption instead of merely detecting it
> after the fact. See **P-015** in `workflow-policies` for the governing policy.

## When to Use

* **Ship Step 6 closure** (mandatory): run `mode: pre`, then `mode: safe-close`
  **instead of** the cascade `backlogit_ship_shipment` call, then `mode: post`.
  The `--classify-only` route (Step 0(c) Command Routing, below) selects and
  records the close path from the machine-checkable P-015 classification.
  Safe-close is entered only on exit 3 (`safe_close` recorded), archives
  manifest item IDs individually, and never calls the cascade op. Exit 0
  (`CASCADE`) routes to the Cascade Close Sub-Procedure, which runs only
  through the mutating run of the cascade-close command (`autoharness shipment cascade-close`)
  after per-closure operator approval. Ship never re-derives the close path.
* **Ship Step 0.5** (sanity check): pre-mode at intake with `expected_status: queued`
  (or `active` if the shipment was already claimed in a prior session)
  to catch Stage-side over-inclusion before any build work begins.
* **Ad-hoc audit**: any time an operator suspects manifest drift.
* **`mode: detect-mixed-role`** (operator-invoked, READ-ONLY, no lock, no mutation):
  any time an operator wants a diagnostic scan for the queued-with-active-work /
  mixed-role silently-dropped-claim signature (936C68F3 part 2, re-scoped
  report-only per 013-DL Addendum G / 112-F) across one shipment or ALL
  shipments. Composed entirely from EXISTING read-only backlogit reads
  (`backlogit_list_shipments` + per-shipment/per-task status reads via
  `backlogit_get_shipment` and per-task `backlogit_get_item`). EMITS a report-only
  diagnostic plus operator-remediation guidance. NEVER mutates, NEVER calls
  `backlogit_claim_shipment` or any status-write operation, and needs no
  `file-lock` acquisition because no backlog/shipment artifact is ever
  mutated — the mode's own diagnostic report, audit-log entry, and telemetry
  event (steps 6–8 below) are additive-only writes to non-backlog-state
  locations, never applied to a queue/archive item.

## Inputs

| Parameter | Required | Values | Notes |
|---|---|---|---|
| `mode` | yes | `pre` \| `post` \| `safe-close` \| `detect-mixed-role` | Controls which check/close/detect phase runs |
| `shipment_id` | yes for `pre`/`post`/`safe-close`; optional for `detect-mixed-role` | e.g. `004-S` | The shipment to reconcile; for `detect-mixed-role`, omit to scan ALL shipments via `backlogit_list_shipments` |
| `expected_status` | pre-mode only | `queued` \| `active` \| `done` | `queued` for fresh intake; `active` when shipment already claimed in a prior session; `done` for pre-ship check |
| `merge_commit_sha` | post-mode and safe-close | git SHA | The merge commit that closed the PR; recorded on archived items for traceability |

## Output

A structured **reconciliation report** stored at
`.backlogit/reconcile/{shipment_id}-{mode}-{timestamp}.md`.

Every item in the manifest is classified as one of:

| Classification | Pre-Mode Meaning | Post-Mode Meaning |
|---|---|---|
| `matched` | Queue file present AND declared status matches `expected_status` | Archive file present for this item |
| `pre-archived` | No queue file found but archive file exists — item already archived before this shipment ran; treated as valid | N/A (all items are expected in archive; use `matched` / `missing`) |
| `missing` | No queue or archive file found for this manifest item | Archive file not found for this manifest item |
| `status-mismatch` | Queue file present but declared status does not match `expected_status` | N/A (post-mode does not check status fields) |
| `orphan` | Queue file declares this `shipment_id` in its frontmatter but is NOT in the manifest | N/A (post-mode does not scan queue files) |

> Classification semantics are mode-dependent. Pre-mode checks the queue
> for status correctness; post-mode checks the archive for file presence only.

### Shipment-Record-Status Classification (record scope, distinct from the five per-item classifications above)

In addition to the five per-item classifications, pre-mode also classifies the
shipment **record's own** `status` against the aggregate status of its manifest
**task** items. **Task-artifact filter (mandatory)**: the manifest `items` list
is untyped and may include the covering feature id (e.g. a fallback-assembled
manifest); this classification aggregates only items whose `artifact_type` is
`task` (already read from each item's frontmatter during the per-item check
above — no new read), excluding any non-task entry, so a covering feature that
happens to be `active`/`done` outside the shipment's own
scope can never be misread as a "conflicting task". This mirrors the task-artifact
filter the Ship agent's intake early-warning already applies to
`custom_fields.items` (`templates/agents/_ship.agent.md.tmpl`). **Scope**: the
three named inconsistency cases below apply only when the
record's own status is `queued` or `blocked` — the two
statuses where a manifest task already being `active`/`done`
is itself the drift signal (the queued/blocked record has not "caught up" to its
tasks). **`blocked` itself is a non-standard/legacy value**: backlogit
1.8.0's `ShipmentStatus` enum is only `queued|active|shipped|abandoned` (see
`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`) — a
persisted `blocked` record can still exist in real workspaces from a
historical `backlogit move` CLI defect that silently accepted invalid status
writes. Pre-mode's `record-blocked-with-active-work`/`record-blocked-with-done-work`
cases below classify such a leftover record defensively (it is real data that may
be present); `mode: detect-mixed-role`'s `malformed-legacy` classification
(defined further below, in the mode's own classification section) reports the
identical underlying fact — a non-standard/legacy status value, never a normal
current-day state — for a different manifest scan. Both agree `blocked` is never fabricated or
transitioned into/out of; only the case label differs by mode. When the record's
own status is `active` or `done`
(or archived), the record is **by definition** `record-consistent` for this
check — an `active` record is the normal in-progress state while its
tasks move `queued` → `active` → `done`, and an
`done`/archived record reflects a shipment already closed. This is an
explicit scope boundary, not a silent default: the four cases below are
**mutually exclusive** because they are evaluated in this fixed order and every
record status value maps to exactly one of them.

| Classification | Condition |
|---|---|
| `record-consistent` | Record status is `active` or `done`/archived (always consistent — the normal in-progress/closed lifecycle, out of scope for this check), OR record status is `queued`/`blocked` and none of the three inconsistency conditions below match (e.g. record `queued` with all task items `queued`) |
| `record-queued-with-active-work` | Record status is `queued` AND at least one **task-artifact** manifest item is `active` or `done` — the classic "silently-dropped claim" inconsistency |
| `record-blocked-with-active-work` | Record status is `blocked` AND at least one **task-artifact** manifest item is `active`. **Precedence**: when a `blocked` record has BOTH an `active` task and a `done` task, classify here — active work takes precedence over `record-blocked-with-done-work` below, because it is the more severe/earlier-stage drift signal |
| `record-blocked-with-done-work` | Record status is `blocked` AND no **task-artifact** manifest item is `active` AND at least one **task-artifact** manifest item is `done` |

These four cases are mutually exclusive: every possible record-status value
(`queued`, `active`, `blocked`, `done`)
is covered — `active`/`done` always resolve to
`record-consistent`, and for `queued`/`blocked` the
active-over-done precedence rule (applied only to task-artifact items) guarantees
exactly one of the remaining three cases applies. This check is
**detect-and-report only — NO auto-repair**: it
never mutates the shipment record or any task; operators must manually
reconcile.

The report ends with a `recommendation`:

* `PROCEED` — all items are `matched` or `pre-archived` AND the
  shipment-record-status classification is `record-consistent`; no action needed
* `HALT — operator reconcile required` — one or more missing, status-mismatch, or orphan items, OR a non-`record-consistent` shipment-record-status classification (pre-mode)
* `HALT — restore archives` — missing archive files or unrestored deletions (post-mode)
* `CLOSED` — safe-close archived every manifest item individually, archived the
  shipment record itself, and the protected set (parent feature + unshipped siblings)
  is intact
* `HALT — cascade detected, revert required` — safe-close found a non-manifest artifact (parent feature or a sibling task) archived or deleted; the unintended change must be reverted before any commit

For `mode: safe-close`, the report also records the **protected set** (the parent
feature file and every unshipped sibling task file that must survive closure) and,
per manifest item, whether it was `matched` (archived by this run) or
`pre-archived` (already archived before this run; skipped to avoid
double-archival and false-positive cascade flags).

### Mixed-Role Detection Classification (`mode: detect-mixed-role` only, distinct from the record-scope classification above)

This classification is **separate** from the "Shipment-Record-Status
Classification" above: that check compares a single shipment record's own
status against the *aggregate* active/done state of its manifest tasks (four
record-scope cases). This check instead classifies **each manifest task
individually** by its per-task **ROLE**, to precisely describe the
mixed-role "silently-dropped-claim" signature — a `queued` shipment
record whose manifest tasks have kept progressing to `active` and
even fully `done`/archived while the record itself never advanced.
Per **013-DL Addendum G** (re-scoped by Copilot PR #304 finding 1), this
classification is used **ONLY to DESCRIBE** the inconsistency in a report —
**NEVER** to gate a mutation. There is no `--confirm` flag because nothing is
ever mutated.

**Per-task ALLOWED ROLE** (a task-artifact manifest item must be a UNIQUE,
NON-CONFLICTING record for exactly one of these three roles):

| Role | Definition |
|---|---|
| `live-queued` | UNIQUE record in `.backlogit/queue/` with `status: queued`; NO archive record for the same id |
| `live-active` | UNIQUE record in `.backlogit/queue/` with `status: active`; NO archive record for the same id |
| `archived-completed(done)` | UNIQUE record in `.backlogit/archive/` ONLY, in EITHER valid representation: (a) TERMINAL RELOCATION — `status: done` (provenance `archived_status`/`archived_from` NOT required), OR (b) EXPLICIT ARCHIVAL — `status: archived` AND `archived_status: done` AND valid, well-formed `archived_from` provenance; NO conflicting live queue record for the same id |

**Per-item ANOMALY** (fail closed — REPORT and HALT on ANY of these; a
manifest task satisfying none of these is role-clean):

| Anomaly | Definition |
|---|---|
| `duplicate` | Same task id present in BOTH `.backlogit/queue/` AND `.backlogit/archive/` |
| `conflicting` | Queue status disagrees with the declared role, an archive record exists alongside a live queue record, or a live `status: done` record is found in the QUEUE (a completed task lives ONLY in `archive/`, never live-done in `queue/`) |
| `missing` | Manifest task id has no file in either `.backlogit/queue/` or `.backlogit/archive/` |
| `malformed-provenance` | An archive record with `status: archived` is missing or has ill-formed `archived_status`/`archived_from` (a `status: done` archive record legitimately carries no provenance and is NOT malformed) |
| `any-other-archived-status` | An archive record whose status is NEITHER `done` NOR `archived`-with-`archived_status: done` |
| `orphan` | A queue file declares this `shipment_id` in its frontmatter but is NOT present in the manifest `items` list (reuses the pre-mode orphan-scan definition) |
| `out-of-role` | Task status falls outside the allowed lifecycle set `queued` \| `active` \| `done`/`archived` (e.g. a non-lifecycle or otherwise malformed status value) |
| `torn-partial` | Any other ambiguous, incomplete, or inconsistent signal for the task that cannot be cleanly assigned to a role (e.g. partially-written frontmatter) |

**Malformed-legacy shipment record**: backlogit 1.8.0 has NO `blocked`
shipment status (`ShipmentStatus` is only `queued|active|shipped|abandoned`).
A shipment record whose persisted status is anything other than a valid 1.8.0
lifecycle value (e.g. a legacy `blocked` value) is described in the report as
`malformed-legacy` — REPORT it, HALT, and never fabricate a `blocked->queued`
or any other transition. This is the same underlying fact the pre-mode
Shipment-Record-Status Classification table above documents for its own
scan (`record-blocked-with-active-work`/`record-blocked-with-done-work`
defensively classify a leftover legacy `blocked` record because it
may exist in real workspaces); the two modes describe the identical
non-standard value under mode-appropriate labels — never treat a persisted
`blocked` value as a normal current-day state in either mode.

**Mixed-role signature**: a shipment record `queued` whose
task-artifact manifest items (filtered by `artifact_type` to exclude any
non-task entry, e.g. a covering feature id, mirroring the existing
task-artifact filter above) include AT LEAST ONE `live-active` or
`archived-completed(done)` role task, with every task otherwise role-clean
(no anomaly), is the reportable "silently-dropped-claim" signature. On ANY
per-item anomaly (duplicate / conflicting / missing / malformed-provenance /
any-other-archived-status / orphan / out-of-role / torn-partial) or any other
ambiguity, REPORT the specific anomaly and HALT — never mutate, never
attempt to resolve or repair.

**Detection outcomes** (structured audit entry + telemetry event on every
run — see "Mixed-Role Detection Audit + Telemetry" below): exactly one of
`DETECTED` (scan completed; no mixed-role signature or anomaly found —
`record-consistent`, nothing to report), `REPORTED` (scan completed; the
mixed-role signature and/or one or more per-item anomalies were found and
described in the report), or `DEGRADED` (backlogit was unreachable; the
degraded condition is reported and the scan halts). There is **NO**
`succeeded` / `repaired` / `refused` / two-active outcome — nothing is ever
mutated or repaired by this mode.

## Behavioral Constraints

* **Report-and-halt only.** This skill NEVER modifies the shipment manifest or
  queue/archive files **outside the safe-close mode's manifest-scoped archival**.
  In pre- and post-mode it only reports; operators must manually reconcile via
  existing backlog tools and re-invoke Ship Step 6.
* **Manifest-scoped mutation only.** In `mode: safe-close`, the ONLY artifacts
  this skill may move or archive are the shipment manifest's explicit item IDs and
  the shipment record itself (`{shipment_id}`). It must NEVER archive the parent
  feature or any sibling task that is not in the manifest. It never calls the cascade
  `backlogit_ship_shipment`; the shipment record is closed as its own single artifact.
* **No prune / no auto-repair.** Auto-mutation of the manifest itself is reserved
  for a future version. Safe-close never prunes the manifest and never
  auto-deletes non-manifest artifacts; on cascade detection it reverts the
  unintended change and halts.
* **Single-writer lock.** When invoked from Ship Step 6, this skill holds the
  `.backlogit/queue/{shipment_id}.md` file lock (via the `file-lock` skill) for
  the duration of pre-mode → safe-close → post-mode. See lock protocol in the
  Required Protocol section below.
* **Halt on RECONCILE_FAIL.** Do not proceed to safe-close unless pre-mode
  returns `PROCEED`. Do not commit backlog state if safe-close returns
  `HALT — cascade detected, revert required`. Surface the report path to the operator.
* **`mode: detect-mixed-role` is strictly READ-ONLY.** It NEVER mutates any
  shipment record or task, NEVER calls `backlogit_claim_shipment` (no
  re-claim, no repair mode — a record-only forward re-claim of a
  queued-with-active-work shipment is UNSUPPORTED by backlogit 1.8.0; see
  "Operator-Remediation Guidance" below), and requires NO `file-lock`
  acquisition because no backlog/shipment artifact is ever mutated (its own
  diagnostic report, audit-log entry, and telemetry event are additive-only
  writes to non-backlog-state locations). DEGRADED (backlogit unreachable)
  is REPORTED and the mode HALTS — it never guesses or acts blind.

## Required Protocol

### Pre-Mode

1. **Acquire single-writer lock** (Ship Step 6 invocations only, not intake):
   Invoke the `file-lock` skill to acquire `.backlogit/queue/{shipment_id}.md`.
   If lock acquisition fails, count as a session stall (circuit-breaker protocol)
   and prompt the operator.

2. **Load manifest** via `backlogit_get_shipment(shipment_id)`.
   Extract the `items` list.

   Classification is read-only, so Pre-Mode stays detect-and-report only.

3. **Check each manifest item**:
   * Attempt to locate the file at `.backlogit/queue/{id}.*`
   * If found, read its frontmatter (including `status` and `artifact_type`) and
     compare `status` to `expected_status` — classify as `matched` or
     `status-mismatch`
   * If NOT found in queue, check `.backlogit/archive/{id}.*`
     — if archive file exists, classify as `pre-archived` (valid; item already shipped)
     — if no file in either location, classify as `missing`

4. **Orphan scan**:
   Scan `.backlogit/queue/` for any files whose YAML frontmatter declares
   `shipment_id: {shipment_id}` but whose ID is NOT present in the manifest `items` list.
   Classify each such file as `orphan`.

5. **Shipment-record-status classification** (reuses in-hand data — NO new scan):
   Using the shipment record's own `status` already loaded via `backlogit_get_shipment`
   in step 2, and the manifest items' statuses already read in step 3, classify the
   record scope per the Shipment-Record-Status Classification table in the Output
   section above, evaluated in this order. **Filter to task artifacts first**: the
   manifest `items` list is untyped and may include the covering feature id (e.g. a
   fallback-assembled manifest); reuse the `artifact_type` already read from each
   item's frontmatter in step 3 to exclude any non-task entry before aggregating
   task statuses below — the same task-artifact filter the Ship agent's intake
   early-warning applies to `custom_fields.items` (`templates/agents/_ship.agent.md.tmpl`)
   — so a covering feature that is `active`/`done` outside the
   shipment's own manifest scope can never be misread as a "conflicting task" and
   falsely halt an otherwise-consistent shipment.
   * Record `active` or `done`/archived →
     `record-consistent` (always — this check is scoped to `queued`/
     `blocked` records only; an active/done record is the normal
     in-progress/closed lifecycle state, not evaluated further).
   * Record `queued` AND any manifest **task** is `active` or
     `done` → `record-queued-with-active-work`.
   * Record `blocked` AND any manifest **task** is `active` →
     `record-blocked-with-active-work` (takes precedence over the case below when
     both an active and a done task are present).
   * Record `blocked` AND no task `active` AND any manifest
     **task** is `done` → `record-blocked-with-done-work`.
   * Record `queued` or `blocked` matching none of the above
     → `record-consistent` (e.g. record `queued` with all tasks
     `queued`).
   This step is **detect-and-report only — NO auto-repair**: it never mutates the
   shipment record or any task.

6. **Produce report** and store at
   `.backlogit/reconcile/{shipment_id}-{mode}-{timestamp}.md`.

7. **Gate decision**:
   * If all items are `matched` or `pre-archived`, no orphans exist, AND the
     shipment-record-status classification is `record-consistent` →
     `recommendation: PROCEED`
   * If any `missing`, `status-mismatch`, or `orphan` items exist, OR the
     shipment-record-status classification is `record-queued-with-active-work`,
     `record-blocked-with-active-work`, or `record-blocked-with-done-work` →
     `recommendation: HALT — operator reconcile required`, naming the shipment id,
     the record's own status, and the conflicting manifest task ids
   * On `HALT`: emit the report path, release the lock, and halt with
     `RECONCILE_FAIL`. Do NOT call `backlogit_ship_shipment`.
   * On `PROCEED` from Ship Step 6: retain the lock until post-mode completes.

### Post-Mode

1. **Verify archive presence**:
   List `.backlogit/archive/` and confirm a file exists for the shipment itself
   (`{shipment_id}.*`).

2. **Per-item archive check**:
   For every item in the manifest, verify a corresponding archive file exists.
   If any are absent, flag them in the report.

3. **Deleted-file guard** (known `backlogit_ship_shipment` quirk — see P-007):
   Run `git status --short -- ".backlogit/archive/"` and inspect for deletions.
   If any archive files are reported as deleted (` D`), restore only the
   individual paths that `git status --short -- .backlogit/archive/` lists as
   deleted, one `git restore -- <path>` per listed path (never a whole
   directory), before the commit step.

4. **Produce post-mode report** per the same schema.

5. **Gate decision**:
   * If all archive files present and no deletions detected → `recommendation: PROCEED`
   * If missing archive files or unrestored deletions detected →
     `recommendation: HALT — restore archives`
   * On `HALT`: release the lock and report. Ship must restore archives before committing.

6. **Release lock** (acquired in step 1 of pre-mode):
   Invoke `file-lock` release for `.backlogit/queue/{shipment_id}.md`.
   If release fails, log a warning — stale locks are operator-recoverable.

### Step 0(c) Command Routing: cascade-close --classify-only (192-F)

<!-- cascade-close-routing:BEGIN step-0c -->
**Command routing (192-F): `--classify-only` first, on every close.**
Every close, on either path, first runs
`autoharness shipment cascade-close --classify-only --shipment <shipment_id> --feature <feature_id> --sha <merge_commit_sha> --json`,
where `<feature_id>` is the shipment's covering feature (derived as
safe-close step 2 derives it). That command performs (a)–(c) below:
the manifest load, the Step 0(b) snapshot, the classification, the
engine-semantics gate, the close-path selection, and the
linked-deliberation disposition snapshot. It records them in the
evidence record
`docs/closure/evidence/{shipment_id}-{feature_id}-close-evidence.json`
(the fixed path `build_evidence_path` builds), whose
`pre_close.engine_semantics`, `pre_close.close_path_selection`, and
`pre_close.linked_deliberation_disposition` fields carry the Step 0(c)
engine-semantics decision, close-path selection, and disposition
snapshot. The same record carries the path-specific baseline:
`pre_close.observation_set` on a selected `safe_close`, and
`pre_close.out_of_manifest_descendants` on either path. The prose in
(a)–(c) below stays the specification the command implements; Ship
never re-derives these values by hand.

Plain `--classify-only` needs no operator approval: it is no-clobber
(an existing `pre_close` record exits 2 and is left untouched) and only
creates a new record. `--classify-only --replace-pre-close` overwrites
an existing `pre_close` record and needs the same destructive-command
approval as the mutating run. The mutating
`autoharness shipment cascade-close` invocation (no `--classify-only`)
**is** the destructive command: it needs the same operator approval
that a direct `backlogit shipment ship` call needs (explicit operator approval
for this closure; no automated check, allowlist, or activation record satisfies
it). A direct
`backlogit_ship_shipment` MCP or `backlogit shipment ship` CLI call is
a **P-005 deviation** on either path, and the closure-evidence gate
refuses the closure it produces, because no valid evidence record
backs it.

Route on the command's exit code:

| Mode | Exit | Skill action |
|---|---|---|
| `--classify-only` | 0 (`CASCADE`) | Obtain destructive-command approval, then run the mutating `cascade-close` (Cascade Close Sub-Procedure) |
| `--classify-only` | 3 (`SAFE_CLOSE` selected: classifier `SAFE_CLOSE`, or engine `UNVERIFIED`); exit 3 AND `selected_close_path` is `safe_close`; otherwise HALT | Safe-close steps 1–10, then the Linked-Deliberation Disposition step, citing the verdict record as `close_evidence` |
| mutating | 0 | Run the Linked-Deliberation Disposition step with its inputs from the evidence record, then write the closure artifact with `close_path: cascade` and `close_evidence` |
| mutating | 3 | Reached only when no `cascade`-selected `--classify-only` record preceded the run, which the routing above never does. Nothing was mutated: HALT, operator review, and never `SAFE_CLOSE` on this result (the Cascade Close Sub-Procedure's No-substitution rule). A selected-path change after a `CASCADE` `--classify-only` selection is exit 4, never 3 |
| either | 2, 4 | HALT. Nothing was mutated. Exit 4 includes an engine re-probe difference and any difference from the `cascade`-selected `--classify-only` record, and is never answered with `SAFE_CLOSE`. Fix the input, or ask the operator |
| either | 5, 6, 7, 8 | HALT. Operator review. No commit of the backlog root, no retry, no direct call |
| either | any other or unrecognized exit (including 1) | HALT (CLI too old or unexpected; do not fall back to cascade). Never call `backlogit shipment ship` directly and never substitute `SAFE_CLOSE`; ask the operator |

**Minimum version**: The pre-flight probe is the capability test:
`autoharness shipment --help` must exit 0 and list `cascade-close`; an older CLI
exits 1 and HALTs. Do not infer capability from the version string. Upgrade
autoharness before closing; there is no hand-run fallback.
<!-- cascade-close-routing:END step-0c -->

**Approval is not waivable by dark mode.** The approval for the mutating `autoharness shipment cascade-close` run is not waivable by dark mode: a `DARK_MODE_ACTIVE` activation record under P-017 never satisfies it, and it is required in dark mode on the same terms as in interactive operation (Constitution Principle VII, Destructive Command Approval, NON-NEGOTIABLE).

**Consume `selected_close_path`.** Ship reads the close path from the `selected_close_path` field of the command's `--json` output (mirrored as `pre_close.close_path_selection` in the evidence record) and never routes on `classifier_verdict` alone. A `CASCADE` classifier verdict with an `UNVERIFIED` engine verdict selects `safe_close`, and only `selected_close_path: cascade` in the command's JSON output (recorded in the evidence record as `pre_close.close_path_selection`) permits the mutating run.

**Exit codes (verbatim from `autoharness shipment --help`):**

```text
  0  mutating: all postconditions passed; --classify-only: CASCADE selected.
  2  input, I/O, no-clobber refusal, planning error, or pre-invocation write failure.
  3  the selected close path is not CASCADE; verdict recorded, nothing invoked.
  4  pre-invocation revalidation drift; nothing invoked. Never answer with SAFE_CLOSE.
  5  a postcondition failed (operator review).
  6  backlogit exited non-zero, timed out, or its output did not parse (operator review).
  7  the pair lock is already held, or an existing evidence record needs review
     (invoking, unreadable, ambiguous, or another pair); nothing invoked
     (operator review). The lock is
     .autoharness/gates/cascade-close/{S}-{F}.lock; it is never broken
     automatically, and removing a stale lock is an operator-only action
     (agents must not delete it).
  8  the post-close evidence write failed after invocation (operator review).
Exits 5, 6, 7, and 8 forbid committing the backlog root, re-running
cascade-close, calling `backlogit shipment ship` directly, or substituting
SAFE_CLOSE until an operator reviews the backlog state and the record.
```

### Safe-Close Mode

Runs **in place of** the destructive cascade `backlogit_ship_shipment` call —
**except** when the `--classify-only` route in Step 0(c) Command Routing (above)
exits 0 (`CASCADE`), in which case only the mutating `autoharness shipment
cascade-close` command runs, after per-closure operator approval, and safe-close
steps 1–10 are skipped entirely. Archives only the shipment manifest's explicit
item IDs, one artifact at a time, verifying after each archival that the parent
feature and any unshipped sibling tasks survive. Invoked between pre-mode
(`PROCEED`) and post-mode, under the lock pre-mode already holds. If invoked
standalone, acquire the lock per pre-mode step 1 first and release it on
completion. A standalone invocation first runs `autoharness shipment cascade-close --classify-only` and proceeds only on exit 3 with `selected_close_path` `safe_close`; a stale record alone never authorizes the path.

0. **Load manifest, snapshot pre-close state, and confirm the recorded close path
   (the path is selected by the `--classify-only` route, never re-selected here)**:
   safe-close runs only on the exit 3 `safe_close` record.
   a. **On the SAFE_CLOSE path, load the manifest first**: invoke `backlogit_get_shipment(shipment_id)` and extract the
      `items` list. This load happens here in Step 0 — not deferred to step 1
      below — because the Step 0(b) snapshot requires it. The cascade path skips steps
      1–10 entirely and takes its baseline from the `pre_close` section of the
      evidence record instead.
   b. **On the SAFE_CLOSE path, snapshot pre-close `parent_id` and declared `status` for every explicit
      manifest item regardless of `artifact_type`**, by reading each item's current
      frontmatter from
      whichever of `.backlogit/queue/` or
      `.backlogit/archive/` currently contains it — a manifest
      task item may already be
      pre-archived when this snapshot runs (see the
      Cascade Close Sub-Procedure's pre-archived-member preamble below), and
      its snapshot must still be captured from wherever it actually resides.
      If a task item's record is found in **both** locations (an
      ambiguous/torn state) or in **neither** (missing), halt immediately
      with `RECONCILE_FAIL_SNAPSHOT_AMBIGUOUS` or
      `RECONCILE_FAIL_SNAPSHOT_MISSING` respectively — never guess which
      copy or location is authoritative. Retain this snapshot in memory
      for the duration of this close operation; it is the SAFE_CLOSE baseline for this close operation (the cascade
      path's baseline is the `pre_close` section of the evidence record, never
      this snapshot), and it must
      be captured **before** any mutating call (cascade or otherwise) runs,
      never reconstructed after the fact.

      **Declared `status` is read from the record's own frontmatter `status`
      field — never inferred from, nor substituted by, which of
      `queue/`/`archive/` currently holds the record.** A record residing in
      `.backlogit/archive/` while declaring `status: done` is
      **not** truly archived; only a declared `status: archived` counts as
      truly archived for the Cascade Close Sub-Procedure's step 3 gate
      below — location alone is never sufficient. This declared-status
      snapshot MUST be captured here, in Step 0(b), **before** the cascade
      invocation, for the identical reason already stated for `parent_id`:
      `status` is the very field the cascade mutates, so a post-close read
      would report `archived` for everything the cascade just archived,
      collapsing `required_ids` (Cascade Close Sub-Procedure step 3) to
      empty and silently disabling the completeness check entirely. Never a
      freshly-read or assumed value. An absent, null, or unparseable `selected_close_path`, `pre_close` snapshot entry, `parent_id`, `declared_status`, `out_of_manifest_descendants`, or `linked_deliberation_disposition` value is a HALT; a missing value is never read as empty or as not archived.
   c. **Entry gate (no classification in this mode)**: entry is only after
      `cascade-close --classify-only` exited 3 and the `selected_close_path` field of the command's JSON output
      (recorded in the evidence record as `pre_close.close_path_selection`) is
      `safe_close`. Do not re-classify to CASCADE here. If
      the evidence record's `pre_close.close_path_selection` is not `safe_close`, HALT with
      operator review and do not call the cascade operation. Safe-close never
      evaluates the P-015 classification to choose a path. `cascade` is permitted
      only when the P-015 INV-6 engine-inertness containment holds and the command
      returned exit 0.

1. **Load manifest** via `backlogit_get_shipment(shipment_id)`. Extract the
   `items` list. These IDs are the **only** artifacts safe-close may move or
   archive.

2. **Compute the protected set** (partial-feature detection):
   * Derive the covering feature ID from the manifest item hierarchy
     (e.g. a task `055.002-T` belongs to feature `055-F`).
   * If the covering feature ID is **not** in the manifest `items`, this is a
     **partial-feature shipment**. Add the covering feature to the protected set.
   * Enumerate every task sharing the covering feature's hierarchy prefix whose ID
     is **not** in the manifest `items` (the unshipped siblings) by scanning **both**
     `.backlogit/queue/` **and** `.backlogit/archive/` (plus the
     feature file's declared children when available). Add each to the protected set.
   * **Sequence-aware exclusion (serial partial-feature shipments)**: when a sibling
     belongs to a predecessor shipment in the same feature-split sequence, exclude it
     from the protected set **ONLY** when that predecessor shipment record itself has
     **verified archived provenance** `archived_status: shipped` (or normalized legacy
     `done`). Mere archive-file presence, `archived_status: active|queued|blocked|abandoned`,
     generic `status: archived` without shipped/done provenance, or missing/ambiguous
     provenance are **NOT** sufficient — in those cases the sibling stays protected
     fail-closed.
   * The **protected set** is the parent feature plus every unshipped sibling task
     that MUST remain in `.backlogit/queue/` after closure. It is computed
     from **expected IDs**, not merely the files currently present in queue, so a
     sibling or parent that was already wrongly archived is still detected.

3. **Baseline integrity gate** (before archiving anything): Run
   `git status --short -- ".backlogit/"` and record the pre-closure
   working-tree state so any later archival or deletion of a protected-set path can be
   attributed to this procedure. Then confirm **every** protected-set member currently
   exists in `.backlogit/queue/`. If any protected-set member is already in
   `.backlogit/archive/` or missing from the working tree, a cascade has
   **already** occurred (or the shipment scope is wrong): halt immediately with
   `HALT — cascade detected, revert required`, name the affected artifact IDs, and do
   NOT archive any manifest item. The `pre-archived` exemption (step 4) applies to
   **manifest items only** — never to the protected set.

4. **Archive each manifest item individually** (loop over `items` ONLY):
   * If the item's file is in `.backlogit/queue/`: move it to
     `done` via `backlogit_move_item`, then archive that single artifact via
     `backlogit_archive_item` (CLI fallback `backlogit archive {{id}}`). When the
     backlog registry's `archive_item` operation supports commit metadata, record
     the merge SHA using that tool's configured field (for backlogit, `commit_sha`);
     otherwise record the merge SHA in the closure report. Classify `matched`.
   * If the item's file is already in `.backlogit/archive/`: classify
     `pre-archived` and **skip** — do not re-archive. Reusing the `pre-archived`
     classification prevents false-positive cascade flags on items that were
     legitimately shipped earlier.
   * If the item's file is in neither location: classify `missing`, halt with
     `RECONCILE_FAIL`, and do not continue archiving.

5. **Verify-after-each invariant** (run immediately after each item's archival):
   * Confirm **every** protected-set member is still present in
     `.backlogit/queue/` — not moved to `.backlogit/archive/`,
     not deleted from the working tree.
   * Run `git status --short -- ".backlogit/"` and confirm no
     protected-set path appears as a deletion, rename into `archive/`, or new
     `archive/` addition beyond the baseline captured in step 3.
   * The protected set was proven fully present in queue at the baseline gate
     (step 3), so **any** protected-set member now found in `archive/` or missing from
     the working tree is a cascade. There is **no** pre-archived exemption for the
     protected set — the exemption in step 4 covers manifest items only.

6. **git-revert-on-cascade**: If the invariant fails (a protected-set artifact was
   archived or deleted by the preceding archival):
   * **Cascade detected.** Immediately restore the unintended change, one path at a
     time: `git restore -- <path>` for each path that this run touched (the paths the
     step-5 `git status --short -- ".backlogit/"` output shows as changed relative to
     the step-3 baseline), never a whole directory such as `.backlogit/queue/` or
     `.backlogit/archive/`, for working-tree moves/deletions, or `git revert <commit>` if
     the cascade was already committed.
   * Re-run the invariant to confirm the protected set is intact again.
   * Halt with `HALT — cascade detected, revert required`, emit a **P-005**
     violation event (naming the cascaded artifact IDs), and do **NOT** commit the
     backlog state. Do not auto-prune the manifest.

7. **Final invariant re-check**: After the loop completes, re-confirm the full
   protected set is intact in `.backlogit/queue/`.

8. **Close the shipment record itself** (single artifact, non-cascading; authoritative order):
   * Move **ONLY** the live shipment record to `status: shipped` via the generic,
     non-cascading `backlogit move <shipment_id> --status shipped`.
   * Re-read and verify the live shipment record now reports `status: shipped`. If the
     record remains `active`, is already `archived`, is missing, or resolves to any other
     shape, halt fail-closed with `RECONCILE_FAIL_SHIPMENT_RECORD_LIVE_STATUS`. Do **NOT**
     auto-retry by calling the cascade op, and do not archive an `active` shipment record.
   * Archive **ONLY** the shipment record via `backlogit archive <shipment_id>` (single-
     artifact archive; this stamps `archived_status` from the live status at archive time).
   * Re-read and verify the archived record now reports `archived_status: shipped`.
     A legacy `archived_status: done` is accepted only when it pre-existed as an older,
     already-correct terminal provenance. Missing archive, live+archived duplication,
     generic archived-without-provenance, or any non-shipped archived provenance halt
     fail-closed with `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`.
   * Re-run the verify-after-each invariant (step 5) to confirm the protected set is still
     intact after the shipment-record close sequence.

9. **Produce safe-close report** per the same schema, recording the protected set,
   each item's classification, the shipment-record move-to-shipped verification, the
   shipment-record archived-provenance verification, and the recommendation.

10. **Gate decision**:
    * All manifest items `matched` or `pre-archived`, the shipment record archived,
      and the protected set intact → `recommendation: CLOSED`. Run the Linked-Deliberation
      Disposition step; only `DISPOSITION_COMPLETE` proceeds to post-mode.
    * Any cascade detected → `recommendation: HALT — cascade detected, revert required`
      (see step 6). Do not proceed to the commit step.

### Cascade Close Sub-Procedure (P-015 `CASCADE` exception ONLY)

Runs **only** when the `--classify-only` route in Step 0(c) Command Routing (above)
exits 0 (`CASCADE` selected), never from Safe-Close Mode Step 0. The cascade
baseline is the pre-close record written by `autoharness shipment cascade-close --classify-only`
(the `pre_close` section of the evidence record); the mutating
command re-verifies that baseline itself. Safe-close Step 0(b) and 0(c) do not
apply to the CASCADE path. This sub-procedure never reloads the manifest or
attempts to reconstruct pre-close state after the fact.
Replaces steps 1–10 above entirely for this shipment's closure; there is no
partial mixing of the two paths.

The mutating `autoharness shipment cascade-close` command is the ONLY caller of the cascade primitive (`backlogit shipment ship`, MCP `backlogit_ship_shipment`) in this workflow. No other path, including Ship, invokes the cascade primitive directly.

<!-- cascade-close-routing:BEGIN cascade-sub-procedure -->
**Command routing (192-F): the mutating `cascade-close` runs this
sub-procedure.** The pre-invocation revalidation below (the classifier
re-run, the linked-deliberation re-collection, and the engine-semantics
re-probe), the baseline-fingerprint capture, step 1's invocation, and
steps 2–6 are performed by the mutating
`autoharness shipment cascade-close --shipment <shipment_id> --feature <feature_id> --sha <merge_commit_sha> --message=<merge_commit_message> --author=<merge_commit_author> --json`. Pass both as argv in the equals form; values must not begin with `-` and must not contain newlines or control characters.
Ship runs it only after the Step 0(c) `--classify-only` run exited 0
(`CASCADE` selected) and only with the destructive-command approval Step
0(c) names. Before it invokes anything, the command re-checks the
`cascade`-selected `--classify-only` record: any difference from that
record, and any engine re-probe difference, exits 4 with nothing mutated.
Its exit 0 is step 6's `recommendation: CLOSED`, and the record it
finalizes (phase `post_close`) is step 5's cascade-close report and the
closure artifact's `close_evidence`. Every other exit follows the Step 0(c)
routing table: HALT, never a `SAFE_CLOSE` substitution. Ship never performs
these steps by hand and never invokes the cascade operation directly: a
direct `backlogit_ship_shipment` MCP or `backlogit shipment ship` CLI call is
a **P-005 deviation**. The command never archives a disposition-set
deliberation; the Linked-Deliberation Disposition step stays its only
archiver. The prose below stays the specification the command implements.

**Timeout sizing (192-F).** The mutating run's `--timeout` is 30-3600
seconds (default 1800). On expiry the command kills the engine mid-cascade
and exits 6 (`mutation_possible: indeterminate`), so Ship sizes it before
the run. Let N = |closure_scope(S)| (the manifest items plus the shipment
record) and compute the budget `B = ceil(1.5 * (F + P * N))` seconds, where
F and P are the fixed and per-artifact `backlogit shipment ship` costs
re-derived from the workspace's own backlogit event logs (reference values,
backlogit 1.11.0: F = 133 s, P = 35 s). If B <= 1800, the default suffices.
If 1800 < B <= 3600, Ship adds `--timeout B` to the invocation above. If
B > 3600, Ship does not start the mutating run and HALTs for an operator
decision; it never invokes the command with a timeout known to be too
short.
The B > 3600 HALT bounds `--timeout`, not the whole command: `--timeout`
bounds only the `backlogit shipment ship` child, and the command's runner
alone owns that child's expiry. The engine probes, preflight, and
revalidation run before that timer starts, and child cleanup, postcondition
checks, evidence persistence, and lock release run after it, so Ship
supervises the command for its supervision budget: the effective timeout (B,
or the 1800 default) + a fixed 600 s supervision margin (two 30 s engine
probes, the bounded kill and reader-join cleanup of each spawned process, at
most 70 s each, and headroom for the untimed workspace reads and writes), at
most 4200 s. Whenever the supervision budget exceeds the agent runtime's
synchronous tool-call limit, Ship MUST start the mutating run in the
runtime's background/async mode, attached to the session (never detached
from it), and poll until the command exits; if the runtime cannot keep the
process alive for the full supervision budget, Ship HALTs before the
mutating run. A runtime that kills the command mid-run leaves `backlogit`
running unsupervised in its own process group, the evidence record at
`invoking`, and the pair lock in place, so the next run exits 7. Ship stays
attached to the run until the command exits and never abandons or kills it
on an agent-tool wait or before the supervision budget elapses. For this run
the supervision budget is the stall bound: the circuit-breaker Stall
Detection table's `autoharness shipment cascade-close` (mutating mode) row
sets it in place of the "Other commands" 5-minute stall timeout.
<!-- cascade-close-routing:END cascade-sub-procedure -->

**Approval is not waivable by dark mode.** Approval for the mutating `autoharness shipment cascade-close` run is never waived by dark mode, under P-017 or otherwise, and is required on the same terms as in interactive operation (Constitution Principle VII, Destructive Command Approval, NON-NEGOTIABLE).

**Pre-archived manifest members (expected and tolerated)**: before invoking
step 1 below, classify each manifest member's **location** as `queued` or
`pre-archived` by checking whether its record currently resides in
`.backlogit/queue/` or `.backlogit/archive/`, and
retain that location set for the cascade-close report (step 5 below). This
location label is **descriptive only** — it names where the record
currently resides, and is **never** a substitute for, nor evidence of, the
record's own declared `status` field. A record residing in
`.backlogit/archive/` while declaring `status: done` is **not**
truly archived; only a declared `status: archived` is truly archived for
step 3's two-set gate below. The sole authority for "truly archived" is the
declared-status pre-close snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`) already captured before this
sub-procedure runs.

A `pre-archived` (by location) manifest member is **expected and tolerated
on this path**: it does **not** disqualify the `CASCADE` verdict, does
**not** constitute a classifier ambiguity or unresolved precondition, and
does **not** authorize a fallback to safe-close. Step 0(c)'s classifier
already resolves each manifest member by scanning **both** `queue/` and
`archive/`, so archived inputs were already accounted for when the
verdict was selected — this clause states the execution-time consequence
of that fact and does not add a new classifier precondition; Step 0(c)'s
precondition wording is unchanged.

This tolerance applies to **manifest members only** — it does not
restate, weaken, or cross-apply to the protected set, which has no
pre-archived exemption (see Safe-Close Mode steps 3/5 above). A manifest
that qualifies for `CASCADE` has no protected set by construction (the
P-015 INV-6 engine-inertness predicate governs out-of-manifest descendants), so no
protected set arises on this path.

**`archived_ids` is a transition log, not a manifest echo.** The cascade
operation invoked in step 1 below reports, in `archived_ids`, only the
artifacts it **actually transitioned** to archived during that invocation
(backlogit engine source, `internal/core/shipment_lifecycle.go`
`archiveItems()`: an item whose declared `status` is already `archived` is
skipped and never appended to the slice that becomes `archived_ids`). A
manifest task item that was already truly `status: archived` before the call
therefore has no transition to report and is **correctly absent** from
`archived_ids` — this is expected engine behavior, not an anomaly and not a
cascade failure. **This never extends to the shipment record or to a
qualifying feature member itself** (155-S, PR #407 review, thread
PRRT_kwDORzpWpM6b0kit): step 3 below makes both unconditionally required
regardless of their own pre-close declared status, so neither can ever be
"correctly absent" the way a task item can — see
step 3 for the full statement of that rule. The live fail-closed guard
over this result is the two-set `allowed_ids` / `required_ids` gate
specified in step 3 below, evaluated **against the pre-close
declared-status snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`)** — never against location, and never against a
post-close re-read.

   **SUPERSESSION NOTE (155-S, 2026-08-24).** This paragraph previously
   claimed the cascade operation "is idempotent over pre-archived members",
   citing
   `docs/spikes/2026-08-18-cascade-close-pre-archived-member-behavior.md`
   as authority for the claim that it "returns \[pre-archived members] in
   its `archived_ids` result exactly as it does newly-archived members",
   and stated that step 3's exact-match post-condition, "evaluated against
   the manifest's full item set", "must never be relaxed". **That claim,
   and the spike cited for it, are WITHDRAWN.** The spike's arms were built
   with `move --status done`, which relocates a record but leaves it
   declaring `status: done` — never truly `status: archived` — so none of
   its arms ever exercised the case this paragraph claimed to cover; its
   finding is valid only for relocated-but-`done` records (see the spike's
   own superseded banner). The two safety properties this paragraph
   protected — nothing out-of-scope archived, nothing required left
   unarchived — are now carried, at full strength, by the two-set gate in
   step 3 below, keyed on declared pre-close status rather than a full-set
   echo of the manifest.

**No-substitution rule**: once the `--classify-only` route selects `cascade` (exit 0), that verdict is
final for this closure — between the verdict and step 1's invocation
below, substituting manual safe-close is a **P-005 process deviation**,
never a permitted fallback, regardless of the manifest's archival state.
This complements, and does not restate or contradict, step 2's separate
rule against falling back to safe-close **after** a cascade has already
executed: together the two rules close both the pre-execution and
post-execution substitution windows. The asymmetry is intentional and
one-directional — this rule forbids `CASCADE -> manual safe-close`
substitution only; it grants no license to invoke cascade when Step 0
selects `safe_close`, which remains governed by the P-015 default
prohibition. If a genuine unhandled error occurs during the cascade
operation, halt and disclose it per the verification steps below — never
silently switch to safe-close instead.

Ship performs **no** manual per-item archive loop on this path: the
cascade operation in step 1 below performs all remaining archival
itself, consistent with the "no partial mixing of the two paths" rule
above.

1. Executed only by the mutating `autoharness shipment cascade-close` command, which
   is the single caller of the cascade primitive. No agent step invokes
   `backlogit_ship_shipment` or `backlogit shipment ship`.
2. **Verify the result matches the classifier's own precondition**:
   `returned_ids` MUST be empty (`[]`). A non-empty `returned_ids` means the
   live engine found an unreleased descendant the classifier's live-workspace
   enumeration did not — this is a TOCTOU/engine-behavior mismatch, not a
   recoverable state. Halt immediately with
   `HALT — cascade returned non-empty returned_ids, classifier/engine
   mismatch` and emit a **P-005** violation; do NOT retry, do NOT fall back to
   safe-close after a cascade has already executed.
3. **Verify `archived_ids` against the two-set `allowed_ids` / `required_ids`
   gate** (replaces exact full-set equality — see the SUPERSESSION NOTE
   above and the P-015 policy's own supersession note for why):
   * Apply the flat `allowed_ids` / `required_ids` postcondition sets defined in the
     section "Cascade Close Postconditions: Flat" (single definition; do not restate
     them here).
   * **Two separately-labelled, independently-failing conditions.** Neither
     may be evaluated as a precondition of the other, and the two MUST NOT
     be merged into a single combined test (conflating two questions into
     one condition is the documented root cause of external defect
     `B57F9E24`):
     - **Unexpected-artifact check**: if `archived_ids - allowed_ids` is
       non-empty, halt with
       `HALT — cascade archived unexpected artifact {id}` and emit a
       **P-005** violation.
     - **Missing-required-artifact check**: if `required_ids - archived_ids`
       is non-empty, halt with
       `HALT — cascade did not archive required artifact {id}` and emit a
       **P-005** violation.
   * An `allowed_ids` **non-shipment, non-feature** member (a manifest item that is
     neither the shipment record nor a qualifying feature member; the qualifying
     feature member itself is unconditionally required, see below) that was already
     truly `status: archived` in the pre-close snapshot MAY be included in or omitted
     from `archived_ids` by the engine — neither outcome fails either check (it is
     outside `required_ids` by construction, and if present in `archived_ids` it is
     still inside `allowed_ids`). **This tolerance never extends to the shipment
     record itself**, which is unconditionally a `required_ids` member per
     the computation above regardless of its own
     pre-close declared status: if the shipment record were ever reported
     pre-close as already truly `status: archived` — an anomalous state for
     an artifact this same closure step is actively transitioning to
     `shipped` — its absence from `archived_ids` still fails the
     missing-required-artifact check exactly as any other missing
     `required_ids` member would; no engine behavior toward the shipment
     record ever gets a pass under this tolerance.

     **Nor does it extend to a qualifying feature member itself (155-S, PR
     #407 review, thread PRRT_kwDORzpWpM6bzlFl).** Backlogit's own
     `ShipShipment` (`internal/core/shipment_lifecycle.go`) unconditionally
     calls `setArtifactStatus(featureID, models.StatusDone, "feature
     released")` for every explicit shipment-member feature — regardless of
     that feature's own pre-close declared status, including an already
     truly `status: archived` one — **before** `collectArchiveCandidateIDs`
     runs. `setArtifactStatus` only no-ops when the artifact's current
     status already equals the requested one, and the requested status here
     is `done`, never `archived`, so an already-archived qualifying feature
     is unconditionally relocated to `done` first, with no terminal-status
     bypass of the kind `completeReleaseScope` grants a manifest task item
     already truly `status: archived` (that task-only skip is exactly what
     makes the tolerance above valid for tasks, and it has no counterpart in
     the feature-forcing loop). By the time `collectArchiveCandidateIDs`
     loads the feature, its declared status is therefore always `done`,
     never still `archived` — that function's own
     `feature.Status != models.StatusArchived` check is always true for it
     — so the feature is always appended to the candidate list
     `archiveItems` archives. A qualifying feature member can therefore
     never be "correctly absent" from `archived_ids` the way a truly
     pre-archived manifest task item can — its
     absence is always an anomaly, never expected engine behavior. A
     qualifying feature member is therefore an unconditional `required_ids`
     member exactly like the shipment record, and is never eligible for
     this tolerance.
4. **Verify no `parent_id` was cleared**: re-read every archived task's
   frontmatter and confirm `parent_id` is unchanged from the pre-close
   snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`) — never a freshly-read or assumed value,
   since the field being verified is the very one a cascade could have just
   cleared. Any cleared or altered `parent_id` is a
   cascade-detection failure equivalent to step 6 of safe-close: halt with
   `HALT — cascade cleared parent_id on {id}, revert required` and emit a
   **P-005** violation; do NOT commit the mutated backlog state.
5. **Produce cascade-close report** recording the classifier's verdict,
   qualifying feature IDs
   (all Step 0(c)), the pre-close declared-status
   snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`), the `backlogit_ship_shipment` result
   (`shipment_status`, `archived_ids`, `returned_ids`, `commit_sha`),
   `allowed_ids`, `required_ids`, and both set differences
   (`archived_ids - allowed_ids` and `required_ids - archived_ids`) — so a
   vacuous `required_ids` is visible in the report rather than silent — and
   the parent_id-preservation verification outcome (against the pre-close
   snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`)).
6. **Gate decision**: `returned_ids` empty, `archived_ids - allowed_ids`
   empty (no unexpected artifact archived), `required_ids - archived_ids`
   empty (no required artifact left unarchived), and every `parent_id`
   preserved (against the pre-close snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`)) →
   `recommendation: CLOSED`. Run the Linked-Deliberation Disposition
   step; only `DISPOSITION_COMPLETE` proceeds to post-mode. Any verification failure above → the corresponding `HALT`; do
   not proceed to post-mode or any commit step.

### Cascade Close Postconditions: Flat `allowed_ids` / `required_ids` Sets

3. **Verify `archived_ids` against the two-set `allowed_ids` / `required_ids`
   gate** (replaces exact full-set equality — see the SUPERSESSION NOTE
   above and the P-015 policy's own supersession note for why):
   * **Compute `allowed_ids`** = `closure_scope(S)`: every explicit manifest
     item ID from `items(S)`, regardless of `artifact_type`, plus the
     shipment record itself. This is the flat set. A disposition-set
     deliberation is **not** added merely because it is linked from a
     manifest member; linked deliberations are handled only by INV-12.
     H10 carve-out: if a deliberation ID is itself an explicit manifest
     member, it is an ordinary `closure_scope(S)` / `allowed_ids` member,
     and its presence in `archived_ids` MUST NOT trip the
     unexpected-artifact check merely because its `artifact_type` is
     `deliberation`.
   * **Compute `required_ids`** = the shipment record and every qualifying
     feature member (**both unconditionally** — never omitted, and never
     conditioned on either artifact's own pre-close declared status) + every
     other manifest item that was **not** truly `status: archived` in the all-member pre-close declared-status snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`). Because the pre-close snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`) covers every explicit manifest member regardless of
     `artifact_type`, no Step 0(c) linked-deliberation extension participates
     in this set. A disposition-set deliberation that is not an explicit
     manifest member is never a `required_ids` member.
   * **Two separately-labelled, independently-failing conditions.** Neither
     may be evaluated as a precondition of the other, and the two MUST NOT
     be merged into a single combined test (conflating two questions into
     one condition is the documented root cause of external defect
     `B57F9E24`):
     - **Unexpected-artifact check**: if `archived_ids - allowed_ids` is
       non-empty, halt with
       `HALT — cascade archived unexpected artifact {id}` and emit a
       **P-005** violation. A disposition-set deliberation that is not an
       explicit manifest member and appears in `archived_ids` fails this
       check: that is engine drift under the verified flat line.
     - **Missing-required-artifact check**: if `required_ids - archived_ids`
       is non-empty, halt with
       `HALT — cascade did not archive required artifact {id}` and emit a
       **P-005** violation.
   * An `allowed_ids` **non-shipment, non-feature** member (any explicit
     manifest item that is neither the shipment record nor a qualifying
     feature member, including an explicit-member deliberation under H10)
     that was already truly `status: archived` in the all-member
     pre-close snapshot recorded in the evidence record (the pre_close section written by `autoharness shipment cascade-close --classify-only`) MAY be included in or omitted from `archived_ids` by
     the engine — neither outcome fails either check (it is outside
     `required_ids` by construction, and if present in `archived_ids` it is
     still inside `allowed_ids`). **This tolerance never extends to the
     shipment record itself**, which is unconditionally a `required_ids`
     member per the computation above regardless of its own pre-close
     declared status: if the shipment record were ever reported pre-close as
     already truly `status: archived` — an anomalous state for an artifact
     this same closure step is actively transitioning to `shipped` — its
     absence from `archived_ids` still fails the missing-required-artifact
     check exactly as any other missing `required_ids` member would; no
     engine behavior toward the shipment record ever gets a pass under this
     tolerance.

     **Nor does it extend to a qualifying feature member itself (155-S, PR
     #407 review, thread PRRT_kwDORzpWpM6bzlFl).** Backlogit's own
     `ShipShipment` (`internal/core/shipment_lifecycle.go`) unconditionally
     calls `setArtifactStatus(featureID, models.StatusDone, "feature
     released")` for every explicit shipment-member feature — regardless of
     that feature's own pre-close declared status, including an already
     truly `status: archived` one — **before** `collectArchiveCandidateIDs`
     runs. `setArtifactStatus` only no-ops when the artifact's current
     status already equals the requested one, and the requested status here
     is `done`, never `archived`, so an already-archived qualifying feature
     is unconditionally relocated to `done` first, with no terminal-status
     bypass of the kind `completeReleaseScope` grants a non-feature manifest
     item already truly `status: archived` (that non-feature skip is exactly
     what makes the tolerance above valid for those members, and it has no
     counterpart in the feature-forcing loop). By the time
     `collectArchiveCandidateIDs` loads the feature, its declared status is
     therefore always `done`, never still `archived` — that function's own
     `feature.Status != models.StatusArchived` check is always true for it
     — so the feature is always appended to the candidate list
     `archiveItems` archives. A qualifying feature member can therefore
     never be "correctly absent" from `archived_ids` the way a truly
     pre-archived non-feature manifest member can — its absence is always an
     anomaly, never expected engine behavior. A qualifying feature member is
     therefore an unconditional `required_ids` member exactly like the
     shipment record, and is never eligible for this tolerance.

### Linked-Deliberation Disposition (P-015 INV-12)

Runs once per closure, after the selected close path's gate returns
`recommendation: CLOSED` (safe-close step 10, or Cascade Close Sub-Procedure
step 6) and before post-mode, under the lock pre-mode already holds. The step
is path-independent and separately sanctioned by **INV-12**: it never widens
`closure_scope(S)`, `allowed_ids(S)`, or `required_ids(S)`, and it mutates a
disposition-set deliberation only through a single-artifact, non-cascading
archive, one ID at a time.

0. **Inputs.** The step consumes these values from Step 0 and the selected
   close path, and never substitutes a recomputed set for the snapshot.
   <!-- cascade-close-routing:BEGIN disposition-inputs -->
   It takes every input from the evidence record
   (`docs/closure/evidence/{shipment_id}-{feature_id}-close-evidence.json`,
   written by the Step 0(c) `autoharness shipment cascade-close` run), never
   from in-session Step 0(c) state, and never recomputes one:
   <!-- cascade-close-routing:END disposition-inputs -->
   * the selected close path and its reason (Step 0(c) close-path selection),
     from `pre_close.close_path_selection`;
   * the Step 0(c) engine-semantics decision (`probe_surface`, `version`,
     `commit`, and verdict/reason), from `pre_close.engine_semantics`,
     rebuilt with `engine_semantics_from_record` (this self-hosting
     repository's own implementation lives at
     `src/autoharness/gates/cascade_evidence.py`);
   * the Step 0(c) linked-deliberation disposition snapshot (the
     **disposition snapshot**), from
     `pre_close.linked_deliberation_disposition`;
   * the path-specific baseline: the safe-close step 3 observation-set
     fingerprints, from `pre_close.observation_set` (`safe_close`), or the
     Cascade Close Sub-Procedure's out-of-manifest descendant baseline
     fingerprints, from `pre_close.out_of_manifest_descendants` (`cascade`).

   The input source is the only thing the evidence record changes here:
   steps 1–6 below run exactly as stated, with the recorded engine decision
   as step 1's `engine=` argument. Under an `UNVERIFIED` engine the step
   still mutates nothing, and it stays the only archiver of a
   disposition-set deliberation.

1. **Plan.** Workspaces with a Python implementation installed plan each
   disposition-set deliberation's outcome with
   `compute_linked_deliberation_disposition(manifest_items, shipment_id, workspace_backlog_dir, engine=<the Step 0(c) engine-semantics decision>)`
   (this self-hosting repository's own implementation lives at
   `src/autoharness/gates/shipment_closure.py`). The call passes no
   `stash_path`, so the planner reads the active stash file at
   `.backlogit/stash.jsonl`. The planner result is consumed in this order. Check `planning_error` first: when `planning_error` is present, it
   is exempt from the planner-vs-snapshot equality check because the
   planner returns `dispositions=()`; the equality check applies only
   when the planner returns without `planning_error`. Before any archive,
   when the planner returns without `planning_error`, the planner's disposition
   set (deliberation IDs, link kinds, linking members, record paths, and
   record hashes) MUST equal the Step 0(c) disposition snapshot, and for
   every deliberation the snapshot already settled as `retained_read_error`,
   `retained_ambiguous`, or `already-archived`, the planner's outcome and
   `reason_code` (and, for `retained_read_error`, its `path`) MUST equal the
   snapshot's settled outcome data, and the planner's `unresolved_references`
   set (exact `{id, reason_code}` pairs) MUST equal the snapshot's,
   so a fresh plan never upgrades snapshot-settled evidence to a planned
   `archive` or drops snapshot evidence. Any difference halts with `HALT — linked-deliberation disposition failed {id}`;
   emit **P-005** once; no mutation. `{id}` is the first differing
   deliberation ID, or the shipment ID for a set-level added/removed-ID
   difference. Other workspaces apply the same stated rules and
   first-match precedence over the eight `LinkedDeliberationOutcome` values
   that INV-12 names (P-015 Vocabulary and Invariant Summary below). The
   list below is the precedence order; other outcome enumerations in this
   skill are unordered:
   `retained_read_error`, `retained_ambiguous`, `already-archived`,
   `retained_engine_unverified`, `retained_live_status`,
   `retained_shared_reference`, `retained_description_mention`, and otherwise
   a planned `archive`. The planned `archive` is the pre-mutation form of the
   `archived` outcome, which only step 4 below assigns. The disposition set,
   its link kinds, and the shared-reference guard's live-referrer scan all use
   the Step 0(c) matcher, referenced here by name and never restated. The
   shared-reference guard counts, as live referrers when they link the
   deliberation: work items outside
   `closure_scope(S)` that are not truly archived, shipments other than this
   one that are not truly archived, and active stash entries. It never counts
   the deliberation itself, any other deliberation, docs, or archived records.
   A planner `planning_error` carries no planned `archive`. Preserve outcomes
   already settled by the Step 0(c) disposition snapshot (`retained_read_error`,
   `retained_ambiguous`, and `already-archived`); for every remaining snapshot
   deliberation, report outcome `retained_ambiguous` with `reason_code:
   planning_error`, and exclude `already-archived` deliberations from the
   `stranded_linked_deliberation` advisory. Also preserve the snapshot's
   `unresolved_references` set, because the planner returns none on
   `planning_error`.

   **Engine UNVERIFIED means no mutation.** When the Step 0(c)
   engine-semantics verdict is `UNVERIFIED`, every disposition-set
   deliberation that an earlier rule (`retained_read_error`,
   `retained_ambiguous`, `already-archived`) does not settle is
   `retained_engine_unverified`, and nothing is mutated, on every close path.
   This includes a classifier `CASCADE` that Step 0(c) routed to `safe_close`
   with reason `ENGINE_SEMANTICS_UNVERIFIED`.

2. **Disposition baseline.** Before the first archive call, capture the
   disposition baseline from these components:
   * (i) the path-specific set: the safe-close observation-set fingerprints,
     or the CASCADE out-of-manifest descendant fingerprints (disposition step 0);
   * (ii) the disposition snapshot;
   * (iii) a pre-disposition `git status --porcelain -- ".backlogit/"`
     capture, together with a location-and-SHA-256 fingerprint of every
     path under `.backlogit/` other than the exempt paths named below,
     because porcelain status codes alone are not byte identity (a path
     that is already modified or untracked keeps the same entry when its
     bytes change again); enumerate those paths without following any
     symlink, junction, or other reparse point, and before reading any
     path's bytes, at this capture and at every later invariance comparison
     in steps 4 and 6, apply the Step 0(c) containment checks to it (lexical
     and canonical containment in the backlog root, and no symlink,
     junction, or other reparse point);
   * minus the `closure_scope(S)` IDs this run already archived (their
     queue→archive moves are the completed closure, not drift). This
     subtraction applies to the (iii) porcelain capture only: the (iii)
     fingerprint is taken after the closure, so it covers those records'
     post-closure bytes and any later change to them violates the baseline.

   Relative to this baseline, an archive call may change only backlogit
   `ArchiveItem`'s own side effects (verified at `v1.11.0`
   `internal/core/archive.go`):
   * the target deliberation's own queue→archive move;
   * the target's frontmatter keys `status`, `archived_status`, and
     `archived_from`;
   * the gitignored item event log (`.backlogit/logs/`) and backlogit's
     own index, which is exactly `.backlogit/backlogit.db` and its SQLite
     sidecars `.backlogit/backlogit.db-wal`, `.backlogit/backlogit.db-shm`,
     and `.backlogit/backlogit.db-journal` at the backlog storage root
     (no other `.db` file, at the root or below it, is part of the index);
   * lock and hook-queue files.

   `ArchiveItem` also calls `ArchiveLinkedStashEntries`
   (`internal/core/stash.go`), which archives every active stash entry that
   the engine's stash-link index links to the target, rewriting
   `.backlogit/stash.jsonl` and appending to
   `.backlogit/archive/stash.jsonl`. That write is never an allowed side
   effect: the step 3 engine stash-link guard's provenance key check and
   exact relation check together retain every target that carries such a
   link, so the call is a no-op for every archived target. Any
   change to `.backlogit/stash.jsonl` or `.backlogit/archive/stash.jsonl`,
   or to any other path outside these allowed side effects, violates the
   disposition baseline.

   The exempt paths, excluded from the (iii) fingerprint, are only the
   gitignored item event log, the four index paths enumerated above, lock and
   hook-queue files, and this
   run's own closure report path under `.backlogit/reconcile/` (written by
   step 5). Every other `.backlogit/reconcile/` path stays in the baseline.

   A path that fails those checks, or cannot be read, while capturing this
   baseline is never opened or hashed and does not halt: record it in the
   (iii) fingerprint by its workspace-relative path and `reason_code`
   (`path_escape`, `symlink_or_reparse_point`, or `unreadable_file`) instead
   of a SHA-256, as Step 0(c) records a read-error record path. At every later
   comparison in steps 4 and 6, compare such a path by location and
   `reason_code` only, never by following or reading it. Any other path that
   fails those checks at a later comparison, or a recorded path whose
   location or `reason_code` changed, is never read or hashed and fails that
   step's verification (step 4 or step 6).

3. **Archive each planned `archive`, one at a time, in ascending ID order.**
   * Immediately before each call, re-run the SHA-256 check after first
     re-listing the deliberation ID's record
     paths under `.backlogit/queue/` and `.backlogit/archive/` and require
     exactly the disposition snapshot's single path, then reapply the Step
     0(c) containment checks to that record path before reading any of its
     bytes (lexical and canonical containment in the backlog root, and no
     symlink, junction, or other reparse point), then check that record
     path against its disposition-snapshot hash **and** the shared-reference
     guard for that ID, so nothing that changed since
     the snapshot is archived (TOCTOU). Retain the record bytes this re-check
     read; step 4 compares against them.
   * A containment failure at this re-check is never opened, hashed, or
     archived; it halts identically with `HALT — linked-deliberation
     disposition failed {id}`; emit a **P-005** violation once through D6,
     scoped to the disposition archive of {id}; no mutation, and the
     completed close-path closure is never rolled back.
   * A record-path re-list mismatch halts identically with `HALT —
     linked-deliberation disposition failed {id}`; emit a **P-005**
     violation once through D6, scoped to the disposition archive of
     {id}; the completed close-path closure is never rolled back.
   * A hash mismatch halts with `HALT — linked-deliberation disposition failed {id}`;
     emit a **P-005** violation once through D6, scoped to the disposition
     archive of {id}; the completed close-path closure is never rolled back.
   * These retention rules apply in the order listed, and the first that
     applies settles the ID.
   * **Archive-resident target.** A target whose disposition-snapshot record
     path lies under `.backlogit/archive/` rather than
     `.backlogit/queue/` is never archived: record outcome
     `retained_ambiguous` with `reason_code: archive_resident_unarchived` for
     that ID, with no mutation, and the loop continues. For such a record,
     `ArchiveItem` stamps `archived_from` with a path that is not a
     disposition-snapshot queue record path (the engine's canonical queue
     restore path when the record already sits at its archive path), so step
     4 could never verify it.
   * **Engine stash-link guard.** Two checks run in order, and both must
     clear before the target is archived.
     * *Provenance key check.* A target whose retained record bytes declare
       a `custom_fields.source_stash_id` key, with any value, is never
       archived: record outcome `retained_shared_reference` with
       `reason_code: engine_stash_link` and the declared value as its
       referrer for that ID, with no mutation, and the loop continues. In
       `v1.11.0`, `ArchiveLinkedStashEntries` reads the engine's stash-link
       index through `GetStashLinksForItem`; harvest, index rehydration, and
       verified provenance corrections write every production row of that
       index, and each such row names an item whose
       `custom_fields.source_stash_id` equals the linked stash ID. This check
       ignores whether that stash entry still looks active, because the
       index can lag the stash file, so it deliberately retains every
       harvest-created deliberation, even one whose link is no longer
       active, and such a deliberation stays in the
       `stranded_linked_deliberation` advisory.
     * *Exact relation check.* A stash-link row need not carry that key (the
       backlogit `v1.11.0` test `TestArchiveItem_ArchivesLinkedStashEntries`
       links an item that has no `source_stash_id`), so the key check alone
       cannot prove the archive is stash-inert. After the key check clears,
       and immediately before the archive call, run the exact relation that
       `ArchiveLinkedStashEntries` reads through `GetStashLinksForItem`
       (`internal/db/stash.go`) against the engine index
       `.backlogit/backlogit.db`, bound to the target ID:
       `SELECT sl.stash_id FROM stash_links sl JOIN stash_entries se ON se.stash_id = sl.stash_id WHERE sl.item_id = ? AND se.state = 'active'`.
       Open the index **read-only** through the SQLite URI
       `file:.backlogit/backlogit.db?mode=ro` and run only that
       `SELECT`; never write, migrate, sync, rehydrate, or hand-edit the
       index (Data Ownership rule), and never substitute the public stash
       views, which do not expose item links. The index path lies inside the
       backlog storage root that the Step 0(c) containment checks already
       bound; reapply those checks to it before opening it.
       If the query returns any row, the target is never archived: record
       outcome `retained_shared_reference` with `reason_code:
       engine_stash_link` and the returned stash IDs, in ascending order, as
       its referrers for that ID, with no mutation, and the loop continues.
       If the index is missing, fails the containment checks, cannot be
       opened read-only, or lacks the `stash_links` or `stash_entries` table
       or a queried column, so the query cannot run, the target is never
       archived (fail-closed): record outcome `retained_engine_unverified`
       with `reason_code: engine_stash_link_unverifiable` for that ID, with
       no mutation, and the loop continues. Only an error-free query that
       returns zero rows clears this check.
     * Together, the two checks make the `ArchiveLinkedStashEntries` call
       inside `ArchiveItem` a no-op for every archived target. The step 4
       disposition-baseline invariance check, which halts on any change to
       `.backlogit/stash.jsonl` or
       `.backlogit/archive/stash.jsonl`, remains the backstop for a
       link written between the query and the archive call.
   * A new live referrer records `retained_shared_reference: [referrer IDs]`
     for that ID with no mutation, and the loop continues.
   * Otherwise archive that single artifact via `backlogit_archive_item`
     (CLI fallback `backlogit archive {id}`) on the probed surface, the same
     `probe_surface` the Step 0(c) engine-semantics gate recorded. Pass **no
     cascade flag**.

4. **Verify-after-each** (immediately after each archive call). All of the
   following must hold:
   * the queue copy is absent, and the archive copy is present exactly once;
   * the archive copy passes the Step 0(c) containment checks (lexical and
     canonical containment in the backlog root, and no symlink, junction, or
     other reparse point) before any of its bytes are read;
   * the archive copy declares `status: archived`;
   * its `archived_status` equals the deliberation's declared status in the
     disposition snapshot;
   * its `archived_from` provenance is present and well-formed, naming the
     deliberation's disposition-snapshot queue record path by exact string
     match, workspace-relative with `/` separators (step 3 archives only
     queue-resident targets, so that path always exists; missing or
     ill-formed provenance fails verification);
   * its frontmatter, compared **semantically** (parsed YAML, because
     `ArchiveItem` re-serializes it), equals the frontmatter of the record
     bytes retained at the step 3 re-check, except for the three engine keys
     `status`, `archived_status`, and `archived_from`;
   * its Markdown body after the closing frontmatter delimiter is
     **byte-exact** against those retained record bytes;
   * disposition-baseline invariance holds: nothing changed relative to the
     step 2 disposition baseline except that call's allowed `ArchiveItem`
     side effects and the side effects of earlier disposition archives in
     this step that already passed verification.

   Any failure: `HALT — linked-deliberation disposition failed {id}`; follow
   the D6 sequence of safe-close step 6, scoped to the disposition archive of
   {id} (capture evidence → HALT → emit P-005 once → request EXPLICIT
   operator approval → REVALIDATE → execute ONLY the approved rollback). There is no retry and
   no rollback of the completed closure: the close path's gate already
   passed, so the closure is complete and only the disposition is not. Do not
   advance to post-mode, and never commit a torn disposition (a partially
   verified archive state; the self-hosting repository records this lesson in
   `docs/compound/2026-08-15-torn-archive-log-entry-without-file-mutation-must-not-be-committed.md`).

5. **Report.** Add to the closure report (safe-close step 9 or cascade-close
   step 5):
   * `linked_deliberation_disposition: [{id, link_kinds, linking_members, outcome, reason_code, path, referrers, pre_sha256, post_sha256, archived_status}]`,
     one entry per disposition-set deliberation, where `outcome` is a
     `LinkedDeliberationOutcome` value (`archived`, `already-archived`,
     `retained_read_error`, `retained_ambiguous`,
     `retained_engine_unverified`, `retained_live_status`,
     `retained_shared_reference`, or `retained_description_mention`),
     `reason_code` is always present and is copied verbatim from the planner
     for every planner outcome other than the planned `archive` (never
     re-derived; an unknown code is carried as-is), except that a verified
     `archived` outcome carries `reason_code: archived` (the outcome-value
     default, replacing the planner's pre-mutation `archive`), a planned
     `archive` that the step 3 re-check settles as
     `retained_shared_reference` carries `reason_code:
     retained_shared_reference`, one that the step 3 engine stash-link guard
     settles carries `reason_code: engine_stash_link` (or, when its exact
     relation check cannot run, outcome `retained_engine_unverified` with
     `reason_code: engine_stash_link_unverifiable`), one that step 3
     settles as an archive-resident target carries `reason_code:
     archive_resident_unarchived`, an outcome preserved from the Step 0(c)
     disposition snapshot keeps that snapshot's own `reason_code` (for
     example a read error's `path_escape`), and an outcome synthesized for a
     `planning_error` (only the remaining snapshot deliberations) carries
     `reason_code: planning_error`, `path` is present
     for `retained_read_error`, `referrers` lists the live referrer IDs of a
     `retained_shared_reference` (for `engine_stash_link`, the declared
     `source_stash_id` value from the provenance key check, or the stash IDs
     the exact relation check returned), and `post_sha256` and `archived_status` are
     present for an `archived` outcome. Report keys map to planner/code fields
     as follows: `id` → `deliberation_id`, `linking_members` →
     `linking_member_ids`, `referrers` → `referrer_ids`, and `pre_sha256`
     records one pre-archive SHA-256 per record. When the planner reports
     `planning_error`, record the `planning_error`. Preserve outcomes already
     settled by the Step 0(c) disposition snapshot (`retained_read_error`,
     `retained_ambiguous`, and `already-archived`); for every remaining snapshot
     deliberation, report outcome `retained_ambiguous` with `reason_code:
     planning_error`, and exclude `already-archived` deliberations from the
     `stranded_linked_deliberation` advisory, and report the snapshot's
     `unresolved_references` set. An empty disposition set records
     `linked_deliberation_disposition: []`;
   * `unresolved_references` (`{id, reason_code}`), plus every planner read
     failure (`{path, reason_code}`);
   * the durable advisories, carried into the closure summary:
     `ENGINE_SEMANTICS_UNVERIFIED` with the Step 0(c) reason whenever the
     engine-semantics verdict is `UNVERIFIED`, and
     `ENGINE_LINE_UNVERIFIED_ADVISORY` naming the probed version whenever its
     minor line is not a P-015 verified engine-semantics line;
   * a `stranded_linked_deliberation` advisory listing every `retained_*`
     outcome, so a deliberation that stays live across closures stays visible
     to the operator.

6. **Gate decision.**
   * Every disposition-set deliberation has exactly one outcome, every
     `archived` outcome passed step 4, and a final disposition-baseline
     invariance check, run after the step 5 report write, passes (nothing
     changed relative to the step 2 disposition baseline except the allowed
     side effects of the verified disposition archives and this run's own
     closure report path) → `recommendation: DISPOSITION_COMPLETE`.
     Proceed to post-mode. Retained outcomes are reported, never halt, and
     never block this gate. `linked_deliberation_disposition: []` is valid
     only when the Step 0(c) disposition snapshot is empty; a `planning_error`
     over a non-empty snapshot preserves snapshot-settled outcomes and reports
     only the remaining deliberations as `retained_ambiguous` with
     `reason_code: planning_error`.
   * A failure of that final invariance check halts with `HALT —
     linked-deliberation disposition failed {id}`, where `{id}` is the
     shipment ID; follow the D6 sequence of safe-close step 6, scoped to the
     diverging paths plus this run's disposition archives (when the run
     archived nothing, the scope is the diverging paths alone), and emit
     **P-005** once. Do not proceed to post-mode.
   * Any halt above → `recommendation: HALT — linked-deliberation disposition failed {id}`.
     Do not proceed to post-mode or to any commit step. On a disposition HALT,
     keep the shipment lock held through the D6 sequence (approval → REVALIDATE
     → approved rollback) and release it only after D6 finishes — either the
     approved rollback is complete or the operator declines — following
     pre-mode step 7's rule to release the lock on pre-mode HALT while retaining
     it after `PROCEED` from Ship Step 6 until post-mode completes.

**Action risk (ProposedAction / ActionRisk).** `backlogit_archive_item` on a
linked deliberation is ActionRisk **medium**: it is reversible with
`backlogit restore`, and it is gated by the shared-reference guard, the
pre-archive hash re-check, and verify-after-each. The engine-version probe is
ActionRisk **none** (read-only). This step adds no other mutation.

**Sanctioned successor.** This step is the sanctioned successor to the
operator-approved 190-S closure deviation, in which the operator approved a
one-off standalone archive of a linked deliberation that the cascade had left
live. That archive now happens only here, under the guards above; no closure
repeats it as an ad hoc deviation.

### Mixed-Role Detection Mode (`mode: detect-mixed-role`, operator-invoked, READ-ONLY)

No lock is acquired for this mode — no backlog/shipment artifact is ever
mutated. (The mode's own diagnostic report, audit-log entry, and telemetry
event — steps 6, 8 below — ARE writes, but they are additive-only writes to
non-backlog-state locations, never applied to a queue/archive item, so no
`file-lock` is required.) This mode is composed entirely from EXISTING
read-only backlogit reads; it introduces NO new gate/CLI code (single
template family: this SKILL's prose only).

1. **Enumerate shipments**: call `backlogit_list_shipments`. If `shipment_id`
   was given, narrow to that single shipment; otherwise scan every shipment
   returned.

2. **Load each candidate shipment record** via `backlogit_get_shipment`.
   Skip (no report entry) any shipment whose record status is
   `active`, `shipped`, `abandoned`, or archived — those are the
   normal in-progress/closed lifecycle **shipment-record** states and are out
   of scope for this check (mirrors the `record-consistent` scope boundary
   above). Note: `shipped`/`abandoned` are the shipment record's own terminal
   statuses (`ShipmentStatus` enum), distinct from `done` which is
   a **task**-artifact status — a live shipment record is never itself
   `done`. If a candidate's persisted status is not a valid
   backlogit 1.8.0 shipment lifecycle value (e.g. a legacy `blocked` value),
   classify it `malformed-legacy`, add it to the report, and continue to the
   next candidate — never fabricate a transition. Any other unrecognized
   persisted value is likewise `malformed-legacy` rather than silently
   skipped or silently matched to the queued branch below — every possible
   persisted value maps to exactly one of: skip (`active`/
   `shipped`/`abandoned`/archived), scan (`queued`, step 3), or
   `malformed-legacy` (anything else).

3. **Filter to task-artifact manifest items**: for each remaining
   `queued` candidate, read its manifest `items` list and each
   item's frontmatter (`status`, `artifact_type`) via per-item reads
   (`backlogit_get_item` per task id). Exclude any non-task entry (e.g. a covering
   feature id in a fallback-assembled manifest) before classifying — the same
   task-artifact filter the Ship agent's intake early-warning applies to
   `custom_fields.items` (`templates/agents/_ship.agent.md.tmpl`).

4. **Classify each task-artifact manifest item** against the per-task ALLOWED
   ROLE table and the per-item ANOMALY table in the "Mixed-Role Detection
   Classification" section above. Locate each id in
   `.backlogit/queue/` and `.backlogit/archive/` to
   determine role/anomaly; for an archive hit, read `status`, `archived_status`,
   and `archived_from` to distinguish the two valid archived-completed
   representations from `malformed-provenance` / `any-other-archived-status`.

5. **Determine the outcome for each candidate**:
   * If any per-item anomaly was found → `REPORTED`, naming the shipment id,
     the anomalous task id(s), and the specific anomaly for each.
   * Else if the mixed-role signature is present (at least one `live-active`
     or `archived-completed(done)` role task, all tasks otherwise role-clean)
     → `REPORTED`, naming the shipment id, record status, and each task's role.
   * Else (all tasks role-clean and no mixed-role signature — e.g. a
     genuinely fresh `queued` shipment with all-`live-queued`
     tasks) → `DETECTED` with no anomaly/signature to report for that
     candidate.
   * If backlogit is unreachable at any point in steps 1–4 → `DEGRADED`;
     report the degraded condition for the affected candidate(s) and HALT the
     scan. Do not guess or proceed on partial data.

6. **Emit the report-only diagnostic** at
   `.backlogit/reconcile/{shipment_id-or-"all"}-detect-mixed-role-{timestamp}.md`,
   listing per candidate: the shipment id, its record status, the per-task
   role classification, any per-item anomaly, and the outcome (`DETECTED` /
   `REPORTED` / `DEGRADED`).

7. **Emit Operator-Remediation Guidance** inline with any `REPORTED` or
   `DEGRADED` entry — see "Operator-Remediation Guidance" below. This mode
   performs NO mutation of any kind in response to what it finds.

8. **Write the audit entry and emit telemetry** for every candidate's outcome
   — see "Mixed-Role Detection Audit + Telemetry" below.

#### Operator-Remediation Guidance

When this mode reports a mixed-role signature or a per-item anomaly, include
this guidance verbatim (adapted with the specific ids/anomaly found) in the
report:

> autoharness performs **NO auto-repair** of this inconsistency. A
> record-only forward re-claim (`queued` → `active` on
> the shipment record alone) is **UNSUPPORTED** by backlogit 1.8.0: evidence
> (read-only inspection of `C:\Source\GitHub\backlogit`, NOT mutated) —
> `ClaimShipment` (`internal/core/shipment_lifecycle.go`) is **manifest-wide**
> activation (it moves the shipment `queued`→`active`
> AND THEN activates every still-`queued` manifest member,
> cascading parent-feature status, with all-or-nothing rollback on any
> mid-flight failure) and is **STRICTLY SINGLE-SHOT**
> (`isValidShipmentTransition` in `internal/core/shipment.go` permits ONLY
> `queued`→`active` and
> `active`→`{shipped,abandoned}`; a re-claim on an already-`active`
> shipment returns `ErrShipmentConflict`). There is **NO**
> `active`→`queued` transition and **NO** `blocked`
> shipment status in 1.8.0 — never expect or fabricate either. The SUPPORTED
> manual remediation path is entirely through backlogit's own sanctioned
> lifecycle transitions: the operator inspects the shipment and its manifest
> tasks directly (`backlogit get <id>` / `backlogit shipment get <id>`) and
> decides, case by case, whether the tasks are legitimately progressing (in
> which case the operator may let the shipment proceed to closure normally
> once all tasks complete, using this skill's own `mode: pre` →
> `mode: safe-close` → `mode: post` sequence) or whether the state reflects a
> genuinely torn/partial session (in which case the operator investigates and
> resolves the affected tasks manually before any closure attempt). This
> guidance is descriptive only; this skill never performs any of these steps
> itself.

#### Mixed-Role Detection Audit + Telemetry

Mirrors the `pipeline-topology` force-audit + telemetry pattern (this
installed skill markdown is the authoritative description of that pattern;
`src/autoharness/cli.py` `_audit_pipeline_topology_force` /
`_emit_pipeline_topology_telemetry` is an upstream reference implementation
in the self-hosting repository that may be absent in other workspaces),
adapted for a detection-only outcome — there is NO
repair/mutation/confirm/post-condition field, because nothing is ever
mutated.

1. **Audit log**: for EVERY candidate outcome (`DETECTED` / `REPORTED` /
   `DEGRADED`), append one structured JSON line to
   `.autoharness/gates/shipment-reconcile-detection-audit.log` (creating the
   `.autoharness/gates/` directory if needed), containing: `timestamp`
   (UTC ISO-8601), `actor` (the invoking operator/session identity, e.g. from
   `USERNAME`/`USER`), `shipment_id`, `record_status`, `outcome` (`DETECTED` \|
   `REPORTED` \| `DEGRADED`), `per_task_roles` (an array of
   `{task_id, role, anomaly}` — `anomaly` is `null` when the task is
   role-clean), `remediation_guidance_emitted` (boolean), and `report_path`
   (the reconciliation report path from step 6). NO `repair`, `mutation`,
   `confirm`, or `post_condition` field is ever present — those concepts do
   not apply to this read-only mode.
2. **Telemetry event**: emit one telemetry event per candidate outcome,
   mirroring the pipeline-topology `ToolTelemetryEvent` shape
   (`schemas/tool-telemetry-event.schema.json`): `tool_surface: "builtin"`
   (the schema's `tool_surface` enum is `mcp` \| `cli` \| `shell` \| `builtin`
   \| `api` \| `unknown` — `"skill"` is NOT a valid value; `shipment-reconcile`
   is an agent-invoked prose skill with no separate CLI/MCP/API backend of its
   own, so `builtin` is the correct surface),
   `tool_name: "shipment-reconcile"`, `operation: "detect-mixed-role"`,
   `status` mapped from the outcome (`DETECTED` → `success`, `REPORTED` →
   `blocked`, `DEGRADED` → `failed`), `sensitivity: "internal"`,
   `redaction_applied` and `metric_sources`/`metric_quality` set per the
   schema's structural requirements, `shipment_id`/`backlog_item_id` set to
   the candidate shipment id, and `artifact_refs` including the audit log
   path and the report path. When invoked inside an active Ship task epoch
   with a live `context_ref` (from `autoharness telemetry begin`), emit via
   `autoharness telemetry event --context-ref <ref> --from-json <payload>`;
   when invoked standalone (e.g. an ad-hoc operator audit outside any Ship
   task loop, the most common case for this mode) there is no `context_ref`
   to attach to, so telemetry emission for that invocation is skipped and the
   audit log entry (step 1) remains the durable record — this mirrors
   telemetry's existing fail-open, observational contract (an absent or
   unavailable telemetry surface is a no-op and NEVER blocks detection,
   reporting, or the HALT decision).

### Lock-Conflict Scenario

If pre-mode cannot acquire the lock because another process holds it:

1. Retry once after 30 seconds.
2. If retry also fails, count as a session stall and prompt the operator:
   `Shipment lock conflict on {shipment_id}. Another process holds the lock.`
3. Do NOT proceed without the lock. Do NOT call `backlogit_ship_shipment`.



## Deterministic Safe-Close Scenario Matrix

* **114-S -> 115-S -> 116-S serial-close success chain**: 114-S safely closes while 115-S/116-S remain protected; once 114-S carries verified `archived_status: shipped` (or legacy `done`), 115-S may exclude 114-S-owned siblings from its protected set; once 115-S carries the same verified provenance, 116-S may exclude 115-S-owned siblings.
* **Negative — archive-while-active**: if the shipment record is archived while still `status: active`, producing `archived_status: active`, halt with `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`.
* **Negative — non-shipped-live-before-archive**: if the shipment record cannot be re-read and verified as live `status: shipped` before the archive step, halt with `RECONCILE_FAIL_SHIPMENT_RECORD_LIVE_STATUS`.
* **Negative — missing archive**: if the archive file for the shipment record is missing after `backlogit archive <shipment_id>`, halt with `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`.
* **Negative — archived abandoned**: if the shipment record archives with `archived_status: abandoned`, halt with `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`.
* **Negative — missing provenance**: if the shipment record is archived but lacks `archived_status: shipped|done`, keep predecessor siblings protected and halt with `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`.

## Quality Criteria

* `mode: pre` runs before closing a shipment in Ship Step 6
* `mode: pre` with `expected_status: queued` (or `active` for already-claimed shipments) runs at Ship Step 0.5 intake
* `mode: safe-close` runs **in place of** the cascade `backlogit_ship_shipment` call and archives only the manifest item IDs (one artifact at a time) plus the shipment record itself — **except** when the `--classify-only` route exits 0 (`CASCADE`), in which case only the mutating `autoharness shipment cascade-close` command runs, after per-closure operator approval, and safe-close steps 1–10 are skipped entirely
* Close-path selection is made **only** from the machine-checkable `--classify-only` result (the `selected_close_path` field of the command's JSON output, recorded in the evidence record as `pre_close.close_path_selection`; Step 0(c)), never inferred from prose or manifest shape alone; any classifier error, ambiguity, or unresolved precondition HALTs under the Step 0(c) routing table, and only exit 3 routes to safe-close
* The P-015 INV-6 "fully covered" check (evaluated by the Step 0(c) `--classify-only` run) walks the qualifying feature's **full descendant tree, at every depth** (via a full `parent_id` graph, not a single-level scan of direct children only) — a manifest such as `[feature, task]` where that task has an out-of-manifest subtask must select `safe_close` (exit 3), never wrongly qualify for `CASCADE` (155-S, PR #407 review, thread PRRT_kwDORzpWpM6b2MJv)
* The Cascade Close Sub-Procedure independently verifies `returned_ids` is empty, `archived_ids` against the two-set `allowed_ids` / `required_ids` gate (step 3 applies the flat `allowed_ids` / `required_ids` sets defined in "Cascade Close Postconditions: Flat"; `archived_ids - allowed_ids` non-empty halts with `HALT — cascade archived unexpected artifact {id}`, `required_ids - archived_ids` non-empty halts with `HALT — cascade did not archive required artifact {id}`, evaluated as two independent, never-merged conditions — a truly pre-archived **non-shipment, non-feature** allowed member (i.e. a manifest task item) has no transition to report and is correctly, expectedly absent from `archived_ids`, never a mismatch, but this tolerance never extends to the shipment record itself, which remains unconditionally required regardless of its own pre-close declared status, nor does it extend to a qualifying feature member itself, which is likewise unconditionally required regardless of its own pre-close declared status — since Backlogit's own `ShipShipment` forces every explicit qualifying feature member through `status: done` before archive-candidate collection ever runs, so it is never still `archived` by that point either), and no `parent_id` was cleared — a mismatch on any of these halts fail-closed with a P-005 violation even though the cascade path was itself permitted
* Safe-close computes the protected set (parent feature + unshipped siblings) from expected IDs and proves it is fully present in queue at a baseline gate before archiving anything
* Safe-close verifies the protected set survives after every single-item archival (verify-after-each invariant), with no pre-archived exemption for protected-set members
* Safe-close moves the shipment record to live `shipped`, verifies it, explicitly archives only that record, and verifies `archived_status: shipped` before closure completes
* Safe-close archives the shipment record as its own single artifact and never via the cascade op
* Safe-close reverts on cascade detection (`git restore -- <path>` per touched path, or `git revert`) and halts with a P-005 violation; it never auto-prunes the manifest
* `mode: post` runs after the safe-close archive sequence in Ship Step 6
* All five item classifications are represented in the schema
* Pre-mode adds a shipment-record-status classification (`record-consistent` /
  `record-queued-with-active-work` / `record-blocked-with-active-work` /
  `record-blocked-with-done-work`) comparing the record's own status against its
  manifest **task-artifact** items' statuses (filtered by `artifact_type` to
  exclude any non-task manifest entry, e.g. a covering feature id, before
  aggregating), with the blocked+active-over-blocked+done precedence
  rule stated explicitly; the four cases are mutually exclusive, computed with
  no new scan (reuses data already read in steps 2–3); any non-`record-consistent`
  classification HALTs the pre-mode recommendation, naming the shipment id,
  record status, and conflicting task ids — detect-and-report only, no auto-repair
* Lock is acquired before pre-mode and released after post-mode (or on any halt)
* Report-and-halt in pre/post mode; safe-close mutation is strictly manifest-scoped with no auto-prune
* `mode: detect-mixed-role` is operator-invoked and strictly READ-ONLY: it
  requires NO `file-lock` acquisition, NEVER mutates any shipment record or
  task, and NEVER calls `backlogit_claim_shipment` or any other status-write
  operation
* `mode: detect-mixed-role` classifies each task-artifact manifest item into
  exactly one per-task ROLE (`live-queued` / `live-active` /
  `archived-completed(done)` — either the terminal-relocation `status: done`
  representation or the explicit-archival `status: archived` +
  `archived_status: done` + valid `archived_from` representation) or flags a
  per-item ANOMALY (`duplicate` / `conflicting` / `missing` /
  `malformed-provenance` / `any-other-archived-status` / `orphan` /
  `out-of-role` / `torn-partial`); role classification is used ONLY to
  DESCRIBE the inconsistency in the report, NEVER to gate a mutation
  (013-DL Addendum G)
* `mode: detect-mixed-role` produces exactly one outcome per candidate
  shipment — `DETECTED` / `REPORTED` / `DEGRADED` — with NO
  `succeeded`/`repaired`/`refused`/two-active outcome, and writes a
  structured audit entry (`.autoharness/gates/shipment-reconcile-detection-audit.log`)
  plus a best-effort telemetry event for every outcome, mirroring the
  `pipeline-topology` force-audit + telemetry pattern
* `mode: detect-mixed-role` includes inline Operator-Remediation Guidance on
  any `REPORTED`/`DEGRADED` outcome, explicitly stating autoharness performs
  NO auto-repair and that a record-only forward re-claim is UNSUPPORTED by
  backlogit 1.8.0 (`ClaimShipment` is manifest-wide + all-or-nothing +
  STRICTLY SINGLE-SHOT; NO `active`→`queued` edge; NO
  `blocked` shipment status), with the read-only source evidence, and never
  fabricates a `blocked`→`queued` or
  `active`→`queued` transition
* `mode: detect-mixed-role` DEGRADED (backlogit unreachable) reports the
  degraded condition and HALTs — it never guesses or acts on partial data

## Related Artifacts

* `.github/skills/file-lock/SKILL.md` — lock acquisition/release primitives
* `.github/agents/_ship.agent.md` — integration points (Step 0.5, Step 6 safe-close)
* `.github/agents/_stage.agent.md` — scope guard (Step 5.5)
* `.github/policies/workflow-policies.md` — P-007 archive integrity policy; P-015 single-artifact closure (cascade prohibition) plus the flat-manifest `CASCADE` exception
* `src/autoharness/gates/shipment_closure.py` — optional upstream reference: this self-hosting repository's own `classify_shipment_close_path` implementation, reused by the Step 0(c) classification specification when present. This installed skill markdown is authoritative; the path is an upstream reference that may be absent in other workspaces.

## Model Routing

This skill operates at **Tier 2 (Standard)** — file scanning and frontmatter
comparison do not require frontier-level reasoning.

Generated by autoharness | Template: shipment-reconcile/SKILL.md.tmpl
