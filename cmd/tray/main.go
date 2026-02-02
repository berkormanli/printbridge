package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/systray"
	"printbridge/pkg/config"
	"printbridge/pkg/update"
	"printbridge/tray"
)

// AppVersion is the current version of the application
const AppVersion = "1.0.8"

var (
	serviceURL  = "http://localhost:9100"
	servicePath string
	configPath  string
)

func main() {
	// Find service binary
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	configDir := config.GetConfigDir()
	
	// Look for printbridge binary in same directory
	switch runtime.GOOS {
	case "windows":
		servicePath = filepath.Join(dir, "printbridge_service.exe")
	default:
		servicePath = filepath.Join(dir, "printbridge_service")
	}

	// Find config - use AppData config directory
	configPath = filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Fallback to local directory for portable mode
		configPath = "config.json"
	}

	// Run systray
	systray.Run(onReady, onExit)
}

var (
	mStatus    *systray.MenuItem
	mStartStop *systray.MenuItem
	mUpdate    *systray.MenuItem
)

func onReady() {
	systray.SetIcon(tray.Icon)
	systray.SetTitle("")
	systray.SetTooltip(fmt.Sprintf("PrintBridge v%s", AppVersion))

	// Status display (disabled, just for info)
	mStatus = systray.AddMenuItem("Checking...", "Service status")
	mStatus.Disable()

	systray.AddSeparator()

	// Start/Stop toggle
	mStartStop = systray.AddMenuItem("Start Service", "Start or stop the service")
	mTestPrint := systray.AddMenuItem("Test Print", "Send a test receipt")
	
	systray.AddSeparator()

	// USB Devices submenu
	mUSBDevices := systray.AddMenuItem("USB Devices", "Select USB printer")
	mScanDevices := mUSBDevices.AddSubMenuItem("Scan for Devices...", "Scan for connected USB printers")

	systray.AddSeparator()
	
	mOpenConfig := systray.AddMenuItem("Open Config", "Open configuration file")
	
	systray.AddSeparator()
	
	// Update menu
	mUpdate = systray.AddMenuItem("Check for Updates", "Check for new versions")
	mVersion := systray.AddMenuItem(fmt.Sprintf("Version: %s", AppVersion), "Current version")
	mVersion.Disable()
	
	systray.AddSeparator()
	
	mQuit := systray.AddMenuItem("Quit Tray", "Close the tray app (service keeps running)")

	// Initial status check
	go updateStatus()

	// Periodic status updates
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			updateStatus()
		}
	}()

	// Check for updates on startup (after a delay)
	go func() {
		time.Sleep(10 * time.Second)
		checkForUpdates(false) // Silent check
	}()

	// Periodic update checks (every 4 hours)
	go func() {
		ticker := time.NewTicker(4 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			checkForUpdates(false) // Silent check
		}
	}()

	// Handle clicks
	go func() {
		for {
			select {
			case <-mStartStop.ClickedCh:
				toggleService()
			case <-mTestPrint.ClickedCh:
				testPrint()
			case <-mScanDevices.ClickedCh:
				scanAndShowDevices(mUSBDevices)
			case <-mOpenConfig.ClickedCh:
				openConfig()
			case <-mUpdate.ClickedCh:
				checkForUpdates(true) // Show notification even if no update
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// Tray exits, but service keeps running
}

func updateStatus() {
	running := isServiceRunning()
	connected := false

	if running {
		connected = isPrinterConnected()
	}

	// Update status text
	var statusText string
	if !running {
		statusText = "⚫ Service: Stopped"
		mStartStop.SetTitle("Start Service")
	} else if connected {
		statusText = "🟢 Service: Running | Printer: Connected"
		mStartStop.SetTitle("Stop Service")
	} else {
		statusText = "🟡 Service: Running | Printer: Disconnected"
		mStartStop.SetTitle("Stop Service")
	}

	mStatus.SetTitle(statusText)
}

func isServiceRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func isPrinterConnected() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(serviceURL + "/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var status struct {
		Connected bool `json:"connected"`
	}
	json.NewDecoder(resp.Body).Decode(&status)
	return status.Connected
}

func toggleService() {
	if isServiceRunning() {
		stopService()
	} else {
		startService()
	}
	
	// Wait a moment and update status
	time.Sleep(500 * time.Millisecond)
	updateStatus()
}

func startService() {
	// Check if service binary exists
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		showNotification("PrintBridge", fmt.Sprintf("Service binary not found: %s", servicePath))
		return
	}

	cmd := exec.Command(servicePath)
	cmd.Dir = filepath.Dir(servicePath)
	
	// Hide console window on Windows
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	
	if err := cmd.Start(); err != nil {
		showNotification("PrintBridge Error", err.Error())
		return
	}

	showNotification("PrintBridge", "Service started")
}

