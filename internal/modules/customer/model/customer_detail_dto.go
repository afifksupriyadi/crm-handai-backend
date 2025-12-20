package model

import "time"

// CustomerDetailRequest for filtering transaction history
type CustomerDetailRequest struct {
	Month string `query:"month" doc:"Filter month (format: YYYY-MM, e.g., 2025-11)"` // Optional
}

// CustomerDetailResponse contains comprehensive customer information
type CustomerDetailResponse struct {
	Customer           CustomerInfo           `json:"customer"`
	Stats              CustomerStats          `json:"stats"`
	LatestPurchase     *LatestPurchaseInfo    `json:"latest_purchase,omitempty"`
	Prediction         *PredictionInfo        `json:"prediction,omitempty"`
	ProductPurchases   []ProductPurchaseInfo  `json:"product_purchases"`
	TransactionHistory TransactionHistoryInfo `json:"transaction_history"`
}

// CustomerInfo basic customer information
type CustomerInfo struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Phone             string     `json:"phone"`
	Segment           *string    `json:"segment,omitempty"`
	IsLoyal           bool       `json:"is_loyal"`
	CreatedAt         time.Time  `json:"created_at"`
	UpgradedFromGuest bool       `json:"upgraded_from_guest"`
	UpgradedAt        *time.Time `json:"upgraded_at,omitempty"`
}

// CustomerStats aggregated statistics
type CustomerStats struct {
	TotalProductsPurchases int     `json:"total_products_purchases"` // This month
	TotalTransaction       string  `json:"total_transaction"`        // Formatted: "Rp. xxx.xxx,00"
	TotalSpent             float64 `json:"total_spent"`              // Raw number
	TotalTransactionCount  int     `json:"total_transaction_count"`  // All time
}

// LatestPurchaseInfo information about the most recent purchase
type LatestPurchaseInfo struct {
	TransactionCode string                  `json:"transaction_code"`
	Date            time.Time               `json:"date"`
	DaysAgo         int                     `json:"days_ago"`
	Products        []TransactionProductDTO `json:"products"`
}

// TransactionProductDTO product in a transaction
type TransactionProductDTO struct {
	Name     string `json:"name"`
	Variant  string `json:"variant,omitempty"`
	Quantity int    `json:"quantity"`
}

// PredictionInfo next purchase prediction from analytics
type PredictionInfo struct {
	NextPurchaseBy      *time.Time `json:"next_purchase_by,omitempty"`
	SupposeNextPurchase *time.Time `json:"suppose_next_purchase,omitempty"`
	DaysUntilNext       *int       `json:"days_until_next,omitempty"`
	ConfidenceScore     *float64   `json:"confidence_score,omitempty"`
	PredictedProducts   []string   `json:"predicted_products,omitempty"`
	AvgDaysBetweenBuy   *float64   `json:"avg_days_between_purchase,omitempty"`
}

// ProductPurchaseInfo aggregated product purchase information
type ProductPurchaseInfo struct {
	ProductName   string `json:"product_name"`
	TotalQuantity int    `json:"total_quantity"`
}

// TransactionHistoryInfo transaction history with filtering
type TransactionHistoryInfo struct {
	FilterMonth  string                      `json:"filter_month"` // e.g., "November 2025"
	Transactions []TransactionHistoryItemDTO `json:"transactions"`
}

// TransactionHistoryItemDTO single transaction in history
type TransactionHistoryItemDTO struct {
	Date            time.Time               `json:"date"`
	DateFormatted   string                  `json:"date_formatted"` // e.g., "Thursday, 20 Nov 2025"
	TransactionCode string                  `json:"transaction_code"`
	Products        []TransactionProductDTO `json:"products"`
	TotalAmount     float64                 `json:"total_amount"`
	TotalFormatted  string                  `json:"total_formatted"`
}
