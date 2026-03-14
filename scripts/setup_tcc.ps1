# Create tools/ and help set up TinyCC so Cortex can compile without gcc/cmake.
# Run from project root. After this, use: cortex -i main.cx -o main.exe (backend=auto will use tcc).
$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$ToolsDir = Join-Path $ProjectRoot "tools"
$TccExe = Join-Path $ToolsDir "tcc.exe"

if (-not (Test-Path $ToolsDir)) {
    New-Item -ItemType Directory -Path $ToolsDir | Out-Null
    Write-Host "Created $ToolsDir"
}

if (Test-Path $TccExe) {
    Write-Host "tcc.exe already present at $TccExe"
    & $TccExe -version 2>$null; if ($LASTEXITCODE -ne 0) { & $TccExe 2>&1 | Select-Object -First 1 }
    exit 0
}

# Try common locations for a system tcc
$inPath = Get-Command tcc -ErrorAction SilentlyContinue
if ($inPath) {
    Write-Host "tcc found in PATH: $($inPath.Source)"
    Write-Host "Cortex will use it automatically with -backend auto"
    exit 0
}

Write-Host ""
Write-Host "TinyCC (tcc) not found. To compile without gcc/cmake:"
Write-Host "  1. Download a Windows build of Tiny C Compiler (tcc)."
Write-Host "     Options:"
Write-Host "       - Prebuilt (if available): https://github.com/nickhutchinson/tinycc/releases"
Write-Host "       - Or build from source: https://repo.or.cz/tinycc.git"
Write-Host "  2. Save the executable as: $TccExe"
Write-Host "  3. Run cortex as usual; it will use tcc when backend=auto (default)."
Write-Host ""
Write-Host "Alternatively install gcc (e.g. MinGW) and use: cortex -backend gcc -i main.cx -o main.exe"
Write-Host ""
