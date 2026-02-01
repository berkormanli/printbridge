package adapter

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modwinspool = windows.NewLazySystemDLL("winspool.drv")

	procOpenPrinterW      = modwinspool.NewProc("OpenPrinterW")
	procClosePrinter      = modwinspool.NewProc("ClosePrinter")
	procStartDocPrinterW  = modwinspool.NewProc("StartDocPrinterW")
	procStartPagePrinter  = modwinspool.NewProc("StartPagePrinter")
	procWritePrinter      = modwinspool.NewProc("WritePrinter")
	procEndPagePrinter    = modwinspool.NewProc("EndPagePrinter")
	procEndDocPrinter     = modwinspool.NewProc("EndDocPrinter")
	procEnumPrintersW     = modwinspool.NewProc("EnumPrintersW")
)

// WindowsPrinter adapters for Windows Spooler API
type WindowsPrinter struct {
	handle windows.Handle
	name   string
}

func NewWindowsPrinter(name string) *WindowsPrinter {
	return &WindowsPrinter{name: name}
}

func (w *WindowsPrinter) Open() error {
	var h windows.Handle
	namePtr, err := syscall.UTF16PtrFromString(w.name)
	if err != nil {
		return err
	}

	// BOOL OpenPrinterW(LPWSTR pPrinterName, LPHANDLE phPrinter, LPPRINTER_DEFAULTSW pDefault);
	r1, _, e1 := procOpenPrinterW.Call(
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&h)),
		0,
	)
	if r1 == 0 {
		return fmt.Errorf("OpenPrinterW failed: %v", e1)
	}
	w.handle = h
	return nil
}

func (w *WindowsPrinter) Write(data []byte) error {
	if w.handle == 0 {
		return fmt.Errorf("printer not open")
	}

	// StartDoc
	docName, _ := syscall.UTF16PtrFromString("PrintBridge Raw Data")
	dataType, _ := syscall.UTF16PtrFromString("RAW")
	di := DOC_INFO_1{
		pDocName:    docName,
		pOutputFile: nil,
		pDatatype:   dataType,
	}

	// DWORD StartDocPrinterW(HANDLE hPrinter, DWORD Level, LPBYTE pDocInfo);
	r1, _, e1 := procStartDocPrinterW.Call(
		uintptr(w.handle),
		1,
		uintptr(unsafe.Pointer(&di)),
	)
	if r1 == 0 {
		return fmt.Errorf("StartDocPrinterW failed: %v", e1)
	}
	// jobID := r1

	// StartPage
	r1, _, e1 = procStartPagePrinter.Call(uintptr(w.handle))
	if r1 == 0 {
		procEndDocPrinter.Call(uintptr(w.handle))
		return fmt.Errorf("StartPagePrinter failed: %v", e1)
	}

	// WritePrinter
	var written uint32
	r1, _, e1 = procWritePrinter.Call(
		uintptr(w.handle),
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		uintptr(unsafe.Pointer(&written)),
	)
	if r1 == 0 {
		procEndPagePrinter.Call(uintptr(w.handle))
		procEndDocPrinter.Call(uintptr(w.handle))
		return fmt.Errorf("WritePrinter failed: %v", e1)
	}

	// EndPage
	procEndPagePrinter.Call(uintptr(w.handle))
	
	// EndDoc
	procEndDocPrinter.Call(uintptr(w.handle))

	return nil
}

func (w *WindowsPrinter) Read() ([]byte, error) {
	// Reading from a raw Windows printer handle is not typically supported 
	// or requires bidirectional communication setup. Returning nil for now.
	return nil, nil
}

func (w *WindowsPrinter) Close() error {
	if w.handle != 0 {
		procClosePrinter.Call(uintptr(w.handle))
		w.handle = 0
	}
	return nil
}

func (w *WindowsPrinter) IsOpen() bool {
	return w.handle != 0
}

type DOC_INFO_1 struct {
	pDocName    *uint16
	pOutputFile *uint16
	pDatatype   *uint16
}

type PRINTER_INFO_4 struct {
	pPrinterName *uint16
	pServerName  *uint16
	Attributes   uint32
}

const (
	PRINTER_ENUM_LOCAL       = 0x00000002
	PRINTER_ENUM_CONNECTIONS = 0x00000004
)

// FindWindowsPrinters enumerates all local and network printers.
func FindWindowsPrinters() ([]PrinterInfo, error) {
	flags := uintptr(PRINTER_ENUM_LOCAL | PRINTER_ENUM_CONNECTIONS)
	var needed, returned uint32

	// First call to get size
	procEnumPrintersW.Call(
		flags,
		0,
		4, // Level 4 is usually safe and fast
		0,
		0,
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)

	if needed == 0 {
		return []PrinterInfo{}, nil
	}

	buffer := make([]byte, needed)
	r1, _, e1 := procEnumPrintersW.Call(
		flags,
		0,
		4,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(needed),
		uintptr(unsafe.Pointer(&needed)),
		uintptr(unsafe.Pointer(&returned)),
	)
	
	if r1 == 0 {
		return nil, fmt.Errorf("EnumPrintersW failed: %v", e1)
	}

	var printers []PrinterInfo
	for i := 0; i < int(returned); i++ {
		// Calculate offset of the i-th struct
		// In C: pPrinterEnum + i
		// In Go, we need unsafe pointer arithmetic
		// Size of PRINTER_INFO_4 is roughly 2 pointers + 1 uint32 = 16 (on 64-bit) or 12 (32-bit).
		// We should rely on unsafe.Sizeof
		
		// Wait, manual parsing is risky. Let's cast to slice of struct if possible, but they contain pointers.
		// It's safer to iterate via unsafe.Pointer
		
		// PRINTER_INFO_4W: pPrinterName (ptr), pServerName (ptr), Attributes (uint32)
		// On 64-bit: 8 + 8 + 4 = 20 bytes + padding = 24 bytes?
		// We'll trust unsafe.Sizeof(PRINTER_INFO_4{}) to be correct for the arch.
	}
	
	// Re-implement loop using unsafe to be correct
	pInfos := (*[1024]PRINTER_INFO_4)(unsafe.Pointer(&buffer[0]))[:returned:returned]
	
	for _, info := range pInfos {
		name := windows.UTF16PtrToString(info.pPrinterName)
		log.Printf("Found printer: %s", name)
		// Add to list
		printers = append(printers, PrinterInfo{
			VendorID:     0, // VIDs not available via Spooler API usually
			ProductID:    0,
			Manufacturer: "Windows Printer",
			Product:      name,
			IsPrinter:    true,
			DeviceType:   "Windows",
		})
	}

	return printers, nil
}
