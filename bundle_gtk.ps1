# Bundle GTK4 with your Cortex GUI application
# Run this script to create a portable distribution

param(
    [string]$OutputDir = "dist",
    [string]$AppExe = "hello_gui.exe"
)

$MSYS2Bin = "C:\msys64\mingw64\bin"
$MSYS2Share = "C:\msys64\mingw64\share"
$MSYS2Lib = "C:\msys64\mingw64\lib"

Write-Host "Creating GTK4 bundled distribution..." -ForegroundColor Cyan

# Create output directory
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path $OutputDir | Out-Null

# Required GTK4 DLLs
$gtkDlls = @(
    "libgtk-4-1.dll",
    "libgdk_pixbuf-2.0-0.dll",
    "libpango-1.0-0.dll",
    "libpangocairo-1.0-0.dll",
    "libpangowin32-1.0-0.dll",
    "libpangoft2-1.0-0.dll",
    "libcairo-2.dll",
    "libcairo-gobject-2.dll",
    "libgio-2.0-0.dll",
    "libglib-2.0-0.dll",
    "libgobject-2.0-0.dll",
    "libharfbuzz-0.dll",
    "libgraphene-1.0-0.dll",
    "libintl-8.dll",
    "libffi-8.dll",
    "libpixman-1-0.dll",
    "libpng16-16.dll",
    "libfreetype-6.dll",
    "libbz2-1.dll",
    "libbrotlidec.dll",
    "libbrotlicommon.dll",
    "libzstd.dll",
    "libexpat-1.dll",
    "libfontconfig-1.dll",
    "libiconv-2.dll",
    "libwinpthread-1.dll",
    "libgcc_s_seh-1.dll",
    "libstdc++-6.dll",
    "zlib1.dll",
    "libcrypto-3-x64.dll",
    "libssl-3-x64.dll",
    "libtiff-6.dll",
    "libjpeg-8.dll",
    "libwebp-7.dll",
    "libjbig-0.dll",
    "libLerc.dll",
    "libdeflate.dll",
    "libidn2-0.dll",
    "libssh2-1.dll",
    "libpsl-5.dll",
    "libunistring-2.dll",
    "libbrotlienc.dll",
    "libfribidi-0.dll"
)

# Copy DLLs
Write-Host "Copying GTK4 DLLs..."
$copiedDlls = 0
foreach ($dll in $gtkDlls) {
    $src = Join-Path $MSYS2Bin $dll
    if (Test-Path $src) {
        Copy-Item $src -Destination $OutputDir -Force
        $copiedDlls++
    } else {
        Write-Host "  Warning: $dll not found" -ForegroundColor Yellow
    }
}
Write-Host "  Copied $copiedDlls DLLs" -ForegroundColor Green

# Copy additional dependencies that might be loaded dynamically
$extraDlls = Get-ChildItem $MSYS2Bin -Filter "*.dll" | 
    Where-Object { $_.Name -match "^(libcurl|libnghttp|libpcre2|libreadline|libncurses|libtinfo|libtermcap|libedit|libsqlite)" }
foreach ($dll in $extraDlls) {
    Copy-Item $dll.FullName -Destination $OutputDir -Force -ErrorAction SilentlyContinue
}

# Create lib directory for GDK-Pixbuf loaders
$libDir = Join-Path $OutputDir "lib"
$loadersDir = Join-Path $libDir "gdk-pixbuf-2.0\2.10.0\loaders"
New-Item -ItemType Directory -Path $loadersDir -Force | Out-Null

# Copy pixbuf loaders (for image loading)
Write-Host "Copying GDK-Pixbuf loaders..."
$loaders = Get-ChildItem "$MSYS2Lib\gdk-pixbuf-2.0\2.10.0\loaders" -Filter "*.dll" -ErrorAction SilentlyContinue
foreach ($loader in $loaders) {
    Copy-Item $loader.FullName -Destination $loadersDir -Force
}
Write-Host "  Copied $($loaders.Count) loaders"

# Copy loaders.cache
$loadersCache = "$MSYS2Lib\gdk-pixbuf-2.0\2.10.0\loaders.cache"
if (Test-Path $loadersCache) {
    Copy-Item $loadersCache -Destination (Join-Path $libDir "gdk-pixbuf-2.0\2.10.0") -Force
}

# Create share directory for icons and schemas
$shareDir = Join-Path $OutputDir "share"
New-Item -ItemType Directory -Path $shareDir -Force | Out-Null

# Copy Adwaita icons (required for GTK4 theme)
Write-Host "Copying GTK icons..."
$iconsDir = Join-Path $shareDir "icons"
New-Item -ItemType Directory -Path $iconsDir -Force | Out-Null

# Copy only essential Adwaita icons
$adwaitaSrc = "$MSYS2Share\icons\Adwaita"
$adwaitaDest = Join-Path $iconsDir "Adwaita"
if (Test-Path $adwaitaSrc) {
    # Copy symbolic icons (needed for window controls)
    Copy-Item -Path "$adwaitaSrc\symbolic" -Destination $adwaitaDest -Recurse -Force -ErrorAction SilentlyContinue
    # Copy index.theme
    Copy-Item -Path "$adwaitaSrc\index.theme" -Destination $adwaitaDest -Force -ErrorAction SilentlyContinue
    Write-Host "  Copied Adwaita icons"
}

# Copy GLib schemas
Write-Host "Copying GLib schemas..."
$schemasDir = Join-Path $shareDir "glib-2.0\schemas"
New-Item -ItemType Directory -Path $schemasDir -Force | Out-Null
Copy-Item -Path "$MSYS2Share\glib-2.0\schemas\*.compiled" -Destination $schemasDir -Force -ErrorAction SilentlyContinue
Copy-Item -Path "$MSYS2Share\glib-2.0\schemas\gschemas.compiled" -Destination $schemasDir -Force -ErrorAction SilentlyContinue

# Copy application executable
if (Test-Path $AppExe) {
    Copy-Item $AppExe -Destination $OutputDir -Force
    Write-Host "Copied $AppExe"
} else {
    Write-Host "Warning: $AppExe not found, skipping" -ForegroundColor Yellow
}

# Create a launcher batch file
$launcherBat = @"
@echo off
set GDK_PIXBUF_MODULE_FILE=%~dp0lib\gdk-pixbuf-2.0\2.10.0\loaders.cache
set GDK_PIXBUF_MODULEDIR=%~dp0lib\gdk-pixbuf-2.0\2.10.0\loaders
set XDG_DATA_DIRS=%~dp0share
set GSETTINGS_SCHEMA_DIR=%~dp0share\glib-2.0\schemas
"%~dp0$AppExe"
"@

Set-Content -Path (Join-Path $OutputDir "run.bat") -Value $launcherBat -Encoding ASCII

# Calculate total size
$totalSize = (Get-ChildItem $OutputDir -Recurse | Measure-Object -Property Length -Sum).Sum / 1MB

Write-Host ""
Write-Host "Bundle created successfully!" -ForegroundColor Green
Write-Host "  Output: $OutputDir"
Write-Host "  Size: $([math]::Round($totalSize, 1)) MB"
Write-Host ""
Write-Host "To distribute: Copy the '$OutputDir' folder to any Windows machine"
Write-Host "To run: Double-click run.bat or run $AppExe directly"
