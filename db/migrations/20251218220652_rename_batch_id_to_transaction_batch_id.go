package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] renaming batch_id to transaction_batch_id in analytics tables...")

		queries := []string{
			`ALTER TABLE analytics.customer_predictions RENAME COLUMN batch_id TO transaction_batch_id`,
			`ALTER TABLE analytics.sales_forecasts RENAME COLUMN batch_id TO transaction_batch_id`,
			`ALTER TABLE analytics.churn_alerts RENAME COLUMN batch_id TO transaction_batch_id`,
			`ALTER TABLE analytics.customer_metrics RENAME COLUMN batch_id TO transaction_batch_id`,
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("failed to rename column: %w", err)
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] renaming transaction_batch_id back to batch_id...")

		queries := []string{
			`ALTER TABLE analytics.customer_predictions RENAME COLUMN transaction_batch_id TO batch_id`,
			`ALTER TABLE analytics.sales_forecasts RENAME COLUMN transaction_batch_id TO batch_id`,
			`ALTER TABLE analytics.churn_alerts RENAME COLUMN transaction_batch_id TO batch_id`,
			`ALTER TABLE analytics.customer_metrics RENAME COLUMN transaction_batch_id TO batch_id`,
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("failed to rename column: %w", err)
			}
		}

		fmt.Println(" done")
		return nil
	})
}
