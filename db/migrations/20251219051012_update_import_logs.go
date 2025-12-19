package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] update import_logs with batch references")

		query := `
			ALTER TABLE import_logs 
			ADD COLUMN IF NOT EXISTS customer_batch_id INT REFERENCES customer_batches(id) ON DELETE SET NULL,
			ADD COLUMN IF NOT EXISTS transaction_batch_id INT REFERENCES transaction_batches(id) ON DELETE SET NULL;

			CREATE INDEX IF NOT EXISTS idx_import_logs_customer_batch ON import_logs(customer_batch_id);
			CREATE INDEX IF NOT EXISTS idx_import_logs_transaction_batch ON import_logs(transaction_batch_id);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to update import_logs: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove batch references from import_logs")

		query := `
			DROP INDEX IF EXISTS idx_import_logs_customer_batch;
			DROP INDEX IF EXISTS idx_import_logs_transaction_batch;
			
			ALTER TABLE import_logs 
			DROP COLUMN IF EXISTS customer_batch_id,
			DROP COLUMN IF EXISTS transaction_batch_id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to remove batch references from import_logs: %w", err)
		}

		return nil
	})
}
