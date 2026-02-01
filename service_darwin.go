//go:build darwin
// +build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const launchAgentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.printbridge.service</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/printbridge.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/printbridge.error.log</string>
    <key>WorkingDirectory</key>
    <string>%s</string>
</dict>
</plist>`

// InstallService installs the service as a macOS LaunchAgent.
func InstallService() error {
	// Get executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	workDir := filepath.Dir(execPath)

	// Create LaunchAgents directory if needed
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	launchAgentsDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Write plist file
	plistPath := filepath.Join(launchAgentsDir, "com.printbridge.service.plist")
	plistContent := fmt.Sprintf(launchAgentPlist, execPath, workDir)

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", plistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load service: %s: %w", string(output), err)
	}

	fmt.Printf("Service installed: %s\n", plistPath)
	fmt.Println("Service will start automatically on login")
	return nil
}

// UninstallService removes the macOS LaunchAgent.
func UninstallService() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.printbridge.service.plist")

	// Unload the service
	cmd := exec.Command("launchctl", "unload", plistPath)
	cmd.Run() // Ignore error if not loaded

	// Remove the plist file
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	fmt.Println("Service uninstalled")
	return nil
}
