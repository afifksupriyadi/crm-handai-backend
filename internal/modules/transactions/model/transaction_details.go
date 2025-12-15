package model

import (
	"time"

	productModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/products/model"

	"github.com/uptrace/bun"
)

type TransactionDetail struct {
	bun.BaseModel `bun:"table:transaction_details"`

	ID              int        `bun:"id,pk,autoincrement"`
	TransactionCode string     `bun:"transaction_code,notnull"`
	ProductID       int        `bun:"product_id,notnull"`
	VariantID       *int       `bun:"variant_id,nullzero"` // Nullable
	Quantity        int        `bun:"quantity,notnull"`
	UnitPrice       float64    `bun:"unit_price,notnull"`
	Subtotal        float64    `bun:"subtotal,notnull"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt       *time.Time `bun:"updated_at,nullzero"`
	DeletedAt       *time.Time `bun:"deleted_at,soft_delete,nullzero"`

	// Relations
	Transaction *Transaction          `bun:"rel:belongs-to,join:transaction_code=code"`
	Product     *productModel.Product `bun:"rel:belongs-to,join:product_id=id"`
	Variant     *productModel.Variant `bun:"rel:belongs-to,join:variant_id=id"`
}
