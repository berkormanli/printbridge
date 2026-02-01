package printer

// ESC/POS Commands based on node-escpos reference

// Control characters
const (
	LF  = "\x0a"
	FS  = "\x1c"
	FF  = "\x0c"
	GS  = "\x1d"
	DLE = "\x10"
	EOT = "\x04"
	NUL = "\x00"
	ESC = "\x1b"
	EOL = "\n"
)

// Hardware commands
var (
	HW_INIT   = []byte{0x1b, 0x40}             // ESC @ - Initialize printer
	HW_SELECT = []byte{0x1b, 0x3d, 0x01}       // ESC = 1 - Select printer
	HW_RESET  = []byte{0x1b, 0x3f, 0x0a, 0x00} // Reset printer
)

// Feed control
var (
	CTL_LF = []byte{0x0a}       // Line feed
	CTL_FF = []byte{0x0c}       // Form feed
	CTL_CR = []byte{0x0d}       // Carriage return
	CTL_HT = []byte{0x09}       // Horizontal tab
	CTL_VT = []byte{0x0b}       // Vertical tab
)

// Text formatting
var (
	TXT_NORMAL   = []byte{0x1b, 0x21, 0x00} // Normal text
	TXT_2HEIGHT  = []byte{0x1b, 0x21, 0x10} // Double height
	TXT_2WIDTH   = []byte{0x1b, 0x21, 0x20} // Double width
	TXT_4SQUARE  = []byte{0x1b, 0x21, 0x30} // Double width & height

	TXT_UNDERL_OFF = []byte{0x1b, 0x2d, 0x00} // Underline off
	TXT_UNDERL_ON  = []byte{0x1b, 0x2d, 0x01} // Underline 1-dot on
	TXT_UNDERL2_ON = []byte{0x1b, 0x2d, 0x02} // Underline 2-dot on

	TXT_BOLD_OFF = []byte{0x1b, 0x45, 0x00} // Bold off
	TXT_BOLD_ON  = []byte{0x1b, 0x45, 0x01} // Bold on

	TXT_ITALIC_OFF = []byte{0x1b, 0x35} // Italic off
	TXT_ITALIC_ON  = []byte{0x1b, 0x34} // Italic on

	TXT_FONT_A = []byte{0x1b, 0x4d, 0x00} // Font A
	TXT_FONT_B = []byte{0x1b, 0x4d, 0x01} // Font B
	TXT_FONT_C = []byte{0x1b, 0x4d, 0x02} // Font C

	TXT_ALIGN_LT = []byte{0x1b, 0x61, 0x00} // Left align
	TXT_ALIGN_CT = []byte{0x1b, 0x61, 0x01} // Center align
	TXT_ALIGN_RT = []byte{0x1b, 0x61, 0x02} // Right align
)

// Paper cutting
var (
	PAPER_FULL_CUT = []byte{0x1d, 0x56, 0x00} // Full cut
	PAPER_PART_CUT = []byte{0x1d, 0x56, 0x01} // Partial cut
)

// Cash drawer
var (
	CD_KICK_2 = []byte{0x1b, 0x70, 0x00, 0x19, 0x78} // Kick pin 2
	CD_KICK_5 = []byte{0x1b, 0x70, 0x01, 0x19, 0x78} // Kick pin 5
)

