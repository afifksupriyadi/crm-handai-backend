package service

import (
	"context"
	"fmt"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/model"
	"github.com/afifksupriyadi/crm-handai-backend/internal/modules/analytics/repository"
	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

type AnalyticsServiceImpl struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) analytics.AnalyticsService {
	return &AnalyticsServiceImpl{
		analyticsRepo: analyticsRepo,
	}
}

func (s *AnalyticsServiceImpl) GetSalesChart(ctx context.Context, req *model.SalesChartRequest) (*model.SalesChartResponse, error) {
	// Get date range
	startDate, endDate, err := req.GetDateRange()
	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrInvalidDateRange, "Failed to calculate date range")
	}

	// Get sales data based on period type
	var dataPoints []*model.SalesDataPoint
	switch req.PeriodType {
	case "DAILY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByDaily(ctx, startDate, endDate)
	case "MONTHLY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByMonthly(ctx, startDate, endDate)
	case "YEARLY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByYearly(ctx, startDate, endDate)
	default:
		return nil, response.WrapAppError(ctx, nil, response.ErrInvalidPeriodType, "Invalid period type")
	}

	if err != nil {
		return nil, response.WrapAppError(ctx, err, response.ErrDatabaseError, "Failed to get sales data")
	}

	// Fill missing periods with zero values
	dataPoints = s.fillMissingPeriods(dataPoints, startDate, endDate, req.PeriodType)

	// Calculate total revenue
	totalRevenue := 0.0
	for _, dp := range dataPoints {
		totalRevenue += dp.Revenue
	}

	// Build chart data
	chartData := make([]model.ChartDataPoint, 0, len(dataPoints))
	for _, dp := range dataPoints {
		chartData = append(chartData, model.ChartDataPoint{
			Period:         dp.Period,
			Value:          dp.Revenue,
			FormattedValue: s.formatCurrency(dp.Revenue),
		})
	}

	resp := &model.SalesChartResponse{
		PeriodType:   req.PeriodType,
		StartDate:    startDate.Format("2006-01-02"),
		EndDate:      endDate.Format("2006-01-02"),
		TotalRevenue: totalRevenue,
		Growth:       0,
		Currency:     "Rp",
		ChartData:    chartData,
	}

	// Get comparison data if requested
	if req.Compare {
		prevStart, prevEnd := req.GetPreviousPeriodRange(startDate, endDate)
		comparison, err := s.getComparisonData(ctx, prevStart, prevEnd, req.PeriodType, totalRevenue)
		if err == nil {
			resp.Comparison = comparison
			resp.Growth = comparison.GrowthRate
		}
	}

	return resp, nil
}

func (s *AnalyticsServiceImpl) getComparisonData(ctx context.Context, prevStart, prevEnd time.Time, periodType string, currentRevenue float64) (*model.ComparisonData, error) {
	var dataPoints []*model.SalesDataPoint
	var err error

	switch periodType {
	case "DAILY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByDaily(ctx, prevStart, prevEnd)
	case "MONTHLY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByMonthly(ctx, prevStart, prevEnd)
	case "YEARLY":
		dataPoints, err = s.analyticsRepo.GetSalesDataByYearly(ctx, prevStart, prevEnd)
	}

	if err != nil {
		return nil, err
	}

	previousRevenue := 0.0
	for _, dp := range dataPoints {
		previousRevenue += dp.Revenue
	}

	difference := currentRevenue - previousRevenue
	growthRate := 0.0
	if previousRevenue > 0 {
		growthRate = (difference / previousRevenue) * 100
	}

	return &model.ComparisonData{
		PreviousRevenue: previousRevenue,
		Difference:      difference,
		GrowthRate:      growthRate / 100, // Convert to decimal
	}, nil
}

func (s *AnalyticsServiceImpl) fillMissingPeriods(dataPoints []*model.SalesDataPoint, start, end time.Time, periodType string) []*model.SalesDataPoint {
	if len(dataPoints) == 0 {
		return dataPoints
	}

	// Create map for quick lookup
	dataMap := make(map[string]*model.SalesDataPoint)
	for _, dp := range dataPoints {
		key := dp.Date.Format("2006-01-02")
		dataMap[key] = dp
	}

	result := make([]*model.SalesDataPoint, 0)
	current := start

	monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for current.Before(end) || current.Equal(end) {
		key := current.Format("2006-01-02")

		if dp, exists := dataMap[key]; exists {
			result = append(result, dp)
		} else {
			// Create zero value entry
			var period string
			var periodOrder int

			switch periodType {
			case "DAILY":
				period = current.Format("02 Jan")
				periodOrder = current.YearDay()
			case "MONTHLY":
				period = monthNames[current.Month()]
				periodOrder = int(current.Month())
			case "YEARLY":
				period = fmt.Sprintf("%d", current.Year())
				periodOrder = current.Year()
			}

			result = append(result, &model.SalesDataPoint{
				Period:      period,
				PeriodOrder: periodOrder,
				Date:        current,
				Revenue:     0,
			})
		}

		// Move to next period
		switch periodType {
		case "DAILY":
			current = current.AddDate(0, 0, 1)
		case "MONTHLY":
			current = current.AddDate(0, 1, 0)
		case "YEARLY":
			current = current.AddDate(1, 0, 0)
		}
	}

	return result
}

func (s *AnalyticsServiceImpl) formatCurrency(amount float64) string {
	// Simple Indonesian Rupiah formatting
	return fmt.Sprintf("Rp. %,.0f", amount)
}
