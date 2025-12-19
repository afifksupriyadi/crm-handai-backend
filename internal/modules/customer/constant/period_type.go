package constant

// PeriodType represents sales forecast period types
type PeriodType string

const (
	PeriodDaily   PeriodType = "DAILY"
	PeriodWeekly  PeriodType = "WEEKLY"
	PeriodMonthly PeriodType = "MONTHLY"
	PeriodYearly  PeriodType = "YEARLY"
)
