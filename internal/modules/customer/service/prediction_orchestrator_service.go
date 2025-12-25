package service

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/repository"
	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	importDataRepo "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type PredictionOrchestratorServiceImpl struct {
	db                             *bun.DB
	trackerRepo                    importDataRepo.ImportTrackerRepository
	predictionRepo                 repository.CustomerPredictionRepository
	predictedProductRepo           repository.CustomerPredictedProductRepository
	segmentRepo                    repository.CustomerSegmentRepository
	customerSvc                    customer.CustomerService
	windowCalculatorSvc            customer.WindowCalculatorService
	predictionCalculatorSvc        customer.PredictionCalculatorService
	productPredictionCalculatorSvc customer.ProductPredictionCalculatorService
	predictionValidatorSvc         customer.PredictionValidatorService
	segmentDeterminerSvc           customer.SegmentDeterminerService
}

// NewPredictionOrchestratorService creates a new instance
func NewPredictionOrchestratorService(
	db *bun.DB,
	trackerRepo importDataRepo.ImportTrackerRepository,
	predictionRepo repository.CustomerPredictionRepository,
	segmentRepo repository.CustomerSegmentRepository,
	predictedProductRepo repository.CustomerPredictedProductRepository,
	customerSvc customer.CustomerService,
	windowCalculatorSvc customer.WindowCalculatorService,
	predictionCalculatorSvc customer.PredictionCalculatorService,
	productPredictionCalculatorSvc customer.ProductPredictionCalculatorService,
	predictionValidatorSvc customer.PredictionValidatorService,
	segmentDeterminerSvc customer.SegmentDeterminerService,
) customer.PredictionOrchestratorService {
	return &PredictionOrchestratorServiceImpl{
		db:                             db,
		trackerRepo:                    trackerRepo,
		predictionRepo:                 predictionRepo,
		segmentRepo:                    segmentRepo,
		predictedProductRepo:           predictedProductRepo,
		customerSvc:                    customerSvc,
		windowCalculatorSvc:            windowCalculatorSvc,
		predictionCalculatorSvc:        predictionCalculatorSvc,
		productPredictionCalculatorSvc: productPredictionCalculatorSvc,
		predictionValidatorSvc:         predictionValidatorSvc,
		segmentDeterminerSvc:           segmentDeterminerSvc,
	}
}

// ProcessPredictions orchestrates the entire prediction process
func (s *PredictionOrchestratorServiceImpl) ProcessPredictions(ctx context.Context, importStartDate, importEndDate time.Time, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	log.Info().
		Str("import_start", importStartDate.Format("2006-01-02")).
		Str("import_end", importEndDate.Format("2006-01-02")).
		Int("batch_id", transactionBatchID).
		Msg("Starting prediction processing")

	// STEP 1: Get tracker and validate import sequence
	tracker, err := s.trackerRepo.GetLatest(ctx)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get import tracker")
	}

	// Define default/initial date (from migration)
	defaultDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

	// Check if this is first import (tracker still has default value)
	if tracker != nil && !tracker.LastImportEndDate.Equal(defaultDate) {
		// This is NOT first import, validate sequence
		expectedStart := tracker.LastImportEndDate.AddDate(0, 0, 1)
		if !importStartDate.Equal(expectedStart) {
			log.Error().
				Str("expected_start", expectedStart.Format("2006-01-02")).
				Str("got_start", importStartDate.Format("2006-01-02")).
				Str("last_import_end", tracker.LastImportEndDate.Format("2006-01-02")).
				Msg("Import sequence gap detected")

			return fmt.Errorf("import gap detected: expected start date %s, got %s",
				expectedStart.Format("2006-01-02"),
				importStartDate.Format("2006-01-02"))
		}

		log.Info().
			Str("last_import_end", tracker.LastImportEndDate.Format("2006-01-02")).
			Str("current_start", importStartDate.Format("2006-01-02")).
			Msg("Sequential validation passed")
	} else {
		log.Info().Msg("First prediction run detected, skipping sequential validation")
	}
	// STEP 2: Calculate windows
	var pendingStart *time.Time
	if tracker != nil {
		pendingStart = tracker.PendingWindowStart
	}

	windows, newPending, err := s.windowCalculatorSvc.CalculateWindows(ctx, importStartDate, importEndDate, pendingStart)
	if err != nil {
		return response.WrapAppError(ctx, err, response.ErrInternalServerError, "Failed to calculate windows")
	}

	log.Info().Int("complete_windows", len(windows)).Msg("Windows calculated")

	// STEP 3: Process each complete window
	for i, window := range windows {
		log.Info().
			Int("window_number", i+1).
			Str("start", window.StartDate.Format("2006-01-02")).
			Str("end", window.EndDate.Format("2006-01-02")).
			Msg("Processing window")

		err = s.processWindow(ctx, window, transactionBatchID)
		if err != nil {
			log.Error().Err(err).Int("window_number", i+1).Msg("Failed to process window")
			return err
		}
	}

	// STEP 4: Update tracker
	var lastWindowEndDate time.Time
	if len(windows) > 0 {
		lastWindowEndDate = windows[len(windows)-1].EndDate
	} else if tracker != nil {
		lastWindowEndDate = tracker.LastWindowEndDate
	} else {
		lastWindowEndDate = importStartDate.AddDate(0, 0, -1) // Default if no windows
	}

	err = s.updateTracker(ctx, importEndDate, lastWindowEndDate, newPending)
	if err != nil {
		return err
	}

	log.Info().Msg("Prediction processing completed successfully")
	return nil
}

