package model

import (
	"context"
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/request"
)

// CreateCustomerRequest represents request to create a new customer
type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// Sanitize cleans up input data
func (r *CreateCustomerRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Phone = strings.TrimSpace(r.Phone)
}

// Validate validates the create customer request
func (r *CreateCustomerRequest) Validate(ctx context.Context) error {
	if err := validateName(ctx, r.Name); err != nil {
		return err
	}
	if err := validatePhone(ctx, r.Phone); err != nil {
		return err
	}
	return nil
}

// UpdateCustomerRequest represents request to update an existing customer
type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// Sanitize cleans up input data
func (r *UpdateCustomerRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Phone = strings.TrimSpace(r.Phone)
}

// Validate validates the update customer request
func (r *UpdateCustomerRequest) Validate(ctx context.Context) error {
	if err := validateName(ctx, r.Name); err != nil {
		return err
	}
	if err := validatePhone(ctx, r.Phone); err != nil {
		return err
	}
	return nil
}

// GetCustomersRequest represents request to get list of customers with pagination
type GetCustomersRequest struct {
	Page      int    `query:"page" default:"1" minimum:"1"`
	Limit     int    `query:"limit" default:"10" minimum:"1" maximum:"100"`
	Search    string `query:"search" default:""`
	SortOrder string `query:"sort_order" default:"asc" enum:"asc,desc"`
}

// CustomerResponse now includes last_transaction_date and is_loyal
type CustomerResponse struct {
	ID                    int        `json:"id"`
	Name                  string     `json:"name"`
	Phone                 string     `json:"phone"`
	Status                *string    `json:"status,omitempty"`
	LastTransactionDate   *time.Time `json:"last_transaction_date,omitempty"`
	SupposeToByBy         *time.Time `json:"suppose_to_by_by,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
	UpgradedFromGuest     bool       `json:"upgraded_from_guest"`
	UpgradedAt            *time.Time `json:"upgraded_at,omitempty"`
	FirstSeenAsGuest      *time.Time `json:"first_seen_as_guest,omitempty"`
	IsLoyal               bool       `json:"is_loyal"`
	DaysSinceLastPurchase *int       `json:"days_since_last_purchase,omitempty"`
}

// CustomerListResponse represents paginated list of customers
type CustomerListResponse struct {
	Data       []*CustomerResponse `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// GetRecentTransactionsQueryRequest for GET request with query params
type GetRecentTransactionsQueryRequest struct {
	request.AuthorizedRequest
	Page  int `query:"page" default:"1" minimum:"1" doc:"Page number for pagination"`
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"100" doc:"Number of items per page"`
}

// GetRecentTransactionsRequest for getting customers with recent transactions
type GetRecentTransactionsRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// CustomerRecentTransactionResponse represents customer with transaction stats
type CustomerRecentTransactionResponse struct {
	ID                       int        `json:"id"`
	Name                     string     `json:"name"`
	Phone                    string     `json:"phone"`
	LastTransactionDate      *time.Time `json:"last_transaction_date,omitempty"`
	TotalTransactions        int        `json:"total_transactions"`
	TotalSpent               float64    `json:"total_spent"`
	DaysSinceLastTransaction *int       `json:"days_since_last_transaction,omitempty"`
	Segment                  *string    `json:"segment,omitempty"`
	IsLoyal                  bool       `json:"is_loyal"`
	AvgDaysBetweenPurchase   *float64   `json:"avg_days_between_purchase,omitempty"`
	ChurnRiskScore           *float64   `json:"churn_risk_score,omitempty"`
}

// CustomerRecentTransactionListResponse for paginated recent transactions
type CustomerRecentTransactionListResponse struct {
	Data       []*CustomerRecentTransactionResponse `json:"data"`
	Pagination PaginationMeta                       `json:"pagination"`
}

// CustomerWithMetrics combines customer info with their metrics
type CustomerWithMetrics struct {
	Customer
	CustomerMetric
}
