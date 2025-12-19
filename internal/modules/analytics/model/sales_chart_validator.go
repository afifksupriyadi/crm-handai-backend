package model

import (
	"context"
	"time"

	"github.com/afifksupriyadi/crm-handai-backend/internal/util/response"
)

var validPeriodTypes = map[string]bool{
	"DAILY":   true,
	"MONTHLY": true,
	"YEARLY":  true,
}

var validPresets = map[string]bool{
	"7d":  true,
	"30d": true,
	"3m":  true,
	"6m":  true,
	"12m": true,
	"ytd": true,
}

// Validate validates the sales chart request
func (r *SalesChartRequest) Validate(ctx context.Context) error {
	// Validate period type
	if r.PeriodType != "" && !validPeriodTypes[r.PeriodType] {
		return response.WrapAppError(ctx, nil, response.ErrInvalidPeriodType, "Invalid period type")
	}

	// Set default if empty
	if r.PeriodType == "" {
		r.PeriodType = "MONTHLY"
	}

	// Validate preset if provided
	if r.Preset != "" && !validPresets[r.Preset] {
		return response.WrapAppError(ctx, nil, response.ErrInvalidPreset, "Invalid preset")
	}

	// If custom dates provided, validate them
	if r.StartDate != "" || r.EndDate != "" {
		if r.StartDate == "" || r.EndDate == "" {
			return response.WrapAppError(ctx, nil, response.ErrInvalidDateRange, "Both start_date and end_date are required")
		}

		startDate, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			return response.WrapAppError(ctx, err, response.ErrInvalidDateFormat, "Invalid start_date format")
		}

		endDate, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			return response.WrapAppError(ctx, err, response.ErrInvalidDateFormat, "Invalid end_date format")
		}

		if endDate.Before(startDate) {
			return response.WrapAppError(ctx, nil, response.ErrInvalidDateRange, "end_date must be after start_date")
		}

		// Check max range for DAILY
		if r.PeriodType == "DAILY" {
			daysDiff := endDate.Sub(startDate).Hours() / 24
			if daysDiff > 365 {
				return response.WrapAppError(ctx, nil, response.ErrDateRangeTooLarge, "Maximum range for DAILY is 365 days")
			}
		}
	}

	return nil
}

// GetDateRange calculates the actual date range based on preset or custom dates
func (r *SalesChartRequest) GetDateRange() (time.Time, time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	// If custom dates provided, use them
	if r.StartDate != "" && r.EndDate != "" {
		start, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err := time.Parse("2006-01-02", r.EndDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		return start, end, nil
	}

	// Use preset
	var start, end time.Time
	end = now

	switch r.Preset {
	case "7d":
		start = now.AddDate(0, 0, -7)
	case "30d":
		start = now.AddDate(0, 0, -30)
	case "3m":
		start = now.AddDate(0, -3, 0)
	case "6m":
		start = now.AddDate(0, -6, 0)
	case "12m":
		start = now.AddDate(0, -12, 0)
	case "ytd":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
	default:
		// Default: last 12 months for MONTHLY, last 30 days for DAILY, last 3 years for YEARLY
		switch r.PeriodType {
		case "DAILY":
			start = now.AddDate(0, 0, -30)
		case "YEARLY":
			start = now.AddDate(-3, 0, 0)
		default: // MONTHLY
			start = now.AddDate(0, -12, 0)
		}
	}

	return start, end, nil
}

// GetPreviousPeriodRange calculates the previous period for comparison
func (r *SalesChartRequest) GetPreviousPeriodRange(start, end time.Time) (time.Time, time.Time) {
	duration := end.Sub(start)
	prevEnd := start.Add(-time.Second)
	prevStart := prevEnd.Add(-duration)
	return prevStart, prevEnd
}
