package repository

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/uptrace/bun"
)

type ImportLogRepository interface {
	CreateImportLog(ctx context.Context, log *model.ImportLog) error
	HasCustomerImportSinceDate(ctx context.Context, date time.Time) (bool, error)
	GetLatestCustomerImport(ctx context.Context) (*model.ImportLog, error)
	GetImportLogsByType(ctx context.Context, importType string) ([]*model.ImportLog, error)
}

type ImportLogRepositoryImpl struct {
	db *bun.DB
}

func NewImportLogRepository(db *bun.DB) ImportLogRepository {
	return &ImportLogRepositoryImpl{db: db}
}

func (r *ImportLogRepositoryImpl) CreateImportLog(ctx context.Context, log *model.ImportLog) error {
	_, err := r.db.NewInsert().
		Model(log).
		Exec(ctx)

	return err
}

func (r *ImportLogRepositoryImpl) HasCustomerImportSinceDate(ctx context.Context, date time.Time) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*model.ImportLog)(nil)).
		Where("import_type = ?", "CUSTOMER").
		Where("file_date >= ?", date).
		Where("status = ?", "SUCCESS").
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *ImportLogRepositoryImpl) GetLatestCustomerImport(ctx context.Context) (*model.ImportLog, error) {
	log := new(model.ImportLog)
	err := r.db.NewSelect().
		Model(log).
		Where("import_type = ?", "CUSTOMER").
		Where("status = ?", "SUCCESS").
		Order("file_date DESC").
		Limit(1).
		Scan(ctx)

	return log, err
}

func (r *ImportLogRepositoryImpl) GetImportLogsByType(ctx context.Context, importType string) ([]*model.ImportLog, error) {
	var logs []*model.ImportLog
	err := r.db.NewSelect().
		Model(&logs).
		Where("import_type = ?", importType).
		Order("imported_at DESC").
		Scan(ctx)

	return logs, err
}
