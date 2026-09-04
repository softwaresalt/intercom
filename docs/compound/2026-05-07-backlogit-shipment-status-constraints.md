---
title: "backlogit shipment status constraints (generic move refuses direct shipment transitions)"
date: 2026-05-07
category: build-errors
tags: [backlogit, shipment, cli, closure]
---

# backlogit shipment status constraints

## Symptom

Running the generic status-mutation command against a shipment artifact
fails, even though the same command works for every other artifact type
(feature/task/subtask):

```text
$ backlogit move 002-S --status shipped
move shipment 002-S to shipped via generic path: backlogit: shipment must be
shipped via ShipShipment, not a direct status update
exit code: 9
```

## Root cause

Backlogit (confirmed on the version installed for `intercom-go`, and
consistent across at least 1.8.0–1.10.x) treats shipment closure as a
structurally different operation from a plain status transition. A
shipment's `active → shipped` transition is *only* reachable through the
dedicated `ShipShipment` engine path (CLI: `backlogit shipment ship <id>
--sha <merge_sha> --message <msg> --author <author>`), because closing a
shipment has release-scope side effects (archiving covering-feature/task/
subtask descendants, stamping `commit_sha` provenance) that a bare status
flip cannot perform safely. The generic `move` command intentionally
rejects the shipment artifact type for this transition rather than silently
performing a partial, unsafe mutation.

Backlogit shipment status values observed in this workspace: `queued`,
`active`, `shipped`, `abandoned`. There is **no** `blocked`
status for shipments — do not assume shipment lifecycle mirrors task
lifecycle (`queued → active → done`/`blocked`); shipments have their own,
narrower state machine.

## Resolution

* Never attempt `backlogit move <shipment_id> --status shipped` (or any
  direct status mutation) on a shipment artifact — it will always fail with
  exit code 9 and the message above.
* Use `backlogit shipment ship <shipment_id> --sha <merge_commit_sha>
  --message <merge_commit_message> --author <merge_commit_author>` instead.
  This is the **only** supported closure path for a shipment once its
  manifest qualifies for cascade close (see the P-015 verified
  fully-covered-root exception in the `shipment-reconcile` skill).
* If the shipment's manifest does *not* qualify for the cascade exception
  (a partial-feature shipment with unshipped siblings), use the
  `shipment-reconcile` skill's Safe-Close Mode instead, which never calls
  `backlogit shipment ship` and instead archives only the manifest's
  explicit item IDs one at a time before closing the shipment record via
  the reconcile skill's own verified sequence.
* A manifest item may be **relocated** to `.backlogit/archive/` (by file
  location) while still declaring `status: done` rather than
  `status: archived` — this is a distinct, valid intermediate state (e.g.
  produced by a per-task completion flow that archives the task but not via
  the full shipment-close sequence). Treat "location in archive/" and
  "declared `status: archived`" as two independent facts: only the
  declared `status` field is authoritative for pre-close snapshot/gate
  computations in the reconcile skill's cascade sub-procedure. `backlogit
  doctor` reports this state as consistent ("No issues found") — it is not
  an integrity error, just an earlier-lifecycle-stage artifact.

## Verification

`backlogit doctor` after closure reports "No issues found"; `backlogit get
<shipment_id> --json` reports `status: archived`; the raw archive
frontmatter reports `archived_status: shipped` and `commit: <merge_sha>`.

## Compounding value

Any Ship-agent template or skill assuming a generic `move --status shipped`
path exists for shipment closure will hard-fail with exit code 9 on first
use against a real backlogit installation. Always route shipment closure
through `backlogit shipment ship` (cascade) or the `shipment-reconcile`
skill's Safe-Close Mode (manual per-item archive) — never through the
generic `move` command.
