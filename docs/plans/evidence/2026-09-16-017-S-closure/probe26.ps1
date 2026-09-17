# Probe 26 - claim-cascade behaviour over the exact 017-S root-included manifest shape
# Stage-owned, isolated-fixture measurement. Closes the plan-review P0:
# "what does `backlogit shipment claim` do to a 13-member manifest containing
#  11 members declaring status: archived (archived_status: queued)?"
$ErrorActionPreference = 'Continue'
$repo = 'C:\Source\GitHub\intercom-go'
$scratch = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p26'

function Emit($k, $v) { Write-Output ("{0}={1}" -f $k, $v) }

# --- engine identity gate -------------------------------------------------
$engine = (Get-Command backlogit).Source
$sha = (Get-FileHash $engine -Algorithm SHA256).Hash
Emit 'ENGINE_RESOLVED_VIA' 'Get-Command backlogit (registered command; not a hardcoded path)'
Emit 'ENGINE_PATH_OBSERVED' $engine
Emit 'ENGINE_SHA256' $sha
Emit 'ENGINE_SHA256_EXPECTED' '1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98'
Emit 'DIGEST_GATE' $(if ($sha -eq '1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98') { 'PASS' } else { 'FAIL' })
Emit 'ENGINE_VERSION' ((backlogit version 2>&1 | Select-Object -First 1) -replace '\s+', ' ')
Emit 'PROBE_HEAD' (git -C $repo rev-parse --short HEAD)
Emit 'PROBE_RUN_BY' 'Stage (pre-claim planning evidence; Ship re-verifies read-only)'

# --- live-backlog isolation baseline --------------------------------------
$dirtyBefore = (git -C $repo status --porcelain -- .backlogit | Measure-Object).Count
Emit 'LIVE_BACKLOGIT_DIRTY_PATHS_BEFORE' $dirtyBefore

if (Test-Path $scratch) { Remove-Item $scratch -Recurse -Force }
New-Item -ItemType Directory -Path $scratch -Force | Out-Null
backlogit init $scratch *>&1 | Out-Null
Emit 'FIXTURE_BACKLOG_ROOT' $scratch
Emit 'FIXTURE_ISOLATED_FROM_LIVE' ((Resolve-Path $scratch).Path -ne (Resolve-Path $repo).Path)

function BL { backlogit --cwd $scratch @args 2>&1 }

# resolve the storage root that `backlogit init` actually created
$storeRoot = if (Test-Path (Join-Path $scratch '.backlogit')) { Join-Path $scratch '.backlogit' } else { Join-Path $scratch '.backlog' }
Emit 'FIXTURE_STORAGE_ROOT' $storeRoot

# --- STEP 1: build the EXACT 017-S topology -------------------------------
Write-Output "`n### STEP 1: build EXACT 017-S topology (root + 3 tasks + 8 subtasks + 1 live task)"
BL add --type feature --title 'Root-included fully-covered-root fixture' | Out-Null
foreach ($n in 1..3) { BL add --type task --title "Legacy task $n" --parent 001-F | Out-Null }
foreach ($p in @('001.001-T', '001.002-T')) { foreach ($n in 1..3) { BL add --type subtask --title "Legacy ST $n" --parent $p | Out-Null } }
foreach ($n in 1..2) { BL add --type subtask --title "Legacy ST $n" --parent '001.003-T' | Out-Null }
BL add --type task --title 'Live task (the 018.008-T analogue)' --parent 001-F | Out-Null
# control artifacts OUTSIDE the manifest
BL add --type feature --title 'Control out-of-manifest feature' | Out-Null
BL add --type task --title 'Control out-of-manifest task' --parent 002-F | Out-Null

$legacy = @('001.001.001-ST', '001.001.002-ST', '001.001.003-ST', '001.002.001-ST', '001.002.002-ST', '001.002.003-ST',
    '001.003.001-ST', '001.003.002-ST', '001.001-T', '001.002-T', '001.003-T')
