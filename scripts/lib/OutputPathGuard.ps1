function Normalize-GuardPath {
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $pathRoot = [System.IO.Path]::GetPathRoot($fullPath)
    if ($fullPath.Length -gt $pathRoot.Length) {
        return $fullPath.TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    }

    return $fullPath
}

function Resolve-ReparseAwarePath {
    param(
        [Parameter(Mandatory)]
        [string]$Path
    )

    $absolutePath = Normalize-GuardPath -Path $Path
    $pathRoot = [System.IO.Path]::GetPathRoot($absolutePath)
    if ([string]::IsNullOrEmpty($pathRoot)) {
        throw "Unable to determine path root for '$absolutePath'."
    }

    $relativePath = $absolutePath.Substring($pathRoot.Length)
    $parts = @()
    if (-not [string]::IsNullOrEmpty($relativePath)) {
        $parts = @($relativePath -split '[\\/]' | Where-Object { $_ })
    }

    $currentPath = $pathRoot
    for ($i = 0; $i -lt $parts.Count; $i++) {
        $candidatePath = Join-Path $currentPath $parts[$i]
        if (-not (Test-Path -LiteralPath $candidatePath)) {
            $resolvedPath = $currentPath
            for ($j = $i; $j -lt $parts.Count; $j++) {
                $resolvedPath = Join-Path $resolvedPath $parts[$j]
            }
            return Normalize-GuardPath -Path $resolvedPath
        }

        $item = Get-Item -LiteralPath $candidatePath -Force
        if ($item.LinkType -eq 'SymbolicLink' -or $item.LinkType -eq 'Junction') {
            $target = $item.ResolveLinkTarget($true)
            if ($null -eq $target) {
                throw "Unable to resolve reparse point '$candidatePath'."
            }
            $currentPath = Normalize-GuardPath -Path $target.FullName
            continue
        }

        $currentPath = Normalize-GuardPath -Path $item.FullName
    }

    return $currentPath
}

function Resolve-OutputPathWithinRoot {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)]
        [string]$RepoRoot,

        [Parameter(Mandatory)]
        [string]$OutputDir,

        [string]$BasePath = (Get-Location).Path
    )

    $repoRoot = Normalize-GuardPath -Path $RepoRoot

    # GetFullPath resolves '..' segments and relative paths without requiring the
    # directory to already exist.
    $resolvedOutputDir = Normalize-GuardPath -Path ([System.IO.Path]::GetFullPath($OutputDir, $BasePath))

    # TOCTOU residual risk: validation resolves the current filesystem state now,
    # but the later directory creation and artifact writes happen at a different
    # time. A concurrent process could still swap an ancestor reparse point
    # between the check and the write. This narrows the lexical bypass; it is not
    # an atomic guarantee.
    $resolvedRepoRoot = Resolve-ReparseAwarePath -Path $repoRoot
    $resolvedOutputTarget = Resolve-ReparseAwarePath -Path $resolvedOutputDir

    # Compare ordinally (case-sensitive) rather than case-insensitively: this
    # guard is a security/safety boundary (invariant I5), so it must fail closed
    # (refuse) on an ambiguous case mismatch rather than risk a false negative on
    # a case-sensitive filesystem (Linux, or macOS in case-sensitive mode).
    $repoRootWithSep = $resolvedRepoRoot + [System.IO.Path]::DirectorySeparatorChar
    $isDescendant = $resolvedOutputTarget.Equals($resolvedRepoRoot, [System.StringComparison]::Ordinal) -or
        $resolvedOutputTarget.StartsWith($repoRootWithSep, [System.StringComparison]::Ordinal)

    if (-not $isDescendant) {
        throw "Refusing to build: resolved output path '$resolvedOutputTarget' is not a descendant of the repository root '$resolvedRepoRoot' (invariant I5)."
    }

    return $resolvedOutputDir
}
