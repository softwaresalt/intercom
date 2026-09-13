# Probe 18 - LIVE-CONFIG-SEEDED prerequisite topology + inbound-edge byte-invariance.
#
# Supersedes the stock-`backlogit init` basis of probes 14/15/16 (rev-5 open item O-7).
# The disposable workspace is seeded from the LIVE .backlogit control files, NOT from
# `backlogit init` defaults:
#   config.yaml      - status enum (NOTE: has NO `shipped` value -> root cause, plan section 2)
#   header-def.yaml  - field definitions (no `complexity` on task)
#   hooks.yaml       - validate_transition + pre_task_completion_gate (absent from stock init)
#   registry.yaml    - queue/archive routing by status
#   migration.yaml   - document class map
#   templates/*.md   - artifact templates
#
# Models the exact PREREQUISITE topology, not a generic fixture:
#   001-F      root covering feature, live in queue, OUTSIDE the manifest   (= 022-F)
#   001.001-T  sole live task, `done` at closure                            (= 022.001-T)
#   001-S      task-only shipment, manifest exactly [001.001-T]             (= 021-S)
#   002-S      INBOUND edge holder: 002-S depends_on 001-S --type blocks    (= 017-S -> 021-S)
#
# Primary question: does the inbound-edge-holding record (002-S) survive the
# prerequisite's ShipShipment call BYTE-IDENTICAL?  This is the record that plan
# section 6.2.2 places in the baseline scope and section 6.2.3 locks.
#
# AUTHORIZATION SCOPE: CLI-only. The MCP surface is NOT probed and NOT authorized.

$ErrorActionPreference = 'Continue'

# ---------------------------------------------------------------- digest gate
$EXPECTED_SHA = '1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98'
$cmd = Get-Command backlogit -ErrorAction Stop
$EXE = $cmd.Source
$sha = (Get-FileHash -Algorithm SHA256 -Path $EXE).Hash
Write-Output "ENGINE_PATH=$EXE"
Write-Output "ENGINE_SHA256=$sha"
Write-Output "ENGINE_SHA256_EXPECTED=$EXPECTED_SHA"
if ($sha -ne $EXPECTED_SHA) {
    Write-Output "DIGEST_GATE=FAIL"
    Write-Output "HALT: engine digest mismatch - probe refresh required before any operation."
    exit 1
}
Write-Output "DIGEST_GATE=PASS"
$ver = (& $EXE --version --no-update-check 2>&1 | Select-Object -First 1)
Write-Output "ENGINE_VERSION=$ver"
Write-Output "AUTHORIZATION_SURFACE=CLI-only (MCP not probed, not authorized)"

# Repo root resolved via git, not by counting `..` segments. The committed
# probe16 script resolved `..\..\..` from this 4-deep evidence directory, which
# lands on `docs\`, not the repo root - a reproducibility defect recorded in the
# plan's evidence-integrity table.
$repo = (& git -C $PSScriptRoot rev-parse --show-toplevel) -replace '/','\'
if (-not $repo -or -not (Test-Path (Join-Path $repo '.backlogit'))) {
    Write-Output "HALT: could not resolve repo root with a live .backlogit directory."
    exit 1
}
Write-Output "REPO_ROOT=$repo"
$ws   = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p18'
function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function Say($m) { Write-Output "### $m" }

