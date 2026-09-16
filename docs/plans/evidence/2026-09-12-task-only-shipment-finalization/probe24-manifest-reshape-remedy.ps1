# Probe 24 - does re-shaping C to a FULLY-COVERED-ROOT manifest dispose `024-F`
#            mechanically, by the engine, with no barrier edge required?
#
# Question (three-shipment bootstrap re-plan, PRIMARY BLOCKER 2, remedy validation):
#   Probe 22 measured: claiming a TASK-ONLY shipment activates its OUT-OF-MANIFEST
#     covering feature parent, and that feature REMAINS `active` after the task-only
#     close. `024-F` therefore becomes a live top-level release unit that trips Ship
#     Step 1's P-001 `Active` gate when `017-S` is routed.
#   Probe 23 measured: the natural barrier - a `blocks` edge from the successor
#     SHIPMENT to the covering FEATURE - is NOT CONSTRUCTIBLE. The engine refuses it:
#       `add shipment block: prerequisite 001-F has type "feature";
#        both endpoints must be shipments`
#     So no dependency-graph barrier can be written against a feature at all.
#
#   REMEDY UNDER TEST: stop creating the un-disposable feature in the first place.
#   Re-shape `023-S` from TASK-ONLY `[024.001-T]` to FULLY-COVERED-ROOT
#   `[024-F, 024.001-T]`, exactly as `021-S` and `022-S` are already shaped. The
#   feature then becomes a MANIFEST MEMBER, so the engine's own cascade archives it as
#   part of the close. No barrier is needed because no residue is created.
#
# MEASURED:
#   Q1 - does the cascade archive the FEATURE as well as the task and the shipment?
#   Q2 - are there ZERO live top-level release units afterwards (P-001 clean)?
#   Q3 - does the successor shipment become eligible with no further Stage action?
#   Q4 - CONTROL: re-confirm the task-only shape leaves the feature `active`, so the
#        difference measured is attributable to manifest shape and nothing else.
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

# ======================================================= ARM Q - remedy under test
Say "ARM Q - C re-shaped to FULLY-COVERED-ROOT [001-F, 001.001-T] (the 023-S remedy)"
$ws = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p24-q'
Seed $ws
function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function StatusOf($id) {
    $out = B get $id
    $line = $out | Where-Object { $_ -match '^\s*status:' } | Select-Object -First 1
    if (-not $line) { return "<not-found>" }
    return ($line -replace '^\s*status:\s*','').Trim()
}
function Readiness($label) {
    $raw = & autoharness gate dag-readiness --workspace $ws --json 2>&1
    $ex = $LASTEXITCODE
    $obj = $null
    try { $obj = ($raw | Out-String) | ConvertFrom-Json } catch { }
    $ready = if ($obj) { @($obj.ready_set) -join ',' } else { '<unparsed>' }
    $next  = if ($obj) { $obj.next_eligible } else { '<unparsed>' }
    Write-Output ("  {0,-38} GATE_EXIT={1} ready_set=[{2}] next_eligible={3}" -f $label, $ex, $ready, $next)
}
Write-Output "SEEDED_FROM_LIVE_CONFIG=True  WS=$ws"

B add --type feature --title "C covering feature - 024-F analogue" | Out-Null
B add --type task --title "C sole task - 024.001-T analogue" --parent 001-F | Out-Null
B add --type feature --title "Successor release unit - 018-F analogue" | Out-Null
B add --type task --title "Successor task - 018.00x-T analogue" --parent 002-F | Out-Null
B shipment create --title "C fully-covered-root shipment - 023-S remedy shape" --items "001-F,001.001-T" | Out-Null
B shipment create --title "Successor task-only shipment - 017-S analogue" --items 002.001-T | Out-Null
B dep add 002-S 001-S --type blocks | Out-Null
Write-Output "MANIFEST_C_REMEDY:"
B shipment get 001-S | Where-Object { $_ -match '"items"|001-F|001.001-T' } | ForEach-Object { Write-Output "  $($_.Trim())" }

Write-Output ("  pre-claim: 001-F=" + (StatusOf '001-F') + " 001.001-T=" + (StatusOf '001.001-T'))
B shipment claim 001-S | Out-Null
Write-Output ("  after claim: 001-F=" + (StatusOf '001-F') + " 001.001-T=" + (StatusOf '001.001-T'))
Readiness "R1 after C claim"

Say "ARM Q - Ship 4.1/4.5 task lifecycle, then Step 6.1(a1) feature completion (A's grant)"
B move 001.001-T --status active | Out-Null
B move 001.001-T --status done | Out-Null
Write-Output ("  after task done: 001-F=" + (StatusOf '001-F') + " 001.001-T=" + (StatusOf '001.001-T'))
$fm = B move 001-F --status done
Write-Output "  A1_FEATURE_COMPLETION_EXIT=$LASTEXITCODE"
$fm | ForEach-Object { Write-Output "    $_" }
Write-Output ("  after a1: 001-F=" + (StatusOf '001-F'))

