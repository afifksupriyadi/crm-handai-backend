package model

import (
	"time"

	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"
	"github.com/uptrace/bun"
)

// CustomerPredictedProduct represents predicted products for a customer prediction
type CustomerPredictedProduct struct {
	bun.BaseModel `bun:"table:analytics.customer_predicted_products,alias:cpp"`

	ID                int       `bun:"id,pk,autoincrement" json:"id"`
	PredictionID      int       `bun:"prediction_id,notnull" json:"prediction_id"`
	ProductID         int       `bun:"product_id,notnull" json:"product_id"`
	VariantID         *int      `bun:"variant_id" json:"variant_id,omitempty"`
	PredictedQuantity int       `bun:"predicted_quantity,notnull,default:1" json:"predicted_quantity"`
	ConfidenceScore   float64   `bun:"confidence_score,type:numeric(5,2),notnull,default:0.00" json:"confidence_score"`
	PurchaseFrequency int       `bun:"purchase_frequency,notnull,default:0" json:"purchase_frequency"`
	AvgQuantity       float64   `bun:"avg_quantity,type:numeric(5,2),notnull,default:1.00" json:"avg_quantity"`
	CreatedAt         time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Prediction *CustomerPrediction   `bun:"rel:belongs-to,join:prediction_id=id" json:"prediction,omitempty"`
	Product    *productModel.Product `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Variant    *productModel.Variant `bun:"rel:belongs-to,join:variant_id=id" json:"variant,omitempty"`
}

// TransactionProduct represents a product purchased in a transaction
type TransactionProduct struct {
	TransactionCode string  `bun:"transaction_code"`
	ProductID       int     `bun:"product_id"`
	ProductName     string  `bun:"product_name"`
	VariantID       *int    `bun:"variant_id"`
	VariantName     *string `bun:"variant_name"`
	Quantity        int     `bun:"quantity"`
}
