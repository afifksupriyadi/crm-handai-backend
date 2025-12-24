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

	// Generate Weekly Forecasts (7 days)
	weeklyForecasts, err := s.generateWeeklyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate weekly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, weeklyForecasts...)
	}

	// Generate Monthly Forecasts (4-5 weeks)
	monthlyForecasts, err := s.generateMonthlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate monthly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, monthlyForecasts...)
	}

	// Generate Yearly Forecasts (12 months)
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

// generateWeeklyForecasts creates 7 daily forecasts for CURRENT WEEK
func (s *SalesForecastServiceImpl) generateWeeklyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Calculate start of current week (Monday)
	weekday := int(endDate.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	currentWeekStart := endDate.AddDate(0, 0, -(weekday - 1))

	// Collect revenues for last 3 weeks (same weekday)
	var forecasts []*model.SalesForecast

	for day := 0; day < 7; day++ {
		forecastDate := currentWeekStart.AddDate(0, 0, day)

		// Get revenue for this weekday from last 3 weeks
		var revenues []float64
		for i := 1; i <= 3; i++ {
			historicalDate := forecastDate.AddDate(0, 0, -7*i)
			dayStart := time.Date(historicalDate.Year(), historicalDate.Month(), historicalDate.Day(), 0, 0, 0, 0, historicalDate.Location())
			dayEnd := dayStart.AddDate(0, 0, 1)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, dayStart, dayEnd)
			if err != nil {
				log.Warn().Err(err).Int("day", day).Msg("Failed to get historical revenue for day")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			log.Warn().Int("day", day).Msg("Insufficient data for day forecast")
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecast := &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodWeekly,
			ForecastDate:       forecastDate,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		}
		forecasts = append(forecasts, forecast)

		log.Debug().
			Str("date", forecastDate.Format("2006-01-02")).
			Str("day", forecastDate.Format("Monday")).
			Float64("min", min).
			Float64("avg", avg).
			Float64("max", max).
			Msg("Daily forecast generated")
	}

	return forecasts, nil
}

// generateMonthlyForecasts creates 4-5 weekly forecasts for CURRENT MONTH
func (s *SalesForecastServiceImpl) generateMonthlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Get first day of current month
	currentMonthStart := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location())
	currentMonthEnd := currentMonthStart.AddDate(0, 1, 0)

	var forecasts []*model.SalesForecast
	weekNum := 1

	// Loop through each week in current month
	for weekStart := currentMonthStart; weekStart.Before(currentMonthEnd); weekStart = weekStart.AddDate(0, 0, 7) {
		weekEnd := weekStart.AddDate(0, 0, 7)
		if weekEnd.After(currentMonthEnd) {
			weekEnd = currentMonthEnd
		}

		// Get revenue for this week from last 3 months (same week position)
		var revenues []float64
		for i := 1; i <= 3; i++ {
			historicalWeekStart := weekStart.AddDate(0, -i, 0)
			historicalWeekEnd := weekEnd.AddDate(0, -i, 0)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, historicalWeekStart, historicalWeekEnd)
			if err != nil {
				log.Warn().Err(err).Int("week", weekNum).Msg("Failed to get historical revenue for week")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			log.Warn().Int("week", weekNum).Msg("Insufficient data for week forecast")
			weekNum++
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecast := &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodMonthly,
			ForecastDate:       weekStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		}
		forecasts = append(forecasts, forecast)

		log.Debug().
			Str("week_start", weekStart.Format("2006-01-02")).
			Int("week_num", weekNum).
			Float64("min", min).
			Float64("avg", avg).
			Float64("max", max).
			Msg("Weekly forecast for month generated")

		weekNum++
	}

	return forecasts, nil
}

