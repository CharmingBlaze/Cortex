@echo off
set PATH=C:\msys64\mingw64\bin;%PATH%
set XDG_DATA_DIRS=C:\msys64\mingw64\share
set GSETTINGS_SCHEMA_DIR=C:\msys64\mingw64\share\glib-2.0\schemas
hello_gui.exe 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Program exited with error code: %ERRORLEVEL%
)
pause
