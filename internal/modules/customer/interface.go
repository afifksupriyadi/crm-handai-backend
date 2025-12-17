package customer

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/uptrace/bun"
)

// CustomerService defines the contract for customer-related operations.
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

	// Batch import operations (with transaction support)
	FindOrCreateCustomerWithNameMatching(ctx context.Context, tx *bun.Tx, name, phone string) (*model.Customer, bool, error)
	UpdateCustomerMetrics(ctx context.Context, tx *bun.Tx, customerID int) error
}

// CustomerRepository defines the contract for customer data access.
type CustomerRepository interface {
	// Regular CRUD operations (without transaction)
	Create(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	FindByID(ctx context.Context, id int) (*model.Customer, error)
	FindByPhone(ctx context.Context, phone string) (*model.Customer, error)
	FindByName(ctx context.Context, name string) (*model.Customer, error)
	FindAll(ctx context.Context, page, limit int, search, sortOrder string) ([]*model.Customer, int, error)
	Update(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)

	// Batch import operations (with transaction support)
	CreateWithTx(ctx context.Context, tx *bun.Tx, customer *model.Customer) (*model.Customer, error)
	FindByPhoneWithTx(ctx context.Context, tx *bun.Tx, phone string) (*model.Customer, error)
	UpdateWithTx(ctx context.Context, tx *bun.Tx, customer *model.Customer) (*model.Customer, error)
	UpdateCustomerMetrics(ctx context.Context, tx *bun.Tx, customerID int) error
	ComputeAndStoreMetrics(ctx context.Context, customerID int, batchID int) error
}