func (s *PredictionOrchestratorServiceImpl) processWindow(ctx context.Context, window customer.Window, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// ========== STEP 0: COMPUTE METRICS ==========
		log.Debug().Msg("Step 0: Computing metrics")
		err := s.computeMetricsForWindow(ctx, window, transactionBatchID)
		if err != nil {
			return err
		}

		// ========== STEP A: VALIDATE OLD PREDICTIONS ==========
		log.Debug().Msg("Step A: Validating old predictions")
		validatedCustomers := make(map[int]bool)

		predictions, err := s.predictionRepo.GetPendingValidations(ctx, tx, window.EndDate)
		if err != nil {
			return err
		}

		for _, prediction := range predictions {
			err = s.predictionValidatorSvc.ValidatePrediction(ctx, tx, prediction, window.EndDate)
			if err != nil {
				return err
			}

			_, err = s.predictionRepo.Update(ctx, tx, prediction)
			if err != nil {
				return err
			}

			validatedCustomers[prediction.CustomerID] = true
		}

		log.Info().Int("validated_count", len(predictions)).Msg("Old predictions validated")

		// ========== STEP B: UPDATE SEGMENTS FOR OLD VALIDATIONS ==========
		log.Debug().Msg("Step B: Updating segments for old validations")
		err = s.updateSegmentsForValidated(ctx, tx, window, transactionBatchID, validatedCustomers)
		if err != nil {
			return err
		}

		// ========== STEP C: GENERATE NEW PREDICTIONS ==========
		log.Debug().Msg("Step C: Generating new predictions")
		immediateValidations, err := s.generateNewPredictions(ctx, tx, window, transactionBatchID)
		if err != nil {
			return err
		}

		// ========== STEP D: UPDATE SEGMENTS FOR IMMEDIATE VALIDATIONS ========== ✅ NEW!
		if len(immediateValidations) > 0 {
			log.Debug().
				Int("immediate_validations", len(immediateValidations)).
				Msg("Step D: Updating segments for immediate validations")
			err = s.updateSegmentsForValidated(ctx, tx, window, transactionBatchID, immediateValidations)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Window processing failed")
		return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to process window")
	}

	return nil
}

// computeMetricsForWindow computes customer_metrics for all customers in window
func (s *PredictionOrchestratorServiceImpl) computeMetricsForWindow(ctx context.Context, window customer.Window, transactionBatchID int) error {
	log := logger.FromContext(ctx, 2)

	// Get customers with transactions in this window
	customerIDs, err := s.predictionRepo.GetCustomerIDsWithTransactionsInWindow(ctx, s.db, window.StartDate, window.EndDate)
	if err != nil {
		return err
	}

	log.Info().Int("customers", len(customerIDs)).Msg("Computing metrics for customers in window")

	// Compute metrics for each customer
	for _, customerID := range customerIDs {
		err := s.customerSvc.ComputeCustomerMetrics(ctx, customerID, transactionBatchID)
		if err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to compute metrics")
			// Don't fail entire process
			continue
		}
	}

	log.Info().Int("computed", len(customerIDs)).Msg("Metrics computed successfully")
	return nil
}

