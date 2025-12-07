package model

// ParsedProductInfo contains extracted and validated product information
type ParsedProductInfo struct {
	OriginalName   string // From Excel (e.g., "Kopi Susu Gula Aren (Medium) (Botol)")
	NormalizedName string // After normalization (e.g., "Kopi Susu Gula Aren (Botol)")
	Category       string // "COFFEE" or "NON-COFFEE"
	BasePrice      int64  // Base price for this product
	HasVariants    bool   // Whether this product has size variants
}

// ParsedVariantInfo contains extracted and validated variant information
type ParsedVariantInfo struct {
	Size          string // "250ml", "500ml", "1000ml"
	PriceModifier int64  // Additional price from base price
	IsDefault     bool   // Is this the default variant (250ml)
}

// ParsedCustomerInfo contains normalized customer information
type ParsedCustomerInfo struct {
	Name  string // Normalized name (title case, trimmed)
	Phone string // Normalized phone (cleaned, no spaces, no +62)
}

// ParsedTransactionInfo contains parsed transaction metadata
type ParsedTransactionInfo struct {
	Code            string // No. Struk
	TransactionDate string // Combined from Tanggal + Jam
	Discount        int64  // Diskon Transaksi
	ShippingCost    int64  // Ongkos Kirim
	PaymentMethod   string // Metode Pembayaran
	Status          string // Status
}
