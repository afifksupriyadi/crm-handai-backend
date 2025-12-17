package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] creating analytics.customer_metrics table...")

		// Create customer_metrics table in analytics schema
		_, err := db.ExecContext(ctx, `
			CREATE TABLE analytics.customer_metrics (
				customer_id INTEGER NOT NULL,
				batch_id INTEGER NOT NULL,
				total_transactions INTEGER DEFAULT 0 NOT NULL,
				total_spent NUMERIC(12,2) DEFAULT 0 NOT NULL,
				last_transaction_date TIMESTAMPTZ,
				avg_days_between_purchase NUMERIC(5,2),
				segment VARCHAR(20) CHECK (segment IN ('NEW', 'POTENTIAL', 'LOYAL', 'CHURN')),
				is_loyal BOOLEAN DEFAULT FALSE NOT NULL,
				churn_risk_score NUMERIC(3,2) CHECK (churn_risk_score >= 0 AND churn_risk_score <= 1),
				computed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
				
				-- Composite primary key
				PRIMARY KEY (customer_id, batch_id),
				
				-- Foreign keys
				CONSTRAINT fk_customer_metrics_customer 
					FOREIGN KEY (customer_id) 
					REFERENCES public.customers(id) 
					ON DELETE CASCADE,
				CONSTRAINT fk_customer_metrics_batch 
					FOREIGN KEY (batch_id) 
					REFERENCES public.batches(id) 
					ON DELETE CASCADE
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create customer_metrics table: %w", err)
		}

		// Create indexes for performance
		indexes := []string{
			`CREATE INDEX idx_customer_metrics_segment ON analytics.customer_metrics(segment)`,
			`CREATE INDEX idx_customer_metrics_batch ON analytics.customer_metrics(batch_id)`,
			`CREATE INDEX idx_customer_metrics_loyal ON analytics.customer_metrics(is_loyal)`,
			`CREATE INDEX idx_customer_metrics_last_transaction ON analytics.customer_metrics(last_transaction_date)`,
			`CREATE INDEX idx_customer_metrics_churn_risk ON analytics.customer_metrics(churn_risk_score)`,
		}

		for _, indexSQL := range indexes {
			if _, err := db.ExecContext(ctx, indexSQL); err != nil {
				return fmt.Errorf("failed to create index: %w", err)
			}
		}

		fmt.Println(" done")
		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] dropping analytics.customer_metrics table...")

		// Drop table (indexes will be dropped automatically)
		_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS analytics.customer_metrics CASCADE`)
		if err != nil {
			return fmt.Errorf("failed to drop customer_metrics table: %w", err)
		}

		fmt.Println(" done")
		return nil
	})
}
