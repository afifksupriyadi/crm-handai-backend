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

	// Get customer detail with full information
	GetCustomerDetail(ctx context.Context, id int, month *time.Time) (*model.CustomerDetailResponse, error)

	// Batch import operations
	BulkImportCustomers(ctx context.Context, customers []*model.Customer) (int, int, error)

	// Manual upgrade operations (for future API endpoint)
	UpgradeGuestToCustomer(ctx context.Context, guestName string, customer *model.Customer) (int, error)

	// Analytics
	ComputeCustomerMetrics(ctx context.Context, customerID int, transactionBatchID int) error

	// Recent transactions
	GetCustomersWithRecentTransactions(ctx context.Context, req *model.GetRecentTransactionsRequest) (*model.CustomerRecentTransactionListResponse, error)
}

// CustomerRepository defines the contract for customer data access
type CustomerRepository interface {
	// CRUD operations - all accept bun.IDB (can be *bun.DB or *bun.Tx)
	Create(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error)
	FindByID(ctx context.Context, db bun.IDB, id int) (*model.Customer, error)
	FindByPhone(ctx context.Context, db bun.IDB, phone string) (*model.Customer, error)
	FindByName(ctx context.Context, db bun.IDB, name string) (*model.Customer, error)
	FindAll(ctx context.Context, page, limit int, search, sortOrder string) ([]*model.CustomerWithLatestMetrics, int, error) // <-- Ubah ini
	Update(ctx context.Context, db bun.IDB, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, db bun.IDB, id int) error
	Exists(ctx context.Context, db bun.IDB, id int) (bool, error)

	// Get comprehensive customer detail data
	GetCustomerDetailData(ctx context.Context, customerID int, month *time.Time) (*model.CustomerDetailData, error)

	// Transaction helper
	WithTx(ctx context.Context, fn func(*bun.Tx) error) error

	// Batch import operations
	LinkPastTransactions(ctx context.Context, db bun.IDB, guestName string, customerID int) (int, error)
	ComputeAndStoreMetrics(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error

	// Recent transactions - joins with analytics.customer_metrics
	FindAllWithRecentTransactions(ctx context.Context, page, limit int) ([]*model.CustomerWithMetrics, int, error)
}
