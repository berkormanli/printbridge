package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const serviceURL = "http://localhost:9100"

// App struct
type App struct {
	ctx    context.Context
	client *http.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// PrinterInfo matches the service's PrinterInfo struct
type PrinterInfo struct {
	VendorID     uint16 `json:"vendor_id"`
	ProductID    uint16 `json:"product_id"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	IsPrinter    bool   `json:"is_printer"`
	DeviceType   string `json:"device_type"`
}

// StatusResponse represents the /status endpoint response
type StatusResponse struct {
	Connected bool          `json:"connected"`
	Service   string        `json:"service"`
	Printers  []PrinterInfo `json:"printers"`
}

// CheckServiceStatus checks if the PrintBridge service is running
func (a *App) CheckServiceStatus() (bool, error) {
	resp, err := a.client.Get(serviceURL + "/health")
	if err != nil {
		return false, nil // Service not running
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

// GetPrinters retrieves the list of printers from the service
func (a *App) GetPrinters() ([]PrinterInfo, error) {
	resp, err := a.client.Get(serviceURL + "/status")
	if err != nil {
		return nil, fmt.Errorf("service not reachable: %v", err)
	}
	defer resp.Body.Close()

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return status.Printers, nil
}

// GetConnectionStatus returns whether a printer is currently connected
func (a *App) GetConnectionStatus() (bool, error) {
	resp, err := a.client.Get(serviceURL + "/status")
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	var status StatusResponse
	json.NewDecoder(resp.Body).Decode(&status)
	return status.Connected, nil
}

// SelectPrinter is a no-op for now since the service handles connection.
// In the future, this could update config.json and restart the service.
func (a *App) SelectPrinter(name string, deviceType string, vid, pid uint16) error {
	// Currently, the service auto-connects to the first available printer.
	// If we want to select a specific printer, we'd need to update config.json
	// and restart the service (like the tray app does).
	// For now, just return success.
	return nil
}

// PrintTest sends a test print request to the service
func (a *App) PrintTest(testType string) error {
	var endpoint string
	var method string
	var body io.Reader

	switch testType {
	case "comprehensive":
		endpoint = serviceURL + "/test"
		method = "GET"
	case "simple":
		endpoint = serviceURL + "/print"
		method = "POST"
		payload := map[string]interface{}{
			"header": "SIMPLE TEST",
			"items":  []interface{}{},
			"total":  0,
			"footer": fmt.Sprintf("PrintBridge Test\n%s", time.Now().Format("2006-01-02 15:04:05")),
		}
		data, _ := json.Marshal(payload)
		body = bytes.NewReader(data)
	default:
		return fmt.Errorf("unknown test type: %s", testType)
	}

	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest("GET", endpoint, nil)
	} else {
		req, err = http.NewRequest("POST", endpoint, body)
		req.Header.Set("Content-Type", "application/json")
	}
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("print failed: %s", string(bodyBytes))
	}

	return nil
}

// SendRaw sends raw data to the printer via the service
// It parses escape sequences like \x1B, \n, \r, \t before sending
func (a *App) SendRaw(data string) error {
	// Parse escape sequences
	parsed := parseEscapeSequences(data)
	
	payload := map[string]interface{}{
		"data": parsed,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := a.client.Post(serviceURL+"/raw", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send failed: %s", string(bodyBytes))
	}

	return nil
}

// parseEscapeSequences converts string escape sequences to actual bytes
// Supports: \x1B (hex), \n, \r, \t, \\
func parseEscapeSequences(s string) []byte {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'x', 'X':
				// Hex escape: \x1B
				if i+3 < len(s) {
					hex := s[i+2 : i+4]
					if b, err := parseHexByte(hex); err == nil {
						result = append(result, b)
						i += 4
						continue
					}
				}
			case 'n':
				result = append(result, '\n')
				i += 2
				continue
			case 'r':
				result = append(result, '\r')
				i += 2
				continue
			case 't':
				result = append(result, '\t')
				i += 2
				continue
			case '\\':
				result = append(result, '\\')
				i += 2
				continue
			}
		}
		result = append(result, s[i])
		i++
	}
	return result
}

// parseHexByte parses a 2-character hex string to a byte
func parseHexByte(hex string) (byte, error) {
	if len(hex) != 2 {
		return 0, fmt.Errorf("invalid hex length")
	}
	var b byte
	for _, c := range hex {
		b <<= 4
		switch {
		case c >= '0' && c <= '9':
			b |= byte(c - '0')
		case c >= 'a' && c <= 'f':
			b |= byte(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			b |= byte(c - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid hex char: %c", c)
		}
	}
	return b, nil
}

// Close is a no-op since we're using HTTP
func (a *App) Close() {
	// Nothing to close for HTTP client
}

// ConfigResponse represents the /config endpoint response
type ConfigResponse struct {
	Config     map[string]interface{} `json:"config"`
	ConfigPath string                 `json:"config_path"`
	ConfigDir  string                 `json:"config_dir"`
}

// GetConfig retrieves the current configuration from the service
func (a *App) GetConfig() (ConfigResponse, error) {
	var result ConfigResponse
	
	resp, err := a.client.Get(serviceURL + "/config")
	if err != nil {
		return result, fmt.Errorf("service not reachable: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to parse response: %v", err)
	}

	return result, nil
}

// UpdateConfig updates a configuration value via the service
func (a *App) UpdateConfig(key string, value interface{}) error {
	payload := map[string]interface{}{
		key: value,
	}
	jsonData, _ := json.Marshal(payload)

	resp, err := a.client.Post(serviceURL+"/config", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed: %s", string(bodyBytes))
	}

	return nil
}

// GetConfigPath returns the config file path from the service
func (a *App) GetConfigPath() (string, error) {
	result, err := a.GetConfig()
	if err != nil {
		return "", err
	}
	return result.ConfigPath, nil
}

