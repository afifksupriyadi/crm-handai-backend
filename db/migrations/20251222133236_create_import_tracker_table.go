package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create import_tracker table")

		query := `
			CREATE TABLE IF NOT EXISTS import_tracker (
				id SERIAL PRIMARY KEY,
				last_import_end_date DATE NOT NULL,
				last_window_end_date DATE NOT NULL,
				pending_window_start DATE NULL,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			-- Only one row should exist in this table
			CREATE UNIQUE INDEX IF NOT EXISTS idx_import_tracker_singleton 
				ON import_tracker((1));

			-- Index for date queries
			CREATE INDEX IF NOT EXISTS idx_import_tracker_dates 
				ON import_tracker(last_import_end_date, last_window_end_date);

			-- Insert initial record (will be updated on first import)
			INSERT INTO import_tracker (
				last_import_end_date,
				last_window_end_date,
				pending_window_start
			) VALUES (
				'1970-01-01'::DATE,
				'1970-01-01'::DATE,
				NULL
			);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create import_tracker table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop import_tracker table")

		query := `
			DROP INDEX IF EXISTS idx_import_tracker_singleton;
			DROP INDEX IF EXISTS idx_import_tracker_dates;
			DROP TABLE IF EXISTS import_tracker CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop import_tracker table: %w", err)
		}

		return nil
	})
}
