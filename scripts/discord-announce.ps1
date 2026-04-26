#Requires -Version 5.1
<#
.SYNOPSIS
    Génère le message Discord pour la dernière release et le copie dans le presse-papier.
.USAGE
    .\scripts\discord-announce.ps1
    .\scripts\discord-announce.ps1 -Version v0.7.1
#>
param(
    [string]$Version = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot

# Dernière version = dernier tag git si non spécifié
if ($Version -eq "") {
    $Version = (git -C $root describe --tags --abbrev=0 2>$null).Trim()
    if (-not $Version) { throw "Aucun tag git trouvé." }
}

$v = $Version.TrimStart('v')
$releaseUrl = "https://github.com/mipsou/mcp-freebox/releases/tag/v$v"

# Titre depuis le message du tag git
$tagTitle = (git -C $root tag -l --format='%(contents:subject)' "v$v" 2>$null).Trim()
if (-not $tagTitle) { $tagTitle = "v$v" }

# Extraire la section du CHANGELOG
$changelog = Get-Content "$root\CHANGELOG.md" -Raw
$section = ""
if ($changelog -match "(?s)## \[$v\][^\n]*\n(.*?)(?=\n## \[|\z)") {
    $section = $Matches[1].Trim()
}

# Reformater pour Discord : ### Ajouté → **Ajouté**, lignes conservées intégralement
$lines = $section -split "`n" | ForEach-Object {
    $line = $_.TrimEnd()
    if ($line -match '^### (.+)') {
        "`n**$($Matches[1])**"
    } elseif ($line -ne "" -and $line -ne "---") {
        $line
    }
} | Where-Object { $_ -ne $null }

$body = ($lines | Where-Object { $_ -ne "" }) -join "`n"
$body = $body -replace "`n{3,}", "`n`n"

# Construire le message final
$msg = "🚀 **freebox-mcp v$v** — $tagTitle`n$releaseUrl`n$body"
$msg = $msg.Trim()

# Copier dans le presse-papier
Set-Clipboard -Value $msg

Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host $msg
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""
Write-Host "[OK] Copié dans le presse-papier — coller directement dans Discord." -ForegroundColor Green
