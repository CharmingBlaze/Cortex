# Raylib Build Script for Windows
# Builds raylib 5.5 as a static library for use with Cortex

param(
    [string]$Compiler = "gcc",
    [string]$BuildDir = "third_party/raylib/src"
)

Write-Host "Building raylib 5.5 with $Compiler..." -ForegroundColor Cyan

$srcPath = Join-Path $BuildDir "raylib.dll.rc"
if (Test-Path $srcPath) {
    Write-Host "Found raylib source at $BuildDir" -ForegroundColor Green
} else {
    Write-Error "Raylib source not found at $BuildDir"
    exit 1
}

# Compile flags
$cflags = @(
    "-c",
    "-O2",
    "-Wall",
    "-std=c99",
    "-DPLATFORM_DESKTOP",
    "-DGRAPHICS_API_OPENGL_33",
    "-DBUILD_LIBTYPE_SHARED=0",
    "-I."
)

# Get all C source files
$cFiles = Get-ChildItem -Path $BuildDir -Filter "*.c" -File | ForEach-Object { $_.FullName }

if ($cFiles.Count -eq 0) {
    Write-Error "No C source files found in $BuildDir"
    exit 1
}

Write-Host "Found $($cFiles.Count) source files" -ForegroundColor Yellow

# Compile each source file
$objFiles = @()
foreach ($file in $cFiles) {
    $objFile = [System.IO.Path]::ChangeExtension($file, ".o")
    Write-Host "Compiling: $(Split-Path $file -Leaf)" -ForegroundColor DarkGray
    
    $compileArgs = $cflags + @($file, "-o", $objFile)
    & $Compiler $compileArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to compile $file"
        exit 1
    }
    
    $objFiles += $objFile
}

# Create static library
$libPath = Join-Path $BuildDir "libraylib.a"
Write-Host "Creating static library: $libPath" -ForegroundColor Yellow

& ar rcs $libPath $objFiles

if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to create static library"
    exit 1
}

Write-Host "Raylib built successfully!" -ForegroundColor Green
Write-Host "Library: $libPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "To use with Cortex:" -ForegroundColor White
Write-Host "  cortex -i game.cx -o game -use raylib" -ForegroundColor Gray
