package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration.
type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Adapter string `json:"adapter"` // usb, network, serial, console

	AutoStart struct {
		Enabled          bool `json:"enabled"`
		InstallOnStartup bool `json:"install_on_startup"`
	} `json:"autostart"`

	USB struct {
		VendorID  uint16 `json:"vendor_id"`
		ProductID uint16 `json:"product_id"`
	} `json:"usb"`

	Network struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	} `json:"network"`

	Serial struct {
		Port     string `json:"port"`
		BaudRate int    `json:"baud_rate"`
	} `json:"serial"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:    "0.0.0.0",
		Port:    9100,
		Adapter: "console", // Safe default for testing
	}
}

// LoadConfig loads configuration from a file.
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file
			if err := SaveConfig(path, config); err != nil {
				return nil, err
			}
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig saves configuration to a file.
func SaveConfig(path string, config *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetConfigPath returns the default config file path.
func GetConfigPath() string {
	// Check environment variable first
	if path := os.Getenv("PRINTBRIDGE_CONFIG"); path != "" {
		return path
	}

	// Check current directory
	if _, err := os.Stat("config.json"); err == nil {
		return "config.json"
	}

	// Default to user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "config.json"
	}

	return filepath.Join(configDir, "printbridge", "config.json")
}

// UpdateUSBDevice updates the USB vendor and product IDs in the config file.
func UpdateUSBDevice(configPath string, vendorID, productID uint16) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	config.USB.VendorID = vendorID
	config.USB.ProductID = productID

	// Ensure adapter is set to USB when selecting a device
	if vendorID != 0 || productID != 0 {
		config.Adapter = "usb"
	}

	return SaveConfig(configPath, config)
}