Say "ARM Q - Q1: the CASCADE close over a fully-covered-root manifest"
$shipOut = B shipment ship 001-S --sha 7bc03180000000000000000000000000000000000 --message "merge: probe24" --author "stage-probe@local"
Write-Output "  SHIP_EXIT=$LASTEXITCODE"
$shipOut | ForEach-Object { Write-Output "    $_" }
B sync | Out-Null
$fQ = StatusOf '001-F'
Write-Output ("  after cascade: 001-F=" + $fQ + " 001.001-T=" + (StatusOf '001.001-T') + " 001-S=" + (StatusOf '001-S'))
$archivedFeature = Test-Path (Join-Path $ws '.backlogit\archive\001-F.md')
Write-Output "  Q1_FEATURE_ARCHIVED_BY_CASCADE=$archivedFeature"

Say "ARM Q - Q2: live top-level release units remaining (P-001 surface)"
$liveFeatures = Get-ChildItem (Join-Path $ws '.backlogit\queue') -Filter '*-F.md' -ErrorAction SilentlyContinue | ForEach-Object { $_.BaseName }
foreach ($lf in $liveFeatures) { Write-Output ("    live feature " + $lf + " = " + (StatusOf $lf)) }
$otherActive = @($liveFeatures | Where-Object { $_ -ne '002-F' -and (StatusOf $_) -eq 'active' })
Write-Output "  Q2_OTHER_ACTIVE_TOPLEVEL_UNITS=$($otherActive.Count)"

Say "ARM Q - Q3: successor eligibility with NO further Stage action"
Readiness "R2 after C cascade close"
$rawQ = & autoharness gate dag-readiness --workspace $ws --json 2>&1
$readySetQ = @((($rawQ | Out-String) | ConvertFrom-Json).ready_set)
$q3 = $readySetQ -contains '002-S'
Write-Output "  Q3_SUCCESSOR_ELIGIBLE_WITHOUT_DISPOSITION=$q3"

