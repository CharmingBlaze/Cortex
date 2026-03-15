# Cortex Cross-Platform Release Builder with Zig
# Bundles Zig CC for all platforms

param(
    [string]$Version = "0.1.0",
    [string]$ZigVersion = "0.13.0"
)

$ErrorActionPreference = "Stop"
$DistDir = "dist"
$BuildDir = "build"

Write-Host "=== Cortex Release v$Version with Zig $ZigVersion ===" -ForegroundColor Cyan

# Zig download URLs
$ZigUrls = @{
    "windows-amd64" = "https://ziglang.org/download/$ZigVersion/zig-x86_64-windows-$ZigVersion.zip"
    "windows-arm64" = "https://ziglang.org/download/$ZigVersion/zig-aarch64-windows-$ZigVersion.zip"
    "linux-amd64" = "https://ziglang.org/download/$ZigVersion/zig-x86_64-linux-$ZigVersion.tar.xz"
    "linux-arm64" = "https://ziglang.org/download/$ZigVersion/zig-aarch64-linux-$ZigVersion.tar.xz"
    "darwin-amd64" = "https://ziglang.org/download/$ZigVersion/zig-x86_64-macos-$ZigVersion.tar.xz"
    "darwin-arm64" = "https://ziglang.org/download/$ZigVersion/zig-aarch64-macos-$ZigVersion.tar.xz"
}

# Build targets
$Targets = @(
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe"; ZigDir = "zig-x86_64-windows-$ZigVersion" }
    @{ OS = "windows"; Arch = "arm64"; Ext = ".exe"; ZigDir = "zig-aarch64-windows-$ZigVersion" }
    @{ OS = "linux"; Arch = "amd64"; Ext = ""; ZigDir = "zig-x86_64-linux-$ZigVersion" }
    @{ OS = "linux"; Arch = "arm64"; Ext = ""; ZigDir = "zig-aarch64-linux-$ZigVersion" }
    @{ OS = "darwin"; Arch = "amd64"; Ext = ""; ZigDir = "zig-x86_64-macos-$ZigVersion" }
    @{ OS = "darwin"; Arch = "arm64"; Ext = ""; ZigDir = "zig-aarch64-macos-$ZigVersion" }
)

# Clean previous builds
if (Test-Path $DistDir) { Remove-Item -Recurse -Force $DistDir }
if (Test-Path $BuildDir) { Remove-Item -Recurse -Force $BuildDir }
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

# Download Zig for all platforms
Write-Host "`nDownloading Zig binaries..." -ForegroundColor Yellow
foreach ($target in $Targets) {
    $key = "$($target.OS)-$($target.Arch)"
    $url = $ZigUrls[$key]
    $zigFile = "$BuildDir\zig-$key"
    
    if ($target.OS -eq "windows") {
        $zigFile += ".zip"
    } else {
        $zigFile += ".tar.xz"
    }
    
    Write-Host "  Downloading Zig for $key..." -ForegroundColor Gray
    Invoke-WebRequest -Uri $url -OutFile $zigFile -UseBasicParsing
}

# Build Cortex for each target
Write-Host "`nBuilding Cortex for all platforms..." -ForegroundColor Yellow
foreach ($target in $Targets) {
    $key = "$($target.OS)-$($target.Arch)"
    $output = "$BuildDir\cortex-$key$($target.Ext)"
    
    Write-Host "  Building for $key..." -ForegroundColor Gray
    
    $Env:GOOS = $target.OS
    $Env:GOARCH = $target.Arch
    go build -ldflags="-s -w -X main.Version=$Version" -o $output ./cmd/cortex
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  Failed to build for $key" -ForegroundColor Red
        continue
    }
}

