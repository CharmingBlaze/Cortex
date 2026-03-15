# Cortex Windows Release Builder
# Creates a complete Windows release with TCC bundled

param(
    [string]$Version = "0.1.0",
    [string]$TccPath = "C:\tcc"
)

$ErrorActionPreference = "Stop"
$DistDir = "dist"
$BuildDir = "build"

Write-Host "=== Cortex Windows Release v$Version ===" -ForegroundColor Cyan

# Clean previous builds
if (Test-Path "$DistDir\cortex-$Version-windows-amd64") { 
    Remove-Item -Recurse -Force "$DistDir\cortex-$Version-windows-amd64" 
}
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

# Build Cortex
Write-Host "Building Cortex..." -ForegroundColor Yellow
$Env:GOOS = "windows"
$Env:GOARCH = "amd64"
go build -ldflags="-s -w -X main.Version=$Version" -o "$BuildDir\cortex.exe" ./cmd/cortex

# Create package directory
$PkgDir = "$DistDir\cortex-$Version-windows-amd64"
New-Item -ItemType Directory -Force -Path "$PkgDir\bin" | Out-Null
New-Item -ItemType Directory -Force -Path "$PkgDir\runtime" | Out-Null
New-Item -ItemType Directory -Force -Path "$PkgDir\examples" | Out-Null
New-Item -ItemType Directory -Force -Path "$PkgDir\docs" | Out-Null
New-Item -ItemType Directory -Force -Path "$PkgDir\tcc" | Out-Null

# Copy Cortex binary
Copy-Item "$BuildDir\cortex.exe" "$PkgDir\bin\"

# Copy TCC if available
if (Test-Path $TccPath) {
    Write-Host "Bundling TCC..." -ForegroundColor Yellow
    Copy-Item -Recurse "$TccPath\*" "$PkgDir\tcc\"
} else {
    Write-Host "TCC not found at $TccPath - skipping TCC bundle" -ForegroundColor Yellow
    Write-Host "Download TCC from https://bellard.org/tcc/ and extract to C:\tcc" -ForegroundColor Yellow
}

# Copy runtime files
Copy-Item runtime\*.c "$PkgDir\runtime\"
Copy-Item runtime\*.h "$PkgDir\runtime\"

# Copy examples
Get-ChildItem -Path examples -Filter "*.cx" -Recurse | ForEach-Object {
    $destDir = "$PkgDir\examples\$($_.Directory.Name)"
    New-Item -ItemType Directory -Force -Path $destDir -ErrorAction SilentlyContinue | Out-Null
    Copy-Item $_.FullName $destDir
}

# Copy documentation
Copy-Item README.md $PkgDir\
Copy-Item LANGUAGE_SPEC.md "$PkgDir\docs\"
Copy-Item LANGUAGE_GUIDE.md "$PkgDir\docs\"
Copy-Item LICENSE $PkgDir\

# Create install script
$InstallBat = @"
@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   Cortex $Version - Windows Installer
echo ========================================
echo.

rem Get the directory where this script is located
set "CORTEX_DIR=%~dp0"
set "CORTEX_DIR=%CORTEX_DIR:~0,-1%"

rem Set up paths
set "CORTEX_BIN=%CORTEX_DIR%\bin"
set "CORTEX_TCC=%CORTEX_DIR%\tcc"

rem Check if Cortex binary exists
if not exist "%CORTEX_BIN%\cortex.exe" (
    echo ERROR: cortex.exe not found in %CORTEX_BIN%
    pause
    exit /b 1
)

echo Installing Cortex to: %CORTEX_DIR%
echo.

rem Add Cortex to user PATH
echo Adding Cortex to PATH...
setx CORTEX_PATH "%CORTEX_BIN%" >nul 2>&1

