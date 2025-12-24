package model

import "time"

type CustomerDetailRequest struct {
	Year  int `query:"year" minimum:"0" doc:"Filter by year (e.g., 2025). If 0, all years"`
	Month int `query:"month" minimum:"0" maximum:"12" doc:"Filter by month (1-12). If 0, all months"`
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
	Status            *string    `json:"status,omitempty"`
	IsLoyal           bool       `json:"is_loyal"`
	CreatedAt         time.Time  `json:"created_at"`
	UpgradedFromGuest bool       `json:"upgraded_from_guest"`
	UpgradedAt        *time.Time `json:"upgraded_at,omitempty"`
}

// CustomerStats aggregated statistics (filtered by year/month if specified)
type CustomerStats struct {
	TotalProductsPurchases        int     `json:"total_products_purchases"`
	TotalProductsVariantPurchases int     `json:"total_products_variant_purchases"`
	TotalTransaction              string  `json:"total_transaction"`
	TotalSpent                    float64 `json:"total_spent"`
	TotalTransactionCount         int     `json:"total_transaction_count"`
}

// LatestPurchaseInfo information about the most recent purchase (in filtered period)
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

// PredictionInfo represents customer's next purchase prediction (always latest, unfiltered)
type PredictionInfo struct {
	LastTransactionDate       string                 `json:"last_transaction_date"`
	PredictedNextPurchaseDate string                 `json:"predicted_next_purchase_date"`
	ActualNextPurchaseDate    *string                `json:"actual_next_purchase_date,omitempty"`
	IsPredictedCorrect        *bool                  `json:"is_predicted_correct,omitempty"`
	DaysUntilPredicted        *int                   `json:"days_until_predicted,omitempty"`
	PredictedProducts         []PredictedProductInfo `json:"predicted_products,omitempty"`
	CreatedAt                 string                 `json:"created_at"`
}

// ProductPurchaseInfo aggregated product purchase information (filtered by year/month)
type ProductPurchaseInfo struct {
	ProductName   string  `json:"product_name"`
	VariantName   *string `json:"variant_name,omitempty"`
	TotalQuantity int     `json:"total_quantity"`
}

// TransactionHistoryInfo transaction history with filtering
type TransactionHistoryInfo struct {
	FilterMonth  string                      `json:"filter_month"`
	Transactions []TransactionHistoryItemDTO `json:"transactions"`
}

// TransactionHistoryItemDTO single transaction in history
type TransactionHistoryItemDTO struct {
	Date            time.Time               `json:"date"`
	DateFormatted   string                  `json:"date_formatted"`
	TransactionCode string                  `json:"transaction_code"`
	Products        []TransactionProductDTO `json:"products"`
	TotalAmount     float64                 `json:"total_amount"`
	TotalFormatted  string                  `json:"total_formatted"`
}
