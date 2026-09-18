# Git commands
Write-Host "=== git rev-parse HEAD ===" -ForegroundColor Green
git rev-parse HEAD
Write-Host ""

Write-Host "=== git log -1 --format=%H%n%an%n%ad%n%s b65b3e3 ===" -ForegroundColor Green
git log -1 --format=%H%n%an%n%ad%n%s b65b3e3
Write-Host ""

Write-Host "=== git status --porcelain ===" -ForegroundColor Green
git status --porcelain
Write-Host ""

Write-Host "=== git show --stat b65b3e3 ===" -ForegroundColor Green
git show --stat b65b3e3
Write-Host ""

# Line count checks
Write-Host "=== Line counts ===" -ForegroundColor Green
$files = @(
    "docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md",
    ".backlogit/queue/021-S.md",
    ".backlogit/queue/022-F.md",
    ".backlogit/queue/022.001-T.md",
    ".backlogit/queue/017-S.md",
    ".backlogit/queue/018-F.md",
    ".backlogit/queue/018.008-T.md",
    "docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md",
    "docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md"
)

foreach ($file in $files) {
    if (Test-Path $file) {
        $count = (Get-Content $file | Measure-Object -Line).Lines
        Write-Host "$file : $count"
    } else {
        Write-Host "$file : NOT FOUND"
    }
}
