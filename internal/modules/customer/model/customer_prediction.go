package model

import (
	"time"

	importDataModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/importdata/model"

	"github.com/uptrace/bun"
)

// CustomerPrediction represents customer purchase predictions (in analytics schema)
type CustomerPrediction struct {
	bun.BaseModel `bun:"table:analytics.customer_predictions,alias:cp"`

	ID                     int        `bun:"id,pk,autoincrement" json:"id"`
	BatchID                int        `bun:"batch_id,notnull" json:"batch_id"`
	CustomerID             int        `bun:"customer_id,notnull" json:"customer_id"`
	NextPurchaseDate       *time.Time `bun:"next_purchase_date" json:"next_purchase_date,omitempty"`
	ConfidenceScore        *float64   `bun:"confidence_score,type:numeric(3,2)" json:"confidence_score,omitempty"`
	PredictedQuantity      *int       `bun:"predicted_quantity" json:"predicted_quantity,omitempty"`
	PredictedProducts      []byte     `bun:"predicted_products,type:jsonb" json:"predicted_products,omitempty"`
	AvgDaysBetweenPurchase *float64   `bun:"avg_days_between_purchase,type:numeric(5,2)" json:"avg_days_between_purchase,omitempty"`
	Last5Purchases         []byte     `bun:"last_5_purchases,type:jsonb" json:"last_5_purchases,omitempty"`
	ModelVersion           *string    `bun:"model_version,type:varchar(20)" json:"model_version,omitempty"`
	ComputedAt             time.Time  `bun:"computed_at,notnull,default:current_timestamp" json:"computed_at"`

	// Relations
	Customer *Customer              `bun:"rel:belongs-to,join:customer_id=id" json:"customer,omitempty"`
	Batch    *importDataModel.Batch `bun:"rel:belongs-to,join:batch_id=id" json:"batch,omitempty"`
}

// PredictedProduct represents a predicted product structure
type PredictedProduct struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}
