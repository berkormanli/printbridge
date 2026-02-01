package printer

import (
	"fmt"

	"printbridge/pkg/adapter"
)

// Printer provides a fluent API for building ESC/POS print jobs.
type Printer struct {
	adapter  adapter.Adapter
	buffer   []byte
	encoding string
	width    int
}

// New creates a new Printer with the given adapter.
func New(a adapter.Adapter) *Printer {
	return &Printer{
		adapter:  a,
		buffer:   make([]byte, 0, 1024),
		encoding: "UTF-8",
		width:    48, // Default character width for 80mm paper
	}
}

// Init initializes the printer.
func (p *Printer) Init() *Printer {
	p.buffer = append(p.buffer, HW_INIT...)
	return p
}

// Text adds text to the buffer.
func (p *Printer) Text(content string) *Printer {
	p.buffer = append(p.buffer, []byte(content)...)
	return p
}

// Println adds text with a newline.
func (p *Printer) Println(content string) *Printer {
	p.buffer = append(p.buffer, []byte(content+EOL)...)
	return p
}

// NewLine adds a line feed.
func (p *Printer) NewLine() *Printer {
	p.buffer = append(p.buffer, CTL_LF...)
	return p
}

// Feed adds multiple line feeds.
func (p *Printer) Feed(n int) *Printer {
	for i := 0; i < n; i++ {
		p.buffer = append(p.buffer, CTL_LF...)
	}
	return p
}

// Align sets text alignment.
func (p *Printer) Align(align string) *Printer {
	switch align {
	case "left", "lt", "LT":
		p.buffer = append(p.buffer, TXT_ALIGN_LT...)
	case "center", "ct", "CT":
		p.buffer = append(p.buffer, TXT_ALIGN_CT...)
	case "right", "rt", "RT":
		p.buffer = append(p.buffer, TXT_ALIGN_RT...)
	}
	return p
}

// Bold sets bold mode.
func (p *Printer) Bold(on bool) *Printer {
	if on {
		p.buffer = append(p.buffer, TXT_BOLD_ON...)
	} else {
		p.buffer = append(p.buffer, TXT_BOLD_OFF...)
	}
	return p
}

// Underline sets underline mode.
func (p *Printer) Underline(mode int) *Printer {
	switch mode {
	case 0:
		p.buffer = append(p.buffer, TXT_UNDERL_OFF...)
	case 1:
		p.buffer = append(p.buffer, TXT_UNDERL_ON...)
	case 2:
		p.buffer = append(p.buffer, TXT_UNDERL2_ON...)
	}
	return p
}

// Size sets custom text size (1-8 for width and height).
func (p *Printer) Size(width, height int) *Printer {
	p.buffer = append(p.buffer, TxtCustomSize(width, height)...)
	return p
}

// Font sets the font type.
func (p *Printer) Font(font string) *Printer {
	switch font {
	case "a", "A":
		p.buffer = append(p.buffer, TXT_FONT_A...)
		p.width = 48
	case "b", "B":
		p.buffer = append(p.buffer, TXT_FONT_B...)
		p.width = 64
	case "c", "C":
		p.buffer = append(p.buffer, TXT_FONT_C...)
		p.width = 64
	}
	return p
}

// Normal resets text formatting.
func (p *Printer) Normal() *Printer {
	p.buffer = append(p.buffer, TXT_NORMAL...)
	return p
}

// DrawLine prints a line of characters.
func (p *Printer) DrawLine(char string) *Printer {
	if char == "" {
		char = "-"
	}
	for i := 0; i < p.width; i++ {
		p.buffer = append(p.buffer, []byte(char)...)
	}
	return p.NewLine()
}

// Cut cuts the paper.
func (p *Printer) Cut(partial bool) *Printer {
	p.Feed(3)
	if partial {
		p.buffer = append(p.buffer, PAPER_PART_CUT...)
	} else {
		p.buffer = append(p.buffer, PAPER_FULL_CUT...)
	}
	return p
}

// CashDraw kicks the cash drawer.
func (p *Printer) CashDraw(pin int) *Printer {
	if pin == 5 {
		p.buffer = append(p.buffer, CD_KICK_5...)
	} else {
		p.buffer = append(p.buffer, CD_KICK_2...)
	}
	return p
}

// Beep makes the printer beep.
func (p *Printer) Beep(times, duration int) *Printer {
	p.buffer = append(p.buffer, BEEP...)
	p.buffer = append(p.buffer, byte(times), byte(duration))
	return p
}

