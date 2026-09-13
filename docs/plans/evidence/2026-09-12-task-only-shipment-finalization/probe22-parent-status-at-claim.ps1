# Probe 22 - covering-feature status under shipment claim and task-only close.
#
# Question (three-shipment bootstrap re-plan, PRIMARY BLOCKER 2):
#   Does claiming a shipment transition the covering FEATURE parent to `active`?
#   And after a task-only close, what status does that feature hold?
#
#   This adjudicates two separate plan claims that were previously ASSERTED, not measured:
#     (a) C plan section 8.3 / R-5: `024-F` is left "live" after C's task-only closure and
#         may trip Ship Step 1's P-001 single-in-flight gate, which tests for status
#         `Active`. If claim never moves an out-of-manifest parent off `queued`, the gate
#         is not tripped and no barrier is needed before `017-S`.
#     (b) A plan PO-1a / B plan section 8 step 1: whether claim leaves an IN-manifest
#         covering feature `active` (A's section 5 condition 5 requires exactly `active`),
#         and whether B's `023-F -> active` is an observed claim effect or an
#         unauthorized direct move by Ship.
#
# TWO ARMS, measured independently in two disposable workspaces:
#   ARM C  (T4 shape)   - manifest is TASK-ONLY [001.001-T]; feature 001-F is live and
#                         OUTSIDE the manifest. Matches `023-S` exactly.
#   ARM AB (root shape) - manifest is FULLY-COVERED-ROOT [001-F, 001.001-T]; the feature
#                         IS a manifest member. Matches `021-S` / `022-S` exactly.
#
# The workspaces are disposable, gitignored, workspace-contained copies seeded from the
# LIVE .backlogit control files and the LIVE .autoharness config. The live backlog is
# never mutated. Scratch is removed by the caller after the transcript is captured.
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
Write-Output ("ENGINE_VERSION=" + (& $EXE --version --no-update-check 2>&1 | Select-Object -First 1))
Write-Output "AUTHORIZATION_SURFACE=CLI-only (MCP not probed, not authorized)"

$repo = (& git -C $PSScriptRoot rev-parse --show-toplevel) -replace '/','\'
if (-not $repo -or -not (Test-Path (Join-Path $repo '.backlogit'))) {
    Write-Output "HALT: could not resolve repo root with a live .backlogit directory."
    exit 1
}
Write-Output "REPO_ROOT=$repo"
Write-Output ("PROBE_HEAD=" + (& git -C $repo rev-parse --short HEAD))

