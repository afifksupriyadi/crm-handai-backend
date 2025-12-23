package model

// SalesForecastResponse represents a single forecast data point
type SalesForecastResponse struct {
	Period         string   `json:"period"`          // "Jan", "Feb", "Week 1", "2025", etc
	MinimumRevenue float64  `json:"minimum_revenue"` // Rp
	NormalRevenue  float64  `json:"normal_revenue"`  // Rp (AVG)
	MaximumRevenue float64  `json:"maximum_revenue"` // Rp
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