// validateOldPredictions validates all pending predictions that fall within this window
func (s *PredictionOrchestratorServiceImpl) validateOldPredictions(ctx context.Context, tx bun.Tx, window customer.Window) error {
	// Get all predictions with NULL status and predicted_date <= window.EndDate
	predictions, err := s.predictionRepo.GetPendingValidations(ctx, tx, window.EndDate)
	if err != nil {
		return err
	}

	for _, prediction := range predictions {
		err = s.predictionValidatorSvc.ValidatePrediction(ctx, tx, prediction, window.EndDate)
		if err != nil {
			return err
		}

		// Update prediction in database
		_, err = s.predictionRepo.Update(ctx, tx, prediction)
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx, 2).Info().Int("validated_count", len(predictions)).Msg("Predictions validated")
	return nil
}

// updateSegmentsForValidated updates segments for customers whose predictions were validated
func (s *PredictionOrchestratorServiceImpl) updateSegmentsForValidated(ctx context.Context, tx bun.Tx, window customer.Window, transactionBatchID int, validatedCustomers map[int]bool) error {
	log := logger.FromContext(ctx, 2)

	log.Info().Int("validated_customers", len(validatedCustomers)).Msg("Updating segments for validated customers")

	// Update segment for each customer
	for customerID := range validatedCustomers {
		err := s.segmentDeterminerSvc.DetermineSegment(ctx, tx, customerID, transactionBatchID)
		if err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to update segment")
			continue
		}
	}

	log.Info().Int("customers_segmented", len(validatedCustomers)).Msg("Segments updated")
	return nil
}