$live = '001.004-T'
$root = '001-F'
$controls = @('002-F', '002.001-T')

# --- STEP 2: archive the 11 legacy DIRECTLY FROM queued -------------------
Write-Output "`n### STEP 2: archive 11 legacy descendants directly from queued (yields archived_status: queued)"
foreach ($id in $legacy) { BL archive $id | Out-Null }

function State($id) {
    $q = Join-Path $storeRoot "queue\$id.md"
    $a = Join-Path $storeRoot "archive\$id.md"
    $loc = if (Test-Path $q) { 'queue' } elseif (Test-Path $a) { 'archive' } else { 'MISSING' }
    $f = if (Test-Path $q) { $q } elseif (Test-Path $a) { $a } else { $null }
    if (-not $f) { return [pscustomobject]@{ id = $id; loc = $loc; status = 'MISSING'; astatus = '-'; parent = '-' } }
    $c = Get-Content $f -Raw
    [pscustomobject]@{
        id      = $id; loc = $loc
        status  = $(if ($c -match '(?m)^status:\s*(\S+)') { $matches[1] } else { '?' })
        astatus = $(if ($c -match '(?m)^archived_status:\s*(\S+)') { $matches[1] } else { 'NONE' })
        parent  = $(if ($c -match '(?m)^parent_id:\s*(\S+)') { $matches[1] } else { '(unset)' })
    }
}

$manifest = @($root) + $legacy + @($live)
$preArchivedCount = ($legacy | ForEach-Object { State $_ } | Where-Object { $_.status -eq 'archived' }).Count
Emit 'TOPOLOGY_MEMBERS' $manifest.Count
Emit 'ROOT_INCLUDED' ($manifest -contains $root)
Emit 'PRE_ARCHIVED' $preArchivedCount

# FIXTURE-VALIDITY GATE: a fixture that never reached the target shape can never
# be reported as a measurement of that shape. Fail closed rather than emit a
# verdict about a topology that was not actually built.
$archivedFlavour = (State $legacy[0]).astatus
Emit 'ARCHIVED_STATUS_FLAVOUR' $archivedFlavour
if ($preArchivedCount -ne 11 -or $archivedFlavour -ne 'queued') {
    Emit 'FIXTURE_VALIDITY_GATE' 'FAIL'
    Emit 'PROBE26_RESULT' 'INVALID - fixture did not reach the 11-pre-archived/archived_status:queued shape; no claim measurement was taken'
    Remove-Item $scratch -Recurse -Force
    exit 1
}
Emit 'FIXTURE_VALIDITY_GATE' 'PASS'

# --- STEP 3: create the shipment -----------------------------------------
Write-Output "`n### STEP 3: create shipment with the 13-member manifest"
BL shipment create --title 'Root-included claim-cascade fixture' --items ($manifest -join ',') | Out-Null

Write-Output "`n### STEP 4: PRE-CLAIM state"
$pre = @{}
foreach ($id in ($manifest + $controls)) { $s = State $id; $pre[$id] = $s; Write-Output ("  PRE  {0,-18} loc={1,-8} status={2,-9} archived_status={3,-8} parent={4}" -f $s.id, $s.loc, $s.status, $s.astatus, $s.parent) }

# --- STEP 5: THE MEASUREMENT - claim the shipment -------------------------
Write-Output "`n### STEP 5: THE MEASUREMENT - backlogit shipment claim 001-S"
$claimOut = BL shipment claim 001-S
Emit 'CLAIM_EXIT_CODE' $LASTEXITCODE
Write-Output ("CLAIM_OUTPUT=" + (($claimOut | Out-String) -replace '(?m)^time=.*$', '' -replace '\s+', ' ').Trim())

Write-Output "`n### STEP 6: POST-CLAIM state"
$post = @{}
foreach ($id in ($manifest + $controls)) { $s = State $id; $post[$id] = $s; Write-Output ("  POST {0,-18} loc={1,-8} status={2,-9} archived_status={3,-8} parent={4}" -f $s.id, $s.loc, $s.status, $s.astatus, $s.parent) }

