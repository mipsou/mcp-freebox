#Requires -Version 5.1
<#
.SYNOPSIS
    Charge le token freebox-mcp depuis Windows Credential Manager.

.DESCRIPTION
    Lit le token stocké par pair-and-save.ps1 et l'exporte comme
    FREEBOX_APP_TOKEN pour la session PowerShell courante.

.NOTES
    Usage : . .\scripts\load-token.ps1  (dot-source pour exporter la variable)
#>

$cred = Get-StoredCredential -Target "freebox-mcp" -ErrorAction SilentlyContinue

if (-not $cred) {
    # Fallback via cmdkey si CredentialManager module absent
    $credLine = cmdkey /list:freebox-mcp 2>&1 | Where-Object { $_ -match "freebox-mcp" }
    if (-not $credLine) {
        throw "Token introuvable. Lancer scripts\pair-and-save.ps1 d'abord."
    }
    # cmdkey ne permet pas de relire le mot de passe — utiliser le module
    throw "Module CredentialManager requis : Install-Module CredentialManager -Scope CurrentUser"
}

$env:FREEBOX_APP_TOKEN = $cred.GetNetworkCredential().Password
Write-Host "[OK] FREEBOX_APP_TOKEN chargé depuis Credential Manager" -ForegroundColor Green
