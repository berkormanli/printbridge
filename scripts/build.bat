@echo off
REM Build script for PrintBridge (Windows)
REM Run this script from the project root directory
REM
REM For USB support, first run: scripts\setup-usb.bat
REM Or install libusb manually and set PKG_CONFIG_PATH

setlocal enabledelayedexpansion

set VERSION=%1
if "%VERSION%"=="" set VERSION=1.0.0

set OUTPUT_DIR=build\windows

REM Set PKG_CONFIG_PATH for libusb (check multiple locations)
if "%PKG_CONFIG_PATH%"=="" (
    if exist "C:\libusb\pkgconfig" (
        set PKG_CONFIG_PATH=C:\libusb\pkgconfig
    ) else if exist "C:\vcpkg\installed\x64-windows\lib\pkgconfig" (
        set PKG_CONFIG_PATH=C:\vcpkg\installed\x64-windows\lib\pkgconfig
    )
)

REM Add libusb DLL to PATH
if exist "C:\libusb\MinGW64\dll" (
    set "PATH=%PATH%;C:\libusb\MinGW64\dll"
) else if exist "C:\vcpkg\installed\x64-windows\bin" (
    set "PATH=%PATH%;C:\vcpkg\installed\x64-windows\bin"
)

REM Fix Strawberry Perl locale warnings that break pkg-config parsing
set LC_ALL=C
set LANG=C

echo Building PrintBridge v%VERSION%...
if defined PKG_CONFIG_PATH (
    echo Using PKG_CONFIG_PATH: %PKG_CONFIG_PATH%
) else (
    echo WARNING: PKG_CONFIG_PATH not set. USB support may fail.
    echo Run scripts\setup-usb.bat first for USB support.
)

REM Create output directory
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM Build main executable
echo Building PrintBridge main executable...
REM CGO_ENABLED=0 to avoid libusb dependency (USB adapter will be disabled)
REM Set CGO_ENABLED=1 if you have libusb installed and need USB support
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o "%OUTPUT_DIR%\printbridge.exe" .
if errorlevel 1 (
    echo ERROR: Failed to build printbridge.exe
    exit /b 1
)

REM Build tray application
echo Building PrintBridge tray application...
go build -ldflags="-s -w" -o "%OUTPUT_DIR%\printbridge-tray.exe" .\cmd\tray\
if errorlevel 1 (
    echo ERROR: Failed to build printbridge-tray.exe
    exit /b 1
)

REM Copy supporting files
echo Copying supporting files...
if exist "config.json" copy /Y "config.json" "%OUTPUT_DIR%\" >nul
if exist "README.md" copy /Y "README.md" "%OUTPUT_DIR%\" >nul
if exist "LICENSE" copy /Y "LICENSE" "%OUTPUT_DIR%\" >nul

echo.
echo ============================================
echo   Windows build complete!
echo ============================================
echo.
echo Output files:
dir /B "%OUTPUT_DIR%"
echo.
echo Location: %OUTPUT_DIR%\

endlocal
