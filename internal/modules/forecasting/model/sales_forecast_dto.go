package model

// SalesForecastResponse represents a single forecast data point
type SalesForecastResponse struct {
	Period         string   `json:"period"`          // "Senin", "Week 1", "Januari"
	Date           string   `json:"date"`            // ISO format: "2025-12-22"
	DateRange      string   `json:"date_range"`      // Human readable: "22 Des 2025" or "1-7 Des 2025"
	MinimumRevenue float64  `json:"minimum_revenue"` // Rp (minimum dari 3 periode sebelumnya)
	NormalRevenue  float64  `json:"normal_revenue"`  // Rp (rata-rata dari 3 periode sebelumnya)
	MaximumRevenue float64  `json:"maximum_revenue"` // Rp (maksimum dari 3 periode sebelumnya)
	ActualRevenue  *float64 `json:"actual_revenue,omitempty"`
}

// GetSalesForecastRequest represents request for getting forecasts
type GetSalesForecastRequest struct {
	Period string `query:"period" validate:"required,oneof=WEEKLY MONTHLY YEARLY"`
	Year   int    `query:"year" validate:"required,min=2020,max=2100"`
}

// GenerateForecastsResponse represents the result of forecast generation
type GenerateForecastsResponse struct {
	WeeklyForecasts  int `json:"weekly_forecasts"`
	MonthlyForecasts int `json:"monthly_forecasts"`
	YearlyForecasts  int `json:"yearly_forecasts"`
	TotalGenerated   int `json:"total_generated"`
}
