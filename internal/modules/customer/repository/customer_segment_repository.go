package repository

import (
	"context"
	"database/sql"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

// CustomerSegmentRepository defines the contract for customer segment data access
type CustomerSegmentRepository interface {
	GetByCustomer(ctx context.Context, customerID int) (*model.CustomerSegment, error)
	Upsert(ctx context.Context, db bun.IDB, segment *model.CustomerSegment) (*model.CustomerSegment, error)
	GetBySegment(ctx context.Context, segmentType string) ([]*model.CustomerSegment, error)
}

type CustomerSegmentRepositoryImpl struct {
	db *bun.DB
}

// NewCustomerSegmentRepository creates a new instance
func NewCustomerSegmentRepository(db *bun.DB) CustomerSegmentRepository {
	return &CustomerSegmentRepositoryImpl{db: db}
}

// GetByCustomer retrieves customer segment by customer ID
func (r *CustomerSegmentRepositoryImpl) GetByCustomer(ctx context.Context, customerID int) (*model.CustomerSegment, error) {
	segment := new(model.CustomerSegment)

	err := r.db.NewSelect().
		Model(segment).
		Where("customer_id = ?", customerID).
		Scan(ctx)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", customerID).Msg("Failed to get customer segment")
		return nil, err
	}

	return segment, nil
}

// Upsert inserts or updates customer segment
func (r *CustomerSegmentRepositoryImpl) Upsert(ctx context.Context, db bun.IDB, segment *model.CustomerSegment) (*model.CustomerSegment, error) {
	_, err := db.NewInsert().
		Model(segment).
		On("CONFLICT (customer_id) DO UPDATE").
		Set("segment = EXCLUDED.segment").
		Set("consecutive_correct_predictions = EXCLUDED.consecutive_correct_predictions").
		Set("total_predictions = EXCLUDED.total_predictions").
		Set("total_correct_predictions = EXCLUDED.total_correct_predictions").
		Set("last_updated_at = EXCLUDED.last_updated_at").
		Set("updated_by_batch_id = EXCLUDED.updated_by_batch_id").
		Exec(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Int("customer_id", segment.CustomerID).Msg("Failed to upsert customer segment")
		return nil, err
	}

	return segment, nil
}

// GetBySegment retrieves all customers with a specific segment
func (r *CustomerSegmentRepositoryImpl) GetBySegment(ctx context.Context, segmentType string) ([]*model.CustomerSegment, error) {
	var segments []*model.CustomerSegment

	err := r.db.NewSelect().
		Model(&segments).
		Where("segment = ?", segmentType).
		Order("last_updated_at DESC").
		Scan(ctx)

	if err != nil {
		logger.FromContext(ctx, 1).Error().Err(err).Str("segment", segmentType).Msg("Failed to get segments by type")
		return nil, err
	}

	return segments, nil
}
