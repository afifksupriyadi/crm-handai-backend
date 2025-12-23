package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

// ImportTrackerRepository defines the contract for import tracker data access
type ImportTrackerRepository interface {
	GetLatest(ctx context.Context) (*model.ImportTracker, error)
	Create(ctx context.Context, db bun.IDB, tracker *model.ImportTracker) (*model.ImportTracker, error)
	Update(ctx context.Context, db bun.IDB, tracker *model.ImportTracker) (*model.ImportTracker, error)
}

type ImportTrackerRepositoryImpl struct {
	db *bun.DB
}

// NewImportTrackerRepository creates a new instance of ImportTrackerRepositoryImpl
func NewImportTrackerRepository(db *bun.DB) ImportTrackerRepository {
	return &ImportTrackerRepositoryImpl{db: db}
}

// GetLatest retrieves the singleton import tracker record
func (r *ImportTrackerRepositoryImpl) GetLatest(ctx context.Context) (*model.ImportTracker, error) {
	tracker := new(model.ImportTracker)

	err := r.db.NewSelect().
		Model(tracker).
		Order("id DESC").
		Limit(1).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to get latest import tracker")
		return nil, err
	}

	return tracker, nil
}

// Create inserts a new import tracker record
func (r *ImportTrackerRepositoryImpl) Create(ctx context.Context, db bun.IDB, tracker *model.ImportTracker) (*model.ImportTracker, error) {
	_, err := db.NewInsert().
		Model(tracker).
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to create import tracker")
		return nil, err
	}

	return tracker, nil
}

// Update updates the existing import tracker record
func (r *ImportTrackerRepositoryImpl) Update(ctx context.Context, db bun.IDB, tracker *model.ImportTracker) (*model.ImportTracker, error) {
	_, err := db.NewUpdate().
		Model(tracker).
		WherePK().
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Msg("Failed to update import tracker")
		return nil, err
	}

	return tracker, nil
}
