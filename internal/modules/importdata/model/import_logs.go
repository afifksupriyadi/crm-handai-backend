package model

import (
	"time"

	"github.com/uptrace/bun"
)

type ImportLog struct {
	bun.BaseModel `bun:"table:import_logs"`

	ID                 int       `bun:"id,pk,autoincrement"`
	ImportType         string    `bun:"import_type,notnull"`
	FileDate           time.Time `bun:"file_date,notnull"`
	Filename           string    `bun:"filename,notnull"`
	RowsImported       int       `bun:"rows_imported,notnull,default:0"`
	Status             string    `bun:"status,notnull"`
	ImportedAt         time.Time `bun:"imported_at,notnull,default:current_timestamp"`
	CustomerBatchID    *int      `bun:"customer_batch_id"`
	TransactionBatchID *int      `bun:"transaction_batch_id"`
}
