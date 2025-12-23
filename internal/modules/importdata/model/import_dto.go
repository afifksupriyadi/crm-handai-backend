package model

import (
	"fmt"
	"mime/multipart"
	"time"
)

// ImportBatchRequest represents batch import request
// Used for validation in handler, NOT passed to service
type ImportBatchRequest struct {
	FileCustomer    *multipart.FileHeader `form:"file_customer"`    // Optional
	FileTransaction *multipart.FileHeader `form:"file_transaction"` // Required
	StartDate       string                `form:"start_date"`       // Required, format: YYYY-MM-DD
	EndDate         string                `form:"end_date"`         // Required, format: YYYY-MM-DD
	Notes           string                `form:"notes"`
}

// Validate validates the import batch request
func (r *ImportBatchRequest) Validate() error {
	if r.FileTransaction == nil {
		return fmt.Errorf("file_transaction is required")
	}

	if r.StartDate == "" {
		return fmt.Errorf("start_date is required")
	}

	if r.EndDate == "" {
		return fmt.Errorf("end_date is required")
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start_date format, expected YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end_date format, expected YYYY-MM-DD")
	}

	// Validate date range
	if endDate.Before(startDate) {
		return fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	return nil
}

// GetParsedDates returns parsed start and end dates
func (r *ImportBatchRequest) GetParsedDates() (time.Time, time.Time, error) {
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startDate, endDate, nil
}

// ImportBatchResponse represents the result of batch import
type ImportBatchResponse struct {
	Batch        *BatchInfo                `json:"batch"`
	Customers    *ImportCustomerSummary    `json:"customers"`
	Transactions *ImportTransactionSummary `json:"transactions"`
}

// BatchInfo contains batch metadata
type BatchInfo struct {
	ID        int    `json:"id"`
	BatchCode string `json:"batch_code"`
	BatchDate string `json:"batch_date"`
	Status    string `json:"status"`
	IsActive  bool   `json:"is_active"`
}

// ImportCustomerSummary shows customer import results
type ImportCustomerSummary struct {
	TotalRows        int              `json:"totalRows"`
	SuccessRows      int              `json:"successRows"`
	FailedRows       int              `json:"failedRows"`
	CustomersCreated int              `json:"customersCreated"`
	CustomersUpdated int              `json:"customersUpdated"`
	Errors           []ImportRowError `json:"errors,omitempty"`
}

// ImportTransactionSummary shows transaction import results
type ImportTransactionSummary struct {
	TotalRows                 int              `json:"totalRows"`
	SuccessRows               int              `json:"successRows"`
	FailedRows                int              `json:"failedRows"`
	TransactionsCreated       int              `json:"transactionsCreated"`
	TransactionDetailsCreated int              `json:"transactionDetailsCreated"`
	ProductsCreated           int              `json:"productsCreated"`
	VariantsCreated           int              `json:"variantsCreated"`
	Errors                    []ImportRowError `json:"errors,omitempty"`
}

// ImportRowError represents error for specific row during import
type ImportRowError struct {
	RowNumber int    `json:"rowNumber"`
	Field     string `json:"field,omitempty"`
	Message   string `json:"message"`
}
