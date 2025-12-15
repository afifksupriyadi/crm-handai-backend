package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

// CustomerRepositoryImpl implements CustomerRepository interface
type CustomerRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerRepository creates a new instance of CustomerRepositoryImpl
func NewCustomerRepository(db *bun.DB) customer.CustomerRepository {
	return &CustomerRepositoryImpl{db: db}
}

// Create inserts a new customer into the database
func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	customer.CreatedAt = time.Now()

	_, err := r.db.NewInsert().
		Model(customer).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create customer")
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", customer.ID).Msg("Customer created successfully")
	return customer, nil
}

// FindByID retrieves a customer by ID
func (r *CustomerRepositoryImpl) FindByID(ctx context.Context, id int) (*model.Customer, error) {
	customer := new(model.Customer)

	err := r.db.NewSelect().
		Model(customer).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, response.WrapAppError(ctx, err, response.ErrCustomerNotFound, "Customer not found")
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by ID")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindByPhone retrieves a customer by phone number
func (r *CustomerRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*model.Customer, error) {
	customer := new(model.Customer)

	err := r.db.NewSelect().
		Model(customer).
		Where("phone = ?", phone).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if not found (not an error)
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by phone")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindByName retrieves a customer by name (exact match)
func (r *CustomerRepositoryImpl) FindByName(ctx context.Context, name string) (*model.Customer, error) {
	customer := new(model.Customer)

	err := r.db.NewSelect().
		Model(customer).
		Where("name = ?", name).
		Where("deleted_at IS NULL").
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if not found (not an error)
		}
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to find customer by name")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to find customer")
	}

	return customer, nil
}

// FindAll retrieves all customers with pagination and search
func (r *CustomerRepositoryImpl) FindAll(ctx context.Context, page, limit int, search string) ([]*model.Customer, int, error) {
	var customers []*model.Customer

	query := r.db.NewSelect().
		Model(&customers).
		Where("deleted_at IS NULL")

	// Apply search filter if provided
	if search != "" {
		searchPattern := fmt.Sprintf("%%%s%%", search)
		query = query.Where("name ILIKE ? OR phone ILIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	totalCount, err := query.Count(ctx)
	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to count customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to count customers")
	}

	// Apply pagination
	offset := (page - 1) * limit
	err = query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customers")
		return nil, 0, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get customers")
	}

	return customers, totalCount, nil
}

// Update updates an existing customer
func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	now := time.Now()
	customer.UpdatedAt = &now

	result, err := r.db.NewUpdate().
		Model(customer).
		Column("name", "phone", "updated_at").
		Where("id = ?", customer.ID).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to update customer")
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, response.WrapAppError(ctx, nil, response.ErrCustomerNotFound, "Customer not found")
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", customer.ID).Msg("Customer updated successfully")
	return customer, nil
}

// Delete soft deletes a customer by setting deleted_at timestamp
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id int) error {
	now := time.Now()

	result, err := r.db.NewUpdate().
		Model((*model.Customer)(nil)).
		Set("deleted_at = ?", now).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to delete customer")
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to delete customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return response.WrapAppError(ctx, nil, response.ErrCustomerNotFound, "Customer not found")
	}

	logger.FromContext(ctx, 1).Info().Int("customer_id", id).Msg("Customer deleted successfully")
	return nil
}

// Exists checks if a customer exists by ID
func (r *CustomerRepositoryImpl) Exists(ctx context.Context, id int) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*model.Customer)(nil)).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to check customer existence")
		return false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check customer existence")
	}

	return count > 0, nil
}
