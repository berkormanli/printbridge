@echo off
setlocal

echo ============================================
echo  PrintBridge Installer Builder
echo ============================================
echo.

:: Check for Inno Setup
set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
if not exist "%ISCC%" (
    set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
)
if not exist "%ISCC%" (
    echo ERROR: Inno Setup 6 not found!
    echo Please install from: https://jrsoftware.org/isinfo.php
    exit /b 1
)

:: Check for required dependencies
echo Checking dependencies...

if not exist "installer\deps\libusb-1.0.dll" (
    echo.
    echo ERROR: libusb-1.0.dll not found!
    echo.
    echo Please download from: https://github.com/libusb/libusb/releases
    echo Extract VS2019\MS64\dll\libusb-1.0.dll
    echo Copy to: installer\deps\libusb-1.0.dll
    echo.
    exit /b 1
)
echo   [OK] libusb-1.0.dll

if not exist "installer\deps\vc_redist.x64.exe" (
    echo.
    echo ERROR: vc_redist.x64.exe not found!
    echo.
    echo Please download from: https://aka.ms/vs/17/release/vc_redist.x64.exe
    echo Save to: installer\deps\vc_redist.x64.exe
    echo.
    exit /b 1
)
echo   [OK] vc_redist.x64.exe

:: Check for built executables
if not exist "installer\build\windows\printbridge_service.exe" (
    echo.
    echo ERROR: printbridge_service.exe not found!
    echo Please run build.bat first.
    exit /b 1
)
echo   [OK] printbridge_service.exe

if not exist "installer\build\windows\printbridge-gui.exe" (
    echo.
    echo ERROR: printbridge-gui.exe not found!
    echo Please run build.bat first.
    exit /b 1
)
echo   [OK] printbridge-gui.exe

if not exist "installer\build\windows\printbridge-tray.exe" (
    echo.
    echo ERROR: printbridge-tray.exe not found!
    echo Please run build.bat first.
    exit /b 1
)
echo   [OK] printbridge-tray.exe

echo.
echo All dependencies found!
echo.

:: Create output directory
if not exist "installer\installer\output" mkdir "installer\installer\output"

:: Run Inno Setup Compiler
echo Compiling installer...
cd installer
"%ISCC%" printbridge.iss
if %errorlevel% neq 0 (
    echo.
    echo ERROR: Installer compilation failed!
    cd ..
    exit /b %errorlevel%
)
cd ..

echo.
echo ============================================
echo  Installer Built Successfully!
echo ============================================
echo.
echo Output: installer\installer\output\
echo.
