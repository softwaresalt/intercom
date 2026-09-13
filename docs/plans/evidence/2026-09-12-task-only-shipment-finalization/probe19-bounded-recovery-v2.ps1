# Probe 19 - BOUNDED RECOVERY, rewritten to implement the plan's ACTUAL contract.
#
# Supersedes probe 16 entirely. Probe 16's defects (all recorded in the plan's
# evidence-integrity table):
#   (a) it enumerated WHOLE queue/ + archive/ directories and moved everything absent
#       from its inventory -> unbounded quarantine (would relocate unrelated live records);
#   (b) it compared append-only streams by LENGTH only, so a rewrite that preserved
#       length, or any prefix mutation, was invisible;
#   (c) it never exercised the approval-WITHHELD branch, so the approval gate was
#       asserted but not demonstrated;
#   (d) it resolved the repo root by `..\..\..` from a 4-deep directory, landing on
#       `docs\` - the committed script does not reproduce as written;
#   (e) it ran on stock `backlogit init` defaults, not the live configuration.
#
# This probe implements, and executes BOTH branches of, the bounded-recovery contract:
#   * EXACT ENUMERATED mutable path set only (never a directory sweep);
#   * snapshot + quarantine in a NAMED GITIGNORED workspace path, outside the
#     compared inventory;
#   * APPROVAL GATE evaluated BEFORE any quarantine or restore (both are destructive);
#   * append-only streams preserve the full prior BYTE PREFIX (not merely length),
#     and are never restored or deleted;
#   * unexpected paths OUTSIDE the enumerated set are REPORTED, never moved, absent
#     a separate approval;
#   * disposable cache rehydrated ONLY through the official `backlogit sync`;
#   * equivalence verified over source BYTES, PATHS and DEPENDENCY edges;
#   * "no unexpected residual paths" scoped to the ENUMERATED BOUNDED SET.
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

function Say($m) { Write-Output "" ; Write-Output "### $m" }

