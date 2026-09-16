# Probe 14 — end-to-end dependency lifecycle over a shipment-blocks edge
# Mirrors the live 017-S -> 021-S topology: successor depends_on predecessor --type blocks.
# Workspace-contained. Emits a deterministic transcript to stdout.
$ErrorActionPreference = 'Continue'
$ws = Join-Path $PSScriptRoot 'p14'
function Say($m) { Write-Output "### $m" }
function B { backlogit --cwd $ws @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }

Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $ws | Out-Null
backlogit init $ws 2>&1 | Where-Object { $_ -notmatch '^time=' }
# Pre-create the archive directory. A freshly-init'd workspace has none until the
# first archival, and dag-readiness reports status=degraded (NOT a real verdict)
# when it is absent. Without this the STEP 4/5 suppression checks would be vacuous.
New-Item -ItemType Directory -Force -Path (Join-Path $ws '.backlog\archive') | Out-Null

Say "STEP 1: build two independent feature/task trees"
B add --type feature --title "Predecessor prerequisite feature"      # 001-F
B add --type task --title "Predecessor sole task" --parent 001-F     # 001.001-T
B add --type feature --title "Successor policy-gap feature"          # 002-F
B add --type task --title "Successor sole task" --parent 002-F       # 002.001-T

Say "STEP 2: create two TASK-ONLY shipments"
B shipment create --title "Predecessor shipment" --items 001.001-T   # 001-S
B shipment create --title "Successor shipment"   --items 002.001-T   # 002-S

Say "STEP 3: write the blocks edge  successor(002-S) depends_on predecessor(001-S)"
B dep add 002-S 001-S --type blocks
Say "edge readback (002-S):"
B dep list 002-S

Say "STEP 4: BASELINE readiness - predecessor QUEUED, successor MUST be suppressed"
B sync | Out-Null
autoharness gate dag-readiness --workspace $ws --json 2>&1
Write-Output "GATE_EXIT=$LASTEXITCODE"

Say "STEP 5: claim predecessor (queued -> active); successor MUST STILL be suppressed"
B shipment claim 001-S
B sync | Out-Null
autoharness gate dag-readiness --workspace $ws --json 2>&1
Write-Output "GATE_EXIT=$LASTEXITCODE"

Say "STEP 6: complete the predecessor's sole task, then close predecessor via ShipShipment"
B move 001.001-T --status done
B shipment ship 001-S --sha 0000000000000000000000000000000000000000 --message "merge: probe14 predecessor" --author "stage-probe@local"
Write-Output "SHIP_EXIT=$LASTEXITCODE"

Say "STEP 7: predecessor archived-record provenance (status + archived_status)"
$arch = Join-Path $ws '.backlog\archive\001-S.md'
if (Test-Path $arch) { Write-Output "PATH=.backlog/archive/001-S.md"; Get-Content $arch -Raw } else { Write-Output "MISSING ARCHIVED PREDECESSOR RECORD" }

Say "STEP 8: official sync (index rehydration)"
B sync

Say "STEP 9: POST-CLOSE readiness - successor MUST now be eligible"
autoharness gate dag-readiness --workspace $ws --json 2>&1
Write-Output "GATE_EXIT=$LASTEXITCODE"

Say "STEP 10: successor + its edge after predecessor archival"
B dep list 002-S
B shipment get 002-S
