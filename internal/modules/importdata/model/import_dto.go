package model

import "mime/multipart"

// ==========================================
// REQUEST DTOs
// ==========================================

// ImportCustomerRequest represents customer import request
type ImportCustomerRequest struct {
	File *multipart.FileHeader `form:"file" required:"true" doc:"Customer Excel file (.xlsx)"`
}

// ImportTransactionRequest represents transaction import request
type ImportTransactionRequest struct {
	File *multipart.FileHeader `form:"file" required:"true" doc:"Transaction Excel file (.xlsx)"`
}

// ImportBatchRequest represents batch import request with both files
type ImportBatchRequest struct {
	FileCustomer     *multipart.FileHeader `form:"file_customer"`      // OPTIONAL
	FileTransaction  *multipart.FileHeader `form:"file_transaction"`   // REQUIRED
	BatchDate        string                `form:"batch_date"`         // YYYY-MM-DD format
	OverwriteIfExist bool                  `form:"overwrite_if_exist"` // If true, delete existing batch for this date
	Notes            string                `form:"notes"`              // Optional
}

// ==========================================
// RESPONSE DTOs
// ==========================================

// ImportCustomerResponse represents the result of customer import
type ImportCustomerResponse struct {
	TotalRows        int              `json:"totalRows"`
	SuccessRows      int              `json:"successRows"`
	FailedRows       int              `json:"failedRows"`
	CustomersCreated int              `json:"customersCreated"`
	Errors           []ImportRowError `json:"errors,omitempty"`
}

// ImportTransactionResponse represents the result of transaction import
type ImportTransactionResponse struct {
	TotalRows   int                       `json:"totalRows"`
	SuccessRows int                       `json:"successRows"`
	FailedRows  int                       `json:"failedRows"`
	Summary     *TransactionImportSummary `json:"summary"`
	Errors      []ImportRowError          `json:"errors,omitempty"`
}

// TransactionImportSummary shows detailed stats for transaction import
type TransactionImportSummary struct {
	TransactionsCreated       int `json:"transactionsCreated"`
	TransactionsWithMember    int `json:"transactionsWithMember"`
	AnonymousTransactions     int `json:"anonymousTransactions"`
	TransactionDetailsCreated int `json:"transactionDetailsCreated"`
	ProductsCreated           int `json:"productsCreated"`
	VariantsCreated           int `json:"variantsCreated"`
}

// ImportRowError represents error for specific row during import
type ImportRowError struct {
	RowNumber int    `json:"rowNumber"`
	Field     string `json:"field,omitempty"`
	Message   string `json:"message"`
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
