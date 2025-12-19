package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type CustomerBatchRepository interface {
	Create(ctx context.Context, db bun.IDB, batch *model.CustomerBatch) (*model.CustomerBatch, error)
	Update(ctx context.Context, db bun.IDB, batch *model.CustomerBatch) (*model.CustomerBatch, error)
	GetLatestActive(ctx context.Context) (*model.CustomerBatch, error)
	GetByDate(ctx context.Context, batchDate time.Time) (*model.CustomerBatch, error)
	SetInactive(ctx context.Context, db bun.IDB, batchID int) error
	SetActive(ctx context.Context, db bun.IDB, batchID int) error
}

type CustomerBatchRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerBatchRepository creates a new instance of CustomerBatchRepositoryImpl
func NewCustomerBatchRepository(db *bun.DB) CustomerBatchRepository {
	return &CustomerBatchRepositoryImpl{db: db}
}

// Create inserts a new customer batch
func (r *CustomerBatchRepositoryImpl) Create(ctx context.Context, db bun.IDB, batch *model.CustomerBatch) (*model.CustomerBatch, error) {
	_, err := db.NewInsert().
		Model(batch).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create customer batch")
		return nil, err
	}

	return batch, nil
}

// Update updates an existing customer batch
func (r *CustomerBatchRepositoryImpl) Update(ctx context.Context, db bun.IDB, batch *model.CustomerBatch) (*model.CustomerBatch, error) {
	_, err := db.NewUpdate().
		Model(batch).
		WherePK().
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to update customer batch")
		return nil, err
	}

	return batch, nil
}

// GetLatestActive retrieves the latest active customer batch
func (r *CustomerBatchRepositoryImpl) GetLatestActive(ctx context.Context) (*model.CustomerBatch, error) {
	batch := new(model.CustomerBatch)

	err := r.db.NewSelect().
		Model(batch).
		Where("is_active = ?", true).
		Order("batch_date DESC").
		Limit(1).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get latest active customer batch")
		return nil, err
	}

	return batch, nil
}

// GetByDate retrieves customer batch by batch date
func (r *CustomerBatchRepositoryImpl) GetByDate(ctx context.Context, batchDate time.Time) (*model.CustomerBatch, error) {
	batch := new(model.CustomerBatch)

	err := r.db.NewSelect().
		Model(batch).
		Where("batch_date = ?", batchDate).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get customer batch by date")
		return nil, err
	}

	return batch, nil
}

// SetInactive sets customer batch as inactive
func (r *CustomerBatchRepositoryImpl) SetInactive(ctx context.Context, db bun.IDB, batchID int) error {
	_, err := db.NewUpdate().
		Model((*model.CustomerBatch)(nil)).
		Set("is_active = ?", false).
		Where("id = ?", batchID).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("batch_id", batchID).Msg("Failed to set customer batch inactive")
		return err
	}

	return nil
}

// SetActive sets customer batch as active (and deactivates others)
func (r *CustomerBatchRepositoryImpl) SetActive(ctx context.Context, db bun.IDB, batchID int) error {
	_, err := db.NewUpdate().
		Model((*model.CustomerBatch)(nil)).
		Set("is_active = ?", false).
		Where("is_active = ?", true).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to deactivate old customer batches")
		return err
	}

	_, err = db.NewUpdate().
		Model((*model.CustomerBatch)(nil)).
		Set("is_active = ?", true).
		Where("id = ?", batchID).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("batch_id", batchID).Msg("Failed to set customer batch active")
		return err
	}

	return nil
}
