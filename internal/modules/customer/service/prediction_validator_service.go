package service

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/uptrace/bun"
)

type PredictionValidatorServiceImpl struct {
	predictionRepo repository.CustomerPredictionRepository
}

// NewPredictionValidatorService creates a new instance
func NewPredictionValidatorService(predictionRepo repository.CustomerPredictionRepository) customer.PredictionValidatorService {
	return &PredictionValidatorServiceImpl{
		predictionRepo: predictionRepo,
	}
}

// ValidatePrediction validates a prediction against actual customer behavior
// isPredictedCorrect = TRUE if actual_date <= predicted_date
// isPredictedCorrect = FALSE if actual_date > predicted_date OR no transaction
func (s *PredictionValidatorServiceImpl) ValidatePrediction(ctx context.Context, db bun.IDB, prediction *model.CustomerPrediction, windowEndDate time.Time) error {
	log := logger.FromContext(ctx, 2)

	// Check if customer has transaction after last_transaction_date up to windowEndDate
	hasTransaction, actualDate, err := s.predictionRepo.CheckCustomerHasTransactionAfter(
		ctx,
		db,
		prediction.CustomerID,
		prediction.LastTransactionDate,
		windowEndDate,
	)

	if err != nil {
		return err
	}

	now := time.Now()

	if !hasTransaction {
		// No transaction found = FALSE
		prediction.IsPredictedCorrect = boolPtr(false)
		prediction.ActualNextPurchaseDate = nil
		prediction.ValidatedAt = &now

		log.Info().
			Int("prediction_id", prediction.ID).
			Int("customer_id", prediction.CustomerID).
			Str("predicted_date", prediction.PredictedNextPurchaseDate.Format("2006-01-02")).
			Bool("result", false).
			Msg("Prediction validated: no transaction found")
	} else {
		// Transaction found, check if actual <= predicted
		isCorrect := !actualDate.After(prediction.PredictedNextPurchaseDate)

		prediction.IsPredictedCorrect = &isCorrect
		prediction.ActualNextPurchaseDate = actualDate
		prediction.ValidatedAt = &now

		log.Info().
			Int("prediction_id", prediction.ID).
			Int("customer_id", prediction.CustomerID).
			Str("predicted_date", prediction.PredictedNextPurchaseDate.Format("2006-01-02")).
			Str("actual_date", actualDate.Format("2006-01-02")).
			Bool("result", isCorrect).
			Msg("Prediction validated")
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
