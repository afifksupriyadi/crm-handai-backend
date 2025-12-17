package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] dropping analytics columns from customers table...")

		// Drop analytics columns from customers table
		columns := []string{
			"segment",
			"is_loyal",
			"total_transactions",
			"total_spent",
			"last_transaction_date",
			"avg_days_between_purchase",
			"churn_risk_score",
		}

		for _, column := range columns {
			_, err := db.ExecContext(ctx, fmt.Sprintf(`
				ALTER TABLE public.customers 
				DROP COLUMN IF EXISTS %s
			`, column))
			if err != nil {
				return fmt.Errorf("failed to drop column %s: %w", column, err)
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] restoring analytics columns to customers table...")

		// Restore analytics columns to customers table
		_, err := db.ExecContext(ctx, `
			ALTER TABLE public.customers
			ADD COLUMN IF NOT EXISTS segment VARCHAR(20) CHECK (segment IN ('NEW', 'POTENTIAL', 'LOYAL', 'CHURN')),
			ADD COLUMN IF NOT EXISTS is_loyal BOOLEAN DEFAULT FALSE,
			ADD COLUMN IF NOT EXISTS total_transactions INTEGER DEFAULT 0,
			ADD COLUMN IF NOT EXISTS total_spent NUMERIC(12,2) DEFAULT 0,
			ADD COLUMN IF NOT EXISTS last_transaction_date TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS avg_days_between_purchase NUMERIC(5,2),
			ADD COLUMN IF NOT EXISTS churn_risk_score NUMERIC(3,2) CHECK (churn_risk_score >= 0 AND churn_risk_score <= 1)
		`)
		if err != nil {
			return fmt.Errorf("failed to restore analytics columns: %w", err)
		}

		// Restore data from customer_metrics back to customers
		_, err = db.ExecContext(ctx, `
			UPDATE public.customers c
			SET 
				segment = cm.segment,
				is_loyal = cm.is_loyal,
				total_transactions = cm.total_transactions,
				total_spent = cm.total_spent,
				last_transaction_date = cm.last_transaction_date,
				avg_days_between_purchase = cm.avg_days_between_purchase,
				churn_risk_score = cm.churn_risk_score
			FROM analytics.customer_metrics cm
			JOIN public.batches b ON cm.batch_id = b.id
			WHERE cm.customer_id = c.id
				AND b.is_active = TRUE
		`)
		if err != nil {
			return fmt.Errorf("failed to restore analytics data: %w", err)
		}

		fmt.Println(" done")
		return nil
	})
}
