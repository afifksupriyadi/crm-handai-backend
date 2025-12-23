package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create customer_segments table")

		query := `
			CREATE TABLE IF NOT EXISTS analytics.customer_segments (
				customer_id INTEGER PRIMARY KEY,
				segment VARCHAR(20) NOT NULL,
				
				consecutive_correct_predictions INTEGER NOT NULL DEFAULT 0,
				total_predictions INTEGER NOT NULL DEFAULT 0,
				total_correct_predictions INTEGER NOT NULL DEFAULT 0,
				
				last_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_by_batch_id INTEGER NULL,
				
				CONSTRAINT fk_customer_segments_customer 
					FOREIGN KEY (customer_id) 
					REFERENCES public.customers(id) 
					ON DELETE CASCADE,
				CONSTRAINT chk_customer_segments_segment 
					CHECK (segment IN ('LOYAL', 'CHURN', 'REGULAR'))
			);

			-- Indexes
			CREATE INDEX IF NOT EXISTS idx_customer_segments_segment 
				ON analytics.customer_segments(segment);
			
			CREATE INDEX IF NOT EXISTS idx_customer_segments_updated 
				ON analytics.customer_segments(last_updated_at DESC);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create customer_segments table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop customer_segments table")

		query := `
			DROP INDEX IF EXISTS idx_customer_segments_segment;
			DROP INDEX IF EXISTS idx_customer_segments_updated;
			DROP TABLE IF EXISTS analytics.customer_segments CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop customer_segments table: %w", err)
		}

		return nil
	})
}
