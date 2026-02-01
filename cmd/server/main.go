package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"printbridge/handlers"
	"printbridge/pkg/adapter"
	"printbridge/pkg/config"
)

func main() {
	// Load configuration from AppData or fallback locations
	configPath := config.GetConfigPath()
	log.Printf("Using config: %s", configPath)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create adapter based on config
	var adpt adapter.Adapter
	adapterType := cfg.Adapter

	// Auto-detect Windows if adapter not specified or is "auto"
	if adapterType == "" || adapterType == "auto" {
		if runtime.GOOS == "windows" {
			adapterType = "windows"
		} else {
			adapterType = "usb"
		}
	}

	switch adapterType {
	case "windows":
		printerName := cfg.Windows.PrinterName
		if printerName == "" {
			// Try to find the first available Windows printer
			printers, err := adapter.FindWindowsPrinters()
			if err == nil && len(printers) > 0 {
				printerName = printers[0].Product
				log.Printf("Auto-selected Windows printer: %s", printerName)
			}
		}
		if printerName == "" {
			log.Println("Warning: No Windows printer configured or found. Using console adapter.")
			adpt = adapter.NewConsoleAdapter()
		} else {
			adpt = adapter.NewWindowsPrinter(printerName)
		}

	case "usb":
		adpt = adapter.NewUSBAdapter(cfg.USB.VendorID, cfg.USB.ProductID)

	case "console":
		adpt = adapter.NewConsoleAdapter()

	default:
		log.Printf("Unknown adapter type '%s', using console", cfg.Adapter)
		adpt = adapter.NewConsoleAdapter()
	}

	// Open the adapter
	if err := adpt.Open(); err != nil {
		log.Printf("Warning: Failed to open adapter: %v", err)
		// Continue anyway - some endpoints don't require printer
	}
	defer adpt.Close()

	// Create print service with templates directory from AppData
	templatesDir := filepath.Join(config.GetConfigDir(), "templates")
	printService := handlers.NewPrintServiceWithTemplates(adpt, templatesDir)

	// Register HTTP handlers with CORS support
	http.HandleFunc("/health", cors(printService.HealthHandler))
	http.HandleFunc("/status", cors(printService.StatusHandler))
	http.HandleFunc("/print", cors(printService.PrintHandler))
	http.HandleFunc("/print/template", cors(printService.TemplatePrintHandler))
	http.HandleFunc("/raw", cors(printService.RawPrintHandler))
	http.HandleFunc("/test", cors(printService.TestPrintHandler))
	
	// Config endpoints
	http.HandleFunc("/config", cors(handleConfig))

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("PrintBridge service starting on %s (adapter: %s)", addr, adapterType)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// cors wraps an HTTP handler with CORS headers
func cors(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// handleConfig handles GET/POST requests for config
func handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodGet:
		cfg, err := config.Load()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
			return
		}
		
		response := map[string]interface{}{
			"config":      cfg,
			"config_path": config.GetConfigPath(),
			"config_dir":  config.GetConfigDir(),
		}
		
		data, _ := json.Marshal(response)
		w.Write(data)
		
	case http.MethodPost:
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid JSON: %v"}`, err), http.StatusBadRequest)
			return
		}
		
		for key, value := range updates {
			if err := config.Update(key, value); err != nil {
				http.Error(w, fmt.Sprintf(`{"error": "Failed to update %s: %v"}`, key, err), http.StatusInternalServerError)
				return
			}
		}
		
		w.Write([]byte(`{"status": "ok", "message": "Config updated. Restart service to apply changes."}`))
		
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