// Barcode prints a barcode.
func (p *Printer) Barcode(code string, barcodeType string, width, height int) *Printer {
	p.buffer = append(p.buffer, BARCODE_TXT_BLW...)
	p.buffer = append(p.buffer, BARCODE_FONT_A...)
	p.buffer = append(p.buffer, BarcodeHeight(height)...)
	p.buffer = append(p.buffer, BarcodeWidth(width)...)

	switch barcodeType {
	case "UPC_A", "UPC-A":
		p.buffer = append(p.buffer, BARCODE_UPC_A...)
	case "UPC_E", "UPC-E":
		p.buffer = append(p.buffer, BARCODE_UPC_E...)
	case "EAN13":
		p.buffer = append(p.buffer, BARCODE_EAN13...)
	case "EAN8":
		p.buffer = append(p.buffer, BARCODE_EAN8...)
	case "CODE39":
		p.buffer = append(p.buffer, BARCODE_CODE39...)
	case "CODE128":
		p.buffer = append(p.buffer, BARCODE_CODE128...)
	default:
		p.buffer = append(p.buffer, BARCODE_CODE39...)
	}

	p.buffer = append(p.buffer, []byte(code)...)
	p.buffer = append(p.buffer, 0x00)
	return p
}

// QRCode prints a QR code with default settings (Model 2, Error Level L).
func (p *Printer) QRCode(content string, size int) *Printer {
	return p.QRCodeAdvanced(content, size, QRErrorL, QRModel2)
}

// QR Code Error Correction Levels
const (
	QRErrorL = 48 // ~7% recovery - for clean environments
	QRErrorM = 49 // ~15% recovery - standard
	QRErrorQ = 50 // ~25% recovery - better for printing
	QRErrorH = 51 // ~30% recovery - best for damaged/dirty conditions
)

// QR Code Models
const (
	QRModel1 = 49 // Model 1 - Original, smaller capacity
	QRModel2 = 50 // Model 2 - Enhanced, recommended
)

// QRCodeAdvanced prints a QR code with full control over settings.
// size: 1-16 (module size in dots)
// errorLevel: QRErrorL/M/Q/H (error correction level)
// model: QRModel1 or QRModel2
func (p *Printer) QRCodeAdvanced(content string, size int, errorLevel int, model int) *Printer {
	if size < 1 {
		size = 6
	}
	if size > 16 {
		size = 16
	}

	// Set QR model
	p.buffer = append(p.buffer, QR_MODEL...)
	p.buffer = append(p.buffer, byte(model), 0x00)

	// Set QR size
	p.buffer = append(p.buffer, QR_SIZE...)
	p.buffer = append(p.buffer, byte(size))

	// Set error correction level
	p.buffer = append(p.buffer, QR_ERROR...)
	p.buffer = append(p.buffer, byte(errorLevel))

	// Store data
	data := []byte(content)
	storeLen := len(data) + 3
	pL := byte(storeLen % 256)
	pH := byte(storeLen / 256)
	p.buffer = append(p.buffer, QR_STORE_PRE...)
	p.buffer = append(p.buffer, pL, pH)
	p.buffer = append(p.buffer, QR_STORE_POST...)
	p.buffer = append(p.buffer, data...)

	// Print QR
	p.buffer = append(p.buffer, QR_PRINT...)

	return p
}

// ============== QR CODE DATA TYPE HELPERS ==============
// These format data correctly for different QR code types

// QRCodeURL prints a URL QR code (just a regular QR with URL content).
func (p *Printer) QRCodeURL(url string, size int) *Printer {
	// URLs work best with higher error correction for scanning
	return p.QRCodeAdvanced(url, size, QRErrorM, QRModel2)
}

// QRCodeWiFi prints a WiFi credential QR code.
// authType: "WPA", "WEP", or "nopass" (open network)
// hidden: true if the network is hidden
func (p *Printer) QRCodeWiFi(ssid, password, authType string, hidden bool) *Printer {
	// Format: WIFI:T:<auth>;S:<ssid>;P:<password>;H:<hidden>;;
	hiddenStr := "false"
	if hidden {
		hiddenStr = "true"
	}
	content := "WIFI:T:" + authType + ";S:" + ssid + ";P:" + password + ";H:" + hiddenStr + ";;"
	return p.QRCodeAdvanced(content, 6, QRErrorM, QRModel2)
}

// QRCodeVCard prints a contact vCard QR code.
func (p *Printer) QRCodeVCard(name, phone, email, org string) *Printer {
	// Simple vCard 3.0 format
	content := "BEGIN:VCARD\n" +
		"VERSION:3.0\n" +
		"N:" + name + "\n" +
		"TEL:" + phone + "\n" +
		"EMAIL:" + email + "\n" +
		"ORG:" + org + "\n" +
		"END:VCARD"
	return p.QRCodeAdvanced(content, 5, QRErrorM, QRModel2)
}

