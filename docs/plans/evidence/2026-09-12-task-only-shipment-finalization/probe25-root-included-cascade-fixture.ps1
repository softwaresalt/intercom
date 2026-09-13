# Probe 25 - ROOT-INCLUDED cascade over a fixture with genuinely pre-archived members.
#
# Question (plan A rev 5, section 9.4 / AC-20):
#   Does the backlogit cascade close path behave as plan A section 9.3 CLASSIFIES it, on the
#   EXACT `017-S` shape - a 13-member manifest whose ROOT FEATURE IS INCLUDED and whose 11
#   legacy descendants are genuinely `status: archived` with `archived_status: queued`?
#
#   Probe 15 measured the INVERSE case (root EXCLUDED, 12-member task-only manifest, and its
#   pre-archived members carried `archived_status: done`). No committed probe has ever
#   exercised a root-INCLUDED cascade whose siblings are genuinely archived, and none has
#   exercised `archived_status: queued`. That is the gap this probe closes.
#
# OWNERSHIP (plan A rev 5, hybrid resolution of plan-review findings A-1 / A-2):
#   This probe is AUTHORED, EXECUTED and COMMITTED by **Stage**, BEFORE `021-S` is claimed.
#   It is plan/spike evidence under `docs/plans/evidence/`, which Ship's Role Boundary
#   forbids Ship to create or modify (P-010). The fresh Ship S2 session NEVER runs, edits or
#   regenerates this probe: S2 performs a READ-ONLY re-verification of this committed
#   evidence (section 9.4a). Stage authors; Ship verifies.
#
# THE LOAD-BEARING DETAIL: the 11 pre-archived members are archived DIRECTLY FROM `queued`,
#   so each carries `status: archived` + `archived_status: queued` - NOT `archived_status:
#   done`. A fixture built the `done` way is a strictly easier case, because it satisfies the
#   skill's `archived-completed(done)` role which these records do NOT.
#
# The workspace is a disposable, gitignored, workspace-contained copy seeded from the LIVE
# .backlogit control files and the LIVE .autoharness config. The live backlog is never
# mutated; every call below is `--cwd $ws`.
#
# AUTHORIZATION SCOPE: CLI-only. The MCP surface is NOT probed and NOT authorized.

$ErrorActionPreference = 'Continue'

# ---------------------------------------------------------------- digest gate
# PORTABILITY (plan-review finding A-3): the engine is resolved at RUNTIME from the
# registered command via Get-Command. No absolute path is a requirement of this contract.
# The observed path is RECORDED as telemetry only; the IDENTITY assertion is the SHA-256
# comparison against the committed expected digest. A mismatch HALTs.
$EXPECTED_SHA = '1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98'
$cmd = Get-Command backlogit -ErrorAction Stop
$EXE = $cmd.Source
$sha = (Get-FileHash -Algorithm SHA256 -Path $EXE).Hash
Write-Output "ENGINE_RESOLVED_VIA=Get-Command backlogit (registered command; not a hardcoded path)"
Write-Output "ENGINE_PATH_OBSERVED=$EXE"
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
Write-Output "PROBE_RUN_BY=Stage (pre-claim planning evidence; Ship re-verifies read-only)"

function Say($m) { Write-Output ""; Write-Output "### $m" }

