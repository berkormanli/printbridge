package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Config represents the application configuration.
type Config struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Adapter string `json:"adapter"` // usb, windows, network, serial, console, auto

	AutoStart struct {
		Enabled          bool `json:"enabled"`
		InstallOnStartup bool `json:"install_on_startup"`
	} `json:"autostart"`

	USB struct {
		VendorID  uint16 `json:"vendor_id"`
		ProductID uint16 `json:"product_id"`
	} `json:"usb"`

	Windows struct {
		PrinterName string `json:"printer_name"`
	} `json:"windows"`

	Network struct {
		Address string `json:"address"`
		Port    int    `json:"port"`
	} `json:"network"`

	Serial struct {
		Port     string `json:"port"`
		BaudRate int    `json:"baud_rate"`
	} `json:"serial"`
}

var (
	configPath string
	configOnce sync.Once
)

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Host:    "0.0.0.0",
		Port:    9100,
		Adapter: "auto",
	}
}

// GetConfigDir returns the PrintBridge config directory path.
// On Windows: %APPDATA%/PrintBridge
// On Linux/Mac: ~/.config/printbridge
func GetConfigDir() string {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "PrintBridge")
		}
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(configDir, "PrintBridge")
}

// GetConfigPath returns the full path to config.json.
// Priority:
// 1. PRINTBRIDGE_CONFIG environment variable
// 2. %APPDATA%/PrintBridge/config.json (or ~/.config/PrintBridge/)
// 3. Executable directory (portable mode)
// 4. Current working directory
func GetConfigPath() string {
	configOnce.Do(func() {
		// 1. Check environment variable
		if path := os.Getenv("PRINTBRIDGE_CONFIG"); path != "" {
			configPath = path
			return
		}

		// 2. AppData directory (primary)
		appDataConfig := filepath.Join(GetConfigDir(), "config.json")
		if _, err := os.Stat(appDataConfig); err == nil {
			configPath = appDataConfig
			return
		}

		// 3. Executable directory (portable mode)
		exe, _ := os.Executable()
		if exe != "" {
			exeDir := filepath.Dir(exe)
			exeConfig := filepath.Join(exeDir, "config.json")
			if _, err := os.Stat(exeConfig); err == nil {
				configPath = exeConfig
				return
			}
		}

		// 4. Current working directory
		if _, err := os.Stat("config.json"); err == nil {
			configPath = "config.json"
			return
		}

		// Default to AppData (will be created if not exists)
		configPath = appDataConfig
	})

	return configPath
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir := GetConfigDir()
	return os.MkdirAll(dir, 0755)
}

// Load loads the configuration from the default path.
func Load() (*Config, error) {
	return LoadFrom(GetConfigPath())
}

// LoadFrom loads configuration from a specific path.
func LoadFrom(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			if err := EnsureConfigDir(); err != nil {
				return config, nil
			}
			if err := SaveTo(path, config); err != nil {
				return config, nil
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

// Save saves the configuration to the default path.
func Save(config *Config) error {
	return SaveTo(GetConfigPath(), config)
}

// SaveTo saves configuration to a specific path.
func SaveTo(path string, config *Config) error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Update updates a specific field in the config and saves.
func Update(key string, value interface{}) error {
	config, err := Load()
	if err != nil {
		return err
	}

	switch key {
	case "host":
		if v, ok := value.(string); ok {
			config.Host = v
		}
	case "port":
		if v, ok := value.(float64); ok {
			config.Port = int(v)
		}
	case "adapter":
		if v, ok := value.(string); ok {
			config.Adapter = v
		}
	case "windows.printer_name":
		if v, ok := value.(string); ok {
			config.Windows.PrinterName = v
		}
	case "usb.vendor_id":
		if v, ok := value.(float64); ok {
			config.USB.VendorID = uint16(v)
		}
	case "usb.product_id":
		if v, ok := value.(float64); ok {
			config.USB.ProductID = uint16(v)
		}
	}

	return Save(config)
}
