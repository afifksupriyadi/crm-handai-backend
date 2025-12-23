package service

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/forecasting/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/logger"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
	"github.com/uptrace/bun"
)

type SalesForecastServiceImpl struct {
	db           *bun.DB
	forecastRepo repository.SalesForecastRepository
}

func NewSalesForecastService(
	db *bun.DB,
	forecastRepo repository.SalesForecastRepository,
) forecasting.SalesForecastService {
	return &SalesForecastServiceImpl{
		db:           db,
		forecastRepo: forecastRepo,
	}
}

// GenerateForecasts creates forecasts after import
func (s *SalesForecastServiceImpl) GenerateForecasts(ctx context.Context, endDate time.Time, transactionBatchID int) (*model.GenerateForecastsResponse, error) {
	log := logger.FromContext(ctx, 2)

	log.Info().
		Str("end_date", endDate.Format("2006-01-02")).
		Int("batch_id", transactionBatchID).
		Msg("Starting forecast generation")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	defer tx.Rollback()

	var allForecasts []*model.SalesForecast

	// Generate Weekly Forecasts
	weeklyForecasts, err := s.generateWeeklyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate weekly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, weeklyForecasts...)
	}

	// Generate Monthly Forecasts
	monthlyForecasts, err := s.generateMonthlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate monthly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, monthlyForecasts...)
	}

	// Generate Yearly Forecasts
	yearlyForecasts, err := s.generateYearlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate yearly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, yearlyForecasts...)
	}

	// Bulk insert all forecasts
	if len(allForecasts) > 0 {
		if err := s.forecastRepo.BulkCreate(ctx, &tx, allForecasts); err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrForecastCalculation, "Failed to save forecasts")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit forecasts")
	}

	resp := &model.GenerateForecastsResponse{
		WeeklyForecasts:  len(weeklyForecasts),
		MonthlyForecasts: len(monthlyForecasts),
		YearlyForecasts:  len(yearlyForecasts),
		TotalGenerated:   len(allForecasts),
	}

	log.Info().
		Int("weekly", resp.WeeklyForecasts).
		Int("monthly", resp.MonthlyForecasts).
		Int("yearly", resp.YearlyForecasts).
		Int("total", resp.TotalGenerated).
		Msg("Forecast generation completed")

	return resp, nil
}

// generateWeeklyForecasts creates forecasts for next week based on last 3 weeks
func (s *SalesForecastServiceImpl) generateWeeklyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Calculate revenue for last 3 weeks
	var revenues []float64
	for i := 1; i <= 3; i++ {
		weekStart := endDate.AddDate(0, 0, -7*i)
		weekEnd := endDate.AddDate(0, 0, -7*(i-1))

		revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, weekStart, weekEnd)
		if err != nil {
			return nil, err
		}
		revenues = append(revenues, revenue)
	}

	if len(revenues) < 3 {
		log.Warn().Msg("Insufficient data for weekly forecast")
		return nil, response.WrapAppError(ctx, nil, response.ErrInsufficientData, "Not enough weekly data")
	}

	// Calculate min, avg, max
	min, avg, max := calculateStats(revenues)

	// Create forecast for next week
	forecastDate := endDate.AddDate(0, 0, 7)
	forecast := &model.SalesForecast{
		TransactionBatchID: batchID,
		ForecastPeriod:     model.ForecastPeriodWeekly,
		ForecastDate:       forecastDate,
		MinimumRevenue:     min,
		NormalRevenue:      avg,
		MaximumRevenue:     max,
	}

	log.Debug().
		Str("forecast_date", forecastDate.Format("2006-01-02")).
		Float64("min", min).
		Float64("avg", avg).
		Float64("max", max).
		Msg("Weekly forecast generated")

	return []*model.SalesForecast{forecast}, nil
}

