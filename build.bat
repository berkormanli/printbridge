@echo off
setlocal enabledelayedexpansion

echo ============================================
echo  PrintBridge Complete Build Script
echo ============================================
echo.

:: Step 1: Compile Windows Resources
echo [1/5] Compiling Resources...

windres cmd/tray/printbridge-tray.rc -O coff -o cmd/tray/printbridge-tray.syso 2>nul
if %errorlevel% neq 0 (
    echo      Warning: windres not found or failed. Skipping tray icon.
)

echo.

:: Step 2: Build Go executables
echo [2/5] Building PrintBridge Service...
set CGO_ENABLED=1
set GOARCH=amd64
go build -o printbridge_service.exe ./cmd/server
if %errorlevel% neq 0 (
    echo      ERROR: Failed to build service!
    exit /b %errorlevel%
)
echo      Built: printbridge_service.exe

echo [3/5] Building PrintBridge Tray App...
go build -ldflags "-H=windowsgui" -o printbridge-tray.exe ./cmd/tray
if %errorlevel% neq 0 (
    echo      ERROR: Failed to build tray app!
    exit /b %errorlevel%
)
echo      Built: printbridge-tray.exe

echo [4/5] Building PrintBridge Desktop App (Wails)...
wails build
if %errorlevel% neq 0 (
    echo      ERROR: Failed to build Wails app!
    exit /b %errorlevel%
)
echo      Built: build\bin\printbridge.exe

echo.

:: Step 5: Stage files for installer
echo [5/5] Staging files for installer...
if not exist "installer\build\windows" mkdir "installer\build\windows"
copy /Y printbridge_service.exe installer\build\windows\ >nul
copy /Y printbridge-tray.exe installer\build\windows\ >nul
copy /Y build\bin\printbridge.exe installer\build\windows\printbridge-gui.exe >nul
copy /Y tray\iconwin.ico installer\build\windows\ >nul
copy /Y config.json installer\build\windows\ >nul
copy /Y README.md installer\build\windows\ >nul
echo      Staged: installer\build\windows\

echo.
echo ============================================
echo  Build Complete!
echo ============================================
echo.
echo Executables:
echo   - printbridge_service.exe
echo   - printbridge-tray.exe
echo   - printbridge-gui.exe
echo.
echo Installer staging:
echo   - installer\build\windows\
echo.
echo To build installer:
echo   1. Ensure deps are in installer\deps\:
echo      - libusb-1.0.dll (from github.com/libusb/libusb/releases)
echo      - vc_redist.x64.exe (from aka.ms/vs/17/release/vc_redist.x64.exe)
echo   2. Run: build-installer.bat
echo.