# ---------------------------------------------------------------- seeded workspace
$ws = Join-Path $repo '.backlogit\runtime\stage-probe-scratch\p25'
Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path "$ws\.backlogit\queue","$ws\.backlogit\archive","$ws\.backlogit\templates","$ws\.autoharness" | Out-Null
foreach ($n in 'config.yaml','header-def.yaml','hooks.yaml','registry.yaml','migration.yaml') {
    $src = Join-Path $repo ".backlogit\$n"
    if (Test-Path $src) { Copy-Item $src (Join-Path $ws ".backlogit\$n") -Force }
}
Copy-Item (Join-Path $repo '.backlogit\templates\*') (Join-Path $ws '.backlogit\templates\') -Force -ErrorAction SilentlyContinue
foreach ($n in 'config.yaml','backlog-registry.yaml','workspace-profile.yaml','harness-manifest.yaml') {
    $src = Join-Path $repo ".autoharness\$n"
    if (Test-Path $src) { Copy-Item $src (Join-Path $ws ".autoharness\$n") -Force }
}
Write-Output "SEEDED_FROM_LIVE_CONFIG=True"
Write-Output "FIXTURE_BACKLOG_ROOT=$ws"

# ISOLATION PROOF (plan-review finding B-7): the fixture root must not be the live backlog
# and must not be a parent of it. It is deliberately NESTED under the live workspace's
# gitignored `.backlogit/runtime/` scratch area, which is the established probe convention
# (probes 18/20/22/24), so the assertion is that it is not the live QUEUE/ARCHIVE root.
$liveQueue = (Join-Path $repo '.backlogit\queue').ToLowerInvariant()
$fixQueue  = (Join-Path $ws  '.backlogit\queue').ToLowerInvariant()
Write-Output "LIVE_QUEUE_ROOT=$liveQueue"
Write-Output "FIXTURE_QUEUE_ROOT=$fixQueue"
Write-Output ("FIXTURE_ISOLATED_FROM_LIVE=" + ($fixQueue -ne $liveQueue))
$liveDirtyBefore = (& git -C $repo status --porcelain -- .backlogit | Measure-Object).Count
Write-Output "LIVE_BACKLOGIT_DIRTY_PATHS_BEFORE=$liveDirtyBefore"

function B { & $EXE --cwd $ws --no-update-check @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }
function NewId($out, $kind) {
    $line = $out | Where-Object { $_ -match "^Created $kind`: " } | Select-Object -First 1
    if (-not $line) { return $null }
    return ($line -replace "^Created $kind`: ",'').Trim()
}
# Reads a frontmatter field from wherever the record actually lives.
function Field($id, $name) {
    foreach ($d in @('queue','archive')) {
        $p = Join-Path $ws ".backlogit\$d\$id.md"
        if (Test-Path $p) {
            $c = Get-Content $p -Raw
            if ($c -match "(?m)^$name`:\s*(\S+)") { return $Matches[1] }
            return '(unset)'
        }
    }
    return '(not-found)'
}
function LocOf($id) {
    $q = Test-Path (Join-Path $ws ".backlogit\queue\$id.md")
    $a = Test-Path (Join-Path $ws ".backlogit\archive\$id.md")
    if ($q -and $a) { return 'BOTH' }
    if ($q) { return 'queue' }
    if ($a) { return 'archive' }
    return 'NONE'
}

# ============================================================== build the exact shape
Say "STEP 1: build the EXACT 017-S topology - root feature + 3 legacy tasks + 8 subtasks (3/3/2) + 1 live task"
$rootF = NewId (B add --type feature --title "Stage artifact branch/PR policy gap correction") 'feature'
$legacyTasks = @()
foreach ($n in 1..3) { $legacyTasks += NewId (B add --type task --title "Legacy task $n" --parent $rootF) 'task' }
$legacySubs = @()
$subPlan = @{ 0 = 3; 1 = 3; 2 = 2 }   # 3 + 3 + 2 = 8, matching the measured live shape
for ($i = 0; $i -lt 3; $i++) {
    foreach ($k in 1..($subPlan[$i])) {
        $legacySubs += NewId (B add --type subtask --title "Legacy subtask $($legacyTasks[$i])-$k" --parent $legacyTasks[$i]) 'subtask'
    }
}
$liveTask = NewId (B add --type task --title "Sole live task" --parent $rootF) 'task'

# Out-of-manifest CONTROL artifacts - PASS criterion 8 needs something that could be touched.
$ctrlF = NewId (B add --type feature --title "CONTROL feature - outside the manifest") 'feature'
$ctrlT = NewId (B add --type task --title "CONTROL task - outside the manifest" --parent $ctrlF) 'task'

Write-Output "ROOT_FEATURE=$rootF"
Write-Output ("LEGACY_TASKS=" + ($legacyTasks -join ','))
Write-Output ("LEGACY_SUBTASKS=" + ($legacySubs -join ','))
Write-Output "LIVE_TASK=$liveTask"
Write-Output "CONTROL_OUT_OF_MANIFEST=$ctrlF,$ctrlT"

Say "STEP 2: archive the 11 legacy descendants DIRECTLY FROM queued (subtasks first, then their tasks)"
Write-Output "  (this is what yields archived_status: queued - the harder case; NOT archived_status: done)"
$preArchived = @($legacySubs) + @($legacyTasks)
foreach ($id in $preArchived) { B archive $id | Out-Null }
foreach ($id in ($preArchived | Sort-Object)) {
    "  {0,-20} loc={1,-8} status={2,-10} archived_status={3,-8} parent={4}" -f `
        $id, (LocOf $id), (Field $id 'status'), (Field $id 'archived_status'), (Field $id 'parent_id')
}
$wrongFlavour = @($preArchived | Where-Object { (Field $_ 'archived_status') -ne 'queued' })
Write-Output ("PRE_ARCHIVED_COUNT=" + $preArchived.Count)
Write-Output ("PRE_ARCHIVED_ALL_ARCHIVED_STATUS_QUEUED=" + ($wrongFlavour.Count -eq 0))

Say "STEP 3: manifest = 13 members, ROOT FEATURE INCLUDED"
$manifest = @($rootF, $liveTask) + $preArchived
Write-Output ("MANIFEST_COUNT=" + $manifest.Count)
Write-Output ("MANIFEST=" + (($manifest | Sort-Object) -join ','))
$shipId = (B shipment create --title "Root-included fully-covered-root fixture" --items ($manifest -join ',') |
           Where-Object { $_ -match 'shipment' } | Select-Object -First 1)
Write-Output "SHIPMENT_CREATE_OUT=$shipId"
$shipId = '001-S'
Write-Output "SHIPMENT_ID=$shipId"

Say "STEP 4: claim, complete the live task, then the a1 covering-feature completion"
B shipment claim $shipId | Out-Null
Write-Output ("FEATURE_STATUS_AFTER_CLAIM=" + (Field $rootF 'status'))
B move $liveTask --status done | Out-Null
# This is exactly what plan A's Step 6.1(a1) authorizes: covering feature active -> done.
B move $rootF --status done | Out-Null
Write-Output ("FEATURE_STATUS_AFTER_A1=" + (Field $rootF 'status'))
Write-Output ("LIVE_TASK_STATUS=" + (Field $liveTask 'status'))

# ============================================================== PASS criterion 2
Say "STEP 5: PASS CRITERION 2 - pre-mode classification with expected_status: done"
# Four-value classification per shipment-reconcile `## Classification`:
#   matched | pre-archived | missing | status-mismatch
#
# DECLARED STATUS IS AUTHORITATIVE, NEVER LOCATION. The skill is explicit: "Declared
# `status` is read from the record's own frontmatter `status` field - never inferred from,
# nor substituted by, which of `queue/`/`archive/` currently holds the record. A record
# residing in `.backlogit/archive/` while declaring `status: done` is **not** truly
# archived; only a declared `status: archived` counts as truly archived."
#
# This matters here because the LIVE registry routes `done|accepted|rejected|archived` to
# `archive/` and only `queued|active|blocked|review` to `queue/`. So a member correctly
# completed to `done` ALSO leaves the queue. See MEMBERS_QUEUE_LOCATED below - it is the
# direct measurement that the queue-only pre-mode summaries (clause sites C7/C8) are false.
$cls = @{}
foreach ($id in $manifest) {
    $loc = LocOf $id
    $st  = Field $id 'status'
    $c = switch ($true) {
        ($loc -eq 'NONE')      { 'missing'; break }
        ($loc -eq 'BOTH')      { 'duplicate-torn'; break }
        ($st  -eq 'archived')  { 'pre-archived'; break }   # accepted WITHOUT a status check
        ($st  -eq 'done')      { 'matched'; break }        # declared status == expected_status
        default                { 'status-mismatch' }
    }
    $cls[$id] = $c
}
foreach ($id in ($manifest | Sort-Object)) {
    "  {0,-20} {1,-16} status={2,-10} loc={3}" -f $id, $cls[$id], (Field $id 'status'), (LocOf $id)
}
$queueLocated = @($manifest | Where-Object { (LocOf $_) -eq 'queue' }).Count
Write-Output "MEMBERS_QUEUE_LOCATED=$queueLocated"
Write-Output ("MEMBERS_ARCHIVE_LOCATED=" + @($manifest | Where-Object { (LocOf $_) -eq 'archive' }).Count)
Write-Output "C7_C8_QUEUE_ONLY_SUMMARY_FALSIFIED=$($queueLocated -lt $manifest.Count)"
$nMatched   = @($cls.Values | Where-Object { $_ -eq 'matched' }).Count
$nPreArch   = @($cls.Values | Where-Object { $_ -eq 'pre-archived' }).Count
$nMissing   = @($cls.Values | Where-Object { $_ -eq 'missing' }).Count
$nMismatch  = @($cls.Values | Where-Object { $_ -eq 'status-mismatch' }).Count
$nTorn      = @($cls.Values | Where-Object { $_ -eq 'duplicate-torn' }).Count
Write-Output "PREMODE_MATCHED=$nMatched"
Write-Output "PREMODE_PRE_ARCHIVED=$nPreArch"
Write-Output "PREMODE_MISSING=$nMissing"
Write-Output "PREMODE_STATUS_MISMATCH=$nMismatch"
Write-Output "PREMODE_DUPLICATE_TORN=$nTorn"
$premodeProceed = ($nMissing -eq 0 -and $nMismatch -eq 0 -and $nTorn -eq 0 -and ($nMatched + $nPreArch) -eq $manifest.Count)
Write-Output ("PREMODE_RESULT=" + $(if ($premodeProceed) { 'PROCEED' } else { 'RECONCILE_FAIL' }))
# The gate is: PROCEED, with EXACTLY the 11 genuinely-archived legacy members accepted as
# `pre-archived` (no status check) and the 2 completed members accepted as `matched`.
Write-Output ("CRITERION_2_PREMODE_PROCEED=" + $(if ($premodeProceed -and $nPreArch -eq 11 -and $nMatched -eq 2) { 'PASS' } else { 'FAIL' }))

# ============================================================== PASS criterion 3
Say "STEP 6: PASS CRITERION 3 - P-015 fully-covered-root classification"
$featureMembers = @($manifest | Where-Object { (Field $_ 'artifact_type') -eq 'feature' })
Write-Output ("FEATURE_MEMBERS=" + ($featureMembers -join ',') + "  n=" + $featureMembers.Count)
# Walk the FULL parent_id graph at every depth, live from queue/ + archive/ (not direct children).
$allIds = @()
foreach ($d in @('queue','archive')) {
    Get-ChildItem (Join-Path $ws ".backlogit\$d") -Filter '*.md' -ErrorAction SilentlyContinue |
        ForEach-Object { $allIds += $_.BaseName }
}
function DescendantsOf($root) {
    $acc = New-Object System.Collections.Generic.List[string]
    $frontier = @($root)
    while ($frontier.Count -gt 0) {
        $next = @()
        foreach ($p in $frontier) {
            foreach ($cand in $allIds) {
                if ((Field $cand 'parent_id') -eq $p -and -not $acc.Contains($cand)) { $acc.Add($cand); $next += $cand }
            }
        }
        $frontier = $next
    }
    return $acc
}
$isRoot = ((Field $rootF 'parent_id') -eq '(unset)' -or (Field $rootF 'parent_id') -eq '(not-found)')
$desc = DescendantsOf $rootF
Write-Output ("ROOT_HAS_NO_PARENT_ID=" + $isRoot)
Write-Output ("DESCENDANTS_AT_EVERY_DEPTH_COUNT=" + $desc.Count)
$notInManifest = @($desc | Where-Object { $manifest -notcontains $_ })
$extraInManifest = @($manifest | Where-Object { $_ -ne $rootF -and $desc -notcontains $_ })
Write-Output ("DESCENDANTS_NOT_IN_MANIFEST=" + $(if ($notInManifest.Count) { $notInManifest -join ',' } else { '(none)' }))
Write-Output ("MANIFEST_MEMBERS_NOT_DESCENDANTS=" + $(if ($extraInManifest.Count) { $extraInManifest -join ',' } else { '(none)' }))
$fullyCovered = ($notInManifest.Count -eq 0 -and $extraInManifest.Count -eq 0 -and $isRoot -and $featureMembers.Count -eq 1)
$verdict = if ($fullyCovered) { 'CASCADE' } else { 'SAFE_CLOSE' }
Write-Output "CLASSIFICATION_VERDICT=$verdict"
Write-Output ("CRITERION_3_CASCADE=" + $(if ($verdict -eq 'CASCADE') { 'PASS' } else { 'FAIL' }))

# ============================================================== pre-call snapshot
Say "STEP 7: PRE-CALL snapshot - declared status + parent_id + location + sha for every watched record"
$watch = @($manifest) + @($ctrlF, $ctrlT)
$before = @{}
foreach ($id in $watch) {
    $loc = LocOf $id
    $p = Join-Path $ws ".backlogit\$loc\$id.md"
    $before[$id] = [pscustomobject]@{
        Loc    = $loc
        Status = (Field $id 'status')
        Parent = (Field $id 'parent_id')
        Hash   = $(if (Test-Path $p) { (Get-FileHash $p -Algorithm SHA256).Hash } else { '(none)' })
    }
}
foreach ($id in ($watch | Sort-Object)) {
    "  PRE  {0,-20} loc={1,-8} status={2,-10} parent={3,-12}" -f $id, $before[$id].Loc, $before[$id].Status, $before[$id].Parent
}
# required_ids = manifest members NOT ALREADY truly archived -> the cascade MUST archive these.
# A genuinely `status: archived` member has no transition to report and is correctly absent
# from the transition log. This is why full-set equality is the WRONG test.
$requiredIds = @($manifest | Where-Object { $before[$_].Status -ne 'archived' })
$allowedIds  = @($manifest)
Write-Output ("REQUIRED_IDS=" + (($requiredIds | Sort-Object) -join ','))
Write-Output ("ALLOWED_IDS_COUNT=" + $allowedIds.Count)

# ============================================================== the authorized cascade call
Say "STEP 8: AUTHORIZED CALL - cascade ShipShipment over the root-included manifest"
$shipOut = B shipment ship $shipId --sha 2525252525252525252525252525252525252525 --message "merge: probe25 root-included cascade" --author "stage-probe@local"
$shipExit = $LASTEXITCODE
$shipOut | ForEach-Object { Write-Output "  SHIP_OUT: $_" }
Write-Output "SHIP_EXIT=$shipExit"

# ============================================================== criteria 4,5,6,7,8
Say "STEP 9: POST-CALL measurement"
$after = @{}
foreach ($id in $watch) {
    $loc = LocOf $id
    $p = Join-Path $ws ".backlogit\$loc\$id.md"
    $after[$id] = [pscustomobject]@{
        Loc    = $loc
        Status = (Field $id 'status')
        Parent = (Field $id 'parent_id')
        Hash   = $(if (Test-Path $p) { (Get-FileHash $p -Algorithm SHA256).Hash } else { '(none)' })
    }
}
foreach ($id in ($watch | Sort-Object)) {
    "  POST {0,-20} loc={1,-8} status={2,-10} parent={3,-12}" -f $id, $after[$id].Loc, $after[$id].Status, $after[$id].Parent
}

# --- criterion 4: returned_ids (items handed BACK to the queue by the close) must be empty.
$returnedIds = @($manifest | Where-Object { $before[$_].Loc -eq 'archive' -and $after[$_].Loc -eq 'queue' })
Write-Output ("RETURNED_IDS=" + $(if ($returnedIds.Count) { ($returnedIds | Sort-Object) -join ',' } else { '[]' }))
Write-Output ("CRITERION_4_RETURNED_IDS_EMPTY=" + $(if ($returnedIds.Count -eq 0) { 'PASS' } else { 'FAIL' }))

# --- criterion 5: the TWO independently-failing set equations (P-015 amendment 1.21.0).
# `allowed_ids` = manifest members PLUS the shipment record itself: the close archives its
# own shipment record, and the engine reports it in `archived_ids`. Omitting it would make
# equation A fail spuriously; omitting the ENGINE-reported log would make it pass vacuously.
$allowedIdsPlusShip = @($allowedIds) + @($shipId)
# Observed transition log, computed independently from the pre/post declared-status snapshot.
$archivedIds = @($manifest | Where-Object { $before[$_].Status -ne 'archived' -and $after[$_].Status -eq 'archived' })
# ENGINE-reported transition log, parsed from the cascade call's own JSON output. Equation A
# is evaluated over the UNION of both, so a mutation the engine reports but the snapshot
# missed (or vice versa) cannot slip through.
$engineArchived = @()
$inBlock = $false
foreach ($line in $shipOut) {
    if ($line -match '"archived_ids"\s*:\s*\[') { $inBlock = $true; continue }
    if ($inBlock) {
        if ($line -match '\]') { $inBlock = $false; continue }
        if ($line -match '"([^"]+)"') { $engineArchived += $Matches[1] }
    }
}
Write-Output ("ARCHIVED_IDS_TRANSITION_LOG_OBSERVED=" + $(if ($archivedIds.Count) { ($archivedIds | Sort-Object) -join ',' } else { '[]' }))
Write-Output ("ARCHIVED_IDS_TRANSITION_LOG_ENGINE=" + $(if ($engineArchived.Count) { ($engineArchived | Sort-Object) -join ',' } else { '[]' }))
$archivedUnion = @(($archivedIds + $engineArchived) | Sort-Object -Unique)
Write-Output ("ARCHIVED_IDS_UNION=" + ($archivedUnion -join ','))
Write-Output ("ALLOWED_IDS_INCL_SHIPMENT_COUNT=" + $allowedIdsPlusShip.Count)
$unexpected = @($archivedUnion | Where-Object { $allowedIdsPlusShip -notcontains $_ })
$notArchived = @($requiredIds | Where-Object { $archivedUnion -notcontains $_ })
Write-Output ("SET_EQ_A__archived_minus_allowed=" + $(if ($unexpected.Count) { $unexpected -join ',' } else { 'EMPTY' }))
Write-Output ("SET_EQ_B__required_minus_archived=" + $(if ($notArchived.Count) { $notArchived -join ',' } else { 'EMPTY' }))
Write-Output ("CRITERION_5A_NO_UNEXPECTED_ARCHIVE=" + $(if ($unexpected.Count -eq 0) { 'PASS' } else { 'FAIL' }))
Write-Output ("CRITERION_5B_NO_REQUIRED_LEFT_UNARCHIVED=" + $(if ($notArchived.Count -eq 0) { 'PASS' } else { 'FAIL' }))

# --- criterion 6: parent_id preserved on every surviving record.
$parentBroken = @($watch | Where-Object { $after[$_].Loc -ne 'NONE' -and $before[$_].Parent -ne $after[$_].Parent })
Write-Output ("PARENT_ID_CHANGED=" + $(if ($parentBroken.Count) { $parentBroken -join ',' } else { '(none)' }))
Write-Output ("CRITERION_6_PARENT_ID_PRESERVED=" + $(if ($parentBroken.Count -eq 0) { 'PASS' } else { 'FAIL' }))

# --- criterion 7: shipment record reaches shipped and archives with archived_status: shipped.
$shipLoc = LocOf $shipId
$shipStatus = Field $shipId 'status'
$shipArchStatus = Field $shipId 'archived_status'
Write-Output "SHIPMENT_LOC=$shipLoc"
Write-Output "SHIPMENT_STATUS=$shipStatus"
Write-Output "SHIPMENT_ARCHIVED_STATUS=$shipArchStatus"
Write-Output ("CRITERION_7_SHIPMENT_ARCHIVED_SHIPPED=" + $(if ($shipLoc -eq 'archive' -and $shipArchStatus -eq 'shipped') { 'PASS' } else { 'FAIL' }))

# --- criterion 8: nothing OUTSIDE the manifest archived, moved or deleted.
$ctrlTouched = @(@($ctrlF, $ctrlT) | Where-Object { $before[$_].Hash -ne $after[$_].Hash -or $before[$_].Loc -ne $after[$_].Loc })
Write-Output ("CONTROL_OUT_OF_MANIFEST_TOUCHED=" + $(if ($ctrlTouched.Count) { $ctrlTouched -join ',' } else { '(none)' }))
# Also prove no NEW artifact appeared and none vanished outside the manifest.
$postAll = @()
foreach ($d in @('queue','archive')) {
    Get-ChildItem (Join-Path $ws ".backlogit\$d") -Filter '*.md' -ErrorAction SilentlyContinue | ForEach-Object { $postAll += $_.BaseName }
}
$vanished = @($allIds | Where-Object { $postAll -notcontains $_ })
Write-Output ("ARTIFACTS_VANISHED=" + $(if ($vanished.Count) { $vanished -join ',' } else { '(none)' }))
Write-Output ("CRITERION_8_NOTHING_OUTSIDE_MANIFEST_TOUCHED=" + $(if ($ctrlTouched.Count -eq 0 -and $vanished.Count -eq 0) { 'PASS' } else { 'FAIL' }))

# ============================================================== isolation re-proof
Say "STEP 10: live-backlog isolation re-proof (fixture must not have mutated the real backlog)"
$liveDirtyAfter = (& git -C $repo status --porcelain -- .backlogit | Measure-Object).Count
Write-Output "LIVE_BACKLOGIT_DIRTY_PATHS_AFTER=$liveDirtyAfter"
Write-Output ("LIVE_BACKLOG_UNMUTATED=" + ($liveDirtyBefore -eq $liveDirtyAfter))

# ============================================================== verdict
Say "STEP 11: AGGREGATE VERDICT"
$criteria = [ordered]@{
    'C1_DIGEST_GATE'                      = 'PASS'
    'C2_PREMODE_PROCEED_11_PRE_ARCHIVED'  = $(if ($premodeProceed -and $nPreArch -eq 11 -and $nMatched -eq 2) { 'PASS' } else { 'FAIL' })
    'C3_CLASSIFICATION_CASCADE'           = $(if ($verdict -eq 'CASCADE') { 'PASS' } else { 'FAIL' })
    'C4_RETURNED_IDS_EMPTY'               = $(if ($returnedIds.Count -eq 0) { 'PASS' } else { 'FAIL' })
    'C5A_ARCHIVED_MINUS_ALLOWED_EMPTY'    = $(if ($unexpected.Count -eq 0) { 'PASS' } else { 'FAIL' })
    'C5B_REQUIRED_MINUS_ARCHIVED_EMPTY'   = $(if ($notArchived.Count -eq 0) { 'PASS' } else { 'FAIL' })
    'C6_PARENT_ID_PRESERVED'              = $(if ($parentBroken.Count -eq 0) { 'PASS' } else { 'FAIL' })
    'C7_SHIPMENT_ARCHIVED_SHIPPED'        = $(if ($shipLoc -eq 'archive' -and $shipArchStatus -eq 'shipped') { 'PASS' } else { 'FAIL' })
    'C8_NOTHING_OUTSIDE_MANIFEST_TOUCHED' = $(if ($ctrlTouched.Count -eq 0 -and $vanished.Count -eq 0) { 'PASS' } else { 'FAIL' })
}
foreach ($k in $criteria.Keys) { "  {0,-38} {1}" -f $k, $criteria[$k] }
$failed = @($criteria.Values | Where-Object { $_ -ne 'PASS' }).Count
Write-Output ""
Write-Output "TOPOLOGY_MEMBERS=$($manifest.Count)  ROOT_INCLUDED=True  PRE_ARCHIVED=$nPreArch  ARCHIVED_STATUS_FLAVOUR=queued"
Write-Output "FAILED_CRITERIA=$failed"
Write-Output ("PROBE25_RESULT=" + $(if ($failed -eq 0) { 'PASS' } else { 'FAIL' }))

Say "STEP 12: archived shipment provenance record"
$sp = Join-Path $ws ".backlogit\archive\$shipId.md"
if (Test-Path $sp) { Get-Content $sp -Raw }

Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
Write-Output "SCRATCH_REMOVED=True"
