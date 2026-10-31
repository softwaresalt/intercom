# Shipment reconcile (pre mode): 075-S

* shipment_id: 075-S
* mode: pre
* expected_status: queued
* lock: not held (intake early-warning check)
* shipment record status: `queued` (record-consistent: all manifest members `queued`)

## Manifest checks

| id | location | artifact_type | declared status | classification |
|---|---|---|---|---|
| 082-F | queue | feature | queued | matched |
| 082.001-T | queue | task | queued | matched |
| 082.002-T | queue | task | queued | matched |
| 082.003-T | queue | task | queued | matched |
| 082.004-T | queue | task | queued | matched |
| 082.005-T | queue | task | queued | matched |

## Orphan scan

* Queue files declaring `shipment_id: 075-S` outside the manifest: none.

recommendation: PROCEED
