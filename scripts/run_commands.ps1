# Script to run all commands sequentially

$commands = @(
    @{num=1; cmd='backlogit query "SELECT id,artifact_type,status,parent_id FROM items WHERE id LIKE ''019%'' OR id LIKE ''020%'' OR id LIKE ''021%'' OR id LIKE ''023%'' OR id LIKE ''024%'' OR id LIKE ''025%'' ORDER BY id"'},
    @{num=2; cmd='backlogit get 025-F'},
    @{num=3; cmd='backlogit get 025.001-T'},
    @{num=4; cmd='backlogit dep list 025.001-T'},
    @{num=5; cmd='backlogit query "SELECT item_id,depends_on,dep_type FROM item_deps WHERE depends_on=''025.001-T''"'},
    @{num=6; cmd='backlogit query "SELECT item_id,depends_on,dep_type FROM item_deps WHERE item_id LIKE ''019%'' OR item_id LIKE ''025%''"'},
    @{num=7; cmd='backlogit shipment list'},
    @{num=8; cmd='Get-ChildItem .backlogit/queue -Name'},
    @{num=9; cmd='Get-ChildItem .backlogit/archive -Name | Select-String -Pattern ''019|025|018'''},
    @{num=10; cmd='backlogit get 021-F'},
    @{num=11; cmd='backlogit get 021.001-T'},
    @{num=12; cmd='git status --porcelain'},
    @{num=13; cmd='git log --oneline -5'},
    @{num=14; cmd='git symbolic-ref --short HEAD'},
    @{num=15; cmd='Select-String -Path .backlogit/stash.jsonl -Pattern ''DB12DA37'''}
)

foreach ($item in $commands) {
    Write-Host "[COMMAND $($item.num)]"
    Write-Host $item.cmd
    Write-Host "[OUTPUT]"
    Invoke-Expression $item.cmd 2>&1 | ForEach-Object { Write-Host $_ }
    Write-Host ""
}

# Command 16 - conditional
Write-Host "[COMMAND 16]"
Write-Host "If file .backlogit/queue/025.001-T.md exists, run: Get-Content .backlogit/queue/025.001-T.md -Raw; otherwise search .backlogit for it"
Write-Host "[OUTPUT]"
if (Test-Path ".backlogit/queue/025.001-T.md") {
    Get-Content ".backlogit/queue/025.001-T.md" -Raw
} else {
    $found = Get-ChildItem -Path ".backlogit" -Filter "025.001-T.md" -Recurse -ErrorAction SilentlyContinue
    if ($found) {
        Get-Content $found.FullName -Raw
    } else {
        Write-Host "File 025.001-T.md not found in .backlogit"
    }
}
Write-Host ""

# Command 17
Write-Host "[COMMAND 17]"
Write-Host "Get-Content .backlogit/archive/018.008-T.md -Raw | head -n 30"
Write-Host "[OUTPUT]"
if (Test-Path ".backlogit/archive/018.008-T.md") {
    Get-Content ".backlogit/archive/018.008-T.md" -Raw | Select-Object -First 30
} else {
    Write-Host "File .backlogit/archive/018.008-T.md not found"
}
Write-Host ""

# Command 18
Write-Host "[COMMAND 18]"
Write-Host "backlogit --help"
Write-Host "[OUTPUT]"
backlogit --help 2>&1
Write-Host ""
