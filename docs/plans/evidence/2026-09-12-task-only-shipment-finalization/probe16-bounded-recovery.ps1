# Probe 16 - injected partial failure + BOUNDED recovery (not atomic rollback).
# Recovery path is WORKSPACE-CONTAINED and GITIGNORED: .autoharness/backups/stage-recovery/
# (outside the compared .backlog queue/archive/log inventory).
# Class discipline under test:
#   mutable current-state markdown (queue/*.md, archive/*.md) -> MAY be restored from snapshot
#   append-only logs/events (logs/, hooks_queue.jsonl, stash.jsonl) -> NEVER rewound/deleted; get a recovery event
#   disposable cache (backlogit.db/-wal/-shm) -> NEVER byte-restored; official `backlogit sync` rehydrates
$ErrorActionPreference = 'Continue'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$ws   = Join-Path $PSScriptRoot 'p16'
$rec  = Join-Path $repo '.autoharness\backups\stage-recovery\p16'
$snap = Join-Path $rec 'snapshot'
$quar = Join-Path $rec 'quarantine'
function Say($m) { Write-Output "### $m" }
function B { backlogit --cwd $ws @args 2>&1 | Where-Object { $_ -notmatch '^time=' } }

Remove-Item $ws   -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item $rec  -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force -Path $ws,$snap,$quar | Out-Null
backlogit init $ws 2>&1 | Where-Object { $_ -notmatch '^time=' }
New-Item -ItemType Directory -Force -Path (Join-Path $ws '.backlog\archive') | Out-Null
Write-Output "RECOVERY_PATH=$($rec.Replace($repo,''))"
$recRel = $rec.Replace($repo,'').TrimStart('\')
git -C $repo check-ignore -q $recRel 2>$null
$recIgnored = ($LASTEXITCODE -eq 0)
Write-Output "RECOVERY_PATH_GITIGNORED=$recIgnored"

Say "STEP 1: minimal authorized topology (root feature + 1 live task), shipment claimed, task done"
B add --type feature --title "Covering root feature" | Out-Null
B add --type task --title "Sole live task" --parent 001-F | Out-Null
B shipment create --title "Recovery probe shipment" --items 001.001-T | Out-Null
B shipment claim 001-S | Out-Null
B move 001.001-T --status done | Out-Null

Say "STEP 2: PRE-CALL snapshot of MUTABLE source-of-truth markdown only + inventory manifest"
$mutable = Get-ChildItem (Join-Path $ws '.backlog\queue'),(Join-Path $ws '.backlog\archive') -Filter *.md -File -ErrorAction SilentlyContinue
$inv = foreach ($f in $mutable) {
  $rel = $f.FullName.Substring($ws.Length).TrimStart('\')
  $dst = Join-Path $snap $rel
  New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
  Copy-Item $f.FullName $dst -Force
  [pscustomobject]@{ rel=$rel; sha256=(Get-FileHash $f.FullName -Algorithm SHA256).Hash }
}
$inv | ConvertTo-Json -Depth 3 | Set-Content (Join-Path $rec 'inventory.json')
$inv | ForEach-Object { "INVENTORY {0,-34} {1}" -f $_.rel,$_.sha256 }

Say "STEP 2b: append-only + disposable baselines (recorded, NEVER restored)"
$logDir = Join-Path $ws '.backlog\logs'
$appendOnly = @{}
foreach ($p in @((Join-Path $ws '.backlog\hooks_queue.jsonl'),(Join-Path $ws '.backlog\stash.jsonl'))) {
  if (Test-Path $p) { $appendOnly[$p] = (Get-Item $p).Length }
}
if (Test-Path $logDir) { foreach ($f in Get-ChildItem $logDir -File -Recurse) { $appendOnly[$f.FullName] = $f.Length } }
$appendOnly.GetEnumerator() | ForEach-Object { "APPEND_ONLY_BASELINE {0,-28} bytes={1}" -f (Split-Path $_.Key -Leaf),$_.Value }
$db = Join-Path $ws '.backlog\backlogit.db'
Write-Output "DISPOSABLE_CACHE backlogit.db present=$(Test-Path $db) (never byte-restored)"

Say "STEP 3: INJECT partial failure - occupy the shipment record's archive destination with a directory"
New-Item -ItemType Directory -Force -Path (Join-Path $ws '.backlog\archive\001-S.md') | Out-Null
Say "STEP 4: AUTHORIZED CALL (expected to fail AFTER partial mutation)"
B shipment ship 001-S --sha 2222222222222222222222222222222222222222 --message "merge: probe16" --author "stage-probe@local"
$shipExit = $LASTEXITCODE
Write-Output "SHIP_EXIT=$shipExit"

Say "STEP 5: observe TORN state"
foreach ($id in @('001-S','001.001-T')) {
  foreach ($d in @('queue','archive')) {
    $p = Join-Path $ws ".backlog\$d\$id.md"
    if (Test-Path $p) {
      $isDir = (Get-Item $p).PSIsContainer
      $st = if ($isDir) { '<DIRECTORY>' } else { (Select-String -Path $p -Pattern '^status:' | Select-Object -First 1).Line }
      "TORN_SCAN {0,-14} {1,-8} {2}" -f $id,$d,$st
    }
  }
}

Say "STEP 6: APPROVAL GATE - destructive quarantine/restore requires explicit approval FIRST"
$APPROVAL = $true   # probe-mode explicit approval; live path requires operator action-risk approval
Write-Output "OPERATOR_ACTION_RISK_APPROVAL=$APPROVAL (gate evaluated BEFORE any destructive move)"
if (-not $APPROVAL) { Write-Output "HALT: approval withheld - no quarantine, no restore"; exit 1 }

Say "STEP 7: RECOVERY (a) enumerate, (b) quarantine unexpected paths - move, never delete"
$now = Get-ChildItem (Join-Path $ws '.backlog\queue'),(Join-Path $ws '.backlog\archive') -Force -ErrorAction SilentlyContinue
$invRel = $inv.rel
foreach ($f in $now) {
  $rel = $f.FullName.Substring($ws.Length).TrimStart('\')
  if ($invRel -notcontains $rel) {
    $dst = Join-Path $quar $rel
    New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
    Move-Item $f.FullName $dst -Force
    "QUARANTINED {0}" -f $rel
  }
}

Say "STEP 8: RECOVERY (c) restore ONLY enumerated mutable markdown from snapshot"
foreach ($e in $inv) {
  $dst = Join-Path $ws $e.rel
  New-Item -ItemType Directory -Force -Path (Split-Path $dst) | Out-Null
  Copy-Item (Join-Path $snap $e.rel) $dst -Force
  "RESTORED {0}" -f $e.rel
}

Say "STEP 9: RECOVERY (d) rehydrate disposable index via OFFICIAL sync (no DB/WAL byte restore)"
B sync

Say "STEP 10: VERIFY byte/path equivalence for mutable state"
$fail = 0
foreach ($e in $inv) {
  $p = Join-Path $ws $e.rel
  if (-not (Test-Path $p)) { "FAIL missing {0}" -f $e.rel; $fail++; continue }
  $h = (Get-FileHash $p -Algorithm SHA256).Hash
  if ($h -eq $e.sha256) { "PASS {0,-34} byte=SAME" -f $e.rel } else { "FAIL {0,-34} byte=DIFF" -f $e.rel; $fail++ }
}
$residual = Get-ChildItem (Join-Path $ws '.backlog\queue'),(Join-Path $ws '.backlog\archive') -Force -ErrorAction SilentlyContinue |
  ForEach-Object { $_.FullName.Substring($ws.Length).TrimStart('\') } | Where-Object { $invRel -notcontains $_ }
Write-Output "RESIDUAL_UNEXPECTED_PATHS=$(@($residual).Count)"
Write-Output "ENGINE_VIEW_AFTER_RECOVERY:"
B shipment get 001-S

Say "STEP 11: VERIFY append-only streams were NOT rewound (length must be >= baseline)"
$rewound = 0
foreach ($kv in $appendOnly.GetEnumerator()) {
  $cur = if (Test-Path $kv.Key) { (Get-Item $kv.Key).Length } else { -1 }
  $ok = $cur -ge $kv.Value
  if (-not $ok) { $rewound++ }
  "APPEND_ONLY {0,-28} baseline={1} current={2} not_rewound={3}" -f (Split-Path $kv.Key -Leaf),$kv.Value,$cur,$ok
}
Write-Output "APPEND_ONLY_REWIND_VIOLATIONS=$rewound"

Say "STEP 12: append a RECOVERY EVENT to the append-only stream (never a rewind)"
$evt = [pscustomobject]@{ ts=(Get-Date).ToUniversalTime().ToString('o'); event='stage_bounded_recovery'; probe='16'; shipment='001-S'; ship_exit=$shipExit; restored=$inv.Count; quarantine=$quar.Replace($repo,''); note='bounded recovery; NOT atomic rollback' } | ConvertTo-Json -Compress
Add-Content -Path (Join-Path $ws '.backlog\stash.jsonl') -Value $evt
Write-Output "RECOVERY_EVENT_APPENDED=$evt"

Write-Output "MUTABLE_EQUIVALENCE_FAILURES=$fail"
Write-Output "QUARANTINE_PRESERVED=$((Get-ChildItem $quar -Recurse -Force -ErrorAction SilentlyContinue | Measure-Object).Count) entries at $($quar.Replace($repo,''))"
Write-Output "TERMINAL_DISPOSITION=HALT (bounded recovery complete; diagnostic evidence preserved)"
