@echo off
REM Setup script for PrintBridge USB support on Windows
REM This script downloads pre-built libusb binaries (no Visual Studio required)

setlocal enabledelayedexpansion

echo ============================================
echo   PrintBridge USB Support Setup
echo ============================================
echo.

set LIBUSB_VERSION=1.0.27
set LIBUSB_DIR=C:\libusb
set DOWNLOAD_URL=https://github.com/libusb/libusb/releases/download/v%LIBUSB_VERSION%/libusb-%LIBUSB_VERSION%.7z
set SEVENZIP_URL=https://www.7-zip.org/a/7zr.exe

REM Create libusb directory
if not exist "%LIBUSB_DIR%" mkdir "%LIBUSB_DIR%"

REM Check for 7-Zip
set SEVENZIP_EXE=
where 7z >nul 2>&1
if not errorlevel 1 (
    set SEVENZIP_EXE=7z
) else (
    if exist "%ProgramFiles%\7-Zip\7z.exe" (
        set "SEVENZIP_EXE=%ProgramFiles%\7-Zip\7z.exe"
    ) else if exist "%ProgramFiles(x86)%\7-Zip\7z.exe" (
        set "SEVENZIP_EXE=%ProgramFiles(x86)%\7-Zip\7z.exe"
    )
)

REM Download 7zr.exe if no 7-Zip found
if "%SEVENZIP_EXE%"=="" (
    echo 7-Zip not found. Downloading 7zr.exe...
    powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%SEVENZIP_URL%' -OutFile '%LIBUSB_DIR%\7zr.exe'"
    if errorlevel 1 (
        echo ERROR: Failed to download 7zr.exe
        exit /b 1
    )
    set "SEVENZIP_EXE=%LIBUSB_DIR%\7zr.exe"
)

echo Using 7-Zip: %SEVENZIP_EXE%
echo.

REM Download libusb
echo Downloading libusb %LIBUSB_VERSION%...
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%LIBUSB_DIR%\libusb.7z'"
if errorlevel 1 (
    echo ERROR: Failed to download libusb.
    exit /b 1
)

echo Extracting libusb...
"%SEVENZIP_EXE%" x "%LIBUSB_DIR%\libusb.7z" -o"%LIBUSB_DIR%" -y
if errorlevel 1 (
    echo ERROR: Failed to extract libusb.
    exit /b 1
)
del "%LIBUSB_DIR%\libusb.7z"

echo.
echo Creating pkg-config file...

REM Create pkgconfig directory
set PKGCONFIG_DIR=%LIBUSB_DIR%\pkgconfig
if not exist "%PKGCONFIG_DIR%" mkdir "%PKGCONFIG_DIR%"

REM Create libusb-1.0.pc file (using forward slashes for pkg-config)
(
echo prefix=C:/libusb
echo exec_prefix=${prefix}
echo libdir=${exec_prefix}/VS2022/MS64/dll
echo includedir=${prefix}/include
echo.
echo Name: libusb-1.0
echo Description: C API for USB device access
echo Version: %LIBUSB_VERSION%
echo Libs: -L${libdir} -lusb-1.0
echo Cflags: -I${includedir}/libusb-1.0
) > "%PKGCONFIG_DIR%\libusb-1.0.pc"

echo.
echo ============================================
echo   Setup Complete!
echo ============================================
echo.
echo libusb installed to: %LIBUSB_DIR%
echo.
echo Now run the following commands in PowerShell (as Administrator) to set environment variables:
echo.
echo   [Environment]::SetEnvironmentVariable("PKG_CONFIG_PATH", "C:\libusb\pkgconfig", "User")
echo   $oldPath = [Environment]::GetEnvironmentVariable("PATH", "User")
echo   [Environment]::SetEnvironmentVariable("PATH", "$oldPath;C:\libusb\VS2022\MS64\dll", "User")
echo.
echo Then restart your terminal and run: .\scripts\build.bat
echo.

endlocal
