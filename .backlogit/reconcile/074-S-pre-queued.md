# Shipment reconcile (pre mode): 074-S

* shipment_id: 074-S
* mode: pre
* expected_status: queued
* lock: not held (intake early-warning check)
* shipment record status: `queued` (record-consistent: all manifest members `queued`)

## Manifest checks

| id | location | artifact_type | declared status | classification |
|---|---|---|---|---|
| 081-F | queue | feature | queued | matched |
| 081.001-T | queue | task | queued | matched |
| 081.002-T | queue | task | queued | matched |
| 081.003-T | queue | task | queued | matched |
| 081.004-T | queue | task | queued | matched |
| 081.005-T | queue | task | queued | matched |

## Orphan scan

* Queue files declaring `shipment_id: 074-S` outside the manifest: none.

recommendation: PROCEED
