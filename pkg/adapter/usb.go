//go:build cgo && !windows
// +build cgo,!windows

package adapter

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/gousb"
)

// USBAdapter communicates with USB receipt printers.
type USBAdapter struct {
	mu        sync.Mutex
	ctx       *gousb.Context
	device    *gousb.Device
	intf      *gousb.Interface
	outEP     *gousb.OutEndpoint
	inEP      *gousb.InEndpoint
	done      func()
	open      bool
	VendorID  uint16
	ProductID uint16
}



// NewUSBAdapter creates a new USB adapter.
// If vendorID and productID are 0, it will auto-detect the first printer.
func NewUSBAdapter(vendorID, productID uint16) *USBAdapter {
	return &USBAdapter{
		VendorID:  vendorID,
		ProductID: productID,
	}
}

// Open connects to the USB printer.
func (u *USBAdapter) Open() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.open {
		return nil
	}

	u.ctx = gousb.NewContext()

	var device *gousb.Device
	var err error

	if u.VendorID != 0 && u.ProductID != 0 {
		// Open specific device
		device, err = u.ctx.OpenDeviceWithVIDPID(gousb.ID(u.VendorID), gousb.ID(u.ProductID))
		if err != nil {
			u.ctx.Close()
			return fmt.Errorf("failed to open device %04x:%04x: %v", u.VendorID, u.ProductID, err)
		}
		if device == nil {
			u.ctx.Close()
			return fmt.Errorf("device %04x:%04x not found", u.VendorID, u.ProductID)
		}
	} else {
		// Auto-detect printer
		devices, err := u.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
			for _, cfg := range desc.Configs {
				for _, intf := range cfg.Interfaces {
					for _, alt := range intf.AltSettings {
						if alt.Class == gousb.ClassPrinter {
							return true
						}
					}
				}
			}
			return false
		})
		if err != nil {
			u.ctx.Close()
			return fmt.Errorf("failed to enumerate USB devices: %v", err)
		}
		if len(devices) == 0 {
			u.ctx.Close()
			return fmt.Errorf("no USB printer found")
		}
		device = devices[0]
		// Close extra devices
		for i := 1; i < len(devices); i++ {
			devices[i].Close()
		}
	}

	u.device = device

	// Set auto-detach kernel driver
	if err := u.device.SetAutoDetach(true); err != nil {
		// Not fatal, just log
	}

	// Get default interface
	intf, done, err := u.device.DefaultInterface()
	if err != nil {
		u.device.Close()
		u.ctx.Close()
		return fmt.Errorf("failed to claim interface: %v", err)
	}
	u.intf = intf
	u.done = done

	// Find OUT endpoint
	for _, ep := range intf.Setting.Endpoints {
		if ep.Direction == gousb.EndpointDirectionOut {
			u.outEP, err = intf.OutEndpoint(ep.Number)
			if err != nil {
				continue
			}
			break
		}
	}

	// Find IN endpoint (optional, for status)
	for _, ep := range intf.Setting.Endpoints {
		if ep.Direction == gousb.EndpointDirectionIn {
			u.inEP, err = intf.InEndpoint(ep.Number)
			if err != nil {
				continue
			}
			break
		}
	}

	if u.outEP == nil {
		u.done()
		u.device.Close()
		u.ctx.Close()
		return fmt.Errorf("no OUT endpoint found")
	}

	u.open = true
	return nil
}

// Write sends data to the printer.
func (u *USBAdapter) Write(data []byte) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.open {
		return fmt.Errorf("adapter not open")
	}

	_, err := u.outEP.Write(data)
	return err
}

// Read reads data from the printer.
func (u *USBAdapter) Read() ([]byte, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.open || u.inEP == nil {
		return nil, nil
	}

	buf := make([]byte, 64)
	n, err := u.inEP.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Close closes the USB connection.
func (u *USBAdapter) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.open {
		return nil
	}

	if u.done != nil {
		u.done()
	}
	if u.device != nil {
		u.device.Close()
	}
	if u.ctx != nil {
		u.ctx.Close()
	}

	u.open = false
	return nil
}

// IsOpen returns true if connected.
func (u *USBAdapter) IsOpen() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.open
}

// FindUSBPrinters returns a list of connected USB devices.
func FindUSBPrinters() ([]PrinterInfo, error) {
	log.Println("[USB] Starting USB device scan...")
	ctx := gousb.NewContext()
	defer ctx.Close()

	var devices []PrinterInfo

	// Collect device descriptors in the callback - we return false to avoid
	// having gousb try to open every device (which fails for system devices)
	_, _ = ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		vid := uint16(desc.Vendor)
		pid := uint16(desc.Product)
		
		// Check if device has printer class interface
		isPrinter := false
		for _, cfg := range desc.Configs {
			for _, intf := range cfg.Interfaces {
				for _, alt := range intf.AltSettings {
					if alt.Class == gousb.ClassPrinter {
						isPrinter = true
						break
					}
				}
				if isPrinter {
					break
				}
			}
			if isPrinter {
				break
			}
		}
		
		log.Printf("[USB] Found device: VID=%04X PID=%04X IsPrinter=%v", vid, pid, isPrinter)
		
		info := PrinterInfo{
			VendorID:  vid,
			ProductID: pid,
			IsPrinter: isPrinter,
		}
		devices = append(devices, info)
		
		// Return false - we don't want to actually open every device
		// as many will fail with LIBUSB_ERROR_NOT_SUPPORTED
		return false
	})

	log.Printf("[USB] Enumerated %d USB devices", len(devices))

	// Now try to get manufacturer/product strings for each device
	// by opening them individually (with error handling)
	for i := range devices {
		dev, err := ctx.OpenDeviceWithVIDPID(
			gousb.ID(devices[i].VendorID),
			gousb.ID(devices[i].ProductID),
		)
		if err != nil || dev == nil {
			log.Printf("[USB] Could not open VID=%04X PID=%04X for details (likely system device)",
				devices[i].VendorID, devices[i].ProductID)
			continue
		}

		if mfr, err := dev.Manufacturer(); err == nil {
			devices[i].Manufacturer = mfr
		}
		if prod, err := dev.Product(); err == nil {
			devices[i].Product = prod
		}
		log.Printf("[USB] Device details: VID=%04X PID=%04X Mfr=%q Product=%q IsPrinter=%v",
			devices[i].VendorID, devices[i].ProductID, devices[i].Manufacturer, devices[i].Product, devices[i].IsPrinter)
		dev.Close()
	}

	log.Printf("[USB] Returning %d devices", len(devices))
	return devices, nil
}


