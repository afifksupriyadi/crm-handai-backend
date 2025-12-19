package service

import (
	"context"
	"strings"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type CustomerServiceImpl struct {
	repo customer.CustomerRepository
	db   *bun.DB
}

// NewCustomerService creates a new instance of CustomerServiceImpl
func NewCustomerService(repo customer.CustomerRepository, db *bun.DB) customer.CustomerService {
	return &CustomerServiceImpl{
		repo: repo,
		db:   db,
	}
}

// CreateCustomer creates a new customer
func (s *CustomerServiceImpl) CreateCustomer(ctx context.Context, req *model.CreateCustomerRequest) (*model.CustomerResponse, error) {
	log := logger.FromContext(ctx, 2)

	existingCustomer, err := s.repo.FindByPhone(ctx, s.db, req.Phone)
	if err != nil {
		log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to check existing customer")
		return nil, err
	}

	if existingCustomer != nil {
		return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
	}

	customer := &model.Customer{Name: req.Name, Phone: req.Phone}
	createdCustomer, err := s.repo.Create(ctx, s.db, customer)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create customer")
		return nil, err
	}

	log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created successfully")
	return mapCustomerToResponse(createdCustomer), nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerServiceImpl) GetCustomerByID(ctx context.Context, id int) (*model.CustomerResponse, error) {
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to get customer by ID")
		return nil, err
	}

	return mapCustomerToResponse(customer), nil
}

// GetCustomerByPhone retrieves a customer by phone number
func (s *CustomerServiceImpl) GetCustomerByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer, err := s.repo.FindByPhone(ctx, s.db, phone)
	if err != nil {
		logger.FromContext(ctx, 2).Error().Err(err).Str("phone", phone).Msg("Failed to get customer by phone")
		return nil, err
	}
	return customer, nil
}

// GetCustomerByName retrieves a customer by name
func (s *CustomerServiceImpl) GetCustomerByName(ctx context.Context, name string) (*model.Customer, error) {
	customer, err := s.repo.FindByName(ctx, s.db, name)
	if err != nil {
		logger.FromContext(ctx, 2).Error().Err(err).Str("name", name).Msg("Failed to get customer by name")
		return nil, err
	}
	return customer, nil
}

// GetOrCreateCustomer gets or creates a customer by name and phone
func (s *CustomerServiceImpl) GetOrCreateCustomer(ctx context.Context, name, phone string) (*model.Customer, error) {
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByPhone(ctx, s.db, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Failed to find customer by phone")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	if customer != nil {
		log.Info().Int("customer_id", customer.ID).Msg("Customer found by phone")
		return customer, nil
	}

	newCustomer := &model.Customer{Name: name, Phone: phone}
	createdCustomer, err := s.repo.Create(ctx, s.db, newCustomer)
	if err != nil {
		log.Error().Err(err).Str("name", name).Str("phone", phone).Msg("Failed to create customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created via import")
	return createdCustomer, nil
}

// GetAllCustomers retrieves all customers with pagination
func (s *CustomerServiceImpl) GetAllCustomers(ctx context.Context, req *model.GetCustomersRequest) (*model.CustomerListResponse, error) {
	log := logger.FromContext(ctx, 2)

	customers, totalCount, err := s.repo.FindAll(ctx, req.Page, req.Limit, req.Search, req.SortOrder)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all customers")
		return nil, err
	}

	customerResponses := make([]*model.CustomerResponse, len(customers))
	for i, customer := range customers {
		customerResponses[i] = mapCustomerToResponse(customer)
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit

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
	log := logger.FromContext(ctx, 2)

	customer, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to find customer for update")
		return nil, err
	}

	if req.Phone != customer.Phone {
		existingWithPhone, err := s.repo.FindByPhone(ctx, s.db, req.Phone)
		if err != nil {
			log.Error().Err(err).Str("phone", req.Phone).Msg("Failed to check phone existence")
			return nil, err
		}
		if existingWithPhone != nil {
			return nil, response.WrapAppError(ctx, nil, response.ErrPhoneAlreadyExists, "Phone number already exists")
		}
	}

	customer.Name = req.Name
	customer.Phone = req.Phone

	updatedCustomer, err := s.repo.Update(ctx, s.db, customer)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to update customer")
		return nil, err
	}

	log.Info().Int("customer_id", updatedCustomer.ID).Msg("Customer updated successfully")
	return mapCustomerToResponse(updatedCustomer), nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerServiceImpl) DeleteCustomer(ctx context.Context, id int) error {
	log := logger.FromContext(ctx, 2)

	_, err := s.repo.FindByID(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Customer not found for deletion")
		return err
	}

	err = s.repo.Delete(ctx, s.db, id)
	if err != nil {
		log.Error().Err(err).Int("id", id).Msg("Failed to delete customer")
		return err
	}

	log.Info().Int("customer_id", id).Msg("Customer deleted successfully")
	return nil
}

