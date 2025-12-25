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
		Msg("Starting forecast generation for all periods")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to start transaction")
	}
	defer tx.Rollback()

	var allForecasts []*model.SalesForecast

	// Generate DAILY forecasts for entire month
	dailyForecasts, err := s.generateDailyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate daily forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, dailyForecasts...)
	}

	// Generate WEEKLY forecasts for entire month
	weeklyForecasts, err := s.generateWeeklyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate weekly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, weeklyForecasts...)
	}

	// Generate MONTHLY forecasts for entire year
	monthlyForecasts, err := s.generateMonthlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate monthly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, monthlyForecasts...)
	}

	// Generate YEARLY forecasts
	yearlyForecasts, err := s.generateYearlyForecasts(ctx, endDate, transactionBatchID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate yearly forecasts (non-critical)")
	} else {
		allForecasts = append(allForecasts, yearlyForecasts...)
	}

	// Bulk insert all forecasts (with UPSERT)
	if len(allForecasts) > 0 {
		if err := s.forecastRepo.BulkCreate(ctx, &tx, allForecasts); err != nil {
			return nil, response.WrapAppError(ctx, err, response.ErrForecastCalculation, "Failed to save forecasts")
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to commit forecasts")
	}

	resp := &model.GenerateForecastsResponse{
		DailyForecasts:   len(dailyForecasts),
		WeeklyForecasts:  len(weeklyForecasts),
		MonthlyForecasts: len(monthlyForecasts),
		YearlyForecasts:  len(yearlyForecasts),
		TotalGenerated:   len(allForecasts),
	}

	log.Info().
		Int("daily", resp.DailyForecasts).
		Int("weekly", resp.WeeklyForecasts).
		Int("monthly", resp.MonthlyForecasts).
		Int("yearly", resp.YearlyForecasts).
		Int("total", resp.TotalGenerated).
		Msg("Forecast generation completed")

	return resp, nil
}

// generateDailyForecasts: Generate forecast untuk setiap hari dalam bulan
// ✅ FIX: Generate untuk semua hari, meskipun data historis tidak cukup (nilai 0)
func (s *SalesForecastServiceImpl) generateDailyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	// Get first and last day of the month
	monthStart := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).AddDate(0, 0, -1)

	var forecasts []*model.SalesForecast

	// Generate forecast untuk setiap hari dalam bulan
	for day := monthStart; !day.After(monthEnd); day = day.AddDate(0, 0, 1) {
		// Ambil revenue dari 3 hari yang sama di 3 bulan sebelumnya
		var revenues []float64
		for i := 1; i <= 3; i++ {
			historicalDate := day.AddDate(0, -i, 0)
			dayStart := time.Date(historicalDate.Year(), historicalDate.Month(), historicalDate.Day(), 0, 0, 0, 0, time.UTC)
			dayEnd := dayStart.AddDate(0, 0, 1)

			revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, dayStart, dayEnd)
			if err != nil {
				log.Warn().Err(err).Str("date", day.Format("2006-01-02")).Msg("Failed to get revenue")
				continue
			}
			revenues = append(revenues, revenue)
		}

		// ✅ PERUBAHAN: Tetap generate meskipun data historis < 3 (nilai 0)
		var min, avg, max float64
		if len(revenues) >= 3 {
			min, avg, max = calculateStats(revenues)
		}
		// Jika < 3, biarkan nilai 0 (default)

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodDaily,
			ForecastDate:       day,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	log.Info().Int("count", len(forecasts)).Str("month", monthStart.Format("Jan 2006")).Msg("Daily forecasts generated")
	return forecasts, nil
}

// generateWeeklyForecasts: Generate forecast untuk setiap minggu dalam bulan
// ✅ FIX: Generate untuk semua minggu, meskipun data historis tidak cukup (nilai 0)
func (s *SalesForecastServiceImpl) generateWeeklyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	monthStart := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)

	var forecasts []*model.SalesForecast

	// Generate untuk setiap minggu dalam bulan
	for weekStart := monthStart; weekStart.Before(monthEnd); weekStart = weekStart.AddDate(0, 0, 7) {
		weekEnd := weekStart.AddDate(0, 0, 7)
		if weekEnd.After(monthEnd) {
			weekEnd = monthEnd
		}

		// Ambil revenue dari 3 minggu yang sama di 3 bulan sebelumnya
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

		// ✅ PERUBAHAN: Tetap generate meskipun data historis < 3 (nilai 0)
		var min, avg, max float64
		if len(revenues) >= 3 {
			min, avg, max = calculateStats(revenues)
		}

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodWeekly,
			ForecastDate:       weekStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	log.Info().Int("count", len(forecasts)).Str("month", monthStart.Format("Jan 2006")).Msg("Weekly forecasts generated")
	return forecasts, nil
}

