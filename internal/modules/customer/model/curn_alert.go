package model

import (
	"time"

	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/uptrace/bun"
)

// ChurnAlert represents customer churn alerts (in analytics schema)
type ChurnAlert struct {
	bun.BaseModel `bun:"table:analytics.churn_alerts,alias:ca"`

	ID                   int        `bun:"id,pk,autoincrement" json:"id"`
	CustomerID           int        `bun:"customer_id,notnull" json:"customer_id"`
	TransactionBatchID   int        `bun:"transaction_batch_id,notnull" json:"transaction_batch_id"`
	AlertType            string     `bun:"alert_type,notnull,type:varchar(20)" json:"alert_type"`
	ExpectedPurchaseDate time.Time  `bun:"expected_purchase_date,notnull" json:"expected_purchase_date"`
	DaysOverdue          int        `bun:"days_overdue,notnull" json:"days_overdue"`
	ChurnProbability     *float64   `bun:"churn_probability,type:numeric(3,2)" json:"churn_probability,omitempty"`
	Status               string     `bun:"status,notnull,type:varchar(20)" json:"status"`
	NotifiedAt           *time.Time `bun:"notified_at" json:"notified_at,omitempty"`
	ResolvedAt           *time.Time `bun:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt            time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt            *time.Time `bun:"updated_at" json:"updated_at,omitempty"`

	// Relations
	Customer         *Customer                         `bun:"rel:belongs-to,join:customer_id=id" json:"customer,omitempty"`
	TransactionBatch *importDataModel.TransactionBatch `bun:"rel:belongs-to,join:transaction_batch_id=id" json:"transaction_batch,omitempty"`
}