// generateMonthlyForecasts creates forecasts for next month based on last 3 months
func (s *SalesForecastServiceImpl) generateMonthlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Calculate revenue for last 3 months
	var revenues []float64
	for i := 1; i <= 3; i++ {
		monthStart := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location()).AddDate(0, -i, 0)
		monthEnd := monthStart.AddDate(0, 1, 0)

		revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, monthStart, monthEnd)
		if err != nil {
			return nil, err
		}
		revenues = append(revenues, revenue)
	}

	if len(revenues) < 3 {
		log.Warn().Msg("Insufficient data for monthly forecast")
		return nil, response.WrapAppError(ctx, nil, response.ErrInsufficientData, "Not enough monthly data")
	}

	// Calculate min, avg, max
	min, avg, max := calculateStats(revenues)

	// Create forecast for next month
	forecastDate := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location()).AddDate(0, 1, 0)
	forecast := &model.SalesForecast{
		TransactionBatchID: batchID,
		ForecastPeriod:     model.ForecastPeriodMonthly,
		ForecastDate:       forecastDate,
		MinimumRevenue:     min,
		NormalRevenue:      avg,
		MaximumRevenue:     max,
	}

	log.Debug().
		Str("forecast_date", forecastDate.Format("2006-01-02")).
		Float64("min", min).
		Float64("avg", avg).
		Float64("max", max).
		Msg("Monthly forecast generated")

	return []*model.SalesForecast{forecast}, nil
}

// generateYearlyForecasts creates forecasts for next year based on last 3 years
func (s *SalesForecastServiceImpl) generateYearlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Calculate revenue for last 3 years
	var revenues []float64
	for i := 1; i <= 3; i++ {
		yearStart := time.Date(endDate.Year()-i, 1, 1, 0, 0, 0, 0, endDate.Location())
		yearEnd := time.Date(endDate.Year()-i+1, 1, 1, 0, 0, 0, 0, endDate.Location())

		revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, yearStart, yearEnd)
		if err != nil {
			return nil, err
		}
		revenues = append(revenues, revenue)
	}

	if len(revenues) < 3 {
		log.Warn().Msg("Insufficient data for yearly forecast")
		return nil, response.WrapAppError(ctx, nil, response.ErrInsufficientData, "Not enough yearly data")
	}

	// Calculate min, avg, max
	min, avg, max := calculateStats(revenues)

	// Create forecast for next year
	forecastDate := time.Date(endDate.Year()+1, 1, 1, 0, 0, 0, 0, endDate.Location())
	forecast := &model.SalesForecast{
		TransactionBatchID: batchID,
		ForecastPeriod:     model.ForecastPeriodYearly,
		ForecastDate:       forecastDate,
		MinimumRevenue:     min,
		NormalRevenue:      avg,
		MaximumRevenue:     max,
	}

	log.Debug().
		Str("forecast_date", forecastDate.Format("2006-01-02")).
		Float64("min", min).
		Float64("avg", avg).
		Float64("max", max).
		Msg("Yearly forecast generated")

	return []*model.SalesForecast{forecast}, nil
}

// GetForecastsByPeriod retrieves forecasts for display
func (s *SalesForecastServiceImpl) GetForecastsByPeriod(ctx context.Context, period model.ForecastPeriod, year int) ([]*model.SalesForecastResponse, error) {
	forecasts, err := s.forecastRepo.GetByPeriodAndYear(ctx, period, year)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get forecasts")
	}

	if len(forecasts) == 0 {
		return nil, response.WrapAppError(ctx, nil, response.ErrForecastNotFound, "No forecasts found")
	}

	// Convert to response format
	var responses []*model.SalesForecastResponse
	for _, f := range forecasts {
		responses = append(responses, &model.SalesForecastResponse{
			Period:         formatPeriod(f.ForecastDate, period),
			MinimumRevenue: f.MinimumRevenue,
			NormalRevenue:  f.NormalRevenue,
			MaximumRevenue: f.MaximumRevenue,
			ActualRevenue:  f.ActualRevenue,
		})
	}

	return responses, nil
}

// Helper: calculateStats computes min, avg, max
func calculateStats(values []float64) (min, avg, max float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	min = values[0]
	max = values[0]
	sum := 0.0

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg = sum / float64(len(values))
	return min, avg, max
}

// Helper: formatPeriod converts date to readable period label
func formatPeriod(date time.Time, period model.ForecastPeriod) string {
	switch period {
	case model.ForecastPeriodWeekly:
		return fmt.Sprintf("Week %d", weekOfMonth(date))
	case model.ForecastPeriodMonthly:
		return date.Format("Jan 2006")
	case model.ForecastPeriodYearly:
		return date.Format("2006")
	}
	return date.Format("2006-01-02")
}

// Helper: weekOfMonth calculates week number in month
func weekOfMonth(date time.Time) int {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return (date.Day()+int(firstDay.Weekday())-1)/7 + 1
}
