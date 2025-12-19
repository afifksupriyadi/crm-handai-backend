package model

import (
	"time"

	"github.com/uptrace/bun"
)

// CustomerBatch represents a customer data import batch
type CustomerBatch struct {
	bun.BaseModel `bun:"table:customer_batches,alias:cb"`

	ID                int       `bun:"id,pk,autoincrement"`
	BatchDate         time.Time `bun:"batch_date,notnull"`
	Filename          string    `bun:"filename,notnull"`
	ImportedAt        time.Time `bun:"imported_at,notnull,default:now()"`
	CustomerCount     int       `bun:"customer_count,notnull,default:0"`
	NewCustomers      int       `bun:"new_customers,notnull,default:0"`
	UpdatedCustomers  int       `bun:"updated_customers,notnull,default:0"`
	UpgradedFromGuest int       `bun:"upgraded_from_guest,notnull,default:0"`
	IsActive          bool      `bun:"is_active,notnull,default:true"`
	Notes             string    `bun:"notes,type:text"`

	// Relations
	TransactionBatches []*TransactionBatch `bun:"rel:has-many,join:id=customer_batch_id"`
}
