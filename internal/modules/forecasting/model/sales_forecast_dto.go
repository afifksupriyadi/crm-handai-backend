package model

import "fmt"

// SalesForecastResponse represents a single forecast data point
type SalesForecastResponse struct {
	Period         string   `json:"period"`                 // "Senin", "Week 1", "Januari"
	Date           string   `json:"date"`                   // ISO format: "2025-12-22"
	DateRange      string   `json:"date_range"`             // Human readable: "22 Des 2025" or "1-7 Des 2025"
	WeekNumber     *int     `json:"week_number,omitempty"`  // Week number in month (1-5)
	MonthNumber    *int     `json:"month_number,omitempty"` // Month number (1-12)
	MinimumRevenue float64  `json:"minimum_revenue"`        // Rp (minimum dari 3 periode sebelumnya)
	NormalRevenue  float64  `json:"normal_revenue"`         // Rp (rata-rata dari 3 periode sebelumnya)
	MaximumRevenue float64  `json:"maximum_revenue"`        // Rp (maksimum dari 3 periode sebelumnya)
	ActualRevenue  *float64 `json:"actual_revenue,omitempty"`
}

// GetSalesForecastRequest represents request for getting forecasts
type GetSalesForecastRequest struct {
	Period string `query:"period" validate:"required,oneof=WEEKLY MONTHLY YEARLY" doc:"Forecast period: WEEKLY, MONTHLY, YEARLY"`
	Year   int    `query:"year" validate:"required,min=2020,max=2100" doc:"Year (2020-2100)"`
	Month  int    `query:"month" validate:"omitempty,min=1,max=12" doc:"Month (1-12) - Required for WEEKLY & MONTHLY"`
	Week   int    `query:"week" validate:"omitempty,min=1,max=5" doc:"Week number in month (1-5) - Required for WEEKLY"`
}

// Validate validates the request based on period type
func (r *GetSalesForecastRequest) Validate() error {
	switch r.Period {
	case "WEEKLY":
		if r.Month == 0 {
			return fmt.Errorf("parameter 'month' wajib diisi untuk period WEEKLY")
		}
		if r.Week == 0 {
			return fmt.Errorf("parameter 'week' wajib diisi untuk period WEEKLY")
		}
		if r.Month < 1 || r.Month > 12 {
			return fmt.Errorf("parameter 'month' harus antara 1-12")
		}
		if r.Week < 1 || r.Week > 5 {
			return fmt.Errorf("parameter 'week' harus antara 1-5")
		}
	case "MONTHLY":
		if r.Month == 0 {
			return fmt.Errorf("parameter 'month' wajib diisi untuk period MONTHLY")
		}
		if r.Month < 1 || r.Month > 12 {
			return fmt.Errorf("parameter 'month' harus antara 1-12")
		}
	case "YEARLY":
		// No additional validation needed for YEARLY
	default:
		return fmt.Errorf("parameter 'period' tidak valid (harus WEEKLY, MONTHLY, atau YEARLY)")
	}
	return nil
}

// GenerateForecastsResponse represents the result of forecast generation
type GenerateForecastsResponse struct {
	WeeklyForecasts  int `json:"weekly_forecasts"`
	MonthlyForecasts int `json:"monthly_forecasts"`
	YearlyForecasts  int `json:"yearly_forecasts"`
	TotalGenerated   int `json:"total_generated"`
}
