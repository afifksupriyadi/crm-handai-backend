package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] create transaction_batches table")

		query := `
			CREATE TABLE IF NOT EXISTS transaction_batches (
				id SERIAL PRIMARY KEY,
				batch_date DATE NOT NULL,
				filename VARCHAR(255) NOT NULL,
				customer_batch_id INT NOT NULL,
				imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				transaction_count INT NOT NULL DEFAULT 0,
				registered_transactions INT NOT NULL DEFAULT 0,
				guest_transactions INT NOT NULL DEFAULT 0,
				notes TEXT,
				
				CONSTRAINT fk_transaction_batches_customer_batch 
					FOREIGN KEY (customer_batch_id) 
					REFERENCES customer_batches(id) 
					ON DELETE RESTRICT
			);

			CREATE INDEX IF NOT EXISTS idx_transaction_batches_date ON transaction_batches(batch_date);
			CREATE INDEX IF NOT EXISTS idx_transaction_batches_customer ON transaction_batches(customer_batch_id);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create transaction_batches table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] drop transaction_batches table")

		query := `
			DROP INDEX IF EXISTS idx_transaction_batches_date;
			DROP INDEX IF EXISTS idx_transaction_batches_customer;
			DROP TABLE IF EXISTS transaction_batches CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop transaction_batches table: %w", err)
		}

		return nil
	})
}
