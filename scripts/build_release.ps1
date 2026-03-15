# Cortex Release Build Script for Windows
# Builds Cortex for Windows, Linux, and macOS

param(
    [string]$Version = "0.1.0"
)

$ErrorActionPreference = "Stop"
$DistDir = "dist"
$BuildDir = "build"

Write-Host "=== Cortex Release Build v$Version ===" -ForegroundColor Cyan

# Clean previous builds
if (Test-Path $DistDir) { Remove-Item -Recurse -Force $DistDir }
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

function Build-Cortex {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$Suffix
    )
    
    Write-Host "Building for $OS/$Arch..." -ForegroundColor Yellow
    
    $Output = if ($OS -eq "windows") { "cortex.exe" } else { "cortex" }
    $Env:GOOS = $OS
    $Env:GOARCH = $Arch
    
    $BuildOutput = "$BuildDir/$Output"
    
    go build -ldflags="-s -w -X main.Version=$Version" -o $BuildOutput ./cmd/cortex
    
    # Create package directory
    $PkgDir = "$DistDir/cortex-$Version-$OS-$Arch"
    New-Item -ItemType Directory -Force -Path "$PkgDir/bin" | Out-Null
    New-Item -ItemType Directory -Force -Path "$PkgDir/runtime" | Out-Null
    New-Item -ItemType Directory -Force -Path "$PkgDir/examples" | Out-Null
    New-Item -ItemType Directory -Force -Path "$PkgDir/docs" | Out-Null
    
    # Copy binary
    Copy-Item $BuildOutput "$PkgDir/bin/"
    
    # Copy runtime files
    Copy-Item runtime/*.c "$PkgDir/runtime/"
    Copy-Item runtime/*.h "$PkgDir/runtime/"
    
    # Copy examples
    Get-ChildItem -Path examples -Filter "*.cx" -Recurse | ForEach-Object {
        $destDir = "$PkgDir/examples/$($_.Directory.Name)"
        New-Item -ItemType Directory -Force -Path $destDir -ErrorAction SilentlyContinue | Out-Null
        Copy-Item $_.FullName $destDir
    }
    
    # Copy documentation
    Copy-Item README.md $PkgDir/
    Copy-Item LANGUAGE_SPEC.md "$PkgDir/docs/"
    Copy-Item LANGUAGE_GUIDE.md "$PkgDir/docs/"
    Copy-Item LICENSE $PkgDir/
    
    # Create install script
    $InstallScript = @"
@echo off
echo Installing Cortex $Version...
setlocal

rem Get the directory where this script is located
set CORTEX_DIR=%~dp0
set CORTEX_BIN=%CORTEX_DIR%bin

rem Add to PATH for current session
set PATH=%CORTEX_BIN%;%PATH%

rem Check if already in user PATH
echo %PATH% | findstr /C:"%CORTEX_BIN%" >nul
if errorlevel 1 (
    echo Adding Cortex to user PATH...
    setx PATH "%PATH%;%CORTEX_BIN%"
    echo Cortex added to PATH. Please restart your terminal.
) else (
    echo Cortex is already in PATH.
)

echo.
echo Cortex $Version installed successfully!
echo Run 'cortex --help' to get started.
"@
    $InstallScript | Out-File -FilePath "$PkgDir/install.bat" -Encoding ASCII
    
    # Create README for the package
    $PkgReadme = @"
# Cortex $Version - $OS $Arch

## Installation

### Windows
1. Extract this archive to a folder (e.g., C:\Cortex)
2. Run install.bat to add Cortex to your PATH
3. Restart your terminal
4. Run: cortex --help

### Linux/macOS
1. Extract this archive
2. Add the bin directory to your PATH:
   export PATH=/path/to/cortex/bin:$PATH
3. Add to your shell profile for persistence
4. Run: cortex --help

## Requirements

Cortex requires a C compiler for building executables:

### Windows
- TCC (recommended - included in full release)
- Or MinGW/GCC

### Linux
- GCC (usually pre-installed)
- On Ubuntu/Debian: sudo apt install gcc

### macOS
- Xcode Command Line Tools: xcode-select --install

## Quick Start

```bash
# Create a new project
cortex new myapp

# Run a file
cortex run main.cx

# Build an executable
cortex build main.cx -o myapp
```

## What's Included

- bin/       - Cortex compiler binary
- runtime/   - C runtime source files
- examples/  - Example Cortex programs
- docs/      - Documentation

## Documentation

- docs/LANGUAGE_SPEC.md - Language specification
- docs/LANGUAGE_GUIDE.md - Full language guide
- README.md - Project overview
"@
    $PkgReadme | Out-File -FilePath "$PkgDir/README.txt" -Encoding UTF8
    
    # Create archive
    $ArchiveName = "cortex-$Version-$OS-$Arch"
    if ($OS -eq "windows") {
        Compress-Archive -Path $PkgDir -DestinationPath "$DistDir/$ArchiveName.zip"
    } else {
        # Use tar on Windows 10+
        tar -czvf "$DistDir/$ArchiveName.tar.gz" -C $DistDir $ArchiveName
    }
    
    Write-Host "Created package for $OS/$Arch" -ForegroundColor Green
}

# Build for all platforms
Build-Cortex -OS "windows" -Arch "amd64" -Suffix ".exe"
Build-Cortex -OS "windows" -Arch "386" -Suffix ".exe"
Build-Cortex -OS "linux" -Arch "amd64" -Suffix ""
Build-Cortex -OS "linux" -Arch "arm64" -Suffix ""
Build-Cortex -OS "darwin" -Arch "amd64" -Suffix ""
Build-Cortex -OS "darwin" -Arch "arm64" -Suffix ""

# Cleanup build dir
Remove-Item -Recurse -Force $BuildDir

Write-Host ""
Write-Host "=== Build Complete ===" -ForegroundColor Cyan
Write-Host "Release packages created in $DistDir/"
Get-ChildItem $DistDir | Format-Table Name, Length
