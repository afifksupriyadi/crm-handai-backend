package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Batch struct {
	bun.BaseModel `bun:"table:batches"`

	ID                  int        `bun:"id,pk,autoincrement"`
	BatchDate           time.Time  `bun:"batch_date,notnull"`
	BatchCode           string     `bun:"batch_code,notnull,unique"`
	Status              string     `bun:"status,notnull"` // PROCESSING, COMPLETED, FAILED
	IsActive            bool       `bun:"is_active,default:false"`
	CustomerImportID    *int       `bun:"customer_import_id"`
	TransactionImportID *int       `bun:"transaction_import_id"`
	CreatedAt           time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt           *time.Time `bun:"updated_at"`

	// Relations
	CustomerImportLog    *ImportLog `bun:"rel:belongs-to,join:customer_import_id=id"`
	TransactionImportLog *ImportLog `bun:"rel:belongs-to,join:transaction_import_id=id"`
}
