package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] add DAILY to sales_forecasts period enum")

		query := `
			ALTER TABLE analytics.sales_forecasts 
			DROP CONSTRAINT IF EXISTS sales_forecasts_forecast_period_check;

			ALTER TABLE analytics.sales_forecasts 
			ADD CONSTRAINT sales_forecasts_forecast_period_check 
			CHECK (forecast_period IN ('DAILY', 'WEEKLY', 'MONTHLY', 'YEARLY'));
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to add DAILY to forecast period: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove DAILY from sales_forecasts period enum")

		query := `
			ALTER TABLE analytics.sales_forecasts 
			DROP CONSTRAINT IF EXISTS sales_forecasts_forecast_period_check;

			ALTER TABLE analytics.sales_forecasts 
			ADD CONSTRAINT sales_forecasts_forecast_period_check 
			CHECK (forecast_period IN ('WEEKLY', 'MONTHLY', 'YEARLY'));
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to remove DAILY from forecast period: %w", err)
		}

		return nil
	})
}
