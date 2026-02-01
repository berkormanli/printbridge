package printer

import (
	"encoding/json"
	"fmt"
	"image"
	"strconv"
	_ "golang.org/x/image/bmp"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateOrder represents an order from a food delivery platform
type TemplateOrder struct {
	Platform string           `json:"platform"`
	Merchant OrderMerchant    `json:"merchant"`
	Order    OrderInfo        `json:"order"`
	Customer OrderCustomer    `json:"customer"`
	Items    []OrderItem      `json:"items"`
	Totals   OrderTotals      `json:"totals"`
	Payment  OrderPayment     `json:"payment"`
	Notes    OrderNotes       `json:"notes"`
}

type OrderMerchant struct {
	Name         string `json:"name"`
	District     string `json:"district"`
	Neighborhood string `json:"neighborhood"`
}

type OrderInfo struct {
	OrderTime string `json:"order_time"`
	OrderType string `json:"order_type"`
}

type OrderCustomer struct {
	Name    string          `json:"name"`
	Address CustomerAddress `json:"address"`
	Phone   string          `json:"phone"`
}

type CustomerAddress struct {
	Neighborhood  string      `json:"neighborhood"`
	StreetAddress string      `json:"street_address"`
	Floor         interface{} `json:"floor"`     // Can be string or int
	Apartment     interface{} `json:"apartment"` // Can be string or int
	District      string      `json:"district"`
	City          string      `json:"city"`
	Description   string      `json:"description"`
}

// GetFloor returns the floor as an int
func (a CustomerAddress) GetFloor() int {
	return toInt(a.Floor)
}

// GetApartment returns the apartment as an int
func (a CustomerAddress) GetApartment() int {
	return toInt(a.Apartment)
}

// toInt converts interface{} to int (handles string, float64, int)
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

type OrderItem struct {
	Name         string  `json:"name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price_try"`
	TotalPrice   float64 `json:"total_price_try"`
}

type OrderTotals struct {
	Subtotal    float64  `json:"subtotal_try"`
	DeliveryFee float64  `json:"delivery_fee_try"`
	VAT         OrderVAT `json:"vat"`
	Total       float64  `json:"total_try"`
}

type OrderVAT struct {
	Included bool `json:"included"`
}

type OrderPayment struct {
	Method string `json:"method"`
	Note   string `json:"note"`
}

type OrderNotes struct {
	CustomerNote *string `json:"customer_note"`
}

// Template represents a receipt template for a food delivery platform
type Template struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LogoPath string `json:"logo"`
}

// PlatformTemplates maps platform names to their template configurations
var PlatformTemplates = map[string]Template{
	"getir_yemek": {
		ID:       "getir_yemek",
		Name:     "Getir Yemek",
		LogoPath: "logos/getir_yemek.bmp",
	},
	"yemeksepeti": {
		ID:       "yemeksepeti",
		Name:     "Yemeksepeti",
		LogoPath: "logos/yemeksepeti.bmp",
	},
	"trendyol_go": {
		ID:       "trendyol_go",
		Name:     "Trendyol Go",
		LogoPath: "logos/trendyol_go.bmp",
	},
	"migros_yemek": {
		ID:       "migros_yemek",
		Name:     "Migros Yemek",
		LogoPath: "logos/migros_yemek.bmp",
	},
}

// NormalizePlatform converts a platform name to its template key
func NormalizePlatform(platform string) string {
	// Convert to lowercase and replace spaces with underscores
	normalized := strings.ToLower(strings.TrimSpace(platform))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	
	// Map common variations
	variations := map[string]string{
		"getir":         "getir_yemek",
		"getiryemek":    "getir_yemek",
		"getir yemek":   "getir_yemek",
		"yemeksepeti":   "yemeksepeti",
		"yemek sepeti":  "yemeksepeti",
		"trendyol":      "trendyol_go",
		"trendyolgo":    "trendyol_go",
		"trendyol go":   "trendyol_go",
		"trendyol_go":   "trendyol_go",
		"migros":        "migros_yemek",
		"migrosyemek":   "migros_yemek",
		"migros yemek":  "migros_yemek",
		"migros_yemek":  "migros_yemek",
	}
	
	if key, ok := variations[normalized]; ok {
		return key
	}
	return normalized
}

// GetTemplate returns the template for a given platform
func GetTemplate(platform string) (Template, bool) {
	key := NormalizePlatform(platform)
	tmpl, ok := PlatformTemplates[key]
	return tmpl, ok
}

// LoadLogo loads a logo image from the templates directory
func LoadLogo(templatesDir, logoPath string) (image.Image, error) {
	fullPath := filepath.Join(templatesDir, logoPath)
	
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open logo: %w", err)
	}
	defer f.Close()
	
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo: %w", err)
	}
	
	return img, nil
}

// ImageToRaster converts an image to ESC/POS raster format (1-bit per pixel)
func ImageToRaster(img image.Image) ([]byte, int, int) {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	
	// Width in bytes (8 pixels per byte)
	widthBytes := (width + 7) / 8
	
	// Create raster data
	data := make([]byte, widthBytes*height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			// Convert to grayscale (0-65535 range)
			gray := (r*299 + g*587 + b*114) / 1000
			
			// Threshold: if dark enough, set the bit (inverted for thermal: black = 1)
			if gray < 32768 { // 50% threshold
				byteIndex := y*widthBytes + x/8
				bitIndex := 7 - (x % 8)
				data[byteIndex] |= 1 << bitIndex
			}
		}
	}
	
	return data, widthBytes, height
}

