package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type TransactionBatchRepository interface {
	Create(ctx context.Context, db bun.IDB, batch *model.TransactionBatch) (*model.TransactionBatch, error)
	Update(ctx context.Context, db bun.IDB, batch *model.TransactionBatch) (*model.TransactionBatch, error)
	GetByID(ctx context.Context, batchID int) (*model.TransactionBatch, error)
	GetByCustomerBatchID(ctx context.Context, customerBatchID int) ([]*model.TransactionBatch, error)
	GetCustomerIDsInBatch(ctx context.Context, transactionBatchID int) ([]int, error)
}

type TransactionBatchRepositoryImpl struct {
	db *bun.DB
}

// NewTransactionBatchRepository creates a new instance of TransactionBatchRepositoryImpl
func NewTransactionBatchRepository(db *bun.DB) TransactionBatchRepository {
	return &TransactionBatchRepositoryImpl{db: db}
}

// Create inserts a new transaction batch
func (r *TransactionBatchRepositoryImpl) Create(ctx context.Context, db bun.IDB, batch *model.TransactionBatch) (*model.TransactionBatch, error) {
	_, err := db.NewInsert().
		Model(batch).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create transaction batch")
		return nil, err
	}

	return batch, nil
}

// Update updates an existing transaction batch
func (r *TransactionBatchRepositoryImpl) Update(ctx context.Context, db bun.IDB, batch *model.TransactionBatch) (*model.TransactionBatch, error) {
	_, err := db.NewUpdate().
		Model(batch).
		WherePK().
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to update transaction batch")
		return nil, err
	}

	return batch, nil
}

// GetByID retrieves transaction batch by ID
func (r *TransactionBatchRepositoryImpl) GetByID(ctx context.Context, batchID int) (*model.TransactionBatch, error) {
	batch := new(model.TransactionBatch)

	err := r.db.NewSelect().
		Model(batch).
		Where("id = ?", batchID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get transaction batch by ID")
		return nil, err
	}

	return batch, nil
}

// GetByCustomerBatchID retrieves all transaction batches linked to a customer batch
func (r *TransactionBatchRepositoryImpl) GetByCustomerBatchID(ctx context.Context, customerBatchID int) ([]*model.TransactionBatch, error) {
	var batches []*model.TransactionBatch

	err := r.db.NewSelect().
		Model(&batches).
		Where("customer_batch_id = ?", customerBatchID).
		Order("batch_date DESC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get transaction batches by customer batch ID")
		return nil, err
	}

	return batches, nil
}

// GetCustomerIDsInBatch returns all customer IDs that have transactions in this batch
func (r *TransactionBatchRepositoryImpl) GetCustomerIDsInBatch(ctx context.Context, transactionBatchID int) ([]int, error) {
	var customerIDs []int

	err := r.db.NewSelect().
		Table("transactions").
		ColumnExpr("DISTINCT customer_id").
		Where("customer_id IS NOT NULL").
		Where("deleted_at IS NULL").
		Join("JOIN transaction_batches tb ON tb.batch_date = DATE(transactions.transaction_date)").
		Where("tb.id = ?", transactionBatchID).
		Scan(ctx, &customerIDs)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("batch_id", transactionBatchID).Msg("Failed to get customer IDs in batch")
		return nil, err
	}

	return customerIDs, nil
}
