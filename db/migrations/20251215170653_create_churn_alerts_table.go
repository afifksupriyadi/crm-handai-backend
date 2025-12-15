package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create churn_alerts table")

		query := `
			CREATE SEQUENCE IF NOT EXISTS churn_alerts_id_seq;

			CREATE TABLE IF NOT EXISTS churn_alerts (
				id INTEGER NOT NULL DEFAULT nextval('churn_alerts_id_seq'::regclass),
				customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
				batch_id INTEGER NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
				alert_type VARCHAR(20) NOT NULL CHECK (alert_type IN ('MISSED_CYCLE', 'HIGH_RISK', 'DORMANT')),
				expected_purchase_date DATE NOT NULL,
				days_overdue INTEGER NOT NULL,
				churn_probability DECIMAL(3,2) CHECK (churn_probability >= 0 AND churn_probability <= 1),
				status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'NOTIFIED', 'RESOLVED', 'IGNORED')),
				notified_at TIMESTAMPTZ,
				resolved_at TIMESTAMPTZ,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ,
				PRIMARY KEY (id)
			);

			CREATE INDEX IF NOT EXISTS idx_alerts_customer ON churn_alerts(customer_id);
			CREATE INDEX IF NOT EXISTS idx_alerts_batch ON churn_alerts(batch_id);
			CREATE INDEX IF NOT EXISTS idx_alerts_status ON churn_alerts(status);
			CREATE INDEX IF NOT EXISTS idx_alerts_expected_date ON churn_alerts(expected_purchase_date);
			CREATE INDEX IF NOT EXISTS idx_alerts_type ON churn_alerts(alert_type);

			ALTER SEQUENCE churn_alerts_id_seq OWNED BY churn_alerts.id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create churn_alerts table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop churn_alerts table")

		query := `
			DROP INDEX IF EXISTS idx_alerts_customer;
			DROP INDEX IF EXISTS idx_alerts_batch;
			DROP INDEX IF EXISTS idx_alerts_status;
			DROP INDEX IF EXISTS idx_alerts_expected_date;
			DROP INDEX IF EXISTS idx_alerts_type;
			DROP TABLE IF EXISTS churn_alerts CASCADE;
			DROP SEQUENCE IF EXISTS churn_alerts_id_seq;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop churn_alerts table: %w", err)
		}

		return nil
	})
}