// FindOrCreateCustomerWithNameMatching finds or creates customer with fuzzy name matching
func (s *CustomerServiceImpl) FindOrCreateCustomerWithNameMatching(ctx context.Context, tx *bun.Tx, name, phone string) (*model.Customer, bool, error) {
	log := logger.FromContext(ctx, 2)

	existing, err := s.repo.FindByPhone(ctx, tx, phone)
	if err != nil {
		log.Error().Err(err).Str("phone", phone).Msg("Failed to get customer by phone")
		return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check existing customer")
	}

	if existing != nil {
		if namesAreSimilar(existing.Name, name) {
			log.Debug().Int("customer_id", existing.ID).Str("existing_name", existing.Name).Str("import_name", name).Msg("Customer matched by phone and similar name")
			return existing, false, nil
		}

		log.Warn().Int("customer_id", existing.ID).Str("old_name", existing.Name).Str("new_name", name).Msg("Updating customer name due to mismatch")
		existing.Name = name
		existing.UpdatedAt = timePtr(time.Now())
		_, err = s.repo.Update(ctx, tx, existing)
		if err != nil {
			log.Error().Err(err).Int("id", existing.ID).Msg("Failed to update customer name")
			return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update customer")
		}
		return existing, false, nil
	}

	newCustomer := &model.Customer{Name: name, Phone: phone}
	createdCustomer, err := s.repo.Create(ctx, tx, newCustomer)
	if err != nil {
		log.Error().Err(err).Str("name", name).Str("phone", phone).Msg("Failed to create customer")
		return nil, false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	log.Info().Int("customer_id", createdCustomer.ID).Str("name", name).Msg("Customer created in batch import")
	return createdCustomer, true, nil
}

// BulkImportCustomers imports multiple customers from Excel and handles retroactive linking
func (s *CustomerServiceImpl) BulkImportCustomers(ctx context.Context, customers []*model.Customer) (int, int, error) {
	log := logger.FromContext(ctx, 2)

	var newCount, updatedCount int

	err := s.repo.WithTx(ctx, func(tx *bun.Tx) error {
		for _, customer := range customers {
			existing, err := s.repo.FindByPhone(ctx, tx, customer.Phone)
			if err != nil {
				log.Error().Err(err).Str("phone", customer.Phone).Msg("Failed to check existing customer")
				continue
			}

			if existing != nil {
				existing.Name = customer.Name
				existing.UpdatedAt = timePtr(time.Now())
				_, err = s.repo.Update(ctx, tx, existing)
				if err != nil {
					log.Error().Err(err).Int("id", existing.ID).Msg("Failed to update customer")
					continue
				}
				updatedCount++
				log.Info().Int("customer_id", existing.ID).Msg("Customer updated")
			} else {
				linkedCount, err := s.repo.LinkPastTransactions(ctx, tx, customer.Name, 0)
				if err != nil {
					log.Error().Err(err).Str("name", customer.Name).Msg("Failed to check past transactions")
				}

				if linkedCount > 0 {
					customer.UpgradedFromGuest = true
					customer.UpgradedAt = timePtr(time.Now())
					customer.FirstSeenAsGuest = timePtr(time.Now())
					log.Info().Str("name", customer.Name).Int("linked_count", linkedCount).Msg("Found past guest transactions")
				}

				createdCustomer, err := s.repo.Create(ctx, tx, customer)
				if err != nil {
					log.Error().Err(err).Str("name", customer.Name).Msg("Failed to insert customer")
					continue
				}

				if linkedCount > 0 {
					finalLinked, err := s.repo.LinkPastTransactions(ctx, tx, customer.Name, createdCustomer.ID)
					if err != nil {
						log.Error().Err(err).Int("customer_id", createdCustomer.ID).Msg("Failed to link past transactions")
					} else {
						log.Info().Int("customer_id", createdCustomer.ID).Int("linked", finalLinked).Msg("Guest upgraded and transactions linked")
					}
				}

				newCount++
				log.Info().Int("customer_id", createdCustomer.ID).Msg("Customer created")
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to bulk import customers")
		return 0, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to bulk import customers")
	}

	log.Info().Int("new", newCount).Int("updated", updatedCount).Msg("Bulk import completed")
	return newCount, updatedCount, nil
}

// UpgradeGuestToCustomer upgrades guest to registered customer and links past transactions
func (s *CustomerServiceImpl) UpgradeGuestToCustomer(ctx context.Context, guestName string, customer *model.Customer) (int, error) {
	log := logger.FromContext(ctx, 2)

	var linkedCount int
	err := s.repo.WithTx(ctx, func(tx *bun.Tx) error {
		customer.UpgradedFromGuest = true
		customer.UpgradedAt = timePtr(time.Now())
		customer.FirstSeenAsGuest = timePtr(time.Now())

		createdCustomer, err := s.repo.Create(ctx, tx, customer)
		if err != nil {
			log.Error().Err(err).Str("name", customer.Name).Msg("Failed to insert customer")
			return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
		}

		count, err := s.repo.LinkPastTransactions(ctx, tx, guestName, createdCustomer.ID)
		if err != nil {
			log.Error().Err(err).Int("customer_id", createdCustomer.ID).Msg("Failed to link past transactions")
			return err
		}

		linkedCount = count
		log.Info().Int("customer_id", createdCustomer.ID).Int("linked", linkedCount).Msg("Guest upgraded to customer")
		return nil
	})

	return linkedCount, err
}

// mapCustomerToResponse maps Customer model to CustomerResponse DTO
func mapCustomerToResponse(customer *model.Customer) *model.CustomerResponse {
	return &model.CustomerResponse{
		ID:                customer.ID,
		Name:              customer.Name,
		Phone:             customer.Phone,
		CreatedAt:         customer.CreatedAt,
		UpdatedAt:         customer.UpdatedAt,
		UpgradedFromGuest: customer.UpgradedFromGuest,
		UpgradedAt:        customer.UpgradedAt,
		FirstSeenAsGuest:  customer.FirstSeenAsGuest,
	}
}

// namesAreSimilar checks if two names are similar using contains logic
func namesAreSimilar(name1, name2 string) bool {
	n1 := strings.ToLower(strings.TrimSpace(name1))
	n2 := strings.ToLower(strings.TrimSpace(name2))
	return strings.Contains(n1, n2) || strings.Contains(n2, n1)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// ComputeCustomerMetrics computes and stores analytics for a customer
func (s *CustomerServiceImpl) ComputeCustomerMetrics(ctx context.Context, customerID int, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	err := s.repo.ComputeAndStoreMetrics(ctx, s.db, customerID, transactionBatchID)
	if err != nil {
		log.Error().Err(err).Int("customer_id", customerID).Int("transaction_batch_id", transactionBatchID).Msg("Failed to compute customer metrics")
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to compute customer metrics")
	}

	log.Debug().Int("customer_id", customerID).Int("transaction_batch_id", transactionBatchID).Msg("Customer metrics computed")
	return nil
}

// GetCustomersWithRecentTransactions retrieves customers with their latest transaction info from analytics
func (s *CustomerServiceImpl) GetCustomersWithRecentTransactions(ctx context.Context, req *model.GetRecentTransactionsRequest) (*model.CustomerRecentTransactionListResponse, error) {
	log := logger.FromContext(ctx, 2)

	// Get customers with metrics from repository
	customersWithMetrics, totalCount, err := s.repo.FindAllWithRecentTransactions(ctx, req.Page, req.Limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get customers with recent transactions")
		return nil, err
	}

	// Map to response DTOs
	now := time.Now()
	customerResponses := make([]*model.CustomerRecentTransactionResponse, 0, len(customersWithMetrics))

	for _, cwm := range customersWithMetrics {
		resp := &model.CustomerRecentTransactionResponse{
			ID:                     cwm.Customer.ID,
			Name:                   cwm.Customer.Name,
			Phone:                  cwm.Customer.Phone,
			LastTransactionDate:    cwm.CustomerMetric.LastTransactionDate,
			TotalTransactions:      cwm.CustomerMetric.TotalTransactions,
			TotalSpent:             cwm.CustomerMetric.TotalSpent,
			Segment:                cwm.CustomerMetric.Segment,
			IsLoyal:                cwm.CustomerMetric.IsLoyal,
			AvgDaysBetweenPurchase: cwm.CustomerMetric.AvgDaysBetweenPurchase,
			ChurnRiskScore:         cwm.CustomerMetric.ChurnRiskScore,
		}

		// Calculate days since last transaction
		if cwm.CustomerMetric.LastTransactionDate != nil {
			daysSince := int(now.Sub(*cwm.CustomerMetric.LastTransactionDate).Hours() / 24)
			resp.DaysSinceLastTransaction = &daysSince
		}

		customerResponses = append(customerResponses, resp)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.Limit - 1) / req.Limit

	return &model.CustomerRecentTransactionListResponse{
		Data: customerResponses,
		Pagination: model.PaginationMeta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}
