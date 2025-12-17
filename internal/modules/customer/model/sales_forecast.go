package model

import (
	"time"

	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"

	"github.com/uptrace/bun"
)

// SalesForecast represents sales predictions (in analytics schema)
type SalesForecast struct {
	bun.BaseModel `bun:"table:analytics.sales_forecasts,alias:sf"`

	ID                      int       `bun:"id,pk,autoincrement" json:"id"`
	BatchID                 int       `bun:"batch_id,notnull" json:"batch_id"`
	ForecastDate            time.Time `bun:"forecast_date,notnull" json:"forecast_date"`
	PeriodType              string    `bun:"period_type,notnull,type:varchar(20)" json:"period_type"` // DAILY, WEEKLY, MONTHLY, YEARLY
	PredictedRevenue        *float64  `bun:"predicted_revenue,type:numeric(12,2)" json:"predicted_revenue,omitempty"`
	PredictedTransactions   *int      `bun:"predicted_transactions" json:"predicted_transactions,omitempty"`
	ConfidenceIntervalLower *float64  `bun:"confidence_interval_lower,type:numeric(12,2)" json:"confidence_interval_lower,omitempty"`
	ConfidenceIntervalUpper *float64  `bun:"confidence_interval_upper,type:numeric(12,2)" json:"confidence_interval_upper,omitempty"`
	TargetRevenue           *float64  `bun:"target_revenue,type:numeric(12,2)" json:"target_revenue,omitempty"`
	ActualRevenue           *float64  `bun:"actual_revenue,type:numeric(12,2)" json:"actual_revenue,omitempty"`
	ModelVersion            *string   `bun:"model_version,type:varchar(20)" json:"model_version,omitempty"`
	ComputedAt              time.Time `bun:"computed_at,notnull,default:current_timestamp" json:"computed_at"`

	// Relations
	Batch *importDataModel.Batch `bun:"rel:belongs-to,join:batch_id=id" json:"batch,omitempty"`
}

// Period type constants
const (
	PeriodDaily   = "DAILY"
	PeriodWeekly  = "WEEKLY"
	PeriodMonthly = "MONTHLY"
	PeriodYearly  = "YEARLY"
)
