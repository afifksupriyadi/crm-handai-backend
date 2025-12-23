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

	// ========== GET LAST 3 VALIDATED PREDICTIONS (WITH TX) ==========
	predictions, err := s.predictionRepo.GetByCustomerValidatedTx(ctx, db, customerID, 3) // ← USE TX
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get validated predictions")
	}

	// ========== COUNT TOTAL & CORRECT (WITH TX) ==========
	totalPredictions, err := s.predictionRepo.CountByCustomerTx(ctx, db, customerID) // ← USE TX
	if err != nil {
		log.Warn().Err(err).Msg("Failed to count total predictions, using 0")
		totalPredictions = 0
	}

	// Count ALL correct predictions
	allPredictions, err := s.predictionRepo.GetByCustomerValidatedTx(ctx, db, customerID, 100) // ← USE TX
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

	// ========== CALCULATE SEGMENT FROM LAST 3 ==========
	segment := s.calculateSegment(predictions)
	consecutiveCorrect := s.countConsecutiveCorrect(predictions)

	// ========== BUILD MODEL ==========
	segmentModel := &model.CustomerSegment{
		CustomerID:                    customerID,
		Segment:                       segment,
		ConsecutiveCorrectPredictions: consecutiveCorrect,
		TotalPredictions:              totalPredictions,
		TotalCorrectPredictions:       totalCorrect,
		LastUpdatedAt:                 time.Now(),
		UpdatedByBatchID:              transactionBatchID,
	}

	// ========== SAVE VIA REPOSITORY ==========
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

// calculateSegment applies segment rules based on LAST 3 predictions
func (s *SegmentDeterminerServiceImpl) calculateSegment(predictions []*model.CustomerPrediction) string {
	if len(predictions) == 0 {
		return model.SegmentRegular
	}

	if len(predictions) < 2 {
		return model.SegmentRegular
	}

	// Predictions already ordered by created_at DESC (most recent first)
	// Check CHURN: 2x FALSE consecutive (check positions 0,1)
	if len(predictions) >= 2 {
		if s.isFalse(predictions[0]) && s.isFalse(predictions[1]) {
			return model.SegmentChurn
		}
	}

	// Check LOYAL: 3x TRUE consecutive (check positions 0,1,2)
	if len(predictions) >= 3 {
		if s.isTrue(predictions[0]) && s.isTrue(predictions[1]) && s.isTrue(predictions[2]) {
			return model.SegmentLoyal
		}
	}

	// Everything else is REGULAR
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
