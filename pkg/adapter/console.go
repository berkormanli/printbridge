package adapter

import (
	"fmt"
	"os"
)

// ConsoleAdapter is a testing adapter that prints to stdout.
// Useful for development and debugging without a physical printer.
type ConsoleAdapter struct {
	open bool
}

// NewConsoleAdapter creates a new console adapter.
func NewConsoleAdapter() *ConsoleAdapter {
	return &ConsoleAdapter{}
}

// Open simulates opening a connection.
func (c *ConsoleAdapter) Open() error {
	c.open = true
	fmt.Println("[ConsoleAdapter] Opened")
	return nil
}

// Write prints data to stdout in a readable format.
func (c *ConsoleAdapter) Write(data []byte) error {
	if !c.open {
		return fmt.Errorf("adapter not open")
	}

	// Print raw bytes in hex for ESC/POS commands, text otherwise
	fmt.Fprintf(os.Stdout, "[PRINT] %s", string(data))
	return nil
}

// Read returns empty data (console doesn't support reading).
func (c *ConsoleAdapter) Read() ([]byte, error) {
	return nil, nil
}

// Close simulates closing the connection.
func (c *ConsoleAdapter) Close() error {
	c.open = false
	fmt.Println("[ConsoleAdapter] Closed")
	return nil
}

// IsOpen returns true if the adapter is open.
func (c *ConsoleAdapter) IsOpen() bool {
	return c.open
}
