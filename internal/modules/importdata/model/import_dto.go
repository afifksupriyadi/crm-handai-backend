package model

import (
	"mime/multipart"
)

// ImportBatchRequest represents batch import request
// Used for validation in handler, NOT passed to service
type ImportBatchRequest struct {
	FileCustomer    *multipart.FileHeader `form:"file_customer"`    // OPTIONAL
	FileTransaction *multipart.FileHeader `form:"file_transaction"` // REQUIRED
	Notes           string                `form:"notes"`            // OPTIONAL
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
