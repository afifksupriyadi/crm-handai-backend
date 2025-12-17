package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] moving analytics tables to analytics schema...")

		// Move tables from public schema to analytics schema
		tables := []string{
			"customer_predictions",
			"sales_forecasts",
			"churn_alerts",
		}

		for _, table := range tables {
			// Check if table exists in public schema
			var exists bool
			err := db.NewSelect().
				ColumnExpr("EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", table).
				Scan(ctx, &exists)
			if err != nil {
				return fmt.Errorf("failed to check if table %s exists: %w", table, err)
			}

			if exists {
				// Move table to analytics schema
				_, err = db.ExecContext(ctx, fmt.Sprintf(`
					ALTER TABLE public.%s SET SCHEMA analytics
				`, table))
				if err != nil {
					return fmt.Errorf("failed to move table %s to analytics schema: %w", table, err)
				}
				fmt.Printf("  - moved %s to analytics schema\n", table)
			} else {
				fmt.Printf("  - table %s not found in public schema, skipping\n", table)
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] moving analytics tables back to public schema...")

		// Move tables from analytics schema back to public schema
		tables := []string{
			"customer_predictions",
			"sales_forecasts",
			"churn_alerts",
		}

		for _, table := range tables {
			// Check if table exists in analytics schema
			var exists bool
			err := db.NewSelect().
				ColumnExpr("EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'analytics' AND table_name = ?)", table).
				Scan(ctx, &exists)
			if err != nil {
				return fmt.Errorf("failed to check if table %s exists: %w", table, err)
			}

			if exists {
				// Move table back to public schema
				_, err = db.ExecContext(ctx, fmt.Sprintf(`
					ALTER TABLE analytics.%s SET SCHEMA public
				`, table))
				if err != nil {
					return fmt.Errorf("failed to move table %s back to public schema: %w", table, err)
				}
				fmt.Printf("  - moved %s back to public schema\n", table)
			}
		}

		fmt.Println(" done")
		return nil
	})
}