// PrintTemplateOrder prints an order using the appropriate template
func (p *Printer) PrintTemplateOrder(order TemplateOrder, templatesDir string) error {
	// Get template for the platform
	tmpl, found := GetTemplate(order.Platform)
	if !found {
		// Use text-only header if no template found
		return p.printOrderWithoutLogo(order, order.Platform)
	}
	
	// Initialize printer
	p.Init()
	
	// Try to load and print logo
	if tmpl.LogoPath != "" {
		if img, err := LoadLogo(templatesDir, tmpl.LogoPath); err == nil {
			rasterData, widthBytes, height := ImageToRaster(img)
			p.Align("center").
				RasterImage(0, widthBytes, height, rasterData).
				NewLine()
		}
	}
	
	// Print platform header
	p.Align("center").
		Bold(true).
		Size(1, 2).
		Println(tmpl.Name).
		Size(1, 1).
		Bold(false).
		Println("Sipariş Fişi").
		NewLine().
		DrawLine("=")
	
	// Print the rest of the order
	return p.printOrderBody(order)
}

// printOrderWithoutLogo prints an order using text-only header
func (p *Printer) printOrderWithoutLogo(order TemplateOrder, platformName string) error {
	p.Init().
		Align("center").
		Reverse(true).
		Size(1, 2).
		Println(fmt.Sprintf(" %s ", strings.ToUpper(platformName))).
		Reverse(false).
		Size(1, 1).
		Println("Sipariş Fişi").
		NewLine().
		DrawLine("=")
	
	return p.printOrderBody(order)
}

// printOrderBody prints the main content of the order
func (p *Printer) printOrderBody(order TemplateOrder) error {
	// Merchant info
	p.Align("center").
		Bold(true).
		Println(order.Merchant.Name).
		Bold(false).
		Println(fmt.Sprintf("%s, %s", order.Merchant.Neighborhood, order.Merchant.District)).
		NewLine()
	
	// Order time
	p.Align("left").
		DrawLine("-")
	
	orderTime := order.Order.OrderTime
	if t, err := time.Parse(time.RFC3339, orderTime); err == nil {
		orderTime = t.Format("02.01.2006 15:04")
	} else if t, err := time.Parse("2006-01-02T15:04:05", orderTime); err == nil {
		orderTime = t.Format("02.01.2006 15:04")
	}
	
	p.Println(fmt.Sprintf("Sipariş Zamanı: %s", orderTime)).
		Println(fmt.Sprintf("Sipariş Tipi: %s", order.Order.OrderType)).
		DrawLine("-")
	
	// Customer info
	p.Bold(true).
		Println("MÜŞTERİ BİLGİLERİ").
		Bold(false).
		Println(fmt.Sprintf("Ad: %s", order.Customer.Name)).
		Println(fmt.Sprintf("Tel: %s", order.Customer.Phone)).
		NewLine().
		Println("Adres:").
		Println(order.Customer.Address.StreetAddress)
	
	if order.Customer.Address.GetFloor() > 0 || order.Customer.Address.GetApartment() > 0 {
		p.Println(fmt.Sprintf("Kat: %d, Daire: %d", order.Customer.Address.GetFloor(), order.Customer.Address.GetApartment()))
	}
	
	p.Println(fmt.Sprintf("%s, %s", order.Customer.Address.Neighborhood, order.Customer.Address.District)).
		Println(order.Customer.Address.City)
	
	if order.Customer.Address.Description != "" {
		p.Println(fmt.Sprintf("Not: %s", order.Customer.Address.Description))
	}
	
	p.DrawLine("-")
	
	// Items
	p.Bold(true).
		Println("SİPARİŞ DETAYI").
		Bold(false)
	
	for _, item := range order.Items {
		name := item.Name
		if len(name) > 24 {
			name = name[:24]
		}
		p.Println(fmt.Sprintf("%-24s", name))
		p.Println(fmt.Sprintf("  %d x %.2f TL = %.2f TL", item.Quantity, item.UnitPrice, item.TotalPrice))
	}
	
	// Totals
	p.DrawLine("-").
		Align("right")
	
	p.Println(fmt.Sprintf("Ara Toplam: %.2f TL", order.Totals.Subtotal))
	
	if order.Totals.DeliveryFee > 0 {
		p.Println(fmt.Sprintf("Paket Servis: %.2f TL", order.Totals.DeliveryFee))
	}
	
	if order.Totals.VAT.Included {
		p.Println("(KDV Dahil)")
	}
	
	p.NewLine().
		Bold(true).
		Size(1, 2).
		Println(fmt.Sprintf("TOPLAM: %.2f TL", order.Totals.Total)).
		Size(1, 1).
		Bold(false)
	
	// Payment
	p.Align("left").
		DrawLine("-").
		Println(fmt.Sprintf("Ödeme: %s", order.Payment.Method))
	
	if order.Payment.Note != "" {
		p.Println(order.Payment.Note)
	}
	
	// Customer notes
	if order.Notes.CustomerNote != nil && *order.Notes.CustomerNote != "" {
		p.DrawLine("-").
			Bold(true).
			Println("MÜŞTERİ NOTU:").
			Bold(false).
			Println(*order.Notes.CustomerNote)
	}
	
	// Footer
	p.DrawLine("=").
		Align("center").
		NewLine().
		Println("Afiyet olsun!").
		NewLine().
		Feed(2).
		Cut(false)
	
	return p.Flush()
}

// ParseTemplateOrder parses JSON data into a TemplateOrder
func ParseTemplateOrder(data []byte) (*TemplateOrder, error) {
	var order TemplateOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to parse order: %w", err)
	}
	return &order, nil
}
