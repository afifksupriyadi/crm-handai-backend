package model

import (
	"time"
)

// SalesChartRequest represents the request for sales chart data
type SalesChartRequest struct {
	PeriodType string `query:"period_type" enum:"DAILY,MONTHLY,YEARLY" default:"MONTHLY" doc:"Type of period aggregation"`
	Preset     string `query:"preset" enum:"7d,30d,3m,6m,12m,ytd" doc:"Preset time range"`
	StartDate  string `query:"start_date" doc:"Custom start date (YYYY-MM-DD)"`
	EndDate    string `query:"end_date" doc:"Custom end date (YYYY-MM-DD)"`
	Compare    bool   `query:"compare" default:"true" doc:"Compare with previous period"`
}

// SalesChartResponse represents the response for sales chart data
type SalesChartResponse struct {
	PeriodType   string           `json:"period_type"`
	StartDate    string           `json:"start_date"`
	EndDate      string           `json:"end_date"`
	TotalRevenue float64          `json:"total_revenue"`
	Growth       float64          `json:"growth_percentage"`
	Currency     string           `json:"currency"`
	ChartData    []ChartDataPoint `json:"chart_data"`
	Comparison   *ComparisonData  `json:"comparison_data,omitempty"`
}

// ChartDataPoint represents a single point in the chart
type ChartDataPoint struct {
	Period         string  `json:"period"`
	Value          float64 `json:"value"`
	FormattedValue string  `json:"formatted_value"`
}

// ComparisonData represents comparison with previous period
type ComparisonData struct {
	PreviousRevenue float64 `json:"previous_period_revenue"`
	Difference      float64 `json:"difference"`
	GrowthRate      float64 `json:"growth_rate"`
}

// SalesDataPoint represents raw data from database
type SalesDataPoint struct {
	Period      string
	PeriodOrder int       // For sorting (month number, day of year, etc)
	Date        time.Time // Actual date for reference
	Revenue     float64
}
