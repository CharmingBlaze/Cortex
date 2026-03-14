@echo off
REM Cortex Language VSCode Extension Installer for Windows
REM This script installs the Cortex language extension for VSCode

echo ============================================
echo   Cortex Language VSCode Extension Installer
echo ============================================
echo.

REM Check if VSCode is installed
where code >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] VSCode is not installed or not in PATH.
    echo Please install VSCode from: https://code.visualstudio.com/
    echo.
    pause
    exit /b 1
)

REM Set paths
set "EXTENSION_SRC=%~dp0..\vscode-extension"
set "EXTENSION_DEST=%USERPROFILE%\.vscode\extensions\cortex-language"

echo Source: %EXTENSION_SRC%
echo Destination: %EXTENSION_DEST%
echo.

REM Create destination directory
if not exist "%EXTENSION_DEST%" mkdir "%EXTENSION_DEST%"
if not exist "%EXTENSION_DEST%\syntaxes" mkdir "%EXTENSION_DEST%\syntaxes"

REM Copy extension files
echo Copying extension files...
copy /Y "%EXTENSION_SRC%\package.json" "%EXTENSION_DEST%\" >nul
copy /Y "%EXTENSION_SRC%\language-configuration.json" "%EXTENSION_DEST%\" >nul
copy /Y "%EXTENSION_SRC%\README.md" "%EXTENSION_DEST%\" >nul
copy /Y "%EXTENSION_SRC%\syntaxes\cortex.tmLanguage.json" "%EXTENSION_DEST%\syntaxes\" >nul

echo.
echo [SUCCESS] Cortex language extension installed!
echo.
echo To verify:
echo   1. Open VSCode
echo   2. Open any .cx file
echo   3. Check that syntax highlighting is working
echo.
echo If syntax highlighting doesn't work, press Ctrl+K M and select "Cortex"
echo.
pause
