# Probe 23 - is a shipment -> FEATURE dependency edge a MECHANICALLY ENFORCED barrier?
#
# Question (three-shipment bootstrap re-plan, PRIMARY BLOCKER 2, second branch):
#   Probe 22 measured that claiming a TASK-ONLY shipment transitions its OUT-OF-MANIFEST
#   covering feature parent to `active`, and that the feature REMAINS `active` after the
#   task-only close. `024-F` therefore becomes a live top-level release unit that Ship
#   Step 1's P-001 gate (line 308, tests for status `Active`) will trip when `017-S` is
#   routed. The re-plan directive requires a MECHANICALLY ENFORCED barrier before `017-S`
#   eligibility - prose-only cleanup (C plan section 8.3) is explicitly insufficient, and
#   an invalid shipment status is not an option (backlogit has no shipment `blocked`).
#
#   CANDIDATE BARRIER: a `blocks` dependency edge from the successor SHIPMENT to the
#   covering FEATURE - `017-S depends_on 024-F`. Probe 14 already proved shipment ->
#   SHIPMENT edges gate `dag-readiness`. It did NOT prove shipment -> FEATURE edges do.
#   That is the gap this probe closes.
#
# MEASURED, in dependency-graph terms, with no live-backlog mutation:
#   Q1 - does `dag-readiness` SUPPRESS the successor shipment while the barrier feature
#        is still `active` (i.e. after the predecessor shipment has already archived)?
#   Q2 - does the successor become ELIGIBLE once Stage disposes of the feature
#        (`move --status done`, then `archive`)?
#   Q3 - is the edge preserved and readable across the disposal?
#
# If Q1 is False the candidate barrier is advisory metadata, not enforcement, and the
# re-plan must return MUST_REPLAN rather than ship a prose-only gate.
#
# The workspace is a disposable, gitignored, workspace-contained copy seeded from the
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
Write-Output ("GATE_ENGINE=" + (Get-Command autoharness -ErrorAction SilentlyContinue).Source)
Write-Output "AUTHORIZATION_SURFACE=CLI-only (MCP not probed, not authorized)"

$repo = (& git -C $PSScriptRoot rev-parse --show-toplevel) -replace '/','\'
if (-not $repo -or -not (Test-Path (Join-Path $repo '.backlogit'))) {
    Write-Output "HALT: could not resolve repo root with a live .backlogit directory."
    exit 1
}
Write-Output "REPO_ROOT=$repo"
Write-Output ("PROBE_HEAD=" + (& git -C $repo rev-parse --short HEAD))

$ws = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p23'
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
Write-Output "SEEDED_FROM_LIVE_CONFIG=True  WS=$ws"

function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function Say($m) { Write-Output ""; Write-Output "### $m" }

function StatusOf($id) {
    $out = B get $id
    $line = $out | Where-Object { $_ -match '^\s*status:' } | Select-Object -First 1
    if (-not $line) { return "<not-found>" }
    return ($line -replace '^\s*status:\s*','').Trim()
}

# Returns the parsed dag-readiness object.
function ReadySet($label) {
    $raw = & autoharness gate dag-readiness --workspace $ws --json 2>&1
    $ex = $LASTEXITCODE
    $obj = $null
    try { $obj = ($raw | Out-String) | ConvertFrom-Json } catch { }
    $set = @()
    if ($obj) { $set = @($obj.ready_set) }
    Write-Host ("  {0,-40} GATE_EXIT={1} ready_set=[{2}] next_eligible={3}" -f $label, $ex, ($set -join ','), $(if($obj){$obj.next_eligible}else{'<unparsed>'}))
    return ,$set
}

Say "STEP 1: build the C analogue (001-F / 001.001-T / 001-S task-only) and the 017-S analogue (002-F / 002.001-T / 002-S)"
B add --type feature --title "C covering feature (outside manifest) - 024-F analogue" | Out-Null
B add --type task --title "C sole manifest task - 024.001-T analogue" --parent 001-F | Out-Null
B add --type feature --title "Successor release unit - 018-F analogue" | Out-Null
B add --type task --title "Successor task - 018.00x-T analogue" --parent 002-F | Out-Null
B shipment create --title "C task-only shipment - 023-S analogue" --items 001.001-T | Out-Null
B shipment create --title "Successor task-only shipment - 017-S analogue" --items 002.001-T | Out-Null

