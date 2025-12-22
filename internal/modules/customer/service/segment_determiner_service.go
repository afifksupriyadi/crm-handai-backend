package service

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type SegmentDeterminerServiceImpl struct {
	predictionRepo repository.CustomerPredictionRepository
	segmentRepo    repository.CustomerSegmentRepository
}

// NewSegmentDeterminerService creates a new instance
func NewSegmentDeterminerService(
	predictionRepo repository.CustomerPredictionRepository,
	segmentRepo repository.CustomerSegmentRepository,
) customer.SegmentDeterminerService {
	return &SegmentDeterminerServiceImpl{
		predictionRepo: predictionRepo,
		segmentRepo:    segmentRepo,
	}
}

// DetermineSegment determines customer segment based on last 3 validated predictions
func (s *SegmentDeterminerServiceImpl) DetermineSegment(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	// Get last 3 validated predictions
	predictions, err := s.predictionRepo.GetByCustomerValidated(ctx, customerID, 3)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get validated predictions")
	}

	// Calculate segment
	segment := s.calculateSegment(predictions)

	// Count stats
	totalCorrect := 0
	consecutiveCorrect := 0
	for _, p := range predictions {
		if p.IsPredictedCorrect != nil && *p.IsPredictedCorrect {
			totalCorrect++
			consecutiveCorrect++
		} else {
			consecutiveCorrect = 0
		}
	}

	// Upsert segment
	segmentModel := &model.CustomerSegment{
		CustomerID:                    customerID,
		Segment:                       segment,
		ConsecutiveCorrectPredictions: consecutiveCorrect,
		TotalPredictions:              len(predictions),
		TotalCorrectPredictions:       totalCorrect,
		LastUpdatedAt:                 time.Now(),
		UpdatedByBatchID:              &transactionBatchID,
	}

	_, err = s.segmentRepo.Upsert(ctx, db, segmentModel)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update segment")
	}

	log.Info().Int("customer_id", customerID).Str("segment", segment).Int("predictions_count", len(predictions)).Msg("Segment determined")

	return nil
}

// calculateSegment applies segment rules based on prediction history
func (s *SegmentDeterminerServiceImpl) calculateSegment(predictions []*model.CustomerPrediction) string {
	if len(predictions) == 0 {
		return model.SegmentRegular
	}

	// Reverse to get oldest first (for pattern checking)
	reversed := make([]*model.CustomerPrediction, len(predictions))
	for i, p := range predictions {
		reversed[len(predictions)-1-i] = p
	}

	// Check for LOYAL: 3x TRUE consecutive
	if len(reversed) >= 3 {
		if s.isTrue(reversed[0]) && s.isTrue(reversed[1]) && s.isTrue(reversed[2]) {
			return model.SegmentLoyal
		}
	}

	// Check for CHURN: 2x FALSE consecutive (at any position in last 3)
	if len(reversed) >= 2 {
		// Check positions: [0,1], [1,2]
		for i := 0; i < len(reversed)-1; i++ {
			if s.isFalse(reversed[i]) && s.isFalse(reversed[i+1]) {
				return model.SegmentChurn
			}
		}
	}

	// Everything else is REGULAR
	return model.SegmentRegular
}

func (s *SegmentDeterminerServiceImpl) isTrue(p *model.CustomerPrediction) bool {
	return p.IsPredictedCorrect != nil && *p.IsPredictedCorrect
}

func (s *SegmentDeterminerServiceImpl) isFalse(p *model.CustomerPrediction) bool {
	return p.IsPredictedCorrect != nil && !*p.IsPredictedCorrect
}
