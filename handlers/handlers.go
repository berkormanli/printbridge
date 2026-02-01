package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"printbridge/pkg/adapter"
	"printbridge/pkg/printer"
)

// PrintService holds the printer and adapter for HTTP handlers.
type PrintService struct {
	Adapter      adapter.Adapter
	Printer      *printer.Printer
	TemplatesDir string
}

// NewPrintService creates a new print service.
func NewPrintService(a adapter.Adapter) *PrintService {
	return &PrintService{
		Adapter:      a,
		Printer:      printer.New(a),
		TemplatesDir: "templates", // Default templates directory
	}
}

// NewPrintServiceWithTemplates creates a print service with custom templates path.
func NewPrintServiceWithTemplates(a adapter.Adapter, templatesDir string) *PrintService {
	return &PrintService{
		Adapter:      a,
		Printer:      printer.New(a),
		TemplatesDir: templatesDir,
	}
}

// HealthHandler responds with service health status.
func (s *PrintService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// StatusHandler responds with printer connection status.
func (s *PrintService) StatusHandler(w http.ResponseWriter, r *http.Request) {
	connected := s.Adapter.IsOpen()

	// Try to connect if not already connected
	if !connected {
		if err := s.Adapter.Open(); err == nil {
			connected = true
		}
	}

	status := map[string]interface{}{
		"connected": connected,
		"service":   "running",
	}

	// Add USB printer info if available
	if printers, err := adapter.FindPrinters(); err == nil && len(printers) > 0 {
		status["printers"] = printers
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ReceiptItem represents an item in a receipt.
type ReceiptItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"qty"`
	Price    float64 `json:"price"`
}

// PrintRequest represents a print job request.
type PrintRequest struct {
	Header string        `json:"header"`
	Items  []ReceiptItem `json:"items"`
	Total  float64       `json:"total"`
	Footer string        `json:"footer"`
}

// PrintHandler handles receipt printing.
func (s *PrintService) PrintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PrintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	p := s.Printer

	// Build receipt
	p.Init().
		Align("center").
		Bold(true).
		Println(req.Header).
		Bold(false).
		NewLine().
		Align("left").
		DrawLine("-")

	// Print items
	for _, item := range req.Items {
		line := fmt.Sprintf("%-20s x%d  $%.2f", truncate(item.Name, 20), item.Quantity, item.Price)
		p.Println(line)
	}

	// Print total
	p.DrawLine("-").
		Align("right").
		Bold(true).
		Println(fmt.Sprintf("TOTAL: $%.2f", req.Total)).
		Bold(false).
		NewLine()

	// Print footer
	if req.Footer != "" {
		p.Align("center").
			Println(req.Footer)
	}

	p.Feed(2).Cut(false)

	// Send to printer
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Receipt printed",
	})
}

// RawPrintRequest represents a raw print request.
type RawPrintRequest struct {
	Data []byte `json:"data"`
}

// RawPrintHandler handles raw ESC/POS printing.
func (s *PrintService) RawPrintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RawPrintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	s.Printer.Raw(req.Data)
	if err := s.Printer.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Raw data sent",
	})
}

// TemplatePrintHandler handles template-based receipt printing for food delivery platforms.
func (s *PrintService) TemplatePrintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request: %v", err), http.StatusBadRequest)
		return
	}

	// Parse the order
	order, err := printer.ParseTemplateOrder(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid order JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Print the order using template
	if err := s.Printer.PrintTemplateOrder(*order, s.TemplatesDir); err != nil {
		http.Error(w, fmt.Sprintf("Print failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"message":  "Order printed",
		"platform": order.Platform,
	})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// TestPrintHandler prints a comprehensive test receipt to verify all features.
func (s *PrintService) TestPrintHandler(w http.ResponseWriter, r *http.Request) {
	p := s.Printer

	// Initialize and build comprehensive test receipt
	p.Init()

	// ===== HEADER SECTION =====
	p.Align("center").
		Size(2, 2).
		Bold(true).
		Println("PRINTBRIDGE").
		Size(1, 1).
		Bold(false).
		Println("COMPREHENSIVE TEST").
		NewLine().
		Println("================================").
		NewLine()
	
	// Flush header immediately
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print header failed: %v", err), http.StatusInternalServerError)
		return
	}

	// ===== STORE INFO =====
	p.Align("center").
		Println("Sample Store Name").
		Println("123 Main Street").
		Println("City, State 12345").
		Println("Tel: (555) 123-4567").
		NewLine()

	// ===== DATE/TIME/RECEIPT# =====
	p.Align("left").
		DrawLine("-")
	
	now := fmt.Sprintf("Date: %s", getCurrentTime())
	p.Println(now).
		Println("Receipt #: TEST-001").
		Println("Cashier: Demo User").
		DrawLine("-")

	// ===== ITEMS SECTION =====
	p.Bold(true).
		Println("ITEMS").
		Bold(false)

	// Sample items with various lengths
	items := []struct {
		name  string
		qty   int
		price float64
	}{
		{"Espresso", 2, 3.50},
		{"Cappuccino Large", 1, 4.75},
		{"Croissant", 3, 2.50},
		{"Blueberry Muffin", 2, 3.25},
		{"Bottled Water 500ml", 1, 1.50},
	}

	subtotal := 0.0
	for _, item := range items {
		total := float64(item.qty) * item.price
		subtotal += total
		line := fmt.Sprintf("%-20s", truncate(item.name, 20))
		p.Println(line)
		qtyLine := fmt.Sprintf("  %d x $%.2f = $%.2f", item.qty, item.price, total)
		p.Println(qtyLine)
	}

	// ===== TOTALS SECTION =====
	p.DrawLine("-")

	tax := subtotal * 0.08 // 8% tax
	total := subtotal + tax

	p.Align("right").
		Println(fmt.Sprintf("Subtotal: $%.2f", subtotal)).
		Println(fmt.Sprintf("Tax (8%%): $%.2f", tax)).
		NewLine().
		Bold(true).
		Size(1, 2).
		Println(fmt.Sprintf("TOTAL: $%.2f", total)).
		Size(1, 1).
		Bold(false)

	// ===== PAYMENT SECTION =====
	p.Align("left").
		DrawLine("-").
		Println("Payment Method: CASH").
		Println(fmt.Sprintf("Amount Tendered: $%.2f", 50.00)).
		Println(fmt.Sprintf("Change: $%.2f", 50.00-total)).
		DrawLine("-")

	// Flush receipt body
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print body failed: %v", err), http.StatusInternalServerError)
		return
	}

	// ===== NEW FEATURES TEST SECTION =====
	p.Align("center").
		NewLine().
		Size(1, 2).
		Bold(true).
		Println("=== NEW FEATURES TEST ===").
		Size(1, 1).
		Bold(false).
		NewLine()

	// --- Reverse Mode Test ---
	p.Align("left").
		Println("1. REVERSE MODE (White on Black):").
		Reverse(true).
		Println("  THIS TEXT IS REVERSED  ").
		Reverse(false).
		NewLine()

	// --- Line Spacing Test ---
	p.Println("2. LINE SPACING TEST:").
		Println("   Default spacing above").
		LineSpacing(10). // Tight spacing
		Println("   Tight spacing (10)").
		Println("   Still tight spacing").
		LineSpacing(60). // Wide spacing
		Println("   Wide spacing (60)").
		Println("   Still wide spacing").
		LineSpacingDefault(). // Reset
		Println("   Back to default").
		NewLine()

	// --- Feed Control Test ---
	p.Println("3. FEED CONTROL:").
		Println("   FeedDots(30) below...").
		FeedDots(30).
		Println("   FeedLines(2) below...").
		FeedLines(2).
		Println("   Back to normal").
		NewLine()

	// --- Underline Test ---
	p.Println("4. UNDERLINE MODES:").
		Underline(1).
		Println("   Single underline (1-dot)").
		Underline(2).
		Println("   Double underline (2-dot)").
		Underline(0).
		Println("   No underline").
		NewLine()

	// --- Font Test ---
	p.Println("5. FONT SELECTION:").
		Font("A").
		Println("   Font A (12x24 default)").
		Font("B").
		Println("   Font B (9x17 smaller)").
		Font("A").
		NewLine()

	// Flush features section 1
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print features 1 failed: %v", err), http.StatusInternalServerError)
		return
	}

	// --- Size Combinations ---
	p.Println("6. SIZE COMBINATIONS:").
		Size(2, 1).
		Println("  2x Width").
		Size(1, 2).
		Println(" 2x Height").
		Size(2, 2).
		Println("2x Both").
		Size(3, 3).
		Println("3x").
		Size(1, 1).
		Normal().
		NewLine()

	// --- Raster Image Demo (simple pattern) ---
	p.Println("7. RASTER IMAGE (8x8 checkerboard):").
		Align("center")
	
	// Create a simple 8x8 checkerboard pattern (1 byte wide, 8 rows tall)
	checkerboard := []byte{
		0xAA, // 10101010
		0x55, // 01010101
		0xAA, // 10101010
		0x55, // 01010101
		0xAA, // 10101010
		0x55, // 01010101
		0xAA, // 10101010
		0x55, // 01010101
	}
	p.RasterImage(0, 1, 8, checkerboard).
		NewLine().
		Align("left").
		Println("   (8x8 pixel pattern)").
		NewLine()

	// ===== BARCODE TEST =====
	p.Align("center").
		Println("8. BARCODE TEST:").
		Barcode("1234567890", "CODE39", 2, 60).
		NewLine()

	// ===== QR CODE TYPES TEST =====
	p.Println("9. QR CODE TYPES:").
		Align("left").
		Println("   a) URL QR (Error Level M):").
		Align("center").
		QRCodeURL("https://www.google.com", 5).
		NewLine().
		Align("left").
		Println("   b) WiFi QR:").
		Align("center").
		QRCodeWiFi("MyNetwork", "password123", "WPA", false).
		NewLine().
		Align("left").
		Println("   c) Phone QR:").
		Align("center").
		QRCodePhone("+905315828192").
		NewLine().
		Align("left").
		Println("   d) High Error Correction (H):").
		Align("center").
		QRCodeAdvanced("Error Level H Test", 5, printer.QRErrorH, printer.QRModel2).
		NewLine()

	// Flush features section 2
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print features 2 failed: %v", err), http.StatusInternalServerError)
		return
	}

	// ===== FOOTER =====
	p.Align("center").
		DrawLine("=").
		NewLine().
		Println("Thank you for your visit!").
		Println("Please come again").
		NewLine().
		Println("*** END OF TEST ***").
		NewLine().
		Bold(true).
		Println("Features Tested:").
		Bold(false).
		Align("left").
		Println("- Text alignment (L/C/R)").
		Println("- Bold text").
		Println("- Double/triple height & width").
		Println("- Line separators").
		Println("- Underline modes (1-dot, 2-dot)").
		Println("- Font A & B selection").
		Println("- REVERSE MODE (new)").
		Println("- LINE SPACING control (new)").
		Println("- FEED DOTS/LINES (new)").
		Println("- RASTER IMAGE (new)").
		Println("- Barcode printing").
		Println("- QR code printing").
		Println("- Paper cut").
		NewLine()

	// Cut paper
	p.Feed(3).Cut(false)

	// Send final chunk
	if err := p.Flush(); err != nil {
		http.Error(w, fmt.Sprintf("Print footer failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Comprehensive test printed (flushed in 4 chunks)",
	})
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

