package constant

// AlertStatus represents churn alert status
type AlertStatus string

const (
	AlertStatusPending  AlertStatus = "PENDING"
	AlertStatusNotified AlertStatus = "NOTIFIED"
	AlertStatusResolved AlertStatus = "RESOLVED"
	AlertStatusIgnored  AlertStatus = "IGNORED"
)
