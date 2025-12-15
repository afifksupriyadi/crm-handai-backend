package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create sales_forecasts table")

		query := `
			CREATE SEQUENCE IF NOT EXISTS sales_forecasts_id_seq;

			CREATE TABLE IF NOT EXISTS sales_forecasts (
				id INTEGER NOT NULL DEFAULT nextval('sales_forecasts_id_seq'::regclass),
				batch_id INTEGER NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
				forecast_date DATE NOT NULL,
				period_type VARCHAR(20) NOT NULL CHECK (period_type IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY')),
				predicted_revenue DECIMAL(12,2),
				predicted_transactions INTEGER,
				confidence_interval_lower DECIMAL(12,2),
				confidence_interval_upper DECIMAL(12,2),
				target_revenue DECIMAL(12,2),
				actual_revenue DECIMAL(12,2),
				model_version VARCHAR(20),
				computed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				CONSTRAINT unique_batch_forecast_date UNIQUE (batch_id, forecast_date, period_type)
			);

			CREATE INDEX IF NOT EXISTS idx_forecasts_batch ON sales_forecasts(batch_id);
			CREATE INDEX IF NOT EXISTS idx_forecasts_date ON sales_forecasts(forecast_date);
			CREATE INDEX IF NOT EXISTS idx_forecasts_period ON sales_forecasts(period_type);

			ALTER SEQUENCE sales_forecasts_id_seq OWNED BY sales_forecasts.id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create sales_forecasts table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop sales_forecasts table")

		query := `
			DROP INDEX IF EXISTS idx_forecasts_batch;
			DROP INDEX IF EXISTS idx_forecasts_date;
			DROP INDEX IF EXISTS idx_forecasts_period;
			DROP TABLE IF EXISTS sales_forecasts CASCADE;
			DROP SEQUENCE IF EXISTS sales_forecasts_id_seq;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop sales_forecasts table: %w", err)
		}

		return nil
	})
}
