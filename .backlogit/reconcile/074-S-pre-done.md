# Shipment reconcile (pre mode): 074-S

* shipment_id: 074-S
* mode: pre
* expected_status: done
* merge_commit_sha: 966f6a977ef6590e1d7a9155c3f64d0ee8861061 (PR #223)
* lock: held on `.backlogit/queue/074-S.md` via the `file-lock` skill
  (`acquire_lock.ps1 .backlogit/queue/074-S.md -WorkspaceRoot C:\Source\GitHub\intercom`);
  token `af7e8827eb65f67c4c8fc57f93d802b5f5812db1d20743530fbd8dc7a0316b95`.
  The default invocation without `-WorkspaceRoot` failed the script's widening guard on this
  workspace ("git-derived root does not match this script's own installed location"); the
  explicit root is the documented remedy. The lock is held across mode: pre and classify-only,
  released cleanly for the per-closure approval pause, and re-acquired before the approved
  mutating run and mode: post.
* shipment record status: `active` (record-consistent: every manifest member `done`)

## Manifest checks

| id | location | artifact_type | declared status | classification |
|---|---|---|---|---|
| 081-F | archive | feature | done | matched (moved `active -> done` at Step 6.1.a1 S4; the backlog tool archived it) |
| 081.001-T | archive | task | done | matched |
| 081.002-T | archive | task | done | matched |
| 081.003-T | archive | task | done | matched |
| 081.004-T | archive | task | done | matched |
| 081.005-T | archive | task | done | matched |

## Orphan scan

* Queue files declaring `shipment_id: 074-S` or `parent_id: 081-F` outside the manifest: none.
* Live descendants of 081-F at every depth: none.

recommendation: PROCEED
