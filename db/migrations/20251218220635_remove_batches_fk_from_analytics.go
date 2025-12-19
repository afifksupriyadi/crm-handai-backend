package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] removing batches FK from analytics tables...")

		queries := []string{
			// Drop FKs from customer_predictions
			`ALTER TABLE analytics.customer_predictions DROP CONSTRAINT IF EXISTS customer_predictions_batch_id_fkey`,
			`ALTER TABLE analytics.customer_predictions DROP CONSTRAINT IF EXISTS fk_customer_predictions_batch`,

			// Drop FKs from sales_forecasts
			`ALTER TABLE analytics.sales_forecasts DROP CONSTRAINT IF EXISTS sales_forecasts_batch_id_fkey`,
			`ALTER TABLE analytics.sales_forecasts DROP CONSTRAINT IF EXISTS fk_sales_forecasts_batch`,

			// Drop FKs from churn_alerts
			`ALTER TABLE analytics.churn_alerts DROP CONSTRAINT IF EXISTS churn_alerts_batch_id_fkey`,
			`ALTER TABLE analytics.churn_alerts DROP CONSTRAINT IF EXISTS fk_churn_alerts_batch`,

			// Drop FKs from customer_metrics
			`ALTER TABLE analytics.customer_metrics DROP CONSTRAINT IF EXISTS fk_customer_metrics_batch`,
			`ALTER TABLE analytics.customer_metrics DROP CONSTRAINT IF EXISTS customer_metrics_batch_id_fkey`,
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				// Ignore errors if constraint doesn't exist
				fmt.Printf(" (constraint may not exist, continuing...)")
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] cannot restore batches FK (batches table will be dropped)")
		return nil
	})
}
