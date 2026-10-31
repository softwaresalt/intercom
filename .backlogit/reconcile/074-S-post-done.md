# Shipment reconcile (post mode): 074-S

* shipment_id: 074-S
* mode: post
* merge_commit_sha: 966f6a977ef6590e1d7a9155c3f64d0ee8861061
* lock: `file-lock` token acquired before the mutating cascade-close and held through this post-mode check, then released

## Post-close checks

| check | result |
|---|---|
| shipment record | archived, `archived_status: shipped`; queue file removed |
| 081-F (feature) | archived, `parent_id` preserved |
| 081.001-T to 081.005-T | archived, `parent_id: 081-F` preserved |
| archive deletions (P-007) | none (`git status --short -- .backlogit/archive/` shows one rename and in-place modifications only) |
| missing or unexpected archived ids | none (`archived_ids` = 074-S, 081-F, 081.001-T to 081.005-T; `returned_ids` empty) |
| linked-deliberation disposition | `DISPOSITION_COMPLETE` (empty dispositions, no drift) |
| close evidence | `docs/closure/evidence/074-S-081-F-close-evidence.json`, `phase: post_close`, `postcondition_verdict: pass`, exit 0 |

recommendation: PROCEED
