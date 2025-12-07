package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type CustomerServiceImpl struct {
	customerRepo repository.CustomerRepository
}

func NewCustomerService(customerRepo repository.CustomerRepository) customer.CustomerService {
	return &CustomerServiceImpl{
		customerRepo: customerRepo,
	}
}

func (s *CustomerServiceImpl) GetCustomerByID(ctx context.Context, id int) (*model.Customer, error) {
	customer, err := s.customerRepo.GetCustomerByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customer")
	}

	return customer, nil
}

func (s *CustomerServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer, err := s.customerRepo.GetCustomerByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customer")
	}

	return customer, nil
}

func (s *CustomerServiceImpl) GetCustomerByName(ctx context.Context, name string) (*model.Customer, error) {
	customer, err := s.customerRepo.GetCustomerByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customer")
	}

	return customer, nil
}

func (s *CustomerServiceImpl) GetOrCreateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	// Try to get existing customer
	existing, err := s.customerRepo.GetCustomerByPhone(ctx, customer.Phone)

	// If found, return it
	if err == nil {
		return existing, nil
	}

	// If not found, create new
	if errors.Is(err, sql.ErrNoRows) {
		err = s.customerRepo.CreateCustomer(ctx, customer)
		if err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
		}
		return customer, nil
	}

	// Other database error
	return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing customer")
}

func (s *CustomerServiceImpl) GetCustomerCount(ctx context.Context) (int, error) {
	count, err := s.customerRepo.GetCustomerCount(ctx)
	if err != nil {
		return 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customer count")
	}

	return count, nil
}