func stopService() {
	// Kill process by name
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd = exec.Command("pkill", "-f", "printbridge_service")
	case "windows":
		cmd = exec.Command("taskkill", "/IM", "printbridge_service.exe", "/F")
	}

	if cmd != nil {
		cmd.Run()
	}
	
	showNotification("PrintBridge", "Service stopped")
}

func testPrint() {
	if !isServiceRunning() {
		showNotification("PrintBridge", "Service is not running")
		return
	}

	// Send test print request
	payload := map[string]interface{}{
		"header": "TEST PRINT",
		"items":  []interface{}{},
		"total":  0,
		"footer": fmt.Sprintf("PrintBridge Test\n%s", time.Now().Format("2006-01-02 15:04:05")),
	}

	data, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(serviceURL+"/print", "application/json", bytes.NewReader(data))
	if err != nil {
		showNotification("PrintBridge Error", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		showNotification("PrintBridge", "Test print sent!")
	} else {
		showNotification("PrintBridge Error", fmt.Sprintf("Status: %d", resp.StatusCode))
	}
}

func openConfig() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", configPath)
	case "linux":
		cmd = exec.Command("xdg-open", configPath)
	case "windows":
		cmd = exec.Command("notepad", configPath)
	}

	if cmd != nil {
		cmd.Start()
	}
}

func showNotification(title, message string) {
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		exec.Command("osascript", "-e", script).Run()
	case "linux":
		exec.Command("notify-send", title, message).Run()
	case "windows":
		showWindowsMessageBox(title, message)
	default:
		fmt.Printf("[%s] %s\n", title, message)
	}
}

// USB device tracking
var (
	currentVID uint16
	currentPID uint16
)

// PrinterInfo for USB device detection
type PrinterInfo struct {
	VendorID     uint16 `json:"vendor_id"`
	ProductID    uint16 `json:"product_id"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	IsPrinter    bool   `json:"is_printer"`
}

// scanAndShowDevices scans for USB printers and displays them
func scanAndShowDevices(parent *systray.MenuItem) {
	if !isServiceRunning() {
		showNotification("PrintBridge", "Service must be running to scan devices")
		return
	}

	// Get printers from service /status endpoint
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(serviceURL + "/status")
	if err != nil {
		showNotification("PrintBridge Error", fmt.Sprintf("Failed to scan: %v", err))
		return
	}
	defer resp.Body.Close()

	var status struct {
		Connected bool          `json:"connected"`
		Printers  []PrinterInfo `json:"printers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		showNotification("PrintBridge Error", "Failed to parse device list")
		return
	}

	if len(status.Printers) == 0 {
		showNotification("PrintBridge", "No USB printers found")
		return
	}

	// Load current config to see selected device
	loadCurrentDevice()

	// Show notification with found devices
	var msg string
	for i, p := range status.Printers {
		name := p.Product
		if name == "" {
			name = fmt.Sprintf("Device %04X:%04X", p.VendorID, p.ProductID)
		}
		if p.Manufacturer != "" {
			name = fmt.Sprintf("%s (%s)", name, p.Manufacturer)
		}

		// Mark current device
		if p.VendorID == currentVID && p.ProductID == currentPID {
			name = "✓ " + name
		}

		// Mark non-printer devices
		if !p.IsPrinter {
			name = name + " [Not a printer]"
		}

		msg += fmt.Sprintf("%d. %s\n", i+1, name)

		// Add submenu item for each device
		item := parent.AddSubMenuItem(name, fmt.Sprintf("Select %s", name))

		// Disable non-printer devices
		if !p.IsPrinter {
			item.Disable()
		}

		// Capture values for closure
		vid, pid, isPrinter := p.VendorID, p.ProductID, p.IsPrinter
		go func() {
			for range item.ClickedCh {
				if isPrinter {
					selectDevice(vid, pid)
				}
			}
		}()
	}

	showNotification("PrintBridge - USB Devices Found", msg)
}

