package adapter

import (
	"fmt"
	"net"
	"time"
)

// NetworkAdapter communicates with network receipt printers (typically port 9100).
type NetworkAdapter struct {
	address string
	port    int
	timeout time.Duration
	conn    net.Conn
	open    bool
}

// NewNetworkAdapter creates a new network adapter.
func NewNetworkAdapter(address string, port int) *NetworkAdapter {
	if port == 0 {
		port = 9100 // Default printer port
	}
	return &NetworkAdapter{
		address: address,
		port:    port,
		timeout: 30 * time.Second,
	}
}

// Open connects to the network printer.
func (n *NetworkAdapter) Open() error {
	if n.open {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", n.address, n.port)
	conn, err := net.DialTimeout("tcp", addr, n.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	n.conn = conn
	n.open = true
	return nil
}

// Write sends data to the printer.
func (n *NetworkAdapter) Write(data []byte) error {
	if !n.open {
		return fmt.Errorf("adapter not open")
	}
	_, err := n.conn.Write(data)
	return err
}

// Read reads data from the printer.
func (n *NetworkAdapter) Read() ([]byte, error) {
	if !n.open {
		return nil, fmt.Errorf("adapter not open")
	}
	buf := make([]byte, 1024)
	n.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	num, err := n.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:num], nil
}

// Close closes the connection.
func (n *NetworkAdapter) Close() error {
	if !n.open {
		return nil
	}
	err := n.conn.Close()
	n.open = false
	return err
}

// IsOpen returns true if connected.
func (n *NetworkAdapter) IsOpen() bool {
	return n.open
}