rem Check if already in PATH
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "USER_PATH=%%b"
echo %USER_PATH% | findstr /C:"%CORTEX_BIN%" >nul
if errorlevel 1 (
    if defined USER_PATH (
        setx PATH "%USER_PATH%;%CORTEX_BIN%" >nul 2>&1
    ) else (
        setx PATH "%CORTEX_BIN%" >nul 2>&1
    )
    echo Added Cortex to user PATH.
) else (
    echo Cortex is already in PATH.
)

rem Set TCC path for this session
if exist "%CORTEX_TCC%\tcc.exe" (
    echo.
    echo TCC compiler found. Setting TCC path...
    setx TCC_PATH "%CORTEX_TCC%" >nul 2>&1
    echo TCC path set.
)

echo.
echo ========================================
echo   Installation Complete!
echo ========================================
echo.
echo Cortex has been installed successfully!
echo.
echo Next steps:
echo   1. Close and reopen your terminal
echo   2. Run: cortex --help
echo   3. Try: cortex run examples/hello.cx
echo.
echo Documentation:
echo   - %CORTEX_DIR%\docs\LANGUAGE_SPEC.md
echo   - %CORTEX_DIR%\docs\LANGUAGE_GUIDE.md
echo.
pause
"@
$InstallBat | Out-File -FilePath "$PkgDir\install.bat" -Encoding ASCII

# Create quick start script
$QuickStart = @"
@echo off
echo Cortex $Version Quick Start
echo ==========================
echo.
echo This will create and run a simple Cortex program.
echo.

set "CORTEX_DIR=%~dp0"
set "PATH=%CORTEX_DIR%\bin;%PATH%"

rem Create a simple test program
echo Creating test program...
(
echo // Hello World in Cortex
echo void main^(^) {
echo     println^("Hello from Cortex!"^);
echo }
) > test_hello.cx

echo Running test program...
cortex run test_hello.cx

echo.
echo If you see "Hello from Cortex!" above, Cortex is working!
echo.
pause
"@
$QuickStart | Out-File -FilePath "$PkgDir\quickstart.bat" -Encoding ASCII

# Create README
$Readme = @"
# Cortex $Version - Windows Release

## Quick Start

1. Run `install.bat` to add Cortex to your PATH
2. Close and reopen your terminal
3. Run `cortex --help`

Or just run `quickstart.bat` to test Cortex immediately!

## What's Included

```
cortex-$Version-windows-amd64/
├── bin/           - Cortex compiler (cortex.exe)
├── tcc/           - TCC C compiler (bundled)
├── runtime/       - C runtime source files
├── examples/      - Example programs
├── docs/          - Documentation
├── install.bat    - Installation script
├── quickstart.bat - Quick test script
├── README.txt     - This file
└── LICENSE        - MIT License
```

## Requirements

- Windows 10 or later (64-bit)
- Cortex comes with TCC bundled, no additional setup needed!

## Examples

```bash
# Run a simple program
cortex run examples/hello.cx

# Create a new project
cortex new myproject

# Build an executable
cortex build main.cx -o myapp.exe
```

## Documentation

- `docs/LANGUAGE_SPEC.md` - Language specification
- `docs/LANGUAGE_GUIDE.md` - Full language guide

## Support

- GitHub: https://github.com/CharmingBlaze/Cortex
- Issues: https://github.com/CharmingBlaze/Cortex/issues
"@
$Readme | Out-File -FilePath "$PkgDir\README.txt" -Encoding UTF8

# Create the zip archive
Write-Host "Creating release archive..." -ForegroundColor Yellow
$ArchivePath = "$DistDir\cortex-$Version-windows-amd64.zip"
if (Test-Path $ArchivePath) { Remove-Item $ArchivePath }
Compress-Archive -Path $PkgDir -DestinationPath $ArchivePath

# Cleanup
Remove-Item -Recurse -Force $BuildDir

Write-Host ""
Write-Host "=== Build Complete ===" -ForegroundColor Green
Write-Host "Release: $ArchivePath"
Write-Host "Size: $((Get-Item $ArchivePath).Length / 1MB) MB"