// generateNewPredictions generates predictions for customers in window
// Returns: map of customers with immediate validations
func (s *PredictionOrchestratorServiceImpl) generateNewPredictions(ctx context.Context, tx bun.Tx, window customer.Window, transactionBatchID int) (map[int]bool, error) {
	log := logger.FromContext(ctx, 2)

	// ✅ NEW: Track immediate validations
	immediateValidations := make(map[int]bool)

	customerIDs, err := s.predictionRepo.GetCustomerIDsWithTransactionsInWindow(ctx, tx, window.StartDate, window.EndDate)
	if err != nil {
		return immediateValidations, err
	}

	// Deduplication
	uniqueMap := make(map[int]bool)
	for _, id := range customerIDs {
		uniqueMap[id] = true
	}

	uniqueIDs := make([]int, 0, len(uniqueMap))
	for id := range uniqueMap {
		uniqueIDs = append(uniqueIDs, id)
	}

	log.Info().
		Int("raw_count", len(customerIDs)).
		Int("unique_count", len(uniqueIDs)).
		Str("window_start", window.StartDate.Format("2006-01-02")).
		Str("window_end", window.EndDate.Format("2006-01-02")).
		Msg("Processing predictions for window")

	generatedCount := 0
	immediateValidationCount := 0

	for _, customerID := range uniqueIDs {
		// Check eligibility
		eligible, err := s.predictionCalculatorSvc.CheckEligibility(ctx, customerID)
		if err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to check eligibility")
			continue
		}

		if !eligible {
			continue
		}

		// Calculate prediction (with window end date)
		prediction, err := s.predictionCalculatorSvc.CalculatePrediction(ctx, customerID, transactionBatchID, window.EndDate)
		if err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to calculate prediction")
			continue
		}

		// Set window info
		prediction.WindowStartDate = window.StartDate
		prediction.WindowEndDate = window.EndDate

		// Create prediction
		createdPrediction, err := s.predictionRepo.Create(ctx, tx, prediction)
		if err != nil {
			log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to create prediction")
			continue
		}

		generatedCount++

		// ========== CALCULATE PREDICTED PRODUCTS ==========
		predictedProducts, err := s.productPredictionCalculatorSvc.CalculateProductPredictions(ctx, customerID, createdPrediction.ID)
		if err != nil {
			log.Warn().Err(err).Int("prediction_id", createdPrediction.ID).Msg("Failed to calculate predicted products")
			// Don't fail the entire prediction, just continue
		} else if len(predictedProducts) > 0 {
			// Save predicted products to database using CreateBatch
			err := s.predictedProductRepo.CreateBatch(ctx, tx, predictedProducts)
			if err != nil {
				log.Warn().Err(err).Int("prediction_id", createdPrediction.ID).Msg("Failed to save predicted products")
			} else {
				log.Debug().Int("prediction_id", createdPrediction.ID).Int("products_count", len(predictedProducts)).Msg("Predicted products saved")
			}
		}

		// ========== IMMEDIATE VALIDATION ==========
		if createdPrediction.PredictedNextPurchaseDate.Before(window.EndDate) ||
			createdPrediction.PredictedNextPurchaseDate.Equal(window.EndDate) {

			hasTransaction, actualDate, err := s.predictionRepo.CheckCustomerHasTransactionAfter(
				ctx, tx, customerID,
				createdPrediction.LastTransactionDate,
				createdPrediction.PredictedNextPurchaseDate,
			)

			if err != nil {
				log.Warn().Err(err).Int("prediction_id", createdPrediction.ID).Msg("Failed to validate prediction")
				continue
			}

			isCorrect := hasTransaction
			now := time.Now()

			createdPrediction.IsPredictedCorrect = &isCorrect
			createdPrediction.ActualNextPurchaseDate = actualDate
			createdPrediction.ValidatedAt = &now

			_, err = s.predictionRepo.Update(ctx, tx, createdPrediction)
			if err != nil {
				log.Warn().Err(err).Int("prediction_id", createdPrediction.ID).Msg("Failed to update validation")
				continue
			}

			// ✅ NEW: TRACK IMMEDIATE VALIDATION
			immediateValidations[customerID] = true
			immediateValidationCount++

			log.Debug().
				Int("prediction_id", createdPrediction.ID).
				Int("customer_id", customerID).
				Bool("is_correct", isCorrect).
				Msg("Prediction immediately validated")
		}

		// Cleanup old predictions
		count, err := s.predictionRepo.CountByCustomerTx(ctx, tx, customerID)
		if err == nil && count > 4 {
			err = s.predictionRepo.DeleteOldest(ctx, tx, customerID)
			if err != nil {
				log.Warn().Err(err).Int("customer_id", customerID).Msg("Failed to delete oldest prediction")
			}
		}
	}

	log.Info().
		Int("generated", generatedCount).
		Int("immediate_validated", immediateValidationCount).
		Str("window", fmt.Sprintf("%s to %s", window.StartDate.Format("2006-01-02"), window.EndDate.Format("2006-01-02"))).
		Msg("New predictions generated for window")

	return immediateValidations, nil // ✅ RETURN immediate validations
}

// updateTracker updates the import tracker
func (s *PredictionOrchestratorServiceImpl) updateTracker(ctx context.Context, importEndDate, lastWindowEndDate time.Time, pendingStart *time.Time) error {
	tracker, err := s.trackerRepo.GetLatest(ctx)
	if err != nil {
		return err
	}

	if tracker == nil {
		// First import - create tracker
		tracker = &importDataModel.ImportTracker{
			LastImportEndDate:  importEndDate,
			LastWindowEndDate:  lastWindowEndDate,
			PendingWindowStart: pendingStart,
		}

		_, err = s.trackerRepo.Create(ctx, s.db, tracker)
		if err != nil {
			return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to create tracker")
		}
	} else {
		// Update existing tracker
		tracker.LastImportEndDate = importEndDate
		tracker.LastWindowEndDate = lastWindowEndDate
		tracker.PendingWindowStart = pendingStart
		tracker.UpdatedAt = time.Now()

		_, err = s.trackerRepo.Update(ctx, s.db, tracker)
		if err != nil {
			return response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to update tracker")
		}
	}

	logger.FromContext(ctx, 2).Info().
		Str("last_import_end", importEndDate.Format("2006-01-02")).
		Str("last_window_end", lastWindowEndDate.Format("2006-01-02")).
		Msg("Tracker updated")

	return nil
}
