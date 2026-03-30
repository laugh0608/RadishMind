[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$textExtensions = @(".md", ".txt", ".json", ".yml", ".yaml", ".ps1", ".sh")
$textNames = @(".gitignore", ".gitattributes", ".editorconfig")
$utf8 = [System.Text.UTF8Encoding]::new($false, $true)
$violations = New-Object System.Collections.Generic.List[string]

function Test-IsTextFile {
    param([string]$RelativePath)

    $leaf = Split-Path $RelativePath -Leaf
    if ($textNames -contains $leaf) {
        return $true
    }

    $extension = [System.IO.Path]::GetExtension($RelativePath).ToLowerInvariant()
    return $textExtensions -contains $extension
}

$trackedFiles = & git -C $repoRoot ls-files --cached --others --exclude-standard
if ($LASTEXITCODE -ne 0) {
    throw "command failed: git -C $repoRoot ls-files --cached --others --exclude-standard"
}

foreach ($relativePath in $trackedFiles) {
    if (-not (Test-IsTextFile -RelativePath $relativePath)) {
        continue
    }

    $fullPath = Join-Path $repoRoot $relativePath
    if (-not (Test-Path -LiteralPath $fullPath -PathType Leaf)) {
        continue
    }

    $bytes = [System.IO.File]::ReadAllBytes($fullPath)

    if (
        $bytes.Length -ge 3 -and
        $bytes[0] -eq 0xEF -and
        $bytes[1] -eq 0xBB -and
        $bytes[2] -eq 0xBF
    ) {
        $violations.Add(("{0}: contains UTF-8 BOM" -f $relativePath))
    }

    try {
        $text = $utf8.GetString($bytes)
    }
    catch {
        $violations.Add(("{0}: is not valid UTF-8" -f $relativePath))
        continue
    }

    if ($text.Contains("`r")) {
        $violations.Add(("{0}: contains CRLF or carriage returns" -f $relativePath))
    }

    if ($bytes.Length -gt 0 -and $bytes[$bytes.Length - 1] -ne 10) {
        $violations.Add(("{0}: is missing a final newline" -f $relativePath))
    }

    $extension = [System.IO.Path]::GetExtension($relativePath).ToLowerInvariant()
    if ($extension -eq ".md") {
        continue
    }

    $lines = $text -split "`n"
    for ($index = 0; $index -lt $lines.Count; $index++) {
        $line = $lines[$index]

        if ($index -eq $lines.Count - 1 -and $line.Length -eq 0) {
            continue
        }

        if ($line -match "[ \t]+$") {
            $lineNumber = $index + 1
            $violations.Add(("{0}:{1}: trailing whitespace" -f $relativePath, $lineNumber))
        }
    }
}

if ($violations.Count -gt 0) {
    $violations | ForEach-Object { Write-Error $_ }
    exit 1
}

Write-Host "text file hygiene checks passed."