function Seed($ws) {
    Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path "$ws\.backlogit\queue","$ws\.backlogit\archive","$ws\.backlogit\templates","$ws\.autoharness" | Out-Null
    foreach ($n in 'config.yaml','header-def.yaml','hooks.yaml','registry.yaml','migration.yaml') {
        $src = Join-Path $repo ".backlogit\$n"
        if (Test-Path $src) { Copy-Item $src (Join-Path $ws ".backlogit\$n") -Force }
    }
    Copy-Item (Join-Path $repo '.backlogit\templates\*') (Join-Path $ws '.backlogit\templates\') -Force
    foreach ($n in 'config.yaml','backlog-registry.yaml','workspace-profile.yaml','harness-manifest.yaml') {
        $src = Join-Path $repo ".autoharness\$n"
        if (Test-Path $src) { Copy-Item $src (Join-Path $ws ".autoharness\$n") -Force }
    }
}

function Say($m) { Write-Output ""; Write-Output "### $m" }

# Status reader - parses the `status:` field of `backlogit get <id>`.
function StatusOf($ws, $id) {
    $out = & $EXE --cwd $ws --no-update-check get $id 2>&1 | Where-Object { $_ -notmatch '^time=' }
    $line = $out | Where-Object { $_ -match '^\s*status:' } | Select-Object -First 1
    if (-not $line) { return "<not-found>" }
    return ($line -replace '^\s*status:\s*','').Trim()
}

function Snapshot($ws, $label, $featureId, $taskId, $shipId) {
    $f = StatusOf $ws $featureId
    $t = StatusOf $ws $taskId
    $s = StatusOf $ws $shipId
    Write-Output ("  {0,-34} feature {1}={2,-10} task {3}={4,-10} shipment {5}={6}" -f $label, $featureId, $f, $taskId, $t, $shipId, $s)
}

# ============================================================== ARM C (T4 shape)
Say "ARM C - TASK-ONLY manifest, covering feature OUTSIDE the manifest (matches 023-S)"
$wsC = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p22-c'
Seed $wsC
function BC { & $EXE --cwd $wsC --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
Write-Output "SEEDED_FROM_LIVE_CONFIG=True  WS=$wsC"

BC add --type feature --title "Covering root feature (outside manifest)" | Out-Null
BC add --type task --title "Sole manifest task" --parent 001-F | Out-Null
BC shipment create --title "Task-only shipment (C shape)" --items 001.001-T | Out-Null
Write-Output "MANIFEST_C:"
BC shipment get 001-S | Where-Object { $_ -match 'items|"id"|status' } | Select-Object -First 12 | ForEach-Object { Write-Output "  $_" }

Snapshot $wsC "C0 after create (pre-claim)" '001-F' '001.001-T' '001-S'
$c0f = StatusOf $wsC '001-F'

Say "ARM C - STEP 1: claim the task-only shipment (Ship Step 0.5.4)"
$claimOut = BC shipment claim 001-S
Write-Output "CLAIM_EXIT=$LASTEXITCODE"
$claimOut | ForEach-Object { Write-Output "  CLAIM_OUT: $_" }
Snapshot $wsC "C1 AFTER CLAIM" '001-F' '001.001-T' '001-S'
$c1f = StatusOf $wsC '001-F'

Say "ARM C - STEP 2: Ship Step 4.1 moves the task active, Step 4.5 moves it done"
BC move 001.001-T --status active | Out-Null
Snapshot $wsC "C2 after task->active" '001-F' '001.001-T' '001-S'
BC move 001.001-T --status done | Out-Null
Snapshot $wsC "C3 after task->done" '001-F' '001.001-T' '001-S'

Say "ARM C - STEP 3: TASK-ONLY close - archive ONLY the manifest item, then close the record"
$arcOut = BC archive 001.001-T
Write-Output "ARCHIVE_TASK_EXIT=$LASTEXITCODE"
$arcOut | ForEach-Object { Write-Output "  ARCHIVE_OUT: $_" }
Snapshot $wsC "C4 after manifest task archived" '001-F' '001.001-T' '001-S'

Write-Output ""
Write-Output "  -- shipment-record close route A: backlogit move <ship> --status shipped (Ship 6.1.b text)"
$mvOut = BC move 001-S --status shipped
$mvExit = $LASTEXITCODE
Write-Output "  MOVE_SHIPPED_EXIT=$mvExit"
$mvOut | ForEach-Object { Write-Output "    MOVE_OUT: $_" }
Snapshot $wsC "C5 after move --status shipped" '001-F' '001.001-T' '001-S'

if ($mvExit -ne 0) {
    Write-Output ""
    Write-Output "  -- route A refused; falling back to the cascade op to reach a CLOSED record (measurement only)"
    $shOut = BC shipment ship 001-S --sha 7bc03180000000000000000000000000000000000 --message "merge: probe22" --author "stage-probe@local"
    Write-Output "  SHIPMENT_SHIP_EXIT=$LASTEXITCODE"
    $shOut | ForEach-Object { Write-Output "    SHIP_OUT: $_" }
    Snapshot $wsC "C6 after cascade shipment ship" '001-F' '001.001-T' '001-S'
}

BC sync | Out-Null
Snapshot $wsC "C_FINAL after sync" '001-F' '001.001-T' '001-S'
$cFinalF = StatusOf $wsC '001-F'

Write-Output ""
Write-Output "ARM_C_FEATURE_STATUS_PRE_CLAIM=$c0f"
Write-Output "ARM_C_FEATURE_STATUS_AFTER_CLAIM=$c1f"
Write-Output "ARM_C_FEATURE_STATUS_AFTER_TASKONLY_CLOSE=$cFinalF"
Write-Output "ARM_C_FEATURE_EVER_ACTIVE=$(@($c0f,$c1f,$cFinalF) -contains 'active')"

# ============================================================= ARM AB (root shape)
Say "ARM AB - FULLY-COVERED-ROOT manifest, covering feature IS a member (matches 021-S / 022-S)"
$wsAB = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p22-ab'
Seed $wsAB
function BA { & $EXE --cwd $wsAB --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
Write-Output "SEEDED_FROM_LIVE_CONFIG=True  WS=$wsAB"

BA add --type feature --title "Covering root feature (manifest member)" | Out-Null
BA add --type task --title "Sole manifest task" --parent 001-F | Out-Null
BA shipment create --title "Fully-covered-root shipment (A/B shape)" --items "001-F,001.001-T" | Out-Null
Write-Output "MANIFEST_AB:"
BA shipment get 001-S | Where-Object { $_ -match 'items|"id"|status' } | Select-Object -First 12 | ForEach-Object { Write-Output "  $_" }

Snapshot $wsAB "AB0 after create (pre-claim)" '001-F' '001.001-T' '001-S'
$a0f = StatusOf $wsAB '001-F'

Say "ARM AB - STEP 1: claim the fully-covered-root shipment"
$claimOutAB = BA shipment claim 001-S
Write-Output "CLAIM_EXIT=$LASTEXITCODE"
$claimOutAB | ForEach-Object { Write-Output "  CLAIM_OUT: $_" }
Snapshot $wsAB "AB1 AFTER CLAIM" '001-F' '001.001-T' '001-S'
$a1f = StatusOf $wsAB '001-F'

Say "ARM AB - STEP 2: task active -> done (Ship 4.1 / 4.5)"
BA move 001.001-T --status active | Out-Null
BA move 001.001-T --status done | Out-Null
Snapshot $wsAB "AB2 after task->done" '001-F' '001.001-T' '001-S'

Say "ARM AB - STEP 3: can the feature be completed by an explicit move (A's a1 grant)?"
$fmOut = BA move 001-F --status done
Write-Output "MOVE_FEATURE_DONE_EXIT=$LASTEXITCODE"
$fmOut | ForEach-Object { Write-Output "  MOVE_OUT: $_" }
Snapshot $wsAB "AB3 after feature->done" '001-F' '001.001-T' '001-S'
$a3f = StatusOf $wsAB '001-F'

Write-Output ""
Write-Output "ARM_AB_FEATURE_STATUS_PRE_CLAIM=$a0f"
Write-Output "ARM_AB_FEATURE_STATUS_AFTER_CLAIM=$a1f"
Write-Output "ARM_AB_FEATURE_ACTIVE_AT_CLAIM=$($a1f -eq 'active')"
Write-Output "ARM_AB_FEATURE_DONE_MOVE_ACCEPTED=$($a3f -eq 'done')"

# ============================================================== P-001 evaluation
Say 'P-001 gate evaluation - Ship Step 1 item 1 tests for status Active'
$shipAgent = Join-Path $repo '.github\agents\_ship.agent.md'
Select-String -Path $shipAgent -Pattern 'P-001 Gate' | Select-Object -First 1 | ForEach-Object {
    Write-Output ("  L{0,-5} {1}" -f $_.LineNumber, ($_.Line.Trim() -replace '\s+',' '))
}

Say "VERDICT"
Write-Output "Q1_CLAIM_ACTIVATES_OUT_OF_MANIFEST_PARENT=$($c1f -eq 'active')"
Write-Output "Q2_CLAIM_ACTIVATES_IN_MANIFEST_PARENT=$($a1f -eq 'active')"
Write-Output "Q3_TASKONLY_CLOSE_LEAVES_PARENT_STATUS=$cFinalF"
Write-Output "Q4_PARENT_IS_ACTIVE_FOR_P001_AFTER_TASKONLY_CLOSE=$($cFinalF -eq 'active')"
