# Cortex Language VSCode Extension Installer for Windows
# PowerShell script

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Cortex Language VSCode Extension Installer" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# Check if VSCode is installed
$vscodeInstalled = Get-Command code -ErrorAction SilentlyContinue
if (-not $vscodeInstalled) {
    Write-Host "[ERROR] VSCode is not installed or not in PATH." -ForegroundColor Red
    Write-Host "Please install VSCode from: https://code.visualstudio.com/" -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

# Set paths
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$extensionSrc = Join-Path $scriptDir "..\vscode-extension"
$extensionDest = Join-Path $env:USERPROFILE ".vscode\extensions\cortex-language"

Write-Host "Source: $extensionSrc"
Write-Host "Destination: $extensionDest"
Write-Host ""

# Create destination directory
if (-not (Test-Path $extensionDest)) {
    New-Item -ItemType Directory -Path $extensionDest -Force | Out-Null
}
$syntaxesDest = Join-Path $extensionDest "syntaxes"
if (-not (Test-Path $syntaxesDest)) {
    New-Item -ItemType Directory -Path $syntaxesDest -Force | Out-Null
}

# Copy extension files
Write-Host "Copying extension files..." -ForegroundColor Green

Copy-Item -Path (Join-Path $extensionSrc "package.json") -Destination $extensionDest -Force
Copy-Item -Path (Join-Path $extensionSrc "language-configuration.json") -Destination $extensionDest -Force
Copy-Item -Path (Join-Path $extensionSrc "README.md") -Destination $extensionDest -Force
Copy-Item -Path (Join-Path $extensionSrc "syntaxes\cortex.tmLanguage.json") -Destination $syntaxesDest -Force

Write-Host ""
Write-Host "[SUCCESS] Cortex language extension installed!" -ForegroundColor Green
Write-Host ""
Write-Host "To verify:" -ForegroundColor Yellow
Write-Host "  1. Open VSCode"
Write-Host "  2. Open any .cx file"
Write-Host "  3. Check that syntax highlighting is working"
Write-Host ""
Write-Host "If syntax highlighting doesn't work, press Ctrl+K M and select 'Cortex'" -ForegroundColor Yellow
Write-Host ""
Read-Host "Press Enter to exit"