# --- STEP 7: criteria -----------------------------------------------------
Write-Output "`n### STEP 7: CRITERIA"
$c1 = $post[$root].status -eq 'active'
Emit 'ROOT_FEATURE_POST_STATUS' $post[$root].status
Emit 'C1_ROOT_FEATURE_QUEUED_TO_ACTIVE' $(if ($c1) { 'PASS' } else { 'FAIL' })

$c2 = $post[$live].status -eq 'active'
Emit 'LIVE_TASK_POST_STATUS' $post[$live].status
Emit 'C2_LIVE_TASK_QUEUED_TO_ACTIVE' $(if ($c2) { 'PASS' } else { 'FAIL' })

$drifted = @()
foreach ($id in $legacy) {
    if ($post[$id].status -ne 'archived' -or $post[$id].astatus -ne $pre[$id].astatus -or $post[$id].loc -ne 'archive') { $drifted += $id }
}
Emit 'ARCHIVED_MEMBERS_DRIFTED' $(if ($drifted.Count -eq 0) { '(none)' } else { $drifted -join ',' })
Emit 'C3_ARCHIVED_MEMBERS_UNCHANGED_BY_CLAIM' $(if ($drifted.Count -eq 0) { 'PASS' } else { 'FAIL' })

$ctrlDrift = @()
foreach ($id in $controls) { if ($post[$id].status -ne $pre[$id].status -or $post[$id].loc -ne $pre[$id].loc) { $ctrlDrift += $id } }
Emit 'CONTROL_OUT_OF_MANIFEST_DRIFTED' $(if ($ctrlDrift.Count -eq 0) { '(none)' } else { $ctrlDrift -join ',' })
Emit 'C4_NOTHING_OUTSIDE_MANIFEST_TOUCHED' $(if ($ctrlDrift.Count -eq 0) { 'PASS' } else { 'FAIL' })

$parentDrift = @()
foreach ($id in $manifest) { if ($post[$id].parent -ne $pre[$id].parent) { $parentDrift += $id } }
Emit 'C5_PARENT_ID_PRESERVED' $(if ($parentDrift.Count -eq 0) { 'PASS' } else { 'FAIL' })

# --- STEP 8: live-backlog isolation re-proof ------------------------------
$dirtyAfter = (git -C $repo status --porcelain -- .backlogit | Measure-Object).Count
Emit 'LIVE_BACKLOGIT_DIRTY_PATHS_AFTER' $dirtyAfter
Emit 'LIVE_BACKLOG_UNMUTATED' ($dirtyAfter -eq $dirtyBefore)

# --- STEP 9: aggregate ----------------------------------------------------
Write-Output "`n### STEP 9: AGGREGATE VERDICT"
$crit = @{ C1_ROOT_FEATURE_QUEUED_TO_ACTIVE = $c1; C2_LIVE_TASK_QUEUED_TO_ACTIVE = $c2;
    C3_ARCHIVED_MEMBERS_UNCHANGED_BY_CLAIM = ($drifted.Count -eq 0); C4_NOTHING_OUTSIDE_MANIFEST_TOUCHED = ($ctrlDrift.Count -eq 0);
    C5_PARENT_ID_PRESERVED = ($parentDrift.Count -eq 0); C6_LIVE_BACKLOG_UNMUTATED = ($dirtyAfter -eq $dirtyBefore)
}
$failed = 0
foreach ($k in ($crit.Keys | Sort-Object)) { $r = $(if ($crit[$k]) { 'PASS' } else { 'FAIL' }); if (-not $crit[$k]) { $failed++ }; Write-Output ("  {0,-42} {1}" -f $k, $r) }
Emit 'FAILED_CRITERIA' $failed
Emit 'PROBE26_RESULT' $(if ($failed -eq 0) { 'PASS' } else { 'FAIL' })

Remove-Item $scratch -Recurse -Force
Emit 'SCRATCH_REMOVED' (-not (Test-Path $scratch))
