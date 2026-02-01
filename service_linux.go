//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const systemdService = `[Unit]
Description=PrintBridge Receipt Printer Service
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=5
WorkingDirectory=%s
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`

// InstallService installs the service as a systemd unit.
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

	// Create systemd user directory if needed
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	systemdDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Write service file
	servicePath := filepath.Join(systemdDir, "printbridge.service")
	serviceContent := fmt.Sprintf(systemdService, execPath, workDir)

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	// Enable and start the service
	cmd := exec.Command("systemctl", "--user", "enable", "--now", "printbridge.service")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable service: %s: %w", string(output), err)
	}

	fmt.Printf("Service installed: %s\n", servicePath)
	fmt.Println("Service enabled and started")
	return nil
}

// UninstallService removes the systemd unit.
func UninstallService() error {
	// Stop and disable the service
	exec.Command("systemctl", "--user", "stop", "printbridge.service").Run()
	exec.Command("systemctl", "--user", "disable", "printbridge.service").Run()

	// Remove the service file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	servicePath := filepath.Join(home, ".config", "systemd", "user", "printbridge.service")
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	exec.Command("systemctl", "--user", "daemon-reload").Run()

	fmt.Println("Service uninstalled")
	return nil
}
