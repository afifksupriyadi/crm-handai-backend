package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] drop old batches table")

		query := `
			-- Drop foreign key and column from transactions first
			ALTER TABLE transactions 
			DROP CONSTRAINT IF EXISTS transactions_batch_id_fkey;
			
			ALTER TABLE transactions 
			DROP COLUMN IF EXISTS batch_id;

			-- Drop indexes
			DROP INDEX IF EXISTS idx_batches_date;
			DROP INDEX IF EXISTS idx_batches_status;
			DROP INDEX IF EXISTS idx_batches_code;
			DROP INDEX IF EXISTS idx_batches_single_active;

			-- Drop the table
			DROP TABLE IF EXISTS batches CASCADE;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to drop old batches table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] recreate old batches table")

		query := `
			CREATE TABLE IF NOT EXISTS batches (
				id SERIAL PRIMARY KEY,
				batch_date DATE NOT NULL,
				batch_code VARCHAR(50) NOT NULL UNIQUE,
				status VARCHAR(20) NOT NULL,
				is_active BOOLEAN DEFAULT FALSE,
				customer_import_id INTEGER,
				transaction_import_id INTEGER,
				created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ
			);

			ALTER TABLE transactions 
			ADD COLUMN IF NOT EXISTS batch_id INTEGER REFERENCES batches(id) ON DELETE SET NULL;

			CREATE INDEX IF NOT EXISTS idx_batches_date ON batches(batch_date);
			CREATE INDEX IF NOT EXISTS idx_transactions_batch_id ON transactions(batch_id);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to recreate old batches table: %w", err)
		}

		return nil
	})
}
