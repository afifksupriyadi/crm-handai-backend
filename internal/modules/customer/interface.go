package customer

import (
	"context"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
)

type CustomerService interface {
	GetCustomerByID(ctx context.Context, id int) (*model.Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error)
	GetCustomerByName(ctx context.Context, name string) (*model.Customer, error)
	GetOrCreateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	GetCustomerCount(ctx context.Context) (int, error)
}