// Barcode format
var (
	BARCODE_TXT_OFF = []byte{0x1d, 0x48, 0x00} // HRI off
	BARCODE_TXT_ABV = []byte{0x1d, 0x48, 0x01} // HRI above
	BARCODE_TXT_BLW = []byte{0x1d, 0x48, 0x02} // HRI below
	BARCODE_TXT_BTH = []byte{0x1d, 0x48, 0x03} // HRI both

	BARCODE_FONT_A = []byte{0x1d, 0x66, 0x00} // Font A for HRI
	BARCODE_FONT_B = []byte{0x1d, 0x66, 0x01} // Font B for HRI

	BARCODE_HEIGHT_DEFAULT = []byte{0x1d, 0x68, 0x64} // Height 100
	BARCODE_WIDTH_DEFAULT  = []byte{0x1d, 0x77, 0x01} // Width 1

	BARCODE_UPC_A   = []byte{0x1d, 0x6b, 0x00} // UPC-A
	BARCODE_UPC_E   = []byte{0x1d, 0x6b, 0x01} // UPC-E
	BARCODE_EAN13   = []byte{0x1d, 0x6b, 0x02} // EAN13
	BARCODE_EAN8    = []byte{0x1d, 0x6b, 0x03} // EAN8
	BARCODE_CODE39  = []byte{0x1d, 0x6b, 0x04} // CODE39
	BARCODE_ITF     = []byte{0x1d, 0x6b, 0x05} // ITF
	BARCODE_NW7     = []byte{0x1d, 0x6b, 0x06} // NW7
	BARCODE_CODE93  = []byte{0x1d, 0x6b, 0x48} // CODE93
	BARCODE_CODE128 = []byte{0x1d, 0x6b, 0x49} // CODE128
)

// QR Code
var (
	QR_MODEL      = []byte{0x1d, 0x28, 0x6b, 0x04, 0x00, 0x31, 0x41} // Set QR model
	QR_SIZE       = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x43} // Set QR size
	QR_ERROR      = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x45} // Set error correction
	QR_STORE_PRE  = []byte{0x1d, 0x28, 0x6b}                         // Store data prefix
	QR_STORE_POST = []byte{0x31, 0x50, 0x30}                         // Store data postfix
	QR_PRINT      = []byte{0x1d, 0x28, 0x6b, 0x03, 0x00, 0x31, 0x51, 0x30} // Print QR
)

// Beep
var BEEP = []byte{0x1b, 0x42}

// TxtCustomSize returns the command for custom text size.
func TxtCustomSize(width, height int) []byte {
	if width < 1 {
		width = 1
	}
	if width > 8 {
		width = 8
	}
	if height < 1 {
		height = 1
	}
	if height > 8 {
		height = 8
	}

	size := byte((width-1)*16 + (height - 1))
	return []byte{0x1d, 0x21, size}
}

// BarcodeHeight returns the command for barcode height.
func BarcodeHeight(height int) []byte {
	if height < 1 {
		height = 1
	}
	if height > 255 {
		height = 255
	}
	return []byte{0x1d, 0x68, byte(height)}
}

// BarcodeWidth returns the command for barcode width.
func BarcodeWidth(width int) []byte {
	if width < 2 {
		width = 2
	}
	if width > 6 {
		width = 6
	}
	return []byte{0x1d, 0x77, byte(width)}
}

// ============== HIGH-PRIORITY ESC/POS COMMANDS ==============

// International character sets (ESC R n)
var (
	CHARSET_USA       = []byte{0x1b, 0x52, 0x00} // USA
	CHARSET_FRANCE    = []byte{0x1b, 0x52, 0x01} // France
	CHARSET_GERMANY   = []byte{0x1b, 0x52, 0x02} // Germany
	CHARSET_UK        = []byte{0x1b, 0x52, 0x03} // UK
	CHARSET_DENMARK   = []byte{0x1b, 0x52, 0x04} // Denmark
	CHARSET_SWEDEN    = []byte{0x1b, 0x52, 0x05} // Sweden
	CHARSET_ITALY     = []byte{0x1b, 0x52, 0x06} // Italy
	CHARSET_SPAIN     = []byte{0x1b, 0x52, 0x07} // Spain
	CHARSET_JAPAN     = []byte{0x1b, 0x52, 0x08} // Japan
	CHARSET_NORWAY    = []byte{0x1b, 0x52, 0x09} // Norway
	CHARSET_LATIN     = []byte{0x1b, 0x52, 0x0c} // Latin
	CHARSET_KOREA     = []byte{0x1b, 0x52, 0x0d} // Korea
	CHARSET_CHINESE   = []byte{0x1b, 0x52, 0x0f} // Chinese
)

