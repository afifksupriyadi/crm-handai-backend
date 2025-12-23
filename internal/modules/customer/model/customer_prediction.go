package model

import (
	"time"

	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"
	"github.com/uptrace/bun"
)

// CustomerPrediction represents customer purchase predictions (in analytics schema)
// Simplified version focused on 7-day window system
type CustomerPrediction struct {
	bun.BaseModel `bun:"table:analytics.customer_predictions,alias:cp"`

	ID                        int        `bun:"id,pk,autoincrement" json:"id"`
	CustomerID                int        `bun:"customer_id,notnull" json:"customer_id"`
	TransactionBatchID        int        `bun:"transaction_batch_id,notnull" json:"transaction_batch_id"`
	LastTransactionDate       time.Time  `bun:"last_transaction_date,notnull" json:"last_transaction_date"`
	PredictedNextPurchaseDate time.Time  `bun:"predicted_next_purchase_date,notnull" json:"predicted_next_purchase_date"`
	ActualNextPurchaseDate    *time.Time `bun:"actual_next_purchase_date" json:"actual_next_purchase_date,omitempty"`
	IsPredictedCorrect        *bool      `bun:"is_predicted_correct" json:"is_predicted_correct,omitempty"`
	CreatedAt                 time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	ValidatedAt               *time.Time `bun:"validated_at" json:"validated_at,omitempty"`

	// Relations
	Customer         *Customer                         `bun:"rel:belongs-to,join:customer_id=id" json:"customer,omitempty"`
	TransactionBatch *importDataModel.TransactionBatch `bun:"rel:belongs-to,join:transaction_batch_id=id" json:"transaction_batch,omitempty"`
}