# ======================================================= ARM T - control (task-only)
Say "ARM T - CONTROL: identical topology, C left TASK-ONLY [001.001-T]"
$wsT = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p24-t'
Seed $wsT
function BT { & $EXE --cwd $wsT --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function StatusOfT($id) {
    $out = BT get $id
    $line = $out | Where-Object { $_ -match '^\s*status:' } | Select-Object -First 1
    if (-not $line) { return "<not-found>" }
    return ($line -replace '^\s*status:\s*','').Trim()
}
BT add --type feature --title "C covering feature - 024-F analogue" | Out-Null
BT add --type task --title "C sole task - 024.001-T analogue" --parent 001-F | Out-Null
BT add --type feature --title "Successor release unit - 018-F analogue" | Out-Null
BT add --type task --title "Successor task - 018.00x-T analogue" --parent 002-F | Out-Null
BT shipment create --title "C task-only shipment - current 023-S shape" --items 001.001-T | Out-Null
BT shipment create --title "Successor task-only shipment - 017-S analogue" --items 002.001-T | Out-Null
BT dep add 002-S 001-S --type blocks | Out-Null
BT shipment claim 001-S | Out-Null
BT move 001.001-T --status active | Out-Null
BT move 001.001-T --status done | Out-Null
BT shipment ship 001-S --sha 7bc03180000000000000000000000000000000000 --message "merge: probe24-control" --author "stage-probe@local" | Out-Null
BT sync | Out-Null
$fT = StatusOfT '001-F'
$archivedFeatureT = Test-Path (Join-Path $wsT '.backlogit\archive\001-F.md')
Write-Output ("  CONTROL after task-only close: 001-F=" + $fT)
Write-Output "  CONTROL_FEATURE_ARCHIVED=$archivedFeatureT"

# ============================================== ARM QR - the LIVE construction order
Say "ARM QR - EXACT live 023-S construction: task-only manifest, THEN 'shipment add' the feature"
Write-Output "  Rationale: ARM Q builds the manifest feature-first via 'shipment create --items 001-F,001.001-T'."
Write-Output "  The LIVE 023-S was built the other way - it already existed as task-only and the feature was"
Write-Output "  APPENDED, giving items [024.001-T, 024-F] (TASK-FIRST). That is the one order never probed,"
Write-Output "  and it is the order the re-shape actually produced. Measured here rather than assumed."
$wsQR = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p24-qr'
Seed $wsQR
function BQ { & $EXE --cwd $wsQR --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function StatusOfQR($id) {
    $out = BQ get $id
    $line = $out | Where-Object { $_ -match '^\s*status:' } | Select-Object -First 1
    if (-not $line) { return "<not-found>" }
    return ($line -replace '^\s*status:\s*','').Trim()
}
BQ add --type feature --title "C covering feature - 024-F analogue" | Out-Null
BQ add --type task --title "C sole task - 024.001-T analogue" --parent 001-F | Out-Null
BQ add --type feature --title "Successor release unit - 018-F analogue" | Out-Null
BQ add --type task --title "Successor task - 018.00x-T analogue" --parent 002-F | Out-Null
BQ shipment create --title "C task-only shipment (pre-reshape)" --items 001.001-T | Out-Null
BQ shipment create --title "Successor task-only shipment - 017-S analogue" --items 002.001-T | Out-Null
BQ dep add 002-S 001-S --type blocks | Out-Null
Write-Output "  -- manifest BEFORE append:"
BQ shipment get 001-S | Where-Object { $_ -match '"items"|001-F|001\.001-T' } | ForEach-Object { Write-Output "     $($_.Trim())" }
$addOut = BQ shipment add 001-S 001-F
Write-Output "  SHIPMENT_ADD_EXIT=$LASTEXITCODE"
$addOut | ForEach-Object { Write-Output "     $_" }
Write-Output "  -- manifest AFTER append (expect TASK-FIRST, matching live 023-S):"
BQ shipment get 001-S | Where-Object { $_ -match '"items"|001-F|001\.001-T|covering_feature' } | ForEach-Object { Write-Output "     $($_.Trim())" }
$itemsQR = ((BQ shipment get 001-S | Out-String) | ConvertFrom-Json).custom_fields.items
Write-Output "  MANIFEST_ORDER_QR=$($itemsQR -join ',')"
Write-Output "  MANIFEST_IS_TASK_FIRST=$($itemsQR[0] -eq '001.001-T')"

BQ shipment claim 001-S | Out-Null
Write-Output ("  after claim: 001-F=" + (StatusOfQR '001-F') + " 001.001-T=" + (StatusOfQR '001.001-T'))
BQ move 001.001-T --status active | Out-Null
BQ move 001.001-T --status done | Out-Null
BQ move 001-F --status done | Out-Null
Write-Output ("  after a1: 001-F=" + (StatusOfQR '001-F'))
$shipQR = BQ shipment ship 001-S --sha 7bc03180000000000000000000000000000000000 --message "merge: probe24-qr" --author "stage-probe@local"
Write-Output "  SHIP_EXIT=$LASTEXITCODE"
$shipQR | ForEach-Object { Write-Output "     $_" }
BQ sync | Out-Null
$fQR = StatusOfQR '001-F'
$archivedFeatureQR = Test-Path (Join-Path $wsQR '.backlogit\archive\001-F.md')
Write-Output ("  after cascade: 001-F=" + $fQR)
Write-Output "  QR_FEATURE_ARCHIVED_BY_CASCADE=$archivedFeatureQR"
$rawQR = & autoharness gate dag-readiness --workspace $wsQR --json 2>&1
$readySetQR = @((($rawQR | Out-String) | ConvertFrom-Json).ready_set)
Write-Output "  QR_READY_SET=$($readySetQR -join ',')"
$qr3 = $readySetQR -contains '002-S'
Write-Output "  QR_SUCCESSOR_ELIGIBLE=$qr3"

Say "VERDICT"
Write-Output "REMEDY_FULLY_COVERED_ROOT_FEATURE_STATUS=$fQ"
Write-Output "CONTROL_TASK_ONLY_FEATURE_STATUS=$fT"
Write-Output "Q1_CASCADE_ARCHIVES_FEATURE=$archivedFeature"
Write-Output "Q2_ZERO_OTHER_ACTIVE_TOPLEVEL_UNITS=$($otherActive.Count -eq 0)"
Write-Output "Q3_SUCCESSOR_ELIGIBLE_NO_BARRIER_NEEDED=$q3"
Write-Output "DIFFERENCE_ATTRIBUTABLE_TO_MANIFEST_SHAPE_ALONE=$($fQ -ne $fT)"
Write-Output "QR_LIVE_ORDER_FEATURE_STATUS=$fQR"
Write-Output "QR_LIVE_ORDER_FEATURE_ARCHIVED=$archivedFeatureQR"
Write-Output "QR_LIVE_ORDER_SUCCESSOR_ELIGIBLE=$qr3"
Write-Output "ORDER_INDEPENDENCE_MEASURED=$(($fQ -eq $fQR) -and ($archivedFeature -eq $archivedFeatureQR))"