// QRCodeSMS prints an SMS QR code.
func (p *Printer) QRCodeSMS(phone, message string) *Printer {
	content := "smsto:" + phone + ":" + message
	return p.QRCodeAdvanced(content, 6, QRErrorM, QRModel2)
}

// QRCodeEmail prints an email QR code.
func (p *Printer) QRCodeEmail(to, subject, body string) *Printer {
	content := "mailto:" + to + "?subject=" + subject + "&body=" + body
	return p.QRCodeAdvanced(content, 6, QRErrorM, QRModel2)
}

// QRCodePhone prints a phone number QR code.
func (p *Printer) QRCodePhone(phone string) *Printer {
	content := "tel:" + phone
	return p.QRCodeAdvanced(content, 6, QRErrorM, QRModel2)
}

// Raw appends raw bytes to the buffer.
func (p *Printer) Raw(data []byte) *Printer {
	p.buffer = append(p.buffer, data...)
	return p
}

// Clear clears the buffer without sending.
func (p *Printer) Clear() *Printer {
	p.buffer = p.buffer[:0]
	return p
}

// Buffer returns the current buffer contents.
func (p *Printer) Buffer() []byte {
	return p.buffer
}

// Flush sends all buffered commands to the printer and clears the buffer.
func (p *Printer) Flush() error {
	if len(p.buffer) == 0 {
		return nil
	}

	if !p.adapter.IsOpen() {
		if err := p.adapter.Open(); err != nil {
			return fmt.Errorf("failed to open adapter: %w", err)
		}
	}

	err := p.adapter.Write(p.buffer)
	p.buffer = p.buffer[:0]
	return err
}

// Close closes the adapter.
func (p *Printer) Close() error {
	return p.adapter.Close()
}

// ============== HIGH-PRIORITY ESC/POS METHODS ==============

// Charset sets the international character set (0-15).
// 0=USA, 1=France, 2=Germany, 3=UK, 4=Denmark, 5=Sweden,
// 6=Italy, 7=Spain, 8=Japan, 9=Norway, 12=Latin, 13=Korea, 15=Chinese
func (p *Printer) Charset(n int) *Printer {
	p.buffer = append(p.buffer, SetCharset(n)...)
	return p
}

// CodePage sets the character code table.
// 0=PC437, 1=Katakana, 2=PC850, 3=PC860, 4=PC863, 5=PC865,
// 6=WestEur, 7=Greek, 8=Hebrew, 16=WPC1252, 17=PC866, 18=PC852, 19=PC858
func (p *Printer) CodePage(n int) *Printer {
	p.buffer = append(p.buffer, SetCodePage(n)...)
	return p
}

// Reverse sets white/black reverse printing mode.
func (p *Printer) Reverse(on bool) *Printer {
	if on {
		p.buffer = append(p.buffer, REVERSE_ON...)
	} else {
		p.buffer = append(p.buffer, REVERSE_OFF...)
	}
	return p
}

// LineSpacing sets line spacing to n/180 inch (or n/60 mm).
// Use 0 for tight spacing, 60 for 1/3 inch, etc.
func (p *Printer) LineSpacing(n int) *Printer {
	p.buffer = append(p.buffer, SetLineSpacing(n)...)
	return p
}

// LineSpacingDefault resets line spacing to default (1/6 inch).
func (p *Printer) LineSpacingDefault() *Printer {
	p.buffer = append(p.buffer, LINE_SPACING_DEFAULT...)
	return p
}

// FeedDots feeds paper by n dots (1 dot = 1/180 inch approx).
func (p *Printer) FeedDots(n int) *Printer {
	p.buffer = append(p.buffer, FeedDots(n)...)
	return p
}

// FeedLines prints and feeds n lines.
func (p *Printer) FeedLines(n int) *Printer {
	p.buffer = append(p.buffer, FeedLines(n)...)
	return p
}

// RasterImage prints a raster bit image.
// mode: 0=normal, 1=double-width, 2=double-height, 3=quadruple
// data: raw bitmap data (1 bit per dot, 8 dots per byte, MSB first)
// widthBytes: width in bytes (widthBytes*8 = width in dots)
// heightDots: height in dots
func (p *Printer) RasterImage(mode int, widthBytes, heightDots int, data []byte) *Printer {
	p.buffer = append(p.buffer, RasterImageCmd(mode, widthBytes, heightDots)...)
	p.buffer = append(p.buffer, data...)
	return p
}
