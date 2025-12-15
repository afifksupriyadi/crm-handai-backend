package model

import (
	"time"

	customerModel "github.com/afifksupriyadi/crm-handai-backend/internal/modules/customer/model"

	"github.com/uptrace/bun"
)

type Transaction struct {
	bun.BaseModel `bun:"table:transactions"`

	Code            string     `bun:"code,pk"`
	CustomerID      *int       `bun:"customer_id,nullzero"`
	TransactionDate time.Time  `bun:"transaction_date,notnull"`
	Discount        float64    `bun:"discount,notnull,default:0"`
	ShippingCost    float64    `bun:"shipping_cost,notnull,default:0"`
	PaymentMethod   string     `bun:"payment_method,notnull"` // 'Tunai' or 'Non Tunai'
	Status          string     `bun:"status,notnull"`         // 'LUNAS' or 'PENDING'
	BatchID         *int       `bun:"batch_id,nullzero"`      // NEW: Link to batch
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt       *time.Time `bun:"updated_at,nullzero"`
	DeletedAt       *time.Time `bun:"deleted_at,soft_delete,nullzero"`

	// Relations
	Customer *customerModel.Customer `bun:"rel:belongs-to,join:customer_id=id"`
	Details  []*TransactionDetail    `bun:"rel:has-many,join:code=transaction_code"`
}
