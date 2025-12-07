package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:products"`

	ID        int        `bun:"id,pk,autoincrement"`
	Name      string     `bun:"name,notnull"`
	Category  string     `bun:"category,notnull"` // 'coffee' or 'non-coffee'
	BasePrice int64      `bun:"base_price,notnull"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt *time.Time `bun:"updated_at,nullzero"`
	DeletedAt *time.Time `bun:"deleted_at,soft_delete,nullzero"`

	// Relations
	Variants []*Variant `bun:"rel:has-many,join:id=product_id"`
}
