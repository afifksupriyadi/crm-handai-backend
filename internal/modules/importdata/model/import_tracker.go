package model

import (
	"time"

	"github.com/uptrace/bun"
)

// ImportTracker tracks the last import dates and window processing state
// Only ONE row should exist in this table (singleton pattern)
type ImportTracker struct {
	bun.BaseModel `bun:"table:import_tracker,alias:it"`

	ID                 int        `bun:"id,pk,autoincrement"`
	LastImportEndDate  time.Time  `bun:"last_import_end_date,notnull"`
	LastWindowEndDate  time.Time  `bun:"last_window_end_date,notnull"`
	PendingWindowStart *time.Time `bun:"pending_window_start"`
	CreatedAt          time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt          time.Time  `bun:"updated_at,notnull,default:current_timestamp"`
}