// SetCharset returns the command for setting international character set.
func SetCharset(n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > 15 {
		n = 15
	}
	return []byte{0x1b, 0x52, byte(n)}
}

// Code page selection (ESC t n)
var (
	CODEPAGE_PC437     = []byte{0x1b, 0x74, 0x00} // PC437 USA/Europe
	CODEPAGE_KATAKANA  = []byte{0x1b, 0x74, 0x01} // Katakana
	CODEPAGE_PC850     = []byte{0x1b, 0x74, 0x02} // PC850 Multilingual
	CODEPAGE_PC860     = []byte{0x1b, 0x74, 0x03} // PC860 Portuguese
	CODEPAGE_PC863     = []byte{0x1b, 0x74, 0x04} // PC863 Canadian French
	CODEPAGE_PC865     = []byte{0x1b, 0x74, 0x05} // PC865 Nordic
	CODEPAGE_WESTEUR   = []byte{0x1b, 0x74, 0x06} // West Europe
	CODEPAGE_GREEK     = []byte{0x1b, 0x74, 0x07} // Greek
	CODEPAGE_HEBREW    = []byte{0x1b, 0x74, 0x08} // Hebrew
	CODEPAGE_WPC1252   = []byte{0x1b, 0x74, 0x10} // Windows-1252
	CODEPAGE_PC866     = []byte{0x1b, 0x74, 0x11} // PC866 Cyrillic
	CODEPAGE_PC852     = []byte{0x1b, 0x74, 0x12} // PC852 Latin2
	CODEPAGE_PC858     = []byte{0x1b, 0x74, 0x13} // PC858 Euro
)

// SetCodePage returns the command for setting code page.
func SetCodePage(n int) []byte {
	if n < 0 {
		n = 0
	}
	return []byte{0x1b, 0x74, byte(n)}
}

// White/Black reverse mode (GS B n)
var (
	REVERSE_OFF = []byte{0x1d, 0x42, 0x00} // Reverse mode off
	REVERSE_ON  = []byte{0x1d, 0x42, 0x01} // Reverse mode on (white on black)
)

// Line spacing (ESC 3 n) - sets line spacing to n/180 inch
func SetLineSpacing(n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	return []byte{0x1b, 0x33, byte(n)}
}

// Default line spacing (ESC 2) - sets line spacing to 1/6 inch
var LINE_SPACING_DEFAULT = []byte{0x1b, 0x32}

// Feed n dots (ESC J n)
func FeedDots(n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	return []byte{0x1b, 0x4a, byte(n)}
}

// Feed n lines (ESC d n)
func FeedLines(n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	return []byte{0x1b, 0x64, byte(n)}
}

// Raster bit image modes (GS v 0 m)
const (
	RASTER_NORMAL        = 0  // Normal mode (200x200 DPI)
	RASTER_DOUBLE_WIDTH  = 1  // Double width (100x200 DPI)
	RASTER_DOUBLE_HEIGHT = 2  // Double height (200x100 DPI)
	RASTER_QUADRUPLE     = 3  // Quadruple (100x100 DPI)
)

// RasterImageCmd returns the command prefix for raster bit image.
// Format: GS v 0 m xL xH yL yH d1...dk
// xL xH: horizontal dots (xL + xH*256) bytes = (xL + xH*256)*8 dots
// yL yH: vertical dots (yL + yH*256) dots
func RasterImageCmd(mode int, widthBytes, heightDots int) []byte {
	if mode < 0 || mode > 3 {
		mode = 0
	}
	xL := byte(widthBytes % 256)
	xH := byte(widthBytes / 256)
	yL := byte(heightDots % 256)
	yH := byte(heightDots / 256)
	return []byte{0x1d, 0x76, 0x30, byte(mode), xL, xH, yL, yH}
}
