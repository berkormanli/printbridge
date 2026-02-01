package tray

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/systray"
)

// PrinterInfo contains USB printer details (matches adapter.PrinterInfo).
type PrinterInfo struct {
	VendorID     uint16 `json:"vendor_id"`
	ProductID    uint16 `json:"product_id"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
}

// App represents the system tray application.
type App struct {
	statusFn       func() (bool, string)
	testPrintFn    func() error
	restartFn      func()
	listPrintersFn func() ([]PrinterInfo, error)
	selectDeviceFn func(vendorID, productID uint16) error
	configPath     string
	serviceURL     string
	mStatus        *systray.MenuItem
	currentVID     uint16
	currentPID     uint16
}

// New creates a new tray application.
func New() *App {
	return &App{
		serviceURL: "http://localhost:9100",
	}
}

// SetStatusFunc sets the function to get printer status.
func (a *App) SetStatusFunc(fn func() (bool, string)) {
	a.statusFn = fn
}

// SetTestPrintFunc sets the function to run a test print.
func (a *App) SetTestPrintFunc(fn func() error) {
	a.testPrintFn = fn
}

// SetRestartFunc sets the function to restart the service.
func (a *App) SetRestartFunc(fn func()) {
	a.restartFn = fn
}

// SetConfigPath sets the config file path for "Open Config" menu.
func (a *App) SetConfigPath(path string) {
	a.configPath = path
}

// SetServiceURL sets the base URL for the HTTP service.
func (a *App) SetServiceURL(url string) {
	a.serviceURL = url
}

// SetListPrintersFn sets the function to list available USB printers.
func (a *App) SetListPrintersFn(fn func() ([]PrinterInfo, error)) {
	a.listPrintersFn = fn
}

// SetSelectDeviceFn sets the function to select a USB device.
func (a *App) SetSelectDeviceFn(fn func(vendorID, productID uint16) error) {
	a.selectDeviceFn = fn
}

// SetCurrentDevice sets the currently configured USB device.
func (a *App) SetCurrentDevice(vendorID, productID uint16) {
	a.currentVID = vendorID
	a.currentPID = productID
}

// Run starts the system tray application.
func (a *App) Run() {
	systray.Run(a.onReady, a.onExit)
}

// RunWithExistingLoop runs systray in an existing event loop (for integration).
func (a *App) RunWithExistingLoop() {
	go systray.Run(a.onReady, a.onExit)
}

func (a *App) onReady() {
	// Set tray icon
	systray.SetIcon(getIcon())
	systray.SetTitle("")
	a.updateTooltip()

	// Menu items - Status is a submenu showing current state
	a.mStatus = systray.AddMenuItem("Service: Checking...", "Service status")
	a.mStatus.Disable() // Status line is informational only

	mRefresh := systray.AddMenuItem("Refresh Status", "Check service status")
	mTestPrint := systray.AddMenuItem("Test Print", "Send a test receipt")
	systray.AddSeparator()

	// USB Devices submenu
	mUSBDevices := systray.AddMenuItem("USB Devices", "Select USB printer")
	mScanDevices := mUSBDevices.AddSubMenuItem("Scan for Devices...", "Scan for connected USB printers")

	systray.AddSeparator()
	mRestart := systray.AddMenuItem("Restart Service", "Restart the print service")
	mOpenConfig := systray.AddMenuItem("Open Config", "Open configuration file")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop service and exit")

	// Initial status check
	go a.refreshStatus()

	// Periodic status updates every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			a.refreshStatus()
		}
	}()

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mRefresh.ClickedCh:
				a.refreshStatus()
				showNotification("PrintBridge", "Status refreshed")
			case <-mTestPrint.ClickedCh:
				a.testPrint()
			case <-mScanDevices.ClickedCh:
				a.scanAndShowDevices(mUSBDevices)
			case <-mRestart.ClickedCh:
				a.restart()
			case <-mOpenConfig.ClickedCh:
				a.openConfig()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

// scanAndShowDevices scans for USB printers and updates the submenu.
func (a *App) scanAndShowDevices(parent *systray.MenuItem) {
	if a.listPrintersFn == nil {
		showNotification("PrintBridge", "USB device scanning not available")
		return
	}

	printers, err := a.listPrintersFn()
	if err != nil {
		showNotification("PrintBridge - Error", fmt.Sprintf("Failed to scan: %v", err))
		return
	}

	if len(printers) == 0 {
		showNotification("PrintBridge", "No USB printers found")
		return
	}

	// Show notification with found devices
	var msg string
	for i, p := range printers {
		name := p.Product
		if name == "" {
			name = fmt.Sprintf("Device %04X:%04X", p.VendorID, p.ProductID)
		}
		if p.Manufacturer != "" {
			name = fmt.Sprintf("%s (%s)", name, p.Manufacturer)
		}

		// Mark current device
		if p.VendorID == a.currentVID && p.ProductID == a.currentPID {
			name = "✓ " + name
		}

		msg += fmt.Sprintf("%d. %s\n", i+1, name)

		// Add submenu item for each printer
		item := parent.AddSubMenuItem(name, fmt.Sprintf("Select %s", name))

		// Capture values for closure
		vid, pid := p.VendorID, p.ProductID
		go func() {
			for range item.ClickedCh {
				a.selectDevice(vid, pid)
			}
		}()
	}

	showNotification("PrintBridge - USB Devices Found", msg)
}

// selectDevice selects a USB device and updates the config.
func (a *App) selectDevice(vendorID, productID uint16) {
	if a.selectDeviceFn == nil {
		showNotification("PrintBridge", "Device selection not available")
		return
	}

	if err := a.selectDeviceFn(vendorID, productID); err != nil {
		showNotification("PrintBridge - Error", fmt.Sprintf("Failed to select device: %v", err))
		return
	}

	a.currentVID = vendorID
	a.currentPID = productID

	showNotification("PrintBridge", fmt.Sprintf("Selected device %04X:%04X. Restarting service...", vendorID, productID))

	// Restart service to apply changes
	if a.restartFn != nil {
		a.restartFn()
	}
}

func (a *App) onExit() {
	// Cleanup if needed
}

func (a *App) updateTooltip() {
	systray.SetTooltip("PrintBridge - Receipt Printer Service")
}

// refreshStatus checks the /health and /status endpoints and updates the menu
func (a *App) refreshStatus() {
	// Check /health endpoint
	healthOK := a.checkHealth()

	// Check /status endpoint
	printerStatus := a.checkPrinterStatus()

	// Update menu item
	var statusText string
	if !healthOK {
		statusText = "Service: Stopped"
	} else if printerStatus.Connected {
		statusText = "Service: Running | Printer: Connected"
	} else {
		statusText = "Service: Running | Printer: Disconnected"
	}

	a.mStatus.SetTitle(statusText)
}

// checkHealth calls the /health endpoint
func (a *App) checkHealth() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(a.serviceURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// StatusResponse represents the response from /status endpoint
type StatusResponse struct {
	Connected bool   `json:"connected"`
	Service   string `json:"service"`
}

// checkPrinterStatus calls the /status endpoint
func (a *App) checkPrinterStatus() StatusResponse {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(a.serviceURL + "/status")
	if err != nil {
		return StatusResponse{}
	}
	defer resp.Body.Close()

	var status StatusResponse
	json.NewDecoder(resp.Body).Decode(&status)
	return status
}

func (a *App) testPrint() {
	if a.testPrintFn == nil {
		showNotification("PrintBridge", "Test print function not configured")
		return
	}

	if err := a.testPrintFn(); err != nil {
		showNotification("PrintBridge - Error", err.Error())
	} else {
		showNotification("PrintBridge", "Test print sent successfully!")
	}
}

func (a *App) restart() {
	if a.restartFn != nil {
		showNotification("PrintBridge", "Restarting service...")
		a.restartFn()
	}
}

func (a *App) openConfig() {
	if a.configPath == "" {
		showNotification("PrintBridge", "Config path not set")
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", a.configPath)
	case "linux":
		cmd = exec.Command("xdg-open", a.configPath)
	case "windows":
		cmd = exec.Command("notepad", a.configPath)
	}

	if cmd != nil {
		cmd.Start()
	}
}

// showNotification displays a system notification or falls back to stdout.
func showNotification(title, message string) {
	// Try to use native notifications
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		exec.Command("osascript", "-e", script).Run()
	case "linux":
		exec.Command("notify-send", title, message).Run()
	case "windows":
		// Windows toast notifications require more setup, fallback to console
		fmt.Printf("[%s] %s\n", title, message)
	default:
		fmt.Printf("[%s] %s\n", title, message)
	}
}

// getIcon returns the tray icon bytes.
func getIcon() []byte {
	return Icon
}
