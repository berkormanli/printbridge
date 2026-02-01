# PrintBridge Installer Build Guide

## Prerequisites

Before building the installer, you need to:

### 1. Install Inno Setup
Download and install Inno Setup 6.x from: https://jrsoftware.org/isinfo.php

### 2. Download Dependencies

Place the following files in the `installer/deps/` folder:

#### libusb-1.0.dll (REQUIRED)
1. Download from: https://github.com/libusb/libusb/releases
2. Extract and locate: `VS2019/MS64/dll/libusb-1.0.dll`
3. Copy to: `installer/deps/libusb-1.0.dll`

#### vc_redist.x64.exe (REQUIRED)
1. Download from: https://aka.ms/vs/17/release/vc_redist.x64.exe
2. Save to: `installer/deps/vc_redist.x64.exe`

### 3. Build the Application

Run the build script to compile the executables:

```batch
build.bat
```

This will create:
- `printbridge.exe` - The service executable
- `printbridge-tray.exe` - The system tray application

### 4. Prepare Build Files

Copy the built files to the installer staging directory:

```batch
copy printbridge.exe installer\build\windows\
copy printbridge-tray.exe installer\build\windows\
copy config.json installer\build\windows\
copy README.md installer\build\windows\
```

## Building the Installer

### Option 1: Using Inno Setup GUI
1. Open `installer/printbridge.iss` in Inno Setup
2. Click **Build > Compile**
3. The installer will be created in `installer/installer/output/`

### Option 2: Using Command Line
```batch
"C:\Program Files (x86)\Inno Setup 6\ISCC.exe" installer\printbridge.iss
```

## Output

The installer will be created at:
```
installer/installer/output/PrintBridge-Setup-1.0.0.exe
```

## Directory Structure

```
installer/
├── printbridge.iss          # Inno Setup script
├── LICENSE                   # License file
├── README.md                 # This file
├── deps/                     # Dependencies folder
│   ├── libusb-1.0.dll       # libusb library (REQUIRED)
│   └── vc_redist.x64.exe    # VC++ Redistributable (REQUIRED)
├── build/
│   └── windows/             # Compiled executables
│       ├── printbridge.exe
│       ├── printbridge-tray.exe
│       ├── config.json
│       └── README.md
└── installer/
    └── output/              # Built installer output
        └── PrintBridge-Setup-1.0.0.exe
```

## What the Installer Does

1. **Installs VC++ Redistributable** (if not already installed)
2. **Copies all files** to `C:\Program Files\PrintBridge\`
3. **Bundles libusb-1.0.dll** (eliminates 0x0000007b error)
4. **Registers Windows Service** (optional)
5. **Adds tray app to startup** (optional)
6. **Creates Start Menu shortcuts**
7. **Creates uninstaller**

## Troubleshooting

### Error: "Source file does not exist"
Make sure all files exist in the correct locations:
- `installer/deps/libusb-1.0.dll`
- `installer/deps/vc_redist.x64.exe`
- `installer/build/windows/printbridge.exe`
- `installer/build/windows/printbridge-tray.exe`

### Error: 0x0000007b on target machine
This error should be resolved by the installer. If it persists:
1. Verify libusb-1.0.dll was copied to the installation directory
2. Verify VC++ Redistributable was installed
3. Ensure you're using 64-bit builds for 64-bit Windows
