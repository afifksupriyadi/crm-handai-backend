package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

// CustomerPredictionRepository defines the contract for customer prediction data access
type CustomerPredictionRepository interface {
	GetByCustomer(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error)
	GetByCustomerValidated(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error)
	GetPendingValidations(ctx context.Context, db bun.IDB, windowEndDate time.Time) ([]*model.CustomerPrediction, error)
	Create(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction) (*model.CustomerPrediction, error)
	Update(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction) (*model.CustomerPrediction, error)
	DeleteOldest(ctx context.Context, db bun.IDB, customerID int) error
	CountByCustomer(ctx context.Context, customerID int) (int, error)
	GetCustomerTransactionDates(ctx context.Context, customerID int, limit int) ([]time.Time, error)
	CheckCustomerHasTransactionAfter(ctx context.Context, db bun.IDB, customerID int, afterDate time.Time, beforeOrEqualDate time.Time) (bool, *time.Time, error)
	CountUniqueTransactionDates(ctx context.Context, customerID int) (int, error)
	GetCustomerIDsWithTransactionsInWindow(ctx context.Context, db bun.IDB, startDate, endDate time.Time) ([]int, error)
}

type CustomerPredictionRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerPredictionRepository creates a new instance
func NewCustomerPredictionRepository(db *bun.DB) CustomerPredictionRepository {
	return &CustomerPredictionRepositoryImpl{db: db}
}

// GetByCustomer retrieves customer predictions ordered by created_at DESC
func (r *CustomerPredictionRepositoryImpl) GetByCustomer(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error) {
	var predictions []*model.CustomerPrediction

	err := r.db.NewSelect().
		Model(&predictions).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to get customer predictions")
		return nil, err
	}

	return predictions, nil
}

// GetByCustomerValidated retrieves only validated predictions (not NULL)
func (r *CustomerPredictionRepositoryImpl) GetByCustomerValidated(ctx context.Context, customerID int, limit int) ([]*model.CustomerPrediction, error) {
	var predictions []*model.CustomerPrediction

	err := r.db.NewSelect().
		Model(&predictions).
		Where("customer_id = ?", customerID).
		Where("is_predicted_correct IS NOT NULL").
		Order("created_at DESC").
		Limit(limit).
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to get validated predictions")
		return nil, err
	}

	return predictions, nil
}

// GetPendingValidations retrieves all predictions that need validation
func (r *CustomerPredictionRepositoryImpl) GetPendingValidations(ctx context.Context, db bun.IDB, windowEndDate time.Time) ([]*model.CustomerPrediction, error) {
	var predictions []*model.CustomerPrediction

	err := db.NewSelect().
		Model(&predictions).
		Where("is_predicted_correct IS NULL").
		Where("predicted_next_purchase_date <= ?", windowEndDate).
		Order("customer_id ASC, created_at ASC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get pending validations")
		return nil, err
	}

	return predictions, nil
}

// Create inserts a new customer prediction
func (r *CustomerPredictionRepositoryImpl) Create(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction) (*model.CustomerPrediction, error) {
	_, err := db.NewInsert().
		Model(prediction).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", prediction.CustomerID).Msg("Failed to create prediction")
		return nil, err
	}

	return prediction, nil
}

// Update updates an existing customer prediction
func (r *CustomerPredictionRepositoryImpl) Update(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction) (*model.CustomerPrediction, error) {
	_, err := db.NewUpdate().
		Model(prediction).
		WherePK().
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("id", prediction.ID).Msg("Failed to update prediction")
		return nil, err
	}

	return prediction, nil
}

// DeleteOldest deletes the oldest prediction for a customer
func (r *CustomerPredictionRepositoryImpl) DeleteOldest(ctx context.Context, db bun.IDB, customerID int) error {
	var oldestID int
	err := db.NewSelect().
		Model((*model.CustomerPrediction)(nil)).
		Column("id").
		Where("customer_id = ?", customerID).
		Order("created_at ASC").
		Limit(1).
		Scan(ctx, &oldestID)

	if err == sql.ErrNoRows {
		return nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to find oldest prediction")
		return err
	}

	_, err = db.NewDelete().
		Model((*model.CustomerPrediction)(nil)).
		Where("id = ?", oldestID).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("id", oldestID).Msg("Failed to delete oldest prediction")
		return err
	}

	return nil
}

// CountByCustomer counts total predictions for a customer
func (r *CustomerPredictionRepositoryImpl) CountByCustomer(ctx context.Context, customerID int) (int, error) {
	count, err := r.db.NewSelect().
		Model((*model.CustomerPrediction)(nil)).
		Where("customer_id = ?", customerID).
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to count predictions")
		return 0, err
	}

	return count, nil
}

// GetCustomerTransactionDates retrieves transaction dates for prediction calculation
func (r *CustomerPredictionRepositoryImpl) GetCustomerTransactionDates(ctx context.Context, customerID int, limit int) ([]time.Time, error) {
	var dates []time.Time

	err := r.db.NewSelect().
		Table("transactions").
		Column("transaction_date").
		Where("customer_id = ?", customerID).
		Where("deleted_at IS NULL").
		Order("transaction_date DESC").
		Limit(limit).
		Scan(ctx, &dates)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to get transaction dates")
		return nil, err
	}

	return dates, nil
}

// CheckCustomerHasTransactionAfter checks if customer has transaction in date range
// Returns: (hasTransaction, actualDate, error)
func (r *CustomerPredictionRepositoryImpl) CheckCustomerHasTransactionAfter(ctx context.Context, db bun.IDB, customerID int, afterDate time.Time, beforeOrEqualDate time.Time) (bool, *time.Time, error) {
	var firstDate time.Time

	err := db.NewSelect().
		Table("transactions").
		Column("transaction_date").
		Where("customer_id = ?", customerID).
		Where("transaction_date > ?", afterDate).
		Where("transaction_date <= ?", beforeOrEqualDate).
		Where("deleted_at IS NULL").
		Order("transaction_date ASC").
		Limit(1).
		Scan(ctx, &firstDate)

	if err == sql.ErrNoRows {
		return false, nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to check transaction")
		return false, nil, err
	}

	return true, &firstDate, nil
}

// CountUniqueTransactionDates counts unique transaction dates for eligibility check
func (r *CustomerPredictionRepositoryImpl) CountUniqueTransactionDates(ctx context.Context, customerID int) (int, error) {
	var count int

	err := r.db.NewSelect().
		Table("transactions").
		ColumnExpr("COUNT(DISTINCT DATE(transaction_date))").
		Where("customer_id = ?", customerID).
		Where("deleted_at IS NULL").
		Scan(ctx, &count)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to count unique dates")
		return 0, err
	}

	return count, nil
}

// GetCustomerIDsWithTransactionsInWindow gets unique customer IDs with transactions in date range
func (r *CustomerPredictionRepositoryImpl) GetCustomerIDsWithTransactionsInWindow(ctx context.Context, db bun.IDB, startDate, endDate time.Time) ([]int, error) {
	var customerIDs []int

	err := db.NewSelect().
		Table("transactions").
		ColumnExpr("DISTINCT customer_id").
		Where("customer_id IS NOT NULL").
		Where("transaction_date >= ?", startDate).
		Where("transaction_date <= ?", endDate).
		Where("deleted_at IS NULL").
		Scan(ctx, &customerIDs)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customer IDs in window")
		return nil, err
	}

	return customerIDs, nil
}
