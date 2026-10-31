# Shipment reconcile (pre mode): 075-S

* shipment_id: 075-S
* mode: pre
* expected_status: done
* merge_commit_sha: a2898942a12c70fe23be050d698ae4622153d97f (PR #225)
* lock: held on `.backlogit/queue/075-S.md` via the `file-lock` skill
  (`acquire_lock.ps1 .backlogit/queue/075-S.md -WorkspaceRoot C:\Source\GitHub\intercom`);
  token `abab7dfb3b2daaf37a54ccda31c3f3ce5cff38717d17e26d06d4464837465a9d`.
  The lock is held across mode: pre and classify-only, released cleanly for the
  per-closure approval pause, and re-acquired before the approved mutating run
  and mode: post.
* shipment record status: `active` (record-consistent: every manifest member `done`)

## Manifest checks

| id | location | artifact_type | declared status | classification |
|---|---|---|---|---|
| 082-F | archive | feature | done | matched (moved `active -> done` at Step 6.1.a1 S4; the backlog tool archived it) |
| 082.001-T | archive | task | done | matched |
| 082.002-T | archive | task | done | matched |
| 082.003-T | archive | task | done | matched |
| 082.004-T | archive | task | done | matched |
| 082.005-T | archive | task | done | matched |

## Orphan scan

* Queue files declaring `shipment_id: 075-S` or `parent_id: 082-F` outside the manifest: none.
* Live descendants of 082-F at every depth: none.

recommendation: PROCEED
