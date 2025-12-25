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

// DetermineSegment determines customer segment based on prediction patterns
func (s *SegmentDeterminerServiceImpl) DetermineSegment(ctx context.Context, db bun.IDB, customerID int, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	// Get last 3 validated predictions (ordered by created_at DESC)
	predictions, err := s.predictionRepo.GetByCustomerValidatedTx(ctx, db, customerID, 3)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get validated predictions")
	}

	// Count total predictions (with transaction support)
	totalPredictions, err := s.predictionRepo.CountByCustomerTx(ctx, db, customerID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count total predictions, using 0")
		totalPredictions = 0
	}

	// Count ALL correct predictions
	allPredictions, err := s.predictionRepo.GetByCustomerValidatedTx(ctx, db, customerID, 100)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get all predictions")
		allPredictions = []*model.CustomerPrediction{}
	}

	totalCorrect := 0
	for _, p := range allPredictions {
		if p.IsPredictedCorrect != nil && *p.IsPredictedCorrect {
			totalCorrect++
		}
	}

	// Calculate segment from last 3 predictions
	segment := s.calculateSegment(predictions)
	consecutiveCorrect := s.countConsecutiveCorrect(predictions)

	// Build segment model
	segmentModel := &model.CustomerSegment{
		CustomerID:                    customerID,
		Segment:                       segment,
		ConsecutiveCorrectPredictions: consecutiveCorrect,
		TotalPredictions:              totalPredictions,
		TotalCorrectPredictions:       totalCorrect,
		LastUpdatedAt:                 time.Now(),
		UpdatedByBatchID:              transactionBatchID,
	}

	// Save via repository
	_, err = s.segmentRepo.Upsert(ctx, db, segmentModel)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update segment")
	}

	log.Info().
		Int("customer_id", customerID).
		Str("segment", segment).
		Int("total_predictions", totalPredictions).
		Int("total_correct", totalCorrect).
		Int("consecutive", consecutiveCorrect).
		Msg("Segment determined")

	return nil
}

// calculateSegment applies segment rules based on prediction patterns
// Predictions ordered by created_at DESC: [0]=newest, [1]=2nd, [2]=3rd
//
// Rules:
// - LOYAL: 3x TRUE consecutive (TTT)
// - CHURN: Latest prediction FALSE (except FFF case)
// - REGULAR: Latest prediction TRUE, or FFF (grace period reset)
func (s *SegmentDeterminerServiceImpl) calculateSegment(predictions []*model.CustomerPrediction) string {
	if len(predictions) == 0 {
		return model.SegmentRegular
	}

	// Rule 1: LOYAL - 3x TRUE consecutive (TTT)
	if len(predictions) >= 3 {
		if s.isTrue(predictions[0]) && s.isTrue(predictions[1]) && s.isTrue(predictions[2]) {
			return model.SegmentLoyal
		}
	}

	// Rule 2: Grace Period - 3x FALSE consecutive (FFF) → REGULAR (reset)
	if len(predictions) >= 3 {
		if s.isFalse(predictions[0]) && s.isFalse(predictions[1]) && s.isFalse(predictions[2]) {
			return model.SegmentRegular // Grace period: 3x churn, kasih kesempatan lagi
		}
	}

	// Rule 3: Check newest prediction
	if s.isFalse(predictions[0]) {
		return model.SegmentChurn // Latest FALSE = CHURN (at risk)
	}

	// Default: Latest TRUE or not enough data = REGULAR
	return model.SegmentRegular
}

// countConsecutiveCorrect counts consecutive TRUE from most recent
func (s *SegmentDeterminerServiceImpl) countConsecutiveCorrect(predictions []*model.CustomerPrediction) int {
	count := 0
	for _, p := range predictions {
		if s.isTrue(p) {
			count++
		} else {
			break // Stop at first non-TRUE
		}
	}
	return count
}

func (s *SegmentDeterminerServiceImpl) isTrue(p *model.CustomerPrediction) bool {
	return p.IsPredictedCorrect != nil && *p.IsPredictedCorrect
}

func (s *SegmentDeterminerServiceImpl) isFalse(p *model.CustomerPrediction) bool {
	return p.IsPredictedCorrect != nil && !*p.IsPredictedCorrect
}
