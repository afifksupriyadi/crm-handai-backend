package service

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type PredictionCalculatorServiceImpl struct {
	predictionRepo repository.CustomerPredictionRepository
}

// NewPredictionCalculatorService creates a new instance
func NewPredictionCalculatorService(predictionRepo repository.CustomerPredictionRepository) customer.PredictionCalculatorService {
	return &PredictionCalculatorServiceImpl{
		predictionRepo: predictionRepo,
	}
}

// CheckEligibility checks if customer is eligible for prediction
func (s *PredictionCalculatorServiceImpl) CheckEligibility(ctx context.Context, customerID int) (bool, error) {
	log := logger.FromContext(ctx, 2)

	uniqueDateCount, err := s.predictionRepo.CountUniqueTransactionDates(ctx, customerID)
	if err != nil {
		return false, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to check eligibility")
	}

	eligible := uniqueDateCount >= 2

	log.Debug().Int("customer_id", customerID).Int("unique_dates", uniqueDateCount).Bool("eligible", eligible).Msg("Checked eligibility")

	return eligible, nil
}

// CalculatePrediction calculates next purchase prediction for a customer
func (s *PredictionCalculatorServiceImpl) CalculatePrediction(ctx context.Context, customerID int, transactionBatchID int) (*model.CustomerPrediction, error) {
	log := logger.FromContext(ctx, 2)

	// Get last 10 transaction dates via repository
	dates, err := s.predictionRepo.GetCustomerTransactionDates(ctx, customerID, 10)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get transaction dates")
	}

	if len(dates) < 2 {
		return nil, fmt.Errorf("insufficient transactions for prediction (need 2, got %d)", len(dates))
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(dates)-1; i < j; i, j = i+1, j-1 {
		dates[i], dates[j] = dates[j], dates[i]
	}

	// Calculate intervals
	var intervals []float64
	for i := 1; i < len(dates); i++ {
		interval := dates[i].Sub(dates[i-1]).Hours() / 24
		intervals = append(intervals, interval)
	}

	// Average interval
	var sum float64
	for _, interval := range intervals {
		sum += interval
	}
	avgInterval := sum / float64(len(intervals))

	// Last transaction date
	lastTxDate := dates[len(dates)-1]

	// Predicted date
	predictedDate := lastTxDate.AddDate(0, 0, int(avgInterval))

	prediction := &model.CustomerPrediction{
		CustomerID:                customerID,
		TransactionBatchID:        transactionBatchID,
		LastTransactionDate:       lastTxDate,
		PredictedNextPurchaseDate: predictedDate,
		IsPredictedCorrect:        nil,
		CreatedAt:                 time.Now(),
	}

	log.Info().Int("customer_id", customerID).Float64("avg_interval_days", avgInterval).Str("predicted_date", predictedDate.Format("2006-01-02")).Msg("Prediction calculated")

	return prediction, nil
}
