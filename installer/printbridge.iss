; PrintBridge Inno Setup Installer Script
; Compile with Inno Setup 6.x: https://jrsoftware.org/isinfo.php

#define MyAppName "PrintBridge"
#define MyAppVersion "1.1.0"
#define MyAppPublisher "Berk Ormanlı"
#define MyAppURL "https://github.com/berkormanli/printbridge"
#define MyAppServiceName "printbridge_service.exe"
#define MyAppDashboardName "printbridge-gui.exe"
#define MyAppTrayName "printbridge-tray.exe"

[Setup]
AppId={{PRINTBRIDGE-8237ADEA-A941-4D37-BFF4-2E8AF34C6F34}}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
LicenseFile=LICENSE
OutputDir=installer\output
OutputBaseFilename=PrintBridge-Setup-{#MyAppVersion}
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startuptray"; Description: "Start tray application at Windows startup"; GroupDescription: "Startup Options:"; Flags: checkedonce

[Files]
; Main application files
Source: "build\windows\printbridge_service.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "build\windows\printbridge-gui.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "build\windows\printbridge-tray.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "build\windows\iconwin.ico"; DestDir: "{app}"; Flags: ignoreversion
Source: "build\windows\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "LICENSE"; DestDir: "{app}"; Flags: ignoreversion

; Config file - copy to AppData directory
Source: "build\windows\config.json"; DestDir: "{userappdata}\PrintBridge"; Flags: onlyifdoesntexist

; ============================================================
; TEMPLATES - Logo assets for food delivery platforms
; ============================================================
Source: "..\templates\logos\*.bmp"; DestDir: "{userappdata}\PrintBridge\templates\logos"; Flags: ignoreversion

; ============================================================
; DEPENDENCIES
; ============================================================

; ============================================================
; VC++ REDISTRIBUTABLE (Required for CGO binaries)
; ============================================================
; Download from: https://aka.ms/vs/17/release/vc_redist.x64.exe
; Place in: installer\deps\vc_redist.x64.exe
Source: "deps\vc_redist.x64.exe"; DestDir: "{tmp}"; Flags: deleteafterinstall; Check: VCRedistNeedsInstall

[Dirs]
; Create config and templates directories in AppData
Name: "{userappdata}\PrintBridge"
Name: "{userappdata}\PrintBridge\templates"
Name: "{userappdata}\PrintBridge\templates\logos"

[Icons]
Name: "{group}\PrintBridge Dashboard"; Filename: "{app}\{#MyAppDashboardName}"; IconFilename: "{app}\iconwin.ico"
Name: "{group}\{#MyAppName} Tray"; Filename: "{app}\{#MyAppTrayName}"; IconFilename: "{app}\iconwin.ico"
Name: "{group}\{#MyAppName} Service"; Filename: "{app}\{#MyAppServiceName}"
Name: "{group}\Open Config Directory"; Filename: "{userappdata}\PrintBridge"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\PrintBridge Dashboard"; Filename: "{app}\{#MyAppDashboardName}"; IconFilename: "{app}\iconwin.ico"; Tasks: desktopicon

[Run]
; Install VC++ Redistributable silently if needed
Filename: "{tmp}\vc_redist.x64.exe"; Parameters: "/install /quiet /norestart"; StatusMsg: "Installing Visual C++ Runtime..."; Check: VCRedistNeedsInstall; Flags: runhidden waituntilterminated

; Start tray after install (tray app manages the service)
Filename: "{app}\{#MyAppTrayName}"; Description: "{cm:LaunchProgram,{#MyAppName} Tray}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
; No explicit service uninstall needed - processes are killed in CurUninstallStepChanged

[Registry]
; Add tray to startup if selected
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "PrintBridge"; ValueData: """{app}\{#MyAppTrayName}"""; Flags: uninsdeletevalue; Tasks: startuptray

[Code]
// Check if VC++ Redistributable is already installed
function VCRedistNeedsInstall: Boolean;
var
  Version: String;
begin
  // Check for VC++ 2015-2022 Redistributable (x64)
  // Registry key exists if installed
  Result := True;
  if RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x64', 'Version', Version) then
  begin
    // Installed version found
    Result := False;
  end;
  if RegQueryStringValue(HKLM, 'SOFTWARE\WOW6432Node\Microsoft\VisualStudio\14.0\VC\Runtimes\x64', 'Version', Version) then
  begin
    Result := False;
  end;
end;

// Stop running processes before install
procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssInstall then
  begin
    // Kill tray app if running
    Exec('cmd.exe', '/c taskkill /IM printbridge-tray.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // Kill service if running
    Exec('cmd.exe', '/c taskkill /IM printbridge_service.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // Kill GUI if running
    Exec('cmd.exe', '/c taskkill /IM printbridge-gui.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  end;
end;

// Stop running processes before uninstall
procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  ResultCode: Integer;
begin
  if CurUninstallStep = usUninstall then
  begin
    // Kill tray app
    Exec('cmd.exe', '/c taskkill /IM printbridge-tray.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // Kill service
    Exec('cmd.exe', '/c taskkill /IM printbridge_service.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // Kill GUI
    Exec('cmd.exe', '/c taskkill /IM printbridge-gui.exe /F 2>nul', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // Wait a moment for processes to fully terminate
    Sleep(500);
  end;
end;

