package adapter

import (
	"fmt"
	"regexp"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modSetupAPI = windows.NewLazySystemDLL("setupapi.dll")

	procSetupDiGetClassDevsW         = modSetupAPI.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo        = modSetupAPI.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceRegistryPropertyW = modSetupAPI.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiDestroyDeviceInfoList = modSetupAPI.NewProc("SetupDiDestroyDeviceInfoList")
	procSetupDiGetDeviceInstanceIdW  = modSetupAPI.NewProc("SetupDiGetDeviceInstanceIdW")
)

// GUID for all USB devices
var GUID_DEVINTERFACE_USB_DEVICE = windows.GUID{
	Data1: 0xA5DCBF10,
	Data2: 0x6530,
	Data3: 0x11D2,
	Data4: [8]byte{0x90, 0x1F, 0x00, 0xC0, 0x4F, 0xB9, 0x51, 0xED},
}

const (
	DIGCF_PRESENT         = 0x00000002
	DIGCF_ALLCLASSES      = 0x00000004
	DIGCF_DEVICEINTERFACE = 0x00000010

	SPDRP_DEVICEDESC         = 0x00000000
	SPDRP_HARDWAREID         = 0x00000001
	SPDRP_COMPATIBLEIDS      = 0x00000002
	SPDRP_CLASS              = 0x00000007
	SPDRP_CLASSGUID          = 0x00000008
	SPDRP_DRIVER             = 0x00000009
	SPDRP_MFG                = 0x0000000B
	SPDRP_FRIENDLYNAME       = 0x0000000C
	SPDRP_LOCATION_INFORMATION = 0x0000000D
	SPDRP_ENUMERATOR_NAME    = 0x00000016

	INVALID_HANDLE_VALUE = ^uintptr(0)
)

type SP_DEVINFO_DATA struct {
	CbSize    uint32
	ClassGuid windows.GUID
	DevInst   uint32
	Reserved  uintptr
}

// USBDeviceInfo represents a USB device discovered via SetupAPI
type USBDeviceInfo struct {
	VendorID     uint16 `json:"vendor_id"`
	ProductID    uint16 `json:"product_id"`
	Description  string `json:"description"`
	Manufacturer string `json:"manufacturer"`
	DeviceClass  string `json:"device_class"`
	InstanceID   string `json:"instance_id"`
	IsPrinter    bool   `json:"is_printer"`
}

// FindAllUSBDevices enumerates all USB devices using Windows SetupAPI
func FindAllUSBDevices() ([]USBDeviceInfo, error) {
	// Get device info set for all present devices
	hDevInfo, _, err := procSetupDiGetClassDevsW.Call(
		0, // No class GUID - enumerate all
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("USB"))), // Enumerator: USB
		0,
		DIGCF_PRESENT|DIGCF_ALLCLASSES,
	)

	if hDevInfo == INVALID_HANDLE_VALUE {
		return nil, fmt.Errorf("SetupDiGetClassDevsW failed: %v", err)
	}
	defer procSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	var devices []USBDeviceInfo
	var devInfoData SP_DEVINFO_DATA
	devInfoData.CbSize = uint32(unsafe.Sizeof(devInfoData))

	for i := uint32(0); ; i++ {
		r1, _, _ := procSetupDiEnumDeviceInfo.Call(
			hDevInfo,
			uintptr(i),
			uintptr(unsafe.Pointer(&devInfoData)),
		)
		if r1 == 0 {
			break // No more devices
		}

		device := USBDeviceInfo{}

		// Get device description
		device.Description = getDeviceRegistryProperty(hDevInfo, &devInfoData, SPDRP_DEVICEDESC)
		if device.Description == "" {
			device.Description = getDeviceRegistryProperty(hDevInfo, &devInfoData, SPDRP_FRIENDLYNAME)
		}

		// Get manufacturer
		device.Manufacturer = getDeviceRegistryProperty(hDevInfo, &devInfoData, SPDRP_MFG)

		// Get device class
		device.DeviceClass = getDeviceRegistryProperty(hDevInfo, &devInfoData, SPDRP_CLASS)

		// Get instance ID (contains VID/PID)
		device.InstanceID = getDeviceInstanceID(hDevInfo, &devInfoData)

		// Parse VID/PID from instance ID (format: USB\VID_XXXX&PID_XXXX\...)
		device.VendorID, device.ProductID = parseVIDPID(device.InstanceID)

		// Check if it's a printer
		device.IsPrinter = (device.DeviceClass == "Printer" || device.DeviceClass == "USB Printing Support")

		// Skip devices without VID/PID (hubs, controllers, etc.)
		if device.VendorID == 0 && device.ProductID == 0 {
			continue
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func getDeviceRegistryProperty(hDevInfo uintptr, devInfoData *SP_DEVINFO_DATA, property uint32) string {
	var dataType uint32
	var buffer [512]uint16
	var requiredSize uint32

	r1, _, _ := procSetupDiGetDeviceRegistryPropertyW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(devInfoData)),
		uintptr(property),
		uintptr(unsafe.Pointer(&dataType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)*2),
		uintptr(unsafe.Pointer(&requiredSize)),
	)

	if r1 == 0 {
		return ""
	}

	return syscall.UTF16ToString(buffer[:])
}

func getDeviceInstanceID(hDevInfo uintptr, devInfoData *SP_DEVINFO_DATA) string {
	var buffer [256]uint16
	var requiredSize uint32

	r1, _, _ := procSetupDiGetDeviceInstanceIdW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(devInfoData)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&requiredSize)),
	)

	if r1 == 0 {
		return ""
	}

	return syscall.UTF16ToString(buffer[:])
}

// parseVIDPID extracts VID and PID from instance ID string
// Example: "USB\VID_1234&PID_5678\123456789" -> 0x1234, 0x5678
func parseVIDPID(instanceID string) (uint16, uint16) {
	re := regexp.MustCompile(`VID_([0-9A-Fa-f]{4})&PID_([0-9A-Fa-f]{4})`)
	matches := re.FindStringSubmatch(instanceID)
	if len(matches) != 3 {
		return 0, 0
	}

	vid, _ := strconv.ParseUint(matches[1], 16, 16)
	pid, _ := strconv.ParseUint(matches[2], 16, 16)

	return uint16(vid), uint16(pid)
}
