package model

import "github.com/afifksupriyadi/crm-handai-backend/internal/util/request"

// GetLoyalCustomersRequest represents request for loyal customers list
type GetLoyalCustomersRequest struct {
	request.AuthorizedRequest
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"50" doc:"Number of customers to return"`
}

// GetChurnCustomersRequest represents request for churn customers list
type GetChurnCustomersRequest struct {
	request.AuthorizedRequest
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"50" doc:"Number of customers to return"`
}

// SegmentCustomerInfo represents customer info in segment lists
type SegmentCustomerInfo struct {
	ID                     int     `json:"id"`
	Name                   string  `json:"name"`
	LastPurchase           *string `json:"last_purchase"`            // ISO 8601 timestamp in WIB: "2025-09-30T10:30:00+07:00"
	DaysSinceLastPurchase  int     `json:"days_since_last_purchase"` // 2
	TotalTransactions      int     `json:"total_transactions"`       // Total number of transactions
	TotalProductsThisWeek  int     `json:"total_products_this_week"`  // For loyal customers
	TotalMonthlyPurchase   int     `json:"total_monthly_purchase"`    // For loyal customers
}

// LoyalCustomersResponse represents list of loyal customers
type LoyalCustomersResponse struct {
	Data []*SegmentCustomerInfo `json:"data"`
}

// ChurnCustomersResponse represents list of churn (at-risk) customers
type ChurnCustomersResponse struct {
	Data []*SegmentCustomerInfo `json:"data"`
}
