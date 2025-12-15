package customer

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
)

// CustomerService defines the contract for customer-related operations.
type CustomerService interface {
	CreateCustomer(ctx context.Context, req *model.CreateCustomerRequest) (*model.CustomerResponse, error)
	GetCustomerByID(ctx context.Context, id int) (*model.CustomerResponse, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error)
	GetCustomerByName(ctx context.Context, name string) (*model.Customer, error)
	GetOrCreateCustomer(ctx context.Context, name, phone string) (*model.Customer, error)
	GetAllCustomers(ctx context.Context, req *model.GetCustomersRequest) (*model.CustomerListResponse, error)
	UpdateCustomer(ctx context.Context, id int, req *model.UpdateCustomerRequest) (*model.CustomerResponse, error)
	DeleteCustomer(ctx context.Context, id int) error
}

// CustomerRepository defines the contract for customer data access.
type CustomerRepository interface {
	Create(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	FindByID(ctx context.Context, id int) (*model.Customer, error)
	FindByPhone(ctx context.Context, phone string) (*model.Customer, error)
	FindByName(ctx context.Context, name string) (*model.Customer, error)
	FindAll(ctx context.Context, page, limit int, search string) ([]*model.Customer, int, error)
	Update(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, id int) error
	Exists(ctx context.Context, id int) (bool, error)
}
