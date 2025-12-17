package model

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Customer struct {
	bun.BaseModel `bun:"table:customers,alias:c"`

	ID        int        `bun:"id,pk,autoincrement" json:"id"`
	Name      string     `bun:"name,notnull" json:"name"`
	Phone     string     `bun:"phone,notnull,unique" json:"phone"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt *time.Time `bun:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `bun:"deleted_at" json:"deleted_at,omitempty"`

	Segment                *string    `bun:"segment" json:"segment,omitempty"`
	IsLoyal                bool       `bun:"is_loyal,default:false" json:"is_loyal"`
	TotalTransactions      int        `bun:"total_transactions,default:0" json:"total_transactions"`
	TotalSpent             float64    `bun:"total_spent,type:decimal(12,2),default:0" json:"total_spent"`
	LastTransactionDate    *time.Time `bun:"last_transaction_date" json:"last_transaction_date,omitempty"`
	AvgDaysBetweenPurchase *float64   `bun:"avg_days_between_purchase,type:decimal(5,2)" json:"avg_days_between_purchase,omitempty"`
	ChurnRiskScore         *float64   `bun:"churn_risk_score,type:decimal(3,2)" json:"churn_risk_score,omitempty"`
}

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (r *CreateCustomerRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Phone = strings.TrimSpace(r.Phone)
}

func (r *CreateCustomerRequest) Validate(ctx context.Context) error {
	if err := validateName(ctx, r.Name); err != nil {
		return err
	}

	if err := validatePhone(ctx, r.Phone); err != nil {
		return err
	}

	return nil
}

type UpdateCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

func (r *UpdateCustomerRequest) Sanitize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Phone = strings.TrimSpace(r.Phone)
}

func (r *UpdateCustomerRequest) Validate(ctx context.Context) error {
	if err := validateName(ctx, r.Name); err != nil {
		return err
	}

	if err := validatePhone(ctx, r.Phone); err != nil {
		return err
	}

	return nil
}

// ✅ UPDATED: Tambah SortOrder field
type GetCustomersRequest struct {
	Page      int    `query:"page" default:"1" minimum:"1"`
	Limit     int    `query:"limit" default:"10" minimum:"1" maximum:"100"`
	Search    string `query:"search" default:""`
	SortOrder string `query:"sort_order" default:"asc" enum:"asc,desc"`
}

type CustomerResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Phone     string     `json:"phone"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CustomerListResponse struct {
	Data       []*CustomerResponse `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func (c *Customer) ToResponse() *CustomerResponse {
	return &CustomerResponse{
		ID:        c.ID,
		Name:      c.Name,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// Request untuk recent transactions
type GetRecentTransactionsRequest struct {
	Page  int `query:"page" default:"1" minimum:"1"`
	Limit int `query:"limit" default:"10" minimum:"1" maximum:"100"`
}

// Response dengan info last transaction
type CustomerRecentTransactionResponse struct {
	ID                       int        `json:"id"`
	Name                     string     `json:"name"`
	Phone                    string     `json:"phone"`
	LastTransactionDate      *time.Time `json:"last_transaction_date"`
	DaysSinceLastTransaction *int       `json:"days_since_last_transaction,omitempty"`
	TotalTransactions        int        `json:"total_transactions"`
	TotalSpent               float64    `json:"total_spent"`
}

type CustomerRecentTransactionListResponse struct {
	Data       []*CustomerRecentTransactionResponse `json:"data"`
	Pagination PaginationMeta                       `json:"pagination"`
}
