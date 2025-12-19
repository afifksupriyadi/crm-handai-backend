package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type ImportLogRepository interface {
	Create(ctx context.Context, db bun.IDB, log *model.ImportLog) (*model.ImportLog, error)
	LinkToBatches(ctx context.Context, db bun.IDB, logID int, customerBatchID *int, transactionBatchID *int) error
	HasCustomerImportSinceDate(ctx context.Context, date time.Time) (bool, error)
	GetLatestCustomerImport(ctx context.Context) (*model.ImportLog, error)
	GetImportLogsByType(ctx context.Context, importType string) ([]*model.ImportLog, error)
}

type ImportLogRepositoryImpl struct {
	db *bun.DB
}

// NewImportLogRepository creates a new instance of ImportLogRepositoryImpl
func NewImportLogRepository(db *bun.DB) ImportLogRepository {
	return &ImportLogRepositoryImpl{db: db}
}

// Create inserts a new import log
func (r *ImportLogRepositoryImpl) Create(ctx context.Context, db bun.IDB, log *model.ImportLog) (*model.ImportLog, error) {
	_, err := db.NewInsert().
		Model(log).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create import log")
		return nil, err
	}

	return log, nil
}

// LinkToBatches links import log to customer and/or transaction batches
func (r *ImportLogRepositoryImpl) LinkToBatches(ctx context.Context, db bun.IDB, logID int, customerBatchID *int, transactionBatchID *int) error {
	_, err := db.NewUpdate().
		Model((*model.ImportLog)(nil)).
		Set("customer_batch_id = ?", customerBatchID).
		Set("transaction_batch_id = ?", transactionBatchID).
		Where("id = ?", logID).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("log_id", logID).Msg("Failed to link import log to batches")
		return err
	}

	return nil
}

// HasCustomerImportSinceDate checks if there's a successful customer import since given date
func (r *ImportLogRepositoryImpl) HasCustomerImportSinceDate(ctx context.Context, date time.Time) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*model.ImportLog)(nil)).
		Where("import_type = ?", "CUSTOMER").
		Where("file_date >= ?", date).
		Where("status = ?", "SUCCESS").
		Count(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to check customer import since date")
		return false, err
	}

	return count > 0, nil
}

// GetLatestCustomerImport retrieves the latest successful customer import
func (r *ImportLogRepositoryImpl) GetLatestCustomerImport(ctx context.Context) (*model.ImportLog, error) {
	log := new(model.ImportLog)

	err := r.db.NewSelect().
		Model(log).
		Where("import_type = ?", "CUSTOMER").
		Where("status = ?", "SUCCESS").
		Order("file_date DESC").
		Limit(1).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get latest customer import")
		return nil, err
	}

	return log, nil
}

// GetImportLogsByType retrieves all import logs by type
func (r *ImportLogRepositoryImpl) GetImportLogsByType(ctx context.Context, importType string) ([]*model.ImportLog, error) {
	var logs []*model.ImportLog

	err := r.db.NewSelect().
		Model(&logs).
		Where("import_type = ?", importType).
		Order("imported_at DESC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Str("import_type", importType).Msg("Failed to get import logs by type")
		return nil, err
	}

	return logs, nil
}
