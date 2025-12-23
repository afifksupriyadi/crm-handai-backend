package model

import (
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
)

// GetLoyalCustomersRequest represents request to get loyal customers list
type GetLoyalCustomersRequest struct {
	request.AuthorizedRequest
	Page  int `query:"page" default:"1" minimum:"1" doc:"Page number for pagination"`
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"100" doc:"Number of items per page"`
}

// LoyalCustomerItem represents a single loyal customer
type LoyalCustomerItem struct {
	ID                        int       `json:"id"`
	Name                      string    `json:"name"`
	Phone                     string    `json:"phone"`
	LastPurchase              time.Time `json:"last_purchase"`
	TotalProductPurchasesWeek int       `json:"total_product_purchases_week"`
	TotalMonthPurchase        int       `json:"total_month_purchase"`
}

// LoyalCustomersResponse represents paginated list of loyal customers
type LoyalCustomersResponse struct {
	Data       []LoyalCustomerItem `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}
