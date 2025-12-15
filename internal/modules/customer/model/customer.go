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

type GetCustomersRequest struct {
	Page   int    `query:"page" default:"1" minimum:"1"`
	Limit  int    `query:"limit" default:"10" minimum:"1" maximum:"100"`
	Search string `query:"search" default:""`
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
