package model

import (
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
)

// GetChurnCustomersRequest represents request to get churn customers list
type GetChurnCustomersRequest struct {
	request.AuthorizedRequest
	Page  int `query:"page" default:"1" minimum:"1" doc:"Page number for pagination"`
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"100" doc:"Number of items per page"`
}

// ChurnCustomerItem represents a single churn customer
type ChurnCustomerItem struct {
	ID                    int       `json:"id"`
	Name                  string    `json:"name"`
	Phone                 string    `json:"phone"`
	LastPurchase          time.Time `json:"last_purchase"`
	DaysSinceLastPurchase int       `json:"days_since_last_purchase"`
}

// ChurnCustomersResponse represents paginated list of churn customers
type ChurnCustomersResponse struct {
	Data       []ChurnCustomerItem `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}
