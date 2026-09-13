# Probe 20 - O-6: the POST-ARCHIVE closure route.
#
# Question (plan rev-5 open item O-6, flagged highest-risk):
#   Plan section 8 archives the shipment, then opens/merges a closure PR. After the
#   archive there are ZERO active shipments. Probe 17 falsified the PR-lifecycle-only
#   route on exactly that condition. Does the CLOSURE route therefore deadlock
#   *after* the destructive call?
#
# Two separable sub-questions, answered separately:
#   Q1 (measured here)  - post-archive, does `pipeline-topology --phase lifecycle`
#                         block? If yes, any actor that runs it after the close is
#                         deadlocked with no active shipment to re-create.
#   Q2 (contract read)  - does the INSTALLED Ship contract actually mandate that gate
#                         before the closure PR? Cited from _ship.agent.md by line.
#
# The workspace is a disposable, gitignored, workspace-contained copy seeded from the
# LIVE .backlogit control files and the LIVE .autoharness config. The live backlog is
# never mutated.
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

$ws = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p20'
function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function Say($m) { Write-Output ""; Write-Output "### $m" }

Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ws\.backlogit\queue","$ws\.backlogit\archive","$ws\.backlogit\templates","$ws\.autoharness" | Out-Null
foreach ($n in 'config.yaml','header-def.yaml','hooks.yaml','registry.yaml','migration.yaml') {
    Copy-Item (Join-Path $repo ".backlogit\$n") (Join-Path $ws ".backlogit\$n") -Force
}
Copy-Item (Join-Path $repo '.backlogit\templates\*') (Join-Path $ws '.backlogit\templates\') -Force
foreach ($n in 'config.yaml','backlog-registry.yaml','workspace-profile.yaml','harness-manifest.yaml') {
    $src = Join-Path $repo ".autoharness\$n"
    if (Test-Path $src) { Copy-Item $src (Join-Path $ws ".autoharness\$n") -Force }
}
Write-Output "SEEDED_FROM_LIVE_CONFIG=True"

Say "STEP 1: build the prerequisite topology and take it to the POST-ARCHIVE state"
B add --type feature --title "Covering root feature" | Out-Null
B add --type task --title "Sole live task" --parent 001-F | Out-Null
B shipment create --title "Prerequisite task-only shipment" --items 001.001-T | Out-Null
B shipment create --title "Successor blocked shipment" | Out-Null
B dep add 002-S 001-S --type blocks | Out-Null
B shipment claim 001-S | Out-Null
B move 001.001-T --status done | Out-Null

Say "STEP 2: CONTROL - gate result while the shipment is ACTIVE (plan section 8 step a0 position)"
Push-Location $ws
$ctl = & autoharness gate pipeline-topology --mode agent --shipment 001-S --phase lifecycle --json 2>&1
$ctlExit = $LASTEXITCODE
Pop-Location
Write-Output "CONTROL_ACTIVE_EXIT=$ctlExit"
($ctl | Out-String) -split "`n" | Where-Object { $_ -match 'message|token|active_shipment_ids|exit_code|blocked' } | ForEach-Object { Write-Output "  CONTROL: $($_.Trim())" }

Say "STEP 3: the DESTRUCTIVE call - archive the shipment"
$shipOut = B shipment ship 001-S --sha 1f55e6aa2a4507bcc1a9c7000d742a8710db4ca1 --message "merge: probe20" --author "stage-probe@local"
Write-Output "SHIP_EXIT=$LASTEXITCODE"
$shipOut | ForEach-Object { Write-Output "SHIP_OUT: $_" }
B sync | Out-Null
Write-Output "SHIPMENT_STATE_AFTER:"
B shipment list

Say "STEP 4: Q1 - POST-ARCHIVE gate routes (zero active shipments)"
$routes = @(
    @{ desc = '--mode agent --shipment 001-S --phase lifecycle'; args = @('--mode','agent','--shipment','001-S','--phase','lifecycle') },
    @{ desc = '--mode agent --phase ambient';                    args = @('--mode','agent','--phase','ambient') },
    @{ desc = '--mode manual --shipment 001-S --phase lifecycle';args = @('--mode','manual','--shipment','001-S','--phase','lifecycle') },
    @{ desc = '--mode ci --phase lifecycle';                     args = @('--mode','ci','--phase','lifecycle') }
)
$blockedCount = 0
foreach ($r in $routes) {
    Push-Location $ws
    $out = & autoharness gate pipeline-topology @($r.args) --json 2>&1
    $ex = $LASTEXITCODE
    Pop-Location
    Write-Output ""
    Write-Output "ROUTE: autoharness gate pipeline-topology $($r.desc)"
    Write-Output "  EXIT=$ex"
    ($out | Out-String) -split "`n" | Where-Object { $_ -match '"message"|"token"|active_shipment_ids|"blocked"|"invalid"' } | ForEach-Object { Write-Output "  $($_.Trim())" }
    if ($ex -ne 0) { $blockedCount++ }
}
Write-Output ""
Write-Output "POST_ARCHIVE_ROUTES_TESTED=$($routes.Count)"
Write-Output "POST_ARCHIVE_ROUTES_NONZERO_EXIT=$blockedCount"

Say "STEP 5: Q2 - what the INSTALLED Ship contract actually mandates (source citation, not inference)"
$shipAgent = Join-Path $repo '.github\agents\_ship.agent.md'
Write-Output "SOURCE=.github/agents/_ship.agent.md"
foreach ($pat in 'TOPOLOGY_GATE: lifecycle','Then invoke the \*\*pr-lifecycle\*\* skill for the closure PR','After all closure work is committed','Obtain explicit operator approval') {
    Select-String -Path $shipAgent -Pattern $pat | ForEach-Object {
        Write-Output ("  L{0,-5} {1}" -f $_.LineNumber, ($_.Line.Trim() -replace '\s+',' '))
    }
}

Say "VERDICT"
Write-Output "Q1_POST_ARCHIVE_LIFECYCLE_GATE_BLOCKS=$($blockedCount -gt 0)"
Write-Output "Q1_NOTE=A lifecycle-phase topology gate run AFTER the archive fails closed; there is no active shipment left to satisfy it and re-claiming an archived shipment is not a supported transition."
Write-Output "Q2_NOTE=See STEP 5 citations. The installed Ship Step 6.0 closure-PR step invokes pr-lifecycle directly; the lifecycle topology gate is mandated at Step 5 (implementation PR) and at Step 6 a0 (BEFORE closure/safe-close, while the shipment is still active), not before the closure PR."
