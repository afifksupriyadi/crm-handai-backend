package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] alter customers add segmentation fields")

		query := `
			ALTER TABLE customers
			ADD COLUMN IF NOT EXISTS segment VARCHAR(20) CHECK (segment IN ('NEW', 'POTENTIAL', 'LOYAL', 'CHURN')),
			ADD COLUMN IF NOT EXISTS is_loyal BOOLEAN DEFAULT FALSE,
			ADD COLUMN IF NOT EXISTS total_transactions INTEGER DEFAULT 0,
			ADD COLUMN IF NOT EXISTS total_spent DECIMAL(12,2) DEFAULT 0,
			ADD COLUMN IF NOT EXISTS last_transaction_date TIMESTAMPTZ,
			ADD COLUMN IF NOT EXISTS avg_days_between_purchase DECIMAL(5,2),
			ADD COLUMN IF NOT EXISTS churn_risk_score DECIMAL(3,2) CHECK (churn_risk_score >= 0 AND churn_risk_score <= 1);

			CREATE INDEX IF NOT EXISTS idx_customers_segment ON customers(segment);
			CREATE INDEX IF NOT EXISTS idx_customers_is_loyal ON customers(is_loyal);
			CREATE INDEX IF NOT EXISTS idx_customers_last_transaction ON customers(last_transaction_date);
			CREATE INDEX IF NOT EXISTS idx_customers_churn_risk ON customers(churn_risk_score);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to alter customers table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove segmentation fields from customers")

		query := `
			DROP INDEX IF EXISTS idx_customers_segment;
			DROP INDEX IF EXISTS idx_customers_is_loyal;
			DROP INDEX IF EXISTS idx_customers_last_transaction;
			DROP INDEX IF EXISTS idx_customers_churn_risk;

			ALTER TABLE customers
			DROP COLUMN IF EXISTS segment,
			DROP COLUMN IF EXISTS is_loyal,
			DROP COLUMN IF EXISTS total_transactions,
			DROP COLUMN IF EXISTS total_spent,
			DROP COLUMN IF EXISTS last_transaction_date,
			DROP COLUMN IF EXISTS avg_days_between_purchase,
			DROP COLUMN IF EXISTS churn_risk_score;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to rollback customers table: %w", err)
		}

		return nil
	})
}
