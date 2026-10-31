# Shipment reconcile (post mode): 075-S

* shipment_id: 075-S
* mode: post
* merge_commit_sha: a2898942a12c70fe23be050d698ae4622153d97f (PR #225)
* close_path: cascade (evidence `docs/closure/evidence/075-S-082-F-close-evidence.json`, phase `post_close`)
* lock: held on `.backlogit/queue/075-S.md` via the `file-lock` skill; token
  `07fad7b284a5962e48a0f12c071c79ed342a321457a5dce1d7fe21b8ee561ae6`, released at the end of this mode
* shipment record: `.backlogit/archive/075-S.md` with `archived_status: shipped`; queue file removed
* linked_deliberation_disposition: DISPOSITION_COMPLETE (no dispositions)

## Archive verification

| id | location | result |
|---|---|---|
| 075-S | archive | verified (archived_status shipped) |
| 082-F | archive | verified |
| 082.001-T | archive | verified |
| 082.002-T | archive | verified |
| 082.003-T | archive | verified |
| 082.004-T | archive | verified |
| 082.005-T | archive | verified |

* archived_ids: 075-S, 082-F, 082.001-T through 082.005-T
* returned_ids: none
* parent_id_preserved: true
* P-007: `git status -- .backlogit/archive/` shows no working-tree deletions

recommendation: PROCEED
