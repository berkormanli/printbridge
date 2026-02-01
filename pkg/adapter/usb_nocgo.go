//go:build !cgo || windows
// +build !cgo windows

package adapter

import (
	"fmt"
)

// USBAdapter stub for non-CGO builds (Windows cross-compile)
// USB support requires native build with CGO enabled
type USBAdapter struct {
	VendorID  uint16
	ProductID uint16
}

func NewUSBAdapter(vendorID, productID uint16) *USBAdapter {
	return &USBAdapter{
		VendorID:  vendorID,
		ProductID: productID,
	}
}

func (u *USBAdapter) Open() error {
	return fmt.Errorf("USB adapter not available: requires native build with CGO. Use 'network' or 'console' adapter instead")
}

func (u *USBAdapter) Write(data []byte) error {
	return fmt.Errorf("USB adapter not available")
}

func (u *USBAdapter) Read() ([]byte, error) {
	return nil, fmt.Errorf("USB adapter not available")
}

func (u *USBAdapter) Close() error {
	return nil
}

func (u *USBAdapter) IsOpen() bool {
	return false
}

// FindUSBPrinters stub - returns empty list on non-CGO builds
func FindUSBPrinters() ([]PrinterInfo, error) {
	return nil, fmt.Errorf("USB printer discovery not available: requires native build with CGO")
}
