# Probe 15 - EXACT 017-S fixture (not a 3-member representative).
# Reproduces the live 017-S topology exactly:
#   root feature (live, excluded from manifest)
#     + 1 live task  (done at closure)          -> manifest member
#     + 3 archived tasks                        -> manifest members
#     + 8 archived subtasks (3/3/2 under those) -> manifest members
#   = 12-member TASK-ONLY manifest.
# Asserts every pre-archived member is byte / path / parent_id UNCHANGED across ShipShipment.
$ErrorActionPreference = 'Continue'
$ws = Join-Path $PSScriptRoot 'p15'
function Say($m) { Write-Output "### $m" }
function B { backlogit --cwd $ws @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }

Remove-Item $ws -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $ws | Out-Null
backlogit init $ws 2>&1 | Where-Object { $_ -notmatch '^time=' }
New-Item -ItemType Directory -Force -Path (Join-Path $ws '.backlog\archive') | Out-Null

Say "STEP 1: root feature + 3 legacy tasks + 8 subtasks (3/3/2) + 1 live task"
B add --type feature --title "Covering root feature" | Out-Null            # 001-F
$legacy = @()
foreach ($n in 1..3) { B add --type task --title "Legacy task $n" --parent 001-F | Out-Null }
$subCounts = @{ '001.001-T' = 3; '001.002-T' = 3; '001.003-T' = 2 }
foreach ($t in @('001.001-T','001.002-T','001.003-T')) {
  foreach ($k in 1..$subCounts[$t]) { B add --type subtask --title "Legacy subtask $t-$k" --parent $t | Out-Null }
}
B add --type task --title "Sole live task" --parent 001-F | Out-Null       # 001.004-T
$live = '001.004-T'

Say "STEP 2: archive the 11 legacy descendants (subtasks first, then their tasks)"
$archivedMembers = @('001.001.001-ST','001.001.002-ST','001.001.003-ST',
                     '001.002.001-ST','001.002.002-ST','001.002.003-ST',
                     '001.003.001-ST','001.003.002-ST',
                     '001.001-T','001.002-T','001.003-T')
foreach ($id in $archivedMembers) { B move $id --status done | Out-Null; B archive $id | Out-Null }

Say "STEP 3: manifest = 12 members (1 live task + 11 pre-archived)"
$manifest = @($live) + $archivedMembers
B shipment create --title "Exact 017-S-shaped shipment" --items ($manifest -join ',')
B shipment get 001-S

Say "STEP 4: claim (queued -> active), then complete the sole live task"
B shipment claim 001-S | Out-Null
B move $live --status done | Out-Null
Say "parent feature state (must be live in queue, excluded from manifest, root):"
Get-Content (Join-Path $ws '.backlog\queue\001-F.md') -Raw

Say "STEP 5: PRE-CALL capture - path + SHA256 + parent_id for all 11 archived members + parent"
function Snap($ids) {
  $r = @{}
  foreach ($id in $ids) {
    foreach ($d in @('queue','archive')) {
      $p = Join-Path $ws ".backlog\$d\$id.md"
      if (Test-Path $p) {
        $c = Get-Content $p -Raw
        $par = if ($c -match '(?m)^parent_id:\s*(\S+)') { $Matches[1] } else { '(none)' }
        $r[$id] = [pscustomobject]@{ Loc=$d; Hash=(Get-FileHash $p -Algorithm SHA256).Hash; Parent=$par }
      }
    }
  }
  return $r
}
$watch = $archivedMembers + @('001-F')
$before = Snap $watch
$before.GetEnumerator() | Sort-Object Name | ForEach-Object { "PRE  {0,-18} loc={1,-8} parent={2,-12} sha={3}" -f $_.Key,$_.Value.Loc,$_.Value.Parent,$_.Value.Hash }

Say "STEP 6: AUTHORIZED CALL - ShipShipment"
B shipment ship 001-S --sha 1111111111111111111111111111111111111111 --message "merge: probe15" --author "stage-probe@local"
Write-Output "SHIP_EXIT=$LASTEXITCODE"

Say "STEP 7: POST-CALL capture + invariance comparison"
$after = Snap $watch
$fail = 0
foreach ($id in ($watch | Sort-Object)) {
  $b = $before[$id]; $a = $after[$id]
  if ($null -eq $a) { Write-Output "FAIL {0} DISAPPEARED" -f $id; $fail++; continue }
  $okB = $b.Hash -eq $a.Hash; $okP = $b.Loc -eq $a.Loc; $okR = $b.Parent -eq $a.Parent
  if ($okB -and $okP -and $okR) { "PASS {0,-18} byte=SAME path=SAME parent=SAME ({1})" -f $id,$a.Parent }
  else { "FAIL {0,-18} byte={1} path={2}->{3} parent={4}->{5}" -f $id,$okB,$b.Loc,$a.Loc,$b.Parent,$a.Parent; $fail++ }
}
Write-Output "INVARIANCE_FAILURES=$fail"

Say "STEP 8: archived shipment provenance + live task lineage"
Get-Content (Join-Path $ws '.backlog\archive\001-S.md') -Raw
Get-Content (Join-Path $ws ".backlog\archive\$live.md") -Raw
