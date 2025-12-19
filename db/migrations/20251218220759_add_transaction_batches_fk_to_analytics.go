package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] adding transaction_batches FK to analytics tables...")

		queries := []string{
			`ALTER TABLE analytics.customer_predictions 
				ADD CONSTRAINT fk_customer_predictions_transaction_batch 
				FOREIGN KEY (transaction_batch_id) 
				REFERENCES transaction_batches(id) 
				ON DELETE CASCADE`,

			`ALTER TABLE analytics.sales_forecasts 
				ADD CONSTRAINT fk_sales_forecasts_transaction_batch 
				FOREIGN KEY (transaction_batch_id) 
				REFERENCES transaction_batches(id) 
				ON DELETE CASCADE`,

			`ALTER TABLE analytics.churn_alerts 
				ADD CONSTRAINT fk_churn_alerts_transaction_batch 
				FOREIGN KEY (transaction_batch_id) 
				REFERENCES transaction_batches(id) 
				ON DELETE CASCADE`,

			`ALTER TABLE analytics.customer_metrics 
				ADD CONSTRAINT fk_customer_metrics_transaction_batch 
				FOREIGN KEY (transaction_batch_id) 
				REFERENCES transaction_batches(id) 
				ON DELETE CASCADE`,
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("failed to add FK: %w", err)
			}
		}

		// Update indexes
		indexQueries := []string{
			`DROP INDEX IF EXISTS analytics.idx_predictions_batch`,
			`CREATE INDEX idx_predictions_transaction_batch ON analytics.customer_predictions(transaction_batch_id)`,

			`DROP INDEX IF EXISTS analytics.idx_forecasts_batch`,
			`CREATE INDEX idx_forecasts_transaction_batch ON analytics.sales_forecasts(transaction_batch_id)`,

			`DROP INDEX IF EXISTS analytics.idx_alerts_batch`,
			`CREATE INDEX idx_alerts_transaction_batch ON analytics.churn_alerts(transaction_batch_id)`,

			`DROP INDEX IF EXISTS analytics.idx_customer_metrics_batch`,
			`CREATE INDEX idx_customer_metrics_transaction_batch ON analytics.customer_metrics(transaction_batch_id)`,
		}

		for _, query := range indexQueries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				// Ignore errors for DROP INDEX (may not exist)
				continue
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] removing transaction_batches FK from analytics tables...")

		queries := []string{
			`ALTER TABLE analytics.customer_predictions DROP CONSTRAINT IF EXISTS fk_customer_predictions_transaction_batch`,
			`ALTER TABLE analytics.sales_forecasts DROP CONSTRAINT IF EXISTS fk_sales_forecasts_transaction_batch`,
			`ALTER TABLE analytics.churn_alerts DROP CONSTRAINT IF EXISTS fk_churn_alerts_transaction_batch`,
			`ALTER TABLE analytics.customer_metrics DROP CONSTRAINT IF EXISTS fk_customer_metrics_transaction_batch`,
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("failed to drop FK: %w", err)
			}
		}

		fmt.Println(" done")
		return nil
	})
}
