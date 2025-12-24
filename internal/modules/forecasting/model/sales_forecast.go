package model

import (
	"time"

	"github.com/uptrace/bun"
)

type ForecastPeriod string

const (
	ForecastPeriodWeekly  ForecastPeriod = "WEEKLY"
	ForecastPeriodMonthly ForecastPeriod = "MONTHLY"
	ForecastPeriodYearly  ForecastPeriod = "YEARLY"
)

func (f ForecastPeriod) String() string {
	return string(f)
}

type SalesForecast struct {
	bun.BaseModel `bun:"table:analytics.sales_forecasts,alias:sf"`

	ID                 int            `bun:"id,pk,autoincrement"`
	TransactionBatchID int            `bun:"transaction_batch_id,notnull"`
	ForecastPeriod     ForecastPeriod `bun:"forecast_period,notnull"`
	ForecastDate       time.Time      `bun:"forecast_date,notnull"`
	MinimumRevenue     float64        `bun:"minimum_revenue,notnull,default:0"`
	NormalRevenue      float64        `bun:"normal_revenue,notnull,default:0"`
	MaximumRevenue     float64        `bun:"maximum_revenue,notnull,default:0"`
	ActualRevenue      *float64       `bun:"actual_revenue"`
	ComputedAt         time.Time      `bun:"computed_at,notnull,default:current_timestamp"`
}