// loadCurrentDevice loads the current VID/PID from config
func loadCurrentDevice() {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	var cfg struct {
		USB struct {
			VendorID  uint16 `json:"vendor_id"`
			ProductID uint16 `json:"product_id"`
		} `json:"usb"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}

	currentVID = cfg.USB.VendorID
	currentPID = cfg.USB.ProductID
}

// selectDevice updates the config with the selected USB device
func selectDevice(vendorID, productID uint16) {
	// Load current config
	data, err := os.ReadFile(configPath)
	if err != nil {
		showNotification("PrintBridge Error", fmt.Sprintf("Failed to read config: %v", err))
		return
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		showNotification("PrintBridge Error", fmt.Sprintf("Failed to parse config: %v", err))
		return
	}

	// Update USB settings
	usb, ok := cfg["usb"].(map[string]interface{})
	if !ok {
		usb = make(map[string]interface{})
	}
	usb["vendor_id"] = vendorID
	usb["product_id"] = productID
	cfg["usb"] = usb

	// Set adapter to USB
	cfg["adapter"] = "usb"

	// Save config
	newData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		showNotification("PrintBridge Error", fmt.Sprintf("Failed to encode config: %v", err))
		return
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		showNotification("PrintBridge Error", fmt.Sprintf("Failed to save config: %v", err))
		return
	}

	currentVID = vendorID
	currentPID = productID

	showNotification("PrintBridge", fmt.Sprintf("Selected device %04X:%04X. Restarting service...", vendorID, productID))

	// Restart service to apply changes
	stopService()
	time.Sleep(500 * time.Millisecond)
	startService()
	time.Sleep(500 * time.Millisecond)
	updateStatus()
}

// showWindowsMessageBox displays a native Windows message box.
func showWindowsMessageBox(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")

	// Convert strings to UTF16 pointers
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	// MB_OK | MB_ICONINFORMATION = 0x40
	const MB_OK = 0x00000000
	const MB_ICONINFORMATION = 0x00000040

	messageBoxW.Call(
		0, // hWnd (null = no parent)
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(MB_OK|MB_ICONINFORMATION),
	)
}

// showWindowsYesNoBox displays a Yes/No message box and returns true if Yes was clicked
func showWindowsYesNoBox(title, message string) bool {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBoxW := user32.NewProc("MessageBoxW")

	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)

	const MB_YESNO = 0x00000004
	const MB_ICONQUESTION = 0x00000020
	const IDYES = 6

	ret, _, _ := messageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(MB_YESNO|MB_ICONQUESTION),
	)

	return ret == IDYES
}

// checkForUpdates checks for available updates
func checkForUpdates(showIfNoUpdate bool) {
	mUpdate.SetTitle("Checking for Updates...")

	info, err := update.CheckForUpdates(AppVersion)
	
	mUpdate.SetTitle("Check for Updates")

	if err != nil {
		if showIfNoUpdate {
			showNotification("PrintBridge Update Error", fmt.Sprintf("Failed to check for updates: %v", err))
		}
		return
	}

	if !info.Available {
		if showIfNoUpdate {
			showNotification("PrintBridge", fmt.Sprintf("You're up to date! (v%s)", AppVersion))
		}
		return
	}

	// Update available!
	msg := fmt.Sprintf("New version available: v%s\n\nYou have: v%s\n\nWould you like to update now?", 
		info.LatestVersion, info.CurrentVersion)

	if runtime.GOOS == "windows" {
		if showWindowsYesNoBox("PrintBridge Update Available", msg) {
			installUpdate(info)
		}
	} else {
		showNotification("PrintBridge Update", fmt.Sprintf("Version %s available! Visit: %s", info.LatestVersion, info.ReleaseURL))
	}
}

// installUpdate downloads and installs the update
func installUpdate(info *update.UpdateInfo) {
	if info.DownloadURL == "" {
		showNotification("PrintBridge Update Error", "No download URL found. Please update manually.")
		// Try to open the release page
		if runtime.GOOS == "windows" {
			exec.Command("cmd", "/c", "start", info.ReleaseURL).Start()
		}
		return
	}

	showNotification("PrintBridge", "Downloading update...")
	mUpdate.SetTitle("Downloading update...")

	// Download the installer
	installerPath, err := update.DownloadInstaller(info.DownloadURL)
	if err != nil {
		showNotification("PrintBridge Update Error", fmt.Sprintf("Download failed: %v", err))
		mUpdate.SetTitle("Check for Updates")
		return
	}

	showNotification("PrintBridge", "Installing update... The application will restart.")
	mUpdate.SetTitle("Installing...")

	// Stop the service first
	stopService()
	time.Sleep(500 * time.Millisecond)

	// Launch the installer
	if runtime.GOOS == "windows" {
		// Use ShellExecuteW with "runas" verb to request admin privileges
		// This is necessary because Inno Setup needs admin rights to write to Program Files
		err := shellExecuteRunAs(installerPath, "/SILENT /CLOSEAPPLICATIONS /RESTARTAPPLICATIONS")
		if err != nil {
			showNotification("PrintBridge Update Error", fmt.Sprintf("Failed to launch installer: %v", err))
			mUpdate.SetTitle("Check for Updates")
			return
		}
	} else {
		cmd := exec.Command(installerPath)
		if err := cmd.Start(); err != nil {
			showNotification("PrintBridge Update Error", fmt.Sprintf("Failed to launch installer: %v", err))
			mUpdate.SetTitle("Check for Updates")
			return
		}
	}

	// Exit the tray app to allow update
	systray.Quit()
}

// shellExecuteRunAs launches a program with admin privileges using ShellExecuteW
func shellExecuteRunAs(path string, args string) error {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	shellExecuteW := shell32.NewProc("ShellExecuteExW")

	// SHELLEXECUTEINFO structure
	type SHELLEXECUTEINFO struct {
		cbSize         uint32
		fMask          uint32
		hwnd           uintptr
		lpVerb         *uint16
		lpFile         *uint16
		lpParameters   *uint16
		lpDirectory    *uint16
		nShow          int32
		hInstApp       uintptr
		lpIDList       uintptr
		lpClass        *uint16
		hkeyClass      uintptr
		dwHotKey       uint32
		hIconOrMonitor uintptr
		hProcess       uintptr
	}

	const (
		SEE_MASK_NOCLOSEPROCESS = 0x00000040
		SW_SHOWNORMAL           = 1
	)

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	filePtr, _ := syscall.UTF16PtrFromString(path)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	sei := SHELLEXECUTEINFO{
		cbSize:       uint32(unsafe.Sizeof(SHELLEXECUTEINFO{})),
		fMask:        SEE_MASK_NOCLOSEPROCESS,
		lpVerb:       verbPtr,
		lpFile:       filePtr,
		lpParameters: argsPtr,
		nShow:        SW_SHOWNORMAL,
	}

	ret, _, err := shellExecuteW.Call(uintptr(unsafe.Pointer(&sei)))
	if ret == 0 {
		return fmt.Errorf("ShellExecuteEx failed: %v", err)
	}

	return nil
}

