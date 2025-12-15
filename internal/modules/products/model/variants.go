package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Variant struct {
	bun.BaseModel `bun:"table:variants"`

	ID            int        `bun:"id,pk,autoincrement"`
	ProductID     int        `bun:"product_id,notnull"`
	Name          string     `bun:"name,notnull"`
	PriceModifier float64    `bun:"price_modifier,notnull,default:0"`
	IsDefault     bool       `bun:"is_default,notnull,default:false"`
	CreatedAt     time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     *time.Time `bun:"updated_at,nullzero"`
	DeletedAt     *time.Time `bun:"deleted_at,soft_delete,nullzero"`

	// Relations
	Product *Product `bun:"rel:belongs-to,join:product_id=id"`
}