# ------------------------------------------------------- live-config seeding
Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ws\.backlogit\queue","$ws\.backlogit\archive","$ws\.backlogit\templates" | Out-Null
$seeded = @()
foreach ($n in 'config.yaml','header-def.yaml','hooks.yaml','registry.yaml','migration.yaml') {
    $src = Join-Path $repo ".backlogit\$n"
    Copy-Item $src (Join-Path $ws ".backlogit\$n") -Force
    $seeded += ('{0} sha256={1}' -f $n, (Get-FileHash $src -Algorithm SHA256).Hash)
}
Copy-Item (Join-Path $repo '.backlogit\templates\*') (Join-Path $ws '.backlogit\templates\') -Force
Say "SEEDED FROM LIVE CONFIG (not `backlogit init` defaults)"
$seeded | ForEach-Object { Write-Output "SEED $_" }
$statusEnum = (Get-Content (Join-Path $ws '.backlogit\config.yaml') -Raw)
Write-Output ("SEED_STATUS_ENUM_HAS_SHIPPED=" + ($statusEnum -match '(?s)status:.*?values:.*?-\s*shipped'))
Write-Output ("SEED_HOOKS_VALIDATE_TRANSITION=" + ((Get-Content (Join-Path $ws '.backlogit\hooks.yaml') -Raw) -match 'validate_transition:\s*true'))
Write-Output ("SEED_HOOKS_PRE_TASK_GATE=" + ((Get-Content (Join-Path $ws '.backlogit\hooks.yaml') -Raw) -match 'pre_task_completion_gate'))

Say "STEP 1: build the exact prerequisite topology under live config"
B add --type feature --title "Covering root feature" | Out-Null
B add --type task --title "Sole live task" --parent 001-F | Out-Null
B shipment create --title "Prerequisite task-only shipment" --items 001.001-T | Out-Null
B shipment create --title "Successor blocked shipment" | Out-Null
B dep add 002-S 001-S --type blocks | Out-Null
Write-Output "TOPOLOGY:"
B list

Say "STEP 2: ROOT CAUSE re-confirmation under LIVE config - generic move to shipped"
B shipment claim 001-S | Out-Null
$mv = B move 001-S --status shipped
$mvExit = $LASTEXITCODE
Write-Output "MOVE_SHIPPED_EXIT=$mvExit"
$mv | ForEach-Object { Write-Output "MOVE_SHIPPED_OUT: $_" }

Say "STEP 3: task to done under live hooks (pre_task_completion_gate observable here)"
$mvT = B move 001.001-T --status done
Write-Output "MOVE_TASK_DONE_EXIT=$LASTEXITCODE"
$mvT | ForEach-Object { Write-Output "MOVE_TASK_DONE_OUT: $_" }

# ------------------------------------------- baseline over the relation closure
Say "STEP 4: BASELINE hash of the relation closure (includes the INBOUND edge holder 002-S)"
$closure = @('queue\001-F.md','queue\001.001-T.md','queue\001-S.md','queue\002-S.md')
$base = @{}
foreach ($rel in $closure) {
    $p = Join-Path $ws ".backlogit\$rel"
    if (Test-Path $p) {
        $base[$rel] = (Get-FileHash $p -Algorithm SHA256).Hash
        Write-Output ("BASELINE {0,-24} {1}" -f $rel, $base[$rel])
    } else {
        Write-Output ("BASELINE {0,-24} ABSENT" -f $rel)
    }
}
$inboundBefore = Get-Content (Join-Path $ws '.backlogit\queue\002-S.md') -Raw
Write-Output "INBOUND_EDGE_RECORD_BEFORE:"
$inboundBefore -split "`n" | ForEach-Object { Write-Output "  | $_" }

Say "STEP 5: AUTHORIZED CALL - ShipShipment on the prerequisite"
$shipOut = B shipment ship 001-S --sha 1f55e6aa2a4507bcc1a9c7000d742a8710db4ca1 --message "merge: probe18 prerequisite" --author "stage-probe@local"
Write-Output "SHIP_EXIT=$LASTEXITCODE"
$shipOut | ForEach-Object { Write-Output "SHIP_OUT: $_" }

Say "STEP 6: INBOUND-EDGE BYTE-INVARIANCE (the load-bearing assertion)"
$inboundAfterPath = Join-Path $ws '.backlogit\queue\002-S.md'
$inboundMoved = $false
if (-not (Test-Path $inboundAfterPath)) {
    $inboundAfterPath = Join-Path $ws '.backlogit\archive\002-S.md'
    $inboundMoved = $true
}
Write-Output "INBOUND_RECORD_PATH_CHANGED=$inboundMoved"
$inboundAfter = if (Test-Path $inboundAfterPath) { Get-Content $inboundAfterPath -Raw } else { $null }
$inboundHashAfter = if (Test-Path $inboundAfterPath) { (Get-FileHash $inboundAfterPath -Algorithm SHA256).Hash } else { 'ABSENT' }
Write-Output "INBOUND_EDGE_SHA_BEFORE=$($base['queue\002-S.md'])"
Write-Output "INBOUND_EDGE_SHA_AFTER =$inboundHashAfter"
$inboundIdentical = ($inboundHashAfter -eq $base['queue\002-S.md']) -and (-not $inboundMoved)
Write-Output "INBOUND_EDGE_BYTE_IDENTICAL=$inboundIdentical"
Write-Output "INBOUND_EDGE_RECORD_AFTER:"
if ($inboundAfter) { $inboundAfter -split "`n" | ForEach-Object { Write-Output "  | $_" } }

Say "STEP 7: parent feature preservation + archived shipment provenance"
$featLive = Test-Path (Join-Path $ws '.backlogit\queue\001-F.md')
Write-Output "PARENT_FEATURE_STILL_LIVE_IN_QUEUE=$featLive"
$featHash = if ($featLive) { (Get-FileHash (Join-Path $ws '.backlogit\queue\001-F.md') -Algorithm SHA256).Hash } else { 'ABSENT' }
Write-Output "PARENT_FEATURE_BYTE_IDENTICAL=$($featHash -eq $base['queue\001-F.md'])"
$arch = Join-Path $ws '.backlogit\archive\001-S.md'
if (Test-Path $arch) {
    Write-Output "ARCHIVED_SHIPMENT_RECORD:"
    Get-Content $arch | Where-Object { $_ -match 'status|archived_status|commit|^id' } | ForEach-Object { Write-Output "  | $_" }
}

Say "STEP 8: dependency edge persistence after official sync"
B sync | Out-Null
Write-Output "DEP_LIST_002S:"
B dep list 002-S
Write-Output "SHIPMENT_LIST:"
B shipment list

Say "STEP 9: residual - no DB/WAL artifacts leave the scratch workspace"
Write-Output "SCRATCH_PATH=$($ws.Replace($repo,''))"
$ignored = $false
Push-Location $repo
git check-ignore -q ".backlogit/runtime/stage-probe-scratch" 2>$null
$ignored = ($LASTEXITCODE -eq 0)
Pop-Location
Write-Output "SCRATCH_GITIGNORED=$ignored"

Say "VERDICT"
Write-Output "INBOUND_EDGE_BYTE_IDENTICAL=$inboundIdentical"
Write-Output "PARENT_FEATURE_PRESERVED=$($featHash -eq $base['queue\001-F.md'])"
