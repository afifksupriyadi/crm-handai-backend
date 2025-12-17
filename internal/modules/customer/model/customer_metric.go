package model

import (
	"time"

	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/uptrace/bun"
)

// CustomerMetric represents customer analytics metrics (in analytics schema)
type CustomerMetric struct {
	bun.BaseModel `bun:"table:analytics.customer_metrics,alias:cm"`

	CustomerID             int        `bun:"customer_id,pk" json:"customer_id"`
	BatchID                int        `bun:"batch_id,pk" json:"batch_id"`
	TotalTransactions      int        `bun:"total_transactions,notnull,default:0" json:"total_transactions"`
	TotalSpent             float64    `bun:"total_spent,type:numeric(12,2),notnull,default:0" json:"total_spent"`
	LastTransactionDate    *time.Time `bun:"last_transaction_date" json:"last_transaction_date,omitempty"`
	AvgDaysBetweenPurchase *float64   `bun:"avg_days_between_purchase,type:numeric(5,2)" json:"avg_days_between_purchase,omitempty"`
	Segment                *string    `bun:"segment,type:varchar(20)" json:"segment,omitempty"` // NEW, POTENTIAL, LOYAL, CHURN
	IsLoyal                bool       `bun:"is_loyal,notnull,default:false" json:"is_loyal"`
	ChurnRiskScore         *float64   `bun:"churn_risk_score,type:numeric(3,2)" json:"churn_risk_score,omitempty"`
	ComputedAt             time.Time  `bun:"computed_at,notnull,default:current_timestamp" json:"computed_at"`

	// Relations
	Customer *Customer              `bun:"rel:belongs-to,join:customer_id=id" json:"customer,omitempty"`
	Batch    *importDataModel.Batch `bun:"rel:belongs-to,join:batch_id=id" json:"batch,omitempty"`
}

// CustomerSegment constants
const (
	SegmentNew       = "NEW"
	SegmentPotential = "POTENTIAL"
	SegmentLoyal     = "LOYAL"
	SegmentChurn     = "CHURN"
)
