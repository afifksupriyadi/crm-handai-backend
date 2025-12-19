package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] migrating customer analytics data to analytics.customer_metrics...")

		// Migrate existing analytics data from customers table to analytics.customer_metrics
		// For each customer, create a record in customer_metrics for the active batch
		_, err := db.ExecContext(ctx, `
			INSERT INTO analytics.customer_metrics (
				customer_id,
				batch_id,
				total_transactions,
				total_spent,
				last_transaction_date,
				avg_days_between_purchase,
				segment,
				is_loyal,
				churn_risk_score,
				computed_at
			)
			SELECT 
				c.id as customer_id,
				b.id as batch_id,
				COALESCE(c.total_transactions, 0) as total_transactions,
				COALESCE(c.total_spent, 0) as total_spent,
				c.last_transaction_date,
				c.avg_days_between_purchase,
				c.segment,
				COALESCE(c.is_loyal, FALSE) as is_loyal,
				c.churn_risk_score,
				CURRENT_TIMESTAMP as computed_at
			FROM public.customers c
			CROSS JOIN (
				-- Get the active batch (or latest batch if no active)
				SELECT id 
				FROM public.batches 
				WHERE is_active = TRUE 
				LIMIT 1
			) b
			WHERE c.deleted_at IS NULL
				-- Only migrate if customer has analytics data
				AND (
					c.total_transactions IS NOT NULL 
					OR c.total_spent IS NOT NULL
					OR c.segment IS NOT NULL
				)
			ON CONFLICT (customer_id, batch_id) DO NOTHING
		`)
		if err != nil {
			return fmt.Errorf("failed to migrate customer analytics data: %w", err)
		}

		// Get count of migrated records
		var count int
		err = db.NewSelect().
			Table("analytics.customer_metrics").
			ColumnExpr("COUNT(*)").
			Scan(ctx, &count)
		if err != nil {
			return fmt.Errorf("failed to count migrated records: %w", err)
		}

		fmt.Printf(" done (%d records migrated)\n", count)
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] clearing analytics.customer_metrics data...")

		// Clear all data from customer_metrics
		// (The columns in customers table will be restored in the next down migration)
		_, err := db.ExecContext(ctx, `TRUNCATE TABLE analytics.customer_metrics CASCADE`)
		if err != nil {
			return fmt.Errorf("failed to clear customer_metrics data: %w", err)
		}

		fmt.Println(" done")
		return nil
	})
}
