package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create customer_predictions table")

		query := `
			CREATE SEQUENCE IF NOT EXISTS customer_predictions_id_seq;

			CREATE TABLE IF NOT EXISTS customer_predictions (
				id INTEGER NOT NULL DEFAULT nextval('customer_predictions_id_seq'::regclass),
				batch_id INTEGER NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
				customer_id INTEGER NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
				next_purchase_date DATE,
				confidence_score DECIMAL(3,2) CHECK (confidence_score >= 0 AND confidence_score <= 1),
				predicted_quantity INTEGER,
				predicted_products JSONB,
				avg_days_between_purchase DECIMAL(5,2),
				last_5_purchases JSONB,
				model_version VARCHAR(20),
				computed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				PRIMARY KEY (id),
				CONSTRAINT unique_customer_batch UNIQUE (batch_id, customer_id)
			);

			CREATE INDEX IF NOT EXISTS idx_predictions_batch ON customer_predictions(batch_id);
			CREATE INDEX IF NOT EXISTS idx_predictions_customer ON customer_predictions(customer_id);
			CREATE INDEX IF NOT EXISTS idx_predictions_next_date ON customer_predictions(next_purchase_date);
			CREATE INDEX IF NOT EXISTS idx_predictions_confidence ON customer_predictions(confidence_score);

			ALTER SEQUENCE customer_predictions_id_seq OWNED BY customer_predictions.id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create customer_predictions table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop customer_predictions table")

		query := `
			DROP INDEX IF EXISTS idx_predictions_batch;
			DROP INDEX IF EXISTS idx_predictions_customer;
			DROP INDEX IF EXISTS idx_predictions_next_date;
			DROP INDEX IF EXISTS idx_predictions_confidence;
			DROP TABLE IF EXISTS customer_predictions CASCADE;
			DROP SEQUENCE IF EXISTS customer_predictions_id_seq;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop customer_predictions table: %w", err)
		}

		return nil
	})
}
