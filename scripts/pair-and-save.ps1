#Requires -Version 5.1
<#
.SYNOPSIS
    Build freebox-pair, pairing interactif Freebox, sauvegarde token.

.DESCRIPTION
    1. Build freebox-pair.exe
    2. Lance le pairing (interactif — stdin/stderr/stdout natifs)
    3. Le token s'affiche à la fin → coller dans le prompt
    4. Sauvegarde dans Windows Credential Manager
    5. Exporte FREEBOX_APP_TOKEN pour la session courante

.NOTES
    Prérequis : Go installé, exécuter depuis D:\infra\mcp-servers\freebox-mcp\
    Usage    : .\scripts\pair-and-save.ps1
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

# ── 1. Build ──────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "── Build freebox-pair ──────────────────────────────────────────" -ForegroundColor Cyan
go build -o freebox-pair.exe ./cmd/freebox-pair/
if ($LASTEXITCODE -ne 0) { throw "Build échoué." }
Write-Host "[OK] freebox-pair.exe prêt" -ForegroundColor Green

# ── 2. Pairing interactif ─────────────────────────────────────────────────────
Write-Host ""
Write-Host "── Pairing Freebox (interactif) ────────────────────────────────" -ForegroundColor Cyan
Write-Host "  Le token apparaîtra sur la dernière ligne." -ForegroundColor Yellow
Write-Host ""

$env:FREEBOX_HOST      = "mafreebox.freebox.fr"
$env:FREEBOX_APP_ID    = "mcp-freebox"
$env:FREEBOX_APP_TOKEN = ""

# Exécution native — stdin/stdout/stderr passent directement au terminal
& "$projectRoot\freebox-pair.exe"
$exitCode = $LASTEXITCODE

if ($exitCode -ne 0) {
    throw "freebox-pair a échoué (exit $exitCode)."
}

# ── 3. Saisie du token ────────────────────────────────────────────────────────
Write-Host ""
Write-Host "── Sauvegarde du token ─────────────────────────────────────────" -ForegroundColor Cyan
$token = Read-Host "  Coller le token affiché ci-dessus"
$token = $token.Trim()

if ($token.Length -lt 32) {
    throw "Token invalide (longueur $($token.Length) < 32)."
}

# ── 4. Credential Manager ─────────────────────────────────────────────────────
cmdkey /generic:freebox-mcp /user:app /pass:$token | Out-Null
if ($LASTEXITCODE -ne 0) { throw "cmdkey échoué." }
Write-Host "[OK] Token sauvegardé → Credential Manager (freebox-mcp)" -ForegroundColor Green

# ── 5. Export env session ─────────────────────────────────────────────────────
$env:FREEBOX_APP_TOKEN = $token
Write-Host "[OK] FREEBOX_APP_TOKEN actif pour cette session" -ForegroundColor Green

# ── 6. Résumé ─────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "── Prêt ────────────────────────────────────────────────────────" -ForegroundColor Cyan
Write-Host "  Pour recharger le token dans une nouvelle session :"
Write-Host '  $env:FREEBOX_APP_TOKEN = (cmdkey /list | Select-String "freebox-mcp" -Context 0,2 | ...'
Write-Host "  → voir scripts/load-token.ps1"
Write-Host ""
