package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create customer_predictions v2 table")

		query := `
			-- Drop old table if exists
			DROP TABLE IF EXISTS analytics.customer_predictions CASCADE;

			-- Create new simplified table
			CREATE TABLE IF NOT EXISTS analytics.customer_predictions (
				id SERIAL PRIMARY KEY,
				customer_id INTEGER NOT NULL,
				transaction_batch_id INTEGER NOT NULL,
				
				last_transaction_date DATE NOT NULL,
				predicted_next_purchase_date DATE NOT NULL,
				
				actual_next_purchase_date DATE NULL,
				is_predicted_correct BOOLEAN NULL,
				
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				validated_at TIMESTAMPTZ NULL,
				
				CONSTRAINT fk_customer_predictions_customer 
					FOREIGN KEY (customer_id) 
					REFERENCES public.customers(id) 
					ON DELETE CASCADE,
				CONSTRAINT fk_customer_predictions_batch 
					FOREIGN KEY (transaction_batch_id) 
					REFERENCES public.transaction_batches(id) 
					ON DELETE CASCADE
			);

			-- Indexes
			CREATE INDEX IF NOT EXISTS idx_customer_predictions_customer 
				ON analytics.customer_predictions(customer_id);
			
			CREATE INDEX IF NOT EXISTS idx_customer_predictions_customer_created 
				ON analytics.customer_predictions(customer_id, created_at DESC);
			
			CREATE INDEX IF NOT EXISTS idx_customer_predictions_validation 
				ON analytics.customer_predictions(is_predicted_correct) 
				WHERE is_predicted_correct IS NULL;
			
			CREATE INDEX IF NOT EXISTS idx_customer_predictions_predicted_date 
				ON analytics.customer_predictions(predicted_next_purchase_date);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create customer_predictions v2 table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop customer_predictions v2 table")

		query := `
			DROP INDEX IF EXISTS idx_customer_predictions_customer;
			DROP INDEX IF EXISTS idx_customer_predictions_customer_created;
			DROP INDEX IF EXISTS idx_customer_predictions_validation;
			DROP INDEX IF EXISTS idx_customer_predictions_predicted_date;
			DROP TABLE IF EXISTS analytics.customer_predictions CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop customer_predictions v2 table: %w", err)
		}

		return nil
	})
}