// generateMonthlyForecasts: Generate forecast untuk SEMUA bulan dalam tahun (Januari-Desember)
// ✅ FIX: Generate untuk semua 12 bulan, meskipun data historis tidak cukup (nilai 0)
func (s *SalesForecastServiceImpl) generateMonthlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	var forecasts []*model.SalesForecast

	// ✅ PERUBAHAN: Generate untuk SEMUA 12 bulan (Januari - Desember)
	for month := 1; month <= 12; month++ {
		monthStart := time.Date(endDate.Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		// Ambil revenue dari 3 bulan yang sama di 3 tahun sebelumnya
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

		// ✅ PERUBAHAN: Tetap generate meskipun data historis < 3 (nilai 0)
		var min, avg, max float64
		if len(revenues) >= 3 {
			min, avg, max = calculateStats(revenues)
		}
		// Jika < 3, biarkan nilai 0 (default) - TETAP MASUKKAN KE FORECASTS

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodMonthly,
			ForecastDate:       monthStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	log.Info().Int("count", len(forecasts)).Int("year", endDate.Year()).Msg("Monthly forecasts generated (all 12 months)")
	return forecasts, nil
}

// generateYearlyForecasts: Generate forecast untuk tahun
// ✅ TETAP: Hanya generate jika ada data historis yang cukup
func (s *SalesForecastServiceImpl) generateYearlyForecasts(ctx context.Context, endDate time.Time, batchID int) ([]*model.SalesForecast, error) {
	log := logger.FromContext(ctx, 2)

	var forecasts []*model.SalesForecast

	yearStart := time.Date(endDate.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := yearStart.AddDate(1, 0, 0)

	// Ambil revenue dari 3 tahun sebelumnya
	var revenues []float64
	for i := 1; i <= 3; i++ {
		histStart := yearStart.AddDate(-i, 0, 0)
		histEnd := yearEnd.AddDate(-i, 0, 0)

		revenue, err := s.forecastRepo.GetHistoricalRevenue(ctx, histStart, histEnd)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get revenue")
			continue
		}
		revenues = append(revenues, revenue)
	}

	// ✅ TETAP: Hanya generate jika ada data historis yang cukup
	if len(revenues) >= 3 {
		min, avg, max := calculateStats(revenues)

		forecasts = append(forecasts, &model.SalesForecast{
			TransactionBatchID: batchID,
			ForecastPeriod:     model.ForecastPeriodYearly,
			ForecastDate:       yearStart,
			MinimumRevenue:     min,
			NormalRevenue:      avg,
			MaximumRevenue:     max,
		})
	}

	log.Info().Int("count", len(forecasts)).Int("year", endDate.Year()).Msg("Yearly forecasts generated")
	return forecasts, nil
}

// ========================================
// GET FORECASTS (API ENDPOINT)
// ========================================

func (s *SalesForecastServiceImpl) GetForecastsByPeriod(ctx context.Context, period model.ForecastPeriod, year, month, week int) ([]*model.SalesForecastResponse, error) {
	var startDate, endDate time.Time

	switch period {
	case model.ForecastPeriodDaily:
		// DAILY: setiap hari dalam bulan (atau minggu jika week diisi)
		monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		if week > 0 {
			// Filter by specific week
			startDate, endDate = getWeekRange(year, month, week)
		} else {
			// All days in month
			startDate = monthStart
			endDate = monthEnd
		}

	case model.ForecastPeriodWeekly:
		// WEEKLY: setiap minggu dalam bulan
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(0, 1, 0)

	case model.ForecastPeriodMonthly:
		// MONTHLY: setiap bulan dalam tahun
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = startDate.AddDate(1, 0, 0)

	case model.ForecastPeriodYearly:
		// YEARLY: tahun tertentu
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

	// ✅ PERUBAHAN: Hanya filter zero values untuk YEARLY
	var responses []*model.SalesForecastResponse
	for _, f := range forecasts {
		// Filter hanya untuk YEARLY (tampilkan yang ada data aja)
		if period == model.ForecastPeriodYearly {
			// Skip jika semua nilai 0
			if f.MinimumRevenue == 0 && f.NormalRevenue == 0 && f.MaximumRevenue == 0 {
				continue
			}
		}
		// Untuk DAILY, WEEKLY, MONTHLY: tampilkan semua (termasuk yang nilai 0)

		resp := convertToResponse(f, period)
		responses = append(responses, resp)
	}

	// ✅ PERUBAHAN: Hanya error jika YEARLY tidak ada data
	if len(responses) == 0 && period == model.ForecastPeriodYearly {
		return nil, response.WrapAppError(ctx, nil, response.ErrInsufficientData, "Data historis tidak cukup untuk membuat forecast di periode ini")
	}

	// Untuk period lain, return empty array jika tidak ada data (bukan error)
	if len(responses) == 0 {
		responses = []*model.SalesForecastResponse{}
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
	case model.ForecastPeriodDaily:
		// "Senin, 22 Des"
		resp.Period = fmt.Sprintf("%s, %d %s", daysIndo[f.ForecastDate.Weekday()], f.ForecastDate.Day(), monthsIndo[f.ForecastDate.Month()-1])
		resp.DateRange = fmt.Sprintf("%d %s %d", f.ForecastDate.Day(), monthsIndo[f.ForecastDate.Month()-1], f.ForecastDate.Year())

	case model.ForecastPeriodWeekly:
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

	case model.ForecastPeriodMonthly:
		monthNum := int(f.ForecastDate.Month())
		resp.Period = monthsIndoFull[monthNum-1]
		resp.MonthNumber = &monthNum
		resp.DateRange = fmt.Sprintf("%s %d", monthsIndoFull[monthNum-1], f.ForecastDate.Year())

	case model.ForecastPeriodYearly:
		resp.Period = fmt.Sprintf("%d", f.ForecastDate.Year())
		resp.DateRange = fmt.Sprintf("Tahun %d", f.ForecastDate.Year())
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
