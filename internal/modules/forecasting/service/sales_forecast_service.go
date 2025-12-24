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

// ========================================
// GENERATE FORECASTS (CALLED AFTER IMPORT)
// ========================================

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

	// Generate Weekly Forecasts for CURRENT WEEK
	weeklyForecasts, err := s.generateWeeklyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate weekly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, weeklyForecasts...)
	}

	// Generate Monthly Forecasts for CURRENT MONTH
	monthlyForecasts, err := s.generateMonthlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate monthly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, monthlyForecasts...)
	}

	// Generate Yearly Forecasts for CURRENT YEAR
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

// generateWeeklyForecasts: 7 days for CURRENT WEEK
func (s *SalesForecastServiceImpl) generateWeeklyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Get Monday of current week
	weekday := int(endDate.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := endDate.AddDate(0, 0, -(weekday - 1))

	var forecasts []*model.SalesForecast

	for day := 0; day < 7; day++ {
		forecastDate := monday.AddDate(0, 0, day)

		// Get revenue from same weekday in last 3 weeks
		var revenues []float64
		for i := 1; i <= 3; i++ {
			historicalDate := forecastDate.AddDate(0, 0, -7*i)
			dayStart := time.Date(historicalDate.Year(), historicalDate.Month(), historicalDate.Day(), 0, 0, 0, 0, time.UTC)
			dayEnd := dayStart.AddDate(0, 0, 1)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, dayStart, dayEnd)
			if err != nil {
				log.Warn().Err(err).Int("day", day).Msg("Failed to get revenue")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			log.Warn().Int("day", day).Msg("Insufficient data")
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodWeekly,
			ForecastDate:       forecastDate,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	return forecasts, nil
}

// generateMonthlyForecasts: 4-5 weeks for CURRENT MONTH
func (s *SalesForecastServiceImpl) generateMonthlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	monthStart := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var forecasts []*model.SalesForecast

	for weekStart := monthStart; weekStart.Before(monthEnd); weekStart = weekStart.AddDate(0, 0, 7) {
		weekEnd := weekStart.AddDate(0, 0, 7)
		if weekEnd.After(monthEnd) {
			weekEnd = monthEnd
		}

		// Get revenue from same week position in last 3 months
		var revenues []float64
		for i := 1; i <= 3; i++ {
			histStart := weekStart.AddDate(0, -i, 0)
			histEnd := weekEnd.AddDate(0, -i, 0)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, histStart, histEnd)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to get revenue")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodMonthly,
			ForecastDate:       weekStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	return forecasts, nil
}

// generateYearlyForecasts: 12 months for CURRENT YEAR
func (s *SalesForecastServiceImpl) generateYearlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	var forecasts []*model.SalesForecast

	for month := 1; month <= 12; month++ {
		monthStart := time.Date(endDate.Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		// Get revenue from same month in last 3 years
		var revenues []float64
		for i := 1; i <= 3; i++ {
			histStart := monthStart.AddDate(-i, 0, 0)
			histEnd := monthEnd.AddDate(-i, 0, 0)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, histStart, histEnd)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to get revenue")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodYearly,
			ForecastDate:       monthStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	return forecasts, nil
}

// ========================================
// GET FORECASTS (API ENDPOINT)
// ========================================

func (s *SalesForecastServiceImpl) GetForecastsByPeriod(ctx context.Context, period model.ForecastPeriod, year, month, week int) ([]*model.SalesForecastResponse, error) {
	var startDate, endDate time.Time

	switch period {
	case model.ForecastPeriodWeekly:
		// Calculate specific week in specific month
		startDate, endDate = getWeekRange(year, month, week)
	case model.ForecastPeriodMonthly:
		// Get all weeks in specific month
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0)
	case model.ForecastPeriodYearly:
		// Get all months in specific year
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(1, 0, 0)
	default:
		return nil, response.WrapAppError(ctx, nil, response.ErrInvalidPeriodType, "Invalid period type")
	}

	forecasts, err := s.forecastRepo.GetByPeriodAndDateRange(ctx, period, startDate, endDate)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get forecasts")
	}

	if len(forecasts) == 0 {
		return nil, response.WrapAppError(ctx, nil, response.ErrForecastNotFound, "No forecasts found")
	}

	// Convert to response
	var responses []*model.SalesForecastResponse
	for _, f := range forecasts {
		resp := convertToResponse(f, period)
		responses = append(responses, resp)
	}

	return responses, nil
}

// ========================================
// HELPER FUNCTIONS
// ========================================

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

func getWeekRange(year, month, weekNum int) (time.Time, time.Time) {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Adjust to Monday
	firstWeekday := int(monthStart.Weekday())
	if firstWeekday == 0 {
		firstWeekday = 7
	}
	firstMonday := monthStart.AddDate(0, 0, -(firstWeekday - 1))
	if firstMonday.Before(monthStart) {
		firstMonday = firstMonday.AddDate(0, 0, 7)
	}

	weekStart := firstMonday.AddDate(0, 0, 7*(weekNum-1))
	weekEnd := weekStart.AddDate(0, 0, 7)

	// Ensure we don't exceed month boundaries
	monthEnd := monthStart.AddDate(0, 1, 0)
	if weekEnd.After(monthEnd) {
		weekEnd = monthEnd
	}

	return weekStart, weekEnd
}

func convertToResponse(f *model.SalesForecast, period model.ForecastPeriod) *model.SalesForecastResponse {
	daysIndo := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	monthsIndo := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	monthsIndoFull := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	resp := &model.SalesForecastResponse{
		Date:           f.ForecastDate.Format("2006-01-02"),
		MinimumRevenue: f.MinimumRevenue,
		NormalRevenue:  f.NormalRevenue,
		MaximumRevenue: f.MaximumRevenue,
		ActualRevenue:  f.ActualRevenue,
	}

	switch period {
	case model.ForecastPeriodWeekly:
		resp.Period = daysIndo[f.ForecastDate.Weekday()]
		resp.DateRange = fmt.Sprintf("%d %s %d", f.ForecastDate.Day(), monthsIndo[f.ForecastDate.Month()-1], f.ForecastDate.Year())

	case model.ForecastPeriodMonthly:
		weekNum := weekOfMonth(f.ForecastDate)
		resp.Period = fmt.Sprintf("Week %d", weekNum)
		resp.WeekNumber = &weekNum
		resp.MonthNumber = intPtr(int(f.ForecastDate.Month()))

		weekEnd := f.ForecastDate.AddDate(0, 0, 6)
		if weekEnd.Month() != f.ForecastDate.Month() {
			lastDay := time.Date(f.ForecastDate.Year(), f.ForecastDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
			weekEnd = lastDay
		}
		resp.DateRange = fmt.Sprintf("%d-%d %s %d", f.ForecastDate.Day(), weekEnd.Day(), monthsIndo[f.ForecastDate.Month()-1], f.ForecastDate.Year())

	case model.ForecastPeriodYearly:
		monthNum := int(f.ForecastDate.Month())
		resp.Period = monthsIndoFull[monthNum-1]
		resp.MonthNumber = &monthNum
		resp.DateRange = fmt.Sprintf("%s %d", monthsIndoFull[monthNum-1], f.ForecastDate.Year())
	}

	return resp
}

func weekOfMonth(date time.Time) int {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	firstWeekday := int(firstDay.Weekday())
	if firstWeekday == 0 {
		firstWeekday = 7
	}
	currentDay := date.Day()
	weekNum := ((currentDay + firstWeekday - 2) / 7) + 1
	return weekNum
}

func intPtr(i int) *int {
	return &i
}
