package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] drop deprecated prediction tables")

		query := `
			-- Drop old prediction-related tables that are no longer used
			DROP TABLE IF EXISTS analytics.churn_alerts CASCADE;
			DROP TABLE IF EXISTS analytics.sales_forecasts CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop deprecated tables: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] restore deprecated prediction tables")

		// Recreate table structures (data cannot be restored)
		query := `
			-- Recreate churn_alerts table (structure only, data lost)
			CREATE TABLE IF NOT EXISTS analytics.churn_alerts (
				id SERIAL PRIMARY KEY,
				customer_id INTEGER NOT NULL,
				alert_type VARCHAR(50) NOT NULL,
				severity VARCHAR(20) NOT NULL,
				predicted_churn_date DATE,
				confidence_score NUMERIC(3,2),
				recommended_actions TEXT,
				is_actioned BOOLEAN DEFAULT FALSE,
				actioned_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				resolved_at TIMESTAMPTZ,
				CONSTRAINT fk_churn_alerts_customer 
					FOREIGN KEY (customer_id) 
					REFERENCES public.customers(id) 
					ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_churn_alerts_customer 
				ON analytics.churn_alerts(customer_id);
			
			CREATE INDEX IF NOT EXISTS idx_churn_alerts_date 
				ON analytics.churn_alerts(predicted_churn_date);
			
			CREATE INDEX IF NOT EXISTS idx_churn_alerts_unactioned 
				ON analytics.churn_alerts(is_actioned) 
				WHERE is_actioned = FALSE;

			-- Recreate sales_forecasts table (structure only, data lost)
			CREATE TABLE IF NOT EXISTS analytics.sales_forecasts (
				id SERIAL PRIMARY KEY,
				customer_id INTEGER NOT NULL,
				forecast_period VARCHAR(20) NOT NULL,
				predicted_next_purchase_date DATE,
				predicted_amount NUMERIC(12,2),
				predicted_products JSONB,
				confidence_score NUMERIC(3,2),
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				actual_purchase_date DATE,
				actual_amount NUMERIC(12,2),
				forecast_accuracy NUMERIC(3,2),
				CONSTRAINT fk_sales_forecasts_customer 
					FOREIGN KEY (customer_id) 
					REFERENCES public.customers(id) 
					ON DELETE CASCADE
			);

			CREATE INDEX IF NOT EXISTS idx_sales_forecasts_customer 
				ON analytics.sales_forecasts(customer_id);
			
			CREATE INDEX IF NOT EXISTS idx_sales_forecasts_date 
				ON analytics.sales_forecasts(predicted_next_purchase_date);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to restore deprecated tables: %w", err)
		}

		fmt.Println(" [WARNING] Tables restored but data cannot be recovered")
		return nil
	})
}
