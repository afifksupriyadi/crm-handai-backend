package service

import (
	"context"
	"math"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

// CustomerServiceImpl implements CustomerService interface
type CustomerServiceImpl struct {
	repo customer.CustomerRepository
}

// NewCustomerService creates a new instance of CustomerServiceImpl
func NewCustomerService(repo customer.CustomerRepository) customer.CustomerService {
	return &CustomerServiceImpl{repo: repo}
}

// CreateCustomer creates a new customer
func (s *CustomerServiceImpl) CreateCustomer(ctx context.Context, req *model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	// Check if phone already exists
	existingCustomer, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}

	if existingCustomer != nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
	}

	// Create customer entity
	customer := &model.Customer{
		Name:  req.Name,
		Phone: req.Phone,
	}

	// Save to database
	createdCustomer, err := s.repo.Create(ctx, customer)
	if err != nil {
		return nil, err
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", createdCustomer.ID).Msg("Customer created successfully")
	return createdCustomer.ToResponse(), nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerServiceImpl) GetCustomerByID(ctx context.Context, id int) (*model.CustomerResponse, error) {
	customer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return customer.ToResponse(), nil
}

// GetCustomerByPhone retrieves a customer by phone number
func (s *CustomerServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByName retrieves a customer by name
func (s *CustomerServiceImpl) GetCustomerByName(ctx context.Context, name string) (*model.Customer, error) {
	customer, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetOrCreateCustomer gets existing customer or creates new one
func (s *CustomerServiceImpl) GetOrCreateCustomer(ctx context.Context, name, phone string) (*model.Customer, error) {
	// Try to find by phone first
	customer, err := s.repo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	// If found, return existing customer
	if customer != nil {
		return customer, nil
	}

	// If not found, create new customer
	newCustomer := &model.Customer{
		Name:  name,
		Phone: phone,
	}

	createdCustomer, err := s.repo.Create(ctx, newCustomer)
	if err != nil {
		return nil, err
	}

	logger.FromContext(ctx, 1).Info().
		Int("customer_id", createdCustomer.ID).
		Str("phone", phone).
		Msg("New customer created via import")

	return createdCustomer, nil
}

// GetAllCustomers retrieves all customers with pagination and search
func (s *CustomerServiceImpl) GetAllCustomers(ctx context.Context, req *model.GetCustomersRequest) (*model.CustomerListResponse, error) {
	customers, totalCount, err := s.repo.FindAll(ctx, req.Page, req.Limit, req.Search)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	customerResponses := make([]*model.CustomerResponse, 0, len(customers))
	for _, customer := range customers {
		customerResponses = append(customerResponses, customer.ToResponse())
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalCount) / float64(req.Limit)))

	return &model.CustomerListResponse{
		Data: customerResponses,
		Pagination: model.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerServiceImpl) UpdateCustomer(ctx context.Context, id int, req *model.UpdateCustomerRequest) (*model.CustomerResponse, error) {
	// Check if customer exists
	existingCustomer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if phone is being changed and if new phone already exists
	if req.Phone != existingCustomer.Phone {
		customerWithPhone, err := s.repo.FindByPhone(ctx, req.Phone)
		if err != nil {
			return nil, err
		}

		if customerWithPhone != nil {
			return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
		}
	}

	// Update customer fields
	existingCustomer.Name = req.Name
	existingCustomer.Phone = req.Phone

	// Save to database
	updatedCustomer, err := s.repo.Update(ctx, existingCustomer)
	if err != nil {
		return nil, err
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", updatedCustomer.ID).Msg("Customer updated successfully")
	return updatedCustomer.ToResponse(), nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerServiceImpl) DeleteCustomer(ctx context.Context, id int) error {
	// Check if customer exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", id).Msg("Customer deleted successfully")
	return nil
}
