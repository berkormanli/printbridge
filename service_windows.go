//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallService installs the service using Windows Task Scheduler.
// For production use, consider using NSSM or Windows Service wrapper.
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

	// Create scheduled task that runs at logon
	cmd := exec.Command("schtasks", "/create",
		"/tn", "PrintBridge",
		"/tr", execPath,
		"/sc", "onlogon",
		"/rl", "highest",
		"/f",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create scheduled task: %s: %w", string(output), err)
	}

	fmt.Println("Service installed as Windows Scheduled Task")
	fmt.Println("Task Name: PrintBridge")
	fmt.Println("Trigger: At logon")
	fmt.Println("")
	fmt.Println("For a proper Windows Service, consider using NSSM:")
	fmt.Println("  nssm install PrintBridge", execPath)
	return nil
}

// UninstallService removes the Windows scheduled task.
func UninstallService() error {
	cmd := exec.Command("schtasks", "/delete",
		"/tn", "PrintBridge",
		"/f",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete scheduled task: %s: %w", string(output), err)
	}

	fmt.Println("Service uninstalled")
	return nil
}
