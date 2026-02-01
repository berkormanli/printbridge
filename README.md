# PrintBridge

A cross-platform print bridge service for thermal/receipt printers with ESC/POS support. Built with Go and [Wails](https://wails.io/) for the desktop GUI.

![Version](https://img.shields.io/badge/version-1.0.7-blue)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- 🖨️ **Multi-Adapter Support** – Connect via USB, Windows Print Spooler, Network, or Serial
- 📋 **ESC/POS Commands** – Full support for thermal printer commands including text formatting, barcodes, QR codes, and images
- 🌐 **HTTP API** – RESTful endpoints for remote printing from any application
- 🖥️ **Desktop Dashboard** – Modern Svelte-based GUI for configuration and testing
- 🔧 **System Tray App** – Background service management with tray icon
- ⚙️ **Auto-Configuration** – Automatic printer detection and selection
- 📦 **Easy Installation** – Windows installer with all dependencies included

## Architecture

PrintBridge consists of three main components:

| Component | Description |
|-----------|-------------|
| **Service** (`printbridge_service.exe`) | HTTP server that handles print requests on port 9100 |
| **Dashboard** (`printbridge-gui.exe`) | Wails desktop app for configuration and testing |
| **Tray App** (`printbridge-tray.exe`) | System tray application for service management |

## Installation

### Windows Installer

Download the latest release from the [Releases](https://github.com/berkormanli/printbridge/releases) page and run the installer.

The installer will:
- Install all components to `Program Files\PrintBridge`
- Create configuration in `%APPDATA%\PrintBridge`
- Optionally add the tray app to Windows startup
- Install required dependencies (VC++ Runtime, libusb)

### Manual Build

#### Prerequisites

- Go 1.21+
- Node.js 18+ with pnpm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- For USB support: libusb-1.0

#### Build Commands

```bash
# Install frontend dependencies
cd frontend && pnpm install && cd ..

# Build all components
wails build                           # Dashboard GUI
go build -o printbridge_service.exe ./cmd/server    # Service
go build -o printbridge-tray.exe ./cmd/tray         # Tray app

# Or use the build script
./build.bat
```

## Configuration

Configuration is stored in `%APPDATA%\PrintBridge\config.json` (Windows) or `~/.config/PrintBridge/config.json` (Linux/macOS).

```json
{
  "host": "localhost",
  "port": 9100,
  "adapter": "auto",
  "usb": {
    "vendor_id": 0,
    "product_id": 0
  },
  "windows": {
    "printer_name": ""
  }
}
```

### Adapter Types

| Adapter | Description |
|---------|-------------|
| `auto` | Auto-detect based on OS (Windows → windows, others → usb) |
| `windows` | Use Windows Print Spooler |
| `usb` | Direct USB connection (requires libusb) |
| `console` | Debug mode - output to console |

## API Reference

The service exposes the following HTTP endpoints on `http://localhost:9100`:

### Health Check
```
GET /health
```
Returns service health status.

### Printer Status
```
GET /status
```
Returns printer connection status and list of available printers.

### Print Receipt
```
POST /print
Content-Type: application/json

{
  "header": "STORE NAME",
  "items": [
    {"name": "Item 1", "qty": 2, "price": 9.99}
  ],
  "total": 19.98,
  "footer": "Thank you!"
}
```

### Raw ESC/POS Print
```
POST /raw
Content-Type: application/json

{
  "data": [27, 64, 72, 101, 108, 108, 111]
}
```
Send raw ESC/POS bytes directly to the printer.

### Test Print
```
GET /test
```
Prints a comprehensive test receipt demonstrating all features.

### Configuration
```
GET /config
POST /config
Content-Type: application/json

{
  "adapter": "windows"
}
```
Get or update service configuration.

### Template Print (Food Delivery)
```
POST /print/template
Content-Type: application/json

{
  "platform": "Yemeksepeti",
  "merchant": {
    "name": "Lezzet Pide & Cafe",
    "district": "Esenyurt",
    "neighborhood": "Güzelyurt Mah."
  },
  "order": {
    "order_time": "2024-03-15T09:17:00",
    "order_type": "Food delivery"
  },
  "customer": {
    "name": "Yılmaz Öz",
    "address": {
      "neighborhood": "Güzelyurt Mah.",
      "street_address": "Cumhuriyet Parkı yanı",
      "floor": 2,
      "apartment": 4,
      "district": "Esenyurt",
      "city": "İstanbul",
      "description": "Parkın karşısı"
    },
    "phone": "+905551234567"
  },
  "items": [
    {
      "name": "Kıymalı Pide",
      "quantity": 1,
      "unit_price_try": 85.00,
      "total_price_try": 85.00
    },
    {
      "name": "Ayran",
      "quantity": 2,
      "unit_price_try": 15.00,
      "total_price_try": 30.00
    }
  ],
  "totals": {
    "subtotal_try": 115.00,
    "delivery_fee_try": 9.99,
    "vat": { "included": true },
    "total_try": 124.99
  },
  "payment": {
    "method": "Online Ödeme",
    "note": ""
  },
  "notes": {
    "customer_note": "Acı olmasın"
  }
}
```

**Supported platforms:** `Getir Yemek`, `Yemeksepeti`, `Trendyol Go`, `Migros Yemek`

The `platform` field auto-selects the branded logo and template styling.

## ESC/POS Command Reference

PrintBridge supports a comprehensive set of ESC/POS commands. Below are the raw byte buffers for all supported commands.

### Control Characters

| Name | Hex | Bytes |
|------|-----|-------|
| LF (Line Feed) | `0x0A` | `[10]` |
| CR (Carriage Return) | `0x0D` | `[13]` |
| ESC (Escape) | `0x1B` | `[27]` |
| GS (Group Separator) | `0x1D` | `[29]` |
| FS (Field Separator) | `0x1C` | `[28]` |

### Hardware Commands

```go
HW_INIT   = []byte{0x1b, 0x40}             // ESC @ - Initialize printer
HW_SELECT = []byte{0x1b, 0x3d, 0x01}       // ESC = 1 - Select printer
HW_RESET  = []byte{0x1b, 0x3f, 0x0a, 0x00} // Reset printer
```

### Feed Control

```go
CTL_LF = []byte{0x0a}  // Line feed
CTL_FF = []byte{0x0c}  // Form feed
CTL_CR = []byte{0x0d}  // Carriage return
CTL_HT = []byte{0x09}  // Horizontal tab
CTL_VT = []byte{0x0b}  // Vertical tab
```

### Text Formatting

```go
// Size
TXT_NORMAL   = []byte{0x1b, 0x21, 0x00}  // Normal text
TXT_2HEIGHT  = []byte{0x1b, 0x21, 0x10}  // Double height
TXT_2WIDTH   = []byte{0x1b, 0x21, 0x20}  // Double width
TXT_4SQUARE  = []byte{0x1b, 0x21, 0x30}  // Double width & height

// Custom size: GS ! n (width 1-8, height 1-8)
// Formula: n = (width-1)*16 + (height-1)
TxtCustomSize = []byte{0x1d, 0x21, n}

// Underline
TXT_UNDERL_OFF = []byte{0x1b, 0x2d, 0x00}  // Off
TXT_UNDERL_ON  = []byte{0x1b, 0x2d, 0x01}  // 1-dot underline
TXT_UNDERL2_ON = []byte{0x1b, 0x2d, 0x02}  // 2-dot underline

// Bold
TXT_BOLD_OFF = []byte{0x1b, 0x45, 0x00}  // Bold off
TXT_BOLD_ON  = []byte{0x1b, 0x45, 0x01}  // Bold on

// Italic
TXT_ITALIC_OFF = []byte{0x1b, 0x35}  // Italic off
TXT_ITALIC_ON  = []byte{0x1b, 0x34}  // Italic on

// Font selection
TXT_FONT_A = []byte{0x1b, 0x4d, 0x00}  // Font A (12x24)
TXT_FONT_B = []byte{0x1b, 0x4d, 0x01}  // Font B (9x17)
TXT_FONT_C = []byte{0x1b, 0x4d, 0x02}  // Font C

// Alignment
TXT_ALIGN_LT = []byte{0x1b, 0x61, 0x00}  // Left
TXT_ALIGN_CT = []byte{0x1b, 0x61, 0x01}  // Center
TXT_ALIGN_RT = []byte{0x1b, 0x61, 0x02}  // Right

// Reverse mode (white on black)
REVERSE_OFF = []byte{0x1d, 0x42, 0x00}  // Off
REVERSE_ON  = []byte{0x1d, 0x42, 0x01}  // On
```

### Line Spacing

```go
// Set line spacing to n/180 inch
SetLineSpacing(n) = []byte{0x1b, 0x33, n}

// Reset to default (1/6 inch)
LINE_SPACING_DEFAULT = []byte{0x1b, 0x32}

// Feed n dots
FeedDots(n) = []byte{0x1b, 0x4a, n}

// Feed n lines
FeedLines(n) = []byte{0x1b, 0x64, n}
```

### Paper Cutting

```go
PAPER_FULL_CUT = []byte{0x1d, 0x56, 0x00}  // Full cut
PAPER_PART_CUT = []byte{0x1d, 0x56, 0x01}  // Partial cut
```

### Cash Drawer

```go
CD_KICK_2 = []byte{0x1b, 0x70, 0x00, 0x19, 0x78}  // Kick pin 2
CD_KICK_5 = []byte{0x1b, 0x70, 0x01, 0x19, 0x78}  // Kick pin 5
```

### Barcode Commands

```go
// HRI (Human Readable Interpretation) position
BARCODE_TXT_OFF = []byte{0x1d, 0x48, 0x00}  // Off
BARCODE_TXT_ABV = []byte{0x1d, 0x48, 0x01}  // Above barcode
BARCODE_TXT_BLW = []byte{0x1d, 0x48, 0x02}  // Below barcode
BARCODE_TXT_BTH = []byte{0x1d, 0x48, 0x03}  // Both

// Height and width
BarcodeHeight(h) = []byte{0x1d, 0x68, h}   // h: 1-255
BarcodeWidth(w)  = []byte{0x1d, 0x77, w}   // w: 2-6

// Barcode types
BARCODE_UPC_A   = []byte{0x1d, 0x6b, 0x00}
BARCODE_UPC_E   = []byte{0x1d, 0x6b, 0x01}
BARCODE_EAN13   = []byte{0x1d, 0x6b, 0x02}
BARCODE_EAN8    = []byte{0x1d, 0x6b, 0x03}
BARCODE_CODE39  = []byte{0x1d, 0x6b, 0x04}
BARCODE_ITF     = []byte{0x1d, 0x6b, 0x05}
BARCODE_NW7     = []byte{0x1d, 0x6b, 0x06}
BARCODE_CODE93  = []byte{0x1d, 0x6b, 0x48}
BARCODE_CODE128 = []byte{0x1d, 0x6b, 0x49}
```

### QR Code Commands

```go
// Set QR model (append model number: 0x31=Model1, 0x32=Model2)
QR_MODEL = []byte{0x1d, 0x28, 0x6b, 0x04, 0x00, 0x31, 0x41, model}

// Set QR size (append size 1-16)
QR_SIZE = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x43, size}

// Set error correction (L=0x30, M=0x31, Q=0x32, H=0x33)
QR_ERROR = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x45, level}

// Store QR data: prefix + length + data
QR_STORE = []byte{0x1d, 0x28, 0x6b, pL, pH, 0x31, 0x50, 0x30} + data

// Print stored QR code
QR_PRINT = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x51, 0x30}
```

### Raster Bit Image

```go
// GS v 0 - Print raster bit image
// mode: 0=normal, 1=double width, 2=double height, 3=quadruple
// xL,xH: width in bytes (xL + xH*256)
// yL,yH: height in dots (yL + yH*256)
RasterImage = []byte{0x1d, 0x76, 0x30, mode, xL, xH, yL, yH} + imageData
```

### Character Sets & Code Pages

```go
// International character sets (ESC R n)
CHARSET_USA     = []byte{0x1b, 0x52, 0x00}
CHARSET_FRANCE  = []byte{0x1b, 0x52, 0x01}
CHARSET_GERMANY = []byte{0x1b, 0x52, 0x02}
CHARSET_UK      = []byte{0x1b, 0x52, 0x03}
CHARSET_SPAIN   = []byte{0x1b, 0x52, 0x07}
CHARSET_JAPAN   = []byte{0x1b, 0x52, 0x08}
CHARSET_KOREA   = []byte{0x1b, 0x52, 0x0d}

// Code pages (ESC t n)
CODEPAGE_PC437    = []byte{0x1b, 0x74, 0x00}  // USA/Europe
CODEPAGE_PC850    = []byte{0x1b, 0x74, 0x02}  // Multilingual
CODEPAGE_PC866    = []byte{0x1b, 0x74, 0x11}  // Cyrillic
CODEPAGE_WPC1252  = []byte{0x1b, 0x74, 0x10}  // Windows-1252
```

---

## Test Print Example

The `/test` endpoint prints a comprehensive receipt demonstrating all features. Here's what it tests:

### Test Receipt Structure

```
┌─────────────────────────────┐
│      PRINTBRIDGE            │  ← 2x size, bold, centered
│   COMPREHENSIVE TEST        │
│ ═══════════════════════════ │
│                             │
│    Sample Store Name        │  ← Store info
│    123 Main Street          │
│    City, State 12345        │
│    Tel: (555) 123-4567      │
│ ─────────────────────────── │
│ Date: 2024-01-15 14:30:00   │  ← Receipt details
│ Receipt #: TEST-001         │
│ Cashier: Demo User          │
│ ─────────────────────────── │
│ ITEMS                       │  ← Bold header
│ Espresso                    │
│   2 x $3.50 = $7.00         │
│ Cappuccino Large            │
│   1 x $4.75 = $4.75         │
│ ... more items ...          │
│ ─────────────────────────── │
│           Subtotal: $26.50  │  ← Right aligned
│           Tax (8%): $2.12   │
│           TOTAL: $28.62     │  ← Bold, 2x height
│ ─────────────────────────── │
│ Payment Method: CASH        │
│ Amount Tendered: $50.00     │
│ Change: $21.38              │
│ ═══════════════════════════ │
│                             │
│ === NEW FEATURES TEST ===   │
│                             │
│ 1. REVERSE MODE:            │
│ █ THIS TEXT IS REVERSED █   │  ← White on black
│                             │
│ 2. LINE SPACING TEST:       │
│    Tight spacing (10)       │
│    Wide spacing (60)        │
│                             │
│ 3. FEED CONTROL:            │
│    FeedDots(30)             │
│    FeedLines(2)             │
│                             │
│ 4. UNDERLINE MODES:         │
│    Single underline         │  ← 1-dot
│    Double underline         │  ← 2-dot
│                             │
│ 5. FONT SELECTION:          │
│    Font A (12x24 default)   │
│    Font B (9x17 smaller)    │
│                             │
│ 6. SIZE COMBINATIONS:       │
│    2x Width                 │
│    2x Height                │
│    2x Both                  │
│    3x                       │
│                             │
│ 7. RASTER IMAGE:            │
│    ▓▓▓▓▓▓▓▓ (checkerboard)  │
│                             │
│ 8. BARCODE TEST:            │
│    |||||||||||||||          │  ← CODE39
│    1234567890               │
│                             │
│ 9. QR CODE TYPES:           │
│    [QR] URL                 │
│    [QR] WiFi                │
│    [QR] Phone               │
│    [QR] High Error (H)      │
│                             │
│ Thank you for your visit!   │
│ *** END OF TEST ***         │
│                             │
│ Features Tested:            │
│ - Text alignment (L/C/R)    │
│ - Bold text                 │
│ - Double/triple size        │
│ - Underline modes           │
│ - Font A & B                │
│ - Reverse mode              │
│ - Line spacing              │
│ - Feed control              │
│ - Raster image              │
│ - Barcode printing          │
│ - QR code printing          │
│ - Paper cut                 │
│                             │
└─────────────────────────────┘
```

### Raw Byte Example

To print "Hello World" with initialization and paper cut:

```go
// Initialize + "Hello World" + Line Feed + Partial Cut
raw := []byte{
    0x1b, 0x40,                         // ESC @ (init)
    0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20,  // "Hello "
    0x57, 0x6f, 0x72, 0x6c, 0x64,        // "World"
    0x0a,                                // Line feed
    0x1d, 0x56, 0x01,                    // Partial cut
}
```

**API call:**
```bash
curl -X POST http://localhost:9100/raw \
  -H "Content-Type: application/json" \
  -d '{"data": [27,64,72,101,108,108,111,32,87,111,114,108,100,10,29,86,1]}'
```

---

## Development

### Live Development

```bash
# Run with hot reload
wails dev
```

This starts:
- Vite dev server for frontend with HMR
- Go backend with live reload
- Dev server at http://localhost:34115 for browser access

### Project Structure

```
PrintBridge/
├── cmd/
│   ├── server/          # HTTP service entry point
│   └── tray/            # System tray app
├── frontend/            # Svelte frontend (Wails)
│   └── src/
│       └── components/  # UI components
├── handlers/            # HTTP request handlers
├── pkg/
│   ├── adapter/         # Printer adapters (USB, Windows, Serial, Network)
│   ├── config/          # Configuration management
│   └── printer/         # ESC/POS printer implementation
├── installer/           # Inno Setup installer scripts
└── scripts/             # Build and utility scripts
```

## Building the Installer

```bash
# Build all components first
./build.bat

# Then compile the installer (requires Inno Setup 6.x)
./build-installer.bat
```

The installer will be output to `installer/output/PrintBridge-Setup-1.0.7.exe`.

## Dependencies

- [Wails](https://wails.io/) - Desktop app framework
- [gousb](https://github.com/google/gousb) - USB device access
- [libusb](https://libusb.info/) - Low-level USB support

## License

MIT License - see [LICENSE](LICENSE) for details.

## Author

**Berk Ormanlı** - [berkormanl@gmail.com](mailto:berkormanl@gmail.com)

---

*Built with ❤️ using Go and Wails*
