package model

import (
	"time"

	"github.com/uptrace/bun"
)

// TransactionBatch represents a transaction data import batch
type TransactionBatch struct {
	bun.BaseModel `bun:"table:transaction_batches,alias:tb"`

	ID                     int       `bun:"id,pk,autoincrement"`
	BatchDate              time.Time `bun:"batch_date,notnull"`
	Filename               string    `bun:"filename,notnull"`
	CustomerBatchID        int       `bun:"customer_batch_id,notnull"`
	ImportedAt             time.Time `bun:"imported_at,notnull,default:now()"`
	TransactionCount       int       `bun:"transaction_count,notnull,default:0"`
	RegisteredTransactions int       `bun:"registered_transactions,notnull,default:0"`
	GuestTransactions      int       `bun:"guest_transactions,notnull,default:0"`
	Notes                  string    `bun:"notes,type:text"`

	// Relations
	CustomerBatch *CustomerBatch `bun:"rel:belongs-to,join:customer_batch_id=id"`
}