// generateYearlyForecasts creates 12 monthly forecasts for CURRENT YEAR
func (s *SalesForecastServiceImpl) generateYearlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	var forecasts []*model.SalesForecast

	// Loop through each month in current year
	for month := 1; month <= 12; month++ {
		monthStart := time.Date(endDate.Year(), time.Month(month), 1, 0, 0, 0, 0, endDate.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)

		// Get revenue for this month from last 3 years
		var revenues []float64
		for i := 1; i <= 3; i++ {
			historicalMonthStart := monthStart.AddDate(-i, 0, 0)
			historicalMonthEnd := monthEnd.AddDate(-i, 0, 0)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, historicalMonthStart, historicalMonthEnd)
			if err != nil {
				log.Warn().Err(err).Int("month", month).Msg("Failed to get historical revenue for month")
				continue
			}
			revenues = append(revenues, revenue)
		}

		if len(revenues) < 3 {
			log.Warn().Int("month", month).Msg("Insufficient data for month forecast")
			continue
		}

		min, avg, max := calculateStats(revenues)

		forecast := &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodYearly,
			ForecastDate:       monthStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		}
		forecasts = append(forecasts, forecast)

		log.Debug().
			Str("month", monthStart.Format("January 2006")).
			Float64("min", min).
			Float64("avg", avg).
			Float64("max", max).
			Msg("Monthly forecast for year generated")
	}

	return forecasts, nil
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

	// Convert to response format with detailed date info
	var responses []*model.SalesForecastResponse
	for _, f := range forecasts {
		periodLabel, dateStr, dateRange := formatPeriodWithDetails(f.ForecastDate, period)

		responses = append(responses, &model.SalesForecastResponse{
			Period:         periodLabel,
			Date:           dateStr,
			DateRange:      dateRange,
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

// Helper: formatPeriodWithDetails returns (periodLabel, dateStr, dateRange)
func formatPeriodWithDetails(date time.Time, period model.ForecastPeriod) (string, string, string) {
	// Indonesian locale
	daysIndo := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	monthsIndo := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	monthsIndoFull := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}

	switch period {
	case model.ForecastPeriodWeekly:
		// WEEKLY: "Senin", "2025-12-22", "22 Des 2025"
		dayName := daysIndo[date.Weekday()]
		dateStr := date.Format("2006-01-02")
		dateRange := fmt.Sprintf("%d %s %d", date.Day(), monthsIndo[date.Month()-1], date.Year())
		return dayName, dateStr, dateRange

	case model.ForecastPeriodMonthly:
		// MONTHLY: "Week 1", "2025-12-01", "1-7 Des 2025"
		weekNum := weekOfMonth(date)
		weekEnd := date.AddDate(0, 0, 6)

		// Pastikan weekEnd tidak melewati akhir bulan
		if weekEnd.Month() != date.Month() {
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
			weekEnd = lastDay
		}

		periodLabel := fmt.Sprintf("Week %d", weekNum)
		dateStr := date.Format("2006-01-02")
		dateRange := fmt.Sprintf("%d-%d %s %d", date.Day(), weekEnd.Day(), monthsIndo[date.Month()-1], date.Year())
		return periodLabel, dateStr, dateRange

	case model.ForecastPeriodYearly:
		// YEARLY: "Januari", "2025-01-01", "Januari 2025"
		monthName := monthsIndoFull[date.Month()-1]
		dateStr := date.Format("2006-01-02")
		dateRange := fmt.Sprintf("%s %d", monthName, date.Year())
		return monthName, dateStr, dateRange
	}

	return date.Format("2006-01-02"), date.Format("2006-01-02"), date.Format("2006-01-02")
}

// Helper: formatPeriod (keep for backward compatibility)
func formatPeriod(date time.Time, period model.ForecastPeriod) string {
	label, _, _ := formatPeriodWithDetails(date, period)
	return label
}

// Helper: weekOfMonth calculates week number in month (1-based)
func weekOfMonth(date time.Time) int {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())

	// Adjust for Monday as first day of week
	firstWeekday := int(firstDay.Weekday())
	if firstWeekday == 0 { // Sunday
		firstWeekday = 7
	}

	currentDay := date.Day()
	weekNum := ((currentDay + firstWeekday - 2) / 7) + 1

	return weekNum
}
