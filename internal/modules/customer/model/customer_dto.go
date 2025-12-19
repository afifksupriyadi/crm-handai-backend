package model

import (
	"context"
	"strings"
	"time"
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

// CustomerResponse represents customer data in API response
type CustomerResponse struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Phone             string     `json:"phone"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty"`
	UpgradedFromGuest bool       `json:"upgraded_from_guest"`
	UpgradedAt        *time.Time `json:"upgraded_at,omitempty"`
	FirstSeenAsGuest  *time.Time `json:"first_seen_as_guest,omitempty"`
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