# Create release packages
Write-Host "`nCreating release packages..." -ForegroundColor Yellow
foreach ($target in $Targets) {
    $key = "$($target.OS)-$($target.Arch)"
    $pkgDir = "$DistDir\cortex-$Version-$key"
    $zigDir = $target.ZigDir
    
    Write-Host "  Packaging $key..." -ForegroundColor Gray
    
    # Create directory structure
    New-Item -ItemType Directory -Force -Path "$pkgDir\bin" | Out-Null
    New-Item -ItemType Directory -Force -Path "$pkgDir\zig" | Out-Null
    New-Item -ItemType Directory -Force -Path "$pkgDir\runtime" | Out-Null
    New-Item -ItemType Directory -Force -Path "$pkgDir\examples" | Out-Null
    New-Item -ItemType Directory -Force -Path "$pkgDir\docs" | Out-Null
    
    # Copy Cortex binary
    Copy-Item "$BuildDir\cortex-$key$($target.Ext)" "$pkgDir\bin\"
    
    # Extract Zig
    $zigArchive = "$BuildDir\zig-$key"
    if ($target.OS -eq "windows") {
        $zigArchive += ".zip"
        Expand-Archive -Path $zigArchive -DestinationPath "$BuildDir\zig-extract-$key" -Force
        Move-Item "$BuildDir\zig-extract-$key\$zigDir\*" "$pkgDir\zig\" -Force
    } else {
        $zigArchive += ".tar.xz"
        # Extract tar.xz (requires 7-zip or tar on Windows 10+)
        tar -xf $zigArchive -C "$BuildDir"
        Move-Item "$BuildDir\$zigDir\*" "$pkgDir\zig\" -Force
    }
    
    # Copy runtime files
    Copy-Item runtime\*.c "$pkgDir\runtime\" -ErrorAction SilentlyContinue
    Copy-Item runtime\*.h "$pkgDir\runtime\" -ErrorAction SilentlyContinue
    
    # Copy examples
    Copy-Item examples\*.cx "$pkgDir\examples\" -ErrorAction SilentlyContinue
    
    # Copy docs
    Copy-Item README.md "$pkgDir\"
    Copy-Item LANGUAGE_SPEC.md "$pkgDir\docs\" -ErrorAction SilentlyContinue
    Copy-Item LANGUAGE_GUIDE.md "$pkgDir\docs\" -ErrorAction SilentlyContinue
    Copy-Item CHANGELOG.md "$pkgDir\docs\" -ErrorAction SilentlyContinue
    Copy-Item -Recurse docs\* "$pkgDir\docs\" -ErrorAction SilentlyContinue
    
    # Create install script
    if ($target.OS -eq "windows") {
        $installScript = @"
@echo off
echo Installing Cortex v$Version...
setlocal
set "CORT_EXE=%~dp0bin\cortex.exe"
set "ZIG_EXE=%~dp0zig\zig.exe"
setx CORT_EXE "%CORT_EXE%" >nul 2>&1
setx ZIG_EXE "%ZIG_EXE%" >nul 2>&1
echo.
echo Cortex installed successfully!
echo.
echo Add to PATH manually or restart your terminal:
echo   %~dp0bin
echo   %~dp0zig
echo.
echo Run: cortex run examples/hello.cx
"@
        Set-Content -Path "$pkgDir\install.bat" -Value $installScript
        
        # Create quickstart script
        $quickstartScript = @"
@echo off
echo Running Cortex Quick Start...
call "%~dp0bin\cortex.exe" run "%~dp0examples\hello.cx"
"@
        Set-Content -Path "$pkgDir\quickstart.bat" -Value $quickstartScript
    } else {
        $installScript = @"
#!/bin/bash
echo "Installing Cortex v$Version..."
CORT_DIR="$(cd "$(dirname "$0")" && pwd)"
echo ""
echo "Add to PATH:"
echo "  export PATH=\"\$CORT_DIR/bin:\$CORT_DIR/zig:\$PATH\""
echo ""
echo "Add to ~/.bashrc or ~/.zshrc for persistence."
echo ""
echo "Run: cortex run examples/hello.cx"
"@
        Set-Content -Path "$pkgDir\install.sh" -Value $installScript
    }
    
    # Create README
    $readme = @"
# Cortex v$Version for $($target.OS) $($target.Arch)

## Quick Start

### Windows
1. Run `install.bat`
2. Restart your terminal
3. Run: `cortex run examples/hello.cx`

### Linux/macOS
1. Run: `source install.sh` or add bin/ and zig/ to PATH
2. Run: `cortex run examples/hello.cx`

## What's Included

- `bin/cortex` - The Cortex compiler
- `zig/` - Zig CC (bundled C compiler, no external dependencies)
- `runtime/` - Cortex runtime libraries
- `examples/` - Example programs
- `docs/` - Documentation

## No External Dependencies

Everything you need is bundled. No Go, no GCC, no external C compiler required.
Zig CC handles all C compilation internally.
"@
    Set-Content -Path "$pkgDir\README.txt" -Value $readme
}

# Create archives
Write-Host "`nCreating release archives..." -ForegroundColor Yellow
foreach ($target in $Targets) {
    $key = "$($target.OS)-$($target.Arch)"
    $pkgDir = "$DistDir\cortex-$Version-$key"
    
    if ($target.OS -eq "windows") {
        $archive = "$DistDir\cortex-$Version-$key.zip"
        Compress-Archive -Path $pkgDir -DestinationPath $archive -Force
    } else {
        $archive = "$DistDir\cortex-$Version-$key.tar.gz"
        tar -czf $archive -C $DistDir "cortex-$Version-$key"
    }
    
    Write-Host "  Created $archive" -ForegroundColor Gray
}

Write-Host "`n=== Release Complete ===" -ForegroundColor Green
Write-Host "Packages created in $DistDir" -ForegroundColor Cyan
Get-ChildItem "$DistDir\*.zip", "$DistDir\*.tar.gz" | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 1)
    Write-Host "  $($_.Name) - ${size}MB" -ForegroundColor Gray
}