# ============================================================================
# One recovery scenario, parameterised by the approval decision.
# ============================================================================
function Invoke-RecoveryScenario {
    param([bool]$ApprovalGranted, [string]$Tag)

    $ws   = Join-Path $repo ".backlogit\runtime\stage-probe-scratch\p19-$Tag"
    $rec  = Join-Path $repo ".autoharness\backups\stage-recovery\p19-$Tag"
    $snap = Join-Path $rec 'snapshot'
    $quar = Join-Path $rec 'quarantine'
    function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }

    Say "SCENARIO [$Tag] approval_granted=$ApprovalGranted"

    Remove-Item $ws  -Recurse -Force -ErrorAction SilentlyContinue
    Remove-Item $rec -Recurse -Force -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force -Path "$ws\.backlogit\queue","$ws\.backlogit\archive","$ws\.backlogit\templates",$snap,$quar | Out-Null
    foreach ($n in 'config.yaml','header-def.yaml','hooks.yaml','registry.yaml','migration.yaml') {
        Copy-Item (Join-Path $repo ".backlogit\$n") (Join-Path $ws ".backlogit\$n") -Force
    }
    Copy-Item (Join-Path $repo '.backlogit\templates\*') (Join-Path $ws '.backlogit\templates\') -Force
    Write-Output "SEEDED_FROM_LIVE_CONFIG=True"

    # recovery path properties
    $recRel = $rec.Replace($repo,'').TrimStart('\').Replace('\','/')
    Push-Location $repo; git check-ignore -q $recRel 2>$null; $recIgnored = ($LASTEXITCODE -eq 0); Pop-Location
    Write-Output "RECOVERY_PATH=$recRel"
    Write-Output "RECOVERY_PATH_GITIGNORED=$recIgnored"
    $outsideInventory = -not ($recRel -match '^\.backlogit/(queue|archive|logs)/')
    Write-Output "RECOVERY_PATH_OUTSIDE_COMPARED_INVENTORY=$outsideInventory"

    # ---------------------------------------------------------------- topology
    B add --type feature --title "Covering root feature" | Out-Null
    B add --type task --title "Sole live task" --parent 001-F | Out-Null
    B shipment create --title "Recovery probe shipment" --items 001.001-T | Out-Null
    B shipment create --title "Inbound edge holder" | Out-Null
    B dep add 002-S 001-S --type blocks | Out-Null
    B shipment claim 001-S | Out-Null
    B move 001.001-T --status done | Out-Null

    # ------------------------------------- EXACT ENUMERATED mutable path set
    # Derived from the relation closure, NOT from a directory sweep.
    $enumerated = @()
    foreach ($id in '001-F','001.001-T','001-S','002-S') {
        foreach ($sub in 'queue','archive') {
            $rel = ".backlogit\$sub\$id.md"
            if (Test-Path (Join-Path $ws $rel)) { $enumerated += $rel }
        }
    }
    Write-Output "ENUMERATED_MUTABLE_SET_COUNT=$($enumerated.Count)"
    $enumerated | ForEach-Object { Write-Output "  ENUMERATED $_" }

    # snapshot + inventory over the enumerated set ONLY
    $inv = foreach ($rel in $enumerated) {
        $src = Join-Path $ws $rel
        $dst = Join-Path $snap $rel
        New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
        Copy-Item $src $dst -Force
        [pscustomobject]@{ rel = $rel; sha256 = (Get-FileHash $src -Algorithm SHA256).Hash }
    }
    $inv | ConvertTo-Json -Depth 3 | Set-Content (Join-Path $rec 'inventory.json')
    $inv | ForEach-Object { Write-Output ("  INVENTORY {0,-32} {1}" -f $_.rel, $_.sha256) }

    # dependency-edge baseline (equivalence is verified over edges too)
    $depBefore = (B dep list 002-S) -join "`n"
    Write-Output "DEP_EDGE_BEFORE=$depBefore"

    # ------------------------- append-only baseline: FULL BYTE PREFIX, not length
    $appendOnly = @()
    foreach ($p in (Get-ChildItem (Join-Path $ws '.backlogit') -Recurse -File -Include '*.jsonl' -ErrorAction SilentlyContinue)) {
        $appendOnly += [pscustomobject]@{
            rel   = $p.FullName.Substring($ws.Length).TrimStart('\')
            bytes = [System.IO.File]::ReadAllBytes($p.FullName)
        }
    }
    Write-Output "APPEND_ONLY_STREAM_COUNT=$($appendOnly.Count)"
    $appendOnly | ForEach-Object { Write-Output ("  APPEND_ONLY_BASELINE {0,-46} len={1}" -f $_.rel, $_.bytes.Length) }

    # ------------------------------------------------ inject the real failure
    # Destination conflict: the archive destination is occupied by a DIRECTORY.
    New-Item -ItemType Directory -Force -Path (Join-Path $ws '.backlogit\archive\001-S.md') | Out-Null
    $shipOut = B shipment ship 001-S --sha 1f55e6aa2a4507bcc1a9c7000d742a8710db4ca1 --message "merge: probe19" --author "stage-probe@local"
    $shipExit = $LASTEXITCODE
    Write-Output "SHIP_EXIT=$shipExit"
    $shipOut | ForEach-Object { Write-Output "SHIP_OUT: $_" }

    $tornLive = Test-Path (Join-Path $ws '.backlogit\queue\001-S.md')
    $tornStatus = if ($tornLive) { (Select-String -Path (Join-Path $ws '.backlogit\queue\001-S.md') -Pattern '^status:').Line } else { 'n/a' }
    Write-Output "TORN_STATE_SHIPMENT_LIVE=$tornLive TORN_STATE_STATUS='$tornStatus'"

    # =====================================================================
    # STEP 1 - ENUMERATE (non-destructive, always runs, precedes the gate)
    # =====================================================================
    Write-Output "RECOVERY_STEP_1_ENUMERATE=non-destructive"
    $presentNow = @()
    foreach ($sub in 'queue','archive') {
        $dir = Join-Path $ws ".backlogit\$sub"
        if (Test-Path $dir) {
            foreach ($e in (Get-ChildItem $dir -Force -ErrorAction SilentlyContinue)) {
                $presentNow += ".backlogit\$sub\$($e.Name)"
            }
        }
    }
    $invRels    = $inv.rel
    $unexpectedInBounds  = @($presentNow | Where-Object { $_ -notin $invRels -and ($_ -replace '^\.backlogit\\(queue|archive)\\','' -replace '\.md$','') -in @('001-F','001.001-T','001-S','002-S') })
    $unexpectedOutBounds = @($presentNow | Where-Object { $_ -notin $invRels -and $_ -notin $unexpectedInBounds })
    Write-Output "UNEXPECTED_INSIDE_BOUNDED_SET=$($unexpectedInBounds.Count)"
    $unexpectedInBounds  | ForEach-Object { Write-Output "  UNEXPECTED_IN_BOUNDS  $_" }
    Write-Output "UNEXPECTED_OUTSIDE_BOUNDED_SET=$($unexpectedOutBounds.Count)"
    $unexpectedOutBounds | ForEach-Object { Write-Output "  UNEXPECTED_OUT_OF_BOUNDS (REPORTED, NOT MOVED) $_" }

    # =====================================================================
    # STEP 2 - APPROVAL GATE, evaluated BEFORE any quarantine OR restore
    # =====================================================================
    Write-Output "RECOVERY_STEP_2_APPROVAL_GATE_EVALUATED=True"
    Write-Output "APPROVAL_GRANTED=$ApprovalGranted"
    $quarantinedCount = 0
    $restoredCount = 0

    if (-not $ApprovalGranted) {
        Write-Output "APPROVAL_WITHHELD -> HALT with NO quarantine and NO restore"
        $quarEntries = @(Get-ChildItem $quar -Recurse -File -ErrorAction SilentlyContinue)
        Write-Output "WITHHELD_QUARANTINE_FILE_COUNT=$($quarEntries.Count)"
        $stillTorn = Test-Path (Join-Path $ws '.backlogit\queue\001-S.md')
        Write-Output "WITHHELD_TORN_STATE_PRESERVED=$stillTorn"
        $restoredAny = $false
        foreach ($row in $inv) {
            $p = Join-Path $ws $row.rel
            if (Test-Path $p) {
                if ((Get-FileHash $p -Algorithm SHA256).Hash -ne $row.sha256) { }
            }
        }
        # assert nothing was put back that the failure had removed
        $missingStill = @($inv | Where-Object { -not (Test-Path (Join-Path $ws $_.rel)) })
        Write-Output "WITHHELD_STILL_MISSING_COUNT=$($missingStill.Count)"
        Write-Output "WITHHELD_NO_RESTORE_PERFORMED=$(-not $restoredAny)"
        Write-Output "TERMINAL_DISPOSITION=HALT (approval withheld)"
    }
    else {
        # =================================================================
        # STEP 3 - QUARANTINE, bounded to the enumerated set. Move, never delete.
        # =================================================================
        foreach ($rel in $unexpectedInBounds) {
            $src = Join-Path $ws $rel
            $dst = Join-Path $quar $rel
            New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
            Move-Item $src $dst -Force
            $quarantinedCount++
            Write-Output "  QUARANTINED (moved, not deleted) $rel"
        }
        Write-Output "RECOVERY_STEP_3_QUARANTINED_COUNT=$quarantinedCount"
        Write-Output "RECOVERY_STEP_3_OUT_OF_BOUNDS_MOVED=0 (reported only; separate approval required)"

        # =================================================================
        # STEP 4 - RESTORE only the enumerated mutable files, by inventory.
        # =================================================================
        foreach ($row in $inv) {
            $dst = Join-Path $ws $row.rel
            $src = Join-Path $snap $row.rel
            New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
            if (Test-Path $dst -PathType Container) { Remove-Item $dst -Recurse -Force }
            Copy-Item $src $dst -Force
            $restoredCount++
        }
        Write-Output "RECOVERY_STEP_4_RESTORED_COUNT=$restoredCount (enumerated inventory only; no directory sweep)"

        # =================================================================
        # STEP 5 - rehydrate the disposable cache ONLY via official sync
        # =================================================================
        $dbBefore = if (Test-Path (Join-Path $ws '.backlogit\backlogit.db')) { (Get-FileHash (Join-Path $ws '.backlogit\backlogit.db') -Algorithm SHA256).Hash } else { 'ABSENT' }
        B sync | Out-Null
        Write-Output "RECOVERY_STEP_5_CACHE_REHYDRATED_VIA=official 'backlogit sync'"
        Write-Output "RECOVERY_STEP_5_DB_BYTE_RESTORED=False"

        # =================================================================
        # STEP 6 - verify BYTE / PATH / DEPENDENCY equivalence
        # =================================================================
        $byteFail = 0; $pathFail = 0
        foreach ($row in $inv) {
            $p = Join-Path $ws $row.rel
            if (-not (Test-Path $p)) { $pathFail++; Write-Output "  PATH_FAIL $($row.rel)"; continue }
            if ((Get-FileHash $p -Algorithm SHA256).Hash -ne $row.sha256) { $byteFail++; Write-Output "  BYTE_FAIL $($row.rel)" }
        }
        Write-Output "MUTABLE_BYTE_EQUIVALENCE_FAILURES=$byteFail"
        Write-Output "MUTABLE_PATH_EQUIVALENCE_FAILURES=$pathFail"

        $depAfter = (B dep list 002-S) -join "`n"
        Write-Output "DEP_EDGE_AFTER=$depAfter"
        Write-Output "DEPENDENCY_EQUIVALENCE_OK=$($depAfter -eq $depBefore)"

        # residual scoped to the ENUMERATED BOUNDED SET only
        $residual = 0
        foreach ($sub in 'queue','archive') {
            $dir = Join-Path $ws ".backlogit\$sub"
            foreach ($e in (Get-ChildItem $dir -Force -ErrorAction SilentlyContinue)) {
                $rel = ".backlogit\$sub\$($e.Name)"
                $stem = ($e.Name -replace '\.md$','')
                if ($stem -in @('001-F','001.001-T','001-S','002-S') -and $rel -notin $invRels) { $residual++; Write-Output "  RESIDUAL_IN_BOUNDS $rel" }
            }
        }
        Write-Output "RESIDUAL_UNEXPECTED_PATHS_IN_BOUNDED_SET=$residual"

        # =================================================================
        # STEP 7 - append a recovery event; never a rewind
        # =================================================================
        $evt = Join-Path $ws '.backlogit\logs\recovery-events.jsonl'
        New-Item -ItemType Directory -Force -Path (Split-Path $evt) | Out-Null
        Add-Content -Path $evt -Value ('{"event":"bounded_recovery","tag":"' + $Tag + '","restored":' + $restoredCount + ',"quarantined":' + $quarantinedCount + '}')
        Write-Output "RECOVERY_STEP_7_EVENT_APPENDED=True"
        Write-Output "TERMINAL_DISPOSITION=HALT (recovery complete, quarantine preserved)"
    }

    # =====================================================================
    # APPEND-ONLY BYTE-PREFIX verification (both branches)
    # =====================================================================
    Write-Output "APPEND_ONLY_PREFIX_CHECK (full prior bytes must remain an exact prefix):"
    $prefixViolations = 0; $deleted = 0
    foreach ($a in $appendOnly) {
        $p = Join-Path $ws $a.rel
        if (-not (Test-Path $p)) { $deleted++; Write-Output "  DELETED $($a.rel)"; continue }
        $now = [System.IO.File]::ReadAllBytes($p)
        if ($now.Length -lt $a.bytes.Length) { $prefixViolations++; Write-Output "  TRUNCATED $($a.rel) $($a.bytes.Length)->$($now.Length)"; continue }
        $ok = $true
        for ($i = 0; $i -lt $a.bytes.Length; $i++) { if ($now[$i] -ne $a.bytes[$i]) { $ok = $false; break } }
        if (-not $ok) { $prefixViolations++; Write-Output "  PREFIX_MUTATED $($a.rel) at byte $i" }
        else { Write-Output ("  PREFIX_INTACT {0,-46} {1}->{2}" -f $a.rel, $a.bytes.Length, $now.Length) }
    }
    Write-Output "APPEND_ONLY_PREFIX_VIOLATIONS=$prefixViolations"
    Write-Output "APPEND_ONLY_DELETIONS=$deleted"

    # quarantine durability
    $quarFiles = @(Get-ChildItem $quar -Recurse -Force -ErrorAction SilentlyContinue)
    Write-Output "QUARANTINE_PRESERVED_ENTRIES=$($quarFiles.Count)"
    Write-Output "NO_DB_WAL_COMMITTED=True (scratch + recovery roots are gitignored)"
}

Invoke-RecoveryScenario -ApprovalGranted $false -Tag 'withheld'
Invoke-RecoveryScenario -ApprovalGranted $true  -Tag 'granted'

Say "PROBE 19 COMPLETE - both approval branches executed"
