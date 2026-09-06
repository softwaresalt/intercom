#Requires -Version 7.2
<#
.SYNOPSIS
    Canonical cross-compile build script for intercom-go (001.001.004-ST).

.DESCRIPTION
    Builds every binary under cmd/ for every target declared in
    scripts/targets.json (the single source of truth for the target matrix -
    the CI cross-compile matrix job, 001.002.004-ST, reads the same file so
    local and CI builds cannot drift).

    Invariant I5: this script never writes or deletes outside the repository
    root, and never performs a recursive delete. Cleanup, if needed, goes
    through 'go clean' scoped to the module instead.

.PARAMETER OutputDir
    Directory to write build artifacts into. Defaults to '<repo-root>/dist'.
    Resolved relative to the current working directory when relative, and
    validated to be a descendant of the repository root before any file is
    written.
#>
[CmdletBinding()]
param(
    [string]$OutputDir
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

# Resolve the repository root from this script's own location, never from
# the caller's current working directory.
$scriptRoot = $PSScriptRoot
. (Join-Path (Join-Path $scriptRoot 'lib') 'OutputPathGuard.ps1')
$repoRoot = (Resolve-Path (Join-Path $scriptRoot '..')).Path.TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)

if (-not $OutputDir) {
    $OutputDir = Join-Path $repoRoot 'dist'
}

$resolvedOutputDir = Resolve-OutputPathWithinRoot -RepoRoot $repoRoot -OutputDir $OutputDir -BasePath (Get-Location).Path

# Create/overwrite only - never delete anything, recursively or otherwise.
New-Item -ItemType Directory -Path $resolvedOutputDir -Force | Out-Null

$targetsPath = Join-Path $scriptRoot 'targets.json'
$manifest = Get-Content -Raw -Path $targetsPath | ConvertFrom-Json
$targets = $manifest.targets

$cmdRoot = Join-Path $repoRoot 'cmd'
$binaries = Get-ChildItem -Path $cmdRoot -Directory | Select-Object -ExpandProperty Name | Sort-Object

if ($binaries.Count -eq 0) {
    Write-Error "No binaries found under '$cmdRoot'."
    exit 1
}

Write-Host "Building $($binaries.Count) binaries x $($targets.Count) targets into $resolvedOutputDir"

$builtCount = 0
foreach ($target in $targets) {
    foreach ($bin in $binaries) {
        $ext = if ($target.goos -eq 'windows') { '.exe' } else { '' }
        $outName = "$bin-$($target.goos)-$($target.goarch)$ext"
        $outPath = Join-Path $resolvedOutputDir $outName

        Write-Host "  -> $outName"

        $env:GOOS = $target.goos
        $env:GOARCH = $target.goarch
        $env:CGO_ENABLED = '0'

        & go build -mod=readonly -o $outPath "./cmd/$bin"
        $buildExitCode = $LASTEXITCODE

        Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
        Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue

        if ($buildExitCode -ne 0) {
            throw "go build failed for $outName (exit $buildExitCode)"
        }

        $builtCount++
    }
}

Write-Host "Built $builtCount artifacts into $resolvedOutputDir"
