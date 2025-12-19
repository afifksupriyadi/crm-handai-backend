package constant

// AlertType represents customer churn alert types
type AlertType string

const (
	AlertTypeMissedCycle AlertType = "MISSED_CYCLE"
	AlertTypeHighRisk    AlertType = "HIGH_RISK"
	AlertTypeDormant     AlertType = "DORMANT"
)
