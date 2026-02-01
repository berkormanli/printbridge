package adapter

import (
	"log"
	"runtime"
)

// FindPrinters aggregates printers from all available sources (Windows Spooler, USB via SetupAPI).
func FindPrinters() ([]PrinterInfo, error) {
	var allPrinters []PrinterInfo

	if runtime.GOOS == "windows" {
		// 1. Windows Spooler Printers
		winPrinters, err := FindWindowsPrinters()
		if err != nil {
			log.Printf("[Discovery] Failed to list Windows printers: %v", err)
		} else {
			allPrinters = append(allPrinters, winPrinters...)
		}

		// 2. All USB Devices (via SetupAPI)
		usbDevices, err := FindAllUSBDevices()
		if err != nil {
			log.Printf("[Discovery] Failed to list USB devices: %v", err)
		} else {
			for _, dev := range usbDevices {
				allPrinters = append(allPrinters, PrinterInfo{
					VendorID:     dev.VendorID,
					ProductID:    dev.ProductID,
					Manufacturer: dev.Manufacturer,
					Product:      dev.Description,
					IsPrinter:    dev.IsPrinter,
					DeviceType:   "USB",
				})
			}
		}
	} else {
		// Non-Windows: use libusb-based discovery
		usbPrinters, err := FindUSBPrinters()
		if err != nil {
			log.Printf("[Discovery] Failed to list USB printers: %v", err)
		} else {
			for i := range usbPrinters {
				usbPrinters[i].DeviceType = "USB"
			}
			allPrinters = append(allPrinters, usbPrinters...)
		}
	}

	return allPrinters, nil
}

