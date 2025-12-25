package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Add window_start_date column
		_, err := db.ExecContext(ctx, `
			ALTER TABLE analytics.customer_predictions
			ADD COLUMN IF NOT EXISTS window_start_date DATE
		`)
		if err != nil {
			return fmt.Errorf("failed to add window_start_date column: %w", err)
		}

		// Add window_end_date column
		_, err = db.ExecContext(ctx, `
			ALTER TABLE analytics.customer_predictions
			ADD COLUMN IF NOT EXISTS window_end_date DATE
		`)
		if err != nil {
			return fmt.Errorf("failed to add window_end_date column: %w", err)
		}

		// Create index for better query performance
		_, err = db.ExecContext(ctx, `
			CREATE INDEX IF NOT EXISTS idx_customer_predictions_window 
			ON analytics.customer_predictions(window_start_date, window_end_date)
		`)
		if err != nil {
			return fmt.Errorf("failed to create window index: %w", err)
		}

		// Add column comments for documentation
		_, err = db.ExecContext(ctx, `
			COMMENT ON COLUMN analytics.customer_predictions.window_start_date 
			IS 'Start date of the processing window when this prediction was generated'
		`)
		if err != nil {
			return fmt.Errorf("failed to add window_start_date comment: %w", err)
		}

		_, err = db.ExecContext(ctx, `
			COMMENT ON COLUMN analytics.customer_predictions.window_end_date 
			IS 'End date of the processing window when this prediction was generated'
		`)
		if err != nil {
			return fmt.Errorf("failed to add window_end_date comment: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		// Rollback: Drop index
		_, err := db.ExecContext(ctx, `
			DROP INDEX IF EXISTS analytics.idx_customer_predictions_window
		`)
		if err != nil {
			return fmt.Errorf("failed to drop window index: %w", err)
		}

		// Rollback: Drop window_end_date column
		_, err = db.ExecContext(ctx, `
			ALTER TABLE analytics.customer_predictions
			DROP COLUMN IF EXISTS window_end_date
		`)
		if err != nil {
			return fmt.Errorf("failed to drop window_end_date column: %w", err)
		}

		// Rollback: Drop window_start_date column
		_, err = db.ExecContext(ctx, `
			ALTER TABLE analytics.customer_predictions
			DROP COLUMN IF EXISTS window_start_date
		`)
		if err != nil {
			return fmt.Errorf("failed to drop window_start_date column: %w", err)
		}

		return nil
	})
}
