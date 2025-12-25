package service

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
)

type WindowCalculatorServiceImpl struct{}

// NewWindowCalculatorService creates a new instance
func NewWindowCalculatorService() customer.WindowCalculatorService {
	return &WindowCalculatorServiceImpl{}
}

// CalculateWindows calculates complete 7-day windows from import date range
// Returns: list of complete windows, new pending start date (or nil if complete)
func (s *WindowCalculatorServiceImpl) CalculateWindows(ctx context.Context, importStartDate, importEndDate time.Time, pendingStart *time.Time) ([]customer.Window, *time.Time, error) {
	log := logger.FromContext(ctx, 2)

	// Determine the actual start date (use pending if exists, otherwise import start)
	var actualStartDate time.Time
	if pendingStart != nil {
		actualStartDate = *pendingStart
		log.Info().
			Str("pending_start", pendingStart.Format("2006-01-02")).
			Str("import_start", importStartDate.Format("2006-01-02")).
			Msg("Using pending window start from previous import")
	} else {
		actualStartDate = importStartDate
		log.Info().
			Str("start", importStartDate.Format("2006-01-02")).
			Msg("No pending window, starting fresh")
	}

	var windows []customer.Window
	currentStart := actualStartDate

	// Calculate complete 7-day windows
	for {
		currentEnd := currentStart.AddDate(0, 0, 7).Add(-1 * time.Second)

		// Check if window is complete (end date <= import end date)
		if currentEnd.After(importEndDate) {
			// Incomplete window - this becomes the new pending
			log.Info().
				Str("pending_start", currentStart.Format("2006-01-02")).
				Str("import_end", importEndDate.Format("2006-01-02")).
				Int("pending_days", int(importEndDate.Sub(currentStart).Hours()/24)+1).
				Msg("Incomplete window found, setting as pending")
			return windows, &currentStart, nil
		}

		// Window is complete, add it
		windows = append(windows, customer.Window{
			StartDate: currentStart,
			EndDate:   currentEnd,
		})

		log.Debug().
			Str("window_start", currentStart.Format("2006-01-02")).
			Str("window_end", currentEnd.Format("2006-01-02")).
			Msg("Complete 7-day window calculated")

		// Move to next window
		currentStart = currentEnd.AddDate(0, 0, 1)

		// If next start exceeds import end, we're done
		if currentStart.After(importEndDate) {
			log.Info().
				Int("total_windows", len(windows)).
				Msg("All windows calculated, no pending")
			return windows, nil, nil
		}
	}
}
