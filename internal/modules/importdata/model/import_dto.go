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
