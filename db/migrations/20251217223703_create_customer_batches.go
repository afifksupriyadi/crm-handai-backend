package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create customer_batches table")

		query := `
			CREATE TABLE IF NOT EXISTS customer_batches (
				id SERIAL PRIMARY KEY,
				batch_date DATE NOT NULL,
				filename VARCHAR(255) NOT NULL,
				imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				customer_count INT NOT NULL DEFAULT 0,
				new_customers INT NOT NULL DEFAULT 0,
				updated_customers INT NOT NULL DEFAULT 0,
				upgraded_from_guest INT NOT NULL DEFAULT 0,
				is_active BOOLEAN NOT NULL DEFAULT true,
				notes TEXT
			);

			CREATE INDEX IF NOT EXISTS idx_customer_batches_date ON customer_batches(batch_date);
			CREATE INDEX IF NOT EXISTS idx_customer_batches_active ON customer_batches(is_active);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create customer_batches table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop customer_batches table")

		query := `
			DROP INDEX IF EXISTS idx_customer_batches_date;
			DROP INDEX IF EXISTS idx_customer_batches_active;
			DROP TABLE IF EXISTS customer_batches CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop customer_batches table: %w", err)
		}

		return nil
	})
}