Say "STEP 2: write BOTH edges - the existing chain edge AND the candidate barrier edge"
B dep add 002-S 001-S --type blocks
B dep add 002-S 001-F --type blocks
Write-Output "### edge readback (002-S):"
B dep list 002-S | ForEach-Object { Write-Output "  $_" }

Say "STEP 3: BASELINE readiness (both predecessors live)"
ReadySet "R0 baseline" | Out-Null
Write-Output ("  statuses: 001-F=" + (StatusOf '001-F') + " 001-S=" + (StatusOf '001-S') + " 002-S=" + (StatusOf '002-S'))

Say "STEP 4: claim the C shipment - Probe 22 showed this activates the out-of-manifest parent"
B shipment claim 001-S | Out-Null
Write-Output ("  statuses: 001-F=" + (StatusOf '001-F') + " 001-S=" + (StatusOf '001-S'))
ReadySet "R1 after C claim" | Out-Null

Say "STEP 5: complete + close the C shipment task-only (its own task archives; 001-F is NOT a member)"
B move 001.001-T --status active | Out-Null
B move 001.001-T --status done | Out-Null
$shipOut = B shipment ship 001-S --sha 7bc03180000000000000000000000000000000000 --message "merge: probe23" --author "stage-probe@local"
Write-Output "SHIP_EXIT=$LASTEXITCODE"
$shipOut | ForEach-Object { Write-Output "  SHIP_OUT: $_" }
B sync | Out-Null
$featAfterClose = StatusOf '001-F'
Write-Output ("  statuses: 001-F=" + $featAfterClose + " 001-S=" + (StatusOf '001-S') + " 002-S=" + (StatusOf '002-S'))

Say "STEP 6: Q1 - with the predecessor SHIPMENT archived but the barrier FEATURE still active, is the successor SUPPRESSED?"
$set2 = ReadySet "R2 barrier feature still active"
$edgeExists = ((B dep list 002-S | Out-String) -match '001-F')
$suppressed = ($edgeExists -and ($set2 -notcontains '002-S'))
Write-Output "  Q1_SUCCESSOR_SUPPRESSED_BY_FEATURE_BARRIER=$suppressed"

Say "STEP 7: CONTROL - confirm the chain edge alone would NOT have suppressed it (predecessor shipment is already archived)"
Write-Output "  001-S archived_status:"
$arc = Join-Path $ws '.backlogit\archive\001-S.md'
if (Test-Path $arc) { Select-String -Path $arc -Pattern 'archived_status|^status:' | ForEach-Object { Write-Output "    $($_.Line.Trim())" } }

Say "STEP 8: STAGE DISPOSITION - the barrier's designated unblocking action (move done, then archive)"
$mv = B move 001-F --status done
Write-Output "  MOVE_FEATURE_DONE_EXIT=$LASTEXITCODE"
$mv | ForEach-Object { Write-Output "    $_" }
$ar = B archive 001-F
Write-Output "  ARCHIVE_FEATURE_EXIT=$LASTEXITCODE"
$ar | ForEach-Object { Write-Output "    $_" }
B sync | Out-Null
Write-Output ("  statuses: 001-F=" + (StatusOf '001-F'))

Say "STEP 9: Q2 - is the successor now ELIGIBLE?"
$set3 = ReadySet "R3 after Stage disposition"
$released = ($set3 -contains '002-S')
Write-Output "  Q2_SUCCESSOR_RELEASED_AFTER_DISPOSITION=$released"

Say "STEP 10: Q3 - edge preserved across disposal?"
B dep list 002-S | ForEach-Object { Write-Output "  $_" }

Say "VERDICT"
Write-Output "Q0_FEATURE_ENDPOINT_EDGE_CONSTRUCTIBLE=$edgeExists"
Write-Output "Q1_FEATURE_EDGE_ENFORCES_BARRIER=$suppressed  (MOOT when Q0 is False - no edge exists to enforce anything)"
Write-Output "Q2_DISPOSITION_RELEASES_SUCCESSOR=$released"
Write-Output "BARRIER_IS_MECHANICALLY_ENFORCED=$($suppressed -and $released)"
Write-Output "FEATURE_STATUS_AFTER_TASKONLY_CLOSE=$featAfterClose"
