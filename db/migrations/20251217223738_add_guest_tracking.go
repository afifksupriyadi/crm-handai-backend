package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [up migration] add guest tracking to transactions")

		query := `
			ALTER TABLE transactions 
			ADD COLUMN IF NOT EXISTS guest_name VARCHAR(50);

			CREATE INDEX IF NOT EXISTS idx_transactions_guest_name 
			ON transactions(guest_name) 
			WHERE customer_id IS NULL;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to add guest tracking to transactions: %w", err)
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Println(" [down migration] remove guest tracking from transactions")

		query := `
			DROP INDEX IF EXISTS idx_transactions_guest_name;
			ALTER TABLE transactions DROP COLUMN IF EXISTS guest_name;
		`

		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to remove guest tracking from transactions: %w", err)
		}

		return nil
	})
}
