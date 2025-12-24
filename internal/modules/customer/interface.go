package customer

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/uptrace/bun"
)

// CustomerService defines the contract for customer-related operations
type CustomerService interface {
	// Regular CRUD operations
	CreateCustomer(ctx context.Context, req *model.CreateCustomerRequest) (*model.CustomerResponse, error)
	GetCustomerByID(ctx context.Context, id int) (*model.CustomerResponse, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error)
	GetCustomerByName(ctx context.Context, name string) (*model.Customer, error)
	GetOrCreateCustomer(ctx context.Context, name, phone string) (*model.Customer, error)
	GetAllCustomers(ctx context.Context, req *model.GetCustomersRequest) (*model.CustomerListResponse, error)
	UpdateCustomer(ctx context.Context, id int, req *model.UpdateCustomerRequest) (*model.CustomerResponse, error)
	DeleteCustomer(ctx context.Context, id int) error

	// Get customer detail with full information - UPDATED SIGNATURE
	GetCustomerDetail(ctx context.Context, id int, year *int, month *int) (*model.CustomerDetailResponse, error)

	// Batch import operations
	BulkImportCustomers(ctx context.Context, customers []*model.Customer) (int, int, error)

	// Manual upgrade operations (for future API endpoint)
	UpgradeGuestToCustomer(ctx context.Context, guestName string, customer *model.Customer) (int, error)

	// Analytics
	ComputeCustomerMetrics(ctx context.Context, customerID int, transactionBatchID int) error

	// Recent transactions
	GetCustomersWithRecentTransactions(ctx context.Context, req *model.GetRecentTransactionsRequest) (*model.CustomerRecentTransactionListResponse, error)

	// GetCustomerPredictions retrieves prediction history for a customer
	GetCustomerPredictions(ctx context.Context, customerID int, limit int) (*model.CustomerPredictionListResponse, error)
}

// WindowCalculatorService defines the contract for window calculation
type WindowCalculatorService interface {
	CalculateWindows(ctx context.Context, importStartDate, importEndDate time.Time, pendingStart *time.Time) ([]Window, *time.Time, error)
}

// PredictionCalculatorService defines the contract for prediction calculation
type PredictionCalculatorService interface {
	CheckEligibility(ctx context.Context, customerID int) (bool, error)
	CalculatePrediction(ctx context.Context, customerID int, transactionBatchID int) (*model.CustomerPrediction, error)
}

// PredictionValidatorService defines the contract for prediction validation
type PredictionValidatorService interface {
	ValidatePrediction(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction, windowEndDate time.Time) error
}

// SegmentDeterminerService defines the contract for segment determination
type SegmentDeterminerService interface {
	DetermineSegment(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error
}

// PredictionOrchestratorService orchestrates the entire prediction process
type PredictionOrchestratorService interface {
	ProcessPredictions(ctx context.Context, importStartDate, importEndDate time.Time, transactionBatchID int) error
}

// ProductPredictionCalculatorService calculates product predictions based on purchase history
type ProductPredictionCalculatorService interface {
	// CalculateProductPredictions analyzes purchase history and predicts products customer will buy
	CalculateProductPredictions(ctx context.Context, customerID int, predictionID int) ([]*model.CustomerPredictedProduct, error)
}

// CustomerRepository defines the contract for customer data access
type CustomerRepository interface {
	// CRUD operations - all accept bun.IDB (can be *bun.DB or *bun.Tx)
	Create(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error)
	FindByID(ctx context.Context, db bun.IDB, id int) (*model.Customer, error)
	FindByPhone(ctx context.Context, db bun.IDB, phone string) (*model.Customer, error)
	FindByName(ctx context.Context, db bun.IDB, name string) (*model.Customer, error)
	FindAll(ctx context.Context, page, limit int, search, sortOrder string) ([]*model.CustomerWithLatestMetrics, int, error)
	Update(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, db bun.IDB, id int) error
	Exists(ctx context.Context, db bun.IDB, id int) (bool, error)

	// Get comprehensive customer detail data - UPDATED SIGNATURE
	GetCustomerDetailData(ctx context.Context, customerID int, year *int, month *int) (*model.CustomerDetailData, error)

	// Transaction helper
	WithTx(ctx context.Context, fn func(*bun.Tx) error) error

	// Batch import operations
	CountPastGuestTransactions(ctx context.Context, db bun.IDB, guestName string) (int, error)
	LinkPastTransactions(ctx context.Context, db bun.IDB, guestName string, customerID int) (int, error)
	ComputeAndStoreMetrics(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error

	// Recent transactions - joins with analytics.customer_metrics
	FindAllWithRecentTransactions(ctx context.Context, page, limit int) ([]*model.CustomerWithMetrics, int, error)

	// GetCustomerPredictions retrieves prediction history
	GetCustomerPredictions(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error)

	// GetCustomersBySegment retrieves customers by segment type
	GetCustomersBySegment(ctx context.Context, segmentType string, limit int) ([]*model.CustomerWithLatestMetrics, error)

	// GetCustomerPurchaseCounts returns weekly and monthly product purchase counts
	GetCustomerPurchaseCounts(ctx context.Context, customerID int) (weeklyCount int, monthlyCount int)

	// GetCustomerTransactionCount returns total number of transactions for a customer
	GetCustomerTransactionCount(ctx context.Context, customerID int) int
}

// Window represents a 7-day processing window
type Window struct {
	StartDate time.Time
	EndDate   time.Time
}
