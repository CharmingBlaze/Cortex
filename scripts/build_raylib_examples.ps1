# Build raylib (if not already built) and compile all Cortex raylib examples.
# Run from project root: .\scripts\build_raylib_examples.ps1

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$RaylibDir = Join-Path $ProjectRoot "third_party\raylib"
$BuildDir = Join-Path $RaylibDir "build"
$ExamplesDir = Join-Path $ProjectRoot "examples\raylib"
$ConfigPath = Join-Path $ProjectRoot "configs\raylib.json"

# Clone raylib if missing
if (-not (Test-Path (Join-Path $RaylibDir ".git"))) {
    Write-Host "Cloning raylib..."
    New-Item -ItemType Directory -Force -Path (Split-Path $RaylibDir) | Out-Null
    git clone --depth 1 https://github.com/raysan5/raylib.git $RaylibDir
}

# Build raylib if not built
if (-not (Test-Path (Join-Path $BuildDir "raylib\libraylib.dll"))) {
    Write-Host "Building raylib (this may take a few minutes)..."
    Push-Location $RaylibDir
    cmake -B build -G "MinGW Makefiles" -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=ON -DCMAKE_C_COMPILER=gcc
    cmake --build build --config Release
    Pop-Location
}

# Build Cortex compiler
Write-Host "Building Cortex compiler..."
Push-Location $ProjectRoot
go build -o cortex.exe .
if (-not $?) { exit 1 }

# Compile each raylib example
$examples = @(
    "core_basic_window",
    "core_input_keys",
    "shapes_basic_shapes"
)
foreach ($name in $examples) {
    $src = Join-Path $ExamplesDir "$name.cx"
    $out = Join-Path $ExamplesDir "$name.exe"
    if (Test-Path $src) {
        Write-Host "Compiling $name..."
        .\cortex.exe -i $src -o $out -config $ConfigPath
        if (-not $?) { exit 1 }
    }
}
Pop-Location

Write-Host ""
Write-Host "Done. Run an example from project root, e.g.:"
Write-Host "  .\examples\raylib\core_basic_window.exe"
Write-Host "On Windows, if the window fails to open, copy libraylib.dll to the exe folder:"
Write-Host "  copy third_party\raylib\build\raylib\libraylib.dll examples\raylib\"
