package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] alter transactions add batch_id")

		query := `
			ALTER TABLE transactions 
			ADD COLUMN IF NOT EXISTS batch_id INTEGER REFERENCES batches(id) ON DELETE SET NULL;

			CREATE INDEX IF NOT EXISTS idx_transactions_batch_id ON transactions(batch_id);
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to alter transactions table: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove batch_id from transactions")

		query := `
			DROP INDEX IF EXISTS idx_transactions_batch_id;
			ALTER TABLE transactions DROP COLUMN IF EXISTS batch_id;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to rollback transactions table: %w", err)
		}

		return nil
	})
}
