package adapter

// Adapter interface defines the contract for all printer adapters.
// This follows the adapter pattern from node-escpos for extensibility.
type Adapter interface {
	// Open establishes a connection to the printer
	Open() error

	// Write sends data to the printer
	Write(data []byte) error

	// Read reads data from the printer (for status checks)
	Read() ([]byte, error)

	// Close terminates the connection
	Close() error

	// IsOpen returns true if the adapter is connected
	IsOpen() bool
}

// PrinterInfo contains device details for discovery.
type PrinterInfo struct {
	VendorID     uint16 `json:"vendor_id"`
	ProductID    uint16 `json:"product_id"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	IsPrinter    bool   `json:"is_printer"`
	DeviceType   string `json:"device_type"` // "USB" or "Windows"
}
