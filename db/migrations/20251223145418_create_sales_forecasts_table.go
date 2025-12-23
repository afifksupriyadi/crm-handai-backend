package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] create sales_forecasts table")

		query := `
			CREATE TABLE IF NOT EXISTS analytics.sales_forecasts (
				id SERIAL PRIMARY KEY,
				transaction_batch_id INTEGER NOT NULL,
				
				forecast_period VARCHAR(20) NOT NULL CHECK (forecast_period IN ('WEEKLY', 'MONTHLY', 'YEARLY')),
				forecast_date DATE NOT NULL,
				
				minimum_revenue NUMERIC(12,2) NOT NULL DEFAULT 0,
				normal_revenue NUMERIC(12,2) NOT NULL DEFAULT 0,
				maximum_revenue NUMERIC(12,2) NOT NULL DEFAULT 0,
				
				actual_revenue NUMERIC(12,2),
				
				computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				
				CONSTRAINT fk_sales_forecasts_batch 
					FOREIGN KEY (transaction_batch_id) 
					REFERENCES transaction_batches(id) 
					ON DELETE CASCADE,
				
				CONSTRAINT unique_batch_period_date 
					UNIQUE (transaction_batch_id, forecast_period, forecast_date)
			);

			CREATE INDEX IF NOT EXISTS idx_sales_forecasts_batch 
				ON analytics.sales_forecasts(transaction_batch_id);
			
			CREATE INDEX IF NOT EXISTS idx_sales_forecasts_period 
				ON analytics.sales_forecasts(forecast_period);
			
			CREATE INDEX IF NOT EXISTS idx_sales_forecasts_date 
				ON analytics.sales_forecasts(forecast_date);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create sales_forecasts table: %w", err)
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] drop sales_forecasts table")

		query := `
			DROP INDEX IF EXISTS analytics.idx_sales_forecasts_batch;
			DROP INDEX IF EXISTS analytics.idx_sales_forecasts_period;
			DROP INDEX IF EXISTS analytics.idx_sales_forecasts_date;
			DROP TABLE IF EXISTS analytics.sales_forecasts CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop sales_forecasts table: %w", err)
		}

		fmt.Println(" done")
		return nil
	})
}
